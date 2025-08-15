package format

import (
	"fmt"
	"os"
	"path/filepath"
)

// Module represents a template module
type Module struct {
	Name        string
	Version     string
	Path        string
	Description string
}

// LoadModule loads a module from the given path
func LoadModule(path string) (*Module, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("module path does not exist: %s", path)
	}
	
	module := &Module{
		Name:        filepath.Base(path),
		Path:        path,
		Description: "Template module",
	}
	
	return module, nil
}

// Save saves the module to the specified directory
func (m *Module) Save(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	return nil
}
