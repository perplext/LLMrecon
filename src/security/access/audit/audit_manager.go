package audit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/common"
)

// AuditManager manages audit logging across the system
type AuditManager struct {
	logger *AuditLogger
	mutex  sync.RWMutex
}

// NewAuditManager creates a new audit manager
func NewAuditManager(logger *AuditLogger) *AuditManager {
	return &AuditManager{
		logger: logger,
	}
}

// LogAccess logs an access event
func (m *AuditManager) LogAccess(ctx context.Context, userID, resource, action string) error {
	if m.logger == nil {
		return fmt.Errorf("audit logger not configured")
	}

	m.logger.LogEventWithStatus("access", "AuditManager", userID, "info", map[string]interface{}{
		"resource":  resource,
		"action":    action,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	return nil
}

// LogSecurity logs a security event
func (m *AuditManager) LogSecurity(ctx context.Context, eventType, details string) error {
	if m.logger == nil {
		return fmt.Errorf("audit logger not configured")
	}

	m.logger.LogEventWithStatus("security", "AuditManager", eventType, "warning", map[string]interface{}{
		"details":   details,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	return nil
}

// LogEvent logs an audit event
func (m *AuditManager) LogEvent(ctx context.Context, event *AuditEvent) error {
	if m.logger == nil {
		return fmt.Errorf("audit logger not configured")
	}

	// Convert AuditEvent to the format expected by the logger
	details := map[string]interface{}{
		"timestamp":   event.Timestamp.Format(time.RFC3339),
		"action":      string(event.Action),
		"resource":    event.Resource,
		"resource_id": event.ResourceID,
		"description": event.Description,
		"ip_address":  event.IPAddress,
		"user_agent":  event.UserAgent,
		"session_id":  event.SessionID,
		"request_id":  event.RequestID,
		"duration_ms": event.DurationMs,
	}

	// Add metadata
	for key, value := range event.Metadata {
		details[key] = value
	}

	// Add changes if present
	if event.Changes != nil {
		details["changes"] = event.Changes
	}

	// Use the severity level from the event to determine log level
	var logLevel string
	switch event.Severity {
	case common.AuditSeverityCritical, common.AuditSeverityEmergency, common.AuditSeverityAlert:
		logLevel = "critical"
	case common.AuditSeverityError, common.AuditSeverityHigh:
		logLevel = "error"
	case common.AuditSeverityWarning, common.AuditSeverityMedium:
		logLevel = "warning"
	case common.AuditSeverityInfo, common.AuditSeverityNotice, common.AuditSeverityLow, common.AuditSeverityDebug:
		logLevel = "info"
	default:
		logLevel = "info"
	}

	m.logger.LogEventWithStatus(string(event.Action), "AuditManager", event.UserID, logLevel, details)

	return nil
}
