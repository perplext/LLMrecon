// Package db provides database implementations of the security access control interfaces
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// Error definitions
var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
)

// SQLUserStore is a SQL implementation of UserStore
type SQLUserStore struct {
	db *sql.DB
}

// Ensure SQLUserStore implements interfaces.UserStore
var _ interfaces.UserStore = (*SQLUserStore)(nil)

// Close closes the user store
func (s *SQLUserStore) Close() error {
	// Nothing to close for SQLUserStore as the DB connection is managed externally
	return nil
}

// NewSQLUserStore creates a new SQL-based user store
func NewSQLUserStore(db *sql.DB) (interfaces.UserStore, error) {
	store := &SQLUserStore{
		db: db,
	}

	// Initialize database schema if needed
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema initializes the database schema
func (s *SQLUserStore) initSchema() error {
	// Create users table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			roles TEXT NOT NULL,
			permissions TEXT,
			mfa_enabled INTEGER NOT NULL DEFAULT 0,
			mfa_methods TEXT,
			mfa_secret TEXT,
			failed_login_attempts INTEGER NOT NULL DEFAULT 0,
			locked INTEGER NOT NULL DEFAULT 0,
			lock_expiration TIMESTAMP,
			last_login TIMESTAMP,
			last_password_change TIMESTAMP NOT NULL,
			password_history TEXT,
			active INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			metadata TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create indexes
	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`)
	if err != nil {
		return fmt.Errorf("failed to create username index: %w", err)
	}

	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`)
	if err != nil {
		return fmt.Errorf("failed to create email index: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (s *SQLUserStore) GetUserByID(ctx context.Context, id string) (*interfaces.User, error) {
	query := `
		SELECT id, username, email, password_hash, roles, permissions, mfa_enabled, mfa_methods, mfa_secret,
		       failed_login_attempts, locked, lock_expiration, last_login, last_password_change, password_history,
		       active, created_at, updated_at, metadata
		FROM users
		WHERE id = ?
	`
	row := s.db.QueryRowContext(ctx, query, id)
	return s.scanUser(row)
}

// GetUserByUsername retrieves a user by username
func (s *SQLUserStore) GetUserByUsername(ctx context.Context, username string) (*interfaces.User, error) {
	query := `
		SELECT id, username, email, password_hash, roles, permissions, mfa_enabled, mfa_methods, mfa_secret,
		       failed_login_attempts, locked, lock_expiration, last_login, last_password_change, password_history,
		       active, created_at, updated_at, metadata
		FROM users
		WHERE username = ?
	`

	row := s.db.QueryRowContext(ctx, query, username)
	return s.scanUser(row)
}

// GetUserByEmail retrieves a user by email
func (s *SQLUserStore) GetUserByEmail(ctx context.Context, email string) (*interfaces.User, error) {
	query := `
		SELECT id, username, email, password_hash, roles, permissions, mfa_enabled, mfa_methods, mfa_secret,
		       failed_login_attempts, locked, lock_expiration, last_login, last_password_change, password_history,
		       active, created_at, updated_at, metadata
		FROM users
		WHERE email = ?
	`

	row := s.db.QueryRowContext(ctx, query, email)
	return s.scanUser(row)
}

// Helper function to handle missing fields in interfaces.User
func getMFASecretForDB() string {
	// MFASecret is not part of interfaces.User
	return ""
}

func getEmptyPasswordHistory() string {
	// PasswordHistory is not part of interfaces.User
	return "[]" // Empty JSON array
}

func getZeroTime() time.Time {
	// Return zero time for missing time fields
	return time.Time{}
}

// CreateUser creates a new user
func (s *SQLUserStore) CreateUser(ctx context.Context, user *interfaces.User) error {
	// Serialize roles
	rolesJSON, err := json.Marshal(user.Roles)
	if err != nil {
		return fmt.Errorf("failed to serialize roles: %w", err)
	}

	// Serialize permissions
	permissionsJSON, err := json.Marshal(user.Permissions)
	if err != nil {
		return fmt.Errorf("failed to serialize permissions: %w", err)
	}

	// Serialize MFA methods
	mfaMethodsJSON, err := json.Marshal(user.MFAMethods)
	if err != nil {
		return fmt.Errorf("failed to serialize MFA methods: %w", err)
	}

	// Password history is not part of interfaces.User
	passwordHistoryJSON := getEmptyPasswordHistory()

	// Serialize metadata
	metadataJSON, err := json.Marshal(user.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	// Insert user
	query := `
		INSERT INTO users (
			id, username, email, password_hash, roles, permissions, mfa_enabled, mfa_methods, mfa_secret,
			failed_login_attempts, locked, lock_expiration, last_login, last_password_change, password_history,
			active, created_at, updated_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		rolesJSON,
		permissionsJSON,
		boolToInt(user.MFAEnabled),
		mfaMethodsJSON,
		getMFASecretForDB(),
		user.FailedLoginAttempts,
		boolToInt(user.Locked),
		formatTime(getZeroTime()), // LockExpiration is not part of interfaces.User
		formatTime(user.LastLogin),
		formatTime(user.LastPasswordChange),
		passwordHistoryJSON,
		boolToInt(user.Active),
		formatTime(user.CreatedAt),
		formatTime(user.UpdatedAt),
		metadataJSON,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			return fmt.Errorf("username already exists")
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return fmt.Errorf("email already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func (s *SQLUserStore) UpdateUser(ctx context.Context, user *interfaces.User) error {
	// Serialize roles
	rolesJSON, err := json.Marshal(user.Roles)
	if err != nil {
		return fmt.Errorf("failed to serialize roles: %w", err)
	}

	// Serialize permissions
	permissionsJSON, err := json.Marshal(user.Permissions)
	if err != nil {
		return fmt.Errorf("failed to serialize permissions: %w", err)
	}

	// Serialize MFA methods
	mfaMethodsJSON, err := json.Marshal(user.MFAMethods)
	if err != nil {
		return fmt.Errorf("failed to serialize MFA methods: %w", err)
	}

	// Password history is not part of interfaces.User
	passwordHistoryJSON := getEmptyPasswordHistory()

	// Serialize metadata
	metadataJSON, err := json.Marshal(user.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	// Update user
	query := `
		UPDATE users SET
			username = ?,
			email = ?,
			password_hash = ?,
			roles = ?,
			permissions = ?,
			mfa_enabled = ?,
			mfa_methods = ?,
			mfa_secret = ?,
			failed_login_attempts = ?,
			locked = ?,
			lock_expiration = ?,
			last_login = ?,
			last_password_change = ?,
			password_history = ?,
			active = ?,
			updated_at = ?,
			metadata = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		rolesJSON,
		permissionsJSON,
		boolToInt(user.MFAEnabled),
		mfaMethodsJSON,
		getMFASecretForDB(),
		user.FailedLoginAttempts,
		boolToInt(user.Locked),
		formatTime(getZeroTime()), // LockExpiration is not part of interfaces.User
		formatTime(user.LastLogin),
		formatTime(user.LastPasswordChange),
		passwordHistoryJSON,
		boolToInt(user.Active),
		formatTime(user.UpdatedAt),
		metadataJSON,
		user.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			return fmt.Errorf("username already exists")
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return fmt.Errorf("email already exists")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser deletes a user by ID
func (s *SQLUserStore) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsers lists all users
func (s *SQLUserStore) ListUsers(ctx context.Context, filter map[string]interface{}, offset, limit int) ([]*interfaces.User, int, error) {
	// Build the query with filters
	baseQuery := `
		SELECT id, username, email, password_hash, roles, permissions, mfa_enabled, mfa_methods, mfa_secret,
		       failed_login_attempts, locked, lock_expiration, last_login, last_password_change, password_history,
		       active, created_at, updated_at, metadata
		FROM users
	`

	// Add WHERE clauses for filters
	whereClause := ""
	args := []interface{}{}

	if len(filter) > 0 {
		whereClause = " WHERE "
		conditions := []string{}

		for key, value := range filter {
			// Handle different filter types
			switch key {
			case "username":
				conditions = append(conditions, "username LIKE ?")
				args = append(args, fmt.Sprintf("%%%s%%", value))
			case "email":
				conditions = append(conditions, "email LIKE ?")
				args = append(args, fmt.Sprintf("%%%s%%", value))
			case "active":
				conditions = append(conditions, "active = ?")
				args = append(args, boolToInt(value.(bool)))
			case "locked":
				conditions = append(conditions, "locked = ?")
				args = append(args, boolToInt(value.(bool)))
			}
		}

		whereClause += strings.Join(conditions, " AND ")
	}

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM users" + whereClause
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Add ORDER BY, LIMIT and OFFSET
	query := baseQuery + whereClause + " ORDER BY username LIMIT ? OFFSET ?"
	// Add limit and offset to args
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*interfaces.User
	for rows.Next() {
		user, err := s.scanUserFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, total, nil
}

// scanUser scans a user from a database row
func (s *SQLUserStore) scanUser(row *sql.Row) (*interfaces.User, error) {
	var user interfaces.User
	var rolesJSON, permissionsJSON, mfaMethodsJSON, metadataJSON string
	var mfaEnabled, locked, active int
	var lastLoginStr, lastPasswordChangeStr, createdAtStr, updatedAtStr sql.NullString
	var mfaSecret string // Not used in interfaces.User
	var lockExpirationStr sql.NullString // Not used in interfaces.User
	var passwordHistoryJSON string // Not used in interfaces.User

	// Scan all fields from the database row
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&rolesJSON,
		&permissionsJSON,
		&mfaEnabled,
		&mfaMethodsJSON,
		&mfaSecret, // Not used in interfaces.User
		&user.FailedLoginAttempts,
		&locked,
		&lockExpirationStr, // Not used in interfaces.User
		&lastLoginStr,
		&lastPasswordChangeStr,
		&passwordHistoryJSON, // Not used in interfaces.User
		&active,
		&createdAtStr,
		&updatedAtStr,
		&metadataJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	// Parse roles
	if err := json.Unmarshal([]byte(rolesJSON), &user.Roles); err != nil {
		return nil, fmt.Errorf("failed to parse roles: %w", err)
	}

	// Parse permissions
	if permissionsJSON != "" {
		if err := json.Unmarshal([]byte(permissionsJSON), &user.Permissions); err != nil {
			return nil, fmt.Errorf("failed to parse permissions: %w", err)
		}
	}

	// Parse MFA methods
	if mfaMethodsJSON != "" {
		if err := json.Unmarshal([]byte(mfaMethodsJSON), &user.MFAMethods); err != nil {
			return nil, fmt.Errorf("failed to parse MFA methods: %w", err)
		}
	}

	// Parse metadata
	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	// Convert boolean fields
	user.MFAEnabled = intToBool(mfaEnabled)
	user.Locked = intToBool(locked)
	user.Active = intToBool(active)

	// Parse time fields
	if lastLoginStr.Valid {
		user.LastLogin, _ = time.Parse(time.RFC3339, lastLoginStr.String)
	}
	if lastPasswordChangeStr.Valid {
		user.LastPasswordChange, _ = time.Parse(time.RFC3339, lastPasswordChangeStr.String)
	}
	if createdAtStr.Valid {
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr.String)
	}
	if updatedAtStr.Valid {
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr.String)
	}

	return &user, nil
}

// scanUserFromRows scans a user from database rows
func (s *SQLUserStore) scanUserFromRows(rows *sql.Rows) (*interfaces.User, error) {
	var user interfaces.User
	var rolesJSON, permissionsJSON, mfaMethodsJSON, metadataJSON string
	var mfaEnabled, locked, active int
	var lastLoginStr, lastPasswordChangeStr, createdAtStr, updatedAtStr sql.NullString
	var mfaSecret string // Not used in interfaces.User
	var lockExpirationStr sql.NullString // Not used in interfaces.User
	var passwordHistoryJSON string // Not used in interfaces.User

	err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&rolesJSON,
		&permissionsJSON,
		&mfaEnabled,
		&mfaMethodsJSON,
		&mfaSecret, // Not used in interfaces.User
		&user.FailedLoginAttempts,
		&locked,
		&lockExpirationStr, // Not used in interfaces.User
		&lastLoginStr,
		&lastPasswordChangeStr,
		&passwordHistoryJSON, // Not used in interfaces.User
		&active,
		&createdAtStr,
		&updatedAtStr,
		&metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	// Parse roles
	if err := json.Unmarshal([]byte(rolesJSON), &user.Roles); err != nil {
		return nil, fmt.Errorf("failed to parse roles: %w", err)
	}

	// Parse permissions
	if permissionsJSON != "" {
		if err := json.Unmarshal([]byte(permissionsJSON), &user.Permissions); err != nil {
			return nil, fmt.Errorf("failed to parse permissions: %w", err)
		}
	}

	// Parse MFA methods
	if mfaMethodsJSON != "" {
		if err := json.Unmarshal([]byte(mfaMethodsJSON), &user.MFAMethods); err != nil {
			return nil, fmt.Errorf("failed to parse MFA methods: %w", err)
		}
	}

	// Password history is not part of interfaces.User

	// Parse metadata
	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	// Convert boolean fields
	user.MFAEnabled = intToBool(mfaEnabled)
	user.Locked = intToBool(locked)
	user.Active = intToBool(active)

	// Parse time fields
	if lastLoginStr.Valid {
		user.LastLogin, _ = time.Parse(time.RFC3339, lastLoginStr.String)
	}
	if lastPasswordChangeStr.Valid {
		user.LastPasswordChange, _ = time.Parse(time.RFC3339, lastPasswordChangeStr.String)
	}
	if createdAtStr.Valid {
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr.String)
	}
	if updatedAtStr.Valid {
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr.String)
	}

	return &user, nil
}

// Helper functions

// boolToInt converts a boolean to an integer (1 for true, 0 for false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool converts an integer to a boolean (true for non-zero, false for zero)
func intToBool(i int) bool {
	return i != 0
}
