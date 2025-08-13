// Package interfaces provides interfaces for template management components
package interfaces

import (
	"context"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
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
	TemplateID string
	
	// Template is the template that was executed
	Template *format.Template
	
	// Success indicates if the template execution was successful
	Success bool
	
	// Response is the response from the LLM
	Response string
	
	// VulnerabilityDetected indicates if a vulnerability was detected
	VulnerabilityDetected bool
	
	// VulnerabilityScore is the vulnerability score (0-100)
	VulnerabilityScore int
	
	// VulnerabilityDetails contains details about the detected vulnerability
	VulnerabilityDetails map[string]interface{}
	
	// Error is the error that occurred during execution, if any
	Error error
	
	// ExecutionTime is the time it took to execute the template
	ExecutionTime int64
	
	// Timestamp is the time the template was executed
	Timestamp int64
	
	// Status is the status of the template execution
	Status string
	
	// StartTime is the time the template execution started
	StartTime time.Time
	
	// EndTime is the time the template execution ended
	EndTime time.Time
	
	// Duration is the duration of the template execution
	Duration time.Duration
	
	// Details contains additional details about the execution
	Details map[string]interface{}
	
	// Detected is a shorthand for VulnerabilityDetected
	Detected bool
	
	// Score is a shorthand for VulnerabilityScore
	Score int
	
	// CompletionTime is when the execution completed
	CompletionTime time.Time
	
	// Provider is the LLM provider used
	Provider string
	
	// ProviderOptions are the options passed to the provider
	ProviderOptions map[string]interface{}
	
	// FromCache indicates if the result was served from cache
	FromCache bool
}

// TemplateExecutor is the interface for executing templates
type TemplateExecutor interface {
	// Execute executes a template
	Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*TemplateResult, error)
	
	// ExecuteBatch executes multiple templates
	ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*TemplateResult, error)
}

// TemplateParser is the interface for parsing templates
type TemplateParser interface {
	// Parse parses a template
	Parse(template *format.Template) error
	
	// Validate validates a template
	Validate(template *format.Template) error
	
	// ResolveVariables resolves variables in a template
	ResolveVariables(template *format.Template, variables map[string]interface{}) error
}

// TemplateReporter is the interface for generating reports
type TemplateReporter interface {
	// GenerateReport generates a report for template execution results
	GenerateReport(results []*TemplateResult, format string) ([]byte, error)
}

// TemplateCache is the interface for caching templates
type TemplateCache interface {
	// Get gets a template from the cache
	Get(id string) (*format.Template, bool)
	// Set sets a template in the cache
	Set(id string, template *format.Template)
	// Delete deletes a template from the cache
	Delete(id string)
	// Clear clears the cache
	Clear()
	// GetStats gets cache statistics
	GetStats() map[string]interface{}
	// Prune removes old entries from the cache
	Prune(maxAge time.Duration) int
}

// TemplateRegistry is the interface for registering templates
type TemplateRegistry interface {
	// Register registers a template
	Register(template *format.Template) error
	// Unregister unregisters a template
	Unregister(id string) error
	// Get gets a template from the registry
	Get(id string) (*format.Template, error)
	// List lists all templates in the registry
	List() []*format.Template
	// Update updates a template in the registry
	Update(template *format.Template) error
	// FindByTag finds templates by tag
	FindByTag(tag string) []*format.Template
	// FindByTags finds templates by tags
	FindByTags(tags []string) []*format.Template
	// GetMetadata gets metadata for a template
	GetMetadata(id string) (map[string]interface{}, error)
	// SetMetadata sets metadata for a template
	SetMetadata(id string, metadata map[string]interface{}) error
	// Count returns the number of templates in the registry
	Count() int
}

// LLMProvider is the interface for interacting with LLM systems
type LLMProvider interface {
	// SendPrompt sends a prompt to the LLM and returns the response
	SendPrompt(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
	
	// GetSupportedModels returns the list of supported models
	GetSupportedModels() []string
	
	// GetName returns the name of the provider
	GetName() string
}

// DetectionEngine is the interface for detecting vulnerabilities in LLM responses
type DetectionEngine interface {
	// Detect detects vulnerabilities in an LLM response
	Detect(ctx context.Context, template *format.Template, response string) (bool, int, map[string]interface{}, error)
	
	// GetName returns the name of the detection engine
	GetName() string
}

// RateLimiter is the interface for rate limiting requests to LLM systems
type RateLimiter interface {
	// Acquire acquires a token
	Acquire(ctx context.Context) error
	
	// AcquireForUser acquires a token for a specific user
	AcquireForUser(ctx context.Context, userID string) error
	
	// Release releases a token
	Release()
	
	// ReleaseForUser releases a token for a specific user
	ReleaseForUser(userID string)
	
	// GetLimit returns the current rate limit
	GetLimit() int
	
	// GetUserLimit returns the current rate limit for a specific user
	GetUserLimit(userID string) int
	
	// SetLimit sets the global rate limit
	SetLimit(limit int)
	
	// SetUserLimit sets the rate limit for a specific user
	SetUserLimit(userID string, limit int)
}

// SchemaValidator is the interface for validating templates against a schema
type SchemaValidator interface {
	// ValidateTemplate validates a template against the schema
	ValidateTemplate(template *format.Template) error
	// ValidateTemplateFile validates a template file against the schema
	ValidateTemplateFile(filePath string) error
	// ValidateJSON validates JSON data against the schema
	ValidateJSON(data []byte) error
	// ValidateYAML validates YAML data against the schema
	ValidateYAML(data []byte) error
}

// TemplateLoader is the interface for loading templates
type TemplateLoader interface {
	// Load loads a template from a file
	Load(filePath string) (*format.Template, error)
	// LoadFromBytes loads a template from bytes
	LoadFromBytes(data []byte, format string) (*format.Template, error)
	// LoadBatch loads multiple templates from a directory
	LoadBatch(directory string) ([]*format.Template, error)
}

// TemplateManagerInternal is the internal interface for template managers
type TemplateManagerInternal interface {
	// GetRegistry returns the template registry
	GetRegistry() TemplateRegistry
	// GetCache returns the template cache
	GetCache() TemplateCache
	// GetExecutor returns the template executor
	GetExecutor() TemplateExecutor
	// GetParser returns the template parser
	GetParser() TemplateParser
	// GetReporter returns the template reporter
	GetReporter() TemplateReporter
	// GetLoader returns the template loader
	GetLoader() TemplateLoader
}
