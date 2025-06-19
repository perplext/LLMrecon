package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/perplext/LLMrecon/src/version"
)

var (
	// currentVersion is the current version of the tool
	// This should be set during build time with -ldflags
	currentVersion = "1.0.0"
)

// getLocalVersions returns the current versions of templates and modules
func getLocalVersions() (version.Version, map[string]version.Version, error) {
	// Default template version
	templateVersion := version.Version{Major: 1, Minor: 0, Patch: 0}
	moduleVersions := make(map[string]version.Version)

	// Try to read versions from local manifest files
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return templateVersion, moduleVersions, nil // Return defaults on error
	}

	// Read template version
	templateManifestPath := filepath.Join(homeDir, ".LLMrecon", "templates", "manifest.json")
	if data, err := os.ReadFile(templateManifestPath); err == nil {
		var manifest struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			if v, err := version.ParseVersion(manifest.Version); err == nil {
				templateVersion = v
			}
		}
	}

	// Read module versions
	modulesDir := filepath.Join(homeDir, ".LLMrecon", "modules")
	entries, err := os.ReadDir(modulesDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				moduleManifestPath := filepath.Join(modulesDir, entry.Name(), "manifest.json")
				if data, err := os.ReadFile(moduleManifestPath); err == nil {
					var manifest struct {
						ID      string `json:"id"`
						Version string `json:"version"`
					}
					if err := json.Unmarshal(data, &manifest); err == nil {
						if v, err := version.ParseVersion(manifest.Version); err == nil {
							moduleID := manifest.ID
							if moduleID == "" {
								moduleID = entry.Name()
							}
							moduleVersions[moduleID] = v
						}
					}
				}
			}
		}
	}

	return templateVersion, moduleVersions, nil
}

