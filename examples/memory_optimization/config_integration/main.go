package main

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution/optimizer"
	"github.com/perplext/LLMrecon/src/template/management/optimization"
	"github.com/perplext/LLMrecon/src/utils/concurrency"
	"github.com/perplext/LLMrecon/src/utils/config"
	"github.com/perplext/LLMrecon/src/utils/profiling"
	"github.com/perplext/LLMrecon/src/utils/resource"
)

// Example demonstrates how to use the memory optimization components
// with environment-specific configurations
func main() {
	// Set environment (can also be set via APP_ENV environment variable)
	if len(os.Args) > 1 {
		os.Setenv("APP_ENV", os.Args[1])
	}

	// Get memory configuration
	memConfig := config.GetMemoryConfig()
	
	fmt.Printf("Running with %s environment configuration\n", memConfig.GetEnvironment())
	
	// Initialize memory profiler if enabled
	var memProfiler *profiling.MemoryProfiler
	if memConfig.ProfilerEnabled {
		memProfiler = initializeMemoryProfiler(memConfig)
		defer memProfiler.Stop()
	}
	
	// Initialize resource pool manager if enabled
	var poolManager *resource.PoolManager
	if memConfig.PoolManagerEnabled {
		poolManager = initializeResourcePoolManager(memConfig)
		defer poolManager.CloseAllPools()
	}
	
	// Initialize concurrency manager if enabled
	var concurrencyManager *concurrency.ConcurrencyManager
	if memConfig.ConcurrencyManagerEnabled {
		concurrencyManager = initializeConcurrencyManager(memConfig)
		defer concurrencyManager.Shutdown()
	}
	
	// Initialize memory optimizer if enabled
	var memoryOptimizer *optimization.MemoryOptimizer
	if memConfig.MemoryOptimizerEnabled {
		memoryOptimizer = initializeMemoryOptimizer(memConfig)
	}
	
	// Initialize execution optimizer if enabled
	var executionOptimizer *optimizer.ExecutionOptimizer
	if memConfig.ExecutionOptimizerEnabled {
		executionOptimizer = initializeExecutionOptimizer(memConfig, memoryOptimizer, concurrencyManager)
	}
	
	// Create sample templates
	templates := createSampleTemplates(100)
	
	// Capture memory before optimization
	if memProfiler != nil {
		memProfiler.CaptureMemorySnapshot("before_optimization")
	}
	
	// Process templates
	processTemplates(templates, executionOptimizer, concurrencyManager, memConfig)
	
	// Capture memory after optimization
	if memProfiler != nil {
		memProfiler.CaptureMemorySnapshot("after_optimization")
		
		// Print memory statistics
		beforeStats, _ := memProfiler.GetMemorySnapshot("before_optimization")
		afterStats, _ := memProfiler.GetMemorySnapshot("after_optimization")
		
		if beforeStats != nil && afterStats != nil {
			memoryReduction := float64(beforeStats.Alloc - afterStats.Alloc) / float64(beforeStats.Alloc) * 100
			fmt.Printf("Memory reduction: %.2f%%\n", memoryReduction)
		}
	}
}

// initializeMemoryProfiler initializes the memory profiler
func initializeMemoryProfiler(memConfig *config.MemoryConfig) *profiling.MemoryProfiler {
	// Create profiler options
	options := &profiling.MemoryProfilerOptions{
		ProfileInterval:  time.Duration(memConfig.ProfilerInterval) * time.Second,
		OutputDir:        memConfig.ProfilerOutputDir,
		MemoryThreshold:  memConfig.MemoryThreshold * 1024 * 1024, // Convert MB to bytes
		GCThreshold:      time.Duration(memConfig.GCThreshold) * time.Millisecond,
		EnableAutoProfile: true,
	}
	
	// Create memory profiler
	profiler := profiling.NewMemoryProfiler(options)
	
	// Start profiler
	profiler.Start()
	
	fmt.Println("Memory profiler initialized and started")
	return profiler
}

// initializeResourcePoolManager initializes the resource pool manager
func initializeResourcePoolManager(memConfig *config.MemoryConfig) *resource.PoolManager {
	// Create pool manager options
	options := &resource.PoolManagerOptions{
		DefaultPoolSize:    memConfig.DefaultPoolSize,
		MinPoolSize:        memConfig.MinPoolSize,
		MaxPoolSize:        memConfig.MaxPoolSize,
		EnablePoolScaling:  memConfig.EnablePoolScaling,
		ScaleUpThreshold:   memConfig.ScaleUpThreshold,
		ScaleDownThreshold: memConfig.ScaleDownThreshold,
	}
	
	// Create resource pool manager
	poolManager := resource.NewPoolManager(options)
	
	fmt.Println("Resource pool manager initialized")
	return poolManager
}

// initializeConcurrencyManager initializes the concurrency manager
func initializeConcurrencyManager(memConfig *config.MemoryConfig) *concurrency.ConcurrencyManager {
	// Create concurrency manager options
	options := &concurrency.ConcurrencyManagerOptions{
		MaxWorkers:        memConfig.MaxWorkers,
		MinWorkers:        memConfig.MinWorkers,
		QueueSize:         memConfig.QueueSize,
		WorkerIdleTimeout: time.Duration(memConfig.WorkerIdleTimeout) * time.Second,
	}
	
	// Create concurrency manager
	manager := concurrency.NewConcurrencyManager(options)
	
	fmt.Println("Concurrency manager initialized")
	return manager
}

// initializeMemoryOptimizer initializes the memory optimizer
func initializeMemoryOptimizer(memConfig *config.MemoryConfig) *optimization.MemoryOptimizer {
	// Create memory optimizer options
	options := &optimization.MemoryOptimizerOptions{
		EnableDeduplication: memConfig.EnableDeduplication,
		EnableCompression:   memConfig.EnableCompression,
		EnableLazyLoading:   memConfig.EnableLazyLoading,
		EnableGCHints:       memConfig.EnableGCHints,
	}
	
	// Create memory optimizer
	optimizer := optimization.NewMemoryOptimizer(options)
	
	fmt.Println("Memory optimizer initialized")
	return optimizer
}

// initializeExecutionOptimizer initializes the execution optimizer
func initializeExecutionOptimizer(
	memConfig *config.MemoryConfig,
	memoryOptimizer *optimization.MemoryOptimizer,
	concurrencyManager *concurrency.ConcurrencyManager,
) *optimizer.ExecutionOptimizer {
	// Create execution optimizer options
	options := &optimizer.ExecutionOptimizerOptions{
		EnableMemoryOptimization: memConfig.MemoryOptimizerEnabled,
		EnableConcurrencyManagement: memConfig.ConcurrencyManagerEnabled,
		EnableBatchProcessing:    memConfig.EnableBatchProcessing,
		BatchSize:                memConfig.BatchSize,
		ResultCacheSize:          memConfig.ResultCacheSize,
		ResultCacheTTL:           time.Duration(memConfig.ResultCacheTTL) * time.Second,
	}
	
	// Create execution optimizer
	executionOptimizer := optimizer.NewExecutionOptimizer(options, memoryOptimizer, concurrencyManager)
	
	fmt.Println("Execution optimizer initialized")
	return executionOptimizer
}

// createSampleTemplates creates sample templates for testing
func createSampleTemplates(count int) []*format.Template {
	templates := make([]*format.Template, count)
	
	for i := 0; i < count; i++ {
		// Create a template with some common content to demonstrate deduplication
		template := &format.Template{
			ID:      fmt.Sprintf("template-%d", i),
			Name:    fmt.Sprintf("Template %d", i),
			Content: fmt.Sprintf("This is template %d with some common content that can be deduplicated across templates.", i),
			Variables: map[string]string{
				"var1": fmt.Sprintf("value%d", i),
				"var2": "common value",
			},
			Metadata: map[string]interface{}{
				"created": time.Now(),
				"version": 1,
			},
		}
		
		templates[i] = template
	}
	
	return templates
}

// processTemplates processes templates using the execution optimizer
func processTemplates(
	templates []*format.Template,
	executionOptimizer *optimizer.ExecutionOptimizer,
	concurrencyManager *concurrency.ConcurrencyManager,
	memConfig *config.MemoryConfig,
) {
	fmt.Printf("Processing %d templates\n", len(templates))
	
	startTime := time.Now()
	processedCount := 0
	
	// Process templates based on configuration
	if executionOptimizer != nil {
		// Use execution optimizer
		for _, template := range templates {
			template := template // Create local copy for closure
			
			// Process template
			_, err := executionOptimizer.ExecuteTemplate(context.Background(), template, nil)
			if err != nil {
				log.Printf("Error executing template %s: %v\n", template.ID, err)
			} else {
				processedCount++
			}
		}
	} else if concurrencyManager != nil && memConfig.ConcurrencyManagerEnabled {
		// Use concurrency manager directly
		for _, template := range templates {
			template := template // Create local copy for closure
			
			err := concurrencyManager.Submit(func(ctx context.Context) error {
if err != nil {
treturn err
if err != nil {
treturn err
}}				// Process template
				_, err := processTemplate(template)
				return err
			})
			
			if err != nil {
				log.Printf("Error submitting template %s: %v\n", template.ID, err)
			}
if err != nil {
treturn err
}		}
		
		// Wait for all tasks to complete
		if err := concurrencyManager.Wait(context.Background()); err != nil {
			log.Printf("Error waiting for tasks to complete: %v\n", err)
		}
		
if err != nil {
treturn err
}		processedCount = len(templates)
	} else {
		// Process templates sequentially
		for _, template := range templates {
			// Process template
			_, err := processTemplate(template)
			if err != nil {
				log.Printf("Error processing template %s: %v\n", template.ID, err)
			} else {
				processedCount++
			}
		}
	}
	
	// Print processing statistics
	duration := time.Since(startTime)
	fmt.Printf("Processed %d templates in %.2f seconds (%.2f templates/sec)\n", 
		processedCount, duration.Seconds(), float64(processedCount)/duration.Seconds())
	
	// Print memory statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory usage: %.2f MB\n", float64(m.Alloc)/1024/1024)
}

// processTemplate simulates processing a template
func processTemplate(template *format.Template) (string, error) {
	// Simulate processing time
	time.Sleep(10 * time.Millisecond)
	
	// Return processed result
	return fmt.Sprintf("Processed: %s", template.Content), nil
}
