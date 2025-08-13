// Package main provides a CLI tool for benchmarking template operations.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/management"
	"github.com/perplext/LLMrecon/src/template/management/benchmark"
	"github.com/perplext/LLMrecon/src/template/management/monitoring"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

// Command line flags
var (
	sourceFlag       = flag.String("source", "", "Template source path or URL")
	sourceTypeFlag   = flag.String("type", "file", "Template source type (file, github, gitlab)")
	iterationsFlag   = flag.Int("iterations", 5, "Number of benchmark iterations")
	concurrencyFlag  = flag.Int("concurrency", 10, "Maximum concurrent operations")
	cacheSizeFlag    = flag.Int("cache-size", 1000, "Maximum cache size")
	cacheTTLFlag     = flag.Int("cache-ttl", 3600, "Cache TTL in seconds")
	outputFileFlag   = flag.String("output", "", "Output file for benchmark results (JSON format)")
	compareFlag      = flag.String("compare", "", "Previous benchmark results file to compare with")
	optimizedFlag    = flag.Bool("optimized", true, "Use optimized components")
	verboseFlag      = flag.Bool("verbose", false, "Enable verbose output")
	monitorFlag      = flag.Bool("monitor", false, "Enable performance monitoring")
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Validate flags
	if *sourceFlag == "" {
		fmt.Println("Error: source flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create context
	ctx := context.Background()

	// Create performance monitor if enabled
	var monitor *monitoring.PerformanceMonitor
	if *monitorFlag {
		monitor = monitoring.NewPerformanceMonitor(&monitoring.MonitorConfig{
			HistorySize:      100,
			SamplingInterval: 1 * time.Second,
			EnableAlerts:     false,
		})
	}

	// Run benchmark
	fmt.Println("Running benchmark...")
	results, err := runBenchmark(ctx, monitor)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print benchmark results
	benchmark.PrintBenchmarkResults(results)

	// Save results to file if specified
	if *outputFileFlag != "" {
		if err := saveResults(results, *outputFileFlag); err != nil {
			fmt.Printf("Error saving results: %v\n", err)
		} else {
			fmt.Printf("Results saved to %s\n", *outputFileFlag)
		}
	}

	// Compare with previous results if specified
	if *compareFlag != "" {
		previousResults, err := loadResults(*compareFlag)
		if err != nil {
			fmt.Printf("Error loading previous results: %v\n", err)
		} else {
			compareResults(results, previousResults)
		}
	}

	// Print performance monitor report if enabled
	if *monitorFlag {
		fmt.Println("\n=== Performance Monitor Report ===")
		report := monitor.GetReport()
		reportJSON, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(reportJSON))
	}
}

// runBenchmark runs the benchmark
func runBenchmark(ctx context.Context, monitor *monitoring.PerformanceMonitor) (map[string]*benchmark.BenchmarkResult, error) {
	// Create template manager
	var manager types.TemplateManager
	var err error

	if *optimizedFlag {
		// Create optimized template manager
		config := &management.ManagerConfig{
			CacheTTL:         time.Duration(*cacheTTLFlag) * time.Second,
			MaxCacheSize:     *cacheSizeFlag,
			ConcurrencyLimit: *concurrencyFlag,
			ExecutionTimeout: 30 * time.Second,
			LoadTimeout:      30 * time.Second,
			Debug:            *verboseFlag,
		}
		manager, err = management.NewOptimizedTemplateManager(config)
	} else {
		// Create standard template manager
		manager, err = management.NewTemplateManager()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create template manager: %w", err)
	}

	// Create template source
	source := types.TemplateSource{
		Path: *sourceFlag,
		Type: *sourceTypeFlag,
	}

	// Run benchmark suite
	results, err := benchmark.RunBenchmarkSuite(ctx, manager, []types.TemplateSource{source}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to run benchmark suite: %w", err)
	}

	// Record benchmark results in monitor if enabled
	if monitor != nil {
		monitor.RecordBenchmarkResults(results)
	}

	return results, nil
}

// saveResults saves benchmark results to a file
func saveResults(results map[string]*benchmark.BenchmarkResult, filePath string) error {
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Encode results as JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("failed to encode results: %w", err)
	}

	return nil
}

// loadResults loads benchmark results from a file
func loadResults(filePath string) (map[string]*benchmark.BenchmarkResult, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode results from JSON
	var results map[string]*benchmark.BenchmarkResult
	if err := json.NewDecoder(file).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

// compareResults compares benchmark results
func compareResults(current, previous map[string]*benchmark.BenchmarkResult) {
	fmt.Println("\n=== Benchmark Comparison ===")

	// Compare each benchmark
	for name, currentResult := range current {
		previousResult, exists := previous[name]
		if !exists {
			fmt.Printf("%s: No previous result\n", name)
			continue
		}

		// Calculate percentage differences
		durationDiff := calculatePercentageDiff(currentResult.Duration.Seconds(), previousResult.Duration.Seconds())
		opsDiff := calculatePercentageDiff(currentResult.OperationsPerSecond, previousResult.OperationsPerSecond)
		latencyDiff := calculatePercentageDiff(currentResult.AverageLatency.Seconds(), previousResult.AverageLatency.Seconds())
		memoryDiff := calculatePercentageDiff(float64(currentResult.MemoryUsage), float64(previousResult.MemoryUsage))

		fmt.Printf("\n%s:\n", name)
		fmt.Printf("  Duration: %v -> %v (%+.2f%%)\n", previousResult.Duration, currentResult.Duration, durationDiff)
		fmt.Printf("  Operations/sec: %.2f -> %.2f (%+.2f%%)\n", previousResult.OperationsPerSecond, currentResult.OperationsPerSecond, opsDiff)
		fmt.Printf("  Avg Latency: %v -> %v (%+.2f%%)\n", previousResult.AverageLatency, currentResult.AverageLatency, latencyDiff)
		fmt.Printf("  Memory Usage: %.2f MB -> %.2f MB (%+.2f%%)\n", 
			float64(previousResult.MemoryUsage)/(1024*1024), 
			float64(currentResult.MemoryUsage)/(1024*1024), 
			memoryDiff)
	}

	// Check for benchmarks that exist only in previous results
	for name := range previous {
		if _, exists := current[name]; !exists {
			fmt.Printf("%s: Only in previous results\n", name)
		}
	}
}

// calculatePercentageDiff calculates the percentage difference between two values
func calculatePercentageDiff(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return ((current - previous) / previous) * 100
}
