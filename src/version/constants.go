// Package version provides utilities for semantic versioning
package version

// FrameworkVersion is the current version of the framework
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
		Version:          FrameworkVersion,
		MinPluginVersion: "0.1.0",
		MaxPluginVersion: "0.2.0",
	}
}

// IsPluginVersionCompatible checks if a plugin version is compatible with the framework
func IsPluginVersionCompatible(pluginVersion, minFrameworkVersion, maxFrameworkVersion string) (bool, error) {
	// Parse plugin version
	plugin, err := Parse(pluginVersion)
	if err != nil {
		return false, err
	}
	
	// Parse min framework version
	minFramework, err := Parse(minFrameworkVersion)
	if err != nil {
		return false, err
	}
	
	// Parse max framework version if provided
	var maxFramework *SemVersion
	if maxFrameworkVersion != "" {
		maxFramework, err = Parse(maxFrameworkVersion)
		if err != nil {
			return false, err
		}
	}
	
	// Check if plugin version is compatible with framework version
	if plugin.LessThan(minFramework) {
		return false, nil
	}
	
	if maxFramework != nil && plugin.GreaterThan(maxFramework) {
		return false, nil
	}
	
	return true, nil
}
