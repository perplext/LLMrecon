# Downgrade Attack Protection

This document provides an overview of the downgrade attack protection mechanisms implemented in the LLMrecon project. Downgrade attacks attempt to force the system to use weaker cryptographic algorithms or parameters, potentially compromising security.

## Overview

The downgrade protection system provides the following security features:

1. **Security Policy Management**: Centralized management of security requirements
2. **TLS Version Enforcement**: Ensures minimum TLS version requirements
3. **Cipher Suite Restrictions**: Limits allowed cipher suites to secure options
4. **Signature Algorithm Validation**: Enforces use of strong signature algorithms
5. **Key Size Requirements**: Ensures cryptographic keys meet minimum size requirements
6. **Certificate Pinning Enforcement**: Requires certificate pinning for secure connections
7. **Revocation Checking**: Enforces certificate revocation checking
8. **Version Validation**: Prevents downgrade to vulnerable versions of components
9. **Policy Signing**: Cryptographically signs security policies to prevent tampering
10. **Secure Client Creation**: Creates HTTP clients that comply with security policies

## Security Policy

The security policy is a central component that defines minimum security requirements:

```go
type SecurityPolicy struct {
    // Minimum allowed TLS version
    MinimumTLSVersion uint16
    
    // List of allowed TLS cipher suites
    AllowedCipherSuites []uint16
    
    // List of allowed signature algorithms
    AllowedSignatureAlgorithms []SignatureAlgorithm
    
    // Minimum key sizes by algorithm type
    MinimumKeySize map[string]int
    
    // Whether certificate pinning is required
    RequireCertificatePinning bool
    
    // Whether certificate revocation checking is required
    RequireRevocationCheck bool
    
    // Minimum allowed versions by component type
    MinimumVersions map[string]string
    
    // When the policy was last updated
    LastUpdateTime time.Time
    
    // Version of the security policy
    PolicyVersion string
    
    // Cryptographic signature of the policy
    PolicySignature string
    
    // Whether signature verification is required
    RequireSignatureVerification bool
}
```

## Usage Examples

### Initializing Downgrade Protection

```go
import (
    "github.com/perplext/LLMrecon/src/security/keystore"
    "github.com/perplext/LLMrecon/src/update"
)

// Create key store
keyStoreOptions := keystore.KeyStoreOptions{
    StoragePath:           "/path/to/keystore.json",
    Passphrase:            "secure-passphrase",
    AutoSave:              true,
    RotationCheckInterval: time.Hour * 24,
}
ks, err := keystore.NewFileKeyStore(keyStoreOptions)
if err != nil {
    log.Fatalf("Failed to create key store: %v", err)
}
defer ks.Close()

// Initialize downgrade protection
dp, err := update.NewDowngradeProtection("/path/to/security_policy.json", ks)
if err != nil {
    log.Fatalf("Failed to initialize downgrade protection: %v", err)
}
```

### Validating Connection Security

```go
// Create connection security options
options := update.DefaultConnectionSecurityOptions()

// Validate against security policy
if err := dp.ValidateConnectionSecurity(options); err != nil {
    log.Fatalf("Connection security validation failed: %v", err)
}
```

### Enforcing Security Policy

```go
// Create connection security options (potentially with weak settings)
options := &update.ConnectionSecurityOptions{
    MinTLSVersion:          tls.VersionTLS10, // Too low
    EnableCertificatePinning: false,          // Should be enabled
    CheckRevocation:        false,            // Should be enabled
}

// Enforce the security policy (modifies options to comply with policy)
dp.EnforceSecurityPolicy(options)

// Now options will comply with the security policy
```

### Creating a Secure Client

```go
// Create a secure HTTP client that complies with the security policy
client, err := dp.CreateSecureClient()
if err != nil {
    log.Fatalf("Failed to create secure client: %v", err)
}

// Use the client for secure HTTP requests
resp, err := client.Get(context.Background(), "https://example.com", nil)
```

### Validating Update Packages

```go
// Validate that an update package meets security requirements
if err := dp.ValidateUpdatePackage(updatePackage); err != nil {
    log.Fatalf("Update package validation failed: %v", err)
}
```

### Signing and Verifying Security Policies

```go
// Sign the security policy
if err := dp.SignPolicy("policy-signing-key-id"); err != nil {
    log.Fatalf("Failed to sign policy: %v", err)
}

// Policy verification happens automatically when loading the policy
```

## Best Practices

1. **Regular Policy Updates**: Regularly update the security policy to address new vulnerabilities and threats
2. **Key Rotation**: Rotate policy signing keys periodically
3. **Monitoring**: Monitor for attempts to bypass security policies
4. **Secure Storage**: Store security policies in secure locations with appropriate access controls
5. **Default Deny**: Configure policies to deny by default and explicitly allow only secure options
6. **Testing**: Regularly test that downgrade protection is working as expected

## Example Implementation

See the [downgrade_example](downgrade_example/main.go) directory for a complete example of how to use the downgrade protection system.

## Security Considerations

1. **Policy Integrity**: The security policy itself must be protected from tampering. Use the policy signing and verification features.
2. **Key Management**: Properly manage the keys used for policy signing and verification.
3. **Secure Defaults**: The default security policy provides reasonable security, but should be reviewed and customized for your specific requirements.
4. **Compatibility**: Be aware that enforcing strict security requirements may impact compatibility with older systems.
5. **Updates**: Keep the security policy updated as new vulnerabilities are discovered and security best practices evolve.

## References

1. [OWASP TLS Cipher String Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/TLS_Cipher_String_Cheat_Sheet.html)
2. [NIST SP 800-52 Rev. 2: Guidelines for TLS Implementations](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-52r2.pdf)
3. [NIST SP 800-57: Recommendation for Key Management](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-57pt1r5.pdf)
