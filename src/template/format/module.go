package format

import (
	"fmt"
	"io/ioutil"
	
	"github.com/perplext/LLMrecon/src/template/compatibility"
	"gopkg.in/yaml.v3"
)

// ModuleType represents the type of module
type ModuleType string

// Available module types
const (
	ProviderModule  ModuleType = "provider"
	UtilityModule   ModuleType = "utility"
	DetectorModule  ModuleType = "detector"
)

// ModuleInfo represents the basic information section of a module
type ModuleInfo struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Version     string   `yaml:"version" json:"version"`
	Author      string   `yaml:"author" json:"author"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	References  []string `yaml:"references,omitempty" json:"references,omitempty"`
}

// ProviderConfig represents the configuration for a provider module
type ProviderConfig struct {
	SupportedModels []string               `yaml:"supported_models" json:"supported_models"`
	Features        []string               `yaml:"features" json:"features"`
	DefaultOptions  map[string]interface{} `yaml:"default_options,omitempty" json:"default_options,omitempty"`
	RateLimits      struct {
		RequestsPerMinute int `yaml:"requests_per_minute,omitempty" json:"requests_per_minute,omitempty"`
		TokensPerMinute   int `yaml:"tokens_per_minute,omitempty" json:"tokens_per_minute,omitempty"`
	} `yaml:"rate_limits,omitempty" json:"rate_limits,omitempty"`
}

// UtilityConfig represents the configuration for a utility module
type UtilityConfig struct {
	Functions     []string               `yaml:"functions" json:"functions"`
	DefaultOptions map[string]interface{} `yaml:"default_options,omitempty" json:"default_options,omitempty"`
}

// DetectorConfig represents the configuration for a detector module
type DetectorConfig struct {
	DetectionType string                 `yaml:"detection_type" json:"detection_type"`
	DefaultOptions map[string]interface{} `yaml:"default_options,omitempty" json:"default_options,omitempty"`
}

// Module represents a module definition (provider, utility, or detector)
type Module struct {
	ID           string                      `yaml:"id" json:"id"`
	Type         ModuleType                  `yaml:"type" json:"type"`
	Info         ModuleInfo                  `yaml:"info" json:"info"`
	Compatibility *compatibility.CompatibilityMetadata `yaml:"compatibility" json:"compatibility"`
	Provider     *ProviderConfig             `yaml:"provider,omitempty" json:"provider,omitempty"`
	Utility      *UtilityConfig              `yaml:"utility,omitempty" json:"utility,omitempty"`
	Detector     *DetectorConfig             `yaml:"detector,omitempty" json:"detector,omitempty"`
}

// NewModule creates a new module with default values
func NewModule(moduleType ModuleType) *Module {
	module := &Module{
		Type:         moduleType,
		Compatibility: compatibility.NewCompatibilityMetadata(),
	}
	
	// Initialize the appropriate config section based on module type
	switch moduleType {
	case ProviderModule:
		module.Provider = &ProviderConfig{
			SupportedModels: []string{},
			Features:        []string{},
			DefaultOptions:  make(map[string]interface{}),
		}
	case UtilityModule:
		module.Utility = &UtilityConfig{
			Functions:     []string{},
			DefaultOptions: make(map[string]interface{}),
		}
	case DetectorModule:
		module.Detector = &DetectorConfig{
			DefaultOptions: make(map[string]interface{}),
		}
	}
	
	return module
}

// LoadModuleFromFile loads a module from a YAML file
func LoadModuleFromFile(filePath string) (*Module, error) {
	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module file: %w", err)
	}
	
	// Parse YAML
	var module Module
	if err := yaml.Unmarshal(data, &module); err != nil {
		return nil, fmt.Errorf("failed to parse module file: %w", err)
	}
	
	return &module, nil
}

// SaveToFile saves a module to a YAML file
func (m *Module) SaveToFile(filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal module to YAML: %w", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write module file: %w", err)
	}
	
	return nil
}

// Validate validates the module structure and content
func (m *Module) Validate() []string {
	var issues []string
	
	// Validate ID
	if m.ID == "" {
		issues = append(issues, "Module ID is required")
	}
	
	// Validate Type
	if m.Type == "" {
		issues = append(issues, "Module type is required")
	} else {
		switch m.Type {
		case ProviderModule, UtilityModule, DetectorModule:
			// Valid type
		default:
			issues = append(issues, fmt.Sprintf("Invalid module type: %s (valid types: provider, utility, detector)", m.Type))
		}
	}
	
	// Validate Info section
	if m.Info.Name == "" {
		issues = append(issues, "Module name is required")
	}
	
	if m.Info.Description == "" {
		issues = append(issues, "Module description is required")
	}
	
	if m.Info.Version == "" {
		issues = append(issues, "Module version is required")
	}
	
	if m.Info.Author == "" {
		issues = append(issues, "Module author is required")
	}
	
	// Validate Compatibility section
	if m.Compatibility == nil {
		issues = append(issues, "Module compatibility section is required")
	}
	
	// Validate type-specific sections
	switch m.Type {
	case ProviderModule:
		if m.Provider == nil {
			issues = append(issues, "Provider configuration is required for provider modules")
		} else {
			if len(m.Provider.SupportedModels) == 0 {
				issues = append(issues, "At least one supported model is required for provider modules")
			}
		}
	case UtilityModule:
		if m.Utility == nil {
			issues = append(issues, "Utility configuration is required for utility modules")
		} else {
			if len(m.Utility.Functions) == 0 {
				issues = append(issues, "At least one function is required for utility modules")
			}
		}
	case DetectorModule:
		if m.Detector == nil {
			issues = append(issues, "Detector configuration is required for detector modules")
		} else {
			if m.Detector.DetectionType == "" {
				issues = append(issues, "Detection type is required for detector modules")
			}
		}
	}
	
	return issues
}

// GetModuleType returns the module type based on the directory
func GetModuleType(dirPath string) (ModuleType, error) {
	dir := filepath.Base(dirPath)
	
	switch dir {
	case "providers":
		return ProviderModule, nil
	case "utils":
		return UtilityModule, nil
	case "detectors":
		return DetectorModule, nil
	default:
		return "", fmt.Errorf("unknown module directory: %s", dir)
	}
}
