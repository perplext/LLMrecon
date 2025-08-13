// Package plugin provides functionality for dynamically loading provider plugins.
package plugin

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/perplext/LLMrecon/src/version"
)

// DefaultPluginValidator is the default implementation of PluginValidator
type DefaultPluginValidator struct {
	// currentVersion is the current framework version
	currentVersion *semver.Version
}

// NewDefaultPluginValidator creates a new default plugin validator
func NewDefaultPluginValidator() *DefaultPluginValidator {
	// Get framework info
	frameworkInfo := version.GetFrameworkInfo()
	
	// Parse current framework version
	currentVersion, err := semver.NewVersion(frameworkInfo.Version)
	if err != nil {
		// If version parsing fails, use a default version
		currentVersion, _ = semver.NewVersion("0.1.0")
	}
	
	return &DefaultPluginValidator{
		currentVersion: currentVersion,
	}
}

// ValidatePlugin validates a plugin
func (v *DefaultPluginValidator) ValidatePlugin(plugin *ProviderPlugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin is nil")
	}
	
	// Validate plugin metadata
	if err := v.ValidateCompatibility(plugin.Metadata); err != nil {
		return err
	}
	
	// Validate provider type
	if plugin.ProviderType == "" {
		return fmt.Errorf("plugin provider type is empty")
	}
	
	return nil
}

// ValidateCompatibility validates compatibility with the framework
func (v *DefaultPluginValidator) ValidateCompatibility(metadata *PluginMetadata) error {
	if metadata == nil {
		return fmt.Errorf("plugin metadata is nil")
	}
	
	// Check if minimum framework version is specified
	if metadata.MinFrameworkVersion == "" {
		return nil // No version constraint
	}
	
	// Parse minimum framework version
	minVersion, err := semver.NewVersion(metadata.MinFrameworkVersion)
	if err != nil {
		return fmt.Errorf("invalid minimum framework version: %s: %w", metadata.MinFrameworkVersion, err)
	}
	
	// Check if current version is greater than or equal to minimum version
	if v.currentVersion.LessThan(minVersion) {
		return fmt.Errorf("plugin requires minimum framework version %s, but current version is %s", 
			metadata.MinFrameworkVersion, v.currentVersion.String())
	}
	
	// Check if maximum framework version is specified
	if metadata.MaxFrameworkVersion != "" {
		// Parse maximum framework version
		maxVersion, err := semver.NewVersion(metadata.MaxFrameworkVersion)
		if err != nil {
			return fmt.Errorf("invalid maximum framework version: %s: %w", metadata.MaxFrameworkVersion, err)
		}
		
		// Check if current version is less than or equal to maximum version
		if v.currentVersion.GreaterThan(maxVersion) {
			return fmt.Errorf("plugin supports maximum framework version %s, but current version is %s", 
				metadata.MaxFrameworkVersion, v.currentVersion.String())
		}
	}
	
	return nil
}
