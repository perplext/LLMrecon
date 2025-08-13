// Package adapter provides adapters for the access control system
package adapter

import (
	"context"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// AuditEvent represents an entry in the audit log
type AuditEvent struct {
	ID          string
	Timestamp   time.Time
	UserID      string
	Username    string
	Action      string
	Resource    string
	ResourceID  string
	Description string
	IPAddress   string
	UserAgent   string
	Severity    string
	Status      string
	SessionID   string
	Details     map[string]interface{}
	Changes     map[string]interface{}
	Metadata    map[string]interface{}
}

// AuditEventFilter represents a filter for querying audit logs
type AuditEventFilter struct {
	UserID     string
	Username   string
	Action     string
	Resource   string
	ResourceID string
	IPAddress  string
	Severity   string
	Status     string
	SessionID  string
	StartTime  time.Time
	EndTime    time.Time
	Offset     int
	Limit      int
}

// ModelsAuditStore defines the interface for audit storage operations using models.AuditEvent
type ModelsAuditStore interface {
	LogEvent(ctx context.Context, event *AuditEvent) error
	GetEventByID(ctx context.Context, id string) (*AuditEvent, error)
	QueryEvents(ctx context.Context, filter *AuditEventFilter, offset, limit int) ([]*AuditEvent, int, error)
	ExportEvents(ctx context.Context, filter *AuditEventFilter, format string) (string, error)
	Close() error
}

// AuditStoreAdapter adapts a models.AuditEvent store to the AuditStore interface
type AuditStoreAdapter struct {
	store ModelsAuditStore
}

// NewAuditStoreAdapter creates a new adapter for models.AuditEvent store
func NewAuditStoreAdapter(store ModelsAuditStore) interfaces.AuditLogger {
	return &AuditStoreAdapter{
		store: store,
	}
}

// convertModelsAuditLogToAuditEvent converts a models.AuditLog to an AuditEvent
func convertModelsAuditLogToAuditEvent(log *models.AuditLog) *AuditEvent {
	if log == nil {
		return nil
	}

	return &AuditEvent{
		ID:          log.ID,
		Timestamp:   log.Timestamp,
		UserID:      log.UserID,
		Username:    log.Username,
		Action:      log.Action,
		Resource:    log.Resource,
		ResourceID:  log.ResourceID,
		Description: log.Description,
		IPAddress:   log.IPAddress,
		UserAgent:   log.UserAgent,
		Severity:    "info",        // Default severity
		Status:      log.Status,
		SessionID:   "",            // Not available in models.AuditLog
		Details:     log.Metadata,
		Changes:     nil,           // Not available in models.AuditLog
		Metadata:    log.Metadata,
	}
}

// convertAuditEventToModelsAuditLog converts an AuditEvent to a models.AuditLog
func convertAuditEventToModelsAuditLog(event *AuditEvent) *models.AuditLog {
	if event == nil {
		return nil
	}

	return &models.AuditLog{
		ID:           event.ID,
		UserID:       event.UserID,
		Username:     event.Username,
		Action:       event.Action,
		Resource:     event.Resource,
		ResourceType: "",              // Not available in AuditEvent
		ResourceID:   event.ResourceID,
		Status:       event.Status,
		Description:  event.Description,
		IPAddress:    event.IPAddress,
		UserAgent:    event.UserAgent,
		Timestamp:    event.Timestamp,
		Metadata:     event.Metadata,
	}
}

// convertAuditEventsToModelsAuditLogs converts a slice of AuditEvent to a slice of models.AuditLog
func convertAuditEventsToModelsAuditLogs(events []*AuditEvent) []*models.AuditLog {
	if events == nil {
		return nil
	}
	result := make([]*models.AuditLog, len(events))
	for i, event := range events {
		result[i] = convertAuditEventToModelsAuditLog(event)
	}
	return result
}

// convertFilterMapToAuditEventFilter converts a map[string]interface{} filter to an AuditEventFilter
func convertFilterMapToAuditEventFilter(filter map[string]interface{}) *AuditEventFilter {
	result := &AuditEventFilter{}

	if userID, ok := filter["user_id"].(string); ok {
		result.UserID = userID
	}
	if username, ok := filter["username"].(string); ok {
		result.Username = username
	}
	if action, ok := filter["action"].(string); ok {
		result.Action = action
	}
	if resource, ok := filter["resource"].(string); ok {
		result.Resource = resource
	}
	if resourceID, ok := filter["resource_id"].(string); ok {
		result.ResourceID = resourceID
	}
	if ipAddress, ok := filter["ip_address"].(string); ok {
		result.IPAddress = ipAddress
	}
	if severity, ok := filter["severity"].(string); ok {
		result.Severity = severity
	}
	if status, ok := filter["status"].(string); ok {
		result.Status = status
	}
	if sessionID, ok := filter["session_id"].(string); ok {
		result.SessionID = sessionID
	}
	if startTime, ok := filter["start_time"].(time.Time); ok {
		result.StartTime = startTime
	}
	if endTime, ok := filter["end_time"].(time.Time); ok {
		result.EndTime = endTime
	}
	if offset, ok := filter["offset"].(int); ok {
		result.Offset = offset
	}
	if limit, ok := filter["limit"].(int); ok {
		result.Limit = limit
	}

	return result
}

// LogEvent logs an audit event
func (a *AuditStoreAdapter) LogEvent(ctx context.Context, event *models.AuditLog) error {
	return a.store.LogEvent(ctx, convertModelsAuditLogToAuditEvent(event))
}

// GetEventByID retrieves an audit event by ID
func (a *AuditStoreAdapter) GetEventByID(ctx context.Context, id string) (*models.AuditLog, error) {
	event, err := a.store.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertAuditEventToModelsAuditLog(event), nil
}

// QueryEvents queries audit events with filtering
func (a *AuditStoreAdapter) QueryEvents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	eventFilter := convertFilterMapToAuditEventFilter(filter)
	eventFilter.Offset = offset
	eventFilter.Limit = limit

	events, count, err := a.store.QueryEvents(ctx, eventFilter, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return convertAuditEventsToModelsAuditLogs(events), count, nil
}

// ExportEvents exports audit events to a file
func (a *AuditStoreAdapter) ExportEvents(ctx context.Context, filter map[string]interface{}, format string) (string, error) {
	eventFilter := convertFilterMapToAuditEventFilter(filter)
	return a.store.ExportEvents(ctx, eventFilter, format)
}

// Close closes the audit store
func (a *AuditStoreAdapter) Close() error {
	return a.store.Close()
}
