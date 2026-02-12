// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/perplext/LLMrecon/src/security/access"
	"github.com/perplext/LLMrecon/src/security/access/common"
)

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Password    string                 `json:"password"`
	Roles       []string               `json:"roles"`
	MFAEnabled  bool                   `json:"mfa_enabled"`
	MFAMethods  []string               `json:"mfa_methods,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email       string                 `json:"email,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Active      *bool                  `json:"active,omitempty"`
	MFAEnabled  *bool                  `json:"mfa_enabled,omitempty"`
	MFAMethods  []string               `json:"mfa_methods,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ResetPasswordRequest represents a request to reset a user's password
type ResetPasswordRequest struct {
	Password string `json:"password"`
}

// ManageMFARequest represents a request to manage a user's MFA settings
type ManageMFARequest struct {
	Enabled bool     `json:"enabled"`
	Methods []string `json:"methods,omitempty"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions,omitempty"`
	MFAEnabled  bool                   `json:"mfa_enabled"`
	MFAMethods  []string               `json:"mfa_methods,omitempty"`
	Active      bool                   `json:"active"`
	Locked      bool                   `json:"locked"`
	LastLogin   string                 `json:"last_login,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// handleListUsers handles listing users
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserList) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Get page and limit parameters
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get role filter
	roleFilter := query.Get("role")

	// Get users
	users, err := s.accessManager.ListUsers(r.Context())
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// Filter users by role if specified
	var filteredUsers []*access.User
	if roleFilter != "" {
		for _, user := range users {
			for _, role := range user.Roles {
				if role == roleFilter {
					filteredUsers = append(filteredUsers, user)
					break
				}
			}
		}
	} else {
		filteredUsers = users
	}

	// Paginate results
	total := len(filteredUsers)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		start = 0
		end = 0
	}
	if end > total {
		end = total
	}

	var paginatedUsers []*access.User
	if start < end {
		paginatedUsers = filteredUsers[start:end]
	}

	// Convert users to response format
	var userResponses []UserResponse
	for _, user := range paginatedUsers {
		userResponses = append(userResponses, convertAccessUserToResponse(user))
	}

	// Create response
	resp := struct {
		Users      []UserResponse `json:"users"`
		TotalCount int            `json:"total_count"`
		Page       int            `json:"page"`
		Limit      int            `json:"limit"`
		TotalPages int            `json:"total_pages"`
	}{
		Users:      userResponses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	}

	WriteSuccessResponse(w, http.StatusOK, "Users retrieved successfully", resp)
}

// handleCreateUser handles creating a new user
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserCreate) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse request
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Create user using AccessControlManager
	user, err := s.accessManager.CreateUser(
		r.Context(),
		req.Username,
		req.Email,
		req.Password,
		req.Roles,
		currentUser.ID,
	)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "username already exists"):
			WriteErrorResponse(w, http.StatusConflict, "Username already exists")
		case strings.Contains(err.Error(), "email already exists"):
			WriteErrorResponse(w, http.StatusConflict, "Email already exists")
		case strings.Contains(err.Error(), "password policy"):
			WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	WriteSuccessResponse(w, http.StatusCreated, "User created successfully", convertAccessUserToResponse(user))
}

// handleGetUser handles retrieving a user
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check permission (can view self or others with permission)
	isSelf := currentUser.ID == userID
	if !isSelf && !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserView) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user
	user, err := s.accessManager.GetUser(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	WriteSuccessResponse(w, http.StatusOK, "User retrieved successfully", convertAccessUserToResponse(user))
}

// handleUpdateUser handles updating a user
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Parse request
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get existing user
	existingUser, err := s.accessManager.GetUser(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Check permissions
	isSelf := currentUser.ID == userID
	if !isSelf && !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserUpdate) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Additional permission checks for role assignment
	if len(req.Roles) > 0 {
		if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserRoleAssign) {
			WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions to update roles")
			return
		}
	}

	// Additional permission check for active status changes (admin only)
	if req.Active != nil {
		hasAdmin := false
		for _, role := range currentUser.Roles {
			if role == access.RoleAdmin {
				hasAdmin = true
				break
			}
		}
		if !hasAdmin {
			WriteErrorResponse(w, http.StatusForbidden, "Only admins can activate/deactivate accounts")
			return
		}
	}

	// Build update fields
	email := existingUser.Email
	if req.Email != "" {
		email = req.Email
	}

	roles := existingUser.Roles
	if len(req.Roles) > 0 {
		roles = req.Roles
	}

	active := existingUser.Active
	if req.Active != nil {
		active = *req.Active
	}

	// Update user using AccessControlManager
	updatedUser, err := s.accessManager.UpdateUser(
		r.Context(),
		userID,
		existingUser.Username,
		email,
		roles,
		active,
		currentUser.ID,
	)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "email already exists"):
			WriteErrorResponse(w, http.StatusConflict, "Email already exists")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	WriteSuccessResponse(w, http.StatusOK, "User updated successfully", convertAccessUserToResponse(updatedUser))
}

// handleDeleteUser handles deleting a user
func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserDelete) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Prevent self-deletion
	if currentUser.ID == userID {
		WriteErrorResponse(w, http.StatusBadRequest, "Cannot delete your own account")
		return
	}

	// Delete user
	if err := s.accessManager.DeleteUser(r.Context(), userID, currentUser.ID); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	WriteSuccessResponse(w, http.StatusOK, "User deleted successfully", nil)
}

// handleResetPassword handles resetting a user's password
func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Parse request
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.Password == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Check permissions (can reset own or others with permission)
	isSelf := currentUser.ID == userID
	if !isSelf && !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserResetPassword) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Reset password
	if err := s.accessManager.ResetPassword(r.Context(), userID, req.Password, currentUser.ID); err != nil {
		switch {
		case strings.Contains(err.Error(), "user not found"):
			WriteErrorResponse(w, http.StatusNotFound, "User not found")
		case strings.Contains(err.Error(), "password policy"):
			WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to reset password")
		}
		return
	}

	WriteSuccessResponse(w, http.StatusOK, "Password reset successfully", nil)
}

// handleLockUser handles locking a user account
func (s *Server) handleLockUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserLock) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Prevent self-locking
	if currentUser.ID == userID {
		WriteErrorResponse(w, http.StatusBadRequest, "Cannot lock your own account")
		return
	}

	// Lock user
	if err := s.accessManager.LockUser(r.Context(), userID, currentUser.ID, "Locked by admin"); err != nil {
		switch {
		case strings.Contains(err.Error(), "user not found"):
			WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to lock user")
		}
		return
	}

	WriteSuccessResponse(w, http.StatusOK, "User locked successfully", nil)
}

// handleUnlockUser handles unlocking a user account
func (s *Server) handleUnlockUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserUnlock) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Unlock user
	if err := s.accessManager.UnlockUser(r.Context(), userID, currentUser.ID); err != nil {
		switch {
		case strings.Contains(err.Error(), "user not found"):
			WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to unlock user")
		}
		return
	}

	WriteSuccessResponse(w, http.StatusOK, "User unlocked successfully", nil)
}

// handleManageUserMFA handles managing a user's MFA settings
func (s *Server) handleManageUserMFA(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Parse request
	var req ManageMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Check permissions (can manage own or others with permission)
	isSelf := currentUser.ID == userID
	if !isSelf && !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionUserManageMFA) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Enable or disable MFA
	if req.Enabled {
		// Default to TOTP if no method specified
		method := common.AuthMethod("totp")
		if len(req.Methods) > 0 {
			method = common.AuthMethod(req.Methods[0])
		}
		if err := s.accessManager.EnableMFA(r.Context(), userID, method, currentUser.ID); err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to enable MFA")
			return
		}
	} else {
		method := common.AuthMethod("totp")
		if len(req.Methods) > 0 {
			method = common.AuthMethod(req.Methods[0])
		}
		if err := s.accessManager.DisableMFA(r.Context(), userID, method, currentUser.ID); err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to disable MFA")
			return
		}
	}

	WriteSuccessResponse(w, http.StatusOK, "MFA settings updated successfully", nil)
}

// convertAccessUserToResponse converts an access.User to a response format
func convertAccessUserToResponse(u *access.User) UserResponse {
	var lastLogin string
	if !u.LastLogin.IsZero() {
		lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
	}

	return UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Roles:       u.Roles,
		Permissions: u.Permissions,
		MFAEnabled:  u.MFAEnabled,
		MFAMethods:  u.MFAMethods,
		Active:      u.Active,
		Locked:      u.Locked,
		LastLogin:   lastLogin,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Metadata:    u.Metadata,
	}
}

// convertUserToResponse converts a user interface{} to a response format (backward compatibility)
func convertUserToResponse(user interface{}) UserResponse {
	switch u := user.(type) {
	case *access.User:
		return convertAccessUserToResponse(u)
	default:
		return UserResponse{}
	}
}
