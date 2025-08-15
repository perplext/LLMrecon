package analytics

import (
	"context"
	"fmt"
	"strings"
)

// ValidationProcessor validates metrics before storage
type ValidationProcessor struct {
	enabled bool
}

func (vp *ValidationProcessor) Process(ctx context.Context, metric Metric) (Metric, error) {
	// Validate required fields
	if metric.Name == "" {
		return metric, fmt.Errorf("metric name cannot be empty")
	}
	
	if metric.ID == "" {
		return metric, fmt.Errorf("metric ID cannot be empty")
	}
	
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}
	
	// Validate metric type
	switch metric.Type {
	case MetricTypeCounter, MetricTypeGauge, MetricTypeHistogram, MetricTypeEvent, MetricTypeCustom:
		// Valid types
	default:
		return metric, fmt.Errorf("invalid metric type: %s", metric.Type)
	}
	
	// Sanitize metric name
	metric.Name = strings.ToLower(strings.ReplaceAll(metric.Name, " ", "_"))
	
	// Ensure labels is not nil
	if metric.Labels == nil {
		metric.Labels = make(map[string]string)
	}
	
	// Ensure metadata is not nil
	if metric.Metadata == nil {
		metric.Metadata = make(map[string]interface{})
	}
	
	return metric, nil

func (vp *ValidationProcessor) GetType() string {
	return "validation"

func (vp *ValidationProcessor) IsEnabled() bool {
	return true // Always enabled

// EnrichmentProcessor adds additional context to metrics
type EnrichmentProcessor struct {
	enabled bool
}

func (ep *EnrichmentProcessor) Process(ctx context.Context, metric Metric) (Metric, error) {
	// Add system context
	if metric.Labels == nil {
		metric.Labels = make(map[string]string)
	}
	
	// Add timestamp-based labels
	metric.Labels["hour"] = fmt.Sprintf("%02d", metric.Timestamp.Hour())
	metric.Labels["day_of_week"] = metric.Timestamp.Weekday().String()
	metric.Labels["month"] = metric.Timestamp.Month().String()
	
	// Add environment context
	metric.Labels["environment"] = "production" // This would come from config
	
	// Enrich based on metric type
	switch metric.Type {
	case MetricTypeEvent:
		if metric.Metadata == nil {
			metric.Metadata = make(map[string]interface{})
		}
		metric.Metadata["event_timestamp"] = metric.Timestamp.Unix()
		metric.Metadata["event_day"] = metric.Timestamp.Format("2006-01-02")
		
	case MetricTypeHistogram:
		// Add histogram buckets for better aggregation
		value := metric.Value
		bucket := getBucket(value)
		metric.Labels["bucket"] = bucket
		
	case MetricTypeCounter:
		// Add rate calculation metadata
		if metric.Metadata == nil {
			metric.Metadata = make(map[string]interface{})
		}
		metric.Metadata["counter_timestamp"] = metric.Timestamp.Unix()
	}
	
	// Add performance classification
	if strings.Contains(metric.Name, "duration") || strings.Contains(metric.Name, "time") {
		classification := classifyPerformance(metric.Value)
		metric.Labels["performance_class"] = classification
	}
	
	return metric, nil

func (ep *EnrichmentProcessor) GetType() string {
	return "enrichment"

func (ep *EnrichmentProcessor) IsEnabled() bool {
	return true

// FilteringProcessor filters metrics based on configuration
type FilteringProcessor struct {
	config  *Config
	enabled bool

func (fp *FilteringProcessor) Process(ctx context.Context, metric Metric) (Metric, error) {
	// Skip metrics that match exclusion patterns
	for _, pattern := range fp.config.Analytics.ExcludePatterns {
		if strings.Contains(metric.Name, pattern) {
			return metric, fmt.Errorf("metric filtered out by pattern: %s", pattern)
		}
	}
	
	// Skip metrics below minimum value threshold
	if fp.config.Analytics.MinValue > 0 && metric.Value < fp.config.Analytics.MinValue {
		return metric, fmt.Errorf("metric value below threshold: %f < %f", metric.Value, fp.config.Analytics.MinValue)
	}
	
	// Skip old metrics
	if time.Since(metric.Timestamp) > fp.config.Analytics.MaxAge {
		return metric, fmt.Errorf("metric too old: %v", time.Since(metric.Timestamp))
	}
	
	return metric, nil

func (fp *FilteringProcessor) GetType() string {
	return "filtering"

func (fp *FilteringProcessor) IsEnabled() bool {
	return fp.config.Analytics.FilteringEnabled

// BasicAggregator provides basic statistical aggregations
type BasicAggregator struct {
	windowSizes []time.Duration
}

func (ba *BasicAggregator) Aggregate(ctx context.Context, metrics []Metric, window TimeWindow) (AggregatedMetric, error) {
	if len(metrics) == 0 {
		return AggregatedMetric{}, fmt.Errorf("no metrics to aggregate")
	}
	
	// Group metrics by name
	metricGroups := make(map[string][]Metric)
	for _, metric := range metrics {
		metricGroups[metric.Name] = append(metricGroups[metric.Name], metric)
	}
	
	aggregations := make(map[string]interface{})
	
	for name, groupMetrics := range metricGroups {
		stats := calculateBasicStats(groupMetrics)
		aggregations[name] = stats
	}
	
	return AggregatedMetric{
		ID:           generateMetricID(),
		TimeWindow:   window,
		MetricCount:  len(metrics),
		Aggregations: aggregations,
		CreatedAt:    time.Now(),
		Type:         "basic",
	}, nil

func (ba *BasicAggregator) GetWindowSizes() []time.Duration {
	if len(ba.windowSizes) == 0 {
		return []time.Duration{
			5 * time.Minute,
			15 * time.Minute,
			time.Hour,
			24 * time.Hour,
		}
	}
	return ba.windowSizes

func (ba *BasicAggregator) Reset() {
	// Basic aggregator is stateless, nothing to reset

// PerformanceAggregator focuses on performance-related metrics
type PerformanceAggregator struct {
	windowSizes []time.Duration
}

func (pa *PerformanceAggregator) Aggregate(ctx context.Context, metrics []Metric, window TimeWindow) (AggregatedMetric, error) {
	// Filter for performance metrics
	perfMetrics := filterPerformanceMetrics(metrics)
	
	if len(perfMetrics) == 0 {
		return AggregatedMetric{}, fmt.Errorf("no performance metrics found")
	}
	
	aggregations := make(map[string]interface{})
	
	// Calculate performance percentiles
	durations := extractDurationValues(perfMetrics)
	if len(durations) > 0 {
		percentiles := calculatePercentiles(durations)
		aggregations["duration_percentiles"] = percentiles
		
		// Calculate throughput
		throughput := float64(len(perfMetrics)) / window.Duration.Seconds()
		aggregations["throughput"] = throughput
		
		// Calculate error rate
		errorRate := calculateErrorRate(perfMetrics)
		aggregations["error_rate"] = errorRate
	}
	
	// Calculate resource utilization trends
	resourceMetrics := filterResourceMetrics(perfMetrics)
	if len(resourceMetrics) > 0 {
		utilization := calculateResourceUtilization(resourceMetrics)
		aggregations["resource_utilization"] = utilization
	}
	
	return AggregatedMetric{
		ID:           generateMetricID(),
		TimeWindow:   window,
		MetricCount:  len(perfMetrics),
		Aggregations: aggregations,
		CreatedAt:    time.Now(),
		Type:         "performance",
	}, nil

func (pa *PerformanceAggregator) GetWindowSizes() []time.Duration {
	if len(pa.windowSizes) == 0 {
		return []time.Duration{
			time.Minute,
			5 * time.Minute,
			15 * time.Minute,
			time.Hour,
		}
	}
	return pa.windowSizes

func (pa *PerformanceAggregator) Reset() {
	// Performance aggregator is stateless, nothing to reset

// SecurityAggregator focuses on security-related metrics
type SecurityAggregator struct {
	windowSizes    []time.Duration
	threatPatterns []string

func (sa *SecurityAggregator) Aggregate(ctx context.Context, metrics []Metric, window TimeWindow) (AggregatedMetric, error) {
	// Filter for security metrics
	securityMetrics := filterSecurityMetrics(metrics)
	
	if len(securityMetrics) == 0 {
		return AggregatedMetric{}, fmt.Errorf("no security metrics found")
	}
	
	aggregations := make(map[string]interface{})
	
	// Calculate vulnerability distribution
	vulnDistribution := calculateVulnerabilityDistribution(securityMetrics)
	aggregations["vulnerability_distribution"] = vulnDistribution
	
	// Calculate threat severity trends
	severityTrends := calculateSeverityTrends(securityMetrics)
	aggregations["severity_trends"] = severityTrends
	
	// Detect anomalies
	anomalies := detectSecurityAnomalies(securityMetrics, window)
	aggregations["anomalies"] = anomalies
	
	// Calculate OWASP category coverage
	owaspCoverage := calculateOWASPCoverage(securityMetrics)
	aggregations["owasp_coverage"] = owaspCoverage
	
	return AggregatedMetric{
		ID:           generateMetricID(),
		TimeWindow:   window,
		MetricCount:  len(securityMetrics),
		Aggregations: aggregations,
		CreatedAt:    time.Now(),
		Type:         "security",
	}, nil

func (sa *SecurityAggregator) GetWindowSizes() []time.Duration {
	if len(sa.windowSizes) == 0 {
		return []time.Duration{
			15 * time.Minute,
			time.Hour,
			6 * time.Hour,
			24 * time.Hour,
		}
	}
	return sa.windowSizes

func (sa *SecurityAggregator) Reset() {
	// Security aggregator is stateless, nothing to reset

// Utility functions

func getBucket(value float64) string {
	switch {
	case value < 0.1:
		return "very_fast"
	case value < 1.0:
		return "fast"
	case value < 5.0:
		return "normal"
	case value < 10.0:
		return "slow"
	default:
		return "very_slow"
	}

func classifyPerformance(value float64) string {
	switch {
	case value < 1.0:
		return "excellent"
	case value < 3.0:
		return "good"
	case value < 5.0:
		return "acceptable"
	case value < 10.0:
		return "poor"
	default:
		return "critical"
	}

func calculateBasicStats(metrics []Metric) map[string]interface{} {
	values := make([]float64, len(metrics))
	for i, metric := range metrics {
		values[i] = metric.Value
	}
	
	return map[string]interface{}{
		"count": len(values),
		"sum":   sum(values),
		"avg":   average(values),
		"min":   min(values),
		"max":   max(values),
		"std":   standardDeviation(values),
	}

func filterPerformanceMetrics(metrics []Metric) []Metric {
	var perfMetrics []Metric
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "duration") ||
			strings.Contains(metric.Name, "time") ||
			strings.Contains(metric.Name, "latency") ||
			strings.Contains(metric.Name, "throughput") ||
			strings.Contains(metric.Name, "cpu") ||
			strings.Contains(metric.Name, "memory") {
			perfMetrics = append(perfMetrics, metric)
		}
	}
	return perfMetrics

func filterSecurityMetrics(metrics []Metric) []Metric {
	var securityMetrics []Metric
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "vulnerability") ||
			strings.Contains(metric.Name, "threat") ||
			strings.Contains(metric.Name, "security") ||
			strings.Contains(metric.Name, "scan") ||
			strings.Contains(metric.Name, "attack") {
			securityMetrics = append(securityMetrics, metric)
		}
	}
	return securityMetrics

func filterResourceMetrics(metrics []Metric) []Metric {
	var resourceMetrics []Metric
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "cpu") ||
			strings.Contains(metric.Name, "memory") ||
			strings.Contains(metric.Name, "disk") ||
			strings.Contains(metric.Name, "network") {
			resourceMetrics = append(resourceMetrics, metric)
		}
	}
	return resourceMetrics

func extractDurationValues(metrics []Metric) []float64 {
	var durations []float64
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "duration") || strings.Contains(metric.Name, "time") {
			durations = append(durations, metric.Value)
		}
	}
	return durations

func calculatePercentiles(values []float64) map[string]float64 {
	if len(values) == 0 {
		return make(map[string]float64)
	}
	
	// Sort values for percentile calculation
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	
	// Simple bubble sort for demonstration
	for i := 0; i < len(sortedValues); i++ {
		for j := 0; j < len(sortedValues)-1-i; j++ {
			if sortedValues[j] > sortedValues[j+1] {
				sortedValues[j], sortedValues[j+1] = sortedValues[j+1], sortedValues[j]
			}
		}
	}
	
	return map[string]float64{
		"p50":  percentile(sortedValues, 0.5),
		"p90":  percentile(sortedValues, 0.9),
		"p95":  percentile(sortedValues, 0.95),
		"p99":  percentile(sortedValues, 0.99),
		"p999": percentile(sortedValues, 0.999),
	}

func calculateErrorRate(metrics []Metric) float64 {
	var total, errors int
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "test") || strings.Contains(metric.Name, "scan") {
			total++
			if errorValue, exists := metric.Metadata["error"]; exists {
				if errorBool, ok := errorValue.(bool); ok && errorBool {
					errors++
				}
			}
		}
	}
	
	if total == 0 {
		return 0.0
	}
	return float64(errors) / float64(total) * 100.0

func calculateResourceUtilization(metrics []Metric) map[string]float64 {
	utilization := make(map[string]float64)
	
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "cpu") {
			utilization["cpu"] = metric.Value
		} else if strings.Contains(metric.Name, "memory") {
			utilization["memory"] = metric.Value
		} else if strings.Contains(metric.Name, "disk") {
			utilization["disk"] = metric.Value
		}
	}
	
	return utilization

func calculateVulnerabilityDistribution(metrics []Metric) map[string]int {
	distribution := make(map[string]int)
	
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "vulnerability") {
			if severity, exists := metric.Labels["severity"]; exists {
				distribution[severity]++
			}
		}
	}
	
	return distribution

func calculateSeverityTrends(metrics []Metric) map[string][]float64 {
	trends := make(map[string][]float64)
	
	// Group by severity and collect values over time
	for _, metric := range metrics {
		if severity, exists := metric.Labels["severity"]; exists {
			trends[severity] = append(trends[severity], metric.Value)
		}
	}
	
	return trends

func detectSecurityAnomalies(metrics []Metric, window TimeWindow) []map[string]interface{} {
	var anomalies []map[string]interface{}
	
	// Simple anomaly detection based on value thresholds
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "vulnerability") && metric.Value > 10 {
			anomaly := map[string]interface{}{
				"metric_name": metric.Name,
				"value":       metric.Value,
				"timestamp":   metric.Timestamp,
				"type":        "high_vulnerability_count",
				"severity":    "high",
			}
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies

func calculateOWASPCoverage(metrics []Metric) map[string]float64 {
	coverage := make(map[string]float64)
	owaspCategories := []string{
		"llm01", "llm02", "llm03", "llm04", "llm05",
		"llm06", "llm07", "llm08", "llm09", "llm10",
	}
	
	totalTests := len(metrics)
	if totalTests == 0 {
		return coverage
	}
	
	for _, category := range owaspCategories {
		count := 0
		for _, metric := range metrics {
			if strings.Contains(metric.Name, category) {
				count++
			}
		}
		coverage[category] = float64(count) / float64(totalTests) * 100.0
	}
	
	return coverage

// Mathematical utility functions
func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	return sum(values) / float64(len(values))

func min(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	minVal := values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal

func standardDeviation(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}
	
	avg := average(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - avg
		sumSquares += diff * diff
	}
	
	variance := sumSquares / float64(len(values)-1)
	return sqrt(variance)

func percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0.0
	}
	if len(sortedValues) == 1 {
		return sortedValues[0]
	}
	
	index := p * float64(len(sortedValues)-1)
	lower := int(index)
	upper := lower + 1
	
	if upper >= len(sortedValues) {
		return sortedValues[len(sortedValues)-1]
	}
	
	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight

// Simple square root implementation for demonstration
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
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
}
}
}
}
}
}
}
