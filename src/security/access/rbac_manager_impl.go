// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// RBACManagerImpl implements the RBACManager interface
type RBACManagerImpl struct {
	mu          sync.RWMutex
	userManager UserManager
	auditLogger AuditLogger
	roleStore   RoleStore
	initialized bool
}

// RoleStore defines the interface for storing and retrieving roles and permissions
type RoleStore interface {
	// GetUserRoles gets a user's roles
	GetUserRoles(ctx context.Context, userID string) ([]string, error)

	// AddRoleToUser adds a role to a user
	AddRoleToUser(ctx context.Context, userID, role string) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, role string) error

	// GetRolePermissions gets the permissions for a role
	GetRolePermissions(ctx context.Context, role string) ([]string, error)

	// AddPermissionToRole adds a permission to a role
	AddPermissionToRole(ctx context.Context, role, permission string) error

	// RemovePermissionFromRole removes a permission from a role
	RemovePermissionFromRole(ctx context.Context, role, permission string) error

	// GetAllRoles gets all roles
	GetAllRoles(ctx context.Context) ([]string, error)

	// GetAllPermissions gets all permissions
	GetAllPermissions(ctx context.Context) ([]string, error)

// NewRBACManagerImpl creates a new RBAC manager implementation
func NewRBACManagerImpl(userManager UserManager, roleStore RoleStore, auditLogger AuditLogger) *RBACManagerImpl {
	return &RBACManagerImpl{
		userManager: userManager,
		roleStore:   roleStore,
		auditLogger: auditLogger,
	}

// Initialize initializes the RBAC manager
func (m *RBACManagerImpl) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = true
	return nil

// HasPermission checks if a user has a permission
func (m *RBACManagerImpl) HasPermission(ctx context.Context, userID string, permission string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get the user's roles
	roles, err := m.roleStore.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if the user has the permission through any of their roles
	for _, role := range roles {
		permissions, err := m.roleStore.GetRolePermissions(ctx, role)
		if err != nil {
			continue
		}

		for _, p := range permissions {
			if p == permission {
				return true, nil
			}
		}
	}

	return false, nil

// HasRole checks if a user has a role
func (m *RBACManagerImpl) HasRole(ctx context.Context, userID string, role string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get the user's roles
	roles, err := m.roleStore.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if the user has the role
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}

	return false, nil

// AddRoleToUser adds a role to a user
func (m *RBACManagerImpl) AddRoleToUser(ctx context.Context, userID string, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify the user exists
	user, err := m.userManager.GetUser(userID)
	if err != nil {
		return err
	}

	// Add the role to the user
	err = m.roleStore.AddRoleToUser(ctx, userID, role)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      getRBACUserIDFromContext(ctx),
		Action:      AuditActionUpdate,
		Resource:    "role",
		ResourceID:  userID,
		Description: "Added role " + role + " to user " + user.Username,
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil

// RemoveRoleFromUser removes a role from a user
func (m *RBACManagerImpl) RemoveRoleFromUser(ctx context.Context, userID string, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify the user exists
	user, err := m.userManager.GetUser(userID)
	if err != nil {
		return err
	}

	// Remove the role from the user
	err = m.roleStore.RemoveRoleFromUser(ctx, userID, role)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      getRBACUserIDFromContext(ctx),
		Action:      AuditActionUpdate,
		Resource:    "role",
		ResourceID:  userID,
		Description: "Removed role " + role + " from user " + user.Username,
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil

// GetUserRoles gets a user's roles
func (m *RBACManagerImpl) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify the user exists
	_, err := m.userManager.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// Get the user's roles
	roles, err := m.roleStore.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      getRBACUserIDFromContext(ctx),
		Action:      AuditActionRead,
		Resource:    "role",
		ResourceID:  userID,
		Description: "Retrieved roles for user",
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return roles, nil
// GetUserPermissions gets a user's permissions
func (m *RBACManagerImpl) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify the user exists
	_, err := m.userManager.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// Get the user's roles
	roles, err := m.roleStore.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get the permissions for each role
	permissionMap := make(map[string]bool)
	for _, role := range roles {
		permissions, err := m.roleStore.GetRolePermissions(ctx, role)
		if err != nil {
			continue
		}

		for _, permission := range permissions {
			permissionMap[permission] = true
		}
	}

	// Convert the map to a slice
	var permissions []string
	for permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	// Log the audit event
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      getRBACUserIDFromContext(ctx),
		Action:      AuditActionRead,
		Resource:    "permission",
		ResourceID:  userID,
		Description: "Retrieved permissions for user",
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return permissions, nil

// Close closes the RBAC manager
func (m *RBACManagerImpl) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = false
	return nil

// InMemoryRoleStore is an in-memory implementation of the RoleStore interface
type InMemoryRoleStore struct {
	mu              sync.RWMutex
	userRoles       map[string][]string
	rolePermissions map[string][]string
}

// NewInMemoryRoleStore creates a new in-memory role store
func NewInMemoryRoleStore() *InMemoryRoleStore {
	return &InMemoryRoleStore{
		userRoles:       make(map[string][]string),
		rolePermissions: make(map[string][]string),
	}

// GetUserRoles gets a user's roles
func (s *InMemoryRoleStore) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles, ok := s.userRoles[userID]
	if !ok {
		return []string{}, nil
	}

	return roles, nil

// AddRoleToUser adds a role to a user
func (s *InMemoryRoleStore) AddRoleToUser(ctx context.Context, userID, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user already has the role
	roles, ok := s.userRoles[userID]
	if ok {
		for _, r := range roles {
			if r == role {
				return nil
			}
		}
	}

	// Add the role to the user
	s.userRoles[userID] = append(s.userRoles[userID], role)
	return nil

// RemoveRoleFromUser removes a role from a user
func (s *InMemoryRoleStore) RemoveRoleFromUser(ctx context.Context, userID, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user has any roles
	roles, ok := s.userRoles[userID]
	if !ok {
		return nil
	}

	// Remove the role from the user
	var newRoles []string
	for _, r := range roles {
		if r != role {
			newRoles = append(newRoles, r)
		}
	}

	s.userRoles[userID] = newRoles
	return nil

// GetRolePermissions gets the permissions for a role
func (s *InMemoryRoleStore) GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permissions, ok := s.rolePermissions[role]
	if !ok {
		return []string{}, nil
	}

	return permissions, nil

// AddPermissionToRole adds a permission to a role
func (s *InMemoryRoleStore) AddPermissionToRole(ctx context.Context, role, permission string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the role already has the permission
	permissions, ok := s.rolePermissions[role]
	if ok {
		for _, p := range permissions {
			if p == permission {
				return nil
			}
		}
	}

	// Add the permission to the role
	s.rolePermissions[role] = append(s.rolePermissions[role], permission)
	return nil

// RemovePermissionFromRole removes a permission from a role
func (s *InMemoryRoleStore) RemovePermissionFromRole(ctx context.Context, role, permission string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the role has any permissions
	permissions, ok := s.rolePermissions[role]
	if !ok {
		return nil
	}

	// Remove the permission from the role
	var newPermissions []string
	for _, p := range permissions {
		if p != permission {
			newPermissions = append(newPermissions, p)
		}
	}

	s.rolePermissions[role] = newPermissions
	return nil

// GetAllRoles gets all roles
func (s *InMemoryRoleStore) GetAllRoles(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var roles []string
	for role := range s.rolePermissions {
		roles = append(roles, role)
	}

	return roles, nil

// GetAllPermissions gets all permissions
func (s *InMemoryRoleStore) GetAllPermissions(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permissionMap := make(map[string]bool)
	for _, permissions := range s.rolePermissions {
		for _, permission := range permissions {
			permissionMap[permission] = true
		}
	}

	var permissions []string
	for permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	return permissions, nil

// getRBACUserIDFromContext extracts the user ID from the context
func getRBACUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return "system"
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
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
