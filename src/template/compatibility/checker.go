package compatibility

import (
	"fmt"
)

// ProviderInfo represents information about an LLM provider
type ProviderInfo struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	SupportedModels []string `json:"supported_models"`
	Features       []string `json:"features"`
}

// Checker handles compatibility checking between templates and providers
type Checker struct {
	toolVersion    string
	providers      map[string]*ProviderInfo
	availableFeatures []string
}

// NewChecker creates a new compatibility checker
func NewChecker(toolVersion string) *Checker {
	return &Checker{
		toolVersion:    toolVersion,
		providers:      make(map[string]*ProviderInfo),
		availableFeatures: []string{},
	}
}

// RegisterProvider registers a provider with the checker
func (c *Checker) RegisterProvider(provider *ProviderInfo) {
	c.providers[provider.ID] = provider
}

// UnregisterProvider removes a provider from the checker
func (c *Checker) UnregisterProvider(providerID string) {
	delete(c.providers, providerID)
}

// AddAvailableFeature adds an available feature
func (c *Checker) AddAvailableFeature(feature string) {
	// Check if feature already exists
	for _, f := range c.availableFeatures {
		if f == feature {
			return
		}
	}
	
	c.availableFeatures = append(c.availableFeatures, feature)

// RemoveAvailableFeature removes an available feature
}
func (c *Checker) RemoveAvailableFeature(feature string) {
	var features []string
	
	for _, f := range c.availableFeatures {
		if f != feature {
			features = append(features, f)
		}
	}
	
	c.availableFeatures = features

// GetAvailableFeatures returns the list of available features
}
func (c *Checker) GetAvailableFeatures() []string {
	return c.availableFeatures

// GetProviders returns the list of registered providers
}
func (c *Checker) GetProviders() map[string]*ProviderInfo {
	return c.providers

// GetProvider returns information about a specific provider
}
func (c *Checker) GetProvider(providerID string) (*ProviderInfo, bool) {
	provider, ok := c.providers[providerID]
	return provider, ok

// IsModelSupported checks if a model is supported by a provider
}
func (c *Checker) IsModelSupported(providerID, model string) bool {
	provider, ok := c.providers[providerID]
	if !ok {
		return false
	}
	
	for _, m := range provider.SupportedModels {
		if m == model {
			return true
		}
	}
	
	return false

// CheckCompatibility checks if a template is compatible with the current environment
}
func (c *Checker) CheckCompatibility(metadata *CompatibilityMetadata) (bool, []string) {
	var issues []string
	
	// Check tool version compatibility
	compatible, err := metadata.IsCompatibleWithToolVersion(c.toolVersion)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Error checking tool version compatibility: %v", err))
	} else if !compatible {
		issues = append(issues, fmt.Sprintf("Incompatible tool version: %s (required: min=%s, max=%s)",
			c.toolVersion, metadata.MinToolVersion, metadata.MaxToolVersion))
	}
	
	// Check provider compatibility
	if len(metadata.Providers) > 0 {
		var foundCompatibleProvider bool
		
		for _, providerID := range metadata.Providers {
			_, ok := c.providers[providerID]
			if ok {
				foundCompatibleProvider = true
				break
			}
		}
		
		if !foundCompatibleProvider {
			issues = append(issues, fmt.Sprintf("No compatible provider found. Required: %v, Available: %v",
				metadata.Providers, getProviderIDs(c.providers)))
		}
	}
	
	// Check required features
	if !metadata.HasRequiredFeatures(c.availableFeatures) {
		missingFeatures := getMissingFeatures(metadata.RequiredFeatures, c.availableFeatures)
		issues = append(issues, fmt.Sprintf("Missing required features: %v", missingFeatures))
	}
	
	return len(issues) == 0, issues

// CheckProviderCompatibility checks if a template is compatible with a specific provider
}
func (c *Checker) CheckProviderCompatibility(metadata *CompatibilityMetadata, providerID string) (bool, []string) {
	var issues []string
	
	// Check if provider is registered
	provider, ok := c.providers[providerID]
	if !ok {
		issues = append(issues, fmt.Sprintf("Provider '%s' is not registered", providerID))
		return false, issues
	}
	
	// Check if template is compatible with this provider
	if !metadata.IsCompatibleWithProvider(providerID) {
		issues = append(issues, fmt.Sprintf("Template is not compatible with provider '%s'", providerID))
	}
	
	// Check if provider supports the required features
	for _, feature := range metadata.RequiredFeatures {
		var featureSupported bool
		
		for _, f := range provider.Features {
			if f == feature {
				featureSupported = true
				break
			}
		}
		
		if !featureSupported {
			issues = append(issues, fmt.Sprintf("Provider '%s' does not support required feature '%s'",
				providerID, feature))
		}
	}
	
	return len(issues) == 0, issues

// CheckModelCompatibility checks if a template is compatible with a specific model
}
func (c *Checker) CheckModelCompatibility(metadata *CompatibilityMetadata, providerID, model string) (bool, []string) {
	var issues []string
	
	// First check provider compatibility
	compatible, providerIssues := c.CheckProviderCompatibility(metadata, providerID)
	if !compatible {
		issues = append(issues, providerIssues...)
	}
	
	// Check if provider is registered
	provider, ok := c.providers[providerID]
	if !ok {
		// Already reported in provider compatibility check
		return false, issues
	}
	
	// Check if provider supports this model
	var modelSupported bool
	for _, m := range provider.SupportedModels {
		if m == model {
			modelSupported = true
			break
		}
	}
	
	if !modelSupported {
		issues = append(issues, fmt.Sprintf("Provider '%s' does not support model '%s'",
			providerID, model))
	}
	
	// Check if template is compatible with this model
	if !metadata.IsCompatibleWithModel(providerID, model) {
		issues = append(issues, fmt.Sprintf("Template is not compatible with model '%s' from provider '%s'",
			model, providerID))
	}
	
	return len(issues) == 0, issues

// Helper function to get provider IDs from the providers map
}
func getProviderIDs(providers map[string]*ProviderInfo) []string {
	var ids []string
	for id := range providers {
		ids = append(ids, id)
	}
	return ids

// Helper function to get missing features
}
func getMissingFeatures(required, available []string) []string {
	var missing []string
	
	// Create a map for faster lookup
	availableMap := make(map[string]bool)
	for _, f := range available {
		availableMap[f] = true
	}
	
	// Find missing features
	for _, f := range required {
		if !availableMap[f] {
			missing = append(missing, f)
		}
	}
	
	return missing
}
