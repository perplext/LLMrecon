package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution/optimizer"
	"github.com/perplext/LLMrecon/src/template/management/optimization"
	"github.com/perplext/LLMrecon/src/utils/concurrency"
	"github.com/perplext/LLMrecon/src/utils/config"
	"github.com/perplext/LLMrecon/src/utils/monitoring"
	"github.com/perplext/LLMrecon/src/utils/profiling"
	"github.com/perplext/LLMrecon/src/utils/resource"
)

// Command-line flags
var (
	envFlag         = flag.String("env", "dev", "Environment (dev, test, prod)")
	templateCountFlag = flag.Int("templates", 1000, "Number of templates to process")
	iterationsFlag  = flag.Int("iterations", 5, "Number of iterations to run")
	verboseFlag     = flag.Bool("verbose", false, "Enable verbose logging")
	profileFlag     = flag.Bool("profile", false, "Enable memory profiling")
	monitorFlag     = flag.Bool("monitor", true, "Enable monitoring")
	optimizeFlag    = flag.Bool("optimize", true, "Enable memory optimization")
	concurrencyFlag = flag.Bool("concurrency", true, "Enable concurrency management")
	poolingFlag     = flag.Bool("pooling", true, "Enable resource pooling")
)

func main() {
	// Parse command-line flags
	flag.Parse()

	// Set environment
	os.Setenv("APP_ENV", *envFlag)
	fmt.Printf("Running in %s environment\n", *envFlag)

	// Get memory configuration
	memConfig := config.GetMemoryConfig()
	fmt.Printf("Loaded configuration for %s environment\n", memConfig.GetEnvironment())

	// Initialize components based on configuration and flags
	var memProfiler *profiling.MemoryProfiler
	var poolManager *resource.PoolManager
	var concurrencyManager *concurrency.ConcurrencyManager
	var memoryOptimizer *optimization.MemoryOptimizer
	var executionOptimizer *optimizer.ExecutionOptimizer
	var monitoringService *monitoring.MonitoringService

	// Initialize memory profiler if enabled
	if *profileFlag && memConfig.ProfilerEnabled {
		memProfiler = initializeMemoryProfiler(memConfig)
		defer memProfiler.Stop()
	}

	// Initialize resource pool manager if enabled
	if *poolingFlag && memConfig.PoolManagerEnabled {
		poolManager = initializeResourcePoolManager(memConfig)
		defer poolManager.CloseAllPools()
	}

	// Initialize concurrency manager if enabled
	if *concurrencyFlag && memConfig.ConcurrencyManagerEnabled {
		concurrencyManager = initializeConcurrencyManager(memConfig)
		defer concurrencyManager.Shutdown()
	}

	// Initialize memory optimizer if enabled
	if *optimizeFlag && memConfig.MemoryOptimizerEnabled {
		memoryOptimizer = initializeMemoryOptimizer(memConfig)
	}

	// Initialize execution optimizer if enabled
	if memConfig.ExecutionOptimizerEnabled {
		executionOptimizer = initializeExecutionOptimizer(memConfig, memoryOptimizer, concurrencyManager)
	}

	// Initialize monitoring service if enabled
	if *monitorFlag {
		var err error
		monitoringService, err = initializeMonitoringService(memConfig)
		if err != nil {
			log.Fatalf("Failed to initialize monitoring service: %v", err)
		}
		defer monitoringService.Stop()

		// Monitor resource pool if available
		if poolManager != nil {
			// Create connection pool for demonstration
			connectionPool := poolManager.CreatePool("connections", func() (interface{}, error) {
				// Simulate creating a connection
				return &struct{}{}, nil
			}, func(resource interface{}) error {
				// Simulate closing a connection
				return nil
			})

			// Monitor the connection pool
			monitoringService.MonitorResourcePool(connectionPool, "connections")
		}

		// Monitor concurrency manager if available
		if concurrencyManager != nil {
			monitoringService.MonitorConcurrencyManager(concurrencyManager)
		}
	}

	// Run the benchmark
	for i := 0; i < *iterationsFlag; i++ {
		fmt.Printf("\nIteration %d/%d\n", i+1, *iterationsFlag)
		
		// Capture memory before creating templates
		var memoryBefore uint64
		if monitoringService != nil {
			memoryBefore = monitoringService.CaptureMemorySnapshot("before_templates")
		}

		// Create sample templates
		templates := createSampleTemplates(*templateCountFlag)
		fmt.Printf("Created %d templates\n", len(templates))

		// Capture memory after creating templates
		if monitoringService != nil {
			monitoringService.CaptureMemorySnapshot("after_templates")
		}

		// Process templates
		startTime := time.Now()
		processTemplates(templates, executionOptimizer, concurrencyManager, memoryOptimizer, monitoringService, memConfig)
		duration := time.Since(startTime)

		// Capture memory after processing templates
		var memoryAfter uint64
		if monitoringService != nil {
			memoryAfter = monitoringService.CaptureMemorySnapshot("after_processing")
		}

		// Print memory reduction
		if memoryBefore > 0 && memoryAfter > 0 {
			memoryReduction := float64(memoryBefore-memoryAfter) / float64(memoryBefore) * 100
			fmt.Printf("Memory reduction: %.2f%%\n", memoryReduction)
		}

		// Print processing statistics
		fmt.Printf("Processed %d templates in %.2f seconds (%.2f templates/sec)\n", 
			len(templates), duration.Seconds(), float64(len(templates))/duration.Seconds())

		// Force garbage collection between iterations
		runtime.GC()
		time.Sleep(1 * time.Second)
	}

	// Print final memory statistics
	if monitoringService != nil {
		monitoringService.LogMemoryStats()
	}

	fmt.Println("\nBenchmark completed successfully")
}

// initializeMemoryProfiler initializes the memory profiler
func initializeMemoryProfiler(memConfig *config.MemoryConfig) *profiling.MemoryProfiler {
	// Create profiler options
	options := &profiling.MemoryProfilerOptions{
		ProfileInterval:   time.Duration(memConfig.ProfilerInterval) * time.Second,
		OutputDir:         memConfig.ProfilerOutputDir,
		MemoryThreshold:   memConfig.MemoryThreshold * 1024 * 1024, // Convert MB to bytes
		GCThreshold:       time.Duration(memConfig.GCThreshold) * time.Millisecond,
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
		EnableMemoryOptimization:   memConfig.MemoryOptimizerEnabled,
		EnableConcurrencyManagement: memConfig.ConcurrencyManagerEnabled,
		EnableBatchProcessing:      memConfig.EnableBatchProcessing,
		BatchSize:                  memConfig.BatchSize,
		ResultCacheSize:            memConfig.ResultCacheSize,
		ResultCacheTTL:             time.Duration(memConfig.ResultCacheTTL) * time.Second,
	}
	
	// Create execution optimizer
	executionOptimizer := optimizer.NewExecutionOptimizer(options, memoryOptimizer, concurrencyManager)
	
	fmt.Println("Execution optimizer initialized")
	return executionOptimizer
}

// initializeMonitoringService initializes the monitoring service
func initializeMonitoringService(memConfig *config.MemoryConfig) (*monitoring.MonitoringService, error) {
	// Create monitoring service options
	options := &monitoring.MonitoringServiceOptions{
		CollectionInterval:   15 * time.Second,
		LogFile:              "logs/monitoring.log",
		EnableConsoleLogging: *verboseFlag,
		HeapAllocWarningMB:   float64(memConfig.MemoryThreshold) * 0.8,
		HeapAllocCriticalMB:  float64(memConfig.MemoryThreshold),
		AlertCooldown:        5 * time.Minute,
	}
	
	// Create monitoring service
	service, err := monitoring.NewMonitoringService(options)
	if err != nil {
		return nil, err
	}
	
	// Start monitoring service
	service.Start()
	
	fmt.Println("Monitoring service initialized and started")
	return service, nil
}

// createSampleTemplates creates sample templates for testing
func createSampleTemplates(count int) []*format.Template {
	templates := make([]*format.Template, count)
	
	for i := 0; i < count; i++ {
		// Create a template with some common content to demonstrate deduplication
		template := &format.Template{
			ID:      fmt.Sprintf("template-%d", i),
			Name:    fmt.Sprintf("Template %d", i),
			Content: fmt.Sprintf("This is template %d with some common content that can be deduplicated across templates. It contains various sections that might be repeated in multiple templates.", i),
			Variables: map[string]string{
				"var1": fmt.Sprintf("value%d", i),
				"var2": "common value",
				"var3": "another common value",
			},
			Metadata: map[string]interface{}{
				"created": time.Now(),
				"version": 1,
				"tags":    []string{"test", "example", fmt.Sprintf("tag-%d", i%10)},
			},
		}
		
		// Add more content to make templates larger
		for j := 0; j < 5; j++ {
			template.Content += fmt.Sprintf("\nSection %d: This is a section with some content that might be similar across templates. It contains information about the template and its usage.", j)
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
	memoryOptimizer *optimization.MemoryOptimizer,
	monitoringService *monitoring.MonitoringService,
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
			
			// Capture memory before execution
			var memoryBefore uint64
			if monitoringService != nil {
				memoryBefore = monitoringService.CaptureMemorySnapshot("")
			}
			
			execStartTime := time.Now()
			
			// Process template
			_, err := executionOptimizer.ExecuteTemplate(context.Background(), template, nil)
			
			execDuration := time.Since(execStartTime)
			
			// Capture memory after execution
			var memoryAfter uint64
			if monitoringService != nil {
				memoryAfter = monitoringService.CaptureMemorySnapshot("")
				
				// Record template execution metrics
				monitoringService.RecordTemplateExecution(execDuration, memoryBefore, memoryAfter, err)
			}
			
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
				execStartTime := time.Now()
				
				// Optimize template if memory optimizer is available
				var optimizedTemplate *format.Template
				var err error
				if memoryOptimizer != nil {
					optimizedTemplate, err = memoryOptimizer.OptimizeTemplate(ctx, template)
					if err != nil {
						return err
					}
				} else {
					optimizedTemplate = template
				}
				
				// Process template
				_, err = processTemplate(optimizedTemplate)
				
				execDuration := time.Since(execStartTime)
				
				// Record execution time if monitoring is enabled
				if monitoringService != nil {
					monitoringService.GetMetricsManager().ObserveHistogram(
						"template.execution.time", 
						float64(execDuration.Milliseconds()),
					)
				}
				
				return err
			})
			
			if err != nil {
				log.Printf("Error submitting template %s: %v\n", template.ID, err)
			}
		}
		
		// Wait for all tasks to complete
		if err := concurrencyManager.Wait(context.Background()); err != nil {
			log.Printf("Error waiting for tasks to complete: %v\n", err)
		}
		
		processedCount = len(templates)
	} else {
		// Process templates sequentially
		for _, template := range templates {
			// Optimize template if memory optimizer is available
			var optimizedTemplate *format.Template
			var err error
			if memoryOptimizer != nil {
				optimizedTemplate, err = memoryOptimizer.OptimizeTemplate(context.Background(), template)
				if err != nil {
					log.Printf("Error optimizing template %s: %v\n", template.ID, err)
					continue
				}
			} else {
				optimizedTemplate = template
			}
			
			// Process template
			_, err = processTemplate(optimizedTemplate)
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
}

// processTemplate simulates processing a template
func processTemplate(template *format.Template) (string, error) {
	// Simulate processing time
	time.Sleep(5 * time.Millisecond)
	
	// Return processed result
	return fmt.Sprintf("Processed: %s", template.Content), nil
}
