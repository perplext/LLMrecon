package compatibility

import (
	"fmt"
	"strings"
)

// MetadataValidator validates template metadata
type MetadataValidator struct {
	// RequiredFields defines required metadata fields
	RequiredFields []string
	// OptionalFields defines optional metadata fields
	OptionalFields []string
}

// NewMetadataValidator creates a new metadata validator
func NewMetadataValidator() *MetadataValidator {
	return &MetadataValidator{
		RequiredFields: []string{"name", "version", "description"},
		OptionalFields: []string{"author", "tags", "category"},
	}
}

// Validate validates metadata
func (v *MetadataValidator) Validate(metadata map[string]interface{}) error {
	// Check required fields
	for _, field := range v.RequiredFields {
		if _, ok := metadata[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// ValidateField validates a single metadata field
func (v *MetadataValidator) ValidateField(field string, value interface{}) error {
	switch field {
	case "name":
		if str, ok := value.(string); !ok || str == "" {
			return fmt.Errorf("name must be a non-empty string")
		}
	case "version":
		if str, ok := value.(string); !ok || str == "" {
			return fmt.Errorf("version must be a non-empty string")
		}
	case "description":
		if str, ok := value.(string); !ok || str == "" {
			return fmt.Errorf("description must be a non-empty string")
		}
	case "author":
		if str, ok := value.(string); ok && str == "" {
			return fmt.Errorf("author must be a non-empty string if provided")
		}
	case "tags":
		if _, ok := value.([]string); !ok {
			if str, ok := value.(string); ok {
				// Allow comma-separated tags
				if str == "" {
					return fmt.Errorf("tags must be non-empty if provided")
				}
			} else {
				return fmt.Errorf("tags must be a string or array of strings")
			}
		}
	case "category":
		if str, ok := value.(string); ok && str == "" {
			return fmt.Errorf("category must be a non-empty string if provided")
		}
	}
	return nil
}

// NormalizeMetadata normalizes metadata fields
func (v *MetadataValidator) NormalizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	for key, value := range metadata {
		// Normalize field names to lowercase
		normalizedKey := strings.ToLower(key)

		// Handle tags field
		if normalizedKey == "tags" {
			if str, ok := value.(string); ok {
				// Convert comma-separated string to array
				tags := strings.Split(str, ",")
				for i, tag := range tags {
					tags[i] = strings.TrimSpace(tag)
				}
				normalized[normalizedKey] = tags
			} else {
				normalized[normalizedKey] = value
			}
		} else {
			normalized[normalizedKey] = value
		}
	}

	return normalized
}

// GetRequiredFields returns the list of required fields
func (v *MetadataValidator) GetRequiredFields() []string {
	return v.RequiredFields
}

// GetOptionalFields returns the list of optional fields
func (v *MetadataValidator) GetOptionalFields() []string {
	return v.OptionalFields
}

// IsRequiredField checks if a field is required
func (v *MetadataValidator) IsRequiredField(field string) bool {
	for _, f := range v.RequiredFields {
		if f == field {
			return true
		}
	}
	return false
}

// IsOptionalField checks if a field is optional
func (v *MetadataValidator) IsOptionalField(field string) bool {
	for _, f := range v.OptionalFields {
		if f == field {
			return true
		}
	}
	return false
}

// IsValidField checks if a field is valid (either required or optional)
func (v *MetadataValidator) IsValidField(field string) bool {
	return v.IsRequiredField(field) || v.IsOptionalField(field)
}

// CompatibilityMetadata represents compatibility metadata for a template
type CompatibilityMetadata struct {
	// MinToolVersion is the minimum tool version required
	MinToolVersion string `json:"min_tool_version"`
	// MaxToolVersion is the maximum tool version supported
	MaxToolVersion string `json:"max_tool_version"`
	// Providers is the list of compatible providers
	Providers []string `json:"providers"`
	// RequiredFeatures is the list of required features
	RequiredFeatures []string `json:"required_features"`
}

// IsCompatibleWithToolVersion checks if the metadata is compatible with a tool version
func (c *CompatibilityMetadata) IsCompatibleWithToolVersion(version string) (bool, error) {
	// Simple string comparison - in production would use semantic versioning
	if c.MinToolVersion != "" && version < c.MinToolVersion {
		return false, nil
	}
	if c.MaxToolVersion != "" && version > c.MaxToolVersion {
		return false, nil
	}
	return true, nil
}

// HasRequiredFeatures checks if the given features contain all required features
func (c *CompatibilityMetadata) HasRequiredFeatures(availableFeatures []string) bool {
	for _, required := range c.RequiredFeatures {
		found := false
		for _, available := range availableFeatures {
			if required == available {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// IsCompatibleWithProvider checks if the metadata is compatible with a provider
func (c *CompatibilityMetadata) IsCompatibleWithProvider(providerID string) bool {
	for _, p := range c.Providers {
		if p == providerID {
			return true
		}
	}
	return false
}

// IsCompatibleWithModel checks if the metadata is compatible with a model
func (c *CompatibilityMetadata) IsCompatibleWithModel(modelID string) bool {
	// For now, we don't have specific model compatibility checks
	// This could be extended to check model-specific requirements
	return true
}
