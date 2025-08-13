package sandbox

import (
	"sync"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/security"
)

// MetricsCollector collects metrics for the template security framework
type MetricsCollector struct {
	// mutex is used for thread safety
	mutex sync.RWMutex

	// Validation metrics
	validationCount      int64
	validationErrors     int64
	validationTime       time.Duration
	issuesByType         map[security.SecurityIssueType]int64
	issuesBySeverity     map[common.SeverityLevel]int64
	templatesByRisk      map[RiskCategory]int64
	averageValidationTime time.Duration

	// Execution metrics
	executionCount      int64
	executionErrors     int64
	executionTime       time.Duration
	resourceUsage       ResourceUsage
	averageExecutionTime time.Duration
	averageResourceUsage ResourceUsage

	// Workflow metrics
	templateVersions    int64
	approvedTemplates   int64
	rejectedTemplates   int64
	pendingTemplates    int64
	deprecatedTemplates int64

	// Alerts
	alerts []Alert
}

// Alert represents a security alert
type Alert struct {
	// Timestamp is the time of the alert
	Timestamp time.Time
	// Level is the alert level
	Level AlertLevel
	// Message is the alert message
	Message string
	// TemplateID is the ID of the template that triggered the alert
	TemplateID string
	// Issues are the security issues that triggered the alert
	Issues []*security.SecurityIssue
}

// AlertLevel represents the level of an alert
type AlertLevel string

const (
	// AlertLevelInfo is an informational alert
	AlertLevelInfo AlertLevel = "info"
	// AlertLevelWarning is a warning alert
	AlertLevelWarning AlertLevel = "warning"
	// AlertLevelError is an error alert
	AlertLevelError AlertLevel = "error"
	// AlertLevelCritical is a critical alert
	AlertLevelCritical AlertLevel = "critical"
)

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		mutex:               sync.RWMutex{},
		issuesByType:        make(map[security.SecurityIssueType]int64),
		issuesBySeverity:    make(map[common.SeverityLevel]int64),
		templatesByRisk:     make(map[RiskCategory]int64),
		resourceUsage:       ResourceUsage{},
		averageResourceUsage: ResourceUsage{},
		alerts:              []Alert{},
	}
}

// RecordValidation records a template validation
func (m *MetricsCollector) RecordValidation(result *ValidationResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Update validation count
	m.validationCount++

	// Update validation time
	m.validationTime += result.ValidationTime
	m.averageValidationTime = time.Duration(int64(m.validationTime) / m.validationCount)

	// Update issues by type and severity
	for _, issue := range result.Issues {
		m.issuesByType[issue.Type]++
		m.issuesBySeverity[issue.Severity]++
	}

	// Update templates by risk
	m.templatesByRisk[result.RiskScore.Category]++

	// Create alerts for high and critical issues
	if result.HasCriticalIssues() {
		m.createAlert(AlertLevelCritical, "Template has critical security issues", result.Template.ID, result.Issues)
	} else if result.HasHighIssues() {
		m.createAlert(AlertLevelWarning, "Template has high severity security issues", result.Template.ID, result.Issues)
	}
}

// RecordExecution records a template execution
func (m *MetricsCollector) RecordExecution(result *ExecutionResult, templateID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Update execution count
	m.executionCount++

	// Update execution time
	m.executionTime += result.ExecutionTime
	m.averageExecutionTime = time.Duration(int64(m.executionTime) / m.executionCount)

	// Update resource usage
	m.resourceUsage.CPUTime += result.ResourceUsage.CPUTime
	m.resourceUsage.MemoryUsage += result.ResourceUsage.MemoryUsage
	m.averageResourceUsage.CPUTime = m.resourceUsage.CPUTime / float64(m.executionCount)
	m.averageResourceUsage.MemoryUsage = m.resourceUsage.MemoryUsage / int64(m.executionCount)

	// Update execution errors
	if !result.Success {
		m.executionErrors++
		m.createAlert(AlertLevelError, "Template execution failed: "+result.Error, templateID, result.SecurityIssues)
	}

	// Create alerts for resource usage
	if result.ResourceUsage.CPUTime > 5.0 {
		m.createAlert(AlertLevelWarning, "Template has high CPU usage", templateID, nil)
	}
	if result.ResourceUsage.MemoryUsage > 500 {
		m.createAlert(AlertLevelWarning, "Template has high memory usage", templateID, nil)
	}
}

// RecordWorkflowAction records a workflow action
func (m *MetricsCollector) RecordWorkflowAction(action string, version *TemplateVersion) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Update template versions
	m.templateVersions++

	// Update template status counts
	switch version.Status {
	case StatusApproved:
		m.approvedTemplates++
	case StatusRejected:
		m.rejectedTemplates++
	case StatusPendingReview:
		m.pendingTemplates++
	case StatusDeprecated:
		m.deprecatedTemplates++
	}

	// Create alerts for high and critical risk templates in the workflow
	if version.RiskScore != nil {
		if version.RiskScore.Category == RiskCategoryCritical {
			m.createAlert(AlertLevelCritical, "Critical risk template in workflow: "+action, version.TemplateID, version.SecurityIssues)
		} else if version.RiskScore.Category == RiskCategoryHigh {
			m.createAlert(AlertLevelWarning, "High risk template in workflow: "+action, version.TemplateID, version.SecurityIssues)
		}
	}
}

// GetValidationMetrics gets the validation metrics
func (m *MetricsCollector) GetValidationMetrics() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"validationCount":       m.validationCount,
		"validationErrors":      m.validationErrors,
		"averageValidationTime": m.averageValidationTime.String(),
		"issuesByType":          m.issuesByType,
		"issuesBySeverity":      m.issuesBySeverity,
		"templatesByRisk":       m.templatesByRisk,
	}
}

// GetExecutionMetrics gets the execution metrics
func (m *MetricsCollector) GetExecutionMetrics() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"executionCount":       m.executionCount,
		"executionErrors":      m.executionErrors,
		"averageExecutionTime": m.averageExecutionTime.String(),
		"averageResourceUsage": map[string]interface{}{
			"cpuTime":     m.averageResourceUsage.CPUTime,
			"memoryUsage": m.averageResourceUsage.MemoryUsage,
		},
	}
}

// GetWorkflowMetrics gets the workflow metrics
func (m *MetricsCollector) GetWorkflowMetrics() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"templateVersions":    m.templateVersions,
		"approvedTemplates":   m.approvedTemplates,
		"rejectedTemplates":   m.rejectedTemplates,
		"pendingTemplates":    m.pendingTemplates,
		"deprecatedTemplates": m.deprecatedTemplates,
	}
}

// GetAlerts gets the alerts
func (m *MetricsCollector) GetAlerts() []Alert {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.alerts
}

// GetAlertsByLevel gets alerts by level
func (m *MetricsCollector) GetAlertsByLevel(level AlertLevel) []Alert {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var alerts []Alert
	for _, alert := range m.alerts {
		if alert.Level == level {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// ClearAlerts clears all alerts
func (m *MetricsCollector) ClearAlerts() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.alerts = []Alert{}
}

// createAlert creates a new alert
func (m *MetricsCollector) createAlert(level AlertLevel, message string, templateID string, issues []*security.SecurityIssue) {
	alert := Alert{
		Timestamp:  time.Now(),
		Level:      level,
		Message:    message,
		TemplateID: templateID,
		Issues:     issues,
	}

	m.alerts = append(m.alerts, alert)
}

// ResetMetrics resets all metrics
func (m *MetricsCollector) ResetMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.validationCount = 0
	m.validationErrors = 0
	m.validationTime = 0
	m.issuesByType = make(map[security.SecurityIssueType]int64)
	m.issuesBySeverity = make(map[common.SeverityLevel]int64)
	m.templatesByRisk = make(map[RiskCategory]int64)
	m.averageValidationTime = 0

	m.executionCount = 0
	m.executionErrors = 0
	m.executionTime = 0
	m.resourceUsage = ResourceUsage{}
	m.averageExecutionTime = 0
	m.averageResourceUsage = ResourceUsage{}

	m.templateVersions = 0
	m.approvedTemplates = 0
	m.rejectedTemplates = 0
	m.pendingTemplates = 0
	m.deprecatedTemplates = 0

	m.alerts = []Alert{}
}
