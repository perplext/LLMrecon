// Package management provides functionality for managing templates in the LLMreconing Tool.
package management

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// Constants for validation
var (
	// ValidSeverityLevels are the valid severity levels
	ValidSeverityLevels = []string{"info", "low", "medium", "high", "critical"}
	// ValidDetectionTypes are the valid detection types
	ValidDetectionTypes = []string{"word", "regex", "condition", "status", "size", "dsl", "time"}
	// ValidConditions are the valid conditions
	ValidConditions = []string{"and", "or", "not"}
)

// SchemaFormat represents the format of a schema
type SchemaFormat string

const (
	// JSONSchema represents a JSON schema
	JSONSchema SchemaFormat = "json"
	// YAMLSchema represents a YAML schema
	YAMLSchema SchemaFormat = "yaml"
)

// SchemaValidator is responsible for validating templates against a schema
type SchemaValidator struct {
	// jsonSchema is the JSON schema for templates
	jsonSchema *gojsonschema.Schema
	// yamlSchema is the YAML schema for templates
	yamlSchema *gojsonschema.Schema
	// customValidators is a map of custom validation functions
	customValidators map[string]func(interface{}) error
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(jsonSchemaPath, yamlSchemaPath string) (*SchemaValidator, error) {
	// Load JSON schema
	jsonSchemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", jsonSchemaPath))
	jsonSchema, err := gojsonschema.NewSchema(jsonSchemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON schema: %w", err)
	}

	// Load YAML schema
	yamlSchemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", yamlSchemaPath))
	yamlSchema, err := gojsonschema.NewSchema(yamlSchemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load YAML schema: %w", err)
	}

	return &SchemaValidator{
		jsonSchema: jsonSchema,
		yamlSchema: yamlSchema,
		customValidators: make(map[string]func(interface{}) error),
	}, nil
}

// AddCustomValidator adds a custom validator function for a specific field
func (v *SchemaValidator) AddCustomValidator(field string, validator func(interface{}) error) {
	v.customValidators[field] = validator
}

// ValidateSchema validates template data against a schema
func (v *SchemaValidator) ValidateSchema(data interface{}) error {
	// Convert data to map for validation
	var templateMap map[string]interface{}
	
	switch d := data.(type) {
	case map[string]interface{}:
		templateMap = d
	case *format.Template:
		var err error
		templateMap, err = templateToMap(d)
		if err != nil {
			return fmt.Errorf("failed to convert template to map: %w", err)
		}
	default:
		return fmt.Errorf("unsupported data type for validation: %T", data)
	}

	// Validate against JSON schema
	jsonLoader := gojsonschema.NewGoLoader(templateMap)
	result, err := v.jsonSchema.Validate(jsonLoader)
	if err != nil {
		return fmt.Errorf("failed to validate data against JSON schema: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, fmt.Sprintf("- %s", err.String()))
		}
		return fmt.Errorf("schema validation failed:\n%s", strings.Join(errMsgs, "\n"))
	}

	// Run custom validators
	for field, validator := range v.customValidators {
		value, ok := getFieldValue(templateMap, field)
		if ok {
			if err := validator(value); err != nil {
				return fmt.Errorf("custom validation failed for field %s: %w", field, err)
			}
		}
	}

	return nil
}

// ValidateTemplate validates a template against the schema
func (v *SchemaValidator) ValidateTemplate(template interface{}) error {
	// Use ValidateSchema for consistency
	return v.ValidateSchema(template)
}

// ValidateTemplateFile validates a template file against the schema
func (v *SchemaValidator) ValidateTemplateFile(filePath string) error {
	// Determine schema format based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	var schemaFormat SchemaFormat
	if ext == ".json" {
		schemaFormat = JSONSchema
	} else if ext == ".yaml" || ext == ".yml" {
		schemaFormat = YAMLSchema
	} else {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Read file
	data, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Validate based on format
	if schemaFormat == JSONSchema {
		return v.ValidateJSON(data)
	} else {
		return v.ValidateYAML(data)
	}
}

// ValidateJSON validates JSON data against the JSON schema
func (v *SchemaValidator) ValidateJSON(data []byte) error {
	// Parse JSON
	var templateMap map[string]interface{}
	if err := json.Unmarshal(data, &templateMap); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate against JSON schema
	jsonLoader := gojsonschema.NewGoLoader(templateMap)
	result, err := v.jsonSchema.Validate(jsonLoader)
	if err != nil {
		return fmt.Errorf("failed to validate JSON against schema: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, fmt.Sprintf("- %s", err.String()))
		}
		return fmt.Errorf("JSON validation failed:\n%s", strings.Join(errMsgs, "\n"))
	}

	// Run custom validators
	for field, validator := range v.customValidators {
		value, ok := getFieldValue(templateMap, field)
		if ok {
			if err := validator(value); err != nil {
				return fmt.Errorf("custom validation failed for field %s: %w", field, err)
			}
		}
	}

	return nil
}

// ValidateYAML validates YAML data against the YAML schema
func (v *SchemaValidator) ValidateYAML(data []byte) error {
	// Parse YAML
	var templateMap map[string]interface{}
	if err := yaml.Unmarshal(data, &templateMap); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate against YAML schema
	yamlLoader := gojsonschema.NewGoLoader(templateMap)
	result, err := v.yamlSchema.Validate(yamlLoader)
	if err != nil {
		return fmt.Errorf("failed to validate YAML against schema: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, fmt.Sprintf("- %s", err.String()))
		}
		return fmt.Errorf("YAML validation failed:\n%s", strings.Join(errMsgs, "\n"))
	}

	// Run custom validators
	for field, validator := range v.customValidators {
		value, ok := getFieldValue(templateMap, field)
		if ok {
			if err := validator(value); err != nil {
				return fmt.Errorf("custom validation failed for field %s: %w", field, err)
			}
		}
	}

	return nil
}

// templateToMap converts a template to a map for validation
func templateToMap(template interface{}) (map[string]interface{}, error) {
	// Marshal to JSON
	data, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template to JSON: %w", err)
	}

	// Unmarshal to map
	var templateMap map[string]interface{}
	if err := json.Unmarshal(data, &templateMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template to map: %w", err)
	}

	return templateMap, nil
}

// getFieldValue gets the value of a field from a map
// The field can be a nested field using dot notation (e.g., "info.name")
func getFieldValue(data map[string]interface{}, field string) (interface{}, bool) {
	parts := strings.Split(field, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, get value
			value, ok := current[part]
			return value, ok
		}

		// Not last part, get next map
		next, ok := current[part]
		if !ok {
			return nil, false
		}
		// Convert to map
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil, false
		}

		current = nextMap
	}

	return nil, false
}

// Standard validators for common fields

// ValidateID validates a template ID
func ValidateID(id interface{}) error {
	idStr, ok := id.(string)
	if !ok {
		return fmt.Errorf("ID must be a string")
	}

	// ID must be alphanumeric with hyphens and underscores
	match, err := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, idStr)
	if err != nil {
		return fmt.Errorf("failed to validate ID: %w", err)
	}

	if !match {
		return fmt.Errorf("ID must be alphanumeric with hyphens and underscores")
	}

	return nil
}

// ValidateVersion validates a version string
func ValidateVersion(version interface{}) error {
	versionStr, ok := version.(string)
	if !ok {
		return fmt.Errorf("version must be a string")
	}

	// Version must be in the format x.y.z
	match, err := regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+$`, versionStr)
	if err != nil {
		return fmt.Errorf("failed to validate version: %w", err)
	}

	if !match {
		return fmt.Errorf("version must be in the format x.y.z")
	}

	return nil
}

// ValidateSeverity validates a severity level
func ValidateSeverity(severity interface{}) error {
	severityStr, ok := severity.(string)
	if !ok {
		return fmt.Errorf("severity must be a string")
	}

	// Severity must be one of the valid levels
	for _, level := range ValidSeverityLevels {
		if severityStr == level {
			return nil
		}
	}

	return fmt.Errorf("severity must be one of: %s", strings.Join(ValidSeverityLevels, ", "))
}

// ValidateDetectionType validates a detection type
func ValidateDetectionType(detectionType interface{}) error {
	typeStr, ok := detectionType.(string)
	if !ok {
		return fmt.Errorf("detection type must be a string")
	}

	// Detection type must be one of the valid types
	for _, validType := range ValidDetectionTypes {
		if typeStr == validType {
			return nil
		}
	}

	return fmt.Errorf("detection type must be one of: %s", strings.Join(ValidDetectionTypes, ", "))
}

// ValidateCondition validates a condition
func ValidateCondition(condition interface{}) error {
	conditionStr, ok := condition.(string)
	if !ok {
		return fmt.Errorf("condition must be a string")
	}

	// Condition must be one of the valid conditions
	for _, validCondition := range ValidConditions {
		if conditionStr == validCondition {
			return nil
		}
	}

	return fmt.Errorf("condition must be one of: %s", strings.Join(ValidConditions, ", "))
}

// SchemaValidatorAdapter adapts SchemaValidator to implement interfaces.SchemaValidator
type SchemaValidatorAdapter struct {
	schemaValidator *SchemaValidator
}

// ValidateSchema validates schema data (interface compatibility)
func (a *SchemaValidatorAdapter) ValidateSchema(data interface{}) error {
	return a.schemaValidator.ValidateSchema(data)
}

// ValidateTemplate validates template (interface compatibility)
func (a *SchemaValidatorAdapter) ValidateTemplate(template interface{}) error {
	return a.schemaValidator.ValidateTemplate(template)
}

// DefaultSchemaValidator creates a default schema validator with standard validators
func DefaultSchemaValidator(jsonSchemaPath, yamlSchemaPath string) (interfaces.TemplateValidator, error) {
	validator, err := NewSchemaValidator(jsonSchemaPath, yamlSchemaPath)
	if err != nil {
		return nil, err
	}

	// Add standard validators
	validator.AddCustomValidator("id", ValidateID)
	validator.AddCustomValidator("info.version", ValidateVersion)
	validator.AddCustomValidator("info.severity", ValidateSeverity)
	validator.AddCustomValidator("test.detection.type", ValidateDetectionType)
	validator.AddCustomValidator("test.detection.condition", ValidateCondition)

	return &TemplateValidatorAdapter{schemaValidator: validator}, nil
}

// DefaultSchemaValidatorForParser creates a schema validator for the parser
func DefaultSchemaValidatorForParser(jsonSchemaPath, yamlSchemaPath string) (interfaces.SchemaValidator, error) {
	validator, err := NewSchemaValidator(jsonSchemaPath, yamlSchemaPath)
	if err != nil {
		return nil, err
	}

	// Add standard validators
	validator.AddCustomValidator("id", ValidateID)
	validator.AddCustomValidator("info.version", ValidateVersion)
	validator.AddCustomValidator("info.severity", ValidateSeverity)
	validator.AddCustomValidator("test.detection.type", ValidateDetectionType)
	validator.AddCustomValidator("test.detection.condition", ValidateCondition)

	return &SchemaValidatorAdapter{schemaValidator: validator}, nil
}

// TemplateValidatorAdapter adapts SchemaValidator to implement interfaces.TemplateValidator
type TemplateValidatorAdapter struct {
	schemaValidator *SchemaValidator
}

// Validate validates a template
func (a *TemplateValidatorAdapter) Validate(template interfaces.Template) error {
	// Convert interfaces.Template to format.Template for validation
	formatTemplate, ok := template.(*format.Template)
	if !ok {
		return fmt.Errorf("template must be of type *format.Template")
	}
	return a.schemaValidator.ValidateTemplate(formatTemplate)
}

// ValidateContent validates template content
func (a *TemplateValidatorAdapter) ValidateContent(content []byte) error {
	// Try to parse as JSON first, then YAML
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		// Try YAML
		if err := yaml.Unmarshal(content, &data); err != nil {
			return fmt.Errorf("failed to parse content as JSON or YAML: %w", err)
		}
		return a.schemaValidator.ValidateYAML(content)
	}
	return a.schemaValidator.ValidateJSON(content)
}

// ValidateSchema validates against a schema
func (a *TemplateValidatorAdapter) ValidateSchema(template interfaces.Template, schema interface{}) error {
	// For now, just use the default schema validation
	return a.Validate(template)
}
