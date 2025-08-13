// Package management provides functionality for managing templates in the LLMreconing Tool.
package management

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// Manager is the implementation of the TemplateManager interface
type Manager struct {
	// loaders is the list of template loaders
	loaders []interfaces.TemplateLoader
	// parser is the template parser
	parser interfaces.TemplateParser
	// executor is the template executor
	executor interfaces.TemplateExecutor
	// reporter is the template reporter
	reporter interfaces.TemplateReporter
	// cache is the template cache
	cache interfaces.TemplateCache
	// registry is the template registry
	registry interfaces.TemplateRegistry
	// preExecutionHooks are functions to run before template execution
	preExecutionHooks []TemplateHook
	// postExecutionHooks are functions to run after template execution
	postExecutionHooks []TemplateHook
	// mutex is a mutex for concurrent operations
	mutex sync.RWMutex
}

// NewManager creates a new template manager
func NewManager(options *ManagerOptions) (*Manager, error) {
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

	return &Manager{
		loaders:           options.Loaders,
		parser:            options.Parser,
		executor:          options.Executor,
		reporter:          options.Reporter,
		cache:             options.Cache,
		registry:          options.Registry,
		preExecutionHooks: options.PreExecutionHooks,
		postExecutionHooks: options.PostExecutionHooks,
	}, nil
}

// ManagerOptions contains options for creating a new template manager
type ManagerOptions struct {
	// Loaders is the list of template loaders
	Loaders []interfaces.TemplateLoader
	// Parser is the template parser
	Parser interfaces.TemplateParser
	// Executor is the template executor
	Executor interfaces.TemplateExecutor
	// Reporter is the template reporter
	Reporter interfaces.TemplateReporter
	// Cache is the template cache
	Cache interfaces.TemplateCache
	// Registry is the template registry
	Registry interfaces.TemplateRegistry
	// PreExecutionHooks are functions to run before template execution
	PreExecutionHooks []TemplateHook
	// PostExecutionHooks are functions to run after template execution
	PostExecutionHooks []TemplateHook
}

// LoadTemplates loads templates from the specified sources
func (m *Manager) LoadTemplates(ctx context.Context, sources []TemplateSource) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, source := range sources {
		for _, loader := range m.loaders {
			// Convert string source to the appropriate source type
			sourceType := string(source)
			
			// Check if loader supports the extended interface
			var templates []*format.Template
			var err error
			
			if extLoader, ok := loader.(interfaces.TemplateLoaderExtended); ok {
				// Use the extended loader interface
				templates, err = extLoader.LoadTemplates(ctx, string(source), sourceType)
			} else {
				// Fall back to basic loader - load single file
				template, err := loader.Load(string(source))
				if err == nil {
					templates = []*format.Template{template}
				}
			}
			
			if err != nil {
				// Try next loader
				continue
			}

			for _, template := range templates {
				if err := m.parser.Validate(template); err != nil {
					return fmt.Errorf("failed to validate template %s: %w", template.ID, err)
				}

				if err := m.registry.Register(template); err != nil {
					return fmt.Errorf("failed to register template %s: %w", template.ID, err)
				}

				if m.cache != nil {
					m.cache.Set(template.ID, template)
				}
			}
		}
	}

	return nil
}

// GetTemplate gets a template by ID
func (m *Manager) GetTemplate(id string) (*format.Template, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check cache first if available
	if m.cache != nil {
		if template, ok := m.cache.Get(id); ok {
			return template, nil
		}
	}

	// Get from registry
	template, err := m.registry.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template %s: %w", id, err)
	}

	// Update cache
	if m.cache != nil {
		m.cache.Set(id, template)
	}

	return template, nil
}

// ListTemplates lists all templates
func (m *Manager) ListTemplates() []*format.Template {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.registry.List()
}

// ValidateTemplate validates a template
func (m *Manager) ValidateTemplate(template *format.Template) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.parser.Validate(template)
}

// ExecuteTemplate executes a template
func (m *Manager) ExecuteTemplate(ctx context.Context, templateID string, options map[string]interface{}) (*TemplateResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get template
	template, err := m.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template %s: %w", templateID, err)
	}

	// Create result
	result := &interfaces.TemplateResult{
		TemplateID: template.ID,
		Status:     string(interfaces.StatusLoaded),
		StartTime:  time.Now(),
	}

	// Run pre-execution hooks
	for _, hook := range m.preExecutionHooks {
		if err := hook(ctx, template, result); err != nil {
			result.Status = string(interfaces.StatusFailed)
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, fmt.Errorf("pre-execution hook failed: %w", err)
		}
	}

	// Execute template
	result.Status = string(interfaces.StatusExecuting)
	execResult, err := m.executor.Execute(ctx, template, options)
	if err != nil {
		result.Status = string(interfaces.StatusFailed)
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("failed to execute template %s: %w", templateID, err)
	}

	// Update result
	result = execResult
	result.Status = string(interfaces.StatusCompleted)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Run post-execution hooks
	for _, hook := range m.postExecutionHooks {
		if err := hook(ctx, template, result); err != nil {
			result.Status = string(interfaces.StatusFailed)
			result.Error = err
			return result, fmt.Errorf("post-execution hook failed: %w", err)
		}
	}

	return result, nil
}

// ExecuteTemplates executes multiple templates
func (m *Manager) ExecuteTemplates(ctx context.Context, templateIDs []string, options map[string]interface{}) ([]*TemplateResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var results []*TemplateResult
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(templateIDs))

	for _, id := range templateIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			result, err := m.ExecuteTemplate(ctx, id, options)
			if err != nil {
				errChan <- fmt.Errorf("failed to execute template %s: %w", id, err)
				return
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(id)
	}

	// Wait for all executions to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("some templates failed to execute: %v", errs)
	}

	return results, nil
}

// GenerateReport generates a report for template execution results
func (m *Manager) GenerateReport(results []*TemplateResult, format string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.reporter.GenerateReport(results, format)
}

// AddHook adds a hook to the template manager
func (m *Manager) AddHook(hook TemplateHook, pre bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if pre {
		m.preExecutionHooks = append(m.preExecutionHooks, hook)
	} else {
		m.postExecutionHooks = append(m.postExecutionHooks, hook)
	}
}

// RemoveHook removes a hook from the template manager
func (m *Manager) RemoveHook(hook TemplateHook, pre bool) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var hooks []TemplateHook
	if pre {
		hooks = m.preExecutionHooks
	} else {
		hooks = m.postExecutionHooks
	}

	for i, h := range hooks {
		if &h == &hook {
			if pre {
				m.preExecutionHooks = append(m.preExecutionHooks[:i], m.preExecutionHooks[i+1:]...)
			} else {
				m.postExecutionHooks = append(m.postExecutionHooks[:i], m.postExecutionHooks[i+1:]...)
			}
			return true
		}
	}

	return false
}

// ClearHooks clears all hooks
func (m *Manager) ClearHooks() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.preExecutionHooks = nil
	m.postExecutionHooks = nil
}

// RegisterTemplate registers a template
func (m *Manager) RegisterTemplate(template *format.Template) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate template
	if err := m.parser.Validate(template); err != nil {
		return fmt.Errorf("failed to validate template: %w", err)
	}

	// Register template
	if err := m.registry.Register(template); err != nil {
		return fmt.Errorf("failed to register template: %w", err)
	}

	// Add to cache
	if m.cache != nil {
		m.cache.Set(template.ID, template)
	}

	return nil
}

// UnregisterTemplate unregisters a template
func (m *Manager) UnregisterTemplate(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Unregister template
	if err := m.registry.Unregister(id); err != nil {
		return fmt.Errorf("failed to unregister template: %w", err)
	}

	// Remove from cache
	if m.cache != nil {
		m.cache.Delete(id)
	}

	return nil
}

// UpdateTemplate updates a template
func (m *Manager) UpdateTemplate(template *format.Template) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate template
	if err := m.parser.Validate(template); err != nil {
		return fmt.Errorf("failed to validate template: %w", err)
	}

	// Update template
	if err := m.registry.Update(template); err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	// Update cache
	if m.cache != nil {
		m.cache.Set(template.ID, template)
	}

	return nil
}

// FindTemplatesByTag finds templates by tag
func (m *Manager) FindTemplatesByTag(tag string) []*format.Template {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.registry.FindByTag(tag)
}

// FindTemplatesByTags finds templates by multiple tags
func (m *Manager) FindTemplatesByTags(tags []string) []*format.Template {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.registry.FindByTags(tags)
}

// GetTemplateMetadata gets metadata for a template
func (m *Manager) GetTemplateMetadata(id string) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.registry.GetMetadata(id)
}

// SetTemplateMetadata sets metadata for a template
func (m *Manager) SetTemplateMetadata(id string, metadata map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.registry.SetMetadata(id, metadata)
}

// GetStats gets statistics about the template manager
func (m *Manager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := map[string]interface{}{
		"templates_count": m.registry.Count(),
	}

	if m.cache != nil {
		stats["cache"] = m.cache.GetStats()
	}

	return stats
}

// ClearCache clears the template cache
func (m *Manager) ClearCache() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.cache != nil {
		m.cache.Clear()
	}
}

// PruneCache removes old entries from the cache
func (m *Manager) PruneCache(maxAge time.Duration) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.cache != nil {
		return m.cache.Prune(maxAge)
	}
	return 0
}
