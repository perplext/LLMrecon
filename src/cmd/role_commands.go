//go:build ignore

// Package cmd provides command-line interface functionality
package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/perplext/LLMrecon/src/security/access"
)

var (
	roleDescription string
	rolePermissions []string
)

// initRoleCommands initializes role management commands
func initRoleCommands() {
	// Role command
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Manage roles and permissions",
		Long:  `Create, update, delete, and list roles and permissions.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	accessControlCmd.AddCommand(roleCmd)

	// List roles command
	listRolesCmd := &cobra.Command{
		Use:   "list",
		Short: "List roles",
		Long:  `List all roles in the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserManageRoles); err != nil {
				return err
			}

			// Get roles from RBAC manager
			ctx := context.Background()
			roles := accessControlManager.rbacManager.GetAllRoles(ctx)

			// Sort roles by name for consistent output
			roleNames := make([]string, 0, len(roles))
			for role := range roles {
				roleNames = append(roleNames, string(role))
			}
			sort.Strings(roleNames)

			fmt.Println("Roles:")
			fmt.Println("------")
			for _, roleName := range roleNames {
				role := access.Role(roleName)
				description := roles[role]
				fmt.Printf("Role: %s\n", role)
				fmt.Printf("Description: %s\n", description)

				// Get permissions for this role
				permissions := accessControlManager.rbacManager.GetRolePermissions(ctx, role)
				if len(permissions) > 0 {
					permStrings := make([]string, len(permissions))
					for i, perm := range permissions {
						permStrings[i] = string(perm)
					}
					fmt.Printf("Permissions: %s\n", strings.Join(permStrings, ", "))
				} else {
					fmt.Printf("Permissions: none\n")
				}
				fmt.Println()
			}

			return nil
		},
	}
	roleCmd.AddCommand(listRolesCmd)

	// List permissions command
	listPermissionsCmd := &cobra.Command{
		Use:   "list-permissions",
		Short: "List permissions",
		Long:  `List all permissions in the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserManageRoles); err != nil {
				return err
			}

			// Get permissions from RBAC manager
			ctx := context.Background()
			permissions := accessControlManager.rbacManager.GetAllPermissions(ctx)

			// Sort permissions by name for consistent output
			permNames := make([]string, 0, len(permissions))
			for perm := range permissions {
				permNames = append(permNames, string(perm))
			}
			sort.Strings(permNames)

			fmt.Println("Permissions:")
			fmt.Println("------------")
			for _, permName := range permNames {
				perm := access.Permission(permName)
				description := permissions[perm]
				fmt.Printf("Permission: %s\n", perm)
				fmt.Printf("Description: %s\n", description)
				fmt.Println()
			}

			return nil
		},
	}
	roleCmd.AddCommand(listPermissionsCmd)

	// Get role command
	getRoleCmd := &cobra.Command{
		Use:   "get [role-name]",
		Short: "Get role details",
		Long:  `Get detailed information about a role.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserManageRoles); err != nil {
				return err
			}

			roleName := args[0]
			role := access.Role(roleName)

			// Get role details
			ctx := context.Background()
			roles := accessControlManager.rbacManager.GetAllRoles(ctx)
			description, exists := roles[role]
			if !exists {
				return fmt.Errorf("role not found: %s", roleName)
			}

			// Get permissions for this role
			permissions := accessControlManager.rbacManager.GetRolePermissions(ctx, role)

			fmt.Printf("Role Details:\n")
			fmt.Printf("Name: %s\n", role)
			fmt.Printf("Description: %s\n", description)

			if len(permissions) > 0 {
				fmt.Printf("Permissions:\n")
				for _, perm := range permissions {
					fmt.Printf("- %s\n", perm)
				}
			} else {
				fmt.Printf("Permissions: none\n")
			}

			return nil
		},
	}
	roleCmd.AddCommand(getRoleCmd)

	// Create role command
	createRoleCmd := &cobra.Command{
		Use:   "create [role-name]",
		Short: "Create a new role",
		Long:  `Create a new role with the specified name, description, and permissions.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserManageRoles); err != nil {
				return err
			}

			roleName := args[0]
			role := access.Role(roleName)

			// Validate input
			if roleDescription == "" {
				return fmt.Errorf("description is required")
			}

			// Check if role already exists
			ctx := context.Background()
			roles := accessControlManager.rbacManager.GetAllRoles(ctx)
			if _, exists := roles[role]; exists {
				return fmt.Errorf("role already exists: %s", roleName)
			}

			// Create role
			accessControlManager.rbacManager.AddRole(role, roleDescription)

			// Add permissions to role if specified
			if len(rolePermissions) > 0 {
				for _, permName := range rolePermissions {
					perm := access.Permission(permName)
					accessControlManager.rbacManager.AddPermissionToRole(role, perm)
				}
			}

			// Get current user
			currentUser, err := getCurrentUser(cmd)
			if err != nil {
				return err
			}

			// Log role creation
			accessControlManager.LogAudit(ctx, &access.AuditLog{
				UserID:      currentUser.ID,
				Username:    currentUser.Username,
				Action:      access.AuditActionCreate,
				Resource:    "role",
				ResourceID:  string(role),
				Description: "Role created",
				Severity:    access.AuditSeverityInfo,
				Status:      "success",
				Metadata: map[string]interface{}{
					"role_name":        roleName,
					"role_description": roleDescription,
					"permissions":      rolePermissions,
				},
			})

			fmt.Printf("Role created successfully:\n")
			fmt.Printf("Name: %s\n", role)
			fmt.Printf("Description: %s\n", roleDescription)

			if len(rolePermissions) > 0 {
				fmt.Printf("Permissions: %s\n", strings.Join(rolePermissions, ", "))
			} else {
				fmt.Printf("Permissions: none\n")
			}

			return nil
		},
	}
	createRoleCmd.Flags().StringVar(&roleDescription, "description", "", "Description of the role")
	createRoleCmd.Flags().StringSliceVar(&rolePermissions, "permissions", nil, "Permissions to assign to the role (comma-separated)")
	roleCmd.AddCommand(createRoleCmd)

	// Delete role command
	deleteRoleCmd := &cobra.Command{
		Use:   "delete [role-name]",
		Short: "Delete a role",
		Long:  `Delete a role from the system.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserManageRoles); err != nil {
				return err
			}

			roleName := args[0]
			role := access.Role(roleName)

			// Check if role exists
			ctx := context.Background()
			roles := accessControlManager.rbacManager.GetAllRoles(ctx)
			if _, exists := roles[role]; !exists {
				return fmt.Errorf("role not found: %s", roleName)
			}

			// Prevent deletion of built-in roles
			builtInRoles := []access.Role{
				access.RoleAdmin,
				access.RoleManager,
				access.RoleOperator,
				access.RoleAuditor,
				access.RoleUser,
				access.RoleGuest,
				access.RoleAutomation,
			}
			for _, builtInRole := range builtInRoles {
				if role == builtInRole {
					return fmt.Errorf("cannot delete built-in role: %s", roleName)
				}
			}

			// Delete role
			accessControlManager.rbacManager.RemoveRole(role)

			// Get current user
			currentUser, err := getCurrentUser(cmd)
			if err != nil {
				return err
			}

			// Log role deletion
			accessControlManager.LogAudit(ctx, &access.AuditLog{
				UserID:      currentUser.ID,
				Username:    currentUser.Username,
				Action:      access.AuditActionDelete,
				Resource:    "role",
				ResourceID:  string(role),
				Description: "Role deleted",
				Severity:    access.AuditSeverityInfo,
				Status:      "success",
				Metadata: map[string]interface{}{
					"role_name": roleName,
				},
			})

			fmt.Printf("Role deleted successfully: %s\n", roleName)
			return nil
		},
	}
	roleCmd.AddCommand(deleteRoleCmd)

	// Add permission to role command
	addPermissionCmd := &cobra.Command{
		Use:   "add-permission [role-name] [permission-name]",
		Short: "Add a permission to a role",
		Long:  `Add a permission to a role.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserManageRoles); err != nil {
				return err
			}

			roleName := args[0]
			permissionName := args[1]
			role := access.Role(roleName)
			permission := access.Permission(permissionName)

			// Check if role exists
			ctx := context.Background()
			roles := accessControlManager.rbacManager.GetAllRoles(ctx)
			if _, exists := roles[role]; !exists {
				return fmt.Errorf("role not found: %s", roleName)
			}

			// Check if permission exists
			permissions := accessControlManager.rbacManager.GetAllPermissions(ctx)
			if _, exists := permissions[permission]; !exists {
				return fmt.Errorf("permission not found: %s", permissionName)
			}

			// Add permission to role
			accessControlManager.rbacManager.AddPermissionToRole(role, permission)

			// Get current user
			currentUser, err := getCurrentUser(cmd)
			if err != nil {
				return err
			}

			// Log permission addition
			accessControlManager.LogAudit(ctx, &access.AuditLog{
				UserID:      currentUser.ID,
				Username:    currentUser.Username,
				Action:      access.AuditActionUpdate,
				Resource:    "role",
				ResourceID:  string(role),
				Description: "Permission added to role",
				Severity:    access.AuditSeverityInfo,
				Status:      "success",
				Metadata: map[string]interface{}{
					"role_name":       roleName,
					"permission_name": permissionName,
				},
			})
			fmt.Printf("Permission added to role successfully:\n")
			fmt.Printf("Role: %s\n", roleName)
			fmt.Printf("Permission: %s\n", permissionName)

			return nil
		},
	}
	roleCmd.AddCommand(addPermissionCmd)

	// Remove permission from role command
	removePermissionCmd := &cobra.Command{
		Use:   "remove-permission [role-name] [permission-name]",
		Short: "Remove a permission from a role",
		Long:  `Remove a permission from a role.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserManageRoles); err != nil {
				return err
			}

			roleName := args[0]
			permissionName := args[1]
			role := access.Role(roleName)
			permission := access.Permission(permissionName)

			// Check if role exists
			ctx := context.Background()
			roles := accessControlManager.rbacManager.GetAllRoles(ctx)
			if _, exists := roles[role]; !exists {
				return fmt.Errorf("role not found: %s", roleName)
			}

			// Check if permission exists
			permissions := accessControlManager.rbacManager.GetAllPermissions(ctx)
			if _, exists := permissions[permission]; !exists {
				return fmt.Errorf("permission not found: %s", permissionName)
			}

			// Remove permission from role
			accessControlManager.rbacManager.RemovePermissionFromRole(role, permission)

			// Get current user
			currentUser, err := getCurrentUser(cmd)
			if err != nil {
				return err
			}

			// Log permission removal
			accessControlManager.LogAudit(ctx, &access.AuditLog{
				UserID:      currentUser.ID,
				Username:    currentUser.Username,
				Action:      access.AuditActionUpdate,
				Resource:    "role",
				ResourceID:  string(role),
				Description: "Permission removed from role",
				Severity:    access.AuditSeverityInfo,
				Status:      "success",
				Metadata: map[string]interface{}{
					"role_name":       roleName,
					"permission_name": permissionName,
				},
			})

			fmt.Printf("Permission removed from role successfully:\n")
			fmt.Printf("Role: %s\n", roleName)
			fmt.Printf("Permission: %s\n", permissionName)
			return nil
		},
	}
	roleCmd.AddCommand(removePermissionCmd)

	// Check user permissions command
	checkPermissionCmd := &cobra.Command{
		Use:   "check-permission [username] [permission-name]",
		Short: "Check if a user has a permission",
		Long:  `Check if a user has a specific permission.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserRead); err != nil {
				return err
			}

			username := args[0]
			permissionName := args[1]
			permission := access.Permission(permissionName)

			// Get user
			ctx := context.Background()
			user, err := accessControlManager.GetUserByUsername(ctx, username)
			if err != nil {
				return fmt.Errorf("error getting user: %v", err)
			}

			// Check permission
			hasPermission := accessControlManager.HasPermission(ctx, user, permission)

			if hasPermission {
				fmt.Printf("User '%s' has permission '%s'\n", username, permissionName)
			} else {
				fmt.Printf("User '%s' does not have permission '%s'\n", username, permissionName)
			}

			return nil
		},
	}
	roleCmd.AddCommand(checkPermissionCmd)

	// Check user role command
	checkRoleCmd := &cobra.Command{
		Use:   "check-role [username] [role-name]",
		Short: "Check if a user has a role",
		Long:  `Check if a user has a specific role.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserRead); err != nil {
				return err
			}

			username := args[0]
			roleName := args[1]
			role := access.Role(roleName)

			// Get user
			ctx := context.Background()
			user, err := accessControlManager.GetUserByUsername(ctx, username)
			if err != nil {
				return fmt.Errorf("error getting user: %v", err)
			}

			// Check role
			hasRole := accessControlManager.HasRole(ctx, user, role)

			if hasRole {
				fmt.Printf("User '%s' has role '%s'\n", username, roleName)
			} else {
				fmt.Printf("User '%s' does not have role '%s'\n", username, roleName)
			}

			return nil
		},
	}
	roleCmd.AddCommand(checkRoleCmd)
