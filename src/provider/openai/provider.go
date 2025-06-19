// Package openai provides an implementation of the Provider interface for OpenAI.
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/provider/middleware"
)

// OpenAIProvider is an implementation of the Provider interface for OpenAI
type OpenAIProvider struct {
	*core.BaseProvider
	client             *http.Client
	rateLimiter        *middleware.RateLimiter
	retryMiddleware    *middleware.RetryMiddleware
	loggingMiddleware  *middleware.LoggingMiddleware
	circuitBreaker     *middleware.CircuitBreakerMiddleware
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *core.ProviderConfig) (core.Provider, error) {
	if config == nil {
		config = &core.ProviderConfig{
			Type:        core.OpenAIProvider,
			BaseURL:     "https://api.openai.com/v1",
			Timeout:     30 * time.Second,
		}
	}

	// Validate configuration
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI provider")
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Create base provider
	baseProvider := core.NewBaseProvider(core.OpenAIProvider, config)

	// Create rate limiter
	var rateLimiter *middleware.RateLimiter
	if config.RateLimitConfig != nil {
		rateLimiter = middleware.NewRateLimiter(
			config.RateLimitConfig.RequestsPerMinute,
			config.RateLimitConfig.TokensPerMinute,
			config.RateLimitConfig.MaxConcurrentRequests,
			config.RateLimitConfig.BurstSize,
		)
	} else {
		// Default rate limits for OpenAI
		rateLimiter = middleware.NewRateLimiter(
			60,    // 60 requests per minute (1 per second)
			100000, // 100K tokens per minute
			10,     // 10 concurrent requests
			10,     // Burst size of 10
		)
	}

	// Create retry middleware
	var retryMiddleware *middleware.RetryMiddleware
	if config.RetryConfig != nil {
		retryMiddleware = middleware.NewRetryMiddleware(config.RetryConfig)
	} else {
		retryMiddleware = middleware.NewRetryMiddleware(nil) // Use default config
	}

	// Create logging middleware
	loggingMiddleware := middleware.NewLoggingMiddleware(middleware.LogLevelInfo, true)
	loggingMiddleware.AddHandler(middleware.LogLevelInfo, middleware.ConsoleLogHandler())

	// Create circuit breaker
	circuitBreaker := middleware.NewCircuitBreaker(middleware.CircuitBreakerConfig{
		FailureThreshold:        5,
		ResetTimeout:            60 * time.Second,
		HalfOpenSuccessThreshold: 2,
	})

	provider := &OpenAIProvider{
		BaseProvider:      baseProvider,
		client:            client,
		rateLimiter:       rateLimiter,
		retryMiddleware:   retryMiddleware,
		loggingMiddleware: loggingMiddleware,
		circuitBreaker:    circuitBreaker,
	}

	// Initialize models
	go provider.updateModels(context.Background())

	return provider, nil
}

// updateModels updates the models cache
func (p *OpenAIProvider) updateModels(ctx context.Context) error {
	// Execute with resilience
	result, err := p.executeWithResilience(ctx, "GetModels", nil, func(ctx context.Context) (interface{}, error) {
		return p.getModelsFromAPI(ctx)
	})

	if err != nil {
		return err
	}

	models := result.([]core.ModelInfo)
	p.SetModels(models)

	return nil
}

// getModelsFromAPI gets models from the OpenAI API
func (p *OpenAIProvider) getModelsFromAPI(ctx context.Context) ([]core.ModelInfo, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", p.GetConfig().BaseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+p.GetConfig().APIKey)
	req.Header.Add("Content-Type", "application/json")
	if p.GetConfig().OrgID != "" {
		req.Header.Add("OpenAI-Organization", p.GetConfig().OrgID)
	}

	// Add additional headers
	for key, value := range p.GetConfig().AdditionalHeaders {
		req.Header.Add(key, value)
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse response
	var modelsResponse struct {
		Data []struct {
			ID    string `json:"id"`
			Owner string `json:"owner"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to ModelInfo
	models := make([]core.ModelInfo, 0, len(modelsResponse.Data))
	for _, model := range modelsResponse.Data {
		// Skip non-OpenAI models
		if !strings.Contains(model.Owner, "openai") {
			continue
		}

		// Determine model type and capabilities
		modelType := core.TextCompletionModel
		capabilities := []core.ModelCapability{core.TextCompletionCapability}

		// Check for chat models
		if strings.HasPrefix(model.ID, "gpt-3.5") || strings.HasPrefix(model.ID, "gpt-4") {
			modelType = core.ChatModel
			capabilities = []core.ModelCapability{
				core.ChatCompletionCapability,
				core.StreamingCapability,
				core.FunctionCallingCapability,
				core.ToolUseCapability,
				core.JSONModeCapability,
			}
		}

		// Check for embedding models
		if strings.Contains(model.ID, "embedding") {
			modelType = core.EmbeddingModel
			capabilities = []core.ModelCapability{core.EmbeddingCapability}
		}

		// Add model info
		models = append(models, core.ModelInfo{
			ID:          model.ID,
			Provider:    core.OpenAIProvider,
			Type:        modelType,
			Capabilities: capabilities,
			// Other fields would be populated with more detailed information
		})
	}

	return models, nil
}

// GetModels returns a list of available models
func (p *OpenAIProvider) GetModels(ctx context.Context) ([]core.ModelInfo, error) {
	// Use the base implementation
	return p.BaseProvider.GetModels(ctx)
}

// TextCompletion generates a text completion
func (p *OpenAIProvider) TextCompletion(ctx context.Context, request *core.TextCompletionRequest) (*core.TextCompletionResponse, error) {
	// Execute with resilience
	result, err := p.executeWithResilience(ctx, "TextCompletion", request, func(ctx context.Context) (interface{}, error) {
		return p.textCompletionFromAPI(ctx, request)
	})

	if err != nil {
		return nil, err
	}

	return result.(*core.TextCompletionResponse), nil
}

// textCompletionFromAPI gets text completion from the OpenAI API
func (p *OpenAIProvider) textCompletionFromAPI(ctx context.Context, request *core.TextCompletionRequest) (*core.TextCompletionResponse, error) {
	// Set default model if not specified
	if request.Model == "" {
		if p.GetConfig().DefaultModel != "" {
			request.Model = p.GetConfig().DefaultModel
		} else {
			request.Model = "text-davinci-003" // Default model
		}
	}

	// Create request body
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+p.GetConfig().APIKey)
	req.Header.Add("Content-Type", "application/json")
	if p.GetConfig().OrgID != "" {
		req.Header.Add("OpenAI-Organization", p.GetConfig().OrgID)
	}

	// Add additional headers
	for key, value := range p.GetConfig().AdditionalHeaders {
		req.Header.Add(key, value)
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse response
	var response core.TextCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// ChatCompletion generates a chat completion
func (p *OpenAIProvider) ChatCompletion(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	// Execute with resilience
	result, err := p.executeWithResilience(ctx, "ChatCompletion", request, func(ctx context.Context) (interface{}, error) {
		return p.chatCompletionFromAPI(ctx, request)
	})

	if err != nil {
		return nil, err
	}

	return result.(*core.ChatCompletionResponse), nil
}

// chatCompletionFromAPI gets chat completion from the OpenAI API
func (p *OpenAIProvider) chatCompletionFromAPI(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	// Set default model if not specified
	if request.Model == "" {
		if p.GetConfig().DefaultModel != "" {
			request.Model = p.GetConfig().DefaultModel
		} else {
			request.Model = "gpt-3.5-turbo" // Default model
		}
	}

	// Create request body
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+p.GetConfig().APIKey)
	req.Header.Add("Content-Type", "application/json")
	if p.GetConfig().OrgID != "" {
		req.Header.Add("OpenAI-Organization", p.GetConfig().OrgID)
	}

	// Add additional headers
	for key, value := range p.GetConfig().AdditionalHeaders {
		req.Header.Add(key, value)
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse response
	var response core.ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// StreamingChatCompletion generates a streaming chat completion
func (p *OpenAIProvider) StreamingChatCompletion(ctx context.Context, request *core.ChatCompletionRequest, callback func(response *core.ChatCompletionResponse) error) error {
	// Set streaming flag
	request.Stream = true

	// Execute with resilience
	_, err := p.executeWithResilience(ctx, "StreamingChatCompletion", request, func(ctx context.Context) (interface{}, error) {
		return nil, p.streamingChatCompletionFromAPI(ctx, request, callback)
	})

	return err
}

// streamingChatCompletionFromAPI gets streaming chat completion from the OpenAI API
func (p *OpenAIProvider) streamingChatCompletionFromAPI(ctx context.Context, request *core.ChatCompletionRequest, callback func(response *core.ChatCompletionResponse) error) error {
	// Set default model if not specified
	if request.Model == "" {
		if p.GetConfig().DefaultModel != "" {
			request.Model = p.GetConfig().DefaultModel
		} else {
			request.Model = "gpt-3.5-turbo" // Default model
		}
	}

	// Create request body
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+p.GetConfig().APIKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "text/event-stream")
	if p.GetConfig().OrgID != "" {
		req.Header.Add("OpenAI-Organization", p.GetConfig().OrgID)
	}

	// Add additional headers
	for key, value := range p.GetConfig().AdditionalHeaders {
		req.Header.Add(key, value)
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check for error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return p.handleErrorResponse(resp.StatusCode, body)
	}

	// Read response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Check for data prefix
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract data
		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			break
		}

		// Parse response
		var response core.ChatCompletionResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		// Call callback
		if err := callback(&response); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	return nil
}

// CreateEmbedding creates an embedding
func (p *OpenAIProvider) CreateEmbedding(ctx context.Context, request *core.EmbeddingRequest) (*core.EmbeddingResponse, error) {
	// Execute with resilience
	result, err := p.executeWithResilience(ctx, "CreateEmbedding", request, func(ctx context.Context) (interface{}, error) {
		return p.createEmbeddingFromAPI(ctx, request)
	})

	if err != nil {
		return nil, err
	}

	return result.(*core.EmbeddingResponse), nil
}

// createEmbeddingFromAPI creates an embedding using the OpenAI API
func (p *OpenAIProvider) createEmbeddingFromAPI(ctx context.Context, request *core.EmbeddingRequest) (*core.EmbeddingResponse, error) {
	// Set default model if not specified
	if request.Model == "" {
		if p.GetConfig().DefaultModel != "" {
			request.Model = p.GetConfig().DefaultModel
		} else {
			request.Model = "text-embedding-ada-002" // Default model
		}
	}

	// Create request body
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+p.GetConfig().APIKey)
	req.Header.Add("Content-Type", "application/json")
	if p.GetConfig().OrgID != "" {
		req.Header.Add("OpenAI-Organization", p.GetConfig().OrgID)
	}

	// Add additional headers
	for key, value := range p.GetConfig().AdditionalHeaders {
		req.Header.Add(key, value)
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse response
	var response core.EmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CountTokens counts the number of tokens in a text
func (p *OpenAIProvider) CountTokens(ctx context.Context, text string, modelID string) (int, error) {
	// This is a simplified implementation
	// In a real implementation, this would use a tokenizer
	// For now, we'll just estimate based on words
	words := strings.Fields(text)
	return len(words) * 4 / 3, nil // Rough estimate: 4 tokens per 3 words
}

// Close closes the provider and releases any resources
func (p *OpenAIProvider) Close() error {
	// Nothing to close for this provider
	return nil
}

// handleErrorResponse handles an error response from the OpenAI API
func (p *OpenAIProvider) handleErrorResponse(statusCode int, body []byte) error {
	// Parse error response
	var errorResponse struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Param   string `json:"param"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return &core.ProviderError{
			StatusCode:  statusCode,
			Message:     fmt.Sprintf("Unknown error (status code: %d)", statusCode),
			RawResponse: string(body),
		}
	}

	return &core.ProviderError{
		StatusCode:  statusCode,
		Type:        errorResponse.Error.Type,
		Message:     errorResponse.Error.Message,
		Param:       errorResponse.Error.Param,
		Code:        errorResponse.Error.Code,
		RawResponse: string(body),
	}
}

// executeWithResilience executes a function with resilience
func (p *OpenAIProvider) executeWithResilience(ctx context.Context, operation string, request interface{}, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// Log request
	requestID := p.loggingMiddleware.LogRequest(ctx, p.GetType(), operation, request, nil)

	// Record start time
	startTime := time.Now()

	// Create a function that executes with rate limiting
	rateLimitedFn := func(ctx context.Context) (interface{}, error) {
		// Wait for rate limiting
		if err := p.rateLimiter.Wait(ctx); err != nil {
			return nil, err
		}
		defer p.rateLimiter.Release()

		// Execute the function
		return fn(ctx)
	}

	// Create a function that executes with circuit breaker
	circuitBreakerFn := func(ctx context.Context) (interface{}, error) {
		// Execute with circuit breaker
		return p.circuitBreaker.Execute(ctx, func(ctx context.Context) (interface{}, error) {
			return rateLimitedFn(ctx)
		})
	}

	// Execute with retry
	result, err := p.retryMiddleware.Execute(ctx, circuitBreakerFn)

	// Calculate duration
	duration := time.Since(startTime)

	// Log response
	p.loggingMiddleware.LogResponse(ctx, p.GetType(), operation, requestID, request, result, err, duration, nil)

	return result, err
}
