package access

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/impl"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// DBAccessControlConfig contains configuration for the database-backed access control manager
type DBAccessControlConfig struct {
	// Database configuration
	DBConfig *interfaces.DBConfig

	// AuthConfig contains authentication configuration
	AuthConfig *AuthConfig

	// RBACConfig contains role-based access control configuration
	RBACConfig *RBACConfig

	// SecurityConfig contains security configuration
	SecurityConfig *AccessControlConfig

	// Default admin user to create if no users exist
	DefaultAdminUsername string
	DefaultAdminPassword string
	DefaultAdminEmail    string
}

// DefaultDBAccessControlConfig returns a default database-backed access control configuration
func DefaultDBAccessControlConfig(dataDir string) *DBAccessControlConfig {
	dbPath := filepath.Join(dataDir, "access_control.db")
	
	return &DBAccessControlConfig{
		DBConfig:             &interfaces.DBConfig{Driver: "sqlite3", DSN: dbPath},
		AuthConfig:           DefaultAuthConfig(),
		RBACConfig:           DefaultRBACConfig(),
		SecurityConfig:       DefaultSecurityConfig(),
		DefaultAdminUsername: "admin",
		DefaultAdminPassword: "changeme",
		DefaultAdminEmail:    "admin@example.com",
	}
}

// DefaultRBACConfig returns default RBAC configuration
func DefaultRBACConfig() *RBACConfig {
	return &RBACConfig{
		DefaultRoles: []string{"admin", "user", "manager", "operator", "auditor", "guest"},
		RolePermissions: map[string][]string{
			"admin": {
				"system:admin",
				"system:config",
				"system:view",
				"user:create",
				"user:read",
				"user:update",
				"user:delete",
				"template:create",
				"template:read",
				"template:update",
				"template:delete",
				"template:execute",
				"security:config",
				"security:view",
				"report:create",
				"report:read",
				"report:export",
			},
			"user": {
				"template:read",
				"template:execute",
				"report:read",
			},
		},
		RoleHierarchy: map[string][]string{
			"admin":    {},
			"manager":  {"user"},
			"operator": {"user"},
			"auditor":  {"user"},
			"user":     {},
			"guest":    {},
		},
		CustomRoles: []Role{},
	}
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *AccessControlConfig {
	return &AccessControlConfig{
		EnableRBAC: true,
		EnableMFA:  false,
		PasswordPolicy: PasswordPolicy{
			MinLength:           8,
			RequireUppercase:    true,
			RequireLowercase:    true,
			RequireNumbers:      true,
			RequireSpecialChars: false,
			MaxAge:              90,
			PreventReuseCount:   5,
		},
		SessionPolicy: SessionPolicy{
			TokenExpiration:         60,
			InactivityTimeout:       30,
			EnforceIPBinding:        false,
			EnforceUserAgentBinding: false,
			CleanupInterval:         30,
		},
		// TODO: Implement account lockout policy
		// AccountLockout: AccountLockoutPolicy{
		//	MaxFailedAttempts: 5,
		//	LockoutDuration:   15,
		//	Enabled:           true,
		// },
	}
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		SessionTimeout:         60 * time.Minute,
		SessionMaxInactive:     30 * time.Minute,
		PasswordMinLength:      8,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireNumber:  true,
		PasswordRequireSpecial: false,
		PasswordMaxAge:         90 * 24 * time.Hour,
		MFAEnabled:             false,
		MFAMethods:             []string{"totp", "sms"},
	}
}

// userStoreAdapter adapts interfaces.UserStore to access.UserStore
type userStoreAdapter struct {
	store interfaces.UserStore
}

// convertInterfacesUserToAccessUser converts from interfaces.User to access.User
func convertInterfacesUserToAccessUser(src *interfaces.User) *User {
	if src == nil {
		return nil
	}

	// Create a deep copy with only the fields that exist in the access.User struct
	result := &User{
		ID:        src.ID,
		Username:  src.Username,
		Email:     src.Email,
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
	}

	// Skip copying Metadata field as it's not compatible

	return result
}

// convertAccessUserToInterfacesUser converts from access.User to interfaces.User
func convertAccessUserToInterfacesUser(src *User) *interfaces.User {
	if src == nil {
		return nil
	}

	// Create a deep copy with only the fields that exist in the interfaces.User struct
	result := &interfaces.User{
		ID:        src.ID,
		Username:  src.Username,
		Email:     src.Email,
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
	}

	// Skip copying Metadata field as it's not compatible

	return result
}

// CreateUser implements access.UserStore
func (a *userStoreAdapter) CreateUser(ctx context.Context, user *User) error {
	interfacesUser := convertAccessUserToInterfacesUser(user)
	return a.store.CreateUser(ctx, interfacesUser)
}

// GetUserByID implements access.UserStore
func (a *userStoreAdapter) GetUserByID(ctx context.Context, id string) (*User, error) {
	user, err := a.store.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertInterfacesUserToAccessUser(user), nil
}

// GetUserByUsername implements access.UserStore
func (a *userStoreAdapter) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := a.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return convertInterfacesUserToAccessUser(user), nil
}

// GetUserByEmail implements access.UserStore
func (a *userStoreAdapter) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := a.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return convertInterfacesUserToAccessUser(user), nil
}

// UpdateUser implements access.UserStore
func (a *userStoreAdapter) UpdateUser(ctx context.Context, user *User) error {
	interfacesUser := convertAccessUserToInterfacesUser(user)
	return a.store.UpdateUser(ctx, interfacesUser)
}

// DeleteUser implements access.UserStore
func (a *userStoreAdapter) DeleteUser(ctx context.Context, id string) error {
	return a.store.DeleteUser(ctx, id)
}

// ListUsers implements access.UserStore
func (a *userStoreAdapter) ListUsers(ctx context.Context) ([]*User, error) {
	users, _, err := a.store.ListUsers(ctx, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	
	// Convert the slice of *interfaces.User to []*User
	result := make([]*User, len(users))
	for i, user := range users {
		result[i] = convertInterfacesUserToAccessUser(user)
	}
	return result, nil
}

// Close implements access.UserStore
func (a *userStoreAdapter) Close() error {
	// The underlying store might not have a close method, so return nil
	return nil
}

// DBAccessControlManager is a database-backed implementation of the access control manager
type DBAccessControlManager struct {
	// Database factory
	factory interfaces.DBFactory

	// Stores
	userStore          UserStore
	sessionStore       SessionStore
	auditLogger        AuditLogger
	incidentStore      IncidentStore
	vulnerabilityStore VulnerabilityStore

	// Managers
	authManager      *AuthManager
	rbacManager      *SimpleRBACManager  // Changed to concrete type
	securityManager  *BasicSecurityManager  // Changed to concrete type
	sessionManager   *SessionManager
	userManager      *UserManager
	boundaryEnforcer *EnhancedContextBoundaryEnforcer

	// Configuration
	config *DBAccessControlConfig

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewDBAccessControlManager creates a new database-backed access control manager
func NewDBAccessControlManager(config *DBAccessControlConfig) (*DBAccessControlManager, error) {
	// Create database factory
	// Note: We need to implement a factory creator in the interfaces package
	// For now, we'll use a placeholder
	var factory interfaces.DBFactory
	var err error
	// TODO: Replace with proper factory implementation
	if err != nil {
		return nil, fmt.Errorf("failed to create database factory: %w", err)
	}

	// Create all stores
	err = factory.CreateAllStores()
	if err != nil {
		return nil, fmt.Errorf("failed to create stores: %w", err)
	}
	
	// Get individual stores
	interfaceUserStore, err := factory.GetUserStore()
	if err != nil {
		factory.Close()
		return nil, fmt.Errorf("failed to get user store: %w", err)
	}
	
	// Create adapters
	userStore := &userStoreAdapter{store: interfaceUserStore}
	
	// TODO: Fix interface mismatches before enabling these stores
	// interfaceSessionStore, err := factory.GetSessionStore()
	// if err != nil {
	//	factory.Close()
	//	return nil, fmt.Errorf("failed to get session store: %w", err)
	// }
	
	// Create session adapter with converter
	// sessionConverter := impl.NewModelConverter()
	// sessionStore := impl.NewSessionStoreAdapter(interfaceSessionStore, sessionConverter)
	
	// auditStore, err := factory.GetAuditStore()
	// if err != nil {
	//	factory.Close()
	//	return nil, fmt.Errorf("failed to get audit store: %w", err)
	// }
	// Create audit logger adapter with converter
	// auditConverter := impl.NewModelConverter()
	// auditLogger := impl.NewAuditLoggerAdapter(auditStore, auditConverter)
	
	// TODO: Fix interface mismatches before enabling these stores
	// incidentStore, err := factory.GetIncidentStore()
	// if err != nil {
	//	factory.Close()
	//	return nil, fmt.Errorf("failed to get incident store: %w", err)
	// }
	
	// vulnerabilityStore, err := factory.GetVulnerabilityStore()
	// if err != nil {
	//	factory.Close()
	//	return nil, fmt.Errorf("failed to get vulnerability store: %w", err)
	// }
	
	// Create the manager
	manager := &DBAccessControlManager{
		factory:            factory,
		userStore:          userStore,
		sessionStore:       nil, // TODO: Fix session store adapter
		auditLogger:        nil, // TODO: Fix audit logger adapter
		incidentStore:      nil, // TODO: Fix incident store adapter
		vulnerabilityStore: nil, // TODO: Fix vulnerability store adapter
		config:             config,
	}
	
	// TODO: Create managers once interfaces are properly aligned
	// For now, leaving managers as nil to avoid compilation errors
	
	// Create simple managers with minimal dependencies
	rbacManager := NewRBACManager(config.SecurityConfig)
	manager.rbacManager = rbacManager
	
	// Create a minimal security manager
	securityManager := NewSecurityManager(config.SecurityConfig, nil, nil, nil)
	manager.securityManager = securityManager
	
	// TODO: Initialize the manager once all dependencies are resolved
	// err = manager.initialize(context.Background())
	// if err != nil {
	//	factory.Close()
	//	return nil, fmt.Errorf("failed to initialize manager: %w", err)
	// }
	
	return manager, nil
}

// initialize initializes the access control system
func (m *DBAccessControlManager) initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// TODO: Initialize all managers once interfaces are aligned
	// err := m.authManager.Initialize(ctx)
	// if err != nil {
	//	return fmt.Errorf("failed to initialize auth manager: %w", err)
	// }
	
	// err = m.rbacManager.Initialize(ctx)
	// if err != nil {
	//	return fmt.Errorf("failed to initialize RBAC manager: %w", err)
	// }
	
	// err = m.securityManager.Initialize(ctx)
	// if err != nil {
	//	return fmt.Errorf("failed to initialize security manager: %w", err)
	// }
	
	// err = m.sessionManager.Initialize(ctx)
	// if err != nil {
	//	return fmt.Errorf("failed to initialize session manager: %w", err)
	// }
	
	// Create default admin user if no users exist
	users, err := m.userStore.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}
	
	if len(users) == 0 {
		// TODO: Create default admin user
	}
	
	return nil
}

// Close closes the access control manager and releases resources
func (m *DBAccessControlManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.factory.Close()
}

// GetAuthManager returns the authentication manager
func (m *DBAccessControlManager) GetAuthManager() *AuthManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.authManager
}

// GetRBACManager returns the role-based access control manager
func (m *DBAccessControlManager) GetRBACManager() *SimpleRBACManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rbacManager
}

// GetSecurityManager returns the security manager
func (m *DBAccessControlManager) GetSecurityManager() *BasicSecurityManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.securityManager
}

// GetSessionManager returns the session manager
func (m *DBAccessControlManager) GetSessionManager() *SessionManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionManager
}

// GetUserManager returns the user manager
func (m *DBAccessControlManager) GetUserManager() *UserManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.userManager
}

// GetBoundaryEnforcer returns the context boundary enforcer
func (m *DBAccessControlManager) GetBoundaryEnforcer() *EnhancedContextBoundaryEnforcer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.boundaryEnforcer
}

// GetAuditLogger returns the audit logger
func (m *DBAccessControlManager) GetAuditLogger() AuditLogger {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.auditLogger
}
