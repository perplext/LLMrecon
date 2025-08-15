// Package errors provides error handling functionality for bundle operations
package errors

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// DefaultErrorHandler is the default error handler implementation
type DefaultErrorHandler struct {
	// Logger is the logger for error events
	Logger interface{}
	// Metrics is the metrics collector
	Metrics interface{}
	// RecoveryStrategy is the strategy for recovering from errors
	RecoveryStrategy RecoveryStrategy
}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{}
}

// HandleError handles an error
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	
	// Log the error
	fmt.Printf("Error: %v\n", err)
	
	// Apply recovery strategy if available
	if h.RecoveryStrategy != nil {
		// Convert error to BundleError if needed
		var bundleErr *BundleError
		if be, ok := err.(*BundleError); ok {
			bundleErr = be
		} else {
			bundleErr = &BundleError{
				Code:    "UNKNOWN_ERROR",
				Message: err.Error(),
				Cause:   err,
			}
		}
		
		recovered, recErr := h.RecoveryStrategy.Recover(ctx, bundleErr)
		if recovered {
			return recErr
		}
	}
	
	return err
}

// SetRecoveryStrategy sets the recovery strategy
func (h *DefaultErrorHandler) SetRecoveryStrategy(strategy RecoveryStrategy) {
	h.RecoveryStrategy = strategy
}

// EnhancedErrorHandler provides enhanced error handling capabilities
type EnhancedErrorHandler struct {
	// ErrorCategorizer categorizes errors
	ErrorCategorizer ErrorCategorizer
	// RecoveryManager manages recovery strategies
	RecoveryManager *RecoveryManager
	// CircuitBreaker provides circuit breaking functionality
	CircuitBreaker *CircuitBreaker
	// RetryPolicy defines retry behavior
	RetryPolicy *RetryPolicy
	// RateLimiter limits error handling rate
	RateLimiter *TokenBucketRateLimiter
	// Metrics tracks error metrics
	Metrics map[string]int
	// mutex protects concurrent access
	mutex sync.Mutex
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	// State is the current state
	State string
	// FailureCount tracks failures
	FailureCount int
	// SuccessCount tracks successes
	SuccessCount int
	// LastFailureTime is the time of last failure
	LastFailureTime time.Time
	// Threshold is the failure threshold
	Threshold int
	// Timeout is the timeout duration
	Timeout time.Duration
	// mutex protects concurrent access
	mutex sync.Mutex
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int
	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the delay multiplier
	Multiplier float64
	// Jitter adds randomness to delays
	Jitter bool
}

// TokenBucketRateLimiter implements token bucket rate limiting
type TokenBucketRateLimiter struct {
	// Capacity is the bucket capacity
	Capacity int
	// RefillRate is the token refill rate
	RefillRate int
	// Tokens is the current token count
	Tokens int
	// LastRefill is the last refill time
	LastRefill time.Time
	// mutex protects concurrent access
	mutex sync.Mutex
}

// NewEnhancedErrorHandler creates a new enhanced error handler
func NewEnhancedErrorHandler() *EnhancedErrorHandler {
	return &EnhancedErrorHandler{
		ErrorCategorizer: NewErrorCategorizer(nil),
		RecoveryManager:  NewRecoveryManager(os.Stderr, nil),
		CircuitBreaker:   NewCircuitBreaker(5, 30*time.Second),
		RetryPolicy:      NewRetryPolicy(3, 1*time.Second, 30*time.Second),
		RateLimiter:      NewTokenBucketRateLimiter(100, 10),
		Metrics:          make(map[string]int),
	}
}

// HandleError handles an error with enhanced capabilities
func (h *EnhancedErrorHandler) HandleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Categorize the error
	category, severity, recoverability := h.ErrorCategorizer.CategorizeError(err)
	
	// Update metrics
	h.Metrics[string(category)]++
	h.Metrics[string(severity)]++
	
	// Check circuit breaker
	if h.CircuitBreaker != nil && !h.CircuitBreaker.CanProceed() {
		return fmt.Errorf("circuit breaker open: %w", err)
	}
	
	// Apply rate limiting
	if h.RateLimiter != nil && !h.RateLimiter.Allow() {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}
	
	// Handle based on recoverability
	if recoverability == RecoverableError {
		// Apply retry policy
		if h.RetryPolicy != nil {
			return h.RetryWithPolicy(ctx, err)
		}
	}
	
	// Apply recovery strategy
	if h.RecoveryManager != nil {
		// Convert error to BundleError if needed
		var bundleErr *BundleError
		if be, ok := err.(*BundleError); ok {
			bundleErr = be
		} else {
			bundleErr = &BundleError{
				Code:    "UNKNOWN_ERROR",
				Message: err.Error(),
				Cause:   err,
			}
		}
		
		recovered, recErr := h.RecoveryManager.AttemptRecovery(ctx, bundleErr)
		if recovered {
			return recErr
		}
	}
	
	// Record failure in circuit breaker
	if h.CircuitBreaker != nil {
		h.CircuitBreaker.RecordFailure()
	}
	
	return err
}

// RetryWithPolicy retries an operation with the configured policy
func (h *EnhancedErrorHandler) RetryWithPolicy(ctx context.Context, err error) error {
	if h.RetryPolicy == nil {
		return err
	}
	
	delay := h.RetryPolicy.InitialDelay
	
	for i := 0; i < h.RetryPolicy.MaxRetries; i++ {
		// Wait before retry
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		
		// Update delay for next retry
		delay = time.Duration(float64(delay) * h.RetryPolicy.Multiplier)
		if delay > h.RetryPolicy.MaxDelay {
			delay = h.RetryPolicy.MaxDelay
		}
		
		// Add jitter if configured
		if h.RetryPolicy.Jitter {
			jitter := time.Duration(randFloat64() * float64(delay) * 0.1)
			delay += jitter
		}
	}
	
	return err
}

// CanProceed checks if the circuit breaker allows proceeding
func (cb *CircuitBreaker) CanProceed() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	switch cb.State {
	case "open":
		// Check if timeout has passed
		if time.Since(cb.LastFailureTime) > cb.Timeout {
			cb.State = "half-open"
			return true
		}
		return false
	case "half-open":
		return true
	default: // closed
		return true
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.SuccessCount++
	
	if cb.State == "half-open" {
		cb.State = "closed"
		cb.FailureCount = 0
		cb.SuccessCount = 0
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.FailureCount++
	cb.LastFailureTime = time.Now()
	
	if cb.FailureCount >= cb.Threshold {
		cb.State = "open"
	}
	
	if cb.State == "half-open" {
		cb.State = "open"
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		State:     "closed",
		Threshold: threshold,
		Timeout:   timeout,
	}
}

// NewRetryPolicy creates a new retry policy
func NewRetryPolicy(maxRetries int, initialDelay, maxDelay time.Duration) *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:   maxRetries,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter
func NewTokenBucketRateLimiter(capacity, refillRate int) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		Capacity:   capacity,
		RefillRate: refillRate,
		Tokens:     capacity,
		LastRefill: time.Now(),
	}
}

// Allow checks if an operation is allowed
func (rl *TokenBucketRateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(rl.LastRefill)
	tokensToAdd := int(elapsed.Seconds()) * rl.RefillRate
	
	if tokensToAdd > 0 {
		rl.Tokens = min(rl.Tokens+tokensToAdd, rl.Capacity)
		rl.LastRefill = now
	}
	
	// Check if token available
	if rl.Tokens > 0 {
		rl.Tokens--
		return true
	}
	
	return false
}

// GetMetrics returns error handling metrics
func (h *EnhancedErrorHandler) GetMetrics() map[string]int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	metrics := make(map[string]int)
	for k, v := range h.Metrics {
		metrics[k] = v
	}
	
	return metrics
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// randFloat64 generates a random float64 between 0 and 1
func randFloat64() float64 {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return float64(binary.BigEndian.Uint64(bytes)) / (1 << 64)
}

// WrapWithContext wraps an error with context information
func WrapWithContext(err error, ctx context.Context, operation string) error {
	if err == nil {
		return nil
	}
	
	contextInfo := make(map[string]string)
	
	// Extract context values if available
	if reqID := ctx.Value("request_id"); reqID != nil {
		contextInfo["request_id"] = fmt.Sprintf("%v", reqID)
	}
	
	if userID := ctx.Value("user_id"); userID != nil {
		contextInfo["user_id"] = fmt.Sprintf("%v", userID)
	}
	
	// Build context string
	var contextParts []string
	for k, v := range contextInfo {
		contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
	}
	
	contextStr := ""
	if len(contextParts) > 0 {
		contextStr = fmt.Sprintf(" [%s]", strings.Join(contextParts, ", "))
	}
	
	return fmt.Errorf("%s%s: %w", operation, contextStr, err)
}

// IsTemporaryError checks if an error is temporary
func IsTemporaryError(err error) bool {
	type temporary interface {
		Temporary() bool
	}
	
	if te, ok := err.(temporary); ok {
		return te.Temporary()
	}
	
	// Check for specific error types
	errStr := err.Error()
	temporaryPatterns := []string{
		"temporary",
		"timeout",
		"connection reset",
		"connection refused",
		"unavailable",
	}
	
	for _, pattern := range temporaryPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}
	
	return false
}
