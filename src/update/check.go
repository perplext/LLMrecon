// Package update provides functionality for checking and applying updates
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/perplext/LLMrecon/src/version"
)

// UpdateInfo represents information about an available update
type UpdateInfo struct {
	Component       string
	CurrentVersion  version.Version
	LatestVersion   version.Version
	ChangeType      version.VersionChangeType
	ChangelogURL    string
	ReleaseDate     time.Time
	ReleaseNotes    string
	DownloadURL     string
	Signature       string
	ChecksumSHA256  string
	Required        bool                // Indicates if this update is required
	SecurityFixes   bool                // Indicates if this update contains security fixes
}

// VersionManifest represents the version information from the update server
type VersionManifest struct {
	Core struct {
		Version       string    `json:"version"`
		ReleaseDate   time.Time `json:"releaseDate"`
		ChangelogURL  string    `json:"changelogURL"`
		ReleaseNotes  string    `json:"releaseNotes"`
		DownloadURL   string    `json:"downloadURL"`
		Signature     string    `json:"signature"`
		ChecksumSHA256 string   `json:"checksumSHA256"`
	} `json:"core"`
	Templates struct {
		Version       string    `json:"version"`
		ReleaseDate   time.Time `json:"releaseDate"`
		ChangelogURL  string    `json:"changelogURL"`
		DownloadURL   string    `json:"downloadURL"`
		Signature     string    `json:"signature"`
		ChecksumSHA256 string   `json:"checksumSHA256"`
	} `json:"templates"`
	Modules []struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		Version       string    `json:"version"`
		ReleaseDate   time.Time `json:"releaseDate"`
		ChangelogURL  string    `json:"changelogURL"`
		DownloadURL   string    `json:"downloadURL"`
		Signature     string    `json:"signature"`
		ChecksumSHA256 string   `json:"checksumSHA256"`
	} `json:"modules"`
}

// UpdateVersionInfo contains information about the current and latest versions
type UpdateVersionInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	RequiredUpdate  bool
	SecurityFixes   bool
	ReleaseDate     string
	ChangelogURL    string
	DownloadURL     string
	Updates         []UpdateInfo
}

// VersionChecker provides functionality to check for updates
type VersionChecker struct {
	UpdateServerURL string
	HTTPClient      *http.Client
	CurrentVersions map[string]version.Version
	Notifier        UpdateNotifier
}

// UpdateNotifier defines the interface for notifying about updates
type UpdateNotifier interface {
	HandleUpdateCheck(ctx context.Context, versionInfo *UpdateVersionInfo) error
	NotifyUpdateSuccess(ctx context.Context, fromVersion, toVersion string) error
	NotifyUpdateFailure(ctx context.Context, fromVersion, toVersion string, err error) error
}

// NewVersionChecker creates a new VersionChecker
func NewVersionChecker(ctx context.Context) (*VersionChecker, error) {
	// Get current versions from the version package
	currentVersions := map[string]version.Version{
		"core": version.Version{Major: 1, Minor: 0, Patch: 0}, // Default version
	}

	// TODO: Add template and module versions

	return &VersionChecker{
		UpdateServerURL: "https://updates.LLMrecon.com", // Default update server URL
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		CurrentVersions: currentVersions,
	}, nil
}

// SetNotifier sets the notifier for the version checker
func (vc *VersionChecker) SetNotifier(notifier UpdateNotifier) {
	vc.Notifier = notifier
}

// CheckVersion checks if updates are available and returns version information
func (vc *VersionChecker) CheckVersion(ctx context.Context) (*UpdateVersionInfo, error) {
	updates, err := vc.CheckForUpdates()
	if err != nil {
		return nil, err
	}

	versionInfo := &UpdateVersionInfo{
		CurrentVersion:  func() string { v := vc.CurrentVersions["core"]; return (&v).String() }(),
		LatestVersion:   func() string { v := vc.CurrentVersions["core"]; return (&v).String() }(),
		UpdateAvailable: false,
		RequiredUpdate:  false,
		SecurityFixes:   false,
		ReleaseDate:     "",
		ChangelogURL:    "",
		DownloadURL:     "",
		Updates:         updates,
	}

	// Find the latest version among all components
	for _, update := range updates {
		versionInfo.UpdateAvailable = true

		// For core component, update the version info
		if update.Component == "core" {
			versionInfo.LatestVersion = update.LatestVersion.String()
			versionInfo.ReleaseDate = update.ReleaseDate.Format(time.RFC3339)
			versionInfo.ChangelogURL = update.ChangelogURL
			versionInfo.DownloadURL = update.DownloadURL
		}

		// Check if any update is required
		if update.Required {
			versionInfo.RequiredUpdate = true
		}

		// Check if any update has security fixes
		if update.SecurityFixes {
			versionInfo.SecurityFixes = true
		}
	}

	// Notify about updates if a notifier is set
	if vc.Notifier != nil && versionInfo.UpdateAvailable {
		if err := vc.Notifier.HandleUpdateCheck(ctx, versionInfo); err != nil {
			// Just log the error, don't fail the check
			fmt.Printf("Warning: Failed to send update notification: %v\n", err)
		}
	}

	return versionInfo, nil
}

// CheckForUpdates checks if updates are available for components
func (vc *VersionChecker) CheckForUpdates() ([]UpdateInfo, error) {
	manifest, err := vc.fetchVersionManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version manifest: %w", err)
	}

	updates := []UpdateInfo{}

	// Check core update
	coreCurrentVersion, hasCoreVersion := vc.CurrentVersions["core"]
	if hasCoreVersion {
		coreLatestVersion, err := version.ParseVersion(manifest.Core.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid core version in manifest: %w", err)
		}

		if coreLatestVersion.GreaterThan(&coreCurrentVersion) {
			updates = append(updates, UpdateInfo{
				Component:      "core",
				CurrentVersion: coreCurrentVersion,
				LatestVersion:  coreLatestVersion,
				ChangeType:     coreCurrentVersion.GetChangeType(&coreLatestVersion),
				ChangelogURL:   manifest.Core.ChangelogURL,
				ReleaseDate:    manifest.Core.ReleaseDate,
				ReleaseNotes:   manifest.Core.ReleaseNotes,
				DownloadURL:    manifest.Core.DownloadURL,
				Signature:      manifest.Core.Signature,
				ChecksumSHA256: manifest.Core.ChecksumSHA256,
				SecurityFixes: false, // This should be determined from the manifest
			})
		}
	}

	// Check templates update
	templatesCurrentVersion, hasTemplatesVersion := vc.CurrentVersions["templates"]
	if hasTemplatesVersion {
		templatesLatestVersion, err := version.ParseVersion(manifest.Templates.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid templates version in manifest: %w", err)
		}

		if templatesLatestVersion.GreaterThan(&templatesCurrentVersion) {
			updates = append(updates, UpdateInfo{
				Component:      "templates",
				CurrentVersion: templatesCurrentVersion,
				LatestVersion:  templatesLatestVersion,
				ChangeType:     templatesCurrentVersion.GetChangeType(&templatesLatestVersion),
				ChangelogURL:   manifest.Templates.ChangelogURL,
				ReleaseDate:    manifest.Templates.ReleaseDate,
				DownloadURL:    manifest.Templates.DownloadURL,
				Signature:      manifest.Templates.Signature,
				ChecksumSHA256: manifest.Templates.ChecksumSHA256,
				SecurityFixes: false, // This should be determined from the manifest
			})
		}
	}

	// Check module updates
	for _, moduleManifest := range manifest.Modules {
		moduleID := moduleManifest.ID
		moduleCurrentVersion, hasModuleVersion := vc.CurrentVersions[fmt.Sprintf("module.%s", moduleID)]
		
		if hasModuleVersion {
			moduleLatestVersion, err := version.ParseVersion(moduleManifest.Version)
			if err != nil {
				return nil, fmt.Errorf("invalid module version in manifest: %w", err)
			}

			if moduleLatestVersion.GreaterThan(&moduleCurrentVersion) {
				updates = append(updates, UpdateInfo{
					Component:      fmt.Sprintf("module.%s", moduleID),
					CurrentVersion: moduleCurrentVersion,
					LatestVersion:  moduleLatestVersion,
					ChangeType:     moduleCurrentVersion.GetChangeType(&moduleLatestVersion),
					ChangelogURL:   moduleManifest.ChangelogURL,
					ReleaseDate:    moduleManifest.ReleaseDate,
					DownloadURL:    moduleManifest.DownloadURL,
					Signature:      moduleManifest.Signature,
					ChecksumSHA256: moduleManifest.ChecksumSHA256,
					SecurityFixes: false, // This should be determined from the manifest
				})
			}
		}
	}

	return updates, nil
}

// fetchVersionManifest fetches the version manifest from the update server
func (vc *VersionChecker) fetchVersionManifest() (VersionManifest, error) {
	var manifest VersionManifest

	resp, err := vc.HTTPClient.Get(fmt.Sprintf("%s/versions.json", vc.UpdateServerURL))
	if err != nil {
		return manifest, fmt.Errorf("failed to connect to update server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return manifest, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return manifest, fmt.Errorf("failed to read response body: %w", err)
	}

	err = json.Unmarshal(body, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("failed to parse version manifest: %w", err)
	}

	return manifest, nil
}

// FormatUpdateInfo formats update information for display
func FormatUpdateInfo(updates []UpdateInfo) string {
	if len(updates) == 0 {
		return "All components are up to date."
	}

	result := "Updates available:\n\n"
	for _, update := range updates {
		result += fmt.Sprintf("Component: %s\n", update.Component)
		result += fmt.Sprintf("Current Version: %s\n", update.CurrentVersion.String())
		result += fmt.Sprintf("Latest Version: %s\n", update.LatestVersion.String())
		result += fmt.Sprintf("Change Type: %s\n", FormatChangeType(update.ChangeType))
		result += fmt.Sprintf("Release Date: %s\n", update.ReleaseDate.Format("2006-01-02"))
		
		if update.ReleaseNotes != "" {
			result += fmt.Sprintf("Release Notes: %s\n", update.ReleaseNotes)
		}
		
		if update.ChangelogURL != "" {
			result += fmt.Sprintf("Changelog: %s\n", update.ChangelogURL)
		}
		
		result += "\n"
	}

	return result
}

// FormatChangeType formats a VersionChangeType as a string
func FormatChangeType(changeType version.VersionChangeType) string {
	switch changeType {
	case version.MajorChange:
		return "Major Update"
	case version.MinorChange:
		return "Minor Update"
	case version.PatchChange:
		return "Patch Update"
	default:
		return "No Change"
	}
}

// MergeUpdates combines updates from multiple sources, with priority rules
func MergeUpdates(githubUpdates, gitlabUpdates []UpdateInfo) []UpdateInfo {
	// Create a map to store the highest priority update for each component
	updateMap := make(map[string]UpdateInfo)
	
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
	result := make([]UpdateInfo, 0, len(updateMap))
	for _, update := range updateMap {
		result = append(result, update)
	}
	
	return result
}
