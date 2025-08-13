// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// SecurityManagerImpl implements the SecurityManager interface
type SecurityManagerImpl struct {
	mu             sync.RWMutex
	incidentStore  ModelsIncidentStore
	vulnStore      ModelsVulnerabilityStore
	auditLogger    ModelsAuditLogger
	initialized    bool
}

// ModelsIncidentStore defines the interface for storing and retrieving security incidents using models
type ModelsIncidentStore interface {
	// CreateIncident creates a new security incident
	CreateIncident(ctx context.Context, incident *models.SecurityIncident) error

	// GetIncidentByID retrieves a security incident by ID
	GetIncidentByID(ctx context.Context, incidentID string) (*models.SecurityIncident, error)

	// UpdateIncident updates a security incident
	UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error

	// ListIncidents lists security incidents with optional filtering
	ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error)
}

// ModelsVulnerabilityStore defines the interface for storing and retrieving security vulnerabilities using models
type ModelsVulnerabilityStore interface {
	// CreateVulnerability creates a new security vulnerability
	CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error

	// GetVulnerabilityByID retrieves a security vulnerability by ID
	GetVulnerabilityByID(ctx context.Context, vulnerabilityID string) (*models.Vulnerability, error)

	// UpdateVulnerability updates a security vulnerability
	UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error

	// ListVulnerabilities lists security vulnerabilities with optional filtering
	ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error)
}

// ModelsAuditLogger defines the interface for logging audit events using models
type ModelsAuditLogger interface {
	// LogAudit logs an audit event
	LogAudit(ctx context.Context, log *models.AuditLog) error

	// GetAuditLogs gets audit logs
	GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error)
}

// NewSecurityManager creates a new security manager
func NewSecurityManagerImpl(incidentStore ModelsIncidentStore, vulnStore ModelsVulnerabilityStore, auditLogger ModelsAuditLogger) *SecurityManagerImpl {
	return &SecurityManagerImpl{
		incidentStore: incidentStore,
		vulnStore:     vulnStore,
		auditLogger:   auditLogger,
	}
}

// Initialize initializes the security manager
func (m *SecurityManagerImpl) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = true
	return nil
}

// ReportIncident reports a security incident
func (m *SecurityManagerImpl) ReportIncident(ctx context.Context, title, description string, severity models.SecurityIncidentSeverity) (*models.SecurityIncident, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident := &models.SecurityIncident{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      models.SecurityIncidentStatusOpen,
		ReportedAt:  time.Now(),
	}

	err := m.incidentStore.CreateIncident(ctx, incident)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionCreate,
		Resource:    "security_incident",
		ResourceID:  incident.ID,
		Description: "Created security incident: " + incident.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return incident, nil
}

// GetIncident retrieves a security incident by ID
func (m *SecurityManagerImpl) GetIncident(ctx context.Context, incidentID string) (*models.SecurityIncident, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incident, err := m.incidentStore.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "security_incident",
		ResourceID:  incidentID,
		Description: "Retrieved security incident: " + incident.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return incident, nil
}

// UpdateIncident updates a security incident
func (m *SecurityManagerImpl) UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the existing incident to compare changes
	existingIncident, err := m.incidentStore.GetIncidentByID(ctx, incident.ID)
	if err != nil {
		return err
	}

	// Update the incident
	err = m.incidentStore.UpdateIncident(ctx, incident)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionUpdate,
		Resource:    "security_incident",
		ResourceID:  incident.ID,
		Description: "Updated security incident: " + incident.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// CloseIncident closes a security incident
func (m *SecurityManagerImpl) CloseIncident(ctx context.Context, incidentID, resolution string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the existing incident
	incident, err := m.incidentStore.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return err
	}

	// Update the incident
	incident.Status = models.SecurityIncidentStatusClosed
	incident.Resolution = resolution
	incident.ResolvedAt = time.Now()
	incident.ResolvedBy = getSecurityManagerUserIDFromContext(ctx)

	err = m.incidentStore.UpdateIncident(ctx, incident)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionUpdate,
		Resource:    "security_incident",
		ResourceID:  incidentID,
		Description: "Closed security incident: " + incident.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// ListIncidents lists security incidents
func (m *SecurityManagerImpl) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incidents, total, err := m.incidentStore.ListIncidents(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "security_incident",
		Description: "Listed security incidents",
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return incidents, total, nil
}

// ReportVulnerability reports a security vulnerability
func (m *SecurityManagerImpl) ReportVulnerability(ctx context.Context, title, description string, severity models.VulnerabilitySeverity) (*models.Vulnerability, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vulnerability := &models.Vulnerability{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      models.VulnerabilityStatusOpen,
		ReportedAt:  time.Now(),
		ReportedBy:  getSecurityManagerUserIDFromContext(ctx),
	}

	err := m.vulnStore.CreateVulnerability(ctx, vulnerability)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionCreate,
		Resource:    "vulnerability",
		ResourceID:  vulnerability.ID,
		Description: "Reported vulnerability: " + vulnerability.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return vulnerability, nil
}

// GetVulnerability retrieves a security vulnerability by ID
func (m *SecurityManagerImpl) GetVulnerability(ctx context.Context, vulnerabilityID string) (*models.Vulnerability, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vulnerability, err := m.vulnStore.GetVulnerabilityByID(ctx, vulnerabilityID)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "vulnerability",
		ResourceID:  vulnerabilityID,
		Description: "Retrieved vulnerability: " + vulnerability.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return vulnerability, nil
}

// UpdateVulnerability updates a security vulnerability
func (m *SecurityManagerImpl) UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the existing vulnerability to compare changes
	existingVulnerability, err := m.vulnStore.GetVulnerabilityByID(ctx, vulnerability.ID)
	if err != nil {
		return err
	}

	// Update the vulnerability
	err = m.vulnStore.UpdateVulnerability(ctx, vulnerability)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionUpdate,
		Resource:    "vulnerability",
		ResourceID:  vulnerability.ID,
		Description: "Updated vulnerability: " + vulnerability.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// MitigateVulnerability marks a security vulnerability as mitigated
func (m *SecurityManagerImpl) MitigateVulnerability(ctx context.Context, vulnerabilityID, mitigation string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the existing vulnerability
	vulnerability, err := m.vulnStore.GetVulnerabilityByID(ctx, vulnerabilityID)
	if err != nil {
		return err
	}

	// Update the vulnerability
	vulnerability.Status = models.VulnerabilityStatusMitigated
	vulnerability.Mitigation = mitigation
	vulnerability.MitigatedAt = time.Now()
	vulnerability.MitigatedBy = getSecurityManagerUserIDFromContext(ctx)

	err = m.vulnStore.UpdateVulnerability(ctx, vulnerability)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionUpdate,
		Resource:    "vulnerability",
		ResourceID:  vulnerabilityID,
		Description: "Mitigated vulnerability: " + vulnerability.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// ResolveVulnerability marks a security vulnerability as resolved
func (m *SecurityManagerImpl) ResolveVulnerability(ctx context.Context, vulnerabilityID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the existing vulnerability
	vulnerability, err := m.vulnStore.GetVulnerabilityByID(ctx, vulnerabilityID)
	if err != nil {
		return err
	}

	// Update the vulnerability
	vulnerability.Status = models.VulnerabilityStatusResolved
	vulnerability.ResolvedAt = time.Now()
	vulnerability.ResolvedBy = getSecurityManagerUserIDFromContext(ctx)

	err = m.vulnStore.UpdateVulnerability(ctx, vulnerability)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionUpdate,
		Resource:    "vulnerability",
		ResourceID:  vulnerabilityID,
		Description: "Resolved vulnerability: " + vulnerability.Title,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// ListVulnerabilities lists security vulnerabilities
func (m *SecurityManagerImpl) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vulnerabilities, total, err := m.vulnStore.ListVulnerabilities(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getSecurityManagerUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "vulnerability",
		Description: "Listed vulnerabilities",
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return vulnerabilities, total, nil
}

// Close closes the security manager
func (m *SecurityManagerImpl) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = false
	return nil
}

// Helper function to get user ID from context
func getSecurityManagerUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return "system"
	}
	
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "system"
	}
	
	return userID
}
