// Package middleware provides middleware components for the Multi-Provider LLM Integration Framework.
package middleware

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// RetryMiddleware provides retry functionality with exponential backoff
type RetryMiddleware struct {
	// config is the retry configuration
	config *core.RetryConfig
}

// NewRetryMiddleware creates a new retry middleware
func NewRetryMiddleware(config *core.RetryConfig) *RetryMiddleware {
	if config == nil {
		config = &core.RetryConfig{
			MaxRetries:          3,
			InitialBackoff:      1 * time.Second,
			MaxBackoff:          60 * time.Second,
			BackoffMultiplier:   2.0,
			RetryableStatusCodes: []int{
				http.StatusTooManyRequests,
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
			},
		}
	}

	return &RetryMiddleware{
		config: config,
	}
}

// Execute executes a function with retries
func (m *RetryMiddleware) Execute(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var result interface{}
	var err error
	var attempt int

	// Initialize random number generator for jitter
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Try the operation until max retries or success
	for attempt = 0; attempt <= m.config.MaxRetries; attempt++ {
		// Execute the function
		result, err = fn(ctx)

		// If no error or not retryable, return the result
		if err == nil || !m.isRetryableError(err) {
			return result, err
		}

		// Check if we've reached max retries
		if attempt == m.config.MaxRetries {
			break
		}

		// Calculate backoff duration with jitter
		backoff := m.calculateBackoff(attempt, rng)

		// Wait for backoff duration or context cancellation
		select {
		case <-time.After(backoff):
			// Continue to next retry
		case <-ctx.Done():
			// Context cancelled, return error
			return nil, ctx.Err()
		}
	}

	// Return the last error
	return result, fmt.Errorf("max retries reached: %w", err)
}

// isRetryableError checks if an error is retryable
func (m *RetryMiddleware) isRetryableError(err error) bool {
	// Check if it's a provider error with a status code
	if providerErr, ok := err.(*core.ProviderError); ok {
		// Check if the status code is in the retryable status codes
		for _, code := range m.config.RetryableStatusCodes {
			if providerErr.StatusCode == code {
				return true
			}
		}

		// Check if it's a rate limit error
		if providerErr.StatusCode == http.StatusTooManyRequests {
			return true
		}

		// Check if it's a server error
		if providerErr.StatusCode >= 500 && providerErr.StatusCode < 600 {
			return true
		}
	}

	// Check if it's a network error or timeout
	// This is a simplified check and may need to be expanded
	return false
}

// calculateBackoff calculates the backoff duration with jitter
func (m *RetryMiddleware) calculateBackoff(attempt int, rng *rand.Rand) time.Duration {
	// Calculate base backoff with exponential increase
	backoff := float64(m.config.InitialBackoff) * math.Pow(m.config.BackoffMultiplier, float64(attempt))

	// Apply jitter (random variation) to avoid thundering herd problem
	jitter := 0.5 + rng.Float64()*0.5 // 50-100% of calculated backoff
	backoff = backoff * jitter

	// Cap at max backoff
	if backoff > float64(m.config.MaxBackoff) {
		backoff = float64(m.config.MaxBackoff)
	}

	return time.Duration(backoff)
}

// UpdateConfig updates the retry configuration
func (m *RetryMiddleware) UpdateConfig(config *core.RetryConfig) {
	if config == nil {
		return
	}

	// Create a copy to avoid race conditions
	configCopy := *config

	// Update the configuration
	m.config = &configCopy
}

// GetConfig returns the retry configuration
func (m *RetryMiddleware) GetConfig() *core.RetryConfig {
	// Create a copy to avoid race conditions
	configCopy := *m.config
	return &configCopy
}
