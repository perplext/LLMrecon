package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/perplext/LLMrecon/src/auth"
	"github.com/perplext/LLMrecon/src/repository"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create config directory
	configDir := filepath.Join(".", "config")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Printf("Failed to create config directory: %v\n", err)
		return
	}

	// Initialize authentication manager
	authManager, err := auth.NewManager(configDir, "example-passphrase")
	if err != nil {
		fmt.Printf("Failed to create authentication manager: %v\n", err)
		return
	}

	// Example 1: Managing Users
	fmt.Println("=== Example 1: Managing Users ===")
	
	// Create a new user
	user := &auth.User{
		ID:          "user1",
		Username:    "testuser",
		Role:        auth.UserRole,
		Permissions: []auth.Permission{auth.ReadPermission, auth.WritePermission},
	}
	
	// Save user
	if err := authManager.SaveUser(user); err != nil {
		fmt.Printf("Failed to save user: %v\n", err)
	} else {
		fmt.Printf("Created user: %s with role %s\n", user.Username, user.Role)
	}
	
	// List all users
	users, err := authManager.ListUsers()
	if err != nil {
		fmt.Printf("Failed to list users: %v\n", err)
	} else {
		fmt.Println("Users:")
		for _, u := range users {
			fmt.Printf("- %s (Role: %s, Permissions: %v)\n", u.Username, u.Role, u.Permissions)
		}
	}
	
	// Check permissions
	fmt.Printf("User '%s' has read permission: %v\n", 
		user.Username, authManager.HasPermission(user, auth.ReadPermission))
	fmt.Printf("User '%s' has admin permission: %v\n", 
		user.Username, authManager.HasPermission(user, auth.AdminPermission))
	
	fmt.Println()

	// Example 2: Managing Credentials
	fmt.Println("=== Example 2: Managing Credentials ===")
	
	// Create GitHub token credentials
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		// Use a placeholder token for the example
		githubToken = "example-github-token"
	}
	
	githubCreds := &auth.Credentials{
		ID:       auth.GenerateCredentialID(auth.GitHubProvider, "github-example"),
		Name:     "GitHub Example",
		Type:     auth.TokenAuth,
		Provider: auth.GitHubProvider,
		Token:    githubToken,
	}
	
	// Save credentials
	if err := authManager.SaveCredentials(githubCreds); err != nil {
		fmt.Printf("Failed to save GitHub credentials: %v\n", err)
	} else {
		fmt.Printf("Created credentials: %s (ID: %s)\n", githubCreds.Name, githubCreds.ID)
	}
	
	// Create GitLab credentials with username/password
	gitlabCreds := &auth.Credentials{
		ID:       auth.GenerateCredentialID(auth.GitLabProvider, "gitlab-example"),
		Name:     "GitLab Example",
		Type:     auth.BasicAuth,
		Provider: auth.GitLabProvider,
		Username: "example-user",
		Password: "example-password",
	}
	
	// Save credentials
	if err := authManager.SaveCredentials(gitlabCreds); err != nil {
		fmt.Printf("Failed to save GitLab credentials: %v\n", err)
	} else {
		fmt.Printf("Created credentials: %s (ID: %s)\n", gitlabCreds.Name, gitlabCreds.ID)
	}
	
	// List all credentials
	creds, err := authManager.ListCredentials()
	if err != nil {
		fmt.Printf("Failed to list credentials: %v\n", err)
	} else {
		fmt.Println("Credentials:")
		for _, c := range creds {
			fmt.Printf("- %s (Type: %s, Provider: %s)\n", c.Name, c.Type, c.Provider)
		}
	}
	
	fmt.Println()

	// Example 3: Authenticating Repositories
	fmt.Println("=== Example 3: Authenticating Repositories ===")
	
	// Create repository authenticator
	repoAuth := auth.NewRepositoryAuthenticator(authManager)
	
	// Create a GitHub repository
	githubConfig := repository.NewConfig(repository.GitHub, "github-repo", "https://github.com/perplext/LLMrecon")
	githubRepo, err := repository.Create(githubConfig)
	if err != nil {
		fmt.Printf("Failed to create GitHub repository: %v\n", err)
	} else {
		fmt.Printf("Created repository: %s (%s)\n", githubRepo.GetName(), githubRepo.GetType())
		
		// Authenticate repository with credentials
		if err := repoAuth.AuthenticateRepository(ctx, githubRepo, githubCreds.ID); err != nil {
			fmt.Printf("Failed to authenticate GitHub repository: %v\n", err)
		} else {
			fmt.Printf("Authenticated repository: %s with credentials: %s\n", 
				githubRepo.GetName(), githubCreds.Name)
		}
		
		// Authorize repository operation
		if err := repoAuth.AuthorizeRepositoryOperation(user, githubRepo, auth.ReadPermission); err != nil {
			fmt.Printf("Failed to authorize read operation: %v\n", err)
		} else {
			fmt.Printf("Authorized read operation for user: %s on repository: %s\n", 
				user.Username, githubRepo.GetName())
		}
		
		// Try to authorize an operation the user doesn't have permission for
		if err := repoAuth.AuthorizeRepositoryOperation(user, githubRepo, auth.AdminPermission); err != nil {
			fmt.Printf("As expected, failed to authorize admin operation: %v\n", err)
		} else {
			fmt.Printf("Unexpectedly authorized admin operation for user: %s on repository: %s\n", 
				user.Username, githubRepo.GetName())
		}
	}
	
	fmt.Println()

	// Example 4: OAuth Authentication (simulated)
	fmt.Println("=== Example 4: OAuth Authentication (Simulated) ===")
	
	// Create OAuth credentials
	oauthCreds := &auth.Credentials{
		ID:           auth.GenerateCredentialID(auth.GitHubProvider, "oauth-example"),
		Name:         "GitHub OAuth Example",
		Type:         auth.OAuthAuth,
		Provider:     auth.GitHubProvider,
		Token:        "example-oauth-token",
		RefreshToken: "example-refresh-token",
		TokenExpiry:  time.Now().Add(-1 * time.Hour), // Expired token
		ClientID:     "example-client-id",
		ClientSecret: "example-client-secret",
	}
	
	// Save credentials
	if err := authManager.SaveCredentials(oauthCreds); err != nil {
		fmt.Printf("Failed to save OAuth credentials: %v\n", err)
	} else {
		fmt.Printf("Created OAuth credentials: %s (ID: %s)\n", oauthCreds.Name, oauthCreds.ID)
		fmt.Printf("Token expiry: %s (expired: %v)\n", 
			oauthCreds.TokenExpiry.Format(time.RFC3339), oauthCreds.TokenExpiry.Before(time.Now()))
		
		// Try to refresh token (will fail in this example as we're using fake tokens)
		err := authManager.RefreshToken(ctx, oauthCreds)
		if err != nil {
			fmt.Printf("Token refresh failed (expected in this example): %v\n", err)
		} else {
			fmt.Printf("Token refreshed successfully. New expiry: %s\n", 
				oauthCreds.TokenExpiry.Format(time.RFC3339))
		}
	}
	
	fmt.Println()
	fmt.Println("Authentication and Authorization System Example Complete")
}
