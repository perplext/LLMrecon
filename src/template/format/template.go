package format

import (
	"fmt"
	"os"
	"path/filepath"
)

// Template represents a security template
type Template struct {
	Name        string
	Version     string
	Path        string
	Content     []byte
	Metadata    map[string]interface{}
}

// LoadTemplate loads a template from the given path
func LoadTemplate(path string) (*Template, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("template path does not exist: %s", path)
	}
	
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}
	
	template := &Template{
		Name:     filepath.Base(path),
		Path:     path,
		Content:  content,
		Metadata: make(map[string]interface{}),
	}
	
	return template, nil
}

// Save saves the template to the specified directory
func (t *Template) Save(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	targetPath := filepath.Join(dir, t.Name)
	if err := os.WriteFile(targetPath, t.Content, 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}
	
	return nil
}
