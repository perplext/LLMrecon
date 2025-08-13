package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/loader"
)

// MockLLMProvider is a mock LLM provider for testing
type MockLLMProvider struct {
	name            string
	responseTime    time.Duration
	supportedModels []string
}

// SendPrompt sends a prompt to the LLM and returns the response
func (p *MockLLMProvider) SendPrompt(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	// Simulate processing time
	select {
	case <-time.After(p.responseTime):
		return fmt.Sprintf("Response to: %s", prompt), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// GetSupportedModels returns the list of supported models
func (p *MockLLMProvider) GetSupportedModels() []string {
	return p.supportedModels
}

// GetName returns the name of the provider
func (p *MockLLMProvider) GetName() string {
	return p.name
}

// MockDetectionEngine is a mock detection engine for testing
type MockDetectionEngine struct {
	name string
}

// Detect detects vulnerabilities in an LLM response
func (e *MockDetectionEngine) Detect(ctx context.Context, template *format.Template, response string) (bool, int, map[string]interface{}, error) {
	// Simulate detection
	return false, 0, map[string]interface{}{"reason": "No vulnerability detected"}, nil
}

// GetName returns the name of the detection engine
func (e *MockDetectionEngine) GetName() string {
	return e.name
}

func main() {
	// Command line flags
	sourcePtr := flag.String("source", "./templates", "Source path for templates")
	sourceTypePtr := flag.String("type", "file", "Source type (file, github, gitlab)")
	executorTypePtr := flag.String("executor", "pipeline", "Executor type (standard, optimized, async, pipeline)")
	concurrencyPtr := flag.Int("concurrency", runtime.NumCPU(), "Concurrency limit")
	iterationsPtr := flag.Int("iterations", 3, "Number of benchmark iterations")
	templateCountPtr := flag.Int("count", 10, "Number of templates to execute")
	responseTimePtr := flag.Int("response-time", 100, "Simulated LLM response time in milliseconds")
	verbosePtr := flag.Bool("verbose", false, "Enable verbose output")
	
	flag.Parse()

	// Create repository manager
	repoManager := repository.NewManager()

	// Create template loader
	loaderOptions := loader.ResourceEfficientLoaderOptions{
		CacheTTL:                  1 * time.Hour,
		MaxCacheSize:              1000,
		ConcurrencyLimit:          *concurrencyPtr,
		EnableOptimization:        true,
		EnableStructureOptimization: true,
	}
	templateLoader := loader.NewResourceEfficientLoader(repoManager, loaderOptions)

	// Create context
	ctx := context.Background()

	// Load templates
	fmt.Printf("Loading templates from %s (%s)...\n", *sourcePtr, *sourceTypePtr)
	startTime := time.Now()
	templates, err := templateLoader.LoadTemplates(ctx, *sourcePtr, *sourceTypePtr)
	if err != nil {
		fmt.Printf("Error loading templates: %v\n", err)
		os.Exit(1)
	}
	loadTime := time.Since(startTime)

	fmt.Printf("Loaded %d templates in %v\n", len(templates), loadTime)

	// Limit template count if needed
	if len(templates) > *templateCountPtr {
		templates = templates[:*templateCountPtr]
		fmt.Printf("Using %d templates for benchmarking\n", len(templates))
	}

	// Create mock LLM provider
	provider := &MockLLMProvider{
		name:            "mock",
		responseTime:    time.Duration(*responseTimePtr) * time.Millisecond,
		supportedModels: []string{"mock-model"},
	}

	// Create mock detection engine
	detectionEngine := &MockDetectionEngine{
		name: "mock",
	}

	// Create execution options
	execOptions := &execution.ExecutionOptions{
		Provider:        provider,
		DetectionEngine: detectionEngine,
		Timeout:         30 * time.Second,
		RetryCount:      3,
		RetryDelay:      1 * time.Second,
		MaxConcurrent:   *concurrencyPtr,
		Variables:       make(map[string]interface{}),
		ProviderOptions: map[string]interface{}{
			"provider": "mock",
		},
	}

	// Create executor based on type
	var executor interfaces.TemplateExecutor
	switch *executorTypePtr {
	case "standard":
		executor = execution.NewTemplateExecutor(execOptions)
	case "optimized":
		executor = execution.NewOptimizedTemplateExecutor(execOptions, 1000, 1*time.Hour, *concurrencyPtr)
	case "async":
		baseExecutor := execution.NewOptimizedTemplateExecutor(execOptions, 1000, 1*time.Hour, *concurrencyPtr)
		executor = execution.NewAsyncTemplateExecutor(baseExecutor, 1000, 1000, 1*time.Hour)
	case "pipeline":
		baseExecutor := execution.NewOptimizedTemplateExecutor(execOptions, 1000, 1*time.Hour, *concurrencyPtr)
		executor = execution.NewPipelineExecutor(baseExecutor, 100)
	default:
		fmt.Printf("Unknown executor type: %s\n", *executorTypePtr)
		os.Exit(1)
	}

	// Run benchmark
	fmt.Printf("Running benchmark with %s executor...\n", *executorTypePtr)
	runBenchmark(ctx, executor, templates, *iterationsPtr, *verbosePtr)
}

// runBenchmark runs a benchmark of template execution
func runBenchmark(ctx context.Context, executor interfaces.TemplateExecutor, templates []*format.Template, iterations int, verbose bool) {
	// Run benchmark iterations
	var totalExecutionTime time.Duration
	var totalTemplates int

	fmt.Printf("Running %d iterations...\n", iterations)
	for i := 0; i < iterations; i++ {
		fmt.Printf("Iteration %d/%d: ", i+1, iterations)
		
		// Execute templates
		startTime := time.Now()
		results, err := executor.ExecuteBatch(ctx, templates, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		executionTime := time.Since(startTime)
		
		totalExecutionTime += executionTime
		totalTemplates += len(results)
		
		fmt.Printf("Executed %d templates in %v\n", len(results), executionTime)

		// Print detailed results if verbose
		if verbose {
			fmt.Println("Results:")
			for j, result := range results {
				fmt.Printf("  %d: %s - %s (Duration: %v)\n", j+1, result.TemplateID, result.Status, result.Duration)
			}
		}
	}

	// Calculate average
	avgExecutionTime := totalExecutionTime / time.Duration(iterations)
	avgTemplatesPerSecond := float64(totalTemplates) / totalExecutionTime.Seconds() * float64(iterations)

	fmt.Println("\nBenchmark Results:")
	fmt.Printf("  Total Templates:        %d\n", totalTemplates/iterations)
	fmt.Printf("  Average Execution Time: %v\n", avgExecutionTime)
	fmt.Printf("  Templates per Second:   %.2f\n", avgTemplatesPerSecond)

	// Print executor-specific stats if available
	switch e := executor.(type) {
	case *execution.OptimizedTemplateExecutor:
		stats := e.GetExecutionStats()
		fmt.Println("\nOptimized Executor Statistics:")
		fmt.Printf("  Total Executions:     %d\n", stats["total_executions"])
		fmt.Printf("  Successful Executions: %d\n", stats["successful_executions"])
		fmt.Printf("  Failed Executions:    %d\n", stats["failed_executions"])
		fmt.Printf("  Cached Responses:     %d\n", stats["cached_responses"])
		fmt.Printf("  Average Execution Time: %v\n", stats["avg_execution_time"])
	case *execution.AsyncTemplateExecutor:
		stats := e.GetExecutionStats()
		fmt.Println("\nAsync Executor Statistics:")
		fmt.Printf("  Total Tasks:          %d\n", stats["total_tasks"])
		fmt.Printf("  Completed Tasks:      %d\n", stats["completed_tasks"])
		fmt.Printf("  Failed Tasks:         %d\n", stats["failed_tasks"])
		fmt.Printf("  Queued Tasks:         %d\n", stats["queued_tasks"])
		fmt.Printf("  Average Execution Time: %v\n", stats["avg_execution_time"])
		fmt.Printf("  Active Workers:       %d\n", stats["active_workers"])
	case *execution.PipelineExecutor:
		stats := e.GetPipelineStats()
		fmt.Println("\nPipeline Executor Statistics:")
		fmt.Printf("  Total Tasks:          %d\n", stats["total_tasks"])
		fmt.Printf("  Completed Tasks:      %d\n", stats["completed_tasks"])
		fmt.Printf("  Failed Tasks:         %d\n", stats["failed_tasks"])
		fmt.Printf("  Average Duration:     %v\n", stats["avg_duration"])
		
		if stageDurations, ok := stats["avg_stage_durations"].(map[string]time.Duration); ok {
			fmt.Println("  Average Stage Durations:")
			for stage, duration := range stageDurations {
				fmt.Printf("    %s: %v\n", stage, duration)
			}
		}
	}
}
