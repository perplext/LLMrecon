// Package cmd provides command-line interface functionality
package cmd

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/perplext/LLMrecon/src/security/access"
	"github.com/perplext/LLMrecon/src/security/access/common"
)

var (
	userUsername string
	userEmail    string
	userPassword string
	userRoles    []string
	userActive   bool
	userMFA      string
)

// initUserCommands initializes user management commands
func initUserCommands() {
	// User command
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  `Create, update, delete, and list users.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	accessControlCmd.AddCommand(userCmd)

	// List users command
	listUsersCmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		Long:  `List all users in the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserRead); err != nil {
				return err
			}

			ctx := context.Background()
			users, err := accessControlSystem.GetAllUsers(ctx)
			if err != nil {
				return fmt.Errorf("error listing users: %v", err)
			}

			fmt.Println("Users:")
			fmt.Println("------")
			for _, user := range users {
				roleStr := strings.Join(convertRolesToStrings(user.Roles), ", ")
				mfaStatus := "Disabled"
				if user.MFAEnabled {
					mfaStatus = "Enabled"
				}
				status := "Active"
				if !user.Active {
					status = "Inactive"
				}
				if user.Locked {
					status = "Locked"
				}
				fmt.Printf("ID: %s\n", user.ID)
				fmt.Printf("Username: %s\n", user.Username)
				fmt.Printf("Email: %s\n", user.Email)
				fmt.Printf("Roles: %s\n", roleStr)
				fmt.Printf("MFA: %s\n", mfaStatus)
				fmt.Printf("Status: %s\n", status)
				fmt.Printf("Created: %s\n", user.CreatedAt.Format(time.RFC3339))
				fmt.Printf("Last Login: %s\n", user.LastLogin.Format(time.RFC3339))
				fmt.Println()
			}

			return nil
		},
	}
	userCmd.AddCommand(listUsersCmd)

	// Create user command
	createUserCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Long:  `Create a new user with the specified username, email, password, and roles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserCreate); err != nil {
				return err
			}

			// Validate input
			if userUsername == "" {
				return fmt.Errorf("username is required")
			}
			if userEmail == "" {
				return fmt.Errorf("email is required")
			}
			if userPassword == "" {
				// Prompt for password if not provided
				fmt.Print("Enter password: ")
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("error reading password: %v", err)
				}
				fmt.Println()
				userPassword = string(passwordBytes)
			}

			// Convert role strings to Role type
			roles := convertStringToRoles(userRoles)

			// Create user
			ctx := context.Background()
			user, err := accessControlSystem.CreateUser(ctx, userUsername, userEmail, userPassword, roles)
			if err != nil {
				return fmt.Errorf("error creating user: %v", err)
			}

			fmt.Printf("User created successfully:\n")
			fmt.Printf("ID: %s\n", user.ID)
			fmt.Printf("Username: %s\n", user.Username)
			fmt.Printf("Email: %s\n", user.Email)
			fmt.Printf("Roles: %s\n", strings.Join(convertRolesToStrings(user.Roles), ", "))

			return nil
		},
	}
	createUserCmd.Flags().StringVar(&userUsername, "username", "", "Username for the new user")
	createUserCmd.Flags().StringVar(&userEmail, "email", "", "Email for the new user")
	createUserCmd.Flags().StringVar(&userPassword, "password", "", "Password for the new user")
	createUserCmd.Flags().StringSliceVar(&userRoles, "roles", []string{"user"}, "Roles for the new user (comma-separated)")
	userCmd.AddCommand(createUserCmd)

	// Update user command
	updateUserCmd := &cobra.Command{
		Use:   "update [user-id]",
		Short: "Update an existing user",
		Long:  `Update an existing user's information.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserUpdate); err != nil {
				return err
			}

			userID := args[0]

			// Convert role strings to Role type
			roles := convertStringToRoles(userRoles)

			// Update user
			ctx := context.Background()
			// Get the user first
			user, err := accessControlSystem.GetUserByID(ctx, userID)
			if err != nil {
				return fmt.Errorf("error getting user: %v", err)
			}

			// Update user fields
			user.Username = userUsername
			user.Email = userEmail
			user.Roles = convertStringToRoles(userRoles)
			user.Active = userActive

			// Save the updated user
			err = accessControlSystem.UpdateUser(ctx, user)
			if err != nil {
				return fmt.Errorf("error updating user: %v", err)
			}

			fmt.Printf("User updated successfully:\n")
			fmt.Printf("ID: %s\n", user.ID)
			fmt.Printf("Username: %s\n", user.Username)
			fmt.Printf("Email: %s\n", user.Email)
			fmt.Printf("Roles: %s\n", strings.Join(convertRolesToStrings(user.Roles), ", "))
			fmt.Printf("Active: %v\n", user.Active)

			return nil
		},
	}
	updateUserCmd.Flags().StringVar(&userUsername, "username", "", "New username (leave empty to keep current)")
	updateUserCmd.Flags().StringVar(&userEmail, "email", "", "New email (leave empty to keep current)")
	updateUserCmd.Flags().StringSliceVar(&userRoles, "roles", nil, "New roles (comma-separated, leave empty to keep current)")
	updateUserCmd.Flags().BoolVar(&userActive, "active", true, "Whether the user is active")
	userCmd.AddCommand(updateUserCmd)

	// Delete user command
	deleteUserCmd := &cobra.Command{
		Use:   "delete [user-id]",
		Short: "Delete a user",
		Long:  `Delete a user from the system.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserDelete); err != nil {
				return err
			}

			userID := args[0]

			// Delete user
			ctx := context.Background()
			if err := accessControlSystem.DeleteUser(ctx, userID); err != nil {
				return fmt.Errorf("error deleting user: %v", err)
			}

			fmt.Printf("User deleted successfully\n")

			return nil
		},
	}
	userCmd.AddCommand(deleteUserCmd)

	// Reset password command
	resetPasswordCmd := &cobra.Command{
		Use:   "reset-password [user-id]",
		Short: "Reset a user's password",
		Long:  `Reset a user's password to a new value.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserUpdate); err != nil {
				return err
			}

			userID := args[0]

			// Prompt for new password if not provided
			if userPassword == "" {
				fmt.Print("Enter new password: ")
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("error reading password: %v", err)
				}
				fmt.Println()
				userPassword = string(passwordBytes)
			}

			// Reset password
			ctx := context.Background()
			if err := accessControlSystem.UpdateUserPassword(ctx, userID, "", userPassword); err != nil {
				return fmt.Errorf("error resetting password: %v", err)
			}

			fmt.Printf("Password reset successfully\n")

			return nil
		},
	}
	resetPasswordCmd.Flags().StringVar(&userPassword, "password", "", "New password")
	userCmd.AddCommand(resetPasswordCmd)

	// Lock user command
	lockUserCmd := &cobra.Command{
		Use:   "lock [user-id]",
		Short: "Lock a user account",
		Long:  `Lock a user account to prevent login.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserUpdate); err != nil {
				return err
			}

			userID := args[0]

			// Lock user
			ctx := context.Background()
			// Get the user first
			user, err := accessControlSystem.GetUserByID(ctx, userID)
			if err != nil {
				return fmt.Errorf("error getting user: %v", err)
			}

			// Lock the user
			user.Locked = true
			if err := accessControlSystem.UpdateUser(ctx, user); err != nil {
				return fmt.Errorf("error locking user: %v", err)
			}

			fmt.Printf("User account locked successfully\n")

			return nil
		},
	}
	userCmd.AddCommand(lockUserCmd)

	// Unlock user command
	unlockUserCmd := &cobra.Command{
		Use:   "unlock [user-id]",
		Short: "Unlock a user account",
		Long:  `Unlock a user account to allow login.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserUpdate); err != nil {
				return err
			}

			userID := args[0]

			// Unlock user
			ctx := context.Background()
			// Get the user first
			user, err := accessControlSystem.GetUserByID(ctx, userID)
			if err != nil {
				return fmt.Errorf("error getting user: %v", err)
			}

			// Unlock the user
			user.Locked = false
			if err := accessControlSystem.UpdateUser(ctx, user); err != nil {
				return fmt.Errorf("error unlocking user: %v", err)
			}

			fmt.Printf("User account unlocked successfully\n")

			return nil
		},
	}
	userCmd.AddCommand(unlockUserCmd)

	// MFA commands
	mfaCmd := &cobra.Command{
		Use:   "mfa",
		Short: "Manage multi-factor authentication",
		Long:  `Enable or disable multi-factor authentication for a user.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	userCmd.AddCommand(mfaCmd)

	// Enable MFA command
	enableMFACmd := &cobra.Command{
		Use:   "enable [user-id]",
		Short: "Enable MFA for a user",
		Long:  `Enable multi-factor authentication for a user.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserUpdate); err != nil {
				return err
			}

			userID := args[0]
			method := common.AuthMethod(userMFA)

			// Validate MFA method
			validMethods := []common.AuthMethod{
				common.AuthMethodTOTP,
				common.AuthMethodSMS,
				common.AuthMethodWebAuthn,
			}
			valid := false
			for _, m := range validMethods {
				if method == m {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid MFA method: %s", userMFA)
			}

			// Enable MFA
			ctx := context.Background()
			if err := accessControlSystem.EnableMFA(ctx, userID, method); err != nil {
				return fmt.Errorf("error enabling MFA: %v", err)
			}

			fmt.Printf("MFA enabled successfully\n")

			return nil
		},
	}
	enableMFACmd.Flags().StringVar(&userMFA, "method", string(common.AuthMethodTOTP), "MFA method (totp, sms, webauthn)")
	mfaCmd.AddCommand(enableMFACmd)

	// Disable MFA command
	disableMFACmd := &cobra.Command{
		Use:   "disable [user-id]",
		Short: "Disable MFA for a user",
		Long:  `Disable multi-factor authentication for a user.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePermission(cmd, access.PermissionUserUpdate); err != nil {
				return err
			}

			userID := args[0]
			method := common.AuthMethod(userMFA)

			// Disable MFA
			ctx := context.Background()
			if err := accessControlSystem.DisableMFA(ctx, userID, method); err != nil {
				return fmt.Errorf("error disabling MFA: %v", err)
			}

			fmt.Printf("MFA disabled successfully\n")

			return nil
		},
	}
	disableMFACmd.Flags().StringVar(&userMFA, "method", string(common.AuthMethodTOTP), "MFA method to disable (totp, sms, webauthn)")
	mfaCmd.AddCommand(disableMFACmd)
}

// Helper functions

// convertStringToRoles converts string role names to Role type
func convertStringToRoles(roleNames []string) []access.Role {
	roles := make([]access.Role, 0, len(roleNames))
	for _, name := range roleNames {
		switch name {
		case "admin":
			roles = append(roles, access.RoleAdmin)
		case "user":
			roles = append(roles, access.RoleUser)
		case "auditor":
			roles = append(roles, access.RoleAuditor)
		case "manager":
			roles = append(roles, access.RoleManager)
		case "readonly":
			roles = append(roles, access.RoleReadOnly)
		default:
			// Skip invalid roles
			fmt.Printf("Warning: ignoring invalid role '%s'\n", name)
		}
	}
	return roles
}

// convertRolesToStrings converts Role type to string names
func convertRolesToStrings(roles []access.Role) []string {
	names := make([]string, 0, len(roles))
	for _, role := range roles {
		names = append(names, string(role))
	}
	return names
}
