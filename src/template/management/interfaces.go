package management

import "context"

// Template represents a security test template
type Template interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetCategory() string
	GetSeverity() string
	GetAuthor() string
	GetVersion() string
	GetTags() []string
	GetReferences() []string
	GetMetadata() map[string]interface{}
	Validate() error

// TemplateManager manages templates
type TemplateManager interface {
	GetTemplate(id string) (Template, error)
	ListTemplates() ([]Template, error)
	GetCategories() ([]string, error)
	LoadTemplate(path string) (Template, error)
	ValidateTemplate(template Template) error

// Engine represents a template execution engine
type Engine interface {
	Execute(ctx context.Context, template Template, target interface{}) (interface{}, error)
	Validate(template Template) error
	GetSupportedFormats() []string
