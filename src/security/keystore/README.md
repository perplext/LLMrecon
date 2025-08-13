# Secure Key Storage Solution

This package provides a secure storage solution for cryptographic keys and sensitive materials. It supports various key types, protection levels, and operations for managing cryptographic keys throughout their lifecycle.

## Features

- **Multiple Key Types**: Support for RSA, ECDSA, Ed25519, and symmetric keys
- **Key Usage Definitions**: Define keys for signing, encryption, authentication, and TLS
- **Protection Levels**: Software-based protection with extensibility for HSM integration
- **Key Management**: Store, retrieve, delete, and list keys with rich metadata
- **Key Rotation**: Automatic key rotation based on configurable policies
- **Key Export/Import**: Export and import keys in various formats (PEM, DER)
- **Type-Specific Operations**: Retrieve keys in their native Go cryptographic types

## Usage

### Creating a Key Store

```go
import (
    "time"
    "github.com/perplext/LLMrecon/src/security/keystore"
)

// Create key store options
options := keystore.KeyStoreOptions{
    StoragePath:           "/path/to/keystore.json",
    Passphrase:            "secure-passphrase",
    AutoSave:              true,
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
```

### Generating Keys

```go
// Generate RSA key
rsaMetadata := &keystore.KeyMetadata{
    Name:            "my-rsa-signing-key",
    Type:            keystore.RSAKey,
    Usage:           keystore.SigningKey,
    ProtectionLevel: keystore.SoftwareProtection,
    Algorithm:       "RSA-2048",
    Tags:            []string{"application", "signing"},
    Description:     "RSA signing key for application",
    RotationPeriod:  90, // 90 days
}

rsaKey, err := ks.GenerateKey(keystore.RSAKey, "RSA-2048", rsaMetadata)
if err != nil {
    log.Fatalf("Failed to generate RSA key: %v", err)
}

// Generate ECDSA key
ecdsaMetadata := &keystore.KeyMetadata{
    Name:            "my-ecdsa-signing-key",
    Type:            keystore.ECDSAKey,
    Usage:           keystore.SigningKey,
    ProtectionLevel: keystore.SoftwareProtection,
    Algorithm:       "ECDSA-P256",
    Tags:            []string{"application", "signing"},
    Description:     "ECDSA signing key for application",
    RotationPeriod:  90, // 90 days
}

ecdsaKey, err := ks.GenerateKey(keystore.ECDSAKey, "ECDSA-P256", ecdsaMetadata)
if err != nil {
    log.Fatalf("Failed to generate ECDSA key: %v", err)
}

// Generate symmetric key
symmetricMetadata := &keystore.KeyMetadata{
    Name:            "my-encryption-key",
    Type:            keystore.SymmetricKey,
    Usage:           keystore.EncryptionKey,
    ProtectionLevel: keystore.SoftwareProtection,
    Algorithm:       "AES-256",
    Tags:            []string{"application", "encryption"},
    Description:     "Symmetric encryption key",
    RotationPeriod:  30, // 30 days
}

symmetricKey, err := ks.GenerateKey(keystore.SymmetricKey, "AES-256", symmetricMetadata)
if err != nil {
    log.Fatalf("Failed to generate symmetric key: %v", err)
}
```

### Retrieving Keys

```go
// Get key by ID
key, err := ks.GetKey("key-id")
if err != nil {
    log.Fatalf("Failed to get key: %v", err)
}

// Get key metadata only
metadata, err := ks.GetKeyMetadata("key-id")
if err != nil {
    log.Fatalf("Failed to get key metadata: %v", err)
}

// Get RSA private key
rsaPrivateKey, err := ks.GetRSAPrivateKey("key-id")
if err != nil {
    log.Fatalf("Failed to get RSA private key: %v", err)
}

// Get RSA public key
rsaPublicKey, err := ks.GetRSAPublicKey("key-id")
if err != nil {
    log.Fatalf("Failed to get RSA public key: %v", err)
}

// Get ECDSA private key
ecdsaPrivateKey, err := ks.GetECDSAPrivateKey("key-id")
if err != nil {
    log.Fatalf("Failed to get ECDSA private key: %v", err)
}

// Get ECDSA public key
ecdsaPublicKey, err := ks.GetECDSAPublicKey("key-id")
if err != nil {
    log.Fatalf("Failed to get ECDSA public key: %v", err)
}

// Get Ed25519 private key
ed25519PrivateKey, err := ks.GetEd25519PrivateKey("key-id")
if err != nil {
    log.Fatalf("Failed to get Ed25519 private key: %v", err)
}

// Get Ed25519 public key
ed25519PublicKey, err := ks.GetEd25519PublicKey("key-id")
if err != nil {
    log.Fatalf("Failed to get Ed25519 public key: %v", err)
}
```

### Listing Keys

```go
// List all keys
keys, err := ks.ListKeys()
if err != nil {
    log.Fatalf("Failed to list keys: %v", err)
}

// List keys by type
rsaKeys, err := ks.ListKeysByType(keystore.RSAKey)
if err != nil {
    log.Fatalf("Failed to list RSA keys: %v", err)
}

// List keys by usage
signingKeys, err := ks.ListKeysByUsage(keystore.SigningKey)
if err != nil {
    log.Fatalf("Failed to list signing keys: %v", err)
}

// List keys by tag
appKeys, err := ks.ListKeysByTag("application")
if err != nil {
    log.Fatalf("Failed to list application keys: %v", err)
}
```

### Key Rotation

```go
// Rotate a key
newKey, err := ks.RotateKey("key-id")
if err != nil {
    log.Fatalf("Failed to rotate key: %v", err)
}
```

### Key Export and Import

```go
// Export a key in PEM format (public key only)
pemData, err := ks.ExportKey("key-id", "PEM", false)
if err != nil {
    log.Fatalf("Failed to export key: %v", err)
}

// Export a key in DER format (with private key)
derData, err := ks.ExportKey("key-id", "DER", true)
if err != nil {
    log.Fatalf("Failed to export key: %v", err)
}

// Import a key from PEM format
importMetadata := &keystore.KeyMetadata{
    Name:            "imported-key",
    Type:            keystore.RSAKey,
    Usage:           keystore.SigningKey,
    ProtectionLevel: keystore.SoftwareProtection,
    Tags:            []string{"imported"},
    Description:     "Imported key",
    RotationPeriod:  90, // 90 days
}

importedKey, err := ks.ImportKey(pemData, "PEM", importMetadata)
if err != nil {
    log.Fatalf("Failed to import key: %v", err)
}
```

### Key Deletion

```go
// Delete a key
err := ks.DeleteKey("key-id")
if err != nil {
    log.Fatalf("Failed to delete key: %v", err)
}
```

## Security Considerations

1. **Passphrase Protection**: The key store is protected by a passphrase. Choose a strong, high-entropy passphrase.
2. **Key Rotation**: Regularly rotate keys according to your security policy. The `RotationPeriod` in key metadata defines how often keys should be rotated.
3. **Access Control**: Implement appropriate access controls to restrict who can access the key store.
4. **Audit Logging**: Enable audit logging to track key operations.
5. **Backup**: Regularly back up your key store to prevent data loss.
6. **HSM Integration**: For high-security environments, consider using hardware security modules (HSMs) for key protection.

## Environment Compatibility

The key storage solution is designed to work across different environments:
- Development
- Testing
- Production

Configuration options can be adjusted based on the environment to balance security and usability.

## Integration with Other Components

This key storage solution integrates with:
- The `vault` package for secure credential management
- The `crypto` package for cryptographic operations
- The `x509` package for certificate handling

## Example

See the [examples](examples/keystore_example.go) directory for a complete example of how to use the key storage solution.
