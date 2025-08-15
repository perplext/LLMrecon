package interfaces

import (
	"io"
)

// TemplateSource represents a source for templates
type TemplateSource string

const (
	// SourceFile indicates the template is from a file
	SourceFile TemplateSource = "file"
	// SourceURL indicates the template is from a URL
	SourceURL TemplateSource = "url"
	// SourceBytes indicates the template is from bytes
	SourceBytes TemplateSource = "bytes"
	// SourceReader indicates the template is from a reader
	SourceReader TemplateSource = "reader"
)

// LoaderOptions represents options for template loading
type LoaderOptions struct {
	// Source is the source type
	Source TemplateSource
	// ValidateOnLoad indicates if validation should occur on load
	ValidateOnLoad bool
	// CacheEnabled indicates if caching is enabled
	CacheEnabled bool
}

// TemplateLoaderExt extends the basic loader interface
type TemplateLoaderExt interface {
	TemplateLoader
	
	// LoadWithOptions loads a template with options
	LoadWithOptions(source interface{}, options LoaderOptions) (Template, error)
	
	// LoadMultiple loads multiple templates
	LoadMultiple(sources []interface{}) ([]Template, error)
	
	// ValidateSource validates a template source
	ValidateSource(source interface{}) error
}

// DefaultLoader provides a default template loader implementation
type DefaultLoader struct {
	validator TemplateValidator
	cache     TemplateCache
}

// NewDefaultLoader creates a new default loader
func NewDefaultLoader(validator TemplateValidator, cache TemplateCache) *DefaultLoader {
	return &DefaultLoader{
		validator: validator,
		cache:     cache,
	}
}

// LoadFromFile loads a template from a file
func (l *DefaultLoader) LoadFromFile(path string) (Template, error) {
	// Implementation would go here
	return nil, nil
}

// LoadFromReader loads a template from a reader
func (l *DefaultLoader) LoadFromReader(reader io.Reader) (Template, error) {
	// Implementation would go here
	return nil, nil
}

// LoadFromBytes loads a template from bytes
func (l *DefaultLoader) LoadFromBytes(data []byte) (Template, error) {
	// Implementation would go here
	return nil, nil
}

// LoadFromURL loads a template from a URL
func (l *DefaultLoader) LoadFromURL(url string) (Template, error) {
	// Implementation would go here
	return nil, nil
}
