// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/mfa"
	"github.com/perplext/LLMrecon/src/security/access/models"
	"github.com/perplext/LLMrecon/src/security/access/rbac"
)

// AccessControlManager is the main entry point for the access control system
type AccessControlManager struct {
	config          *AccessControlConfig
	rbacManager     RBACManager
	authManager     *AuthManager
	sessionManager  *SessionManager
	userManager     UserManager
	auditManager    *AuditManager
	securityManager SecurityManager
	userStore       UserStore
	sessionStore    SessionStore
	auditLogger     AuditLogger
	incidentStore   IncidentStore
	vulnStore       VulnerabilityStore
	mu              sync.RWMutex
}

// NewAccessControlManager creates a new access control manager
func NewAccessControlManager(config *AccessControlConfig) (*AccessControlManager, error) {
	if config == nil {
		config = DefaultAccessControlConfig()
	}

	// Create stores
	userStore := NewInMemoryUserStore()
	sessionStore := NewInMemorySessionStore()

	// Create audit logger
	var auditLogger AuditLogger
	if config.AuditConfig.LogFilePath != "" {
		fileLogger := NewFileAuditLogger(config.AuditConfig.LogFilePath)
		memLogger := NewInMemoryAuditLogger()
		auditLogger = NewMultiAuditLogger(fileLogger, memLogger)
	} else {
		auditLogger = NewInMemoryAuditLogger()
	}

	// Create audit manager
	auditManager, err := NewAuditManager(auditLogger, &config.AuditConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit manager: %w", err)
	}

	// Create RBAC manager
	simpleRBACManager := NewRBACManager(config)
	rbacManager := NewRBACManagerAdapter(simpleRBACManager)

	// Create auth manager
	mfaManager := mfa.NewMockMFAManager()
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
	authManager, err := NewAuthManager(userStore, sessionStore, auditLogger, mfaManager, authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Create session manager
	sessionManager := NewSessionManager(sessionStore, &config.SessionPolicy, auditLogger)

	// Create user manager using stub implementation
	userManager := NewUserManager()

	// Create security incident and vulnerability stores
	incidentStore := NewLocalInMemoryIncidentStore()
	vulnStore := NewLocalInMemoryVulnerabilityStore()

	// Create security manager
	securityManagerImpl := NewSecurityManager(config, incidentStore, vulnStore, auditLogger)
	securityManager := NewSecurityManagerAdapter(securityManagerImpl)

	return &AccessControlManager{
		config:          config,
		rbacManager:     rbacManager,
		authManager:     authManager,
		sessionManager:  sessionManager,
		userManager:     userManager,
		auditManager:    auditManager,
		securityManager: securityManager,
		userStore:       userStore,
		sessionStore:    sessionStore,
		auditLogger:     auditLogger,
		incidentStore:   incidentStore,
		vulnStore:       vulnStore,
	}, nil
}

// GetRBACManager returns the RBAC manager
func (m *AccessControlManager) GetRBACManager() RBACManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rbacManager
}

// GetUserManager returns the user manager
func (m *AccessControlManager) GetUserManager() UserManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.userManager
}

// Initialize initializes the access control system
func (m *AccessControlManager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize default roles and permissions
	m.initializeDefaultRoles()

	// Create admin user if it doesn't exist
	if err := m.createAdminUser(ctx); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Log initialization
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		Action:      AuditActionSystem,
		Resource:    "access_control",
		Description: "Access control system initialized",
		Severity:    AuditSeverityInfo,
		Status:      "success",
	})

	return nil
}

// Close closes the access control system
func (m *AccessControlManager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop session manager
	m.sessionManager.Stop()

	// Close audit manager
	if err := m.auditManager.Close(); err != nil {
		return fmt.Errorf("failed to close audit manager: %w", err)
	}

	// Log shutdown
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		Action:      AuditActionSystem,
		Resource:    "access_control",
		Description: "Access control system shutdown",
		Severity:    AuditSeverityInfo,
		Status:      "success",
	})

	return nil
}

// Login authenticates a user and creates a new session
func (m *AccessControlManager) Login(ctx context.Context, username, password string, ipAddress, userAgent string) (*Session, error) {
	return m.authManager.Login(ctx, username, password, ipAddress, userAgent)
}

// Logout logs out a user by invalidating their session
func (m *AccessControlManager) Logout(ctx context.Context, sessionID string) error {
	return m.authManager.Logout(ctx, sessionID)
}

// ValidateSession validates a session and returns whether it's valid
func (m *AccessControlManager) ValidateSession(ctx context.Context, token string) (bool, error) {
	return m.authManager.VerifySession(ctx, token)
}

// RefreshSession refreshes a session and returns a new token
func (m *AccessControlManager) RefreshSession(ctx context.Context, refreshToken string) (*Session, error) {
	return m.authManager.RefreshSession(ctx, refreshToken)
}

// VerifyMFA verifies a multi-factor authentication code
func (m *AccessControlManager) VerifyMFA(ctx context.Context, sessionID, code string) error {
	return m.authManager.VerifyMFA(ctx, sessionID, code)
}

// HasPermission checks if a user has a specific permission
func (m *AccessControlManager) HasPermission(ctx context.Context, user *User, permission Permission) bool {
	hasPermission, err := m.rbacManager.HasPermission(user.ID, string(permission))
	if err != nil {
		return false
	}
	return hasPermission
}

// HasRole checks if a user has a specific role
func (m *AccessControlManager) HasRole(ctx context.Context, user *User, role rbac.Role) bool {
	hasRole, err := m.rbacManager.HasRole(user.ID, role.Name)
	if err != nil {
		return false
	}
	return hasRole
}

// CreateUser creates a new user
func (m *AccessControlManager) CreateUser(ctx context.Context, username, email, password string, roles []string, createdBy string) (*User, error) {
	// Create a models.User to pass to the user manager
	user := &models.User{
		Username: username,
		Email:    email,
		Roles:    roles,
		Active:   true,
	}

	err := m.userManager.CreateUser(user)
	if err != nil {
		return nil, err
	}

	// Convert back to local User type for return
	// This is a simplified conversion - in real implementation we'd need proper mapping
	return &User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
		Active:   user.Active,
	}, nil
}

// UpdateUser updates an existing user
func (m *AccessControlManager) UpdateUser(ctx context.Context, id, username, email string, roles []string, active bool, updatedBy string) (*User, error) {
	// Create a models.User to pass to the user manager
	user := &models.User{
		ID:       id,
		Username: username,
		Email:    email,
		Roles:    roles,
		Active:   active,
	}

	err := m.userManager.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	// Convert back to local User type for return
	return &User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
		Active:   user.Active,
	}, nil
}

// DeleteUser deletes a user
func (m *AccessControlManager) DeleteUser(ctx context.Context, id, deletedBy string) error {
	return m.userManager.DeleteUser(id)
}

// GetUser retrieves a user by ID
func (m *AccessControlManager) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := m.userManager.GetUser(id)
	if err != nil {
		return nil, err
	}

	// Convert models.User to local User type
	return &User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
		Active:   user.Active,
	}, nil
}

// GetUserByUsername retrieves a user by username
func (m *AccessControlManager) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := m.userManager.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Convert models.User to local User type
	return &User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
		Active:   user.Active,
	}, nil
}

// ListUsers lists all users
func (m *AccessControlManager) ListUsers(ctx context.Context) ([]*User, error) {
	users, _, err := m.userManager.ListUsers(nil, 0, 100) // Default pagination
	if err != nil {
		return nil, err
	}

	// Convert models.User slice to local User slice
	result := make([]*User, len(users))
	for i, user := range users {
		result[i] = &User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Roles:    user.Roles,
			Active:   user.Active,
		}
	}

	return result, nil
}

// ChangePassword changes a user's password
func (m *AccessControlManager) ChangePassword(ctx context.Context, id, currentPassword, newPassword, updatedBy string) error {
	// Use the auth manager for password operations
	return m.authManager.UpdateUserPassword(ctx, id, currentPassword, newPassword)
}

// ResetPassword resets a user's password (admin function)
func (m *AccessControlManager) ResetPassword(ctx context.Context, id, newPassword, resetBy string) error {
	// Use the auth manager for password operations
	return m.authManager.UpdateUserPassword(ctx, id, "", newPassword) // Admin reset, no current password needed
}

// LockUser locks a user account
func (m *AccessControlManager) LockUser(ctx context.Context, id, lockedBy string, reason string) error {
	// Get the user and update their locked status
	user, err := m.userManager.GetUser(id)
	if err != nil {
		return err
	}

	user.Locked = true
	return m.userManager.UpdateUser(user)
}

// UnlockUser unlocks a user account
func (m *AccessControlManager) UnlockUser(ctx context.Context, id, unlockedBy string) error {
	// Get the user and update their locked status
	user, err := m.userManager.GetUser(id)
	if err != nil {
		return err
	}

	user.Locked = false
	return m.userManager.UpdateUser(user)
}

// EnableMFA enables multi-factor authentication for a user
func (m *AccessControlManager) EnableMFA(ctx context.Context, id string, method common.AuthMethod, enabledBy string) error {
	// Use the auth manager for MFA operations
	return m.authManager.EnableMFA(ctx, id, method)
}

// DisableMFA disables multi-factor authentication for a user
func (m *AccessControlManager) DisableMFA(ctx context.Context, id string, method common.AuthMethod, disabledBy string) error {
	// Use the auth manager for MFA operations
	return m.authManager.DisableMFA(ctx, id, method)
}

// LogAudit logs an audit event
func (m *AccessControlManager) LogAudit(ctx context.Context, log *AuditLog) error {
	// Note: ProcessSecurityAuditLog is not part of the SecurityManager interface
	// This would need to be implemented separately if needed

	return m.auditLogger.LogAudit(ctx, log)
}

// QueryAuditLogs queries audit logs
func (m *AccessControlManager) QueryAuditLogs(ctx context.Context, filter *AuditLogFilter) ([]*AuditLog, error) {
	// Note: QueryAuditLogs is not available on AuditManager
	// This is a placeholder implementation
	return []*AuditLog{}, nil
}

// CreateIncident creates a new security incident
func (m *AccessControlManager) CreateIncident(ctx context.Context, title, description string, severity AuditSeverity, reportedBy string, auditLogIDs []string, metadata map[string]interface{}) (*SecurityIncident, error) {
	// Convert severity and use SecurityManager interface
	modelsSeverity := models.SecurityIncidentSeverity(severity)
	modelsIncident, err := m.securityManager.ReportIncident(title, description, modelsSeverity)
	if err != nil {
		return nil, err
	}

	// Convert back to local SecurityIncident type
	return &SecurityIncident{
		ID:          modelsIncident.ID,
		Title:       modelsIncident.Title,
		Description: modelsIncident.Description,
		Severity:    AuditSeverity(modelsIncident.Severity),
		Status:      IncidentStatus(modelsIncident.Status),
		CreatedAt:   modelsIncident.ReportedAt,
		ReportedBy:  modelsIncident.ReportedBy,
		AuditLogIDs: auditLogIDs,
		Metadata:    metadata,
	}, nil
}

// UpdateIncidentStatus updates the status of a security incident
func (m *AccessControlManager) UpdateIncidentStatus(ctx context.Context, id string, status IncidentStatus, assignedTo, updatedBy string) error {
	// Get the incident first
	modelsIncident, err := m.securityManager.GetIncident(id)
	if err != nil {
		return err
	}

	// Update the status
	modelsIncident.Status = models.SecurityIncidentStatus(status)

	// Update the incident
	return m.securityManager.UpdateIncident(modelsIncident)
}

// GetIncident retrieves a security incident by ID
func (m *AccessControlManager) GetIncident(ctx context.Context, id string) (*SecurityIncident, error) {
	modelsIncident, err := m.securityManager.GetIncident(id)
	if err != nil {
		return nil, err
	}

	// Convert to local SecurityIncident type
	return &SecurityIncident{
		ID:          modelsIncident.ID,
		Title:       modelsIncident.Title,
		Description: modelsIncident.Description,
		Severity:    AuditSeverity(modelsIncident.Severity),
		Status:      IncidentStatus(modelsIncident.Status),
		CreatedAt:   modelsIncident.ReportedAt,
		ReportedBy:  modelsIncident.ReportedBy,
	}, nil
}

// ListIncidents lists security incidents based on filters
func (m *AccessControlManager) ListIncidents(ctx context.Context, filter *IncidentFilter) ([]*SecurityIncident, error) {
	// Convert filter to map[string]interface{} format
	filterMap := make(map[string]interface{})
	if filter != nil {
		if filter.Severity != "" {
			filterMap["severity"] = filter.Severity
		}
		if filter.Status != "" {
			filterMap["status"] = filter.Status
		}
		if filter.AssigneeID != "" {
			filterMap["assignee_id"] = filter.AssigneeID
		}
	}

	// Call SecurityManager with default pagination
	modelsIncidents, _, err := m.securityManager.ListIncidents(filterMap, 0, 100)
	if err != nil {
		return nil, err
	}

	// Convert to local SecurityIncident types
	result := make([]*SecurityIncident, len(modelsIncidents))
	for i, modelsIncident := range modelsIncidents {
		result[i] = &SecurityIncident{
			ID:          modelsIncident.ID,
			Title:       modelsIncident.Title,
			Description: modelsIncident.Description,
			Severity:    AuditSeverity(modelsIncident.Severity),
			Status:      IncidentStatus(modelsIncident.Status),
			CreatedAt:   modelsIncident.ReportedAt,
			ReportedBy:  modelsIncident.ReportedBy,
		}
	}

	return result, nil
}

// CreateVulnerability creates a new vulnerability
func (m *AccessControlManager) CreateVulnerability(ctx context.Context, title, description string, severity AuditSeverity, affectedSystem, cve, reportedBy string, metadata map[string]interface{}) (*Vulnerability, error) {
	// Convert severity and use SecurityManager interface
	modelsSeverity := models.VulnerabilitySeverity(severity)
	modelsVuln, err := m.securityManager.ReportVulnerability(title, description, modelsSeverity)
	if err != nil {
		return nil, err
	}

	// Convert back to local Vulnerability type
	return &Vulnerability{
		ID:             modelsVuln.ID,
		Title:          modelsVuln.Title,
		Description:    modelsVuln.Description,
		Severity:       AuditSeverity(modelsVuln.Severity),
		Status:         VulnerabilityStatus(modelsVuln.Status),
		CreatedAt:      modelsVuln.ReportedAt,
		ReportedBy:     modelsVuln.ReportedBy,
		AffectedSystem: affectedSystem,
		CVE:            cve,
		Metadata:       metadata,
	}, nil
}

// UpdateVulnerabilityStatus updates the status of a vulnerability
func (m *AccessControlManager) UpdateVulnerabilityStatus(ctx context.Context, id string, status VulnerabilityStatus, assignedTo, remediationPlan, updatedBy string) error {
	// Get the vulnerability first
	modelsVuln, err := m.securityManager.GetVulnerability(id)
	if err != nil {
		return err
	}

	// Update the status
	modelsVuln.Status = models.VulnerabilityStatus(status)

	// Update the vulnerability
	return m.securityManager.UpdateVulnerability(modelsVuln)
}

// GetVulnerability retrieves a vulnerability by ID
func (m *AccessControlManager) GetVulnerability(ctx context.Context, id string) (*Vulnerability, error) {
	modelsVuln, err := m.securityManager.GetVulnerability(id)
	if err != nil {
		return nil, err
	}

	// Convert to local Vulnerability type
	return &Vulnerability{
		ID:          modelsVuln.ID,
		Title:       modelsVuln.Title,
		Description: modelsVuln.Description,
		Severity:    AuditSeverity(modelsVuln.Severity),
		Status:      VulnerabilityStatus(modelsVuln.Status),
		CreatedAt:   modelsVuln.ReportedAt,
		ReportedBy:  modelsVuln.ReportedBy,
	}, nil
}

// ListVulnerabilities lists vulnerabilities based on filters
func (m *AccessControlManager) ListVulnerabilities(ctx context.Context, filter *VulnerabilityFilter) ([]*Vulnerability, error) {
	// Convert filter to map[string]interface{} format
	filterMap := make(map[string]interface{})
	if filter != nil {
		if filter.Severity != "" {
			filterMap["severity"] = filter.Severity
		}
		if filter.Status != "" {
			filterMap["status"] = filter.Status
		}
		if filter.Component != "" {
			filterMap["component"] = filter.Component
		}
		if filter.CveID != "" {
			filterMap["cve_id"] = filter.CveID
		}
	}

	// Call SecurityManager with default pagination
	modelsVulns, _, err := m.securityManager.ListVulnerabilities(filterMap, 0, 100)
	if err != nil {
		return nil, err
	}

	// Convert to local Vulnerability types
	result := make([]*Vulnerability, len(modelsVulns))
	for i, modelsVuln := range modelsVulns {
		result[i] = &Vulnerability{
			ID:          modelsVuln.ID,
			Title:       modelsVuln.Title,
			Description: modelsVuln.Description,
			Severity:    AuditSeverity(modelsVuln.Severity),
			Status:      VulnerabilityStatus(modelsVuln.Status),
			CreatedAt:   modelsVuln.ReportedAt,
			ReportedBy:  modelsVuln.ReportedBy,
		}
	}

	return result, nil
}

// Private helper methods

// initializeDefaultRoles initializes the default roles and permissions
func (m *AccessControlManager) initializeDefaultRoles() {
	// Note: The RBACManager interface doesn't expose AddRole/AddPermission methods
	// Role and permission initialization would need to be handled elsewhere
	// or the interface would need to be extended to support these operations
	// For now, this is a placeholder that does nothing
}

// createAdminUser creates the default admin user if it doesn't exist
func (m *AccessControlManager) createAdminUser(ctx context.Context) error {
	// Check if admin user already exists
	_, err := m.userStore.GetUserByUsername(ctx, "admin")
	if err == nil {
		// Admin user already exists
		return nil
	} else if !errors.Is(err, ErrUserNotFound) {
		return err
	}

	// Create admin user
	adminUser := &User{
		ID:                  generateRandomID(),
		Username:            "admin",
		Email:               "admin@example.com",
		PasswordHash:        "admin", // This should be properly hashed in a real implementation
		Roles:               []string{RoleAdmin},
		MFAEnabled:          false,
		FailedLoginAttempts: 0,
		Locked:              false,
		LastPasswordChange:  time.Now(),
		Active:              true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Save admin user
	if err := m.userStore.CreateUser(ctx, adminUser); err != nil {
		return err
	}

	// Log admin user creation
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		Action:      AuditActionCreate,
		Resource:    "user",
		ResourceID:  adminUser.ID,
		Description: "Default admin user created",
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Metadata: map[string]interface{}{
			"username": adminUser.Username,
			"roles":    adminUser.Roles,
		},
	})

	return nil
}
