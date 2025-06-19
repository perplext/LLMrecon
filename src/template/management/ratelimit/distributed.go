// Package ratelimit provides rate limiting functionality for template execution
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// DistributedLimiter implements a distributed rate limiter using Redis
// as a backend for coordination across multiple instances.
//
// This limiter uses Redis to maintain shared state across multiple instances,
// allowing for coordinated rate limiting in distributed environments.
// It implements the same interface as AdaptiveLimiter, making it a drop-in
// replacement for applications that need distributed rate limiting.
type DistributedLimiter struct {
	// Redis client for distributed coordination
	client redis.UniversalClient
	
	// Key prefix for Redis keys
	keyPrefix string
	
	// Default rate limit for users without a specific policy
	defaultUserLimit rate.Limit
	
	// Default burst size for users without a specific policy
	defaultUserBurst int
	
	// User policies define custom limits and priorities for each user
	userPolicies map[string]*UserRateLimitPolicy
	
	// Mutex for concurrent access to limiter state
	mu sync.RWMutex
	
	// Statistics collector for monitoring and debugging
	stats *StatsCollector
	
	// Whether to collect statistics
	statsEnabled bool
	
	// Script for atomic token bucket operations in Redis
	tokenBucketScript *redis.Script
	
	// Local adaptive limiter for fallback if Redis is unavailable
	localLimiter *AdaptiveLimiter
	
	// Whether to use local fallback if Redis is unavailable
	localFallbackEnabled bool
}

// NewDistributedLimiter creates a new distributed rate limiter using Redis
func NewDistributedLimiter(redisClient redis.UniversalClient, keyPrefix string, globalQPS float64, globalBurst int, defaultUserQPS float64, defaultUserBurst int) *DistributedLimiter {
	// Create a local adaptive limiter for fallback
	localLimiter := NewAdaptiveLimiter(globalQPS, globalBurst, defaultUserQPS, defaultUserBurst)
	
	// Create the token bucket Lua script for Redis
	// This script implements a token bucket algorithm atomically in Redis
	tokenBucketScript := redis.NewScript(`
		local key = KEYS[1]
		local tokens_key = key .. ":tokens"
		local timestamp_key = key .. ":ts"
		local rate = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local requested = tonumber(ARGV[4])
		
		-- Get the current token count or initialize it
		local tokens = tonumber(redis.call("get", tokens_key))
		if tokens == nil then
			tokens = burst
		end
		
		-- Get the last refill timestamp or initialize it
		local last_refill = tonumber(redis.call("get", timestamp_key))
		if last_refill == nil then
			last_refill = 0
		end
		
		-- Calculate time since last refill in seconds
		local elapsed = now - last_refill
		if elapsed > 0 then
			-- Refill tokens based on rate and elapsed time
			tokens = math.min(burst, tokens + (rate * elapsed))
			redis.call("set", timestamp_key, now)
		end
		
		-- Check if we have enough tokens
		local allowed = 0
		if tokens >= requested then
			-- Consume tokens
			tokens = tokens - requested
			allowed = 1
		end
		
		-- Store updated token count with TTL of 2 minutes
		redis.call("set", tokens_key, tokens)
		redis.call("expire", tokens_key, 120)
		redis.call("expire", timestamp_key, 120)
		
		return {allowed, tokens}
	`)
	
	return &DistributedLimiter{
		client:              redisClient,
		keyPrefix:           keyPrefix,
		defaultUserLimit:    rate.Limit(defaultUserQPS),
		defaultUserBurst:    defaultUserBurst,
		userPolicies:        make(map[string]*UserRateLimitPolicy),
		stats:               NewStatsCollector(1000), // Keep the last 1000 events
		statsEnabled:        true,
		tokenBucketScript:   tokenBucketScript,
		localLimiter:        localLimiter,
		localFallbackEnabled: true,
	}
}

// Acquire acquires a token from the global limiter
func (l *DistributedLimiter) Acquire(ctx context.Context) error {
	startTime := time.Now()
	
	// Try to acquire a token from Redis
	allowed, err := l.tryAcquireRedis(ctx, "global", float64(l.defaultUserLimit), l.defaultUserBurst, 1)
	
	if err != nil {
		// Redis error, fall back to local limiter if enabled
		if l.localFallbackEnabled {
			if l.statsEnabled {
				l.stats.RecordEvent(RateLimitEvent{
					Type:         "redis_error",
					UserID:       "global",
					Priority:     0,
					Timestamp:    time.Now(),
					WaitDuration: time.Since(startTime),
					ErrorMessage: err.Error(),
				})
			}
			return l.localLimiter.Acquire(ctx)
		}
		return fmt.Errorf("redis error: %w", err)
	}
	
	if !allowed {
		// Record the rejection event if stats are enabled
		if l.statsEnabled {
			l.stats.RecordEvent(RateLimitEvent{
				Type:         EventTypeGlobalLimitExceed,
				UserID:       "global",
				Priority:     0,
				Timestamp:    time.Now(),
				WaitDuration: time.Since(startTime),
				ErrorMessage: "global rate limit exceeded",
			})
		}
		return fmt.Errorf("global rate limit exceeded")
	}
	
	// Record the successful acquisition if stats are enabled
	if l.statsEnabled {
		l.stats.RecordEvent(RateLimitEvent{
			Type:         EventTypeAcquire,
			UserID:       "global",
			Priority:     0,
			Timestamp:    time.Now(),
			WaitDuration: time.Since(startTime),
		})
	}
	
	return nil
}

// AcquireForUser acquires a token for a specific user
func (l *DistributedLimiter) AcquireForUser(ctx context.Context, userID string) error {
	startTime := time.Now()
	
	// First, check global limit
	if err := l.Acquire(ctx); err != nil {
		return err
	}
	
	// Get user policy
	l.mu.RLock()
	policy, exists := l.userPolicies[userID]
	l.mu.RUnlock()
	
	var qps float64
	var burst int
	var priority int
	
	if exists {
		qps = policy.QPS
		burst = policy.Burst
		priority = policy.Priority
	} else {
		qps = float64(l.defaultUserLimit)
		burst = l.defaultUserBurst
		priority = 1
	}
	
	// Try to acquire a token from Redis
	allowed, err := l.tryAcquireRedis(ctx, userID, qps, burst, 1)
	
	if err != nil {
		// Redis error, fall back to local limiter if enabled
		if l.localFallbackEnabled {
			if l.statsEnabled {
				l.stats.RecordEvent(RateLimitEvent{
					Type:         "redis_error",
					UserID:       userID,
					Priority:     priority,
					Timestamp:    time.Now(),
					WaitDuration: time.Since(startTime),
					ErrorMessage: err.Error(),
				})
			}
			return l.localLimiter.AcquireForUser(ctx, userID)
		}
		return fmt.Errorf("redis error: %w", err)
	}
	
	if !allowed {
		// Record the rejection event if stats are enabled
		if l.statsEnabled {
			l.stats.RecordEvent(RateLimitEvent{
				Type:         EventTypeUserLimitExceed,
				UserID:       userID,
				Priority:     priority,
				Timestamp:    time.Now(),
				WaitDuration: time.Since(startTime),
				ErrorMessage: "user rate limit exceeded",
			})
		}
		return fmt.Errorf("user rate limit exceeded for %s", userID)
	}
	
	// Record the successful acquisition if stats are enabled
	if l.statsEnabled {
		l.stats.RecordEvent(RateLimitEvent{
			Type:         EventTypeAcquire,
			UserID:       userID,
			Priority:     priority,
			Timestamp:    time.Now(),
			WaitDuration: time.Since(startTime),
		})
	}
	
	return nil
}

// tryAcquireRedis attempts to acquire tokens from Redis using the token bucket algorithm
func (l *DistributedLimiter) tryAcquireRedis(ctx context.Context, key string, rate float64, burst int, tokens int) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	now := float64(time.Now().Unix())
	
	// Execute the token bucket script
	result, err := l.tokenBucketScript.Run(ctx, l.client, []string{redisKey}, rate, burst, now, tokens).Result()
	if err != nil {
		return false, err
	}
	
	// Parse the result
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) < 1 {
		return false, fmt.Errorf("unexpected result from Redis: %v", result)
	}
	
	allowed, ok := resultArray[0].(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type from Redis: %T", resultArray[0])
	}
	
	return allowed == 1, nil
}

// Release releases a token (no-op for token bucket)
func (l *DistributedLimiter) Release() {
	// No-op for token bucket
}

// ReleaseForUser releases a token for a specific user (no-op for token bucket)
func (l *DistributedLimiter) ReleaseForUser(userID string) {
	// No-op for token bucket
}

// SetUserPolicy sets a rate limit policy for a specific user
func (l *DistributedLimiter) SetUserPolicy(policy *UserRateLimitPolicy) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.userPolicies[policy.UserID] = policy
	
	// Also set in local limiter for fallback
	if l.localFallbackEnabled {
		l.localLimiter.SetUserPolicy(policy)
	}
}

// GetUserPolicy gets the rate limit policy for a specific user
func (l *DistributedLimiter) GetUserPolicy(userID string) *UserRateLimitPolicy {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.userPolicies[userID]
}

// EnableLocalFallback enables or disables local fallback
func (l *DistributedLimiter) EnableLocalFallback(enabled bool) {
	l.localFallbackEnabled = enabled
}

// GetStats returns statistics about the rate limiter
func (l *DistributedLimiter) GetStats() map[string]interface{} {
	if !l.statsEnabled {
		return map[string]interface{}{
			"stats_enabled": false,
		}
	}
	
	stats := l.stats.GetStats()
	stats["local_fallback_enabled"] = l.localFallbackEnabled
	
	// Add Redis info if available
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	info, err := l.client.Info(ctx).Result()
	if err == nil {
		stats["redis_info"] = info
	}
	
	return stats
}

// FlushRedis flushes all rate limit data from Redis
// This should be used with caution, as it will reset all rate limits
func (l *DistributedLimiter) FlushRedis(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", l.keyPrefix)
	
	// Scan for keys matching the pattern
	iter := l.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		
		// Delete in batches of 100 to avoid large commands
		if len(keys) >= 100 {
			if err := l.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	
	// Delete any remaining keys
	if len(keys) > 0 {
		if err := l.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	
	if err := iter.Err(); err != nil {
		return err
	}
	
	return nil
}

// Close closes the distributed limiter and releases resources
func (l *DistributedLimiter) Close() error {
	return l.client.Close()
}
