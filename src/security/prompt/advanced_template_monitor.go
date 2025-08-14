// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"time"
	"context"
	"fmt"
	"sync"
)

// AdvancedTemplateMonitor provides real-time monitoring for unusual template patterns
type AdvancedTemplateMonitor struct {
	config                *ProtectionConfig
	patternLibrary        *EnhancedInjectionPatternLibrary
	templateStats         map[string]*TemplateStats
	userStats             map[string]*UserStats
	sessionStats          map[string]*SessionStats
	anomalyDetector       *AnomalyDetector
	alertManager          *AlertManager
	monitoringActive      bool
	monitoringInterval    time.Duration
	lastMonitoringTime    time.Time
	monitoringThreshold   float64
	maxTemplateStats      int
	maxUserStats          int
	maxSessionStats       int
	mu                    sync.RWMutex
	stopChan              chan struct{}
}

// TemplateStats tracks statistics for a template
type TemplateStats struct {
	TemplateID          string    `json:"template_id"`
	TemplateName        string    `json:"template_name"`
	ExecutionCount      int       `json:"execution_count"`
	FirstSeen           time.Time `json:"first_seen"`
	LastSeen            time.Time `json:"last_seen"`
	AverageRiskScore    float64   `json:"average_risk_score"`
	DetectionCount      int       `json:"detection_count"`
	DetectionTypes      map[DetectionType]int `json:"detection_types"`
	SuccessRate         float64   `json:"success_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// UserStats tracks statistics for a user
type UserStats struct {
	UserID              string    `json:"user_id"`
	TemplateUsage       map[string]int `json:"template_usage"`
	TotalExecutions     int       `json:"total_executions"`
	FirstSeen           time.Time `json:"first_seen"`
	LastSeen            time.Time `json:"last_seen"`
	AverageRiskScore    float64   `json:"average_risk_score"`
	DetectionCount      int       `json:"detection_count"`
	DetectionTypes      map[DetectionType]int `json:"detection_types"`
	SuccessRate         float64   `json:"success_rate"`
	AnomalyScore        float64   `json:"anomaly_score"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// SessionStats tracks statistics for a session
type SessionStats struct {
	SessionID           string    `json:"session_id"`
	UserID              string    `json:"user_id"`
	StartTime           time.Time `json:"start_time"`
	LastActivityTime    time.Time `json:"last_activity_time"`
	ExecutionCount      int       `json:"execution_count"`
	AverageRiskScore    float64   `json:"average_risk_score"`
	DetectionCount      int       `json:"detection_count"`
	DetectionTypes      map[DetectionType]int `json:"detection_types"`
	TemplateUsage       map[string]int `json:"template_usage"`
	AnomalyScore        float64   `json:"anomaly_score"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// AnomalyDetector detects anomalies in template usage
type AnomalyDetector struct {
	baselineTemplateStats map[string]*TemplateStats
	baselineUserStats     map[string]*UserStats
	anomalyThresholds     map[string]float64
	detectionAlgorithms   map[string]func(interface{}, interface{}) float64
}

// AlertManager manages alerts for template monitoring
type AlertManager struct {
	alertHandlers        map[string]func(context.Context, *Alert) error
	alertHistory         []*Alert
	maxAlertHistory      int
	alertThresholds      map[string]float64
}

// Alert represents a monitoring alert
type Alert struct {
	AlertID             string    `json:"alert_id"`
	Timestamp           time.Time `json:"timestamp"`
	Severity            string    `json:"severity"`
	Type                string    `json:"type"`
	Message             string    `json:"message"`
	TemplateID          string    `json:"template_id,omitempty"`
	UserID              string    `json:"user_id,omitempty"`
	SessionID           string    `json:"session_id,omitempty"`
	DetectionType       DetectionType `json:"detection_type,omitempty"`
	RiskScore           float64   `json:"risk_score"`
	AnomalyScore        float64   `json:"anomaly_score"`
	RelatedAlerts       []string  `json:"related_alerts,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// NewAdvancedTemplateMonitor creates a new advanced template monitor
func NewAdvancedTemplateMonitor(config *ProtectionConfig, patternLibrary *EnhancedInjectionPatternLibrary) *AdvancedTemplateMonitor {
	// Create anomaly detector
	anomalyDetector := &AnomalyDetector{
		baselineTemplateStats: make(map[string]*TemplateStats),
		baselineUserStats:     make(map[string]*UserStats),
		anomalyThresholds: map[string]float64{
			"template_risk_score": 0.7,
			"user_risk_score":     0.8,
			"detection_frequency": 0.6,
			"pattern_deviation":   0.5,
			"usage_pattern":       0.65,
		},
		detectionAlgorithms: make(map[string]func(interface{}, interface{}) float64),
	}
	
	// Initialize detection algorithms
	anomalyDetector.detectionAlgorithms["template_risk_score"] = func(current, baseline interface{}) float64 {
		currentStats := current.(*TemplateStats)
		baselineStats := baseline.(*TemplateStats)
		
		if baselineStats.AverageRiskScore == 0 {
			return 0.0
		}
		
		deviation := (currentStats.AverageRiskScore - baselineStats.AverageRiskScore) / baselineStats.AverageRiskScore
		return deviation
	}
	
	// Create alert manager
	alertManager := &AlertManager{
		alertHandlers:   make(map[string]func(context.Context, *Alert) error),
		alertHistory:    make([]*Alert, 0),
		maxAlertHistory: 100,
		alertThresholds: map[string]float64{
			"high":   0.8,
			"medium": 0.6,
			"low":    0.4,
		},
	}
	
	// Register default alert handlers
	alertManager.alertHandlers["log"] = func(ctx context.Context, alert *Alert) error {
		// In a real implementation, this would log to a file or database
		fmt.Printf("[%s] %s: %s (Risk: %.2f, Anomaly: %.2f)\n", 
			alert.Severity, 
			alert.Type, 
			alert.Message, 
			alert.RiskScore, 
			alert.AnomalyScore)
		return nil
	}
	
	return &AdvancedTemplateMonitor{
		config:              config,
		patternLibrary:      patternLibrary,
		templateStats:       make(map[string]*TemplateStats),
		userStats:           make(map[string]*UserStats),
		sessionStats:        make(map[string]*SessionStats),
		anomalyDetector:     anomalyDetector,
		alertManager:        alertManager,
		monitoringActive:    false,
		monitoringInterval:  config.MonitoringInterval,
		lastMonitoringTime:  time.Now(),
		monitoringThreshold: 0.7,
		maxTemplateStats:    1000,
		maxUserStats:        1000,
		maxSessionStats:     1000,
		stopChan:            make(chan struct{}),
	}
}

// StartMonitoring starts the template monitoring
func (m *AdvancedTemplateMonitor) StartMonitoring(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.monitoringActive {
		return fmt.Errorf("monitoring is already active")
	}
	
	m.monitoringActive = true
	
	// Start monitoring in a goroutine
	go func() {
		ticker := time.NewTicker(m.monitoringInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.performMonitoring(ctx)
			case <-m.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return nil
}

// StopMonitoring stops the template monitoring
func (m *AdvancedTemplateMonitor) StopMonitoring() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.monitoringActive {
		return
	}
	
	m.monitoringActive = false
	m.stopChan <- struct{}{}
}

// MonitorTemplate monitors a template execution
func (m *AdvancedTemplateMonitor) MonitorTemplate(ctx context.Context, templateID string, templateName string, userID string, sessionID string, prompt string, result *ProtectionResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Update template stats
	m.updateTemplateStats(templateID, templateName, result)
	
	// Update user stats
	m.updateUserStats(userID, templateID, result)
	
	// Update session stats
	m.updateSessionStats(sessionID, userID, templateID, result)
	
	// Check for anomalies
	anomalies := m.detectAnomalies(templateID, userID, sessionID)
	
	// Generate alerts for anomalies
	for _, anomaly := range anomalies {
		alert := &Alert{
			AlertID:      fmt.Sprintf("alert-%d", time.Now().UnixNano()),
			Timestamp:    time.Now(),
			Severity:     getSeverityForAnomalyScore(anomaly.Score),
			Type:         "template_anomaly",
			Message:      anomaly.Message,
			TemplateID:   templateID,
			UserID:       userID,
			SessionID:    sessionID,
			RiskScore:    result.RiskScore,
			AnomalyScore: anomaly.Score,
			Metadata:     anomaly.Metadata,
		}
		
		// Process the alert
		m.processAlert(ctx, alert)
	}
	
	return nil
}

// updateTemplateStats updates the statistics for a template
func (m *AdvancedTemplateMonitor) updateTemplateStats(templateID string, templateName string, result *ProtectionResult) {
	stats, ok := m.templateStats[templateID]
	if !ok {
		// Create new stats
		stats = &TemplateStats{
			TemplateID:       templateID,
			TemplateName:     templateName,
			ExecutionCount:   0,
			FirstSeen:        time.Now(),
			LastSeen:         time.Now(),
			AverageRiskScore: 0,
			DetectionCount:   0,
			DetectionTypes:   make(map[DetectionType]int),
			SuccessRate:      1.0,
			Metadata:         make(map[string]interface{}),
		}
		m.templateStats[templateID] = stats
	}
	
	// Update stats
	stats.ExecutionCount++
	stats.LastSeen = time.Now()
	
	// Update average risk score
	stats.AverageRiskScore = ((stats.AverageRiskScore * float64(stats.ExecutionCount-1)) + result.RiskScore) / float64(stats.ExecutionCount)
	
	// Update detection count and types
	if len(result.Detections) > 0 {
		stats.DetectionCount += len(result.Detections)
		
		for _, detection := range result.Detections {
			stats.DetectionTypes[detection.Type]++
		}
	}
	
	// Update success rate (consider blocked or warned as failures)
	if result.ActionTaken == ActionBlocked || result.ActionTaken == ActionWarned {
		stats.SuccessRate = ((stats.SuccessRate * float64(stats.ExecutionCount-1)) + 0) / float64(stats.ExecutionCount)
	} else {
		stats.SuccessRate = ((stats.SuccessRate * float64(stats.ExecutionCount-1)) + 1) / float64(stats.ExecutionCount)
	}
	
	// Update average response time
	if stats.Metadata["average_response_time"] == nil {
		stats.Metadata["average_response_time"] = result.ProcessingTime
	} else {
		avgTime := stats.Metadata["average_response_time"].(time.Duration)
		stats.Metadata["average_response_time"] = ((avgTime * time.Duration(stats.ExecutionCount-1)) + result.ProcessingTime) / time.Duration(stats.ExecutionCount)
	}
}

// updateUserStats updates the statistics for a user
func (m *AdvancedTemplateMonitor) updateUserStats(userID string, templateID string, result *ProtectionResult) {
	stats, ok := m.userStats[userID]
	if !ok {
		// Create new stats
		stats = &UserStats{
			UserID:           userID,
			TemplateUsage:    make(map[string]int),
			TotalExecutions:  0,
			FirstSeen:        time.Now(),
			LastSeen:         time.Now(),
			AverageRiskScore: 0,
			DetectionCount:   0,
			DetectionTypes:   make(map[DetectionType]int),
			SuccessRate:      1.0,
			AnomalyScore:     0,
			Metadata:         make(map[string]interface{}),
		}
		m.userStats[userID] = stats
	}
	
	// Update stats
	stats.TotalExecutions++
	stats.LastSeen = time.Now()
	
	// Update template usage
	stats.TemplateUsage[templateID]++
	
	// Update average risk score
	stats.AverageRiskScore = ((stats.AverageRiskScore * float64(stats.TotalExecutions-1)) + result.RiskScore) / float64(stats.TotalExecutions)
	
	// Update detection count and types
	if len(result.Detections) > 0 {
		stats.DetectionCount += len(result.Detections)
		
		for _, detection := range result.Detections {
			stats.DetectionTypes[detection.Type]++
		}
	}
	
	// Update success rate (consider blocked or warned as failures)
	if result.ActionTaken == ActionBlocked || result.ActionTaken == ActionWarned {
		stats.SuccessRate = ((stats.SuccessRate * float64(stats.TotalExecutions-1)) + 0) / float64(stats.TotalExecutions)
	} else {
		stats.SuccessRate = ((stats.SuccessRate * float64(stats.TotalExecutions-1)) + 1) / float64(stats.TotalExecutions)
	}
}

// updateSessionStats updates the statistics for a session
func (m *AdvancedTemplateMonitor) updateSessionStats(sessionID string, userID string, templateID string, result *ProtectionResult) {
	stats, ok := m.sessionStats[sessionID]
	if !ok {
		// Create new stats
		stats = &SessionStats{
			SessionID:        sessionID,
			UserID:           userID,
			StartTime:        time.Now(),
			LastActivityTime: time.Now(),
			ExecutionCount:   0,
			AverageRiskScore: 0,
			DetectionCount:   0,
			DetectionTypes:   make(map[DetectionType]int),
			TemplateUsage:    make(map[string]int),
			AnomalyScore:     0,
			Metadata:         make(map[string]interface{}),
		}
		m.sessionStats[sessionID] = stats
	}
	
	// Update stats
	stats.ExecutionCount++
	stats.LastActivityTime = time.Now()
	
	// Update template usage
	stats.TemplateUsage[templateID]++
	
	// Update average risk score
	stats.AverageRiskScore = ((stats.AverageRiskScore * float64(stats.ExecutionCount-1)) + result.RiskScore) / float64(stats.ExecutionCount)
	
	// Update detection count and types
	if len(result.Detections) > 0 {
		stats.DetectionCount += len(result.Detections)
		
		for _, detection := range result.Detections {
			stats.DetectionTypes[detection.Type]++
		}
	}
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	Type     string                 `json:"type"`
	Message  string                 `json:"message"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// detectAnomalies detects anomalies in template usage
func (m *AdvancedTemplateMonitor) detectAnomalies(templateID string, userID string, sessionID string) []*Anomaly {
	anomalies := make([]*Anomaly, 0)
	
	// Get stats
	templateStats, templateOk := m.templateStats[templateID]
	userStats, userOk := m.userStats[userID]
	sessionStats, sessionOk := m.sessionStats[sessionID]
	
	if !templateOk || !userOk || !sessionOk {
		return anomalies
	}
	
	// Check for high risk scores
	if templateStats.AverageRiskScore > m.anomalyDetector.anomalyThresholds["template_risk_score"] {
		anomalies = append(anomalies, &Anomaly{
			Type:    "high_risk_template",
			Message: fmt.Sprintf("Template %s has a high average risk score of %.2f", templateID, templateStats.AverageRiskScore),
			Score:   templateStats.AverageRiskScore,
			Metadata: map[string]interface{}{
				"template_id":   templateID,
				"template_name": templateStats.TemplateName,
				"risk_score":    templateStats.AverageRiskScore,
			},
		})
	}
	
	if userStats.AverageRiskScore > m.anomalyDetector.anomalyThresholds["user_risk_score"] {
		anomalies = append(anomalies, &Anomaly{
			Type:    "high_risk_user",
			Message: fmt.Sprintf("User %s has a high average risk score of %.2f", userID, userStats.AverageRiskScore),
			Score:   userStats.AverageRiskScore,
			Metadata: map[string]interface{}{
				"user_id":    userID,
				"risk_score": userStats.AverageRiskScore,
			},
		})
	}
	
	// Check for unusual detection patterns
	if templateStats.DetectionCount > 0 {
		detectionRate := float64(templateStats.DetectionCount) / float64(templateStats.ExecutionCount)
		if detectionRate > m.anomalyDetector.anomalyThresholds["detection_frequency"] {
			anomalies = append(anomalies, &Anomaly{
				Type:    "high_detection_rate",
				Message: fmt.Sprintf("Template %s has a high detection rate of %.2f", templateID, detectionRate),
				Score:   detectionRate,
				Metadata: map[string]interface{}{
					"template_id":     templateID,
					"template_name":   templateStats.TemplateName,
					"detection_rate":  detectionRate,
					"detection_count": templateStats.DetectionCount,
				},
			})
		}
	}
	
	// Check for unusual session patterns
	if sessionStats.ExecutionCount > 10 {
		// Check for rapid template switching
		if len(sessionStats.TemplateUsage) > 5 {
			switchRate := float64(len(sessionStats.TemplateUsage)) / float64(sessionStats.ExecutionCount)
			if switchRate > 0.5 {
				anomalies = append(anomalies, &Anomaly{
					Type:    "rapid_template_switching",
					Message: fmt.Sprintf("Session %s is rapidly switching between templates (rate: %.2f)", sessionID, switchRate),
					Score:   switchRate,
					Metadata: map[string]interface{}{
						"session_id":   sessionID,
						"user_id":      userID,
						"switch_rate":  switchRate,
						"template_count": len(sessionStats.TemplateUsage),
					},
				})
			}
		}
		
		// Check for unusual execution frequency
		sessionDuration := time.Since(sessionStats.StartTime)
		executionRate := float64(sessionStats.ExecutionCount) / sessionDuration.Seconds()
		if executionRate > 0.2 { // More than 1 execution per 5 seconds
			anomalies = append(anomalies, &Anomaly{
				Type:    "high_execution_rate",
				Message: fmt.Sprintf("Session %s has a high execution rate of %.2f per second", sessionID, executionRate),
				Score:   min(executionRate*5, 1.0), // Scale to 0-1
				Metadata: map[string]interface{}{
					"session_id":     sessionID,
					"user_id":        userID,
					"execution_rate": executionRate,
					"session_duration": sessionDuration.String(),
				},
			})
		}
	}
	
	// Check for unusual detection types
	for detectionType, count := range templateStats.DetectionTypes {
		if count > 3 {
			anomalies = append(anomalies, &Anomaly{
				Type:    "repeated_detection_type",
				Message: fmt.Sprintf("Template %s has %d detections of type %s", templateID, count, detectionType),
				Score:   min(float64(count)/10.0, 1.0), // Scale to 0-1
				Metadata: map[string]interface{}{
					"template_id":    templateID,
					"template_name":  templateStats.TemplateName,
					"detection_type": string(detectionType),
					"count":          count,
				},
			})
		}
	}
	
	return anomalies
}

// processAlert processes an alert
func (m *AdvancedTemplateMonitor) processAlert(ctx context.Context, alert *Alert) {
	// Add to alert history
	m.alertManager.alertHistory = append(m.alertManager.alertHistory, alert)
	
	// Trim if too many
	if len(m.alertManager.alertHistory) > m.alertManager.maxAlertHistory {
		m.alertManager.alertHistory = m.alertManager.alertHistory[1:]
	}
	
	// Process with alert handlers
	for _, handler := range m.alertManager.alertHandlers {
		if err := handler(ctx, alert); err != nil {
			// Log error but continue processing
			fmt.Printf("Error processing alert: %v\n", err)
		}
	}
}

// performMonitoring performs periodic monitoring tasks
func (m *AdvancedTemplateMonitor) performMonitoring(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Update last monitoring time
	m.lastMonitoringTime = time.Now()
	
	// Check for inactive sessions
	for sessionID, stats := range m.sessionStats {
		if time.Since(stats.LastActivityTime) > time.Hour {
			// Session is inactive, remove it
			delete(m.sessionStats, sessionID)
		}
	}
	
	// Check for templates with high risk scores
	for templateID, stats := range m.templateStats {
		if stats.AverageRiskScore > m.monitoringThreshold {
			alert := &Alert{
				AlertID:      fmt.Sprintf("alert-%d", time.Now().UnixNano()),
				Timestamp:    time.Now(),
				Severity:     getSeverityForRiskScore(stats.AverageRiskScore),
				Type:         "high_risk_template",
				Message:      fmt.Sprintf("Template %s has a high average risk score of %.2f", templateID, stats.AverageRiskScore),
				TemplateID:   templateID,
				RiskScore:    stats.AverageRiskScore,
				AnomalyScore: 0,
				Metadata: map[string]interface{}{
					"template_name":     stats.TemplateName,
					"execution_count":   stats.ExecutionCount,
					"detection_count":   stats.DetectionCount,
					"success_rate":      stats.SuccessRate,
				},
			}
			
			// Process the alert
			m.processAlert(ctx, alert)
		}
	}
	
	// Check for users with high risk scores
	for userID, stats := range m.userStats {
		if stats.AverageRiskScore > m.monitoringThreshold {
			alert := &Alert{
				AlertID:      fmt.Sprintf("alert-%d", time.Now().UnixNano()),
				Timestamp:    time.Now(),
				Severity:     getSeverityForRiskScore(stats.AverageRiskScore),
				Type:         "high_risk_user",
				Message:      fmt.Sprintf("User %s has a high average risk score of %.2f", userID, stats.AverageRiskScore),
				UserID:       userID,
				RiskScore:    stats.AverageRiskScore,
				AnomalyScore: 0,
				Metadata: map[string]interface{}{
					"total_executions": stats.TotalExecutions,
					"detection_count":  stats.DetectionCount,
					"success_rate":     stats.SuccessRate,
				},
			}
			
			// Process the alert
			m.processAlert(ctx, alert)
		}
	}
}

// RegisterAlertHandler registers a handler for alerts
func (m *AdvancedTemplateMonitor) RegisterAlertHandler(name string, handler func(context.Context, *Alert) error) {
	m.alertManager.alertHandlers[name] = handler
}

// GetTemplateStats gets statistics for a template
func (m *AdvancedTemplateMonitor) GetTemplateStats(templateID string) (*TemplateStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats, ok := m.templateStats[templateID]
	if !ok {
		return nil, fmt.Errorf("template stats not found")
	}
	
	return stats, nil
}

// GetUserStats gets statistics for a user
func (m *AdvancedTemplateMonitor) GetUserStats(userID string) (*UserStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats, ok := m.userStats[userID]
	if !ok {
		return nil, fmt.Errorf("user stats not found")
	}
	
	return stats, nil
}

// GetSessionStats gets statistics for a session
func (m *AdvancedTemplateMonitor) GetSessionStats(sessionID string) (*SessionStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats, ok := m.sessionStats[sessionID]
	if !ok {
		return nil, fmt.Errorf("session stats not found")
	}
	
	return stats, nil
}

// GetAlertHistory gets the alert history
func (m *AdvancedTemplateMonitor) GetAlertHistory() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.alertManager.alertHistory
}

// getSeverityForRiskScore gets the severity level for a risk score
func getSeverityForRiskScore(score float64) string {
	if score >= 0.8 {
		return "high"
	} else if score >= 0.5 {
		return "medium"
	} else {
		return "low"
	}
}

// getSeverityForAnomalyScore gets the severity level for an anomaly score
func getSeverityForAnomalyScore(score float64) string {
	if score >= 0.8 {
		return "high"
	} else if score >= 0.5 {
		return "medium"
	} else {
		return "low"
	}
}

// minFloat returns the minimum of two float64 values
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
