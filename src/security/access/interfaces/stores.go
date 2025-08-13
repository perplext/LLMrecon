// Package interfaces defines the interfaces for the access control system
package interfaces

import (
	"context"

	"github.com/perplext/LLMrecon/src/security/access/models"
)

// User represents a user in the system
type User = models.User

// Session represents a user session
type Session = models.Session

// AuditEvent represents an entry in the audit log
type AuditEvent = models.AuditLog

// SecurityIncident represents a security incident
type SecurityIncident = models.SecurityIncident

// Vulnerability represents a security vulnerability
type Vulnerability = models.Vulnerability

// Role represents a role in the system
type Role struct {
	ID          string
	Name        string
	Description string
	Permissions []string
	IsBuiltIn   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserStore defines the interface for user storage operations
type UserStore interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *User) error
	
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id string) (*User, error)
	
	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	
	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	
	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *User) error
	
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id string) error
	
	// ListUsers lists users with optional filtering
	ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*User, int, error)
	
	// Close closes the user store
	Close() error
}

// SessionStore defines the interface for session storage operations
type SessionStore interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *Session) error
	
	// GetSessionByID retrieves a session by ID
	GetSessionByID(ctx context.Context, id string) (*Session, error)
	
	// GetSessionByToken retrieves a session by token
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	
	// GetSessionByRefreshToken retrieves a session by refresh token
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	
	// UpdateSession updates an existing session
	UpdateSession(ctx context.Context, session *Session) error
	
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, id string) error
	
	// DeleteSessionsByUserID deletes all sessions for a user
	DeleteSessionsByUserID(ctx context.Context, userID string) error
	
	// ListSessionsByUserID lists sessions for a user
	ListSessionsByUserID(ctx context.Context, userID string) ([]*Session, error)
	
	// CleanExpiredSessions removes expired sessions
	CleanExpiredSessions(ctx context.Context) (int, error)
	
	// Close closes the session store
	Close() error
}

// AuditLogger defines the interface for audit logging operations
type AuditLogger interface {
	// LogEvent logs an audit event
	LogEvent(ctx context.Context, event *AuditEvent) error
	
	// GetEventByID retrieves an audit event by ID
	GetEventByID(ctx context.Context, id string) (*AuditEvent, error)
	
	// QueryEvents queries audit events with filtering
	QueryEvents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*AuditEvent, int, error)
	
	// ExportEvents exports audit events to a file
	ExportEvents(ctx context.Context, filter map[string]interface{}, format string) (string, error)
	
	// Close closes the audit logger
	Close() error
}

// IncidentStore defines the interface for security incident storage operations
type IncidentStore interface {
	// CreateIncident creates a new security incident
	CreateIncident(ctx context.Context, incident *SecurityIncident) error
	
	// GetIncidentByID retrieves a security incident by ID
	GetIncidentByID(ctx context.Context, id string) (*SecurityIncident, error)
	
	// UpdateIncident updates an existing security incident
	UpdateIncident(ctx context.Context, incident *SecurityIncident) error
	
	// DeleteIncident deletes a security incident
	DeleteIncident(ctx context.Context, id string) error
	
	// ListIncidents lists security incidents with optional filtering
	ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*SecurityIncident, int, error)
	
	// Close closes the incident store
	Close() error
}

// VulnerabilityStore defines the interface for vulnerability storage operations
type VulnerabilityStore interface {
	// CreateVulnerability creates a new vulnerability
	CreateVulnerability(ctx context.Context, vulnerability *Vulnerability) error
	
	// GetVulnerabilityByID retrieves a vulnerability by ID
	GetVulnerabilityByID(ctx context.Context, id string) (*Vulnerability, error)
	
	// GetVulnerabilityByCVE retrieves a vulnerability by CVE ID
	GetVulnerabilityByCVE(ctx context.Context, cve string) (*Vulnerability, error)
	
	// UpdateVulnerability updates an existing vulnerability
	UpdateVulnerability(ctx context.Context, vulnerability *Vulnerability) error
	
	// DeleteVulnerability deletes a vulnerability
	DeleteVulnerability(ctx context.Context, id string) error
	
	// ListVulnerabilities lists vulnerabilities with optional filtering
	ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*Vulnerability, int, error)
	
	// Close closes the vulnerability store
	Close() error
}

// RoleStore defines the interface for role storage operations
type RoleStore interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, role *Role) error
	
	// GetRoleByID retrieves a role by ID
	GetRoleByID(ctx context.Context, id string) (*Role, error)
	
	// GetRoleByName retrieves a role by name
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	
	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, role *Role) error
	
	// DeleteRole deletes a role
	DeleteRole(ctx context.Context, id string) error
	
	// ListRoles lists roles with optional filtering
	ListRoles(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*Role, int, error)
	
	// Close closes the role store
	Close() error
}
