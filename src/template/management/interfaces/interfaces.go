package interfaces

import (
	"context"
	"io"
	"time"
)

// Template represents a template
type Template interface {
	// GetID returns the template ID
	GetID() string

	// GetName returns the template name
	GetName() string

	// GetVersion returns the template version
	GetVersion() string

	// GetContent returns the template content
	GetContent() ([]byte, error)

	// Validate validates the template
	Validate() error
}

// TemplateManager manages templates
type TemplateManager interface {
	// ListTemplates lists all templates
	ListTemplates(ctx context.Context) ([]Template, error)

	// GetTemplate gets a template by ID
	GetTemplate(ctx context.Context, id string) (Template, error)

	// CreateTemplate creates a new template
	CreateTemplate(ctx context.Context, template Template) error

	// UpdateTemplate updates an existing template
	UpdateTemplate(ctx context.Context, id string, template Template) error

	// DeleteTemplate deletes a template
	DeleteTemplate(ctx context.Context, id string) error

	// ValidateTemplate validates a template
	ValidateTemplate(ctx context.Context, template Template) error
}

// TemplateRepository provides template storage
type TemplateRepository interface {
	// List lists all templates
	List(ctx context.Context) ([]Template, error)

	// Get gets a template by ID
	Get(ctx context.Context, id string) (Template, error)

	// Create creates a new template
	Create(ctx context.Context, template Template) error

	// Update updates an existing template
	Update(ctx context.Context, id string, template Template) error

	// Delete deletes a template
	Delete(ctx context.Context, id string) error

	// Exists checks if a template exists
	Exists(ctx context.Context, id string) (bool, error)
}

// TemplateLoader loads templates from various sources
type TemplateLoader interface {
	// LoadFromFile loads a template from a file
	LoadFromFile(path string) (Template, error)

	// LoadFromReader loads a template from a reader
	LoadFromReader(reader io.Reader) (Template, error)

	// LoadFromBytes loads a template from bytes
	LoadFromBytes(data []byte) (Template, error)

	// LoadFromURL loads a template from a URL
	LoadFromURL(url string) (Template, error)
}

// TemplateCache provides template caching
type TemplateCache interface {
	// Get gets a template from cache
	Get(key string) (Template, bool)

	// Set sets a template in cache
	Set(key string, template Template, ttl time.Duration)

	// Delete deletes a template from cache
	Delete(key string)

	// Clear clears the cache
	Clear()

	// Size returns the cache size
	Size() int
}

// InputValidationRule defines a rule for validating template inputs
type InputValidationRule interface {
	// GetName returns the name of the rule
	GetName() string

	// GetDescription returns the description of the rule
	GetDescription() string

	// Validate validates a template against the rule
	Validate(ctx context.Context, template interface{}) error
}

// SchemaValidator validates templates against a schema
type SchemaValidator interface {
	// ValidateSchema validates template data against a schema
	ValidateSchema(data interface{}) error
	
	// ValidateTemplate validates a template against a schema
	ValidateTemplate(template interface{}) error
}

// TemplateValidator validates templates
type TemplateValidator interface {
	// Validate validates a template
	Validate(template Template) error

	// ValidateContent validates template content
	ValidateContent(content []byte) error

	// ValidateSchema validates against a schema
	ValidateSchema(template Template, schema interface{}) error
}

// TemplateExecutor executes templates
type TemplateExecutor interface {
	// Execute executes a template
	Execute(ctx context.Context, template Template, data interface{}) ([]byte, error)

	// ExecuteWithOptions executes with options
	ExecuteWithOptions(ctx context.Context, template Template, data interface{}, options map[string]interface{}) ([]byte, error)
}

// TemplateRegistry provides template registration
type TemplateRegistry interface {
	// Register registers a template
	Register(template Template) error

	// Unregister unregisters a template
	Unregister(id string) error

	// Get gets a registered template
	Get(id string) (Template, bool)

	// List lists all registered templates
	List() []Template

	// Clear clears the registry
	Clear()
}

// Additional TemplateSource constants for management layer
const (
	// FileSource indicates the template is from a file
	FileSource TemplateSource = "file"
	// DirectorySource indicates the template is from a directory
	DirectorySource TemplateSource = "directory"
	// GitHubSource indicates the template is from GitHub
	GitHubSource TemplateSource = "github"
	// GitLabSource indicates the template is from GitLab
	GitLabSource TemplateSource = "gitlab"
	// HTTPSource indicates the template is from HTTP
	HTTPSource TemplateSource = "http"
	// DatabaseSource indicates the template is from a database
	DatabaseSource TemplateSource = "database"
)

// TemplateStatus represents the status of a template
type TemplateStatus string

const (
	// StatusLoaded indicates the template has been loaded
	StatusLoaded TemplateStatus = "loaded"
	// StatusValidated indicates the template has been validated
	StatusValidated TemplateStatus = "validated"
	// StatusExecuting indicates the template is being executed
	StatusExecuting TemplateStatus = "executing"
	// StatusCompleted indicates the template execution has completed
	StatusCompleted TemplateStatus = "completed"
	// StatusFailed indicates the template execution has failed
	StatusFailed TemplateStatus = "failed"
)

// TemplateResult represents the result of a template execution
type TemplateResult struct {
	// TemplateID is the ID of the template
	TemplateID string `json:"template_id"`
	// TemplateName is the name of the template
	TemplateName string `json:"template_name"`
	// Description is the description of the template
	Description string `json:"description"`
	// Status is the status of the template
	Status string `json:"status"`
	// StartTime is the time the template execution started
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the template execution ended
	EndTime time.Time `json:"end_time"`
	// Duration is the duration of the template execution
	Duration time.Duration `json:"duration"`
	// Error is any error that occurred during template execution
	Error error `json:"error,omitempty"`
	// Response is the response from the LLM
	Response string `json:"response,omitempty"`
	// Detected indicates whether the vulnerability was detected
	Detected bool `json:"detected"`
	// Score is the score of the template execution (0-100)
	Score int `json:"score"`
	// Details contains additional details about the template execution
	Details map[string]interface{} `json:"details,omitempty"`
	// Tags contains tags associated with the template
	Tags []string `json:"tags,omitempty"`
	// Metadata contains additional metadata about the template
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Input contains the input to the template
	Input string `json:"input,omitempty"`
	// Output contains the expected output from the template
	Output string `json:"output,omitempty"`
}

// RateLimiter provides rate limiting functionality for template execution
type RateLimiter interface {
	// Allow checks if an operation is allowed under the current rate limit
	Allow(ctx context.Context, userID string) (bool, error)
	
	// Wait blocks until the operation is allowed under the rate limit
	Wait(ctx context.Context, userID string) error
	
	// Reserve reserves the right to perform an operation at the given time
	Reserve(ctx context.Context, userID string) (bool, time.Duration, error)
	
	// SetUserLimit sets a custom rate limit for a specific user
	SetUserLimit(userID string, qps float64, burst int) error
	
	// GetUserLimit gets the current rate limit for a specific user
	GetUserLimit(userID string) (float64, int, error)
	
	// GetStats returns statistics about the rate limiter
	GetStats() map[string]interface{}
}

// InputValidator validates template inputs before execution
type InputValidator interface {
	// Validate validates a template input
	Validate(ctx context.Context, template interface{}) error
	
	// ValidateContent validates content against validation rules
	ValidateContent(ctx context.Context, content string) error
	
	// AddRule adds a validation rule
	AddRule(rule InputValidationRule) error
	
	// RemoveRule removes a validation rule by name
	RemoveRule(name string) error
	
	// GetRules returns all validation rules
	GetRules() []InputValidationRule
	
	// SetStrictMode sets whether validation errors should fail execution
	SetStrictMode(strict bool)
	
	// IsStrictMode returns whether strict mode is enabled
	IsStrictMode() bool
}

// LLMProvider provides an interface for interacting with LLM systems
type LLMProvider interface {
	// SendPrompt sends a prompt to the LLM and returns the response
	SendPrompt(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
	
	// GetSupportedModels returns the list of supported models
	GetSupportedModels() []string
	
	// GetName returns the name of the provider
	GetName() string
	
	// GetID returns a unique identifier for the provider
	GetID() string
	
	// IsHealthy checks if the provider is healthy and available
	IsHealthy(ctx context.Context) (bool, error)
	
	// GetUsage returns usage statistics for the provider
	GetUsage() map[string]interface{}
	
	// SetTimeout sets the request timeout for the provider
	SetTimeout(timeout time.Duration)
	
	// GetTimeout gets the current request timeout
	GetTimeout() time.Duration
}

// DetectionEngine provides an interface for detecting vulnerabilities in LLM responses
type DetectionEngine interface {
	// Detect analyzes a template and response to detect vulnerabilities
	Detect(ctx context.Context, template Template, response string) (bool, int, map[string]interface{}, error)
	
	// GetName returns the name of the detection engine
	GetName() string
	
	// GetVersion returns the version of the detection engine
	GetVersion() string
	
	// GetSupportedCategories returns the vulnerability categories this engine can detect
	GetSupportedCategories() []string
	
	// Configure configures the detection engine with options
	Configure(options map[string]interface{}) error
	
	// GetConfiguration returns the current configuration
	GetConfiguration() map[string]interface{}
	
	// IsEnabled returns whether the detection engine is enabled
	IsEnabled() bool
	
	// SetEnabled enables or disables the detection engine
	SetEnabled(enabled bool)
}
