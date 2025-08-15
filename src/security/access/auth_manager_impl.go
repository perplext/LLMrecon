// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthManagerImpl implements the AuthManager interface
type AuthManagerImpl struct {
	mu           sync.RWMutex
	userStore    interfaces.UserStore
	sessionStore interfaces.SessionStore
	auditLogger  AuditLogger
	initialized  bool
	config       *AuthConfig

// AuthConfig contains configuration for the auth manager
type AuthConfig struct {
	// Session configuration
	SessionTimeout     time.Duration
	SessionMaxInactive time.Duration
}

	// Password configuration
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireLower   bool
	PasswordRequireNumber  bool
	PasswordRequireSpecial bool
	PasswordMaxAge         time.Duration

	// MFA configuration
	MFAEnabled bool
	MFAMethods []string

// NewAuthManagerImpl creates a new auth manager implementation
func NewAuthManagerImpl(userStore interfaces.UserStore, sessionStore interfaces.SessionStore, auditLogger AuditLogger, config *AuthConfig) *AuthManagerImpl {
	// Set default configuration if not provided
	if config == nil {
		config = &AuthConfig{
			SessionTimeout:     24 * time.Hour,
			SessionMaxInactive: 1 * time.Hour,
			PasswordMinLength:  8,
		}
	}

	return &AuthManagerImpl{
		userStore:    userStore,
		sessionStore: sessionStore,
		auditLogger:  auditLogger,
		config:       config,
	}

// Initialize initializes the auth manager
func (m *AuthManagerImpl) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = true
	return nil

// Login authenticates a user
func (m *AuthManagerImpl) Login(ctx context.Context, username, password string) (*models.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the user by username
	user, err := m.userStore.GetUserByUsername(ctx, username)
	if err != nil {
		// Log failed login attempt
		ipAddress := getIPFromContext(ctx)
		userAgent := getUserAgentFromContext(ctx)
		auditLog := &AuditLog{
			ID:          uuid.New().String(),
			UserID:      "",
			Username:    username,
			Action:      AuditActionLogin,
			Resource:    "auth",
			ResourceID:  "",
			Description: "Failed login attempt for user: " + username,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed",
			Timestamp:   time.Now(),
		}

		if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		return nil, errors.New("invalid username or password")
	}

	// Check if the user is active
	if !user.Active {
		// Log failed login attempt for inactive user
		ipAddress := getIPFromContext(ctx)
		userAgent := getUserAgentFromContext(ctx)
		auditLog := &AuditLog{
			ID:          uuid.New().String(),
			UserID:      user.ID,
			Username:    username,
			Action:      AuditActionLogin,
			Resource:    "auth",
			ResourceID:  user.ID,
			Description: "Failed login attempt for inactive user: " + username,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed_inactive",
			Timestamp:   time.Now(),
		}
		if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		return nil, errors.New("user account is inactive")
	}

	// Check if the user is locked
	if user.Locked {
		// Log failed login attempt for locked user
		ipAddress := getIPFromContext(ctx)
		userAgent := getUserAgentFromContext(ctx)
		auditLog := &AuditLog{
			ID:          uuid.New().String(),
			UserID:      user.ID,
			Username:    username,
			Action:      AuditActionLogin,
			Resource:    "auth",
			ResourceID:  user.ID,
			Description: "Failed login attempt for locked user: " + username,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityHigh,
			Status:      "failed_locked",
			Timestamp:   time.Now(),
		}

		if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		return nil, errors.New("user account is locked")
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Increment failed login attempts
		user.FailedLoginAttempts++

		// Lock the account if too many failed attempts
		if user.FailedLoginAttempts >= 5 {
			user.Locked = true
		}

		// Update the user
		if err := m.userStore.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		// Log failed login attempt
		ipAddress := getIPFromContext(ctx)
		userAgent := getUserAgentFromContext(ctx)
		auditLog := &AuditLog{
			ID:          uuid.New().String(),
			UserID:      user.ID,
			Username:    username,
			Action:      AuditActionLogin,
			Resource:    "auth",
			ResourceID:  user.ID,
			Description: "Failed login attempt (incorrect password) for user: " + username,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed_password",
			Timestamp:   time.Now(),
		}

		if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}
		return nil, errors.New("invalid username or password")
	}

	// Reset failed login attempts
	user.FailedLoginAttempts = 0
	user.LastLogin = time.Now()

	// Update the user
	if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Create a new session
	session := &models.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		Token:        uuid.New().String(),
		IPAddress:    getIPFromContext(ctx),
		UserAgent:    getUserAgentFromContext(ctx),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(m.config.SessionTimeout),
		LastActivity: time.Now(),
	}

	// Store the session
	err = m.sessionStore.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// Log successful login
	ipAddress := getIPFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		Username:    username,
		Action:      AuditActionLogin,
		Resource:    "auth",
		ResourceID:  user.ID,
		Description: "Successful login for user: " + username,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return session, nil

// Logout logs out a user
func (m *AuthManagerImpl) Logout(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the session
	session, err := m.sessionStore.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Mark session as expired by setting expiry to now
	session.ExpiresAt = time.Now()

	// Update the session
	err = m.sessionStore.UpdateSession(ctx, session)
	if err != nil {
		return err
	}

	// Log logout
	ipAddress := getIPFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)
	user, _ := m.userStore.GetUserByID(ctx, session.UserID)
	username := ""
	if user != nil {
		username = user.Username
	}
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      session.UserID,
		Username:    username,
		Action:      AuditActionLogout,
		Resource:    "auth",
		ResourceID:  session.UserID,
		Description: "Logout for user: " + username,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil

// ValidateSession validates a session
func (m *AuthManagerImpl) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get the session
	session, err := m.sessionStore.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	// Check if the session has expired
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session has expired")
	}

	// Check if the session has been inactive for too long
	if time.Now().Sub(session.LastActivity) > m.config.SessionMaxInactive {
		// Mark session as expired
		session.ExpiresAt = time.Now()
		if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		return nil, errors.New("session has been inactive for too long")
	}

	// Update the last activity time
	session.LastActivity = time.Now()
	if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return session, nil

// RefreshSession refreshes a session
func (m *AuthManagerImpl) RefreshSession(ctx context.Context, sessionID string) (*models.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the session
	session, err := m.sessionStore.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if the session has expired
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session has expired")
	}

	// Refresh the session
	session.ExpiresAt = time.Now().Add(m.config.SessionTimeout)
	session.LastActivity = time.Now()

	// Update the session
	err = m.sessionStore.UpdateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil

// ChangePassword changes a user's password
func (m *AuthManagerImpl) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the user
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify the old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		// Log failed password change
		auditLog := &AuditLog{
			ID:          uuid.New().String(),
			UserID:      userID,
			Action:      AuditActionUpdate,
			Resource:    "auth",
			ResourceID:  userID,
			Description: "Failed password change: incorrect old password",
			Timestamp:   time.Now(),
			Severity:    AuditSeverityMedium,
			Status:      "failed",
		}

		if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		return errors.New("incorrect old password")
	}

	// Validate the new password
	err = m.validatePassword(newPassword)
	if err != nil {
		// Log failed password change
		ipAddress := getIPFromContext(ctx)
		userAgent := getUserAgentFromContext(ctx)
		auditLog := &AuditLog{
			ID:          uuid.New().String(),
			UserID:      userID,
			Action:      AuditActionUpdate,
			Resource:    "auth",
			ResourceID:  userID,
			Description: "Failed password change: " + err.Error(),
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed",
			Timestamp:   time.Now(),
		}

		if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		return err
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update the user
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	err = m.userStore.UpdateUser(ctx, user)
	if err != nil {
		return err
	}
	// Log password change
	ipAddress := getIPFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      userID,
		Action:      AuditActionUpdate,
		Resource:    "auth",
		ResourceID:  userID,
		Description: "Password changed successfully",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil

// ResetPassword resets a user's password
func (m *AuthManagerImpl) ResetPassword(ctx context.Context, userID, newPassword string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the user
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Validate the new password
	err = m.validatePassword(newPassword)
	if err != nil {
		// Log failed password reset
		auditLog := &AuditLog{
			ID:          uuid.New().String(),
			UserID:      getUserIDFromContext(ctx),
			Action:      AuditActionUpdate,
			Resource:    "auth",
			ResourceID:  userID,
			Description: "Failed password reset: " + err.Error(),
			Timestamp:   time.Now(),
			Severity:    AuditSeverityMedium,
			Status:      "failed",
		}

		if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}

		return err
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update the user
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	user.FailedLoginAttempts = 0
	user.Locked = false

	err = m.userStore.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	// Log password reset
	ipAddress := getIPFromContext(ctx)
	userAgent := getUserAgentFromContext(ctx)
	auditLog := &AuditLog{
		ID:          uuid.New().String(),
		UserID:      getUserIDFromContext(ctx),
		Action:      AuditActionUpdate,
		Resource:    "auth",
		ResourceID:  userID,
		Description: "Password reset for user: " + user.Username,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		Timestamp:   time.Now(),
	}

	if err := m.auditLogger.LogAudit(ctx, auditLog); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil

// Close closes the auth manager
func (m *AuthManagerImpl) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = false
	return nil

// validatePassword validates a password against the password policy
func (m *AuthManagerImpl) validatePassword(password string) error {
	// Check minimum length
	if len(password) < m.config.PasswordMinLength {
		return errors.New("password is too short")
	}

	// Check for uppercase letters
	if m.config.PasswordRequireUpper {
		hasUpper := false
		for _, c := range password {
			if c >= 'A' && c <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return errors.New("password must contain at least one uppercase letter")
		}
	}

	// Check for lowercase letters
	if m.config.PasswordRequireLower {
		hasLower := false
		for _, c := range password {
			if c >= 'a' && c <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return errors.New("password must contain at least one lowercase letter")
		}
	}

	// Check for numbers
	if m.config.PasswordRequireNumber {
		hasNumber := false
		for _, c := range password {
			if c >= '0' && c <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return errors.New("password must contain at least one number")
		}
	}

	// Check for special characters
	if m.config.PasswordRequireSpecial {
		hasSpecial := false
		for _, c := range password {
			if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			return errors.New("password must contain at least one special character")
		}
	}

	return nil

// Helper functions to get context values
func getIPFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	ip, ok := ctx.Value("ip").(string)
	if !ok {
		return ""
	}

	return ip

func getUserAgentFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userAgent, ok := ctx.Value("user_agent").(string)
	if !ok {
		return ""
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
