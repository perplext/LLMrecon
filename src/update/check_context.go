package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/perplext/LLMrecon/src/version"
)

// NewVersionCheckerWithURL creates a new VersionChecker with the given update server URL and current versions
func NewVersionCheckerWithURL(updateServerURL string, currentVersions map[string]version.Version) *VersionChecker {
	return &VersionChecker{
		UpdateServerURL: updateServerURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		CurrentVersions: currentVersions,
	}
}

// CheckForUpdatesContext checks for updates with context support
func (vc *VersionChecker) CheckForUpdatesContext(ctx context.Context) ([]ExtendedUpdateInfo, error) {
	// Fetch version manifest from update server
	manifest, err := vc.fetchEnhancedVersionManifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching version manifest: %w", err)
	}

	var updates []ExtendedUpdateInfo

	// Check core/binary updates
	if coreVersion, ok := vc.CurrentVersions["binary"]; ok {
		if manifest.Core.Version != "" {
			latestCore, err := version.ParseVersion(manifest.Core.Version)
			if err == nil && latestCore.GreaterThan(&coreVersion) {
				updates = append(updates, ExtendedUpdateInfo{
					Component:       "binary",
					CurrentVersion:  coreVersion,
					LatestVersion:   latestCore,
					ChangeType:      coreVersion.GetChangeType(&latestCore),
					ReleaseDate:     manifest.Core.ReleaseDate,
					ReleaseNotes:    manifest.Core.ReleaseNotes,
					DownloadURL:     manifest.Core.DownloadURL,
					ChecksumSHA256:  manifest.Core.ChecksumSHA256,
					Signature:       manifest.Core.Signature,
					Size:            0, // Not available in manifest
					Dependencies:    []string{}, // Not available in manifest
					BreakingChanges: false, // Not available in manifest
				})
			}
		}
	}

	// Check template updates
	if templateVersion, ok := vc.CurrentVersions["templates"]; ok {
		if manifest.Templates.Version != "" {
			latestTemplates, err := version.ParseVersion(manifest.Templates.Version)
			if err == nil && latestTemplates.GreaterThan(&templateVersion) {
				updates = append(updates, ExtendedUpdateInfo{
					Component:       "templates",
					CurrentVersion:  templateVersion,
					LatestVersion:   latestTemplates,
					ChangeType:      templateVersion.GetChangeType(&latestTemplates),
					ReleaseDate:     manifest.Templates.ReleaseDate,
					ReleaseNotes:    "", // Not available in manifest
					DownloadURL:     manifest.Templates.DownloadURL,
					ChecksumSHA256:  manifest.Templates.ChecksumSHA256,
					Signature:       manifest.Templates.Signature,
					Size:            0, // Not available in manifest
					Dependencies:    []string{}, // Not available in manifest
					BreakingChanges: false, // Not available in manifest
				})
			}
		}
	}

	// Check module updates
	for _, module := range manifest.Modules {
		moduleKey := fmt.Sprintf("module.%s", module.ID)
		if currentVersion, ok := vc.CurrentVersions[moduleKey]; ok {
			latestVersion, err := version.ParseVersion(module.Version)
			if err == nil && latestVersion.GreaterThan(&currentVersion) {
				updates = append(updates, ExtendedUpdateInfo{
					Component:       moduleKey,
					CurrentVersion:  currentVersion,
					LatestVersion:   latestVersion,
					ChangeType:      currentVersion.GetChangeType(&latestVersion),
					ReleaseDate:     module.ReleaseDate,
					ReleaseNotes:    "", // Not available in manifest
					DownloadURL:     module.DownloadURL,
					ChecksumSHA256:  module.ChecksumSHA256,
					Signature:       module.Signature,
					Size:            0, // Not available in manifest
					Dependencies:    []string{}, // Not available in manifest
					BreakingChanges: false, // Not available in manifest
				})
			}
		}
	}

	return updates, nil
}

// fetchEnhancedVersionManifest fetches the enhanced version manifest from the update server
func (vc *VersionChecker) fetchEnhancedVersionManifest(ctx context.Context) (*EnhancedVersionManifest, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", vc.UpdateServerURL+"/manifest.json", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := vc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var manifest EnhancedVersionManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decoding manifest: %w", err)
	}

	return &manifest, nil
}

// EnhancedVersionManifest extends VersionManifest with additional fields
type EnhancedVersionManifest struct {
	Core struct {
		Version         string    `json:"version"`
		ReleaseDate     time.Time `json:"releaseDate"`
		ChangelogURL    string    `json:"changelogURL"`
		ReleaseNotes    string    `json:"releaseNotes"`
		DownloadURL     string    `json:"downloadURL"`
		Signature       string    `json:"signature"`
		ChecksumSHA256  string    `json:"checksumSHA256"`
		Size            int64     `json:"size"`
		Dependencies    []string  `json:"dependencies"`
		BreakingChanges bool      `json:"breakingChanges"`
	} `json:"core"`
	Templates struct {
		Version         string    `json:"version"`
		ReleaseDate     time.Time `json:"releaseDate"`
		ChangelogURL    string    `json:"changelogURL"`
		ReleaseNotes    string    `json:"releaseNotes"`
		DownloadURL     string    `json:"downloadURL"`
		Signature       string    `json:"signature"`
		ChecksumSHA256  string    `json:"checksumSHA256"`
		Size            int64     `json:"size"`
		Dependencies    []string  `json:"dependencies"`
		BreakingChanges bool      `json:"breakingChanges"`
	} `json:"templates"`
	Modules []struct {
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		Version         string    `json:"version"`
		ReleaseDate     time.Time `json:"releaseDate"`
		ChangelogURL    string    `json:"changelogURL"`
		ReleaseNotes    string    `json:"releaseNotes"`
		DownloadURL     string    `json:"downloadURL"`
		Signature       string    `json:"signature"`
		ChecksumSHA256  string    `json:"checksumSHA256"`
		Size            int64     `json:"size"`
		Dependencies    []string  `json:"dependencies"`
		BreakingChanges bool      `json:"breakingChanges"`
	} `json:"modules"`
}

// ExtendedUpdateInfo represents information about an available update with additional fields
type ExtendedUpdateInfo struct {
	Component       string                    `json:"component"`
	CurrentVersion  version.Version           `json:"current_version"`
	LatestVersion   version.Version           `json:"latest_version"`
	ChangeType      version.VersionChangeType `json:"change_type"`
	ChangelogURL    string                    `json:"changelog_url"`
	ReleaseDate     time.Time                 `json:"release_date"`
	ReleaseNotes    string                    `json:"release_notes"`
	DownloadURL     string                    `json:"download_url"`
	Signature       string                    `json:"signature"`
	ChecksumSHA256  string                    `json:"checksum_sha256"`
	Required        bool                      `json:"required"`
	SecurityFixes   bool                      `json:"security_fixes"`
	Size            int64                     `json:"size"`
	Dependencies    []string                  `json:"dependencies"`
	BreakingChanges bool                      `json:"breaking_changes"`
}

// MergeExtendedUpdates merges and deduplicates extended update lists
func MergeExtendedUpdates(updateLists ...[]ExtendedUpdateInfo) []ExtendedUpdateInfo {
	seen := make(map[string]ExtendedUpdateInfo)
	
	for _, updates := range updateLists {
		for _, update := range updates {
			key := fmt.Sprintf("%s-%s", update.Component, update.LatestVersion.String())
			if existing, ok := seen[key]; !ok || update.ReleaseDate.After(existing.ReleaseDate) {
				seen[key] = update
			}
		}
	}
	
	result := make([]ExtendedUpdateInfo, 0, len(seen))
	for _, update := range seen {
		result = append(result, update)
	}
	
	return result
}