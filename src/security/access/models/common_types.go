// Package models provides common data models for the security system
package models

import (
)

// SecurityIncidentSeverity represents the severity of a security incident
type SecurityIncidentSeverity string

// SecurityIncidentStatus represents the status of a security incident
type SecurityIncidentStatus string

// VulnerabilitySeverity represents the severity of a vulnerability
type VulnerabilitySeverity string

// VulnerabilityStatus represents the status of a vulnerability
type VulnerabilityStatus string

// AuditAction represents an action that can be audited
type AuditAction string

// String returns the string representation of the AuditAction
func (a AuditAction) String() string {
	return string(a)
}

// Constants for security incident severity
const (
	SecurityIncidentSeverityLow      SecurityIncidentSeverity = "low"
	SecurityIncidentSeverityMedium   SecurityIncidentSeverity = "medium"
	SecurityIncidentSeverityHigh     SecurityIncidentSeverity = "high"
	SecurityIncidentSeverityCritical SecurityIncidentSeverity = "critical"
)

// Constants for security incident status
const (
	SecurityIncidentStatusOpen      SecurityIncidentStatus = "open"
	SecurityIncidentStatusInProgress SecurityIncidentStatus = "in-progress"
	SecurityIncidentStatusResolved  SecurityIncidentStatus = "resolved"
	SecurityIncidentStatusClosed    SecurityIncidentStatus = "closed"
)

// Constants for vulnerability severity
const (
	VulnerabilitySeverityLow      VulnerabilitySeverity = "low"
	VulnerabilitySeverityMedium   VulnerabilitySeverity = "medium"
	VulnerabilitySeverityHigh     VulnerabilitySeverity = "high"
	VulnerabilitySeverityCritical VulnerabilitySeverity = "critical"
)

// Constants for vulnerability status
const (
	VulnerabilityStatusOpen      VulnerabilityStatus = "open"
	VulnerabilityStatusMitigated VulnerabilityStatus = "mitigated"
	VulnerabilityStatusResolved  VulnerabilityStatus = "resolved"
	VulnerabilityStatusClosed    VulnerabilityStatus = "closed"
)

// Constants for audit actions
const (
	AuditActionCreate AuditAction = "create"
	AuditActionRead   AuditAction = "read"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionLogin  AuditAction = "login"
	AuditActionLogout AuditAction = "logout"
)

// User represents a user in the system
type User struct {
	ID                 string                 `json:"id"`
	Username           string                 `json:"username"`
	Email              string                 `json:"email"`
	PasswordHash       string                 `json:"password_hash,omitempty"`
	Roles              []string               `json:"roles"`
	Permissions        []string               `json:"permissions,omitempty"`
	MFAEnabled         bool                   `json:"mfa_enabled"`
	MFAMethod          string                 `json:"mfa_method,omitempty"`
	MFAMethods         []string               `json:"mfa_methods,omitempty"`
	MFASecret          string                 `json:"mfa_secret,omitempty"`
	LastLogin          time.Time              `json:"last_login"`
	LastPasswordChange time.Time              `json:"last_password_change"`
	FailedLoginAttempts int                   `json:"failed_login_attempts"`
	Locked             bool                   `json:"locked"`
	Active             bool                   `json:"active"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// Session represents a user session
type Session struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Token           string                 `json:"token"`
	RefreshToken    string                 `json:"refresh_token,omitempty"`
	IPAddress       string                 `json:"ip_address"`
	UserAgent       string                 `json:"user_agent"`
	ExpiresAt       time.Time              `json:"expires_at"`
	LastActivity    time.Time              `json:"last_activity"`
	MFACompleted    bool                   `json:"mfa_completed"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Username     string                 `json:"username"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Status       string                 `json:"status"`
	Description  string                 `json:"description"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID              string                  `json:"id"`
	Title           string                  `json:"title"`
	Description     string                  `json:"description"`
	Severity        SecurityIncidentSeverity `json:"severity"`
	Status          SecurityIncidentStatus  `json:"status"`
	ReportedAt      time.Time               `json:"reported_at"`
	ReportedBy      string                  `json:"reported_by"`
	AssignedTo      string                  `json:"assigned_to,omitempty"`
	Resolution      string                  `json:"resolution,omitempty"`
	ResolvedAt      time.Time               `json:"resolved_at,omitempty"`
	ResolvedBy      string                  `json:"resolved_by,omitempty"`
	AffectedSystems []string                `json:"affected_systems,omitempty"`
	RelatedLogs     []string                `json:"related_logs,omitempty"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID              string                `json:"id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Severity        VulnerabilitySeverity `json:"severity"`
	Status          VulnerabilityStatus   `json:"status"`
	ReportedAt      time.Time             `json:"reported_at"`
	ReportedBy      string                `json:"reported_by"`
	AssignedTo      string                `json:"assigned_to,omitempty"`
	Mitigation      string                `json:"mitigation,omitempty"`
	MitigatedAt     time.Time             `json:"mitigated_at,omitempty"`
	MitigatedBy     string                `json:"mitigated_by,omitempty"`
	ResolvedAt      time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy      string                `json:"resolved_by,omitempty"`
	AffectedSystems []string              `json:"affected_systems,omitempty"`
	CVE             string                `json:"cve,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
