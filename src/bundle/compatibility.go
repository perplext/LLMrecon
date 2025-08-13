package bundle

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/perplext/LLMrecon/src/version"
)

// CompatibilityResult contains the result of compatibility checks
type CompatibilityResult struct {
	Compatible bool                  `json:"compatible"`
	Issues     []CompatibilityIssue  `json:"issues"`
	Warnings   []string              `json:"warnings"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// CompatibilityIssue represents a compatibility problem
type CompatibilityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"` // error, warning, info
	Message     string `json:"message"`
	Component   string `json:"component,omitempty"`
	Action      string `json:"action,omitempty"`
	CanOverride bool   `json:"canOverride"`
}

// CompatibilityConfig defines compatibility checking configuration
type CompatibilityConfig struct {
	EnforceStrict         bool                   `json:"enforceStrict"`
	AllowDowngrade        bool                   `json:"allowDowngrade"`
	AllowPrerelease       bool                   `json:"allowPrerelease"`
	CheckVersion          bool                   `json:"checkVersion"`
	CheckComponents       bool                   `json:"checkComponents"`
	CheckEnvironment      bool                   `json:"checkEnvironment"`
	UpgradePaths          []UpgradePath          `json:"upgradePaths"`
	ComponentRequirements map[string]Requirement `json:"componentRequirements"`
	EnvironmentReqs       EnvironmentRequirement `json:"environmentRequirements"`
}

// UpgradePath defines a supported upgrade path
type UpgradePath struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Direct       bool   `json:"direct"`
	Intermediate string `json:"intermediate,omitempty"`
}

// Requirement defines version requirements for a component
type Requirement struct {
	MinVersion string `json:"minVersion"`
	MaxVersion string `json:"maxVersion,omitempty"`
	Required   bool   `json:"required"`
}

// EnvironmentRequirement defines environment requirements
type EnvironmentRequirement struct {
	Platforms      []Platform         `json:"platforms"`
	MinDiskSpace   int64              `json:"minDiskSpace"`
	MinMemory      int64              `json:"minMemory"`
	RuntimeDeps    map[string]string  `json:"runtimeDeps"`
}

// Platform defines a supported platform
type Platform struct {
	OS   string   `json:"os"`
	Arch []string `json:"arch"`
}

// Environment represents the current system environment
type Environment struct {
	OS                 string
	Arch               string
	Version            string
	AvailableDiskSpace int64
	AvailableMemory    int64
	IsProduction       bool
	InstalledPackages  map[string]string
}

// CompatibilityChecker performs compatibility verification
type CompatibilityChecker struct {
	currentVersion *version.SemVersion
	targetVersion  *version.SemVersion
	environment    *Environment
	config         *CompatibilityConfig
}

// NewCompatibilityChecker creates a new compatibility checker
func NewCompatibilityChecker(current, target string, env *Environment, config *CompatibilityConfig) (*CompatibilityChecker, error) {
	currentVer, err := version.ParseVersion(current)
	if err != nil {
		return nil, fmt.Errorf("invalid current version: %w", err)
	}

	targetVer, err := version.ParseVersion(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target version: %w", err)
	}

	if env == nil {
		env = DetectEnvironment()
	}

	if config == nil {
		config = DefaultCompatibilityConfig()
	}

	return &CompatibilityChecker{
		currentVersion: &currentVer,
		targetVersion:  &targetVer,
		environment:    env,
		config:         config,
	}, nil
}

// VerifyCompatibility performs all compatibility checks
func (c *CompatibilityChecker) VerifyCompatibility() (*CompatibilityResult, error) {
	result := &CompatibilityResult{
		Compatible: true,
		Issues:     []CompatibilityIssue{},
		Warnings:   []string{},
		Metadata:   make(map[string]interface{}),
	}

	// Add metadata
	result.Metadata["currentVersion"] = c.currentVersion.String()
	result.Metadata["targetVersion"] = c.targetVersion.String()
	result.Metadata["platform"] = fmt.Sprintf("%s/%s", c.environment.OS, c.environment.Arch)

	// Version compatibility
	if c.config.CheckVersion {
		if err := c.checkVersionCompatibility(result); err != nil {
			return nil, err
		}
	}

	// Component compatibility
	if c.config.CheckComponents {
		if err := c.checkComponentCompatibility(result); err != nil {
			return nil, err
		}
	}

	// Environment compatibility
	if c.config.CheckEnvironment {
		if err := c.checkEnvironmentCompatibility(result); err != nil {
			return nil, err
		}
	}

	// Feature compatibility
	if err := c.checkFeatureCompatibility(result); err != nil {
		return nil, err
	}

	return result, nil
}

// checkVersionCompatibility checks version compatibility
func (c *CompatibilityChecker) checkVersionCompatibility(result *CompatibilityResult) error {
	current := c.currentVersion
	target := c.targetVersion

	// Rule 1: No downgrades unless explicitly allowed
	if target.LessThan(current) && !c.config.AllowDowngrade {
		result.Compatible = false
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:        "version_downgrade",
			Severity:    "error",
			Message:     fmt.Sprintf("Cannot downgrade from %s to %s", current, target),
			CanOverride: true,
		})
	}

	// Rule 2: No prerelease versions unless allowed
	if target.Prerelease != "" && !c.config.AllowPrerelease {
		if c.environment.IsProduction {
			result.Compatible = false
			result.Issues = append(result.Issues, CompatibilityIssue{
				Type:        "prerelease_in_production",
				Severity:    "error",
				Message:     fmt.Sprintf("Cannot install prerelease version %s in production", target),
				CanOverride: true,
			})
		} else {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Installing prerelease version %s", target))
		}
	}

	// Rule 3: Major version changes require migration
	if target.Major != current.Major {
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:        "major_version_change",
			Severity:    "warning",
			Message:     fmt.Sprintf("Major version change from %d to %d may require migration", current.Major, target.Major),
			Action:      "Review migration guide before proceeding",
			CanOverride: false,
		})
	}

	// Rule 4: Check supported upgrade paths
	if !c.isUpgradePathSupported(current, target) {
		result.Compatible = false
		intermediate := c.findIntermediateVersion(current, target)
		action := "No upgrade path available"
		if intermediate != "" {
			action = fmt.Sprintf("Upgrade to %s first, then to %s", intermediate, target)
		}
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:        "unsupported_upgrade_path",
			Severity:    "error",
			Message:     fmt.Sprintf("Direct upgrade from %s to %s is not supported", current, target),
			Action:      action,
			CanOverride: false,
		})
	}

	return nil
}

// checkComponentCompatibility checks component compatibility
func (c *CompatibilityChecker) checkComponentCompatibility(result *CompatibilityResult) error {
	// This would check installed components against requirements
	// For now, we'll add a placeholder implementation
	
	for name, req := range c.config.ComponentRequirements {
		if req.Required {
			// Check if component exists and version is compatible
			// This is a simplified check - real implementation would query installed components
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Component '%s' compatibility check pending", name))
		}
	}

	return nil
}

// checkEnvironmentCompatibility checks environment compatibility
func (c *CompatibilityChecker) checkEnvironmentCompatibility(result *CompatibilityResult) error {
	env := c.environment
	reqs := c.config.EnvironmentReqs

	// Platform check
	platformSupported := false
	for _, platform := range reqs.Platforms {
		if platform.OS == env.OS {
			for _, arch := range platform.Arch {
				if arch == env.Arch {
					platformSupported = true
					break
				}
			}
		}
	}

	if !platformSupported {
		result.Compatible = false
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:        "unsupported_platform",
			Severity:    "error",
			Message:     fmt.Sprintf("Platform %s/%s is not supported", env.OS, env.Arch),
			CanOverride: false,
		})
	}

	// Disk space check
	if env.AvailableDiskSpace < reqs.MinDiskSpace {
		result.Compatible = false
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:     "insufficient_disk_space",
			Severity: "error",
			Message: fmt.Sprintf("Insufficient disk space: %s available, %s required",
				formatBytesCompat(env.AvailableDiskSpace), formatBytesCompat(reqs.MinDiskSpace)),
			CanOverride: false,
		})
	}

	// Memory check (warning only)
	if env.AvailableMemory < reqs.MinMemory {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Low memory: %s available, %s recommended",
				formatBytesCompat(env.AvailableMemory), formatBytesCompat(reqs.MinMemory)))
	}

	return nil
}

// checkFeatureCompatibility checks for feature-specific compatibility
func (c *CompatibilityChecker) checkFeatureCompatibility(result *CompatibilityResult) error {
	// Check for deprecated features
	deprecated := c.getDeprecatedFeatures()
	for _, feature := range deprecated {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Feature '%s' is deprecated and will be removed in version %s",
				feature.Name, feature.RemovalVersion))
	}

	// Check for breaking changes
	breaking := c.getBreakingChanges()
	for _, change := range breaking {
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:        "breaking_change",
			Severity:    "warning",
			Message:     change.Description,
			Action:      change.MigrationGuide,
			CanOverride: true,
		})
	}

	return nil
}

// isUpgradePathSupported checks if an upgrade path is supported
func (c *CompatibilityChecker) isUpgradePathSupported(from, to *version.SemVersion) bool {
	// Direct patch and minor updates are always supported
	if from.Major == to.Major {
		if from.Minor == to.Minor || (to.Minor == from.Minor+1) {
			return true
		}
	}

	// Check configured upgrade paths
	for _, path := range c.config.UpgradePaths {
		if c.versionMatches(from.String(), path.From) && c.versionMatches(to.String(), path.To) {
			return path.Direct
		}
	}

	return false
}

// findIntermediateVersion finds an intermediate version for multi-step upgrades
func (c *CompatibilityChecker) findIntermediateVersion(from, to *version.SemVersion) string {
	for _, path := range c.config.UpgradePaths {
		if c.versionMatches(from.String(), path.From) && c.versionMatches(to.String(), path.To) {
			if !path.Direct && path.Intermediate != "" {
				return path.Intermediate
			}
		}
	}
	return ""
}

// versionMatches checks if a version matches a pattern
func (c *CompatibilityChecker) versionMatches(version, pattern string) bool {
	// Simple pattern matching: supports x for wildcards
	// e.g., "1.x" matches "1.0", "1.1", etc.
	pattern = strings.ReplaceAll(pattern, "x", "*")
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = "^" + pattern + "$"
	
	// This is simplified - real implementation would use proper regex
	return strings.HasPrefix(version, strings.TrimSuffix(pattern, ".*$"))
}

// getDeprecatedFeatures returns deprecated features in the target version
func (c *CompatibilityChecker) getDeprecatedFeatures() []DeprecatedFeature {
	// This would be loaded from a configuration file or database
	return []DeprecatedFeature{
		{
			Name:           "legacy_templates",
			DeprecatedIn:   "2.0.0",
			RemovalVersion: "3.0.0",
			Alternative:    "Use new template format",
		},
	}
}

// getBreakingChanges returns breaking changes between versions
func (c *CompatibilityChecker) getBreakingChanges() []BreakingChange {
	// This would be loaded from a configuration file or database
	changes := []BreakingChange{}
	
	// Check for known breaking changes between versions
	if c.currentVersion.Major == 1 && c.targetVersion.Major == 2 {
		changes = append(changes, BreakingChange{
			Version:        "2.0.0",
			Description:    "Template format has changed to YAML v2",
			MigrationGuide: "Run 'llm-redteam migrate templates' to convert existing templates",
		})
	}
	
	return changes
}

// DeprecatedFeature represents a deprecated feature
type DeprecatedFeature struct {
	Name           string `json:"name"`
	DeprecatedIn   string `json:"deprecatedIn"`
	RemovalVersion string `json:"removalVersion"`
	Alternative    string `json:"alternative"`
}

// BreakingChange represents a breaking change
type BreakingChange struct {
	Version        string `json:"version"`
	Description    string `json:"description"`
	MigrationGuide string `json:"migrationGuide"`
}

// DetectEnvironment detects the current system environment
func DetectEnvironment() *Environment {
	env := &Environment{
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		Version:           runtime.Version(),
		InstalledPackages: make(map[string]string),
	}

	// Get disk space (simplified - real implementation would be platform-specific)
	env.AvailableDiskSpace = 10 * 1024 * 1024 * 1024 // 10GB placeholder

	// Get available memory (simplified)
	env.AvailableMemory = 4 * 1024 * 1024 * 1024 // 4GB placeholder

	// Check if production (simplified - check for env var)
	env.IsProduction = os.Getenv("LLM_REDTEAM_ENV") == "production"

	return env
}

// DefaultCompatibilityConfig returns the default compatibility configuration
func DefaultCompatibilityConfig() *CompatibilityConfig {
	return &CompatibilityConfig{
		EnforceStrict:    true,
		AllowDowngrade:   false,
		AllowPrerelease:  false,
		CheckVersion:     true,
		CheckComponents:  true,
		CheckEnvironment: true,
		UpgradePaths: []UpgradePath{
			{From: "1.0.x", To: "1.1.x", Direct: true},
			{From: "1.x", To: "2.0.x", Direct: false, Intermediate: "1.9.x"},
			{From: "2.0.x", To: "2.1.x", Direct: true},
		},
		ComponentRequirements: map[string]Requirement{
			"templates": {MinVersion: "1.0.0", Required: true},
			"providers": {MinVersion: "1.1.0", Required: true},
			"modules":   {MinVersion: "1.0.0", Required: false},
		},
		EnvironmentReqs: EnvironmentRequirement{
			Platforms: []Platform{
				{OS: "linux", Arch: []string{"amd64", "arm64"}},
				{OS: "darwin", Arch: []string{"amd64", "arm64"}},
				{OS: "windows", Arch: []string{"amd64"}},
			},
			MinDiskSpace: 1 * 1024 * 1024 * 1024,  // 1GB
			MinMemory:    512 * 1024 * 1024,       // 512MB
		},
	}
}

// formatBytesCompat formats bytes into human-readable format for compatibility reports
func formatBytesCompat(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ApplyOverrides applies compatibility overrides
func (c *CompatibilityChecker) ApplyOverrides(overrides map[string]bool) {
	if force, ok := overrides["force"]; ok && force {
		c.config.EnforceStrict = false
		c.config.AllowDowngrade = true
		c.config.AllowPrerelease = true
	}

	if skipVersion, ok := overrides["skipVersion"]; ok && skipVersion {
		c.config.CheckVersion = false
	}

	if skipComponents, ok := overrides["skipComponents"]; ok && skipComponents {
		c.config.CheckComponents = false
	}

	if skipEnv, ok := overrides["skipEnvironment"]; ok && skipEnv {
		c.config.CheckEnvironment = false
	}
}

// GetRecommendations provides recommendations based on compatibility issues
func GetRecommendations(result *CompatibilityResult) []string {
	recommendations := []string{}

	for _, issue := range result.Issues {
		switch issue.Type {
		case "version_downgrade":
			recommendations = append(recommendations,
				"Consider backing up your data before downgrading")
			if issue.CanOverride {
				recommendations = append(recommendations,
					"Use --allow-downgrade flag to override this check")
			}

		case "unsupported_platform":
			recommendations = append(recommendations,
				"Check the documentation for supported platforms")
			recommendations = append(recommendations,
				"Consider using Docker for platform-independent deployment")

		case "insufficient_disk_space":
			recommendations = append(recommendations,
				"Free up disk space or use a different installation directory")

		case "major_version_change":
			recommendations = append(recommendations,
				"Review the migration guide for breaking changes")
			recommendations = append(recommendations,
				"Consider testing the upgrade in a non-production environment first")
		}

		if issue.Action != "" {
			recommendations = append(recommendations, issue.Action)
		}
	}

	return recommendations
}