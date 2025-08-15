package dashboard

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
)

// ExecutiveDashboard provides high-level security insights for LLM red teaming
type ExecutiveDashboard struct {
    mu              sync.RWMutex
    metrics         *MetricsSystem
    visualizer      *DataVisualizer
    alerts          *AlertSystem
    reporting       *ExecutiveReporting
    insights        *InsightsEngine
    forecaster      *TrendForecaster
    scorecard       *SecurityScorecard
    riskAnalyzer    *RiskAnalyzer
    widgets         map[string]*Widget
    layouts         map[string]*Layout
    config          DashboardConfig

}
// DashboardConfig holds configuration for executive dashboard
type DashboardConfig struct {
    RefreshInterval    time.Duration
    DataRetention      time.Duration
    AlertThresholds    map[string]float64
    EnableForecasting  bool
    EnableRealtime     bool
    MaxWidgets         int

}
// MetricsSystem manages dashboard metrics
type MetricsSystem struct {
    collectors  map[string]*MetricCollector
    aggregators map[string]*MetricAggregator
    storage     *MetricStorage
    mu          sync.RWMutex

}
// MetricCollector collects specific metrics
type MetricCollector struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        MetricType             `json:"type"`
    Source      string                 `json:"source"`
    Interval    time.Duration          `json:"interval"`
    LastCollect time.Time             `json:"last_collect"`
    Status      CollectorStatus        `json:"status"`
    Config      map[string]interface{} `json:"config"`

}
// MetricType defines metric types
type MetricType string

const (
    MetricCounter     MetricType = "counter"
    MetricGauge       MetricType = "gauge"
    MetricHistogram   MetricType = "histogram"
    MetricSummary     MetricType = "summary"
)

// CollectorStatus defines collector status
type CollectorStatus string

const (
    CollectorActive   CollectorStatus = "active"
    CollectorPaused   CollectorStatus = "paused"
    CollectorError    CollectorStatus = "error"
)

// MetricAggregator aggregates metrics
type MetricAggregator struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Metrics     []string            `json:"metrics"`
    Function    AggregationFunction `json:"function"`
    Window      time.Duration       `json:"window"`
    Output      string              `json:"output"`

}
// AggregationFunction defines aggregation functions
type AggregationFunction string

const (
    AggSum      AggregationFunction = "sum"
    AggAvg      AggregationFunction = "average"
    AggMin      AggregationFunction = "min"
    AggMax      AggregationFunction = "max"
    AggCount    AggregationFunction = "count"
    AggP95      AggregationFunction = "p95"
    AggP99      AggregationFunction = "p99"
)

// MetricStorage stores metric data
type MetricStorage struct {
    timeseries map[string]*TimeSeries
    mu         sync.RWMutex

}
// TimeSeries represents time series data
type TimeSeries struct {
    ID         string       `json:"id"`
    Name       string       `json:"name"`
    Points     []DataPoint  `json:"points"`
    Resolution time.Duration `json:"resolution"`
    Retention  time.Duration `json:"retention"`
}

}
// DataPoint represents a data point
type DataPoint struct {
    Timestamp time.Time              `json:"timestamp"`
    Value     float64                `json:"value"`
    Labels    map[string]string      `json:"labels"`
    Metadata  map[string]interface{} `json:"metadata"`

}
// DataVisualizer creates visualizations
type DataVisualizer struct {
    charts     map[string]*Chart
    renderers  map[ChartType]Renderer
    mu         sync.RWMutex
}

}
// Chart represents a data visualization
type Chart struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Type        ChartType              `json:"type"`
    DataSource  string                 `json:"data_source"`
    Options     ChartOptions           `json:"options"`
    LastUpdate  time.Time             `json:"last_update"`
    Data        interface{}            `json:"data"`
}

}
// ChartType defines chart types
type ChartType string

const (
    ChartLine       ChartType = "line"
    ChartBar        ChartType = "bar"
    ChartPie        ChartType = "pie"
    ChartScatter    ChartType = "scatter"
    ChartHeatmap    ChartType = "heatmap"
    ChartGauge      ChartType = "gauge"
    ChartTreemap    ChartType = "treemap"
    ChartSankey     ChartType = "sankey"
)

// ChartOptions contains chart configuration
type ChartOptions struct {
    Width       int                    `json:"width"`
    Height      int                    `json:"height"`
    Colors      []string               `json:"colors"`
    Legend      bool                   `json:"legend"`
    Interactive bool                   `json:"interactive"`
    Annotations []Annotation           `json:"annotations"`
    Custom      map[string]interface{} `json:"custom"`
}

}
// Annotation represents chart annotation
type Annotation struct {
    Type     string      `json:"type"`
    Position interface{} `json:"position"`
    Text     string      `json:"text"`
    Style    string      `json:"style"`

}
// Renderer interface for chart rendering
type Renderer interface {
    Render(chart *Chart) ([]byte, error)

// AlertSystem manages dashboard alerts
}
type AlertSystem struct {
    alerts      map[string]*Alert
    rules       map[string]*AlertRule
    channels    map[string]*AlertChannel
    mu          sync.RWMutex

}
// Alert represents a dashboard alert
type Alert struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Message     string                 `json:"message"`
    Severity    AlertSeverity          `json:"severity"`
    Source      string                 `json:"source"`
    TriggeredAt time.Time             `json:"triggered_at"`
    ResolvedAt  *time.Time            `json:"resolved_at,omitempty"`
    Status      AlertStatus            `json:"status"`
    Actions     []AlertAction          `json:"actions"`
    Metadata    map[string]interface{} `json:"metadata"`

}
// AlertSeverity defines alert severity levels
type AlertSeverity string

const (
    AlertCritical AlertSeverity = "critical"
    AlertHigh     AlertSeverity = "high"
    AlertMedium   AlertSeverity = "medium"
    AlertLow      AlertSeverity = "low"
    AlertInfo     AlertSeverity = "info"
)

// AlertStatus defines alert status
type AlertStatus string

const (
    AlertActive       AlertStatus = "active"
    AlertAcknowledged AlertStatus = "acknowledged"
    AlertResolved     AlertStatus = "resolved"
    AlertSuppressed   AlertStatus = "suppressed"
)

// AlertAction represents an alert action
type AlertAction struct {
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    Executed    bool                   `json:"executed"`
    Result      string                 `json:"result,omitempty"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// AlertRule defines alert triggering rules
type AlertRule struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Condition   string                 `json:"condition"`
    Severity    AlertSeverity          `json:"severity"`
    Actions     []string               `json:"actions"`
    Cooldown    time.Duration          `json:"cooldown"`
    Enabled     bool                   `json:"enabled"`
    LastFired   *time.Time            `json:"last_fired,omitempty"`
    Metadata    map[string]interface{} `json:"metadata"`

}
// AlertChannel represents an alert notification channel
type AlertChannel struct {
    ID       string      `json:"id"`
    Name     string      `json:"name"`
    Type     ChannelType `json:"type"`
    Config   interface{} `json:"config"`
    Enabled  bool        `json:"enabled"`
}

}
// ChannelType defines alert channel types
type ChannelType string

const (
    ChannelEmail    ChannelType = "email"
    ChannelSlack    ChannelType = "slack"
    ChannelWebhook  ChannelType = "webhook"
    ChannelSMS      ChannelType = "sms"
    ChannelPagerDuty ChannelType = "pagerduty"
)

// ExecutiveReporting generates executive reports
type ExecutiveReporting struct {
    reports     map[string]*ExecutiveReport
    templates   map[string]*ReportTemplate
    scheduler   *ReportScheduler
    mu          sync.RWMutex

}
// ExecutiveReport represents an executive report
type ExecutiveReport struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Period      ReportPeriod           `json:"period"`
    Summary     ExecutiveSummary       `json:"summary"`
    Sections    []ReportSection        `json:"sections"`
    Generated   time.Time             `json:"generated"`
    Recipients  []string               `json:"recipients"`
    Format      string                 `json:"format"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// ReportPeriod defines report period
type ReportPeriod struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
    Label string    `json:"label"`
}

}
// ExecutiveSummary contains executive summary
type ExecutiveSummary struct {
    KeyMetrics      map[string]float64     `json:"key_metrics"`
    Trends          []TrendSummary         `json:"trends"`
    Risks           []RiskSummary          `json:"risks"`
    Achievements    []string               `json:"achievements"`
    Concerns        []string               `json:"concerns"`
    Recommendations []string               `json:"recommendations"`

}
// TrendSummary summarizes a trend
type TrendSummary struct {
    Metric    string  `json:"metric"`
    Direction string  `json:"direction"`
    Change    float64 `json:"change"`
    Impact    string  `json:"impact"`

}
// RiskSummary summarizes a risk
type RiskSummary struct {
    Name        string  `json:"name"`
    Level       string  `json:"level"`
    Likelihood  float64 `json:"likelihood"`
    Impact      string  `json:"impact"`
    Mitigation  string  `json:"mitigation"`
}

}
// ReportSection represents a report section
type ReportSection struct {
    Title    string      `json:"title"`
    Content  interface{} `json:"content"`
    Charts   []string    `json:"charts"`
    Tables   []Table     `json:"tables"`
    Priority int         `json:"priority"`
}

}
// Table represents a data table
type Table struct {
    Headers []string   `json:"headers"`
    Rows    [][]string `json:"rows"`

}
// ReportTemplate defines report template
type ReportTemplate struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Sections    []string `json:"sections"`
    Schedule    string   `json:"schedule"`
    Recipients  []string `json:"recipients"`
}

}
// ReportScheduler schedules report generation
type ReportScheduler struct {
    schedules map[string]*Schedule
    mu        sync.RWMutex

}
// Schedule represents a report schedule
type Schedule struct {
    ID        string    `json:"id"`
    Frequency string    `json:"frequency"`
    NextRun   time.Time `json:"next_run"`
    Enabled   bool      `json:"enabled"`

}
// InsightsEngine generates insights
type InsightsEngine struct {
    analyzers map[string]*InsightAnalyzer
    insights  map[string]*Insight
    mu        sync.RWMutex
}

}
// InsightAnalyzer analyzes data for insights
type InsightAnalyzer struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Type        string   `json:"type"`
    Metrics     []string `json:"metrics"`
    Algorithm   string   `json:"algorithm"`
    Threshold   float64  `json:"threshold"`

}
// Insight represents a generated insight
type Insight struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    Type        InsightType            `json:"type"`
    Severity    InsightSeverity        `json:"severity"`
    Confidence  float64                `json:"confidence"`
    Evidence    []string               `json:"evidence"`
    Actions     []string               `json:"actions"`
    Generated   time.Time             `json:"generated"`
    ExpiresAt   time.Time             `json:"expires_at"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// InsightType defines insight types
type InsightType string

const (
    InsightAnomaly      InsightType = "anomaly"
    InsightTrend        InsightType = "trend"
    InsightPrediction   InsightType = "prediction"
    InsightOptimization InsightType = "optimization"
    InsightCompliance   InsightType = "compliance"
)

// InsightSeverity defines insight severity
type InsightSeverity string

const (
    InsightCritical InsightSeverity = "critical"
    InsightHigh     InsightSeverity = "high"
    InsightMedium   InsightSeverity = "medium"
    InsightLow      InsightSeverity = "low"
)

// TrendForecaster forecasts trends
type TrendForecaster struct {
    models      map[string]*ForecastModel
    forecasts   map[string]*Forecast
    mu          sync.RWMutex
}

}
// ForecastModel represents a forecasting model
type ForecastModel struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Parameters  map[string]interface{} `json:"parameters"`
    Accuracy    float64                `json:"accuracy"`
    LastTrained time.Time             `json:"last_trained"`

}
// Forecast represents a trend forecast
type Forecast struct {
    ID          string          `json:"id"`
    Metric      string          `json:"metric"`
    Horizon     time.Duration   `json:"horizon"`
    Points      []ForecastPoint `json:"points"`
    Confidence  float64         `json:"confidence"`
    Generated   time.Time       `json:"generated"`

}
// ForecastPoint represents a forecast point
type ForecastPoint struct {
    Timestamp  time.Time `json:"timestamp"`
    Value      float64   `json:"value"`
    Upper      float64   `json:"upper_bound"`
    Lower      float64   `json:"lower_bound"`
    Confidence float64   `json:"confidence"`
}

}
// SecurityScorecard tracks security metrics
type SecurityScorecard struct {
    scores      map[string]*ScoreMetric
    categories  map[string]*ScoreCategory
    history     map[string][]*ScoreHistory
    mu          sync.RWMutex

}
// ScoreMetric represents a score metric
type ScoreMetric struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Category    string    `json:"category"`
    Value       float64   `json:"value"`
    MaxValue    float64   `json:"max_value"`
    Weight      float64   `json:"weight"`
    Trend       string    `json:"trend"`
    LastUpdate  time.Time `json:"last_update"`

}
// ScoreCategory represents a score category
type ScoreCategory struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Metrics     []string `json:"metrics"`
    Weight      float64  `json:"weight"`
    Score       float64  `json:"score"`

}
// ScoreHistory represents score history
type ScoreHistory struct {
    Timestamp time.Time `json:"timestamp"`
    Score     float64   `json:"score"`
    Delta     float64   `json:"delta"`
}

}
// RiskAnalyzer analyzes security risks
type RiskAnalyzer struct {
    risks       map[string]*Risk
    scenarios   map[string]*RiskScenario
    mitigations map[string]*Mitigation
    mu          sync.RWMutex

}
// Risk represents a security risk
type Risk struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Category    string                 `json:"category"`
    Likelihood  float64                `json:"likelihood"`
    Impact      ImpactLevel            `json:"impact"`
    Score       float64                `json:"risk_score"`
    Status      RiskStatus             `json:"status"`
    Owner       string                 `json:"owner"`
    Mitigations []string               `json:"mitigations"`
    LastAssessed time.Time            `json:"last_assessed"`
    Metadata    map[string]interface{} `json:"metadata"`

}
// ImpactLevel defines impact levels
type ImpactLevel string

const (
    ImpactCritical ImpactLevel = "critical"
    ImpactHigh     ImpactLevel = "high"
    ImpactMedium   ImpactLevel = "medium"
    ImpactLow      ImpactLevel = "low"
    ImpactMinimal  ImpactLevel = "minimal"
)

// RiskStatus defines risk status
type RiskStatus string

const (
    RiskActive      RiskStatus = "active"
    RiskMitigated   RiskStatus = "mitigated"
    RiskAccepted    RiskStatus = "accepted"
    RiskTransferred RiskStatus = "transferred"
)

// RiskScenario represents a risk scenario
type RiskScenario struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Triggers    []string               `json:"triggers"`
    Outcomes    []string               `json:"outcomes"`
    Probability float64                `json:"probability"`
    SimResults  map[string]interface{} `json:"simulation_results"`
}

}
// Mitigation represents a risk mitigation
type Mitigation struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Type            MitigationType         `json:"type"`
    Effectiveness   float64                `json:"effectiveness"`
    Cost            float64                `json:"cost"`
    Implementation  string                 `json:"implementation"`
    Status          MitigationStatus       `json:"status"`
    Metadata        map[string]interface{} `json:"metadata"`
}

}
// MitigationType defines mitigation types
type MitigationType string

const (
    MitigationPreventive   MitigationType = "preventive"
    MitigationDetective    MitigationType = "detective"
    MitigationCorrective   MitigationType = "corrective"
    MitigationCompensating MitigationType = "compensating"
)

// MitigationStatus defines mitigation status
type MitigationStatus string

const (
    MitigationPlanned      MitigationStatus = "planned"
    MitigationImplementing MitigationStatus = "implementing"
    MitigationImplemented  MitigationStatus = "implemented"
    MitigationVerified     MitigationStatus = "verified"
)

// Widget represents a dashboard widget
type Widget struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        WidgetType             `json:"type"`
    DataSource  string                 `json:"data_source"`
    Config      WidgetConfig           `json:"config"`
    Position    Position               `json:"position"`
    Size        Size                   `json:"size"`
    LastUpdate  time.Time             `json:"last_update"`
    RefreshRate time.Duration          `json:"refresh_rate"`
    Metadata    map[string]interface{} `json:"metadata"`

}
// WidgetType defines widget types
type WidgetType string

const (
    WidgetMetric     WidgetType = "metric"
    WidgetChart      WidgetType = "chart"
    WidgetTable      WidgetType = "table"
    WidgetAlert      WidgetType = "alert"
    WidgetScorecard  WidgetType = "scorecard"
    WidgetHeatmap    WidgetType = "heatmap"
    WidgetTimeline   WidgetType = "timeline"
)

// WidgetConfig contains widget configuration
type WidgetConfig struct {
    Title       string                 `json:"title"`
    Subtitle    string                 `json:"subtitle"`
    ShowLegend  bool                   `json:"show_legend"`
    Interactive bool                   `json:"interactive"`
    Thresholds  []Threshold            `json:"thresholds"`
    Actions     []WidgetAction         `json:"actions"`
    Custom      map[string]interface{} `json:"custom"`
}

}
// Threshold represents a widget threshold
type Threshold struct {
    Value    float64 `json:"value"`
    Color    string  `json:"color"`
    Label    string  `json:"label"`
    Operator string  `json:"operator"`

}
// WidgetAction represents a widget action
type WidgetAction struct {
    Type    string `json:"type"`
    Label   string `json:"label"`
    Target  string `json:"target"`
    Payload string `json:"payload"`

}
// Position represents widget position
type Position struct {
    X int `json:"x"`
    Y int `json:"y"`

}
// Size represents widget size
type Size struct {
    Width  int `json:"width"`
    Height int `json:"height"`

}
// Layout represents dashboard layout
type Layout struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Widgets     []string               `json:"widgets"`
    Grid        GridConfig             `json:"grid"`
    Theme       string                 `json:"theme"`
    Metadata    map[string]interface{} `json:"metadata"`
}

}
// GridConfig defines grid configuration
type GridConfig struct {
    Columns    int     `json:"columns"`
    Rows       int     `json:"rows"`
    GutterSize int     `json:"gutter_size"`
    Responsive bool    `json:"responsive"`

}
// NewExecutiveDashboard creates a new executive dashboard
func NewExecutiveDashboard(config DashboardConfig) *ExecutiveDashboard {
    return &ExecutiveDashboard{
        metrics:      NewMetricsSystem(),
        visualizer:   NewDataVisualizer(),
        alerts:       NewAlertSystem(),
        reporting:    NewExecutiveReporting(),
        insights:     NewInsightsEngine(),
        forecaster:   NewTrendForecaster(),
        scorecard:    NewSecurityScorecard(),
        riskAnalyzer: NewRiskAnalyzer(),
        widgets:      make(map[string]*Widget),
        layouts:      make(map[string]*Layout),
        config:       config,
    }

// GetOverview returns dashboard overview
}
func (ed *ExecutiveDashboard) GetOverview(ctx context.Context) (*DashboardOverview, error) {
    ed.mu.RLock()
    defer ed.mu.RUnlock()

    overview := &DashboardOverview{
        Timestamp: time.Now(),
        Status:    ed.calculateOverallStatus(),
        Metrics:   ed.getKeyMetrics(),
        Alerts:    ed.getActiveAlerts(),
        Insights:  ed.getTopInsights(),
        Risks:     ed.getTopRisks(),
        Score:     ed.getSecurityScore(),
    }

    return overview, nil

// DashboardOverview contains dashboard overview
type DashboardOverview struct {
    Timestamp time.Time              `json:"timestamp"`
    Status    string                 `json:"status"`
    Metrics   map[string]interface{} `json:"metrics"`
    Alerts    []*Alert               `json:"alerts"`
    Insights  []*Insight             `json:"insights"`
    Risks     []*Risk                `json:"risks"`
    Score     float64                `json:"security_score"`
}

}
// AddWidget adds a widget to the dashboard
func (ed *ExecutiveDashboard) AddWidget(ctx context.Context, widget *Widget) error {
    ed.mu.Lock()
    defer ed.mu.Unlock()

    if len(ed.widgets) >= ed.config.MaxWidgets {
        return fmt.Errorf("maximum widgets limit reached")
    }

    if widget.ID == "" {
        widget.ID = generateWidgetID()
    }

    widget.LastUpdate = time.Now()
    ed.widgets[widget.ID] = widget

    // Start widget updates if realtime enabled
    if ed.config.EnableRealtime {
        go ed.updateWidget(ctx, widget)
    }

    return nil

// CreateLayout creates a dashboard layout
}
func (ed *ExecutiveDashboard) CreateLayout(ctx context.Context, layout *Layout) error {
    ed.mu.Lock()
    defer ed.mu.Unlock()

    if layout.ID == "" {
        layout.ID = generateLayoutID()
    }

    ed.layouts[layout.ID] = layout
    return nil

// NewMetricsSystem creates a new metrics system
}
func NewMetricsSystem() *MetricsSystem {
    return &MetricsSystem{
        collectors:  make(map[string]*MetricCollector),
        aggregators: make(map[string]*MetricAggregator),
        storage:     NewMetricStorage(),
    }

// CollectMetric collects a metric
}
func (ms *MetricsSystem) CollectMetric(name string, value float64, labels map[string]string) {
    ms.storage.Store(name, DataPoint{
        Timestamp: time.Now(),
        Value:     value,
        Labels:    labels,
    })

// NewMetricStorage creates new metric storage
}
func NewMetricStorage() *MetricStorage {
    return &MetricStorage{
        timeseries: make(map[string]*TimeSeries),
    }

// Store stores a data point
}
func (ms *MetricStorage) Store(name string, point DataPoint) {
    ms.mu.Lock()
    defer ms.mu.Unlock()

    ts, exists := ms.timeseries[name]
    if !exists {
        ts = &TimeSeries{
            ID:     generateTimeSeriesID(),
            Name:   name,
            Points: []DataPoint{},
        }
        ms.timeseries[name] = ts
    }

    ts.Points = append(ts.Points, point)

// NewDataVisualizer creates a new data visualizer
}
func NewDataVisualizer() *DataVisualizer {
    return &DataVisualizer{
        charts:    make(map[string]*Chart),
        renderers: make(map[ChartType]Renderer),
    }

// CreateChart creates a new chart
}
func (dv *DataVisualizer) CreateChart(chartType ChartType, title string, dataSource string) *Chart {
    dv.mu.Lock()
    defer dv.mu.Unlock()

    chart := &Chart{
        ID:         generateChartID(),
        Title:      title,
        Type:       chartType,
        DataSource: dataSource,
        LastUpdate: time.Now(),
        Options:    ChartOptions{},
    }

    dv.charts[chart.ID] = chart
    return chart

// NewAlertSystem creates a new alert system
}
func NewAlertSystem() *AlertSystem {
    return &AlertSystem{
        alerts:   make(map[string]*Alert),
        rules:    make(map[string]*AlertRule),
        channels: make(map[string]*AlertChannel),
    }

// TriggerAlert triggers a new alert
}
func (as *AlertSystem) TriggerAlert(rule *AlertRule, message string) *Alert {
    as.mu.Lock()
    defer as.mu.Unlock()

    alert := &Alert{
        ID:          generateAlertID(),
        Title:       rule.Name,
        Message:     message,
        Severity:    rule.Severity,
        Source:      rule.ID,
        TriggeredAt: time.Now(),
        Status:      AlertActive,
        Actions:     []AlertAction{},
    }

    as.alerts[alert.ID] = alert
    rule.LastFired = &alert.TriggeredAt

    // Execute alert actions
    for _, action := range rule.Actions {
        as.executeAction(alert, action)
    }

    return alert

// executeAction executes an alert action
}
func (as *AlertSystem) executeAction(alert *Alert, actionID string) {
    // Implement action execution
    action := AlertAction{
        Type:        actionID,
        Description: fmt.Sprintf("Execute %s for alert %s", actionID, alert.ID),
        Executed:    true,
    }
    alert.Actions = append(alert.Actions, action)

// NewExecutiveReporting creates new executive reporting
}
func NewExecutiveReporting() *ExecutiveReporting {
    return &ExecutiveReporting{
        reports:   make(map[string]*ExecutiveReport),
        templates: make(map[string]*ReportTemplate),
        scheduler: NewReportScheduler(),
    }

// GenerateReport generates an executive report
}
func (er *ExecutiveReporting) GenerateReport(period ReportPeriod) *ExecutiveReport {
    er.mu.Lock()
    defer er.mu.Unlock()

    report := &ExecutiveReport{
        ID:        generateReportID(),
        Title:     fmt.Sprintf("Executive Security Report - %s", period.Label),
        Period:    period,
        Generated: time.Now(),
        Summary:   er.generateSummary(period),
        Sections:  er.generateSections(period),
    }

    er.reports[report.ID] = report
    return report

// generateSummary generates report summary
}
func (er *ExecutiveReporting) generateSummary(period ReportPeriod) ExecutiveSummary {
    return ExecutiveSummary{
        KeyMetrics: map[string]float64{
            "security_score":     95.5,
            "incidents_resolved": 47,
            "mttr_hours":        2.5,
            "compliance_rate":   98.2,
        },
        Trends: []TrendSummary{
            {
                Metric:    "attack_success_rate",
                Direction: "decreasing",
                Change:    -15.3,
                Impact:    "positive",
            },
        },
        Risks: []RiskSummary{
            {
                Name:       "Advanced Persistent Threats",
                Level:      "high",
                Likelihood: 0.7,
                Impact:     "critical",
                Mitigation: "Enhanced monitoring deployed",
            },
        },
        Achievements:    []string{"Zero critical incidents", "100% uptime maintained"},
        Concerns:        []string{"Increasing sophistication of attacks"},
        Recommendations: []string{"Increase threat hunting resources"},
    }

// generateSections generates report sections
}
func (er *ExecutiveReporting) generateSections(period ReportPeriod) []ReportSection {
    return []ReportSection{
        {
            Title:    "Security Posture",
            Priority: 1,
        },
        {
            Title:    "Threat Landscape",
            Priority: 2,
        },
        {
            Title:    "Risk Assessment",
            Priority: 3,
        },
    }

// NewReportScheduler creates a new report scheduler
}
func NewReportScheduler() *ReportScheduler {
    return &ReportScheduler{
        schedules: make(map[string]*Schedule),
    }

// NewInsightsEngine creates a new insights engine
}
func NewInsightsEngine() *InsightsEngine {
    return &InsightsEngine{
        analyzers: make(map[string]*InsightAnalyzer),
        insights:  make(map[string]*Insight),
    }

// GenerateInsight generates a new insight
}
func (ie *InsightsEngine) GenerateInsight(analyzer *InsightAnalyzer, data interface{}) *Insight {
    ie.mu.Lock()
    defer ie.mu.Unlock()

    insight := &Insight{
        ID:          generateInsightID(),
        Title:       "Anomaly Detected in Attack Patterns",
        Description: "Unusual spike in prompt injection attempts detected",
        Type:        InsightAnomaly,
        Severity:    InsightHigh,
        Confidence:  0.85,
        Generated:   time.Now(),
        ExpiresAt:   time.Now().Add(24 * time.Hour),
    }

    ie.insights[insight.ID] = insight
    return insight

// NewTrendForecaster creates a new trend forecaster
}
func NewTrendForecaster() *TrendForecaster {
    return &TrendForecaster{
        models:    make(map[string]*ForecastModel),
        forecasts: make(map[string]*Forecast),
    }

// ForecastTrend forecasts a trend
}
func (tf *TrendForecaster) ForecastTrend(metric string, horizon time.Duration) *Forecast {
    tf.mu.Lock()
    defer tf.mu.Unlock()

    forecast := &Forecast{
        ID:         generateForecastID(),
        Metric:     metric,
        Horizon:    horizon,
        Points:     tf.generateForecastPoints(horizon),
        Confidence: 0.8,
        Generated:  time.Now(),
    }

    tf.forecasts[forecast.ID] = forecast
    return forecast

// generateForecastPoints generates forecast points
}
func (tf *TrendForecaster) generateForecastPoints(horizon time.Duration) []ForecastPoint {
    var points []ForecastPoint
    
    steps := int(horizon.Hours() / 24) // Daily points
    for i := 0; i < steps; i++ {
        point := ForecastPoint{
            Timestamp:  time.Now().Add(time.Duration(i) * 24 * time.Hour),
            Value:      100 + float64(i)*2,
            Upper:      110 + float64(i)*2,
            Lower:      90 + float64(i)*2,
            Confidence: 0.95 - float64(i)*0.01,
        }
        points = append(points, point)
    }
    
    return points

// NewSecurityScorecard creates a new security scorecard
}
func NewSecurityScorecard() *SecurityScorecard {
    return &SecurityScorecard{
        scores:     make(map[string]*ScoreMetric),
        categories: make(map[string]*ScoreCategory),
        history:    make(map[string][]*ScoreHistory),
    }

// UpdateScore updates a score metric
}
func (ss *SecurityScorecard) UpdateScore(metricID string, value float64) {
    ss.mu.Lock()
    defer ss.mu.Unlock()

    metric, exists := ss.scores[metricID]
    if !exists {
        metric = &ScoreMetric{
            ID:       metricID,
            MaxValue: 100,
        }
        ss.scores[metricID] = metric
    }

    oldValue := metric.Value
    metric.Value = value
    metric.LastUpdate = time.Now()

    if value > oldValue {
        metric.Trend = "improving"
    } else if value < oldValue {
        metric.Trend = "declining"
    } else {
        metric.Trend = "stable"
    }

    // Record history
    history := &ScoreHistory{
        Timestamp: time.Now(),
        Score:     value,
        Delta:     value - oldValue,
    }
    ss.history[metricID] = append(ss.history[metricID], history)

// NewRiskAnalyzer creates a new risk analyzer
}
func NewRiskAnalyzer() *RiskAnalyzer {
    return &RiskAnalyzer{
        risks:       make(map[string]*Risk),
        scenarios:   make(map[string]*RiskScenario),
        mitigations: make(map[string]*Mitigation),
    }

// AnalyzeRisk analyzes a security risk
}
func (ra *RiskAnalyzer) AnalyzeRisk(risk *Risk) float64 {
    ra.mu.Lock()
    defer ra.mu.Unlock()

    // Calculate risk score
    impactScore := map[ImpactLevel]float64{
        ImpactCritical: 5.0,
        ImpactHigh:     4.0,
        ImpactMedium:   3.0,
        ImpactLow:      2.0,
        ImpactMinimal:  1.0,
    }

    risk.Score = risk.Likelihood * impactScore[risk.Impact]
    risk.LastAssessed = time.Now()
    
    ra.risks[risk.ID] = risk
    
    return risk.Score

// Helper functions
}
func (ed *ExecutiveDashboard) calculateOverallStatus() string {
    // Calculate overall dashboard status
    activeAlerts := ed.alerts.getActiveCount()
    
    if activeAlerts > 10 {
        return "critical"
    } else if activeAlerts > 5 {
        return "warning"
    }
    return "healthy"

func (ed *ExecutiveDashboard) getKeyMetrics() map[string]interface{} {
    return map[string]interface{}{
        "total_attacks_blocked": 1523,
        "active_campaigns":      12,
        "security_score":        95.5,
        "mttr_minutes":         145,
    }

}
func (ed *ExecutiveDashboard) getActiveAlerts() []*Alert {
    return ed.alerts.getActive()

}
func (as *AlertSystem) getActiveCount() int {
    as.mu.RLock()
    defer as.mu.RUnlock()
    
    count := 0
    for _, alert := range as.alerts {
        if alert.Status == AlertActive {
            count++
        }
    }
    return count

func (as *AlertSystem) getActive() []*Alert {
    as.mu.RLock()
    defer as.mu.RUnlock()
    
    var active []*Alert
    for _, alert := range as.alerts {
        if alert.Status == AlertActive {
            active = append(active, alert)
        }
    }
    return active

func (ed *ExecutiveDashboard) getTopInsights() []*Insight {
    return ed.insights.getTop(5)

}
func (ie *InsightsEngine) getTop(limit int) []*Insight {
    ie.mu.RLock()
    defer ie.mu.RUnlock()
    
    var insights []*Insight
    count := 0
    for _, insight := range ie.insights {
        if count >= limit {
            break
        }
        insights = append(insights, insight)
        count++
    }
    return insights

func (ed *ExecutiveDashboard) getTopRisks() []*Risk {
    return ed.riskAnalyzer.getTopRisks(5)

}
func (ra *RiskAnalyzer) getTopRisks(limit int) []*Risk {
    ra.mu.RLock()
    defer ra.mu.RUnlock()
    
    var risks []*Risk
    count := 0
    for _, risk := range ra.risks {
        if count >= limit {
            break
        }
        if risk.Status == RiskActive {
            risks = append(risks, risk)
            count++
        }
    }
    return risks

func (ed *ExecutiveDashboard) getSecurityScore() float64 {
    return ed.scorecard.getOverallScore()

}
func (ss *SecurityScorecard) getOverallScore() float64 {
    ss.mu.RLock()
    defer ss.mu.RUnlock()
    
    if len(ss.scores) == 0 {
        return 0
    }
    
    totalScore := 0.0
    totalWeight := 0.0
    
    for _, metric := range ss.scores {
        totalScore += metric.Value * metric.Weight
        totalWeight += metric.Weight
    }
    
    if totalWeight == 0 {
        return 0
    }
    
    return totalScore / totalWeight

func (ed *ExecutiveDashboard) updateWidget(ctx context.Context, widget *Widget) {
    ticker := time.NewTicker(widget.RefreshRate)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Update widget data
            widget.LastUpdate = time.Now()
        case <-ctx.Done():
            return
        }
    }

func generateWidgetID() string {
    return fmt.Sprintf("widget_%d", time.Now().UnixNano())

}
func generateLayoutID() string {
    return fmt.Sprintf("layout_%d", time.Now().UnixNano())

}
func generateTimeSeriesID() string {
    return fmt.Sprintf("ts_%d", time.Now().UnixNano())

}
func generateChartID() string {
    return fmt.Sprintf("chart_%d", time.Now().UnixNano())

}
func generateAlertID() string {
    return fmt.Sprintf("alert_%d", time.Now().UnixNano())

}
func generateReportID() string {
    return fmt.Sprintf("report_%d", time.Now().UnixNano())

}
func generateInsightID() string {
    return fmt.Sprintf("insight_%d", time.Now().UnixNano())

}
func generateForecastID() string {
    return fmt.Sprintf("forecast_%d", time.Now().UnixNano())
