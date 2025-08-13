// Package interfaces provides interfaces for template management components
package interfaces

import (
	"context"

	"github.com/perplext/LLMrecon/src/template/format"
)

// InputValidationRule defines a rule for validating template inputs
type InputValidationRule interface {
	// Validate validates a template against the rule
	Validate(ctx context.Context, template *format.Template) error
	// GetName returns the name of the rule
	GetName() string
	// GetDescription returns the description of the rule
	GetDescription() string
}

// InputValidator validates template inputs before sending to LLM providers
type InputValidator interface {
	// ValidateTemplate validates a template against all rules
	ValidateTemplate(ctx context.Context, template *format.Template) error
	// SanitizePrompt sanitizes a prompt to make it safer for execution
	SanitizePrompt(prompt string) string
	// AddRule adds a validation rule
	AddRule(rule InputValidationRule)
	// RemoveRule removes a validation rule by name
	RemoveRule(name string) bool
	// SetStrictMode sets the strict mode
	SetStrictMode(strict bool)
}
