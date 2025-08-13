// Package config provides configuration management for the LLMreconing Tool
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// UpdateSources defines the URLs for update sources
	UpdateSources struct {
		// GitHub is the URL for the official GitHub repository for updates
		GitHub string `mapstructure:"github"`
		// GitLab is the URL for an optional internal GitLab repository for updates
		GitLab string `mapstructure:"gitlab"`
	} `mapstructure:"update_sources"`

	// APIKeys stores API keys for different services
	APIKeys struct {
		// OpenAI API key for OpenAI provider
		OpenAI string `mapstructure:"openai"`
		// Anthropic API key for Anthropic provider
		Anthropic string `mapstructure:"anthropic"`
	} `mapstructure:"api_keys"`

	// Templates configuration
	Templates struct {
		// Directory where templates are stored
		Dir string `mapstructure:"dir"`
		// Default template categories to use
		DefaultCategories []string `mapstructure:"default_categories"`
	} `mapstructure:"templates"`

	// Modules configuration
	Modules struct {
		// Directory where modules are stored
		Dir string `mapstructure:"dir"`
		// Enabled modules
		Enabled []string `mapstructure:"enabled"`
	} `mapstructure:"modules"`

	// Security settings
	Security struct {
		// Whether to verify signatures of updates
		VerifySignatures bool `mapstructure:"verify_signatures"`
		// Public key for verifying signatures
		PublicKey string `mapstructure:"public_key"`
	} `mapstructure:"security"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	cfg := &Config{}

	// Set default update sources
	cfg.UpdateSources.GitHub = "https://api.github.com/repos/LLMrecon/LLMrecon/releases/latest"
	cfg.UpdateSources.GitLab = ""

	// Set default template and module directories
	homeDir, err := os.UserHomeDir()
	if err == nil {
		cfg.Templates.Dir = filepath.Join(homeDir, ".LLMrecon", "templates")
		cfg.Modules.Dir = filepath.Join(homeDir, ".LLMrecon", "modules")
	} else {
		cfg.Templates.Dir = "./templates"
		cfg.Modules.Dir = "./modules"
	}

	// Set default template categories
	cfg.Templates.DefaultCategories = []string{"prompt-injection", "data-leakage", "insecure-output"}

	// Set default security settings
	cfg.Security.VerifySignatures = true
	cfg.Security.PublicKey = ""

	return cfg
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig() (*Config, error) {
	// Start with default configuration
	cfg := DefaultConfig()

	// Set up viper
	v := viper.New()
	v.SetConfigName(".LLMrecon")
	v.SetConfigType("yaml")

	// Add config paths
	homeDir, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(homeDir)
	}
	v.AddConfigPath(".")

	// Set environment variable prefix
	v.SetEnvPrefix("LLMRT")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional, so we'll just use defaults if it's not found
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables for sensitive values
	if apiKey := os.Getenv("LLMRT_OPENAI_API_KEY"); apiKey != "" {
		cfg.APIKeys.OpenAI = apiKey
	}
	if apiKey := os.Getenv("LLMRT_ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.APIKeys.Anthropic = apiKey
	}

	return cfg, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(cfg *Config) error {
	v := viper.New()
	v.SetConfigName(".LLMrecon")
	v.SetConfigType("yaml")

	// Set config values from struct
	v.Set("update_sources.github", cfg.UpdateSources.GitHub)
	v.Set("update_sources.gitlab", cfg.UpdateSources.GitLab)
	v.Set("templates.dir", cfg.Templates.Dir)
	v.Set("templates.default_categories", cfg.Templates.DefaultCategories)
	v.Set("modules.dir", cfg.Modules.Dir)
	v.Set("modules.enabled", cfg.Modules.Enabled)
	v.Set("security.verify_signatures", cfg.Security.VerifySignatures)
	v.Set("security.public_key", cfg.Security.PublicKey)

	// Don't save API keys to config file for security
	// They should be stored in environment variables

	// Save config to user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".LLMrecon.yaml")
	return v.WriteConfigAs(configPath)
}
