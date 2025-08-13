// Package plugin provides functionality for dynamically loading provider plugins.
package plugin

import (
	"github.com/perplext/LLMrecon/src/provider/core"
)

// PluginMetadata represents metadata about a provider plugin
type PluginMetadata struct {
	// Name is the name of the plugin
	Name string `json:"name"`
	// Version is the version of the plugin
	Version string `json:"version"`
	// Author is the author of the plugin
	Author string `json:"author"`
	// Description is a description of the plugin
	Description string `json:"description"`
	// ProviderType is the type of provider
	ProviderType core.ProviderType `json:"provider_type"`
	// SupportedModels is a list of models supported by the plugin
	SupportedModels []string `json:"supported_models,omitempty"`
	// MinFrameworkVersion is the minimum framework version required by the plugin
	MinFrameworkVersion string `json:"min_framework_version"`
	// MaxFrameworkVersion is the maximum framework version supported by the plugin
	MaxFrameworkVersion string `json:"max_framework_version,omitempty"`
	// Tags is a list of tags for the plugin
	Tags []string `json:"tags,omitempty"`
}

// PluginInterface defines the interface that all provider plugins must implement
type PluginInterface interface {
	// GetMetadata returns metadata about the plugin
	GetMetadata() *PluginMetadata
	
	// CreateProvider creates a new provider instance
	CreateProvider(config *core.ProviderConfig) (core.Provider, error)
	
	// ValidateConfig validates the provider configuration
	ValidateConfig(config *core.ProviderConfig) error
}

// PluginValidator defines the interface for validating plugins
type PluginValidator interface {
	// ValidatePlugin validates a plugin
	ValidatePlugin(plugin *ProviderPlugin) error
	
	// ValidateCompatibility validates compatibility with the framework
	ValidateCompatibility(metadata *PluginMetadata) error
}
