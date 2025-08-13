package cmd

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/manifest"
	"github.com/spf13/cobra"
)

var (
	templateName        string
	templateDescription string
	templateVersion     string
	templateAuthor      string
	templateCategory    string
	templateSeverity    string
	templateTags        []string

	moduleName        string
	moduleDescription string
	moduleVersion     string
	moduleAuthor      string
	moduleType        string
	moduleTags        []string
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage vulnerability templates",
	Long:  `Create, list, and manage vulnerability templates for the LLMreconing Tool.`,
}

// templateListCmd represents the template list command
var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	Long:  `List all available vulnerability templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get base directory
		baseDir, err := getBaseDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Create manifest manager
		manager := manifest.NewManager(baseDir)

		// Load manifests
		if err := manager.LoadManifests(); err != nil {
			fmt.Printf("Error loading manifests: %v\n", err)
			return
		}

		// Get template manifest
		templateManifest := manager.GetTemplateManifest()

		// Print templates
		fmt.Println("Available Templates:")
		fmt.Println("====================")

		if len(templateManifest.Templates) == 0 {
			fmt.Println("No templates found.")
			return
		}

		// Print by category
		for category, categoryInfo := range templateManifest.Categories {
			fmt.Printf("\nCategory: %s\n", category)
			fmt.Printf("Description: %s\n", categoryInfo.Description)
			fmt.Println("Templates:")

			for _, id := range categoryInfo.Templates {
				if template, exists := templateManifest.Templates[id]; exists {
					fmt.Printf("  - %s (v%s): %s\n", template.Name, template.Version, template.Description)
				}
			}
		}
	},
}

// templateCreateCmd represents the template create command
var templateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new template",
	Long:  `Create a new vulnerability template.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required flags
		if templateName == "" || templateDescription == "" || templateVersion == "" ||
			templateAuthor == "" || templateCategory == "" || templateSeverity == "" {
			fmt.Println("Error: All required flags must be specified.")
			fmt.Println("Required flags: --name, --description, --version, --author, --category, --severity")
			return
		}

		// Get base directory
		baseDir, err := getBaseDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Create template
		template := format.NewTemplate()
		template.Info = format.TemplateInfo{
			Name:        templateName,
			Description: templateDescription,
			Version:     templateVersion,
			Author:      templateAuthor,
			Severity:    templateSeverity,
			Tags:        templateTags,
		}

		// Generate ID
		template.ID = fmt.Sprintf("%s_%s_v%s", templateCategory, format.SanitizeFilename(templateName), templateVersion)

		// Create test definition with placeholder
		template.Test = format.TestDefinition{
			Prompt:           "Add your test prompt here",
			ExpectedBehavior: "Describe the expected behavior here",
			Detection: format.DetectionCriteria{
				Type:      "string_match",
				Match:     "Add detection string here",
				Condition: "contains",
			},
		}

		// Create manifest manager
		manager := manifest.NewManager(baseDir)

		// Load manifests
		if err := manager.LoadManifests(); err != nil {
			fmt.Printf("Error loading manifests: %v\n", err)
			return
		}

		// Create category directory if it doesn't exist
		categoryDir := filepath.Join(baseDir, "templates", templateCategory)
		if err := os.MkdirAll(categoryDir, 0755); err != nil {
			fmt.Printf("Error creating category directory: %v\n", err)
			return
		}

		// Save template to file
		templatePath := filepath.Join(categoryDir, fmt.Sprintf("%s_v%s.yaml", format.SanitizeFilename(templateName), templateVersion))
		if err := template.SaveToFile(templatePath); err != nil {
			fmt.Printf("Error saving template: %v\n", err)
			return
		}

		// Register template in manifest
		if err := manager.RegisterTemplate(template); err != nil {
			fmt.Printf("Error registering template: %v\n", err)
			return
		}

		// Save manifests
		if err := manager.SaveManifests(); err != nil {
			fmt.Printf("Error saving manifests: %v\n", err)
			return
		}

		fmt.Printf("Template '%s' created successfully at %s\n", templateName, templatePath)
		fmt.Println("Remember to edit the template to add your test prompt and detection criteria.")
	},
}

// moduleCmd represents the module command
var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Manage modules",
	Long:  `Create, list, and manage modules for the LLMreconing Tool.`,
}

// moduleListCmd represents the module list command
var moduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all modules",
	Long:  `List all available modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get base directory
		baseDir, err := getBaseDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Create manifest manager
		manager := manifest.NewManager(baseDir)

		// Load manifests
		if err := manager.LoadManifests(); err != nil {
			fmt.Printf("Error loading manifests: %v\n", err)
			return
		}

		// Get module manifest
		moduleManifest := manager.GetModuleManifest()

		// Print modules
		fmt.Println("Available Modules:")
		fmt.Println("==================")

		if len(moduleManifest.Modules) == 0 {
			fmt.Println("No modules found.")
			return
		}

		// Print by type
		for moduleType, typeInfo := range moduleManifest.Types {
			fmt.Printf("\nType: %s\n", moduleType)
			fmt.Printf("Description: %s\n", typeInfo.Description)
			fmt.Println("Modules:")

			for _, id := range typeInfo.Modules {
				if module, exists := moduleManifest.Modules[id]; exists {
					fmt.Printf("  - %s (v%s): %s\n", module.Name, module.Version, module.Description)
				}
			}
		}
	},
}

// moduleCreateCmd represents the module create command
var moduleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new module",
	Long:  `Create a new module (provider, utility, or detector).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required flags
		if moduleName == "" || moduleDescription == "" || moduleVersion == "" ||
			moduleAuthor == "" || moduleType == "" {
			fmt.Println("Error: All required flags must be specified.")
			fmt.Println("Required flags: --name, --description, --version, --author, --type")
			return
		}

		// Validate module type
		var moduleTypeEnum format.ModuleType
		switch moduleType {
		case "provider":
			moduleTypeEnum = format.ProviderModule
		case "utility":
			moduleTypeEnum = format.UtilityModule
		case "detector":
			moduleTypeEnum = format.DetectorModule
		default:
			fmt.Printf("Error: Invalid module type '%s'. Valid types: provider, utility, detector\n", moduleType)
			return
		}

		// Get base directory
		baseDir, err := getBaseDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Create module
		module := format.NewModule(moduleTypeEnum)
		module.Info = format.ModuleInfo{
			Name:        moduleName,
			Description: moduleDescription,
			Version:     moduleVersion,
			Author:      moduleAuthor,
			Tags:        moduleTags,
		}

		// Generate ID
		module.ID = fmt.Sprintf("%s_%s_v%s", format.SanitizeFilename(moduleName), moduleType, moduleVersion)

		// Create manifest manager
		manager := manifest.NewManager(baseDir)

		// Load manifests
		if err := manager.LoadManifests(); err != nil {
			fmt.Printf("Error loading manifests: %v\n", err)
			return
		}

		// Determine module directory
		var moduleDir string
		switch moduleTypeEnum {
		case format.ProviderModule:
			moduleDir = filepath.Join(baseDir, "modules", "providers")
		case format.UtilityModule:
			moduleDir = filepath.Join(baseDir, "modules", "utils")
		case format.DetectorModule:
			moduleDir = filepath.Join(baseDir, "modules", "detectors")
		}

		// Create module directory if it doesn't exist
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			fmt.Printf("Error creating module directory: %v\n", err)
			return
		}

		// Save module to file
		modulePath := filepath.Join(moduleDir, fmt.Sprintf("%s_v%s.yaml", format.SanitizeFilename(moduleName), moduleVersion))
		if err := module.SaveToFile(modulePath); err != nil {
			fmt.Printf("Error saving module: %v\n", err)
			return
		}

		// Register module in manifest
		if err := manager.RegisterModule(module); err != nil {
			fmt.Printf("Error registering module: %v\n", err)
			return
		}

		// Save manifests
		if err := manager.SaveManifests(); err != nil {
			fmt.Printf("Error saving manifests: %v\n", err)
			return
		}

		fmt.Printf("Module '%s' created successfully at %s\n", moduleName, modulePath)
		fmt.Println("Remember to edit the module to add your specific configuration.")
	},
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan and update manifests",
	Long:  `Scan templates and modules directories and update manifests.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get base directory
		baseDir, err := getBaseDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Create manifest manager
		manager := manifest.NewManager(baseDir)

		// Scan and register templates
		fmt.Println("Scanning templates...")
		if err := manager.ScanAndRegisterTemplates(); err != nil {
			fmt.Printf("Error scanning templates: %v\n", err)
			return
		}

		// Scan and register modules
		fmt.Println("Scanning modules...")
		if err := manager.ScanAndRegisterModules(); err != nil {
			fmt.Printf("Error scanning modules: %v\n", err)
			return
		}

		// Save manifests
		if err := manager.SaveManifests(); err != nil {
			fmt.Printf("Error saving manifests: %v\n", err)
			return
		}

		// Get counts
		templateManifest := manager.GetTemplateManifest()
		moduleManifest := manager.GetModuleManifest()

		fmt.Printf("Scan complete. Found %d templates and %d modules.\n",
			len(templateManifest.Templates), len(moduleManifest.Modules))
	},
}

// getBaseDir returns the base directory for the LLMreconing Tool
func getBaseDir() (string, error) {
	// For now, use the current working directory
	// In a real implementation, this would use a configuration file or environment variable
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return cwd, nil
}

func init() {
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(moduleCmd)
	rootCmd.AddCommand(scanCmd)

	// Add subcommands to template command
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateCreateCmd)

	// Add flags to template create command
	templateCreateCmd.Flags().StringVar(&templateName, "name", "", "Template name")
	templateCreateCmd.Flags().StringVar(&templateDescription, "description", "", "Template description")
	templateCreateCmd.Flags().StringVar(&templateVersion, "version", "", "Template version")
	templateCreateCmd.Flags().StringVar(&templateAuthor, "author", "", "Template author")
	templateCreateCmd.Flags().StringVar(&templateCategory, "category", "", "Template category")
	templateCreateCmd.Flags().StringVar(&templateSeverity, "severity", "", "Template severity (low, medium, high, critical)")
	templateCreateCmd.Flags().StringSliceVar(&templateTags, "tags", []string{}, "Template tags (comma-separated)")

	// Add subcommands to module command
	moduleCmd.AddCommand(moduleListCmd)
	moduleCmd.AddCommand(moduleCreateCmd)

	// Add flags to module create command
	moduleCreateCmd.Flags().StringVar(&moduleName, "name", "", "Module name")
	moduleCreateCmd.Flags().StringVar(&moduleDescription, "description", "", "Module description")
	moduleCreateCmd.Flags().StringVar(&moduleVersion, "version", "", "Module version")
	moduleCreateCmd.Flags().StringVar(&moduleAuthor, "author", "", "Module author")
	moduleCreateCmd.Flags().StringVar(&moduleType, "type", "", "Module type (provider, utility, detector)")
	moduleCreateCmd.Flags().StringSliceVar(&moduleTags, "tags", []string{}, "Module tags (comma-separated)")
}
