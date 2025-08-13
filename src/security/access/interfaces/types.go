// Package interfaces defines the interfaces for the access control system
package interfaces

// Use common.AuthMethod instead of redefining it here

// AuditAction represents an action recorded in the audit log
type AuditAction string

// Constants for audit actions
const (
	AuditActionLogin           AuditAction = "login"
	AuditActionLogout          AuditAction = "logout"
	AuditActionCreate          AuditAction = "create"
	AuditActionRead            AuditAction = "read"
	AuditActionUpdate          AuditAction = "update"
	AuditActionDelete          AuditAction = "delete"
	AuditActionExecute         AuditAction = "execute"
	AuditActionApprove         AuditAction = "approve"
	AuditActionReject          AuditAction = "reject"
	AuditActionConfigChange    AuditAction = "config_change"
	AuditActionPermissionChange AuditAction = "permission_change"
	AuditActionRoleChange      AuditAction = "role_change"
	AuditActionSecurityAlert   AuditAction = "security_alert"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

// Constants for audit severity
const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityLow      AuditSeverity = "low"
	AuditSeverityMedium   AuditSeverity = "medium"
	AuditSeverityHigh     AuditSeverity = "high"
	AuditSeverityCritical AuditSeverity = "critical"
)

// SecurityIncidentStatus represents the status of a security incident
type SecurityIncidentStatus string

// Constants for security incident status
const (
	IncidentStatusOpen       SecurityIncidentStatus = "open"
	IncidentStatusInProgress SecurityIncidentStatus = "in_progress"
	IncidentStatusContained  SecurityIncidentStatus = "contained"
	IncidentStatusResolved   SecurityIncidentStatus = "resolved"
	IncidentStatusClosed     SecurityIncidentStatus = "closed"
)

// VulnerabilityStatus represents the status of a vulnerability
type VulnerabilityStatus string

// Constants for vulnerability status
const (
	VulnerabilityStatusOpen       VulnerabilityStatus = "open"
	VulnerabilityStatusInProgress VulnerabilityStatus = "in_progress"
	VulnerabilityStatusMitigated  VulnerabilityStatus = "mitigated"
	VulnerabilityStatusResolved   VulnerabilityStatus = "resolved"
	VulnerabilityStatusClosed     VulnerabilityStatus = "closed"
	VulnerabilityStatusDeferred   VulnerabilityStatus = "deferred"
)

// VulnerabilitySeverity represents the severity of a vulnerability
type VulnerabilitySeverity string

// Constants for vulnerability severity
const (
	VulnerabilitySeverityLow      VulnerabilitySeverity = "low"
	VulnerabilitySeverityMedium   VulnerabilitySeverity = "medium"
	VulnerabilitySeverityHigh     VulnerabilitySeverity = "high"
	VulnerabilitySeverityCritical VulnerabilitySeverity = "critical"
)
