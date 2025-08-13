// Package adapter provides adapters between database interfaces and domain models
package adapter

import (
	"context"
)

// Session represents a user session
type Session struct {
	ID             string
	UserID         string
	Token          string
	RefreshToken   string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	LastActivityAt time.Time
	IPAddress      string
	UserAgent      string
	Metadata       string
}

// SessionStore defines the interface for session storage operations
type SessionStore interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *Session) error
	
	// GetSessionByID retrieves a session by ID
	GetSessionByID(ctx context.Context, id string) (*Session, error)
	
	// GetSessionByToken retrieves a session by token
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	
	// GetSessionByRefreshToken retrieves a session by refresh token
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	
	// UpdateSession updates an existing session
	UpdateSession(ctx context.Context, session *Session) error
	
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, id string) error
	
	// DeleteSessionsByUserID deletes all sessions for a user
	DeleteSessionsByUserID(ctx context.Context, userID string) error
	
	// ListSessionsByUserID lists sessions for a user
	ListSessionsByUserID(ctx context.Context, userID string) ([]*Session, error)
	
	// CleanExpiredSessions removes expired sessions
	CleanExpiredSessions(ctx context.Context) (int, error)
	
	// Close closes the session store
	Close() error
}
