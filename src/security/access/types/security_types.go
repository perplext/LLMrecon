// Package types defines common types for the security access control system
package types

import (
	"time"

	"github.com/perplext/LLMrecon/src/security/access/common"
)

// SecurityConfig defines configuration for the security subsystem
type SecurityConfig struct {
	// IncidentEscalationThreshold is the minimum severity level for incident escalation
	IncidentEscalationThreshold common.AuditSeverity `json:"incidentEscalationThreshold"`

	// VulnerabilityEscalationThreshold is the minimum severity level for vulnerability escalation
	VulnerabilityEscalationThreshold common.VulnerabilitySeverity `json:"vulnerabilityEscalationThreshold"`

	// AuditLogRetentionDays is the number of days to retain audit logs
	AuditLogRetentionDays int `json:"auditLogRetentionDays"`

	// EnableRealTimeAlerts enables real-time alerts for high-severity incidents
	EnableRealTimeAlerts bool `json:"enableRealTimeAlerts"`

	// AlertEndpoints defines endpoints for sending security alerts
	AlertEndpoints []string `json:"alertEndpoints"`
}

// SecurityIncident represents a security incident
type SecurityIncident struct {
	// ID is the unique identifier for the incident
	ID string `json:"id"`

	// Title is the title of the incident
	Title string `json:"title"`

	// Description is a detailed description of the incident
	Description string `json:"description"`

	// Severity is the severity level of the incident
	Severity common.AuditSeverity `json:"severity"`

	// Status is the current status of the incident
	Status common.SecurityIncidentStatus `json:"status"`

	// DetectedAt is when the incident was detected
	DetectedAt time.Time `json:"detectedAt"`

	// ReportedBy is the ID of the user who reported the incident
	ReportedBy string `json:"reportedBy"`

	// ClosedAt is when the incident was closed (if applicable)
	ClosedAt time.Time `json:"closedAt,omitempty"`

	// Resolution is the resolution of the incident (if applicable)
	Resolution string `json:"resolution,omitempty"`

	// AffectedSystems is a list of systems affected by the incident
	AffectedSystems []string `json:"affectedSystems"`

	// RelatedLogs is a list of related log IDs
	RelatedLogs []string `json:"relatedLogs"`

	// Metadata is additional metadata for the incident
	Metadata map[string]interface{} `json:"metadata"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	// ID is the unique identifier for the vulnerability
	ID string `json:"id"`

	// Title is the title of the vulnerability
	Title string `json:"title"`

	// Description is a detailed description of the vulnerability
	Description string `json:"description"`

	// Severity is the severity level of the vulnerability
	Severity common.VulnerabilitySeverity `json:"severity"`

	// Status is the current status of the vulnerability
	Status common.VulnerabilityStatus `json:"status"`

	// DiscoveredAt is when the vulnerability was discovered
	DiscoveredAt time.Time `json:"discoveredAt"`

	// DiscoveredBy is the ID of the user who discovered the vulnerability
	DiscoveredBy string `json:"discoveredBy"`

	// MitigatedAt is when the vulnerability was mitigated (if applicable)
	MitigatedAt time.Time `json:"mitigatedAt,omitempty"`

	// Mitigation is the mitigation for the vulnerability (if applicable)
	Mitigation string `json:"mitigation,omitempty"`

	// ResolvedAt is when the vulnerability was resolved (if applicable)
	ResolvedAt time.Time `json:"resolvedAt,omitempty"`

	// AffectedSystems is a list of systems affected by the vulnerability
	AffectedSystems []string `json:"affectedSystems"`

	// Metadata is additional metadata for the vulnerability
	Metadata map[string]interface{} `json:"metadata"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	// ID is the unique identifier for the audit log
	ID string `json:"id,omitempty"`

	// Timestamp is when the action occurred
	Timestamp time.Time `json:"timestamp"`

	// UserID is the ID of the user who performed the action
	UserID string `json:"userId"`

	// Action is the action that was performed
	Action common.AuditAction `json:"action"`

	// Resource is the type of resource that was affected
	Resource string `json:"resource"`

	// ResourceID is the ID of the resource that was affected
	ResourceID string `json:"resourceId"`

	// Description is a description of the action
	Description string `json:"description"`

	// Severity is the severity level of the action
	Severity common.AuditSeverity `json:"severity"`

	// Changes is a map of changes that were made
	Changes map[string]interface{} `json:"changes,omitempty"`

	// Metadata is additional metadata for the audit log
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AccessControlConfig is an alias for the config in the parent package
// This is imported to avoid circular dependencies
type AccessControlConfig struct {
	// PasswordPolicy defines password requirements
	PasswordPolicy struct {
		MinLength        int  `json:"minLength"`
		RequireUppercase bool `json:"requireUppercase"`
		RequireLowercase bool `json:"requireLowercase"`
		RequireNumbers   bool `json:"requireNumbers"`
		RequireSymbols   bool `json:"requireSymbols"`
		MaxAge           int  `json:"maxAge"`
	} `json:"passwordPolicy"`

	// SessionPolicy defines session management settings
	SessionPolicy struct {
		MaxDuration      int  `json:"maxDuration"`
		IdleTimeout      int  `json:"idleTimeout"`
		RequireMFA       bool `json:"requireMFA"`
		SecureCookies    bool `json:"secureCookies"`
		AllowConcurrent  bool `json:"allowConcurrent"`
	} `json:"sessionPolicy"`

	// RolePermissions maps roles to their permissions
	RolePermissions map[string][]string `json:"rolePermissions,omitempty"`
}
