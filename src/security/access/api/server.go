// Package api provides a RESTful API for the access control system
package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/perplext/LLMrecon/src/security/access"
)

// APIConfig contains configuration for the API server
type APIConfig struct {
	// Port is the port to listen on
	Port int

	// BasePath is the base path for all API endpoints
	BasePath string

	// EnableCORS enables Cross-Origin Resource Sharing
	EnableCORS bool

	// AllowedOrigins is a list of allowed origins for CORS
	AllowedOrigins []string

	// EnableRateLimit enables rate limiting
	EnableRateLimit bool

	// RateLimitPerMinute is the number of requests allowed per minute per IP
	RateLimitPerMinute int

	// EnableRequestLogging enables logging of all API requests
	EnableRequestLogging bool
}

// DefaultAPIConfig returns a default API configuration
func DefaultAPIConfig() *APIConfig {
	return &APIConfig{
		Port:               8080,
		BasePath:           "/api/v1",
		EnableCORS:         true,
		AllowedOrigins:     []string{"*"},
		EnableRateLimit:    true,
		RateLimitPerMinute: 60,
		EnableRequestLogging: true,
	}
}

// Server is the API server for the access control system
type Server struct {
	// Configuration
	config *APIConfig

	// HTTP server
	httpServer *http.Server

	// Router
	router *mux.Router

	// Access control manager
	accessManager access.AccessControlManager

	// Middleware
	authMiddleware   *AuthMiddleware
	rbacMiddleware   *RBACMiddleware
	loggingMiddleware *LoggingMiddleware
	rateLimitMiddleware *RateLimitMiddleware
}

// NewServer creates a new API server
func NewServer(config *APIConfig, accessManager access.AccessControlManager) *Server {
	router := mux.NewRouter()

	// Create server
	server := &Server{
		config:        config,
		router:        router,
		accessManager: accessManager,
	}

	// Create middleware
	server.authMiddleware = NewAuthMiddleware(accessManager)
	server.rbacMiddleware = NewRBACMiddleware(accessManager)
	server.loggingMiddleware = NewLoggingMiddleware(accessManager)
	server.rateLimitMiddleware = NewRateLimitMiddleware(config.RateLimitPerMinute)

	// Configure HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Register routes
	server.registerRoutes()

	return server
}

// Start starts the API server
func (s *Server) Start() error {
	log.Printf("Starting API server on port %d", s.config.Port)
	return s.httpServer.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping API server")
	return s.httpServer.Shutdown(ctx)
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	// Create API subrouter
	api := s.router.PathPrefix(s.config.BasePath).Subrouter()

	// Apply global middleware
	if s.config.EnableRequestLogging {
		api.Use(s.loggingMiddleware.Middleware)
	}
	if s.config.EnableRateLimit {
		api.Use(s.rateLimitMiddleware.Middleware)
	}
	if s.config.EnableCORS {
		api.Use(s.corsMiddleware)
	}

	// Authentication routes
	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", s.handleLogin).Methods("POST")
	authRouter.HandleFunc("/logout", s.handleLogout).Methods("POST")
	authRouter.HandleFunc("/refresh", s.handleRefreshToken).Methods("POST")
	authRouter.HandleFunc("/status", s.handleAuthStatus).Methods("GET")
	authRouter.HandleFunc("/mfa/verify", s.handleMFAVerify).Methods("POST")
	
	// User routes (require authentication)
	userRouter := api.PathPrefix("/users").Subrouter()
	userRouter.Use(s.authMiddleware.Middleware)
	userRouter.HandleFunc("", s.handleListUsers).Methods("GET")
	userRouter.HandleFunc("", s.handleCreateUser).Methods("POST")
	userRouter.HandleFunc("/{id}", s.handleGetUser).Methods("GET")
	userRouter.HandleFunc("/{id}", s.handleUpdateUser).Methods("PUT")
	userRouter.HandleFunc("/{id}", s.handleDeleteUser).Methods("DELETE")
	userRouter.HandleFunc("/{id}/password", s.handleResetPassword).Methods("POST")
	userRouter.HandleFunc("/{id}/lock", s.handleLockUser).Methods("POST")
	userRouter.HandleFunc("/{id}/unlock", s.handleUnlockUser).Methods("POST")
	userRouter.HandleFunc("/{id}/mfa", s.handleManageUserMFA).Methods("PUT")
	
	// Role routes (require authentication and admin permission)
	roleRouter := api.PathPrefix("/roles").Subrouter()
	roleRouter.Use(s.authMiddleware.Middleware)
	roleRouter.HandleFunc("", s.handleListRoles).Methods("GET")
	roleRouter.HandleFunc("", s.handleCreateRole).Methods("POST")
	roleRouter.HandleFunc("/{name}", s.handleGetRole).Methods("GET")
	roleRouter.HandleFunc("/{name}", s.handleUpdateRole).Methods("PUT")
	roleRouter.HandleFunc("/{name}", s.handleDeleteRole).Methods("DELETE")
	roleRouter.HandleFunc("/{name}/permissions", s.handleAddPermission).Methods("POST")
	roleRouter.HandleFunc("/{name}/permissions/{permission}", s.handleRemovePermission).Methods("DELETE")
	
	// Audit routes (require authentication and audit permission)
	auditRouter := api.PathPrefix("/audit").Subrouter()
	auditRouter.Use(s.authMiddleware.Middleware)
	auditRouter.HandleFunc("", s.handleListAuditLogs).Methods("GET")
	auditRouter.HandleFunc("/{id}", s.handleGetAuditLog).Methods("GET")
	auditRouter.HandleFunc("/export", s.handleExportAuditLogs).Methods("GET")
	
	// Security incident routes (require authentication)
	incidentRouter := api.PathPrefix("/incidents").Subrouter()
	incidentRouter.Use(s.authMiddleware.Middleware)
	incidentRouter.HandleFunc("", s.handleListIncidents).Methods("GET")
	incidentRouter.HandleFunc("", s.handleCreateIncident).Methods("POST")
	incidentRouter.HandleFunc("/{id}", s.handleGetIncident).Methods("GET")
	incidentRouter.HandleFunc("/{id}", s.handleUpdateIncident).Methods("PUT")
	incidentRouter.HandleFunc("/{id}", s.handleDeleteIncident).Methods("DELETE")
	
	// Vulnerability routes (require authentication)
	vulnRouter := api.PathPrefix("/vulnerabilities").Subrouter()
	vulnRouter.Use(s.authMiddleware.Middleware)
	vulnRouter.HandleFunc("", s.handleListVulnerabilities).Methods("GET")
	vulnRouter.HandleFunc("", s.handleCreateVulnerability).Methods("POST")
	vulnRouter.HandleFunc("/{id}", s.handleGetVulnerability).Methods("GET")
	vulnRouter.HandleFunc("/{id}", s.handleUpdateVulnerability).Methods("PUT")
	vulnRouter.HandleFunc("/{id}", s.handleDeleteVulnerability).Methods("DELETE")
	
	// Health check route (no authentication required)
	api.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
}

// corsMiddleware handles Cross-Origin Resource Sharing
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Handler methods for authentication routes
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in auth_handlers.go
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in auth_handlers.go
}

func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in auth_handlers.go
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in auth_handlers.go
}

func (s *Server) handleMFAVerify(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in auth_handlers.go
}

// Handler methods for user routes
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleLockUser(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleUnlockUser(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

func (s *Server) handleManageUserMFA(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in user_handlers.go
}

// Handler methods for role routes
func (s *Server) handleListRoles(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in role_handlers.go
}

func (s *Server) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in role_handlers.go
}

func (s *Server) handleGetRole(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in role_handlers.go
}

func (s *Server) handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in role_handlers.go
}

func (s *Server) handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in role_handlers.go
}

func (s *Server) handleAddPermission(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in role_handlers.go
}

func (s *Server) handleRemovePermission(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in role_handlers.go
}

// Handler methods for audit routes
func (s *Server) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in audit_handlers.go
}

func (s *Server) handleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in audit_handlers.go
}

func (s *Server) handleExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in audit_handlers.go
}

// Handler methods for security incident routes
func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

// Handler methods for vulnerability routes
func (s *Server) handleListVulnerabilities(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleCreateVulnerability(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleGetVulnerability(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleUpdateVulnerability(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

func (s *Server) handleDeleteVulnerability(w http.ResponseWriter, r *http.Request) {
	// Implementation will be in security_handlers.go
}

// Health check handler
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
