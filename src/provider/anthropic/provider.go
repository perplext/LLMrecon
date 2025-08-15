// Package anthropic provides an implementation of the Provider interface for Anthropic.
package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/provider/middleware"
)

// AnthropicProvider is an implementation of the Provider interface for Anthropic
type AnthropicProvider struct {
	*core.BaseProvider
	client             *http.Client
	connectionPool     *core.ProviderConnectionPool
	rateLimiter        *middleware.RateLimiter
	retryMiddleware    *middleware.RetryMiddleware
	loggingMiddleware  *middleware.LoggingMiddleware
	circuitBreaker     *middleware.CircuitBreakerMiddleware
	requestQueue       *middleware.RequestQueueMiddleware
	usageTracker       *middleware.UsageTracker
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config *core.ProviderConfig) (core.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("API key is required")
	}

	// Validate configuration
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Set default base URL if not specified
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}

	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Create connection pool configuration
	poolConfig := core.DefaultConnectionPoolConfig()
	poolConfig.ProviderType = core.AnthropicProvider
	poolConfig.BaseURL = config.BaseURL
	if config.Timeout > 0 {
		poolConfig.ResponseHeaderTimeout = config.Timeout
	}
	
	// Create connection pool manager
	logger := core.NewDefaultLogger()
	poolManager := core.NewConnectionPoolManager(poolConfig, logger)
	connectionPool, err := poolManager.CreatePool(core.AnthropicProvider, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	
	// Get HTTP client from connection pool
	client := connectionPool.GetClient()

	// Create base provider
	baseProvider := core.NewBaseProvider(core.AnthropicProvider, config)

	// Create middleware with appropriate parameters
	var rateLimiter *middleware.RateLimiter
	if config.RateLimitConfig != nil {
		rateLimiter = middleware.NewRateLimiter(
			config.RateLimitConfig.RequestsPerMinute,
			config.RateLimitConfig.TokensPerMinute,
			config.RateLimitConfig.MaxConcurrentRequests,
			config.RateLimitConfig.BurstSize,
		)
	} else {
		// Default values
		rateLimiter = middleware.NewRateLimiter(60, 100000, 10, 20)
	}

	var retryMiddleware *middleware.RetryMiddleware
	if config.RetryConfig != nil {
		retryMiddleware = middleware.NewRetryMiddleware(config.RetryConfig)
	} else {
		retryMiddleware = middleware.NewRetryMiddleware(nil)
	}

	// Using LogLevelInfo and no PII redaction as defaults
	loggingMiddleware := middleware.NewLoggingMiddleware(middleware.LogLevelInfo, false)

	// Create circuit breaker middleware with appropriate configuration
	circuitBreakerConfig := middleware.CircuitBreakerConfig{
		FailureThreshold:         5,  // 5 consecutive failures will open the circuit
		ResetTimeout:             30 * time.Second, // Wait 30 seconds before trying again
		HalfOpenSuccessThreshold: 2,  // 2 consecutive successes will close the circuit
	}
	circuitBreaker := middleware.NewCircuitBreakerMiddleware(circuitBreakerConfig)
	
	// Create request queue middleware with appropriate configuration
	requestQueueConfig := middleware.RequestQueueConfig{
		MaxQueueSize:   100, // Maximum number of requests that can be queued
		MaxWaitTime:    60 * time.Second, // Maximum time a request can wait in the queue
		PriorityLevels: 3,  // Number of priority levels (0 = highest, 2 = lowest)
	}
	requestQueue := middleware.NewRequestQueueMiddleware(requestQueueConfig, 5) // 5 worker goroutines
	// Start the request queue workers
	requestQueue.Start()
	
	// Create usage tracker with daily reset interval
	usageTracker := middleware.NewUsageTracker(24 * time.Hour)

	return &AnthropicProvider{
		BaseProvider:      baseProvider,
		client:            client,
		connectionPool:    connectionPool,
		rateLimiter:       rateLimiter,
		retryMiddleware:   retryMiddleware,
		loggingMiddleware: loggingMiddleware,
		circuitBreaker:    circuitBreaker,
		requestQueue:      requestQueue,
		usageTracker:      usageTracker,
	}, nil

// GetModels returns a list of available models
func (p *AnthropicProvider) GetModels(ctx context.Context) ([]core.ModelInfo, error) {
	result, err := p.executeWithResilience(ctx, "GetModels", nil, func(ctx context.Context) (interface{}, error) {
			// Create request
		req, err := http.NewRequestWithContext(ctx, "GET", p.GetConfig().BaseURL+"/v1/models", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", "Bearer "+p.GetConfig().APIKey)
		req.Header.Set("Anthropic-Version", "2023-06-01")
		req.Header.Set("X-Api-Client", "anthropic-LLMrecon/1.0.0")

		// Add additional headers
		for key, value := range p.GetConfig().AdditionalHeaders {
			req.Header.Set(key, value)
		}

		// Execute request
		resp, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

		// Check for error
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, p.handleErrorResponse(resp.StatusCode, body)
		}

		// Parse response
		var response struct {
			Models []struct {
				Name        string    `json:"name"`
				Description string    `json:"description"`
				MaxTokens   int       `json:"max_tokens"`
				Created     time.Time `json:"created"`
			} `json:"models"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Convert to ModelInfo
		models := make([]core.ModelInfo, 0, len(response.Models))
		for _, model := range response.Models {
			modelType := core.ChatModel
			if strings.Contains(model.Name, "embedding") {
				modelType = core.EmbeddingModel
			}

			capabilities := []core.ModelCapability{core.ChatCompletionCapability, core.StreamingCapability}
			if strings.Contains(model.Name, "claude-3") {
				capabilities = append(capabilities, core.FunctionCallingCapability, core.ToolUseCapability, core.JSONModeCapability)
			}

			models = append(models, core.ModelInfo{
				ID:             model.Name,
				Provider:       core.AnthropicProvider,
				Type:           modelType,
				Capabilities:   capabilities,
				MaxTokens:      model.MaxTokens,
				TrainingCutoff: model.Created,
				Description:    model.Description,
			})
		}

		return models, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]core.ModelInfo), nil

// GetModelInfo returns information about a specific model
func (p *AnthropicProvider) GetModelInfo(ctx context.Context, modelID string) (*core.ModelInfo, error) {
	models, err := p.GetModels(ctx)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.ID == modelID {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model %s not found", modelID)

// TextCompletion generates a text completion
func (p *AnthropicProvider) TextCompletion(ctx context.Context, request *core.TextCompletionRequest) (*core.TextCompletionResponse, error) {
	result, err := p.executeWithResilience(ctx, "TextCompletion", request, func(ctx context.Context) (interface{}, error) {
		// Validate request
		if request.Prompt == "" {
			return nil, fmt.Errorf("prompt is required")
		}

		// Set default model if not specified
		if request.Model == "" {
			request.Model = p.GetConfig().DefaultModel
			if request.Model == "" {
				request.Model = "claude-3-opus-20240229"
			}
		}

		// Create Anthropic request
		anthropicRequest := map[string]interface{}{
			"model":       request.Model,
			"messages":    []map[string]interface{}{{"role": "user", "content": request.Prompt}},
			"temperature": request.Temperature,
			"max_tokens":  request.MaxTokens,
			"top_p":       request.TopP,
		}

		// Add stop sequences if specified
		if len(request.Stop) > 0 {
			anthropicRequest["stop_sequences"] = request.Stop
		}
		// Add additional parameters
		for key, value := range p.GetConfig().AdditionalParams {
			anthropicRequest[key] = value
		}

		// Create request body
		requestBody, err := json.Marshal(anthropicRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/v1/messages", bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", "Bearer "+p.GetConfig().APIKey)
		req.Header.Set("Anthropic-Version", "2023-06-01")
		req.Header.Set("X-Api-Client", "anthropic-LLMrecon/1.0.0")

		// Add additional headers
		for key, value := range p.GetConfig().AdditionalHeaders {
			req.Header.Set(key, value)
		}

		// Execute request
		resp, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

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
		var response struct {
			ID        string `json:"id"`
			Type      string `json:"type"`
			Model     string `json:"model"`
			StopReason string `json:"stop_reason"`
			Content   []struct {
				Type string `json:"type"`
				Text string `json:"text"`
				Role string `json:"role"`
			} `json:"content"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Extract content text
		content := getContentText(response.Content)

		// Create text completion response
		// Estimate token counts using a simple approach
		promptTokens := len(strings.Fields(request.Prompt)) * 4 / 3
		completionTokens := len(strings.Fields(content)) * 4 / 3
		totalTokens := promptTokens + completionTokens

		return &core.TextCompletionResponse{
			ID:      response.ID,
			Object:  "text_completion",
			Created: time.Now().Unix(),
			Model:   response.Model,
			Choices: []core.TextCompletionChoice{
				{
					Text:         content,
					Index:        0,
					FinishReason: convertStopReasonToFinishReason(response.StopReason),
				},
			},
			Usage: &core.TokenUsage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      totalTokens,
			},
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*core.TextCompletionResponse), nil

// ChatCompletion generates a chat completion
func (p *AnthropicProvider) ChatCompletion(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	result, _ := p.executeWithResilience(ctx, "ChatCompletion", request, func(ctx context.Context) (interface{}, error) {
		// Validate request
		if len(request.Messages) == 0 {
			return nil, fmt.Errorf("messages are required")
		}

		// Set default model if not specified
		if request.Model == "" {
			request.Model = p.GetConfig().DefaultModel
			if request.Model == "" {
				request.Model = "claude-3-opus-20240229"
			}
		}

		// Create Anthropic request
		anthropicRequest := map[string]interface{}{
			"model":       request.Model,
			"messages":    convertMessagesToAnthropicFormat(request.Messages),
			"temperature": request.Temperature,
			"max_tokens":  request.MaxTokens,
			"top_p":       request.TopP,
		}

		// Add stop sequences if specified
		if len(request.Stop) > 0 {
			anthropicRequest["stop_sequences"] = request.Stop
		}

		// Add tools if specified
		if len(request.Tools) > 0 {
			anthropicRequest["tools"] = convertToolsToAnthropicFormat(request.Tools)
		}

		// Extract system message if present in the messages
		for _, msg := range request.Messages {
			if msg.Role == "system" {
				anthropicRequest["system"] = msg.Content
				break
			}
		}

		// Add additional parameters
		for key, value := range p.GetConfig().AdditionalParams {
			anthropicRequest[key] = value
		}
		// Create request body
		requestBody, err := json.Marshal(anthropicRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/v1/messages", bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", "Bearer "+p.GetConfig().APIKey)
		req.Header.Set("Anthropic-Version", "2023-06-01")
		req.Header.Set("X-Api-Client", "anthropic-LLMrecon/1.0.0")

		// Add additional headers
		for key, value := range p.GetConfig().AdditionalHeaders {
			req.Header.Set(key, value)
		}

		// Execute request
		resp, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

		// Check for error
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, p.handleErrorResponse(resp.StatusCode, body)
		}

		// Parse response
		var response struct {
			ID        string `json:"id"`
			Type      string `json:"type"`
			Model     string `json:"model"`
			StopReason string `json:"stop_reason"`
			Content   []struct {
				Type string `json:"type"`
				Text string `json:"text"`
				Role string `json:"role"`
			} `json:"content"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Create chat completion response
		return &core.ChatCompletionResponse{
			ID:      response.ID,
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   response.Model,
			Choices: []core.ChatCompletionChoice{
				{
					Index: 0,
					Message: core.Message{
						Role:    "assistant",
						Content: getContentText(response.Content),
					},
					FinishReason: convertStopReasonToFinishReason(response.StopReason),
				},
			},
			Usage: &core.TokenUsage{
				PromptTokens:     response.Usage.InputTokens,
				CompletionTokens: response.Usage.OutputTokens,
				TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
			},
		}, nil
	})

	return result.(*core.ChatCompletionResponse), nil

// StreamingChatCompletion generates a streaming chat completion
func (p *AnthropicProvider) StreamingChatCompletion(ctx context.Context, request *core.ChatCompletionRequest, callback func(response *core.ChatCompletionResponse) error) error {
	// Validate request
	if len(request.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}

	// Set default model if not specified
	if request.Model == "" {
		request.Model = p.GetConfig().DefaultModel
		if request.Model == "" {
			request.Model = "claude-3-opus-20240229"
		}
	}

	// Create Anthropic request
	anthropicRequest := map[string]interface{}{
		"model":       request.Model,
		"messages":    convertMessagesToAnthropicFormat(request.Messages),
		"temperature": request.Temperature,
		"max_tokens":  request.MaxTokens,
		"top_p":       request.TopP,
		"stream":      true,
	}
	
	// Stream parameter is added directly to the URL in the request creation

	// Add stop sequences if specified
	if len(request.Stop) > 0 {
		anthropicRequest["stop_sequences"] = request.Stop
	}

	// Extract system message if present in the messages
	for _, msg := range request.Messages {
		if msg.Role == "system" {
			anthropicRequest["system"] = msg.Content
			break
		}
	}

	// Add additional parameters
	for key, value := range p.GetConfig().AdditionalParams {
		anthropicRequest[key] = value
	}

	// Create request body
	requestBody, err := json.Marshal(anthropicRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with stream parameter in URL for test compatibility
	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/v1/messages?stream=true", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", "Bearer "+p.GetConfig().APIKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")
	req.Header.Set("X-Api-Client", "anthropic-LLMrecon/1.0.0")
	req.Header.Set("Accept", "text/event-stream")

	// Add additional headers
	for key, value := range p.GetConfig().AdditionalHeaders {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Check for error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return p.handleErrorResponse(resp.StatusCode, body)
	}

	// Read response line by line
	scanner := bufio.NewScanner(resp.Body)
	var responseID string
	var model string
	var stopReason string
	var contentBuffer strings.Builder
	var inputTokens, outputTokens int

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

		// Parse event
		var event struct {
			Type         string `json:"type"`
			Index        int    `json:"index,omitempty"`
			Message      struct {
				ID         string `json:"id"`
				Role       string `json:"role"`
				Content    []interface{} `json:"content"`
				Model      string `json:"model"`
				StopReason interface{} `json:"stop_reason"`
				Usage      struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens,omitempty"`
				} `json:"usage"`
			} `json:"message,omitempty"`
			ContentBlock struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content_block,omitempty"`
			Delta        struct {
				Type       string `json:"type"`
				Text       string `json:"text,omitempty"`
				StopReason string `json:"stop_reason,omitempty"`
				Usage      struct {
					OutputTokens int `json:"output_tokens,omitempty"`
				} `json:"usage,omitempty"`
			} `json:"delta,omitempty"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("failed to parse event: %w", err)
		}

		// Process event based on type
		switch event.Type {
		case "message_start":
			responseID = event.Message.ID
			model = event.Message.Model
			inputTokens = event.Message.Usage.InputTokens
		case "content_block_start":
			// Nothing to do
		case "content_block_delta":
			if event.Delta.Type == "text_delta" {
				contentBuffer.WriteString(event.Delta.Text)
			}
		case "content_block_stop":
			// Nothing to do
		case "message_delta":
			// Update stop reason and usage if available
			if event.Delta.StopReason != "" {
				stopReason = event.Delta.StopReason
			}
			if event.Delta.Usage.OutputTokens > 0 {
				outputTokens = event.Delta.Usage.OutputTokens
			}
		case "message_stop":
			// Create final response
			response := &core.ChatCompletionResponse{
				ID:      responseID,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
				Choices: []core.ChatCompletionChoice{
					{
						Index: 0,
						Message: core.Message{
							Role:    "assistant",
							Content: contentBuffer.String(),
						},
						FinishReason: convertStopReasonToFinishReason(stopReason),
					},
				},
				Usage: &core.TokenUsage{
					PromptTokens:     inputTokens,
					CompletionTokens: outputTokens,
					TotalTokens:      inputTokens + outputTokens,
				},
			}
			// Call callback with final response
			if err := callback(response); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}

		// Call callback with chunk if there's content
		if event.Type == "content_block_delta" && event.Delta.Type == "text" && event.Delta.Text != "" {
			response := &core.ChatCompletionResponse{
				ID:      responseID,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
				Choices: []core.ChatCompletionChoice{
					{
						Index: 0,
						Message: core.Message{
							Role:    "assistant",
							Content: event.Delta.Text,
						},
						FinishReason: "",
					},
				},
			}
			if err := callback(response); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	return nil

// CreateEmbedding creates an embedding
func (p *AnthropicProvider) CreateEmbedding(ctx context.Context, request *core.EmbeddingRequest) (*core.EmbeddingResponse, error) {
	return nil, fmt.Errorf("embeddings are not supported by Anthropic provider")

// CountTokens counts the number of tokens in a text
func (p *AnthropicProvider) CountTokens(ctx context.Context, text string, modelID string) (int, error) {
	// This is a simplified implementation
	// In a real implementation, this would use a tokenizer
	// For now, we'll just estimate based on words
	words := strings.Fields(text)
	return len(words) * 4 / 3, nil // Rough estimate: 4 tokens per 3 words

// Close closes the provider and releases any resources
func (p *AnthropicProvider) Close() error {
	// Stop the request queue workers
	p.requestQueue.Stop()
	
	// Stop the connection pool
	if p.connectionPool != nil {
		p.connectionPool.Stop()
	}
	
	// Close the HTTP client if it implements io.Closer
	if closer, ok := interface{}(p.client).(io.Closer); ok {
		closer.Close()
	}
	
	return nil

// GetUsageMetrics returns the usage metrics for a specific model
func (p *AnthropicProvider) GetUsageMetrics(modelID string) (*core.UsageMetrics, error) {
	if modelID == "" {
		return nil, fmt.Errorf("model ID is required")
	}

	// Get metrics for the specific model
	middlewareMetrics := p.usageTracker.GetMetrics(modelID)
	if middlewareMetrics == nil {
		return nil, fmt.Errorf("no metrics found for model %s", modelID)
	}

	// Calculate average response time
	var avgResponseTime time.Duration
	if middlewareMetrics.Requests > 0 {
		avgResponseTime = time.Duration(middlewareMetrics.TotalRequestDuration.Nanoseconds() / middlewareMetrics.Requests)
	}

	// Calculate tokens and requests per minute based on the last hour
	var tokensPerMinute, requestsPerMinute float64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	if middlewareMetrics.LastRequestTime.After(oneHourAgo) {
		elapsedMinutes := time.Since(oneHourAgo).Minutes()
		if elapsedMinutes > 0 {
			tokensPerMinute = float64(middlewareMetrics.Tokens) / elapsedMinutes
			requestsPerMinute = float64(middlewareMetrics.Requests) / elapsedMinutes
		}
	}

	// Convert middleware.UsageMetrics to core.UsageMetrics
	coreMetrics := &core.UsageMetrics{
		ModelID:              modelID,
		Requests:             middlewareMetrics.Requests,
		Tokens:               middlewareMetrics.Tokens,
		Errors:               middlewareMetrics.Errors,
		LastRequestTime:      middlewareMetrics.LastRequestTime,
		TotalRequestDuration: middlewareMetrics.TotalRequestDuration,
		AverageResponseTime:  avgResponseTime,
		TokensPerMinute:      tokensPerMinute,
		RequestsPerMinute:    requestsPerMinute,
	}

	return coreMetrics, nil

// GetAllUsageMetrics returns the usage metrics for all models
func (p *AnthropicProvider) GetAllUsageMetrics() (map[string]*core.UsageMetrics, error) {
	middlewareMetrics := p.usageTracker.GetAllMetrics()
	if middlewareMetrics == nil {
		return nil, fmt.Errorf("no metrics available")
	}

	// Convert middleware.UsageMetrics to core.UsageMetrics
	coreMetrics := make(map[string]*core.UsageMetrics)
	for modelID, metrics := range middlewareMetrics {
		// Calculate average response time
		var avgResponseTime time.Duration
		if metrics.Requests > 0 {
			avgResponseTime = time.Duration(metrics.TotalRequestDuration.Nanoseconds() / metrics.Requests)
		}

		// Calculate tokens and requests per minute based on the last hour
		var tokensPerMinute, requestsPerMinute float64
		oneHourAgo := time.Now().Add(-1 * time.Hour)
		if metrics.LastRequestTime.After(oneHourAgo) {
			elapsedMinutes := time.Since(oneHourAgo).Minutes()
			if elapsedMinutes > 0 {
				tokensPerMinute = float64(metrics.Tokens) / elapsedMinutes
				requestsPerMinute = float64(metrics.Requests) / elapsedMinutes
			}
		}

		coreMetrics[modelID] = &core.UsageMetrics{
			ModelID:              modelID,
			Requests:             metrics.Requests,
			Tokens:               metrics.Tokens,
			Errors:               metrics.Errors,
			LastRequestTime:      metrics.LastRequestTime,
			TotalRequestDuration: metrics.TotalRequestDuration,
			AverageResponseTime:  avgResponseTime,
			TokensPerMinute:      tokensPerMinute,
			RequestsPerMinute:    requestsPerMinute,
		}
	}

	return coreMetrics, nil

// ResetUsageMetrics resets the usage metrics
func (p *AnthropicProvider) ResetUsageMetrics() error {
	p.usageTracker.ResetMetrics()
	return nil

// GetRateLimitConfig returns the rate limit configuration
func (p *AnthropicProvider) GetRateLimitConfig() *core.RateLimitConfig {
	requestsPerMinute, tokensPerMinute, maxConcurrentRequests, burstSize := p.rateLimiter.GetLimits()
	return &core.RateLimitConfig{
		RequestsPerMinute:     requestsPerMinute,
		TokensPerMinute:       tokensPerMinute,
		MaxConcurrentRequests: maxConcurrentRequests,
		BurstSize:             burstSize,
	}

// UpdateRateLimitConfig updates the rate limit configuration
func (p *AnthropicProvider) UpdateRateLimitConfig(config *core.RateLimitConfig) error {
	if config == nil {
		return fmt.Errorf("rate limit config cannot be nil")
	}
	
	// Update the rate limiter with the new configuration
	p.rateLimiter.UpdateLimits(
		config.RequestsPerMinute,
		config.TokensPerMinute,
		config.MaxConcurrentRequests,
		config.BurstSize,
	)
	
	return nil

// GetRetryConfig returns the retry configuration
func (p *AnthropicProvider) GetRetryConfig() *core.RetryConfig {
	return p.retryMiddleware.GetConfig()
	

// UpdateRetryConfig updates the retry configuration
func (p *AnthropicProvider) UpdateRetryConfig(config *core.RetryConfig) error {
	if config == nil {
		return fmt.Errorf("retry config cannot be nil")
	}
	
	// Update the retry middleware with the new configuration
	p.retryMiddleware.UpdateConfig(config)
	return nil

// handleErrorResponse handles an error response from the Anthropic API
func (p *AnthropicProvider) handleErrorResponse(statusCode int, body []byte) error {
	// Parse error response
	var errorResponse struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
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
		RawResponse: string(body),
	}

// executeWithResilience executes a function with resilience
func (p *AnthropicProvider) executeWithResilience(ctx context.Context, operation string, request interface{}, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
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
		return p.circuitBreaker.Execute(ctx, rateLimitedFn)
	}

	// Create a function that executes with retry
	retryFn := func(ctx context.Context) (interface{}, error) {
		// Execute with retry
		return p.retryMiddleware.Execute(ctx, circuitBreakerFn)
	}

	// Determine priority based on operation type
	priority := 1 // Default to medium priority
	if operation == "GetModels" {
		priority = 2 // Lower priority for non-critical operations
	} else if strings.Contains(operation, "Streaming") {
		priority = 0 // Higher priority for streaming operations
	}

	// Execute with request queue
	result, err := p.requestQueue.Execute(ctx, priority, retryFn)

	// Calculate duration
	duration := time.Since(startTime)

	// Log response
	p.loggingMiddleware.LogResponse(ctx, p.GetType(), operation, requestID, request, result, err, duration, nil)
	
	// Track usage metrics
	if err == nil {
		// Estimate token usage based on request and response
		var tokenCount int64
		modelID := "claude-3"
		
		// Extract model ID if available
		if req, ok := request.(map[string]interface{}); ok {
			if model, ok := req["model"].(string); ok {
				modelID = model
			}
		}
		
		// Estimate token count based on operation and result
		switch operation {
		case "ChatCompletion":
			// For chat completions, estimate based on input and output text
			if resp, ok := result.(map[string]interface{}); ok {
				if content, ok := resp["content"].([]interface{}); ok && len(content) > 0 {
					if textContent, ok := content[0].(map[string]interface{}); ok {
						if text, ok := textContent["text"].(string); ok {
							// Rough estimate: 1 token per 4 characters
							tokenCount = int64(len(text) / 4)
						}
					}
				}
			}
		case "TextCompletion":
			// For text completions, estimate based on output text
			if resp, ok := result.(map[string]interface{}); ok {
				if completion, ok := resp["completion"].(string); ok {
					// Rough estimate: 1 token per 4 characters
					tokenCount = int64(len(completion) / 4)
				}
			}
		default:
			// Default token count for other operations
			tokenCount = 10
		}
		
		// Track the usage
		p.usageTracker.TrackRequest(modelID, tokenCount, duration, nil)
	} else {
		// Track error
		p.usageTracker.TrackRequest("error", 0, duration, err)
	}

	return result, err

// Helper functions

// convertMessagesToAnthropicFormat converts messages to Anthropic API format
func convertMessagesToAnthropicFormat(messages []core.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
	for _, message := range messages {
		// Skip system messages as they're handled separately in Anthropic
		if message.Role == "system" {
			continue
		}

		// Map role
		role := message.Role
		if role == "user" {
			role = "user"
		} else if role == "assistant" {
			role = "assistant"
		} else {
			// Skip unknown roles
			continue
		}

		// Add message
		result = append(result, map[string]interface{}{
			"role":    role,
			"content": message.Content,
		})
	}
	return result

// convertToolsToAnthropicFormat converts tools to Anthropic API format
func convertToolsToAnthropicFormat(tools []core.Tool) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		if tool.Type == "function" && tool.Function != nil {
			result = append(result, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Function.Name,
					"description": tool.Function.Description,
					"parameters":  tool.Function.Parameters,
				},
			})
		}
	}
	return result

// getContentText extracts text content from Anthropic response content
func getContentText(content []struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Role  string `json:"role"`
) string {
	var result strings.Builder
	for _, block := range content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}
	return result.String()

// convertStopReasonToFinishReason converts Anthropic stop reason to standard finish reason
func convertStopReasonToFinishReason(stopReason string) string {
	switch stopReason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	default:
		return stopReason
	}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
