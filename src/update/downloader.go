package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// UpdateDownloader handles downloading files for updates
type UpdateDownloader struct {
	config *Config
	client *http.Client
	logger Logger
}

// DownloadProgress represents download progress
type DownloadProgress struct {
	URL           string
	Filename      string
	TotalBytes    int64
	DownloadedBytes int64
	Speed         float64
	ETA           time.Duration
	StartTime     time.Time
}

// ProgressCallback is called during download progress
type ProgressCallback func(*DownloadProgress)

// NewUpdateDownloader creates a new update downloader
func NewUpdateDownloader(config *Config, logger Logger) *UpdateDownloader {
	client := &http.Client{
		Timeout: config.Timeout,
	}
	
	// Configure proxy if specified
	if config.ProxyURL != "" {
		if proxyURL, err := url.Parse(config.ProxyURL); err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}
	
	return &UpdateDownloader{
		config: config,
		client: client,
		logger: logger,
	}
}

// DownloadFile downloads a file from the given URL
func (d *UpdateDownloader) DownloadFile(ctx context.Context, url, filename string) (string, error) {
	return d.DownloadFileWithProgress(ctx, url, filename, nil)
}

// DownloadFileWithProgress downloads a file with progress callback
func (d *UpdateDownloader) DownloadFileWithProgress(ctx context.Context, url, filename string, progressCallback ProgressCallback) (string, error) {
	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), "LLMrecon-updates")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Create destination file path
	destPath := filepath.Join(tempDir, filename)
	
	d.logger.Info(fmt.Sprintf("Downloading %s to %s", url, destPath))
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", d.config.UserAgent)
	
	var attempt int
	for attempt < d.config.MaxRetries {
		if attempt > 0 {
			d.logger.Info(fmt.Sprintf("Retry attempt %d/%d", attempt+1, d.config.MaxRetries))
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		
		if err := d.downloadWithRetry(ctx, req, destPath, progressCallback); err != nil {
			d.logger.Error(fmt.Sprintf("Download attempt %d failed", attempt+1), err)
			attempt++
			continue
		}
		
		break
	}
	
	if attempt >= d.config.MaxRetries {
		return "", fmt.Errorf("download failed after %d attempts", d.config.MaxRetries)
	}
	
	d.logger.Info(fmt.Sprintf("Successfully downloaded %s", filename))
	return destPath, nil
}

// downloadWithRetry performs a single download attempt
func (d *UpdateDownloader) downloadWithRetry(ctx context.Context, req *http.Request, destPath string, progressCallback ProgressCallback) error {
	// Execute request
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	
	// Get content length
	contentLength := resp.ContentLength
	if contentLengthStr := resp.Header.Get("Content-Length"); contentLengthStr != "" {
		if cl, err := strconv.ParseInt(contentLengthStr, 10, 64); err == nil {
			contentLength = cl
		}
	}
	
	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	// Setup progress tracking
	progress := &DownloadProgress{
		URL:           req.URL.String(),
		Filename:      filepath.Base(destPath),
		TotalBytes:    contentLength,
		DownloadedBytes: 0,
		StartTime:     time.Now(),
	}
	
	// Create progress reader
	reader := &progressReader{
		reader:   resp.Body,
		progress: progress,
		callback: progressCallback,
		logger:   d.logger,
	}
	
	// Copy with progress
	_, err = io.Copy(destFile, reader)
	if err != nil {
		os.Remove(destPath) // Clean up on error
		return fmt.Errorf("failed to download file: %w", err)
	}
	
	// Final progress callback
	if progressCallback != nil {
		progress.DownloadedBytes = progress.TotalBytes
		progressCallback(progress)
	}
	
	return nil
}

// DownloadFileToPath downloads a file to a specific path
func (d *UpdateDownloader) DownloadFileToPath(ctx context.Context, url, destPath string, progressCallback ProgressCallback) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", d.config.UserAgent)
	
	return d.downloadWithRetry(ctx, req, destPath, progressCallback)
}

// DownloadArchive downloads and extracts an archive
func (d *UpdateDownloader) DownloadArchive(ctx context.Context, url, destDir string, progressCallback ProgressCallback) error {
	// Download archive to temporary file
	tempFile, err := d.DownloadFileWithProgress(ctx, url, "archive.zip", progressCallback)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer os.Remove(tempFile)
	
	// Extract archive
	return d.extractArchive(tempFile, destDir)
}

// extractArchive extracts an archive file
func (d *UpdateDownloader) extractArchive(archivePath, destDir string) error {
	// Determine archive type from extension
	ext := strings.ToLower(filepath.Ext(archivePath))
	
	switch ext {
	case ".zip":
		return d.extractZip(archivePath, destDir)
	case ".tar", ".tgz", ".tar.gz":
		return d.extractTar(archivePath, destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// extractZip extracts a ZIP archive
func (d *UpdateDownloader) extractZip(archivePath, destDir string) error {
	// Implementation would use archive/zip package
	d.logger.Info(fmt.Sprintf("Extracting ZIP archive %s to %s", archivePath, destDir))
	return fmt.Errorf("ZIP extraction not yet implemented")
}

// extractTar extracts a TAR archive
func (d *UpdateDownloader) extractTar(archivePath, destDir string) error {
	// Implementation would use archive/tar package
	d.logger.Info(fmt.Sprintf("Extracting TAR archive %s to %s", archivePath, destDir))
	return fmt.Errorf("TAR extraction not yet implemented")
}

// GetFileSize gets the size of a remote file without downloading
func (d *UpdateDownloader) GetFileSize(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", d.config.UserAgent)
	
	resp, err := d.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	
	contentLength := resp.ContentLength
	if contentLengthStr := resp.Header.Get("Content-Length"); contentLengthStr != "" {
		if cl, err := strconv.ParseInt(contentLengthStr, 10, 64); err == nil {
			contentLength = cl
		}
	}
	
	return contentLength, nil
}

// VerifyFileSize verifies that a downloaded file has the expected size
func (d *UpdateDownloader) VerifyFileSize(filePath string, expectedSize int64) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	
	if info.Size() != expectedSize {
		return fmt.Errorf("file size mismatch: expected %d, got %d", expectedSize, info.Size())
	}
	
	return nil
}

// CleanupTempFiles removes temporary download files
func (d *UpdateDownloader) CleanupTempFiles() error {
	tempDir := filepath.Join(os.TempDir(), "LLMrecon-updates")
	if err := os.RemoveAll(tempDir); err != nil {
		return fmt.Errorf("failed to cleanup temp directory: %w", err)
	}
	
	d.logger.Info("Cleaned up temporary download files")
	return nil
}

// progressReader wraps an io.Reader to track download progress
type progressReader struct {
	reader     io.Reader
	progress   *DownloadProgress
	callback   ProgressCallback
	logger     Logger
	lastUpdate time.Time
}

// Read implements io.Reader
func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	
	if n > 0 {
		pr.progress.DownloadedBytes += int64(n)
		
		// Calculate speed and ETA
		elapsed := time.Since(pr.progress.StartTime)
		if elapsed > 0 {
			pr.progress.Speed = float64(pr.progress.DownloadedBytes) / elapsed.Seconds()
			
			if pr.progress.TotalBytes > 0 && pr.progress.Speed > 0 {
				remaining := pr.progress.TotalBytes - pr.progress.DownloadedBytes
				pr.progress.ETA = time.Duration(float64(remaining)/pr.progress.Speed) * time.Second
			}
		}
		
		// Call progress callback (throttled to avoid spam)
		now := time.Now()
		if pr.callback != nil && (now.Sub(pr.lastUpdate) > time.Second || err != nil) {
			pr.callback(pr.progress)
			pr.lastUpdate = now
		}
	}
	
	return n, err
}

// DownloadStats represents download statistics
type DownloadStats struct {
	TotalDownloads   int
	TotalBytes       int64
	SuccessfulDownloads int
	FailedDownloads  int
	AverageSpeed     float64
	TotalTime        time.Duration
}

// GetDownloadStats returns download statistics
func (d *UpdateDownloader) GetDownloadStats() *DownloadStats {
	// Implementation would track statistics
	return &DownloadStats{
		TotalDownloads:   0,
		TotalBytes:       0,
		SuccessfulDownloads: 0,
		FailedDownloads:  0,
		AverageSpeed:     0,
		TotalTime:        0,
	}
}

