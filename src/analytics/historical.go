package analytics

import (
	"context"
	"fmt"
	"time"
)

// HistoricalDataManager manages long-term data storage and archival
type HistoricalDataManager struct {
	config      *Config
	storage     DataStorage
	archiver    DataArchiver
	compressor  DataCompressor
	logger      Logger
	
	// Retention policies
	policies    map[string]RetentionPolicy
	
	// Archival state
	lastArchival time.Time
	archivalLock bool
}

// RetentionPolicy defines how long to keep different types of data
type RetentionPolicy struct {
	DataType        string        `json:"data_type"`
	HotRetention    time.Duration `json:"hot_retention"`    // Keep in primary storage
	WarmRetention   time.Duration `json:"warm_retention"`   // Keep in secondary storage
	ColdRetention   time.Duration `json:"cold_retention"`   // Keep in archive storage
	CompressionAge  time.Duration `json:"compression_age"`  // When to compress
	DeletionAge     time.Duration `json:"deletion_age"`     // When to delete permanently
}

// DataArchiver interface for archival operations
type DataArchiver interface {
	Archive(ctx context.Context, data []Metric, archivePath string) error
	Retrieve(ctx context.Context, archivePath string, timeRange TimeWindow) ([]Metric, error)
	List(ctx context.Context, pattern string) ([]ArchiveInfo, error)
	Delete(ctx context.Context, archivePath string) error
}

// DataCompressor interface for data compression
type DataCompressor interface {
	Compress(ctx context.Context, data []byte) ([]byte, error)
	Decompress(ctx context.Context, compressedData []byte) ([]byte, error)
	GetCompressionRatio() float64
}

// ArchiveInfo contains information about archived data
type ArchiveInfo struct {
	Path         string    `json:"path"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	MetricCount  int       `json:"metric_count"`
	CompressedSize int64   `json:"compressed_size"`
	OriginalSize   int64   `json:"original_size"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewHistoricalDataManager creates a new historical data manager
func NewHistoricalDataManager(config *Config, storage DataStorage, logger Logger) *HistoricalDataManager {
	manager := &HistoricalDataManager{
		config:   config,
		storage:  storage,
		archiver: &FileSystemArchiver{basePath: config.Analytics.ArchivePath, logger: logger},
		compressor: &GzipCompressor{},
		logger:   logger,
		policies: make(map[string]RetentionPolicy),
	}
	
	// Set up default retention policies
	manager.setupDefaultPolicies()
	
	return manager
}

// Archive moves old data to archive storage
func (hdm *HistoricalDataManager) Archive(ctx context.Context) error {
	if hdm.archivalLock {
		return fmt.Errorf("archival already in progress")
	}
	
	hdm.archivalLock = true
	defer func() { hdm.archivalLock = false }()
	
	hdm.logger.Info("Starting data archival process")
	
	for dataType, policy := range hdm.policies {
		if err := hdm.archiveDataType(ctx, dataType, policy); err != nil {
			hdm.logger.Error("Failed to archive data type", "dataType", dataType, "error", err)
			continue
		}
	}
	
	hdm.lastArchival = time.Now()
	hdm.logger.Info("Data archival process completed")
	
	return nil
}

// GetHistoricalData retrieves historical data across storage tiers
func (hdm *HistoricalDataManager) GetHistoricalData(ctx context.Context, metricName string, timeRange TimeWindow) ([]Metric, error) {
	var allMetrics []Metric
	
	// Get data from primary storage (hot data)
	hotData, err := hdm.storage.GetMetricsByNameAndTimeRange(ctx, metricName, timeRange.Start, timeRange.End)
	if err != nil {
		hdm.logger.Warn("Failed to get hot data", "error", err)
	} else {
		allMetrics = append(allMetrics, hotData...)
	}
	
	// Get data from archive storage (cold data)
	coldData, err := hdm.getArchivedData(ctx, metricName, timeRange)
	if err != nil {
		hdm.logger.Warn("Failed to get archived data", "error", err)
	} else {
		allMetrics = append(allMetrics, coldData...)
	}
	
	// Sort by timestamp
	hdm.sortMetricsByTime(allMetrics)
	
	return allMetrics, nil
}

// GetRetentionStatus returns current retention status
func (hdm *HistoricalDataManager) GetRetentionStatus(ctx context.Context) (map[string]RetentionStatus, error) {
	status := make(map[string]RetentionStatus)
	
	for dataType, policy := range hdm.policies {
		retentionStatus, err := hdm.calculateRetentionStatus(ctx, dataType, policy)
		if err != nil {
			hdm.logger.Error("Failed to calculate retention status", "dataType", dataType, "error", err)
			continue
		}
		status[dataType] = retentionStatus
	}
	
	return status, nil
}

// SetRetentionPolicy sets a custom retention policy
func (hdm *HistoricalDataManager) SetRetentionPolicy(dataType string, policy RetentionPolicy) {
	hdm.policies[dataType] = policy
	hdm.logger.Info("Set retention policy", "dataType", dataType, "policy", policy)
}

// Cleanup removes expired data according to retention policies
func (hdm *HistoricalDataManager) Cleanup(ctx context.Context) error {
	hdm.logger.Info("Starting data cleanup process")
	
	for dataType, policy := range hdm.policies {
		if err := hdm.cleanupDataType(ctx, dataType, policy); err != nil {
			hdm.logger.Error("Failed to cleanup data type", "dataType", dataType, "error", err)
			continue
		}
	}
	
	hdm.logger.Info("Data cleanup process completed")
	return nil
}

// Internal methods

func (hdm *HistoricalDataManager) setupDefaultPolicies() {
	// Default policy for metrics
	hdm.policies["metrics"] = RetentionPolicy{
		DataType:       "metrics",
		HotRetention:   7 * 24 * time.Hour,     // 7 days
		WarmRetention:  30 * 24 * time.Hour,    // 30 days
		ColdRetention:  365 * 24 * time.Hour,   // 1 year
		CompressionAge: 24 * time.Hour,         // 1 day
		DeletionAge:    2 * 365 * 24 * time.Hour, // 2 years
	}
	
	// Policy for aggregated data
	hdm.policies["aggregated"] = RetentionPolicy{
		DataType:       "aggregated",
		HotRetention:   30 * 24 * time.Hour,    // 30 days
		WarmRetention:  90 * 24 * time.Hour,    // 90 days
		ColdRetention:  2 * 365 * 24 * time.Hour, // 2 years
		CompressionAge: 7 * 24 * time.Hour,     // 7 days
		DeletionAge:    5 * 365 * 24 * time.Hour, // 5 years
	}
	
	// Policy for scan results
	hdm.policies["scan_results"] = RetentionPolicy{
		DataType:       "scan_results",
		HotRetention:   14 * 24 * time.Hour,    // 14 days
		WarmRetention:  90 * 24 * time.Hour,    // 90 days
		ColdRetention:  365 * 24 * time.Hour,   // 1 year
		CompressionAge: 3 * 24 * time.Hour,     // 3 days
		DeletionAge:    3 * 365 * 24 * time.Hour, // 3 years
	}
}

func (hdm *HistoricalDataManager) archiveDataType(ctx context.Context, dataType string, policy RetentionPolicy) error {
	cutoffTime := time.Now().Add(-policy.HotRetention)
	
	// Get old metrics that need archiving
	oldMetrics, err := hdm.storage.GetMetricsByTimeRange(ctx, time.Time{}, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get old metrics: %w", err)
	}
	
	if len(oldMetrics) == 0 {
		return nil // Nothing to archive
	}
	
	// Create archive path
	archivePath := fmt.Sprintf("%s/%s_%s.archive", 
		dataType, 
		cutoffTime.Format("2006-01-02"), 
		time.Now().Format("150405"))
	
	// Archive the data
	if err := hdm.archiver.Archive(ctx, oldMetrics, archivePath); err != nil {
		return fmt.Errorf("failed to archive data: %w", err)
	}
	
	// Remove from primary storage
	if err := hdm.storage.DeleteMetricsByTimeRange(ctx, time.Time{}, cutoffTime); err != nil {
		hdm.logger.Warn("Failed to delete archived metrics from primary storage", "error", err)
	}
	
	hdm.logger.Info("Archived data", "dataType", dataType, "count", len(oldMetrics), "path", archivePath)
	
	return nil
}

func (hdm *HistoricalDataManager) getArchivedData(ctx context.Context, metricName string, timeRange TimeWindow) ([]Metric, error) {
	// List relevant archives
	pattern := fmt.Sprintf("*%s*", metricName)
	archives, err := hdm.archiver.List(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list archives: %w", err)
	}
	
	var allMetrics []Metric
	
	// Retrieve data from relevant archives
	for _, archive := range archives {
		// Check if archive overlaps with requested time range
		if hdm.overlapsTimeRange(archive, timeRange) {
			metrics, err := hdm.archiver.Retrieve(ctx, archive.Path, timeRange)
			if err != nil {
				hdm.logger.Warn("Failed to retrieve from archive", "path", archive.Path, "error", err)
				continue
			}
			allMetrics = append(allMetrics, metrics...)
		}
	}
	
	return allMetrics, nil
}

func (hdm *HistoricalDataManager) calculateRetentionStatus(ctx context.Context, dataType string, policy RetentionPolicy) (RetentionStatus, error) {
	now := time.Now()
	
	// Count metrics in different age ranges
	hotCutoff := now.Add(-policy.HotRetention)
	warmCutoff := now.Add(-policy.WarmRetention)
	coldCutoff := now.Add(-policy.ColdRetention)
	
	hotCount, _ := hdm.storage.CountMetricsByTimeRange(ctx, hotCutoff, now)
	warmCount, _ := hdm.storage.CountMetricsByTimeRange(ctx, warmCutoff, hotCutoff)
	coldCount, _ := hdm.storage.CountMetricsByTimeRange(ctx, coldCutoff, warmCutoff)
	
	return RetentionStatus{
		DataType:    dataType,
		HotCount:    hotCount,
		WarmCount:   warmCount,
		ColdCount:   coldCount,
		LastArchival: hdm.lastArchival,
		NextArchival: hdm.lastArchival.Add(24 * time.Hour), // Daily archival
	}, nil
}

func (hdm *HistoricalDataManager) cleanupDataType(ctx context.Context, dataType string, policy RetentionPolicy) error {
	deletionCutoff := time.Now().Add(-policy.DeletionAge)
	
	// List archives older than deletion age
	archives, err := hdm.archiver.List(ctx, fmt.Sprintf("%s_*", dataType))
	if err != nil {
		return fmt.Errorf("failed to list archives: %w", err)
	}
	
	var deletedCount int
	for _, archive := range archives {
		if archive.EndTime.Before(deletionCutoff) {
			if err := hdm.archiver.Delete(ctx, archive.Path); err != nil {
				hdm.logger.Warn("Failed to delete archive", "path", archive.Path, "error", err)
				continue
			}
			deletedCount++
		}
	}
	
	if deletedCount > 0 {
		hdm.logger.Info("Deleted expired archives", "dataType", dataType, "count", deletedCount)
	}
	
	return nil
}

func (hdm *HistoricalDataManager) overlapsTimeRange(archive ArchiveInfo, timeRange TimeWindow) bool {
	return archive.StartTime.Before(timeRange.End) && archive.EndTime.After(timeRange.Start)
}

func (hdm *HistoricalDataManager) sortMetricsByTime(metrics []Metric) {
	// Simple bubble sort for demonstration
	for i := 0; i < len(metrics); i++ {
		for j := 0; j < len(metrics)-1-i; j++ {
			if metrics[j].Timestamp.After(metrics[j+1].Timestamp) {
				metrics[j], metrics[j+1] = metrics[j+1], metrics[j]
			}
		}
	}
}

// RetentionStatus represents current retention status
type RetentionStatus struct {
	DataType     string    `json:"data_type"`
	HotCount     int       `json:"hot_count"`
	WarmCount    int       `json:"warm_count"`
	ColdCount    int       `json:"cold_count"`
	LastArchival time.Time `json:"last_archival"`
	NextArchival time.Time `json:"next_archival"`
}

// FileSystemArchiver implements DataArchiver for filesystem storage
type FileSystemArchiver struct {
	basePath string
	logger   Logger
}

func (fsa *FileSystemArchiver) Archive(ctx context.Context, data []Metric, archivePath string) error {
	// Mock implementation - would actually write to filesystem
	fsa.logger.Info("Archiving data to filesystem", "path", archivePath, "count", len(data))
	return nil
}

func (fsa *FileSystemArchiver) Retrieve(ctx context.Context, archivePath string, timeRange TimeWindow) ([]Metric, error) {
	// Mock implementation - would actually read from filesystem
	fsa.logger.Info("Retrieving data from archive", "path", archivePath)
	return []Metric{}, nil
}

func (fsa *FileSystemArchiver) List(ctx context.Context, pattern string) ([]ArchiveInfo, error) {
	// Mock implementation - would actually list files
	return []ArchiveInfo{}, nil
}

func (fsa *FileSystemArchiver) Delete(ctx context.Context, archivePath string) error {
	// Mock implementation - would actually delete file
	fsa.logger.Info("Deleting archive", "path", archivePath)
	return nil
}

// GzipCompressor implements DataCompressor using gzip
type GzipCompressor struct{}

func (gc *GzipCompressor) Compress(ctx context.Context, data []byte) ([]byte, error) {
	// Mock implementation - would actually compress data
	return data, nil
}

func (gc *GzipCompressor) Decompress(ctx context.Context, compressedData []byte) ([]byte, error) {
	// Mock implementation - would actually decompress data
	return compressedData, nil
}

func (gc *GzipCompressor) GetCompressionRatio() float64 {
	return 0.3 // Mock 70% compression ratio
}