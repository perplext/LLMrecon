// Package monitoring provides performance monitoring for template operations.
package monitoring

import (
	"fmt"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/management/benchmark"
)

// PerformanceMonitor tracks performance metrics for template operations
type PerformanceMonitor struct {
	// metrics is a map of metric name to metric data
	metrics map[string]*PerformanceMetric
	// mutex protects the metrics map
	mutex sync.RWMutex
	// startTime is the time the monitor was started
	startTime time.Time
	// config is the monitor configuration
	config *MonitorConfig
}

// PerformanceMetric represents a performance metric
type PerformanceMetric struct {
	// Name is the name of the metric
	Name string
	// Description is the description of the metric
	Description string
	// Value is the current value of the metric
	Value float64
	// Count is the number of times the metric has been updated
	Count int64
	// Min is the minimum value of the metric
	Min float64
	// Max is the maximum value of the metric
	Max float64
	// Sum is the sum of all values of the metric
	Sum float64
	// LastUpdated is the time the metric was last updated
	LastUpdated time.Time
	// Unit is the unit of the metric
	Unit string
	// Tags are additional tags for the metric
	Tags map[string]string
	// History stores historical values for trending
	History []HistoryEntry
}

// HistoryEntry represents a historical metric value
type HistoryEntry struct {
	// Timestamp is the time the metric was recorded
	Timestamp time.Time
	// Value is the value of the metric
	Value float64
}

// MonitorConfig contains configuration for the performance monitor
type MonitorConfig struct {
	// HistorySize is the number of historical values to keep
	HistorySize int
	// SamplingInterval is the interval at which to sample metrics
	SamplingInterval time.Duration
	// AlertThresholds is a map of metric name to alert threshold
	AlertThresholds map[string]float64
	// EnableAlerts enables or disables alerts
	EnableAlerts bool
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config *MonitorConfig) *PerformanceMonitor {
	// Set default values
	if config == nil {
		config = &MonitorConfig{
			HistorySize:      100,
			SamplingInterval: 1 * time.Minute,
			AlertThresholds:  make(map[string]float64),
			EnableAlerts:     false,
		}
	}

	return &PerformanceMonitor{
		metrics:   make(map[string]*PerformanceMetric),
		startTime: time.Now(),
		config:    config,
	}
}

// RegisterMetric registers a new metric
func (m *PerformanceMonitor) RegisterMetric(name string, description string, unit string, tags map[string]string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if metric already exists
	if _, exists := m.metrics[name]; exists {
		return
	}

	// Create new metric
	m.metrics[name] = &PerformanceMetric{
		Name:        name,
		Description: description,
		Min:         float64(^uint64(0) >> 1), // Max float64 value
		Max:         -1,
		Unit:        unit,
		Tags:        tags,
		LastUpdated: time.Now(),
		History:     make([]HistoryEntry, 0, m.config.HistorySize),
	}
}

// RecordMetric records a value for a metric
func (m *PerformanceMonitor) RecordMetric(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if metric exists
	metric, exists := m.metrics[name]
	if !exists {
		// Create new metric with default values
		metric = &PerformanceMetric{
			Name:        name,
			Description: name,
			Min:         float64(^uint64(0) >> 1), // Max float64 value
			Max:         -1,
			Unit:        "count",
			Tags:        make(map[string]string),
			LastUpdated: time.Now(),
			History:     make([]HistoryEntry, 0, m.config.HistorySize),
		}
		m.metrics[name] = metric
	}

	// Update metric
	metric.Value = value
	metric.Count++
	metric.Sum += value
	metric.LastUpdated = time.Now()

	// Update min/max
	if value < metric.Min {
		metric.Min = value
	}
	if value > metric.Max {
		metric.Max = value
	}

	// Add to history
	historyEntry := HistoryEntry{
		Timestamp: time.Now(),
		Value:     value,
	}
	
	// Maintain history size
	if len(metric.History) >= m.config.HistorySize {
		// Remove oldest entry
		metric.History = metric.History[1:]
	}
	
	// Add new entry
	metric.History = append(metric.History, historyEntry)

	// Check alert threshold
	if m.config.EnableAlerts {
		threshold, hasThreshold := m.config.AlertThresholds[name]
		if hasThreshold && value > threshold {
			m.triggerAlert(name, value, threshold)
		}
	}
}

// GetMetric gets a metric by name
func (m *PerformanceMonitor) GetMetric(name string) (*PerformanceMetric, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metric, exists := m.metrics[name]
	return metric, exists
}

// GetMetrics gets all metrics
func (m *PerformanceMonitor) GetMetrics() map[string]*PerformanceMetric {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a copy of the metrics map
	metrics := make(map[string]*PerformanceMetric, len(m.metrics))
	for name, metric := range m.metrics {
		metrics[name] = metric
	}

	return metrics
}

// GetMetricAverage gets the average value of a metric
func (m *PerformanceMonitor) GetMetricAverage(name string) (float64, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metric, exists := m.metrics[name]
	if !exists || metric.Count == 0 {
		return 0, false
	}

	return metric.Sum / float64(metric.Count), true
}

// GetMetricHistory gets the history of a metric
func (m *PerformanceMonitor) GetMetricHistory(name string) ([]HistoryEntry, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metric, exists := m.metrics[name]
	if !exists {
		return nil, false
	}

	// Create a copy of the history
	history := make([]HistoryEntry, len(metric.History))
	copy(history, metric.History)

	return history, true
}

// GetMetricsByTag gets metrics by tag
func (m *PerformanceMonitor) GetMetricsByTag(tag string, value string) map[string]*PerformanceMetric {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metrics := make(map[string]*PerformanceMetric)
	for name, metric := range m.metrics {
		if tagValue, hasTag := metric.Tags[tag]; hasTag && tagValue == value {
			metrics[name] = metric
		}
	}

	return metrics
}

// ResetMetric resets a metric
func (m *PerformanceMonitor) ResetMetric(name string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric, exists := m.metrics[name]
	if !exists {
		return false
	}

	// Reset metric
	metric.Value = 0
	metric.Count = 0
	metric.Min = float64(^uint64(0) >> 1) // Max float64 value
	metric.Max = -1
	metric.Sum = 0
	metric.LastUpdated = time.Now()
	metric.History = make([]HistoryEntry, 0, m.config.HistorySize)

	return true
}

// ResetAllMetrics resets all metrics
func (m *PerformanceMonitor) ResetAllMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, metric := range m.metrics {
		// Reset metric
		metric.Value = 0
		metric.Count = 0
		metric.Min = float64(^uint64(0) >> 1) // Max float64 value
		metric.Max = -1
		metric.Sum = 0
		metric.LastUpdated = time.Now()
		metric.History = make([]HistoryEntry, 0, m.config.HistorySize)
	}
}

// RemoveMetric removes a metric
func (m *PerformanceMonitor) RemoveMetric(name string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, exists := m.metrics[name]
	if !exists {
		return false
	}

	delete(m.metrics, name)
	return true
}

// SetAlertThreshold sets an alert threshold for a metric
func (m *PerformanceMonitor) SetAlertThreshold(name string, threshold float64) {
	m.config.AlertThresholds[name] = threshold
}

// EnableAlerts enables alerts
func (m *PerformanceMonitor) EnableAlerts(enable bool) {
	m.config.EnableAlerts = enable
}

// GetUptime gets the uptime of the monitor
func (m *PerformanceMonitor) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// GetReport generates a performance report
func (m *PerformanceMonitor) GetReport() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	report := make(map[string]interface{})
	report["uptime"] = m.GetUptime().String()
	report["metric_count"] = len(m.metrics)
	
	metricSummaries := make(map[string]map[string]interface{})
	for name, metric := range m.metrics {
		summary := make(map[string]interface{})
		summary["value"] = metric.Value
		summary["count"] = metric.Count
		summary["min"] = metric.Min
		summary["max"] = metric.Max
		
		// Calculate average
		avg := float64(0)
		if metric.Count > 0 {
			avg = metric.Sum / float64(metric.Count)
		}
		summary["avg"] = avg
		
		summary["unit"] = metric.Unit
		summary["last_updated"] = metric.LastUpdated.Format(time.RFC3339)
		
		metricSummaries[name] = summary
	}
	
	report["metrics"] = metricSummaries
	
	return report
}

// triggerAlert triggers an alert for a metric
func (m *PerformanceMonitor) triggerAlert(name string, value float64, threshold float64) {
	// This is a placeholder for alert functionality
	// In a real implementation, this would send an alert to a monitoring system
	fmt.Printf("ALERT: Metric %s exceeded threshold %.2f with value %.2f\n", name, threshold, value)
}

// RecordBenchmarkResult records a benchmark result as metrics
func (m *PerformanceMonitor) RecordBenchmarkResult(result *benchmark.BenchmarkResult) {
	// Record main metrics
	m.RecordMetric(fmt.Sprintf("%s.duration_ms", result.Name), float64(result.Duration.Milliseconds()))
	m.RecordMetric(fmt.Sprintf("%s.ops_per_sec", result.Name), result.OperationsPerSecond)
	m.RecordMetric(fmt.Sprintf("%s.avg_latency_ms", result.Name), float64(result.AverageLatency.Milliseconds()))
	m.RecordMetric(fmt.Sprintf("%s.memory_usage_mb", result.Name), float64(result.MemoryUsage)/(1024*1024))
	m.RecordMetric(fmt.Sprintf("%s.errors", result.Name), float64(result.Errors))
	
	// Record operation count
	m.RecordMetric(fmt.Sprintf("%s.operation_count", result.Name), float64(result.OperationCount))
}

// RecordBenchmarkResults records multiple benchmark results
func (m *PerformanceMonitor) RecordBenchmarkResults(results map[string]*benchmark.BenchmarkResult) {
	for _, result := range results {
		m.RecordBenchmarkResult(result)
	}
}

// GetMetricTrend gets the trend of a metric over time
func (m *PerformanceMonitor) GetMetricTrend(name string, duration time.Duration) (float64, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metric, exists := m.metrics[name]
	if !exists || len(metric.History) < 2 {
		return 0, false
	}

	// Get history entries within the specified duration
	now := time.Now()
	startTime := now.Add(-duration)
	
	var oldestValue, newestValue float64
	var oldestFound, newestFound bool
	
	// Find oldest and newest values within the duration
	for _, entry := range metric.History {
		if entry.Timestamp.After(startTime) {
			if !oldestFound {
				oldestValue = entry.Value
				oldestFound = true
			}
			newestValue = entry.Value
			newestFound = true
		}
	}
	
	if !oldestFound || !newestFound {
		return 0, false
	}
	
	// Calculate trend (positive means increasing, negative means decreasing)
	return newestValue - oldestValue, true
}

// GetPerformanceSummary gets a summary of performance metrics
func (m *PerformanceMonitor) GetPerformanceSummary() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	summary := make(map[string]interface{})
	summary["uptime"] = m.GetUptime().String()
	
	// Group metrics by category
	categories := make(map[string]map[string]interface{})
	
	for name, metric := range m.metrics {
		// Extract category from metric name (e.g., "template.load.duration" -> "template")
		var category string
		if len(name) > 0 && name[0] != '.' {
			for i, c := range name {
				if c == '.' {
					category = name[:i]
					break
				}
			}
			if category == "" {
				category = "other"
			}
		} else {
			category = "other"
		}
		
		// Create category if it doesn't exist
		if _, exists := categories[category]; !exists {
			categories[category] = make(map[string]interface{})
		}
		
		// Add metric to category
		metricName := name
		if len(category) > 0 && len(name) > len(category) {
			metricName = name[len(category)+1:] // +1 for the dot
		}
		
		// Calculate average
		avg := float64(0)
		if metric.Count > 0 {
			avg = metric.Sum / float64(metric.Count)
		}
		
		categories[category][metricName] = map[string]interface{}{
			"value": metric.Value,
			"avg":   avg,
			"min":   metric.Min,
			"max":   metric.Max,
			"unit":  metric.Unit,
		}
	}
	
	summary["categories"] = categories
	
	return summary
}
