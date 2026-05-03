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
		githubChecker, err := update.NewVersionChecker(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating GitHub version checker: %v\n", err)
			os.Exit(1)
		}
		githubChecker.UpdateServerURL = cfg.UpdateSources.GitHub
		githubChecker.CurrentVersions = currentVersions
		githubUpdates, err := githubChecker.CheckForUpdatesContext(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking GitHub updates: %v\n", err)
			// Continue to check GitLab if GitHub fails
		}

		// Check GitLab updates if configured
		var gitlabUpdates []update.ExtendedUpdateInfo
		if cfg.UpdateSources.GitLab != "" {
			gitlabChecker, err := update.NewVersionChecker(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating GitLab version checker: %v\n", err)
				// Continue with GitHub updates if GitLab fails
			} else {
				gitlabChecker.UpdateServerURL = cfg.UpdateSources.GitLab
				gitlabChecker.CurrentVersions = currentVersions
				gitlabUpdates, err = gitlabChecker.CheckForUpdatesContext(ctx)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking GitLab updates: %v\n", err)
				// Continue with GitHub updates if GitLab fails
			}
		}

		// Merge updates - for now just append them
		var allUpdates []update.ExtendedUpdateInfo
		allUpdates = append(allUpdates, githubUpdates...)
		if gitlabUpdates != nil {
			allUpdates = append(allUpdates, gitlabUpdates...)
		}

		// Filter updates based on component flag
		var updates []update.ExtendedUpdateInfo
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
			fmt.Scanln(&response) // #nosec G104 -- error reading stdin is handled by checking response value
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

		// Track per-update apply failures so we can exit non-zero at the
		// end if any update failed to land. Mirrors how the Tier 1
		// "not implemented" stubs surface: each is an error that should
		// signal failure, not be papered over with a generic completion message.
		applyErrorCount := 0

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
				applyErrorCount++
				continue
			}

			fmt.Printf("Successfully updated %s to version %s.\n", u.Component, u.LatestVersion.String())
		}

		// v0.10.0 #174 Tier 1: any apply path that errored (including the
		// "not implemented" stubs) means the on-disk state did NOT change
		// to the version the operator asked for. Exit non-zero so any
		// surrounding shell/CI workflow notices, instead of trusting a
		// 0-exit-with-error-printed signal.
		if applyErrorCount > 0 {
			fmt.Fprintf(os.Stderr, "\nUpdate process completed with %d error(s); see above. No on-disk changes were made for the failing components.\n", applyErrorCount)
			os.Exit(1)
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

// applyNotImplementedHint is the standard guidance string operators see
// when an update-apply path is unimplemented in this release. Centralized
// so the message is consistent across the four stubs and easy to update
// once Tier 2 (#174) lands real implementations.
// applyNotImplementedHint is what operators see when an update-apply
// path is unimplemented in this release. The path interpolated is the
// already-downloaded bundle: applyXxxUpdate is only reached AFTER
// DownloadWithProgress has succeeded, so the bundle is sitting on disk
// at downloadPath. Operators can extract it manually until Tier 2 ships.
//
// Side effect (relied upon): os.Exit(1) at the loop tail skips the
// defer os.RemoveAll(tempDir), which is desirable on this error path —
// preserving the bundle is exactly what the message directs operators
// toward. (Tier 2 will add explicit cleanup once apply is real.)
const applyNotImplementedHint = "on-disk update apply not implemented in this version; bundle downloaded to %q — extract manually, or wait for v0.10.0 #174 Tier 2"

// createBackup creates a backup of the current installation.
//
// v0.10.0 #174 Tier 1: returns a non-nil error rather than silently no-op'ing.
// Tier 2 will implement (cp -a of install dir with timestamp suffix). Operator
// who explicitly passed --backup gets a hard failure rather than a fake
// "Backup created successfully" message they shouldn't trust.
func createBackup(cfg *config.Config) error {
	return fmt.Errorf("createBackup: not implemented in this version (deferred to v0.10.0 #174 Tier 2)")
}

// applyCoreBinaryUpdate applies an update to the core binary.
//
// v0.10.0 #174 Tier 1: stops printing fake success. Binary self-replace
// is high-risk (atomic swap of running process's binary, Windows file
// locks, permission preservation) and is deferred to v0.11.0 — the
// templates/modules paths in Tier 2 are lower-risk and ship first.
func applyCoreBinaryUpdate(downloadPath string) error {
	return fmt.Errorf("applyCoreBinaryUpdate: not implemented in this version (binary self-replace deferred; see v0.10.0 #174 Tier 1; download is at %q)", downloadPath)
}

// applyTemplatesUpdate applies an update to the templates directory.
//
// v0.10.0 #174 Tier 1: returns error instead of nil so the caller's
// "Successfully updated" message is suppressed. Tier 2 will implement
// via ZIP extract + atomic os.Rename over the templates dir.
func applyTemplatesUpdate(downloadPath, templatesDir string) error {
	return fmt.Errorf(applyNotImplementedHint+" (component=templates, target=%q)", downloadPath, templatesDir)
}

// applyModuleUpdate applies an update to a specific module.
//
// v0.10.0 #174 Tier 1: returns error instead of nil. Tier 2 will
// implement via the same atomic-replace pattern as templates, scoped to
// modules/<id>/.
func applyModuleUpdate(downloadPath, moduleID, modulesDir string) error {
	return fmt.Errorf(applyNotImplementedHint+" (component=module.%s, target=%q)", downloadPath, moduleID, modulesDir)
}
