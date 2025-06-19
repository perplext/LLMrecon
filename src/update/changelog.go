package update

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ChangelogManager handles changelog operations
type ChangelogManager struct {
	config *Config
	logger Logger
}

// NewChangelogManager creates a new changelog manager
func NewChangelogManager(config *Config, logger Logger) *ChangelogManager {
	return &ChangelogManager{
		config: config,
		logger: logger,
	}
}

// GetChangelog retrieves and parses the changelog
func (cm *ChangelogManager) GetChangelog(version string) (*Changelog, error) {
	// Try to get changelog from different sources
	changelog, err := cm.getRemoteChangelog()
	if err != nil {
		cm.logger.Debug("Failed to fetch remote changelog, trying local")
		changelog, err = cm.getLocalChangelog()
		if err != nil {
			return nil, fmt.Errorf("failed to get changelog: %w", err)
		}
	}
	
	// Filter entries if version specified
	if version != "" {
		changelog = cm.filterChangelogByVersion(changelog, version)
	}
	
	return changelog, nil
}

// GetChangesSince gets changelog entries since a specific version
func (cm *ChangelogManager) GetChangesSince(sinceVersion string) ([]ChangelogEntry, error) {
	changelog, err := cm.GetChangelog("")
	if err != nil {
		return nil, err
	}
	
	var entries []ChangelogEntry
	foundSince := false
	
	// Sort entries by version (newest first)
	sort.Slice(changelog.Entries, func(i, j int) bool {
		return cm.compareVersions(changelog.Entries[i].Version, changelog.Entries[j].Version) > 0
	})
	
	for _, entry := range changelog.Entries {
		if entry.Version == sinceVersion {
			foundSince = true
			break
		}
		entries = append(entries, entry)
	}
	
	if !foundSince && sinceVersion != "" {
		cm.logger.Warn(fmt.Sprintf("Version %s not found in changelog", sinceVersion))
	}
	
	return entries, nil
}

// DisplayChangelog displays changelog in a formatted way
func (cm *ChangelogManager) DisplayChangelog(changelog *Changelog) {
	fmt.Printf("# %s Changelog\n\n", changelog.Project)
	
	if changelog.Format != "" {
		fmt.Printf("Format: %s\n", changelog.Format)
	}
	
	fmt.Printf("Last Updated: %s\n\n", changelog.LastUpdated.Format("2006-01-02"))
	
	currentVersion := ""
	for _, entry := range changelog.Entries {
		if entry.Version != currentVersion {
			currentVersion = entry.Version
			fmt.Printf("## Version %s\n", entry.Version)
			if !entry.Date.IsZero() {
				fmt.Printf("*Released: %s*\n\n", entry.Date.Format("2006-01-02"))
			}
		}
		
		cm.displayChangelogEntry(entry)
	}
}

// DisplayChangesSince displays changes since a version
func (cm *ChangelogManager) DisplayChangesSince(sinceVersion string) error {
	changes, err := cm.GetChangesSince(sinceVersion)
	if err != nil {
		return err
	}
	
	if len(changes) == 0 {
		fmt.Printf("No changes since version %s\n", sinceVersion)
		return nil
	}
	
	fmt.Printf("# Changes since %s\n\n", sinceVersion)
	
	currentVersion := ""
	for _, entry := range changes {
		if entry.Version != currentVersion {
			currentVersion = entry.Version
			fmt.Printf("## Version %s\n", entry.Version)
			if !entry.Date.IsZero() {
				fmt.Printf("*Released: %s*\n\n", entry.Date.Format("2006-01-02"))
			}
		}
		
		cm.displayChangelogEntry(entry)
	}
	
	return nil
}

// GetLatestChanges gets the most recent changelog entries
func (cm *ChangelogManager) GetLatestChanges(count int) ([]ChangelogEntry, error) {
	changelog, err := cm.GetChangelog("")
	if err != nil {
		return nil, err
	}
	
	// Sort entries by date (newest first)
	sort.Slice(changelog.Entries, func(i, j int) bool {
		return changelog.Entries[i].Date.After(changelog.Entries[j].Date)
	})
	
	if count > len(changelog.Entries) {
		count = len(changelog.Entries)
	}
	
	return changelog.Entries[:count], nil
}

// getRemoteChangelog fetches changelog from remote source
func (cm *ChangelogManager) getRemoteChangelog() (*Changelog, error) {
	// Try different common changelog URLs
	urls := []string{
		"https://raw.githubusercontent.com/LLMrecon/LLMrecon/main/CHANGELOG.md",
		"https://raw.githubusercontent.com/LLMrecon/LLMrecon/main/CHANGELOG.json",
		"https://api.github.com/repos/LLMrecon/LLMrecon/releases",
	}
	
	for _, url := range urls {
		changelog, err := cm.fetchChangelogFromURL(url)
		if err == nil {
			return changelog, nil
		}
		cm.logger.Debug(fmt.Sprintf("Failed to fetch from %s: %v", url, err))
	}
	
	return nil, fmt.Errorf("no remote changelog found")
}

// getLocalChangelog gets changelog from local files
func (cm *ChangelogManager) getLocalChangelog() (*Changelog, error) {
	// Try different local paths
	paths := []string{
		"CHANGELOG.md",
		"CHANGELOG.json",
		"docs/CHANGELOG.md",
		filepath.Join(cm.config.TemplateDirectory, "CHANGELOG.md"),
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return cm.parseLocalChangelog(path)
		}
	}
	
	return nil, fmt.Errorf("no local changelog found")
}

// fetchChangelogFromURL fetches changelog from a URL
func (cm *ChangelogManager) fetchChangelogFromURL(url string) (*Changelog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", cm.config.UserAgent)
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}
	
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Determine format and parse
	if strings.Contains(url, ".json") || strings.Contains(url, "/releases") {
		return cm.parseJSONChangelog(content, url)
	} else {
		return cm.parseMarkdownChangelog(string(content))
	}
}

// parseLocalChangelog parses a local changelog file
func (cm *ChangelogManager) parseLocalChangelog(path string) (*Changelog, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	if strings.HasSuffix(path, ".json") {
		return cm.parseJSONChangelog(content, path)
	} else {
		return cm.parseMarkdownChangelog(string(content))
	}
}

// parseMarkdownChangelog parses a Markdown changelog
func (cm *ChangelogManager) parseMarkdownChangelog(content string) (*Changelog, error) {
	changelog := &Changelog{
		Project:     "LLMrecon",
		Format:      "Markdown",
		LastUpdated: time.Now(),
		Entries:     make([]ChangelogEntry, 0),
	}
	
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	var currentEntry *ChangelogEntry
	var currentSection string
	
	// Regular expressions for parsing
	versionRegex := regexp.MustCompile(`^##\s+(?:Version\s+)?([^\s]+)(?:\s+[-(](.+)[)-])?`)
	dateRegex := regexp.MustCompile(`\*?(?:Released|Date):\s*([0-9]{4}-[0-9]{1,2}-[0-9]{1,2})\*?`)
	sectionRegex := regexp.MustCompile(`^###\s+(.+)`)
	listItemRegex := regexp.MustCompile(`^[-*]\s+(.+)`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and project title
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}
		
		// Check for version header
		if matches := versionRegex.FindStringSubmatch(line); matches != nil {
			// Save previous entry
			if currentEntry != nil {
				changelog.Entries = append(changelog.Entries, *currentEntry)
			}
			
			// Start new entry
			currentEntry = &ChangelogEntry{
				Version: matches[1],
				Type:    "release",
			}
			
			// Parse date from version line if present
			if len(matches) > 2 && matches[2] != "" {
				if date, err := time.Parse("2006-01-02", matches[2]); err == nil {
					currentEntry.Date = date
				}
			}
			
			currentSection = ""
			continue
		}
		
		// Check for date line
		if currentEntry != nil && currentEntry.Date.IsZero() {
			if matches := dateRegex.FindStringSubmatch(line); matches != nil {
				if date, err := time.Parse("2006-01-02", matches[1]); err == nil {
					currentEntry.Date = date
				}
				continue
			}
		}
		
		// Check for section header (Added, Changed, Fixed, etc.)
		if matches := sectionRegex.FindStringSubmatch(line); matches != nil {
			currentSection = strings.ToLower(matches[1])
			continue
		}
		
		// Check for list items
		if currentEntry != nil && currentSection != "" {
			if matches := listItemRegex.FindStringSubmatch(line); matches != nil {
				entry := ChangelogEntry{
					Version:     currentEntry.Version,
					Date:        currentEntry.Date,
					Type:        currentSection,
					Category:    cm.inferCategory(matches[1]),
					Title:       matches[1],
					Description: matches[1],
					Breaking:    cm.isBreakingChange(matches[1]),
					Security:    cm.isSecurityChange(matches[1]),
				}
				
				changelog.Entries = append(changelog.Entries, entry)
			}
		}
	}
	
	// Save last entry
	if currentEntry != nil {
		changelog.Entries = append(changelog.Entries, *currentEntry)
	}
	
	return changelog, nil
}

// parseJSONChangelog parses a JSON changelog or GitHub releases
func (cm *ChangelogManager) parseJSONChangelog(content []byte, source string) (*Changelog, error) {
	changelog := &Changelog{
		Project:     "LLMrecon",
		Format:      "JSON",
		LastUpdated: time.Now(),
		Entries:     make([]ChangelogEntry, 0),
	}
	
	// Check if it's GitHub releases format
	if strings.Contains(source, "/releases") {
		return cm.parseGitHubReleases(content)
	}
	
	// Try to parse as direct changelog JSON
	var changelogData Changelog
	if err := json.Unmarshal(content, &changelogData); err == nil {
		return &changelogData, nil
	}
	
	// Try to parse as array of entries
	var entries []ChangelogEntry
	if err := json.Unmarshal(content, &entries); err == nil {
		changelog.Entries = entries
		return changelog, nil
	}
	
	return nil, fmt.Errorf("unsupported JSON format")
}

// parseGitHubReleases parses GitHub releases API response
func (cm *ChangelogManager) parseGitHubReleases(content []byte) (*Changelog, error) {
	var releases []struct {
		TagName     string    `json:"tag_name"`
		Name        string    `json:"name"`
		Body        string    `json:"body"`
		Draft       bool      `json:"draft"`
		Prerelease  bool      `json:"prerelease"`
		CreatedAt   time.Time `json:"created_at"`
		PublishedAt time.Time `json:"published_at"`
	}
	
	if err := json.Unmarshal(content, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub releases: %w", err)
	}
	
	changelog := &Changelog{
		Project:     "LLMrecon",
		Format:      "GitHub Releases",
		LastUpdated: time.Now(),
		Entries:     make([]ChangelogEntry, 0),
	}
	
	for _, release := range releases {
		if release.Draft {
			continue
		}
		
		entry := ChangelogEntry{
			Version:     strings.TrimPrefix(release.TagName, "v"),
			Date:        release.PublishedAt,
			Type:        "release",
			Title:       release.Name,
			Description: release.Body,
			Security:    cm.isSecurityChange(release.Body),
			Breaking:    cm.isBreakingChange(release.Body),
		}
		
		if release.Prerelease {
			entry.Type = "prerelease"
		}
		
		changelog.Entries = append(changelog.Entries, entry)
	}
	
	return changelog, nil
}

// filterChangelogByVersion filters changelog entries for a specific version
func (cm *ChangelogManager) filterChangelogByVersion(changelog *Changelog, version string) *Changelog {
	filtered := &Changelog{
		Project:     changelog.Project,
		Format:      changelog.Format,
		LastUpdated: changelog.LastUpdated,
		Entries:     make([]ChangelogEntry, 0),
	}
	
	for _, entry := range changelog.Entries {
		if entry.Version == version {
			filtered.Entries = append(filtered.Entries, entry)
		}
	}
	
	return filtered
}

// displayChangelogEntry displays a single changelog entry
func (cm *ChangelogManager) displayChangelogEntry(entry ChangelogEntry) {
	icon := "•"
	switch entry.Type {
	case "added":
		icon = "✨"
	case "changed":
		icon = "🔄"
	case "fixed":
		icon = "🐛"
	case "removed":
		icon = "🗑️"
	case "security":
		icon = "🔒"
	case "deprecated":
		icon = "⚠️"
	}
	
	fmt.Printf("- %s %s", icon, entry.Title)
	
	if entry.Breaking {
		fmt.Printf(" **[BREAKING]**")
	}
	
	if entry.Security {
		fmt.Printf(" **[SECURITY]**")
	}
	
	fmt.Println()
	
	if entry.Description != entry.Title && entry.Description != "" {
		// Split description into lines and indent
		lines := strings.Split(entry.Description, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("  %s\n", strings.TrimSpace(line))
			}
		}
	}
	
	if len(entry.References) > 0 {
		fmt.Printf("  References: %s\n", strings.Join(entry.References, ", "))
	}
	
	fmt.Println()
}

// Helper functions

func (cm *ChangelogManager) inferCategory(description string) string {
	description = strings.ToLower(description)
	
	if strings.Contains(description, "template") {
		return "templates"
	}
	if strings.Contains(description, "module") || strings.Contains(description, "provider") {
		return "modules"
	}
	if strings.Contains(description, "cli") || strings.Contains(description, "command") {
		return "cli"
	}
	if strings.Contains(description, "api") {
		return "api"
	}
	if strings.Contains(description, "security") || strings.Contains(description, "vulnerability") {
		return "security"
	}
	if strings.Contains(description, "performance") {
		return "performance"
	}
	if strings.Contains(description, "ui") || strings.Contains(description, "interface") {
		return "ui"
	}
	
	return "general"
}

func (cm *ChangelogManager) isBreakingChange(description string) bool {
	breakingKeywords := []string{
		"breaking", "break", "incompatible", "removed", "deprecated",
		"changed api", "changed interface", "migration required",
	}
	
	description = strings.ToLower(description)
	for _, keyword := range breakingKeywords {
		if strings.Contains(description, keyword) {
			return true
		}
	}
	
	return false
}

func (cm *ChangelogManager) isSecurityChange(description string) bool {
	securityKeywords := []string{
		"security", "vulnerability", "cve", "exploit", "patch",
		"sensitive", "credential", "authentication", "authorization",
		"injection", "xss", "csrf", "dos",
	}
	
	description = strings.ToLower(description)
	for _, keyword := range securityKeywords {
		if strings.Contains(description, keyword) {
			return true
		}
	}
	
	return false
}

func (cm *ChangelogManager) compareVersions(v1, v2 string) int {
	// Simple version comparison - in production would use semver
	if v1 == v2 {
		return 0
	}
	if v1 > v2 {
		return 1
	}
	return -1
}

// CreateChangelogEntry creates a new changelog entry
func (cm *ChangelogManager) CreateChangelogEntry(entry ChangelogEntry) error {
	// This would be used for automated changelog generation
	changelogPath := "CHANGELOG.md"
	
	// Read existing changelog
	content := ""
	if data, err := os.ReadFile(changelogPath); err == nil {
		content = string(data)
	}
	
	// Insert new entry at the top
	newEntry := cm.formatChangelogEntry(entry)
	
	if content == "" {
		content = fmt.Sprintf("# LLMrecon Changelog\n\n%s", newEntry)
	} else {
		// Insert after the first header
		lines := strings.Split(content, "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
			lines = append(lines[:1], append([]string{"", newEntry}, lines[1:]...)...)
		} else {
			lines = append([]string{newEntry}, lines...)
		}
		content = strings.Join(lines, "\n")
	}
	
	// Write back to file
	return os.WriteFile(changelogPath, []byte(content), 0644)
}

// formatChangelogEntry formats a changelog entry for Markdown
func (cm *ChangelogManager) formatChangelogEntry(entry ChangelogEntry) string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("## Version %s\n", entry.Version))
	if !entry.Date.IsZero() {
		result.WriteString(fmt.Sprintf("*Released: %s*\n", entry.Date.Format("2006-01-02")))
	}
	result.WriteString("\n")
	
	// Group by type
	typeMap := map[string]string{
		"added":      "### Added",
		"changed":    "### Changed",
		"fixed":      "### Fixed",
		"removed":    "### Removed",
		"security":   "### Security",
		"deprecated": "### Deprecated",
	}
	
	if section, ok := typeMap[entry.Type]; ok {
		result.WriteString(section + "\n")
	}
	
	result.WriteString(fmt.Sprintf("- %s", entry.Title))
	
	if entry.Breaking {
		result.WriteString(" **[BREAKING]**")
	}
	
	if entry.Security {
		result.WriteString(" **[SECURITY]**")
	}
	
	result.WriteString("\n")
	
	if entry.Description != entry.Title && entry.Description != "" {
		result.WriteString(fmt.Sprintf("  %s\n", entry.Description))
	}
	
	result.WriteString("\n")
	
	return result.String()
}