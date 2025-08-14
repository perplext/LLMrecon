// Package audit provides shared audit logging functionality
package audit

import (
	"time"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os/user"
)

// AuditFormat represents the format for exporting audit logs
type AuditFormat string

const (
	// JSONAuditFormat represents JSON format for audit logs
	JSONAuditFormat AuditFormat = "json"
	// CSVAuditFormat represents CSV format for audit logs
	CSVAuditFormat AuditFormat = "csv"
)

// AuditEvent represents an audit event
type AuditEvent struct {
	// Timestamp is the time the event occurred
	Timestamp time.Time `json:"timestamp"`
	// EventType is the type of event
	EventType string `json:"event_type"`
	// Component is the component associated with the event
	Component string `json:"component"`
	// User is the user who triggered the event
	User string `json:"user,omitempty"`
	// ID is the ID associated with the event
	ID string `json:"id,omitempty"`
	// Operation is the specific operation being performed
	Operation string `json:"operation,omitempty"`
	// Status is the status of the operation (success, failure, in-progress)
	Status string `json:"status,omitempty"`
	// Details contains additional details about the event
	Details map[string]interface{} `json:"details,omitempty"`
}

// AuditLogger implements audit logging functionality
type AuditLogger struct {
	// Writer is the writer for audit logs
	Writer io.Writer
	// User is the user performing the operation
	User string
	// Events stores audit events if enabled
	Events []AuditEvent
	// StoreEvents indicates whether to store events in memory
	StoreEvents bool
}

// AuditLoggerOptions contains options for creating a new audit logger
type AuditLoggerOptions struct {
	// Writer is the writer for audit logs
	Writer io.Writer
	// User is the user performing the operation
	User string
	// EnableConsoleOutput indicates whether to output to console
	EnableConsoleOutput bool
	// StoreEvents indicates whether to store events in memory
	StoreEvents bool
	// LogFilePath is the path to the log file
	LogFilePath string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(writer io.Writer, user string) *AuditLogger {
	return &AuditLogger{
		Writer:      writer,
		User:        user,
		Events:      make([]AuditEvent, 0),
		StoreEvents: false,
	}
}

// NewFileAuditLogger creates a new audit logger that writes to a file
func NewFileAuditLogger(filePath string) (*AuditLogger, error) {
	// Create or open the file with append mode
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	// Get the current user if possible
	username := "system"
	if currentUser, err := user.Current(); err == nil {
		username = currentUser.Username
	}

	return &AuditLogger{
		Writer:      file,
		User:        username,
		Events:      make([]AuditEvent, 0),
		StoreEvents: true,
	}, nil
}

// LogEvent logs an event
func (l *AuditLogger) LogEvent(event, component, id string, details map[string]interface{}) {
	l.LogEventWithStatus(event, component, id, "info", details)
}

// LogEventWithStatus logs an event with a status
func (l *AuditLogger) LogEventWithStatus(event, component, id, status string, details map[string]interface{}) {
	if l.Writer == nil && !l.StoreEvents {
		return
	}
	
	timestamp := time.Now()
	user := l.User
	if user == "" {
		user = "system"
	}
	
	// Create audit event
	auditEvent := AuditEvent{
		Timestamp: timestamp,
		EventType: event,
		Component: component,
		User:      user,
		ID:        id,
		Status:    status,
		Details:   details,
	}
	
	// Store event if enabled
	if l.StoreEvents {
		l.Events = append(l.Events, auditEvent)
	}
	
	// Write to log if writer is available
	if l.Writer != nil {
		// Format the details as a string
		detailsStr := ""
		for k, v := range details {
			detailsStr += fmt.Sprintf(" %s=%v", k, v)
		}
		
		// Write the audit log entry
		fmt.Fprintf(l.Writer, "[%s] [%s] [%s] [%s] [%s] [%s]%s\n",
			timestamp.Format(time.RFC3339), status, user, component, event, id, detailsStr)
	}
}

// LogImportStart logs the start of an import operation
func (l *AuditLogger) LogImportStart(bundleID, bundlePath string, options map[string]interface{}) {
	l.LogEvent("import_started", "BundleImporter", bundleID, map[string]interface{}{
		"bundle_path": bundlePath,
		"options":     options,
	})
}

// LogImportComplete logs the completion of an import operation
func (l *AuditLogger) LogImportComplete(bundleID string, success bool, details map[string]interface{}) {
	status := "success"
	if !success {
		status = "failure"
	}
	l.LogEventWithStatus("import_completed", "BundleImporter", bundleID, status, details)
}

// LogValidation logs a validation event
func (l *AuditLogger) LogValidation(bundleID, bundlePath string, level string, success bool, details map[string]interface{}) {
	status := "success"
	if !success {
		status = "failure"
	}
	details["validation_level"] = level
	details["bundle_path"] = bundlePath
	l.LogEventWithStatus("validation", "BundleImporter", bundleID, status, details)
}

// LogBackupCreated logs a backup creation event
func (l *AuditLogger) LogBackupCreated(bundleID, targetDir, backupPath string) {
	l.LogEvent("backup_created", "BundleImporter", bundleID, map[string]interface{}{
		"target_dir":  targetDir,
		"backup_path": backupPath,
	})
}

// LogFileInstallation logs a file installation event
func (l *AuditLogger) LogFileInstallation(bundleID, filePath string, success bool, details map[string]interface{}) {
	status := "success"
	if !success {
		status = "failure"
	}
	details["file_path"] = filePath
	l.LogEventWithStatus("file_installation", "BundleImporter", bundleID, status, details)
}

// LogImportSummary logs a summary of the import operation
func (l *AuditLogger) LogImportSummary(bundleID string, stats map[string]interface{}) {
	l.LogEvent("import_summary", "BundleImporter", bundleID, stats)
}

// FilterEvents filters audit events based on the provided options
func (l *AuditLogger) FilterEvents(options FilterOptions) []AuditEvent {
	if !l.StoreEvents {
		return nil
	}
	
	var filtered []AuditEvent
	
	for _, event := range l.Events {
		// Filter by time range
		if options.StartTime != nil && event.Timestamp.Before(*options.StartTime) {
			continue
		}
		if options.EndTime != nil && event.Timestamp.After(*options.EndTime) {
			continue
		}
		
		// Filter by event type
		if len(options.EventTypes) > 0 && !contains(options.EventTypes, event.EventType) {
			continue
		}
		
		// Filter by ID
		if len(options.IDs) > 0 && !contains(options.IDs, event.ID) {
			continue
		}
		
		// Filter by bundle ID
		if len(options.BundleIDs) > 0 && !contains(options.BundleIDs, event.ID) {
			continue
		}
		
		// Filter by status
		if len(options.Statuses) > 0 && !contains(options.Statuses, event.Status) {
			continue
		}
		
		// Filter by user
		if len(options.Users) > 0 && !contains(options.Users, event.User) {
			continue
		}
		
		filtered = append(filtered, event)
	}
	
	return filtered
}

// FilterOptions defines options for filtering audit events
type FilterOptions struct {
	// StartTime is the start time for filtering events
	StartTime *time.Time
	// EndTime is the end time for filtering events
	EndTime *time.Time
	// EventTypes is a list of event types to include
	EventTypes []string
	// IDs is a list of IDs to include
	IDs []string
	// BundleIDs is a list of bundle IDs to include
	BundleIDs []string
	// Statuses is a list of statuses to include
	Statuses []string
	// Users is a list of users to include
	Users []string
}

// contains checks if a string is in a slice of strings
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ComplianceReportType defines the type of compliance report to generate
type ComplianceReportType string

const (
	// DetailedReport includes all audit events with full details
	DetailedReport ComplianceReportType = "detailed"
	// SummaryReport includes summarized information about audit events
	SummaryReport ComplianceReportType = "summary"
	// ActivityReport includes activity-based information
	ActivityReport ComplianceReportType = "activity"
)

// ComplianceReportOptions defines options for generating compliance reports
type ComplianceReportOptions struct {
	// ReportType is the type of report to generate
	ReportType ComplianceReportType
	// Filter contains filtering options for the report
	Filter FilterOptions
	// Format is the format of the report
	Format AuditFormat
	// IncludeSystemEvents indicates whether to include system events
	IncludeSystemEvents bool
}

// ComplianceReportSummary contains summary information for a compliance report
type ComplianceReportSummary struct {
	// ReportType is the type of report
	ReportType ComplianceReportType `json:"report_type"`
	// GeneratedAt is the time the report was generated
	GeneratedAt time.Time `json:"generated_at"`
	// TimeRange contains the time range for the report
	TimeRange struct {
		Start time.Time `json:"start,omitempty"`
		End   time.Time `json:"end,omitempty"`
	} `json:"time_range,omitempty"`
	// EventCounts contains counts of events by type
	EventCounts map[string]int `json:"event_counts,omitempty"`
	// BundleCounts contains counts of events by bundle ID
	BundleCounts map[string]int `json:"bundle_counts,omitempty"`
	// StatusCounts contains counts of events by status
	StatusCounts map[string]int `json:"status_counts,omitempty"`
	// UserCounts contains counts of events by user
	UserCounts map[string]int `json:"user_counts,omitempty"`
	// TotalEvents is the total number of events
	TotalEvents int `json:"total_events"`
}

// GenerateComplianceReport generates a compliance report based on audit logs
func (l *AuditLogger) GenerateComplianceReport(writer io.Writer, options ComplianceReportOptions) error {
	if !l.StoreEvents {
		return fmt.Errorf("cannot generate compliance report: audit events not stored")
	}

	// Filter events based on options
	events := l.FilterEvents(options.Filter)

	// Generate report based on type
	switch options.ReportType {
	case DetailedReport:
		// For detailed report, just export all events
		if options.Format == JSONAuditFormat {
			return json.NewEncoder(writer).Encode(events)
		}
		// CSV format
		return l.exportEventsListAsCSV(events, writer)
	case SummaryReport:
		return l.generateSummaryReport(events, writer, options)
	case ActivityReport:
		return l.generateActivityReport(events, writer, options)
	default:
		return fmt.Errorf("unsupported report type: %s", options.ReportType)
	}
}

// generateSummaryReport generates a summary report
func (l *AuditLogger) generateSummaryReport(events []AuditEvent, writer io.Writer, options ComplianceReportOptions) error {
	// Create summary
	summary := ComplianceReportSummary{
		ReportType:  options.ReportType,
		GeneratedAt: time.Now(),
		EventCounts: make(map[string]int),
		BundleCounts: make(map[string]int),
		StatusCounts: make(map[string]int),
		UserCounts:   make(map[string]int),
		TotalEvents: len(events),
	}

	// Set time range if provided
	if options.Filter.StartTime != nil {
		summary.TimeRange.Start = *options.Filter.StartTime
	}
	if options.Filter.EndTime != nil {
		summary.TimeRange.End = *options.Filter.EndTime
	}

	// Count events by type, bundle ID, status, and user
	for _, event := range events {
		summary.EventCounts[event.EventType]++
		summary.BundleCounts[event.ID]++
		summary.StatusCounts[event.Status]++
		summary.UserCounts[event.User]++
	}

	// Write summary in requested format
	if options.Format == JSONAuditFormat {
		return json.NewEncoder(writer).Encode(summary)
	}

	// CSV format
	csv := csv.NewWriter(writer)
	defer csv.Flush()

	// Write header
	csv.Write([]string{"Report Type", "Generated At", "Total Events"})
	csv.Write([]string{string(summary.ReportType), summary.GeneratedAt.Format(time.RFC3339), fmt.Sprintf("%d", summary.TotalEvents)})

	// Write time range
	csv.Write([]string{"Time Range Start", "Time Range End"})
	csv.Write([]string{summary.TimeRange.Start.Format(time.RFC3339), summary.TimeRange.End.Format(time.RFC3339)})

	// Write event counts
	csv.Write([]string{"Event Type", "Count"})
	for eventType, count := range summary.EventCounts {
		csv.Write([]string{eventType, fmt.Sprintf("%d", count)})
	}

	// Write bundle counts
	csv.Write([]string{"Bundle ID", "Count"})
	for bundleID, count := range summary.BundleCounts {
		csv.Write([]string{bundleID, fmt.Sprintf("%d", count)})
	}

	// Write status counts
	csv.Write([]string{"Status", "Count"})
	for status, count := range summary.StatusCounts {
		csv.Write([]string{status, fmt.Sprintf("%d", count)})
	}

	// Write user counts
	csv.Write([]string{"User", "Count"})
	for user, count := range summary.UserCounts {
		csv.Write([]string{user, fmt.Sprintf("%d", count)})
	}

	return nil
}

// generateActivityReport generates an activity report
func (l *AuditLogger) generateActivityReport(events []AuditEvent, writer io.Writer, options ComplianceReportOptions) error {
	// Group events by day
	activitiesByDay := make(map[string][]AuditEvent)
	for _, event := range events {
		day := event.Timestamp.Format("2006-01-02")
		activitiesByDay[day] = append(activitiesByDay[day], event)
	}

	// Create activity report
	activityReport := struct {
		ReportType  ComplianceReportType       `json:"report_type"`
		GeneratedAt time.Time                  `json:"generated_at"`
		TimeRange   struct {
			Start time.Time `json:"start,omitempty"`
			End   time.Time `json:"end,omitempty"`
		} `json:"time_range,omitempty"`
		TotalEvents int                        `json:"total_events"`
		Activities  map[string][]AuditEvent    `json:"activities"`
	}{
		ReportType:  options.ReportType,
		GeneratedAt: time.Now(),
		TotalEvents: len(events),
		Activities:  activitiesByDay,
	}

	// Set time range if provided
	if options.Filter.StartTime != nil {
		activityReport.TimeRange.Start = *options.Filter.StartTime
	}
	if options.Filter.EndTime != nil {
		activityReport.TimeRange.End = *options.Filter.EndTime
	}

	// Write report in requested format
	if options.Format == JSONAuditFormat {
		return json.NewEncoder(writer).Encode(activityReport)
	}

	// CSV format
	csv := csv.NewWriter(writer)
	defer csv.Flush()

	// Write header
	csv.Write([]string{"Report Type", "Generated At", "Total Events"})
	csv.Write([]string{string(activityReport.ReportType), activityReport.GeneratedAt.Format(time.RFC3339), fmt.Sprintf("%d", activityReport.TotalEvents)})

	// Write time range
	csv.Write([]string{"Time Range Start", "Time Range End"})
	csv.Write([]string{activityReport.TimeRange.Start.Format(time.RFC3339), activityReport.TimeRange.End.Format(time.RFC3339)})

	// Write activities by day
	csv.Write([]string{"Date", "Timestamp", "Event Type", "Component", "ID", "User", "Status", "Details"})
	for day, activities := range activityReport.Activities {
		for _, activity := range activities {
			// Format details as string
			detailsStr := ""
			for k, v := range activity.Details {
				detailsStr += fmt.Sprintf("%s=%v; ", k, v)
			}

			csv.Write([]string{
				day,
				activity.Timestamp.Format(time.RFC3339),
				activity.EventType,
				activity.Component,
				activity.ID,
				activity.User,
				activity.Status,
				detailsStr,
			})
		}
	}

	return nil
}

// exportEventsListAsCSV exports a list of audit events as CSV
func (l *AuditLogger) exportEventsListAsCSV(events []AuditEvent, writer io.Writer) error {
	csv := csv.NewWriter(writer)
	defer csv.Flush()

	// Write header
	csv.Write([]string{"Timestamp", "Event Type", "Component", "ID", "User", "Status", "Details"})

	// Write events
	for _, event := range events {
		// Format details as string
		detailsStr := ""
		for k, v := range event.Details {
			detailsStr += fmt.Sprintf("%s=%v; ", k, v)
		}

		csv.Write([]string{
			event.Timestamp.Format(time.RFC3339),
			event.EventType,
			event.Component,
			event.ID,
			event.User,
			event.Status,
			detailsStr,
		})
	}

	return nil
}
