// Package errors provides error handling functionality for bundle operations
package errors

import (
	"io"
	"time"
)

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ErrorReporter defines the interface for error reporting
type ErrorReporter interface {
	// ReportError reports an error
	ReportError(ctx context.Context, err *BundleError) error
	// GenerateErrorReport generates an error report for a list of errors
	GenerateErrorReport(ctx context.Context, errors []*BundleError, outputPath string) error
	// GetErrorStatistics returns statistics about errors
	GetErrorStatistics(errors []*BundleError) map[string]interface{}
}

// ErrorReport represents a structured error report
type ErrorReport struct {
	// ReportID is the unique identifier for the report
	ReportID string `json:"report_id"`
	// Timestamp is the time the report was generated
	Timestamp time.Time `json:"timestamp"`
	// Errors is the list of errors in the report
	Errors []*BundleError `json:"errors"`
	// Statistics is the error statistics
	Statistics map[string]interface{} `json:"statistics"`
	// SystemInfo is information about the system
	SystemInfo map[string]interface{} `json:"system_info"`
}

// DefaultErrorReporter is the default implementation of ErrorReporter
type DefaultErrorReporter struct {
	// Logger is the logger for error reporting
	Logger io.Writer
	// ReportsDir is the directory for error reports
	ReportsDir string
	// NotificationService is the service for sending notifications
	NotificationService NotificationService
	// IncludeSystemInfo indicates whether to include system information in reports
	IncludeSystemInfo bool
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(logger io.Writer, reportsDir string) ErrorReporter {
	if logger == nil {
		logger = os.Stdout
	}
	
	if reportsDir == "" {
		reportsDir = filepath.Join(os.TempDir(), "error-reports")
	}
	
	// Ensure reports directory exists
	os.MkdirAll(reportsDir, 0755)
	
	return &DefaultErrorReporter{
		Logger:            logger,
		ReportsDir:        reportsDir,
		IncludeSystemInfo: true,
	}
}

// ReportError reports an error
func (r *DefaultErrorReporter) ReportError(ctx context.Context, bundleErr *BundleError) error {
	if bundleErr == nil {
		return nil
	}
	
	// Log the error
	fmt.Fprintf(r.Logger, "[ERROR] %s (ID: %s, Category: %s, Severity: %s)\n",
		bundleErr.Message, bundleErr.ErrorID, bundleErr.Category, bundleErr.Severity)
	
	// Generate a report file name
	reportFile := filepath.Join(r.ReportsDir, fmt.Sprintf("error_%s_%s.json",
		bundleErr.ErrorID, time.Now().Format("20060102_150405")))
	
	// Create a report with a single error
	report := &ErrorReport{
		ReportID:   bundleErr.ErrorID,
		Timestamp:  time.Now(),
		Errors:     []*BundleError{bundleErr},
		Statistics: map[string]interface{}{
			"total_errors": 1,
			"categories":   map[string]int{string(bundleErr.Category): 1},
			"severities":   map[string]int{string(bundleErr.Severity): 1},
		},
	}
	
	// Add system information if enabled
	if r.IncludeSystemInfo {
		report.SystemInfo = getSystemInfo()
	}
	
	// Write the report to file
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal error report: %w", err)
	}
	
	if err := os.WriteFile(reportFile, reportJSON, 0644); err != nil {
		return fmt.Errorf("failed to write error report: %w", err)
	}
	
	fmt.Fprintf(r.Logger, "Error report written to: %s\n", reportFile)
	
	// Send notification if notification service is configured and error is critical
	if r.NotificationService != nil && bundleErr.Severity == CriticalSeverity {
		if notifyErr := r.NotificationService.SendNotification(ctx, "Critical Error", fmt.Sprintf(
			"Critical error occurred: %s (ID: %s)", bundleErr.Message, bundleErr.ErrorID)); notifyErr != nil {
			fmt.Fprintf(r.Logger, "Failed to send notification: %s\n", notifyErr.Error())
		}
	}
	
	return nil
}

// GenerateErrorReport generates an error report for a list of errors
func (r *DefaultErrorReporter) GenerateErrorReport(ctx context.Context, errors []*BundleError, outputPath string) error {
	if len(errors) == 0 {
		return nil
	}
	
	// Generate a report ID
	reportID := fmt.Sprintf("report_%s", time.Now().Format("20060102_150405"))
	
	// Create a report
	report := &ErrorReport{
		ReportID:   reportID,
		Timestamp:  time.Now(),
		Errors:     errors,
		Statistics: r.GetErrorStatistics(errors),
	}
	
	// Add system information if enabled
	if r.IncludeSystemInfo {
		report.SystemInfo = getSystemInfo()
	}
	
	// If output path is not specified, generate one
	if outputPath == "" {
		outputPath = filepath.Join(r.ReportsDir, fmt.Sprintf("error_report_%s.json", reportID))
	}
	
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(outputPath), 0755)
	
	// Write the report to file
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal error report: %w", err)
	}
	
	if err := os.WriteFile(outputPath, reportJSON, 0644); err != nil {
		return fmt.Errorf("failed to write error report: %w", err)
	}
	
	fmt.Fprintf(r.Logger, "Error report written to: %s\n", outputPath)
	
	// Generate a human-readable summary
	summaryPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + "_summary.txt"
	if err := r.generateErrorSummary(report, summaryPath); err != nil {
		fmt.Fprintf(r.Logger, "Failed to generate error summary: %s\n", err.Error())
	}
	
	return nil
}

// generateErrorSummary generates a human-readable summary of an error report
func (r *DefaultErrorReporter) generateErrorSummary(report *ErrorReport, outputPath string) error {
	// Open the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %w", err)
	}
	defer file.Close()
	
	// Write the summary header
	fmt.Fprintf(file, "Error Report Summary\n")
	fmt.Fprintf(file, "===================\n\n")
	fmt.Fprintf(file, "Report ID: %s\n", report.ReportID)
	fmt.Fprintf(file, "Timestamp: %s\n\n", report.Timestamp.Format(time.RFC3339))
	
	// Write error statistics
	fmt.Fprintf(file, "Error Statistics\n")
	fmt.Fprintf(file, "----------------\n")
	fmt.Fprintf(file, "Total Errors: %d\n\n", report.Statistics["total_errors"])
	
	// Write error categories
	if categories, ok := report.Statistics["categories"].(map[string]int); ok {
		fmt.Fprintf(file, "Error Categories:\n")
		for category, count := range categories {
			fmt.Fprintf(file, "  - %s: %d\n", category, count)
		}
		fmt.Fprintf(file, "\n")
	}
	
	// Write error severities
	if severities, ok := report.Statistics["severities"].(map[string]int); ok {
		fmt.Fprintf(file, "Error Severities:\n")
		for severity, count := range severities {
			fmt.Fprintf(file, "  - %s: %d\n", severity, count)
		}
		fmt.Fprintf(file, "\n")
	}
	
	// Write detailed error list
	fmt.Fprintf(file, "Detailed Error List\n")
	fmt.Fprintf(file, "-------------------\n")
	for i, err := range report.Errors {
		fmt.Fprintf(file, "Error #%d:\n", i+1)
		fmt.Fprintf(file, "  ID: %s\n", err.ErrorID)
		fmt.Fprintf(file, "  Message: %s\n", err.Message)
		if err.Original != nil {
			fmt.Fprintf(file, "  Original Error: %s\n", err.Original.Error())
		}
		fmt.Fprintf(file, "  Category: %s\n", err.Category)
		fmt.Fprintf(file, "  Severity: %s\n", err.Severity)
		fmt.Fprintf(file, "  Recoverability: %s\n", err.Recoverability)
		fmt.Fprintf(file, "  Timestamp: %s\n", err.Timestamp.Format(time.RFC3339))
		
		// Write error context
		if len(err.Context) > 0 {
			fmt.Fprintf(file, "  Context:\n")
			for k, v := range err.Context {
				fmt.Fprintf(file, "    %s: %v\n", k, v)
			}
		}
		
		fmt.Fprintf(file, "\n")
	}
	
	// Write system information
	if report.SystemInfo != nil && len(report.SystemInfo) > 0 {
		fmt.Fprintf(file, "System Information\n")
		fmt.Fprintf(file, "------------------\n")
		for k, v := range report.SystemInfo {
			fmt.Fprintf(file, "%s: %v\n", k, v)
		}
	}
	
	fmt.Fprintf(r.Logger, "Error summary written to: %s\n", outputPath)
	return nil
}

// GetErrorStatistics returns statistics about errors
func (r *DefaultErrorReporter) GetErrorStatistics(errors []*BundleError) map[string]interface{} {
	if len(errors) == 0 {
		return map[string]interface{}{
			"total_errors": 0,
		}
	}
	
	// Initialize statistics
	statistics := map[string]interface{}{
		"total_errors": len(errors),
		"categories":   make(map[string]int),
		"severities":   make(map[string]int),
		"recoverable":  0,
		"nonrecoverable": 0,
		"retry_attempts": 0,
	}
	
	// Categorize errors
	categories := statistics["categories"].(map[string]int)
	severities := statistics["severities"].(map[string]int)
	
	for _, err := range errors {
		// Count by category
		categories[string(err.Category)]++
		
		// Count by severity
		severities[string(err.Severity)]++
		
		// Count by recoverability
		if err.Recoverability == RecoverableError {
			statistics["recoverable"] = statistics["recoverable"].(int) + 1
		} else {
			statistics["nonrecoverable"] = statistics["nonrecoverable"].(int) + 1
		}
		
		// Count retry attempts
		statistics["retry_attempts"] = statistics["retry_attempts"].(int) + err.RetryAttempt
	}
	
	return statistics
}

// getSystemInfo returns information about the system
func getSystemInfo() map[string]interface{} {
	info := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"hostname":  "unknown",
	}
	
	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}
	
	// Get OS information
	info["os"] = map[string]interface{}{
		"name": os.Getenv("GOOS"),
		"arch": os.Getenv("GOARCH"),
	}
	
	return info
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	// SendNotification sends a notification
	SendNotification(ctx context.Context, subject, message string) error
}

// EmailNotificationService implements NotificationService for email notifications
type EmailNotificationService struct {
	// Recipients is the list of email recipients
	Recipients []string
	// SenderEmail is the sender email address
	SenderEmail string
	// SMTPServer is the SMTP server
	SMTPServer string
	// SMTPPort is the SMTP port
	SMTPPort int
	// SMTPUsername is the SMTP username
	SMTPUsername string
	// SMTPPassword is the SMTP password
	SMTPPassword string
	// Logger is the logger for notification events
	Logger io.Writer
}

// NewEmailNotificationService creates a new email notification service
func NewEmailNotificationService(recipients []string, senderEmail, smtpServer string, smtpPort int, logger io.Writer) *EmailNotificationService {
	if logger == nil {
		logger = os.Stdout
	}
	
	return &EmailNotificationService{
		Recipients:   recipients,
		SenderEmail:  senderEmail,
		SMTPServer:   smtpServer,
		SMTPPort:     smtpPort,
		Logger:       logger,
	}
}

// SendNotification sends an email notification
func (s *EmailNotificationService) SendNotification(ctx context.Context, subject, message string) error {
	if len(s.Recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}
	
	if s.SMTPServer == "" {
		return fmt.Errorf("SMTP server not configured")
	}
	
	// In a real implementation, this would send an email
	// For now, we'll just log the notification
	fmt.Fprintf(s.Logger, "Would send email notification:\n")
	fmt.Fprintf(s.Logger, "  From: %s\n", s.SenderEmail)
	fmt.Fprintf(s.Logger, "  To: %s\n", strings.Join(s.Recipients, ", "))
	fmt.Fprintf(s.Logger, "  Subject: %s\n", subject)
	fmt.Fprintf(s.Logger, "  Message: %s\n", message)
	
	return nil
}

// SlackNotificationService implements NotificationService for Slack notifications
type SlackNotificationService struct {
	// WebhookURL is the Slack webhook URL
	WebhookURL string
	// Channel is the Slack channel
	Channel string
	// Username is the username to use for notifications
	Username string
	// Logger is the logger for notification events
	Logger io.Writer
}

// NewSlackNotificationService creates a new Slack notification service
func NewSlackNotificationService(webhookURL, channel, username string, logger io.Writer) *SlackNotificationService {
	if logger == nil {
		logger = os.Stdout
	}
	
	return &SlackNotificationService{
		WebhookURL: webhookURL,
		Channel:    channel,
		Username:   username,
		Logger:     logger,
	}
}

// SendNotification sends a Slack notification
func (s *SlackNotificationService) SendNotification(ctx context.Context, subject, message string) error {
	if s.WebhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}
	
	// In a real implementation, this would send a Slack message
	// For now, we'll just log the notification
	fmt.Fprintf(s.Logger, "Would send Slack notification:\n")
	fmt.Fprintf(s.Logger, "  Channel: %s\n", s.Channel)
	fmt.Fprintf(s.Logger, "  Username: %s\n", s.Username)
	fmt.Fprintf(s.Logger, "  Subject: %s\n", subject)
	fmt.Fprintf(s.Logger, "  Message: %s\n", message)
	
	return nil
}
