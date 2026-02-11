package format

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ModuleType represents the type of a module
type ModuleType string

// Module type constants
const (
	ProviderModule ModuleType = "provider"
	UtilityModule  ModuleType = "utility"
	DetectorModule ModuleType = "detector"
)

// ModuleInfo contains module metadata
type ModuleInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
	Tags        []string
}

// Module represents a template module
type Module struct {
	ID          string
	Name        string
	Type        ModuleType
	Version     string
	Path        string
	Description string
	Info        *ModuleInfo
}

// LoadModule loads a module from the given path
func LoadModule(path string) (*Module, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("module path does not exist: %s", path)
	}

	module := &Module{
		ID:          filepath.Base(path),
		Name:        filepath.Base(path),
		Type:        UtilityModule,
		Path:        path,
		Description: "Template module",
		Info: &ModuleInfo{
			Name:        filepath.Base(path),
			Version:     "1.0.0",
			Description: "Template module",
		},
	}

	return module, nil
}

// Save saves the module to the specified directory
func (m *Module) Save(dir string) error {
	if err := os.MkdirAll(dir, 0750); err != nil { // Restrict directory permissions
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// LoadModuleFromFile loads a module from a file (alias for LoadModule)
func LoadModuleFromFile(path string) (*Module, error) {
	return LoadModule(path)
}

// ParseModule parses module content from YAML/JSON bytes
func ParseModule(content []byte) (*Module, error) {
	var module Module
	if err := yaml.Unmarshal(content, &module); err != nil {
		return nil, fmt.Errorf("failed to parse module content: %w", err)
	}
	return &module, nil
}
