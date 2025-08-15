package management

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateManagerWrapper wraps DefaultTemplateManager to implement the TemplateManager interface
type TemplateManagerWrapper struct {
	manager *DefaultTemplateManager
}

// NewTemplateManagerWrapper creates a new wrapper for DefaultTemplateManager
func NewTemplateManagerWrapper(manager *DefaultTemplateManager) TemplateManager {
	return &TemplateManagerWrapper{manager: manager}

// GetTemplate implements TemplateManager interface
func (w *TemplateManagerWrapper) GetTemplate(id string) (Template, error) {
	template, err := w.manager.GetTemplate(id)
	if err != nil {
		return nil, err
	}
	return template, nil

// ListTemplates implements TemplateManager interface
func (w *TemplateManagerWrapper) ListTemplates() ([]Template, error) {
	templates := w.manager.ListTemplates()
	result := make([]Template, len(templates))
	for i, template := range templates {
		result[i] = template
	}
	return result, nil

// GetCategories implements TemplateManager interface
func (w *TemplateManagerWrapper) GetCategories() ([]string, error) {
	return w.manager.GetCategories()

// LoadTemplate implements TemplateManager interface
func (w *TemplateManagerWrapper) LoadTemplate(path string) (Template, error) {
	template, err := w.manager.LoadTemplate(context.Background(), path, "file")
	if err != nil {
		return nil, err
	}
	return template, nil

// ValidateTemplate implements TemplateManager interface
func (w *TemplateManagerWrapper) ValidateTemplate(template Template) error {
	if formatTemplate, ok := template.(*format.Template); ok {
		return w.manager.ValidateTemplate(formatTemplate)
	}
}
}
}
}
}
