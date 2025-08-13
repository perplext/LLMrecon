package reporting

import (
	"github.com/perplext/LLMrecon/src/reporting/common"
	// Import formats package to ensure formatters are registered
	_ "github.com/perplext/LLMrecon/src/reporting/formats"
)

// init initializes the reporting package
func init() {
	// Formatters are registered in their respective packages
}

// CreateFormatter creates a formatter for the specified format
func CreateFormatter(format common.ReportFormat, options map[string]interface{}) (common.ReportFormatter, error) {
	return common.CreateFormatter(format, options)
}

// CreateDefaultReportGenerator creates a default report generator with all formatters registered
func CreateDefaultReportGenerator() (*DefaultReportGenerator, error) {
	generator := NewReportGenerator()
	
	// Register all formatters
	formats := []common.ReportFormat{
		common.JSONFormat,
		common.JSONLFormat,
		common.CSVFormat,
		common.ExcelFormat,
		common.TextFormat,
		common.MarkdownFormat,
		common.PDFFormat,
		common.HTMLFormat,
	}
	
	for _, format := range formats {
		formatter, err := CreateFormatter(format, nil)
		if err != nil {
			return nil, err
		}
		generator.RegisterFormatter(formatter)
	}
	
	return generator, nil
}
