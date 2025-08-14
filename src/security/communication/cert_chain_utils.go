// Package communication provides secure communication utilities for the LLMrecon tool.
package communication

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CertificateFormat represents the format of a certificate
type CertificateFormat int

const (
	// CertFormatPEM represents PEM format
	CertFormatPEM CertificateFormat = iota
	// CertFormatDER represents DER format
	CertFormatDER
)

// ExportCertificate exports a certificate to a file
func ExportCertificate(cert *x509.Certificate, filePath string, format CertificateFormat) error {
	if cert == nil {
		return errors.New("certificate cannot be nil")
	}

	var data []byte
	switch format {
	case CertFormatPEM:
		// Convert to PEM
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		data = pem.EncodeToMemory(block)
	case CertFormatDER:
		data = cert.Raw
	default:
		return fmt.Errorf("unsupported certificate format: %d", format)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write certificate to file: %w", err)
	}

	return nil
}

// ImportCertificatesFromDirectory imports certificates from a directory
func (m *TrustChainManager) ImportCertificatesFromDirectory(dirPath string, isRoot bool) error {
	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dirPath)
	}

	// Read directory
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Process each file
	for _, file := range files {
		// Skip directories and hidden files
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".pem" && ext != ".crt" && ext != ".cer" {
			continue
		}

		// Import certificate
		filePath := filepath.Join(dirPath, file.Name())
		var importErr error
		if isRoot {
			importErr = m.AddRootCertificateFromFile(filePath)
		} else {
			importErr = m.AddIntermediateCertificateFromFile(filePath)
		}

		if importErr != nil {
			// Log error but continue with other files
			fmt.Printf("Failed to import certificate from %s: %v\n", filePath, importErr)
		}
	}

	return nil
}

// CreateCertificatePool creates an x509.CertPool from the trusted certificates
func (m *TrustChainManager) CreateCertificatePool() *x509.CertPool {
	pool := x509.NewCertPool()

	// Add root certificates
	for _, cert := range m.GetRootCertificates() {
		pool.AddCert(cert)
	}

	// Add intermediate certificates
	for _, cert := range m.GetIntermediateCertificates() {
		pool.AddCert(cert)
	}

	return pool
}

// CreateTLSConfig creates a tls.Config with the trusted certificates
func (m *TrustChainManager) CreateTLSConfig() *tls.Config {
	return &tls.Config{
		RootCAs:            m.CreateCertificatePool(),
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}
}

// CreateHTTPClient creates an HTTP client with the trusted certificates
func (m *TrustChainManager) CreateHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: m.CreateTLSConfig(),
		},
		Timeout: 30 * time.Second,
	}
}

// GetCertificateInfo returns information about a certificate
func GetCertificateInfo(cert *x509.Certificate) map[string]interface{} {
	if cert == nil {
		return nil
	}

	info := map[string]interface{}{
		"Subject":      cert.Subject.String(),
		"Issuer":       cert.Issuer.String(),
		"SerialNumber": cert.SerialNumber.String(),
		"NotBefore":    cert.NotBefore,
		"NotAfter":     cert.NotAfter,
		"IsCA":         cert.IsCA,
		"KeyUsage":     cert.KeyUsage,
	}

	// Add extended key usage
	extKeyUsage := make([]string, 0, len(cert.ExtKeyUsage))
	for _, usage := range cert.ExtKeyUsage {
		switch usage {
		case x509.ExtKeyUsageAny:
			extKeyUsage = append(extKeyUsage, "Any")
		case x509.ExtKeyUsageServerAuth:
			extKeyUsage = append(extKeyUsage, "ServerAuth")
		case x509.ExtKeyUsageClientAuth:
			extKeyUsage = append(extKeyUsage, "ClientAuth")
		case x509.ExtKeyUsageCodeSigning:
			extKeyUsage = append(extKeyUsage, "CodeSigning")
		case x509.ExtKeyUsageEmailProtection:
			extKeyUsage = append(extKeyUsage, "EmailProtection")
		case x509.ExtKeyUsageTimeStamping:
			extKeyUsage = append(extKeyUsage, "TimeStamping")
		case x509.ExtKeyUsageOCSPSigning:
			extKeyUsage = append(extKeyUsage, "OCSPSigning")
		default:
			extKeyUsage = append(extKeyUsage, fmt.Sprintf("Unknown(%d)", usage))
		}
	}
	info["ExtKeyUsage"] = extKeyUsage

	// Add CRL distribution points
	info["CRLDistributionPoints"] = cert.CRLDistributionPoints

	// Add OCSP server
	info["OCSPServer"] = cert.OCSPServer

	// Add issuing certificate URL
	info["IssuingCertificateURL"] = cert.IssuingCertificateURL

	// Add DNS names
	info["DNSNames"] = cert.DNSNames

	// Add IP addresses
	ips := make([]string, 0, len(cert.IPAddresses))
	for _, ip := range cert.IPAddresses {
		ips = append(ips, ip.String())
	}
	info["IPAddresses"] = ips

	return info
}

// FormatCertificateInfo formats certificate information as a string
func FormatCertificateInfo(cert *x509.Certificate) string {
	if cert == nil {
		return "Certificate: nil"
	}

	info := GetCertificateInfo(cert)
	var builder strings.Builder

	builder.WriteString("Certificate Information:\n")
	builder.WriteString(fmt.Sprintf("  Subject: %s\n", info["Subject"]))
	builder.WriteString(fmt.Sprintf("  Issuer: %s\n", info["Issuer"]))
	builder.WriteString(fmt.Sprintf("  Serial Number: %s\n", info["SerialNumber"]))
	builder.WriteString(fmt.Sprintf("  Valid From: %s\n", info["NotBefore"].(time.Time).Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("  Valid Until: %s\n", info["NotAfter"].(time.Time).Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("  Is CA: %t\n", info["IsCA"]))

	// Add extended key usage
	builder.WriteString("  Extended Key Usage:\n")
	for _, usage := range info["ExtKeyUsage"].([]string) {
		builder.WriteString(fmt.Sprintf("    - %s\n", usage))
	}

	// Add CRL distribution points
	builder.WriteString("  CRL Distribution Points:\n")
	for _, crlDP := range info["CRLDistributionPoints"].([]string) {
		builder.WriteString(fmt.Sprintf("    - %s\n", crlDP))
	}

	// Add OCSP server
	builder.WriteString("  OCSP Server:\n")
	for _, ocspServer := range info["OCSPServer"].([]string) {
		builder.WriteString(fmt.Sprintf("    - %s\n", ocspServer))
	}

	// Add issuing certificate URL
	builder.WriteString("  Issuing Certificate URL:\n")
	for _, issuingURL := range info["IssuingCertificateURL"].([]string) {
		builder.WriteString(fmt.Sprintf("    - %s\n", issuingURL))
	}

	// Add DNS names
	builder.WriteString("  DNS Names:\n")
	for _, dnsName := range info["DNSNames"].([]string) {
		builder.WriteString(fmt.Sprintf("    - %s\n", dnsName))
	}

	// Add IP addresses
	builder.WriteString("  IP Addresses:\n")
	for _, ip := range info["IPAddresses"].([]string) {
		builder.WriteString(fmt.Sprintf("    - %s\n", ip))
	}

	return builder.String()
}

// VerifyCertificateChain verifies a certificate chain
func VerifyCertificateChain(certs []*x509.Certificate, roots *x509.CertPool) error {
	if len(certs) == 0 {
		return errors.New("certificate chain is empty")
	}

	// Create intermediates pool
	intermediates := x509.NewCertPool()
	for i := 1; i < len(certs); i++ {
		intermediates.AddCert(certs[i])
	}

	// Create verification options
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	// Verify the leaf certificate
	_, err := certs[0].Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// GetCertificateChainFromPEM extracts a certificate chain from PEM data
func GetCertificateChainFromPEM(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	// Parse all certificates in the PEM data
	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}

		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, errors.New("no certificates found in PEM data")
	}

	return certs, nil
}

// GetCertificateChainFromFile extracts a certificate chain from a file
func GetCertificateChainFromFile(filePath string) ([]*x509.Certificate, error) {
	pemData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	return GetCertificateChainFromPEM(pemData)
}
