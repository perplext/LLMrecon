// Package profiling provides tools for profiling and performance measurement.
package profiling

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// MetricType defines the type of metric being measured
type MetricType string

const (
	// TemplateLoadTime measures time to load templates
	TemplateLoadTime MetricType = "template_load_time"
	// TemplateRenderTime measures time to render templates
	TemplateRenderTime MetricType = "template_render_time"
	// MemoryUsage measures memory consumption
	MemoryUsage MetricType = "memory_usage"
	// CPUUsage measures CPU utilization
	CPUUsage MetricType = "cpu_usage"
	// ResponseTime measures time to generate a response
	ResponseTime MetricType = "response_time"
	// ThroughputMetric measures operations per second
	ThroughputMetric MetricType = "throughput"
	// LatencyMetric measures operation latency
	LatencyMetric MetricType = "latency"
	// ErrorRateMetric measures error rate
	ErrorRateMetric MetricType = "error_rate"
	// CacheHitRateMetric measures cache hit rate
	CacheHitRateMetric MetricType = "cache_hit_rate"
	// DatabaseQueryCount measures number of database queries
	DatabaseQueryCount MetricType = "db_query_count"
	// NetworkIOMetric measures network I/O
	NetworkIOMetric MetricType = "network_io"
	// DiskIOMetric measures disk I/O
	DiskIOMetric MetricType = "disk_io"
	// GoroutineCount measures number of goroutines
	GoroutineCount MetricType = "goroutine_count"
)

// MetricUnit defines the unit of measurement for a metric
type MetricUnit string

const (
	// Milliseconds for time measurements
	Milliseconds MetricUnit = "ms"
	// Bytes for memory measurements
	Bytes MetricUnit = "bytes"
	// Megabytes for memory measurements
	Megabytes MetricUnit = "MB"
	// Percentage for utilization measurements
	Percentage MetricUnit = "percent"
	// Count for counting measurements
	Count MetricUnit = "count"
	// OpsPerSecond for throughput measurements
	OpsPerSecond MetricUnit = "ops/sec"
	// BytesPerSecond for I/O measurements
	BytesPerSecond MetricUnit = "bytes/sec"
)

// MetricValue represents a single metric measurement
type MetricValue struct {
	// Value is the measured value
	Value float64
	// Unit is the unit of measurement
	Unit MetricUnit
	// Timestamp is when the measurement was taken
	Timestamp time.Time
	// Labels are additional metadata for the measurement
	Labels map[string]string
}

// MetricSeries represents a series of metric measurements
type MetricSeries struct {
	// Type is the type of metric
	Type MetricType
	// Description is a human-readable description of the metric
	Description string
	// Unit is the unit of measurement
	Unit MetricUnit
	// Values are the measured values
	Values []MetricValue
	// Summary contains summary statistics
	Summary MetricSummary
	// mutex protects the metric series
	mutex sync.RWMutex
}

// MetricSummary contains summary statistics for a metric series
type MetricSummary struct {
	// Min is the minimum value
	Min float64
	// Max is the maximum value
	Max float64
	// Mean is the mean value
	Mean float64
	// Median is the median value
	Median float64
	// P95 is the 95th percentile value
	P95 float64
	// P99 is the 99th percentile value
	P99 float64
	// StdDev is the standard deviation
	StdDev float64
	// Count is the number of measurements
	Count int
}

// Profiler collects and analyzes performance metrics
type Profiler struct {
	// metrics is a map of metric name to metric series
	metrics map[string]*MetricSeries
	// mutex protects the metrics map
	mutex sync.RWMutex
	// startTime is when the profiler was started
	startTime time.Time
	// cpuProfileFile is the file for CPU profiling
	cpuProfileFile *os.File
	// memProfileFile is the file for memory profiling
	memProfileFile *os.File
	// isRunning indicates if the profiler is running
	isRunning bool
	// config is the profiler configuration
	config *ProfilerConfig
}

// ProfilerConfig contains configuration for the profiler
type ProfilerConfig struct {
	// EnableCPUProfiling enables CPU profiling
	EnableCPUProfiling bool
	// EnableMemProfiling enables memory profiling
	EnableMemProfiling bool
	// CPUProfilePath is the path to save CPU profiles
	CPUProfilePath string
	// MemProfilePath is the path to save memory profiles
	MemProfilePath string
	// SamplingInterval is how often to sample metrics
	SamplingInterval time.Duration
	// MaxSamples is the maximum number of samples to keep
	MaxSamples int
	// Tags are additional metadata for all metrics
	Tags map[string]string
}

// NewProfiler creates a new profiler
func NewProfiler(config *ProfilerConfig) *Profiler {
	// Set default values
	if config == nil {
		config = &ProfilerConfig{
			EnableCPUProfiling: false,
			EnableMemProfiling: false,
			CPUProfilePath:     "cpu.pprof",
			MemProfilePath:     "mem.pprof",
			SamplingInterval:   1 * time.Second,
			MaxSamples:         1000,
			Tags:               make(map[string]string),
		}
	}

	return &Profiler{
		metrics:   make(map[string]*MetricSeries),
		startTime: time.Now(),
		config:    config,
	}
}

// Start starts the profiler
func (p *Profiler) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isRunning {
		return fmt.Errorf("profiler is already running")
	}

	// Start CPU profiling if enabled
	if p.config.EnableCPUProfiling {
		var err error
		p.cpuProfileFile, err = os.Create(p.config.CPUProfilePath)
		if err != nil {
			return fmt.Errorf("failed to create CPU profile file: %w", err)
		}
		if err := pprof.StartCPUProfile(p.cpuProfileFile); err != nil {
			p.cpuProfileFile.Close()
			return fmt.Errorf("failed to start CPU profile: %w", err)
		}
	}

	p.isRunning = true
	p.startTime = time.Now()

	return nil
}

// Stop stops the profiler
func (p *Profiler) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return fmt.Errorf("profiler is not running")
	}

	// Stop CPU profiling if enabled
	if p.config.EnableCPUProfiling {
		pprof.StopCPUProfile()
		if p.cpuProfileFile != nil {
			p.cpuProfileFile.Close()
			p.cpuProfileFile = nil
		}
	}

	// Write memory profile if enabled
	if p.config.EnableMemProfiling {
		f, err := os.Create(p.config.MemProfilePath)
		if err != nil {
			return fmt.Errorf("failed to create memory profile file: %w", err)
		}
		defer f.Close()
		runtime.GC() // Get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("failed to write memory profile: %w", err)
		}
	}

	p.isRunning = false

	return nil
}

// RegisterMetric registers a new metric
func (p *Profiler) RegisterMetric(name string, metricType MetricType, description string, unit MetricUnit) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if metric already exists
	if _, exists := p.metrics[name]; exists {
		return
	}

	// Create new metric series
	p.metrics[name] = &MetricSeries{
		Type:        metricType,
		Description: description,
		Unit:        unit,
		Values:      make([]MetricValue, 0, p.config.MaxSamples),
		Summary: MetricSummary{
			Min:   float64(^uint64(0) >> 1), // Max float64 value
			Max:   -1,
			Count: 0,
		},
	}
}

// RecordMetric records a metric value
func (p *Profiler) RecordMetric(name string, value float64, labels map[string]string) {
	p.mutex.RLock()
	metric, exists := p.metrics[name]
	p.mutex.RUnlock()

	if !exists {
		// Metric doesn't exist, create it with default values
		p.RegisterMetric(name, MetricType(name), name, Count)
		p.mutex.RLock()
		metric = p.metrics[name]
		p.mutex.RUnlock()
	}

	// Create merged labels
	mergedLabels := make(map[string]string)
	for k, v := range p.config.Tags {
		mergedLabels[k] = v
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}

	// Record metric value
	metric.mutex.Lock()
	defer metric.mutex.Unlock()

	// Create metric value
	metricValue := MetricValue{
		Value:     value,
		Unit:      metric.Unit,
		Timestamp: time.Now(),
		Labels:    mergedLabels,
	}

	// Add to values
	metric.Values = append(metric.Values, metricValue)

	// Trim values if needed
	if len(metric.Values) > p.config.MaxSamples {
		metric.Values = metric.Values[1:]
	}

	// Update summary
	metric.Summary.Count++
	if value < metric.Summary.Min {
		metric.Summary.Min = value
	}
	if value > metric.Summary.Max {
		metric.Summary.Max = value
	}

	// Recalculate mean
	sum := 0.0
	for _, v := range metric.Values {
		sum += v.Value
	}
	metric.Summary.Mean = sum / float64(len(metric.Values))

	// Update other summary statistics periodically
	if metric.Summary.Count%10 == 0 {
		p.updateMetricSummary(metric)
	}
}

// RecordDuration records a duration metric
func (p *Profiler) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	p.RecordMetric(name, float64(duration.Milliseconds()), labels)
}

// StartTimer starts a timer for measuring durations
func (p *Profiler) StartTimer(name string, labels map[string]string) func() {
	startTime := time.Now()
	return func() {
		duration := time.Since(startTime)
		p.RecordDuration(name, duration, labels)
	}
}

// GetMetric gets a metric by name
func (p *Profiler) GetMetric(name string) (*MetricSeries, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	metric, exists := p.metrics[name]
	return metric, exists
}

// GetMetrics gets all metrics
func (p *Profiler) GetMetrics() map[string]*MetricSeries {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Create a copy of the metrics map
	metrics := make(map[string]*MetricSeries, len(p.metrics))
	for name, metric := range p.metrics {
		metrics[name] = metric
	}

	return metrics
}

// GetReport generates a profiling report
func (p *Profiler) GetReport() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	report := make(map[string]interface{})
	report["uptime"] = time.Since(p.startTime).String()
	report["metric_count"] = len(p.metrics)
	report["timestamp"] = time.Now().Format(time.RFC3339)

	// Add runtime statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	report["memory"] = map[string]interface{}{
		"alloc":       memStats.Alloc,
		"total_alloc": memStats.TotalAlloc,
		"sys":         memStats.Sys,
		"num_gc":      memStats.NumGC,
	}
	report["goroutines"] = runtime.NumGoroutine()

	// Add metrics
	metricSummaries := make(map[string]map[string]interface{})
	for name, metric := range p.metrics {
		metric.mutex.RLock()
		summary := make(map[string]interface{})
		summary["type"] = string(metric.Type)
		summary["description"] = metric.Description
		summary["unit"] = string(metric.Unit)
		summary["min"] = metric.Summary.Min
		summary["max"] = metric.Summary.Max
		summary["mean"] = metric.Summary.Mean
		summary["median"] = metric.Summary.Median
		summary["p95"] = metric.Summary.P95
		summary["p99"] = metric.Summary.P99
		summary["std_dev"] = metric.Summary.StdDev
		summary["count"] = metric.Summary.Count

		// Add recent values
		recentValues := make([]map[string]interface{}, 0, 5)
		for i := len(metric.Values) - 1; i >= 0 && i >= len(metric.Values)-5; i-- {
			v := metric.Values[i]
			recentValues = append(recentValues, map[string]interface{}{
				"value":     v.Value,
				"timestamp": v.Timestamp.Format(time.RFC3339),
			})
		}
		summary["recent_values"] = recentValues
		metric.mutex.RUnlock()

		metricSummaries[name] = summary
	}
	report["metrics"] = metricSummaries

	return report
}

// SaveReport saves a profiling report to a file
func (p *Profiler) SaveReport(filePath string) error {
	report := p.GetReport()

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// Write report header
	fmt.Fprintf(file, "Profiling Report - %s\n", report["timestamp"])
	fmt.Fprintf(file, "Uptime: %s\n", report["uptime"])
	fmt.Fprintf(file, "Metric Count: %d\n", report["metric_count"])
	fmt.Fprintf(file, "Goroutines: %d\n", report["goroutines"])

	// Write memory statistics
	memStats := report["memory"].(map[string]interface{})
	fmt.Fprintf(file, "\nMemory Statistics:\n")
	fmt.Fprintf(file, "  Alloc: %d bytes\n", memStats["alloc"])
	fmt.Fprintf(file, "  Total Alloc: %d bytes\n", memStats["total_alloc"])
	fmt.Fprintf(file, "  Sys: %d bytes\n", memStats["sys"])
	fmt.Fprintf(file, "  GC Cycles: %d\n", memStats["num_gc"])

	// Write metrics
	fmt.Fprintf(file, "\nMetrics:\n")
	metrics := report["metrics"].(map[string]map[string]interface{})
	for name, metric := range metrics {
		fmt.Fprintf(file, "\n%s (%s):\n", name, metric["description"])
		fmt.Fprintf(file, "  Type: %s\n", metric["type"])
		fmt.Fprintf(file, "  Unit: %s\n", metric["unit"])
		fmt.Fprintf(file, "  Min: %.2f\n", metric["min"])
		fmt.Fprintf(file, "  Max: %.2f\n", metric["max"])
		fmt.Fprintf(file, "  Mean: %.2f\n", metric["mean"])
		fmt.Fprintf(file, "  Median: %.2f\n", metric["median"])
		fmt.Fprintf(file, "  P95: %.2f\n", metric["p95"])
		fmt.Fprintf(file, "  P99: %.2f\n", metric["p99"])
		fmt.Fprintf(file, "  StdDev: %.2f\n", metric["std_dev"])
		fmt.Fprintf(file, "  Count: %d\n", metric["count"])

		// Write recent values
		recentValues := metric["recent_values"].([]map[string]interface{})
		if len(recentValues) > 0 {
			fmt.Fprintf(file, "  Recent Values:\n")
			for _, v := range recentValues {
				fmt.Fprintf(file, "    %.2f (%s)\n", v["value"], v["timestamp"])
			}
		}
	}

	return nil
}

// CaptureMemoryProfile captures a memory profile
func (p *Profiler) CaptureMemoryProfile(filePath string) error {
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %w", err)
	}
	defer file.Close()

	// Run garbage collection to get up-to-date statistics
	runtime.GC()

	// Write heap profile
	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("failed to write memory profile: %w", err)
	}

	return nil
}

// CaptureCPUProfile captures a CPU profile
func (p *Profiler) CaptureCPUProfile(filePath string, duration time.Duration) error {
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}
	defer file.Close()

	// Start CPU profiling
	if err := pprof.StartCPUProfile(file); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	// Profile for the specified duration
	time.Sleep(duration)

	// Stop CPU profiling
	pprof.StopCPUProfile()

	return nil
}

// RecordMemoryUsage records current memory usage
func (p *Profiler) RecordMemoryUsage(name string, labels map[string]string) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	p.RecordMetric(name, float64(memStats.Alloc), labels)
}

// RecordGoroutineCount records current goroutine count
func (p *Profiler) RecordGoroutineCount(name string, labels map[string]string) {
	p.RecordMetric(name, float64(runtime.NumGoroutine()), labels)
}

// updateMetricSummary updates the summary statistics for a metric
func (p *Profiler) updateMetricSummary(metric *MetricSeries) {
	// Sort values for percentile calculations
	values := make([]float64, len(metric.Values))
	for i, v := range metric.Values {
		values[i] = v.Value
	}
	sortFloat64s(values)

	// Calculate median
	if len(values) > 0 {
		if len(values)%2 == 0 {
			metric.Summary.Median = (values[len(values)/2-1] + values[len(values)/2]) / 2
		} else {
			metric.Summary.Median = values[len(values)/2]
		}
	}

	// Calculate percentiles
	if len(values) > 0 {
		p95Index := int(float64(len(values)) * 0.95)
		if p95Index >= len(values) {
			p95Index = len(values) - 1
		}
		metric.Summary.P95 = values[p95Index]

		p99Index := int(float64(len(values)) * 0.99)
		if p99Index >= len(values) {
			p99Index = len(values) - 1
		}
		metric.Summary.P99 = values[p99Index]
	}

	// Calculate standard deviation
	if len(values) > 1 {
		sumSquaredDiff := 0.0
		for _, v := range values {
			diff := v - metric.Summary.Mean
			sumSquaredDiff += diff * diff
		}
		metric.Summary.StdDev = (sumSquaredDiff / float64(len(values)-1))
	}
}

// sortFloat64s sorts a slice of float64 values
func sortFloat64s(values []float64) {
	// Simple bubble sort for small slices
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}
