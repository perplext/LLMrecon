// Package api provides a RESTful API for the access control system
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	access "github.com/perplext/LLMrecon/src/security/access"
)

// Permission is a type alias for permission strings used in middleware
type Permission = string

// Response is a standard API response format
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// WriteJSON writes a JSON response to the HTTP response writer
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data) // Best effort, headers already sent
}

// WriteErrorResponse writes an error response to the HTTP response writer
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := Response{
		Success: false,
		Error:   message,
	}
	WriteJSON(w, statusCode, response)
}

// WriteSuccessResponse writes a success response to the HTTP response writer
func WriteSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	WriteJSON(w, statusCode, response)
}

// AuthMiddleware handles authentication for API requests
type AuthMiddleware struct {
	accessManager access.AccessControlManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(accessManager access.AccessControlManager) *AuthMiddleware {
	return &AuthMiddleware{
		accessManager: accessManager,
	}
}

// Middleware is the HTTP middleware function for authentication
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteErrorResponse(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Check if the header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format, expected 'Bearer <token>'")
			return
		}

		token := parts[1]

		// Validate the token via AuthManager
		authManager := m.accessManager.GetAuthManager()
		session, err := authManager.ValidateSession(r.Context(), token)
		if err != nil {
			WriteErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Check if MFA is required but not completed
		if !session.MFACompleted {
			// Get the user to check if MFA is enabled
			user, err := authManager.GetUserByID(r.Context(), session.UserID)
			if err == nil && user.MFAEnabled {
				WriteErrorResponse(w, http.StatusForbidden, "MFA verification required")
				return
			}
		}

		// Get the user
		user, err := authManager.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Check if the user is active
		if !user.Active {
			WriteErrorResponse(w, http.StatusForbidden, "User account is inactive")
			return
		}

		// Check if the user is locked
		if user.Locked {
			WriteErrorResponse(w, http.StatusForbidden, "User account is locked")
			return
		}

		// Store user and session in request context
		ctx := context.WithValue(r.Context(), "user", user)
		ctx = context.WithValue(ctx, "session", session)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RBACMiddleware handles role-based access control for API requests
type RBACMiddleware struct {
	accessManager access.AccessControlManager
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(accessManager access.AccessControlManager) *RBACMiddleware {
	return &RBACMiddleware{
		accessManager: accessManager,
	}
}

// RequirePermission returns a middleware function that requires a specific permission
func (m *RBACMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := r.Context().Value("user").(*access.User)
			if !ok {
				WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			// Check if the user has the required permission
			rbacManager := m.accessManager.GetRBACManager()
			hasPermission, err := rbacManager.HasPermission(user.ID, permission)
			if err != nil || !hasPermission {
				WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns a middleware function that requires a specific role
func (m *RBACMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := r.Context().Value("user").(*access.User)
			if !ok {
				WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			// Check if the user has the required role
			rbacManager := m.accessManager.GetRBACManager()
			hasRole, err := rbacManager.HasRole(user.ID, role)
			if err != nil || !hasRole {
				WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware handles logging of API requests
type LoggingMiddleware struct {
	accessManager access.AccessControlManager
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(accessManager access.AccessControlManager) *LoggingMiddleware {
	return &LoggingMiddleware{
		accessManager: accessManager,
	}
}

// Middleware is the HTTP middleware function for logging
func (m *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the request
		duration := time.Since(start)

		// Get user ID if available
		var userID string
		if user, ok := r.Context().Value("user").(*access.User); ok {
			userID = user.ID
		}

		// Get client IP
		ip := getClientIP(r)

		// Log the request to the audit log
		auditLog := &access.AuditLog{
			Action:     access.AuditActionRead,
			Resource:   "API",
			ResourceID: r.URL.Path,
			Severity:   access.AuditSeverityInfo,
			Status:     fmt.Sprintf("%d", rw.statusCode),
			UserID:     userID,
			IPAddress:  ip,
			UserAgent:  r.UserAgent(),
			Metadata: map[string]interface{}{
				"method":   r.Method,
				"path":     r.URL.Path,
				"query":    r.URL.RawQuery,
				"duration": duration.Milliseconds(),
				"status":   rw.statusCode,
			},
		}

		// Only log errors or important requests to the audit log
		if rw.statusCode >= 400 || strings.Contains(r.URL.Path, "/auth/") {
			auditLogger := m.accessManager.GetAuditLogger()
			if err := auditLogger.LogAudit(r.Context(), auditLog); err != nil {
				fmt.Printf("Failed to log audit event: %v\n", err)
			}
		}

		// Log to stdout for all requests
		fmt.Printf("%s - %s %s %s - %d - %s - %dms\n",
			time.Now().Format(time.RFC3339),
			ip,
			r.Method,
			r.URL.Path,
			rw.statusCode,
			userID,
			duration.Milliseconds(),
		)
	})
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and calls the underlying ResponseWriter
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RateLimitMiddleware handles rate limiting for API requests
type RateLimitMiddleware struct {
	// requestsPerMinute is the number of requests allowed per minute per IP
	requestsPerMinute int

	// ipRequests maps IP addresses to request counts and timestamps
	ipRequests map[string]*ipRequestInfo

	// mutex protects the ipRequests map
	mutex sync.Mutex

	// cleanupInterval is the interval at which to clean up expired entries
	cleanupInterval time.Duration

	// stopCleanup is a channel to signal the cleanup goroutine to stop
	stopCleanup chan struct{}
}

// ipRequestInfo tracks request information for an IP address
type ipRequestInfo struct {
	count     int
	timestamp time.Time
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(requestsPerMinute int) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		requestsPerMinute: requestsPerMinute,
		ipRequests:        make(map[string]*ipRequestInfo),
		cleanupInterval:   5 * time.Minute,
		stopCleanup:       make(chan struct{}),
	}

	// Start cleanup goroutine
	go m.cleanup()

	return m
}

// Stop stops the cleanup goroutine
func (m *RateLimitMiddleware) Stop() {
	close(m.stopCleanup)
}

// cleanup periodically removes expired entries from the ipRequests map
func (m *RateLimitMiddleware) cleanup() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mutex.Lock()
			now := time.Now()
			for ip, info := range m.ipRequests {
				// Remove entries older than 1 minute
				if now.Sub(info.timestamp) > time.Minute {
					delete(m.ipRequests, ip)
				}
			}
			m.mutex.Unlock()
		case <-m.stopCleanup:
			return
		}
	}
}

// Middleware is the HTTP middleware function for rate limiting
func (m *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := getClientIP(r)

		// Check rate limit
		m.mutex.Lock()
		info, exists := m.ipRequests[ip]
		now := time.Now()

		if !exists || now.Sub(info.timestamp) > time.Minute {
			// First request in the current minute
			m.ipRequests[ip] = &ipRequestInfo{
				count:     1,
				timestamp: now,
			}
		} else {
			// Increment request count
			info.count++

			// Check if rate limit exceeded
			if info.count > m.requestsPerMinute {
				m.mutex.Unlock()
				WriteErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}
		}
		m.mutex.Unlock()

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check for X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Use RemoteAddr as fallback
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, return the whole RemoteAddr
		return r.RemoteAddr
	}
	return ip
}
