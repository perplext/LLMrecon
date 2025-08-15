// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"encoding/json"
	"fmt"
)

// LoadBundle loads a bundle from the specified path
func LoadBundle(bundlePath string) (*Bundle, error) {
	// Validate bundle path
	info, err := os.Stat(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access bundle path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("bundle path is not a directory: %s", bundlePath)
	}

	// Load manifest
	manifestPath := filepath.Join(bundlePath, "manifest.json")
	manifestData, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Unmarshal manifest
	var manifest BundleManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// Create bundle
	bundle := &Bundle{
		Manifest:   manifest,
		BundlePath: bundlePath,
		IsVerified: false,
	}

	return bundle, nil

// SaveBundle saves a bundle to the specified path
func SaveBundle(bundle *Bundle) error {
	// Create bundle directory if it doesn't exist
	if err := os.MkdirAll(bundle.BundlePath, 0700); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}

	// Marshal manifest
	manifestData, err := json.MarshalIndent(bundle.Manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write manifest
	manifestPath := filepath.Join(bundle.BundlePath, "manifest.json")
	if err := os.WriteFile(filepath.Clean(manifestPath, manifestData, 0600)); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil

// CreateEmptyBundle creates a new empty bundle
func CreateEmptyBundle(bundlePath string, schemaVersion string, bundleID string, bundleType BundleType, name string, description string, version string) (*Bundle, error) {
	// Validate parameters
	if bundlePath == "" {
		return nil, fmt.Errorf("bundle path is required")
	}
	if bundleID == "" {
		return nil, fmt.Errorf("bundle ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("bundle name is required")
	}
	if version == "" {
		return nil, fmt.Errorf("bundle version is required")
	}

	// Create manifest
	manifest := BundleManifest{
		SchemaVersion: schemaVersion,
		BundleID:      bundleID,
		BundleType:    bundleType,
		Name:          name,
		Description:   description,
		Version:       version,
		CreatedAt:     time.Now(),
		Content:       []ContentItem{},
		Checksums: Checksums{
			Manifest: "",
			Content:  make(map[string]string),
		},
		Compatibility: Compatibility{
			MinVersion: "1.0.0",
		},
	}

	// Create bundle
	bundle := &Bundle{
		Manifest:   manifest,
		BundlePath: bundlePath,
		IsVerified: false,
	}

