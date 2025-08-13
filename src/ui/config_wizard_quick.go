package ui

import (
	"fmt"
	"os"
	"path/filepath"
)

// Quick setup implementations

func (qs *QuickSetup) basicSetup() error {
	qs.wizard.terminal.Info("Setting up basic configuration...")

	// Select one provider
	providers := []string{"OpenAI", "Anthropic", "Google PaLM", "Local Model"}
	choice, _ := qs.wizard.terminal.Select("Select your LLM provider:", providers)
	
	provider := ProviderConfig{
		Name:    providers[choice],
		Enabled: true,
	}

	// Configure selected provider
	switch choice {
	case 0: // OpenAI
		provider.Type = "openai"
		provider.Endpoint = "https://api.openai.com/v1"
		apiKey, _ := qs.wizard.promptAPIKey("OpenAI")
		provider.APIKey = apiKey
		provider.Model = "gpt-3.5-turbo"
		
	case 1: // Anthropic
		provider.Type = "anthropic"
		provider.Endpoint = "https://api.anthropic.com/v1"
		apiKey, _ := qs.wizard.promptAPIKey("Anthropic")
		provider.APIKey = apiKey
		provider.Model = "claude-3-sonnet-20240229"
		
	case 2: // Google
		provider.Type = "google"
		provider.Endpoint = "https://generativelanguage.googleapis.com/v1"
		apiKey, _ := qs.wizard.promptAPIKey("Google")
		provider.APIKey = apiKey
		provider.Model = "gemini-pro"
		
	case 3: // Local
		provider.Type = "local"
		endpoint, _ := qs.wizard.terminal.Prompt("Local model endpoint (default: http://localhost:8080): ")
		if endpoint == "" {
			endpoint = "http://localhost:8080"
		}
		provider.Endpoint = endpoint
	}

	// Basic settings
	provider.Settings = map[string]interface{}{
		"temperature": 0.7,
		"max_tokens":  2048,
	}

	qs.wizard.config.Providers = []ProviderConfig{provider}

	// Default test settings
	qs.wizard.config.Test = TestConfig{
		ConcurrentTests: 1,
		Timeout:         30,
		MaxRetries:      2,
		RetryDelay:      1,
		RateLimit:       30,
		Categories:      []string{"OWASP LLM Top 10"},
	}

	// Simple output
	qs.wizard.config.Output = OutputConfig{
		Formats:      []string{"json", "markdown"},
		Directory:    "./reports",
		Filename:     "scan-{{.timestamp}}",
		ColorOutput:  true,
		ShowProgress: true,
		Verbose:      false,
	}

	// Basic security
	qs.wizard.config.Security = SecurityConfig{
		APIKeyStorage: "environment",
		EncryptKeys:   false,
		TLSVerify:     true,
	}

	// Minimal advanced settings
	qs.wizard.config.Advanced = AdvancedConfig{
		TemplateDir:     "./templates",
		CacheDir:        "./.cache",
		LogLevel:        "info",
		EnableTelemetry: false,
		AutoUpdate:      true,
	}

	return qs.saveQuickConfig("basic")
}

func (qs *QuickSetup) professionalSetup() error {
	qs.wizard.terminal.Info("Setting up professional configuration...")

	// Multiple providers
	qs.wizard.terminal.Info("Let's configure multiple providers for flexibility.")
	
	// OpenAI
	openai := ProviderConfig{
		Name:     "OpenAI",
		Type:     "openai",
		Enabled:  true,
		Endpoint: "https://api.openai.com/v1",
		Model:    "gpt-4",
		Settings: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  4096,
			"top_p":       0.95,
		},
	}
	
	apiKey, _ := qs.wizard.promptAPIKey("OpenAI")
	openai.APIKey = apiKey
	
	// Anthropic
	anthropic := ProviderConfig{
		Name:     "Anthropic",
		Type:     "anthropic",
		Enabled:  true,
		Endpoint: "https://api.anthropic.com/v1",
		Model:    "claude-3-opus-20240229",
		Settings: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  4096,
			"top_k":       40,
		},
	}
	
	apiKey2, _ := qs.wizard.promptAPIKey("Anthropic")
	anthropic.APIKey = apiKey2

	qs.wizard.config.Providers = []ProviderConfig{openai, anthropic}

	// Professional test settings
	qs.wizard.config.Test = TestConfig{
		ConcurrentTests: 5,
		Timeout:         45,
		MaxRetries:      3,
		RetryDelay:      2,
		RateLimit:       60,
		Categories: []string{
			"OWASP LLM Top 10",
			"Prompt Injection",
			"Data Leakage",
		},
	}

	// Multiple output formats
	qs.wizard.config.Output = OutputConfig{
		Formats:      []string{"json", "markdown", "html", "pdf"},
		Directory:    "./security-reports",
		Filename:     "{{.date}}/scan-{{.scan_id}}-{{.timestamp}}",
		ColorOutput:  true,
		ShowProgress: true,
		Verbose:      true,
	}

	// Enhanced security
	qs.wizard.config.Security = SecurityConfig{
		APIKeyStorage: "encrypted_file",
		EncryptKeys:   true,
		TLSVerify:     true,
		AllowedDomains: []string{
			"api.openai.com",
			"api.anthropic.com",
		},
	}

	// Professional advanced settings
	qs.wizard.config.Advanced = AdvancedConfig{
		TemplateDir:     "./templates",
		CacheDir:        "./.cache",
		LogLevel:        "debug",
		EnableTelemetry: false,
		AutoUpdate:      true,
		Experimental: map[string]interface{}{
			"parallel_providers": true,
		},
	}

	return qs.saveQuickConfig("professional")
}

func (qs *QuickSetup) enterpriseSetup() error {
	qs.wizard.terminal.Info("Setting up enterprise configuration...")

	// Multiple providers with fallback
	providers := []ProviderConfig{
		{
			Name:     "Primary OpenAI",
			Type:     "openai",
			Enabled:  true,
			Endpoint: "https://api.openai.com/v1",
			Model:    "gpt-4-turbo-preview",
			Settings: map[string]interface{}{
				"temperature":        0.3,
				"max_tokens":         8192,
				"top_p":              0.95,
				"frequency_penalty":  0.0,
				"presence_penalty":   0.0,
			},
		},
		{
			Name:     "Fallback Anthropic",
			Type:     "anthropic",
			Enabled:  true,
			Endpoint: "https://api.anthropic.com/v1",
			Model:    "claude-3-opus-20240229",
			Settings: map[string]interface{}{
				"temperature": 0.3,
				"max_tokens":  8192,
				"top_k":       40,
			},
		},
		{
			Name:     "Internal LLM",
			Type:     "custom",
			Enabled:  false,
			Endpoint: "https://llm.internal.company.com/v1",
			Settings: map[string]interface{}{
				"temperature": 0.5,
				"max_tokens":  4096,
			},
		},
	}

	// Get API keys for enabled providers
	for i := range providers {
		if providers[i].Enabled && providers[i].Type != "custom" {
			apiKey, _ := qs.wizard.promptAPIKey(providers[i].Name)
			providers[i].APIKey = apiKey
		}
	}

	qs.wizard.config.Providers = providers

	// Enterprise test settings
	qs.wizard.config.Test = TestConfig{
		ConcurrentTests: 10,
		Timeout:         60,
		MaxRetries:      5,
		RetryDelay:      5,
		RateLimit:       120,
		Categories: []string{
			"OWASP LLM Top 10",
			"Prompt Injection",
			"Data Leakage",
			"Model Manipulation",
			"Content Safety",
		},
		SkipCategories: []string{
			"Performance Testing",
		},
	}

	// Comprehensive output
	qs.wizard.config.Output = OutputConfig{
		Formats:      []string{"json", "markdown", "html", "pdf", "excel", "jsonl"},
		Directory:    "/var/log/llm-security",
		Filename:     "{{.date}}/{{.scan_id}}/report-{{.timestamp}}",
		ColorOutput:  true,
		ShowProgress: true,
		Verbose:      true,
	}

	// Enterprise security
	proxyURL, _ := qs.wizard.terminal.Prompt("Corporate proxy URL (leave empty if none): ")
	
	qs.wizard.config.Security = SecurityConfig{
		APIKeyStorage: "keychain",
		EncryptKeys:   true,
		TLSVerify:     true,
		AllowedDomains: []string{
			"api.openai.com",
			"api.anthropic.com",
			"llm.internal.company.com",
		},
		ProxyURL: proxyURL,
	}

	// Enterprise advanced settings
	qs.wizard.config.Advanced = AdvancedConfig{
		TemplateDir:     "/opt/LLMrecon/templates",
		CacheDir:        "/var/cache/LLMrecon",
		LogLevel:        "info",
		EnableTelemetry: false,
		AutoUpdate:      false, // Manual updates in enterprise
		Experimental: map[string]interface{}{
			"parallel_providers":  true,
			"adaptive_testing":    true,
			"audit_logging":       true,
			"compliance_mode":     true,
			"data_retention_days": 90,
		},
	}

	return qs.saveQuickConfig("enterprise")
}

func (qs *QuickSetup) developmentSetup() error {
	qs.wizard.terminal.Info("Setting up development configuration...")

	// Local and test providers
	qs.wizard.config.Providers = []ProviderConfig{
		{
			Name:     "Local LLaMA",
			Type:     "local",
			Enabled:  true,
			Endpoint: "http://localhost:8080",
			Model:    "llama-2-7b",
			Settings: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  2048,
			},
		},
		{
			Name:     "Test OpenAI",
			Type:     "openai",
			Enabled:  false,
			Endpoint: "https://api.openai.com/v1",
			Model:    "gpt-3.5-turbo",
			APIKey:   os.Getenv("OPENAI_API_KEY"),
			Settings: map[string]interface{}{
				"temperature": 0.9,
				"max_tokens":  1024,
			},
		},
	}

	// Development test settings
	qs.wizard.config.Test = TestConfig{
		ConcurrentTests: 1,
		Timeout:         120, // Longer timeout for debugging
		MaxRetries:      0,   // No retries in dev
		RetryDelay:      0,
		RateLimit:       10, // Low rate for debugging
		Categories:      []string{"Custom Tests"},
	}

	// Development output
	qs.wizard.config.Output = OutputConfig{
		Formats:      []string{"json", "markdown"},
		Directory:    "./dev-reports",
		Filename:     "dev-{{.timestamp}}",
		ColorOutput:  true,
		ShowProgress: true,
		Verbose:      true, // Always verbose in dev
	}

	// Development security (relaxed)
	qs.wizard.config.Security = SecurityConfig{
		APIKeyStorage: "environment",
		EncryptKeys:   false,
		TLSVerify:     false, // Allow self-signed certs
	}

	// Development advanced settings
	qs.wizard.config.Advanced = AdvancedConfig{
		TemplateDir:     "./dev-templates",
		CacheDir:        "./.dev-cache",
		LogLevel:        "debug",
		EnableTelemetry: false,
		AutoUpdate:      false,
		Experimental: map[string]interface{}{
			"debug_mode":         true,
			"save_raw_responses": true,
			"mock_providers":     true,
		},
	}

	return qs.saveQuickConfig("development")
}

func (qs *QuickSetup) cicdSetup() error {
	qs.wizard.terminal.Info("Setting up CI/CD pipeline configuration...")

	// CI/CD providers (using environment variables)
	qs.wizard.config.Providers = []ProviderConfig{
		{
			Name:     "CI OpenAI",
			Type:     "openai",
			Enabled:  true,
			Endpoint: "https://api.openai.com/v1",
			Model:    "gpt-3.5-turbo", // Cost-effective for CI
			APIKey:   "${OPENAI_API_KEY}",
			Settings: map[string]interface{}{
				"temperature": 0.0, // Deterministic
				"max_tokens":  2048,
			},
		},
	}

	// CI/CD test settings
	qs.wizard.config.Test = TestConfig{
		ConcurrentTests: 3,
		Timeout:         30,
		MaxRetries:      1,
		RetryDelay:      1,
		RateLimit:       30,
		Categories: []string{
			"OWASP LLM Top 10",
			"Critical Tests Only",
		},
	}

	// CI/CD output
	qs.wizard.config.Output = OutputConfig{
		Formats: []string{
			"json",     // For parsing
			"junit",    // For CI integration
			"markdown", // For PR comments
		},
		Directory:    "${CI_ARTIFACTS_DIR:-./artifacts}",
		Filename:     "security-scan-${CI_BUILD_ID:-local}",
		ColorOutput:  false, // No color in CI logs
		ShowProgress: false, // Clean CI output
		Verbose:      false,
	}

	// CI/CD security
	qs.wizard.config.Security = SecurityConfig{
		APIKeyStorage: "environment",
		EncryptKeys:   false,
		TLSVerify:     true,
	}

	// CI/CD advanced settings
	qs.wizard.config.Advanced = AdvancedConfig{
		TemplateDir:     "./templates",
		CacheDir:        "${CI_CACHE_DIR:-./.cache}",
		LogLevel:        "warning", // Only important messages
		EnableTelemetry: false,
		AutoUpdate:      false,
		Experimental: map[string]interface{}{
			"fail_fast":          true,
			"exit_on_high_risk":  true,
			"github_annotations": true,
			"gitlab_integration": true,
		},
	}

	return qs.saveQuickConfig("cicd")
}

// Helper methods

func (qs *QuickSetup) promptAPIKey(provider string) (string, error) {
	return qs.wizard.promptAPIKey(provider)
}

func (qs *QuickSetup) saveQuickConfig(preset string) error {
	qs.wizard.terminal.Info("\nConfiguration Summary:")
	qs.wizard.terminal.KeyValue("Preset", preset)
	qs.wizard.terminal.KeyValue("Providers", len(qs.wizard.config.Providers))
	qs.wizard.terminal.KeyValue("Output Formats", len(qs.wizard.config.Output.Formats))

	// Determine save location based on preset
	var configPath string
	switch preset {
	case "cicd":
		configPath = "./LLMrecon-ci.yaml"
	case "development":
		configPath = "./LLMrecon-dev.yaml"
	default:
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".LLMrecon")
		os.MkdirAll(configDir, 0755)
		configPath = filepath.Join(configDir, "config.yaml")
	}

	// Save configuration
	if err := qs.wizard.saveConfig(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	qs.wizard.terminal.Success("Configuration saved to: %s", configPath)

	// Show next steps
	qs.wizard.terminal.Section("Next Steps")
	
	switch preset {
	case "basic":
		qs.wizard.terminal.List([]string{
			"Run 'LLMrecon scan --help' to see scanning options",
			"Download templates with 'LLMrecon template update'",
			"Start your first scan with 'LLMrecon scan <target>'",
		}, true)
		
	case "professional":
		qs.wizard.terminal.List([]string{
			"Review and customize templates in ./templates",
			"Set up scheduled scans for continuous monitoring",
			"Configure webhook notifications for findings",
			"Explore advanced scanning options",
		}, true)
		
	case "enterprise":
		qs.wizard.terminal.List([]string{
			"Configure LDAP/SSO integration for team access",
			"Set up audit logging and compliance reporting",
			"Deploy scanning infrastructure",
			"Review security policies and procedures",
			"Schedule training for security team",
		}, true)
		
	case "development":
		qs.wizard.terminal.List([]string{
			"Create custom templates in ./dev-templates",
			"Enable debug mode for detailed logging",
			"Use mock providers for testing",
			"Contribute improvements back to the project",
		}, true)
		
	case "cicd":
		qs.wizard.terminal.List([]string{
			"Add security scanning to your CI pipeline",
			"Configure failure thresholds",
			"Set up notifications for high-risk findings",
			"Create custom templates for your use cases",
		}, true)
	}

	return nil
}