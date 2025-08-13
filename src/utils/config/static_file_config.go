package config

import "time"

// StaticFileHandlerConfig contains configuration options for the static file handler
type StaticFileHandlerConfig struct {
	// Root directory for static files
	RootDir string `json:"root_dir"`
	
	// Whether to enable file caching
	EnableCache *bool `json:"enable_cache"`
	
	// Maximum cache size in bytes
	MaxCacheSize int64 `json:"max_cache_size"`
	
	// Whether to enable gzip compression
	EnableCompression *bool `json:"enable_compression"`
	
	// Minimum file size for compression in bytes
	MinCompressSize int64 `json:"min_compress_size"`
	
	// Cache expiration time in seconds
	CacheExpirationSeconds int `json:"cache_expiration_seconds"`
	
	// File extensions to compress
	CompressExtensions []string `json:"compress_extensions"`
}

// DefaultStaticFileHandlerConfig returns default configuration for the static file handler
func DefaultStaticFileHandlerConfig() *StaticFileHandlerConfig {
	enableCache := true
	enableCompression := true
	
	return &StaticFileHandlerConfig{
		RootDir:                "static",
		EnableCache:            &enableCache,
		MaxCacheSize:           100 * 1024 * 1024, // 100 MB
		EnableCompression:      &enableCompression,
		MinCompressSize:        1024, // 1 KB
		CacheExpirationSeconds: int(time.Hour.Seconds()),
		CompressExtensions:     []string{".html", ".css", ".js", ".json", ".xml", ".txt", ".md"},
	}
}
