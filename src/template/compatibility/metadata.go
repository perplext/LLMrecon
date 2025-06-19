package compatibility

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// CompatibilityMetadata represents compatibility information for templates and modules
type CompatibilityMetadata struct {
	// Compatible providers (e.g., "openai", "anthropic")
	Providers []string `json:"providers,omitempty" yaml:"providers,omitempty"`
	
	// Minimum tool version required (e.g., "1.0.0")
	MinToolVersion string `json:"min_tool_version,omitempty" yaml:"min_tool_version,omitempty"`
	
	// Maximum tool version supported (e.g., "2.0.0") - optional
	MaxToolVersion string `json:"max_tool_version,omitempty" yaml:"max_tool_version,omitempty"`
	
	// Supported models for each provider
	SupportedModels map[string][]string `json:"supported_models,omitempty" yaml:"supported_models,omitempty"`
	
	// Required features that must be supported
	RequiredFeatures []string `json:"required_features,omitempty" yaml:"required_features,omitempty"`
	
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// NewCompatibilityMetadata creates a new compatibility metadata instance
func NewCompatibilityMetadata() *CompatibilityMetadata {
	return &CompatibilityMetadata{
		Providers:       []string{},
		SupportedModels: make(map[string][]string),
		RequiredFeatures: []string{},
		Metadata:        make(map[string]interface{}),
	}
}

// AddProvider adds a provider to the compatibility metadata
func (cm *CompatibilityMetadata) AddProvider(provider string) {
	// Check if provider already exists
	for _, p := range cm.Providers {
		if p == provider {
			return
		}
	}
	
	cm.Providers = append(cm.Providers, provider)
}

// RemoveProvider removes a provider from the compatibility metadata
func (cm *CompatibilityMetadata) RemoveProvider(provider string) {
	var providers []string
	
	for _, p := range cm.Providers {
		if p != provider {
			providers = append(providers, p)
		}
	}
	
	cm.Providers = providers
	
	// Also remove supported models for this provider
	delete(cm.SupportedModels, provider)
}

// AddSupportedModel adds a supported model for a provider
func (cm *CompatibilityMetadata) AddSupportedModel(provider, model string) {
	// Ensure provider is in the list
	cm.AddProvider(provider)
	
	// Check if model already exists
	models, ok := cm.SupportedModels[provider]
	if !ok {
		models = []string{}
	}
	
	for _, m := range models {
		if m == model {
			return
		}
	}
	
	models = append(models, model)
	cm.SupportedModels[provider] = models
}

// RemoveSupportedModel removes a supported model for a provider
func (cm *CompatibilityMetadata) RemoveSupportedModel(provider, model string) {
	models, ok := cm.SupportedModels[provider]
	if !ok {
		return
	}
	
	var newModels []string
	for _, m := range models {
		if m != model {
			newModels = append(newModels, m)
		}
	}
	
	if len(newModels) == 0 {
		delete(cm.SupportedModels, provider)
	} else {
		cm.SupportedModels[provider] = newModels
	}
}

// AddRequiredFeature adds a required feature
func (cm *CompatibilityMetadata) AddRequiredFeature(feature string) {
	// Check if feature already exists
	for _, f := range cm.RequiredFeatures {
		if f == feature {
			return
		}
	}
	
	cm.RequiredFeatures = append(cm.RequiredFeatures, feature)
}

// RemoveRequiredFeature removes a required feature
func (cm *CompatibilityMetadata) RemoveRequiredFeature(feature string) {
	var features []string
	
	for _, f := range cm.RequiredFeatures {
		if f != feature {
			features = append(features, f)
		}
	}
	
	cm.RequiredFeatures = features
}

// SetMinToolVersion sets the minimum tool version required
func (cm *CompatibilityMetadata) SetMinToolVersion(version string) error {
	// Validate version
	_, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}
	
	cm.MinToolVersion = version
	return nil
}

// SetMaxToolVersion sets the maximum tool version supported
func (cm *CompatibilityMetadata) SetMaxToolVersion(version string) error {
	// Validate version
	_, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}
	
	cm.MaxToolVersion = version
	return nil
}

// IsCompatibleWithToolVersion checks if the metadata is compatible with a tool version
func (cm *CompatibilityMetadata) IsCompatibleWithToolVersion(version string) (bool, error) {
	// Parse version
	toolVersion, err := semver.NewVersion(version)
	if err != nil {
		return false, fmt.Errorf("invalid tool version format: %w", err)
	}
	
	// Check minimum version
	if cm.MinToolVersion != "" {
		minVersion, err := semver.NewVersion(cm.MinToolVersion)
		if err != nil {
			return false, fmt.Errorf("invalid minimum version format: %w", err)
		}
		
		if toolVersion.LessThan(minVersion) {
			return false, nil
		}
	}
	
	// Check maximum version
	if cm.MaxToolVersion != "" {
		maxVersion, err := semver.NewVersion(cm.MaxToolVersion)
		if err != nil {
			return false, fmt.Errorf("invalid maximum version format: %w", err)
		}
		
		if toolVersion.GreaterThan(maxVersion) {
			return false, nil
		}
	}
	
	return true, nil
}

// IsCompatibleWithProvider checks if the metadata is compatible with a provider
func (cm *CompatibilityMetadata) IsCompatibleWithProvider(provider string) bool {
	// If no providers are specified, assume compatible with all
	if len(cm.Providers) == 0 {
		return true
	}
	
	for _, p := range cm.Providers {
		if p == provider {
			return true
		}
	}
	
	return false
}

// IsCompatibleWithModel checks if the metadata is compatible with a model
func (cm *CompatibilityMetadata) IsCompatibleWithModel(provider, model string) bool {
	// Check if provider is compatible
	if !cm.IsCompatibleWithProvider(provider) {
		return false
	}
	
	// If no models are specified for this provider, assume compatible with all
	models, ok := cm.SupportedModels[provider]
	if !ok || len(models) == 0 {
		return true
	}
	
	for _, m := range models {
		// Support wildcard matching (e.g., "gpt-*" matches "gpt-3.5-turbo", "gpt-4", etc.)
		if strings.HasSuffix(m, "*") {
			prefix := strings.TrimSuffix(m, "*")
			if strings.HasPrefix(model, prefix) {
				return true
			}
		} else if m == model {
			return true
		}
	}
	
	return false
}

// HasRequiredFeatures checks if all required features are available
func (cm *CompatibilityMetadata) HasRequiredFeatures(availableFeatures []string) bool {
	if len(cm.RequiredFeatures) == 0 {
		return true
	}
	
	// Create a map for faster lookup
	featureMap := make(map[string]bool)
	for _, f := range availableFeatures {
		featureMap[f] = true
	}
	
	// Check if all required features are available
	for _, f := range cm.RequiredFeatures {
		if !featureMap[f] {
			return false
		}
	}
	
	return true
}

// SetMetadata sets a metadata value
func (cm *CompatibilityMetadata) SetMetadata(key string, value interface{}) {
	cm.Metadata[key] = value
}

// GetMetadata gets a metadata value
func (cm *CompatibilityMetadata) GetMetadata(key string) (interface{}, bool) {
	value, ok := cm.Metadata[key]
	return value, ok
}

// RemoveMetadata removes a metadata value
func (cm *CompatibilityMetadata) RemoveMetadata(key string) {
	delete(cm.Metadata, key)
}

// Merge merges another compatibility metadata into this one
func (cm *CompatibilityMetadata) Merge(other *CompatibilityMetadata) {
	// Merge providers
	for _, provider := range other.Providers {
		cm.AddProvider(provider)
	}
	
	// Merge supported models
	for provider, models := range other.SupportedModels {
		for _, model := range models {
			cm.AddSupportedModel(provider, model)
		}
	}
	
	// Merge required features
	for _, feature := range other.RequiredFeatures {
		cm.AddRequiredFeature(feature)
	}
	
	// Merge metadata
	for key, value := range other.Metadata {
		cm.SetMetadata(key, value)
	}
	
	// Update min/max tool versions if they're more restrictive
	if other.MinToolVersion != "" {
		otherMin, err := semver.NewVersion(other.MinToolVersion)
		if err == nil {
			if cm.MinToolVersion == "" {
				cm.MinToolVersion = other.MinToolVersion
			} else {
				currentMin, err := semver.NewVersion(cm.MinToolVersion)
				if err == nil && otherMin.GreaterThan(currentMin) {
					cm.MinToolVersion = other.MinToolVersion
				}
			}
		}
	}
	
	if other.MaxToolVersion != "" {
		otherMax, err := semver.NewVersion(other.MaxToolVersion)
		if err == nil {
			if cm.MaxToolVersion == "" {
				cm.MaxToolVersion = other.MaxToolVersion
			} else {
				currentMax, err := semver.NewVersion(cm.MaxToolVersion)
				if err == nil && otherMax.LessThan(currentMax) {
					cm.MaxToolVersion = other.MaxToolVersion
				}
			}
		}
	}
}
