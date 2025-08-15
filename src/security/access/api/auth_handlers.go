// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/perplext/LLMrecon/src/security/access"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`

// LoginResponse represents a login response
type LoginResponse struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	MFARequired  bool   `json:"mfa_required"`
	MFAMethods   []string `json:"mfa_methods,omitempty"`

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

	// Attempt login
	authManager := s.accessManager.GetAuthManager()
	result, err := authManager.Login(r.Context(), req.Username, req.Password, ip, userAgent)
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

	// Create response
	resp := LoginResponse{
		UserID:       result.User.ID,
		Username:     result.User.Username,
		Email:        result.User.Email,
		Token:        result.Session.Token,
		RefreshToken: result.Session.RefreshToken,
		ExpiresAt:    result.Session.ExpiresAt.Unix(),
		MFARequired:  result.MFARequired,
	}

	// Add MFA methods if MFA is required
	if result.MFARequired {
		resp.MFAMethods = result.User.MFAMethods
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Login successful", resp)

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
	// Logout
	authManager := s.accessManager.GetAuthManager()
	if err := authManager.Logout(r.Context(), token); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Logout successful", nil)
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
	// Get client information
	ip := getClientIP(r)
	userAgent := r.UserAgent()

	// Refresh token
	authManager := s.accessManager.GetAuthManager()
	result, err := authManager.RefreshToken(r.Context(), req.RefreshToken, ip, userAgent)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// Create response
	resp := LoginResponse{
		UserID:       result.User.ID,
		Username:     result.User.Username,
		Email:        result.User.Email,
		Token:        result.Session.Token,
		RefreshToken: result.Session.RefreshToken,
		ExpiresAt:    result.Session.ExpiresAt.Unix(),
		MFARequired:  result.MFARequired,
	}

	// Add MFA methods if MFA is required
	if result.MFARequired {
		resp.MFAMethods = result.User.MFAMethods
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Token refreshed", resp)

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
	// Validate token
	sessionManager := s.accessManager.GetSessionManager()
	session, err := sessionManager.ValidateToken(r.Context(), token)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Get user
	userManager := s.accessManager.GetUserManager()
	user, err := userManager.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Check if MFA is required but not completed
	mfaRequired := sessionManager.RequiresMFA(r.Context(), session)
	mfaCompleted := session.MFACompleted

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
		Roles:        user.Roles,
		MFARequired:  mfaRequired,
		MFACompleted: mfaCompleted,
		ExpiresAt:    session.ExpiresAt,
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Authentication status", resp)

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

	// Verify MFA
	authManager := s.accessManager.GetAuthManager()
	if err := authManager.VerifyMFA(r.Context(), req.Token, req.Method, req.Code); err != nil {
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
}
}
}
