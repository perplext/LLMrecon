// Package trail provides audit trail management
package trail

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages audit trail operations
type Manager struct {
	trail  *AuditTrail
	logger *FileLogger
	mutex  sync.Mutex
}

// NewManager creates a new audit trail manager
func NewManager(logPath string) (*Manager, error) {
	logger, err := NewFileLogger(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file logger: %w", err)
	}
	
	trail := NewAuditTrail(logger)
	
	return &Manager{
		trail:  trail,
		logger: logger,
	}, nil
}

// LogOperation logs an operation to the audit trail
func (m *Manager) LogOperation(ctx context.Context, log *AuditLog) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	return m.trail.LogOperation(ctx, log)
}

// Close closes the audit trail manager
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if err := m.trail.Close(); err != nil {
		return fmt.Errorf("failed to close audit trail: %w", err)
	}
	
	if err := m.logger.Close(); err != nil {
		return fmt.Errorf("failed to close logger: %w", err)
	}
	
	return nil
}

// GetTrail returns the audit trail
func (m *Manager) GetTrail() *AuditTrail {
	return m.trail
}
