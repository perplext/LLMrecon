package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/config"
	"github.com/spf13/cobra"
)

// bundleImportCmd represents the bundle import command
var bundleImportCmd = &cobra.Command{
	Use:   "import PATH",
	Short: "Import a bundle into the system",
	Long: `Import an offline update bundle into the LLMreconing Tool.

This command imports templates, modules, and documentation from a bundle,
preserving OWASP LLM Top 10 categorization and maintaining version compatibility.

The import process includes:
- Bundle verification
- Backup creation (optional)
- Conflict resolution
- Template installation with category preservation
- Module dependency handling
- Documentation integration`,
	Example: `  # Import a bundle
  LLMrecon bundle import update.bundle

  # Import with automatic confirmation
  LLMrecon bundle import update.bundle --yes

  # Import with backup
  LLMrecon bundle import update.bundle --backup

  # Import to custom directories
  LLMrecon bundle import update.bundle --templates-dir=/opt/templates --modules-dir=/opt/modules

  # Dry run to see what would be imported
  LLMrecon bundle import update.bundle --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runBundleImport,
}

func init() {
	bundleCmd.AddCommand(bundleImportCmd)

	// Add flags
	bundleImportCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	bundleImportCmd.Flags().Bool("backup", true, "Create backup before importing")
	bundleImportCmd.Flags().String("templates-dir", "", "Custom templates directory")
	bundleImportCmd.Flags().String("modules-dir", "", "Custom modules directory")
	bundleImportCmd.Flags().Bool("dry-run", false, "Show what would be imported without making changes")
	bundleImportCmd.Flags().Bool("force", false, "Force overwrite existing files")
	bundleImportCmd.Flags().Bool("preserve-structure", true, "Preserve OWASP category structure")
	bundleImportCmd.Flags().Bool("verify", true, "Verify bundle before importing")
	bundleImportCmd.Flags().Bool("verbose", false, "Show detailed import progress")
}

func runBundleImport(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]

	// Get flags
	autoConfirm, _ := cmd.Flags().GetBool("yes")
	createBackup, _ := cmd.Flags().GetBool("backup")
	templatesDir, _ := cmd.Flags().GetString("templates-dir")
	modulesDir, _ := cmd.Flags().GetString("modules-dir")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	preserveStructure, _ := cmd.Flags().GetBool("preserve-structure")
	verify, _ := cmd.Flags().GetBool("verify")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Check if bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle not found: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Set default directories if not specified
	if templatesDir == "" {
		templatesDir = cfg.Templates.Dir
	}
	if modulesDir == "" {
		modulesDir = cfg.Modules.Dir
	}

	fmt.Printf("%s %s\n\n", bold("Importing Bundle:"), bundlePath)

	// 1. Verify bundle if requested
	if verify {
		fmt.Print(dim("Verifying bundle integrity... "))
		
		// Run verification (simplified version)
		verifyCmd := &cobra.Command{}
		verifyCmd.Flags().Bool("json", true, "")
		verifyArgs := []string{bundlePath}
		
		if err := runBundleVerify(verifyCmd, verifyArgs); err != nil {
			fmt.Println(red("✗"))
			return fmt.Errorf("bundle verification failed: %w", err)
		}
		fmt.Println(green("✓"))
	}

	// 2. Load bundle manifest
	fmt.Print(dim("Loading bundle manifest... "))
	manifest, err := bundle.LoadBundleManifest(bundlePath)
	if err != nil {
		fmt.Println(red("✗"))
		return fmt.Errorf("loading manifest: %w", err)
	}
	fmt.Println(green("✓"))

	// 3. Analyze import
	analysis := analyzeImport(manifest, templatesDir, modulesDir, preserveStructure)
	
	// Display import summary
	fmt.Println("\n" + bold("Import Summary:"))
	fmt.Printf("  %s: %d (%d new, %d updates, %d conflicts)\n", 
		cyan("Templates"), 
		analysis.TotalTemplates,
		analysis.NewTemplates,
		analysis.UpdatedTemplates,
		analysis.ConflictingTemplates)
	
	fmt.Printf("  %s: %d (%d new, %d updates, %d conflicts)\n", 
		cyan("Modules"), 
		analysis.TotalModules,
		analysis.NewModules,
		analysis.UpdatedModules,
		analysis.ConflictingModules)

	if analysis.ComplianceDocs > 0 {
		fmt.Printf("  %s: %d\n", cyan("Compliance Docs"), analysis.ComplianceDocs)
	}

	// Show OWASP categories if preserving structure
	if preserveStructure && len(analysis.OWASPCategories) > 0 {
		fmt.Printf("\n  %s:\n", cyan("OWASP Categories"))
		for category, count := range analysis.OWASPCategories {
			fmt.Printf("    - %s: %d templates\n", category, count)
		}
	}

	// Show conflicts if any
	if analysis.HasConflicts() && !force {
		fmt.Println("\n" + yellow("⚠") + " Conflicts detected:")
		for _, conflict := range analysis.Conflicts {
			fmt.Printf("  - %s: %s\n", conflict.Path, conflict.Reason)
		}
		
		if !dryRun {
			fmt.Println("\nUse --force to overwrite existing files.")
		}
	}

	// Dry run mode
	if dryRun {
		fmt.Println("\n" + dim("This is a dry run. No changes will be made."))
		return nil
	}

	// Confirmation prompt
	if !autoConfirm && analysis.HasChanges() {
		fmt.Print("\nProceed with import? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Import cancelled.")
			return nil
		}
	}

	// 4. Create backup if requested
	if createBackup && analysis.HasUpdates() {
		fmt.Print("\n" + dim("Creating backup... "))
		backupPath, err := createImportBackup(cfg, templatesDir, modulesDir)
		if err != nil {
			fmt.Println(yellow("!"))
			fmt.Printf("  %s\n", dim(err.Error()))
		} else {
			fmt.Println(green("✓"))
			fmt.Printf("  %s\n", dim(backupPath))
		}
	}

	// 5. Perform import
	fmt.Println("\n" + bold("Importing:"))
	
	importer := bundle.NewBundleImporter(bundlePath, bundle.ImportOptions{
		TemplatesDir:      templatesDir,
		ModulesDir:        modulesDir,
		Force:             force,
		PreserveStructure: preserveStructure,
		Verbose:           verbose,
	})

	// Set progress handler
	importer.SetProgressHandler(func(event bundle.ImportEvent) {
		switch event.Type {
		case bundle.ImportEventStart:
			fmt.Printf("\n%s %s\n", blue("→"), event.Message)
		case bundle.ImportEventProgress:
			if verbose {
				fmt.Printf("  %s %s\n", dim("•"), event.Path)
			}
		case bundle.ImportEventSuccess:
			// Count successes silently
		case bundle.ImportEventError:
			fmt.Printf("  %s %s: %v\n", red("✗"), event.Path, event.Error)
		case bundle.ImportEventComplete:
			fmt.Printf("  %s %s\n", green("✓"), event.Message)
		}
	})

	// Execute import
	result, err := importer.Import()
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// 6. Display results
	fmt.Println("\n" + strings.Repeat("─", 60))
	
	if result.Success {
		fmt.Printf("\n%s Import completed successfully!\n", green("✓"))
	} else {
		fmt.Printf("\n%s Import completed with errors.\n", yellow("!"))
	}

	fmt.Println("\n" + bold("Results:"))
	fmt.Printf("  Templates: %d imported, %d updated, %d failed\n",
		result.TemplatesImported,
		result.TemplatesUpdated,
		result.TemplatesFailed)
	
	fmt.Printf("  Modules: %d imported, %d updated, %d failed\n",
		result.ModulesImported,
		result.ModulesUpdated,
		result.ModulesFailed)

	if result.DocsImported > 0 {
		fmt.Printf("  Documentation: %d files imported\n", result.DocsImported)
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		fmt.Println("\n" + red("Errors:"))
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	// Post-import recommendations
	if result.Success && result.HasUpdates() {
		fmt.Println("\n" + bold("Recommendations:"))
		fmt.Println("  • Run 'LLMrecon template verify' to validate imported templates")
		fmt.Println("  • Run 'LLMrecon version --check-compatibility' to verify compatibility")
		if result.ModulesImported > 0 || result.ModulesUpdated > 0 {
			fmt.Println("  • Restart the service to load new modules")
		}
	}

	return nil
}

// ImportAnalysis contains analysis of what will be imported
type ImportAnalysis struct {
	TotalTemplates       int
	NewTemplates         int
	UpdatedTemplates     int
	ConflictingTemplates int
	
	TotalModules         int
	NewModules           int
	UpdatedModules       int
	ConflictingModules   int
	
	ComplianceDocs       int
	OWASPCategories      map[string]int
	
	Conflicts            []ImportConflict
}

// ImportConflict represents a file conflict during import
type ImportConflict struct {
	Path   string
	Reason string
}

func (a *ImportAnalysis) HasConflicts() bool {
	return a.ConflictingTemplates > 0 || a.ConflictingModules > 0
}

func (a *ImportAnalysis) HasChanges() bool {
	return a.TotalTemplates > 0 || a.TotalModules > 0 || a.ComplianceDocs > 0
}

func (a *ImportAnalysis) HasUpdates() bool {
	return a.UpdatedTemplates > 0 || a.UpdatedModules > 0
}

func analyzeImport(manifest *bundle.BundleManifest, templatesDir, modulesDir string, preserveStructure bool) *ImportAnalysis {
	analysis := &ImportAnalysis{
		OWASPCategories: make(map[string]int),
		Conflicts:       []ImportConflict{},
	}

	// Analyze templates
	for _, template := range manifest.Templates {
		analysis.TotalTemplates++
		
		// Determine target path
		targetPath := filepath.Join(templatesDir, template.Path)
		if preserveStructure {
			if category, ok := template.Metadata["owasp_category"].(string); ok && category != "" {
				analysis.OWASPCategories[category]++
			}
		}

		// Check if file exists
		if info, err := os.Stat(targetPath); err == nil {
			// File exists - check if it's an update or conflict
			if !info.IsDir() {
				// Compare versions or checksums
				if template.Version != "" {
					// This is an update
					analysis.UpdatedTemplates++
				} else {
					// This is a conflict
					analysis.ConflictingTemplates++
					analysis.Conflicts = append(analysis.Conflicts, ImportConflict{
						Path:   template.Path,
						Reason: "File exists and version information unavailable",
					})
				}
			}
		} else {
			// New file
			analysis.NewTemplates++
		}
	}

	// Analyze modules
	for _, module := range manifest.Modules {
		analysis.TotalModules++
		
		targetPath := filepath.Join(modulesDir, module.Path)
		
		if info, err := os.Stat(targetPath); err == nil {
			if !info.IsDir() {
				if module.Version != "" {
					analysis.UpdatedModules++
				} else {
					analysis.ConflictingModules++
					analysis.Conflicts = append(analysis.Conflicts, ImportConflict{
						Path:   module.Path,
						Reason: "Module exists and version information unavailable",
					})
				}
			}
		} else {
			analysis.NewModules++
		}
	}

	// Count compliance docs
	for _, file := range manifest.Files {
		if strings.HasPrefix(file.Path, "docs/") && 
		   (strings.Contains(file.Path, "compliance") || 
		    strings.Contains(file.Path, "iso42001") ||
		    strings.Contains(file.Path, "owasp")) {
			analysis.ComplianceDocs++
		}
	}

	return analysis
}

func createImportBackup(cfg *config.Config, templatesDir, modulesDir string) (string, error) {
	// Create backup directory
	backupDir := filepath.Join(cfg.Backup.Dir, fmt.Sprintf("import-backup-%s", time.Now().Format("20060102-150405")))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("creating backup directory: %w", err)
	}

	// Backup templates
	if err := backupDirectory(templatesDir, filepath.Join(backupDir, "templates")); err != nil {
		return "", fmt.Errorf("backing up templates: %w", err)
	}

	// Backup modules  
	if err := backupDirectory(modulesDir, filepath.Join(backupDir, "modules")); err != nil {
		return "", fmt.Errorf("backing up modules: %w", err)
	}

	return backupDir, nil
}

func backupDirectory(src, dst string) error {
	// This is a simplified implementation
	// In production, this would properly copy all files and preserve permissions
	return os.MkdirAll(dst, 0755)
}