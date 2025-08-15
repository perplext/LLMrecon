// Package adapters provides adapter implementations for security interfaces
package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// InMemoryIncidentStoreAdapter is an in-memory implementation of IncidentStore
type InMemoryIncidentStoreAdapter struct {
	incidents map[string]*models.SecurityIncident
	mu        sync.RWMutex
}

// NewInMemoryIncidentStoreAdapter creates a new in-memory incident store adapter
func NewInMemoryIncidentStoreAdapter() *InMemoryIncidentStoreAdapter {
	return &InMemoryIncidentStoreAdapter{
		incidents: make(map[string]*models.SecurityIncident),
	}
}

// Close closes the incident store
func (a *InMemoryIncidentStoreAdapter) Close() error {
	// Nothing to close for in-memory store
	return nil
}

// CreateIncident creates a new security incident
func (a *InMemoryIncidentStoreAdapter) CreateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.incidents[incident.ID] = incident
	return nil
}

// GetIncidentByID retrieves a security incident by ID
func (a *InMemoryIncidentStoreAdapter) GetIncidentByID(ctx context.Context, id string) (*models.SecurityIncident, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	incident, ok := a.incidents[id]
	if !ok {
		return nil, interfaces.ErrNotFound
	}
	return incident, nil
}

// UpdateIncident updates an existing security incident
func (a *InMemoryIncidentStoreAdapter) UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if _, ok := a.incidents[incident.ID]; !ok {
		return interfaces.ErrNotFound
	}
	
	a.incidents[incident.ID] = incident
	return nil
}

// DeleteIncident deletes a security incident
func (a *InMemoryIncidentStoreAdapter) DeleteIncident(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if _, ok := a.incidents[id]; !ok {
		return interfaces.ErrNotFound
	}
	
	delete(a.incidents, id)
	return nil
}

// ListIncidents lists security incidents with filtering
func (a *InMemoryIncidentStoreAdapter) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	var results []*models.SecurityIncident
	
	// Simple implementation without filtering
	for _, incident := range a.incidents {
		results = append(results, incident)
	}
	
	total := len(results)
	
	// Apply pagination
	if offset < len(results) {
		end := offset + limit
		if end > len(results) {
			end = len(results)
		}
		results = results[offset:end]
	} else {
		results = []*models.SecurityIncident{}
	}
	
	return results, total, nil
}

// InMemoryVulnerabilityStoreAdapter is an in-memory implementation of VulnerabilityStore
type InMemoryVulnerabilityStoreAdapter struct {
	vulnerabilities map[string]*models.Vulnerability
	mu              sync.RWMutex
}

// NewInMemoryVulnerabilityStoreAdapter creates a new in-memory vulnerability store adapter
func NewInMemoryVulnerabilityStoreAdapter() *InMemoryVulnerabilityStoreAdapter {
	return &InMemoryVulnerabilityStoreAdapter{
		vulnerabilities: make(map[string]*models.Vulnerability),
	}
}

// Close closes the vulnerability store
func (a *InMemoryVulnerabilityStoreAdapter) Close() error {
	// Nothing to close for in-memory store
	return nil
}

// CreateVulnerability creates a new vulnerability
func (a *InMemoryVulnerabilityStoreAdapter) CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.vulnerabilities[vulnerability.ID] = vulnerability
	return nil
}

// GetVulnerabilityByID retrieves a vulnerability by ID
func (a *InMemoryVulnerabilityStoreAdapter) GetVulnerabilityByID(ctx context.Context, id string) (*models.Vulnerability, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	vulnerability, ok := a.vulnerabilities[id]
	if !ok {
		return nil, interfaces.ErrNotFound
	}
	return vulnerability, nil
}

// UpdateVulnerability updates an existing vulnerability
func (a *InMemoryVulnerabilityStoreAdapter) UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if _, ok := a.vulnerabilities[vulnerability.ID]; !ok {
		return interfaces.ErrNotFound
	}
	
	a.vulnerabilities[vulnerability.ID] = vulnerability
	return nil
}

// DeleteVulnerability deletes a vulnerability
func (a *InMemoryVulnerabilityStoreAdapter) DeleteVulnerability(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if _, ok := a.vulnerabilities[id]; !ok {
		return interfaces.ErrNotFound
	}
	
	delete(a.vulnerabilities, id)
	return nil
}

// ListVulnerabilities lists vulnerabilities with filtering
func (a *InMemoryVulnerabilityStoreAdapter) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	var results []*models.Vulnerability
	
	// Simple implementation without filtering
	for _, vulnerability := range a.vulnerabilities {
		results = append(results, vulnerability)
	}
	
	total := len(results)
	
	// Apply pagination
	if offset < len(results) {
		end := offset + limit
		if end > len(results) {
			end = len(results)
		}
		results = results[offset:end]
	} else {
		results = []*models.Vulnerability{}
	}
	
	return results, total, nil
}

// InMemoryAuditLoggerAdapter is an in-memory implementation of AuditLogger
type InMemoryAuditLoggerAdapter struct {
	logs map[string]*models.AuditLog
	mu   sync.RWMutex

}
// NewInMemoryAuditLoggerAdapter creates a new in-memory audit logger adapter
func NewInMemoryAuditLoggerAdapter() *InMemoryAuditLoggerAdapter {
	return &InMemoryAuditLoggerAdapter{
		logs: make(map[string]*models.AuditLog),
	}
}

// LogEvent logs an audit event
func (a *InMemoryAuditLoggerAdapter) LogEvent(ctx context.Context, event *models.AuditLog) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.logs[event.ID] = event
	return nil
}

// GetEventByID retrieves an audit event by ID
func (a *InMemoryAuditLoggerAdapter) GetEventByID(ctx context.Context, id string) (*models.AuditLog, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	event, ok := a.logs[id]
	if !ok {
		return nil, interfaces.ErrNotFound
	}
	return event, nil
}

// QueryEvents queries audit events with filtering
func (a *InMemoryAuditLoggerAdapter) QueryEvents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	var results []*models.AuditLog
	
	// Simple implementation without filtering
	for _, event := range a.logs {
		results = append(results, event)
	}
	
	total := len(results)
	
	// Apply pagination
	if offset < len(results) {
		end := offset + limit
		if end > len(results) {
			end = len(results)
		}
		results = results[offset:end]
	} else {
		results = []*models.AuditLog{}
	}
	
	return results, total, nil
}

// ExportEvents exports audit events to a file
func (a *InMemoryAuditLoggerAdapter) ExportEvents(ctx context.Context, filter map[string]interface{}, format string) (string, error) {
	// In-memory implementation just returns a placeholder
	return "audit_export_" + time.Now().Format("20060102150405") + "." + format, nil
}

// Close closes the audit logger
func (a *InMemoryAuditLoggerAdapter) Close() error {
	return nil
}
