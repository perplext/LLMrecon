// Package config provides functionality for managing provider configurations.
package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// ConfigVersion represents a version of the configuration
type ConfigVersion struct {
	// Version is the version number
	Version int `json:"version"`
	// Timestamp is the timestamp of the version
	Timestamp time.Time `json:"timestamp"`
	// Changes is a description of the changes
	Changes string `json:"changes,omitempty"`
}

// ConfigHistory represents the history of configuration changes
type ConfigHistory struct {
	// Current is the current version
	Current int `json:"current"`
	// Versions is a map of version numbers to versions
	Versions map[int]ConfigVersion `json:"versions"`
}

// ConfigManager is responsible for managing provider configurations
type ConfigManager struct {
	// configs is a map of provider types to configurations
	configs map[core.ProviderType]*core.ProviderConfig
	// configFile is the path to the configuration file
	configFile string
	// encryptionKey is the key used for encrypting sensitive data
	encryptionKey []byte
	// mutex is a mutex for concurrent access to configs
	mutex sync.RWMutex
	// envVarPrefix is the prefix for environment variables
	envVarPrefix string
	// history is a map of provider types to configuration history
	history map[core.ProviderType]*ConfigHistory
	// encryptData is a function that encrypts data
	encryptData func(data []byte) ([]byte, error)
	// decryptData is a function that decrypts data
	decryptData func(data []byte) ([]byte, error)
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configFile string, encryptionKey []byte, envVarPrefix string) (*ConfigManager, error) {
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configFile = filepath.Join(homeDir, ".LLMrecon", "provider_config.json")
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set default environment variable prefix
	if envVarPrefix == "" {
		envVarPrefix = "LLM_RED_TEAM"
	}
	
	// Ensure encryption key is valid (32 bytes for AES-256)
	var normalizedKey []byte
	if len(encryptionKey) > 0 {
		// If key is provided but not 32 bytes, hash it to get a 32-byte key
		if len(encryptionKey) != 32 {
			hash := sha256.Sum256(encryptionKey)
			normalizedKey = hash[:]
		} else {
			normalizedKey = encryptionKey
		}
	}

	manager := &ConfigManager{
		configs:       make(map[core.ProviderType]*core.ProviderConfig),
		configFile:    configFile,
		encryptionKey: normalizedKey,
		envVarPrefix:  envVarPrefix,
		history:       make(map[core.ProviderType]*ConfigHistory),
	}

	// Initialize encryption functions
	manager.encryptData = func(data []byte) ([]byte, error) {
		// If no encryption key is provided, return the data as is
		if len(manager.encryptionKey) == 0 {
			return data, nil
		}

		// Create a new AES cipher block
		block, err := aes.NewCipher(manager.encryptionKey)
		if err != nil {
			return nil, err
		}

		// Create a new GCM cipher
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		// Create a nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}

		// Encrypt the data
		return gcm.Seal(nonce, nonce, data, nil), nil
	}

	manager.decryptData = func(data []byte) ([]byte, error) {
		// If no encryption key is provided, return the data as is
		if len(manager.encryptionKey) == 0 {
			return data, nil
		}

		// Create a new AES cipher block
		block, err := aes.NewCipher(manager.encryptionKey)
		if err != nil {
			return nil, err
		}

		// Create a new GCM cipher
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		// Check if the data is long enough
		if len(data) < gcm.NonceSize() {
			return nil, fmt.Errorf("ciphertext too short")
		}

		// Get the nonce and ciphertext
		nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

		// Decrypt the data
		return gcm.Open(nil, nonce, ciphertext, nil)
	}

	// Load configurations from file
	if err := manager.Load(); err != nil {
		// If the file doesn't exist, it's not an error
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load configurations: %w", err)
		}
	}

	// Load configurations from environment variables
	if err := manager.LoadFromEnv(); err != nil {
		return nil, fmt.Errorf("failed to load configurations from environment variables: %w", err)
	}

	return manager, nil
}

// Load loads configurations from the configuration file
func (m *ConfigManager) Load() error {
	// Check if file exists
	if _, err := os.Stat(m.configFile); os.IsNotExist(err) {
		return err
	}

	// Read file - this doesn't need a lock
	data, err := ioutil.ReadFile(m.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Decrypt data if decryption function is available
	if m.decryptData != nil {
		decryptedData, err := m.decryptData(data)
		if err != nil {
			return fmt.Errorf("failed to decrypt config file: %w", err)
		}
		data = decryptedData
	}

	// Parse JSON
	var configs map[string]*core.ProviderConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Now acquire the lock to update the configs
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Convert string keys to ProviderType
	m.configs = make(map[core.ProviderType]*core.ProviderConfig)
	for key, config := range configs {
		m.configs[core.ProviderType(key)] = config
	}

	return nil
}

// Save saves configurations to the configuration file
func (m *ConfigManager) Save() error {
	// Make a copy of the configs to avoid holding the lock during I/O operations
	var configsCopy map[string]*core.ProviderConfig
	
	// Acquire lock only for reading the configs
	m.mutex.RLock()
	// Convert ProviderType keys to string
	configsCopy = make(map[string]*core.ProviderConfig)
	for key, config := range m.configs {
		// Create a deep copy of the config
		configCopy := *config
		configsCopy[string(key)] = &configCopy
	}
	m.mutex.RUnlock()

	// Marshal to JSON
	data, err := json.MarshalIndent(configsCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configs to JSON: %w", err)
	}

	// Encrypt data if encryption key is provided
	if m.encryptData != nil {
		encryptedData, err := m.encryptData(data)
		if err != nil {
			return fmt.Errorf("failed to encrypt config file: %w", err)
		}
		data = encryptedData
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(m.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(m.configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadFromEnv loads configurations from environment variables
func (m *ConfigManager) LoadFromEnv() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get all environment variables
	for _, env := range os.Environ() {
		// Split key and value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Check if key has the correct prefix
		if !strings.HasPrefix(key, m.envVarPrefix+"_") {
			continue
		}

		// Remove prefix
		key = strings.TrimPrefix(key, m.envVarPrefix+"_")

		// Parse key to get provider type and config field
		keyParts := strings.SplitN(key, "_", 2)
		if len(keyParts) != 2 {
			continue
		}

		providerType := core.ProviderType(strings.ToLower(keyParts[0]))
		configField := keyParts[1]

		// Get or create provider config
		config, ok := m.configs[providerType]
		if !ok {
			config = &core.ProviderConfig{
				Type: providerType,
			}
			m.configs[providerType] = config
		}

		// Set config field
		switch strings.ToLower(configField) {
		case "apikey":
			config.APIKey = value
		case "orgid":
			config.OrgID = value
		case "baseurl":
			config.BaseURL = value
		case "timeout":
			timeout, err := time.ParseDuration(value)
			if err == nil {
				config.Timeout = timeout
			}
		case "defaultmodel":
			config.DefaultModel = value
		}
	}

	return nil
}

// GetConfig returns the configuration for a provider type
func (m *ConfigManager) GetConfig(providerType core.ProviderType) (*core.ProviderConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, ok := m.configs[providerType]
	if !ok {
		return nil, fmt.Errorf("configuration for provider type %s not found", providerType)
	}

	// Return a copy of the configuration to prevent modification
	configCopy := *config
	
	// Decrypt sensitive data if decryption is available
	if m.decryptData != nil {
		if err := m.DecryptSensitiveData(&configCopy); err != nil {
			return nil, fmt.Errorf("failed to decrypt sensitive data: %w", err)
		}
	}
	
	return &configCopy, nil
}

// SetConfig sets the configuration for a provider type
func (m *ConfigManager) SetConfig(providerType core.ProviderType, config *core.ProviderConfig) error {
	// Validate config
	if err := m.ValidateConfig(config); err != nil {
		return err
	}

	// Create a deep copy of the config to avoid modifying the original
	configCopy := *config
	
	// Encrypt sensitive data if encryption is available
	if m.encryptData != nil {
		if err := m.EncryptSensitiveData(&configCopy); err != nil {
			return fmt.Errorf("failed to encrypt sensitive data: %w", err)
		}
	}
	
	// Check if this is a new config or an update and update the configs map
	var exists bool
	var existingConfig *core.ProviderConfig
	
	m.mutex.Lock()
	existingConfig, exists = m.configs[providerType]
	m.configs[providerType] = &configCopy
	m.mutex.Unlock()

	// Save config without holding the lock
	if err := m.Save(); err != nil {
		return err
	}

	// Add version to history
	changes := "Initial configuration"
	if exists {
		changes = "Updated configuration"
		if existingConfig.APIKey != config.APIKey {
			changes += ", changed API key"
		}
		if existingConfig.BaseURL != config.BaseURL {
			changes += fmt.Sprintf(", changed base URL from %s to %s", existingConfig.BaseURL, config.BaseURL)
		}
		if existingConfig.Timeout != config.Timeout {
			changes += fmt.Sprintf(", changed timeout from %v to %v", existingConfig.Timeout, config.Timeout)
		}
		if existingConfig.RetryConfig != nil && config.RetryConfig != nil && 
		   existingConfig.RetryConfig.MaxRetries != config.RetryConfig.MaxRetries {
			changes += fmt.Sprintf(", changed max retries from %d to %d", 
				existingConfig.RetryConfig.MaxRetries, config.RetryConfig.MaxRetries)
		}
	}

	// Add version to history
	return m.AddConfigVersion(providerType, changes)
}

// UpdateConfig updates the configuration for a provider type
func (m *ConfigManager) UpdateConfig(providerType core.ProviderType, updates *core.ProviderConfig) error {
	m.mutex.Lock()
	
	// Get existing config
	config, ok := m.configs[providerType]
	if !ok {
		m.mutex.Unlock()
		return fmt.Errorf("no configuration found for provider type %s", providerType)
	}

	// Track changes for versioning
	changes := "Updated configuration:"

	// Update config
	if updates.APIKey != "" {
		// Store the API key directly (encryption will be handled by Save)
		config.APIKey = updates.APIKey
		changes += " changed API key;"
	}

	if updates.OrgID != "" && updates.OrgID != config.OrgID {
		config.OrgID = updates.OrgID
		changes += fmt.Sprintf(" changed organization ID from '%s' to '%s';", config.OrgID, updates.OrgID)
	}

	if updates.BaseURL != "" && updates.BaseURL != config.BaseURL {
		changes += fmt.Sprintf(" changed base URL from '%s' to '%s';", config.BaseURL, updates.BaseURL)
		config.BaseURL = updates.BaseURL
	}

	if updates.Timeout > 0 && updates.Timeout != config.Timeout {
		changes += fmt.Sprintf(" changed timeout from %v to %v;", config.Timeout, updates.Timeout)
		config.Timeout = updates.Timeout
	}

	if updates.DefaultModel != "" && updates.DefaultModel != config.DefaultModel {
		changes += fmt.Sprintf(" changed default model from '%s' to '%s';", config.DefaultModel, updates.DefaultModel)
		config.DefaultModel = updates.DefaultModel
	}

	if updates.RetryConfig != nil {
		if config.RetryConfig == nil {
			config.RetryConfig = &core.RetryConfig{}
			changes += " added retry configuration;"
		} else {
			changes += " updated retry configuration;"
		}

		if updates.RetryConfig.MaxRetries > 0 && updates.RetryConfig.MaxRetries != config.RetryConfig.MaxRetries {
			config.RetryConfig.MaxRetries = updates.RetryConfig.MaxRetries
		}

		if updates.RetryConfig.InitialBackoff > 0 && updates.RetryConfig.InitialBackoff != config.RetryConfig.InitialBackoff {
			config.RetryConfig.InitialBackoff = updates.RetryConfig.InitialBackoff
		}

		if updates.RetryConfig.MaxBackoff > 0 && updates.RetryConfig.MaxBackoff != config.RetryConfig.MaxBackoff {
			config.RetryConfig.MaxBackoff = updates.RetryConfig.MaxBackoff
		}

		if updates.RetryConfig.BackoffMultiplier > 0 && updates.RetryConfig.BackoffMultiplier != config.RetryConfig.BackoffMultiplier {
			config.RetryConfig.BackoffMultiplier = updates.RetryConfig.BackoffMultiplier
		}

		if updates.RetryConfig.RetryableStatusCodes != nil {
			config.RetryConfig.RetryableStatusCodes = updates.RetryConfig.RetryableStatusCodes
		}
	}

	if updates.RateLimitConfig != nil {
		if config.RateLimitConfig == nil {
			config.RateLimitConfig = &core.RateLimitConfig{}
			changes += " added rate limit configuration;"
		} else {
			changes += " updated rate limit configuration;"
		}

		if updates.RateLimitConfig.RequestsPerMinute > 0 && updates.RateLimitConfig.RequestsPerMinute != config.RateLimitConfig.RequestsPerMinute {
			config.RateLimitConfig.RequestsPerMinute = updates.RateLimitConfig.RequestsPerMinute
		}

		if updates.RateLimitConfig.TokensPerMinute > 0 && updates.RateLimitConfig.TokensPerMinute != config.RateLimitConfig.TokensPerMinute {
			config.RateLimitConfig.TokensPerMinute = updates.RateLimitConfig.TokensPerMinute
		}

		if updates.RateLimitConfig.MaxConcurrentRequests > 0 && updates.RateLimitConfig.MaxConcurrentRequests != config.RateLimitConfig.MaxConcurrentRequests {
			config.RateLimitConfig.MaxConcurrentRequests = updates.RateLimitConfig.MaxConcurrentRequests
		}

		if updates.RateLimitConfig.BurstSize > 0 && updates.RateLimitConfig.BurstSize != config.RateLimitConfig.BurstSize {
			config.RateLimitConfig.BurstSize = updates.RateLimitConfig.BurstSize
		}
	}

	// Update additional headers if provided
	if updates.AdditionalHeaders != nil && len(updates.AdditionalHeaders) > 0 {
		if config.AdditionalHeaders == nil {
			config.AdditionalHeaders = make(map[string]string)
			changes += " added additional headers;"
		} else {
			changes += " updated additional headers;"
		}

		for k, v := range updates.AdditionalHeaders {
			config.AdditionalHeaders[k] = v
		}
	}

	// Update additional params if provided
	if updates.AdditionalParams != nil && len(updates.AdditionalParams) > 0 {
		if config.AdditionalParams == nil {
			config.AdditionalParams = make(map[string]interface{})
			changes += " added additional parameters;"
		} else {
			changes += " updated additional parameters;"
		}

		for k, v := range updates.AdditionalParams {
			config.AdditionalParams[k] = v
		}
	}

	// Make a copy of the updated config and store it in the map
	m.configs[providerType] = config
	
	// Release the lock before I/O operations
	m.mutex.Unlock()

	// Save config
	if err := m.Save(); err != nil {
		return err
	}

	// Add version to history
	return m.AddConfigVersion(providerType, changes)
}

// DeleteConfig deletes the configuration for a provider type
func (m *ConfigManager) DeleteConfig(providerType core.ProviderType) error {
	m.mutex.Lock()
	
	// Check if config exists
	_, ok := m.configs[providerType]
	if !ok {
		m.mutex.Unlock()
		return fmt.Errorf("no configuration found for provider type %s", providerType)
	}

	// Delete config
	delete(m.configs, providerType)
	
	// Release the lock before I/O operations
	m.mutex.Unlock()

	// Save config
	if err := m.Save(); err != nil {
		return err
	}

	// Add version to history
	return m.AddConfigVersion(providerType, "Configuration deleted")
}

// GetAllConfigs returns all configurations
func (m *ConfigManager) GetAllConfigs() map[core.ProviderType]*core.ProviderConfig {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy of the configurations to prevent modification
	configsCopy := make(map[core.ProviderType]*core.ProviderConfig)
	for providerType, config := range m.configs {
		configCopy := *config
		configsCopy[providerType] = &configCopy
	}

	return configsCopy
}

// GetAllProviderTypes returns all provider types with configurations
func (m *ConfigManager) GetAllProviderTypes() []core.ProviderType {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	providerTypes := make([]core.ProviderType, 0, len(m.configs))
	for providerType := range m.configs {
		providerTypes = append(providerTypes, providerType)
	}

	return providerTypes
}

// SetEncryptionKey sets the encryption key
func (m *ConfigManager) SetEncryptionKey(encryptionKey []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.encryptionKey = encryptionKey
}

// SetConfigFile sets the configuration file path
func (m *ConfigManager) SetConfigFile(configFile string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.configFile = configFile
}

// SetEnvVarPrefix sets the environment variable prefix
func (m *ConfigManager) SetEnvVarPrefix(envVarPrefix string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.envVarPrefix = envVarPrefix
}

// ValidateConfig validates a provider configuration
func (m *ConfigManager) ValidateConfig(config *core.ProviderConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if config.Type == "" {
		return fmt.Errorf("provider type is required")
	}

	// Validate required fields based on provider type
	switch config.Type {
	case core.OpenAIProvider, core.AzureOpenAIProvider:
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for provider type %s", config.Type)
		}
	case core.AnthropicProvider:
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for provider type %s", config.Type)
		}
	case core.HuggingFaceProvider:
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for provider type %s", config.Type)
		}
	}

	return nil
}

// encrypt encrypts data using AES-GCM encryption
func encrypt(data []byte, key []byte) ([]byte, error) {
	// Ensure the key is the right size for AES-256
	if len(key) < 32 {
		// If key is too short, derive a 32-byte key using SHA-256
		hash := sha256.Sum256(key)
		key = hash[:]
	} else if len(key) > 32 {
		// If key is too long, truncate it
		key = key[:32]
	}

	// Create a new cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and seal the data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM decryption
func decrypt(data []byte, key []byte) ([]byte, error) {
	// Ensure the key is the right size for AES-256
	if len(key) < 32 {
		// If key is too short, derive a 32-byte key using SHA-256
		hash := sha256.Sum256(key)
		key = hash[:]
	} else if len(key) > 32 {
		// If key is too long, truncate it
		key = key[:32]
	}

	// Create a new cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Ensure the data is long enough
	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract the nonce and ciphertext
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
