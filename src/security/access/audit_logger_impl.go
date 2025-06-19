// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/models"
)

// AuditLoggerImpl implements the AuditLogger interface
type AuditLoggerImpl struct {
	mu          sync.RWMutex
	auditStore  AuditStore
	config      *AuditLoggerConfig
	initialized bool
}

// AuditLoggerConfig contains configuration for the audit logger implementation
type AuditLoggerConfig struct {
	// Logging configuration
	LogToConsole bool
	LogToFile    bool
	LogFile      string
	LogFormat    string
	
	// Retention configuration
	RetentionDays int
}

// AuditStore defines the interface for storing and retrieving audit logs
type AuditStore interface {
	// StoreAuditLog stores an audit log
	StoreAuditLog(ctx context.Context, log *models.AuditLog) error
	
	// GetAuditLog retrieves an audit log by ID
	GetAuditLog(ctx context.Context, id string) (*models.AuditLog, error)
	
	// GetAuditLogs retrieves audit logs with optional filtering
	GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error)
	
	// DeleteAuditLog deletes an audit log
	DeleteAuditLog(ctx context.Context, id string) error
	
	// DeleteAuditLogsBefore deletes audit logs before a specific time
	DeleteAuditLogsBefore(ctx context.Context, before time.Time) (int, error)
	
	// Close closes the audit store
	Close() error
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(auditStore AuditStore, config *AuditLoggerConfig) *AuditLoggerImpl {
	// Set default configuration if not provided
	if config == nil {
		config = &AuditLoggerConfig{
			LogToConsole:  true,
			LogToFile:     false,
			LogFormat:     "json",
			RetentionDays: 90,
		}
	}
	
	return &AuditLoggerImpl{
		auditStore: auditStore,
		config:     config,
	}
}

// Initialize initializes the audit logger
func (l *AuditLoggerImpl) Initialize(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Create the log file if needed
	if l.config.LogToFile && l.config.LogFile != "" {
		// Create the directory if it doesn't exist
		dir := l.config.LogFile[:len(l.config.LogFile)-len("/audit.log")]
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}
		}
		
		// Create or open the log file
		file, err := os.OpenFile(l.config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		file.Close()
	}
	
	l.initialized = true
	return nil
}

// LogAudit logs an audit event
func (l *AuditLoggerImpl) LogAudit(ctx context.Context, log *models.AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Store the audit log
	err := l.auditStore.StoreAuditLog(ctx, log)
	if err != nil {
		return err
	}
	
	// Log to console if enabled
	if l.config.LogToConsole {
		l.logToConsole(log)
	}
	
	// Log to file if enabled
	if l.config.LogToFile && l.config.LogFile != "" {
		l.logToFile(log)
	}
	
	return nil
}

// GetAuditLogs gets audit logs
func (l *AuditLoggerImpl) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.auditStore.GetAuditLogs(ctx, filter, offset, limit)
}

// CleanupOldLogs deletes audit logs older than the retention period
func (l *AuditLoggerImpl) CleanupOldLogs(ctx context.Context) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Calculate the cutoff time
	cutoff := time.Now().AddDate(0, 0, -l.config.RetentionDays)
	
	// Delete old logs
	return l.auditStore.DeleteAuditLogsBefore(ctx, cutoff)
}

// Close closes the audit logger
func (l *AuditLoggerImpl) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.auditStore != nil {
		if err := l.auditStore.Close(); err != nil {
			return err
		}
	}
	
	l.initialized = false
	return nil
}

// logToConsole logs an audit event to the console
func (l *AuditLoggerImpl) logToConsole(log *models.AuditLog) {
	var logStr string
	
	if l.config.LogFormat == "json" {
		logStr = fmt.Sprintf(`{"timestamp":"%s","user_id":"%s","action":"%s","resource":"%s","resource_id":"%s","description":"%s"}`,
			log.Timestamp.Format(time.RFC3339),
			log.UserID,
			log.Action,
			log.Resource,
			log.ResourceID,
			log.Description)
	} else {
		logStr = fmt.Sprintf("[%s] %s %s %s:%s - %s",
			log.Timestamp.Format(time.RFC3339),
			log.UserID,
			log.Action,
			log.Resource,
			log.ResourceID,
			log.Description)
	}
	
	fmt.Println(logStr)
}

// logToFile logs an audit event to a file
func (l *AuditLoggerImpl) logToFile(log *models.AuditLog) error {
	var logStr string
	
	if l.config.LogFormat == "json" {
		logStr = fmt.Sprintf(`{"timestamp":"%s","user_id":"%s","action":"%s","resource":"%s","resource_id":"%s","description":"%s"}`,
			log.Timestamp.Format(time.RFC3339),
			log.UserID,
			log.Action,
			log.Resource,
			log.ResourceID,
			log.Description)
	} else {
		logStr = fmt.Sprintf("[%s] %s %s %s:%s - %s",
			log.Timestamp.Format(time.RFC3339),
			log.UserID,
			log.Action,
			log.Resource,
			log.ResourceID,
			log.Description)
	}
	
	// Append to the log file
	file, err := os.OpenFile(l.config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	if _, err := file.WriteString(logStr + "\n"); err != nil {
		return err
	}
	
	return nil
}

// InMemoryAuditStore is an in-memory implementation of the AuditStore interface
type InMemoryAuditStore struct {
	mu    sync.RWMutex
	logs  map[string]*models.AuditLog
	count int
}

// NewInMemoryAuditStore creates a new in-memory audit store
func NewInMemoryAuditStore() *InMemoryAuditStore {
	return &InMemoryAuditStore{
		logs: make(map[string]*models.AuditLog),
	}
}

// StoreAuditLog stores an audit log
func (s *InMemoryAuditStore) StoreAuditLog(ctx context.Context, log *models.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.logs[log.ID] = log
	s.count++
	
	return nil
}

// GetAuditLog retrieves an audit log by ID
func (s *InMemoryAuditStore) GetAuditLog(ctx context.Context, id string) (*models.AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	log, ok := s.logs[id]
	if !ok {
		return nil, errors.New("audit log not found")
	}
	
	return log, nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *InMemoryAuditStore) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var logs []*models.AuditLog
	
	// Apply filters
	for _, log := range s.logs {
		match := true
		
		// Apply each filter
		for key, value := range filter {
			switch key {
			case "user_id":
				if log.UserID != value.(string) {
					match = false
				}
			case "action":
				if string(log.Action) != value.(string) {
					match = false
				}
			case "resource":
				if log.Resource != value.(string) {
					match = false
				}
			case "resource_id":
				if log.ResourceID != value.(string) {
					match = false
				}
			case "timestamp_from":
				if log.Timestamp.Before(value.(time.Time)) {
					match = false
				}
			case "timestamp_to":
				if log.Timestamp.After(value.(time.Time)) {
					match = false
				}
			}
			
			if !match {
				break
			}
		}
		
		if match {
			logs = append(logs, log)
		}
	}
	
	// Sort logs by timestamp (newest first)
	// In a real implementation, we would sort the logs here
	
	// Apply pagination
	total := len(logs)
	if offset >= total {
		return []*models.AuditLog{}, total, nil
	}
	
	end := offset + limit
	if end > total {
		end = total
	}
	
	return logs[offset:end], total, nil
}

// DeleteAuditLog deletes an audit log
func (s *InMemoryAuditStore) DeleteAuditLog(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, ok := s.logs[id]; !ok {
		return errors.New("audit log not found")
	}
	
	delete(s.logs, id)
	s.count--
	
	return nil
}

// DeleteAuditLogsBefore deletes audit logs before a specific time
func (s *InMemoryAuditStore) DeleteAuditLogsBefore(ctx context.Context, before time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var deleted int
	
	for id, log := range s.logs {
		if log.Timestamp.Before(before) {
			delete(s.logs, id)
			deleted++
		}
	}
	
	s.count -= deleted
	
	return deleted, nil
}

// Close closes the audit store
func (s *InMemoryAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.logs = make(map[string]*models.AuditLog)
	s.count = 0
	
	return nil
}
