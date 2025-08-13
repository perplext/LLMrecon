// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"sync"
)

// TemplateMonitor monitors templates for unusual patterns
type TemplateMonitor struct {
	config          *ProtectionConfig
	monitoringConfig *MonitoringConfig
	patternLibrary   *InjectionPatternLibrary
	patternStats     map[string]*TemplatePatternStats
	stopChan         chan struct{}
	mu               sync.RWMutex
	running          bool
}

// NewTemplateMonitor creates a new template monitor
func NewTemplateMonitor(config *ProtectionConfig, patternLibrary *InjectionPatternLibrary) *TemplateMonitor {
	// Create default monitoring config if not specified
	monitoringConfig := &MonitoringConfig{
		MonitoringInterval: config.MonitoringInterval,
		MaxPatternHistory:  100,
		AnomalyThreshold:   0.8,
		EnableAnomalyDetection: true,
	}

	return &TemplateMonitor{
		config:          config,
		monitoringConfig: monitoringConfig,
		patternLibrary:   patternLibrary,
		patternStats:     make(map[string]*TemplatePatternStats),
		stopChan:         make(chan struct{}),
	}
}

// Start starts the template monitor
func (m *TemplateMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.running {
		return
	}
	
	m.running = true
	
	// Start the monitoring loop
	go m.monitoringLoop(ctx)
}

// Stop stops the template monitor
func (m *TemplateMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.running {
		return
	}
	
	m.running = false
	close(m.stopChan)
}

// monitoringLoop is the main monitoring loop
func (m *TemplateMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(m.monitoringConfig.MonitoringInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.detectAnomalies(ctx)
		case <-m.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// detectAnomalies detects anomalies in the pattern statistics
func (m *TemplateMonitor) detectAnomalies(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if !m.monitoringConfig.EnableAnomalyDetection {
		return
	}
	
	// Implement anomaly detection logic here
	// This is a placeholder for future implementation
	// Potential approaches:
	// - Detect sudden spikes in pattern frequency
	// - Detect unusual combinations of patterns
	// - Detect patterns with consistently high risk scores
}

// MonitorPrompt monitors a prompt for unusual patterns
func (m *TemplateMonitor) MonitorPrompt(ctx context.Context, result *ProtectionResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Update pattern statistics for each detection
	for _, detection := range result.Detections {
		patternKey := string(detection.Type) + ":" + detection.Pattern
		
		// Get or create pattern stats
		stats, exists := m.patternStats[patternKey]
		if !exists {
			stats = &TemplatePatternStats{
				Pattern:          detection.Pattern,
				Count:            0,
				FirstSeen:        time.Now(),
				LastSeen:         time.Now(),
				AverageRiskScore: 0,
				DetectionTypes:   make(map[DetectionType]int),
				Examples:         make([]string, 0),
				Metadata:         make(map[string]interface{}),
			}
			m.patternStats[patternKey] = stats
		}
		
		// Update stats
		stats.Count++
		stats.LastSeen = time.Now()
		
		// Update average risk score
		stats.AverageRiskScore = ((stats.AverageRiskScore * float64(stats.Count-1)) + detection.Confidence) / float64(stats.Count)
		
		// Update detection types
		stats.DetectionTypes[detection.Type]++
		
		// Add example if we don't have too many
		if len(stats.Examples) < 5 && detection.Location != nil {
			stats.Examples = append(stats.Examples, detection.Location.Context)
		}
		
		// Check for unusual pattern
		if m.isUnusualPattern(stats) {
			// Add unusual pattern detection
			unusualDetection := &Detection{
				Type:        DetectionTypeUnusualPattern,
				Confidence:  0.7,
				Description: "Unusual pattern detected: " + detection.Pattern,
				Pattern:     detection.Pattern,
				Remediation: "Monitor this pattern for potential security issues",
				Metadata: map[string]interface{}{
					"pattern_stats": stats,
				},
			}
			
			result.Detections = append(result.Detections, unusualDetection)
			result.RiskScore = max(result.RiskScore, 0.7)
		}
	}
	
	// Prune old patterns if we have too many
	if len(m.patternStats) > m.monitoringConfig.MaxPatternHistory {
		m.pruneOldPatterns()
	}
}

// isUnusualPattern determines if a pattern is unusual
func (m *TemplateMonitor) isUnusualPattern(stats *TemplatePatternStats) bool {
	// Consider a pattern unusual if:
	// 1. It's the first time we've seen it
	// 2. It has a high average risk score
	// 3. It's seen infrequently but consistently
	
	if stats.Count == 1 {
		// First time seeing this pattern
		return true
	}
	
	if stats.AverageRiskScore >= m.monitoringConfig.AnomalyThreshold {
		// High risk score
		return true
	}
	
	// More sophisticated detection would be implemented here
	
	return false
}

// pruneOldPatterns removes old patterns from the statistics
func (m *TemplateMonitor) pruneOldPatterns() {
	// Find the oldest patterns
	oldPatterns := make([]patternAge, 0, len(m.patternStats))
	for key, stats := range m.patternStats {
		oldPatterns = append(oldPatterns, patternAge{pattern: key, lastSeen: stats.LastSeen})
	}
	
	// Sort by last seen time (oldest first)
	sortPatternsByAge(oldPatterns)
	
	// Remove the oldest patterns until we're under the limit
	numToRemove := len(m.patternStats) - m.monitoringConfig.MaxPatternHistory
	for i := 0; i < numToRemove && i < len(oldPatterns); i++ {
		delete(m.patternStats, oldPatterns[i].pattern)
	}
}

// sortPatternsByAge sorts patterns by age (oldest first)
func sortPatternsByAge(patterns []patternAge) {
	// Simple bubble sort for now
	for i := 0; i < len(patterns); i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[i].lastSeen.After(patterns[j].lastSeen) {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}
}

// GetPatternStats gets the statistics for a pattern
func (m *TemplateMonitor) GetPatternStats(pattern string) *TemplatePatternStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.patternStats[pattern]
}

// GetAllPatternStats gets all pattern statistics
func (m *TemplateMonitor) GetAllPatternStats() []*TemplatePatternStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := make([]*TemplatePatternStats, 0, len(m.patternStats))
	for _, stat := range m.patternStats {
		stats = append(stats, stat)
	}
	
	return stats
}
