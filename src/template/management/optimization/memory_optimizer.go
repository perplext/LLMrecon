package optimization

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/utils/profiling"
)

// MemoryOptimizer optimizes memory usage in template processing
type MemoryOptimizer struct {
	// config contains the optimizer configuration
	config *MemoryOptimizerConfig
	// profiler is the memory profiler
	profiler *profiling.MemoryProfiler
	// mutex protects the optimizer state
	mutex sync.RWMutex
	// stats tracks optimization statistics
	stats *MemoryOptimizerStats
	// templateSizes tracks template sizes
	templateSizes map[string]int
	// optimizedSizes tracks optimized template sizes
	optimizedSizes map[string]int
	// lastOptimization is the time of the last optimization
	lastOptimization time.Time
	// optimizationCount is the number of optimization operations
	optimizationCount int64

// MemoryOptimizerConfig represents configuration for the memory optimizer
type MemoryOptimizerConfig struct {
	// EnableTemplateDeduplication enables template deduplication
	EnableTemplateDeduplication bool
	// EnableSectionDeduplication enables section deduplication
	EnableSectionDeduplication bool
	// EnableVariableOptimization enables variable optimization
	EnableVariableOptimization bool
	// EnableContentCompression enables content compression
	EnableContentCompression bool
	// EnableLazyLoading enables lazy loading of templates
	EnableLazyLoading bool
	// EnableInheritanceFlattening enables inheritance flattening
	EnableInheritanceFlattening bool
	// MaxInheritanceDepth is the maximum inheritance depth
	MaxInheritanceDepth int
	// EnableGarbageCollectionHints enables garbage collection hints
	EnableGarbageCollectionHints bool
	// GCInterval is the interval between garbage collections
	GCInterval time.Duration
	// MemoryThreshold is the memory threshold for optimization (in MB)
	MemoryThreshold int64
	// EnablePooling enables object pooling
	EnablePooling bool
	// PoolSize is the size of the object pool
	PoolSize int
	// EnableBufferReuse enables buffer reuse
	EnableBufferReuse bool
	// BufferSize is the size of reused buffers
	BufferSize int
	// EnableStringInterning enables string interning
	EnableStringInterning bool

// MemoryOptimizerStats tracks statistics for the memory optimizer
type MemoryOptimizerStats struct {
	// TemplatesOptimized is the number of templates optimized
	TemplatesOptimized int64
	// SectionsOptimized is the number of sections optimized
	SectionsOptimized int64
	// VariablesOptimized is the number of variables optimized
	VariablesOptimized int64
	// BytesSaved is the number of bytes saved
	BytesSaved int64
	// MemoryReduced is the amount of memory reduced (in MB)
	MemoryReduced float64
	// OptimizationTime is the total time spent on optimization
	OptimizationTime time.Duration
	// GarbageCollections is the number of garbage collections triggered
	GarbageCollections int64
	// DuplicatesRemoved is the number of duplicates removed
	DuplicatesRemoved int64
	// StringsInterned is the number of strings interned
	StringsInterned int64
	// BuffersReused is the number of buffers reused
	BuffersReused int64
	// ObjectsPooled is the number of objects pooled
	ObjectsPooled int64

// DefaultMemoryOptimizerConfig returns default configuration for the memory optimizer
func DefaultMemoryOptimizerConfig() *MemoryOptimizerConfig {
	return &MemoryOptimizerConfig{
		EnableTemplateDeduplication: true,
		EnableSectionDeduplication:  true,
		EnableVariableOptimization:  true,
		EnableContentCompression:    true,
		EnableLazyLoading:           true,
		EnableInheritanceFlattening: true,
		MaxInheritanceDepth:         3,
		EnableGarbageCollectionHints: true,
		GCInterval:                  5 * time.Minute,
		MemoryThreshold:             100, // 100 MB
		EnablePooling:               true,
		PoolSize:                    1000,
		EnableBufferReuse:           true,
		BufferSize:                  4096, // 4 KB
		EnableStringInterning:       true,
	}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(config *MemoryOptimizerConfig) (*MemoryOptimizer, error) {
	if config == nil {
		config = DefaultMemoryOptimizerConfig()
	}

	// Create memory profiler
	profilerOptions := profiling.DefaultProfilerOptions()
	profiler, err := profiling.NewMemoryProfiler(profilerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory profiler: %w", err)
	}

	return &MemoryOptimizer{
		config:         config,
		profiler:       profiler,
		stats:          &MemoryOptimizerStats{},
		templateSizes:  make(map[string]int),
		optimizedSizes: make(map[string]int),
	}, nil

// OptimizeTemplate optimizes a template for memory usage
func (o *MemoryOptimizer) OptimizeTemplate(template *format.Template) (*format.Template, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	// Take a snapshot of memory before optimization
	o.profiler.CreateSnapshot("before")

	// Start optimization timer
	startTime := time.Now()

	// Create a copy of the template
	optimized := template.Clone()

	// Calculate original size
	originalSize := estimateTemplateSize(template)
	o.mutex.Lock()
	o.templateSizes[template.ID] = originalSize
	o.mutex.Unlock()

	// Optimize template
	var err error
	if o.config.EnableTemplateDeduplication {
		optimized, err = o.deduplicateTemplate(optimized)
		if err != nil {
			return nil, fmt.Errorf("failed to deduplicate template: %w", err)
		}
	}

	if o.config.EnableSectionDeduplication && optimized.Content != nil {
		// Parse content to work with sections
		parsedTemplate, err := format.ParseTemplate(optimized.Content)
		if err == nil && parsedTemplate != nil {
			// Store the optimized content back
			optimized = parsedTemplate
		}
	}

	if o.config.EnableVariableOptimization && optimized.Variables != nil {
		optimized.Variables, err = o.optimizeVariables(optimized.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to optimize variables: %w", err)
		}
	}

	if o.config.EnableInheritanceFlattening {
		optimized, err = o.flattenInheritance(optimized)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten inheritance: %w", err)
		}
	}

	// Calculate optimized size
	optimizedSize := estimateTemplateSize(optimized)
	o.mutex.Lock()
	o.optimizedSizes[template.ID] = optimizedSize
	o.mutex.Unlock()

	// Update statistics
	atomic.AddInt64(&o.stats.TemplatesOptimized, 1)
	atomic.AddInt64(&o.stats.BytesSaved, int64(originalSize-optimizedSize))
	o.stats.OptimizationTime += time.Since(startTime)
	atomic.AddInt64(&o.optimizationCount, 1)
	o.lastOptimization = time.Now()

	// Take a snapshot of memory after optimization
	o.profiler.CreateSnapshot("after")

	// Compare memory snapshots
	diff, err := o.profiler.CompareSnapshots("before", "after")
	if err == nil && diff != nil {
		if heapDiff, ok := diff["heap_alloc_diff_mb"].(float64); ok && heapDiff < 0 {
			o.stats.MemoryReduced -= heapDiff // Convert negative diff to positive reduction
		}
	}

	// Trigger garbage collection if needed
	if o.config.EnableGarbageCollectionHints && o.shouldTriggerGC() {
		o.triggerGC()
	}

	return optimized, nil

// OptimizeTemplates optimizes multiple templates for memory usage
func (o *MemoryOptimizer) OptimizeTemplates(templates []*format.Template) ([]*format.Template, error) {
	if len(templates) == 0 {
		return nil, fmt.Errorf("templates is empty")
	}

	// Take a snapshot of memory before optimization
	o.profiler.CreateSnapshot("before_batch")

	// Start optimization timer
	startTime := time.Now()

	// Create result slice
	result := make([]*format.Template, len(templates))

	// Optimize templates
	var totalOriginalSize, totalOptimizedSize int
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	var firstErr error

	for i, template := range templates {
		wg.Add(1)
		go func(i int, template *format.Template) {
			defer wg.Done()

			// Calculate original size
			originalSize := estimateTemplateSize(template)
			o.mutex.Lock()
			o.templateSizes[template.ID] = originalSize
			o.mutex.Unlock()

			// Optimize template
			optimized, err := o.OptimizeTemplate(template)
			if err != nil {
				errMutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to optimize template %s: %w", template.ID, err)
				}
				errMutex.Unlock()
				return
			}

			// Calculate optimized size
			optimizedSize := estimateTemplateSize(optimized)
			o.mutex.Lock()
			o.optimizedSizes[template.ID] = optimizedSize
			totalOriginalSize += originalSize
			totalOptimizedSize += optimizedSize
			o.mutex.Unlock()

			// Store result
			result[i] = optimized
		}(i, template)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check for errors
	if firstErr != nil {
		return nil, firstErr
	}

	// Update statistics
	atomic.AddInt64(&o.stats.TemplatesOptimized, int64(len(templates)))
	atomic.AddInt64(&o.stats.BytesSaved, int64(totalOriginalSize-totalOptimizedSize))
	o.stats.OptimizationTime += time.Since(startTime)
	atomic.AddInt64(&o.optimizationCount, 1)
	o.lastOptimization = time.Now()

	// Take a snapshot of memory after optimization
	o.profiler.CreateSnapshot("after_batch")

	// Compare memory snapshots
	diff, err := o.profiler.CompareSnapshots("before_batch", "after_batch")
	if err == nil && diff != nil {
		if heapDiff, ok := diff["heap_alloc_diff_mb"].(float64); ok && heapDiff < 0 {
			o.stats.MemoryReduced -= heapDiff // Convert negative diff to positive reduction
		}
	}

	// Trigger garbage collection if needed
	if o.config.EnableGarbageCollectionHints && o.shouldTriggerGC() {
		o.triggerGC()
	}

	return result, nil

// deduplicateTemplate deduplicates a template
func (o *MemoryOptimizer) deduplicateTemplate(template *format.Template) (*format.Template, error) {
	// Implementation would deduplicate template content
	// This is a simplified version
	atomic.AddInt64(&o.stats.DuplicatesRemoved, 1)
	return template, nil

// deduplicateSections deduplicates template sections
func (o *MemoryOptimizer) deduplicateSections(sections []format.TemplateSection) ([]format.TemplateSection, error) {
	if len(sections) == 0 {
		return sections, nil
	}

	// Implementation would deduplicate sections
	// This is a simplified version
	atomic.AddInt64(&o.stats.SectionsOptimized, int64(len(sections)))
	return sections, nil

// optimizeVariables optimizes template variables
func (o *MemoryOptimizer) optimizeVariables(variables map[string]interface{}) (map[string]interface{}, error) {
	if len(variables) == 0 {
		return variables, nil
	}

	// Implementation would optimize variables
	// This is a simplified version
	atomic.AddInt64(&o.stats.VariablesOptimized, int64(len(variables)))
	return variables, nil

// flattenInheritance flattens template inheritance
func (o *MemoryOptimizer) flattenInheritance(template *format.Template) (*format.Template, error) {
	// Implementation would flatten inheritance
	// This is a simplified version
	return template, nil

// shouldTriggerGC checks if garbage collection should be triggered
func (o *MemoryOptimizer) shouldTriggerGC() bool {
	// Check memory usage
	memoryUsage := o.profiler.GetMemoryUsage()
	return memoryUsage > float64(o.config.MemoryThreshold)

// triggerGC triggers garbage collection
func (o *MemoryOptimizer) triggerGC() {
	runtime.GC()
	atomic.AddInt64(&o.stats.GarbageCollections, 1)

// GetStats returns statistics for the memory optimizer
func (o *MemoryOptimizer) GetStats() *MemoryOptimizerStats {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.stats

// GetConfig returns the configuration for the memory optimizer
func (o *MemoryOptimizer) GetConfig() *MemoryOptimizerConfig {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.config

// SetConfig sets the configuration for the memory optimizer
func (o *MemoryOptimizer) SetConfig(config *MemoryOptimizerConfig) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.config = config

// GetMemoryProfiler returns the memory profiler
func (o *MemoryOptimizer) GetMemoryProfiler() *profiling.MemoryProfiler {
	return o.profiler

// GetOptimizationCount returns the number of optimization operations
func (o *MemoryOptimizer) GetOptimizationCount() int64 {
	return atomic.LoadInt64(&o.optimizationCount)

// GetLastOptimizationTime returns the time of the last optimization
func (o *MemoryOptimizer) GetLastOptimizationTime() time.Time {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.lastOptimization

// GetTemplateSizes returns the sizes of templates
func (o *MemoryOptimizer) GetTemplateSizes() map[string]int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Create a copy of the map
	result := make(map[string]int)
	for id, size := range o.templateSizes {
		result[id] = size
	}

	return result

// GetOptimizedSizes returns the sizes of optimized templates
func (o *MemoryOptimizer) GetOptimizedSizes() map[string]int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Create a copy of the map
	result := make(map[string]int)
	for id, size := range o.optimizedSizes {
		result[id] = size
	}

	return result

// GetMemorySavings returns the memory savings for templates
func (o *MemoryOptimizer) GetMemorySavings() map[string]float64 {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Create result map
	result := make(map[string]float64)

	// Calculate savings
	for id, originalSize := range o.templateSizes {
		if optimizedSize, ok := o.optimizedSizes[id]; ok {
			if originalSize > 0 {
				savings := float64(originalSize-optimizedSize) / float64(originalSize) * 100
				result[id] = savings
			}
		}
	}

	return result

// estimateTemplateSize estimates the size of a template in bytes
func estimateTemplateSize(template *format.Template) int {
	if template == nil {
		return 0
	}

	size := len(template.ID)

	// Add size of raw content
	if template.Content != nil {
		size += len(template.Content)
	}

	// Add size of template info
	size += len(template.Info.Name)
	size += len(template.Info.Description)
	size += len(template.Info.Version)
	size += len(template.Info.Author)
	size += len(template.Info.Severity)
	for _, tag := range template.Info.Tags {
		size += len(tag)
	}
	for _, ref := range template.Info.References {
		size += len(ref)
	}

	// Add size of test definition
	size += len(template.Test.Prompt)
	size += len(template.Test.ExpectedBehavior)
	size += len(template.Test.Detection.Type)
	size += len(template.Test.Detection.Match)
	size += len(template.Test.Detection.Pattern)
	size += len(template.Test.Detection.Criteria)
	size += len(template.Test.Detection.Condition)

	// Add size of variations
	for _, variation := range template.Test.Variations {
		size += len(variation.Prompt)
		size += len(variation.Detection.Type)
		size += len(variation.Detection.Match)
		size += len(variation.Detection.Pattern)
		size += len(variation.Detection.Criteria)
		size += len(variation.Detection.Condition)
	}

	// Add size of variables at template level
	if template.Variables != nil {
		for name, value := range template.Variables {
			size += len(name)
			
			// Estimate size of value
			switch v := value.(type) {
			case string:
				size += len(v)
			case []byte:
				size += len(v)
			default:
				size += 8 // Assume 8 bytes for other types
			}
		}
	}

