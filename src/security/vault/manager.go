// Package vault provides a secure credential management system for the LLMreconing Tool.
package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	securityAudit "github.com/perplext/LLMrecon/src/security/audit"
	"github.com/perplext/LLMrecon/src/provider/core"
)

// CredentialManager manages credentials for the application
type CredentialManager struct {
	// vault is the secure vault for storing credentials
	vault *SecureVault
	// envPrefix is the prefix for environment variables
	envPrefix string
	// configDir is the directory for configuration files
	configDir string
	// mutex protects the manager during operations
	mutex sync.RWMutex
	// serviceToProviderMap maps service names to provider types
	serviceToProviderMap map[string]core.ProviderType
	// providerToServiceMap maps provider types to service names
	providerToServiceMap map[core.ProviderType]string
	// gitHookInstalled tracks whether the git hook is installed
	gitHookInstalled bool
}

// ManagerOptions contains options for creating a credential manager
type ManagerOptions struct {
	// ConfigDir is the directory for configuration files
	ConfigDir string
	// Passphrase is used to derive the encryption key
	Passphrase string
	// EnvPrefix is the prefix for environment variables
	EnvPrefix string
	// AuditLogger is used for logging credential access
	AuditLogger *securityAudit.AuditLoggerAdapter
	// AutoSave determines whether to automatically save after changes
	AutoSave bool
	// RotationCheckInterval is how often to check for credentials that need rotation
	RotationCheckInterval time.Duration
	// InstallGitHook determines whether to install a git hook to prevent credential leakage
	InstallGitHook bool
	// GitDir is the custom git directory to use for git hook installation (for testing)
	GitDir string
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(options ManagerOptions) (*CredentialManager, error) {
	// Set default config directory if not specified
	configDir := options.ConfigDir
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".LLMrecon")
	}

	// Set default environment variable prefix if not specified
	envPrefix := options.EnvPrefix
	if envPrefix == "" {
		envPrefix = "LLMRT"
	}

	// Create secure vault
	vaultPath := filepath.Join(configDir, "credentials.vault")
	vault, err := NewSecureVault(vaultPath, VaultOptions{
		Passphrase:           options.Passphrase,
		AuditLogger:          options.AuditLogger,
		AutoSave:             options.AutoSave,
		RotationCheckInterval: options.RotationCheckInterval,
		AlertCallback: func(credential *Credential, daysUntilExpiration int) {
			// Log alert about credential rotation
			if options.AuditLogger != nil {
				options.AuditLogger.LogAlert(
					fmt.Sprintf("Credential '%s' for service '%s' needs rotation in %d days", 
						credential.Name, 
						credential.Service, 
						daysUntilExpiration),
					"credential_rotation",
					map[string]string{
						"credential_id":        credential.ID,
						"credential_name":      credential.Name,
						"service":              credential.Service,
						"days_until_expiration": fmt.Sprintf("%d", daysUntilExpiration),
					},
				)
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create secure vault: %w", err)
	}

	// Create manager
	manager := &CredentialManager{
		vault:               vault,
		envPrefix:           envPrefix,
		configDir:           configDir,
		serviceToProviderMap: make(map[string]core.ProviderType),
		providerToServiceMap: make(map[core.ProviderType]string),
	}

	// Initialize service to provider mappings
	manager.initServiceMappings()

	// Install git hook if requested
	if options.InstallGitHook {
		var err error
		if options.GitDir != "" {
			err = manager.InstallGitHookInDir(options.GitDir)
		} else {
			err = manager.InstallGitHook()
		}
		
		if err != nil {
			fmt.Printf("Warning: Failed to install git hook: %v\n", err)
		} else {
			manager.gitHookInstalled = true
		}
	}

	// Load credentials from environment variables
	if err := manager.LoadFromEnv(); err != nil {
		// Log error but continue
		fmt.Printf("Warning: Failed to load credentials from environment variables: %v\n", err)
	}

	return manager, nil
}

// initServiceMappings initializes the service to provider mappings
func (m *CredentialManager) initServiceMappings() {
	// Map service names to provider types
	m.serviceToProviderMap = map[string]core.ProviderType{
		"openai":   core.OpenAIProvider,
		"anthropic": core.AnthropicProvider,
		"azure-openai": core.AzureOpenAIProvider,
		"huggingface": core.HuggingFaceProvider,
		"local": core.LocalProvider,
	}

	// Create reverse mapping
	for service, provider := range m.serviceToProviderMap {
		m.providerToServiceMap[provider] = service
	}
}

// InstallGitHook installs a git hook to prevent credential leakage
func (m *CredentialManager) InstallGitHook() error {
	return m.InstallGitHookInDir("")
}

// InstallGitHookInDir installs a git hook in the specified directory or finds the git directory if empty
func (m *CredentialManager) InstallGitHookInDir(customDir string) error {
	// Find git directory
	var gitDir string
	var err error
	
	if customDir != "" {
		// Use the custom directory for testing
		gitDir = filepath.Join(customDir, ".git")
		// Verify it exists
		if _, err := os.Stat(gitDir); err != nil {
			return fmt.Errorf("custom git directory not found: %w", err)
		}
	} else {
		// Find git directory in the current working directory
		gitDir, err = findGitDir()
		if err != nil {
			return fmt.Errorf("failed to find git directory: %w", err)
		}
	}

	// Create pre-commit hook
	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
	
	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		// Read existing hook
		existingHook, err := os.ReadFile(hookPath)
		if err != nil {
			return fmt.Errorf("failed to read existing git hook: %w", err)
		}

		// Check if our hook is already installed
		if strings.Contains(string(existingHook), "# LLMrecon credential check") {
			// Hook already installed
			return nil
		}

		// Backup existing hook
		backupPath := hookPath + ".backup"
		if err := os.WriteFile(backupPath, existingHook, 0755); err != nil {
			return fmt.Errorf("failed to backup existing git hook: %w", err)
		}
	}

	// Create hook content with proper escaping and error handling
	hookContent := `#!/bin/bash
# LLMrecon credential check
# This hook prevents committing files that might contain API keys or credentials

# Function to display error message
show_error() {
  echo "\033[1;31mERROR:\033[0m $1"
  echo "$2"
  exit 1
}

# Check for potential API keys and tokens
if git diff --cached | grep -E '(api[_-]?key|api[_-]?token|access[_-]?token|secret[_-]?key|password|credential)["'\''']?\s*[:=]\s*["'\''']?[A-Za-z0-9_\-]{20,}'; then
  show_error "Potential API key or credential found in commit." "Please remove the credential or add it to .gitignore."
fi

# Check for .env files
if git diff --cached --name-only | grep -E '\.env(\..*)?$'; then
  show_error "Attempting to commit .env file." "Please add .env files to .gitignore."
fi

# Check for credential files
if git diff --cached --name-only | grep -E '(credentials?|secrets?|api[_-]?keys?|auth)\.(json|yaml|yml|xml|txt|ini|conf)$'; then
  show_error "Attempting to commit a credential file." "Please add credential files to .gitignore."
fi

# Continue with existing hook if it exists
if [ -f "${hookPath}.backup" ]; then
  exec "${hookPath}.backup" "$@"
fi

exit 0
`

	// Write hook
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to write git hook: %w", err)
	}

	return nil
}

// findGitDir finds the git directory for the current repository
func findGitDir() (string, error) {
	// Start with current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for .git
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return gitDir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			return "", fmt.Errorf("not in a git repository")
		}
		dir = parent
	}
}

// LoadFromEnv loads credentials from environment variables
func (m *CredentialManager) LoadFromEnv() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get all environment variables
	for _, env := range os.Environ() {
		// Parse environment variable
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]

		// Check if this is one of our environment variables
		if !strings.HasPrefix(key, m.envPrefix+"_") {
			continue
		}

		// Extract service name from environment variable
		servicePart := strings.TrimPrefix(key, m.envPrefix+"_")
		servicePart = strings.TrimSuffix(servicePart, "_API_KEY")
		service := strings.ToLower(servicePart)

		// Skip if value is empty
		if value == "" {
			continue
		}

		// Create credential
		cred := &Credential{
			ID:          GenerateCredentialID(service, "env"),
			Name:        fmt.Sprintf("%s API Key (from env)", strings.Title(service)),
			Type:        APIKeyCredential,
			Service:     service,
			Value:       value,
			Description: "API key loaded from environment variable",
			Tags:        []string{"api_key", "env"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Store credential
		if err := m.vault.StoreCredential(cred); err != nil {
			return fmt.Errorf("failed to store credential from environment variable: %w", err)
		}
	}

	return nil
}

// GetAPIKey gets an API key for a provider
func (m *CredentialManager) GetAPIKey(providerType core.ProviderType) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get service name for provider
	service, exists := m.providerToServiceMap[providerType]
	if !exists {
		return "", fmt.Errorf("unknown provider type: %s", providerType)
	}

	// Get credentials for service
	creds, err := m.vault.ListCredentialsByService(service)
	if err != nil {
		return "", fmt.Errorf("failed to list credentials for service %s: %w", service, err)
	}

	// Find API key credential
	for _, cred := range creds {
		if cred.Type == APIKeyCredential {
			return cred.Value, nil
		}
	}

	// Check environment variable as fallback
	envKey := m.envPrefix + "_" + strings.ToUpper(service) + "_API_KEY"
	if apiKey := os.Getenv(envKey); apiKey != "" {
		return apiKey, nil
	}

	return "", fmt.Errorf("no API key found for provider %s", providerType)
}

// SetAPIKey sets an API key for a provider
func (m *CredentialManager) SetAPIKey(providerType core.ProviderType, apiKey string, description string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get service name for provider
	service, exists := m.providerToServiceMap[providerType]
	if !exists {
		return fmt.Errorf("unknown provider type: %s", providerType)
	}

	// Create credential
	cred := &Credential{
		ID:          GenerateCredentialID(service, "api_key"),
		Name:        fmt.Sprintf("%s API Key", strings.Title(service)),
		Type:        APIKeyCredential,
		Service:     service,
		Value:       apiKey,
		Description: description,
		Tags:        []string{"api_key"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		RotationPolicy: &RotationPolicy{
			Enabled:      true,
			IntervalDays: 90,  // Default to 90-day rotation
			WarningDays:  14,  // Warn 14 days before expiration
			LastRotation: time.Now(),
		},
	}

	// Store credential
	return m.vault.StoreCredential(cred)
}

// GetCredential gets a credential by ID
func (m *CredentialManager) GetCredential(id string) (*Credential, error) {
	return m.vault.GetCredential(id)
}

// StoreCredential stores a credential
func (m *CredentialManager) StoreCredential(cred *Credential) error {
	return m.vault.StoreCredential(cred)
}

// DeleteCredential deletes a credential by ID
func (m *CredentialManager) DeleteCredential(id string) error {
	return m.vault.DeleteCredential(id)
}

// ListCredentials lists all credentials
func (m *CredentialManager) ListCredentials() ([]*Credential, error) {
	return m.vault.ListCredentials()
}

// ListCredentialsByService lists credentials for a specific service
func (m *CredentialManager) ListCredentialsByService(service string) ([]*Credential, error) {
	return m.vault.ListCredentialsByService(service)
}

// ListCredentialsByType lists credentials of a specific type
func (m *CredentialManager) ListCredentialsByType(credType CredentialType) ([]*Credential, error) {
	return m.vault.ListCredentialsByType(credType)
}

// ListCredentialsByTag lists credentials with a specific tag
func (m *CredentialManager) ListCredentialsByTag(tag string) ([]*Credential, error) {
	return m.vault.ListCredentialsByTag(tag)
}

// RotateCredential rotates a credential
func (m *CredentialManager) RotateCredential(id string, newValue string) error {
	return m.vault.RotateCredential(id, newValue)
}

// GetCredentialsNeedingRotation returns credentials that need rotation
func (m *CredentialManager) GetCredentialsNeedingRotation() ([]*Credential, error) {
	return m.vault.GetCredentialsNeedingRotation()
}

// Close closes the credential manager
func (m *CredentialManager) Close() error {
	return m.vault.Close()
}

// DefaultManager is the default credential manager
var DefaultManager *CredentialManager

// InitDefaultManager initializes the default credential manager
func InitDefaultManager(options ManagerOptions) error {
	var err error
	DefaultManager, err = NewCredentialManager(options)
	return err
}
