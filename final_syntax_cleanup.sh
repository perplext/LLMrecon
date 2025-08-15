#!/bin/bash

echo "Final syntax cleanup for remaining Go compilation errors..."

# Fix the specific remaining syntax errors
echo "Fixing src/bundle/errors/reporting.go..."
cat > src/bundle/errors/reporting.go << 'EOF'
// Package errors provides error handling functionality for bundle operations
package errors

import (
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
	GeneratedAt time.Time     `json:"generated_at"`
	TotalErrors int           `json:"total_errors"`
	Statistics  ErrorStats    `json:"statistics"`
	Errors      []*BundleError `json:"errors"`
}

// ErrorStats contains error statistics
type ErrorStats struct {
	BySeverity      map[string]int `json:"by_severity"`
	ByCategory      map[string]int `json:"by_category"`
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
		r.AuditLogger.LogEventWithStatus(
			"error_reported",
			"ErrorReporter",
			err.ErrorID,
			"info",
			map[string]interface{}{
				"error_id":  err.ErrorID,
				"category":  string(err.Category),
				"severity":  string(err.Severity),
				"timestamp": time.Now().Format(time.RFC3339),
			},
		)
	}
	
	return nil
}

// GenerateReport generates a comprehensive error report
func (r *DefaultErrorReporter) GenerateReport(ctx context.Context, errors []*BundleError) (*ErrorReport, error) {
	if errors == nil {
		errors = []*BundleError{}
	}
	
	report := &ErrorReport{
		GeneratedAt: time.Now(),
		TotalErrors: len(errors),
		Statistics:  calculateErrorStats(errors),
		Errors:      errors,
	}
	
	return report, nil
}

// calculateErrorStats calculates statistics for the given errors
func calculateErrorStats(errors []*BundleError) ErrorStats {
	stats := ErrorStats{
		BySeverity:       make(map[string]int),
		ByCategory:       make(map[string]int),
		ByRecoverability: make(map[string]int),
	}
	
	for _, err := range errors {
		if err != nil {
			stats.BySeverity[string(err.Severity)]++
			stats.ByCategory[string(err.Category)]++
			stats.ByRecoverability[string(err.Recoverability)]++
		}
	}
	
	return stats
}

// WriteReportJSON writes the error report as JSON
func WriteReportJSON(report *ErrorReport, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// WriteReportText writes the error report as formatted text
func WriteReportText(report *ErrorReport, writer io.Writer) error {
	fmt.Fprintf(writer, "Error Report - Generated at %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(writer, "Total Errors: %d\n\n", report.TotalErrors)
	
	fmt.Fprintf(writer, "Statistics:\n")
	fmt.Fprintf(writer, "  By Severity:\n")
	for severity, count := range report.Statistics.BySeverity {
		fmt.Fprintf(writer, "    %s: %d\n", severity, count)
	}
	
	fmt.Fprintf(writer, "  By Category:\n")
	for category, count := range report.Statistics.ByCategory {
		fmt.Fprintf(writer, "    %s: %d\n", category, count)
	}
	
	fmt.Fprintf(writer, "  By Recoverability:\n")
	for recoverability, count := range report.Statistics.ByRecoverability {
		fmt.Fprintf(writer, "    %s: %d\n", recoverability, count)
	}
	
	fmt.Fprintf(writer, "\nDetailed Errors:\n")
	for i, err := range report.Errors {
		fmt.Fprintf(writer, "%d. %s (ID: %s, Category: %s, Severity: %s)\n", 
			i+1, err.Message, err.ErrorID, err.Category, err.Severity)
	}
	
	return nil
}
EOF

echo "Done with final syntax cleanup!"