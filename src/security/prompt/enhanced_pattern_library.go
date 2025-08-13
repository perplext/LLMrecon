// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// EnhancedInjectionPatternLibrary extends the InjectionPatternLibrary with more sophisticated pattern management
type EnhancedInjectionPatternLibrary struct {
	*InjectionPatternLibrary
	categorizedPatterns map[string][]*InjectionPattern
	patternStats        map[string]*PatternStats
	emergingPatterns    []*EmergingPattern
	customPatterns      []*CustomPattern
	patternSources      map[string]string
	lastUpdateTime      time.Time
	updateInterval      time.Duration
	dataDir             string
	mu                  sync.RWMutex
}

// PatternStats tracks statistics for pattern matches
type PatternStats struct {
	Pattern       string    `json:"pattern"`
	MatchCount    int       `json:"match_count"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	EffectivenessScore float64 `json:"effectiveness_score"`
	FalsePositiveRate  float64 `json:"false_positive_rate"`
	Categories    []string  `json:"categories"`
}

// EmergingPattern represents a newly discovered pattern that hasn't been fully validated
type EmergingPattern struct {
	Pattern       string    `json:"pattern"`
	Description   string    `json:"description"`
	Source        string    `json:"source"`
	DiscoveryTime time.Time `json:"discovery_time"`
	LastSeen      time.Time `json:"last_seen"`
	Confidence    float64   `json:"confidence"`
	Examples      []string  `json:"examples"`
	Validated     bool      `json:"validated"`
}

// CustomPattern represents a user-defined pattern
type CustomPattern struct {
	Pattern       string    `json:"pattern"`
	Description   string    `json:"description"`
	Creator       string    `json:"creator"`
	CreationTime  time.Time `json:"creation_time"`
	Enabled       bool      `json:"enabled"`
	Categories    []string  `json:"categories"`
}

// NewEnhancedInjectionPatternLibrary creates a new enhanced injection pattern library
func NewEnhancedInjectionPatternLibrary(dataDir string) (*EnhancedInjectionPatternLibrary, error) {
	baseLibrary := NewInjectionPatternLibrary()
	
	// Create the data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	library := &EnhancedInjectionPatternLibrary{
		InjectionPatternLibrary: baseLibrary,
		categorizedPatterns:     make(map[string][]*InjectionPattern),
		patternStats:            make(map[string]*PatternStats),
		emergingPatterns:        make([]*EmergingPattern, 0),
		customPatterns:          make([]*CustomPattern, 0),
		patternSources:          make(map[string]string),
		lastUpdateTime:          time.Now(),
		updateInterval:          time.Hour * 24, // Update patterns daily
		dataDir:                 dataDir,
	}
	
	// Initialize pattern categories
	library.initializePatternCategories()
	
	// Load patterns from disk
	if err := library.loadPatternsFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load patterns from disk: %w", err)
	}
	
	return library, nil
}

// initializePatternCategories initializes the pattern categories
func (l *EnhancedInjectionPatternLibrary) initializePatternCategories() {
	// Define pattern categories
	categories := []string{
		"direct_injection",
		"indirect_injection",
		"jailbreak",
		"role_change",
		"system_prompt",
		"boundary_violation",
		"delimiter_misuse",
		"unusual_pattern",
		"prohibited_content",
		"sensitive_info",
		"obfuscation",
		"multi_stage",
		"emerging",
	}
	
	// Initialize each category
	for _, category := range categories {
		l.categorizedPatterns[category] = make([]*InjectionPattern, 0)
	}
	
	// Add default patterns to categories
	for _, pattern := range l.patterns {
		switch {
		case regexp.MustCompile(`(?i)(ignore|disregard) (previous|above|earlier|all) (instructions|prompts|directives|guidance)`).MatchString(pattern.Pattern):
			l.categorizedPatterns["direct_injection"] = append(l.categorizedPatterns["direct_injection"], pattern)
		
		case regexp.MustCompile(`(?i)(you are|act as|pretend to be) (a|an) ([a-zA-Z\s]+)`).MatchString(pattern.Pattern):
			l.categorizedPatterns["role_change"] = append(l.categorizedPatterns["role_change"], pattern)
		
		case regexp.MustCompile(`(?i)(system|instruction|prompt):`).MatchString(pattern.Pattern):
			l.categorizedPatterns["system_prompt"] = append(l.categorizedPatterns["system_prompt"], pattern)
		
		case regexp.MustCompile(`(?i)(DAN|DUDE|STAN|jailbreak|waluigi)`).MatchString(pattern.Pattern):
			l.categorizedPatterns["jailbreak"] = append(l.categorizedPatterns["jailbreak"], pattern)
		
		default:
			// Default to unusual pattern if no specific category matches
			l.categorizedPatterns["unusual_pattern"] = append(l.categorizedPatterns["unusual_pattern"], pattern)
		}
	}
}

// loadPatternsFromDisk loads patterns from disk
func (l *EnhancedInjectionPatternLibrary) loadPatternsFromDisk() error {
	// Load pattern stats
	statsFile := filepath.Join(l.dataDir, "pattern_stats.json")
	if _, err := os.Stat(statsFile); err == nil {
		data, err := ioutil.ReadFile(statsFile)
		if err != nil {
			return fmt.Errorf("failed to read pattern stats file: %w", err)
		}
		
		var stats map[string]*PatternStats
		if err := json.Unmarshal(data, &stats); err != nil {
			return fmt.Errorf("failed to unmarshal pattern stats: %w", err)
		}
		
		l.patternStats = stats
	}
	
	// Load emerging patterns
	emergingFile := filepath.Join(l.dataDir, "emerging_patterns.json")
	if _, err := os.Stat(emergingFile); err == nil {
		data, err := ioutil.ReadFile(emergingFile)
		if err != nil {
			return fmt.Errorf("failed to read emerging patterns file: %w", err)
		}
		
		var patterns []*EmergingPattern
		if err := json.Unmarshal(data, &patterns); err != nil {
			return fmt.Errorf("failed to unmarshal emerging patterns: %w", err)
		}
		
		l.emergingPatterns = patterns
	}
	
	// Load custom patterns
	customFile := filepath.Join(l.dataDir, "custom_patterns.json")
	if _, err := os.Stat(customFile); err == nil {
		data, err := ioutil.ReadFile(customFile)
		if err != nil {
			return fmt.Errorf("failed to read custom patterns file: %w", err)
		}
		
		var patterns []*CustomPattern
		if err := json.Unmarshal(data, &patterns); err != nil {
			return fmt.Errorf("failed to unmarshal custom patterns: %w", err)
		}
		
		l.customPatterns = patterns
		
		// Add enabled custom patterns to the base library
		for _, pattern := range patterns {
			if pattern.Enabled {
				patternObj := &InjectionPattern{
					Pattern:     pattern.Pattern,
					Description: pattern.Description,
					Confidence:  0.8,
				}
				l.AddPattern(patternObj)
			}
		}
	}
	
	return nil
}

// savePatternsToDisc saves patterns to disk
func (l *EnhancedInjectionPatternLibrary) savePatternsToDisc() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Save pattern stats
	statsFile := filepath.Join(l.dataDir, "pattern_stats.json")
	statsData, err := json.MarshalIndent(l.patternStats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pattern stats: %w", err)
	}
	
	if err := ioutil.WriteFile(statsFile, statsData, 0644); err != nil {
		return fmt.Errorf("failed to write pattern stats file: %w", err)
	}
	
	// Save emerging patterns
	emergingFile := filepath.Join(l.dataDir, "emerging_patterns.json")
	emergingData, err := json.MarshalIndent(l.emergingPatterns, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal emerging patterns: %w", err)
	}
	
	if err := ioutil.WriteFile(emergingFile, emergingData, 0644); err != nil {
		return fmt.Errorf("failed to write emerging patterns file: %w", err)
	}
	
	// Save custom patterns
	customFile := filepath.Join(l.dataDir, "custom_patterns.json")
	customData, err := json.MarshalIndent(l.customPatterns, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal custom patterns: %w", err)
	}
	
	if err := ioutil.WriteFile(customFile, customData, 0644); err != nil {
		return fmt.Errorf("failed to write custom patterns file: %w", err)
	}
	
	return nil
}

// DetectPatternsEnhanced detects patterns in a prompt with enhanced detection
func (l *EnhancedInjectionPatternLibrary) DetectPatternsEnhanced(prompt string, result *ProtectionResult) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// First run the base detection
	l.DetectPatterns(prompt, result)
	
	// Check for emerging patterns
	for _, pattern := range l.emergingPatterns {
		if regexp.MustCompile(pattern.Pattern).MatchString(prompt) {
			// Create detection
			detection := &Detection{
				Type:        DetectionTypeUnusualPattern,
				Description: pattern.Description,
				Confidence:  pattern.Confidence,
				Pattern:     pattern.Pattern,
				Location:    findPatternLocation(prompt, pattern.Pattern),
				Metadata: map[string]interface{}{
					"category": "emerging",
					"source":   pattern.Source,
					"validated": pattern.Validated,
				},
			}
			
			result.Detections = append(result.Detections, detection)
			result.RiskScore = max(result.RiskScore, pattern.Confidence)
			
			// Update pattern stats
			l.updatePatternStats(pattern.Pattern, "emerging")
		}
	}
	
	// Check for custom patterns
	for _, pattern := range l.customPatterns {
		if pattern.Enabled && regexp.MustCompile(pattern.Pattern).MatchString(prompt) {
			// Create detection
			detection := &Detection{
				Type:        DetectionTypeInjection,
				Description: pattern.Description,
				Confidence:  0.8, // Default confidence for custom patterns
				Pattern:     pattern.Pattern,
				Location:    findPatternLocation(prompt, pattern.Pattern),
				Metadata: map[string]interface{}{
					"category": "custom",
					"creator":  pattern.Creator,
					"categories": pattern.Categories,
				},
			}
			
			result.Detections = append(result.Detections, detection)
			result.RiskScore = max(result.RiskScore, 0.8)
			
			// Update pattern stats
			for _, category := range pattern.Categories {
				l.updatePatternStats(pattern.Pattern, category)
			}
		}
	}
	
	// Check if it's time to save patterns
	if time.Since(l.lastUpdateTime) > l.updateInterval {
		go func() {
			if err := l.savePatternsToDisc(); err != nil {
				// Log error but don't fail the detection
				fmt.Printf("Failed to save patterns to disk: %v\n", err)
			}
			l.lastUpdateTime = time.Now()
		}()
	}
}

// AddEmergingPattern adds a new emerging pattern
func (l *EnhancedInjectionPatternLibrary) AddEmergingPattern(pattern string, description string, source string, examples []string, confidence float64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Validate the pattern
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}
	
	// Check if the pattern already exists
	for _, p := range l.emergingPatterns {
		if p.Pattern == pattern {
			// Update existing pattern
			p.Description = description
			p.Source = source
			p.Examples = append(p.Examples, examples...)
			p.Confidence = confidence
			p.LastSeen = time.Now()
			return nil
		}
	}
	
	// Add new pattern
	l.emergingPatterns = append(l.emergingPatterns, &EmergingPattern{
		Pattern:       pattern,
		Description:   description,
		Source:        source,
		DiscoveryTime: time.Now(),
		Confidence:    confidence,
		Examples:      examples,
		Validated:     false,
	})
	
	// Save to disk
	return l.savePatternsToDisc()
}

// AddCustomPattern adds a new custom pattern
func (l *EnhancedInjectionPatternLibrary) AddCustomPattern(pattern string, description string, creator string, categories []string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Validate the pattern
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}
	
	// Check if the pattern already exists
	for _, p := range l.customPatterns {
		if p.Pattern == pattern {
			// Update existing pattern
			p.Description = description
			p.Categories = categories
			return nil
		}
	}
	
	// Add new pattern
	l.customPatterns = append(l.customPatterns, &CustomPattern{
		Pattern:      pattern,
		Description:  description,
		Creator:      creator,
		CreationTime: time.Now(),
		Enabled:      true,
		Categories:   categories,
	})
	
	// Add to base library
	patternObj := &InjectionPattern{
		Pattern:     pattern,
		Description: description,
		Confidence:  0.8,
	}
	l.AddPattern(patternObj)
	
	// Save to disk
	return l.savePatternsToDisc()
}

// ValidateEmergingPattern validates an emerging pattern
func (l *EnhancedInjectionPatternLibrary) ValidateEmergingPattern(pattern string, validated bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Find the pattern
	for _, p := range l.emergingPatterns {
		if p.Pattern == pattern {
			p.Validated = validated
			
			// If validated, add to the base library
			if validated {
				patternObj := &InjectionPattern{
					Pattern:     p.Pattern,
					Description: p.Description,
					Confidence:  p.Confidence,
				}
				l.AddPattern(patternObj)
			}
			
			// Save to disk
			return l.savePatternsToDisc()
		}
	}
	
	return fmt.Errorf("pattern not found")
}

// EnableCustomPattern enables or disables a custom pattern
func (l *EnhancedInjectionPatternLibrary) EnableCustomPattern(pattern string, enabled bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Find the pattern
	for _, p := range l.customPatterns {
		if p.Pattern == pattern {
			p.Enabled = enabled
			
			// Save to disk
			return l.savePatternsToDisc()
		}
	}
	
	return fmt.Errorf("pattern not found")
}

// GetPatternsByCategory gets patterns by category
func (l *EnhancedInjectionPatternLibrary) GetPatternsByCategory(category string) ([]*InjectionPattern, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	patterns, ok := l.categorizedPatterns[category]
	if !ok {
		return nil, fmt.Errorf("category not found")
	}
	
	return patterns, nil
}

// GetEmergingPatterns gets all emerging patterns
func (l *EnhancedInjectionPatternLibrary) GetEmergingPatterns() []*EmergingPattern {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.emergingPatterns
}

// GetCustomPatterns gets all custom patterns
func (l *EnhancedInjectionPatternLibrary) GetCustomPatterns() []*CustomPattern {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.customPatterns
}

// GetPatternStats gets statistics for a pattern
func (l *EnhancedInjectionPatternLibrary) GetPatternStats(pattern string) (*PatternStats, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	stats, ok := l.patternStats[pattern]
	if !ok {
		return nil, fmt.Errorf("pattern stats not found")
	}
	
	return stats, nil
}

// updatePatternStats updates statistics for a pattern
func (l *EnhancedInjectionPatternLibrary) updatePatternStats(pattern string, category string) {
	now := time.Now()
	
	stats, ok := l.patternStats[pattern]
	if !ok {
		// Create new stats
		stats = &PatternStats{
			Pattern:    pattern,
			MatchCount: 0,
			FirstSeen:  now,
			LastSeen:   now,
			EffectivenessScore: 0.5, // Default effectiveness
			FalsePositiveRate:  0.0, // Default false positive rate
			Categories: []string{category},
		}
		l.patternStats[pattern] = stats
	} else {
		// Update existing stats
		stats.MatchCount++
		stats.LastSeen = now
		
		// Add category if not already present
		categoryFound := false
		for _, c := range stats.Categories {
			if c == category {
				categoryFound = true
				break
			}
		}
		
		if !categoryFound {
			stats.Categories = append(stats.Categories, category)
		}
	}
}

// findPatternLocation finds the location of a pattern in a prompt
func findPatternLocation(prompt string, pattern string) *DetectionLocation {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	
	indices := re.FindStringIndex(prompt)
	if indices == nil {
		return nil
	}
	
	start := indices[0]
	end := indices[1]
	
	// Extract context (50 chars before and after)
	contextStart := start - 50
	if contextStart < 0 {
		contextStart = 0
	}
	
	contextEnd := end + 50
	if contextEnd > len(prompt) {
		contextEnd = len(prompt)
	}
	
	context := prompt[contextStart:contextEnd]
	
	// Add ellipsis if truncated
	if contextStart > 0 {
		context = "..." + context
	}
	if contextEnd < len(prompt) {
		context = context + "..."
	}
	
	return &DetectionLocation{
		Start:   start,
		End:     end,
		Context: context,
	}
}
