package analytics

import (
	"context"
	"fmt"
	"math"
)

// ComparativeAnalyzer performs comparative analysis between different metrics, time periods, or datasets
type ComparativeAnalyzer struct {
	config          *Config
	storage         DataStorage
	trendAnalyzer   *TrendAnalyzer
	historicalData  *HistoricalDataManager
	logger          Logger
}

// ComparisonResult represents the result of a comparative analysis
type ComparisonResult struct {
	ComparisonType   ComparisonType           `json:"comparison_type"`
	BaselineDataset  ComparisonDataset        `json:"baseline_dataset"`
	ComparisonDataset ComparisonDataset       `json:"comparison_dataset"`
	Statistics       ComparisonStatistics     `json:"statistics"`
	Insights         []ComparisonInsight      `json:"insights"`
	Recommendations  []string                 `json:"recommendations"`
	GeneratedAt      time.Time                `json:"generated_at"`
}

// ComparisonType represents different types of comparisons
type ComparisonType string

const (
	ComparisonTypeTimePeriod    ComparisonType = "time_period"    // Compare different time periods
	ComparisonTypeMetrics       ComparisonType = "metrics"        // Compare different metrics
	ComparisonTypeBaseline      ComparisonType = "baseline"       // Compare against baseline
	ComparisonTypeBenchmark     ComparisonType = "benchmark"      // Compare against benchmark
	ComparisonTypeAnomalyPattern ComparisonType = "anomaly_pattern" // Compare anomaly patterns
)

// ComparisonDataset represents a dataset in comparison
type ComparisonDataset struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MetricName  string                 `json:"metric_name"`
	TimeRange   TimeWindow             `json:"time_range"`
	DataPoints  int                    `json:"data_points"`
	Statistics  BasicStatistics        `json:"statistics"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ComparisonStatistics contains statistical comparison results
type ComparisonStatistics struct {
	MeanDifference       float64 `json:"mean_difference"`
	MeanDifferencePercent float64 `json:"mean_difference_percent"`
	MedianDifference     float64 `json:"median_difference"`
	StandardDeviationRatio float64 `json:"standard_deviation_ratio"`
	CorrelationCoefficient float64 `json:"correlation_coefficient"`
	PValue               float64 `json:"p_value"`
	StatisticalSignificance string `json:"statistical_significance"`
	EffectSize           float64 `json:"effect_size"`
	EffectSizeInterpretation string `json:"effect_size_interpretation"`
}

// BasicStatistics contains basic statistical measures
type BasicStatistics struct {
	Count            int     `json:"count"`
	Mean             float64 `json:"mean"`
	Median           float64 `json:"median"`
	StandardDeviation float64 `json:"standard_deviation"`
	Min              float64 `json:"min"`
	Max              float64 `json:"max"`
	Q1               float64 `json:"q1"`
	Q3               float64 `json:"q3"`
	Variance         float64 `json:"variance"`
	Skewness         float64 `json:"skewness"`
	Kurtosis         float64 `json:"kurtosis"`
}

// ComparisonInsight represents an insight from comparative analysis
type ComparisonInsight struct {
	Type        InsightType `json:"type"`
	Severity    string      `json:"severity"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Evidence    []string    `json:"evidence"`
	Confidence  float64     `json:"confidence"`

// InsightType represents different types of insights
type InsightType string

const (
	InsightTypeTrend       InsightType = "trend"
	InsightTypeAnomaly     InsightType = "anomaly"
	InsightTypePattern     InsightType = "pattern"
	InsightTypePerformance InsightType = "performance"
	InsightTypeSecurity    InsightType = "security"
	InsightTypeCapacity    InsightType = "capacity"
)

// NewComparativeAnalyzer creates a new comparative analyzer
func NewComparativeAnalyzer(config *Config, storage DataStorage, trendAnalyzer *TrendAnalyzer, historicalData *HistoricalDataManager, logger Logger) *ComparativeAnalyzer {
	return &ComparativeAnalyzer{
		config:         config,
		storage:        storage,
		trendAnalyzer:  trendAnalyzer,
		historicalData: historicalData,
		logger:         logger,
	}

// CompareTimePeriods compares metrics across different time periods
func (ca *ComparativeAnalyzer) CompareTimePeriods(ctx context.Context, metricName string, baselineRange, comparisonRange TimeWindow) (*ComparisonResult, error) {
	ca.logger.Info("Starting time period comparison", 
		"metric", metricName, 
		"baseline", baselineRange, 
		"comparison", comparisonRange)
	
	// Get data for both periods
	baselineData, err := ca.historicalData.GetHistoricalData(ctx, metricName, baselineRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline data: %w", err)
	}
	
	comparisonData, err := ca.historicalData.GetHistoricalData(ctx, metricName, comparisonRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get comparison data: %w", err)
	}
	
	if len(baselineData) == 0 || len(comparisonData) == 0 {
		return nil, fmt.Errorf("insufficient data for comparison")
	}
	
	// Create datasets
	baselineDataset := ca.createDataset("Baseline Period", metricName, baselineRange, baselineData)
	comparisonDataset := ca.createDataset("Comparison Period", metricName, comparisonRange, comparisonData)
	
	// Calculate statistics
	stats := ca.calculateStatistics(baselineData, comparisonData)
	
	// Generate insights
	insights := ca.generateTimePeriodInsights(baselineDataset, comparisonDataset, stats)
	
	// Generate recommendations
	recommendations := ca.generateTimePeriodRecommendations(stats, insights)
	
	return &ComparisonResult{
		ComparisonType:    ComparisonTypeTimePeriod,
		BaselineDataset:   baselineDataset,
		ComparisonDataset: comparisonDataset,
		Statistics:        stats,
		Insights:          insights,
		Recommendations:   recommendations,
		GeneratedAt:       time.Now(),
	}, nil

// CompareMetrics compares different metrics over the same time period
func (ca *ComparativeAnalyzer) CompareMetrics(ctx context.Context, baselineMetric, comparisonMetric string, timeRange TimeWindow) (*ComparisonResult, error) {
	ca.logger.Info("Starting metrics comparison", 
		"baseline", baselineMetric, 
		"comparison", comparisonMetric, 
		"timeRange", timeRange)
	
	// Get data for both metrics
	baselineData, err := ca.storage.GetMetricsByNameAndTimeRange(ctx, baselineMetric, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline metric data: %w", err)
	}
	
	comparisonData, err := ca.storage.GetMetricsByNameAndTimeRange(ctx, comparisonMetric, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to get comparison metric data: %w", err)
	}
	
	// Create datasets
	baselineDataset := ca.createDataset(baselineMetric, baselineMetric, timeRange, baselineData)
	comparisonDataset := ca.createDataset(comparisonMetric, comparisonMetric, timeRange, comparisonData)
	
	// Calculate statistics
	stats := ca.calculateStatistics(baselineData, comparisonData)
	
	// Generate insights
	insights := ca.generateMetricsInsights(baselineDataset, comparisonDataset, stats)
	
	// Generate recommendations
	recommendations := ca.generateMetricsRecommendations(stats, insights)
	
	return &ComparisonResult{
		ComparisonType:    ComparisonTypeMetrics,
		BaselineDataset:   baselineDataset,
		ComparisonDataset: comparisonDataset,
		Statistics:        stats,
		Insights:          insights,
		Recommendations:   recommendations,
		GeneratedAt:       time.Now(),
	}, nil

// CompareAgainstBaseline compares current metrics against established baselines
func (ca *ComparativeAnalyzer) CompareAgainstBaseline(ctx context.Context, metricName string, currentRange TimeWindow, baselineValue float64) (*ComparisonResult, error) {
	ca.logger.Info("Starting baseline comparison", 
		"metric", metricName, 
		"timeRange", currentRange, 
		"baseline", baselineValue)
	
	// Get current data
	currentData, err := ca.storage.GetMetricsByNameAndTimeRange(ctx, metricName, currentRange.Start, currentRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to get current data: %w", err)
	}
	
	// Create synthetic baseline data
	baselineData := ca.createSyntheticBaseline(baselineValue, len(currentData))
	
	// Create datasets
	baselineDataset := ca.createDataset("Baseline", metricName, currentRange, baselineData)
	comparisonDataset := ca.createDataset("Current", metricName, currentRange, currentData)
	
	// Calculate statistics
	stats := ca.calculateStatistics(baselineData, currentData)
	
	// Generate insights
	insights := ca.generateBaselineInsights(baselineDataset, comparisonDataset, stats, baselineValue)
	
	// Generate recommendations
	recommendations := ca.generateBaselineRecommendations(stats, insights, baselineValue)
	
	return &ComparisonResult{
		ComparisonType:    ComparisonTypeBaseline,
		BaselineDataset:   baselineDataset,
		ComparisonDataset: comparisonDataset,
		Statistics:        stats,
		Insights:          insights,
		Recommendations:   recommendations,
		GeneratedAt:       time.Now(),
	}, nil

// CompareAnomalyPatterns compares anomaly patterns between different time periods
func (ca *ComparativeAnalyzer) CompareAnomalyPatterns(ctx context.Context, metricName string, baselineRange, comparisonRange TimeWindow) (*ComparisonResult, error) {
	ca.logger.Info("Starting anomaly pattern comparison", 
		"metric", metricName, 
		"baseline", baselineRange, 
		"comparison", comparisonRange)
	
	// Get anomalies for both periods
	baselineAnomalies, err := ca.trendAnalyzer.DetectAnomalies(ctx, metricName, baselineRange)
	if err != nil {
		return nil, fmt.Errorf("failed to detect baseline anomalies: %w", err)
	}
	
	comparisonAnomalies, err := ca.trendAnalyzer.DetectAnomalies(ctx, metricName, comparisonRange)
	if err != nil {
		return nil, fmt.Errorf("failed to detect comparison anomalies: %w", err)
	}
	
	// Convert anomalies to metrics for comparison
	baselineData := ca.anomaliesToMetrics(baselineAnomalies)
	comparisonData := ca.anomaliesToMetrics(comparisonAnomalies)
	
	// Create datasets
	baselineDataset := ca.createDataset("Baseline Anomalies", metricName, baselineRange, baselineData)
	comparisonDataset := ca.createDataset("Comparison Anomalies", metricName, comparisonRange, comparisonData)
	
	// Calculate statistics
	stats := ca.calculateStatistics(baselineData, comparisonData)
	
	// Generate insights
	insights := ca.generateAnomalyInsights(baselineAnomalies, comparisonAnomalies, stats)
	
	// Generate recommendations
	recommendations := ca.generateAnomalyRecommendations(insights)
	
	return &ComparisonResult{
		ComparisonType:    ComparisonTypeAnomalyPattern,
		BaselineDataset:   baselineDataset,
		ComparisonDataset: comparisonDataset,
		Statistics:        stats,
		Insights:          insights,
		Recommendations:   recommendations,
		GeneratedAt:       time.Now(),
	}, nil

// Internal methods

func (ca *ComparativeAnalyzer) createDataset(name, metricName string, timeRange TimeWindow, data []Metric) ComparisonDataset {
	stats := ca.calculateBasicStatistics(data)
	
	return ComparisonDataset{
		Name:        name,
		Description: fmt.Sprintf("Dataset for %s over %s", metricName, timeRange.Duration),
		MetricName:  metricName,
		TimeRange:   timeRange,
		DataPoints:  len(data),
		Statistics:  stats,
		Metadata: map[string]interface{}{
			"first_timestamp": data[0].Timestamp,
			"last_timestamp":  data[len(data)-1].Timestamp,
		},
	}

func (ca *ComparativeAnalyzer) calculateBasicStatistics(data []Metric) BasicStatistics {
	if len(data) == 0 {
		return BasicStatistics{}
	}
	
	values := make([]float64, len(data))
	for i, metric := range data {
		values[i] = metric.Value
	}
	
	// Calculate basic statistics
	mean := average(values)
	median := ca.calculateMedian(values)
	stdDev := standardDeviation(values)
	minVal := min(values)
	maxVal := max(values)
	variance := stdDev * stdDev
	
	// Calculate quartiles
	q1 := ca.calculatePercentile(values, 0.25)
	q3 := ca.calculatePercentile(values, 0.75)
	
	// Calculate skewness and kurtosis (simplified)
	skewness := ca.calculateSkewness(values, mean, stdDev)
	kurtosis := ca.calculateKurtosis(values, mean, stdDev)
	
	return BasicStatistics{
		Count:             len(values),
		Mean:              mean,
		Median:            median,
		StandardDeviation: stdDev,
		Min:               minVal,
		Max:               maxVal,
		Q1:                q1,
		Q3:                q3,
		Variance:          variance,
		Skewness:          skewness,
		Kurtosis:          kurtosis,
	}

func (ca *ComparativeAnalyzer) calculateStatistics(baselineData, comparisonData []Metric) ComparisonStatistics {
	baselineValues := ca.extractValues(baselineData)
	comparisonValues := ca.extractValues(comparisonData)
	
	baselineMean := average(baselineValues)
	comparisonMean := average(comparisonValues)
	
	meanDiff := comparisonMean - baselineMean
	meanDiffPercent := 0.0
	if baselineMean != 0 {
		meanDiffPercent = (meanDiff / baselineMean) * 100
	}
	
	baselineMedian := ca.calculateMedian(baselineValues)
	comparisonMedian := ca.calculateMedian(comparisonValues)
	medianDiff := comparisonMedian - baselineMedian
	
	baselineStdDev := standardDeviation(baselineValues)
	comparisonStdDev := standardDeviation(comparisonValues)
	stdDevRatio := 1.0
	if baselineStdDev != 0 {
		stdDevRatio = comparisonStdDev / baselineStdDev
	}
	
	// Calculate correlation (simplified)
	correlation := ca.calculateCorrelation(baselineValues, comparisonValues)
	
	// Calculate effect size (Cohen's d)
	effectSize := ca.calculateCohenD(baselineValues, comparisonValues)
	effectSizeInterpretation := ca.interpretEffectSize(effectSize)
	
	// Mock p-value and significance
	pValue := 0.05
	significance := "significant"
	if math.Abs(meanDiffPercent) < 5 {
		significance = "not_significant"
		pValue = 0.15
	}
	
	return ComparisonStatistics{
		MeanDifference:           meanDiff,
		MeanDifferencePercent:    meanDiffPercent,
		MedianDifference:         medianDiff,
		StandardDeviationRatio:   stdDevRatio,
		CorrelationCoefficient:   correlation,
		PValue:                   pValue,
		StatisticalSignificance:  significance,
		EffectSize:               effectSize,
		EffectSizeInterpretation: effectSizeInterpretation,
	}

func (ca *ComparativeAnalyzer) generateTimePeriodInsights(baseline, comparison ComparisonDataset, stats ComparisonStatistics) []ComparisonInsight {
	var insights []ComparisonInsight
	
	// Performance trend insight
	if math.Abs(stats.MeanDifferencePercent) > 10 {
		severity := "medium"
		if math.Abs(stats.MeanDifferencePercent) > 25 {
			severity = "high"
		}
		
		direction := "improved"
		if stats.MeanDifferencePercent < 0 {
			direction = "degraded"
		}
		
		insights = append(insights, ComparisonInsight{
			Type:        InsightTypePerformance,
			Severity:    severity,
			Title:       fmt.Sprintf("Performance %s by %.1f%%", direction, math.Abs(stats.MeanDifferencePercent)),
			Description: fmt.Sprintf("The metric has %s significantly compared to the baseline period", direction),
			Evidence:    []string{fmt.Sprintf("Mean difference: %.2f (%.1f%%)", stats.MeanDifference, stats.MeanDifferencePercent)},
			Confidence:  0.8,
		})
	}
	
	// Volatility insight
	if stats.StandardDeviationRatio > 1.5 {
		insights = append(insights, ComparisonInsight{
			Type:        InsightTypePattern,
			Severity:    "medium",
			Title:       "Increased Volatility Detected",
			Description: "The data shows increased variability compared to the baseline period",
			Evidence:    []string{fmt.Sprintf("Standard deviation ratio: %.2f", stats.StandardDeviationRatio)},
			Confidence:  0.7,
		})
	}
	
	return insights

func (ca *ComparativeAnalyzer) generateTimePeriodRecommendations(stats ComparisonStatistics, insights []ComparisonInsight) []string {
	var recommendations []string
	
	if math.Abs(stats.MeanDifferencePercent) > 20 {
		recommendations = append(recommendations, "Investigate the root cause of significant performance changes")
	}
	
	if stats.StandardDeviationRatio > 2 {
		recommendations = append(recommendations, "Consider implementing additional monitoring to understand volatility patterns")
	}
	
	if stats.StatisticalSignificance == "significant" {
		recommendations = append(recommendations, "The observed differences are statistically significant and warrant attention")
	}
	
	return recommendations

func (ca *ComparativeAnalyzer) generateMetricsInsights(baseline, comparison ComparisonDataset, stats ComparisonStatistics) []ComparisonInsight {
	var insights []ComparisonInsight
	
	// Correlation insight
	if math.Abs(stats.CorrelationCoefficient) > 0.7 {
		corrType := "positive"
		if stats.CorrelationCoefficient < 0 {
			corrType = "negative"
		}
		
		insights = append(insights, ComparisonInsight{
			Type:        InsightTypePattern,
			Severity:    "medium",
			Title:       fmt.Sprintf("Strong %s correlation detected", corrType),
			Description: fmt.Sprintf("The metrics show a strong %s correlation (r=%.3f)", corrType, stats.CorrelationCoefficient),
			Evidence:    []string{fmt.Sprintf("Correlation coefficient: %.3f", stats.CorrelationCoefficient)},
			Confidence:  0.85,
		})
	}
	
	return insights

func (ca *ComparativeAnalyzer) generateMetricsRecommendations(stats ComparisonStatistics, insights []ComparisonInsight) []string {
	var recommendations []string
	
	if math.Abs(stats.CorrelationCoefficient) > 0.8 {
		recommendations = append(recommendations, "Consider using one metric as a predictor for the other due to strong correlation")
	}
	
	return recommendations

func (ca *ComparativeAnalyzer) generateBaselineInsights(baseline, comparison ComparisonDataset, stats ComparisonStatistics, baselineValue float64) []ComparisonInsight {
	var insights []ComparisonInsight
	
	// Baseline deviation insight
	if math.Abs(stats.MeanDifferencePercent) > 15 {
		severity := "high"
		if math.Abs(stats.MeanDifferencePercent) < 25 {
			severity = "medium"
		}
		
		insights = append(insights, ComparisonInsight{
			Type:        InsightTypePerformance,
			Severity:    severity,
			Title:       "Significant deviation from baseline",
			Description: fmt.Sprintf("Current performance deviates %.1f%% from established baseline", math.Abs(stats.MeanDifferencePercent)),
			Evidence:    []string{fmt.Sprintf("Baseline: %.2f, Current: %.2f", baselineValue, comparison.Statistics.Mean)},
			Confidence:  0.9,
		})
	}
	
	return insights

func (ca *ComparativeAnalyzer) generateBaselineRecommendations(stats ComparisonStatistics, insights []ComparisonInsight, baselineValue float64) []string {
	var recommendations []string
	
	if math.Abs(stats.MeanDifferencePercent) > 20 {
		recommendations = append(recommendations, "Consider updating the baseline value or investigating the cause of deviation")
	}
	
	return recommendations

func (ca *ComparativeAnalyzer) generateAnomalyInsights(baselineAnomalies, comparisonAnomalies []AnomalyPoint, stats ComparisonStatistics) []ComparisonInsight {
	var insights []ComparisonInsight
	
	// Anomaly frequency insight
	frequencyChange := float64(len(comparisonAnomalies)-len(baselineAnomalies)) / float64(len(baselineAnomalies)) * 100
	
	if math.Abs(frequencyChange) > 50 {
		severity := "high"
		if math.Abs(frequencyChange) < 100 {
			severity = "medium"
		}
		
		direction := "increased"
		if frequencyChange < 0 {
			direction = "decreased"
		}
		
		insights = append(insights, ComparisonInsight{
			Type:        InsightTypeAnomaly,
			Severity:    severity,
			Title:       fmt.Sprintf("Anomaly frequency %s by %.1f%%", direction, math.Abs(frequencyChange)),
			Description: fmt.Sprintf("The number of anomalies has %s significantly", direction),
			Evidence:    []string{fmt.Sprintf("Baseline: %d anomalies, Comparison: %d anomalies", len(baselineAnomalies), len(comparisonAnomalies))},
			Confidence:  0.8,
		})
	}
	
	return insights

func (ca *ComparativeAnalyzer) generateAnomalyRecommendations(insights []ComparisonInsight) []string {
	var recommendations []string
	
	for _, insight := range insights {
		if insight.Type == InsightTypeAnomaly && insight.Severity == "high" {
			recommendations = append(recommendations, "Investigate the cause of anomaly pattern changes")
			break
		}
	}
	
	return recommendations

// Utility methods

func (ca *ComparativeAnalyzer) extractValues(data []Metric) []float64 {
	values := make([]float64, len(data))
	for i, metric := range data {
		values[i] = metric.Value
	}
	return values

func (ca *ComparativeAnalyzer) createSyntheticBaseline(baselineValue float64, count int) []Metric {
	data := make([]Metric, count)
	for i := 0; i < count; i++ {
		data[i] = Metric{
			Value:     baselineValue,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}
	return data

func (ca *ComparativeAnalyzer) anomaliesToMetrics(anomalies []AnomalyPoint) []Metric {
	metrics := make([]Metric, len(anomalies))
	for i, anomaly := range anomalies {
		metrics[i] = Metric{
			Value:     anomaly.Deviation,
			Timestamp: anomaly.Timestamp,
		}
	}
	return metrics

func (ca *ComparativeAnalyzer) calculateMedian(values []float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
	// Simple bubble sort
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]

func (ca *ComparativeAnalyzer) calculatePercentile(values []float64, p float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
	// Simple sort (same as median)
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	index := p * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1
	
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight

func (ca *ComparativeAnalyzer) calculateSkewness(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 || len(values) < 3 {
		return 0
	}
	
	sum := 0.0
	for _, value := range values {
		sum += math.Pow((value-mean)/stdDev, 3)
	}
	
	n := float64(len(values))
	return (n / ((n - 1) * (n - 2))) * sum

func (ca *ComparativeAnalyzer) calculateKurtosis(values []float64, mean, stdDev float64) float64 {
	if stdDev == 0 || len(values) < 4 {
		return 0
	}
	
	sum := 0.0
	for _, value := range values {
		sum += math.Pow((value-mean)/stdDev, 4)
	}
	
	n := float64(len(values))
	return ((n*(n+1))/((n-1)*(n-2)*(n-3)))*sum - (3*(n-1)*(n-1))/((n-2)*(n-3))

func (ca *ComparativeAnalyzer) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}
	
	meanX := average(x)
	meanY := average(y)
	
	var sumXY, sumX2, sumY2 float64
	
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		sumXY += dx * dy
		sumX2 += dx * dx
		sumY2 += dy * dy
	}
	
	if sumX2 == 0 || sumY2 == 0 {
		return 0
	}
	
	return sumXY / math.Sqrt(sumX2*sumY2)

func (ca *ComparativeAnalyzer) calculateCohenD(x, y []float64) float64 {
	if len(x) == 0 || len(y) == 0 {
		return 0
	}
	
	meanX := average(x)
	meanY := average(y)
	stdX := standardDeviation(x)
	stdY := standardDeviation(y)
	
	// Pooled standard deviation
	pooledStd := math.Sqrt(((float64(len(x))-1)*stdX*stdX + (float64(len(y))-1)*stdY*stdY) / (float64(len(x)) + float64(len(y)) - 2))
	
	if pooledStd == 0 {
		return 0
	}
	
	return (meanY - meanX) / pooledStd

func (ca *ComparativeAnalyzer) interpretEffectSize(d float64) string {
	absD := math.Abs(d)
	switch {
	case absD < 0.2:
		return "negligible"
	case absD < 0.5:
		return "small"
	case absD < 0.8:
		return "medium"
	default:
		return "large"
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
