// Package adapter provides adapters between database interfaces and domain models
package adapter

import (
	"context"
	"encoding/json"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// SessionStoreAdapter adapts between the adapter.SessionStore and interfaces.SessionStore
type SessionStoreAdapter struct {
	store SessionStore // This is adapter.SessionStore
}

// NewSessionStoreAdapter creates a new session store adapter
func NewSessionStoreAdapter(store SessionStore) *SessionStoreAdapter {
	return &SessionStoreAdapter{
		store: store,
	}

// Close closes the session store
func (a *SessionStoreAdapter) Close() error {
	return a.store.Close()

// CreateSession creates a new session
func (a *SessionStoreAdapter) CreateSession(ctx context.Context, session *interfaces.Session) error {
	// Convert interfaces.Session to adapter.Session
	adapterSession := &Session{
		ID:             session.ID,
		UserID:         session.UserID,
		Token:          session.Token,
		RefreshToken:   session.RefreshToken,
		ExpiresAt:      session.ExpiresAt,
		CreatedAt:      session.CreatedAt,
		LastActivityAt: session.LastActivity,
		IPAddress:      session.IPAddress,
		UserAgent:      session.UserAgent,
		Metadata:       convertMetadataToString(session.Metadata),
	}
	return a.store.CreateSession(ctx, adapterSession)

// GetSessionByID retrieves a session by ID
func (a *SessionStoreAdapter) GetSessionByID(ctx context.Context, id string) (*interfaces.Session, error) {
	adapterSession, err := a.store.GetSessionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertAdapterSessionToInterface(adapterSession), nil

// GetSessionByToken retrieves a session by token
func (a *SessionStoreAdapter) GetSessionByToken(ctx context.Context, token string) (*interfaces.Session, error) {
	adapterSession, err := a.store.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return convertAdapterSessionToInterface(adapterSession), nil

// GetSessionByRefreshToken retrieves a session by refresh token
func (a *SessionStoreAdapter) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*interfaces.Session, error) {
	adapterSession, err := a.store.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	return convertAdapterSessionToInterface(adapterSession), nil

// UpdateSession updates an existing session
func (a *SessionStoreAdapter) UpdateSession(ctx context.Context, session *interfaces.Session) error {
	// Convert interfaces.Session to adapter.Session
	adapterSession := &Session{
		ID:             session.ID,
		UserID:         session.UserID,
		Token:          session.Token,
		RefreshToken:   session.RefreshToken,
		ExpiresAt:      session.ExpiresAt,
		CreatedAt:      session.CreatedAt,
		LastActivityAt: session.LastActivity,
		IPAddress:      session.IPAddress,
		UserAgent:      session.UserAgent,
		Metadata:       convertMetadataToString(session.Metadata),
	}
	return a.store.UpdateSession(ctx, adapterSession)

// DeleteSession deletes a session by ID
func (a *SessionStoreAdapter) DeleteSession(ctx context.Context, id string) error {
	return a.store.DeleteSession(ctx, id)

// DeleteSessionsByUserID deletes all sessions for a user
func (a *SessionStoreAdapter) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	return a.store.DeleteSessionsByUserID(ctx, userID)

// ListSessionsByUserID lists sessions for a user
func (a *SessionStoreAdapter) ListSessionsByUserID(ctx context.Context, userID string) ([]*interfaces.Session, error) {
	adapterSessions, err := a.store.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	interfaceSessions := make([]*interfaces.Session, len(adapterSessions))
	for i, adapterSession := range adapterSessions {
		interfaceSessions[i] = convertAdapterSessionToInterface(adapterSession)
	}
	return interfaceSessions, nil

// CleanExpiredSessions removes all expired sessions
func (a *SessionStoreAdapter) CleanExpiredSessions(ctx context.Context) (int, error) {
	return a.store.CleanExpiredSessions(ctx)

// convertAdapterSessionToInterface converts an adapter.Session to interfaces.Session
func convertAdapterSessionToInterface(adapterSession *Session) *interfaces.Session {
	if adapterSession == nil {
		return nil
	}
	
	return &interfaces.Session{
		ID:             adapterSession.ID,
		UserID:         adapterSession.UserID,
		Token:          adapterSession.Token,
		RefreshToken:   adapterSession.RefreshToken,
		ExpiresAt:      adapterSession.ExpiresAt,
		CreatedAt:      adapterSession.CreatedAt,
		LastActivity:   adapterSession.LastActivityAt,
		IPAddress:      adapterSession.IPAddress,
		UserAgent:      adapterSession.UserAgent,
		Metadata:       convertStringToMetadata(adapterSession.Metadata),
	}

// convertMetadataToString converts metadata map to JSON string
func convertMetadataToString(metadata map[string]interface{}) string {
	if metadata == nil || len(metadata) == 0 {
		return ""
	}
	
	data, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	return string(data)

// convertStringToMetadata converts JSON string to metadata map
func convertStringToMetadata(metadataStr string) map[string]interface{} {
	if metadataStr == "" {
		return make(map[string]interface{})
	}
	
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return make(map[string]interface{})
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
