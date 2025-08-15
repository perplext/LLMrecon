package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/perplext/LLMrecon/src/config"
	"github.com/perplext/LLMrecon/src/update"
	"github.com/perplext/LLMrecon/src/version"
	"github.com/spf13/cobra"
)

// Color functions for output
var (
	green   = color.New(color.FgGreen).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	bold    = color.New(color.Bold).SprintFunc()
	dim     = color.New(color.Faint).SprintFunc()
)

// updateCheckCmd represents the update check command
var updateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for available updates",
	Long: `Check for available updates for the LLMreconing Tool and its components.

This command queries configured update sources to find available updates for:
- Core binary
- Templates
- Modules

It displays version differences, change types, and release notes without applying any changes.`,
	Example: `  # Check for all updates
  LLMrecon update check

  # Check only for template updates
  LLMrecon update check --component=templates

  # Check for updates with detailed output
  LLMrecon update check --verbose

  # Check updates from specific source
  LLMrecon update check --source=github`,
	RunE: runUpdateCheck,

func init() {
	updateCmd.AddCommand(updateCheckCmd)

	// Add flags
	updateCheckCmd.Flags().StringP("component", "c", "all", "Component to check (all, binary, templates, modules)")
	updateCheckCmd.Flags().BoolP("verbose", "v", false, "Show detailed update information")
	updateCheckCmd.Flags().String("source", "all", "Update source to check (all, github, gitlab, s3)")
	updateCheckCmd.Flags().Bool("json", false, "Output results in JSON format")
	updateCheckCmd.Flags().Bool("no-color", false, "Disable colored output")
	updateCheckCmd.Flags().Duration("timeout", 30*time.Second, "Timeout for update checks")

func runUpdateCheck(cmd *cobra.Command, args []string) error {
	// Get flags
	componentFlag, _ := cmd.Flags().GetString("component")
	verboseFlag, _ := cmd.Flags().GetBool("verbose")
	sourceFlag, _ := cmd.Flags().GetString("source")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	noColorFlag, _ := cmd.Flags().GetBool("no-color")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	// Disable colors if requested
	if noColorFlag {
		color.NoColor = true
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Get current versions
	currentVersions, err := getUpdateCurrentVersions(cfg)
	if err != nil {
		return fmt.Errorf("getting current versions: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Check for updates from configured sources
	updates, err := checkUpdatesFromSources(ctx, cfg, currentVersions, sourceFlag)
	if err != nil {
		return fmt.Errorf("checking updates: %w", err)
	}

	// Filter updates by component
	filteredUpdates := filterUpdatesByComponent(updates, componentFlag)

	// Output results
	if jsonFlag {
		return outputUpdateJSON(filteredUpdates)
	}

	return outputTable(filteredUpdates, verboseFlag)

func getUpdateCurrentVersions(cfg *config.Config) (map[string]version.Version, error) {
	versions := make(map[string]version.Version)

	// Get core version
	coreVer, err := version.ParseVersion(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("parsing core version: %w", err)
	}
	versions["binary"] = coreVer

	// Get template and module versions
	templateVer, moduleVersions, err := getLocalVersions()
	if err != nil {
		return nil, fmt.Errorf("getting local versions: %w", err)
	}

	versions["templates"] = templateVer
	for id, ver := range moduleVersions {
		versions[fmt.Sprintf("module.%s", id)] = ver
	}

	return versions, nil

func checkUpdatesFromSources(ctx context.Context, cfg *config.Config, currentVersions map[string]version.Version, source string) ([]update.ExtendedUpdateInfo, error) {
	var allUpdates []update.ExtendedUpdateInfo

	// Check GitHub if configured
	if (source == "all" || source == "github") && cfg.UpdateSources.GitHub != "" {
		fmt.Print(dim("Checking GitHub for updates... "))
		checker, err := update.NewVersionChecker(ctx)
		if err != nil {
			fmt.Println(red("✗"))
			if source == "github" {
				return nil, fmt.Errorf("GitHub checker creation failed: %w", err)
			}
			// Continue with other sources if not exclusive
		} else {
			checker.UpdateServerURL = cfg.UpdateSources.GitHub
			checker.CurrentVersions = currentVersions
			updates, err := checker.CheckForUpdatesContext(ctx)
			if err != nil {
				fmt.Println(red("✗"))
				if source == "github" {
					return nil, fmt.Errorf("GitHub check failed: %w", err)
				}
				// Continue with other sources if not exclusive
			} else {
				fmt.Println(green("✓"))
				allUpdates = append(allUpdates, updates...)
			}
		}
	}

	// Check GitLab if configured
	if (source == "all" || source == "gitlab") && cfg.UpdateSources.GitLab != "" {
		fmt.Print(dim("Checking GitLab for updates... "))
		checker, err := update.NewVersionChecker(ctx)
		if err != nil {
			fmt.Println(red("✗"))
			if source == "gitlab" {
				return nil, fmt.Errorf("GitLab checker creation failed: %w", err)
			}
		} else {
			checker.UpdateServerURL = cfg.UpdateSources.GitLab
			checker.CurrentVersions = currentVersions
			updates, err := checker.CheckForUpdatesContext(ctx)
			if err != nil {
				fmt.Println(red("✗"))
				if source == "gitlab" {
					return nil, fmt.Errorf("GitLab check failed: %w", err)
				}
			} else {
				fmt.Println(green("✓"))
				allUpdates = append(allUpdates, updates...)
			}
		}
	}

	// Check S3 if configured
	// Note: S3 support is not currently available in the config
	/*
		if (source == "all" || source == "s3") && cfg.UpdateSources.S3 != "" {
			fmt.Print(dim("Checking S3 for updates... "))
			checker := update.NewVersionChecker(cfg.UpdateSources.S3, currentVersions)
			updates, err := checker.CheckForUpdatesContext(ctx)
			if err != nil {
				fmt.Println(red("✗"))
				if source == "s3" {
					return nil, fmt.Errorf("S3 check failed: %w", err)
				}
			} else {
				fmt.Println(green("✓"))
				allUpdates = append(allUpdates, updates...)
			}
		}
	*/

	// Return updates directly (merging is done elsewhere if needed)
	return allUpdates, nil

func filterUpdatesByComponent(updates []update.ExtendedUpdateInfo, component string) []update.ExtendedUpdateInfo {
	if component == "all" {
		return updates
	}

	var filtered []update.ExtendedUpdateInfo
	for _, u := range updates {
		switch component {
		case "binary":
			if u.Component == "binary" || u.Component == "core" {
				filtered = append(filtered, u)
			}
		case "templates":
			if u.Component == "templates" {
				filtered = append(filtered, u)
			}
		case "modules":
			if strings.HasPrefix(u.Component, "module.") {
				filtered = append(filtered, u)
			}
		default:
			if u.Component == component {
				filtered = append(filtered, u)
			}
		}
	}
	return filtered

func outputTable(updates []update.ExtendedUpdateInfo, verbose bool) error {
	if len(updates) == 0 {
		fmt.Println("\n" + green("✓") + " All components are up to date!")
		return nil
	}

	fmt.Println("\n" + bold("Available Updates:"))
	fmt.Println()

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		bold("Component"),
		bold("Current"),
		bold("Available"),
		bold("Type"),
		bold("Size"))
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		strings.Repeat("-", 20),
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
		strings.Repeat("-", 8),
		strings.Repeat("-", 10))

	// Updates
	for _, u := range updates {
		changeColor := getChangeTypeColor(u.ChangeType)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			cyan(u.Component),
			u.CurrentVersion.String(),
			green(u.LatestVersion.String()),
			changeColor(formatChangeType(u.ChangeType)),
			formatUpdateSize(u.Size))
	}

	w.Flush()

	// Verbose output
	if verbose {
		fmt.Println("\n" + bold("Details:"))
		for _, u := range updates {
			fmt.Printf("\n%s %s:\n", bold("→"), cyan(u.Component))
			fmt.Printf("  Release Date: %s\n", u.ReleaseDate.Format("2006-01-02"))
			if u.ReleaseNotes != "" {
				fmt.Printf("  Release Notes:\n")
				for _, line := range strings.Split(u.ReleaseNotes, "\n") {
					if line != "" {
						fmt.Printf("    %s\n", dim(line))
					}
				}
			}
			if len(u.Dependencies) > 0 {
				fmt.Printf("  Dependencies:\n")
				for _, dep := range u.Dependencies {
					fmt.Printf("    - %s\n", dep)
				}
			}
			fmt.Printf("  Download URL: %s\n", dim(u.DownloadURL))
			if u.ChecksumSHA256 != "" {
				fmt.Printf("  SHA256: %s\n", dim(u.ChecksumSHA256[:16]+"..."))
			}
		}
	}

	// Summary
	fmt.Printf("\n%s %d update(s) available. Run '%s' to apply.\n",
		yellow("→"),
		len(updates),
		bold("LLMrecon update apply"))

	return nil

func outputUpdateJSON(updates []update.ExtendedUpdateInfo) error {
	// Convert to JSON-friendly format
	type jsonUpdate struct {
		Component       string    `json:"component"`
		CurrentVersion  string    `json:"current_version"`
		LatestVersion   string    `json:"latest_version"`
		ChangeType      string    `json:"change_type"`
		Size            int64     `json:"size"`
		ReleaseDate     time.Time `json:"release_date"`
		ReleaseNotes    string    `json:"release_notes,omitempty"`
		DownloadURL     string    `json:"download_url"`
		ChecksumSHA256  string    `json:"checksum_sha256,omitempty"`
		Dependencies    []string  `json:"dependencies,omitempty"`
		BreakingChanges bool      `json:"breaking_changes"`
	}

	jsonUpdates := make([]jsonUpdate, len(updates))
	for i, u := range updates {
		jsonUpdates[i] = jsonUpdate{
			Component:       u.Component,
			CurrentVersion:  u.CurrentVersion.String(),
			LatestVersion:   u.LatestVersion.String(),
			ChangeType:      string(u.ChangeType),
			Size:            u.Size,
			ReleaseDate:     u.ReleaseDate,
			ReleaseNotes:    u.ReleaseNotes,
			DownloadURL:     u.DownloadURL,
			ChecksumSHA256:  u.ChecksumSHA256,
			Dependencies:    u.Dependencies,
			BreakingChanges: u.BreakingChanges,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonUpdates)

func getChangeTypeColor(changeType version.VersionChangeType) func(...interface{}) string {
	switch changeType {
	case version.MajorChange:
		return red
	case version.MinorChange:
		return yellow
	case version.PatchChange:
		return green
	default:
		return blue
	}

func formatChangeType(changeType version.VersionChangeType) string {
	switch changeType {
	case version.MajorChange:
		return "Major"
	case version.MinorChange:
		return "Minor"
	case version.PatchChange:
		return "Patch"
	case version.PreReleaseChange:
		return "Pre-release"
	case version.BuildChange:
		return "Build"
	default:
		return "Unknown"
	}

func formatUpdateSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
