package registry

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/template/format"
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
}

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
}

// DeleteTemplate deletes a template

func (a *RegistryAdapter) DeleteTemplate(ctx context.Context, id string) error {
	return a.registry.Unregister(id)
}

// ValidateTemplate validates a template
func (a *RegistryAdapter) ValidateTemplate(ctx context.Context, template interfaces.Template) error {
	return template.Validate()
}

// NewTemplateRegistryAdapter creates a new template registry adapter from a concrete registry
func NewTemplateRegistryAdapter(concreteRegistry *TemplateRegistry) *RegistryAdapter {
	return &RegistryAdapter{
		registry: &ConcreteRegistryWrapper{concreteRegistry: concreteRegistry},
	}
}

// ConcreteRegistryWrapper wraps the concrete TemplateRegistry to implement interfaces.TemplateRegistry
type ConcreteRegistryWrapper struct {
	concreteRegistry *TemplateRegistry
}

// Register registers a template
func (w *ConcreteRegistryWrapper) Register(template interfaces.Template) error {
	// Convert interfaces.Template to *format.Template
	formatTemplate, ok := template.(*format.Template)
	if !ok {
		return fmt.Errorf("template must be of type *format.Template")
	}
	return w.concreteRegistry.Register(formatTemplate)
}

// Unregister unregisters a template
func (w *ConcreteRegistryWrapper) Unregister(id string) error {
	return w.concreteRegistry.Unregister(id)
}

// Get gets a registered template
func (w *ConcreteRegistryWrapper) Get(id string) (interfaces.Template, bool) {
	template, found := w.concreteRegistry.Get(id)
	if !found {
		return nil, false
	}
	return template, true
}

// List lists all registered templates
func (w *ConcreteRegistryWrapper) List() []interfaces.Template {
	concreteList := w.concreteRegistry.List()
	interfaceList := make([]interfaces.Template, len(concreteList))
	for i, template := range concreteList {
		interfaceList[i] = template
	}
	return interfaceList
}

// Clear clears the registry
func (w *ConcreteRegistryWrapper) Clear() {
	w.concreteRegistry.Clear()
}

// Additional methods for TemplateRegistryExtended interface

// FindByTag finds templates by tag
func (a *RegistryAdapter) FindByTag(tag string) []*format.Template {
	// Use the concrete registry method
	if concreteWrapper, ok := a.registry.(*ConcreteRegistryWrapper); ok {
		return concreteWrapper.concreteRegistry.FindByTag(tag)
	}
	return []*format.Template{}
}

// FindByTags finds templates by multiple tags
func (a *RegistryAdapter) FindByTags(tags []string) []*format.Template {
	// Use the concrete registry method
	if concreteWrapper, ok := a.registry.(*ConcreteRegistryWrapper); ok {
		return concreteWrapper.concreteRegistry.FindByTags(tags)
	}
	return []*format.Template{}
}

// GetMetadata gets metadata for a template
func (a *RegistryAdapter) GetMetadata(id string) (map[string]interface{}, error) {
	// Use the concrete registry method
	if concreteWrapper, ok := a.registry.(*ConcreteRegistryWrapper); ok {
		return concreteWrapper.concreteRegistry.GetMetadata(id)
	}
	return nil, fmt.Errorf("metadata not supported by this registry implementation")
}

// SetMetadata sets metadata for a template
func (a *RegistryAdapter) SetMetadata(id string, metadata map[string]interface{}) error {
	// Use the concrete registry method
	if concreteWrapper, ok := a.registry.(*ConcreteRegistryWrapper); ok {
		return concreteWrapper.concreteRegistry.SetMetadata(id, metadata)
	}
	return fmt.Errorf("metadata not supported by this registry implementation")
}

// Count returns the number of templates
func (a *RegistryAdapter) Count() int {
	// Use the concrete registry method
	if concreteWrapper, ok := a.registry.(*ConcreteRegistryWrapper); ok {
		return concreteWrapper.concreteRegistry.Count()
	}
	return 0
}

// Update updates a template
func (a *RegistryAdapter) Update(template *format.Template) error {
	// Use the concrete registry method
	if concreteWrapper, ok := a.registry.(*ConcreteRegistryWrapper); ok {
		return concreteWrapper.concreteRegistry.Update(template)
	}
	return fmt.Errorf("update not supported by this registry implementation")
}
