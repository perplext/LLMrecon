package audit

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// AuditLogger provides logging functionality for security auditing
type AuditLogger struct {
	writer io.Writer
	mutex  sync.Mutex

// NewAuditLogger creates a new audit logger
func NewAuditLogger(writer io.Writer) *AuditLogger {
	return &AuditLogger{
		writer: writer,
	}

// Log writes an audit log entry
func (l *AuditLogger) Log(ctx context.Context, level, message string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, message)
	
	_, err := l.writer.Write([]byte(entry))
	return err

// LogEvent logs an audit event
func (l *AuditLogger) LogEvent(event, component, id string, details map[string]interface{}) {
	l.LogEventWithStatus(event, component, id, "info", details)

// LogEventWithStatus logs an audit event with status
func (l *AuditLogger) LogEventWithStatus(event, component, id, status string, details map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s] [%s] [%s] [%s] [%s]", timestamp, status, component, event, id)
	
	if details != nil {
		for k, v := range details {
			entry += fmt.Sprintf(" %s=%v", k, v)
		}
	}
	entry += "\n"
	
	l.writer.Write([]byte(entry))
