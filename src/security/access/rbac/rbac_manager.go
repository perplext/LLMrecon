// Package rbac provides enhanced role-based access control functionality
package rbac

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/security/access/audit"
	"github.com/perplext/LLMrecon/src/security/access/common"
)

// RBACManager manages role-based access control with enhanced features
type RBACManager struct {
	config          *RBACConfig
	roleStore       RoleStore
	permissionStore PermissionStore
	auditManager    *audit.AuditManager
	mu              sync.RWMutex
}

// RBACConfig defines configuration for the RBAC system
type RBACConfig struct {
	// Whether RBAC is enabled
	Enabled bool `json:"enabled"`
	
	// Whether to enforce strict role hierarchy
	StrictHierarchy bool `json:"strict_hierarchy"`
	
	// Whether to allow direct permission assignments to users
	AllowDirectPermissions bool `json:"allow_direct_permissions"`
	
	// Default roles to create
	DefaultRoles []Role `json:"default_roles"`
	
	// Maximum depth of role hierarchy
	MaxHierarchyDepth int `json:"max_hierarchy_depth"`
	
	// Whether to cache permission checks
	EnablePermissionCache bool `json:"enable_permission_cache"`
	
	// Permission cache TTL in seconds
	PermissionCacheTTL int `json:"permission_cache_ttl"`
	
	// Whether to automatically create missing permissions
	AutoCreatePermissions bool `json:"auto_create_permissions"`
	
	// Whether to log permission checks
	LogPermissionChecks bool `json:"log_permission_checks"`
	
	// Minimum severity for logging permission checks
	LogPermissionCheckSeverity common.AuditSeverity `json:"log_permission_check_severity"`

// NewRBACManager creates a new RBAC manager
func NewRBACManager(config *RBACConfig, roleStore RoleStore, permissionStore PermissionStore, auditManager *audit.AuditManager) (*RBACManager, error) {
	if config == nil {
		config = DefaultRBACConfig()
	}

	manager := &RBACManager{
		config:          config,
		roleStore:       roleStore,
		permissionStore: permissionStore,
		auditManager:    auditManager,
	}

	// Initialize default roles if configured
	if err := manager.initializeDefaultRoles(); err != nil {
		return nil, fmt.Errorf("failed to initialize default roles: %w", err)
	}

	return manager, nil

// initializeDefaultRoles initializes default roles
func (m *RBACManager) initializeDefaultRoles() error {
	// If no default roles are configured, use system defaults
	if len(m.config.DefaultRoles) == 0 {
		m.config.DefaultRoles = DefaultRoles()
	}

	// Create default roles
	for _, role := range m.config.DefaultRoles {
		// Check if role already exists
		exists, err := m.roleStore.RoleExists(context.Background(), role.ID)
		if err != nil {
			return fmt.Errorf("failed to check if role exists: %w", err)
		}
		if !exists {
			// Create the role
			if err := m.roleStore.CreateRole(context.Background(), &role); err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}
		}
	}

	return nil

// HasPermission checks if a user has a specific permission
func (m *RBACManager) HasPermission(ctx context.Context, userID string, permissionID string) (bool, error) {
	// If RBAC is disabled, allow all permissions
	if !m.config.Enabled {
		return true, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get user's direct permissions
	if m.config.AllowDirectPermissions {
		directPermissions, err := m.permissionStore.GetUserPermissions(ctx, userID)
		if err != nil {
			return false, fmt.Errorf("failed to get user permissions: %w", err)
		}

		// Check if the user has the permission directly
		for _, permission := range directPermissions {
			if permission.ID == permissionID {
				// Log permission check if enabled
				if m.config.LogPermissionChecks {
					m.logPermissionCheck(ctx, userID, permissionID, true, "direct permission")
				}
				return true, nil
			}
		}
	}
	// Get user's roles
	userRoles, err := m.roleStore.GetUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check if any of the user's roles have the permission
	for _, role := range userRoles {
		hasPermission, err := m.roleHasPermission(ctx, role.ID, permissionID, 0)
		if err != nil {
			return false, fmt.Errorf("failed to check if role has permission: %w", err)
		}

		if hasPermission {
			// Log permission check if enabled
			if m.config.LogPermissionChecks {
				m.logPermissionCheck(ctx, userID, permissionID, true, fmt.Sprintf("role %s", role.ID))
			}
			return true, nil
		}
	}

	// Log permission check if enabled
	if m.config.LogPermissionChecks {
		m.logPermissionCheck(ctx, userID, permissionID, false, "no matching role or permission")
	}

	return false, nil

// roleHasPermission checks if a role has a specific permission
func (m *RBACManager) roleHasPermission(ctx context.Context, roleID string, permissionID string, depth int) (bool, error) {
	// Check for max hierarchy depth to prevent infinite recursion
	if m.config.MaxHierarchyDepth > 0 && depth >= m.config.MaxHierarchyDepth {
		return false, nil
	}

	// Get role
	role, err := m.roleStore.GetRole(ctx, roleID)
	if err != nil {
		return false, fmt.Errorf("failed to get role: %w", err)
	}

	// Check if the role has the permission directly
	for _, permission := range role.Permissions {
		if permission == permissionID {
			return true, nil
		}
	}

	// Check parent roles if hierarchy is enabled
	for _, parentRoleID := range role.ParentRoles {
		hasPermission, err := m.roleHasPermission(ctx, parentRoleID, permissionID, depth+1)
		if err != nil {
			return false, fmt.Errorf("failed to check if parent role has permission: %w", err)
		}

		if hasPermission {
			return true, nil
		}
	}

	return false, nil

// logPermissionCheck logs a permission check
func (m *RBACManager) logPermissionCheck(ctx context.Context, userID string, permissionID string, granted bool, reason string) {
	status := "denied"
	if granted {
		status = "granted"
	}

	event := audit.NewAuditEvent(
		common.AuditActionResourceAccess,
		"permission",
		fmt.Sprintf("Permission check %s: %s", status, permissionID),
	).WithUserInfo(userID, "").
		WithResourceID(permissionID).
		WithStatus(status).
		WithSeverity(m.config.LogPermissionCheckSeverity).
		WithMetadata("reason", reason)

	// Log the event
	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Just log to stdout if audit logging fails
		fmt.Printf("Failed to log permission check: %v\n", err)
	}
// AssignRoleToUser assigns a role to a user
func (m *RBACManager) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	exists, err := m.roleStore.RoleExists(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("role %s does not exist", roleID)
	}

	// Assign role to user
	if err := m.roleStore.AssignRoleToUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	// Audit the action
	event := audit.NewAuditEvent(
		common.AuditActionRoleAssign,
		"role",
		fmt.Sprintf("Role %s assigned to user %s", roleID, userID),
	).WithUserInfo(userID, "").
		WithResourceID(roleID).
		WithStatus("success").
		WithSeverity(common.AuditSeverityInfo)

	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log role assignment: %v\n", err)
	}

	return nil

// RevokeRoleFromUser revokes a role from a user
func (m *RBACManager) RevokeRoleFromUser(ctx context.Context, userID string, roleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Revoke role from user
	if err := m.roleStore.RevokeRoleFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to revoke role from user: %w", err)
	}
	// Audit the action
	event := audit.NewAuditEvent(
		common.AuditActionRoleRevoke,
		"role",
		fmt.Sprintf("Role %s revoked from user %s", roleID, userID),
	).WithUserInfo(userID, "").
		WithResourceID(roleID).
		WithStatus("success").
		WithSeverity(common.AuditSeverityInfo)
	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log role revocation: %v\n", err)
	}

	return nil

// CreateRole creates a new role
func (m *RBACManager) CreateRole(ctx context.Context, role *Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role already exists
	exists, err := m.roleStore.RoleExists(ctx, role.ID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if exists {
		return fmt.Errorf("role %s already exists", role.ID)
	}
	// Validate role hierarchy
	if m.config.StrictHierarchy {
		for _, parentRoleID := range role.ParentRoles {
			exists, err := m.roleStore.RoleExists(ctx, parentRoleID)
			if err != nil {
				return fmt.Errorf("failed to check if parent role exists: %w", err)
			}

			if !exists {
				return fmt.Errorf("parent role %s does not exist", parentRoleID)
			}
		}
	}

	// Create the role
	if err := m.roleStore.CreateRole(ctx, role); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	// Audit the action
	event := audit.NewAuditEvent(
		common.AuditActionRoleCreate,
		"role",
		fmt.Sprintf("Role %s created", role.ID),
	).WithResourceID(role.ID).
		WithStatus("success").
		WithSeverity(common.AuditSeverityInfo).
		WithMetadata("role_name", role.Name).
		WithMetadata("role_description", role.Description)

	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log role creation: %v\n", err)
	}

	return nil

// UpdateRole updates an existing role
func (m *RBACManager) UpdateRole(ctx context.Context, role *Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	exists, err := m.roleStore.RoleExists(ctx, role.ID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("role %s does not exist", role.ID)
	}

	// Get the original role for audit logging
	originalRole, err := m.roleStore.GetRole(ctx, role.ID)
	if err != nil {
		return fmt.Errorf("failed to get original role: %w", err)
	}

	// Validate role hierarchy
	if m.config.StrictHierarchy {
		for _, parentRoleID := range role.ParentRoles {
			exists, err := m.roleStore.RoleExists(ctx, parentRoleID)
			if err != nil {
				return fmt.Errorf("failed to check if parent role exists: %w", err)
			}

			if !exists {
				return fmt.Errorf("parent role %s does not exist", parentRoleID)
			}
		}
	}

	// Update the role
	if err := m.roleStore.UpdateRole(ctx, role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Prepare changes for audit log
	changes := make(map[string]interface{})
	if originalRole.Name != role.Name {
		changes["name"] = map[string]interface{}{
			"old": originalRole.Name,
			"new": role.Name,
		}
	}
	if originalRole.Description != role.Description {
		changes["description"] = map[string]interface{}{
			"old": originalRole.Description,
			"new": role.Description,
		}
	}
	// Compare permissions
	addedPermissions, removedPermissions := diffStringSlices(originalRole.Permissions, role.Permissions)
	if len(addedPermissions) > 0 || len(removedPermissions) > 0 {
		changes["permissions"] = map[string]interface{}{
			"added":   addedPermissions,
			"removed": removedPermissions,
		}
	}
	// Compare parent roles
	addedParentRoles, removedParentRoles := diffStringSlices(originalRole.ParentRoles, role.ParentRoles)
	if len(addedParentRoles) > 0 || len(removedParentRoles) > 0 {
		changes["parent_roles"] = map[string]interface{}{
			"added":   addedParentRoles,
			"removed": removedParentRoles,
		}
	}

	// Audit the action
	event := audit.NewAuditEvent(
		common.AuditActionRoleUpdate,
		"role",
		fmt.Sprintf("Role %s updated", role.ID),
	).WithResourceID(role.ID).
		WithStatus("success").
		WithSeverity(common.AuditSeverityInfo).
		WithChanges(changes)

	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log role update: %v\n", err)
	}

	return nil

// DeleteRole deletes a role
func (m *RBACManager) DeleteRole(ctx context.Context, roleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	exists, err := m.roleStore.RoleExists(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("role %s does not exist", roleID)
	}

	// Get the role for audit logging
	role, err := m.roleStore.GetRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Delete the role
	if err := m.roleStore.DeleteRole(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Audit the action
	event := audit.NewAuditEvent(
		common.AuditActionRoleDelete,
		"role",
		fmt.Sprintf("Role %s deleted", roleID),
	).WithResourceID(roleID).
		WithStatus("success").
		WithSeverity(common.AuditSeverityInfo).
		WithMetadata("role_name", role.Name).
		WithMetadata("role_description", role.Description)

	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log role deletion: %v\n", err)
	}

	return nil

// GetRole gets a role by ID
func (m *RBACManager) GetRole(ctx context.Context, roleID string) (*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.roleStore.GetRole(ctx, roleID)

// ListRoles lists all roles
func (m *RBACManager) ListRoles(ctx context.Context) ([]*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.roleStore.ListRoles(ctx)

// GetUserRoles gets all roles assigned to a user
func (m *RBACManager) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.roleStore.GetUserRoles(ctx, userID)

// AddPermissionToRole adds a permission to a role
func (m *RBACManager) AddPermissionToRole(ctx context.Context, roleID string, permissionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	exists, err := m.roleStore.RoleExists(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("role %s does not exist", roleID)
	}

	// Check if permission exists
	exists, err = m.permissionStore.PermissionExists(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check if permission exists: %w", err)
	}

	if !exists {
		if m.config.AutoCreatePermissions {
			// Auto-create the permission
			permission := &Permission{
				ID:          permissionID,
				Name:        permissionID,
				Description: fmt.Sprintf("Auto-created permission: %s", permissionID),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := m.permissionStore.CreatePermission(ctx, permission); err != nil {
				return fmt.Errorf("failed to auto-create permission: %w", err)
			}
		} else {
			return fmt.Errorf("permission %s does not exist", permissionID)
		}
	}

	// Add permission to role
	if err := m.roleStore.AddPermissionToRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to add permission to role: %w", err)
	}

	// Audit the action
	event := audit.NewAuditEvent(
		common.AuditActionRolePermissionAdd,
		"role_permission",
		fmt.Sprintf("Permission %s added to role %s", permissionID, roleID),
	).WithResourceID(fmt.Sprintf("%s:%s", roleID, permissionID)).
		WithStatus("success").
		WithSeverity(common.AuditSeverityInfo)

	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log permission addition: %v\n", err)
	}

	return nil

// RemovePermissionFromRole removes a permission from a role
func (m *RBACManager) RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	exists, err := m.roleStore.RoleExists(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check if role exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("role %s does not exist", roleID)
	}
	// Remove permission from role
	if err := m.roleStore.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	// Audit the action
	event := audit.NewAuditEvent(
		common.AuditActionRolePermissionRemove,
		"role_permission",
		fmt.Sprintf("Permission %s removed from role %s", permissionID, roleID),
	).WithResourceID(fmt.Sprintf("%s:%s", roleID, permissionID)).
		WithStatus("success").
		WithSeverity(common.AuditSeverityInfo)

	if err := m.auditManager.LogAudit(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to log permission removal: %v\n", err)
	}

	return nil

// GetRolePermissions gets all permissions assigned to a role
func (m *RBACManager) GetRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get role
	role, err := m.roleStore.GetRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role.Permissions, nil

// GetUserPermissions gets all permissions assigned to a user (including from roles)
func (m *RBACManager) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get user's direct permissions
	var allPermissions []string
	if m.config.AllowDirectPermissions {
		directPermissions, err := m.permissionStore.GetUserPermissions(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user permissions: %w", err)
		}

		for _, permission := range directPermissions {
			allPermissions = append(allPermissions, permission.ID)
		}
	}

	// Get user's roles
	userRoles, err := m.roleStore.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Get permissions from roles
	rolePermissions := make(map[string]bool)
	for _, role := range userRoles {
		for _, permission := range role.Permissions {
			rolePermissions[permission] = true
		}

		// Get permissions from parent roles
		if err := m.addParentRolePermissions(ctx, role.ParentRoles, rolePermissions, 0); err != nil {
			return nil, fmt.Errorf("failed to get parent role permissions: %w", err)
		}
	}

	// Add role permissions to all permissions
	for permission := range rolePermissions {
		allPermissions = append(allPermissions, permission)
	}

	return allPermissions, nil

// addParentRolePermissions adds permissions from parent roles to the permissions map
func (m *RBACManager) addParentRolePermissions(ctx context.Context, parentRoleIDs []string, permissions map[string]bool, depth int) error {
	// Check for max hierarchy depth to prevent infinite recursion
	if m.config.MaxHierarchyDepth > 0 && depth >= m.config.MaxHierarchyDepth {
		return nil
	}

	for _, parentRoleID := range parentRoleIDs {
		// Get parent role
		parentRole, err := m.roleStore.GetRole(ctx, parentRoleID)
		if err != nil {
			return fmt.Errorf("failed to get parent role: %w", err)
		}

		// Add parent role permissions
		for _, permission := range parentRole.Permissions {
			permissions[permission] = true
		}

		// Add permissions from parent's parent roles
		if err := m.addParentRolePermissions(ctx, parentRole.ParentRoles, permissions, depth+1); err != nil {
			return fmt.Errorf("failed to get parent role permissions: %w", err)
		}
	}

	return nil

// DefaultRBACConfig returns the default RBAC configuration
func DefaultRBACConfig() *RBACConfig {
	return &RBACConfig{
		Enabled:                  true,
		StrictHierarchy:          true,
		AllowDirectPermissions:   true,
		MaxHierarchyDepth:        5,
		EnablePermissionCache:    true,
		PermissionCacheTTL:       300, // 5 minutes
		AutoCreatePermissions:    false,
		LogPermissionChecks:      true,
		LogPermissionCheckSeverity: common.AuditSeverityInfo,
	}

// diffStringSlices returns the elements that are in slice2 but not in slice1 (added),
// and the elements that are in slice1 but not in slice2 (removed)
func diffStringSlices(slice1, slice2 []string) (added, removed []string) {
	// Create maps for faster lookups
	map1 := make(map[string]bool)
	map2 := make(map[string]bool)

	for _, item := range slice1 {
		map1[item] = true
	}

	for _, item := range slice2 {
		map2[item] = true
	}

	// Find added items (in slice2 but not in slice1)
	for _, item := range slice2 {
		if !map1[item] {
			added = append(added, item)
		}
	}

	// Find removed items (in slice1 but not in slice2)
	for _, item := range slice1 {
		if !map2[item] {
			removed = append(removed, item)
		}
	}

}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
