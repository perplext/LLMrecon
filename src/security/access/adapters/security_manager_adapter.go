// Package adapters provides adapter implementations for security interfaces
package adapters

import (
	"context"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// SecurityManagerAdapter adapts the legacy security manager to the new interface
type SecurityManagerAdapter struct {
	incidentStore      interfaces.IncidentStore
	vulnerabilityStore interfaces.VulnerabilityStore
	auditLogger        interfaces.AuditLogger
}

// NewSecurityManagerAdapter creates a new security manager adapter
func NewSecurityManagerAdapter(
	incidentStore interfaces.IncidentStore,
	vulnerabilityStore interfaces.VulnerabilityStore,
	auditLogger interfaces.AuditLogger,
) *SecurityManagerAdapter {
	return &SecurityManagerAdapter{
		incidentStore:      incidentStore,
		vulnerabilityStore: vulnerabilityStore,
		auditLogger:        auditLogger,
	}
}

// CreateIncident creates a new security incident
func (a *SecurityManagerAdapter) CreateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	return a.incidentStore.CreateIncident(ctx, incident)
}

// GetIncidentByID retrieves a security incident by ID
func (a *SecurityManagerAdapter) GetIncidentByID(ctx context.Context, id string) (*models.SecurityIncident, error) {
	return a.incidentStore.GetIncidentByID(ctx, id)
}

// UpdateIncident updates an existing security incident
func (a *SecurityManagerAdapter) UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	return a.incidentStore.UpdateIncident(ctx, incident)
}

// DeleteIncident deletes a security incident
func (a *SecurityManagerAdapter) DeleteIncident(ctx context.Context, id string) error {
	return a.incidentStore.DeleteIncident(ctx, id)
}

// ListIncidents lists security incidents with filtering
func (a *SecurityManagerAdapter) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	return a.incidentStore.ListIncidents(ctx, filter, offset, limit)
}

// CreateVulnerability creates a new vulnerability
func (a *SecurityManagerAdapter) CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	return a.vulnerabilityStore.CreateVulnerability(ctx, vulnerability)
}

// GetVulnerabilityByID retrieves a vulnerability by ID
func (a *SecurityManagerAdapter) GetVulnerabilityByID(ctx context.Context, id string) (*models.Vulnerability, error) {
	return a.vulnerabilityStore.GetVulnerabilityByID(ctx, id)
}

// UpdateVulnerability updates an existing vulnerability
func (a *SecurityManagerAdapter) UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	return a.vulnerabilityStore.UpdateVulnerability(ctx, vulnerability)
}

// DeleteVulnerability deletes a vulnerability
func (a *SecurityManagerAdapter) DeleteVulnerability(ctx context.Context, id string) error {
	return a.vulnerabilityStore.DeleteVulnerability(ctx, id)
}

// ListVulnerabilities lists vulnerabilities with filtering
func (a *SecurityManagerAdapter) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	return a.vulnerabilityStore.ListVulnerabilities(ctx, filter, offset, limit)
}

// UpdateVulnerabilityStatus updates the status of a vulnerability
func (a *SecurityManagerAdapter) UpdateVulnerabilityStatus(ctx context.Context, id string, status models.VulnerabilityStatus) error {
	vulnerability, err := a.GetVulnerabilityByID(ctx, id)
	if err != nil {
		return err
	}
	
	vulnerability.Status = status
	return a.UpdateVulnerability(ctx, vulnerability)
}
