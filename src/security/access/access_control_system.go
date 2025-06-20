// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/mfa"
)

// simpleRBACManager is a simple implementation of RBAC
type simpleRBACManager struct {
	config          *AccessControlConfig
	rolePermissions map[string][]string
	userRoles       map[string][]string
	mu              sync.RWMutex
}

// Initialize initializes the RBAC manager
func (r *simpleRBACManager) Initialize(ctx context.Context) error {
	// Initialize role permissions from config
	if r.config.RolePermissions != nil {
		for roleStr, perms := range r.config.RolePermissions {
			r.rolePermissions[roleStr] = perms
		}
	}
	return nil
}

// RoleHasPermission checks if a role has a specific permission
func (r *simpleRBACManager) RoleHasPermission(role, permission string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	perms, exists := r.rolePermissions[role]
	if !exists {
		return false
	}

	for _, perm := range perms {
		if perm == permission {
			return true
		}
	}
	return false
}

// simpleSecurityManager manages security incidents and vulnerabilities
type simpleSecurityManager struct {
	auditLogger AuditLogger
	config      *SecurityConfig
}

// Initialize initializes the security manager
func (s *simpleSecurityManager) Initialize(ctx context.Context) error {
	return nil
}

// AccessControlSystem is the main entry point for the access control system
type AccessControlSystem struct {
	config          *AccessControlConfig
	authManager     *AuthManager
	rbacManager     *simpleRBACManager
	auditManager    *AuditManager
	securityManager *simpleSecurityManager
	mfaManager      mfa.MFAManager
	mu              sync.RWMutex
}

// NewAccessControlSystem creates a new access control system
func NewAccessControlSystem(config *AccessControlConfig) (*AccessControlSystem, error) {
	if config == nil {
		config = DefaultAccessControlConfig()
	}

	// Create the audit manager
	auditLogger := NewMultiAuditLogger(
		NewInMemoryAuditLogger(),
		NewFileAuditLogger(config.AuditConfig.LogFilePath),
	)
	auditManager, err := NewAuditManager(auditLogger, &config.AuditConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit manager: %w", err)
	}

	// Initialize the audit logger
	if err := auditLogger.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize audit logger: %w", err)
	}

	// Create user and session stores
	userStore := NewInMemoryUserStore()
	sessionStore := NewInMemorySessionStore()

	// Create MFA manager
	mfaManager := mfa.NewMockMFAManager()

	// Create auth config from access control config
	authConfig := &AuthConfig{
		SessionTimeout:         time.Duration(config.SessionPolicy.TokenExpiration) * time.Minute,
		SessionMaxInactive:     time.Duration(config.SessionPolicy.InactivityTimeout) * time.Minute,
		PasswordMinLength:      config.PasswordPolicy.MinLength,
		PasswordRequireUpper:   config.PasswordPolicy.RequireUppercase,
		PasswordRequireLower:   config.PasswordPolicy.RequireLowercase,
		PasswordRequireNumber:  config.PasswordPolicy.RequireNumbers,
		PasswordRequireSpecial: config.PasswordPolicy.RequireSpecialChars,
		PasswordMaxAge:         time.Duration(config.PasswordPolicy.MaxAge) * 24 * time.Hour,
		MFAEnabled:             config.EnableMFA,
	}

	// Create the auth manager
	authManager, err := NewAuthManager(userStore, sessionStore, auditLogger, mfaManager, authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Create the RBAC manager
	rbacManager := &simpleRBACManager{
		config:          config,
		rolePermissions: make(map[string][]string),
		userRoles:       make(map[string][]string),
	}

	// MFA manager already created above

	// Create the security manager if configured
	var securityManager *simpleSecurityManager
	if config.SecurityIncidentConfig.EnableIncidentTracking || config.VulnerabilityConfig.EnableVulnerabilityTracking {

		// Create security config from access control config
		securityConfig := &SecurityConfig{
			IncidentNotificationEmails:      config.SecurityIncidentConfig.NotificationEmails,
			IncidentEscalationThreshold:     common.AuditSeverity(config.SecurityIncidentConfig.EscalationThreshold),
			IncidentAutoClose:               time.Duration(config.SecurityIncidentConfig.ResponseTimeoutMinutes) * time.Minute,
			VulnerabilityCheckPeriod:        24 * time.Hour,
			VulnerabilityNotificationEmails: config.VulnerabilityConfig.ReportRecipients,
		}

		// Create the security manager
		securityManager = &simpleSecurityManager{
			auditLogger: auditLogger,
			config:      securityConfig,
		}
	}

	return &AccessControlSystem{
		config:          config,
		authManager:     authManager,
		rbacManager:     rbacManager,
		auditManager:    auditManager,
		securityManager: securityManager,
		mfaManager:      mfaManager,
	}, nil
}

// Initialize initializes the access control system
func (s *AccessControlSystem) Initialize(ctx context.Context) error {
	// Initialize the auth manager
	if err := s.authManager.Initialize(ctx); err != nil {
		return fmt.Errorf("error initializing auth manager: %w", err)
	}

	// Initialize the RBAC manager
	if err := s.rbacManager.Initialize(ctx); err != nil {
		return fmt.Errorf("error initializing RBAC manager: %w", err)
	}

	// Initialize the security manager if available
	if s.securityManager != nil {
		if err := s.securityManager.Initialize(ctx); err != nil {
			return fmt.Errorf("error initializing security manager: %w", err)
		}
	}

	return nil
}

// Auth returns the auth manager
func (s *AccessControlSystem) Auth() *AuthManager {
	return s.authManager
}

// RBAC returns the RBAC manager
func (s *AccessControlSystem) RBAC() *simpleRBACManager {
	return s.rbacManager
}

// Audit returns the audit manager
func (s *AccessControlSystem) Audit() *AuditManager {
	return s.auditManager
}

// Security returns the security manager
func (s *AccessControlSystem) Security() *simpleSecurityManager {
	return s.securityManager
}

// MFA returns the multi-factor authentication manager
func (s *AccessControlSystem) MFA() mfa.MFAManager {
	return s.mfaManager
}

// GetAllUsers returns all users in the system
func (s *AccessControlSystem) GetAllUsers(ctx context.Context) ([]*User, error) {
	return s.authManager.GetAllUsers(ctx)
}

// GetUserByID retrieves a user by ID
func (s *AccessControlSystem) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.authManager.GetUserByID(ctx, id)
}

// GetUserByUsername retrieves a user by username
func (s *AccessControlSystem) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.authManager.GetUserByUsername(ctx, username)
}

// CreateUser creates a new user
func (s *AccessControlSystem) CreateUser(ctx context.Context, username, email, password string, roles []string) (*User, error) {
	user, err := s.authManager.CreateUser(ctx, username, email, password, roles)
	if err != nil {
		return nil, err
	}

	// Log the action if security manager is available
	if s.securityManager != nil && user != nil {
		s.securityManager.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      getUserIDFromContext(ctx),
			Action:      AuditAction(common.AuditActionUserCreate),
			Resource:    "user",
			ResourceID:  user.ID,
			Description: fmt.Sprintf("Created user %s", username),
			Severity:    AuditSeverity(common.AuditSeverityInfo),
		})
	}

	return user, nil
}

// UpdateUser updates an existing user
func (a *AccessControlSystem) UpdateUser(ctx context.Context, user *User) error {
	// Ensure user is not nil and has required fields
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// Get the old user for comparison
	oldUser, err := a.authManager.GetUserByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// Update the user
	if err := a.authManager.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Log the action
	changes := getChanges(oldUser, user)
	if a.securityManager != nil {
		a.securityManager.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      getUserIDFromContext(ctx),
			Action:      AuditAction(common.AuditActionUserUpdate),
			Resource:    "user",
			ResourceID:  user.ID,
			Description: fmt.Sprintf("Updated user %s", user.Username),
			Severity:    AuditSeverity(common.AuditSeverityInfo),
			Changes:     changes,
		})
	}

	return nil
}

// DeleteUser deletes a user
func (a *AccessControlSystem) DeleteUser(ctx context.Context, userID string) error {
	// Get the user for logging
	user, err := a.authManager.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Delete the user
	if err := a.authManager.DeleteUser(ctx, userID); err != nil {
		return err
	}

	// Log the action
	if a.securityManager != nil && user != nil {
		a.securityManager.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      getUserIDFromContext(ctx),
			Action:      AuditAction(common.AuditActionUserDelete),
			Resource:    "user",
			ResourceID:  userID,
			Description: fmt.Sprintf("Deleted user %s", user.Username),
			Severity:    AuditSeverity(common.AuditSeverityInfo),
		})
	}

	return nil
}

// UpdateUserPassword updates a user's password
func (a *AccessControlSystem) UpdateUserPassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get the user for logging
	user, err := a.authManager.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Update the password
	if err := a.authManager.UpdateUserPassword(ctx, userID, currentPassword, newPassword); err != nil {
		return err
	}

	// Log the action
	if a.securityManager != nil && user != nil {
		a.securityManager.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      getUserIDFromContext(ctx),
			Action:      AuditAction(common.AuditActionUserPasswordChange),
			Resource:    "user",
			ResourceID:  userID,
			Description: fmt.Sprintf("Updated password for user %s", user.Username),
			Severity:    AuditSeverity(common.AuditSeverityInfo),
		})
	}

	return nil
}

// EnableMFA enables multi-factor authentication for a user
func (a *AccessControlSystem) EnableMFA(ctx context.Context, userID string, method common.AuthMethod) error {
	// Get the user for logging
	user, err := a.authManager.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Enable MFA
	if err := a.authManager.EnableMFA(ctx, userID, method); err != nil {
		return err
	}

	// Log the action
	if a.securityManager != nil && user != nil {
		a.securityManager.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      getUserIDFromContext(ctx),
			Action:      AuditAction(common.AuditActionMfaEnable),
			Resource:    "user",
			ResourceID:  userID,
			Description: fmt.Sprintf("Enabled MFA (%s) for user %s", method, user.Username),
			Severity:    AuditSeverity(common.AuditSeverityInfo),
			Changes: map[string]interface{}{
				"mfa_enabled": true,
				"mfa_method":  string(method),
			},
		})
	}

	return nil
}

// DisableMFA disables multi-factor authentication for a user
func (a *AccessControlSystem) DisableMFA(ctx context.Context, userID string, method common.AuthMethod) error {
	// Get the user for logging
	user, err := a.authManager.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Disable MFA
	if err := a.authManager.DisableMFA(ctx, userID, method); err != nil {
		return err
	}

	// Log the action
	if a.securityManager != nil && user != nil {
		a.securityManager.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      getUserIDFromContext(ctx),
			Action:      AuditAction(common.AuditActionMfaDisable),
			Resource:    "user",
			ResourceID:  userID,
			Description: fmt.Sprintf("Disabled MFA (%s) for user %s", method, user.Username),
			Severity:    AuditSeverity(common.AuditSeverityInfo),
			Changes: map[string]interface{}{
				"mfa_enabled": false,
				"mfa_method":  string(method),
			},
		})
	}

	return nil
}

// Helper function to get changes between two users
func getChanges(oldUser, newUser *User) map[string]interface{} {
	changes := make(map[string]interface{})

	// Compare fields
	oldVal := reflect.ValueOf(oldUser).Elem()
	newVal := reflect.ValueOf(newUser).Elem()
	typeOfUser := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := typeOfUser.Field(i)
		fieldName := field.Name

		// Skip some fields
		if fieldName == "PasswordHash" || fieldName == "MFASecret" || fieldName == "UpdatedAt" {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		// Compare values
		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			changes[fieldName] = newField.Interface()
		}
	}

	return changes
}

// Close closes the access control system and releases any resources
func (a *AccessControlSystem) Close() error {
	// Close managers
	if err := a.authManager.Close(); err != nil {
		return err
	}

	return nil
}

// DefaultAccessControlConfig returns the default access control configuration
func DefaultAccessControlConfig() *AccessControlConfig {
	return DefaultAccessControlConfigV2()
}

// Helper function to get user ID from context
func getUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return ""
	}

	return userID
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	// Security incident management
	IncidentNotificationEmails  []string
	IncidentEscalationThreshold common.AuditSeverity
	IncidentAutoClose           time.Duration

	// Vulnerability management
	VulnerabilityCheckPeriod        time.Duration
	VulnerabilityNotificationEmails []string
}
