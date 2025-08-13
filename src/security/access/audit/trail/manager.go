// Package trail provides a comprehensive audit trail system for tracking all operations
package trail

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/audit"
)

// AuditTrailManager manages the audit trail system
type AuditTrailManager struct {
	config       *AuditTrailConfig
	loggers      []Logger
	mu           sync.RWMutex
	lastLogHash  string
	auditManager *audit.AuditManager
}

// NewAuditTrailManager creates a new audit trail manager
func NewAuditTrailManager(config *AuditTrailConfig, auditManager *audit.AuditManager) (*AuditTrailManager, error) {
	if config == nil {
		config = DefaultAuditTrailConfig()
	}

	// Create log directory if it doesn't exist
	if config.LogDirectory != "" {
		if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	manager := &AuditTrailManager{
		config:       config,
		loggers:      make([]Logger, 0),
		auditManager: auditManager,
	}

	// Initialize file logger
	if config.LogDirectory != "" {
		fileLogger, err := NewFileLogger(filepath.Join(config.LogDirectory, "audit_trail.log"), config)
		if err != nil {
			return nil, fmt.Errorf("failed to create file logger: %w", err)
		}
		manager.loggers = append(manager.loggers, fileLogger)
	}

	// Always add in-memory logger for immediate querying
	memLogger := NewInMemoryLogger(config)
	manager.loggers = append(manager.loggers, memLogger)

	return manager, nil
}

// AddLogger adds a logger to the manager
func (m *AuditTrailManager) AddLogger(logger Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loggers = append(m.loggers, logger)
}

// RemoveLogger removes a logger from the manager
func (m *AuditTrailManager) RemoveLogger(logger Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, l := range m.loggers {
		if l == logger {
			m.loggers = append(m.loggers[:i], m.loggers[i+1:]...)
			break
		}
	}
}

// LogOperation logs an operation to the audit trail
func (m *AuditTrailManager) LogOperation(ctx context.Context, log *AuditLog) error {
	if !m.config.Enabled {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Redact sensitive data if configured
	if len(m.config.RedactFields) > 0 {
		log = RedactSensitiveData(log, m.config.RedactFields)
	}

	// Add hash chain if enabled
	if m.config.EnableHashChain && m.lastLogHash != "" {
		log.PreviousHash = m.lastLogHash
	}

	// Sign the log if configured
	if m.config.SignLogs && m.config.SigningKey != "" {
		if err := log.Sign(m.config.SigningKey); err != nil {
			return fmt.Errorf("failed to sign audit log: %w", err)
		}
	}

	// Update the last log hash
	if m.config.EnableHashChain {
		logJSON, err := json.Marshal(log)
		if err != nil {
			return fmt.Errorf("failed to marshal log for hash chain: %w", err)
		}
		hash := sha256.Sum256(logJSON)
		m.lastLogHash = base64.StdEncoding.EncodeToString(hash[:])
	}

	// Log to all loggers
	var firstErr error
	for _, logger := range m.loggers {
		if err := logger.Log(ctx, log); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Also log to the audit manager if available
	if m.auditManager != nil {
		auditEvent := ConvertAuditLogToAuditEvent(log)
		if err := m.auditManager.LogAudit(ctx, auditEvent); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// LogCreate logs a create operation
func (m *AuditTrailManager) LogCreate(ctx context.Context, userID, username, resourceType, resourceID, description string, details map[string]interface{}) error {
	log := NewAuditLog("create", resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID)

	if details != nil {
		for k, v := range details {
			log.WithDetail(k, v)
		}
	}

	return m.LogOperation(ctx, log)
}

// LogRead logs a read operation
func (m *AuditTrailManager) LogRead(ctx context.Context, userID, username, resourceType, resourceID, description string, details map[string]interface{}) error {
	log := NewAuditLog("read", resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID)

	if details != nil {
		for k, v := range details {
			log.WithDetail(k, v)
		}
	}

	return m.LogOperation(ctx, log)
}

// LogUpdate logs an update operation
func (m *AuditTrailManager) LogUpdate(ctx context.Context, userID, username, resourceType, resourceID, description string, previousState, newState map[string]interface{}) error {
	log := NewAuditLog("update", resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID)

	if m.config.IncludeState {
		log.WithStates(previousState, newState)
	} else {
		// Calculate changes without storing full state
		changes := calculateChanges(previousState, newState)
		log.WithChanges(changes)
	}

	return m.LogOperation(ctx, log)
}

// LogDelete logs a delete operation
func (m *AuditTrailManager) LogDelete(ctx context.Context, userID, username, resourceType, resourceID, description string, deletedState map[string]interface{}) error {
	log := NewAuditLog("delete", resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID)

	if m.config.IncludeState && deletedState != nil {
		log.WithStates(deletedState, nil)
	}

	return m.LogOperation(ctx, log)
}

// LogVerification logs a verification operation
func (m *AuditTrailManager) LogVerification(ctx context.Context, userID, username, resourceType, resourceID, description, verificationType, result string, metadata map[string]interface{}) error {
	log := NewAuditLog("verify", resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID).
		WithVerification(verificationType, username, result)

	if metadata != nil {
		for k, v := range metadata {
			log.WithVerificationMetadata(k, v)
		}
	}

	return m.LogOperation(ctx, log)
}

// LogCompliance logs a compliance-related operation
func (m *AuditTrailManager) LogCompliance(ctx context.Context, userID, username, resourceType, resourceID, description string, frameworks, controls []string, dataClassification, retentionPeriod string) error {
	log := NewAuditLog("compliance", resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID).
		WithCompliance(frameworks, controls, dataClassification, retentionPeriod)

	return m.LogOperation(ctx, log)
}

// LogApproval logs an approval operation
func (m *AuditTrailManager) LogApproval(ctx context.Context, userID, username, resourceType, resourceID, description, status string, details map[string]interface{}) error {
	log := NewAuditLog("approval", resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID).
		WithApproval(true, status, username)

	if details != nil {
		for k, v := range details {
			log.WithDetail(k, v)
		}
	}

	return m.LogOperation(ctx, log)
}

// LogCustom logs a custom operation
func (m *AuditTrailManager) LogCustom(ctx context.Context, operation, userID, username, resourceType, resourceID, description string, details map[string]interface{}) error {
	log := NewAuditLog(operation, resourceType, description).
		WithUser(userID, username).
		WithResource(resourceID)

	if details != nil {
		for k, v := range details {
			log.WithDetail(k, v)
		}
	}

	return m.LogOperation(ctx, log)
}

// QueryLogs queries audit logs based on filters
func (m *AuditTrailManager) QueryLogs(ctx context.Context, filters map[string]interface{}, startTime, endTime time.Time, limit, offset int) ([]*AuditLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect logs from all loggers
	allLogs := make([]*AuditLog, 0)
	for _, logger := range m.loggers {
		logs, err := logger.Query(ctx, filters, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to query logs from logger: %w", err)
		}
		allLogs = append(allLogs, logs...)
	}

	// Remove duplicates (by ID)
	uniqueLogs := make(map[string]*AuditLog)
	for _, log := range allLogs {
		uniqueLogs[log.ID] = log
	}

	// Convert map to slice
	result := make([]*AuditLog, 0, len(uniqueLogs))
	for _, log := range uniqueLogs {
		result = append(result, log)
	}

	// Sort logs by timestamp (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	// Apply pagination
	if offset >= len(result) {
		return []*AuditLog{}, nil
	}

	end := offset + limit
	if end > len(result) || limit <= 0 {
		end = len(result)
	}

	return result[offset:end], nil
}

// GetLog retrieves a specific log by ID
func (m *AuditTrailManager) GetLog(ctx context.Context, id string) (*AuditLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, logger := range m.loggers {
		log, err := logger.GetLog(ctx, id)
		if err == nil && log != nil {
			return log, nil
		}
	}

	return nil, fmt.Errorf("log with ID %s not found", id)
}

// ExportLogs exports logs to a file in the specified format
func (m *AuditTrailManager) ExportLogs(ctx context.Context, filters map[string]interface{}, startTime, endTime time.Time, format, filePath string) error {
	// Query logs
	logs, err := m.QueryLogs(ctx, filters, startTime, endTime, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to query logs for export: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Export logs in the specified format
	switch strings.ToLower(format) {
	case "json":
		return ExportLogsToJSON(logs, filePath)
	case "csv":
		return ExportLogsToCSV(logs, filePath)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// VerifyLogIntegrity verifies the integrity of logs
func (m *AuditTrailManager) VerifyLogIntegrity(ctx context.Context, startTime, endTime time.Time) (bool, []string, error) {
	if !m.config.SignLogs || m.config.SigningKey == "" {
		return false, []string{"Log signing is not enabled"}, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Query logs for the specified time range
	logs, err := m.QueryLogs(ctx, nil, startTime, endTime, 0, 0)
	if err != nil {
		return false, []string{fmt.Sprintf("Failed to query logs: %v", err)}, err
	}

	// Sort logs by timestamp (oldest first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.Before(logs[j].Timestamp)
	})

	// Verify each log's signature
	issues := make([]string, 0)
	for i, log := range logs {
		// Verify signature
		if log.Signature == "" {
			issues = append(issues, fmt.Sprintf("Log %s has no signature", log.ID))
			continue
		}

		valid, err := log.Verify(m.config.SigningKey)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Failed to verify log %s: %v", log.ID, err))
			continue
		}

		if !valid {
			issues = append(issues, fmt.Sprintf("Log %s has invalid signature", log.ID))
		}

		// Verify hash chain if enabled
		if m.config.EnableHashChain && i > 0 && log.PreviousHash != "" {
			prevLog := logs[i-1]
			prevLogJSON, err := json.Marshal(prevLog)
			if err != nil {
				issues = append(issues, fmt.Sprintf("Failed to marshal log %s for hash verification: %v", prevLog.ID, err))
				continue
			}

			hash := sha256.Sum256(prevLogJSON)
			expectedHash := base64.StdEncoding.EncodeToString(hash[:])

			if log.PreviousHash != expectedHash {
				issues = append(issues, fmt.Sprintf("Log %s has invalid previous hash", log.ID))
			}
		}
	}

	return len(issues) == 0, issues, nil
}

// PurgeLogs deletes logs older than the retention period
func (m *AuditTrailManager) PurgeLogs(ctx context.Context) error {
	if m.config.RetentionDays <= 0 {
		return nil // No retention policy
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	cutoffTime := time.Now().UTC().AddDate(0, 0, -m.config.RetentionDays)

	for _, logger := range m.loggers {
		if purger, ok := logger.(LogPurger); ok {
			if err := purger.PurgeLogs(ctx, cutoffTime); err != nil {
				return fmt.Errorf("failed to purge logs: %w", err)
			}
		}
	}

	return nil
}

// Close closes the audit trail manager and all loggers
func (m *AuditTrailManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for _, logger := range m.loggers {
		if closer, ok := logger.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}

	return firstErr
}

// LogFromAuditEvent converts and logs an AuditEvent
func (m *AuditTrailManager) LogFromAuditEvent(ctx context.Context, event *audit.AuditEvent) error {
	log := ConvertAuditEventToAuditLog(event)
	return m.LogOperation(ctx, log)
}
