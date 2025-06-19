// Package access provides access control and security auditing functionality
package access

import (
	"context"
)

// UserStore defines the interface for user storage
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
	
	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id string) error
	
	// ListUsers lists all users
	ListUsers(ctx context.Context) ([]*User, error)
	
	// Close closes the store
	Close() error
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *Session) error
	
	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, id string) (*Session, error)
	
	// UpdateSession updates an existing session
	UpdateSession(ctx context.Context, session *Session) error
	
	// DeleteSession deletes a session by ID
	DeleteSession(ctx context.Context, id string) error
	
	// GetUserSessions retrieves all sessions for a user
	GetUserSessions(ctx context.Context, userID string) ([]*Session, error)
	
	// CleanExpiredSessions removes all expired sessions
	CleanExpiredSessions(ctx context.Context) error
	
	// Close closes the store
	Close() error
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	// Initialize initializes the audit logger
	Initialize(ctx context.Context) error
	
	// LogAudit logs an audit event
	LogAudit(ctx context.Context, log *AuditLog) error
	
	// GetAuditLogs retrieves audit logs
	GetAuditLogs(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*AuditLog, int, error)
	
	// GetAuditLogByID retrieves an audit log by ID
	GetAuditLogByID(ctx context.Context, id string) (*AuditLog, error)
	
	// Close closes the logger
	Close(ctx context.Context) error
}

// TypesIncidentStore defines the interface for security incident storage using types.SecurityIncident
type TypesIncidentStore interface {
	// CreateIncident creates a new security incident
	CreateIncident(ctx context.Context, incident *SecurityIncident) error
	
	// GetIncidentByID retrieves a security incident by ID
	GetIncidentByID(ctx context.Context, id string) (*SecurityIncident, error)
	
	// UpdateIncident updates an existing security incident
	UpdateIncident(ctx context.Context, incident *SecurityIncident) error
	
	// DeleteIncident deletes a security incident by ID
	DeleteIncident(ctx context.Context, id string) error
	
	// ListIncidents lists security incidents
	ListIncidents(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*SecurityIncident, int, error)
	
	// Close closes the store
	Close() error
}

// TypesVulnerabilityStore defines the interface for vulnerability storage using types.Vulnerability
type TypesVulnerabilityStore interface {
	// CreateVulnerability creates a new vulnerability
	CreateVulnerability(ctx context.Context, vulnerability *Vulnerability) error
	
	// GetVulnerabilityByID retrieves a vulnerability by ID
	GetVulnerabilityByID(ctx context.Context, id string) (*Vulnerability, error)
	
	// UpdateVulnerability updates an existing vulnerability
	UpdateVulnerability(ctx context.Context, vulnerability *Vulnerability) error
	
	// DeleteVulnerability deletes a vulnerability by ID
	DeleteVulnerability(ctx context.Context, id string) error
	
	// ListVulnerabilities lists vulnerabilities
	ListVulnerabilities(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*Vulnerability, int, error)
	
	// Close closes the store
	Close() error
}
