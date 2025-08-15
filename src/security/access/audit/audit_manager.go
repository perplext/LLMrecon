package audit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AuditManager manages audit logging across the system
type AuditManager struct {
	logger *AuditLogger
	mutex  sync.RWMutex

// NewAuditManager creates a new audit manager
func NewAuditManager(logger *AuditLogger) *AuditManager {
	return &AuditManager{
		logger: logger,
	}

// LogAccess logs an access event
func (m *AuditManager) LogAccess(ctx context.Context, userID, resource, action string) error {
	if m.logger == nil {
		return fmt.Errorf("audit logger not configured")
	}
	
	m.logger.LogEventWithStatus("access", "AuditManager", userID, "info", map[string]interface{}{
		"resource": resource,
		"action":   action,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	
	return nil

// LogSecurity logs a security event
func (m *AuditManager) LogSecurity(ctx context.Context, eventType, details string) error {
	if m.logger == nil {
		return fmt.Errorf("audit logger not configured")
	}
	
	m.logger.LogEventWithStatus("security", "AuditManager", eventType, "warning", map[string]interface{}{
		"details": details,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	
