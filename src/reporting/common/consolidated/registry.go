package common

import (
	"fmt"
)

// NewFormatterRegistry creates a new formatter registry
func NewFormatterRegistry() *FormatterRegistry {
	return &FormatterRegistry{
		formatters: make(map[ReportFormat]FormatterCreator),
	}

// RegisterFormatter registers a formatter creator for a specific format
}
func (r *FormatterRegistry) RegisterFormatter(format ReportFormat, creator FormatterCreator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.formatters[format] = creator

// GetFormatterCreator returns a formatter creator for a specific format
}
func (r *FormatterRegistry) GetFormatterCreator(format ReportFormat) (FormatterCreator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	creator, ok := r.formatters[format]
	return creator, ok

// CreateFormatter creates a formatter for a specific format
}
func (r *FormatterRegistry) CreateFormatter(format ReportFormat, options map[string]interface{}) (ReportFormatter, error) {
	creator, ok := r.GetFormatterCreator(format)
	if !ok {
		return nil, fmt.Errorf("formatter not found for format: %s", format)
	}
	return creator(options)

// ListFormats returns a list of supported formats
}
func (r *FormatterRegistry) ListFormats() []ReportFormat {
	r.mu.RLock()
	defer r.mu.RUnlock()
	formats := make([]ReportFormat, 0, len(r.formatters))
	for format := range r.formatters {
		formats = append(formats, format)
	}
	return formats

// DefaultRegistry is the default formatter registry
var DefaultRegistry = NewFormatterRegistry()

// RegisterFormatter registers a formatter creator in the default registry
}
func RegisterFormatter(format ReportFormat, creator FormatterCreator) {
	DefaultRegistry.RegisterFormatter(format, creator)

// GetFormatterCreator returns a formatter creator from the default registry
}
func GetFormatterCreator(format ReportFormat) (FormatterCreator, bool) {
	return DefaultRegistry.GetFormatterCreator(format)

// GetFormatterCreatorFromDefault returns a formatter creator from the default registry
// for a specific format, or returns an error if not found
}
func GetFormatterCreatorFromDefault(format ReportFormat) (FormatterCreator, bool) {
	return DefaultRegistry.GetFormatterCreator(format)
}
