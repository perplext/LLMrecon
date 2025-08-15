// Package customization provides customization registry
package customization

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Registry manages customization registration
type Registry struct {
	FilePath        string
	Customizations  []Customization
	mutex           sync.RWMutex
}

// NewRegistry creates a new customization registry
func NewRegistry(filePath string) *Registry {
	return &Registry{
		FilePath:       filePath,
		Customizations: make([]Customization, 0),
	}
}

// Load loads the registry from disk
func (r *Registry) Load() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	data, err := os.ReadFile(filepath.Clean(r.FilePath))
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's OK
			return nil
		}
		return fmt.Errorf("failed to read registry file: %w", err)
	}
	
	if err := json.Unmarshal(data, &r.Customizations); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}
	
	return nil
}

// Save saves the registry to disk
func (r *Registry) Save() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	data, err := json.MarshalIndent(r.Customizations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}
	
	// Create directory if needed
	dir := filepath.Dir(r.FilePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}
	
	if err := os.WriteFile(filepath.Clean(r.FilePath), data, 0640); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}
	
	return nil
}

// Register registers a customization
func (r *Registry) Register(custom Customization) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Check if already registered
	for i, existing := range r.Customizations {
		if existing.Path == custom.Path {
			// Update existing registration
			r.Customizations[i] = custom
			return nil
		}
	}
	
	// Add new registration
	r.Customizations = append(r.Customizations, custom)
	return nil
}

// GetCustomizations returns all registered customizations
func (r *Registry) GetCustomizations() []Customization {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make([]Customization, len(r.Customizations))
	copy(result, r.Customizations)
	return result
}

// FindByPath finds a customization by path
func (r *Registry) FindByPath(path string) (*Customization, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	for _, custom := range r.Customizations {
		if custom.Path == path {
			c := custom
			return &c, true
		}
	}
	
	return nil, false
}
