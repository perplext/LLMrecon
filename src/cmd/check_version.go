package cmd

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/config"
	"github.com/perplext/LLMrecon/src/update"
	"github.com/perplext/LLMrecon/src/version"
	"github.com/spf13/cobra"
)

// Current version of the application
// This would typically be set during the build process
var (
	currentVersion = "1.0.0" // Default version for development
)

// checkVersionCmd represents the check-version command
var checkVersionCmd = &cobra.Command{
	Use:   "check-version",
	Short: "Check for available updates",
	Long: `Check for available updates for the core binary, templates, and modules.
This command connects to the configured update servers and compares the current
versions with the latest available versions. It displays information about any
available updates including version differences, change types, and release notes.`,
	Run: func(cmd *cobra.Command, args []string) {
		quiet, _ := cmd.Flags().GetBool("quiet")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		// Get update server URLs from config
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

		// Check for updates from primary source (GitHub)
		if !quiet {
			fmt.Println("Checking for updates from GitHub...")
		}

		githubChecker, err := update.NewVersionChecker(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating GitHub version checker: %v\n", err)
			os.Exit(1)
		}
		// Set the update server URL
		githubChecker.UpdateServerURL = cfg.UpdateSources.GitHub
		githubChecker.CurrentVersions = currentVersions
		githubUpdates, err := githubChecker.CheckForUpdates()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking GitHub updates: %v\n", err)
			// Continue to check GitLab if GitHub fails
		}

		// Check for updates from secondary source (GitLab) if configured
		var gitlabUpdates []update.UpdateInfo
		if cfg.UpdateSources.GitLab != "" {
			if !quiet {
				fmt.Println("Checking for updates from GitLab...")
			}

			gitlabChecker, err := update.NewVersionChecker(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating GitLab version checker: %v\n", err)
				// Continue with GitHub updates if GitLab fails
			} else {
				// Set the update server URL
				gitlabChecker.UpdateServerURL = cfg.UpdateSources.GitLab
				gitlabChecker.CurrentVersions = currentVersions
				gitlabUpdates, err = gitlabChecker.CheckForUpdates()
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking GitLab updates: %v\n", err)
				// Continue with GitHub updates if GitLab fails
			}
		}

		// Merge updates, preferring GitHub for core and GitLab for custom templates/modules
		updates := mergeUpdates(githubUpdates, gitlabUpdates)

		// Output results
		if jsonOutput {
			outputJSON(updates)
		} else {
			fmt.Println(update.FormatUpdateInfo(updates))

			// Show update command hint if updates are available
			if len(updates) > 0 {
				fmt.Println("\nTo apply updates, run: LLMrecon update")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(checkVersionCmd)

	// Add flags
	checkVersionCmd.Flags().BoolP("quiet", "q", false, "Suppress informational output")
	checkVersionCmd.Flags().BoolP("json", "j", false, "Output results in JSON format")
}

// getLocalVersions retrieves the current versions of templates and modules
func getLocalVersions() (version.Version, map[string]version.Version, error) {
	// This would typically read from a local state file or database
	// For now, we'll return placeholder versions
	templateVersion, _ := version.ParseVersion("1.0.0")

	moduleVersions := map[string]version.Version{}
	// Add some example module versions
	openaiVersion, _ := version.ParseVersion("1.0.0")
	anthropicVersion, _ := version.ParseVersion("0.9.0")

	moduleVersions["openai"] = openaiVersion
	moduleVersions["anthropic"] = anthropicVersion

	return templateVersion, moduleVersions, nil
}

// mergeUpdates combines updates from multiple sources, with priority rules
func mergeUpdates(githubUpdates, gitlabUpdates []update.UpdateInfo) []update.UpdateInfo {
	// Create a map to store the highest priority update for each component
	updateMap := make(map[string]update.UpdateInfo)

	// Process GitHub updates first (they take precedence for core components)
	for _, update := range githubUpdates {
		updateMap[update.Component] = update
	}

	// Process GitLab updates, which take precedence for custom templates and modules
	for _, update := range gitlabUpdates {
		// For core components, only use GitLab if GitHub doesn't have an update
		if update.Component == "core" {
			if _, exists := updateMap[update.Component]; !exists {
				updateMap[update.Component] = update
			}
		} else {
			// For templates and modules, prefer GitLab (internal/development versions)
			updateMap[update.Component] = update
		}
	}

	// Convert map back to slice
	result := make([]update.UpdateInfo, 0, len(updateMap))
	for _, update := range updateMap {
		result = append(result, update)
	}

	return result
}

// outputJSON outputs the update information in JSON format
func outputJSON(updates []update.UpdateInfo) {
	// This would typically use json.Marshal and print the result
	// For now, we'll just print a placeholder
	fmt.Println("{\"updates\": []}")
}
