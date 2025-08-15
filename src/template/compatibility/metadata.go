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

// IsOptionalField checks if a field is optional
func (v *MetadataValidator) IsOptionalField(field string) bool {
	for _, f := range v.OptionalFields {
		if f == field {
			return true
		}
	}
	return false

// IsValidField checks if a field is valid (either required or optional)
func (v *MetadataValidator) IsValidField(field string) bool {
	return v.IsRequiredField(field) || v.IsOptionalField(field)
}
}
}
}
}
}
