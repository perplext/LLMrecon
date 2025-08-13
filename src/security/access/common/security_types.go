// Package common provides common constants and utilities for the security access control system
package common

import (
)

// AuthMethod represents an authentication method
type AuthMethod string

// Constants for authentication methods
const (
	AuthMethodPassword    AuthMethod = "password"
	AuthMethodOAuth       AuthMethod = "oauth"
	AuthMethodAPIKey      AuthMethod = "api_key"
	AuthMethodTOTP        AuthMethod = "totp"
	AuthMethodBackupCode  AuthMethod = "backup_code"
	AuthMethodWebAuthn    AuthMethod = "webauthn"
	AuthMethodSMS         AuthMethod = "sms"
	AuthMethodCertificate AuthMethod = "certificate"
	AuthMethodSSO         AuthMethod = "sso"
)

// SecurityIncidentStatus defines the status of a security incident
type SecurityIncidentStatus string

// VulnerabilitySeverity defines the severity level of a vulnerability
type VulnerabilitySeverity int

// VulnerabilityStatus defines the status of a vulnerability
type VulnerabilityStatus string

// Security incident status constants
const (
	SecurityIncidentStatusOpen    SecurityIncidentStatus = "open"
	SecurityIncidentStatusClosed  SecurityIncidentStatus = "closed"
	SecurityIncidentStatusPending SecurityIncidentStatus = "pending"
)

// Vulnerability severity constants
const (
	VulnerabilitySeverityLow VulnerabilitySeverity = iota
	VulnerabilitySeverityMedium
	VulnerabilitySeverityHigh
	VulnerabilitySeverityCritical
)

// Vulnerability status constants
const (
	VulnerabilityStatusOpen      VulnerabilityStatus = "open"
	VulnerabilityStatusMitigated VulnerabilityStatus = "mitigated"
	VulnerabilityStatusResolved  VulnerabilityStatus = "resolved"
	VulnerabilityStatusPending   VulnerabilityStatus = "pending"
)

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Type            string                 `json:"type"`
	Severity        string                 `json:"severity"`
	Status          SecurityIncidentStatus `json:"status"`
	Description     string                 `json:"description"`
	DetectedAt      time.Time              `json:"detected_at"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
	CreatedBy       string                 `json:"created_by"`
	ReportedBy      string                 `json:"reported_by"`
	AssignedTo      string                 `json:"assigned_to,omitempty"`
	Details         map[string]interface{} `json:"details,omitempty"`
	Actions         []string               `json:"actions,omitempty"`
	Resolution      string                 `json:"resolution,omitempty"`
	AffectedSystems []string               `json:"affected_systems,omitempty"`
	RelatedLogs     []string               `json:"related_logs,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// User represents a user in the system (same as models.User but in common package)
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

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID              string                `json:"id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Severity        VulnerabilitySeverity `json:"severity"`
	Status          VulnerabilityStatus   `json:"status"`
	DiscoveredAt    time.Time             `json:"discovered_at"`
	DiscoveredBy    string                `json:"discovered_by"`
	MitigatedAt     *time.Time            `json:"mitigated_at,omitempty"`
	Mitigation      string                `json:"mitigation,omitempty"`
	ResolvedAt      *time.Time            `json:"resolved_at,omitempty"`
	AffectedSystems []string              `json:"affected_systems,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AuditSeverityFromVulnerabilitySeverity converts a vulnerability severity to an audit severity
func AuditSeverityFromVulnerabilitySeverity(severity VulnerabilitySeverity) AuditSeverity {
	switch severity {
	case VulnerabilitySeverityLow:
		return AuditSeverityInfo
	case VulnerabilitySeverityMedium:
		return AuditSeverityWarning
	case VulnerabilitySeverityHigh:
		return AuditSeverityError
	case VulnerabilitySeverityCritical:
		return AuditSeverityCritical
	default:
		return AuditSeverityInfo
	}
}
