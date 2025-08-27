// Package cache provides caching functionality for templates.
package cache

import (
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateCache is responsible for caching templates
type TemplateCache struct {
	// cache is a map of template ID to cache entry
	cache map[string]*CacheEntry
	// mutex is a mutex for the cache
	mutex sync.RWMutex
	// defaultTTL is the default time-to-live for cache entries
	defaultTTL time.Duration
	// maxSize is the maximum size of the cache
	maxSize int
	// evictionPolicy is the policy for evicting cache entries
	evictionPolicy EvictionPolicy
}

// CacheEntry represents a cached template
type CacheEntry struct {
	// Template is the cached template
	Template *format.Template
	// CreatedAt is the time the entry was created
	CreatedAt time.Time
	// ExpiresAt is the time the entry expires
	ExpiresAt time.Time
	// LastAccessed is the time the entry was last accessed
	LastAccessed time.Time
	// AccessCount is the number of times the entry has been accessed
	AccessCount int
}

// EvictionPolicy represents the policy for evicting cache entries
type EvictionPolicy string

const (
	// LRU (Least Recently Used) evicts the least recently used entries first
	LRU EvictionPolicy = "lru"
	// LFU (Least Frequently Used) evicts the least frequently used entries first
	LFU EvictionPolicy = "lfu"
	// FIFO (First In, First Out) evicts the oldest entries first
	FIFO EvictionPolicy = "fifo"
)

// NewTemplateCache creates a new template cache
func NewTemplateCache(defaultTTL time.Duration, maxSize int, evictionPolicy EvictionPolicy) *TemplateCache {
	// Set default values
	if defaultTTL == 0 {
		defaultTTL = 1 * time.Hour
	}
	if maxSize <= 0 {
		maxSize = 100
	}
	if evictionPolicy == "" {
		evictionPolicy = LRU
	}

	return &TemplateCache{
		cache:          make(map[string]*CacheEntry),
		defaultTTL:     defaultTTL,
		maxSize:        maxSize,
		evictionPolicy: evictionPolicy,
	}
}

// Get gets a template from the cache
func (c *TemplateCache) Get(id string) (*format.Template, bool) {
	c.mutex.RLock()
	entry, exists := c.cache[id]
	c.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		// Entry has expired, remove it
		c.mutex.Lock()
		delete(c.cache, id)
		c.mutex.Unlock()
		return nil, false
	}

	// Update access information
	c.mutex.Lock()
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	c.mutex.Unlock()

	return entry.Template, true
}

// Set sets a template in the cache
func (c *TemplateCache) Set(id string, template *format.Template) {
	c.SetWithTTL(id, template, c.defaultTTL)
}

// SetWithTTL sets a template in the cache with a specific TTL
func (c *TemplateCache) SetWithTTL(id string, template *format.Template, ttl time.Duration) {
	now := time.Now()
	entry := &CacheEntry{
		Template:     template,
		CreatedAt:    now,
		ExpiresAt:    now.Add(ttl),
		LastAccessed: now,
		AccessCount:  0,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if cache is full
	if len(c.cache) >= c.maxSize && c.cache[id] == nil {
		// Cache is full, evict an entry
		c.evict()
	}

	// Add entry to cache
	c.cache[id] = entry
}

// Delete deletes a template from the cache
func (c *TemplateCache) Delete(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.cache, id)
}

// Clear clears the cache
func (c *TemplateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]*CacheEntry)
}

// Size returns the number of templates in the cache
func (c *TemplateCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// Prune removes entries from the cache that are older than the specified duration
func (c *TemplateCache) Prune(maxAge time.Duration) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0
	now := time.Now()
	cutoff := now.Add(-maxAge)

	for id, entry := range c.cache {
		// Remove if expired or older than the specified duration
		if now.After(entry.ExpiresAt) || entry.CreatedAt.Before(cutoff) {
			delete(c.cache, id)
			count++
		}
	}

	return count
}

// evict evicts an entry from the cache based on the eviction policy
func (c *TemplateCache) evict() {
	switch c.evictionPolicy {
	case LRU:
		c.evictLRU()
	case LFU:
		c.evictLFU()
	case FIFO:
		c.evictFIFO()
	default:
		c.evictLRU()
	}
}

// evictLRU evicts the least recently used entry
func (c *TemplateCache) evictLRU() {
	var oldestID string
	var oldestTime time.Time

	// Find the least recently used entry
	for id, entry := range c.cache {
		if oldestID == "" || entry.LastAccessed.Before(oldestTime) {
			oldestID = id
			oldestTime = entry.LastAccessed
		}
	}

	// Evict the entry
	if oldestID != "" {
		delete(c.cache, oldestID)
	}
}

// evictLFU evicts the least frequently used entry
func (c *TemplateCache) evictLFU() {
	var leastUsedID string
	var leastUsedCount int = -1

	// Find the least frequently used entry
	for id, entry := range c.cache {
		if leastUsedID == "" || entry.AccessCount < leastUsedCount {
			leastUsedID = id
			leastUsedCount = entry.AccessCount
		}
	}

	// Evict the entry
	if leastUsedID != "" {
		delete(c.cache, leastUsedID)
	}
}

// evictFIFO evicts the oldest entry
func (c *TemplateCache) evictFIFO() {
	var oldestID string
	var oldestTime time.Time

	// Find the oldest entry
	for id, entry := range c.cache {
		if oldestID == "" || entry.CreatedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = entry.CreatedAt
		}
	}

	// Evict the entry
	if oldestID != "" {
		delete(c.cache, oldestID)
	}
}

// GetKeys returns the keys of all templates in the cache
func (c *TemplateCache) GetKeys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}

	return keys
}

// GetStats returns statistics about the cache
func (c *TemplateCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := map[string]interface{}{
		"size":            len(c.cache),
		"max_size":        c.maxSize,
		"eviction_policy": string(c.evictionPolicy),
		"default_ttl_ms":  c.defaultTTL.Milliseconds(),
	}

	// Count expired entries
	now := time.Now()
	expiredCount := 0
	for _, entry := range c.cache {
		if now.After(entry.ExpiresAt) {
			expiredCount++
		}
	}
	stats["expired_count"] = expiredCount

	return stats
}

// SetMaxSize sets the maximum size of the cache
func (c *TemplateCache) SetMaxSize(maxSize int) {
	if maxSize <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.maxSize = maxSize

	// Evict entries if cache is now too large
	for len(c.cache) > c.maxSize {
		c.evict()
	}
}

// SetEvictionPolicy sets the eviction policy
func (c *TemplateCache) SetEvictionPolicy(policy EvictionPolicy) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.evictionPolicy = policy
}

// SetDefaultTTL sets the default TTL for cache entries
func (c *TemplateCache) SetDefaultTTL(ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.defaultTTL = ttl
}

// Refresh refreshes the expiration time of a cache entry
func (c *TemplateCache) Refresh(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.cache[id]
	if !exists {
		return false
	}

	// Update expiration time
	entry.ExpiresAt = time.Now().Add(c.defaultTTL)
	return true
}

// RefreshWithTTL refreshes the expiration time of a cache entry with a specific TTL
func (c *TemplateCache) RefreshWithTTL(id string, ttl time.Duration) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.cache[id]
	if !exists {
		return false
	}

	// Update expiration time
	entry.ExpiresAt = time.Now().Add(ttl)
	return true
}
