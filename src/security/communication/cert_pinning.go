// Package communication provides secure communication utilities for the LLMrecon tool.
package communication

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// CertificatePinner implements certificate pinning for HTTP clients
type CertificatePinner struct {
	// pins maps hostnames to sets of pinned certificate hashes
	pins     map[string]map[string]bool
	mu       sync.RWMutex
	enforced bool
}

// NewCertificatePinner creates a new certificate pinner
func NewCertificatePinner(enforced bool) *CertificatePinner {
	return &CertificatePinner{
		pins:     make(map[string]map[string]bool),
		enforced: enforced,
	}
}

// AddPin adds a certificate pin for a host
// The pin should be a base64-encoded SHA-256 hash of the certificate's public key
func (p *CertificatePinner) AddPin(hostname string, pin string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Normalize hostname
	hostname = strings.ToLower(hostname)

	// Create pin set if it doesn't exist
	if _, ok := p.pins[hostname]; !ok {
		p.pins[hostname] = make(map[string]bool)
	}

	// Add pin
	p.pins[hostname][pin] = true
}

// AddPins adds multiple certificate pins for a host
func (p *CertificatePinner) AddPins(hostname string, pins []string) {
	for _, pin := range pins {
		p.AddPin(hostname, pin)
	}
}

// RemovePin removes a certificate pin for a host
func (p *CertificatePinner) RemovePin(hostname string, pin string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Normalize hostname
	hostname = strings.ToLower(hostname)

	// Check if pin set exists
	pinSet, ok := p.pins[hostname]
	if !ok {
		return
	}

	// Remove pin
	delete(pinSet, pin)

	// Remove pin set if empty
	if len(pinSet) == 0 {
		delete(p.pins, hostname)
	}
}

// ClearPins removes all pins for a host
func (p *CertificatePinner) ClearPins(hostname string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Normalize hostname
	hostname = strings.ToLower(hostname)

	// Remove pin set
	delete(p.pins, hostname)
}

// SetEnforced sets whether certificate pinning is enforced
func (p *CertificatePinner) SetEnforced(enforced bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enforced = enforced
}

// IsEnforced returns whether certificate pinning is enforced
func (p *CertificatePinner) IsEnforced() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.enforced
}

// VerifyCertificate verifies that a certificate matches a pinned hash for a host
func (p *CertificatePinner) VerifyCertificate(hostname string, cert *x509.Certificate) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Normalize hostname
	hostname = strings.ToLower(hostname)

	// Check if we have pins for this host
	pinSet, ok := p.pins[hostname]
	if !ok {
		// No pins for this host, so it's valid
		return nil
	}

	// Calculate the hash of the certificate's public key
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	b64Hash := base64.StdEncoding.EncodeToString(hash[:])

	// Check if the hash matches any of the pins
	if pinSet[b64Hash] {
		return nil
	}

	// If we're not enforcing pins, just log the error
	if !p.enforced {
		return fmt.Errorf("certificate pin verification failed for %s (not enforced)", hostname)
	}

	// Otherwise, return an error
	return fmt.Errorf("certificate pin verification failed for %s", hostname)
}

// WrapTransport wraps an HTTP transport with certificate pinning
func (p *CertificatePinner) WrapTransport(transport *http.Transport) *http.Transport {
	// Create a copy of the transport
	newTransport := transport.Clone()

	// Get the original verification function
	originalVerifyPeerCert := newTransport.TLSClientConfig.VerifyPeerCertificate

	// Set a new verification function that includes pin checking
	newTransport.TLSClientConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// Call the original verification function if it exists
		if originalVerifyPeerCert != nil {
			if err := originalVerifyPeerCert(rawCerts, verifiedChains); err != nil {
				return err
			}
		}

		// If there are no verified chains, we can't verify pins
		if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
			return fmt.Errorf("no verified certificate chain")
		}

		// Get the hostname from the TLS config
		hostname := newTransport.TLSClientConfig.ServerName

		// Verify the certificate against pins
		return p.VerifyCertificate(hostname, verifiedChains[0][0])
	}

	return newTransport
}

// CreatePinnedClient creates an HTTP client with certificate pinning
func (p *CertificatePinner) CreatePinnedClient(hostname string, pins []string) *http.Client {
	// Add pins
	p.AddPins(hostname, pins)

	// Create a transport with TLS configuration
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: hostname,
			MinVersion: tls.VersionTLS12,
		},
	}

	// Wrap the transport with certificate pinning
	pinnedTransport := p.WrapTransport(transport)

	// Create a client with the pinned transport
	return &http.Client{
		Transport: pinnedTransport,
	}
}

// ExtractCertificatePin extracts a certificate pin from a certificate
func ExtractCertificatePin(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// FetchCertificatePin fetches a certificate pin from a host
func FetchCertificatePin(hostname string, port int) (string, error) {
	// Connect to the host
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s:%d: %w", hostname, port, err)
	}
	defer conn.Close()

	// Get the peer certificates
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return "", fmt.Errorf("no certificates found for %s:%d", hostname, port)
	}

	// Extract the pin from the first certificate
	return ExtractCertificatePin(certs[0]), nil
}
