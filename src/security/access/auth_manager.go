// Package access provides access control and security auditing functionality
package access

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/mfa"
	"golang.org/x/crypto/bcrypt"
)

// AuthManager handles authentication and user management
type AuthManager struct {
	userStore    UserStore
	sessionStore SessionStore
	auditLogger  AuditLogger
	mfaManager   mfa.MFAManager
	config       *AuthConfig
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(
	userStore UserStore,
	sessionStore SessionStore,
	auditLogger AuditLogger,
	mfaManager mfa.MFAManager,
	config *AuthConfig,
) (*AuthManager, error) {
	if userStore == nil {
		return nil, errors.New("user store is required")
	}
	if sessionStore == nil {
		return nil, errors.New("session store is required")
	}
	if auditLogger == nil {
		return nil, errors.New("audit logger is required")
	}
	if mfaManager == nil {
		return nil, errors.New("MFA manager is required")
	}
	if config == nil {
		return nil, errors.New("auth config is required")
	}

	return &AuthManager{
		userStore:    userStore,
		sessionStore: sessionStore,
		auditLogger:  auditLogger,
		mfaManager:   mfaManager,
		config:       config,
	}, nil

// Initialize initializes the authentication manager
func (m *AuthManager) Initialize(ctx context.Context) error {
	// Nothing to initialize for now
	return nil

// CreateUser creates a new user
func (m *AuthManager) CreateUser(ctx context.Context, username, email, password string, roles []string) (*User, error) {
	// Validate input
	if username == "" {
		return nil, errors.New("username is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password := os.Getenv("PASSWORD") {
		return nil, errors.New("password is required")
	}

	// Check if username or email already exists
	_, err := m.userStore.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("username %s already exists", username)
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("error checking username: %w", err)
	}

	_, err = m.userStore.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, fmt.Errorf("email %s already exists", email)
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("error checking email: %w", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Generate a unique ID
	userID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("error generating ID: %w", err)
	}

	// Create the user
	now := time.Now()
	user := &User{
		ID:                  userID,
		Username:            username,
		Email:               email,
		PasswordHash:        string(hashedPassword),
		Roles:               roles,
		Permissions:         []string{},
		MFAEnabled:          false,
		MFAMethods:          []string{},
		FailedLoginAttempts: 0,
		Locked:              false,
		Active:              true,
		CreatedAt:           now,
		UpdatedAt:           now,
		Metadata:            map[string]interface{}{},
	}

	// Store the user
	if err := m.userStore.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("error storing user: %w", err)
	}

	// Log the action
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   now,
		UserID:      getUserIDFromAuthContext(ctx),
		Action:      AuditActionCreate,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("Created user %s", username),
		Severity:    AuditSeverityInfo,
	})

	return user, nil

// GetUserByID retrieves a user by ID
func (m *AuthManager) GetUserByID(ctx context.Context, id string) (*User, error) {
	return m.userStore.GetUserByID(ctx, id)

// GetUserByUsername retrieves a user by username
func (m *AuthManager) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return m.userStore.GetUserByUsername(ctx, username)

// GetUserByEmail retrieves a user by email
func (m *AuthManager) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return m.userStore.GetUserByEmail(ctx, email)

// UpdateUser updates an existing user
func (m *AuthManager) UpdateUser(ctx context.Context, user *User) error {
	// Validate input
	if user.ID == "" {
		return errors.New("user ID is required")
	}
	if user.Username == "" {
		return errors.New("username is required")
	}
	if user.Email == "" {
		return errors.New("email is required")
	}

	// Check if user exists
	existingUser, err := m.userStore.GetUserByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	// Check if username or email already exists (if changed)
	if user.Username != existingUser.Username {
		_, err := m.userStore.GetUserByUsername(ctx, user.Username)
		if err == nil {
			return fmt.Errorf("username %s already exists", user.Username)
		}
		if !errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("error checking username: %w", err)
		}
	}

	if user.Email != existingUser.Email {
		_, err := m.userStore.GetUserByEmail(ctx, user.Email)
		if err == nil {
			return fmt.Errorf("email %s already exists", user.Email)
		}
		if !errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("error checking email: %w", err)
		}
	}

	// Update the user
	user.UpdatedAt = time.Now()
	if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	// Log the action
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      getUserIDFromAuthContext(ctx),
		Action:      AuditActionUpdate,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: fmt.Sprintf("Updated user %s", user.Username),
		Severity:    AuditSeverityInfo,
	})

	return nil
	

// DeleteUser deletes a user
func (m *AuthManager) DeleteUser(ctx context.Context, userID string) error {
	// Check if user exists
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Delete the user
	if err := m.userStore.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	// Log the action
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      getUserIDFromAuthContext(ctx),
		Action:      AuditActionDelete,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("Deleted user %s", user.Username),
		Severity:    AuditSeverityInfo,
	})

	return nil
// UpdateUserPassword updates a user's password
func (m *AuthManager) UpdateUserPassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Check if user exists
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Verify current password if provided
	if currentPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
			return ErrInvalidCredentials
		}
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Update the user
	user.PasswordHash = string(hashedPassword)
	user.LastPasswordChange = time.Now()
	user.UpdatedAt = time.Now()
	if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	// Log the action
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      getUserIDFromAuthContext(ctx),
		Action:      AuditActionUpdate,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("Updated password for user %s", user.Username),
		Severity:    AuditSeverityInfo,
	})

	return nil

// Login authenticates a user and creates a session
func (m *AuthManager) Login(ctx context.Context, username, password, ipAddress, userAgent string) (*Session, error) {
	// Get the user
	user, err := m.userStore.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is locked
	if user.Locked {
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      user.ID,
			Action:      AuditActionUnauthorized,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: fmt.Sprintf("Login failed for user %s: account locked", username),
			Severity:    AuditSeverityMedium,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})
		return nil, fmt.Errorf("account is locked")
	}

	// Check if user is active
	if !user.Active {
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      user.ID,
			Action:      AuditActionUnauthorized,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: fmt.Sprintf("Login failed for user %s: account inactive", username),
			Severity:    AuditSeverityMedium,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment failed login attempts
		user.FailedLoginAttempts++

		// Lock account if max attempts reached
		// Use a default of 5 max login attempts
		maxAttempts := 5
		if user.FailedLoginAttempts >= maxAttempts {
			user.Locked = true
		}

		// Update user
		m.userStore.UpdateUser(ctx, user)

		// Log the failed login
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      user.ID,
			Action:      AuditActionUnauthorized,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: fmt.Sprintf("Login failed for user %s: invalid credentials", username),
			Severity:    AuditSeverityMedium,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})

		return nil, ErrInvalidCredentials
	}

	// Reset failed login attempts
	user.FailedLoginAttempts = 0
	user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()
	m.userStore.UpdateUser(ctx, user)

	// Generate session ID
	sessionID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("error generating session ID: %w", err)
	}

	// Generate tokens
	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}
	refreshToken, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	// Create session
	now := time.Now()
	session := &Session{
		ID:           sessionID,
		UserID:       user.ID,
		Token:        token,
		RefreshToken: refreshToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    now.Add(m.config.SessionTimeout),
		LastActivity: now,
		MFACompleted: !user.MFAEnabled, // If MFA is not enabled, mark as completed
		CreatedAt:    now,
		Metadata:     map[string]interface{}{},
	}
	// Store session
	if err := m.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	// Log the login
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   now,
		UserID:      user.ID,
		Action:      AuditActionLogin,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: fmt.Sprintf("User %s logged in", username),
		Severity:    AuditSeverityInfo,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		SessionID:   sessionID,
	})

	// If MFA is required, return the session but indicate MFA is needed
	if user.MFAEnabled {
		return session, ErrMFARequired
	}

	return session, nil
	

// Logout ends a user session
func (m *AuthManager) Logout(ctx context.Context, sessionID string) error {
	// Get the session
	session, err := m.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}

	// Delete the session
	if err := m.sessionStore.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	// Log the logout
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      session.UserID,
		Action:      AuditActionLogout,
		Resource:    "user",
		ResourceID:  session.UserID,
		Description: "User logged out",
		Severity:    AuditSeverityInfo,
		IPAddress:   session.IPAddress,
		UserAgent:   session.UserAgent,
		SessionID:   sessionID,
	})

	return nil

// VerifySession verifies a session token
func (m *AuthManager) VerifySession(ctx context.Context, token string) (bool, error) {
	// Find the session by token
	session, err := m.getSessionByToken(ctx, token)
	if err != nil {
		return false, err
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete the expired session
		m.sessionStore.DeleteSession(ctx, session.ID)
		return false, ErrSessionExpired
	}

	// Check if MFA is completed
	if !session.MFACompleted {
		return false, ErrMFARequired
	}

	// Update last activity
	session.LastActivity = time.Now()
	if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
		return false, fmt.Errorf("error updating session: %w", err)
	}

	return true, nil

// RefreshSession refreshes a session using the refresh token
func (m *AuthManager) RefreshSession(ctx context.Context, refreshToken string) (*Session, error) {
	// Find the session by refresh token
	session, err := m.getSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete the expired session
		m.sessionStore.DeleteSession(ctx, session.ID)
		return nil, ErrSessionExpired
	}

	// Generate new tokens
	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}
	newRefreshToken, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	// Update session
	session.Token = token
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(m.config.SessionTimeout)
	session.LastActivity = time.Now()
	if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("error updating session: %w", err)
	}

	// Log the refresh
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      session.UserID,
		Action:      AuditActionUpdate,
		Resource:    "user",
		ResourceID:  session.UserID,
		Description: "Session refreshed",
		Severity:    AuditSeverityInfo,
		IPAddress:   session.IPAddress,
		UserAgent:   session.UserAgent,
		SessionID:   session.ID,
	})
	return session, nil

// VerifyMFA verifies a multi-factor authentication code
func (m *AuthManager) VerifyMFA(ctx context.Context, sessionID, code string) error {
	// Get the session
	session, err := m.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}

	// Get the user
	user, err := m.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return fmt.Errorf("MFA is not enabled for this user")
	}
	// Verify the MFA code
	mfaMethod := mfa.MFAMethodTOTP // Default to TOTP
	if user.MFAMethod != "" {
		mfaMethod = mfa.MFAMethod(user.MFAMethod)
	}
	valid, err := m.mfaManager.VerifyMFA(ctx, user.ID, mfaMethod, code)
	if err != nil {
		return fmt.Errorf("error verifying MFA code: %w", err)
	}
	if !valid {
		// Log the failed MFA verification
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      user.ID,
			Action:      AuditActionUnauthorized,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: "MFA verification failed",
			Severity:    AuditSeverityMedium,
			IPAddress:   session.IPAddress,
			UserAgent:   session.UserAgent,
			SessionID:   sessionID,
		})
		return fmt.Errorf("invalid MFA code")
	}

	// Mark the session as MFA completed
	session.MFACompleted = true
	if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("error updating session: %w", err)
	}

	// Log the successful MFA verification
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      user.ID,
		Action:      AuditActionAuthorize,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "MFA verification successful",
		Severity:    AuditSeverityInfo,
		IPAddress:   session.IPAddress,
		UserAgent:   session.UserAgent,
		SessionID:   sessionID,
	})

	return nil

// EnableMFA enables multi-factor authentication for a user
func (m *AuthManager) EnableMFA(ctx context.Context, userID string, method common.AuthMethod) error {
	// Get the user
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Setup MFA based on method
	var secret string
	switch method {
	case common.AuthMethodTOTP:
		totpConfig, err := m.mfaManager.SetupTOTP(ctx, userID, user.Username)
		if err != nil {
			return fmt.Errorf("error setting up TOTP: %w", err)
		}
		secret = totpConfig.Secret
	case common.AuthMethodSMS:
		// For SMS, we don't generate a secret, just enable it
		// The phone number should be already set on the user
		if err := m.mfaManager.SetupSMS(ctx, userID, user.Email); err != nil {
			return fmt.Errorf("error setting up SMS: %w", err)
		}
	default:
		return fmt.Errorf("unsupported MFA method: %s", method)
	}

	// Update user
	user.MFAEnabled = true
	user.MFAMethod = string(method)
	user.MFAMethods = append(user.MFAMethods, string(method))
	if secret := os.Getenv("SECRET_KEY") {
		user.MFASecret = secret
	}
	user.UpdatedAt = time.Now()
	if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	// Log the action
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      getUserIDFromAuthContext(ctx),
		Action:      AuditActionUpdate,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("Enabled MFA (%s) for user %s", method, user.Username),
		Severity:    AuditSeverityInfo,
	})

	return nil

// DisableMFA disables multi-factor authentication for a user
func (m *AuthManager) DisableMFA(ctx context.Context, userID string, method common.AuthMethod) error {
	// Get the user
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return fmt.Errorf("MFA is not enabled for this user")
	}

	// Check if the method is enabled
	methodEnabled := false
	for _, m := range user.MFAMethods {
		if m == string(method) {
			methodEnabled = true
			break
		}
	}
	if !methodEnabled {
		return fmt.Errorf("MFA method %s is not enabled for this user", method)
	}

	// Remove the method from the list
	var newMethods []string
	for _, m := range user.MFAMethods {
		if m != string(method) {
			newMethods = append(newMethods, m)
		}
	}
	user.MFAMethods = newMethods

	// If no methods left, disable MFA
	if len(user.MFAMethods) == 0 {
		user.MFAEnabled = false
		user.MFAMethod = ""
		user.MFASecret = ""
	} else {
		// Set the first method as the default
		user.MFAMethod = user.MFAMethods[0]
	}

	// Update user
	user.UpdatedAt = time.Now()
	if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	// Log the action
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
		UserID:      getUserIDFromAuthContext(ctx),
		Action:      AuditActionUpdate,
		Resource:    "user",
		ResourceID:  userID,
		Description: fmt.Sprintf("Disabled MFA (%s) for user %s", method, user.Username),
		Severity:    AuditSeverityInfo,
	})

	return nil
// GetAllUsers returns all users
func (m *AuthManager) GetAllUsers(ctx context.Context) ([]*User, error) {
	return m.userStore.ListUsers(ctx)

// getSessionByToken retrieves a session by token
func (m *AuthManager) getSessionByToken(ctx context.Context, token string) (*Session, error) {
	// TODO: This is a temporary implementation. In production, we should have a way to
	// lookup sessions by token directly or maintain a token->sessionID map
	return nil, fmt.Errorf("getSessionByToken not implemented")

// getSessionByRefreshToken retrieves a session by refresh token
func (m *AuthManager) getSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	// TODO: This is a temporary implementation. In production, we should have a way to
	// lookup sessions by refresh token directly or maintain a refreshToken->sessionID map
	return nil, fmt.Errorf("getSessionByRefreshToken not implemented")

// Close closes the auth manager
func (m *AuthManager) Close() error {
	var errs []error

	// Close the user store
	if err := m.userStore.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing user store: %w", err))
	}

	// Close the session store
	if err := m.sessionStore.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing session store: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing auth manager: %v", errs)
	}

	return nil

// ValidateSession validates a session token and returns the session
func (m *AuthManager) ValidateSession(ctx context.Context, token string) (*Session, error) {
	// This is a simplified implementation that builds on VerifySession
	valid, err := m.VerifySession(ctx, token)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid session")
	}

	// For now, return a basic session. In a real implementation,
	// this would retrieve the actual session from storage
	return &Session{
		Token: token,
		// Other fields would be populated from actual session storage
	}, nil

// UpdateSession updates a session
func (m *AuthManager) UpdateSession(ctx context.Context, sessionID string, updates map[string]interface{}) error {
	// This is a placeholder implementation
	// In a real implementation, this would update the session in storage
	return nil

// HasRole checks if a user has a specific role
func (m *AuthManager) HasRole(ctx context.Context, userID string, role string) (bool, error) {
	user, err := m.GetUserByID(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, userRole := range user.Roles {
		if userRole == role {
			return true, nil
		}
	}

	return false, nil

// Helper functions

// generateID generates a unique ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil

// generateToken generates a random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil

// getUserIDFromAuthContext gets the user ID from the context
func getUserIDFromAuthContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userID, ok := ctx.Value("user_id").(string)
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
