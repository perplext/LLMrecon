package format

import (
	"fmt"
	"io/ioutil"
	
	"github.com/perplext/LLMrecon/src/template/compatibility"
	"gopkg.in/yaml.v3"
)

// TemplateInfo represents the basic information section of a template
type TemplateInfo struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Version     string   `yaml:"version" json:"version"`
	Author      string   `yaml:"author" json:"author"`
	Severity    string   `yaml:"severity" json:"severity"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	References  []string `yaml:"references,omitempty" json:"references,omitempty"`
	Compliance  struct {
		OWASP string `yaml:"owasp,omitempty" json:"owasp,omitempty"`
		ISO   string `yaml:"iso,omitempty" json:"iso,omitempty"`
	} `yaml:"compliance,omitempty" json:"compliance,omitempty"`
}

// DetectionCriteria represents the detection criteria for a test
type DetectionCriteria struct {
	Type      string `yaml:"type" json:"type"`
	Match     string `yaml:"match,omitempty" json:"match,omitempty"`
	Pattern   string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Criteria  string `yaml:"criteria,omitempty" json:"criteria,omitempty"`
	Condition string `yaml:"condition,omitempty" json:"condition,omitempty"`
}

// TestVariation represents a variation of a test
type TestVariation struct {
	Prompt    string           `yaml:"prompt" json:"prompt"`
	Detection DetectionCriteria `yaml:"detection" json:"detection"`
}

// TestDefinition represents the test section of a template
type TestDefinition struct {
	Prompt           string           `yaml:"prompt" json:"prompt"`
	ExpectedBehavior string           `yaml:"expected_behavior,omitempty" json:"expected_behavior,omitempty"`
	Detection        DetectionCriteria `yaml:"detection" json:"detection"`
	Variations       []TestVariation   `yaml:"variations,omitempty" json:"variations,omitempty"`
}

// Template represents a vulnerability test template
type Template struct {
	ID           string                      `yaml:"id" json:"id"`
	Info         TemplateInfo                `yaml:"info" json:"info"`
	Compatibility *compatibility.CompatibilityMetadata `yaml:"compatibility" json:"compatibility"`
	Test         TestDefinition              `yaml:"test" json:"test"`
	Content      []byte                      `json:"-"` // Raw content for caching
	Variables    map[string]interface{}      `yaml:"variables,omitempty" json:"variables,omitempty"`
	Parent       string                      `yaml:"parent,omitempty" json:"parent,omitempty"`
	Metadata     map[string]interface{}      `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Version returns the version from the template info
func (t *Template) Version() string {
	return t.Info.Version
}

// Template interface implementation

// GetID returns the template ID
func (t *Template) GetID() string {
	return t.ID
}

// GetName returns the template name
func (t *Template) GetName() string {
	return t.Info.Name
}

// GetDescription returns the template description
func (t *Template) GetDescription() string {
	return t.Info.Description
}

// GetCategory returns the template category
func (t *Template) GetCategory() string {
	if t.Metadata != nil {
		if category, ok := t.Metadata["category"].(string); ok {
			return category
		}
	}
	// Fallback to tags or severity
	if len(t.Info.Tags) > 0 {
		return t.Info.Tags[0]
	}
	return t.Info.Severity
}

// GetSeverity returns the template severity
func (t *Template) GetSeverity() string {
	return t.Info.Severity
}

// GetAuthor returns the template author
func (t *Template) GetAuthor() string {
	return t.Info.Author
}

// GetVersion returns the template version
func (t *Template) GetVersion() string {
	return t.Info.Version
}

// GetTags returns the template tags
func (t *Template) GetTags() []string {
	return t.Info.Tags
}

// GetReferences returns the template references
func (t *Template) GetReferences() []string {
	return t.Info.References
}

// GetMetadata returns the template metadata
func (t *Template) GetMetadata() map[string]interface{} {
	return t.Metadata
}

// ValidateStructure validates the template structure and returns validation issues
func (t *Template) ValidateStructure() []string {
	// This will be implemented by renaming the existing Validate method
	return t.validateInternal()
}

// Validate validates the template and returns an error if invalid (implements Template interface)
func (t *Template) Validate() error {
	issues := t.validateInternal()
	if len(issues) > 0 {
		return fmt.Errorf("template validation failed: %v", issues)
	}
	return nil
}

// Clone creates a deep copy of the template
func (t *Template) Clone() *Template {
	if t == nil {
		return nil
	}
	
	// Create a new template
	clone := &Template{
		ID:     t.ID,
		Info:   t.Info,
		Test:   t.Test,
		Parent: t.Parent,
	}
	
	// Clone compatibility metadata
	if t.Compatibility != nil {
		clone.Compatibility = &compatibility.CompatibilityMetadata{
			Providers:        append([]string{}, t.Compatibility.Providers...),
			MinToolVersion:   t.Compatibility.MinToolVersion,
			MaxToolVersion:   t.Compatibility.MaxToolVersion,
			SupportedModels:  make(map[string][]string),
			RequiredFeatures: append([]string{}, t.Compatibility.RequiredFeatures...),
			Metadata:         make(map[string]interface{}),
		}
		// Clone supported models
		for provider, models := range t.Compatibility.SupportedModels {
			clone.Compatibility.SupportedModels[provider] = append([]string{}, models...)
		}
		// Clone metadata
		for k, v := range t.Compatibility.Metadata {
			clone.Compatibility.Metadata[k] = v
		}
	}
	
	// Clone content
	if t.Content != nil {
		clone.Content = make([]byte, len(t.Content))
		copy(clone.Content, t.Content)
	}
	
	// Clone variables
	if t.Variables != nil {
		clone.Variables = make(map[string]interface{})
		for k, v := range t.Variables {
			clone.Variables[k] = v
		}
	}
	
	return clone
}

// ValidSeverityLevels defines the valid severity levels for templates
var ValidSeverityLevels = []string{"info", "low", "medium", "high", "critical"}

// ValidDetectionTypes defines the valid detection types for templates
var ValidDetectionTypes = []string{"string_match", "regex_match", "semantic_match"}

// ValidConditions defines the valid conditions for detection criteria
var ValidConditions = []string{"contains", "not_contains"}

// NewTemplate creates a new template with default values
func NewTemplate() *Template {
	return &Template{
		Compatibility: compatibility.NewCompatibilityMetadata(),
	}
}

// LoadFromFile loads a template from a YAML file
func LoadFromFile(filePath string) (*Template, error) {
	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}
	
	// Parse YAML
	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template file: %w", err)
	}
	
	return &template, nil
}

// SaveToFile saves a template to a YAML file
func (t *Template) SaveToFile(filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}
	
	return nil
}

// ParseTemplate parses a template from bytes
func ParseTemplate(content []byte) (*Template, error) {
	// Parse YAML
	var template Template
	if err := yaml.Unmarshal(content, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	
	return &template, nil
}

// validateInternal validates the template structure and content
func (t *Template) validateInternal() []string {
	var issues []string
	
	// Validate ID
	if t.ID == "" {
		issues = append(issues, "Template ID is required")
	}
	
	// Validate Info section
	if t.Info.Name == "" {
		issues = append(issues, "Template name is required")
	}
	
	if t.Info.Description == "" {
		issues = append(issues, "Template description is required")
	}
	
	if t.Info.Version == "" {
		issues = append(issues, "Template version is required")
	}
	
	if t.Info.Author == "" {
		issues = append(issues, "Template author is required")
	}
	
	if t.Info.Severity == "" {
		issues = append(issues, "Template severity is required")
	} else {
		var validSeverity bool
		for _, level := range ValidSeverityLevels {
			if t.Info.Severity == level {
				validSeverity = true
				break
			}
		}
		
		if !validSeverity {
			issues = append(issues, fmt.Sprintf("Invalid severity level: %s (valid levels: %v)",
				t.Info.Severity, ValidSeverityLevels))
		}
	}
	
	// Validate Compatibility section
	if t.Compatibility == nil {
		issues = append(issues, "Template compatibility section is required")
	} else if len(t.Compatibility.Providers) == 0 {
		issues = append(issues, "At least one compatible provider is required")
	}
	
	// Validate Test section
	if t.Test.Prompt == "" {
		issues = append(issues, "Test prompt is required")
	}
	
	if t.Test.Detection.Type == "" {
		issues = append(issues, "Detection type is required")
	} else {
		var validType bool
		for _, detectionType := range ValidDetectionTypes {
			if t.Test.Detection.Type == detectionType {
				validType = true
				break
			}
		}
		
		if !validType {
			issues = append(issues, fmt.Sprintf("Invalid detection type: %s (valid types: %v)",
				t.Test.Detection.Type, ValidDetectionTypes))
		}
	}
	
	// Validate detection criteria based on type
	issues = append(issues, validateDetectionCriteria(t.Test.Detection)...)
	
	// Validate test variations
	for i, variation := range t.Test.Variations {
		if variation.Prompt == "" {
			issues = append(issues, fmt.Sprintf("Prompt is required for test variation %d", i+1))
		}
		
		issues = append(issues, validateDetectionCriteria(variation.Detection)...)
	}
	
	return issues
}

// validateDetectionCriteria validates the detection criteria based on its type
func validateDetectionCriteria(criteria DetectionCriteria) []string {
	var issues []string
	
	switch criteria.Type {
	case "string_match":
		if criteria.Match == "" {
			issues = append(issues, "Match string is required for string_match detection")
		}
	case "regex_match":
		if criteria.Pattern == "" {
			issues = append(issues, "Pattern is required for regex_match detection")
		}
	case "semantic_match":
		if criteria.Criteria == "" {
			issues = append(issues, "Criteria is required for semantic_match detection")
		}
	}
	
	if criteria.Condition != "" {
		var validCondition bool
		for _, condition := range ValidConditions {
			if criteria.Condition == condition {
				validCondition = true
				break
			}
		}
		
		if !validCondition {
			issues = append(issues, fmt.Sprintf("Invalid condition: %s (valid conditions: %v)",
				criteria.Condition, ValidConditions))
		}
	}
	
	return issues
}

// ensureDir ensures that a directory exists
func ensureDir(dir string) error {
	return nil // Placeholder for actual implementation
}
