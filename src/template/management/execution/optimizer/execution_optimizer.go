package optimizer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/optimization"
	"github.com/perplext/LLMrecon/src/utils/concurrency"
	"github.com/perplext/LLMrecon/src/utils/profiling"
)

// ExecutionOptimizer optimizes template execution for memory efficiency and concurrency
type ExecutionOptimizer struct {
	// engine is the template execution engine
	engine *execution.Engine
	// memoryOptimizer is the memory optimizer
	memoryOptimizer *optimization.MemoryOptimizer
	// concurrencyManager is the concurrency manager
	concurrencyManager *concurrency.ConcurrencyManager
	// profiler is the memory profiler
	profiler *profiling.MemoryProfiler
	// config contains the optimizer configuration
	config *ExecutionOptimizerConfig
	// mutex protects the optimizer state
	mutex sync.RWMutex
	// stats tracks optimization statistics
	stats *ExecutionOptimizerStats
	// running indicates if the optimizer is running
	running bool
	// stopChan is used to stop the optimizer
	stopChan chan struct{}
}

// ExecutionOptimizerConfig represents configuration for the execution optimizer
type ExecutionOptimizerConfig struct {
	// EnableMemoryOptimization enables memory optimization
	EnableMemoryOptimization bool
	// EnableConcurrencyOptimization enables concurrency optimization
	EnableConcurrencyOptimization bool
	// EnableResultCaching enables result caching
	EnableResultCaching bool
	// ResultCacheSize is the size of the result cache
	ResultCacheSize int
	// ResultCacheTTL is the TTL for cached results
	ResultCacheTTL time.Duration
	// EnableBatchProcessing enables batch processing of templates
	EnableBatchProcessing bool
	// BatchSize is the size of template batches
	BatchSize int
	// MaxConcurrentExecutions is the maximum number of concurrent executions
	MaxConcurrentExecutions int
	// MemoryThreshold is the memory threshold for optimization (in MB)
	MemoryThreshold int64
	// ExecutionTimeout is the timeout for template execution
	ExecutionTimeout time.Duration
	// EnableAdaptiveTimeouts enables adaptive timeouts based on template complexity
	EnableAdaptiveTimeouts bool
}

// ExecutionOptimizerStats tracks statistics for the execution optimizer
type ExecutionOptimizerStats struct {
	// TemplatesExecuted is the number of templates executed
	TemplatesExecuted int64
	// TemplatesOptimized is the number of templates optimized
	TemplatesOptimized int64
	// MemorySaved is the amount of memory saved (in MB)
	MemorySaved float64
	// ExecutionTime is the total execution time
	ExecutionTime time.Duration
	// AverageExecutionTime is the average execution time per template
	AverageExecutionTime time.Duration
	// CacheHits is the number of cache hits
	CacheHits int64
	// CacheMisses is the number of cache misses
	CacheMisses int64
	// BatchesProcessed is the number of batches processed
	BatchesProcessed int64
	// ExecutionErrors is the number of execution errors
	ExecutionErrors int64
	// TimeoutErrors is the number of timeout errors
	TimeoutErrors int64
	// MemoryErrors is the number of memory-related errors
	MemoryErrors int64
}

// DefaultExecutionOptimizerConfig returns default configuration for the execution optimizer
func DefaultExecutionOptimizerConfig() *ExecutionOptimizerConfig {
	return &ExecutionOptimizerConfig{
		EnableMemoryOptimization:    true,
		EnableConcurrencyOptimization: true,
		EnableResultCaching:         true,
		ResultCacheSize:             1000,
		ResultCacheTTL:              30 * time.Minute,
		EnableBatchProcessing:       true,
		BatchSize:                   10,
		MaxConcurrentExecutions:     100,
		MemoryThreshold:             100, // 100 MB
		ExecutionTimeout:            30 * time.Second,
		EnableAdaptiveTimeouts:      true,
	}
}

// NewExecutionOptimizer creates a new execution optimizer
func NewExecutionOptimizer(engine *execution.Engine, config *ExecutionOptimizerConfig) (*ExecutionOptimizer, error) {
	if engine == nil {
		return nil, fmt.Errorf("execution engine is required")
	}

	if config == nil {
		config = DefaultExecutionOptimizerConfig()
	}

	// Create memory profiler
	profilerOptions := profiling.DefaultProfilerOptions()
	profiler, err := profiling.NewMemoryProfiler(profilerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory profiler: %w", err)
	}

	// Create memory optimizer if enabled
	var memoryOptimizer *optimization.MemoryOptimizer
	if config.EnableMemoryOptimization {
		optimizerConfig := optimization.DefaultMemoryOptimizerConfig()
		memoryOptimizer, err = optimization.NewMemoryOptimizer(optimizerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create memory optimizer: %w", err)
		}
	}

	// Create concurrency manager if enabled
	var concurrencyManager *concurrency.ConcurrencyManager
	if config.EnableConcurrencyOptimization {
		managerConfig := concurrency.DefaultManagerConfig()
		managerConfig.MaxWorkers = config.MaxConcurrentExecutions
		concurrencyManager, err = concurrency.NewConcurrencyManager(managerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create concurrency manager: %w", err)
		}
	}

	return &ExecutionOptimizer{
		engine:             engine,
		memoryOptimizer:    memoryOptimizer,
		concurrencyManager: concurrencyManager,
		profiler:           profiler,
		config:             config,
		stats:              &ExecutionOptimizerStats{},
		stopChan:           make(chan struct{}),
	}, nil
}

// Start starts the execution optimizer
func (o *ExecutionOptimizer) Start() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.running {
		return fmt.Errorf("execution optimizer is already running")
	}

	o.running = true
	o.stopChan = make(chan struct{})

	// Start memory profiler
	if err := o.profiler.StartAutomaticProfiling(); err != nil {
		o.running = false
		return fmt.Errorf("failed to start memory profiler: %w", err)
	}

	// Start concurrency manager if enabled
	if o.config.EnableConcurrencyOptimization && o.concurrencyManager != nil {
		if err := o.concurrencyManager.Start(); err != nil {
			o.running = false
			o.profiler.StopAutomaticProfiling()
			return fmt.Errorf("failed to start concurrency manager: %w", err)
		}
	}

	return nil
}

// Stop stops the execution optimizer
func (o *ExecutionOptimizer) Stop() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.running {
		return
	}

	close(o.stopChan)
	o.running = false

	// Stop memory profiler
	o.profiler.StopAutomaticProfiling()

	// Stop concurrency manager if enabled
	if o.config.EnableConcurrencyOptimization && o.concurrencyManager != nil {
		o.concurrencyManager.Stop()
	}
}

// ExecuteTemplate executes a template with optimization
func (o *ExecutionOptimizer) ExecuteTemplate(ctx context.Context, template *format.Template, data interface{}) (string, error) {
	if template == nil {
		return "", fmt.Errorf("template is nil")
	}

	o.mutex.RLock()
	running := o.running
	o.mutex.RUnlock()

	if !running {
		return "", fmt.Errorf("execution optimizer is not running")
	}

	// Take a snapshot of memory before execution
	o.profiler.CreateSnapshot("before_execution")

	// Start execution timer
	startTime := time.Now()

	// Optimize template if enabled
	var optimizedTemplate *format.Template
	var err error
	if o.config.EnableMemoryOptimization && o.memoryOptimizer != nil {
		optimizedTemplate, err = o.memoryOptimizer.OptimizeTemplate(template)
		if err != nil {
			return "", fmt.Errorf("failed to optimize template: %w", err)
		}
		atomic.AddInt64(&o.stats.TemplatesOptimized, 1)
	} else {
		optimizedTemplate = template
	}

	// Execute template
	var result string
	if o.config.EnableConcurrencyOptimization && o.concurrencyManager != nil {
		// Execute with concurrency manager
		result, err = o.executeWithConcurrencyManager(ctx, optimizedTemplate, data)
	} else {
		// Execute directly
		result, err = o.engine.ExecuteTemplate(ctx, optimizedTemplate, data)
	}

	// Update execution time statistics
	executionTime := time.Since(startTime)
	o.mutex.Lock()
	o.stats.ExecutionTime += executionTime
	o.stats.TemplatesExecuted++
	if o.stats.TemplatesExecuted > 0 {
		o.stats.AverageExecutionTime = o.stats.ExecutionTime / time.Duration(o.stats.TemplatesExecuted)
	}
	o.mutex.Unlock()

	// Take a snapshot of memory after execution
	o.profiler.CreateSnapshot("after_execution")

	// Compare memory snapshots
	diff, diffErr := o.profiler.CompareSnapshots("before_execution", "after_execution")
	if diffErr == nil && diff != nil {
		if heapDiff, ok := diff["heap_alloc_diff_mb"].(float64); ok && heapDiff < 0 {
			o.mutex.Lock()
			o.stats.MemorySaved -= heapDiff // Convert negative diff to positive savings
			o.mutex.Unlock()
		}
	}

	// Update error statistics if needed
	if err != nil {
		o.mutex.Lock()
		o.stats.ExecutionErrors++
		if err == context.DeadlineExceeded {
			o.stats.TimeoutErrors++
		}
		o.mutex.Unlock()
	}

	return result, err
}

// executeWithConcurrencyManager executes a template using the concurrency manager
func (o *ExecutionOptimizer) executeWithConcurrencyManager(ctx context.Context, template *format.Template, data interface{}) (string, error) {
	// Create a task for the concurrency manager
	task := &templateExecutionTask{
		id:       template.ID,
		template: template,
		data:     data,
		engine:   o.engine,
		result:   "",
		err:      nil,
		done:     make(chan struct{}),
	}

	// Submit task to concurrency manager
	if err := o.concurrencyManager.Submit(task); err != nil {
		return "", fmt.Errorf("failed to submit task: %w", err)
	}

	// Wait for task completion or context cancellation
	select {
	case <-task.done:
		return task.result, task.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// ExecuteTemplates executes multiple templates with optimization
func (o *ExecutionOptimizer) ExecuteTemplates(ctx context.Context, templates []*format.Template, data interface{}) ([]string, error) {
	if len(templates) == 0 {
		return nil, fmt.Errorf("templates is empty")
	}

	o.mutex.RLock()
	running := o.running
	enableBatchProcessing := o.config.EnableBatchProcessing
	batchSize := o.config.BatchSize
	o.mutex.RUnlock()

	if !running {
		return nil, fmt.Errorf("execution optimizer is not running")
	}

	// Take a snapshot of memory before execution
	o.profiler.CreateSnapshot("before_batch_execution")

	// Start execution timer
	startTime := time.Now()

	// Optimize templates if enabled
	var optimizedTemplates []*format.Template
	if o.config.EnableMemoryOptimization && o.memoryOptimizer != nil {
		var err error
		optimizedTemplates, err = o.memoryOptimizer.OptimizeTemplates(templates)
		if err != nil {
			return nil, fmt.Errorf("failed to optimize templates: %w", err)
		}
		atomic.AddInt64(&o.stats.TemplatesOptimized, int64(len(templates)))
	} else {
		optimizedTemplates = templates
	}

	// Execute templates
	var results []string
	var err error
	if enableBatchProcessing {
		// Execute in batches
		results, err = o.executeTemplatesInBatches(ctx, optimizedTemplates, data, batchSize)
	} else {
		// Execute individually
		results = make([]string, len(optimizedTemplates))
		for i, template := range optimizedTemplates {
			results[i], err = o.ExecuteTemplate(ctx, template, data)
			if err != nil {
				return results, fmt.Errorf("failed to execute template %s: %w", template.ID, err)
			}
		}
	}

	// Update execution time statistics
	executionTime := time.Since(startTime)
	o.mutex.Lock()
	o.stats.ExecutionTime += executionTime
	o.stats.TemplatesExecuted += int64(len(templates))
	if o.stats.TemplatesExecuted > 0 {
		o.stats.AverageExecutionTime = o.stats.ExecutionTime / time.Duration(o.stats.TemplatesExecuted)
	}
	o.mutex.Unlock()

	// Take a snapshot of memory after execution
	o.profiler.CreateSnapshot("after_batch_execution")

	// Compare memory snapshots
	diff, diffErr := o.profiler.CompareSnapshots("before_batch_execution", "after_batch_execution")
	if diffErr == nil && diff != nil {
		if heapDiff, ok := diff["heap_alloc_diff_mb"].(float64); ok && heapDiff < 0 {
			o.mutex.Lock()
			o.stats.MemorySaved -= heapDiff // Convert negative diff to positive savings
			o.mutex.Unlock()
		}
	}

	return results, err
}

// executeTemplatesInBatches executes templates in batches
func (o *ExecutionOptimizer) executeTemplatesInBatches(ctx context.Context, templates []*format.Template, data interface{}, batchSize int) ([]string, error) {
	if batchSize <= 0 {
		batchSize = len(templates)
	}

	// Create result slice
	results := make([]string, len(templates))

	// Process templates in batches
	for i := 0; i < len(templates); i += batchSize {
		// Create batch
		end := i + batchSize
		if end > len(templates) {
			end = len(templates)
		}
		batch := templates[i:end]

		// Execute batch
		batchResults, err := o.executeBatch(ctx, batch, data)
		if err != nil {
			return results, fmt.Errorf("failed to execute batch: %w", err)
		}

		// Copy batch results to results slice
		copy(results[i:end], batchResults)

		// Update batch statistics
		atomic.AddInt64(&o.stats.BatchesProcessed, 1)
	}

	return results, nil
}

// executeBatch executes a batch of templates
func (o *ExecutionOptimizer) executeBatch(ctx context.Context, templates []*format.Template, data interface{}) ([]string, error) {
	// Create result slice
	results := make([]string, len(templates))

	// Create wait group for batch execution
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	var firstErr error

	// Execute templates in parallel
	for i, template := range templates {
		wg.Add(1)
		go func(i int, template *format.Template) {
			defer wg.Done()

			// Execute template
			result, err := o.ExecuteTemplate(ctx, template, data)
			if err != nil {
				errMutex.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to execute template %s: %w", template.ID, err)
				}
				errMutex.Unlock()
				return
			}

			// Store result
			results[i] = result
		}(i, template)
	}

	// Wait for all templates to be executed
	wg.Wait()

	return results, firstErr
}

// GetStats returns statistics for the execution optimizer
func (o *ExecutionOptimizer) GetStats() *ExecutionOptimizerStats {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Create a copy of the stats
	stats := &ExecutionOptimizerStats{
		TemplatesExecuted:    atomic.LoadInt64(&o.stats.TemplatesExecuted),
		TemplatesOptimized:   atomic.LoadInt64(&o.stats.TemplatesOptimized),
		MemorySaved:          o.stats.MemorySaved,
		ExecutionTime:        o.stats.ExecutionTime,
		AverageExecutionTime: o.stats.AverageExecutionTime,
		CacheHits:            atomic.LoadInt64(&o.stats.CacheHits),
		CacheMisses:          atomic.LoadInt64(&o.stats.CacheMisses),
		BatchesProcessed:     atomic.LoadInt64(&o.stats.BatchesProcessed),
		ExecutionErrors:      atomic.LoadInt64(&o.stats.ExecutionErrors),
		TimeoutErrors:        atomic.LoadInt64(&o.stats.TimeoutErrors),
		MemoryErrors:         atomic.LoadInt64(&o.stats.MemoryErrors),
	}

	return stats
}

// GetConfig returns the configuration for the execution optimizer
func (o *ExecutionOptimizer) GetConfig() *ExecutionOptimizerConfig {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.config
}

// SetConfig sets the configuration for the execution optimizer
func (o *ExecutionOptimizer) SetConfig(config *ExecutionOptimizerConfig) {
	if config == nil {
		return
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.config = config
}

// IsRunning returns if the execution optimizer is running
func (o *ExecutionOptimizer) IsRunning() bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return o.running
}

// GetMemoryProfiler returns the memory profiler
func (o *ExecutionOptimizer) GetMemoryProfiler() *profiling.MemoryProfiler {
	return o.profiler
}

// GetMemoryOptimizer returns the memory optimizer
func (o *ExecutionOptimizer) GetMemoryOptimizer() *optimization.MemoryOptimizer {
	return o.memoryOptimizer
}

// GetConcurrencyManager returns the concurrency manager
func (o *ExecutionOptimizer) GetConcurrencyManager() *concurrency.ConcurrencyManager {
	return o.concurrencyManager
}

// templateExecutionTask represents a template execution task for the concurrency manager
type templateExecutionTask struct {
	id       string
	template *format.Template
	data     interface{}
	engine   *execution.Engine
	result   string
	err      error
	done     chan struct{}
}

// Execute executes the template execution task
func (t *templateExecutionTask) Execute(ctx context.Context) error {
	// Execute template
	result, err := t.engine.ExecuteTemplate(ctx, t.template, t.data)

	// Store result and error
	t.result = result
	t.err = err

	// Signal completion
	close(t.done)

	return err
}

// ID returns the task ID
func (t *templateExecutionTask) ID() string {
	return t.id
}

// Priority returns the task priority
func (t *templateExecutionTask) Priority() int {
	return 0
}
