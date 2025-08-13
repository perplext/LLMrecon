// Package types provides common types and interfaces for template management.
package types

import (
	"context"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// TemplateSource represents a source of templates
type TemplateSource struct {
	// Path is the path to the source
	Path string
	// Type is the type of the source
	Type string
}

// TemplateLoader defines the interface for loading templates
type TemplateLoader interface {
	// LoadTemplate loads a template from a source
	LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error)
	
	// LoadTemplates loads multiple templates from a source
	LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error)
}

// OptimizedTemplateLoader extends TemplateLoader with additional optimized methods
type OptimizedTemplateLoader interface {
	TemplateLoader
	
	// LoadTemplateWithTimeout loads a template with a timeout
	LoadTemplateWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) (*format.Template, error)
	
	// LoadTemplatesWithTimeout loads multiple templates with a timeout
	LoadTemplatesWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) ([]*format.Template, error)
	
	// ClearCache clears the template cache
	ClearCache()
	
	// GetCacheStats returns statistics about the cache
	GetCacheStats() map[string]interface{}
	
	// GetLoaderStats returns statistics about the loader
	GetLoaderStats() map[string]interface{}
	
	// SetConcurrencyLimit sets the concurrency limit
	SetConcurrencyLimit(limit int)
}

// TemplateManager defines the interface for managing templates
type TemplateManager interface {
	// LoadTemplate loads a template from a source
	LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error)
	
	// LoadTemplates loads multiple templates from a source
	LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error)
	
	// Execute executes a template
	Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*interfaces.TemplateResult, error)
	
	// ExecuteBatch executes multiple templates
	ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*interfaces.TemplateResult, error)
	
	// GetLoader returns the template loader
	GetLoader() TemplateLoader
	
	// GetExecutor returns the template executor
	GetExecutor() interfaces.TemplateExecutor
}

// OptimizedTemplateManager extends TemplateManager with additional optimized methods
type OptimizedTemplateManager interface {
	TemplateManager
	
	// ClearCache clears the template cache
	ClearCache()
	
	// GetStats returns statistics about the manager
	GetStats() map[string]interface{}
	
	// GetTemplateStats returns statistics about a specific template
	GetTemplateStats(templateID string) map[string]interface{}
	
	// GetTemplateIDs returns all template IDs
	GetTemplateIDs() []string
	
	// SetConcurrencyLimit sets the concurrency limit
	SetConcurrencyLimit(limit int)
	
	// SetCacheTTL sets the cache TTL
	SetCacheTTL(ttl time.Duration)
	
	// SetCacheSize sets the cache size
	SetCacheSize(size int)
	
	// SetExecutionTimeout sets the execution timeout
	SetExecutionTimeout(timeout time.Duration)
	
	// SetLoadTimeout sets the load timeout
	SetLoadTimeout(timeout time.Duration)
	
	// SetDebug sets the debug flag
	SetDebug(debug bool)
}
