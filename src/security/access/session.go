// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Token           string                 `json:"token"`
	RefreshToken    string                 `json:"refresh_token,omitempty"`
	IPAddress       string                 `json:"ip_address"`
	UserAgent       string                 `json:"user_agent"`
	ExpiresAt       time.Time              `json:"expires_at"`
	LastActivity    time.Time              `json:"last_activity"`
	MFACompleted    bool                   `json:"mfa_completed"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SimpleInMemorySessionStore is a simple in-memory implementation of SessionStore
type SimpleInMemorySessionStore struct {
	sessions map[string]*Session
	userSessions map[string][]string // Maps userID to session IDs
	mu       sync.RWMutex
}

// NewSimpleInMemorySessionStore creates a new simple in-memory session store
func NewSimpleInMemorySessionStore() *SimpleInMemorySessionStore {
	return &SimpleInMemorySessionStore{
		sessions: make(map[string]*Session),
		userSessions: make(map[string][]string),
	}
}

// CreateSession creates a new session
func (s *SimpleInMemorySessionStore) CreateSession(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store session
	s.sessions[session.ID] = session

	// Add to user sessions
	s.userSessions[session.UserID] = append(s.userSessions[session.UserID], session.ID)

	return nil
}

// GetSession retrieves a session by ID
func (s *SimpleInMemorySessionStore) GetSession(ctx context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// UpdateSession updates an existing session
func (s *SimpleInMemorySessionStore) UpdateSession(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	_, ok := s.sessions[session.ID]
	if !ok {
		return ErrSessionExpired
	}

	// Update session
	s.sessions[session.ID] = session

	return nil
}

// DeleteSession deletes a session by ID
func (s *SimpleInMemorySessionStore) DeleteSession(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session exists
	session, ok := s.sessions[id]
	if !ok {
		return nil
	}

	// Remove from user sessions
	if userSessions, ok := s.userSessions[session.UserID]; ok {
		newUserSessions := make([]string, 0, len(userSessions))
		for _, sid := range userSessions {
			if sid != id {
				newUserSessions = append(newUserSessions, sid)
			}
		}
		s.userSessions[session.UserID] = newUserSessions
	}

	// Delete session
	delete(s.sessions, id)

	return nil
}

// GetUserSessions retrieves all sessions for a user
func (s *SimpleInMemorySessionStore) GetUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get session IDs for user
	sessionIDs, ok := s.userSessions[userID]
	if !ok {
		return []*Session{}, nil
	}

	// Get sessions
	sessions := make([]*Session, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		if session, ok := s.sessions[id]; ok {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// CleanExpiredSessions removes expired sessions
func (s *SimpleInMemorySessionStore) CleanExpiredSessions(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiredIDs := make([]string, 0)

	// Find expired sessions
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			expiredIDs = append(expiredIDs, id)
		}
	}

	// Delete expired sessions
	for _, id := range expiredIDs {
		session := s.sessions[id]
		
		// Remove from user sessions
		if userSessions, ok := s.userSessions[session.UserID]; ok {
			newUserSessions := make([]string, 0, len(userSessions))
			for _, sid := range userSessions {
				if sid != id {
					newUserSessions = append(newUserSessions, sid)
				}
			}
			s.userSessions[session.UserID] = newUserSessions
		}
		
		// Delete session
		delete(s.sessions, id)
	}

	return nil
}

// PersistentSessionStore is an interface for session stores that persist data
type PersistentSessionStore interface {
	SessionStore
	Initialize(ctx context.Context) error
	CloseWithContext(ctx context.Context) error
}

// SessionManager manages user sessions
type SessionManager struct {
	store      SessionStore
	config     *SessionPolicy
	auditLogger AuditLogger
	cleanupTicker *time.Ticker
	stopChan   chan struct{}
}

// NewSessionManager creates a new session manager
func NewSessionManager(store SessionStore, config *SessionPolicy, auditLogger AuditLogger) *SessionManager {
	manager := &SessionManager{
		store:      store,
		config:     config,
		auditLogger: auditLogger,
		stopChan:   make(chan struct{}),
	}

	// Start cleanup routine if cleanup interval is set
	if config != nil && config.CleanupInterval > 0 {
		manager.cleanupTicker = time.NewTicker(time.Duration(config.CleanupInterval) * time.Minute)
		go manager.cleanupRoutine()
	}

	return manager
}

// cleanupRoutine periodically cleans up expired sessions
func (m *SessionManager) cleanupRoutine() {
	for {
		select {
		case <-m.cleanupTicker.C:
			ctx := context.Background()
			if err := m.store.CleanExpiredSessions(ctx); err != nil {
				// Log error
				if m.auditLogger != nil {
					m.auditLogger.LogAudit(ctx, &AuditLog{
						Timestamp:  time.Now(),
						Action:     AuditActionSystem,
						Resource:   "session",
						Description: "Failed to clean up expired sessions",
						Severity:   AuditSeverityError,
						Status:     "failed",
						Metadata: map[string]interface{}{
							"error": err.Error(),
						},
					})
				}
			}
		case <-m.stopChan:
			if m.cleanupTicker != nil {
				m.cleanupTicker.Stop()
			}
			return
		}
	}
}

// Stop stops the session manager
func (m *SessionManager) Stop() {
	close(m.stopChan)
}

// CreateSession creates a new session
func (m *SessionManager) CreateSession(ctx context.Context, userID, ipAddress, userAgent string, mfaCompleted bool) (*Session, error) {
	// Create session
	session := &Session{
		ID:           generateRandomID(),
		UserID:       userID,
		Token:        generateRandomToken(),
		RefreshToken: generateRandomToken(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		MFACompleted: mfaCompleted,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Set expiration time
	tokenExpiration := 60 // Default 60 minutes
	if m.config != nil && m.config.TokenExpiration > 0 {
		tokenExpiration = m.config.TokenExpiration
	}
	session.ExpiresAt = time.Now().Add(time.Duration(tokenExpiration) * time.Minute)

	// Save session
	if err := m.store.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	// Log session creation
	if m.auditLogger != nil {
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:  time.Now(),
			UserID:     userID,
			Action:     AuditActionLogin,
			Resource:   "session",
			ResourceID: session.ID,
			Description: "Session created",
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Severity:   AuditSeverityInfo,
			Status:     "success",
			SessionID:  session.ID,
			Metadata: map[string]interface{}{
				"mfa_completed": mfaCompleted,
			},
		})
	}

	return session, nil
}

// ValidateSession validates a session
func (m *SessionManager) ValidateSession(ctx context.Context, sessionID, token string, ipAddress, userAgent string) (*Session, error) {
	// Get session
	session, err := m.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		m.store.DeleteSession(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	// Verify token
	if session.Token != token {
		return nil, ErrInvalidToken
	}

	// Check if MFA is completed if required
	if !session.MFACompleted {
		return nil, ErrMFARequired
	}

	// Check IP binding if enabled
	if m.config != nil && m.config.EnforceIPBinding && session.IPAddress != ipAddress {
		// Log suspicious activity
		if m.auditLogger != nil {
			m.auditLogger.LogAudit(ctx, &AuditLog{
				Timestamp:  time.Now(),
				UserID:     session.UserID,
				Action:     AuditActionLogin,
				Resource:   "session",
				ResourceID: session.ID,
				Description: "Session IP mismatch",
				IPAddress:  ipAddress,
				UserAgent:  userAgent,
				Severity:   AuditSeverityHigh,
				Status:     "failed",
				SessionID:  session.ID,
				Metadata: map[string]interface{}{
					"expected_ip": session.IPAddress,
					"actual_ip":   ipAddress,
				},
			})
		}
		return nil, ErrInvalidToken
	}

	// Check user agent binding if enabled
	if m.config != nil && m.config.EnforceUserAgentBinding && session.UserAgent != userAgent {
		// Log suspicious activity
		if m.auditLogger != nil {
			m.auditLogger.LogAudit(ctx, &AuditLog{
				Timestamp:  time.Now(),
				UserID:     session.UserID,
				Action:     AuditActionLogin,
				Resource:   "session",
				ResourceID: session.ID,
				Description: "Session user agent mismatch",
				IPAddress:  ipAddress,
				UserAgent:  userAgent,
				Severity:   AuditSeverityMedium,
				Status:     "failed",
				SessionID:  session.ID,
				Metadata: map[string]interface{}{
					"expected_user_agent": session.UserAgent,
					"actual_user_agent":   userAgent,
				},
			})
		}
		return nil, ErrInvalidToken
	}

	// Check inactivity timeout
	if m.config != nil && m.config.InactivityTimeout > 0 {
		inactivityTimeout := time.Duration(m.config.InactivityTimeout) * time.Minute
		if time.Since(session.LastActivity) > inactivityTimeout {
			// Delete inactive session
			m.store.DeleteSession(ctx, sessionID)
			return nil, ErrSessionExpired
		}
	}

	// Update last activity
	session.LastActivity = time.Now()
	if err := m.store.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// RefreshSession refreshes a session and returns a new token
func (m *SessionManager) RefreshSession(ctx context.Context, sessionID, refreshToken string) (*Session, error) {
	// Get session
	session, err := m.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify refresh token
	if session.RefreshToken != refreshToken {
		return nil, ErrInvalidToken
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		m.store.DeleteSession(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	// Generate new tokens
	session.Token = generateRandomToken()
	session.RefreshToken = generateRandomToken()
	session.LastActivity = time.Now()

	// Set new expiration time
	tokenExpiration := 60 // Default 60 minutes
	if m.config != nil && m.config.TokenExpiration > 0 {
		tokenExpiration = m.config.TokenExpiration
	}
	session.ExpiresAt = time.Now().Add(time.Duration(tokenExpiration) * time.Minute)

	// Update session
	if err := m.store.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	// Log session refresh
	if m.auditLogger != nil {
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:  time.Now(),
			UserID:     session.UserID,
			Action:     AuditActionLogin,
			Resource:   "session",
			ResourceID: session.ID,
			Description: "Session refreshed",
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			Severity:   AuditSeverityInfo,
			Status:     "success",
			SessionID:  session.ID,
		})
	}

	return session, nil
}

// InvalidateSession invalidates a session
func (m *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	// Get session
	session, err := m.store.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Delete session
	if err := m.store.DeleteSession(ctx, sessionID); err != nil {
		return err
	}

	// Log session invalidation
	if m.auditLogger != nil {
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:  time.Now(),
			UserID:     session.UserID,
			Action:     AuditActionLogout,
			Resource:   "session",
			ResourceID: session.ID,
			Description: "Session invalidated",
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			Severity:   AuditSeverityInfo,
			Status:     "success",
			SessionID:  session.ID,
		})
	}

	return nil
}

// InvalidateUserSessions invalidates all sessions for a user
func (m *SessionManager) InvalidateUserSessions(ctx context.Context, userID string) error {
	// Get user sessions
	sessions, err := m.store.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	// Delete each session
	for _, session := range sessions {
		if err := m.store.DeleteSession(ctx, session.ID); err != nil {
			return err
		}

		// Log session invalidation
		if m.auditLogger != nil {
			m.auditLogger.LogAudit(ctx, &AuditLog{
				Timestamp:  time.Now(),
				UserID:     userID,
				Action:     AuditActionLogout,
				Resource:   "session",
				ResourceID: session.ID,
				Description: "Session invalidated (user logout)",
				IPAddress:  session.IPAddress,
				UserAgent:  session.UserAgent,
				Severity:   AuditSeverityInfo,
				Status:     "success",
				SessionID:  session.ID,
			})
		}
	}

	return nil
}

// GetUserSessions retrieves all sessions for a user
func (m *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	return m.store.GetUserSessions(ctx, userID)
}

// GetSessionFromContext extracts a session from the request context
func (m *SessionManager) GetSessionFromContext(ctx context.Context) (*Session, error) {
	// Check if session is stored in context
	session, ok := ctx.Value("session").(*Session)
	if !ok {
		return nil, ErrSessionExpired
	}
	return session, nil
}
