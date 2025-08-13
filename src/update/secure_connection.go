// Package update provides functionality for checking and applying updates
package update

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// PinnedCertificate represents a pinned certificate for a specific host
type PinnedCertificate struct {
	Host            string   // Hostname
	PublicKeyHashes []string // SHA-256 hashes of the public key in base64
	SubjectName     string   // Expected subject name
	Issuer          string   // Expected issuer
}

// ConnectionSecurityOptions contains security options for connections
type ConnectionSecurityOptions struct {
	// Whether to enable certificate pinning
	EnableCertificatePinning bool
	// Pinned certificates by host
	PinnedCertificates []PinnedCertificate
	// Whether to check certificate revocation
	CheckRevocation bool
	// Minimum TLS version
	MinTLSVersion uint16
	// Cipher suites (nil means use defaults)
	CipherSuites []uint16
	// Whether to enable HTTP/2
	EnableHTTP2 bool
	// Timeout for connections
	ConnectionTimeout time.Duration
	// Timeout for TLS handshake
	TLSHandshakeTimeout time.Duration
	// Timeout for idle connections
	IdleConnectionTimeout time.Duration
	// Maximum number of idle connections
	MaxIdleConnections int
	// Maximum number of idle connections per host
	MaxIdleConnectionsPerHost int
	// Retry configuration
	RetryConfig RetryConfig
}

// RetryConfig contains configuration for retry behavior
type RetryConfig struct {
	// Maximum number of retry attempts
	MaxRetries int
	// Initial delay before first retry
	InitialDelay time.Duration
	// Maximum delay between retries
	MaxDelay time.Duration
	// Whether to use exponential backoff
	UseExponentialBackoff bool
	// Jitter factor (0.0-1.0) to add randomness to retry delays
	JitterFactor float64
	// Status codes that should trigger a retry
	RetryableStatusCodes []int
	// Network errors that should trigger a retry
	RetryableNetworkErrors []string
}

// DefaultConnectionSecurityOptions returns the default security options
func DefaultConnectionSecurityOptions() *ConnectionSecurityOptions {
	return &ConnectionSecurityOptions{
		EnableCertificatePinning:  false,
		PinnedCertificates:        []PinnedCertificate{},
		CheckRevocation:           true,
		MinTLSVersion:             tls.VersionTLS13,
		CipherSuites:              nil, // Use defaults
		EnableHTTP2:               true,
		ConnectionTimeout:         30 * time.Second,
		TLSHandshakeTimeout:       10 * time.Second,
		IdleConnectionTimeout:     90 * time.Second,
		MaxIdleConnections:        100,
		MaxIdleConnectionsPerHost: 10,
		RetryConfig: RetryConfig{
			MaxRetries:            3,
			InitialDelay:          1 * time.Second,
			MaxDelay:              30 * time.Second,
			UseExponentialBackoff: true,
			JitterFactor:          0.1,
			RetryableStatusCodes: []int{
				http.StatusRequestTimeout,
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
			},
			RetryableNetworkErrors: []string{
				"connection refused",
				"connection reset",
				"connection closed",
				"no such host",
				"timeout",
				"temporary failure",
			},
		},
	}
}

// SecureClient provides a secure HTTP client with enhanced security features
type SecureClient struct {
	client  *http.Client
	options *ConnectionSecurityOptions
	mutex   sync.RWMutex
}

// NewSecureClient creates a new SecureClient with the specified options
func NewSecureClient(options *ConnectionSecurityOptions) (*SecureClient, error) {
	if options == nil {
		options = DefaultConnectionSecurityOptions()
	}

	// Create root CA pool
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		// If we can't get the system cert pool, create a new one
		rootCAs = x509.NewCertPool()
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		MinVersion: options.MinTLSVersion,
		RootCAs:    rootCAs,
	}

	// Set cipher suites if specified
	if options.CipherSuites != nil {
		tlsConfig.CipherSuites = options.CipherSuites
	}

	// Set verification function if certificate pinning is enabled
	if options.EnableCertificatePinning {
		tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
			return verifyPinnedCertificate(state, options.PinnedCertificates)
		}
	}

	// Create transport
	transport := &http.Transport{
		TLSClientConfig:       tlsConfig,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           createDialContext(options.ConnectionTimeout),
		TLSHandshakeTimeout:   options.TLSHandshakeTimeout,
		IdleConnTimeout:       options.IdleConnectionTimeout,
		MaxIdleConns:          options.MaxIdleConnections,
		MaxIdleConnsPerHost:   options.MaxIdleConnectionsPerHost,
		MaxConnsPerHost:       0, // No limit
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     options.EnableHTTP2,
		DisableCompression:    false,
		ResponseHeaderTimeout: options.ConnectionTimeout,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Create client
	client := &http.Client{
		Transport: transport,
		Timeout:   options.ConnectionTimeout * 2, // Overall timeout is twice the connection timeout
	}

	return &SecureClient{
		client:  client,
		options: options,
		mutex:   sync.RWMutex{},
	}, nil
}

// createDialContext creates a custom dial context function with timeout
func createDialContext(timeout time.Duration) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}
		return dialer.DialContext(ctx, network, addr)
	}
}

// verifyPinnedCertificate verifies that the certificate chain matches the pinned certificates
func verifyPinnedCertificate(state tls.ConnectionState, pinnedCerts []PinnedCertificate) error {
	// Find pinned certificate for this host
	var pinnedCert *PinnedCertificate
	for i, cert := range pinnedCerts {
		if cert.Host == state.ServerName {
			pinnedCert = &pinnedCerts[i]
			break
		}
	}

	// If no pinned certificate for this host, accept the connection
	if pinnedCert == nil {
		return nil
	}

	// Check peer certificates
	if len(state.PeerCertificates) == 0 {
		return errors.New("no peer certificates presented")
	}

	// Check subject name if specified
	if pinnedCert.SubjectName != "" {
		if state.PeerCertificates[0].Subject.CommonName != pinnedCert.SubjectName {
			return fmt.Errorf("certificate subject mismatch: expected %s, got %s",
				pinnedCert.SubjectName, state.PeerCertificates[0].Subject.CommonName)
		}
	}

	// Check issuer if specified
	if pinnedCert.Issuer != "" {
		if state.PeerCertificates[0].Issuer.CommonName != pinnedCert.Issuer {
			return fmt.Errorf("certificate issuer mismatch: expected %s, got %s",
				pinnedCert.Issuer, state.PeerCertificates[0].Issuer.CommonName)
		}
	}

	// Check public key hash
	cert := state.PeerCertificates[0]
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])

	// Check if the hash matches any of the pinned hashes
	for _, pinnedHash := range pinnedCert.PublicKeyHashes {
		if hashBase64 == pinnedHash {
			return nil // Match found
		}
	}

	return fmt.Errorf("certificate public key hash mismatch for %s", state.ServerName)
}

// Do performs an HTTP request with retry logic
func (c *SecureClient) Do(req *http.Request) (*http.Response, error) {
	c.mutex.RLock()
	retryConfig := c.options.RetryConfig
	c.mutex.RUnlock()

	var resp *http.Response
	var err error
	var delay time.Duration = 0

	// Create a context with timeout if not already set
	ctx := req.Context()
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.client.Timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	// Try the request with retries
	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		// Wait before retry (except for first attempt)
		if attempt > 0 {
			// Apply jitter to delay
			jitter := 1.0
			if retryConfig.JitterFactor > 0 {
				jitter = 1.0 + (retryConfig.JitterFactor * (2.0*float64(attempt)/float64(retryConfig.MaxRetries) - 1.0))
			}

			// Calculate delay with exponential backoff if enabled
			if retryConfig.UseExponentialBackoff {
				backoffFactor := 1 << uint(attempt-1)
				delay = time.Duration(float64(retryConfig.InitialDelay) * jitter * float64(backoffFactor))
			} else {
				delay = time.Duration(float64(retryConfig.InitialDelay) * jitter)
			}

			// Cap delay at max delay
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}

			// Wait for delay or context cancellation
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			// Clone the request body if needed for retry
			if req.Body != nil {
				// If the body is not nil, we need to clone it for retry
				// This is typically handled by the caller by providing a GetBody function
				if req.GetBody == nil {
					return nil, fmt.Errorf("request body cannot be reused for retry (GetBody not set)")
				}

				body, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to get request body for retry: %w", err)
				}
				req.Body = body
			}
		}

		// Perform the request
		resp, err = c.client.Do(req)

		// Check if we should retry based on error
		if err != nil {
			// Check if context was canceled or deadline exceeded
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			// Check if error is retryable
			if !isRetryableError(err, retryConfig.RetryableNetworkErrors) {
				return nil, err
			}

			// Retryable error, continue to next attempt
			continue
		}

		// Check if we should retry based on status code
		if isRetryableStatusCode(resp.StatusCode, retryConfig.RetryableStatusCodes) {
			resp.Body.Close() // Close the body before retry
			continue
		}

		// Success
		return resp, nil
	}

	// If we get here, all retries failed
	if resp != nil {
		resp.Body.Close()
	}

	if err != nil {
		return nil, fmt.Errorf("all retry attempts failed: %w", err)
	}

	return nil, fmt.Errorf("all retry attempts failed with status code: %d", resp.StatusCode)
}

// isRetryableError checks if an error is retryable based on its message
func isRetryableError(err error, retryableErrors []string) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	for _, retryableErr := range retryableErrors {
		if retryableErr != "" && containsIgnoreCase(errMsg, retryableErr) {
			return true
		}
	}

	return false
}

// isRetryableStatusCode checks if a status code is retryable
func isRetryableStatusCode(statusCode int, retryableStatusCodes []int) bool {
	for _, code := range retryableStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks if a string contains a substring, ignoring case
func containsIgnoreCase(s, substr string) bool {
	s, substr = s, substr
	for i := 0; i+len(substr) <= len(s); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalIgnoreCase checks if two strings are equal, ignoring case
func equalIgnoreCase(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}

// AddPinnedCertificate adds a pinned certificate to the client
func (c *SecureClient) AddPinnedCertificate(pinnedCert PinnedCertificate) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Enable certificate pinning if not already enabled
	c.options.EnableCertificatePinning = true

	// Check if we already have a pinned certificate for this host
	for i, cert := range c.options.PinnedCertificates {
		if cert.Host == pinnedCert.Host {
			// Update existing pinned certificate
			c.options.PinnedCertificates[i] = pinnedCert
			return
		}
	}

	// Add new pinned certificate
	c.options.PinnedCertificates = append(c.options.PinnedCertificates, pinnedCert)
}

// RemovePinnedCertificate removes a pinned certificate for a host
func (c *SecureClient) RemovePinnedCertificate(host string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Find and remove pinned certificate
	for i, cert := range c.options.PinnedCertificates {
		if cert.Host == host {
			// Remove by replacing with last element and truncating
			lastIndex := len(c.options.PinnedCertificates) - 1
			c.options.PinnedCertificates[i] = c.options.PinnedCertificates[lastIndex]
			c.options.PinnedCertificates = c.options.PinnedCertificates[:lastIndex]
			break
		}
	}

	// Disable certificate pinning if no pinned certificates remain
	if len(c.options.PinnedCertificates) == 0 {
		c.options.EnableCertificatePinning = false
	}
}

// GetPinnedCertificates returns a copy of the pinned certificates
func (c *SecureClient) GetPinnedCertificates() []PinnedCertificate {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Create a copy to avoid race conditions
	pinnedCerts := make([]PinnedCertificate, len(c.options.PinnedCertificates))
	copy(pinnedCerts, c.options.PinnedCertificates)

	return pinnedCerts
}

// UpdateRetryConfig updates the retry configuration
func (c *SecureClient) UpdateRetryConfig(retryConfig RetryConfig) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.options.RetryConfig = retryConfig
}

// GetRetryConfig returns a copy of the retry configuration
func (c *SecureClient) GetRetryConfig() RetryConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.options.RetryConfig
}

// Get performs a GET request with the secure client
func (c *SecureClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.Do(req)
}

// Head performs a HEAD request with the secure client
func (c *SecureClient) Head(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.Do(req)
}
