// Package api provides the HTTP server for the LLMrecon API
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"

	"github.com/perplext/LLMrecon/src/api/scan"
	"github.com/perplext/LLMrecon/src/security"
	"github.com/perplext/LLMrecon/src/security/communication"
	"github.com/perplext/LLMrecon/src/security/prompt"
)

// Server represents the API server
type Server struct {
	server                  *http.Server
	mux                     *http.ServeMux
	scanSvc                 *scan.Service
	scanHdlr                *scan.Handler
	securityManager         *security.SecurityManager
	promptProtectionMiddleware *prompt.PromptProtectionMiddleware
	config                  *ServerConfig
}

// ServerConfig represents the configuration for the API server
type ServerConfig struct {
	// Address is the server address
	Address string
	// UseTLS indicates whether to use TLS
	UseTLS bool
	// CertFile is the path to the certificate file
	CertFile string
	// KeyFile is the path to the key file
	KeyFile string
	// SecurityConfig is the security configuration
	SecurityConfig *security.SecurityConfig
	// PromptProtectionConfig is the configuration for prompt injection protection
	PromptProtectionConfig *prompt.ProtectionConfig

// DefaultServerConfig returns the default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:               ":8080",
		UseTLS:                false,
		SecurityConfig:        security.DefaultSecurityConfig(),
		PromptProtectionConfig: prompt.DefaultProtectionConfig(),
	}

// NewServer creates a new API server
func NewServer(config *ServerConfig) (*Server, error) {
	if config == nil {
		config = DefaultServerConfig()
	}

	// Create security manager
	securityManager, err := security.NewSecurityManager(config.SecurityConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create security manager: %w", err)
	}

	// Create prompt protection middleware
	promptProtectionMiddleware, err := prompt.NewPromptProtectionMiddleware(config.PromptProtectionConfig)
	if err != nil {
		securityManager.Close()
		return nil, fmt.Errorf("failed to create prompt protection middleware: %w", err)
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Create a new HTTP server
	server := &http.Server{
		Addr:         config.Address,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Configure TLS if enabled
	if config.UseTLS {
		// Update TLS configuration
		config.SecurityConfig.TLSConfig.CertFile = config.CertFile
		config.SecurityConfig.TLSConfig.KeyFile = config.KeyFile

		// Configure TLS for server
		tlsConfig, err := communication.ConfigureTLSForServer(config.SecurityConfig.TLSConfig)
		if err != nil {
			securityManager.Close()
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}

		// Set TLS config
		server.TLSConfig = tlsConfig
	}

	// Create scan service and handler
	scanStorage := scan.NewMemoryStorage()
	scanSvc := scan.NewService(scanStorage)
	scanHdlr := scan.NewHandler(scanSvc)

	return &Server{
		server:                  server,
		mux:                     mux,
		scanSvc:                 scanSvc,
		scanHdlr:                scanHdlr,
		securityManager:         securityManager,
		promptProtectionMiddleware: promptProtectionMiddleware,
		config:                  config,
	}, nil

// Start starts the server
func (s *Server) Start() error {
	// Register routes
	s.registerRoutes()

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on %s", s.server.Addr)
		var err error
		if s.config.UseTLS {
			// Start with TLS
			err = s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		} else {
			// Start without TLS
			err = s.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	// Close security manager
	if err := s.securityManager.Close(); err != nil {
		return fmt.Errorf("failed to close security manager: %w", err)
	}

	log.Println("Server exited properly")
	return nil

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	// Register scan routes with middleware
	s.scanHdlr.RegisterRoutes(s.mux, s.withMiddleware)

	// Add middleware to index route
	s.mux.Handle("/", s.withMiddleware(http.HandlerFunc(s.handleIndex)))

	// Register security routes
	s.registerSecurityRoutes()

// handleIndex handles the root path
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","message":"LLMrecon API","version":"1.0.0"}`)

// registerSecurityRoutes registers security-related routes
func (s *Server) registerSecurityRoutes() {
	// Health check endpoint (exempt from rate limiting and IP allowlisting)
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Metrics endpoint (protected)
	s.mux.Handle("/metrics", s.withMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Get anomalies
		anomalies := s.securityManager.GetAnomalyDetector().GetAnomalies()

		// Get rate limiting statistics
		rateLimitStats := s.securityManager.GetRateLimiter().GetStatistics()

		// Create response
		response := map[string]interface{}{
			"anomalies": anomalies,
			"rate_limit_stats": rateLimitStats,
			"timestamp": time.Now(),
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})))

// generateRequestID is implemented in middleware.go

// withMiddleware applies common middleware to handlers
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	// Chain middleware in order
	handler := next

	// Apply prompt protection middleware (outermost, executed last)
	handler = s.promptProtectionMiddleware.Middleware(handler)

	// Apply rate limiting middleware
	handler = s.securityManager.GetRateLimiter().Middleware(handler)

	// Apply IP allowlist middleware
	handler = s.securityManager.GetIPAllowlist().Middleware(handler)

	// Apply secure logging middleware
	handler = s.securityManager.GetSecureLogger().Middleware(handler)

	// Apply anomaly detection middleware
	handler = s.securityManager.GetAnomalyDetector().Middleware(handler)

	// Apply error handling wrapper
	errorHandler := s.securityManager.GetErrorHandler()
	// Wrap the handler with error handling
	handlerWithErrorHandling := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set up recovery to handle panics
		defer func() {
			if err := recover(); err != nil {
				// Get request ID from context
				requestID := r.Context().Value("request_id").(string)
				// Log the panic
				s.securityManager.Log(3, requestID, "Panic recovered", fmt.Errorf("%v", err))
				// Return a 500 error
			}
		}()
		
		// Call the next handler
		handler.ServeHTTP(w, r)
	})
	
	handler = handlerWithErrorHandling

	// Apply request ID and client IP middleware (innermost, executed first)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a request ID
		requestID := generateRequestID()

		// Add request ID to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add client IP to context
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
			if i := strings.LastIndex(clientIP, ":"); i != -1 {
				clientIP = clientIP[:i]
			}
		}
		ctx = context.WithValue(ctx, "client_ip", clientIP)
		r = r.WithContext(ctx)

		// Call the next handler
		handler.ServeHTTP(w, r)
	})
}
}
}
}
}
}
