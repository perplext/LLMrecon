// Package cache provides caching functionality for templates.
package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
)

// OptimizedTemplateCache is an enhanced template cache with LRU eviction and TTL support
type OptimizedTemplateCache struct {
	// cache is a map of template ID to cache entry
	cache map[string]*OptimizedCacheEntry
	// evictionList is a doubly linked list for LRU eviction
	evictionList *list.List
	// evictionMap maps template IDs to list elements for O(1) lookup
	evictionMap map[string]*list.Element
	// mutex is a mutex for the cache
	mutex sync.RWMutex
	// defaultTTL is the default time-to-live for cache entries
	defaultTTL time.Duration
	// maxSize is the maximum size of the cache
	maxSize int
	// stats tracks cache statistics
	stats CacheStats
}

// OptimizedCacheEntry represents a cached template with additional metadata
type OptimizedCacheEntry struct {
	// Template is the cached template
	Template *format.Template
	// CreatedAt is the time the entry was created
	CreatedAt time.Time
	// ExpiresAt is the time the entry expires
	ExpiresAt time.Time
	// Size is an estimate of the template's memory size
	Size int
}

// CacheStats tracks cache statistics
type CacheStats struct {
	// Hits is the number of cache hits
	Hits int64
	// Misses is the number of cache misses
	Misses int64
	// Evictions is the number of cache evictions
	Evictions int64
	// Expirations is the number of expired entries
	Expirations int64
	// TotalLookups is the total number of lookups
	TotalLookups int64
}

// NewOptimizedTemplateCache creates a new optimized template cache
func NewOptimizedTemplateCache(defaultTTL time.Duration, maxSize int) *OptimizedTemplateCache {
	// Set default values
	if defaultTTL == 0 {
		defaultTTL = 1 * time.Hour
	}
	if maxSize <= 0 {
		maxSize = 100
	}

	return &OptimizedTemplateCache{
		cache:        make(map[string]*OptimizedCacheEntry),
		evictionList: list.New(),
		evictionMap:  make(map[string]*list.Element),
		defaultTTL:   defaultTTL,
		maxSize:      maxSize,
	}
}

// Get gets a template from the cache
func (c *OptimizedTemplateCache) Get(id string) (*format.Template, bool) {
	c.mutex.RLock()
	entry, exists := c.cache[id]
	c.mutex.RUnlock()

	c.stats.TotalLookups++

	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		// Entry has expired, remove it
		c.mutex.Lock()
		c.removeEntry(id)
		c.stats.Expirations++
		c.mutex.Unlock()
		c.stats.Misses++
		return nil, false
	}

	// Update position in eviction list (mark as recently used)
	c.mutex.Lock()
	c.updateEntryPosition(id)
	c.mutex.Unlock()

	c.stats.Hits++
	return entry.Template, true
}

// Set sets a template in the cache
func (c *OptimizedTemplateCache) Set(id string, template *format.Template) {
	c.SetWithTTL(id, template, c.defaultTTL)
}

// SetWithTTL sets a template in the cache with a specific TTL
func (c *OptimizedTemplateCache) SetWithTTL(id string, template *format.Template, ttl time.Duration) {
	// Calculate size estimate
	size := estimateTemplateSize(template)

	// Create entry
	entry := &OptimizedCacheEntry{
		Template:  template,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		Size:      size,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if entry already exists
	if elem, exists := c.evictionMap[id]; exists {
		// Update existing entry
		c.evictionList.MoveToFront(elem)
		c.cache[id] = entry
		elem.Value = id
	} else {
		// Add new entry
		elem := c.evictionList.PushFront(id)
		c.cache[id] = entry
		c.evictionMap[id] = elem

		// Check if cache exceeds max size
		c.evictIfNeeded()
	}
}

// Delete deletes a template from the cache
func (c *OptimizedTemplateCache) Delete(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.removeEntry(id)
}

// Clear clears the cache
func (c *OptimizedTemplateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*OptimizedCacheEntry)
	c.evictionList = list.New()
	c.evictionMap = make(map[string]*list.Element)
}

// Size returns the number of templates in the cache
func (c *OptimizedTemplateCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// Prune removes entries from the cache that are older than the specified duration
func (c *OptimizedTemplateCache) Prune(maxAge time.Duration) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0
	now := time.Now()
	threshold := now.Add(-maxAge)

	// Iterate through the cache and remove old entries
	for id, entry := range c.cache {
		if entry.CreatedAt.Before(threshold) || now.After(entry.ExpiresAt) {
			c.removeEntry(id)
			count++
			c.stats.Expirations++
		}
	}

	return count
}

// GetKeys returns the keys of all templates in the cache
func (c *OptimizedTemplateCache) GetKeys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}

	return keys
}

// GetStats returns statistics about the cache
func (c *OptimizedTemplateCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	hitRate := float64(0)
	if c.stats.TotalLookups > 0 {
		hitRate = float64(c.stats.Hits) / float64(c.stats.TotalLookups) * 100
	}

	return map[string]interface{}{
		"size":        len(c.cache),
		"max_size":    c.maxSize,
		"hits":        c.stats.Hits,
		"misses":      c.stats.Misses,
		"evictions":   c.stats.Evictions,
		"expirations": c.stats.Expirations,
		"hit_rate":    hitRate,
	}
}

// SetMaxSize sets the maximum size of the cache
func (c *OptimizedTemplateCache) SetMaxSize(maxSize int) {
	if maxSize <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.maxSize = maxSize

	// Evict entries if needed
	c.evictIfNeeded()
}

// SetDefaultTTL sets the default TTL for cache entries
func (c *OptimizedTemplateCache) SetDefaultTTL(ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.defaultTTL = ttl
}

// Refresh refreshes the expiration time of a cache entry
func (c *OptimizedTemplateCache) Refresh(id string) bool {
	return c.RefreshWithTTL(id, c.defaultTTL)
}

// RefreshWithTTL refreshes the expiration time of a cache entry with a specific TTL
func (c *OptimizedTemplateCache) RefreshWithTTL(id string, ttl time.Duration) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.cache[id]
	if !exists {
		return false
	}

	// Update expiration time
	entry.ExpiresAt = time.Now().Add(ttl)

	// Update position in eviction list
	c.updateEntryPosition(id)

	return true
}

// PreloadTemplates preloads templates into the cache
func (c *OptimizedTemplateCache) PreloadTemplates(templates map[string]*format.Template) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for id, template := range templates {
		// Calculate size estimate
		size := estimateTemplateSize(template)

		// Create entry
		entry := &OptimizedCacheEntry{
			Template:  template,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(c.defaultTTL),
			Size:      size,
		}

		// Add to cache
		elem := c.evictionList.PushFront(id)
		c.cache[id] = entry
		c.evictionMap[id] = elem
	}

	// Evict entries if needed
	c.evictIfNeeded()
}

// removeEntry removes an entry from the cache
func (c *OptimizedTemplateCache) removeEntry(id string) {
	if elem, exists := c.evictionMap[id]; exists {
		c.evictionList.Remove(elem)
		delete(c.evictionMap, id)
	}
	delete(c.cache, id)
}

// updateEntryPosition updates the position of an entry in the eviction list
func (c *OptimizedTemplateCache) updateEntryPosition(id string) {
	if elem, exists := c.evictionMap[id]; exists {
		c.evictionList.MoveToFront(elem)
	}
}

// evictIfNeeded evicts entries if the cache exceeds the maximum size
func (c *OptimizedTemplateCache) evictIfNeeded() {
	for len(c.cache) > c.maxSize {
		// Get the least recently used entry
		elem := c.evictionList.Back()
		if elem == nil {
			break
		}

		// Remove the entry
		id := elem.Value.(string)
		c.removeEntry(id)
		c.stats.Evictions++
	}
}

// estimateTemplateSize estimates the size of a template in bytes
func estimateTemplateSize(template *format.Template) int {
	if template == nil {
		return 0
	}

	// Base size for the template struct
	size := 100

	// Add size for the template ID
	size += len(template.ID)

	// Add size for the template name
	size += len(template.Info.Name)

	// Add size for the template description
	size += len(template.Info.Description)

	// Add size for the template content
	size += len(template.Test.Prompt) * 2 // Unicode characters

	// This is a rough estimate and could be improved with more detailed analysis
	return size
}
