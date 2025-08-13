// Package audit provides comprehensive security audit logging functionality
package audit

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/security/access/common"
)

// AuditManager is responsible for managing security audit logging
type AuditManager struct {
	config     *AuditConfig
	loggers    []AuditLogger
	mu         sync.RWMutex
	alertRules []AlertRule
}

// AuditConfig defines the configuration for audit logging
type AuditConfig struct {
	// Whether to enable audit logging
	Enabled bool `json:"enabled"`
	
	// Minimum severity level to log
	MinSeverity common.AuditSeverity `json:"min_severity"`
	
	// File path for audit logs
	LogFilePath string `json:"log_file_path"`
	
	// Whether to log to syslog
	EnableSyslog bool `json:"enable_syslog"`
	
	// Whether to log to database
	EnableDatabase bool `json:"enable_database"`
	
	// Database connection string
	DatabaseURL string `json:"database_url"`
	
	// Log retention period in days
	RetentionDays int `json:"retention_days"`
	
	// Whether to enable real-time alerting
	EnableAlerts bool `json:"enable_alerts"`
	
	// Whether to encrypt audit logs
	EncryptLogs bool `json:"encrypt_logs"`
	
	// Encryption key for audit logs (if encryption is enabled)
	EncryptionKey string `json:"encryption_key,omitempty"`
	
	// Whether to sign audit logs for integrity verification
	SignLogs bool `json:"sign_logs"`
	
	// Signing key ID for audit logs (if signing is enabled)
	SigningKeyID string `json:"signing_key_id,omitempty"`
	
	// Whether to include sensitive data in audit logs
	IncludeSensitiveData bool `json:"include_sensitive_data"`
	
	// List of fields to redact from audit logs
	RedactFields []string `json:"redact_fields"`
	
	// Whether to compress audit logs
	CompressLogs bool `json:"compress_logs"`
	
	// Maximum size of a single audit log file in MB
	MaxLogFileSize int `json:"max_log_file_size"`
	
	// Maximum number of audit log files to keep
	MaxLogFiles int `json:"max_log_files"`
}

// AlertRule defines a rule for triggering alerts based on audit events
type AlertRule struct {
	// Unique identifier for the rule
	ID string `json:"id"`
	
	// Human-readable name for the rule
	Name string `json:"name"`
	
	// Description of the rule
	Description string `json:"description"`
	
	// Minimum severity level to trigger the alert
	MinSeverity common.AuditSeverity `json:"min_severity"`
	
	// Actions that should trigger the alert
	Actions []common.AuditAction `json:"actions"`
	
	// Resources that should trigger the alert
	Resources []string `json:"resources"`
	
	// Alert notification channels
	Channels []AlertChannel `json:"channels"`
	
	// Whether the rule is enabled
	Enabled bool `json:"enabled"`
	
	// Cooldown period between alerts in seconds
	CooldownSeconds int `json:"cooldown_seconds"`
	
	// Last time this rule was triggered
	lastTriggered time.Time
}

// AlertChannel defines a notification channel for security alerts
type AlertChannel struct {
	// Type of alert channel (email, sms, webhook, etc.)
	Type string `json:"type"`
	
	// Configuration for the alert channel
	Config map[string]interface{} `json:"config"`
}

// NewAuditManager creates a new audit manager
func NewAuditManager(config *AuditConfig) (*AuditManager, error) {
	if config == nil {
		config = DefaultAuditConfig()
	}

	manager := &AuditManager{
		config:     config,
		loggers:    make([]AuditLogger, 0),
		alertRules: make([]AlertRule, 0),
	}

	// Initialize file logger if enabled
	if config.LogFilePath != "" {
		// Ensure directory exists
		logDir := filepath.Dir(config.LogFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		fileLogger, err := NewFileAuditLogger(config.LogFilePath, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create file logger: %w", err)
		}
		manager.loggers = append(manager.loggers, fileLogger)
	}

	// Initialize syslog logger if enabled
	if config.EnableSyslog {
		syslogLogger, err := NewSyslogAuditLogger(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create syslog logger: %w", err)
		}
		manager.loggers = append(manager.loggers, syslogLogger)
	}

	// Initialize database logger if enabled
	if config.EnableDatabase && config.DatabaseURL != "" {
		dbLogger, err := NewDatabaseAuditLogger(config.DatabaseURL, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create database logger: %w", err)
		}
		manager.loggers = append(manager.loggers, dbLogger)
	}

	// Always add in-memory logger for immediate querying
	memLogger := NewInMemoryAuditLogger(config)
	manager.loggers = append(manager.loggers, memLogger)

	// Load default alert rules
	manager.loadDefaultAlertRules()

	return manager, nil
}

// loadDefaultAlertRules loads the default alert rules
func (m *AuditManager) loadDefaultAlertRules() {
	// Rule for failed authentication attempts
	m.alertRules = append(m.alertRules, AlertRule{
		ID:             "auth-failures",
		Name:           "Authentication Failures",
		Description:    "Alert on multiple failed authentication attempts",
		MinSeverity:    common.AuditSeverityMedium,
		Actions:        []common.AuditAction{common.AuditActionLoginFailed},
		Enabled:        true,
		CooldownSeconds: 300, // 5 minutes
	})

	// Rule for privilege escalation
	m.alertRules = append(m.alertRules, AlertRule{
		ID:             "privilege-escalation",
		Name:           "Privilege Escalation",
		Description:    "Alert on privilege escalation attempts",
		MinSeverity:    common.AuditSeverityHigh,
		Actions:        []common.AuditAction{common.AuditActionRoleChange},
		Enabled:        true,
		CooldownSeconds: 60, // 1 minute
	})

	// Rule for security configuration changes
	m.alertRules = append(m.alertRules, AlertRule{
		ID:             "security-config-change",
		Name:           "Security Configuration Change",
		Description:    "Alert on security configuration changes",
		MinSeverity:    common.AuditSeverityHigh,
		Actions:        []common.AuditAction{common.AuditActionUserUpdate},
		Resources:      []string{"security_config", "access_control", "mfa_config"},
		Enabled:        true,
		CooldownSeconds: 300, // 5 minutes
	})

	// Rule for critical operations
	m.alertRules = append(m.alertRules, AlertRule{
		ID:             "critical-operation",
		Name:           "Critical Operation",
		Description:    "Alert on critical operations",
		MinSeverity:    common.AuditSeverityCritical,
		Enabled:        true,
		CooldownSeconds: 0, // No cooldown for critical alerts
	})
}

// LogAudit logs an audit event to all configured loggers
func (m *AuditManager) LogAudit(ctx context.Context, log *AuditEvent) error {
	if !m.config.Enabled {
		return nil
	}

	// Skip logging if severity is below minimum
	if severityLevel(log.Severity) < severityLevel(m.config.MinSeverity) {
		return nil
	}

	// Generate ID if not provided
	if log.ID == "" {
		id, err := generateRandomID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
		log.ID = id
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now().UTC()
	}

	// Redact sensitive fields if configured
	if !m.config.IncludeSensitiveData && len(m.config.RedactFields) > 0 {
		if log.Metadata != nil {
			for _, field := range m.config.RedactFields {
				if _, exists := log.Metadata[field]; exists {
					log.Metadata[field] = "[REDACTED]"
				}
			}
		}
	}

	// Sign the log if configured
	if m.config.SignLogs && m.config.SigningKeyID != "" {
		signature, err := m.signAuditLog(log)
		if err != nil {
			return fmt.Errorf("failed to sign audit log: %w", err)
		}
		if log.Metadata == nil {
			log.Metadata = make(map[string]interface{})
		}
		log.Metadata["signature"] = signature
	}

	// Log to all configured loggers
	var lastErr error
	for _, logger := range m.loggers {
		if err := logger.LogAudit(ctx, log); err != nil {
			lastErr = err
		}
	}

	// Check alert rules
	if m.config.EnableAlerts {
		m.checkAlertRules(ctx, log)
	}

	return lastErr
}

// signAuditLog signs an audit log for integrity verification
func (m *AuditManager) signAuditLog(log *AuditEvent) (string, error) {
	// In a real implementation, this would use a cryptographic signing mechanism
	// For now, we'll just return a placeholder
	return "signed-" + log.ID, nil
}

// checkAlertRules checks if any alert rules should be triggered for this audit event
func (m *AuditManager) checkAlertRules(ctx context.Context, log *AuditEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()

	for i, rule := range m.alertRules {
		if !rule.Enabled {
			continue
		}

		// Skip if severity is below the rule's minimum
		if severityLevel(log.Severity) < severityLevel(rule.MinSeverity) {
			continue
		}

		// Skip if cooldown period hasn't elapsed
		if !rule.lastTriggered.IsZero() && now.Sub(rule.lastTriggered).Seconds() < float64(rule.CooldownSeconds) {
			continue
		}

		// Check if action matches
		actionMatches := len(rule.Actions) == 0
		for _, action := range rule.Actions {
			if log.Action == action {
				actionMatches = true
				break
			}
		}
		if !actionMatches {
			continue
		}

		// Check if resource matches
		resourceMatches := len(rule.Resources) == 0
		for _, resource := range rule.Resources {
			if log.Resource == resource {
				resourceMatches = true
				break
			}
		}
		if !resourceMatches {
			continue
		}

		// Rule matches, trigger alert
		m.triggerAlert(ctx, rule, log)

		// Update last triggered time
		m.alertRules[i].lastTriggered = now
	}
}

// triggerAlert triggers an alert for a matching rule
func (m *AuditManager) triggerAlert(ctx context.Context, rule AlertRule, log *AuditEvent) {
	// In a real implementation, this would send alerts through configured channels
	// For now, we'll just log the alert
	alertLog := &AuditEvent{
		Action:      common.AuditActionAlert,
		Resource:    "alert_rule",
		ResourceID:  rule.ID,
		Description: fmt.Sprintf("Alert triggered: %s", rule.Name),
		Severity:    log.Severity,
		UserID:      log.UserID,
		Username:    log.Username,
		IPAddress:   log.IPAddress,
		Metadata: map[string]interface{}{
			"rule_id":          rule.ID,
			"rule_name":        rule.Name,
			"rule_description": rule.Description,
			"trigger_log_id":   log.ID,
		},
	}

	// Log the alert as a separate audit event
	for _, logger := range m.loggers {
		logger.LogAudit(ctx, alertLog)
	}
}

// QueryAuditLogs queries audit logs based on filters
func (m *AuditManager) QueryAuditLogs(ctx context.Context, filter *AuditQueryFilter) ([]*AuditEvent, error) {
	// Use the in-memory logger for queries by default
	// In a real implementation, this would query the appropriate storage backend
	for _, logger := range m.loggers {
		if memLogger, ok := logger.(*InMemoryAuditLogger); ok {
			return memLogger.QueryAuditLogs(ctx, filter)
		}
	}

	return nil, fmt.Errorf("no suitable logger found for querying")
}

// GetAuditLog retrieves a specific audit log by ID
func (m *AuditManager) GetAuditLog(ctx context.Context, id string) (*AuditEvent, error) {
	// Use the in-memory logger for queries by default
	for _, logger := range m.loggers {
		if memLogger, ok := logger.(*InMemoryAuditLogger); ok {
			return memLogger.GetAuditLog(ctx, id)
		}
	}

	return nil, fmt.Errorf("no suitable logger found for querying")
}

// AddAlertRule adds a new alert rule
func (m *AuditManager) AddAlertRule(rule AlertRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if rule with same ID already exists
	for _, r := range m.alertRules {
		if r.ID == rule.ID {
			return fmt.Errorf("alert rule with ID %s already exists", rule.ID)
		}
	}

	// Add the rule
	m.alertRules = append(m.alertRules, rule)
	return nil
}

// UpdateAlertRule updates an existing alert rule
func (m *AuditManager) UpdateAlertRule(rule AlertRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and update the rule
	for i, r := range m.alertRules {
		if r.ID == rule.ID {
			m.alertRules[i] = rule
			return nil
		}
	}

	return fmt.Errorf("alert rule with ID %s not found", rule.ID)
}

// DeleteAlertRule deletes an alert rule
func (m *AuditManager) DeleteAlertRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and delete the rule
	for i, r := range m.alertRules {
		if r.ID == id {
			m.alertRules = append(m.alertRules[:i], m.alertRules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("alert rule with ID %s not found", id)
}

// GetAlertRules returns all alert rules
func (m *AuditManager) GetAlertRules() []AlertRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent modification
	rules := make([]AlertRule, len(m.alertRules))
	copy(rules, m.alertRules)
	return rules
}

// EnableAlertRule enables an alert rule
func (m *AuditManager) EnableAlertRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and enable the rule
	for i, r := range m.alertRules {
		if r.ID == id {
			m.alertRules[i].Enabled = true
			return nil
		}
	}

	return fmt.Errorf("alert rule with ID %s not found", id)
}

// DisableAlertRule disables an alert rule
func (m *AuditManager) DisableAlertRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and disable the rule
	for i, r := range m.alertRules {
		if r.ID == id {
			m.alertRules[i].Enabled = false
			return nil
		}
	}

	return fmt.Errorf("alert rule with ID %s not found", id)
}

// Close closes the audit manager and all loggers
func (m *AuditManager) Close() error {
	var lastErr error
	for _, logger := range m.loggers {
		if closer, ok := logger.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

// DefaultAuditConfig returns the default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Enabled:             true,
		MinSeverity:         common.AuditSeverityInfo,
		LogFilePath:         "logs/audit.log",
		EnableSyslog:        false,
		EnableDatabase:      false,
		RetentionDays:       90,
		EnableAlerts:        true,
		EncryptLogs:         false,
		SignLogs:            false,
		IncludeSensitiveData: false,
		RedactFields:        []string{"password", "token", "secret", "key"},
		CompressLogs:        true,
		MaxLogFileSize:      10, // 10 MB
		MaxLogFiles:         5,
	}
}

// severityLevel returns the numeric level for a severity
func severityLevel(severity common.AuditSeverity) int {
	switch severity {
	case common.AuditSeverityInfo:
		return 1
	case common.AuditSeverityLow:
		return 2
	case common.AuditSeverityMedium:
		return 3
	case common.AuditSeverityHigh:
		return 4
	case common.AuditSeverityCritical:
		return 5
	default:
		return 0
	}
}

// generateRandomID generates a random ID for audit logs
func generateRandomID() (string, error) {
	// In a real implementation, this would use a secure random ID generator
	return fmt.Sprintf("audit-%d", time.Now().UnixNano()), nil
}
