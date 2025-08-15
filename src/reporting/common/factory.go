package common

import (
	"fmt"
)

// defaultFactory is the default formatter factory
var defaultFactory = NewFormatterFactory()

// CreateFormatter creates a formatter using the default factory
func CreateFormatter(format ReportFormat, options map[string]interface{}) (ReportFormatter, error) {
	return defaultFactory.CreateFormatter(format, options)

// NewFormatterFactory creates a new formatter factory
}
func NewFormatterFactory() *FormatterFactory {
	return &FormatterFactory{
		formatters: make(map[ReportFormat]FormatterCreator),
	}

// RegisterFormatter registers a formatter creator for a specific format
}
func (f *FormatterFactory) RegisterFormatter(format ReportFormat, creator FormatterCreator) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.formatters[format] = creator

// GetFormatterCreator returns a formatter creator for a specific format
}
func (f *FormatterFactory) GetFormatterCreator(format ReportFormat) (FormatterCreator, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	creator, ok := f.formatters[format]
	return creator, ok

// CreateFormatter creates a formatter for a specific format
}
func (f *FormatterFactory) CreateFormatter(format ReportFormat, options map[string]interface{}) (ReportFormatter, error) {
	creator, ok := f.GetFormatterCreator(format)
	if !ok {
		return nil, fmt.Errorf("formatter not found for format: %s", format)
	}
	return creator(options)

// ListFormats returns a list of supported formats
}
func (f *FormatterFactory) ListFormats() []ReportFormat {
	f.mu.RLock()
	defer f.mu.RUnlock()
	formats := make([]ReportFormat, 0, len(f.formatters))
	for format := range f.formatters {
		formats = append(formats, format)
	}
}
