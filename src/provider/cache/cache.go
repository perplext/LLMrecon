// Package cache provides caching functionality for the Multi-Provider LLM Integration Framework.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// CacheKey represents a key for the cache
type CacheKey string

// CacheEntry represents an entry in the cache
type CacheEntry struct {
	// Value is the cached value
	Value interface{}
	// Expiration is the expiration time
	Expiration time.Time

// Cache is a simple in-memory cache
type Cache struct {
	// entries is a map of keys to entries
	entries map[CacheKey]*CacheEntry
	// mutex is a mutex for concurrent access to entries
	mutex sync.RWMutex
	// defaultTTL is the default time-to-live for cache entries
	defaultTTL time.Duration
	// maxEntries is the maximum number of entries in the cache
	maxEntries int
	// evictionPolicy is the eviction policy for the cache
	evictionPolicy EvictionPolicy
	// metrics tracks cache metrics
	metrics *CacheMetrics

// EvictionPolicy is the policy for evicting entries from the cache
type EvictionPolicy string

const (
	// LRU is the least recently used eviction policy
	LRU EvictionPolicy = "lru"
	// LFU is the least frequently used eviction policy
	LFU EvictionPolicy = "lfu"
	// FIFO is the first in, first out eviction policy
	FIFO EvictionPolicy = "fifo"
)

// CacheMetrics tracks metrics for the cache
type CacheMetrics struct {
	// Hits is the number of cache hits
	Hits int
	// Misses is the number of cache misses
	Misses int
	// Evictions is the number of cache evictions
	Evictions int
	// mutex is a mutex for concurrent access to metrics
	mutex sync.RWMutex

// NewCache creates a new cache
func NewCache(defaultTTL time.Duration, maxEntries int, evictionPolicy EvictionPolicy) *Cache {
	return &Cache{
		entries:        make(map[CacheKey]*CacheEntry),
		defaultTTL:     defaultTTL,
		maxEntries:     maxEntries,
		evictionPolicy: evictionPolicy,
		metrics:        &CacheMetrics{},
	}

// Set sets a value in the cache
func (c *Cache) Set(key CacheKey, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if cache is full
	if c.maxEntries > 0 && len(c.entries) >= c.maxEntries && c.entries[key] == nil {
		// Evict an entry
		c.evict()
	}

	// Set expiration time
	expiration := time.Now().Add(ttl)
	if ttl == 0 {
		expiration = time.Now().Add(c.defaultTTL)
	}

	// Create entry
	c.entries[key] = &CacheEntry{
		Value:      value,
		Expiration: expiration,
	}

// Get gets a value from the cache
func (c *Cache) Get(key CacheKey) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Get entry
	entry, ok := c.entries[key]
	if !ok {
		// Cache miss
		c.metrics.mutex.Lock()
		c.metrics.Misses++
		c.metrics.mutex.Unlock()
		return nil, false
	}

	// Check if entry is expired
	if time.Now().After(entry.Expiration) {
		// Entry is expired
		c.metrics.mutex.Lock()
		c.metrics.Misses++
		c.metrics.mutex.Unlock()
		return nil, false
	}

	// Cache hit
	c.metrics.mutex.Lock()
	c.metrics.Hits++
	c.metrics.mutex.Unlock()
	return entry.Value, true

// Delete deletes a value from the cache
func (c *Cache) Delete(key CacheKey) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, key)

// Clear clears the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[CacheKey]*CacheEntry)

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.entries)

// Keys returns the keys in the cache
func (c *Cache) Keys() []CacheKey {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]CacheKey, 0, len(c.entries))
	for key := range c.entries {
		keys = append(keys, key)
	}

	return keys

// evict evicts an entry from the cache
func (c *Cache) evict() {
	// Increment eviction count
	c.metrics.mutex.Lock()
	c.metrics.Evictions++
	c.metrics.mutex.Unlock()

	// Evict based on policy
	switch c.evictionPolicy {
	case LRU:
		c.evictLRU()
	case LFU:
		c.evictLFU()
	case FIFO:
		c.evictFIFO()
	default:
		c.evictLRU() // Default to LRU
	}

// evictLRU evicts the least recently used entry
func (c *Cache) evictLRU() {
	var oldestKey CacheKey
	var oldestTime time.Time

	// Find the oldest entry
	for key, entry := range c.entries {
		if oldestTime.IsZero() || entry.Expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Expiration
		}
	}

	// Delete the oldest entry
	delete(c.entries, oldestKey)

// evictLFU evicts the least frequently used entry
// This is a simplified implementation that uses expiration time as a proxy for frequency
func (c *Cache) evictLFU() {
	// For now, just use LRU
	c.evictLRU()

// evictFIFO evicts the first in, first out entry
// This is a simplified implementation that uses expiration time as a proxy for insertion time
func (c *Cache) evictFIFO() {
	// For now, just use LRU
	c.evictLRU()

// GetMetrics returns the cache metrics
func (c *Cache) GetMetrics() *CacheMetrics {
	c.metrics.mutex.RLock()
	defer c.metrics.mutex.RUnlock()

	// Return a copy of the metrics
	return &CacheMetrics{
		Hits:      c.metrics.Hits,
		Misses:    c.metrics.Misses,
		Evictions: c.metrics.Evictions,
	}

// ResetMetrics resets the cache metrics
func (c *Cache) ResetMetrics() {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.Hits = 0
	c.metrics.Misses = 0
	c.metrics.Evictions = 0

// GenerateKey generates a cache key from a request
func GenerateKey(providerType core.ProviderType, operation string, request interface{}) (CacheKey, error) {
	// Marshal request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create key
	key := fmt.Sprintf("%s:%s:%s", providerType, operation, requestJSON)

	// Hash key
	hash := sha256.Sum256([]byte(key))
	return CacheKey(hex.EncodeToString(hash[:])), nil

// ProviderCache is a cache for provider responses
type ProviderCache struct {
	// cache is the underlying cache
	cache *Cache
	// enabled indicates whether caching is enabled
	enabled bool
	// mutex is a mutex for concurrent access to enabled
	mutex sync.RWMutex

// NewProviderCache creates a new provider cache
func NewProviderCache(defaultTTL time.Duration, maxEntries int, evictionPolicy EvictionPolicy) *ProviderCache {
	return &ProviderCache{
		cache:   NewCache(defaultTTL, maxEntries, evictionPolicy),
		enabled: true,
	}

// Get gets a value from the cache
func (c *ProviderCache) Get(providerType core.ProviderType, operation string, request interface{}) (interface{}, bool) {
	c.mutex.RLock()
	enabled := c.enabled
	c.mutex.RUnlock()

	if !enabled {
		return nil, false
	}

	// Generate key
	key, err := GenerateKey(providerType, operation, request)
	if err != nil {
		return nil, false
	}

	// Get from cache
	return c.cache.Get(key)

// Set sets a value in the cache
func (c *ProviderCache) Set(providerType core.ProviderType, operation string, request interface{}, response interface{}, ttl time.Duration) error {
	c.mutex.RLock()
	enabled := c.enabled
	c.mutex.RUnlock()

	if !enabled {
		return nil
	}

	// Generate key
	key, err := GenerateKey(providerType, operation, request)
	if err != nil {
		return err
	}

	// Set in cache
	c.cache.Set(key, response, ttl)

	return nil

// Delete deletes a value from the cache
func (c *ProviderCache) Delete(providerType core.ProviderType, operation string, request interface{}) error {
	// Generate key
	key, err := GenerateKey(providerType, operation, request)
	if err != nil {
		return err
	}

	// Delete from cache
	c.cache.Delete(key)

	return nil

// Clear clears the cache
func (c *ProviderCache) Clear() {
	c.cache.Clear()

// Enable enables caching
func (c *ProviderCache) Enable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.enabled = true

// Disable disables caching
func (c *ProviderCache) Disable() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.enabled = false

// IsEnabled returns whether caching is enabled
func (c *ProviderCache) IsEnabled() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.enabled

// GetMetrics returns the cache metrics
func (c *ProviderCache) GetMetrics() *CacheMetrics {
	return c.cache.GetMetrics()

// ResetMetrics resets the cache metrics
func (c *ProviderCache) ResetMetrics() {
	c.cache.ResetMetrics()

// Size returns the number of entries in the cache
func (c *ProviderCache) Size() int {
	return c.cache.Size()

// CachingProvider wraps a provider with caching
type CachingProvider struct {
	// provider is the underlying provider
	provider core.Provider
	// cache is the cache
	cache *ProviderCache

// NewCachingProvider creates a new caching provider
func NewCachingProvider(provider core.Provider, cache *ProviderCache) *CachingProvider {
	return &CachingProvider{
		provider: provider,
		cache:    cache,
	}

// GetType returns the type of provider
func (p *CachingProvider) GetType() core.ProviderType {
	return p.provider.GetType()

// GetConfig returns the configuration for the provider
func (p *CachingProvider) GetConfig() *core.ProviderConfig {
	return p.provider.GetConfig()

// GetModels returns a list of available models
func (p *CachingProvider) GetModels(ctx context.Context) ([]core.ModelInfo, error) {
	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "GetModels", nil); ok {
		return cached.([]core.ModelInfo), nil
	}

	// Call provider
	models, err := p.provider.GetModels(ctx)
	if err != nil {
		return nil, err
	}

	// Cache response
	p.cache.Set(p.GetType(), "GetModels", nil, models, 1*time.Hour)

	return models, nil

// GetModelInfo returns information about a specific model
func (p *CachingProvider) GetModelInfo(ctx context.Context, modelID string) (*core.ModelInfo, error) {
	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "GetModelInfo", modelID); ok {
		return cached.(*core.ModelInfo), nil
	}

	// Call provider
	modelInfo, err := p.provider.GetModelInfo(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// Cache response
	p.cache.Set(p.GetType(), "GetModelInfo", modelID, modelInfo, 1*time.Hour)

	return modelInfo, nil

// TextCompletion generates a text completion
func (p *CachingProvider) TextCompletion(ctx context.Context, request *core.TextCompletionRequest) (*core.TextCompletionResponse, error) {
	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "TextCompletion", request); ok {
		return cached.(*core.TextCompletionResponse), nil
	}

	// Call provider
	response, err := p.provider.TextCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	// Cache response
	p.cache.Set(p.GetType(), "TextCompletion", request, response, 24*time.Hour)

	return response, nil

// ChatCompletion generates a chat completion
func (p *CachingProvider) ChatCompletion(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "ChatCompletion", request); ok {
		return cached.(*core.ChatCompletionResponse), nil
	}

	// Call provider
	response, err := p.provider.ChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	// Cache response
	p.cache.Set(p.GetType(), "ChatCompletion", request, response, 24*time.Hour)

	return response, nil

// StreamingChatCompletion generates a streaming chat completion
func (p *CachingProvider) StreamingChatCompletion(ctx context.Context, request *core.ChatCompletionRequest, callback func(response *core.ChatCompletionResponse) error) error {
	// Streaming responses are not cached
	return p.provider.StreamingChatCompletion(ctx, request, callback)

// CreateEmbedding creates an embedding
func (p *CachingProvider) CreateEmbedding(ctx context.Context, request *core.EmbeddingRequest) (*core.EmbeddingResponse, error) {
	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "CreateEmbedding", request); ok {
		return cached.(*core.EmbeddingResponse), nil
	}

	// Call provider
	response, err := p.provider.CreateEmbedding(ctx, request)
	if err != nil {
		return nil, err
	}

	// Cache response
	p.cache.Set(p.GetType(), "CreateEmbedding", request, response, 24*time.Hour)

	return response, nil

// CountTokens counts the number of tokens in a text
func (p *CachingProvider) CountTokens(ctx context.Context, text string, modelID string) (int, error) {
	// Create a key for the cache
	key := struct {
		Text    string
		ModelID string
	}{
		Text:    text,
		ModelID: modelID,
	}

	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "CountTokens", key); ok {
		return cached.(int), nil
	}

	// Call provider
	count, err := p.provider.CountTokens(ctx, text, modelID)
	if err != nil {
		return 0, err
	}

	// Cache response
	p.cache.Set(p.GetType(), "CountTokens", key, count, 24*time.Hour)

	return count, nil

// SupportsModel returns whether the provider supports a specific model
func (p *CachingProvider) SupportsModel(ctx context.Context, modelID string) bool {
	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "SupportsModel", modelID); ok {
		return cached.(bool)
	}

	// Call provider
	supports := p.provider.SupportsModel(ctx, modelID)

	// Cache response
	p.cache.Set(p.GetType(), "SupportsModel", modelID, supports, 1*time.Hour)

	return supports

// SupportsCapability returns whether the provider supports a specific capability
func (p *CachingProvider) SupportsCapability(ctx context.Context, capability core.ModelCapability) bool {
	// Check cache
	if cached, ok := p.cache.Get(p.GetType(), "SupportsCapability", capability); ok {
		return cached.(bool)
	}

	// Call provider
	supports := p.provider.SupportsCapability(ctx, capability)

	// Cache response
	p.cache.Set(p.GetType(), "SupportsCapability", capability, supports, 1*time.Hour)

	return supports

// Close closes the provider and releases any resources
func (p *CachingProvider) Close() error {
	return p.provider.Close()
