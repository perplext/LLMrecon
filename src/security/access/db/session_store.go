// Package db provides database implementations of the access control interfaces
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/perplext/LLMrecon/src/security/access/db/adapter"
)

// SQLSessionStore is a SQL implementation of SessionStore
type SQLSessionStore struct {
	db *sql.DB
}

// NewSQLSessionStore creates a new SQL-based session store
func NewSQLSessionStore(db *sql.DB) (adapter.SessionStore, error) {
	store := &SQLSessionStore{
		db: db,
	}

	// Initialize database schema if needed
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema initializes the database schema
func (s *SQLSessionStore) initSchema() error {
	// Create sessions table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			refresh_token TEXT UNIQUE,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			last_activity_at TIMESTAMP NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			metadata TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create indexes
	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`)
	if err != nil {
		return fmt.Errorf("failed to create user_id index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`)
	if err != nil {
		return fmt.Errorf("failed to create token index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token)`)
	if err != nil {
		return fmt.Errorf("failed to create refresh_token index: %w", err)
	}

	return nil
}

// CreateSession creates a new session
func (s *SQLSessionStore) CreateSession(ctx context.Context, session *adapter.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token, refresh_token, expires_at, created_at, last_activity_at,
			ip_address, user_agent, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.Token,
		session.RefreshToken,
		formatTime(session.ExpiresAt),
		formatTime(session.CreatedAt),
		formatTime(session.LastActivityAt),
		session.IPAddress,
		session.UserAgent,
		session.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByID retrieves a session by ID
func (s *SQLSessionStore) GetSessionByID(ctx context.Context, id string) (*adapter.Session, error) {
	query := `
		SELECT id, user_id, token, refresh_token, expires_at, created_at, last_activity_at,
		       ip_address, user_agent, metadata
		FROM sessions
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanSession(row)
}

// GetSessionByToken retrieves a session by token
func (s *SQLSessionStore) GetSessionByToken(ctx context.Context, token string) (*adapter.Session, error) {
	query := `
		SELECT id, user_id, token, refresh_token, expires_at, created_at, last_activity_at,
		       ip_address, user_agent, metadata
		FROM sessions
		WHERE token = ?
	`

	row := s.db.QueryRowContext(ctx, query, token)
	return s.scanSession(row)
}

// GetSessionByRefreshToken retrieves a session by refresh token
func (s *SQLSessionStore) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*adapter.Session, error) {
	query := `
		SELECT id, user_id, token, refresh_token, expires_at, created_at, last_activity_at,
		       ip_address, user_agent, metadata
		FROM sessions
		WHERE refresh_token = ?
	`

	row := s.db.QueryRowContext(ctx, query, refreshToken)
	return s.scanSession(row)
}

// UpdateSession updates an existing session
func (s *SQLSessionStore) UpdateSession(ctx context.Context, session *adapter.Session) error {
	query := `
		UPDATE sessions
		SET token = ?, refresh_token = ?, expires_at = ?, last_activity_at = ?,
		    ip_address = ?, user_agent = ?, metadata = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(
		ctx,
		query,
		session.Token,
		session.RefreshToken,
		formatTime(session.ExpiresAt),
		formatTime(session.LastActivityAt),
		session.IPAddress,
		session.UserAgent,
		session.Metadata,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// DeleteSession deletes a session
func (s *SQLSessionStore) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// DeleteSessionsByUserID deletes all sessions for a user
func (s *SQLSessionStore) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions for user: %w", err)
	}

	return nil
}

// ListSessionsByUserID lists sessions for a user
func (s *SQLSessionStore) ListSessionsByUserID(ctx context.Context, userID string) ([]*adapter.Session, error) {
	query := `
		SELECT id, user_id, token, refresh_token, expires_at, created_at, last_activity_at,
		       ip_address, user_agent, metadata
		FROM sessions
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*adapter.Session
	for rows.Next() {
		session, err := s.scanSessionFromRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// CleanExpiredSessions removes expired sessions
func (s *SQLSessionStore) CleanExpiredSessions(ctx context.Context) (int, error) {
	query := `DELETE FROM sessions WHERE expires_at < ?`
	result, err := s.db.ExecContext(ctx, query, formatTime(time.Now()))
	if err != nil {
		return 0, fmt.Errorf("failed to clean expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// scanSession scans a session from a database row
func (s *SQLSessionStore) scanSession(row *sql.Row) (*adapter.Session, error) {
	var session adapter.Session
	var expiresAtStr, createdAtStr, lastActivityAtStr string
	var refreshToken, ipAddress, userAgent, metadata sql.NullString

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&refreshToken,
		&expiresAtStr,
		&createdAtStr,
		&lastActivityAtStr,
		&ipAddress,
		&userAgent,
		&metadata,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to scan session: %w", err)
	}

	// Parse timestamps
	session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)
	session.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	session.LastActivityAt, _ = time.Parse(time.RFC3339, lastActivityAtStr)

	// Handle nullable fields
	if refreshToken.Valid {
		session.RefreshToken = refreshToken.String
	}
	if ipAddress.Valid {
		session.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		session.UserAgent = userAgent.String
	}
	if metadata.Valid {
		session.Metadata = metadata.String
	}

	return &session, nil
}

// scanSessionFromRows scans a session from database rows
func (s *SQLSessionStore) scanSessionFromRows(rows *sql.Rows) (*adapter.Session, error) {
	var session adapter.Session
	var expiresAtStr, createdAtStr, lastActivityAtStr string
	var refreshToken, ipAddress, userAgent, metadata sql.NullString

	err := rows.Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&refreshToken,
		&expiresAtStr,
		&createdAtStr,
		&lastActivityAtStr,
		&ipAddress,
		&userAgent,
		&metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan session: %w", err)
	}

	// Parse timestamps
	session.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)
	session.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	session.LastActivityAt, _ = time.Parse(time.RFC3339, lastActivityAtStr)

	// Handle nullable fields
	if refreshToken.Valid {
		session.RefreshToken = refreshToken.String
	}
	if ipAddress.Valid {
		session.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		session.UserAgent = userAgent.String
	}
	if metadata.Valid {
		session.Metadata = metadata.String
	}

	return &session, nil
}

// Close closes the SQL connection
func (s *SQLSessionStore) Close() error {
	return s.db.Close()
}
