// Package management provides functionality for managing templates in the LLMreconing Tool.
package management

import (
	"context"
	"fmt"
	"time"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/cache"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/loader"
	"github.com/perplext/LLMrecon/src/template/management/parser"
	"github.com/perplext/LLMrecon/src/template/management/registry"
	"github.com/perplext/LLMrecon/src/template/management/reporting"
)

// CreateDefaultManager creates a default template manager with all components
func CreateDefaultManager(ctx context.Context, options *DefaultManagerOptions) (*Manager, error) {
	// Create schema validator
	schemaValidator, err := DefaultSchemaValidator(options.JSONSchemaPath, options.YAMLSchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema validator: %w", err)
	}

	// Create template parser
	templateParser, err := parser.NewTemplateParser(schemaValidator)
	if err != nil {
		return nil, fmt.Errorf("failed to create template parser: %w", err)
	}

	// Create repository manager
	repoManager := repository.NewManager()
	// Configure the repository manager if needed
	if options.RepositoryConfig != nil {
		// Apply configuration to repoManager
	}

	// Create template loader
	templateLoader := loader.NewTemplateLoader(options.CacheTTL, repoManager)

	// Create template cache
	templateCache := cache.NewTemplateCache(options.CacheTTL, options.CacheMaxSize, cache.LRU)

	// Create template registry
	templateRegistry := registry.NewTemplateRegistry()

	// Create detection engine
	detectionEngine := execution.NewDefaultDetectionEngine()

	// Create template executor
	execOptions := &execution.ExecutionOptions{
		DetectionEngine: detectionEngine,
		Timeout:         options.ExecutionTimeout,
		RetryCount:      options.RetryCount,
		RetryDelay:      options.RetryDelay,
		MaxConcurrent:   options.MaxConcurrent,
		Variables:       make(map[string]interface{}),
		ProviderOptions: make(map[string]interface{}),
	}
	templateExecutor := execution.NewTemplateExecutor(execOptions)

	// Register providers
	for _, provider := range options.Providers {
		templateExecutor.RegisterProvider(provider)
	}

	// Create template reporter
	templateReporter, err := reporting.NewTemplateReporter()
	if err != nil {
		return nil, fmt.Errorf("failed to create template reporter: %w", err)
	}

	// Create adapter for cache and registry
	cacheAdapter := cache.NewTemplateCacheAdapter(templateCache)
	registryAdapter := registry.NewTemplateRegistryAdapter(templateRegistry)

	// Create manager options
	managerOptions := &ManagerOptions{
		Loaders:           []interfaces.TemplateLoader{templateLoader},
		Parser:            templateParser,
		Executor:          templateExecutor,
		Reporter:          templateReporter,
		Cache:             cacheAdapter,
		Registry:          registryAdapter,
		PreExecutionHooks: options.PreExecutionHooks,
		PostExecutionHooks: options.PostExecutionHooks,
	}

	// Create manager
	manager, err := NewManager(managerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Load templates if paths are provided
	if len(options.TemplatePaths) > 0 {
		for _, path := range options.TemplatePaths {
			templates, err := templateLoader.LoadFromPath(ctx, path, true)
			if err != nil {
				return nil, fmt.Errorf("failed to load templates from path %s: %w", path, err)
			}

			for _, template := range templates {
				if err := manager.RegisterTemplate(template); err != nil {
					return nil, fmt.Errorf("failed to register template %s: %w", template.ID, err)
				}
			}
		}
	}

	return manager, nil
}

// DefaultManagerOptions contains options for creating a default template manager
type DefaultManagerOptions struct {
	// JSONSchemaPath is the path to the JSON schema file
	JSONSchemaPath string
	// YAMLSchemaPath is the path to the YAML schema file
	YAMLSchemaPath string
	// TemplatePaths is the list of paths to load templates from
	TemplatePaths []string
	// RepositoryConfig is the configuration for the repository manager
	RepositoryConfig *repository.Config
	// CacheTTL is the time-to-live for cached templates
	CacheTTL time.Duration
	// CacheMaxSize is the maximum size of the cache
	CacheMaxSize int
	// ExecutionTimeout is the timeout for template execution
	ExecutionTimeout time.Duration
	// RetryCount is the number of retries for failed requests
	RetryCount int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	// MaxConcurrent is the maximum number of concurrent executions
	MaxConcurrent int
	// Providers is the list of LLM providers
	Providers []execution.LLMProvider
	// PreExecutionHooks are functions to run before template execution
	PreExecutionHooks []TemplateHook
	// PostExecutionHooks are functions to run after template execution
	PostExecutionHooks []TemplateHook
}

// DefaultManagerOptionsWithDefaults creates default manager options with default values
func DefaultManagerOptionsWithDefaults() *DefaultManagerOptions {
	return &DefaultManagerOptions{
		JSONSchemaPath:    "src/template/management/schemas/template.json",
		YAMLSchemaPath:    "src/template/management/schemas/template.yaml",
		TemplatePaths:     []string{"examples/templates"},
		RepositoryConfig:  &repository.Config{},
		CacheTTL:          1 * time.Hour,
		CacheMaxSize:      100,
		ExecutionTimeout:  30 * time.Second,
		RetryCount:        3,
		RetryDelay:        1 * time.Second,
		MaxConcurrent:     10,
		Providers:         []execution.LLMProvider{},
		PreExecutionHooks: []TemplateHook{},
		PostExecutionHooks: []TemplateHook{},
	}
}

// RunTemplate runs a template with the specified options
func RunTemplate(ctx context.Context, manager *Manager, templateID string, options map[string]interface{}) (*TemplateResult, error) {
	// Execute template
	result, err := manager.ExecuteTemplate(ctx, templateID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templateID, err)
	}

	return result, nil
}

// RunTemplates runs multiple templates with the specified options
func RunTemplates(ctx context.Context, manager *Manager, templateIDs []string, options map[string]interface{}) ([]*TemplateResult, error) {
	// Execute templates
	results, err := manager.ExecuteTemplates(ctx, templateIDs, options)
	if err != nil {
		return nil, fmt.Errorf("failed to execute templates: %w", err)
	}

	return results, nil
}

// GenerateTemplateReport generates a report for template execution results
func GenerateTemplateReport(manager *Manager, results []*TemplateResult, format string) ([]byte, error) {
	// Generate report
	report, err := manager.GenerateReport(results, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return report, nil
}

// ListAllTemplates lists all templates
func ListAllTemplates(manager *Manager) []*format.Template {
	return manager.ListTemplates()
}

// FindTemplatesByTag finds templates by tag
func FindTemplatesByTag(manager *Manager, tag string) []*format.Template {
	return manager.FindTemplatesByTag(tag)
}

// FindTemplatesByTags finds templates by multiple tags
func FindTemplatesByTags(manager *Manager, tags []string) []*format.Template {
	return manager.FindTemplatesByTags(tags)
}

// GetTemplateByID gets a template by ID
func GetTemplateByID(manager *Manager, id string) (*format.Template, error) {
	return manager.GetTemplate(id)
}
