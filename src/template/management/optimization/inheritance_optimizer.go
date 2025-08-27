package optimization

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
)

// InheritanceOptimizer optimizes template inheritance by flattening inheritance chains
// and reducing the depth of template inheritance hierarchies
type InheritanceOptimizer struct {
	// maxInheritanceDepth is the maximum allowed inheritance depth
	maxInheritanceDepth int
	// flattenInheritance indicates if inheritance should be flattened
	flattenInheritance bool
	// cacheOptimizedTemplates indicates if optimized templates should be cached
	cacheOptimizedTemplates bool
	// templateCache caches optimized templates
	templateCache map[string]*format.Template
	// mutex protects the template cache
	mutex sync.RWMutex
}

// InheritanceOptimizerOptions contains options for the inheritance optimizer
type InheritanceOptimizerOptions struct {
	// MaxInheritanceDepth is the maximum allowed inheritance depth
	MaxInheritanceDepth int
	// FlattenInheritance indicates if inheritance should be flattened
	FlattenInheritance bool
	// CacheOptimizedTemplates indicates if optimized templates should be cached
	CacheOptimizedTemplates bool
}

// DefaultInheritanceOptimizerOptions returns default options for the inheritance optimizer
func DefaultInheritanceOptimizerOptions() *InheritanceOptimizerOptions {
	return &InheritanceOptimizerOptions{
		MaxInheritanceDepth:     3,
		FlattenInheritance:      true,
		CacheOptimizedTemplates: true,
	}
}

// NewInheritanceOptimizer creates a new inheritance optimizer
func NewInheritanceOptimizer(options *InheritanceOptimizerOptions) *InheritanceOptimizer {
	if options == nil {
		options = DefaultInheritanceOptimizerOptions()
	}

	return &InheritanceOptimizer{
		maxInheritanceDepth:     options.MaxInheritanceDepth,
		flattenInheritance:      options.FlattenInheritance,
		cacheOptimizedTemplates: options.CacheOptimizedTemplates,
		templateCache:           make(map[string]*format.Template),
	}
}

// loadTemplate loads a template from the cache
func (o *InheritanceOptimizer) loadTemplate(id string) (*format.Template, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	template, ok := o.templateCache[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	return template, nil
}

// OptimizeTemplate optimizes a template by flattening inheritance chains
// and reducing the depth of template inheritance hierarchies
func (o *InheritanceOptimizer) OptimizeTemplate(ctx context.Context, template *format.Template) (*format.Template, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	// Check if template is already optimized
	if o.cacheOptimizedTemplates {
		o.mutex.RLock()
		cachedTemplate, ok := o.templateCache[template.Name]
		o.mutex.RUnlock()
		if ok {
			return cachedTemplate, nil
		}
	}

	// Create a copy of the template
	optimizedTemplate := template.Clone()

	// Cache optimized template
	if o.cacheOptimizedTemplates {
		o.mutex.Lock()
		o.templateCache[template.Name] = optimizedTemplate
		o.mutex.Unlock()
	}

	return optimizedTemplate, nil
}

// flattenTemplateInheritance is a placeholder since the current Template struct doesn't support inheritance
func (o *InheritanceOptimizer) flattenTemplateInheritance(ctx context.Context, template *format.Template) error {
	// The current Template struct doesn't support inheritance patterns
	// This function is kept for compatibility but performs no operations
	return nil
}

// calculateInheritanceDepth returns 0 since the current Template struct doesn't support inheritance
func (o *InheritanceOptimizer) calculateInheritanceDepth(template *format.Template) int {
	// The current Template struct doesn't support inheritance patterns
	return 0
}

// mergeWithParents is a placeholder since the current Template struct doesn't support inheritance
func (o *InheritanceOptimizer) mergeWithParents(ctx context.Context, template *format.Template) error {
	// The current Template struct doesn't support inheritance patterns
	return nil
}

// mergeTemplates merges metadata from a parent template into a child template
func (o *InheritanceOptimizer) mergeTemplates(child, parent *format.Template) {
	// Merge content if child content is nil or empty
	if child.Content == nil || len(child.Content) == 0 {
		child.Content = make([]byte, len(parent.Content))
		copy(child.Content, parent.Content)
	}

	// Merge metadata
	if child.Metadata == nil {
		child.Metadata = make(map[string]interface{})
	}
	for key, value := range parent.Metadata {
		if _, ok := child.Metadata[key]; !ok {
			child.Metadata[key] = value
		}
	}
}

// OptimizeTemplates optimizes multiple templates
func (o *InheritanceOptimizer) OptimizeTemplates(ctx context.Context, templates []*format.Template) ([]*format.Template, error) {
	optimizedTemplates := make([]*format.Template, len(templates))
	var wg sync.WaitGroup
	errChan := make(chan error, len(templates))
	for i, template := range templates {
		wg.Add(1)
		go func(i int, template *format.Template) {
			defer wg.Done()
			optimizedTemplate, err := o.OptimizeTemplate(ctx, template)
			if err != nil {
				errChan <- err
				return
			}
			optimizedTemplates[i] = optimizedTemplate
		}(i, template)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return optimizedTemplates, nil
}

// ClearCache clears the template cache
func (o *InheritanceOptimizer) ClearCache() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.templateCache = make(map[string]*format.Template)
}

// GetCacheSize returns the size of the template cache
func (o *InheritanceOptimizer) GetCacheSize() int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return len(o.templateCache)
}

// SetMaxInheritanceDepth sets the maximum allowed inheritance depth
func (o *InheritanceOptimizer) SetMaxInheritanceDepth(depth int) {
	o.maxInheritanceDepth = depth
}

// GetMaxInheritanceDepth returns the maximum allowed inheritance depth
func (o *InheritanceOptimizer) GetMaxInheritanceDepth() int {
	return o.maxInheritanceDepth
}

// SetFlattenInheritance sets whether inheritance should be flattened
func (o *InheritanceOptimizer) SetFlattenInheritance(flatten bool) {
	o.flattenInheritance = flatten
}

// GetFlattenInheritance returns whether inheritance is flattened
func (o *InheritanceOptimizer) GetFlattenInheritance() bool {
	return o.flattenInheritance
}

// SetCacheOptimizedTemplates sets whether optimized templates should be cached
func (o *InheritanceOptimizer) SetCacheOptimizedTemplates(cache bool) {
	o.cacheOptimizedTemplates = cache
}

// GetCacheOptimizedTemplates returns whether optimized templates are cached
func (o *InheritanceOptimizer) GetCacheOptimizedTemplates() bool {
	return o.cacheOptimizedTemplates
}
