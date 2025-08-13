// Package common provides common types and interfaces for the reporting system
package common

import (
	"fmt"
	"sync"
)

// FormatterCreator is a function that creates a formatter
type FormatterCreator func(options map[string]interface{}) (ReportFormatter, error)

// FormatterRegistry is a registry for formatter creators
type FormatterRegistry struct {
	formatters map[ReportFormat]FormatterCreator
	mu         sync.RWMutex
}

// NewFormatterRegistry creates a new formatter registry
func NewFormatterRegistry() *FormatterRegistry {
	return &FormatterRegistry{
		formatters: make(map[ReportFormat]FormatterCreator),
	}
}

// RegisterFormatter registers a formatter creator for a specific format
func (r *FormatterRegistry) RegisterFormatter(format ReportFormat, creator FormatterCreator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.formatters[format] = creator
}

// GetFormatterCreator returns the formatter creator for a specific format
func (r *FormatterRegistry) GetFormatterCreator(format ReportFormat) (FormatterCreator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	creator, ok := r.formatters[format]
	return creator, ok
}

// CreateFormatter creates a formatter for the specified format
func (r *FormatterRegistry) CreateFormatter(format ReportFormat, options map[string]interface{}) (ReportFormatter, error) {
	creator, ok := r.GetFormatterCreator(format)
	if !ok {
		return nil, fmt.Errorf("unsupported report format: %s", format)
	}
	
	return creator(options)
}

// DefaultRegistry is the global formatter registry
var DefaultRegistry = NewFormatterRegistry()

// RegisterFormatter registers a formatter creator with the default registry
func RegisterFormatter(format ReportFormat, creator FormatterCreator) {
	DefaultRegistry.RegisterFormatter(format, creator)
}

// GetFormatterCreator returns a formatter creator from the default registry
func GetFormatterCreator(format ReportFormat) (FormatterCreator, bool) {
	return DefaultRegistry.GetFormatterCreator(format)
}

// CreateFormatter creates a formatter using the default registry
func CreateFormatter(format ReportFormat, options map[string]interface{}) (ReportFormatter, error) {
	return DefaultRegistry.CreateFormatter(format, options)
}
