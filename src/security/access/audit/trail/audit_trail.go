// Package trail provides audit trail functionality
package trail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// AuditTrail manages audit logging
type AuditTrail struct {
	Writer io.Writer
	mutex  sync.Mutex

}
// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Operation    string                 `json:"operation"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	IPAddress    string                 `json:"ip_address"`
	Details      map[string]interface{} `json:"details"`
}


// NewAuditTrail creates a new audit trail
func NewAuditTrail(writer io.Writer) *AuditTrail {
	return &AuditTrail{
		Writer: writer,
	}
}

// LogOperation logs an operation to the audit trail
func (a *AuditTrail) LogOperation(ctx context.Context, log *AuditLog) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}
	
	_, err = a.Writer.Write(append(data, '\n'))
	return err

}// Close closes the audit trail
func (a *AuditTrail) Close() error {
	if closer, ok := a.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
