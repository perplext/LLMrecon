package repository

import (
	"context"
	"fmt"
	"strings"
)

// LocalFSRepository implements the Repository interface for local file system repositories
type LocalFSRepository struct {
	*BaseRepository
	
	// rootPath is the root path of the repository
	rootPath string
}

// NewLocalFSRepository creates a new local file system repository
func NewLocalFSRepository(config *Config) (Repository, error) {
	// Create base repository
	base := NewBaseRepository(config)
	
	// Parse URL to extract path
	rootPath, err := parseLocalPath(config.URL)
	if err != nil {
		return nil, err
	}
	
	// Check if path exists
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("invalid local path: %w", err)
	}
	
	// Check if path is a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("local path is not a directory: %s", rootPath)
	}
	
	return &LocalFSRepository{
		BaseRepository: base,
		rootPath:       rootPath,
	}, nil
}

// parseLocalPath parses a local path from a URL or file path
func parseLocalPath(urlOrPath string) (string, error) {
	// Handle file:// URLs
	if strings.HasPrefix(urlOrPath, "file://") {
		return strings.TrimPrefix(urlOrPath, "file://"), nil
	}
	
	// Handle absolute paths
	if filepath.IsAbs(urlOrPath) {
		return urlOrPath, nil
	}
	
	// Handle relative paths
	absPath, err := filepath.Abs(urlOrPath)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %w", err)
	}
	
	return absPath, nil
}

// Connect establishes a connection to the local file system repository
func (r *LocalFSRepository) Connect(ctx context.Context) error {
	// Check if already connected
	if r.IsConnected() {
		return nil
	}
	
	// Check if path exists
	_, err := os.Stat(r.rootPath)
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to connect to local repository: %w", err)
	}
	
	// Set connected flag
	r.setConnected(true)
	
	return nil
}

// Disconnect closes the connection to the local file system repository
func (r *LocalFSRepository) Disconnect() error {
	// Set connected flag to false
	r.setConnected(false)
	
	return nil
}

// ListFiles lists files in the local file system repository matching the pattern
func (r *LocalFSRepository) ListFiles(ctx context.Context, pattern string) ([]FileInfo, error) {
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
		// Walk the directory
		return filepath.Walk(r.rootPath, func(path string, info os.FileInfo, err error) error {
			// Check for context cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}
			
			// Check for errors
			if err != nil {
				return err
			}
			
			// Skip the root directory itself
			if path == r.rootPath {
				return nil
			}
			
			// Get relative path
			relPath, err := filepath.Rel(r.rootPath, path)
			if err != nil {
				return err
			}
			
			// Convert to forward slashes for consistency
			relPath = filepath.ToSlash(relPath)
			
			// Skip if not matching pattern
			if pattern != "" && !matchPattern(filepath.Base(relPath), pattern) {
				// Continue walking if it's a directory
				if info.IsDir() {
					return nil
				}
				// Skip this file
				return nil
			}
			
			// Create file info
			fileInfo := FileInfo{
				Path:         relPath,
				Name:         info.Name(),
				Size:         info.Size(),
				LastModified: info.ModTime(),
				IsDirectory:  info.IsDir(),
			}
			
			// Add to result
			result = append(result, fileInfo)
			
			return nil
		})
	})
	
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetFile retrieves a file from the local file system repository
func (r *LocalFSRepository) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return nil, err
	}
	
	// Create full path
	fullPath := filepath.Join(r.rootPath, filepath.FromSlash(path))
	
	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		r.ReleaseConnection()
		return nil, err
	}
	
	// Create a wrapper for the file that releases the connection when closed
	return &connectionCloser{
		ReadCloser: file,
		release: func() {
			r.ReleaseConnection()
		},
	}, nil
}

// FileExists checks if a file exists in the local file system repository
func (r *LocalFSRepository) FileExists(ctx context.Context, path string) (bool, error) {
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return false, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return false, err
	}
	defer r.ReleaseConnection()
	
	// Create full path
	fullPath := filepath.Join(r.rootPath, filepath.FromSlash(path))
	
	// Use WithRetry for the operation
	var exists bool
	err := r.WithRetry(ctx, func() error {
		// Check if file exists
		_, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				exists = false
				return nil
			}
			return err
		}
		
		exists = true
		return nil
	})
	
	if err != nil {
		return false, err
	}
	
	return exists, nil
}

// GetBranch returns the branch of the repository
// Local file system repositories don't have branches, so this returns an empty string
func (r *LocalFSRepository) GetBranch() string {
	return ""
}

// GetLastModified gets the last modified time of a file in the local file system repository
func (r *LocalFSRepository) GetLastModified(ctx context.Context, path string) (time.Time, error) {
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return time.Time{}, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return time.Time{}, err
	}
	defer r.ReleaseConnection()
	
	// Create full path
	fullPath := filepath.Join(r.rootPath, filepath.FromSlash(path))
	
	// Use WithRetry for the operation
	var lastModified time.Time
	err := r.WithRetry(ctx, func() error {
		// Get file info
		info, err := os.Stat(fullPath)
		if err != nil {
			return err
		}
		
		lastModified = info.ModTime()
		return nil
	})
	
	if err != nil {
		return time.Time{}, err
	}
	
	return lastModified, nil
}
