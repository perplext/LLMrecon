// Package adapter provides adapters between database interfaces and domain models
package adapter

import (
	"context"

	"github.com/perplext/LLMrecon/src/security/access/converters"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// UserStoreAdapter adapts between the interfaces.UserStore and domain types
type UserStoreAdapter struct {
	store interfaces.UserStore
}

// NewUserStoreAdapter creates a new user store adapter
func NewUserStoreAdapter(store interfaces.UserStore) *UserStoreAdapter {
	return &UserStoreAdapter{
		store: store,
	}

// Close closes the user store
func (a *UserStoreAdapter) Close() error {
	return a.store.Close()

// CreateUser creates a new user
func (a *UserStoreAdapter) CreateUser(ctx context.Context, user *models.User) error {
	// Convert domain user to interface user
	interfaceUser := converters.ModelUserToInterfaceUser(user)
	
	// Create user in store
	err := a.store.CreateUser(ctx, interfaceUser)
	if err != nil {
		return err
	}
	
	return nil

// GetUserByID retrieves a user by ID
func (a *UserStoreAdapter) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	// Get user from store
	interfaceUser, err := a.store.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Convert interface user to domain user
	user := converters.InterfaceUserToModelUser(interfaceUser)
	
	return user, nil

// GetUserByUsername retrieves a user by username
func (a *UserStoreAdapter) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	// Get user from store
	interfaceUser, err := a.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	
	// Convert interface user to domain user
	user := converters.InterfaceUserToModelUser(interfaceUser)
	
	return user, nil

// GetUserByEmail retrieves a user by email
func (a *UserStoreAdapter) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Get user from store
	interfaceUser, err := a.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	
	// Convert interface user to domain user
	user := converters.InterfaceUserToModelUser(interfaceUser)
	
	return user, nil

// UpdateUser updates an existing user
func (a *UserStoreAdapter) UpdateUser(ctx context.Context, user *models.User) error {
	// Convert domain user to interface user
	interfaceUser := converters.ModelUserToInterfaceUser(user)
	
	// Update user in store
	err := a.store.UpdateUser(ctx, interfaceUser)
	if err != nil {
		return err
	}
	
	return nil

// DeleteUser deletes a user by ID
func (a *UserStoreAdapter) DeleteUser(ctx context.Context, id string) error {
	// Delete user from store
	return a.store.DeleteUser(ctx, id)

// ListUsers lists all users
func (a *UserStoreAdapter) ListUsers(ctx context.Context) ([]*models.User, error) {
	// List users from store
	interfaceUsers, _, err := a.store.ListUsers(ctx, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	
	// Convert interface users to domain users
	users := make([]*models.User, len(interfaceUsers))
	for i, interfaceUser := range interfaceUsers {
		users[i] = converters.InterfaceUserToModelUser(interfaceUser)
	}
	
}
}
}
}
}
}
}
}
