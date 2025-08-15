package api

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// Security middleware enhancements

// SecurityHeaders represents security headers configuration
type SecurityHeaders struct {
	ContentSecurityPolicy   string
	XContentTypeOptions     string
	XFrameOptions          string
	XSSProtection          string
	StrictTransportSecurity string
	ReferrerPolicy         string
	PermissionsPolicy      string
}

// DefaultSecurityHeaders returns default security headers
func DefaultSecurityHeaders() SecurityHeaders {
	return SecurityHeaders{
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'",
		XContentTypeOptions:     "nosniff",
		XFrameOptions:          "DENY",
		XSSProtection:          "1; mode=block",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		ReferrerPolicy:         "strict-origin-when-cross-origin",
		PermissionsPolicy:      "geolocation=(), microphone=(), camera=()",
	}

// securityHeadersMiddleware adds security headers to responses
func securityHeadersMiddleware(headers SecurityHeaders) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			if headers.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", headers.ContentSecurityPolicy)
			}
			if headers.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", headers.XContentTypeOptions)
			}
			if headers.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", headers.XFrameOptions)
			}
			if headers.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", headers.XSSProtection)
			}
			if headers.StrictTransportSecurity != "" && r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", headers.StrictTransportSecurity)
			}
			if headers.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", headers.ReferrerPolicy)
			}
			if headers.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", headers.PermissionsPolicy)
			}
			
			next.ServeHTTP(w, r)
		})
	}

// IPWhitelist manages IP whitelisting
type IPWhitelist struct {
	allowedIPs map[string]bool
	allowedCIDRs []string
	mu         sync.RWMutex
}

// NewIPWhitelist creates a new IP whitelist
func NewIPWhitelist(ips []string, cidrs []string) *IPWhitelist {
	whitelist := &IPWhitelist{
		allowedIPs:   make(map[string]bool),
		allowedCIDRs: cidrs,
	}
	
	for _, ip := range ips {
		whitelist.allowedIPs[ip] = true
	}
	
	return whitelist

// IsAllowed checks if an IP is whitelisted
func (w *IPWhitelist) IsAllowed(ip string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	// Check exact match
	if w.allowedIPs[ip] {
		return true
	}
	
	// Check CIDR ranges (simplified - real implementation would parse CIDR)
	// This is a placeholder for CIDR matching logic
	
	return false

// ipWhitelistMiddleware enforces IP whitelisting
func ipWhitelistMiddleware(whitelist *IPWhitelist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if whitelist != nil {
				clientIP := getClientIP(r)
				if !whitelist.IsAllowed(clientIP) {
					writeError(w, http.StatusForbidden, NewAPIError(ErrCodeForbidden, "Access denied from this IP"))
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}

// RequestSizeLimit limits request body size
func requestSizeLimitMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}

// ScopeValidator validates API key scopes
type ScopeValidator struct {
	requiredScopes map[string][]string // endpoint -> required scopes
}

// NewScopeValidator creates a new scope validator
func NewScopeValidator() *ScopeValidator {
	return &ScopeValidator{
		requiredScopes: map[string][]string{
			"/api/v1/scans":     {"scan:write"},
			"/api/v1/templates": {"template:read"},
			"/api/v1/modules":   {"module:read"},
			"/api/v1/update":    {"system:update"},
			"/api/v1/auth/keys": {"admin"},
			"/api/v1/auth/users": {"admin"},
		},
	}

// scopeValidationMiddleware validates API key scopes
func scopeValidationMiddleware(validator *ScopeValidator, authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from context
			apiKeyStr, ok := r.Context().Value(contextKeyAPIKey).(string)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			
			// Validate API key and get details
			apiKey, err := authService.ValidateAPIKey(apiKeyStr)
			if err != nil {
				writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Invalid API key"))
				return
			}
			
			// Check scopes for endpoint
			path := r.URL.Path
			requiredScopes, exists := validator.requiredScopes[path]
			if !exists {
				// Check for prefix match
				for endpoint, scopes := range validator.requiredScopes {
					if strings.HasPrefix(path, endpoint) {
						requiredScopes = scopes
						break
					}
				}
			}
			
			if len(requiredScopes) > 0 {
				hasScope := false
				for _, required := range requiredScopes {
					for _, scope := range apiKey.Scopes {
						if scope == required || scope == "admin" {
							hasScope = true
							break
						}
					}
					if hasScope {
						break
					}
				}
				
				if !hasScope {
					writeError(w, http.StatusForbidden, NewAPIError(ErrCodeForbidden, "Insufficient permissions"))
					return
				}
			}
			
			// Add API key details to context
			ctx := context.WithValue(r.Context(), "apiKeyDetails", apiKey)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}

// AuditLogger logs security-relevant events
type AuditLogger struct {
	logSensitive bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logSensitive bool) *AuditLogger {
	return &AuditLogger{
		logSensitive: logSensitive,
	}

// auditLoggingMiddleware logs security events
func auditLoggingMiddleware(logger *AuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Extract relevant info
			requestID := r.Context().Value(contextKeyRequestID).(string)
			clientIP := getClientIP(r)
			userAgent := r.UserAgent()
			
			// Log security event start
			event := log.Info().
				Str("event_type", "api_access").
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("client_ip", clientIP).
				Str("user_agent", userAgent)
			
			// Add auth info if available
			if apiKey, ok := r.Context().Value(contextKeyAPIKey).(string); ok {
				event.Str("api_key", maskAPIKey(apiKey))
			}
			if userID, ok := r.Context().Value("userID").(string); ok {
				event.Str("user_id", userID)
			}
			
			event.Msg("API request initiated")
			
			// Process request
			next.ServeHTTP(wrapped, r)
			
			// Log completion
			duration := time.Since(start)
			completion := log.Info().
				Str("event_type", "api_access_complete").
				Str("request_id", requestID).
				Int("status", wrapped.statusCode).
				Dur("duration", duration)
			
			if wrapped.statusCode >= 400 {
				completion = log.Warn().
					Str("event_type", "api_access_failed").
					Str("request_id", requestID).
					Int("status", wrapped.statusCode).
					Dur("duration", duration)
			}
			
			completion.Msg("API request completed")
		})
	}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	
	return ip

// maskAPIKey masks an API key for logging
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]

// timeoutMiddleware adds request timeout
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			
			r = r.WithContext(ctx)
			
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()
			
			select {
			case <-done:
				// Request completed
			case <-ctx.Done():
				// Timeout occurred
				writeError(w, http.StatusRequestTimeout, NewAPIError(ErrCodeTimeout, "Request timeout"))
			}
		})
	}

// compressionMiddleware adds response compression
func compressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		
		// Wrap response writer with gzip writer
		w.Header().Set("Content-Encoding", "gzip")
		// Implementation would use gzip.Writer here
		
		next.ServeHTTP(w, r)
	})

// requestIDMiddleware ensures request ID is set
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		ctx := context.WithValue(r.Context(), contextKeyRequestID, requestID)
		r = r.WithContext(ctx)
		
		w.Header().Set("X-Request-ID", requestID)
		
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
}
