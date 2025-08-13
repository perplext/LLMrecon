// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/perplext/LLMrecon/src/security/access/db/adapter"
)

// Error definitions
var (
	// ErrUserNotFound is already defined in auth.go
	// ErrInvalidCredentials is already defined in auth.go
	// ErrMFARequired is already defined in auth.go
	// ErrSessionExpired is already defined in auth.go
	
	// ErrPermissionDenied is returned when a user does not have the required permission
	ErrPermissionDenied = errors.New("permission denied")
)

// InitializeAccessControl initializes the access control system
func InitializeAccessControl(dataDir string) (*AccessControlSystem, error) {
	// Create the access control factory
	factory := NewAccessControlFactory(nil)

	// Determine if we should use in-memory or database-backed storage
	if dataDir == "" {
		// Use in-memory storage
		return factory.CreateInMemoryAccessControlSystem()
	}

	// Use database-backed storage
	dbPath := filepath.Join(dataDir, "access_control.db")
	
	system, err := factory.CreateDatabaseAccessControlSystem("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database-backed access control system: %w", err)
	}

	// Initialize the system
	if err := system.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize access control system: %w", err)
	}

	return system, nil
}

// CreateDefaultAdminUser creates a default admin user if no users exist
func CreateDefaultAdminUser(system *AccessControlSystem, username, email, password string) error {
	// Check if any users exist
	ctx := context.Background()
	users, err := system.GetAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for existing users: %w", err)
	}

	// If users exist, don't create a default admin
	if len(users) > 0 {
		return nil
	}

	// Create the admin user
	_, err = system.CreateUser(ctx, username, email, password, []string{RoleAdmin})
	if err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	return nil
}

// HasPermission checks if a user has the specified permission
func HasPermission(ctx context.Context, system *AccessControlSystem, userID string, permission string) (bool, error) {
	// Get the user
	user, err := system.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if the user has the permission directly
	for _, p := range user.Permissions {
		if p == permission {
			return true, nil
		}
	}

	// Check if the user has the permission through a role
	for _, role := range user.Roles {
		hasPermission := system.RBAC().RoleHasPermission(role, permission)
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// RequirePermission checks if a user has the specified permission and returns an error if not
func RequirePermission(ctx context.Context, system *AccessControlSystem, userID string, permission Permission) error {
	hasPermission, err := HasPermission(ctx, system, userID, permission)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}
	return nil
}
