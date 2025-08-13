package server

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/utils/config"
)

// StaticFileHandler provides optimized handling of static files
// with features like caching, compression, and ETags
type StaticFileHandler struct {
	// rootDir is the root directory for static files
	rootDir string
	// urlPrefix is the URL prefix for static files
	urlPrefix string
	// maxAge is the max-age value for Cache-Control header in seconds
	maxAge int
	// enableCompression enables gzip compression
	enableCompression bool
	// enableETag enables ETag generation
	enableETag bool
	// fileCache caches file information
	fileCache map[string]*fileCacheEntry
	// cacheMutex protects the fileCache
	cacheMutex sync.RWMutex
	// logger is the handler logger
	logger *log.Logger
	// config is the memory configuration
	config *config.MemoryConfig
}

// fileCacheEntry represents a cached file
type fileCacheEntry struct {
	// modTime is the file modification time
	modTime time.Time
	// size is the file size
	size int64
	// etag is the ETag for the file
	etag string
	// contentType is the content type of the file
	contentType string
	// compressible indicates if the file is compressible
	compressible bool
	// data is the cached file data (optional)
	data []byte
}

// StaticFileHandlerOptions contains options for the static file handler
type StaticFileHandlerOptions struct {
	// RootDir is the root directory for static files
	RootDir string
	// URLPrefix is the URL prefix for static files
	URLPrefix string
	// MaxAge is the max-age value for Cache-Control header in seconds
	MaxAge int
	// EnableCompression enables gzip compression
	EnableCompression bool
	// EnableETag enables ETag generation
	EnableETag bool
	// EnableFileCache enables caching of file data in memory
	EnableFileCache bool
	// LogFile is the file to log to
	LogFile string
}

// DefaultStaticFileHandlerOptions returns default options for the static file handler
func DefaultStaticFileHandlerOptions() *StaticFileHandlerOptions {
	return &StaticFileHandlerOptions{
		RootDir:           "static",
		URLPrefix:         "/static/",
		MaxAge:            86400, // 1 day
		EnableCompression: true,
		EnableETag:        true,
		EnableFileCache:   true,
		LogFile:           "logs/static_handler.log",
	}
}

// NewStaticFileHandler creates a new static file handler
func NewStaticFileHandler(options *StaticFileHandlerOptions) (*StaticFileHandler, error) {
	if options == nil {
		options = DefaultStaticFileHandlerOptions()
	}

	// Create logger
	var logger *log.Logger
	if options.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(options.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		logFile, err := os.OpenFile(options.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		// Create logger
		logger = log.New(logFile, "[STATIC] ", log.LstdFlags)
	} else {
		// Log to stderr
		logger = log.New(os.Stderr, "[STATIC] ", log.LstdFlags)
	}

	// Create static file handler
	handler := &StaticFileHandler{
		rootDir:           options.RootDir,
		urlPrefix:         options.URLPrefix,
		maxAge:            options.MaxAge,
		enableCompression: options.EnableCompression,
		enableETag:        options.EnableETag,
		fileCache:         make(map[string]*fileCacheEntry),
		logger:            logger,
		config:            config.GetMemoryConfig(),
	}

	// Create root directory if it doesn't exist
	if err := os.MkdirAll(options.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	// Preload file cache if enabled
	if options.EnableFileCache {
		go handler.preloadFileCache()
	}

	return handler, nil
}

// preloadFileCache preloads the file cache with information about all files
func (h *StaticFileHandler) preloadFileCache() {
	h.logger.Println("Preloading file cache...")

	// Walk the root directory
	err := filepath.Walk(h.rootDir, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(h.rootDir, filePath)
		if err != nil {
			return err
		}

		// Convert to URL path
		urlPath := filepath.ToSlash(relPath)

		// Cache file information
		h.cacheFileInfo(urlPath, filePath, info)

		return nil
	})

	if err != nil {
		h.logger.Printf("Error preloading file cache: %v", err)
	} else {
		h.logger.Printf("File cache preloaded with %d files", len(h.fileCache))
	}
}

// cacheFileInfo caches information about a file
func (h *StaticFileHandler) cacheFileInfo(urlPath, filePath string, info fs.FileInfo) {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()

	// Determine content type
	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Determine if file is compressible
	compressible := isCompressibleContentType(contentType)

	// Create cache entry
	entry := &fileCacheEntry{
		modTime:      info.ModTime(),
		size:         info.Size(),
		contentType:  contentType,
		compressible: compressible,
	}

	// Generate ETag if enabled
	if h.enableETag {
		// Use modification time and size for ETag
		etag := fmt.Sprintf("%x-%x", info.ModTime().Unix(), info.Size())
		entry.etag = etag
	}

	// Add to cache
	h.fileCache[urlPath] = entry
}

// ServeHTTP implements the http.Handler interface
func (h *StaticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get path relative to URL prefix
	if !strings.HasPrefix(r.URL.Path, h.urlPrefix) {
		http.NotFound(w, r)
		return
	}

	urlPath := strings.TrimPrefix(r.URL.Path, h.urlPrefix)
	urlPath = path.Clean(urlPath)

	// Prevent directory traversal
	if strings.Contains(urlPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Get file path
	filePath := filepath.Join(h.rootDir, filepath.FromSlash(urlPath))

	// Check if file exists and get info
	var fileInfo fs.FileInfo
	var err error
	var cacheEntry *fileCacheEntry

	// Check cache first
	h.cacheMutex.RLock()
	cacheEntry = h.fileCache[urlPath]
	h.cacheMutex.RUnlock()

	if cacheEntry != nil {
		// Check if file has been modified
		fileInfo, err = os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				h.logger.Printf("Error stating file %s: %v", filePath, err)
			}
			return
		}

		// Check if file has been modified
		if fileInfo.ModTime().After(cacheEntry.modTime) || fileInfo.Size() != cacheEntry.size {
			// File has been modified, update cache
			h.cacheFileInfo(urlPath, filePath, fileInfo)
			h.cacheMutex.RLock()
			cacheEntry = h.fileCache[urlPath]
			h.cacheMutex.RUnlock()
		}
	} else {
		// Not in cache, get file info
		fileInfo, err = os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				h.logger.Printf("Error stating file %s: %v", filePath, err)
			}
			return
		}

		// Cache file info
		h.cacheFileInfo(urlPath, filePath, fileInfo)
		h.cacheMutex.RLock()
		cacheEntry = h.fileCache[urlPath]
		h.cacheMutex.RUnlock()
	}

	// Check if file is a directory
	if fileInfo.IsDir() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if client has a valid cached copy
	if h.enableETag && cacheEntry.etag != "" {
		if r.Header.Get("If-None-Match") == cacheEntry.etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Check if client has a valid cached copy based on modification time
	if r.Header.Get("If-Modified-Since") != "" {
		ifModifiedSince, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since"))
		if err == nil && cacheEntry.modTime.Unix() <= ifModifiedSince.Unix() {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Set content type
	w.Header().Set("Content-Type", cacheEntry.contentType)

	// Set Last-Modified header
	w.Header().Set("Last-Modified", cacheEntry.modTime.UTC().Format(http.TimeFormat))

	// Set ETag if enabled
	if h.enableETag && cacheEntry.etag != "" {
		w.Header().Set("ETag", cacheEntry.etag)
	}

	// Set Cache-Control header
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", h.maxAge))

	// Check if compression is supported
	var useCompression bool
	if h.enableCompression && cacheEntry.compressible {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		useCompression = strings.Contains(acceptEncoding, "gzip")
	}

	// Serve from cache if data is available
	if cacheEntry.data != nil {
		if useCompression {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			// Compress and serve
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()
			gzipWriter.Write(cacheEntry.data)
		} else {
			// Serve uncompressed
			w.Header().Set("Content-Length", strconv.FormatInt(int64(len(cacheEntry.data)), 10))
			w.Write(cacheEntry.data)
		}
		return
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		h.logger.Printf("Error opening file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	// Serve file with compression if enabled
	if useCompression {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()
		io.Copy(gzipWriter, file)
	} else {
		// Set Content-Length header
		w.Header().Set("Content-Length", strconv.FormatInt(cacheEntry.size, 10))
		io.Copy(w, file)
	}
}

// isCompressibleContentType returns true if the content type is compressible
func isCompressibleContentType(contentType string) bool {
	compressibleTypes := []string{
		"text/",
		"application/javascript",
		"application/json",
		"application/xml",
		"application/xhtml+xml",
		"image/svg+xml",
		"application/wasm",
	}

	for _, t := range compressibleTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}

	return false
}

// GenerateETag generates an ETag for a file
func GenerateETag(filePath string) (string, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create MD5 hash
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Generate ETag
	etag := hex.EncodeToString(hash.Sum(nil))
	return etag, nil
}

// ClearCache clears the file cache
func (h *StaticFileHandler) ClearCache() {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()

	h.fileCache = make(map[string]*fileCacheEntry)
	h.logger.Println("File cache cleared")
}

// RefreshCache refreshes the file cache
func (h *StaticFileHandler) RefreshCache() {
	h.ClearCache()
	go h.preloadFileCache()
}

// GetCacheStats returns statistics about the file cache
func (h *StaticFileHandler) GetCacheStats() map[string]interface{} {
	h.cacheMutex.RLock()
	defer h.cacheMutex.RUnlock()

	stats := make(map[string]interface{})
	stats["cache_size"] = len(h.fileCache)

	var totalSize int64
	var compressibleCount int
	var dataCount int

	for _, entry := range h.fileCache {
		totalSize += entry.size
		if entry.compressible {
			compressibleCount++
		}
		if entry.data != nil {
			dataCount++
		}
	}

	stats["total_size"] = totalSize
	stats["compressible_count"] = compressibleCount
	stats["data_count"] = dataCount

	return stats
}

// RegisterStaticRoute registers the static file handler with an HTTP server
func (h *StaticFileHandler) RegisterStaticRoute(mux *http.ServeMux) {
	mux.Handle(h.urlPrefix, h)
	h.logger.Printf("Registered static file handler for %s", h.urlPrefix)
}

// CacheFile caches a file in memory
func (h *StaticFileHandler) CacheFile(urlPath string) error {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()

	// Get file path
	filePath := filepath.Join(h.rootDir, filepath.FromSlash(urlPath))

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Check if file is a directory
	if fileInfo.IsDir() {
		return fmt.Errorf("cannot cache directory")
	}

	// Get cache entry
	cacheEntry, ok := h.fileCache[urlPath]
	if !ok {
		// Cache file info
		h.cacheFileInfo(urlPath, filePath, fileInfo)
		cacheEntry = h.fileCache[urlPath]
	}

	// Read file data
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Store data in cache
	cacheEntry.data = data

	return nil
}

// UncacheFile removes a file from memory cache
func (h *StaticFileHandler) UncacheFile(urlPath string) {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()

	// Get cache entry
	cacheEntry, ok := h.fileCache[urlPath]
	if !ok {
		return
	}

	// Clear data
	cacheEntry.data = nil
}

// GetFilePath returns the file path for a URL path
func (h *StaticFileHandler) GetFilePath(urlPath string) string {
	return filepath.Join(h.rootDir, filepath.FromSlash(urlPath))
}

// GetURLPath returns the URL path for a file path
func (h *StaticFileHandler) GetURLPath(filePath string) (string, error) {
	// Get relative path
	relPath, err := filepath.Rel(h.rootDir, filePath)
	if err != nil {
		return "", err
	}

	// Convert to URL path
	urlPath := filepath.ToSlash(relPath)

	return path.Join(h.urlPrefix, urlPath), nil
}
