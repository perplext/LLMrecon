package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/perplext/LLMrecon/src/utils/config"
	"github.com/perplext/LLMrecon/src/utils/monitoring"
	"github.com/perplext/LLMrecon/src/utils/profiling"
	"github.com/perplext/LLMrecon/src/utils/static"
)

var (
	// Command line flags
	envFlag        = flag.String("env", "dev", "Environment (dev, test, prod)")
	portFlag       = flag.Int("port", 8080, "HTTP server port")
	staticDirFlag  = flag.String("static-dir", "static", "Static files directory")
	profileFlag    = flag.Bool("profile", false, "Enable memory profiling")
	monitorFlag    = flag.Bool("monitor", false, "Enable monitoring")
	compressionFlag = flag.Bool("compression", true, "Enable gzip compression")
	cachingFlag    = flag.Bool("cache", true, "Enable file caching")
)

func main() {
	flag.Parse()

	// Set environment
	var env config.Environment
	switch *envFlag {
	case "dev":
		env = config.Development
	case "test":
		env = config.Testing
	case "prod":
		env = config.Production
	default:
		env = config.Development
	}

	// Load configuration
	cfg := config.GetMemoryConfig()
	cfg.SetEnvironment(env)

	// Create static file handler with custom options
	fileHandlerOptions := static.DefaultFileHandlerOptions()
	fileHandlerOptions.RootDir = *staticDirFlag
	fileHandlerOptions.EnableCompression = *compressionFlag
	fileHandlerOptions.EnableCache = *cachingFlag

	// Override with environment-specific settings if available
	if cfg.StaticFileHandler != nil {
		fileHandlerOptions = static.LoadFromConfig(cfg)
		
		// Command-line flags take precedence
		if *staticDirFlag != "static" {
			fileHandlerOptions.RootDir = *staticDirFlag
		}
		if *compressionFlag != true {
			enableCompression := *compressionFlag
			fileHandlerOptions.EnableCompression = enableCompression
		}
		if *cachingFlag != true {
			enableCache := *cachingFlag
			fileHandlerOptions.EnableCache = enableCache
		}
	}

	// Create the file handler
	fileHandler := static.NewFileHandler(fileHandlerOptions)

	// Setup memory profiler if enabled
	if *profileFlag || cfg.ProfilerEnabled {
		profiler, err := profiling.NewMemoryProfiler(&profiling.ProfilerOptions{
if err != nil {
treturn err
}			Interval:        time.Duration(cfg.ProfilerInterval) * time.Second,
			OutputDir:       cfg.ProfilerOutputDir,
			MemoryThreshold: int64(cfg.MemoryThreshold), // Already in MB
			GCThreshold:     100, // 100 ms default
		})
		if err != nil {
			log.Fatalf("Failed to create memory profiler: %v", err)
		}
		
		go profiler.Start(context.Background())
		defer profiler.Stop()
		
		log.Printf("Memory profiler started with interval %d seconds", cfg.ProfilerInterval)
	}

	// Setup monitoring if enabled
	var monitoringService *monitoring.MonitoringService
	if *monitorFlag || cfg.ProfilerEnabled {
		monitoringService, err = monitoring.NewMonitoringService(&monitoring.MonitoringServiceOptions{
			EnableConsoleLogging: true,
			CollectionInterval:   5 * time.Second,
			HeapAllocWarningMB:   float64(cfg.MemoryThreshold) * 0.8,
			HeapAllocCriticalMB:  float64(cfg.MemoryThreshold),
			AlertCooldown:        time.Minute,
		})
		if err != nil {
			log.Fatalf("Failed to create monitoring service: %v", err)
		}
		
		// Get metrics manager and register static file handler metrics
		metricsManager := monitoringService.GetMetricsManager()
		if mm, ok := metricsManager.(*monitoring.MetricsManager); ok {
			mm.RegisterGauge("static_file_cache_size", "Size of static file cache in bytes", nil)
			mm.RegisterGauge("static_file_cache_items", "Number of items in static file cache", nil)
			mm.RegisterCounter("static_file_requests", "Total static file requests", nil)
			mm.RegisterCounter("static_file_cache_hits", "Total cache hits", nil)
			mm.RegisterCounter("static_file_cache_misses", "Total cache misses", nil)
		}
		
		// Start monitoring service
		go monitoringService.Start()
		defer monitoringService.Stop()
		
		log.Println("Monitoring service started")
		
		// Update cache metrics periodically
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			
			for range ticker.C {
				monitoringService.UpdateMetric("static_file_cache_size", float64(fileHandler.GetCacheSize()))
				monitoringService.UpdateMetric("static_file_cache_items", float64(fileHandler.GetCacheItemCount()))
				
				// Log memory stats
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				log.Printf("Memory stats: Alloc=%v MiB, TotalAlloc=%v MiB, Sys=%v MiB, NumGC=%v",
					m.Alloc/1024/1024,
					m.TotalAlloc/1024/1024,
					m.Sys/1024/1024,
					m.NumGC)
			}
		}()
	}

	// Create HTTP server with the file handler
	mux := http.NewServeMux()
	
	// Register the static file handler for all requests
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if monitoringService != nil {
			monitoringService.IncrementMetric("static_file_requests", 1)
		}
		fileHandler.ServeHTTP(w, r)
	}))
	
	// Add a simple API endpoint for cache control
	mux.HandleFunc("/api/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		fileHandler.ClearCache()
		fmt.Fprintf(w, "Cache cleared successfully")
		log.Println("Cache cleared via API request")
	})
	
	// Add a simple API endpoint for cache stats
	mux.HandleFunc("/api/cache/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"cache_size_bytes": %d, "cache_items": %d}`, 
			fileHandler.GetCacheSize(), fileHandler.GetCacheItemCount())
	})

	// Create server with optimized settings
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", *portFlag),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting static file server on port %d in %s environment", *portFlag, env)
		log.Printf("Static files directory: %s", fileHandlerOptions.RootDir)
		log.Printf("Compression enabled: %v", *fileHandlerOptions.EnableCompression)
		log.Printf("Caching enabled: %v", *fileHandlerOptions.EnableCache)
if err != nil {
treturn err
}		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Shutdown server gracefully
	log.Println("Shutting down server...")
if err != nil {
treturn err
}	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}
	
	log.Println("Server stopped")
}
