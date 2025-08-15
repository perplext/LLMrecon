package ui

import (
	"os"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ConfigWizard provides interactive configuration setup
type ConfigWizard struct {
	terminal *Terminal
	config   *Configuration
	steps    []WizardStep
}

// Configuration represents the tool configuration
type Configuration struct {
	// Provider settings
	Providers []ProviderConfig `yaml:"providers"`
	
	// Test settings
	Test TestConfig `yaml:"test"`
	
	// Output settings
	Output OutputConfig `yaml:"output"`
	
	// Security settings
	Security SecurityConfig `yaml:"security"`
	
	// Advanced settings
	Advanced AdvancedConfig `yaml:"advanced"`

// ProviderConfig represents LLM provider configuration
type ProviderConfig struct {
	Name     string            `yaml:"name"`
	Type     string            `yaml:"type"`
	Enabled  bool              `yaml:"enabled"`
	APIKey   string            `yaml:"api_key,omitempty"`
	Endpoint string            `yaml:"endpoint,omitempty"`
	Model    string            `yaml:"model,omitempty"`
	Settings map[string]interface{} `yaml:"settings,omitempty"`
}

// TestConfig represents test execution configuration
type TestConfig struct {
	ConcurrentTests int      `yaml:"concurrent_tests"`
	Timeout         int      `yaml:"timeout"`
	MaxRetries      int      `yaml:"max_retries"`
	RetryDelay      int      `yaml:"retry_delay"`
	RateLimit       int      `yaml:"rate_limit"`
	Categories      []string `yaml:"default_categories,omitempty"`
	SkipCategories  []string `yaml:"skip_categories,omitempty"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Formats      []string `yaml:"formats"`
	Directory    string   `yaml:"directory"`
	Filename     string   `yaml:"filename_pattern"`
	Verbose      bool     `yaml:"verbose"`
	ColorOutput  bool     `yaml:"color_output"`
	ShowProgress bool     `yaml:"show_progress"`

// SecurityConfig represents security settings
type SecurityConfig struct {
	APIKeyStorage    string   `yaml:"api_key_storage"`
	EncryptKeys      bool     `yaml:"encrypt_keys"`
	TLSVerify        bool     `yaml:"tls_verify"`
	AllowedDomains   []string `yaml:"allowed_domains,omitempty"`
	ProxyURL         string   `yaml:"proxy_url,omitempty"`
}

// AdvancedConfig represents advanced settings
type AdvancedConfig struct {
	TemplateDir      string   `yaml:"template_dir"`
	CacheDir         string   `yaml:"cache_dir"`
	LogLevel         string   `yaml:"log_level"`
	EnableTelemetry  bool     `yaml:"enable_telemetry"`
	AutoUpdate       bool     `yaml:"auto_update"`
	Experimental     map[string]interface{} `yaml:"experimental,omitempty"`

// WizardStep represents a configuration step
type WizardStep struct {
	Name        string
	Description string
	Configure   func() error
	Skip        func() bool

// NewConfigWizard creates a new configuration wizard
func NewConfigWizard(terminal *Terminal) *ConfigWizard {
	wizard := &ConfigWizard{
		terminal: terminal,
		config: &Configuration{
			Providers: []ProviderConfig{},
			Test: TestConfig{
				ConcurrentTests: 5,
				Timeout:        30,
				MaxRetries:     3,
				RetryDelay:     1,
				RateLimit:      60,
			},
			Output: OutputConfig{
				Formats:      []string{"json", "markdown"},
				Directory:    "./reports",
				Filename:     "scan-{{.timestamp}}-{{.scan_id}}",
				ColorOutput:  true,
				ShowProgress: true,
			},
			Security: SecurityConfig{
				APIKeyStorage: "encrypted_file",
				EncryptKeys:   true,
				TLSVerify:     true,
			},
			Advanced: AdvancedConfig{
				TemplateDir:     "./templates",
				CacheDir:        "./.cache",
				LogLevel:        "info",
				EnableTelemetry: false,
				AutoUpdate:      true,
			},
		},
	}

	// Define wizard steps
	wizard.steps = []WizardStep{
		{
			Name:        "Welcome",
			Description: "Introduction to the configuration wizard",
			Configure:   wizard.showWelcome,
		},
		{
			Name:        "Providers",
			Description: "Configure LLM providers",
			Configure:   wizard.configureProviders,
		},
		{
			Name:        "Test Settings",
			Description: "Configure test execution parameters",
			Configure:   wizard.configureTestSettings,
		},
		{
			Name:        "Output Settings",
			Description: "Configure output formats and options",
			Configure:   wizard.configureOutput,
		},
		{
			Name:        "Security",
			Description: "Configure security settings",
			Configure:   wizard.configureSecurity,
		},
		{
			Name:        "Advanced",
			Description: "Configure advanced options",
			Configure:   wizard.configureAdvanced,
			Skip: func() bool {
				// Skip if user chooses basic setup
				skip, _ := wizard.terminal.Confirm("Skip advanced configuration?", true)
				return skip
			},
		},
		{
			Name:        "Review",
			Description: "Review and save configuration",
			Configure:   wizard.reviewAndSave,
		},
	}

	return wizard

// Run executes the configuration wizard
func (w *ConfigWizard) Run() error {
	// Check for existing configuration
	if w.checkExistingConfig() {
		useExisting, _ := w.terminal.Confirm("An existing configuration was found. Would you like to modify it?", true)
		if useExisting {
			if err := w.loadExistingConfig(); err != nil {
				w.terminal.Error("Failed to load existing configuration: %v", err)
			}
		}
	}

	// Execute wizard steps
	totalSteps := len(w.steps)
	for i, step := range w.steps {
		// Show progress
		w.terminal.Header(fmt.Sprintf("Configuration Wizard - Step %d/%d: %s", i+1, totalSteps, step.Name))
		
		if step.Description != "" {
			w.terminal.Info(step.Description)
			fmt.Println()
		}

		// Check if step should be skipped
		if step.Skip != nil && step.Skip() {
			w.terminal.Info("Skipping %s...", step.Name)
			continue
		}

		// Execute step
		if err := step.Configure(); err != nil {
			w.terminal.Error("Error in %s: %v", step.Name, err)
			
			retry, _ := w.terminal.Confirm("Would you like to retry this step?", true)
			if retry {
				i-- // Retry current step
				continue
			}
			
			return err
		}

		fmt.Println()
	}

	w.terminal.Success("Configuration wizard completed successfully!")
	return nil

// Step implementations

func (w *ConfigWizard) showWelcome() error {
	w.terminal.Print("Welcome to the LLMrecon configuration wizard!")
	w.terminal.Print("This wizard will help you set up the tool for first use.")
	w.terminal.Print("")
	w.terminal.Print("The wizard will guide you through:")
	w.terminal.List([]string{
		"Configuring LLM providers (OpenAI, Anthropic, etc.)",
		"Setting up test execution parameters",
		"Choosing output formats and locations",
		"Configuring security settings",
		"Advanced options (optional)",
	}, false)
	
	w.terminal.Print("")
	ready, _ := w.terminal.Confirm("Ready to begin?", true)
	if !ready {
		return fmt.Errorf("wizard cancelled by user")
	}
	
	return nil

func (w *ConfigWizard) configureProviders() error {
	w.terminal.Info("Let's configure your LLM providers.")
	w.terminal.Print("You can configure multiple providers and switch between them during testing.\n")

	// Available providers
	availableProviders := []struct {
		Name        string
		Type        string
		NeedsAPIKey bool
		DefaultURL  string
	}{
		{"OpenAI", "openai", true, "https://api.openai.com/v1"},
		{"Anthropic", "anthropic", true, "https://api.anthropic.com/v1"},
		{"Google PaLM", "google", true, "https://generativelanguage.googleapis.com/v1"},
		{"Cohere", "cohere", true, "https://api.cohere.ai/v1"},
		{"Hugging Face", "huggingface", true, "https://api-inference.huggingface.co"},
		{"Local Model", "local", false, "https://localhost:8080"},
		{"Custom API", "custom", false, ""},
	}

	// Provider selection
	providerNames := make([]string, len(availableProviders))
	for i, p := range availableProviders {
		providerNames[i] = p.Name
	}

	selected, err := w.terminal.MultiSelect("Select providers to configure:", providerNames)
	if err != nil {
		return err
	}

	// Configure each selected provider
	for _, idx := range selected {
		provider := availableProviders[idx]
		w.terminal.Subheader(fmt.Sprintf("Configuring %s", provider.Name))

		config := ProviderConfig{
			Name:    provider.Name,
			Type:    provider.Type,
			Enabled: true,
		}

		// API Key
		if provider.NeedsAPIKey {
			apiKey, err := w.promptAPIKey(provider.Name)
			if err != nil {
				continue
			}
			config.APIKey = apiKey
		}

		// Endpoint
		defaultEndpoint := provider.DefaultURL
		if defaultEndpoint != "" {
			w.terminal.Info("Default endpoint: %s", defaultEndpoint)
			customEndpoint, _ := w.terminal.Confirm("Use custom endpoint?", false)
			if customEndpoint {
				endpoint, _ := w.terminal.Prompt("Enter endpoint URL: ")
				if endpoint != "" {
					config.Endpoint = endpoint
				} else {
					config.Endpoint = defaultEndpoint
				}
			} else {
				config.Endpoint = defaultEndpoint
			}
		} else {
			endpoint, _ := w.terminal.Prompt("Enter endpoint URL: ")
			config.Endpoint = endpoint
		}

		// Model selection
		models := w.getAvailableModels(provider.Type)
		if len(models) > 0 {
			modelChoice, _ := w.terminal.Select("Select default model:", models)
			config.Model = models[modelChoice]
		}

		// Provider-specific settings
		config.Settings = w.getProviderSettings(provider.Type)

		w.config.Providers = append(w.config.Providers, config)
		w.terminal.Success("✓ %s configured", provider.Name)
	}

	if len(w.config.Providers) == 0 {
		w.terminal.Warning("No providers configured. You'll need at least one provider to run tests.")
	}

	return nil

func (w *ConfigWizard) configureTestSettings() error {
	w.terminal.Info("Now let's configure test execution settings.")

	// Concurrent tests
	concurrent, _ := w.terminal.Prompt(fmt.Sprintf("Number of concurrent tests (default: %d): ", w.config.Test.ConcurrentTests))
	if concurrent != "" {
		fmt.Sscanf(concurrent, "%d", &w.config.Test.ConcurrentTests)
	}

	// Timeout
	timeout, _ := w.terminal.Prompt(fmt.Sprintf("Test timeout in seconds (default: %d): ", w.config.Test.Timeout))
	if timeout != "" {
		fmt.Sscanf(timeout, "%d", &w.config.Test.Timeout)
	}

	// Retries
	enableRetries, _ := w.terminal.Confirm("Enable automatic retries for failed tests?", true)
	if enableRetries {
		maxRetries, _ := w.terminal.Prompt(fmt.Sprintf("Maximum retries (default: %d): ", w.config.Test.MaxRetries))
		if maxRetries != "" {
			fmt.Sscanf(maxRetries, "%d", &w.config.Test.MaxRetries)
		}

		retryDelay, _ := w.terminal.Prompt(fmt.Sprintf("Retry delay in seconds (default: %d): ", w.config.Test.RetryDelay))
		if retryDelay != "" {
			fmt.Sscanf(retryDelay, "%d", &w.config.Test.RetryDelay)
		}
	} else {
		w.config.Test.MaxRetries = 0
	}

	// Rate limiting
	rateLimit, _ := w.terminal.Prompt(fmt.Sprintf("API rate limit per minute (default: %d): ", w.config.Test.RateLimit))
	if rateLimit != "" {
		fmt.Sscanf(rateLimit, "%d", &w.config.Test.RateLimit)
	}

	// Default categories
	categories := []string{
		"OWASP LLM Top 10",
		"Prompt Injection",
		"Data Leakage",
		"Model Manipulation",
		"Content Safety",
		"Performance Testing",
		"Custom Tests",
	}

	w.terminal.Info("\nSelect default test categories to run:")
	selected, _ := w.terminal.MultiSelect("Categories:", categories)
	
	w.config.Test.Categories = make([]string, len(selected))
	for i, idx := range selected {
		w.config.Test.Categories[i] = categories[idx]
	}

	return nil

func (w *ConfigWizard) configureOutput() error {
	w.terminal.Info("Configure output settings for test results.")

	// Output formats
	formats := []string{
		"JSON - Machine-readable format",
		"Markdown - Human-readable documentation",
		"HTML - Web-viewable reports",
		"PDF - Printable reports",
		"CSV - Spreadsheet format",
		"Excel - Full spreadsheet with formatting",
		"JSONL - Line-delimited JSON for streaming",
	}

	selected, _ := w.terminal.MultiSelect("Select output formats:", formats)
	
	w.config.Output.Formats = make([]string, 0)
	for _, idx := range selected {
		switch idx {
		case 0:
			w.config.Output.Formats = append(w.config.Output.Formats, "json")
		case 1:
			w.config.Output.Formats = append(w.config.Output.Formats, "markdown")
		case 2:
			w.config.Output.Formats = append(w.config.Output.Formats, "html")
		case 3:
			w.config.Output.Formats = append(w.config.Output.Formats, "pdf")
		case 4:
			w.config.Output.Formats = append(w.config.Output.Formats, "csv")
		case 5:
			w.config.Output.Formats = append(w.config.Output.Formats, "excel")
		case 6:
			w.config.Output.Formats = append(w.config.Output.Formats, "jsonl")
		}
	}

	// Output directory
	outputDir, _ := w.terminal.Prompt(fmt.Sprintf("Output directory (default: %s): ", w.config.Output.Directory))
	if outputDir != "" {
		w.config.Output.Directory = outputDir
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(w.config.Output.Directory, 0700); err != nil {
		w.terminal.Warning("Failed to create output directory: %v", err)
	}
	// Filename pattern
	w.terminal.Info("\nFilename pattern can use variables: {{.timestamp}}, {{.scan_id}}, {{.date}}")
	pattern, _ := w.terminal.Prompt(fmt.Sprintf("Filename pattern (default: %s): ", w.config.Output.Filename))
	if pattern != "" {
		w.config.Output.Filename = pattern
	}

	// Display options
	w.config.Output.Verbose, if err := w.terminal.Confirm("Enable verbose output?", false); err != nil { return err }
	w.config.Output.ColorOutput, if err := w.terminal.Confirm("Enable colored terminal output?", true); err != nil { return err }
	w.config.Output.ShowProgress, if err := w.terminal.Confirm("Show progress indicators?", true); err != nil { return err }

	return nil

func (w *ConfigWizard) configureSecurity() error {
	w.terminal.Info("Configure security settings.")

	// API key storage
	storageOptions := []string{
		"Encrypted file - Store encrypted keys in local file",
		"Environment variables - Read from environment",
		"System keychain - Use OS keychain/credential manager",
		"Plain text file - Not recommended",
	}

	choice, _ := w.terminal.Select("How should API keys be stored?", storageOptions)
	
	switch choice {
	case 0:
		w.config.Security.APIKeyStorage = "encrypted_file"
		w.config.Security.EncryptKeys = true
	case 1:
		w.config.Security.APIKeyStorage = "environment"
		w.config.Security.EncryptKeys = false
	case 2:
		w.config.Security.APIKeyStorage = "keychain"
		w.config.Security.EncryptKeys = false
	case 3:
		w.config.Security.APIKeyStorage = "plain_file"
		w.config.Security.EncryptKeys = false
		w.terminal.Warning("Storing API keys in plain text is not recommended!")
	}

	// TLS verification
	w.config.Security.TLSVerify, if err := w.terminal.Confirm("Verify TLS certificates?", true); err != nil { return err }

	// Domain restrictions
	restrictDomains, _ := w.terminal.Confirm("Restrict API calls to specific domains?", false)
	if restrictDomains {
		domains, _ := w.terminal.Prompt("Enter allowed domains (comma-separated): ")
		if domains != "" {
			w.config.Security.AllowedDomains = strings.Split(domains, ",")
			for i := range w.config.Security.AllowedDomains {
				w.config.Security.AllowedDomains[i] = strings.TrimSpace(w.config.Security.AllowedDomains[i])
			}
		}
	}

	// Proxy settings
	useProxy, _ := w.terminal.Confirm("Use HTTP proxy?", false)
	if useProxy {
		proxyURL, _ := w.terminal.Prompt("Proxy URL (e.g., https://proxy.company.com:8080): ")
		w.config.Security.ProxyURL = proxyURL
	}

	return nil

func (w *ConfigWizard) configureAdvanced() error {
	w.terminal.Info("Configure advanced options.")

	// Template directory
	templateDir, _ := w.terminal.Prompt(fmt.Sprintf("Template directory (default: %s): ", w.config.Advanced.TemplateDir))
	if templateDir != "" {
		w.config.Advanced.TemplateDir = templateDir
	}
	// Cache directory
	cacheDir, _ := w.terminal.Prompt(fmt.Sprintf("Cache directory (default: %s): ", w.config.Advanced.CacheDir))
	if cacheDir != "" {
		w.config.Advanced.CacheDir = cacheDir
	}

	// Log level
	logLevels := []string{"debug", "info", "warning", "error"}
	levelChoice, _ := w.terminal.Select("Log level:", logLevels)
	w.config.Advanced.LogLevel = logLevels[levelChoice]

	// Telemetry
	w.config.Advanced.EnableTelemetry, if err := w.terminal.Confirm("Enable anonymous usage telemetry?", false); err != nil { return err }
	if w.config.Advanced.EnableTelemetry {
		w.terminal.Info("Telemetry helps improve the tool. No sensitive data is collected.")
	}

	// Auto-update
	w.config.Advanced.AutoUpdate, if err := w.terminal.Confirm("Enable automatic updates?", true); err != nil { return err }

	// Experimental features
	enableExperimental, _ := w.terminal.Confirm("Enable experimental features?", false)
	if enableExperimental {
		w.config.Advanced.Experimental = map[string]interface{}{
			"parallel_providers": true,
			"adaptive_testing":   true,
			"ml_analysis":        false,
		}
		w.terminal.Warning("Experimental features may be unstable!")
	}

	return nil

func (w *ConfigWizard) reviewAndSave() error {
	w.terminal.Info("Configuration Review")
	
	// Show configuration summary
	w.terminal.Section("Providers")
	for _, provider := range w.config.Providers {
		status := "Configured"
		if provider.APIKey == "" && provider.Type != "local" {
			status = "Missing API Key"
		}
		w.terminal.KeyValue(provider.Name, status)
	}

	w.terminal.Section("Test Settings")
	w.terminal.KeyValue("Concurrent Tests", w.config.Test.ConcurrentTests)
	w.terminal.KeyValue("Timeout", fmt.Sprintf("%d seconds", w.config.Test.Timeout))
	w.terminal.KeyValue("Max Retries", w.config.Test.MaxRetries)
	w.terminal.KeyValue("Rate Limit", fmt.Sprintf("%d/minute", w.config.Test.RateLimit))

	w.terminal.Section("Output Settings")
	w.terminal.KeyValue("Formats", strings.Join(w.config.Output.Formats, ", "))
	w.terminal.KeyValue("Directory", w.config.Output.Directory)
	w.terminal.KeyValue("Color Output", w.config.Output.ColorOutput)

	w.terminal.Section("Security")
	w.terminal.KeyValue("API Key Storage", w.config.Security.APIKeyStorage)
	w.terminal.KeyValue("TLS Verify", w.config.Security.TLSVerify)

	// Save options
	fmt.Println()
	save, _ := w.terminal.Confirm("Save this configuration?", true)
	if !save {
		return fmt.Errorf("configuration not saved")
	}

	// Choose save location
	locations := []string{
		"Default location (~/.LLMrecon/config.yaml)",
		"Current directory (./config.yaml)",
		"Custom location",
	}

	choice, _ := w.terminal.Select("Where to save configuration?", locations)
	
	var configPath string
	switch choice {
	case 0:
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".LLMrecon")
		os.MkdirAll(configDir, 0700)
		configPath = filepath.Join(configDir, "config.yaml")
	case 1:
		configPath = "./config.yaml"
	case 2:
		configPath, if err := w.terminal.Prompt("Enter path: "); err != nil { return err }
	}

	// Save configuration
	if err := w.saveConfig(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	w.terminal.Success("Configuration saved to: %s", configPath)
	
	// Set as default
	setDefault, _ := w.terminal.Confirm("Set this as the default configuration?", true)
	if setDefault {
		viper.SetConfigFile(configPath)
	}

	return nil

// Helper methods

func (w *ConfigWizard) checkExistingConfig() bool {
	// Check common configuration locations
	locations := []string{
		"./config.yaml",
		"./.LLMrecon.yaml",
	}
	
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		locations = append(locations,
			filepath.Join(homeDir, ".LLMrecon", "config.yaml"),
			filepath.Join(homeDir, ".config", "LLMrecon", "config.yaml"),
		)
	}
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return true
		}
	}

	return false

func (w *ConfigWizard) loadExistingConfig() error {
	// Implementation would load existing configuration
	w.terminal.Info("Loading existing configuration...")
	return nil

func (w *ConfigWizard) promptAPIKey(provider string) (string, error) {
	// Check environment variable first
	envVar := fmt.Sprintf("%s_API_KEY", strings.ToUpper(strings.ReplaceAll(provider, " ", "_")))
	if envKey := os.Getenv(envVar); envKey != "" {
		useEnv, _ := w.terminal.Confirm(fmt.Sprintf("Found API key in environment variable %s. Use it?", envVar), true)
		if useEnv {
			return envKey, nil
		}
	}

	// Prompt for key
	apiKey, err := w.terminal.Prompt(fmt.Sprintf("Enter API key for %s: ", provider))
	if err != nil {
		return "", err
	}

	// Validate format
	if !w.validateAPIKey(provider, apiKey) {
		w.terminal.Warning("API key format appears invalid")
		retry, _ := w.terminal.Confirm("Try again?", true)
		if retry {
			return w.promptAPIKey(provider)
		}
	}

	return apiKey, nil

func (w *ConfigWizard) validateAPIKey(provider, key string) bool {
	if key == "" {
		return false
	}

	// Basic format validation
	patterns := map[string]string{
		"openai":     `^sk-[a-zA-Z0-9]{48}$`,
		"anthropic":  `^sk-ant-[a-zA-Z0-9-]{40,}$`,
		"cohere":     `^[a-zA-Z0-9]{40}$`,
		"huggingface": `^hf_[a-zA-Z0-9]{30,}$`,
	}

	if pattern, exists := patterns[strings.ToLower(provider)]; exists {
		matched, _ := regexp.MatchString(pattern, key)
		return matched
	}

	// Default: just check length
	return len(key) >= 20

func (w *ConfigWizard) getAvailableModels(providerType string) []string {
	models := map[string][]string{
		"openai": {
			"gpt-4-turbo-preview",
			"gpt-4",
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-16k",
		},
		"anthropic": {
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
			"claude-2.1",
			"claude-instant-1.2",
		},
		"google": {
			"gemini-pro",
			"gemini-pro-vision",
			"palm-2",
		},
		"cohere": {
			"command",
			"command-light",
			"command-nightly",
		},
	}

	if m, exists := models[providerType]; exists {
		return m
	}

	return []string{}

func (w *ConfigWizard) getProviderSettings(providerType string) map[string]interface{} {
	settings := make(map[string]interface{})

	// Common settings
	temperature, _ := w.terminal.Prompt("Temperature (0.0-1.0, default: 0.7): ")
	if temperature != "" {
		var temp float64
		fmt.Sscanf(temperature, "%f", &temp)
		settings["temperature"] = temp
	}

	maxTokens, _ := w.terminal.Prompt("Max tokens (default: 2048): ")
	if maxTokens != "" {
		var tokens int
		fmt.Sscanf(maxTokens, "%d", &tokens)
		settings["max_tokens"] = tokens
	}
	// Provider-specific settings
	switch providerType {
	case "openai":
		topP, _ := w.terminal.Prompt("Top P (default: 1.0): ")
		if topP != "" {
			var p float64
			fmt.Sscanf(topP, "%f", &p)
			settings["top_p"] = p
		}
		
		settings["presence_penalty"] = 0.0
		settings["frequency_penalty"] = 0.0
		
	case "anthropic":
		topK, _ := w.terminal.Prompt("Top K (default: 40): ")
		if topK != "" {
			var k int
			fmt.Sscanf(topK, "%d", &k)
			settings["top_k"] = k
		}
	}

	return settings

func (w *ConfigWizard) saveConfig(path string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(w.config)
	if err != nil {
		return err
	}

	// Write file
	return os.WriteFile(filepath.Clean(path, data, 0600))

// QuickSetup provides a streamlined setup for common scenarios
type QuickSetup struct {
	wizard *ConfigWizard
}

// NewQuickSetup creates a quick setup helper
func NewQuickSetup(wizard *ConfigWizard) *QuickSetup {
	return &QuickSetup{wizard: wizard}

// Run executes quick setup
func (qs *QuickSetup) Run() error {
	scenarios := []string{
		"Basic - Single provider, default settings",
		"Professional - Multiple providers, custom output",
		"Enterprise - Full configuration with security",
		"Development - Local testing setup",
		"CI/CD - Automated pipeline configuration",
	}

	choice, err := qs.wizard.terminal.Select("Select setup scenario:", scenarios)
	if err != nil {
		return err
	}

	switch choice {
	case 0:
		return qs.basicSetup()
	case 1:
		return qs.professionalSetup()
	case 2:
		return qs.enterpriseSetup()
	case 3:
		return qs.developmentSetup()
	case 4:
		return qs.cicdSetup()
	}

}
}
}
}
}
}
}
}
}
}
}
}
}
