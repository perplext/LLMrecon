// Package communication provides secure communication utilities for the LLMrecon tool.
package communication

import (
	"time"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ValidateCertificate validates a certificate and returns certificate information
func (m *TrustChainManager) ValidateCertificate(cert *x509.Certificate) (*CertificateInfo, error) {
	if cert == nil {
		return nil, errors.New("certificate cannot be nil")
	}

	m.mu.RLock()
	// Check cache first
	key := fmt.Sprintf("%s-%s", cert.Subject.String(), cert.SerialNumber.String())
	cachedInfo, found := m.certInfoCache[key]
	m.mu.RUnlock()

	// If found in cache and not expired, return cached info
	if found && time.Since(cachedInfo.LastChecked) < m.cacheExpiration {
		return cachedInfo, nil
	}

	// Build certificate info
	certInfo := &CertificateInfo{
		Certificate: cert,
		LastChecked: time.Now(),
	}

	// Build verification options
	opts := x509.VerifyOptions{
		Roots:         x509.NewCertPool(),
		Intermediates: x509.NewCertPool(),
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	// Add root certificates to the pool
	m.mu.RLock()
	for _, rootCert := range m.rootCerts {
		opts.Roots.AddCert(rootCert)
	}

	// Add intermediate certificates to the pool
	for _, intermediateCert := range m.intermediateCerts {
		opts.Intermediates.AddCert(intermediateCert)
	}
	m.mu.RUnlock()

	// Verify the certificate
	chains, err := cert.Verify(opts)
	if err != nil {
		certInfo.Status = CertStatusUntrusted
		certInfo.ValidationError = fmt.Errorf("certificate verification failed: %w", err)
		
		// Cache the result
		m.mu.Lock()
		m.certInfoCache[key] = certInfo
		m.mu.Unlock()
		
		return certInfo, certInfo.ValidationError
	}

	// Store the trust chain (use the first chain)
	if len(chains) > 0 {
		certInfo.TrustChain = chains[0]
	}

	// Check if the certificate is expired
	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		certInfo.Status = CertStatusExpired
		certInfo.ValidationError = fmt.Errorf("certificate is expired or not yet valid")
		
		// Cache the result
		m.mu.Lock()
		m.certInfoCache[key] = certInfo
		m.mu.Unlock()
		
		return certInfo, certInfo.ValidationError
	}

	// Check revocation status if enabled
	if m.crlCheckEnabled {
		isRevoked, err := m.checkCertificateRevocation(cert)
		if err != nil {
			// Don't fail validation for CRL check errors, but log them
			certInfo.ValidationError = fmt.Errorf("CRL check error: %w", err)
		} else if isRevoked {
			certInfo.Status = CertStatusRevoked
			certInfo.ValidationError = fmt.Errorf("certificate is revoked")
			
			// Cache the result
			m.mu.Lock()
			m.certInfoCache[key] = certInfo
			m.mu.Unlock()
			
			return certInfo, certInfo.ValidationError
		}
	}

	// If we got here, the certificate is valid
	certInfo.Status = CertStatusValid

	// Cache the result
	m.mu.Lock()
	m.certInfoCache[key] = certInfo
	m.mu.Unlock()

	return certInfo, nil

// ValidateCertificateFromPEM validates a certificate from PEM data
func (m *TrustChainManager) ValidateCertificateFromPEM(pemData []byte) (*CertificateInfo, error) {
	cert, err := parseCertificateFromPEM(pemData)
	if err != nil {
		return nil, err
	}

	return m.ValidateCertificate(cert)

// ValidateCertificateFromFile validates a certificate from a file
func (m *TrustChainManager) ValidateCertificateFromFile(filePath string) (*CertificateInfo, error) {
	pemData, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	return m.ValidateCertificateFromPEM(pemData)

// checkCertificateRevocation checks if a certificate is revoked
func (m *TrustChainManager) checkCertificateRevocation(cert *x509.Certificate) (bool, error) {
	// If no CRL distribution points, we can't check revocation
	if len(cert.CRLDistributionPoints) == 0 {
		return false, nil
	}

	// Check each CRL distribution point
	for _, crlDP := range cert.CRLDistributionPoints {
		// Check if we have a cached CRL
		m.mu.RLock()
		crlInfo, found := m.crlCache[crlDP]
		m.mu.RUnlock()

		// If found and not expired, use cached CRL
		if found && time.Now().Before(crlInfo.NextUpdate) {
			// Check if the certificate is in the CRL
			for _, revokedCert := range crlInfo.CRL.TBSCertList.RevokedCertificates {
				if revokedCert.SerialNumber.Cmp(cert.SerialNumber) == 0 {
					return true, nil
				}
			}
			continue
		}

		// Fetch the CRL
		crl, err := m.fetchCRL(crlDP)
		if err != nil {
			return false, fmt.Errorf("failed to fetch CRL from %s: %w", crlDP, err)
		}

		// Check if the certificate is in the CRL
		for _, revokedCert := range crl.TBSCertList.RevokedCertificates {
			if revokedCert.SerialNumber.Cmp(cert.SerialNumber) == 0 {
				return true, nil
			}
		}

		// Cache the CRL
		m.mu.Lock()
		m.crlCache[crlDP] = &CRLInfo{
			CRL:         crl,
			LastUpdated: time.Now(),
			NextUpdate:  crl.TBSCertList.NextUpdate,
			URL:         crlDP,
		}
		m.mu.Unlock()
	}

	// Certificate is not revoked
	return false, nil

// fetchCRL fetches a CRL from a URL
func (m *TrustChainManager) fetchCRL(url string) (*pkix.CertificateList, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the CRL
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CRL: %w", err)
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch CRL: status code %d", resp.StatusCode)
	}

	// Read response body
	crlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read CRL data: %w", err)
	}

	// Try to parse as DER
	crl, err := x509.ParseCRL(crlData)
	if err != nil {
		// Try to parse as PEM
		block, _ := pem.Decode(crlData)
		if block != nil && block.Type == "X509 CRL" {
			crl, err = x509.ParseCRL(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse CRL: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse CRL: %w", err)
		}
	}

	return crl, nil

// GetTrustedCertificates returns all trusted certificates (roots and intermediates)
func (m *TrustChainManager) GetTrustedCertificates() []*x509.Certificate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certs := make([]*x509.Certificate, 0, len(m.rootCerts)+len(m.intermediateCerts))
	
	for _, cert := range m.rootCerts {
		certs = append(certs, cert)
	}
	
	for _, cert := range m.intermediateCerts {
		certs = append(certs, cert)
	}
	
	return certs

// GetRootCertificates returns all root certificates
func (m *TrustChainManager) GetRootCertificates() []*x509.Certificate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certs := make([]*x509.Certificate, 0, len(m.rootCerts))
	
	for _, cert := range m.rootCerts {
		certs = append(certs, cert)
	}
	
	return certs

// GetIntermediateCertificates returns all intermediate certificates
func (m *TrustChainManager) GetIntermediateCertificates() []*x509.Certificate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certs := make([]*x509.Certificate, 0, len(m.intermediateCerts))
	
	for _, cert := range m.intermediateCerts {
		certs = append(certs, cert)
	}
	
	return certs

// ClearCRLCache clears the CRL cache
func (m *TrustChainManager) ClearCRLCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.crlCache = make(map[string]*CRLInfo)

// ClearCertificateCache clears the certificate validation cache
func (m *TrustChainManager) ClearCertificateCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

