# Secure Credential Management

This document outlines the secure credential management system implemented in the LLMrecon Tool. The system provides a secure way to store, retrieve, and manage credentials such as API keys.

## Overview

The secure credential management system consists of the following components:

1. **SecureVault**: Provides encryption and secure storage of credentials
2. **CredentialManager**: Manages credential operations (add, update, retrieve, delete)
3. **Audit Logging**: Tracks all credential access and operations
4. **Rotation Policies**: Enforces credential rotation for enhanced security
5. **Integration with Configuration**: Seamlessly works with the existing configuration system

## Using the Credential Manager

### Setting up the Credential Manager

The credential manager can be initialized with the following code:

```go
manager, err := vault.NewCredentialManager(vault.ManagerOptions{
    ConfigDir:     "/path/to/config/dir",
    Passphrase:    "your-secure-passphrase",
    EnvPrefix:     "LLMRT",
    AutoSave:      true,
    InstallGitHook: true,
})
```

Or you can use the default manager:

```go
err := vault.InitDefaultManager("/path/to/config/dir", "your-secure-passphrase", "LLMRT")
```

### Managing API Keys

```go
// Set an API key
err := manager.SetAPIKey("openai", "your-api-key", "OpenAI API Key")

// Get an API key
apiKey, err := manager.GetAPIKey("openai")

// List all credentials
credentials, err := manager.ListCredentials()

// List credentials for a specific service
credentials, err := manager.ListCredentialsByService("openai")

// Delete a credential
err := manager.DeleteCredential("credential-id")
```

### Environment Variables

The credential manager can load API keys from environment variables. The environment variables should be named with the following pattern:

```
{ENV_PREFIX}_{SERVICE}_API_KEY
```

For example, if the environment prefix is `LLMRT`, the environment variable for the OpenAI API key would be `LLMRT_OPENAI_API_KEY`.

## Security Features

### Encryption

Credentials are encrypted using AES-GCM with a key derived from the passphrase using scrypt. This ensures that credentials are secure both at rest and in transit.

### Rotation Policies

Credentials can have rotation policies that specify how often they should be rotated. The credential manager can identify credentials that need rotation:

```go
// Get credentials that need rotation
credentials, err := manager.GetCredentialsNeedingRotation()

// Rotate a credential
err := manager.RotateCredential("credential-id", "new-value")
```

### Audit Logging

All credential operations are logged for security and compliance purposes. The audit logs include:

- Timestamp
- Event type (access, create, update, delete, rotate)
- Credential ID
- Service
- User ID
- Source IP
- Success/failure status
- Error message (if applicable)
- Additional metadata

```go
// Get audit events
events, err := auditLogger.GetAuditEvents(100, map[string]string{
    "event_type": "access",
})
```

### Anomaly Detection

The system can detect anomalous access patterns and unused credentials:

```go
// Get credentials with anomalous access patterns
anomalous, err := auditIntegration.GetCredentialsWithAnomalousAccess(10)

// Get unused credentials
unused, err := auditIntegration.GetUnusedCredentials(30)
```

### Git Integration

The credential manager can install git hooks to prevent committing sensitive files and update the `.gitignore` file to exclude credential files.

## Command-Line Interface

The credential management system includes a CLI tool for managing credentials:

```
# List all credentials
LLMrecon credentials list

# Add a credential
LLMrecon credentials add --service openai --value your-api-key --description "OpenAI API Key"

# Get a credential
LLMrecon credentials get --service openai

# Delete a credential
LLMrecon credentials delete --id credential-id

# Rotate a credential
LLMrecon credentials rotate --id credential-id --value new-api-key
```

## Best Practices

1. **Use a strong passphrase**: The security of the credential vault depends on the strength of the passphrase.
2. **Rotate credentials regularly**: Set up rotation policies for all credentials and follow the rotation recommendations.
3. **Monitor audit logs**: Regularly review the audit logs for suspicious activity.
4. **Use environment variables in production**: In production environments, use environment variables to provide credentials.
5. **Backup the credential vault**: Regularly backup the credential vault to prevent data loss.
6. **Never commit credentials to version control**: Use the provided git hooks to prevent accidentally committing credentials.

## Environment Variables

The following environment variables are used by the credential management system:

- `LLMRT_VAULT_PASSPHRASE`: The passphrase used to encrypt the credential vault. If not set, the passphrase must be provided programmatically.
- `LLMRT_{SERVICE}_API_KEY`: API keys for various services. These will be automatically loaded by the credential manager.

## File Locations

- **Credential Vault**: `{ConfigDir}/credentials.vault`
- **Audit Log**: `{ConfigDir}/credential-audit.log`

## Implementation Details

The secure credential management system is implemented in the following packages:

- `github.com/perplext/LLMrecon/src/security/vault`: Core vault and credential manager implementation
- `github.com/perplext/LLMrecon/src/security/audit`: Audit logging implementation
- `github.com/perplext/LLMrecon/src/cmd/credentials.go`: CLI implementation

## Security Considerations

1. **Memory Safety**: Credentials are only decrypted when needed and are not stored in memory longer than necessary.
2. **Encryption**: AES-GCM is used for encryption, which provides both confidentiality and integrity.
3. **Key Derivation**: scrypt is used for key derivation, which is designed to be resistant to hardware acceleration attacks.
4. **Audit Logging**: All operations are logged for security and compliance purposes.
5. **Rotation Policies**: Credentials can have rotation policies to ensure they are rotated regularly.
6. **Git Integration**: Git hooks prevent accidentally committing sensitive files.

## Troubleshooting

### Common Issues

1. **"Failed to decrypt credential vault"**: This usually means the passphrase is incorrect. Make sure you're using the correct passphrase.
2. **"Credential not found"**: The requested credential doesn't exist in the vault. Check the service name or credential ID.
3. **"Failed to save credential vault"**: This could be due to permission issues. Make sure the process has write access to the config directory.

### Resetting the Credential Vault

If you need to reset the credential vault, you can delete the `credentials.vault` file in the config directory. This will remove all stored credentials.

## Contributing

When contributing to the credential management system, please follow these guidelines:

1. **Security First**: Always prioritize security over convenience.
2. **Test Thoroughly**: All changes must be thoroughly tested, especially encryption and decryption functionality.
3. **Audit Logging**: All credential operations must be logged for security and compliance purposes.
4. **Documentation**: Update this documentation when making changes to the credential management system.

## License

The credential management system is part of the LLMrecon Tool and is licensed under the same license as the rest of the project.
