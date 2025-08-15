package sandbox

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
)

// ValidationLevel defines the level of validation to perform
type ValidationLevel string

const (
	// ValidationLevelBasic performs basic validation
	ValidationLevelBasic ValidationLevel = "basic"
	// ValidationLevelStandard performs standard validation
	ValidationLevelStandard ValidationLevel = "standard"
	// ValidationLevelStrict performs strict validation
	ValidationLevelStrict ValidationLevel = "strict"
)

// ValidationOptions defines options for template validation
type ValidationOptions struct {
	// Level is the validation level
	Level ValidationLevel
	// CustomChecks is a list of custom checks to perform
	CustomChecks []security.SecurityCheck
	// IgnorePatterns is a list of patterns to ignore
	IgnorePatterns []*regexp.Regexp
	// AllowedFunctions is a list of allowed functions
	AllowedFunctions []string
	// AllowedPackages is a list of allowed packages
	AllowedPackages []string
	// DisallowedFunctions is a list of disallowed functions
	DisallowedFunctions []string
	// DisallowedPackages is a list of disallowed packages
	DisallowedPackages []string
	// MaxComplexity is the maximum allowed complexity
	MaxComplexity int
	// MaxLineLength is the maximum allowed line length
	MaxLineLength int
	// RequireComments determines if comments are required
	RequireComments bool
	// RequireValidation determines if input validation is required
	RequireValidation bool
	// SecurityOptions are the security verification options
	SecurityOptions *security.VerificationOptions

// DefaultValidationOptions returns the default validation options
func DefaultValidationOptions() *ValidationOptions {
	return &ValidationOptions{
		Level:              ValidationLevelStandard,
		CustomChecks:       []security.SecurityCheck{},
		IgnorePatterns:     []*regexp.Regexp{},
		AllowedFunctions:   []string{},
		AllowedPackages:    []string{},
		DisallowedFunctions: []string{
			"os.Exit",
			"syscall",
			"unsafe",
			"runtime.SetFinalizer",
		},
		DisallowedPackages: []string{
			"os/exec",
			"syscall",
			"unsafe",
		},
		MaxComplexity:     10,
		MaxLineLength:     100,
		RequireComments:   false,
		RequireValidation: true,
		SecurityOptions:   security.DefaultVerificationOptions(),
	}

// TemplateValidator is responsible for validating templates
type TemplateValidator struct {
	verifier security.TemplateVerifier
	options  *ValidationOptions

// NewTemplateValidator creates a new template validator
func NewTemplateValidator(verifier security.TemplateVerifier, options *ValidationOptions) *TemplateValidator {
	if options == nil {
		options = DefaultValidationOptions()
	}
	
	return &TemplateValidator{
		verifier: verifier,
		options:  options,
	}

// Validate validates a template
func (v *TemplateValidator) Validate(ctx context.Context, template *format.Template) ([]*security.SecurityIssue, error) {
	var allIssues []*security.SecurityIssue
	
	// Perform security verification
	result, err := v.verifier.VerifyTemplate(ctx, template, v.options.SecurityOptions)
	if err != nil {
		return nil, fmt.Errorf("security verification failed: %w", err)
	}
	
	allIssues = append(allIssues, result.Issues...)
	
	// Perform syntax validation
	syntaxIssues, err := v.validateSyntax(template)
	if err != nil {
		return nil, fmt.Errorf("syntax validation failed: %w", err)
	}
	
	allIssues = append(allIssues, syntaxIssues...)
	
	// Perform semantic validation
	semanticIssues, err := v.validateSemantics(template)
	if err != nil {
		return nil, fmt.Errorf("semantic validation failed: %w", err)
	}
	
	allIssues = append(allIssues, semanticIssues...)
	
	// Perform custom checks
	for _, check := range v.options.CustomChecks {
		customIssues := check.Check(template, v.options.SecurityOptions)
		allIssues = append(allIssues, customIssues...)
	}
	
	return allIssues, nil

// ValidateFile validates a template file
func (v *TemplateValidator) ValidateFile(ctx context.Context, templatePath string) ([]*security.SecurityIssue, error) {
	// Read the template file
	content, err := ioutil.ReadFile(filepath.Clean(templatePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}
	
	// Parse the template
	template, err := format.ParseTemplate(string(content), filepath.Base(templatePath))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Set template path
	template.Path = templatePath
	
	// Validate the template
	return v.Validate(ctx, template)

// validateSyntax validates the syntax of a template
func (v *TemplateValidator) validateSyntax(template *format.Template) ([]*security.SecurityIssue, error) {
	var issues []*security.SecurityIssue
	
	if template.Content == "" {
		issues = append(issues, &security.SecurityIssue{
			Type:        security.TemplateFormatError,
			Description: "Template content is empty",
			Severity:    common.SeverityHigh,
			Remediation: "Provide content for the template",
		})
	}
	
	// Check for balanced braces, parentheses, and brackets
	if !hasBalancedDelimiters(template.Content) {
		issues = append(issues, &security.SecurityIssue{
			Type:        security.TemplateFormatError,
			Description: "Template has unbalanced delimiters (braces, parentheses, or brackets)",
			Severity:    common.SeverityHigh,
			Remediation: "Ensure all delimiters are properly balanced",
			Context:     template.Content,
		})
	}
	
	// Check for line length
	lines := strings.Split(template.Content, "\n")
	for i, line := range lines {
		if len(line) > v.options.MaxLineLength {
			issues = append(issues, &security.SecurityIssue{
				Type:        security.TemplateFormatError,
				Description: fmt.Sprintf("Line %d exceeds maximum length of %d characters", i+1, v.options.MaxLineLength),
				Severity:    common.SeverityLow,
				Remediation: "Reduce the line length",
				LineNumber:  i + 1,
				Context:     line,
			})
		}
	}
	
	return issues, nil

// validateSemantics validates the semantics of a template
func (v *TemplateValidator) validateSemantics(template *format.Template) ([]*security.SecurityIssue, error) {
	var issues []*security.SecurityIssue
	
	// Check for disallowed functions
	for _, disallowedFunc := range v.options.DisallowedFunctions {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(disallowedFunc))
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		if re.MatchString(template.Content) {
			issues = append(issues, &security.SecurityIssue{
				Type:        security.InsecurePattern,
				Description: fmt.Sprintf("Template contains disallowed function: %s", disallowedFunc),
				Severity:    common.SeverityHigh,
				Remediation: fmt.Sprintf("Remove the use of the disallowed function: %s", disallowedFunc),
				Context:     template.Content,
			})
		}
	}
	
	// Check for disallowed packages
	for _, disallowedPkg := range v.options.DisallowedPackages {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(disallowedPkg))
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		if re.MatchString(template.Content) {
			issues = append(issues, &security.SecurityIssue{
				Type:        security.InsecurePattern,
				Description: fmt.Sprintf("Template contains disallowed package: %s", disallowedPkg),
				Severity:    common.SeverityHigh,
				Remediation: fmt.Sprintf("Remove the use of the disallowed package: %s", disallowedPkg),
				Context:     template.Content,
			})
		}
	}
	
	// Check for input validation if required
	if v.options.RequireValidation {
		// Look for common validation patterns
		validationPatterns := []string{
			`validate`,
			`validation`,
			`check`,
			`verify`,
			`sanitize`,
			`escape`,
		}
		
		hasValidation := false
		for _, pattern := range validationPatterns {
			if strings.Contains(strings.ToLower(template.Content), pattern) {
				hasValidation = true
				break
			}
		}
		
		if !hasValidation {
			issues = append(issues, &security.SecurityIssue{
				Type:        security.MissingValidation,
				Description: "Template does not appear to include input validation",
				Severity:    common.SeverityMedium,
				Remediation: "Add input validation to the template",
				Context:     template.Content,
			})
		}
	}
	
	// Check for comments if required
	if v.options.RequireComments {
		// Look for comment patterns
		commentPatterns := []string{
			`//`,
			`/*`,
			`#`,
		}
		
		hasComments := false
		for _, pattern := range commentPatterns {
			if strings.Contains(template.Content, pattern) {
				hasComments = true
				break
			}
		}
		
		if !hasComments {
			issues = append(issues, &security.SecurityIssue{
				Type:        security.TemplateFormatError,
				Description: "Template does not include comments",
				Severity:    common.SeverityLow,
				Remediation: "Add comments to the template for better readability",
				Context:     template.Content,
			})
		}
	}
	
	return issues, nil

// hasBalancedDelimiters checks if a string has balanced delimiters
func hasBalancedDelimiters(s string) bool {
	var stack []rune
	
	for _, c := range s {
		switch c {
		case '(', '[', '{':
			stack = append(stack, c)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return false
			}
			stack = stack[:len(stack)-1]
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	
