# Provider Configuration Management

This document describes the configuration management system for the Multi-Provider LLM Integration Framework, including how to securely store and retrieve API keys, manage configuration versions, and customize provider settings.

## Overview

The configuration management system provides a secure way to manage provider configurations, including API keys, endpoints, and other provider-specific settings. It supports:

- Secure storage of API keys with encryption
- Loading configurations from environment variables, files, and runtime parameters
- Validation of configuration parameters
- Configuration versioning to track changes
- Rollback to previous configurations

## Configuration Structure

Each provider configuration includes the following fields:

```go
type ProviderConfig struct {
    // Type is the type of provider
    Type ProviderType
    // APIKey is the API key for the provider
    APIKey string
    // OrgID is the organization ID for the provider
    OrgID string
    // BaseURL is the base URL for the provider API
    BaseURL string
    // Timeout is the timeout for requests to the provider
    Timeout time.Duration
    // RetryConfig is the configuration for retries
    RetryConfig *RetryConfig
    // RateLimitConfig is the configuration for rate limiting
    RateLimitConfig *RateLimitConfig
    // DefaultModel is the default model to use
    DefaultModel string
    // AdditionalHeaders is a map of additional headers to include in requests
    AdditionalHeaders map[string]string
    // AdditionalParams is a map of additional parameters to include in requests
    AdditionalParams map[string]interface{}
}
```

## Usage

### Creating a Configuration Manager

```go
import (
    "github.com/perplext/LLMrecon/src/provider/config"
)

// Create a configuration manager with default settings
configManager, err := config.NewConfigManager("", nil, "")

// Create a configuration manager with custom settings
configFile := "/path/to/config.json"
encryptionKey := []byte("your-encryption-key")
envVarPrefix := "LLM_RED_TEAM"
configManager, err := config.NewConfigManager(configFile, encryptionKey, envVarPrefix)
```

### Setting Provider Configurations

```go
// Create a provider configuration
providerConfig := &core.ProviderConfig{
    Type:    core.OpenAIProvider,
    APIKey:  "your-api-key",
    BaseURL: "https://api.openai.com",
    Timeout: 30 * time.Second,
    RetryConfig: &core.RetryConfig{
        MaxRetries:           3,
        InitialBackoff:       1 * time.Second,
        MaxBackoff:           60 * time.Second,
        BackoffMultiplier:    2.0,
        RetryableStatusCodes: []int{429, 500, 502, 503, 504},
    },
    RateLimitConfig: &core.RateLimitConfig{
        RequestsPerMinute:    60,
        TokensPerMinute:      100000,
        MaxConcurrentRequests: 10,
        BurstSize:            5,
    },
    DefaultModel: "gpt-4",
}

// Set the configuration
err := configManager.SetConfig(core.OpenAIProvider, providerConfig)
```

### Getting Provider Configurations

```go
// Get a provider configuration
providerConfig, err := configManager.GetConfig(core.OpenAIProvider)

// Get all provider configurations
allConfigs := configManager.GetAllConfigs()

// Get all provider types
allProviderTypes := configManager.GetAllProviderTypes()
```

### Updating Provider Configurations

```go
// Create an update configuration
updates := &core.ProviderConfig{
    APIKey:  "updated-api-key",
    Timeout: 60 * time.Second,
}

// Update the configuration
err := configManager.UpdateConfig(core.OpenAIProvider, updates)
```

### Deleting Provider Configurations

```go
// Delete a provider configuration
err := configManager.DeleteConfig(core.OpenAIProvider)
```

### Setting Provider API Keys

```go
// Set a provider API key
err := configManager.SetProviderAPIKey(core.OpenAIProvider, "your-api-key")
```

## Environment Variables

The configuration manager supports loading configurations from environment variables. The environment variables are in the format `<PREFIX>_<PROVIDER>_<FIELD>`, where:

- `<PREFIX>` is the environment variable prefix (default: `LLM_RED_TEAM`)
- `<PROVIDER>` is the provider type (e.g., `OPENAI`, `ANTHROPIC`)
- `<FIELD>` is the configuration field (e.g., `APIKEY`, `BASEURL`)

For example:

```
LLM_RED_TEAM_OPENAI_APIKEY=your-api-key
LLM_RED_TEAM_OPENAI_BASEURL=https://api.openai.com
LLM_RED_TEAM_ANTHROPIC_APIKEY=your-anthropic-api-key
```

## Configuration Versioning

The configuration manager supports versioning of configurations, allowing you to track changes and roll back to previous versions.

### Getting Configuration History

```go
// Get the configuration history for a provider
history, err := configManager.GetConfigHistory(core.OpenAIProvider)

// Get a specific version of a configuration
version, err := configManager.GetConfigVersion(core.OpenAIProvider, 1)

// Get all versions of a configuration
versions, err := configManager.GetConfigVersions(core.OpenAIProvider)

// Get all configuration history
allHistory := configManager.GetAllConfigHistory()
```

### Rolling Back Configurations

```go
// Roll back to a specific version
err := configManager.RollbackConfig(core.OpenAIProvider, 1)
```

## Security

### Encryption

The configuration manager supports encryption of sensitive data, such as API keys. The encryption is performed using AES-256-GCM, which provides both confidentiality and integrity.

To enable encryption, provide an encryption key when creating the configuration manager:

```go
encryptionKey := []byte("your-encryption-key")
configManager, err := config.NewConfigManager("", encryptionKey, "")
```

### Validation

The configuration manager validates configurations before storing them, ensuring that required fields are present and that the configuration is valid.

## Best Practices

1. **Use Environment Variables**: Store API keys in environment variables to avoid hardcoding them in your code.
2. **Enable Encryption**: Always enable encryption by providing an encryption key to protect sensitive data.
3. **Use Version Control**: Use the configuration versioning system to track changes and roll back if needed.
4. **Validate Configurations**: Always validate configurations before using them to ensure they are valid.
5. **Use Default Values**: Provide default values for optional fields to ensure your application works correctly even if the configuration is incomplete.

## Example: Complete Configuration Management

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/perplext/LLMrecon/src/provider/config"
    "github.com/perplext/LLMrecon/src/provider/core"
    "github.com/perplext/LLMrecon/src/provider/factory"
)

func main() {
    // Create a configuration manager
    configManager, err := config.NewConfigManager("", nil, "LLM_RED_TEAM")
    if err != nil {
        log.Fatalf("Failed to create configuration manager: %v", err)
    }

    // Set API keys from environment variables
    // LLM_RED_TEAM_OPENAI_APIKEY and LLM_RED_TEAM_ANTHROPIC_APIKEY should be set

    // Create a provider factory
    providerFactory := factory.NewProviderFactory(configManager)

    // Get a provider
    provider, err := providerFactory.GetProvider(core.OpenAIProvider)
    if err != nil {
        log.Fatalf("Failed to get provider: %v", err)
    }

    // Use the provider
    ctx := context.Background()
    models, err := provider.GetModels(ctx)
    if err != nil {
        log.Fatalf("Failed to get models: %v", err)
    }

    fmt.Printf("Available models: %v\n", models)

    // Close the provider
    err = provider.Close()
    if err != nil {
        log.Fatalf("Failed to close provider: %v", err)
    }
}
```

## Troubleshooting

### Common Issues

1. **Configuration Not Found**: Ensure that the configuration file exists and is readable.
2. **Invalid Configuration**: Ensure that the configuration is valid and contains all required fields.
3. **Encryption Key Mismatch**: Ensure that the same encryption key is used for encryption and decryption.
4. **Environment Variables Not Set**: Ensure that the environment variables are set correctly.

### Debugging

To debug configuration issues, you can:

1. Check the configuration file to ensure it exists and contains the expected configurations.
2. Check the environment variables to ensure they are set correctly.
3. Use the `GetConfig` method to retrieve the configuration and inspect it.
4. Use the `GetConfigHistory` method to view the configuration history and track changes.

## Contributing

To contribute to the configuration management system, please follow these guidelines:

1. Add tests for any new functionality.
2. Ensure that all existing tests pass.
3. Update the documentation to reflect any changes.
4. Follow the existing code style and conventions.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
