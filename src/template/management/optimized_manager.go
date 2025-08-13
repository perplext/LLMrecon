// Package management provides functionality for managing templates.
package management

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/loader"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

// ManagerConfig represents configuration for the template manager
type ManagerConfig struct {
	// CacheTTL is the TTL for cached templates
	CacheTTL time.Duration
	// MaxCacheSize is the maximum number of cached templates
	MaxCacheSize int
	// ConcurrencyLimit is the maximum number of concurrent operations
	ConcurrencyLimit int
	// ExecutionTimeout is the timeout for template execution
	ExecutionTimeout time.Duration
	// LoadTimeout is the timeout for template loading
	LoadTimeout time.Duration
	// RetryCount is the number of retries for failed operations
	RetryCount int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	// Debug enables debug logging
	Debug bool
}

// OptimizedTemplateManager is an enhanced template manager with improved performance
type OptimizedTemplateManager struct {
	// loader is the template loader
	loader *loader.OptimizedTemplateLoader
	// executor is the template executor
	executor *execution.OptimizedTemplateExecutor
	// repoManager is the repository manager
	repoManager *repository.Manager
	// config is the template manager configuration
	config *ManagerConfig
	// templateIndex is a map of template ID to source information
	templateIndex map[string]*TemplateIndexEntry
	// indexMutex protects the templateIndex
	indexMutex sync.RWMutex
	// stats tracks manager statistics
	stats ManagerStats
	// statsMutex protects the stats
	statsMutex sync.RWMutex
}

// TemplateIndexEntry contains information about a template source
type TemplateIndexEntry struct {
	// Source is the source of the template
	Source string
	// SourceType is the type of the source
	SourceType string
	// LastAccessed is the time the template was last accessed
	LastAccessed time.Time
	// AccessCount is the number of times the template has been accessed
	AccessCount int
}

// ManagerStats tracks manager statistics
type ManagerStats struct {
	// TotalTemplates is the total number of templates managed
	TotalTemplates int
	// TotalSources is the total number of sources managed
	TotalSources int
	// TotalLoads is the total number of template loads
	TotalLoads int64
	// TotalExecutions is the total number of template executions
	TotalExecutions int64
	// CacheHitRate is the cache hit rate
	CacheHitRate float64
}

// NewOptimizedTemplateManager creates a new optimized template manager
func NewOptimizedTemplateManager(config *ManagerConfig) (*OptimizedTemplateManager, error) {
	// Create repository manager
	repoManager := repository.NewManager()

	// Set default values for config
	if config.CacheTTL == 0 {
		config.CacheTTL = 1 * time.Hour
	}
	if config.MaxCacheSize == 0 {
		config.MaxCacheSize = 1000
	}
	if config.ConcurrencyLimit == 0 {
		config.ConcurrencyLimit = 10
	}
	if config.ExecutionTimeout == 0 {
		config.ExecutionTimeout = 30 * time.Second
	}

	// Create optimized template loader
	optimizedLoader := loader.NewOptimizedTemplateLoader(
		config.CacheTTL,
		config.MaxCacheSize,
		repoManager,
		config.ConcurrencyLimit,
	)

	// Create execution options
	executionOptions := &execution.ExecutionOptions{
		Timeout:         config.ExecutionTimeout,
		RetryCount:      config.RetryCount,
		RetryDelay:      config.RetryDelay,
		MaxConcurrent:   config.ConcurrencyLimit,
		Variables:       make(map[string]interface{}),
		ProviderOptions: make(map[string]interface{}),
	}

	// Create optimized template executor
	optimizedExecutor := execution.NewOptimizedTemplateExecutor(
		executionOptions,
		config.MaxCacheSize,
		config.CacheTTL,
		config.ConcurrencyLimit,
	)

	return &OptimizedTemplateManager{
		loader:        optimizedLoader,
		executor:      optimizedExecutor,
		repoManager:   repoManager,
		config:        config,
		templateIndex: make(map[string]*TemplateIndexEntry),
	}, nil
}

// LoadTemplate loads a template from a source
func (m *OptimizedTemplateManager) LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error) {
	startTime := time.Now()
	m.statsMutex.Lock()
	m.stats.TotalLoads++
	m.statsMutex.Unlock()

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, m.config.LoadTimeout)
	defer cancel()

	// Load template
	template, err := m.loader.LoadTemplateWithTimeout(ctxWithTimeout, source, sourceType, m.config.LoadTimeout)
	if err != nil {
		return nil, err
	}

	// Update template index
	m.updateTemplateIndex(template.ID, source, sourceType)

	// Update stats
	m.statsMutex.Lock()
	m.stats.TotalTemplates = len(m.templateIndex)
	m.statsMutex.Unlock()

	// Log performance metrics if debug is enabled
	if m.config.Debug {
		fmt.Printf("Template load time: %v\n", time.Since(startTime))
	}

	return template, nil
}

// LoadTemplates loads multiple templates from a source
func (m *OptimizedTemplateManager) LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error) {
	startTime := time.Now()
	m.statsMutex.Lock()
	m.stats.TotalLoads++
	m.statsMutex.Unlock()

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, m.config.LoadTimeout)
	defer cancel()

	// Load templates
	templates, err := m.loader.LoadTemplatesWithTimeout(ctxWithTimeout, source, sourceType, m.config.LoadTimeout)
	if err != nil {
		return nil, err
	}

	// Update template index
	for _, template := range templates {
		m.updateTemplateIndex(template.ID, source, sourceType)
	}

	// Update stats
	m.statsMutex.Lock()
	m.stats.TotalTemplates = len(m.templateIndex)
	m.statsMutex.Unlock()

	// Log performance metrics if debug is enabled
	if m.config.Debug {
		fmt.Printf("Templates load time: %v for %d templates\n", time.Since(startTime), len(templates))
	}

	return templates, nil
}

// Execute executes a template
func (m *OptimizedTemplateManager) Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*interfaces.TemplateResult, error) {
	startTime := time.Now()
	m.statsMutex.Lock()
	m.stats.TotalExecutions++
	m.statsMutex.Unlock()

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, m.config.ExecutionTimeout)
	defer cancel()

	// Execute template
	result, err := m.executor.Execute(ctxWithTimeout, template, options)
	if err != nil {
		return nil, err
	}

	// Update template index
	m.updateTemplateAccessTime(template.ID)

	// Log performance metrics if debug is enabled
	if m.config.Debug {
		fmt.Printf("Template execution time: %v\n", time.Since(startTime))
	}

	return result, nil
}

// ExecuteBatch executes multiple templates
func (m *OptimizedTemplateManager) ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*interfaces.TemplateResult, error) {
	startTime := time.Now()
	m.statsMutex.Lock()
	m.stats.TotalExecutions += int64(len(templates))
	m.statsMutex.Unlock()

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, m.config.ExecutionTimeout)
	defer cancel()

	// Execute templates
	results, err := m.executor.ExecuteBatch(ctxWithTimeout, templates, options)
	if err != nil {
		return nil, err
	}

	// Update template index
	for _, template := range templates {
		m.updateTemplateAccessTime(template.ID)
	}

	// Log performance metrics if debug is enabled
	if m.config.Debug {
		fmt.Printf("Batch execution time: %v for %d templates\n", time.Since(startTime), len(templates))
	}

	return results, nil
}

// GetLoader returns the template loader
func (m *OptimizedTemplateManager) GetLoader() types.TemplateLoader {
	return m.loader
}

// GetExecutor returns the template executor
func (m *OptimizedTemplateManager) GetExecutor() interfaces.TemplateExecutor {
	return m.executor
}

// RegisterProvider registers an LLM provider
func (m *OptimizedTemplateManager) RegisterProvider(provider execution.LLMProvider) {
	m.executor.RegisterProvider(provider)
}

// RegisterDetectionEngine registers a detection engine
func (m *OptimizedTemplateManager) RegisterDetectionEngine(engine interfaces.DetectionEngine) {
	m.executor.RegisterDetectionEngine(engine)
}

// AddRepository adds a repository to the manager
func (m *OptimizedTemplateManager) AddRepository(config *repository.Config) error {
	_, err := m.repoManager.CreateRepository(config)
	return err
}

// GetRepository gets a repository by name
func (m *OptimizedTemplateManager) GetRepository(name string) (repository.Repository, error) {
	return m.repoManager.GetRepository(name)
}

// GetRepositories gets all repositories
func (m *OptimizedTemplateManager) GetRepositories() []repository.Repository {
	return m.repoManager.ListRepositories()
}

// RemoveRepository removes a repository by name
func (m *OptimizedTemplateManager) RemoveRepository(name string) error {
	return m.repoManager.RemoveRepository(name)
}

// ClearCache clears the template cache
func (m *OptimizedTemplateManager) ClearCache() {
	m.loader.ClearCache()
	m.executor.ClearCache()
}

// GetStats returns statistics about the manager
func (m *OptimizedTemplateManager) GetStats() map[string]interface{} {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	// Get loader stats
	loaderStats := m.loader.GetLoaderStats()
	
	// Get cache stats
	loaderCacheStats := m.loader.GetCacheStats()
	executorCacheStats := m.executor.GetCacheStats()
	
	// Get executor stats
	executorStats := m.executor.GetExecutionStats()

	// Calculate overall cache hit rate
	loaderHitRate := float64(0)
	if totalLoaderLookups := loaderCacheStats["hits"].(int64) + loaderCacheStats["misses"].(int64); totalLoaderLookups > 0 {
		loaderHitRate = float64(loaderCacheStats["hits"].(int64)) / float64(totalLoaderLookups) * 100
	}

	executorHitRate := float64(0)
	if totalExecutorLookups := executorCacheStats["hits"].(int64) + executorCacheStats["misses"].(int64); totalExecutorLookups > 0 {
		executorHitRate = float64(executorCacheStats["hits"].(int64)) / float64(totalExecutorLookups) * 100
	}

	overallHitRate := (loaderHitRate + executorHitRate) / 2

	return map[string]interface{}{
		"total_templates":     m.stats.TotalTemplates,
		"total_sources":       m.stats.TotalSources,
		"total_loads":         m.stats.TotalLoads,
		"total_executions":    m.stats.TotalExecutions,
		"cache_hit_rate":      overallHitRate,
		"loader_stats":        loaderStats,
		"loader_cache_stats":  loaderCacheStats,
		"executor_stats":      executorStats,
		"executor_cache_stats": executorCacheStats,
	}
}

// GetTemplateStats returns statistics about a specific template
func (m *OptimizedTemplateManager) GetTemplateStats(templateID string) map[string]interface{} {
	m.indexMutex.RLock()
	defer m.indexMutex.RUnlock()

	entry, exists := m.templateIndex[templateID]
	if !exists {
		return map[string]interface{}{
			"exists": false,
		}
	}

	return map[string]interface{}{
		"exists":        true,
		"source":        entry.Source,
		"source_type":   entry.SourceType,
		"last_accessed": entry.LastAccessed,
		"access_count":  entry.AccessCount,
	}
}

// GetTemplateIDs returns all template IDs
func (m *OptimizedTemplateManager) GetTemplateIDs() []string {
	m.indexMutex.RLock()
	defer m.indexMutex.RUnlock()

	ids := make([]string, 0, len(m.templateIndex))
	for id := range m.templateIndex {
		ids = append(ids, id)
	}

	return ids
}

// SetConcurrencyLimit sets the concurrency limit
func (m *OptimizedTemplateManager) SetConcurrencyLimit(limit int) {
	if limit <= 0 {
		return
	}

	m.config.ConcurrencyLimit = limit
	m.loader.SetConcurrencyLimit(limit)
	m.executor.SetMaxConcurrent(limit)
}

// SetCacheTTL sets the cache TTL
func (m *OptimizedTemplateManager) SetCacheTTL(ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	m.config.CacheTTL = ttl
	m.executor.SetCacheTTL(ttl)
}

// SetCacheSize sets the cache size
func (m *OptimizedTemplateManager) SetCacheSize(size int) {
	if size <= 0 {
		return
	}

	m.config.MaxCacheSize = size
	m.executor.SetCacheSize(size)
}

// SetExecutionTimeout sets the execution timeout
func (m *OptimizedTemplateManager) SetExecutionTimeout(timeout time.Duration) {
	if timeout <= 0 {
		return
	}

	m.config.ExecutionTimeout = timeout
}

// SetLoadTimeout sets the load timeout
func (m *OptimizedTemplateManager) SetLoadTimeout(timeout time.Duration) {
	if timeout <= 0 {
		return
	}

	m.config.LoadTimeout = timeout
}

// SetDebug sets the debug flag
func (m *OptimizedTemplateManager) SetDebug(debug bool) {
	m.config.Debug = debug
}

// updateTemplateIndex updates the template index
func (m *OptimizedTemplateManager) updateTemplateIndex(templateID string, source string, sourceType string) {
	m.indexMutex.Lock()
	defer m.indexMutex.Unlock()

	// Check if template already exists in index
	if entry, exists := m.templateIndex[templateID]; exists {
		// Update existing entry
		entry.LastAccessed = time.Now()
		entry.AccessCount++
	} else {
		// Create new entry
		m.templateIndex[templateID] = &TemplateIndexEntry{
			Source:       source,
			SourceType:   sourceType,
			LastAccessed: time.Now(),
			AccessCount:  1,
		}
	}

	// Update sources count
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	m.statsMutex.Lock()
	m.stats.TotalSources = len(sourceKey)
	m.statsMutex.Unlock()
}

// updateTemplateAccessTime updates the last access time for a template
func (m *OptimizedTemplateManager) updateTemplateAccessTime(templateID string) {
	m.indexMutex.Lock()
	defer m.indexMutex.Unlock()

	if entry, exists := m.templateIndex[templateID]; exists {
		entry.LastAccessed = time.Now()
		entry.AccessCount++
	}
}
