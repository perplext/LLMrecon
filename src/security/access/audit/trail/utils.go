// Package trail provides audit trail utilities
package trail

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateAuditID generates a unique audit ID
func GenerateAuditID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	return fmt.Sprintf("%x", bytes), nil
}

// FormatTimestamp formats a timestamp for audit logs
func FormatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ValidateAuditLog validates an audit log entry
func ValidateAuditLog(log *AuditLog) error {
	if log == nil {
		return fmt.Errorf("audit log is nil")
	}
	
	if log.ID == "" {
		return fmt.Errorf("audit log ID is required")
	}
	
	if log.Operation == "" {
		return fmt.Errorf("audit log operation is required")
	}
	
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	
	return nil
}
