package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// AutoCompleter provides command auto-completion and suggestions
type AutoCompleter struct {
	commands     map[string]*CommandInfo
	aliases      map[string]string
	history      []string
	maxHistory   int
	suggestions  SuggestionEngine
}

// CommandInfo contains command metadata
type CommandInfo struct {
	Name        string
	Description string
	Usage       string
	Aliases     []string
	Flags       []FlagInfo
	SubCommands []*CommandInfo
	Examples    []Example
}

// FlagInfo contains flag metadata
type FlagInfo struct {
	Name        string
	Shorthand   string
	Description string
	Type        string
	Default     string
	Required    bool
	Values      []string // Possible values for enum flags
}

// Example contains command example
type Example struct {
	Command     string
	Description string
}

// NewAutoCompleter creates a new auto-completer
func NewAutoCompleter() *AutoCompleter {
	ac := &AutoCompleter{
		commands:    make(map[string]*CommandInfo),
		aliases:     make(map[string]string),
		history:     make([]string, 0),
		maxHistory:  1000,
		suggestions: NewSuggestionEngine(),
	}
	
	// Initialize with default commands
	ac.initializeCommands()
	
	return ac
}

// initializeCommands sets up command information
func (ac *AutoCompleter) initializeCommands() {
	// Main commands
	ac.RegisterCommand(&CommandInfo{
		Name:        "scan",
		Description: "Run security scans against LLM endpoints",
		Usage:       "scan [target] [flags]",
		Aliases:     []string{"s", "test"},
		Flags: []FlagInfo{
			{Name: "template", Shorthand: "t", Description: "Template file or directory", Type: "string"},
			{Name: "output", Shorthand: "o", Description: "Output format", Type: "string", Values: []string{"json", "markdown", "html", "pdf"}},
			{Name: "config", Shorthand: "c", Description: "Configuration file", Type: "string"},
			{Name: "provider", Shorthand: "p", Description: "LLM provider", Type: "string", Values: []string{"openai", "anthropic", "google", "local"}},
			{Name: "concurrent", Description: "Number of concurrent tests", Type: "int", Default: "5"},
			{Name: "timeout", Description: "Test timeout in seconds", Type: "int", Default: "30"},
			{Name: "verbose", Shorthand: "v", Description: "Verbose output", Type: "bool"},
		},
		Examples: []Example{
			{Command: "scan https://api.openai.com/v1/chat/completions", Description: "Scan OpenAI endpoint"},
			{Command: "scan -t owasp-llm -o pdf report.pdf", Description: "Run OWASP tests and save as PDF"},
			{Command: "scan --provider anthropic --template prompt-injection", Description: "Test Anthropic with specific template"},
		},
	})

	ac.RegisterCommand(&CommandInfo{
		Name:        "template",
		Description: "Manage test templates",
		Usage:       "template [command]",
		Aliases:     []string{"t", "tmpl"},
		SubCommands: []*CommandInfo{
			{
				Name:        "list",
				Description: "List available templates",
				Usage:       "template list [flags]",
				Flags: []FlagInfo{
					{Name: "category", Description: "Filter by category", Type: "string"},
					{Name: "severity", Description: "Filter by severity", Type: "string", Values: []string{"critical", "high", "medium", "low", "info"}},
				},
			},
			{
				Name:        "get",
				Description: "Download templates from repository",
				Usage:       "template get [name]",
			},
			{
				Name:        "validate",
				Description: "Validate template syntax",
				Usage:       "template validate [file]",
			},
			{
				Name:        "create",
				Description: "Create new template interactively",
				Usage:       "template create",
			},
		},
	})

	ac.RegisterCommand(&CommandInfo{
		Name:        "config",
		Description: "Manage configuration",
		Usage:       "config [command]",
		Aliases:     []string{"cfg"},
		SubCommands: []*CommandInfo{
			{
				Name:        "init",
				Description: "Initialize configuration interactively",
				Usage:       "config init",
			},
			{
				Name:        "set",
				Description: "Set configuration value",
				Usage:       "config set [key] [value]",
				Examples: []Example{
					{Command: "config set provider.openai.api_key sk-...", Description: "Set OpenAI API key"},
					{Command: "config set test.concurrent 10", Description: "Set concurrent test limit"},
				},
			},
			{
				Name:        "get",
				Description: "Get configuration value",
				Usage:       "config get [key]",
			},
			{
				Name:        "list",
				Description: "List all configuration",
				Usage:       "config list",
			},
		},
	})

	ac.RegisterCommand(&CommandInfo{
		Name:        "report",
		Description: "Generate and manage reports",
		Usage:       "report [command]",
		Aliases:     []string{"r"},
		SubCommands: []*CommandInfo{
			{
				Name:        "generate",
				Description: "Generate report from scan results",
				Usage:       "report generate [scan-id]",
				Flags: []FlagInfo{
					{Name: "format", Shorthand: "f", Description: "Output format", Type: "string", Values: []string{"pdf", "html", "excel"}},
					{Name: "template", Description: "Report template", Type: "string"},
				},
			},
			{
				Name:        "list",
				Description: "List recent reports",
				Usage:       "report list",
			},
			{
				Name:        "export",
				Description: "Export report in different format",
				Usage:       "report export [report-id] [format]",
			},
		},
	})

	ac.RegisterCommand(&CommandInfo{
		Name:        "provider",
		Description: "Manage LLM providers",
		Usage:       "provider [command]",
		SubCommands: []*CommandInfo{
			{
				Name:        "list",
				Description: "List configured providers",
				Usage:       "provider list",
			},
			{
				Name:        "add",
				Description: "Add new provider",
				Usage:       "provider add [name]",
			},
			{
				Name:        "test",
				Description: "Test provider connection",
				Usage:       "provider test [name]",
			},
			{
				Name:        "remove",
				Description: "Remove provider",
				Usage:       "provider remove [name]",
			},
		},
	})
}

// RegisterCommand registers a command for auto-completion
func (ac *AutoCompleter) RegisterCommand(cmd *CommandInfo) {
	ac.commands[cmd.Name] = cmd
	
	// Register aliases
	for _, alias := range cmd.Aliases {
		ac.aliases[alias] = cmd.Name
	}
	
	// Register sub-commands recursively
	for _, sub := range cmd.SubCommands {
		fullName := cmd.Name + " " + sub.Name
		ac.commands[fullName] = sub
	}
}

// Complete returns completions for the given input
func (ac *AutoCompleter) Complete(input string) []string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ac.getRootCommands()
	}

	// Check if it's a known command or alias
	cmdName := parts[0]
	if alias, exists := ac.aliases[cmdName]; exists {
		cmdName = alias
	}

	cmd, exists := ac.commands[cmdName]
	if !exists {
		// Partial command match
		return ac.findCommandMatches(parts[0])
	}

	// Complete sub-commands
	if len(parts) == 1 && !strings.HasSuffix(input, " ") {
		// Still typing the command
		return ac.findCommandMatches(parts[0])
	}

	// Check for sub-commands
	if len(parts) > 1 && len(cmd.SubCommands) > 0 {
		subInput := strings.Join(parts[1:], " ")
		return ac.completeSubCommand(cmd, subInput)
	}

	// Complete flags
	lastPart := parts[len(parts)-1]
	if strings.HasPrefix(lastPart, "-") {
		return ac.completeFlagsForCommand(cmd, lastPart)
	}

	// Check if previous part was a flag that needs a value
	if len(parts) >= 2 {
		prevPart := parts[len(parts)-2]
		if strings.HasPrefix(prevPart, "-") {
			return ac.completeValueForFlag(cmd, prevPart)
		}
	}

	// Suggest common patterns
	return ac.suggestions.GetSuggestions(cmd.Name, input)
}

// CompleteFlags returns flag completions for cobra command
func (ac *AutoCompleter) CompleteFlags(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cmdInfo := ac.findCommandInfo(cmd.Name())
	if cmdInfo == nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	completions := ac.completeFlagsForCommand(cmdInfo, toComplete)
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// CompleteArgs returns argument completions
func (ac *AutoCompleter) CompleteArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cmdName := cmd.Name()
	
	switch cmdName {
	case "scan":
		if len(args) == 0 {
			// Complete target URLs
			return ac.completeTargets(toComplete), cobra.ShellCompDirectiveNoFileComp
		}
		
	case "template":
		if len(args) > 0 {
			switch args[0] {
			case "get", "validate":
				return ac.completeTemplates(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
		}
		
	case "provider":
		if len(args) > 0 {
			switch args[0] {
			case "test", "remove":
				return ac.completeProviders(toComplete), cobra.ShellCompDirectiveNoFileComp
			}
		}
		
	case "report":
		if len(args) > 0 {
			switch args[0] {
			case "generate":
				return ac.completeScanIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
			case "export":
				if len(args) == 1 {
					return ac.completeReportIDs(toComplete), cobra.ShellCompDirectiveNoFileComp
				} else if len(args) == 2 {
					return []string{"pdf", "html", "excel", "json", "markdown"}, cobra.ShellCompDirectiveNoFileComp
				}
			}
		}
	}

	return nil, cobra.ShellCompDirectiveDefault
}

// Helper methods

func (ac *AutoCompleter) getRootCommands() []string {
	commands := make([]string, 0, len(ac.commands))
	for name, cmd := range ac.commands {
		if !strings.Contains(name, " ") { // Root commands only
			commands = append(commands, name)
		}
	}
	sort.Strings(commands)
	return commands
}

func (ac *AutoCompleter) findCommandMatches(prefix string) []string {
	matches := make([]string, 0)
	
	// Check exact commands
	for name := range ac.commands {
		if strings.HasPrefix(name, prefix) && !strings.Contains(name, " ") {
			matches = append(matches, name)
		}
	}
	
	// Check aliases
	for alias, cmdName := range ac.aliases {
		if strings.HasPrefix(alias, prefix) {
			matches = append(matches, alias)
		}
	}
	
	sort.Strings(matches)
	return matches
}

func (ac *AutoCompleter) completeSubCommand(parent *CommandInfo, input string) []string {
	completions := make([]string, 0)
	
	parts := strings.Fields(input)
	if len(parts) == 0 || (len(parts) == 1 && !strings.HasSuffix(input, " ")) {
		// Complete sub-command names
		prefix := ""
		if len(parts) == 1 {
			prefix = parts[0]
		}
		
		for _, sub := range parent.SubCommands {
			if strings.HasPrefix(sub.Name, prefix) {
				completions = append(completions, parent.Name+" "+sub.Name)
			}
		}
	}
	
	return completions
}

func (ac *AutoCompleter) completeFlagsForCommand(cmd *CommandInfo, prefix string) []string {
	completions := make([]string, 0)
	
	for _, flag := range cmd.Flags {
		// Long form
		longFlag := "--" + flag.Name
		if strings.HasPrefix(longFlag, prefix) {
			completion := longFlag
			if flag.Type != "bool" {
				completion += "="
			}
			completions = append(completions, completion)
		}
		
		// Short form
		if flag.Shorthand != "" {
			shortFlag := "-" + flag.Shorthand
			if strings.HasPrefix(shortFlag, prefix) {
				completions = append(completions, shortFlag)
			}
		}
	}
	
	sort.Strings(completions)
	return completions
}

func (ac *AutoCompleter) completeValueForFlag(cmd *CommandInfo, flagName string) []string {
	// Remove flag prefix
	flagName = strings.TrimPrefix(flagName, "--")
	flagName = strings.TrimPrefix(flagName, "-")
	
	for _, flag := range cmd.Flags {
		if flag.Name == flagName || flag.Shorthand == flagName {
			if len(flag.Values) > 0 {
				return flag.Values
			}
			
			// Special handling for common flag types
			switch flag.Type {
			case "bool":
				return []string{"true", "false"}
			case "string":
				if strings.Contains(flag.Name, "file") || strings.Contains(flag.Name, "path") {
					// File completion would be handled by shell
					return nil
				}
			}
		}
	}
	
	return nil
}

func (ac *AutoCompleter) findCommandInfo(name string) *CommandInfo {
	if cmd, exists := ac.commands[name]; exists {
		return cmd
	}
	
	// Check aliases
	if alias, exists := ac.aliases[name]; exists {
		return ac.commands[alias]
	}
	
	return nil
}

// Completion data providers

func (ac *AutoCompleter) completeTargets(prefix string) []string {
	// Common LLM API endpoints
	targets := []string{
		"https://api.openai.com/v1/chat/completions",
		"https://api.anthropic.com/v1/messages",
		"https://generativelanguage.googleapis.com/v1/models",
		"http://localhost:8080/v1/completions",
	}
	
	completions := make([]string, 0)
	for _, target := range targets {
		if strings.HasPrefix(target, prefix) {
			completions = append(completions, target)
		}
	}
	
	return completions
}

func (ac *AutoCompleter) completeTemplates(prefix string) []string {
	// Would load actual templates from filesystem
	templates := []string{
		"owasp-llm-top10",
		"prompt-injection",
		"data-leakage",
		"model-manipulation",
		"content-safety",
	}
	
	completions := make([]string, 0)
	for _, tmpl := range templates {
		if strings.HasPrefix(tmpl, prefix) {
			completions = append(completions, tmpl)
		}
	}
	
	return completions
}

func (ac *AutoCompleter) completeProviders(prefix string) []string {
	// Would load from config
	providers := []string{
		"openai",
		"anthropic",
		"google",
		"cohere",
		"local",
	}
	
	completions := make([]string, 0)
	for _, provider := range providers {
		if strings.HasPrefix(provider, prefix) {
			completions = append(completions, provider)
		}
	}
	
	return completions
}

func (ac *AutoCompleter) completeScanIDs(prefix string) []string {
	// Would load from scan history
	return []string{
		"scan-2024-01-20-001",
		"scan-2024-01-20-002",
		"scan-2024-01-19-015",
	}
}

func (ac *AutoCompleter) completeReportIDs(prefix string) []string {
	// Would load from report history
	return []string{
		"report-2024-01-20-001",
		"report-2024-01-20-002",
	}
}

// SuggestionEngine provides intelligent command suggestions
type SuggestionEngine struct {
	patterns map[string][]string
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine() SuggestionEngine {
	se := SuggestionEngine{
		patterns: make(map[string][]string),
	}
	
	// Initialize common patterns
	se.patterns["scan"] = []string{
		"--template owasp-llm",
		"--output json",
		"--provider openai",
		"--concurrent 10",
		"--timeout 60",
	}
	
	se.patterns["template"] = []string{
		"list --category prompt-injection",
		"get owasp-llm-top10",
		"validate ./my-template.yaml",
		"create",
	}
	
	se.patterns["config"] = []string{
		"init",
		"set provider.openai.api_key",
		"get test.concurrent",
		"list",
	}
	
	return se
}

// GetSuggestions returns suggestions based on context
func (se SuggestionEngine) GetSuggestions(command, input string) []string {
	if patterns, exists := se.patterns[command]; exists {
		suggestions := make([]string, 0)
		for _, pattern := range patterns {
			fullSuggestion := command + " " + pattern
			if strings.HasPrefix(fullSuggestion, input) {
				suggestions = append(suggestions, fullSuggestion)
			}
		}
		return suggestions
	}
	
	return nil
}

// BashCompletionScript generates bash completion script
func GenerateBashCompletionScript(rootCmd *cobra.Command) string {
	var b strings.Builder
	
	b.WriteString("#!/bin/bash\n\n")
	b.WriteString("# LLMrecon bash completion script\n")
	b.WriteString("# Generated by LLMrecon completion bash\n\n")
	
	rootCmd.GenBashCompletion(&b)
	
	return b.String()
}

// ZshCompletionScript generates zsh completion script
func GenerateZshCompletionScript(rootCmd *cobra.Command) string {
	var b strings.Builder
	
	b.WriteString("#compdef LLMrecon\n\n")
	b.WriteString("# LLMrecon zsh completion script\n")
	b.WriteString("# Generated by LLMrecon completion zsh\n\n")
	
	rootCmd.GenZshCompletion(&b)
	
	return b.String()
}

// InstallCompletions installs shell completions
func InstallCompletions(shell string, rootCmd *cobra.Command) error {
	var script string
	var installPath string
	
	switch shell {
	case "bash":
		script = GenerateBashCompletionScript(rootCmd)
		installPath = "/etc/bash_completion.d/LLMrecon"
		
	case "zsh":
		script = GenerateZshCompletionScript(rootCmd)
		// Get zsh completion directory
		homeDir, _ := os.UserHomeDir()
		installPath = filepath.Join(homeDir, ".zsh/completions/_LLMrecon")
		
	case "fish":
		var b strings.Builder
		rootCmd.GenFishCompletion(&b, true)
		script = b.String()
		homeDir, _ := os.UserHomeDir()
		installPath = filepath.Join(homeDir, ".config/fish/completions/LLMrecon.fish")
		
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
	
	// Create directory if needed
	dir := filepath.Dir(installPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write completion file
	if err := os.WriteFile(installPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write completion file: %w", err)
	}
	
	fmt.Printf("Completions installed to: %s\n", installPath)
	fmt.Println("Restart your shell or source the completion file to enable completions.")
	
	return nil
}