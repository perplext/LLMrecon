// Package benchmark provides tools for benchmarking template operations.
package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

// BenchmarkResult contains the results of a benchmark
type BenchmarkResult struct {
	// Name is the name of the benchmark
	Name string
	// Duration is the total duration of the benchmark
	Duration time.Duration
	// OperationCount is the number of operations performed
	OperationCount int
	// OperationsPerSecond is the number of operations per second
	OperationsPerSecond float64
	// AverageLatency is the average latency per operation
	AverageLatency time.Duration
	// MemoryUsage is the memory usage in bytes
	MemoryUsage int64
	// Errors is the number of errors encountered
	Errors int
	// Details contains additional details about the benchmark
	Details map[string]interface{}
}

// TemplateLoadBenchmark benchmarks template loading
func TemplateLoadBenchmark(ctx context.Context, loader types.TemplateLoader, source string, sourceType string, iterations int) (*BenchmarkResult, error) {
	result := &BenchmarkResult{
		Name:    "TemplateLoad",
		Details: make(map[string]interface{}),
	}

	// Record start time
	startTime := time.Now()
	
	// Get initial memory stats
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Perform benchmark
	for i := 0; i < iterations; i++ {
		_, err := loader.LoadTemplates(ctx, source, sourceType)
		if err != nil {
			result.Errors++
		}
		result.OperationCount++
	}

	// Get final memory stats
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)
	result.MemoryUsage = int64(memStatsAfter.Alloc - memStatsBefore.Alloc)

	// Record end time
	endTime := time.Now()
	result.Duration = endTime.Sub(startTime)
	result.AverageLatency = result.Duration / time.Duration(result.OperationCount)
	result.OperationsPerSecond = float64(result.OperationCount) / result.Duration.Seconds()

	// Add details
	result.Details["source"] = source
	result.Details["sourceType"] = sourceType
	result.Details["iterations"] = iterations

	return result, nil
}

// TemplateExecuteBenchmark benchmarks template execution
func TemplateExecuteBenchmark(ctx context.Context, executor interfaces.TemplateExecutor, templates []*format.Template, options map[string]interface{}, iterations int) (*BenchmarkResult, error) {
	result := &BenchmarkResult{
		Name:    "TemplateExecute",
		Details: make(map[string]interface{}),
	}

	// Record start time
	startTime := time.Now()
	
	// Get initial memory stats
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Perform benchmark
	for i := 0; i < iterations; i++ {
		_, err := executor.ExecuteBatch(ctx, templates, options)
		if err != nil {
			result.Errors++
		}
		result.OperationCount += len(templates)
	}
	
	// Get final memory stats
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)
	result.MemoryUsage = int64(memStatsAfter.Alloc - memStatsBefore.Alloc)

	// Record end time
	endTime := time.Now()
	result.Duration = endTime.Sub(startTime)
	result.AverageLatency = result.Duration / time.Duration(result.OperationCount)
	result.OperationsPerSecond = float64(result.OperationCount) / result.Duration.Seconds()

	// Add details
	result.Details["templateCount"] = len(templates)
	result.Details["iterations"] = iterations

	return result, nil
}

// RunBenchmarkSuite runs a suite of benchmarks and returns the results
func RunBenchmarkSuite(ctx context.Context, manager types.TemplateManager, sources []types.TemplateSource, options map[string]interface{}) (map[string]*BenchmarkResult, error) {
	results := make(map[string]*BenchmarkResult)

	// Load templates benchmark
	for _, source := range sources {
		benchName := fmt.Sprintf("Load_%s_%s", source.Type, source.Path)
		// Skip load benchmarks for now - interface mismatch
		// TODO: Fix when LoadTemplate method is added to TemplateLoader interface
		_ = benchName
		continue
	}

	// Load templates for execution benchmark
	templates := make([]*format.Template, 0)
	for _, source := range sources {
		// Use LoadBatch for directory sources
		if source.Type == "directory" {
			loadedTemplates, err := manager.(interfaces.TemplateManagerInternal).GetLoader().LoadBatch(source.Path)
			if err != nil {
				continue
			}
			templates = append(templates, loadedTemplates...)
		}
	}

	// Execute templates benchmark
	if len(templates) > 0 {
		result, err := TemplateExecuteBenchmark(ctx, manager.(interfaces.TemplateManagerInternal).GetExecutor(), templates, options, 3)
		if err != nil {
			return results, err
		}
		results["Execute"] = result
	}

	return results, nil
}

// PrintBenchmarkResults prints benchmark results in a human-readable format
func PrintBenchmarkResults(results map[string]*BenchmarkResult) {
	fmt.Println("=== Benchmark Results ===")
	for name, result := range results {
		fmt.Printf("\n%s:\n", name)
		fmt.Printf("  Duration: %v\n", result.Duration)
		fmt.Printf("  Operations: %d\n", result.OperationCount)
		fmt.Printf("  Operations/sec: %.2f\n", result.OperationsPerSecond)
		fmt.Printf("  Avg Latency: %v\n", result.AverageLatency)
		fmt.Printf("  Memory Usage: %.2f MB\n", float64(result.MemoryUsage)/(1024*1024))
		fmt.Printf("  Errors: %d\n", result.Errors)
		fmt.Println("  Details:")
		for k, v := range result.Details {
			fmt.Printf("    %s: %v\n", k, v)
		}
	}
}
