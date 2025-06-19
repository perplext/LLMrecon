// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"crypto/ed25519"
	"time"

	"github.com/perplext/LLMrecon/src/version"
)

// BundleType represents the type of bundle
type BundleType string

const (
	// TemplateBundleType represents a bundle containing templates
	TemplateBundleType BundleType = "templates"
	// ModuleBundleType represents a bundle containing modules
	ModuleBundleType BundleType = "modules"
	// MixedBundleType represents a bundle containing both templates and modules
	MixedBundleType BundleType = "mixed"
)

// ContentType represents the type of content in a bundle
type ContentType string

const (
	// TemplateContentType represents template content
	TemplateContentType ContentType = "template"
	// ModuleContentType represents module content
	ModuleContentType ContentType = "module"
	// ConfigContentType represents configuration content
	ConfigContentType ContentType = "config"
	// ResourceContentType represents resource content (e.g., images, data files)
	ResourceContentType ContentType = "resource"
)

// ValidationLevel represents the level of validation to perform
type ValidationLevel string

const (
	// BasicValidation represents basic validation (manifest integrity, signature)
	BasicValidation ValidationLevel = "basic"
	// StandardValidation represents standard validation (basic + content integrity)
	StandardValidation ValidationLevel = "standard"
	// StrictValidation represents strict validation (standard + compatibility)
	StrictValidation ValidationLevel = "strict"
	// ManifestValidationLevel represents manifest validation
	ManifestValidationLevel ValidationLevel = "manifest"
	// ChecksumValidationLevel represents checksum validation
	ChecksumValidationLevel ValidationLevel = "checksum"
	// SignatureValidationLevel represents signature validation
	SignatureValidationLevel ValidationLevel = "signature"
	// CompatibilityValidationLevel represents compatibility validation
	CompatibilityValidationLevel ValidationLevel = "compatibility"
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	// Valid indicates whether the validation was successful
	Valid bool
	// IsValid is an alias for Valid for compatibility with reporting
	IsValid bool
	// Level is the validation level that was performed
	Level ValidationLevel
	// Message contains a human-readable message about the validation
	Message string
	// Errors contains any errors encountered during validation
	Errors []string
	// Error is the primary error that occurred during validation
	Error error
	// Warnings contains any warnings encountered during validation
	Warnings []string
	// Details contains additional details about the validation
	Details map[string]interface{}
	// Timestamp is when the validation was performed
	Timestamp time.Time
}

// Author represents the author of a bundle
type Author struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	URL     string `json:"url,omitempty"`
	KeyID   string `json:"key_id,omitempty"`
}

// ContentItem represents an item in a bundle
type ContentItem struct {
	Path        string                 `json:"path"`
	Type        ContentType            `json:"type"`
	ID          string                 `json:"id,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Description string                 `json:"description,omitempty"`
	Checksum    string                 `json:"checksum"`
	// BundleID is the ID of the bundle this item belongs to
	BundleID    string                 `json:"bundle_id,omitempty"`
	// Metadata stores additional metadata for the item
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	// Size is the file size in bytes
	Size        int64                  `json:"size,omitempty"`
}

// Checksums contains checksums for bundle components
type Checksums struct {
	Manifest  string            `json:"manifest"`
	Content   map[string]string `json:"content"`
}

// BundleChecksums is an alias for Checksums for compatibility
type BundleChecksums struct {
	Algorithm string            `json:"algorithm"`
	Manifest  string            `json:"manifest,omitempty"`
	Content   map[string]string `json:"content"`
}

// Compatibility represents compatibility information for a bundle
type Compatibility struct {
	MinVersion    string   `json:"min_version"`
	MaxVersion    string   `json:"max_version,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
	Incompatible  []string `json:"incompatible,omitempty"`
}

// BundleManifest represents the manifest of a bundle
type BundleManifest struct {
	SchemaVersion   string                 `json:"schema_version"`
	ManifestVersion string                 `json:"manifest_version,omitempty"`
	BundleID        string                 `json:"bundle_id"`
	BundleType      BundleType             `json:"bundle_type"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Version         string                 `json:"version"`
	BundleVersion   string                 `json:"bundle_version,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	Author          Author                 `json:"author"`
	Content         []ContentItem          `json:"content"`
	Checksums       Checksums              `json:"checksums"`
	Compatibility   Compatibility          `json:"compatibility"`
	Signature       string                 `json:"signature,omitempty"`
	Dependencies    map[string][]string    `json:"dependencies,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Bundle represents a bundle for import/export
type Bundle struct {
	Manifest    BundleManifest
	BundlePath  string
	IsVerified  bool
}

// BundleValidator defines the interface for bundle validation
type BundleValidator interface {
	// Validate validates a bundle with the specified validation level
	Validate(bundle *Bundle, level ValidationLevel) (*ValidationResult, error)
	// ValidateManifest validates a bundle manifest
	ValidateManifest(manifest *BundleManifest) (*ValidationResult, error)
	// ValidateSignature validates a bundle signature
	ValidateSignature(bundle *Bundle, publicKey ed25519.PublicKey) (*ValidationResult, error)
	// ValidateChecksums validates bundle checksums
	ValidateChecksums(bundle *Bundle) (*ValidationResult, error)
	// ValidateCompatibility validates bundle compatibility with current versions
	ValidateCompatibility(bundle *Bundle, currentVersions map[string]*version.SemVersion) (*ValidationResult, error)
}
