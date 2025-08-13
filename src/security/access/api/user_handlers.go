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
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
)

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	Password    string            `json:"password"`
	Roles       []string          `json:"roles"`
	MFAEnabled  bool              `json:"mfa_enabled"`
	MFAMethods  []common.AuthMethod `json:"mfa_methods,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email       string            `json:"email,omitempty"`
	Roles       []string          `json:"roles,omitempty"`
	Active      *bool             `json:"active,omitempty"`
	MFAEnabled  *bool             `json:"mfa_enabled,omitempty"`
	MFAMethods  []common.AuthMethod `json:"mfa_methods,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ResetPasswordRequest represents a request to reset a user's password
type ResetPasswordRequest struct {
	Password string `json:"password"`
}

// ManageMFARequest represents a request to manage a user's MFA settings
type ManageMFARequest struct {
	Enabled  bool     `json:"enabled"`
	Methods  []string `json:"methods,omitempty"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID             string                 `json:"id"`
	Username       string                 `json:"username"`
	Email          string                 `json:"email"`
	Roles          []string               `json:"roles"`
	Permissions    []string               `json:"permissions,omitempty"`
	MFAEnabled     bool                   `json:"mfa_enabled"`
	MFAMethods     []string               `json:"mfa_methods,omitempty"`
	Active         bool                   `json:"active"`
	Locked         bool                   `json:"locked"`
	LastLogin      string                 `json:"last_login,omitempty"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// handleListUsers handles listing users
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserList) {
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
	userManager := s.accessManager.GetUserManager()
	users, err := userManager.ListUsers(r.Context())
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// Filter users by role if specified
	// Create a slice to hold users - could be either access.User or interfaces.User
	var filteredUsers []interface{}
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
	start := (page - 1) * limit
	end := start + limit
	if start >= len(filteredUsers) {
		start = 0
		end = 0
	}
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}
	
	paginatedUsers := filteredUsers
	if start < end {
		paginatedUsers = filteredUsers[start:end]
	} else {
		paginatedUsers = []interface{}{}
	}

	// Convert users to response format
	var userResponses []UserResponse
	for _, user := range paginatedUsers {
		userResponses = append(userResponses, convertUserToResponse(user))
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
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserCreate) {
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

	// Check if the current user has permission to assign the requested roles
	for _, role := range req.Roles {
		if !rbacManager.CanAssignRole(r.Context(), currentUser, role) {
			WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions to assign role: "+role)
			return
		}
	}

	// Create user
	// We need to check which type the UserManager expects
	userManager := s.accessManager.GetUserManager()
	
	// Try to use interfaces.User first
	user := &interfaces.User{
		Username:    req.Username,
		Email:       req.Email,
		Roles:       req.Roles,
		MFAEnabled:  req.MFAEnabled,
		MFAMethods:  req.MFAMethods,
		Permissions: req.Permissions,
		Active:      true,
		Metadata:    req.Metadata,
	}
	
	// Convert roles to access.Role for compatibility
	accessRoles := make([]access.Role, len(req.Roles))
	for i, role := range req.Roles {
		accessRoles[i] = access.Role(role)
	}
	
	// Convert permissions to access.Permission for compatibility
	accessPermissions := make([]access.Permission, len(req.Permissions))
	for i, perm := range req.Permissions {
		accessPermissions[i] = access.Permission(perm)
	}
	
	// Convert MFA methods to common.AuthMethod for compatibility
	accessMFAMethods := make([]common.AuthMethod, len(req.MFAMethods))
	for i, method := range req.MFAMethods {
		accessMFAMethods[i] = common.AuthMethod(method)
	}
	
	// Try to create the user
	var err error
	
	// First try with interfaces.User
	err = userManager.CreateUser(r.Context(), user, req.Password)
	
	// If that fails with a type error, try with access.User
	if err != nil && strings.Contains(err.Error(), "invalid type") {
		// Fall back to access.User
		accessUser := &access.User{
			Username:    req.Username,
			Email:       req.Email,
			Roles:       accessRoles,
			MFAEnabled:  req.MFAEnabled,
			MFAMethods:  accessMFAMethods,
			Permissions: accessPermissions,
			Active:      true,
			Metadata:    req.Metadata,
		}
		
		// Try again with access.User
		err = userManager.CreateUser(r.Context(), accessUser, req.Password)
		
		// If successful, update our user variable for the response
		if err == nil {
			user = &interfaces.User{
				ID:          accessUser.ID,
				Username:    accessUser.Username,
				Email:       accessUser.Email,
				Roles:       req.Roles, // Use original roles as strings
				Permissions: req.Permissions,
				MFAEnabled:  accessUser.MFAEnabled,
				MFAMethods:  req.MFAMethods,
				Active:      accessUser.Active,
				CreatedAt:   accessUser.CreatedAt,
				UpdatedAt:   accessUser.UpdatedAt,
				Metadata:    accessUser.Metadata,
			}
		}
	}
	
	// Handle any errors
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
	WriteSuccessResponse(w, http.StatusCreated, "User created successfully", convertUserToResponse(user))
}

// handleGetUser handles retrieving a user
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
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
	canViewOthers := rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserView)
	
	if !isSelf && !canViewOthers {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user
	userManager := s.accessManager.GetUserManager()
	user, err := userManager.GetUserByID(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "User retrieved successfully", convertUserToResponse(user))
}

// handleUpdateUser handles updating a user
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
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

	// Get user to update
	userManager := s.accessManager.GetUserManager()
	user, err := userManager.GetUserByID(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Check permissions
	rbacManager := s.accessManager.GetRBACManager()
	isSelf := currentUser.ID == userID
	canUpdateOthers := rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserUpdate)
	
	if !isSelf && !canUpdateOthers {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Additional permission checks for specific operations
	if len(req.Roles) > 0 {
		// Check if the current user has permission to update roles
		if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleAssign) {
			WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions to update roles")
			return
		}

		// Check if the current user has permission to assign the requested roles
		for _, role := range req.Roles {
			if !rbacManager.CanAssignRole(r.Context(), currentUser, role) {
				WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions to assign role: "+role)
				return
			}
		}
	}

	// Update user fields
	if req.Email != "" {
		user.Email = req.Email
	}
	
	if len(req.Roles) > 0 {
		user.Roles = req.Roles
	}
	
	if req.Active != nil {
		// Only admins can activate/deactivate accounts
		if !rbacManager.HasRole(r.Context(), currentUser, access.RoleAdmin) {
			WriteErrorResponse(w, http.StatusForbidden, "Only admins can activate/deactivate accounts")
			return
		}
		user.Active = *req.Active
	}
	
	if len(req.Permissions) > 0 {
		// Only admins can update direct permissions
		if !rbacManager.HasRole(r.Context(), currentUser, access.RoleAdmin) {
			WriteErrorResponse(w, http.StatusForbidden, "Only admins can update permissions")
			return
		}
		user.Permissions = req.Permissions
	}
	
	if req.Metadata != nil {
		user.Metadata = req.Metadata
	}

	// Update user
	if err := userManager.UpdateUser(r.Context(), user); err != nil {
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
	WriteSuccessResponse(w, http.StatusOK, "User updated successfully", convertUserToResponse(user))
}

// handleDeleteUser handles deleting a user
func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserDelete) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Prevent self-deletion
	if currentUser.ID == userID {
		WriteErrorResponse(w, http.StatusBadRequest, "Cannot delete your own account")
		return
	}

	// Delete user
	userManager := s.accessManager.GetUserManager()
	if err := userManager.DeleteUser(r.Context(), userID); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "User deleted successfully", nil)
}

// handleResetPassword handles resetting a user's password
func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
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
	canResetOthers := rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserResetPassword)
	
	if !isSelf && !canResetOthers {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Reset password
	userManager := s.accessManager.GetUserManager()
	if err := userManager.ResetPassword(r.Context(), userID, req.Password); err != nil {
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
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserLock) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Prevent self-locking
	if currentUser.ID == userID {
		WriteErrorResponse(w, http.StatusBadRequest, "Cannot lock your own account")
		return
	}

	// Lock user
	userManager := s.accessManager.GetUserManager()
	if err := userManager.LockUser(r.Context(), userID); err != nil {
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
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserUnlock) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Unlock user
	userManager := s.accessManager.GetUserManager()
	if err := userManager.UnlockUser(r.Context(), userID); err != nil {
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
	// Get current user from context - could be either access.User or interfaces.User
currentUserVal := r.Context().Value("user")
var currentUser *access.User
var ok bool

// Try to cast to access.User first
currentUser, ok = currentUserVal.(*access.User)
if !ok {
	// If that fails, try to cast to interfaces.User and convert
	if interfaceUser, ok := currentUserVal.(*interfaces.User); ok {
		// Convert interfaces.User to access.User for compatibility
		currentUser = &access.User{
			ID:                 interfaceUser.ID,
			Username:           interfaceUser.Username,
			Email:              interfaceUser.Email,
			MFAEnabled:         interfaceUser.MFAEnabled,
			FailedLoginAttempts: interfaceUser.FailedLoginAttempts,
			Locked:             interfaceUser.Locked,
			Active:             interfaceUser.Active,
			LastLogin:          interfaceUser.LastLogin,
			CreatedAt:          interfaceUser.CreatedAt,
			UpdatedAt:          interfaceUser.UpdatedAt,
		}
		ok = true
	}
}
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
	canManageOthersMFA := rbacManager.HasPermission(r.Context(), currentUser, access.PermissionUserManageMFA)
	
	if !isSelf && !canManageOthersMFA {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get user
	userManager := s.accessManager.GetUserManager()
	user, err := userManager.GetUserByID(r.Context(), userID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Update MFA settings
	user.MFAEnabled = req.Enabled
	if len(req.Methods) > 0 {
		user.MFAMethods = req.Methods
	}

	// Update user
	if err := userManager.UpdateUser(r.Context(), user); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update MFA settings")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "MFA settings updated successfully", nil)
}

// convertUserToResponse converts a user to a response format
func convertUserToResponse(user interface{}) UserResponse {
	// Handle both access.User and interfaces.User types
	switch u := user.(type) {
	case *access.User:
		var lastLogin string
		if !u.LastLogin.IsZero() {
			lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
		}

		// Convert roles and permissions to string slices
		roles := make([]string, len(u.Roles))
		for i, role := range u.Roles {
			roles[i] = string(role)
		}

		permissions := make([]string, len(u.Permissions))
		for i, perm := range u.Permissions {
			permissions[i] = string(perm)
		}

		// Extract MFA method names from common.AuthMethod objects
		mfaMethods := make([]string, len(u.MFAMethods))
		for i, method := range u.MFAMethods {
			mfaMethods[i] = string(method)
		}

		return UserResponse{
			ID:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			Roles:       roles,
			Permissions: permissions,
			MFAEnabled:  u.MFAEnabled,
			MFAMethods:  mfaMethods,
			Active:      u.Active,
			Locked:      u.Locked,
			LastLogin:   lastLogin,
			CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Metadata:    u.Metadata,
		}

	case *interfaces.User:
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

	default:
		// Return an empty response if the type is not recognized
		return UserResponse{}
	}
}
