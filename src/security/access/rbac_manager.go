// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"fmt"
	"sync"

)

// TypesRBACManager handles role-based access control with types config
type TypesRBACManager struct {
	config *AccessControlConfig
	mu     sync.RWMutex
}

// NewTypesRBACManager creates a new RBAC manager with types config
func NewTypesRBACManager(config *AccessControlConfig) *TypesRBACManager {
	return &TypesRBACManager{
		config: config,
	}
}

// Initialize initializes the RBAC manager
func (m *TypesRBACManager) Initialize(ctx context.Context) error {
	// Nothing to initialize for now
	return nil
}

// GetRoles returns all available roles
func (m *TypesRBACManager) GetRoles(ctx context.Context) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config.RBACConfig.DefaultRoles
}

// GetRolePermissions returns the permissions for a role
func (m *TypesRBACManager) GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if role exists
	roleExists := false
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return nil, fmt.Errorf("role %s does not exist", role)
	}

	// Get permissions for the role
	permissions, ok := m.config.RBACConfig.RolePermissions[role]
	if !ok {
		return []string{}, nil
	}

	return permissions, nil
}

// RoleHasPermission checks if a role has a specific permission
func (m *TypesRBACManager) RoleHasPermission(ctx context.Context, role string, permission string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if role exists
	roleExists := false
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return false, fmt.Errorf("role %s does not exist", role)
	}

	// Get permissions for the role
	permissions, ok := m.config.RBACConfig.RolePermissions[role]
	if !ok {
		return false, nil
	}

	// Check if the role has the permission
	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// AddRole adds a new role
func (m *TypesRBACManager) AddRole(ctx context.Context, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role already exists
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r == role {
			return fmt.Errorf("role %s already exists", role)
		}
	}

	// Add the role
	m.config.RBACConfig.DefaultRoles = append(m.config.RBACConfig.DefaultRoles, role)
	m.config.RBACConfig.RolePermissions[role] = []string{}

	return nil
}

// RemoveRole removes a role
func (m *TypesRBACManager) RemoveRole(ctx context.Context, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	roleExists := false
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return fmt.Errorf("role %s does not exist", role)
	}

	// Remove the role from the list
	var newRoles []string
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r != role {
			newRoles = append(newRoles, r)
		}
	}
	m.config.RBACConfig.DefaultRoles = newRoles

	// Remove the role's permissions
	delete(m.config.RBACConfig.RolePermissions, role)

	return nil
}

// AddPermissionToRole adds a permission to a role
func (m *TypesRBACManager) AddPermissionToRole(ctx context.Context, role string, permission string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	roleExists := false
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return fmt.Errorf("role %s does not exist", role)
	}

	// Get permissions for the role
	permissions, ok := m.config.RBACConfig.RolePermissions[role]
	if !ok {
		permissions = []string{}
	}

	// Check if permission already exists
	for _, p := range permissions {
		if p == permission {
			return fmt.Errorf("permission %s already exists for role %s", permission, role)
		}
	}

	// Add the permission
	permissions = append(permissions, permission)
	m.config.RBACConfig.RolePermissions[role] = permissions

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (m *TypesRBACManager) RemovePermissionFromRole(ctx context.Context, role string, permission string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	roleExists := false
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return fmt.Errorf("role %s does not exist", role)
	}

	// Get permissions for the role
	permissions, ok := m.config.RBACConfig.RolePermissions[role]
	if !ok {
		return fmt.Errorf("role %s has no permissions", role)
	}

	// Check if permission exists
	permissionExists := false
	for _, p := range permissions {
		if p == permission {
			permissionExists = true
			break
		}
	}
	if !permissionExists {
		return fmt.Errorf("permission %s does not exist for role %s", permission, role)
	}

	// Remove the permission
	var newPermissions []string
	for _, p := range permissions {
		if p != permission {
			newPermissions = append(newPermissions, p)
		}
	}
	m.config.RBACConfig.RolePermissions[role] = newPermissions

	return nil
}

// UserHasPermission checks if a user has a specific permission
func (m *TypesRBACManager) UserHasPermission(ctx context.Context, user *User, permission string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if user has the permission directly
	for _, p := range user.Permissions {
		if p == permission {
			return true, nil
		}
	}

	// Check if user has the permission through a role
	for _, role := range user.Roles {
		hasPermission, err := m.RoleHasPermission(ctx, role, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// AddPermissionToUser adds a permission directly to a user
func (m *TypesRBACManager) AddPermissionToUser(ctx context.Context, user *User, permission string) error {
	// Check if user already has the permission
	for _, p := range user.Permissions {
		if p == permission {
			return fmt.Errorf("user already has permission %s", permission)
		}
	}

	// Add the permission
	user.Permissions = append(user.Permissions, permission)

	return nil
}

// RemovePermissionFromUser removes a permission directly from a user
func (m *TypesRBACManager) RemovePermissionFromUser(ctx context.Context, user *User, permission string) error {
	// Check if user has the permission
	permissionExists := false
	for _, p := range user.Permissions {
		if p == permission {
			permissionExists = true
			break
		}
	}
	if !permissionExists {
		return fmt.Errorf("user does not have permission %s", permission)
	}

	// Remove the permission
	var newPermissions []string
	for _, p := range user.Permissions {
		if p != permission {
			newPermissions = append(newPermissions, p)
		}
	}
	user.Permissions = newPermissions

	return nil
}

// AddRoleToUser adds a role to a user
func (m *TypesRBACManager) AddRoleToUser(ctx context.Context, user *User, role string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if role exists
	roleExists := false
	for _, r := range m.config.RBACConfig.DefaultRoles {
		if r == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return fmt.Errorf("role %s does not exist", role)
	}

	// Check if user already has the role
	for _, r := range user.Roles {
		if r == role {
			return fmt.Errorf("user already has role %s", role)
		}
	}

	// Add the role
	user.Roles = append(user.Roles, role)

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (m *TypesRBACManager) RemoveRoleFromUser(ctx context.Context, user *User, role string) error {
	// Check if user has the role
	roleExists := false
	for _, r := range user.Roles {
		if r == role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return fmt.Errorf("user does not have role %s", role)
	}

	// Remove the role
	var newRoles []string
	for _, r := range user.Roles {
		if r != role {
			newRoles = append(newRoles, r)
		}
	}
	user.Roles = newRoles

	return nil
}

// GetUserPermissions returns all permissions a user has
func (m *TypesRBACManager) GetUserPermissions(ctx context.Context, user *User) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Start with direct permissions
	permissions := make(map[string]bool)
	for _, p := range user.Permissions {
		permissions[p] = true
	}

	// Add permissions from roles
	for _, role := range user.Roles {
		rolePermissions, err := m.GetRolePermissions(ctx, role)
		if err != nil {
			return nil, err
		}
		for _, p := range rolePermissions {
			permissions[p] = true
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(permissions))
	for p := range permissions {
		result = append(result, p)
	}

	return result, nil
}
