package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
)

// DownloadWithProgressBar downloads a file with a progress bar
func DownloadWithProgressBar(ctx context.Context, url, destPath string) error {
	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create the destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer out.Close()

	// Create progress bar
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintf(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// Copy with progress
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	URL      string
	Path     string
	Size     int64
	Duration time.Duration
	Error    error
}

// BatchDownloader downloads multiple files concurrently with progress
type BatchDownloader struct {
	MaxConcurrent int
	Client        *http.Client
}

// NewBatchDownloader creates a new batch downloader
func NewBatchDownloader(maxConcurrent int) *BatchDownloader {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	
	return &BatchDownloader{
		MaxConcurrent: maxConcurrent,
		Client: &http.Client{
			Timeout: 30 * time.Minute,
		},
	}
}

// Download downloads multiple files concurrently
func (bd *BatchDownloader) Download(ctx context.Context, downloads map[string]string) []DownloadResult {
	results := make([]DownloadResult, 0, len(downloads))
	resultChan := make(chan DownloadResult, len(downloads))
	
	// Create a semaphore for concurrency control
	sem := make(chan struct{}, bd.MaxConcurrent)
	
	// Start downloads
	for url, destPath := range downloads {
		go func(url, destPath string) {
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			start := time.Now()
			err := DownloadWithProgressBar(ctx, url, destPath)
			
			// Get file size
			var size int64
			if err == nil {
				if info, statErr := os.Stat(destPath); statErr == nil {
					size = info.Size()
				}
			}
			
			resultChan <- DownloadResult{
				URL:      url,
				Path:     destPath,
				Size:     size,
				Duration: time.Since(start),
				Error:    err,
			}
		}(url, destPath)
	}
	
	// Collect results
	for i := 0; i < len(downloads); i++ {
		results = append(results, <-resultChan)
	}
	
	return results
}