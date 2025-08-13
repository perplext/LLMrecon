// Package rbac provides enhanced role-based access control functionality
package rbac

import (
	"time"
)

// Role represents a role in the RBAC system
type Role struct {
	// Unique identifier for the role
	ID string `json:"id"`
	
	// Human-readable name for the role
	Name string `json:"name"`
	
	// Description of the role
	Description string `json:"description"`
	
	// Permissions assigned to this role
	Permissions []string `json:"permissions"`
	
	// Parent roles in the role hierarchy
	ParentRoles []string `json:"parent_roles"`
	
	// Whether the role is a system role (cannot be deleted)
	SystemRole bool `json:"system_role"`
	
	// Time when the role was created
	CreatedAt time.Time `json:"created_at"`
	
	// Time when the role was last updated
	UpdatedAt time.Time `json:"updated_at"`
	
	// Additional metadata for the role
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Permission represents a permission in the RBAC system
type Permission struct {
	// Unique identifier for the permission
	ID string `json:"id"`
	
	// Human-readable name for the permission
	Name string `json:"name"`
	
	// Description of the permission
	Description string `json:"description"`
	
	// Resource type this permission applies to
	ResourceType string `json:"resource_type"`
	
	// Action this permission allows
	Action string `json:"action"`
	
	// Whether this is a system permission (cannot be deleted)
	SystemPermission bool `json:"system_permission"`
	
	// Time when the permission was created
	CreatedAt time.Time `json:"created_at"`
	
	// Time when the permission was last updated
	UpdatedAt time.Time `json:"updated_at"`
	
	// Additional metadata for the permission
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UserRole represents a role assignment to a user
type UserRole struct {
	// User ID
	UserID string `json:"user_id"`
	
	// Role ID
	RoleID string `json:"role_id"`
	
	// Time when the role was assigned
	AssignedAt time.Time `json:"assigned_at"`
	
	// User who assigned the role
	AssignedBy string `json:"assigned_by,omitempty"`
	
	// Expiration time for the role assignment (if any)
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// UserPermission represents a direct permission assignment to a user
type UserPermission struct {
	// User ID
	UserID string `json:"user_id"`
	
	// Permission ID
	PermissionID string `json:"permission_id"`
	
	// Time when the permission was assigned
	AssignedAt time.Time `json:"assigned_at"`
	
	// User who assigned the permission
	AssignedBy string `json:"assigned_by,omitempty"`
	
	// Expiration time for the permission assignment (if any)
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// NewRole creates a new role
func NewRole(id, name, description string) *Role {
	now := time.Now()
	return &Role{
		ID:          id,
		Name:        name,
		Description: description,
		Permissions: make([]string, 0),
		ParentRoles: make([]string, 0),
		SystemRole:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]interface{}),
	}
}

// NewPermission creates a new permission
func NewPermission(id, name, description, resourceType, action string) *Permission {
	now := time.Now()
	return &Permission{
		ID:               id,
		Name:             name,
		Description:      description,
		ResourceType:     resourceType,
		Action:           action,
		SystemPermission: false,
		CreatedAt:        now,
		UpdatedAt:        now,
		Metadata:         make(map[string]interface{}),
	}
}

// DefaultRoles returns the default system roles
func DefaultRoles() []Role {
	now := time.Now()
	
	// Admin role
	adminRole := Role{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full administrative access to all system functions",
		Permissions: []string{
			"system.admin",
			"system.config",
			"user.create",
			"user.read",
			"user.update",
			"user.delete",
			"role.create",
			"role.read",
			"role.update",
			"role.delete",
			"permission.create",
			"permission.read",
			"permission.update",
			"permission.delete",
		},
		ParentRoles: []string{},
		SystemRole:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// Manager role
	managerRole := Role{
		ID:          "manager",
		Name:        "Manager",
		Description: "Management access to users and content",
		Permissions: []string{
			"user.create",
			"user.read",
			"user.update",
			"role.read",
			"permission.read",
			"content.create",
			"content.read",
			"content.update",
			"content.delete",
			"report.create",
			"report.read",
			"report.export",
		},
		ParentRoles: []string{},
		SystemRole:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// User role
	userRole := Role{
		ID:          "user",
		Name:        "User",
		Description: "Standard user access",
		Permissions: []string{
			"user.read.self",
			"user.update.self",
			"content.read",
			"content.create",
			"content.update.own",
			"content.delete.own",
			"report.read",
		},
		ParentRoles: []string{},
		SystemRole:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// Guest role
	guestRole := Role{
		ID:          "guest",
		Name:        "Guest",
		Description: "Limited guest access",
		Permissions: []string{
			"content.read.public",
		},
		ParentRoles: []string{},
		SystemRole:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// Security Auditor role
	securityAuditorRole := Role{
		ID:          "security_auditor",
		Name:        "Security Auditor",
		Description: "Access to security logs and audit information",
		Permissions: []string{
			"audit.read",
			"audit.export",
			"security.incident.read",
			"security.vulnerability.read",
			"user.read",
			"role.read",
			"permission.read",
		},
		ParentRoles: []string{},
		SystemRole:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// Security Administrator role
	securityAdminRole := Role{
		ID:          "security_admin",
		Name:        "Security Administrator",
		Description: "Management of security settings and incidents",
		Permissions: []string{
			"security.config",
			"security.incident.create",
			"security.incident.update",
			"security.incident.resolve",
			"security.vulnerability.create",
			"security.vulnerability.update",
			"security.vulnerability.resolve",
		},
		ParentRoles: []string{"security_auditor"},
		SystemRole:  true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	return []Role{
		adminRole,
		managerRole,
		userRole,
		guestRole,
		securityAuditorRole,
		securityAdminRole,
	}
}

// DefaultPermissions returns the default system permissions
func DefaultPermissions() []Permission {
	now := time.Now()
	
	return []Permission{
		{
			ID:               "system.admin",
			Name:             "System Administration",
			Description:      "Full administrative access to the system",
			ResourceType:     "system",
			Action:           "admin",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "system.config",
			Name:             "System Configuration",
			Description:      "Configure system settings",
			ResourceType:     "system",
			Action:           "config",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "user.create",
			Name:             "Create Users",
			Description:      "Create new users",
			ResourceType:     "user",
			Action:           "create",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "user.read",
			Name:             "Read Users",
			Description:      "View user information",
			ResourceType:     "user",
			Action:           "read",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "user.update",
			Name:             "Update Users",
			Description:      "Update user information",
			ResourceType:     "user",
			Action:           "update",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "user.delete",
			Name:             "Delete Users",
			Description:      "Delete users",
			ResourceType:     "user",
			Action:           "delete",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "user.read.self",
			Name:             "Read Own User",
			Description:      "View own user information",
			ResourceType:     "user",
			Action:           "read.self",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "user.update.self",
			Name:             "Update Own User",
			Description:      "Update own user information",
			ResourceType:     "user",
			Action:           "update.self",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "role.create",
			Name:             "Create Roles",
			Description:      "Create new roles",
			ResourceType:     "role",
			Action:           "create",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "role.read",
			Name:             "Read Roles",
			Description:      "View role information",
			ResourceType:     "role",
			Action:           "read",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "role.update",
			Name:             "Update Roles",
			Description:      "Update role information",
			ResourceType:     "role",
			Action:           "update",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "role.delete",
			Name:             "Delete Roles",
			Description:      "Delete roles",
			ResourceType:     "role",
			Action:           "delete",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "permission.create",
			Name:             "Create Permissions",
			Description:      "Create new permissions",
			ResourceType:     "permission",
			Action:           "create",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "permission.read",
			Name:             "Read Permissions",
			Description:      "View permission information",
			ResourceType:     "permission",
			Action:           "read",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "permission.update",
			Name:             "Update Permissions",
			Description:      "Update permission information",
			ResourceType:     "permission",
			Action:           "update",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "permission.delete",
			Name:             "Delete Permissions",
			Description:      "Delete permissions",
			ResourceType:     "permission",
			Action:           "delete",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "audit.read",
			Name:             "Read Audit Logs",
			Description:      "View audit logs",
			ResourceType:     "audit",
			Action:           "read",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "audit.export",
			Name:             "Export Audit Logs",
			Description:      "Export audit logs",
			ResourceType:     "audit",
			Action:           "export",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.config",
			Name:             "Configure Security",
			Description:      "Configure security settings",
			ResourceType:     "security",
			Action:           "config",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.incident.create",
			Name:             "Create Security Incidents",
			Description:      "Create security incidents",
			ResourceType:     "security.incident",
			Action:           "create",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.incident.read",
			Name:             "Read Security Incidents",
			Description:      "View security incidents",
			ResourceType:     "security.incident",
			Action:           "read",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.incident.update",
			Name:             "Update Security Incidents",
			Description:      "Update security incidents",
			ResourceType:     "security.incident",
			Action:           "update",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.incident.resolve",
			Name:             "Resolve Security Incidents",
			Description:      "Resolve security incidents",
			ResourceType:     "security.incident",
			Action:           "resolve",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.vulnerability.create",
			Name:             "Create Security Vulnerabilities",
			Description:      "Create security vulnerabilities",
			ResourceType:     "security.vulnerability",
			Action:           "create",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.vulnerability.read",
			Name:             "Read Security Vulnerabilities",
			Description:      "View security vulnerabilities",
			ResourceType:     "security.vulnerability",
			Action:           "read",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.vulnerability.update",
			Name:             "Update Security Vulnerabilities",
			Description:      "Update security vulnerabilities",
			ResourceType:     "security.vulnerability",
			Action:           "update",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "security.vulnerability.resolve",
			Name:             "Resolve Security Vulnerabilities",
			Description:      "Resolve security vulnerabilities",
			ResourceType:     "security.vulnerability",
			Action:           "resolve",
			SystemPermission: true,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
}
