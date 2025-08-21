// Package management provides functionality for managing templates in the LLMreconing Tool.
// It includes components for template loading, parsing, execution, and reporting.
package management

import (
	"context"
	"fmt"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/cache"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/parser"
	"github.com/perplext/LLMrecon/src/template/management/registry"
	"github.com/perplext/LLMrecon/src/template/management/reporting"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

const (
	// FileSource indicates the template is from a file
	FileSource = interfaces.FileSource
	// DirectorySource indicates the template is from a directory
	DirectorySource = interfaces.DirectorySource
	// GitHubSource indicates the template is from GitHub
	GitHubSource = interfaces.GitHubSource
	// GitLabSource indicates the template is from GitLab
	GitLabSource = interfaces.GitLabSource
	// HTTPSource indicates the template is from HTTP
	HTTPSource = interfaces.HTTPSource
	// DatabaseSource indicates the template is from a database
	DatabaseSource = interfaces.DatabaseSource
)

// TemplateStatus is an alias for interfaces.TemplateStatus
type TemplateStatus = interfaces.TemplateStatus

const (
	// StatusLoaded indicates the template has been loaded
	StatusLoaded = interfaces.StatusLoaded
	// StatusValidated indicates the template has been validated
	StatusValidated = interfaces.StatusValidated
	// StatusExecuting indicates the template is being executed
	StatusExecuting = interfaces.StatusExecuting
	// StatusCompleted indicates the template execution has completed
	StatusCompleted = interfaces.StatusCompleted
	// StatusFailed indicates the template execution has failed
	StatusFailed = interfaces.StatusFailed
)

// TemplateManagerOptions contains options for the template manager
type TemplateManagerOptions struct {
	// Loaders is the list of template loaders
	Loaders []types.TemplateLoader
	// Parser is the template parser
	Parser *parser.TemplateParser
	// Executor is the template executor
	Executor *execution.TemplateExecutor
	// Reporter is the template reporter
	Reporter *reporting.TemplateReporter
	// Cache is the template cache
	Cache *cache.TemplateCache
	// Registry is the template registry
	Registry *registry.TemplateRegistry
	// PreExecutionHooks are functions to run before template execution
	PreExecutionHooks []func(ctx context.Context, template *format.Template, result *interfaces.TemplateResult) error
	// PostExecutionHooks are functions to run after template execution
	PostExecutionHooks []func(ctx context.Context, template *format.Template, result *interfaces.TemplateResult) error
}

// DefaultTemplateManager is the default implementation of TemplateManager
type DefaultTemplateManager struct {
	// loaders is the list of template loaders
	loaders []types.TemplateLoader
	// parser is the template parser
	parser *parser.TemplateParser
	// executor is the template executor
	executor *execution.TemplateExecutor
	// reporter is the template reporter
	reporter *reporting.TemplateReporter
	// cache is the template cache
	cache *cache.TemplateCache
	// registry is the template registry
	registry *registry.TemplateRegistry
	// preExecutionHooks are functions to run before template execution
	preExecutionHooks []func(ctx context.Context, template *format.Template, result *interfaces.TemplateResult) error
	// postExecutionHooks are functions to run after template execution
	postExecutionHooks []func(ctx context.Context, template *format.Template, result *interfaces.TemplateResult) error
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(options *TemplateManagerOptions) (*DefaultTemplateManager, error) {
	if options.Parser == nil {
		return nil, fmt.Errorf("parser is required")
	}
	if options.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}
	if options.Reporter == nil {
		return nil, fmt.Errorf("reporter is required")
	}
	if options.Registry == nil {
		return nil, fmt.Errorf("registry is required")
	}

	manager := &DefaultTemplateManager{
		loaders:            options.Loaders,
		parser:             options.Parser,
		executor:           options.Executor,
		reporter:           options.Reporter,
		cache:              options.Cache,
		registry:           options.Registry,
		preExecutionHooks:  options.PreExecutionHooks,
		postExecutionHooks: options.PostExecutionHooks,
	}
	return manager, nil
}

// LoadTemplate loads a template from a source
func (m *DefaultTemplateManager) LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error) {
	// Try each loader until one succeeds
	for _, loader := range m.loaders {
		// Check if this loader can handle LoadTemplate
		if templateLoader, ok := loader.(types.TemplateLoader); ok {
			template, err := templateLoader.LoadTemplate(ctx, source, sourceType)
			if err == nil {
				// Validate and register the template
				if err := m.parser.Validate(template); err != nil {
					return nil, fmt.Errorf("failed to validate template: %w", err)
				}
				if err := m.registry.Register(template); err != nil {
					return nil, fmt.Errorf("failed to register template: %w", err)
				}

				if m.cache != nil {
					m.cache.Set(template.ID, template)
				}

				return template, nil
			}
		}
	}

	return nil, fmt.Errorf("no loader could handle source type: %s", sourceType)
}

// LoadTemplates loads multiple templates from a source
func (m *DefaultTemplateManager) LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error) {
	// Try each loader until one succeeds
	for _, loader := range m.loaders {
		// Check if this loader can handle LoadTemplates
		if templateLoader, ok := loader.(types.TemplateLoader); ok {
			templates, err := templateLoader.LoadTemplates(ctx, source, sourceType)
			if err == nil {
				// Validate and register each template
				for _, template := range templates {
					if err := m.parser.Validate(template); err != nil {
						return nil, fmt.Errorf("failed to validate template %s: %w", template.ID, err)
					}

					if err := m.registry.Register(template); err != nil {
						return nil, fmt.Errorf("failed to register template %s: %w", template.ID, err)
					}

					if m.cache != nil {
						m.cache.Set(template.ID, template)
					}
				}

				return templates, nil
			}
		}
	}

	return nil, fmt.Errorf("no loader could handle source type: %s", sourceType)
}

// LoadTemplatesFromSources loads templates from the specified sources
func (m *DefaultTemplateManager) LoadTemplatesFromSources(ctx context.Context, sources []types.TemplateSource) error {
	for _, source := range sources {
		// Load templates using the type-aware method
		_, err := m.LoadTemplates(ctx, source.Path, source.Type)
		if err != nil {
			return fmt.Errorf("failed to load templates from source %s: %w", source.Path, err)
		}

		// Templates are already validated and registered by LoadTemplates method
	}

	return nil
}

// GetTemplate gets a template by ID
func (m *DefaultTemplateManager) GetTemplate(id string) (*format.Template, error) {
	// Try to get from cache first
	if m.cache != nil {
		if template, found := m.cache.Get(id); found {
			return template, nil
		}
	}

	// Get from registry
	template, found := m.registry.Get(id)
	if !found {
		return nil, fmt.Errorf("template %s not found", id)
	}

	// Cache the template
	if m.cache != nil {
		m.cache.Set(id, template)
	}

	return template, nil
}

// ListTemplates lists all templates
func (m *DefaultTemplateManager) ListTemplates() []*format.Template {
	return m.registry.List()
}

// ValidateTemplate validates a template
func (m *DefaultTemplateManager) ValidateTemplate(template *format.Template) error {
	return m.parser.Validate(template)
}

// GetCategories returns a list of all template categories
func (m *DefaultTemplateManager) GetCategories() ([]string, error) {
	templates := m.ListTemplates()
	categories := make(map[string]bool)
	for _, template := range templates {
		if template.Metadata != nil {
			if category, ok := template.Metadata["category"].(string); ok {
				categories[category] = true
			}
		}
	}

	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	return result, nil
}

// Interface wrapper methods to match TemplateManager interface

// GetTemplateInterface wraps GetTemplate to match interface
func (m *DefaultTemplateManager) GetTemplateInterface(id string) (*format.Template, error) {
	template, err := m.GetTemplate(id)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// ListTemplatesInterface wraps ListTemplates to match interface
func (m *DefaultTemplateManager) ListTemplatesInterface() ([]*format.Template, error) {
	templates := m.ListTemplates()
	result := make([]*format.Template, len(templates))
	for i, template := range templates {
		result[i] = template
	}
	return result, nil
}

// LoadTemplateInterface wraps LoadTemplate to match interface
func (m *DefaultTemplateManager) LoadTemplateInterface(path string) (*format.Template, error) {
	template, err := m.LoadTemplate(context.Background(), path, "file")
	if err != nil {
		return nil, err
	}
	return template, nil
}

// ValidateTemplateInterface wraps ValidateTemplate to match interface
func (m *DefaultTemplateManager) ValidateTemplateInterface(template *format.Template) error {
	return m.ValidateTemplate(template)
}

// ExecuteTemplate executes a template
func (m *DefaultTemplateManager) ExecuteTemplate(ctx context.Context, templateID string, options map[string]interface{}) (*types.TemplateResult, error) {
	// Get template
	template, err := m.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Create a copy of the template to avoid modifying the original
	templateCopy := *template

	// Resolve variables if provided
	if variables, ok := options["variables"].(map[string]interface{}); ok && len(variables) > 0 {
		if err := m.parser.ResolveVariables(&templateCopy, variables); err != nil {
			return nil, fmt.Errorf("failed to resolve variables: %w", err)
		}
	}

	// Create result
	result := &types.TemplateResult{
		TemplateID:   templateID,
		TemplateName: template.Info.Name,
		Description:  template.Info.Description,
		StartTime:    time.Now(),
		Status:       types.StatusExecuting,
		Tags:         template.Info.Tags,
	}
	// Create interface result for hooks
	ifaceResult := &interfaces.TemplateResult{
		TemplateID: templateID,
		StartTime:  result.StartTime,
		Status:     string(types.StatusExecuting),
	}

	// Run pre-execution hooks
	for _, hook := range m.preExecutionHooks {
		if err := hook(ctx, &templateCopy, ifaceResult); err != nil {
			result.Status = types.StatusFailed
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, fmt.Errorf("pre-execution hook failed: %w", err)
		}
	}

	// Execute template
	execResult, err := m.executor.Execute(ctx, &templateCopy, options)
	if err != nil {
		result.Status = types.StatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("execution failed: %w", err)
	}

	// Update result with execution result
	result.Response = execResult.Response
	result.Detected = execResult.Detected
	result.Score = execResult.Score
	result.Status = types.StatusCompleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Details = execResult.Details

	// Update interface result
	ifaceResult.Response = execResult.Response
	ifaceResult.Detected = execResult.Detected
	ifaceResult.Score = execResult.Score
	ifaceResult.Status = string(types.StatusCompleted)
	ifaceResult.EndTime = result.EndTime
	ifaceResult.Duration = result.Duration
	ifaceResult.Details = execResult.Details

	// Run post-execution hooks
	for _, hook := range m.postExecutionHooks {
		if err := hook(ctx, &templateCopy, ifaceResult); err != nil {
			// Don't fail the execution, just log the error
			if result.Details == nil {
				result.Details = make(map[string]interface{})
			}
			result.Details["postHookError"] = err.Error()
		}
	}

	return result, nil
}

// ExecuteTemplates executes multiple templates
func (m *DefaultTemplateManager) ExecuteTemplates(ctx context.Context, templateIDs []string, options map[string]interface{}) ([]*types.TemplateResult, error) {
	// Create result slice
	results := make([]*types.TemplateResult, 0, len(templateIDs))

	// Execute each template
	for _, templateID := range templateIDs {
		result, err := m.ExecuteTemplate(ctx, templateID, options)
		if err != nil {
			// Don't fail the entire batch, just record the error
			result = &types.TemplateResult{
				TemplateID: templateID,
				Error:      err,
				Status:     types.StatusFailed,
				StartTime:  time.Now(),
				EndTime:    time.Now(),
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// Execute executes a template
func (m *DefaultTemplateManager) Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*types.TemplateResult, error) {
	// Create result
	result := &types.TemplateResult{
		TemplateID:   template.ID,
		TemplateName: template.Info.Name,
		Description:  template.Info.Description,
		StartTime:    time.Now(),
		Status:       types.StatusExecuting,
		Tags:         template.Info.Tags,
	}

	// Execute template
	execResult, err := m.executor.Execute(ctx, template, options)
	if err != nil {
		result.Status = types.StatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Update result
	result.Response = execResult.Response
	result.Detected = execResult.Detected
	result.Score = execResult.Score
	result.Status = types.StatusCompleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Details = execResult.Details

	return result, nil
}

// ExecuteBatch executes multiple templates
func (m *DefaultTemplateManager) ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*types.TemplateResult, error) {
	results := make([]*types.TemplateResult, 0, len(templates))
	for _, template := range templates {
		result, err := m.Execute(ctx, template, options)
		if err != nil {
			// Don't fail the entire batch, just record the error
			result = &types.TemplateResult{
				TemplateID: template.ID,
				Error:      err,
				Status:     types.StatusFailed,
				StartTime:  time.Now(),
				EndTime:    time.Now(),
			}
		}
		results = append(results, result)
	}
	return results, nil
}

// GetLoader returns the template loader
func (m *DefaultTemplateManager) GetLoader() types.TemplateLoader {
	// Return the first loader if available
	if len(m.loaders) > 0 {
		return m.loaders[0]
	}
	return nil
}

// GetExecutor returns the template executor
func (m *DefaultTemplateManager) GetExecutor() interface{} {
	return m.executor
}

// GenerateReport generates a report for template execution results
func (m *DefaultTemplateManager) GenerateReport(results []*types.TemplateResult, format string) ([]byte, error) {
	// Validate format
	if format == "" {
		format = "json" // Default format
	}

	// Convert types.TemplateResult to interfaces.TemplateResult for the reporter
	ifaceResults := make([]*interfaces.TemplateResult, 0, len(results))
	for _, r := range results {
		ifaceResult := &interfaces.TemplateResult{
			TemplateID: r.TemplateID,
			Response:   r.Response,
			Error:      r.Error,
			Status:     string(r.Status),
			StartTime:  r.StartTime,
			EndTime:    r.EndTime,
			Duration:   r.Duration,
			Detected:   r.Detected,
			Score:      r.Score,
			Details:    r.Details,
		}
		ifaceResults = append(ifaceResults, ifaceResult)
	}

	// Generate report
	report, err := m.reporter.GenerateReport(ifaceResults, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return report, nil
}
