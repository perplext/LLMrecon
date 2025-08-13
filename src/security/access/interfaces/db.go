// Package interfaces defines common interfaces and types for the access control system
package interfaces

// DBConfig contains configuration for the database
type DBConfig struct {
	// Driver is the database driver name (e.g., "sqlite3", "postgres")
	Driver string

	// DSN is the data source name or connection string
	DSN string

	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns int

	// MaxIdleConns is the maximum number of connections in the idle connection pool
	MaxIdleConns int

	// ConnMaxLifetime is the maximum amount of time a connection may be reused
	ConnMaxLifetime int

	// MigrationDir is the directory containing database migration files
	MigrationDir string

	// AutoMigrate indicates whether to automatically run migrations on startup
	AutoMigrate bool
}

// DBFactory defines an interface for creating database connections and stores
type DBFactory interface {
	// GetUserStore returns a user store implementation
	GetUserStore() (UserStore, error)

	// GetSessionStore returns a session store implementation
	GetSessionStore() (SessionStore, error)

	// GetAuditStore returns an audit store implementation
	GetAuditStore() (AuditLogger, error)

	// GetIncidentStore returns an incident store implementation
	GetIncidentStore() (IncidentStore, error)

	// GetVulnerabilityStore returns a vulnerability store implementation
	GetVulnerabilityStore() (VulnerabilityStore, error)

	// CreateAllStores creates all stores and returns any errors
	CreateAllStores() error

	// Close closes all database connections
	Close() error
}
