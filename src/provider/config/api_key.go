package config

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// SetProviderAPIKey sets the API key for a provider
func (m *ConfigManager) SetProviderAPIKey(providerType core.ProviderType, apiKey string) error {
	// Check if the API key is empty
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Get the existing config or create a new one
	var config *core.ProviderConfig
	var isNewConfig bool
	
	// First check if the config exists
	m.mutex.RLock()
	existingConfig, exists := m.configs[providerType]
	m.mutex.RUnlock()
	
	if !exists {
		// Create a new config if it doesn't exist
		config = &core.ProviderConfig{
			Type:   providerType,
			APIKey: apiKey,
		}
		isNewConfig = true
	} else {
		// Make a copy of the existing config and update the API key
		config = &core.ProviderConfig{}
		*config = *existingConfig
		config.APIKey = apiKey
	}

	// Update the configs map with the new or updated config
	m.mutex.Lock()
	// Create a deep copy to avoid modifying the original
	configCopy := *config
	
	// Encrypt sensitive data if encryption is available
	if m.encryptData != nil {
		if err := m.EncryptSensitiveData(&configCopy); err != nil {
			m.mutex.Unlock()
			return fmt.Errorf("failed to encrypt sensitive data: %w", err)
		}
	}
	
	// Set config
	m.configs[providerType] = &configCopy
	
	// Prepare changes message
	changes := "Updated API key"
	if isNewConfig {
		changes = "Initial configuration with API key"
	}
	
	// Release the lock before I/O operations
	m.mutex.Unlock()
	
	// Add a version to the history
	if err := m.AddConfigVersion(providerType, changes); err != nil {
		return fmt.Errorf("failed to add config version: %w", err)
	}
	
	// Save the updated configs
	if err := m.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
