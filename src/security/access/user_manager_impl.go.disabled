// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// UserManagerImpl implements the UserManager interface
type UserManagerImpl struct {
	mu          sync.RWMutex
	userStore   interfaces.UserStore
	auditLogger AuditLogger
	initialized bool
}

// NewUserManager creates a new user manager
func NewUserManager(userStore interfaces.UserStore, auditLogger AuditLogger) *UserManagerImpl {
	return &UserManagerImpl{
		userStore:   userStore,
		auditLogger: auditLogger,
	}
}

// UserManagerAdapter wraps UserManagerImpl to implement the UserManager interface
type UserManagerAdapter struct {
	impl *UserManagerImpl
}

// NewUserManagerAdapter creates a new user manager adapter
func NewUserManagerAdapter(impl *UserManagerImpl) *UserManagerAdapter {
	return &UserManagerAdapter{impl: impl}
}

// CreateUser creates a new user
func (a *UserManagerAdapter) CreateUser(user *models.User) error {
	return a.impl.CreateUser(context.Background(), user)
}

// GetUser retrieves a user by ID
func (a *UserManagerAdapter) GetUser(userID string) (*models.User, error) {
	return a.impl.GetUserByID(context.Background(), userID)
}

// GetUserByUsername retrieves a user by username
func (a *UserManagerAdapter) GetUserByUsername(username string) (*models.User, error) {
	return a.impl.GetUserByUsername(context.Background(), username)
}

// GetUserByEmail retrieves a user by email
func (a *UserManagerAdapter) GetUserByEmail(email string) (*models.User, error) {
	return a.impl.GetUserByEmail(context.Background(), email)
}

// UpdateUser updates a user
func (a *UserManagerAdapter) UpdateUser(user *models.User) error {
	return a.impl.UpdateUser(context.Background(), user)
}

// DeleteUser deletes a user
func (a *UserManagerAdapter) DeleteUser(userID string) error {
	return a.impl.DeleteUser(context.Background(), userID)
}

// ListUsers lists users
func (a *UserManagerAdapter) ListUsers(filter map[string]interface{}, offset, limit int) ([]*models.User, int, error) {
	return a.impl.ListUsers(context.Background(), filter, offset, limit)
}

// Close closes the user manager
func (a *UserManagerAdapter) Close() error {
	return a.impl.Close()
}

// Initialize initializes the user manager
func (m *UserManagerImpl) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = true
	return nil
}

// CreateUser creates a new user
func (m *UserManagerImpl) CreateUser(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure the user has an ID
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create the user
	err := m.userStore.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      models.AuditActionCreate,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "Created user: " + user.Username,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// GetUser retrieves a user by ID
func (m *UserManagerImpl) GetUser(ctx context.Context, userID string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "user",
		ResourceID:  userID,
		Description: "Retrieved user: " + user.Username,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (m *UserManagerImpl) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, err := m.userStore.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "Retrieved user by username: " + username,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (m *UserManagerImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, err := m.userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "Retrieved user by email: " + email,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return user, nil
}

// UpdateUser updates a user
func (m *UserManagerImpl) UpdateUser(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the existing user to compare changes
	existingUser, err := m.userStore.GetUserByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// Update the timestamp
	user.UpdatedAt = time.Now()
	user.CreatedAt = existingUser.CreatedAt // Preserve the creation timestamp

	// Update the user
	err = m.userStore.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      models.AuditActionUpdate,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "Updated user: " + user.Username,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// DeleteUser deletes a user
func (m *UserManagerImpl) DeleteUser(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the user to log the username
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Delete the user
	err = m.userStore.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      models.AuditActionDelete,
		Resource:    "user",
		ResourceID:  userID,
		Description: "Deleted user: " + user.Username,
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return nil
}

// ListUsers lists users
func (m *UserManagerImpl) ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.User, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if the store implements the ListUsersWithFilter method
	if store, ok := m.userStore.(interface {
		ListUsersWithFilter(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.User, int, error)
	}); ok {
		users, total, err := store.ListUsersWithFilter(ctx, filter, offset, limit)
		if err != nil {
			return nil, 0, err
		}

		// Log the audit event
		auditLog := &models.AuditLog{
			ID:          uuid.New().String(),
			UserID:      getUserIDFromContext(ctx),
			Action:      models.AuditActionRead,
			Resource:    "user",
			Description: "Listed users",
			Timestamp:   time.Now(),
		}
		
		_ = m.auditLogger.LogAudit(ctx, auditLog)

		return users, total, nil
	}

	// Fall back to the basic ListUsers method
	users, err := m.userStore.ListUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply filtering and pagination manually
	var filteredUsers []*models.User
	for _, user := range users {
		// In a real implementation, we would apply filters here
		filteredUsers = append(filteredUsers, user)
	}

	// Apply pagination
	total := len(filteredUsers)
	if offset >= total {
		return []*models.User{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// Log the audit event
	auditLog := &models.AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      models.AuditActionRead,
		Resource:    "user",
		Description: "Listed users",
		Timestamp:   time.Now(),
	}
	
	_ = m.auditLogger.LogAudit(ctx, auditLog)

	return filteredUsers[offset:end], total, nil
}

// Close closes the user manager
func (m *UserManagerImpl) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = false
	
	// Close the user store if it implements Close
	if closer, ok := m.userStore.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return errors.New("failed to close user store: " + err.Error())
		}
	}

	return nil
}
