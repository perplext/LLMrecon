// Package interfaces provides common interfaces for the reporting system
package interfaces

import (
	"github.com/perplext/LLMrecon/src/reporting/common"
)

// ReportFormat represents the format of a report
type ReportFormat = common.ReportFormat

// Supported report formats
const (
	TextFormat     = common.TextFormat
	MarkdownFormat = common.MarkdownFormat
	JSONFormat     = common.JSONFormat
	JSONLFormat    = common.JSONLFormat
	CSVFormat      = common.CSVFormat
	ExcelFormat    = common.ExcelFormat
	PDFFormat      = common.PDFFormat
	HTMLFormat     = common.HTMLFormat
)

// SeverityLevelMapping maps string representations to SeverityLevel constants
var SeverityLevelMapping = map[string]common.SeverityLevel{
	"critical": common.Critical,
	"high":     common.High,
	"medium":   common.Medium,
	"low":      common.Low,
	"info":     common.Info,

// ReportFormatter is the interface for report formatters
type ReportFormatter = common.ReportFormatter

// ReportGenerator is the interface for report generators
