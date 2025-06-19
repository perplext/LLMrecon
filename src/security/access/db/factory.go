// Package db provides database implementations of the access control interfaces
package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/perplext/LLMrecon/src/security/access/db/adapter"
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
	db *sql.DB
}

// NewFactory creates a new database factory
func NewFactory(config *DBConfig) (*Factory, error) {
	// Open database connection
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Factory{db: db}, nil
}

// Close closes the database connection
func (f *Factory) Close() error {
	return f.db.Close()
}

// CreateUserStore creates a new SQL-based user store
func (f *Factory) CreateUserStore() (interfaces.UserStore, error) {
	return NewSQLUserStore(f.db)
}

// CreateSessionStore creates a new SQL-based session store
func (f *Factory) CreateSessionStore() (interfaces.SessionStore, error) {
	// Create the SQL session store
	sqlStore, err := NewSQLSessionStore(f.db)
	if err != nil {
		return nil, err
	}
	
	// Wrap it with the adapter
	return adapter.NewSessionStoreAdapter(sqlStore), nil
}

// CreateAuditStore creates a new SQL-based audit store
func (f *Factory) CreateAuditStore() (interfaces.AuditLogger, error) {
	// Create the SQL audit store
	sqlStore, err := NewSQLAuditStore(f.db)
	if err != nil {
		return nil, err
	}
	
	// Wrap it with the adapter
	return adapter.NewAuditStoreAdapter(sqlStore), nil
}

// CreateIncidentStore creates a new SQL-based incident store
func (f *Factory) CreateIncidentStore() (interfaces.IncidentStore, error) {
	// Create the SQL incident store
	sqlStore, err := NewSQLIncidentStore(f.db)
	if err != nil {
		return nil, err
	}
	
	// Wrap it with the adapter
	return adapter.NewIncidentStoreAdapter(sqlStore), nil
}

// CreateVulnerabilityStore creates a new SQL-based vulnerability store
func (f *Factory) CreateVulnerabilityStore() (interfaces.VulnerabilityStore, error) {
	// Create the SQL vulnerability store
	sqlStore, err := NewSQLVulnerabilityStore(f.db)
	if err != nil {
		return nil, err
	}
	
	// Wrap it with the adapter
	return adapter.NewVulnerabilityStoreAdapter(sqlStore), nil
}

// CreateAllStores creates all stores and returns them in a map
func (f *Factory) CreateAllStores() (map[string]interface{}, error) {
	stores := make(map[string]interface{})

	// Create user store
	userStore, err := f.CreateUserStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create user store: %w", err)
	}
	stores["userStore"] = userStore

	// Create session store
	sessionStore, err := f.CreateSessionStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create session store: %w", err)
	}
	stores["sessionStore"] = sessionStore

	// Create audit store
	auditStore, err := f.CreateAuditStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create audit store: %w", err)
	}
	stores["auditStore"] = auditStore

	// Create incident store
	incidentStore, err := f.CreateIncidentStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create incident store: %w", err)
	}
	stores["incidentStore"] = incidentStore

	// Create vulnerability store
	vulnerabilityStore, err := f.CreateVulnerabilityStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create vulnerability store: %w", err)
	}
	stores["vulnerabilityStore"] = vulnerabilityStore

	return stores, nil
}

// ensureDir ensures that the specified directory exists
func ensureDir(dir string) error {
	// This is a placeholder. In a real implementation, you would use os.MkdirAll
	// to create the directory if it doesn't exist.
	return nil
}
