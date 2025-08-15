package optimization

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
)

// ContextOptimizer optimizes context variable usage in templates
// to reduce memory footprint and improve performance
type ContextOptimizer struct {
	// enableDeduplication enables variable deduplication
	enableDeduplication bool
	// enableLazyLoading enables lazy loading of variables
	enableLazyLoading bool
	// enableCompression enables variable compression
	enableCompression bool
	// sharedVariables contains variables shared across templates
	sharedVariables map[string]string
	// variableUsageCount tracks variable usage across templates
	variableUsageCount map[string]int
	// mutex protects shared data
	mutex sync.RWMutex

// ContextOptimizerOptions contains options for the context optimizer
type ContextOptimizerOptions struct {
	// EnableDeduplication enables variable deduplication
	EnableDeduplication bool
	// EnableLazyLoading enables lazy loading of variables
	EnableLazyLoading bool
	// EnableCompression enables variable compression
	EnableCompression bool

// DefaultContextOptimizerOptions returns default options for the context optimizer
func DefaultContextOptimizerOptions() *ContextOptimizerOptions {
	return &ContextOptimizerOptions{
		EnableDeduplication: true,
		EnableLazyLoading:   true,
		EnableCompression:   false,
	}

// NewContextOptimizer creates a new context optimizer
func NewContextOptimizer(options *ContextOptimizerOptions) *ContextOptimizer {
	if options == nil {
		options = DefaultContextOptimizerOptions()
	}

	return &ContextOptimizer{
		enableDeduplication: options.EnableDeduplication,
		enableLazyLoading:   options.EnableLazyLoading,
		enableCompression:   options.EnableCompression,
		sharedVariables:     make(map[string]string),
		variableUsageCount:  make(map[string]int),
	}

// OptimizeTemplate optimizes context variable usage in a template
func (o *ContextOptimizer) OptimizeTemplate(ctx context.Context, template *format.Template) (*format.Template, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	// Create a copy of the template
	optimizedTemplate := template.Clone()

	// Optimize variables
	if err := o.optimizeVariables(optimizedTemplate); err != nil {
		return nil, err
	}

	return optimizedTemplate, nil

// optimizeVariables optimizes variables in a template
func (o *ContextOptimizer) optimizeVariables(template *format.Template) error {
	// Skip if template has no variables
	if template.Variables == nil || len(template.Variables) == 0 {
		return nil
	}

	// Deduplicate variables if enabled
	if o.enableDeduplication {
		o.deduplicateVariables(template)
	}

	// Track variable usage
	o.trackVariableUsage(template)

	return nil

// deduplicateVariables deduplicates variables by sharing common values
func (o *ContextOptimizer) deduplicateVariables(template *format.Template) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Create a new variables map
	newVariables := make(map[string]interface{})

	// Process each variable
	for key, value := range template.Variables {
		// Check if value is already shared (only for string values)
		shared := false
		if strValue, ok := value.(string); ok {
			for _, sharedValue := range o.sharedVariables {
				if strValue == sharedValue {
					// Use shared value
					newVariables[key] = sharedValue
					shared = true
					break
				}
			}
		}

		if !shared {
			// Add to shared variables if it's a string
			if strValue, ok := value.(string); ok {
				o.sharedVariables[key] = strValue
			}
			newVariables[key] = value
		}
	}

	// Update template variables
	template.Variables = newVariables

// trackVariableUsage tracks variable usage across templates
func (o *ContextOptimizer) trackVariableUsage(template *format.Template) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Update variable usage count
	for key := range template.Variables {
		o.variableUsageCount[key]++
	}

// OptimizeTemplates optimizes context variable usage in multiple templates
func (o *ContextOptimizer) OptimizeTemplates(ctx context.Context, templates []*format.Template) ([]*format.Template, error) {
	// First pass: analyze variable usage across all templates
	o.analyzeVariableUsage(templates)

	// Second pass: optimize each template
	optimizedTemplates := make([]*format.Template, len(templates))
	var wg sync.WaitGroup
	errChan := make(chan error, len(templates))

	for i, template := range templates {
		wg.Add(1)
		go func(i int, template *format.Template) {
			defer wg.Done()
			optimizedTemplate, err := o.OptimizeTemplate(ctx, template)
			if err != nil {
				errChan <- err
				return
			}
			optimizedTemplates[i] = optimizedTemplate
		}(i, template)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return optimizedTemplates, nil

// analyzeVariableUsage analyzes variable usage across all templates
func (o *ContextOptimizer) analyzeVariableUsage(templates []*format.Template) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Reset variable usage count
	o.variableUsageCount = make(map[string]int)

	// Count variable usage across all templates
	for _, template := range templates {
		if template.Variables == nil {
			continue
		}

		for key := range template.Variables {
			o.variableUsageCount[key]++
		}
	}

// GetSharedVariables returns the shared variables
func (o *ContextOptimizer) GetSharedVariables() map[string]string {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Create a copy of shared variables
	sharedVars := make(map[string]string)
	for key, value := range o.sharedVariables {
		sharedVars[key] = value
	}

	return sharedVars

// GetVariableUsageCount returns the variable usage count
func (o *ContextOptimizer) GetVariableUsageCount() map[string]int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Create a copy of variable usage count
	usageCount := make(map[string]int)
	for key, count := range o.variableUsageCount {
		usageCount[key] = count
	}

	return usageCount

// GetHighUsageVariables returns variables with high usage
func (o *ContextOptimizer) GetHighUsageVariables(threshold int) map[string]int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Find variables with usage count above threshold
	highUsage := make(map[string]int)
	for key, count := range o.variableUsageCount {
		if count >= threshold {
			highUsage[key] = count
		}
	}

	return highUsage

// GetLowUsageVariables returns variables with low usage
func (o *ContextOptimizer) GetLowUsageVariables(threshold int) map[string]int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Find variables with usage count below threshold
	lowUsage := make(map[string]int)
	for key, count := range o.variableUsageCount {
		if count <= threshold {
			lowUsage[key] = count
		}
	}

	return lowUsage

// ClearSharedVariables clears the shared variables
func (o *ContextOptimizer) ClearSharedVariables() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.sharedVariables = make(map[string]string)

// SetEnableDeduplication sets whether variable deduplication is enabled
func (o *ContextOptimizer) SetEnableDeduplication(enable bool) {
	o.enableDeduplication = enable

// GetEnableDeduplication returns whether variable deduplication is enabled
func (o *ContextOptimizer) GetEnableDeduplication() bool {
	return o.enableDeduplication

// SetEnableLazyLoading sets whether lazy loading of variables is enabled
func (o *ContextOptimizer) SetEnableLazyLoading(enable bool) {
	o.enableLazyLoading = enable

// GetEnableLazyLoading returns whether lazy loading of variables is enabled
func (o *ContextOptimizer) GetEnableLazyLoading() bool {
	return o.enableLazyLoading

// SetEnableCompression sets whether variable compression is enabled
func (o *ContextOptimizer) SetEnableCompression(enable bool) {
	o.enableCompression = enable

// GetEnableCompression returns whether variable compression is enabled
func (o *ContextOptimizer) GetEnableCompression() bool {
	return o.enableCompression
