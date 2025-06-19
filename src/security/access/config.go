// Package access provides access control and security auditing functionality
package access

import (
	"time"
)

// AccessControlConfig represents the configuration for the access control system
type AccessControlConfig struct {
	// EnableRBAC enables role-based access control
	EnableRBAC bool `json:"enable_rbac"`

	// EnableMFA enables multi-factor authentication
	EnableMFA bool `json:"enable_mfa"`

	// MFARequiredRoles specifies roles that require MFA
	MFARequiredRoles []string `json:"mfa_required_roles"`

	// PasswordPolicy defines password requirements
	PasswordPolicy PasswordPolicy `json:"password_policy"`

	// SessionPolicy defines session management settings
	SessionPolicy SessionPolicy `json:"session_policy"`

	// AuditConfig defines audit logging settings
	AuditConfig AuditConfig `json:"audit_config"`

	// SecurityIncidentConfig defines security incident management settings
	SecurityIncidentConfig SecurityIncidentConfig `json:"security_incident_config"`

	// VulnerabilityConfig defines vulnerability management settings
	VulnerabilityConfig VulnerabilityConfig `json:"vulnerability_config"`

	// RBACConfig defines RBAC-specific settings
	RBACConfig RBACConfigSettings `json:"rbac_config"`

	// RolePermissions maps roles to their permissions
	RolePermissions map[string][]string `json:"role_permissions,omitempty"`
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	// MinLength is the minimum password length
	MinLength int `json:"min_length"`

	// RequireUppercase requires at least one uppercase letter
	RequireUppercase bool `json:"require_uppercase"`

	// RequireLowercase requires at least one lowercase letter
	RequireLowercase bool `json:"require_lowercase"`

	// RequireNumbers requires at least one number
	RequireNumbers bool `json:"require_numbers"`

	// RequireSpecialChars requires at least one special character
	RequireSpecialChars bool `json:"require_special_chars"`

	// MaxAge is the maximum password age in days (0 = no expiration)
	MaxAge int `json:"max_age"`

	// PreventReuseCount prevents reusing the last N passwords
	PreventReuseCount int `json:"prevent_reuse_count"`

	// LockoutThreshold is the number of failed login attempts before account lockout
	LockoutThreshold int `json:"lockout_threshold"`

	// LockoutDuration is the duration of account lockout in minutes
	LockoutDuration int `json:"lockout_duration"`
}

// SessionPolicy defines session management settings
type SessionPolicy struct {
	// TokenExpiration is the token expiration time in minutes
	TokenExpiration int `json:"token_expiration"`

	// RefreshTokenExpiration is the refresh token expiration time in minutes
	RefreshTokenExpiration int `json:"refresh_token_expiration"`

	// InactivityTimeout is the session inactivity timeout in minutes
	InactivityTimeout int `json:"inactivity_timeout"`

	// MaxConcurrentSessions is the maximum number of concurrent sessions per user
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`

	// EnforceIPBinding enforces session binding to IP address
	EnforceIPBinding bool `json:"enforce_ip_binding"`

	// EnforceUserAgentBinding enforces session binding to user agent
	EnforceUserAgentBinding bool `json:"enforce_user_agent_binding"`

	// CleanupInterval is the interval for cleaning up expired sessions in minutes
	CleanupInterval int `json:"cleanup_interval"`
}

// SecurityIncidentConfig defines security incident management settings
type SecurityIncidentConfig struct {
	// EnableIncidentTracking enables security incident tracking
	EnableIncidentTracking bool `json:"enable_incident_tracking"`

	// AutoCreateIncidents automatically creates incidents for high-severity events
	AutoCreateIncidents bool `json:"auto_create_incidents"`

	// NotificationEmails are email addresses to notify for security incidents
	NotificationEmails []string `json:"notification_emails"`

	// EscalationThreshold is the severity threshold for incident escalation
	EscalationThreshold AuditSeverity `json:"escalation_threshold"`

	// ResponseTimeoutMinutes is the maximum response time for incidents in minutes
	ResponseTimeoutMinutes int `json:"response_timeout_minutes"`
}

// VulnerabilityConfig defines vulnerability management settings
type VulnerabilityConfig struct {
	// EnableVulnerabilityTracking enables vulnerability tracking
	EnableVulnerabilityTracking bool `json:"enable_vulnerability_tracking"`

	// AutoScanEnabled enables automatic vulnerability scanning
	AutoScanEnabled bool `json:"auto_scan_enabled"`

	// ScanSchedule defines when automatic scans are performed
	ScanSchedule string `json:"scan_schedule"`

	// ReportRecipients are email addresses to receive vulnerability reports
	ReportRecipients []string `json:"report_recipients"`

	// RemediationDeadlineDays is the number of days to remediate vulnerabilities
	RemediationDeadlineDays map[string]int `json:"remediation_deadline_days"`
}

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    AuditSeverity          `json:"severity"`
	Status      IncidentStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  time.Time              `json:"resolved_at,omitempty"`
	AssignedTo  string                 `json:"assigned_to,omitempty"`
	ReportedBy  string                 `json:"reported_by,omitempty"`
	AuditLogIDs []string               `json:"audit_log_ids,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// IncidentStatus represents the status of a security incident
type IncidentStatus string

// Incident statuses
const (
	IncidentStatusNew        IncidentStatus = "new"
	IncidentStatusInProgress IncidentStatus = "in_progress"
	IncidentStatusResolved   IncidentStatus = "resolved"
	IncidentStatusClosed     IncidentStatus = "closed"
	IncidentStatusDuplicate  IncidentStatus = "duplicate"
)

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Severity       AuditSeverity          `json:"severity"`
	Status         VulnerabilityStatus    `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ResolvedAt     time.Time              `json:"resolved_at,omitempty"`
	AssignedTo     string                 `json:"assigned_to,omitempty"`
	ReportedBy     string                 `json:"reported_by,omitempty"`
	AffectedSystem string                 `json:"affected_system,omitempty"`
	CVE            string                 `json:"cve,omitempty"`
	RemediationPlan string                `json:"remediation_plan,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// VulnerabilityStatus represents the status of a vulnerability
type VulnerabilityStatus string

// Vulnerability statuses
const (
	VulnerabilityStatusNew        VulnerabilityStatus = "new"
	VulnerabilityStatusValidated  VulnerabilityStatus = "validated"
	VulnerabilityStatusInProgress VulnerabilityStatus = "in_progress"
	VulnerabilityStatusRemediated VulnerabilityStatus = "remediated"
	VulnerabilityStatusVerified   VulnerabilityStatus = "verified"
	VulnerabilityStatusRejected   VulnerabilityStatus = "rejected"
	VulnerabilityStatusDeferred   VulnerabilityStatus = "deferred"
)

// DefaultAccessControlConfigV2 returns the default access control configuration (version 2)
func DefaultAccessControlConfigV2() *AccessControlConfig {
	return &AccessControlConfig{
		EnableRBAC: true,
		EnableMFA:  true,
		MFARequiredRoles: []string{
			RoleAdmin,
			RoleManager,
		},
		PasswordPolicy: PasswordPolicy{
			MinLength:         12,
			RequireUppercase:  true,
			RequireLowercase:  true,
			RequireNumbers:    true,
			RequireSpecialChars: true,
			MaxAge:            90,
			PreventReuseCount: 5,
			LockoutThreshold:  5,
			LockoutDuration:   30,
		},
		SessionPolicy: SessionPolicy{
			TokenExpiration:        60,
			RefreshTokenExpiration: 1440, // 24 hours
			InactivityTimeout:      30,
			MaxConcurrentSessions:  5,
			EnforceIPBinding:       true,
			EnforceUserAgentBinding: true,
			CleanupInterval:        15,
		},
		AuditConfig: AuditConfig{
			EnabledSeverities: []AuditSeverity{
				AuditSeverityInfo,
				AuditSeverityLow,
				AuditSeverityMedium,
				AuditSeverityHigh,
				AuditSeverityCritical,
				AuditSeverityError,
			},
			LogFilePath:   "audit.log",
			RetentionDays: 90,
		},
		SecurityIncidentConfig: SecurityIncidentConfig{
			EnableIncidentTracking: true,
			AutoCreateIncidents:    true,
			NotificationEmails:     []string{},
			EscalationThreshold:    AuditSeverityHigh,
			ResponseTimeoutMinutes: 60,
		},
		VulnerabilityConfig: VulnerabilityConfig{
			EnableVulnerabilityTracking: true,
			AutoScanEnabled:             true,
			ScanSchedule:                "0 0 * * *", // Daily at midnight
			ReportRecipients:            []string{},
			RemediationDeadlineDays: map[string]int{
				string(AuditSeverityCritical): 7,
				string(AuditSeverityHigh):     14,
				string(AuditSeverityMedium):   30,
				string(AuditSeverityLow):      90,
			},
		},
	}
}

// RBACConfigSettings defines RBAC-specific configuration settings
type RBACConfigSettings struct {
	// DefaultRoles specifies the default roles available in the system
	DefaultRoles []string `json:"default_roles"`
	
	// RolePermissions maps role names to their permissions
	RolePermissions map[string][]string `json:"role_permissions"`
	
	// StrictHierarchy enables strict role hierarchy enforcement
	StrictHierarchy bool `json:"strict_hierarchy"`
	
	// AllowDirectPermissions allows direct permission assignments to users
	AllowDirectPermissions bool `json:"allow_direct_permissions"`
}
