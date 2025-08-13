// Package adapter provides adapters between database interfaces and domain models
package adapter

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"os"

	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// DBConfig contains configuration for database connections
type DBConfig struct {
	// Driver is the database driver to use (e.g., "sqlite3", "postgres")
	Driver string

	// DSN is the data source name for the database connection
	DSN string

	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections in the connection pool
	MaxIdleConns int
}

// DefaultSQLiteConfig returns a default SQLite configuration
func DefaultSQLiteConfig(dbPath string) *DBConfig {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := ensureDir(dir); err != nil {
			// Just log the error, don't fail
			fmt.Printf("Warning: failed to create directory for SQLite database: %v\n", err)
		}
	}

	return &DBConfig{
		Driver:       "sqlite3",
		DSN:          dbPath,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}
}

// Factory is a factory for creating database-backed access control components
type Factory struct {
	db       *sql.DB
	dbConfig *DBConfig
}

// NewFactory creates a new database factory
func NewFactory(config *DBConfig) (*Factory, error) {
	if config == nil {
		return nil, fmt.Errorf("database configuration is required")
	}

	// Open database connection
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Factory{
		db:       db,
		dbConfig: config,
	}, nil
}

// Close closes the database connection
func (f *Factory) Close() error {
	return f.db.Close()
}

// CreateUserStoreAdapter creates a new user store adapter
func (f *Factory) CreateUserStoreAdapter() (*UserStoreAdapter, error) {
	// Create the underlying user store
	userStore, err := f.createUserStore()
	if err != nil {
		return nil, err
	}

	// Create the adapter
	return NewUserStoreAdapter(userStore), nil
}

// CreateSessionStoreAdapter creates a new session store adapter
func (f *Factory) CreateSessionStoreAdapter() (*SessionStoreAdapter, error) {
	// Create the underlying session store
	sessionStore, err := f.createSessionStore()
	if err != nil {
		return nil, err
	}

	// Create the adapter
	return NewSessionStoreAdapter(sessionStore), nil
}

// createUserStore creates a new SQL-based user store
func (f *Factory) createUserStore() (interfaces.UserStore, error) {
	// Import the actual implementation from the db package
	// This is done to avoid import cycles
	return nil, fmt.Errorf("not implemented yet")
}

// createSessionStore creates a new SQL-based session store
func (f *Factory) createSessionStore() (SessionStore, error) {
	// This would create an adapter.SessionStore implementation
	// For now, return an error as this is a stub
	return nil, fmt.Errorf("not implemented yet")
}

// ensureDir ensures that the specified directory exists
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
