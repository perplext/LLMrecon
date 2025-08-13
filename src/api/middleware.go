package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// Context keys
type contextKey string

const (
	contextKeyUser      contextKey = "user"
	contextKeyRequestID contextKey = "request_id"
	contextKeyAPIKey    contextKey = "api_key"
)

// RateLimiter tracks rate limits per API key
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     int
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(ratePerMinute int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     ratePerMinute,
		burst:    ratePerMinute, // Allow burst equal to rate
	}
}

// GetLimiter returns a rate limiter for the given key
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()
	
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(rl.rate)), rl.burst)
		rl.mu.Lock()
		rl.limiters[key] = limiter
		rl.mu.Unlock()
	}
	
	return limiter
}

// Global rate limiter instance
var globalRateLimiter *RateLimiter

// loggingMiddleware logs all API requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Generate request ID
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), contextKeyRequestID, requestID)
		r = r.WithContext(ctx)
		
		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Log request
		log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("API request started")
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Log response
		duration := time.Since(start)
		log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Msg("API request completed")
	})
}

// corsMiddleware handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Configure based on config
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// jsonContentTypeMiddleware sets JSON content type
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// authMiddleware handles API authentication
func authMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.EnableAuth {
				next.ServeHTTP(w, r)
				return
			}
			
			// Extract API key from header or query parameter
			apiKey := extractAPIKey(r)
			if apiKey == "" {
				writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Missing API key"))
				return
			}
			
			// Validate API key
			if !isValidAPIKey(apiKey, config.APIKeys) {
				writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Invalid API key"))
				return
			}
			
			// Add API key to context
			ctx := context.WithValue(r.Context(), contextKeyAPIKey, apiKey)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimitMiddleware implements rate limiting
func rateLimitMiddleware(config *Config) func(http.Handler) http.Handler {
	// Initialize global rate limiter
	if globalRateLimiter == nil {
		globalRateLimiter = NewRateLimiter(config.RateLimit)
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.EnableRateLimit {
				next.ServeHTTP(w, r)
				return
			}
			
			// Get API key from context
			apiKey, ok := r.Context().Value(contextKeyAPIKey).(string)
			if !ok {
				apiKey = "anonymous"
			}
			
			// Get limiter for this API key
			limiter := globalRateLimiter.GetLimiter(apiKey)
			
			// Check rate limit
			if !limiter.Allow() {
				writeError(w, http.StatusTooManyRequests, NewAPIError(ErrCodeRateLimited, "Rate limit exceeded"))
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts API key from request
func extractAPIKey(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}
	
	// Check X-API-Key header
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	
	// Check query parameter (for convenience, but less secure)
	if key := r.URL.Query().Get("api_key"); key != "" {
		return key
	}
	
	return ""
}

// isValidAPIKey checks if the provided API key is valid
func isValidAPIKey(key string, validKeys []string) bool {
	normalizedKey := normalizeAPIKey(key)
	for _, validKey := range validKeys {
		if normalizeAPIKey(validKey) == normalizedKey {
			return true
		}
	}
	return false
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(data)
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, apiErr *APIError) {
	response := Response{
		Success: false,
		Error: &Error{
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Details: apiErr.Details,
		},
	}
	writeJSON(w, status, response)
}

// writeSuccess writes success response
func writeSuccess(w http.ResponseWriter, data interface{}) {
	response := Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Version: APIVersion,
		},
	}
	writeJSON(w, http.StatusOK, response)
}

// writeSuccessWithMeta writes success response with metadata
func writeSuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	if meta == nil {
		meta = &Meta{}
	}
	meta.Version = APIVersion
	
	response := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
	writeJSON(w, http.StatusOK, response)
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), generateRandomInt())
}

// generateRandomInt generates a random integer
func generateRandomInt() int {
	return int(time.Now().UnixNano() % 1000000)
}

// Error codes are defined in types.go

// paginate applies pagination to a slice
func paginate(page, perPage, total int) (offset, limit int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	
	offset = (page - 1) * perPage
	limit = perPage
	
	if offset >= total {
		offset = total
		limit = 0
	} else if offset+limit > total {
		limit = total - offset
	}
	
	return offset, limit
}

// calculateTotalPages calculates total pages for pagination
func calculateTotalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	pages := total / perPage
	if total%perPage > 0 {
		pages++
	}
	return pages
}