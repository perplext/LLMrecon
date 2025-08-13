// Package keystore provides secure storage for cryptographic keys and sensitive materials.
package keystore

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// GetRSAPrivateKey gets an RSA private key by ID
func (ks *FileKeyStore) GetRSAPrivateKey(id string) (*rsa.PrivateKey, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Check key type
	if key.Metadata.Type != RSAKey {
		return nil, fmt.Errorf("key is not an RSA key, it is %s", key.Metadata.Type)
	}

	// Check if private key is available
	if !key.Metadata.HasPrivateKey {
		return nil, errors.New("private key is not available")
	}

	// Log the key access
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("access", key.Metadata.ID, "Retrieved RSA private key")
	}

	// Handle HSM-protected keys
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.GetRSAPrivateKey(id)
	}

	// Parse the private key
	switch key.Material.Format {
	case "PEM":
		// Decode PEM
		block, _ := pem.Decode(key.Material.Private)
		if block == nil {
			return nil, errors.New("failed to decode PEM block")
		}

		// Parse private key
		if block.Type == "RSA PRIVATE KEY" {
			// PKCS#1 format
			return x509.ParsePKCS1PrivateKey(block.Bytes)
		} else if block.Type == "PRIVATE KEY" {
			// PKCS#8 format
			pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
			}
			rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("key is not an RSA private key")
			}
			return rsaKey, nil
		}
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)

	case "DER":
		// Try PKCS#1 format first
		privateKey, err := x509.ParsePKCS1PrivateKey(key.Material.Private)
		if err == nil {
			return privateKey, nil
		}

		// Try PKCS#8 format
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(key.Material.Private)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("key is not an RSA private key")
		}
		return rsaKey, nil

	default:
		return nil, fmt.Errorf("unsupported key format: %s", key.Material.Format)
	}
}

// GetRSAPublicKey gets an RSA public key by ID
func (ks *FileKeyStore) GetRSAPublicKey(id string) (*rsa.PublicKey, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Check key type
	if key.Metadata.Type != RSAKey {
		return nil, fmt.Errorf("key is not an RSA key, it is %s", key.Metadata.Type)
	}

	// Check if public key is available
	if !key.Metadata.HasPublicKey {
		return nil, errors.New("public key is not available")
	}

	// Log the key access
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("access", key.Metadata.ID, "Retrieved RSA public key")
	}

	// Handle HSM-protected keys
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.GetRSAPublicKey(id)
	}

	// Parse the public key
	switch key.Material.Format {
	case "PEM":
		// Decode PEM
		block, _ := pem.Decode(key.Material.Public)
		if block == nil {
			return nil, errors.New("failed to decode PEM block")
		}

		// Parse public key
		if block.Type == "RSA PUBLIC KEY" {
			// PKCS#1 format
			return x509.ParsePKCS1PublicKey(block.Bytes)
		} else if block.Type == "PUBLIC KEY" {
			// PKIX format
			pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
			}
			rsaKey, ok := pubKey.(*rsa.PublicKey)
			if !ok {
				return nil, errors.New("key is not an RSA public key")
			}
			return rsaKey, nil
		}
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)

	case "DER":
		// Try PKCS#1 format first
		publicKey, err := x509.ParsePKCS1PublicKey(key.Material.Public)
		if err == nil {
			return publicKey, nil
		}

		// Try PKIX format
		pubKey, err := x509.ParsePKIXPublicKey(key.Material.Public)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		rsaKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("key is not an RSA public key")
		}
		return rsaKey, nil

	default:
		return nil, fmt.Errorf("unsupported key format: %s", key.Material.Format)
	}
}

// GetECDSAPrivateKey gets an ECDSA private key by ID
func (ks *FileKeyStore) GetECDSAPrivateKey(id string) (*ecdsa.PrivateKey, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Check key type
	if key.Metadata.Type != ECDSAKey {
		return nil, fmt.Errorf("key is not an ECDSA key, it is %s", key.Metadata.Type)
	}

	// Check if private key is available
	if !key.Metadata.HasPrivateKey {
		return nil, errors.New("private key is not available")
	}

	// Log the key access
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("access", key.Metadata.ID, "Retrieved ECDSA private key")
	}

	// Handle HSM-protected keys
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.GetECDSAPrivateKey(id)
	}

	// Parse the private key
	switch key.Material.Format {
	case "PEM":
		// Decode PEM
		block, _ := pem.Decode(key.Material.Private)
		if block == nil {
			return nil, errors.New("failed to decode PEM block")
		}

		// Parse private key
		if block.Type == "EC PRIVATE KEY" {
			// SEC1 format
			return x509.ParseECPrivateKey(block.Bytes)
		} else if block.Type == "PRIVATE KEY" {
			// PKCS#8 format
			pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
			}
			ecdsaKey, ok := pkcs8Key.(*ecdsa.PrivateKey)
			if !ok {
				return nil, errors.New("key is not an ECDSA private key")
			}
			return ecdsaKey, nil
		}
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)

	case "DER":
		// Try SEC1 format first
		privateKey, err := x509.ParseECPrivateKey(key.Material.Private)
		if err == nil {
			return privateKey, nil
		}

		// Try PKCS#8 format
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(key.Material.Private)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		ecdsaKey, ok := pkcs8Key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("key is not an ECDSA private key")
		}
		return ecdsaKey, nil

	default:
		return nil, fmt.Errorf("unsupported key format: %s", key.Material.Format)
	}
}

// GetECDSAPublicKey gets an ECDSA public key by ID
func (ks *FileKeyStore) GetECDSAPublicKey(id string) (*ecdsa.PublicKey, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Check key type
	if key.Metadata.Type != ECDSAKey {
		return nil, fmt.Errorf("key is not an ECDSA key, it is %s", key.Metadata.Type)
	}

	// Check if public key is available
	if !key.Metadata.HasPublicKey {
		return nil, errors.New("public key is not available")
	}

	// Log the key access
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("access", key.Metadata.ID, "Retrieved ECDSA public key")
	}

	// Handle HSM-protected keys
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.GetECDSAPublicKey(id)
	}

	// Parse the public key
	switch key.Material.Format {
	case "PEM":
		// Decode PEM
		block, _ := pem.Decode(key.Material.Public)
		if block == nil {
			return nil, errors.New("failed to decode PEM block")
		}

		// Parse public key
		if block.Type == "EC PUBLIC KEY" || block.Type == "PUBLIC KEY" {
			// PKIX format
			pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
			}
			ecdsaKey, ok := pubKey.(*ecdsa.PublicKey)
			if !ok {
				return nil, errors.New("key is not an ECDSA public key")
			}
			return ecdsaKey, nil
		}
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)

	case "DER":
		// Try PKIX format
		pubKey, err := x509.ParsePKIXPublicKey(key.Material.Public)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		ecdsaKey, ok := pubKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("key is not an ECDSA public key")
		}
		return ecdsaKey, nil

	default:
		return nil, fmt.Errorf("unsupported key format: %s", key.Material.Format)
	}
}

// GetEd25519PrivateKey gets an Ed25519 private key by ID
func (ks *FileKeyStore) GetEd25519PrivateKey(id string) (ed25519.PrivateKey, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Check key type
	if key.Metadata.Type != Ed25519Key {
		return nil, fmt.Errorf("key is not an Ed25519 key, it is %s", key.Metadata.Type)
	}

	// Check if private key is available
	if !key.Metadata.HasPrivateKey {
		return nil, errors.New("private key is not available")
	}

	// Log the key access
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("access", key.Metadata.ID, "Retrieved Ed25519 private key")
	}

	// Handle HSM-protected keys
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.GetEd25519PrivateKey(id)
	}

	// Parse the private key
	switch key.Material.Format {
	case "PEM":
		// Decode PEM
		block, _ := pem.Decode(key.Material.Private)
		if block == nil {
			return nil, errors.New("failed to decode PEM block")
		}

		// Parse private key (Ed25519 keys are typically in PKCS#8 format)
		if block.Type == "PRIVATE KEY" {
			pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
			}
			ed25519Key, ok := pkcs8Key.(ed25519.PrivateKey)
			if !ok {
				return nil, errors.New("key is not an Ed25519 private key")
			}
			return ed25519Key, nil
		}
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)

	case "DER":
		// Try PKCS#8 format
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(key.Material.Private)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		ed25519Key, ok := pkcs8Key.(ed25519.PrivateKey)
		if !ok {
			return nil, errors.New("key is not an Ed25519 private key")
		}
		return ed25519Key, nil

	default:
		return nil, fmt.Errorf("unsupported key format: %s", key.Material.Format)
	}
}

// GetEd25519PublicKey gets an Ed25519 public key by ID
func (ks *FileKeyStore) GetEd25519PublicKey(id string) (ed25519.PublicKey, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Check key type
	if key.Metadata.Type != Ed25519Key {
		return nil, fmt.Errorf("key is not an Ed25519 key, it is %s", key.Metadata.Type)
	}

	// Check if public key is available
	if !key.Metadata.HasPublicKey {
		return nil, errors.New("public key is not available")
	}

	// Log the key access
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("access", key.Metadata.ID, "Retrieved Ed25519 public key")
	}

	// Handle HSM-protected keys
	if key.Metadata.ProtectionLevel == HSMProtection && ks.hsmManager != nil {
		return ks.hsmManager.GetEd25519PublicKey(id)
	}

	// Parse the public key
	switch key.Material.Format {
	case "PEM":
		// Decode PEM
		block, _ := pem.Decode(key.Material.Public)
		if block == nil {
			return nil, errors.New("failed to decode PEM block")
		}

		// Parse public key
		if block.Type == "PUBLIC KEY" {
			// PKIX format
			pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
			}
			ed25519Key, ok := pubKey.(ed25519.PublicKey)
			if !ok {
				return nil, errors.New("key is not an Ed25519 public key")
			}
			return ed25519Key, nil
		}
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)

	case "DER":
		// Try PKIX format
		pubKey, err := x509.ParsePKIXPublicKey(key.Material.Public)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		ed25519Key, ok := pubKey.(ed25519.PublicKey)
		if !ok {
			return nil, errors.New("key is not an Ed25519 public key")
		}
		return ed25519Key, nil

	default:
		return nil, fmt.Errorf("unsupported key format: %s", key.Material.Format)
	}
}

// GetCertificate gets a certificate by ID
func (ks *FileKeyStore) GetCertificate(id string) (*x509.Certificate, error) {
	// Get the key
	key, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Check if certificate is available
	if len(key.Material.Certificate) == 0 {
		return nil, errors.New("certificate is not available")
	}

	// Log the certificate access
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("access", key.Metadata.ID, "Retrieved certificate")
	}

	// Parse the certificate
	switch key.Material.Format {
	case "PEM":
		// Decode PEM
		block, _ := pem.Decode(key.Material.Certificate)
		if block == nil {
			return nil, errors.New("failed to decode PEM block")
		}

		// Parse certificate
		if block.Type == "CERTIFICATE" {
			return x509.ParseCertificate(block.Bytes)
		}
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)

	case "DER":
		// Parse DER-encoded certificate
		return x509.ParseCertificate(key.Material.Certificate)

	default:
		return nil, fmt.Errorf("unsupported certificate format: %s", key.Material.Format)
	}
}
