// Package main provides an example of using the Multi-Provider LLM Integration Framework.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/perplext/LLMrecon/src/provider/anthropic"
	"github.com/perplext/LLMrecon/src/provider/config"
	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/provider/factory"
	"github.com/perplext/LLMrecon/src/provider/openai"
	"github.com/perplext/LLMrecon/src/provider/registry"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a configuration manager
	configManager, err := config.NewConfigManager("", nil, "LLM_RED_TEAM")
	if err != nil {
		fmt.Printf("Failed to create configuration manager: %v\n", err)
		os.Exit(1)
	}

	// Load configurations from environment variables
	if err := configManager.LoadFromEnv(); err != nil {
		fmt.Printf("Failed to load configurations from environment variables: %v\n", err)
		os.Exit(1)
	}

	// Create a provider factory
	providerFactory := factory.NewProviderFactory(configManager)

	// Register provider constructors
	providerFactory.RegisterProvider(core.OpenAIProvider, openai.NewOpenAIProvider)
	providerFactory.RegisterProvider(core.AnthropicProvider, anthropic.NewAnthropicProvider)

	// Create a provider registry
	providerRegistry := registry.NewProviderRegistry()

	// Create a model registry
	modelRegistry := registry.NewModelRegistry()

	// Get OpenAI provider
	openaiProvider, err := providerFactory.GetProvider(core.OpenAIProvider)
	if err != nil {
		fmt.Printf("Failed to get OpenAI provider: %v\n", err)
		os.Exit(1)
	}

	// Get Anthropic provider
	anthropicProvider, err := providerFactory.GetProvider(core.AnthropicProvider)
	if err != nil {
		fmt.Printf("Failed to get Anthropic provider: %v\n", err)
		os.Exit(1)
	}

	// Register providers
	providerRegistry.RegisterProvider(openaiProvider)
	providerRegistry.RegisterProvider(anthropicProvider)

	// Sync models from providers
	if err := modelRegistry.SyncModelsFromProviders([]core.Provider{openaiProvider, anthropicProvider}); err != nil {
		fmt.Printf("Failed to sync models from providers: %v\n", err)
		os.Exit(1)
	}

	// Get all models
	models := modelRegistry.GetAllModels()
	fmt.Printf("Available models: %d\n", len(models))
	for _, model := range models {
		fmt.Printf("- %s (%s): %s\n", model.ID, model.Provider, model.Type)
	}

	// Example: Chat completion with OpenAI
	fmt.Println("\nChat completion with OpenAI:")
	openaiChatResponse, err := openaiProvider.ChatCompletion(ctx, &core.ChatCompletionRequest{
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
	if err != nil {
		fmt.Printf("Failed to generate chat completion with OpenAI: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", openaiChatResponse.Choices[0].Message.Content)
	}

	// Example: Chat completion with Anthropic
	fmt.Println("\nChat completion with Anthropic:")
	anthropicChatResponse, err := anthropicProvider.ChatCompletion(ctx, &core.ChatCompletionRequest{
		Model: "claude-3-sonnet-20240229",
		Messages: []core.Message{
			{
				Role:    "user",
				Content: "Hello, who are you?",
			},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	})
	if err != nil {
		fmt.Printf("Failed to generate chat completion with Anthropic: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", anthropicChatResponse.Choices[0].Message.Content)
	}

	// Example: Get provider by capability
	fmt.Println("\nGet provider by capability:")
	embeddingProvider, err := providerRegistry.GetProviderByCapability(core.EmbeddingCapability)
	if err != nil {
		fmt.Printf("Failed to get provider by capability: %v\n", err)
	} else {
		fmt.Printf("Provider for embedding: %s\n", embeddingProvider.GetType())
	}

	// Example: Get provider by model
	fmt.Println("\nGet provider by model:")
	claudeProvider, err := providerRegistry.GetProviderByModel("claude-3-opus-20240229")
	if err != nil {
		fmt.Printf("Failed to get provider by model: %v\n", err)
	} else {
		fmt.Printf("Provider for claude-3-opus-20240229: %s\n", claudeProvider.GetType())
	}

	// Close all providers
	if err := providerRegistry.CloseAllProviders(); err != nil {
		fmt.Printf("Failed to close all providers: %v\n", err)
	}
}
