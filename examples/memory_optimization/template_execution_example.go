package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/execution/optimizer"
	"github.com/perplext/LLMrecon/src/template/management/optimization"
	"github.com/perplext/LLMrecon/src/utils/concurrency"
	"github.com/perplext/LLMrecon/src/utils/profiling"
	"github.com/perplext/LLMrecon/src/utils/resource"
)

// Configuration constants
const (
	NumTemplates            = 100
	TemplateSize            = 5000
	NumVariables            = 20
	NumConcurrent           = 10
	EnableMemoryOptimizer   = true
	EnableConcurrencyManager = true
	EnableBatchProcessing   = true
	BatchSize               = 10
	EnableResourcePooling   = true
	PoolSize                = 100
)

func main() {
	// Parse environment variables for configuration
	numTemplates := getEnvInt("NUM_TEMPLATES", NumTemplates)
	templateSize := getEnvInt("TEMPLATE_SIZE", TemplateSize)
	numVariables := getEnvInt("NUM_VARIABLES", NumVariables)
	numConcurrent := getEnvInt("NUM_CONCURRENT", NumConcurrent)
	enableMemoryOptimizer := getEnvBool("ENABLE_MEMORY_OPTIMIZER", EnableMemoryOptimizer)
	enableConcurrencyManager := getEnvBool("ENABLE_CONCURRENCY_MANAGER", EnableConcurrencyManager)
	enableBatchProcessing := getEnvBool("ENABLE_BATCH_PROCESSING", EnableBatchProcessing)
	batchSize := getEnvInt("BATCH_SIZE", BatchSize)
	enableResourcePooling := getEnvBool("ENABLE_RESOURCE_POOLING", EnableResourcePooling)
	poolSize := getEnvInt("POOL_SIZE", PoolSize)

	// Print configuration
	fmt.Println("Template Execution with Memory Optimization Example")
	fmt.Println("--------------------------------------------------")
	fmt.Println("Configuration:")
	fmt.Printf("- Number of Templates: %d\n", numTemplates)
	fmt.Printf("- Template Size: %d bytes\n", templateSize)
	fmt.Printf("- Number of Variables: %d\n", numVariables)
	fmt.Printf("- Number of Concurrent Operations: %d\n", numConcurrent)
	fmt.Printf("- Memory Optimizer Enabled: %t\n", enableMemoryOptimizer)
	fmt.Printf("- Concurrency Manager Enabled: %t\n", enableConcurrencyManager)
	fmt.Printf("- Batch Processing Enabled: %t\n", enableBatchProcessing)
	fmt.Printf("- Batch Size: %d\n", batchSize)
	fmt.Printf("- Resource Pooling Enabled: %t\n", enableResourcePooling)
	fmt.Printf("- Pool Size: %d\n", poolSize)
	fmt.Println()

	// Create memory profiler
	profilerOptions := profiling.DefaultProfilerOptions()
	profiler, err := profiling.NewMemoryProfiler(profilerOptions)
	if err != nil {
		log.Fatalf("Failed to create memory profiler: %v", err)
	}

	// Start memory profiling
	profiler.StartAutomaticProfiling()
	defer profiler.StopAutomaticProfiling()

	// Create resource pool manager if enabled
	var poolManager *resource.ResourcePoolManager
	if enableResourcePooling {
		poolManagerConfig := resource.DefaultPoolManagerConfig()
		poolManagerConfig.DefaultPoolSize = poolSize
		poolManager = resource.NewResourcePoolManager(poolManagerConfig)
	}

	// Create execution engine
	engine := execution.NewEngine()
	engine.SetMaxConcurrent(numConcurrent)

	// Create execution optimizer if enabled
	var executionOptimizer *optimizer.ExecutionOptimizer
	if enableMemoryOptimizer || enableConcurrencyManager || enableBatchProcessing {
		optimizerConfig := optimizer.DefaultExecutionOptimizerConfig()
		optimizerConfig.EnableMemoryOptimization = enableMemoryOptimizer
		optimizerConfig.EnableConcurrencyOptimization = enableConcurrencyManager
		optimizerConfig.EnableBatchProcessing = enableBatchProcessing
		optimizerConfig.BatchSize = batchSize
		optimizerConfig.MaxConcurrentExecutions = numConcurrent

		executionOptimizer, err = optimizer.NewExecutionOptimizer(engine, optimizerConfig)
		if err != nil {
			log.Fatalf("Failed to create execution optimizer: %v", err)
		}

		// Start execution optimizer
		if err := executionOptimizer.Start(); err != nil {
			log.Fatalf("Failed to start execution optimizer: %v", err)
		}
		defer executionOptimizer.Stop()
	}

	// Create template pool if resource pooling is enabled
	var templatePool *resource.ResourcePool
	if enableResourcePooling && poolManager != nil {
		templatePool, err = poolManager.CreatePool("templates", poolSize, 
			func() (interface{}, error) {
				return generateTemplate(fmt.Sprintf("template-%d", time.Now().UnixNano()), 
					templateSize, numVariables)
			}, 
			func(obj interface{}) {
				// Cleanup template
			})
		if err != nil {
			log.Fatalf("Failed to create template pool: %v", err)
		}
	}

	// Take initial memory snapshot
	profiler.CreateSnapshot("initial")
	initialMemStats := profiler.GetFormattedMemoryStats()
	fmt.Println("Initial Memory Usage:")
	fmt.Printf("- Heap Alloc: %.2f MB\n", initialMemStats["heap_alloc_mb"].(float64))
	fmt.Printf("- Heap Sys: %.2f MB\n", initialMemStats["heap_sys_mb"].(float64))
	fmt.Printf("- Heap Objects: %d\n", initialMemStats["heap_objects"].(uint64))
	fmt.Println()

	// Generate templates or use template pool
	var templates []*format.Template
	if templatePool != nil {
		// Use template pool
		templates = make([]*format.Template, numTemplates)
		for i := 0; i < numTemplates; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			obj, err := templatePool.Acquire(ctx)
			cancel()
			if err != nil {
				log.Fatalf("Failed to acquire template from pool: %v", err)
			}
			templates[i] = obj.(*format.Template)
			defer templatePool.Release(obj)
		}
	} else {
		// Generate templates
		templates = make([]*format.Template, numTemplates)
		for i := 0; i < numTemplates; i++ {
			templates[i], err = generateTemplate(fmt.Sprintf("template-%d", i), 
				templateSize, numVariables)
			if err != nil {
				log.Fatalf("Failed to generate template: %v", err)
			}
		}
	}

	// Generate test data
	testData := generateTestData(numVariables)

	// Take post-generation memory snapshot
	profiler.CreateSnapshot("post-generation")
	postGenMemStats := profiler.GetFormattedMemoryStats()
	fmt.Println("Post-Generation Memory Usage:")
	fmt.Printf("- Heap Alloc: %.2f MB\n", postGenMemStats["heap_alloc_mb"].(float64))
	fmt.Printf("- Heap Sys: %.2f MB\n", postGenMemStats["heap_sys_mb"].(float64))
	fmt.Printf("- Heap Objects: %d\n", postGenMemStats["heap_objects"].(uint64))
	fmt.Println()

	// Execute templates
	fmt.Println("Executing templates...")
	startTime := time.Now()

	if executionOptimizer != nil {
		// Execute with optimizer
		if enableBatchProcessing {
			// Execute in batches
			results, err := executionOptimizer.ExecuteTemplates(context.Background(), templates, testData)
			if err != nil {
				log.Fatalf("Failed to execute templates: %v", err)
			}
			fmt.Printf("Executed %d templates in batches\n", len(results))
		} else {
			// Execute individually
			for i, template := range templates {
				_, err := executionOptimizer.ExecuteTemplate(context.Background(), template, testData)
				if err != nil {
					log.Fatalf("Failed to execute template %d: %v", i, err)
				}
			}
			fmt.Printf("Executed %d templates individually\n", len(templates))
		}
	} else {
		// Execute directly with engine
		for i, template := range templates {
			_, err := engine.ExecuteTemplate(context.Background(), template, testData)
			if err != nil {
				log.Fatalf("Failed to execute template %d: %v", i, err)
			}
		}
		fmt.Printf("Executed %d templates directly\n", len(templates))
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("Execution completed in %s\n", elapsedTime)
	fmt.Printf("Average time per template: %s\n", elapsedTime/time.Duration(len(templates)))
	fmt.Println()

	// Take post-execution memory snapshot
	profiler.CreateSnapshot("post-execution")
	postExecMemStats := profiler.GetFormattedMemoryStats()
	fmt.Println("Post-Execution Memory Usage:")
	fmt.Printf("- Heap Alloc: %.2f MB\n", postExecMemStats["heap_alloc_mb"].(float64))
	fmt.Printf("- Heap Sys: %.2f MB\n", postExecMemStats["heap_sys_mb"].(float64))
	fmt.Printf("- Heap Objects: %d\n", postExecMemStats["heap_objects"].(uint64))
	fmt.Println()

	// Force garbage collection
	runtime.GC()

	// Take final memory snapshot
	profiler.CreateSnapshot("final")
	finalMemStats := profiler.GetFormattedMemoryStats()
	fmt.Println("Final Memory Usage (After GC):")
	fmt.Printf("- Heap Alloc: %.2f MB\n", finalMemStats["heap_alloc_mb"].(float64))
	fmt.Printf("- Heap Sys: %.2f MB\n", finalMemStats["heap_sys_mb"].(float64))
	fmt.Printf("- Heap Objects: %d\n", finalMemStats["heap_objects"].(uint64))
	fmt.Println()

	// Print memory comparison
	fmt.Println("Memory Comparison:")
	initialHeapAlloc := initialMemStats["heap_alloc_mb"].(float64)
	finalHeapAlloc := finalMemStats["heap_alloc_mb"].(float64)
	heapAllocDiff := finalHeapAlloc - initialHeapAlloc
	heapAllocPercent := (heapAllocDiff / initialHeapAlloc) * 100
	fmt.Printf("- Initial Heap Alloc: %.2f MB\n", initialHeapAlloc)
	fmt.Printf("- Final Heap Alloc: %.2f MB\n", finalHeapAlloc)
	fmt.Printf("- Difference: %.2f MB (%.2f%%)\n", heapAllocDiff, heapAllocPercent)
	fmt.Println()

	// Print optimizer statistics if available
	if executionOptimizer != nil {
		stats := executionOptimizer.GetStats()
		fmt.Println("Execution Optimizer Statistics:")
		fmt.Printf("- Templates Executed: %d\n", stats.TemplatesExecuted)
		fmt.Printf("- Templates Optimized: %d\n", stats.TemplatesOptimized)
		fmt.Printf("- Memory Saved: %.2f MB\n", stats.MemorySaved)
		fmt.Printf("- Execution Time: %s\n", stats.ExecutionTime)
		fmt.Printf("- Average Execution Time: %s\n", stats.AverageExecutionTime)
		fmt.Printf("- Batches Processed: %d\n", stats.BatchesProcessed)
		fmt.Println()
	}

	// Print resource pool manager statistics if available
	if poolManager != nil {
		stats := poolManager.GetStats()
		fmt.Println("Resource Pool Manager Statistics:")
		fmt.Printf("- Total Pools: %d\n", stats.TotalPools)
		fmt.Printf("- Total Resources: %d\n", stats.TotalResources)
		fmt.Printf("- Total Available: %d\n", stats.TotalAvailable)
		fmt.Printf("- Total In Use: %d\n", stats.TotalInUse)
		fmt.Println()
	}

	fmt.Println("Example completed successfully")
}

// generateTemplate generates a template with random content
func generateTemplate(id string, size, numVariables int) (*format.Template, error) {
	// Create template
	template := &format.Template{
		ID:        id,
		Variables: make(map[string]interface{}),
		Info: format.TemplateInfo{
			Name:        fmt.Sprintf("Test Template %s", id),
			Description: "Generated test template",
			Version:     "1.0.0",
		},
	}

	// Generate template content as raw bytes
	templateContent := format.NewTemplateContent()

	// Add sections to reach desired size
	remainingSize := size
	sectionSize := size / 10
	if sectionSize < 100 {
		sectionSize = 100
	}

	for remainingSize > 0 {
		currentSize := min(sectionSize, remainingSize)
		content := generateRandomString(currentSize)
		
		templateContent.AddSection("text", content)
		
		remainingSize -= currentSize
	}

	// Add variables to both template content and template
	for i := 0; i < numVariables; i++ {
		varName := fmt.Sprintf("var%d", i)
		varValue := generateRandomString(20)
		templateContent.AddVariable(varName, varValue)
		template.Variables[varName] = varValue
	}

	// For this example, we'll store the content as a simple string representation
	// In a real implementation, this would be properly serialized YAML/JSON
	template.Content = []byte(fmt.Sprintf("# Template %s\n# Size: %d\n# Variables: %d\n", id, size, numVariables))

	return template, nil
}

// generateTestData generates test data for template execution
func generateTestData(numVariables int) map[string]interface{} {
	data := make(map[string]interface{})
	
	for i := 0; i < numVariables; i++ {
		data[fmt.Sprintf("var%d", i)] = generateRandomString(20)
	}
	
	return data
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// getEnvInt gets an integer from an environment variable
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return intValue
}

// getEnvBool gets a boolean from an environment variable
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	
	return boolValue
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
