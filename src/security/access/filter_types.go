// Package access provides access control and security auditing functionality
package access

import "time"

// AuditEvent represents an audit event in the access control system
type AuditEvent struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	UserID     string                 `json:"user_id"`
	Username   string                 `json:"username"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	Severity   string                 `json:"severity"`
	Status     string                 `json:"status"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Details    map[string]interface{} `json:"details"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UserFilter defines filters for querying users
type UserFilter struct {
	Username  string   `json:"username,omitempty"`
	Email     string   `json:"email,omitempty"`
	Roles     []string `json:"roles,omitempty"`
	Active    *bool    `json:"active,omitempty"`
	Locked    *bool    `json:"locked,omitempty"`
	SortBy    string   `json:"sort_by,omitempty"`
	SortOrder string   `json:"sort_order,omitempty"`
	Offset    int      `json:"offset,omitempty"`
	Limit     int      `json:"limit,omitempty"`
}

// SessionFilter defines filters for querying sessions
type SessionFilter struct {
	UserID        string `json:"user_id,omitempty"`
	IPAddress     string `json:"ip_address,omitempty"`
	UserAgent     string `json:"user_agent,omitempty"`
	MFACompleted  *bool  `json:"mfa_completed,omitempty"`
	Active        *bool  `json:"active,omitempty"`
	CreatedAfter  string `json:"created_after,omitempty"`
	CreatedBefore string `json:"created_before,omitempty"`
	SortBy        string `json:"sort_by,omitempty"`
	SortOrder     string `json:"sort_order,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

// IncidentFilter defines filters for querying security incidents
type IncidentFilter struct {
	Severity      string `json:"severity,omitempty"`
	Status        string `json:"status,omitempty"`
	AssigneeID    string `json:"assignee_id,omitempty"`
	ReportedAfter string `json:"reported_after,omitempty"`
	SortBy        string `json:"sort_by,omitempty"`
	SortOrder     string `json:"sort_order,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

// VulnerabilityFilter defines filters for querying vulnerabilities
type VulnerabilityFilter struct {
	Severity      string `json:"severity,omitempty"`
	Status        string `json:"status,omitempty"`
	CveID         string `json:"cve_id,omitempty"`
	Component     string `json:"component,omitempty"`
	ReportedAfter string `json:"reported_after,omitempty"`
	SortBy        string `json:"sort_by,omitempty"`
	SortOrder     string `json:"sort_order,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

// AuditEventFilter defines filters for querying audit events
type AuditEventFilter struct {
	UserID        string     `json:"user_id,omitempty"`
	Username      string     `json:"username,omitempty"`
	Action        string     `json:"action,omitempty"`
	Resource      string     `json:"resource,omitempty"`
	ResourceID    string     `json:"resource_id,omitempty"`
	Severity      string     `json:"severity,omitempty"`
	Status        string     `json:"status,omitempty"`
	IPAddress     string     `json:"ip_address,omitempty"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	CreatedAfter  string     `json:"created_after,omitempty"`
	CreatedBefore string     `json:"created_before,omitempty"`
	SortBy        string     `json:"sort_by,omitempty"`
	SortOrder     string     `json:"sort_order,omitempty"`
	Offset        int        `json:"offset,omitempty"`
	Limit         int        `json:"limit,omitempty"`
}
