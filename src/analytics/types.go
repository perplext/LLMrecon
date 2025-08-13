package analytics

import (
	"database/sql"
)

// TimeRange represents a time period for analytics queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ScanResult represents a scan result for analytics processing
type ScanResult struct {
	ID               string            `json:"id"`
	Timestamp        time.Time         `json:"timestamp"`
	Duration         time.Duration     `json:"duration"`
	Target           string            `json:"target"`
	TemplatesUsed    []string          `json:"templates_used"`
	TotalTests       int               `json:"total_tests"`
	PassedTests      int               `json:"passed_tests"`
	FailedTests      int               `json:"failed_tests"`
	Vulnerabilities  []Vulnerability   `json:"vulnerabilities"`
	Metadata         map[string]string `json:"metadata"`
	Success          bool              `json:"success"`
	ErrorMessage     string            `json:"error_message,omitempty"`
}

// Vulnerability represents a discovered vulnerability
type Vulnerability struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Severity    string            `json:"severity"`
	Category    string            `json:"category"`
	Template    string            `json:"template"`
	Description string            `json:"description"`
	Confidence  float64           `json:"confidence"`
	CVSS        float64           `json:"cvss,omitempty"`
	CWE         string            `json:"cwe,omitempty"`
	OWASP       string            `json:"owasp,omitempty"`
	Evidence    map[string]string `json:"evidence"`
	Remediation string            `json:"remediation,omitempty"`
}

// Metric represents a custom metric data point
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Tags      map[string]string `json:"tags"`
	Timestamp time.Time         `json:"timestamp"`
	Source    string            `json:"source"`
}

// MetricsQuery represents a query for metrics data
type MetricsQuery struct {
	TimeRange  TimeRange         `json:"time_range"`
	Metrics    []string          `json:"metrics"`
	GroupBy    []string          `json:"group_by"`
	Filters    map[string]string `json:"filters"`
	Aggregation string           `json:"aggregation"` // sum, avg, count, max, min
	Interval   string            `json:"interval"`    // 1h, 1d, 1w, 1m
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// MetricsResult represents query results
type MetricsResult struct {
	Query     *MetricsQuery `json:"query"`
	Data      []DataPoint   `json:"data"`
	Total     int           `json:"total"`
	Cached    bool          `json:"cached"`
	QueryTime time.Duration `json:"query_time"`
}

// DataPoint represents a single data point in analytics
type DataPoint struct {
	Timestamp time.Time          `json:"timestamp"`
	Values    map[string]float64 `json:"values"`
	Tags      map[string]string  `json:"tags"`
}

// Dashboard represents dashboard data
type Dashboard struct {
	GeneratedAt time.Time      `json:"generated_at"`
	TimeRange   TimeRange      `json:"time_range"`
	Widgets     []Widget       `json:"widgets"`
	Summary     DashboardSummary `json:"summary"`
	Alerts      []Alert        `json:"alerts"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"` // chart, table, metric, alert
	Title    string      `json:"title"`
	Data     interface{} `json:"data"`
	Position Position    `json:"position"`
	Config   WidgetConfig `json:"config"`
}

// Position represents widget position on dashboard
type Position struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// WidgetConfig represents widget configuration
type WidgetConfig struct {
	ChartType    string            `json:"chart_type,omitempty"`    // line, bar, pie, area
	ColorScheme  string            `json:"color_scheme,omitempty"`  // default, danger, warning, success
	ShowLegend   bool              `json:"show_legend,omitempty"`
	ShowGrid     bool              `json:"show_grid,omitempty"`
	Aggregation  string            `json:"aggregation,omitempty"`
	ThresholdValues map[string]float64 `json:"threshold_values,omitempty"`
	RefreshInterval string           `json:"refresh_interval,omitempty"`
}

// DashboardSummary represents high-level dashboard metrics
type DashboardSummary struct {
	TotalScans           int     `json:"total_scans"`
	TotalVulnerabilities int     `json:"total_vulnerabilities"`
	AverageScanDuration  float64 `json:"average_scan_duration"`
	SuccessRate          float64 `json:"success_rate"`
	CriticalVulns        int     `json:"critical_vulns"`
	HighVulns            int     `json:"high_vulns"`
	TrendDirection       string  `json:"trend_direction"` // up, down, stable
	TrendPercentage      float64 `json:"trend_percentage"`
}

// Alert represents a dashboard alert
type Alert struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // threshold, anomaly, trend
	Severity    string            `json:"severity"` // info, warning, error, critical
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	Timestamp   time.Time         `json:"timestamp"`
	Acknowledged bool             `json:"acknowledged"`
	Tags        map[string]string `json:"tags"`
	Actions     []AlertAction     `json:"actions"`
}

// AlertAction represents an action for an alert
type AlertAction struct {
	Type  string `json:"type"`  // dismiss, acknowledge, escalate
	Label string `json:"label"`
	URL   string `json:"url,omitempty"`
}

// TrendParams represents parameters for trend analysis
type TrendParams struct {
	TimeRange    TimeRange `json:"time_range"`
	Metrics      []string  `json:"metrics"`
	Granularity  string    `json:"granularity"` // hour, day, week, month
	Smoothing    bool      `json:"smoothing"`
	ForecastDays int       `json:"forecast_days"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	Params      *TrendParams   `json:"params"`
	Trends      []Trend        `json:"trends"`
	Summary     TrendSummary   `json:"summary"`
	Forecast    []ForecastPoint `json:"forecast,omitempty"`
	Anomalies   []Anomaly      `json:"anomalies"`
	GeneratedAt time.Time      `json:"generated_at"`
}

// Trend represents a single metric trend
type Trend struct {
	Metric      string      `json:"metric"`
	Direction   string      `json:"direction"` // increasing, decreasing, stable
	Strength    float64     `json:"strength"`  // 0-1, how strong the trend is
	Change      float64     `json:"change"`    // percentage change
	DataPoints  []DataPoint `json:"data_points"`
	RSquared    float64     `json:"r_squared"` // correlation coefficient
	Slope       float64     `json:"slope"`
	Confidence  float64     `json:"confidence"`
}

// TrendSummary represents overall trend summary
type TrendSummary struct {
	OverallDirection string  `json:"overall_direction"`
	StrongestTrend   string  `json:"strongest_trend"`
	WeakestTrend     string  `json:"weakest_trend"`
	AverageChange    float64 `json:"average_change"`
	Volatility       float64 `json:"volatility"`
}

// ForecastPoint represents a forecasted data point
type ForecastPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	Confidence float64   `json:"confidence"`
	Lower      float64   `json:"lower_bound"`
	Upper      float64   `json:"upper_bound"`
}

// Anomaly represents an anomalous data point
type Anomaly struct {
	Timestamp   time.Time `json:"timestamp"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Deviation   float64   `json:"deviation"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}

// ReportParams represents parameters for report generation
type ReportParams struct {
	Type        string            `json:"type"` // summary, detailed, executive, compliance
	TimeRange   TimeRange         `json:"time_range"`
	Targets     []string          `json:"targets"`
	Templates   []string          `json:"templates"`
	Format      string            `json:"format"` // pdf, html, docx, md
	Filters     map[string]string `json:"filters"`
	Sections    []string          `json:"sections"`
	Template    string            `json:"template_name,omitempty"`
	Recipients  []string          `json:"recipients,omitempty"`
	Schedule    string            `json:"schedule,omitempty"` // for recurring reports
}

// Report represents a generated analytics report
type Report struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	GeneratedAt time.Time         `json:"generated_at"`
	GeneratedBy string            `json:"generated_by"`
	Params      *ReportParams     `json:"params"`
	Sections    []ReportSection   `json:"sections"`
	Metadata    map[string]string `json:"metadata"`
	Size        int64             `json:"size"`
	Format      string            `json:"format"`
	FilePath    string            `json:"file_path,omitempty"`
}

// ReportSection represents a section within a report
type ReportSection struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Type     string      `json:"type"` // text, chart, table, summary
	Content  interface{} `json:"content"`
	Order    int         `json:"order"`
	PageBreak bool       `json:"page_break"`
}

// ComparisonParams represents parameters for comparative analysis
type ComparisonParams struct {
	Type        ComparisonType `json:"type"`
	TimeRanges  []TimeRange    `json:"time_ranges,omitempty"`
	Targets     []string       `json:"targets,omitempty"`
	Templates   []string       `json:"templates,omitempty"`
	Metrics     []string       `json:"metrics"`
	Filters     map[string]string `json:"filters"`
}

// ComparisonType represents the type of comparison
type ComparisonType string

const (
	ComparisonTypeTimeRange ComparisonType = "time_range"
	ComparisonTypeTargets   ComparisonType = "targets"
	ComparisonTypeTemplates ComparisonType = "templates"
)

// ComparisonResult represents comparison analysis results
type ComparisonResult struct {
	ComparisonType ComparisonType `json:"comparison_type"`
	Comparisons    []Comparison   `json:"comparisons"`
	Summary        ComparisonSummary `json:"summary"`
	GeneratedAt    time.Time      `json:"generated_at"`
}

// Comparison represents a single comparison
type Comparison struct {
	Label   string             `json:"label"`
	Metrics map[string]float64 `json:"metrics"`
	Deltas  map[string]float64 `json:"deltas,omitempty"` // percentage changes
	Rank    int                `json:"rank,omitempty"`
}

// ComparisonSummary represents overall comparison summary
type ComparisonSummary struct {
	BestPerforming  string  `json:"best_performing"`
	WorstPerforming string  `json:"worst_performing"`
	AverageChange   float64 `json:"average_change"`
	MaxChange       float64 `json:"max_change"`
	MinChange       float64 `json:"min_change"`
}

// ExportParams represents parameters for data export
type ExportParams struct {
	Format    string            `json:"format"` // json, csv, excel, xml
	DataType  string            `json:"data_type"` // metrics, trends, reports
	TimeRange TimeRange         `json:"time_range"`
	Metrics   []string          `json:"metrics"`
	Filters   map[string]string `json:"filters"`
	Filename  string            `json:"filename,omitempty"`
}

// AnalyticsSummary represents a high-level analytics summary
type AnalyticsSummary struct {
	GeneratedAt           time.Time     `json:"generated_at"`
	TotalScans            int           `json:"total_scans"`
	TotalVulnerabilities  int           `json:"total_vulnerabilities"`
	AverageScanDuration   float64       `json:"average_scan_duration"`
	MedianScanDuration    float64       `json:"median_scan_duration"`
	SuccessRate           float64       `json:"success_rate"`
	TrendData             TrendSummary  `json:"trend_data"`
	StorageStats          StorageStats  `json:"storage_stats"`
	TopVulnerabilities    []VulnSummary `json:"top_vulnerabilities"`
	TopTargets            []TargetSummary `json:"top_targets"`
	RecentActivity        []ActivitySummary `json:"recent_activity"`
}

// StorageStats represents storage statistics
type StorageStats struct {
	TotalSize    int64 `json:"total_size"`
	TotalRecords int64 `json:"total_records"`
	CompressedSize int64 `json:"compressed_size,omitempty"`
	IndexSize    int64 `json:"index_size,omitempty"`
}

// VulnSummary represents a vulnerability summary
type VulnSummary struct {
	Type        string  `json:"type"`
	Count       int     `json:"count"`
	Severity    string  `json:"severity"`
	Percentage  float64 `json:"percentage"`
	Trend       string  `json:"trend"`
}

// TargetSummary represents a target summary
type TargetSummary struct {
	Target      string  `json:"target"`
	ScanCount   int     `json:"scan_count"`
	VulnCount   int     `json:"vuln_count"`
	LastScan    time.Time `json:"last_scan"`
	RiskScore   float64 `json:"risk_score"`
}

// ActivitySummary represents recent activity summary
type ActivitySummary struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // scan, vulnerability, alert
	Description string    `json:"description"`
	Target      string    `json:"target,omitempty"`
	Severity    string    `json:"severity,omitempty"`
}

// DataStorage interface defines storage operations
type DataStorage interface {
	// Lifecycle
	Initialize() error
	Close() error
	
	// Metrics operations
	StoreMetric(metric *Metric) error
	StoreScanResult(result *ScanResult) error
	QueryMetrics(query *MetricsQuery) (*MetricsResult, error)
	
	// Cleanup operations
	DeleteRawData(before time.Time) error
	ArchiveData(before time.Time) error
	
	// Statistics
	GetStorageSize() (int64, error)
	GetRecordCount() (int64, error)
	
	// Advanced queries
	GetAggregatedData(query *MetricsQuery) (*MetricsResult, error)
	GetTimeSeriesData(metric string, timeRange TimeRange) ([]DataPoint, error)
}

// Constants for metric names
const (
	MetricScanDuration      = "scan_duration"
	MetricVulnerabilityCount = "vulnerability_count"
	MetricTestCount         = "test_count"
	MetricSuccessRate       = "success_rate"
	MetricCriticalVulns     = "critical_vulnerabilities"
	MetricHighVulns         = "high_vulnerabilities"
	MetricMediumVulns       = "medium_vulnerabilities"
	MetricLowVulns          = "low_vulnerabilities"
	MetricTargetCount       = "target_count"
	MetricTemplateUsage     = "template_usage"
	MetricErrorRate         = "error_rate"
)

// Constants for severities
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Constants for trend directions
const (
	TrendDirectionUp     = "up"
	TrendDirectionDown   = "down"
	TrendDirectionStable = "stable"
)

// Constants for widget types
const (
	WidgetTypeChart  = "chart"
	WidgetTypeTable  = "table"
	WidgetTypeMetric = "metric"
	WidgetTypeAlert  = "alert"
	WidgetTypeGauge  = "gauge"
	WidgetTypeSparkline = "sparkline"
)

// Constants for chart types
const (
	ChartTypeLine = "line"
	ChartTypeBar  = "bar"
	ChartTypePie  = "pie"
	ChartTypeArea = "area"
	ChartTypeDoughnut = "doughnut"
	ChartTypeScatter = "scatter"
)

// Validation functions

// Validate validates a TimeRange
func (tr TimeRange) Validate() error {
	if tr.Start.IsZero() || tr.End.IsZero() {
		return fmt.Errorf("start and end times must be specified")
	}
	if tr.Start.After(tr.End) {
		return fmt.Errorf("start time must be before end time")
	}
	return nil
}

// Duration returns the duration of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Contains checks if a time is within the range
func (tr TimeRange) Contains(t time.Time) bool {
	return !t.Before(tr.Start) && !t.After(tr.End)
}

// Validate validates MetricsQuery
func (mq *MetricsQuery) Validate() error {
	if err := mq.TimeRange.Validate(); err != nil {
		return fmt.Errorf("invalid time range: %w", err)
	}
	
	if mq.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	
	if mq.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	
	validAggregations := []string{"sum", "avg", "count", "max", "min"}
	if mq.Aggregation != "" {
		found := false
		for _, valid := range validAggregations {
			if mq.Aggregation == valid {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid aggregation: %s", mq.Aggregation)
		}
	}
	
	return nil
}

import "fmt"