package static

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/utils/config"
	"github.com/perplext/LLMrecon/src/utils/monitoring"
)

// FileHandlerOptions contains options for the static file handler
type FileHandlerOptions struct {
	// Root directory for static files
	RootDir string
	// Whether to enable file caching
	EnableCache bool
	// Maximum cache size in bytes
	MaxCacheSize int64
	// Whether to enable gzip compression
	EnableCompression bool
	// Minimum file size for compression in bytes
	MinCompressSize int64
	// Cache expiration time
	CacheExpiration time.Duration
	// File extensions to compress
	CompressExtensions []string
}

// DefaultFileHandlerOptions returns default options for the file handler
func DefaultFileHandlerOptions() *FileHandlerOptions {
	return &FileHandlerOptions{
		RootDir:            "static",
		EnableCache:        true,
		MaxCacheSize:       100 * 1024 * 1024, // 100 MB
		EnableCompression:  true,
		MinCompressSize:    1024, // 1 KB
		CacheExpiration:    1 * time.Hour,
		CompressExtensions: []string{".html", ".css", ".js", ".json", ".xml", ".txt", ".md"},
	}
}

// CachedFile represents a cached static file
type CachedFile struct {
	Content       []byte
	CompressedContent []byte
	ContentType   string
	LastModified  time.Time
	ETag          string
	Expiration    time.Time
}

// FileHandler handles static file serving with memory optimization
type FileHandler struct {
	options *FileHandlerOptions
	cache   map[string]*CachedFile
	mutex   sync.RWMutex
	cacheSize int64
}

// NewFileHandler creates a new static file handler
func NewFileHandler(options *FileHandlerOptions) *FileHandler {
	if options == nil {
		options = DefaultFileHandlerOptions()
	}

	return &FileHandler{
		options:   options,
		cache:     make(map[string]*CachedFile),
		mutex:     sync.RWMutex{},
		cacheSize: 0,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path to prevent directory traversal attacks
	path := filepath.Clean(r.URL.Path)
	if path == "/" {
		path = "/index.html"
	}

	// Get the absolute file path
	filePath := filepath.Join(h.options.RootDir, path)

	// Try to serve from cache if enabled
	if h.options.EnableCache {
		if cachedFile := h.getFromCache(filePath); cachedFile != nil {
			// Check if the client has a valid cached version
			if h.checkClientCache(w, r, cachedFile) {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			// Check if we can serve compressed content
			if h.options.EnableCompression && len(cachedFile.CompressedContent) > 0 && h.acceptsGzip(r) {
				h.serveCompressedContent(w, r, cachedFile)
				return
			}

			// Serve uncompressed content
			h.serveContent(w, r, cachedFile)
			return
		}
	}

	// File not in cache, serve directly from disk
	h.serveFromDisk(w, r, filePath)
}

// getFromCache retrieves a file from the cache
func (h *FileHandler) getFromCache(filePath string) *CachedFile {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	cachedFile, exists := h.cache[filePath]
	if !exists {
		return nil
	}

	// Check if the cache has expired
	if time.Now().After(cachedFile.Expiration) {
		return nil
	}

	return cachedFile
}

// checkClientCache checks if the client has a valid cached version
func (h *FileHandler) checkClientCache(w http.ResponseWriter, r *http.Request, cachedFile *CachedFile) bool {
	// Set cache control headers
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("ETag", cachedFile.ETag)
	w.Header().Set("Last-Modified", cachedFile.LastModified.Format(http.TimeFormat))

	// Check If-None-Match header
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == cachedFile.ETag {
			return true
		}
	}

	// Check If-Modified-Since header
	if modifiedSince := r.Header.Get("If-Modified-Since"); modifiedSince != "" {
		if t, err := time.Parse(http.TimeFormat, modifiedSince); err == nil {
			if cachedFile.LastModified.Before(t.Add(1 * time.Second)) {
				return true
			}
		}
	}

	return false
}

// serveContent serves uncompressed content
func (h *FileHandler) serveContent(w http.ResponseWriter, r *http.Request, cachedFile *CachedFile) {
	w.Header().Set("Content-Type", cachedFile.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(cachedFile.Content)))
	w.WriteHeader(http.StatusOK)
	w.Write(cachedFile.Content)
}

// serveCompressedContent serves gzip compressed content
func (h *FileHandler) serveCompressedContent(w http.ResponseWriter, r *http.Request, cachedFile *CachedFile) {
	w.Header().Set("Content-Type", cachedFile.ContentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(cachedFile.CompressedContent)))
	w.WriteHeader(http.StatusOK)
	w.Write(cachedFile.CompressedContent)
}

// acceptsGzip checks if the client accepts gzip encoding
func (h *FileHandler) acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

// serveFromDisk serves a file directly from disk
func (h *FileHandler) serveFromDisk(w http.ResponseWriter, r *http.Request, filePath string) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Don't serve directories
	if info.IsDir() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Determine content type
	contentType := getContentType(filePath)

	// Cache the file if caching is enabled
	if h.options.EnableCache {
		h.cacheFile(filePath, file, info, contentType)
		
		// Try again from cache
		if cachedFile := h.getFromCache(filePath); cachedFile != nil {
			if h.checkClientCache(w, r, cachedFile) {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			if h.options.EnableCompression && len(cachedFile.CompressedContent) > 0 && h.acceptsGzip(r) {
				h.serveCompressedContent(w, r, cachedFile)
				return
			}

			h.serveContent(w, r, cachedFile)
			return
		}
	}

	// If we get here, serve the file directly
	http.ServeFile(w, r, filePath)
}

// cacheFile caches a file
func (h *FileHandler) cacheFile(filePath string, file *os.File, info os.FileInfo, contentType string) {
	// Don't cache files that are too large
	if info.Size() > h.options.MaxCacheSize {
		return
	}

	// Read the file content
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	// Reset file position
	file.Seek(0, 0)

	// Calculate ETag
	hash := md5.Sum(content)
	etag := hex.EncodeToString(hash[:])

	// Create cached file
	cachedFile := &CachedFile{
		Content:      content,
		ContentType:  contentType,
		LastModified: info.ModTime(),
		ETag:         etag,
		Expiration:   time.Now().Add(h.options.CacheExpiration),
	}

	// Compress the file if needed
	if h.options.EnableCompression && info.Size() >= h.options.MinCompressSize && shouldCompress(filePath, h.options.CompressExtensions) {
		cachedFile.CompressedContent = compressContent(content)
	}

	// Add to cache
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check if we need to evict some files to make room
	newSize := h.cacheSize + info.Size()
	if newSize > h.options.MaxCacheSize {
		h.evictCache(newSize - h.options.MaxCacheSize)
	}

	h.cache[filePath] = cachedFile
	h.cacheSize += info.Size()
}

// evictCache evicts files from the cache to free up space
func (h *FileHandler) evictCache(bytesToFree int64) {
	// Simple LRU eviction - remove oldest files first
	type cacheEntry struct {
		path       string
		expiration time.Time
		size       int64
	}

	entries := make([]cacheEntry, 0, len(h.cache))
	for path, file := range h.cache {
		entries = append(entries, cacheEntry{
			path:       path,
			expiration: file.Expiration,
			size:       int64(len(file.Content)),
		})
	}

	// Sort by expiration time (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].expiration.Before(entries[j].expiration)
	})

	// Remove entries until we've freed enough space
	var freedBytes int64
	for _, entry := range entries {
		if freedBytes >= bytesToFree {
			break
		}

		delete(h.cache, entry.path)
		freedBytes += entry.size
		h.cacheSize -= entry.size
	}
}

// getContentType determines the content type of a file
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".xml":
		return "application/xml; charset=utf-8"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".md":
		return "text/markdown; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// shouldCompress checks if a file should be compressed
func shouldCompress(filePath string, compressExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, compressExt := range compressExtensions {
		if ext == compressExt {
			return true
		}
	}
	return false
}

// compressContent compresses content using gzip
func compressContent(content []byte) []byte {
	var b bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
	gz.Write(content)
	gz.Close()
	return b.Bytes()
}

// ClearCache clears the file cache
func (h *FileHandler) ClearCache() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.cache = make(map[string]*CachedFile)
	h.cacheSize = 0
}

// GetCacheSize returns the current cache size in bytes
func (h *FileHandler) GetCacheSize() int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.cacheSize
}

// GetCacheItemCount returns the number of items in the cache
func (h *FileHandler) GetCacheItemCount() int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return int64(len(h.cache))
}

// LoadFromConfig loads file handler options from configuration
func LoadFromConfig(cfg *config.MemoryConfig) *FileHandlerOptions {
	options := DefaultFileHandlerOptions()
	
	if cfg.StaticFileHandler != nil {
		if cfg.StaticFileHandler.RootDir != "" {
			options.RootDir = cfg.StaticFileHandler.RootDir
		}
		
		if cfg.StaticFileHandler.EnableCache != nil {
			options.EnableCache = *cfg.StaticFileHandler.EnableCache
		}
		
		if cfg.StaticFileHandler.MaxCacheSize > 0 {
			options.MaxCacheSize = cfg.StaticFileHandler.MaxCacheSize
		}
		
		if cfg.StaticFileHandler.EnableCompression != nil {
			options.EnableCompression = *cfg.StaticFileHandler.EnableCompression
		}
		
		if cfg.StaticFileHandler.MinCompressSize > 0 {
			options.MinCompressSize = cfg.StaticFileHandler.MinCompressSize
		}
		
		if cfg.StaticFileHandler.CacheExpirationSeconds > 0 {
			options.CacheExpiration = time.Duration(cfg.StaticFileHandler.CacheExpirationSeconds) * time.Second
		}
		
		if len(cfg.StaticFileHandler.CompressExtensions) > 0 {
			options.CompressExtensions = cfg.StaticFileHandler.CompressExtensions
		}
	}
	
	return options
}

// FileResponse represents a response from the file handler
type FileResponse struct {
	Content        []byte
	ContentType    string
	ContentLength  int
	StatusCode     int
	ETag           string
	LastModified   time.Time
	Compressed     bool
	FromCache      bool
}

// HandlerStats contains statistics for the file handler
type HandlerStats struct {
	FilesServed      int64
	CacheHits        int64
	CacheMisses      int64
	CompressedFiles  int64
	TotalSize        int64
	CompressedSize   int64
	CompressionRatio float64
	AverageServeTime time.Duration
	TotalServeTime   time.Duration
}

// Stats for the file handler
var stats = HandlerStats{}
var statsMutex sync.RWMutex

// ServeFile serves a file with the given name and content
// This method is primarily for benchmarking and testing
func (h *FileHandler) ServeFile(fileName string, content []byte) *FileResponse {
	startTime := time.Now()
	
	// Update stats
	statsMutex.Lock()
	stats.FilesServed++
	stats.TotalSize += int64(len(content))
	statsMutex.Unlock()
	
	// Determine content type
	contentType := getContentType(fileName)
	
	// Calculate ETag
	hash := md5.Sum(content)
	etag := hex.EncodeToString(hash[:])
	
	// Create response
	response := &FileResponse{
		Content:       content,
		ContentType:   contentType,
		ContentLength: len(content),
		StatusCode:    http.StatusOK,
		ETag:          etag,
		LastModified:  time.Now(),
		Compressed:    false,
		FromCache:     false,
	}
	
	// Check if file is in cache
	if h.options.EnableCache {
		cachedFile := h.getFromCache(fileName)
		if cachedFile != nil {
			// Update stats
			statsMutex.Lock()
			stats.CacheHits++
			statsMutex.Unlock()
			
			response.FromCache = true
			response.Content = cachedFile.Content
			response.ETag = cachedFile.ETag
			response.LastModified = cachedFile.LastModified
		} else {
			// Update stats
			statsMutex.Lock()
			stats.CacheMisses++
			statsMutex.Unlock()
			
			// Cache the file
			h.cacheFileContent(fileName, content, contentType)
		}
	}
	
	// Compress the file if needed
	if h.options.EnableCompression && len(content) >= int(h.options.MinCompressSize) && 
	   shouldCompress(fileName, h.options.CompressExtensions) {
		compressedContent := compressContent(content)
		
		// Update stats
		statsMutex.Lock()
		stats.CompressedFiles++
		stats.CompressedSize += int64(len(compressedContent))
		statsMutex.Unlock()
		
		response.Content = compressedContent
		response.Compressed = true
		response.ContentLength = len(compressedContent)
	}
	
	// Update serve time stats
	serveTime := time.Since(startTime)
	statsMutex.Lock()
	stats.TotalServeTime += serveTime
	stats.AverageServeTime = stats.TotalServeTime / time.Duration(stats.FilesServed)
	statsMutex.Unlock()
	
	return response
}

// cacheFileContent caches file content
func (h *FileHandler) cacheFileContent(fileName string, content []byte, contentType string) {
	// Don't cache files that are too large
	if int64(len(content)) > h.options.MaxCacheSize {
		return
	}
	
	// Calculate ETag
	hash := md5.Sum(content)
	etag := hex.EncodeToString(hash[:])
	
	// Create cached file
	cachedFile := &CachedFile{
		Content:      content,
		ContentType:  contentType,
		LastModified: time.Now(),
		ETag:         etag,
		Expiration:   time.Now().Add(h.options.CacheExpiration),
	}
	
	// Compress the file if needed
	if h.options.EnableCompression && int64(len(content)) >= h.options.MinCompressSize && 
	   shouldCompress(fileName, h.options.CompressExtensions) {
		cachedFile.CompressedContent = compressContent(content)
	}
	
	// Add to cache
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Check if we need to evict some files to make room
	newSize := h.cacheSize + int64(len(content))
	if newSize > h.options.MaxCacheSize {
		h.evictCache(newSize - h.options.MaxCacheSize)
	}
	
	h.cache[fileName] = cachedFile
	h.cacheSize += int64(len(content))
}

// GetStats returns statistics for the file handler
func (h *FileHandler) GetStats() *monitoring.Stats {
	statsMutex.RLock()
	defer statsMutex.RUnlock()
	
	// Calculate compression ratio
	var compressionRatio float64
	if stats.TotalSize > 0 && stats.CompressedSize > 0 {
		compressionRatio = 1.0 - (float64(stats.CompressedSize) / float64(stats.TotalSize))
	}
	
	// Return a copy of the stats as monitoring.Stats
	return &monitoring.Stats{
		FilesServed:      stats.FilesServed,
		CacheHits:        stats.CacheHits,
		CacheMisses:      stats.CacheMisses,
		CompressedFiles:  stats.CompressedFiles,
		TotalSize:        stats.TotalSize,
		CompressedSize:   stats.CompressedSize,
		CompressionRatio: compressionRatio,
		AverageServeTime: stats.AverageServeTime,
	}
}
