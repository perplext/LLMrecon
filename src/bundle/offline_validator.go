// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"crypto/ed25519"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/version"
)

// OfflineBundleValidator is a validator for offline bundles
type OfflineBundleValidator struct {
	// BaseValidator is the underlying bundle validator
	BaseValidator BundleValidator
	// Logger is the logger for validation operations
	Logger io.Writer

// NewOfflineBundleValidator creates a new offline bundle validator
func NewOfflineBundleValidator(logger io.Writer) *OfflineBundleValidator {
	if logger == nil {
		logger = os.Stdout
	}
	return &OfflineBundleValidator{
		BaseValidator: NewBundleValidator(logger),
		Logger:        logger,
	}

// Validate validates an offline bundle
func (v *OfflineBundleValidator) Validate(bundle *OfflineBundle, level ValidationLevel) (*ValidationResult, error) {
	return v.ValidateOfflineBundle(bundle, level)

// ValidateOfflineBundle validates an offline bundle
func (v *OfflineBundleValidator) ValidateOfflineBundle(bundle *OfflineBundle, level ValidationLevel) (*ValidationResult, error) {
	// Log validation start
	fmt.Fprintf(v.Logger, "Validating offline bundle: %s (level: %s)\n", bundle.BundlePath, level)

	// Create a validation result
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     level,
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Validate base bundle
	baseResult, err := v.BaseValidator.Validate(&bundle.Bundle, level)
	if err != nil {
		return baseResult, err
	}

	if !baseResult.Valid {
		return baseResult, fmt.Errorf("base bundle validation failed: %s", baseResult.Message)
	}

	// Validate enhanced manifest
	manifestResult, err := v.ValidateEnhancedManifest(&bundle.EnhancedManifest, level)
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Message = fmt.Sprintf("Enhanced manifest validation failed: %v", err)
		result.Error = err
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if !manifestResult.Valid {
		result.Valid = false
		result.IsValid = false
		result.Message = fmt.Sprintf("Enhanced manifest validation failed: %s", manifestResult.Message)
		result.Errors = append(result.Errors, manifestResult.Errors...)
		result.Warnings = append(result.Warnings, manifestResult.Warnings...)
		return result, fmt.Errorf("enhanced manifest validation failed: %s", manifestResult.Message)
	}

	// Validate directory structure
	if level == StrictValidation {
		dirResult, err := v.ValidateDirectoryStructure(bundle)
		if err != nil {
			result.Valid = false
			result.IsValid = false
			result.Message = fmt.Sprintf("Directory structure validation failed: %v", err)
			result.Error = err
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}

		if !dirResult.Valid {
			result.Valid = false
			result.IsValid = false
			result.Message = fmt.Sprintf("Directory structure validation failed: %s", dirResult.Message)
			result.Errors = append(result.Errors, dirResult.Errors...)
			result.Warnings = append(result.Warnings, dirResult.Warnings...)
			return result, fmt.Errorf("directory structure validation failed: %s", dirResult.Message)
		}
	}

	// Validate compliance mappings
	if level == StrictValidation {
		complianceResult, err := v.ValidateComplianceMappings(bundle)
		if err != nil {
			result.Valid = false
			result.IsValid = false
			result.Message = fmt.Sprintf("Compliance mapping validation failed: %v", err)
			result.Error = err
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}

		if !complianceResult.Valid {
			result.Valid = false
			result.IsValid = false
			result.Message = fmt.Sprintf("Compliance mapping validation failed: %s", complianceResult.Message)
			result.Errors = append(result.Errors, complianceResult.Errors...)
			result.Warnings = append(result.Warnings, complianceResult.Warnings...)
			return result, fmt.Errorf("compliance mapping validation failed: %s", complianceResult.Message)
		}
	}

	// Validate incremental bundle if applicable
	if bundle.IsIncremental {
		incrementalResult, err := v.ValidateIncrementalBundle(bundle)
		if err != nil {
			result.Valid = false
			result.IsValid = false
			result.Message = fmt.Sprintf("Incremental bundle validation failed: %v", err)
			result.Error = err
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}

		if !incrementalResult.Valid {
			result.Valid = false
			result.IsValid = false
			result.Message = fmt.Sprintf("Incremental bundle validation failed: %s", incrementalResult.Message)
			result.Errors = append(result.Errors, incrementalResult.Errors...)
			result.Warnings = append(result.Warnings, incrementalResult.Warnings...)
			return result, fmt.Errorf("incremental bundle validation failed: %s", incrementalResult.Message)
		}
	}

	// Log validation success
	fmt.Fprintf(v.Logger, "Offline bundle validation successful\n")

	return &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     level,
		Message:   "Offline bundle validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}, nil

// ValidateEnhancedManifest validates an enhanced bundle manifest
func (v *OfflineBundleValidator) ValidateEnhancedManifest(manifest *EnhancedBundleManifest, level ValidationLevel) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     level,
		Message:   "Enhanced manifest validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Validate base manifest
	baseResult, err := v.BaseValidator.ValidateManifest(&manifest.BundleManifest)
	if err != nil {
		return baseResult, err
	}

	if !baseResult.Valid {
		return baseResult, fmt.Errorf("base manifest validation failed: %s", baseResult.Message)
	}

	// Validate incremental bundle fields if applicable
	if manifest.IsIncremental {
		if manifest.BaseVersion == "" {
			result.Valid = false
			result.IsValid = false
			result.Errors = append(result.Errors, "Incremental bundle must specify a base version")
		}
	}

	// Validate compliance mappings if present
	if len(manifest.Compliance.OwaspLLMTop10) > 0 {
		for category, contentIDs := range manifest.Compliance.OwaspLLMTop10 {
			if len(contentIDs) == 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("OWASP LLM Top 10 category %s has no mapped content items", category))
			}

			for _, contentID := range contentIDs {
				found := false
				for _, item := range manifest.Content {
					if item.ID == contentID {
						found = true
						break
					}
				}

				if !found {
					result.Errors = append(result.Errors, fmt.Sprintf("Content ID %s referenced in OWASP LLM Top 10 mapping for category %s does not exist", contentID, category))
				}
			}
		}
	}

	// Validate ISO/IEC 42001 mappings if present
	if len(manifest.Compliance.ISOIEC42001) > 0 {
		for control, contentIDs := range manifest.Compliance.ISOIEC42001 {
			if len(contentIDs) == 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("ISO/IEC 42001 control %s has no mapped content items", control))
			}

			for _, contentID := range contentIDs {
				found := false
				for _, item := range manifest.Content {
					if item.ID == contentID {
						found = true
						break
					}
				}

				if !found {
					result.Errors = append(result.Errors, fmt.Sprintf("Content ID %s referenced in ISO/IEC 42001 mapping for control %s does not exist", contentID, control))
				}
			}
		}
	}

	// Validate changelog if present
	if len(manifest.Changelog) > 0 {
		for i, entry := range manifest.Changelog {
			if entry.Version == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Changelog entry %d has no version", i))
			}

			if entry.Date.IsZero() {
				result.Errors = append(result.Errors, fmt.Sprintf("Changelog entry for version %s has no date", entry.Version))
			}

			if len(entry.Changes) == 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Changelog entry for version %s has no changes", entry.Version))
			}
		}
	}

	// Validate documentation if present
	if len(manifest.Documentation) > 0 {
		for docType, path := range manifest.Documentation {
			if path == "" {
				result.Errors = append(result.Errors, fmt.Sprintf("Documentation type %s has an empty path", docType))
			}
		}
	}

	// Check if there are any errors
	if len(result.Errors) > 0 {
		result.Valid = false
		result.IsValid = false
		result.Message = "Enhanced manifest validation failed"
	}

	return result, nil

// ValidateDirectoryStructure validates the directory structure of an offline bundle
func (v *OfflineBundleValidator) ValidateDirectoryStructure(bundle *OfflineBundle) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     StrictValidation,
		Message:   "Directory structure validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Check required directories
	for _, dir := range bundle.Format.RequiredDirectories {
		path := filepath.Join(bundle.BundlePath, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.Valid = false
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Required directory or file missing: %s", dir))
		}
	}

	entries, err := os.ReadDir(bundle.BundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		isRequired := false
		isOptional := false

		// Check if it's a required directory
		for _, dir := range bundle.Format.RequiredDirectories {
			if dir == name || strings.HasPrefix(dir, name+"/") {
				isRequired = true
				break
			}
		}

		// Check if it's an optional directory
		if !isRequired {
			for _, dir := range bundle.Format.OptionalDirectories {
				if dir == name || strings.HasPrefix(dir, name+"/") {
					isOptional = true
					break
				}
			}
		}

		if !isRequired && !isOptional {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Unexpected top-level directory or file: %s", name))
		}
	}

	// Check if there are any errors
	if len(result.Errors) > 0 {
		result.Valid = false
		result.IsValid = false
		result.Message = "Directory structure validation failed"
	}

	return result, nil

// ValidateComplianceMappings validates the compliance mappings of an offline bundle
func (v *OfflineBundleValidator) ValidateComplianceMappings(bundle *OfflineBundle) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     StrictValidation,
		Message:   "Compliance mapping validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Get compliance mappings
	mappings := bundle.GetComplianceMappings()

	// Check for content items without compliance mappings
	for _, item := range bundle.Manifest.Content {
		if item.Type == TemplateContentType {
			found := false
			for _, mapping := range mappings {
				if mapping.ContentID == item.ID {
					found = true
					break
				}
			}

			if !found {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Template content item %s has no compliance mappings", item.ID))
			}
		}
	}

	// Check for OWASP LLM Top 10 categories without mappings
	if len(bundle.EnhancedManifest.Compliance.OwaspLLMTop10) > 0 {
		// Define expected OWASP LLM Top 10 categories
		expectedCategories := []string{
			"LLM01:PromptInjection",
			"LLM02:InsecureOutput",
			"LLM03:TrainingDataPoisoning",
			"LLM04:ModelDenialOfService",
			"LLM05:SupplyChainVulnerabilities",
			"LLM06:SensitiveInformationDisclosure",
			"LLM07:InsecurePluginDesign",
			"LLM08:ExcessiveAgency",
			"LLM09:Overreliance",
			"LLM10:ModelTheft",
		}

		// Check for missing categories
		for _, category := range expectedCategories {
			if _, ok := bundle.EnhancedManifest.Compliance.OwaspLLMTop10[category]; !ok {
				result.Warnings = append(result.Warnings, fmt.Sprintf("OWASP LLM Top 10 category %s has no mappings", category))
			}
		}
	}

	// Check if there are any errors
	if len(result.Errors) > 0 {
		result.Valid = false
		result.IsValid = false
		result.Message = "Compliance mapping validation failed"
	}

	return result, nil

// ValidateIncrementalBundle validates an incremental bundle
func (v *OfflineBundleValidator) ValidateIncrementalBundle(bundle *OfflineBundle) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     StrictValidation,
		Message:   "Incremental bundle validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Check that base version is specified
	if bundle.EnhancedManifest.BaseVersion == "" {
		result.Valid = false
		result.IsValid = false
		result.Errors = append(result.Errors, "Incremental bundle must specify a base version")
	}

	// Check that version is different from base version
	if bundle.EnhancedManifest.Version == bundle.EnhancedManifest.BaseVersion {
		result.Valid = false
		result.IsValid = false
		result.Errors = append(result.Errors, "Incremental bundle version must be different from base version")
	}

	// Check that changelog includes an entry for the current version
	versionFound := false
	for _, entry := range bundle.EnhancedManifest.Changelog {
		if entry.Version == bundle.EnhancedManifest.Version {
			versionFound = true
			break
		}
	}

	if !versionFound {
		result.Warnings = append(result.Warnings, "Incremental bundle does not include a changelog entry for the current version")
	}

	// Check if there are any errors
	if len(result.Errors) > 0 {
		result.Valid = false
		result.IsValid = false
		result.Message = "Incremental bundle validation failed"
	}

	return result, nil

// ValidateSignature validates the signature of an offline bundle
func (v *OfflineBundleValidator) ValidateSignature(bundle *OfflineBundle, publicKey ed25519.PublicKey) (*ValidationResult, error) {
	return v.BaseValidator.ValidateSignature(&bundle.Bundle, publicKey)

// ValidateChecksums validates the checksums of an offline bundle
func (v *OfflineBundleValidator) ValidateChecksums(bundle *OfflineBundle) (*ValidationResult, error) {
	return v.BaseValidator.ValidateChecksums(&bundle.Bundle)

// ValidateCompatibility validates the compatibility of an offline bundle
func (v *OfflineBundleValidator) ValidateCompatibility(bundle *OfflineBundle, currentVersions map[string]version.Version) (*ValidationResult, error) {
	// Convert version.Version map to *version.SemVersion map
	semVersions := make(map[string]*version.SemVersion)
	for key, ver := range currentVersions {
		semVersions[key] = &version.SemVersion{
			Major: ver.Major,
			Minor: ver.Minor,
			Patch: ver.Patch,
		}
	}
	return v.BaseValidator.ValidateCompatibility(&bundle.Bundle, semVersions)

// StandardBundleValidator is a standard implementation of BundleValidator
type StandardBundleValidator struct {
	// Logger is the logger for validation operations
	Logger io.Writer

// NewStandardBundleValidator creates a new standard bundle validator
func NewStandardBundleValidator(logger io.Writer) *StandardBundleValidator {
	if logger == nil {
		logger = os.Stdout
	}
	return &StandardBundleValidator{
		Logger: logger,
	}

// Validate validates a bundle with the specified validation level
func (v *StandardBundleValidator) Validate(bundle *Bundle, level ValidationLevel) (*ValidationResult, error) {
	// Create default validator
	defaultValidator := &DefaultBundleValidator{
		Logger: v.Logger,
	}
	return defaultValidator.Validate(bundle, level)

// ValidateManifest validates a bundle manifest
func (v *StandardBundleValidator) ValidateManifest(manifest *BundleManifest) (*ValidationResult, error) {
	// Create default validator
	defaultValidator := &DefaultBundleValidator{
		Logger: v.Logger,
	}
	return defaultValidator.ValidateManifest(manifest)

// ValidateSignature validates a bundle signature
func (v *StandardBundleValidator) ValidateSignature(bundle *Bundle, publicKey ed25519.PublicKey) (*ValidationResult, error) {
	// Create default validator
	defaultValidator := &DefaultBundleValidator{
		Logger: v.Logger,
	}
	return defaultValidator.ValidateSignature(bundle, publicKey)

// ValidateChecksums validates bundle checksums
func (v *StandardBundleValidator) ValidateChecksums(bundle *Bundle) (*ValidationResult, error) {
	// Create default validator
	defaultValidator := &DefaultBundleValidator{
		Logger: v.Logger,
	}
	return defaultValidator.ValidateChecksums(bundle)

// ValidateCompatibility validates bundle compatibility with current versions
func (v *StandardBundleValidator) ValidateCompatibility(bundle *Bundle, currentVersions map[string]version.Version) (*ValidationResult, error) {
	// Create default validator
	defaultValidator := &DefaultBundleValidator{
		Logger: v.Logger,
	}
	// Convert version.Version map to *version.SemVersion map
	semVersions := make(map[string]*version.SemVersion)
	for key, ver := range currentVersions {
		semVersions[key] = &version.SemVersion{
			Major: ver.Major,
			Minor: ver.Minor,
			Patch: ver.Patch,
		}
	}
