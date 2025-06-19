package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/loader"
)

func main() {
	// Command line flags
	sourcePtr := flag.String("source", "./templates", "Source path for templates")
	sourceTypePtr := flag.String("type", "file", "Source type (file, github, gitlab)")
	cacheSizePtr := flag.Int("cache-size", 1000, "Maximum cache size")
	cacheTTLPtr := flag.Int("cache-ttl", 3600, "Cache TTL in seconds")
	concurrencyPtr := flag.Int("concurrency", runtime.NumCPU(), "Concurrency limit")
	optimizePtr := flag.Bool("optimize", true, "Enable template optimization")
	structureOptimizePtr := flag.Bool("structure-optimize", true, "Enable template structure optimization")
	compressPtr := flag.Bool("compress", false, "Enable template compression")
	minifyPtr := flag.Bool("minify", true, "Enable template minification")
	outputDirPtr := flag.String("output", "", "Output directory for optimized templates")
	verbosePtr := flag.Bool("verbose", false, "Enable verbose output")
	benchmarkPtr := flag.Bool("benchmark", false, "Run benchmark")
	benchmarkIterationsPtr := flag.Int("iterations", 5, "Number of benchmark iterations")
	
	flag.Parse()

	// Create repository manager
	repoManager := repository.NewManager()

	// Create loader options
	options := loader.ResourceEfficientLoaderOptions{
		CacheTTL:                  time.Duration(*cacheTTLPtr) * time.Second,
		MaxCacheSize:              *cacheSizePtr,
		ConcurrencyLimit:          *concurrencyPtr,
		EnableOptimization:        *optimizePtr,
		EnableStructureOptimization: *structureOptimizePtr,
		EnableCompression:         *compressPtr,
		EnableMinification:        *minifyPtr,
	}

	// Create resource-efficient loader
	efficientLoader := loader.NewResourceEfficientLoader(repoManager, options)

	// Create context
	ctx := context.Background()

	if *benchmarkPtr {
		// Run benchmark
		fmt.Println("Running template loading benchmark...")
		runBenchmark(ctx, efficientLoader, *sourcePtr, *sourceTypePtr, *benchmarkIterationsPtr, *verbosePtr)
	} else {
		// Load templates
		fmt.Printf("Loading templates from %s (%s)...\n", *sourcePtr, *sourceTypePtr)
		startTime := time.Now()
		templates, err := efficientLoader.LoadTemplates(ctx, *sourcePtr, *sourceTypePtr)
		if err != nil {
			fmt.Printf("Error loading templates: %v\n", err)
			os.Exit(1)
		}
		loadTime := time.Since(startTime)

		fmt.Printf("Loaded %d templates in %v\n", len(templates), loadTime)

		// Print stats
		stats := efficientLoader.GetLoaderStats()
		fmt.Println("\nLoader Statistics:")
		fmt.Printf("  Total Loads:     %d\n", stats["total_loads"])
		fmt.Printf("  Cache Hits:      %d\n", stats["cache_hits"])
		fmt.Printf("  Cache Misses:    %d\n", stats["cache_misses"])
		fmt.Printf("  Load Errors:     %d\n", stats["load_errors"])
		fmt.Printf("  Avg Load Time:   %v\n", stats["avg_load_time"])
		fmt.Printf("  Indexed Sources: %d\n", stats["indexed_sources"])

		// Print optimizer stats if optimization is enabled
		if *optimizePtr {
			optimizerStats := stats["optimizer_stats"].(map[string]interface{})
			fmt.Println("\nOptimizer Statistics:")
			fmt.Printf("  Total Optimizations:   %d\n", optimizerStats["total_optimizations"])
			fmt.Printf("  Original Size:         %d bytes\n", optimizerStats["total_bytes_original"])
			fmt.Printf("  Optimized Size:        %d bytes\n", optimizerStats["total_bytes_optimized"])
			fmt.Printf("  Compression Ratio:     %.2f\n", optimizerStats["compression_ratio"])
		}

		// Print structure optimizer stats if structure optimization is enabled
		if *structureOptimizePtr {
			structureStats := stats["structure_stats"].(map[string]interface{})
			fmt.Println("\nStructure Optimizer Statistics:")
			fmt.Printf("  Total Optimizations:   %d\n", structureStats["total_optimizations"])
			
			if *verbosePtr && structureStats["optimizations_by_rule"] != nil {
				ruleStats := structureStats["optimizations_by_rule"].(map[string]interface{})
				fmt.Println("  Optimizations by Rule:")
				for rule, count := range ruleStats {
					fmt.Printf("    %s: %d\n", rule, count)
				}
			}
		}

		// Save optimized templates if output directory is specified
		if *outputDirPtr != "" {
			saveOptimizedTemplates(templates, *outputDirPtr, *verbosePtr)
		}
	}
}

// runBenchmark runs a benchmark of template loading
func runBenchmark(ctx context.Context, loader *loader.ResourceEfficientLoader, source string, sourceType string, iterations int, verbose bool) {
	// Clear cache before benchmark
	loader.ClearCache()
	loader.ClearAllSourceIndices()

	// Run benchmark iterations
	var totalLoadTime time.Duration
	var totalTemplates int

	fmt.Printf("Running %d iterations...\n", iterations)
	for i := 0; i < iterations; i++ {
		// Clear cache between iterations
		loader.ClearCache()
		
		fmt.Printf("Iteration %d/%d: ", i+1, iterations)
		
		// Load templates
		startTime := time.Now()
		templates, err := loader.LoadTemplates(ctx, source, sourceType)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		loadTime := time.Since(startTime)
		
		totalLoadTime += loadTime
		totalTemplates += len(templates)
		
		fmt.Printf("Loaded %d templates in %v\n", len(templates), loadTime)
	}

	// Calculate average
	avgLoadTime := totalLoadTime / time.Duration(iterations)
	avgTemplatesPerSecond := float64(totalTemplates) / totalLoadTime.Seconds() * float64(iterations)

	fmt.Println("\nBenchmark Results:")
	fmt.Printf("  Total Templates:        %d\n", totalTemplates/iterations)
	fmt.Printf("  Average Load Time:      %v\n", avgLoadTime)
	fmt.Printf("  Templates per Second:   %.2f\n", avgTemplatesPerSecond)

	// Print loader stats
	if verbose {
		stats := loader.GetLoaderStats()
		fmt.Println("\nLoader Statistics:")
		fmt.Printf("  Total Loads:     %d\n", stats["total_loads"])
		fmt.Printf("  Cache Hits:      %d\n", stats["cache_hits"])
		fmt.Printf("  Cache Misses:    %d\n", stats["cache_misses"])
		fmt.Printf("  Load Errors:     %d\n", stats["load_errors"])
	}
}

// saveOptimizedTemplates saves optimized templates to the specified directory
func saveOptimizedTemplates(templates []*format.Template, outputDir string, verbose bool) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	fmt.Printf("Saving %d optimized templates to %s...\n", len(templates), outputDir)

	// Save templates
	for _, template := range templates {
		// Create filename
		filename := template.ID
		if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") && !strings.HasSuffix(filename, ".json") {
			filename += ".yaml"
		}
		
		// Create file path
		filePath := filepath.Join(outputDir, filename)
		
		// Save template
		if err := template.SaveToFile(filePath); err != nil {
			fmt.Printf("Error saving template %s: %v\n", template.ID, err)
			continue
		}
		
		if verbose {
			fmt.Printf("Saved template %s to %s\n", template.ID, filePath)
		}
	}

	fmt.Printf("Saved %d templates to %s\n", len(templates), outputDir)
}
