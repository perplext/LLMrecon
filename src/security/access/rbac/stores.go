// Package rbac provides enhanced role-based access control functionality
package rbac

import (
	"context"
	"fmt"
	"sync"
)

// RoleStore defines the interface for role storage
type RoleStore interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, role *Role) error
	
	// GetRole retrieves a role by ID
	GetRole(ctx context.Context, roleID string) (*Role, error)
	
	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, role *Role) error
	
	// DeleteRole deletes a role
	DeleteRole(ctx context.Context, roleID string) error
	
	// ListRoles lists all roles
	ListRoles(ctx context.Context) ([]*Role, error)
	
	// RoleExists checks if a role exists
	RoleExists(ctx context.Context, roleID string) (bool, error)
	
	// AssignRoleToUser assigns a role to a user
	AssignRoleToUser(ctx context.Context, userID string, roleID string) error
	
	// RevokeRoleFromUser revokes a role from a user
	RevokeRoleFromUser(ctx context.Context, userID string, roleID string) error
	
	// GetUserRoles gets all roles assigned to a user
	GetUserRoles(ctx context.Context, userID string) ([]*Role, error)
	
	// AddPermissionToRole adds a permission to a role
	AddPermissionToRole(ctx context.Context, roleID string, permissionID string) error
	
	// RemovePermissionFromRole removes a permission from a role
	RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error
}

// PermissionStore defines the interface for permission storage
type PermissionStore interface {
	// CreatePermission creates a new permission
	CreatePermission(ctx context.Context, permission *Permission) error
	
	// GetPermission retrieves a permission by ID
	GetPermission(ctx context.Context, permissionID string) (*Permission, error)
	
	// UpdatePermission updates an existing permission
	UpdatePermission(ctx context.Context, permission *Permission) error
	
	// DeletePermission deletes a permission
	DeletePermission(ctx context.Context, permissionID string) error
	
	// ListPermissions lists all permissions
	ListPermissions(ctx context.Context) ([]*Permission, error)
	
	// PermissionExists checks if a permission exists
	PermissionExists(ctx context.Context, permissionID string) (bool, error)
	
	// AssignPermissionToUser assigns a permission directly to a user
	AssignPermissionToUser(ctx context.Context, userID string, permissionID string) error
	
	// RevokePermissionFromUser revokes a permission from a user
	RevokePermissionFromUser(ctx context.Context, userID string, permissionID string) error
	
	// GetUserPermissions gets all permissions directly assigned to a user
	GetUserPermissions(ctx context.Context, userID string) ([]*Permission, error)
}

// InMemoryRoleStore is an in-memory implementation of RoleStore
type InMemoryRoleStore struct {
	roles     map[string]*Role
	userRoles map[string][]string // userID -> roleIDs
	mu        sync.RWMutex
}

// NewInMemoryRoleStore creates a new in-memory role store
func NewInMemoryRoleStore() *InMemoryRoleStore {
	return &InMemoryRoleStore{
		roles:     make(map[string]*Role),
		userRoles: make(map[string][]string),
	}
}

// CreateRole creates a new role
func (s *InMemoryRoleStore) CreateRole(ctx context.Context, role *Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role already exists
	if _, exists := s.roles[role.ID]; exists {
		return fmt.Errorf("role with ID %s already exists", role.ID)
	}

	// Create a copy of the role to prevent modification
	roleCopy := *role
	s.roles[role.ID] = &roleCopy

	return nil
}

// GetRole retrieves a role by ID
func (s *InMemoryRoleStore) GetRole(ctx context.Context, roleID string) (*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role with ID %s not found", roleID)
	}

	// Return a copy to prevent modification
	roleCopy := *role
	return &roleCopy, nil
}

// UpdateRole updates an existing role
func (s *InMemoryRoleStore) UpdateRole(ctx context.Context, role *Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role exists
	if _, exists := s.roles[role.ID]; !exists {
		return fmt.Errorf("role with ID %s not found", role.ID)
	}

	// Update the role
	roleCopy := *role
	roleCopy.UpdatedAt = time.Now()
	s.roles[role.ID] = &roleCopy

	return nil
}

// DeleteRole deletes a role
func (s *InMemoryRoleStore) DeleteRole(ctx context.Context, roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role exists
	if _, exists := s.roles[roleID]; !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// Check if the role is a system role
	if s.roles[roleID].SystemRole {
		return fmt.Errorf("cannot delete system role %s", roleID)
	}

	// Delete the role
	delete(s.roles, roleID)

	// Remove the role from all users
	for userID, roles := range s.userRoles {
		newRoles := make([]string, 0, len(roles))
		for _, id := range roles {
			if id != roleID {
				newRoles = append(newRoles, id)
			}
		}
		s.userRoles[userID] = newRoles
	}

	// Remove the role from parent roles of other roles
	for id, role := range s.roles {
		newParentRoles := make([]string, 0, len(role.ParentRoles))
		for _, parentID := range role.ParentRoles {
			if parentID != roleID {
				newParentRoles = append(newParentRoles, parentID)
			}
		}
		s.roles[id].ParentRoles = newParentRoles
	}

	return nil
}

// ListRoles lists all roles
func (s *InMemoryRoleStore) ListRoles(ctx context.Context) ([]*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]*Role, 0, len(s.roles))
	for _, role := range s.roles {
		// Return copies to prevent modification
		roleCopy := *role
		roles = append(roles, &roleCopy)
	}

	return roles, nil
}

// RoleExists checks if a role exists
func (s *InMemoryRoleStore) RoleExists(ctx context.Context, roleID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.roles[roleID]
	return exists, nil
}

// AssignRoleToUser assigns a role to a user
func (s *InMemoryRoleStore) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role exists
	if _, exists := s.roles[roleID]; !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// Check if user already has the role
	roles, exists := s.userRoles[userID]
	if exists {
		for _, id := range roles {
			if id == roleID {
				return nil // User already has the role
			}
		}
		// Add the role
		s.userRoles[userID] = append(roles, roleID)
	} else {
		// Create new role list for user
		s.userRoles[userID] = []string{roleID}
	}

	return nil
}

// RevokeRoleFromUser revokes a role from a user
func (s *InMemoryRoleStore) RevokeRoleFromUser(ctx context.Context, userID string, roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user has any roles
	roles, exists := s.userRoles[userID]
	if !exists {
		return nil // User has no roles
	}

	// Remove the role
	newRoles := make([]string, 0, len(roles))
	for _, id := range roles {
		if id != roleID {
			newRoles = append(newRoles, id)
		}
	}
	s.userRoles[userID] = newRoles

	return nil
}

// GetUserRoles gets all roles assigned to a user
func (s *InMemoryRoleStore) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get role IDs for user
	roleIDs, exists := s.userRoles[userID]
	if !exists {
		return []*Role{}, nil // User has no roles
	}

	// Get role objects
	roles := make([]*Role, 0, len(roleIDs))
	for _, id := range roleIDs {
		role, exists := s.roles[id]
		if exists {
			// Return a copy to prevent modification
			roleCopy := *role
			roles = append(roles, &roleCopy)
		}
	}

	return roles, nil
}

// AddPermissionToRole adds a permission to a role
func (s *InMemoryRoleStore) AddPermissionToRole(ctx context.Context, roleID string, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role exists
	role, exists := s.roles[roleID]
	if !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// Check if role already has the permission
	for _, id := range role.Permissions {
		if id == permissionID {
			return nil // Role already has the permission
		}
	}

	// Add the permission
	role.Permissions = append(role.Permissions, permissionID)
	role.UpdatedAt = time.Now()

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *InMemoryRoleStore) RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role exists
	role, exists := s.roles[roleID]
	if !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// Remove the permission
	newPermissions := make([]string, 0, len(role.Permissions))
	for _, id := range role.Permissions {
		if id != permissionID {
			newPermissions = append(newPermissions, id)
		}
	}
	role.Permissions = newPermissions
	role.UpdatedAt = time.Now()

	return nil
}

// InMemoryPermissionStore is an in-memory implementation of PermissionStore
type InMemoryPermissionStore struct {
	permissions     map[string]*Permission
	userPermissions map[string][]string // userID -> permissionIDs
	mu              sync.RWMutex
}

// NewInMemoryPermissionStore creates a new in-memory permission store
func NewInMemoryPermissionStore() *InMemoryPermissionStore {
	return &InMemoryPermissionStore{
		permissions:     make(map[string]*Permission),
		userPermissions: make(map[string][]string),
	}
}

// CreatePermission creates a new permission
func (s *InMemoryPermissionStore) CreatePermission(ctx context.Context, permission *Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if permission already exists
	if _, exists := s.permissions[permission.ID]; exists {
		return fmt.Errorf("permission with ID %s already exists", permission.ID)
	}

	// Create a copy of the permission to prevent modification
	permissionCopy := *permission
	s.permissions[permission.ID] = &permissionCopy

	return nil
}

// GetPermission retrieves a permission by ID
func (s *InMemoryPermissionStore) GetPermission(ctx context.Context, permissionID string) (*Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permission, exists := s.permissions[permissionID]
	if !exists {
		return nil, fmt.Errorf("permission with ID %s not found", permissionID)
	}

	// Return a copy to prevent modification
	permissionCopy := *permission
	return &permissionCopy, nil
}

// UpdatePermission updates an existing permission
func (s *InMemoryPermissionStore) UpdatePermission(ctx context.Context, permission *Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if permission exists
	if _, exists := s.permissions[permission.ID]; !exists {
		return fmt.Errorf("permission with ID %s not found", permission.ID)
	}

	// Update the permission
	permissionCopy := *permission
	permissionCopy.UpdatedAt = time.Now()
	s.permissions[permission.ID] = &permissionCopy

	return nil
}

// DeletePermission deletes a permission
func (s *InMemoryPermissionStore) DeletePermission(ctx context.Context, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if permission exists
	if _, exists := s.permissions[permissionID]; !exists {
		return fmt.Errorf("permission with ID %s not found", permissionID)
	}

	// Check if the permission is a system permission
	if s.permissions[permissionID].SystemPermission {
		return fmt.Errorf("cannot delete system permission %s", permissionID)
	}

	// Delete the permission
	delete(s.permissions, permissionID)

	// Remove the permission from all users
	for userID, permissions := range s.userPermissions {
		newPermissions := make([]string, 0, len(permissions))
		for _, id := range permissions {
			if id != permissionID {
				newPermissions = append(newPermissions, id)
			}
		}
		s.userPermissions[userID] = newPermissions
	}

	return nil
}

// ListPermissions lists all permissions
func (s *InMemoryPermissionStore) ListPermissions(ctx context.Context) ([]*Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permissions := make([]*Permission, 0, len(s.permissions))
	for _, permission := range s.permissions {
		// Return copies to prevent modification
		permissionCopy := *permission
		permissions = append(permissions, &permissionCopy)
	}

	return permissions, nil
}

// PermissionExists checks if a permission exists
func (s *InMemoryPermissionStore) PermissionExists(ctx context.Context, permissionID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.permissions[permissionID]
	return exists, nil
}

// AssignPermissionToUser assigns a permission directly to a user
func (s *InMemoryPermissionStore) AssignPermissionToUser(ctx context.Context, userID string, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if permission exists
	if _, exists := s.permissions[permissionID]; !exists {
		return fmt.Errorf("permission with ID %s not found", permissionID)
	}

	// Check if user already has the permission
	permissions, exists := s.userPermissions[userID]
	if exists {
		for _, id := range permissions {
			if id == permissionID {
				return nil // User already has the permission
			}
		}
		// Add the permission
		s.userPermissions[userID] = append(permissions, permissionID)
	} else {
		// Create new permission list for user
		s.userPermissions[userID] = []string{permissionID}
	}

	return nil
}

// RevokePermissionFromUser revokes a permission from a user
func (s *InMemoryPermissionStore) RevokePermissionFromUser(ctx context.Context, userID string, permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user has any permissions
	permissions, exists := s.userPermissions[userID]
	if !exists {
		return nil // User has no permissions
	}

	// Remove the permission
	newPermissions := make([]string, 0, len(permissions))
	for _, id := range permissions {
		if id != permissionID {
			newPermissions = append(newPermissions, id)
		}
	}
	s.userPermissions[userID] = newPermissions

	return nil
}

// GetUserPermissions gets all permissions directly assigned to a user
func (s *InMemoryPermissionStore) GetUserPermissions(ctx context.Context, userID string) ([]*Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get permission IDs for user
	permissionIDs, exists := s.userPermissions[userID]
	if !exists {
		return []*Permission{}, nil // User has no permissions
	}

	// Get permission objects
	permissions := make([]*Permission, 0, len(permissionIDs))
	for _, id := range permissionIDs {
		permission, exists := s.permissions[id]
		if exists {
			// Return a copy to prevent modification
			permissionCopy := *permission
			permissions = append(permissions, &permissionCopy)
		}
	}

	return permissions, nil
}
