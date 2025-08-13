package performance

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// DistributedRateLimiter implements distributed rate limiting using Redis
type DistributedRateLimiter struct {
	client     *redis.Client
	config     DistributedRateLimitConfig
	scripts    *RateLimitScripts
	logger     Logger
	metrics    *RateLimitMetrics
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// DistributedRateLimitConfig defines configuration for distributed rate limiting
type DistributedRateLimitConfig struct {
	// Redis connection
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	
	// Rate limiting configuration
	KeyPrefix           string        `json:"key_prefix"`
	DefaultLimit        int64         `json:"default_limit"`
	DefaultWindow       time.Duration `json:"default_window"`
	DefaultBurst        int64         `json:"default_burst"`
	
	// Algorithm settings
	Algorithm           RateLimitAlgorithm `json:"algorithm"`
	SlidingWindowParts  int                `json:"sliding_window_parts"`
	
	// Cleanup and maintenance
	CleanupInterval     time.Duration `json:"cleanup_interval"`
	KeyExpiration       time.Duration `json:"key_expiration"`
	
	// Performance settings
	EnablePipelining    bool          `json:"enable_pipelining"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	
	// Monitoring
	EnableMetrics       bool          `json:"enable_metrics"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
}

// RateLimitAlgorithm defines the rate limiting algorithm
type RateLimitAlgorithm string

const (
	AlgorithmTokenBucket    RateLimitAlgorithm = "token_bucket"
	AlgorithmSlidingWindow  RateLimitAlgorithm = "sliding_window"
	AlgorithmFixedWindow    RateLimitAlgorithm = "fixed_window"
	AlgorithmLeakyBucket    RateLimitAlgorithm = "leaky_bucket"
)

// RateLimitRequest represents a rate limit check request
type RateLimitRequest struct {
	Key       string        `json:"key"`
	Limit     int64         `json:"limit"`
	Window    time.Duration `json:"window"`
	Burst     int64         `json:"burst"`
	Cost      int64         `json:"cost"`
	Timestamp time.Time     `json:"timestamp"`
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool          `json:"allowed"`
	Remaining     int64         `json:"remaining"`
	ResetTime     time.Time     `json:"reset_time"`
	RetryAfter    time.Duration `json:"retry_after"`
	TotalLimit    int64         `json:"total_limit"`
	WindowSize    time.Duration `json:"window_size"`
	CurrentUsage  int64         `json:"current_usage"`
}

// RateLimitMetrics tracks rate limiting performance
type RateLimitMetrics struct {
	TotalRequests      int64 `json:"total_requests"`
	AllowedRequests    int64 `json:"allowed_requests"`
	DeniedRequests     int64 `json:"denied_requests"`
	ActiveKeys         int64 `json:"active_keys"`
	RedisOperations    int64 `json:"redis_operations"`
	RedisErrors        int64 `json:"redis_errors"`
	AverageLatency     time.Duration `json:"average_latency"`
	AllowRate          float64 `json:"allow_rate"`
}

// RateLimitScripts contains Lua scripts for atomic Redis operations
type RateLimitScripts struct {
	TokenBucket    *redis.Script
	SlidingWindow  *redis.Script
	FixedWindow    *redis.Script
	LeakyBucket    *redis.Script
	Cleanup        *redis.Script
}

// DefaultDistributedRateLimitConfig returns default configuration
func DefaultDistributedRateLimitConfig() DistributedRateLimitConfig {
	return DistributedRateLimitConfig{
		RedisAddr:          "localhost:6379",
		RedisPassword:      "",
		RedisDB:            0,
		KeyPrefix:          "ratelimit",
		DefaultLimit:       100,
		DefaultWindow:      time.Minute,
		DefaultBurst:       10,
		Algorithm:          AlgorithmTokenBucket,
		SlidingWindowParts: 10,
		CleanupInterval:    5 * time.Minute,
		KeyExpiration:      time.Hour,
		EnablePipelining:   true,
		MaxRetries:         3,
		RetryDelay:         100 * time.Millisecond,
		EnableMetrics:      true,
		MetricsInterval:    30 * time.Second,
	}
}

// NewDistributedRateLimiter creates a new distributed rate limiter
func NewDistributedRateLimiter(config DistributedRateLimitConfig, logger Logger) (*DistributedRateLimiter, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
	
	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	// Initialize Lua scripts
	scripts := &RateLimitScripts{
		TokenBucket:   redis.NewScript(tokenBucketScript),
		SlidingWindow: redis.NewScript(slidingWindowScript),
		FixedWindow:   redis.NewScript(fixedWindowScript),
		LeakyBucket:   redis.NewScript(leakyBucketScript),
		Cleanup:       redis.NewScript(cleanupScript),
	}
	
	limiter := &DistributedRateLimiter{
		client:  client,
		config:  config,
		scripts: scripts,
		logger:  logger,
		metrics: &RateLimitMetrics{},
		ctx:     ctx,
		cancel:  cancel,
	}
	
	return limiter, nil
}

// Start starts the distributed rate limiter
func (d *DistributedRateLimiter) Start() error {
	d.logger.Info("Starting distributed rate limiter", "algorithm", d.config.Algorithm)
	
	// Start cleanup loop
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.cleanupLoop()
	}()
	
	// Start metrics collection if enabled
	if d.config.EnableMetrics {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			d.metricsLoop()
		}()
	}
	
	d.logger.Info("Distributed rate limiter started")
	return nil
}

// Stop stops the distributed rate limiter
func (d *DistributedRateLimiter) Stop() error {
	d.logger.Info("Stopping distributed rate limiter")
	
	d.cancel()
	d.wg.Wait()
	
	// Close Redis connection
	if err := d.client.Close(); err != nil {
		d.logger.Error("Error closing Redis connection", "error", err)
		return err
	}
	
	d.logger.Info("Distributed rate limiter stopped")
	return nil
}

// Allow checks if a request is allowed under the rate limit
func (d *DistributedRateLimiter) Allow(request *RateLimitRequest) (*RateLimitResult, error) {
	start := time.Now()
	d.metrics.TotalRequests++
	
	// Set defaults if not provided
	if request.Limit == 0 {
		request.Limit = d.config.DefaultLimit
	}
	if request.Window == 0 {
		request.Window = d.config.DefaultWindow
	}
	if request.Burst == 0 {
		request.Burst = d.config.DefaultBurst
	}
	if request.Cost == 0 {
		request.Cost = 1
	}
	if request.Timestamp.IsZero() {
		request.Timestamp = time.Now()
	}
	
	// Execute rate limiting algorithm
	result, err := d.executeRateLimit(request)
	if err != nil {
		d.metrics.RedisErrors++
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}
	
	// Update metrics
	d.updateRequestMetrics(result, time.Since(start))
	
	d.logger.Debug("Rate limit check", 
		"key", request.Key,
		"allowed", result.Allowed,
		"remaining", result.Remaining,
		"usage", result.CurrentUsage,
		"limit", result.TotalLimit,
	)
	
	return result, nil
}

// AllowN checks if N requests are allowed under the rate limit
func (d *DistributedRateLimiter) AllowN(key string, n int64) (*RateLimitResult, error) {
	request := &RateLimitRequest{
		Key:  key,
		Cost: n,
	}
	return d.Allow(request)
}

// Reset resets the rate limit for a key
func (d *DistributedRateLimiter) Reset(key string) error {
	fullKey := d.getRedisKey(key)
	
	err := d.client.Del(d.ctx, fullKey).Err()
	if err != nil {
		d.metrics.RedisErrors++
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}
	
	d.logger.Info("Rate limit reset", "key", key)
	return nil
}

// GetStatus returns the current status for a key without consuming quota
func (d *DistributedRateLimiter) GetStatus(key string) (*RateLimitResult, error) {
	request := &RateLimitRequest{
		Key:  key,
		Cost: 0, // Don't consume quota
	}
	return d.Allow(request)
}

// GetMetrics returns current rate limiting metrics
func (d *DistributedRateLimiter) GetMetrics() *RateLimitMetrics {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	// Calculate allow rate
	if d.metrics.TotalRequests > 0 {
		d.metrics.AllowRate = float64(d.metrics.AllowedRequests) / float64(d.metrics.TotalRequests)
	}
	
	return d.metrics
}

// Private methods

// executeRateLimit executes the rate limiting algorithm
func (d *DistributedRateLimiter) executeRateLimit(request *RateLimitRequest) (*RateLimitResult, error) {
	d.metrics.RedisOperations++
	
	switch d.config.Algorithm {
	case AlgorithmTokenBucket:
		return d.executeTokenBucket(request)
	case AlgorithmSlidingWindow:
		return d.executeSlidingWindow(request)
	case AlgorithmFixedWindow:
		return d.executeFixedWindow(request)
	case AlgorithmLeakyBucket:
		return d.executeLeakyBucket(request)
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", d.config.Algorithm)
	}
}

// executeTokenBucket implements token bucket algorithm
func (d *DistributedRateLimiter) executeTokenBucket(request *RateLimitRequest) (*RateLimitResult, error) {
	key := d.getRedisKey(request.Key)
	now := request.Timestamp.Unix()
	
	// Arguments: [key, limit, window_seconds, burst, cost, now]
	args := []interface{}{
		request.Limit,
		int64(request.Window.Seconds()),
		request.Burst,
		request.Cost,
		now,
	}
	
	result, err := d.scripts.TokenBucket.Run(d.ctx, d.client, []string{key}, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("token bucket script failed: %w", err)
	}
	
	return d.parseScriptResult(result, request)
}

// executeSlidingWindow implements sliding window algorithm
func (d *DistributedRateLimiter) executeSlidingWindow(request *RateLimitRequest) (*RateLimitResult, error) {
	key := d.getRedisKey(request.Key)
	now := request.Timestamp.Unix()
	windowStart := now - int64(request.Window.Seconds())
	
	// Arguments: [key, limit, window_start, now, cost, parts]
	args := []interface{}{
		request.Limit,
		windowStart,
		now,
		request.Cost,
		d.config.SlidingWindowParts,
	}
	
	result, err := d.scripts.SlidingWindow.Run(d.ctx, d.client, []string{key}, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("sliding window script failed: %w", err)
	}
	
	return d.parseScriptResult(result, request)
}

// executeFixedWindow implements fixed window algorithm
func (d *DistributedRateLimiter) executeFixedWindow(request *RateLimitRequest) (*RateLimitResult, error) {
	key := d.getRedisKey(request.Key)
	now := request.Timestamp.Unix()
	window := int64(request.Window.Seconds())
	windowStart := (now / window) * window
	
	// Arguments: [key, limit, window_start, window_end, cost]
	args := []interface{}{
		request.Limit,
		windowStart,
		windowStart + window,
		request.Cost,
	}
	
	result, err := d.scripts.FixedWindow.Run(d.ctx, d.client, []string{key}, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("fixed window script failed: %w", err)
	}
	
	return d.parseScriptResult(result, request)
}

// executeLeakyBucket implements leaky bucket algorithm
func (d *DistributedRateLimiter) executeLeakyBucket(request *RateLimitRequest) (*RateLimitResult, error) {
	key := d.getRedisKey(request.Key)
	now := request.Timestamp.Unix()
	
	// Arguments: [key, capacity, leak_rate, cost, now]
	leakRate := float64(request.Limit) / request.Window.Seconds()
	args := []interface{}{
		request.Limit, // capacity
		leakRate,
		request.Cost,
		now,
	}
	
	result, err := d.scripts.LeakyBucket.Run(d.ctx, d.client, []string{key}, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("leaky bucket script failed: %w", err)
	}
	
	return d.parseScriptResult(result, request)
}

// parseScriptResult parses the result from Lua scripts
func (d *DistributedRateLimiter) parseScriptResult(result interface{}, request *RateLimitRequest) (*RateLimitResult, error) {
	values, ok := result.([]interface{})
	if !ok || len(values) < 4 {
		return nil, fmt.Errorf("invalid script result format")
	}
	
	// Parse script results: [allowed, remaining, reset_time, current_usage]
	allowed, _ := strconv.ParseInt(fmt.Sprintf("%v", values[0]), 10, 64)
	remaining, _ := strconv.ParseInt(fmt.Sprintf("%v", values[1]), 10, 64)
	resetTime, _ := strconv.ParseInt(fmt.Sprintf("%v", values[2]), 10, 64)
	currentUsage, _ := strconv.ParseInt(fmt.Sprintf("%v", values[3]), 10, 64)
	
	resetTimestamp := time.Unix(resetTime, 0)
	var retryAfter time.Duration
	if !resetTimestamp.IsZero() && resetTimestamp.After(time.Now()) {
		retryAfter = resetTimestamp.Sub(time.Now())
	}
	
	return &RateLimitResult{
		Allowed:      allowed == 1,
		Remaining:    remaining,
		ResetTime:    resetTimestamp,
		RetryAfter:   retryAfter,
		TotalLimit:   request.Limit,
		WindowSize:   request.Window,
		CurrentUsage: currentUsage,
	}, nil
}

// getRedisKey returns the full Redis key for rate limiting
func (d *DistributedRateLimiter) getRedisKey(key string) string {
	return fmt.Sprintf("%s:%s", d.config.KeyPrefix, key)
}

// updateRequestMetrics updates request-level metrics
func (d *DistributedRateLimiter) updateRequestMetrics(result *RateLimitResult, latency time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	if result.Allowed {
		d.metrics.AllowedRequests++
	} else {
		d.metrics.DeniedRequests++
	}
	
	// Update average latency
	if d.metrics.AverageLatency == 0 {
		d.metrics.AverageLatency = latency
	} else {
		d.metrics.AverageLatency = (d.metrics.AverageLatency + latency) / 2
	}
}

// cleanupLoop performs periodic cleanup of expired keys
func (d *DistributedRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(d.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			d.performCleanup()
		case <-d.ctx.Done():
			return
		}
	}
}

// performCleanup removes expired rate limit keys
func (d *DistributedRateLimiter) performCleanup() {
	pattern := d.config.KeyPrefix + ":*"
	expireTime := time.Now().Add(-d.config.KeyExpiration).Unix()
	
	// Arguments: [pattern, expire_time]
	args := []interface{}{pattern, expireTime}
	
	result, err := d.scripts.Cleanup.Run(d.ctx, d.client, []string{}, args...).Result()
	if err != nil {
		d.logger.Error("Cleanup script failed", "error", err)
		d.metrics.RedisErrors++
		return
	}
	
	cleaned, _ := strconv.ParseInt(fmt.Sprintf("%v", result), 10, 64)
	if cleaned > 0 {
		d.logger.Info("Cleaned up expired rate limit keys", "count", cleaned)
	}
}

// metricsLoop periodically updates metrics
func (d *DistributedRateLimiter) metricsLoop() {
	ticker := time.NewTicker(d.config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			d.updateMetrics()
		case <-d.ctx.Done():
			return
		}
	}
}

// updateMetrics updates system-level metrics
func (d *DistributedRateLimiter) updateMetrics() {
	// Count active keys
	pattern := d.config.KeyPrefix + ":*"
	keys, err := d.client.Keys(d.ctx, pattern).Result()
	if err != nil {
		d.logger.Error("Failed to count active keys", "error", err)
		d.metrics.RedisErrors++
		return
	}
	
	d.mutex.Lock()
	d.metrics.ActiveKeys = int64(len(keys))
	d.mutex.Unlock()
}

// Lua scripts for atomic operations

const tokenBucketScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local burst = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local now = tonumber(ARGV[5])

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1]) or limit
local last_refill = tonumber(bucket[2]) or now

-- Calculate tokens to add based on time elapsed
local elapsed = now - last_refill
local tokens_to_add = math.floor(elapsed * limit / window)
tokens = math.min(limit, tokens + tokens_to_add)

local allowed = 0
local remaining = tokens
local reset_time = now + window

if cost <= tokens then
    tokens = tokens - cost
    allowed = 1
    remaining = tokens
end

-- Update bucket state
redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
redis.call('EXPIRE', key, window * 2)

return {allowed, remaining, reset_time, limit - tokens}
`

const slidingWindowScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_start = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])
local parts = tonumber(ARGV[5])

-- Clean old entries
redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

-- Count current usage
local current = redis.call('ZCARD', key)
local allowed = 0
local remaining = limit - current

if cost <= remaining then
    -- Add new entry
    for i = 1, cost do
        redis.call('ZADD', key, now, now .. ':' .. i)
    end
    allowed = 1
    remaining = remaining - cost
    current = current + cost
end

redis.call('EXPIRE', key, (now - window_start) + 60)

return {allowed, remaining, now + (now - window_start), current}
`

const fixedWindowScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_start = tonumber(ARGV[2])
local window_end = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])

local window_key = key .. ':' .. window_start
local current = tonumber(redis.call('GET', window_key)) or 0
local allowed = 0
local remaining = limit - current

if cost <= remaining then
    current = redis.call('INCRBY', window_key, cost)
    allowed = 1
    remaining = limit - current
    redis.call('EXPIRE', window_key, window_end - window_start + 60)
end

return {allowed, remaining, window_end, current}
`

const leakyBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local leak_rate = tonumber(ARGV[2])
local cost = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local bucket = redis.call('HMGET', key, 'volume', 'last_leak')
local volume = tonumber(bucket[1]) or 0
local last_leak = tonumber(bucket[2]) or now

-- Calculate volume leaked since last check
local elapsed = now - last_leak
local leaked = elapsed * leak_rate
volume = math.max(0, volume - leaked)

local allowed = 0
local remaining = capacity - volume

if cost <= remaining then
    volume = volume + cost
    allowed = 1
    remaining = capacity - volume
end

-- Update bucket state
redis.call('HMSET', key, 'volume', volume, 'last_leak', now)
redis.call('EXPIRE', key, capacity / leak_rate + 60)

-- Calculate reset time (when bucket will be empty)
local reset_time = now + (volume / leak_rate)

return {allowed, remaining, reset_time, volume}
`

const cleanupScript = `
local pattern = ARGV[1]
local expire_time = tonumber(ARGV[2])
local keys = redis.call('KEYS', pattern)
local cleaned = 0

for i = 1, #keys do
    local key = keys[i]
    local ttl = redis.call('TTL', key)
    if ttl > 0 and ttl < (expire_time - redis.call('TIME')[1]) then
        redis.call('DEL', key)
        cleaned = cleaned + 1
    end
end

return cleaned
`