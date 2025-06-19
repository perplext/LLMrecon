// Package version provides utilities for semantic versioning
package version

// Current version of the framework
const FrameworkVersion = "0.1.0"

// FrameworkInfo contains information about the framework
type FrameworkInfo struct {
	// Version is the current version of the framework
	Version string
	// MinPluginVersion is the minimum plugin version supported
	MinPluginVersion string
	// MaxPluginVersion is the maximum plugin version supported
	MaxPluginVersion string
}

// GetFrameworkInfo returns information about the framework
func GetFrameworkInfo() *FrameworkInfo {
	return &FrameworkInfo{
		Version:         Version,
		MinPluginVersion: "0.1.0",
		MaxPluginVersion: "", // No maximum version constraint
	}
}

// IsPluginVersionCompatible checks if a plugin version is compatible with the framework
func IsPluginVersionCompatible(pluginVersion, minFrameworkVersion, maxFrameworkVersion string) (bool, error) {
	// Parse versions
	frameworkVer, err := ParseVersion(Version)
	if err != nil {
		return false, err
	}
	
	// We don't actually need to use pluginVersion for compatibility check
	// but we validate it's a valid semver
	_, err = ParseVersion(pluginVersion)
	if err != nil {
		return false, err
	}
	
	// Check minimum framework version
	if minFrameworkVersion != "" {
		minVer, err := ParseVersion(minFrameworkVersion)
		if err != nil {
			return false, err
		}
		
		if Compare(frameworkVer, minVer) < 0 {
			return false, nil
		}
	}
	
	// Check maximum framework version
	if maxFrameworkVersion != "" {
		maxVer, err := ParseVersion(maxFrameworkVersion)
		if err != nil {
			return false, err
		}
		
		if Compare(frameworkVer, maxVer) > 0 {
			return false, nil
		}
	}
	
	return true, nil
}
