// Package access provides access control and security auditing functionality
package access

import "time"

// Role represents a role in the system with associated permissions
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	ParentRoles []string  `json:"parent_roles,omitempty"`
	IsBuiltIn   bool      `json:"is_built_in"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RBACConfig represents the configuration for role-based access control
type RBACConfig struct {
	DefaultRoles    []string            `json:"default_roles"`
	RolePermissions map[string][]string `json:"role_permissions"`
	RoleHierarchy   map[string][]string `json:"role_hierarchy"`
	CustomRoles     []Role              `json:"custom_roles"`
}

// NewRBACConfig creates a new RBAC configuration with default values
func NewRBACConfig() *RBACConfig {
	return &RBACConfig{
		DefaultRoles: []string{
			"admin",
			"user",
			"manager",
			"operator",
			"auditor",
			"guest",
		},
		RolePermissions: map[string][]string{
			"admin": {"*"},
			"user":  {"user.view"},
		},
		RoleHierarchy: map[string][]string{
			"admin":    {},
			"manager":  {"user"},
			"operator": {"user"},
			"auditor":  {"user"},
			"user":     {},
			"guest":    {},
		},
		CustomRoles: []Role{},
	}
}
