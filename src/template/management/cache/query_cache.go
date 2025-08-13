package cache

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"math"
	"sync"
)

// QueryCacheEntry represents a cached query result with metadata
type QueryCacheEntry struct {
	// Value is the cached query result
	Value interface{}
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
	// Compressed indicates if the value is compressed
	Compressed bool
}

// QueryCache is a cache for database query results
type QueryCache struct {
	// cache is a map of query hash to cache entry
	cache map[string]*QueryCacheEntry
	// evictionList is a doubly linked list for LRU eviction
	evictionList *list.List
	// evictionMap maps query hashes to list elements for O(1) lookup
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
	// enableCompression enables compression of cached values
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

// NewQueryCache creates a new query cache
func NewQueryCache(defaultTTL time.Duration, maxSize int, enableCompression bool) *QueryCache {
	// Set default values
	if defaultTTL == 0 {
		defaultTTL = 30 * time.Minute
	}
	if maxSize <= 0 {
		maxSize = 100
	}

	return &QueryCache{
		cache:             make(map[string]*QueryCacheEntry),
		evictionList:      list.New(),
		evictionMap:       make(map[string]*list.Element),
		defaultTTL:        defaultTTL,
		maxSize:           maxSize,
		enableCompression: enableCompression,
		compressionLevel:  6, // Default compression level
		adaptiveTTL:       true,
		minTTL:            5 * time.Minute,
		maxTTL:            24 * time.Hour,
	}
}

// Get gets a query result from the cache
func (c *QueryCache) Get(query string) (interface{}, bool) {
	// Generate query hash
	queryHash := c.hashQuery(query)

	c.mutex.RLock()
	entry, exists := c.cache[queryHash]
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
		c.removeEntry(queryHash)
		c.stats.Expirations++
		c.mutex.Unlock()
		c.stats.Misses++
		return nil, false
	}

	// Update position in eviction list (mark as recently used)
	c.mutex.Lock()
	c.updateEntryPosition(queryHash)
	
	// Update access statistics
	entry.AccessCount++
	entry.LastAccessed = time.Now()
	
	// If adaptive TTL is enabled, extend TTL based on access count
	if c.adaptiveTTL {
		c.extendTTL(entry)
	}
	
	// Get the value (decompress if needed)
	value := entry.Value
	if entry.Compressed {
		value = c.decompress(value)
	}
	
	c.mutex.Unlock()

	c.stats.Hits++
	return value, true
}

// Set sets a query result in the cache
func (c *QueryCache) Set(query string, value interface{}) {
	c.SetWithTTL(query, value, c.defaultTTL)
}

// SetWithTTL sets a query result in the cache with a specific TTL
func (c *QueryCache) SetWithTTL(query string, value interface{}, ttl time.Duration) {
	// Generate query hash
	queryHash := c.hashQuery(query)

	// Compress value if enabled
	compressed := false
	if c.enableCompression {
		value = c.compress(value)
		compressed = true
	}

	// Calculate size estimate
	size := estimateSize(value)

	// Create entry
	entry := &QueryCacheEntry{
		Value:        value,
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
	if elem, exists := c.evictionMap[queryHash]; exists {
		// Update existing entry
		oldEntry := c.cache[queryHash]
		c.currentSize -= oldEntry.Size
		c.currentSize += size
		
		c.evictionList.MoveToFront(elem)
		c.cache[queryHash] = entry
		elem.Value = queryHash
	} else {
		// Add new entry
		elem := c.evictionList.PushFront(queryHash)
		c.cache[queryHash] = entry
		c.evictionMap[queryHash] = elem
		c.currentSize += size

		// Check if cache exceeds max size
		c.evictIfNeeded()
	}
}

// Delete deletes a query result from the cache
func (c *QueryCache) Delete(query string) {
	// Generate query hash
	queryHash := c.hashQuery(query)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.removeEntry(queryHash)
}

// Clear clears the cache
func (c *QueryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*QueryCacheEntry)
	c.evictionList = list.New()
	c.evictionMap = make(map[string]*list.Element)
	c.currentSize = 0
}

// Size returns the number of entries in the cache
func (c *QueryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// Prune removes entries from the cache that are older than the specified duration
func (c *QueryCache) Prune(maxAge time.Duration) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0
	now := time.Now()
	threshold := now.Add(-maxAge)

	// Iterate through the cache and remove old entries
	for queryHash, entry := range c.cache {
		if entry.CreatedAt.Before(threshold) || now.After(entry.ExpiresAt) {
			c.removeEntry(queryHash)
			count++
			c.stats.Expirations++
		}
	}

	return count
}

// GetStats returns statistics about the cache
func (c *QueryCache) GetStats() map[string]interface{} {
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
func (c *QueryCache) SetMaxSize(maxSize int) {
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
func (c *QueryCache) SetDefaultTTL(ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.defaultTTL = ttl
}

// SetCompressionEnabled sets whether compression is enabled
func (c *QueryCache) SetCompressionEnabled(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.enableCompression = enabled
}

// SetCompressionLevel sets the compression level (1-9)
func (c *QueryCache) SetCompressionLevel(level int) {
	if level < 1 || level > 9 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.compressionLevel = level
}

// SetAdaptiveTTL sets whether adaptive TTL is enabled
func (c *QueryCache) SetAdaptiveTTL(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.adaptiveTTL = enabled
}

// SetAdaptiveTTLRange sets the range for adaptive TTL
func (c *QueryCache) SetAdaptiveTTLRange(min, max time.Duration) {
	if min <= 0 || max <= 0 || min > max {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.minTTL = min
	c.maxTTL = max
}

// hashQuery generates a hash for a query
func (c *QueryCache) hashQuery(query string) string {
	hash := md5.Sum([]byte(query))
	return hex.EncodeToString(hash[:])
}

// removeEntry removes an entry from the cache
func (c *QueryCache) removeEntry(queryHash string) {
	entry, exists := c.cache[queryHash]
	if !exists {
		return
	}

	// Remove from eviction list
	if elem, ok := c.evictionMap[queryHash]; ok {
		c.evictionList.Remove(elem)
		delete(c.evictionMap, queryHash)
	}

	// Remove from cache
	c.currentSize -= entry.Size
	delete(c.cache, queryHash)
}

// updateEntryPosition updates the position of an entry in the eviction list
func (c *QueryCache) updateEntryPosition(queryHash string) {
	if elem, ok := c.evictionMap[queryHash]; ok {
		c.evictionList.MoveToFront(elem)
	}
}

// evictIfNeeded evicts entries if the cache exceeds the maximum size
func (c *QueryCache) evictIfNeeded() {
	for c.currentSize > c.maxSize && c.evictionList.Len() > 0 {
		// Get the least recently used entry
		elem := c.evictionList.Back()
		if elem == nil {
			break
		}

		// Get the query hash
		queryHash := elem.Value.(string)

		// Remove the entry
		c.removeEntry(queryHash)
		c.stats.Evictions++
	}
}

// compress compresses a value
func (c *QueryCache) compress(value interface{}) interface{} {
	// Serialize the value
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return value
	}

	// Compress the serialized value
	var compressed bytes.Buffer
	gzw, err := gzip.NewWriterLevel(&compressed, c.compressionLevel)
	if err != nil {
		return value
	}

	_, err = io.Copy(gzw, &buf)
	gzw.Close()
	if err != nil {
		return value
	}

	return compressed.Bytes()
}

// decompress decompresses a value
func (c *QueryCache) decompress(value interface{}) interface{} {
	// Check if value is compressed
	compressedBytes, ok := value.([]byte)
	if !ok {
		return value
	}

	// Decompress the value
	gzr, err := gzip.NewReader(bytes.NewReader(compressedBytes))
	if err != nil {
		return value
	}

	var decompressed bytes.Buffer
	_, err = io.Copy(&decompressed, gzr)
	gzr.Close()
	if err != nil {
		return value
	}

	// Deserialize the value
	var result interface{}
	dec := gob.NewDecoder(&decompressed)
	err = dec.Decode(&result)
	if err != nil {
		return value
	}

	return result
}

// extendTTL extends the TTL of an entry based on access count
func (c *QueryCache) extendTTL(entry *QueryCacheEntry) {
	// Calculate TTL extension factor based on access count
	// More frequently accessed items get longer TTLs
	factor := 1.0
	if entry.AccessCount > 0 {
		factor = math.Min(float64(entry.AccessCount)/10.0+1.0, 5.0)
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

// estimateSize estimates the size of a value in bytes
func estimateSize(value interface{}) int {
	switch v := value.(type) {
	case []byte:
		return len(v)
	case string:
		return len(v)
	case nil:
		return 0
	default:
		// For complex types, serialize and measure
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(value)
		if err != nil {
			return 1024 // Default size estimate
		}
		return buf.Len()
	}
}
