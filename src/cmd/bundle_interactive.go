package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var bundleInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive bundle management wizard",
	Long:  `Launch an interactive wizard for bundle operations with guided prompts and validation`,
	RunE:  runBundleInteractive,
}

func init() {
	bundleCmd.AddCommand(bundleInteractiveCmd)
}

func runBundleInteractive(cmd *cobra.Command, args []string) error {
	// Welcome message
	fmt.Println()
	color.Cyan("🎯 LLMrecon Bundle Management Interactive Wizard")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()

	// Main menu
	var operation string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Create a new bundle",
			"Verify an existing bundle",
			"Import a bundle",
			"View bundle information",
			"Generate compliance report",
			"Export templates by category",
			"Exit",
		},
	}

	if err := survey.AskOne(prompt, &operation); err != nil {
		return fmt.Errorf("operation selection failed: %w", err)
	}

	switch operation {
	case "Create a new bundle":
		return interactiveCreateBundle()
	case "Verify an existing bundle":
		return interactiveVerifyBundle()
	case "Import a bundle":
		return interactiveImportBundle()
	case "View bundle information":
		return interactiveViewBundleInfo()
	case "Generate compliance report":
		return interactiveGenerateReport()
	case "Export templates by category":
		return interactiveExportByCategory()
	case "Exit":
		color.Green("👋 Thank you for using the Bundle Management Wizard!")
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

func interactiveCreateBundle() error {
	fmt.Println()
	color.Yellow("📦 Bundle Creation Wizard")
	fmt.Println()

	// Bundle name
	var bundleName string
	namePrompt := &survey.Input{
		Message: "Enter bundle name:",
		Default: "security-bundle",
	}
	if err := survey.AskOne(namePrompt, &bundleName, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Source selection
	var source string
	sourcePrompt := &survey.Select{
		Message: "Select template source:",
		Options: []string{
			"Local directory",
			"GitHub repository",
			"GitLab repository",
			"Multiple sources",
		},
	}
	if err := survey.AskOne(sourcePrompt, &source); err != nil {
		return err
	}

	var sources []string
	switch source {
	case "Local directory":
		var localPath string
		pathPrompt := &survey.Input{
			Message: "Enter local directory path:",
			Default: "./templates",
		}
		if err := survey.AskOne(pathPrompt, &localPath); err != nil {
			return err
		}
		sources = append(sources, localPath)

	case "GitHub repository":
		var repoURL string
		repoPrompt := &survey.Input{
			Message: "Enter GitHub repository URL:",
			Help:    "Example: https://github.com/owner/repo",
		}
		if err := survey.AskOne(repoPrompt, &repoURL, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		sources = append(sources, repoURL)

	case "GitLab repository":
		var repoURL string
		repoPrompt := &survey.Input{
			Message: "Enter GitLab repository URL:",
			Help:    "Example: https://gitlab.com/owner/repo",
		}
		if err := survey.AskOne(repoPrompt, &repoURL, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		sources = append(sources, repoURL)

	case "Multiple sources":
		fmt.Println("Enter sources (one per line, empty line to finish):")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}
			sources = append(sources, line)
		}
	}

	// OWASP categories selection
	var selectedCategories []string
	categoryPrompt := &survey.MultiSelect{
		Message: "Select OWASP LLM Top 10 categories to include:",
		Options: []string{
			"llm01-prompt-injection",
			"llm02-insecure-output",
			"llm03-training-data-poisoning",
			"llm04-model-denial-of-service",
			"llm05-supply-chain",
			"llm06-sensitive-information",
			"llm07-insecure-plugin",
			"llm08-excessive-agency",
			"llm09-overreliance",
			"llm10-model-theft",
			"All categories",
		},
	}
	if err := survey.AskOne(categoryPrompt, &selectedCategories); err != nil {
		return err
	}

	// Check if "All categories" was selected
	includeAll := false
	for _, cat := range selectedCategories {
		if cat == "All categories" {
			includeAll = true
			break
		}
	}

	// Bundle options
	var compress, encrypt, sign bool
	
	compressPrompt := &survey.Confirm{
		Message: "Compress bundle?",
		Default: true,
	}
	if err := survey.AskOne(compressPrompt, &compress); err != nil {
		return err
	}

	encryptPrompt := &survey.Confirm{
		Message: "Encrypt bundle?",
		Default: false,
	}
	if err := survey.AskOne(encryptPrompt, &encrypt); err != nil {
		return err
	}

	signPrompt := &survey.Confirm{
		Message: "Sign bundle?",
		Default: true,
	}
	if err := survey.AskOne(signPrompt, &sign); err != nil {
		return err
	}

	// Build command arguments
	args := []string{"bundle", "create", bundleName}
	for _, src := range sources {
		args = append(args, "--source", src)
	}
	if !includeAll {
		for _, cat := range selectedCategories {
			args = append(args, "--category", cat)
		}
	}
	if compress {
		args = append(args, "--compress")
	}
	if encrypt {
		args = append(args, "--encrypt")
	}
	if sign {
		args = append(args, "--sign")
	}

	// Show preview
	fmt.Println()
	color.Cyan("Command to execute:")
	fmt.Printf("LLMrecon %s\n", strings.Join(args, " "))
	fmt.Println()

	var proceed bool
	proceedPrompt := &survey.Confirm{
		Message: "Proceed with bundle creation?",
		Default: true,
	}
	if err := survey.AskOne(proceedPrompt, &proceed); err != nil {
		return err
	}

	if !proceed {
		color.Yellow("Bundle creation cancelled")
		return nil
	}

	// Execute the command
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func interactiveVerifyBundle() error {
	fmt.Println()
	color.Yellow("🔍 Bundle Verification Wizard")
	fmt.Println()

	// Bundle file selection
	var bundlePath string
	pathPrompt := &survey.Input{
		Message: "Enter bundle file path:",
		Help:    "Path to .bundle file",
	}
	if err := survey.AskOne(pathPrompt, &bundlePath, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Verification options
	var skipSignature, skipIntegrity, verbose bool
	
	skipSigPrompt := &survey.Confirm{
		Message: "Skip signature verification?",
		Default: false,
	}
	if err := survey.AskOne(skipSigPrompt, &skipSignature); err != nil {
		return err
	}

	skipIntPrompt := &survey.Confirm{
		Message: "Skip integrity verification?",
		Default: false,
	}
	if err := survey.AskOne(skipIntPrompt, &skipIntegrity); err != nil {
		return err
	}

	verbosePrompt := &survey.Confirm{
		Message: "Show detailed output?",
		Default: true,
	}
	if err := survey.AskOne(verbosePrompt, &verbose); err != nil {
		return err
	}

	// Build command
	args := []string{"bundle", "verify", bundlePath}
	if skipSignature {
		args = append(args, "--skip-signature")
	}
	if skipIntegrity {
		args = append(args, "--skip-integrity")
	}
	if verbose {
		args = append(args, "--verbose")
	}

	// Execute
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func interactiveImportBundle() error {
	fmt.Println()
	color.Yellow("📥 Bundle Import Wizard")
	fmt.Println()

	// Bundle file
	var bundlePath string
	pathPrompt := &survey.Input{
		Message: "Enter bundle file path:",
		Help:    "Path to .bundle file to import",
	}
	if err := survey.AskOne(pathPrompt, &bundlePath, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Target directory
	var targetDir string
	targetPrompt := &survey.Input{
		Message: "Enter target directory:",
		Default: "./imported-templates",
	}
	if err := survey.AskOne(targetPrompt, &targetDir); err != nil {
		return err
	}

	// Import options
	var dryRun, force, preserveOWASP bool
	
	dryRunPrompt := &survey.Confirm{
		Message: "Perform dry run first?",
		Default: true,
	}
	if err := survey.AskOne(dryRunPrompt, &dryRun); err != nil {
		return err
	}

	if dryRun {
		// Execute dry run
		args := []string{"bundle", "import", bundlePath, targetDir, "--dry-run"}
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			return err
		}

		// Ask if they want to proceed with actual import
		var proceedImport bool
		proceedPrompt := &survey.Confirm{
			Message: "Proceed with actual import?",
			Default: true,
		}
		if err := survey.AskOne(proceedPrompt, &proceedImport); err != nil {
			return err
		}
		if !proceedImport {
			color.Yellow("Import cancelled")
			return nil
		}
	}

	forcePrompt := &survey.Confirm{
		Message: "Force overwrite existing templates?",
		Default: false,
	}
	if err := survey.AskOne(forcePrompt, &force); err != nil {
		return err
	}

	preservePrompt := &survey.Confirm{
		Message: "Preserve OWASP category structure?",
		Default: true,
	}
	if err := survey.AskOne(preservePrompt, &preserveOWASP); err != nil {
		return err
	}

	// Build final command
	args := []string{"bundle", "import", bundlePath, targetDir}
	if force {
		args = append(args, "--force")
	}
	if preserveOWASP {
		args = append(args, "--preserve-owasp-structure")
	}

	// Execute
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func interactiveViewBundleInfo() error {
	fmt.Println()
	color.Yellow("ℹ️  Bundle Information Viewer")
	fmt.Println()

	// Bundle file
	var bundlePath string
	pathPrompt := &survey.Input{
		Message: "Enter bundle file path:",
		Help:    "Path to .bundle file",
	}
	if err := survey.AskOne(pathPrompt, &bundlePath, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Display options
	var format string
	formatPrompt := &survey.Select{
		Message: "Select output format:",
		Options: []string{
			"Summary",
			"Detailed",
			"JSON",
		},
		Default: "Detailed",
	}
	if err := survey.AskOne(formatPrompt, &format); err != nil {
		return err
	}

	// Build command
	args := []string{"bundle", "info", bundlePath}
	switch format {
	case "Summary":
		args = append(args, "--summary")
	case "JSON":
		args = append(args, "--json")
	}

	// Execute
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func interactiveGenerateReport() error {
	fmt.Println()
	color.Yellow("📊 Compliance Report Generator")
	fmt.Println()

	// Report type
	var reportType string
	typePrompt := &survey.Select{
		Message: "Select report type:",
		Options: []string{
			"OWASP LLM Top 10 Compliance",
			"ISO/IEC 42001 Compliance",
			"Full Security Assessment",
			"Template Coverage Report",
			"Vulnerability Summary",
		},
	}
	if err := survey.AskOne(typePrompt, &reportType); err != nil {
		return err
	}

	// Input source
	var source string
	sourcePrompt := &survey.Select{
		Message: "Select input source:",
		Options: []string{
			"Bundle file",
			"Template directory",
			"Scan results",
		},
	}
	if err := survey.AskOne(sourcePrompt, &source); err != nil {
		return err
	}

	var inputPath string
	pathPrompt := &survey.Input{
		Message: fmt.Sprintf("Enter %s path:", strings.ToLower(source)),
	}
	if err := survey.AskOne(pathPrompt, &inputPath, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Output format
	var formats []string
	formatPrompt := &survey.MultiSelect{
		Message: "Select output formats:",
		Options: []string{
			"HTML",
			"PDF",
			"JSON",
			"Markdown",
			"CSV",
		},
		Default: []string{"HTML", "JSON"},
	}
	if err := survey.AskOne(formatPrompt, &formats); err != nil {
		return err
	}

	// Output directory
	var outputDir string
	outPrompt := &survey.Input{
		Message: "Enter output directory:",
		Default: "./reports",
	}
	if err := survey.AskOne(outPrompt, &outputDir); err != nil {
		return err
	}

	fmt.Println()
	color.Green("✅ Generating compliance report...")
	color.Cyan("Report Type: %s", reportType)
	color.Cyan("Input: %s", inputPath)
	color.Cyan("Output Formats: %s", strings.Join(formats, ", "))
	color.Cyan("Output Directory: %s", outputDir)
	
	// TODO: Execute actual report generation when reporting commands are implemented
	
	return nil
}

func interactiveExportByCategory() error {
	fmt.Println()
	color.Yellow("🗂️  Export Templates by Category")
	fmt.Println()

	// Source selection
	var sourcePath string
	sourcePrompt := &survey.Input{
		Message: "Enter template directory or bundle file:",
		Default: "./templates",
	}
	if err := survey.AskOne(sourcePrompt, &sourcePath, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Category selection
	var categories []string
	categoryPrompt := &survey.MultiSelect{
		Message: "Select categories to export:",
		Options: []string{
			"llm01-prompt-injection",
			"llm02-insecure-output",
			"llm03-training-data-poisoning",
			"llm04-model-denial-of-service",
			"llm05-supply-chain",
			"llm06-sensitive-information",
			"llm07-insecure-plugin",
			"llm08-excessive-agency",
			"llm09-overreliance",
			"llm10-model-theft",
		},
	}
	if err := survey.AskOne(categoryPrompt, &categories); err != nil {
		return err
	}

	// Export format
	var exportFormat string
	formatPrompt := &survey.Select{
		Message: "Select export format:",
		Options: []string{
			"Separate bundles per category",
			"Single bundle with categories",
			"Directory structure",
		},
		Default: "Separate bundles per category",
	}
	if err := survey.AskOne(formatPrompt, &exportFormat); err != nil {
		return err
	}

	// Output directory
	var outputDir string
	outPrompt := &survey.Input{
		Message: "Enter output directory:",
		Default: "./exports",
	}
	if err := survey.AskOne(outPrompt, &outputDir); err != nil {
		return err
	}

	fmt.Println()
	color.Green("✅ Exporting templates by category...")
	color.Cyan("Source: %s", sourcePath)
	color.Cyan("Categories: %s", strings.Join(categories, ", "))
	color.Cyan("Format: %s", exportFormat)
	color.Cyan("Output: %s", outputDir)

	// TODO: Execute actual export when category export is implemented

	return nil
}