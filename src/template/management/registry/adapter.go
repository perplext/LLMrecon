package registry

import (
	"context"
	"fmt"
	
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// RegistryAdapter adapts the registry to the template manager interface
type RegistryAdapter struct {
	registry interfaces.TemplateRegistry
}

// NewRegistryAdapter creates a new registry adapter
func NewRegistryAdapter(registry interfaces.TemplateRegistry) *RegistryAdapter {
	return &RegistryAdapter{
		registry: registry,
	}
}

// Register registers a template
func (a *RegistryAdapter) Register(template interfaces.Template) error {
	return a.registry.Register(template)
}

// Unregister unregisters a template
func (a *RegistryAdapter) Unregister(id string) error {
	return a.registry.Unregister(id)
}

// Get gets a registered template
func (a *RegistryAdapter) Get(id string) (interfaces.Template, bool) {
	return a.registry.Get(id)
}

// List lists all registered templates
func (a *RegistryAdapter) List() []interfaces.Template {
	return a.registry.List()
}

// Clear clears the registry
func (a *RegistryAdapter) Clear() {
	a.registry.Clear()
}

// ListTemplates lists all templates
func (a *RegistryAdapter) ListTemplates(ctx context.Context) ([]interfaces.Template, error) {
	return a.registry.List(), nil
}

// GetTemplate gets a template by ID
func (a *RegistryAdapter) GetTemplate(ctx context.Context, id string) (interfaces.Template, error) {
	template, ok := a.registry.Get(id)
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil

// CreateTemplate creates a new template
func (a *RegistryAdapter) CreateTemplate(ctx context.Context, template interfaces.Template) error {
	return a.registry.Register(template)
}

// UpdateTemplate updates an existing template
func (a *RegistryAdapter) UpdateTemplate(ctx context.Context, id string, template interfaces.Template) error {
	if err := a.registry.Unregister(id); err != nil {
		return err
	}
	return a.registry.Register(template)

// DeleteTemplate deletes a template
func (a *RegistryAdapter) DeleteTemplate(ctx context.Context, id string) error {
	return a.registry.Unregister(id)
}

// ValidateTemplate validates a template
func (a *RegistryAdapter) ValidateTemplate(ctx context.Context, template interfaces.Template) error {
	return template.Validate()
}
}
