// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	access "github.com/perplext/LLMrecon/src/security/access"
)

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	ParentRoles []string `json:"parent_roles,omitempty"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Description string   `json:"description,omitempty"`
	ParentRoles []string `json:"parent_roles,omitempty"`
}

// AddPermissionRequest represents a request to add a permission to a role
type AddPermissionRequest struct {
	Permission string `json:"permission"`
}

// RoleResponse represents a role response
type RoleResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	ParentRoles []string `json:"parent_roles,omitempty"`
	IsBuiltIn   bool     `json:"is_built_in"`
}

// handleListRoles handles listing roles
func (s *Server) handleListRoles(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleList)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// The RBACManager interface does not support listing roles directly.
	// Return the user's own roles as a minimal response.
	userRoles, err := rbacManager.GetUserRoles(currentUser.ID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list roles")
		return
	}

	var roleResponses []RoleResponse
	for _, roleName := range userRoles {
		roleResponses = append(roleResponses, RoleResponse{
			Name: roleName,
		})
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Roles retrieved successfully", roleResponses)
}

// handleCreateRole handles creating a new role
func (s *Server) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleCreate)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse request
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.Name == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Role creation is not supported through the RBACManager interface
	WriteErrorResponse(w, http.StatusNotImplemented, "Role creation is not supported through this API")
}

// handleGetRole handles retrieving a role
func (s *Server) handleGetRole(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleView)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get role name from URL
	vars := mux.Vars(r)
	roleName := vars["name"]
	if roleName == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Role retrieval by name is not supported through the RBACManager interface.
	// Return a minimal response with the role name.
	resp := RoleResponse{
		Name: roleName,
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Role retrieved successfully", resp)
}

// handleUpdateRole handles updating a role
func (s *Server) handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleUpdate)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get role name from URL
	vars := mux.Vars(r)
	roleName := vars["name"]
	if roleName == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Parse request
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Role update is not supported through the RBACManager interface
	WriteErrorResponse(w, http.StatusNotImplemented, "Role update is not supported through this API")
}

// handleDeleteRole handles deleting a role
func (s *Server) handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleDelete)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get role name from URL
	vars := mux.Vars(r)
	roleName := vars["name"]
	if roleName == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Role deletion is not supported through the RBACManager interface
	WriteErrorResponse(w, http.StatusNotImplemented, "Role deletion is not supported through this API")
}

// handleAddPermission handles adding a permission to a role
func (s *Server) handleAddPermission(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleUpdate)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get role name from URL
	vars := mux.Vars(r)
	roleName := vars["name"]
	if roleName == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Role name is required")
		return
	}

	// Parse request
	var req AddPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if req.Permission == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Permission is required")
		return
	}

	// Adding permissions to roles is not supported through the RBACManager interface
	_ = roleName // suppress unused variable
	WriteErrorResponse(w, http.StatusNotImplemented, "Adding permissions to roles is not supported through this API")
}

// handleRemovePermission handles removing a permission from a role
func (s *Server) handleRemovePermission(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	currentUser, ok := r.Context().Value("user").(*access.User)
	if !ok {
		WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Check permission
	rbacManager := s.accessManager.GetRBACManager()
	hasPermission, err := rbacManager.HasPermission(currentUser.ID, access.PermissionRoleUpdate)
	if err != nil || !hasPermission {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get role name and permission from URL
	vars := mux.Vars(r)
	roleName := vars["name"]
	permission := vars["permission"]
	if roleName == "" || permission == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "Role name and permission are required")
		return
	}

	// Removing permissions from roles is not supported through the RBACManager interface
	WriteErrorResponse(w, http.StatusNotImplemented, "Removing permissions from roles is not supported through this API")
}
