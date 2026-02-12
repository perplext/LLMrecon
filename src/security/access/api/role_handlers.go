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
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionRoleList) {
		WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Return the built-in role names since the RBACManager interface
	// does not expose role listing operations
	builtInRoles := []RoleResponse{
		{Name: access.RoleAdmin, Description: "Administrator", IsBuiltIn: true},
		{Name: access.RoleManager, Description: "Manager", IsBuiltIn: true},
		{Name: access.RoleOperator, Description: "Operator", IsBuiltIn: true},
		{Name: access.RoleAuditor, Description: "Auditor", IsBuiltIn: true},
		{Name: access.RoleUser, Description: "User", IsBuiltIn: true},
		{Name: access.RoleGuest, Description: "Guest", IsBuiltIn: true},
		{Name: access.RoleAutomation, Description: "Automation", IsBuiltIn: true},
	}

	WriteSuccessResponse(w, http.StatusOK, "Roles retrieved successfully", builtInRoles)
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
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionRoleCreate) {
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

	// Role CRUD operations are not available through the current RBACManager interface
	WriteErrorResponse(w, http.StatusNotImplemented, "Role creation is not supported through the current interface")
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
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionRoleView) {
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

	// Check if it's a known built-in role
	builtInRoles := map[string]string{
		access.RoleAdmin:      "Administrator with full system access",
		access.RoleManager:    "Manager with team oversight capabilities",
		access.RoleOperator:   "Operator with operational access",
		access.RoleAuditor:    "Auditor with read-only audit access",
		access.RoleUser:       "Standard user with basic access",
		access.RoleGuest:      "Guest with minimal access",
		access.RoleAutomation: "Automation service account",
	}

	if desc, ok := builtInRoles[roleName]; ok {
		resp := RoleResponse{
			Name:        roleName,
			Description: desc,
			IsBuiltIn:   true,
		}
		WriteSuccessResponse(w, http.StatusOK, "Role retrieved successfully", resp)
		return
	}

	WriteErrorResponse(w, http.StatusNotFound, "Role not found")
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
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionRoleUpdate) {
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

	// Parse request to validate JSON format
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Role update operations are not available through the current RBACManager interface
	_ = roleName
	WriteErrorResponse(w, http.StatusNotImplemented, "Role update is not supported through the current interface")
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
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionRoleDelete) {
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

	// Cannot delete built-in roles
	builtInRoles := []string{
		access.RoleAdmin, access.RoleManager, access.RoleOperator,
		access.RoleAuditor, access.RoleUser, access.RoleGuest, access.RoleAutomation,
	}
	for _, builtInRole := range builtInRoles {
		if strings.EqualFold(roleName, builtInRole) {
			WriteErrorResponse(w, http.StatusForbidden, "Cannot delete built-in roles")
			return
		}
	}

	// Role delete operations are not available through the current RBACManager interface
	WriteErrorResponse(w, http.StatusNotImplemented, "Role deletion is not supported through the current interface")
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
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionRoleUpdate) {
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

	// Role permission operations are not available through the current RBACManager interface
	_ = roleName
	WriteErrorResponse(w, http.StatusNotImplemented, "Adding permissions to roles is not supported through the current interface")
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
	if !s.accessManager.HasPermission(r.Context(), currentUser, access.PermissionRoleUpdate) {
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

	// Role permission operations are not available through the current RBACManager interface
	WriteErrorResponse(w, http.StatusNotImplemented, "Removing permissions from roles is not supported through the current interface")
}

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
