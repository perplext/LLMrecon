// Package core provides the core interfaces and types for the Multi-Provider LLM Integration Framework.
package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BaseProvider is a base implementation of the Provider interface
// that can be embedded in specific provider implementations
type BaseProvider struct {
	// providerType is the type of provider
	providerType ProviderType
	// config is the configuration for the provider
	config *ProviderConfig
	// models is a cache of available models
	models []ModelInfo
	// modelsMutex is a mutex for concurrent access to models
	modelsMutex sync.RWMutex
	// modelsLastUpdated is the time when models were last updated
	modelsLastUpdated time.Time
	// modelsCacheTTL is the TTL for the models cache
	modelsCacheTTL time.Duration
	// capabilities is a map of supported capabilities
	capabilities map[ModelCapability]bool
	// capabilitiesMutex is a mutex for concurrent access to capabilities
	capabilitiesMutex sync.RWMutex
	// usageMetrics is a map of model ID to usage metrics
	usageMetrics map[string]*UsageMetrics
	// usageMetricsMutex is a mutex for concurrent access to usage metrics
	usageMetricsMutex sync.RWMutex
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(providerType ProviderType, config *ProviderConfig) *BaseProvider {
	if config == nil {
		config = &ProviderConfig{
			Type: providerType,
		}
	}

	return &BaseProvider{
		providerType:   providerType,
		config:         config,
		models:         make([]ModelInfo, 0),
		modelsCacheTTL: 1 * time.Hour,
		capabilities:   make(map[ModelCapability]bool),
		usageMetrics:   make(map[string]*UsageMetrics),
	}
}

// GetType returns the type of provider
func (p *BaseProvider) GetType() ProviderType {
	return p.providerType
}

// GetConfig returns the configuration for the provider
func (p *BaseProvider) GetConfig() *ProviderConfig {
	return p.config
}

// SetModels sets the available models
func (p *BaseProvider) SetModels(models []ModelInfo) {
	p.modelsMutex.Lock()
	defer p.modelsMutex.Unlock()

	p.models = models
	p.modelsLastUpdated = time.Now()

	// Update capabilities based on models
	p.capabilitiesMutex.Lock()
	defer p.capabilitiesMutex.Unlock()

	p.capabilities = make(map[ModelCapability]bool)
	for _, model := range models {
		for _, capability := range model.Capabilities {
			p.capabilities[capability] = true
		}
	}
}

// GetModels returns a list of available models
func (p *BaseProvider) GetModels(ctx context.Context) ([]ModelInfo, error) {
	p.modelsMutex.RLock()
	modelsLastUpdated := p.modelsLastUpdated
	p.modelsMutex.RUnlock()

	// Check if models cache is expired
	if time.Since(modelsLastUpdated) > p.modelsCacheTTL {
		// Cache is expired, but we'll return the cached models
		// and update the cache asynchronously
		go p.updateModels(context.Background())
	}

	p.modelsMutex.RLock()
	defer p.modelsMutex.RUnlock()

	// Return a copy of the models slice to prevent concurrent modification
	modelsCopy := make([]ModelInfo, len(p.models))
	copy(modelsCopy, p.models)

	return modelsCopy, nil
}

// updateModels updates the models cache
// This method should be overridden by specific provider implementations
func (p *BaseProvider) updateModels(ctx context.Context) error {
	// This is a placeholder implementation
	// Specific providers should override this method
	return nil
}

// GetModelInfo returns information about a specific model
func (p *BaseProvider) GetModelInfo(ctx context.Context, modelID string) (*ModelInfo, error) {
	models, err := p.GetModels(ctx)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.ID == modelID {
			// Return a copy of the model to prevent modification
			modelCopy := model
			return &modelCopy, nil
		}
	}

	return nil, fmt.Errorf("model with ID %s not found", modelID)
}

// TextCompletion generates a text completion
// This method should be overridden by specific provider implementations
func (p *BaseProvider) TextCompletion(ctx context.Context, request *TextCompletionRequest) (*TextCompletionResponse, error) {
	return nil, fmt.Errorf("text completion not implemented for provider %s", p.providerType)
}

// ChatCompletion generates a chat completion
// This method should be overridden by specific provider implementations
func (p *BaseProvider) ChatCompletion(ctx context.Context, request *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return nil, fmt.Errorf("chat completion not implemented for provider %s", p.providerType)
}

// StreamingChatCompletion generates a streaming chat completion
// This method should be overridden by specific provider implementations
func (p *BaseProvider) StreamingChatCompletion(ctx context.Context, request *ChatCompletionRequest, callback func(response *ChatCompletionResponse) error) error {
	return fmt.Errorf("streaming chat completion not implemented for provider %s", p.providerType)
}

// CreateEmbedding creates an embedding
// This method should be overridden by specific provider implementations
func (p *BaseProvider) CreateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("create embedding not implemented for provider %s", p.providerType)
}

// CountTokens counts the number of tokens in a text
// This method should be overridden by specific provider implementations
func (p *BaseProvider) CountTokens(ctx context.Context, text string, modelID string) (int, error) {
	return 0, fmt.Errorf("count tokens not implemented for provider %s", p.providerType)
}

// SupportsModel returns whether the provider supports a specific model
func (p *BaseProvider) SupportsModel(ctx context.Context, modelID string) bool {
	models, err := p.GetModels(ctx)
	if err != nil {
		return false
	}

	for _, model := range models {
		if model.ID == modelID {
			return true
		}
	}

	return false
}

// SupportsCapability returns whether the provider supports a specific capability
func (p *BaseProvider) SupportsCapability(ctx context.Context, capability ModelCapability) bool {
	p.capabilitiesMutex.RLock()
	defer p.capabilitiesMutex.RUnlock()

	return p.capabilities[capability]
}

// Close closes the provider and releases any resources
func (p *BaseProvider) Close() error {
	// This is a placeholder implementation
	// Specific providers should override this method if needed
	return nil
}

// SetModelsCacheTTL sets the TTL for the models cache
func (p *BaseProvider) SetModelsCacheTTL(ttl time.Duration) {
	p.modelsMutex.Lock()
	defer p.modelsMutex.Unlock()

	p.modelsCacheTTL = ttl
}

// AddCapability adds a capability to the provider
func (p *BaseProvider) AddCapability(capability ModelCapability) {
	p.capabilitiesMutex.Lock()
	defer p.capabilitiesMutex.Unlock()

	p.capabilities[capability] = true
}

// RemoveCapability removes a capability from the provider
func (p *BaseProvider) RemoveCapability(capability ModelCapability) {
	p.capabilitiesMutex.Lock()
	defer p.capabilitiesMutex.Unlock()

	delete(p.capabilities, capability)
}

// GetCapabilities returns the capabilities supported by the provider
func (p *BaseProvider) GetCapabilities() []ModelCapability {
	p.capabilitiesMutex.RLock()
	defer p.capabilitiesMutex.RUnlock()

	capabilities := make([]ModelCapability, 0, len(p.capabilities))
	for capability := range p.capabilities {
		capabilities = append(capabilities, capability)
	}

	return capabilities
}

// ValidateConfig validates the provider configuration
func (p *BaseProvider) ValidateConfig() error {
	if p.config == nil {
		return fmt.Errorf("provider configuration is nil")
	}

	if p.config.Type != p.providerType {
		return fmt.Errorf("provider type mismatch: expected %s, got %s", p.providerType, p.config.Type)
	}

	return nil
}

// SetConfig sets the provider configuration
func (p *BaseProvider) SetConfig(config *ProviderConfig) error {
	if config == nil {
		return fmt.Errorf("provider configuration is nil")
	}

	if config.Type != p.providerType {
		return fmt.Errorf("provider type mismatch: expected %s, got %s", p.providerType, config.Type)
	}

	p.config = config
	return nil
}

// UpdateConfig updates the provider configuration
func (p *BaseProvider) UpdateConfig(updates *ProviderConfig) error {
	if updates == nil {
		return fmt.Errorf("updates is nil")
	}

	if updates.Type != p.providerType && updates.Type != "" {
		return fmt.Errorf("provider type mismatch: expected %s, got %s", p.providerType, updates.Type)
	}

	// Update non-empty fields
	if updates.APIKey != "" {
		p.config.APIKey = updates.APIKey
	}
	if updates.OrgID != "" {
		p.config.OrgID = updates.OrgID
	}
	if updates.BaseURL != "" {
		p.config.BaseURL = updates.BaseURL
	}
	if updates.Timeout > 0 {
		p.config.Timeout = updates.Timeout
	}
	if updates.DefaultModel != "" {
		p.config.DefaultModel = updates.DefaultModel
	}

	// Update retry config if provided
	if updates.RetryConfig != nil {
		if p.config.RetryConfig == nil {
			p.config.RetryConfig = updates.RetryConfig
		} else {
			if updates.RetryConfig.MaxRetries > 0 {
				p.config.RetryConfig.MaxRetries = updates.RetryConfig.MaxRetries
			}
			if updates.RetryConfig.InitialBackoff > 0 {
				p.config.RetryConfig.InitialBackoff = updates.RetryConfig.InitialBackoff
			}
			if updates.RetryConfig.MaxBackoff > 0 {
				p.config.RetryConfig.MaxBackoff = updates.RetryConfig.MaxBackoff
			}
			if updates.RetryConfig.BackoffMultiplier > 0 {
				p.config.RetryConfig.BackoffMultiplier = updates.RetryConfig.BackoffMultiplier
			}
			if len(updates.RetryConfig.RetryableStatusCodes) > 0 {
				p.config.RetryConfig.RetryableStatusCodes = updates.RetryConfig.RetryableStatusCodes
			}
		}
	}

	// Update rate limit config if provided
	if updates.RateLimitConfig != nil {
		if p.config.RateLimitConfig == nil {
			p.config.RateLimitConfig = updates.RateLimitConfig
		} else {
			if updates.RateLimitConfig.RequestsPerMinute > 0 {
				p.config.RateLimitConfig.RequestsPerMinute = updates.RateLimitConfig.RequestsPerMinute
			}
			if updates.RateLimitConfig.TokensPerMinute > 0 {
				p.config.RateLimitConfig.TokensPerMinute = updates.RateLimitConfig.TokensPerMinute
			}
			if updates.RateLimitConfig.MaxConcurrentRequests > 0 {
				p.config.RateLimitConfig.MaxConcurrentRequests = updates.RateLimitConfig.MaxConcurrentRequests
			}
			if updates.RateLimitConfig.BurstSize > 0 {
				p.config.RateLimitConfig.BurstSize = updates.RateLimitConfig.BurstSize
			}
		}
	}

	// Update additional headers if provided
	if len(updates.AdditionalHeaders) > 0 {
		if p.config.AdditionalHeaders == nil {
			p.config.AdditionalHeaders = make(map[string]string)
		}
		for k, v := range updates.AdditionalHeaders {
			p.config.AdditionalHeaders[k] = v
		}
	}

	// Update additional params if provided
	if len(updates.AdditionalParams) > 0 {
		if p.config.AdditionalParams == nil {
			p.config.AdditionalParams = make(map[string]interface{})
		}
		for k, v := range updates.AdditionalParams {
			p.config.AdditionalParams[k] = v
		}
	}

	return nil
}

// GetRateLimitConfig returns the rate limit configuration
func (p *BaseProvider) GetRateLimitConfig() *RateLimitConfig {
	if p.config == nil {
		return nil
	}
	return p.config.RateLimitConfig
}

// UpdateRateLimitConfig updates the rate limit configuration
func (p *BaseProvider) UpdateRateLimitConfig(config *RateLimitConfig) error {
	if p.config == nil {
		return fmt.Errorf("provider configuration is nil")
	}
	p.config.RateLimitConfig = config
	return nil
}

// GetRetryConfig returns the retry configuration
func (p *BaseProvider) GetRetryConfig() *RetryConfig {
	if p.config == nil {
		return nil
	}
	return p.config.RetryConfig
}

// UpdateRetryConfig updates the retry configuration
func (p *BaseProvider) UpdateRetryConfig(config *RetryConfig) error {
	if p.config == nil {
		return fmt.Errorf("provider configuration is nil")
	}
	p.config.RetryConfig = config
	return nil
}

// GetUsageMetrics returns the usage metrics for a specific model
func (p *BaseProvider) GetUsageMetrics(modelID string) (*UsageMetrics, error) {
	p.usageMetricsMutex.RLock()
	defer p.usageMetricsMutex.RUnlock()

	metrics, ok := p.usageMetrics[modelID]
	if !ok {
		return nil, fmt.Errorf("no usage metrics found for model %s", modelID)
	}

	// Return a copy to prevent external modification
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// GetAllUsageMetrics returns the usage metrics for all models
func (p *BaseProvider) GetAllUsageMetrics() (map[string]*UsageMetrics, error) {
	p.usageMetricsMutex.RLock()
	defer p.usageMetricsMutex.RUnlock()

	// Create a copy of the map to prevent external modification
	metricsCopy := make(map[string]*UsageMetrics)
	for modelID, metrics := range p.usageMetrics {
		metricsCopy[modelID] = &UsageMetrics{
			Requests:             metrics.Requests,
			Tokens:               metrics.Tokens,
			Errors:               metrics.Errors,
			LastRequestTime:      metrics.LastRequestTime,
			TotalRequestDuration: metrics.TotalRequestDuration,
			AverageResponseTime:  metrics.AverageResponseTime,
			TokensPerMinute:      metrics.TokensPerMinute,
			RequestsPerMinute:    metrics.RequestsPerMinute,
			ModelID:              metrics.ModelID,
		}
	}

	return metricsCopy, nil
}

// ResetUsageMetrics resets the usage metrics
func (p *BaseProvider) ResetUsageMetrics() error {
	p.usageMetricsMutex.Lock()
	defer p.usageMetricsMutex.Unlock()

	for _, metrics := range p.usageMetrics {
		metrics.Reset()
	}

	return nil
}

// UpdateUsageMetrics updates the usage metrics for a model
func (p *BaseProvider) UpdateUsageMetrics(modelID string, tokens int64, duration time.Duration, err error) {
	p.usageMetricsMutex.Lock()
	defer p.usageMetricsMutex.Unlock()

	metrics, ok := p.usageMetrics[modelID]
	if !ok {
		metrics = NewUsageMetrics(modelID)
		p.usageMetrics[modelID] = metrics
	}

	metrics.AddRequest(tokens, duration, err)
}
