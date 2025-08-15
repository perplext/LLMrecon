// Package middleware provides middleware components for the Multi-Provider LLM Integration Framework.
package middleware

import (
	"time"
	"context"
	"errors"
	"sync"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	// CircuitBreakerStateClosed indicates that the circuit is closed and requests are allowed
	CircuitBreakerStateClosed CircuitBreakerState = iota
	// CircuitBreakerStateOpen indicates that the circuit is open and requests are not allowed
	CircuitBreakerStateOpen
	// CircuitBreakerStateHalfOpen indicates that the circuit is half-open and a limited number of requests are allowed
	CircuitBreakerStateHalfOpen
)

// CircuitBreakerConfig represents the configuration for the circuit breaker
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures that will open the circuit
	FailureThreshold int
	// ResetTimeout is the time to wait before trying again after the circuit is opened
	ResetTimeout time.Duration
	// HalfOpenSuccessThreshold is the number of consecutive successes that will close the circuit
	HalfOpenSuccessThreshold int

// CircuitBreakerMiddleware provides circuit breaking functionality
type CircuitBreakerMiddleware struct {
	config CircuitBreakerConfig
	state  CircuitBreakerState
	// consecutiveFailures is the number of consecutive failures
	consecutiveFailures int
	// consecutiveSuccesses is the number of consecutive successes in half-open state
	consecutiveSuccesses int
	// lastStateChange is the time of the last state change
	lastStateChange time.Time
	mutex           sync.RWMutex
}

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware
func NewCircuitBreakerMiddleware(config CircuitBreakerConfig) *CircuitBreakerMiddleware {
	// Set default values if not specified
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 60 * time.Second
	}
	if config.HalfOpenSuccessThreshold <= 0 {
		config.HalfOpenSuccessThreshold = 2
	}

	return &CircuitBreakerMiddleware{
		config:               config,
		state:                CircuitBreakerStateClosed,
		consecutiveFailures:  0,
		consecutiveSuccesses: 0,
		lastStateChange:      time.Now(),
		mutex:                sync.RWMutex{},
	}

// Execute executes a function with circuit breaking
func (cb *CircuitBreakerMiddleware) Execute(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// Check if the circuit is open
	if !cb.allowRequest() {
		return nil, errors.New("circuit breaker is open")
	}

	// Execute the function
	result, err := fn(ctx)
	// Update the circuit breaker state based on the result
	cb.updateState(err == nil)

	return result, err

// allowRequest checks if a request is allowed based on the circuit breaker state
func (cb *CircuitBreakerMiddleware) allowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitBreakerStateClosed:
		// Always allow requests when the circuit is closed
		return true
	case CircuitBreakerStateOpen:
		// Check if enough time has passed to try again
		if time.Since(cb.lastStateChange) > cb.config.ResetTimeout {
			// Transition to half-open state
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = CircuitBreakerStateHalfOpen
			cb.lastStateChange = time.Now()
			cb.consecutiveSuccesses = 0
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case CircuitBreakerStateHalfOpen:
		// Allow a limited number of requests in half-open state
		return true
	default:
		return false
	}

// updateState updates the circuit breaker state based on the result of a request
func (cb *CircuitBreakerMiddleware) updateState(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if success {
		// Reset consecutive failures on success
		cb.consecutiveFailures = 0

		// If in half-open state, increment consecutive successes
		if cb.state == CircuitBreakerStateHalfOpen {
			cb.consecutiveSuccesses++

			// If enough consecutive successes, close the circuit
			if cb.consecutiveSuccesses >= cb.config.HalfOpenSuccessThreshold {
				cb.state = CircuitBreakerStateClosed
				cb.lastStateChange = time.Now()
			}
		}
	} else {
		// Increment consecutive failures on failure
		cb.consecutiveFailures++

		// Reset consecutive successes on failure
		cb.consecutiveSuccesses = 0

		// If enough consecutive failures, open the circuit
		if (cb.state == CircuitBreakerStateClosed && cb.consecutiveFailures >= cb.config.FailureThreshold) ||
			cb.state == CircuitBreakerStateHalfOpen {
			cb.state = CircuitBreakerStateOpen
			cb.lastStateChange = time.Now()
		}
	}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreakerMiddleware) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreakerMiddleware) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = CircuitBreakerStateClosed
	cb.consecutiveFailures = 0
	cb.consecutiveSuccesses = 0
	cb.lastStateChange = time.Now()

// UpdateConfig updates the circuit breaker configuration
func (cb *CircuitBreakerMiddleware) UpdateConfig(config CircuitBreakerConfig) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.config = config

// NewCircuitBreaker is an alias for NewCircuitBreakerMiddleware for backward compatibility
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreakerMiddleware {
	return NewCircuitBreakerMiddleware(config)
}
}
}
}
}
}
}
