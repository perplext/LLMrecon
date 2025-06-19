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

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update management commands",
	Long: `Manage updates for the LLMreconing Tool and its components.

This command provides subcommands for:
- Checking for available updates
- Applying updates
- Viewing version information
- Displaying changelogs

Use 'LLMrecon update <subcommand> --help' for more information about each subcommand.`,
	Example: `  # Check for updates
  LLMrecon update check
  
  # Apply all available updates
  LLMrecon update apply
  
  # Check version information
  LLMrecon version --verbose`,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
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
		githubChecker := update.NewVersionChecker(cfg.UpdateSources.GitHub, currentVersions)
		githubUpdates, err := githubChecker.CheckForUpdates()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking GitHub updates: %v\n", err)
			// Continue to check GitLab if GitHub fails
		}
		
		// Check GitLab updates if configured
		var gitlabUpdates []update.UpdateInfo
		if cfg.UpdateSources.GitLab != "" {
			gitlabChecker := update.NewVersionChecker(cfg.UpdateSources.GitLab, currentVersions)
			gitlabUpdates, err = gitlabChecker.CheckForUpdates()
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
				formatChangeType(u.ChangeType))
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
			case u.Component == "core":
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
	rootCmd.AddCommand(updateCmd)
	
	// Add flags
	updateCmd.Flags().StringP("component", "c", "all", "Component to update (core, templates, modules, or all)")
	updateCmd.Flags().BoolP("force", "f", false, "Apply updates without confirmation")
	updateCmd.Flags().Bool("no-verify", false, "Skip signature verification")
}

// formatChangeType is a wrapper around update.formatChangeType
func formatChangeType(changeType version.VersionChangeType) string {
	return update.FormatChangeType(changeType)
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
