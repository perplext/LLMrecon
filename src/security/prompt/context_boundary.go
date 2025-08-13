// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"math"
	"regexp"
	"strings"
)

// ContextBoundaryEnforcer enforces context boundaries to prevent prompt injection
type ContextBoundaryEnforcer struct {
	config                *ProtectionConfig
	boundaryConfig        *ContextBoundaryConfig
	sanitizationPatterns  map[int][]*regexp.Regexp
	normalizationPatterns []*regexp.Regexp
}

// NewContextBoundaryEnforcer creates a new context boundary enforcer
func NewContextBoundaryEnforcer(config *ProtectionConfig) *ContextBoundaryEnforcer {
	// Create default boundary config if not specified
	boundaryConfig := &ContextBoundaryConfig{
		EnableTokenization: true,
		EnableSanitization: true,
		EnableNormalization: true,
		MaxPromptLength:    config.MaxPromptLength,
		SanitizationLevel:  config.SanitizationLevel,
	}

	// Initialize sanitization patterns for different levels
	sanitizationPatterns := make(map[int][]*regexp.Regexp)
	
	// Level 1 (Basic) sanitization patterns
	sanitizationPatterns[1] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)system\s*:\s*`),
		regexp.MustCompile(`(?i)<\s*system\s*>\s*([^<]*)<\s*/\s*system\s*>`),
		regexp.MustCompile(`(?i)\[\s*system\s*\]\s*([^\[]*)\[\s*/\s*system\s*\]`),
		// Add triple backticks pattern to level 1 for high priority detection
		regexp.MustCompile(`(?i)\x60{3}\s*system\b`),
	}
	
	// Level 2 (Medium) sanitization patterns - includes Level 1 plus additional patterns
	sanitizationPatterns[2] = append(sanitizationPatterns[1], []*regexp.Regexp{
		regexp.MustCompile(`(?i)ignore\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)disregard\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)forget\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)you\s+are\s+now\s+(a|an)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)act\s+as\s+(a|an)\s+([a-zA-Z\s]+)`),
	}...)
	
	// Level 3 (High) sanitization patterns - includes Level 2 plus additional patterns
	sanitizationPatterns[3] = append(sanitizationPatterns[2], []*regexp.Regexp{
		regexp.MustCompile(`(?i)(` + "`" + `{3}|'''|""")`),
		regexp.MustCompile(`(?i)<\s*[a-zA-Z]+\s*>`),
		regexp.MustCompile(`(?i)\[\s*[a-zA-Z]+\s*\]`),
		regexp.MustCompile(`(?i)\{\s*[a-zA-Z]+\s*\}`),
		regexp.MustCompile(`(?i)#\s*[a-zA-Z]+\s*#`),
		regexp.MustCompile(`(?i)pretend\s+to\s+be\s+(a|an)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)from\s+now\s+on\s+you\s+are\s+(a|an)\s+([a-zA-Z\s]+)`),
	}...)

	// Normalization patterns
	normalizationPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\s+`),                    // Multiple whitespace
		regexp.MustCompile(`[^\x20-\x7E\s]`),         // Non-printable ASCII characters
		regexp.MustCompile(`(?i)\\[nrt]`),            // Escape sequences
		regexp.MustCompile(`(?i)\\u[0-9a-f]{4}`),     // Unicode escape sequences
		regexp.MustCompile(`(?i)\\x[0-9a-f]{2}`),     // Hex escape sequences
	}

	return &ContextBoundaryEnforcer{
		config:                config,
		boundaryConfig:        boundaryConfig,
		sanitizationPatterns:  sanitizationPatterns,
		normalizationPatterns: normalizationPatterns,
	}
}

// EnforceBoundaries enforces context boundaries on a prompt
func (e *ContextBoundaryEnforcer) EnforceBoundaries(ctx context.Context, prompt string) (string, *ProtectionResult, error) {
	startTime := time.Now()
	
	result := &ProtectionResult{
		OriginalPrompt:   prompt,
		ProtectedPrompt:  prompt,
		Detections:       make([]*Detection, 0),
		RiskScore:        0.0,
		ActionTaken:      ActionNone,
		Timestamp:        startTime,
	}

	// Check prompt length
	if len(prompt) > e.boundaryConfig.MaxPromptLength {
		result.ProtectedPrompt = prompt[:e.boundaryConfig.MaxPromptLength]
		result.Detections = append(result.Detections, &Detection{
			Type:        DetectionTypeBoundaryViolation,
			Confidence:  1.0,
			Description: "Prompt exceeds maximum allowed length",
			Location: &DetectionLocation{
				Start:   e.boundaryConfig.MaxPromptLength,
				End:     len(prompt),
				Context: prompt[maxInt(0, e.boundaryConfig.MaxPromptLength-20):e.boundaryConfig.MaxPromptLength] + "...",
			},
			Remediation: "Truncate the prompt to the maximum allowed length",
		})
		result.RiskScore = 0.5
		result.ActionTaken = ActionModified
		prompt = result.ProtectedPrompt
	}

	// Apply normalization if enabled
	if e.boundaryConfig.EnableNormalization {
		normalizedPrompt, normalizationDetections := e.normalizePrompt(prompt)
		if normalizedPrompt != prompt {
			result.ProtectedPrompt = normalizedPrompt
			result.Detections = append(result.Detections, normalizationDetections...)
			result.RiskScore = math.Max(result.RiskScore, 0.3) // Normalization is a lower risk
			result.ActionTaken = ActionModified
			prompt = normalizedPrompt
		}
	}

	// Apply sanitization if enabled
	if e.boundaryConfig.EnableSanitization {
		sanitizedPrompt, sanitizationDetections := e.sanitizePrompt(prompt)
		if sanitizedPrompt != prompt {
			result.ProtectedPrompt = sanitizedPrompt
			result.Detections = append(result.Detections, sanitizationDetections...)
			
			// Check for specific high-risk patterns that should be blocked
			for _, detection := range sanitizationDetections {
				if detection.Type == DetectionTypeBoundaryViolation {
					// Always block boundary violations with maximum risk score
					result.RiskScore = 1.0
					result.ActionTaken = ActionBlocked
					result.ProtectedPrompt = ""
					return result.ProtectedPrompt, result, nil
				}
			}
			
			result.RiskScore = math.Max(result.RiskScore, 0.7) // Sanitization indicates higher risk
			result.ActionTaken = ActionModified
			prompt = sanitizedPrompt
		}
	}

	// Apply tokenization if enabled (for future implementation)
	if e.boundaryConfig.EnableTokenization {
		// Tokenization would be implemented here
		// This is a placeholder for future implementation
	}

	result.ProcessingTime = time.Since(startTime)
	return result.ProtectedPrompt, result, nil
}

// normalizePrompt normalizes a prompt by removing or replacing problematic characters
func (e *ContextBoundaryEnforcer) normalizePrompt(prompt string) (string, []*Detection) {
	detections := make([]*Detection, 0)
	normalizedPrompt := prompt
	
	// Replace multiple whitespace with a single space
	multipleWhitespacePattern := e.normalizationPatterns[0]
	if multipleWhitespacePattern.MatchString(normalizedPrompt) {
		normalizedPrompt = multipleWhitespacePattern.ReplaceAllString(normalizedPrompt, " ")
	}
	
	// Remove non-printable ASCII characters
	nonPrintablePattern := e.normalizationPatterns[1]
	if nonPrintablePattern.MatchString(normalizedPrompt) {
		matches := nonPrintablePattern.FindAllStringIndex(prompt, -1)
		if len(matches) > 0 {
			detections = append(detections, &Detection{
				Type:        DetectionTypeBoundaryViolation,
				Confidence:  0.8,
				Description: "Non-printable characters detected and removed",
				Location: &DetectionLocation{
					Start:   matches[0][0],
					End:     matches[0][1],
					Context: getContext(prompt, matches[0][0], matches[0][1]),
				},
				Remediation: "Remove non-printable characters from the prompt",
			})
		}
		normalizedPrompt = nonPrintablePattern.ReplaceAllString(normalizedPrompt, "")
	}
	
	// Replace escape sequences
	escapeSequencePattern := e.normalizationPatterns[2]
	if escapeSequencePattern.MatchString(normalizedPrompt) {
		matches := escapeSequencePattern.FindAllStringIndex(prompt, -1)
		if len(matches) > 0 {
			detections = append(detections, &Detection{
				Type:        DetectionTypeBoundaryViolation,
				Confidence:  0.7,
				Description: "Escape sequences detected and normalized",
				Location: &DetectionLocation{
					Start:   matches[0][0],
					End:     matches[0][1],
					Context: getContext(prompt, matches[0][0], matches[0][1]),
				},
				Remediation: "Remove escape sequences from the prompt",
			})
		}
		
		// Replace common escape sequences with their actual characters
		normalizedPrompt = strings.ReplaceAll(normalizedPrompt, "\\n", " ")
		normalizedPrompt = strings.ReplaceAll(normalizedPrompt, "\\r", " ")
		normalizedPrompt = strings.ReplaceAll(normalizedPrompt, "\\t", " ")
	}
	
	// Replace Unicode escape sequences
	unicodeEscapePattern := e.normalizationPatterns[3]
	if unicodeEscapePattern.MatchString(normalizedPrompt) {
		matches := unicodeEscapePattern.FindAllStringIndex(prompt, -1)
		if len(matches) > 0 {
			detections = append(detections, &Detection{
				Type:        DetectionTypeBoundaryViolation,
				Confidence:  0.8,
				Description: "Unicode escape sequences detected and removed",
				Location: &DetectionLocation{
					Start:   matches[0][0],
					End:     matches[0][1],
					Context: getContext(prompt, matches[0][0], matches[0][1]),
				},
				Remediation: "Remove Unicode escape sequences from the prompt",
			})
		}
		normalizedPrompt = unicodeEscapePattern.ReplaceAllString(normalizedPrompt, "")
	}
	
	// Replace hex escape sequences
	hexEscapePattern := e.normalizationPatterns[4]
	if hexEscapePattern.MatchString(normalizedPrompt) {
		matches := hexEscapePattern.FindAllStringIndex(prompt, -1)
		if len(matches) > 0 {
			detections = append(detections, &Detection{
				Type:        DetectionTypeBoundaryViolation,
				Confidence:  0.8,
				Description: "Hex escape sequences detected and removed",
				Location: &DetectionLocation{
					Start:   matches[0][0],
					End:     matches[0][1],
					Context: getContext(prompt, matches[0][0], matches[0][1]),
				},
				Remediation: "Remove hex escape sequences from the prompt",
			})
		}
		normalizedPrompt = hexEscapePattern.ReplaceAllString(normalizedPrompt, "")
	}
	
	return normalizedPrompt, detections
}

// sanitizePrompt sanitizes a prompt by removing or replacing potential injection patterns
func (e *ContextBoundaryEnforcer) sanitizePrompt(prompt string) (string, []*Detection) {
	detections := make([]*Detection, 0)
	sanitizedPrompt := prompt
	
	// Get sanitization patterns for the configured level
	level := e.boundaryConfig.SanitizationLevel
	if level < 1 {
		level = 1
	} else if level > 3 {
		level = 3
	}
	
	patterns := e.sanitizationPatterns[level]
	
	// Apply each pattern
	for _, pattern := range patterns {
		if pattern.MatchString(sanitizedPrompt) {
			matches := pattern.FindAllStringIndex(sanitizedPrompt, -1)
			for _, match := range matches {
				startIndex := match[0]
				endIndex := match[1]
				matchedText := sanitizedPrompt[startIndex:endIndex]
				
				// Determine the detection type based on the pattern
				detectionType := DetectionTypePromptInjection
				confidence := 0.9
				
				// Check for specific patterns that indicate boundary violations
				if strings.Contains(pattern.String(), "system") || 
				   strings.Contains(pattern.String(), "\\x60{3}") || // Triple backticks
				   strings.Contains(pattern.String(), "'''") || 
				   strings.Contains(pattern.String(), "\"\"\"") {
					detectionType = DetectionTypeBoundaryViolation
					confidence = 1.0 // Increase confidence to 1.0 for boundary violations
				}
				
				detections = append(detections, &Detection{
					Type:        detectionType,
					Confidence:  confidence,
					Description: "Potential prompt injection pattern detected and removed: " + matchedText,
					Location: &DetectionLocation{
						Start:   startIndex,
						End:     endIndex,
						Context: getContext(sanitizedPrompt, startIndex, endIndex),
					},
					Pattern:     pattern.String(),
					Remediation: "Remove or sanitize the prompt injection pattern",
				})
			}
			
			// Replace the pattern with a safe placeholder
			sanitizedPrompt = pattern.ReplaceAllString(sanitizedPrompt, "[FILTERED]")
		}
	}
	
	return sanitizedPrompt, detections
}

// getContext extracts context around a match
func getContext(text string, start, end int) string {
	contextStart := maxInt(0, start-20)
	contextEnd := minInt(len(text), end+20)
	return text[contextStart:contextEnd]
}

// Helper function to get the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to get the maximum of two integers
func maxInt(a, b int) int {
	if a < b {
		return b
	}
	return a
}
