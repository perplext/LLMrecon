// Package cmd provides command-line commands for the LLMrecon tool
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/perplext/LLMrecon/src/security/prompt"
	"github.com/spf13/cobra"
)

// promptProtectionCmd represents the prompt-protection command
var promptProtectionCmd = &cobra.Command{
	Use:   "prompt-protection",
	Short: "Manage prompt injection protection",
	Long: `Manage the enhanced prompt injection protection system for detecting and preventing 
prompt injection attacks and other LLM-specific security threats.

This command provides subcommands for configuring and managing the prompt
injection protection system, including:
- Configuring protection levels and features
- Managing pattern libraries
- Viewing detection reports
- Testing prompts for potential security issues
- Monitoring template usage
- Managing approval workflows`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},

// configureCmd represents the prompt-protection configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the prompt injection protection system",
	Long: `Configure the prompt injection protection system settings.

This command allows you to set the protection level, enable or disable specific
protection features, and configure other settings for the prompt injection
protection system.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		level, _ := cmd.Flags().GetString("level")
		configFile, _ := cmd.Flags().GetString("config")

		// Convert level string to ProtectionLevel
		var protectionLevel prompt.ProtectionLevel
		switch strings.ToLower(level) {
		case "low":
			protectionLevel = prompt.LevelLow
		case "medium":
			protectionLevel = prompt.LevelMedium
		case "high":
			protectionLevel = prompt.LevelHigh
		case "custom":
			protectionLevel = prompt.LevelCustom
		default:
			fmt.Println("Invalid protection level. Using medium level.")
			protectionLevel = prompt.LevelMedium
		}

		// Create a default config based on the protection level
		var config *prompt.ProtectionConfig
		switch protectionLevel {
		case prompt.LevelLow:
			config = prompt.DefaultProtectionConfig()
			config.Level = prompt.LevelLow
			config.EnableJailbreakDetection = true
			config.EnableContextBoundaries = true
			config.EnableRealTimeMonitoring = false
			config.EnableContentFiltering = true
			config.EnableApprovalWorkflow = false
			config.EnableReportingSystem = true
			config.SanitizationLevel = 1
		case prompt.LevelMedium:
			config = prompt.DefaultProtectionConfig()
		case prompt.LevelHigh:
			config = prompt.HighSecurityProtectionConfig()
		case prompt.LevelCustom:
			// Load from config file if provided
			if configFile != "" {
				var err error
				config, err = loadConfigFromFile(configFile)
				if err != nil {
					fmt.Printf("Error loading config file: %v\n", err)
					fmt.Println("Using default medium level configuration.")
					config = prompt.DefaultProtectionConfig()
				}
			} else {
				// Start with medium config and let user customize
				config = prompt.DefaultProtectionConfig()
				fmt.Println("Using custom configuration. You can modify it with --enable and --disable flags.")
			}
		}

		// Apply enable/disable flags
		enableFeatures, _ := cmd.Flags().GetStringSlice("enable")
		disableFeatures, _ := cmd.Flags().GetStringSlice("disable")

		for _, feature := range enableFeatures {
			enableFeature(config, feature)
		}

		for _, feature := range disableFeatures {
			disableFeature(config, feature)
		}

		// Save the configuration
		outputFile, _ := cmd.Flags().GetString("output")
		if outputFile == "" {
			outputFile = "config/prompt_protection.json"
		}

		err := saveConfigToFile(config, outputFile)
		if err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		fmt.Printf("Prompt protection configuration saved to %s\n", outputFile)
		fmt.Println("Configuration summary:")
		printConfigSummary(config)
	},

// testCmd represents the prompt-protection test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test a prompt for security issues",
	Long:  `Test a prompt for potential security issues such as prompt injection attempts, jailbreaking attempts, or other security threats using the enhanced protection system.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		promptText, _ := cmd.Flags().GetString("prompt")
		promptFile, _ := cmd.Flags().GetString("file")
		configFile, _ := cmd.Flags().GetString("config")

		// Load the prompt text
		if promptText == "" && promptFile == "" {
			fmt.Println("Error: Either --prompt or --file must be provided.")
			cmd.Help()
			return
		}

		if promptFile != "" {
			data, err := os.ReadFile(filepath.Clean(promptFile))
			if err != nil {
				fmt.Printf("Error reading prompt file: %v\n", err)
				return
			}
			promptText = string(data)
		}

		// Load the configuration
		var config *prompt.ProtectionConfig
		if configFile != "" {
			var err error
			config, err = loadConfigFromFile(configFile)
			if err != nil {
				fmt.Printf("Error loading config file: %v\n", err)
				fmt.Println("Using default configuration.")
				config = prompt.DefaultProtectionConfig()
			}
		} else {
			config = prompt.DefaultProtectionConfig()
		}

		// Create an enhanced protection manager
		manager, err := prompt.NewEnhancedProtectionManager(config)
		if err != nil {
			fmt.Printf("Error creating enhanced protection manager: %v\n", err)
			return
		}

		// Create a context
		ctx := context.Background()

		// Generate session and user IDs
		sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
		userID := "cli-user"
		templateID := "cli-template"

		// Protect the prompt with enhanced protection
		protectedPrompt, result, err := manager.ProtectPromptEnhanced(ctx, promptText, userID, sessionID, templateID)
		if err != nil {
			fmt.Printf("Error protecting prompt: %v\n", err)
			return
		}

		// Print the results
		fmt.Println("Prompt Protection Test Results")
		fmt.Println("==============================")
		fmt.Printf("Risk Score: %.2f\n", result.RiskScore)
		fmt.Printf("Action Taken: %s\n", actionTypeToString(result.ActionTaken))
		fmt.Printf("Processing Time: %v\n", result.ProcessingTime)
		fmt.Printf("Detections: %d\n", len(result.Detections))

		if len(result.Detections) > 0 {
			fmt.Println("\nDetections:")
			for i, detection := range result.Detections {
				fmt.Printf("\n%d. Type: %s\n", i+1, detection.Type)
				fmt.Printf("   Confidence: %.2f\n", detection.Confidence)
				fmt.Printf("   Description: %s\n", detection.Description)
				if detection.Location != nil {
					fmt.Printf("   Context: %s\n", detection.Location.Context)
				}
				if detection.Remediation != "" {
					fmt.Printf("   Remediation: %s\n", detection.Remediation)
				}
			}
		}

		if protectedPrompt != promptText {
			fmt.Println("\nOriginal Prompt:")
			fmt.Println(promptText)
			fmt.Println("\nProtected Prompt:")
			fmt.Println(protectedPrompt)
		} else {
			fmt.Println("\nNo modifications were made to the prompt.")
		}

		// Print risk assessment
		fmt.Println("\nRisk Assessment:")
		if result.RiskScore >= 0.9 {
			fmt.Println("HIGH RISK - This prompt contains serious security issues and should be blocked.")
		} else if result.RiskScore >= 0.7 {
			fmt.Println("MEDIUM RISK - This prompt contains potential security issues and should be reviewed.")
		} else if result.RiskScore >= 0.5 {
			fmt.Println("LOW RISK - This prompt contains minor security issues but may be acceptable with modifications.")
		} else {
			fmt.Println("MINIMAL RISK - This prompt appears to be safe.")
		}
	},

// patternsCmd represents the prompt-protection patterns command
var patternsCmd = &cobra.Command{
	Use:   "patterns",
	Short: "Manage prompt injection patterns",
	Long: `Manage the library of prompt injection patterns.

This command allows you to list, add, remove, enable, or disable patterns in the
prompt injection pattern library.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		action, _ := cmd.Flags().GetString("action")
		patternID, _ := cmd.Flags().GetString("id")
		patternFile, _ := cmd.Flags().GetString("file")

		// Create a pattern library
		patternLibrary := prompt.NewInjectionPatternLibrary()

		// Load patterns from file if provided
		if patternFile != "" && action != "save" {
			err := patternLibrary.LoadPatternsFromFile(patternFile)
			if err != nil {
				fmt.Printf("Error loading patterns from file: %v\n", err)
				return
			}
			fmt.Printf("Loaded patterns from %s\n", patternFile)
		}

		// Perform the requested action
		switch action {
		case "list":
			listPatterns(patternLibrary)
		case "enable":
			if patternID == "" {
				fmt.Println("Error: Pattern ID must be provided with --id flag.")
				return
			}
			patternLibrary.EnablePattern(patternID)
			fmt.Printf("Enabled pattern %s\n", patternID)
		case "disable":
			if patternID == "" {
				fmt.Println("Error: Pattern ID must be provided with --id flag.")
				return
			}
			patternLibrary.DisablePattern(patternID)
			fmt.Printf("Disabled pattern %s\n", patternID)
		case "save":
			if patternFile == "" {
				fmt.Println("Error: Pattern file must be provided with --file flag.")
				return
			}
			err := patternLibrary.SavePatternsToFile(patternFile)
			if err != nil {
				fmt.Printf("Error saving patterns to file: %v\n", err)
				return
			}
			fmt.Printf("Saved patterns to %s\n", patternFile)
		default:
			fmt.Printf("Unknown action: %s\n", action)
			cmd.Help()
		}
	},

// reportsCmd represents the prompt-protection reports command
var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "View prompt injection reports",
	Long: `View reports of prompt injection attempts.

This command allows you to view reports of prompt injection attempts that have
been detected by the prompt injection protection system.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		reportsDir, _ := cmd.Flags().GetString("dir")
		reportID, _ := cmd.Flags().GetString("id")

		// Ensure reports directory exists
		if reportsDir == "" {
			reportsDir = "reports"
		}

		if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
			fmt.Printf("Reports directory %s does not exist.\n", reportsDir)
			return
		}

		// If a specific report ID is provided, show that report
		if reportID != "" {
			showReport(reportsDir, reportID)
			return
		}

		// Otherwise, list all reports
		listReports(reportsDir)
	},

// monitorCmd represents the prompt-protection monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor template usage",
	Long:  `Monitor template usage for unusual patterns and behaviors that may indicate security issues.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		duration, _ := cmd.Flags().GetDuration("duration")
		if duration == 0 {
			duration = time.Minute * 10 // Default to 10 minutes
		}

		// Create an enhanced protection manager
		config := prompt.DefaultProtectionConfig()
		manager, err := prompt.NewEnhancedProtectionManager(config)
		if err != nil {
			fmt.Printf("Error creating enhanced protection manager: %v\n", err)
			return
		}
		defer func() { if err := manager.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		// Start monitoring
		fmt.Printf("Starting template monitoring for %v...\n", duration)
		if err := manager.StartMonitoring(ctx); err != nil {
			fmt.Printf("Error starting monitoring: %v\n", err)
			return
		}

		// Wait for monitoring to complete
		<-ctx.Done()

		// Get template monitor
		monitor := manager.GetTemplateMonitor()
		if monitor == nil {
			fmt.Println("Template monitor not available")
			return
		}

		// Get alert history
		alerts := monitor.GetAlertHistory()

		// Print alerts
		fmt.Printf("Monitoring completed. Found %d alerts.\n", len(alerts))
		for i, alert := range alerts {
			fmt.Printf("Alert %d:\n", i+1)
			fmt.Printf("  ID: %s\n", alert.AlertID)
			fmt.Printf("  Timestamp: %s\n", alert.Timestamp.Format(time.RFC3339))
			fmt.Printf("  Severity: %s\n", alert.Severity)
			fmt.Printf("  Type: %s\n", alert.Type)
			fmt.Printf("  Message: %s\n", alert.Message)
			if alert.TemplateID != "" {
				fmt.Printf("  Template ID: %s\n", alert.TemplateID)
			}
			if alert.UserID != "" {
				fmt.Printf("  User ID: %s\n", alert.UserID)
			}
			if alert.SessionID != "" {
				fmt.Printf("  Session ID: %s\n", alert.SessionID)
			}
			fmt.Printf("  Risk Score: %.2f\n", alert.RiskScore)
			fmt.Printf("  Anomaly Score: %.2f\n", alert.AnomalyScore)
			fmt.Println()
		}
	},

// approvalCmd represents the prompt-protection approval command
var approvalCmd = &cobra.Command{
	Use:   "approval",
	Short: "Manage approval workflow",
	Long:  `Manage the approval workflow for high-risk operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		action, _ := cmd.Flags().GetString("action")
		id, _ := cmd.Flags().GetString("id")

		// Create an enhanced protection manager
		config := prompt.DefaultProtectionConfig()
		manager, err := prompt.NewEnhancedProtectionManager(config)
		if err != nil {
			fmt.Printf("Error creating enhanced protection manager: %v\n", err)
			return
		}
		defer func() { if err := manager.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

		// Get approval workflow
		workflow := manager.GetApprovalWorkflow()
		if workflow == nil {
			fmt.Println("Approval workflow not available")
			return
		}

		// Handle different actions
		switch action {
		case "list":
			// List pending approvals
			pendingApprovals := workflow.GetPendingApprovals()
			fmt.Printf("Found %d pending approvals.\n", len(pendingApprovals))
			for i, approval := range pendingApprovals {
				fmt.Printf("Approval %d:\n", i+1)
				fmt.Printf("  ID: %s\n", approval.RequestID)
				fmt.Printf("  Timestamp: %s\n", approval.Timestamp.Format(time.RFC3339))
				fmt.Printf("  Risk Score: %.2f\n", approval.RiskScore)
				fmt.Printf("  Status: %s\n", approval.Status)
				fmt.Println()
			}
		case "approve":
			// Approve a specific request
			if id == "" {
				fmt.Println("Error: Approval ID is required")
				return
			}
			if err := workflow.ApproveRequest(id, "CLI-User", "Approved via CLI"); err != nil {
				fmt.Printf("Error approving request: %v\n", err)
				return
			}
			fmt.Printf("Request %s approved successfully.\n", id)
		case "reject":
			// Reject a specific request
			if id == "" {
				fmt.Println("Error: Approval ID is required")
				return
			}
			if err := workflow.RejectRequest(id, "CLI-User", "Rejected via CLI"); err != nil {
				fmt.Printf("Error rejecting request: %v\n", err)
				return
			}
			fmt.Printf("Request %s rejected successfully.\n", id)
		default:
			fmt.Println("Invalid action. Use 'list', 'approve', or 'reject'.")
		}
	},

func init() {
	rootCmd.AddCommand(promptProtectionCmd)
	promptProtectionCmd.AddCommand(configureCmd)
	promptProtectionCmd.AddCommand(testCmd)
	promptProtectionCmd.AddCommand(patternsCmd)
	promptProtectionCmd.AddCommand(reportsCmd)
	promptProtectionCmd.AddCommand(monitorCmd)
	promptProtectionCmd.AddCommand(approvalCmd)

	// Configure command flags
	configureCmd.Flags().String("level", "medium", "Protection level (low, medium, high, custom)")
	configureCmd.Flags().String("config", "", "Path to a configuration file")
	configureCmd.Flags().String("output", "", "Path to save the configuration")
	configureCmd.Flags().StringSlice("enable", []string{}, "Features to enable (comma-separated)")
	configureCmd.Flags().StringSlice("disable", []string{}, "Features to disable (comma-separated)")

	// Test command flags
	testCmd.Flags().String("prompt", "", "Prompt text to test")
	testCmd.Flags().String("file", "", "Path to a file containing the prompt to test")
	testCmd.Flags().String("config", "", "Path to a configuration file")

	// Patterns command flags
	patternsCmd.Flags().String("action", "list", "Action to perform (list, enable, disable, save)")
	patternsCmd.Flags().String("id", "", "Pattern ID")
	patternsCmd.Flags().String("file", "", "Path to a pattern file")
	patternsCmd.Flags().String("config", "", "Path to a configuration file")

	// Reports command flags
	reportsCmd.Flags().String("dir", "", "Directory containing reports")
	reportsCmd.Flags().String("id", "", "Report ID")

// Helper functions

// loadConfigFromFile loads a configuration from a JSON file
func loadConfigFromFile(filePath string) (*prompt.ProtectionConfig, error) {
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}

	var config prompt.ProtectionConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil

// saveConfigToFile saves a configuration to a JSON file
func saveConfigToFile(config *prompt.ProtectionConfig, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Clean(filePath, data, 0600))

// enableFeature enables a feature in the configuration
func enableFeature(config *prompt.ProtectionConfig, feature string) {
	switch strings.ToLower(feature) {
	case "context-boundaries", "boundaries":
		config.EnableContextBoundaries = true
	case "jailbreak-detection", "jailbreak":
		config.EnableJailbreakDetection = true
	case "real-time-monitoring", "monitoring":
		config.EnableRealTimeMonitoring = true
	case "content-filtering", "filtering":
		config.EnableContentFiltering = true
	case "approval-workflow", "approval":
		config.EnableApprovalWorkflow = true
	case "reporting-system", "reporting":
		config.EnableReportingSystem = true
	default:
		fmt.Printf("Unknown feature: %s\n", feature)
	}

// disableFeature disables a feature in the configuration
func disableFeature(config *prompt.ProtectionConfig, feature string) {
	switch strings.ToLower(feature) {
	case "context-boundaries", "boundaries":
		config.EnableContextBoundaries = false
	case "jailbreak-detection", "jailbreak":
		config.EnableJailbreakDetection = false
	case "real-time-monitoring", "monitoring":
		config.EnableRealTimeMonitoring = false
	case "content-filtering", "filtering":
		config.EnableContentFiltering = false
	case "approval-workflow", "approval":
		config.EnableApprovalWorkflow = false
	case "reporting-system", "reporting":
		config.EnableReportingSystem = false
	default:
		fmt.Printf("Unknown feature: %s\n", feature)
	}

// printConfigSummary prints a summary of the configuration
func printConfigSummary(config *prompt.ProtectionConfig) {
	fmt.Printf("Protection Level: %s\n", protectionLevelToString(config.Level))
	fmt.Printf("Context Boundaries: %s\n", boolToEnabledString(config.EnableContextBoundaries))
	fmt.Printf("Jailbreak Detection: %s\n", boolToEnabledString(config.EnableJailbreakDetection))
	fmt.Printf("Real-time Monitoring: %s\n", boolToEnabledString(config.EnableRealTimeMonitoring))
	fmt.Printf("Content Filtering: %s\n", boolToEnabledString(config.EnableContentFiltering))
	fmt.Printf("Approval Workflow: %s\n", boolToEnabledString(config.EnableApprovalWorkflow))
	fmt.Printf("Reporting System: %s\n", boolToEnabledString(config.EnableReportingSystem))
	fmt.Printf("Sanitization Level: %d\n", config.SanitizationLevel)
	fmt.Printf("Max Prompt Length: %d\n", config.MaxPromptLength)
	fmt.Printf("Approval Threshold: %.2f\n", config.ApprovalThreshold)
	fmt.Printf("Monitoring Interval: %v\n", config.MonitoringInterval)

// protectionLevelToString converts a ProtectionLevel to a string
func protectionLevelToString(level prompt.ProtectionLevel) string {
	switch level {
	case prompt.LevelLow:
		return "Low"
	case prompt.LevelMedium:
		return "Medium"
	case prompt.LevelHigh:
		return "High"
	case prompt.LevelCustom:
		return "Custom"
	default:
		return "Unknown"
	}

// actionTypeToString converts an ActionType to a string
func actionTypeToString(action prompt.ActionType) string {
	switch action {
	case prompt.ActionNone:
		return "None"
	case prompt.ActionModified:
		return "Modified"
	case prompt.ActionWarned:
		return "Warned"
	case prompt.ActionBlocked:
		return "Blocked"
	case prompt.ActionLogged:
		return "Logged"
	case prompt.ActionReported:
		return "Reported"
	default:
		return "Unknown"
	}

// boolToEnabledString converts a bool to an "Enabled" or "Disabled" string
func boolToEnabledString(b bool) string {
	if b {
		return "Enabled"
	}
	return "Disabled"

// listPatterns lists all patterns in the library
func listPatterns(library *prompt.InjectionPatternLibrary) {
	patterns := library.GetAllPatterns()

	if len(patterns) == 0 {
		fmt.Println("No patterns found.")
		return
	}

	fmt.Printf("Found %d patterns:\n\n", len(patterns))

	// Group patterns by category
	patternsByCategory := make(map[prompt.PatternCategory][]*prompt.InjectionPattern)
	for _, pattern := range patterns {
		patternsByCategory[pattern.Category] = append(patternsByCategory[pattern.Category], pattern)
	}

	// Print patterns by category
	for category, patterns := range patternsByCategory {
		fmt.Printf("Category: %s (%d patterns)\n", category, len(patterns))
		fmt.Println(strings.Repeat("-", 50))

		for _, pattern := range patterns {
			status := "Enabled"
			if !pattern.Enabled {
				status = "Disabled"
			}

			fmt.Printf("ID: %s (%s)\n", pattern.ID, status)
			fmt.Printf("Name: %s\n", pattern.Name)
			fmt.Printf("Description: %s\n", pattern.Description)
			fmt.Printf("Pattern: %s\n", pattern.Pattern)
			fmt.Printf("Confidence: %.2f, Severity: %.2f\n", pattern.Confidence, pattern.Severity)

			if len(pattern.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(pattern.Tags, ", "))
			}
			if len(pattern.Examples) > 0 {
				fmt.Println("Examples:")
				for i, example := range pattern.Examples {
					fmt.Printf("  %d. %s\n", i+1, example)
				}
			}

			fmt.Println()
		}

		fmt.Println()
	}

// showReport shows a specific report
func showReport(reportsDir, reportID string) {
	// Look for the report file
	var reportFile string
	err := filepath.Walk(reportsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.Contains(path, reportID) {
			reportFile = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error searching for report: %v\n", err)
		return
	}

	if reportFile == "" {
		fmt.Printf("Report with ID %s not found.\n", reportID)
		return
	}

	// Read the report file
	data, err := os.ReadFile(filepath.Clean(reportFile))
	if err != nil {
		fmt.Printf("Error reading report file: %v\n", err)
		return
	}

	// Parse the report
	var report prompt.InjectionReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		fmt.Printf("Error parsing report: %v\n", err)
		return
	}

	// Print the report
	fmt.Println("Prompt Injection Report")
	fmt.Println("=======================")
	fmt.Printf("Report ID: %s\n", report.ReportID)
	fmt.Printf("Detection Type: %s\n", report.DetectionType)
	fmt.Printf("Confidence: %.2f\n", report.Confidence)
	fmt.Printf("Severity: %.2f\n", report.Severity)
	fmt.Printf("Timestamp: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Description: %s\n", report.Description)

	if report.Pattern != "" {
		fmt.Printf("Pattern: %s\n", report.Pattern)
	}

	if report.Example != "" {
		fmt.Println("\nExample:")
		fmt.Println(report.Example)
	}

	if report.Source != "" {
		fmt.Printf("\nSource: %s\n", report.Source)
	}

	if len(report.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		for key, value := range report.Metadata {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

// listReports lists all reports in the directory
func listReports(reportsDir string) {
	// Find all report files
	var reportFiles []string
	err := filepath.Walk(reportsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(filepath.Base(path), "report_") {
			reportFiles = append(reportFiles, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error searching for reports: %v\n", err)
		return
	}

	if len(reportFiles) == 0 {
		fmt.Println("No reports found.")
		return
	}

	fmt.Printf("Found %d reports:\n\n", len(reportFiles))

	// Process each report file
	for i, file := range reportFiles {
		// Read the report file
		data, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			fmt.Printf("Error reading report file %s: %v\n", file, err)
			continue
		}

		// Parse the report
		var report prompt.InjectionReport
		err = json.Unmarshal(data, &report)
		if err != nil {
			fmt.Printf("Error parsing report from file %s: %v\n", file, err)
			continue
		}

		// Print a summary of the report
		fmt.Printf("%d. Report ID: %s\n", i+1, report.ReportID)
		fmt.Printf("   Type: %s\n", report.DetectionType)
		fmt.Printf("   Confidence: %.2f, Severity: %.2f\n", report.Confidence, report.Severity)
		fmt.Printf("   Timestamp: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Description: %s\n", report.Description)
		fmt.Println()
	}

