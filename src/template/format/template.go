package format

import (
	"fmt"
	"os"
	"path/filepath"
)

// Template represents a security template
type Template struct {
	ID            string
	Name          string
	Version       string
	Path          string
	Content       []byte
	Metadata      map[string]interface{}
	Variables     map[string]interface{}
	Info          *TemplateInfo
	Test          *TemplateTest
	Compatibility *TemplateCompatibility
}

// TemplateCompatibility contains template compatibility information
type TemplateCompatibility struct {
	MinToolVersion string
	MaxToolVersion string
}

// TemplateInfo contains template information
type TemplateInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
	Severity    string
	Tags        []string
}

// TemplateDetection contains detection criteria for templates
type TemplateDetection struct {
	Type     string `json:"type"`
	Pattern  string `json:"pattern,omitempty"`
	Match    string `json:"match,omitempty"`
	Criteria string `json:"criteria,omitempty"`
}

// TemplateVariation contains a variation of the template test
type TemplateVariation struct {
	Prompt    string              `json:"prompt"`
	Expected  string              `json:"expected,omitempty"`
	Detection *TemplateDetection  `json:"detection,omitempty"`
}

// TemplateTest contains template test information
type TemplateTest struct {
	Prompt     string                `json:"prompt"`
	Expected   string                `json:"expected,omitempty"`
	Detection  *TemplateDetection    `json:"detection,omitempty"`
	Variations []*TemplateVariation  `json:"variations,omitempty"`
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
		ID:        filepath.Base(path),
		Name:      filepath.Base(path),
		Path:      path,
		Content:   content,
		Metadata:  make(map[string]interface{}),
		Variables: make(map[string]interface{}),
		Info: &TemplateInfo{
			Name:     filepath.Base(path),
			Version:  "1.0.0",
			Severity: "medium",
		},
		Test: &TemplateTest{
			Prompt:     "",
			Expected:   "",
			Detection:  &TemplateDetection{},
			Variations: make([]*TemplateVariation, 0),
		},
	}

	return template, nil
}

// Save saves the template to the specified directory
func (t *Template) Save(dir string) error {
	if err := os.MkdirAll(dir, 0750); err != nil { // Restrict directory permissions
		return fmt.Errorf("failed to create directory: %w", err)
	}

	targetPath := filepath.Join(dir, t.Name)
	if err := os.WriteFile(targetPath, t.Content, 0600); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}

// ValidateStructure validates the template structure
func (t *Template) ValidateStructure() error {
	if t.ID == "" {
		return fmt.Errorf("template ID is required")
	}
	if t.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if t.Info == nil {
		return fmt.Errorf("template info is required")
	}
	if t.Test == nil {
		return fmt.Errorf("template test is required")
	}
	return nil
}

// GetID returns the template ID
func (t *Template) GetID() string {
	return t.ID
}

// GetName returns the template name
func (t *Template) GetName() string {
	return t.Name
}

// GetVersion returns the template version
func (t *Template) GetVersion() string {
	return t.Version
}

// GetContent returns the template content
func (t *Template) GetContent() ([]byte, error) {
	if len(t.Content) == 0 {
		return nil, fmt.Errorf("template content is empty")
	}
	return t.Content, nil
}

// GetCategory returns the template category (derived from tags or metadata)
func (t *Template) GetCategory() string {
	if t.Info != nil && len(t.Info.Tags) > 0 {
		return t.Info.Tags[0] // Use first tag as category
	}
	if category, exists := t.Metadata["category"].(string); exists {
		return category
	}
	return "general"
}

// GetSeverity returns the template severity
func (t *Template) GetSeverity() string {
	if t.Info != nil && t.Info.Severity != "" {
		return t.Info.Severity
	}
	return "medium"
}

// GetReferences returns template references
func (t *Template) GetReferences() []string {
	if refs, exists := t.Metadata["references"].([]string); exists {
		return refs
	}
	if refs, exists := t.Metadata["references"].([]interface{}); exists {
		stringRefs := make([]string, len(refs))
		for i, ref := range refs {
			if str, ok := ref.(string); ok {
				stringRefs[i] = str
			}
		}
		return stringRefs
	}
	return []string{}
}

// GetDescription returns the template description
func (t *Template) GetDescription() string {
	if t.Info != nil && t.Info.Description != "" {
		return t.Info.Description
	}
	if desc, exists := t.Metadata["description"].(string); exists {
		return desc
	}
	return ""
}

// GetAuthor returns the template author
func (t *Template) GetAuthor() string {
	if t.Info != nil && t.Info.Author != "" {
		return t.Info.Author
	}
	if author, exists := t.Metadata["author"].(string); exists {
		return author
	}
	return ""
}

// GetTags returns the template tags
func (t *Template) GetTags() []string {
	if t.Info != nil && t.Info.Tags != nil {
		return t.Info.Tags
	}
	return []string{}
}

// Validate validates the template
func (t *Template) Validate() error {
	return t.ValidateStructure()
}

// Clone creates a deep copy of the template
func (t *Template) Clone() *Template {
	if t == nil {
		return nil
	}

	// Create a new template
	clone := &Template{
		ID:      t.ID,
		Name:    t.Name,
		Version: t.Version,
		Path:    t.Path,
	}

	// Copy content
	if t.Content != nil {
		clone.Content = make([]byte, len(t.Content))
		copy(clone.Content, t.Content)
	}

	// Copy metadata
	if t.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range t.Metadata {
			clone.Metadata[k] = v
		}
	}

	// Copy variables
	if t.Variables != nil {
		clone.Variables = make(map[string]interface{})
		for k, v := range t.Variables {
			clone.Variables[k] = v
		}
	}

	// Copy info
	if t.Info != nil {
		clone.Info = &TemplateInfo{
			Name:        t.Info.Name,
			Version:     t.Info.Version,
			Description: t.Info.Description,
			Author:      t.Info.Author,
			Severity:    t.Info.Severity,
		}
		if t.Info.Tags != nil {
			clone.Info.Tags = make([]string, len(t.Info.Tags))
			copy(clone.Info.Tags, t.Info.Tags)
		}
	}

	// Copy test
	if t.Test != nil {
		clone.Test = &TemplateTest{
			Prompt:   t.Test.Prompt,
			Expected: t.Test.Expected,
		}
		
		// Copy detection
		if t.Test.Detection != nil {
			clone.Test.Detection = &TemplateDetection{
				Type:     t.Test.Detection.Type,
				Pattern:  t.Test.Detection.Pattern,
				Match:    t.Test.Detection.Match,
				Criteria: t.Test.Detection.Criteria,
			}
		}
		
		// Copy variations
		if t.Test.Variations != nil {
			clone.Test.Variations = make([]*TemplateVariation, len(t.Test.Variations))
			for i, variation := range t.Test.Variations {
				cloneVariation := &TemplateVariation{
					Prompt:   variation.Prompt,
					Expected: variation.Expected,
				}
				
				// Copy variation detection
				if variation.Detection != nil {
					cloneVariation.Detection = &TemplateDetection{
						Type:     variation.Detection.Type,
						Pattern:  variation.Detection.Pattern,
						Match:    variation.Detection.Match,
						Criteria: variation.Detection.Criteria,
					}
				}
				
				clone.Test.Variations[i] = cloneVariation
			}
		}
	}

	// Copy compatibility
	if t.Compatibility != nil {
		clone.Compatibility = &TemplateCompatibility{
			MinToolVersion: t.Compatibility.MinToolVersion,
			MaxToolVersion: t.Compatibility.MaxToolVersion,
		}
	}

	return clone
}

// LoadFromFile loads a template from a file (alias for LoadTemplate)
func LoadFromFile(path string) (*Template, error) {
	return LoadTemplate(path)
}

// ParseTemplate parses a template from the given content bytes
func ParseTemplate(content []byte) (*Template, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("template content is empty")
	}

	// Create a basic template structure
	template := &Template{
		ID:        "parsed-template",
		Name:      "parsed-template",
		Content:   content,
		Metadata:  make(map[string]interface{}),
		Variables: make(map[string]interface{}),
		Info: &TemplateInfo{
			Name:     "parsed-template",
			Version:  "1.0.0",
			Severity: "medium",
		},
		Test: &TemplateTest{
			Prompt:     "",
			Expected:   "",
			Detection:  &TemplateDetection{},
			Variations: make([]*TemplateVariation, 0),
		},
	}

	// TODO: Add actual parsing logic for YAML/JSON content
	// For now, this creates a basic template structure

	return template, nil
}
