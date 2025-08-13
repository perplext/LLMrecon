// Package adapter provides adapters between database interfaces and domain models
package adapter

import (
	"context"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// IncidentEvent represents a security incident
type IncidentEvent struct {
	ID          string
	Title       string
	Description string
	Severity    string
	Status      string
	ReportedBy  string
	AssignedTo  string
	CreatedAt   string
	UpdatedAt   string
	ResolvedAt  string
	Metadata    map[string]interface{}
}

// IncidentStore defines the interface for security incident storage operations
type IncidentStore interface {
	// CreateIncident creates a new security incident
	CreateIncident(ctx context.Context, incident *IncidentEvent) error
	
	// GetIncidentByID retrieves a security incident by ID
	GetIncidentByID(ctx context.Context, id string) (*IncidentEvent, error)
	
	// UpdateIncident updates an existing security incident
	UpdateIncident(ctx context.Context, incident *IncidentEvent) error
	
	// DeleteIncident deletes a security incident
	DeleteIncident(ctx context.Context, id string) error
	
	// ListIncidents lists security incidents with optional filtering
	ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*IncidentEvent, int, error)
	
	// Close closes the incident store
	Close() error
}

// IncidentStoreAdapter adapts between the IncidentStore and domain types
type IncidentStoreAdapter struct {
	store IncidentStore
}

// NewIncidentStoreAdapter creates a new incident store adapter
func NewIncidentStoreAdapter(store IncidentStore) interfaces.IncidentStore {
	return &IncidentStoreAdapter{
		store: store,
	}
}

// convertInterfacesIncidentToModelIncident converts an interfaces.SecurityIncident to an IncidentEvent
func convertInterfacesIncidentToModelIncident(incident *interfaces.SecurityIncident) *IncidentEvent {
	if incident == nil {
		return nil
	}

	var resolvedAt string
	if !incident.ResolvedAt.IsZero() {
		resolvedAt = incident.ResolvedAt.Format(time.RFC3339)
	}

	return &IncidentEvent{
		ID:          incident.ID,
		Title:       incident.Title,
		Description: incident.Description,
		Severity:    string(incident.Severity),
		Status:      string(incident.Status),
		ReportedBy:  incident.ReportedBy,
		AssignedTo:  incident.AssignedTo,
		CreatedAt:   incident.ReportedAt.Format(time.RFC3339),
		UpdatedAt:   incident.ReportedAt.Format(time.RFC3339), // Using ReportedAt as UpdatedAt is not in the model
		ResolvedAt:  resolvedAt,
		Metadata:    incident.Metadata,
	}
}

// convertModelIncidentToInterfacesIncident converts an IncidentEvent to an interfaces.SecurityIncident
func convertModelIncidentToInterfacesIncident(incident *IncidentEvent) *interfaces.SecurityIncident {
	if incident == nil {
		return nil
	}

	interfacesIncident := &interfaces.SecurityIncident{
		ID:          incident.ID,
		Title:       incident.Title,
		Description: incident.Description,
		Severity:    models.SecurityIncidentSeverity(incident.Severity),
		Status:      models.SecurityIncidentStatus(incident.Status),
		ReportedBy:  incident.ReportedBy,
		AssignedTo:  incident.AssignedTo,
		Metadata:    incident.Metadata,
	}

	// Parse time fields
	if incident.CreatedAt != "" {
		interfacesIncident.ReportedAt, _ = time.Parse(time.RFC3339, incident.CreatedAt)
	}
	if incident.ResolvedAt != "" {
		interfacesIncident.ResolvedAt, _ = time.Parse(time.RFC3339, incident.ResolvedAt)
	}

	return interfacesIncident
}

// convertModelIncidentsToInterfacesIncidents converts a slice of IncidentEvent to a slice of interfaces.SecurityIncident
func convertModelIncidentsToInterfacesIncidents(incidents []*IncidentEvent) []*interfaces.SecurityIncident {
	if incidents == nil {
		return nil
	}
	result := make([]*interfaces.SecurityIncident, len(incidents))
	for i, incident := range incidents {
		result[i] = convertModelIncidentToInterfacesIncident(incident)
	}
	return result
}

// CreateIncident creates a new security incident
func (a *IncidentStoreAdapter) CreateIncident(ctx context.Context, incident *interfaces.SecurityIncident) error {
	return a.store.CreateIncident(ctx, convertInterfacesIncidentToModelIncident(incident))
}

// GetIncidentByID retrieves a security incident by ID
func (a *IncidentStoreAdapter) GetIncidentByID(ctx context.Context, id string) (*interfaces.SecurityIncident, error) {
	incident, err := a.store.GetIncidentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertModelIncidentToInterfacesIncident(incident), nil
}

// UpdateIncident updates an existing security incident
func (a *IncidentStoreAdapter) UpdateIncident(ctx context.Context, incident *interfaces.SecurityIncident) error {
	return a.store.UpdateIncident(ctx, convertInterfacesIncidentToModelIncident(incident))
}

// DeleteIncident deletes a security incident
func (a *IncidentStoreAdapter) DeleteIncident(ctx context.Context, id string) error {
	return a.store.DeleteIncident(ctx, id)
}

// ListIncidents lists security incidents with optional filtering
func (a *IncidentStoreAdapter) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*interfaces.SecurityIncident, int, error) {
	incidents, count, err := a.store.ListIncidents(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return convertModelIncidentsToInterfacesIncidents(incidents), count, nil
}

// Close closes the incident store
func (a *IncidentStoreAdapter) Close() error {
	return a.store.Close()
}
