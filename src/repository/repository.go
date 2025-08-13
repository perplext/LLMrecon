package repository

import (
	"context"

	"github.com/perplext/LLMrecon/src/audit"
)

// RepositoryType represents the type of repository
type RepositoryType string

const (
	// GitHub repository type
	GitHub RepositoryType = "github"
	// GitLab repository type
	GitLab RepositoryType = "gitlab"
	// LocalFS repository type (local file system)
	LocalFS RepositoryType = "local"
	// HTTP repository type (generic HTTP/HTTPS)
	HTTP RepositoryType = "http"
	// Database repository type
	Database RepositoryType = "database"
	// S3 repository type
	S3 RepositoryType = "s3"
)

// Repository defines the interface for interacting with a template/module repository
type Repository interface {
	// Connect establishes a connection to the repository
	Connect(ctx context.Context) error
	
	// Disconnect closes the connection to the repository
	Disconnect() error
	
	// IsConnected returns true if the repository is connected
	IsConnected() bool
	
	// GetName returns the name of the repository
	GetName() string
	
	// GetBranch returns the branch of the repository
	GetBranch() string
	
	// GetType returns the type of the repository
	GetType() RepositoryType
	
	// GetURL returns the repository URL
	GetURL() string
	
	// ListFiles lists files in the repository matching the pattern
	ListFiles(ctx context.Context, pattern string) ([]FileInfo, error)
	
	// GetFile retrieves a file from the repository
	GetFile(ctx context.Context, filePath string) (io.ReadCloser, error)
	
	// FileExists checks if a file exists in the repository
	FileExists(ctx context.Context, filePath string) (bool, error)
	
	// GetLastModified gets the last modified time of a file in the repository
	GetLastModified(ctx context.Context, filePath string) (time.Time, error)
}

// FileInfo represents information about a file in a repository
type FileInfo struct {
	// Path is the path to the file within the repository
	Path string
	
	// Name is the name of the file
	Name string
	
	// Size is the size of the file in bytes
	Size int64
	
	// LastModified is the time the file was last modified
	LastModified time.Time
	
	// IsDirectory indicates if the file is a directory
	IsDirectory bool
}

// RepositoryInfo represents information about a repository
type RepositoryInfo struct {
	// Type is the repository type
	Type RepositoryType
	// Name is the repository name
	Name string
	// URL is the repository URL
	URL string
	// Branch is the repository branch
	Branch string
	// LocalPath is the local path where the repository is stored
	LocalPath string
	// CurrentVersion is the current version (commit hash or tag)
	CurrentVersion string
	// LatestVersion is the latest available version
	LatestVersion string
	// Description is a description of the repository
	Description string
	// LastSynced is the time the repository was last synced
	LastSynced time.Time
}

// Config represents the configuration for a repository
type Config struct {
	// Type is the repository type
	Type RepositoryType
	
	// Name is the repository name
	Name string
	
	// URL is the repository URL
	URL string
	
	// Branch is the repository branch (for Git repositories)
	Branch string
	
	// Username is the username for authentication
	Username string
	
	// Password is the password or token for authentication
	Password string
	
	// Timeout is the timeout for repository operations
	Timeout time.Duration
	
	// AuditLogger is the audit logger for repository operations
	AuditLogger *audit.AuditLogger
	
	// ProxyURL is the URL of the proxy server
	ProxyURL string
	
	// InsecureSkipVerify disables TLS certificate verification
	InsecureSkipVerify bool
	
	// MaxConnections is the maximum number of connections to the repository
	MaxConnections int
	
	// RetryCount is the number of times to retry a failed operation
	RetryCount int
	
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// NewConfig creates a new repository configuration with default values
func NewConfig(repoType RepositoryType, name, url string) *Config {
	return &Config{
		Type:            repoType,
		Name:            name,
		URL:             url,
		Branch:          "main",
		Timeout:         30 * time.Second,
		MaxConnections:  5,
		RetryCount:      3,
		RetryDelay:      2 * time.Second,
	}
}
