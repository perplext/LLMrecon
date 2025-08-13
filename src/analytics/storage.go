package analytics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements DataStorage interface using SQLite
type SQLiteStorage struct {
	config *Config
	db     *sql.DB
	logger Logger
	mu     sync.RWMutex
}

// MemoryStorage implements DataStorage interface using in-memory storage
type MemoryStorage struct {
	config      *Config
	metrics     []Metric
	scanResults []ScanResult
	logger      Logger
	mu          sync.RWMutex
}

// NewDataStorage creates a new data storage instance based on configuration
func NewDataStorage(config *Config, logger Logger) (DataStorage, error) {
	switch config.StorageType {
	case StorageTypeMemory:
		return NewMemoryStorage(config, logger), nil
	case StorageTypeSQLite:
		return NewSQLiteStorage(config, logger)
	case StorageTypePostgreSQL:
		return NewPostgreSQLStorage(config, logger)
	case StorageTypeMySQL:
		return NewMySQLStorage(config, logger)
	case StorageTypeInfluxDB:
		return NewInfluxDBStorage(config, logger)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.StorageType)
	}
}

// SQLite Storage Implementation

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(config *Config, logger Logger) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	
	storage := &SQLiteStorage{
		config: config,
		db:     db,
		logger: logger,
	}
	
	return storage, nil
}

// Initialize initializes the SQLite storage
func (s *SQLiteStorage) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Create tables
	if err := s.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	
	// Create indexes
	if err := s.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	s.logger.Info("SQLite storage initialized")
	return nil
}

// Close closes the SQLite database connection
func (s *SQLiteStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}

// createTables creates the necessary database tables
func (s *SQLiteStorage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value REAL NOT NULL,
			unit TEXT,
			tags TEXT, -- JSON
			timestamp DATETIME NOT NULL,
			source TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS scan_results (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			duration INTEGER NOT NULL, -- nanoseconds
			target TEXT NOT NULL,
			templates_used TEXT, -- JSON array
			total_tests INTEGER NOT NULL,
			passed_tests INTEGER NOT NULL,
			failed_tests INTEGER NOT NULL,
			vulnerabilities TEXT, -- JSON array
			metadata TEXT, -- JSON
			success BOOLEAN NOT NULL,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS vulnerabilities (
			id TEXT PRIMARY KEY,
			scan_id TEXT NOT NULL,
			type TEXT NOT NULL,
			severity TEXT NOT NULL,
			category TEXT NOT NULL,
			template TEXT NOT NULL,
			description TEXT NOT NULL,
			confidence REAL NOT NULL,
			cvss REAL,
			cwe TEXT,
			owasp TEXT,
			evidence TEXT, -- JSON
			remediation TEXT,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY (scan_id) REFERENCES scan_results (id)
		)`,
		`CREATE TABLE IF NOT EXISTS aggregated_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			metric_name TEXT NOT NULL,
			aggregation_type TEXT NOT NULL, -- sum, avg, count, max, min
			value REAL NOT NULL,
			interval_start DATETIME NOT NULL,
			interval_end DATETIME NOT NULL,
			tags TEXT, -- JSON
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	
	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}
	
	return nil
}

// createIndexes creates database indexes for performance
func (s *SQLiteStorage) createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name)",
		"CREATE INDEX IF NOT EXISTS idx_metrics_name_timestamp ON metrics(name, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_scan_results_timestamp ON scan_results(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_scan_results_target ON scan_results(target)",
		"CREATE INDEX IF NOT EXISTS idx_vulnerabilities_scan_id ON vulnerabilities(scan_id)",
		"CREATE INDEX IF NOT EXISTS idx_vulnerabilities_severity ON vulnerabilities(severity)",
		"CREATE INDEX IF NOT EXISTS idx_vulnerabilities_type ON vulnerabilities(type)",
		"CREATE INDEX IF NOT EXISTS idx_aggregated_metrics_interval ON aggregated_metrics(interval_start, interval_end)",
	}
	
	for _, index := range indexes {
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	
	return nil
}

// StoreMetric stores a metric in the database
func (s *SQLiteStorage) StoreMetric(metric *Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	tagsJSON, err := json.Marshal(metric.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	
	query := `INSERT INTO metrics (name, value, unit, tags, timestamp, source) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err = s.db.Exec(query, metric.Name, metric.Value, metric.Unit, 
		string(tagsJSON), metric.Timestamp, metric.Source)
	if err != nil {
		return fmt.Errorf("failed to store metric: %w", err)
	}
	
	return nil
}

// StoreScanResult stores a scan result in the database
func (s *SQLiteStorage) StoreScanResult(result *ScanResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Marshal complex fields
	templatesJSON, _ := json.Marshal(result.TemplatesUsed)
	vulnsJSON, _ := json.Marshal(result.Vulnerabilities)
	metadataJSON, _ := json.Marshal(result.Metadata)
	
	// Store scan result
	query := `INSERT INTO scan_results 
			  (id, timestamp, duration, target, templates_used, total_tests, 
			   passed_tests, failed_tests, vulnerabilities, metadata, success, error_message) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err = tx.Exec(query, result.ID, result.Timestamp, result.Duration.Nanoseconds(),
		result.Target, string(templatesJSON), result.TotalTests, result.PassedTests,
		result.FailedTests, string(vulnsJSON), string(metadataJSON), 
		result.Success, result.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to store scan result: %w", err)
	}
	
	// Store individual vulnerabilities
	for _, vuln := range result.Vulnerabilities {
		evidenceJSON, _ := json.Marshal(vuln.Evidence)
		
		vulnQuery := `INSERT INTO vulnerabilities 
					  (id, scan_id, type, severity, category, template, description, 
					   confidence, cvss, cwe, owasp, evidence, remediation, timestamp) 
					  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
		_, err = tx.Exec(vulnQuery, vuln.ID, result.ID, vuln.Type, vuln.Severity,
			vuln.Category, vuln.Template, vuln.Description, vuln.Confidence,
			nullableFloat64(vuln.CVSS), nullableString(vuln.CWE), 
			nullableString(vuln.OWASP), string(evidenceJSON), 
			nullableString(vuln.Remediation), result.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to store vulnerability: %w", err)
		}
	}
	
	// Store derived metrics
	metrics := s.extractMetricsFromScanResult(result)
	for _, metric := range metrics {
		if err := s.storeMetricInTx(tx, &metric); err != nil {
			s.logger.Error("Failed to store derived metric", err)
		}
	}
	
	return tx.Commit()
}

// QueryMetrics queries metrics from the database
func (s *SQLiteStorage) QueryMetrics(query *MetricsQuery) (*MetricsResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	startTime := time.Now()
	
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}
	
	// Build SQL query
	sqlQuery, args := s.buildMetricsQuery(query)
	
	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	
	// Parse results
	dataPoints := make([]DataPoint, 0)
	for rows.Next() {
		var timestamp time.Time
		var name string
		var value float64
		var tagsJSON string
		
		if err := rows.Scan(&timestamp, &name, &value, &tagsJSON); err != nil {
			continue
		}
		
		var tags map[string]string
		json.Unmarshal([]byte(tagsJSON), &tags)
		
		// Find or create data point for this timestamp
		var dataPoint *DataPoint
		for i := range dataPoints {
			if dataPoints[i].Timestamp.Equal(timestamp) {
				dataPoint = &dataPoints[i]
				break
			}
		}
		
		if dataPoint == nil {
			dataPoints = append(dataPoints, DataPoint{
				Timestamp: timestamp,
				Values:    make(map[string]float64),
				Tags:      tags,
			})
			dataPoint = &dataPoints[len(dataPoints)-1]
		}
		
		dataPoint.Values[name] = value
	}
	
	result := &MetricsResult{
		Query:     query,
		Data:      dataPoints,
		Total:     len(dataPoints),
		Cached:    false,
		QueryTime: time.Since(startTime),
	}
	
	return result, nil
}

// buildMetricsQuery builds SQL query from MetricsQuery
func (s *SQLiteStorage) buildMetricsQuery(query *MetricsQuery) (string, []interface{}) {
	var sql strings.Builder
	var args []interface{}
	
	sql.WriteString("SELECT timestamp, name, value, tags FROM metrics WHERE ")
	sql.WriteString("timestamp >= ? AND timestamp <= ?")
	args = append(args, query.TimeRange.Start, query.TimeRange.End)
	
	// Add metric name filter
	if len(query.Metrics) > 0 {
		placeholders := make([]string, len(query.Metrics))
		for i, metric := range query.Metrics {
			placeholders[i] = "?"
			args = append(args, metric)
		}
		sql.WriteString(" AND name IN (")
		sql.WriteString(strings.Join(placeholders, ","))
		sql.WriteString(")")
	}
	
	// Add filters
	for key, value := range query.Filters {
		sql.WriteString(" AND JSON_EXTRACT(tags, '$.")
		sql.WriteString(key)
		sql.WriteString("') = ?")
		args = append(args, value)
	}
	
	sql.WriteString(" ORDER BY timestamp")
	
	// Add limit and offset
	if query.Limit > 0 {
		sql.WriteString(" LIMIT ?")
		args = append(args, query.Limit)
		
		if query.Offset > 0 {
			sql.WriteString(" OFFSET ?")
			args = append(args, query.Offset)
		}
	}
	
	return sql.String(), args
}

// DeleteRawData deletes raw data older than the specified time
func (s *SQLiteStorage) DeleteRawData(before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	queries := []string{
		"DELETE FROM metrics WHERE timestamp < ?",
		"DELETE FROM vulnerabilities WHERE timestamp < ?",
		"DELETE FROM scan_results WHERE timestamp < ?",
	}
	
	for _, query := range queries {
		_, err := s.db.Exec(query, before)
		if err != nil {
			return fmt.Errorf("failed to delete data: %w", err)
		}
	}
	
	// Vacuum to reclaim space
	_, err := s.db.Exec("VACUUM")
	if err != nil {
		s.logger.Warn("Failed to vacuum database: " + err.Error())
	}
	
	return nil
}

// ArchiveData archives data older than the specified time
func (s *SQLiteStorage) ArchiveData(before time.Time) error {
	// For SQLite, we'll just compress old data by creating aggregated metrics
	return s.createAggregatedMetrics(before)
}

// GetStorageSize returns the storage size in bytes
func (s *SQLiteStorage) GetStorageSize() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var size int64
	err := s.db.QueryRow("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").Scan(&size)
	return size, err
}

// GetRecordCount returns the total number of records
func (s *SQLiteStorage) GetRecordCount() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var count int64
	err := s.db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM metrics) + 
			(SELECT COUNT(*) FROM scan_results) + 
			(SELECT COUNT(*) FROM vulnerabilities)
	`).Scan(&count)
	return count, err
}

// GetAggregatedData returns aggregated metrics data
func (s *SQLiteStorage) GetAggregatedData(query *MetricsQuery) (*MetricsResult, error) {
	// Implementation for aggregated queries
	return s.QueryMetrics(query)
}

// GetTimeSeriesData returns time series data for a specific metric
func (s *SQLiteStorage) GetTimeSeriesData(metric string, timeRange TimeRange) ([]DataPoint, error) {
	query := &MetricsQuery{
		TimeRange: timeRange,
		Metrics:   []string{metric},
	}
	
	result, err := s.QueryMetrics(query)
	if err != nil {
		return nil, err
	}
	
	return result.Data, nil
}

// Helper methods

func (s *SQLiteStorage) extractMetricsFromScanResult(result *ScanResult) []Metric {
	timestamp := result.Timestamp
	tags := map[string]string{
		"target": result.Target,
		"scan_id": result.ID,
	}
	
	metrics := []Metric{
		{
			Name: MetricScanDuration,
			Value: float64(result.Duration.Seconds()),
			Unit: "seconds",
			Tags: tags,
			Timestamp: timestamp,
			Source: "scan_result",
		},
		{
			Name: MetricVulnerabilityCount,
			Value: float64(len(result.Vulnerabilities)),
			Unit: "count",
			Tags: tags,
			Timestamp: timestamp,
			Source: "scan_result",
		},
		{
			Name: MetricTestCount,
			Value: float64(result.TotalTests),
			Unit: "count",
			Tags: tags,
			Timestamp: timestamp,
			Source: "scan_result",
		},
		{
			Name: MetricSuccessRate,
			Value: float64(result.PassedTests) / float64(result.TotalTests) * 100,
			Unit: "percentage",
			Tags: tags,
			Timestamp: timestamp,
			Source: "scan_result",
		},
	}
	
	// Count vulnerabilities by severity
	severityCounts := make(map[string]int)
	for _, vuln := range result.Vulnerabilities {
		severityCounts[vuln.Severity]++
	}
	
	for severity, count := range severityCounts {
		metricName := fmt.Sprintf("%s_vulnerabilities", severity)
		metrics = append(metrics, Metric{
			Name: metricName,
			Value: float64(count),
			Unit: "count",
			Tags: tags,
			Timestamp: timestamp,
			Source: "scan_result",
		})
	}
	
	return metrics
}

func (s *SQLiteStorage) storeMetricInTx(tx *sql.Tx, metric *Metric) error {
	tagsJSON, _ := json.Marshal(metric.Tags)
	
	query := `INSERT INTO metrics (name, value, unit, tags, timestamp, source) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := tx.Exec(query, metric.Name, metric.Value, metric.Unit, 
		string(tagsJSON), metric.Timestamp, metric.Source)
	return err
}

func (s *SQLiteStorage) createAggregatedMetrics(before time.Time) error {
	// Create hourly aggregates for data older than the specified time
	query := `
		INSERT INTO aggregated_metrics (metric_name, aggregation_type, value, interval_start, interval_end, tags)
		SELECT 
			name,
			'avg' as aggregation_type,
			AVG(value) as value,
			datetime(timestamp, 'start of hour') as interval_start,
			datetime(timestamp, 'start of hour', '+1 hour') as interval_end,
			'{}' as tags
		FROM metrics 
		WHERE timestamp < ?
		GROUP BY name, datetime(timestamp, 'start of hour')
	`
	
	_, err := s.db.Exec(query, before)
	return err
}

// Helper functions

func nullableFloat64(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Memory Storage Implementation

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage(config *Config, logger Logger) *MemoryStorage {
	return &MemoryStorage{
		config:      config,
		metrics:     make([]Metric, 0),
		scanResults: make([]ScanResult, 0),
		logger:      logger,
	}
}

// Initialize initializes the memory storage
func (m *MemoryStorage) Initialize() error {
	m.logger.Info("Memory storage initialized")
	return nil
}

// Close closes the memory storage (no-op for memory)
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.metrics = nil
	m.scanResults = nil
	return nil
}

// StoreMetric stores a metric in memory
func (m *MemoryStorage) StoreMetric(metric *Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.metrics = append(m.metrics, *metric)
	
	// Apply retention policy
	if m.config.RetentionPolicy.RawDataDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -m.config.RetentionPolicy.RawDataDays)
		m.cleanupMetrics(cutoff)
	}
	
	return nil
}

// StoreScanResult stores a scan result in memory
func (m *MemoryStorage) StoreScanResult(result *ScanResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.scanResults = append(m.scanResults, *result)
	
	// Store derived metrics
	metrics := m.extractMetricsFromScanResult(result)
	for _, metric := range metrics {
		m.metrics = append(m.metrics, metric)
	}
	
	return nil
}

// QueryMetrics queries metrics from memory
func (m *MemoryStorage) QueryMetrics(query *MetricsQuery) (*MetricsResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	startTime := time.Now()
	
	var filteredMetrics []Metric
	
	// Filter by time range and metrics
	for _, metric := range m.metrics {
		if !query.TimeRange.Contains(metric.Timestamp) {
			continue
		}
		
		if len(query.Metrics) > 0 {
			found := false
			for _, name := range query.Metrics {
				if metric.Name == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Apply filters
		if !m.matchesFilters(metric, query.Filters) {
			continue
		}
		
		filteredMetrics = append(filteredMetrics, metric)
	}
	
	// Convert to data points
	dataPoints := m.metricsToDataPoints(filteredMetrics)
	
	// Apply limit and offset
	if query.Limit > 0 {
		start := query.Offset
		end := start + query.Limit
		
		if start >= len(dataPoints) {
			dataPoints = []DataPoint{}
		} else {
			if end > len(dataPoints) {
				end = len(dataPoints)
			}
			dataPoints = dataPoints[start:end]
		}
	}
	
	return &MetricsResult{
		Query:     query,
		Data:      dataPoints,
		Total:     len(dataPoints),
		QueryTime: time.Since(startTime),
	}, nil
}

// DeleteRawData deletes old metrics from memory
func (m *MemoryStorage) DeleteRawData(before time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.cleanupMetrics(before)
	m.cleanupScanResults(before)
	
	return nil
}

// ArchiveData is a no-op for memory storage
func (m *MemoryStorage) ArchiveData(before time.Time) error {
	return nil
}

// GetStorageSize returns approximate memory usage
func (m *MemoryStorage) GetStorageSize() (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Rough estimate of memory usage
	size := int64(len(m.metrics)*100 + len(m.scanResults)*1000)
	return size, nil
}

// GetRecordCount returns the total number of records in memory
func (m *MemoryStorage) GetRecordCount() (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return int64(len(m.metrics) + len(m.scanResults)), nil
}

// GetAggregatedData returns aggregated data from memory
func (m *MemoryStorage) GetAggregatedData(query *MetricsQuery) (*MetricsResult, error) {
	return m.QueryMetrics(query)
}

// GetTimeSeriesData returns time series data from memory
func (m *MemoryStorage) GetTimeSeriesData(metric string, timeRange TimeRange) ([]DataPoint, error) {
	query := &MetricsQuery{
		TimeRange: timeRange,
		Metrics:   []string{metric},
	}
	
	result, err := m.QueryMetrics(query)
	if err != nil {
		return nil, err
	}
	
	return result.Data, nil
}

// Helper methods for memory storage

func (m *MemoryStorage) cleanupMetrics(before time.Time) {
	filtered := make([]Metric, 0)
	for _, metric := range m.metrics {
		if metric.Timestamp.After(before) {
			filtered = append(filtered, metric)
		}
	}
	m.metrics = filtered
}

func (m *MemoryStorage) cleanupScanResults(before time.Time) {
	filtered := make([]ScanResult, 0)
	for _, result := range m.scanResults {
		if result.Timestamp.After(before) {
			filtered = append(filtered, result)
		}
	}
	m.scanResults = filtered
}

func (m *MemoryStorage) matchesFilters(metric Metric, filters map[string]string) bool {
	for key, value := range filters {
		if tagValue, ok := metric.Tags[key]; !ok || tagValue != value {
			return false
		}
	}
	return true
}

func (m *MemoryStorage) metricsToDataPoints(metrics []Metric) []DataPoint {
	pointMap := make(map[time.Time]*DataPoint)
	
	for _, metric := range metrics {
		key := metric.Timestamp.Truncate(time.Minute) // Group by minute
		
		if point, ok := pointMap[key]; ok {
			point.Values[metric.Name] = metric.Value
		} else {
			pointMap[key] = &DataPoint{
				Timestamp: key,
				Values:    map[string]float64{metric.Name: metric.Value},
				Tags:      metric.Tags,
			}
		}
	}
	
	// Convert map to slice
	dataPoints := make([]DataPoint, 0, len(pointMap))
	for _, point := range pointMap {
		dataPoints = append(dataPoints, *point)
	}
	
	return dataPoints
}

func (m *MemoryStorage) extractMetricsFromScanResult(result *ScanResult) []Metric {
	// Same implementation as SQLite storage
	timestamp := result.Timestamp
	tags := map[string]string{
		"target": result.Target,
		"scan_id": result.ID,
	}
	
	metrics := []Metric{
		{
			Name: MetricScanDuration,
			Value: float64(result.Duration.Seconds()),
			Unit: "seconds",
			Tags: tags,
			Timestamp: timestamp,
			Source: "scan_result",
		},
		{
			Name: MetricVulnerabilityCount,
			Value: float64(len(result.Vulnerabilities)),
			Unit: "count",
			Tags: tags,
			Timestamp: timestamp,
			Source: "scan_result",
		},
	}
	
	return metrics
}

// Placeholder implementations for other storage types

func NewPostgreSQLStorage(config *Config, logger Logger) (DataStorage, error) {
	return nil, fmt.Errorf("PostgreSQL storage not yet implemented")
}

func NewMySQLStorage(config *Config, logger Logger) (DataStorage, error) {
	return nil, fmt.Errorf("MySQL storage not yet implemented")
}

func NewInfluxDBStorage(config *Config, logger Logger) (DataStorage, error) {
	return nil, fmt.Errorf("InfluxDB storage not yet implemented")
}