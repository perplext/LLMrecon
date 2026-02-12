// Package trail provides a comprehensive audit trail and logging system
package trail

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditLogger defines the interface for basic audit logging
type AuditLogger interface {
	// GetID returns the unique identifier for this logger
	GetID() string

	// Log records an audit log entry
	Log(ctx context.Context, log *AuditLog) error

	// Close releases any resources used by the logger
	Close() error
}

// AuditQueryLogger extends AuditLogger with query capabilities
type AuditQueryLogger interface {
	AuditLogger

	// Query searches for audit logs matching the specified criteria
	Query(ctx context.Context, query *LogQuery) (*LogQueryResult, error)
}

// AuditExporter extends AuditLogger with export capabilities
type AuditExporter interface {
	AuditLogger

	// Export exports audit logs in the specified format
	Export(ctx context.Context, logs []*AuditLog, format ExportFormat) ([]byte, error)
}

// FileLogger implements audit logging to files
type FileLogger struct {
	id           string
	directory    string
	currentFile  *os.File
	maxFileSize  int64 // in bytes
	maxFiles     int
	compress     bool
	mu           sync.Mutex
	rotationTime time.Time
}

// NewFileLogger creates a new file-based audit logger
func NewFileLogger(directory string, maxFileSize int64, maxFiles int, compress bool) (*FileLogger, error) {
	// Ensure directory exists
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &FileLogger{
		id:          "file-logger",
		directory:   directory,
		maxFileSize: maxFileSize,
		maxFiles:    maxFiles,
		compress:    compress,
	}

	// Open initial log file
	if err := logger.openLogFile(); err != nil {
		return nil, err
	}

	return logger, nil
}

// GetID returns the unique identifier for this logger
func (l *FileLogger) GetID() string {
	return l.id
}

// Log records an audit log entry to a file
func (l *FileLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Convert log to JSON
	data, err := log.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert log to JSON: %w", err)
	}

	// Check if we need to rotate the log file
	if l.currentFile != nil {
		info, err := l.currentFile.Stat()
		if err == nil && info.Size() >= l.maxFileSize {
			if err := l.rotateLogFile(); err != nil {
				return fmt.Errorf("failed to rotate log file: %w", err)
			}
		}
	}

	// Write the log entry with a newline
	if _, err := l.currentFile.WriteString(data + "\n"); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

// Close closes the file logger
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentFile != nil {
		err := l.currentFile.Close()
		l.currentFile = nil
		return err
	}

	return nil
}

// Helper methods for FileLogger

// openLogFile opens a new log file
func (l *FileLogger) openLogFile() error {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("audit-%s.log", timestamp)
	filePath := filepath.Join(l.directory, filename)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.currentFile = file
	l.rotationTime = time.Now()

	return nil
}

// rotateLogFile rotates the current log file
func (l *FileLogger) rotateLogFile() error {
	// Close the current file
	if l.currentFile != nil {
		if err := l.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		l.currentFile = nil
	}

	// Open a new log file
	if err := l.openLogFile(); err != nil {
		return err
	}

	return nil
}

// InMemoryLogger implements in-memory audit logging
type InMemoryLogger struct {
	id      string
	logs    []*AuditLog
	maxLogs int
	mu      sync.RWMutex
}

// NewInMemoryLogger creates a new in-memory audit logger
func NewInMemoryLogger(maxLogs int) *InMemoryLogger {
	if maxLogs <= 0 {
		maxLogs = 1000 // Default to 1000 logs
	}

	return &InMemoryLogger{
		id:      "memory-logger",
		logs:    make([]*AuditLog, 0, maxLogs),
		maxLogs: maxLogs,
	}
}

// GetID returns the unique identifier for this logger
func (l *InMemoryLogger) GetID() string {
	return l.id
}

// Log records an audit log entry in memory
func (l *InMemoryLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a copy of the log to prevent modification
	logCopy := *log

	// Add to logs
	l.logs = append(l.logs, &logCopy)

	// Trim if we have too many logs
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[len(l.logs)-l.maxLogs:]
	}

	return nil
}

// Close closes the in-memory logger
func (l *InMemoryLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logs = nil

	return nil
}
