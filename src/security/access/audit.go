// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// AuditAction represents the type of action being audited
type AuditAction string

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

// Audit actions
const (
	AuditActionLogin        AuditAction = "login"
	AuditActionLogout       AuditAction = "logout"
	AuditActionCreate       AuditAction = "create"
	AuditActionRead         AuditAction = "read"
	AuditActionUpdate       AuditAction = "update"
	AuditActionDelete       AuditAction = "delete"
	AuditActionExecute      AuditAction = "execute"
	AuditActionAuthorize    AuditAction = "authorize"
	AuditActionUnauthorized AuditAction = "unauthorized"
	AuditActionSystem       AuditAction = "system"
	AuditActionSecurity     AuditAction = "security"
)

// Audit severity levels
const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityLow      AuditSeverity = "low"
	AuditSeverityMedium   AuditSeverity = "medium"
	AuditSeverityHigh     AuditSeverity = "high"
	AuditSeverityCritical AuditSeverity = "critical"
	AuditSeverityError    AuditSeverity = "error"
)

// AuditLog represents a security audit log entry
type AuditLog struct {
	ID          string                 `json:"id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Action      AuditAction            `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Description string                 `json:"description"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Severity    AuditSeverity          `json:"severity"`
	Status      string                 `json:"status"`
	SessionID   string                 `json:"session_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
}

// AuditLogger defines the interface for audit logging
// AuditLogger interface is now defined in store_interfaces.go

// AuditLogFilter defines filters for querying audit logs
type AuditLogFilter struct {
	UserID     string        `json:"user_id,omitempty"`
	Username   string        `json:"username,omitempty"`
	Action     AuditAction   `json:"action,omitempty"`
	Resource   string        `json:"resource,omitempty"`
	ResourceID string        `json:"resource_id,omitempty"`
	IPAddress  string        `json:"ip_address,omitempty"`
	Severity   AuditSeverity `json:"severity,omitempty"`
	Status     string        `json:"status,omitempty"`
	SessionID  string        `json:"session_id,omitempty"`
	StartTime  time.Time     `json:"start_time,omitempty"`
	EndTime    time.Time     `json:"end_time,omitempty"`
	Limit      int           `json:"limit,omitempty"`
	Offset     int           `json:"offset,omitempty"`
}

// InMemoryAuditLogger is a simple in-memory implementation of AuditLogger
type InMemoryAuditLogger struct {
	logs []*AuditLog
	mu   sync.RWMutex
}

// NewInMemoryAuditLogger creates a new in-memory audit logger
func NewInMemoryAuditLogger() *InMemoryAuditLogger {
	return &InMemoryAuditLogger{
		logs: make([]*AuditLog, 0),
	}
}

// LogAudit logs an audit event
func (l *InMemoryAuditLogger) LogAudit(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Generate ID if not provided
	if log.ID == "" {
		log.ID = generateRandomID()
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// Add log
	l.logs = append(l.logs, log)

	return nil
}

// GetAuditLogs retrieves audit logs
func (l *InMemoryAuditLogger) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*AuditLog, int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	// Default offset if not specified
	if offset < 0 {
		offset = 0
	}

	// Filter logs
	filtered := make([]*AuditLog, 0)
	for _, log := range l.logs {
		if filter != nil {
			// Apply filters
			if userID, ok := filter["user_id"].(string); ok && userID != "" && log.UserID != userID {
				continue
			}
			if username, ok := filter["username"].(string); ok && username != "" && log.Username != username {
				continue
			}
			if action, ok := filter["action"].(string); ok && action != "" && string(log.Action) != action {
				continue
			}
			if resource, ok := filter["resource"].(string); ok && resource != "" && log.Resource != resource {
				continue
			}
			if resourceID, ok := filter["resource_id"].(string); ok && resourceID != "" && log.ResourceID != resourceID {
				continue
			}
			if ipAddress, ok := filter["ip_address"].(string); ok && ipAddress != "" && log.IPAddress != ipAddress {
				continue
			}
			if severity, ok := filter["severity"].(string); ok && severity != "" && string(log.Severity) != severity {
				continue
			}
			if status, ok := filter["status"].(string); ok && status != "" && log.Status != status {
				continue
			}
			if sessionID, ok := filter["session_id"].(string); ok && sessionID != "" && log.SessionID != sessionID {
				continue
			}
			if startTime, ok := filter["start_time"].(time.Time); ok && !startTime.IsZero() && log.Timestamp.Before(startTime) {
				continue
			}
			if endTime, ok := filter["end_time"].(time.Time); ok && !endTime.IsZero() && log.Timestamp.After(endTime) {
				continue
			}
		}

		filtered = append(filtered, log)
	}

	totalCount := len(filtered)

	// Apply pagination
	start := offset
	if start >= len(filtered) {
		return []*AuditLog{}, totalCount, nil
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], totalCount, nil
}

// GetAuditLogByID retrieves an audit log by ID
func (l *InMemoryAuditLogger) GetAuditLogByID(ctx context.Context, id string) (*AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, log := range l.logs {
		if log.ID == id {
			return log, nil
		}
	}

	return nil, fmt.Errorf("audit log not found: %s", id)
}

// Initialize initializes the audit logger
func (l *InMemoryAuditLogger) Initialize(ctx context.Context) error {
	// Nothing to initialize for in-memory logger
	return nil
}

// Close closes the audit logger
func (l *InMemoryAuditLogger) Close(ctx context.Context) error {
	// Nothing to close for in-memory logger
	return nil
}

// FileAuditLogger is a file-based implementation of AuditLogger
type FileAuditLogger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// NewFileAuditLogger creates a new file-based audit logger
func NewFileAuditLogger(filePath string) *FileAuditLogger {
	return &FileAuditLogger{
		filePath: filePath,
	}
}

// Initialize initializes the file audit logger
func (l *FileAuditLogger) Initialize(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Open file for appending
	file, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}

	l.file = file
	return nil
}

// Close closes the file audit logger
func (l *FileAuditLogger) Close(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return fmt.Errorf("failed to close audit log file: %w", err)
		}
		l.file = nil
	}

	return nil
}

// LogAudit logs an audit event to the file
func (l *FileAuditLogger) LogAudit(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Generate ID if not provided
	if log.ID == "" {
		log.ID = generateRandomID()
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// Check if file is open
	if l.file == nil {
		return fmt.Errorf("audit log file is not open")
	}

	// Convert log to JSON
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	// Write log to file
	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs
func (l *FileAuditLogger) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*AuditLog, int, error) {
	// This is a simplified implementation that reads the entire file
	// In a real-world scenario, you would use a database or more efficient storage

	// Open file for reading
	file, err := os.Open(l.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*AuditLog{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer file.Close()

	// Read logs
	logs := make([]*AuditLog, 0)
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var log AuditLog
		if err := decoder.Decode(&log); err != nil {
			return nil, 0, fmt.Errorf("failed to decode audit log: %w", err)
		}

		// Apply filters
		if filter != nil {
			if userID, ok := filter["user_id"].(string); ok && userID != "" && log.UserID != userID {
				continue
			}
			if username, ok := filter["username"].(string); ok && username != "" && log.Username != username {
				continue
			}
			if action, ok := filter["action"].(string); ok && action != "" && string(log.Action) != action {
				continue
			}
			if resource, ok := filter["resource"].(string); ok && resource != "" && log.Resource != resource {
				continue
			}
			if resourceID, ok := filter["resource_id"].(string); ok && resourceID != "" && log.ResourceID != resourceID {
				continue
			}
			if ipAddress, ok := filter["ip_address"].(string); ok && ipAddress != "" && log.IPAddress != ipAddress {
				continue
			}
			if severity, ok := filter["severity"].(string); ok && severity != "" && string(log.Severity) != severity {
				continue
			}
			if status, ok := filter["status"].(string); ok && status != "" && log.Status != status {
				continue
			}
			if sessionID, ok := filter["session_id"].(string); ok && sessionID != "" && log.SessionID != sessionID {
				continue
			}
			if startTime, ok := filter["start_time"].(time.Time); ok && !startTime.IsZero() && log.Timestamp.Before(startTime) {
				continue
			}
			if endTime, ok := filter["end_time"].(time.Time); ok && !endTime.IsZero() && log.Timestamp.After(endTime) {
				continue
			}
		}

		logs = append(logs, &log)
	}

	totalCount := len(logs)

	// Apply pagination
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	start := offset
	if start >= len(logs) {
		return []*AuditLog{}, totalCount, nil
	}

	end := start + limit
	if end > len(logs) {
		end = len(logs)
	}

	return logs[start:end], totalCount, nil
}

// GetAuditLogByID retrieves an audit log by ID
func (l *FileAuditLogger) GetAuditLogByID(ctx context.Context, id string) (*AuditLog, error) {
	// Open file for reading
	file, err := os.Open(l.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("audit log not found: %s", id)
		}
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer file.Close()

	// Read logs and find the one with matching ID
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var log AuditLog
		if err := decoder.Decode(&log); err != nil {
			return nil, fmt.Errorf("failed to decode audit log: %w", err)
		}

		if log.ID == id {
			return &log, nil
		}
	}

	return nil, fmt.Errorf("audit log not found: %s", id)
}

// MultiAuditLogger is an implementation of AuditLogger that logs to multiple loggers
type MultiAuditLogger struct {
	loggers []AuditLogger
}

// NewMultiAuditLogger creates a new multi-logger
func NewMultiAuditLogger(loggers ...AuditLogger) *MultiAuditLogger {
	return &MultiAuditLogger{
		loggers: loggers,
	}
}

// LogAudit logs an audit event to all loggers
func (l *MultiAuditLogger) LogAudit(ctx context.Context, log *AuditLog) error {
	var lastErr error
	for _, logger := range l.loggers {
		if err := logger.LogAudit(ctx, log); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// GetAuditLogs retrieves audit logs from the primary logger
func (l *MultiAuditLogger) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*AuditLog, int, error) {
	if len(l.loggers) == 0 {
		return []*AuditLog{}, 0, nil
	}
	return l.loggers[0].GetAuditLogs(ctx, filter, offset, limit)
}

// GetAuditLogByID retrieves an audit log by ID from the primary logger
func (l *MultiAuditLogger) GetAuditLogByID(ctx context.Context, id string) (*AuditLog, error) {
	if len(l.loggers) == 0 {
		return nil, fmt.Errorf("no audit loggers configured")
	}
	return l.loggers[0].GetAuditLogByID(ctx, id)
}

// Initialize initializes all loggers
func (l *MultiAuditLogger) Initialize(ctx context.Context) error {
	var lastErr error
	for _, logger := range l.loggers {
		if err := logger.Initialize(ctx); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Close closes all loggers
func (l *MultiAuditLogger) Close(ctx context.Context) error {
	var lastErr error
	for _, logger := range l.loggers {
		if err := logger.Close(ctx); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// AuditManager manages audit logging
type AuditManager struct {
	logger AuditLogger
	config *AuditConfig
}

// NewAuditManager creates a new audit manager
func NewAuditManager(logger AuditLogger, config *AuditConfig) (*AuditManager, error) {
	if logger == nil {
		return nil, fmt.Errorf("audit logger is required")
	}

	// Initialize logger
	if err := logger.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize audit logger: %w", err)
	}

	return &AuditManager{
		logger: logger,
		config: config,
	}, nil
}

// LogAudit logs an audit event
func (m *AuditManager) LogAudit(ctx context.Context, log *AuditLog) error {
	// Check if logging is enabled for this severity
	if m.config != nil && !m.config.IsEnabled(log.Severity) {
		return nil
	}

	return m.logger.LogAudit(ctx, log)
}

// GetAuditLogs retrieves audit logs
func (m *AuditManager) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*AuditLog, int, error) {
	return m.logger.GetAuditLogs(ctx, filter, offset, limit)
}

// GetAuditLogByID retrieves an audit log by ID
func (m *AuditManager) GetAuditLogByID(ctx context.Context, id string) (*AuditLog, error) {
	return m.logger.GetAuditLogByID(ctx, id)
}

// Close closes the audit manager
func (m *AuditManager) Close() error {
	return m.logger.Close(context.Background())
}

// AuditConfig defines configuration for audit logging
type AuditConfig struct {
	EnabledSeverities []AuditSeverity `json:"enabled_severities"`
	LogFilePath       string          `json:"log_file_path"`
	RetentionDays     int             `json:"retention_days"`
}

// IsEnabled checks if logging is enabled for a severity level
func (c *AuditConfig) IsEnabled(severity AuditSeverity) bool {
	if c == nil || len(c.EnabledSeverities) == 0 {
		// If no config or no enabled severities, log everything
		return true
	}

	for _, s := range c.EnabledSeverities {
		if s == severity {
			return true
		}
	}

	return false
}

// DefaultAuditConfig returns the default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		EnabledSeverities: []AuditSeverity{
			AuditSeverityInfo,
			AuditSeverityLow,
			AuditSeverityMedium,
			AuditSeverityHigh,
			AuditSeverityCritical,
			AuditSeverityError,
		},
		LogFilePath:   "audit.log",
		RetentionDays: 90,
	}
}
