package update

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// SignatureGenerator handles the generation of digital signatures
type SignatureGenerator struct {
	// Algorithm is the signature algorithm used
	Algorithm SignatureAlgorithm
	// PrivateKey is the private key used for signing
	PrivateKey interface{}
}

// NewSignatureGenerator creates a new SignatureGenerator with the given private key
func NewSignatureGenerator(privateKeyData string) (*SignatureGenerator, error) {
	if privateKeyData == "" {
		return nil, fmt.Errorf("private key is required for signature generation")
	}

	// Try to parse as PEM first
	block, _ := pem.Decode([]byte(privateKeyData))
	if block != nil {
		return parsePrivateKeyFromPEM(block)
	}

	// Try to parse as base64-encoded key
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Try to parse as different key types
	// First try Ed25519
	if len(privateKeyBytes) == ed25519.PrivateKeySize {
		return &SignatureGenerator{
			Algorithm: Ed25519Algorithm,
			PrivateKey: ed25519.PrivateKey(privateKeyBytes),
		}, nil
	}

	// Try X.509 encoded private key (PKCS#8)
	privKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err == nil {
		return createGeneratorFromParsedKey(privKey)
	}

	return nil, fmt.Errorf("unsupported private key format")
}

// parsePrivateKeyFromPEM parses a private key from a PEM block
func parsePrivateKeyFromPEM(block *pem.Block) (*SignatureGenerator, error) {
	switch block.Type {
	case "PRIVATE KEY":
		// PKCS#8, ASN.1 DER form
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
		}
		return createGeneratorFromParsedKey(privKey)

	case "RSA PRIVATE KEY":
		// PKCS#1, RSA private key
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#1 private key: %w", err)
		}
		return &SignatureGenerator{
			Algorithm: RSAAlgorithm,
			PrivateKey: privKey,
		}, nil

	case "EC PRIVATE KEY":
		// SEC1, EC private key
		privKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SEC1 private key: %w", err)
		}
		return &SignatureGenerator{
			Algorithm: ECDSAAlgorithm,
			PrivateKey: privKey,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}

// createGeneratorFromParsedKey creates a SignatureGenerator from a parsed private key
func createGeneratorFromParsedKey(privKey interface{}) (*SignatureGenerator, error) {
	switch key := privKey.(type) {
	case *rsa.PrivateKey:
		return &SignatureGenerator{
			Algorithm: RSAAlgorithm,
			PrivateKey: key,
		}, nil
	case *ecdsa.PrivateKey:
		return &SignatureGenerator{
			Algorithm: ECDSAAlgorithm,
			PrivateKey: key,
		}, nil
	case ed25519.PrivateKey:
		return &SignatureGenerator{
			Algorithm: Ed25519Algorithm,
			PrivateKey: key,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported private key type: %T", privKey)
	}
}

// GenerateSignature generates a digital signature for a file
func (g *SignatureGenerator) GenerateSignature(filePath string) (string, error) {
	// Read the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate the signature based on the algorithm
	var signature []byte
	switch g.Algorithm {
	case Ed25519Algorithm:
		privKey, ok := g.PrivateKey.(ed25519.PrivateKey)
		if !ok {
			return "", fmt.Errorf("invalid private key type for Ed25519")
		}
		signature = ed25519.Sign(privKey, fileContent)

	case RSAAlgorithm:
		privKey, ok := g.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("invalid private key type for RSA")
		}
		hash := sha256.Sum256(fileContent)
		var err error
		signature, err = rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
		if err != nil {
			return "", fmt.Errorf("failed to generate RSA signature: %w", err)
		}

	case ECDSAAlgorithm:
		privKey, ok := g.PrivateKey.(*ecdsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("invalid private key type for ECDSA")
		}
		hash := sha256.Sum256(fileContent)
		var err error
		signature, err = ecdsa.SignASN1(rand.Reader, privKey, hash[:])
		if err != nil {
			return "", fmt.Errorf("failed to generate ECDSA signature: %w", err)
		}

	default:
		return "", fmt.Errorf("unsupported signature algorithm: %s", g.Algorithm)
	}

	// Encode the signature as base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

// GenerateSignatureForData generates a digital signature for the provided data
func (g *SignatureGenerator) GenerateSignatureForData(data []byte) (string, error) {
	// Generate the signature based on the algorithm
	var signature []byte
	switch g.Algorithm {
	case Ed25519Algorithm:
		privKey, ok := g.PrivateKey.(ed25519.PrivateKey)
		if !ok {
			return "", fmt.Errorf("invalid private key type for Ed25519")
		}
		signature = ed25519.Sign(privKey, data)

	case RSAAlgorithm:
		privKey, ok := g.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("invalid private key type for RSA")
		}
		hash := sha256.Sum256(data)
		var err error
		signature, err = rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
		if err != nil {
			return "", fmt.Errorf("failed to generate RSA signature: %w", err)
		}

	case ECDSAAlgorithm:
		privKey, ok := g.PrivateKey.(*ecdsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("invalid private key type for ECDSA")
		}
		hash := sha256.Sum256(data)
		var err error
		signature, err = ecdsa.SignASN1(rand.Reader, privKey, hash[:])
		if err != nil {
			return "", fmt.Errorf("failed to generate ECDSA signature: %w", err)
		}

	default:
		return "", fmt.Errorf("unsupported signature algorithm: %s", g.Algorithm)
	}

	// Encode the signature as base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

// GenerateKeyPair generates a new key pair for the specified algorithm
func GenerateKeyPair(algorithm SignatureAlgorithm) (privateKeyPEM, publicKeyPEM string, err error) {
	switch algorithm {
	case Ed25519Algorithm:
		return generateEd25519KeyPair()
	case RSAAlgorithm:
		return generateRSAKeyPair(2048) // 2048 bits is a reasonable default
	case ECDSAAlgorithm:
		return generateECDSAKeyPair()
	default:
		return "", "", fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}
}

// generateEd25519KeyPair generates a new Ed25519 key pair
func generateEd25519KeyPair() (privateKeyPEM, publicKeyPEM string, err error) {
	// Generate key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	// Marshal private key to PKCS#8
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal Ed25519 private key: %w", err)
	}

	// Marshal public key to PKIX
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal Ed25519 public key: %w", err)
	}

	// Encode private key to PEM
	privateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	// Encode public key to PEM
	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	return privateKeyPEM, publicKeyPEM, nil
}

// generateRSAKeyPair generates a new RSA key pair with the specified bit size
func generateRSAKeyPair(bits int) (privateKeyPEM, publicKeyPEM string, err error) {
	// Generate key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Marshal private key to PKCS#1
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	// Marshal public key to PKIX
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal RSA public key: %w", err)
	}

	// Encode private key to PEM
	privateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	// Encode public key to PEM
	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	return privateKeyPEM, publicKeyPEM, nil
}

// generateECDSAKeyPair generates a new ECDSA key pair using P-256 curve
func generateECDSAKeyPair() (privateKeyPEM, publicKeyPEM string, err error) {
	// Generate key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA key pair: %w", err)
	}

	// Marshal private key to SEC1
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal ECDSA private key: %w", err)
	}

	// Marshal public key to PKIX
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal ECDSA public key: %w", err)
	}

	// Encode private key to PEM
	privateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}))

	// Encode public key to PEM
	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	return privateKeyPEM, publicKeyPEM, nil
}

// CalculateChecksum calculates the SHA-256 checksum of a file
func CalculateChecksum(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Calculate the SHA-256 hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Get the calculated checksum as hex
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
