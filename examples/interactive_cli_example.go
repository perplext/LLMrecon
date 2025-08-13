package main

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/ui"
)

func main() {
	// Create terminal with progress support
	terminal := ui.NewTerminal(ui.TerminalOptions{
		Output:       os.Stdout,
		Input:        os.Stdin,
		ColorOutput:  true,
		ShowProgress: true,
		ClearScreen:  false,
	})

	// Show welcome message
	terminal.Header("LLMrecon - Interactive CLI Demo")
	terminal.Info("This demo showcases various UI features available in the tool.\n")

	// Menu loop
	for {
		options := []string{
			"Progress Indicators Demo",
			"Interactive Task Runner",
			"Template Selection Demo",
			"Configuration Wizard Demo",
			"Dashboard View Demo",
			"Exit",
		}

		choice, err := terminal.Select("Select a demo to run:", options)
		if err != nil {
			terminal.Error("Selection error: %v", err)
			break
		}

		switch choice {
		case 0:
			demoProgressIndicators(terminal)
		case 1:
			demoInteractiveTaskRunner(terminal)
		case 2:
			demoTemplateSelection(terminal)
		case 3:
			demoConfigurationWizard(terminal)
		case 4:
			demoDashboardView(terminal)
		case 5:
			terminal.Success("Thanks for using LLMrecon!")
			return
		}

		terminal.Print("\n")
	}
}

// demoProgressIndicators demonstrates various progress indicators
func demoProgressIndicators(terminal *ui.Terminal) {
	terminal.Header("Progress Indicators Demo")

	// 1. Simple spinner
	terminal.Info("1. Spinner for indeterminate tasks:")
	stop := terminal.StartSpinner("Connecting to LLM provider")
	time.Sleep(3 * time.Second)
	stop()
	terminal.Success("Connected successfully!")

	// 2. Progress bar
	terminal.Info("\n2. Progress bar for determinate tasks:")
	terminal.StartProgress("scan", "Scanning for vulnerabilities", 50)
	for i := 0; i <= 50; i++ {
		terminal.UpdateProgress("scan", int64(i))
		time.Sleep(50 * time.Millisecond)
	}
	terminal.FinishProgress("scan")
	terminal.Success("Scan completed!")

	// 3. Multi-task progress
	terminal.Info("\n3. Multiple concurrent tasks:")
	multiProg := ui.NewMultiProgress(os.Stdout, 5, true)
	
	// Simulate concurrent tasks
	go func() {
		task1 := multiProg.AddTask("task1", "Loading templates")
		multiProg.UpdateTask(task1.ID, ui.TaskRunning, 0.0, "Fetching from repository...")
		time.Sleep(1 * time.Second)
		multiProg.UpdateTask(task1.ID, ui.TaskRunning, 0.5, "Parsing YAML files...")
		time.Sleep(1 * time.Second)
		multiProg.UpdateTask(task1.ID, ui.TaskCompleted, 1.0, "156 templates loaded")
	}()

	go func() {
		time.Sleep(500 * time.Millisecond)
		task2 := multiProg.AddTask("task2", "Initializing providers")
		multiProg.UpdateTask(task2.ID, ui.TaskRunning, 0.0, "Setting up OpenAI...")
		time.Sleep(1500 * time.Millisecond)
		multiProg.UpdateTask(task2.ID, ui.TaskRunning, 0.7, "Configuring rate limits...")
		time.Sleep(1 * time.Second)
		multiProg.UpdateTask(task2.ID, ui.TaskCompleted, 1.0, "3 providers ready")
	}()

	go func() {
		time.Sleep(1 * time.Second)
		task3 := multiProg.AddTask("task3", "Validating configuration")
		multiProg.UpdateTask(task3.ID, ui.TaskRunning, 0.0, "Checking API keys...")
		time.Sleep(2 * time.Second)
		multiProg.UpdateTask(task3.ID, ui.TaskFailed, 0.8, "Invalid API key format")
	}()

	// Show progress updates
	for i := 0; i < 10; i++ {
		terminal.ClearPreviousLines(5)
		terminal.Print(multiProg.Render())
		time.Sleep(500 * time.Millisecond)
	}
}

// demoInteractiveTaskRunner demonstrates the interactive task runner
func demoInteractiveTaskRunner(terminal *ui.Terminal) {
	terminal.Header("Interactive Task Runner Demo")

	// Create runner with options
	runnerOpts := ui.DefaultRunnerOptions()
	runnerOpts.ConcurrentTasks = 2
	runnerOpts.ShowTaskDetails = true
	
	runner := ui.NewInteractiveRunner(terminal, runnerOpts)

	// Add various tasks
	runner.AddTask(ui.RunnerTask{
		ID:          "init",
		Name:        "Initialize Environment",
		Description: "Set up directories and load configuration",
		Required:    true,
		Execute: func(ctx context.Context, progress ui.ProgressReporter) error {
			progress.SetStatus("Creating directories...")
			time.Sleep(1 * time.Second)
			progress.SetStatus("Loading configuration...")
			time.Sleep(1 * time.Second)
			return nil
		},
	})

	runner.AddTask(ui.RunnerTask{
		ID:          "download",
		Name:        "Download Templates",
		Description: "Fetch latest vulnerability templates from repository",
		Retryable:   true,
		Execute: func(ctx context.Context, progress ui.ProgressReporter) error {
			progress.SetTotal(100)
			for i := 0; i <= 100; i += 10 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					progress.SetCurrent(int64(i))
					progress.SetDetails(fmt.Sprintf("Downloaded %d/100 templates", i))
					time.Sleep(200 * time.Millisecond)
				}
			}
			return nil
		},
	})

	runner.AddTask(ui.RunnerTask{
		ID:          "validate",
		Name:        "Validate Templates",
		Description: "Check template syntax and compatibility",
		Execute: func(ctx context.Context, progress ui.ProgressReporter) error {
			templates := []string{"prompt-injection", "data-leakage", "jailbreak", "toxic-content"}
			progress.SetTotal(int64(len(templates)))
			
			for i, template := range templates {
				progress.AddSubTask(fmt.Sprintf("Validating %s", template))
				time.Sleep(500 * time.Millisecond)
				progress.SetCurrent(int64(i + 1))
			}
			return nil
		},
	})

	runner.AddTask(ui.RunnerTask{
		ID:          "connect",
		Name:        "Connect to LLM Providers",
		Description: "Establish connections to configured LLM APIs",
		Retryable:   true,
		Execute: func(ctx context.Context, progress ui.ProgressReporter) error {
			providers := []string{"OpenAI", "Anthropic", "Google"}
			for _, provider := range providers {
				progress.SetStatus(fmt.Sprintf("Connecting to %s...", provider))
				time.Sleep(800 * time.Millisecond)
			}
			return nil
		},
	})

	// Allow user to select tasks
	if err := runner.SelectTasks(); err != nil {
		terminal.Error("Task selection cancelled: %v", err)
		return
	}

	// Run selected tasks
	ctx := context.Background()
	if err := runner.Run(ctx); err != nil {
		terminal.Error("Task execution failed: %v", err)
	}
}

// demoTemplateSelection demonstrates interactive template selection
func demoTemplateSelection(terminal *ui.Terminal) {
	terminal.Header("Template Selection Demo")

	// Categories
	categories := []string{
		"OWASP LLM Top 10",
		"Prompt Injection",
		"Data Leakage",
		"Model Manipulation",
		"Content Safety",
		"All Templates",
	}

	categoryChoice, err := terminal.Select("Select template category:", categories)
	if err != nil {
		return
	}

	// Templates based on category
	var templates []string
	switch categoryChoice {
	case 0: // OWASP
		templates = []string{
			"LLM01 - Prompt Injection",
			"LLM02 - Insecure Output Handling",
			"LLM03 - Training Data Poisoning",
			"LLM04 - Model Denial of Service",
			"LLM05 - Supply Chain Vulnerabilities",
			"LLM06 - Sensitive Information Disclosure",
			"LLM07 - Insecure Plugin Design",
			"LLM08 - Excessive Agency",
			"LLM09 - Overreliance",
			"LLM10 - Model Theft",
		}
	case 1: // Prompt Injection
		templates = []string{
			"Basic Prompt Override",
			"System Prompt Extraction",
			"Role-Playing Attack",
			"Instruction Bypass",
			"Context Manipulation",
		}
	default:
		templates = []string{"Template 1", "Template 2", "Template 3"}
	}

	// Multi-select templates
	selected, err := terminal.MultiSelect("Select templates to run", templates)
	if err != nil {
		return
	}

	// Show selected templates
	terminal.Success("\nSelected %d templates:", len(selected))
	for _, idx := range selected {
		terminal.Print("  • %s", templates[idx])
	}

	// Simulate loading templates
	stop := terminal.StartSpinner("Loading selected templates")
	time.Sleep(2 * time.Second)
	stop()

	terminal.Success("Templates loaded and ready for execution!")
}

// demoConfigurationWizard demonstrates configuration wizard
func demoConfigurationWizard(terminal *ui.Terminal) {
	terminal.Header("Configuration Wizard Demo")

	terminal.Info("This wizard will help you configure LLMrecon.\n")

	// Step 1: Provider selection
	terminal.Print("Step 1: Select LLM Providers")
	providers := []string{"OpenAI", "Anthropic", "Google PaLM", "Cohere", "Local Model"}
	selectedProviders, err := terminal.MultiSelect("Which providers would you like to configure?", providers)
	if err != nil {
		return
	}

	// Step 2: API Keys
	terminal.Print("\nStep 2: API Configuration")
	for _, idx := range selectedProviders {
		provider := providers[idx]
		if provider != "Local Model" {
			apiKey, err := terminal.Prompt(fmt.Sprintf("Enter API key for %s: ", provider))
			if err != nil {
				continue
			}
			terminal.Success("API key for %s configured (hidden)", provider)
			_ = apiKey // Would store securely
		}
	}

	// Step 3: Test configuration
	terminal.Print("\nStep 3: Test Configuration")
	testConfig := []string{
		"Enable concurrent testing",
		"Set timeout (seconds)",
		"Maximum retries",
		"Rate limiting",
	}

	// Concurrent testing
	concurrent, _ := terminal.Confirm("Enable concurrent testing?", true)
	if concurrent {
		concurrency, _ := terminal.Prompt("Number of concurrent tests (default: 5): ")
		if concurrency == "" {
			concurrency = "5"
		}
		terminal.Info("Concurrent testing: %s tests", concurrency)
	}

	// Timeout
	timeout, _ := terminal.Prompt("Test timeout in seconds (default: 30): ")
	if timeout == "" {
		timeout = "30"
	}
	terminal.Info("Test timeout: %s seconds", timeout)

	// Step 4: Output configuration
	terminal.Print("\nStep 4: Output Configuration")
	formats := []string{"JSON", "Markdown", "PDF", "Excel", "CSV", "HTML"}
	selectedFormats, _ := terminal.MultiSelect("Select output formats", formats)
	
	terminal.Success("\nConfiguration Summary:")
	terminal.Print("Providers: %d configured", len(selectedProviders))
	terminal.Print("Concurrent tests: %v", concurrent)
	terminal.Print("Timeout: %s seconds", timeout)
	terminal.Print("Output formats: %d selected", len(selectedFormats))

	save, _ := terminal.Confirm("\nSave configuration?", true)
	if save {
		stop := terminal.StartSpinner("Saving configuration")
		time.Sleep(1 * time.Second)
		stop()
		terminal.Success("Configuration saved to config.yaml")
	}
}

// demoDashboardView demonstrates dashboard-style output
func demoDashboardView(terminal *ui.Terminal) {
	terminal.Header("Dashboard View Demo")

	// Scan summary
	terminal.Print("Current Scan: security-scan-2024-01-20")
	terminal.Print("Status: Running")
	terminal.Print("Duration: 5m 23s\n")

	// Progress overview
	headers := []string{"Category", "Total", "Completed", "Failed", "Progress"}
	rows := [][]string{
		{"Prompt Injection", "25", "20", "2", "[=========>  ] 80%"},
		{"Data Leakage", "15", "15", "0", "[============] 100%"},
		{"Model Manipulation", "10", "5", "1", "[=====>      ] 50%"},
		{"Content Safety", "20", "8", "0", "[===>        ] 40%"},
	}
	terminal.Table(headers, rows)

	// Recent findings
	terminal.Print("\nRecent Findings:")
	findings := []string{
		"HIGH: System prompt extraction possible via role-play attack",
		"MEDIUM: Model reveals training data when prompted with specific patterns",
		"LOW: Inconsistent content filtering for edge cases",
		"INFO: Model version information exposed in error messages",
	}

	for _, finding := range findings {
		if strings.HasPrefix(finding, "HIGH:") {
			terminal.Error("  %s", finding)
		} else if strings.HasPrefix(finding, "MEDIUM:") {
			terminal.Warning("  %s", finding)
		} else if strings.HasPrefix(finding, "LOW:") {
			terminal.Info("  %s", finding)
		} else {
			terminal.Debug("  %s", finding)
		}
	}

	// Live updates simulation
	terminal.Print("\nLive Test Feed:")
	stop := terminal.StartSpinner("Running test: prompt-injection-advanced-001")
	time.Sleep(2 * time.Second)
	stop()
	terminal.Success("Test completed: prompt-injection-advanced-001 [PASS]")

	stop = terminal.StartSpinner("Running test: data-extraction-personal-info")
	time.Sleep(1500 * time.Millisecond)
	stop()
	terminal.Error("Test completed: data-extraction-personal-info [FAIL]")

	// Summary stats
	terminal.Print("\nSummary Statistics:")
	stats := [][]string{
		{"Total Tests Run", "68"},
		{"Tests Passed", "55"},
		{"Tests Failed", "3"},
		{"Tests Skipped", "10"},
		{"Success Rate", "94.8%"},
		{"Average Response Time", "342ms"},
	}

	statsHeaders := []string{"Metric", "Value"}
	terminal.Table(statsHeaders, stats)
}