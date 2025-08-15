// Package parser provides functionality for parsing and validating templates.
package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// TemplateParser is responsible for parsing and validating templates
type TemplateParser struct {
	// schemaValidator is the schema validator for templates
	schemaValidator interfaces.SchemaValidator
	// variablePattern is the regex pattern for template variables
	variablePattern *regexp.Regexp

// NewTemplateParser creates a new template parser
func NewTemplateParser(schemaValidator interfaces.SchemaValidator) (*TemplateParser, error) {
	// Compile variable pattern regex
	variablePattern, err := regexp.Compile(`\{\{\s*([a-zA-Z0-9_\.]+)\s*\}\}`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile variable pattern regex: %w", err)
	}

	return &TemplateParser{
		schemaValidator: schemaValidator,
		variablePattern: variablePattern,
	}, nil

// Parse parses a template
func (p *TemplateParser) Parse(template *format.Template) error {
	// Validate template structure
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}
	if template.Info.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Info.Version == "" {
		return fmt.Errorf("template version is required")
	}
	if template.Test.Prompt == "" {
		return fmt.Errorf("template prompt is required")
	}

	// Validate template against schema
	if p.schemaValidator != nil {
		if err := p.schemaValidator.ValidateTemplate(template); err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
	}

	return nil

// Validate validates a template
func (p *TemplateParser) Validate(template *format.Template) error {
	// Parse template first
	if err := p.Parse(template); err != nil {
		return err
	}

	// Additional validation logic
	errors := template.ValidateStructure()
	if len(errors) > 0 {
		return fmt.Errorf("template validation failed: %s", strings.Join(errors, ", "))
	}

	return nil

// ResolveVariables resolves variables in a template
func (p *TemplateParser) ResolveVariables(template *format.Template, variables map[string]interface{}) error {
	// Convert template to JSON for easier manipulation
	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template to JSON: %w", err)
	}

	// Replace variables in JSON string
	templateStr := string(templateJSON)
	resolvedStr := p.variablePattern.ReplaceAllStringFunc(templateStr, func(match string) string {
		// Extract variable name
		varName := p.variablePattern.FindStringSubmatch(match)[1]

		// Get variable value
		value, ok := getNestedValue(variables, varName)
		if !ok {
			// Keep original variable if not found
			return match
		}

		// Convert value to string
		valueStr, err := convertToString(value)
		if err != nil {
			// Keep original variable if conversion fails
			return match
		}

		return valueStr
	})

	// Unmarshal resolved JSON back to template
	if err := json.Unmarshal([]byte(resolvedStr), template); err != nil {
		return fmt.Errorf("failed to unmarshal resolved template: %w", err)
	}

	return nil

// ExtractVariables extracts variables from a template
func (p *TemplateParser) ExtractVariables(template *format.Template) ([]string, error) {
	// Convert template to JSON for easier manipulation
	templateJSON, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template to JSON: %w", err)
	}

	// Find all variables in JSON string
	templateStr := string(templateJSON)
	matches := p.variablePattern.FindAllStringSubmatch(templateStr, -1)

	// Extract variable names
	var variables []string
	variableMap := make(map[string]bool) // Use map to deduplicate
	for _, match := range matches {
		varName := match[1]
		if !variableMap[varName] {
			variables = append(variables, varName)
			variableMap[varName] = true
		}
	}

	return variables, nil

// getNestedValue gets a nested value from a map using dot notation
func getNestedValue(data map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
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

// convertToString converts a value to a string representation for JSON
func convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		// Escape quotes for JSON
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(v, "\"", "\\\"")), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		// Use fmt.Sprintf for basic types
		return fmt.Sprintf("%v", v), nil
	default:
		// Marshal complex types to JSON
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal value to JSON: %w", err)
		}
		return string(data), nil
	}

// ValidateTemplateInheritance validates template inheritance
func (p *TemplateParser) ValidateTemplateInheritance(template *format.Template, parentTemplate *format.Template) error {
	// Check compatibility
	if template.Compatibility != nil && parentTemplate.Compatibility != nil {
		// Check minimum version
		if template.Compatibility.MinToolVersion != "" && parentTemplate.Compatibility.MinToolVersion != "" {
			// Compare versions
			// This is a simplified version comparison, a real implementation would use semver
			if template.Compatibility.MinToolVersion < parentTemplate.Compatibility.MinToolVersion {
				return fmt.Errorf("template minimum version (%s) is lower than parent minimum version (%s)",
					template.Compatibility.MinToolVersion, parentTemplate.Compatibility.MinToolVersion)
			}
		}
	}

	return nil

// MergeTemplates merges a child template with its parent template
func (p *TemplateParser) MergeTemplates(childTemplate, parentTemplate *format.Template) (*format.Template, error) {
	// Validate inheritance
	if err := p.ValidateTemplateInheritance(childTemplate, parentTemplate); err != nil {
		return nil, fmt.Errorf("invalid template inheritance: %w", err)
	}

	// Create a copy of the parent template
	mergedTemplate := *parentTemplate

	// Override with child template values
	mergedTemplate.ID = childTemplate.ID
	mergedTemplate.Info = childTemplate.Info

	// Merge compatibility
	if childTemplate.Compatibility != nil {
		mergedTemplate.Compatibility = childTemplate.Compatibility
	}

	// Override test definition
	mergedTemplate.Test = childTemplate.Test

