// Package cmd provides command-line interface functionality
package cmd

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/perplext/LLMrecon/src/security/access"
)

var (
	authUsername string
	authPassword string
	authMFACode  string
)

// initAuthCommands initializes authentication commands
func initAuthCommands() {
	// Auth command
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  `Login, logout, and manage authentication sessions.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	accessControlCmd.AddCommand(authCmd)

	// Login command
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to the system",
		Long:  `Authenticate with the system to gain access to protected resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate input
			if authUsername == "" {
				return fmt.Errorf("username is required")
			}

			// Get password if not provided
			if authPassword == "" {
				fmt.Print("Enter password: ")
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("error reading password: %v", err)
				}
				fmt.Println()
				authPassword = string(passwordBytes)
			}

			// Get IP address and user agent
			ipAddress := "127.0.0.1" // Local CLI usage
			userAgent := "LLMrecon CLI"

			// Login
			ctx := context.Background()
			session, err := accessControlSystem.Auth().Login(ctx, authUsername, authPassword, ipAddress, userAgent)
			if err != nil {
				if err == access.ErrMFARequired {
					// MFA required
					fmt.Println("Multi-factor authentication required")
					fmt.Print("Enter MFA code: ")
					fmt.Scanln(&authMFACode)

					// Verify MFA
					if err := accessControlSystem.Auth().VerifyMFA(ctx, session.ID, authMFACode); err != nil {
						return fmt.Errorf("MFA verification failed: %v", err)
					}
				} else {
					return fmt.Errorf("login failed: %v", err)
				}
			}

			// Get user
			user, err := accessControlSystem.GetUserByUsername(ctx, authUsername)
			if err != nil {
				return fmt.Errorf("error getting user: %v", err)
			}

			// Set current user and session
			currentUser = user
			currentSession = session

			// Save session to file for persistence
			if err := saveSession(session); err != nil {
				fmt.Printf("Warning: Failed to save session: %v\n", err)
			}

			fmt.Printf("Logged in as %s\n", user.Username)
			return nil
		},
	}
	loginCmd.Flags().StringVar(&authUsername, "username", "", "Username for login")
	loginCmd.Flags().StringVar(&authPassword, "password", "", "Password for login")
	authCmd.AddCommand(loginCmd)

	// Logout command
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from the system",
		Long:  `End the current authentication session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if logged in
			if currentSession == nil {
				return fmt.Errorf("not logged in")
			}

			// Logout
			ctx := context.Background()
			if err := accessControlSystem.Auth().Logout(ctx, currentSession.ID); err != nil {
				return fmt.Errorf("logout failed: %v", err)
			}

			// Clear current user and session
			currentUser = nil
			currentSession = nil

			// Remove session file
			if err := removeSessionFile(); err != nil {
				fmt.Printf("Warning: Failed to remove session file: %v\n", err)
			}

			fmt.Println("Logged out successfully")
			return nil
		},
	}
	authCmd.AddCommand(logoutCmd)

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  `Display information about the current authentication session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load session if not already loaded
			if currentSession == nil {
				loadSession()
			}

			// Check if logged in
			if currentSession == nil {
				fmt.Println("Not logged in")
				return nil
			}

			// Get user
			ctx := context.Background()
			user, err := accessControlSystem.GetUserByID(ctx, currentSession.UserID)
			if err != nil {
				return fmt.Errorf("error getting user: %v", err)
			}

			// Verify session
			valid, err := accessControlSystem.Auth().VerifySession(ctx, currentSession.Token)
			if err != nil || !valid {
				fmt.Println("Session expired or invalid")
				// Clear current user and session
				currentUser = nil
				currentSession = nil
				// Remove session file
				removeSessionFile()
				return nil
			}

			// Display status
			fmt.Println("Authentication Status:")
			fmt.Printf("Logged in as: %s (%s)\n", user.Username, user.Email)
			fmt.Printf("User ID: %s\n", user.ID)
			fmt.Printf("Roles: %v\n", user.Roles)
			fmt.Printf("Session ID: %s\n", currentSession.ID)
			fmt.Printf("Session expires: %s\n", currentSession.ExpiresAt.Format(time.RFC3339))
			fmt.Printf("Last activity: %s\n", currentSession.LastActivity.Format(time.RFC3339))
			fmt.Printf("MFA completed: %v\n", currentSession.MFACompleted)

			return nil
		},
	}
	authCmd.AddCommand(statusCmd)

	// Refresh command
	refreshCmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh the current session",
		Long:  `Refresh the current authentication session to extend its validity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load session if not already loaded
			if currentSession == nil {
				loadSession()
			}

			// Check if logged in
			if currentSession == nil {
				return fmt.Errorf("not logged in")
			}

			// Refresh session
			ctx := context.Background()
			newSession, err := accessControlSystem.Auth().RefreshSession(ctx, currentSession.RefreshToken)
			if err != nil {
				return fmt.Errorf("session refresh failed: %v", err)
			}

			// Update current session
			currentSession = newSession

			// Save session to file for persistence
			if err := saveSession(newSession); err != nil {
				fmt.Printf("Warning: Failed to save session: %v\n", err)
			}

			fmt.Println("Session refreshed successfully")
			fmt.Printf("New expiration: %s\n", newSession.ExpiresAt.Format(time.RFC3339))

			return nil
		},
	}
	authCmd.AddCommand(refreshCmd)
}

// Session file operations

// getSessionFilePath returns the path to the session file
func getSessionFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".LLMrecon-session"
	}
	return fmt.Sprintf("%s/.LLMrecon-session", homeDir)
}

// saveSession saves the session to a file
func saveSession(session *access.Session) error {
	filePath := getSessionFilePath()
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write session data
	fmt.Fprintf(file, "ID=%s\n", session.ID)
	fmt.Fprintf(file, "UserID=%s\n", session.UserID)
	fmt.Fprintf(file, "Token=%s\n", session.Token)
	fmt.Fprintf(file, "RefreshToken=%s\n", session.RefreshToken)
	fmt.Fprintf(file, "ExpiresAt=%s\n", session.ExpiresAt.Format(time.RFC3339))
	fmt.Fprintf(file, "LastActivity=%s\n", session.LastActivity.Format(time.RFC3339))
	fmt.Fprintf(file, "MFACompleted=%v\n", session.MFACompleted)

	return nil
}

// loadSession loads the session from a file
func loadSession() {
	filePath := getSessionFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		// No session file or can't read it
		return
	}

	// Parse session data
	lines := splitLines(string(data))
	session := &access.Session{}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		parts := splitKeyValue(line)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "ID":
			session.ID = value
		case "UserID":
			session.UserID = value
		case "Token":
			session.Token = value
		case "RefreshToken":
			session.RefreshToken = value
		case "ExpiresAt":
			t, err := time.Parse(time.RFC3339, value)
			if err == nil {
				session.ExpiresAt = t
			}
		case "LastActivity":
			t, err := time.Parse(time.RFC3339, value)
			if err == nil {
				session.LastActivity = t
			}
		case "MFACompleted":
			session.MFACompleted = value == "true"
		}
	}

	// Set current session
	if session.ID != "" && session.Token != "" {
		currentSession = session
	}
}

// removeSessionFile removes the session file
func removeSessionFile() error {
	filePath := getSessionFilePath()
	return os.Remove(filePath)
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	var line string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, line)
			line = ""
		} else {
			line += string(r)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// splitKeyValue splits a line into key and value
func splitKeyValue(line string) []string {
	var parts []string
	var key string
	var value string
	var inKey = true
	for _, r := range line {
		if r == '=' && inKey {
			inKey = false
		} else if inKey {
			key += string(r)
		} else {
			value += string(r)
		}
	}
	if key != "" {
		parts = append(parts, key)
		parts = append(parts, value)
	}
	return parts
}
