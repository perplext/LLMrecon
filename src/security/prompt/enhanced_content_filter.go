// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// EnhancedContentFilter extends the ContentFilter with more sophisticated filtering capabilities
type EnhancedContentFilter struct {
	*ContentFilter
	config              *ProtectionConfig
	filterConfig        *EnhancedFilterConfig
	sensitivePatterns   map[string]*regexp.Regexp
	prohibitedPatterns  map[string]*regexp.Regexp
	categoryFilters     map[string]func(string) (bool, float64, string)
	replacementStrategies map[string]func(string, string) string
	filterStats         map[string]*FilterStats
	mu                  sync.RWMutex
}

// EnhancedFilterConfig defines the configuration for enhanced content filtering
type EnhancedFilterConfig struct {
	EnableProfanityFilter bool                   `json:"enable_profanity_filter"`
	EnablePIIFilter       bool                   `json:"enable_pii_filter"`
	EnableCodeFilter      bool                   `json:"enable_code_filter"`
	EnableURLFilter       bool                   `json:"enable_url_filter"`
	EnableJailbreakFilter bool                   `json:"enable_jailbreak_filter"`
	EnableSensitiveInfoFilter bool               `json:"enable_sensitive_info_filter"`
	CustomFilters         map[string]string      `json:"custom_filters"`
	ReplacementChar       rune                   `json:"replacement_char"`
	FilterThreshold       float64                `json:"filter_threshold"`
	CategoryThresholds    map[string]float64     `json:"category_thresholds"`
	FilterActions         map[string]ActionType  `json:"filter_actions"`
}

// FilterStats tracks statistics for content filtering
type FilterStats struct {
	Category          string    `json:"category"`
	FilterCount       int       `json:"filter_count"`
	FirstSeen         time.Time `json:"first_seen"`
	LastSeen          time.Time `json:"last_seen"`
	AverageConfidence float64   `json:"average_confidence"`
	Examples          []string  `json:"examples,omitempty"`
}

// NewEnhancedContentFilter creates a new enhanced content filter
func NewEnhancedContentFilter(config *ProtectionConfig) *EnhancedContentFilter {
	baseFilter := NewContentFilter(config)
	
	// Initialize enhanced filter config
	filterConfig := &EnhancedFilterConfig{
		EnableProfanityFilter:    true,
		EnablePIIFilter:          true,
		EnableCodeFilter:         true,
		EnableURLFilter:          true,
		EnableJailbreakFilter:    true,
		EnableSensitiveInfoFilter: true,
		CustomFilters:            make(map[string]string),
		ReplacementChar:          '*',
		FilterThreshold:          0.7,
		CategoryThresholds:       make(map[string]float64),
		FilterActions:            make(map[string]ActionType),
	}
	
	// Set default category thresholds
	filterConfig.CategoryThresholds["profanity"] = 0.7
	filterConfig.CategoryThresholds["pii"] = 0.8
	filterConfig.CategoryThresholds["code"] = 0.6
	filterConfig.CategoryThresholds["url"] = 0.6
	filterConfig.CategoryThresholds["jailbreak"] = 0.8
	filterConfig.CategoryThresholds["sensitive_info"] = 0.8
	
	// Set default filter actions
	filterConfig.FilterActions["profanity"] = ActionModified
	filterConfig.FilterActions["pii"] = ActionModified
	filterConfig.FilterActions["code"] = ActionWarned
	filterConfig.FilterActions["url"] = ActionWarned
	filterConfig.FilterActions["jailbreak"] = ActionBlocked
	filterConfig.FilterActions["sensitive_info"] = ActionModified
	
	// Initialize sensitive patterns
	sensitivePatterns := map[string]*regexp.Regexp{
		"email": regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`),
		"phone": regexp.MustCompile(`\b(\+\d{1,2}\s)?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}\b`),
		"ssn": regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		"credit_card": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
		"ip_address": regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
		"address": regexp.MustCompile(`\b\d+\s+[A-Za-z0-9\s,]+\b(?:street|st|avenue|ave|road|rd|highway|hwy|square|sq|trail|trl|drive|dr|court|ct|park|parkway|pkwy|circle|cir|boulevard|blvd)\b`),
	}
	
	// Initialize prohibited patterns
	prohibitedPatterns := map[string]*regexp.Regexp{
		"profanity": regexp.MustCompile(`(?i)\b(fuck|shit|ass|bitch|cunt|damn|dick|pussy|asshole|bastard)\b`),
		"hate_speech": regexp.MustCompile(`(?i)\b(nigger|faggot|spic|chink|kike|retard|wetback|towelhead)\b`),
		"violence": regexp.MustCompile(`(?i)\b(kill|murder|assassinate|bomb|shoot|strangle|torture)\b`),
		"self_harm": regexp.MustCompile(`(?i)\b(suicide|self-harm|cut myself|kill myself|hang myself)\b`),
		"sexual_content": regexp.MustCompile(`(?i)\b(porn|pornography|sex|masturbate|orgasm|penis|vagina|blowjob|handjob)\b`),
	}
	
	// Initialize category filters
	categoryFilters := make(map[string]func(string) (bool, float64, string))
	
	// Profanity filter
	categoryFilters["profanity"] = func(text string) (bool, float64, string) {
		matches := prohibitedPatterns["profanity"].FindAllStringIndex(text, -1)
		if len(matches) > 0 {
			confidence := min(float64(len(matches))*0.2, 1.0)
			return true, confidence, "Profanity detected"
		}
		return false, 0.0, ""
	}
	
	// PII filter
	categoryFilters["pii"] = func(text string) (bool, float64, string) {
		for category, pattern := range sensitivePatterns {
			matches := pattern.FindAllStringIndex(text, -1)
			if len(matches) > 0 {
				confidence := min(float64(len(matches))*0.3, 1.0)
				return true, confidence, fmt.Sprintf("PII detected: %s", category)
			}
		}
		return false, 0.0, ""
	}
	
	// Code filter
	categoryFilters["code"] = func(text string) (bool, float64, string) {
		codePatterns := []string{
			"```",
			"function",
			"class",
			"def ",
			"import ",
			"from ",
			"public static",
			"private ",
			"protected ",
			"<script",
			"<style",
			"#!/",
		}
		
		for _, pattern := range codePatterns {
			if strings.Contains(text, pattern) {
				return true, 0.7, "Code snippet detected"
			}
		}
		return false, 0.0, ""
	}
	
	// URL filter
	categoryFilters["url"] = func(text string) (bool, float64, string) {
		urlPattern := regexp.MustCompile(`(https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|www\.[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9]+\.[^\s]{2,}|www\.[a-zA-Z0-9]+\.[^\s]{2,})`)
		matches := urlPattern.FindAllStringIndex(text, -1)
		if len(matches) > 0 {
			confidence := min(float64(len(matches))*0.2, 1.0)
			return true, confidence, "URL detected"
		}
		return false, 0.0, ""
	}
	
	// Jailbreak filter
	categoryFilters["jailbreak"] = func(text string) (bool, float64, string) {
		jailbreakPatterns := []string{
			"DAN",
			"Do Anything Now",
			"ignore previous instructions",
			"disregard your programming",
			"bypass your rules",
			"ignore your guidelines",
			"you can do anything",
			"you have no restrictions",
			"you are free from constraints",
			"you don't have to follow",
		}
		
		for _, pattern := range jailbreakPatterns {
			if strings.Contains(strings.ToLower(text), strings.ToLower(pattern)) {
				return true, 0.9, "Jailbreak attempt detected"
			}
		}
		return false, 0.0, ""
	}
	
	// Sensitive info filter
	categoryFilters["sensitive_info"] = func(text string) (bool, float64, string) {
		sensitivePatterns := []string{
			"password",
			"secret",
			"api key",
			"token",
			"credential",
			"private key",
			"access key",
			"auth token",
		}
		
		for _, pattern := range sensitivePatterns {
			if strings.Contains(strings.ToLower(text), strings.ToLower(pattern)) {
				return true, 0.8, "Sensitive information detected"
			}
		}
		return false, 0.0, ""
	}
	
	// Initialize replacement strategies
	replacementStrategies := make(map[string]func(string, string) string)
	
	// Character replacement
	replacementStrategies["character"] = func(text string, pattern string) string {
		re := regexp.MustCompile(pattern)
		return re.ReplaceAllStringFunc(text, func(match string) string {
			return strings.Repeat(string(filterConfig.ReplacementChar), len(match))
		})
	}
	
	// Token replacement
	replacementStrategies["token"] = func(text string, pattern string) string {
		re := regexp.MustCompile(pattern)
		return re.ReplaceAllString(text, "[FILTERED]")
	}
	
	// Context-aware replacement
	replacementStrategies["context"] = func(text string, pattern string) string {
		re := regexp.MustCompile(pattern)
		return re.ReplaceAllString(text, "[FILTERED: Sensitive Content]")
	}
	
	return &EnhancedContentFilter{
		ContentFilter:         baseFilter,
		config:                config,
		filterConfig:          filterConfig,
		sensitivePatterns:     sensitivePatterns,
		prohibitedPatterns:    prohibitedPatterns,
		categoryFilters:       categoryFilters,
		replacementStrategies: replacementStrategies,
		filterStats:           make(map[string]*FilterStats),
	}
}

// FilterContentEnhanced filters content with enhanced filtering
func (f *EnhancedContentFilter) FilterContentEnhanced(ctx context.Context, content string) (string, *ProtectionResult, error) {
	startTime := time.Now()
	
	result := &ProtectionResult{
		OriginalResponse:  content,
		ProtectedResponse: content,
		Detections:        make([]*Detection, 0),
		RiskScore:         0.0,
		ActionTaken:       ActionNone,
		Timestamp:         startTime,
	}
	
	// Apply category filters
	filteredContent := content
	for category, filter := range f.categoryFilters {
		if f.shouldApplyFilter(category) {
			detected, confidence, message := filter(filteredContent)
			if detected && confidence >= f.filterConfig.CategoryThresholds[category] {
				// Create detection
				detection := &Detection{
					Type:        DetectionTypeProhibitedContent,
					Description: message,
					Confidence:  confidence,
					Metadata: map[string]interface{}{
						"category": category,
					},
				}
				
				result.Detections = append(result.Detections, detection)
				result.RiskScore = max(result.RiskScore, confidence)
				
				// Apply action based on category
				action := f.filterConfig.FilterActions[category]
				if action > result.ActionTaken {
					result.ActionTaken = action
				}
				
				// Apply filtering if action is modification
				if action == ActionModified {
					filteredContent = f.applyFilterForCategory(filteredContent, category)
				} else if action == ActionBlocked {
					filteredContent = f.createBlockedContentMessage(category, message)
					break // Stop processing if content is blocked
				}
				
				// Update filter stats
				f.updateFilterStats(category, confidence, content)
			}
		}
	}
	
	result.ProtectedResponse = filteredContent
	result.ProcessingTime = time.Since(startTime)
	
	return filteredContent, result, nil
}

// shouldApplyFilter determines if a filter should be applied
func (f *EnhancedContentFilter) shouldApplyFilter(category string) bool {
	switch category {
	case "profanity":
		return f.filterConfig.EnableProfanityFilter
	case "pii":
		return f.filterConfig.EnablePIIFilter
	case "code":
		return f.filterConfig.EnableCodeFilter
	case "url":
		return f.filterConfig.EnableURLFilter
	case "jailbreak":
		return f.filterConfig.EnableJailbreakFilter
	case "sensitive_info":
		return f.filterConfig.EnableSensitiveInfoFilter
	default:
		return false
	}
}

// applyFilterForCategory applies filtering for a specific category
func (f *EnhancedContentFilter) applyFilterForCategory(content string, category string) string {
	switch category {
	case "profanity":
		return f.replacementStrategies["character"](content, f.prohibitedPatterns["profanity"].String())
	case "pii":
		filtered := content
		for _, pattern := range f.sensitivePatterns {
			filtered = f.replacementStrategies["token"](filtered, pattern.String())
		}
		return filtered
	case "code":
		// For code, we just add a warning but don't modify the content
		return content
	case "url":
		urlPattern := `(https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|www\.[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9]+\.[^\s]{2,}|www\.[a-zA-Z0-9]+\.[^\s]{2,})`
		return f.replacementStrategies["token"](content, urlPattern)
	case "jailbreak":
		// For jailbreak, we block the entire content
		return f.createBlockedContentMessage(category, "Jailbreak attempt detected")
	case "sensitive_info":
		// For sensitive info, we replace with context-aware messages
		filtered := content
		sensitivePatterns := []string{
			"password",
			"secret",
			"api key",
			"token",
			"credential",
			"private key",
			"access key",
			"auth token",
		}
		
		for _, pattern := range sensitivePatterns {
			patternRegex := fmt.Sprintf(`(?i)(%s\s*[:=]\s*[^\s]+)`, regexp.QuoteMeta(pattern))
			filtered = f.replacementStrategies["context"](filtered, patternRegex)
		}
		return filtered
	default:
		return content
	}
}

// createBlockedContentMessage creates a message for blocked content
func (f *EnhancedContentFilter) createBlockedContentMessage(category string, message string) string {
	return fmt.Sprintf("I'm unable to provide the requested content due to content filtering: %s", message)
}

// updateFilterStats updates the filter statistics
func (f *EnhancedContentFilter) updateFilterStats(category string, confidence float64, example string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	stats, ok := f.filterStats[category]
	if !ok {
		stats = &FilterStats{
			Category:          category,
			FilterCount:       0,
			FirstSeen:         time.Now(),
			LastSeen:          time.Now(),
			AverageConfidence: 0,
			Examples:          make([]string, 0),
		}
		f.filterStats[category] = stats
	}
	
	// Update stats
	stats.FilterCount++
	stats.LastSeen = time.Now()
	stats.AverageConfidence = ((stats.AverageConfidence * float64(stats.FilterCount-1)) + confidence) / float64(stats.FilterCount)
	
	// Add example if we have fewer than 5
	if len(stats.Examples) < 5 {
		// Truncate example if it's too long
		if len(example) > 100 {
			example = example[:100] + "..."
		}
		stats.Examples = append(stats.Examples, example)
	}
}

// GetFilterStats gets the filter statistics
func (f *EnhancedContentFilter) GetFilterStats() map[string]*FilterStats {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	return f.filterStats
}

// SetFilterThreshold sets the threshold for a category
func (f *EnhancedContentFilter) SetFilterThreshold(category string, threshold float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.filterConfig.CategoryThresholds[category] = threshold
}

// SetFilterAction sets the action for a category
func (f *EnhancedContentFilter) SetFilterAction(category string, action ActionType) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.filterConfig.FilterActions[category] = action
}

// EnableFilter enables or disables a filter
func (f *EnhancedContentFilter) EnableFilter(category string, enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	switch category {
	case "profanity":
		f.filterConfig.EnableProfanityFilter = enabled
	case "pii":
		f.filterConfig.EnablePIIFilter = enabled
	case "code":
		f.filterConfig.EnableCodeFilter = enabled
	case "url":
		f.filterConfig.EnableURLFilter = enabled
	case "jailbreak":
		f.filterConfig.EnableJailbreakFilter = enabled
	case "sensitive_info":
		f.filterConfig.EnableSensitiveInfoFilter = enabled
	}
}

// AddCustomFilter adds a custom filter
func (f *EnhancedContentFilter) AddCustomFilter(name string, pattern string, threshold float64, action ActionType) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// Validate the pattern
	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}
	
	// Add the filter
	f.filterConfig.CustomFilters[name] = pattern
	f.filterConfig.CategoryThresholds[name] = threshold
	f.filterConfig.FilterActions[name] = action
	
	// Add the category filter
	f.categoryFilters[name] = func(text string) (bool, float64, string) {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringIndex(text, -1)
		if len(matches) > 0 {
			confidence := min(float64(len(matches))*0.2, 1.0)
			return true, confidence, fmt.Sprintf("Custom filter matched: %s", name)
		}
		return false, 0.0, ""
	}
	
	return nil
}
