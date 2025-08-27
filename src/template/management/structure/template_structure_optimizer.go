// Package structure provides functionality for optimizing template structure.
package structure

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
)

// TemplateStructureOptimizer provides functionality for optimizing template structure
type TemplateStructureOptimizer struct {
	// mutex protects the optimizer
	mutex sync.RWMutex
	// optimizationStats tracks optimization statistics
	optimizationStats OptimizationStats
	// optimizationRules contains the rules for optimizing templates
	optimizationRules []OptimizationRule
}

// OptimizationStats tracks optimization statistics
type OptimizationStats struct {
	// TotalOptimizations is the total number of template optimizations
	TotalOptimizations int64
	// OptimizationsByRule tracks optimizations by rule
	OptimizationsByRule map[string]int64
}

// OptimizationRule represents a rule for optimizing templates
type OptimizationRule struct {
	// Name is the name of the rule
	Name string
	// Description is the description of the rule
	Description string
	// ApplyFunc is the function to apply the rule
	ApplyFunc func(*format.Template) (bool, error)
}

// NewTemplateStructureOptimizer creates a new template structure optimizer
func NewTemplateStructureOptimizer() *TemplateStructureOptimizer {
	optimizer := &TemplateStructureOptimizer{
		optimizationStats: OptimizationStats{
			OptimizationsByRule: make(map[string]int64),
		},
	}

	// Initialize optimization rules
	optimizer.initializeRules()

	return optimizer
}

// initializeRules initializes the optimization rules
func (o *TemplateStructureOptimizer) initializeRules() {
	o.optimizationRules = []OptimizationRule{
		{
			Name:        "remove_redundant_whitespace",
			Description: "Removes redundant whitespace from template content",
			ApplyFunc:   o.removeRedundantWhitespace,
		},
		{
			Name:        "optimize_regex_patterns",
			Description: "Optimizes regex patterns in detection criteria",
			ApplyFunc:   o.optimizeRegexPatterns,
		},
		{
			Name:        "consolidate_variations",
			Description: "Consolidates similar test variations",
			ApplyFunc:   o.consolidateVariations,
		},
		{
			Name:        "optimize_detection_criteria",
			Description: "Optimizes detection criteria for faster matching",
			ApplyFunc:   o.optimizeDetectionCriteria,
		},
		{
			Name:        "normalize_template_structure",
			Description: "Normalizes template structure for consistent processing",
			ApplyFunc:   o.normalizeTemplateStructure,
		},
	}
}

// OptimizeTemplate optimizes a template by applying various structural optimizations
func (o *TemplateStructureOptimizer) OptimizeTemplate(template *format.Template) (*format.Template, error) {
	if template == nil {
		return nil, fmt.Errorf("cannot optimize nil template")
	}

	// Create a deep copy of the template to avoid modifying the original
	optimizedTemplate := o.cloneTemplate(template)

	// Apply each optimization rule
	for _, rule := range o.optimizationRules {
		applied, err := rule.ApplyFunc(optimizedTemplate)
		if err != nil {
			return nil, fmt.Errorf("error applying rule %s: %w", rule.Name, err)
		}

		if applied {
			o.mutex.Lock()
			o.optimizationStats.OptimizationsByRule[rule.Name]++
			o.mutex.Unlock()
		}
	}

	// Update optimization statistics
	o.mutex.Lock()
	o.optimizationStats.TotalOptimizations++
	o.mutex.Unlock()

	return optimizedTemplate, nil
}

// OptimizeTemplates optimizes multiple templates
func (o *TemplateStructureOptimizer) OptimizeTemplates(templates []*format.Template) ([]*format.Template, error) {
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

// removeRedundantWhitespace removes redundant whitespace from template content
func (o *TemplateStructureOptimizer) removeRedundantWhitespace(template *format.Template) (bool, error) {
	modified := false

	// Optimize prompt
	if template.Test.Prompt != "" {
		originalPrompt := template.Test.Prompt
		template.Test.Prompt = o.optimizeWhitespace(template.Test.Prompt)
		if originalPrompt != template.Test.Prompt {
			modified = true
		}
	}

	// Optimize expected behavior
	if template.Test.Expected != "" {
		originalExpected := template.Test.Expected
		template.Test.Expected = o.optimizeWhitespace(template.Test.Expected)
		if originalExpected != template.Test.Expected {
			modified = true
		}
	}

	// Optimize variations
	for i := range template.Test.Variations {
		if template.Test.Variations[i].Prompt != "" {
			originalPrompt := template.Test.Variations[i].Prompt
			template.Test.Variations[i].Prompt = o.optimizeWhitespace(template.Test.Variations[i].Prompt)
			if originalPrompt != template.Test.Variations[i].Prompt {
				modified = true
			}
		}
	}

	return modified, nil
}

// optimizeWhitespace optimizes whitespace in a string while preserving code blocks
func (o *TemplateStructureOptimizer) optimizeWhitespace(content string) string {
	// Preserve markdown and code blocks
	lines := strings.Split(content, "\n")
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

	return strings.TrimSpace(result.String())
}

// optimizeRegexPatterns optimizes regex patterns in detection criteria
func (o *TemplateStructureOptimizer) optimizeRegexPatterns(template *format.Template) (bool, error) {
	modified := false

	// Optimize main detection criteria
	if template.Test.Detection.Type == "regex_match" && template.Test.Detection.Pattern != "" {
		originalPattern := template.Test.Detection.Pattern
		optimizedPattern, err := o.optimizeRegex(originalPattern)
		if err != nil {
			return false, err
		}

		if originalPattern != optimizedPattern {
			template.Test.Detection.Pattern = optimizedPattern
			modified = true
		}
	}

	// Optimize variations
	for i := range template.Test.Variations {
		if template.Test.Variations[i].Detection.Type == "regex_match" &&
			template.Test.Variations[i].Detection.Pattern != "" {
			originalPattern := template.Test.Variations[i].Detection.Pattern
			optimizedPattern, err := o.optimizeRegex(originalPattern)
			if err != nil {
				return false, err
			}

			if originalPattern != optimizedPattern {
				template.Test.Variations[i].Detection.Pattern = optimizedPattern
				modified = true
			}
		}
	}

	return modified, nil
}

// optimizeRegex optimizes a regex pattern for better performance
func (o *TemplateStructureOptimizer) optimizeRegex(pattern string) (string, error) {
	// Validate the pattern first
	_, err := regexp.Compile(pattern)
	if err != nil {
		return pattern, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Apply regex optimizations
	optimized := pattern

	// Optimize common inefficient patterns
	optimized = strings.ReplaceAll(optimized, ".*.*", ".*")
	optimized = strings.ReplaceAll(optimized, "[a-zA-Z0-9_]", "\\w")
	optimized = strings.ReplaceAll(optimized, "[0-9]", "\\d")

	// Add anchors if missing and appropriate
	if !strings.HasPrefix(optimized, "^") && !strings.HasPrefix(optimized, ".*") {
		// Only add anchor if it's meant to match from the beginning
		if strings.HasPrefix(pattern, "[") || strings.HasPrefix(pattern, "\\") {
			optimized = "^" + optimized
		}
	}

	// Validate the optimized pattern
	_, err = regexp.Compile(optimized)
	if err != nil {
		// If optimization broke the pattern, return the original
		return pattern, nil
	}

	return optimized, nil
}

// consolidateVariations consolidates similar test variations
func (o *TemplateStructureOptimizer) consolidateVariations(template *format.Template) (bool, error) {
	if len(template.Test.Variations) <= 1 {
		return false, nil
	}

	// Map to track variations by detection criteria
	variationMap := make(map[string][]int)

	// Group variations by detection criteria
	for i, variation := range template.Test.Variations {
		key := fmt.Sprintf("%s:%s:%s",
			variation.Detection.Type,
			variation.Detection.Pattern,
			variation.Detection.Match)
		variationMap[key] = append(variationMap[key], i)
	}

	// Check if any consolidation is possible
	consolidationPossible := false
	for _, indices := range variationMap {
		if len(indices) > 1 {
			consolidationPossible = true
			break
		}
	}

	if !consolidationPossible {
		return false, nil
	}

	// Perform consolidation
	newVariations := make([]*format.TemplateVariation, 0)
	for _, indices := range variationMap {
		if len(indices) == 1 {
			// Single variation, keep as is
			newVariations = append(newVariations, template.Test.Variations[indices[0]])
		} else {
			// Multiple variations with same detection criteria
			// Combine prompts with a delimiter
			combinedPrompt := ""
			for _, idx := range indices {
				if combinedPrompt != "" {
					combinedPrompt += "\n---\n"
				}
				combinedPrompt += template.Test.Variations[idx].Prompt
			}

			// Create consolidated variation
			original := template.Test.Variations[indices[0]]
			consolidated := &format.TemplateVariation{
				Prompt:   combinedPrompt,
				Expected: original.Expected,
			}
			if original.Detection != nil {
				consolidated.Detection = &format.TemplateDetection{
					Type:     original.Detection.Type,
					Pattern:  original.Detection.Pattern,
					Match:    original.Detection.Match,
					Criteria: original.Detection.Criteria,
				}
			}
			newVariations = append(newVariations, consolidated)
		}
	}

	// Update template if variations were consolidated
	if len(newVariations) < len(template.Test.Variations) {
		template.Test.Variations = newVariations
		return true, nil
	}

	return false, nil
}

// optimizeDetectionCriteria optimizes detection criteria for faster matching
func (o *TemplateStructureOptimizer) optimizeDetectionCriteria(template *format.Template) (bool, error) {
	modified := false

	// Optimize main detection criteria
	if template.Test.Detection.Type == "string_match" && template.Test.Detection.Match != "" {
		originalMatch := template.Test.Detection.Match
		optimizedMatch := o.optimizeStringMatch(originalMatch)

		if originalMatch != optimizedMatch {
			template.Test.Detection.Match = optimizedMatch
			modified = true
		}
	}

	// Optimize variations
	for i := range template.Test.Variations {
		if template.Test.Variations[i].Detection.Type == "string_match" &&
			template.Test.Variations[i].Detection.Match != "" {
			originalMatch := template.Test.Variations[i].Detection.Match
			optimizedMatch := o.optimizeStringMatch(originalMatch)

			if originalMatch != optimizedMatch {
				template.Test.Variations[i].Detection.Match = optimizedMatch
				modified = true
			}
		}
	}

	return modified, nil
}

// optimizeStringMatch optimizes a string match pattern
func (o *TemplateStructureOptimizer) optimizeStringMatch(match string) string {
	// Trim whitespace
	return strings.TrimSpace(match)
}

// normalizeTemplateStructure normalizes template structure for consistent processing
func (o *TemplateStructureOptimizer) normalizeTemplateStructure(template *format.Template) (bool, error) {
	modified := false

	// Ensure template has a valid ID
	if template.ID == "" {
		// This should be handled by validation, but we'll set a placeholder
		template.ID = "placeholder_id"
		modified = true
	}

	// Ensure template has compatibility information
	if template.Compatibility == nil {
		// Create default compatibility section
		template.Compatibility = &format.TemplateCompatibility{
			MinToolVersion: "1.0.0",
			MaxToolVersion: "",
		}
		modified = true
	}

	// Ensure detection criteria is valid
	if template.Test.Detection.Type == "" {
		template.Test.Detection.Type = "string_match"
		modified = true
	}

	return modified, nil
}

// cloneTemplate creates a deep copy of a template
func (o *TemplateStructureOptimizer) cloneTemplate(template *format.Template) *format.Template {
	clone := &format.Template{
		ID:      template.ID,
		Name:    template.Name,
		Version: template.Version,
		Path:    template.Path,
	}

	// Copy content
	if template.Content != nil {
		clone.Content = make([]byte, len(template.Content))
		copy(clone.Content, template.Content)
	}

	// Copy metadata
	if template.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range template.Metadata {
			clone.Metadata[k] = v
		}
	}

	// Copy variables
	if template.Variables != nil {
		clone.Variables = make(map[string]interface{})
		for k, v := range template.Variables {
			clone.Variables[k] = v
		}
	}

	// Copy info
	if template.Info != nil {
		clone.Info = &format.TemplateInfo{
			Name:        template.Info.Name,
			Version:     template.Info.Version,
			Description: template.Info.Description,
			Author:      template.Info.Author,
		}
		if template.Info.Tags != nil {
			clone.Info.Tags = make([]string, len(template.Info.Tags))
			copy(clone.Info.Tags, template.Info.Tags)
		}
	}

	// Copy test
	if template.Test != nil {
		clone.Test = &format.TemplateTest{
			Prompt:   template.Test.Prompt,
			Expected: template.Test.Expected,
		}

		// Copy detection
		if template.Test.Detection != nil {
			clone.Test.Detection = &format.TemplateDetection{
				Type:     template.Test.Detection.Type,
				Pattern:  template.Test.Detection.Pattern,
				Match:    template.Test.Detection.Match,
				Criteria: template.Test.Detection.Criteria,
			}
		}

		// Copy variations
		if template.Test.Variations != nil {
			clone.Test.Variations = make([]*format.TemplateVariation, len(template.Test.Variations))
			for i, variation := range template.Test.Variations {
				cloneVariation := &format.TemplateVariation{
					Prompt:   variation.Prompt,
					Expected: variation.Expected,
				}

				// Copy variation detection
				if variation.Detection != nil {
					cloneVariation.Detection = &format.TemplateDetection{
						Type:     variation.Detection.Type,
						Pattern:  variation.Detection.Pattern,
						Match:    variation.Detection.Match,
						Criteria: variation.Detection.Criteria,
					}
				}

				clone.Test.Variations[i] = cloneVariation
			}
		}
	}

	// Copy compatibility
	if template.Compatibility != nil {
		clone.Compatibility = &format.TemplateCompatibility{
			MinToolVersion: template.Compatibility.MinToolVersion,
			MaxToolVersion: template.Compatibility.MaxToolVersion,
		}
	}

	return clone
}

// GetOptimizationStats returns statistics about the optimizer
func (o *TemplateStructureOptimizer) GetOptimizationStats() map[string]interface{} {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	// Copy the map to avoid concurrent access issues
	ruleStats := make(map[string]int64)
	for rule, count := range o.optimizationStats.OptimizationsByRule {
		ruleStats[rule] = count
	}

	return map[string]interface{}{
		"total_optimizations":   o.optimizationStats.TotalOptimizations,
		"optimizations_by_rule": ruleStats,
	}
}
