// Package security provides template security verification mechanisms
package security

import (
	"context"
	"fmt"
	"regexp"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateVerifier is the interface for template security verifiers
type TemplateVerifier interface {
	// VerifyTemplate verifies a template for security issues
	VerifyTemplate(ctx context.Context, template *format.Template, options *VerificationOptions) (*VerificationResult, error)
	
	// VerifyTemplateFile verifies a template file for security issues
	VerifyTemplateFile(ctx context.Context, templatePath string, options *VerificationOptions) (*VerificationResult, error)
	
	// VerifyTemplateDirectory verifies all templates in a directory for security issues
	VerifyTemplateDirectory(ctx context.Context, directoryPath string, options *VerificationOptions) ([]*VerificationResult, error)
	
	// RegisterCheck registers a custom security check
	RegisterCheck(name string, check SecurityCheck)
	
	// GetChecks returns all registered security checks
	GetChecks() map[string]SecurityCheck
}

// SecurityCheck is the interface for template security checks
type SecurityCheck interface {
	// Name returns the name of the check
	Name() string
	
	// Description returns a description of the check
	Description() string
	
	// Check checks a template for security issues
	Check(template *format.Template, options *VerificationOptions) []*SecurityIssue
}

// DefaultTemplateVerifier is the default implementation of TemplateVerifier
type DefaultTemplateVerifier struct {
	checks map[string]SecurityCheck
}

// NewTemplateVerifier creates a new template verifier with default security checks
func NewTemplateVerifier() *DefaultTemplateVerifier {
	verifier := &DefaultTemplateVerifier{
		checks: make(map[string]SecurityCheck),
	}
	
	// Register default security checks
	verifier.RegisterCheck("injection_patterns", NewInjectionPatternCheck())
	verifier.RegisterCheck("regex_safety", NewRegexSafetyCheck())
	verifier.RegisterCheck("input_validation", NewInputValidationCheck())
	verifier.RegisterCheck("template_format", NewTemplateFormatCheck())
	verifier.RegisterCheck("data_leakage", NewDataLeakageCheck())
	
	return verifier
}

// VerifyTemplate verifies a template for security issues
func (v *DefaultTemplateVerifier) VerifyTemplate(ctx context.Context, template *format.Template, options *VerificationOptions) (*VerificationResult, error) {
	if template == nil {
		return nil, fmt.Errorf("template cannot be nil")
	}
	
	if options == nil {
		options = DefaultVerificationOptions()
	}
	
	result := &VerificationResult{
		TemplateID:   template.ID,
		TemplateName: template.Info.Name,
		Issues:       []*SecurityIssue{},
		Passed:       true,
		Score:        100.0,
		MaxScore:     100.0,
		Metadata:     make(map[string]interface{}),
	}
	
	// Run all security checks
	for name, check := range v.checks {
		// Skip checks that are not in the custom checks list if it's provided
		if len(options.CustomChecks) > 0 {
			found := false
			for _, customCheck := range options.CustomChecks {
				if customCheck == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		issues := check.Check(template, options)
		
		// Filter issues based on severity threshold
		var filteredIssues []*SecurityIssue
		for _, issue := range issues {
			if !options.IncludeInfo && issue.Severity == common.Info {
				continue
			}
			
			if isSeverityHigher(options.SeverityThreshold, issue.Severity) {
				continue
			}
			
			// Check if the issue should be ignored based on patterns
			ignored := false
			for _, pattern := range options.IgnorePatterns {
				if matched, _ := regexp.MatchString(pattern, issue.Description); matched {
					ignored = true
					break
				}
			}
			
			if !ignored {
				filteredIssues = append(filteredIssues, issue)
			}
		}
		
		result.Issues = append(result.Issues, filteredIssues...)
	}
	
	// Calculate score based on issues
	if len(result.Issues) > 0 {
		// Deduct points for each issue based on severity
		totalDeduction := 0.0
		for _, issue := range result.Issues {
			switch issue.Severity {
			case common.Critical:
				totalDeduction += 30.0
			case common.High:
				totalDeduction += 20.0
			case common.Medium:
				totalDeduction += 10.0
			case common.Low:
				totalDeduction += 5.0
			case common.Info:
				totalDeduction += 1.0
			}
		}
		
		// Ensure score doesn't go below 0
		result.Score = max(0.0, 100.0-totalDeduction)
		
		// Set passed flag based on score threshold
		result.Passed = result.Score >= 70.0
	}
	
	return result, nil
}

// VerifyTemplateFile verifies a template file for security issues
func (v *DefaultTemplateVerifier) VerifyTemplateFile(ctx context.Context, templatePath string, options *VerificationOptions) (*VerificationResult, error) {
	// Load template from file
	template, err := format.LoadFromFile(templatePath)
	if err != nil {
		return &VerificationResult{
			TemplatePath: templatePath,
			TemplateID:   filepath.Base(templatePath),
			TemplateName: filepath.Base(templatePath),
			Issues: []*SecurityIssue{
				{
					Type:        TemplateFormatError,
					Description: fmt.Sprintf("Failed to load template: %v", err),
					Location:    templatePath,
					Severity:    common.High,
					Remediation: "Fix the template format according to the error message",
				},
			},
			Passed:   false,
			Score:    0.0,
			MaxScore: 100.0,
			Metadata: make(map[string]interface{}),
		}, nil
	}
	
	// Verify the template
	result, err := v.VerifyTemplate(ctx, template, options)
	if err != nil {
		return nil, err
	}
	
	// Set the template path
	result.TemplatePath = templatePath
	
	return result, nil
}

// VerifyTemplateDirectory verifies all templates in a directory for security issues
func (v *DefaultTemplateVerifier) VerifyTemplateDirectory(ctx context.Context, directoryPath string, options *VerificationOptions) ([]*VerificationResult, error) {
	// Find all template files in the directory
	templateFiles, err := filepath.Glob(filepath.Join(directoryPath, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to find template files: %w", err)
	}
	
	// Also check for .yml files
	ymlFiles, err := filepath.Glob(filepath.Join(directoryPath, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to find template files: %w", err)
	}
	
	templateFiles = append(templateFiles, ymlFiles...)
	
	// Verify each template file
	var results []*VerificationResult
	for _, templateFile := range templateFiles {
		result, err := v.VerifyTemplateFile(ctx, templateFile, options)
		if err != nil {
			return nil, fmt.Errorf("failed to verify template %s: %w", templateFile, err)
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// RegisterCheck registers a custom security check
func (v *DefaultTemplateVerifier) RegisterCheck(name string, check SecurityCheck) {
	v.checks[name] = check
}

// GetChecks returns all registered security checks
func (v *DefaultTemplateVerifier) GetChecks() map[string]SecurityCheck {
	return v.checks
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
