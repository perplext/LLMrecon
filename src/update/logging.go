// Package update provides functionality for checking and applying updates
package update

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	// LogLevelDebug is for debug messages
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo is for informational messages
	LogLevelInfo LogLevel = "info"
	// LogLevelWarning is for warning messages
	LogLevelWarning LogLevel = "warning"
	// LogLevelError is for error messages
	LogLevelError LogLevel = "error"
)

// LogEntry represents a single log entry
type LogEntry struct {
	// Timestamp is the time the log entry was created
	Timestamp time.Time `json:"timestamp"`
	// Level is the severity level of the log entry
	Level LogLevel `json:"level"`
	// Message is the log message
	Message string `json:"message"`
	// Component is the component that generated the log entry
	Component string `json:"component"`
	// TransactionID is the ID of the transaction associated with the log entry
	TransactionID string `json:"transaction_id,omitempty"`
	// Details contains additional details about the log entry
	Details map[string]interface{} `json:"details,omitempty"`
}

// UpdateLogger handles logging for update operations
type UpdateLogger struct {
	// Writer is the writer for log output
	Writer io.Writer
	// JSONWriter is the writer for JSON log output
	JSONWriter io.Writer
	// MinLevel is the minimum log level to output
	MinLevel LogLevel
	// IncludeDetails determines whether to include details in log output
	IncludeDetails bool
}

// LoggerOptions contains options for the UpdateLogger
type LoggerOptions struct {
	// Writer is the writer for log output
	Writer io.Writer
	// JSONWriter is the writer for JSON log output
	JSONWriter io.Writer
	// MinLevel is the minimum log level to output
	MinLevel LogLevel
	// IncludeDetails determines whether to include details in log output
	IncludeDetails bool
}

// NewUpdateLogger creates a new update logger
func NewUpdateLogger(options *LoggerOptions) *UpdateLogger {
	// Set default writer if not provided
	writer := options.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// Set default minimum level if not provided
	minLevel := options.MinLevel
	if minLevel == "" {
		minLevel = LogLevelInfo
	}

	return &UpdateLogger{
		Writer:         writer,
		JSONWriter:     options.JSONWriter,
		MinLevel:       minLevel,
		IncludeDetails: options.IncludeDetails,
	}
}

// CreateJSONLogFile creates a JSON log file for the update
func CreateJSONLogFile(logDir, packageID string) (*os.File, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file
	logPath := filepath.Join(logDir, fmt.Sprintf("update-%s-%d.json", packageID, time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Write opening bracket for JSON array
	if _, err := logFile.WriteString("[\n"); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to write to log file: %w", err)
	}

	return logFile, nil
}

// CloseJSONLogFile closes a JSON log file
func CloseJSONLogFile(logFile *os.File) error {
	// Write closing bracket for JSON array
	if _, err := logFile.WriteString("\n]\n"); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	// Close file
	if err := logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}

	return nil
}

// shouldLog determines whether a log entry should be output based on its level
func (l *UpdateLogger) shouldLog(level LogLevel) bool {
	switch l.MinLevel {
	case LogLevelDebug:
		return true
	case LogLevelInfo:
		return level != LogLevelDebug
	case LogLevelWarning:
		return level == LogLevelWarning || level == LogLevelError
	case LogLevelError:
		return level == LogLevelError
	default:
		return true
	}
}

// Log logs a message
func (l *UpdateLogger) Log(level LogLevel, component, message string, transactionID string, details map[string]interface{}) {
	// Check if this log level should be output
	if !l.shouldLog(level) {
		return
	}

	// Create log entry
	entry := &LogEntry{
		Timestamp:     time.Now(),
		Level:         level,
		Message:       message,
		Component:     component,
		TransactionID: transactionID,
	}

	// Include details if enabled
	if l.IncludeDetails && details != nil {
		entry.Details = details
	}

	// Format log entry for text output
	var levelPrefix string
	switch level {
	case LogLevelDebug:
		levelPrefix = "DEBUG"
	case LogLevelInfo:
		levelPrefix = "INFO"
	case LogLevelWarning:
		levelPrefix = "WARNING"
	case LogLevelError:
		levelPrefix = "ERROR"
	default:
		levelPrefix = "UNKNOWN"
	}

	// Write text log entry
	fmt.Fprintf(l.Writer, "[%s] [%s] [%s] %s", 
		entry.Timestamp.Format(time.RFC3339),
		levelPrefix,
		component,
		message)
	
	// Include transaction ID if provided
	if transactionID != "" {
		fmt.Fprintf(l.Writer, " (Transaction: %s)", transactionID)
	}
	
	// End line
	fmt.Fprintln(l.Writer)

	// Write JSON log entry if JSON writer is provided
	if l.JSONWriter != nil {
		// Marshal log entry to JSON
		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(l.Writer, "[%s] [ERROR] [Logger] Failed to marshal log entry to JSON: %v\n", 
				time.Now().Format(time.RFC3339), err)
			return
		}

		// Write JSON log entry
		if _, err := l.JSONWriter.Write(data); err != nil {
			fmt.Fprintf(l.Writer, "[%s] [ERROR] [Logger] Failed to write JSON log entry: %v\n", 
				time.Now().Format(time.RFC3339), err)
			return
		}

		// Write newline and comma for JSON array
		if _, err := l.JSONWriter.Write([]byte(",\n")); err != nil {
			fmt.Fprintf(l.Writer, "[%s] [ERROR] [Logger] Failed to write JSON log entry: %v\n", 
				time.Now().Format(time.RFC3339), err)
			return
		}
	}
}

// Debug logs a debug message
func (l *UpdateLogger) Debug(component, message string, transactionID string, details map[string]interface{}) {
	l.Log(LogLevelDebug, component, message, transactionID, details)
}

// Info logs an informational message
func (l *UpdateLogger) Info(component, message string, transactionID string, details map[string]interface{}) {
	l.Log(LogLevelInfo, component, message, transactionID, details)
}

// Warning logs a warning message
func (l *UpdateLogger) Warning(component, message string, transactionID string, details map[string]interface{}) {
	l.Log(LogLevelWarning, component, message, transactionID, details)
}

// Error logs an error message
func (l *UpdateLogger) Error(component, message string, transactionID string, details map[string]interface{}) {
	l.Log(LogLevelError, component, message, transactionID, details)
}

// AuditEvent represents an audit event for the update process
type AuditEvent struct {
	// Timestamp is the time the event occurred
	Timestamp time.Time `json:"timestamp"`
	// EventType is the type of event
	EventType string `json:"event_type"`
	// Component is the component associated with the event
	Component string `json:"component"`
	// User is the user who triggered the event
	User string `json:"user,omitempty"`
	// TransactionID is the ID of the transaction associated with the event
	TransactionID string `json:"transaction_id,omitempty"`
	// PackageID is the ID of the package associated with the event
	PackageID string `json:"package_id,omitempty"`
	// Details contains additional details about the event
	Details map[string]interface{} `json:"details,omitempty"`
}

// AuditLogger handles audit logging for update operations
type AuditLogger struct {
	// Writer is the writer for audit log output
	Writer io.Writer
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(writer io.Writer) *AuditLogger {
	return &AuditLogger{
		Writer: writer,
	}
}

// LogEvent logs an audit event
func (l *AuditLogger) LogEvent(eventType, component, user, transactionID, packageID string, details map[string]interface{}) {
	// Create audit event
	event := &AuditEvent{
		Timestamp:     time.Now(),
		EventType:     eventType,
		Component:     component,
		User:          user,
		TransactionID: transactionID,
		PackageID:     packageID,
		Details:       details,
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] [AuditLogger] Failed to marshal audit event to JSON: %v\n", 
			time.Now().Format(time.RFC3339), err)
		return
	}

	// Write JSON event
	if _, err := l.Writer.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] [AuditLogger] Failed to write audit event: %v\n", 
			time.Now().Format(time.RFC3339), err)
		return
	}

	// Write newline
	if _, err := l.Writer.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] [AuditLogger] Failed to write audit event: %v\n", 
			time.Now().Format(time.RFC3339), err)
		return
	}
}
