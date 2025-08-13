// Package ratelimit provides rate limiting functionality for template execution.
//
// This package implements both global and user-specific rate limiting with
// adaptive behavior based on system load. It includes priority-based fairness
// mechanisms to ensure high-priority users receive preferential treatment
// during high contention periods.
package ratelimit

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// UserRateLimitPolicy defines a policy for user rate limiting.
// Each user can have their own rate limiting policy with custom limits and priority.
type UserRateLimitPolicy struct {
	// UserID is the ID of the user
	UserID string
	
	// QPS is the queries per second limit
	// This controls how many requests per second the user can make under normal conditions
	QPS float64
	
	// Burst is the maximum burst size
	// This allows users to occasionally exceed their QPS for short periods
	Burst int
	
	// Priority is the priority of the user (higher = more priority)
	// During high contention periods, higher priority users will be served first
	// Priority values typically range from 1 (lowest) to 10 (highest)
	Priority int
	
	// MaxTokens is the maximum number of tokens the user can consume
	// This provides a hard cap on resource usage over time
	MaxTokens int
	
	// ResetInterval is the interval at which the user's tokens are reset
	// This controls how often the user's token allocation is refreshed
	ResetInterval time.Duration
}

// AdaptiveLimiter implements an adaptive rate limiter with fairness mechanisms.
//
// The AdaptiveLimiter provides several key features:
//  1. Global rate limiting - Controls the overall system throughput
//  2. User-specific rate limiting - Controls individual user throughput
//  3. Priority-based fairness - Ensures higher priority users get preferential treatment
//  4. Dynamic adjustment - Adapts limits based on system load
//  5. Token bucket tracking - Prevents abuse by limiting total resource consumption
//  6. Statistics collection - Tracks rate limiting events for monitoring and debugging
//
// During periods of high contention (when loadFactor < 0.8), the limiter activates
// its fairness mechanisms, which queue requests and process them based on user priority.
// This ensures that high-priority operations continue to function even when the system
// is under heavy load.
type AdaptiveLimiter struct {
	// Global limiter controls the overall system throughput
	globalLimiter *rate.Limiter
	
	// User-specific limiters control per-user throughput
	userLimiters map[string]*rate.Limiter
	
	// User policies define custom limits and priorities for each user
	userPolicies map[string]*UserRateLimitPolicy
	
	// Default rate limit for new users without a specific policy
	defaultUserLimit rate.Limit
	
	// Default burst size for new users without a specific policy
	defaultUserBurst int
	
	// Usage tracking counts token consumption per user
	userUsage map[string]int
	
	// Last reset time tracks when token usage was last reset
	lastResetTime time.Time
	
	// Mutex for concurrent access to limiter state
	mu sync.RWMutex
	
	// Fairness queue for high contention periods
	fairnessEnabled bool
	
	// Priority-based request queue manages requests during high contention
	requestQueue *priorityQueue
	queueMu      sync.Mutex
	
	// Dynamic adjustment based on system load
	dynamicAdjustment bool
	
	// System load factor (1.0 = normal, <1.0 = reduce limits, >1.0 = increase limits)
	// This represents the current system capacity and affects how requests are processed
	loadFactor float64
	
	// Statistics collector for monitoring and debugging
	stats *StatsCollector
	
	// Whether to collect statistics
	statsEnabled bool
}

// NewAdaptiveLimiter creates a new adaptive rate limiter
func NewAdaptiveLimiter(globalQPS float64, globalBurst int, defaultUserQPS float64, defaultUserBurst int) *AdaptiveLimiter {
	return &AdaptiveLimiter{
		globalLimiter:     rate.NewLimiter(rate.Limit(globalQPS), globalBurst),
		userLimiters:      make(map[string]*rate.Limiter),
		userPolicies:      make(map[string]*UserRateLimitPolicy),
		defaultUserLimit:  rate.Limit(defaultUserQPS),
		defaultUserBurst:  defaultUserBurst,
		userUsage:         make(map[string]int),
		lastResetTime:     time.Now(),
		fairnessEnabled:   true,
		requestQueue:      newPriorityQueue(),
		dynamicAdjustment: true,
		loadFactor:        1.0,
		stats:             NewStatsCollector(1000), // Keep the last 1000 events
		statsEnabled:      true,
	}
}

// Acquire acquires a token from the global limiter
func (l *AdaptiveLimiter) Acquire(ctx context.Context) error {
	startTime := time.Now()
	
	if err := l.globalLimiter.Wait(ctx); err != nil {
		// Record the rejection event if stats are enabled
		if l.statsEnabled {
			l.stats.RecordEvent(RateLimitEvent{
				Type:         EventTypeGlobalLimitExceed,
				UserID:       "global",
				Priority:     0,
				Timestamp:    time.Now(),
				WaitDuration: time.Since(startTime),
				ErrorMessage: err.Error(),
				LoadFactor:   l.loadFactor,
			})
		}
		return fmt.Errorf("global rate limit exceeded: %w", err)
	}
	
	// Record the successful acquisition if stats are enabled
	if l.statsEnabled {
		l.stats.RecordEvent(RateLimitEvent{
			Type:         EventTypeAcquire,
			UserID:       "global",
			Priority:     0,
			Timestamp:    time.Now(),
			WaitDuration: time.Since(startTime),
			LoadFactor:   l.loadFactor,
		})
	}
	
	return nil
}

// AcquireForUser acquires a token for a specific user with priority-based fairness.
//
// This method implements the core rate limiting logic with the following steps:
// 1. Check global rate limit to control overall system throughput
// 2. Apply priority-based fairness during high load conditions
// 3. Apply user-specific rate limits based on their policy
// 4. Track token usage for quota enforcement
//
// The fairness mechanism activates when the system is under high load (loadFactor < 0.8).
// When activated, requests are queued and processed based on user priority, ensuring
// that high-priority users receive preferential treatment.
//
// Returns an error if any rate limit is exceeded or if the context is canceled.
func (l *AdaptiveLimiter) AcquireForUser(ctx context.Context, userID string) error {
	startTime := time.Now()
	
	// First, check global limit
	if err := l.Acquire(ctx); err != nil {
		return err
	}
	
	// Get user priority from policy
	userPriority := 1 // Default priority
	l.mu.RLock()
	policy, exists := l.userPolicies[userID]
	l.mu.RUnlock()
	if exists && policy.Priority > 0 {
		userPriority = policy.Priority
	}
	
	// Apply fairness if enabled and under high load
	if l.fairnessEnabled && l.loadFactor < 0.8 {
		// System is under high load, use priority queue
		l.queueMu.Lock()
		queueLen := l.requestQueue.Len()
		l.queueMu.Unlock()
		
		// If queue has items or we're under very high load, use priority queue
		if queueLen > 0 || l.loadFactor < 0.5 {
			// Wait for our turn in the priority queue
			queueStartTime := time.Now()
			if !l.requestQueue.waitForTurn(userID, userPriority, ctx) {
				// Context was canceled while waiting
				if l.statsEnabled {
					l.stats.RecordEvent(RateLimitEvent{
						Type:         EventTypeQueueTimeout,
						UserID:       userID,
						Priority:     userPriority,
						Timestamp:    time.Now(),
						WaitDuration: time.Since(queueStartTime),
						ErrorMessage: ctx.Err().Error(),
						LoadFactor:   l.loadFactor,
					})
				}
				return ctx.Err()
			}
		}
	}
	
	// Get user-specific limiter
	limiter := l.getUserLimiter(userID)
	
	// Apply dynamic adjustments if enabled
	if l.dynamicAdjustment {
		l.applyDynamicAdjustment(userID)
	}
	
	// Check if user has exceeded their token allocation
	if !l.checkUserTokens(userID) {
		if l.statsEnabled {
			l.stats.RecordEvent(RateLimitEvent{
				Type:         EventTypeTokensExceed,
				UserID:       userID,
				Priority:     userPriority,
				Timestamp:    time.Now(),
				WaitDuration: time.Since(startTime),
				ErrorMessage: "token allocation exceeded",
				LoadFactor:   l.loadFactor,
			})
		}
		return fmt.Errorf("user %s has exceeded their token allocation", userID)
	}
	
	// Apply rate limiting
	limitStartTime := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		if l.statsEnabled {
			l.stats.RecordEvent(RateLimitEvent{
				Type:         EventTypeUserLimitExceed,
				UserID:       userID,
				Priority:     userPriority,
				Timestamp:    time.Now(),
				WaitDuration: time.Since(limitStartTime),
				ErrorMessage: err.Error(),
				LoadFactor:   l.loadFactor,
			})
		}
		return fmt.Errorf("user rate limit exceeded for %s: %w", userID, err)
	}
	
	// Track usage
	l.trackUsage(userID)
	
	// Record successful acquisition
	if l.statsEnabled {
		l.stats.RecordEvent(RateLimitEvent{
			Type:         EventTypeAcquire,
			UserID:       userID,
			Priority:     userPriority,
			Timestamp:    time.Now(),
			WaitDuration: time.Since(startTime),
			LoadFactor:   l.loadFactor,
		})
	}
	
	return nil
}

// Release releases a token (no-op for token bucket)
func (l *AdaptiveLimiter) Release() {
	// No-op for token bucket limiter
}

// ReleaseForUser releases a token for a specific user (no-op for token bucket)
func (l *AdaptiveLimiter) ReleaseForUser(userID string) {
	// No-op for token bucket limiter
}

// getUserLimiter gets or creates a rate limiter for a specific user
func (l *AdaptiveLimiter) getUserLimiter(userID string) *rate.Limiter {
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
	
	// Check if there's a policy for this user
	policy, hasPolicyForUser := l.userPolicies[userID]
	if hasPolicyForUser {
		limiter = rate.NewLimiter(rate.Limit(policy.QPS), policy.Burst)
	} else {
		limiter = rate.NewLimiter(l.defaultUserLimit, l.defaultUserBurst)
	}
	
	l.userLimiters[userID] = limiter
	return limiter
}

// GetLimit returns the current global rate limit
func (l *AdaptiveLimiter) GetLimit() int {
	return l.globalLimiter.Burst()
}

// GetUserLimit returns the current rate limit for a specific user
func (l *AdaptiveLimiter) GetUserLimit(userID string) int {
	limiter := l.getUserLimiter(userID)
	return limiter.Burst()
}

// SetLimit sets the global rate limit
func (l *AdaptiveLimiter) SetLimit(limit int) {
	l.globalLimiter.SetBurst(limit)
}

// SetUserLimit sets the rate limit for a specific user
func (l *AdaptiveLimiter) SetUserLimit(userID string, limit int) {
	limiter := l.getUserLimiter(userID)
	limiter.SetBurst(limit)
}

// SetUserPolicy sets a rate limit policy for a specific user
func (l *AdaptiveLimiter) SetUserPolicy(policy *UserRateLimitPolicy) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.userPolicies[policy.UserID] = policy
	
	// Update limiter if it exists
	if limiter, exists := l.userLimiters[policy.UserID]; exists {
		limiter.SetLimit(rate.Limit(policy.QPS))
		limiter.SetBurst(policy.Burst)
	}
}

// GetUserPolicy gets the rate limit policy for a specific user
func (l *AdaptiveLimiter) GetUserPolicy(userID string) *UserRateLimitPolicy {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.userPolicies[userID]
}

// applyDynamicAdjustment applies dynamic adjustments based on system load and user priority.
// 
// This method adjusts rate limits based on:
// 1. Current system load factor
// 2. User priority (higher priority users maintain higher limits during contention)
// 3. Historical usage patterns
//
// The adjustment algorithm ensures that even during high load, high-priority users
// maintain reasonable throughput while low-priority users are throttled more aggressively.
func (l *AdaptiveLimiter) applyDynamicAdjustment(userID string) {
	if !l.dynamicAdjustment {
		return
	}
	
	l.mu.RLock()
	limiter, exists := l.userLimiters[userID]
	policy, hasPolicyForUser := l.userPolicies[userID]
	loadFactor := l.loadFactor
	usage := l.userUsage[userID]
	l.mu.RUnlock()
	
	if !exists {
		return
	}
	
	// Get user priority (default to 1 if not specified)
	userPriority := 1
	if hasPolicyForUser && policy.Priority > 0 {
		userPriority = policy.Priority
	}
	
	// Calculate priority factor (ranges from 0.5 to 1.5)
	// Higher priority users get a boost, lower priority users get reduced limits
	priorityFactor := 0.5 + float64(userPriority)/10.0
	
	// Calculate usage factor (ranges from 0.7 to 1.2)
	// Users with lower historical usage get a slight boost
	usageFactor := 1.0
	if hasPolicyForUser && policy.MaxTokens > 0 {
		// Calculate percentage of max tokens used
		percentUsed := float64(usage) / float64(policy.MaxTokens)
		
		// Map to a factor: lower usage = higher factor
		usageFactor = 1.2 - (percentUsed * 0.5)
		
		// Clamp to reasonable range
		if usageFactor < 0.7 {
			usageFactor = 0.7
		} else if usageFactor > 1.2 {
			usageFactor = 1.2
		}
	}
	
	// Combined adjustment factor
	combinedFactor := loadFactor * priorityFactor * usageFactor
	
	// Adjust limits based on combined factor
	var adjustedLimit rate.Limit
	var adjustedBurst int
	
	if hasPolicyForUser {
		// Use policy-specific values
		adjustedLimit = rate.Limit(policy.QPS * combinedFactor)
		adjustedBurst = int(float64(policy.Burst) * combinedFactor)
	} else {
		// Use default values
		adjustedLimit = l.defaultUserLimit * rate.Limit(combinedFactor)
		adjustedBurst = int(float64(l.defaultUserBurst) * combinedFactor)
	}
	
	// Ensure minimum values
	if adjustedLimit < 0.1 {
		adjustedLimit = 0.1
	}
	if adjustedBurst < 1 {
		adjustedBurst = 1
	}
	
	// Apply maximum limits for very high priority users during extreme load
	if userPriority >= 9 && loadFactor < 0.3 {
		// Ensure critical users maintain at least 50% of their normal capacity
		// even during extreme load conditions
		if hasPolicyForUser {
			minLimit := rate.Limit(policy.QPS * 0.5)
			minBurst := int(float64(policy.Burst) * 0.5)
			
			if adjustedLimit < minLimit {
				adjustedLimit = minLimit
			}
			if adjustedBurst < minBurst {
				adjustedBurst = minBurst
			}
		}
	}
	
	// Update limiter
	limiter.SetLimit(adjustedLimit)
	limiter.SetBurst(adjustedBurst)
}

// SetLoadFactor sets the system load factor for dynamic adjustments
func (l *AdaptiveLimiter) SetLoadFactor(factor float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.loadFactor = factor
}

// EnableDynamicAdjustment enables or disables dynamic adjustment
func (l *AdaptiveLimiter) EnableDynamicAdjustment(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.dynamicAdjustment = enabled
}

// EnableFairness enables or disables fairness mechanisms
func (l *AdaptiveLimiter) EnableFairness(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.fairnessEnabled = enabled
}

// trackUsage tracks usage for a user
func (l *AdaptiveLimiter) trackUsage(userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.userUsage[userID]++
	
	// Check if we need to reset usage counters
	now := time.Now()
	policy, hasPolicyForUser := l.userPolicies[userID]
	
	if hasPolicyForUser && policy.ResetInterval > 0 {
		if now.Sub(l.lastResetTime) > policy.ResetInterval {
			// Reset all usage counters
			l.userUsage = make(map[string]int)
			l.lastResetTime = now
		}
	}
}

// checkUserTokens checks if a user has exceeded their token allocation
func (l *AdaptiveLimiter) checkUserTokens(userID string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	policy, hasPolicyForUser := l.userPolicies[userID]
	if !hasPolicyForUser || policy.MaxTokens <= 0 {
		// No token limit for this user
		return true
	}
	
	usage, hasUsage := l.userUsage[userID]
	if !hasUsage {
		// No usage recorded yet
		return true
	}
	
	// Check if user has exceeded their token allocation
	return usage < policy.MaxTokens
}

// GetUserUsage gets the current usage for a user
func (l *AdaptiveLimiter) GetUserUsage(userID string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.userUsage[userID]
}

// ResetUserUsage resets the usage counter for a user
func (l *AdaptiveLimiter) ResetUserUsage(userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	delete(l.userUsage, userID)
}

// ResetAllUserUsage resets all usage counters
func (l *AdaptiveLimiter) ResetAllUserUsage() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.userUsage = make(map[string]int)
	l.lastResetTime = time.Now()
}
