// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// MockLLMProviderImpl is a mock implementation of the LLM provider for testing OWASP vulnerabilities
type MockLLMProviderImpl struct {
	// Base provider implementation
	*core.BaseProvider
	// Configuration
	config *core.ProviderConfig
	// Vulnerable responses for specific test cases
	vulnerableResponses map[string]string
	// Default response for test cases without specific vulnerable responses
	defaultResponse string
	// Response delay to simulate latency
	responseDelay time.Duration
	// Error rate for responses (0.0 to 1.0)
	errorRate float64
	// Mutex for thread safety
	mu sync.RWMutex
}

// NewMockLLMProvider creates a new mock LLM provider for testing
func NewMockLLMProvider(config *core.ProviderConfig) *MockLLMProviderImpl {
	if config == nil {
		config = &core.ProviderConfig{
			Type:         core.CustomProvider,
			DefaultModel: "mock-llm-model",
		}
	}

	baseProvider := core.NewBaseProvider(core.CustomProvider, config)
	
	// Add default models
	models := []core.ModelInfo{
		{
			ID:           "mock-llm-model",
			Provider:     core.CustomProvider,
			Type:         core.ChatModel,
			Capabilities: []core.ModelCapability{core.ChatCompletionCapability},
			Description:  "Mock LLM model for OWASP vulnerability testing",
		},
		{
			ID:           "mock-llm-model-vulnerable",
			Provider:     core.CustomProvider,
			Type:         core.ChatModel,
			Capabilities: []core.ModelCapability{core.ChatCompletionCapability},
			Description:  "Vulnerable mock LLM model for OWASP vulnerability testing",
		},
	}
	baseProvider.SetModels(models)

	return &MockLLMProviderImpl{
		BaseProvider:       baseProvider,
		config:             config,
		vulnerableResponses: make(map[string]string),
		defaultResponse:    "This is a default response from the mock LLM provider.",
		responseDelay:      0,
		errorRate:          0.0,
	}
}

// SetVulnerableResponses sets the vulnerable responses for specific test cases
func (p *MockLLMProviderImpl) SetVulnerableResponses(vulnerableResponses map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.vulnerableResponses = vulnerableResponses
}

// GetVulnerableResponse gets a vulnerable response for a specific test case
func (p *MockLLMProviderImpl) GetVulnerableResponse(testCaseID string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	response, ok := p.vulnerableResponses[testCaseID]
	return response, ok
}

// SetDefaultResponse sets the default response for test cases without specific vulnerable responses
func (p *MockLLMProviderImpl) SetDefaultResponse(response string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.defaultResponse = response
}

// SetResponseDelay sets a delay for responses to simulate latency
func (p *MockLLMProviderImpl) SetResponseDelay(delay time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.responseDelay = delay
}

// SetErrorRate sets the error rate for responses (0.0 to 1.0)
func (p *MockLLMProviderImpl) SetErrorRate(rate float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if rate < 0.0 {
		rate = 0.0
	} else if rate > 1.0 {
		rate = 1.0
	}
	p.errorRate = rate
}

// ResetState resets the state of the mock provider
func (p *MockLLMProviderImpl) ResetState() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.vulnerableResponses = make(map[string]string)
	p.defaultResponse = "This is a default response from the mock LLM provider."
	p.responseDelay = 0
	p.errorRate = 0.0
}

// ChatCompletion generates a chat completion
func (p *MockLLMProviderImpl) ChatCompletion(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simulate response delay
	if p.responseDelay > 0 {
		time.Sleep(p.responseDelay)
	}

	// Simulate errors based on error rate
	if p.errorRate > 0 && rand.Float64() < p.errorRate {
		return nil, errors.New("simulated error from mock LLM provider")
	}

	// Extract test case ID from request metadata if available
	var testCaseID string
	if request.Metadata != nil {
		if id, ok := request.Metadata["test_case_id"].(string); ok {
			testCaseID = id
		}
	}

	// Get response content based on test case ID or use default
	var responseContent string
	if testCaseID != "" {
		if response, ok := p.vulnerableResponses[testCaseID]; ok {
			responseContent = response
		} else {
			responseContent = p.defaultResponse
		}
	} else {
		// If no test case ID, try to extract from the last message
		if len(request.Messages) > 0 {
			_ = request.Messages[len(request.Messages)-1] // lastMessage - could be used for future test case ID extraction
			// Check if the message contains a test case ID marker
			// Format: [TEST_CASE_ID:123]
			// TODO: Implement more sophisticated extraction if needed
			responseContent = p.defaultResponse
		} else {
			responseContent = p.defaultResponse
		}
	}

	// Create response
	response := &core.ChatCompletionResponse{
		ID:      fmt.Sprintf("mock-chat-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   request.Model,
		Choices: []core.ChatCompletionChoice{
			{
				Index: 0,
				Message: core.Message{
					Role:    "assistant",
					Content: responseContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: &core.TokenUsage{
			PromptTokens:     100, // Mock values
			CompletionTokens: 50,  // Mock values
			TotalTokens:      150, // Mock values
		},
	}

	return response, nil
}

// TextCompletion generates a text completion
func (p *MockLLMProviderImpl) TextCompletion(ctx context.Context, request *core.TextCompletionRequest) (*core.TextCompletionResponse, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simulate response delay
	if p.responseDelay > 0 {
		time.Sleep(p.responseDelay)
	}

	// Simulate errors based on error rate
	if p.errorRate > 0 && rand.Float64() < p.errorRate {
		return nil, errors.New("simulated error from mock LLM provider")
	}

	// Extract test case ID from request metadata if available
	var testCaseID string
	if request.User != "" {
		// Use User field to pass test case ID
		testCaseID = request.User
	}

	// Get response content based on test case ID or use default
	var responseContent string
	if testCaseID != "" {
		if response, ok := p.vulnerableResponses[testCaseID]; ok {
			responseContent = response
		} else {
			responseContent = p.defaultResponse
		}
	} else {
		responseContent = p.defaultResponse
	}

	// Create response
	response := &core.TextCompletionResponse{
		ID:      fmt.Sprintf("mock-text-%d", time.Now().UnixNano()),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   request.Model,
		Choices: []core.TextCompletionChoice{
			{
				Text:         responseContent,
				Index:        0,
				FinishReason: "stop",
			},
		},
		Usage: &core.TokenUsage{
			PromptTokens:     100, // Mock values
			CompletionTokens: 50,  // Mock values
			TotalTokens:      150, // Mock values
		},
	}

	return response, nil
}

// StreamingChatCompletion generates a streaming chat completion
func (p *MockLLMProviderImpl) StreamingChatCompletion(ctx context.Context, request *core.ChatCompletionRequest, callback func(response *core.ChatCompletionResponse) error) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simulate response delay
	if p.responseDelay > 0 {
		time.Sleep(p.responseDelay)
	}

	// Simulate errors based on error rate
	if p.errorRate > 0 && rand.Float64() < p.errorRate {
		return errors.New("simulated error from mock LLM provider")
	}

	// Extract test case ID from request metadata if available
	var testCaseID string
	if request.Metadata != nil {
		if id, ok := request.Metadata["test_case_id"].(string); ok {
			testCaseID = id
		}
	}

	// Get response content based on test case ID or use default
	var responseContent string
	if testCaseID != "" {
		if response, ok := p.vulnerableResponses[testCaseID]; ok {
			responseContent = response
		} else {
			responseContent = p.defaultResponse
		}
	} else {
		responseContent = p.defaultResponse
	}

	// Split the response into chunks for streaming
	chunks := splitIntoChunks(responseContent, 10) // 10 characters per chunk

	// Stream each chunk
	for i, chunk := range chunks {
		response := &core.ChatCompletionResponse{
			ID:      fmt.Sprintf("mock-chat-stream-%d", time.Now().UnixNano()),
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   request.Model,
			Choices: []core.ChatCompletionChoice{
				{
					Index: 0,
					Message: core.Message{
						Role:    "assistant",
						Content: chunk,
					},
					FinishReason: "",
				},
			},
		}

		// Set finish reason for the last chunk
		if i == len(chunks)-1 {
			response.Choices[0].FinishReason = "stop"
		}

		// Call the callback with the chunk
		if err := callback(response); err != nil {
			return err
		}

		// Small delay between chunks to simulate streaming
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

// CreateEmbedding creates an embedding
func (p *MockLLMProviderImpl) CreateEmbedding(ctx context.Context, request *core.EmbeddingRequest) (*core.EmbeddingResponse, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simulate response delay
	if p.responseDelay > 0 {
		time.Sleep(p.responseDelay)
	}

	// Simulate errors based on error rate
	if p.errorRate > 0 && rand.Float64() < p.errorRate {
		return nil, errors.New("simulated error from mock LLM provider")
	}

	// Generate mock embeddings
	var inputs []string
	switch v := request.Input.(type) {
	case string:
		inputs = []string{v}
	case []string:
		inputs = v
	default:
		return nil, fmt.Errorf("unsupported input type: %T", request.Input)
	}

	embeddings := make([]core.Embedding, len(inputs))
	for i, input := range inputs {
		// Generate a deterministic but unique embedding for each input
		embedding := generateMockEmbedding(input, 1536) // 1536 is a common embedding dimension
		embeddings[i] = core.Embedding{
			Object:    "embedding",
			Embedding: embedding,
			Index:     i,
		}
	}

	response := &core.EmbeddingResponse{
		Object: "list",
		Data:   embeddings,
		Model:  request.Model,
		Usage: &core.TokenUsage{
			PromptTokens:     len(inputs) * 10, // Mock values
			CompletionTokens: 0,                // No completion tokens for embeddings
			TotalTokens:      len(inputs) * 10, // Mock values
		},
	}

	return response, nil
}

// CountTokens counts the number of tokens in a text
func (p *MockLLMProviderImpl) CountTokens(ctx context.Context, text string, modelID string) (int, error) {
	// Mock implementation - just count words as a simple approximation
	words := splitIntoChunks(text, 1)
	return len(words), nil
}

// GetAllUsageMetrics returns all usage metrics for the provider
func (p *MockLLMProviderImpl) GetAllUsageMetrics() (map[string]*core.UsageMetrics, error) {
	// Mock implementation - return empty metrics
	return map[string]*core.UsageMetrics{
		"mock-llm-model": {
			Requests:             0,
			Tokens:               0,
			Errors:               0,
			LastRequestTime:      time.Now(),
			TotalRequestDuration: 0,
			AverageResponseTime:  0,
			TokensPerMinute:      0,
			RequestsPerMinute:    0,
			ModelID:              "mock-llm-model",
		},
	}, nil
}

// GetRateLimitConfig returns the rate limit configuration
func (p *MockLLMProviderImpl) GetRateLimitConfig() *core.RateLimitConfig {
	// Mock implementation - return default rate limit config
	return &core.RateLimitConfig{
		RequestsPerMinute:    100,
		TokensPerMinute:      10000,
		MaxConcurrentRequests: 10,
		BurstSize:            20,
	}
}

// GetRetryConfig returns the retry configuration
func (p *MockLLMProviderImpl) GetRetryConfig() *core.RetryConfig {
	// Mock implementation - return default retry config
	return &core.RetryConfig{
		MaxRetries:           3,
		InitialBackoff:       time.Second,
		MaxBackoff:           time.Second * 10,
		BackoffMultiplier:    2.0,
		RetryableStatusCodes: []int{429, 500, 502, 503, 504},
	}
}

// GetUsageMetrics returns the usage metrics for a specific model
func (p *MockLLMProviderImpl) GetUsageMetrics(modelID string) (*core.UsageMetrics, error) {
	// Mock implementation - return empty metrics
	return &core.UsageMetrics{
		Requests:             0,
		Tokens:               0,
		Errors:               0,
		LastRequestTime:      time.Now(),
		TotalRequestDuration: 0,
		AverageResponseTime:  0,
		TokensPerMinute:      0,
		RequestsPerMinute:    0,
		ModelID:              modelID,
	}, nil
}

// ResetUsageMetrics resets the usage metrics
func (p *MockLLMProviderImpl) ResetUsageMetrics() error {
	// Mock implementation - nothing to reset
	return nil
}

// UpdateRateLimitConfig updates the rate limit configuration
func (p *MockLLMProviderImpl) UpdateRateLimitConfig(config *core.RateLimitConfig) error {
	// Mock implementation - nothing to update
	return nil
}

// UpdateRetryConfig updates the retry configuration
func (p *MockLLMProviderImpl) UpdateRetryConfig(config *core.RetryConfig) error {
	// Mock implementation - nothing to update
	return nil
}

// Helper function to split a string into chunks
func splitIntoChunks(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{s}
	}

	var chunks []string
	runes := []rune(s)
	
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	
	return chunks
}

// Helper function to generate a mock embedding
func generateMockEmbedding(input string, dimensions int) []float64 {
	// Use a simple hash of the input to seed the random number generator
	// This ensures the same input always generates the same embedding
	var seed int64
	for i, c := range input {
		seed += int64(c) * int64(i+1)
	}
	r := rand.New(rand.NewSource(seed))

	// Generate random embedding values
	embedding := make([]float64, dimensions)
	for i := 0; i < dimensions; i++ {
		embedding[i] = r.Float64()*2 - 1 // Values between -1 and 1
	}

	// Normalize the embedding to unit length
	var sum float64
	for _, v := range embedding {
		sum += v * v
	}
	magnitude := float64(1.0)
	if sum > 0 {
		magnitude = 1.0 / float64(sum)
	}
	for i := range embedding {
		embedding[i] *= magnitude
	}

	return embedding
}
