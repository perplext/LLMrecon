// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// InMemoryIncidentStoreAdapter is an in-memory implementation of interfaces.IncidentStore
type InMemoryIncidentStoreAdapter struct {
	incidents map[string]*models.SecurityIncident
	mu        sync.RWMutex

// NewInMemoryIncidentStoreAdapter creates a new in-memory incident store adapter
func NewInMemoryIncidentStoreAdapter() interfaces.IncidentStore {
	return &InMemoryIncidentStoreAdapter{
		incidents: make(map[string]*models.SecurityIncident),
	}

// CreateIncident creates a new security incident
func (s *InMemoryIncidentStoreAdapter) CreateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the incident already exists
	if _, exists := s.incidents[incident.ID]; exists {
		return fmt.Errorf("incident already exists: %s", incident.ID)
	}

	// Store the incident
	s.incidents[incident.ID] = incident

	return nil

// GetIncidentByID retrieves a security incident by ID
func (s *InMemoryIncidentStoreAdapter) GetIncidentByID(ctx context.Context, id string) (*models.SecurityIncident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incident, exists := s.incidents[id]
	if !exists {
		return nil, fmt.Errorf("incident not found: %s", id)
	}

	return incident, nil

// UpdateIncident updates a security incident
func (s *InMemoryIncidentStoreAdapter) UpdateIncident(ctx context.Context, incident *models.SecurityIncident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the incident exists
	if _, exists := s.incidents[incident.ID]; !exists {
		return fmt.Errorf("incident not found: %s", incident.ID)
	}

	// Update the incident
	s.incidents[incident.ID] = incident

	return nil

// DeleteIncident deletes a security incident
func (s *InMemoryIncidentStoreAdapter) DeleteIncident(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the incident exists
	if _, exists := s.incidents[id]; !exists {
		return fmt.Errorf("incident not found: %s", id)
	}

	// Delete the incident
	delete(s.incidents, id)

	return nil

// ListIncidents lists security incidents with optional filtering
func (s *InMemoryIncidentStoreAdapter) ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*models.SecurityIncident

	// Apply filters
	for _, incident := range s.incidents {
		if filter != nil {
			// Filter by severity
			if severity, ok := filter["severity"].(string); ok && severity != "" && string(incident.Severity) != severity {
				continue
			}

			// Filter by status
			if status, ok := filter["status"].(string); ok && status != "" && string(incident.Status) != status {
				continue
			}

			// Filter by assigned to
			if assignedTo, ok := filter["assigned_to"].(string); ok && assignedTo != "" && incident.AssignedTo != assignedTo {
				continue
			}

			// Filter by reported by
			if reportedBy, ok := filter["reported_by"].(string); ok && reportedBy != "" && incident.ReportedBy != reportedBy {
				continue
			}

			// Filter by time range
			if startTime, ok := filter["start_time"].(time.Time); ok && !startTime.IsZero() && incident.ReportedAt.Before(startTime) {
				continue
			}

			if endTime, ok := filter["end_time"].(time.Time); ok && !endTime.IsZero() && incident.ReportedAt.After(endTime) {
				continue
			}
		}

		results = append(results, incident)
	}

	// Get total count before applying offset and limit
	totalCount := len(results)

	// Apply offset and limit
	if offset > 0 && offset < len(results) {
		results = results[offset:]
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, totalCount, nil

// Close closes the incident store
func (s *InMemoryIncidentStoreAdapter) Close() error {
	return nil

// InMemoryVulnerabilityStoreAdapter is an in-memory implementation of interfaces.VulnerabilityStore
type InMemoryVulnerabilityStoreAdapter struct {
	vulnerabilities map[string]*models.Vulnerability
	mu              sync.RWMutex

// NewInMemoryVulnerabilityStoreAdapter creates a new in-memory vulnerability store adapter
func NewInMemoryVulnerabilityStoreAdapter() interfaces.VulnerabilityStore {
	return &InMemoryVulnerabilityStoreAdapter{
		vulnerabilities: make(map[string]*models.Vulnerability),
	}

// CreateVulnerability creates a new vulnerability
func (s *InMemoryVulnerabilityStoreAdapter) CreateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the vulnerability already exists
	if _, exists := s.vulnerabilities[vulnerability.ID]; exists {
		return fmt.Errorf("vulnerability already exists: %s", vulnerability.ID)
	}

	// Store the vulnerability
	s.vulnerabilities[vulnerability.ID] = vulnerability

	return nil

// GetVulnerabilityByID retrieves a vulnerability by ID
func (s *InMemoryVulnerabilityStoreAdapter) GetVulnerabilityByID(ctx context.Context, id string) (*models.Vulnerability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vulnerability, exists := s.vulnerabilities[id]
	if !exists {
		return nil, fmt.Errorf("vulnerability not found: %s", id)
	}

	return vulnerability, nil

// GetVulnerabilityByCVE retrieves a vulnerability by CVE ID
func (s *InMemoryVulnerabilityStoreAdapter) GetVulnerabilityByCVE(ctx context.Context, cve string) (*models.Vulnerability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, vulnerability := range s.vulnerabilities {
		if vulnerability.CVE == cve {
			return vulnerability, nil
		}
	}

	return nil, fmt.Errorf("vulnerability not found for CVE: %s", cve)

// UpdateVulnerability updates a vulnerability
func (s *InMemoryVulnerabilityStoreAdapter) UpdateVulnerability(ctx context.Context, vulnerability *models.Vulnerability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the vulnerability exists
	if _, exists := s.vulnerabilities[vulnerability.ID]; !exists {
		return fmt.Errorf("vulnerability not found: %s", vulnerability.ID)
	}

	// Update the vulnerability
	s.vulnerabilities[vulnerability.ID] = vulnerability

	return nil

// DeleteVulnerability deletes a vulnerability
func (s *InMemoryVulnerabilityStoreAdapter) DeleteVulnerability(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the vulnerability exists
	if _, exists := s.vulnerabilities[id]; !exists {
		return fmt.Errorf("vulnerability not found: %s", id)
	}

	// Delete the vulnerability
	delete(s.vulnerabilities, id)

	return nil

// ListVulnerabilities lists vulnerabilities with optional filtering
func (s *InMemoryVulnerabilityStoreAdapter) ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*models.Vulnerability

	// Apply filters
	for _, vulnerability := range s.vulnerabilities {
		if filter != nil {
			// Filter by severity
			if severity, ok := filter["severity"].(string); ok && severity != "" && string(vulnerability.Severity) != severity {
				continue
			}

			// Filter by status
			if status, ok := filter["status"].(string); ok && status != "" && string(vulnerability.Status) != status {
				continue
			}

			// Filter by assigned to
			if assignedTo, ok := filter["assigned_to"].(string); ok && assignedTo != "" && vulnerability.AssignedTo != assignedTo {
				continue
			}

			// Filter by reported by
			if reportedBy, ok := filter["reported_by"].(string); ok && reportedBy != "" && vulnerability.ReportedBy != reportedBy {
				continue
			}

			// Filter by affected system
			if affectedSystem, ok := filter["affected_system"].(string); ok && affectedSystem != "" {
				found := false
				for _, system := range vulnerability.AffectedSystems {
					if system == affectedSystem {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Filter by CVE
			if cve, ok := filter["cve"].(string); ok && cve != "" && vulnerability.CVE != cve {
				continue
			}

			// Filter by time range
			if startTime, ok := filter["start_time"].(time.Time); ok && !startTime.IsZero() && vulnerability.ReportedAt.Before(startTime) {
				continue
			}

			if endTime, ok := filter["end_time"].(time.Time); ok && !endTime.IsZero() && vulnerability.ReportedAt.After(endTime) {
				continue
			}
		}

		results = append(results, vulnerability)
	}

	// Get total count before applying offset and limit
	totalCount := len(results)

	// Apply offset and limit
	if offset > 0 && offset < len(results) {
		results = results[offset:]
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, totalCount, nil

// Close closes the vulnerability store
func (s *InMemoryVulnerabilityStoreAdapter) Close() error {
	return nil

// InMemoryAuditLoggerAdapter is an in-memory implementation of interfaces.AuditLogger
type InMemoryAuditLoggerAdapter struct {
	logs map[string]*models.AuditLog
	mu   sync.RWMutex

// NewInMemoryAuditLoggerAdapter creates a new in-memory audit logger adapter
func NewInMemoryAuditLoggerAdapter() interfaces.AuditLogger {
	return &InMemoryAuditLoggerAdapter{
		logs: make(map[string]*models.AuditLog),
	}

// LogAudit logs an audit event
func (l *InMemoryAuditLoggerAdapter) LogAudit(ctx context.Context, log *models.AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Generate ID if not provided
	if log.ID == "" {
		id, err := generateInMemoryID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
		log.ID = id
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// Store the log
	l.logs[log.ID] = log

	return nil

// GetAuditLogByID retrieves an audit log by ID
func (l *InMemoryAuditLoggerAdapter) GetAuditLogByID(ctx context.Context, id string) (*models.AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	log, exists := l.logs[id]
	if !exists {
		return nil, fmt.Errorf("audit log not found: %s", id)
	}

	return log, nil

// GetEventByID retrieves an audit event by ID (alias for GetAuditLogByID to satisfy interface)
func (l *InMemoryAuditLoggerAdapter) GetEventByID(ctx context.Context, id string) (*models.AuditLog, error) {
	return l.GetAuditLogByID(ctx, id)

// LogEvent logs an audit event (alias for LogAudit to satisfy interface)
func (l *InMemoryAuditLoggerAdapter) LogEvent(ctx context.Context, event *models.AuditLog) error {
	return l.LogAudit(ctx, event)

// QueryEvents queries audit events with filtering (alias for ListAuditLogs to satisfy interface)
func (l *InMemoryAuditLoggerAdapter) QueryEvents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	return l.ListAuditLogs(ctx, filter, offset, limit)

// ListAuditLogs lists audit logs with optional filtering
func (l *InMemoryAuditLoggerAdapter) ListAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var results []*models.AuditLog

	// Apply filters
	for _, log := range l.logs {
		if filter != nil {
			// Filter by user ID
			if userID, ok := filter["user_id"].(string); ok && userID != "" && log.UserID != userID {
				continue
			}

			// Filter by action
			if action, ok := filter["action"].(string); ok && action != "" && string(log.Action) != action {
				continue
			}

			// Filter by resource
			if resource, ok := filter["resource"].(string); ok && resource != "" && log.Resource != resource {
				continue
			}

			// Filter by resource ID
			if resourceID, ok := filter["resource_id"].(string); ok && resourceID != "" && log.ResourceID != resourceID {
				continue
			}

			// Note: AuditLog in models doesn't have Severity field
			// Skip severity filtering for now

			// Filter by time range
			if startTime, ok := filter["start_time"].(time.Time); ok && !startTime.IsZero() && log.Timestamp.Before(startTime) {
				continue
			}

			if endTime, ok := filter["end_time"].(time.Time); ok && !endTime.IsZero() && log.Timestamp.After(endTime) {
				continue
			}
		}

		results = append(results, log)
	}

	// Get total count before applying offset and limit
	totalCount := len(results)

	// Apply offset and limit
	if offset > 0 && offset < len(results) {
		results = results[offset:]
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, totalCount, nil

// ExportEvents exports audit events to a file
func (l *InMemoryAuditLoggerAdapter) ExportEvents(ctx context.Context, filter map[string]interface{}, format string) (string, error) {
	// For now, return a simple implementation
	return fmt.Sprintf("exported_%d_events.%s", len(l.logs), format), nil

// Close closes the audit logger
func (l *InMemoryAuditLoggerAdapter) Close() error {
	return nil

// Helper function to generate ID
func generateInMemoryID() (string, error) {
	// Simple ID generation - in real implementation, use UUID or similar
