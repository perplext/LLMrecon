// Package customization provides functionality for identifying, preserving, and reapplying
// user customizations during template and module updates.
package customization

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// CustomizationType represents the type of customization
type CustomizationType string

const (
	// TemplateCustomization represents a customization to a template
	TemplateCustomization CustomizationType = "template"
	// ModuleCustomization represents a customization to a module
	ModuleCustomization CustomizationType = "module"
)

// CustomizationEntry represents a single customization entry in the registry
type CustomizationEntry struct {
	// ID is the unique identifier for the customization
	ID string `json:"id"`
	// Type is the type of customization
	Type CustomizationType `json:"type"`
	// Path is the path to the customized file, relative to the component root
	Path string `json:"path"`
	// ComponentID is the ID of the template or module
	ComponentID string `json:"component_id"`
	// BaseVersion is the version of the component that was customized
	BaseVersion string `json:"base_version"`
	// CustomizationDate is the date the customization was made
	CustomizationDate time.Time `json:"customization_date"`
	// Hash is the hash of the original file content
	OriginalHash string `json:"original_hash"`
	// CustomizedHash is the hash of the customized file content
	CustomizedHash string `json:"customized_hash"`
	// Markers contains the customization markers found in the file
	Markers []CustomizationMarker `json:"markers,omitempty"`
	// Policy is the preservation policy for this customization
	Policy PreservationPolicy `json:"policy"`
}

// CustomizationMarker represents a marker for customization in a file
type CustomizationMarker struct {
	// Type is the type of marker
	Type string `json:"type"`
	// StartLine is the line number where the marker starts
	StartLine int `json:"start_line"`
	// EndLine is the line number where the marker ends
	EndLine int `json:"end_line"`
	// Content is the content of the marker
	Content string `json:"content,omitempty"`
}

// PreservationPolicy represents the policy for preserving customizations
type PreservationPolicy string

const (
	// AlwaysPreserve means the customization should always be preserved
	AlwaysPreserve PreservationPolicy = "always_preserve"
	// PreserveWithConflictResolution means the customization should be preserved with conflict resolution
	PreserveWithConflictResolution PreservationPolicy = "preserve_with_conflict_resolution"
	// AskUser means the user should be asked what to do with the customization
	AskUser PreservationPolicy = "ask_user"
	// Discard means the customization should be discarded
	Discard PreservationPolicy = "discard"
)

// Registry represents the customization registry
type Registry struct {
	// Entries is a map of customization entries by ID
	Entries map[string]*CustomizationEntry `json:"entries"`
	// RegistryPath is the path to the registry file
	RegistryPath string `json:"-"`
}

// NewRegistry creates a new customization registry
func NewRegistry(registryPath string) (*Registry, error) {
	registry := &Registry{
		Entries:      make(map[string]*CustomizationEntry),
		RegistryPath: registryPath,
	}

	// Create registry directory if it doesn't exist
	registryDir := filepath.Dir(registryPath)
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Load registry if it exists
	if _, err := os.Stat(registryPath); err == nil {
		if err := registry.Load(); err != nil {
			return nil, fmt.Errorf("failed to load registry: %w", err)
		}
	}

	return registry, nil
}

// Load loads the registry from disk
func (r *Registry) Load() error {
	data, err := ioutil.ReadFile(r.RegistryPath)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	if err := json.Unmarshal(data, &r.Entries); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}

	return nil
}

// Save saves the registry to disk
func (r *Registry) Save() error {
	data, err := json.MarshalIndent(r.Entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}

	if err := ioutil.WriteFile(r.RegistryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	return nil
}

// AddEntry adds a customization entry to the registry
func (r *Registry) AddEntry(entry *CustomizationEntry) {
	r.Entries[entry.ID] = entry
}

// RemoveEntry removes a customization entry from the registry
func (r *Registry) RemoveEntry(id string) {
	delete(r.Entries, id)
}

// GetEntry gets a customization entry from the registry
func (r *Registry) GetEntry(id string) *CustomizationEntry {
	return r.Entries[id]
}

// GetEntriesByComponentID gets all customization entries for a component
func (r *Registry) GetEntriesByComponentID(componentID string, entryType CustomizationType) []*CustomizationEntry {
	var entries []*CustomizationEntry
	for _, entry := range r.Entries {
		if entry.ComponentID == componentID && entry.Type == entryType {
			entries = append(entries, entry)
		}
	}
	return entries
}

// GetEntriesByPath gets all customization entries for a path
func (r *Registry) GetEntriesByPath(path string, entryType CustomizationType) []*CustomizationEntry {
	var entries []*CustomizationEntry
	for _, entry := range r.Entries {
		if entry.Path == path && entry.Type == entryType {
			entries = append(entries, entry)
		}
	}
	return entries
}
