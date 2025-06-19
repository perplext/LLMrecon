// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/perplext/LLMrecon/src/update"
)

// SignatureAlgorithm represents the algorithm used for digital signatures
type SignatureAlgorithm string

const (
	// Ed25519Algorithm represents the Ed25519 signature algorithm
	Ed25519Algorithm SignatureAlgorithm = "ed25519"
	// RSAAlgorithm represents the RSA signature algorithm
	RSAAlgorithm SignatureAlgorithm = "rsa"
	// ECDSAAlgorithm represents the ECDSA signature algorithm
	ECDSAAlgorithm SignatureAlgorithm = "ecdsa"
)

// SignatureManager handles the generation and verification of digital signatures
type SignatureManager struct {
	// Algorithm is the signature algorithm used
	Algorithm SignatureAlgorithm
	// Generator is the signature generator
	Generator *update.SignatureGenerator
	// PublicKey is the public key used for verification
	PublicKey ed25519.PublicKey
}

// NewSignatureManager creates a new SignatureManager with the specified algorithm
func NewSignatureManager(algorithm SignatureAlgorithm, privateKey []byte, publicKey []byte) (*SignatureManager, error) {
	// Create signature generator if private key is provided
	var generator *update.SignatureGenerator
	if privateKey != nil && len(privateKey) > 0 {
		privateKeyStr := base64.StdEncoding.EncodeToString(privateKey)
		var err error
		generator, err = update.NewSignatureGenerator(privateKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to create signature generator: %w", err)
		}
	}

	// Set public key if provided
	var pubKey ed25519.PublicKey
	if publicKey != nil && len(publicKey) > 0 {
		if len(publicKey) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(publicKey))
		}
		pubKey = ed25519.PublicKey(publicKey)
	}

	return &SignatureManager{
		Algorithm: algorithm,
		Generator: generator,
		PublicKey: pubKey,
	}, nil
}

// GenerateKeyPair generates a new key pair for signing and verification
func GenerateKeyPair() ([]byte, []byte, error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	return privateKey, publicKey, nil
}

// CalculateManifestSignature calculates a signature for a bundle manifest
func (sm *SignatureManager) CalculateManifestSignature(manifest *BundleManifest) (string, error) {
	if sm.Generator == nil {
		return "", fmt.Errorf("signature generator not initialized")
	}

	// Create a copy of the manifest without the signature field
	manifestCopy := *manifest
	manifestCopy.Signature = ""

	// Marshal the manifest to JSON
	jsonData, err := json.Marshal(manifestCopy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Generate signature
	signature, err := sm.Generator.GenerateSignatureForData(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to generate signature: %w", err)
	}

	return signature, nil
}

// VerifyManifestSignature verifies the signature of a bundle manifest
func (sm *SignatureManager) VerifyManifestSignature(manifest *BundleManifest) (bool, error) {
	if sm.PublicKey == nil {
		return false, fmt.Errorf("public key not initialized")
	}

	// Get the signature from the manifest
	signature := manifest.Signature
	if signature == "" {
		return false, fmt.Errorf("manifest does not have a signature")
	}

	// Create a copy of the manifest without the signature field
	manifestCopy := *manifest
	manifestCopy.Signature = ""

	// Marshal the manifest to JSON
	jsonData, err := json.Marshal(manifestCopy)
	if err != nil {
		return false, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Decode the signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify the signature
	valid := ed25519.Verify(sm.PublicKey, jsonData, signatureBytes)
	return valid, nil
}

// CalculateFileChecksum calculates the SHA-256 checksum of a file
func CalculateFileChecksum(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Calculate checksum
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Return the checksum as a hex string with algorithm prefix
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

// UpdateBundleChecksums updates all checksums in a bundle
func UpdateBundleChecksums(bundle *Bundle) error {
	// Initialize checksums map if not already initialized
	if bundle.Manifest.Checksums.Content == nil {
		bundle.Manifest.Checksums.Content = make(map[string]string)
	}

	// Calculate checksums for all content items
	for _, item := range bundle.Manifest.Content {
		// Get the full path to the content item
		contentPath := filepath.Join(bundle.BundlePath, item.Path)

		// Calculate checksum
		checksum, err := CalculateFileChecksum(contentPath)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum for %s: %w", item.Path, err)
		}

		// Update checksum in content item
		item.Checksum = checksum

		// Update checksum in checksums map
		bundle.Manifest.Checksums.Content[item.Path] = checksum
	}

	// Create a copy of the manifest without the signature field for manifest checksum
	manifestCopy := bundle.Manifest
	manifestCopy.Signature = ""
	manifestCopy.Checksums.Manifest = ""

	// Marshal the manifest copy to calculate its checksum
	manifestData, err := json.Marshal(manifestCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Calculate manifest checksum
	hash := sha256.New()
	hash.Write(manifestData)
	bundle.Manifest.Checksums.Manifest = fmt.Sprintf("sha256:%x", hash.Sum(nil))

	return nil
}

// VerifyBundleChecksums verifies all checksums in a bundle
func VerifyBundleChecksums(bundle *Bundle) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     ChecksumValidationLevel,
		Message:   "All checksums are valid",
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Verify checksums for all content items
	for _, item := range bundle.Manifest.Content {
		// Get the full path to the content item
		contentPath := filepath.Join(bundle.BundlePath, item.Path)

		// Calculate actual checksum
		actualChecksum, err := CalculateFileChecksum(contentPath)
		if err != nil {
			result.Valid = false
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to calculate checksum for %s: %v", item.Path, err))
			continue
		}

		// Get expected checksum from manifest
		expectedChecksum := item.Checksum
		if expectedChecksum == "" {
			// Try to get from checksums map
			expectedChecksum = bundle.Manifest.Checksums.Content[item.Path]
		}

		if expectedChecksum == "" {
			result.Valid = false
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("No checksum found for %s", item.Path))
			continue
		}

		// Compare checksums
		if !strings.EqualFold(actualChecksum, expectedChecksum) {
			result.Valid = false
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Checksum mismatch for %s: expected %s, got %s", item.Path, expectedChecksum, actualChecksum))
		}
	}

	// Verify manifest checksum
	manifestCopy := bundle.Manifest
	expectedManifestChecksum := manifestCopy.Checksums.Manifest
	manifestCopy.Signature = ""
	manifestCopy.Checksums.Manifest = ""

	// Marshal the manifest copy to calculate its checksum
	manifestData, err := json.Marshal(manifestCopy)
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to marshal manifest: %v", err))
	} else {
		// Calculate actual manifest checksum
		hash := sha256.New()
		hash.Write(manifestData)
		actualManifestChecksum := fmt.Sprintf("sha256:%x", hash.Sum(nil))

		// Compare checksums
		if !strings.EqualFold(actualManifestChecksum, expectedManifestChecksum) {
			result.Valid = false
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Manifest checksum mismatch: expected %s, got %s", expectedManifestChecksum, actualManifestChecksum))
		}
	}

	// Update message and error if validation failed
	if !result.Valid {
		result.Message = "Checksum validation failed"
		result.Error = fmt.Errorf("one or more checksums are invalid")
	}

	return result, nil
}

// SignBundle signs a bundle using the provided private key
func SignBundle(bundle *Bundle, privateKey []byte) error {
	// Create signature manager
	signatureManager, err := NewSignatureManager(Ed25519Algorithm, privateKey, nil)
	if err != nil {
		return fmt.Errorf("failed to create signature manager: %w", err)
	}

	// Update bundle checksums
	if err := UpdateBundleChecksums(bundle); err != nil {
		return fmt.Errorf("failed to update bundle checksums: %w", err)
	}

	// Calculate signature
	signature, err := signatureManager.CalculateManifestSignature(&bundle.Manifest)
	if err != nil {
		return fmt.Errorf("failed to calculate manifest signature: %w", err)
	}

	// Set signature in manifest
	bundle.Manifest.Signature = signature

	return nil
}

// VerifyBundle verifies a bundle's signature and checksums
func VerifyBundle(bundle *Bundle, publicKey []byte) (*ValidationResult, error) {
	// Create signature manager
	signatureManager, err := NewSignatureManager(Ed25519Algorithm, nil, publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature manager: %w", err)
	}

	// Create validation result
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     SignatureValidationLevel,
		Message:   "Bundle signature is valid",
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Verify signature
	valid, err := signatureManager.VerifyManifestSignature(&bundle.Manifest)
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Error = err
		result.Message = "Failed to verify bundle signature"
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	if !valid {
		result.Valid = false
		result.IsValid = false
		result.Message = "Bundle signature is invalid"
		result.Errors = append(result.Errors, "Bundle signature verification failed")
		result.Error = fmt.Errorf("invalid bundle signature")
		return result, nil
	}

	// Verify checksums
	checksumResult, err := VerifyBundleChecksums(bundle)
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Error = err
		result.Message = "Failed to verify bundle checksums"
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	if !checksumResult.Valid {
		result.Valid = false
		result.IsValid = false
		result.Message = "Bundle checksums are invalid"
		result.Errors = append(result.Errors, checksumResult.Errors...)
		result.Error = checksumResult.Error
	}

	return result, nil
}
