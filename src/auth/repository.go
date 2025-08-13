package auth

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/repository"
)

// RepositoryAuthenticator authenticates repositories
type RepositoryAuthenticator struct {
	// authManager is the authentication manager
	authManager AuthManager
}

// NewRepositoryAuthenticator creates a new repository authenticator
func NewRepositoryAuthenticator(authManager AuthManager) *RepositoryAuthenticator {
	return &RepositoryAuthenticator{
		authManager: authManager,
	}
}

// AuthenticateRepository authenticates a repository with credentials
func (a *RepositoryAuthenticator) AuthenticateRepository(ctx context.Context, repo repository.Repository, credID string) error {
	// Get credentials
	creds, err := a.authManager.GetCredentials(credID)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}
	
	// Authenticate with the authentication manager
	authenticated, err := a.authManager.Authenticate(ctx, creds)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	
	if !authenticated {
		return fmt.Errorf("invalid credentials")
	}
	
	// Apply credentials to repository config
	config := getRepositoryConfig(repo)
	if config == nil {
		return fmt.Errorf("failed to get repository config")
	}
	
	// Update config with credentials
	updateConfigWithCredentials(config, creds)
	
	return nil
}

// getRepositoryConfig gets the configuration from a repository
func getRepositoryConfig(repo repository.Repository) *repository.Config {
	// This is a simplified implementation
	// In a real system, we would need to access the repository's config directly
	
	// Create a new config based on the repository's type, name, and URL
	config := repository.NewConfig(
		repo.GetType(),
		repo.GetName(),
		repo.GetURL(),
	)
	
	return config
}

// updateConfigWithCredentials updates a repository configuration with credentials
func updateConfigWithCredentials(config *repository.Config, creds *Credentials) {
	switch creds.Type {
	case BasicAuth:
		config.Username = creds.Username
		config.Password = creds.Password
	case TokenAuth:
		// For GitHub and GitLab, the token is used as the password
		if creds.Provider == GitHubProvider || creds.Provider == GitLabProvider {
			config.Password = creds.Token
		} else {
			config.Username = creds.Username
			config.Password = creds.Token
		}
	case OAuthAuth:
		// OAuth tokens are used the same way as regular tokens
		config.Password = creds.Token
	case CertAuth:
		// Set certificate paths
		// This would require extending the repository.Config type
		// to support certificate-based authentication
	}
}

// AuthorizeRepositoryOperation authorizes a repository operation
func (a *RepositoryAuthenticator) AuthorizeRepositoryOperation(user *User, repo repository.Repository, operation Permission) error {
	// Check if user has permission
	if !a.authManager.HasPermission(user, operation) {
		return fmt.Errorf("user does not have permission to perform this operation")
	}
	
	return nil
}

// DefaultRepositoryAuthenticator is the default repository authenticator
var DefaultRepositoryAuthenticator *RepositoryAuthenticator

// InitDefaultRepositoryAuthenticator initializes the default repository authenticator
func InitDefaultRepositoryAuthenticator(authManager AuthManager) {
	DefaultRepositoryAuthenticator = NewRepositoryAuthenticator(authManager)
}
