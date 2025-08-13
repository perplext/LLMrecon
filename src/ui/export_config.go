package ui

import (
	"fmt"
	"strings"
)

// ExportConfigurator provides interactive export configuration
type ExportConfigurator struct {
	terminal *Terminal
	preview  *ExportPreview
}

// NewExportConfigurator creates a new export configurator
func NewExportConfigurator(terminal *Terminal) *ExportConfigurator {
	return &ExportConfigurator{
		terminal: terminal,
		preview:  NewExportPreview(terminal),
	}
}

// ConfigureExport interactively configures export settings
func (ec *ExportConfigurator) ConfigureExport(data interface{}) (*ExportConfig, error) {
	ec.terminal.Clear()
	ec.terminal.HeaderBox("Configure Export Settings")
	
	config := &ExportConfig{}
	
	// Step 1: Select format
	format, err := ec.selectFormat(data)
	if err != nil {
		return nil, err
	}
	config.Format = format
	
	// Step 2: Configure output
	if err := ec.configureOutput(config); err != nil {
		return nil, err
	}
	
	// Step 3: Configure format-specific options
	if err := ec.configureFormatOptions(config); err != nil {
		return nil, err
	}
	
	// Step 4: Configure filtering
	if err := ec.configureFiltering(config); err != nil {
		return nil, err
	}
	
	// Step 5: Review and confirm
	if err := ec.reviewConfiguration(config); err != nil {
		return nil, err
	}
	
	return config, nil
}

// Step 1: Format selection with preview
func (ec *ExportConfigurator) selectFormat(data interface{}) (string, error) {
	return ec.preview.ShowFormatSelection(data)
}

// Step 2: Output configuration
func (ec *ExportConfigurator) configureOutput(config *ExportConfig) error {
	ec.terminal.Section("Output Configuration")
	
	// Filename
	defaultName := fmt.Sprintf("scan-report-%s", config.Format)
	filename, err := ec.terminal.Input("Output filename:", defaultName)
	if err != nil {
		return err
	}
	
	// Add extension if missing
	ext := ec.getDefaultExtension(config.Format)
	if ext != "" && !strings.HasSuffix(filename, ext) {
		filename += ext
	}
	config.Filename = filename
	
	// Output directory
	defaultDir := "./reports"
	dir, err := ec.terminal.Input("Output directory:", defaultDir)
	if err != nil {
		return err
	}
	config.OutputPath = filepath.Join(dir, config.Filename)
	
	// Overwrite handling
	overwrite, err := ec.terminal.Select("If file exists:", []string{
		"Ask before overwriting",
		"Always overwrite",
		"Create numbered backup",
		"Append timestamp",
	})
	if err != nil {
		return err
	}
	
	switch overwrite {
	case 0:
		config.OverwriteMode = "ask"
	case 1:
		config.OverwriteMode = "overwrite"
	case 2:
		config.OverwriteMode = "backup"
	case 3:
		config.OverwriteMode = "timestamp"
	}
	
	return nil
}

// Step 3: Format-specific options
func (ec *ExportConfigurator) configureFormatOptions(config *ExportConfig) error {
	ec.terminal.Section("Format Options: " + strings.ToUpper(config.Format))
	
	options := make(map[string]interface{})
	
	switch config.Format {
	case "json":
		options["pretty"] = ec.askYesNo("Pretty print (indented)?", true)
		options["include_metadata"] = ec.askYesNo("Include metadata?", true)
		options["minify"] = ec.askYesNo("Minify output?", false)
		
	case "yaml":
		options["include_comments"] = ec.askYesNo("Include explanatory comments?", true)
		options["flow_style"] = ec.askYesNo("Use flow style for arrays?", false)
		
	case "markdown":
		options["toc"] = ec.askYesNo("Include table of contents?", true)
		options["emoji"] = ec.askYesNo("Use emoji indicators?", true)
		options["github_flavored"] = ec.askYesNo("GitHub flavored markdown?", true)
		
	case "html":
		theme, _ := ec.terminal.Select("Theme:", []string{"Light", "Dark", "Auto"})
		options["theme"] = []string{"light", "dark", "auto"}[theme]
		options["include_charts"] = ec.askYesNo("Include interactive charts?", true)
		options["standalone"] = ec.askYesNo("Single file (embed assets)?", true)
		
	case "pdf":
		size, _ := ec.terminal.Select("Page size:", []string{"A4", "Letter", "Legal"})
		options["page_size"] = []string{"A4", "Letter", "Legal"}[size]
		options["toc"] = ec.askYesNo("Include table of contents?", true)
		options["page_numbers"] = ec.askYesNo("Include page numbers?", true)
		options["watermark"] = ec.askYesNo("Add confidential watermark?", false)
		
	case "csv":
		delimiter, _ := ec.terminal.Input("Field delimiter:", ",")
		options["delimiter"] = delimiter
		options["headers"] = ec.askYesNo("Include headers?", true)
		options["quote_all"] = ec.askYesNo("Quote all fields?", false)
		
	case "sarif":
		options["schema_version"] = "2.1.0"
		options["include_rules"] = ec.askYesNo("Include rule definitions?", true)
		options["include_graphs"] = ec.askYesNo("Include code flow graphs?", false)
		
	case "jira":
		options["project_key"], _ = ec.terminal.Input("JIRA project key:", "SEC")
		options["issue_type"], _ = ec.terminal.Select("Issue type:", 
			[]string{"Bug", "Security Vulnerability", "Task"})
		options["auto_assign"] = ec.askYesNo("Auto-assign issues?", true)
	}
	
	config.FormatOptions = options
	return nil
}

// Step 4: Filtering configuration
func (ec *ExportConfigurator) configureFiltering(config *ExportConfig) error {
	ec.terminal.Section("Data Filtering")
	
	// Severity filter
	includeSeverities := []string{}
	severities := []string{"Critical", "High", "Medium", "Low", "Info"}
	
	ec.terminal.Info("Select severities to include:")
	for _, sev := range severities {
		if ec.askYesNo(fmt.Sprintf("Include %s?", sev), true) {
			includeSeverities = append(includeSeverities, sev)
		}
	}
	config.Filters.Severities = includeSeverities
	
	// Category filter
	includeAll := ec.askYesNo("Include all vulnerability categories?", true)
	if !includeAll {
		categories := []string{
			"Prompt Injection",
			"Data Leakage",
			"Model Manipulation",
			"Output Handling",
			"Access Control",
			"Misconfiguration",
		}
		
		includeCategories := []string{}
		ec.terminal.Info("Select categories to include:")
		for _, cat := range categories {
			if ec.askYesNo(fmt.Sprintf("Include %s?", cat), true) {
				includeCategories = append(includeCategories, cat)
			}
		}
		config.Filters.Categories = includeCategories
	}
	
	// Status filter
	config.Filters.IncludeResolved = ec.askYesNo("Include resolved findings?", false)
	config.Filters.IncludeFalsePositives = ec.askYesNo("Include false positives?", false)
	
	// Data redaction
	config.Redaction.Enabled = ec.askYesNo("Enable sensitive data redaction?", true)
	if config.Redaction.Enabled {
		config.Redaction.RedactPII = ec.askYesNo("Redact PII?", true)
		config.Redaction.RedactSecrets = ec.askYesNo("Redact secrets/tokens?", true)
		config.Redaction.RedactURLs = ec.askYesNo("Redact internal URLs?", false)
	}
	
	return nil
}

// Step 5: Review configuration
func (ec *ExportConfigurator) reviewConfiguration(config *ExportConfig) error {
	ec.terminal.Clear()
	ec.terminal.HeaderBox("Export Configuration Review")
	
	// Display configuration summary
	ec.terminal.Section("Summary")
	
	summary := fmt.Sprintf(`Format: %s
Output: %s
Overwrite: %s

Included Severities: %s
Data Redaction: %s`,
		strings.ToUpper(config.Format),
		config.OutputPath,
		config.OverwriteMode,
		strings.Join(config.Filters.Severities, ", "),
		ec.boolToString(config.Redaction.Enabled, "Enabled", "Disabled"),
	)
	
	ec.terminal.Box("Configuration", summary)
	
	// Format-specific options
	if len(config.FormatOptions) > 0 {
		ec.terminal.Section("Format Options")
		for key, value := range config.FormatOptions {
			ec.terminal.Info(fmt.Sprintf("• %s: %v", 
				ec.humanizeKey(key), value))
		}
	}
	
	// Actions
	ec.terminal.Section("Actions")
	actions := []string{
		"Export with this configuration",
		"Save configuration as template",
		"Modify configuration",
		"Cancel export",
	}
	
	choice, err := ec.terminal.Select("Choose action:", actions)
	if err != nil {
		return err
	}
	
	switch choice {
	case 0:
		// Proceed with export
		return nil
	case 1:
		// Save as template
		return ec.saveConfigTemplate(config)
	case 2:
		// Restart configuration
		return fmt.Errorf("restart configuration")
	case 3:
		// Cancel
		return fmt.Errorf("export cancelled")
	}
	
	return nil
}

// Helper methods

func (ec *ExportConfigurator) askYesNo(question string, defaultYes bool) bool {
	result, _ := ec.terminal.Confirm(question)
	return result
}

func (ec *ExportConfigurator) getDefaultExtension(format string) string {
	extensions := map[string]string{
		"json":     ".json",
		"yaml":     ".yaml",
		"markdown": ".md",
		"html":     ".html",
		"pdf":      ".pdf",
		"csv":      ".csv",
		"sarif":    ".sarif",
		"jira":     "", // No file extension
	}
	return extensions[format]
}

func (ec *ExportConfigurator) humanizeKey(key string) string {
	// Convert snake_case to Title Case
	words := strings.Split(key, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	return strings.Join(words, " ")
}

func (ec *ExportConfigurator) boolToString(value bool, trueStr, falseStr string) string {
	if value {
		return trueStr
	}
	return falseStr
}

// saveConfigTemplate saves the configuration as a reusable template
func (ec *ExportConfigurator) saveConfigTemplate(config *ExportConfig) error {
	ec.terminal.Section("Save Configuration Template")
	
	name, err := ec.terminal.Input("Template name:", "my-export-config")
	if err != nil {
		return err
	}
	
	description, err := ec.terminal.Input("Description:", "")
	if err != nil {
		return err
	}
	
	// Save template (in real implementation)
	ec.terminal.Success(fmt.Sprintf("Template '%s' saved successfully!", name))
	ec.terminal.Info("Use --export-template " + name + " to reuse this configuration")
	
	return nil
}

// LoadTemplate loads a saved configuration template
func (ec *ExportConfigurator) LoadTemplate(name string) (*ExportConfig, error) {
	// In real implementation, load from storage
	// For now, return a sample configuration
	
	if name == "compliance-report" {
		return &ExportConfig{
			Format:     "pdf",
			OutputPath: "./reports/compliance-report.pdf",
			OverwriteMode: "timestamp",
			FormatOptions: map[string]interface{}{
				"page_size":    "A4",
				"toc":          true,
				"page_numbers": true,
				"watermark":    true,
			},
			Filters: ExportFilters{
				Severities:            []string{"Critical", "High"},
				IncludeResolved:       false,
				IncludeFalsePositives: false,
			},
			Redaction: RedactionConfig{
				Enabled:       true,
				RedactPII:     true,
				RedactSecrets: true,
				RedactURLs:    false,
			},
		}, nil
	}
	
	return nil, fmt.Errorf("template not found: %s", name)
}

// QuickExport provides common export presets
func (ec *ExportConfigurator) QuickExport(preset string, data interface{}) (*ExportConfig, error) {
	presets := map[string]*ExportConfig{
		"executive-summary": {
			Format:        "pdf",
			OutputPath:    "./reports/executive-summary.pdf",
			OverwriteMode: "timestamp",
			FormatOptions: map[string]interface{}{
				"page_size": "A4",
				"toc":       false,
				"watermark": false,
			},
			Filters: ExportFilters{
				Severities: []string{"Critical", "High"},
			},
		},
		"technical-details": {
			Format:        "json",
			OutputPath:    "./reports/technical-details.json",
			OverwriteMode: "overwrite",
			FormatOptions: map[string]interface{}{
				"pretty":           true,
				"include_metadata": true,
			},
			Filters: ExportFilters{
				Severities: []string{"Critical", "High", "Medium", "Low"},
			},
		},
		"compliance-audit": {
			Format:        "html",
			OutputPath:    "./reports/compliance-audit.html",
			OverwriteMode: "backup",
			FormatOptions: map[string]interface{}{
				"theme":          "light",
				"include_charts": true,
				"standalone":     true,
			},
		},
		"data-analysis": {
			Format:        "csv",
			OutputPath:    "./reports/vulnerability-data.csv",
			OverwriteMode: "timestamp",
			FormatOptions: map[string]interface{}{
				"delimiter": ",",
				"headers":   true,
			},
		},
	}
	
	if config, ok := presets[preset]; ok {
		// Show preview
		ec.terminal.Info("Using preset: " + preset)
		ec.preview.ShowPreview(config.Format, data)
		
		// Confirm
		if confirmed, _ := ec.terminal.Confirm("Use this preset?"); confirmed {
			return config, nil
		}
	}
	
	return nil, fmt.Errorf("unknown preset: %s", preset)
}

// Data structures

type ExportConfig struct {
	Format        string
	Filename      string
	OutputPath    string
	OverwriteMode string
	FormatOptions map[string]interface{}
	Filters       ExportFilters
	Redaction     RedactionConfig
}

type ExportFilters struct {
	Severities            []string
	Categories            []string
	DateRange             DateRange
	IncludeResolved       bool
	IncludeFalsePositives bool
}

type DateRange struct {
	Start string
	End   string
}

type RedactionConfig struct {
	Enabled       bool
	RedactPII     bool
	RedactSecrets bool
	RedactURLs    bool
	CustomPatterns []string
}