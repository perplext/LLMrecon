// Package loader provides functionality for loading templates from various sources.
package loader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/cache"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/structure"
)

// ResourceEfficientLoader is a template loader optimized for minimal resource usage
type ResourceEfficientLoader struct {
	// cache is the template cache
	cache *cache.OptimizedTemplateCache
	// repoManager is the repository manager
	repoManager *repository.Manager
	// optimizer is the template optimizer
	optimizer *TemplateOptimizer
	// structureOptimizer is the template structure optimizer
	structureOptimizer *structure.TemplateStructureOptimizer
	// indexedSources tracks indexed sources for lazy loading
	indexedSources map[string]*SourceIndex
	// indexMutex protects the indexedSources map
	indexMutex sync.RWMutex
	// loadSemaphore limits concurrent loading operations
	loadSemaphore chan struct{}
	// stats tracks loader statistics
	stats LoaderStats
	// statsMutex protects the stats
	statsMutex sync.RWMutex
	// options contains loader configuration options
	options ResourceEfficientLoaderOptions
}

// ResourceEfficientLoaderOptions contains configuration options for the loader
type ResourceEfficientLoaderOptions struct {
	// CacheTTL is the time-to-live for cached templates
	CacheTTL time.Duration
	// MaxCacheSize is the maximum number of templates in the cache
	MaxCacheSize int
	// ConcurrencyLimit limits concurrent loading operations
	ConcurrencyLimit int
	// EnableOptimization enables template optimization
	EnableOptimization bool
	// EnableStructureOptimization enables template structure optimization
	EnableStructureOptimization bool
	// EnableCompression enables template compression
	EnableCompression bool
	// EnableMinification enables template minification
	EnableMinification bool
	// ChunkSize is the size of chunks for streaming operations
	ChunkSize int
	// MaxMemoryUsage is the maximum memory usage in bytes
	MaxMemoryUsage int64
}

// NewResourceEfficientLoader creates a new resource-efficient template loader
func NewResourceEfficientLoader(repoManager *repository.Manager, options ResourceEfficientLoaderOptions) *ResourceEfficientLoader {
	// Set default values
	if options.CacheTTL == 0 {
		options.CacheTTL = 1 * time.Hour
	}
	if options.MaxCacheSize <= 0 {
		options.MaxCacheSize = 1000
	}
	if options.ConcurrencyLimit <= 0 {
		options.ConcurrencyLimit = runtime.NumCPU()
	}
	if options.ChunkSize <= 0 {
		options.ChunkSize = 4096
	}
	if options.MaxMemoryUsage <= 0 {
		options.MaxMemoryUsage = 1 << 30 // 1GB default
	}

	return &ResourceEfficientLoader{
		cache:              cache.NewOptimizedTemplateCache(options.CacheTTL, options.MaxCacheSize),
		repoManager:        repoManager,
		optimizer:          NewTemplateOptimizer(options.EnableMinification, options.EnableCompression),
		structureOptimizer: structure.NewTemplateStructureOptimizer(),
		indexedSources:     make(map[string]*SourceIndex),
		loadSemaphore:      make(chan struct{}, options.ConcurrencyLimit),
		options:            options,
	}
}

// LoadTemplate loads a template from a source
func (l *ResourceEfficientLoader) LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error) {
	startTime := time.Now()
	l.statsMutex.Lock()
	l.stats.TotalLoads++
	l.statsMutex.Unlock()

	// Check if source is indexed
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	l.indexMutex.RLock()
	sourceIndex, indexed := l.indexedSources[sourceKey]
	l.indexMutex.RUnlock()

	// If source is not indexed, index it first
	if !indexed {
		if err := l.indexSource(ctx, source, sourceType); err != nil {
			l.recordLoadError(startTime)
			return nil, err
		}

		// Get the newly created index
		l.indexMutex.RLock()
		sourceIndex = l.indexedSources[sourceKey]
		l.indexMutex.RUnlock()
	}

	// If source has no templates, return error
	if len(sourceIndex.TemplateIDs) == 0 {
		l.recordLoadError(startTime)
		return nil, fmt.Errorf("no templates found in source %s", source)
	}

	// Load the first template
	template, err := l.loadTemplateByID(ctx, sourceIndex.TemplateIDs[0], sourceIndex)
	
	if err != nil {
		l.recordLoadError(startTime)
	} else {
		l.recordLoadSuccess(startTime)
	}
	
	return template, err
}

// LoadTemplates loads multiple templates from a source
func (l *ResourceEfficientLoader) LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error) {
	startTime := time.Now()
	
	// Check if source is indexed
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	l.indexMutex.RLock()
	sourceIndex, indexed := l.indexedSources[sourceKey]
	l.indexMutex.RUnlock()

	// If source is not indexed, index it first
	if !indexed {
		if err := l.indexSource(ctx, source, sourceType); err != nil {
			l.recordLoadError(startTime)
			return nil, err
		}

		// Get the newly created index
		l.indexMutex.RLock()
		sourceIndex = l.indexedSources[sourceKey]
		l.indexMutex.RUnlock()
	}

	// Load templates in batches to control memory usage
	batchSize := l.calculateBatchSize(len(sourceIndex.TemplateIDs))
	templates := make([]*format.Template, 0, len(sourceIndex.TemplateIDs))
	
	for i := 0; i < len(sourceIndex.TemplateIDs); i += batchSize {
		end := i + batchSize
		if end > len(sourceIndex.TemplateIDs) {
			end = len(sourceIndex.TemplateIDs)
		}
		
		batchIDs := sourceIndex.TemplateIDs[i:end]
		batchTemplates, err := l.loadTemplateBatch(ctx, batchIDs, sourceIndex)
		if err != nil {
			l.recordLoadError(startTime)
			return nil, err
		}
		
		templates = append(templates, batchTemplates...)
		
		// Allow garbage collection between batches
		runtime.GC()
	}
	
	l.recordLoadSuccess(startTime)
	return templates, nil
}

// loadTemplateBatch loads a batch of templates
func (l *ResourceEfficientLoader) loadTemplateBatch(ctx context.Context, templateIDs []string, sourceIndex *SourceIndex) ([]*format.Template, error) {
	templates := make([]*format.Template, 0, len(templateIDs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(templateIDs))

	for _, id := range templateIDs {
		wg.Add(1)
		go func(templateID string) {
			defer wg.Done()

			// Acquire semaphore
			l.loadSemaphore <- struct{}{}
			defer func() { <-l.loadSemaphore }()

			// Load template
			template, err := l.loadTemplateByID(ctx, templateID, sourceIndex)
			if err != nil {
				errorsChan <- err
				return
			}

			mu.Lock()
			templates = append(templates, template)
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	if len(errorsChan) > 0 {
		return nil, <-errorsChan
	}

	return templates, nil
}

// indexSource indexes a template source
func (l *ResourceEfficientLoader) indexSource(ctx context.Context, source string, sourceType string) error {
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	
	// Create a new source index
	index := &SourceIndex{
		Type:       sourceType,
		Path:       source,
		LastIndexed: time.Now(),
		TemplateIDs: make([]string, 0),
		FileMap:     make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}
	
	var err error
	
	switch interfaces.TemplateSource(sourceType) {
	case interfaces.FileSource:
		err = l.indexLocalPath(ctx, source, index)
	case interfaces.GitHubSource, interfaces.GitLabSource:
		err = l.indexRepository(ctx, source, sourceType, index)
	default:
		err = fmt.Errorf("unsupported source type: %s", sourceType)
	}
	
	if err != nil {
		return err
	}
	
	// Store the index
	l.indexMutex.Lock()
	l.indexedSources[sourceKey] = index
	l.indexMutex.Unlock()
	
	return nil
}

// indexLocalPath indexes a local path
func (l *ResourceEfficientLoader) indexLocalPath(ctx context.Context, path string, index *SourceIndex) error {
	// Check if path is a file or directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if fileInfo.IsDir() {
		// Index directory
		return l.indexDirectory(ctx, path, index)
	}

	// Index file
	return l.indexFile(ctx, path, index)
}

// indexDirectory indexes a directory
func (l *ResourceEfficientLoader) indexDirectory(ctx context.Context, dirPath string, index *SourceIndex) error {
	// Walk the directory
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-template files
		if !isTemplateFile(path) {
			return nil
		}

		// Index file
		return l.indexFile(ctx, path, index)
	})
}

// indexFile indexes a file
func (l *ResourceEfficientLoader) indexFile(ctx context.Context, filePath string, index *SourceIndex) error {
	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse template to get ID
	template, err := format.ParseTemplate(content)
	if err != nil {
		return fmt.Errorf("failed to parse template from file %s: %w", filePath, err)
	}

	// Add template ID to index
	index.TemplateIDs = append(index.TemplateIDs, template.ID)
	index.FileMap[template.ID] = filePath

	return nil
}

// indexRepository indexes a repository
func (l *ResourceEfficientLoader) indexRepository(ctx context.Context, repoURL string, repoType string, index *SourceIndex) error {
	// Create repository configuration
	var repoTypeEnum repository.RepositoryType
	if repoType == string(interfaces.GitHubSource) {
		repoTypeEnum = repository.GitHub
	} else if repoType == string(interfaces.GitLabSource) {
		repoTypeEnum = repository.GitLab
	} else {
		return fmt.Errorf("unsupported repository type: %s", repoType)
	}
	
	repoConfig := &repository.Config{
		Type: repoTypeEnum,
		URL:  repoURL,
		Name: fmt.Sprintf("%s-%s", repoType, repoURL),
	}
	
	// Create repository
	repo, err := l.repoManager.CreateRepository(repoConfig)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	
	// Connect to repository
	if err := repo.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to repository: %w", err)
	}

	// List template files
	fileInfos, err := repo.ListFiles(ctx, "**/*.{yaml,yml,json}")
	if err != nil {
		return fmt.Errorf("failed to list template files in repository %s: %w", repoURL, err)
	}
	
	// Convert FileInfo to file paths
	files := make([]string, 0, len(fileInfos))
	for _, fi := range fileInfos {
		files = append(files, fi.Path)
	}

	// Process files in batches to control memory usage
	batchSize := l.calculateBatchSize(len(files))
	
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}
		
		batchFiles := files[i:end]
		
		// Process batch
		var wg sync.WaitGroup
		var mu sync.Mutex
		errorsChan := make(chan error, len(batchFiles))
		
		for _, file := range batchFiles {
			wg.Add(1)
			go func(filePath string) {
				defer wg.Done()
				
				// Acquire semaphore
				l.loadSemaphore <- struct{}{}
				defer func() { <-l.loadSemaphore }()
				
				// Get file reader
				reader, err := repo.GetFile(ctx, filePath)
				if err != nil {
					errorsChan <- fmt.Errorf("failed to get file for %s: %w", filePath, err)
					return
				}
				defer reader.Close()
				
				// Read file content
				content, err := ioutil.ReadAll(reader)
				if err != nil {
					errorsChan <- fmt.Errorf("failed to read file content for %s: %w", filePath, err)
					return
				}
				
				// Parse template to get ID
				template, err := format.ParseTemplate(content)
				if err != nil {
					// Skip files that are not valid templates
					return
				}
				
				// Add template ID to index
				mu.Lock()
				index.TemplateIDs = append(index.TemplateIDs, template.ID)
				index.FileMap[template.ID] = filePath
				mu.Unlock()
			}(file)
		}
		
		wg.Wait()
		close(errorsChan)
		
		// Check for errors
		if len(errorsChan) > 0 {
			return <-errorsChan
		}
		
		// Allow garbage collection between batches
		runtime.GC()
	}
	
	return nil
}

// loadTemplateByID loads a template by ID from a source index
func (l *ResourceEfficientLoader) loadTemplateByID(ctx context.Context, id string, index *SourceIndex) (*format.Template, error) {
	// Check cache first
	if template, found := l.cache.Get(id); found {
		l.statsMutex.Lock()
		l.stats.CacheHits++
		l.statsMutex.Unlock()
		return template, nil
	}
	
	l.statsMutex.Lock()
	l.stats.CacheMisses++
	l.statsMutex.Unlock()
	
	// Get file path from index
	filePath, exists := index.FileMap[id]
	if !exists {
		return nil, fmt.Errorf("template ID %s not found in source index", id)
	}
	
	var template *format.Template
	var err error
	
	switch interfaces.TemplateSource(index.Type) {
	case interfaces.FileSource:
		// Load from local file
		template, err = l.loadFromLocalFile(ctx, filePath)
	case interfaces.GitHubSource, interfaces.GitLabSource:
		// Load from repository
		template, err = l.loadFromRepository(ctx, filePath, index.Type, index.Path)
	default:
		err = fmt.Errorf("unsupported source type: %s", index.Type)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Optimize template if enabled
	if l.options.EnableOptimization {
		template, err = l.optimizer.OptimizeTemplate(template)
		if err != nil {
			return nil, fmt.Errorf("failed to optimize template: %w", err)
		}
	}
	
	// Optimize template structure if enabled
	if l.options.EnableStructureOptimization {
		template, err = l.structureOptimizer.OptimizeTemplate(template)
		if err != nil {
			return nil, fmt.Errorf("failed to optimize template structure: %w", err)
		}
	}
	
	// Cache the template
	l.cache.Set(id, template)
	
	return template, nil
}

// loadFromLocalFile loads a template from a local file
func (l *ResourceEfficientLoader) loadFromLocalFile(ctx context.Context, filePath string) (*format.Template, error) {
	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	// Parse template
	template, err := format.ParseTemplate(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template from file %s: %w", filePath, err)
	}
	
	return template, nil
}

// loadFromRepository loads a template from a repository
func (l *ResourceEfficientLoader) loadFromRepository(ctx context.Context, filePath string, repoType string, repoURL string) (*format.Template, error) {
	// Create repository configuration
	var repoTypeEnum repository.RepositoryType
	if repoType == string(interfaces.GitHubSource) {
		repoTypeEnum = repository.GitHub
	} else if repoType == string(interfaces.GitLabSource) {
		repoTypeEnum = repository.GitLab
	} else {
		return nil, fmt.Errorf("unsupported repository type: %s", repoType)
	}
	
	repoConfig := &repository.Config{
		Type: repoTypeEnum,
		URL:  repoURL,
		Name: fmt.Sprintf("%s-%s", repoType, repoURL),
	}
	
	// Get or create repository
	repo, err := l.repoManager.GetRepository(repoConfig.Name)
	if err != nil {
		// Repository not found, create it
		repo, err = l.repoManager.CreateRepository(repoConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create repository: %w", err)
		}
		
		// Connect to repository
		if err := repo.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to repository: %w", err)
		}
	}
	
	// Get file reader
	reader, err := repo.GetFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file for %s: %w", filePath, err)
	}
	defer reader.Close()
	
	// Read file content
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content for %s: %w", filePath, err)
	}
	
	// Parse template
	template, err := format.ParseTemplate(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template from file %s: %w", filePath, err)
	}
	
	return template, nil
}

// calculateBatchSize calculates the batch size based on the total number of items
func (l *ResourceEfficientLoader) calculateBatchSize(totalItems int) int {
	// Default batch size
	batchSize := 10
	
	// Adjust based on concurrency limit
	if l.options.ConcurrencyLimit > 0 {
		batchSize = l.options.ConcurrencyLimit * 2
	}
	
	// Ensure batch size is not larger than total items
	if batchSize > totalItems {
		batchSize = totalItems
	}
	
	return batchSize
}

// recordLoadSuccess records a successful load operation
func (l *ResourceEfficientLoader) recordLoadSuccess(startTime time.Time) {
	l.statsMutex.Lock()
	defer l.statsMutex.Unlock()
	
	l.stats.TotalLoadTime += time.Since(startTime)
}

// recordLoadError records a failed load operation
func (l *ResourceEfficientLoader) recordLoadError(startTime time.Time) {
	l.statsMutex.Lock()
	defer l.statsMutex.Unlock()
	
	l.stats.LoadErrors++
	l.stats.TotalLoadTime += time.Since(startTime)
}

// GetLoaderStats returns statistics about the loader
func (l *ResourceEfficientLoader) GetLoaderStats() map[string]interface{} {
	l.statsMutex.RLock()
	defer l.statsMutex.RUnlock()
	
	cacheStats := l.cache.GetStats()
	optimizerStats := l.optimizer.GetOptimizationStats()
	structureOptimizerStats := l.structureOptimizer.GetOptimizationStats()
	
	avgLoadTime := time.Duration(0)
	if l.stats.TotalLoads > 0 {
		avgLoadTime = time.Duration(int64(l.stats.TotalLoadTime) / l.stats.TotalLoads)
	}
	
	return map[string]interface{}{
		"total_loads":      l.stats.TotalLoads,
		"cache_hits":       l.stats.CacheHits,
		"cache_misses":     l.stats.CacheMisses,
		"load_errors":      l.stats.LoadErrors,
		"total_load_time":  l.stats.TotalLoadTime,
		"avg_load_time":    avgLoadTime,
		"cache_stats":      cacheStats,
		"optimizer_stats":  optimizerStats,
		"structure_stats":  structureOptimizerStats,
		"indexed_sources":  len(l.indexedSources),
	}
}

// ClearCache clears the template cache
func (l *ResourceEfficientLoader) ClearCache() {
	l.cache.Clear()
}

// ClearSourceIndex clears the source index for a specific source
func (l *ResourceEfficientLoader) ClearSourceIndex(source string, sourceType string) {
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	
	l.indexMutex.Lock()
	delete(l.indexedSources, sourceKey)
	l.indexMutex.Unlock()
}

// ClearAllSourceIndices clears all source indices
func (l *ResourceEfficientLoader) ClearAllSourceIndices() {
	l.indexMutex.Lock()
	l.indexedSources = make(map[string]*SourceIndex)
	l.indexMutex.Unlock()
}

// SetConcurrencyLimit sets the concurrency limit for loading operations
func (l *ResourceEfficientLoader) SetConcurrencyLimit(limit int) {
	if limit <= 0 {
		limit = runtime.NumCPU()
	}
	
	// Create a new semaphore with the new limit
	l.loadSemaphore = make(chan struct{}, limit)
	l.options.ConcurrencyLimit = limit
}
