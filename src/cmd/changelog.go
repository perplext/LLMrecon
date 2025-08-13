package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/perplext/LLMrecon/src/config"
	"github.com/perplext/LLMrecon/src/version"
	"github.com/spf13/cobra"
)

// changelogCmd represents the changelog command
var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Display version history and changes",
	Long: `Display the changelog for the LLMreconing Tool and its components.

This command fetches and displays version history, including:
- Release notes for each version
- Breaking changes
- New features
- Bug fixes
- Security updates

You can filter the changelog by component and version range.`,
	Example: `  # Show full changelog
  LLMrecon changelog

  # Show changelog for specific component
  LLMrecon changelog --component=templates

  # Show changes since a specific version
  LLMrecon changelog --from-version=1.2.0

  # Show only the latest 5 releases
  LLMrecon changelog --limit=5

  # Output in JSON format
  LLMrecon changelog --json`,
	RunE: runChangelog,
}

func init() {
	rootCmd.AddCommand(changelogCmd)

	// Add flags
	changelogCmd.Flags().StringP("component", "c", "all", "Component to show (all, binary, templates, modules)")
	changelogCmd.Flags().String("from-version", "", "Show changes from this version onwards")
	changelogCmd.Flags().String("to-version", "", "Show changes up to this version")
	changelogCmd.Flags().IntP("limit", "l", 10, "Limit number of releases to show")
	changelogCmd.Flags().Bool("json", false, "Output in JSON format")
	changelogCmd.Flags().Bool("only-breaking", false, "Show only breaking changes")
	changelogCmd.Flags().Bool("only-security", false, "Show only security updates")
	changelogCmd.Flags().Duration("timeout", 30*time.Second, "Timeout for fetching changelog")
}

func runChangelog(cmd *cobra.Command, args []string) error {
	// Get flags
	componentFlag, _ := cmd.Flags().GetString("component")
	fromVersion, _ := cmd.Flags().GetString("from-version")
	toVersion, _ := cmd.Flags().GetString("to-version")
	limit, _ := cmd.Flags().GetInt("limit")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	onlyBreaking, _ := cmd.Flags().GetBool("only-breaking")
	onlySecurity, _ := cmd.Flags().GetBool("only-security")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil && !jsonFlag {
		fmt.Fprintf(os.Stderr, yellow("Warning: Could not load configuration: %v\n"), err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Fetch changelog
	changelog, err := fetchChangelog(ctx, cfg, componentFlag)
	if err != nil {
		// Try to load local changelog as fallback
		changelog, err = loadLocalChangelog(componentFlag)
		if err != nil {
			return fmt.Errorf("fetching changelog: %w", err)
		}
		if !jsonFlag {
			fmt.Println(yellow("Note: Showing cached changelog (offline mode)"))
		}
	}

	// Filter changelog
	filtered := filterChangelog(changelog, fromVersion, toVersion, limit, onlyBreaking, onlySecurity)

	// Output results
	if jsonFlag {
		return outputChangelogJSON(filtered)
	}

	return outputChangelogText(filtered)
}

// ChangelogEntry represents a single changelog entry
type ChangelogEntry struct {
	Version         string    `json:"version"`
	Date            time.Time `json:"date"`
	Component       string    `json:"component"`
	Summary         string    `json:"summary"`
	BreakingChanges []string  `json:"breaking_changes,omitempty"`
	Features        []string  `json:"features,omitempty"`
	Improvements    []string  `json:"improvements,omitempty"`
	BugFixes        []string  `json:"bug_fixes,omitempty"`
	SecurityFixes   []string  `json:"security_fixes,omitempty"`
	Contributors    []string  `json:"contributors,omitempty"`
	DownloadURL     string    `json:"download_url,omitempty"`
}

// Changelog represents the full changelog
type Changelog struct {
	Component string           `json:"component"`
	Entries   []ChangelogEntry `json:"entries"`
	Generated time.Time        `json:"generated"`
}

func fetchChangelog(ctx context.Context, cfg *config.Config, component string) (*Changelog, error) {
	// Determine changelog URL based on component
	var changelogURL string
	if cfg != nil && cfg.UpdateSources.GitHub != "" {
		changelogURL = fmt.Sprintf("%s/changelog-%s.json", cfg.UpdateSources.GitHub, component)
	} else {
		// Default URL
		changelogURL = fmt.Sprintf("https://api.github.com/repos/LLMrecon/LLMrecon/releases")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", changelogURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add headers for GitHub API
	if strings.Contains(changelogURL, "github.com") {
		req.Header.Set("Accept", "application/vnd.github.v3+json")
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching changelog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response based on format
	if strings.Contains(changelogURL, "github.com") && strings.Contains(changelogURL, "/releases") {
		return parseGitHubReleases(resp.Body, component)
	}

	// Parse as standard changelog format
	var changelog Changelog
	if err := json.NewDecoder(resp.Body).Decode(&changelog); err != nil {
		return nil, fmt.Errorf("parsing changelog: %w", err)
	}

	return &changelog, nil
}

func parseGitHubReleases(r io.Reader, component string) (*Changelog, error) {
	var releases []struct {
		TagName     string    `json:"tag_name"`
		Name        string    `json:"name"`
		Body        string    `json:"body"`
		PublishedAt time.Time `json:"published_at"`
		HTMLURL     string    `json:"html_url"`
	}

	if err := json.NewDecoder(r).Decode(&releases); err != nil {
		return nil, fmt.Errorf("parsing GitHub releases: %w", err)
	}

	changelog := &Changelog{
		Component: component,
		Generated: time.Now(),
		Entries:   make([]ChangelogEntry, 0, len(releases)),
	}

	for _, release := range releases {
		entry := ChangelogEntry{
			Version:     strings.TrimPrefix(release.TagName, "v"),
			Date:        release.PublishedAt,
			Component:   component,
			Summary:     release.Name,
			DownloadURL: release.HTMLURL,
		}

		// Parse release body for different sections
		parseReleaseBody(release.Body, &entry)

		changelog.Entries = append(changelog.Entries, entry)
	}

	return changelog, nil
}

func parseReleaseBody(body string, entry *ChangelogEntry) {
	lines := strings.Split(body, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for section headers
		switch {
		case strings.HasPrefix(strings.ToLower(line), "## breaking"):
			currentSection = "breaking"
		case strings.HasPrefix(strings.ToLower(line), "## feature"):
			currentSection = "features"
		case strings.HasPrefix(strings.ToLower(line), "## improvement"):
			currentSection = "improvements"
		case strings.HasPrefix(strings.ToLower(line), "## bug fix"):
			currentSection = "bugfixes"
		case strings.HasPrefix(strings.ToLower(line), "## security"):
			currentSection = "security"
		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* "):
			// Parse list items
			item := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			switch currentSection {
			case "breaking":
				entry.BreakingChanges = append(entry.BreakingChanges, item)
			case "features":
				entry.Features = append(entry.Features, item)
			case "improvements":
				entry.Improvements = append(entry.Improvements, item)
			case "bugfixes":
				entry.BugFixes = append(entry.BugFixes, item)
			case "security":
				entry.SecurityFixes = append(entry.SecurityFixes, item)
			}
		}
	}
}

func loadLocalChangelog(component string) (*Changelog, error) {
	// Try to load from local cache
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cacheFile := fmt.Sprintf("%s/.LLMrecon/cache/changelog-%s.json", homeDir, component)
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("no cached changelog available")
	}

	var changelog Changelog
	if err := json.Unmarshal(data, &changelog); err != nil {
		return nil, fmt.Errorf("parsing cached changelog: %w", err)
	}

	return &changelog, nil
}

func filterChangelog(changelog *Changelog, fromVersion, toVersion string, limit int, onlyBreaking, onlySecurity bool) *Changelog {
	filtered := &Changelog{
		Component: changelog.Component,
		Generated: changelog.Generated,
		Entries:   []ChangelogEntry{},
	}

	var fromVer, toVer *version.Version
	if fromVersion != "" {
		v, _ := version.ParseVersion(fromVersion)
		fromVer = &v
	}
	if toVersion != "" {
		v, _ := version.ParseVersion(toVersion)
		toVer = &v
	}

	for _, entry := range changelog.Entries {
		// Filter by version range
		if fromVer != nil || toVer != nil {
			entryVer, err := version.ParseVersion(entry.Version)
			if err != nil {
				continue
			}

			if fromVer != nil && entryVer.Compare(fromVer) < 0 {
				continue
			}
			if toVer != nil && entryVer.Compare(toVer) > 0 {
				continue
			}
		}

		// Filter by type
		if onlyBreaking && len(entry.BreakingChanges) == 0 {
			continue
		}
		if onlySecurity && len(entry.SecurityFixes) == 0 {
			continue
		}

		// Apply filters to entry
		filteredEntry := entry
		if onlyBreaking {
			filteredEntry = ChangelogEntry{
				Version:         entry.Version,
				Date:            entry.Date,
				Component:       entry.Component,
				Summary:         entry.Summary,
				BreakingChanges: entry.BreakingChanges,
			}
		} else if onlySecurity {
			filteredEntry = ChangelogEntry{
				Version:       entry.Version,
				Date:          entry.Date,
				Component:     entry.Component,
				Summary:       entry.Summary,
				SecurityFixes: entry.SecurityFixes,
			}
		}

		filtered.Entries = append(filtered.Entries, filteredEntry)

		// Apply limit
		if limit > 0 && len(filtered.Entries) >= limit {
			break
		}
	}

	return filtered
}

func outputChangelogText(changelog *Changelog) error {
	if len(changelog.Entries) == 0 {
		fmt.Println("No changelog entries found matching the criteria.")
		return nil
	}

	fmt.Printf("%s %s\n\n", bold("Changelog for"), cyan(changelog.Component))

	for i, entry := range changelog.Entries {
		// Version header
		fmt.Printf("%s %s %s\n",
			bold(fmt.Sprintf("v%s", entry.Version)),
			dim("-"),
			dim(entry.Date.Format("2006-01-02")))

		if entry.Summary != "" {
			fmt.Printf("%s\n", entry.Summary)
		}
		fmt.Println()

		// Breaking changes
		if len(entry.BreakingChanges) > 0 {
			fmt.Println(red("Breaking Changes:"))
			for _, change := range entry.BreakingChanges {
				fmt.Printf("  • %s\n", change)
			}
			fmt.Println()
		}

		// Security fixes
		if len(entry.SecurityFixes) > 0 {
			fmt.Println(yellow("Security Fixes:"))
			for _, fix := range entry.SecurityFixes {
				fmt.Printf("  • %s\n", fix)
			}
			fmt.Println()
		}

		// Features
		if len(entry.Features) > 0 {
			fmt.Println(green("New Features:"))
			for _, feature := range entry.Features {
				fmt.Printf("  • %s\n", feature)
			}
			fmt.Println()
		}

		// Improvements
		if len(entry.Improvements) > 0 {
			fmt.Println(blue("Improvements:"))
			for _, improvement := range entry.Improvements {
				fmt.Printf("  • %s\n", improvement)
			}
			fmt.Println()
		}

		// Bug fixes
		if len(entry.BugFixes) > 0 {
			fmt.Println(magenta("Bug Fixes:"))
			for _, fix := range entry.BugFixes {
				fmt.Printf("  • %s\n", fix)
			}
			fmt.Println()
		}

		// Add separator between entries (except last)
		if i < len(changelog.Entries)-1 {
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println()
		}
	}

	return nil
}

func outputChangelogJSON(changelog *Changelog) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(changelog)
}

// CacheChangelog saves changelog to local cache
func CacheChangelog(changelog *Changelog) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cacheDir := fmt.Sprintf("%s/.LLMrecon/cache", homeDir)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	cacheFile := fmt.Sprintf("%s/changelog-%s.json", cacheDir, changelog.Component)
	data, err := json.MarshalIndent(changelog, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}
