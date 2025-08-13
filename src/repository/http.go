package repository

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HTTPRepository implements the Repository interface for HTTP/HTTPS repositories
type HTTPRepository struct {
	*BaseRepository
	
	// client is the HTTP client
	client *http.Client
	
	// baseURL is the base URL of the repository
	baseURL string
	
	// auditLogger is the audit logger for repository operations
	auditLogger *RepositoryAuditLogger
}

// NewHTTPRepository creates a new HTTP repository
func NewHTTPRepository(config *Config) (Repository, error) {
	// Create base repository
	base := NewBaseRepository(config)
	
	// Validate URL
	_, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP URL: %w", err)
	}
	
	// Ensure URL ends with a slash
	baseURL := config.URL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	
	// Create audit logger if audit logging is enabled
	var auditLogger *RepositoryAuditLogger
	if config.AuditLogger != nil {
		auditLogger = NewRepositoryAuditLogger(config.AuditLogger, "HTTP", baseURL)
	}
	
	return &HTTPRepository{
		BaseRepository: base,
		baseURL:        baseURL,
		auditLogger:    auditLogger,
	}, nil
}

// Connect establishes a connection to the HTTP repository
func (r *HTTPRepository) Connect(ctx context.Context) error {
	// Log repository connection attempt
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryConnect(ctx, r.baseURL)
	}
	
	// Check if already connected
	if r.IsConnected() {
		return nil
	}
	
	// Create transport with custom settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: r.config.InsecureSkipVerify,
		},
		MaxIdleConns:        r.config.MaxConnections,
		MaxIdleConnsPerHost: r.config.MaxConnections,
		IdleConnTimeout:     90 * time.Second,
	}
	
	// Set proxy if configured
	if r.config.ProxyURL != "" {
		proxyURL, err := url.Parse(r.config.ProxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	
	// Create HTTP client
	r.client = &http.Client{
		Transport: transport,
		Timeout:   r.config.Timeout,
	}
	
	// Test connection by sending a HEAD request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, r.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add basic authentication if provided
	if r.config.Username != "" && r.config.Password != "" {
		req.SetBasicAuth(r.config.Username, r.config.Password)
	}
	
	// Send request
	resp, err := r.client.Do(req)
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to connect to HTTP repository: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode >= 400 {
		err := fmt.Errorf("HTTP repository returned status code %d", resp.StatusCode)
		r.setLastError(err)
		return err
	}
	
	// Set connected flag
	r.setConnected(true)
	
	return nil
}

// Disconnect closes the connection to the HTTP repository
func (r *HTTPRepository) Disconnect() error {
	// Log repository disconnection
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryDisconnect(context.Background(), r.baseURL)
	}
	
	// Set connected flag to false
	r.setConnected(false)
	
	// Clear client
	r.client = nil
	
	return nil
}

// ListFiles lists files in the HTTP repository matching the pattern
// Note: HTTP repositories typically don't support directory listing,
// so this implementation is limited and may not work for all HTTP servers
func (r *HTTPRepository) ListFiles(ctx context.Context, pattern string) ([]FileInfo, error) {
	// HTTP repositories typically don't support directory listing
	// This is a simplified implementation that assumes a specific structure
	// or server that supports directory listing
	
	// Log file listing operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryListFiles(ctx, r.baseURL, pattern)
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
		// Create request for directory listing
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.baseURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		
		// Add basic authentication if provided
		if r.config.Username != "" && r.config.Password != "" {
			req.SetBasicAuth(r.config.Username, r.config.Password)
		}
		
		// Send request
		resp, err := r.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		// Check response status
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP repository returned status code %d", resp.StatusCode)
		}
		
		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		
		// Parse HTML response to extract links
		// This is a simplified implementation and may not work for all HTTP servers
		links := extractLinksFromHTML(string(body))
		
		// Process links
		for _, link := range links {
			// Skip parent directory links
			if link == ".." || link == "../" {
				continue
			}
			
			// Skip if not matching pattern
			if pattern != "" && !matchPattern(link, pattern) {
				continue
			}
			
			// Create file info
			isDir := strings.HasSuffix(link, "/")
			name := strings.TrimSuffix(link, "/")
			
			fileInfo := FileInfo{
				Path:        link,
				Name:        name,
				IsDirectory: isDir,
			}
			
			// Get file details if it's a file
			if !fileInfo.IsDirectory {
				// Get file size and last modified time
				fileURL := r.baseURL + link
				fileReq, err := http.NewRequestWithContext(ctx, http.MethodHead, fileURL, nil)
				if err != nil {
					continue
				}
				
				// Add basic authentication if provided
				if r.config.Username != "" && r.config.Password != "" {
					fileReq.SetBasicAuth(r.config.Username, r.config.Password)
				}
				
				// Send request
				fileResp, err := r.client.Do(fileReq)
				if err != nil {
					continue
				}
				fileResp.Body.Close()
				
				// Get file size
				fileInfo.Size = fileResp.ContentLength
				
				// Get last modified time
				lastModified := fileResp.Header.Get("Last-Modified")
				if lastModified != "" {
					t, err := time.Parse(time.RFC1123, lastModified)
					if err == nil {
						fileInfo.LastModified = t
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

// extractLinksFromHTML extracts links from HTML content
// This is a simplified implementation and may not work for all HTML formats
func extractLinksFromHTML(html string) []string {
	var links []string
	
	// Find all href attributes
	hrefStart := `href="`
	for {
		// Find start of href
		startIndex := strings.Index(html, hrefStart)
		if startIndex == -1 {
			break
		}
		
		// Move to start of URL
		startIndex += len(hrefStart)
		
		// Find end of URL
		endIndex := strings.Index(html[startIndex:], `"`)
		if endIndex == -1 {
			break
		}
		
		// Extract URL
		url := html[startIndex : startIndex+endIndex]
		
		// Add to links if it's a relative link
		if !strings.Contains(url, "://") && !strings.HasPrefix(url, "#") {
			links = append(links, url)
		}
		
		// Move past this URL
		html = html[startIndex+endIndex:]
	}
	
	return links
}

// GetFile retrieves a file from the HTTP repository
func (r *HTTPRepository) GetFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	// Log file download start
	if r.auditLogger != nil {
		r.auditLogger.LogFileDownloadStart(ctx, r.baseURL, filePath)
	}
	
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		// Log failure if connection fails
		if r.auditLogger != nil {
			r.auditLogger.LogFileDownloadFailure(ctx, r.baseURL, filePath, err)
		}
		return nil, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return nil, err
	}
	
	// Create URL for the file
	fileURL := r.baseURL + filePath
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		r.ReleaseConnection()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add basic authentication if provided
	if r.config.Username != "" && r.config.Password != "" {
		req.SetBasicAuth(r.config.Username, r.config.Password)
	}
	
	// Send request
	resp, err := r.client.Do(req)
	if err != nil {
		r.ReleaseConnection()
		
		// Log download failure
		if r.auditLogger != nil {
			r.auditLogger.LogFileDownloadFailure(ctx, r.baseURL, filePath, err)
		}
		
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		r.ReleaseConnection()
		
		// Create error for status code
		statusErr := fmt.Errorf("HTTP repository returned status code %d for file %s", resp.StatusCode, filePath)
		
		// Log download failure
		if r.auditLogger != nil {
			r.auditLogger.LogFileDownloadFailure(ctx, r.baseURL, filePath, statusErr)
		}
		
		return nil, statusErr
	}
	
	// Log successful download
	if r.auditLogger != nil {
		// Get content length if available
		var size int64
		if resp.ContentLength > 0 {
			size = resp.ContentLength
		}
		r.auditLogger.LogFileDownloadSuccess(ctx, r.baseURL, filePath, size)
	}
	
	// Create a wrapper for the response body that releases the connection when closed
	return &connectionCloser{
		ReadCloser: resp.Body,
		release: func() {
			r.ReleaseConnection()
		},
		ctx:        ctx,
		auditLogger: r.auditLogger,
		filePath:    filePath,
		baseURL:     r.baseURL,
	}, nil
}

// connectionCloser wraps an io.ReadCloser and calls a release function when closed
type connectionCloser struct {
	io.ReadCloser
	release     func()
	ctx         context.Context
	auditLogger *RepositoryAuditLogger
	filePath    string
	baseURL     string
}

// Close closes the underlying ReadCloser and calls the release function
func (c *connectionCloser) Close() error {
	// Close the underlying ReadCloser
	err := c.ReadCloser.Close()
	
	// Call the release function
	c.release()
	
	// Log download completion if there was no error
	if err == nil && c.auditLogger != nil {
		// We already logged the success at the beginning, but we could add additional logging here if needed
	}
	
	return err
}

// FileExists checks if a file exists in the HTTP repository
func (r *HTTPRepository) FileExists(ctx context.Context, filePath string) (bool, error) {
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		// Log file exists check failure
		if r.auditLogger != nil {
			r.auditLogger.LogFileExistsFailure(ctx, r.baseURL, filePath, err)
		}
		return false, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return false, err
	}
	defer r.ReleaseConnection()
	
	// Create URL for the file
	fileURL := r.baseURL + filePath
	
	// Use WithRetry for the operation
	var exists bool
	err := r.WithRetry(ctx, func() error {
		// Create request
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, fileURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		
		// Add basic authentication if provided
		if r.config.Username != "" && r.config.Password != "" {
			req.SetBasicAuth(r.config.Username, r.config.Password)
		}
		
		// Send request
		resp, err := r.client.Do(req)
		if err != nil {
			// Log file exists check failure
			if r.auditLogger != nil {
				r.auditLogger.LogFileExistsFailure(ctx, r.baseURL, filePath, err)
			}
			return err
		}
		defer resp.Body.Close()
		
		// Check response status
		if resp.StatusCode == http.StatusOK {
			exists = true
			// Log successful file existence check
			if r.auditLogger != nil {
				r.auditLogger.LogFileExists(ctx, r.baseURL, filePath, true)
			}
		} else if resp.StatusCode == http.StatusNotFound {
			exists = false
			// Log file does not exist
			if r.auditLogger != nil {
				r.auditLogger.LogFileExists(ctx, r.baseURL, filePath, false)
			}
		} else {
			// Create error for unexpected status code
			statusErr := fmt.Errorf("HTTP repository returned status code %d for file %s", resp.StatusCode, filePath)
			
			// Log file exists check failure
			if r.auditLogger != nil {
				r.auditLogger.LogFileExistsFailure(ctx, r.baseURL, filePath, statusErr)
			}
			
			return statusErr
		}
		
		return nil
	})
	
	if err != nil {
		return false, err
	}
	
	return exists, nil
}

// GetBranch returns the branch of the repository
// HTTP repositories don't have branches, so this returns an empty string
func (r *HTTPRepository) GetBranch() string {
	return ""
}

// GetLastModified gets the last modified time of a file in the HTTP repository
func (r *HTTPRepository) GetLastModified(ctx context.Context, filePath string) (time.Time, error) {
	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return time.Time{}, err
	}
	
	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return time.Time{}, err
	}
	defer r.ReleaseConnection()
	
	// Create URL for the file
	fileURL := r.baseURL + filePath
	
	// Use WithRetry for the operation
	var lastModified time.Time
	err := r.WithRetry(ctx, func() error {
		// Create request
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, fileURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		
		// Add basic authentication if provided
		if r.config.Username != "" && r.config.Password != "" {
			req.SetBasicAuth(r.config.Username, r.config.Password)
		}
		
		// Send request
		resp, err := r.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		// Check response status
		if resp.StatusCode != http.StatusOK {
			statusErr := fmt.Errorf("HTTP repository returned status code %d for file %s", resp.StatusCode, filePath)
			// Log last modified check failure
			if r.auditLogger != nil {
				r.auditLogger.LogGetLastModifiedFailure(ctx, r.baseURL, filePath, statusErr)
			}
			return statusErr
		}
		
		// Get last modified time
		lastModifiedStr := resp.Header.Get("Last-Modified")
		if lastModifiedStr == "" {
			headerErr := fmt.Errorf("HTTP repository did not return Last-Modified header for file %s", filePath)
			// Log last modified check failure
			if r.auditLogger != nil {
				r.auditLogger.LogGetLastModifiedFailure(ctx, r.baseURL, filePath, headerErr)
			}
			return headerErr
		}
		
		// Parse last modified time
		t, err := time.Parse(time.RFC1123, lastModifiedStr)
		if err != nil {
			parseErr := fmt.Errorf("failed to parse Last-Modified header: %w", err)
			// Log last modified check failure
			if r.auditLogger != nil {
				r.auditLogger.LogGetLastModifiedFailure(ctx, r.baseURL, filePath, parseErr)
			}
			return parseErr
		}
		
		lastModified = t
		return nil
	})
	
	if err != nil {
		// Log last modified check failure
		if r.auditLogger != nil {
			r.auditLogger.LogGetLastModifiedFailure(ctx, r.baseURL, filePath, err)
		}
		return time.Time{}, err
	}
	
	// Log successful last modified check
	if r.auditLogger != nil {
		r.auditLogger.LogGetLastModified(ctx, r.baseURL, filePath, lastModified)
	}
	
	return lastModified, nil
}
