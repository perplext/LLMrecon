package sandbox

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
)

// SecurityFramework provides a comprehensive security framework for templates
type SecurityFramework struct {
	// validator is the template validator
	validator *TemplateValidator
	// sandbox is the template sandbox
	sandbox TemplateSandbox
	// scorer is the template scorer
	scorer *TemplateScorer
	// workflow is the approval workflow
	workflow *ApprovalWorkflow
	// options are the framework options
	options *FrameworkOptions
	// metrics is the metrics collector
	metrics *MetricsCollector
	// mutex is used for thread safety
	mutex sync.RWMutex
}

// FrameworkOptions contains options for the security framework
type FrameworkOptions struct {
	// ValidationOptions are the validation options
	ValidationOptions *ValidationOptions
	// SandboxOptions are the sandbox options
	SandboxOptions *SandboxOptions
	// WorkflowStorageDir is the directory for storing workflow data
	WorkflowStorageDir string
	// EnableContainerSandbox enables the container-based sandbox
	EnableContainerSandbox bool
	// ContainerOptions are the container sandbox options
	ContainerOptions *ContainerSandboxOptions
	// EnableLogging enables logging
	EnableLogging bool
	// LogDirectory is the directory for logs
	LogDirectory string
	// EnableMetrics enables metrics collection
	EnableMetrics bool
}

// DefaultFrameworkOptions returns the default framework options
func DefaultFrameworkOptions() *FrameworkOptions {
	return &FrameworkOptions{
		ValidationOptions:      DefaultValidationOptions(),
		SandboxOptions:         DefaultSandboxOptions(),
		WorkflowStorageDir:     "",
		EnableContainerSandbox: false,
		ContainerOptions:       DefaultContainerSandboxOptions(),
		EnableLogging:          true,
		LogDirectory:           "",
		EnableMetrics:          true,
	}
}

// NewSecurityFramework creates a new security framework
func NewSecurityFramework(options *FrameworkOptions) (*SecurityFramework, error) {
	if options == nil {
		options = DefaultFrameworkOptions()
	}
	
	// Create the verifier
	verifier := security.NewTemplateVerifier()
	
	// Create the validator
	validator := NewTemplateValidator(verifier, options.ValidationOptions)
	
	// Create the scorer
	scorer := NewTemplateScorer()
	
	// Create the sandbox
	var sandbox TemplateSandbox
	var err error
	
	if options.EnableContainerSandbox {
		// Create a container-based sandbox
		sandbox, err = NewContainerSandbox(verifier, options.SandboxOptions, options.ContainerOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create container sandbox: %w", err)
		}
	} else {
		// Create a default sandbox
		sandbox = NewSandbox(verifier, options.SandboxOptions)
	}
	
	// Create the workflow
	workflow := NewApprovalWorkflow(validator, scorer, options.WorkflowStorageDir)
	
	// Create the log directory if needed
	if options.EnableLogging && options.LogDirectory != "" {
		if err := os.MkdirAll(options.LogDirectory, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	
	// Create metrics collector if metrics are enabled
	var metrics *MetricsCollector
	if options.EnableMetrics {
		metrics = NewMetricsCollector()
	}

	return &SecurityFramework{
		validator: validator,
		sandbox:   sandbox,
		scorer:    scorer,
		workflow:  workflow,
		options:   options,
		metrics:   metrics,
		mutex:     sync.RWMutex{},
	}, nil
}

// ValidateTemplate validates a template
func (f *SecurityFramework) ValidateTemplate(ctx context.Context, template *format.Template) (*ValidationResult, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	startTime := time.Now()
	
	// Validate the template
	issues, err := f.validator.Validate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}
	
	// Score the template
	riskScore := f.scorer.ScoreTemplate(template, issues)
	
	// Create the validation result
	result := &ValidationResult{
		Template:      template,
		Issues:        issues,
		RiskScore:     riskScore,
		ValidationTime: time.Since(startTime),
		Timestamp:     time.Now(),
	}
	
	// Log the validation result
	if f.options.EnableLogging {
		f.logValidationResult(result)
	}
	
	// Record metrics
	if f.options.EnableMetrics {
		f.metrics.RecordValidation(result)
	}
	
	return result, nil
}

// ValidateTemplateFile validates a template file
func (f *SecurityFramework) ValidateTemplateFile(ctx context.Context, templatePath string) (*ValidationResult, error) {
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
	return f.ValidateTemplate(ctx, template)
}

// ExecuteTemplate executes a template in the sandbox
func (f *SecurityFramework) ExecuteTemplate(ctx context.Context, template *format.Template) (*ExecutionResult, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	// Execute the template in the sandbox
	result, err := f.sandbox.Execute(ctx, template, f.options.SandboxOptions)
	if err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}
	
	// Log the execution result
	if f.options.EnableLogging {
		f.logExecutionResult(result, template)
	}
	
	// Record metrics
	if f.options.EnableMetrics {
		f.metrics.RecordExecution(result, template.ID)
	}
	
	return result, nil
}

// ExecuteTemplateFile executes a template file in the sandbox
func (f *SecurityFramework) ExecuteTemplateFile(ctx context.Context, templatePath string) (*ExecutionResult, error) {
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
	
	// Execute the template
	return f.ExecuteTemplate(ctx, template)
}

// CreateTemplateVersion creates a new version of a template
func (f *SecurityFramework) CreateTemplateVersion(ctx context.Context, template *format.Template, user string) (*TemplateVersion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	// Create a new version
	version, err := f.workflow.CreateVersion(ctx, template, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create template version: %w", err)
	}
	
	// Record metrics
	if f.options.EnableMetrics {
		f.metrics.RecordWorkflowAction("create", version)
	}
	
	return version, nil
}

// GetTemplateVersion gets a template version
func (f *SecurityFramework) GetTemplateVersion(templateID, versionID string) (*TemplateVersion, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	return f.workflow.GetVersion(templateID, versionID)
}

// GetLatestTemplateVersion gets the latest version of a template
func (f *SecurityFramework) GetLatestTemplateVersion(templateID string) (*TemplateVersion, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	return f.workflow.GetLatestVersion(templateID)
}

// GetLatestApprovedTemplateVersion gets the latest approved version of a template
func (f *SecurityFramework) GetLatestApprovedTemplateVersion(templateID string) (*TemplateVersion, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	return f.workflow.GetLatestApprovedVersion(templateID)
}

// SubmitTemplateForReview submits a template version for review
func (f *SecurityFramework) SubmitTemplateForReview(templateID, versionID, user string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	err := f.workflow.SubmitForReview(templateID, versionID, user)
	if err != nil {
		return err
	}
	
	// Record metrics
	if f.options.EnableMetrics {
		version, _ := f.workflow.GetVersion(templateID, versionID)
		if version != nil {
			f.metrics.RecordWorkflowAction("submit", version)
		}
	}
	
	return nil
}

// ApproveTemplate approves a template version
func (f *SecurityFramework) ApproveTemplate(templateID, versionID, user string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	err := f.workflow.ApproveVersion(templateID, versionID, user)
	if err != nil {
		return err
	}
	
	// Record metrics
	if f.options.EnableMetrics {
		version, _ := f.workflow.GetVersion(templateID, versionID)
		if version != nil {
			f.metrics.RecordWorkflowAction("approve", version)
		}
	}
	
	return nil
}

// RejectTemplate rejects a template version
func (f *SecurityFramework) RejectTemplate(templateID, versionID, user, reason string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	err := f.workflow.RejectVersion(templateID, versionID, user, reason)
	if err != nil {
		return err
	}
	
	// Record metrics
	if f.options.EnableMetrics {
		version, _ := f.workflow.GetVersion(templateID, versionID)
		if version != nil {
			f.metrics.RecordWorkflowAction("reject", version)
		}
	}
	
	return nil
}

// AddApprover adds an approver to the workflow
func (f *SecurityFramework) AddApprover(approver string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	f.workflow.AddApprover(approver)
}

// IsApprover checks if a user is an approver
func (f *SecurityFramework) IsApprover(user string) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	return f.workflow.IsApprover(user)
}

// logValidationResult logs a validation result
func (f *SecurityFramework) logValidationResult(result *ValidationResult) {
	if f.options.LogDirectory == "" {
		return
	}
	
	// Create a log entry
	logEntry := fmt.Sprintf("[%s] Validation: %s (Risk: %s, Score: %.2f)\n", 
		time.Now().Format(time.RFC3339),
		result.Template.Name,
		result.RiskScore.Category,
		result.RiskScore.Score)
	
	// Add issues
	for _, issue := range result.Issues {
		logEntry += fmt.Sprintf("  - %s: %s (Severity: %s)\n", 
			issue.Type,
			issue.Description,
			issue.Severity)
	}
	
	// Write to log file
	logFile := filepath.Join(f.options.LogDirectory, "validation.log")
	f.appendToLogFile(logFile, logEntry)
}

// logExecutionResult logs an execution result
func (f *SecurityFramework) logExecutionResult(result *ExecutionResult, template *format.Template) {
	if f.options.LogDirectory == "" {
		return
	}
	
	// Create a log entry
	logEntry := fmt.Sprintf("[%s] Execution: %s (Success: %t, Time: %s)\n", 
		time.Now().Format(time.RFC3339),
		template.Name,
		result.Success,
		result.ExecutionTime)
	
	// Add error if any
	if result.Error != "" {
		logEntry += fmt.Sprintf("  Error: %s\n", result.Error)
	}
	
	// Add resource usage
	logEntry += fmt.Sprintf("  Resources: CPU=%.2fs, Memory=%dMB\n", 
		result.ResourceUsage.CPUTime,
		result.ResourceUsage.MemoryUsage)
	
	// Write to log file
	logFile := filepath.Join(f.options.LogDirectory, "execution.log")
	f.appendToLogFile(logFile, logEntry)
}

// appendToLogFile appends a log entry to a file
func (f *SecurityFramework) appendToLogFile(logFile, logEntry string) {
	// Open the log file in append mode
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	
	// Write the log entry
	file.WriteString(logEntry)
}

// GetMetrics returns the metrics from the metrics collector
func (f *SecurityFramework) GetMetrics() map[string]interface{} {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	if !f.options.EnableMetrics {
		return map[string]interface{}{
			"metrics_enabled": false,
		}
	}
	
	return map[string]interface{}{
		"validation": f.metrics.GetValidationMetrics(),
		"execution":  f.metrics.GetExecutionMetrics(),
		"workflow":   f.metrics.GetWorkflowMetrics(),
	}
}

// GetAlerts returns the alerts from the metrics collector
func (f *SecurityFramework) GetAlerts() []Alert {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	if !f.options.EnableMetrics {
		return []Alert{}
	}
	
	return f.metrics.GetAlerts()
}

// GetAlertsByLevel returns alerts by level from the metrics collector
func (f *SecurityFramework) GetAlertsByLevel(level AlertLevel) []Alert {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	if !f.options.EnableMetrics {
		return []Alert{}
	}
	
	return f.metrics.GetAlertsByLevel(level)
}

// ClearAlerts clears all alerts from the metrics collector
func (f *SecurityFramework) ClearAlerts() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	if !f.options.EnableMetrics {
		return
	}
	
	f.metrics.ClearAlerts()
}

// ResetMetrics resets all metrics in the metrics collector
func (f *SecurityFramework) ResetMetrics() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	if !f.options.EnableMetrics {
		return
	}
	
	f.metrics.ResetMetrics()
}

// ValidationResult represents the result of template validation
type ValidationResult struct {
	// Template is the validated template
	Template *format.Template
	// Issues are the security issues found
	Issues []*security.SecurityIssue
	// RiskScore is the risk score
	RiskScore *RiskScore
	// ValidationTime is the time taken for validation
	ValidationTime time.Duration
	// Timestamp is the time of validation
	Timestamp time.Time
}

// HasCriticalIssues checks if the validation result has critical issues
func (r *ValidationResult) HasCriticalIssues() bool {
	for _, issue := range r.Issues {
		if issue.Severity == common.SeverityCritical {
			return true
		}
	}
	return false
}

// HasHighIssues checks if the validation result has high severity issues
func (r *ValidationResult) HasHighIssues() bool {
	for _, issue := range r.Issues {
		if issue.Severity == common.SeverityHigh {
			return true
		}
	}
	return false
}

// GetIssuesBySeverity gets issues by severity
func (r *ValidationResult) GetIssuesBySeverity(severity common.SeverityLevel) []*security.SecurityIssue {
	var issues []*security.SecurityIssue
	for _, issue := range r.Issues {
		if issue.Severity == severity {
			issues = append(issues, issue)
		}
	}
	return issues
}
