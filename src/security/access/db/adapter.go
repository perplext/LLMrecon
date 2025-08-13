// Package db provides database implementations for the access control system
package db

import (
	"context"

	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// ModelsUserStore defines the interface for user storage operations using models.User
type ModelsUserStore interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.User, int, error)
	Close() error
}

// UserStore defines the interface for user storage operations using common.User
type UserStore interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *common.User) error
	
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id string) (*common.User, error)
	
	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*common.User, error)
	
	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*common.User, error)
	
	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *common.User) error
	
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id string) error
	
	// ListUsers lists users with optional filtering
	ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*common.User, int, error)
	
	// Close closes the user store
	Close() error
}

// UserStoreAdapter adapts a models.User store to the UserStore interface
type UserStoreAdapter struct {
	store ModelsUserStore
}

// NewUserStoreAdapter creates a new adapter for models.User store
func NewUserStoreAdapter(store ModelsUserStore) UserStore {
	return &UserStoreAdapter{
		store: store,
	}
}

// convertModelsUserToCommonUser converts a models.User to a common.User
func convertModelsUserToCommonUser(user *models.User) *common.User {
	if user == nil {
		return nil
	}

	return &common.User{
		ID:                  user.ID,
		Username:            user.Username,
		Email:               user.Email,
		PasswordHash:        user.PasswordHash,
		Roles:               user.Roles,
		Permissions:         user.Permissions,
		MFAEnabled:          user.MFAEnabled,
		MFAMethod:           user.MFAMethod,
		MFAMethods:          user.MFAMethods,
		MFASecret:           user.MFASecret,
		LastLogin:           user.LastLogin,
		LastPasswordChange:  user.LastPasswordChange,
		FailedLoginAttempts: user.FailedLoginAttempts,
		Locked:              user.Locked,
		Active:              user.Active,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
		Metadata:            user.Metadata,
	}
}

// convertCommonUserToModelsUser converts a common.User to a models.User
func convertCommonUserToModelsUser(user *common.User) *models.User {
	if user == nil {
		return nil
	}

	return &models.User{
		ID:                  user.ID,
		Username:            user.Username,
		Email:               user.Email,
		PasswordHash:        user.PasswordHash,
		Roles:               user.Roles,
		Permissions:         user.Permissions,
		MFAEnabled:          user.MFAEnabled,
		MFAMethod:           user.MFAMethod,
		MFAMethods:          user.MFAMethods,
		MFASecret:           user.MFASecret,
		LastLogin:           user.LastLogin,
		LastPasswordChange:  user.LastPasswordChange,
		FailedLoginAttempts: user.FailedLoginAttempts,
		Locked:              user.Locked,
		Active:              user.Active,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
		Metadata:            user.Metadata,
	}
}

// convertModelsUsersToCommonUsers converts a slice of models.User to a slice of common.User
func convertModelsUsersToCommonUsers(users []*models.User) []*common.User {
	if users == nil {
		return nil
	}
	result := make([]*common.User, len(users))
	for i, user := range users {
		result[i] = convertModelsUserToCommonUser(user)
	}
	return result
}

// CreateUser creates a new user
func (a *UserStoreAdapter) CreateUser(ctx context.Context, user *common.User) error {
	return a.store.CreateUser(ctx, convertCommonUserToModelsUser(user))
}

// GetUserByID retrieves a user by ID
func (a *UserStoreAdapter) GetUserByID(ctx context.Context, id string) (*common.User, error) {
	user, err := a.store.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertModelsUserToCommonUser(user), nil
}

// GetUserByUsername retrieves a user by username
func (a *UserStoreAdapter) GetUserByUsername(ctx context.Context, username string) (*common.User, error) {
	user, err := a.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return convertModelsUserToCommonUser(user), nil
}

// GetUserByEmail retrieves a user by email
func (a *UserStoreAdapter) GetUserByEmail(ctx context.Context, email string) (*common.User, error) {
	user, err := a.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return convertModelsUserToCommonUser(user), nil
}

// UpdateUser updates an existing user
func (a *UserStoreAdapter) UpdateUser(ctx context.Context, user *common.User) error {
	return a.store.UpdateUser(ctx, convertCommonUserToModelsUser(user))
}

// DeleteUser deletes a user
func (a *UserStoreAdapter) DeleteUser(ctx context.Context, id string) error {
	return a.store.DeleteUser(ctx, id)
}

// ListUsers lists users with optional filtering
func (a *UserStoreAdapter) ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*common.User, int, error) {
	users, count, err := a.store.ListUsers(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return convertModelsUsersToCommonUsers(users), count, nil
}

// Close closes the user store
func (a *UserStoreAdapter) Close() error {
	return a.store.Close()
}