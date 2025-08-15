package interfaces

// ValidationResult represents the result of validation
type ValidationResult struct {
	// Valid indicates if validation passed
	Valid bool
	// Errors contains validation errors
	Errors []string
	// Warnings contains validation warnings
	Warnings []string
}

// TemplateValidatorExt extends the basic validator interface
type TemplateValidatorExt interface {
	TemplateValidator
	
	// ValidateWithResult validates and returns detailed result
	ValidateWithResult(template Template) ValidationResult
	
	// ValidateContentWithResult validates content and returns detailed result
	ValidateContentWithResult(content []byte) ValidationResult
	
	// GetValidationRules returns the validation rules
	GetValidationRules() []ValidationRule
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	// Name is the rule name
	Name string
	// Description is the rule description
	Description string
	// Severity is the rule severity
	Severity string
	// Validate is the validation function
	Validate func(template Template) error
}

// DefaultValidator provides a default validator implementation
type DefaultValidator struct {
	rules []ValidationRule
}

// NewDefaultValidator creates a new default validator
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{
		rules: []ValidationRule{},
	}
}

// Validate validates a template
func (v *DefaultValidator) Validate(template Template) error {
	for _, rule := range v.rules {
		if err := rule.Validate(template); err != nil {
			return err
		}
	}
	return nil
}

// ValidateContent validates template content
func (v *DefaultValidator) ValidateContent(content []byte) error {
	// Implementation would go here
	return nil
}

// ValidateSchema validates against a schema
func (v *DefaultValidator) ValidateSchema(template Template, schema interface{}) error {
	// Implementation would go here
	return nil
}

// AddRule adds a validation rule
func (v *DefaultValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}
