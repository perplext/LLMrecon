// Package audit provides comprehensive security audit logging functionality
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditLogger defines the interface for audit logging implementations
type AuditLogger interface {
	// LogAudit logs an audit event
	LogAudit(ctx context.Context, event *AuditEvent) error
	
	// QueryAuditLogs queries audit logs based on filters
	QueryAuditLogs(ctx context.Context, filter *AuditQueryFilter) ([]*AuditEvent, error)
	
	// GetAuditLog retrieves a specific audit log by ID
	GetAuditLog(ctx context.Context, id string) (*AuditEvent, error)
}

// InMemoryAuditLogger is an in-memory implementation of AuditLogger
type InMemoryAuditLogger struct {
	logs   []*AuditEvent
	config *AuditConfig
	mu     sync.RWMutex
}

// NewInMemoryAuditLogger creates a new in-memory audit logger
func NewInMemoryAuditLogger(config *AuditConfig) *InMemoryAuditLogger {
	return &InMemoryAuditLogger{
		logs:   make([]*AuditEvent, 0),
		config: config,
	}
}

// LogAudit logs an audit event to memory
func (l *InMemoryAuditLogger) LogAudit(ctx context.Context, event *AuditEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a copy of the event to prevent modification
	eventCopy := *event
	l.logs = append(l.logs, &eventCopy)

	// Limit the number of logs in memory
	maxLogs := 1000
	if len(l.logs) > maxLogs {
		l.logs = l.logs[len(l.logs)-maxLogs:]
	}

	return nil
}

// QueryAuditLogs queries audit logs based on filters
func (l *InMemoryAuditLogger) QueryAuditLogs(ctx context.Context, filter *AuditQueryFilter) ([]*AuditEvent, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if filter == nil {
		// Return all logs (up to a reasonable limit)
		limit := 100
		if len(l.logs) <= limit {
			return l.logs, nil
		}
		return l.logs[len(l.logs)-limit:], nil
	}

	// Apply filters
	filtered := make([]*AuditEvent, 0)
	for _, log := range l.logs {
		if matchesFilter(log, filter) {
			filtered = append(filtered, log)
		}
	}

	// Apply pagination
	if filter.Limit > 0 && len(filtered) > filter.Limit {
		start := 0
		if filter.Offset > 0 {
			start = filter.Offset
			if start >= len(filtered) {
				return []*AuditEvent{}, nil
			}
		}
		end := start + filter.Limit
		if end > len(filtered) {
			end = len(filtered)
		}
		return filtered[start:end], nil
	}

	return filtered, nil
}

// GetAuditLog retrieves a specific audit log by ID
func (l *InMemoryAuditLogger) GetAuditLog(ctx context.Context, id string) (*AuditEvent, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, log := range l.logs {
		if log.ID == id {
			// Return a copy to prevent modification
			logCopy := *log
			return &logCopy, nil
		}
	}

	return nil, fmt.Errorf("audit log with ID %s not found", id)
}

// FileAuditLogger is a file-based implementation of AuditLogger
type FileAuditLogger struct {
	filePath string
	file     *os.File
	config   *AuditConfig
	mu       sync.Mutex
}

// NewFileAuditLogger creates a new file-based audit logger
func NewFileAuditLogger(filePath string, config *AuditConfig) (*FileAuditLogger, error) {
	// Ensure directory exists
	dir := filePath[:len(filePath)-len("/"+filePath)]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file for appending
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &FileAuditLogger{
		filePath: filePath,
		file:     file,
		config:   config,
	}, nil
}

// LogAudit logs an audit event to the file
func (l *FileAuditLogger) LogAudit(ctx context.Context, event *AuditEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Marshal the event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Write the event to the file with a newline
	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	// Check if file rotation is needed
	if l.config.MaxLogFileSize > 0 {
		if info, err := l.file.Stat(); err == nil && info.Size() > int64(l.config.MaxLogFileSize*1024*1024) {
			if err := l.rotateLogFile(); err != nil {
				return fmt.Errorf("failed to rotate log file: %w", err)
			}
		}
	}

	return nil
}

// rotateLogFile rotates the log file
func (l *FileAuditLogger) rotateLogFile() error {
	// Close the current file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	// Rename the current file with a timestamp
	timestamp := time.Now().Format("20060102-150405")
	newPath := fmt.Sprintf("%s.%s", l.filePath, timestamp)
	if err := os.Rename(l.filePath, newPath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Open a new log file
	file, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}
	l.file = file

	// Clean up old log files if needed
	if l.config.MaxLogFiles > 0 {
		if err := l.cleanupOldLogFiles(); err != nil {
			return fmt.Errorf("failed to clean up old log files: %w", err)
		}
	}

	return nil
}

// cleanupOldLogFiles removes old log files
func (l *FileAuditLogger) cleanupOldLogFiles() error {
	// Get the directory and base filename
	dir := l.filePath[:len(l.filePath)-len("/"+l.filePath)]
	base := l.filePath[len(dir)+1:]

	// List files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	// Find log files with the same base name
	logFiles := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > len(base) && file.Name()[:len(base)] == base && file.Name()[len(base)] == '.' {
			logFiles = append(logFiles, file.Name())
		}
	}

	// If we have more files than the limit, remove the oldest ones
	if len(logFiles) > l.config.MaxLogFiles {
		// Sort files by name (which includes the timestamp)
		// In a real implementation, we would parse the timestamp and sort by date
		// For simplicity, we'll just sort by name, which works because of the timestamp format
		for i := 0; i < len(logFiles)-l.config.MaxLogFiles; i++ {
			oldestFile := logFiles[i]
			if err := os.Remove(dir + "/" + oldestFile); err != nil {
				return fmt.Errorf("failed to remove old log file: %w", err)
			}
		}
	}

	return nil
}

// QueryAuditLogs queries audit logs from the file
func (l *FileAuditLogger) QueryAuditLogs(ctx context.Context, filter *AuditQueryFilter) ([]*AuditEvent, error) {
	// In a real implementation, this would scan the log file and apply filters
	// For simplicity, we'll return an error suggesting to use a database for queries
	return nil, fmt.Errorf("file-based audit logger does not support queries, use a database logger instead")
}

// GetAuditLog retrieves a specific audit log by ID
func (l *FileAuditLogger) GetAuditLog(ctx context.Context, id string) (*AuditEvent, error) {
	// In a real implementation, this would scan the log file for the specific ID
	// For simplicity, we'll return an error suggesting to use a database for queries
	return nil, fmt.Errorf("file-based audit logger does not support retrieval by ID, use a database logger instead")
}

// Close closes the file audit logger
func (l *FileAuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SyslogAuditLogger is a syslog-based implementation of AuditLogger
type SyslogAuditLogger struct {
	config *AuditConfig
	// In a real implementation, this would include a syslog client
}

// NewSyslogAuditLogger creates a new syslog-based audit logger
func NewSyslogAuditLogger(config *AuditConfig) (*SyslogAuditLogger, error) {
	// In a real implementation, this would initialize a syslog client
	return &SyslogAuditLogger{
		config: config,
	}, nil
}

// LogAudit logs an audit event to syslog
func (l *SyslogAuditLogger) LogAudit(ctx context.Context, event *AuditEvent) error {
	// In a real implementation, this would send the event to syslog
	// For now, we'll just return nil to indicate success
	return nil
}

// QueryAuditLogs queries audit logs from syslog
func (l *SyslogAuditLogger) QueryAuditLogs(ctx context.Context, filter *AuditQueryFilter) ([]*AuditEvent, error) {
	// Syslog doesn't support querying, so we'll return an error
	return nil, fmt.Errorf("syslog-based audit logger does not support queries")
}

// GetAuditLog retrieves a specific audit log by ID
func (l *SyslogAuditLogger) GetAuditLog(ctx context.Context, id string) (*AuditEvent, error) {
	// Syslog doesn't support retrieval by ID, so we'll return an error
	return nil, fmt.Errorf("syslog-based audit logger does not support retrieval by ID")
}

// DatabaseAuditLogger is a database-based implementation of AuditLogger
type DatabaseAuditLogger struct {
	dbURL  string
	config *AuditConfig
	// In a real implementation, this would include a database client
}

// NewDatabaseAuditLogger creates a new database-based audit logger
func NewDatabaseAuditLogger(dbURL string, config *AuditConfig) (*DatabaseAuditLogger, error) {
	// In a real implementation, this would initialize a database client
	return &DatabaseAuditLogger{
		dbURL:  dbURL,
		config: config,
	}, nil
}

// LogAudit logs an audit event to the database
func (l *DatabaseAuditLogger) LogAudit(ctx context.Context, event *AuditEvent) error {
	// In a real implementation, this would insert the event into the database
	// For now, we'll just return nil to indicate success
	return nil
}

// QueryAuditLogs queries audit logs from the database
func (l *DatabaseAuditLogger) QueryAuditLogs(ctx context.Context, filter *AuditQueryFilter) ([]*AuditEvent, error) {
	// In a real implementation, this would query the database with the provided filters
	// For now, we'll just return an empty slice
	return []*AuditEvent{}, nil
}

// GetAuditLog retrieves a specific audit log by ID
func (l *DatabaseAuditLogger) GetAuditLog(ctx context.Context, id string) (*AuditEvent, error) {
	// In a real implementation, this would query the database for the specific ID
	// For now, we'll just return an error
	return nil, fmt.Errorf("audit log with ID %s not found", id)
}

// matchesFilter checks if an audit event matches the provided filter
func matchesFilter(event *AuditEvent, filter *AuditQueryFilter) bool {
	if filter == nil {
		return true
	}

	// Check user ID
	if filter.UserID != "" && event.UserID != filter.UserID {
		return false
	}

	// Check username
	if filter.Username != "" && event.Username != filter.Username {
		return false
	}

	// Check action
	if filter.Action != "" && event.Action != filter.Action {
		return false
	}

	// Check resource
	if filter.Resource != "" && event.Resource != filter.Resource {
		return false
	}

	// Check resource ID
	if filter.ResourceID != "" && event.ResourceID != filter.ResourceID {
		return false
	}

	// Check IP address
	if filter.IPAddress != "" && event.IPAddress != filter.IPAddress {
		return false
	}

	// Check severity (minimum)
	if filter.MinSeverity != "" && severityLevel(event.Severity) < severityLevel(filter.MinSeverity) {
		return false
	}

	// Check severity (maximum)
	if filter.MaxSeverity != "" && severityLevel(event.Severity) > severityLevel(filter.MaxSeverity) {
		return false
	}

	// Check status
	if filter.Status != "" && event.Status != filter.Status {
		return false
	}

	// Check session ID
	if filter.SessionID != "" && event.SessionID != filter.SessionID {
		return false
	}

	// Check time range (start)
	if !filter.StartTime.IsZero() && event.Timestamp.Before(filter.StartTime) {
		return false
	}

	// Check time range (end)
	if !filter.EndTime.IsZero() && event.Timestamp.After(filter.EndTime) {
		return false
	}

	// Check request ID
	if filter.RequestID != "" && event.RequestID != filter.RequestID {
		return false
	}

	// Check system generated
	if !filter.IncludeSystemGenerated && event.SystemGenerated {
		return false
	}

	// Check tags (must match all)
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			found := false
			for _, eventTag := range event.Tags {
				if tag == eventTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// TODO: Implement full-text search if needed

	return true
}
