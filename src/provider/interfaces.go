package provider

import "context"

// Provider interface for LLM providers
type Provider interface {
	GetName() string
	GetVersion() string
	IsConfigured() bool
	Configure(config map[string]interface{}) error
	Execute(ctx context.Context, request *Request) (*Response, error)
	Validate() error
	Close() error

// Factory interface for creating provider instances
type Factory interface {
	CreateProvider(name string, config map[string]interface{}) (Provider, error)
	GetSupportedProviders() []string
	GetProviderInfo(name string) (*ProviderInfo, error)
	RegisterProvider(name string, creator ProviderCreator) error

// ProviderCreator is a function that creates a provider instance
type ProviderCreator func(config map[string]interface{}) (Provider, error)

// Request represents a request to a provider
type Request struct {
	Method     string                 `json:"method"`
	Endpoint   string                 `json:"endpoint"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Body       interface{}            `json:"body,omitempty"`
	Timeout    int                    `json:"timeout,omitempty"`

// Response represents a response from a provider
type Response struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       interface{}            `json:"body,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderInfo contains information about a provider
type ProviderInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Capabilities []string          `json:"capabilities"`
	ConfigSchema map[string]string `json:"config_schema"`
