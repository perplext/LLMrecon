package update

import (
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Verifier handles cryptographic verification of updates
type Verifier struct {
	config     *Config
	logger     Logger
	trustedKeys map[string]*rsa.PublicKey
}

// NewVerifier creates a new verifier
func NewVerifier(config *Config, logger Logger) *Verifier {
	verifier := &Verifier{
		config:      config,
		logger:      logger,
		trustedKeys: make(map[string]*rsa.PublicKey),
	}
	
	// Load trusted keys
	verifier.loadTrustedKeys()
	
	return verifier
}

// FileVerificationResult represents the result of a file verification
type FileVerificationResult struct {
	Verified       bool
	ChecksumValid  bool
	SignatureValid bool
	Algorithm      string
	KeyID          string
	Error          error
}

// VerifyFile verifies a file's integrity and authenticity
func (v *Verifier) VerifyFile(filePath, expectedChecksum, signatureURL string) error {
	v.logger.Debug(fmt.Sprintf("Verifying file: %s", filePath))
	
	// Verify checksum if provided
	if expectedChecksum != "" {
		if err := v.verifyChecksum(filePath, expectedChecksum); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		v.logger.Debug("Checksum verification passed")
	}
	
	// Verify signature if URL provided and verification is enabled
	if signatureURL != "" && v.config.VerifySignatures {
		if err := v.verifySignature(filePath, signatureURL); err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
		v.logger.Debug("Signature verification passed")
	}
	
	return nil
}

// verifyChecksum verifies file checksum
func (v *Verifier) verifyChecksum(filePath, expectedChecksum string) error {
	// Determine hash algorithm based on checksum length
	var hasher hash.Hash
	var algorithm string
	
	expectedChecksum = strings.ToLower(expectedChecksum)
	checksumLength := len(expectedChecksum)
	
	switch checksumLength {
	case 32: // MD5
		hasher = md5.New()
		algorithm = "MD5"
	case 40: // SHA1
		hasher = sha1.New()
		algorithm = "SHA1"
	case 64: // SHA256
		hasher = sha256.New()
		algorithm = "SHA256"
	case 128: // SHA512
		hasher = sha512.New()
		algorithm = "SHA512"
	default:
		return fmt.Errorf("unsupported checksum format (length: %d)", checksumLength)
	}
	
	v.logger.Debug(fmt.Sprintf("Using %s algorithm for checksum verification", algorithm))
	
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Calculate hash
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}
	
	// Get computed checksum
	computedChecksum := hex.EncodeToString(hasher.Sum(nil))
	
	// Compare checksums
	if computedChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, computedChecksum)
	}
	
	return nil
}

// verifySignature verifies file signature
func (v *Verifier) verifySignature(filePath, signatureURL string) error {
	// Download signature
	signature, err := v.downloadSignature(signatureURL)
	if err != nil {
		return fmt.Errorf("failed to download signature: %w", err)
	}
	
	// Calculate file hash
	fileHash, err := v.calculateFileHash(filePath, crypto.SHA256)
	if err != nil {
		return fmt.Errorf("failed to calculate file hash: %w", err)
	}
	
	// Try to verify with each trusted key
	var lastError error
	for keyID, publicKey := range v.trustedKeys {
		v.logger.Debug(fmt.Sprintf("Trying verification with key: %s", keyID))
		
		err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, fileHash, signature)
		if err == nil {
			v.logger.Debug(fmt.Sprintf("Signature verified with key: %s", keyID))
			return nil
		}
		lastError = err
	}
	
	if len(v.trustedKeys) == 0 {
		return fmt.Errorf("no trusted keys loaded")
	}
	
	return fmt.Errorf("signature verification failed with all keys: %w", lastError)
}

// downloadSignature downloads a signature from URL
func (v *Verifier) downloadSignature(signatureURL string) ([]byte, error) {
	resp, err := http.Get(signatureURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download signature: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signature download failed: %d %s", resp.StatusCode, resp.Status)
	}
	
	signatureData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}
	
	// Try to decode as base64 if it's text
	if v.isTextData(signatureData) {
		if decoded, err := base64.StdEncoding.DecodeString(string(signatureData)); err == nil {
			return decoded, nil
		}
	}
	
	return signatureData, nil
}

// calculateFileHash calculates hash of a file
func (v *Verifier) calculateFileHash(filePath string, hashType crypto.Hash) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	var hasher hash.Hash
	switch hashType {
	case crypto.SHA256:
		hasher = sha256.New()
	case crypto.SHA512:
		hasher = sha512.New()
	case crypto.SHA1:
		hasher = sha1.New()
	default:
		return nil, fmt.Errorf("unsupported hash type: %v", hashType)
	}
	
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}
	
	return hasher.Sum(nil), nil
}

// loadTrustedKeys loads trusted public keys
func (v *Verifier) loadTrustedKeys() {
	// Load keys from configuration
	for _, keyData := range v.config.TrustedKeys {
		if key, keyID, err := v.parsePublicKey(keyData); err == nil {
			v.trustedKeys[keyID] = key
			v.logger.Debug(fmt.Sprintf("Loaded trusted key: %s", keyID))
		} else {
			v.logger.Error(fmt.Sprintf("Failed to parse trusted key: %s", keyData), err)
		}
	}
	
	// Load default embedded keys
	v.loadEmbeddedKeys()
}

// parsePublicKey parses a public key from various formats
func (v *Verifier) parsePublicKey(keyData string) (*rsa.PublicKey, string, error) {
	// Handle PEM format
	if strings.Contains(keyData, "BEGIN") {
		return v.parsePEMKey(keyData)
	}
	
	// Handle base64 encoded key
	if decoded, err := base64.StdEncoding.DecodeString(keyData); err == nil {
		if key, err := x509.ParsePKCS1PublicKey(decoded); err == nil {
			keyID := v.calculateKeyID(key)
			return key, keyID, nil
		}
	}
	
	// Handle hex encoded key
	if decoded, err := hex.DecodeString(keyData); err == nil {
		if key, err := x509.ParsePKCS1PublicKey(decoded); err == nil {
			keyID := v.calculateKeyID(key)
			return key, keyID, nil
		}
	}
	
	return nil, "", fmt.Errorf("unsupported key format")
}

// parsePEMKey parses a PEM formatted public key
func (v *Verifier) parsePEMKey(keyData string) (*rsa.PublicKey, string, error) {
	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return nil, "", fmt.Errorf("failed to decode PEM block")
	}
	
	var publicKey *rsa.PublicKey
	var err error
	
	switch block.Type {
	case "RSA PUBLIC KEY":
		publicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		pubKeyInterface, parseErr := x509.ParsePKIXPublicKey(block.Bytes)
		if parseErr != nil {
			return nil, "", parseErr
		}
		var ok bool
		publicKey, ok = pubKeyInterface.(*rsa.PublicKey)
		if !ok {
			return nil, "", fmt.Errorf("not an RSA public key")
		}
	default:
		return nil, "", fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
	
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse public key: %w", err)
	}
	
	keyID := v.calculateKeyID(publicKey)
	return publicKey, keyID, nil
}

// calculateKeyID calculates a unique ID for a public key
func (v *Verifier) calculateKeyID(key *rsa.PublicKey) string {
	// Serialize the key to DER format
	derBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "unknown"
	}
	
	// Calculate SHA256 hash
	hash := sha256.Sum256(derBytes)
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes as ID
}

// loadEmbeddedKeys loads default embedded keys
func (v *Verifier) loadEmbeddedKeys() {
	// These would be the official LLMrecon public keys
	// In a real implementation, these would be embedded at build time
	
	defaultKeys := []string{
		// Official LLMrecon signing key (example)
		`-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1234567890...
-----END RSA PUBLIC KEY-----`,
	}
	
	for _, keyData := range defaultKeys {
		if key, keyID, err := v.parsePublicKey(keyData); err == nil {
			if _, exists := v.trustedKeys[keyID]; !exists {
				v.trustedKeys[keyID] = key
				v.logger.Debug(fmt.Sprintf("Loaded embedded key: %s", keyID))
			}
		}
	}
}

// isTextData checks if data is likely text (for base64 detection)
func (v *Verifier) isTextData(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	
	// Check if all bytes are printable ASCII
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

// VerifyBundle verifies an entire update bundle
func (v *Verifier) VerifyBundle(bundlePath string) (*FileVerificationResult, error) {
	result := &FileVerificationResult{
		Algorithm: "SHA256",
	}
	
	// Check if bundle exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("bundle file not found: %s", bundlePath)
		return result, result.Error
	}
	
	// Look for checksum file
	checksumPath := bundlePath + ".sha256"
	if _, err := os.Stat(checksumPath); err == nil {
		checksum, err := os.ReadFile(checksumPath)
		if err != nil {
			result.Error = fmt.Errorf("failed to read checksum file: %w", err)
			return result, result.Error
		}
		
		checksumStr := strings.TrimSpace(string(checksum))
		if err := v.verifyChecksum(bundlePath, checksumStr); err != nil {
			result.Error = err
			return result, result.Error
		}
		result.ChecksumValid = true
	}
	
	// Look for signature file
	signaturePath := bundlePath + ".sig"
	if _, err := os.Stat(signaturePath); err == nil && v.config.VerifySignatures {
		if err := v.verifyBundleSignature(bundlePath, signaturePath); err != nil {
			result.Error = err
			return result, result.Error
		}
		result.SignatureValid = true
	}
	
	result.Verified = result.ChecksumValid || result.SignatureValid
	return result, nil
}

// verifyBundleSignature verifies a bundle signature from a local file
func (v *Verifier) verifyBundleSignature(bundlePath, signaturePath string) error {
	signature, err := os.ReadFile(signaturePath)
	if err != nil {
		return fmt.Errorf("failed to read signature file: %w", err)
	}
	
	// Try to decode as base64 if it's text
	if v.isTextData(signature) {
		if decoded, err := base64.StdEncoding.DecodeString(string(signature)); err == nil {
			signature = decoded
		}
	}
	
	// Calculate bundle hash
	fileHash, err := v.calculateFileHash(bundlePath, crypto.SHA256)
	if err != nil {
		return fmt.Errorf("failed to calculate bundle hash: %w", err)
	}
	
	// Try to verify with each trusted key
	var lastError error
	for keyID, publicKey := range v.trustedKeys {
		v.logger.Debug(fmt.Sprintf("Trying bundle verification with key: %s", keyID))
		
		err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, fileHash, signature)
		if err == nil {
			v.logger.Debug(fmt.Sprintf("Bundle signature verified with key: %s", keyID))
			return nil
		}
		lastError = err
	}
	
	if len(v.trustedKeys) == 0 {
		return fmt.Errorf("no trusted keys loaded")
	}
	
	return fmt.Errorf("bundle signature verification failed with all keys: %w", lastError)
}

// CreateChecksum creates a checksum file for a given file
func (v *Verifier) CreateChecksum(filePath string, algorithm string) error {
	var hasher hash.Hash
	var extension string
	
	switch strings.ToUpper(algorithm) {
	case "MD5":
		hasher = md5.New()
		extension = ".md5"
	case "SHA1":
		hasher = sha1.New()
		extension = ".sha1"
	case "SHA256":
		hasher = sha256.New()
		extension = ".sha256"
	case "SHA512":
		hasher = sha512.New()
		extension = ".sha512"
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	
	// Calculate hash
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}
	
	checksum := hex.EncodeToString(hasher.Sum(nil))
	
	// Write checksum file
	checksumPath := filePath + extension
	checksumContent := fmt.Sprintf("%s  %s\n", checksum, filepath.Base(filePath))
	
	if err := os.WriteFile(checksumPath, []byte(checksumContent), 0644); err != nil {
		return fmt.Errorf("failed to write checksum file: %w", err)
	}
	
	v.logger.Info(fmt.Sprintf("Created %s checksum: %s", algorithm, checksumPath))
	return nil
}

// AddTrustedKey adds a trusted public key
func (v *Verifier) AddTrustedKey(keyData string) error {
	key, keyID, err := v.parsePublicKey(keyData)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}
	
	v.trustedKeys[keyID] = key
	v.logger.Info(fmt.Sprintf("Added trusted key: %s", keyID))
	return nil
}

// RemoveTrustedKey removes a trusted public key
func (v *Verifier) RemoveTrustedKey(keyID string) error {
	if _, exists := v.trustedKeys[keyID]; !exists {
		return fmt.Errorf("key not found: %s", keyID)
	}
	
	delete(v.trustedKeys, keyID)
	v.logger.Info(fmt.Sprintf("Removed trusted key: %s", keyID))
	return nil
}

// ListTrustedKeys returns a list of trusted key IDs
func (v *Verifier) ListTrustedKeys() []string {
	keys := make([]string, 0, len(v.trustedKeys))
	for keyID := range v.trustedKeys {
		keys = append(keys, keyID)
	}
	return keys
}

// GetKeyInfo returns information about a trusted key
func (v *Verifier) GetKeyInfo(keyID string) (*KeyInfo, error) {
	key, exists := v.trustedKeys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	
	return &KeyInfo{
		ID:       keyID,
		KeySize:  key.Size() * 8, // Convert bytes to bits
		Algorithm: "RSA",
		Usage:    "signature verification",
	}, nil
}

// KeyInfo represents information about a cryptographic key
type KeyInfo struct {
	ID        string `json:"id"`
	KeySize   int    `json:"key_size"`
	Algorithm string `json:"algorithm"`
	Usage     string `json:"usage"`
}

