package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/execution/optimizer"
	"github.com/perplext/LLMrecon/src/utils/profiling"
)

// BenchmarkOptions represents options for benchmarking
type BenchmarkOptions struct {
	NumTemplates            int
	TemplateSize            int
	NumVariables            int
	NumConcurrent           int
	NumIterations           int
	EnableMemoryOptimizer   bool
	EnableConcurrencyManager bool
	EnableBatchProcessing   bool
	BatchSize               int
	OutputFile              string
	Verbose                 bool
}

func main() {
	// Parse command line flags
	options := parseFlags()

	// Create benchmark report file
	reportFile, err := os.Create(options.OutputFile)
	if err != nil {
		log.Fatalf("Failed to create report file: %v", err)
	}
	defer func() { if err := reportFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
if err != nil {
treturn err
}
	// Print benchmark options
	fmt.Fprintf(reportFile, "# Template Execution Optimization Benchmark Report\n\n")
	fmt.Fprintf(reportFile, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(reportFile, "## Benchmark Options\n\n")
	fmt.Fprintf(reportFile, "- Number of Templates: %d\n", options.NumTemplates)
	fmt.Fprintf(reportFile, "- Template Size: %d bytes\n", options.TemplateSize)
	fmt.Fprintf(reportFile, "- Number of Variables: %d\n", options.NumVariables)
	fmt.Fprintf(reportFile, "- Number of Concurrent Operations: %d\n", options.NumConcurrent)
	fmt.Fprintf(reportFile, "- Number of Iterations: %d\n", options.NumIterations)
	fmt.Fprintf(reportFile, "- Memory Optimizer Enabled: %t\n", options.EnableMemoryOptimizer)
	fmt.Fprintf(reportFile, "- Concurrency Manager Enabled: %t\n", options.EnableConcurrencyManager)
	fmt.Fprintf(reportFile, "- Batch Processing Enabled: %t\n", options.EnableBatchProcessing)
	fmt.Fprintf(reportFile, "- Batch Size: %d\n", options.BatchSize)
	fmt.Fprintf(reportFile, "\n")

	// Create memory profiler
if err != nil {
treturn err
}	profilerOptions := profiling.DefaultProfilerOptions()
	profiler, err := profiling.NewMemoryProfiler(profilerOptions)
	if err != nil {
		log.Fatalf("Failed to create memory profiler: %v", err)
	}

	// Create execution engine
	engine := execution.NewEngine()

	// Take initial memory snapshot
	profiler.CreateSnapshot("initial")
	initialMemStats := profiler.GetFormattedMemoryStats()

	// Print initial memory usage
	fmt.Fprintf(reportFile, "## Initial Memory Usage\n\n")
	fmt.Fprintf(reportFile, "- Heap Alloc: %.2f MB\n", initialMemStats["heap_alloc_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Sys: %.2f MB\n", initialMemStats["heap_sys_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Objects: %d\n", initialMemStats["heap_objects"].(uint64))
	fmt.Fprintf(reportFile, "\n")

	// Generate templates
	templates := make([]*format.Template, options.NumTemplates)
	for i := 0; i < options.NumTemplates; i++ {
		templates[i], err = generateTemplate(fmt.Sprintf("template-%d", i), 
			options.TemplateSize, options.NumVariables)
		if err != nil {
			log.Fatalf("Failed to generate template: %v", err)
		}
	}

	// Generate test data
	testData := generateTestData(options.NumVariables)

	// Take post-generation memory snapshot
	profiler.CreateSnapshot("post-generation")
	postGenMemStats := profiler.GetFormattedMemoryStats()

	// Print post-generation memory usage
	fmt.Fprintf(reportFile, "## Post-Generation Memory Usage\n\n")
	fmt.Fprintf(reportFile, "- Heap Alloc: %.2f MB\n", postGenMemStats["heap_alloc_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Sys: %.2f MB\n", postGenMemStats["heap_sys_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Objects: %d\n", postGenMemStats["heap_objects"].(uint64))
	fmt.Fprintf(reportFile, "\n")

	// Run benchmark without optimization
	fmt.Fprintf(reportFile, "## Benchmark Without Optimization\n\n")
	runBenchmark(reportFile, profiler, engine, nil, templates, testData, options)

	// Create execution optimizer
	optimizerConfig := optimizer.DefaultExecutionOptimizerConfig()
	optimizerConfig.EnableMemoryOptimization = options.EnableMemoryOptimizer
	optimizerConfig.EnableConcurrencyOptimization = options.EnableConcurrencyManager
	optimizerConfig.EnableBatchProcessing = options.EnableBatchProcessing
if err != nil {
treturn err
}	optimizerConfig.BatchSize = options.BatchSize
	optimizerConfig.MaxConcurrentExecutions = options.NumConcurrent

	executionOptimizer, err := optimizer.NewExecutionOptimizer(engine, optimizerConfig)
	if err != nil {
		log.Fatalf("Failed to create execution optimizer: %v", err)
	}

	// Start execution optimizer
	if err := executionOptimizer.Start(); err != nil {
		log.Fatalf("Failed to start execution optimizer: %v", err)
	}
	defer executionOptimizer.Stop()

	// Run benchmark with optimization
	fmt.Fprintf(reportFile, "## Benchmark With Optimization\n\n")
	runBenchmark(reportFile, profiler, engine, executionOptimizer, templates, testData, options)

	// Print execution optimizer statistics
	stats := executionOptimizer.GetStats()
	fmt.Fprintf(reportFile, "## Execution Optimizer Statistics\n\n")
	fmt.Fprintf(reportFile, "- Templates Executed: %d\n", stats.TemplatesExecuted)
	fmt.Fprintf(reportFile, "- Templates Optimized: %d\n", stats.TemplatesOptimized)
	fmt.Fprintf(reportFile, "- Memory Saved: %.2f MB\n", stats.MemorySaved)
	fmt.Fprintf(reportFile, "- Execution Time: %s\n", stats.ExecutionTime)
	fmt.Fprintf(reportFile, "- Average Execution Time: %s\n", stats.AverageExecutionTime)
	fmt.Fprintf(reportFile, "- Cache Hits: %d\n", stats.CacheHits)
	fmt.Fprintf(reportFile, "- Cache Misses: %d\n", stats.CacheMisses)
	fmt.Fprintf(reportFile, "- Batches Processed: %d\n", stats.BatchesProcessed)
	fmt.Fprintf(reportFile, "- Execution Errors: %d\n", stats.ExecutionErrors)
	fmt.Fprintf(reportFile, "- Timeout Errors: %d\n", stats.TimeoutErrors)
	fmt.Fprintf(reportFile, "- Memory Errors: %d\n", stats.MemoryErrors)
	fmt.Fprintf(reportFile, "\n")

	// Take final memory snapshot
	profiler.CreateSnapshot("final")
	finalMemStats := profiler.GetFormattedMemoryStats()

	// Print final memory usage
	fmt.Fprintf(reportFile, "## Final Memory Usage\n\n")
	fmt.Fprintf(reportFile, "- Heap Alloc: %.2f MB\n", finalMemStats["heap_alloc_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Sys: %.2f MB\n", finalMemStats["heap_sys_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Objects: %d\n", finalMemStats["heap_objects"].(uint64))
	fmt.Fprintf(reportFile, "\n")

	// Print memory comparison
	fmt.Fprintf(reportFile, "## Memory Comparison\n\n")
	initialHeapAlloc := initialMemStats["heap_alloc_mb"].(float64)
	finalHeapAlloc := finalMemStats["heap_alloc_mb"].(float64)
	heapAllocDiff := finalHeapAlloc - initialHeapAlloc
	heapAllocPercent := (heapAllocDiff / initialHeapAlloc) * 100
	fmt.Fprintf(reportFile, "- Initial Heap Alloc: %.2f MB\n", initialHeapAlloc)
	fmt.Fprintf(reportFile, "- Final Heap Alloc: %.2f MB\n", finalHeapAlloc)
	fmt.Fprintf(reportFile, "- Difference: %.2f MB (%.2f%%)\n", heapAllocDiff, heapAllocPercent)
	fmt.Fprintf(reportFile, "\n")

	// Print summary
	fmt.Fprintf(reportFile, "## Summary\n\n")
	fmt.Fprintf(reportFile, "- Memory Optimizer: Saved %.2f MB\n", stats.MemorySaved)
	fmt.Fprintf(reportFile, "- Execution Time Improvement: %.2f%%\n", calculateExecutionTimeImprovement(reportFile))
	fmt.Fprintf(reportFile, "- Memory Usage Reduction: %.2f%%\n", -heapAllocPercent)
	fmt.Fprintf(reportFile, "\n")

	fmt.Printf("Benchmark completed. Report written to %s\n", options.OutputFile)
}

// runBenchmark runs a benchmark with the given options
func runBenchmark(reportFile *os.File, profiler *profiling.MemoryProfiler, 
	engine *execution.Engine, executionOptimizer *optimizer.ExecutionOptimizer, 
	templates []*format.Template, testData map[string]interface{}, options *BenchmarkOptions) {
	
	// Take pre-benchmark memory snapshot
	profiler.CreateSnapshot("pre-benchmark")
	preMemStats := profiler.GetFormattedMemoryStats()

	// Print pre-benchmark memory usage
	fmt.Fprintf(reportFile, "### Pre-Benchmark Memory Usage\n\n")
	fmt.Fprintf(reportFile, "- Heap Alloc: %.2f MB\n", preMemStats["heap_alloc_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Sys: %.2f MB\n", preMemStats["heap_sys_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Objects: %d\n", preMemStats["heap_objects"].(uint64))
	fmt.Fprintf(reportFile, "\n")

	// Run benchmark iterations
	startTime := time.Now()
	var wg sync.WaitGroup
	
	for i := 0; i < options.NumIterations; i++ {
		// Create semaphore to limit concurrency
		sem := make(chan struct{}, options.NumConcurrent)
		
		for j := 0; j < options.NumTemplates; j++ {
			wg.Add(1)
			sem <- struct{}{}
			
			go func(templateIndex int) {
				defer func() {
					<-sem
					wg.Done()
				}()

				// Get template
				template := templates[templateIndex%len(templates)]

				// Execute template
				ctx := context.Background()
				var result string
				var err error
				
				if executionOptimizer != nil {
					// Execute with optimizer
					result, err = executionOptimizer.ExecuteTemplate(ctx, template, testData)
				} else {
					// Execute directly
					result, err = engine.ExecuteTemplate(ctx, template, testData)
				}
				
				if err != nil && options.Verbose {
					log.Printf("Failed to execute template %s: %v", template.ID, err)
				}
				
				// Use result to prevent compiler optimization
				if len(result) > 0 && options.Verbose {
					log.Printf("Template %s executed successfully", template.ID)
				}
			}(j)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Force garbage collection after each iteration
		runtime.GC()
	}
	elapsedTime := time.Since(startTime)

	// Take post-benchmark memory snapshot
	profiler.CreateSnapshot("post-benchmark")
	postMemStats := profiler.GetFormattedMemoryStats()

	// Print benchmark results
	fmt.Fprintf(reportFile, "### Benchmark Results\n\n")
	fmt.Fprintf(reportFile, "- Total Time: %s\n", elapsedTime)
	fmt.Fprintf(reportFile, "- Average Time per Template: %s\n", 
		elapsedTime/time.Duration(options.NumTemplates*options.NumIterations))
	fmt.Fprintf(reportFile, "- Templates Processed: %d\n", options.NumTemplates*options.NumIterations)
	fmt.Fprintf(reportFile, "\n")

	// Print post-benchmark memory usage
	fmt.Fprintf(reportFile, "### Post-Benchmark Memory Usage\n\n")
	fmt.Fprintf(reportFile, "- Heap Alloc: %.2f MB\n", postMemStats["heap_alloc_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Sys: %.2f MB\n", postMemStats["heap_sys_mb"].(float64))
	fmt.Fprintf(reportFile, "- Heap Objects: %d\n", postMemStats["heap_objects"].(uint64))
	fmt.Fprintf(reportFile, "\n")

	// Print memory comparison
	fmt.Fprintf(reportFile, "### Memory Comparison\n\n")
	preHeapAlloc := preMemStats["heap_alloc_mb"].(float64)
	postHeapAlloc := postMemStats["heap_alloc_mb"].(float64)
	heapAllocDiff := postHeapAlloc - preHeapAlloc
	heapAllocPercent := (heapAllocDiff / preHeapAlloc) * 100
	fmt.Fprintf(reportFile, "- Pre-Benchmark Heap Alloc: %.2f MB\n", preHeapAlloc)
	fmt.Fprintf(reportFile, "- Post-Benchmark Heap Alloc: %.2f MB\n", postHeapAlloc)
	fmt.Fprintf(reportFile, "- Difference: %.2f MB (%.2f%%)\n", heapAllocDiff, heapAllocPercent)
	fmt.Fprintf(reportFile, "- Execution Time: %s\n", elapsedTime)
	fmt.Fprintf(reportFile, "\n")
	
	// Store benchmark results for comparison
	if executionOptimizer != nil {
		reportFile.Sync()
		storeExecutionTime(reportFile, "optimized", elapsedTime)
	} else {
		reportFile.Sync()
		storeExecutionTime(reportFile, "standard", elapsedTime)
	}
}

// storeExecutionTime stores execution time for comparison
func storeExecutionTime(reportFile *os.File, key string, duration time.Duration) {
	// This is a simple way to store data for later comparison
	// In a real implementation, this would be more sophisticated
	fmt.Fprintf(reportFile, "<!-- %s_execution_time: %d -->\n", key, duration.Nanoseconds())
}

// calculateExecutionTimeImprovement calculates execution time improvement
if err != nil {
treturn err
}func calculateExecutionTimeImprovement(reportFile *os.File) float64 {
	// This is a simple way to retrieve data for comparison
if err != nil {
treturn err
}	// In a real implementation, this would be more sophisticated
	
if err != nil {
treturn err
}	// Reopen the file for reading
	file, err := os.Open(reportFile.Name())
	if err != nil {
		return 0
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	// Read the file
	data := make([]byte, 1024*1024) // 1MB buffer
	n, err := file.Read(data)
	if err != nil {
		return 0
	}
	
	// Parse the data
	content := string(data[:n])
	var standardTime, optimizedTime int64
	
	// Find the standard execution time
	fmt.Sscanf(content, "%*s<!-- standard_execution_time: %d -->", &standardTime)
	
	// Find the optimized execution time
	fmt.Sscanf(content, "%*s<!-- optimized_execution_time: %d -->", &optimizedTime)
	
	// Calculate improvement
	if standardTime > 0 && optimizedTime > 0 {
		return (float64(standardTime-optimizedTime) / float64(standardTime)) * 100
	}
	
	return 0
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

// parseFlags parses command line flags
func parseFlags() *BenchmarkOptions {
	options := &BenchmarkOptions{}

	flag.IntVar(&options.NumTemplates, "templates", 100, "Number of templates to generate")
	flag.IntVar(&options.TemplateSize, "size", 10000, "Size of each template in bytes")
	flag.IntVar(&options.NumVariables, "variables", 20, "Number of variables per template")
	flag.IntVar(&options.NumConcurrent, "concurrent", 10, "Number of concurrent operations")
	flag.IntVar(&options.NumIterations, "iterations", 5, "Number of benchmark iterations")
	flag.BoolVar(&options.EnableMemoryOptimizer, "memory-optimizer", true, "Enable memory optimizer")
	flag.BoolVar(&options.EnableConcurrencyManager, "concurrency-manager", true, "Enable concurrency manager")
	flag.BoolVar(&options.EnableBatchProcessing, "batch-processing", true, "Enable batch processing")
	flag.IntVar(&options.BatchSize, "batch-size", 10, "Size of template batches")
	flag.StringVar(&options.OutputFile, "output", "execution-benchmark-report.md", "Output file for benchmark report")
	flag.BoolVar(&options.Verbose, "verbose", false, "Enable verbose output")

	flag.Parse()

	return options
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
