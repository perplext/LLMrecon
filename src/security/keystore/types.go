// Package keystore provides secure storage for cryptographic keys and sensitive materials.
package keystore

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
)

// KeyType represents the type of cryptographic key
type KeyType string

const (
	// RSAKey represents an RSA key
	RSAKey KeyType = "rsa"
	// ECDSAKey represents an ECDSA key
	ECDSAKey KeyType = "ecdsa"
	// Ed25519Key represents an Ed25519 key
	Ed25519Key KeyType = "ed25519"
	// SymmetricKey represents a symmetric key
	SymmetricKey KeyType = "symmetric"
	// CertificateKey represents a certificate with private key
	CertificateKey KeyType = "certificate"
)

// KeyUsage represents the intended usage of a key
type KeyUsage string

const (
	// SigningKey represents a key used for digital signatures
	SigningKey KeyUsage = "signing"
	// EncryptionKey represents a key used for encryption
	EncryptionKey KeyUsage = "encryption"
	// AuthenticationKey represents a key used for authentication
	AuthenticationKey KeyUsage = "authentication"
	// TLSKey represents a key used for TLS
	TLSKey KeyUsage = "tls"
)

// KeyProtectionLevel represents the level of protection for a key
type KeyProtectionLevel string

const (
	// SoftwareProtection represents software-based protection
	SoftwareProtection KeyProtectionLevel = "software"
	// HSMProtection represents hardware security module protection
	HSMProtection KeyProtectionLevel = "hsm"
	// TPMProtection represents TPM-based protection
	TPMProtection KeyProtectionLevel = "tpm"
	// SecureEnclaveProtection represents secure enclave protection
	SecureEnclaveProtection KeyProtectionLevel = "secure_enclave"
)

// KeyMetadata contains metadata about a key
type KeyMetadata struct {
	// ID is a unique identifier for the key
	ID string `json:"id"`
	// Name is a human-readable name for the key
	Name string `json:"name"`
	// Type is the type of key
	Type KeyType `json:"type"`
	// Usage is the intended usage of the key
	Usage KeyUsage `json:"usage"`
	// ProtectionLevel is the level of protection for the key
	ProtectionLevel KeyProtectionLevel `json:"protection_level"`
	// Algorithm is the specific algorithm (e.g., "RSA-2048", "ECDSA-P256")
	Algorithm string `json:"algorithm"`
	// HasPrivateKey indicates whether a private key is available
	HasPrivateKey bool `json:"has_private_key"`
	// HasPublicKey indicates whether a public key is available
	HasPublicKey bool `json:"has_public_key"`
	// Tags are optional tags for organizing keys
	Tags []string `json:"tags,omitempty"`
	// Description is an optional description of the key
	Description string `json:"description,omitempty"`
	// CreatedAt is when the key was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the key was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt is when the key expires (if applicable)
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	// LastUsedAt is when the key was last used
	LastUsedAt time.Time `json:"last_used_at,omitempty"`
	// LastRotatedAt is when the key was last rotated
	LastRotatedAt time.Time `json:"last_rotated_at,omitempty"`
	// RotationPeriod is the recommended rotation period in days
	RotationPeriod int `json:"rotation_period,omitempty"`
	// Fingerprint is a unique fingerprint of the key
	Fingerprint string `json:"fingerprint,omitempty"`
}

// KeyMaterial contains the actual key material
type KeyMaterial struct {
	// Private contains the private key data
	Private []byte `json:"private,omitempty"`
	// Public contains the public key data
	Public []byte `json:"public,omitempty"`
	// Certificate contains the certificate data (if applicable)
	Certificate []byte `json:"certificate,omitempty"`
	// Format is the format of the key data (e.g., "PEM", "DER")
	Format string `json:"format"`
}

// Key represents a cryptographic key with its metadata
type Key struct {
	// Metadata contains metadata about the key
	Metadata KeyMetadata `json:"metadata"`
	// Material contains the actual key material
	Material KeyMaterial `json:"material,omitempty"`
}

// KeyStore defines the interface for key storage operations
type KeyStore interface {
	// StoreKey stores a key in the key store
	StoreKey(key *Key) error
	
	// GetKey retrieves a key by ID
	GetKey(id string) (*Key, error)
	
	// GetKeyMetadata retrieves key metadata by ID
	GetKeyMetadata(id string) (*KeyMetadata, error)
	
	// DeleteKey deletes a key by ID
	DeleteKey(id string) error
	
	// ListKeys lists all keys in the key store
	ListKeys() ([]*KeyMetadata, error)
	
	// ListKeysByType lists keys of a specific type
	ListKeysByType(keyType KeyType) ([]*KeyMetadata, error)
	
	// ListKeysByUsage lists keys with a specific usage
	ListKeysByUsage(usage KeyUsage) ([]*KeyMetadata, error)
	
	// ListKeysByTag lists keys with a specific tag
	ListKeysByTag(tag string) ([]*KeyMetadata, error)
	
	// RotateKey rotates a key by generating a new key and updating references
	RotateKey(id string) (*Key, error)
	
	// ExportKey exports a key in the specified format
	ExportKey(id string, format string, includePrivate bool) ([]byte, error)
	
	// ImportKey imports a key from the specified format
	ImportKey(keyData []byte, format string, metadata *KeyMetadata) (*Key, error)
	
	// GenerateKey generates a new key with the specified parameters
	GenerateKey(keyType KeyType, algorithm string, metadata *KeyMetadata) (*Key, error)
	
	// GetRSAPrivateKey gets an RSA private key by ID
	GetRSAPrivateKey(id string) (*rsa.PrivateKey, error)
	
	// GetRSAPublicKey gets an RSA public key by ID
	GetRSAPublicKey(id string) (*rsa.PublicKey, error)
	
	// GetECDSAPrivateKey gets an ECDSA private key by ID
	GetECDSAPrivateKey(id string) (*ecdsa.PrivateKey, error)
	
	// GetECDSAPublicKey gets an ECDSA public key by ID
	GetECDSAPublicKey(id string) (*ecdsa.PublicKey, error)
	
	// GetEd25519PrivateKey gets an Ed25519 private key by ID
	GetEd25519PrivateKey(id string) (ed25519.PrivateKey, error)
	
	// GetEd25519PublicKey gets an Ed25519 public key by ID
	GetEd25519PublicKey(id string) (ed25519.PublicKey, error)
	
	// GetCertificate gets a certificate by ID
	GetCertificate(id string) (*x509.Certificate, error)
	
	// Close closes the key store
	Close() error
}

// KeyStoreOptions contains options for creating a key store
type KeyStoreOptions struct {
	// StoragePath is the path to the key store
	StoragePath string
	
	// Passphrase is used to derive the encryption key
	Passphrase string
	
	// HSMConfig contains configuration for HSM integration
	HSMConfig *HSMConfig
	
	// AutoSave determines whether to automatically save after changes
	AutoSave bool
	
	// RotationCheckInterval is how often to check for keys that need rotation
	RotationCheckInterval time.Duration
	
	// AlertCallback is called when a key needs rotation
	AlertCallback func(key *KeyMetadata, daysUntilExpiration int)
}

// HSMConfig contains configuration for HSM integration
type HSMConfig struct {
	// Enabled indicates whether HSM integration is enabled
	Enabled bool
	
	// Provider is the HSM provider (e.g., "pkcs11", "cng", "kms")
	Provider string
	
	// LibraryPath is the path to the HSM library
	LibraryPath string
	
	// SlotID is the HSM slot ID
	SlotID uint
	
	// TokenLabel is the HSM token label
	TokenLabel string
	
	// PIN is the HSM PIN
	PIN string
	
	// KeyLabel is the prefix for key labels in the HSM
	KeyLabel string
}

// KeyRotationPolicy defines when keys should be rotated
type KeyRotationPolicy struct {
	// Enabled indicates whether automatic rotation is enabled
	Enabled bool
	
	// IntervalDays is the number of days between rotations
	IntervalDays int
	
	// LastRotation is the timestamp of the last rotation
	LastRotation time.Time
	
	// WarningDays is the number of days before expiration to start showing warnings
	WarningDays int
}
