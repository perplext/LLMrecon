// Package update provides functionality for checking and applying updates
package update

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// DownloadOptions represents options for downloading files
type DownloadOptions struct {
	// Whether to verify TLS certificates
	VerifyCertificate bool
	// Number of retry attempts
	RetryAttempts int
	// Delay between retries (exponential backoff will be applied)
	RetryDelay time.Duration
	// Timeout for the entire download
	Timeout time.Duration
	// Progress callback
	ProgressCallback func(totalBytes, downloadedBytes int64, percentage float64)
	// Whether to resume a partial download if possible
	Resume bool
	// Custom HTTP headers
	Headers map[string]string

// DefaultDownloadOptions returns the default download options
func DefaultDownloadOptions() *DownloadOptions {
	return &DownloadOptions{
		VerifyCertificate: true,
		RetryAttempts:     3,
		RetryDelay:        2 * time.Second,
		Timeout:           10 * time.Minute,
		Resume:            true,
		Headers:           make(map[string]string),
	}

// Downloader handles secure downloading of files
type Downloader struct {
	client *http.Client
	mutex  sync.Mutex

// NewDownloader creates a new Downloader
func NewDownloader(options *DownloadOptions) *Downloader {
	if options == nil {
		options = DefaultDownloadOptions()
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !options.VerifyCertificate,
		},
		// Set reasonable defaults for production use
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   options.Timeout,
	}

	return &Downloader{
		client: client,
		mutex:  sync.Mutex{},
	}

// Download downloads a file from the given URL to the destination path
func (d *Downloader) Download(ctx context.Context, url, destPath string, options *DownloadOptions) error {
	if options == nil {
		options = DefaultDownloadOptions()
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Get file size and check if resume is possible
	fileSize, supportsResume, err := d.getFileInfo(url, options.Headers)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Check if file already exists
	var file *os.File
	var startOffset int64 = 0
	if options.Resume && supportsResume {
		// Try to open existing file for appending
		if stat, err := os.Stat(destPath); err == nil {
			startOffset = stat.Size()
			if startOffset < fileSize {
				file, err = os.OpenFile(destPath, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					// If we can't open for appending, start from scratch
					startOffset = 0
				}
			} else if startOffset == fileSize {
				// File is already complete
				return nil
			}
		}
	}

	// If file wasn't opened for appending, create a new one
	if file == nil {
		file, err = os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		startOffset = 0
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Perform the download with retries
	var lastErr error
	for attempt := 0; attempt <= options.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			retryDelay := options.RetryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-time.After(retryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}

			// Check if file size has changed
			newFileSize, _, err := d.getFileInfo(url, options.Headers)
			if err == nil && newFileSize != fileSize {
				// File has changed on the server, start over
				file.Close()
				file, err = os.Create(destPath)
				if err != nil {
					return fmt.Errorf("failed to recreate file: %w", err)
				}
				startOffset = 0
				fileSize = newFileSize
			}
		}

		err := d.downloadWithRange(ctx, url, file, startOffset, fileSize, options)
		if err == nil {
			// Download successful
			return nil
		}

		lastErr = err
		// If context is canceled, don't retry
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", options.RetryAttempts, lastErr)

// getFileInfo gets information about the file at the given URL
func (d *Downloader) getFileInfo(url string, headers map[string]string) (size int64, supportsResume bool, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, false, err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	if resp.StatusCode != http.StatusOK {
	}

	// Check if server supports range requests
	supportsResume = resp.Header.Get("Accept-Ranges") == "bytes"

	// Get content length
	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		// If Content-Length is not provided, we can't determine file size
		return 0, supportsResume, nil
	}

	size, err = strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, supportsResume, fmt.Errorf("invalid Content-Length: %w", err)
	}

	return size, supportsResume, nil

// downloadWithRange downloads a file with range requests
func (d *Downloader) downloadWithRange(ctx context.Context, url string, file *os.File, startOffset, fileSize int64, options *DownloadOptions) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// Add headers
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	// Set range header if resuming
	if startOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startOffset))
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	if startOffset > 0 && resp.StatusCode != http.StatusPartialContent {
		// Server doesn't support range requests or range is invalid
		// Start over from the beginning
		file.Close()
		file, err = os.Create(file.Name())
		if err != nil {
			return fmt.Errorf("failed to recreate file: %w", err)
		}
		startOffset = 0
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
	}

	// Create a buffer for copying
	buf := make([]byte, 32*1024) // 32KB buffer

	// Track progress
	downloadedBytes := startOffset
	lastProgressUpdate := time.Now()

	for {
		// Check if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read a chunk from the response body
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// Write the chunk to the file
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			// Update progress
			downloadedBytes += int64(n)
			if options.ProgressCallback != nil && time.Since(lastProgressUpdate) > 100*time.Millisecond {
				percentage := 0.0
				if fileSize > 0 {
					percentage = float64(downloadedBytes) / float64(fileSize) * 100
				}
				options.ProgressCallback(fileSize, downloadedBytes, percentage)
				lastProgressUpdate = time.Now()
			}
		}

		if err != nil {
			if err == io.EOF {
				// End of file, download complete
				break
			}
			return fmt.Errorf("error reading response body: %w", err)
		}
	}

	// Final progress update
	if options.ProgressCallback != nil {
		percentage := 0.0
		if fileSize > 0 {
			percentage = float64(downloadedBytes) / float64(fileSize) * 100
		}
		options.ProgressCallback(fileSize, downloadedBytes, percentage)
	}

	return nil

// DownloadWithProgress downloads a file with progress reporting to stdout
func DownloadWithProgress(ctx context.Context, url, destPath string) error {
	options := DefaultDownloadOptions()
	options.ProgressCallback = func(totalBytes, downloadedBytes int64, percentage float64) {
		// Only print progress every 5%
		if int(percentage)%5 == 0 {
			if totalBytes > 0 {
				fmt.Printf("\rDownloading: %.1f%% (%d/%d bytes)", percentage, downloadedBytes, totalBytes)
			} else {
				fmt.Printf("\rDownloading: %d bytes", downloadedBytes)
			}
		}
	}

	downloader := NewDownloader(options)
	err := downloader.Download(ctx, url, destPath, options)
	if err == nil {
		fmt.Println("\rDownload complete.                                ")
	} else {
		fmt.Println("\rDownload failed.                                  ")
	}
