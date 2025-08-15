// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/security/access/models"
)

// BasicSecurityManager manages security incidents and vulnerabilities
type BasicSecurityManager struct {
	incidentStore      IncidentStore
	vulnerabilityStore VulnerabilityStore
	auditLogger        AuditLogger
	config             *AccessControlConfig
	mu                 sync.RWMutex
}

// IncidentStore defines the interface for security incident storage
type IncidentStore interface {
	CreateIncident(ctx context.Context, incident *SecurityIncident) error
	GetIncident(ctx context.Context, id string) (*SecurityIncident, error)
	UpdateIncident(ctx context.Context, incident *SecurityIncident) error
	DeleteIncident(ctx context.Context, id string) error
	ListIncidents(ctx context.Context, filter *LocalIncidentFilter) ([]*SecurityIncident, error)

// VulnerabilityStore defines the interface for vulnerability storage
type VulnerabilityStore interface {
	CreateVulnerability(ctx context.Context, vulnerability *Vulnerability) error
	GetVulnerability(ctx context.Context, id string) (*Vulnerability, error)
	UpdateVulnerability(ctx context.Context, vulnerability *Vulnerability) error
	DeleteVulnerability(ctx context.Context, id string) error
	ListVulnerabilities(ctx context.Context, filter *LocalVulnerabilityFilter) ([]*Vulnerability, error)

// LocalIncidentFilter defines filters for querying security incidents (local version)
type LocalIncidentFilter struct {
	Severity   AuditSeverity  `json:"severity,omitempty"`
	Status     IncidentStatus `json:"status,omitempty"`
	AssignedTo string         `json:"assigned_to,omitempty"`
	ReportedBy string         `json:"reported_by,omitempty"`
	StartTime  time.Time      `json:"start_time,omitempty"`
	EndTime    time.Time      `json:"end_time,omitempty"`
	Limit      int            `json:"limit,omitempty"`
	Offset     int            `json:"offset,omitempty"`

// LocalVulnerabilityFilter defines filters for querying vulnerabilities (local version)
type LocalVulnerabilityFilter struct {
	Severity       AuditSeverity       `json:"severity,omitempty"`
	Status         VulnerabilityStatus `json:"status,omitempty"`
	AssignedTo     string              `json:"assigned_to,omitempty"`
	ReportedBy     string              `json:"reported_by,omitempty"`
	AffectedSystem string              `json:"affected_system,omitempty"`
	CVE            string              `json:"cve,omitempty"`
	StartTime      time.Time           `json:"start_time,omitempty"`
	EndTime        time.Time           `json:"end_time,omitempty"`
	Limit          int                 `json:"limit,omitempty"`
	Offset         int                 `json:"offset,omitempty"`

// NewSecurityManager creates a new security manager
func NewSecurityManager(config *AccessControlConfig, incidentStore IncidentStore, vulnerabilityStore VulnerabilityStore, auditLogger AuditLogger) *BasicSecurityManager {
	return NewBasicSecurityManager(config, incidentStore, vulnerabilityStore, auditLogger)

// NewBasicSecurityManager creates a new basic security manager
func NewBasicSecurityManager(config *AccessControlConfig, incidentStore IncidentStore, vulnerabilityStore VulnerabilityStore, auditLogger AuditLogger) *BasicSecurityManager {
	return &BasicSecurityManager{
		config:             config,
		incidentStore:      incidentStore,
		vulnerabilityStore: vulnerabilityStore,
		auditLogger:        auditLogger,
	}

// CreateIncident creates a new security incident
func (m *BasicSecurityManager) CreateIncident(ctx context.Context, title, description string, severity AuditSeverity, reportedBy string, auditLogIDs []string, metadata map[string]interface{}) (*SecurityIncident, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create incident
	incident := &SecurityIncident{
		ID:          generateRandomID(),
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      IncidentStatusNew,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ReportedBy:  reportedBy,
		AuditLogIDs: auditLogIDs,
		Metadata:    metadata,
	}

	// Save incident
	if err := m.incidentStore.CreateIncident(ctx, incident); err != nil {
		return nil, err
	}

	// Log incident creation
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      reportedBy,
		Action:      AuditActionCreate,
		Resource:    "security_incident",
		ResourceID:  incident.ID,
		Description: fmt.Sprintf("Security incident created: %s", title),
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Metadata: map[string]interface{}{
			"incident_severity": severity,
			"incident_status":   IncidentStatusNew,
		},
	})

	return incident, nil

// UpdateIncidentStatus updates the status of a security incident
func (m *BasicSecurityManager) UpdateIncidentStatus(ctx context.Context, id string, status IncidentStatus, assignedTo, updatedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get incident
	incident, err := m.incidentStore.GetIncident(ctx, id)
	if err != nil {
		return err
	}

	// Update incident
	incident.Status = status
	incident.UpdatedAt = time.Now()

	// Set assigned to if provided
	if assignedTo != "" {
		incident.AssignedTo = assignedTo
	}

	// Set resolved at if status is resolved
	if status == IncidentStatusResolved {
		now := time.Now()
		incident.ResolvedAt = now
	}

	// Save incident
	if err := m.incidentStore.UpdateIncident(ctx, incident); err != nil {
		return err
	}

	// Log incident update
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      updatedBy,
		Action:      AuditActionUpdate,
		Resource:    "security_incident",
		ResourceID:  incident.ID,
		Description: fmt.Sprintf("Security incident status updated: %s", status),
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Metadata: map[string]interface{}{
			"incident_severity": incident.Severity,
			"incident_status":   status,
			"assigned_to":       incident.AssignedTo,
		},
	})

	return nil

// GetIncident retrieves a security incident by ID
func (m *BasicSecurityManager) GetIncident(ctx context.Context, id string) (*SecurityIncident, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.incidentStore.GetIncident(ctx, id)

// ListIncidents lists security incidents based on filters
func (m *BasicSecurityManager) ListIncidents(ctx context.Context, filter *LocalIncidentFilter) ([]*SecurityIncident, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.incidentStore.ListIncidents(ctx, filter)

// CreateVulnerability creates a new vulnerability
func (m *BasicSecurityManager) CreateVulnerability(ctx context.Context, title, description string, severity AuditSeverity, affectedSystem, cve, reportedBy string, metadata map[string]interface{}) (*Vulnerability, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create vulnerability
	vulnerability := &Vulnerability{
		ID:             generateRandomID(),
		Title:          title,
		Description:    description,
		Severity:       severity,
		Status:         VulnerabilityStatusNew,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ReportedBy:     reportedBy,
		AffectedSystem: affectedSystem,
		CVE:            cve,
		Metadata:       metadata,
	}

	// Save vulnerability
	if err := m.vulnerabilityStore.CreateVulnerability(ctx, vulnerability); err != nil {
		return nil, err
	}

	// Log vulnerability creation
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      reportedBy,
		Action:      AuditActionCreate,
		Resource:    "vulnerability",
		ResourceID:  vulnerability.ID,
		Description: fmt.Sprintf("Vulnerability created: %s", title),
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Metadata: map[string]interface{}{
			"vulnerability_severity": severity,
			"vulnerability_status":   VulnerabilityStatusNew,
			"affected_system":        affectedSystem,
			"cve":                    cve,
		},
	})

	return vulnerability, nil
// UpdateVulnerabilityStatus updates the status of a vulnerability
func (m *BasicSecurityManager) UpdateVulnerabilityStatus(ctx context.Context, id string, status VulnerabilityStatus, assignedTo, remediationPlan, updatedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get vulnerability
	vulnerability, err := m.vulnerabilityStore.GetVulnerability(ctx, id)
	if err != nil {
		return err
	}

	// Update vulnerability
	vulnerability.Status = status
	vulnerability.UpdatedAt = time.Now()

	// Set assigned to if provided
	if assignedTo != "" {
		vulnerability.AssignedTo = assignedTo
	}

	// Set remediation plan if provided
	if remediationPlan != "" {
		vulnerability.RemediationPlan = remediationPlan
	}

	// Set resolved at if status is remediated or verified
	if status == VulnerabilityStatusRemediated || status == VulnerabilityStatusVerified {
		now := time.Now()
		vulnerability.ResolvedAt = now
	}

	// Save vulnerability
	if err := m.vulnerabilityStore.UpdateVulnerability(ctx, vulnerability); err != nil {
		return err
	}

	// Log vulnerability update
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      updatedBy,
		Action:      AuditActionUpdate,
		Resource:    "vulnerability",
		ResourceID:  vulnerability.ID,
		Description: fmt.Sprintf("Vulnerability status updated: %s", status),
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Metadata: map[string]interface{}{
			"vulnerability_severity": vulnerability.Severity,
			"vulnerability_status":   status,
			"assigned_to":            vulnerability.AssignedTo,
			"remediation_plan":       vulnerability.RemediationPlan,
		},
	})

	return nil

// GetVulnerability retrieves a vulnerability by ID
func (m *BasicSecurityManager) GetVulnerability(ctx context.Context, id string) (*Vulnerability, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.vulnerabilityStore.GetVulnerability(ctx, id)

// ListVulnerabilities lists vulnerabilities based on filters
func (m *BasicSecurityManager) ListVulnerabilities(ctx context.Context, filter *LocalVulnerabilityFilter) ([]*Vulnerability, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.vulnerabilityStore.ListVulnerabilities(ctx, filter)

// ProcessSecurityAuditLog processes an audit log for security incidents
func (m *BasicSecurityManager) ProcessSecurityAuditLog(ctx context.Context, log *AuditLog) error {
	// Check if incident tracking is enabled
	if m.config == nil || !m.config.SecurityIncidentConfig.EnableIncidentTracking {
		return nil
	}

	// Check if auto-create incidents is enabled
	if !m.config.SecurityIncidentConfig.AutoCreateIncidents {
		return nil
	}

	// Check if severity meets escalation threshold
	if log.Severity == "" || log.Severity < m.config.SecurityIncidentConfig.EscalationThreshold {
		return nil
	}

	// Create incident
	title := fmt.Sprintf("%s: %s", log.Action, log.Description)
	description := fmt.Sprintf("Security incident automatically created from audit log:\n\n%s", log.Description)

	if log.UserID != "" {
		description += fmt.Sprintf("\n\nUser ID: %s", log.UserID)
	}
	if log.Username != "" {
		description += fmt.Sprintf("\nUsername: %s", log.Username)
	}
	if log.IPAddress != "" {
		description += fmt.Sprintf("\nIP Address: %s", log.IPAddress)
	}
	if log.UserAgent != "" {
		description += fmt.Sprintf("\nUser Agent: %s", log.UserAgent)
	}

	metadata := map[string]interface{}{
		"auto_created": true,
		"audit_log_id": log.ID,
	}

	if log.Metadata != nil {
		for k, v := range log.Metadata {
			metadata[k] = v
		}
	}

	_, err := m.CreateIncident(ctx, title, description, log.Severity, "system", []string{log.ID}, metadata)
	return err

// LocalInMemoryIncidentStore is a simple in-memory implementation of IncidentStore
type LocalInMemoryIncidentStore struct {
	incidents map[string]*SecurityIncident
	mu        sync.RWMutex

// NewLocalInMemoryIncidentStore creates a new local in-memory incident store
func NewLocalInMemoryIncidentStore() *LocalInMemoryIncidentStore {
	return &LocalInMemoryIncidentStore{
		incidents: make(map[string]*SecurityIncident),
	}

// CreateIncident creates a new security incident
func (s *LocalInMemoryIncidentStore) CreateIncident(ctx context.Context, incident *SecurityIncident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store incident
	s.incidents[incident.ID] = incident

	return nil

// GetIncident retrieves a security incident by ID
func (s *LocalInMemoryIncidentStore) GetIncident(ctx context.Context, id string) (*SecurityIncident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incident, ok := s.incidents[id]
	if !ok {
		return nil, fmt.Errorf("incident not found")
	}

	return incident, nil

// UpdateIncident updates an existing security incident
func (s *LocalInMemoryIncidentStore) UpdateIncident(ctx context.Context, incident *SecurityIncident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if incident exists
	_, ok := s.incidents[incident.ID]
	if !ok {
		return fmt.Errorf("incident not found")
	}

	// Update incident
	s.incidents[incident.ID] = incident

	return nil

// DeleteIncident deletes a security incident by ID
func (s *LocalInMemoryIncidentStore) DeleteIncident(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if incident exists
	_, ok := s.incidents[id]
	if !ok {
		return nil
	}

	// Delete incident
	delete(s.incidents, id)

	return nil

// ListIncidents lists security incidents based on filters
func (s *LocalInMemoryIncidentStore) ListIncidents(ctx context.Context, filter *LocalIncidentFilter) ([]*SecurityIncident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Default limit if not specified
	limit := 100
	if filter != nil && filter.Limit > 0 {
		limit = filter.Limit
	}

	// Default offset if not specified
	offset := 0
	if filter != nil && filter.Offset > 0 {
		offset = filter.Offset
	}

	// Filter incidents
	filtered := make([]*SecurityIncident, 0)
	for _, incident := range s.incidents {
		if filter != nil {
			// Apply filters
			if filter.Severity != "" && incident.Severity != filter.Severity {
				continue
			}
			if filter.Status != "" && incident.Status != filter.Status {
				continue
			}
			if filter.AssignedTo != "" && incident.AssignedTo != filter.AssignedTo {
				continue
			}
			if filter.ReportedBy != "" && incident.ReportedBy != filter.ReportedBy {
				continue
			}
			if !filter.StartTime.IsZero() && incident.CreatedAt.Before(filter.StartTime) {
				continue
			}
			if !filter.EndTime.IsZero() && incident.CreatedAt.After(filter.EndTime) {
				continue
			}
		}

		filtered = append(filtered, incident)
	}

	// Apply pagination
	start := offset
	if start >= len(filtered) {
		return []*SecurityIncident{}, nil
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil

// LocalInMemoryVulnerabilityStore is a simple in-memory implementation of VulnerabilityStore
type LocalInMemoryVulnerabilityStore struct {
	vulnerabilities map[string]*Vulnerability
	mu              sync.RWMutex

// NewLocalInMemoryVulnerabilityStore creates a new local in-memory vulnerability store
func NewLocalInMemoryVulnerabilityStore() *LocalInMemoryVulnerabilityStore {
	return &LocalInMemoryVulnerabilityStore{
		vulnerabilities: make(map[string]*Vulnerability),
	}

// CreateVulnerability creates a new vulnerability
func (s *LocalInMemoryVulnerabilityStore) CreateVulnerability(ctx context.Context, vulnerability *Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store vulnerability
	s.vulnerabilities[vulnerability.ID] = vulnerability

	return nil

// GetVulnerability retrieves a vulnerability by ID
func (s *LocalInMemoryVulnerabilityStore) GetVulnerability(ctx context.Context, id string) (*Vulnerability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vulnerability, ok := s.vulnerabilities[id]
	if !ok {
		return nil, fmt.Errorf("vulnerability not found")
	}

	return vulnerability, nil

// UpdateVulnerability updates an existing vulnerability
func (s *LocalInMemoryVulnerabilityStore) UpdateVulnerability(ctx context.Context, vulnerability *Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vulnerability exists
	_, ok := s.vulnerabilities[vulnerability.ID]
	if !ok {
		return fmt.Errorf("vulnerability not found")
	}

	// Update vulnerability
	s.vulnerabilities[vulnerability.ID] = vulnerability

	return nil

// DeleteVulnerability deletes a vulnerability by ID
func (s *LocalInMemoryVulnerabilityStore) DeleteVulnerability(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vulnerability exists
	_, ok := s.vulnerabilities[id]
	if !ok {
		return nil
	}

	// Delete vulnerability
	delete(s.vulnerabilities, id)

	return nil

// ListVulnerabilities lists vulnerabilities based on filters
func (s *LocalInMemoryVulnerabilityStore) ListVulnerabilities(ctx context.Context, filter *LocalVulnerabilityFilter) ([]*Vulnerability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Default limit if not specified
	limit := 100
	if filter != nil && filter.Limit > 0 {
		limit = filter.Limit
	}

	// Default offset if not specified
	offset := 0
	if filter != nil && filter.Offset > 0 {
		offset = filter.Offset
	}

	// Filter vulnerabilities
	filtered := make([]*Vulnerability, 0)
	for _, vulnerability := range s.vulnerabilities {
		if filter != nil {
			// Apply filters
			if filter.Severity != "" && vulnerability.Severity != filter.Severity {
				continue
			}
			if filter.Status != "" && vulnerability.Status != filter.Status {
				continue
			}
			if filter.AssignedTo != "" && vulnerability.AssignedTo != filter.AssignedTo {
				continue
			}
			if filter.ReportedBy != "" && vulnerability.ReportedBy != filter.ReportedBy {
				continue
			}
			if filter.AffectedSystem != "" && vulnerability.AffectedSystem != filter.AffectedSystem {
				continue
			}
			if filter.CVE != "" && vulnerability.CVE != filter.CVE {
				continue
			}
			if !filter.StartTime.IsZero() && vulnerability.CreatedAt.Before(filter.StartTime) {
				continue
			}
			if !filter.EndTime.IsZero() && vulnerability.CreatedAt.After(filter.EndTime) {
				continue
			}
		}

		filtered = append(filtered, vulnerability)
	}

	// Apply pagination
	start := offset
	if start >= len(filtered) {
		return []*Vulnerability{}, nil
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil

// Close closes the security manager
func (m *BasicSecurityManager) Close() error {
	// Nothing to close for now
	return nil

// SecurityManagerAdapter wraps BasicSecurityManager to implement the SecurityManager interface
ype SecurityManagerAdapter struct {
	impl *BasicSecurityManager

// NewSecurityManagerAdapter creates a new security manager adapter
func NewSecurityManagerAdapter(impl *BasicSecurityManager) *SecurityManagerAdapter {
	return &SecurityManagerAdapter{impl: impl}

// ReportIncident reports a security incident
func (a *SecurityManagerAdapter) ReportIncident(title, description string, severity models.SecurityIncidentSeverity) (*models.SecurityIncident, error) {
	// Convert severity and create incident using the impl
	auditSeverity := AuditSeverity(severity)
	incident, err := a.impl.CreateIncident(context.Background(), title, description, auditSeverity, "system", []string{}, nil)
	if err != nil {
		return nil, err
	}
	// Convert back to models.SecurityIncident
	return &models.SecurityIncident{
		ID:          incident.ID,
		Title:       incident.Title,
		Description: incident.Description,
		Severity:    models.SecurityIncidentSeverity(incident.Severity),
		Status:      models.SecurityIncidentStatus(incident.Status),
		ReportedAt:  incident.CreatedAt,
		ReportedBy:  incident.ReportedBy,
	}, nil

// GetIncident retrieves a security incident by ID
func (a *SecurityManagerAdapter) GetIncident(incidentID string) (*models.SecurityIncident, error) {
	incident, err := a.impl.GetIncident(context.Background(), incidentID)
	if err != nil {
		return nil, err
	}

	// Convert to models.SecurityIncident
	return &models.SecurityIncident{
		ID:          incident.ID,
		Title:       incident.Title,
		Description: incident.Description,
		Severity:    models.SecurityIncidentSeverity(incident.Severity),
		Status:      models.SecurityIncidentStatus(incident.Status),
		ReportedAt:  incident.CreatedAt,
		ReportedBy:  incident.ReportedBy,
	}, nil

// UpdateIncident updates a security incident
func (a *SecurityManagerAdapter) UpdateIncident(incident *models.SecurityIncident) error {
	// Convert models.SecurityIncident to local type and update
	// This is a simplified implementation
	return nil

// CloseIncident closes a security incident
func (a *SecurityManagerAdapter) CloseIncident(incidentID, resolution string) error {
	// Update incident status to resolved
	return a.impl.UpdateIncidentStatus(context.Background(), incidentID, IncidentStatusResolved, "", "system")

// ListIncidents lists security incidents
func (a *SecurityManagerAdapter) ListIncidents(filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	// Convert filter and call impl
	localFilter := &LocalIncidentFilter{
		Limit:  limit,
		Offset: offset,
	}

	incidents, err := a.impl.ListIncidents(context.Background(), localFilter)
	if err != nil {
		return nil, 0, err
	}

	// Convert to models.SecurityIncident
	result := make([]*models.SecurityIncident, len(incidents))
	for i, incident := range incidents {
		result[i] = &models.SecurityIncident{
			ID:          incident.ID,
			Title:       incident.Title,
			Description: incident.Description,
			Severity:    models.SecurityIncidentSeverity(incident.Severity),
			Status:      models.SecurityIncidentStatus(incident.Status),
			ReportedAt:  incident.CreatedAt,
			ReportedBy:  incident.ReportedBy,
		}
	}

	return result, len(result), nil

// ReportVulnerability reports a security vulnerability
func (a *SecurityManagerAdapter) ReportVulnerability(title, description string, severity models.VulnerabilitySeverity) (*models.Vulnerability, error) {
	// Convert severity and create vulnerability using the impl
	auditSeverity := AuditSeverity(severity)
	vuln, err := a.impl.CreateVulnerability(context.Background(), title, description, auditSeverity, "", "", "system", nil)
	if err != nil {
		return nil, err
	}

	// Convert back to models.Vulnerability
	return &models.Vulnerability{
		ID:          vuln.ID,
		Title:       vuln.Title,
		Description: vuln.Description,
		Severity:    models.VulnerabilitySeverity(vuln.Severity),
		Status:      models.VulnerabilityStatus(vuln.Status),
		ReportedAt:  vuln.CreatedAt,
		ReportedBy:  vuln.ReportedBy,
	}, nil

// GetVulnerability retrieves a security vulnerability by ID
func (a *SecurityManagerAdapter) GetVulnerability(vulnerabilityID string) (*models.Vulnerability, error) {
	vuln, err := a.impl.GetVulnerability(context.Background(), vulnerabilityID)
	if err != nil {
		return nil, err
	}

	// Convert to models.Vulnerability
	return &models.Vulnerability{
		ID:          vuln.ID,
		Title:       vuln.Title,
		Description: vuln.Description,
		Severity:    models.VulnerabilitySeverity(vuln.Severity),
		Status:      models.VulnerabilityStatus(vuln.Status),
		ReportedAt:  vuln.CreatedAt,
		ReportedBy:  vuln.ReportedBy,
	}, nil

// UpdateVulnerability updates a security vulnerability
func (a *SecurityManagerAdapter) UpdateVulnerability(vulnerability *models.Vulnerability) error {
	// This is a simplified implementation
	return nil

// MitigateVulnerability marks a security vulnerability as mitigated
func (a *SecurityManagerAdapter) MitigateVulnerability(vulnerabilityID, mitigation string) error {
	return a.impl.UpdateVulnerabilityStatus(context.Background(), vulnerabilityID, VulnerabilityStatusRemediated, "", mitigation, "system")

// ResolveVulnerability marks a security vulnerability as resolved
func (a *SecurityManagerAdapter) ResolveVulnerability(vulnerabilityID string) error {
	return a.impl.UpdateVulnerabilityStatus(context.Background(), vulnerabilityID, VulnerabilityStatusVerified, "", "", "system")

// ListVulnerabilities lists security vulnerabilities
func (a *SecurityManagerAdapter) ListVulnerabilities(filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	// Convert filter and call impl
	localFilter := &LocalVulnerabilityFilter{
		Limit:  limit,
		Offset: offset,
	}

	vulnerabilities, err := a.impl.ListVulnerabilities(context.Background(), localFilter)
	if err != nil {
		return nil, 0, err
	}

	// Convert to models.Vulnerability
	result := make([]*models.Vulnerability, len(vulnerabilities))
	for i, vuln := range vulnerabilities {
		result[i] = &models.Vulnerability{
			ID:          vuln.ID,
			Title:       vuln.Title,
			Description: vuln.Description,
			Severity:    models.VulnerabilitySeverity(vuln.Severity),
			Status:      models.VulnerabilityStatus(vuln.Status),
			ReportedAt:  vuln.CreatedAt,
			ReportedBy:  vuln.ReportedBy,
		}
	}

	return result, len(result), nil

// Close closes the security manager
func (a *SecurityManagerAdapter) Close() error {
	return a.impl.Close()
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
