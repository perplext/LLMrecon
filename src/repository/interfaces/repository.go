// Package interfaces defines interfaces for repository operations
package interfaces

import (
	"context"
	"io"
	"time"
)

// FileInfo represents information about a file in a repository
type FileInfo struct {
	// Path is the path of the file
	Path string
	// Size is the size of the file in bytes
	Size int64
	// ModTime is the last modified time of the file
	ModTime time.Time
	// IsDir indicates if the file is a directory
	IsDir bool
}

// Repository defines the interface for repository operations
type Repository interface {
	// Connect establishes a connection to the repository
	Connect(ctx context.Context) error
	
	// Disconnect closes the connection to the repository
	Disconnect() error
	
	// ListFiles lists files in the repository matching the pattern
	ListFiles(ctx context.Context, pattern string) ([]FileInfo, error)
	
	// GetFile retrieves a file from the repository
	GetFile(ctx context.Context, filePath string) (io.ReadCloser, error)
	
	// FileExists checks if a file exists in the repository
	FileExists(ctx context.Context, filePath string) (bool, error)
	
	// GetLastModified gets the last modified time of a file in the repository
	GetLastModified(ctx context.Context, filePath string) (time.Time, error)
	
	// GetURL returns the URL of the repository
	GetURL() string
	
	// GetType returns the type of the repository
	GetType() string
}

// Config represents configuration for a repository
type Config struct {
	// URL is the URL of the repository
	URL string
	
	// Type is the type of repository (HTTP, File, etc.)
	Type string
	
	// Credentials are the credentials for the repository
	Credentials map[string]string
	
	// Options are additional options for the repository
	Options map[string]interface{}
	
	// Timeout is the timeout for repository operations
	Timeout time.Duration
}

// AuditLogger defines the interface for repository audit logging
type AuditLogger interface {
	// LogRepositoryConnect logs a repository connection event
	LogRepositoryConnect(ctx context.Context, repositoryID string)
	
	// LogRepositoryConnectSuccess logs a successful repository connection event
	LogRepositoryConnectSuccess(ctx context.Context, repositoryID string)
	
	// LogRepositoryConnectFailure logs a failed repository connection event
	LogRepositoryConnectFailure(ctx context.Context, repositoryID string, err error)
	
	// LogRepositoryDisconnect logs a repository disconnection event
	LogRepositoryDisconnect(ctx context.Context, repositoryID string)
	
	// LogRepositoryAccess logs a repository access event
	LogRepositoryAccess(ctx context.Context, repositoryID string, filePath string)
	
	// LogRepositoryError logs a repository error event
	LogRepositoryError(ctx context.Context, repositoryID string, err error)
}

// Factory creates repository instances
type Factory interface {
	// CreateRepository creates a repository from a configuration
	CreateRepository(config *Config) (Repository, error)
}
