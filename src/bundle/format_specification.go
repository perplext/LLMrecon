// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
)

// OfflineBundleFormat represents the format specification for offline bundles
type OfflineBundleFormat struct {
	// SchemaVersion is the version of the bundle schema
	SchemaVersion string `json:"schema_version"`
	// FormatVersion is the version of the offline bundle format
	FormatVersion string `json:"format_version"`
	// SupportedTypes lists the supported bundle types
	SupportedTypes []BundleType `json:"supported_types"`
	// RequiredDirectories lists the required directories in the bundle
	RequiredDirectories []string `json:"required_directories"`
	// OptionalDirectories lists the optional directories in the bundle
	OptionalDirectories []string `json:"optional_directories"`
	// ManifestSchema defines the schema for the manifest file
	ManifestSchema map[string]interface{} `json:"manifest_schema"`
	// ValidationLevels lists the supported validation levels
	ValidationLevels []ValidationLevel `json:"validation_levels"`
	// SignatureAlgorithm specifies the algorithm used for signatures
	SignatureAlgorithm string `json:"signature_algorithm"`
	// ChecksumAlgorithm specifies the algorithm used for checksums
	ChecksumAlgorithm string `json:"checksum_algorithm"`
	// UpdatedAt is the timestamp when the format was last updated
	UpdatedAt time.Time `json:"updated_at"`

// DefaultOfflineBundleFormat returns the default offline bundle format specification
func DefaultOfflineBundleFormat() *OfflineBundleFormat {
	return &OfflineBundleFormat{
		SchemaVersion: "1.0",
		FormatVersion: "1.0",
		SupportedTypes: []BundleType{
			TemplateBundleType,
			ModuleBundleType,
			MixedBundleType,
		},
		RequiredDirectories: []string{
			"manifest.json",
			"templates",
			"documentation",
			"signatures",
		},
		OptionalDirectories: []string{
			"binary",
			"modules",
			"repository-config",
			"resources",
		},
		ManifestSchema: map[string]interface{}{
			"type": "object",
			"required": []string{
				"schema_version",
				"bundle_id",
				"bundle_type",
				"name",
				"description",
				"version",
				"created_at",
				"author",
				"content",
				"checksums",
				"compatibility",
			},
			"properties": map[string]interface{}{
				"schema_version": map[string]interface{}{
					"type": "string",
				},
				"bundle_id": map[string]interface{}{
					"type": "string",
				},
				"bundle_type": map[string]interface{}{
					"type": "string",
					"enum": []string{
						string(TemplateBundleType),
						string(ModuleBundleType),
						string(MixedBundleType),
					},
				},
				"name": map[string]interface{}{
					"type": "string",
				},
				"description": map[string]interface{}{
					"type": "string",
				},
				"version": map[string]interface{}{
					"type": "string",
				},
				"created_at": map[string]interface{}{
					"type": "string",
					"format": "date-time",
				},
				"author": map[string]interface{}{
					"type": "object",
					"required": []string{"name", "email"},
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"email": map[string]interface{}{
							"type": "string",
						},
						"url": map[string]interface{}{
							"type": "string",
						},
						"key_id": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"content": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"required": []string{"path", "type", "checksum"},
						"properties": map[string]interface{}{
							"path": map[string]interface{}{
								"type": "string",
							},
							"type": map[string]interface{}{
								"type": "string",
								"enum": []string{
									string(TemplateContentType),
									string(ModuleContentType),
									string(ConfigContentType),
									string(ResourceContentType),
								},
							},
							"id": map[string]interface{}{
								"type": "string",
							},
							"version": map[string]interface{}{
								"type": "string",
							},
							"description": map[string]interface{}{
								"type": "string",
							},
							"checksum": map[string]interface{}{
								"type": "string",
							},
							"bundle_id": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
				"checksums": map[string]interface{}{
					"type": "object",
					"required": []string{"manifest", "content"},
					"properties": map[string]interface{}{
						"manifest": map[string]interface{}{
							"type": "string",
						},
						"content": map[string]interface{}{
							"type": "object",
							"additionalProperties": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
				"compatibility": map[string]interface{}{
					"type": "object",
					"required": []string{"min_version"},
					"properties": map[string]interface{}{
						"min_version": map[string]interface{}{
							"type": "string",
						},
						"max_version": map[string]interface{}{
							"type": "string",
						},
						"dependencies": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
						"incompatible": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
				"signature": map[string]interface{}{
					"type": "string",
				},
				"compliance": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"owasp_llm_top10": map[string]interface{}{
							"type": "object",
							"additionalProperties": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
						},
						"iso_iec_42001": map[string]interface{}{
							"type": "object",
							"additionalProperties": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				"changelog": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"required": []string{"version", "date", "changes"},
						"properties": map[string]interface{}{
							"version": map[string]interface{}{
								"type": "string",
							},
							"date": map[string]interface{}{
								"type": "string",
								"format": "date-time",
							},
							"changes": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
		},
		ValidationLevels: []ValidationLevel{
			BasicValidation,
			StandardValidation,
			StrictValidation,
			ManifestValidationLevel,
			ChecksumValidationLevel,
			SignatureValidationLevel,
			CompatibilityValidationLevel,
		},
		SignatureAlgorithm: "Ed25519",
		ChecksumAlgorithm:  "SHA-256",
		UpdatedAt:          time.Now().UTC(),
	}

// EnhancedBundleManifest extends BundleManifest with additional fields for offline bundles
type EnhancedBundleManifest struct {
	BundleManifest
	// Compliance contains compliance mapping information
	Compliance struct {
		// OwaspLLMTop10 maps OWASP LLM Top 10 categories to content items
		OwaspLLMTop10 map[string][]string `json:"owasp_llm_top10,omitempty"`
		// ISOIEC42001 maps ISO/IEC 42001 controls to content items
		ISOIEC42001 map[string][]string `json:"iso_iec_42001,omitempty"`
	} `json:"compliance,omitempty"`
	// Changelog contains version history information
	Changelog []ChangelogEntry `json:"changelog,omitempty"`
	// Documentation contains paths to documentation files
	Documentation map[string]string `json:"documentation,omitempty"`
	// IsIncremental indicates whether this is an incremental bundle
	IsIncremental bool `json:"is_incremental,omitempty"`
	// BaseVersion is the base version for incremental bundles
	BaseVersion string `json:"base_version,omitempty"`

// ChangelogEntry represents an entry in the changelog
type ChangelogEntry struct {
	// Version is the version associated with this changelog entry
	Version string `json:"version"`
	// Date is when the version was released
	Date time.Time `json:"date"`
	// Changes is a list of changes in this version
	Changes []string `json:"changes"`

// ComplianceMapping represents a mapping between content items and compliance frameworks
type ComplianceMapping struct {
	// ContentID is the ID of the content item
	ContentID string `json:"content_id"`
	// OwaspLLMCategories lists the OWASP LLM Top 10 categories this item addresses
	OwaspLLMCategories []string `json:"owasp_llm_categories,omitempty"`
	// ISOIECControls lists the ISO/IEC 42001 controls this item addresses
	ISOIECControls []string `json:"iso_iec_controls,omitempty"`
	// Description provides additional context about the compliance mapping
	Description string `json:"description,omitempty"`

// OfflineBundle extends Bundle with additional functionality for offline bundles
type OfflineBundle struct {
	Bundle
	// EnhancedManifest contains the enhanced manifest information
	EnhancedManifest EnhancedBundleManifest
	// Format is the format specification for this bundle
	Format *OfflineBundleFormat
	// IsIncremental indicates whether this is an incremental bundle
	IsIncremental bool
	// ComplianceMappings contains detailed compliance mapping information
	ComplianceMappings []ComplianceMapping

// CreateOfflineBundle creates a new offline bundle
func CreateOfflineBundle(manifest EnhancedBundleManifest, contentDir, outputPath string) (*OfflineBundle, error) {
	// Create standard bundle
	bundle, err := CreateBundle(manifest.BundleManifest, contentDir, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create base bundle: %w", err)
	}

	// Create offline bundle
	offlineBundle := &OfflineBundle{
		Bundle:          *bundle,
		EnhancedManifest: manifest,
		Format:          DefaultOfflineBundleFormat(),
		IsIncremental:   manifest.IsIncremental,
	}

	return offlineBundle, nil

// OpenOfflineBundle opens an offline bundle from the given path
func OpenOfflineBundle(path string) (*OfflineBundle, error) {
	// Open standard bundle
	bundle, err := OpenBundle(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open base bundle: %w", err)
	}

	// Create offline bundle
	offlineBundle := &OfflineBundle{
		Bundle: *bundle,
		Format: DefaultOfflineBundleFormat(),
	}

	// Read enhanced manifest
	err = offlineBundle.readEnhancedManifest()
	if err != nil {
		return nil, err
	}

	return offlineBundle, nil

// readEnhancedManifest reads the enhanced manifest from the bundle
func (b *OfflineBundle) readEnhancedManifest() error {
	// Read manifest file
	manifestPath := filepath.Join(b.BundlePath, "manifest.json")
	manifestData, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Unmarshal enhanced manifest
	err = json.Unmarshal(manifestData, &b.EnhancedManifest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal enhanced manifest: %w", err)
	}

	// Set incremental flag
	b.IsIncremental = b.EnhancedManifest.IsIncremental

	return nil

// ValidateOfflineBundle validates an offline bundle with enhanced validation
func ValidateOfflineBundle(bundle *OfflineBundle, level ValidationLevel, publicKey ed25519.PublicKey) (*ValidationResult, error) {
	// Perform standard validation
	result, err := ValidateBundle(bundle.Bundle, level, publicKey)
	if err != nil {
		return nil, err
	}

	// If standard validation failed, return the result
	if !result.Valid {
		return result, nil
	}

	// Perform enhanced validation for offline bundles
	enhancedResult := &ValidationResult{
		Valid:      true,
		IsValid:    true,
		Level:      level,
		Message:    "Offline bundle validation successful",
		Errors:     []string{},
		Warnings:   []string{},
		Details:    make(map[string]interface{}),
		Timestamp:  time.Now().UTC(),
	}

	// Validate required directories
	if level == StrictValidation {
		for _, dir := range bundle.Format.RequiredDirectories {
			path := filepath.Join(bundle.BundlePath, dir)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				enhancedResult.Valid = false
				enhancedResult.IsValid = false
				enhancedResult.Message = "Offline bundle validation failed"
				enhancedResult.Errors = append(enhancedResult.Errors, fmt.Sprintf("Required directory or file missing: %s", dir))
			}
		}
	}

	// Validate compliance mappings if present
	if level == StrictValidation && len(bundle.EnhancedManifest.Compliance.OwaspLLMTop10) > 0 {
		// Check that all mapped content IDs exist
		for category, contentIDs := range bundle.EnhancedManifest.Compliance.OwaspLLMTop10 {
			for _, contentID := range contentIDs {
				if bundle.GetContentItem(contentID) == nil {
					enhancedResult.Valid = false
					enhancedResult.IsValid = false
					enhancedResult.Message = "Offline bundle validation failed"
					enhancedResult.Errors = append(enhancedResult.Errors, 
						fmt.Sprintf("Content ID %s referenced in OWASP LLM Top 10 mapping for category %s does not exist", 
							contentID, category))
				}
			}
		}
	}

	// Validate ISO/IEC 42001 mappings if present
	if level == StrictValidation && len(bundle.EnhancedManifest.Compliance.ISOIEC42001) > 0 {
		// Check that all mapped content IDs exist
		for control, contentIDs := range bundle.EnhancedManifest.Compliance.ISOIEC42001 {
			for _, contentID := range contentIDs {
				if bundle.GetContentItem(contentID) == nil {
					enhancedResult.Valid = false
					enhancedResult.IsValid = false
					enhancedResult.Message = "Offline bundle validation failed"
					enhancedResult.Errors = append(enhancedResult.Errors, 
						fmt.Sprintf("Content ID %s referenced in ISO/IEC 42001 mapping for control %s does not exist", 
							contentID, control))
				}
			}
		}
	}

	// Validate incremental bundle if applicable
	if bundle.IsIncremental && bundle.EnhancedManifest.BaseVersion == "" {
		enhancedResult.Valid = false
		enhancedResult.IsValid = false
		enhancedResult.Message = "Offline bundle validation failed"
		enhancedResult.Errors = append(enhancedResult.Errors, "Incremental bundle missing base version")
	}

	return enhancedResult, nil

// ValidateBundle validates a bundle with the specified validation level
func ValidateBundle(bundle Bundle, level ValidationLevel, publicKey ed25519.PublicKey) (*ValidationResult, error) {
	// Create validator
	validator := &StandardBundleValidator{}

	// Validate bundle
	result, err := validator.Validate(&bundle, level)
	if err != nil {
		return nil, err
	}

	// If signature validation is requested and a public key is provided, validate the signature
	if (level == BasicValidation || level == StandardValidation || level == StrictValidation || level == SignatureValidationLevel) && 
		publicKey != nil {
		sigResult, err := validator.ValidateSignature(&bundle, publicKey)
		if err != nil {
			return nil, err
		}
		
		// If signature validation failed, update the result
		if !sigResult.Valid {
			result.Valid = false
			result.IsValid = false
			result.Message = "Bundle validation failed: invalid signature"
			result.Errors = append(result.Errors, sigResult.Errors...)
		}
	}

	return result, nil

// CreateIncrementalBundle creates an incremental bundle based on a base bundle
func CreateIncrementalBundle(baseBundle *OfflineBundle, newManifest EnhancedBundleManifest, 
	contentDir, outputPath string) (*OfflineBundle, error) {
	
	// Set incremental flag and base version
	newManifest.IsIncremental = true
	newManifest.BaseVersion = baseBundle.EnhancedManifest.Version

	// Create the incremental bundle
	bundle, err := CreateOfflineBundle(newManifest, contentDir, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create incremental bundle: %w", err)
	}

	return bundle, nil

// MergeIncrementalBundle merges an incremental bundle into a base bundle
func MergeIncrementalBundle(baseBundle, incrementalBundle *OfflineBundle, outputPath string) (*OfflineBundle, error) {
	// Verify that the incremental bundle is actually incremental
	if !incrementalBundle.IsIncremental {
		return nil, fmt.Errorf("bundle is not incremental")
	}

	// Verify that the incremental bundle's base version matches the base bundle's version
	if incrementalBundle.EnhancedManifest.BaseVersion != baseBundle.EnhancedManifest.Version {
		return nil, fmt.Errorf("incremental bundle base version %s does not match base bundle version %s",
			incrementalBundle.EnhancedManifest.BaseVersion, baseBundle.EnhancedManifest.Version)
	}

	// Create a new manifest for the merged bundle
	mergedManifest := baseBundle.EnhancedManifest

	// Update version to the incremental bundle's version
	mergedManifest.Version = incrementalBundle.EnhancedManifest.Version

	// Update creation timestamp
	mergedManifest.CreatedAt = time.Now().UTC()

	// Merge content items
	contentMap := make(map[string]ContentItem)
	
	// Add base bundle content items
	for _, item := range baseBundle.Manifest.Content {
		contentMap[item.ID] = item
	}
	
	// Add or replace with incremental bundle content items
	for _, item := range incrementalBundle.Manifest.Content {
		contentMap[item.ID] = item
	}
	
	// Convert map back to slice
	mergedManifest.Content = make([]ContentItem, 0, len(contentMap))
	for _, item := range contentMap {
		mergedManifest.Content = append(mergedManifest.Content, item)
	}

	// Merge compliance mappings
	if incrementalBundle.EnhancedManifest.Compliance.OwaspLLMTop10 != nil {
		if mergedManifest.Compliance.OwaspLLMTop10 == nil {
			mergedManifest.Compliance.OwaspLLMTop10 = make(map[string][]string)
		}
		for category, contentIDs := range incrementalBundle.EnhancedManifest.Compliance.OwaspLLMTop10 {
			mergedManifest.Compliance.OwaspLLMTop10[category] = contentIDs
		}
	}
	
	if incrementalBundle.EnhancedManifest.Compliance.ISOIEC42001 != nil {
		if mergedManifest.Compliance.ISOIEC42001 == nil {
			mergedManifest.Compliance.ISOIEC42001 = make(map[string][]string)
		}
		for control, contentIDs := range incrementalBundle.EnhancedManifest.Compliance.ISOIEC42001 {
			mergedManifest.Compliance.ISOIEC42001[control] = contentIDs
		}
	}

	// Merge changelog entries
	mergedManifest.Changelog = append(mergedManifest.Changelog, incrementalBundle.EnhancedManifest.Changelog...)

	// Create temporary directory for merged content
	tempDir, err := os.MkdirTemp("", "merged-bundle-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy base bundle content
	for _, item := range baseBundle.Manifest.Content {
		srcPath := filepath.Join(baseBundle.BundlePath, item.Path)
		dstPath := filepath.Join(tempDir, item.Path)
		
		// Create parent directories
		err = os.MkdirAll(filepath.Dir(dstPath), 0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", item.Path, err)
		}
		
		// Copy file
		err = copyFile(srcPath, dstPath)
		if err != nil {
			return nil, fmt.Errorf("failed to copy %s: %w", item.Path, err)
		}
	}

	// Copy or overwrite with incremental bundle content
	for _, item := range incrementalBundle.Manifest.Content {
		srcPath := filepath.Join(incrementalBundle.BundlePath, item.Path)
		dstPath := filepath.Join(tempDir, item.Path)
		
		// Create parent directories
		err = os.MkdirAll(filepath.Dir(dstPath), 0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", item.Path, err)
		}
		
		// Copy file
		err = copyFile(srcPath, dstPath)
		if err != nil {
			return nil, fmt.Errorf("failed to copy %s: %w", item.Path, err)
		}
	}

	// Create the merged bundle
	mergedBundle, err := CreateOfflineBundle(mergedManifest, tempDir, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create merged bundle: %w", err)
	}

	return mergedBundle, nil

// GetComplianceMappings returns the compliance mappings for a bundle
func (b *OfflineBundle) GetComplianceMappings() []ComplianceMapping {
	mappings := make([]ComplianceMapping, 0)
	
	// Process OWASP LLM Top 10 mappings
	for category, contentIDs := range b.EnhancedManifest.Compliance.OwaspLLMTop10 {
		for _, contentID := range contentIDs {
			// Check if mapping already exists for this content ID
			var mapping *ComplianceMapping
			for i := range mappings {
				if mappings[i].ContentID == contentID {
					mapping = &mappings[i]
					break
				}
			}
			
			// Create new mapping if it doesn't exist
			if mapping == nil {
				mappings = append(mappings, ComplianceMapping{
					ContentID: contentID,
				})
				mapping = &mappings[len(mappings)-1]
			}
			
			// Add OWASP category
			mapping.OwaspLLMCategories = append(mapping.OwaspLLMCategories, category)
		}
	}
	
	// Process ISO/IEC 42001 mappings
	for control, contentIDs := range b.EnhancedManifest.Compliance.ISOIEC42001 {
		for _, contentID := range contentIDs {
			// Check if mapping already exists for this content ID
			var mapping *ComplianceMapping
			for i := range mappings {
				if mappings[i].ContentID == contentID {
					mapping = &mappings[i]
					break
				}
			}
			
			// Create new mapping if it doesn't exist
			if mapping == nil {
				mappings = append(mappings, ComplianceMapping{
					ContentID: contentID,
				})
				mapping = &mappings[len(mappings)-1]
			}
			
			// Add ISO/IEC control
			mapping.ISOIECControls = append(mapping.ISOIECControls, control)
		}
	}
	
	return mappings

// AddComplianceMapping adds a compliance mapping to the bundle
func (b *OfflineBundle) AddComplianceMapping(mapping ComplianceMapping) error {
	// Verify that the content ID exists
	if b.GetContentItem(mapping.ContentID) == nil {
		return fmt.Errorf("content ID %s does not exist", mapping.ContentID)
	}
	
	// Add OWASP LLM Top 10 mappings
	for _, category := range mapping.OwaspLLMCategories {
		if b.EnhancedManifest.Compliance.OwaspLLMTop10 == nil {
			b.EnhancedManifest.Compliance.OwaspLLMTop10 = make(map[string][]string)
		}
		
		// Check if content ID is already mapped to this category
		exists := false
		for _, id := range b.EnhancedManifest.Compliance.OwaspLLMTop10[category] {
			if id == mapping.ContentID {
				exists = true
				break
			}
		}
		
		// Add mapping if it doesn't exist
		if !exists {
			b.EnhancedManifest.Compliance.OwaspLLMTop10[category] = append(
				b.EnhancedManifest.Compliance.OwaspLLMTop10[category], mapping.ContentID)
		}
	}
	
	// Add ISO/IEC 42001 mappings
	for _, control := range mapping.ISOIECControls {
		if b.EnhancedManifest.Compliance.ISOIEC42001 == nil {
			b.EnhancedManifest.Compliance.ISOIEC42001 = make(map[string][]string)
		}
		
		// Check if content ID is already mapped to this control
		exists := false
		for _, id := range b.EnhancedManifest.Compliance.ISOIEC42001[control] {
			if id == mapping.ContentID {
				exists = true
				break
			}
		}
		
		// Add mapping if it doesn't exist
		if !exists {
			b.EnhancedManifest.Compliance.ISOIEC42001[control] = append(
				b.EnhancedManifest.Compliance.ISOIEC42001[control], mapping.ContentID)
		}
	}
	
	return nil

// AddChangelogEntry adds a changelog entry to the bundle
func (b *OfflineBundle) AddChangelogEntry(version string, changes []string) {
	entry := ChangelogEntry{
		Version: version,
		Date:    time.Now().UTC(),
		Changes: changes,
	}
	
	b.EnhancedManifest.Changelog = append(b.EnhancedManifest.Changelog, entry)

// GetDocumentationPath returns the path to a documentation file
func (b *OfflineBundle) GetDocumentationPath(docType string) (string, error) {
	if path, ok := b.EnhancedManifest.Documentation[docType]; ok {
		return filepath.Join(b.BundlePath, path), nil
	}
	
	return "", fmt.Errorf("documentation type %s not found", docType)

// AddDocumentation adds a documentation file to the bundle
func (b *OfflineBundle) AddDocumentation(docType, path string) {
	if b.EnhancedManifest.Documentation == nil {
		b.EnhancedManifest.Documentation = make(map[string]string)
	}
	
