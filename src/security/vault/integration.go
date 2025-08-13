// Package vault provides a secure credential management system for the LLMreconing Tool.
package vault

import (
	"fmt"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/config"
	securityAudit "github.com/perplext/LLMrecon/src/security/audit"
	"github.com/perplext/LLMrecon/src/provider/core"
)

// ConfigIntegration integrates the secure vault with the existing configuration system
type ConfigIntegration struct {
	// credManager is the credential manager
	credManager *CredentialManager
	// config is the application configuration
	config *config.Config
	// mutex protects the integration during operations
	mutex sync.RWMutex
	// initialized indicates whether the integration has been initialized
	initialized bool
}

// NewConfigIntegration creates a new configuration integration
func NewConfigIntegration(credManager *CredentialManager, cfg *config.Config) *ConfigIntegration {
	return &ConfigIntegration{
		credManager: credManager,
		config:      cfg,
	}
}

// Initialize initializes the integration
func (i *ConfigIntegration) Initialize() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.initialized {
		return nil
	}

	// Import existing API keys from config
	if err := i.importExistingAPIKeys(); err != nil {
		return fmt.Errorf("failed to import existing API keys: %w", err)
	}

	i.initialized = true
	return nil
}

// importExistingAPIKeys imports existing API keys from the configuration
func (i *ConfigIntegration) importExistingAPIKeys() error {
	// Import OpenAI API key
	if i.config.APIKeys.OpenAI != "" {
		if err := i.credManager.SetAPIKey(
			core.OpenAIProvider,
			i.config.APIKeys.OpenAI,
			"Imported from config file",
		); err != nil {
			return fmt.Errorf("failed to import OpenAI API key: %w", err)
		}
	}

	// Import Anthropic API key
	if i.config.APIKeys.Anthropic != "" {
		if err := i.credManager.SetAPIKey(
			core.AnthropicProvider,
			i.config.APIKeys.Anthropic,
			"Imported from config file",
		); err != nil {
			return fmt.Errorf("failed to import Anthropic API key: %w", err)
		}
	}

	return nil
}

// UpdateConfig updates the application configuration with credentials from the vault
func (i *ConfigIntegration) UpdateConfig() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Get OpenAI API key
	openaiKey, err := i.credManager.GetAPIKey(core.OpenAIProvider)
	if err == nil && openaiKey != "" {
		i.config.APIKeys.OpenAI = openaiKey
	}

	// Get Anthropic API key
	anthropicKey, err := i.credManager.GetAPIKey(core.AnthropicProvider)
	if err == nil && anthropicKey != "" {
		i.config.APIKeys.Anthropic = anthropicKey
	}

	return nil
}

// SaveConfig saves the application configuration without sensitive data
func (i *ConfigIntegration) SaveConfig() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Create a copy of the config without API keys
	configCopy := *i.config
	configCopy.APIKeys.OpenAI = ""
	configCopy.APIKeys.Anthropic = ""

	// Save the config
	return config.SaveConfig(&configCopy)
}

// SetupGitIgnore ensures that sensitive files are added to .gitignore
func (i *ConfigIntegration) SetupGitIgnore() error {
	// Find git directory
	gitDir, err := findGitDir()
	if err != nil {
		return fmt.Errorf("failed to find git directory: %w", err)
	}

	// Get .gitignore path
	gitignorePath := filepath.Join(filepath.Dir(gitDir), ".gitignore")

	// Read existing .gitignore
	var existingContent string
	if _, err := os.Stat(gitignorePath); err == nil {
		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			return fmt.Errorf("failed to read .gitignore: %w", err)
		}
		existingContent = string(content)
	}

	// Check if we need to add entries
	entriesToAdd := []string{
		"# LLMrecon sensitive files",
		".env",
		"*.env",
		"credentials.vault",
		"credentials.enc",
		"users.json",
		".LLMrecon.yaml",
	}

	// Check if entries already exist
	needsUpdate := false
	for _, entry := range entriesToAdd {
		if !strings.Contains(existingContent, entry) {
			needsUpdate = true
			break
		}
	}

	// Update .gitignore if needed
	if needsUpdate {
		// Prepare new content
		var newContent string
		if existingContent != "" {
			if !strings.HasSuffix(existingContent, "\n") {
				existingContent += "\n"
			}
			newContent = existingContent + "\n"
		}

		// Add entries
		for _, entry := range entriesToAdd {
			if !strings.Contains(existingContent, entry) {
				newContent += entry + "\n"
			}
		}

		// Write updated .gitignore
		if err := os.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update .gitignore: %w", err)
		}
	}

	return nil
}

// DefaultIntegration is the default configuration integration
var DefaultIntegration *ConfigIntegration

// InitDefaultIntegration initializes the default configuration integration
func InitDefaultIntegration(configDir string, passphrase string, auditLogger *securityAudit.AuditLoggerAdapter) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize credential manager
	if err := InitDefaultManager(ManagerOptions{
		ConfigDir:            configDir,
		Passphrase:           passphrase,
		EnvPrefix:            "LLMRT",
		AuditLogger:          auditLogger,
		AutoSave:             true,
		RotationCheckInterval: 24 * 60 * 60 * 1000000000, // 24 hours
		InstallGitHook:       true,
	}); err != nil {
		return fmt.Errorf("failed to initialize credential manager: %w", err)
	}

	// Create integration
	DefaultIntegration = NewConfigIntegration(DefaultManager, cfg)

	// Initialize integration
	if err := DefaultIntegration.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize integration: %w", err)
	}

	// Setup .gitignore
	if err := DefaultIntegration.SetupGitIgnore(); err != nil {
		// Log error but continue
		fmt.Printf("Warning: Failed to setup .gitignore: %v\n", err)
	}

	return nil
}
