// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// BasicInMemoryUserStore is an in-memory implementation of UserStore
type BasicInMemoryUserStore struct {
	users map[string]*User
	mu    sync.RWMutex
}

// NewBasicInMemoryUserStore creates a new basic in-memory user store
func NewBasicInMemoryUserStore() *BasicInMemoryUserStore {
	return &BasicInMemoryUserStore{
		users: make(map[string]*User),
	}
}

// CreateUser creates a new user
func (s *BasicInMemoryUserStore) CreateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user with the same ID already exists
	if _, exists := s.users[user.ID]; exists {
		return fmt.Errorf("user with ID %s already exists", user.ID)
	}

	// Check if user with the same username already exists
	for _, u := range s.users {
		if u.Username == user.Username {
			return fmt.Errorf("user with username %s already exists", user.Username)
		}
		if u.Email == user.Email {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	// Create a copy to prevent modification
	userCopy := *user
	s.users[user.ID] = &userCopy

	return nil
}

// GetUserByID retrieves a user by ID
func (s *BasicInMemoryUserStore) GetUserByID(ctx context.Context, userID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	// Return a copy to prevent modification
	userCopy := *user
	return &userCopy, nil
}

// GetUserByUsername retrieves a user by username
func (s *BasicInMemoryUserStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			// Return a copy to prevent modification
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, errors.New("user not found")
}

// GetUserByEmail retrieves a user by email
func (s *BasicInMemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			// Return a copy to prevent modification
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, errors.New("user not found")
}

// UpdateUser updates an existing user
func (s *BasicInMemoryUserStore) UpdateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	// Check if username is unique
	for id, u := range s.users {
		if id != user.ID && u.Username == user.Username {
			return fmt.Errorf("user with username %s already exists", user.Username)
		}
		if id != user.ID && u.Email == user.Email {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	// Update the user
	userCopy := *user
	s.users[user.ID] = &userCopy

	return nil
}

// DeleteUser deletes a user
func (s *BasicInMemoryUserStore) DeleteUser(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[userID]; !exists {
		return errors.New("user not found")
	}

	// Delete the user
	delete(s.users, userID)

	return nil
}

// ListUsers lists all users
func (s *BasicInMemoryUserStore) ListUsers(ctx context.Context) ([]*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		// Return copies to prevent modification
		userCopy := *user
		users = append(users, &userCopy)
	}

	return users, nil
}

// BasicInMemorySessionStore is an in-memory implementation of SessionStore
type BasicInMemorySessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewBasicInMemorySessionStore creates a new basic in-memory session store
func NewBasicInMemorySessionStore() *BasicInMemorySessionStore {
	return &BasicInMemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session
func (s *BasicInMemorySessionStore) CreateSession(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session with the same ID already exists
	if _, exists := s.sessions[session.ID]; exists {
		return fmt.Errorf("session with ID %s already exists", session.ID)
	}

	// Create a copy to prevent modification
	sessionCopy := *session
	s.sessions[session.ID] = &sessionCopy

	return nil
}

// GetSession retrieves a session by ID
func (s *BasicInMemorySessionStore) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Return a copy to prevent modification
	sessionCopy := *session
	return &sessionCopy, nil
}

// UpdateSession updates an existing session
func (s *BasicInMemorySessionStore) UpdateSession(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	if _, exists := s.sessions[session.ID]; !exists {
		return errors.New("session not found")
	}

	// Update the session
	sessionCopy := *session
	s.sessions[session.ID] = &sessionCopy

	return nil
}

// DeleteSession deletes a session
func (s *BasicInMemorySessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	if _, exists := s.sessions[sessionID]; !exists {
		return errors.New("session not found")
	}

	// Delete the session
	delete(s.sessions, sessionID)

	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (s *BasicInMemorySessionStore) DeleteUserSessions(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and delete all sessions for the user
	for id, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, id)
		}
	}

	return nil
}

// ListUserSessions lists all sessions for a user
func (s *BasicInMemorySessionStore) ListUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0)
	for _, session := range s.sessions {
		if session.UserID == userID {
			// Return a copy to prevent modification
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// CleanupExpiredSessions cleans up expired sessions
func (s *BasicInMemorySessionStore) CleanupExpiredSessions(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if session.ExpiresAt.Before(now) {
			delete(s.sessions, id)
		}
	}

	return nil
}
