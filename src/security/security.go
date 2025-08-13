// Package security provides security utilities for the LLMrecon tool.
package security

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/perplext/LLMrecon/src/security/api"
	"github.com/perplext/LLMrecon/src/security/communication"
)

// SecurityConfig represents the configuration for the security manager
type SecurityConfig struct {
	// TLS configuration
	TLSConfig *communication.TLSConfig
	// Rate limiter configuration
	RateLimiterConfig *api.RateLimiterConfig
	// IP allowlist configuration
	IPAllowlistConfig *api.IPAllowlistConfig
	// Secure logger configuration
	SecureLoggerConfig *api.SecureLoggerConfig
	// Anomaly detector configuration
	AnomalyDetectorConfig *api.AnomalyDetectorConfig
	// Development mode
	DevelopmentMode bool
	// Log file path
	LogFilePath string
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		TLSConfig:             communication.DefaultTLSConfig(),
		RateLimiterConfig:     api.DefaultRateLimiterConfig(),
		IPAllowlistConfig:     api.DefaultIPAllowlistConfig(),
		SecureLoggerConfig:    api.DefaultSecureLoggerConfig(),
		AnomalyDetectorConfig: api.DefaultAnomalyDetectorConfig(),
		DevelopmentMode:       false,
		LogFilePath:           "logs/security.log",
	}
}

// SecurityManager manages security components
type SecurityManager struct {
	config          *SecurityConfig
	tlsManager      *communication.TLSManager
	rateLimiter     *api.RateLimiter
	ipAllowlist     *api.IPAllowlist
	secureLogger    *api.SecureLogger
	anomalyDetector *api.AnomalyDetector
	errorHandler    *communication.ErrorHandler
	certPinner      *communication.CertificatePinner
	logFile         *os.File
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config *SecurityConfig) (*SecurityManager, error) {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Create TLS manager
	tlsManager := communication.NewTLSManager()

	// Create rate limiter
	rateLimiter := api.NewRateLimiter(config.RateLimiterConfig)

	// Create IP allowlist
	ipAllowlist, err := api.NewIPAllowlist(config.IPAllowlistConfig)
	if err != nil {
		return nil, err
	}

	// Open log file if specified
	var logFile *os.File
	var logWriter io.Writer
	if config.LogFilePath != "" {
		// Create directory if it doesn't exist
		dir := config.LogFilePath[:len(config.LogFilePath)-len("/security.log")]
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		// Open log file
		logFile, err = os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		logWriter = logFile
	} else {
		logWriter = os.Stdout
	}

	// Update secure logger config with log writer
	config.SecureLoggerConfig.OutputWriter = logWriter

	// Create secure logger
	secureLogger, err := api.NewSecureLogger(config.SecureLoggerConfig)
	if err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return nil, err
	}

	// Create anomaly detector with alert callback
	config.AnomalyDetectorConfig.AlertCallback = func(anomaly *api.Anomaly) {
		// Log the anomaly
		secureLogger.Log(api.LogLevelWarning, anomaly.RequestID, "Anomaly detected: "+anomaly.Description, nil)
	}
	anomalyDetector := api.NewAnomalyDetector(config.AnomalyDetectorConfig)

	// Create error handler
	errorHandler := communication.NewErrorHandler(config.DevelopmentMode)

	// Create certificate pinner
	certPinner := communication.NewCertificatePinner(true)

	return &SecurityManager{
		config:          config,
		tlsManager:      tlsManager,
		rateLimiter:     rateLimiter,
		ipAllowlist:     ipAllowlist,
		secureLogger:    secureLogger,
		anomalyDetector: anomalyDetector,
		errorHandler:    errorHandler,
		certPinner:      certPinner,
		logFile:         logFile,
	}, nil
}

// Close closes the security manager and releases resources
func (sm *SecurityManager) Close() error {
	// Close anomaly detector
	sm.anomalyDetector.Close()

	// Close log file
	if sm.logFile != nil {
		return sm.logFile.Close()
	}

	return nil
}

// GetTLSManager returns the TLS manager
func (sm *SecurityManager) GetTLSManager() *communication.TLSManager {
	return sm.tlsManager
}

// GetRateLimiter returns the rate limiter
func (sm *SecurityManager) GetRateLimiter() *api.RateLimiter {
	return sm.rateLimiter
}

// GetIPAllowlist returns the IP allowlist
func (sm *SecurityManager) GetIPAllowlist() *api.IPAllowlist {
	return sm.ipAllowlist
}

// GetSecureLogger returns the secure logger
func (sm *SecurityManager) GetSecureLogger() *api.SecureLogger {
	return sm.secureLogger
}

// GetAnomalyDetector returns the anomaly detector
func (sm *SecurityManager) GetAnomalyDetector() *api.AnomalyDetector {
	return sm.anomalyDetector
}

// GetErrorHandler returns the error handler
func (sm *SecurityManager) GetErrorHandler() *communication.ErrorHandler {
	return sm.errorHandler
}

// GetCertificatePinner returns the certificate pinner
func (sm *SecurityManager) GetCertificatePinner() *communication.CertificatePinner {
	return sm.certPinner
}

// ApplyMiddleware applies all security middleware to a handler
func (sm *SecurityManager) ApplyMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in the correct order
	// 1. Secure logging (to log all requests)
	handler = sm.secureLogger.Middleware(handler)
	// 2. IP allowlist (to block disallowed IPs)
	handler = sm.ipAllowlist.Middleware(handler)
	// 3. Rate limiter (to limit request rate)
	handler = sm.rateLimiter.Middleware(handler)
	// 4. Anomaly detection (to detect unusual patterns)
	handler = sm.anomalyDetector.Middleware(handler)

	return handler
}

// CreateSecureClient creates a secure HTTP client
func (sm *SecurityManager) CreateSecureClient(name string, config *communication.TLSConfig) (*http.Client, error) {
	// Add TLS configuration
	if err := sm.tlsManager.AddConfig(name, config); err != nil {
		return nil, err
	}

	// Get client
	return sm.tlsManager.GetClient(name)
}

// CreatePinnedClient creates a client with certificate pinning
func (sm *SecurityManager) CreatePinnedClient(hostname string, pins []string) *http.Client {
	return sm.certPinner.CreatePinnedClient(hostname, pins)
}

// HandleError handles an error securely
func (sm *SecurityManager) HandleError(w http.ResponseWriter, r *http.Request, err error, defaultMessage string) {
	sm.errorHandler.HandleError(w, r, err, defaultMessage)
}

// NewSecureError creates a new secure error
func (sm *SecurityManager) NewSecureError(code string, message string, level communication.ErrorLevel, originalError error) *communication.SecureError {
	return communication.NewSecureError(code, message, level, originalError)
}

// Log logs a message
func (sm *SecurityManager) Log(level api.LogLevel, requestID string, message string, err error) {
	sm.secureLogger.Log(level, requestID, message, err)
}

// ConfigureTLSForServer configures TLS for an HTTP server
func (sm *SecurityManager) ConfigureTLSForServer() (*http.Server, error) {
	// Configure TLS
	tlsConfig, err := communication.ConfigureTLSForServer(sm.config.TLSConfig)
	if err != nil {
		return nil, err
	}

	// Create server
	server := &http.Server{
		Addr:         ":8443", // Default HTTPS port
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, nil
}
