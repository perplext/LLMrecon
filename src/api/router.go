package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

// Router creates and configures the API router
func NewRouter(config *Config) *mux.Router {
	r := mux.NewRouter()
	
	// Initialize services
	authService := NewAuthService(config)
	scopeValidator := NewScopeValidator()
	var ipWhitelist *IPWhitelist
	if config.EnableIPWhitelist {
		ipWhitelist = NewIPWhitelist(config.WhitelistedIPs, config.WhitelistedCIDRs)
	}
	auditLogger := NewAuditLogger(config.EnableAuditLogging)
	
	// API v1 routes
	v1 := r.PathPrefix("/api/v1").Subrouter()
	
	// Apply global middleware
	v1.Use(requestIDMiddleware)
	v1.Use(loggingMiddleware)
	if config.EnableSecurityHeaders {
		v1.Use(securityHeadersMiddleware(config.SecurityHeaders))
	}
	if config.EnableIPWhitelist {
		v1.Use(ipWhitelistMiddleware(ipWhitelist))
	}
	v1.Use(corsMiddleware)
	v1.Use(jsonContentTypeMiddleware)
	v1.Use(requestSizeLimitMiddleware(config.MaxRequestSize))
	if config.EnableCompression {
		v1.Use(compressionMiddleware)
	}
	if config.EnableAuditLogging {
		v1.Use(auditLoggingMiddleware(auditLogger))
	}
	
	// Health check (no auth required)
	v1.HandleFunc("/health", handleHealth).Methods("GET")
	
	// Version info (no auth required)
	v1.HandleFunc("/version", handleVersion).Methods("GET")
	
	// Add auth service to context
	v1.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "authService", authService)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	
	// Protected routes (API key auth)
	protected := v1.PathPrefix("").Subrouter()
	protected.Use(authMiddleware(config))
	protected.Use(rateLimitMiddleware(config))
	protected.Use(scopeValidationMiddleware(scopeValidator, authService))
	
	// JWT protected routes
	jwtProtected := v1.PathPrefix("").Subrouter()
	jwtProtected.Use(jwtMiddleware(authService))
	
	// Scan endpoints
	protected.HandleFunc("/scans", handleCreateScan).Methods("POST")
	protected.HandleFunc("/scans", handleListScans).Methods("GET")
	protected.HandleFunc("/scans/{id}", handleGetScan).Methods("GET")
	protected.HandleFunc("/scans/{id}", handleCancelScan).Methods("DELETE")
	protected.HandleFunc("/scans/{id}/results", handleGetScanResults).Methods("GET")
	
	// Template endpoints
	protected.HandleFunc("/templates", handleListTemplates).Methods("GET")
	protected.HandleFunc("/templates/{id}", handleGetTemplate).Methods("GET")
	protected.HandleFunc("/templates/categories", handleListCategories).Methods("GET")
	
	// Module endpoints
	protected.HandleFunc("/modules", handleListModules).Methods("GET")
	protected.HandleFunc("/modules/{id}", handleGetModule).Methods("GET")
	protected.HandleFunc("/modules/{id}/config", handleUpdateModuleConfig).Methods("PUT")
	
	// Update endpoints
	protected.HandleFunc("/update", handleCheckUpdate).Methods("GET")
	protected.HandleFunc("/update", handlePerformUpdate).Methods("POST")
	
	// Bundle endpoints
	protected.HandleFunc("/bundles", handleListBundles).Methods("GET")
	protected.HandleFunc("/bundles/export", handleExportBundle).Methods("POST")
	protected.HandleFunc("/bundles/import", handleImportBundle).Methods("POST")
	
	// Compliance endpoints
	protected.HandleFunc("/compliance/report", handleGenerateComplianceReport).Methods("POST")
	protected.HandleFunc("/compliance/check", handleCheckCompliance).Methods("GET")
	
	// Authentication endpoints (no auth required)
	v1.HandleFunc("/auth/login", handleLogin).Methods("POST")
	v1.HandleFunc("/auth/refresh", handleRefreshToken).Methods("POST")
	v1.HandleFunc("/auth/register", handleCreateUser).Methods("POST")
	
	// User management endpoints (JWT auth required)
	jwtProtected.HandleFunc("/auth/password", handleUpdatePassword).Methods("PUT")
	jwtProtected.HandleFunc("/auth/profile", handleGetProfile).Methods("GET")
	
	// API key management endpoints (JWT auth required, admin only)
	jwtProtected.HandleFunc("/auth/keys", handleCreateAPIKey).Methods("POST")
	jwtProtected.HandleFunc("/auth/keys", handleListAPIKeys).Methods("GET")
	jwtProtected.HandleFunc("/auth/keys/{id}", handleGetAPIKey).Methods("GET")
	jwtProtected.HandleFunc("/auth/keys/{id}", handleRevokeAPIKey).Methods("DELETE")
	
	// OpenAPI documentation
	v1.HandleFunc("/openapi.json", handleOpenAPISpec).Methods("GET")
	
	// Static documentation (Swagger UI)
	if config.EnableSwaggerUI {
		r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./api/docs/"))))
	}
	
	return r

// Route represents an API route definition
type Route struct {
	Name        string
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
	RequireAuth bool
}

// Config holds API server configuration
type Config struct {
	Port                  int
	Host                  string
	APIKeys               []string
	EnableAuth            bool
	EnableRateLimit       bool
	RateLimit             int    // requests per minute
	EnableCORS            bool
	AllowedOrigins        []string
	EnableSwaggerUI       bool
	LogLevel              string
	TLSCert               string
	TLSKey                string
	JWTSecret             string
	JWTExpiration         int    // hours
	EnableSecurityHeaders bool
	SecurityHeaders       SecurityHeaders
	EnableIPWhitelist     bool
	WhitelistedIPs        []string
	WhitelistedCIDRs      []string
	MaxRequestSize        int64  // bytes
	RequestTimeout        int    // seconds
	EnableCompression     bool
	EnableAuditLogging    bool
	EnableMetrics         bool   // enable metrics collection

// DefaultConfig returns default API configuration
func DefaultConfig() *Config {
	return &Config{
		Port:                  8080,
		Host:                  "localhost",
		EnableAuth:            true,
		EnableRateLimit:       true,
		RateLimit:             60,
		EnableCORS:            true,
		AllowedOrigins:        []string{"*"},
		EnableSwaggerUI:       true,
		LogLevel:              "info",
		JWTSecret:             "change-me-in-production",
		JWTExpiration:         24,
		EnableSecurityHeaders: true,
		SecurityHeaders:       DefaultSecurityHeaders(),
		EnableIPWhitelist:     false,
		WhitelistedIPs:        []string{},
		WhitelistedCIDRs:      []string{},
		MaxRequestSize:        10 * 1024 * 1024, // 10MB
		RequestTimeout:        30,
		EnableCompression:     true,
		EnableAuditLogging:    true,
	}

// ValidateConfig validates API configuration
func ValidateConfig(config *Config) error {
	if config.Port < 1 || config.Port > 65535 {
		return NewAPIError("INVALID_CONFIG", "Invalid port number")
	}
	
	if config.EnableAuth && len(config.APIKeys) == 0 {
		return NewAPIError("INVALID_CONFIG", "Authentication enabled but no API keys provided")
	}
	
	if config.EnableRateLimit && config.RateLimit < 1 {
		return NewAPIError("INVALID_CONFIG", "Invalid rate limit value")
	}
	
	if config.TLSCert != "" && config.TLSKey == "" {
		return NewAPIError("INVALID_CONFIG", "TLS certificate provided but no key")
	}
	
	if config.EnableAuth && config.JWTSecret == "change-me-in-production" {
		return NewAPIError("INVALID_CONFIG", "JWT secret must be changed from default value")
	}
	
	if config.MaxRequestSize < 1024 {
		return NewAPIError("INVALID_CONFIG", "Max request size too small")
	}
	
	if config.RequestTimeout < 1 {
		return NewAPIError("INVALID_CONFIG", "Request timeout must be at least 1 second")
	}
	
	return nil

// APIError represents an API error
type APIError struct {
	Code    string
	Message string
	Details string
}

func (e *APIError) Error() string {
	return e.Message

// NewAPIError creates a new API error
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(code, message, details string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}

// Error codes are defined in types.go

// GetRoutes returns all API routes for documentation
func GetRoutes() []Route {
	return []Route{
		// Health & Version
		{Name: "Health", Method: "GET", Pattern: "/api/v1/health", RequireAuth: false},
		{Name: "Version", Method: "GET", Pattern: "/api/v1/version", RequireAuth: false},
		
		// Authentication
		{Name: "Login", Method: "POST", Pattern: "/api/v1/auth/login", RequireAuth: false},
		{Name: "RefreshToken", Method: "POST", Pattern: "/api/v1/auth/refresh", RequireAuth: false},
		{Name: "Register", Method: "POST", Pattern: "/api/v1/auth/register", RequireAuth: false},
		
		// User Management (JWT auth)
		{Name: "GetProfile", Method: "GET", Pattern: "/api/v1/auth/profile", RequireAuth: true},
		{Name: "UpdatePassword", Method: "PUT", Pattern: "/api/v1/auth/password", RequireAuth: true},
		
		// API Key Management (JWT auth)
		{Name: "CreateAPIKey", Method: "POST", Pattern: "/api/v1/auth/keys", RequireAuth: true},
		{Name: "ListAPIKeys", Method: "GET", Pattern: "/api/v1/auth/keys", RequireAuth: true},
		{Name: "GetAPIKey", Method: "GET", Pattern: "/api/v1/auth/keys/{id}", RequireAuth: true},
		{Name: "RevokeAPIKey", Method: "DELETE", Pattern: "/api/v1/auth/keys/{id}", RequireAuth: true},
		
		// Scans
		{Name: "CreateScan", Method: "POST", Pattern: "/api/v1/scans", RequireAuth: true},
		{Name: "ListScans", Method: "GET", Pattern: "/api/v1/scans", RequireAuth: true},
		{Name: "GetScan", Method: "GET", Pattern: "/api/v1/scans/{id}", RequireAuth: true},
		{Name: "CancelScan", Method: "DELETE", Pattern: "/api/v1/scans/{id}", RequireAuth: true},
		{Name: "GetScanResults", Method: "GET", Pattern: "/api/v1/scans/{id}/results", RequireAuth: true},
		
		// Templates
		{Name: "ListTemplates", Method: "GET", Pattern: "/api/v1/templates", RequireAuth: true},
		{Name: "GetTemplate", Method: "GET", Pattern: "/api/v1/templates/{id}", RequireAuth: true},
		{Name: "ListCategories", Method: "GET", Pattern: "/api/v1/templates/categories", RequireAuth: true},
		
		// Modules
		{Name: "ListModules", Method: "GET", Pattern: "/api/v1/modules", RequireAuth: true},
		{Name: "GetModule", Method: "GET", Pattern: "/api/v1/modules/{id}", RequireAuth: true},
		{Name: "UpdateModuleConfig", Method: "PUT", Pattern: "/api/v1/modules/{id}/config", RequireAuth: true},
		
		// Updates
		{Name: "CheckUpdate", Method: "GET", Pattern: "/api/v1/update", RequireAuth: true},
		{Name: "PerformUpdate", Method: "POST", Pattern: "/api/v1/update", RequireAuth: true},
		
		// Bundles
		{Name: "ListBundles", Method: "GET", Pattern: "/api/v1/bundles", RequireAuth: true},
		{Name: "ExportBundle", Method: "POST", Pattern: "/api/v1/bundles/export", RequireAuth: true},
		{Name: "ImportBundle", Method: "POST", Pattern: "/api/v1/bundles/import", RequireAuth: true},
		
		// Compliance
		{Name: "GenerateComplianceReport", Method: "POST", Pattern: "/api/v1/compliance/report", RequireAuth: true},
		{Name: "CheckCompliance", Method: "GET", Pattern: "/api/v1/compliance/check", RequireAuth: true},
		
		// Documentation
		{Name: "OpenAPISpec", Method: "GET", Pattern: "/api/v1/openapi.json", RequireAuth: false},
	}

// normalizeAPIKey is implemented in auth_service.go
}
}
}
}
}
