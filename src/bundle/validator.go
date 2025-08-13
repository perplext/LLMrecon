// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/perplext/LLMrecon/src/version"
)

// DefaultBundleValidator is the default implementation of BundleValidator
type DefaultBundleValidator struct {
	// Logger is the logger for validation operations
	Logger io.Writer
}

// NewBundleValidator creates a new bundle validator
func NewBundleValidator(logger io.Writer) BundleValidator {
	if logger == nil {
		logger = os.Stdout
	}
	return &DefaultBundleValidator{
		Logger: logger,
	}
}

// Validate validates a bundle with the specified validation level
func (v *DefaultBundleValidator) Validate(bundle *Bundle, level ValidationLevel) (*ValidationResult, error) {
	// Log validation start
	fmt.Fprintf(v.Logger, "Validating bundle: %s (level: %s)\n", bundle.BundlePath, level)

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

	// Validate manifest
	manifestResult, err := v.ValidateManifest(&bundle.Manifest)
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Message = fmt.Sprintf("Manifest validation failed: %v", err)
		result.Error = err
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	if !manifestResult.Valid {
		result.Valid = false
		result.IsValid = false
		result.Message = fmt.Sprintf("Manifest validation failed: %s", manifestResult.Message)
		result.Errors = append(result.Errors, manifestResult.Errors...)
		result.Warnings = append(result.Warnings, manifestResult.Warnings...)
		return result, fmt.Errorf("manifest validation failed: %s", manifestResult.Message)
	}

	// For basic validation, we only validate the manifest
	if level == BasicValidation {
		return manifestResult, nil
	}

	// Validate checksums for standard and strict validation
	checksumResult, err := v.ValidateChecksums(bundle)
	if err != nil {
		return checksumResult, err
	}

	if !checksumResult.Valid {
		return checksumResult, fmt.Errorf("checksum validation failed: %s", checksumResult.Message)
	}

	// For strict validation, we also validate compatibility
	if level == StrictValidation {
		// Get current versions
		currentVersions, err := getCurrentVersions()
		if err != nil {
			return &ValidationResult{
				Valid:   false,
				Message: "Failed to get current versions",
				Errors:  []string{err.Error()},
			}, fmt.Errorf("failed to get current versions: %w", err)
		}

		compatResult, err := v.ValidateCompatibility(bundle, currentVersions)
		if err != nil {
			return compatResult, err
		}

		if !compatResult.Valid {
			return compatResult, fmt.Errorf("compatibility validation failed: %s", compatResult.Message)
		}
	}

	// Log validation success
	fmt.Fprintf(v.Logger, "Bundle validation successful\n")

	return &ValidationResult{
		Valid:   true,
		Message: "Bundle validation successful",
	}, nil
}

// ValidateManifest validates a bundle manifest
func (v *DefaultBundleValidator) ValidateManifest(manifest *BundleManifest) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     ManifestValidationLevel,
		Message:   "Manifest validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Check required fields
	if manifest.SchemaVersion == "" {
		result.Errors = append(result.Errors, "Schema version is required")
	}

	if manifest.BundleID == "" {
		result.Errors = append(result.Errors, "Bundle ID is required")
	}

	if manifest.Name == "" {
		result.Errors = append(result.Errors, "Name is required")
	}

	if manifest.BundleType == "" {
		result.Errors = append(result.Errors, "Bundle type is required")
	}

	// Validate bundle type
	validBundleTypes := map[BundleType]bool{
		TemplateBundleType: true,
		ModuleBundleType:   true,
		MixedBundleType:    true,
	}

	if !validBundleTypes[manifest.BundleType] {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid bundle type: %s", manifest.BundleType))
	}

	// Check content items
	if len(manifest.Content) == 0 {
		result.Warnings = append(result.Warnings, "Bundle has no content items")
	}

	// Validate content items
	for i, item := range manifest.Content {
		if item.Path == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Content item %d has no path", i))
		}

		if item.Type == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Content item %s has no type", item.Path))
		}

		if item.Checksum == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Content item %s has no checksum", item.Path))
		}

		// Validate content type
		validContentTypes := []ContentType{TemplateContentType, ModuleContentType, ConfigContentType, ResourceContentType}
		validContentType := false
		for _, t := range validContentTypes {
			if item.Type == t {
				validContentType = true
				break
			}
		}
		if !validContentType {
			result.Errors = append(result.Errors, fmt.Sprintf("Content item %s has invalid type: %s", item.Path, item.Type))
		}

		// Check bundle type and content type consistency
		if manifest.BundleType == TemplateBundleType && item.Type == ModuleContentType {
			result.Errors = append(result.Errors, fmt.Sprintf("Content item %s (module) is not allowed in a template bundle", item.Path))
		}
		if manifest.BundleType == ModuleBundleType && item.Type == TemplateContentType {
			result.Errors = append(result.Errors, fmt.Sprintf("Content item %s (template) is not allowed in a module bundle", item.Path))
		}
	}

	// Check compatibility
	if manifest.Compatibility.MinVersion == "" {
		result.Warnings = append(result.Warnings, "Compatibility minimum version is recommended")
	}

	// Update result status
	result.Valid = len(result.Errors) == 0
	result.IsValid = result.Valid
	
	if !result.Valid {
		result.Message = "Manifest validation failed"
		if len(result.Errors) > 0 {
			result.Error = fmt.Errorf("%s", result.Errors[0])
		}
		return result, fmt.Errorf("manifest validation failed: %d errors", len(result.Errors))
	}

	// Return success with warnings if any
	if len(result.Warnings) > 0 {
		result.Message = fmt.Sprintf("Manifest validation successful with %d warnings", len(result.Warnings))
	}

	return result, nil
}

// ValidateSignature validates a bundle signature
func (v *DefaultBundleValidator) ValidateSignature(bundle *Bundle, publicKey ed25519.PublicKey) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     SignatureValidationLevel,
		Message:   "Signature validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Check if signature exists
	if bundle.Manifest.Signature == "" {
		result.Valid = false
		result.IsValid = false
		result.Message = "Bundle signature is missing"
		result.Errors = append(result.Errors, "Signature is required for validation")
		result.Error = fmt.Errorf("bundle signature is missing")
		return result, result.Error
	}

	// Extract signature
	signatureBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(bundle.Manifest.Signature, "base64:"))
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Message = "Failed to decode signature"
		result.Errors = append(result.Errors, err.Error())
		result.Error = fmt.Errorf("failed to decode signature: %w", err)
		return result, result.Error
	}

	// Create a copy of the manifest without the signature for verification
	manifestCopy := bundle.Manifest
	manifestCopy.Signature = ""

	// Marshal the manifest to JSON
	manifestJSON, err := json.Marshal(manifestCopy)
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Message = "Failed to marshal manifest for signature verification"
		result.Errors = append(result.Errors, err.Error())
		result.Error = fmt.Errorf("failed to marshal manifest for signature verification: %w", err)
		return result, result.Error
	}

	// Verify the signature
	if !ed25519.Verify(publicKey, manifestJSON, signatureBytes) {
		result.Valid = false
		result.IsValid = false
		result.Message = "Signature verification failed"
		result.Errors = append(result.Errors, "Invalid signature")
		result.Error = fmt.Errorf("signature verification failed")
		return result, result.Error
	}

	// Return success
	return result, nil
}

// ValidateChecksums validates bundle checksums
func (v *DefaultBundleValidator) ValidateChecksums(bundle *Bundle) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     ChecksumValidationLevel,
		Message:   "Checksum validation successful",
		Timestamp: time.Now(),
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
	}

	// Check if bundle path exists
	if _, err := os.Stat(bundle.BundlePath); os.IsNotExist(err) {
		result.Valid = false
		result.IsValid = false
		result.Message = "Bundle path does not exist"
		result.Errors = append(result.Errors, err.Error())
		result.Error = fmt.Errorf("bundle path does not exist: %w", err)
		return result, result.Error
	}

	// Validate manifest checksum if provided
	manifestPath := filepath.Join(bundle.BundlePath, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		result.Valid = false
		result.IsValid = false
		result.Message = "Manifest file not found"
		result.Errors = append(result.Errors, err.Error())
		result.Error = fmt.Errorf("manifest file not found: %w", err)
		return result, result.Error
	}

	// Read manifest file
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		result.Valid = false
		result.IsValid = false
		result.Message = "Failed to read manifest file"
		result.Errors = append(result.Errors, err.Error())
		result.Error = fmt.Errorf("failed to read manifest file: %w", err)
		return result, result.Error
	}

	// Calculate manifest checksum
	manifestHash := calculateHash(manifestData)
	if bundle.Manifest.Checksums.Manifest != "" && manifestHash != bundle.Manifest.Checksums.Manifest {
		result.Errors = append(result.Errors, fmt.Sprintf("Manifest checksum mismatch: expected %s, got %s", 
			bundle.Manifest.Checksums.Manifest, manifestHash))
	}

	// Validate content checksums
	for _, item := range bundle.Manifest.Content {
		itemPath := filepath.Join(bundle.BundlePath, item.Path)
		
		// Check if item exists
		if _, err := os.Stat(itemPath); os.IsNotExist(err) {
			result.Errors = append(result.Errors, fmt.Sprintf("Content item not found: %s", item.Path))
			continue
		}

		// Calculate checksum
		var itemHash string
		fileInfo, err := os.Stat(itemPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to get file info for %s: %v", item.Path, err))
			continue
		}

		if fileInfo.IsDir() {
			// Calculate directory hash
			dirHash, err := calculateDirectoryHash(itemPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to calculate directory hash for %s: %v", item.Path, err))
				continue
			}
			itemHash = dirHash
		} else {
			// Calculate file hash
			fileData, err := os.ReadFile(itemPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to read file %s: %v", item.Path, err))
				continue
			}
			itemHash = calculateHash(fileData)
		}

		// Compare checksums
		if itemHash != item.Checksum {
			result.Errors = append(result.Errors, fmt.Sprintf("Checksum mismatch for %s: expected %s, got %s", 
				item.Path, item.Checksum, itemHash))
		}
	}

	// Update result status
	result.Valid = len(result.Errors) == 0
	result.IsValid = result.Valid
	
	if !result.Valid {
		result.Message = "Checksum validation failed"
		if len(result.Errors) > 0 {
			result.Error = fmt.Errorf("%s", result.Errors[0])
		}
		return result, fmt.Errorf("checksum validation failed: %d errors", len(result.Errors))
	}

	// Return success with warnings if any
	if len(result.Warnings) > 0 {
		result.Message = fmt.Sprintf("Checksum validation successful with %d warnings", len(result.Warnings))
	}
	
	return result, nil
}

// ValidateCompatibility validates bundle compatibility with current versions
func (v *DefaultBundleValidator) ValidateCompatibility(bundle *Bundle, currentVersions map[string]*version.SemVersion) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:   true,
		IsValid: true,
		Level:   "compatibility",
		Message: "Compatibility validation successful",
		Errors:  []string{},
		Warnings: []string{},
	}

	// Check minimum version compatibility
	if bundle.Manifest.Compatibility.MinVersion != "" {
		minVersion, err := version.Parse(bundle.Manifest.Compatibility.MinVersion)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid minimum version format: %s", err.Error()))
		} else {
			coreVersion, ok := currentVersions["core"]
			if !ok {
				result.Warnings = append(result.Warnings, "Core version not found in current versions")
			} else {
				if coreVersion.Compare(minVersion) < 0 {
					result.Errors = append(result.Errors, fmt.Sprintf("Bundle requires minimum version %s, but current version is %s", 
						bundle.Manifest.Compatibility.MinVersion, coreVersion.String()))
				}
			}
		}
	}

	// Check maximum version compatibility
	if bundle.Manifest.Compatibility.MaxVersion != "" {
		maxVersion, err := version.Parse(bundle.Manifest.Compatibility.MaxVersion)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid max_version format: %s", err.Error()))
		} else {
			coreVersion, ok := currentVersions["core"]
			if !ok {
				result.Warnings = append(result.Warnings, "Core version not found in current versions")
			} else {
				if coreVersion.Compare(maxVersion) > 0 {
					result.Errors = append(result.Errors, fmt.Sprintf("Bundle supports maximum version %s, but current version is %s", 
						bundle.Manifest.Compatibility.MaxVersion, coreVersion.String()))
				}
			}
		}
	}

	// Check dependencies
	for _, dep := range bundle.Manifest.Compatibility.Dependencies {
		parts := strings.SplitN(dep, ":", 2)
		if len(parts) != 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid dependency format: %s", dep))
			continue
		}

		depName := parts[0]
		depVersion, err := version.Parse(parts[1])
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid dependency version format: %s", err.Error()))
			continue
		}

		currentVersion, ok := currentVersions[depName]
		if !ok {
			result.Errors = append(result.Errors, fmt.Sprintf("Dependency %s not found", depName))
			continue
		}

		if currentVersion.Compare(depVersion) < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Bundle requires %s version %s, but current version is %s", 
				depName, parts[1], currentVersion.String()))
		}
	}



	// Check incompatibilities
	for _, incomp := range bundle.Manifest.Compatibility.Incompatible {
		parts := strings.SplitN(incomp, ":", 2)
		if len(parts) != 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid incompatibility format: %s", incomp))
			continue
		}

		incompName := parts[0]
		incompVersion, err := version.Parse(parts[1])
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid incompatibility version format: %s", err.Error()))
			continue
		}

		// Check if incompatible component exists
		currentVersion, ok := currentVersions[incompName]
		if !ok {
			// Component not installed, no incompatibility
			continue
		}

		// Check version incompatibility
		if currentVersion.Compare(incompVersion) == 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Bundle is incompatible with %s version %s", 
				incompName, parts[1]))
		}
	}

	// Check if there are any errors
	if len(result.Errors) > 0 {
		result.Valid = false
		result.IsValid = false
		result.Message = "Compatibility validation failed"
		return result, fmt.Errorf("compatibility validation failed: %d errors", len(result.Errors))
	}

	// Return success with warnings if any
	if len(result.Warnings) > 0 {
		result.Message = fmt.Sprintf("Compatibility validation successful with %d warnings", len(result.Warnings))
	}

	return result, nil
}

// calculateHash calculates the SHA-256 hash of data
func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash)
}

// calculateDirectoryHash calculates the SHA-256 hash of a directory
func calculateDirectoryHash(dirPath string) (string, error) {
	// Create a hash
	h := sha256.New()

	// Walk the directory tree
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == dirPath {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return err
		}

		// Add file path and mode to hash
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		
		// Add path and file mode to hash
		fmt.Fprintf(h, "%s:%d:", relPath, info.Mode())

		// If it's a regular file, add its contents to the hash
		if !d.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			h.Write(data)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Return the hash
	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// getCurrentVersions gets the current versions of components
func getCurrentVersions() (map[string]*version.SemVersion, error) {
	// In a real implementation, this would query the version system
	// For now, we'll return a mock version
	coreVersion, err := version.ParseVersion("1.0.0")
	if err != nil {
		return nil, err
	}

	return map[string]*version.SemVersion{
		"core": &coreVersion,
	}, nil
}
