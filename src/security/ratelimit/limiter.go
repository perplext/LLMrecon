package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu        sync.Mutex
	tokens    float64
	maxTokens float64
	refillRate float64
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

// AllowN checks if n requests are allowed
func (rl *RateLimiter) AllowN(n float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= n {
		rl.tokens -= n
		return true
	}

	return false
}

// refill adds tokens based on time elapsed
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	tokensToAdd := elapsed * rl.refillRate

	rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
	rl.lastRefill = now
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Check again
		}
	}
}

// IPRateLimiter manages rate limits per IP address
type IPRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*RateLimiter
	maxTokens float64
	refillRate float64
	cleanupInterval time.Duration
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(maxTokens, refillRate float64) *IPRateLimiter {
	rl := &IPRateLimiter{
		limiters:        make(map[string]*RateLimiter),
		maxTokens:       maxTokens,
		refillRate:      refillRate,
		cleanupInterval: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from an IP is allowed
func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = NewRateLimiter(rl.maxTokens, rl.refillRate)
		rl.limiters[ip] = limiter
	}
	rl.mu.Unlock()

	return limiter.Allow()
}

// cleanup removes inactive rate limiters
func (rl *IPRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// In a production system, we'd track last access time
		// and remove limiters that haven't been used recently
		// For now, we'll keep this simple
		if len(rl.limiters) > 10000 {
			// Clear if too many entries (prevent memory leak)
			rl.limiters = make(map[string]*RateLimiter)
		}
		rl.mu.Unlock()
	}
}

// APIKeyRateLimiter manages rate limits per API key
type APIKeyRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*RateLimiter
	limits   map[string]RateLimitConfig
}

// RateLimitConfig defines rate limit configuration
type RateLimitConfig struct {
	MaxTokens  float64
	RefillRate float64
	BurstSize  int
}

// NewAPIKeyRateLimiter creates a new API key-based rate limiter
func NewAPIKeyRateLimiter() *APIKeyRateLimiter {
	return &APIKeyRateLimiter{
		limiters: make(map[string]*RateLimiter),
		limits:   make(map[string]RateLimitConfig),
	}
}

// SetLimit sets the rate limit for an API key
func (rl *APIKeyRateLimiter) SetLimit(apiKey string, config RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limits[apiKey] = config
	// Reset the limiter if it exists
	delete(rl.limiters, apiKey)
}

// Allow checks if a request with an API key is allowed
func (rl *APIKeyRateLimiter) Allow(apiKey string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	config, exists := rl.limits[apiKey]
	if !exists {
		return false, fmt.Errorf("API key not configured: %s", apiKey[:min(8, len(apiKey))]+"...")
	}

	limiter, exists := rl.limiters[apiKey]
	if !exists {
		limiter = NewRateLimiter(config.MaxTokens, config.RefillRate)
		rl.limiters[apiKey] = limiter
	}

	return limiter.Allow(), nil
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}