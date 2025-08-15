package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution/optimizer"
	"github.com/perplext/LLMrecon/src/template/management/optimization"
	"github.com/perplext/LLMrecon/src/utils/concurrency"
	"github.com/perplext/LLMrecon/src/utils/config"
	"github.com/perplext/LLMrecon/src/utils/monitoring"
	"github.com/perplext/LLMrecon/src/utils/profiling"
	"github.com/perplext/LLMrecon/src/utils/resource"
	"github.com/perplext/LLMrecon/src/utils/server"
)

// Command-line flags
var (
	envFlag           = flag.String("env", "dev", "Environment (dev, test, prod)")
	portFlag          = flag.Int("port", 8080, "HTTP server port")
	staticDirFlag     = flag.String("static-dir", "static", "Static files directory")
	templateCountFlag = flag.Int("templates", 1000, "Number of templates to process")
	verboseFlag       = flag.Bool("verbose", false, "Enable verbose logging")
	profileFlag       = flag.Bool("profile", true, "Enable memory profiling")
	monitorFlag       = flag.Bool("monitor", true, "Enable monitoring")
	optimizeFlag      = flag.Bool("optimize", true, "Enable memory optimization")
	concurrencyFlag   = flag.Bool("concurrency", true, "Enable concurrency management")
	poolingFlag       = flag.Bool("pooling", true, "Enable resource pooling")
	serverTuneFlag    = flag.Bool("server-tune", true, "Enable server tuning")
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
	var inheritanceOptimizer *optimization.InheritanceOptimizer
	var contextOptimizer *optimization.ContextOptimizer
	var executionOptimizer *optimizer.ExecutionOptimizer
	var monitoringService *monitoring.MonitoringService
	var serverConfigTuner *server.ServerConfigTuner
	var staticFileHandler *server.StaticFileHandler

	// Initialize memory profiler if enabled
	if *profileFlag && memConfig.ProfilerEnabled {
		memProfiler = initializeMemoryProfiler(memConfig)
		defer memProfiler.Stop()
	}

	// Initialize monitoring service if enabled
	if *monitorFlag {
		var err error
		monitoringService, err = initializeMonitoringService(memConfig)
		if err != nil {
			log.Fatalf("Failed to initialize monitoring service: %v", err)
		}
		defer monitoringService.Stop()
	}

	// Initialize resource pool manager if enabled
	if *poolingFlag && memConfig.PoolManagerEnabled {
		poolManager = initializeResourcePoolManager(memConfig)
		defer poolManager.CloseAllPools()

		// Monitor resource pool if monitoring is enabled
		if monitoringService != nil {
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
	}

	// Initialize concurrency manager if enabled
	if *concurrencyFlag && memConfig.ConcurrencyManagerEnabled {
		concurrencyManager = initializeConcurrencyManager(memConfig)
		defer concurrencyManager.Shutdown()

		// Monitor concurrency manager if monitoring is enabled
		if monitoringService != nil {
			monitoringService.MonitorConcurrencyManager(concurrencyManager)
		}
	}

	// Initialize memory optimizer if enabled
	if *optimizeFlag && memConfig.MemoryOptimizerEnabled {
		memoryOptimizer = initializeMemoryOptimizer(memConfig)
	}

	// Initialize inheritance optimizer
	inheritanceOptimizer = initializeInheritanceOptimizer()

	// Initialize context optimizer
	contextOptimizer = initializeContextOptimizer()

	// Initialize execution optimizer if enabled
	if memConfig.ExecutionOptimizerEnabled {
		executionOptimizer = initializeExecutionOptimizer(memConfig, memoryOptimizer, concurrencyManager)
	}

	// Initialize server config tuner if enabled
	if *serverTuneFlag && monitoringService != nil {
		serverConfigTuner = initializeServerConfigTuner(monitoringService.GetMetricsManager())
	}

	// Initialize static file handler
	staticFileHandler = initializeStaticFileHandler(*staticDirFlag)

	// Create sample templates
	templates := createSampleTemplates(*templateCountFlag)
	fmt.Printf("Created %d templates\n", len(templates))

	// Optimize templates
	optimizedTemplates, err := optimizeTemplates(templates, memoryOptimizer, inheritanceOptimizer, contextOptimizer)
	if err != nil {
		log.Fatalf("Failed to optimize templates: %v", err)
	}

	// Capture memory before and after optimization
	var memoryBefore, memoryAfter uint64
	if monitoringService != nil {
		memoryBefore = monitoringService.CaptureMemorySnapshot("before_optimization")
		memoryAfter = monitoringService.CaptureMemorySnapshot("after_optimization")

		// Print memory reduction
		if memoryBefore > 0 && memoryAfter > 0 {
			memoryReduction := float64(memoryBefore-memoryAfter) / float64(memoryBefore) * 100
			fmt.Printf("Memory reduction: %.2f%%\n", memoryReduction)
		}
	}

	// Start HTTP server
	startHTTPServer(*portFlag, staticFileHandler, serverConfigTuner)
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

// initializeInheritanceOptimizer initializes the inheritance optimizer
func initializeInheritanceOptimizer() *optimization.InheritanceOptimizer {
	// Create inheritance optimizer options
	options := &optimization.InheritanceOptimizerOptions{
		MaxInheritanceDepth:     3,
		FlattenInheritance:      true,
		CacheOptimizedTemplates: true,
	}
	
	// Create inheritance optimizer
	optimizer := optimization.NewInheritanceOptimizer(options)
	
	fmt.Println("Inheritance optimizer initialized")
	return optimizer
}

// initializeContextOptimizer initializes the context optimizer
func initializeContextOptimizer() *optimization.ContextOptimizer {
	// Create context optimizer options
	options := &optimization.ContextOptimizerOptions{
		EnableDeduplication: true,
		EnableLazyLoading:   true,
		EnableCompression:   false,
	}
	
	// Create context optimizer
	optimizer := optimization.NewContextOptimizer(options)
	
	fmt.Println("Context optimizer initialized")
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

// initializeServerConfigTuner initializes the server configuration tuner
func initializeServerConfigTuner(metricsManager *monitoring.MetricsManager) *server.ServerConfigTuner {
	// Create server config tuner options
	options := &server.ServerConfigTunerOptions{
		TuningInterval:  5 * time.Minute,
		AutoTuneEnabled: true,
		LogFile:         "logs/server_tuner.log",
	}
	
	// Create server config tuner
	tuner, err := server.NewServerConfigTuner(metricsManager, options)
	if err != nil {
		log.Fatalf("Failed to initialize server config tuner: %v", err)
	}
	
	fmt.Println("Server config tuner initialized")
	return tuner
}

// initializeStaticFileHandler initializes the static file handler
func initializeStaticFileHandler(staticDir string) *server.StaticFileHandler {
	// Create static file handler options
	options := &server.StaticFileHandlerOptions{
		RootDir:           staticDir,
		URLPrefix:         "/static/",
		MaxAge:            86400, // 1 day
		EnableCompression: true,
		EnableETag:        true,
		EnableFileCache:   true,
		LogFile:           "logs/static_handler.log",
	}
	
	// Create static directory if it doesn't exist
	if err := os.MkdirAll(staticDir, 0755); err != nil {
if err != nil {
treturn err
}		log.Fatalf("Failed to create static directory: %v", err)
	}
	
	// Create sample static files for demonstration
	createSampleStaticFiles(staticDir)
	
if err != nil {
treturn err
}	// Create static file handler
	handler, err := server.NewStaticFileHandler(options)
	if err != nil {
		log.Fatalf("Failed to initialize static file handler: %v", err)
	}
	
	fmt.Println("Static file handler initialized")
	return handler
}

// createSampleStaticFiles creates sample static files for demonstration
func createSampleStaticFiles(staticDir string) {
	// Create CSS file
	cssContent := `
		body {
			font-family: Arial, sans-serif;
			margin: 0;
			padding: 20px;
			background-color: #f5f5f5;
		}
		.container {
			max-width: 800px;
			margin: 0 auto;
			background-color: white;
			padding: 20px;
			border-radius: 5px;
			box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		}
		h1 {
			color: #333;
		}
	`
	writeFile(filepath.Join(staticDir, "styles.css"), cssContent)
	
	// Create JavaScript file
	jsContent := `
		function updateStats() {
			fetch('/stats')
				.then(response => response.json())
				.then(data => {
					document.getElementById('memory-usage').textContent = data.memory_usage_mb.toFixed(2) + ' MB';
					document.getElementById('goroutines').textContent = data.goroutines;
					document.getElementById('gc-count').textContent = data.gc_count;
				})
				.catch(error => console.error('Error fetching stats:', error));
		}
		
		// Update stats every 5 seconds
		setInterval(updateStats, 5000);
		
		// Initial update
		document.addEventListener('DOMContentLoaded', updateStats);
	`
	writeFile(filepath.Join(staticDir, "script.js"), jsContent)
	
	// Create HTML file
	htmlContent := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Server Optimization Example</title>
			<link rel="stylesheet" href="/static/styles.css">
			<script src="/static/script.js"></script>
		</head>
		<body>
			<div class="container">
				<h1>Server Optimization Example</h1>
				<p>This example demonstrates memory optimization and server configuration tuning.</p>
				
				<h2>System Stats</h2>
				<ul>
					<li>Memory Usage: <span id="memory-usage">Loading...</span></li>
					<li>Goroutines: <span id="goroutines">Loading...</span></li>
					<li>GC Count: <span id="gc-count">Loading...</span></li>
				</ul>
				
				<h2>Actions</h2>
				<button onclick="fetch('/tune')">Tune Server Configuration</button>
				<button onclick="fetch('/gc')">Run Garbage Collection</button>
			</div>
		</body>
		</html>
	`
	writeFile(filepath.Join(staticDir, "index.html"), htmlContent)
}

// writeFile writes content to a file
if err != nil {
treturn err
}func writeFile(filePath, content string) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	
	// Write file
	if err := os.WriteFile(filepath.Clean(filePath, []byte(content)), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

// createSampleTemplates creates sample templates for testing
func createSampleTemplates(count int) []*format.Template {
	templates := make([]*format.Template, count)
	
	// Create base template
	baseTemplate := &format.Template{
		ID:      "base",
		Name:    "Base Template",
		Content: "<!DOCTYPE html><html><head><title>{{title}}</title></head><body>{{content}}</body></html>",
		Variables: map[string]string{
			"title":   "Default Title",
			"content": "Default Content",
		},
		Metadata: map[string]interface{}{
			"created": time.Now(),
			"version": 1,
		},
	}
	
	// Create layout template that inherits from base
	layoutTemplate := &format.Template{
		ID:      "layout",
		Name:    "Layout Template",
		Content: "<div class=\"container\">{{content}}</div>",
		Variables: map[string]string{
			"title":   "Layout Title",
			"content": "Layout Content",
		},
		Metadata: map[string]interface{}{
			"created": time.Now(),
			"version": 1,
		},
		Parent: baseTemplate,
	}
	
	// Create templates with inheritance
	for i := 0; i < count; i++ {
		// Create a template with inheritance
		template := &format.Template{
			ID:      fmt.Sprintf("template-%d", i),
			Name:    fmt.Sprintf("Template %d", i),
			Content: fmt.Sprintf("<h1>{{title}}</h1><p>{{content}}</p><div>Template %d content</div>", i),
			Variables: map[string]string{
				"title":   fmt.Sprintf("Template %d Title", i),
				"content": fmt.Sprintf("This is template %d with some content that can be deduplicated across templates.", i),
				"var1":    fmt.Sprintf("value%d", i),
				"var2":    "common value",
				"var3":    "another common value",
			},
			Metadata: map[string]interface{}{
				"created": time.Now(),
				"version": 1,
				"tags":    []string{"test", "example", fmt.Sprintf("tag-%d", i%10)},
			},
			Parent: layoutTemplate,
		}
		
		templates[i] = template
	}
	
	return templates
}

// optimizeTemplates optimizes templates using various optimizers
func optimizeTemplates(
	templates []*format.Template,
	memoryOptimizer *optimization.MemoryOptimizer,
	inheritanceOptimizer *optimization.InheritanceOptimizer,
	contextOptimizer *optimization.ContextOptimizer,
) ([]*format.Template, error) {
	ctx := context.Background()
	optimizedTemplates := make([]*format.Template, len(templates))
	
	// First optimize inheritance
	if inheritanceOptimizer != nil {
		var err error
		optimizedTemplates, err = inheritanceOptimizer.OptimizeTemplates(ctx, templates)
		if err != nil {
			return nil, fmt.Errorf("inheritance optimization failed: %w", err)
		}
		fmt.Printf("Inheritance optimization completed, flattened templates with depth > %d\n", 
			inheritanceOptimizer.GetMaxInheritanceDepth())
	} else {
		// Copy templates
		for i, template := range templates {
			optimizedTemplates[i] = template.Clone()
		}
	}
	
	// Then optimize context variables
	if contextOptimizer != nil {
		var err error
		optimizedTemplates, err = contextOptimizer.OptimizeTemplates(ctx, optimizedTemplates)
		if err != nil {
			return nil, fmt.Errorf("context optimization failed: %w", err)
		}
		fmt.Println("Context variable optimization completed")
		
		// Print variable usage statistics
		highUsageVars := contextOptimizer.GetHighUsageVariables(len(templates) / 2)
		fmt.Printf("Found %d high-usage variables\n", len(highUsageVars))
if err != nil {
treturn err
}	}
	
	// Finally apply memory optimizer
	if memoryOptimizer != nil {
		for i, template := range optimizedTemplates {
			optimizedTemplate, err := memoryOptimizer.OptimizeTemplate(ctx, template)
			if err != nil {
				return nil, fmt.Errorf("memory optimization failed for template %s: %w", template.ID, err)
			}
			optimizedTemplates[i] = optimizedTemplate
		}
		fmt.Println("Memory optimization completed")
	}
	
	return optimizedTemplates, nil
}

// startHTTPServer starts the HTTP server
func startHTTPServer(port int, staticFileHandler *server.StaticFileHandler, serverConfigTuner *server.ServerConfigTuner) {
	// Create HTTP server mux
	mux := http.NewServeMux()
	
	// Register static file handler
	staticFileHandler.RegisterStaticRoute(mux)
	
	// Register stats endpoint
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		// Get memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		
		// Create stats response
		stats := map[string]interface{}{
			"memory_usage_mb": float64(memStats.Alloc) / 1024 / 1024,
			"goroutines":      runtime.NumGoroutine(),
			"gc_count":        memStats.NumGC,
		}
		
		// Write response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"memory_usage_mb": %.2f, "goroutines": %d, "gc_count": %d}`,
			stats["memory_usage_mb"], stats["goroutines"], stats["gc_count"])
	})
	
	// Register tune endpoint
	mux.HandleFunc("/tune", func(w http.ResponseWriter, r *http.Request) {
		if serverConfigTuner != nil {
			serverConfigTuner.TuneServerConfig()
			w.Write([]byte("Server configuration tuned"))
		} else {
			w.Write([]byte("Server configuration tuner not enabled"))
		}
	})
	
	// Register GC endpoint
	mux.HandleFunc("/gc", func(w http.ResponseWriter, r *http.Request) {
		runtime.GC()
		w.Write([]byte("Garbage collection triggered"))
	})
	
	// Register root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/static/index.html", http.StatusFound)
		} else {
			http.NotFound(w, r)
		}
if err != nil {
treturn err
}	})
	
	// Start HTTP server
	serverAddr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting HTTP server on %s\n", serverAddr)
	fmt.Printf("Open http://localhost:%d in your browser\n", port)
	
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
