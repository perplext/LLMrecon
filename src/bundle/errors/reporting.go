// Package errors provides error handling functionality for bundle operations
package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ErrorReporter defines the interface for error reporting
type ErrorReporter interface {
	Report(ctx context.Context, err *BundleError) error
	GenerateReport(ctx context.Context, errors []*BundleError) (*ErrorReport, error)
}

// ErrorReport represents a collection of errors with statistics
type ErrorReport struct {
	GeneratedAt time.Time      `json:"generated_at"`
	TotalErrors int            `json:"total_errors"`
	Statistics  ErrorStats     `json:"statistics"`
	Errors      []*BundleError `json:"errors"`
}

// ErrorStats contains error statistics
type ErrorStats struct {
	BySeverity       map[string]int `json:"by_severity"`
	ByCategory       map[string]int `json:"by_category"`
	ByRecoverability map[string]int `json:"by_recoverability"`
}

// DefaultErrorReporter is the default implementation of ErrorReporter
type DefaultErrorReporter struct {
	Writer      io.Writer
	AuditLogger *AuditLogger
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(writer io.Writer, auditLogger *AuditLogger) *DefaultErrorReporter {
	if writer == nil {
		writer = os.Stdout
	}
	
	return &DefaultErrorReporter{
		Writer:      writer,
		AuditLogger: auditLogger,
	}
}

// Report reports a single error
func (r *DefaultErrorReporter) Report(ctx context.Context, err *BundleError) error {
	if err == nil {
		return nil
	}
	
	// Log the error
	fmt.Fprintf(r.Writer, "Error Report: %s (ID: %s, Category: %s, Severity: %s)\n", 
		err.Message, err.ErrorID, err.Category, err.Severity)
	
	// Log audit event
	if r.AuditLogger != nil {
		details := map[string]interface{}{
			"error_id":       err.ErrorID,
			"category":       err.Category,
			"severity":       err.Severity,
			"recoverability": err.Recoverability,
			"message":        err.Message,
		}
		
		// Add context if available
		for k, v := range err.Context {
			details[k] = v
		}
		
		r.AuditLogger.LogEventWithStatus("error_reported", "ErrorReporter", err.ErrorID, "error", details)
	}
	
	return nil
}

// GenerateReport generates a comprehensive error report
func (r *DefaultErrorReporter) GenerateReport(ctx context.Context, errors []*BundleError) (*ErrorReport, error) {
	stats := calculateErrorStats(errors)
	
	report := &ErrorReport{
		GeneratedAt: time.Now(),
		TotalErrors: len(errors),
		Statistics:  stats,
		Errors:      errors,
	}
	
	return report, nil
}

// calculateErrorStats calculates statistics from errors
func calculateErrorStats(errors []*BundleError) ErrorStats {
	stats := ErrorStats{
		BySeverity:       make(map[string]int),
		ByCategory:       make(map[string]int),
		ByRecoverability: make(map[string]int),
	}
	
	for _, err := range errors {
		stats.BySeverity[string(err.Severity)]++
		stats.ByCategory[string(err.Category)]++
		stats.ByRecoverability[string(err.Recoverability)]++
	}
	
	return stats
}

// WriteReportJSON writes an error report as JSON
func WriteReportJSON(w io.Writer, report *ErrorReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// WriteReportText writes an error report as text
func WriteReportText(w io.Writer, report *ErrorReport) error {
	fmt.Fprintf(w, "Error Report\n")
	fmt.Fprintf(w, "============\n\n")
	fmt.Fprintf(w, "Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Total Errors: %d\n\n", report.TotalErrors)
	
	// Write statistics
	fmt.Fprintf(w, "Statistics:\n")
	fmt.Fprintf(w, "-----------\n")
	
	fmt.Fprintf(w, "By Severity:\n")
	for severity, count := range report.Statistics.BySeverity {
		fmt.Fprintf(w, "  %s: %d\n", severity, count)
	}
	
	fmt.Fprintf(w, "\nBy Category:\n")
	for category, count := range report.Statistics.ByCategory {
		fmt.Fprintf(w, "  %s: %d\n", category, count)
	}
	
	fmt.Fprintf(w, "\nBy Recoverability:\n")
	for recoverability, count := range report.Statistics.ByRecoverability {
		fmt.Fprintf(w, "  %s: %d\n", recoverability, count)
	}
	
	// Write individual errors
	fmt.Fprintf(w, "\n\nErrors:\n")
	fmt.Fprintf(w, "-------\n")
	for i, err := range report.Errors {
		fmt.Fprintf(w, "\n%d. %s\n", i+1, err.Message)
		fmt.Fprintf(w, "   ID: %s\n", err.ErrorID)
		fmt.Fprintf(w, "   Category: %s\n", err.Category)
		fmt.Fprintf(w, "   Severity: %s\n", err.Severity)
		fmt.Fprintf(w, "   Recoverability: %s\n", err.Recoverability)
		
		if len(err.Context) > 0 {
			fmt.Fprintf(w, "   Context:\n")
			for k, v := range err.Context {
				fmt.Fprintf(w, "     %s: %v\n", k, v)
			}
		}
	}
	
	return nil
}
