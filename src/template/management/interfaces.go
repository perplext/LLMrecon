package management

import (
	"context"
	"github.com/perplext/LLMrecon/src/template/format"
)

// Template is an alias for format.Template to maintain consistency
type Template = format.Template

// TemplateManager manages templates
type TemplateManager interface {
	GetTemplate(id string) (Template, error)
	ListTemplates() ([]Template, error)
	GetCategories() ([]string, error)
	LoadTemplate(path string) (Template, error)
	ValidateTemplate(template Template) error
}

// Engine represents a template execution engine
type Engine interface {
	Execute(ctx context.Context, template Template, target interface{}) (interface{}, error)
	Validate(template Template) error
	GetSupportedFormats() []string
}

// TemplateLoader represents a template loader interface
type TemplateLoader interface {
	LoadFromPath(ctx context.Context, path string, recursive bool) ([]*Template, error)
}

// TemplateParser represents a template parser interface
type TemplateParser interface {
	Parse(data []byte) (*Template, error)
	Validate(template *Template) error
}

// TemplateExecutor represents a template executor interface
type TemplateExecutor interface {
	Execute(ctx context.Context, template *Template, options map[string]interface{}) (interface{}, error)
	RegisterProvider(provider interface{})
}

// TemplateReporter represents a template reporter interface
type TemplateReporter interface {
	GenerateReport(results interface{}, format string) ([]byte, error)
}

// TemplateCache represents a template cache interface
type TemplateCache interface {
	Get(key string) (*Template, bool)
	Set(key string, template *Template)
}

// TemplateRegistry represents a template registry interface
type TemplateRegistry interface {
	Register(template *Template) error
	Get(id string) (*Template, error)
	List() []*Template
	FindByTag(tag string) []*Template
	FindByTags(tags []string) []*Template
}

