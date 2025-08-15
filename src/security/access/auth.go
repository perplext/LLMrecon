// Package access provides access control and security auditing functionality
package access

import (
	"time"
	"os"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/mfa"
	"golang.org/x/crypto/bcrypt"
)

// Common errors - using the ones defined in interfaces package
var (
	ErrInvalidCredentials = interfaces.ErrInvalidCredentials
	ErrUserNotFound       = interfaces.ErrUserNotFound
	ErrUserLocked         = interfaces.ErrUserLocked
	ErrUserInactive       = interfaces.ErrUserInactive
	ErrMFARequired        = interfaces.ErrMFARequired
	ErrInvalidMFACode     = interfaces.ErrInvalidMFACode
	ErrSessionExpired     = interfaces.ErrSessionExpired
	ErrInvalidToken       = interfaces.ErrInvalidToken
)

// LegacyAuthManager manages authentication and session management (legacy version)
type LegacyAuthManager struct {
	config           *AccessControlConfig
	userStore        UserStore
	sessionStore     SessionStore
	auditLogger      AuditLogger
	mfaManager       mfa.MFAManager
	boundaryEnforcer *EnhancedContextBoundaryEnforcer
	mu               sync.RWMutex
}

// UserStore and SessionStore are now defined in store_interfaces.go

// NewLegacyAuthManager creates a new legacy authentication manager
func NewLegacyAuthManager(config *AccessControlConfig, userStore UserStore, sessionStore SessionStore, auditLogger AuditLogger) *LegacyAuthManager {
	// Create a mock MFA manager for backward compatibility
	mockMFAManager := mfa.NewMockMFAManager()

	// Create a boundary enforcer
	boundaryEnforcer := NewEnhancedContextBoundaryEnforcer()

	return &LegacyAuthManager{
		config:           config,
		userStore:        userStore,
		sessionStore:     sessionStore,
		auditLogger:      auditLogger,
		mfaManager:       mockMFAManager,
		boundaryEnforcer: boundaryEnforcer,
	}

// NewLegacyAuthManagerWithMFA creates a new legacy authentication manager with MFA support
func NewLegacyAuthManagerWithMFA(config *AccessControlConfig, userStore UserStore, sessionStore SessionStore, auditLogger AuditLogger, mfaManager mfa.MFAManager) *LegacyAuthManager {
	// Create a boundary enforcer
	boundaryEnforcer := NewEnhancedContextBoundaryEnforcer()

	return &LegacyAuthManager{
		config:           config,
		userStore:        userStore,
		sessionStore:     sessionStore,
		auditLogger:      auditLogger,
		mfaManager:       mfaManager,
		boundaryEnforcer: boundaryEnforcer,
	}

// Login authenticates a user and creates a new session
func (m *LegacyAuthManager) Login(ctx context.Context, username, password string, ipAddress, userAgent string) (*Session, error) {
	// Get user by username
	user, err := m.userStore.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Log failed login attempt
			m.auditLogger.LogAudit(ctx, &AuditLog{
				Timestamp:   time.Now(),
				Action:      AuditActionLogin,
				Resource:    "user",
				Description: "Failed login attempt: user not found",
				IPAddress:   ipAddress,
				UserAgent:   userAgent,
				Severity:    AuditSeverityMedium,
				Status:      "failed",
				Metadata: map[string]interface{}{
					"username": username,
				},
			})
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.Active {
		// Log failed login attempt
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
					UserID:      user.ID,
			Username:    user.Username,
			Action:      AuditActionLogin,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: "Failed login attempt: user inactive",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed",
		})
		return nil, ErrUserInactive
	}

	// Check if user is locked
	if user.Locked {
		// Log failed login attempt
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
					UserID:      user.ID,
			Username:    user.Username,
			Action:      AuditActionLogin,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: "Failed login attempt: user locked",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed",
		})
		return nil, ErrUserLocked
	}

	// Verify password
	if !m.verifyPassword(password, user.PasswordHash) {
		// Increment failed login attempts
		user.FailedLoginAttempts++

		// Check if account should be locked
		if m.config != nil && m.config.PasswordPolicy.LockoutThreshold > 0 &&
			user.FailedLoginAttempts >= m.config.PasswordPolicy.LockoutThreshold {
			user.Locked = true
		}

		// Update user
		if err := m.userStore.UpdateUser(ctx, user); err != nil {
			return nil, err
		}

		// Log failed login attempt
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
					UserID:      user.ID,
			Username:    user.Username,
			Action:      AuditActionLogin,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: "Failed login attempt: invalid password",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed",
			Metadata: map[string]interface{}{
				"failed_attempts": user.FailedLoginAttempts,
				"account_locked":  user.Locked,
			},
		})

		return nil, ErrInvalidCredentials
	}

	// Reset failed login attempts
	user.FailedLoginAttempts = 0
	user.LastLogin = time.Now()

	// Update user
	if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	// Check if MFA is required
	mfaRequired := false
	if m.config != nil && m.config.EnableMFA {
		if user.MFAEnabled {
			mfaRequired = true
		} else if len(m.config.MFARequiredRoles) > 0 {
			// Check if user has any of the roles that require MFA
			for _, role := range user.Roles {
				for _, requiredRole := range m.config.MFARequiredRoles {
					if role == requiredRole {
						mfaRequired = true
						break
					}
				}
				if mfaRequired {
					break
				}
			}
		}
	}

	// Create session
	session := &Session{
		ID:           generateRandomID(),
		UserID:       user.ID,
		Token:        generateRandomToken(),
		RefreshToken: generateRandomToken(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		MFACompleted: !mfaRequired,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Set expiration time
	tokenExpiration := 60 // Default 60 minutes
	if m.config != nil && m.config.SessionPolicy.TokenExpiration > 0 {
				tokenExpiration = m.config.SessionPolicy.TokenExpiration
	}
	session.ExpiresAt = time.Now().Add(time.Duration(tokenExpiration) * time.Minute)

	// Save session
	if err := m.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	// Log successful login
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
				UserID:      user.ID,
		Username:    user.Username,
		Action:      AuditActionLogin,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "Successful login",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		SessionID:   session.ID,
		Metadata: map[string]interface{}{
			"mfa_required":  mfaRequired,
			"mfa_completed": !mfaRequired,
		},
	})

	if mfaRequired && !session.MFACompleted {
		return session, ErrMFARequired
	}

	return session, nil

// VerifyMFA verifies a multi-factor authentication code
func (m *LegacyAuthManager) VerifyMFA(ctx context.Context, sessionID, code string) error {
	// Get session
	session, err := m.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Get user
	user, err := m.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return err
	}

	// Determine the MFA method from user settings
	method := common.AuthMethodTOTP // Default to TOTP
	if len(user.MFAMethods) > 0 {
		method = common.AuthMethod(user.MFAMethods[0]) // Use the first method in the list
	}

	// Verify MFA code using the MFA manager
	valid, err := m.mfaManager.VerifyMFA(ctx, user.ID, mfa.MFAMethod(method), code)
	if err != nil || !valid {
		// Log failed MFA attempt
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
					UserID:      user.ID,
			Username:    user.Username,
			Action:      AuditActionLogin,
			Resource:    "user",
			ResourceID:  user.ID,
			Description: "Failed MFA verification",
			IPAddress:   session.IPAddress,
			UserAgent:   session.UserAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed",
			SessionID:   session.ID,
		})
		return ErrInvalidMFACode
	}

	// Update session
	session.MFACompleted = true
	session.LastActivity = time.Now()
	if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
		return err
	}

	// Log successful MFA verification
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
				UserID:      user.ID,
		Username:    user.Username,
		Action:      AuditActionLogin,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "Successful MFA verification",
		IPAddress:   session.IPAddress,
		UserAgent:   session.UserAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		SessionID:   session.ID,
	})

	// Update the context with MFA status
	ctx = WithMFAStatus(ctx, MFAStatusCompleted)
	ctx = WithMFAUserID(ctx, user.ID)
	ctx = WithMFAMethod(ctx, method)

	// Enforce boundaries with the updated context
	if err := m.boundaryEnforcer.EnforceBoundaries(ctx); err != nil {
		return err
	}
	return nil

// Logout logs out a user by invalidating their session
func (m *LegacyAuthManager) Logout(ctx context.Context, sessionID string) error {
	// Get session
	session, err := m.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Get user
	user, err := m.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return err
	}

	// Delete session
	if err := m.sessionStore.DeleteSession(ctx, sessionID); err != nil {
		return err
	}

	// Log logout
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
				UserID:      user.ID,
		Username:    user.Username,
		Action:      AuditActionLogout,
		Resource:    "user",
		ResourceID:  user.ID,
		Description: "User logout",
		IPAddress:   session.IPAddress,
		UserAgent:   session.UserAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		SessionID:   session.ID,
	})

	return nil

// ValidateSession validates a session and returns the associated user
func (m *LegacyAuthManager) ValidateSession(ctx context.Context, sessionID, token string, ipAddress, userAgent string) (*User, error) {
	// Get session
	session, err := m.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify token
	if !m.verifyToken(token, session.Token) {
		return nil, ErrInvalidToken
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		m.sessionStore.DeleteSession(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	// Check if MFA is completed if required
	if !session.MFACompleted {
		return nil, ErrMFARequired
	}

	// Check IP binding if enabled
	if m.config != nil && m.config.SessionPolicy.EnforceIPBinding && session.IPAddress != ipAddress {
		// Log suspicious activity
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      session.UserID,
			Action:      AuditActionLogin,
			Resource:    "session",
			ResourceID:  session.ID,
			Description: "Session IP mismatch",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityHigh,
			Status:      "failed",
			SessionID:   session.ID,
			Metadata: map[string]interface{}{
				"expected_ip": session.IPAddress,
				"actual_ip":   ipAddress,
			},
		})
		return nil, ErrInvalidToken
	}

	// Check user agent binding if enabled
	if m.config != nil && m.config.SessionPolicy.EnforceUserAgentBinding && session.UserAgent != userAgent {
		// Log suspicious activity
		m.auditLogger.LogAudit(ctx, &AuditLog{
			Timestamp:   time.Now(),
			UserID:      session.UserID,
			Action:      AuditActionLogin,
			Resource:    "session",
			ResourceID:  session.ID,
			Description: "Session user agent mismatch",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    AuditSeverityMedium,
			Status:      "failed",
			SessionID:   session.ID,
			Metadata: map[string]interface{}{
				"expected_user_agent": session.UserAgent,
				"actual_user_agent":   userAgent,
			},
		})
		return nil, ErrInvalidToken
	}

	// Check inactivity timeout
	if m.config != nil && m.config.SessionPolicy.InactivityTimeout > 0 {
		inactivityTimeout := time.Duration(m.config.SessionPolicy.InactivityTimeout) * time.Minute
		if time.Since(session.LastActivity) > inactivityTimeout {
			// Delete inactive session
			m.sessionStore.DeleteSession(ctx, sessionID)
			return nil, ErrSessionExpired
		}
	}

	// Get user
	user, err := m.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	// Check if user is still active
	if !user.Active {
		// Delete session
		m.sessionStore.DeleteSession(ctx, sessionID)
		return nil, ErrUserInactive
	}

	// Check if user is locked
	if user.Locked {
		// Delete session
		m.sessionStore.DeleteSession(ctx, sessionID)
		return nil, ErrUserLocked
	}

	// Update last activity
	session.LastActivity = time.Now()
	if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return user, nil

// RefreshSession refreshes a session and returns a new token
func (m *LegacyAuthManager) RefreshSession(ctx context.Context, sessionID, refreshToken string, ipAddress, userAgent string) (*Session, error) {
	// Get session
	session, err := m.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify refresh token
	if !m.verifyToken(refreshToken, session.RefreshToken) {
		return nil, ErrInvalidToken
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		m.sessionStore.DeleteSession(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	// Get user
	user, err := m.userStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Check if user is still active
	if !user.Active {
		// Delete session
		m.sessionStore.DeleteSession(ctx, sessionID)
		return nil, ErrUserInactive
	}

	// Check if user is locked
	if user.Locked {
		// Delete session
		m.sessionStore.DeleteSession(ctx, sessionID)
		return nil, ErrUserLocked
	}

	// Generate new tokens
	session.Token = generateRandomToken()
	session.RefreshToken = generateRandomToken()
	session.LastActivity = time.Now()

	// Set new expiration time
	tokenExpiration := 60 // Default 60 minutes
	if m.config != nil && m.config.SessionPolicy.TokenExpiration > 0 {
				tokenExpiration = m.config.SessionPolicy.TokenExpiration
	}
	session.ExpiresAt = time.Now().Add(time.Duration(tokenExpiration) * time.Minute)

	// Update session
	if err := m.sessionStore.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	// Log session refresh
	m.auditLogger.LogAudit(ctx, &AuditLog{
		Timestamp:   time.Now(),
				UserID:      user.ID,
		Username:    user.Username,
		Action:      AuditActionLogin,
		Resource:    "session",
		ResourceID:  session.ID,
		Description: "Session refreshed",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    AuditSeverityInfo,
		Status:      "success",
		SessionID:   session.ID,
	})

	return session, nil

// CreateUser creates a new user
func (m *LegacyAuthManager) CreateUser(ctx context.Context, username, email, password string, roles []string) (*User, error) {
	// Check if username already exists
	if _, err := m.userStore.GetUserByUsername(ctx, username); err == nil {
		return nil, fmt.Errorf("username already exists")
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// Check if email already exists
	if _, err := m.userStore.GetUserByEmail(ctx, email); err == nil {
		return nil, fmt.Errorf("email already exists")
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// Hash password
	passwordHash, err := m.hashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &User{
		ID:                 generateRandomID(),
		Username:           username,
		Email:              email,
		PasswordHash:       passwordHash,
		Roles:              roles,
		MFAEnabled:         false,
		LastPasswordChange: time.Now(),
		Active:             true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Save user
	if err := m.userStore.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil

// UpdateUserPassword updates a user's password
func (m *LegacyAuthManager) UpdateUserPassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !m.verifyPassword(currentPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	// Check password policy
	if err := m.validatePasswordPolicy(newPassword); err != nil {
		return err
	}

	// Hash new password
	passwordHash, err := m.hashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user
	user.PasswordHash = passwordHash
	user.LastPasswordChange = time.Now()
	user.UpdatedAt = time.Now()

	// Save user
	    if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return err
	}
	// Invalidate all sessions
	sessions, err := m.sessionStore.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		m.sessionStore.DeleteSession(ctx, session.ID)
	}

	return nil

// EnableMFA enables multi-factor authentication for a user
func (m *LegacyAuthManager) EnableMFA(ctx context.Context, userID string, method common.AuthMethod) error {
	// Get user
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if MFA is already enabled
	if user.MFAEnabled {
		// Check if method is already enabled
		for _, m := range user.MFAMethods {
			if m == string(method) {
				return nil
			}
		}

		// Add method
		user.MFAMethods = append(user.MFAMethods, string(method))
	} else {
		// Enable MFA
		user.MFAEnabled = true
		user.MFAMethods = []string{string(method)}
	}
	// Update user
	user.UpdatedAt = time.Now()
	    if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return err
	}

	return nil

// DisableMFA disables multi-factor authentication for a user
func (m *LegacyAuthManager) DisableMFA(ctx context.Context, userID string, method common.AuthMethod) error {
	// Get user
	user, err := m.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return nil
	}

	// Remove method
	methods := make([]string, 0, len(user.MFAMethods))
	for _, m := range user.MFAMethods {
		if m != string(method) {
			methods = append(methods, m)
		}
	}

	// Update user
	user.MFAMethods = methods
	if len(methods) == 0 {
		user.MFAEnabled = false
	}
	user.UpdatedAt = time.Now()
	    if err := m.userStore.UpdateUser(ctx, user); err != nil {
		return err
	}

	return nil

// Helper functions

// hashPassword hashes a password using bcrypt
func (m *LegacyAuthManager) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil

// verifyPassword verifies a password against a hash
func (m *LegacyAuthManager) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil

// verifyToken verifies a token using constant-time comparison
func (m *LegacyAuthManager) verifyToken(token, expectedToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1

// verifyMFACode verifies an MFA code
func (m *LegacyAuthManager) verifyMFACode(user *User, code string) bool {
	// Enhanced MFA verification
	// This is a placeholder implementation that will be replaced with a more robust solution
	// when the MFA manager is fully integrated

	// If no MFA methods are configured, fall back to simple verification
	if len(user.MFAMethods) == 0 {
		return code == "123456"
	}

	// Check which MFA method is being used
	method := common.AuthMethodTOTP
	if len(user.MFAMethods) > 0 {
		method = common.AuthMethod(user.MFAMethods[0])
	}

	// Verify based on method
	switch method {
	case common.AuthMethodTOTP:
		// TOTP verification (time-based one-time password)
		// In a real implementation, this would validate against the user's TOTP secret
		// using a proper TOTP algorithm with time window validation
		return verifyTOTPCode(user.MFASecret, code)

	case common.AuthMethodBackupCode:
		// Backup code verification
		// In a real implementation, this would check against stored backup codes
		return code == "BACKUP-123456"

	case common.AuthMethodWebAuthn:
		// WebAuthn verification
		// This would normally validate a WebAuthn assertion
		// For now, we'll just return false as this requires browser integration
		return false

	case common.AuthMethodSMS:
		// SMS verification
		// In a real implementation, this would check against a recently sent SMS code
		return code == "SMS-123456"

	default:
		// Unknown method, fall back to simple verification
		return code == "123456"
	}

// verifyTOTPCode verifies a TOTP code against a secret
func verifyTOTPCode(secret, code string) bool {
	// If no secret is provided, fall back to simple verification
	if secret == "" {
		return code == "123456"
	}

	// Decode secret
	secret = strings.TrimRight(secret, "=")
	missingPadding := len(secret) % 8
	if missingPadding > 0 {
		secret := os.Getenv("SECRET_KEY"), 8-missingPadding)
	}

	secretBytes, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		// If there's an error decoding the secret, fall back to simple verification
		return code == "123456"
	}

	// Get current time
	now := time.Now().Unix()

	// Check codes within time window (30 seconds before and after)
	for offset := -1; offset <= 1; offset++ {
		// Calculate counter
		counter := uint64((now / 30) + int64(offset))

		// Convert counter to bytes
		counterBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(counterBytes, counter)

		// Calculate HMAC
		h := hmac.New(sha1.New, secretBytes)
		h.Write(counterBytes)
		hash := h.Sum(nil)

		// Dynamic truncation
		offset := hash[len(hash)-1] & 0x0F
		binaryCode := ((int(hash[offset]) & 0x7F) << 24) |
			((int(hash[offset+1]) & 0xFF) << 16) |
			((int(hash[offset+2]) & 0xFF) << 8) |
			(int(hash[offset+3]) & 0xFF)

		// Generate 6-digit code
		otp := binaryCode % 1000000
		generatedCode := fmt.Sprintf("%06d", otp)

		// Compare codes
		if generatedCode == code {
			return true
		}
	}

	return false

// validatePasswordPolicy validates a password against the password policy
func (m *LegacyAuthManager) validatePasswordPolicy(password string) error {
	if m.config == nil || m.config.PasswordPolicy.MinLength == 0 {
		// No policy or minimal policy
		return nil
	}

	policy := m.config.PasswordPolicy

	// Check length
	if len(password) < policy.MinLength {
		return fmt.Errorf("password must be at least %d characters long", policy.MinLength)
	}

	// Check uppercase
	if policy.RequireUppercase {
		hasUpper := false
		for _, c := range password {
			if c >= 'A' && c <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
	}

	// Check lowercase
	if policy.RequireLowercase {
		hasLower := false
		for _, c := range password {
			if c >= 'a' && c <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return fmt.Errorf("password must contain at least one lowercase letter")
		}
	}

	// Check numbers
	if policy.RequireNumbers {
		hasNumber := false
		for _, c := range password {
			if c >= '0' && c <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return fmt.Errorf("password must contain at least one number")
		}
	}

	// Check special characters
	if policy.RequireSpecialChars {
		hasSpecial := false
		for _, c := range password {
			if strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", c) {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	return nil

// generateRandomID generates a random ID
func generateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)

// generateRandomToken generates a random token
func generateRandomToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
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
