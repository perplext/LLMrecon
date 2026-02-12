//go:build windows

package trail

import (
	"context"
	"fmt"
	"sync"
)

// SyslogPriority is a placeholder for syslog.Priority on Windows
type SyslogPriority int

// SyslogLogger implements a stub audit logger on Windows where syslog is unavailable
type SyslogLogger struct {
	id string
	mu sync.Mutex
}

// NewSyslogLogger returns an error on Windows since syslog is not supported
func NewSyslogLogger(facility SyslogPriority, tag string) (*SyslogLogger, error) {
	return nil, fmt.Errorf("syslog is not supported on Windows")
}

// GetID returns the unique identifier for this logger
func (l *SyslogLogger) GetID() string {
	return l.id
}

// Log is a stub that returns an error on Windows
func (l *SyslogLogger) Log(ctx context.Context, log *AuditLog) error {
	return fmt.Errorf("syslog is not supported on Windows")
}

// Close is a no-op on Windows
func (l *SyslogLogger) Close() error {
	return nil
}
