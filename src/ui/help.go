package ui

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// HelpSystem provides context-aware help and documentation
type HelpSystem struct {
	terminal    *Terminal
	suggester   *CommandSuggester
	examples    map[string][]Example
	tips        map[string][]string
	faqs        map[string][]FAQ
	context     *HelpContext
}

// HelpContext represents the current help context
type HelpContext struct {
	Command       string
	Subcommand    string
	ErrorContext  string
	LastCommand   string
	UserLevel     string // beginner, intermediate, expert
	PreferredLang string
}

// Example represents a command example
type Example struct {
	Command     string
	Description string
	Output      string
	Tags        []string
}

// FAQ represents a frequently asked question
type FAQ struct {
	Question string
	Answer   string
	Related  []string
}

// NewHelpSystem creates a new help system
func NewHelpSystem(terminal *Terminal, suggester *CommandSuggester) *HelpSystem {
	hs := &HelpSystem{
		terminal:  terminal,
		suggester: suggester,
		examples:  make(map[string][]Example),
		tips:      make(map[string][]string),
		faqs:      make(map[string][]FAQ),
		context:   &HelpContext{UserLevel: "intermediate"},
	}
	
	hs.initializeContent()
	return hs
}

// initializeContent sets up the help content
func (hs *HelpSystem) initializeContent() {
	// Command examples
	hs.examples["scan"] = []Example{
		{
			Command:     "LLMrecon scan --template prompt-injection --target api.example.com",
			Description: "Run a prompt injection scan against an API",
			Output:      "Scanning api.example.com with prompt-injection template...\n✓ 15 test cases executed\n⚠ 3 vulnerabilities found",
			Tags:        []string{"security", "api", "prompt-injection"},
		},
		{
			Command:     "LLMrecon scan --owasp-category LLM01 --severity high",
			Description: "Scan for high-severity OWASP LLM01 vulnerabilities",
			Output:      "Loading templates for OWASP LLM01 (high severity)...\n✓ 8 templates loaded\n✓ Starting scan...",
			Tags:        []string{"owasp", "compliance", "severity"},
		},
	}
	
	hs.examples["template"] = []Example{
		{
			Command:     "LLMrecon template list --category prompt-injection",
			Description: "List all prompt injection templates",
			Output:      "Available prompt-injection templates:\n- basic-injection-v1.0\n- indirect-injection-v2.0\n- jailbreak-attempts-v1.5",
			Tags:        []string{"template", "list", "discovery"},
		},
		{
			Command:     "LLMrecon template validate my-template.yaml",
			Description: "Validate a custom template",
			Output:      "Validating my-template.yaml...\n✓ Schema validation passed\n✓ Syntax check passed\n✓ Template is valid",
			Tags:        []string{"template", "validation", "development"},
		},
	}
	
	// Tips
	hs.tips["scan"] = []string{
		"Use --parallel to run scans concurrently and reduce execution time",
		"Add --report-format html to generate visual scan reports",
		"Use --dry-run to preview what tests will be executed",
		"Combine multiple templates with --template template1,template2",
		"Set --timeout to prevent long-running tests from blocking",
	}
	
	hs.tips["template"] = []string{
		"Templates can inherit from base templates to reduce duplication",
		"Use variables in templates for dynamic test generation",
		"Tag templates for easier discovery and filtering",
		"Version your templates to track changes over time",
		"Test templates locally before deploying to production",
	}
	
	// FAQs
	hs.faqs["general"] = []FAQ{
		{
			Question: "How do I get started with LLMrecon?",
			Answer:   "Run 'LLMrecon init' to create a configuration file, then use 'LLMrecon scan --help' to see scanning options.",
			Related:  []string{"init", "config", "scan"},
		},
		{
			Question: "What's the difference between templates and test cases?",
			Answer:   "Templates define the structure and parameters for tests. Test cases are specific instances generated from templates with actual payloads.",
			Related:  []string{"template", "test", "payload"},
		},
	}
	
	hs.faqs["security"] = []FAQ{
		{
			Question: "How are sensitive results protected?",
			Answer:   "Scan results are encrypted at rest, access is logged, and sensitive data is redacted in reports by default.",
			Related:  []string{"encryption", "audit", "privacy"},
		},
		{
			Question: "Can I scan production systems?",
			Answer:   "Yes, but use --safe-mode to limit request rates and --read-only to prevent any state changes.",
			Related:  []string{"production", "safety", "rate-limiting"},
		},
	}
}

// ShowHelp displays context-aware help
func (hs *HelpSystem) ShowHelp(cmd *cobra.Command, args []string) error {
	// Update context
	hs.updateContext(cmd, args)
	
	// Clear screen for better readability
	hs.terminal.Clear()
	
	// Show command-specific help
	hs.terminal.HeaderBox("Help: " + cmd.Name())
	
	// Basic usage
	hs.showUsage(cmd)
	
	// Context-aware content
	if hs.context.ErrorContext != "" {
		hs.showErrorHelp()
	}
	
	// Examples
	hs.showExamples(cmd.Name())
	
	// Tips
	hs.showTips(cmd.Name())
	
	// Related commands
	hs.showRelatedCommands(cmd)
	
	// Quick actions
	hs.showQuickActions(cmd)
	
	return nil
}

// ShowInteractiveHelp provides an interactive help experience
func (hs *HelpSystem) ShowInteractiveHelp() error {
	for {
		hs.terminal.Clear()
		hs.terminal.HeaderBox("Interactive Help System")
		
		options := []string{
			"Browse Commands",
			"Search Topics",
			"View Examples",
			"FAQ",
			"Troubleshooting",
			"Quick Start Guide",
			"Exit Help",
		}
		
		choice, err := hs.terminal.Select("What would you like help with?", options)
		if err != nil {
			return err
		}
		
		switch choice {
		case 0:
			hs.browseCommands()
		case 1:
			hs.searchTopics()
		case 2:
			hs.viewExamples()
		case 3:
			hs.showFAQ()
		case 4:
			hs.troubleshoot()
		case 5:
			hs.quickStart()
		case 6:
			return nil
		}
	}
}

// browseCommands allows browsing through available commands
func (hs *HelpSystem) browseCommands() {
	categories := map[string][]string{
		"Scanning": {"scan", "validate", "analyze"},
		"Templates": {"template list", "template create", "template validate"},
		"Configuration": {"config init", "config set", "config show"},
		"Reporting": {"report generate", "report export", "report view"},
		"Security": {"auth login", "auth logout", "auth status"},
	}
	
	hs.terminal.Section("Command Categories")
	
	for category, commands := range categories {
		hs.terminal.Subsection(category)
		for _, cmd := range commands {
			hs.terminal.Info("  " + cmd)
			if examples, ok := hs.examples[strings.Split(cmd, " ")[0]]; ok && len(examples) > 0 {
				hs.terminal.Muted("    Example: " + examples[0].Command)
			}
		}
		fmt.Println()
	}
	
	hs.terminal.Prompt("Press Enter to continue...")
}

// searchTopics allows searching for help topics
func (hs *HelpSystem) searchTopics() {
	query, err := hs.terminal.Input("Search for help topic:", "")
	if err != nil {
		return
	}
	
	results := hs.searchContent(query)
	
	if len(results) == 0 {
		hs.terminal.Warning("No results found for: " + query)
		return
	}
	
	hs.terminal.Section("Search Results for: " + query)
	
	for _, result := range results {
		switch r := result.(type) {
		case Example:
			hs.terminal.Subsection("Example")
			hs.terminal.Code(r.Command)
			hs.terminal.Info(r.Description)
		case FAQ:
			hs.terminal.Subsection("FAQ")
			hs.terminal.Bold(r.Question)
			hs.terminal.Info(r.Answer)
		case string:
			hs.terminal.Info("• " + r)
		}
		fmt.Println()
	}
	
	hs.terminal.Prompt("Press Enter to continue...")
}

// viewExamples shows categorized examples
func (hs *HelpSystem) viewExamples() {
	categories := []string{}
	for cat := range hs.examples {
		categories = append(categories, cat)
	}
	
	choice, err := hs.terminal.Select("Select example category:", categories)
	if err != nil {
		return
	}
	
	category := categories[choice]
	examples := hs.examples[category]
	
	hs.terminal.Section(fmt.Sprintf("%s Examples", strings.Title(category)))
	
	for i, ex := range examples {
		hs.terminal.Subsection(fmt.Sprintf("Example %d: %s", i+1, ex.Description))
		hs.terminal.Code(ex.Command)
		if ex.Output != "" {
			hs.terminal.Box("Expected Output", ex.Output)
		}
		if len(ex.Tags) > 0 {
			hs.terminal.Muted("Tags: " + strings.Join(ex.Tags, ", "))
		}
		fmt.Println()
	}
	
	hs.terminal.Prompt("Press Enter to continue...")
}

// showFAQ displays frequently asked questions
func (hs *HelpSystem) showFAQ() {
	hs.terminal.Section("Frequently Asked Questions")
	
	for category, faqs := range hs.faqs {
		hs.terminal.Subsection(strings.Title(category))
		for _, faq := range faqs {
			hs.terminal.Bold("Q: " + faq.Question)
			hs.terminal.Info("A: " + faq.Answer)
			if len(faq.Related) > 0 {
				hs.terminal.Muted("   Related: " + strings.Join(faq.Related, ", "))
			}
			fmt.Println()
		}
	}
	
	hs.terminal.Prompt("Press Enter to continue...")
}

// troubleshoot provides troubleshooting guidance
func (hs *HelpSystem) troubleshoot() {
	issues := []string{
		"Scan is running slowly",
		"Authentication errors",
		"Template validation failures",
		"Network connectivity issues",
		"Permission denied errors",
		"Out of memory errors",
	}
	
	choice, err := hs.terminal.Select("What issue are you experiencing?", issues)
	if err != nil {
		return
	}
	
	hs.terminal.Section("Troubleshooting: " + issues[choice])
	
	// Provide specific troubleshooting steps
	switch choice {
	case 0: // Slow scans
		hs.showSlowScanTroubleshooting()
	case 1: // Auth errors
		hs.showAuthTroubleshooting()
	case 2: // Template validation
		hs.showTemplateValidationTroubleshooting()
	case 3: // Network issues
		hs.showNetworkTroubleshooting()
	case 4: // Permission errors
		hs.showPermissionTroubleshooting()
	case 5: // Memory errors
		hs.showMemoryTroubleshooting()
	}
	
	hs.terminal.Prompt("Press Enter to continue...")
}

// quickStart shows a quick start guide
func (hs *HelpSystem) quickStart() {
	hs.terminal.Section("Quick Start Guide")
	
	steps := []struct {
		Title       string
		Command     string
		Description string
	}{
		{
			Title:       "1. Initialize Configuration",
			Command:     "LLMrecon init",
			Description: "Creates a default configuration file with common settings",
		},
		{
			Title:       "2. Configure Authentication",
			Command:     "LLMrecon auth login",
			Description: "Set up authentication for API access",
		},
		{
			Title:       "3. List Available Templates",
			Command:     "LLMrecon template list",
			Description: "See all available security test templates",
		},
		{
			Title:       "4. Run Your First Scan",
			Command:     "LLMrecon scan --template basic-prompt-injection --target http://localhost:8080",
			Description: "Execute a basic security scan",
		},
		{
			Title:       "5. View Results",
			Command:     "LLMrecon report view latest",
			Description: "Review the scan results and findings",
		},
	}
	
	for _, step := range steps {
		hs.terminal.Subsection(step.Title)
		hs.terminal.Code(step.Command)
		hs.terminal.Info(step.Description)
		fmt.Println()
	}
	
	hs.terminal.Box("Next Steps", `
• Explore more templates: LLMrecon template list --detailed
• Create custom templates: LLMrecon template create
• Set up continuous scanning: LLMrecon schedule create
• Join the community: https://github.com/LLMrecon/community
`)
	
	hs.terminal.Prompt("Press Enter to continue...")
}

// Helper methods

func (hs *HelpSystem) updateContext(cmd *cobra.Command, args []string) {
	hs.context.Command = cmd.Name()
	if len(args) > 0 {
		hs.context.Subcommand = args[0]
	}
	// Update other context as needed
}

func (hs *HelpSystem) showUsage(cmd *cobra.Command) {
	hs.terminal.Section("Usage")
	hs.terminal.Code(cmd.UseLine())
	if cmd.Long != "" {
		hs.terminal.Info(cmd.Long)
	}
	fmt.Println()
}

func (hs *HelpSystem) showErrorHelp() {
	hs.terminal.Section("Error Resolution")
	hs.terminal.Error("Last error: " + hs.context.ErrorContext)
	
	// Provide context-specific error help
	suggestions := hs.suggester.GetErrorSuggestions(hs.context.ErrorContext)
	if len(suggestions) > 0 {
		hs.terminal.Subsection("Suggested Solutions")
		for _, sug := range suggestions {
			hs.terminal.Info("• " + sug.Description)
			if sug.Command != "" {
				hs.terminal.Code("  " + sug.Command)
			}
		}
	}
	fmt.Println()
}

func (hs *HelpSystem) showExamples(command string) {
	if examples, ok := hs.examples[command]; ok && len(examples) > 0 {
		hs.terminal.Section("Examples")
		for i, ex := range examples[:min(3, len(examples))] {
			hs.terminal.Subsection(fmt.Sprintf("Example %d", i+1))
			hs.terminal.Code(ex.Command)
			hs.terminal.Muted(ex.Description)
		}
		fmt.Println()
	}
}

func (hs *HelpSystem) showTips(command string) {
	if tips, ok := hs.tips[command]; ok && len(tips) > 0 {
		hs.terminal.Section("Pro Tips")
		for _, tip := range tips[:min(3, len(tips))] {
			hs.terminal.Info("💡 " + tip)
		}
		fmt.Println()
	}
}

func (hs *HelpSystem) showRelatedCommands(cmd *cobra.Command) {
	if cmd.Parent() != nil && len(cmd.Parent().Commands()) > 1 {
		hs.terminal.Section("Related Commands")
		count := 0
		for _, sibling := range cmd.Parent().Commands() {
			if sibling != cmd && count < 5 {
				hs.terminal.Info(fmt.Sprintf("• %s - %s", sibling.Name(), sibling.Short))
				count++
			}
		}
		fmt.Println()
	}
}

func (hs *HelpSystem) showQuickActions(cmd *cobra.Command) {
	hs.terminal.Section("Quick Actions")
	actions := []string{
		"Press 'h' for interactive help",
		"Press 'e' to see more examples",
		"Press 's' to search help topics",
		"Press 'q' to quit help",
	}
	for _, action := range actions {
		hs.terminal.Muted(action)
	}
}

func (hs *HelpSystem) searchContent(query string) []interface{} {
	var results []interface{}
	query = strings.ToLower(query)
	
	// Search examples
	for _, examples := range hs.examples {
		for _, ex := range examples {
			if strings.Contains(strings.ToLower(ex.Command), query) ||
				strings.Contains(strings.ToLower(ex.Description), query) {
				results = append(results, ex)
			}
		}
	}
	
	// Search FAQs
	for _, faqs := range hs.faqs {
		for _, faq := range faqs {
			if strings.Contains(strings.ToLower(faq.Question), query) ||
				strings.Contains(strings.ToLower(faq.Answer), query) {
				results = append(results, faq)
			}
		}
	}
	
	// Search tips
	for _, tips := range hs.tips {
		for _, tip := range tips {
			if strings.Contains(strings.ToLower(tip), query) {
				results = append(results, tip)
			}
		}
	}
	
	return results
}

// Troubleshooting methods

func (hs *HelpSystem) showSlowScanTroubleshooting() {
	steps := []string{
		"1. Check network latency: ping your target endpoint",
		"2. Enable parallel scanning: add --parallel flag",
		"3. Reduce template complexity: use --simple-mode",
		"4. Increase timeout values: --timeout 60s",
		"5. Use local caching: --enable-cache",
		"6. Monitor resource usage: LLMrecon debug stats",
	}
	
	hs.terminal.Subsection("Troubleshooting Steps")
	for _, step := range steps {
		hs.terminal.Info(step)
		time.Sleep(100 * time.Millisecond) // Visual effect
	}
	
	hs.terminal.Box("Advanced Diagnostics", "Run 'LLMrecon debug performance' for detailed performance analysis")
}

func (hs *HelpSystem) showAuthTroubleshooting() {
	hs.terminal.Subsection("Common Authentication Issues")
	
	issues := map[string]string{
		"Invalid credentials":    "LLMrecon auth login --force",
		"Expired token":          "LLMrecon auth refresh",
		"Missing permissions":    "LLMrecon auth status",
		"Configuration missing":  "LLMrecon config set auth.provider <provider>",
	}
	
	for issue, solution := range issues {
		hs.terminal.Bold(issue)
		hs.terminal.Code(solution)
		fmt.Println()
	}
}

func (hs *HelpSystem) showTemplateValidationTroubleshooting() {
	hs.terminal.Subsection("Template Validation Checklist")
	
	checklist := []string{
		"✓ Verify YAML syntax is correct",
		"✓ Check required fields are present",
		"✓ Validate variable references",
		"✓ Ensure template version compatibility",
		"✓ Test with minimal template first",
	}
	
	for _, item := range checklist {
		hs.terminal.Info(item)
	}
	
	hs.terminal.Code("LLMrecon template validate <template> --verbose")
}

func (hs *HelpSystem) showNetworkTroubleshooting() {
	hs.terminal.Subsection("Network Diagnostics")
	
	commands := []struct {
		Desc string
		Cmd  string
	}{
		{"Test connectivity", "LLMrecon debug ping <target>"},
		{"Check DNS resolution", "LLMrecon debug dns <hostname>"},
		{"Verify SSL/TLS", "LLMrecon debug tls <target>"},
		{"Test with proxy", "LLMrecon --proxy http://proxy:8080 scan ..."},
		{"Bypass SSL verify", "LLMrecon --insecure scan ... (NOT for production)"},
	}
	
	for _, cmd := range commands {
		hs.terminal.Bold(cmd.Desc)
		hs.terminal.Code(cmd.Cmd)
		fmt.Println()
	}
}

func (hs *HelpSystem) showPermissionTroubleshooting() {
	hs.terminal.Subsection("Permission Issues")
	
	hs.terminal.Info("Common causes and solutions:")
	fmt.Println()
	
	solutions := []struct {
		Issue    string
		Solution string
	}{
		{"Config file permissions", "chmod 600 ~/.LLMrecon/config.yaml"},
		{"Log directory access", "mkdir -p ~/.LLMrecon/logs && chmod 755 ~/.LLMrecon/logs"},
		{"Template directory", "LLMrecon config set template.path /path/to/writable/dir"},
		{"System-wide install", "Use sudo for system-wide operations or install to user directory"},
	}
	
	for _, sol := range solutions {
		hs.terminal.Bold(sol.Issue)
		hs.terminal.Code(sol.Solution)
		fmt.Println()
	}
}

func (hs *HelpSystem) showMemoryTroubleshooting() {
	hs.terminal.Subsection("Memory Optimization")
	
	hs.terminal.Info("Reduce memory usage with these options:")
	fmt.Println()
	
	optimizations := []string{
		"--stream-results: Process results as they arrive",
		"--max-concurrent 2: Limit parallel operations",
		"--no-cache: Disable result caching",
		"--light-mode: Use minimal UI features",
		"--batch-size 10: Process in smaller batches",
	}
	
	for _, opt := range optimizations {
		hs.terminal.Code(opt)
	}
	
	hs.terminal.Box("Monitor Memory", "LLMrecon debug memory --watch")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}