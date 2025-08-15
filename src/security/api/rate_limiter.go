// Package api provides API protection mechanisms for the LLMrecon tool.
package api


import (
	"time"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiterConfig represents the configuration for a rate limiter
type RateLimiterConfig struct {
	// RequestsPerMinute is the maximum number of requests per minute
	RequestsPerMinute int
	// BurstSize is the maximum burst size
	BurstSize int
	// IPHeaderName is the name of the header containing the client IP
	IPHeaderName string
	// TrustedProxies is a list of trusted proxy IPs
	TrustedProxies []string
	// ExemptIPs is a list of IPs exempt from rate limiting
	ExemptIPs []string
	// ExemptPaths is a list of paths exempt from rate limiting
	ExemptPaths []string

// DefaultRateLimiterConfig returns the default rate limiter configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		IPHeaderName:      "X-Forwarded-For",
		TrustedProxies:    []string{"127.0.0.1", "::1"},
		ExemptPaths:       []string{"/health", "/metrics"},
	}

// RateLimiter implements rate limiting for API requests
type RateLimiter struct {
	config     *RateLimiterConfig
	limiters   map[string]*rate.Limiter
	mu         sync.RWMutex
	exemptIPs  map[string]bool
	exemptCIDR []*net.IPNet
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	// Create exempt IP map
	exemptIPs := make(map[string]bool)
	var exemptCIDR []*net.IPNet
	for _, ip := range config.ExemptIPs {
		if _, ipnet, err := net.ParseCIDR(ip); err == nil {
			exemptCIDR = append(exemptCIDR, ipnet)
		} else {
			exemptIPs[ip] = true
		}
	}

	return &RateLimiter{
		config:     config,
		limiters:   make(map[string]*rate.Limiter),
		exemptIPs:  exemptIPs,
		exemptCIDR: exemptCIDR,
	}

// GetLimiter gets a rate limiter for a client
func (rl *RateLimiter) GetLimiter(clientIP string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[clientIP]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		defer rl.mu.Unlock()

		// Check again in case another goroutine created it
		limiter, exists = rl.limiters[clientIP]
		if !exists {
			// Create a new limiter with the configured rate
			limiter = rate.NewLimiter(rate.Limit(rl.config.RequestsPerMinute)/60, rl.config.BurstSize)
			rl.limiters[clientIP] = limiter
		}
	}

	return limiter

// CleanupLimiters removes expired limiters
func (rl *RateLimiter) CleanupLimiters(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// This is a simple implementation that removes all limiters
	// In a production environment, you would track the last access time
	rl.limiters = make(map[string]*rate.Limiter)

// IsExempt checks if a client is exempt from rate limiting
func (rl *RateLimiter) IsExempt(clientIP string, path string) bool {
	// Check if the path is exempt
	for _, exemptPath := range rl.config.ExemptPaths {
		if path == exemptPath {
			return true
		}
	}

	// Check if the IP is exempt
	if rl.exemptIPs[clientIP] {
		return true
	}

	// Check if the IP is in an exempt CIDR
	ip := net.ParseIP(clientIP)
	if ip != nil {
		for _, ipnet := range rl.exemptCIDR {
			if ipnet.Contains(ip) {
				return true
			}
		}
	}

	return false

// GetClientIP gets the client IP from a request
func (rl *RateLimiter) GetClientIP(r *http.Request) string {
	// Check for IP in header (e.g., X-Forwarded-For)
	if rl.config.IPHeaderName != "" {
		ip := r.Header.Get(rl.config.IPHeaderName)
		if ip != "" {
			// The header might contain multiple IPs (e.g., "client, proxy1, proxy2")
			// We want the leftmost non-trusted-proxy IP
			ips := splitIP(ip)
			for i := 0; i < len(ips); i++ {
				if !rl.isTrustedProxy(ips[i]) {
					return ips[i]
				}
			}
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If there's an error, just use RemoteAddr as is
		return r.RemoteAddr
	}
	return ip

// isTrustedProxy checks if an IP is a trusted proxy
func (rl *RateLimiter) isTrustedProxy(ip string) bool {
	for _, trustedProxy := range rl.config.TrustedProxies {
		if ip == trustedProxy {
			return true
		}
	}
	return false

// splitIP splits a comma-separated list of IPs
func splitIP(ip string) []string {
	ips := make([]string, 0)
	for _, s := range split(ip, ',') {
		ips = append(ips, s)
	}
	return ips

// split splits a string by a separator and trims spaces
func split(s string, sep rune) []string {
	var result []string
	var builder []rune
	for _, r := range s {
		if r == sep {
			if len(builder) > 0 {
				result = append(result, string(builder))
				builder = builder[:0]
			}
		} else if r != ' ' && r != '\t' {
			builder = append(builder, r)
		}
	}
	if len(builder) > 0 {
		result = append(result, string(builder))
	}
	return result

// RateLimiterStats represents statistics about the rate limiter
type RateLimiterStats struct {
	// ActiveLimiters is the number of active rate limiters
	ActiveLimiters int `json:"active_limiters"`
	// RequestsPerMinute is the configured requests per minute
	RequestsPerMinute int `json:"requests_per_minute"`
	// BurstSize is the configured burst size
	BurstSize int `json:"burst_size"`
	// ExemptPathsCount is the number of exempt paths
	ExemptPathsCount int `json:"exempt_paths_count"`
	// ExemptIPsCount is the number of exempt IPs
	ExemptIPsCount int `json:"exempt_ips_count"`

// GetStatistics returns statistics about the rate limiter
func (rl *RateLimiter) GetStatistics() *RateLimiterStats {
	rl.mu.RLock()
	activeLimiters := len(rl.limiters)
	rl.mu.RUnlock()

	return &RateLimiterStats{
		ActiveLimiters:   activeLimiters,
		RequestsPerMinute: rl.config.RequestsPerMinute,
		BurstSize:        rl.config.BurstSize,
		ExemptPathsCount: len(rl.config.ExemptPaths),
		ExemptIPsCount:   len(rl.exemptIPs) + len(rl.exemptCIDR),
	}

// Middleware returns a middleware function for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the client IP
		clientIP := rl.GetClientIP(r)

		// Check if the client is exempt
		if rl.IsExempt(clientIP, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get the limiter for this client
		limiter := rl.GetLimiter(clientIP)

		// Check if the request is allowed
		if !limiter.Allow() {
			// Return a 429 Too Many Requests response
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","code":"RATE_LIMIT_EXCEEDED"}`))
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
}
}
}
}
}
}
}
}
}
