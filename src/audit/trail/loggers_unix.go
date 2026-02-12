//go:build !windows

package trail

import (
	"context"
	"fmt"
	"log/syslog"
	"sync"
)

// SyslogLogger implements audit logging to syslog
type SyslogLogger struct {
	id     string
	writer *syslog.Writer
	mu     sync.Mutex
}

// NewSyslogLogger creates a new syslog-based audit logger
func NewSyslogLogger(facility syslog.Priority, tag string) (*SyslogLogger, error) {
	writer, err := syslog.New(facility, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}

	return &SyslogLogger{
		id:     "syslog-logger",
		writer: writer,
	}, nil
}

// GetID returns the unique identifier for this logger
func (l *SyslogLogger) GetID() string {
	return l.id
}

// Log records an audit log entry to syslog
func (l *SyslogLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Convert log to JSON
	data, err := log.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert log to JSON: %w", err)
	}

	// Write to syslog with info priority for now
	if err := l.writer.Info(data); err != nil {
		return fmt.Errorf("failed to write to syslog: %w", err)
	}

	return nil
}

// Close closes the syslog logger
func (l *SyslogLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer != nil {
		return l.writer.Close()
	}

	return nil
}
