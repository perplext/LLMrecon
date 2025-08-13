package execution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
)

// Engine is a simplified template execution engine for benchmarking
type Engine struct {
	// maxConcurrent is the maximum number of concurrent executions
	maxConcurrent int
	// executionDelay is the simulated execution delay
	executionDelay time.Duration
}

// NewEngine creates a new execution engine
func NewEngine() *Engine {
	return &Engine{
		maxConcurrent:  10,
		executionDelay: 10 * time.Millisecond,
	}
}

// ExecuteTemplate executes a template
func (e *Engine) ExecuteTemplate(ctx context.Context, template *format.Template, data interface{}) (string, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Continue
	}

	// Simulate execution delay
	select {
	case <-time.After(e.executionDelay):
		// Continue
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Simple template rendering
	result, err := e.renderTemplate(template, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return result, nil
}

// ExecuteTemplates executes multiple templates
func (e *Engine) ExecuteTemplates(ctx context.Context, templates []*format.Template, data interface{}) ([]string, error) {
	results := make([]string, len(templates))
	
	for i, template := range templates {
		result, err := e.ExecuteTemplate(ctx, template, data)
		if err != nil {
			return results, fmt.Errorf("failed to execute template %s: %w", template.ID, err)
		}
		
		results[i] = result
	}
	
	return results, nil
}

// renderTemplate renders a template with data
func (e *Engine) renderTemplate(template *format.Template, data interface{}) (string, error) {
	if template == nil || template.Content == nil {
		return "", fmt.Errorf("template or content is nil")
	}
	
	// Get template content
	content := template.Test.Prompt
	
	// Replace variables
	if data != nil {
		if dataMap, ok := data.(map[string]interface{}); ok {
			for key, value := range dataMap {
				if strValue, ok := value.(string); ok {
					placeholder := fmt.Sprintf("{{%s}}", key)
					content = strings.ReplaceAll(content, placeholder, strValue)
				}
			}
		}
	}
	
	return content, nil
}

// SetMaxConcurrent sets the maximum number of concurrent executions
func (e *Engine) SetMaxConcurrent(max int) {
	if max > 0 {
		e.maxConcurrent = max
	}
}

// SetExecutionDelay sets the simulated execution delay
func (e *Engine) SetExecutionDelay(delay time.Duration) {
	if delay > 0 {
		e.executionDelay = delay
	}
}

// GetMaxConcurrent returns the maximum number of concurrent executions
func (e *Engine) GetMaxConcurrent() int {
	return e.maxConcurrent
}

// GetExecutionDelay returns the simulated execution delay
func (e *Engine) GetExecutionDelay() time.Duration {
	return e.executionDelay
}
