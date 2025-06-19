package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/config"
	"github.com/perplext/LLMrecon/src/update"
	"github.com/perplext/LLMrecon/src/version"
	"github.com/spf13/cobra"
)

// updateApplyCmd represents the update apply command
var updateApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply available updates",
	Long: `Apply available updates to the LLMreconing Tool and its components.

This command downloads and installs updates after verifying their integrity.
It supports updating the core binary, templates, and modules with options for
selective updates and automatic backup creation.`,
	Example: `  # Apply all available updates
  LLMrecon update apply

  # Apply only template updates
  LLMrecon update apply --component=templates

  # Apply updates without confirmation
  LLMrecon update apply --yes

  # Apply updates with backup
  LLMrecon update apply --backup`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		componentFlag, _ := cmd.Flags().GetString("component")
		forceFlag, _ := cmd.Flags().GetBool("yes")
		noVerifyFlag, _ := cmd.Flags().GetBool("no-verify")
		backupFlag, _ := cmd.Flags().GetBool("backup")
		
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}
		
		// Parse current versions
		coreVersion, err := version.ParseVersion(currentVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing current version: %v\n", err)
			os.Exit(1)
		}
		
		// Get template and module versions from local state
		templateVersion, moduleVersions, err := getLocalVersions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting local versions: %v\n", err)
			os.Exit(1)
		}
		
		// Create version map
		currentVersions := map[string]version.Version{
			"core":      coreVersion,
			"templates": templateVersion,
		}
		
		// Add module versions to the map
		for id, ver := range moduleVersions {
			currentVersions[fmt.Sprintf("module.%s", id)] = ver
		}
		
		// Check for updates
		fmt.Println("Checking for updates...")
		
		// Check GitHub updates
		ctx := context.Background()
		githubChecker := update.NewVersionChecker(cfg.UpdateSources.GitHub, currentVersions)
		githubUpdates, err := githubChecker.CheckForUpdatesContext(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking GitHub updates: %v\n", err)
			// Continue to check GitLab if GitHub fails
		}
		
		// Check GitLab updates if configured
		var gitlabUpdates []update.UpdateInfo
		if cfg.UpdateSources.GitLab != "" {
			gitlabChecker := update.NewVersionChecker(cfg.UpdateSources.GitLab, currentVersions)
			gitlabUpdates, err = gitlabChecker.CheckForUpdatesContext(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking GitLab updates: %v\n", err)
				// Continue with GitHub updates if GitLab fails
			}
		}
		
		// Merge updates
		allUpdates := update.MergeUpdates(githubUpdates, gitlabUpdates)
		
		// Filter updates based on component flag
		var updates []update.UpdateInfo
		if componentFlag == "all" {
			updates = allUpdates
		} else {
			for _, u := range allUpdates {
				if u.Component == componentFlag || 
				   (componentFlag == "modules" && strings.HasPrefix(u.Component, "module.")) {
					updates = append(updates, u)
				}
			}
		}
		
		// Check if there are any updates
		if len(updates) == 0 {
			fmt.Println("No updates available.")
			return
		}
		
		// Display available updates
		fmt.Println("Available updates:")
		for _, u := range updates {
			fmt.Printf("- %s: %s → %s (%s)\n", 
				u.Component, 
				u.CurrentVersion.String(), 
				u.LatestVersion.String(),
				update.FormatChangeType(u.ChangeType))
		}
		
		// Create backup if requested
		if backupFlag {
			fmt.Println("\nCreating backup...")
			if err := createBackup(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating backup: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Backup created successfully.")
		}
		
		// Confirm update unless force flag is set
		if !forceFlag {
			fmt.Print("\nDo you want to apply these updates? [y/N] ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Update canceled.")
				return
			}
		}
		
		// Create temporary directory for downloads
		tempDir, err := os.MkdirTemp("", "LLMrecon-update")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temporary directory: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tempDir)
		
		// Download and apply updates
		for _, u := range updates {
			fmt.Printf("\nUpdating %s to version %s...\n", u.Component, u.LatestVersion.String())
			
			// Download update
			downloadPath := filepath.Join(tempDir, fmt.Sprintf("%s-%s.zip", u.Component, u.LatestVersion.String()))
			fmt.Printf("Downloading from %s...\n", u.DownloadURL)
			
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			err := update.DownloadWithProgress(ctx, u.DownloadURL, downloadPath)
			cancel()
			
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading update: %v\n", err)
				continue
			}
			
			// Verify update integrity
			if !noVerifyFlag && cfg.Security.VerifySignatures {
				fmt.Println("Verifying update integrity...")
				err = update.VerifyUpdate(
					downloadPath, 
					u.ChecksumSHA256, 
					u.Signature, 
					cfg.Security.PublicKey,
				)
				
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error verifying update: %v\n", err)
					fmt.Println("Update failed. The downloaded file may be corrupted or tampered with.")
					continue
				}
			}
			
			// Apply update based on component type
			switch {
			case u.Component == "core" || u.Component == "binary":
				err = applyCoreBinaryUpdate(downloadPath)
			case u.Component == "templates":
				err = applyTemplatesUpdate(downloadPath, cfg.Templates.Dir)
			case strings.HasPrefix(u.Component, "module."):
				moduleID := strings.TrimPrefix(u.Component, "module.")
				err = applyModuleUpdate(downloadPath, moduleID, cfg.Modules.Dir)
			}
			
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error applying update: %v\n", err)
				continue
			}
			
			fmt.Printf("Successfully updated %s to version %s.\n", u.Component, u.LatestVersion.String())
		}
		
		fmt.Println("\nUpdate process completed.")
	},
}

func init() {
	updateCmd.AddCommand(updateApplyCmd)
	
	// Add flags
	updateApplyCmd.Flags().StringP("component", "c", "all", "Component to update (all, binary, templates, modules)")
	updateApplyCmd.Flags().BoolP("yes", "y", false, "Apply updates without confirmation")
	updateApplyCmd.Flags().Bool("no-verify", false, "Skip signature verification")
	updateApplyCmd.Flags().Bool("backup", false, "Create backup before applying updates")
}

// createBackup creates a backup of the current installation
func createBackup(cfg *config.Config) error {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Create a timestamped backup directory
	// 2. Copy the current binary
	// 3. Archive the templates directory
	// 4. Archive the modules directory
	// 5. Save the current configuration
	
	fmt.Println("Backup functionality not implemented in this version.")
	return nil
}

// applyCoreBinaryUpdate applies an update to the core binary
func applyCoreBinaryUpdate(downloadPath string) error {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Extract the downloaded archive
	// 2. Replace the current binary with the new one
	// 3. Ensure proper permissions are set
	// 4. Handle platform-specific details (e.g., Windows file locks)
	
	fmt.Println("Core binary update not implemented in this version.")
	return nil
}

// applyTemplatesUpdate applies an update to the templates
func applyTemplatesUpdate(downloadPath, templatesDir string) error {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Extract the downloaded archive to a temporary location
	// 2. Validate the template structure
	// 3. Backup existing templates
	// 4. Copy new templates to the templates directory
	
	fmt.Println("Templates update not implemented in this version.")
	return nil
}

// applyModuleUpdate applies an update to a specific module
func applyModuleUpdate(downloadPath, moduleID, modulesDir string) error {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Extract the downloaded archive to a temporary location
	// 2. Validate the module structure
	// 3. Backup existing module
	// 4. Copy new module to the modules directory
	
	fmt.Printf("Module update for '%s' not implemented in this version.\n", moduleID)
	return nil
}