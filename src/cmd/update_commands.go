package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/perplext/LLMrecon/src/update"
)

var (
	// Update flags
	updateCheck          bool
	updateForce          bool
	updateYes            bool
	updateTemplatesOnly  bool
	updateModulesOnly    bool
	updateBinaryOnly     bool
	updateTimeout        time.Duration
	updateSkipVerify     bool
	updateSkipBackup     bool
	updatePrerelease     bool
	updateVerbose        bool
	
	// Bundle flags
	bundleOutput         string
	bundleDescription    string
	bundleIncludeBinary  bool
	bundleIncludeTemplates bool
	bundleIncludeModules bool
	bundleSign           bool
	bundleExpires        string
	bundleForceImport    bool
	bundleVerifyIntegrity bool
	bundleCreateBackup   bool
	bundleDryRun         bool
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the tool and its components",
	Long: `Update the LLMrecon tool, templates, and modules to the latest versions.

This command can update:
- The main binary (self-update)
- Vulnerability templates from repositories
- Provider modules from repositories

The update process includes verification of downloads and automatic backup creation.`,
	Example: `  # Update everything
  LLMrecon update

  # Check for updates without applying them
  LLMrecon update --check

  # Update only templates
  LLMrecon update --templates-only

  # Force update without prompts
  LLMrecon update --yes

  # Update with verbose output
  LLMrecon update --verbose`,
	RunE: runUpdate,
}

// versionCmd shows current version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display detailed version information for the tool and its components.`,
	Example: `  # Show version information
  LLMrecon version

  # Show detailed component versions
  LLMrecon version --verbose`,
	RunE: runVersion,
}

// bundleExportCmd exports an offline update bundle
var bundleExportCmd = &cobra.Command{
	Use:   "bundle-export",
	Short: "Export an offline update bundle",
	Long: `Create an offline update bundle that can be transferred to air-gapped systems.

The bundle contains templates, modules, and optionally the binary itself,
packaged in a verified archive that can be imported on offline systems.`,
	Example: `  # Export full bundle
  LLMrecon bundle-export --output update-bundle.zip

  # Export only templates
  LLMrecon bundle-export --templates-only --output templates-bundle.zip

  # Export signed bundle with expiration
  LLMrecon bundle-export --sign --expires 30d --output signed-bundle.zip`,
	RunE: runBundleExport,
}

// bundleImportCmd imports an offline update bundle
var bundleImportCmd = &cobra.Command{
	Use:   "bundle-import <bundle-file>",
	Short: "Import an offline update bundle",
	Long: `Import an offline update bundle to update the tool in air-gapped environments.

The bundle is verified for integrity and compatibility before applying updates.`,
	Example: `  # Import bundle
  LLMrecon bundle-import update-bundle.zip

  # Import with backup creation
  LLMrecon bundle-import --backup update-bundle.zip

  # Force import (skip verification)
  LLMrecon bundle-import --force update-bundle.zip

  # Dry run (preview changes)
  LLMrecon bundle-import --dry-run update-bundle.zip`,
	Args: cobra.ExactArgs(1),
	RunE: runBundleImport,
}

func init() {
	// Add update command and subcommands
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(bundleExportCmd)
	rootCmd.AddCommand(bundleImportCmd)
	
	// Update command flags
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Check for updates without applying them")
	updateCmd.Flags().BoolVar(&updateForce, "force", false, "Force update even if current version is newer")
	updateCmd.Flags().BoolVar(&updateYes, "yes", false, "Automatically confirm all prompts")
	updateCmd.Flags().BoolVar(&updateTemplatesOnly, "templates-only", false, "Update only templates")
	updateCmd.Flags().BoolVar(&updateModulesOnly, "modules-only", false, "Update only modules")
	updateCmd.Flags().BoolVar(&updateBinaryOnly, "binary-only", false, "Update only the binary")
	updateCmd.Flags().DurationVar(&updateTimeout, "timeout", 5*time.Minute, "Timeout for update operations")
	updateCmd.Flags().BoolVar(&updateSkipVerify, "skip-verify", false, "Skip cryptographic verification")
	updateCmd.Flags().BoolVar(&updateSkipBackup, "skip-backup", false, "Skip backup creation")
	updateCmd.Flags().BoolVar(&updatePrerelease, "prerelease", false, "Include prerelease versions")
	updateCmd.Flags().BoolVar(&updateVerbose, "verbose", false, "Verbose output")
	
	// Version command flags
	versionCmd.Flags().BoolVar(&updateVerbose, "verbose", false, "Show detailed version information")
	
	// Bundle export flags
	bundleExportCmd.Flags().StringVar(&bundleOutput, "output", "", "Output bundle file path (required)")
	bundleExportCmd.Flags().StringVar(&bundleDescription, "description", "", "Bundle description")
	bundleExportCmd.Flags().BoolVar(&bundleIncludeBinary, "binary", true, "Include binary in bundle")
	bundleExportCmd.Flags().BoolVar(&bundleIncludeTemplates, "templates", true, "Include templates in bundle")
	bundleExportCmd.Flags().BoolVar(&bundleIncludeModules, "modules", true, "Include modules in bundle")
	bundleExportCmd.Flags().BoolVar(&bundleSign, "sign", false, "Sign the bundle")
	bundleExportCmd.Flags().StringVar(&bundleExpires, "expires", "", "Bundle expiration (e.g., 30d, 1m, 1y)")
	bundleExportCmd.MarkFlagRequired("output")
	
	// Bundle import flags
	bundleImportCmd.Flags().BoolVar(&bundleForceImport, "force", false, "Force import (skip verification)")
	bundleImportCmd.Flags().BoolVar(&bundleVerifyIntegrity, "verify", true, "Verify bundle integrity")
	bundleImportCmd.Flags().BoolVar(&bundleCreateBackup, "backup", true, "Create backup before import")
	bundleImportCmd.Flags().BoolVar(&bundleDryRun, "dry-run", false, "Preview changes without applying")
}

// runUpdate executes the update command
func runUpdate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()
	
	// Initialize update manager
	config := createUpdateConfig()
	logger := &CLILogger{verbose: updateVerbose}
	manager := update.NewManager(config, logger)
	
	// Check for updates
	fmt.Println("Checking for updates...")
	updateCheck, err := manager.CheckForUpdates(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	
	if !updateCheck.UpdatesAvailable {
		fmt.Println("✅ All components are up to date")
		return nil
	}
	
	// Display available updates
	fmt.Println("\n📦 Available updates:")
	for component, update := range updateCheck.Components {
		if update.Available {
			icon := "🔄"
			if update.SecurityUpdate {
				icon = "🔒"
			}
			if update.Critical {
				icon = "⚠️"
			}
			
			fmt.Printf("  %s %s: %s → %s", icon, component, update.CurrentVersion, update.LatestVersion)
			if update.UpdateSize > 0 {
				fmt.Printf(" (%s)", update.FormatFileSize(update.UpdateSize))
			}
			fmt.Println()
			
			if update.SecurityUpdate {
				fmt.Println("    🔒 Security update")
			}
			if update.Critical {
				fmt.Println("    ⚠️  Critical update")
			}
		}
	}
	
	// If only checking, return here
	if updateCheck {
		return nil
	}
	
	// Confirm update
	if !updateYes {
		if !confirmUpdate() {
			fmt.Println("Update cancelled")
			return nil
		}
	}
	
	// Determine which components to update
	components := getUpdateComponents()
	
	// Apply updates
	fmt.Println("\n🚀 Starting update process...")
	summary, err := manager.ApplyUpdates(ctx, components)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	
	// Display results
	displayUpdateResults(summary)
	
	// Restart prompt if needed
	if summary.RestartRequired {
		fmt.Println("\n⚠️  Restart required to complete the update")
		if updateYes || confirmRestart() {
			fmt.Println("Restarting...")
			// Implementation would restart the application
		}
	}
	
	return nil
}

// runVersion executes the version command
func runVersion(cmd *cobra.Command, args []string) error {
	config := createUpdateConfig()
	logger := &CLILogger{verbose: updateVerbose}
	manager := update.NewManager(config, logger)
	
	// Get version information
	versionInfo := getVersionInfo()
	
	// Basic version info
	fmt.Printf("LLMrecon %s\n", versionInfo.Version)
	fmt.Printf("Build Date: %s\n", versionInfo.BuildDate.Format("2006-01-02 15:04:05"))
	fmt.Printf("Platform: %s/%s\n", versionInfo.Platform, versionInfo.Arch)
	
	if updateVerbose {
		fmt.Printf("Commit: %s\n", versionInfo.Commit)
		fmt.Printf("Branch: %s\n", versionInfo.Branch)
		fmt.Printf("Go Version: %s\n", versionInfo.GoVersion)
		
		if len(versionInfo.Components) > 0 {
			fmt.Println("\nComponent Versions:")
			for name, version := range versionInfo.Components {
				fmt.Printf("  %s: %s\n", name, version)
			}
		}
		
		// Check for updates
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		fmt.Println("\nChecking for updates...")
		if updateCheck, err := manager.CheckForUpdates(ctx); err == nil {
			if updateCheck.UpdatesAvailable {
				fmt.Println("📦 Updates available! Run 'LLMrecon update' to update.")
			} else {
				fmt.Println("✅ All components are up to date")
			}
		}
	}
	
	return nil
}

// runBundleExport executes the bundle export command
func runBundleExport(cmd *cobra.Command, args []string) error {
	// Parse expiration
	var expiresAt *time.Time
	if bundleExpires != "" {
		exp, err := parseExpiration(bundleExpires)
		if err != nil {
			return fmt.Errorf("invalid expiration format: %w", err)
		}
		expiresAt = &exp
	}
	
	// Create export options
	options := &update.BundleExportOptions{
		Version:          "1.0",
		Type:             determineBundleType(),
		Description:      bundleDescription,
		SourceVersion:    getCurrentVersion(),
		TargetVersion:    getCurrentVersion(),
		Platforms:        []string{getPlatformString()},
		OutputPath:       bundleOutput,
		CreatedBy:        "LLMrecon-cli",
		ExpiresAt:        expiresAt,
		IncludeBinary:    bundleIncludeBinary,
		IncludeTemplates: bundleIncludeTemplates,
		IncludeModules:   bundleIncludeModules,
		SignBundle:       bundleSign,
	}
	
	// Initialize bundle manager
	config := createUpdateConfig()
	logger := &CLILogger{verbose: true}
	bundleManager := update.NewBundleManager(config, logger)
	
	fmt.Printf("📦 Creating bundle: %s\n", bundleOutput)
	
	// Export bundle
	bundleInfo, err := bundleManager.ExportBundle(options)
	if err != nil {
		return fmt.Errorf("failed to export bundle: %w", err)
	}
	
	// Display results
	fmt.Printf("✅ Bundle created successfully!\n")
	fmt.Printf("   Path: %s\n", bundleInfo.Path)
	fmt.Printf("   Size: %s\n", update.FormatFileSize(bundleInfo.Size))
	fmt.Printf("   Components: %d\n", bundleInfo.ComponentCount)
	if bundleInfo.Signed {
		fmt.Printf("   Signed: Yes\n")
	}
	if expiresAt != nil {
		fmt.Printf("   Expires: %s\n", expiresAt.Format("2006-01-02"))
	}
	
	return nil
}

// runBundleImport executes the bundle import command
func runBundleImport(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	// Check if bundle exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle file not found: %s", bundlePath)
	}
	
	// Create import options
	options := &update.BundleImportOptions{
		VerifyIntegrity: bundleVerifyIntegrity,
		ForceImport:     bundleForceImport,
		CreateBackup:    bundleCreateBackup,
		DryRun:          bundleDryRun,
	}
	
	// Initialize bundle manager
	config := createUpdateConfig()
	logger := &CLILogger{verbose: true}
	bundleManager := update.NewBundleManager(config, logger)
	
	fmt.Printf("📦 Importing bundle: %s\n", bundlePath)
	
	// Import bundle
	result, err := bundleManager.ImportBundle(bundlePath, options)
	if err != nil {
		return fmt.Errorf("failed to import bundle: %w", err)
	}
	
	// Display results
	if result.Success {
		fmt.Printf("✅ Bundle imported successfully!\n")
		if len(result.ComponentsUpdated) > 0 {
			fmt.Printf("   Updated components: %s\n", strings.Join(result.ComponentsUpdated, ", "))
		}
		if result.RestartRequired {
			fmt.Printf("   ⚠️  Restart required\n")
		}
	} else {
		fmt.Printf("❌ Bundle import completed with errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("   - %s\n", err)
		}
	}
	
	return nil
}

// Helper functions

func createUpdateConfig() *update.Config {
	return update.DefaultConfig()
}

func confirmUpdate() bool {
	fmt.Print("Do you want to proceed with the update? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func confirmRestart() bool {
	fmt.Print("Restart now? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func getUpdateComponents() []string {
	if updateTemplatesOnly {
		return []string{"templates"}
	}
	if updateModulesOnly {
		return []string{"modules"}
	}
	if updateBinaryOnly {
		return []string{"binary"}
	}
	return []string{} // Empty means all components
}

func displayUpdateResults(summary *update.UpdateSummary) {
	fmt.Println("\n📊 Update Results:")
	fmt.Printf("   Duration: %s\n", summary.TotalDuration.Round(time.Second))
	
	if summary.Success {
		fmt.Printf("   Status: ✅ Success\n")
	} else {
		fmt.Printf("   Status: ❌ Failed\n")
	}
	
	for _, result := range summary.Results {
		icon := "✅"
		if !result.Success {
			icon = "❌"
		}
		
		fmt.Printf("   %s %s: ", icon, result.Component)
		if result.Success {
			if result.OldVersion != result.NewVersion {
				fmt.Printf("%s → %s", result.OldVersion, result.NewVersion)
			} else {
				fmt.Printf("up to date")
			}
		} else {
			fmt.Printf("failed (%v)", result.Error)
		}
		fmt.Println()
	}
}

func determineBundleType() string {
	if bundleIncludeBinary && bundleIncludeTemplates && bundleIncludeModules {
		return update.BundleTypeFull
	}
	if bundleIncludeTemplates && !bundleIncludeBinary && !bundleIncludeModules {
		return update.BundleTypeTemplates
	}
	if bundleIncludeModules && !bundleIncludeBinary && !bundleIncludeTemplates {
		return update.BundleTypeModules
	}
	return update.BundleTypeMixed
}

func parseExpiration(expires string) (time.Time, error) {
	// Simple parsing - could be more sophisticated
	if strings.HasSuffix(expires, "d") {
		days := strings.TrimSuffix(expires, "d")
		if d, err := time.ParseDuration(days + "d"); err == nil {
			return time.Now().Add(d), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported expiration format: %s", expires)
}

func getCurrentVersion() string {
	version, _ := update.GetCurrentVersion()
	return version
}

func getPlatformString() string {
	return update.GetPlatformString()
}

func getVersionInfo() *update.VersionInfo {
	return &update.VersionInfo{
		Version:   getCurrentVersion(),
		BuildDate: time.Now(), // Would be set at build time
		Platform:  getPlatformString(),
		Arch:      update.GetArchString(),
		Commit:    "unknown",
		Branch:    "main",
		GoVersion: "1.21",
		Components: map[string]string{
			"templates": "1.0.0",
			"modules":   "1.0.0",
		},
	}
}

// CLILogger implements the Logger interface for CLI output
type CLILogger struct {
	verbose bool
}

func (l *CLILogger) Info(msg string) {
	fmt.Printf("ℹ️  %s\n", msg)
}

func (l *CLILogger) Error(msg string, err error) {
	if err != nil {
		fmt.Printf("❌ %s: %v\n", msg, err)
	} else {
		fmt.Printf("❌ %s\n", msg)
	}
}

func (l *CLILogger) Debug(msg string) {
	if l.verbose {
		fmt.Printf("🔍 %s\n", msg)
	}
}

func (l *CLILogger) Warn(msg string) {
	fmt.Printf("⚠️  %s\n", msg)
}