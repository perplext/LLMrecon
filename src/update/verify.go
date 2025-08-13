package update

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

// SignatureAlgorithm represents the type of signature algorithm
type SignatureAlgorithm string

const (
	// Ed25519Algorithm represents the Ed25519 signature algorithm
	Ed25519Algorithm SignatureAlgorithm = "ed25519"
	// RSAAlgorithm represents the RSA signature algorithm
	RSAAlgorithm SignatureAlgorithm = "rsa"
	// ECDSAAlgorithm represents the ECDSA signature algorithm
	ECDSAAlgorithm SignatureAlgorithm = "ecdsa"
)

// SignatureVerifier handles verification of update signatures
type SignatureVerifier struct {
	// Algorithm is the signature algorithm used
	Algorithm SignatureAlgorithm
	// PublicKey is the public key used for verification
	PublicKey interface{}
}

// NewSignatureVerifier creates a new SignatureVerifier with the given public key
func NewSignatureVerifier(publicKeyData string) (*SignatureVerifier, error) {
	if publicKeyData == "" {
		return nil, fmt.Errorf("public key is required for signature verification")
	}

	// Try to parse as PEM first
	block, _ := pem.Decode([]byte(publicKeyData))
	if block != nil {
		return parsePublicKeyFromPEM(block)
	}

	// Try to parse as base64-encoded key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Try to parse as different key types
	// First try Ed25519
	if len(publicKeyBytes) == ed25519.PublicKeySize {
		return &SignatureVerifier{
			Algorithm: Ed25519Algorithm,
			PublicKey: ed25519.PublicKey(publicKeyBytes),
		}, nil
	}

	// Try X.509 encoded public key
	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err == nil {
		return createVerifierFromParsedKey(pubKey)
	}

	return nil, fmt.Errorf("unsupported public key format")
}

// parsePublicKeyFromPEM parses a public key from a PEM block
func parsePublicKeyFromPEM(block *pem.Block) (*SignatureVerifier, error) {
	switch block.Type {
	case "PUBLIC KEY":
		// PKIX, ASN.1 DER form
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
		}
		return createVerifierFromParsedKey(pubKey)

	case "RSA PUBLIC KEY":
		// PKCS#1, RSA public key
		pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#1 public key: %w", err)
		}
		return &SignatureVerifier{
			Algorithm: RSAAlgorithm,
			PublicKey: pubKey,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}

// createVerifierFromParsedKey creates a SignatureVerifier from a parsed public key
func createVerifierFromParsedKey(pubKey interface{}) (*SignatureVerifier, error) {
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		return &SignatureVerifier{
			Algorithm: RSAAlgorithm,
			PublicKey: key,
		}, nil
	case *ecdsa.PublicKey:
		return &SignatureVerifier{
			Algorithm: ECDSAAlgorithm,
			PublicKey: key,
		}, nil
	case ed25519.PublicKey:
		return &SignatureVerifier{
			Algorithm: Ed25519Algorithm,
			PublicKey: key,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pubKey)
	}
}

// VerifySignature verifies the signature of a file
func (v *SignatureVerifier) VerifySignature(filePath, signatureBase64 string) error {
	// Read the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Calculate the SHA-256 hash of the file content
	hash := sha256.Sum256(fileContent)

	// Verify the signature based on the algorithm
	switch v.Algorithm {
	case Ed25519Algorithm:
		pubKey, ok := v.PublicKey.(ed25519.PublicKey)
		if !ok {
			return fmt.Errorf("invalid public key type for Ed25519")
		}
		if !ed25519.Verify(pubKey, fileContent, signature) {
			return fmt.Errorf("Ed25519 signature verification failed")
		}

	case RSAAlgorithm:
		pubKey, ok := v.PublicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("invalid public key type for RSA")
		}
		if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature); err != nil {
			return fmt.Errorf("RSA signature verification failed: %w", err)
		}

	case ECDSAAlgorithm:
		pubKey, ok := v.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("invalid public key type for ECDSA")
		}
		if !ecdsa.VerifyASN1(pubKey, hash[:], signature) {
			return fmt.Errorf("ECDSA signature verification failed")
		}

	default:
		return fmt.Errorf("unsupported signature algorithm: %s", v.Algorithm)
	}

	return nil
}

// VerifyChecksum verifies the SHA-256 checksum of a file
func VerifyChecksum(filePath, expectedChecksumHex string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Calculate the SHA-256 hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Get the calculated checksum
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Compare with the expected checksum
	if calculatedChecksum != expectedChecksumHex {
		return fmt.Errorf("checksum verification failed: expected %s, got %s", expectedChecksumHex, calculatedChecksum)
	}

	return nil
}

// VerifyUpdate performs both signature and checksum verification on an update file
func VerifyUpdate(filePath, checksumHex, signatureBase64, publicKeyBase64 string) error {
	// First verify the checksum
	if err := VerifyChecksum(filePath, checksumHex); err != nil {
		return err
	}

	// If signature verification is requested
	if signatureBase64 != "" && publicKeyBase64 != "" {
		verifier, err := NewSignatureVerifier(publicKeyBase64)
		if err != nil {
			return err
		}

		if err := verifier.VerifySignature(filePath, signatureBase64); err != nil {
			return err
		}
	}

	return nil
}

// CalculateChecksum calculates the SHA256 checksum of a file
func CalculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("calculating checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// IntegrityReport represents the result of integrity verification
type IntegrityReport struct {
	FilePath         string
	ChecksumValid    bool
	ExpectedChecksum string
	ActualChecksum   string
	SignatureValid   bool
	SignatureError   error
	FileSize         int64
	VerifiedAt       string
}

// VerifyIntegrity performs comprehensive integrity verification
func VerifyIntegrity(filePath, expectedChecksum, signature, publicKey string) (*IntegrityReport, error) {
	report := &IntegrityReport{
		FilePath:         filePath,
		ExpectedChecksum: expectedChecksum,
		VerifiedAt:       time.Now().Format(time.RFC3339),
	}

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return report, fmt.Errorf("getting file info: %w", err)
	}
	report.FileSize = info.Size()

	// Calculate actual checksum
	actualChecksum, err := CalculateChecksum(filePath)
	if err != nil {
		return report, fmt.Errorf("calculating checksum: %w", err)
	}
	report.ActualChecksum = actualChecksum

	// Verify checksum
	report.ChecksumValid = actualChecksum == expectedChecksum

	// Verify signature if provided
	if signature != "" && publicKey != "" {
		verifier, err := NewSignatureVerifier(publicKey)
		if err != nil {
			report.SignatureError = err
		} else {
			err = verifier.VerifySignature(filePath, signature)
			report.SignatureValid = err == nil
			report.SignatureError = err
		}
	}

	return report, nil
}
