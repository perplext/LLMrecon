package cache

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"encoding/gob"
	"math"
	"sync"

	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// ResultCacheEntry represents a cached template execution result with metadata
type ResultCacheEntry struct {
	// Result is the cached template execution result
	Result *interfaces.TemplateResult
	// CreatedAt is the time the entry was created
	CreatedAt time.Time
	// ExpiresAt is the time the entry expires
	ExpiresAt time.Time
	// Size is an estimate of the entry's memory size
	Size int
	// AccessCount is the number of times the entry has been accessed
	AccessCount int
	// LastAccessed is the time the entry was last accessed
	LastAccessed time.Time
	// Compressed indicates if the result is compressed
	Compressed bool
}

// ResultCache is a cache for template execution results
type ResultCache struct {
	// cache is a map of result key to cache entry
	cache map[string]*ResultCacheEntry
	// evictionList is a doubly linked list for LRU eviction
	evictionList *list.List
	// evictionMap maps result keys to list elements for O(1) lookup
	evictionMap map[string]*list.Element
	// mutex is a mutex for the cache
	mutex sync.RWMutex
	// defaultTTL is the default time-to-live for cache entries
	defaultTTL time.Duration
	// maxSize is the maximum size of the cache
	maxSize int
	// currentSize is the current size of the cache
	currentSize int
	// stats tracks cache statistics
	stats CacheStats
	// enableCompression enables compression of cached results
	enableCompression bool
	// compressionLevel is the compression level (1-9)
	compressionLevel int
	// adaptiveTTL enables adaptive TTL based on access patterns
	adaptiveTTL bool
	// minTTL is the minimum TTL for adaptive TTL
	minTTL time.Duration
	// maxTTL is the maximum TTL for adaptive TTL
	maxTTL time.Duration
}

// NewResultCache creates a new result cache
func NewResultCache(defaultTTL time.Duration, maxSize int, enableCompression bool) *ResultCache {
	// Set default values
	if defaultTTL == 0 {
		defaultTTL = 15 * time.Minute
	}
	if maxSize <= 0 {
		maxSize = 100
	}

	return &ResultCache{
		cache:             make(map[string]*ResultCacheEntry),
		evictionList:      list.New(),
		evictionMap:       make(map[string]*list.Element),
		defaultTTL:        defaultTTL,
		maxSize:           maxSize,
		enableCompression: enableCompression,
		compressionLevel:  6, // Default compression level
		adaptiveTTL:       true,
		minTTL:            1 * time.Minute,
		maxTTL:            1 * time.Hour,
	}
}

// Get gets a template execution result from the cache
func (c *ResultCache) Get(key string) (*interfaces.TemplateResult, bool) {
	c.mutex.RLock()
	entry, exists := c.cache[key]
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
		c.removeEntry(key)
		c.stats.Expirations++
		c.mutex.Unlock()
		c.stats.Misses++
		return nil, false
	}

	// Update position in eviction list (mark as recently used)
	c.mutex.Lock()
	c.updateEntryPosition(key)
	
	// Update access statistics
	entry.AccessCount++
	entry.LastAccessed = time.Now()
	
	// If adaptive TTL is enabled, extend TTL based on access count
	if c.adaptiveTTL {
		c.extendTTL(entry)
	}
	
	// Get the result (decompress if needed)
	result := entry.Result
	if entry.Compressed {
		result = c.decompressResult(entry.Result)
	}
	
	c.mutex.Unlock()

	c.stats.Hits++
	return result, true
}

// Set sets a template execution result in the cache
func (c *ResultCache) Set(key string, result *interfaces.TemplateResult) {
	c.SetWithTTL(key, result, c.defaultTTL)
}

// SetWithTTL sets a template execution result in the cache with a specific TTL
func (c *ResultCache) SetWithTTL(key string, result *interfaces.TemplateResult, ttl time.Duration) {
	if result == nil {
		return
	}

	// Compress result if enabled
	compressed := false
	if c.enableCompression {
		result = c.compressResult(result)
		compressed = true
	}

	// Calculate size estimate
	size := estimateResultSize(result)

	// Create entry
	entry := &ResultCacheEntry{
		Result:       result,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		Size:         size,
		AccessCount:  0,
		LastAccessed: time.Now(),
		Compressed:   compressed,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if entry already exists
	if elem, exists := c.evictionMap[key]; exists {
		// Update existing entry
		oldEntry := c.cache[key]
		c.currentSize -= oldEntry.Size
		c.currentSize += size
		
		c.evictionList.MoveToFront(elem)
		c.cache[key] = entry
		elem.Value = key
	} else {
		// Add new entry
		elem := c.evictionList.PushFront(key)
		c.cache[key] = entry
		c.evictionMap[key] = elem
		c.currentSize += size

		// Check if cache exceeds max size
		c.evictIfNeeded()
	}
}

// Delete deletes a template execution result from the cache
func (c *ResultCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.removeEntry(key)
}

// Clear clears the cache
func (c *ResultCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*ResultCacheEntry)
	c.evictionList = list.New()
	c.evictionMap = make(map[string]*list.Element)
	c.currentSize = 0
}

// Size returns the number of entries in the cache
func (c *ResultCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// Prune removes entries from the cache that are older than the specified duration
func (c *ResultCache) Prune(maxAge time.Duration) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0
	now := time.Now()
	threshold := now.Add(-maxAge)

	// Iterate through the cache and remove old entries
	for key, entry := range c.cache {
		if entry.CreatedAt.Before(threshold) || now.After(entry.ExpiresAt) {
			c.removeEntry(key)
			count++
			c.stats.Expirations++
		}
	}

	return count
}

// GetStats returns statistics about the cache
func (c *ResultCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	hitRatio := 0.0
	if c.stats.TotalLookups > 0 {
		hitRatio = float64(c.stats.Hits) / float64(c.stats.TotalLookups)
	}

	return map[string]interface{}{
		"hits":              c.stats.Hits,
		"misses":            c.stats.Misses,
		"evictions":         c.stats.Evictions,
		"expirations":       c.stats.Expirations,
		"total_lookups":     c.stats.TotalLookups,
		"hit_ratio":         hitRatio,
		"current_size":      c.currentSize,
		"max_size":          c.maxSize,
		"entry_count":       len(c.cache),
		"compression":       c.enableCompression,
		"adaptive_ttl":      c.adaptiveTTL,
		"memory_usage_bytes": c.currentSize,
	}
}

// SetMaxSize sets the maximum size of the cache
func (c *ResultCache) SetMaxSize(maxSize int) {
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
func (c *ResultCache) SetDefaultTTL(ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.defaultTTL = ttl
}

// SetCompressionEnabled sets whether compression is enabled
func (c *ResultCache) SetCompressionEnabled(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.enableCompression = enabled
}

// SetCompressionLevel sets the compression level (1-9)
func (c *ResultCache) SetCompressionLevel(level int) {
	if level < 1 || level > 9 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.compressionLevel = level
}

// SetAdaptiveTTL sets whether adaptive TTL is enabled
func (c *ResultCache) SetAdaptiveTTL(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.adaptiveTTL = enabled
}

// SetAdaptiveTTLRange sets the range for adaptive TTL
func (c *ResultCache) SetAdaptiveTTLRange(min, max time.Duration) {
	if min <= 0 || max <= 0 || min > max {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.minTTL = min
	c.maxTTL = max
}

// removeEntry removes an entry from the cache
func (c *ResultCache) removeEntry(key string) {
	entry, exists := c.cache[key]
	if !exists {
		return
	}

	// Remove from eviction list
	if elem, ok := c.evictionMap[key]; ok {
		c.evictionList.Remove(elem)
		delete(c.evictionMap, key)
	}

	// Remove from cache
	c.currentSize -= entry.Size
	delete(c.cache, key)
}

// updateEntryPosition updates the position of an entry in the eviction list
func (c *ResultCache) updateEntryPosition(key string) {
	if elem, ok := c.evictionMap[key]; ok {
		c.evictionList.MoveToFront(elem)
	}
}

// evictIfNeeded evicts entries if the cache exceeds the maximum size
func (c *ResultCache) evictIfNeeded() {
	for c.currentSize > c.maxSize && c.evictionList.Len() > 0 {
		// Get the least recently used entry
		elem := c.evictionList.Back()
		if elem == nil {
			break
		}

		// Get the key
		key := elem.Value.(string)

		// Remove the entry
		c.removeEntry(key)
		c.stats.Evictions++
	}
}

// compressResult compresses a template execution result
func (c *ResultCache) compressResult(result *interfaces.TemplateResult) *interfaces.TemplateResult {
	// Create a copy of the result
	compressedResult := &interfaces.TemplateResult{
		TemplateID:           result.TemplateID,
		Success:              result.Success,
		VulnerabilityDetected: result.VulnerabilityDetected,
		VulnerabilityScore:   result.VulnerabilityScore,
		Error:                result.Error,
		ExecutionTime:        result.ExecutionTime,
		Timestamp:            result.Timestamp,
		Status:               result.Status,
		StartTime:            result.StartTime,
		EndTime:              result.EndTime,
		Duration:             result.Duration,
		Detected:             result.Detected,
		Score:                result.Score,
	}

	// Compress the response
	if result.Response != "" {
		var buf bytes.Buffer
		gzw, err := gzip.NewWriterLevel(&buf, c.compressionLevel)
		if err == nil {
			_, err = gzw.Write([]byte(result.Response))
			gzw.Close()
			if err == nil {
				// Store compressed response as base64-encoded string
				compressedResult.Response = string(buf.Bytes())
			} else {
				compressedResult.Response = result.Response
			}
		} else {
			compressedResult.Response = result.Response
		}
	}

	// Compress vulnerability details if present
	if result.VulnerabilityDetails != nil {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(result.VulnerabilityDetails)
		if err == nil {
			var compressed bytes.Buffer
			gzw, err := gzip.NewWriterLevel(&compressed, c.compressionLevel)
			if err == nil {
				_, err = io.Copy(gzw, &buf)
				gzw.Close()
				if err == nil {
					// Store compressed details as map with special key
					compressedResult.VulnerabilityDetails = map[string]interface{}{
						"__compressed__": compressed.Bytes(),
					}
				} else {
					compressedResult.VulnerabilityDetails = result.VulnerabilityDetails
				}
			} else {
				compressedResult.VulnerabilityDetails = result.VulnerabilityDetails
			}
		} else {
			compressedResult.VulnerabilityDetails = result.VulnerabilityDetails
		}
	}

	// Compress details if present
	if result.Details != nil {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(result.Details)
		if err == nil {
			var compressed bytes.Buffer
			gzw, err := gzip.NewWriterLevel(&compressed, c.compressionLevel)
			if err == nil {
				_, err = io.Copy(gzw, &buf)
				gzw.Close()
				if err == nil {
					// Store compressed details as map with special key
					compressedResult.Details = map[string]interface{}{
						"__compressed__": compressed.Bytes(),
					}
				} else {
					compressedResult.Details = result.Details
				}
			} else {
				compressedResult.Details = result.Details
			}
		} else {
			compressedResult.Details = result.Details
		}
	}

	return compressedResult
}

// decompressResult decompresses a template execution result
func (c *ResultCache) decompressResult(result *interfaces.TemplateResult) *interfaces.TemplateResult {
	// Create a copy of the result
	decompressedResult := &interfaces.TemplateResult{
		TemplateID:           result.TemplateID,
		Success:              result.Success,
		VulnerabilityDetected: result.VulnerabilityDetected,
		VulnerabilityScore:   result.VulnerabilityScore,
		Error:                result.Error,
		ExecutionTime:        result.ExecutionTime,
		Timestamp:            result.Timestamp,
		Status:               result.Status,
		StartTime:            result.StartTime,
		EndTime:              result.EndTime,
		Duration:             result.Duration,
		Detected:             result.Detected,
		Score:                result.Score,
	}

	// Decompress the response
	if result.Response != "" {
		gzr, err := gzip.NewReader(bytes.NewReader([]byte(result.Response)))
		if err == nil {
			var decompressed bytes.Buffer
			_, err = io.Copy(&decompressed, gzr)
			gzr.Close()
			if err == nil {
				decompressedResult.Response = decompressed.String()
			} else {
				decompressedResult.Response = result.Response
			}
		} else {
			decompressedResult.Response = result.Response
		}
	}

	// Decompress vulnerability details if present
	if result.VulnerabilityDetails != nil {
		// Check if details are compressed
		if compressed, ok := result.VulnerabilityDetails["__compressed__"].([]byte); ok {
			gzr, err := gzip.NewReader(bytes.NewReader(compressed))
			if err == nil {
				var decompressed bytes.Buffer
				_, err = io.Copy(&decompressed, gzr)
				gzr.Close()
				if err == nil {
					var details map[string]interface{}
					dec := gob.NewDecoder(&decompressed)
					err = dec.Decode(&details)
					if err == nil {
						decompressedResult.VulnerabilityDetails = details
					} else {
						decompressedResult.VulnerabilityDetails = result.VulnerabilityDetails
					}
				} else {
					decompressedResult.VulnerabilityDetails = result.VulnerabilityDetails
				}
			} else {
				decompressedResult.VulnerabilityDetails = result.VulnerabilityDetails
			}
		} else {
			decompressedResult.VulnerabilityDetails = result.VulnerabilityDetails
		}
	}

	// Decompress details if present
	if result.Details != nil {
		// Check if details are compressed
		if compressed, ok := result.Details["__compressed__"].([]byte); ok {
			gzr, err := gzip.NewReader(bytes.NewReader(compressed))
			if err == nil {
				var decompressed bytes.Buffer
				_, err = io.Copy(&decompressed, gzr)
				gzr.Close()
				if err == nil {
					var details map[string]interface{}
					dec := gob.NewDecoder(&decompressed)
					err = dec.Decode(&details)
					if err == nil {
						decompressedResult.Details = details
					} else {
						decompressedResult.Details = result.Details
					}
				} else {
					decompressedResult.Details = result.Details
				}
			} else {
				decompressedResult.Details = result.Details
			}
		} else {
			decompressedResult.Details = result.Details
		}
	}

	return decompressedResult
}

// extendTTL extends the TTL of an entry based on access count
func (c *ResultCache) extendTTL(entry *ResultCacheEntry) {
	// Calculate TTL extension factor based on access count
	// More frequently accessed items get longer TTLs
	factor := 1.0
	if entry.AccessCount > 0 {
		factor = math.Min(float64(entry.AccessCount)/5.0+1.0, 4.0)
	}

	// Calculate new TTL
	baseTTL := c.defaultTTL
	newTTL := time.Duration(float64(baseTTL) * factor)

	// Clamp to min/max range
	if newTTL < c.minTTL {
		newTTL = c.minTTL
	} else if newTTL > c.maxTTL {
		newTTL = c.maxTTL
	}

	// Update expiration time
	entry.ExpiresAt = time.Now().Add(newTTL)
}

// estimateResultSize estimates the size of a template execution result in bytes
func estimateResultSize(result *interfaces.TemplateResult) int {
	if result == nil {
		return 0
	}

	size := 0

	// Add size of string fields
	size += len(result.TemplateID)
	size += len(result.Response)
	size += len(result.Status)

	// Add size of maps
	if result.VulnerabilityDetails != nil {
		size += 1024 // Estimate for map
	}
	if result.Details != nil {
		size += 1024 // Estimate for map
	}

	// Add fixed sizes for other fields
	size += 100 // Estimate for timestamps, durations, etc.

	return size
}
