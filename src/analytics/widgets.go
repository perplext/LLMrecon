package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
)

// ChartWidget displays data in various chart formats
type ChartWidget struct {
	storage DataStorage
	logger  Logger
}

func (cw *ChartWidget) GetType() WidgetType {
	return WidgetTypeChart
}

func (cw *ChartWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	switch config.Type {
	case DataSourceTypeMetrics:
		return cw.getMetricsData(ctx, config)
	case DataSourceTypeTrends:
		return cw.getTrendsData(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", config.Type)
	}
}

func (cw *ChartWidget) getMetricsData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	metrics, err := cw.storage.GetMetricsByNameAndTimeRange(ctx, config.MetricName, config.TimeRange.Start, config.TimeRange.End)
	if err != nil {
		return nil, err
	}
	
	// Convert to chart data format
	chartData := ChartData{
		Labels: make([]string, len(metrics)),
		Datasets: []ChartDataset{
			{
				Label: config.MetricName,
				Data:  make([]float64, len(metrics)),
				BorderColor: "#007bff",
				BackgroundColor: "rgba(0, 123, 255, 0.1)",
			},
		},
	}
	
	for i, metric := range metrics {
		chartData.Labels[i] = metric.Timestamp.Format("15:04")
		chartData.Datasets[0].Data[i] = metric.Value
	}
	
	return chartData, nil
}

func (cw *ChartWidget) getTrendsData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	// This would integrate with TrendAnalyzer
	// For now, return mock trend data
	return map[string]interface{}{
		"trend": "upward",
		"slope": 0.5,
		"confidence": 0.85,
	}, nil
}

func (cw *ChartWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	chartData, ok := data.(ChartData)
	if !ok {
		return "", fmt.Errorf("invalid data type for chart widget")
	}
	
	// Generate Chart.js compatible HTML
	chartHTML := fmt.Sprintf(`
<div class="chart-widget" style="border: %dpx %s %s; border-radius: %dpx; background-color: %s; padding: %dpx;">
	<canvas id="chart-%d" width="400" height="200"></canvas>
	<script>
	const ctx%d = document.getElementById('chart-%d').getContext('2d');
	new Chart(ctx%d, {
		type: 'line',
		data: %s,
		options: {
			responsive: true,
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true
				}
			},
			plugins: {
				legend: {
					labels: {
						color: '%s'
					}
				}
			}
		}
	});
	</script>
</div>`,
		style.Borders.Width, style.Borders.Style, style.Borders.Color,
		style.Borders.Radius, style.Background.Color, 16,
		time.Now().UnixNano(),
		time.Now().UnixNano(), time.Now().UnixNano(),
		time.Now().UnixNano(),
		cw.serializeChartData(chartData),
		style.Fonts.Color,
	)
	
	return chartHTML, nil
}

func (cw *ChartWidget) Validate(config map[string]interface{}) error {
	if _, ok := config["chart_type"]; !ok {
		return fmt.Errorf("chart_type is required")
	}
	return nil
}

func (cw *ChartWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{
		Name:        "Chart Widget",
		Description: "Displays data in various chart formats (line, bar, pie, etc.)",
		Version:     "1.0.0",
		Author:      "LLMrecon",
		Tags:        []string{"chart", "visualization", "data"},
		MinSize:     WidgetSize{Width: 2, Height: 2},
		MaxSize:     WidgetSize{Width: 12, Height: 8},
	}
}

func (cw *ChartWidget) serializeChartData(data ChartData) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// TableWidget displays data in tabular format
type TableWidget struct {
	storage DataStorage
	logger  Logger
}

func (tw *TableWidget) GetType() WidgetType {
	return WidgetTypeTable
}

func (tw *TableWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	metrics, err := tw.storage.GetMetricsByNameAndTimeRange(ctx, config.MetricName, config.TimeRange.Start, config.TimeRange.End)
	if err != nil {
		return nil, err
	}
	
	// Convert to table data
	tableData := TableData{
		Headers: []string{"Timestamp", "Value", "Labels"},
		Rows:    make([][]interface{}, len(metrics)),
	}
	
	for i, metric := range metrics {
		labelsStr := ""
		for k, v := range metric.Labels {
			labelsStr += fmt.Sprintf("%s:%s ", k, v)
		}
		
		tableData.Rows[i] = []interface{}{
			metric.Timestamp.Format("2006-01-02 15:04:05"),
			metric.Value,
			labelsStr,
		}
	}
	
	return tableData, nil
}

func (tw *TableWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	tableData, ok := data.(TableData)
	if !ok {
		return "", fmt.Errorf("invalid data type for table widget")
	}
	
	html := fmt.Sprintf(`
<div class="table-widget" style="border: %dpx %s %s; border-radius: %dpx; background-color: %s; padding: 16px;">
	<table style="width: 100%%; border-collapse: collapse; color: %s; font-family: %s; font-size: %dpx;">
		<thead>
			<tr style="background-color: %s;">`,
		style.Borders.Width, style.Borders.Style, style.Borders.Color,
		style.Borders.Radius, style.Background.Color,
		style.Fonts.Color, style.Fonts.Family, style.Fonts.Size,
		style.Colors.Primary,
	)
	
	// Add headers
	for _, header := range tableData.Headers {
		html += fmt.Sprintf(`<th style="padding: 8px; border: 1px solid %s; color: white;">%s</th>`, 
			style.Borders.Color, html.EscapeString(header))
	}
	
	html += `</tr></thead><tbody>`
	
	// Add rows
	for i, row := range tableData.Rows {
		bgColor := "transparent"
		if i%2 == 0 {
			bgColor = "rgba(0,0,0,0.05)"
		}
		
		html += fmt.Sprintf(`<tr style="background-color: %s;">`, bgColor)
		for _, cell := range row {
			html += fmt.Sprintf(`<td style="padding: 8px; border: 1px solid %s;">%s</td>`, 
				style.Borders.Color, html.EscapeString(fmt.Sprintf("%v", cell)))
		}
		html += `</tr>`
	}
	
	html += `</tbody></table></div>`
	
	return html, nil
}

func (tw *TableWidget) Validate(config map[string]interface{}) error {
	return nil // Table widget has no required config
}

func (tw *TableWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{
		Name:        "Table Widget",
		Description: "Displays data in tabular format with sorting and filtering",
		Version:     "1.0.0",
		Author:      "LLMrecon",
		Tags:        []string{"table", "data", "list"},
		MinSize:     WidgetSize{Width: 2, Height: 2},
		MaxSize:     WidgetSize{Width: 12, Height: 8},
	}
}

// MetricWidget displays a single metric value
type MetricWidget struct {
	storage DataStorage
	logger  Logger
}

func (mw *MetricWidget) GetType() WidgetType {
	return WidgetTypeMetric
}

func (mw *MetricWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	metrics, err := mw.storage.GetMetricsByNameAndTimeRange(ctx, config.MetricName, config.TimeRange.Start, config.TimeRange.End)
	if err != nil {
		return nil, err
	}
	
	if len(metrics) == 0 {
		return MetricData{Value: 0, Label: config.MetricName}, nil
	}
	
	// Get the latest metric value
	latest := metrics[len(metrics)-1]
	
	// Calculate change from previous value
	var change float64
	var changePercent float64
	if len(metrics) > 1 {
		previous := metrics[len(metrics)-2]
		change = latest.Value - previous.Value
		if previous.Value != 0 {
			changePercent = (change / previous.Value) * 100
		}
	}
	
	return MetricData{
		Value:         latest.Value,
		Label:         config.MetricName,
		Change:        change,
		ChangePercent: changePercent,
		Timestamp:     latest.Timestamp,
	}, nil
}

func (mw *MetricWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	metricData, ok := data.(MetricData)
	if !ok {
		return "", fmt.Errorf("invalid data type for metric widget")
	}
	
	changeColor := style.Colors.Success
	changeIcon := "↑"
	if metricData.Change < 0 {
		changeColor = style.Colors.Error
		changeIcon = "↓"
	} else if metricData.Change == 0 {
		changeColor = style.Colors.Secondary
		changeIcon = "→"
	}
	
	html := fmt.Sprintf(`
<div class="metric-widget" style="border: %dpx %s %s; border-radius: %dpx; background-color: %s; padding: 24px; text-align: center;">
	<div style="color: %s; font-family: %s; font-size: %dpx; margin-bottom: 8px;">
		%s
	</div>
	<div style="color: %s; font-family: %s; font-size: 36px; font-weight: bold; margin-bottom: 8px;">
		%.2f
	</div>
	<div style="color: %s; font-family: %s; font-size: 14px;">
		<span style="color: %s;">%s %.2f%% (%.2f)</span>
	</div>
	<div style="color: %s; font-family: %s; font-size: 12px; margin-top: 8px;">
		Last updated: %s
	</div>
</div>`,
		style.Borders.Width, style.Borders.Style, style.Borders.Color,
		style.Borders.Radius, style.Background.Color,
		style.Colors.Secondary, style.Fonts.Family, style.Fonts.Size,
		html.EscapeString(metricData.Label),
		style.Fonts.Color, style.Fonts.Family,
		metricData.Value,
		style.Fonts.Color, style.Fonts.Family,
		changeColor, changeIcon, metricData.ChangePercent, metricData.Change,
		style.Colors.Secondary, style.Fonts.Family,
		metricData.Timestamp.Format("15:04:05"),
	)
	
	return html, nil
}

func (mw *MetricWidget) Validate(config map[string]interface{}) error {
	return nil
}

func (mw *MetricWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{
		Name:        "Metric Widget",
		Description: "Displays a single metric value with trend indicator",
		Version:     "1.0.0",
		Author:      "LLMrecon",
		Tags:        []string{"metric", "kpi", "value"},
		MinSize:     WidgetSize{Width: 1, Height: 1},
		MaxSize:     WidgetSize{Width: 3, Height: 2},
	}
}

// GaugeWidget displays metrics as a gauge/dial
type GaugeWidget struct {
	storage DataStorage
	logger  Logger
}

func (gw *GaugeWidget) GetType() WidgetType {
	return WidgetTypeGauge
}

func (gw *GaugeWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	metrics, err := gw.storage.GetMetricsByNameAndTimeRange(ctx, config.MetricName, config.TimeRange.Start, config.TimeRange.End)
	if err != nil {
		return nil, err
	}
	
	if len(metrics) == 0 {
		return GaugeData{Value: 0, Min: 0, Max: 100, Label: config.MetricName}, nil
	}
	
	latest := metrics[len(metrics)-1]
	
	// Calculate min/max from data
	min := latest.Value
	max := latest.Value
	for _, metric := range metrics {
		if metric.Value < min {
			min = metric.Value
		}
		if metric.Value > max {
			max = metric.Value
		}
	}
	
	// Add some padding to min/max
	padding := (max - min) * 0.1
	min -= padding
	max += padding
	
	return GaugeData{
		Value: latest.Value,
		Min:   min,
		Max:   max,
		Label: config.MetricName,
		Unit:  "units",
	}, nil
}

func (gw *GaugeWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	gaugeData, ok := data.(GaugeData)
	if !ok {
		return "", fmt.Errorf("invalid data type for gauge widget")
	}
	
	// Calculate percentage for gauge position
	percentage := ((gaugeData.Value - gaugeData.Min) / (gaugeData.Max - gaugeData.Min)) * 100
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	
	// Determine color based on percentage
	gaugeColor := style.Colors.Success
	if percentage > 80 {
		gaugeColor = style.Colors.Warning
	}
	if percentage > 90 {
		gaugeColor = style.Colors.Error
	}
	
	html := fmt.Sprintf(`
<div class="gauge-widget" style="border: %dpx %s %s; border-radius: %dpx; background-color: %s; padding: 24px; text-align: center;">
	<div style="color: %s; font-family: %s; font-size: %dpx; margin-bottom: 16px;">
		%s
	</div>
	<div class="gauge-container" style="position: relative; width: 150px; height: 150px; margin: 0 auto;">
		<svg width="150" height="150" viewBox="0 0 150 150">
			<circle cx="75" cy="75" r="60" fill="none" stroke="#e0e0e0" stroke-width="10"/>
			<circle cx="75" cy="75" r="60" fill="none" stroke="%s" stroke-width="10" 
					stroke-dasharray="%.2f 377" stroke-dashoffset="94.25" 
					style="transform: rotate(-90deg); transform-origin: 75px 75px;"/>
		</svg>
		<div style="position: absolute; top: 50%%; left: 50%%; transform: translate(-50%%, -50%%);">
			<div style="color: %s; font-family: %s; font-size: 24px; font-weight: bold;">
				%.1f
			</div>
			<div style="color: %s; font-family: %s; font-size: 12px;">
				%s
			</div>
		</div>
	</div>
	<div style="color: %s; font-family: %s; font-size: 12px; margin-top: 8px;">
		Range: %.1f - %.1f
	</div>
</div>`,
		style.Borders.Width, style.Borders.Style, style.Borders.Color,
		style.Borders.Radius, style.Background.Color,
		style.Colors.Secondary, style.Fonts.Family, style.Fonts.Size,
		html.EscapeString(gaugeData.Label),
		gaugeColor, percentage * 3.77, // 377 is circumference, so percentage * 3.77 gives the arc length
		style.Fonts.Color, style.Fonts.Family,
		gaugeData.Value,
		style.Colors.Secondary, style.Fonts.Family,
		gaugeData.Unit,
		style.Colors.Secondary, style.Fonts.Family,
		gaugeData.Min, gaugeData.Max,
	)
	
	return html, nil
}

func (gw *GaugeWidget) Validate(config map[string]interface{}) error {
	return nil
}

func (gw *GaugeWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{
		Name:        "Gauge Widget",
		Description: "Displays metrics as a circular gauge with min/max ranges",
		Version:     "1.0.0",
		Author:      "LLMrecon",
		Tags:        []string{"gauge", "dial", "range"},
		MinSize:     WidgetSize{Width: 2, Height: 2},
		MaxSize:     WidgetSize{Width: 4, Height: 4},
	}
}

// HeatmapWidget displays data as a heatmap
type HeatmapWidget struct {
	storage       DataStorage
	trendAnalyzer *TrendAnalyzer
	logger        Logger
}

func (hw *HeatmapWidget) GetType() WidgetType {
	return WidgetTypeHeatmap
}

func (hw *HeatmapWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	metrics, err := hw.storage.GetMetricsByNameAndTimeRange(ctx, config.MetricName, config.TimeRange.Start, config.TimeRange.End)
	if err != nil {
		return nil, err
	}
	
	// Group metrics by hour and day for heatmap
	heatmapData := make(map[string]map[int]float64)
	
	for _, metric := range metrics {
		day := metric.Timestamp.Format("2006-01-02")
		hour := metric.Timestamp.Hour()
		
		if heatmapData[day] == nil {
			heatmapData[day] = make(map[int]float64)
		}
		
		// Average values for the same hour
		if existing, exists := heatmapData[day][hour]; exists {
			heatmapData[day][hour] = (existing + metric.Value) / 2
		} else {
			heatmapData[day][hour] = metric.Value
		}
	}
	
	return heatmapData, nil
}

func (hw *HeatmapWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	heatmapData, ok := data.(map[string]map[int]float64)
	if !ok {
		return "", fmt.Errorf("invalid data type for heatmap widget")
	}
	
	// Get all days and sort them
	var days []string
	for day := range heatmapData {
		days = append(days, day)
	}
	sort.Strings(days)
	
	// Find min/max values for color scaling
	var minVal, maxVal float64
	first := true
	for _, dayData := range heatmapData {
		for _, value := range dayData {
			if first {
				minVal = value
				maxVal = value
				first = false
			} else {
				if value < minVal {
					minVal = value
				}
				if value > maxVal {
					maxVal = value
				}
			}
		}
	}
	
	html := fmt.Sprintf(`
<div class="heatmap-widget" style="border: %dpx %s %s; border-radius: %dpx; background-color: %s; padding: 16px;">
	<div style="color: %s; font-family: %s; font-size: %dpx; margin-bottom: 16px; text-align: center;">
		Activity Heatmap
	</div>
	<div class="heatmap-grid" style="display: grid; grid-template-columns: repeat(24, 1fr); gap: 2px;">`,
		style.Borders.Width, style.Borders.Style, style.Borders.Color,
		style.Borders.Radius, style.Background.Color,
		style.Colors.Primary, style.Fonts.Family, style.Fonts.Size,
	)
	
	// Add hour headers
	for hour := 0; hour < 24; hour++ {
		html += fmt.Sprintf(`<div style="font-size: 10px; text-align: center; color: %s;">%02d</div>`, 
			style.Colors.Secondary, hour)
	}
	
	// Add data cells
	for _, day := range days {
		dayData := heatmapData[day]
		for hour := 0; hour < 24; hour++ {
			value := dayData[hour]
			intensity := 0.0
			if maxVal > minVal {
				intensity = (value - minVal) / (maxVal - minVal)
			}
			
			// Create color based on intensity
			color := hw.getHeatmapColor(intensity, style.Colors.Primary)
			
			html += fmt.Sprintf(`
<div style="width: 20px; height: 20px; background-color: %s; border: 1px solid %s; 
     display: flex; align-items: center; justify-content: center; font-size: 8px; color: white;" 
     title="Day: %s, Hour: %02d, Value: %.2f">
</div>`,
				color, style.Borders.Color, day, hour, value)
		}
	}
	
	html += `</div></div>`
	
	return html, nil
}

func (hw *HeatmapWidget) getHeatmapColor(intensity float64, baseColor string) string {
	// Convert intensity to RGB
	red := int(255 * intensity)
	green := int(255 * (1 - intensity) * 0.5)
	blue := int(255 * (1 - intensity) * 0.5)
	
	if red > 255 {
		red = 255
	}
	if green > 255 {
		green = 255
	}
	if blue > 255 {
		blue = 255
	}
	
	return fmt.Sprintf("rgb(%d, %d, %d)", red, green, blue)
}

func (hw *HeatmapWidget) Validate(config map[string]interface{}) error {
	return nil
}

func (hw *HeatmapWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{
		Name:        "Heatmap Widget",
		Description: "Displays data intensity as a color-coded heatmap",
		Version:     "1.0.0",
		Author:      "LLMrecon",
		Tags:        []string{"heatmap", "intensity", "temporal"},
		MinSize:     WidgetSize{Width: 4, Height: 3},
		MaxSize:     WidgetSize{Width: 12, Height: 6},
	}
}

// Additional widget types with basic implementations

// TimelineWidget displays events on a timeline
type TimelineWidget struct {
	storage DataStorage
	logger  Logger
}

func (tw *TimelineWidget) GetType() WidgetType          { return WidgetTypeTimeline }
func (tw *TimelineWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	return []interface{}{}, nil // Mock implementation
}
func (tw *TimelineWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	return `<div class="timeline-widget">Timeline Widget</div>`, nil
}
func (tw *TimelineWidget) Validate(config map[string]interface{}) error { return nil }
func (tw *TimelineWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{Name: "Timeline Widget", Version: "1.0.0"}
}

// AlertWidget displays alerts and notifications
type AlertWidget struct {
	storage DataStorage
	logger  Logger
}

func (aw *AlertWidget) GetType() WidgetType          { return WidgetTypeAlert }
func (aw *AlertWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	return []interface{}{}, nil
}
func (aw *AlertWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	return `<div class="alert-widget">Alert Widget</div>`, nil
}
func (aw *AlertWidget) Validate(config map[string]interface{}) error { return nil }
func (aw *AlertWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{Name: "Alert Widget", Version: "1.0.0"}
}

// ProgressWidget displays progress bars
type ProgressWidget struct {
	storage DataStorage
	logger  Logger
}

func (pw *ProgressWidget) GetType() WidgetType          { return WidgetTypeProgress }
func (pw *ProgressWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	return map[string]interface{}{"progress": 75}, nil
}
func (pw *ProgressWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	return `<div class="progress-widget">Progress Widget</div>`, nil
}
func (pw *ProgressWidget) Validate(config map[string]interface{}) error { return nil }
func (pw *ProgressWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{Name: "Progress Widget", Version: "1.0.0"}
}

// TextWidget displays static text content
type TextWidget struct {
	logger Logger
}

func (tw *TextWidget) GetType() WidgetType          { return WidgetTypeText }
func (tw *TextWidget) GetData(ctx context.Context, config DataSourceConfig) (interface{}, error) {
	return map[string]interface{}{"content": "Static text content"}, nil
}
func (tw *TextWidget) Render(data interface{}, style WidgetStyle) (string, error) {
	return `<div class="text-widget">Text Widget</div>`, nil
}
func (tw *TextWidget) Validate(config map[string]interface{}) error { return nil }
func (tw *TextWidget) GetMetadata() WidgetMetadata {
	return WidgetMetadata{Name: "Text Widget", Version: "1.0.0"}
}

// Data structures for widgets

type ChartData struct {
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

type ChartDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BorderColor     string    `json:"borderColor"`
	BackgroundColor string    `json:"backgroundColor"`
}

type TableData struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

type MetricData struct {
	Value         float64   `json:"value"`
	Label         string    `json:"label"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	Timestamp     time.Time `json:"timestamp"`
}

type GaugeData struct {
	Value float64 `json:"value"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Label string  `json:"label"`
	Unit  string  `json:"unit"`
}