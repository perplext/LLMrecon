package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
	"github.com/perplext/LLMrecon/src/template/security/sandbox"
)

func main() {
	// Parse command line flags
	templatePath := flag.String("template", "", "Path to the template file")
	mode := flag.String("mode", "strict", "Execution mode (strict, permissive, audit)")
	enableContainer := flag.Bool("container", false, "Use container-based sandbox")
	enableWorkflow := flag.Bool("workflow", false, "Enable approval workflow")
	storageDir := flag.String("storage", "./template_storage", "Storage directory for workflow data")
	logDir := flag.String("logs", "./logs", "Directory for logs")
	validate := flag.Bool("validate", true, "Validate the template")
	execute := flag.Bool("execute", true, "Execute the template")
	user := flag.String("user", "admin", "User for workflow operations")
	dashboard := flag.Bool("dashboard", false, "Start the dashboard server")
	port := flag.Int("port", 8080, "Dashboard server port")
	batch := flag.Bool("batch", false, "Run batch processing of all templates in a directory")
	templateDir := flag.String("template-dir", "./sample_templates", "Directory containing templates for batch processing")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Check if template path is provided when not in dashboard or batch mode
	if *templatePath == "" && !*dashboard && !*batch {
		fmt.Println("Error: Template path is required when not in dashboard or batch mode")
		flag.Usage()
		os.Exit(1)
	}

	// Create the storage directory if needed
	if *enableWorkflow {
		if err := os.MkdirAll(*storageDir, 0755); err != nil {
			fmt.Printf("Error creating storage directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Create the log directory if needed
	if err := os.MkdirAll(*logDir, 0755); err != nil {
		fmt.Printf("Error creating log directory: %v\n", err)
		os.Exit(1)
	}

	// Create framework options
	var execMode sandbox.ExecutionMode
	switch *mode {
	case "strict":
		execMode = sandbox.ModeStrict
	case "permissive":
		execMode = sandbox.ModePermissive
	case "audit":
		execMode = sandbox.ModeAudit
	default:
		fmt.Printf("Invalid mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}

	// Set sandbox options
	sandboxOptions := &sandbox.SandboxOptions{
		Mode: execMode,
		ResourceLimits: sandbox.ResourceLimits{
			MaxCPUTime:       1.0,
			MaxMemory:        100,
			MaxExecutionTime: 5 * time.Second,
			MaxFileSize:      1024 * 1024,
			MaxOpenFiles:     10,
			MaxProcesses:     5,
			NetworkAccess:    false,
			FileSystemAccess: true,
		},
		TimeoutDuration: 10 * time.Second,
	}

	// Set validation options
	validationOptions := &sandbox.ValidationOptions{
		SecurityOptions: &security.VerificationOptions{
			IncludeStandardChecks: true,
			IncludeCustomChecks:   true,
		},
		SyntaxCheck:       true,
		SemanticCheck:     true,
		PerformanceCheck:  true,
		ComplexityCheck:   true,
		CompatibilityCheck: true,
	}

	// Set framework options
	frameworkOptions := &sandbox.FrameworkOptions{
		ValidationOptions:      validationOptions,
		SandboxOptions:         sandboxOptions,
		WorkflowStorageDir:     *storageDir,
		EnableContainerSandbox: *enableContainer,
		EnableLogging:          true,
		LogDirectory:           *logDir,
		EnableMetrics:          true,
	}

	// Create the security framework
	framework, err := sandbox.NewSecurityFramework(frameworkOptions)
	if err != nil {
		fmt.Printf("Error creating security framework: %v\n", err)
		os.Exit(1)
	}

	// Add approvers
	framework.AddApprover("admin")
	framework.AddApprover("security-team")
	
	// Start the dashboard if requested
	if *dashboard {
		fmt.Println("Starting dashboard server...")
		go RunDashboard(framework, *port)
		
		// Keep the main thread alive
		if !*batch && *templatePath == "" {
			fmt.Printf("Dashboard is running at http://localhost:%d\n", *port)
			fmt.Println("Press Ctrl+C to exit")
			select {}
		}
	}

	// Process a single template if path is provided
	if *templatePath != "" {
		// Read the template file
		content, err := ioutil.ReadFile(*templatePath)
		if err != nil {
			fmt.Printf("Error reading template file: %v\n", err)
			os.Exit(1)
		}

		// Parse the template
		template, err := format.ParseTemplate(string(content), filepath.Base(*templatePath))
		if err != nil {
			fmt.Printf("Error parsing template: %v\n", err)
			os.Exit(1)
		}

		// Set template path
		template.Path = *templatePath
		
		// Process the template
		processTemplate(framework, template, *validate, *execute, *enableWorkflow, *user, *verbose)
	}

	// Run batch processing if requested
	if *batch {
		fmt.Println("Running batch processing of templates...")
		processTemplateDirectory(framework, *templateDir, *validate, *execute, *enableWorkflow, *user, *verbose)
	}

	fmt.Println("Done")
}

// processTemplate processes a single template
func processTemplate(framework *sandbox.SecurityFramework, template *format.Template, validate, execute, enableWorkflow bool, user string, verbose bool) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate the template
	if validate {
		if verbose {
			fmt.Printf("Validating template: %s...\n", template.Path)
		}
		result, err := framework.ValidateTemplate(ctx, template)
		if err != nil {
			fmt.Printf("Error validating template: %v\n", err)
			return
		}

		// Print validation result
		fmt.Printf("Validation completed in %s\n", result.ValidationTime)
		fmt.Printf("Risk Score: %.2f (%s)\n", result.RiskScore.Score, result.RiskScore.Category)
		
		// Print issues
		if len(result.Issues) > 0 {
			fmt.Printf("Found %d security issues:\n", len(result.Issues))
			for i, issue := range result.Issues {
				fmt.Printf("%d. [%s] %s (Severity: %s)\n", i+1, issue.Type, issue.Description, issue.Severity)
			}
		} else {
			fmt.Println("No security issues found")
		}

		// Check if template has critical issues
		if result.HasCriticalIssues() {
			fmt.Println("Template has critical security issues. Execution aborted.")
			if !enableWorkflow {
				return
			}
		}
	}

	// Execute the template
	if execute {
		// Check if we should execute based on validation results
		shouldExecute := true
		if validate {
			validationResult, _ := framework.ValidateTemplate(ctx, template)
			shouldExecute = !validationResult.HasCriticalIssues()
		}

		if shouldExecute {
			if verbose {
				fmt.Printf("Executing template: %s...\n", template.Path)
			}
			result, err := framework.ExecuteTemplate(ctx, template)
			if err != nil {
				fmt.Printf("Error executing template: %v\n", err)
				return
			}

			// Print execution result
			fmt.Printf("Execution completed in %s\n", result.ExecutionTime)
			fmt.Printf("Success: %t\n", result.Success)
			if result.Error != "" {
				fmt.Printf("Error: %s\n", result.Error)
			}
			fmt.Printf("Resource Usage: CPU=%.2fs, Memory=%dMB\n", 
				result.ResourceUsage.CPUTime,
				result.ResourceUsage.MemoryUsage)
			
			// Print output
			fmt.Println("Output:")
			fmt.Println(result.Output)
		} else {
			fmt.Println("Template execution skipped due to critical security issues")
		}
	}

	// Handle workflow if enabled
	if enableWorkflow {
		if verbose {
			fmt.Printf("Creating template version for: %s...\n", template.Path)
		}
		version, err := framework.CreateTemplateVersion(ctx, template, user)
		if err != nil {
			fmt.Printf("Error creating template version: %v\n", err)
			return
		}

		fmt.Printf("Created template version: %s (Status: %s)\n", version.ID, version.Status)
		fmt.Printf("Risk Score: %.2f (%s)\n", version.RiskScore.Score, version.RiskScore.Category)

		// Submit for review
		fmt.Println("Submitting template for review...")
		err = framework.SubmitTemplateForReview(template.ID, version.ID, user)
		if err != nil {
			fmt.Printf("Error submitting template for review: %v\n", err)
			return
		}

		fmt.Printf("Template submitted for review\n")

		// Approve the template if user is an approver
		if framework.IsApprover(user) {
			fmt.Println("Approving template...")
			err = framework.ApproveTemplate(template.ID, version.ID, user)
			if err != nil {
				fmt.Printf("Error approving template: %v\n", err)
				return
			}

			fmt.Printf("Template approved\n")
		}

		// Get the latest approved version
		approvedVersion, err := framework.GetLatestApprovedTemplateVersion(template.ID)
		if err != nil {
			fmt.Printf("No approved version found: %v\n", err)
		} else {
			fmt.Printf("Latest approved version: %s (Approved by: %s at %s)\n", 
				approvedVersion.ID, 
				approvedVersion.ApprovedBy, 
				approvedVersion.ApprovedAt.Format(time.RFC3339))
		}
	}

// processTemplateDirectory processes all templates in a directory
func processTemplateDirectory(framework *sandbox.SecurityFramework, templateDir string, validate, execute, enableWorkflow bool, user string, verbose bool) {
	// Check if the directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		fmt.Printf("Error: Template directory does not exist: %s\n", templateDir)
		return
	}

	// Get all template files in the directory
	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		fmt.Printf("Error reading template directory: %v\n", err)
		return
	}

	// Filter for template files
	var templateFiles []string
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".tmpl" || filepath.Ext(file.Name()) == ".template") {
			templatePath := filepath.Join(templateDir, file.Name())
			templateFiles = append(templateFiles, templatePath)
		}
	}

	if len(templateFiles) == 0 {
		fmt.Println("No template files found in the directory")
		return
	}

	fmt.Printf("Found %d template files\n", len(templateFiles))

	// Process each template file
	var wg sync.WaitGroup
	for _, templatePath := range templateFiles {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			// Read the template file
			content, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading template file %s: %v\n", path, err)
				return
			}

			// Parse the template
			template, err := format.ParseTemplate(string(content), filepath.Base(path))
			if err != nil {
				fmt.Printf("Error parsing template %s: %v\n", path, err)
				return
			}

			// Set template path
			template.Path = path

			// Process the template
			fmt.Printf("\n--- Processing template: %s ---\n", path)
			processTemplate(framework, template, validate, execute, enableWorkflow, user, verbose)
		}(templatePath)
	}

	// Wait for all templates to be processed
	wg.Wait()
	fmt.Println("Batch processing completed")
}
