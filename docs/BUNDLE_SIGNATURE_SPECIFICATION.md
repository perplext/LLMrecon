# Bundle Digital Signature and Integrity Specification

## Overview

This document specifies the digital signature and integrity mechanism for LLMrecon offline bundles. The system ensures authenticity, integrity, and non-repudiation of bundle contents.

## Signature Architecture

### 1. Cryptographic Standards

- **Signature Algorithm**: Ed25519 (RFC 8032)
  - 128-bit security level
  - Fast signature generation and verification
  - Small key and signature sizes (32 and 64 bytes respectively)
  
- **Hash Algorithm**: SHA-256
  - Used for content digest generation
  - Industry standard with broad support

- **Encoding**: Base64 URL-safe encoding for signatures and keys

### 2. Key Management

#### Key Structure
```json
{
  "keyId": "llm-redteam-2024-01",
  "algorithm": "Ed25519",
  "publicKey": "base64url-encoded-public-key",
  "created": "2024-01-01T00:00:00Z",
  "expires": "2025-01-01T00:00:00Z",
  "usage": ["bundle-signing"]
}
```

#### Key Storage
- Private keys: Hardware Security Module (HSM) or secure key vault
- Public keys: Distributed with tool and available at https://keys.llm-redteam.io/

### 3. Signature Process

#### 3.1 Content Preparation
1. Generate SHA-256 hash for each file in the bundle
2. Create content manifest with file paths and hashes
3. Sort manifest entries by file path for deterministic ordering
4. Generate canonical JSON representation of manifest

#### 3.2 Signature Generation
```go
type BundleSignature struct {
    Version     string            `json:"version"`
    Algorithm   string            `json:"algorithm"`
    KeyID       string            `json:"keyId"`
    Timestamp   time.Time         `json:"timestamp"`
    ContentHash string            `json:"contentHash"`
    Signature   string            `json:"signature"`
    Metadata    SignatureMetadata `json:"metadata"`
}

type SignatureMetadata struct {
    Signer      string   `json:"signer"`
    Environment string   `json:"environment"`
    BuildID     string   `json:"buildId"`
    Tags        []string `json:"tags"`
}
```

#### 3.3 Signature File Location
```
bundle/
  ├── manifest.json
  ├── signatures/
  │   ├── bundle.sig      # Primary signature
  │   └── bundle.sig.asc  # Optional PGP signature
  └── ...
```

### 4. Verification Process

#### 4.1 Verification Steps
1. Extract signature file from bundle
2. Validate signature format and version
3. Retrieve public key for specified keyId
4. Verify key validity (not expired, not revoked)
5. Compute content hash of bundle files
6. Verify signature against computed hash
7. Validate timestamp is within acceptable range

#### 4.2 Verification API
```go
func VerifyBundle(bundlePath string) (*VerificationResult, error) {
    // Load bundle and signature
    bundle, err := LoadBundle(bundlePath)
    if err != nil {
        return nil, err
    }
    
    // Verify signature
    result := &VerificationResult{
        Valid:     false,
        Timestamp: time.Now(),
        Details:   make(map[string]interface{}),
    }
    
    // Implementation details...
    return result, nil
}
```

### 5. Integrity Verification

#### 5.1 File-Level Integrity
Each file in the bundle includes:
- SHA-256 hash
- File size
- Permissions (Unix mode)
- Optional: SHA-512 hash for critical files

#### 5.2 Manifest Integrity
```json
{
  "manifestVersion": "1.0",
  "bundleId": "unique-bundle-identifier",
  "created": "2024-01-15T10:00:00Z",
  "files": [
    {
      "path": "binary/llm-redteam",
      "hash": "sha256:abcdef...",
      "size": 10485760,
      "mode": "0755"
    }
  ],
  "contentHash": "sha256:123456...",
  "integrityAlgorithm": "sha256"
}
```

### 6. Security Considerations

#### 6.1 Key Rotation
- Regular key rotation schedule (annual)
- Overlapping validity periods for smooth transitions
- Revocation list for compromised keys

#### 6.2 Time-based Validation
- Signatures include timestamp
- Bundles expire after specified period (default: 90 days)
- Clock skew tolerance: ±5 minutes

#### 6.3 Offline Verification
- Public keys bundled with tool installation
- Offline revocation checking via bundled CRL
- Fallback to online verification when available

### 7. Implementation Example

```go
package signature

import (
    "crypto/ed25519"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "time"
)

// SignBundle creates a digital signature for the bundle
func SignBundle(bundlePath string, privateKey ed25519.PrivateKey) (*BundleSignature, error) {
    // Calculate content hash
    contentHash, err := calculateBundleHash(bundlePath)
    if err != nil {
        return nil, err
    }
    
    // Create signature object
    sig := &BundleSignature{
        Version:     "1.0",
        Algorithm:   "Ed25519",
        KeyID:       getCurrentKeyID(),
        Timestamp:   time.Now().UTC(),
        ContentHash: contentHash,
        Metadata: SignatureMetadata{
            Signer:      "LLMrecon Build System",
            Environment: "production",
            BuildID:     getBuildID(),
        },
    }
    
    // Sign the content
    message := sig.getSigningMessage()
    signature := ed25519.Sign(privateKey, message)
    sig.Signature = base64.URLEncoding.EncodeToString(signature)
    
    return sig, nil
}

// VerifyBundleSignature verifies the bundle signature
func VerifyBundleSignature(bundlePath string, signature *BundleSignature) error {
    // Get public key
    publicKey, err := getPublicKey(signature.KeyID)
    if err != nil {
        return err
    }
    
    // Verify timestamp
    if err := verifyTimestamp(signature.Timestamp); err != nil {
        return err
    }
    
    // Calculate current hash
    currentHash, err := calculateBundleHash(bundlePath)
    if err != nil {
        return err
    }
    
    // Verify content hasn't changed
    if currentHash != signature.ContentHash {
        return ErrContentModified
    }
    
    // Verify signature
    sigBytes, err := base64.URLEncoding.DecodeString(signature.Signature)
    if err != nil {
        return err
    }
    
    message := signature.getSigningMessage()
    if !ed25519.Verify(publicKey, message, sigBytes) {
        return ErrInvalidSignature
    }
    
    return nil
}
```

### 8. CLI Integration

```bash
# Sign a bundle
llm-redteam bundle sign --bundle bundle.tar.gz --key-id production-2024

# Verify a bundle
llm-redteam bundle verify --bundle bundle.tar.gz

# Extract signature information
llm-redteam bundle signature --bundle bundle.tar.gz --format json
```

### 9. Error Handling

Common verification errors:
- `SIGNATURE_NOT_FOUND`: No signature file in bundle
- `INVALID_SIGNATURE_FORMAT`: Malformed signature file
- `KEY_NOT_FOUND`: Unknown key ID
- `KEY_EXPIRED`: Signing key has expired
- `SIGNATURE_EXPIRED`: Bundle signature too old
- `CONTENT_MODIFIED`: Bundle contents don't match signature
- `INVALID_SIGNATURE`: Cryptographic verification failed

## Testing Requirements

1. Unit tests for signature generation and verification
2. Integration tests with various bundle formats
3. Performance tests for large bundles
4. Security tests including tampered bundles
5. Compatibility tests with different key formats