package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Auth handlers

// handleLogin handles user login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, NewAPIError(ErrCodeValidation, "Invalid request body"))
		return
	}
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// Authenticate user
	user, err := authService.Authenticate(request.Username, request.Password)
	if err != nil {
		if err == ErrInvalidCredentials {
			writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Invalid credentials"))
			return
		}
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Authentication failed"))
		return
	}
	
	// Generate JWT
	token, err := authService.GenerateJWT(user.ID, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to generate token"))
		return
	}
	
	// Return token and user info
	response := map[string]interface{}{
		"token": token,
		"user":  user,
	}
	writeSuccess(w, response)

// handleRefreshToken handles token refresh
func handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Missing authorization header"))
		return
	}
	
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Invalid authorization header"))
		return
	}
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// Refresh token
	newToken, err := authService.RefreshJWT(parts[1])
	if err != nil {
		if err == ErrTokenExpired || err == ErrTokenInvalid {
			writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, err.Error()))
			return
		}
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to refresh token"))
		return
	}
	
	response := map[string]interface{}{
		"token": newToken,
	}
	writeSuccess(w, response)

// handleCreateUser handles user creation
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var request CreateUserRequest
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, NewAPIError(ErrCodeValidation, "Invalid request body"))
		return
	}
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// Create user
	user, err := authService.CreateUser(request)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeError(w, http.StatusConflict, NewAPIError(ErrCodeConflict, err.Error()))
			return
		}
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to create user"))
		return
	}
	
	writeSuccess(w, user)
	

// handleUpdatePassword handles password update
func handleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, NewAPIError(ErrCodeValidation, "Invalid request body"))
		return
	}
	
	// Get user ID from JWT claims (would be set by JWT middleware)
	userID := r.Context().Value("userID").(string)
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// Update password
	if err := authService.UpdatePassword(userID, request.OldPassword, request.NewPassword); err != nil {
		if err == ErrInvalidCredentials {
			writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Invalid old password"))
			return
		}
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to update password"))
		return
	}
	
	response := map[string]interface{}{
		"message": "Password updated successfully",
	}
	writeSuccess(w, response)

// API Key handlers

// handleCreateAPIKey handles API key creation
func handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var request CreateAPIKeyRequest
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, NewAPIError(ErrCodeValidation, "Invalid request body"))
		return
	}
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// Create API key
	apiKey, err := authService.CreateAPIKey(request)
	if err != nil {
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to create API key"))
		return
	}
	
	writeSuccess(w, apiKey)

// handleListAPIKeys handles API key listing
func handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	filter := APIKeyFilter{}
	
	// Active filter
	if active := query.Get("active"); active != "" {
		activeBool := active == "true"
		filter.Active = &activeBool
	}
	
	// Expired filter
	filter.ExpiredOnly = query.Get("expired") == "true"
	
	// Revoked filter
	filter.RevokedOnly = query.Get("revoked") == "true"
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// List API keys
	apiKeys, err := authService.ListAPIKeys(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to list API keys"))
		return
	}
	
	writeSuccess(w, apiKeys)

// handleGetAPIKey handles getting a specific API key
func handleGetAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["id"]
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// Get API key
	apiKey, err := authService.GetAPIKey(keyID)
	if err != nil {
		if err == ErrAPIKeyNotFound {
			writeError(w, http.StatusNotFound, NewAPIError(ErrCodeNotFound, "API key not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to get API key"))
		return
	}
	
	writeSuccess(w, apiKey)

// handleRevokeAPIKey handles API key revocation
func handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID := vars["id"]
	
	// Get auth service from context
	authService := r.Context().Value("authService").(AuthService)
	
	// Revoke API key
	if err := authService.RevokeAPIKey(keyID); err != nil {
		if err == ErrAPIKeyNotFound {
			writeError(w, http.StatusNotFound, NewAPIError(ErrCodeNotFound, "API key not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, "Failed to revoke API key"))
		return
	}
	
	response := map[string]interface{}{
		"message": "API key revoked successfully",
	}
	writeSuccess(w, response)

// jwtMiddleware validates JWT tokens
func jwtMiddleware(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Missing authorization header"))
				return
			}
			
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Invalid authorization header"))
				return
			}
			
			// Validate token
			claims, err := authService.ValidateJWT(parts[1])
			if err != nil {
				writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, err.Error()))
				return
			}
			
			// Add claims to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "jwtClaims", claims)
			ctx = context.WithValue(ctx, "userID", claims.UserID)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}

// handleGetProfile returns the current user's profile
func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Get JWT claims from context
	claims, ok := r.Context().Value("jwtClaims").(*JWTClaims)
	if !ok {
		writeError(w, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, "Invalid authentication"))
		return
	}
	
	// Get auth service from context
	// authService := r.Context().Value("authService").(AuthService)
	
	// In a real implementation, we would fetch the full user profile using authService
	// For now, return basic info from JWT claims
	profile := map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
		"extra":    claims.Extra,
	}
	
