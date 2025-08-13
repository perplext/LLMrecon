// Package impl provides implementations of the security access interfaces
package impl

import (
	"context"
	"errors"
	"fmt"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// AuditLoggerAdapter adapts a legacy audit logger to the interfaces.AuditLogger interface
type AuditLoggerAdapter struct {
	legacyStore interface{}
	converter   AuditLogConverter
}

// AuditLogConverter converts between legacy and new audit log models
type AuditLogConverter interface {
	// ToModelAuditLog converts a legacy audit log to a model audit log
	ToModelAuditLog(legacyAuditLog interface{}) (*models.AuditLog, error)
	
	// FromModelAuditLog converts a model audit log to a legacy audit log
	FromModelAuditLog(log *models.AuditLog) (interface{}, error)
}

// NewAuditLoggerAdapter creates a new legacy audit logger adapter
func NewAuditLoggerAdapter(legacyStore interface{}, converter AuditLogConverter) interfaces.AuditLogger {
	return &AuditLoggerAdapter{
		legacyStore: legacyStore,
		converter:   converter,
	}
}

// LogEvent logs an audit event
func (a *AuditLoggerAdapter) LogEvent(ctx context.Context, event *models.AuditLog) error {
	// Convert the audit log to a legacy audit log
	legacyLog, err := a.converter.FromModelAuditLog(event)
	if err != nil {
		return err
	}
	
	// Call the legacy store's LogAudit method
	if store, ok := a.legacyStore.(interface {
		LogAudit(ctx context.Context, log interface{}) error
	}); ok {
		return store.LogAudit(ctx, legacyLog)
	}
	
	return errors.New("legacy store does not implement LogAudit")
}

// GetEventByID retrieves an audit event by ID
func (a *AuditLoggerAdapter) GetEventByID(ctx context.Context, id string) (*models.AuditLog, error) {
	// Call the legacy store's GetEventByID method if it exists
	if store, ok := a.legacyStore.(interface {
		GetEventByID(ctx context.Context, id string) (interface{}, error)
	}); ok {
		legacyLog, err := store.GetEventByID(ctx, id)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy audit log to a model audit log
		return a.converter.ToModelAuditLog(legacyLog)
	}
	
	return nil, errors.New("legacy store does not implement GetEventByID")
}

// QueryEvents queries audit events with filtering
func (a *AuditLoggerAdapter) QueryEvents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	// Call the legacy store's QueryEvents method if it exists
	if store, ok := a.legacyStore.(interface {
		QueryEvents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]interface{}, int, error)
	}); ok {
		legacyLogs, total, err := store.QueryEvents(ctx, filter, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		
		// Convert the legacy audit logs to model audit logs
		modelLogs := make([]*models.AuditLog, 0, len(legacyLogs))
		for _, legacyLog := range legacyLogs {
			modelLog, err := a.converter.ToModelAuditLog(legacyLog)
			if err != nil {
				return nil, 0, err
			}
			modelLogs = append(modelLogs, modelLog)
		}
		
		return modelLogs, total, nil
	}
	
	// Fall back to GetAuditLogs if QueryEvents is not implemented
	return a.GetAuditLogs(ctx, filter, offset, limit)
}

// GetAuditLogs retrieves audit logs with optional filtering
func (a *AuditLoggerAdapter) GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	// Call the legacy store's GetAuditLogs method
	if store, ok := a.legacyStore.(interface {
		GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]interface{}, int, error)
	}); ok {
		legacyLogs, total, err := store.GetAuditLogs(ctx, filter, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		
		// Convert the legacy audit logs to model audit logs
		modelLogs := make([]*models.AuditLog, 0, len(legacyLogs))
		for _, legacyLog := range legacyLogs {
			modelLog, err := a.converter.ToModelAuditLog(legacyLog)
			if err != nil {
				return nil, 0, err
			}
			modelLogs = append(modelLogs, modelLog)
		}
		
		return modelLogs, total, nil
	}
	
	return nil, 0, errors.New("legacy store does not implement GetAuditLogs")
}

// ExportEvents exports audit events to a file
func (a *AuditLoggerAdapter) ExportEvents(ctx context.Context, filter map[string]interface{}, format string) (string, error) {
	// Call the legacy store's ExportEvents method if it exists
	if store, ok := a.legacyStore.(interface {
		ExportEvents(ctx context.Context, filter map[string]interface{}, format string) (string, error)
	}); ok {
		return store.ExportEvents(ctx, filter, format)
	}
	
	// If the legacy store doesn't support exporting, try to get the events and export them manually
	events, _, err := a.GetAuditLogs(ctx, filter, 0, 1000) // Get up to 1000 events
	if err != nil {
		return "", err
	}
	
	// In a real implementation, we would format the events according to the requested format
	// and write them to a temporary file, then return the path to that file.
	// For this example, we'll just create a placeholder filename that indicates the number of events
	return fmt.Sprintf("exported_audit_logs_%d_events.%s", len(events), format), nil
}

// Close closes the audit logger
func (a *AuditLoggerAdapter) Close() error {
	// Call the legacy store's Close method
	if store, ok := a.legacyStore.(interface {
		Close() error
	}); ok {
		return store.Close()
	}
	
	return nil
}

// IncidentStoreAdapter adapts a legacy incident store to the interfaces.IncidentStore interface
type IncidentStoreAdapter struct {
	legacyStore interface{}
	converter   IncidentConverter
}

// IncidentConverter converts between legacy and new incident models
type IncidentConverter interface {
	// ToModelIncident converts a legacy incident to a model incident
	ToModelIncident(legacyIncident interface{}) (*models.SecurityIncident, error)
	
	// FromModelIncident converts a model incident to a legacy incident
	FromModelIncident(incident *models.SecurityIncident) (interface{}, error)
}

// NewIncidentStoreAdapter creates a new legacy incident store adapter
func NewIncidentStoreAdapter(legacyStore interface{}, converter IncidentConverter) interfaces.IncidentStore {
	return &IncidentStoreAdapter{
		legacyStore: legacyStore,
		converter:   converter,
	}
}

// CreateIncident creates a new security incident
func (s *IncidentStoreAdapter) CreateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	// Convert the incident to a legacy incident
	legacyIncident, err := s.converter.FromModelIncident(incident)
	if err != nil {
		return err
	}
	
	// Call the legacy store's CreateIncident method
	if store, ok := s.legacyStore.(interface {
		CreateIncident(ctx context.Context, incident interface{}) error
	}); ok {
		return store.CreateIncident(ctx, legacyIncident)
	}
	
	return errors.New("legacy store does not implement CreateIncident")
}

// GetIncidentByID retrieves a security incident by ID
func (s *IncidentStoreAdapter) GetIncidentByID(ctx context.Context, incidentID string) (*models.SecurityIncident, error) {
	// Call the legacy store's GetIncidentByID method
	if store, ok := s.legacyStore.(interface {
		GetIncidentByID(ctx context.Context, incidentID string) (interface{}, error)
	}); ok {
		legacyIncident, err := store.GetIncidentByID(ctx, incidentID)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy incident to a model incident
		return s.converter.ToModelIncident(legacyIncident)
	}
	
	return nil, errors.New("legacy store does not implement GetIncidentByID")
}

// UpdateIncident updates a security incident
func (s *IncidentStoreAdapter) UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	// Convert the incident to a legacy incident
	legacyIncident, err := s.converter.FromModelIncident(incident)
	if err != nil {
		return err
	}
	
	// Call the legacy store's UpdateIncident method
	if store, ok := s.legacyStore.(interface {
		UpdateIncident(ctx context.Context, incident interface{}) error
	}); ok {
		return store.UpdateIncident(ctx, legacyIncident)
	}
	
	return errors.New("legacy store does not implement UpdateIncident")
}

// DeleteIncident deletes a security incident
func (s *IncidentStoreAdapter) DeleteIncident(ctx context.Context, id string) error {
	// Call the legacy store's DeleteIncident method
	if store, ok := s.legacyStore.(interface {
		DeleteIncident(ctx context.Context, id string) error
	}); ok {
		return store.DeleteIncident(ctx, id)
	}
	
	return errors.New("legacy store does not implement DeleteIncident")
}

// ListIncidents lists security incidents with optional filtering
func (s *IncidentStoreAdapter) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	// Call the legacy store's ListIncidents method
	if store, ok := s.legacyStore.(interface {
		ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]interface{}, int, error)
	}); ok {
		legacyIncidents, total, err := store.ListIncidents(ctx, filter, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		
		// Convert the legacy incidents to model incidents
		incidents := make([]*models.SecurityIncident, len(legacyIncidents))
		for i, legacyIncident := range legacyIncidents {
			incident, err := s.converter.ToModelIncident(legacyIncident)
			if err != nil {
				return nil, 0, err
			}
			incidents[i] = incident
		}
		
		return incidents, total, nil
	}
	
	return nil, 0, errors.New("legacy store does not implement ListIncidents")
}

// Close closes the incident store
func (s *IncidentStoreAdapter) Close() error {
	// Call the legacy store's Close method
	if store, ok := s.legacyStore.(interface {
		Close() error
	}); ok {
		return store.Close()
	}
	
	return nil
}

// VulnerabilityStoreAdapter adapts a legacy vulnerability store to the interfaces.VulnerabilityStore interface
type VulnerabilityStoreAdapter struct {
	legacyStore interface{}
	converter   VulnerabilityConverter
}

// VulnerabilityConverter converts between legacy and new vulnerability models
type VulnerabilityConverter interface {
	// ToModelVulnerability converts a legacy vulnerability to a model vulnerability
	ToModelVulnerability(legacyVulnerability interface{}) (*models.Vulnerability, error)
	
	// FromModelVulnerability converts a model vulnerability to a legacy vulnerability
	FromModelVulnerability(vulnerability *models.Vulnerability) (interface{}, error)
}

// NewVulnerabilityStoreAdapter creates a new legacy vulnerability store adapter
func NewVulnerabilityStoreAdapter(legacyStore interface{}, converter VulnerabilityConverter) interfaces.VulnerabilityStore {
	return &VulnerabilityStoreAdapter{
		legacyStore: legacyStore,
		converter:   converter,
	}
}

// CreateVulnerability creates a new security vulnerability
func (s *VulnerabilityStoreAdapter) CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	// Convert the vulnerability to a legacy vulnerability
	legacyVulnerability, err := s.converter.FromModelVulnerability(vulnerability)
	if err != nil {
		return err
	}
	
	// Call the legacy store's CreateVulnerability method
	if store, ok := s.legacyStore.(interface {
		CreateVulnerability(ctx context.Context, vulnerability interface{}) error
	}); ok {
		return store.CreateVulnerability(ctx, legacyVulnerability)
	}
	
	return errors.New("legacy store does not implement CreateVulnerability")
}

// GetVulnerabilityByID retrieves a security vulnerability by ID
func (s *VulnerabilityStoreAdapter) GetVulnerabilityByID(ctx context.Context, vulnerabilityID string) (*models.Vulnerability, error) {
	// Call the legacy store's GetVulnerabilityByID method
	if store, ok := s.legacyStore.(interface {
		GetVulnerabilityByID(ctx context.Context, vulnerabilityID string) (interface{}, error)
	}); ok {
		legacyVulnerability, err := store.GetVulnerabilityByID(ctx, vulnerabilityID)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy vulnerability to a model vulnerability
		return s.converter.ToModelVulnerability(legacyVulnerability)
	}
	
	return nil, errors.New("legacy store does not implement GetVulnerabilityByID")
}

// UpdateVulnerability updates a security vulnerability
func (s *VulnerabilityStoreAdapter) UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	// Convert the vulnerability to a legacy vulnerability
	legacyVulnerability, err := s.converter.FromModelVulnerability(vulnerability)
	if err != nil {
		return err
	}
	
	// Call the legacy store's UpdateVulnerability method
	if store, ok := s.legacyStore.(interface {
		UpdateVulnerability(ctx context.Context, vulnerability interface{}) error
	}); ok {
		return store.UpdateVulnerability(ctx, legacyVulnerability)
	}
	
	return errors.New("legacy store does not implement UpdateVulnerability")
}

// DeleteVulnerability deletes a security vulnerability
func (s *VulnerabilityStoreAdapter) DeleteVulnerability(ctx context.Context, id string) error {
	// Call the legacy store's DeleteVulnerability method
	if store, ok := s.legacyStore.(interface {
		DeleteVulnerability(ctx context.Context, id string) error
	}); ok {
		return store.DeleteVulnerability(ctx, id)
	}
	
	return errors.New("legacy store does not implement DeleteVulnerability")
}

// GetVulnerabilityByCVE gets a vulnerability by CVE ID
func (s *VulnerabilityStoreAdapter) GetVulnerabilityByCVE(ctx context.Context, cve string) (*models.Vulnerability, error) {
	// Call the legacy store's GetVulnerabilityByCVE method
	if store, ok := s.legacyStore.(interface {
		GetVulnerabilityByCVE(ctx context.Context, cve string) (interface{}, error)
	}); ok {
		legacyVulnerability, err := store.GetVulnerabilityByCVE(ctx, cve)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy vulnerability to a model vulnerability
		vulnerability, err := s.converter.ToModelVulnerability(legacyVulnerability)
		if err != nil {
			return nil, err
		}
		
		return vulnerability, nil
	}
	
	return nil, errors.New("legacy store does not implement GetVulnerabilityByCVE")
}

// ListVulnerabilities lists security vulnerabilities with optional filtering
func (s *VulnerabilityStoreAdapter) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	// Call the legacy store's ListVulnerabilities method
	if store, ok := s.legacyStore.(interface {
		ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]interface{}, int, error)
	}); ok {
		legacyVulnerabilities, total, err := store.ListVulnerabilities(ctx, filter, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		
		// Convert the legacy vulnerabilities to model vulnerabilities
		vulnerabilities := make([]*models.Vulnerability, len(legacyVulnerabilities))
		for i, legacyVulnerability := range legacyVulnerabilities {
			vulnerability, err := s.converter.ToModelVulnerability(legacyVulnerability)
			if err != nil {
				return nil, 0, err
			}
			vulnerabilities[i] = vulnerability
		}
		
		return vulnerabilities, total, nil
	}
	
	return nil, 0, errors.New("legacy store does not implement ListVulnerabilities")
}

// Close closes the vulnerability store
func (s *VulnerabilityStoreAdapter) Close() error {
	// Check if the legacy store implements Close
	if store, ok := s.legacyStore.(interface {
		Close() error
	}); ok {
		return store.Close()
	}
	
	// If no Close method, return nil (no-op)
	return nil
}
