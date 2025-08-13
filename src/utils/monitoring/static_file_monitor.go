package monitoring

import (
	"sync"
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

		_ = m.CollectMetrics()
		_ = m.CheckAlerts()
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
	
	cacheSize := m.fileHandler.GetCacheSize()
	cacheItemCount := m.fileHandler.GetCacheItemCount()

	// Calculate cache hit ratio
	cacheHitRatio := float64(0)
	if stats.CacheHits+stats.CacheMisses > 0 {
		cacheHitRatio = float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses)
	}

	// Create metrics
	metrics := &StaticFileMetrics{
		FilesServed:      stats.FilesServed,
		CacheHits:        stats.CacheHits,
		CacheMisses:      stats.CacheMisses,
		CacheHitRatio:    cacheHitRatio,
		CompressedFiles:  stats.CompressedFiles,
		TotalSize:        stats.TotalSize,
		CompressedSize:   stats.CompressedSize,
		CompressionRatio: stats.CompressionRatio,
		CacheSize:        cacheSize,
		CacheItemCount:   cacheItemCount,
		AverageServeTime: stats.AverageServeTime,
	}

	// Record metrics
	if m.metricsManager != nil {
		m.recordMetrics(metrics)
	}

	// Update last stats
	m.lastStats = stats
	return nil
}

// recordMetrics records metrics to the metrics manager
func (m *StaticFileMonitor) recordMetrics(metrics *StaticFileMetrics) error {
	// Record counter metrics
	_ = m.metricsManager.RecordCounter("static_file.files_served", metrics.FilesServed, nil)
	_ = m.metricsManager.RecordCounter("static_file.cache_hits", metrics.CacheHits, nil)
	_ = m.metricsManager.RecordCounter("static_file.cache_misses", metrics.CacheMisses, nil)
	_ = m.metricsManager.RecordCounter("static_file.compressed_files", metrics.CompressedFiles, nil)
	_ = m.metricsManager.RecordCounter("static_file.total_size", metrics.TotalSize, nil)
	_ = m.metricsManager.RecordCounter("static_file.compressed_size", metrics.CompressedSize, nil)

	// Record gauge metrics
	_ = m.metricsManager.RecordGauge("static_file.cache_hit_ratio", metrics.CacheHitRatio, nil)
	_ = m.metricsManager.RecordGauge("static_file.compression_ratio", metrics.CompressionRatio, nil)
	_ = m.metricsManager.RecordGauge("static_file.cache_size", metrics.CacheSize, nil)
	_ = m.metricsManager.RecordGauge("static_file.cache_item_count", metrics.CacheItemCount, nil)
	_ = m.metricsManager.RecordGauge("static_file.average_serve_time_ms", metrics.AverageServeTime.Milliseconds(), nil)
	
	return nil
}

// CheckAlerts checks for alert conditions
func (m *StaticFileMonitor) CheckAlerts() error {
	if m.alertManager == nil {
		return nil
	}

	metrics := m.GetMetrics()
	if metrics == nil {
		return nil
	}

	// Check cache size
	_ = m.alertManager.CheckThreshold("static_file.cache_size", metrics.CacheSize, nil)

	// Check cache hit ratio
	_ = m.alertManager.CheckThreshold("static_file.cache_hit_ratio", metrics.CacheHitRatio, nil)

	// Check average serve time
	_ = m.alertManager.CheckThreshold("static_file.average_serve_time_ms", metrics.AverageServeTime.Milliseconds(), nil)

	return nil
}

// GetMetrics returns the current static file metrics
func (m *StaticFileMonitor) GetMetrics() *StaticFileMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.fileHandler == nil {
		return nil
	}

	stats := m.fileHandler.GetStats()
	if stats == nil {
		return nil
	}
	
	cacheSize := m.fileHandler.GetCacheSize()
	cacheItemCount := m.fileHandler.GetCacheItemCount()

	cacheHitRatio := float64(0)
	if stats.CacheHits+stats.CacheMisses > 0 {
		cacheHitRatio = float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses)
	}

	return &StaticFileMetrics{
		FilesServed:      stats.FilesServed,
		CacheHits:        stats.CacheHits,
		CacheMisses:      stats.CacheMisses,
		CacheHitRatio:    cacheHitRatio,
		CompressedFiles:  stats.CompressedFiles,
		TotalSize:        stats.TotalSize,
		CompressedSize:   stats.CompressedSize,
		CompressionRatio: stats.CompressionRatio,
		CacheSize:        cacheSize,
		CacheItemCount:   cacheItemCount,
		AverageServeTime: stats.AverageServeTime,
	}
}
