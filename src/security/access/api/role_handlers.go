// Package api provides a RESTful API for the access control system
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/perplext/LLMrecon/src/security/access"
)

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	ParentRoles []string `json:"parent_roles,omitempty"`

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Description string   `json:"description,omitempty"`
	ParentRoles []string `json:"parent_roles,omitempty"`

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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleList) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Get roles
	roles, err := rbacManager.ListRoles(r.Context())
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list roles")
		return
	}

	// Convert roles to response format
	var roleResponses []RoleResponse
	for _, role := range roles {
		roleResponses = append(roleResponses, convertRoleToResponse(role))
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Roles retrieved successfully", roleResponses)

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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleCreate) {
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

	// Check if the role name is reserved
	builtInRoles := []string{
		access.RoleAdmin,
		access.RoleManager,
		access.RoleOperator,
		access.RoleAuditor,
		access.RoleUser,
		access.RoleGuest,
		access.RoleAutomation,
	}
	for _, builtInRole := range builtInRoles {
		if strings.EqualFold(req.Name, builtInRole) {
			WriteErrorResponse(w, http.StatusBadRequest, "Cannot create a role with a reserved name")
			return
		}
	}

	// Check if the current user has permission to grant the requested permissions
	for _, permission := range req.Permissions {
		if !rbacManager.CanGrantPermission(r.Context(), currentUser, permission) {
			WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions to grant permission: "+permission)
			return
		}
	}

	// Create role
	role := &access.Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		ParentRoles: req.ParentRoles,
		IsBuiltIn:   false,
	}
	if err := rbacManager.CreateRole(r.Context(), role); err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "already exists"):
			WriteErrorResponse(w, http.StatusConflict, "Role already exists")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create role")
		}
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusCreated, "Role created successfully", convertRoleToResponse(role))

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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleView) {
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

	// Get role
	role, err := rbacManager.GetRole(r.Context(), roleName)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Role retrieved successfully", convertRoleToResponse(role))

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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleUpdate) {
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

	// Get role
	role, err := rbacManager.GetRole(r.Context(), roleName)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	// Cannot update built-in roles
	if role.IsBuiltIn {
		WriteErrorResponse(w, http.StatusForbidden, "Cannot update built-in roles")
		return
	}

	// Update role fields
	if req.Description != "" {
		role.Description = req.Description
	}

	if len(req.ParentRoles) > 0 {
		role.ParentRoles = req.ParentRoles
	}

	// Update role
	if err := rbacManager.UpdateRole(r.Context(), role); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Role updated successfully", convertRoleToResponse(role))

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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleDelete) {
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

	// Get role
	role, err := rbacManager.GetRole(r.Context(), roleName)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	// Cannot delete built-in roles
	if role.IsBuiltIn {
		WriteErrorResponse(w, http.StatusForbidden, "Cannot delete built-in roles")
		return
	}

	// Delete role
	if err := rbacManager.DeleteRole(r.Context(), roleName); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Role deleted successfully", nil)

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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleUpdate) {
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

	// Get role
	role, err := rbacManager.GetRole(r.Context(), roleName)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	// Cannot update built-in roles
	if role.IsBuiltIn {
		WriteErrorResponse(w, http.StatusForbidden, "Cannot update built-in roles")
		return
	}

	// Check if the current user has permission to grant the requested permission
	if !rbacManager.CanGrantPermission(r.Context(), currentUser, req.Permission) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions to grant permission: "+req.Permission)
		return
	}

	// Add permission to role
	if err := rbacManager.AddPermissionToRole(r.Context(), roleName, req.Permission); err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "already has permission"):
			WriteErrorResponse(w, http.StatusConflict, "Role already has this permission")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to add permission to role")
		}
		return
	}

	// Get updated role
	updatedRole, err := rbacManager.GetRole(r.Context(), roleName)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get updated role")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Permission added to role successfully", convertRoleToResponse(updatedRole))

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
	if !rbacManager.HasPermission(r.Context(), currentUser, access.PermissionRoleUpdate) {
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
	// Get role
	role, err := rbacManager.GetRole(r.Context(), roleName)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	// Cannot update built-in roles
	if role.IsBuiltIn {
		WriteErrorResponse(w, http.StatusForbidden, "Cannot update built-in roles")
		return
	}

	// Remove permission from role
	if err := rbacManager.RemovePermissionFromRole(r.Context(), roleName, permission); err != nil {
		// Handle specific error types
		switch {
		case strings.Contains(err.Error(), "does not have permission"):
			WriteErrorResponse(w, http.StatusBadRequest, "Role does not have this permission")
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, "Failed to remove permission from role")
		}
		return
	}

	// Get updated role
	updatedRole, err := rbacManager.GetRole(r.Context(), roleName)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get updated role")
		return
	}

	// Return success response
	WriteSuccessResponse(w, http.StatusOK, "Permission removed from role successfully", convertRoleToResponse(updatedRole))

// convertRoleToResponse converts a role to a response format
func convertRoleToResponse(role *access.Role) RoleResponse {
	return RoleResponse{
		Name:        role.Name,
		Description: role.Description,
		Permissions: role.Permissions,
		ParentRoles: role.ParentRoles,
		IsBuiltIn:   role.IsBuiltIn,
	}
}
}
}
}
}
}
