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
