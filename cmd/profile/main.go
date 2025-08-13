// Package main provides a CLI tool for profiling template operations.
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/perplext/LLMrecon/src/profiling"
	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/management"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

// Command line flags
var (
	sourceFlag       = flag.String("source", "", "Template source path or URL")
	sourceTypeFlag   = flag.String("type", "file", "Template source type (file, github, gitlab)")
	iterationsFlag   = flag.Int("iterations", 5, "Number of profiling iterations")
	concurrencyFlag  = flag.Int("concurrency", 10, "Maximum concurrent operations")
	cacheSizeFlag    = flag.Int("cache-size", 1000, "Maximum cache size")
	cacheTTLFlag     = flag.Int("cache-ttl", 3600, "Cache TTL in seconds")
	baselineFlag     = flag.Bool("baseline", false, "Establish baseline metrics")
	compareFlag      = flag.Bool("compare", false, "Compare with baseline metrics")
	reportFileFlag   = flag.String("report", "profile_report.txt", "Output file for profiling report")
	baselineFileFlag = flag.String("baseline-file", "template_baseline.txt", "File for baseline metrics")
	cpuProfileFlag   = flag.Bool("cpu-profile", false, "Enable CPU profiling")
	memProfileFlag   = flag.Bool("mem-profile", false, "Enable memory profiling")
	optimizedFlag    = flag.Bool("optimized", true, "Use optimized components")
	verboseFlag      = flag.Bool("verbose", false, "Enable verbose output")
	monitorFlag      = flag.Bool("monitor", false, "Enable continuous monitoring")
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

	// Create template manager
	templateManager, err := createTemplateManager()
	if err != nil {
		fmt.Printf("Error creating template manager: %v\n", err)
		os.Exit(1)
	}

	// Create profiler configuration
	profilerConfig := &profiling.ProfilerConfig{
		EnableCPUProfiling: *cpuProfileFlag,
		EnableMemProfiling: *memProfileFlag,
		CPUProfilePath:     "cpu.pprof",
		MemProfilePath:     "mem.pprof",
		SamplingInterval:   1 * time.Second,
		MaxSamples:         1000,
		Tags: map[string]string{
			"source":      *sourceFlag,
			"source_type": *sourceTypeFlag,
		},
	}

	// Create template profiler configuration
	templateProfilerConfig := &profiling.TemplateProfilerConfig{
		ProfilerConfig:           profilerConfig,
		EnableDetailedProfiling:  *verboseFlag,
		EnableContinuousMonitoring: *monitorFlag,
		MonitoringInterval:       5 * time.Minute,
		BaselineFilePath:         *baselineFileFlag,
		ReportFilePath:           *reportFileFlag,
	}

	// Create template profiler
	templateProfiler := profiling.NewTemplateProfiler(templateManager, templateProfilerConfig)

	// Start profiler
	if err := templateProfiler.Start(); err != nil {
		fmt.Printf("Error starting profiler: %v\n", err)
		os.Exit(1)
	}
	defer templateProfiler.Stop()

	// Create template source
	source := types.TemplateSource{
		Path: *sourceFlag,
		Type: *sourceTypeFlag,
	}

	// Establish baseline if requested
	if *baselineFlag {
		if err := templateProfiler.EstablishBaseline(ctx, []types.TemplateSource{source}, *iterationsFlag); err != nil {
			fmt.Printf("Error establishing baseline: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Baseline metrics saved to %s\n", *baselineFileFlag)
		return
	}

	// Run profiling
	fmt.Println("Running template profiling...")

	// Load templates
	startTime := time.Now()
	templates, err := templateProfiler.ProfileTemplateLoadBatch(ctx, source.Path, source.Type)
	if err != nil {
		fmt.Printf("Error loading templates: %v\n", err)
		os.Exit(1)
	}
	loadTime := time.Since(startTime)

	fmt.Printf("Loaded %d templates in %v\n", len(templates), loadTime)

	// Execute templates
	if len(templates) > 0 {
		// Limit to a reasonable number of templates for execution
		execTemplates := templates
		if len(templates) > 10 {
			execTemplates = templates[:10]
		}

		startTime = time.Now()
		results, err := templateProfiler.ProfileTemplateExecutionBatch(ctx, execTemplates, nil)
		if err != nil {
			fmt.Printf("Error executing templates: %v\n", err)
		} else {
			execTime := time.Since(startTime)
			fmt.Printf("Executed %d templates in %v (%.2f ms/template)\n",
				len(execTemplates), execTime, float64(execTime.Milliseconds())/float64(len(execTemplates)))
			fmt.Printf("Generated %d results\n", len(results))
		}
	}

	// Compare with baseline if requested
	if *compareFlag {
		comparisonFile := "comparison_report.txt"
		if err := templateProfiler.SaveComparisonReport(comparisonFile); err != nil {
			fmt.Printf("Error saving comparison report: %v\n", err)
		} else {
			fmt.Printf("Comparison report saved to %s\n", comparisonFile)
		}
	}

	// Save profiling report
	if err := templateProfiler.profiler.SaveReport(*reportFileFlag); err != nil {
		fmt.Printf("Error saving profiling report: %v\n", err)
	} else {
		fmt.Printf("Profiling report saved to %s\n", *reportFileFlag)
	}
}

// createTemplateManager creates a template manager
func createTemplateManager() (types.TemplateManager, error) {
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
		return management.NewOptimizedTemplateManager(config)
	}

	// Create standard template manager
	return management.NewTemplateManager()
}

// ensureDirectoryExists ensures a directory exists
func ensureDirectoryExists(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}
