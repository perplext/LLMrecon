package analytics

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// ExportManager handles data export and integration APIs
type ExportManager struct {
	config           *Config
	storage          DataStorage
	reportGenerator  *ExecutiveReportGenerator
	logger           Logger
	exporters        map[string]Exporter
	integrations     map[string]Integration
}

// Exporter interface for different export formats
type Exporter interface {
	Export(ctx context.Context, data interface{}, writer io.Writer) error
	GetFormat() ExportFormat
	GetContentType() string
	Validate(data interface{}) error
}

// Integration interface for external system integrations
type Integration interface {
	Send(ctx context.Context, data interface{}) error
	GetName() string
	GetEndpoint() string
	IsEnabled() bool
	Validate() error
}

// ExportFormat represents different export formats
type ExportFormat string

const (
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatCSV      ExportFormat = "csv"
	ExportFormatXML      ExportFormat = "xml"
	ExportFormatPDF      ExportFormat = "pdf"
	ExportFormatExcel    ExportFormat = "xlsx"
	ExportFormatPrometheus ExportFormat = "prometheus"
	ExportFormatSplunk   ExportFormat = "splunk"
	ExportFormatElastic  ExportFormat = "elasticsearch"
)

// ExportRequest represents a data export request
type ExportRequest struct {
	ID          string                 `json:"id"`
	Format      ExportFormat           `json:"format"`
	DataType    string                 `json:"data_type"`
	TimeRange   TimeWindow             `json:"time_range"`
	Filters     map[string]interface{} `json:"filters"`
	Options     ExportOptions          `json:"options"`
	RequestedBy string                 `json:"requested_by"`
	RequestedAt time.Time              `json:"requested_at"`
}

// ExportOptions configures export behavior
type ExportOptions struct {
	IncludeMetadata  bool     `json:"include_metadata"`
	CompressOutput   bool     `json:"compress_output"`
	ChunkSize        int      `json:"chunk_size"`
	FieldSelection   []string `json:"field_selection"`
	CustomFormat     map[string]interface{} `json:"custom_format"`
	Aggregation      AggregationConfig `json:"aggregation"`
}

// ExportResult contains export operation results
type ExportResult struct {
	ID           string        `json:"id"`
	Status       ExportStatus  `json:"status"`
	Format       ExportFormat  `json:"format"`
	Size         int64         `json:"size"`
	RecordCount  int           `json:"record_count"`
	Duration     time.Duration `json:"duration"`
	DownloadURL  string        `json:"download_url"`
	ExpiresAt    time.Time     `json:"expires_at"`
	Error        string        `json:"error,omitempty"`
	CompletedAt  time.Time     `json:"completed_at"`
}

// ExportStatus represents export operation status
type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "pending"
	ExportStatusProcessing ExportStatus = "processing"
	ExportStatusCompleted  ExportStatus = "completed"
	ExportStatusFailed     ExportStatus = "failed"
	ExportStatusExpired    ExportStatus = "expired"
)

// IntegrationConfig configures external integrations
type IntegrationConfig struct {
	Name        string                 `json:"name"`
	Type        IntegrationType        `json:"type"`
	Endpoint    string                 `json:"endpoint"`
	Credentials map[string]string      `json:"credentials"`
	Settings    map[string]interface{} `json:"settings"`
	Enabled     bool                   `json:"enabled"`
	Schedule    string                 `json:"schedule"`
}

// IntegrationType represents different integration types
type IntegrationType string

const (
	IntegrationTypeSIEM      IntegrationType = "siem"
	IntegrationTypeTicketing IntegrationType = "ticketing"
	IntegrationTypeMonitoring IntegrationType = "monitoring"
	IntegrationTypeBI        IntegrationType = "business_intelligence"
	IntegrationTypeWebhook   IntegrationType = "webhook"
	IntegrationTypeAPI       IntegrationType = "api"
)

// NewExportManager creates a new export manager
func NewExportManager(config *Config, storage DataStorage, reportGenerator *ExecutiveReportGenerator, logger Logger) *ExportManager {
	manager := &ExportManager{
		config:          config,
		storage:         storage,
		reportGenerator: reportGenerator,
		logger:          logger,
		exporters:       make(map[string]Exporter),
		integrations:    make(map[string]Integration),
	}
	
	// Register default exporters
	manager.registerDefaultExporters()
	
	// Register default integrations
	manager.registerDefaultIntegrations()
	
	return manager
}

// ExportData exports data in the specified format
func (em *ExportManager) ExportData(ctx context.Context, request ExportRequest, writer io.Writer) (*ExportResult, error) {
	startTime := time.Now()
	
	em.logger.Info("Starting data export", "id", request.ID, "format", request.Format, "dataType", request.DataType)
	
	// Get exporter for format
	exporter, exists := em.exporters[string(request.Format)]
	if !exists {
		return nil, fmt.Errorf("unsupported export format: %s", request.Format)
	}
	
	// Fetch data based on request
	data, err := em.fetchDataForExport(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	
	// Validate data for export
	if err := exporter.Validate(data); err != nil {
		return nil, fmt.Errorf("data validation failed: %w", err)
	}
	
	// Perform export
	if err := exporter.Export(ctx, data, writer); err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}
	
	duration := time.Since(startTime)
	
	result := &ExportResult{
		ID:          request.ID,
		Status:      ExportStatusCompleted,
		Format:      request.Format,
		Duration:    duration,
		CompletedAt: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Default 24-hour expiry
	}
	
	em.logger.Info("Data export completed", "id", request.ID, "duration", duration)
	
	return result, nil
}

// ExportReport exports an executive report
func (em *ExportManager) ExportReport(ctx context.Context, reportID string, format ExportFormat, writer io.Writer) error {
	// This would fetch the report and export it
	// For now, generate a sample report
	report, err := em.reportGenerator.GenerateWeeklyReport(ctx, "system")
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	
	exporter, exists := em.exporters[string(format)]
	if !exists {
		return fmt.Errorf("unsupported export format: %s", format)
	}
	
	return exporter.Export(ctx, report, writer)
}

// SendToIntegration sends data to an external integration
func (em *ExportManager) SendToIntegration(ctx context.Context, integrationName string, data interface{}) error {
	integration, exists := em.integrations[integrationName]
	if !exists {
		return fmt.Errorf("integration not found: %s", integrationName)
	}
	
	if !integration.IsEnabled() {
		return fmt.Errorf("integration is disabled: %s", integrationName)
	}
	
	em.logger.Info("Sending data to integration", "integration", integrationName)
	
	if err := integration.Send(ctx, data); err != nil {
		return fmt.Errorf("failed to send to integration %s: %w", integrationName, err)
	}
	
	em.logger.Info("Data sent successfully to integration", "integration", integrationName)
	
	return nil
}

// RegisterExporter adds a custom exporter
func (em *ExportManager) RegisterExporter(exporter Exporter) {
	em.exporters[string(exporter.GetFormat())] = exporter
	em.logger.Info("Registered exporter", "format", exporter.GetFormat())
}

// RegisterIntegration adds a custom integration
func (em *ExportManager) RegisterIntegration(integration Integration) {
	em.integrations[integration.GetName()] = integration
	em.logger.Info("Registered integration", "name", integration.GetName())
}

// GetSupportedFormats returns list of supported export formats
func (em *ExportManager) GetSupportedFormats() []ExportFormat {
	var formats []ExportFormat
	for format := range em.exporters {
		formats = append(formats, ExportFormat(format))
	}
	return formats
}

// GetAvailableIntegrations returns list of available integrations
func (em *ExportManager) GetAvailableIntegrations() []string {
	var integrations []string
	for name := range em.integrations {
		integrations = append(integrations, name)
	}
	return integrations
}

// Internal methods

func (em *ExportManager) fetchDataForExport(ctx context.Context, request ExportRequest) (interface{}, error) {
	switch request.DataType {
	case "metrics":
		return em.storage.GetMetricsByTimeRange(ctx, request.TimeRange.Start, request.TimeRange.End)
	case "scan_results":
		return em.storage.GetScanResultsByTimeRange(ctx, request.TimeRange.Start, request.TimeRange.End)
	case "aggregated":
		return em.storage.GetAggregatedMetricsByTimeRange(ctx, request.TimeRange.Start, request.TimeRange.End)
	default:
		return nil, fmt.Errorf("unsupported data type: %s", request.DataType)
	}
}

func (em *ExportManager) registerDefaultExporters() {
	em.exporters[string(ExportFormatJSON)] = &JSONExporter{}
	em.exporters[string(ExportFormatCSV)] = &CSVExporter{}
	em.exporters[string(ExportFormatPrometheus)] = &PrometheusExporter{}
	em.exporters[string(ExportFormatSplunk)] = &SplunkExporter{}
}

func (em *ExportManager) registerDefaultIntegrations() {
	em.integrations["splunk"] = &SplunkIntegration{
		endpoint: "https://splunk.example.com/api/collect",
		enabled:  false,
	}
	em.integrations["elasticsearch"] = &ElasticsearchIntegration{
		endpoint: "https://elastic.example.com:9200",
		enabled:  false,
	}
	em.integrations["webhook"] = &WebhookIntegration{
		endpoint: "https://webhook.example.com/analytics",
		enabled:  false,
	}
}

// Default Exporters

// JSONExporter exports data as JSON
type JSONExporter struct{}

func (je *JSONExporter) Export(ctx context.Context, data interface{}, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (je *JSONExporter) GetFormat() ExportFormat { return ExportFormatJSON }
func (je *JSONExporter) GetContentType() string { return "application/json" }
func (je *JSONExporter) Validate(data interface{}) error { return nil }

// CSVExporter exports data as CSV
type CSVExporter struct{}

func (ce *CSVExporter) Export(ctx context.Context, data interface{}, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()
	
	switch d := data.(type) {
	case []Metric:
		// Write CSV header
		header := []string{"ID", "Name", "Type", "Value", "Timestamp", "Labels"}
		if err := csvWriter.Write(header); err != nil {
			return err
		}
		
		// Write data rows
		for _, metric := range d {
			labels := ""
			for k, v := range metric.Labels {
				labels += fmt.Sprintf("%s:%s;", k, v)
			}
			
			record := []string{
				metric.ID,
				metric.Name,
				string(metric.Type),
				fmt.Sprintf("%.6f", metric.Value),
				metric.Timestamp.Format(time.RFC3339),
				labels,
			}
			
			if err := csvWriter.Write(record); err != nil {
				return err
			}
		}
		
	default:
		return fmt.Errorf("unsupported data type for CSV export: %T", data)
	}
	
	return nil
}

func (ce *CSVExporter) GetFormat() ExportFormat { return ExportFormatCSV }
func (ce *CSVExporter) GetContentType() string { return "text/csv" }
func (ce *CSVExporter) Validate(data interface{}) error { return nil }

// PrometheusExporter exports data in Prometheus format
type PrometheusExporter struct{}

func (pe *PrometheusExporter) Export(ctx context.Context, data interface{}, writer io.Writer) error {
	switch d := data.(type) {
	case []Metric:
		for _, metric := range d {
			// Convert to Prometheus format
			metricName := strings.ReplaceAll(metric.Name, ".", "_")
			metricName = strings.ReplaceAll(metricName, "-", "_")
			
			// Write metric with labels
			labelStr := ""
			if len(metric.Labels) > 0 {
				var labels []string
				for k, v := range metric.Labels {
					labels = append(labels, fmt.Sprintf(`%s="%s"`, k, v))
				}
				labelStr = fmt.Sprintf("{%s}", strings.Join(labels, ","))
			}
			
			line := fmt.Sprintf("%s%s %.6f %d\n", 
				metricName, 
				labelStr, 
				metric.Value, 
				metric.Timestamp.Unix()*1000)
			
			if _, err := writer.Write([]byte(line)); err != nil {
				return err
			}
		}
		
	default:
		return fmt.Errorf("unsupported data type for Prometheus export: %T", data)
	}
	
	return nil
}

func (pe *PrometheusExporter) GetFormat() ExportFormat { return ExportFormatPrometheus }
func (pe *PrometheusExporter) GetContentType() string { return "text/plain" }
func (pe *PrometheusExporter) Validate(data interface{}) error { return nil }

// SplunkExporter exports data in Splunk format
type SplunkExporter struct{}

func (se *SplunkExporter) Export(ctx context.Context, data interface{}, writer io.Writer) error {
	switch d := data.(type) {
	case []Metric:
		for _, metric := range d {
			// Create Splunk event
			event := map[string]interface{}{
				"time":       metric.Timestamp.Unix(),
				"sourcetype": "analytics:metric",
				"event": map[string]interface{}{
					"metric_id":    metric.ID,
					"metric_name":  metric.Name,
					"metric_type":  metric.Type,
					"metric_value": metric.Value,
					"labels":       metric.Labels,
					"metadata":     metric.Metadata,
				},
			}
			
			jsonData, err := json.Marshal(event)
			if err != nil {
				return err
			}
			
			if _, err := writer.Write(append(jsonData, '\n')); err != nil {
				return err
			}
		}
		
	default:
		return fmt.Errorf("unsupported data type for Splunk export: %T", data)
	}
	
	return nil
}

func (se *SplunkExporter) GetFormat() ExportFormat { return ExportFormatSplunk }
func (se *SplunkExporter) GetContentType() string { return "application/json" }
func (se *SplunkExporter) Validate(data interface{}) error { return nil }

// Default Integrations

// SplunkIntegration sends data to Splunk
type SplunkIntegration struct {
	endpoint string
	token    string
	enabled  bool
}

func (si *SplunkIntegration) Send(ctx context.Context, data interface{}) error {
	// Mock implementation - would actually send HTTP POST to Splunk
	fmt.Printf("Sending data to Splunk: %s\n", si.endpoint)
	return nil
}

func (si *SplunkIntegration) GetName() string { return "splunk" }
func (si *SplunkIntegration) GetEndpoint() string { return si.endpoint }
func (si *SplunkIntegration) IsEnabled() bool { return si.enabled }
func (si *SplunkIntegration) Validate() error { return nil }

// ElasticsearchIntegration sends data to Elasticsearch
type ElasticsearchIntegration struct {
	endpoint string
	index    string
	enabled  bool
}

func (ei *ElasticsearchIntegration) Send(ctx context.Context, data interface{}) error {
	// Mock implementation - would actually send to Elasticsearch
	fmt.Printf("Sending data to Elasticsearch: %s/%s\n", ei.endpoint, ei.index)
	return nil
}

func (ei *ElasticsearchIntegration) GetName() string { return "elasticsearch" }
func (ei *ElasticsearchIntegration) GetEndpoint() string { return ei.endpoint }
func (ei *ElasticsearchIntegration) IsEnabled() bool { return ei.enabled }
func (ei *ElasticsearchIntegration) Validate() error { return nil }

// WebhookIntegration sends data to a webhook endpoint
type WebhookIntegration struct {
	endpoint string
	headers  map[string]string
	enabled  bool
}

func (wi *WebhookIntegration) Send(ctx context.Context, data interface{}) error {
	// Mock implementation - would actually send HTTP POST
	fmt.Printf("Sending data to webhook: %s\n", wi.endpoint)
	return nil
}

func (wi *WebhookIntegration) GetName() string { return "webhook" }
func (wi *WebhookIntegration) GetEndpoint() string { return wi.endpoint }
func (wi *WebhookIntegration) IsEnabled() bool { return wi.enabled }
func (wi *WebhookIntegration) Validate() error { return nil }