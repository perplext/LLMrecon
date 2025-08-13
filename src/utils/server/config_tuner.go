package server

import (
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/perplext/LLMrecon/src/utils/config"
	"github.com/perplext/LLMrecon/src/utils/monitoring"
)

// ServerConfigTuner provides automatic tuning of server configuration parameters
// based on system load and resource utilization metrics
type ServerConfigTuner struct {
	// config is the memory configuration
	config *config.MemoryConfig
	// metricsManager is used to collect and analyze metrics
	metricsManager *monitoring.MetricsManager
	// logger is the tuner logger
	logger *log.Logger
	// mutex protects configuration changes
	mutex sync.Mutex
	// lastTuneTime is the time of the last tuning operation
	lastTuneTime time.Time
	// tuningInterval is the minimum time between tuning operations
	tuningInterval time.Duration
	// autoTuneEnabled indicates if automatic tuning is enabled
	autoTuneEnabled bool
	// stopChan is used to stop the auto-tuning goroutine
	stopChan chan struct{}
}

// ServerConfigTunerOptions contains options for the server configuration tuner
type ServerConfigTunerOptions struct {
	// TuningInterval is the interval at which automatic tuning occurs
	TuningInterval time.Duration
	// AutoTuneEnabled indicates if automatic tuning is enabled
	AutoTuneEnabled bool
	// LogFile is the file to log tuning operations to
	LogFile string
}

// DefaultServerConfigTunerOptions returns default options for the server configuration tuner
func DefaultServerConfigTunerOptions() *ServerConfigTunerOptions {
	return &ServerConfigTunerOptions{
		TuningInterval:  5 * time.Minute,
		AutoTuneEnabled: true,
		LogFile:         "logs/server_tuner.log",
	}
}

// NewServerConfigTuner creates a new server configuration tuner
func NewServerConfigTuner(metricsManager *monitoring.MetricsManager, options *ServerConfigTunerOptions) (*ServerConfigTuner, error) {
	if options == nil {
		options = DefaultServerConfigTunerOptions()
	}

	// Create logger
	var logger *log.Logger
	if options.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file
		logFile, err := os.OpenFile(options.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		// Log to both file and console
		multiWriter := monitoring.NewMultiWriter(logFile, os.Stderr)
		logger = log.New(multiWriter, "[SERVER TUNER] ", log.LstdFlags)
	} else {
		// Log to console only
		logger = log.New(os.Stderr, "[SERVER TUNER] ", log.LstdFlags)
	}

	// Get memory configuration
	memConfig := config.GetMemoryConfig()

	// Create server configuration tuner
	tuner := &ServerConfigTuner{
		config:         memConfig,
		metricsManager: metricsManager,
		logger:         logger,
		lastTuneTime:   time.Now(),
		tuningInterval: options.TuningInterval,
		autoTuneEnabled: options.AutoTuneEnabled,
		stopChan:       make(chan struct{}),
	}

	// Start automatic tuning if enabled
	if options.AutoTuneEnabled {
		tuner.StartAutoTuning()
	}

	return tuner, nil
}

// StartAutoTuning starts automatic tuning at the configured interval
func (t *ServerConfigTuner) StartAutoTuning() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.autoTuneEnabled {
		go func() {
			ticker := time.NewTicker(t.tuningInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					t.TuneServerConfig()
				case <-t.stopChan:
					return
				}
			}
		}()

		t.logger.Println("Automatic server configuration tuning started")
	}
}

// StopAutoTuning stops automatic tuning
func (t *ServerConfigTuner) StopAutoTuning() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.autoTuneEnabled {
		close(t.stopChan)
		t.autoTuneEnabled = false
		t.logger.Println("Automatic server configuration tuning stopped")
	}
}

// TuneServerConfig tunes server configuration parameters based on system metrics
func (t *ServerConfigTuner) TuneServerConfig() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Check if enough time has passed since the last tuning
	if time.Since(t.lastTuneTime) < t.tuningInterval {
		return
	}

	t.logger.Println("Tuning server configuration...")

	// Get system metrics
	memoryMetrics := t.metricsManager.GetMetricsByPrefix("system.memory")
	gcMetrics := t.metricsManager.GetMetricsByPrefix("system.gc")
	concurrencyMetrics := t.metricsManager.GetMetricsByPrefix("concurrency")

	// Tune worker count
	t.tuneWorkerCount(concurrencyMetrics)

	// Tune connection pool size
	t.tuneConnectionPoolSize(concurrencyMetrics)

	// Tune GC percentage
	t.tuneGCPercentage(memoryMetrics, gcMetrics)

	// Tune buffer sizes
	t.tuneBufferSizes(memoryMetrics)

	// Update last tune time
	t.lastTuneTime = time.Now()

	t.logger.Println("Server configuration tuning completed")
}

// tuneWorkerCount tunes the number of worker processes based on concurrency metrics
func (t *ServerConfigTuner) tuneWorkerCount(metrics map[string]*monitoring.Metric) {
	// Get current worker metrics
	workerUtilization, ok := metrics["concurrency.worker.utilization"]
	if !ok {
		return
	}

	// Get current worker count
	currentMaxWorkers := t.config.MaxWorkers
	currentMinWorkers := t.config.MinWorkers

	// Calculate optimal worker count based on utilization
	numCPU := runtime.NumCPU()
	utilization := workerUtilization.Value

	var newMaxWorkers, newMinWorkers int

	if utilization > 0.8 {
		// High utilization, increase worker count
		newMaxWorkers = int(float64(currentMaxWorkers) * 1.25)
		// Cap at a reasonable multiple of CPU count
		if newMaxWorkers > numCPU*8 {
			newMaxWorkers = numCPU * 8
		}
		newMinWorkers = int(float64(currentMinWorkers) * 1.2)
	} else if utilization < 0.3 {
		// Low utilization, decrease worker count
		newMaxWorkers = int(float64(currentMaxWorkers) * 0.8)
		// Ensure minimum reasonable value
		if newMaxWorkers < numCPU {
			newMaxWorkers = numCPU
		}
		newMinWorkers = int(float64(currentMinWorkers) * 0.8)
		if newMinWorkers < 1 {
			newMinWorkers = 1
		}
	} else {
		// Utilization is in a good range, no change needed
		return
	}

	// Ensure min workers is less than max workers
	if newMinWorkers >= newMaxWorkers {
		newMinWorkers = newMaxWorkers / 2
		if newMinWorkers < 1 {
			newMinWorkers = 1
		}
	}

	// Update configuration if changed
	if newMaxWorkers != currentMaxWorkers || newMinWorkers != currentMinWorkers {
		t.logger.Printf("Tuning worker count: min=%d→%d, max=%d→%d (utilization: %.2f)",
			currentMinWorkers, newMinWorkers, currentMaxWorkers, newMaxWorkers, utilization)

		t.config.MaxWorkers = newMaxWorkers
		t.config.MinWorkers = newMinWorkers

		// Save configuration
		if err := t.config.SaveConfig(); err != nil {
			t.logger.Printf("Failed to save configuration: %v", err)
		}

		// Apply changes to environment variables for immediate effect
		os.Setenv("MAX_WORKERS", strconv.Itoa(newMaxWorkers))
		os.Setenv("MIN_WORKERS", strconv.Itoa(newMinWorkers))
	}
}

// tuneConnectionPoolSize tunes the connection pool size based on concurrency metrics
func (t *ServerConfigTuner) tuneConnectionPoolSize(metrics map[string]*monitoring.Metric) {
	// Get current connection pool metrics
	poolUtilization, ok := metrics["resource.pool.utilization"]
	if !ok {
		return
	}

	// Get current connection pool size
	currentPoolSize := t.config.ConnectionPoolSize

	// Calculate optimal pool size based on utilization
	utilization := poolUtilization.Value
	var newPoolSize int

	if utilization > 0.8 {
		// High utilization, increase pool size
		newPoolSize = int(float64(currentPoolSize) * 1.25)
		// Cap at a reasonable value
		if newPoolSize > 200 {
			newPoolSize = 200
		}
	} else if utilization < 0.3 {
		// Low utilization, decrease pool size
		newPoolSize = int(float64(currentPoolSize) * 0.8)
		// Ensure minimum reasonable value
		if newPoolSize < 5 {
			newPoolSize = 5
		}
	} else {
		// Utilization is in a good range, no change needed
		return
	}

	// Update configuration if changed
	if newPoolSize != currentPoolSize {
		t.logger.Printf("Tuning connection pool size: %d→%d (utilization: %.2f)",
			currentPoolSize, newPoolSize, utilization)

		t.config.ConnectionPoolSize = newPoolSize

		// Save configuration
		if err := t.config.SaveConfig(); err != nil {
			t.logger.Printf("Failed to save configuration: %v", err)
		}

		// Apply changes to environment variables for immediate effect
		os.Setenv("CONNECTION_POOL_SIZE", strconv.Itoa(newPoolSize))
	}
}

// tuneGCPercentage tunes the garbage collection percentage based on memory metrics
func (t *ServerConfigTuner) tuneGCPercentage(memoryMetrics, gcMetrics map[string]*monitoring.Metric) {
	// Get current GC metrics
	gcCPUFraction, ok := gcMetrics["system.gc.cpu_fraction"]
	if !ok {
		return
	}

	// Get current GC percentage
	currentGCPercent := t.config.GCPercent

	// Calculate optimal GC percentage based on CPU fraction
	gcCPUUsage := gcCPUFraction.Value
	var newGCPercent int

	if gcCPUUsage > 0.1 {
		// GC is using too much CPU, increase GC percentage to run less frequently
		newGCPercent = int(float64(currentGCPercent) * 1.2)
		// Cap at a reasonable value
		if newGCPercent > 500 {
			newGCPercent = 500
		}
	} else if gcCPUUsage < 0.01 && currentGCPercent > 100 {
		// GC is using very little CPU and current percentage is high, decrease to run more frequently
		newGCPercent = int(float64(currentGCPercent) * 0.9)
		// Ensure minimum reasonable value
		if newGCPercent < 50 {
			newGCPercent = 50
		}
	} else {
		// GC CPU usage is in a good range, no change needed
		return
	}

	// Update configuration if changed
	if newGCPercent != currentGCPercent {
		t.logger.Printf("Tuning GC percentage: %d→%d (GC CPU usage: %.4f)",
			currentGCPercent, newGCPercent, gcCPUUsage)

		t.config.GCPercent = newGCPercent

		// Save configuration
		if err := t.config.SaveConfig(); err != nil {
			t.logger.Printf("Failed to save configuration: %v", err)
		}

		// Apply changes to runtime for immediate effect
		debug.SetGCPercent(newGCPercent)
	}
}

// tuneBufferSizes tunes buffer sizes based on memory metrics
func (t *ServerConfigTuner) tuneBufferSizes(metrics map[string]*monitoring.Metric) {
	// Get current memory metrics
	heapAlloc, ok := metrics["system.memory.heap_alloc"]
	if !ok {
		return
	}

	// Get current buffer pool size
	currentBufferPoolSize := t.config.BufferPoolSize

	// Calculate optimal buffer pool size based on memory usage
	heapAllocMB := heapAlloc.Value / (1024 * 1024)
	memoryThresholdMB := float64(t.config.MemoryThreshold)
	memoryUsageRatio := heapAllocMB / memoryThresholdMB
	var newBufferPoolSize int

	if memoryUsageRatio > 0.8 {
		// Memory usage is high, decrease buffer pool size
		newBufferPoolSize = int(float64(currentBufferPoolSize) * 0.8)
		// Ensure minimum reasonable value
		if newBufferPoolSize < 100 {
			newBufferPoolSize = 100
		}
	} else if memoryUsageRatio < 0.4 {
		// Memory usage is low, increase buffer pool size
		newBufferPoolSize = int(float64(currentBufferPoolSize) * 1.2)
		// Cap at a reasonable value
		if newBufferPoolSize > 10000 {
			newBufferPoolSize = 10000
		}
	} else {
		// Memory usage is in a good range, no change needed
		return
	}

	// Update configuration if changed
	if newBufferPoolSize != currentBufferPoolSize {
		t.logger.Printf("Tuning buffer pool size: %d→%d (memory usage: %.2f MB, %.2f%% of threshold)",
			currentBufferPoolSize, newBufferPoolSize, heapAllocMB, memoryUsageRatio*100)

		t.config.BufferPoolSize = newBufferPoolSize

		// Save configuration
		if err := t.config.SaveConfig(); err != nil {
			t.logger.Printf("Failed to save configuration: %v", err)
		}

		// Apply changes to environment variables for immediate effect
		os.Setenv("BUFFER_POOL_SIZE", strconv.Itoa(newBufferPoolSize))
	}
}

// GetRecommendations returns recommendations for server configuration
func (t *ServerConfigTuner) GetRecommendations() map[string]string {
	recommendations := make(map[string]string)

	// Get system metrics
	memoryMetrics := t.metricsManager.GetMetricsByPrefix("system.memory")
	gcMetrics := t.metricsManager.GetMetricsByPrefix("system.gc")
	concurrencyMetrics := t.metricsManager.GetMetricsByPrefix("concurrency")

	// Worker count recommendation
	if workerUtilization, ok := concurrencyMetrics["concurrency.worker.utilization"]; ok {
		utilization := workerUtilization.Value
		if utilization > 0.8 {
			recommendations["worker_count"] = fmt.Sprintf(
				"Worker utilization is high (%.2f). Consider increasing max_workers from %d to %d.",
				utilization, t.config.MaxWorkers, int(float64(t.config.MaxWorkers)*1.25))
		} else if utilization < 0.3 {
			recommendations["worker_count"] = fmt.Sprintf(
				"Worker utilization is low (%.2f). Consider decreasing max_workers from %d to %d to save resources.",
				utilization, t.config.MaxWorkers, int(float64(t.config.MaxWorkers)*0.8))
		}
	}

	// Connection pool recommendation
	if poolUtilization, ok := concurrencyMetrics["resource.pool.utilization"]; ok {
		utilization := poolUtilization.Value
		if utilization > 0.8 {
			recommendations["connection_pool"] = fmt.Sprintf(
				"Connection pool utilization is high (%.2f). Consider increasing connection_pool_size from %d to %d.",
				utilization, t.config.ConnectionPoolSize, int(float64(t.config.ConnectionPoolSize)*1.25))
		} else if utilization < 0.3 {
			recommendations["connection_pool"] = fmt.Sprintf(
				"Connection pool utilization is low (%.2f). Consider decreasing connection_pool_size from %d to %d to save resources.",
				utilization, t.config.ConnectionPoolSize, int(float64(t.config.ConnectionPoolSize)*0.8))
		}
	}

	// GC percentage recommendation
	if gcCPUFraction, ok := gcMetrics["system.gc.cpu_fraction"]; ok {
		gcCPUUsage := gcCPUFraction.Value
		if gcCPUUsage > 0.1 {
			recommendations["gc_percent"] = fmt.Sprintf(
				"GC is using %.2f%% of CPU time. Consider increasing gc_percent from %d to %d to reduce GC frequency.",
				gcCPUUsage*100, t.config.GCPercent, int(float64(t.config.GCPercent)*1.2))
		} else if gcCPUUsage < 0.01 && t.config.GCPercent > 100 {
			recommendations["gc_percent"] = fmt.Sprintf(
				"GC is using only %.2f%% of CPU time. Consider decreasing gc_percent from %d to %d for more frequent garbage collection.",
				gcCPUUsage*100, t.config.GCPercent, int(float64(t.config.GCPercent)*0.9))
		}
	}

	// Buffer size recommendation
	if heapAlloc, ok := memoryMetrics["system.memory.heap_alloc"]; ok {
		heapAllocMB := heapAlloc.Value / (1024 * 1024)
		memoryThresholdMB := float64(t.config.MemoryThreshold)
		memoryUsageRatio := heapAllocMB / memoryThresholdMB
		if memoryUsageRatio > 0.8 {
			recommendations["buffer_size"] = fmt.Sprintf(
				"Memory usage is high (%.2f MB, %.2f%% of threshold). Consider decreasing buffer_pool_size from %d to %d.",
				heapAllocMB, memoryUsageRatio*100, t.config.BufferPoolSize, int(float64(t.config.BufferPoolSize)*0.8))
		} else if memoryUsageRatio < 0.4 {
			recommendations["buffer_size"] = fmt.Sprintf(
				"Memory usage is low (%.2f MB, %.2f%% of threshold). Consider increasing buffer_pool_size from %d to %d for better performance.",
				heapAllocMB, memoryUsageRatio*100, t.config.BufferPoolSize, int(float64(t.config.BufferPoolSize)*1.2))
		}
	}

	return recommendations
}

// ApplyRecommendations applies all current recommendations
func (t *ServerConfigTuner) ApplyRecommendations() {
	t.TuneServerConfig()
}

// GetCurrentConfig returns the current server configuration
func (t *ServerConfigTuner) GetCurrentConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	config["max_workers"] = t.config.MaxWorkers
	config["min_workers"] = t.config.MinWorkers
	config["connection_pool_size"] = t.config.ConnectionPoolSize
	config["gc_percent"] = t.config.GCPercent
	config["buffer_pool_size"] = t.config.BufferPoolSize
	config["max_concurrent_requests"] = t.config.MaxConcurrentRequests
	
	return config
}

// SetAutoTuneEnabled enables or disables automatic tuning
func (t *ServerConfigTuner) SetAutoTuneEnabled(enabled bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if enabled && !t.autoTuneEnabled {
		t.autoTuneEnabled = true
		t.stopChan = make(chan struct{})
		t.StartAutoTuning()
	} else if !enabled && t.autoTuneEnabled {
		close(t.stopChan)
		t.autoTuneEnabled = false
	}
}

// IsAutoTuneEnabled returns true if automatic tuning is enabled
func (t *ServerConfigTuner) IsAutoTuneEnabled() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	return t.autoTuneEnabled
}
