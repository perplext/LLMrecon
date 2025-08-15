package ui

import (
	"fmt"
	"strings"
)

// CommandSuggester provides intelligent command suggestions
type CommandSuggester struct {
	history      *CommandHistory
	context      *CommandContext
	terminal     *Terminal
	suggestions  map[string][]Suggestion

// CommandHistory tracks command usage
type CommandHistory struct {
	commands   []HistoryEntry
	maxEntries int

// HistoryEntry represents a command in history
type HistoryEntry struct {
	Command   string
	Timestamp time.Time
	Success   bool
	Duration  time.Duration

// CommandContext tracks current command context
type CommandContext struct {
	LastCommand    string
	LastError      error
	WorkingDir     string
	ActiveProvider string
	LoadedTemplate string
	CurrentScan    string

// Suggestion represents a command suggestion
type Suggestion struct {
	Command     string
	Description string
	Confidence  float64
	Reason      string

// NewCommandSuggester creates a new command suggester
func NewCommandSuggester(terminal *Terminal) *CommandSuggester {
	cs := &CommandSuggester{
		terminal: terminal,
		history: &CommandHistory{
			commands:   make([]HistoryEntry, 0),
			maxEntries: 100,
		},
		context:     &CommandContext{},
		suggestions: make(map[string][]Suggestion),
	}
	
	cs.initializeSuggestions()
	return cs

// initializeSuggestions sets up suggestion patterns
func (cs *CommandSuggester) initializeSuggestions() {
	// Error-based suggestions
	cs.suggestions["missing_api_key"] = []Suggestion{
		{
			Command:     "config set provider.{provider}.api_key YOUR_API_KEY",
			Description: "Set the API key for the provider",
			Confidence:  0.95,
			Reason:      "Missing API key error detected",
		},
		{
			Command:     "export {PROVIDER}_API_KEY=YOUR_API_KEY",
			Description: "Set API key via environment variable",
			Confidence:  0.85,
			Reason:      "Alternative: use environment variable",
		},
	}

	cs.suggestions["template_not_found"] = []Suggestion{
		{
			Command:     "template list",
			Description: "List available templates",
			Confidence:  0.9,
			Reason:      "See what templates are available",
		},
		{
			Command:     "template get {template_name}",
			Description: "Download the missing template",
			Confidence:  0.95,
			Reason:      "Template not found locally",
		},
	}

	cs.suggestions["connection_failed"] = []Suggestion{
		{
			Command:     "provider test {provider}",
			Description: "Test provider connection",
			Confidence:  0.9,
			Reason:      "Connection issue detected",
		},
		{
			Command:     "config get provider.{provider}.endpoint",
			Description: "Check provider endpoint configuration",
			Confidence:  0.85,
			Reason:      "Verify endpoint is correct",
		},
	}

	// Workflow suggestions
	cs.suggestions["after_config_init"] = []Suggestion{
		{
			Command:     "template get owasp-llm-top10",
			Description: "Download OWASP LLM Top 10 templates",
			Confidence:  0.95,
			Reason:      "Common next step after configuration",
		},
		{
			Command:     "provider test all",
			Description: "Test all configured providers",
			Confidence:  0.9,
			Reason:      "Verify provider configuration",
		},
	}

	cs.suggestions["after_scan_complete"] = []Suggestion{
		{
			Command:     "report generate {scan_id} --format pdf",
			Description: "Generate PDF report for the scan",
			Confidence:  0.9,
			Reason:      "Create shareable report",
		},
		{
			Command:     "scan --template {next_template}",
			Description: "Run additional security tests",
			Confidence:  0.85,
			Reason:      "Continue security assessment",
		},
	}

	cs.suggestions["first_run"] = []Suggestion{
		{
			Command:     "config init",
			Description: "Run interactive configuration wizard",
			Confidence:  0.95,
			Reason:      "No configuration detected",
		},
		{
			Command:     "help",
			Description: "Show available commands",
			Confidence:  0.8,
			Reason:      "Learn about the tool",
		},
	}

// RecordCommand records a command execution
func (cs *CommandSuggester) RecordCommand(command string, success bool, duration time.Duration) {
	entry := HistoryEntry{
		Command:   command,
		Timestamp: time.Now(),
		Success:   success,
		Duration:  duration,
	}
	
	cs.history.commands = append(cs.history.commands, entry)
	
	// Maintain history size
	if len(cs.history.commands) > cs.history.maxEntries {
		cs.history.commands = cs.history.commands[1:]
	}

// UpdateContext updates command context
func (cs *CommandSuggester) UpdateContext(key string, value interface{}) {
	switch key {
	case "last_command":
		cs.context.LastCommand = value.(string)
	case "last_error":
		cs.context.LastError = value.(error)
	case "working_dir":
		cs.context.WorkingDir = value.(string)
	case "active_provider":
		cs.context.ActiveProvider = value.(string)
	case "loaded_template":
		cs.context.LoadedTemplate = value.(string)
	case "current_scan":
		cs.context.CurrentScan = value.(string)
	}

// GetSuggestions returns relevant suggestions based on context
func (cs *CommandSuggester) GetSuggestions() []Suggestion {
	suggestions := make([]Suggestion, 0)
	
	// Check for error-based suggestions
	if cs.context.LastError != nil {
		errorStr := cs.context.LastError.Error()
		
		if strings.Contains(errorStr, "API key") || strings.Contains(errorStr, "unauthorized") {
			suggestions = append(suggestions, cs.processsSuggestions("missing_api_key")...)
		} else if strings.Contains(errorStr, "template not found") {
			suggestions = append(suggestions, cs.processsSuggestions("template_not_found")...)
		} else if strings.Contains(errorStr, "connection") || strings.Contains(errorStr, "timeout") {
			suggestions = append(suggestions, cs.processsSuggestions("connection_failed")...)
		}
	}
	
	// Check for workflow suggestions
	if cs.context.LastCommand != "" {
		if strings.Contains(cs.context.LastCommand, "config init") {
			suggestions = append(suggestions, cs.processsSuggestions("after_config_init")...)
		} else if strings.Contains(cs.context.LastCommand, "scan") && cs.context.CurrentScan != "" {
			suggestions = append(suggestions, cs.processsSuggestions("after_scan_complete")...)
		}
	}
	
	// First run suggestions
	if len(cs.history.commands) == 0 {
		suggestions = append(suggestions, cs.processsSuggestions("first_run")...)
	}
	
	// Get contextual suggestions
	suggestions = append(suggestions, cs.getContextualSuggestions()...)
	
	// Sort by confidence
	sortSuggestionsByConfidence(suggestions)
	
	// Limit to top suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	
	return suggestions

// ShowSuggestions displays suggestions to the user
func (cs *CommandSuggester) ShowSuggestions() {
	suggestions := cs.GetSuggestions()
	
	if len(suggestions) == 0 {
		return
	}
	
	cs.terminal.Section("Suggested Commands")
	
	for i, suggestion := range suggestions {
		// Format command with placeholders highlighted
		command := cs.highlightPlaceholders(suggestion.Command)
		
		cs.terminal.Print("%d. %s", i+1, command)
		cs.terminal.Print("   %s", cs.terminal.formatter.Muted(suggestion.Description))
		
		if suggestion.Reason != "" {
			cs.terminal.Print("   %s %s", 
				cs.terminal.formatter.Info("→"),
				cs.terminal.formatter.Muted(suggestion.Reason),
			)
		}
		
		fmt.Println()
	}
	
	// Quick select option
	cs.terminal.Info("Type a number to run the suggested command, or press Enter to continue")

// getContextualSuggestions returns suggestions based on current context
func (cs *CommandSuggester) getContextualSuggestions() []Suggestion {
	suggestions := make([]Suggestion, 0)
	
	// If a template is loaded but no scan running
	if cs.context.LoadedTemplate != "" && cs.context.CurrentScan == "" {
		suggestions = append(suggestions, Suggestion{
			Command:     fmt.Sprintf("scan --template %s", cs.context.LoadedTemplate),
			Description: "Run scan with loaded template",
			Confidence:  0.8,
			Reason:      "Template loaded but not used",
		})
	}
	
	// If provider is set but not tested
	if cs.context.ActiveProvider != "" && !cs.hasRecentProviderTest() {
		suggestions = append(suggestions, Suggestion{
			Command:     fmt.Sprintf("provider test %s", cs.context.ActiveProvider),
			Description: "Test provider connection",
			Confidence:  0.7,
			Reason:      "Provider not recently tested",
		})
	}
	
	// Suggest frequently used commands
	frequent := cs.getFrequentCommands()
	for _, cmd := range frequent {
		suggestions = append(suggestions, Suggestion{
			Command:     cmd,
			Description: "Frequently used command",
			Confidence:  0.6,
			Reason:      "Based on your history",
		})
	}
	
	return suggestions

// processsSuggestions processes suggestion templates
func (cs *CommandSuggester) processsSuggestions(key string) []Suggestion {
	templates, exists := cs.suggestions[key]
	if !exists {
		return nil
	}
	
	processed := make([]Suggestion, 0, len(templates))
	
	for _, template := range templates {
		suggestion := template
		
		// Replace placeholders with context values
		suggestion.Command = cs.replacePlaceholders(suggestion.Command)
		suggestion.Description = cs.replacePlaceholders(suggestion.Description)
		
		processed = append(processed, suggestion)
	}
	
	return processed

// replacePlaceholders replaces placeholders with actual values
func (cs *CommandSuggester) replacePlaceholders(text string) string {
	replacements := map[string]string{
		"{provider}":      cs.context.ActiveProvider,
		"{PROVIDER}":      strings.ToUpper(cs.context.ActiveProvider),
		"{template_name}": cs.context.LoadedTemplate,
		"{scan_id}":       cs.context.CurrentScan,
		"{working_dir}":   cs.context.WorkingDir,
	}
	
	// Handle next template suggestion
	if strings.Contains(text, "{next_template}") {
		nextTemplate := cs.suggestNextTemplate()
		replacements["{next_template}"] = nextTemplate
	}
	
	result := text
	for placeholder, value := range replacements {
		if value != "" {
			result = strings.ReplaceAll(result, placeholder, value)
		}
	}
	
	return result

// highlightPlaceholders highlights placeholders in command
func (cs *CommandSuggester) highlightPlaceholders(command string) string {
	// Find placeholders
	placeholders := []string{
		"YOUR_API_KEY",
		"{provider}",
		"{template_name}",
		"{scan_id}",
	}
	
	result := command
	for _, placeholder := range placeholders {
		if strings.Contains(result, placeholder) {
			highlighted := cs.terminal.formatter.Highlight(placeholder)
			result = strings.ReplaceAll(result, placeholder, highlighted)
		}
	}
	
	return result

// Helper methods

func (cs *CommandSuggester) hasRecentProviderTest() bool {
	cutoff := time.Now().Add(-24 * time.Hour)
	
	for i := len(cs.history.commands) - 1; i >= 0; i-- {
		cmd := cs.history.commands[i]
		if cmd.Timestamp.Before(cutoff) {
			break
		}
		
		if strings.Contains(cmd.Command, "provider test") && cmd.Success {
			return true
		}
	}
	
	return false

func (cs *CommandSuggester) getFrequentCommands() []string {
	// Count command frequency
	counts := make(map[string]int)
	
	for _, entry := range cs.history.commands {
		// Extract base command (first word)
		parts := strings.Fields(entry.Command)
		if len(parts) > 0 {
			baseCmd := parts[0]
			counts[baseCmd]++
		}
	}
	
	// Find top commands
	type cmdCount struct {
		command string
		count   int
	}
	
	var sorted []cmdCount
	for cmd, count := range counts {
		sorted = append(sorted, cmdCount{cmd, count})
	}
	
	// Sort by count
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	// Return top 3
	result := make([]string, 0, 3)
	for i := 0; i < 3 && i < len(sorted); i++ {
		result = append(result, sorted[i].command)
	}
	
	return result

func (cs *CommandSuggester) suggestNextTemplate() string {
	// Suggest templates based on what's been run
	allTemplates := []string{
		"prompt-injection",
		"data-leakage",
		"model-manipulation",
		"content-safety",
		"dos-attacks",
	}
	
	// Find templates not yet run
	for _, template := range allTemplates {
		found := false
		for _, entry := range cs.history.commands {
			if strings.Contains(entry.Command, template) {
				found = true
				break
			}
		}
		
		if !found {
			return template
		}
	}
	
	return "advanced-tests"

func sortSuggestionsByConfidence(suggestions []Suggestion) {
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Confidence > suggestions[i].Confidence {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

// InteractiveSuggestions provides interactive command selection
type InteractiveSuggestions struct {
	suggester *CommandSuggester
	terminal  *Terminal

// NewInteractiveSuggestions creates interactive suggestions
func NewInteractiveSuggestions(suggester *CommandSuggester, terminal *Terminal) *InteractiveSuggestions {
	return &InteractiveSuggestions{
		suggester: suggester,
		terminal:  terminal,
	}

// ShowAndSelect shows suggestions and allows selection
func (is *InteractiveSuggestions) ShowAndSelect() (string, bool) {
	suggestions := is.suggester.GetSuggestions()
	
	if len(suggestions) == 0 {
		return "", false
	}
	
	is.terminal.Section("Command Suggestions")
	
	// Build options
	options := make([]string, len(suggestions))
	for i, suggestion := range suggestions {
		options[i] = fmt.Sprintf("%s - %s", suggestion.Command, suggestion.Description)
	}
	
	// Add skip option
	options = append(options, "Skip suggestions")
	
	choice, err := is.terminal.Select("Select a command to run:", options)
	if err != nil || choice == len(suggestions) {
		return "", false
	}
	
	selectedCmd := suggestions[choice].Command
	
	// Check if command has placeholders
	if strings.Contains(selectedCmd, "YOUR_") || strings.Contains(selectedCmd, "{") {
		is.terminal.Warning("This command contains placeholders that need to be filled:")
		is.terminal.Code(selectedCmd)
		
		// Interactive placeholder filling
		selectedCmd = is.fillPlaceholders(selectedCmd)
	}
	
	// Confirm execution
	is.terminal.Print("\nCommand to execute:")
	is.terminal.Code(selectedCmd)
	
	confirm, _ := is.terminal.Confirm("Run this command?", true)
	
	return selectedCmd, confirm

// fillPlaceholders interactively fills command placeholders
func (is *InteractiveSuggestions) fillPlaceholders(command string) string {
	result := command
	
	// Find and replace YOUR_API_KEY
	if strings.Contains(result, "YOUR_API_KEY") {
		apiKey, _ := is.terminal.Prompt("Enter API key: ")
		result = strings.ReplaceAll(result, "YOUR_API_KEY", apiKey)
	}
	
	// Find and replace other placeholders
	placeholders := []struct {
		placeholder string
		prompt      string
	}{
		{"{provider}", "Enter provider name: "},
		{"{template_name}", "Enter template name: "},
		{"{scan_id}", "Enter scan ID: "},
		{"{format}", "Enter output format: "},
	}
	
	for _, p := range placeholders {
		if strings.Contains(result, p.placeholder) {
			value, _ := is.terminal.Prompt(p.prompt)
			result = strings.ReplaceAll(result, p.placeholder, value)
		}
	}
	
