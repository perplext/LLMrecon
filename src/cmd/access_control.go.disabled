// Package cmd provides command-line interface functionality
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/perplext/LLMrecon/src/security/access"
	"github.com/perplext/LLMrecon/src/security/access/types"
)

var (
	accessControlSystem *access.AccessControlSystem
	currentUser         *access.User
	currentSession      *access.Session
)

// accessControlCmd represents the access-control command
var accessControlCmd = &cobra.Command{
	Use:   "access-control",
	Short: "Manage access control and security auditing",
	Long: `Manage access control and security auditing for the LLMreconing Tool.
This includes user management, role-based access control, audit logging,
security incident management, and vulnerability tracking.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(accessControlCmd)

	// Initialize access control system
	config := access.DefaultAccessControlConfig()
	var err error
	accessControlSystem, err = access.NewAccessControlSystem(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing access control: %v\n", err)
		os.Exit(1)
	}

	// Add subcommands
	initUserCommands()
	initRoleCommands()
	initAuditCommands()
	initSecurityCommands()
	initAuthCommands()
}

// getCurrentUser gets the current authenticated user
func getCurrentUser(cmd *cobra.Command) (*access.User, error) {
	if currentUser == nil {
		return nil, fmt.Errorf("not authenticated, please login first")
	}
	
	// Validate the session if it exists
	if currentSession != nil {
		// Get the auth manager and validate the session
		authManager := accessControlSystem.Auth()
		user, err := authManager.ValidateSession(
			context.Background(),
			currentSession.ID,
			"127.0.0.1",
			"cli",
		)
		if err != nil {
			currentUser = nil
			currentSession = nil
			return nil, fmt.Errorf("session expired: %w", err)
		}
		currentUser = user
	}
	
	return currentUser, nil
}

// requirePermission checks if the current user has the required permission
func requirePermission(cmd *cobra.Command, permission access.Permission) error {
	user, err := getCurrentUser(cmd)
	if err != nil {
		return err
	}

	// Get the RBAC manager and check permissions
	rbacManager := accessControlSystem.RBAC()
	if !rbacManager.HasPermission(context.Background(), user, permission) {
		// Log unauthorized access attempt
		accessControlSystem.Audit().LogAudit(context.Background(), &types.AuditLog{
			Timestamp:   time.Now(),
			UserID:      user.ID,
			Username:    user.Username,
			Action:      types.AuditActionUnauthorized,
			Resource:    "permission",
			ResourceID:  string(permission),
			Description: fmt.Sprintf("Unauthorized access attempt to %s", permission),
			IPAddress:   "127.0.0.1",
			UserAgent:   "cli",
			Severity:    types.AuditSeverityMedium,
		})
		return fmt.Errorf("insufficient permissions: %s required", permission)
	}

	return nil
}

// requireRole checks if the current user has the required role
func requireRole(cmd *cobra.Command, role access.Role) error {
	user, err := getCurrentUser(cmd)
	if err != nil {
		return err
	}

	// Get the RBAC manager and check roles
	rbacManager := accessControlSystem.RBAC()
	if !rbacManager.HasRole(context.Background(), user, role) {
		// Log unauthorized access attempt
		accessControlSystem.Audit().LogAudit(context.Background(), &types.AuditLog{
			Timestamp:   time.Now(),
			UserID:      user.ID,
			Username:    user.Username,
			Action:      types.AuditActionUnauthorized,
			Resource:    "role",
			ResourceID:  role.ID,
			Description: fmt.Sprintf("Unauthorized access attempt requiring role %s", role),
			IPAddress:   "127.0.0.1",
			UserAgent:   "cli",
			Severity:    types.AuditSeverityMedium,
		})
		return fmt.Errorf("insufficient role: %s required", role)
	}

	return nil
}


