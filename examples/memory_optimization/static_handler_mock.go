package main

import (
	"fmt"
	"net/http"
	"sync"
)

// MockFileHandlerOptions represents options for the mock file handler
type MockFileHandlerOptions struct {
	RootDir           string
	EnableCache       bool
	EnableCompression bool
	MaxCacheSize      int64
	CacheExpiration   time.Duration
	MinCompressSize   int64
	CompressExtensions []string
}

// DefaultMockFileHandlerOptions returns default options for the mock file handler
func DefaultMockFileHandlerOptions() *MockFileHandlerOptions {
	return &MockFileHandlerOptions{
		RootDir:           "./static",
		EnableCache:       true,
		EnableCompression: true,
		MaxCacheSize:      100 * 1024 * 1024, // 100MB
		CacheExpiration:   time.Hour,
		MinCompressSize:   1024, // 1KB
		CompressExtensions: []string{
			".html", ".css", ".js", ".json", ".xml", ".txt", ".md",
		},
	}
}

// MockFileHandler is a simplified version of the static file handler for the example
type MockFileHandler struct {
	options *MockFileHandlerOptions
	stats   MockFileHandlerStats
	mutex   sync.RWMutex
}

// NewMockFileHandler creates a new mock file handler
func NewMockFileHandler(options *MockFileHandlerOptions) *MockFileHandler {
	if options == nil {
		options = DefaultMockFileHandlerOptions()
	}
	
	return &MockFileHandler{
		options: options,
		stats: MockFileHandlerStats{
			FilesServed:      0,
			CacheHits:        0,
			CacheMisses:      0,
			CompressedFiles:  0,
			TotalSize:        0,
			CompressedSize:   0,
			CompressionRatio: 0.0,
			AverageServeTime: 5 * time.Millisecond,
		},
	}
}

// ServeHTTP implements the http.Handler interface
func (h *MockFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Update stats
	h.stats.FilesServed++
	
	if h.options.EnableCache {
		h.stats.CacheHits++
	} else {
		h.stats.CacheMisses++
	}
	
	if h.options.EnableCompression {
		h.stats.CompressedFiles++
	}
	
	// Simulate file serving
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("Mock file content for %s", r.URL.Path)))
}

// GetStats returns the current stats
func (h *MockFileHandler) GetStats() MockFileHandlerStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return h.stats
}

// GetCacheSize returns the current cache size
func (h *MockFileHandler) GetCacheSize() int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	// Simulate cache size
	return h.stats.FilesServed * 1024
}

// GetCacheItemCount returns the current number of items in the cache
func (h *MockFileHandler) GetCacheItemCount() int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	// Simulate cache item count
	return h.stats.FilesServed
}

// ClearCache clears the cache
func (h *MockFileHandler) ClearCache() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	fmt.Println("Mock cache cleared")
}

// MockFileHandlerStats represents stats for the mock file handler
type MockFileHandlerStats struct {
	FilesServed      int64
	CacheHits        int64
	CacheMisses      int64
	CompressedFiles  int64
	TotalSize        int64
	CompressedSize   int64
	CompressionRatio float64
	AverageServeTime time.Duration
}
