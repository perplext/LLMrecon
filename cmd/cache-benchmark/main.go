package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/cache"
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

// BenchmarkOptions represents options for benchmarking
type BenchmarkOptions struct {
	// TemplateCount is the number of templates to use
	TemplateCount int
	// Iterations is the number of iterations to run
	Iterations int
	// ConcurrencyLevel is the concurrency level
	ConcurrencyLevel int
	// ResponseTime is the simulated LLM response time
	ResponseTime time.Duration
	// EnableCache enables caching
	EnableCache bool
	// CacheSize is the maximum size of the cache
	CacheSize int
	// CacheTTL is the TTL for cache entries
	CacheTTL time.Duration
	// EnableCompression enables compression of cached items
	EnableCompression bool
	// EnableMultiLevelCache enables multi-level caching
	EnableMultiLevelCache bool
	// EnableAdaptiveTTL enables adaptive TTL
	EnableAdaptiveTTL bool
	// EnableSharding enables sharding of the cache
	EnableSharding bool
	// ShardCount is the number of shards
	ShardCount int
	// QueryCacheSize is the size of the query cache
	QueryCacheSize int
	// ResultCacheSize is the size of the result cache
	ResultCacheSize int
	// EnableRepositoryCache enables repository caching
	EnableRepositoryCache bool
	// RepositoryCacheSize is the size of the repository cache
	RepositoryCacheSize int
}

func main() {
	// Command line flags
	sourcePtr := flag.String("source", "./templates", "Source path for templates")
	sourceTypePtr := flag.String("type", "file", "Source type (file, github, gitlab)")
	iterationsPtr := flag.Int("iterations", 3, "Number of benchmark iterations")
	templateCountPtr := flag.Int("count", 10, "Number of templates to execute")
	concurrencyPtr := flag.Int("concurrency", runtime.NumCPU(), "Concurrency level")
	responseTimePtr := flag.Int("response-time", 100, "Simulated LLM response time in milliseconds")
	enableCachePtr := flag.Bool("cache", true, "Enable caching")
	cacheSizePtr := flag.Int("cache-size", 1000, "Maximum cache size")
	cacheTTLPtr := flag.Int("cache-ttl", 3600, "Cache TTL in seconds")
	enableCompressionPtr := flag.Bool("compression", true, "Enable compression")
	enableMultiLevelCachePtr := flag.Bool("multi-level", true, "Enable multi-level caching")
	enableAdaptiveTTLPtr := flag.Bool("adaptive-ttl", true, "Enable adaptive TTL")
	enableShardingPtr := flag.Bool("sharding", true, "Enable cache sharding")
	shardCountPtr := flag.Int("shard-count", 8, "Number of cache shards")
	queryCacheSizePtr := flag.Int("query-cache-size", 200, "Query cache size")
	resultCacheSizePtr := flag.Int("result-cache-size", 100, "Result cache size")
	enableRepositoryCachePtr := flag.Bool("repo-cache", true, "Enable repository caching")
	repositoryCacheSizePtr := flag.Int("repo-cache-size", 500, "Repository cache size")
	verbosePtr := flag.Bool("verbose", false, "Enable verbose output")
	
	flag.Parse()

	// Create benchmark options
	options := &BenchmarkOptions{
		TemplateCount:        *templateCountPtr,
		Iterations:           *iterationsPtr,
		ConcurrencyLevel:     *concurrencyPtr,
		ResponseTime:         time.Duration(*responseTimePtr) * time.Millisecond,
		EnableCache:          *enableCachePtr,
		CacheSize:            *cacheSizePtr,
		CacheTTL:             time.Duration(*cacheTTLPtr) * time.Second,
		EnableCompression:    *enableCompressionPtr,
		EnableMultiLevelCache: *enableMultiLevelCachePtr,
		EnableAdaptiveTTL:    *enableAdaptiveTTLPtr,
		EnableSharding:       *enableShardingPtr,
		ShardCount:           *shardCountPtr,
		QueryCacheSize:       *queryCacheSizePtr,
		ResultCacheSize:      *resultCacheSizePtr,
		EnableRepositoryCache: *enableRepositoryCachePtr,
		RepositoryCacheSize:  *repositoryCacheSizePtr,
	}

	// Create repository manager
	repoManager := repository.NewManager()

	// Create repository cache manager if enabled
	var repoCacheManager *repository.CacheManager
	if options.EnableRepositoryCache {
		repoCacheOptions := &repository.CacheManagerOptions{
			EnableCache:      true,
			DefaultTTL:       options.CacheTTL,
			MaxSize:          options.RepositoryCacheSize,
			EnableCompression: options.EnableCompression,
			CompressionLevel: 6,
			PruneInterval:    10 * time.Minute,
		}
		repoCacheManager = repository.NewCacheManager(repoManager, repoCacheOptions)
	}

	// Create template loader
	loaderOptions := loader.ResourceEfficientLoaderOptions{
		CacheTTL:                  options.CacheTTL,
		MaxCacheSize:              options.CacheSize,
		ConcurrencyLimit:          options.ConcurrencyLevel,
		EnableOptimization:        true,
		EnableStructureOptimization: true,
	}
	
	var templateLoader loader.TemplateLoader
	if repoCacheManager != nil {
		// Use repository cache manager with loader
		templateLoader = loader.NewResourceEfficientLoader(repoManager, loaderOptions)
	} else {
		templateLoader = loader.NewResourceEfficientLoader(repoManager, loaderOptions)
	}

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
	if len(templates) > options.TemplateCount {
		templates = templates[:options.TemplateCount]
		fmt.Printf("Using %d templates for benchmarking\n", len(templates))
	}

	// Create mock LLM provider
	provider := &MockLLMProvider{
		name:            "mock",
		responseTime:    options.ResponseTime,
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
		MaxConcurrent:   options.ConcurrencyLevel,
		Variables:       make(map[string]interface{}),
		ProviderOptions: map[string]interface{}{
			"provider": "mock",
		},
	}

	// Run benchmarks with and without caching
	fmt.Println("\nRunning benchmarks...")

	// Benchmark without caching first
	if options.EnableCache {
		fmt.Println("\n1. Benchmark WITHOUT caching:")
		noCacheExecutor := execution.NewOptimizedTemplateExecutor(execOptions, 0, 0, options.ConcurrencyLevel)
		runBenchmark(ctx, noCacheExecutor, templates, options.Iterations, *verbosePtr)
	}

	// Benchmark with standard caching
	fmt.Println("\n2. Benchmark WITH standard caching:")
	standardCacheExecutor := execution.NewOptimizedTemplateExecutor(
		execOptions, 
		options.CacheSize, 
		options.CacheTTL, 
		options.ConcurrencyLevel,
	)
	runBenchmark(ctx, standardCacheExecutor, templates, options.Iterations, *verbosePtr)

	// Benchmark with multi-level caching if enabled
	if options.EnableMultiLevelCache {
		fmt.Println("\n3. Benchmark WITH multi-level caching:")
		
		// Create multi-level cache options
		multiLevelOptions := cache.DefaultCacheOptions()
		multiLevelOptions.EnableFragmentCache = true
		multiLevelOptions.EnableTemplateCache = true
		multiLevelOptions.EnableQueryCache = true
		multiLevelOptions.EnableResultCache = true
		multiLevelOptions.FragmentCacheSize = options.CacheSize
		multiLevelOptions.TemplateCacheSize = options.CacheSize
		multiLevelOptions.QueryCacheSize = options.QueryCacheSize
		multiLevelOptions.ResultCacheSize = options.ResultCacheSize
		multiLevelOptions.FragmentTTL = options.CacheTTL
		multiLevelOptions.TemplateTTL = options.CacheTTL
		multiLevelOptions.QueryTTL = options.CacheTTL
		multiLevelOptions.ResultTTL = options.CacheTTL
		multiLevelOptions.EnableCompression = options.EnableCompression
		multiLevelOptions.EnableSharding = options.EnableSharding
		multiLevelOptions.ShardCount = options.ShardCount
		multiLevelOptions.EnableAdaptiveTTL = options.EnableAdaptiveTTL
		
		// Create multi-level cache
		multiLevelCache := cache.NewMultiLevelCache(multiLevelOptions)
		
		// Create executor with multi-level cache
		multiLevelExecutor := execution.NewOptimizedTemplateExecutor(
			execOptions,
			options.CacheSize, 
			options.CacheTTL, 
			options.ConcurrencyLevel,
		)
		
		runBenchmark(ctx, multiLevelExecutor, templates, options.Iterations, *verbosePtr)
	}

	// Print summary
	fmt.Println("\nBenchmark Summary:")
	fmt.Printf("- Templates: %d\n", len(templates))
	fmt.Printf("- Iterations: %d\n", options.Iterations)
	fmt.Printf("- Concurrency: %d\n", options.ConcurrencyLevel)
	fmt.Printf("- Response Time: %v\n", options.ResponseTime)
	
	if options.EnableCache {
		fmt.Printf("- Cache Size: %d\n", options.CacheSize)
		fmt.Printf("- Cache TTL: %v\n", options.CacheTTL)
		fmt.Printf("- Compression: %v\n", options.EnableCompression)
		
		if options.EnableMultiLevelCache {
			fmt.Printf("- Multi-Level Cache: %v\n", options.EnableMultiLevelCache)
			fmt.Printf("- Adaptive TTL: %v\n", options.EnableAdaptiveTTL)
			fmt.Printf("- Sharding: %v (Shards: %d)\n", options.EnableSharding, options.ShardCount)
			fmt.Printf("- Query Cache Size: %d\n", options.QueryCacheSize)
			fmt.Printf("- Result Cache Size: %d\n", options.ResultCacheSize)
		}
		
		if options.EnableRepositoryCache {
			fmt.Printf("- Repository Cache: %v (Size: %d)\n", options.EnableRepositoryCache, options.RepositoryCacheSize)
		}
	}
}

// runBenchmark runs a benchmark of template execution
func runBenchmark(ctx context.Context, executor interfaces.TemplateExecutor, templates []*format.Template, iterations int, verbose bool) {
	// Run benchmark iterations
	var totalExecutionTime time.Duration
	var totalTemplates int

	fmt.Printf("Running %d iterations...\n", iterations)
	
	// First iteration - warm up
	fmt.Print("Warm-up iteration: ")
	results, err := executor.ExecuteBatch(ctx, templates, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Executed %d templates\n", len(results))
	}
	
	// Clear any statistics
	if clearable, ok := executor.(interface{ ClearStats() }); ok {
		clearable.ClearStats()
	}
	
	// Run actual benchmark iterations
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
	}
}

// generateRandomTemplates generates random templates for testing
func generateRandomTemplates(count int) []*format.Template {
	templates := make([]*format.Template, count)
	
	for i := 0; i < count; i++ {
		template := &format.Template{
			ID: fmt.Sprintf("template-%d", i),
			Info: format.TemplateInfo{
				Name:        fmt.Sprintf("Test Template %d", i),
				Description: "Generated test template",
				Version:     "1.0.0",
			},
			Variables: map[string]interface{}{
				"var1": fmt.Sprintf("value-%d", i),
				"var2": i,
			},
			// Store content as raw bytes
			Content: []byte(fmt.Sprintf("# Template %d\nPrompt: Tell me about topic %d", i, rand.Intn(100))),
		}
		templates[i] = template
	}
	
	return templates
}
