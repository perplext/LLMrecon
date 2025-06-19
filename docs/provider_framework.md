# Multi-Provider LLM Integration Framework

This framework provides a unified interface for interacting with multiple LLM providers, including OpenAI, Anthropic, and others. It's designed to be extensible, resilient, and configurable.

## Features

### Core Interface Design
- Abstract `Provider` interface defining standard methods for text completion, chat, embeddings, and other common LLM operations
- Provider-specific implementations (e.g., `OpenAIProvider`, `AnthropicProvider`) that inherit from the base interface
- Factory pattern for instantiating appropriate provider instances

### Configuration Management
- Secure configuration system for storing and retrieving API keys
- Support for environment variables and configuration files
- Encryption for sensitive credentials
- Validation system for configuration parameters

### Request Handling
- Middleware for handling authentication headers and request formatting
- Proper error handling with meaningful error messages
- Request/response logging with PII redaction

### Rate Limiting and Resilience
- Configurable rate limiting based on provider specifications
- Retry mechanisms with exponential backoff for transient failures
- Circuit breaker pattern to prevent cascading failures

### Model Management
- Registry of available models for each provider
- Support for versioning and compatibility checks

### Extensibility
- Plugin architecture for adding new providers without modifying core code
- Documentation templates for provider implementation

### Performance Optimization
- Connection pooling where applicable
- Caching mechanisms for frequently used responses

## Usage

### Environment Variables

Set the following environment variables to configure the providers:

```
# OpenAI
LLM_RED_TEAM_OPENAI_APIKEY=your_openai_api_key
LLM_RED_TEAM_OPENAI_ORGID=your_openai_org_id (optional)

# Anthropic
LLM_RED_TEAM_ANTHROPIC_APIKEY=your_anthropic_api_key
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/perplext/LLMrecon/src/provider/config"
    "github.com/perplext/LLMrecon/src/provider/core"
    "github.com/perplext/LLMrecon/src/provider/factory"
    "github.com/perplext/LLMrecon/src/provider/openai"
)

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Create a configuration manager
    configManager, _ := config.NewConfigManager("", nil, "LLM_RED_TEAM")

    // Create a provider factory
    providerFactory := factory.NewProviderFactory(configManager)

    // Register provider constructors
    providerFactory.RegisterProvider(core.OpenAIProvider, openai.NewOpenAIProvider)

    // Get OpenAI provider
    openaiProvider, _ := providerFactory.GetProvider(core.OpenAIProvider)

    // Chat completion with OpenAI
    response, _ := openaiProvider.ChatCompletion(ctx, &core.ChatCompletionRequest{
        Model: "gpt-3.5-turbo",
        Messages: []core.Message{
            {
                Role:    "system",
                Content: "You are a helpful assistant.",
            },
            {
                Role:    "user",
                Content: "Hello, who are you?",
            },
        },
        MaxTokens:   100,
        Temperature: 0.7,
    })

    fmt.Printf("Response: %s\n", response.Choices[0].Message.Content)
}
```

## Architecture

The framework is organized into the following packages:

- `core`: Core interfaces and types
- `config`: Configuration management
- `factory`: Provider factory
- `middleware`: Middleware components (rate limiting, retry, logging)
- `openai`: OpenAI provider implementation
- `anthropic`: Anthropic provider implementation
- `registry`: Provider and model registry

## Adding a New Provider

To add a new provider:

1. Create a new package for the provider
2. Implement the `Provider` interface
3. Register the provider constructor with the provider factory

Example:

```go
// Register provider constructor
providerFactory.RegisterProvider(core.CustomProvider, custom.NewCustomProvider)
```

## Error Handling

The framework provides standardized error handling through the `ProviderError` type, which includes:

- Status code
- Error type
- Error message
- Additional parameters
- Raw response

## Rate Limiting

Rate limiting is configurable per provider:

```go
// Configure rate limiting
provider.GetConfig().RateLimitConfig = &core.RateLimitConfig{
    RequestsPerMinute:     60,
    TokensPerMinute:       100000,
    MaxConcurrentRequests: 10,
    BurstSize:             10,
}
```

## Retry Mechanism

Retry behavior is configurable per provider:

```go
// Configure retry
provider.GetConfig().RetryConfig = &core.RetryConfig{
    MaxRetries:          3,
    InitialBackoff:      1 * time.Second,
    MaxBackoff:          60 * time.Second,
    BackoffMultiplier:   2.0,
    RetryableStatusCodes: []int{429, 500, 502, 503, 504},
}
```

## Logging

The framework provides configurable logging with PII redaction:

```go
// Configure logging
loggingMiddleware := middleware.NewLoggingMiddleware(middleware.LogLevelInfo, true)
loggingMiddleware.AddHandler(middleware.LogLevelInfo, middleware.ConsoleLogHandler())
```

## Security

- API keys are stored securely
- Sensitive information is redacted in logs
- Configuration files can be encrypted

## Examples

See the `examples/provider` directory for complete examples of using the framework.
