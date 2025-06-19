// Package validation provides functionality for validating templates and inputs before execution.
package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// ruleAdapter adapts our validation rules to the interfaces.InputValidationRule interface
type ruleAdapter struct {
	rule interfaces.InputValidationRule
}

// InputValidator validates template inputs before sending to LLM providers
type InputValidator struct {
	// rules is the list of validation rules
	rules []interfaces.InputValidationRule
	// strictMode determines if validation errors should fail execution
	strictMode bool
}

// NewInputValidator creates a new input validator with default rules
func NewInputValidator(strictMode bool) *InputValidator {
	validator := &InputValidator{
		strictMode: strictMode,
		rules:      make([]interfaces.InputValidationRule, 0),
	}

	// Add default rules
	validator.AddRule(NewNoJailbreakPatternRule())
	validator.AddRule(NewNoSensitiveDataRule())
	validator.AddRule(NewMaxPromptLengthRule(8000)) // Reasonable default for most LLMs
	validator.AddRule(NewSanitizeHTMLRule())
	validator.AddRule(NewSanitizeScriptRule())
	validator.AddRule(NewNoSQLInjectionRule())
	validator.AddRule(NewNoCommandInjectionRule())

	return validator
}

// AddRule adds a validation rule
func (v *InputValidator) AddRule(rule interfaces.InputValidationRule) {
	v.rules = append(v.rules, rule)
}

// RemoveRule removes a validation rule by name
func (v *InputValidator) RemoveRule(name string) bool {
	for i, rule := range v.rules {
		if rule.GetName() == name {
			v.rules = append(v.rules[:i], v.rules[i+1:]...)
			return true
		}
	}
	return false
}

// SetStrictMode sets the strict mode
func (v *InputValidator) SetStrictMode(strict bool) {
	v.strictMode = strict
}

// ValidateTemplate validates a template against all rules
func (v *InputValidator) ValidateTemplate(ctx context.Context, template *format.Template) error {
	if template == nil {
		return fmt.Errorf("template is nil")
	}

	// Collect all validation errors
	var validationErrors []string

	for _, rule := range v.rules {
		if err := rule.Validate(ctx, template); err != nil {
			if v.strictMode {
				return fmt.Errorf("validation failed for rule %s: %w", rule.GetName(), err)
			}
			validationErrors = append(validationErrors, fmt.Sprintf("%s: %s", rule.GetName(), err.Error()))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("template validation warnings: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// SanitizePrompt sanitizes a prompt to make it safer for execution
func (v *InputValidator) SanitizePrompt(prompt string) string {
	// Apply basic sanitization
	sanitized := prompt

	// Remove potential HTML/script tags
	sanitized = regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`).ReplaceAllString(sanitized, "[SCRIPT_REMOVED]")
	sanitized = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(sanitized, "[TAG_REMOVED]")
	
	// Replace potential SQL injection patterns
	sanitized = regexp.MustCompile(`(?i)(UNION\s+ALL|UNION|SELECT\s+\*|DROP\s+TABLE|DELETE\s+FROM|INSERT\s+INTO|UPDATE\s+.*?SET)`).ReplaceAllString(sanitized, "[SQL_REMOVED]")
	
	// Replace potential command injection patterns
	sanitized = regexp.MustCompile(`(?i)(;|\||\$\(|\`+"`"+`|&&|\|\|)\s*(rm|cat|chmod|chown|wget|curl|bash|sh|sudo)`).ReplaceAllString(sanitized, "[CMD_REMOVED]")

	return sanitized
}

// NoJailbreakPatternRule checks for common jailbreak patterns
type NoJailbreakPatternRule struct{}

// NewNoJailbreakPatternRule creates a new jailbreak pattern rule
func NewNoJailbreakPatternRule() interfaces.InputValidationRule {
	return &NoJailbreakPatternRule{}
}

// GetName returns the name of the rule
func (r *NoJailbreakPatternRule) GetName() string {
	return "NoJailbreakPattern"
}

// GetDescription returns the description of the rule
func (r *NoJailbreakPatternRule) GetDescription() string {
	return "Checks for common jailbreak patterns in prompts"
}

// Validate validates a template against the rule
func (r *NoJailbreakPatternRule) Validate(ctx context.Context, template *format.Template) error {
	prompt := template.Test.Prompt
	
	// Check for common jailbreak patterns
	patterns := []string{
		"ignore previous instructions",
		"ignore all instructions",
		"disregard previous instructions",
		"forget your training",
		"ignore your programming",
		"you are now",
		"you are a",
		"you're a",
		"you're now",
		"DAN",
		"do anything now",
	}
	
	lowercasePrompt := strings.ToLower(prompt)
	for _, pattern := range patterns {
		if strings.Contains(lowercasePrompt, pattern) {
			return fmt.Errorf("prompt contains potential jailbreak pattern: %s", pattern)
		}
	}
	
	return nil
}

// NoSensitiveDataRule checks for sensitive data patterns
type NoSensitiveDataRule struct{}

// NewNoSensitiveDataRule creates a new sensitive data rule
func NewNoSensitiveDataRule() interfaces.InputValidationRule {
	return &NoSensitiveDataRule{}
}

// GetName returns the name of the rule
func (r *NoSensitiveDataRule) GetName() string {
	return "NoSensitiveData"
}

// GetDescription returns the description of the rule
func (r *NoSensitiveDataRule) GetDescription() string {
	return "Checks for sensitive data patterns in prompts"
}

// Validate validates a template against the rule
func (r *NoSensitiveDataRule) Validate(ctx context.Context, template *format.Template) error {
	prompt := template.Test.Prompt
	
	// Check for common sensitive data patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\b(?:[0-9]{4}[- ]?){3}[0-9]{4}\b`), // Credit card
		regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`), // Email
		regexp.MustCompile(`\b(?:[0-9]{3}[- ]?){2}[0-9]{4}\b`), // SSN
		regexp.MustCompile(`\b(?:[0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b`), // MAC address
		regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`), // IP address
	}
	
	for _, pattern := range patterns {
		if pattern.MatchString(prompt) {
			return fmt.Errorf("prompt contains potential sensitive data pattern")
		}
	}
	
	return nil
}

// MaxPromptLengthRule checks if the prompt exceeds the maximum length
type MaxPromptLengthRule struct {
	maxLength int
}

// NewMaxPromptLengthRule creates a new maximum prompt length rule
func NewMaxPromptLengthRule(maxLength int) interfaces.InputValidationRule {
	return &MaxPromptLengthRule{
		maxLength: maxLength,
	}
}

// GetName returns the name of the rule
func (r *MaxPromptLengthRule) GetName() string {
	return "MaxPromptLength"
}

// GetDescription returns the description of the rule
func (r *MaxPromptLengthRule) GetDescription() string {
	return fmt.Sprintf("Checks if the prompt exceeds %d characters", r.maxLength)
}

// Validate validates a template against the rule
func (r *MaxPromptLengthRule) Validate(ctx context.Context, template *format.Template) error {
	if len(template.Test.Prompt) > r.maxLength {
		return fmt.Errorf("prompt exceeds maximum length of %d characters", r.maxLength)
	}
	return nil
}

// SanitizeHTMLRule checks for and sanitizes HTML content
type SanitizeHTMLRule struct{}

// NewSanitizeHTMLRule creates a new HTML sanitization rule
func NewSanitizeHTMLRule() interfaces.InputValidationRule {
	return &SanitizeHTMLRule{}
}

// GetName returns the name of the rule
func (r *SanitizeHTMLRule) GetName() string {
	return "SanitizeHTML"
}

// GetDescription returns the description of the rule
func (r *SanitizeHTMLRule) GetDescription() string {
	return "Checks for and warns about HTML content in prompts"
}

// Validate validates a template against the rule
func (r *SanitizeHTMLRule) Validate(ctx context.Context, template *format.Template) error {
	prompt := template.Test.Prompt
	
	if regexp.MustCompile(`<[^>]*>`).MatchString(prompt) {
		return fmt.Errorf("prompt contains HTML tags that should be sanitized")
	}
	
	return nil
}

// SanitizeScriptRule checks for script tags and JavaScript code
type SanitizeScriptRule struct{}

// NewSanitizeScriptRule creates a new script sanitization rule
func NewSanitizeScriptRule() interfaces.InputValidationRule {
	return &SanitizeScriptRule{}
}

// GetName returns the name of the rule
func (r *SanitizeScriptRule) GetName() string {
	return "SanitizeScript"
}

// GetDescription returns the description of the rule
func (r *SanitizeScriptRule) GetDescription() string {
	return "Checks for script tags and JavaScript code in prompts"
}

// Validate validates a template against the rule
func (r *SanitizeScriptRule) Validate(ctx context.Context, template *format.Template) error {
	prompt := template.Test.Prompt
	
	if regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`).MatchString(prompt) {
		return fmt.Errorf("prompt contains script tags that should be removed")
	}
	
	jsPatterns := []string{
		"javascript:",
		"document.cookie",
		"document.write",
		"window.location",
		"eval\\(",
		"setTimeout\\(",
		"setInterval\\(",
	}
	
	for _, pattern := range jsPatterns {
		if regexp.MustCompile(pattern).MatchString(prompt) {
			return fmt.Errorf("prompt contains JavaScript code that should be removed")
		}
	}
	
	return nil
}

// NoSQLInjectionRule checks for SQL injection patterns
type NoSQLInjectionRule struct{}

// NewNoSQLInjectionRule creates a new SQL injection rule
func NewNoSQLInjectionRule() interfaces.InputValidationRule {
	return &NoSQLInjectionRule{}
}

// GetName returns the name of the rule
func (r *NoSQLInjectionRule) GetName() string {
	return "NoSQLInjection"
}

// GetDescription returns the description of the rule
func (r *NoSQLInjectionRule) GetDescription() string {
	return "Checks for SQL injection patterns in prompts"
}

// Validate validates a template against the rule
func (r *NoSQLInjectionRule) Validate(ctx context.Context, template *format.Template) error {
	prompt := template.Test.Prompt
	
	sqlPatterns := []string{
		"(?i)\\bUNION\\s+ALL\\b",
		"(?i)\\bUNION\\b",
		"(?i)\\bSELECT\\s+\\*\\b",
		"(?i)\\bDROP\\s+TABLE\\b",
		"(?i)\\bDELETE\\s+FROM\\b",
		"(?i)\\bINSERT\\s+INTO\\b",
		"(?i)\\bUPDATE\\s+.*?SET\\b",
		"(?i)\\bALTER\\s+TABLE\\b",
		"(?i)\\bCREATE\\s+TABLE\\b",
		"(?i)\\bEXEC\\s+\\b",
		"(?i)\\bEXECUTE\\s+\\b",
	}
	
	for _, pattern := range sqlPatterns {
		if regexp.MustCompile(pattern).MatchString(prompt) {
			return fmt.Errorf("prompt contains potential SQL injection pattern")
		}
	}
	
	return nil
}

// NoCommandInjectionRule checks for command injection patterns
type NoCommandInjectionRule struct{}

// NewNoCommandInjectionRule creates a new command injection rule
func NewNoCommandInjectionRule() interfaces.InputValidationRule {
	return &NoCommandInjectionRule{}
}

// GetName returns the name of the rule
func (r *NoCommandInjectionRule) GetName() string {
	return "NoCommandInjection"
}

// GetDescription returns the description of the rule
func (r *NoCommandInjectionRule) GetDescription() string {
	return "Checks for command injection patterns in prompts"
}

// Validate validates a template against the rule
func (r *NoCommandInjectionRule) Validate(ctx context.Context, template *format.Template) error {
	prompt := template.Test.Prompt
	
	// Check for common command injection patterns
	cmdPatterns := []string{
		"(?i)(;|\\||\\$\\(|`|&&|\\|\\|)\\s*(rm|cat|chmod|chown|wget|curl|bash|sh|sudo)",
		"(?i)\\b(system|exec|popen|subprocess\\.call)\\s*\\(",
		"(?i)\\b(os\\.system|os\\.popen|os\\.exec)\\s*\\(",
	}
	
	for _, pattern := range cmdPatterns {
		if regexp.MustCompile(pattern).MatchString(prompt) {
			return fmt.Errorf("prompt contains potential command injection pattern")
		}
	}
	
	return nil
}
