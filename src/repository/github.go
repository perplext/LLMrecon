package repository

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// GitHubRepository implements the Repository interface for GitHub repositories
type GitHubRepository struct {
	*BaseRepository
	
	// client is the GitHub API client
	client *github.Client
	
	// owner is the GitHub repository owner
	owner string
	
	// repo is the GitHub repository name
	repo string
}

// NewGitHubRepository creates a new GitHub repository
func NewGitHubRepository(config *Config) (Repository, error) {
	// Create base repository
	base := NewBaseRepository(config)
	
	// Parse GitHub URL to extract owner and repo
	owner, repo, err := parseGitHubURL(config.URL)
	if err != nil {
		return nil, err
	}
	
	return &GitHubRepository{
		BaseRepository: base,
		owner:          owner,
		repo:           repo,
	}, nil
}

// parseGitHubURL parses a GitHub URL to extract owner and repo
func parseGitHubURL(url string) (string, string, error) {
	// Remove protocol and domain
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	
	// Split by slash
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
	}
	
	owner := parts[0]
	repo := parts[1]
	
	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")
	
	return owner, repo, nil
}

// Connect establishes a connection to the GitHub repository
func (r *GitHubRepository) Connect(ctx context.Context) error {
	// Check if already connected
	if r.IsConnected() {
		return nil
	}
	
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: r.config.Timeout,
	}
	
	// Add authentication if provided
	if r.config.Username != "" && r.config.Password != "" {
		// Use token-based authentication
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: r.config.Password},
		)
		httpClient = oauth2.NewClient(ctx, ts)
	}
	
	// Create GitHub client
	r.client = github.NewClient(httpClient)
	
	// Test connection by fetching repository info
	_, _, err := r.client.Repositories.Get(ctx, r.owner, r.repo)
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to connect to GitHub repository: %w", err)
	}
	
	// Set connected flag
	r.setConnected(true)
	
	return nil
}

// Disconnect closes the connection to the GitHub repository
func (r *GitHubRepository) Disconnect() error {
	// Set connected flag to false
	r.setConnected(false)
	
	// Clear client
	r.client = nil
	
	return nil
}

// ListFiles lists files in the GitHub repository matching the pattern
func (r *GitHubRepository) ListFiles(ctx context.Context, pattern string) ([]FileInfo, error) {
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
		// Get content of the repository
		_, contents, _, err := r.client.Repositories.GetContents(
			ctx,
			r.owner,
			r.repo,
			"", // Root directory
			&github.RepositoryContentGetOptions{
				Ref: r.config.Branch,
			},
		)
		if err != nil {
			return err
		}
		
		// Process contents
		for _, content := range contents {
			// Skip if not matching pattern
			if pattern != "" && !matchPattern(content.GetName(), pattern) {
				continue
			}
			
			// Create file info
			fileInfo := FileInfo{
				Path:        content.GetPath(),
				Name:        content.GetName(),
				Size:        int64(content.GetSize()),
				IsDirectory: content.GetType() == "dir",
			}
			
			// Get last modified time
			if !fileInfo.IsDirectory {
				lastModified, err := r.getFileLastModified(ctx, content.GetPath())
				if err == nil {
					fileInfo.LastModified = lastModified
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

// GetFile retrieves a file from the GitHub repository
func (r *GitHubRepository) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
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
			fileContent, _, _, err := r.client.Repositories.GetContents(
				ctx,
				r.owner,
				r.repo,
				path,
				&github.RepositoryContentGetOptions{
					Ref: r.config.Branch,
				},
			)
			if err != nil {
				return err
			}
			
			// Check if it's a file
			if fileContent == nil {
				return fmt.Errorf("path is not a file: %s", path)
			}
			
			// Get content
			content, err := fileContent.GetContent()
			if err != nil {
				return err
			}
			
			// Write content to pipe
			_, err = pw.Write([]byte(content))
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

// FileExists checks if a file exists in the GitHub repository
func (r *GitHubRepository) FileExists(ctx context.Context, path string) (bool, error) {
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
		// Get file content
		_, _, resp, err := r.client.Repositories.GetContents(
			ctx,
			r.owner,
			r.repo,
			path,
			&github.RepositoryContentGetOptions{
				Ref: r.config.Branch,
			},
		)
		
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
func (r *GitHubRepository) GetBranch() string {
	return r.config.Branch
}

// GetLastModified gets the last modified time of a file in the GitHub repository
func (r *GitHubRepository) GetLastModified(ctx context.Context, path string) (time.Time, error) {
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
func (r *GitHubRepository) getFileLastModified(ctx context.Context, path string) (time.Time, error) {
	var lastModified time.Time
	
	// Use WithRetry for the operation
	err := r.WithRetry(ctx, func() error {
		// Get commits for the file
		commits, _, err := r.client.Repositories.ListCommits(
			ctx,
			r.owner,
			r.repo,
			&github.CommitsListOptions{
				Path: path,
				SHA:  r.config.Branch,
				ListOptions: github.ListOptions{
					PerPage: 1,
				},
			},
		)
		if err != nil {
			return err
		}
		
		// Check if any commits found
		if len(commits) == 0 {
			return fmt.Errorf("no commits found for file: %s", path)
		}
		
		// Get last commit date
		lastModified = commits[0].Commit.Author.GetDate()
		return nil
	})
	
	if err != nil {
		return time.Time{}, err
	}
	
	return lastModified, nil
}

// matchPattern checks if a string matches a pattern (simple wildcard matching)
func matchPattern(s, pattern string) bool {
	// If pattern is empty or "*", match everything
	if pattern == "" || pattern == "*" {
		return true
	}
	
	// Split pattern by "*"
	parts := strings.Split(pattern, "*")
	
	// If no wildcards, exact match
	if len(parts) == 1 {
		return s == pattern
	}
	
	// Check if string starts with first part
	if parts[0] != "" && !strings.HasPrefix(s, parts[0]) {
		return false
	}
	
	// Check if string ends with last part
	if parts[len(parts)-1] != "" && !strings.HasSuffix(s, parts[len(parts)-1]) {
		return false
	}
	
	// Check middle parts
	current := s
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		if part == "" {
			continue
		}
		
		// Find part in current string
		index := strings.Index(current, part)
		if index == -1 {
			return false
		}
		
		// Move current to after the found part
		current = current[index+len(part):]
	}
	
	return true
}
