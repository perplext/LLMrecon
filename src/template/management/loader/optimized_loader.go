// Package loader provides functionality for loading templates from various sources.
package loader

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/cache"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// OptimizedTemplateLoader is an enhanced template loader with lazy loading and improved performance
type OptimizedTemplateLoader struct {
	// cache is the template cache
	cache *cache.OptimizedTemplateCache
	// repoManager is the repository manager for loading templates from repositories
	repoManager *repository.Manager
	// indexedSources tracks indexed sources for lazy loading
	indexedSources map[string]*SourceIndex
	// indexMutex protects the indexedSources map
	indexMutex sync.RWMutex
	// concurrencyLimit limits concurrent loading operations
	concurrencyLimit int
	// loadSemaphore is a channel for limiting concurrent loading operations
	loadSemaphore chan struct{}
	// stats tracks loader statistics
	stats LoaderStats
	// statsMutex protects the stats
	statsMutex sync.RWMutex

// SourceIndex contains metadata about a template source
type SourceIndex struct {
	// Type is the type of the source
	Type string
	// Path is the path to the source
	Path string
	// LastIndexed is the time the source was last indexed
	LastIndexed time.Time
	// TemplateIDs is a list of template IDs in the source
	TemplateIDs []string
	// FileMap maps template IDs to file paths
	FileMap map[string]string
	// Metadata contains additional metadata about the source
	Metadata map[string]interface{}

// LoaderStats tracks loader statistics
type LoaderStats struct {
	// TotalLoads is the total number of template loads
	TotalLoads int64
	// CacheHits is the number of cache hits
	CacheHits int64
	// CacheMisses is the number of cache misses
	CacheMisses int64
	// LoadErrors is the number of load errors
	LoadErrors int64
	// TotalLoadTime is the total time spent loading templates
	TotalLoadTime time.Duration

// NewOptimizedTemplateLoader creates a new optimized template loader
func NewOptimizedTemplateLoader(cacheTTL time.Duration, maxCacheSize int, repoManager *repository.Manager, concurrencyLimit int) *OptimizedTemplateLoader {
	// Set default values
	if concurrencyLimit <= 0 {
		concurrencyLimit = 10
	}

	return &OptimizedTemplateLoader{
		cache:            cache.NewOptimizedTemplateCache(cacheTTL, maxCacheSize),
		repoManager:      repoManager,
		indexedSources:   make(map[string]*SourceIndex),
		concurrencyLimit: concurrencyLimit,
		loadSemaphore:    make(chan struct{}, concurrencyLimit),
	}

// LoadTemplateWithTimeout loads a template with a timeout
func (l *OptimizedTemplateLoader) LoadTemplateWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) (*format.Template, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Call the regular LoadTemplate with the timeout context
	return l.LoadTemplate(ctxWithTimeout, source, sourceType)

// LoadTemplatesWithTimeout loads multiple templates with a timeout
func (l *OptimizedTemplateLoader) LoadTemplatesWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) ([]*format.Template, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Call the regular LoadTemplates with the timeout context
	return l.LoadTemplates(ctxWithTimeout, source, sourceType)

// LoadTemplate loads a template from a source
func (l *OptimizedTemplateLoader) LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error) {
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
			l.statsMutex.Lock()
			l.stats.LoadErrors++
			l.stats.TotalLoadTime += time.Since(startTime)
			l.statsMutex.Unlock()
			return nil, err
		}

		// Get the newly created index
		l.indexMutex.RLock()
		sourceIndex = l.indexedSources[sourceKey]
		l.indexMutex.RUnlock()
	}

	// If source has no templates, return error
	if len(sourceIndex.TemplateIDs) == 0 {
		l.statsMutex.Lock()
		l.stats.LoadErrors++
		l.stats.TotalLoadTime += time.Since(startTime)
		l.statsMutex.Unlock()
		return nil, fmt.Errorf("no templates found in source %s", source)
	}

	// Load the first template
	template, err := l.loadTemplateByID(ctx, sourceIndex.TemplateIDs[0], sourceIndex)
	
	l.statsMutex.Lock()
	if err != nil {
		l.stats.LoadErrors++
	}
	l.stats.TotalLoadTime += time.Since(startTime)
	l.statsMutex.Unlock()
	
	return template, err

// LoadTemplates loads multiple templates from a source
func (l *OptimizedTemplateLoader) LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error) {
	startTime := time.Now()
	
	// Check if source is indexed
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	l.indexMutex.RLock()
	sourceIndex, indexed := l.indexedSources[sourceKey]
	l.indexMutex.RUnlock()
	// If source is not indexed, index it first
	if !indexed {
		if err := l.indexSource(ctx, source, sourceType); err != nil {
			l.statsMutex.Lock()
			l.stats.LoadErrors++
			l.stats.TotalLoadTime += time.Since(startTime)
			l.statsMutex.Unlock()
			return nil, err
		}

		// Get the newly created index
		l.indexMutex.RLock()
		sourceIndex = l.indexedSources[sourceKey]
		l.indexMutex.RUnlock()
	}

	// Load templates concurrently
	templates := make([]*format.Template, 0, len(sourceIndex.TemplateIDs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(sourceIndex.TemplateIDs))

	for _, id := range sourceIndex.TemplateIDs {
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

			// Add template to result
			mu.Lock()
			templates = append(templates, template)
			mu.Unlock()
		}(id)
	}
	// Wait for all goroutines to finish
	wg.Wait()
	close(errorsChan)

	// Check for errors
	var lastError error
	for err := range errorsChan {
		lastError = err
		l.statsMutex.Lock()
		l.stats.LoadErrors++
		l.statsMutex.Unlock()
	}

	l.statsMutex.Lock()
	l.stats.TotalLoadTime += time.Since(startTime)
	l.statsMutex.Unlock()

	// If no templates were loaded and there was an error, return the error
	if len(templates) == 0 && lastError != nil {
		return nil, lastError
	}

	return templates, nil

// indexSource indexes a template source
func (l *OptimizedTemplateLoader) indexSource(ctx context.Context, source string, sourceType string) error {
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	
	// Create a new source index
	sourceIndex := &SourceIndex{
		Type:        sourceType,
		Path:        source,
		LastIndexed: time.Now(),
		TemplateIDs: make([]string, 0),
		FileMap:     make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}

	switch interfaces.TemplateSource(sourceType) {
	case interfaces.FileSource:
		// Check if path exists
		if _, err := os.Stat(source); os.IsNotExist(err) {
			return fmt.Errorf("path %s does not exist", source)
		}

		if err := l.indexLocalPath(ctx, source, sourceIndex); err != nil {
			return err
		}

	case interfaces.GitHubSource, interfaces.GitLabSource:
		// Index repository
		if err := l.indexRepository(ctx, source, sourceType, sourceIndex); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported source type: %s", sourceType)
	}

	// Store the index
	l.indexMutex.Lock()
	l.indexedSources[sourceKey] = sourceIndex
	l.indexMutex.Unlock()

	return nil

// indexLocalPath indexes a local path
func (l *OptimizedTemplateLoader) indexLocalPath(ctx context.Context, path string, index *SourceIndex) error {
	// Check if path is a file or directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if fileInfo.IsDir() {
		// Index directory
		return l.indexDirectory(ctx, path, index)
	}

	// Index single file
	return l.indexFile(ctx, path, index)

// indexDirectory indexes a directory
func (l *OptimizedTemplateLoader) indexDirectory(ctx context.Context, dirPath string, index *SourceIndex) error {
	// Walk the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
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

	return err

// indexFile indexes a file
func (l *OptimizedTemplateLoader) indexFile(ctx context.Context, filePath string, index *SourceIndex) error {
	// Read file content
	content, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil // Skip files that can't be read
	}

	// Parse template to get ID
	template, err := format.ParseTemplate(content)
	if err != nil {
		return nil // Skip files that can't be parsed
	}

	// Add template ID to index
	index.TemplateIDs = append(index.TemplateIDs, template.ID)
	index.FileMap[template.ID] = filePath

	return nil

// indexRepository indexes a repository
func (l *OptimizedTemplateLoader) indexRepository(ctx context.Context, repoURL string, repoType string, index *SourceIndex) error {
	// Get repository options
	options := make(map[string]interface{})
	options["repo_url"] = repoURL
	
	var repoConfig *repository.Config
	
	if repoType == string(interfaces.GitHubSource) {
		repoConfig = &repository.Config{
			Type: repository.GitHub,
			URL:  repoURL,
		}
	} else if repoType == string(interfaces.GitLabSource) {
		repoConfig = &repository.Config{
			Type: repository.GitLab,
			URL:  repoURL,
		}
	} else {
		return fmt.Errorf("unsupported repository type: %s", repoType)
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
	files, err := repo.ListFiles(ctx, "**/*.{yaml,yml,json}")
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Index files
	for _, file := range files {
		// Skip non-template files
		if !isTemplateFile(file.Path) {
			continue
		}

		// Get file content
		reader, err := repo.GetFile(ctx, file.Path)
		if err != nil {
			continue // Skip files that can't be read
		}

		// Read file content
		content, err := ioutil.ReadAll(reader)
		reader.Close()
		if err != nil {
			continue // Skip files that can't be read
		}

		// Parse template to get ID
		template, err := format.ParseTemplate(content)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		// Add template ID to index
		index.TemplateIDs = append(index.TemplateIDs, template.ID)
		index.FileMap[template.ID] = file.Path
	}

	// Add repository to index metadata
	index.Metadata["repository"] = repo.GetName()

	return nil

// loadTemplateByID loads a template by ID from a source index
func (l *OptimizedTemplateLoader) loadTemplateByID(ctx context.Context, id string, index *SourceIndex) (*format.Template, error) {
	// Check cache first
	template, found := l.cache.Get(id)
	if found {
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
		return nil, fmt.Errorf("template with ID %s not found in source", id)
	}

	var content []byte
	var err error

	switch interfaces.TemplateSource(index.Type) {
	case interfaces.FileSource:
		// Read file content from local file system
		content, err = ioutil.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

	case interfaces.GitHubSource, interfaces.GitLabSource:
		// Get repository name from metadata
		repoName, ok := index.Metadata["repository"].(string)
		if !ok {
			return nil, fmt.Errorf("repository name not found in source metadata")
		}

		// Get repository
		repo, err := l.repoManager.GetRepository(repoName)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository: %w", err)
		}

		// Get file content from repository
		reader, err := repo.GetFile(ctx, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get file from repository: %w", err)
		}

		// Read file content
		content, err = ioutil.ReadAll(reader)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported source type: %s", index.Type)
	}

	// Parse template
	template, err = format.ParseTemplate(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Add to cache
	l.cache.Set(id, template)

	return template, nil

// GetCacheStats returns statistics about the cache
func (l *OptimizedTemplateLoader) GetCacheStats() map[string]interface{} {
	return l.cache.GetStats()

// GetLoaderStats returns statistics about the loader
func (l *OptimizedTemplateLoader) GetLoaderStats() map[string]interface{} {
	l.statsMutex.RLock()
	defer l.statsMutex.RUnlock()

	avgLoadTime := time.Duration(0)
	if l.stats.TotalLoads > 0 {
		avgLoadTime = l.stats.TotalLoadTime / time.Duration(l.stats.TotalLoads)
	}

	cacheHitRate := float64(0)
	totalCacheOps := l.stats.CacheHits + l.stats.CacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(l.stats.CacheHits) / float64(totalCacheOps) * 100
	}

	return map[string]interface{}{
		"total_loads":     l.stats.TotalLoads,
		"cache_hits":      l.stats.CacheHits,
		"cache_misses":    l.stats.CacheMisses,
		"load_errors":     l.stats.LoadErrors,
		"total_load_time": l.stats.TotalLoadTime,
		"avg_load_time":   avgLoadTime,
		"cache_hit_rate":  cacheHitRate,
	}

// ClearCache clears the template cache
func (l *OptimizedTemplateLoader) ClearCache() {
	l.cache.Clear()

// ClearSourceIndex clears the source index for a specific source
func (l *OptimizedTemplateLoader) ClearSourceIndex(source string, sourceType string) {
	sourceKey := fmt.Sprintf("%s:%s", sourceType, source)
	
	l.indexMutex.Lock()
	delete(l.indexedSources, sourceKey)
	l.indexMutex.Unlock()

// ClearAllSourceIndices clears all source indices
func (l *OptimizedTemplateLoader) ClearAllSourceIndices() {
	l.indexMutex.Lock()
	l.indexedSources = make(map[string]*SourceIndex)
	l.indexMutex.Unlock()

// SetConcurrencyLimit sets the concurrency limit for loading operations
func (l *OptimizedTemplateLoader) SetConcurrencyLimit(limit int) {
	if limit <= 0 {
		return
	}

	l.concurrencyLimit = limit
	l.loadSemaphore = make(chan struct{}, limit)

