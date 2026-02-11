// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	access "github.com/perplext/LLMrecon/src/security/access"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresAt    int64    `json:"expires_at"`
	MFARequired  bool     `json:"mfa_required"`
	MFAMethods   []string `json:"mfa_methods,omitempty"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// MFAVerifyRequest represents an MFA verification request
type MFAVerifyRequest struct {
	Token  string `json:"token"`
	Method string `json:"method"`
	Code   string `json:"code"`
}

// handleLogin handles user login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get client information
	ip := getClientIP(r)
	userAgent := r.UserAgent()

	// Attempt login via AccessControlManager
	session, err := s.accessManager.Login(r.Context(), req.Username, req.Password, ip, userAgent)
	if err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "invalid credentials"):
			WriteErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
		case strings.Contains(err.Error(), "account locked"):
			WriteErrorResponse(w, http.StatusForbidden, "Account is locked")
		case strings.Contains(err.Error(), "account inactive"):
			WriteErrorResponse(w, http.StatusForbidden, "Account is inactive")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	// Get the user for response
	user, err := s.accessManager.GetUser(r.Context(), session.UserID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user details")
		return
	}

	// Create response
	resp := LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Token:        session.Token,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt.Unix(),
		MFARequired:  user.MFAEnabled && !session.MFACompleted,
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Login successful", resp)
}

// handleLogout handles user logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authorization header is required")
		return
	}

	// Check if the header has the correct format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		WriteErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format, expected 'Bearer <token>'")
		return
	}

	token := parts[1]
	// Logout via AccessControlManager
	if err := s.accessManager.Logout(r.Context(), token); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Logout successful", nil)
}

// handleRefreshToken handles token refresh
func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.RefreshToken == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Refresh token via AccessControlManager
	session, err := s.accessManager.RefreshSession(r.Context(), req.RefreshToken)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// Get user for response
	user, err := s.accessManager.GetUser(r.Context(), session.UserID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user details")
		return
	}

	// Create response
	resp := LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Token:        session.Token,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt.Unix(),
		MFARequired:  user.MFAEnabled && !session.MFACompleted,
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Token refreshed", resp)
}

// handleAuthStatus handles authentication status check
func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authorization header is required")
		return
	}

	// Check if the header has the correct format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		WriteErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format, expected 'Bearer <token>'")
		return
	}

	token := parts[1]
	// Validate token via AccessControlManager
	valid, err := s.accessManager.ValidateSession(r.Context(), token)
	if err != nil || !valid {
		WriteErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Get the user from context if available, otherwise look up by auth manager
	authManager := s.accessManager.GetAuthManager()
	session, err := authManager.ValidateSession(r.Context(), token)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	user, err := authManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Create response
	resp := struct {
		UserID       string    `json:"user_id"`
		Username     string    `json:"username"`
		Email        string    `json:"email"`
		Roles        []string  `json:"roles"`
		MFARequired  bool      `json:"mfa_required"`
		MFACompleted bool      `json:"mfa_completed"`
		ExpiresAt    time.Time `json:"expires_at"`
	}{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Roles:        getRolesAsStrings(user),
		MFARequired:  user.MFAEnabled,
		MFACompleted: session.MFACompleted,
		ExpiresAt:    session.ExpiresAt,
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Authentication status", resp)
}

// handleMFAVerify handles MFA verification
func (s *Server) handleMFAVerify(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.Token == "" || req.Method == "" || req.Code == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Token, method, and code are required")
		return
	}

	// Verify MFA via AccessControlManager
	if err := s.accessManager.VerifyMFA(r.Context(), req.Token, req.Code); err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "invalid token"):
			WriteErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		case strings.Contains(err.Error(), "invalid code"):
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid MFA code")
		case strings.Contains(err.Error(), "unsupported method"):
			WriteErrorResponse(w, http.StatusBadRequest, "Unsupported MFA method")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "MFA verification failed")
		}
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "MFA verification successful", nil)
}

// getRolesAsStrings converts user roles to string slice
func getRolesAsStrings(user *access.User) []string {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = string(r)
	}
	return roles
}
