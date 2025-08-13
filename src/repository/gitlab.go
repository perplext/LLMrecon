package repository

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	
	"github.com/xanzy/go-gitlab"
)

// GitLabRepository implements the Repository interface for GitLab repositories
type GitLabRepository struct {
	*BaseRepository
	
	// client is the GitLab API client
	client *gitlab.Client
	
	// projectID is the GitLab project ID or path
	projectID string
}

// NewGitLabRepository creates a new GitLab repository
func NewGitLabRepository(config *Config) (Repository, error) {
	// Create base repository
	base := NewBaseRepository(config)
	
	// Parse GitLab URL to extract project ID
	projectID, err := parseGitLabURL(config.URL)
	if err != nil {
		return nil, err
	}
	
	return &GitLabRepository{
		BaseRepository: base,
		projectID:      projectID,
	}, nil
}

// parseGitLabURL parses a GitLab URL to extract project ID or path
func parseGitLabURL(urlStr string) (string, error) {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid GitLab URL: %w", err)
	}
	
	// Extract path
	path := parsedURL.Path
	
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	
	// Remove .git suffix if present
	path = strings.TrimSuffix(path, ".git")
	
	// URL encode the path
	projectID := url.PathEscape(path)
	
	return projectID, nil
}

// Connect establishes a connection to the GitLab repository
func (r *GitLabRepository) Connect(ctx context.Context) error {
	// Check if already connected
	if r.IsConnected() {
		return nil
	}
	
	var err error
	
	// Create GitLab client options
	opts := []gitlab.ClientOptionFunc{
		gitlab.WithBaseURL(getGitLabBaseURL(r.config.URL)),
	}
	
	// Add authentication if provided
	if r.config.Password != "" {
		// Use token-based authentication
		// Pass token directly to NewClient instead of using WithToken
	}
	// Basic auth is not directly supported in newer gitlab client versions
	
	// Create GitLab client with token if available
	if r.config.Password != "" {
		r.client, err = gitlab.NewClient(r.config.Password, opts...)
	} else {
		r.client, err = gitlab.NewClient("", opts...)
	}
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}
	
	// Test connection by fetching project info
	_, _, err = r.client.Projects.GetProject(r.projectID, &gitlab.GetProjectOptions{})
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to connect to GitLab repository: %w", err)
	}
	
	// Set connected flag
	r.setConnected(true)
	
	return nil
}

// getGitLabBaseURL extracts the base URL from a GitLab repository URL
func getGitLabBaseURL(repoURL string) string {
	// Parse URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "https://gitlab.com"
	}
	
	// Extract scheme and host
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	
	// If no scheme, use https
	if parsedURL.Scheme == "" {
		baseURL = fmt.Sprintf("https://%s", parsedURL.Host)
	}
	
	// If no host, use gitlab.com
	if parsedURL.Host == "" {
		baseURL = "https://gitlab.com"
	}
	
	return baseURL
}

// Disconnect closes the connection to the GitLab repository
func (r *GitLabRepository) Disconnect() error {
	// Set connected flag to false
	r.setConnected(false)
	
	// Clear client
	r.client = nil
	
	return nil
}

// ListFiles lists files in the GitLab repository matching the pattern
func (r *GitLabRepository) ListFiles(ctx context.Context, pattern string) ([]FileInfo, error) {
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
		// Get tree of the repository
		treeItems, _, err := r.client.Repositories.ListTree(r.projectID, &gitlab.ListTreeOptions{
			Ref:       &r.config.Branch,
			Recursive: gitlab.Bool(true),
		})
		if err != nil {
			return err
		}
		
		// Process tree items
		for _, item := range treeItems {
			// Skip if not matching pattern
			if pattern != "" && !matchPattern(item.Name, pattern) {
				continue
			}
			
			// Create file info
			fileInfo := FileInfo{
				Path:        item.Path,
				Name:        item.Name,
				IsDirectory: item.Type == "tree",
			}
			
			// Get file details if it's a file
			if !fileInfo.IsDirectory {
				// Get file size and last modified time
				file, _, err := r.client.RepositoryFiles.GetFile(r.projectID, item.Path, &gitlab.GetFileOptions{
					Ref: gitlab.String(r.config.Branch),
				})
				if err == nil {
					fileInfo.Size = int64(file.Size)
					
					// Get last modified time
					lastModified, err := r.getFileLastModified(ctx, item.Path)
					if err == nil {
						fileInfo.LastModified = lastModified
					}
				}
			}
			
			// Add to result
			result = append(result, fileInfo)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetFile retrieves a file from the GitLab repository
func (r *GitLabRepository) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return nil, err
	}
	
	// Create a pipe for streaming the content
	pr, pw := io.Pipe()
	
	// Fetch and write content in a goroutine
	go func() {
		defer r.ReleaseConnection()
		
		var fetchErr error
		
		// Use WithRetry for the operation
		fetchErr = r.WithRetry(ctx, func() error {
			// Get file content
			file, _, err := r.client.RepositoryFiles.GetFile(r.projectID, path, &gitlab.GetFileOptions{
				Ref: gitlab.String(r.config.Branch),
			})
			if err != nil {
				return err
			}
			
			// Get content from file
			// The GitLab API returns base64 encoded content in the Content field
			content, err := base64.StdEncoding.DecodeString(file.Content)
			if err != nil {
				return err
			}
			
			// Write content to pipe
			_, err = pw.Write(content)
			return err
		})
		
		// Close pipe with error if any
		if fetchErr != nil {
			pw.CloseWithError(fetchErr)
		} else {
			pw.Close()
		}
	}()
	
	return pr, nil
}

// FileExists checks if a file exists in the GitLab repository
func (r *GitLabRepository) FileExists(ctx context.Context, path string) (bool, error) {
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
		// Get file
		_, resp, err := r.client.RepositoryFiles.GetFile(r.projectID, path, &gitlab.GetFileOptions{
			Ref: gitlab.String(r.config.Branch),
		})
		
		// Check if file exists
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
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
func (r *GitLabRepository) GetBranch() string {
	return r.config.Branch
}

// GetLastModified gets the last modified time of a file in the GitLab repository
func (r *GitLabRepository) GetLastModified(ctx context.Context, path string) (time.Time, error) {
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return time.Time{}, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return time.Time{}, err
	}
	defer r.ReleaseConnection()
	
	return r.getFileLastModified(ctx, path)
}

// getFileLastModified gets the last modified time of a file (internal implementation)
func (r *GitLabRepository) getFileLastModified(ctx context.Context, path string) (time.Time, error) {
	var lastModified time.Time
	
	// Use WithRetry for the operation
	err := r.WithRetry(ctx, func() error {
		// Get commits for the file
		opts := &gitlab.ListCommitsOptions{
			Path:    gitlab.String(path),
			RefName: gitlab.String(r.config.Branch),
		}
		// Set pagination options
		opts.Page = 1
		opts.PerPage = 1
		commits, _, err := r.client.Commits.ListCommits(r.projectID, opts)
		if err != nil {
			return err
		}
		
		// Check if any commits found
		if len(commits) == 0 {
			return fmt.Errorf("no commits found for file: %s", path)
		}
		
		// Get last commit date
		lastModified = *commits[0].CommittedDate
		return nil
	})
	
	if err != nil {
		return time.Time{}, err
	}
	
	return lastModified, nil
}
