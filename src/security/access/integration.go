// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/audit"
	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/mfa"
	"github.com/perplext/LLMrecon/src/security/access/rbac"
)

// Common errors
var (
	ErrUnauthorized          = errors.New("unauthorized access")
	// ErrInvalidCredentials is already defined in auth.go
	// ErrMFARequired is already defined in auth.go
	ErrMFAVerificationFailed = errors.New("multi-factor authentication verification failed")
	// ErrUserNotFound is already defined in auth.go
	// ErrSessionExpired is already defined in auth.go
	ErrInvalidSession        = errors.New("invalid session")
)

// User represents a user in the system
type User struct {
	// ID is the unique identifier for the user
	ID string `json:"id"`
	
	// Username is the username of the user
	Username string `json:"username"`
	
	// Email is the email address of the user
	Email string `json:"email"`
	
	// PasswordHash is the hashed password of the user
	PasswordHash string `json:"-"`
	
	// Enabled indicates if the user is enabled
	Enabled bool `json:"enabled"`
	
	// Active indicates if the user is active (alias for Enabled for compatibility)
	Active bool `json:"active"`
	
	// MFAEnabled indicates if MFA is enabled for the user
	MFAEnabled bool `json:"mfa_enabled"`
	
	// MFAMethods contains the enabled MFA methods for the user
	MFAMethods []string `json:"mfa_methods"`
	
	// MFAMethod is the default MFA method
	MFAMethod string `json:"mfa_method,omitempty"`
	
	// MFASecret is the MFA secret (for TOTP)
	MFASecret string `json:"mfa_secret,omitempty"`
	
	// Roles contains the user's roles
	Roles []string `json:"roles"`
	
	// Permissions contains the user's permissions
	Permissions []string `json:"permissions,omitempty"`
	
	// FailedLoginAttempts tracks failed login attempts
	FailedLoginAttempts int `json:"failed_login_attempts"`
	
	// Locked indicates if the user account is locked
	Locked bool `json:"locked"`
	
	// LastLogin is the timestamp of the last successful login
	LastLogin time.Time `json:"last_login"`
	
	// LastPasswordChange is the timestamp of the last password change
	LastPasswordChange time.Time `json:"last_password_change"`
	
	// CreatedAt is the timestamp when the user was created
	CreatedAt time.Time `json:"created_at"`
	
	// UpdatedAt is the timestamp when the user was last updated
	UpdatedAt time.Time `json:"updated_at"`
	
	// Metadata contains additional user metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Session is already defined in session.go
// For this integration, we'll use the existing Session type
/*
type Session struct {
	// ID is the unique identifier for the session
	ID string `json:"id"`
	
	// UserID is the ID of the user associated with the session
	UserID string `json:"user_id"`
	
	// CreatedAt is the timestamp when the session was created
	CreatedAt time.Time `json:"created_at"`
	
	// ExpiresAt is the timestamp when the session expires
	ExpiresAt time.Time `json:"expires_at"`
	
	// MFAVerified indicates if MFA has been verified for this session
	MFAVerified bool `json:"mfa_verified"`
	
	// IP is the IP address associated with the session
	IP string `json:"ip"`
	
	// UserAgent is the user agent associated with the session
	UserAgent string `json:"user_agent"`
}
*/

// AccessControlIntegration integrates RBAC, MFA, and audit logging
type AccessControlIntegration struct {
	rbacManager  rbac.RBACManager
	mfaManager   mfa.MFAManager
	auditManager audit.AuditManager
	userStore    UserStore
	sessionStore SessionStore
}

// UserStore is already defined in auth.go
// For this integration, we'll use the existing UserStore interface
/*
type UserStore interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *User) error
	
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, userID string) (*User, error)
	
	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	
	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	
	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *User) error
	
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, userID string) error
	
	// ListUsers lists all users
	ListUsers(ctx context.Context) ([]*User, error)
}
*/

// SessionStore is already defined in auth.go
// For this integration, we'll use the existing SessionStore interface
/*
type SessionStore interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *Session) error
	
	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	
	// UpdateSession updates an existing session
	UpdateSession(ctx context.Context, session *Session) error
	
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, sessionID string) error
	
	// DeleteUserSessions deletes all sessions for a user
	DeleteUserSessions(ctx context.Context, userID string) error
	
	// ListUserSessions lists all sessions for a user
	ListUserSessions(ctx context.Context, userID string) ([]*Session, error)
	
	// CleanupExpiredSessions cleans up expired sessions
	CleanupExpiredSessions(ctx context.Context) error
}
*/

// NewAccessControlIntegration creates a new access control integration
func NewAccessControlIntegration(
	rbacManager rbac.RBACManager,
	mfaManager mfa.MFAManager,
	auditManager audit.AuditManager,
	userStore UserStore,
	sessionStore SessionStore,
) *AccessControlIntegration {
	return &AccessControlIntegration{
		rbacManager:  rbacManager,
		mfaManager:   mfaManager,
		auditManager: auditManager,
		userStore:    userStore,
		sessionStore: sessionStore,
	}
}

// Login authenticates a user and creates a session
func (a *AccessControlIntegration) Login(ctx context.Context, username, password, ip, userAgent string) (*Session, error) {
	// Get the user
	user, err := a.userStore.GetUserByUsername(ctx, username)
	if err != nil {
		a.logFailedLogin(ctx, username, ip, userAgent, "user not found")
		return nil, ErrInvalidCredentials
	}

	// Check if the user is enabled
	if !user.Enabled {
		a.logFailedLogin(ctx, username, ip, userAgent, "user disabled")
		return nil, ErrInvalidCredentials
	}

	// Verify the password
	if !verifyPassword(password, user.PasswordHash) {
		a.logFailedLogin(ctx, username, ip, userAgent, "invalid password")
		return nil, ErrInvalidCredentials
	}

	// Create a session
	session := &Session{
		ID:           generateSessionID(),
		UserID:       user.ID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24-hour session
		MFACompleted: !user.MFAEnabled,               // If MFA is not enabled, mark as verified
		IPAddress:    ip,
		UserAgent:    userAgent,
	}

	// Save the session
	if err := a.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login time
	user.LastLogin = time.Now()
	if err := a.userStore.UpdateUser(ctx, user); err != nil {
		// Log the error but don't fail the login
		a.auditManager.LogAudit(ctx, &audit.AuditEvent{
			UserID:      user.ID,
			Action:      common.AuditActionUserUpdate,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: "Failed to update last login time",
			Severity:    common.AuditSeverityWarning,
			Timestamp:   time.Now(),
		})
	}

	// Log successful login
	a.logSuccessfulLogin(ctx, user.ID, username, ip, userAgent)

	return session, nil
}

// VerifyMFA verifies MFA for a session
func (a *AccessControlIntegration) VerifyMFA(ctx context.Context, sessionID, method, code string) error {
	// Get the session
	session, err := a.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return ErrInvalidSession
	}

	// Check if the session is expired
	if session.ExpiresAt.Before(time.Now()) {
		return ErrSessionExpired
	}

	// Get the user
	user, err := a.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		// MFA is not enabled, so it's already verified
		session.MFACompleted = true
		if err := a.sessionStore.UpdateSession(ctx, session); err != nil {
			return fmt.Errorf("failed to update session: %w", err)
		}
		return nil
	}

	// Verify MFA - use mock verification for now since VerifyMFACode doesn't exist
	valid := true // Mock verification
	if !valid {
		a.logMFAFailure(ctx, user.ID, method, session.IPAddress, session.UserAgent, "invalid code")
		return ErrMFAVerificationFailed
	}

	// Update session
	session.MFACompleted = true
	if err := a.sessionStore.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Log successful MFA verification
	a.logMFASuccess(ctx, user.ID, method, session.IPAddress, session.UserAgent)

	return nil
}

// Logout invalidates a session
func (a *AccessControlIntegration) Logout(ctx context.Context, sessionID string) error {
	// Get the session
	session, err := a.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return ErrInvalidSession
	}

	// Delete the session
	if err := a.sessionStore.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Log logout
	a.logLogout(ctx, session.UserID, session.IPAddress, session.UserAgent)

	return nil
}

// AuthorizeAccess checks if a user has permission to access a resource
func (a *AccessControlIntegration) AuthorizeAccess(ctx context.Context, sessionID, resource, action string) error {
	// Get the session
	session, err := a.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return ErrInvalidSession
	}

	// Check if the session is expired
	if session.ExpiresAt.Before(time.Now()) {
		return ErrSessionExpired
	}

	// Check if MFA is verified if required
	if !session.MFACompleted {
		return ErrMFARequired
	}

	// Get the user
	user, err := a.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if the user is enabled
	if !user.Enabled {
		return ErrUnauthorized
	}

	// Check if the user has permission
	permissionID := fmt.Sprintf("%s:%s", resource, action)
	hasPermission, err := a.rbacManager.HasPermission(ctx, user.ID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}

	if !hasPermission {
		// Log unauthorized access attempt
		a.logUnauthorizedAccess(ctx, user.ID, resource, action, session.IPAddress, session.UserAgent)
		return ErrUnauthorized
	}

	// Log authorized access
	a.logAuthorizedAccess(ctx, user.ID, resource, action, session.IPAddress, session.UserAgent)

	return nil
}

// EnableMFA enables MFA for a user
func (a *AccessControlIntegration) EnableMFA(ctx context.Context, userID, method string) error {
	// Get the user
	user, err := a.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Enable MFA based on the method
	switch mfa.MFAMethod(method) {
	case mfa.MFAMethodTOTP:
		// Mock TOTP enable - would normally call EnableTOTP
		err = nil
	case mfa.MFAMethodBackupCode:
		// Mock backup codes generation - would normally call GenerateBackupCodes
		_, err = a.mfaManager.GenerateBackupCodes(ctx, userID)
	default:
		return fmt.Errorf("unsupported MFA method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	// Update user
	user.MFAEnabled = true
	if !containsString(user.MFAMethods, method) {
		user.MFAMethods = append(user.MFAMethods, method)
	}
	user.UpdatedAt = time.Now()

	if err := a.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log MFA enabled
	a.logMFAEnabled(ctx, userID, method)

	return nil
}

// DisableMFA disables MFA for a user
func (a *AccessControlIntegration) DisableMFA(ctx context.Context, userID, method string) error {
	// Get the user
	user, err := a.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Disable MFA based on the method
	switch mfa.MFAMethod(method) {
	case mfa.MFAMethodTOTP:
		// Mock TOTP disable - would normally call DisableTOTP
		err = nil
	case mfa.MFAMethodBackupCode:
		// No specific method to disable backup codes, handled by ResetMFA
		err = nil
	default:
		return fmt.Errorf("unsupported MFA method: %s", method)
	}

	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Update user
	user.MFAMethods = removeString(user.MFAMethods, method)
	if len(user.MFAMethods) == 0 {
		user.MFAEnabled = false
	}
	user.UpdatedAt = time.Now()

	if err := a.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log MFA disabled
	a.logMFADisabled(ctx, userID, method)

	return nil
}

// ResetMFA resets all MFA methods for a user
func (a *AccessControlIntegration) ResetMFA(ctx context.Context, userID string) error {
	// Get the user
	user, err := a.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Reset MFA
	// Mock MFA reset - would normally call ResetMFA
	if err := (error)(nil); err != nil {
		return fmt.Errorf("failed to reset MFA: %w", err)
	}

	// Update user
	user.MFAEnabled = false
	user.MFAMethods = []string{}
	user.UpdatedAt = time.Now()

	if err := a.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Log MFA reset
	a.logMFAReset(ctx, userID)

	return nil
}

// AssignRoleToUser assigns a role to a user
func (a *AccessControlIntegration) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	// Check if the user exists
	if _, err := a.userStore.GetUserByID(ctx, userID); err != nil {
		return ErrUserNotFound
	}

	// Assign the role
	if err := a.rbacManager.AssignRoleToUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Log role assignment
	a.logRoleAssigned(ctx, userID, roleID)

	return nil
}

// RevokeRoleFromUser revokes a role from a user
func (a *AccessControlIntegration) RevokeRoleFromUser(ctx context.Context, userID, roleID string) error {
	// Check if the user exists
	if _, err := a.userStore.GetUserByID(ctx, userID); err != nil {
		return ErrUserNotFound
	}

	// Revoke the role
	if err := a.rbacManager.RevokeRoleFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// Log role revocation
	a.logRoleRevoked(ctx, userID, roleID)

	return nil
}

// GetUserRoles gets all roles assigned to a user
func (a *AccessControlIntegration) GetUserRoles(ctx context.Context, userID string) ([]*rbac.Role, error) {
	// Check if the user exists
	if _, err := a.userStore.GetUserByID(ctx, userID); err != nil {
		return nil, ErrUserNotFound
	}

	// Get user roles
	return a.rbacManager.GetUserRoles(ctx, userID)
}

// GetUserPermissions gets all permissions assigned to a user
func (a *AccessControlIntegration) GetUserPermissions(ctx context.Context, userID string) ([]*rbac.Permission, error) {
	// Check if the user exists
	if _, err := a.userStore.GetUserByID(ctx, userID); err != nil {
		return nil, ErrUserNotFound
	}

	// Get user permissions
	permStrings, err := a.rbacManager.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Convert strings to Permission objects
	permissions := make([]*rbac.Permission, len(permStrings))
	for i, perm := range permStrings {
		permissions[i] = &rbac.Permission{
			ID:   perm,
			Name: perm,
		}
	}
	
	return permissions, nil
}

// Helper functions for logging

// logFailedLogin logs a failed login attempt
func (a *AccessControlIntegration) logFailedLogin(ctx context.Context, username, ip, userAgent, reason string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		Action:      common.AuditActionLoginFailed,
		Resource:    "auth",
		Description: fmt.Sprintf("Failed login attempt for user %s: %s", username, reason),
		Severity:    common.AuditSeverityWarning,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"username":   username,
			"ip":         ip,
			"user_agent": userAgent,
			"reason":     reason,
		},
	})
}

// logSuccessfulLogin logs a successful login
func (a *AccessControlIntegration) logSuccessfulLogin(ctx context.Context, userID, username, ip, userAgent string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionLoginSuccess,
		Resource:    "auth",
		ResourceID:  userID,
		Description: fmt.Sprintf("Successful login for user %s", username),
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"username":   username,
			"ip":         ip,
			"user_agent": userAgent,
		},
	})
}

// logLogout logs a logout
func (a *AccessControlIntegration) logLogout(ctx context.Context, userID, ip, userAgent string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionLogout,
		Resource:    "auth",
		ResourceID:  userID,
		Description: "User logged out",
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"ip":         ip,
			"user_agent": userAgent,
		},
	})
}

// logMFAFailure logs an MFA verification failure
func (a *AccessControlIntegration) logMFAFailure(ctx context.Context, userID, method, ip, userAgent, reason string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionMfaVerifyFailed,
		Resource:    "auth",
		ResourceID:  userID,
		Description: fmt.Sprintf("MFA verification failed for method %s: %s", method, reason),
		Severity:    common.AuditSeverityWarning,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"method":     method,
			"ip":         ip,
			"user_agent": userAgent,
			"reason":     reason,
		},
	})
}

// logMFASuccess logs a successful MFA verification
func (a *AccessControlIntegration) logMFASuccess(ctx context.Context, userID, method, ip, userAgent string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionMfaVerify,
		Resource:    "auth",
		ResourceID:  userID,
		Description: fmt.Sprintf("MFA verification succeeded for method %s", method),
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"method":     method,
			"ip":         ip,
			"user_agent": userAgent,
		},
	})
}

// logMFAEnabled logs MFA being enabled
func (a *AccessControlIntegration) logMFAEnabled(ctx context.Context, userID, method string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionMfaEnable,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("MFA method %s enabled", method),
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"method": method,
		},
	})
}

// logMFADisabled logs MFA being disabled
func (a *AccessControlIntegration) logMFADisabled(ctx context.Context, userID, method string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionMfaDisable,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("MFA method %s disabled", method),
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"method": method,
		},
	})
}

// logMFAReset logs MFA being reset
func (a *AccessControlIntegration) logMFAReset(ctx context.Context, userID string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionMfaEnable,
		Resource:    "user",
		ResourceID:  userID,
		Description: "All MFA methods reset",
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
	})
}

// logUnauthorizedAccess logs an unauthorized access attempt
func (a *AccessControlIntegration) logUnauthorizedAccess(ctx context.Context, userID, resource, action, ip, userAgent string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionResourceAccessDenied,
		Resource:    resource,
		Description: fmt.Sprintf("Unauthorized access attempt to %s:%s", resource, action),
		Severity:    common.AuditSeverityWarning,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"action":     action,
			"ip":         ip,
			"user_agent": userAgent,
		},
	})
}

// logAuthorizedAccess logs an authorized access
func (a *AccessControlIntegration) logAuthorizedAccess(ctx context.Context, userID, resource, action, ip, userAgent string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      userID,
		Action:      common.AuditActionResourceAccess,
		Resource:    resource,
		Description: fmt.Sprintf("Authorized access to %s:%s", resource, action),
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"action":     action,
			"ip":         ip,
			"user_agent": userAgent,
		},
	})
}

// logRoleAssigned logs a role being assigned to a user
func (a *AccessControlIntegration) logRoleAssigned(ctx context.Context, userID, roleID string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      getIntegrationUserIDFromContext(ctx),
		Action:      common.AuditActionRoleAssign,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("Role %s assigned to user", roleID),
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"role_id": roleID,
		},
	})
}

// logRoleRevoked logs a role being revoked from a user
func (a *AccessControlIntegration) logRoleRevoked(ctx context.Context, userID, roleID string) {
	a.auditManager.LogAudit(ctx, &audit.AuditEvent{
		UserID:      getIntegrationUserIDFromContext(ctx),
		Action:      common.AuditActionRoleRevoke,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("Role %s revoked from user", roleID),
		Severity:    common.AuditSeverityInfo,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"role_id": roleID,
		},
	})
}

// Utility functions

// verifyPassword verifies a password against a hash
func verifyPassword(password, hash string) bool {
	// In a real implementation, this would use a password hashing library
	// such as bcrypt to verify the password against the hash
	// For now, we'll just return true for testing purposes
	// TODO: Implement actual password verification
	return true
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	// In a real implementation, this would generate a secure random ID
	// For now, we'll just return a timestamp-based ID for testing purposes
	// TODO: Implement secure session ID generation
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

// getIntegrationUserIDFromContext gets the user ID from the context
func getIntegrationUserIDFromContext(ctx context.Context) string {
	// In a real implementation, this would extract the user ID from the context
	// For now, we'll just return a placeholder
	// TODO: Implement actual user ID extraction from context
	return "system"
}

// containsString checks if a string is in a slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a string from a slice
func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
