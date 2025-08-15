// Package impl provides implementations of the security access interfaces
package impl

import (
	"context"
	"errors"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// UserStoreAdapter adapts a legacy user store to the interfaces.UserStore interface
type UserStoreAdapter struct {
	legacyStore interface{}
	converter   UserConverter

// UserConverter converts between legacy and new user models
type UserConverter interface {
	// ToModelUser converts a legacy user to a model user
	ToModelUser(legacyUser interface{}) (*models.User, error)
	
	// FromModelUser converts a model user to a legacy user
	FromModelUser(user *models.User) (interface{}, error)

// NewUserStoreAdapter creates a new legacy user store adapter
func NewUserStoreAdapter(legacyStore interface{}, converter UserConverter) interfaces.UserStore {
	return &UserStoreAdapter{
		legacyStore: legacyStore,
		converter:   converter,
	}

// CreateUser creates a new user
func (s *UserStoreAdapter) CreateUser(ctx context.Context, user *models.User) error {
	// Convert the user to a legacy user
	legacyUser, err := s.converter.FromModelUser(user)
	if err != nil {
		return err
	}
	
	// Call the legacy store's CreateUser method
	if store, ok := s.legacyStore.(interface {
		CreateUser(ctx context.Context, user interface{}) error
	}); ok {
		return store.CreateUser(ctx, legacyUser)
	}
	
	return errors.New("legacy store does not implement CreateUser")

// GetUserByID retrieves a user by ID
func (s *UserStoreAdapter) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	// Call the legacy store's GetUserByID method
	if store, ok := s.legacyStore.(interface {
		GetUserByID(ctx context.Context, id string) (interface{}, error)
	}); ok {
		legacyUser, err := store.GetUserByID(ctx, id)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy user to a model user
		return s.converter.ToModelUser(legacyUser)
	}
	
	return nil, errors.New("legacy store does not implement GetUserByID")

// GetUserByUsername retrieves a user by username
func (s *UserStoreAdapter) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	// Call the legacy store's GetUserByUsername method
	if store, ok := s.legacyStore.(interface {
		GetUserByUsername(ctx context.Context, username string) (interface{}, error)
	}); ok {
		legacyUser, err := store.GetUserByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy user to a model user
		return s.converter.ToModelUser(legacyUser)
	}
	
	return nil, errors.New("legacy store does not implement GetUserByUsername")

// GetUserByEmail retrieves a user by email
func (s *UserStoreAdapter) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Call the legacy store's GetUserByEmail method
	if store, ok := s.legacyStore.(interface {
		GetUserByEmail(ctx context.Context, email string) (interface{}, error)
	}); ok {
		legacyUser, err := store.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy user to a model user
		return s.converter.ToModelUser(legacyUser)
	}
	
	return nil, errors.New("legacy store does not implement GetUserByEmail")

// UpdateUser updates an existing user
func (s *UserStoreAdapter) UpdateUser(ctx context.Context, user *models.User) error {
	// Convert the user to a legacy user
	legacyUser, err := s.converter.FromModelUser(user)
	if err != nil {
		return err
	}
	
	// Call the legacy store's UpdateUser method
	if store, ok := s.legacyStore.(interface {
		UpdateUser(ctx context.Context, user interface{}) error
	}); ok {
		return store.UpdateUser(ctx, legacyUser)
	}
	
	return errors.New("legacy store does not implement UpdateUser")

// DeleteUser deletes a user by ID
func (s *UserStoreAdapter) DeleteUser(ctx context.Context, id string) error {
	// Call the legacy store's DeleteUser method
	if store, ok := s.legacyStore.(interface {
		DeleteUser(ctx context.Context, id string) error
	}); ok {
		return store.DeleteUser(ctx, id)
	}
	
	return errors.New("legacy store does not implement DeleteUser")

// ListUsers lists all users
func (s *UserStoreAdapter) ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*interfaces.User, int, error) {
	// Call the legacy store's ListUsersWithFilter method
	if store, ok := s.legacyStore.(interface {
		ListUsersWithFilter(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]interface{}, int, error)
	}); ok {
		legacyUsers, total, err := store.ListUsersWithFilter(ctx, filter, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		
		// Convert the legacy users to interface users
		var users []*interfaces.User
		for _, legacyUser := range legacyUsers {
			modelUser, err := s.converter.ToModelUser(legacyUser)
			if err != nil {
				return nil, 0, err
			}
			// interfaces.User is an alias for models.User, so direct assignment works
			users = append(users, modelUser)
		}
		
		return users, total, nil
	}
	
	return nil, 0, errors.New("legacy store does not implement ListUsersWithFilter")

// ListUsersWithFilter lists users with optional filtering
func (s *UserStoreAdapter) ListUsersWithFilter(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*models.User, int, error) {
	// Call the legacy store's ListUsersWithFilter method
	if store, ok := s.legacyStore.(interface {
		ListUsersWithFilter(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]interface{}, int, error)
	}); ok {
		legacyUsers, total, err := store.ListUsersWithFilter(ctx, filter, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		
		// Convert the legacy users to model users
		var users []*models.User
		for _, legacyUser := range legacyUsers {
			user, err := s.converter.ToModelUser(legacyUser)
			if err != nil {
				return nil, 0, err
			}
			users = append(users, user)
		}
		
		return users, total, nil
	}
	
	// Fall back to the basic ListUsers method
	users, total, err := s.ListUsers(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	
	// Return the users directly since they're already filtered by the method call
	return users, total, nil

// Close closes the user store
func (s *UserStoreAdapter) Close() error {
	// Call the legacy store's Close method
	if store, ok := s.legacyStore.(interface {
		Close() error
	}); ok {
		return store.Close()
	}
	
	return nil

// SessionStoreAdapter adapts a legacy session store to the interfaces.SessionStore interface
type SessionStoreAdapter struct {
	legacyStore interface{}
	converter   SessionConverter

// SessionConverter converts between legacy and new session models
type SessionConverter interface {
	// ToModelSession converts a legacy session to a model session
	ToModelSession(legacySession interface{}) (*models.Session, error)
	
	// FromModelSession converts a model session to a legacy session
	FromModelSession(session *models.Session) (interface{}, error)

// NewSessionStoreAdapter creates a new legacy session store adapter
func NewSessionStoreAdapter(legacyStore interface{}, converter SessionConverter) interfaces.SessionStore {
	return &SessionStoreAdapter{
		legacyStore: legacyStore,
		converter:   converter,
	}

// CreateSession creates a new session
func (s *SessionStoreAdapter) CreateSession(ctx context.Context, session *models.Session) error {
	// Convert the session to a legacy session
	legacySession, err := s.converter.FromModelSession(session)
	if err != nil {
		return err
	}
	
	// Call the legacy store's CreateSession method
	if store, ok := s.legacyStore.(interface {
		CreateSession(ctx context.Context, session interface{}) error
	}); ok {
		return store.CreateSession(ctx, legacySession)
	}
	
	return errors.New("legacy store does not implement CreateSession")

// GetSession retrieves a session by ID
func (s *SessionStoreAdapter) GetSession(ctx context.Context, id string) (*models.Session, error) {
	// Call the legacy store's GetSession method
	if store, ok := s.legacyStore.(interface {
		GetSession(ctx context.Context, id string) (interface{}, error)
	}); ok {
		legacySession, err := store.GetSession(ctx, id)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy session to a model session
		return s.converter.ToModelSession(legacySession)
	}
	
	return nil, errors.New("legacy store does not implement GetSession")

// GetSessionByToken retrieves a session by token
func (s *SessionStoreAdapter) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	// Call the legacy store's GetSessionByToken method
	if store, ok := s.legacyStore.(interface {
		GetSessionByToken(ctx context.Context, token string) (interface{}, error)
	}); ok {
		legacySession, err := store.GetSessionByToken(ctx, token)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy session to a model session
		return s.converter.ToModelSession(legacySession)
	}
	
	return nil, errors.New("legacy store does not implement GetSessionByToken")

// UpdateSession updates an existing session
func (s *SessionStoreAdapter) UpdateSession(ctx context.Context, session *models.Session) error {
	// Convert the session to a legacy session
	legacySession, err := s.converter.FromModelSession(session)
	if err != nil {
		return err
	}
	
	// Call the legacy store's UpdateSession method
	if store, ok := s.legacyStore.(interface {
		UpdateSession(ctx context.Context, session interface{}) error
	}); ok {
		return store.UpdateSession(ctx, legacySession)
	}
	
	return errors.New("legacy store does not implement UpdateSession")

// DeleteSession deletes a session by ID
func (s *SessionStoreAdapter) DeleteSession(ctx context.Context, id string) error {
	// Call the legacy store's DeleteSession method
	if store, ok := s.legacyStore.(interface {
		DeleteSession(ctx context.Context, id string) error
	}); ok {
		return store.DeleteSession(ctx, id)
	}
	
	return errors.New("legacy store does not implement DeleteSession")

// GetUserSessions retrieves all sessions for a user
func (s *SessionStoreAdapter) GetUserSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	// Call the legacy store's GetUserSessions method
	if store, ok := s.legacyStore.(interface {
		GetUserSessions(ctx context.Context, userID string) ([]interface{}, error)
	}); ok {
		legacySessions, err := store.GetUserSessions(ctx, userID)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy sessions to model sessions
		var sessions []*models.Session
		for _, legacySession := range legacySessions {
			session, err := s.converter.ToModelSession(legacySession)
			if err != nil {
				return nil, err
			}
			sessions = append(sessions, session)
		}
		
		return sessions, nil
	}
	
	return nil, errors.New("legacy store does not implement GetUserSessions")

// CleanExpiredSessions removes all expired sessions and returns the count
func (s *SessionStoreAdapter) CleanExpiredSessions(ctx context.Context) (int, error) {
	// Call the legacy store's CleanExpiredSessions method if it returns count
	if store, ok := s.legacyStore.(interface {
		CleanExpiredSessions(ctx context.Context) (int, error)
	}); ok {
		return store.CleanExpiredSessions(ctx)
	}
	
	// Fallback: try legacy method that returns only error
	if store, ok := s.legacyStore.(interface {
		CleanExpiredSessions(ctx context.Context) error
	}); ok {
		err := store.CleanExpiredSessions(ctx)
		if err != nil {
			return 0, err
		}
		// Return 0 as we don't know the actual count
		return 0, nil
	}
	
	return 0, errors.New("legacy store does not implement CleanExpiredSessions")

// DeleteSessionsByUserID deletes all sessions for a user
func (s *SessionStoreAdapter) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	// Call the legacy store's DeleteSessionsByUserID method
	if store, ok := s.legacyStore.(interface {
		DeleteSessionsByUserID(ctx context.Context, userID string) error
	}); ok {
		return store.DeleteSessionsByUserID(ctx, userID)
	}
	
	return errors.New("legacy store does not implement DeleteSessionsByUserID")

// GetSessionByID retrieves a session by ID  
func (s *SessionStoreAdapter) GetSessionByID(ctx context.Context, id string) (*models.Session, error) {
	// Call the legacy store's GetSessionByID method
	if store, ok := s.legacyStore.(interface {
		GetSessionByID(ctx context.Context, id string) (interface{}, error)
	}); ok {
		legacySession, err := store.GetSessionByID(ctx, id)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy session to a model session
		return s.converter.ToModelSession(legacySession)
	}
	
	return nil, errors.New("legacy store does not implement GetSessionByID")

// GetSessionByRefreshToken retrieves a session by refresh token
func (s *SessionStoreAdapter) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	// Call the legacy store's GetSessionByRefreshToken method
	if store, ok := s.legacyStore.(interface {
		GetSessionByRefreshToken(ctx context.Context, refreshToken string) (interface{}, error)
	}); ok {
		legacySession, err := store.GetSessionByRefreshToken(ctx, refreshToken)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy session to a model session
		return s.converter.ToModelSession(legacySession)
	}
	
	return nil, errors.New("legacy store does not implement GetSessionByRefreshToken")

// ListSessionsByUserID lists sessions for a user
func (s *SessionStoreAdapter) ListSessionsByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	// Call the legacy store's ListSessionsByUserID method
	if store, ok := s.legacyStore.(interface {
		ListSessionsByUserID(ctx context.Context, userID string) ([]interface{}, error)
	}); ok {
		legacySessions, err := store.ListSessionsByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		
		// Convert the legacy sessions to model sessions
		sessions := make([]*models.Session, len(legacySessions))
		for i, legacySession := range legacySessions {
			session, err := s.converter.ToModelSession(legacySession)
			if err != nil {
				return nil, err
			}
			sessions[i] = session
		}
		
		return sessions, nil
	}
	
	return nil, errors.New("legacy store does not implement ListSessionsByUserID")

// Close closes the session store
func (s *SessionStoreAdapter) Close() error {
	// Call the legacy store's Close method
	if store, ok := s.legacyStore.(interface {
		Close() error
	}); ok {
		return store.Close()
	}
	
