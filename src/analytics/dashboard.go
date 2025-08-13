package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// DashboardEngine manages interactive dashboard components
type DashboardEngine struct {
	config        *Config
	storage       DataStorage
	trendAnalyzer *TrendAnalyzer
	logger        Logger
	
	// Dashboard state
	dashboards    map[string]*Dashboard
	widgets       map[string]Widget
	layouts       map[string]*Layout
	themes        map[string]*Theme
	
	// Real-time updates
	subscribers   map[string][]DashboardSubscriber
	updateChannel chan DashboardUpdate
	
	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// Dashboard represents a complete dashboard configuration
type Dashboard struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Layout      *Layout                `json:"layout"`
	Widgets     []WidgetConfig         `json:"widgets"`
	Theme       string                 `json:"theme"`
	Filters     map[string]interface{} `json:"filters"`
	RefreshRate time.Duration          `json:"refresh_rate"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Tags        []string               `json:"tags"`
	IsPublic    bool                   `json:"is_public"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Layout defines dashboard layout configuration
type Layout struct {
	Type        LayoutType `json:"type"`
	Columns     int        `json:"columns"`
	Rows        int        `json:"rows"`
	GridSize    GridSize   `json:"grid_size"`
	Responsive  bool       `json:"responsive"`
	Breakpoints map[string]BreakpointConfig `json:"breakpoints"`
}

// LayoutType represents different layout types
type LayoutType string

const (
	LayoutTypeGrid       LayoutType = "grid"
	LayoutTypeFlexible   LayoutType = "flexible"
	LayoutTypeMasonry    LayoutType = "masonry"
	LayoutTypeDashboard  LayoutType = "dashboard"
)

// GridSize defines grid dimensions
type GridSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// BreakpointConfig defines responsive breakpoint settings
type BreakpointConfig struct {
	MinWidth int `json:"min_width"`
	Columns  int `json:"columns"`
	Margin   int `json:"margin"`
}

// WidgetConfig defines widget configuration within a dashboard
type WidgetConfig struct {
	ID       string                 `json:"id"`
	Type     WidgetType             `json:"type"`
	Title    string                 `json:"title"`
	Position WidgetPosition         `json:"position"`
	Size     WidgetSize             `json:"size"`
	Config   map[string]interface{} `json:"config"`
	DataSource DataSourceConfig     `json:"data_source"`
	Style    WidgetStyle            `json:"style"`
}

// Widget interface for all dashboard widgets
type Widget interface {
	GetType() WidgetType
	GetData(ctx context.Context, config DataSourceConfig) (interface{}, error)
	Render(data interface{}, style WidgetStyle) (string, error)
	Validate(config map[string]interface{}) error
	GetMetadata() WidgetMetadata
}

// WidgetType represents different widget types
type WidgetType string

const (
	WidgetTypeChart      WidgetType = "chart"
	WidgetTypeTable      WidgetType = "table"
	WidgetTypeMetric     WidgetType = "metric"
	WidgetTypeGauge      WidgetType = "gauge"
	WidgetTypeHeatmap    WidgetType = "heatmap"
	WidgetTypeTimeline   WidgetType = "timeline"
	WidgetTypeAlert      WidgetType = "alert"
	WidgetTypeProgress   WidgetType = "progress"
	WidgetTypeText       WidgetType = "text"
	WidgetTypeIframe     WidgetType = "iframe"
)

// WidgetPosition defines widget position in layout
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WidgetSize defines widget dimensions
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DataSourceConfig defines how widgets fetch data
type DataSourceConfig struct {
	Type       DataSourceType         `json:"type"`
	MetricName string                 `json:"metric_name"`
	TimeRange  TimeWindow             `json:"time_range"`
	Filters    map[string]interface{} `json:"filters"`
	Aggregation AggregationConfig     `json:"aggregation"`
	RefreshInterval time.Duration      `json:"refresh_interval"`
}

// DataSourceType represents different data source types
type DataSourceType string

const (
	DataSourceTypeMetrics     DataSourceType = "metrics"
	DataSourceTypeTrends      DataSourceType = "trends"
	DataSourceTypeAggregated  DataSourceType = "aggregated"
	DataSourceTypeRealTime    DataSourceType = "realtime"
	DataSourceTypeStatic      DataSourceType = "static"
)

// AggregationConfig defines data aggregation settings
type AggregationConfig struct {
	Function string        `json:"function"`
	Interval time.Duration `json:"interval"`
	GroupBy  []string      `json:"group_by"`
}

// WidgetStyle defines widget styling
type WidgetStyle struct {
	Colors     ColorScheme            `json:"colors"`
	Fonts      FontConfig             `json:"fonts"`
	Borders    BorderConfig           `json:"borders"`
	Background BackgroundConfig       `json:"background"`
	Custom     map[string]interface{} `json:"custom"`
}

// ColorScheme defines color configuration
type ColorScheme struct {
	Primary   string   `json:"primary"`
	Secondary string   `json:"secondary"`
	Success   string   `json:"success"`
	Warning   string   `json:"warning"`
	Error     string   `json:"error"`
	Info      string   `json:"info"`
	Palette   []string `json:"palette"`
}

// FontConfig defines font styling
type FontConfig struct {
	Family string `json:"family"`
	Size   int    `json:"size"`
	Weight string `json:"weight"`
	Color  string `json:"color"`
}

// BorderConfig defines border styling
type BorderConfig struct {
	Width  int    `json:"width"`
	Style  string `json:"style"`
	Color  string `json:"color"`
	Radius int    `json:"radius"`
}

// BackgroundConfig defines background styling
type BackgroundConfig struct {
	Color    string `json:"color"`
	Pattern  string `json:"pattern"`
	Opacity  float64 `json:"opacity"`
}

// WidgetMetadata contains widget metadata
type WidgetMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	MinSize     WidgetSize `json:"min_size"`
	MaxSize     WidgetSize `json:"max_size"`
}

// Theme defines dashboard theming
type Theme struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Colors      ColorScheme            `json:"colors"`
	Fonts       FontConfig             `json:"fonts"`
	Layout      ThemeLayoutConfig      `json:"layout"`
	Components  map[string]interface{} `json:"components"`
}

// ThemeLayoutConfig defines theme layout settings
type ThemeLayoutConfig struct {
	Spacing    int    `json:"spacing"`
	Padding    int    `json:"padding"`
	Margin     int    `json:"margin"`
	GridGutter int    `json:"grid_gutter"`
}

// DashboardSubscriber interface for real-time updates
type DashboardSubscriber interface {
	OnDashboardUpdate(update DashboardUpdate) error
	GetSubscriptionID() string
}

// DashboardUpdate represents real-time dashboard updates
type DashboardUpdate struct {
	DashboardID string                 `json:"dashboard_id"`
	WidgetID    string                 `json:"widget_id"`
	UpdateType  UpdateType             `json:"update_type"`
	Data        interface{}            `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateType represents different update types
type UpdateType string

const (
	UpdateTypeData      UpdateType = "data"
	UpdateTypeConfig    UpdateType = "config"
	UpdateTypeStyle     UpdateType = "style"
	UpdateTypeLayout    UpdateType = "layout"
	UpdateTypeWidget    UpdateType = "widget"
)

// NewDashboardEngine creates a new dashboard engine
func NewDashboardEngine(config *Config, storage DataStorage, trendAnalyzer *TrendAnalyzer, logger Logger) *DashboardEngine {
	ctx, cancel := context.WithCancel(context.Background())
	
	engine := &DashboardEngine{
		config:        config,
		storage:       storage,
		trendAnalyzer: trendAnalyzer,
		logger:        logger,
		dashboards:    make(map[string]*Dashboard),
		widgets:       make(map[string]Widget),
		layouts:       make(map[string]*Layout),
		themes:        make(map[string]*Theme),
		subscribers:   make(map[string][]DashboardSubscriber),
		updateChannel: make(chan DashboardUpdate, 1000),
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// Register default widgets
	engine.registerDefaultWidgets()
	
	// Register default themes
	engine.registerDefaultThemes()
	
	// Register default layouts
	engine.registerDefaultLayouts()
	
	// Start update processor
	go engine.processUpdates()
	
	return engine
}

// CreateDashboard creates a new dashboard
func (de *DashboardEngine) CreateDashboard(name, description, createdBy string) (*Dashboard, error) {
	dashboard := &Dashboard{
		ID:          generateDashboardID(),
		Name:        name,
		Description: description,
		Layout:      de.layouts["default"],
		Widgets:     make([]WidgetConfig, 0),
		Theme:       "default",
		Filters:     make(map[string]interface{}),
		RefreshRate: 30 * time.Second,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Tags:        make([]string, 0),
		IsPublic:    false,
		Metadata:    make(map[string]interface{}),
	}
	
	de.dashboards[dashboard.ID] = dashboard
	de.logger.Info("Created dashboard", "id", dashboard.ID, "name", name)
	
	return dashboard, nil
}

// AddWidget adds a widget to a dashboard
func (de *DashboardEngine) AddWidget(dashboardID string, widgetConfig WidgetConfig) error {
	dashboard, exists := de.dashboards[dashboardID]
	if !exists {
		return fmt.Errorf("dashboard not found: %s", dashboardID)
	}
	
	// Validate widget configuration
	if widget, exists := de.widgets[string(widgetConfig.Type)]; exists {
		if err := widget.Validate(widgetConfig.Config); err != nil {
			return fmt.Errorf("invalid widget configuration: %w", err)
		}
	}
	
	// Generate widget ID if not provided
	if widgetConfig.ID == "" {
		widgetConfig.ID = generateWidgetID()
	}
	
	dashboard.Widgets = append(dashboard.Widgets, widgetConfig)
	dashboard.UpdatedAt = time.Now()
	
	de.logger.Info("Added widget to dashboard", "dashboardID", dashboardID, "widgetID", widgetConfig.ID, "type", widgetConfig.Type)
	
	// Send update notification
	update := DashboardUpdate{
		DashboardID: dashboardID,
		WidgetID:    widgetConfig.ID,
		UpdateType:  UpdateTypeWidget,
		Data:        widgetConfig,
		Timestamp:   time.Now(),
	}
	
	de.sendUpdate(update)
	
	return nil
}

// GetDashboard retrieves a dashboard by ID
func (de *DashboardEngine) GetDashboard(dashboardID string) (*Dashboard, error) {
	dashboard, exists := de.dashboards[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardID)
	}
	
	return dashboard, nil
}

// ListDashboards returns all dashboards (with optional filtering)
func (de *DashboardEngine) ListDashboards(filters map[string]interface{}) ([]*Dashboard, error) {
	var dashboards []*Dashboard
	
	for _, dashboard := range de.dashboards {
		if de.matchesFilters(dashboard, filters) {
			dashboards = append(dashboards, dashboard)
		}
	}
	
	return dashboards, nil
}

// RenderDashboard renders a complete dashboard
func (de *DashboardEngine) RenderDashboard(ctx context.Context, dashboardID string) (string, error) {
	dashboard, exists := de.dashboards[dashboardID]
	if !exists {
		return "", fmt.Errorf("dashboard not found: %s", dashboardID)
	}
	
	theme := de.themes[dashboard.Theme]
	if theme == nil {
		theme = de.themes["default"]
	}
	
	var renderedWidgets []string
	
	// Render each widget
	for _, widgetConfig := range dashboard.Widgets {
		rendered, err := de.renderWidget(ctx, widgetConfig, theme)
		if err != nil {
			de.logger.Error("Failed to render widget", "widgetID", widgetConfig.ID, "error", err)
			continue
		}
		renderedWidgets = append(renderedWidgets, rendered)
	}
	
	// Combine widgets into dashboard layout
	dashboardHTML := de.renderDashboardLayout(dashboard, renderedWidgets, theme)
	
	return dashboardHTML, nil
}

// GetWidgetData retrieves data for a specific widget
func (de *DashboardEngine) GetWidgetData(ctx context.Context, widgetConfig WidgetConfig) (interface{}, error) {
	widget, exists := de.widgets[string(widgetConfig.Type)]
	if !exists {
		return nil, fmt.Errorf("unknown widget type: %s", widgetConfig.Type)
	}
	
	return widget.GetData(ctx, widgetConfig.DataSource)
}

// UpdateWidget updates widget configuration
func (de *DashboardEngine) UpdateWidget(dashboardID, widgetID string, updates map[string]interface{}) error {
	dashboard, exists := de.dashboards[dashboardID]
	if !exists {
		return fmt.Errorf("dashboard not found: %s", dashboardID)
	}
	
	// Find and update widget
	for i, widget := range dashboard.Widgets {
		if widget.ID == widgetID {
			// Apply updates
			if title, ok := updates["title"].(string); ok {
				dashboard.Widgets[i].Title = title
			}
			if config, ok := updates["config"].(map[string]interface{}); ok {
				dashboard.Widgets[i].Config = config
			}
			if style, ok := updates["style"].(WidgetStyle); ok {
				dashboard.Widgets[i].Style = style
			}
			
			dashboard.UpdatedAt = time.Now()
			
			// Send update notification
			update := DashboardUpdate{
				DashboardID: dashboardID,
				WidgetID:    widgetID,
				UpdateType:  UpdateTypeConfig,
				Data:        dashboard.Widgets[i],
				Timestamp:   time.Now(),
			}
			
			de.sendUpdate(update)
			
			return nil
		}
	}
	
	return fmt.Errorf("widget not found: %s", widgetID)
}

// Subscribe adds a subscriber for dashboard updates
func (de *DashboardEngine) Subscribe(dashboardID string, subscriber DashboardSubscriber) {
	if de.subscribers[dashboardID] == nil {
		de.subscribers[dashboardID] = make([]DashboardSubscriber, 0)
	}
	
	de.subscribers[dashboardID] = append(de.subscribers[dashboardID], subscriber)
	de.logger.Debug("Added dashboard subscriber", "dashboardID", dashboardID, "subscriberID", subscriber.GetSubscriptionID())
}

// Unsubscribe removes a subscriber
func (de *DashboardEngine) Unsubscribe(dashboardID string, subscriberID string) {
	subscribers := de.subscribers[dashboardID]
	for i, subscriber := range subscribers {
		if subscriber.GetSubscriptionID() == subscriberID {
			// Remove subscriber
			de.subscribers[dashboardID] = append(subscribers[:i], subscribers[i+1:]...)
			de.logger.Debug("Removed dashboard subscriber", "dashboardID", dashboardID, "subscriberID", subscriberID)
			break
		}
	}
}

// RegisterWidget registers a custom widget type
func (de *DashboardEngine) RegisterWidget(widget Widget) {
	de.widgets[string(widget.GetType())] = widget
	de.logger.Info("Registered widget type", "type", widget.GetType())
}

// RegisterTheme registers a custom theme
func (de *DashboardEngine) RegisterTheme(theme *Theme) {
	de.themes[theme.ID] = theme
	de.logger.Info("Registered theme", "id", theme.ID, "name", theme.Name)
}

// Shutdown gracefully shuts down the dashboard engine
func (de *DashboardEngine) Shutdown() {
	de.cancel()
	close(de.updateChannel)
	de.logger.Info("Dashboard engine shut down")
}

// Internal methods

func (de *DashboardEngine) renderWidget(ctx context.Context, widgetConfig WidgetConfig, theme *Theme) (string, error) {
	widget, exists := de.widgets[string(widgetConfig.Type)]
	if !exists {
		return "", fmt.Errorf("unknown widget type: %s", widgetConfig.Type)
	}
	
	// Get widget data
	data, err := widget.GetData(ctx, widgetConfig.DataSource)
	if err != nil {
		return "", fmt.Errorf("failed to get widget data: %w", err)
	}
	
	// Apply theme styles to widget style
	style := de.applyThemeToWidgetStyle(widgetConfig.Style, theme)
	
	// Render widget
	return widget.Render(data, style)
}

func (de *DashboardEngine) renderDashboardLayout(dashboard *Dashboard, widgets []string, theme *Theme) string {
	// Simple grid layout implementation
	html := fmt.Sprintf(`
<div class="dashboard" id="dashboard-%s" style="background-color: %s;">
    <div class="dashboard-header">
        <h1 style="color: %s;">%s</h1>
        <p style="color: %s;">%s</p>
    </div>
    <div class="dashboard-content" style="display: grid; grid-template-columns: repeat(%d, 1fr); gap: %dpx;">
`,
		dashboard.ID,
		theme.Colors.Background,
		theme.Colors.Primary,
		dashboard.Name,
		theme.Colors.Secondary,
		dashboard.Description,
		dashboard.Layout.Columns,
		theme.Layout.GridGutter,
	)
	
	for _, widget := range widgets {
		html += fmt.Sprintf(`<div class="widget-container">%s</div>`, widget)
	}
	
	html += `
    </div>
</div>`
	
	return html
}

func (de *DashboardEngine) applyThemeToWidgetStyle(widgetStyle WidgetStyle, theme *Theme) WidgetStyle {
	// Apply theme colors if widget style colors are not set
	if widgetStyle.Colors.Primary == "" {
		widgetStyle.Colors.Primary = theme.Colors.Primary
	}
	if widgetStyle.Colors.Secondary == "" {
		widgetStyle.Colors.Secondary = theme.Colors.Secondary
	}
	
	// Apply theme fonts if widget fonts are not set
	if widgetStyle.Fonts.Family == "" {
		widgetStyle.Fonts.Family = theme.Fonts.Family
	}
	if widgetStyle.Fonts.Size == 0 {
		widgetStyle.Fonts.Size = theme.Fonts.Size
	}
	
	return widgetStyle
}

func (de *DashboardEngine) matchesFilters(dashboard *Dashboard, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}
	
	// Apply filters
	if createdBy, ok := filters["created_by"].(string); ok {
		if dashboard.CreatedBy != createdBy {
			return false
		}
	}
	
	if isPublic, ok := filters["is_public"].(bool); ok {
		if dashboard.IsPublic != isPublic {
			return false
		}
	}
	
	if tags, ok := filters["tags"].([]string); ok {
		for _, tag := range tags {
			found := false
			for _, dashboardTag := range dashboard.Tags {
				if dashboardTag == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	return true
}

func (de *DashboardEngine) sendUpdate(update DashboardUpdate) {
	select {
	case de.updateChannel <- update:
		// Update sent successfully
	default:
		de.logger.Warn("Update channel full, dropping update", "dashboardID", update.DashboardID)
	}
}

func (de *DashboardEngine) processUpdates() {
	for update := range de.updateChannel {
		subscribers := de.subscribers[update.DashboardID]
		
		for _, subscriber := range subscribers {
			go func(sub DashboardSubscriber) {
				if err := sub.OnDashboardUpdate(update); err != nil {
					de.logger.Error("Failed to send update to subscriber", 
						"subscriberID", sub.GetSubscriptionID(), 
						"error", err)
				}
			}(subscriber)
		}
	}
}

func (de *DashboardEngine) registerDefaultWidgets() {
	de.widgets[string(WidgetTypeChart)] = &ChartWidget{storage: de.storage, logger: de.logger}
	de.widgets[string(WidgetTypeTable)] = &TableWidget{storage: de.storage, logger: de.logger}
	de.widgets[string(WidgetTypeMetric)] = &MetricWidget{storage: de.storage, logger: de.logger}
	de.widgets[string(WidgetTypeGauge)] = &GaugeWidget{storage: de.storage, logger: de.logger}
	de.widgets[string(WidgetTypeHeatmap)] = &HeatmapWidget{storage: de.storage, trendAnalyzer: de.trendAnalyzer, logger: de.logger}
	de.widgets[string(WidgetTypeTimeline)] = &TimelineWidget{storage: de.storage, logger: de.logger}
	de.widgets[string(WidgetTypeAlert)] = &AlertWidget{storage: de.storage, logger: de.logger}
	de.widgets[string(WidgetTypeProgress)] = &ProgressWidget{storage: de.storage, logger: de.logger}
	de.widgets[string(WidgetTypeText)] = &TextWidget{logger: de.logger}
}

func (de *DashboardEngine) registerDefaultThemes() {
	// Default light theme
	de.themes["default"] = &Theme{
		ID:   "default",
		Name: "Default Light",
		Colors: ColorScheme{
			Primary:   "#007bff",
			Secondary: "#6c757d",
			Success:   "#28a745",
			Warning:   "#ffc107",
			Error:     "#dc3545",
			Info:      "#17a2b8",
			Palette:   []string{"#007bff", "#28a745", "#ffc107", "#dc3545", "#17a2b8"},
			Background: "#ffffff",
		},
		Fonts: FontConfig{
			Family: "Arial, sans-serif",
			Size:   14,
			Weight: "normal",
			Color:  "#333333",
		},
		Layout: ThemeLayoutConfig{
			Spacing:    16,
			Padding:    12,
			Margin:     8,
			GridGutter: 16,
		},
	}
	
	// Dark theme
	de.themes["dark"] = &Theme{
		ID:   "dark",
		Name: "Dark",
		Colors: ColorScheme{
			Primary:   "#0d6efd",
			Secondary: "#6c757d",
			Success:   "#198754",
			Warning:   "#ffc107",
			Error:     "#dc3545",
			Info:      "#0dcaf0",
			Palette:   []string{"#0d6efd", "#198754", "#ffc107", "#dc3545", "#0dcaf0"},
			Background: "#212529",
		},
		Fonts: FontConfig{
			Family: "Arial, sans-serif",
			Size:   14,
			Weight: "normal",
			Color:  "#ffffff",
		},
		Layout: ThemeLayoutConfig{
			Spacing:    16,
			Padding:    12,
			Margin:     8,
			GridGutter: 16,
		},
	}
}

func (de *DashboardEngine) registerDefaultLayouts() {
	de.layouts["default"] = &Layout{
		Type:       LayoutTypeGrid,
		Columns:    3,
		Rows:       3,
		GridSize:   GridSize{Width: 1200, Height: 800},
		Responsive: true,
		Breakpoints: map[string]BreakpointConfig{
			"mobile":  {MinWidth: 0, Columns: 1, Margin: 8},
			"tablet":  {MinWidth: 768, Columns: 2, Margin: 12},
			"desktop": {MinWidth: 1024, Columns: 3, Margin: 16},
		},
	}
	
	de.layouts["flexible"] = &Layout{
		Type:       LayoutTypeFlexible,
		Columns:    4,
		Rows:       4,
		GridSize:   GridSize{Width: 1400, Height: 1000},
		Responsive: true,
		Breakpoints: map[string]BreakpointConfig{
			"mobile":  {MinWidth: 0, Columns: 1, Margin: 8},
			"tablet":  {MinWidth: 768, Columns: 2, Margin: 12},
			"desktop": {MinWidth: 1024, Columns: 4, Margin: 16},
		},
	}
}

// Utility functions

func generateDashboardID() string {
	return fmt.Sprintf("dashboard_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateWidgetID() string {
	return fmt.Sprintf("widget_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}