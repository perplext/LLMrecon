// Package ratelimit provides rate limiting functionality for template execution.
//
// This package implements both simple token bucket rate limiting and advanced
// adaptive rate limiting with priority-based fairness mechanisms. It can be used
// to control the rate of template executions and prevent system overload while
// ensuring critical operations continue to function during high load periods.
//
// See the README.md and doc.go files for detailed usage examples and best practices.
package ratelimit

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// TokenBucketLimiter implements a simple token bucket rate limiter.
//
// This limiter provides basic rate limiting functionality using the token bucket algorithm.
// It supports both global rate limiting (applied to all requests) and user-specific rate
// limiting (applied to individual users). The token bucket algorithm allows for bursts of
// traffic up to a configurable limit while maintaining the desired average rate.
//
// For more advanced rate limiting with priority-based fairness, see AdaptiveLimiter.
type TokenBucketLimiter struct {
	// Global limiter controls the overall system throughput
	globalLimiter *rate.Limiter
	
	// User-specific limiters control per-user throughput
	userLimiters map[string]*rate.Limiter
	
	// Default rate limit for new users without a specific limit
	defaultUserLimit rate.Limit
	
	// Default burst size for new users without a specific limit
	defaultUserBurst int
	
	// Mutex for concurrent access to user limiters
	mu sync.RWMutex

// NewTokenBucketLimiter creates a new token bucket rate limiter with the specified parameters.
//
// Parameters:
//   - globalQPS: The global queries per second limit for all requests combined
//   - globalBurst: The maximum burst size for the global limiter
//   - defaultUserQPS: The default queries per second limit for individual users
//   - defaultUserBurst: The default burst size for individual users
//
// This constructor creates a basic rate limiter without priority-based fairness.
// For advanced rate limiting with fairness mechanisms, use NewAdaptiveLimiter instead.
//
// Example:
//
//     limiter := NewTokenBucketLimiter(100, 50, 10, 5)
//
// This creates a limiter with a global limit of 100 QPS with bursts up to 50,
// and a default per-user limit of 10 QPS with bursts up to 5.
func NewTokenBucketLimiter(globalQPS float64, globalBurst int, defaultUserQPS float64, defaultUserBurst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		globalLimiter:    rate.NewLimiter(rate.Limit(globalQPS), globalBurst),
		userLimiters:     make(map[string]*rate.Limiter),
		defaultUserLimit: rate.Limit(defaultUserQPS),
		defaultUserBurst: defaultUserBurst,
	}

// Acquire acquires a token from the global limiter.
//
// This method blocks until a token is available or the context is canceled.
// It applies only the global rate limit, not user-specific limits.
//
// Parameters:
//   - ctx: A context.Context that can be used to cancel the wait
//
// Returns:
//   - error: An error if the context is canceled before a token is acquired,
//     or nil if a token was successfully acquired
//
// Example:
//
//     ctx := context.Background()
//     err := limiter.Acquire(ctx)
//     if err != nil {
//         // Handle rate limit exceeded error
//         return err
//     }
//
// For user-specific rate limiting, use AcquireForUser instead.
func (l *TokenBucketLimiter) Acquire(ctx context.Context) error {
	if err := l.globalLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("global rate limit exceeded: %w", err)
	}
	return nil

// AcquireForUser acquires a token for a specific user, applying both global and user-specific limits.
//
// This method blocks until a token is available from both the global limiter and
// the user-specific limiter, or until the context is canceled. It first checks the
// global limit, then the user-specific limit.
//
// Parameters:
//   - ctx: A context.Context that can be used to cancel the wait
//   - userID: The ID of the user to acquire a token for
//
// Returns:
//   - error: An error if either the global or user-specific rate limit is exceeded,
//     or nil if a token was successfully acquired
//
// Example:
//
//     ctx := context.Background()
//     err := limiter.AcquireForUser(ctx, "user123")
//     if err != nil {
//         // Handle rate limit exceeded error
//         return err
//     }
//
// Note that this method in TokenBucketLimiter does not implement priority-based fairness.
// For priority-based fairness, use AdaptiveLimiter instead.
func (l *TokenBucketLimiter) AcquireForUser(ctx context.Context, userID string) error {
	// First, check global limit
	if err := l.Acquire(ctx); err != nil {
		return err
	}
	
	// Get user-specific limiter
	limiter := l.getUserLimiter(userID)
	
	// Check user-specific limit
	if err := limiter.Wait(ctx); err != nil {
		return fmt.Errorf("user rate limit exceeded for %s: %w", userID, err)
	}
	
	return nil

// Release releases a token (no-op for token bucket)
func (l *TokenBucketLimiter) Release() {
	// No-op for token bucket limiter

// ReleaseForUser releases a token for a specific user (no-op for token bucket)
func (l *TokenBucketLimiter) ReleaseForUser(userID string) {
	// No-op for token bucket limiter

// getUserLimiter gets or creates a rate limiter for a specific user
func (l *TokenBucketLimiter) getUserLimiter(userID string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.userLimiters[userID]
	l.mu.RUnlock()
	
	if exists {
		return limiter
	}
	
	// Create new limiter for this user
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Double-check to avoid race condition
	if limiter, exists = l.userLimiters[userID]; exists {
		return limiter
	}
	
	limiter = rate.NewLimiter(l.defaultUserLimit, l.defaultUserBurst)
	l.userLimiters[userID] = limiter
	return limiter

// GetLimit returns the current global rate limit
func (l *TokenBucketLimiter) GetLimit() int {
	return l.globalLimiter.Burst()

// GetUserLimit returns the current rate limit for a specific user
func (l *TokenBucketLimiter) GetUserLimit(userID string) int {
	limiter := l.getUserLimiter(userID)
	return limiter.Burst()

// SetLimit sets the global rate limit
func (l *TokenBucketLimiter) SetLimit(limit int) {
	l.globalLimiter.SetBurst(limit)

// SetUserLimit sets the rate limit for a specific user
func (l *TokenBucketLimiter) SetUserLimit(userID string, limit int) {
	limiter := l.getUserLimiter(userID)
	limiter.SetBurst(limit)

// SetGlobalQPS sets the global queries per second limit
func (l *TokenBucketLimiter) SetGlobalQPS(qps float64) {
	l.globalLimiter.SetLimit(rate.Limit(qps))

// SetUserQPS sets the queries per second limit for a specific user
func (l *TokenBucketLimiter) SetUserQPS(userID string, qps float64) {
	limiter := l.getUserLimiter(userID)
	limiter.SetLimit(rate.Limit(qps))

// GetGlobalQPS gets the global queries per second limit
func (l *TokenBucketLimiter) GetGlobalQPS() float64 {
	return float64(l.globalLimiter.Limit())

// GetUserQPS gets the queries per second limit for a specific user
func (l *TokenBucketLimiter) GetUserQPS(userID string) float64 {
	limiter := l.getUserLimiter(userID)
	return float64(limiter.Limit())

// GetUserLimiters returns all user limiters
func (l *TokenBucketLimiter) GetUserLimiters() map[string]float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	result := make(map[string]float64, len(l.userLimiters))
	for userID, limiter := range l.userLimiters {
		result[userID] = float64(limiter.Limit())
	}
	
	return result

// ResetUserLimiter resets the rate limiter for a specific user
func (l *TokenBucketLimiter) ResetUserLimiter(userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	delete(l.userLimiters, userID)

// ResetAllUserLimiters resets all user-specific rate limiters
func (l *TokenBucketLimiter) ResetAllUserLimiters() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
