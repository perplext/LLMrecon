package monitoring

import (
	"fmt"
	"sync"
	"time"
)

// StaticFileMetrics contains metrics for the static file handler
type StaticFileMetrics struct {
	FilesServed      int64
	CacheHits        int64
	CacheMisses      int64
	CacheHitRatio    float64
	CompressedFiles  int64
	TotalSize        int64
	CompressedSize   int64
	CompressionRatio float64
	CacheSize        int64
	CacheItemCount   int64
	AverageServeTime time.Duration
}

// StaticFileMonitor monitors a static file handler
type StaticFileMonitor struct {
	fileHandler    FileHandlerInterface
	metricsManager MetricsManagerInterface
	alertManager   AlertManagerInterface
	sampleInterval time.Duration
	lastStats      *Stats
	enabled        bool
	mu             sync.RWMutex
}

// NewStaticFileMonitor creates a new static file handler monitor
func NewStaticFileMonitor(fileHandler FileHandlerInterface, metricsManager MetricsManagerInterface, alertManager AlertManagerInterface) *StaticFileMonitor {
	return &StaticFileMonitor{
		fileHandler:    fileHandler,
		metricsManager: metricsManager,
		alertManager:   alertManager,
		sampleInterval: time.Minute,
		enabled:        true,
	}
}

// SetSampleInterval sets the sample interval for metrics collection
func (m *StaticFileMonitor) SetSampleInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sampleInterval = interval
}

// Enable enables the monitor
func (m *StaticFileMonitor) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables the monitor
func (m *StaticFileMonitor) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// Start starts the monitoring process
func (m *StaticFileMonitor) Start() {
	// Start monitoring loop
	go m.monitorLoop()
}

// monitorLoop periodically collects metrics
func (m *StaticFileMonitor) monitorLoop() {
	ticker := time.NewTicker(m.sampleInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		m.mu.RLock()
		enabled := m.enabled
		m.mu.RUnlock()
		
		if !enabled {
			continue
		}

		if err := m.CollectMetrics(); err != nil {
			fmt.Printf("Error collecting metrics: %v\n", err)
		}
		if err := m.CheckAlerts(); err != nil {
			fmt.Printf("Error checking alerts: %v\n", err)
		}
	}
}

// CollectMetrics collects metrics from the static file handler
func (m *StaticFileMonitor) CollectMetrics() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.fileHandler == nil {
		return nil
	}

	// Get current stats
	stats := m.fileHandler.GetStats()
	if stats == nil {
		return nil
	}

	// Store last stats for comparison
	m.lastStats = stats

	// Record metrics
	tags := map[string]string{"component": "static_file_handler"}
	
	if err := m.metricsManager.RecordGauge("static_file.files_served", stats.FilesServed, tags); err != nil {
		return fmt.Errorf("failed to record files_served metric: %w", err)
	}
	
	if err := m.metricsManager.RecordGauge("static_file.cache_hits", stats.CacheHits, tags); err != nil {
		return fmt.Errorf("failed to record cache_hits metric: %w", err)
	}
	
	if err := m.metricsManager.RecordGauge("static_file.cache_misses", stats.CacheMisses, tags); err != nil {
		return fmt.Errorf("failed to record cache_misses metric: %w", err)
	}

	return nil
}

// CheckAlerts checks for alert conditions
func (m *StaticFileMonitor) CheckAlerts() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.alertManager == nil || m.lastStats == nil {
		return nil
	}

	tags := map[string]string{"component": "static_file_handler"}
	
	// Check cache hit ratio
	if m.lastStats.CacheHits+m.lastStats.CacheMisses > 0 {
		hitRatio := float64(m.lastStats.CacheHits) / float64(m.lastStats.CacheHits+m.lastStats.CacheMisses)
		if err := m.alertManager.CheckThreshold("static_file.cache_hit_ratio", hitRatio, tags); err != nil {
			return fmt.Errorf("failed to check cache hit ratio alert: %w", err)
		}
	}

	return nil
}

// GetMetrics returns the current metrics
func (m *StaticFileMonitor) GetMetrics() *StaticFileMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.lastStats == nil {
		return &StaticFileMetrics{}
	}

	hitRatio := 0.0
	if m.lastStats.CacheHits+m.lastStats.CacheMisses > 0 {
		hitRatio = float64(m.lastStats.CacheHits) / float64(m.lastStats.CacheHits+m.lastStats.CacheMisses)
	}

	return &StaticFileMetrics{
		FilesServed:      m.lastStats.FilesServed,
		CacheHits:        m.lastStats.CacheHits,
		CacheMisses:      m.lastStats.CacheMisses,
		CacheHitRatio:    hitRatio,
		CompressedFiles:  m.lastStats.CompressedFiles,
		TotalSize:        m.lastStats.TotalSize,
		CompressedSize:   m.lastStats.CompressedSize,
		CompressionRatio: m.lastStats.CompressionRatio,
		CacheSize:        m.fileHandler.GetCacheSize(),
		CacheItemCount:   m.fileHandler.GetCacheItemCount(),
		AverageServeTime: m.lastStats.AverageServeTime,
	}
}