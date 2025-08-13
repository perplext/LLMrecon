package utils

import (
	"context"
	"sync"
	"time"
)

// RateLimiter provides rate limiting functionality for API calls
type RateLimiter struct {
	mutex       sync.Mutex
	requestsMap map[string]*providerRequests
}

// providerRequests tracks requests for a specific provider
type providerRequests struct {
	requests       []time.Time
	requestsPerMin int
	tokensPerMin   int
	tokensUsed     int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requestsMap: make(map[string]*providerRequests),
	}
}

// RegisterProvider registers a provider with rate limits
func (rl *RateLimiter) RegisterProvider(provider string, requestsPerMin, tokensPerMin int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	rl.requestsMap[provider] = &providerRequests{
		requests:       []time.Time{},
		requestsPerMin: requestsPerMin,
		tokensPerMin:   tokensPerMin,
		tokensUsed:     0,
	}
}

// Wait blocks until a request can be made according to rate limits
func (rl *RateLimiter) Wait(ctx context.Context, provider string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if rl.canProceed(provider) {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// RecordRequest records that a request was made
func (rl *RateLimiter) RecordRequest(provider string, tokenCount int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	if pr, ok := rl.requestsMap[provider]; ok {
		now := time.Now()
		pr.requests = append(pr.requests, now)
		pr.tokensUsed += tokenCount
		
		// Clean up old requests (older than 1 minute)
		var newRequests []time.Time
		for _, t := range pr.requests {
			if now.Sub(t) < time.Minute {
				newRequests = append(newRequests, t)
			}
		}
		pr.requests = newRequests
		
		// Reset token count if it's been more than a minute since the first request
		if len(pr.requests) > 0 && now.Sub(pr.requests[0]) >= time.Minute {
			pr.tokensUsed = tokenCount
		}
	}
}

// canProceed checks if a request can proceed based on rate limits
func (rl *RateLimiter) canProceed(provider string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	pr, ok := rl.requestsMap[provider]
	if !ok {
		// If provider is not registered, allow the request
		return true
	}
	
	now := time.Now()
	
	// Count requests in the last minute
	var requestsInLastMinute int
	for _, t := range pr.requests {
		if now.Sub(t) < time.Minute {
			requestsInLastMinute++
		}
	}
	
	// Check if we're under the rate limits
	return requestsInLastMinute < pr.requestsPerMin && pr.tokensUsed < pr.tokensPerMin
}

// GetUsage returns the current usage statistics
func (rl *RateLimiter) GetUsage(provider string) map[string]interface{} {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	pr, ok := rl.requestsMap[provider]
	if !ok {
		return map[string]interface{}{
			"requests_per_min": 0,
			"tokens_per_min":   0,
			"requests_used":    0,
			"tokens_used":      0,
		}
	}
	
	now := time.Now()
	var requestsInLastMinute int
	for _, t := range pr.requests {
		if now.Sub(t) < time.Minute {
			requestsInLastMinute++
		}
	}
	
	return map[string]interface{}{
		"requests_per_min": pr.requestsPerMin,
		"tokens_per_min":   pr.tokensPerMin,
		"requests_used":    requestsInLastMinute,
		"tokens_used":      pr.tokensUsed,
	}
}
