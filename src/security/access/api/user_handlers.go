// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	access "github.com/perplext/LLMrecon/src/security/access"
	"github.com/perplext/LLMrecon/src/security/access/common"
)

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Password    string                 `json:"password"`
	Roles       []string               `json:"roles"`
	MFAEnabled  bool                   `json:"mfa_enabled"`
	Permissions []string               `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email       string                 `json:"email,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Active      *bool                  `json:"active,omitempty"`
	MFAEnabled  *bool                  `json:"mfa_enabled,omitempty"`
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
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserList)
	if err != nil || !hasPermission {
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

	// Get users via AccessControlManager
	users, err := s.accessManager.ListUsers(r.Context())
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// Get role filter
	roleFilter := query.Get("role")

	// Filter users by role if specified
	var filteredUsers []*access.User
	if roleFilter != "" {
		for _, user := range users {
			for _, role := range user.Roles {
				if string(role) == roleFilter {
					filteredUsers = append(filteredUsers, user)
					break
				}
			}
		}
	} else {
		filteredUsers = users
	}

	// Paginate results
	start := (page - 1) * limit
	end := start + limit
	if start >= len(filteredUsers) {
		start = 0
		end = 0
	}
	if end > len(filteredUsers) {
		end = len(filteredUsers)
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
		TotalCount: len(filteredUsers),
		Page:       page,
		Limit:      limit,
		TotalPages: (len(filteredUsers) + limit - 1) / limit,
	}

	// Return success response
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
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserCreate)
	if err != nil || !hasPermission {
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

	// Create user via AccessControlManager
	user, err := s.accessManager.CreateUser(r.Context(), req.Username, req.Email, req.Password, req.Roles, currentUser.ID)
	if err != nil {
		// Handle specific error types
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

	// Return success response
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

	// Check permission (can view others or self)
	rbacManager := s.accessManager.GetRBACManager()
	isSelf := currentUser.ID == userID
	hasViewPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserView)
	canViewOthers := err == nil && hasViewPermission

	if !isSelf && !canViewOthers {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user via AccessControlManager
	user, err := s.accessManager.GetUser(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Return success response
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

	// Check permissions
	rbacManager := s.accessManager.GetRBACManager()
	isSelf := currentUser.ID == userID
	hasUpdatePermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserUpdate)
	canUpdateOthers := err == nil && hasUpdatePermission

	if !isSelf && !canUpdateOthers {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get current user data
	user, err := s.accessManager.GetUser(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Update fields
	email := user.Email
	if req.Email != "" {
		email = req.Email
	}

	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = string(r)
	}
	if len(req.Roles) > 0 {
		// Check role assignment permission
		hasRoleAssign, roleErr := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleAssign)
		if roleErr != nil || !hasRoleAssign {
			WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions to update roles")
			return
		}
		roles = req.Roles
	}

	active := user.Active
	if req.Active != nil {
		// Only admins can activate/deactivate accounts
		hasAdminRole, roleErr := rbacManager.HasRole(currentUser.ID, access.RoleAdmin)
		if roleErr != nil || !hasAdminRole {
			WriteErrorResponse(w, http.StatusForbidden, "Only admins can activate/deactivate accounts")
			return
		}
		active = *req.Active
	}

	// Update user via AccessControlManager
	updatedUser, err := s.accessManager.UpdateUser(r.Context(), userID, user.Username, email, roles, active, currentUser.ID)
	if err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "email already exists"):
			WriteErrorResponse(w, http.StatusConflict, "Email already exists")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	// Return success response
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

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserDelete)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Prevent self-deletion
	if currentUser.ID == userID {
		WriteErrorResponse(w, http.StatusBadRequest, "Cannot delete your own account")
		return
	}

	// Delete user via AccessControlManager
	if err := s.accessManager.DeleteUser(r.Context(), userID, currentUser.ID); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// Return success response
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

	// Check permissions
	rbacManager := s.accessManager.GetRBACManager()
	isSelf := currentUser.ID == userID
	hasResetPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserResetPassword)
	canResetOthers := err == nil && hasResetPermission

	if !isSelf && !canResetOthers {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Reset password via AccessControlManager
	if err := s.accessManager.ResetPassword(r.Context(), userID, req.Password, currentUser.ID); err != nil {
		// Handle specific error types
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

	// Return success response
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

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserLock)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Prevent self-locking
	if currentUser.ID == userID {
		WriteErrorResponse(w, http.StatusBadRequest, "Cannot lock your own account")
		return
	}

	// Lock user via AccessControlManager
	if err := s.accessManager.LockUser(r.Context(), userID, currentUser.ID, "Locked via API"); err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "user not found"):
			WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to lock user")
		}
		return
	}

	// Return success response
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

	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserUnlock)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Unlock user via AccessControlManager
	if err := s.accessManager.UnlockUser(r.Context(), userID, currentUser.ID); err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "user not found"):
			WriteErrorResponse(w, http.StatusNotFound, "User not found")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to unlock user")
		}
		return
	}

	// Return success response
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

	// Check permissions
	rbacManager := s.accessManager.GetRBACManager()
	isSelf := currentUser.ID == userID
	hasMFAPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionUserManageMFA)
	canManageOthersMFA := err == nil && hasMFAPermission

	if !isSelf && !canManageOthersMFA {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user via AccessControlManager
	_, err = s.accessManager.GetUser(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Enable/disable MFA via AccessControlManager
	// Note: The MFA methods are handled via separate enable/disable calls
	if req.Enabled {
		// Enable MFA - use a default method if none specified
		if err := s.accessManager.EnableMFA(r.Context(), userID, common.AuthMethodTOTP, currentUser.ID); err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update MFA settings")
			return
		}
	} else {
		if err := s.accessManager.DisableMFA(r.Context(), userID, common.AuthMethodTOTP, currentUser.ID); err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update MFA settings")
			return
		}
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "MFA settings updated successfully", nil)
}

// convertAccessUserToResponse converts an access.User to a UserResponse
func convertAccessUserToResponse(user *access.User) UserResponse {
	var lastLogin string
	if !user.LastLogin.IsZero() {
		lastLogin = user.LastLogin.Format("2006-01-02T15:04:05Z")
	}

	// Convert roles to string slice
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = string(role)
	}

	// Convert MFA methods to string slice
	mfaMethods := make([]string, len(user.MFAMethods))
	for i, method := range user.MFAMethods {
		mfaMethods[i] = string(method)
	}

	return UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Roles:      roles,
		MFAEnabled: user.MFAEnabled,
		MFAMethods: mfaMethods,
		Active:     user.Active,
		Locked:     user.Locked,
		LastLogin:  lastLogin,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
