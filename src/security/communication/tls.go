// Package communication provides secure communication utilities for the LLMrecon tool.
package communication

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// TLSConfig represents the configuration for TLS
type TLSConfig struct {
	// MinVersion is the minimum TLS version to use
	MinVersion uint16
	// CertFile is the path to the certificate file
	CertFile string
	// KeyFile is the path to the key file
	KeyFile string
	// CAFile is the path to the CA certificate file
	CAFile string
	// InsecureSkipVerify skips certificate verification (not recommended for production)
	InsecureSkipVerify bool
	// CertPinning enables certificate pinning
	CertPinning bool
	// PinnedCerts is a list of pinned certificate hashes
	PinnedCerts []string
	// ServerName is the server name to check against
	ServerName string
}

// DefaultTLSConfig returns the default TLS configuration
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
		CertPinning:        false,
	}
}

// TLSManager manages TLS configurations and clients
type TLSManager struct {
	configs     map[string]*TLSConfig
	clients     map[string]*http.Client
	certPool    *x509.CertPool
	mu          sync.RWMutex
	defaultName string
}

// NewTLSManager creates a new TLS manager
func NewTLSManager() *TLSManager {
	return &TLSManager{
		configs:  make(map[string]*TLSConfig),
		clients:  make(map[string]*http.Client),
		certPool: x509.NewCertPool(),
	}
}

// AddConfig adds a TLS configuration
func (m *TLSManager) AddConfig(name string, config *TLSConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create TLS configuration
	tlsConfig := &tls.Config{
		MinVersion:         config.MinVersion,
		InsecureSkipVerify: config.InsecureSkipVerify,
		ServerName:         config.ServerName,
	}

	// Load certificates if provided
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load certificates: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if config.CAFile != "" {
		caCert, err := ioutil.ReadFile(config.CAFile)
		if err != nil {
			return fmt.Errorf("failed to read CA certificate: %w", err)
		}
		if !m.certPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = m.certPool
	}

	// Create HTTP client with TLS configuration
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		// Set reasonable defaults for production
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Store configuration and client
	m.configs[name] = config
	m.clients[name] = client

	// Set as default if first configuration
	if m.defaultName == "" {
		m.defaultName = name
	}

	return nil
}

// GetClient returns an HTTP client with the specified TLS configuration
func (m *TLSManager) GetClient(name string) (*http.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[name]
	if !ok {
		return nil, fmt.Errorf("TLS configuration not found: %s", name)
	}

	return client, nil
}

// GetDefaultClient returns the default HTTP client
func (m *TLSManager) GetDefaultClient() (*http.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultName == "" {
		return nil, fmt.Errorf("no default TLS configuration")
	}

	return m.clients[m.defaultName], nil
}

// SetDefaultConfig sets the default TLS configuration
func (m *TLSManager) SetDefaultConfig(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.configs[name]; !ok {
		return fmt.Errorf("TLS configuration not found: %s", name)
	}

	m.defaultName = name
	return nil
}

// LoadSystemCertificates loads system certificates into the cert pool
func (m *TLSManager) LoadSystemCertificates() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	systemPool, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("failed to load system certificates: %w", err)
	}

	m.certPool = systemPool
	return nil
}

// VerifyCertificate verifies a certificate against the cert pool
func (m *TLSManager) VerifyCertificate(cert *x509.Certificate) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	opts := x509.VerifyOptions{
		Roots:         m.certPool,
		CurrentTime:   time.Now(),
		Intermediates: x509.NewCertPool(),
	}

	_, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// ConfigureTLSForServer configures TLS for an HTTP server
func ConfigureTLSForServer(config *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: config.MinVersion,
	}

	// Load certificates if provided
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificates: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else {
		return nil, fmt.Errorf("certificate and key files are required for server TLS")
	}

	// Load CA certificate if provided
	if config.CAFile != "" {
		caCert, err := ioutil.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.ClientCAs = certPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}
