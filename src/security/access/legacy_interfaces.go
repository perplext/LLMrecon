// Package access provides access control functionality
package access

import (
	"github.com/perplext/LLMrecon/src/security/access/models"
)


// SecurityManager defines the interface for managing security incidents and vulnerabilities
type SecurityManager interface {
	// ReportIncident reports a security incident
	ReportIncident(title, description string, severity models.SecurityIncidentSeverity) (*models.SecurityIncident, error)

	// GetIncident retrieves a security incident by ID
	GetIncident(incidentID string) (*models.SecurityIncident, error)

	// UpdateIncident updates a security incident
	UpdateIncident(incident *models.SecurityIncident) error

	// CloseIncident closes a security incident
	CloseIncident(incidentID, resolution string) error

	// ListIncidents lists security incidents
	ListIncidents(filter map[string]interface{}, offset, limit int) ([]*models.SecurityIncident, int, error)

	// ReportVulnerability reports a security vulnerability
	ReportVulnerability(title, description string, severity models.VulnerabilitySeverity) (*models.Vulnerability, error)

	// GetVulnerability retrieves a security vulnerability by ID
	GetVulnerability(vulnerabilityID string) (*models.Vulnerability, error)

	// UpdateVulnerability updates a security vulnerability
	UpdateVulnerability(vulnerability *models.Vulnerability) error

	// MitigateVulnerability marks a security vulnerability as mitigated
	MitigateVulnerability(vulnerabilityID, mitigation string) error

	// ResolveVulnerability marks a security vulnerability as resolved
	ResolveVulnerability(vulnerabilityID string) error

	// ListVulnerabilities lists security vulnerabilities
	ListVulnerabilities(filter map[string]interface{}, offset, limit int) ([]*models.Vulnerability, int, error)

	// Close closes the security manager
	Close() error
}

// UserManager defines the interface for managing users
type UserManager interface {
	// CreateUser creates a new user
	CreateUser(user *models.User) error

	// GetUser retrieves a user by ID
	GetUser(userID string) (*models.User, error)

	// GetUserByUsername retrieves a user by username
	GetUserByUsername(username string) (*models.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(email string) (*models.User, error)

	// UpdateUser updates a user
	UpdateUser(user *models.User) error

	// DeleteUser deletes a user
	DeleteUser(userID string) error

	// ListUsers lists users
	ListUsers(filter map[string]interface{}, offset, limit int) ([]*models.User, int, error)

	// Close closes the user manager
	Close() error
}

// LegacyAuthManagerInterface defines the interface for authentication (legacy)
type LegacyAuthManagerInterface interface {
	// Login authenticates a user
	Login(username, password string) (*models.Session, error)

	// Logout logs out a user
	Logout(sessionID string) error

	// ValidateSession validates a session
	ValidateSession(sessionID string) (*models.Session, error)

	// RefreshSession refreshes a session
	RefreshSession(sessionID string) (*models.Session, error)

	// ChangePassword changes a user's password
	ChangePassword(userID, oldPassword, newPassword string) error

	// ResetPassword resets a user's password
	ResetPassword(userID, newPassword string) error

	// Close closes the auth manager
	Close() error
}

// RBACManager defines the interface for role-based access control
type RBACManager interface {
	// HasPermission checks if a user has a permission
	HasPermission(userID string, permission string) (bool, error)

	// HasRole checks if a user has a role
	HasRole(userID string, role string) (bool, error)

	// AddRoleToUser adds a role to a user
	AddRoleToUser(userID string, role string) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(userID string, role string) error

	// GetUserRoles gets a user's roles
	GetUserRoles(userID string) ([]string, error)

	// GetUserPermissions gets a user's permissions
	GetUserPermissions(userID string) ([]string, error)

	// Close closes the RBAC manager
	Close() error
}

// LegacyAuditLogger defines the interface for audit logging (legacy)
type LegacyAuditLogger interface {
	// LogAudit logs an audit event
	LogAudit(log *models.AuditLog) error

	// GetAuditLogs gets audit logs
	GetAuditLogs(filter map[string]interface{}, offset, limit int) ([]*models.AuditLog, int, error)

	// Close closes the audit logger
	Close() error
}

// Factory creates access control system components
type Factory interface {
	// CreateAccessControlSystem creates a new access control system
	CreateAccessControlSystem() (AccessControlSystem, error)

	// CreateSecurityManager creates a new security manager
	CreateSecurityManager() (SecurityManager, error)

	// CreateUserManager creates a new user manager
	CreateUserManager() (UserManager, error)

	// CreateAuthManager creates a new auth manager
	CreateAuthManager() (LegacyAuthManagerInterface, error)

	// CreateRBACManager creates a new RBAC manager
	CreateRBACManager() (RBACManager, error)

	// CreateAuditLogger creates a new audit logger
	CreateAuditLogger() (LegacyAuditLogger, error)
}
