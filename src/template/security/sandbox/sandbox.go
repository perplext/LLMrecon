package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
)

// DefaultSandbox is the default implementation of TemplateSandbox
type DefaultSandbox struct {
	verifier  security.TemplateVerifier
	allowList *AllowList
	options   *SandboxOptions
}

// NewSandbox creates a new template sandbox
func NewSandbox(verifier security.TemplateVerifier, options *SandboxOptions) *DefaultSandbox {
	if options == nil {
		options = DefaultSandboxOptions()
	}
	
	return &DefaultSandbox{
		verifier:  verifier,
		allowList: NewAllowList(),
		options:   options,
	}
}

// Execute executes a template in the sandbox
func (s *DefaultSandbox) Execute(ctx context.Context, template *format.Template, options *SandboxOptions) (*ExecutionResult, error) {
	if options == nil {
		options = s.options
	}
	
	// Validate the template first
	issues, err := s.Validate(ctx, template, options)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Template validation failed: %v", err),
		}, err
	}
	
	// If there are critical security issues, don't execute the template
	for _, issue := range issues {
		if issue.Severity == "critical" {
			return &ExecutionResult{
				Success:       false,
				Error:         fmt.Sprintf("Critical security issue found: %s", issue.Description),
				SecurityIssues: issues,
			}, errors.New("critical security issue found")
		}
	}
	
	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, options.TimeoutDuration)
	defer cancel()
	
	startTime := time.Now()
	
	// Execute the template in a controlled environment
	result, err := s.executeInSandbox(execCtx, template, options)
	
	executionTime := time.Since(startTime)
	
	if err != nil {
		return &ExecutionResult{
			Success:       false,
			Error:         fmt.Sprintf("Template execution failed: %v", err),
			ExecutionTime: executionTime,
			SecurityIssues: issues,
		}, err
	}
	
	result.ExecutionTime = executionTime
	result.SecurityIssues = issues
	
	return result, nil
}

// executeInSandbox executes a template in a controlled environment
func (s *DefaultSandbox) executeInSandbox(ctx context.Context, template *format.Template, options *SandboxOptions) (*ExecutionResult, error) {
	// Create a temporary directory for sandbox execution
	tempDir, err := ioutil.TempDir("", "template-sandbox-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Apply resource limits
	resourceLimits := options.ResourceLimits
	
	// TODO: Implement actual containerization using Docker or similar
	// For now, we'll just simulate the sandbox execution
	
	// Check if the context is done (timeout or cancellation)
	select {
	case <-ctx.Done():
		return &ExecutionResult{
			Success: false,
			Error:   "Template execution timed out",
			ResourceUsage: ResourceUsage{
				ExecutionTime: options.TimeoutDuration,
			},
		}, ctx.Err()
	default:
		// Continue execution
	}
	
	// Simulate template execution
	output, err := s.simulateTemplateExecution(template, tempDir, options)
	
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Template execution failed: %v", err),
			ResourceUsage: ResourceUsage{
				ExecutionTime: time.Since(time.Now().Add(-options.TimeoutDuration)),
				CPUTime:       0.1,
				MemoryUsage:   10,
			},
		}, err
	}
	
	return &ExecutionResult{
		Success: true,
		Output:  output,
		ResourceUsage: ResourceUsage{
			ExecutionTime: time.Since(time.Now().Add(-options.TimeoutDuration)),
			CPUTime:       0.1,
			MemoryUsage:   10,
		},
	}, nil
}

// simulateTemplateExecution simulates template execution
// In a real implementation, this would use actual containerization
func (s *DefaultSandbox) simulateTemplateExecution(template *format.Template, tempDir string, options *SandboxOptions) (string, error) {
	// Check for disallowed functions and packages
	content := template.Content
	
	for _, disallowedFunc := range options.DisallowedFunctions {
		if strings.Contains(content, disallowedFunc) {
			return "", fmt.Errorf("template contains disallowed function: %s", disallowedFunc)
		}
	}
	
	for _, disallowedPkg := range options.DisallowedPackages {
		if strings.Contains(content, disallowedPkg) {
			return "", fmt.Errorf("template contains disallowed package: %s", disallowedPkg)
		}
	}
	
	// In a real implementation, we would execute the template in a container
	// For now, we'll just return a simulated output
	return fmt.Sprintf("Simulated execution of template: %s", template.Name), nil
}

// ExecuteFile executes a template file in the sandbox
func (s *DefaultSandbox) ExecuteFile(ctx context.Context, templatePath string, options *SandboxOptions) (*ExecutionResult, error) {
	if options == nil {
		options = s.options
	}
	
	// Read the template file
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to read template file: %v", err),
		}, err
	}
	
	// Parse the template
	template, err := format.ParseTemplate(string(content), filepath.Base(templatePath))
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse template: %v", err),
		}, err
	}
	
	// Set template path
	template.Path = templatePath
	
	// Execute the template
	return s.Execute(ctx, template, options)
}

// Validate validates a template against security rules
func (s *DefaultSandbox) Validate(ctx context.Context, template *format.Template, options *SandboxOptions) ([]*security.SecurityIssue, error) {
	if options == nil {
		options = s.options
	}
	
	// Use the verifier to validate the template
	result, err := s.verifier.VerifyTemplate(ctx, template, options.ValidationOptions)
	if err != nil {
		return nil, err
	}
	
	// Perform additional sandbox-specific validation
	additionalIssues := s.validateSandboxRules(template, options)
	
	// Combine issues
	allIssues := append(result.Issues, additionalIssues...)
	
	return allIssues, nil
}

// validateSandboxRules performs sandbox-specific validation
func (s *DefaultSandbox) validateSandboxRules(template *format.Template, options *SandboxOptions) []*security.SecurityIssue {
	var issues []*security.SecurityIssue
	
	// Check for disallowed functions
	for _, disallowedFunc := range options.DisallowedFunctions {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(disallowedFunc))
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		if re.MatchString(template.Content) {
			issues = append(issues, &security.SecurityIssue{
				Type:        security.InsecurePattern,
				Description: fmt.Sprintf("Template contains disallowed function: %s", disallowedFunc),
				Severity:    "high",
				Remediation: fmt.Sprintf("Remove the use of the disallowed function: %s", disallowedFunc),
				Context:     template.Content,
			})
		}
	}
	
	// Check for disallowed packages
	for _, disallowedPkg := range options.DisallowedPackages {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(disallowedPkg))
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		if re.MatchString(template.Content) {
			issues = append(issues, &security.SecurityIssue{
				Type:        security.InsecurePattern,
				Description: fmt.Sprintf("Template contains disallowed package: %s", disallowedPkg),
				Severity:    "high",
				Remediation: fmt.Sprintf("Remove the use of the disallowed package: %s", disallowedPkg),
				Context:     template.Content,
			})
		}
	}
	
	return issues
}

// ValidateFile validates a template file against security rules
func (s *DefaultSandbox) ValidateFile(ctx context.Context, templatePath string, options *SandboxOptions) ([]*security.SecurityIssue, error) {
	if options == nil {
		options = s.options
	}
	
	// Read the template file
	content, err := ioutil.ReadFile(templatePath)
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
	return s.Validate(ctx, template, options)
}

// GetAllowList returns the allow list for template execution
func (s *DefaultSandbox) GetAllowList() *AllowList {
	return s.allowList
}

// SetAllowList sets the allow list for template execution
func (s *DefaultSandbox) SetAllowList(allowList *AllowList) {
	s.allowList = allowList
}
