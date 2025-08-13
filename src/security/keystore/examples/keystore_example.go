// Example program demonstrating the usage of the keystore package
package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/perplext/LLMrecon/src/security/keystore"
)

func main() {
	// Create a directory for the key store
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	keyStorePath := filepath.Join(homeDir, ".LLMrecon", "keystore", "keys.json")
	keyStoreDir := filepath.Dir(keyStorePath)
	
	if err := os.MkdirAll(keyStoreDir, 0700); err != nil {
		log.Fatalf("Failed to create key store directory: %v", err)
	}

	// Create key store options
	options := keystore.KeyStoreOptions{
		StoragePath:          keyStorePath,
		Passphrase:           "example-passphrase", // In production, use a secure passphrase
		AutoSave:             true,
		RotationCheckInterval: time.Hour * 24, // Check for rotation daily
		AlertCallback: func(key *keystore.KeyMetadata, daysUntilExpiration int) {
			log.Printf("Key %s (%s) will expire in %d days", key.Name, key.ID, daysUntilExpiration)
		},
	}

	// Create key store
	ks, err := keystore.NewFileKeyStore(options)
	if err != nil {
		log.Fatalf("Failed to create key store: %v", err)
	}
	defer ks.Close()

	// Generate RSA signing key
	rsaKey, err := generateRSASigningKey(ks)
	if err != nil {
		log.Fatalf("Failed to generate RSA signing key: %v", err)
	}
	log.Printf("Generated RSA signing key: %s (%s)", rsaKey.Metadata.Name, rsaKey.Metadata.ID)

	// Generate ECDSA signing key
	ecdsaKey, err := generateECDSASigningKey(ks)
	if err != nil {
		log.Fatalf("Failed to generate ECDSA signing key: %v", err)
	}
	log.Printf("Generated ECDSA signing key: %s (%s)", ecdsaKey.Metadata.Name, ecdsaKey.Metadata.ID)

	// Generate symmetric encryption key
	symmetricKey, err := generateSymmetricKey(ks)
	if err != nil {
		log.Fatalf("Failed to generate symmetric key: %v", err)
	}
	log.Printf("Generated symmetric key: %s (%s)", symmetricKey.Metadata.Name, symmetricKey.Metadata.ID)

	// List all keys
	listAllKeys(ks)

	// Demonstrate key usage
	demonstrateRSAKeyUsage(ks, rsaKey.Metadata.ID)

	// Demonstrate key rotation
	demonstrateKeyRotation(ks, rsaKey.Metadata.ID)

	// Demonstrate key export and import
	demonstrateKeyExportImport(ks, ecdsaKey.Metadata.ID)

	log.Println("Key store example completed successfully")
}

// generateRSASigningKey generates an RSA key for digital signatures
func generateRSASigningKey(ks *keystore.FileKeyStore) (*keystore.Key, error) {
	metadata := &keystore.KeyMetadata{
		Name:            "example-rsa-signing-key",
		Type:            keystore.RSAKey,
		Usage:           keystore.SigningKey,
		ProtectionLevel: keystore.SoftwareProtection,
		Algorithm:       "RSA-2048",
		Tags:            []string{"example", "signing"},
		Description:     "Example RSA signing key",
		RotationPeriod:  90, // 90 days
	}

	return ks.GenerateKey(keystore.RSAKey, "RSA-2048", metadata)
}

// generateECDSASigningKey generates an ECDSA key for digital signatures
func generateECDSASigningKey(ks *keystore.FileKeyStore) (*keystore.Key, error) {
	metadata := &keystore.KeyMetadata{
		Name:            "example-ecdsa-signing-key",
		Type:            keystore.ECDSAKey,
		Usage:           keystore.SigningKey,
		ProtectionLevel: keystore.SoftwareProtection,
		Algorithm:       "ECDSA-P256",
		Tags:            []string{"example", "signing"},
		Description:     "Example ECDSA signing key",
		RotationPeriod:  90, // 90 days
	}

	return ks.GenerateKey(keystore.ECDSAKey, "ECDSA-P256", metadata)
}

// generateSymmetricKey generates a symmetric key for encryption
func generateSymmetricKey(ks *keystore.FileKeyStore) (*keystore.Key, error) {
	metadata := &keystore.KeyMetadata{
		Name:            "example-symmetric-key",
		Type:            keystore.SymmetricKey,
		Usage:           keystore.EncryptionKey,
		ProtectionLevel: keystore.SoftwareProtection,
		Algorithm:       "AES-256",
		Tags:            []string{"example", "encryption"},
		Description:     "Example symmetric encryption key",
		RotationPeriod:  30, // 30 days
	}

	return ks.GenerateKey(keystore.SymmetricKey, "AES-256", metadata)
}

// listAllKeys lists all keys in the key store
func listAllKeys(ks *keystore.FileKeyStore) {
	// List all keys
	keys, err := ks.ListKeys()
	if err != nil {
		log.Printf("Failed to list keys: %v", err)
		return
	}

	log.Printf("Found %d keys in the key store:", len(keys))
	for _, key := range keys {
		log.Printf("  - %s (%s): %s, %s, %s", key.Name, key.ID, key.Type, key.Usage, key.Algorithm)
	}

	// List keys by type
	rsaKeys, err := ks.ListKeysByType(keystore.RSAKey)
	if err != nil {
		log.Printf("Failed to list RSA keys: %v", err)
		return
	}
	log.Printf("Found %d RSA keys", len(rsaKeys))

	// List keys by usage
	signingKeys, err := ks.ListKeysByUsage(keystore.SigningKey)
	if err != nil {
		log.Printf("Failed to list signing keys: %v", err)
		return
	}
	log.Printf("Found %d signing keys", len(signingKeys))

	// List keys by tag
	exampleKeys, err := ks.ListKeysByTag("example")
	if err != nil {
		log.Printf("Failed to list keys with tag 'example': %v", err)
		return
	}
	log.Printf("Found %d keys with tag 'example'", len(exampleKeys))
}

// demonstrateRSAKeyUsage demonstrates how to use an RSA key for signing and verification
func demonstrateRSAKeyUsage(ks *keystore.FileKeyStore, keyID string) {
	// Get the RSA private key
	privateKey, err := ks.GetRSAPrivateKey(keyID)
	if err != nil {
		log.Printf("Failed to get RSA private key: %v", err)
		return
	}

	// Get the RSA public key
	publicKey, err := ks.GetRSAPublicKey(keyID)
	if err != nil {
		log.Printf("Failed to get RSA public key: %v", err)
		return
	}

	// Data to sign
	data := []byte("Hello, world!")

	// Calculate hash
	hash := sha256.Sum256(data)

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		log.Printf("Failed to sign data: %v", err)
		return
	}
	log.Printf("Created signature: %s", hex.EncodeToString(signature))

	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		log.Printf("Signature verification failed: %v", err)
		return
	}
	log.Println("Signature verified successfully")
}

// demonstrateKeyRotation demonstrates key rotation
func demonstrateKeyRotation(ks *keystore.FileKeyStore, keyID string) {
	// Get the original key metadata
	originalKey, err := ks.GetKeyMetadata(keyID)
	if err != nil {
		log.Printf("Failed to get original key metadata: %v", err)
		return
	}
	log.Printf("Original key: %s (%s)", originalKey.Name, originalKey.ID)

	// Rotate the key
	rotatedKey, err := ks.RotateKey(keyID)
	if err != nil {
		log.Printf("Failed to rotate key: %v", err)
		return
	}
	log.Printf("Rotated key: %s (%s)", rotatedKey.Metadata.Name, rotatedKey.Metadata.ID)

	// List all keys to see both the original and rotated keys
	keys, err := ks.ListKeys()
	if err != nil {
		log.Printf("Failed to list keys: %v", err)
		return
	}

	log.Printf("Found %d keys after rotation:", len(keys))
	for _, key := range keys {
		log.Printf("  - %s (%s): %s, %s, %s", key.Name, key.ID, key.Type, key.Usage, key.Algorithm)
	}
}

// demonstrateKeyExportImport demonstrates key export and import
func demonstrateKeyExportImport(ks *keystore.FileKeyStore, keyID string) {
	// Get the original key metadata
	originalKey, err := ks.GetKeyMetadata(keyID)
	if err != nil {
		log.Printf("Failed to get original key metadata: %v", err)
		return
	}
	log.Printf("Original key: %s (%s)", originalKey.Name, originalKey.ID)

	// Export the key in PEM format (public key only)
	exportedPEM, err := ks.ExportKey(keyID, "PEM", false)
	if err != nil {
		log.Printf("Failed to export key in PEM format: %v", err)
		return
	}
	log.Printf("Exported public key in PEM format (%d bytes)", len(exportedPEM))
	fmt.Println(string(exportedPEM))

	// Create metadata for imported key
	importMetadata := &keystore.KeyMetadata{
		Name:            "imported-" + originalKey.Name,
		Type:            originalKey.Type,
		Usage:           originalKey.Usage,
		ProtectionLevel: originalKey.ProtectionLevel,
		Algorithm:       originalKey.Algorithm,
		Tags:            append(originalKey.Tags, "imported"),
		Description:     "Imported copy of " + originalKey.Name,
		RotationPeriod:  originalKey.RotationPeriod,
	}

	// Import the key from PEM
	importedKey, err := ks.ImportKey(exportedPEM, "PEM", importMetadata)
	if err != nil {
		log.Printf("Failed to import key from PEM: %v", err)
		return
	}
	log.Printf("Imported key: %s (%s)", importedKey.Metadata.Name, importedKey.Metadata.ID)

	// List all keys to see both the original and imported keys
	keys, err := ks.ListKeys()
	if err != nil {
		log.Printf("Failed to list keys: %v", err)
		return
	}

	log.Printf("Found %d keys after import:", len(keys))
	for _, key := range keys {
		log.Printf("  - %s (%s): %s, %s, %s", key.Name, key.ID, key.Type, key.Usage, key.Algorithm)
	}
}
