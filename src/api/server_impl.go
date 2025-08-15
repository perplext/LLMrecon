package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// ServerImpl represents the API server implementation
type ServerImpl struct {
	config     *Config
	router     *mux.Router
	httpServer *http.Server
	scanStore  ScanStore
	services   *Services
	wg         sync.WaitGroup
	shutdown   chan struct{}
}

// NewServerImpl creates a new API server implementation
func NewServerImpl(config *Config, services *Services) (*ServerImpl, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	s := &ServerImpl{
		config:    config,
		services:  services,
		scanStore: NewInMemoryScanStore(),
		shutdown:  make(chan struct{}),
	}
	
	// Create router
	s.router = NewRouter(config)
	
	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	// Configure TLS if enabled
	if config.TLSCert != "" && config.TLSKey != "" {
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{ // #nosec G402 - These are secure cipher suites
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
		s.httpServer.TLSConfig = tlsConfig
	}
	
	// Set server instance for handlers
	serverInstance = s
	
	return s, nil

// Start starts the API server
func (s *ServerImpl) Start() error {
	// Setup graceful shutdown
	s.setupGracefulShutdown()
	
	// Start background workers
	s.startBackgroundWorkers()
	
	// Log startup message
	log.Info().
		Str("host", s.config.Host).
		Int("port", s.config.Port).
		Bool("tls", s.config.TLSCert != "").
		Bool("auth", s.config.EnableAuth).
		Msg("Starting API server")
	
	// Start server
	var err error
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		err = s.httpServer.ListenAndServeTLS(s.config.TLSCert, s.config.TLSKey)
	} else {
		err = s.httpServer.ListenAndServe()
	}
	
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}
	
	return nil

// Stop gracefully stops the API server
func (s *ServerImpl) Stop(timeout time.Duration) error {
	log.Info().Msg("Stopping API server")
	
	// Signal shutdown
	close(s.shutdown)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown HTTP server gracefully")
		return err
	}
	
	// Wait for background workers
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Info().Msg("All background workers stopped")
	case <-ctx.Done():
		log.Warn().Msg("Timeout waiting for background workers")
	}
	
	log.Info().Msg("API server stopped")
	return nil

// setupGracefulShutdown sets up signal handling for graceful shutdown
func (s *ServerImpl) setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Info().Msg("Received shutdown signal")
		
		// Give server 30 seconds to shutdown gracefully
		if err := s.Stop(30 * time.Second); err != nil {
			log.Error().Err(err).Msg("Failed to stop server gracefully")
			os.Exit(1)
		}
		os.Exit(0)
	}()

// startBackgroundWorkers starts background tasks
func (s *ServerImpl) startBackgroundWorkers() {
	// Scan cleanup worker
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.scanCleanupWorker()
	}()
	
	// Metrics collection worker (if needed)
	if s.config.EnableMetrics {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.metricsWorker()
		}()
	}

// scanCleanupWorker periodically cleans up old scans
func (s *ServerImpl) scanCleanupWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := s.scanStore.CleanupOldScans(24 * time.Hour); err != nil {
				log.Error().Err(err).Msg("Failed to cleanup old scans")
			}
		case <-s.shutdown:
			return
		}
	}

// metricsWorker collects and reports metrics
func (s *ServerImpl) metricsWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Collect metrics (implement as needed)
			log.Debug().Msg("Collecting metrics")
		case <-s.shutdown:
			return
		}
	}

// RunServer starts the API server with the given configuration
func RunServer(config *Config, services *Services) error {
	server, err := NewServerImpl(config, services)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	
}
}
}
}
}
}
}
