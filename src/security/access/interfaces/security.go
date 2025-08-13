// Package interfaces defines the interfaces for the access control system
package interfaces

import (
	"context"

	"github.com/perplext/LLMrecon/src/security/access/models"
)

// SecurityManager defines the interface for security management operations
type SecurityManager interface {
	// Initialize initializes the security manager
	Initialize(ctx context.Context) error

	// CreateIncident creates a new security incident
	CreateIncident(ctx context.Context, incident *models.SecurityIncident) error

	// GetIncidentByID retrieves a security incident by ID
	GetIncidentByID(ctx context.Context, id string) (*models.SecurityIncident, error)

	// UpdateIncident updates an existing security incident
	UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error

	// DeleteIncident deletes a security incident
	DeleteIncident(ctx context.Context, id string) error

	// ListIncidents lists security incidents with filtering
	ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error)

	// CreateVulnerability creates a new vulnerability
	CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error

	// GetVulnerabilityByID retrieves a vulnerability by ID
	GetVulnerabilityByID(ctx context.Context, id string) (*models.Vulnerability, error)

	// UpdateVulnerability updates an existing vulnerability
	UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error

	// DeleteVulnerability deletes a vulnerability
	DeleteVulnerability(ctx context.Context, id string) error

	// ListVulnerabilities lists vulnerabilities with filtering
	ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error)

	// UpdateVulnerabilityStatus updates the status of a vulnerability
	UpdateVulnerabilityStatus(ctx context.Context, id string, status models.VulnerabilityStatus) error

	// Close closes the security manager
	Close() error
}
