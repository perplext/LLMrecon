package analytics

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/management/execution"
)

// MetricsCollector handles the collection and processing of various metrics
type MetricsCollector struct {
	config      *Config
	storage     DataStorage
	processors  map[string]MetricProcessor
	aggregators map[string]MetricAggregator
	hooks       []CollectionHook
	logger      Logger
	mu          sync.RWMutex
	
	// Real-time metrics
	activeScans    map[string]*ScanTracker
	systemMetrics  *SystemMetrics
	metricsBuffer  chan Metric
	
	// Collection state
	enabled       bool
	batchSize     int
	flushInterval time.Duration
	
	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// MetricProcessor interface for processing individual metrics
type MetricProcessor interface {
	Process(ctx context.Context, metric Metric) (Metric, error)
	GetType() string
	IsEnabled() bool
}

// MetricAggregator interface for aggregating metrics over time windows
type MetricAggregator interface {
	Aggregate(ctx context.Context, metrics []Metric, window TimeWindow) (AggregatedMetric, error)
	GetWindowSizes() []time.Duration
	Reset()
}

// CollectionHook allows custom logic during metric collection
type CollectionHook interface {
	PreCollection(ctx context.Context, scanID string) error
	PostCollection(ctx context.Context, scanID string, metrics []Metric) error
	OnError(ctx context.Context, err error, scanID string)
}

// ScanTracker tracks metrics for an active scan
type ScanTracker struct {
	ScanID        string
	StartTime     time.Time
	LastUpdate    time.Time
	TestsRun      int
	TestsPassed   int
	TestsFailed   int
	TemplatesUsed []string
	CurrentPhase  string
	Metrics       []Metric
	mu            sync.RWMutex
}

// SystemMetrics tracks system-level performance metrics
type SystemMetrics struct {
	CPUUsage      float64
	MemoryUsage   int64
	DiskUsage     int64
	NetworkIO     NetworkStats
	ProcessCount  int
	ThreadCount   int
	LastUpdated   time.Time
	mu            sync.RWMutex
}

// NetworkStats tracks network I/O statistics
type NetworkStats struct {
	BytesSent     int64
	BytesReceived int64
	PacketsSent   int64
	PacketsReceived int64
	Connections   int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *Config, storage DataStorage, logger Logger) *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())
	
	collector := &MetricsCollector{
		config:        config,
		storage:       storage,
		processors:    make(map[string]MetricProcessor),
		aggregators:   make(map[string]MetricAggregator),
		hooks:         make([]CollectionHook, 0),
		logger:        logger,
		activeScans:   make(map[string]*ScanTracker),
		systemMetrics: &SystemMetrics{},
		metricsBuffer: make(chan Metric, config.Analytics.BufferSize),
		enabled:       config.Analytics.CollectionEnabled,
		batchSize:     config.Analytics.BatchSize,
		flushInterval: config.Analytics.FlushInterval,
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// Register default processors
	collector.registerDefaultProcessors()
	collector.registerDefaultAggregators()
	
	// Start background workers
	if collector.enabled {
		collector.startWorkers()
	}
	
	return collector

// StartScanTracking begins tracking metrics for a new scan
func (mc *MetricsCollector) StartScanTracking(scanID, target string, templates []string) error {
	if !mc.enabled {
		return nil
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Run pre-collection hooks
	for _, hook := range mc.hooks {
		if err := hook.PreCollection(mc.ctx, scanID); err != nil {
			mc.logger.Warn("Pre-collection hook failed", "scanID", scanID, "error", err)
		}
	}
	
	tracker := &ScanTracker{
		ScanID:        scanID,
		StartTime:     time.Now(),
		LastUpdate:    time.Now(),
		TemplatesUsed: templates,
		CurrentPhase:  "initialization",
		Metrics:       make([]Metric, 0),
	}
	
	mc.activeScans[scanID] = tracker
	
	// Emit scan started metric
	metric := Metric{
		ID:        generateMetricID(),
		Name:      "scan_started",
		Type:      MetricTypeEvent,
		Value:     1,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"scan_id": scanID,
			"target":  target,
		},
		Metadata: map[string]interface{}{
			"templates_count": len(templates),
			"templates":       templates,
		},
	}
	
	return mc.collectMetric(metric)

// UpdateScanProgress updates the progress of an active scan
func (mc *MetricsCollector) UpdateScanProgress(scanID string, phase string, testsRun, testsPassed, testsFailed int) error {
	if !mc.enabled {
		return nil
	}
	
	mc.mu.Lock()
	tracker, exists := mc.activeScans[scanID]
	if !exists {
		mc.mu.Unlock()
		return fmt.Errorf("scan tracker not found for ID: %s", scanID)
	}
	
	tracker.mu.Lock()
	tracker.CurrentPhase = phase
	tracker.TestsRun = testsRun
	tracker.TestsPassed = testsPassed
	tracker.TestsFailed = testsFailed
	tracker.LastUpdate = time.Now()
	tracker.mu.Unlock()
	mc.mu.Unlock()
	
	// Emit progress metric
	metric := Metric{
		ID:        generateMetricID(),
		Name:      "scan_progress",
		Type:      MetricTypeGauge,
		Value:     float64(testsRun),
		Timestamp: time.Now(),
		Labels: map[string]string{
			"scan_id": scanID,
			"phase":   phase,
		},
		Metadata: map[string]interface{}{
			"tests_passed": testsPassed,
			"tests_failed": testsFailed,
			"success_rate": calculateSuccessRate(testsPassed, testsFailed),
		},
	}
	
	return mc.collectMetric(metric)

// FinishScanTracking completes tracking for a scan
func (mc *MetricsCollector) FinishScanTracking(scanID string, result *ScanResult) error {
	if !mc.enabled {
		return nil
	}
	
	mc.mu.Lock()
	tracker, exists := mc.activeScans[scanID]
	if !exists {
		mc.mu.Unlock()
		return fmt.Errorf("scan tracker not found for ID: %s", scanID)
	}
	
	duration := time.Since(tracker.StartTime)
	delete(mc.activeScans, scanID)
	mc.mu.Unlock()
	
	// Generate completion metrics
	metrics := mc.generateScanCompletionMetrics(scanID, tracker, result, duration)
	
	// Collect all metrics
	for _, metric := range metrics {
		if err := mc.collectMetric(metric); err != nil {
			mc.logger.Error("Failed to collect scan completion metric", "error", err)
		}
	}
	
	// Run post-collection hooks
	for _, hook := range mc.hooks {
		if err := hook.PostCollection(mc.ctx, scanID, metrics); err != nil {
			mc.logger.Warn("Post-collection hook failed", "scanID", scanID, "error", err)
		}
	}
	
	return nil

// CollectCustomMetric allows collection of custom metrics
func (mc *MetricsCollector) CollectCustomMetric(name string, value float64, labels map[string]string, metadata map[string]interface{}) error {
	if !mc.enabled {
		return nil
	}
	
	metric := Metric{
		ID:        generateMetricID(),
		Name:      name,
		Type:      MetricTypeCustom,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
		Metadata:  metadata,
	}
	
	return mc.collectMetric(metric)

// GetActiveScans returns information about currently active scans
func (mc *MetricsCollector) GetActiveScans() map[string]*ScanTracker {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	result := make(map[string]*ScanTracker)
	for id, tracker := range mc.activeScans {
		result[id] = tracker
	}
	
	return result

// GetSystemMetrics returns current system metrics
func (mc *MetricsCollector) GetSystemMetrics() *SystemMetrics {
	mc.systemMetrics.mu.RLock()
	defer mc.systemMetrics.mu.RUnlock()
	
	return &SystemMetrics{
		CPUUsage:     mc.systemMetrics.CPUUsage,
		MemoryUsage:  mc.systemMetrics.MemoryUsage,
		DiskUsage:    mc.systemMetrics.DiskUsage,
		NetworkIO:    mc.systemMetrics.NetworkIO,
		ProcessCount: mc.systemMetrics.ProcessCount,
		ThreadCount:  mc.systemMetrics.ThreadCount,
		LastUpdated:  mc.systemMetrics.LastUpdated,
	}

// RegisterProcessor adds a custom metric processor
func (mc *MetricsCollector) RegisterProcessor(processor MetricProcessor) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.processors[processor.GetType()] = processor
	mc.logger.Info("Registered metric processor", "type", processor.GetType())

// RegisterAggregator adds a custom metric aggregator
func (mc *MetricsCollector) RegisterAggregator(name string, aggregator MetricAggregator) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.aggregators[name] = aggregator
	mc.logger.Info("Registered metric aggregator", "name", name)

// RegisterHook adds a collection hook
func (mc *MetricsCollector) RegisterHook(hook CollectionHook) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.hooks = append(mc.hooks, hook)
	mc.logger.Info("Registered collection hook")

// Shutdown gracefully shuts down the metrics collector
func (mc *MetricsCollector) Shutdown(timeout time.Duration) error {
	mc.logger.Info("Shutting down metrics collector")
	
	mc.cancel()
	
	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		mc.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		mc.logger.Info("Metrics collector shut down successfully")
		return nil
	case <-time.After(timeout):
		mc.logger.Warn("Metrics collector shutdown timed out")
		return fmt.Errorf("shutdown timed out after %v", timeout)
	}

// Internal methods

func (mc *MetricsCollector) collectMetric(metric Metric) error {
	// Process the metric through registered processors
	processedMetric := metric
	for _, processor := range mc.processors {
		if processor.IsEnabled() {
			var err error
			processedMetric, err = processor.Process(mc.ctx, processedMetric)
			if err != nil {
				mc.logger.Error("Metric processing failed", "processor", processor.GetType(), "error", err)
				continue
			}
		}
	}
	
	// Send to buffer for batch processing
	select {
	case mc.metricsBuffer <- processedMetric:
		return nil
	case <-mc.ctx.Done():
		return mc.ctx.Err()
	default:
		mc.logger.Warn("Metrics buffer full, dropping metric", "metric", metric.Name)
		return fmt.Errorf("metrics buffer full")
	}

func (mc *MetricsCollector) startWorkers() {
	// Start metrics processing worker
	mc.wg.Add(1)
	go mc.metricsProcessingWorker()
	
	// Start system metrics collection worker
	mc.wg.Add(1)
	go mc.systemMetricsWorker()
	
	// Start aggregation worker
	mc.wg.Add(1)
	go mc.aggregationWorker()

func (mc *MetricsCollector) metricsProcessingWorker() {
	defer mc.wg.Done()
	
	ticker := time.NewTicker(mc.flushInterval)
	defer ticker.Stop()
	
	batch := make([]Metric, 0, mc.batchSize)
	
	for {
		select {
		case metric := <-mc.metricsBuffer:
			batch = append(batch, metric)
			
			if len(batch) >= mc.batchSize {
				mc.flushBatch(batch)
				batch = batch[:0]
			}
			
		case <-ticker.C:
			if len(batch) > 0 {
				mc.flushBatch(batch)
				batch = batch[:0]
			}
			
		case <-mc.ctx.Done():
			// Flush remaining metrics
			if len(batch) > 0 {
				mc.flushBatch(batch)
			}
			return
		}
	}

func (mc *MetricsCollector) systemMetricsWorker() {
	defer mc.wg.Done()
	
	ticker := time.NewTicker(time.Duration(mc.config.Analytics.SystemMetricsInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.collectSystemMetrics()
		case <-mc.ctx.Done():
			return
		}
	}

func (mc *MetricsCollector) aggregationWorker() {
	defer mc.wg.Done()
	
	ticker := time.NewTicker(time.Duration(mc.config.Analytics.AggregationInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.runAggregation()
		case <-mc.ctx.Done():
			return
		}
	}

func (mc *MetricsCollector) flushBatch(batch []Metric) {
	if err := mc.storage.StoreMetrics(mc.ctx, batch); err != nil {
		mc.logger.Error("Failed to store metrics batch", "size", len(batch), "error", err)
		
		// Notify hooks of error
		for _, hook := range mc.hooks {
			hook.OnError(mc.ctx, err, "batch_storage")
		}
	}

func (mc *MetricsCollector) collectSystemMetrics() {
	// This would integrate with actual system monitoring
	// For now, we'll simulate some metrics
	mc.systemMetrics.mu.Lock()
	defer mc.systemMetrics.mu.Unlock()
	
	mc.systemMetrics.CPUUsage = getCurrentCPUUsage()
	mc.systemMetrics.MemoryUsage = getCurrentMemoryUsage()
	mc.systemMetrics.DiskUsage = getCurrentDiskUsage()
	mc.systemMetrics.NetworkIO = getCurrentNetworkStats()
	mc.systemMetrics.ProcessCount = getCurrentProcessCount()
	mc.systemMetrics.ThreadCount = getCurrentThreadCount()
	mc.systemMetrics.LastUpdated = time.Now()
	
	// Store as metrics
	systemMetrics := []Metric{
		{
			ID:        generateMetricID(),
			Name:      "system_cpu_usage",
			Type:      MetricTypeGauge,
			Value:     mc.systemMetrics.CPUUsage,
			Timestamp: time.Now(),
			Labels:    map[string]string{"component": "system"},
		},
		{
			ID:        generateMetricID(),
			Name:      "system_memory_usage",
			Type:      MetricTypeGauge,
			Value:     float64(mc.systemMetrics.MemoryUsage),
			Timestamp: time.Now(),
			Labels:    map[string]string{"component": "system"},
		},
	}
	
	for _, metric := range systemMetrics {
		mc.collectMetric(metric)
	}

func (mc *MetricsCollector) runAggregation() {
	now := time.Now()
	
	for name, aggregator := range mc.aggregators {
		for _, windowSize := range aggregator.GetWindowSizes() {
			startTime := now.Add(-windowSize)
			
			// Get metrics for the time window
			metrics, err := mc.storage.GetMetricsByTimeRange(mc.ctx, startTime, now)
			if err != nil {
				mc.logger.Error("Failed to get metrics for aggregation", "aggregator", name, "error", err)
				continue
			}
			
			// Run aggregation
			window := TimeWindow{
				Start:    startTime,
				End:      now,
				Duration: windowSize,
			}
			
			aggregated, err := aggregator.Aggregate(mc.ctx, metrics, window)
			if err != nil {
				mc.logger.Error("Aggregation failed", "aggregator", name, "error", err)
				continue
			}
			
			// Store aggregated metric
			if err := mc.storage.StoreAggregatedMetric(mc.ctx, aggregated); err != nil {
				mc.logger.Error("Failed to store aggregated metric", "error", err)
			}
		}
	}

func (mc *MetricsCollector) generateScanCompletionMetrics(scanID string, tracker *ScanTracker, result *ScanResult, duration time.Duration) []Metric {
	baseLabels := map[string]string{
		"scan_id": scanID,
		"target":  result.Target,
	}
	
	metrics := []Metric{
		{
			ID:        generateMetricID(),
			Name:      "scan_completed",
			Type:      MetricTypeEvent,
			Value:     1,
			Timestamp: time.Now(),
			Labels:    baseLabels,
			Metadata: map[string]interface{}{
				"duration_seconds": duration.Seconds(),
				"success":          result.Success,
			},
		},
		{
			ID:        generateMetricID(),
			Name:      "scan_duration",
			Type:      MetricTypeHistogram,
			Value:     duration.Seconds(),
			Timestamp: time.Now(),
			Labels:    baseLabels,
		},
		{
			ID:        generateMetricID(),
			Name:      "scan_tests_total",
			Type:      MetricTypeCounter,
			Value:     float64(result.TotalTests),
			Timestamp: time.Now(),
			Labels:    baseLabels,
		},
		{
			ID:        generateMetricID(),
			Name:      "scan_vulnerabilities_found",
			Type:      MetricTypeCounter,
			Value:     float64(len(result.Vulnerabilities)),
			Timestamp: time.Now(),
			Labels:    baseLabels,
		},
	}
	
	return metrics

func (mc *MetricsCollector) registerDefaultProcessors() {
	// Add default processors
	mc.processors["validation"] = &ValidationProcessor{}
	mc.processors["enrichment"] = &EnrichmentProcessor{}
	mc.processors["filtering"] = &FilteringProcessor{config: mc.config}

func (mc *MetricsCollector) registerDefaultAggregators() {
	// Add default aggregators
	mc.aggregators["basic"] = &BasicAggregator{}
	mc.aggregators["performance"] = &PerformanceAggregator{}
	mc.aggregators["security"] = &SecurityAggregator{}

// Utility functions
func generateMetricID() string {
	return fmt.Sprintf("metric_%d_%d", time.Now().UnixNano(), time.Now().Unix())

func calculateSuccessRate(passed, failed int) float64 {
	total := passed + failed
	if total == 0 {
		return 0.0
	}
	return float64(passed) / float64(total) * 100.0

// System metrics functions (would be replaced with actual system calls)
func getCurrentCPUUsage() float64     { return 45.2 } // Mock implementation
func getCurrentMemoryUsage() int64    { return 1024 * 1024 * 512 } // Mock: 512MB
func getCurrentDiskUsage() int64      { return 1024 * 1024 * 1024 * 10 } // Mock: 10GB
func getCurrentNetworkStats() NetworkStats {
	return NetworkStats{
		BytesSent:       1024 * 1024,
		BytesReceived:   2 * 1024 * 1024,
		PacketsSent:     1000,
		PacketsReceived: 1500,
		Connections:     25,
	}
func getCurrentProcessCount() int { return 120 }
func getCurrentThreadCount() int  { return 480 }
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
