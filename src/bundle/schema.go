package bundle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidator provides functionality for validating bundle manifests against a schema
type SchemaValidator struct {
	schemaLoader gojsonschema.JSONLoader
}

// NewSchemaValidator creates a new schema validator with the specified schema path
func NewSchemaValidator(schemaPath string) (*SchemaValidator, error) {
	// Check if the schema file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("schema file not found: %s", schemaPath)
	}

	// Create a schema loader
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)

	return &SchemaValidator{
		schemaLoader: schemaLoader,
	}, nil
}

// NewDefaultSchemaValidator creates a new schema validator with the default schema path
func NewDefaultSchemaValidator() (*SchemaValidator, error) {
	// Get the executable directory
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	schemaPath := filepath.Join(execDir, "..", "schemas", "bundle-manifest-schema.json")

	// Check if the schema file exists at the default location
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Try to find the schema in the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}

		schemaPath = filepath.Join(cwd, "schemas", "bundle-manifest-schema.json")
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("schema file not found at default locations")
		}
	}

	return NewSchemaValidator(schemaPath)
}

// ValidateManifestFile validates a manifest file against the schema
func (v *SchemaValidator) ValidateManifestFile(manifestPath string) (*ValidationResult, error) {
	// Check if the manifest file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file not found: %s", manifestPath)
	}

	// Create a document loader
	documentLoader := gojsonschema.NewReferenceLoader("file://" + manifestPath)

	// Validate the manifest
	result, err := gojsonschema.Validate(v.schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create a validation result
	validationResult := &ValidationResult{
		Level:     BasicValidation,
		IsValid:   result.Valid(),
		Timestamp: getCurrentTime(),
	}

	// If the manifest is not valid, add the errors
	if !result.Valid() {
		for _, desc := range result.Errors() {
			validationResult.Errors = append(validationResult.Errors, desc.String())
		}
		validationResult.Message = "Manifest validation failed"
	} else {
		validationResult.Message = "Manifest validation succeeded"
	}

	return validationResult, nil
}

// ValidateManifestJSON validates a manifest JSON string against the schema
func (v *SchemaValidator) ValidateManifestJSON(manifestJSON string) (*ValidationResult, error) {
	// Create a document loader
	documentLoader := gojsonschema.NewStringLoader(manifestJSON)

	// Validate the manifest
	result, err := gojsonschema.Validate(v.schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create a validation result
	validationResult := &ValidationResult{
		Level:     BasicValidation,
		IsValid:   result.Valid(),
		Timestamp: getCurrentTime(),
	}

	// If the manifest is not valid, add the errors
	if !result.Valid() {
		for _, desc := range result.Errors() {
			validationResult.Errors = append(validationResult.Errors, desc.String())
		}
		validationResult.Message = "Manifest validation failed"
	} else {
		validationResult.Message = "Manifest validation succeeded"
	}

	return validationResult, nil
}

// ValidateManifestStruct validates a manifest struct against the schema
func (v *SchemaValidator) ValidateManifestStruct(manifest *BundleManifest) (*ValidationResult, error) {
	// Convert the manifest to JSON
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Validate the manifest JSON
	return v.ValidateManifestJSON(string(manifestJSON))
}

// LoadSchemaFromFile loads a JSON schema from a file
func LoadSchemaFromFile(schemaPath string) (map[string]interface{}, error) {
	// Read the schema file
	schemaData, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	// Parse the schema JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	return schema, nil
}

// GenerateExampleManifest generates an example manifest based on the schema
func GenerateExampleManifest() *BundleManifest {
	return &BundleManifest{
		SchemaVersion: "1.0.0",
		BundleID:      "example-bundle",
		BundleType:    TemplateBundleType,
		Name:          "Example Bundle",
		Description:   "An example bundle for demonstration purposes",
		Version:       "1.0.0",
		CreatedAt:     getCurrentTime(),
		Author: Author{
			Name:  "John Doe",
			Email: "john@example.com",
		},
		Content: []ContentItem{
			{
				Path:        "templates/example.json",
				Type:        TemplateContentType,
				ID:          "template-001",
				Version:     "1.0.0",
				Description: "An example template",
				Checksum:    "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
		},
		Checksums: Checksums{
			Manifest: "sha256:0987654321fedcba0987654321fedcba0987654321fedcba0987654321fedcba",
			Content: map[string]string{
				"templates/example.json": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
		},
		Compatibility: Compatibility{
			MinVersion: "1.0.0",
		},
	}
}

// SaveExampleManifest saves an example manifest to a file
func SaveExampleManifest(outputPath string) error {
	// Generate an example manifest
	manifest := GenerateExampleManifest()

	// Convert the manifest to JSON
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write the manifest to the file
	if err := ioutil.WriteFile(outputPath, manifestJSON, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// getCurrentTime returns the current time
func getCurrentTime() time.Time {
	return time.Now()
}
