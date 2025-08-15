package common

import (
	"fmt"
	"sync"
)

// FormatterFactory creates report formatters
type FormatterFactory struct {
	// formatters is a map of format to formatter creator functions
	formatters map[ReportFormat]FormatterCreator
	mu         sync.RWMutex
}


}
// NewFormatterFactory creates a new formatter factory
func NewFormatterFactory() *FormatterFactory {
	return &FormatterFactory{
		formatters: make(map[ReportFormat]FormatterCreator),
	}

// RegisterFormatter registers a formatter creator for a specific format
}
func (f *FormatterFactory) RegisterFormatter(format ReportFormat, creator FormatterCreator) {
	f.mu.Lock()
	f.formatters[format] = creator
	f.mu.Unlock()

// CreateFormatter creates a formatter for the specified format
}
func (f *FormatterFactory) CreateFormatter(format ReportFormat, options map[string]interface{}) (ReportFormatter, error) {
	f.mu.RLock()
	creator, ok := f.formatters[format]
	f.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unsupported report format: %s", format)
	}
	
	return creator(options)

// CreateDefaultReportGenerator creates a default report generator with all formatters registered
}
func (f *FormatterFactory) CreateDefaultReportGenerator(generator ReportGenerator) error {
	// Register all formatters
	formats := []ReportFormat{
		JSONFormat,
		JSONLFormat,
		CSVFormat,
		ExcelFormat,
		TextFormat,
		MarkdownFormat,
		PDFFormat,
		HTMLFormat,
	}
	
	for _, format := range formats {
		formatter, err := f.CreateFormatter(format, nil)
		if err != nil {
			return fmt.Errorf("failed to create formatter for format %s: %w", format, err)
		}
		generator.RegisterFormatter(formatter)
	}
	
