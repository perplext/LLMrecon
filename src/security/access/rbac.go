// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Common errors
var (
	// ErrUnauthorized is already declared in integration.go
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidPermission  = errors.New("invalid permission")
	ErrRoleAlreadyExists  = errors.New("role already exists")
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionNotFound = errors.New("permission not found")
)

// SimpleRBACManager manages role-based access control
type SimpleRBACManager struct {
	config          *AccessControlConfig
	rolePermissions map[string][]string
	mu              sync.RWMutex
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(config *AccessControlConfig) *SimpleRBACManager {
	return NewSimpleRBACManager(config)
}

// NewSimpleRBACManager creates a new simple RBAC manager
func NewSimpleRBACManager(config *AccessControlConfig) *SimpleRBACManager {
	manager := &SimpleRBACManager{
		config:          config,
		rolePermissions: make(map[string][]string),
	}

	// Initialize with default role permissions if provided
	if config != nil && config.RolePermissions != nil {
		manager.rolePermissions = config.RolePermissions
	} else {
		// Set up default role permissions
		manager.setupDefaultRolePermissions()
	}

	return manager
}

// setupDefaultRolePermissions sets up default role permissions
func (m *SimpleRBACManager) setupDefaultRolePermissions() {
	// Admin role has all permissions
	adminPermissions := []string{
		PermissionSystemAdmin,
		PermissionSystemConfig,
		PermissionSystemView,
		PermissionSystemMonitor,
		PermissionSystemAudit,
		PermissionUserCreate,
		PermissionUserRead,
		PermissionUserUpdate,
		PermissionUserDelete,
		PermissionUserRoleAssign,
		PermissionTemplateCreate,
		PermissionTemplateRead,
		PermissionTemplateUpdate,
		PermissionTemplateDelete,
		PermissionTemplateExecute,
		PermissionTemplateApprove,
		PermissionSecurityConfig,
		PermissionSecurityView,
		PermissionSecurityTest,
		PermissionSecurityIncident,
		PermissionSecurityAudit,
		PermissionReportCreate,
		PermissionReportRead,
		PermissionReportUpdate,
		PermissionReportDelete,
		PermissionReportExport,
	}
	m.rolePermissions[RoleAdmin] = adminPermissions

	// Manager role has most permissions except system admin
	managerPermissions := []string{
		PermissionSystemConfig,
		PermissionSystemView,
		PermissionSystemMonitor,
		PermissionSystemAudit,
		PermissionUserCreate,
		PermissionUserRead,
		PermissionUserUpdate,
		PermissionUserRoleAssign,
		PermissionTemplateCreate,
		PermissionTemplateRead,
		PermissionTemplateUpdate,
		PermissionTemplateDelete,
		PermissionTemplateExecute,
		PermissionTemplateApprove,
		PermissionSecurityView,
		PermissionSecurityTest,
		PermissionSecurityIncident,
		PermissionSecurityAudit,
		PermissionReportCreate,
		PermissionReportRead,
		PermissionReportUpdate,
		PermissionReportExport,
	}
	m.rolePermissions[RoleManager] = managerPermissions

	// Operator role has operational permissions
	operatorPermissions := []string{
		PermissionSystemView,
		PermissionSystemMonitor,
		PermissionUserRead,
		PermissionTemplateCreate,
		PermissionTemplateRead,
		PermissionTemplateUpdate,
		PermissionTemplateExecute,
		PermissionSecurityView,
		PermissionSecurityTest,
		PermissionReportCreate,
		PermissionReportRead,
		PermissionReportExport,
	}
	m.rolePermissions[RoleOperator] = operatorPermissions

	// Auditor role has read-only and audit permissions
	auditorPermissions := []string{
		PermissionSystemView,
		PermissionSystemMonitor,
		PermissionSystemAudit,
		PermissionUserRead,
		PermissionTemplateRead,
		PermissionSecurityView,
		PermissionSecurityAudit,
		PermissionReportRead,
		PermissionReportExport,
	}
	m.rolePermissions[RoleAuditor] = auditorPermissions

	// User role has basic permissions
	userPermissions := []string{
		PermissionTemplateRead,
		PermissionTemplateExecute,
		PermissionReportRead,
	}
	m.rolePermissions[RoleUser] = userPermissions

	// Guest role has minimal permissions
	guestPermissions := []string{
		PermissionTemplateRead,
	}
	m.rolePermissions[RoleGuest] = guestPermissions

	// Automation role has specific permissions for automated tasks
	automationPermissions := []string{
		PermissionTemplateExecute,
		PermissionReportCreate,
		PermissionReportRead,
		PermissionReportExport,
	}
	m.rolePermissions[RoleAutomation] = automationPermissions
}

// HasPermission checks if a user has a specific permission
func (m *SimpleRBACManager) HasPermission(user *User, permission string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If RBAC is disabled, allow all permissions
	if m.config != nil && !m.config.EnableRBAC {
		return true
	}

	// Check direct permissions first
	for _, p := range user.Permissions {
		if p == permission {
			return true
		}
	}

	// Check role-based permissions
	for _, role := range user.Roles {
		permissions, exists := m.rolePermissions[role]
		if !exists {
			continue
		}

		for _, p := range permissions {
			if p == permission {
				return true
			}
		}
	}

	return false
}

// HasPermissionWithContext checks if a user has a specific permission (with context)
func (m *SimpleRBACManager) HasPermissionWithContext(ctx context.Context, user *User, permission Permission) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If RBAC is disabled, allow all permissions
	if m.config != nil && !m.config.EnableRBAC {
		return true
	}

	// Check direct permissions first
	for _, p := range user.Permissions {
		if p == permission {
			return true
		}
	}

	// Check role-based permissions
	for _, role := range user.Roles {
		permissions, exists := m.rolePermissions[role]
		if !exists {
			continue
		}

		for _, p := range permissions {
			if p == permission {
				return true
			}
		}
	}

	return false
}

// HasRole checks if a user has a specific role
func (m *SimpleRBACManager) HasRole(user *User, role string) bool {
	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasRoleWithContext checks if a user has a specific role (with context)
func (m *SimpleRBACManager) HasRoleWithContext(ctx context.Context, user *User, role string) bool {
	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// AssignRole assigns a role to a user
func (m *SimpleRBACManager) AssignRole(ctx context.Context, user *User, role string) error {
	m.mu.RLock()
	_, exists := m.rolePermissions[role]
	m.mu.RUnlock()

	if !exists {
		return ErrInvalidRole
	}

	// Check if user already has this role
	for _, r := range user.Roles {
		if r == role {
			return nil // User already has this role
		}
	}

	// Add the role
	user.Roles = append(user.Roles, role)
	return nil
}

// RevokeRole revokes a role from a user
func (m *SimpleRBACManager) RevokeRole(ctx context.Context, user *User, role string) error {
	// Find and remove the role
	for i, r := range user.Roles {
		if r == role {
			// Remove the role by replacing it with the last element and truncating
			user.Roles[i] = user.Roles[len(user.Roles)-1]
			user.Roles = user.Roles[:len(user.Roles)-1]
			return nil
		}
	}

	return ErrRoleNotFound
}

// AssignPermission assigns a direct permission to a user
func (m *SimpleRBACManager) AssignPermission(ctx context.Context, user *User, permission Permission) error {
	// Check if user already has this permission
	for _, p := range user.Permissions {
		if p == permission {
			return nil // User already has this permission
		}
	}

	// Add the permission
	user.Permissions = append(user.Permissions, permission)
	return nil
}

// RevokePermission revokes a direct permission from a user
func (m *SimpleRBACManager) RevokePermission(ctx context.Context, user *User, permission Permission) error {
	// Find and remove the permission
	for i, p := range user.Permissions {
		if p == permission {
			// Remove the permission by replacing it with the last element and truncating
			user.Permissions[i] = user.Permissions[len(user.Permissions)-1]
			user.Permissions = user.Permissions[:len(user.Permissions)-1]
			return nil
		}
	}

	return ErrPermissionNotFound
}

// GetRolePermissions gets all permissions for a role
func (m *SimpleRBACManager) GetRolePermissions(ctx context.Context, role string) ([]Permission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	permissions, exists := m.rolePermissions[role]
	if !exists {
		return nil, ErrRoleNotFound
	}

	// Return a copy to prevent modification
	result := make([]Permission, len(permissions))
	copy(result, permissions)
	return result, nil
}

// GetUserPermissions gets all permissions for a user (both direct and role-based)
func (m *SimpleRBACManager) GetUserPermissions(ctx context.Context, user *User) []Permission {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Use a map to avoid duplicates
	permissionMap := make(map[Permission]bool)

	// Add direct permissions
	for _, p := range user.Permissions {
		permissionMap[p] = true
	}

	// Add role-based permissions
	for _, role := range user.Roles {
		permissions, exists := m.rolePermissions[role]
		if !exists {
			continue
		}

		for _, p := range permissions {
			permissionMap[p] = true
		}
	}

	// Convert map to slice
	result := make([]Permission, 0, len(permissionMap))
	for p := range permissionMap {
		result = append(result, p)
	}

	return result
}

// AddRolePermission adds a permission to a role
func (m *SimpleRBACManager) AddRolePermission(ctx context.Context, role string, permission Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	permissions, exists := m.rolePermissions[role]
	if !exists {
		m.rolePermissions[role] = []Permission{permission}
		return nil
	}

	// Check if the role already has this permission
	for _, p := range permissions {
		if p == permission {
			return nil // Role already has this permission
		}
	}

	// Add the permission
	m.rolePermissions[role] = append(permissions, permission)
	return nil
}

// RemoveRolePermission removes a permission from a role
func (m *SimpleRBACManager) RemoveRolePermission(ctx context.Context, role string, permission Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	permissions, exists := m.rolePermissions[role]
	if !exists {
		return ErrRoleNotFound
	}

	// Find and remove the permission
	for i, p := range permissions {
		if p == permission {
			// Remove the permission by replacing it with the last element and truncating
			permissions[i] = permissions[len(permissions)-1]
			m.rolePermissions[role] = permissions[:len(permissions)-1]
			return nil
		}
	}

	return ErrPermissionNotFound
}

// CreateRole creates a new role with the specified permissions
func (m *SimpleRBACManager) CreateRole(ctx context.Context, role string, permissions []Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rolePermissions[role]; exists {
		return ErrRoleAlreadyExists
	}

	// Create a copy of the permissions
	permissionsCopy := make([]Permission, len(permissions))
	copy(permissionsCopy, permissions)

	m.rolePermissions[role] = permissionsCopy
	return nil
}

// DeleteRole deletes a role
func (m *SimpleRBACManager) DeleteRole(ctx context.Context, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rolePermissions[role]; !exists {
		return ErrRoleNotFound
	}

	delete(m.rolePermissions, role)
	return nil
}

// GetAllRoles gets all defined roles
func (m *SimpleRBACManager) GetAllRoles(ctx context.Context) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	roles := make([]string, 0, len(m.rolePermissions))
	for role := range m.rolePermissions {
		roles = append(roles, role)
	}

	return roles
}

// AddRole adds a role with default permissions
func (m *SimpleRBACManager) AddRole(role string, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rolePermissions[role]; exists {
		return ErrRoleAlreadyExists
	}

	m.rolePermissions[role] = []Permission{}
	return nil
}

// AddPermission adds a permission to the system
func (m *SimpleRBACManager) AddPermission(permission Permission, description string) error {
	// This is a simple implementation since we don't have a separate permissions store
	// In a real system, you might want to store permissions in a database
	// For now, we just return success as permissions are defined as constants
	return nil
}

// AddPermissionToRole adds a permission to a role
func (m *SimpleRBACManager) AddPermissionToRole(role string, permission Permission) error {
	return m.AddRolePermission(context.Background(), role, permission)
}

// Authorize checks if a user is authorized to perform an action on a resource
func (m *SimpleRBACManager) Authorize(ctx context.Context, user *User, permission Permission) error {
	if !m.HasPermissionWithContext(ctx, user, permission) {
		return fmt.Errorf("%w: user %s does not have permission %s", ErrUnauthorized, user.Username, permission)
	}
	return nil
}

// RequirePermission is a middleware-style function that checks if a user has a permission
func (m *SimpleRBACManager) RequirePermission(permission Permission) func(ctx context.Context, user *User) error {
	return func(ctx context.Context, user *User) error {
		return m.Authorize(ctx, user, permission)
	}
}

// RequireRole is a middleware-style function that checks if a user has a role
func (m *SimpleRBACManager) RequireRole(role string) func(ctx context.Context, user *User) error {
	return func(ctx context.Context, user *User) error {
		if !m.HasRoleWithContext(ctx, user, role) {
			return fmt.Errorf("%w: user %s does not have role %s", ErrUnauthorized, user.Username, role)
		}
		return nil
	}
}

// RequireAnyRole is a middleware-style function that checks if a user has any of the specified roles
func (m *SimpleRBACManager) RequireAnyRole(roles ...string) func(ctx context.Context, user *User) error {
	return func(ctx context.Context, user *User) error {
		for _, role := range roles {
			if m.HasRoleWithContext(ctx, user, role) {
				return nil
			}
		}
		return fmt.Errorf("%w: user %s does not have any required roles", ErrUnauthorized, user.Username)
	}
}

// RequireAllRoles is a middleware-style function that checks if a user has all of the specified roles
func (m *SimpleRBACManager) RequireAllRoles(roles ...string) func(ctx context.Context, user *User) error {
	return func(ctx context.Context, user *User) error {
		for _, role := range roles {
			if !m.HasRoleWithContext(ctx, user, role) {
				return fmt.Errorf("%w: user %s does not have required role %s", ErrUnauthorized, user.Username, role)
			}
		}
		return nil
	}
}

// RequireAnyPermission is a middleware-style function that checks if a user has any of the specified permissions
func (m *SimpleRBACManager) RequireAnyPermission(permissions ...Permission) func(ctx context.Context, user *User) error {
	return func(ctx context.Context, user *User) error {
		for _, permission := range permissions {
			if m.HasPermissionWithContext(ctx, user, permission) {
				return nil
			}
		}
		return fmt.Errorf("%w: user %s does not have any required permissions", ErrUnauthorized, user.Username)
	}
}

// RequireAllPermissions is a middleware-style function that checks if a user has all of the specified permissions
func (m *SimpleRBACManager) RequireAllPermissions(permissions ...Permission) func(ctx context.Context, user *User) error {
	return func(ctx context.Context, user *User) error {
		for _, permission := range permissions {
			if !m.HasPermissionWithContext(ctx, user, permission) {
				return fmt.Errorf("%w: user %s does not have required permission %s", ErrUnauthorized, user.Username, permission)
			}
		}
		return nil
	}
}

// RBACManagerAdapter wraps SimpleRBACManager to implement the RBACManager interface
type RBACManagerAdapter struct {
	manager *SimpleRBACManager
}

// NewRBACManagerAdapter creates a new RBAC manager adapter
func NewRBACManagerAdapter(manager *SimpleRBACManager) *RBACManagerAdapter {
	return &RBACManagerAdapter{manager: manager}
}

// HasPermission checks if a user has a permission
func (a *RBACManagerAdapter) HasPermission(userID string, permission string) (bool, error) {
	// This is a simplified implementation
	// In a real implementation, we would look up the user and check permissions
	return false, nil
}

// HasRole checks if a user has a role
func (a *RBACManagerAdapter) HasRole(userID string, role string) (bool, error) {
	// This is a simplified implementation
	// In a real implementation, we would look up the user and check roles
	return false, nil
}

// AddRoleToUser adds a role to a user
func (a *RBACManagerAdapter) AddRoleToUser(userID string, role string) error {
	// This is a simplified implementation
	// In a real implementation, we would update the user's roles
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (a *RBACManagerAdapter) RemoveRoleFromUser(userID string, role string) error {
	// This is a simplified implementation  
	// In a real implementation, we would update the user's roles
	return nil
}

// GetUserRoles gets a user's roles
func (a *RBACManagerAdapter) GetUserRoles(userID string) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, we would look up the user's roles
	return []string{}, nil
}

// GetUserPermissions gets a user's permissions
func (a *RBACManagerAdapter) GetUserPermissions(userID string) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, we would look up the user's permissions
	return []string{}, nil
}

// Close closes the RBAC manager
func (a *RBACManagerAdapter) Close() error {
	return nil
}
