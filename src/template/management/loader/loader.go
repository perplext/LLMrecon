// Package loader provides functionality for loading templates from various sources.
package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// TemplateLoader is responsible for loading templates from various sources
type TemplateLoader struct {
	// cacheTTL is the time-to-live for cached templates
	cacheTTL time.Duration
	// cache is a map of template ID to template and expiration time
	cache map[string]cacheEntry
	// cacheMutex is a mutex for the cache
	cacheMutex sync.RWMutex
	// repoManager is the repository manager for loading templates from repositories
	repoManager *repository.Manager

// cacheEntry represents a cached template
type cacheEntry struct {
	// template is the cached template
	template *format.Template
	// expiration is the expiration time of the cache entry
	expiration time.Time

// NewTemplateLoader creates a new template loader
func NewTemplateLoader(cacheTTL time.Duration, repoManager *repository.Manager) *TemplateLoader {
	return &TemplateLoader{
		cacheTTL:    cacheTTL,
		cache:       make(map[string]cacheEntry),
		repoManager: repoManager,
	}

// LoadTemplateWithTimeout loads a template with a timeout
func (l *TemplateLoader) LoadTemplateWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) (*format.Template, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Call the regular LoadTemplate with the timeout context
	return l.LoadTemplate(ctxWithTimeout, source, sourceType)

// LoadTemplatesWithTimeout loads multiple templates with a timeout
func (l *TemplateLoader) LoadTemplatesWithTimeout(ctx context.Context, source string, sourceType string, timeout time.Duration) ([]*format.Template, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Call the regular LoadTemplates with the timeout context
	return l.LoadTemplates(ctxWithTimeout, source, sourceType)

// Load loads a template from a file
func (l *TemplateLoader) Load(filePath string) (*format.Template, error) {
	return l.loadFromFile(context.Background(), filePath)

// LoadFromBytes loads a template from bytes
func (l *TemplateLoader) LoadFromBytes(data []byte, formatType string) (*format.Template, error) {
	return format.ParseTemplate(data)

// LoadBatch loads multiple templates from a directory
func (l *TemplateLoader) LoadBatch(directory string) ([]*format.Template, error) {
	return l.loadFromDirectory(context.Background(), directory)

// LoadTemplate loads a template from a source
func (l *TemplateLoader) LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error) {
	templates, err := l.LoadTemplates(ctx, source, sourceType)
	if err != nil {
		return nil, err
	}
	
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found")
	}
	
	return templates[0], nil

// LoadTemplates loads multiple templates from a source
func (l *TemplateLoader) LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error) {
	switch interfaces.TemplateSource(sourceType) {
	case interfaces.FileSource:
		// Check if path exists
		if _, err := os.Stat(source); os.IsNotExist(err) {
			return nil, fmt.Errorf("path %s does not exist", source)
		}

		return l.loadFromLocalPath(ctx, source)
	case interfaces.GitHubSource:
		// Get repository URL from options
		options := make(map[string]interface{})
		options["repo_url"] = source
		return l.LoadFromRepository(ctx, source, options)
	case interfaces.GitLabSource:
		// GitLab source not implemented yet
		return nil, fmt.Errorf("GitLab source not implemented yet")
	case interfaces.DatabaseSource:
		// Database source not implemented yet
		return nil, fmt.Errorf("database source not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}

// loadFromLocalPath loads templates from a local path
func (l *TemplateLoader) loadFromLocalPath(ctx context.Context, path string) ([]*format.Template, error) {
	// Check if path is a file or directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if fileInfo.IsDir() {
		// Load templates from directory
		return l.loadFromDirectory(ctx, path)
	}

	// Load template from file
	template, err := l.loadFromFile(ctx, path)
	if err != nil {
		return nil, err
	}

	return []*format.Template{template}, nil

// loadFromDirectory loads templates from a directory
func (l *TemplateLoader) loadFromDirectory(ctx context.Context, dirPath string) ([]*format.Template, error) {
	var templates []*format.Template

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
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Load template from file
		template, err := l.loadFromFile(ctx, path)
		if err != nil {
			return err
		}

		templates = append(templates, template)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	return templates, nil
	

// loadFromFile loads a template from a file
func (l *TemplateLoader) loadFromFile(ctx context.Context, filePath string) (*format.Template, error) {
	// Read file content
	content, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	// Parse template
	template, err := format.ParseTemplate(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template from file %s: %w", filePath, err)
	}

	return template, nil

// LoadFromPath loads templates from a specific path
func (l *TemplateLoader) LoadFromPath(ctx context.Context, path string, recursive bool) ([]*format.Template, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	var templates []*format.Template

	if info.IsDir() {
		// Load templates from directory
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}
		for _, file := range files {
			filePath := filepath.Join(path, file.Name())

			if file.IsDir() {
				if recursive {
					// Recursively load templates from subdirectory
					subTemplates, err := l.LoadFromPath(ctx, filePath, recursive)
					if err != nil {
						// Log error but continue with other files
						fmt.Printf("Error loading templates from %s: %v\n", filePath, err)
						continue
					}
					templates = append(templates, subTemplates...)
				}
			} else {
				// Check if file is a template
				if isTemplateFile(file.Name()) {
					// Load template from file
					template, err := l.loadTemplateFromFile(filePath)
					if err != nil {
						// Log error but continue with other files
						fmt.Printf("Error loading template from %s: %v\n", filePath, err)
						continue
					}
					templates = append(templates, template)
				}
			}
		}
	} else {
		// Load template from file
		if isTemplateFile(path) {
			template, err := l.loadTemplateFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to load template from file: %w", err)
			}
			templates = append(templates, template)
		} else {
			return nil, fmt.Errorf("not a template file: %s", path)
		}
	}

	return templates, nil
	

// LoadFromRepository loads templates from a remote repository
func (l *TemplateLoader) LoadFromRepository(ctx context.Context, repoURL string, options map[string]interface{}) ([]*format.Template, error) {
	// Check if repository manager is available
	if l.repoManager == nil {
		return nil, fmt.Errorf("repository manager not available")
	}

	// Get repository
	repo, err := l.repoManager.GetRepository(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Get templates directory from options
	templatesDir := "templates"
	if templatesDirOpt, ok := options["templates_dir"].(string); ok {
		templatesDir = templatesDirOpt
	}

	// Note: Repository implementation handles recursion internally
	// We keep the recursive option in the API for future compatibility
	// Get files from repository
	files, err := repo.ListFiles(ctx, templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in repository: %w", err)
	}
	var templates []*format.Template
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 10) // Limit concurrent goroutines

	for _, file := range files {
		if isTemplateFile(file.Path) {
			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore

			go func(fileInfo repository.FileInfo) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release semaphore

				// Get file content
				fileReader, err := repo.GetFile(ctx, fileInfo.Path)
				if err != nil {
					// Log error but continue with other files
					fmt.Printf("Error getting file content from repository: %v\n", err)
					return
				}
				defer func() { if err := fileReader.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

				// Read file content
				content, err := ioutil.ReadAll(fileReader)
				if err != nil {
					// Log error but continue with other files
					fmt.Printf("Error reading file content from repository: %v\n", err)
					return
				}

				// Parse template
				template, err := parseTemplateFromContent(fileInfo.Path, content)
				if err != nil {
					// Log error but continue with other files
					fmt.Printf("Error parsing template from repository file %s: %v\n", fileInfo.Path, err)
					return
				}

				// Add template to list
				mu.Lock()
				templates = append(templates, template)
				mu.Unlock()
			}(file)
		}
	}

	wg.Wait()
	return templates, nil
	

// loadTemplateFromFile loads a template from a file
func (l *TemplateLoader) loadTemplateFromFile(filePath string) (*format.Template, error) {
	// Check cache first
	l.cacheMutex.RLock()
	if entry, ok := l.cache[filePath]; ok && time.Now().Before(entry.expiration) {
		l.cacheMutex.RUnlock()
		return entry.template, nil
	}
	l.cacheMutex.RUnlock()

	// Load template from file
	template, err := format.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template from file: %w", err)
	}

	// Update cache
	l.cacheMutex.Lock()
	l.cache[filePath] = cacheEntry{
		template:   template,
		expiration: time.Now().Add(l.cacheTTL),
	}
	l.cacheMutex.Unlock()

	return template, nil

// parseTemplateFromContent parses a template from file content
func parseTemplateFromContent(filePath string, content []byte) (*format.Template, error) {
	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	var template format.Template

	if ext == ".json" {
		// Parse JSON
		if err := json.Unmarshal(content, &template); err != nil {
			return nil, fmt.Errorf("failed to parse JSON template: %w", err)
		}
	} else if ext == ".yaml" || ext == ".yml" {
		// Parse YAML
		if err := yaml.Unmarshal(content, &template); err != nil {
			return nil, fmt.Errorf("failed to parse YAML template: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return &template, nil

// isTemplateFile checks if a file is a template file
func isTemplateFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".json" || ext == ".yaml" || ext == ".yml"

// ClearCache clears the template cache
func (l *TemplateLoader) ClearCache() {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()
	l.cache = make(map[string]cacheEntry)

// GetCacheSize returns the number of templates in the cache
func (l *TemplateLoader) GetCacheSize() int {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()
	return len(l.cache)

// PruneCache removes expired entries from the cache
func (l *TemplateLoader) PruneCache() int {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	count := 0
	now := time.Now()

	for key, entry := range l.cache {
		if now.After(entry.expiration) {
			delete(l.cache, key)
			count++
		}
	}

