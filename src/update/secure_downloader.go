// Package update provides functionality for checking and applying updates
package update

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// SecureDownloadOptions represents options for secure downloading
type SecureDownloadOptions struct {
	// Security options for the connection
	SecurityOptions *ConnectionSecurityOptions
	// Progress callback
	ProgressCallback func(totalBytes, downloadedBytes int64, percentage float64)
	// Whether to resume a partial download if possible
	Resume bool
	// Custom HTTP headers
	Headers map[string]string
	// Chunk size for downloading (0 means no chunking)
	ChunkSize int64
	// Verify file after download
	VerifyAfterDownload bool

// DefaultSecureDownloadOptions returns the default secure download options
func DefaultSecureDownloadOptions() *SecureDownloadOptions {
	return &SecureDownloadOptions{
		SecurityOptions:     DefaultConnectionSecurityOptions(),
		Resume:              true,
		Headers:             make(map[string]string),
		ChunkSize:           4 * 1024 * 1024, // 4MB chunks
		VerifyAfterDownload: true,
	}

// SecureDownloader provides secure downloading functionality
type SecureDownloader struct {
	client  *SecureClient
	options *SecureDownloadOptions
	mutex   sync.RWMutex
}

// NewSecureDownloader creates a new SecureDownloader
func NewSecureDownloader(options *SecureDownloadOptions) (*SecureDownloader, error) {
	if options == nil {
		options = DefaultSecureDownloadOptions()
	}

	// Create secure client
	client, err := NewSecureClient(options.SecurityOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure client: %w", err)
	}

	return &SecureDownloader{
		client:  client,
		options: options,
		mutex:   sync.RWMutex{},
	}, nil

// Download securely downloads a file from the given URL to the destination path
func (d *SecureDownloader) Download(ctx context.Context, url, destPath string, options *SecureDownloadOptions) error {
	if options == nil {
		d.mutex.RLock()
		options = d.options
		d.mutex.RUnlock()
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Get file size and check if resume is possible
	fileSize, supportsResume, err := d.getFileInfo(ctx, url, options.Headers)
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
				if options.VerifyAfterDownload {
					return d.verifyDownload(ctx, url, destPath, options.Headers)
				}
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

	// Download the file
	if options.ChunkSize > 0 && fileSize > 0 {
		// Download in chunks
		err = d.downloadInChunks(ctx, url, file, startOffset, fileSize, options)
	} else {
		// Download in a single request
		err = d.downloadWithRange(ctx, url, file, startOffset, fileSize, options)
	}

	if err != nil {
		return err
	}

	// Verify download if requested
	if options.VerifyAfterDownload {
		file.Close() // Close the file before verification
		return d.verifyDownload(ctx, url, destPath, options.Headers)
	}

	return nil
	

// getFileInfo gets information about the file at the given URL
func (d *SecureDownloader) getFileInfo(ctx context.Context, url string, headers map[string]string) (size int64, supportsResume bool, err error) {
	// Make HEAD request to get file info
	resp, err := d.client.Head(ctx, url, headers)
	if err != nil {
		return 0, false, fmt.Errorf("HEAD request failed: %w", err)
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
func (d *SecureDownloader) downloadWithRange(ctx context.Context, url string, file *os.File, startOffset, fileSize int64, options *SecureDownloadOptions) error {
	// Create headers for the request
	headers := make(map[string]string)
	for k, v := range options.Headers {
		headers[k] = v
	}
	// Add range header if starting from an offset
	if startOffset > 0 {
		headers["Range"] = fmt.Sprintf("bytes=%d-", startOffset)
	}

	// Make the request
	resp, err := d.client.Get(ctx, url, headers)
	if err != nil {
		return fmt.Errorf("GET request failed: %w", err)
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Check status code
	if startOffset > 0 && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server doesn't support range requests, got status code: %d", resp.StatusCode)
	} else if startOffset == 0 && resp.StatusCode != http.StatusOK {
	}

	// Get content length for progress reporting
	var contentLength int64
	contentLengthStr := resp.Header.Get("Content-Length")
	if contentLengthStr != "" {
		contentLength, err = strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			// Not fatal, just means we can't report accurate progress
			contentLength = 0
		}
	}

	// If we don't know the file size from the HEAD request, use the content length
	if fileSize == 0 {
		fileSize = contentLength
	}

	// Create buffer for copying
	buf := make([]byte, 32*1024) // 32KB buffer

	// Track progress
	var downloadedBytes int64 = 0
	lastProgressUpdate := time.Now()
	progressInterval := 100 * time.Millisecond

	// Copy data from response to file
	for {
		// Check if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue with download
		}

		// Read data
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// Write data to file
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			// Update progress
			downloadedBytes += int64(n)
			if options.ProgressCallback != nil && time.Since(lastProgressUpdate) >= progressInterval {
				percentage := 0.0
				if fileSize > 0 {
					percentage = float64(startOffset+downloadedBytes) / float64(fileSize) * 100
				}
				options.ProgressCallback(fileSize, startOffset+downloadedBytes, percentage)
				lastProgressUpdate = time.Now()
			}
		}

		// Check for end of file or error
		if err != nil {
			if err == io.EOF {
				// End of file, success
				break
			}
			return fmt.Errorf("failed to read from response: %w", err)
		}
	}

	// Final progress update
	if options.ProgressCallback != nil {
		percentage := 0.0
		if fileSize > 0 {
			percentage = float64(startOffset+downloadedBytes) / float64(fileSize) * 100
		}
		options.ProgressCallback(fileSize, startOffset+downloadedBytes, percentage)
	}

	return nil

// downloadInChunks downloads a file in chunks for better reliability
func (d *SecureDownloader) downloadInChunks(ctx context.Context, url string, file *os.File, startOffset, fileSize int64, options *SecureDownloadOptions) error {
	// Calculate chunks
	chunkSize := options.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 4 * 1024 * 1024 // Default to 4MB chunks
	}

	// Track total downloaded bytes for progress reporting
	var totalDownloaded int64 = 0
	lastProgressUpdate := time.Now()
	progressInterval := 100 * time.Millisecond

	// Download each chunk
	for offset := startOffset; offset < fileSize; offset += chunkSize {
		// Check if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue with download
		}
		// Calculate end of this chunk
		end := offset + chunkSize - 1
		if end >= fileSize {
			end = fileSize - 1
		}

		// Create headers for the request
		headers := make(map[string]string)
		for k, v := range options.Headers {
			headers[k] = v
		}
		headers["Range"] = fmt.Sprintf("bytes=%d-%d", offset, end)

		// Make the request
		resp, err := d.client.Get(ctx, url, headers)
		if err != nil {
			return fmt.Errorf("GET request failed for chunk %d-%d: %w", offset, end, err)
		}

		// Check status code
		if resp.StatusCode != http.StatusPartialContent {
			resp.Body.Close()
			return fmt.Errorf("server doesn't support range requests for chunk %d-%d, got status code: %d", offset, end, resp.StatusCode)
		}

		// Seek to the correct position in the file
		_, err = file.Seek(offset, io.SeekStart)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to seek in file: %w", err)
		}

		// Copy data from response to file
		var chunkDownloaded int64 = 0
		buf := make([]byte, 32*1024) // 32KB buffer

		for {
			// Check if context is canceled
			select {
			case <-ctx.Done():
				resp.Body.Close()
				return ctx.Err()
			default:
				// Continue with download
			}

			// Read data
			n, err := resp.Body.Read(buf)
			if n > 0 {
				// Write data to file
				_, writeErr := file.Write(buf[:n])
				if writeErr != nil {
					resp.Body.Close()
					return fmt.Errorf("failed to write to file: %w", writeErr)
				}

				// Update progress
				chunkDownloaded += int64(n)
				totalDownloaded += int64(n)

				// Report progress
				if options.ProgressCallback != nil && time.Since(lastProgressUpdate) >= progressInterval {
					percentage := float64(offset+chunkDownloaded) / float64(fileSize) * 100
					options.ProgressCallback(fileSize, offset+chunkDownloaded, percentage)
					lastProgressUpdate = time.Now()
				}
			}

			// Check for end of chunk or error
			if err != nil {
				if err == io.EOF {
					// End of chunk, success
					break
				}
				resp.Body.Close()
				return fmt.Errorf("failed to read from response: %w", err)
			}
		}
		resp.Body.Close()
	}

	// Final progress update
	if options.ProgressCallback != nil {
		options.ProgressCallback(fileSize, fileSize, 100.0)
	}

	return nil

// verifyDownload verifies the integrity of a downloaded file
func (d *SecureDownloader) verifyDownload(ctx context.Context, url, filePath string, headers map[string]string) error {
	// Get file size from server
	serverSize, _, err := d.getFileInfo(ctx, url, headers)
	if err != nil {
		return fmt.Errorf("failed to get file info for verification: %w", err)
	}

	// Get local file size
	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	// Compare sizes
	if stat.Size() != serverSize {
		return fmt.Errorf("file size mismatch: expected %d bytes, got %d bytes", serverSize, stat.Size())
	}

	return nil

// DownloadWithSecureProgress downloads a file with progress reporting using the secure downloader
func DownloadWithSecureProgress(ctx context.Context, url, destPath string, pinnedCerts []PinnedCertificate) error {
	// Create options with progress callback
	options := DefaultSecureDownloadOptions()
	options.ProgressCallback = func(totalBytes, downloadedBytes int64, percentage float64) {
		// Only update progress every 2% to avoid console spam
		if int(percentage)%2 == 0 {
			fmt.Printf("\rDownloading: %.1f%% (%d/%d bytes)", percentage, downloadedBytes, totalBytes)
		}
	}

	// Add certificate pinning if provided
	if len(pinnedCerts) > 0 {
		options.SecurityOptions.EnableCertificatePinning = true
		options.SecurityOptions.PinnedCertificates = pinnedCerts
	}

	// Create downloader
	downloader, err := NewSecureDownloader(options)
	if err != nil {
		return err
	}

	// Download file
	err = downloader.Download(ctx, url, destPath, options)
	if err != nil {
		fmt.Println() // End the progress line
		return err
	}

	// Complete progress
	fmt.Println("\rDownload complete.                                    ")
	return nil
}
}
}
}
}
}
}
