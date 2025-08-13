package analytics

import (
	"context"
	"fmt"
)

// LoggingHook logs collection events
type LoggingHook struct {
	logger Logger
}

func NewLoggingHook(logger Logger) *LoggingHook {
	return &LoggingHook{
		logger: logger,
	}
}

func (lh *LoggingHook) PreCollection(ctx context.Context, scanID string) error {
	lh.logger.Info("Starting metrics collection", "scanID", scanID, "timestamp", time.Now())
	return nil
}

func (lh *LoggingHook) PostCollection(ctx context.Context, scanID string, metrics []Metric) error {
	lh.logger.Info("Completed metrics collection", 
		"scanID", scanID,
		"metricsCount", len(metrics),
		"timestamp", time.Now())
	return nil
}

func (lh *LoggingHook) OnError(ctx context.Context, err error, scanID string) {
	lh.logger.Error("Metrics collection error", "error", err, "scanID", scanID)
}

// NotificationHook sends notifications for important events
type NotificationHook struct {
	notifier      Notifier
	thresholds    map[string]float64
	enabled       bool
}

type Notifier interface {
	SendNotification(ctx context.Context, message string, severity string) error
}

func NewNotificationHook(notifier Notifier, thresholds map[string]float64) *NotificationHook {
	return &NotificationHook{
		notifier:   notifier,
		thresholds: thresholds,
		enabled:    true,
	}
}

func (nh *NotificationHook) PreCollection(ctx context.Context, scanID string) error {
	if !nh.enabled {
		return nil
	}
	
	message := fmt.Sprintf("Starting security scan: %s", scanID)
	return nh.notifier.SendNotification(ctx, message, "info")
}

func (nh *NotificationHook) PostCollection(ctx context.Context, scanID string, metrics []Metric) error {
	if !nh.enabled {
		return nil
	}
	
	// Check for threshold violations
	for _, metric := range metrics {
		if threshold, exists := nh.thresholds[metric.Name]; exists {
			if metric.Value > threshold {
				message := fmt.Sprintf("Threshold exceeded for %s: %.2f > %.2f (Scan: %s)", 
					metric.Name, metric.Value, threshold, scanID)
				nh.notifier.SendNotification(ctx, message, "warning")
			}
		}
	}
	
	// Send completion notification
	vulnerabilityCount := nh.countVulnerabilities(metrics)
	severity := "info"
	if vulnerabilityCount > 5 {
		severity = "warning"
	}
	if vulnerabilityCount > 10 {
		severity = "critical"
	}
	
	message := fmt.Sprintf("Scan completed: %s - Found %d vulnerabilities", scanID, vulnerabilityCount)
	return nh.notifier.SendNotification(ctx, message, severity)
}

func (nh *NotificationHook) OnError(ctx context.Context, err error, scanID string) {
	if !nh.enabled {
		return
	}
	
	message := fmt.Sprintf("Scan error in %s: %v", scanID, err)
	nh.notifier.SendNotification(ctx, message, "error")
}

func (nh *NotificationHook) countVulnerabilities(metrics []Metric) int {
	count := 0
	for _, metric := range metrics {
		if metric.Name == "scan_vulnerabilities_found" {
			count += int(metric.Value)
		}
	}
	return count
}

// AuditHook maintains audit trail for compliance
type AuditHook struct {
	auditLogger   AuditLogger
	includeData   bool
	enabled       bool
}

type AuditLogger interface {
	LogAuditEvent(ctx context.Context, event AuditEvent) error
}

type AuditEvent struct {
	EventID     string                 `json:"event_id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	ScanID      string                 `json:"scan_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Details     map[string]interface{} `json:"details"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func NewAuditHook(auditLogger AuditLogger, includeData bool) *AuditHook {
	return &AuditHook{
		auditLogger: auditLogger,
		includeData: includeData,
		enabled:     true,
	}
}

func (ah *AuditHook) PreCollection(ctx context.Context, scanID string) error {
	if !ah.enabled {
		return nil
	}
	
	event := AuditEvent{
		EventID:   generateAuditID(),
		Timestamp: time.Now(),
		EventType: "metrics_collection_started",
		ScanID:    scanID,
		Details: map[string]interface{}{
			"action": "pre_collection",
		},
		Metadata: map[string]interface{}{
			"source": "analytics_collector",
		},
	}
	
	return ah.auditLogger.LogAuditEvent(ctx, event)
}

func (ah *AuditHook) PostCollection(ctx context.Context, scanID string, metrics []Metric) error {
	if !ah.enabled {
		return nil
	}
	
	details := map[string]interface{}{
		"action":        "post_collection",
		"metrics_count": len(metrics),
	}
	
	if ah.includeData {
		// Include summary of collected metrics
		metricSummary := ah.summarizeMetrics(metrics)
		details["metrics_summary"] = metricSummary
	}
	
	event := AuditEvent{
		EventID:   generateAuditID(),
		Timestamp: time.Now(),
		EventType: "metrics_collection_completed",
		ScanID:    scanID,
		Details:   details,
		Metadata: map[string]interface{}{
			"source": "analytics_collector",
		},
	}
	
	return ah.auditLogger.LogAuditEvent(ctx, event)
}

func (ah *AuditHook) OnError(ctx context.Context, err error, scanID string) {
	if !ah.enabled {
		return
	}
	
	event := AuditEvent{
		EventID:   generateAuditID(),
		Timestamp: time.Now(),
		EventType: "metrics_collection_error",
		ScanID:    scanID,
		Details: map[string]interface{}{
			"error":  err.Error(),
			"action": "error_handling",
		},
		Metadata: map[string]interface{}{
			"source": "analytics_collector",
		},
	}
	
	ah.auditLogger.LogAuditEvent(ctx, event)
}

func (ah *AuditHook) summarizeMetrics(metrics []Metric) map[string]interface{} {
	summary := make(map[string]interface{})
	
	// Count by type
	typeCounts := make(map[string]int)
	for _, metric := range metrics {
		typeCounts[string(metric.Type)]++
	}
	summary["type_distribution"] = typeCounts
	
	// Count by name prefix
	nameCounts := make(map[string]int)
	for _, metric := range metrics {
		prefix := getMetricPrefix(metric.Name)
		nameCounts[prefix]++
	}
	summary["name_distribution"] = nameCounts
	
	return summary
}

// PerformanceHook monitors collection performance
type PerformanceHook struct {
	performanceLogger Logger
	slowThreshold     time.Duration
	enabled           bool
	startTimes        map[string]time.Time
	mu                sync.RWMutex
}

func NewPerformanceHook(logger Logger, slowThreshold time.Duration) *PerformanceHook {
	return &PerformanceHook{
		performanceLogger: logger,
		slowThreshold:     slowThreshold,
		enabled:           true,
		startTimes:        make(map[string]time.Time),
	}
}

func (ph *PerformanceHook) PreCollection(ctx context.Context, scanID string) error {
	if !ph.enabled {
		return nil
	}
	
	ph.mu.Lock()
	ph.startTimes[scanID] = time.Now()
	ph.mu.Unlock()
	
	return nil
}

func (ph *PerformanceHook) PostCollection(ctx context.Context, scanID string, metrics []Metric) error {
	if !ph.enabled {
		return nil
	}
	
	ph.mu.Lock()
	startTime, exists := ph.startTimes[scanID]
	delete(ph.startTimes, scanID)
	ph.mu.Unlock()
	
	if !exists {
		return nil
	}
	
	duration := time.Since(startTime)
	
	if duration > ph.slowThreshold {
		ph.performanceLogger.Warn("Slow metrics collection detected",
			"scanID", scanID,
			"duration", duration,
			"metricsCount", len(metrics),
			"threshold", ph.slowThreshold)
	}
	
	ph.performanceLogger.Debug("Metrics collection performance",
		"scanID", scanID,
		"duration", duration,
		"metricsCount", len(metrics),
		"metricsPerSecond", float64(len(metrics))/duration.Seconds())
	
	return nil
}

func (ph *PerformanceHook) OnError(ctx context.Context, err error, scanID string) {
	if !ph.enabled {
		return
	}
	
	ph.mu.Lock()
	startTime, exists := ph.startTimes[scanID]
	delete(ph.startTimes, scanID)
	ph.mu.Unlock()
	
	duration := time.Duration(0)
	if exists {
		duration = time.Since(startTime)
	}
	
	ph.performanceLogger.Error("Metrics collection failed",
		"scanID", scanID,
		"duration", duration,
		"error", err)
}

// ComplianceHook ensures compliance with regulations
type ComplianceHook struct {
	regulations   []string
	validator     ComplianceValidator
	enabled       bool
}

type ComplianceValidator interface {
	ValidateCompliance(ctx context.Context, scanID string, metrics []Metric, regulations []string) error
}

func NewComplianceHook(validator ComplianceValidator, regulations []string) *ComplianceHook {
	return &ComplianceHook{
		regulations: regulations,
		validator:   validator,
		enabled:     true,
	}
}

func (ch *ComplianceHook) PreCollection(ctx context.Context, scanID string) error {
	if !ch.enabled {
		return nil
	}
	
	// Pre-collection compliance checks can be added here
	return nil
}

func (ch *ComplianceHook) PostCollection(ctx context.Context, scanID string, metrics []Metric) error {
	if !ch.enabled {
		return nil
	}
	
	return ch.validator.ValidateCompliance(ctx, scanID, metrics, ch.regulations)
}

func (ch *ComplianceHook) OnError(ctx context.Context, err error, scanID string) {
	// Compliance hooks typically don't need error handling
	// but can log compliance-related errors
}

// MetricsValidationHook validates metrics quality
type MetricsValidationHook struct {
	validator QualityValidator
	enabled   bool
}

type QualityValidator interface {
	ValidateQuality(ctx context.Context, metrics []Metric) []ValidationIssue
}

type ValidationIssue struct {
	MetricID    string `json:"metric_id"`
	Issue       string `json:"issue"`
	Severity    string `json:"severity"`
	Suggestion  string `json:"suggestion"`
}

func NewMetricsValidationHook(validator QualityValidator) *MetricsValidationHook {
	return &MetricsValidationHook{
		validator: validator,
		enabled:   true,
	}
}

func (mvh *MetricsValidationHook) PreCollection(ctx context.Context, scanID string) error {
	return nil
}

func (mvh *MetricsValidationHook) PostCollection(ctx context.Context, scanID string, metrics []Metric) error {
	if !mvh.enabled {
		return nil
	}
	
	issues := mvh.validator.ValidateQuality(ctx, metrics)
	
	if len(issues) > 0 {
		// Log validation issues but don't fail the collection
		for _, issue := range issues {
			// This would typically use the logger from the collector
			fmt.Printf("Metrics quality issue: %s - %s (Severity: %s)\n", 
				issue.MetricID, issue.Issue, issue.Severity)
		}
	}
	
	return nil
}

func (mvh *MetricsValidationHook) OnError(ctx context.Context, err error, scanID string) {
	// Quality validation hooks typically don't need error handling
}

// Utility functions for hooks

func generateAuditID() string {
	return fmt.Sprintf("audit_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func getMetricPrefix(name string) string {
	if len(name) == 0 {
		return "unknown"
	}
	
	// Extract prefix (first word before underscore)
	for i, char := range name {
		if char == '_' {
			return name[:i]
		}
	}
	
	return name
}

// Mock implementations for interfaces

// SimpleNotifier is a basic implementation of the Notifier interface
type SimpleNotifier struct {
	logger Logger
}

func NewSimpleNotifier(logger Logger) *SimpleNotifier {
	return &SimpleNotifier{logger: logger}
}

func (sn *SimpleNotifier) SendNotification(ctx context.Context, message string, severity string) error {
	sn.logger.Info("Notification", "message", message, "severity", severity)
	return nil
}

// SimpleAuditLogger is a basic implementation of the AuditLogger interface
type SimpleAuditLogger struct {
	logger Logger
}

func NewSimpleAuditLogger(logger Logger) *SimpleAuditLogger {
	return &SimpleAuditLogger{logger: logger}
}

func (sal *SimpleAuditLogger) LogAuditEvent(ctx context.Context, event AuditEvent) error {
	sal.logger.Info("Audit Event",
		"eventID", event.EventID,
		"eventType", event.EventType,
		"scanID", event.ScanID,
		"timestamp", event.Timestamp)
	return nil
}

// SimpleComplianceValidator is a basic implementation of the ComplianceValidator interface
type SimpleComplianceValidator struct {
	logger Logger
}

func NewSimpleComplianceValidator(logger Logger) *SimpleComplianceValidator {
	return &SimpleComplianceValidator{logger: logger}
}

func (scv *SimpleComplianceValidator) ValidateCompliance(ctx context.Context, scanID string, metrics []Metric, regulations []string) error {
	// Basic compliance validation
	for _, regulation := range regulations {
		switch regulation {
		case "GDPR":
			if err := scv.validateGDPR(metrics); err != nil {
				return fmt.Errorf("GDPR compliance violation: %w", err)
			}
		case "SOX":
			if err := scv.validateSOX(metrics); err != nil {
				return fmt.Errorf("SOX compliance violation: %w", err)
			}
		}
	}
	
	scv.logger.Info("Compliance validation passed", "scanID", scanID, "regulations", regulations)
	return nil
}

func (scv *SimpleComplianceValidator) validateGDPR(metrics []Metric) error {
	// Check for PII in metrics
	for _, metric := range metrics {
		for _, value := range metric.Labels {
			if containsPII(value) {
				return fmt.Errorf("potential PII found in metric labels")
			}
		}
	}
	return nil
}

func (scv *SimpleComplianceValidator) validateSOX(metrics []Metric) error {
	// Check for financial data handling compliance
	for _, metric := range metrics {
		if strings.Contains(metric.Name, "financial") || strings.Contains(metric.Name, "payment") {
			// Additional SOX-specific validations would go here
		}
	}
	return nil
}

// SimpleQualityValidator is a basic implementation of the QualityValidator interface
type SimpleQualityValidator struct{}

func NewSimpleQualityValidator() *SimpleQualityValidator {
	return &SimpleQualityValidator{}
}

func (sqv *SimpleQualityValidator) ValidateQuality(ctx context.Context, metrics []Metric) []ValidationIssue {
	var issues []ValidationIssue
	
	for _, metric := range metrics {
		// Check for missing required fields
		if metric.Name == "" {
			issues = append(issues, ValidationIssue{
				MetricID:   metric.ID,
				Issue:      "Missing metric name",
				Severity:   "high",
				Suggestion: "Ensure all metrics have a valid name",
			})
		}
		
		// Check for suspicious values
		if metric.Value < 0 && !strings.Contains(metric.Name, "delta") {
			issues = append(issues, ValidationIssue{
				MetricID:   metric.ID,
				Issue:      "Negative value for non-delta metric",
				Severity:   "medium",
				Suggestion: "Verify that negative values are expected for this metric",
			})
		}
		
		// Check for missing labels on certain metric types
		if metric.Type == MetricTypeEvent && len(metric.Labels) == 0 {
			issues = append(issues, ValidationIssue{
				MetricID:   metric.ID,
				Issue:      "Event metric missing labels",
				Severity:   "low",
				Suggestion: "Add relevant labels to improve metric categorization",
			})
		}
	}
	
	return issues
}

// Utility function to detect PII (basic implementation)
func containsPII(value string) bool {
	// Basic PII detection patterns
	piiPatterns := []string{
		"email", "ssn", "social", "credit", "card", "phone", "address",
	}
	
	lowerValue := strings.ToLower(value)
	for _, pattern := range piiPatterns {
		if strings.Contains(lowerValue, pattern) {
			return true
		}
	}
	
	return false
}