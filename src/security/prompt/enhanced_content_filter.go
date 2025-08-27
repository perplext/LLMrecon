package prompt

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// EnhancedContentFilter represents an enhanced content filtering system
type EnhancedContentFilter struct {
	patterns map[string]*regexp.Regexp
	mu       sync.RWMutex
	enabled  bool
	config   FilterConfig
}

// FilterConfig holds configuration for content filtering
type FilterConfig struct {
	MaxContentLength int
	Timeout          time.Duration
	StrictMode       bool
	AllowedPatterns  []string
	BlockedPatterns  []string
}

// FilterResult represents the result of content filtering
type FilterResult struct {
	Allowed     bool
	Reason      string
	Score       float64
	Violations  []string
	Detections  []*Detection
	RiskScore   float64
	ActionTaken ActionType
	ProcessTime time.Duration
}

// NewEnhancedContentFilter creates a new enhanced content filter
func NewEnhancedContentFilter(config FilterConfig) *EnhancedContentFilter {
	cf := &EnhancedContentFilter{
		patterns: make(map[string]*regexp.Regexp),
		enabled:  true,
		config:   config,
	}

	// Initialize patterns
	cf.initializePatterns()

	return cf
}

// initializePatterns compiles regex patterns for filtering
func (cf *EnhancedContentFilter) initializePatterns() {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	// Compile blocked patterns
	for _, pattern := range cf.config.BlockedPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			cf.patterns["blocked_"+pattern] = regex
		}
	}

	// Compile allowed patterns
	for _, pattern := range cf.config.AllowedPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			cf.patterns["allowed_"+pattern] = regex
		}
	}
}

// FilterContent filters the given content and returns the result
func (cf *EnhancedContentFilter) FilterContent(ctx context.Context, content string) (*FilterResult, error) {
	if !cf.enabled {
		return &FilterResult{Allowed: true, Reason: "filtering disabled"}, nil
	}

	startTime := time.Now()

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, cf.config.Timeout)
	defer cancel()

	result := &FilterResult{
		Allowed:     true,
		Violations:  make([]string, 0),
		ProcessTime: 0,
	}

	// Check content length
	if len(content) > cf.config.MaxContentLength {
		result.Allowed = false
		result.Reason = "content too long"
		result.ProcessTime = time.Since(startTime)
		return result, nil
	}

	// Run filtering in goroutine with timeout
	done := make(chan bool, 1)
	go func() {
		cf.performFiltering(content, result)
		done <- true
	}()

	select {
	case <-done:
		result.ProcessTime = time.Since(startTime)
		return result, nil
	case <-timeoutCtx.Done():
		result.Allowed = false
		result.Reason = "filtering timeout"
		result.ProcessTime = time.Since(startTime)
		return result, timeoutCtx.Err()
	}
}

// performFiltering performs the actual content filtering
func (cf *EnhancedContentFilter) performFiltering(content string, result *FilterResult) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	// Check against blocked patterns
	for name, pattern := range cf.patterns {
		if strings.HasPrefix(name, "blocked_") {
			if pattern.MatchString(content) {
				result.Allowed = false
				result.Violations = append(result.Violations, name)
				result.Reason = fmt.Sprintf("matched blocked pattern: %s", name)

				if cf.config.StrictMode {
					return // Stop at first violation in strict mode
				}
			}
		}
	}

	// Calculate risk score
	result.Score = cf.calculateRiskScore(content)
}

// calculateRiskScore calculates a risk score for the content
func (cf *EnhancedContentFilter) calculateRiskScore(content string) float64 {
	score := 0.0

	// Basic scoring logic
	if len(content) > 1000 {
		score += 0.1
	}

	// Check for suspicious patterns
	suspiciousWords := []string{"inject", "bypass", "exploit", "hack"}
	for _, word := range suspiciousWords {
		if strings.Contains(strings.ToLower(content), word) {
			score += 0.2
		}
	}

	return score
}

// UpdatePatterns updates the filtering patterns
func (cf *EnhancedContentFilter) UpdatePatterns(allowedPatterns, blockedPatterns []string) error {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	// Clear existing patterns
	cf.patterns = make(map[string]*regexp.Regexp)

	// Update config
	cf.config.AllowedPatterns = allowedPatterns
	cf.config.BlockedPatterns = blockedPatterns

	// Reinitialize patterns
	cf.initializePatterns()

	return nil
}

// Enable enables the content filter
func (cf *EnhancedContentFilter) Enable() {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.enabled = true
}

// Disable disables the content filter
func (cf *EnhancedContentFilter) Disable() {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.enabled = false
}

// IsEnabled returns whether the filter is enabled
func (cf *EnhancedContentFilter) IsEnabled() bool {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.enabled
}

// GetConfig returns a copy of the current configuration
func (cf *EnhancedContentFilter) GetConfig() FilterConfig {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.config
}
