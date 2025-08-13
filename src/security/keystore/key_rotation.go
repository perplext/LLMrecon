// Package keystore provides secure storage for cryptographic keys and sensitive materials.
package keystore

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	
	"github.com/perplext/LLMrecon/src/security/vault"
)

// RotateKey rotates a key by generating a new key and updating references
func (ks *FileKeyStore) RotateKey(id string) (*Key, error) {
	// Get the existing key
	oldKey, err := ks.GetKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get key for rotation: %w", err)
	}

	// Log the rotation attempt
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("rotate", oldKey.Metadata.ID, "Attempting key rotation")
	}

	// Create new metadata based on the old key
	newMetadata := oldKey.Metadata
	newMetadata.ID = generateKeyID(newMetadata.Name, newMetadata.Type)
	newMetadata.CreatedAt = time.Now()
	newMetadata.UpdatedAt = time.Now()
	newMetadata.LastRotatedAt = time.Now()
	
	// Set expiration if rotation period is defined
	if newMetadata.RotationPeriod > 0 {
		newMetadata.ExpiresAt = time.Now().AddDate(0, 0, newMetadata.RotationPeriod)
	}

	// Generate a new key with the same parameters
	newKey, err := ks.GenerateKey(oldKey.Metadata.Type, oldKey.Metadata.Algorithm, &newMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new key during rotation: %w", err)
	}

	// Store the relationship between the old and new key
	// This is useful for audit trails and for applications that need to know about key rotation
	rotationInfo := map[string]string{
		"previous_key_id": oldKey.Metadata.ID,
		"new_key_id":      newKey.Metadata.ID,
		"rotation_time":   time.Now().Format(time.RFC3339),
	}

	rotationInfoBytes, err := json.Marshal(rotationInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rotation info: %w", err)
	}

	// Store the rotation information in the vault
	rotationCredential := &vault.Credential{
		ID:          fmt.Sprintf("rotation_%s_%s", oldKey.Metadata.ID, newKey.Metadata.ID),
		Name:        fmt.Sprintf("Key Rotation %s -> %s", oldKey.Metadata.ID, newKey.Metadata.ID),
		Value:       string(rotationInfoBytes),
		Description: "Key rotation relationship information",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiresAt:   time.Time{}, // Never expires
		Tags:        []string{"key_rotation", oldKey.Metadata.ID, newKey.Metadata.ID},
	}

	if err := ks.vault.StoreCredential(rotationCredential); err != nil {
		return nil, fmt.Errorf("failed to store rotation information: %w", err)
	}

	// Log the successful rotation
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("rotate", oldKey.Metadata.ID, 
			fmt.Sprintf("Key rotated successfully, new key ID: %s", newKey.Metadata.ID))
	}

	// Save changes if auto-save is enabled
	if ks.autoSave {
		if err := ks.save(); err != nil {
			return nil, fmt.Errorf("failed to save key store after rotation: %w", err)
		}
	}

	return newKey, nil
}

// GenerateKey generates a new key with the specified parameters
func (ks *FileKeyStore) GenerateKey(keyType KeyType, algorithm string, metadata *KeyMetadata) (*Key, error) {
	// Initialize metadata if not provided
	if metadata == nil {
		return nil, errors.New("metadata is required for key generation")
	}

	// Set default values for metadata
	if metadata.ID == "" {
		metadata.ID = generateKeyID(metadata.Name, keyType)
	}
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}
	if metadata.UpdatedAt.IsZero() {
		metadata.UpdatedAt = time.Now()
	}
	
	// Set expiration if rotation period is defined
	if metadata.RotationPeriod > 0 && metadata.ExpiresAt.IsZero() {
		metadata.ExpiresAt = time.Now().AddDate(0, 0, metadata.RotationPeriod)
	}

	// Set key type and algorithm
	metadata.Type = keyType
	if algorithm != "" {
		metadata.Algorithm = algorithm
	}

	// Generate key material based on type and algorithm
	var keyMaterial KeyMaterial
	var err error

	switch keyType {
	case RSAKey:
		keyMaterial, err = generateRSAKey(algorithm)
	case ECDSAKey:
		keyMaterial, err = generateECDSAKey(algorithm)
	case Ed25519Key:
		keyMaterial, err = generateEd25519Key()
	case SymmetricKey:
		keyMaterial, err = generateSymmetricKey(algorithm)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate key material: %w", err)
	}

	// Calculate fingerprint
	fingerprint, err := calculateKeyFingerprint(keyMaterial.Public)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate key fingerprint: %w", err)
	}
	metadata.Fingerprint = fingerprint

	// Set key flags
	metadata.HasPrivateKey = len(keyMaterial.Private) > 0
	metadata.HasPublicKey = len(keyMaterial.Public) > 0

	// Create the key
	key := &Key{
		Metadata: *metadata,
		Material: keyMaterial,
	}

	// Store the key
	if err := ks.StoreKey(key); err != nil {
		return nil, fmt.Errorf("failed to store generated key: %w", err)
	}

	// Log the key generation
	if ks.auditLogger != nil {
		ks.auditLogger.LogKeyOperation("generate", metadata.ID, 
			fmt.Sprintf("Generated new %s key with algorithm %s", keyType, algorithm))
	}

	return key, nil
}

// generateRSAKey generates an RSA key pair
func generateRSAKey(algorithm string) (KeyMaterial, error) {
	// Determine key size from algorithm
	keySize := 2048 // Default
	if algorithm == "RSA-4096" {
		keySize = 4096
	} else if algorithm == "RSA-3072" {
		keySize = 3072
	} else if algorithm == "RSA-2048" {
		keySize = 2048
	} else if algorithm != "" && algorithm != "RSA-2048" {
		return KeyMaterial{}, fmt.Errorf("unsupported RSA algorithm: %s", algorithm)
	}

	// Generate key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return KeyMaterial{}, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return KeyMaterial{}, fmt.Errorf("failed to marshal RSA public key: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return KeyMaterial{
		Private: privatePEM,
		Public:  publicPEM,
		Format:  "PEM",
	}, nil
}

// generateECDSAKey generates an ECDSA key pair
func generateECDSAKey(algorithm string) (KeyMaterial, error) {
	// Determine curve from algorithm
	var curve elliptic.Curve
	switch algorithm {
	case "ECDSA-P256":
		curve = elliptic.P256()
	case "ECDSA-P384":
		curve = elliptic.P384()
	case "ECDSA-P521":
		curve = elliptic.P521()
	case "":
		curve = elliptic.P256() // Default
	default:
		return KeyMaterial{}, fmt.Errorf("unsupported ECDSA algorithm: %s", algorithm)
	}

	// Generate key
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return KeyMaterial{}, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	// Encode private key to PEM
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return KeyMaterial{}, fmt.Errorf("failed to marshal ECDSA private key: %w", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return KeyMaterial{}, fmt.Errorf("failed to marshal ECDSA public key: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return KeyMaterial{
		Private: privatePEM,
		Public:  publicPEM,
		Format:  "PEM",
	}, nil
}

// generateEd25519Key generates an Ed25519 key pair
func generateEd25519Key() (KeyMaterial, error) {
	// Generate key
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyMaterial{}, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	// Encode private key to PEM
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	})

	// Encode public key to PEM
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	})

	return KeyMaterial{
		Private: privatePEM,
		Public:  publicPEM,
		Format:  "PEM",
	}, nil
}

// generateSymmetricKey generates a symmetric key
func generateSymmetricKey(algorithm string) (KeyMaterial, error) {
	// Determine key size from algorithm
	keySize := 32 // Default (256 bits)
	if algorithm == "AES-128" {
		keySize = 16
	} else if algorithm == "AES-192" {
		keySize = 24
	} else if algorithm == "AES-256" {
		keySize = 32
	} else if algorithm != "" && algorithm != "AES-256" {
		return KeyMaterial{}, fmt.Errorf("unsupported symmetric algorithm: %s", algorithm)
	}

	// Generate random bytes
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return KeyMaterial{}, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	return KeyMaterial{
		Private: key,
		Format:  "RAW",
	}, nil
}
