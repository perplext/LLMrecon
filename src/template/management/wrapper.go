package management

import (
	"context"
	"fmt"
)

// TemplateManagerWrapper wraps DefaultTemplateManager to implement the TemplateManager interface
type TemplateManagerWrapper struct {
	manager *DefaultTemplateManager
}

// NewTemplateManagerWrapper creates a new wrapper for DefaultTemplateManager
func NewTemplateManagerWrapper(manager *DefaultTemplateManager) TemplateManager {
	return &TemplateManagerWrapper{manager: manager}
}

// GetTemplate implements TemplateManager interface
func (w *TemplateManagerWrapper) GetTemplate(id string) (Template, error) {
	template, err := w.manager.GetTemplate(id)
	if err != nil {
		return Template{}, err
	}
	if template == nil {
		return Template{}, fmt.Errorf("template not found: %s", id)
	}
	return *template, nil
}

// ListTemplates implements TemplateManager interface
func (w *TemplateManagerWrapper) ListTemplates() ([]Template, error) {
	templates := w.manager.ListTemplates()
	if templates == nil {
		return []Template{}, nil
	}
	result := make([]Template, 0, len(templates))
	for _, template := range templates {
		if template != nil {
			result = append(result, *template)
		}
	}
	return result, nil
}

// GetCategories implements TemplateManager interface
func (w *TemplateManagerWrapper) GetCategories() ([]string, error) {
	return w.manager.GetCategories()
}

// LoadTemplate implements TemplateManager interface
func (w *TemplateManagerWrapper) LoadTemplate(path string) (Template, error) {
	template, err := w.manager.LoadTemplate(context.Background(), path, "file")
	if err != nil {
		return Template{}, err
	}
	if template == nil {
		return Template{}, fmt.Errorf("failed to load template from path: %s", path)
	}
	return *template, nil
}

// ValidateTemplate implements TemplateManager interface
func (w *TemplateManagerWrapper) ValidateTemplate(template Template) error {
	// Since Template is an alias for format.Template, we can use it directly
	// Pass a pointer to the template to match the expected signature
	return w.manager.ValidateTemplate(&template)
}
