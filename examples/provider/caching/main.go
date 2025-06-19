// Package main provides an example of using the caching system with the Multi-Provider LLM Integration Framework.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/perplext/LLMrecon/src/provider/cache"
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

	// Set API keys (in a real application, these would come from environment variables or secure storage)
	err = configManager.SetProviderAPIKey(core.OpenAIProvider, "your-openai-api-key")
	if err != nil {
		log.Fatalf("Failed to set OpenAI API key: %v", err)
	}

	// Create a provider factory
	providerFactory := factory.NewProviderFactory(configManager)

	// Create a provider cache
	providerCache := cache.NewProviderCache(1*time.Hour, 1000, cache.LRU)

	// Get a provider
	provider, err := providerFactory.GetProvider(core.OpenAIProvider)
	if err != nil {
		log.Fatalf("Failed to get provider: %v", err)
	}

	// Wrap the provider with caching
	cachingProvider := cache.NewCachingProvider(provider, providerCache)

	// Use the caching provider
	ctx := context.Background()

	// Example 1: Get models (will be cached)
	fmt.Println("Getting models (first call)...")
	startTime := time.Now()
	models, err := cachingProvider.GetModels(ctx)
	if err != nil {
		log.Fatalf("Failed to get models: %v", err)
	}
	fmt.Printf("Got %d models in %v\n", len(models), time.Since(startTime))

	// Get models again (should be cached)
	fmt.Println("Getting models (second call, should be cached)...")
	startTime = time.Now()
	models, err = cachingProvider.GetModels(ctx)
	if err != nil {
		log.Fatalf("Failed to get models: %v", err)
	}
	fmt.Printf("Got %d models in %v\n", len(models), time.Since(startTime))

	// Example 2: Text completion (will be cached)
	request := &core.TextCompletionRequest{
		Model:       "text-davinci-003",
		Prompt:      "Hello, world!",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	fmt.Println("Generating text completion (first call)...")
	startTime = time.Now()
	response, err := cachingProvider.TextCompletion(ctx, request)
	if err != nil {
		log.Fatalf("Failed to generate text completion: %v", err)
	}
	fmt.Printf("Generated text completion in %v: %s\n", time.Since(startTime), response.Choices[0].Text)

	// Generate text completion again (should be cached)
	fmt.Println("Generating text completion (second call, should be cached)...")
	startTime = time.Now()
	response, err = cachingProvider.TextCompletion(ctx, request)
	if err != nil {
		log.Fatalf("Failed to generate text completion: %v", err)
	}
	fmt.Printf("Generated text completion in %v: %s\n", time.Since(startTime), response.Choices[0].Text)

	// Example 3: Chat completion (will be cached)
	chatRequest := &core.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []core.Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	fmt.Println("Generating chat completion (first call)...")
	startTime = time.Now()
	chatResponse, err := cachingProvider.ChatCompletion(ctx, chatRequest)
	if err != nil {
		log.Fatalf("Failed to generate chat completion: %v", err)
	}
	fmt.Printf("Generated chat completion in %v: %s\n", time.Since(startTime), chatResponse.Choices[0].Message.Content)

	// Generate chat completion again (should be cached)
	fmt.Println("Generating chat completion (second call, should be cached)...")
	startTime = time.Now()
	chatResponse, err = cachingProvider.ChatCompletion(ctx, chatRequest)
	if err != nil {
		log.Fatalf("Failed to generate chat completion: %v", err)
	}
	fmt.Printf("Generated chat completion in %v: %s\n", time.Since(startTime), chatResponse.Choices[0].Message.Content)

	// Example 4: Embedding (will be cached)
	embeddingRequest := &core.EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: "Hello, world!",
	}

	fmt.Println("Creating embedding (first call)...")
	startTime = time.Now()
	embeddingResponse, err := cachingProvider.CreateEmbedding(ctx, embeddingRequest)
	if err != nil {
		log.Fatalf("Failed to create embedding: %v", err)
	}
	fmt.Printf("Created embedding in %v with %d dimensions\n", time.Since(startTime), len(embeddingResponse.Data[0].Embedding))

	// Create embedding again (should be cached)
	fmt.Println("Creating embedding (second call, should be cached)...")
	startTime = time.Now()
	embeddingResponse, err = cachingProvider.CreateEmbedding(ctx, embeddingRequest)
	if err != nil {
		log.Fatalf("Failed to create embedding: %v", err)
	}
	fmt.Printf("Created embedding in %v with %d dimensions\n", time.Since(startTime), len(embeddingResponse.Data[0].Embedding))

	// Print cache metrics
	metrics := providerCache.GetMetrics()
	fmt.Printf("Cache metrics: %d hits, %d misses, %d evictions\n", metrics.Hits, metrics.Misses, metrics.Evictions)
	fmt.Printf("Cache size: %d entries\n", providerCache.Size())

	// Disable caching
	fmt.Println("Disabling caching...")
	providerCache.Disable()

	// Generate text completion again (should not be cached)
	fmt.Println("Generating text completion (with caching disabled)...")
	startTime = time.Now()
	response, err = cachingProvider.TextCompletion(ctx, request)
	if err != nil {
		log.Fatalf("Failed to generate text completion: %v", err)
	}
	fmt.Printf("Generated text completion in %v: %s\n", time.Since(startTime), response.Choices[0].Text)

	// Enable caching
	fmt.Println("Enabling caching...")
	providerCache.Enable()

	// Generate text completion again (should not be cached because it was generated with caching disabled)
	fmt.Println("Generating text completion (with caching enabled again)...")
	startTime = time.Now()
	response, err = cachingProvider.TextCompletion(ctx, request)
	if err != nil {
		log.Fatalf("Failed to generate text completion: %v", err)
	}
	fmt.Printf("Generated text completion in %v: %s\n", time.Since(startTime), response.Choices[0].Text)

	// Generate text completion again (should be cached now)
	fmt.Println("Generating text completion (should be cached now)...")
	startTime = time.Now()
	response, err = cachingProvider.TextCompletion(ctx, request)
	if err != nil {
		log.Fatalf("Failed to generate text completion: %v", err)
	}
	fmt.Printf("Generated text completion in %v: %s\n", time.Since(startTime), response.Choices[0].Text)

	// Clear cache
	fmt.Println("Clearing cache...")
	providerCache.Clear()
	fmt.Printf("Cache size after clearing: %d entries\n", providerCache.Size())

	// Close provider
	err = cachingProvider.Close()
	if err != nil {
		log.Fatalf("Failed to close provider: %v", err)
	}
}
