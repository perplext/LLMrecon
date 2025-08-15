// Package middleware provides middleware components for the Multi-Provider LLM Integration Framework.
package middleware

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality for API requests
type RateLimiter struct {
	// requestLimiter limits the number of requests per minute
	requestLimiter *rate.Limiter
	// tokenLimiter limits the number of tokens per minute
	tokenLimiter *rate.Limiter
	// concurrencyLimiter limits the number of concurrent requests
	concurrencyLimiter chan struct{}
	// mutex is a mutex for concurrent access
	mutex sync.RWMutex
	// enabled indicates whether rate limiting is enabled
	enabled bool

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60 // Default to 1 request per second
	}
	if tokensPerMinute <= 0 {
		tokensPerMinute = 100000 // Default to ~1.6K tokens per second
	}
	if maxConcurrentRequests <= 0 {
		maxConcurrentRequests = 10 // Default to 10 concurrent requests
	}
	if burstSize <= 0 {
		burstSize = requestsPerMinute / 10 // Default to 10% of requests per minute
		if burstSize < 1 {
			burstSize = 1
		}
	}

	return &RateLimiter{
		requestLimiter:     rate.NewLimiter(rate.Limit(float64(requestsPerMinute)/60.0), burstSize),
		tokenLimiter:       rate.NewLimiter(rate.Limit(float64(tokensPerMinute)/60.0), tokensPerMinute/10),
		concurrencyLimiter: make(chan struct{}, maxConcurrentRequests),
		enabled:            true,
	}

// Wait waits for rate limiting to allow a request
func (l *RateLimiter) Wait(ctx context.Context) error {
	l.mutex.RLock()
	enabled := l.enabled
	l.mutex.RUnlock()

	if !enabled {
		return nil
	}

	// Wait for request limiter
	if err := l.requestLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("request rate limit exceeded: %w", err)
	}

	// Acquire concurrency limiter
	select {
	case l.concurrencyLimiter <- struct{}{}:
		// Acquired
	case <-ctx.Done():
		return fmt.Errorf("concurrency limit wait cancelled: %w", ctx.Err())
	}

	return nil

// Release releases the concurrency limiter
func (l *RateLimiter) Release() {
	l.mutex.RLock()
	enabled := l.enabled
	l.mutex.RUnlock()

	if !enabled {
		return
	}

	select {
	case <-l.concurrencyLimiter:
		// Released
	default:
		// Nothing to release
	}

// WaitForTokens waits for rate limiting to allow tokens
func (l *RateLimiter) WaitForTokens(ctx context.Context, tokens int) error {
	l.mutex.RLock()
	enabled := l.enabled
	l.mutex.RUnlock()

	if !enabled {
		return nil
	}

	// Wait for token limiter
	if err := l.tokenLimiter.WaitN(ctx, tokens); err != nil {
		return fmt.Errorf("token rate limit exceeded: %w", err)
	}

	return nil

// Enable enables rate limiting
func (l *RateLimiter) Enable() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.enabled = true

// Disable disables rate limiting
func (l *RateLimiter) Disable() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.enabled = false

// IsEnabled returns whether rate limiting is enabled
func (l *RateLimiter) IsEnabled() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.enabled

// UpdateLimits updates the rate limits
func (l *RateLimiter) UpdateLimits(requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if requestsPerMinute > 0 {
		if burstSize <= 0 {
			burstSize = requestsPerMinute / 10 // Default to 10% of requests per minute
			if burstSize < 1 {
				burstSize = 1
			}
		}
		l.requestLimiter = rate.NewLimiter(rate.Limit(float64(requestsPerMinute)/60.0), burstSize)
	}

	if tokensPerMinute > 0 {
		l.tokenLimiter = rate.NewLimiter(rate.Limit(float64(tokensPerMinute)/60.0), tokensPerMinute/10)
	}

	if maxConcurrentRequests > 0 {
		// Create a new concurrency limiter with the new size
		newConcurrencyLimiter := make(chan struct{}, maxConcurrentRequests)

		// Transfer existing tokens to the new limiter
		for {
			select {
			case <-l.concurrencyLimiter:
				// Try to add to the new limiter
				select {
				case newConcurrencyLimiter <- struct{}{}:
					// Added
				default:
					// New limiter is full, discard
				}
			default:
				// No more tokens to transfer
				l.concurrencyLimiter = newConcurrencyLimiter
				return
			}
		}
	}

// GetLimits returns the current rate limits
func (l *RateLimiter) GetLimits() (requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize int) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	requestsPerMinute = int(l.requestLimiter.Limit() * 60.0)
	tokensPerMinute = int(l.tokenLimiter.Limit() * 60.0)
	maxConcurrentRequests = cap(l.concurrencyLimiter)
	burstSize = l.requestLimiter.Burst()

	return

// RateLimiterMiddleware is middleware that applies rate limiting to a provider
type RateLimiterMiddleware struct {
	// rateLimiter is the rate limiter
	rateLimiter *RateLimiter

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rateLimiter: NewRateLimiter(requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize),
	}

// Execute executes a function with rate limiting
func (m *RateLimiterMiddleware) Execute(ctx context.Context, tokens int, fn func(ctx context.Context) error) error {
	// Wait for rate limiting
	if err := m.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	defer m.rateLimiter.Release()

	// Wait for token rate limiting
	if tokens > 0 {
		if err := m.rateLimiter.WaitForTokens(ctx, tokens); err != nil {
			return err
		}
	}

	// Execute the function
	return fn(ctx)

// GetRateLimiter returns the rate limiter
func (m *RateLimiterMiddleware) GetRateLimiter() *RateLimiter {
	return m.rateLimiter

// ProviderRateLimiter manages rate limiters for multiple providers
type ProviderRateLimiter struct {
	// rateLimiters is a map of provider types to rate limiters
	rateLimiters map[string]*RateLimiter
	// mutex is a mutex for concurrent access to rateLimiters
	mutex sync.RWMutex

// NewProviderRateLimiter creates a new provider rate limiter
func NewProviderRateLimiter() *ProviderRateLimiter {
	return &ProviderRateLimiter{
		rateLimiters: make(map[string]*RateLimiter),
	}

// GetRateLimiter gets or creates a rate limiter for a provider
func (p *ProviderRateLimiter) GetRateLimiter(providerKey string, requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize int) *RateLimiter {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	rateLimiter, ok := p.rateLimiters[providerKey]
	if !ok {
		rateLimiter = NewRateLimiter(requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize)
		p.rateLimiters[providerKey] = rateLimiter
	}

	return rateLimiter

// UpdateRateLimiter updates a rate limiter for a provider
func (p *ProviderRateLimiter) UpdateRateLimiter(providerKey string, requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	rateLimiter, ok := p.rateLimiters[providerKey]
	if !ok {
		rateLimiter = NewRateLimiter(requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize)
		p.rateLimiters[providerKey] = rateLimiter
	} else {
		rateLimiter.UpdateLimits(requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize)
	}

// RemoveRateLimiter removes a rate limiter for a provider
func (p *ProviderRateLimiter) RemoveRateLimiter(providerKey string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.rateLimiters, providerKey)

// GetAllRateLimiters returns all rate limiters
func (p *ProviderRateLimiter) GetAllRateLimiters() map[string]*RateLimiter {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Return a copy of the map to prevent concurrent modification
	rateLimitersCopy := make(map[string]*RateLimiter)
	for key, rateLimiter := range p.rateLimiters {
		rateLimitersCopy[key] = rateLimiter
	}

	return rateLimitersCopy

// EnableRateLimiter enables a rate limiter for a provider
func (p *ProviderRateLimiter) EnableRateLimiter(providerKey string) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if rateLimiter, ok := p.rateLimiters[providerKey]; ok {
		rateLimiter.Enable()
	}

// DisableRateLimiter disables a rate limiter for a provider
func (p *ProviderRateLimiter) DisableRateLimiter(providerKey string) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if rateLimiter, ok := p.rateLimiters[providerKey]; ok {
		rateLimiter.Disable()
	}

// EnableAllRateLimiters enables all rate limiters
func (p *ProviderRateLimiter) EnableAllRateLimiters() {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, rateLimiter := range p.rateLimiters {
		rateLimiter.Enable()
	}

// DisableAllRateLimiters disables all rate limiters
func (p *ProviderRateLimiter) DisableAllRateLimiters() {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, rateLimiter := range p.rateLimiters {
		rateLimiter.Disable()
	}
