// Package audit provides audit logging functionality for security-sensitive operations.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CredentialAuditEvent represents an audit event for credential operations
type CredentialAuditEvent struct {
	// Timestamp is the time of the event
	Timestamp time.Time `json:"timestamp"`
	// EventType is the type of event (access, create, update, delete, rotate)
	EventType string `json:"event_type"`
	// CredentialID is the ID of the credential
	CredentialID string `json:"credential_id"`
	// Service is the service the credential is for
	Service string `json:"service,omitempty"`
	// UserID is the ID of the user who performed the operation
	UserID string `json:"user_id,omitempty"`
	// SourceIP is the IP address of the user
	SourceIP string `json:"source_ip,omitempty"`
	// Success indicates whether the operation was successful
	Success bool `json:"success"`
	// ErrorMessage is the error message if the operation failed
	ErrorMessage string `json:"error_message,omitempty"`
	// Metadata is additional metadata for the event
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CredentialAuditLogger logs credential audit events
type CredentialAuditLogger struct {
	// filePath is the path to the audit log file
	filePath string
	// mutex protects the logger during operations
	mutex sync.Mutex
	// userIDProvider provides the current user ID
	userIDProvider func() string
	// sourceIPProvider provides the source IP address
	sourceIPProvider func() string
}

// CredentialAuditLoggerOptions contains options for creating a credential audit logger
type CredentialAuditLoggerOptions struct {
	// UserIDProvider provides the current user ID
	UserIDProvider func() string
	// SourceIPProvider provides the source IP address
	SourceIPProvider func() string
}

// NewCredentialAuditLogger creates a new credential audit logger
func NewCredentialAuditLogger(filePath string, options CredentialAuditLoggerOptions) (*CredentialAuditLogger, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create logger
	logger := &CredentialAuditLogger{
		filePath:         filePath,
		userIDProvider:   options.UserIDProvider,
		sourceIPProvider: options.SourceIPProvider,
	}

	return logger, nil
}

// LogCredentialEvent logs a credential event
func (l *CredentialAuditLogger) LogCredentialEvent(eventType string, credentialID string, service string, success bool, errorMessage string, metadata map[string]string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Create event
	event := CredentialAuditEvent{
		Timestamp:    time.Now().UTC(),
		EventType:    eventType,
		CredentialID: credentialID,
		Service:      service,
		Success:      success,
		ErrorMessage: errorMessage,
		Metadata:     metadata,
	}

	// Add user ID if available
	if l.userIDProvider != nil {
		event.UserID = l.userIDProvider()
	}

	// Add source IP if available
	if l.sourceIPProvider != nil {
		event.SourceIP = l.sourceIPProvider()
	}

	// Convert to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer func() { 
		if err := file.Close(); err != nil { 
			fmt.Printf("Failed to close: %v\n", err) 
		} 
	}()

	// Write event
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to audit log file: %w", err)
	}

	return nil
}

// LogCredentialAccess logs a credential access event
func (l *CredentialAuditLogger) LogCredentialAccess(credentialID, service, operation string) error {
	return l.LogCredentialEvent(
		fmt.Sprintf("%s_credential", operation),
		credentialID,
		service,
		true,
		"",
		map[string]string{
			"operation": operation,
		},
	)
}

// LogCredentialError logs a credential error event
func (l *CredentialAuditLogger) LogCredentialError(credentialID, service, operation string, err error) error {
	return l.LogCredentialEvent(
		fmt.Sprintf("%s_credential_error", operation),
		credentialID,
		service,
		false,
		err.Error(),
		map[string]string{
			"operation": operation,
		},
	)
}

// LogAlert logs an alert event
func (l *CredentialAuditLogger) LogAlert(message, alertType string, metadata map[string]string) error {
	return l.LogCredentialEvent(
		"alert",
		"",
		"",
		true,
		"",
		map[string]string{
			"alert_type": alertType,
			"message":    message,
			"metadata":   fmt.Sprintf("%v", metadata),
		},
	)
}

// GetAuditEvents returns audit events from the log file
func (l *CredentialAuditLogger) GetAuditEvents(limit int, filter map[string]string) ([]CredentialAuditEvent, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(l.filePath); os.IsNotExist(err) {
		return []CredentialAuditEvent{}, nil
	}

	// Open file
	file, err := os.Open(filepath.Clean(l.filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer func() { 
		if err := file.Close(); err != nil { 
			fmt.Printf("Failed to close: %v\n", err) 
		} 
	}()

	// Read events
	var events []CredentialAuditEvent
	decoder := json.NewDecoder(file)
	for {
		var event CredentialAuditEvent
		if err := decoder.Decode(&event); err != nil {
			break
		}

		// Apply filters
		if filter != nil {
			match := true
			for key, value := range filter {
				switch key {
				case "event_type":
					if event.EventType != value {
						match = false
					}
				case "credential_id":
					if event.CredentialID != value {
						match = false
					}
				case "service":
					if event.Service != value {
						match = false
					}
				case "user_id":
					if event.UserID != value {
						match = false
					}
				case "success":
					if value == "true" && !event.Success {
						match = false
					} else if value == "false" && event.Success {
						match = false
					}
				}
				if !match {
					break
				}
			}
			if !match {
				continue
			}
		}

		events = append(events, event)

		// Apply limit
		if limit > 0 && len(events) >= limit {
			break
		}
	}

	return events, nil
}

// RotateLogFile rotates the audit log file
func (l *CredentialAuditLogger) RotateLogFile() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(l.filePath); os.IsNotExist(err) {
		return nil
	}

	// Create backup filename
	backupPath := fmt.Sprintf("%s.%d", l.filePath, time.Now().Unix())

	// Rename current file
	if err := os.Rename(l.filePath, backupPath); err != nil {
		return fmt.Errorf("failed to rotate audit log file: %w", err)
	}

	return nil
}
