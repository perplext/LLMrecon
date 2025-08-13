// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
)

// OfflineBundleCreator creates offline bundles
type OfflineBundleCreator struct {
	// Generator is the manifest generator
	Generator *ManifestGenerator
	// Validator is the offline bundle validator
	Validator *OfflineBundleValidator
	// Format is the offline bundle format
	Format *OfflineBundleFormat
	// Logger is the logger for bundle creation operations
	Logger io.Writer
	// AuditTrail is the audit trail manager for logging operations
	AuditTrail *trail.AuditTrailManager
}

// NewOfflineBundleCreator creates a new offline bundle creator
func NewOfflineBundleCreator(signingKey ed25519.PrivateKey, author Author, logger io.Writer, auditTrail *trail.AuditTrailManager) *OfflineBundleCreator {
	if logger == nil {
		logger = os.Stdout
	}

	return &OfflineBundleCreator{
		Generator: NewManifestGenerator(signingKey, author),
		Validator: NewOfflineBundleValidator(logger),
		Format:    DefaultOfflineBundleFormat(),
		Logger:    logger,
		AuditTrail: auditTrail,
	}
}

// CreateOfflineBundle creates a new offline bundle
func (c *OfflineBundleCreator) CreateOfflineBundle(name, description, version string, bundleType BundleType, outputPath string) (*OfflineBundle, error) {
	// Log creation start
	fmt.Fprintf(c.Logger, "Creating offline bundle: %s (version: %s)\n", name, version)

	// Create enhanced manifest
	manifest := c.Generator.GenerateEnhancedManifest(name, description, version, bundleType)

	// Create output directory
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create required directories
	for _, dir := range c.Format.RequiredDirectories {
		dirPath := filepath.Join(outputPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create manifest file
	manifestPath := filepath.Join(outputPath, "manifest.json")
	if err := c.Generator.WriteEnhancedManifest(manifest, manifestPath); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Create README.md file with basic information
	readmePath := filepath.Join(outputPath, "README.md")
	readmeContent := fmt.Sprintf(`# %s

%s

## Version

%s

## Bundle Type

%s

## Created

%s

## Author

%s (%s)

## Contents

This offline bundle contains templates and modules for LLM red teaming.

## Usage

See the documentation directory for usage instructions.
`, name, description, version, bundleType, manifest.CreatedAt.Format(time.RFC3339),
		manifest.Author.Name, manifest.Author.Email)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write README.md: %w", err)
	}

	// Add README.md to documentation
	c.Generator.AddDocumentation(manifest, "README", "README.md")

	// Create offline bundle
	bundle := &OfflineBundle{
		Bundle: Bundle{
			BundlePath: outputPath,
			Manifest:   manifest.BundleManifest,
		},
		EnhancedManifest:   *manifest,
		Format:             c.Format,
		IsIncremental:      false,
		ComplianceMappings: []ComplianceMapping{},
	}

	// Log audit event
	if c.AuditTrail != nil {
		auditLog := &trail.AuditLog{
			Operation:     "create_offline_bundle",
			ResourceType:  "offline_bundle",
			ResourceID:    manifest.BundleID,
			Status:        "success",
			Timestamp:     time.Now(),
			UserID:        manifest.Author.Email,
			Username:      manifest.Author.Name,
			IPAddress:     "",
			Details: map[string]interface{}{
				"name":        name,
				"version":     version,
				"bundle_type": string(bundleType),
				"output_path": outputPath,
				"bundle_id":   manifest.BundleID,
				"created_at":  manifest.CreatedAt,
				"is_incremental": false,
			},
		}
		
		if err := c.AuditTrail.LogOperation(nil, auditLog); err != nil {
			fmt.Fprintf(c.Logger, "Warning: Failed to log audit event: %v\n", err)
		}
	}

	// Log creation success
	fmt.Fprintf(c.Logger, "Offline bundle created successfully: %s\n", outputPath)

	return bundle, nil
}

// AddContentToOfflineBundle adds content to an offline bundle
func (c *OfflineBundleCreator) AddContentToOfflineBundle(bundle *OfflineBundle, sourcePath, targetPath string, contentType ContentType, id, version, description string) error {
	// Check if source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// Determine target directory based on content type
	var targetDir string
	switch contentType {
	case TemplateContentType:
		targetDir = filepath.Join(bundle.BundlePath, "templates")
	case ModuleContentType:
		targetDir = filepath.Join(bundle.BundlePath, "modules")
	case ConfigContentType:
		targetDir = filepath.Join(bundle.BundlePath, "config")
	case ResourceContentType:
		targetDir = filepath.Join(bundle.BundlePath, "resources")
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Determine full target path
	fullTargetPath := filepath.Join(targetDir, targetPath)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullTargetPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Copy file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(fullTargetPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	// Add content to manifest
	relPath := filepath.Join(filepath.Base(targetDir), targetPath)
	c.Generator.AddContentItemToEnhancedManifest(&bundle.EnhancedManifest, relPath, contentType, id, version, description)

	// Update checksums
	if err := c.Generator.UpdateChecksumsForEnhancedManifest(&bundle.EnhancedManifest, bundle.BundlePath); err != nil {
		return fmt.Errorf("failed to update checksums: %w", err)
	}

	// Sign manifest
	if err := c.Generator.SignEnhancedManifest(&bundle.EnhancedManifest); err != nil {
		return fmt.Errorf("failed to sign manifest: %w", err)
	}

	// Update bundle manifest
	bundle.Manifest = bundle.EnhancedManifest.BundleManifest

	// Write updated manifest
	manifestPath := filepath.Join(bundle.BundlePath, "manifest.json")
	if err := c.Generator.WriteEnhancedManifest(&bundle.EnhancedManifest, manifestPath); err != nil {
		return fmt.Errorf("failed to write updated manifest: %w", err)
	}

	// Log audit event
	if c.AuditTrail != nil {
		auditLog := &trail.AuditLog{
			Operation:     "add_content_to_offline_bundle",
			ResourceType:  "offline_bundle",
			ResourceID:    bundle.Manifest.BundleID,
			Status:        "success",
			Timestamp:     time.Now(),
			UserID:        bundle.Manifest.Author.Email,
			Username:      bundle.Manifest.Author.Name,
			IPAddress:     "",
			Details: map[string]interface{}{
				"source_path": sourcePath,
				"target_path": targetPath,
				"content_type": string(contentType),
				"content_id": id,
				"content_version": version,
			},
		}
		
		if err := c.AuditTrail.LogOperation(nil, auditLog); err != nil {
			fmt.Fprintf(c.Logger, "Warning: Failed to log audit event: %v\n", err)
		}
	}

	return nil
}

// AddComplianceMappingToOfflineBundle adds a compliance mapping to an offline bundle
func (c *OfflineBundleCreator) AddComplianceMappingToOfflineBundle(bundle *OfflineBundle, contentID string, owaspCategories, isoControls []string) error {
	// Add compliance mapping to manifest
	if err := c.Generator.AddComplianceMapping(&bundle.EnhancedManifest, contentID, owaspCategories, isoControls); err != nil {
		return fmt.Errorf("failed to add compliance mapping: %w", err)
	}

	// Create compliance mapping object
	mapping := ComplianceMapping{
		ContentID:          contentID,
		OwaspLLMCategories: owaspCategories,
		ISOIECControls:     isoControls,
		Description:        fmt.Sprintf("Mapped by %s", bundle.Manifest.Author.Name),
	}

	// Add to bundle
	bundle.ComplianceMappings = append(bundle.ComplianceMappings, mapping)

	// Sign manifest
	if err := c.Generator.SignEnhancedManifest(&bundle.EnhancedManifest); err != nil {
		return fmt.Errorf("failed to sign manifest: %w", err)
	}

	// Update bundle manifest
	bundle.Manifest = bundle.EnhancedManifest.BundleManifest

	// Write updated manifest
	manifestPath := filepath.Join(bundle.BundlePath, "manifest.json")
	if err := c.Generator.WriteEnhancedManifest(&bundle.EnhancedManifest, manifestPath); err != nil {
		return fmt.Errorf("failed to write updated manifest: %w", err)
	}

	// Write compliance mappings to file
	compliancePath := filepath.Join(bundle.BundlePath, "compliance", "mappings.json")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(compliancePath), 0755); err != nil {
		return fmt.Errorf("failed to create compliance directory: %w", err)
	}
	
	// Write mappings
	mappingsData, err := json.MarshalIndent(bundle.ComplianceMappings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal compliance mappings: %w", err)
	}
	
	if err := os.WriteFile(compliancePath, mappingsData, 0644); err != nil {
		return fmt.Errorf("failed to write compliance mappings: %w", err)
	}

	// Log audit event
	if c.AuditTrail != nil {
		auditLog := &trail.AuditLog{
			Operation:     "add_compliance_mapping_to_offline_bundle",
			ResourceType:  "offline_bundle",
			ResourceID:    bundle.Manifest.BundleID,
			Status:        "success",
			Timestamp:     time.Now(),
			UserID:        bundle.Manifest.Author.Email,
			Username:      bundle.Manifest.Author.Name,
			IPAddress:     "",
			Details: map[string]interface{}{
				"content_id": contentID,
				"owasp_categories": owaspCategories,
				"iso_controls": isoControls,
			},
		}
		
		if err := c.AuditTrail.LogOperation(nil, auditLog); err != nil {
			fmt.Fprintf(c.Logger, "Warning: Failed to log audit event: %v\n", err)
		}
	}

	return nil
}

// AddDocumentationToOfflineBundle adds documentation to an offline bundle
func (c *OfflineBundleCreator) AddDocumentationToOfflineBundle(bundle *OfflineBundle, docType, sourcePath string) error {
	// Check if source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// Determine target path
	targetDir := filepath.Join(bundle.BundlePath, "documentation")
	
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create documentation directory: %w", err)
	}

	// Determine target filename
	targetFilename := filepath.Base(sourcePath)
	targetPath := filepath.Join(targetDir, targetFilename)

	// Copy file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(targetPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write documentation file: %w", err)
	}

	// Add to manifest
	relPath := filepath.Join("documentation", targetFilename)
	c.Generator.AddDocumentation(&bundle.EnhancedManifest, docType, relPath)

	// Sign manifest
	if err := c.Generator.SignEnhancedManifest(&bundle.EnhancedManifest); err != nil {
		return fmt.Errorf("failed to sign manifest: %w", err)
	}

	// Update bundle manifest
	bundle.Manifest = bundle.EnhancedManifest.BundleManifest

	// Write updated manifest
	manifestPath := filepath.Join(bundle.BundlePath, "manifest.json")
	if err := c.Generator.WriteEnhancedManifest(&bundle.EnhancedManifest, manifestPath); err != nil {
		return fmt.Errorf("failed to write updated manifest: %w", err)
	}

	// Log audit event
	if c.AuditTrail != nil {
		auditLog := &trail.AuditLog{
			Operation:     "add_documentation_to_offline_bundle",
			ResourceType:  "offline_bundle",
			ResourceID:    bundle.Manifest.BundleID,
			Status:        "success",
			Timestamp:     time.Now(),
			UserID:        bundle.Manifest.Author.Email,
			Username:      bundle.Manifest.Author.Name,
			IPAddress:     "",
			Details: map[string]interface{}{
				"doc_type": docType,
				"source_path": sourcePath,
				"target_path": relPath,
			},
		}
		
		if err := c.AuditTrail.LogOperation(nil, auditLog); err != nil {
			fmt.Fprintf(c.Logger, "Warning: Failed to log audit event: %v\n", err)
		}
	}

	return nil
}

// CreateIncrementalBundle creates an incremental bundle based on an existing bundle
func (c *OfflineBundleCreator) CreateIncrementalBundle(baseBundle *OfflineBundle, newVersion string, changes []string, outputPath string) (*OfflineBundle, error) {
	// Log creation start
	fmt.Fprintf(c.Logger, "Creating incremental bundle: %s (base: %s, new: %s)\n", 
		baseBundle.Manifest.Name, baseBundle.EnhancedManifest.Version, newVersion)

	// Generate incremental manifest
	manifest := c.Generator.GenerateIncrementalManifest(&baseBundle.EnhancedManifest, newVersion, changes)

	// Create output directory
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create required directories
	for _, dir := range c.Format.RequiredDirectories {
		dirPath := filepath.Join(outputPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create manifest file
	manifestPath := filepath.Join(outputPath, "manifest.json")
	if err := c.Generator.WriteEnhancedManifest(manifest, manifestPath); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Create README.md file with basic information
	readmePath := filepath.Join(outputPath, "README.md")
	readmeContent := fmt.Sprintf(`# %s (Incremental Update)

%s

## Version

%s (based on %s)

## Bundle Type

%s

## Created

%s

## Author

%s (%s)

## Changes

%s

## Contents

This is an incremental update to the base bundle. It contains only the changes since version %s.

## Usage

See the documentation directory for usage instructions.
`, manifest.Name, manifest.Description, newVersion, manifest.BaseVersion, 
		manifest.BundleType, manifest.CreatedAt.Format(time.RFC3339),
		manifest.Author.Name, manifest.Author.Email, 
		formatChanges(changes), manifest.BaseVersion)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write README.md: %w", err)
	}

	// Add README.md to documentation
	c.Generator.AddDocumentation(manifest, "README", "README.md")

	// Create incremental bundle
	bundle := &OfflineBundle{
		Bundle: Bundle{
			BundlePath: outputPath,
			Manifest:   manifest.BundleManifest,
		},
		EnhancedManifest:   *manifest,
		Format:             c.Format,
		IsIncremental:      true,
		ComplianceMappings: []ComplianceMapping{},
	}

	// Log audit event
	if c.AuditTrail != nil {
		auditLog := &trail.AuditLog{
			Operation:     "create_incremental_offline_bundle",
			ResourceType:  "offline_bundle",
			ResourceID:    manifest.BundleID,
			Status:        "success",
			Timestamp:     time.Now(),
			UserID:        manifest.Author.Email,
			Username:      manifest.Author.Name,
			IPAddress:     "",
			Details: map[string]interface{}{
				"base_bundle_id": baseBundle.Manifest.BundleID,
				"base_version":   baseBundle.EnhancedManifest.Version,
				"new_version":    newVersion,
				"output_path":    outputPath,
				"bundle_id":      manifest.BundleID,
				"created_at":     manifest.CreatedAt,
				"is_incremental": true,
				"changes":        changes,
			},
		}
		
		if err := c.AuditTrail.LogOperation(nil, auditLog); err != nil {
			fmt.Fprintf(c.Logger, "Warning: Failed to log audit event: %v\n", err)
		}
	}

	// Log creation success
	fmt.Fprintf(c.Logger, "Incremental bundle created successfully: %s\n", outputPath)

	return bundle, nil
}

// ValidateOfflineBundle validates an offline bundle
func (c *OfflineBundleCreator) ValidateOfflineBundle(bundle *OfflineBundle, level ValidationLevel) (*ValidationResult, error) {
	return c.Validator.ValidateOfflineBundle(bundle, level)
}

// LoadOfflineBundle loads an offline bundle from a directory
func (c *OfflineBundleCreator) LoadOfflineBundle(bundlePath string) (*OfflineBundle, error) {
	// Check if bundle path exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("bundle path does not exist: %s", bundlePath)
	}

	// Load manifest
	manifestPath := filepath.Join(bundlePath, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse manifest
	var enhancedManifest EnhancedBundleManifest
	if err := json.Unmarshal(manifestData, &enhancedManifest); err != nil {
		// Try parsing as regular manifest
		var manifest BundleManifest
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse manifest: %w", err)
		}

		// Convert to enhanced manifest
		enhancedManifest = EnhancedBundleManifest{
			BundleManifest: manifest,
		}
	}

	// Load compliance mappings if they exist
	var complianceMappings []ComplianceMapping
	compliancePath := filepath.Join(bundlePath, "compliance", "mappings.json")
	if _, err := os.Stat(compliancePath); err == nil {
		mappingsData, err := os.ReadFile(compliancePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read compliance mappings: %w", err)
		}

		if err := json.Unmarshal(mappingsData, &complianceMappings); err != nil {
			return nil, fmt.Errorf("failed to parse compliance mappings: %w", err)
		}
	}

	// Create bundle
	bundle := &OfflineBundle{
		Bundle: Bundle{
			BundlePath: bundlePath,
			Manifest:   enhancedManifest.BundleManifest,
		},
		EnhancedManifest:   enhancedManifest,
		Format:             c.Format,
		IsIncremental:      enhancedManifest.IsIncremental,
		ComplianceMappings: complianceMappings,
	}

	// Log audit event
	if c.AuditTrail != nil {
		auditLog := &trail.AuditLog{
			Operation:     "load_offline_bundle",
			ResourceType:  "offline_bundle",
			ResourceID:    enhancedManifest.BundleID,
			Status:        "success",
			Timestamp:     time.Now(),
			UserID:        enhancedManifest.Author.Email,
			Username:      enhancedManifest.Author.Name,
			IPAddress:     "",
			Details: map[string]interface{}{
				"bundle_path": bundlePath,
				"bundle_id":      enhancedManifest.BundleID,
				"version":        enhancedManifest.Version,
				"is_incremental": enhancedManifest.IsIncremental,
			},
		}
		
		if err := c.AuditTrail.LogOperation(nil, auditLog); err != nil {
			fmt.Fprintf(c.Logger, "Warning: Failed to log audit event: %v\n", err)
		}
	}

	return bundle, nil
}

// ExportOfflineBundle exports an offline bundle to a zip file
func (c *OfflineBundleCreator) ExportOfflineBundle(bundle *OfflineBundle, outputPath string) error {
	// Create zip file
	zipPath := outputPath
	if !strings.HasSuffix(zipPath, ".zip") {
		zipPath += ".zip"
	}

	// Create zip file
	return createZipFromDir(bundle.BundlePath, zipPath)
}

// formatChanges formats a list of changes for display
func formatChanges(changes []string) string {
	result := ""
	for _, change := range changes {
		result += fmt.Sprintf("- %s\n", change)
	}
	return result
}

