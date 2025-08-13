package config

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/perplext/LLMrecon/src/utils/profiling"
)

// ConfigTuner provides automatic configuration tuning based on system metrics
type ConfigTuner struct {
	// config is the current configuration
	config *TunerConfig
	// profiler is the memory profiler
	profiler *profiling.MemoryProfiler
	// mutex protects the tuner state
	mutex sync.RWMutex
	// running indicates if automatic tuning is running
	running bool
	// stopChan is used to stop automatic tuning
	stopChan chan struct{}
	// lastTuneTime is the time of the last tuning
	lastTuneTime time.Time
	// tuneCount is the number of tuning operations
	tuneCount int
	// recommendations stores tuning recommendations
	recommendations []string
	// onConfigChange is called when configuration changes
	onConfigChange func(*TunerConfig)
}

// TunerConfig represents configuration for the tuner
type TunerConfig struct {
	// WorkerCount is the number of worker processes
	WorkerCount int
	// MaxConcurrentRequests is the maximum number of concurrent requests
	MaxConcurrentRequests int
	// ConnectionPoolSize is the size of the connection pool
	ConnectionPoolSize int
	// KeepAliveTimeout is the keep-alive timeout for connections
	KeepAliveTimeout time.Duration
	// ReadTimeout is the read timeout for requests
	ReadTimeout time.Duration
	// WriteTimeout is the write timeout for responses
	WriteTimeout time.Duration
	// IdleTimeout is the idle timeout for connections
	IdleTimeout time.Duration
	// MaxHeaderBytes is the maximum size of request headers
	MaxHeaderBytes int
	// BufferPoolSize is the size of the buffer pool
	BufferPoolSize int
	// TemplateCache is the size of the template cache
	TemplateCache int
	// StaticFileCache is the size of the static file cache
	StaticFileCache int
	// GzipCompression enables gzip compression
	GzipCompression bool
	// CompressionLevel is the compression level (1-9)
	CompressionLevel int
	// MaxRequestBodySize is the maximum size of request bodies
	MaxRequestBodySize int64
	// EnableHTTP2 enables HTTP/2
	EnableHTTP2 bool
	// GCPercent is the garbage collection target percentage
	GCPercent int
	// MaxMemory is the maximum memory usage (in MB)
	MaxMemory int64
}

// DefaultTunerConfig returns default configuration for the tuner
func DefaultTunerConfig() *TunerConfig {
	numCPU := runtime.NumCPU()
	
	return &TunerConfig{
		WorkerCount:          numCPU,
		MaxConcurrentRequests: numCPU * 100,
		ConnectionPoolSize:   numCPU * 10,
		KeepAliveTimeout:     60 * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		IdleTimeout:          120 * time.Second,
		MaxHeaderBytes:       1 << 20, // 1 MB
		BufferPoolSize:       1000,
		TemplateCache:        1000,
		StaticFileCache:      1000,
		GzipCompression:      true,
		CompressionLevel:     6,
		MaxRequestBodySize:   10 << 20, // 10 MB
		EnableHTTP2:          true,
		GCPercent:            100,
		MaxMemory:            0, // No limit
	}
}

// NewConfigTuner creates a new configuration tuner
func NewConfigTuner(config *TunerConfig, onConfigChange func(*TunerConfig)) (*ConfigTuner, error) {
	if config == nil {
		config = DefaultTunerConfig()
	}

	// Create memory profiler
	profilerOptions := profiling.DefaultProfilerOptions()
	profiler, err := profiling.NewMemoryProfiler(profilerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory profiler: %w", err)
	}

	return &ConfigTuner{
		config:        config,
		profiler:      profiler,
		stopChan:      make(chan struct{}),
		onConfigChange: onConfigChange,
	}, nil
}

// StartAutomaticTuning starts automatic configuration tuning
func (t *ConfigTuner) StartAutomaticTuning(interval time.Duration) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.running {
		return fmt.Errorf("automatic tuning is already running")
	}

	t.running = true
	t.stopChan = make(chan struct{})

	// Start memory profiling
	if err := t.profiler.StartAutomaticProfiling(); err != nil {
		t.running = false
		return fmt.Errorf("failed to start memory profiling: %w", err)
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.TuneConfiguration()
			case <-t.stopChan:
				return
			}
		}
	}()

	return nil
}

// StopAutomaticTuning stops automatic configuration tuning
func (t *ConfigTuner) StopAutomaticTuning() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.running {
		return
	}

	close(t.stopChan)
	t.running = false

	// Stop memory profiling
	t.profiler.StopAutomaticProfiling()
}

// TuneConfiguration tunes the configuration based on system metrics
func (t *ConfigTuner) TuneConfiguration() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Update tune count and last tune time
	t.tuneCount++
	t.lastTuneTime = time.Now()

	// Get memory stats
	memStats := t.profiler.GetFormattedMemoryStats()
	gcStats := t.profiler.GetGCStats()

	// Clear recommendations
	t.recommendations = nil

	// Tune worker count based on CPU count and load
	numCPU := runtime.NumCPU()
	numGoroutines := memStats["num_goroutines"].(int)
	
	if numGoroutines > numCPU*1000 {
		// Too many goroutines, reduce worker count
		if t.config.WorkerCount > 1 {
			t.config.WorkerCount = max(1, t.config.WorkerCount/2)
			t.recommendations = append(t.recommendations, 
				fmt.Sprintf("Reduced worker count to %d due to high goroutine count (%d)", 
					t.config.WorkerCount, numGoroutines))
		}
	} else if numGoroutines < numCPU*10 && t.config.WorkerCount < numCPU*2 {
		// Few goroutines, increase worker count
		t.config.WorkerCount = min(numCPU*2, t.config.WorkerCount*2)
		t.recommendations = append(t.recommendations, 
			fmt.Sprintf("Increased worker count to %d due to low goroutine count (%d)", 
				t.config.WorkerCount, numGoroutines))
	}

	// Tune max concurrent requests based on memory usage
	heapAllocMB := memStats["heap_alloc_mb"].(float64)
	heapSysMB := memStats["heap_sys_mb"].(float64)
	
	if heapAllocMB > 1000 && t.config.MaxConcurrentRequests > numCPU*10 {
		// High memory usage, reduce max concurrent requests
		t.config.MaxConcurrentRequests = max(numCPU*10, t.config.MaxConcurrentRequests/2)
		t.recommendations = append(t.recommendations, 
			fmt.Sprintf("Reduced max concurrent requests to %d due to high memory usage (%.2f MB)", 
				t.config.MaxConcurrentRequests, heapAllocMB))
	} else if heapAllocMB < 100 && heapSysMB < 500 && t.config.MaxConcurrentRequests < numCPU*200 {
		// Low memory usage, increase max concurrent requests
		t.config.MaxConcurrentRequests = min(numCPU*200, t.config.MaxConcurrentRequests*2)
		t.recommendations = append(t.recommendations, 
			fmt.Sprintf("Increased max concurrent requests to %d due to low memory usage (%.2f MB)", 
				t.config.MaxConcurrentRequests, heapAllocMB))
	}

	// Tune connection pool size based on max concurrent requests
	t.config.ConnectionPoolSize = max(10, t.config.MaxConcurrentRequests/10)

	// Tune GC percent based on GC stats
	gcCPUFraction := gcStats["gc_cpu_fraction"].(float64)
	
	if gcCPUFraction > 0.1 && t.config.GCPercent > 50 {
		// GC is taking too much CPU time, increase GC percent
		t.config.GCPercent = min(1000, t.config.GCPercent*2)
		t.recommendations = append(t.recommendations, 
			fmt.Sprintf("Increased GC percent to %d due to high GC CPU usage (%.2f%%)", 
				t.config.GCPercent, gcCPUFraction*100))
	} else if gcCPUFraction < 0.01 && t.config.GCPercent > 25 {
		// GC is taking very little CPU time, decrease GC percent
		t.config.GCPercent = max(25, t.config.GCPercent/2)
		t.recommendations = append(t.recommendations, 
			fmt.Sprintf("Decreased GC percent to %d due to low GC CPU usage (%.2f%%)", 
				t.config.GCPercent, gcCPUFraction*100))
	}

	// Tune buffer pool size based on memory usage
	if heapAllocMB > 500 && t.config.BufferPoolSize > 100 {
		// High memory usage, reduce buffer pool size
		t.config.BufferPoolSize = max(100, t.config.BufferPoolSize/2)
		t.recommendations = append(t.recommendations, 
			fmt.Sprintf("Reduced buffer pool size to %d due to high memory usage (%.2f MB)", 
				t.config.BufferPoolSize, heapAllocMB))
	} else if heapAllocMB < 100 && t.config.BufferPoolSize < 10000 {
		// Low memory usage, increase buffer pool size
		t.config.BufferPoolSize = min(10000, t.config.BufferPoolSize*2)
		t.recommendations = append(t.recommendations, 
			fmt.Sprintf("Increased buffer pool size to %d due to low memory usage (%.2f MB)", 
				t.config.BufferPoolSize, heapAllocMB))
	}

	// Notify of configuration changes
	if len(t.recommendations) > 0 && t.onConfigChange != nil {
		t.onConfigChange(t.config)
	}
}

// GetConfig returns the current configuration
func (t *ConfigTuner) GetConfig() *TunerConfig {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.config
}

// SetConfig sets the configuration
func (t *ConfigTuner) SetConfig(config *TunerConfig) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.config = config

	// Notify of configuration changes
	if t.onConfigChange != nil {
		t.onConfigChange(t.config)
	}
}

// GetRecommendations returns tuning recommendations
func (t *ConfigTuner) GetRecommendations() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return append([]string{}, t.recommendations...)
}

// GetTuneCount returns the number of tuning operations
func (t *ConfigTuner) GetTuneCount() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.tuneCount
}

// GetLastTuneTime returns the time of the last tuning
func (t *ConfigTuner) GetLastTuneTime() time.Time {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.lastTuneTime
}

// IsRunning returns if automatic tuning is running
func (t *ConfigTuner) IsRunning() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.running
}

// GetMemoryProfiler returns the memory profiler
func (t *ConfigTuner) GetMemoryProfiler() *profiling.MemoryProfiler {
	return t.profiler
}

// SaveConfigToFile saves the configuration to a file
func (t *ConfigTuner) SaveConfigToFile(filename string) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Create file
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Write configuration
	fmt.Fprintf(f, "# Configuration generated by ConfigTuner on %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "worker_count = %d\n", t.config.WorkerCount)
	fmt.Fprintf(f, "max_concurrent_requests = %d\n", t.config.MaxConcurrentRequests)
	fmt.Fprintf(f, "connection_pool_size = %d\n", t.config.ConnectionPoolSize)
	fmt.Fprintf(f, "keep_alive_timeout = %s\n", t.config.KeepAliveTimeout)
	fmt.Fprintf(f, "read_timeout = %s\n", t.config.ReadTimeout)
	fmt.Fprintf(f, "write_timeout = %s\n", t.config.WriteTimeout)
	fmt.Fprintf(f, "idle_timeout = %s\n", t.config.IdleTimeout)
	fmt.Fprintf(f, "max_header_bytes = %d\n", t.config.MaxHeaderBytes)
	fmt.Fprintf(f, "buffer_pool_size = %d\n", t.config.BufferPoolSize)
	fmt.Fprintf(f, "template_cache = %d\n", t.config.TemplateCache)
	fmt.Fprintf(f, "static_file_cache = %d\n", t.config.StaticFileCache)
	fmt.Fprintf(f, "gzip_compression = %t\n", t.config.GzipCompression)
	fmt.Fprintf(f, "compression_level = %d\n", t.config.CompressionLevel)
	fmt.Fprintf(f, "max_request_body_size = %d\n", t.config.MaxRequestBodySize)
	fmt.Fprintf(f, "enable_http2 = %t\n", t.config.EnableHTTP2)
	fmt.Fprintf(f, "gc_percent = %d\n", t.config.GCPercent)
	fmt.Fprintf(f, "max_memory = %d\n", t.config.MaxMemory)

	return nil
}

// LoadConfigFromFile loads the configuration from a file
func (t *ConfigTuner) LoadConfigFromFile(filename string) error {
	// Not implemented - would parse the file format above
	return fmt.Errorf("not implemented")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
