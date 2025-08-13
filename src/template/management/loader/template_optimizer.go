// Package loader provides functionality for loading templates from various sources.
package loader

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateOptimizer provides functionality for optimizing templates
type TemplateOptimizer struct {
	// minifyEnabled determines whether to minify template content
	minifyEnabled bool
	// compressEnabled determines whether to compress template content
	compressEnabled bool
	// mutex protects the optimizer
	mutex sync.RWMutex
	// optimizationStats tracks optimization statistics
	optimizationStats OptimizerStats
}

// OptimizerStats tracks optimizer statistics
type OptimizerStats struct {
	// TotalOptimizations is the total number of template optimizations
	TotalOptimizations int64
	// TotalBytesOriginal is the total size of templates before optimization
	TotalBytesOriginal int64
	// TotalBytesOptimized is the total size of templates after optimization
	TotalBytesOptimized int64
	// CompressionRatio is the average compression ratio
	CompressionRatio float64
}

// NewTemplateOptimizer creates a new template optimizer
func NewTemplateOptimizer(minifyEnabled, compressEnabled bool) *TemplateOptimizer {
	return &TemplateOptimizer{
		minifyEnabled:    minifyEnabled,
		compressEnabled:  compressEnabled,
		optimizationStats: OptimizerStats{},
	}
}

// OptimizeTemplate optimizes a template by applying various optimizations
func (o *TemplateOptimizer) OptimizeTemplate(template *format.Template) (*format.Template, error) {
	if template == nil {
		return nil, fmt.Errorf("cannot optimize nil template")
	}

	// Create a deep copy of the template to avoid modifying the original
	optimizedTemplate := o.cloneTemplate(template)

	// Optimize prompt content
	originalSize := o.estimateTemplateSize(optimizedTemplate)
	
	// Apply optimizations
	if o.minifyEnabled {
		o.minifyPrompt(&optimizedTemplate.Test.Prompt)
		o.minifyExpectedBehavior(&optimizedTemplate.Test.ExpectedBehavior)
		
		// Optimize variations
		for i := range optimizedTemplate.Test.Variations {
			o.minifyPrompt(&optimizedTemplate.Test.Variations[i].Prompt)
		}
	}
	
	// Update optimization statistics
	optimizedSize := o.estimateTemplateSize(optimizedTemplate)
	o.mutex.Lock()
	o.optimizationStats.TotalOptimizations++
	o.optimizationStats.TotalBytesOriginal += int64(originalSize)
	o.optimizationStats.TotalBytesOptimized += int64(optimizedSize)
	if o.optimizationStats.TotalBytesOriginal > 0 {
		o.optimizationStats.CompressionRatio = float64(o.optimizationStats.TotalBytesOptimized) / float64(o.optimizationStats.TotalBytesOriginal)
	}
	o.mutex.Unlock()

	return optimizedTemplate, nil
}

// OptimizeTemplates optimizes multiple templates
func (o *TemplateOptimizer) OptimizeTemplates(templates []*format.Template) ([]*format.Template, error) {
	if templates == nil {
		return nil, fmt.Errorf("cannot optimize nil templates")
	}

	optimizedTemplates := make([]*format.Template, len(templates))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(templates))

	for i, template := range templates {
		wg.Add(1)
		go func(idx int, tmpl *format.Template) {
			defer wg.Done()

			optimizedTemplate, err := o.OptimizeTemplate(tmpl)
			if err != nil {
				errorsChan <- err
				return
			}

			mu.Lock()
			optimizedTemplates[idx] = optimizedTemplate
			mu.Unlock()
		}(i, template)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	if len(errorsChan) > 0 {
		return nil, <-errorsChan
	}

	return optimizedTemplates, nil
}

// minifyPrompt removes unnecessary whitespace from prompt content
func (o *TemplateOptimizer) minifyPrompt(prompt *string) {
	if prompt == nil || *prompt == "" {
		return
	}

	// Preserve markdown and code blocks
	lines := strings.Split(*prompt, "\n")
	var inCodeBlock bool
	var result strings.Builder

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Check for code block markers
		if strings.HasPrefix(trimmedLine, "```") {
			inCodeBlock = !inCodeBlock
			result.WriteString(trimmedLine)
			result.WriteString("\n")
			continue
		}

		// Preserve code blocks
		if inCodeBlock {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Optimize normal text
		if trimmedLine == "" {
			// Collapse multiple empty lines into one
			if i > 0 && strings.TrimSpace(lines[i-1]) == "" {
				continue
			}
			result.WriteString("\n")
		} else {
			result.WriteString(trimmedLine)
			result.WriteString("\n")
		}
	}

	*prompt = strings.TrimSpace(result.String())
}

// minifyExpectedBehavior optimizes expected behavior content
func (o *TemplateOptimizer) minifyExpectedBehavior(expectedBehavior *string) {
	o.minifyPrompt(expectedBehavior)
}

// compressContent compresses content using gzip
func (o *TemplateOptimizer) compressContent(content string) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	
	_, err := gzipWriter.Write([]byte(content))
	if err != nil {
		return nil, err
	}
	
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// decompressContent decompresses gzipped content
func (o *TemplateOptimizer) decompressContent(compressed []byte) (string, error) {
	buf := bytes.NewBuffer(compressed)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()
	
	decompressed, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return "", err
	}
	
	return string(decompressed), nil
}

// cloneTemplate creates a deep copy of a template
func (o *TemplateOptimizer) cloneTemplate(template *format.Template) *format.Template {
	// Serialize to JSON and back for a deep copy
	data, _ := json.Marshal(template)
	var clone format.Template
	_ = json.Unmarshal(data, &clone)
	return &clone
}

// estimateTemplateSize estimates the size of a template in bytes
func (o *TemplateOptimizer) estimateTemplateSize(template *format.Template) int {
	data, _ := json.Marshal(template)
	return len(data)
}

// GetOptimizationStats returns statistics about the optimizer
func (o *TemplateOptimizer) GetOptimizationStats() map[string]interface{} {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	
	return map[string]interface{}{
		"total_optimizations":    o.optimizationStats.TotalOptimizations,
		"total_bytes_original":   o.optimizationStats.TotalBytesOriginal,
		"total_bytes_optimized":  o.optimizationStats.TotalBytesOptimized,
		"compression_ratio":      o.optimizationStats.CompressionRatio,
		"minify_enabled":         o.minifyEnabled,
		"compress_enabled":       o.compressEnabled,
	}
}

// SetMinifyEnabled enables or disables minification
func (o *TemplateOptimizer) SetMinifyEnabled(enabled bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.minifyEnabled = enabled
}

// SetCompressEnabled enables or disables compression
func (o *TemplateOptimizer) SetCompressEnabled(enabled bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.compressEnabled = enabled
}
