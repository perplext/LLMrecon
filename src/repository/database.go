package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	// Import database drivers as needed
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	// MySQL database type
	MySQL DatabaseType = "mysql"
	// PostgreSQL database type
	PostgreSQL DatabaseType = "postgres"
	// SQLite database type
	SQLite DatabaseType = "sqlite"
)

// DatabaseRepository implements the Repository interface for database repositories
type DatabaseRepository struct {
	*BaseRepository

	// db is the database connection
	db *sql.DB

	// dbType is the database type
	dbType DatabaseType

	// tableName is the name of the table storing templates
	tableName string

	// auditLogger is the audit logger for repository operations
	auditLogger *RepositoryAuditLogger
}

// DatabaseConfig extends the repository Config with database-specific options
type DatabaseConfig struct {
	// Type is the database type
	Type DatabaseType

	// TableName is the name of the table storing templates
	TableName string

	// ConnectionString is the database connection string
	ConnectionString string
}

// NewDatabaseRepository creates a new database repository
func NewDatabaseRepository(config *Config) (Repository, error) {
	// Create base repository
	base := NewBaseRepository(config)

	// Parse database URL to extract database type and connection string
	dbType, tableName, connStr, err := parseDatabaseURL(config.URL)
	if err != nil {
		return nil, err
	}

	// Create audit logger if audit logging is enabled
	var auditLogger *RepositoryAuditLogger
	if config.AuditLogger != nil {
		auditLogger = NewRepositoryAuditLogger(config.AuditLogger, string(dbType), config.URL)
	}

	// Open database connection
	db, err := sql.Open(string(dbType), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	return &DatabaseRepository{
		BaseRepository: base,
		db:             db,
		dbType:         dbType,
		tableName:      tableName,
		auditLogger:    auditLogger,
	}, nil
}

// parseDatabaseURL parses a database URL to extract database type, table name, and connection string
// Format: dbtype://connection_string#table_name
func parseDatabaseURL(urlStr string) (DatabaseType, string, string, error) {
	// Split URL by protocol separator
	parts := strings.SplitN(urlStr, "://", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid database URL format: %s", urlStr)
	}

	// Extract database type
	dbTypeStr := parts[0]
	var dbType DatabaseType
	switch strings.ToLower(dbTypeStr) {
	case "mysql":
		dbType = MySQL
	case "postgres", "postgresql":
		dbType = PostgreSQL
	case "sqlite", "sqlite3":
		dbType = SQLite
	default:
		return "", "", "", fmt.Errorf("unsupported database type: %s", dbTypeStr)
	}

	// Extract connection string and table name
	connStrAndTable := parts[1]
	tableName := "templates" // Default table name

	// Check if table name is specified
	if idx := strings.LastIndex(connStrAndTable, "#"); idx != -1 {
		tableName = connStrAndTable[idx+1:]
		connStrAndTable = connStrAndTable[:idx]
	}

	return dbType, tableName, connStrAndTable, nil
}

// Connect establishes a connection to the database repository
func (r *DatabaseRepository) Connect(ctx context.Context) error {
	// Log repository connection attempt
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryConnect(ctx, r.config.URL)
	}

	// Check if already connected
	if r.IsConnected() {
		return nil
	}

	// Parse database URL to extract connection string
	_, _, connStr, err := parseDatabaseURL(r.config.URL)
	if err != nil {
		return err
	}

	// Open database connection
	var driverName string
	switch r.dbType {
	case MySQL:
		driverName = "mysql"
	case PostgreSQL:
		driverName = "postgres"
	case SQLite:
		driverName = "sqlite3"
	default:
		return fmt.Errorf("unsupported database type: %s", r.dbType)
	}

	// Open database connection
	r.db, err = sql.Open(driverName, connStr)
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	r.db.SetMaxOpenConns(r.config.MaxConnections)
	r.db.SetMaxIdleConns(r.config.MaxConnections)
	r.db.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	if err := r.db.PingContext(ctx); err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Ensure the templates table exists
	if err := r.ensureTableExists(ctx); err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to ensure table exists: %w", err)
	}

	// Set connected flag
	r.setConnected(true)

	return nil
}

// ensureTableExists ensures the templates table exists in the database
func (r *DatabaseRepository) ensureTableExists(ctx context.Context) error {
	var createTableSQL string

	switch r.dbType {
	case MySQL:
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				path VARCHAR(255) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				content LONGBLOB NOT NULL,
				size BIGINT NOT NULL,
				is_directory BOOLEAN NOT NULL,
				last_modified TIMESTAMP NOT NULL
			)
		`, r.tableName)
	case PostgreSQL:
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				path VARCHAR(255) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				content BYTEA NOT NULL,
				size BIGINT NOT NULL,
				is_directory BOOLEAN NOT NULL,
				last_modified TIMESTAMP NOT NULL
			)
		`, r.tableName)
	case SQLite:
		createTableSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				path TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				content BLOB NOT NULL,
				size INTEGER NOT NULL,
				is_directory INTEGER NOT NULL,
				last_modified TIMESTAMP NOT NULL
			)
		`, r.tableName)
	default:
		return fmt.Errorf("unsupported database type: %s", r.dbType)
	}

	// Execute create table statement
	_, err := r.db.ExecContext(ctx, createTableSQL)
	return err
}

// Disconnect closes the connection to the database repository
func (r *DatabaseRepository) Disconnect() error {
	// Log repository disconnection
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryDisconnect(context.Background(), r.config.URL)
	}

	// Check if connected
	if !r.IsConnected() {
		return nil
	}

	// Close database connection
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
		r.db = nil
	}

	// Set connected flag to false
	r.setConnected(false)

	return nil
}

// ListFiles lists files in the database repository matching the pattern
func (r *DatabaseRepository) ListFiles(ctx context.Context, pattern string) ([]FileInfo, error) {
	// Log file listing operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryListFiles(ctx, r.config.URL, pattern)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return nil, err
	}
	defer r.ReleaseConnection()

	// Create result slice
	var result []FileInfo

	// Use WithRetry for the operation
	err := r.WithRetry(ctx, func() error {
		// Prepare query
		var query string
		var args []interface{}

		if pattern != "" {
			// Use LIKE for pattern matching
			query = fmt.Sprintf("SELECT path, name, size, is_directory, last_modified FROM %s WHERE name LIKE ?", r.tableName)
			args = append(args, "%"+strings.ReplaceAll(pattern, "*", "%")+"%")
		} else {
			query = fmt.Sprintf("SELECT path, name, size, is_directory, last_modified FROM %s", r.tableName)
		}

		// Execute query
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		// Process results
		for rows.Next() {
			var fileInfo FileInfo
			var isDir int
			var lastModified time.Time

			if err := rows.Scan(&fileInfo.Path, &fileInfo.Name, &fileInfo.Size, &isDir, &lastModified); err != nil {
				return err
			}

			fileInfo.IsDirectory = isDir != 0
			fileInfo.LastModified = lastModified

			result = append(result, fileInfo)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetFile retrieves a file from the database repository
func (r *DatabaseRepository) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// Log file retrieval operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryGetFile(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return nil, err
	}

	// Use WithRetry for the operation
	var content []byte
	err := r.WithRetry(ctx, func() error {
		// Prepare query
		query := fmt.Sprintf("SELECT content FROM %s WHERE path = ?", r.tableName)

		// Execute query
		row := r.db.QueryRowContext(ctx, query, path)

		// Scan result
		return row.Scan(&content)
	})

	if err != nil {
		r.ReleaseConnection()
		return nil, err
	}

	// Create a reader for the content
	reader := io.NopCloser(strings.NewReader(string(content)))

	// Create a wrapper for the reader that releases the connection when closed
	return &connectionCloser{
		ReadCloser: reader,
		release: func() {
			r.ReleaseConnection()
		},
		ctx:         ctx,
		auditLogger: r.auditLogger,
		filePath:    path,
		baseURL:     r.config.URL,
	}, nil
}

// FileExists checks if a file exists in the database repository
func (r *DatabaseRepository) FileExists(ctx context.Context, path string) (bool, error) {
	// Log file existence check operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryFileExists(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return false, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return false, err
	}
	defer r.ReleaseConnection()

	// Use WithRetry for the operation
	var exists bool
	err := r.WithRetry(ctx, func() error {
		// Prepare query
		query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE path = ?)", r.tableName)

		// Execute query
		row := r.db.QueryRowContext(ctx, query, path)

		// Scan result
		return row.Scan(&exists)
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetBranch returns the branch of the repository
// Database repositories don't have branches, so this returns an empty string
func (r *DatabaseRepository) GetBranch() string {
	return ""
}

// GetLastModified gets the last modified time of a file in the database repository
func (r *DatabaseRepository) GetLastModified(ctx context.Context, path string) (time.Time, error) {
	// Log last modified time retrieval operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryGetLastModified(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return time.Time{}, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return time.Time{}, err
	}
	defer r.ReleaseConnection()

	// Use WithRetry for the operation
	var lastModified time.Time
	err := r.WithRetry(ctx, func() error {
		// Prepare query
		query := fmt.Sprintf("SELECT last_modified FROM %s WHERE path = ?", r.tableName)

		// Execute query
		row := r.db.QueryRowContext(ctx, query, path)

		// Scan result
		return row.Scan(&lastModified)
	})

	if err != nil {
		return time.Time{}, err
	}

	return lastModified, nil
}

// StoreFile stores a file in the database repository
func (r *DatabaseRepository) StoreFile(ctx context.Context, path string, name string, content []byte, isDirectory bool) error {
	// Log file storage operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryStoreFile(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return err
	}
	defer r.ReleaseConnection()

	// Use WithRetry for the operation
	return r.WithRetry(ctx, func() error {
		// Prepare query
		query := fmt.Sprintf(`
			INSERT INTO %s (path, name, content, size, is_directory, last_modified)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (path) DO UPDATE SET
				name = EXCLUDED.name,
				content = EXCLUDED.content,
				size = EXCLUDED.size,
				is_directory = EXCLUDED.is_directory,
				last_modified = EXCLUDED.last_modified
		`, r.tableName)

		// For SQLite, use a different syntax for upsert
		if r.dbType == SQLite {
			query = fmt.Sprintf(`
				INSERT OR REPLACE INTO %s (path, name, content, size, is_directory, last_modified)
				VALUES (?, ?, ?, ?, ?, ?)
			`, r.tableName)
		}

		// Execute query
		_, err := r.db.ExecContext(
			ctx,
			query,
			path,
			name,
			content,
			len(content),
			isDirectory,
			time.Now(),
		)

		return err
	})
}

// DeleteFile deletes a file from the database repository
func (r *DatabaseRepository) DeleteFile(ctx context.Context, path string) error {
	// Log file deletion operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryDeleteFile(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return err
	}
	defer r.ReleaseConnection()

	// Use WithRetry for the operation
	return r.WithRetry(ctx, func() error {
		// Prepare query
		query := fmt.Sprintf("DELETE FROM %s WHERE path = ?", r.tableName)

		// Execute query
		_, err := r.db.ExecContext(ctx, query, path)
		return err
	})
}

// init registers the database repository type with the default factory
func init() {
	// Register the Database repository type
	DefaultFactory.Register("database", NewDatabaseRepository)
}
