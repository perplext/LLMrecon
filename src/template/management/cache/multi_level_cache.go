package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// CacheLevel represents the level of caching
type CacheLevel int

const (
	// LevelFragment represents fragment-level caching (template parts)
	LevelFragment CacheLevel = iota
	// LevelTemplate represents template-level caching (full templates)
	LevelTemplate
	// LevelQuery represents query-level caching (data queries)
	LevelQuery
	// LevelResult represents result-level caching (execution results)
	LevelResult
)

// CacheOptions represents options for the multi-level cache
type CacheOptions struct {
	// EnableFragmentCache enables fragment-level caching
	EnableFragmentCache bool
	// EnableTemplateCache enables template-level caching
	EnableTemplateCache bool
	// EnableQueryCache enables query-level caching
	EnableQueryCache bool
	// EnableResultCache enables result-level caching
	EnableResultCache bool
	// FragmentTTL is the TTL for fragment cache entries
	FragmentTTL time.Duration
	// TemplateTTL is the TTL for template cache entries
	TemplateTTL time.Duration
	// QueryTTL is the TTL for query cache entries
	QueryTTL time.Duration
	// ResultTTL is the TTL for result cache entries
	ResultTTL time.Duration
	// FragmentCacheSize is the maximum size of the fragment cache
	FragmentCacheSize int
	// TemplateCacheSize is the maximum size of the template cache
	TemplateCacheSize int
	// QueryCacheSize is the maximum size of the query cache
	QueryCacheSize int
	// ResultCacheSize is the maximum size of the result cache
	ResultCacheSize int
	// EnableCompression enables compression of cached items
	EnableCompression bool
	// CompressionLevel is the compression level (1-9)
	CompressionLevel int
	// EnableSharding enables sharding of the cache
	EnableSharding bool
	// ShardCount is the number of shards
	ShardCount int
	// EnablePrefetching enables prefetching of related items
	EnablePrefetching bool
	// PrefetchCount is the number of related items to prefetch
	PrefetchCount int
	// EnableAdaptiveTTL enables adaptive TTL based on access patterns
	EnableAdaptiveTTL bool
	// MinAdaptiveTTL is the minimum TTL for adaptive TTL
	MinAdaptiveTTL time.Duration
	// MaxAdaptiveTTL is the maximum TTL for adaptive TTL
	MaxAdaptiveTTL time.Duration
}

// DefaultCacheOptions returns default cache options
func DefaultCacheOptions() *CacheOptions {
	return &CacheOptions{
		EnableFragmentCache: true,
		EnableTemplateCache: true,
		EnableQueryCache:    true,
		EnableResultCache:   true,
		FragmentTTL:         1 * time.Hour,
		TemplateTTL:         2 * time.Hour,
		QueryTTL:            30 * time.Minute,
		ResultTTL:           15 * time.Minute,
		FragmentCacheSize:   1000,
		TemplateCacheSize:   500,
		QueryCacheSize:      200,
		ResultCacheSize:     100,
		EnableCompression:   true,
		CompressionLevel:    6,
		EnableSharding:      true,
		ShardCount:          8,
		EnablePrefetching:   false,
		PrefetchCount:       5,
		EnableAdaptiveTTL:   true,
		MinAdaptiveTTL:      5 * time.Minute,
		MaxAdaptiveTTL:      24 * time.Hour,
	}
}

// MultiLevelCache is a hierarchical caching system with multiple levels
type MultiLevelCache struct {
	// fragmentCache caches template fragments
	fragmentCache *OptimizedTemplateCache
	// templateCache caches full templates
	templateCache *OptimizedTemplateCache
	// queryCache caches data query results
	queryCache *QueryCache
	// resultCache caches template execution results
	resultCache *ResultCache
	// options contains the cache configuration
	options *CacheOptions
	// stats tracks cache statistics
	stats *MultiLevelCacheStats
	// mutex protects the cache
	mutex sync.RWMutex
}

// MultiLevelCacheStats tracks statistics for the multi-level cache
type MultiLevelCacheStats struct {
	// FragmentStats tracks fragment cache statistics
	FragmentStats CacheStats
	// TemplateStats tracks template cache statistics
	TemplateStats CacheStats
	// QueryStats tracks query cache statistics
	QueryStats CacheStats
	// ResultStats tracks result cache statistics
	ResultStats CacheStats
	// TotalHits is the total number of cache hits across all levels
	TotalHits int64
	// TotalMisses is the total number of cache misses across all levels
	TotalMisses int64
	// TotalLookups is the total number of lookups across all levels
	TotalLookups int64
	// HitRatio is the overall hit ratio
	HitRatio float64
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(options *CacheOptions) *MultiLevelCache {
	if options == nil {
		options = DefaultCacheOptions()
	}

	cache := &MultiLevelCache{
		options: options,
		stats:   &MultiLevelCacheStats{},
	}

	// Initialize fragment cache
	if options.EnableFragmentCache {
		cache.fragmentCache = NewOptimizedTemplateCache(options.FragmentTTL, options.FragmentCacheSize)
	}

	// Initialize template cache
	if options.EnableTemplateCache {
		cache.templateCache = NewOptimizedTemplateCache(options.TemplateTTL, options.TemplateCacheSize)
	}

	// Initialize query cache
	if options.EnableQueryCache {
		cache.queryCache = NewQueryCache(options.QueryTTL, options.QueryCacheSize, options.EnableCompression)
	}

	// Initialize result cache
	if options.EnableResultCache {
		cache.resultCache = NewResultCache(options.ResultTTL, options.ResultCacheSize, options.EnableCompression)
	}

	return cache
}

// GetTemplate gets a template from the cache
func (c *MultiLevelCache) GetTemplate(id string) (*format.Template, bool) {
	if !c.options.EnableTemplateCache || c.templateCache == nil {
		c.stats.TotalMisses++
		c.stats.TotalLookups++
		return nil, false
	}

	template, found := c.templateCache.Get(id)
	c.stats.TotalLookups++

	if found {
		c.stats.TotalHits++
		c.stats.TemplateStats.Hits++
	} else {
		c.stats.TotalMisses++
		c.stats.TemplateStats.Misses++
	}

	// Update hit ratio
	c.updateHitRatio()

	return template, found
}

// SetTemplate sets a template in the cache
func (c *MultiLevelCache) SetTemplate(id string, template *format.Template) {
	if !c.options.EnableTemplateCache || c.templateCache == nil {
		return
	}

	c.templateCache.Set(id, template)

	// If fragment caching is enabled, cache template fragments
	if c.options.EnableFragmentCache && c.fragmentCache != nil && template != nil {
		c.cacheTemplateFragments(id, template)
	}
}

// GetFragment gets a template fragment from the cache
func (c *MultiLevelCache) GetFragment(id string) (*format.Template, bool) {
	if !c.options.EnableFragmentCache || c.fragmentCache == nil {
		c.stats.TotalMisses++
		c.stats.TotalLookups++
		return nil, false
	}

	fragment, found := c.fragmentCache.Get(id)
	c.stats.TotalLookups++

	if found {
		c.stats.TotalHits++
		c.stats.FragmentStats.Hits++
	} else {
		c.stats.TotalMisses++
		c.stats.FragmentStats.Misses++
	}

	// Update hit ratio
	c.updateHitRatio()

	return fragment, found
}

// SetFragment sets a template fragment in the cache
func (c *MultiLevelCache) SetFragment(id string, fragment *format.Template) {
	if !c.options.EnableFragmentCache || c.fragmentCache == nil {
		return
	}

	c.fragmentCache.Set(id, fragment)
}

// GetQueryResult gets a query result from the cache
func (c *MultiLevelCache) GetQueryResult(query string) (interface{}, bool) {
	if !c.options.EnableQueryCache || c.queryCache == nil {
		c.stats.TotalMisses++
		c.stats.TotalLookups++
		return nil, false
	}

	result, found := c.queryCache.Get(query)
	c.stats.TotalLookups++

	if found {
		c.stats.TotalHits++
		c.stats.QueryStats.Hits++
	} else {
		c.stats.TotalMisses++
		c.stats.QueryStats.Misses++
	}

	// Update hit ratio
	c.updateHitRatio()

	return result, found
}

// SetQueryResult sets a query result in the cache
func (c *MultiLevelCache) SetQueryResult(query string, result interface{}) {
	if !c.options.EnableQueryCache || c.queryCache == nil {
		return
	}

	c.queryCache.Set(query, result)
}

// GetExecutionResult gets a template execution result from the cache
func (c *MultiLevelCache) GetExecutionResult(templateID string, options string) (*interfaces.TemplateResult, bool) {
	if !c.options.EnableResultCache || c.resultCache == nil {
		c.stats.TotalMisses++
		c.stats.TotalLookups++
		return nil, false
	}

	// Create a cache key that includes both template ID and options
	key := fmt.Sprintf("%s:%s", templateID, options)

	result, found := c.resultCache.Get(key)
	c.stats.TotalLookups++

	if found {
		c.stats.TotalHits++
		c.stats.ResultStats.Hits++
	} else {
		c.stats.TotalMisses++
		c.stats.ResultStats.Misses++
	}

	// Update hit ratio
	c.updateHitRatio()

	return result, found
}

// SetExecutionResult sets a template execution result in the cache
func (c *MultiLevelCache) SetExecutionResult(templateID string, options string, result *interfaces.TemplateResult) {
	if !c.options.EnableResultCache || c.resultCache == nil {
		return
	}

	// Create a cache key that includes both template ID and options
	key := fmt.Sprintf("%s:%s", templateID, options)

	c.resultCache.Set(key, result)
}

// Clear clears all caches
func (c *MultiLevelCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.fragmentCache != nil {
		c.fragmentCache.Clear()
	}

	if c.templateCache != nil {
		c.templateCache.Clear()
	}

	if c.queryCache != nil {
		c.queryCache.Clear()
	}

	if c.resultCache != nil {
		c.resultCache.Clear()
	}

	// Reset statistics
	c.stats = &MultiLevelCacheStats{}
}

// ClearLevel clears a specific cache level
func (c *MultiLevelCache) ClearLevel(level CacheLevel) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	switch level {
	case LevelFragment:
		if c.fragmentCache != nil {
			c.fragmentCache.Clear()
			c.stats.FragmentStats = CacheStats{}
		}
	case LevelTemplate:
		if c.templateCache != nil {
			c.templateCache.Clear()
			c.stats.TemplateStats = CacheStats{}
		}
	case LevelQuery:
		if c.queryCache != nil {
			c.queryCache.Clear()
			c.stats.QueryStats = CacheStats{}
		}
	case LevelResult:
		if c.resultCache != nil {
			c.resultCache.Clear()
			c.stats.ResultStats = CacheStats{}
		}
	}

	// Update hit ratio
	c.updateHitRatio()
}

// GetStats returns statistics about the cache
func (c *MultiLevelCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_hits":     c.stats.TotalHits,
		"total_misses":   c.stats.TotalMisses,
		"total_lookups":  c.stats.TotalLookups,
		"hit_ratio":      c.stats.HitRatio,
		"fragment_stats": c.getFragmentStats(),
		"template_stats": c.getTemplateStats(),
		"query_stats":    c.getQueryStats(),
		"result_stats":   c.getResultStats(),
	}

	return stats
}

// Prune removes old entries from all caches
func (c *MultiLevelCache) Prune(ctx context.Context) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0

	if c.fragmentCache != nil {
		count += c.fragmentCache.Prune(c.options.FragmentTTL)
	}

	if c.templateCache != nil {
		count += c.templateCache.Prune(c.options.TemplateTTL)
	}

	if c.queryCache != nil {
		count += c.queryCache.Prune(c.options.QueryTTL)
	}

	if c.resultCache != nil {
		count += c.resultCache.Prune(c.options.ResultTTL)
	}

	return count
}

// PruneLevel prunes a specific cache level
func (c *MultiLevelCache) PruneLevel(level CacheLevel, maxAge time.Duration) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0

	switch level {
	case LevelFragment:
		if c.fragmentCache != nil {
			count = c.fragmentCache.Prune(maxAge)
		}
	case LevelTemplate:
		if c.templateCache != nil {
			count = c.templateCache.Prune(maxAge)
		}
	case LevelQuery:
		if c.queryCache != nil {
			count = c.queryCache.Prune(maxAge)
		}
	case LevelResult:
		if c.resultCache != nil {
			count = c.resultCache.Prune(maxAge)
		}
	}

	return count
}

// PreloadTemplates preloads templates into the cache
func (c *MultiLevelCache) PreloadTemplates(templates map[string]*format.Template) {
	if !c.options.EnableTemplateCache || c.templateCache == nil {
		return
	}

	c.templateCache.PreloadTemplates(templates)

	// If fragment caching is enabled, cache template fragments
	if c.options.EnableFragmentCache && c.fragmentCache != nil {
		for id, template := range templates {
			c.cacheTemplateFragments(id, template)
		}
	}
}

// cacheTemplateFragments caches fragments of a template
func (c *MultiLevelCache) cacheTemplateFragments(templateID string, template *format.Template) {
	if template == nil || len(template.Content) == 0 {
		return
	}

	// TODO: Implement template fragment caching
	// The current Template.Content field is []byte, not a structured type
	// This functionality needs to be redesigned to work with the actual template structure
	
	// For now, just return without caching fragments
}

// updateHitRatio updates the overall hit ratio
func (c *MultiLevelCache) updateHitRatio() {
	if c.stats.TotalLookups > 0 {
		c.stats.HitRatio = float64(c.stats.TotalHits) / float64(c.stats.TotalLookups)
	}
}

// getFragmentStats returns statistics for the fragment cache
func (c *MultiLevelCache) getFragmentStats() map[string]interface{} {
	if c.fragmentCache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := c.fragmentCache.GetStats()
	stats["enabled"] = true
	stats["size"] = c.fragmentCache.Size()
	stats["hit_ratio"] = calculateHitRatio(c.stats.FragmentStats.Hits, c.stats.FragmentStats.TotalLookups)

	return stats
}

// getTemplateStats returns statistics for the template cache
func (c *MultiLevelCache) getTemplateStats() map[string]interface{} {
	if c.templateCache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := c.templateCache.GetStats()
	stats["enabled"] = true
	stats["size"] = c.templateCache.Size()
	stats["hit_ratio"] = calculateHitRatio(c.stats.TemplateStats.Hits, c.stats.TemplateStats.TotalLookups)

	return stats
}

// getQueryStats returns statistics for the query cache
func (c *MultiLevelCache) getQueryStats() map[string]interface{} {
	if c.queryCache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := c.queryCache.GetStats()
	stats["enabled"] = true
	stats["size"] = c.queryCache.Size()
	stats["hit_ratio"] = calculateHitRatio(c.stats.QueryStats.Hits, c.stats.QueryStats.TotalLookups)

	return stats
}

// getResultStats returns statistics for the result cache
func (c *MultiLevelCache) getResultStats() map[string]interface{} {
	if c.resultCache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := c.resultCache.GetStats()
	stats["enabled"] = true
	stats["size"] = c.resultCache.Size()
	stats["hit_ratio"] = calculateHitRatio(c.stats.ResultStats.Hits, c.stats.ResultStats.TotalLookups)

	return stats
}

// calculateHitRatio calculates the hit ratio
func calculateHitRatio(hits, lookups int64) float64 {
	if lookups == 0 {
		return 0
	}
	return float64(hits) / float64(lookups)
}
