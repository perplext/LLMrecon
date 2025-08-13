// Package communication provides secure communication utilities for the LLMrecon tool.
package communication

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

// CertificateStatus represents the status of a certificate
type CertificateStatus int

const (
	// CertStatusValid indicates a valid certificate
	CertStatusValid CertificateStatus = iota
	// CertStatusExpired indicates an expired certificate
	CertStatusExpired
	// CertStatusRevoked indicates a revoked certificate
	CertStatusRevoked
	// CertStatusUntrusted indicates an untrusted certificate
	CertStatusUntrusted
	// CertStatusInvalid indicates an invalid certificate
	CertStatusInvalid
)

// CertificateInfo contains information about a certificate
type CertificateInfo struct {
	Certificate     *x509.Certificate
	Status          CertificateStatus
	TrustChain      []*x509.Certificate
	ValidationError error
	LastChecked     time.Time
}

// CRLInfo contains information about a Certificate Revocation List
type CRLInfo struct {
	CRL         *pkix.CertificateList
	LastUpdated time.Time
	NextUpdate  time.Time
	URL         string
}

// TrustChainManager manages certificate trust chains and validation
type TrustChainManager struct {
	// Root certificates (trusted anchors)
	rootCerts map[string]*x509.Certificate
	// Intermediate certificates
	intermediateCerts map[string]*x509.Certificate
	// Certificate info cache
	certInfoCache map[string]*CertificateInfo
	// CRL cache
	crlCache map[string]*CRLInfo
	// CRL check enabled
	crlCheckEnabled bool
	// OCSP check enabled
	ocspCheckEnabled bool
	// Cache expiration duration
	cacheExpiration time.Duration
	// Mutex for thread safety
	mu sync.RWMutex
}

// NewTrustChainManager creates a new trust chain manager
func NewTrustChainManager() *TrustChainManager {
	return &TrustChainManager{
		rootCerts:         make(map[string]*x509.Certificate),
		intermediateCerts: make(map[string]*x509.Certificate),
		certInfoCache:     make(map[string]*CertificateInfo),
		crlCache:          make(map[string]*CRLInfo),
		crlCheckEnabled:   true,
		ocspCheckEnabled:  true,
		cacheExpiration:   24 * time.Hour,
	}
}

// AddRootCertificate adds a root certificate to the trust store
func (m *TrustChainManager) AddRootCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("certificate cannot be nil")
	}

	// Check if it's a CA certificate
	if !cert.IsCA {
		return errors.New("certificate is not a CA certificate")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Use subject and serial as the key
	key := fmt.Sprintf("%s-%s", cert.Subject.String(), cert.SerialNumber.String())
	m.rootCerts[key] = cert

	// Clear cache
	m.certInfoCache = make(map[string]*CertificateInfo)

	return nil
}

// AddRootCertificateFromPEM adds a root certificate from PEM data
func (m *TrustChainManager) AddRootCertificateFromPEM(pemData []byte) error {
	cert, err := parseCertificateFromPEM(pemData)
	if err != nil {
		return err
	}

	return m.AddRootCertificate(cert)
}

// AddRootCertificateFromFile adds a root certificate from a file
func (m *TrustChainManager) AddRootCertificateFromFile(filePath string) error {
	pemData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	return m.AddRootCertificateFromPEM(pemData)
}

// AddIntermediateCertificate adds an intermediate certificate to the trust store
func (m *TrustChainManager) AddIntermediateCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("certificate cannot be nil")
	}

	// Check if it's a CA certificate
	if !cert.IsCA {
		return errors.New("certificate is not a CA certificate")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Use subject and serial as the key
	key := fmt.Sprintf("%s-%s", cert.Subject.String(), cert.SerialNumber.String())
	m.intermediateCerts[key] = cert

	// Clear cache
	m.certInfoCache = make(map[string]*CertificateInfo)

	return nil
}

// AddIntermediateCertificateFromPEM adds an intermediate certificate from PEM data
func (m *TrustChainManager) AddIntermediateCertificateFromPEM(pemData []byte) error {
	cert, err := parseCertificateFromPEM(pemData)
	if err != nil {
		return err
	}

	return m.AddIntermediateCertificate(cert)
}

// AddIntermediateCertificateFromFile adds an intermediate certificate from a file
func (m *TrustChainManager) AddIntermediateCertificateFromFile(filePath string) error {
	pemData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	return m.AddIntermediateCertificateFromPEM(pemData)
}

// RemoveCertificate removes a certificate from the trust store
func (m *TrustChainManager) RemoveCertificate(cert *x509.Certificate) {
	if cert == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Use subject and serial as the key
	key := fmt.Sprintf("%s-%s", cert.Subject.String(), cert.SerialNumber.String())
	delete(m.rootCerts, key)
	delete(m.intermediateCerts, key)

	// Clear cache
	m.certInfoCache = make(map[string]*CertificateInfo)
}

// SetCRLCheckEnabled sets whether CRL checking is enabled
func (m *TrustChainManager) SetCRLCheckEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.crlCheckEnabled = enabled
}

// SetOCSPCheckEnabled sets whether OCSP checking is enabled
func (m *TrustChainManager) SetOCSPCheckEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ocspCheckEnabled = enabled
}

// SetCacheExpiration sets the cache expiration duration
func (m *TrustChainManager) SetCacheExpiration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheExpiration = duration
}

// parseCertificateFromPEM parses a certificate from PEM data
func parseCertificateFromPEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}
