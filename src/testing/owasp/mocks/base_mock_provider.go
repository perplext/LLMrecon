// Package mocks provides mock implementations of LLM providers for OWASP testing
package mocks

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// BaseMockProviderImpl is a concrete implementation of the BaseMockProvider
type BaseMockProviderImpl struct {
	*BaseMockProvider
	mu                sync.RWMutex
	requestCount      int64
	usageMetrics      map[string]*core.UsageMetrics
	usageMetricsMutex sync.RWMutex
}

// NewBaseMockProviderImpl creates a new base mock provider implementation
func NewBaseMockProviderImpl(config *MockProviderConfig) *BaseMockProviderImpl {
	base := NewBaseMockProvider(config)
	
	// Set up standard models for this provider type
	models := CreateStandardMockModels(config.ProviderType)
	base.SetModels(models)
	
	return &BaseMockProviderImpl{
		BaseMockProvider: base,
		usageMetrics:     make(map[string]*core.UsageMetrics),
	}
}

// SetVulnerableResponses sets the vulnerable responses for specific test cases
func (p *BaseMockProviderImpl) SetVulnerableResponses(vulnerableResponses map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.VulnerableResponses = vulnerableResponses
}

// GetVulnerableResponse gets a vulnerable response for a specific test case
func (p *BaseMockProviderImpl) GetVulnerableResponse(testCaseID string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	response, ok := p.config.VulnerableResponses[testCaseID]
	return response, ok
}

// SetDefaultResponse sets the default response for test cases without specific vulnerable responses
func (p *BaseMockProviderImpl) SetDefaultResponse(response string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.DefaultResponse = response
}

// SetResponseDelay sets a delay for responses to simulate latency
func (p *BaseMockProviderImpl) SetResponseDelay(delay time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.ResponseDelay = delay
}

// SetErrorRate sets the error rate for responses (0.0 to 1.0)
func (p *BaseMockProviderImpl) SetErrorRate(rate float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.ErrorRate = rate
}

// ResetState resets the state of the mock provider
func (p *BaseMockProviderImpl) ResetState() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.requestCount = 0
	p.config.VulnerableResponses = make(map[string]string)
	p.config.DefaultResponse = "This is a default response from the mock LLM provider."
	p.config.ResponseDelay = 0
	p.config.ErrorRate = 0.0
	p.config.SimulateRateLimiting = false
	p.config.SimulateTimeout = false
	p.config.SimulateNetworkErrors = false
	p.config.SimulateServerErrors = false
	
	// Reset usage metrics
	p.usageMetricsMutex.Lock()
	defer p.usageMetricsMutex.Unlock()
	p.usageMetrics = make(map[string]*core.UsageMetrics)
}

// EnableVulnerability enables a specific vulnerability type
func (p *BaseMockProviderImpl) EnableVulnerability(vulnerabilityType types.VulnerabilityType, behavior *VulnerabilityBehavior) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.config.VulnerabilityBehaviors == nil {
		p.config.VulnerabilityBehaviors = make(map[types.VulnerabilityType]*VulnerabilityBehavior)
	}
	
	if behavior == nil {
		behavior = &VulnerabilityBehavior{
			Enabled:         true,
			ResponsePatterns: []string{"This is a default vulnerable response for " + string(vulnerabilityType)},
			TriggerPhrases:  []string{},
			Severity:        core.SeverityMedium,
			Metadata:        make(map[string]interface{}),
		}
	} else {
		behavior.Enabled = true
	}
	
	p.config.VulnerabilityBehaviors[vulnerabilityType] = behavior
}

// DisableVulnerability disables a specific vulnerability type
func (p *BaseMockProviderImpl) DisableVulnerability(vulnerabilityType types.VulnerabilityType) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.config.VulnerabilityBehaviors == nil {
		return
	}
	
	if behavior, ok := p.config.VulnerabilityBehaviors[vulnerabilityType]; ok {
		behavior.Enabled = false
	}
}

// IsVulnerabilityEnabled checks if a specific vulnerability type is enabled
func (p *BaseMockProviderImpl) IsVulnerabilityEnabled(vulnerabilityType types.VulnerabilityType) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if p.config.VulnerabilityBehaviors == nil {
		return false
	}
	
	if behavior, ok := p.config.VulnerabilityBehaviors[vulnerabilityType]; ok {
		return behavior.Enabled
	}
	
	return false
}

// GetVulnerabilityBehavior gets the behavior configuration for a specific vulnerability type
func (p *BaseMockProviderImpl) GetVulnerabilityBehavior(vulnerabilityType types.VulnerabilityType) *VulnerabilityBehavior {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if p.config.VulnerabilityBehaviors == nil {
		return nil
	}
	
	return p.config.VulnerabilityBehaviors[vulnerabilityType]
}

// SimulateRateLimiting enables or disables rate limiting simulation
func (p *BaseMockProviderImpl) SimulateRateLimiting(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.SimulateRateLimiting = enabled
}

// SimulateTimeout enables or disables timeout simulation
func (p *BaseMockProviderImpl) SimulateTimeout(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.SimulateTimeout = enabled
}

// SimulateNetworkErrors enables or disables network error simulation
func (p *BaseMockProviderImpl) SimulateNetworkErrors(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.SimulateNetworkErrors = enabled
}

// SimulateServerErrors enables or disables server error simulation
func (p *BaseMockProviderImpl) SimulateServerErrors(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config.SimulateServerErrors = enabled
}

// TextCompletion generates a text completion
func (p *BaseMockProviderImpl) TextCompletion(ctx context.Context, request *core.TextCompletionRequest) (*core.TextCompletionResponse, error) {
	p.mu.Lock()
	p.requestCount++
	requestCount := p.requestCount
	p.mu.Unlock()
	
	startTime := time.Now()
	
	// Simulate response delay
	if p.config.ResponseDelay > 0 {
		time.Sleep(p.config.ResponseDelay)
	}
	
	// Simulate errors based on error rate
	if p.shouldReturnError() {
		return nil, p.generateError(requestCount)
	}
	
	// Get response based on test case ID or default
	response := p.config.DefaultResponse
	if request.Metadata != nil {
		if testCaseID, ok := request.Metadata["test_case_id"].(string); ok {
			if vulnResponse, ok := p.GetVulnerableResponse(testCaseID); ok {
				response = vulnResponse
			}
		}
	}
	
	// Create token usage
	tokenUsage := p.getTokenUsage(request.Prompt, response)
	
	// Update usage metrics
	duration := time.Since(startTime)
	p.updateUsageMetrics(request.ModelID, tokenUsage.TotalTokens, duration, nil)
	
	return &core.TextCompletionResponse{
		ID:          fmt.Sprintf("mock-completion-%d", requestCount),
		Object:      "text_completion",
		Created:     time.Now().Unix(),
		Model:       request.ModelID,
		Choices: []core.TextCompletionChoice{
			{
				Text:         response,
				Index:        0,
				FinishReason: "stop",
			},
		},
		Usage:       tokenUsage,
	}, nil
}

// ChatCompletion generates a chat completion
func (p *BaseMockProviderImpl) ChatCompletion(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	p.mu.Lock()
	p.requestCount++
	requestCount := p.requestCount
	p.mu.Unlock()
	
	startTime := time.Now()
	
	// Simulate response delay
	if p.config.ResponseDelay > 0 {
		time.Sleep(p.config.ResponseDelay)
	}
	
	// Simulate errors based on error rate
	if p.shouldReturnError() {
		return nil, p.generateError(requestCount)
	}
	
	// Extract the last user message for vulnerability checking
	var lastUserMessage string
	if len(request.Messages) > 0 {
		for i := len(request.Messages) - 1; i >= 0; i-- {
			if request.Messages[i].Role == "user" {
				lastUserMessage = request.Messages[i].Content
				break
			}
		}
	}
	
	// Get response based on test case ID, vulnerability triggers, or default
	response := p.getResponseForChatRequest(request, lastUserMessage)
	
	// Create token usage
	tokenUsage := p.getTokenUsageForChat(request.Messages, response)
	
	// Update usage metrics
	duration := time.Since(startTime)
	p.updateUsageMetrics(request.ModelID, tokenUsage.TotalTokens, duration, nil)
	
	return &core.ChatCompletionResponse{
		ID:          fmt.Sprintf("mock-chat-%d", requestCount),
		Object:      "chat.completion",
		Created:     time.Now().Unix(),
		Model:       request.ModelID,
		Choices: []core.ChatCompletionChoice{
			{
				Index:        0,
				Message: core.ChatMessage{
					Role:    "assistant",
					Content: response,
				},
				FinishReason: "stop",
			},
		},
		Usage:       tokenUsage,
	}, nil
}

// StreamingChatCompletion generates a streaming chat completion
func (p *BaseMockProviderImpl) StreamingChatCompletion(ctx context.Context, request *core.ChatCompletionRequest, callback func(response *core.ChatCompletionResponse) error) error {
	p.mu.Lock()
	p.requestCount++
	requestCount := p.requestCount
	p.mu.Unlock()
	
	startTime := time.Now()
	
	// Simulate response delay
	if p.config.ResponseDelay > 0 {
		time.Sleep(p.config.ResponseDelay)
	}
	
	// Simulate errors based on error rate
	if p.shouldReturnError() {
		return p.generateError(requestCount)
	}
	
	// Extract the last user message for vulnerability checking
	var lastUserMessage string
	if len(request.Messages) > 0 {
		for i := len(request.Messages) - 1; i >= 0; i-- {
			if request.Messages[i].Role == "user" {
				lastUserMessage = request.Messages[i].Content
				break
			}
		}
	}
	
	// Get response based on test case ID, vulnerability triggers, or default
	response := p.getResponseForChatRequest(request, lastUserMessage)
	
	// Create token usage
	tokenUsage := p.getTokenUsageForChat(request.Messages, response)
	
	// Split the response into chunks for streaming
	chunks := p.splitResponseIntoChunks(response, 20) // 20 characters per chunk
	
	// Stream each chunk with a small delay
	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Create a partial response
			streamResponse := &core.ChatCompletionResponse{
				ID:          fmt.Sprintf("mock-chat-stream-%d-%d", requestCount, i),
				Object:      "chat.completion.chunk",
				Created:     time.Now().Unix(),
				Model:       request.ModelID,
				Choices: []core.ChatCompletionChoice{
					{
						Index:        0,
						Message: core.ChatMessage{
							Role:    "assistant",
							Content: chunk,
						},
						FinishReason: func() string {
							if i == len(chunks)-1 {
								return "stop"
							}
							return ""
						}(),
					},
				},
			}
			
			// Only include usage in the final chunk
			if i == len(chunks)-1 {
				streamResponse.Usage = tokenUsage
			}
			
			// Send the chunk through the callback
			if err := callback(streamResponse); err != nil {
				return err
			}
			
			// Small delay between chunks to simulate streaming
			time.Sleep(50 * time.Millisecond)
		}
	}
	
	// Update usage metrics
	duration := time.Since(startTime)
	p.updateUsageMetrics(request.ModelID, tokenUsage.TotalTokens, duration, nil)
	
	return nil
}

// CreateEmbedding creates an embedding
func (p *BaseMockProviderImpl) CreateEmbedding(ctx context.Context, request *core.EmbeddingRequest) (*core.EmbeddingResponse, error) {
	p.mu.Lock()
	p.requestCount++
	requestCount := p.requestCount
	p.mu.Unlock()
	
	startTime := time.Now()
	
	// Simulate response delay
	if p.config.ResponseDelay > 0 {
		time.Sleep(p.config.ResponseDelay)
	}
	
	// Simulate errors based on error rate
	if p.shouldReturnError() {
		return nil, p.generateError(requestCount)
	}
	
	// Generate mock embeddings
	embeddings := make([]core.Embedding, 0, len(request.Input))
	for i, text := range request.Input {
		// Generate deterministic but seemingly random embedding vector
		vector := p.generateMockEmbedding(text, 1536) // 1536 is a common embedding dimension
		
		embeddings = append(embeddings, core.Embedding{
			Index:     i,
			Object:    "embedding",
			Embedding: vector,
		})
	}
	
	// Calculate token usage
	tokenCount := 0
	for _, text := range request.Input {
		tokenCount += p.estimateTokenCount(text)
	}
	
	tokenUsage := &core.TokenUsage{
		PromptTokens:     int64(tokenCount),
		CompletionTokens: 0,
		TotalTokens:      int64(tokenCount),
	}
	
	// Update usage metrics
	duration := time.Since(startTime)
	p.updateUsageMetrics(request.ModelID, tokenUsage.TotalTokens, duration, nil)
	
	return &core.EmbeddingResponse{
		Object:    "list",
		Data:      embeddings,
		Model:     request.ModelID,
		Usage:     tokenUsage,
	}, nil
}

// CountTokens counts the number of tokens in a text
func (p *BaseMockProviderImpl) CountTokens(ctx context.Context, text string, modelID string) (int, error) {
	// Simple token count estimation
	return p.estimateTokenCount(text), nil
}

// Helper methods

// shouldReturnError determines if an error should be returned based on the error rate
func (p *BaseMockProviderImpl) shouldReturnError() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if p.config.ErrorRate <= 0 {
		return false
	}
	
	return rand.Float64() < p.config.ErrorRate
}

// generateError generates an appropriate error based on the provider configuration
func (p *BaseMockProviderImpl) generateError(requestID int64) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if p.config.SimulateRateLimiting {
		return fmt.Errorf("rate limit exceeded: please retry after 60s")
	}
	
	if p.config.SimulateTimeout {
		return context.DeadlineExceeded
	}
	
	if p.config.SimulateNetworkErrors {
		return errors.New("network error: connection reset by peer")
	}
	
	if p.config.SimulateServerErrors {
		return fmt.Errorf("server error: internal server error (500)")
	}
	
	// Default error
	return fmt.Errorf("mock provider error for request %d", requestID)
}

// getResponseForChatRequest gets the appropriate response for a chat request
func (p *BaseMockProviderImpl) getResponseForChatRequest(request *core.ChatCompletionRequest, lastUserMessage string) string {
	// Try to get response from test case ID in metadata
	if request.Metadata != nil {
		if testCaseID, ok := request.Metadata["test_case_id"].(string); ok {
			if vulnResponse, ok := p.GetVulnerableResponse(testCaseID); ok {
				return vulnResponse
			}
		}
	}
	
	// Extract test case ID from messages if not found in metadata
	testCaseID := ExtractTestCaseID(request)
	if testCaseID != "" {
		if vulnResponse, ok := p.GetVulnerableResponse(testCaseID); ok {
			return vulnResponse
		}
	}
	
	// Check for vulnerability triggers in the last user message
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if p.config.VulnerabilityBehaviors != nil && lastUserMessage != "" {
		for vulnType, behavior := range p.config.VulnerabilityBehaviors {
			if behavior.Enabled && MessageTriggerVulnerability(lastUserMessage, behavior) {
				if pattern := GetRandomResponsePattern(behavior); pattern != "" {
					return pattern
				}
			}
		}
	}
	
	// Default response
	return p.config.DefaultResponse
}

// getTokenUsage calculates token usage for a text completion
func (p *BaseMockProviderImpl) getTokenUsage(prompt, completion string) *core.TokenUsage {
	promptTokens := p.estimateTokenCount(prompt)
	completionTokens := p.estimateTokenCount(completion)
	
	return &core.TokenUsage{
		PromptTokens:     int64(promptTokens),
		CompletionTokens: int64(completionTokens),
		TotalTokens:      int64(promptTokens + completionTokens),
	}
}

// getTokenUsageForChat calculates token usage for a chat completion
func (p *BaseMockProviderImpl) getTokenUsageForChat(messages []core.ChatMessage, completion string) *core.TokenUsage {
	promptTokens := 0
	for _, msg := range messages {
		promptTokens += p.estimateTokenCount(msg.Content)
	}
	
	completionTokens := p.estimateTokenCount(completion)
	
	return &core.TokenUsage{
		PromptTokens:     int64(promptTokens),
		CompletionTokens: int64(completionTokens),
		TotalTokens:      int64(promptTokens + completionTokens),
	}
}

// estimateTokenCount estimates the token count for a text
// This is a simple implementation that assumes 4 characters per token on average
func (p *BaseMockProviderImpl) estimateTokenCount(text string) int {
	if text == "" {
		return 0
	}
	
	// Simple estimation: ~4 characters per token on average
	return (len(text) + 3) / 4
}

// splitResponseIntoChunks splits a response into chunks for streaming
func (p *BaseMockProviderImpl) splitResponseIntoChunks(response string, chunkSize int) []string {
	if response == "" {
		return []string{""}
	}
	
	var chunks []string
	for i := 0; i < len(response); i += chunkSize {
		end := i + chunkSize
		if end > len(response) {
			end = len(response)
		}
		chunks = append(chunks, response[i:end])
	}
	
	return chunks
}

// generateMockEmbedding generates a mock embedding vector
func (p *BaseMockProviderImpl) generateMockEmbedding(text string, dimension int) []float32 {
	// Use the text as a seed for deterministic but seemingly random embeddings
	seed := int64(0)
	for i, c := range text {
		seed += int64(c) * int64(i+1)
	}
	
	// Create a new random source with the seed
	source := rand.NewSource(seed)
	rng := rand.New(source)
	
	// Generate the embedding vector
	embedding := make([]float32, dimension)
	for i := 0; i < dimension; i++ {
		embedding[i] = float32(rng.Float64()*2 - 1) // Values between -1 and 1
	}
	
	// Normalize the vector
	sum := float32(0)
	for _, v := range embedding {
		sum += v * v
	}
	
	magnitude := float32(0)
	if sum > 0 {
		magnitude = float32(1.0 / float64(sum))
	}
	
	for i := range embedding {
		embedding[i] *= magnitude
	}
	
	return embedding
}

// updateUsageMetrics updates the usage metrics for a model
func (p *BaseMockProviderImpl) updateUsageMetrics(modelID string, tokens int64, duration time.Duration, err error) {
	p.usageMetricsMutex.Lock()
	defer p.usageMetricsMutex.Unlock()
	
	metrics, ok := p.usageMetrics[modelID]
	if !ok {
		metrics = core.NewUsageMetrics(modelID)
		p.usageMetrics[modelID] = metrics
	}
	
	metrics.AddRequest(tokens, duration, err)
}

// GetUsageMetrics gets the usage metrics for a model
func (p *BaseMockProviderImpl) GetUsageMetrics(modelID string) *core.UsageMetrics {
	p.usageMetricsMutex.RLock()
	defer p.usageMetricsMutex.RUnlock()
	
	if metrics, ok := p.usageMetrics[modelID]; ok {
		return metrics
	}
	
	return core.NewUsageMetrics(modelID)
}

// GetAllUsageMetrics gets all usage metrics
func (p *BaseMockProviderImpl) GetAllUsageMetrics() map[string]*core.UsageMetrics {
	p.usageMetricsMutex.RLock()
	defer p.usageMetricsMutex.RUnlock()
	
	// Create a copy to avoid concurrent modification
	metricsCopy := make(map[string]*core.UsageMetrics, len(p.usageMetrics))
	for k, v := range p.usageMetrics {
		metricsCopy[k] = v
	}
	
	return metricsCopy
}
