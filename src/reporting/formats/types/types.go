// Package types provides common types for the reporting formatters
package types

import (
	"context"

	reportTypes "github.com/perplext/LLMrecon/src/reporting/types"
)

// ReportFormatter is the interface for report formatters
type ReportFormatter interface {
	// GetFormat returns the format of the formatter
	GetFormat() reportTypes.ReportFormat
	// Format formats a report
	Format(ctx context.Context, report *reportTypes.Report, options *reportTypes.ReportOptions) ([]byte, error)
	// WriteToFile writes a report to a file
	WriteToFile(ctx context.Context, report *reportTypes.Report, options *reportTypes.ReportOptions, filePath string) error
}
