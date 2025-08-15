package manifest

import (
)

// TemplateEntry represents an entry in the template manifest
type TemplateEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Severity    string   `json:"severity"`
	Tags        []string `json:"tags,omitempty"`
	Path        string   `json:"path"`
	AddedAt     string   `json:"added_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

// CategoryInfo represents information about a template category
type CategoryInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Templates   []string `json:"templates"`
}

// TemplateManifest represents the manifest file for templates
type TemplateManifest struct {
	SchemaVersion string                  `json:"schema_version"`
	LastUpdated   string                  `json:"last_updated"`
	Templates     map[string]TemplateEntry `json:"templates"`
	Categories    map[string]CategoryInfo  `json:"categories"`

// NewTemplateManifest creates a new template manifest
func NewTemplateManifest() *TemplateManifest {
	return &TemplateManifest{
		SchemaVersion: "1.0",
		LastUpdated:   time.Now().Format(time.RFC3339),
		Templates:     make(map[string]TemplateEntry),
		Categories:    make(map[string]CategoryInfo),
	}

// ModuleEntry represents an entry in the module manifest
type ModuleEntry struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags,omitempty"`
	Path        string   `json:"path"`
	AddedAt     string   `json:"added_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`

// TypeInfo represents information about a module type
type TypeInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Modules     []string `json:"modules"`
}

// ModuleManifest represents the manifest file for modules
type ModuleManifest struct {
	SchemaVersion string                 `json:"schema_version"`
	LastUpdated   string                 `json:"last_updated"`
	Modules       map[string]ModuleEntry `json:"modules"`
	Types         map[string]TypeInfo    `json:"types"`

// NewModuleManifest creates a new module manifest
func NewModuleManifest() *ModuleManifest {
	return &ModuleManifest{
		SchemaVersion: "1.0",
		LastUpdated:   time.Now().Format(time.RFC3339),
		Modules:       make(map[string]ModuleEntry),
		Types:         make(map[string]TypeInfo),
	}
