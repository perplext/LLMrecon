// Package core provides the core interfaces and types for the Multi-Provider LLM Integration Framework.
package core

import (
	"context"
	"time"
)

// ModelType represents the type of LLM model
type ModelType string

const (
	// TextCompletionModel represents a text completion model
	TextCompletionModel ModelType = "text-completion"
	// ChatModel represents a chat model
	ChatModel ModelType = "chat"
	// EmbeddingModel represents an embedding model
	EmbeddingModel ModelType = "embedding"
	// ImageGenerationModel represents an image generation model
	ImageGenerationModel ModelType = "image-generation"
	// ImageAnalysisModel represents an image analysis model
	ImageAnalysisModel ModelType = "image-analysis"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	// OpenAIProvider represents the OpenAI provider
	OpenAIProvider ProviderType = "openai"
	// AnthropicProvider represents the Anthropic provider
	AnthropicProvider ProviderType = "anthropic"
	// AzureOpenAIProvider represents the Azure OpenAI provider
	AzureOpenAIProvider ProviderType = "azure-openai"
	// HuggingFaceProvider represents the HuggingFace provider
	HuggingFaceProvider ProviderType = "huggingface"
	// LocalProvider represents a local provider (e.g., running models locally)
	LocalProvider ProviderType = "local"
	// CustomProvider represents a custom provider
	CustomProvider ProviderType = "custom"
)

// ModelCapability represents a capability of a model
type ModelCapability string

const (
	// TextCompletionCapability represents the capability to generate text completions
	TextCompletionCapability ModelCapability = "text-completion"
	// ChatCompletionCapability represents the capability to generate chat completions
	ChatCompletionCapability ModelCapability = "chat-completion"
	// EmbeddingCapability represents the capability to generate embeddings
	EmbeddingCapability ModelCapability = "embedding"
	// ImageGenerationCapability represents the capability to generate images
	ImageGenerationCapability ModelCapability = "image-generation"
	// ImageAnalysisCapability represents the capability to analyze images
	ImageAnalysisCapability ModelCapability = "image-analysis"
	// StreamingCapability represents the capability to stream responses
	StreamingCapability ModelCapability = "streaming"
	// FunctionCallingCapability represents the capability to call functions
	FunctionCallingCapability ModelCapability = "function-calling"
	// ToolUseCapability represents the capability to use tools
	ToolUseCapability ModelCapability = "tool-use"
	// JSONModeCapability represents the capability to output JSON
	JSONModeCapability ModelCapability = "json-mode"
)

// Message represents a message in a conversation
type Message struct {
	// Role is the role of the message sender (e.g., "system", "user", "assistant")
	Role string `json:"role"`
	// Content is the content of the message
	Content string `json:"content"`
	// Name is an optional name for the message sender
	Name string `json:"name,omitempty"`
	// FunctionCall is an optional function call
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
	// ToolCalls is an optional list of tool calls
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
	// Timestamp is the timestamp of the message
	Timestamp time.Time `json:"timestamp,omitempty"`

// FunctionCall represents a function call
type FunctionCall struct {
	// Name is the name of the function
	Name string `json:"name"`
	// Arguments is a JSON string of arguments
	Arguments string `json:"arguments"`

// ToolCall represents a tool call
type ToolCall struct {
	// ID is the ID of the tool call
	ID string `json:"id"`
	// Type is the type of tool call (e.g., "function")
	Type string `json:"type"`
	// Function is the function call
	Function *FunctionCall `json:"function"`

// Function represents a function definition
type Function struct {
	// Name is the name of the function
	Name string `json:"name"`
	// Description is a description of the function
	Description string `json:"description"`
	// Parameters is a JSON schema of parameters
	Parameters map[string]interface{} `json:"parameters"`

// Tool represents a tool definition
type Tool struct {
	// Type is the type of tool (e.g., "function")
	Type string `json:"type"`
	// Function is the function definition
	Function *Function `json:"function"`

// TextCompletionRequest represents a request for text completion
type TextCompletionRequest struct {
	// Prompt is the prompt for text completion
	Prompt string `json:"prompt"`
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature controls randomness (0.0 to 2.0)
	Temperature float64 `json:"temperature,omitempty"`
	// TopP controls diversity via nucleus sampling (0.0 to 1.0)
	TopP float64 `json:"top_p,omitempty"`
	// N is the number of completions to generate
	N int `json:"n,omitempty"`
	// Stop is a list of tokens at which to stop generation
	Stop []string `json:"stop,omitempty"`
	// Stream indicates whether to stream the response
	Stream bool `json:"stream,omitempty"`
	// LogProbs is the number of log probabilities to return
	LogProbs int `json:"logprobs,omitempty"`
	// Echo indicates whether to echo the prompt
	Echo bool `json:"echo,omitempty"`
	// PresencePenalty penalizes new tokens based on presence in text so far (0.0 to 2.0)
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
	// FrequencyPenalty penalizes new tokens based on frequency in text so far (0.0 to 2.0)
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	// User is an optional user identifier
	User string `json:"user,omitempty"`
	// Model is the model to use
	Model string `json:"model,omitempty"`

// TextCompletionResponse represents a response from text completion
type TextCompletionResponse struct {
	// ID is the ID of the completion
	ID string `json:"id"`
	// Object is the object type
	Object string `json:"object"`
	// Created is the Unix timestamp of when the completion was created
	Created int64 `json:"created"`
	// Model is the model used
	Model string `json:"model"`
	// Choices is the list of completion choices
	Choices []TextCompletionChoice `json:"choices"`
	// Usage is the token usage information
	Usage *TokenUsage `json:"usage,omitempty"`

// TextCompletionChoice represents a choice in a text completion response
type TextCompletionChoice struct {
	// Text is the completed text
	Text string `json:"text"`
	// Index is the index of the choice
	Index int `json:"index"`
	// LogProbs is the log probabilities
	LogProbs *LogProbs `json:"logprobs,omitempty"`
	// FinishReason is the reason the completion finished
	FinishReason string `json:"finish_reason"`

// LogProbs represents log probabilities
type LogProbs struct {
	// Tokens is the list of tokens
	Tokens []string `json:"tokens"`
	// TokenLogProbs is the list of log probabilities for tokens
	TokenLogProbs []float64 `json:"token_logprobs"`
	// TopLogProbs is a list of maps from tokens to log probabilities
	TopLogProbs []map[string]float64 `json:"top_logprobs"`
	// TextOffset is a list of offsets in the text
	TextOffset []int `json:"text_offset"`

// ChatCompletionRequest represents a request for chat completion
type ChatCompletionRequest struct {
	// Messages is the list of messages in the conversation
	Messages []Message `json:"messages"`
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature controls randomness (0.0 to 2.0)
	Temperature float64 `json:"temperature,omitempty"`
	// TopP controls diversity via nucleus sampling (0.0 to 1.0)
	TopP float64 `json:"top_p,omitempty"`
	// N is the number of completions to generate
	N int `json:"n,omitempty"`
	// Stop is a list of tokens at which to stop generation
	Stop []string `json:"stop,omitempty"`
	// Stream indicates whether to stream the response
	Stream bool `json:"stream,omitempty"`
	// PresencePenalty penalizes new tokens based on presence in text so far (0.0 to 2.0)
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
	// FrequencyPenalty penalizes new tokens based on frequency in text so far (0.0 to 2.0)
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	// LogitBias is a map from token IDs to bias values (-100 to 100)
	LogitBias map[string]float64 `json:"logit_bias,omitempty"`
	// User is an optional user identifier
	User string `json:"user,omitempty"`
	// Functions is a list of function definitions
	Functions []Function `json:"functions,omitempty"`
	// FunctionCall controls function calling behavior
	FunctionCall interface{} `json:"function_call,omitempty"`
	// Tools is a list of tool definitions
	Tools []Tool `json:"tools,omitempty"`
	// ToolChoice controls tool choice behavior
	ToolChoice interface{} `json:"tool_choice,omitempty"`
	// ResponseFormat specifies the format of the response
	ResponseFormat map[string]string `json:"response_format,omitempty"`
	// Model is the model to use
	Model string `json:"model,omitempty"`
	// Metadata is optional metadata for the request
	Metadata map[string]interface{} `json:"metadata,omitempty"`

// ChatCompletionResponse represents a response from chat completion
type ChatCompletionResponse struct {
	// ID is the ID of the completion
	ID string `json:"id"`
	// Object is the object type
	Object string `json:"object"`
	// Created is the Unix timestamp of when the completion was created
	Created int64 `json:"created"`
	// Model is the model used
	Model string `json:"model"`
	// Choices is the list of completion choices
	Choices []ChatCompletionChoice `json:"choices"`
	// Usage is the token usage information
	Usage *TokenUsage `json:"usage,omitempty"`

// ChatCompletionChoice represents a choice in a chat completion response
type ChatCompletionChoice struct {
	// Index is the index of the choice
	Index int `json:"index"`
	// Message is the message
	Message Message `json:"message"`
	// FinishReason is the reason the completion finished
	FinishReason string `json:"finish_reason"`

// EmbeddingRequest represents a request for embeddings
type EmbeddingRequest struct {
	// Input is the input to embed
	Input interface{} `json:"input"`
	// Model is the model to use
	Model string `json:"model"`
	// EncodingFormat is the encoding format
	EncodingFormat string `json:"encoding_format,omitempty"`
	// User is an optional user identifier
	User string `json:"user,omitempty"`
	// Dimensions is the number of dimensions for the embeddings
	Dimensions int `json:"dimensions,omitempty"`

// EmbeddingResponse represents a response from embedding
type EmbeddingResponse struct {
	// Object is the object type
	Object string `json:"object"`
	// Data is the list of embeddings
	Data []Embedding `json:"data"`
	// Model is the model used
	Model string `json:"model"`
	// Usage is the token usage information
	Usage *TokenUsage `json:"usage,omitempty"`

// Embedding represents an embedding
type Embedding struct {
	// Object is the object type
	Object string `json:"object"`
	// Embedding is the embedding vector
	Embedding []float64 `json:"embedding"`
	// Index is the index of the embedding
	Index int `json:"index"`

// TokenUsage represents token usage information
type TokenUsage struct {
	// PromptTokens is the number of tokens in the prompt
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens in the completion
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is the total number of tokens
	TotalTokens int `json:"total_tokens"`

// ProviderConfig represents the configuration for a provider
type ProviderConfig struct {
	// Type is the type of provider
	Type ProviderType `json:"type"`
	// APIKey is the API key for the provider
	APIKey string `json:"api_key,omitempty"`
	// OrgID is the organization ID for the provider
	OrgID string `json:"org_id,omitempty"`
	// BaseURL is the base URL for the provider API
	BaseURL string `json:"base_url,omitempty"`
	// Timeout is the timeout for requests to the provider
	Timeout time.Duration `json:"timeout,omitempty"`
	// RetryConfig is the configuration for retries
	RetryConfig *RetryConfig `json:"retry_config,omitempty"`
	// RateLimitConfig is the configuration for rate limiting
	RateLimitConfig *RateLimitConfig `json:"rate_limit_config,omitempty"`
	// DefaultModel is the default model to use
	DefaultModel string `json:"default_model,omitempty"`
	// AdditionalHeaders is a map of additional headers to include in requests
	AdditionalHeaders map[string]string `json:"additional_headers,omitempty"`
	// AdditionalParams is a map of additional parameters to include in requests
	AdditionalParams map[string]interface{} `json:"additional_params,omitempty"`

// RetryConfig represents the configuration for retries
type RetryConfig struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries"`
	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration `json:"initial_backoff"`
	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration `json:"max_backoff"`
	// BackoffMultiplier is the multiplier for backoff duration
	BackoffMultiplier float64 `json:"backoff_multiplier"`
	// RetryableStatusCodes is a list of HTTP status codes that should be retried
	RetryableStatusCodes []int `json:"retryable_status_codes"`

// RateLimitConfig represents the configuration for rate limiting
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum number of requests per minute
	RequestsPerMinute int `json:"requests_per_minute"`
	// TokensPerMinute is the maximum number of tokens per minute
	TokensPerMinute int `json:"tokens_per_minute"`
	// MaxConcurrentRequests is the maximum number of concurrent requests
	MaxConcurrentRequests int `json:"max_concurrent_requests"`
	// BurstSize is the maximum burst size
	BurstSize int `json:"burst_size"`

// ModelInfo represents information about a model
type ModelInfo struct {
	// ID is the ID of the model
	ID string `json:"id"`
	// Provider is the provider of the model
	Provider ProviderType `json:"provider"`
	// Type is the type of model
	Type ModelType `json:"type"`
	// Capabilities is a list of capabilities of the model
	Capabilities []ModelCapability `json:"capabilities"`
	// MaxTokens is the maximum number of tokens the model can process
	MaxTokens int `json:"max_tokens"`
	// TrainingCutoff is the training cutoff date of the model
	TrainingCutoff time.Time `json:"training_cutoff,omitempty"`
	// Version is the version of the model
	Version string `json:"version,omitempty"`
	// Description is a description of the model
	Description string `json:"description,omitempty"`
	// PricingInfo is information about pricing for the model
	PricingInfo *ModelPricingInfo `json:"pricing_info,omitempty"`

// ModelPricingInfo represents pricing information for a model
type ModelPricingInfo struct {
	// InputPricePerToken is the price per input token
	InputPricePerToken float64 `json:"input_price_per_token"`
	// OutputPricePerToken is the price per output token
	OutputPricePerToken float64 `json:"output_price_per_token"`
	// Currency is the currency of the prices
	Currency string `json:"currency"`

// ProviderError represents an error from a provider
type ProviderError struct {
	// StatusCode is the HTTP status code
	StatusCode int `json:"status_code,omitempty"`
	// Type is the type of error
	Type string `json:"type,omitempty"`
	// Message is the error message
	Message string `json:"message"`
	// Param is the parameter that caused the error
	Param string `json:"param,omitempty"`
	// Code is the error code
	Code string `json:"code,omitempty"`
	// RawResponse is the raw response from the provider
	RawResponse string `json:"-"`

// Error returns the error message
func (e *ProviderError) Error() string {
	return e.Message

// Provider is the interface that all LLM providers must implement
type Provider interface {
	// GetType returns the type of provider
	GetType() ProviderType
	// GetConfig returns the configuration for the provider
	GetConfig() *ProviderConfig
	// GetModels returns a list of available models
	GetModels(ctx context.Context) ([]ModelInfo, error)
	// GetModelInfo returns information about a specific model
	GetModelInfo(ctx context.Context, modelID string) (*ModelInfo, error)
	// TextCompletion generates a text completion
	TextCompletion(ctx context.Context, request *TextCompletionRequest) (*TextCompletionResponse, error)
	// ChatCompletion generates a chat completion
	ChatCompletion(ctx context.Context, request *ChatCompletionRequest) (*ChatCompletionResponse, error)
	// StreamingChatCompletion generates a streaming chat completion
	StreamingChatCompletion(ctx context.Context, request *ChatCompletionRequest, callback func(response *ChatCompletionResponse) error) error
	// CreateEmbedding creates an embedding
	CreateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error)
	// CountTokens counts the number of tokens in a text
	CountTokens(ctx context.Context, text string, modelID string) (int, error)
	// SupportsModel returns whether the provider supports a specific model
	SupportsModel(ctx context.Context, modelID string) bool
	// SupportsCapability returns whether the provider supports a specific capability
	SupportsCapability(ctx context.Context, capability ModelCapability) bool
	// Close closes the provider and releases any resources
	Close() error

	// Resilience and monitoring methods

	// GetRateLimitConfig returns the rate limit configuration
	GetRateLimitConfig() *RateLimitConfig
	// UpdateRateLimitConfig updates the rate limit configuration
	UpdateRateLimitConfig(config *RateLimitConfig) error
	// GetRetryConfig returns the retry configuration
	GetRetryConfig() *RetryConfig
	// UpdateRetryConfig updates the retry configuration
	UpdateRetryConfig(config *RetryConfig) error
	// GetUsageMetrics returns the usage metrics for a specific model
	GetUsageMetrics(modelID string) (*UsageMetrics, error)
	// GetAllUsageMetrics returns the usage metrics for all models
	GetAllUsageMetrics() (map[string]*UsageMetrics, error)
	// ResetUsageMetrics resets the usage metrics
	ResetUsageMetrics() error

// ProviderFactory is the interface for creating providers
type ProviderFactory interface {
	// CreateProvider creates a provider with the given configuration
	CreateProvider(config *ProviderConfig) (Provider, error)
	// GetSupportedProviderTypes returns the provider types supported by this factory
	GetSupportedProviderTypes() []ProviderType

// ProviderRegistry is the interface for registering and retrieving providers
type ProviderRegistry interface {
	// RegisterProvider registers a provider
	RegisterProvider(provider Provider) error
	
	// RegisterProviderFactory registers a provider factory
	RegisterProviderFactory(factory ProviderFactory) error
	
	// GetProvider returns a provider by type
	GetProvider(providerType ProviderType) (Provider, error)
	
	// GetProviderByModel returns a provider that supports a specific model
	GetProviderByModel(modelID string) (Provider, error)
	
	// GetProviderByCapability returns a provider that supports a specific capability
	GetProviderByCapability(capability ModelCapability) (Provider, error)
	
	// GetAllProviders returns all registered providers
	GetAllProviders() []Provider
	
	// GetAllProviderTypes returns all registered provider types
	GetAllProviderTypes() []ProviderType

// ModelRegistry is the interface for registering and retrieving models
type ModelRegistry interface {
	// RegisterModel registers a model
	RegisterModel(model *ModelInfo) error
	
	// GetModel returns a model by ID
	GetModel(modelID string) (*ModelInfo, error)
	
	// GetModelsByProvider returns models by provider
	GetModelsByProvider(providerType ProviderType) ([]*ModelInfo, error)
	
	// GetModelsByType returns models by type
	GetModelsByType(modelType ModelType) ([]*ModelInfo, error)
	
	// GetModelsByCapability returns models by capability
	GetModelsByCapability(capability ModelCapability) ([]*ModelInfo, error)
	
	// GetAllModels returns all registered models
