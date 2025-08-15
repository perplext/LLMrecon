// Package registry provides functionality for registering and managing templates.
package registry

import (
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateRegistry is responsible for registering and managing templates
type TemplateRegistry struct {
	// templates is a map of template ID to template
	templates map[string]*format.Template
	// templatesMutex is a mutex for the templates map
	templatesMutex sync.RWMutex
	// metadata is a map of template ID to template metadata
	metadata map[string]*TemplateMetadata
	// metadataMutex is a mutex for the metadata map
	metadataMutex sync.RWMutex

// TemplateMetadata contains metadata about a template
type TemplateMetadata struct {
	// RegisteredAt is the time the template was registered
	RegisteredAt time.Time
	// LastUsedAt is the time the template was last used
	LastUsedAt time.Time
	// UsageCount is the number of times the template has been used
	UsageCount int
	// Tags is a list of tags for the template
	Tags []string
	// Source is the source of the template
	Source string
	// Path is the path to the template file
	Path string

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*format.Template),
		metadata:  make(map[string]*TemplateMetadata),
	}

// Register registers a template
func (r *TemplateRegistry) Register(template *format.Template) error {
	// Check if template is valid
	if template == nil {
		return fmt.Errorf("template is nil")
	}
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	// Check if template already exists
	r.templatesMutex.RLock()
	_, exists := r.templates[template.ID]
	r.templatesMutex.RUnlock()

	if exists {
		return fmt.Errorf("template with ID %s already exists", template.ID)
	}

	// Register template
	r.templatesMutex.Lock()
	r.templates[template.ID] = template
	r.templatesMutex.Unlock()

	// Create metadata
	r.metadataMutex.Lock()
	r.metadata[template.ID] = &TemplateMetadata{
		RegisteredAt: time.Now(),
		LastUsedAt:   time.Now(),
		UsageCount:   0,
		Tags:         template.Info.Tags,
	}
	r.metadataMutex.Unlock()

	return nil

// Unregister unregisters a template
func (r *TemplateRegistry) Unregister(id string) error {
	// Check if template exists
	r.templatesMutex.RLock()
	_, exists := r.templates[id]
	r.templatesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("template with ID %s does not exist", id)
	}

	// Unregister template
	r.templatesMutex.Lock()
	delete(r.templates, id)
	r.templatesMutex.Unlock()

	// Delete metadata
	r.metadataMutex.Lock()
	delete(r.metadata, id)
	r.metadataMutex.Unlock()

	return nil

// Update updates a template
func (r *TemplateRegistry) Update(template *format.Template) error {
	// Check if template is valid
	if template == nil {
		return fmt.Errorf("template is nil")
	}
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	// Check if template exists
	r.templatesMutex.RLock()
	_, exists := r.templates[template.ID]
	r.templatesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("template with ID %s does not exist", template.ID)
	}

	// Update template
	r.templatesMutex.Lock()
	r.templates[template.ID] = template
	r.templatesMutex.Unlock()

	// Update metadata
	r.metadataMutex.Lock()
	if metadata, ok := r.metadata[template.ID]; ok {
		metadata.Tags = template.Info.Tags
	}
	r.metadataMutex.Unlock()

	return nil

// Get gets a template by ID
func (r *TemplateRegistry) Get(id string) (*format.Template, error) {
	// Get template
	r.templatesMutex.RLock()
	template, exists := r.templates[id]
	r.templatesMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("template with ID %s does not exist", id)
	}

	// Update metadata
	r.metadataMutex.Lock()
	if metadata, ok := r.metadata[id]; ok {
		metadata.LastUsedAt = time.Now()
		metadata.UsageCount++
	}
	r.metadataMutex.Unlock()

	return template, nil

// List lists all templates
func (r *TemplateRegistry) List() []*format.Template {
	r.templatesMutex.RLock()
	defer r.templatesMutex.RUnlock()

	templates := make([]*format.Template, 0, len(r.templates))
	for _, template := range r.templates {
		templates = append(templates, template)
	}

	return templates

// GetMetadata gets metadata for a template
func (r *TemplateRegistry) GetMetadata(id string) (map[string]interface{}, error) {
	r.metadataMutex.RLock()
	metadata, exists := r.metadata[id]
	r.metadataMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("metadata for template with ID %s does not exist", id)
	}

	// Convert TemplateMetadata to map[string]interface{}
	result := make(map[string]interface{})
	result["registeredAt"] = metadata.RegisteredAt
	result["lastUsedAt"] = metadata.LastUsedAt
	result["usageCount"] = metadata.UsageCount
	result["tags"] = metadata.Tags
	result["source"] = metadata.Source
	result["path"] = metadata.Path

	return result, nil

// SetMetadata sets metadata for a template
func (r *TemplateRegistry) SetMetadata(id string, metadata map[string]interface{}) error {
	// Check if template exists
	r.templatesMutex.RLock()
	_, exists := r.templates[id]
	r.templatesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("template with ID %s does not exist", id)
	}

	// Convert map[string]interface{} to TemplateMetadata
	templateMetadata := &TemplateMetadata{}
	
	// Set fields from map if they exist
	if val, ok := metadata["registeredAt"]; ok {
		if t, ok := val.(time.Time); ok {
			templateMetadata.RegisteredAt = t
		}
	} else {
		templateMetadata.RegisteredAt = time.Now()
	}
	
	if val, ok := metadata["lastUsedAt"]; ok {
		if t, ok := val.(time.Time); ok {
			templateMetadata.LastUsedAt = t
		}
	}
	
	if val, ok := metadata["usageCount"]; ok {
		if count, ok := val.(int); ok {
			templateMetadata.UsageCount = count
		}
	}
	
	if val, ok := metadata["tags"]; ok {
		if tags, ok := val.([]string); ok {
			templateMetadata.Tags = tags
		}
	}
	
	if val, ok := metadata["source"]; ok {
		if source, ok := val.(string); ok {
			templateMetadata.Source = source
		}
	}
	
	if val, ok := metadata["path"]; ok {
		if path, ok := val.(string); ok {
			templateMetadata.Path = path
		}
	}

	// Set metadata
	r.metadataMutex.Lock()
	r.metadata[id] = templateMetadata
	r.metadataMutex.Unlock()

	return nil

// FindByTag finds templates by tag
func (r *TemplateRegistry) FindByTag(tag string) []*format.Template {
	r.templatesMutex.RLock()
	defer r.templatesMutex.RUnlock()

	var templates []*format.Template
	for id, template := range r.templates {
		r.metadataMutex.RLock()
		metadata, ok := r.metadata[id]
		r.metadataMutex.RUnlock()

		if ok {
			for _, t := range metadata.Tags {
				if t == tag {
					templates = append(templates, template)
					break
				}
			}
		}
	}

	return templates

// FindByTags finds templates by multiple tags (AND logic)
func (r *TemplateRegistry) FindByTags(tags []string) []*format.Template {
	r.templatesMutex.RLock()
	defer r.templatesMutex.RUnlock()

	var templates []*format.Template
	for id, template := range r.templates {
		r.metadataMutex.RLock()
		metadata, ok := r.metadata[id]
		r.metadataMutex.RUnlock()

		if ok {
			// Check if template has all tags
			hasAllTags := true
			for _, tag := range tags {
				found := false
				for _, t := range metadata.Tags {
					if t == tag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}

			if hasAllTags {
				templates = append(templates, template)
			}
		}
	}

	return templates

// FindBySource finds templates by source
func (r *TemplateRegistry) FindBySource(source string) []*format.Template {
	r.templatesMutex.RLock()
	defer r.templatesMutex.RUnlock()

	var templates []*format.Template
	for id, template := range r.templates {
		r.metadataMutex.RLock()
		metadata, ok := r.metadata[id]
		r.metadataMutex.RUnlock()

		if ok && metadata.Source == source {
			templates = append(templates, template)
		}
	}

	return templates

// Count returns the number of registered templates
func (r *TemplateRegistry) Count() int {
	r.templatesMutex.RLock()
	defer r.templatesMutex.RUnlock()
	return len(r.templates)

// Clear clears all templates
func (r *TemplateRegistry) Clear() {
	r.templatesMutex.Lock()
	r.templates = make(map[string]*format.Template)
	r.templatesMutex.Unlock()

	r.metadataMutex.Lock()
	r.metadata = make(map[string]*TemplateMetadata)
	r.metadataMutex.Unlock()

// SetSource sets the source of a template
func (r *TemplateRegistry) SetSource(id string, source string) error {
	// Check if template exists
	r.templatesMutex.RLock()
	_, exists := r.templates[id]
	r.templatesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("template with ID %s does not exist", id)
	}

	// Set source
	r.metadataMutex.Lock()
	if metadata, ok := r.metadata[id]; ok {
		metadata.Source = source
	} else {
		r.metadata[id] = &TemplateMetadata{
			RegisteredAt: time.Now(),
			LastUsedAt:   time.Now(),
			UsageCount:   0,
			Source:       source,
		}
	}
	r.metadataMutex.Unlock()

	return nil

// SetPath sets the path of a template
func (r *TemplateRegistry) SetPath(id string, path string) error {
	// Check if template exists
	r.templatesMutex.RLock()
	_, exists := r.templates[id]
	r.templatesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("template with ID %s does not exist", id)
	}

	// Set path
	r.metadataMutex.Lock()
	if metadata, ok := r.metadata[id]; ok {
		metadata.Path = path
	} else {
		r.metadata[id] = &TemplateMetadata{
			RegisteredAt: time.Now(),
			LastUsedAt:   time.Now(),
			UsageCount:   0,
			Path:         path,
		}
	}
	r.metadataMutex.Unlock()

	return nil

// GetPath gets the path of a template
func (r *TemplateRegistry) GetPath(id string) (string, error) {
	r.metadataMutex.RLock()
	metadata, exists := r.metadata[id]
	r.metadataMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("metadata for template with ID %s does not exist", id)
	}

	return metadata.Path, nil

// GetSource gets the source of a template
func (r *TemplateRegistry) GetSource(id string) (string, error) {
	r.metadataMutex.RLock()
	metadata, exists := r.metadata[id]
	r.metadataMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("metadata for template with ID %s does not exist", id)
	}

