package update

import (
	"os"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ComponentVersionChecker handles version checking for all components
type ComponentVersionChecker struct {
	config *Config
	client *http.Client
	logger Logger
}

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
		ContentType        string `json:"content_type"`
	} `json:"assets"`
	HTMLURL string `json:"html_url"`

// GitLabRelease represents a GitLab release response
type GitLabRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	ReleasedAt  string `json:"released_at"`
	Assets      struct {
		Links []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"links"`
	} `json:"assets"`
	WebURL string `json:"web_url"`

// TemplateManifest represents a template repository manifest
type TemplateManifest struct {
	Version     string                    `json:"version"`
	LastUpdated time.Time                 `json:"last_updated"`
	Templates   map[string]TemplateInfo   `json:"templates"`
	Categories  map[string][]string       `json:"categories"`
	Statistics  map[string]interface{}    `json:"statistics"`
	Repository  RepositoryInfo            `json:"repository"`

// ModuleManifest represents a module repository manifest
type ModuleManifest struct {
	Version     string                  `json:"version"`
	LastUpdated time.Time               `json:"last_updated"`
	Modules     map[string]ModuleInfo   `json:"modules"`
	Providers   map[string][]string     `json:"providers"`
	Statistics  map[string]interface{}  `json:"statistics"`
	Repository  RepositoryMetadata          `json:"repository"`


// RepositoryMetadata contains metadata about a repository
type RepositoryMetadata struct {
	URL         string    `json:"url"`
	Branch      string    `json:"branch"`
	Commit      string    `json:"commit"`
	LastUpdated time.Time `json:"last_updated"`
	Size        int64     `json:"size"`
	FileCount   int       `json:"file_count"`

// NewComponentVersionChecker creates a new component version checker
func NewComponentVersionChecker(config *Config, logger Logger) *ComponentVersionChecker {
	client := &http.Client{
		Timeout: config.Timeout,
	}
	
	return &ComponentVersionChecker{
		config: config,
		client: client,
		logger: logger,
	}

// CheckBinaryUpdate checks for binary updates
func (vc *ComponentVersionChecker) CheckBinaryUpdate(ctx context.Context) (*ComponentUpdate, error) {
	vc.logger.Debug("Checking for binary updates...")
	
	// Get current version
	currentVersion, err := GetCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}
	
	// Get latest release
	latestRelease, err := vc.GetLatestBinaryRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}
	
	// Compare versions
	current, err := semver.NewVersion(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return nil, fmt.Errorf("invalid current version %s: %w", currentVersion, err)
	}
	
	latest, err := semver.NewVersion(strings.TrimPrefix(latestRelease.Version, "v"))
	if err != nil {
		return nil, fmt.Errorf("invalid latest version %s: %w", latestRelease.Version, err)
	}
	
	update := &ComponentUpdate{
		Component:      ComponentBinary,
		CurrentVersion: currentVersion,
		LatestVersion:  latestRelease.Version,
		ChangelogURL:   latestRelease.ChangelogURL,
		ReleaseNotes:   latestRelease.Description,
		Available:      latest.GreaterThan(current),
	}
	
	// Calculate update size
	if update.Available {
		asset := vc.selectBestAsset(latestRelease.Assets)
		if asset != nil {
			update.UpdateSize = asset.Size
		}
		
		// Check if this is a security update
		update.SecurityUpdate = vc.isSecurityUpdate(latestRelease.Description)
		update.Critical = vc.isCriticalUpdate(latestRelease.Description)
	}
	
	return update, nil

// CheckTemplateUpdates checks for template updates
func (vc *ComponentVersionChecker) CheckTemplateUpdates(ctx context.Context) (*ComponentUpdate, error) {
	vc.logger.Debug("Checking for template updates...")
	
	// Get current template versions
	currentVersions := vc.getCurrentTemplateVersions()
	
	update := &ComponentUpdate{
		Component:      ComponentTemplates,
		CurrentVersion: formatVersionMap(currentVersions),
		Available:      false,
	}
	
	var latestVersions map[string]string
	var totalSize int64
	
	// Check each repository
	for _, repo := range vc.config.TemplateRepos {
		repoVersions, size, err := vc.checkRepositoryUpdates(ctx, repo, "templates")
		if err != nil {
			vc.logger.Error(fmt.Sprintf("Failed to check %s repository", repo.Name), err)
			continue
		}
		
		totalSize += size
		
		// Merge versions
		if latestVersions == nil {
			latestVersions = make(map[string]string)
		}
		for name, version := range repoVersions {
			latestVersions[name] = version
		}
	}
	
	if latestVersions != nil {
		update.LatestVersion = formatVersionMap(latestVersions)
		update.UpdateSize = totalSize
		
		// Check if updates are available
		update.Available = vc.hasUpdates(currentVersions, latestVersions)
	}
	
	return update, nil

// CheckModuleUpdates checks for module updates
func (vc *ComponentVersionChecker) CheckModuleUpdates(ctx context.Context) (*ComponentUpdate, error) {
	vc.logger.Debug("Checking for module updates...")
	
	// Get current module versions
	currentVersions := vc.getCurrentModuleVersions()
	
	update := &ComponentUpdate{
		Component:      ComponentModules,
		CurrentVersion: formatVersionMap(currentVersions),
		Available:      false,
	}
	
	var latestVersions map[string]string
	var totalSize int64
	
	// Check each repository
	for _, repo := range vc.config.ModuleRepos {
		repoVersions, size, err := vc.checkRepositoryUpdates(ctx, repo, "modules")
		if err != nil {
			vc.logger.Error(fmt.Sprintf("Failed to check %s repository", repo.Name), err)
			continue
		}
		
		totalSize += size
		
		// Merge versions
		if latestVersions == nil {
			latestVersions = make(map[string]string)
		}
		for name, version := range repoVersions {
			latestVersions[name] = version
		}
	}
	
	if latestVersions != nil {
		update.LatestVersion = formatVersionMap(latestVersions)
		update.UpdateSize = totalSize
		
		// Check if updates are available
		update.Available = vc.hasUpdates(currentVersions, latestVersions)
	}
	
	return update, nil

// GetLatestBinaryRelease gets the latest binary release
func (vc *ComponentVersionChecker) GetLatestBinaryRelease(ctx context.Context) (*Release, error) {
	// Parse repository URL to determine provider
	if strings.Contains(vc.config.BinaryRepo, "github.com") {
		return vc.getLatestGitHubRelease(ctx)
	} else if strings.Contains(vc.config.BinaryRepo, "gitlab.com") {
		return vc.getLatestGitLabRelease(ctx)
	}
	
	return nil, fmt.Errorf("unsupported repository provider")

// getLatestGitHubRelease gets the latest release from GitHub
func (vc *ComponentVersionChecker) getLatestGitHubRelease(ctx context.Context) (*Release, error) {
	url := vc.config.BinaryUpdateURL
	if url == "" {
		// Construct URL from repository
		parts := strings.Split(strings.TrimPrefix(vc.config.BinaryRepo, "github.com/"), "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid GitHub repository format")
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", parts[0], parts[1])
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", vc.config.UserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	// Add authentication if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	resp, err := vc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}
	
	var ghRelease GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert to internal Release format
	release := &Release{
		Version:      strings.TrimPrefix(ghRelease.TagName, "v"),
		Name:         ghRelease.Name,
		Description:  ghRelease.Body,
		ChangelogURL: ghRelease.HTMLURL,
		TagName:      ghRelease.TagName,
		Prerelease:   ghRelease.Prerelease,
		Draft:        ghRelease.Draft,
		Assets:       make([]ReleaseAsset, len(ghRelease.Assets)),
	}
	
	// Parse release date
	if ghRelease.PublishedAt != "" {
		if releaseDate, err := time.Parse(time.RFC3339, ghRelease.PublishedAt); err == nil {
			release.ReleaseDate = releaseDate
		}
	}
	
	// Convert assets
	for i, asset := range ghRelease.Assets {
		release.Assets[i] = ReleaseAsset{
			Name:         asset.Name,
			DownloadURL:  asset.BrowserDownloadURL,
			Size:         asset.Size,
			ContentType:  asset.ContentType,
			Platform:     vc.extractPlatform(asset.Name),
			Architecture: vc.extractArchitecture(asset.Name),
		}
	}
	
	return release, nil

// getLatestGitLabRelease gets the latest release from GitLab
func (vc *ComponentVersionChecker) getLatestGitLabRelease(ctx context.Context) (*Release, error) {
	// Similar implementation for GitLab API
	return nil, fmt.Errorf("GitLab releases not yet implemented")

// checkRepositoryUpdates checks updates for a repository
func (vc *ComponentVersionChecker) checkRepositoryUpdates(ctx context.Context, repo RepositoryConfig, component string) (map[string]string, int64, error) {
	switch repo.Type {
	case RepositoryTypeGitHub:
		return vc.checkGitHubRepository(ctx, repo, component)
	case RepositoryTypeGitLab:
		return vc.checkGitLabRepository(ctx, repo, component)
	case RepositoryTypeHTTP:
		return vc.checkHTTPRepository(ctx, repo, component)
	default:
		return nil, 0, fmt.Errorf("unsupported repository type: %s", repo.Type)
	}

// checkGitHubRepository checks a GitHub repository for updates
func (vc *ComponentVersionChecker) checkGitHubRepository(ctx context.Context, repo RepositoryConfig, component string) (map[string]string, int64, error) {
	// Construct manifest URL
	manifestURL := fmt.Sprintf("%s/raw/%s/manifest.json", repo.URL, repo.Branch)
	
	req, err := http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", vc.config.UserAgent)
	
	// Add authentication if available
	if repo.Token != "" {
		req.Header.Set("Authorization", "Bearer "+repo.Token)
	} else if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	resp, err := vc.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("manifest not found: %d", resp.StatusCode)
	}
	
	// Parse manifest based on component type
	versions := make(map[string]string)
	var totalSize int64
	
	if component == "templates" {
		var manifest TemplateManifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			return nil, 0, fmt.Errorf("failed to decode template manifest: %w", err)
		}
		
		for name, template := range manifest.Templates {
			versions[name] = template.Version
			totalSize += template.Size
		}
	} else if component == "modules" {
		var manifest ModuleManifest
		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			return nil, 0, fmt.Errorf("failed to decode module manifest: %w", err)
		}
		
		for name, module := range manifest.Modules {
			versions[name] = module.Version
			totalSize += module.Size
		}
	}
	
	return versions, totalSize, nil

// checkGitLabRepository checks a GitLab repository for updates
func (vc *ComponentVersionChecker) checkGitLabRepository(ctx context.Context, repo RepositoryConfig, component string) (map[string]string, int64, error) {
	// Similar implementation for GitLab
	return nil, 0, fmt.Errorf("GitLab repositories not yet implemented")

// checkHTTPRepository checks an HTTP repository for updates
func (vc *ComponentVersionChecker) checkHTTPRepository(ctx context.Context, repo RepositoryConfig, component string) (map[string]string, int64, error) {
	// Implementation for HTTP-based repositories
	return nil, 0, fmt.Errorf("HTTP repositories not yet implemented")

// getCurrentTemplateVersions gets current template versions
func (vc *ComponentVersionChecker) getCurrentTemplateVersions() map[string]string {
	versions := make(map[string]string)
	
	templateDir := vc.config.TemplateDirectory
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return versions
	}
	
	// Walk through template directory
	filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			// Read template file and extract version
			if version := vc.extractTemplateVersion(path); version != "" {
				relPath, _ := filepath.Rel(templateDir, path)
				versions[relPath] = version
			}
		}
		
		return nil
	})
	
	return versions

// getCurrentModuleVersions gets current module versions
func (vc *ComponentVersionChecker) getCurrentModuleVersions() map[string]string {
	versions := make(map[string]string)
	
	moduleDir := vc.config.ModuleDirectory
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return versions
	}
	
	// Walk through module directory
	filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() {
			// Read module file and extract version
			if version := vc.extractModuleVersion(path); version != "" {
				relPath, _ := filepath.Rel(moduleDir, path)
				versions[relPath] = version
			}
		}
		
		return nil
	})
	
	return versions

// extractTemplateVersion extracts version from a template file
func (vc *ComponentVersionChecker) extractTemplateVersion(filePath string) string {
	// Simple implementation - read file and look for version field
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return ""
	}
	
	// Look for version in YAML format
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "version:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			}
		}
	}
	
	return "unknown"

// extractModuleVersion extracts version from a module file
func (vc *ComponentVersionChecker) extractModuleVersion(filePath string) string {
	// Similar to template version extraction
	return vc.extractTemplateVersion(filePath)

// hasUpdates checks if there are any updates available
func (vc *ComponentVersionChecker) hasUpdates(current, latest map[string]string) bool {
	for name, latestVersion := range latest {
		currentVersion, exists := current[name]
		if !exists {
			return true // New item
		}
		
		if currentVersion != latestVersion {
			return true // Version changed
		}
	}
	
	return false

// selectBestAsset selects the best asset for the current platform
func (vc *ComponentVersionChecker) selectBestAsset(assets []ReleaseAsset) *ReleaseAsset {
	currentPlatform := GetPlatformString()
	currentArch := GetArchString()
	
	// First, try to find exact match
	for _, asset := range assets {
		if asset.Platform == currentPlatform && asset.Architecture == currentArch {
			return &asset
		}
	}
	
	// Fallback to name-based matching
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, currentPlatform) && strings.Contains(name, currentArch) {
			return &asset
		}
	}
	
	return nil

// extractPlatform extracts platform from asset name
func (vc *ComponentVersionChecker) extractPlatform(name string) string {
	name = strings.ToLower(name)
	
	if strings.Contains(name, "linux") {
		return PlatformLinux
	}
	if strings.Contains(name, "darwin") || strings.Contains(name, "macos") {
		return PlatformDarwin
	}
	if strings.Contains(name, "windows") || strings.Contains(name, "win") {
		return PlatformWindows
	}
	if strings.Contains(name, "freebsd") {
		return PlatformFreeBSD
	}
	
	return ""

// extractArchitecture extracts architecture from asset name
func (vc *ComponentVersionChecker) extractArchitecture(name string) string {
	name = strings.ToLower(name)
	
	if strings.Contains(name, "amd64") || strings.Contains(name, "x64") {
		return ArchAMD64
	}
	if strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") {
		return ArchARM64
	}
	if strings.Contains(name, "386") || strings.Contains(name, "x86") {
		return Arch386
	}
	if strings.Contains(name, "arm") {
		return ArchARM
	}
	
	return ""

// isSecurityUpdate checks if a release is a security update
func (vc *ComponentVersionChecker) isSecurityUpdate(description string) bool {
	securityKeywords := []string{
		"security", "vulnerability", "cve-", "exploit", "patch",
		"critical", "urgent", "hotfix", "security fix",
	}
	
	description = strings.ToLower(description)
	for _, keyword := range securityKeywords {
		if strings.Contains(description, keyword) {
			return true
		}
	}
	
	return false

// isCriticalUpdate checks if a release is critical
func (vc *ComponentVersionChecker) isCriticalUpdate(description string) bool {
	criticalKeywords := []string{
		"critical", "urgent", "breaking", "major", "important",
		"hotfix", "emergency", "severe",
	}
	
	description = strings.ToLower(description)
	for _, keyword := range criticalKeywords {
		if strings.Contains(description, keyword) {
			return true
		}
	}
	
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
