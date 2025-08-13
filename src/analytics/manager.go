package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Manager orchestrates all analytics operations
type Manager struct {
	config          *Config
	collector       *MetricsCollector
	storage         DataStorage
	trendAnalyzer   *TrendAnalyzer
	reportGenerator *ReportGenerator
	dashboardEngine *DashboardEngine
	logger          Logger
	mu              sync.RWMutex
}

// Config holds analytics configuration
type Config struct {
	// Storage settings
	StorageType     StorageType
	DatabaseURL     string
	RetentionPolicy RetentionPolicy
	
	// Collection settings
	MetricsEnabled     bool
	CollectionInterval time.Duration
	BatchSize          int
	BufferSize         int
	
	// Analysis settings
	TrendWindowDays    int
	AnalysisInterval   time.Duration
	AnomalyThreshold   float64
	BaselineWindow     time.Duration
	
	// Dashboard settings
	DashboardEnabled   bool
	RealTimeUpdates    bool
	RefreshInterval    time.Duration
	CacheTTL           time.Duration
	
	// Export settings
	ExportFormats      []string
	APIEnabled         bool
	WebhookURLs        []string
	
	// Security settings
	DataEncryption     bool
	AccessControl      bool
	AuditLogging       bool
}

// StorageType represents different storage backends
type StorageType string

const (
	StorageTypeMemory     StorageType = "memory"
	StorageTypeSQLite     StorageType = "sqlite"
	StorageTypePostgreSQL StorageType = "postgresql"
	StorageTypeMySQL      StorageType = "mysql"
	StorageTypeInfluxDB   StorageType = "influxdb"
)

// RetentionPolicy defines data retention rules
type RetentionPolicy struct {
	RawDataDays       int
	AggregatedDays    int
	TrendDataDays     int
	CompressAfterDays int
	ArchiveAfterDays  int
}

// Logger interface for analytics logging
type Logger interface {
	Info(msg string)
	Error(msg string, err error)
	Debug(msg string)
	Warn(msg string)
}

// NewManager creates a new analytics manager
func NewManager(config *Config, logger Logger) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Initialize storage
	storage, err := NewDataStorage(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	
	manager := &Manager{
		config:          config,
		collector:       NewMetricsCollector(config, logger),
		storage:         storage,
		trendAnalyzer:   NewTrendAnalyzer(config, logger),
		reportGenerator: NewReportGenerator(config, logger),
		dashboardEngine: NewDashboardEngine(config, logger),
		logger:          logger,
	}
	
	return manager, nil
}

// DefaultConfig returns default analytics configuration
func DefaultConfig() *Config {
	return &Config{
		StorageType:        StorageTypeSQLite,
		DatabaseURL:        "./analytics.db",
		RetentionPolicy: RetentionPolicy{
			RawDataDays:       30,
			AggregatedDays:    365,
			TrendDataDays:     90,
			CompressAfterDays: 7,
			ArchiveAfterDays:  180,
		},
		MetricsEnabled:     true,
		CollectionInterval: 1 * time.Minute,
		BatchSize:          100,
		BufferSize:         1000,
		TrendWindowDays:    30,
		AnalysisInterval:   1 * time.Hour,
		AnomalyThreshold:   2.0,
		BaselineWindow:     7 * 24 * time.Hour,
		DashboardEnabled:   true,
		RealTimeUpdates:    true,
		RefreshInterval:    30 * time.Second,
		CacheTTL:           5 * time.Minute,
		ExportFormats:      []string{"json", "csv", "excel"},
		APIEnabled:         true,
		DataEncryption:     false,
		AccessControl:      false,
		AuditLogging:       true,
	}
}

// Start initializes and starts analytics collection
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("Starting analytics manager...")
	
	// Initialize storage
	if err := m.storage.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	
	// Start metrics collection
	if m.config.MetricsEnabled {
		if err := m.collector.Start(ctx, m.storage); err != nil {
			return fmt.Errorf("failed to start metrics collection: %w", err)
		}
	}
	
	// Start trend analysis
	if err := m.trendAnalyzer.Start(ctx, m.storage); err != nil {
		return fmt.Errorf("failed to start trend analysis: %w", err)
	}
	
	// Start dashboard engine
	if m.config.DashboardEnabled {
		if err := m.dashboardEngine.Start(ctx, m.storage); err != nil {
			return fmt.Errorf("failed to start dashboard engine: %w", err)
		}
	}
	
	// Start cleanup routine
	go m.runCleanup(ctx)
	
	m.logger.Info("Analytics manager started successfully")
	return nil
}

// Stop gracefully shuts down analytics components
func (m *Manager) Stop() error {
	m.logger.Info("Stopping analytics manager...")
	
	// Stop components
	if err := m.collector.Stop(); err != nil {
		m.logger.Error("Failed to stop metrics collector", err)
	}
	
	if err := m.trendAnalyzer.Stop(); err != nil {
		m.logger.Error("Failed to stop trend analyzer", err)
	}
	
	if err := m.dashboardEngine.Stop(); err != nil {
		m.logger.Error("Failed to stop dashboard engine", err)
	}
	
	// Close storage
	if err := m.storage.Close(); err != nil {
		m.logger.Error("Failed to close storage", err)
		return err
	}
	
	m.logger.Info("Analytics manager stopped")
	return nil
}

// RecordScanResult records a scan result for analytics
func (m *Manager) RecordScanResult(result *ScanResult) error {
	if !m.config.MetricsEnabled {
		return nil
	}
	
	return m.collector.RecordScanResult(result)
}

// RecordMetric records a custom metric
func (m *Manager) RecordMetric(metric *Metric) error {
	if !m.config.MetricsEnabled {
		return nil
	}
	
	return m.collector.RecordMetric(metric)
}

// GetDashboard returns dashboard data
func (m *Manager) GetDashboard(timeRange TimeRange) (*Dashboard, error) {
	return m.dashboardEngine.GenerateDashboard(timeRange)
}

// GetTrends returns trend analysis data
func (m *Manager) GetTrends(params *TrendParams) (*TrendAnalysis, error) {
	return m.trendAnalyzer.AnalyzeTrends(params)
}

// GetReport generates an analytics report
func (m *Manager) GetReport(params *ReportParams) (*Report, error) {
	return m.reportGenerator.GenerateReport(params)
}

// GetMetrics retrieves metrics data
func (m *Manager) GetMetrics(query *MetricsQuery) (*MetricsResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.storage.QueryMetrics(query)
}

// GetComparisonAnalysis performs comparative analysis
func (m *Manager) GetComparisonAnalysis(params *ComparisonParams) (*ComparisonResult, error) {
	return m.performComparison(params)
}

// ExportData exports analytics data in specified format
func (m *Manager) ExportData(params *ExportParams) ([]byte, error) {
	switch params.Format {
	case "json":
		return m.exportJSON(params)
	case "csv":
		return m.exportCSV(params)
	case "excel":
		return m.exportExcel(params)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", params.Format)
	}
}

// GetAnalyticsSummary returns a high-level analytics summary
func (m *Manager) GetAnalyticsSummary() (*AnalyticsSummary, error) {
	summary := &AnalyticsSummary{
		GeneratedAt: time.Now(),
	}
	
	// Get basic metrics
	if metrics, err := m.storage.QueryMetrics(&MetricsQuery{
		TimeRange: TimeRange{
			Start: time.Now().AddDate(0, 0, -30),
			End:   time.Now(),
		},
	}); err == nil {
		summary.TotalScans = len(metrics.Data)
		summary.calculateBasicStats(metrics)
	}
	
	// Get trends
	if trends, err := m.trendAnalyzer.AnalyzeTrends(&TrendParams{
		TimeRange: TimeRange{
			Start: time.Now().AddDate(0, 0, -7),
			End:   time.Now(),
		},
		Metrics: []string{"vulnerability_count", "scan_duration"},
	}); err == nil {
		summary.TrendData = trends.Summary
	}
	
	// Get storage stats
	summary.StorageStats = m.getStorageStats()
	
	return summary, nil
}

// runCleanup performs periodic cleanup of old data
func (m *Manager) runCleanup(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Daily cleanup
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.performCleanup(); err != nil {
				m.logger.Error("Cleanup failed", err)
			}
		}
	}
}

// performCleanup removes old data according to retention policy
func (m *Manager) performCleanup() error {
	m.logger.Info("Starting data cleanup...")
	
	policy := m.config.RetentionPolicy
	now := time.Now()
	
	// Clean raw data
	if policy.RawDataDays > 0 {
		cutoff := now.AddDate(0, 0, -policy.RawDataDays)
		if err := m.storage.DeleteRawData(cutoff); err != nil {
			return fmt.Errorf("failed to clean raw data: %w", err)
		}
	}
	
	// Archive old data
	if policy.ArchiveAfterDays > 0 {
		cutoff := now.AddDate(0, 0, -policy.ArchiveAfterDays)
		if err := m.storage.ArchiveData(cutoff); err != nil {
			return fmt.Errorf("failed to archive data: %w", err)
		}
	}
	
	m.logger.Info("Data cleanup completed")
	return nil
}

// performComparison performs comparative analysis
func (m *Manager) performComparison(params *ComparisonParams) (*ComparisonResult, error) {
	result := &ComparisonResult{
		ComparisonType: params.Type,
		GeneratedAt:    time.Now(),
		Comparisons:    make([]Comparison, 0),
	}
	
	switch params.Type {
	case ComparisonTypeTimeRange:
		return m.compareTimeRanges(params)
	case ComparisonTypeTargets:
		return m.compareTargets(params)
	case ComparisonTypeTemplates:
		return m.compareTemplates(params)
	default:
		return nil, fmt.Errorf("unsupported comparison type: %s", params.Type)
	}
}

// compareTimeRanges compares metrics across different time ranges
func (m *Manager) compareTimeRanges(params *ComparisonParams) (*ComparisonResult, error) {
	result := &ComparisonResult{
		ComparisonType: ComparisonTypeTimeRange,
		GeneratedAt:    time.Now(),
	}
	
	// Get metrics for each time range
	for _, timeRange := range params.TimeRanges {
		metrics, err := m.storage.QueryMetrics(&MetricsQuery{
			TimeRange: timeRange,
			Metrics:   params.Metrics,
		})
		if err != nil {
			return nil, err
		}
		
		comparison := Comparison{
			Label:   fmt.Sprintf("%s to %s", timeRange.Start.Format("2006-01-02"), timeRange.End.Format("2006-01-02")),
			Metrics: m.aggregateMetrics(metrics),
		}
		
		result.Comparisons = append(result.Comparisons, comparison)
	}
	
	// Calculate deltas
	result.calculateDeltas()
	
	return result, nil
}

// Helper methods

func (m *Manager) getStorageStats() StorageStats {
	stats := StorageStats{}
	
	// Get storage size and record counts
	if size, err := m.storage.GetStorageSize(); err == nil {
		stats.TotalSize = size
	}
	
	if count, err := m.storage.GetRecordCount(); err == nil {
		stats.TotalRecords = count
	}
	
	return stats
}

func (m *Manager) aggregateMetrics(metrics *MetricsResult) map[string]float64 {
	aggregated := make(map[string]float64)
	
	for _, dataPoint := range metrics.Data {
		for key, value := range dataPoint.Values {
			aggregated[key] += value
		}
	}
	
	return aggregated
}

func (m *Manager) exportJSON(params *ExportParams) ([]byte, error) {
	data, err := m.gatherExportData(params)
	if err != nil {
		return nil, err
	}
	
	return json.MarshalIndent(data, "", "  ")
}

func (m *Manager) exportCSV(params *ExportParams) ([]byte, error) {
	// Implementation for CSV export
	return nil, fmt.Errorf("CSV export not yet implemented")
}

func (m *Manager) exportExcel(params *ExportParams) ([]byte, error) {
	// Implementation for Excel export
	return nil, fmt.Errorf("Excel export not yet implemented")
}

func (m *Manager) gatherExportData(params *ExportParams) (interface{}, error) {
	switch params.DataType {
	case "metrics":
		return m.storage.QueryMetrics(&MetricsQuery{
			TimeRange: params.TimeRange,
			Metrics:   params.Metrics,
		})
	case "trends":
		return m.trendAnalyzer.AnalyzeTrends(&TrendParams{
			TimeRange: params.TimeRange,
			Metrics:   params.Metrics,
		})
	default:
		return nil, fmt.Errorf("unsupported data type: %s", params.DataType)
	}
}

// compareTargets compares metrics across different targets
func (m *Manager) compareTargets(params *ComparisonParams) (*ComparisonResult, error) {
	// Implementation for target comparison
	return nil, fmt.Errorf("target comparison not yet implemented")
}

// compareTemplates compares metrics across different templates
func (m *Manager) compareTemplates(params *ComparisonParams) (*ComparisonResult, error) {
	// Implementation for template comparison
	return nil, fmt.Errorf("template comparison not yet implemented")
}

// calculateBasicStats calculates basic statistics from metrics
func (summary *AnalyticsSummary) calculateBasicStats(metrics *MetricsResult) {
	if len(metrics.Data) == 0 {
		return
	}
	
	// Calculate averages, totals, etc.
	var totalVulns, totalDuration float64
	durations := make([]float64, 0)
	
	for _, dataPoint := range metrics.Data {
		if v, ok := dataPoint.Values["vulnerability_count"]; ok {
			totalVulns += v
		}
		if d, ok := dataPoint.Values["scan_duration"]; ok {
			totalDuration += d
			durations = append(durations, d)
		}
	}
	
	summary.TotalVulnerabilities = int(totalVulns)
	summary.AverageScanDuration = totalDuration / float64(len(metrics.Data))
	
	// Calculate median duration
	if len(durations) > 0 {
		sort.Float64s(durations)
		if len(durations)%2 == 0 {
			summary.MedianScanDuration = (durations[len(durations)/2-1] + durations[len(durations)/2]) / 2
		} else {
			summary.MedianScanDuration = durations[len(durations)/2]
		}
	}
}

// calculateDeltas calculates percentage changes between comparisons
func (result *ComparisonResult) calculateDeltas() {
	if len(result.Comparisons) < 2 {
		return
	}
	
	baseline := result.Comparisons[0]
	
	for i := 1; i < len(result.Comparisons); i++ {
		comparison := &result.Comparisons[i]
		comparison.Deltas = make(map[string]float64)
		
		for metric, value := range comparison.Metrics {
			if baseValue, ok := baseline.Metrics[metric]; ok && baseValue != 0 {
				delta := ((value - baseValue) / baseValue) * 100
				comparison.Deltas[metric] = delta
			}
		}
	}
}