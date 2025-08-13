// Package config provides functionality for managing provider configurations.
package config

import (
	"encoding/json"
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// ConfigHistoryFile represents the file where configuration history is stored
type ConfigHistoryFile struct {
	// History is a map of provider types to configuration history
	History map[string]*ConfigHistory `json:"history"`
}

// AddConfigVersion adds a new version to the configuration history
func (m *ConfigManager) AddConfigVersion(providerType core.ProviderType, changes string) error {
	// Create new version
	var newVersion int
	var history *ConfigHistory
	
	// Update the history under a lock
	m.mutex.Lock()
	history, ok := m.history[providerType]
	if !ok {
		history = &ConfigHistory{
			Current:  0,
			Versions: make(map[int]ConfigVersion),
		}
		m.history[providerType] = history
	}
	
	// Create new version
	newVersion = history.Current + 1
	version := ConfigVersion{
		Version:   newVersion,
		Timestamp: time.Now(),
		Changes:   changes,
	}

	// Add version to history
	history.Versions[newVersion] = version
	history.Current = newVersion
	m.mutex.Unlock()

	// Save history without holding the lock
	return m.SaveHistory()
}

// GetConfigHistory returns the configuration history for a provider type
func (m *ConfigManager) GetConfigHistory(providerType core.ProviderType) (*ConfigHistory, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get history for provider
	history, ok := m.history[providerType]
	if !ok {
		return nil, fmt.Errorf("no history found for provider type %s", providerType)
	}

	return history, nil
}

// GetConfigVersion returns a specific version of a provider configuration
func (m *ConfigManager) GetConfigVersion(providerType core.ProviderType, version int) (*ConfigVersion, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get history for provider
	history, ok := m.history[providerType]
	if !ok {
		return nil, fmt.Errorf("no history found for provider type %s", providerType)
	}

	// Get version
	configVersion, ok := history.Versions[version]
	if !ok {
		return nil, fmt.Errorf("version %d not found for provider type %s", version, providerType)
	}

	return &configVersion, nil
}

// GetAllConfigHistory returns the configuration history for all providers
func (m *ConfigManager) GetAllConfigHistory() map[core.ProviderType]*ConfigHistory {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy of the history map
	historyCopy := make(map[core.ProviderType]*ConfigHistory)
	for providerType, history := range m.history {
		historyCopy[providerType] = history
	}

	return historyCopy
}

// LoadHistory loads configuration history from file
func (m *ConfigManager) LoadHistory() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get history file path
	historyFilePath := getHistoryFilePath(m.configFile)

	// Check if file exists
	if _, err := os.Stat(historyFilePath); os.IsNotExist(err) {
		// Initialize empty history
		m.history = make(map[core.ProviderType]*ConfigHistory)
		return nil
	}

	// Read file
	data, err := os.ReadFile(historyFilePath)
	if err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Decrypt data if encryption key is provided
	if len(m.encryptionKey) > 0 {
		decryptedData, err := DecryptData(data, string(m.encryptionKey))
		if err != nil {
			return fmt.Errorf("failed to decrypt history file: %w", err)
		}
		data = decryptedData
	}

	// Parse JSON
	var historyFileData ConfigHistoryFile
	if err := json.Unmarshal(data, &historyFileData); err != nil {
		return fmt.Errorf("failed to parse history file: %w", err)
	}

	// Convert string keys to ProviderType
	m.history = make(map[core.ProviderType]*ConfigHistory)
	for key, history := range historyFileData.History {
		m.history[core.ProviderType(key)] = history
	}

	return nil
}

// SaveHistory saves configuration history to file
func (m *ConfigManager) SaveHistory() error {
	// Get a copy of the history data under a lock
	var historyFilePath string
	var historyData map[string]*ConfigHistory
	var encryptionKey []byte
	
	m.mutex.RLock()
	// Get history file path
	historyFilePath = getHistoryFilePath(m.configFile)

	// Create a deep copy of the history map
	historyData = make(map[string]*ConfigHistory)
	for key, value := range m.history {
		// Create a deep copy of each history entry
		historyCopy := &ConfigHistory{
			Current:  value.Current,
			Versions: make(map[int]ConfigVersion),
		}
		
		// Copy all versions
		for versionNum, version := range value.Versions {
			historyCopy.Versions[versionNum] = version
		}
		
		historyData[string(key)] = historyCopy
	}
	
	// Copy encryption key if available
	if len(m.encryptionKey) > 0 {
		encryptionKey = make([]byte, len(m.encryptionKey))
		copy(encryptionKey, m.encryptionKey)
	}
	m.mutex.RUnlock()

	// Create history file
	historyFileData := ConfigHistoryFile{
		History: historyData,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(historyFileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history to JSON: %w", err)
	}

	// Encrypt data if encryption key is provided
	if len(encryptionKey) > 0 {
		encryptedData, err := EncryptData(data, string(encryptionKey))
		if err != nil {
			return fmt.Errorf("failed to encrypt history file: %w", err)
		}
		data = encryptedData
	}
	
	// Create directory if it doesn't exist
	historyDir := filepath.Dir(historyFilePath)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(historyFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// getHistoryFilePath returns the path to the history file
func getHistoryFilePath(configFile string) string {
	// Get directory and base name
	dir := filepath.Dir(configFile)
	base := filepath.Base(configFile)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	// Create history file path
	return filepath.Join(dir, name+".history"+ext)
}

// RollbackConfig rolls back a provider configuration to a specific version
func (m *ConfigManager) RollbackConfig(providerType core.ProviderType, version int) error {
	m.mutex.Lock()
	
	// Get history for provider
	history, ok := m.history[providerType]
	if !ok {
		m.mutex.Unlock()
		return fmt.Errorf("no history found for provider type %s", providerType)
	}

	// Check if version exists
	_, ok = history.Versions[version]
	if !ok {
		m.mutex.Unlock()
		return fmt.Errorf("version %d not found for provider type %s", version, providerType)
	}

	// Add rollback version
	newVersion := history.Current + 1
	rollbackVersion := ConfigVersion{
		Version:   newVersion,
		Timestamp: time.Now(),
		Changes:   fmt.Sprintf("Rolled back to version %d", version),
	}

	// Add version to history
	history.Versions[newVersion] = rollbackVersion
	history.Current = newVersion
	
	// Release the lock before I/O operations
	m.mutex.Unlock()

	// Save history
	return m.SaveHistory()
}

// GetConfigVersions returns all versions of a provider configuration
func (m *ConfigManager) GetConfigVersions(providerType core.ProviderType) ([]ConfigVersion, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get history for provider
	history, ok := m.history[providerType]
	if !ok {
		return nil, fmt.Errorf("no history found for provider type %s", providerType)
	}

	// Convert map to slice
	versions := make([]ConfigVersion, 0, len(history.Versions))
	for _, version := range history.Versions {
		versions = append(versions, version)
	}

	return versions, nil
}
