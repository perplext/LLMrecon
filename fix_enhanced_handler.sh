#!/bin/bash

echo "Fixing enhanced_handler.go..."

cat > src/bundle/errors/enhanced_handler.go << 'EOF'
// Package errors provides error handling functionality for bundle operations
package errors

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sync"
	"time"
)

// DefaultErrorHandler provides basic error handling functionality
type DefaultErrorHandler struct {
	logger      io.Writer
	auditLogger *AuditLogger
	categorizer ErrorCategorizer
}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler(logger io.Writer, auditLogger *AuditLogger) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		logger:      logger,
		auditLogger: auditLogger,
		categorizer: NewErrorCategorizer(logger),
	}
}

// HandleError handles an error using the default strategy
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error) *BundleError {
	if err == nil {
		return nil
	}

	// Categorize the error
	category, severity, recoverability := h.categorizer.CategorizeError(err)
	
	// Create a BundleError
	bundleError := NewBundleError(
		ImportErrorCode,
		category,
		severity,
		recoverability,
		err.Error(),
	).WithCause(err)

	// Log the error
	if h.logger != nil {
		fmt.Fprintf(h.logger, "Error handled: %s\n", bundleError.Error())
	}

	return bundleError
}

// EnhancedErrorHandler extends the DefaultErrorHandler with advanced error handling capabilities
type EnhancedErrorHandler struct {
	*DefaultErrorHandler
	
	// circuitBreaker tracks consecutive errors to implement circuit breaker pattern
	circuitBreaker *CircuitBreaker
	
	// retryPolicy defines retry behavior
	retryPolicy *RetryPolicy
	
	// errorCounters track error statistics
	errorCounters map[ErrorCategory]int64
	countersMutex sync.RWMutex
	
	// rateLimiter controls error handling rate
	rateLimiter *TokenBucketRateLimiter
}

// CircuitBreaker implements the circuit breaker pattern for error handling
type CircuitBreaker struct {
	failureThreshold int
	resetTimeout     time.Duration
	consecutiveFailures int
	lastFailureTime  time.Time
	state           CircuitBreakerState
	mutex           sync.RWMutex
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// RetryPolicy defines retry behavior for error handling
type RetryPolicy struct {
	maxRetries    int
	baseDelay     time.Duration
	maxDelay      time.Duration
	backoffFactor float64
}

// TokenBucketRateLimiter implements a token bucket rate limiter
type TokenBucketRateLimiter struct {
	capacity       int64
	tokens         int64
	refillRate     int64
	lastRefillTime time.Time
	mutex          sync.Mutex
}

// NewEnhancedErrorHandler creates a new enhanced error handler
func NewEnhancedErrorHandler(logger io.Writer, auditLogger *AuditLogger) *EnhancedErrorHandler {
	defaultHandler := NewDefaultErrorHandler(logger, auditLogger)
	
	return &EnhancedErrorHandler{
		DefaultErrorHandler: defaultHandler,
		circuitBreaker: &CircuitBreaker{
			failureThreshold: 5,
			resetTimeout:     time.Minute * 5,
			state:           CircuitBreakerClosed,
		},
		retryPolicy: &RetryPolicy{
			maxRetries:    3,
			baseDelay:     time.Millisecond * 100,
			maxDelay:      time.Second * 10,
			backoffFactor: 2.0,
		},
		errorCounters: make(map[ErrorCategory]int64),
		rateLimiter: &TokenBucketRateLimiter{
			capacity:   100,
			tokens:     100,
			refillRate: 10,
			lastRefillTime: time.Now(),
		},
	}
}

// HandleErrorWithRetry handles an error with retry logic
func (h *EnhancedErrorHandler) HandleErrorWithRetry(ctx context.Context, err error, operation func() error) *BundleError {
	if err == nil {
		return nil
	}

	// Check circuit breaker
	if !h.circuitBreaker.CanExecute() {
		return NewBundleError(
			ImportErrorCode,
			NetworkError,
			HighSeverity,
			NonRecoverableError,
			"circuit breaker is open",
		)
	}

	// Apply rate limiting
	if !h.rateLimiter.Allow() {
		return NewBundleError(
			ImportErrorCode,
			NetworkError,
			MediumSeverity,
			RecoverableError,
			"rate limit exceeded",
		)
	}

	bundleError := h.HandleError(ctx, err)
	
	// Update error counters
	h.updateErrorCounters(bundleError.Category)
	
	// Attempt retry if applicable
	if bundleError.Recoverability == RecoverableError && operation != nil {
		bundleError = h.retryOperation(ctx, bundleError, operation)
	}
	
	// Update circuit breaker
	h.circuitBreaker.RecordResult(bundleError == nil)
	
	return bundleError
}

// Allow checks if the rate limiter allows an operation
func (r *TokenBucketRateLimiter) Allow() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Refill tokens
	now := time.Now()
	timeSinceLastRefill := now.Sub(r.lastRefillTime)
	tokensToAdd := int64(timeSinceLastRefill.Seconds()) * r.refillRate
	
	if tokensToAdd > 0 {
		r.tokens = min(r.capacity, r.tokens+tokensToAdd)
		r.lastRefillTime = now
	}
	
	// Check if we have tokens
	if r.tokens > 0 {
		r.tokens--
		return true
	}
	
	return false
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// Check if reset timeout has passed
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitBreakerHalfOpen
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// RecordResult records the result of an operation
func (cb *CircuitBreaker) RecordResult(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if success {
		cb.consecutiveFailures = 0
		if cb.state == CircuitBreakerHalfOpen {
			cb.state = CircuitBreakerClosed
		}
	} else {
		cb.consecutiveFailures++
		cb.lastFailureTime = time.Now()
		
		if cb.consecutiveFailures >= cb.failureThreshold {
			cb.state = CircuitBreakerOpen
		}
	}
}

// updateErrorCounters updates error statistics
func (h *EnhancedErrorHandler) updateErrorCounters(category ErrorCategory) {
	h.countersMutex.Lock()
	defer h.countersMutex.Unlock()
	
	h.errorCounters[category]++
}

// retryOperation attempts to retry an operation with exponential backoff
func (h *EnhancedErrorHandler) retryOperation(ctx context.Context, originalError *BundleError, operation func() error) *BundleError {
	for attempt := 0; attempt < h.retryPolicy.maxRetries; attempt++ {
		// Calculate delay with jitter
		delay := h.calculateRetryDelay(attempt)
		
		// Wait for delay or context cancellation
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return originalError
		}
		
		// Retry the operation
		if err := operation(); err == nil {
			return nil // Success
		}
	}
	
	return originalError
}

// calculateRetryDelay calculates the delay for a retry attempt
func (h *EnhancedErrorHandler) calculateRetryDelay(attempt int) time.Duration {
	delay := float64(h.retryPolicy.baseDelay) * math.Pow(h.retryPolicy.backoffFactor, float64(attempt))
	
	// Add jitter
	jitter := rand.Float64() * 0.1 * delay
	delay += jitter
	
	// Cap at max delay
	if delay > float64(h.retryPolicy.maxDelay) {
		delay = float64(h.retryPolicy.maxDelay)
	}
	
	return time.Duration(delay)
}

// GetErrorStatistics returns error statistics
func (h *EnhancedErrorHandler) GetErrorStatistics() map[ErrorCategory]int64 {
	h.countersMutex.RLock()
	defer h.countersMutex.RUnlock()
	
	// Return a copy to prevent external modification
	stats := make(map[ErrorCategory]int64)
	for category, count := range h.errorCounters {
		stats[category] = count
	}
	
	return stats
}

// Helper function for min
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
EOF

echo "Done fixing enhanced_handler.go!"