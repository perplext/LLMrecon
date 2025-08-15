// Package mfa provides multi-factor authentication functionality
package mfa

import (
	"context"
	"errors"
	"net/http"

	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// Common MFA-related errors
var (
	ErrMFARequired        = errors.New("multi-factor authentication required")
	ErrInvalidMFACode     = errors.New("invalid MFA code")
	ErrMFANotEnabled      = errors.New("MFA not enabled for user")
	ErrMFAAlreadyEnabled  = errors.New("MFA already enabled for user")
	ErrMFAMethodNotFound  = errors.New("MFA method not found")
	ErrMFASetupIncomplete = errors.New("MFA setup is incomplete")
)

// MFABoundaryEnforcer enforces MFA boundaries for requests
type MFABoundaryEnforcer struct {
	mfaManager MFAManager
	sessionManager interfaces.SessionStore

// NewMFABoundaryEnforcer creates a new MFA boundary enforcer
func NewMFABoundaryEnforcer(mfaManager MFAManager, sessionManager interfaces.SessionStore) *MFABoundaryEnforcer {
	return &MFABoundaryEnforcer{
		mfaManager:     mfaManager,
		sessionManager: sessionManager,
	}

// RequireMFA creates middleware that requires MFA completion
func (e *MFABoundaryEnforcer) RequireMFA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get session from context
		session, err := getSessionFromContext(ctx)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Check if MFA is completed
		if !session.MFACompleted {
			http.Error(w, "MFA required", http.StatusForbidden)
			return
		}
		
		next.ServeHTTP(w, r)
	})

// RequireMFAForUser creates middleware that requires MFA for specific users
func (e *MFABoundaryEnforcer) RequireMFAForUser(userID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get session from context
		session, err := getSessionFromContext(ctx)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Check if this is the target user
		if session.UserID == userID {
			// Check if MFA is completed
			if !session.MFACompleted {
				http.Error(w, "MFA required", http.StatusForbidden)
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})

// RequireMFAForRoles creates middleware that requires MFA for specific roles
func (e *MFABoundaryEnforcer) RequireMFAForRoles(roles []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get session from context
		session, err := getSessionFromContext(ctx)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Get user roles (this would need to be implemented based on your user store)
		userRoles, err := getUserRoles(ctx, session.UserID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		// Check if user has any of the specified roles
		requiresMFA := false
		for _, role := range roles {
			if contains(userRoles, role) {
				requiresMFA = true
				break
			}
		}
		
		// If user has a role that requires MFA, check if MFA is completed
		if requiresMFA && !session.MFACompleted {
			http.Error(w, "MFA required", http.StatusForbidden)
			return
		}
		
		next.ServeHTTP(w, r)
	})

// Helper function to get session from context
func getSessionFromContext(ctx context.Context) (*interfaces.Session, error) {
	// This is a placeholder - implement based on your session management
	sessionValue := ctx.Value("session")
	if sessionValue == nil {
		return nil, errors.New("no session in context")
	}
	
	session, ok := sessionValue.(*interfaces.Session)
	if !ok {
		return nil, errors.New("invalid session type in context")
	}
	
	return session, nil

// Helper function to get user roles
func getUserRoles(ctx context.Context, userID string) ([]string, error) {
	// This is a placeholder - implement based on your user store
	// For example:
	// user, err := userStore.GetUserByID(ctx, userID)
	// if err != nil {
	//     return nil, err
	// }
	// return user.Roles, nil
	return []string{}, nil

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
