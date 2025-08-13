// Package registry provides functionality for registering and managing templates.
package registry

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateRegistryAdapter adapts the TemplateRegistry to implement the interfaces.TemplateRegistry interface
type TemplateRegistryAdapter struct {
	registry *TemplateRegistry
}

// NewTemplateRegistryAdapter creates a new template registry adapter
func NewTemplateRegistryAdapter(registry *TemplateRegistry) *TemplateRegistryAdapter {
	return &TemplateRegistryAdapter{
		registry: registry,
	}
}

// Register registers a template
func (a *TemplateRegistryAdapter) Register(template *format.Template) error {
	return a.registry.Register(template)
}

// Unregister unregisters a template
func (a *TemplateRegistryAdapter) Unregister(id string) error {
	return a.registry.Unregister(id)
}

// Get gets a template
func (a *TemplateRegistryAdapter) Get(id string) (*format.Template, error) {
	return a.registry.Get(id)
}

// List lists all templates
func (a *TemplateRegistryAdapter) List() []*format.Template {
	return a.registry.List()
}

// FindByTag finds templates by tag
func (a *TemplateRegistryAdapter) FindByTag(tag string) []*format.Template {
	return a.registry.FindByTag(tag)
}

// FindByTags finds templates by tags
func (a *TemplateRegistryAdapter) FindByTags(tags []string) []*format.Template {
	return a.registry.FindByTags(tags)
}

// GetMetadata gets metadata for a template
func (a *TemplateRegistryAdapter) GetMetadata(id string) (map[string]interface{}, error) {
	return a.registry.GetMetadata(id)
}

// SetMetadata sets metadata for a template
func (a *TemplateRegistryAdapter) SetMetadata(id string, metadata map[string]interface{}) error {
	return a.registry.SetMetadata(id, metadata)
}

// Count returns the number of templates
func (a *TemplateRegistryAdapter) Count() int {
	return a.registry.Count()
}

// Update updates a template in the registry
func (a *TemplateRegistryAdapter) Update(template *format.Template) error {
	// First unregister the template
	err := a.registry.Unregister(template.ID)
	if err != nil {
		return fmt.Errorf("failed to unregister template before update: %w", err)
	}
	
	// Then register it again with the updated data
	return a.registry.Register(template)
}
