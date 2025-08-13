package repository

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/management/cache"
)

// CacheManagerOptions represents options for the cache manager
type CacheManagerOptions struct {
	// EnableCache enables caching
	EnableCache bool
	// DefaultTTL is the default time-to-live for cache entries
	DefaultTTL time.Duration
	// MaxSize is the maximum size of the cache
	MaxSize int
	// EnableCompression enables compression of cached data
	EnableCompression bool
	// CompressionLevel is the compression level (1-9)
	CompressionLevel int
	// PruneInterval is the interval at which to prune the cache
	PruneInterval time.Duration
}

// DefaultCacheManagerOptions returns default cache manager options
func DefaultCacheManagerOptions() *CacheManagerOptions {
	return &CacheManagerOptions{
		EnableCache:      true,
		DefaultTTL:       30 * time.Minute,
		MaxSize:          1000,
		EnableCompression: true,
		CompressionLevel: 6,
		PruneInterval:    10 * time.Minute,
	}
}

// CacheManager manages caching for repository operations
type CacheManager struct {
	// manager is the repository manager
	manager *Manager
	// queryCache is the cache for query results
	queryCache *cache.QueryCache
	// fileCache is the cache for file contents
	fileCache *cache.QueryCache
	// options contains the cache configuration
	options *CacheManagerOptions
	// mutex protects the cache manager
	mutex sync.RWMutex
	// pruneTimer is the timer for pruning the cache
	pruneTimer *time.Timer
	// stats tracks cache statistics
	stats *CacheStats
}

// CacheStats tracks statistics for the cache manager
type CacheStats struct {
	// QueryHits is the number of query cache hits
	QueryHits int64
	// QueryMisses is the number of query cache misses
	QueryMisses int64
	// FileHits is the number of file cache hits
	FileHits int64
	// FileMisses is the number of file cache misses
	FileMisses int64
	// TotalHits is the total number of cache hits
	TotalHits int64
	// TotalMisses is the total number of cache misses
	TotalMisses int64
	// TotalLookups is the total number of lookups
	TotalLookups int64
	// HitRatio is the overall hit ratio
	HitRatio float64
}

// NewCacheManager creates a new cache manager
func NewCacheManager(manager *Manager, options *CacheManagerOptions) *CacheManager {
	if options == nil {
		options = DefaultCacheManagerOptions()
	}

	cacheManager := &CacheManager{
		manager:  manager,
		options:  options,
		stats:    &CacheStats{},
	}

	if options.EnableCache {
		// Initialize query cache
		cacheManager.queryCache = cache.NewQueryCache(options.DefaultTTL, options.MaxSize, options.EnableCompression)
		
		// Initialize file cache
		cacheManager.fileCache = cache.NewQueryCache(options.DefaultTTL, options.MaxSize, options.EnableCompression)
		
		// Start prune timer
		cacheManager.startPruneTimer()
	}

	return cacheManager
}

// FindFile finds a file in all repositories with caching
func (c *CacheManager) FindFile(ctx context.Context, path string) (Repository, error) {
	if !c.options.EnableCache || c.queryCache == nil {
		// If caching is disabled, delegate to the manager
		return c.manager.FindFile(ctx, path)
	}

	// Generate cache key
	key := fmt.Sprintf("find_file:%s", path)

	// Try to get from cache
	c.stats.TotalLookups++
	if cachedResult, found := c.queryCache.Get(key); found {
		c.stats.TotalHits++
		c.stats.QueryHits++
		
		// Update hit ratio
		c.updateHitRatio()
		
		// Check if result is an error
		if errStr, ok := cachedResult.(string); ok && strings.HasPrefix(errStr, "error:") {
			return nil, fmt.Errorf(strings.TrimPrefix(errStr, "error:"))
		}
		
		// Get repository by name
		repoName, ok := cachedResult.(string)
		if !ok {
			c.stats.TotalMisses++
			c.stats.QueryMisses++
			c.updateHitRatio()
			return c.manager.FindFile(ctx, path)
		}
		
		repo, err := c.manager.GetRepository(repoName)
		if err != nil {
			c.stats.TotalMisses++
			c.stats.QueryMisses++
			c.updateHitRatio()
			return c.manager.FindFile(ctx, path)
		}
		
		return repo, nil
	}

	// Cache miss
	c.stats.TotalMisses++
	c.stats.QueryMisses++
	c.updateHitRatio()

	// Get from manager
	repo, err := c.manager.FindFile(ctx, path)
	
	// Cache the result
	if err != nil {
		// Cache the error
		c.queryCache.Set(key, fmt.Sprintf("error:%s", err.Error()))
	} else {
		// Cache the repository name
		c.queryCache.Set(key, repo.GetName())
	}
	
	return repo, err
}

// FindFiles finds files matching a pattern in all repositories with caching
func (c *CacheManager) FindFiles(ctx context.Context, pattern string) (map[Repository][]FileInfo, error) {
	if !c.options.EnableCache || c.queryCache == nil {
		// If caching is disabled, delegate to the manager
		return c.manager.FindFiles(ctx, pattern)
	}

	// Generate cache key
	key := fmt.Sprintf("find_files:%s", pattern)

	// Try to get from cache
	c.stats.TotalLookups++
	if cachedResult, found := c.queryCache.Get(key); found {
		c.stats.TotalHits++
		c.stats.QueryHits++
		
		// Update hit ratio
		c.updateHitRatio()
		
		// Check if result is an error
		if errStr, ok := cachedResult.(string); ok && strings.HasPrefix(errStr, "error:") {
			return nil, fmt.Errorf(strings.TrimPrefix(errStr, "error:"))
		}
		
		// Convert cached result to map
		cachedMap, ok := cachedResult.(map[string][]FileInfo)
		if !ok {
			c.stats.TotalMisses++
			c.stats.QueryMisses++
			c.updateHitRatio()
			return c.manager.FindFiles(ctx, pattern)
		}
		
		// Convert map keys from repository names to repositories
		result := make(map[Repository][]FileInfo)
		for repoName, files := range cachedMap {
			repo, err := c.manager.GetRepository(repoName)
			if err != nil {
				continue
			}
			result[repo] = files
		}
		
		return result, nil
	}

	// Cache miss
	c.stats.TotalMisses++
	c.stats.QueryMisses++
	c.updateHitRatio()

	// Get from manager
	result, err := c.manager.FindFiles(ctx, pattern)
	
	// Cache the result
	if err != nil {
		// Cache the error
		c.queryCache.Set(key, fmt.Sprintf("error:%s", err.Error()))
	} else {
		// Convert map keys from repositories to repository names
		cachedMap := make(map[string][]FileInfo)
		for repo, files := range result {
			cachedMap[repo.GetName()] = files
		}
		
		// Cache the map
		c.queryCache.Set(key, cachedMap)
	}
	
	return result, err
}

// GetFile gets a file from any repository that has it with caching
func (c *CacheManager) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	if !c.options.EnableCache || c.fileCache == nil {
		// If caching is disabled, delegate to the manager
		return c.manager.GetFile(ctx, path)
	}

	// Generate cache key
	key := fmt.Sprintf("file:%s", path)

	// Try to get from cache
	c.stats.TotalLookups++
	if cachedResult, found := c.fileCache.Get(key); found {
		c.stats.TotalHits++
		c.stats.FileHits++
		
		// Update hit ratio
		c.updateHitRatio()
		
		// Check if result is an error
		if errStr, ok := cachedResult.(string); ok && strings.HasPrefix(errStr, "error:") {
			return nil, fmt.Errorf(strings.TrimPrefix(errStr, "error:"))
		}
		
		// Convert cached result to file content
		content, ok := cachedResult.([]byte)
		if !ok {
			c.stats.TotalMisses++
			c.stats.FileMisses++
			c.updateHitRatio()
			return c.manager.GetFile(ctx, path)
		}
		
		// Return file content as a ReadCloser
		return ioutil.NopCloser(strings.NewReader(string(content))), nil
	}

	// Cache miss
	c.stats.TotalMisses++
	c.stats.FileMisses++
	c.updateHitRatio()

	// Get from manager
	reader, err := c.manager.GetFile(ctx, path)
	
	// Cache the result
	if err != nil {
		// Cache the error
		c.fileCache.Set(key, fmt.Sprintf("error:%s", err.Error()))
		return nil, err
	}
	
	// Read the file content
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	
	// Close the reader
	reader.Close()
	
	// Cache the file content
	c.fileCache.Set(key, content)
	
	// Return file content as a ReadCloser
	return ioutil.NopCloser(strings.NewReader(string(content))), nil
}

// GetFileFromRepo gets a file from a specific repository with caching
func (c *CacheManager) GetFileFromRepo(ctx context.Context, repoName, path string) (io.ReadCloser, error) {
	if !c.options.EnableCache || c.fileCache == nil {
		// If caching is disabled, delegate to the manager
		return c.manager.GetFileFromRepo(ctx, repoName, path)
	}

	// Generate cache key
	key := fmt.Sprintf("file:%s:%s", repoName, path)

	// Try to get from cache
	c.stats.TotalLookups++
	if cachedResult, found := c.fileCache.Get(key); found {
		c.stats.TotalHits++
		c.stats.FileHits++
		
		// Update hit ratio
		c.updateHitRatio()
		
		// Check if result is an error
		if errStr, ok := cachedResult.(string); ok && strings.HasPrefix(errStr, "error:") {
			return nil, fmt.Errorf(strings.TrimPrefix(errStr, "error:"))
		}
		
		// Convert cached result to file content
		content, ok := cachedResult.([]byte)
		if !ok {
			c.stats.TotalMisses++
			c.stats.FileMisses++
			c.updateHitRatio()
			return c.manager.GetFileFromRepo(ctx, repoName, path)
		}
		
		// Return file content as a ReadCloser
		return ioutil.NopCloser(strings.NewReader(string(content))), nil
	}

	// Cache miss
	c.stats.TotalMisses++
	c.stats.FileMisses++
	c.updateHitRatio()

	// Get from manager
	reader, err := c.manager.GetFileFromRepo(ctx, repoName, path)
	
	// Cache the result
	if err != nil {
		// Cache the error
		c.fileCache.Set(key, fmt.Sprintf("error:%s", err.Error()))
		return nil, err
	}
	
	// Read the file content
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	
	// Close the reader
	reader.Close()
	
	// Cache the file content
	c.fileCache.Set(key, content)
	
	// Return file content as a ReadCloser
	return ioutil.NopCloser(strings.NewReader(string(content))), nil
}

// Clear clears all caches
func (c *CacheManager) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.queryCache != nil {
		c.queryCache.Clear()
	}

	if c.fileCache != nil {
		c.fileCache.Clear()
	}

	// Reset statistics
	c.stats = &CacheStats{}
}

// ClearQueryCache clears the query cache
func (c *CacheManager) ClearQueryCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.queryCache != nil {
		c.queryCache.Clear()
	}
}

// ClearFileCache clears the file cache
func (c *CacheManager) ClearFileCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.fileCache != nil {
		c.fileCache.Clear()
	}
}

// GetStats returns statistics about the cache
func (c *CacheManager) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := map[string]interface{}{
		"query_hits":     c.stats.QueryHits,
		"query_misses":   c.stats.QueryMisses,
		"file_hits":      c.stats.FileHits,
		"file_misses":    c.stats.FileMisses,
		"total_hits":     c.stats.TotalHits,
		"total_misses":   c.stats.TotalMisses,
		"total_lookups":  c.stats.TotalLookups,
		"hit_ratio":      c.stats.HitRatio,
		"enabled":        c.options.EnableCache,
	}

	if c.queryCache != nil {
		stats["query_cache"] = c.queryCache.GetStats()
	}

	if c.fileCache != nil {
		stats["file_cache"] = c.fileCache.GetStats()
	}

	return stats
}

// Prune removes old entries from all caches
func (c *CacheManager) Prune() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0

	if c.queryCache != nil {
		count += c.queryCache.Prune(c.options.DefaultTTL)
	}

	if c.fileCache != nil {
		count += c.fileCache.Prune(c.options.DefaultTTL)
	}

	return count
}

// SetMaxSize sets the maximum size of the caches
func (c *CacheManager) SetMaxSize(maxSize int) {
	if maxSize <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.options.MaxSize = maxSize

	if c.queryCache != nil {
		c.queryCache.SetMaxSize(maxSize)
	}

	if c.fileCache != nil {
		c.fileCache.SetMaxSize(maxSize)
	}
}

// SetDefaultTTL sets the default TTL for cache entries
func (c *CacheManager) SetDefaultTTL(ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.options.DefaultTTL = ttl

	if c.queryCache != nil {
		c.queryCache.SetDefaultTTL(ttl)
	}

	if c.fileCache != nil {
		c.fileCache.SetDefaultTTL(ttl)
	}
}

// SetCompressionEnabled sets whether compression is enabled
func (c *CacheManager) SetCompressionEnabled(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.options.EnableCompression = enabled

	if c.queryCache != nil {
		c.queryCache.SetCompressionEnabled(enabled)
	}

	if c.fileCache != nil {
		c.fileCache.SetCompressionEnabled(enabled)
	}
}

// SetCompressionLevel sets the compression level (1-9)
func (c *CacheManager) SetCompressionLevel(level int) {
	if level < 1 || level > 9 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.options.CompressionLevel = level

	if c.queryCache != nil {
		c.queryCache.SetCompressionLevel(level)
	}

	if c.fileCache != nil {
		c.fileCache.SetCompressionLevel(level)
	}
}

// SetPruneInterval sets the interval at which to prune the cache
func (c *CacheManager) SetPruneInterval(interval time.Duration) {
	if interval <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.options.PruneInterval = interval

	// Restart prune timer
	if c.pruneTimer != nil {
		c.pruneTimer.Stop()
		c.startPruneTimer()
	}
}

// startPruneTimer starts the prune timer
func (c *CacheManager) startPruneTimer() {
	c.pruneTimer = time.AfterFunc(c.options.PruneInterval, func() {
		c.Prune()
		c.startPruneTimer()
	})
}

// updateHitRatio updates the overall hit ratio
func (c *CacheManager) updateHitRatio() {
	if c.stats.TotalLookups > 0 {
		c.stats.HitRatio = float64(c.stats.TotalHits) / float64(c.stats.TotalLookups)
	}
}

// hashQuery generates a hash for a query
func hashQuery(query string) string {
	hash := md5.Sum([]byte(query))
	return hex.EncodeToString(hash[:])
}

// DefaultCacheManager is the default cache manager
var DefaultCacheManager = NewCacheManager(DefaultManager, DefaultCacheManagerOptions())

// FindFile finds a file in all repositories with caching using the default cache manager
func FindFileWithCache(ctx context.Context, path string) (Repository, error) {
	return DefaultCacheManager.FindFile(ctx, path)
}

// FindFiles finds files matching a pattern in all repositories with caching using the default cache manager
func FindFilesWithCache(ctx context.Context, pattern string) (map[Repository][]FileInfo, error) {
	return DefaultCacheManager.FindFiles(ctx, pattern)
}

// GetFile gets a file from any repository that has it with caching using the default cache manager
func GetFileWithCache(ctx context.Context, path string) (io.ReadCloser, error) {
	return DefaultCacheManager.GetFile(ctx, path)
}

// GetFileFromRepo gets a file from a specific repository with caching using the default cache manager
func GetFileFromRepoWithCache(ctx context.Context, repoName, path string) (io.ReadCloser, error) {
	return DefaultCacheManager.GetFileFromRepo(ctx, repoName, path)
}
