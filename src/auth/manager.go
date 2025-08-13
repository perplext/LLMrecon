package auth

import (
	"context"
	"fmt"
)

// Manager implements the AuthManager interface
type Manager struct {
	// credStore is the credential store
	credStore *CredentialStore
	
	// userStore is the user store
	userStore *UserStore
	
	// providers is a map of authentication providers
	providers map[ProviderType]AuthProvider
}

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// Authenticate authenticates with the provider
	Authenticate(ctx context.Context, creds *Credentials) (bool, error)
	
	// RefreshToken refreshes a token
	RefreshToken(ctx context.Context, creds *Credentials) error
}

// NewManager creates a new authentication manager
func NewManager(configDir string, passphrase string) (*Manager, error) {
	// Create credential store
	credStore, err := NewCredentialStore(
		filepath.Join(configDir, "credentials.enc"),
		passphrase,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential store: %w", err)
	}
	
	// Create user store
	userStore, err := NewUserStore(
		filepath.Join(configDir, "users.json"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user store: %w", err)
	}
	
	// Create manager
	manager := &Manager{
		credStore:  credStore,
		userStore:  userStore,
		providers:  make(map[ProviderType]AuthProvider),
	}
	
	// Register default providers
	manager.RegisterProvider(GitHubProvider, NewGitHubAuthProvider())
	manager.RegisterProvider(GitLabProvider, NewGitLabAuthProvider())
	manager.RegisterProvider(GenericProvider, NewGenericAuthProvider())
	
	return manager, nil
}

// RegisterProvider registers an authentication provider
func (m *Manager) RegisterProvider(providerType ProviderType, provider AuthProvider) {
	m.providers[providerType] = provider
}

// Authenticate authenticates a user with the given credentials
func (m *Manager) Authenticate(ctx context.Context, creds *Credentials) (bool, error) {
	// Get provider
	provider, exists := m.providers[creds.Provider]
	if !exists {
		return false, fmt.Errorf("unsupported authentication provider: %s", creds.Provider)
	}
	
	// Check if token is expired
	if creds.Type == TokenAuth || creds.Type == OAuthAuth {
		if !creds.TokenExpiry.IsZero() && creds.TokenExpiry.Before(time.Now()) {
			// Try to refresh token
			if err := m.RefreshToken(ctx, creds); err != nil {
				return false, fmt.Errorf("failed to refresh expired token: %w", err)
			}
		}
	}
	
	// Authenticate with provider
	authenticated, err := provider.Authenticate(ctx, creds)
	if err != nil {
		return false, err
	}
	
	// Update last used timestamp if authenticated
	if authenticated {
		if err := m.credStore.UpdateLastUsed(creds.ID); err != nil {
			// Log error but don't fail authentication
			fmt.Printf("Failed to update last used timestamp: %v\n", err)
		}
	}
	
	return authenticated, nil
}

// GetCredentials gets credentials by ID
func (m *Manager) GetCredentials(id string) (*Credentials, error) {
	return m.credStore.GetCredentials(id)
}

// SaveCredentials saves credentials
func (m *Manager) SaveCredentials(creds *Credentials) error {
	return m.credStore.SaveCredentials(creds)
}

// DeleteCredentials deletes credentials by ID
func (m *Manager) DeleteCredentials(id string) error {
	return m.credStore.DeleteCredentials(id)
}

// ListCredentials lists all credentials
func (m *Manager) ListCredentials() ([]*Credentials, error) {
	return m.credStore.ListCredentials()
}

// RefreshToken refreshes a token if it's expired
func (m *Manager) RefreshToken(ctx context.Context, creds *Credentials) error {
	// Get provider
	provider, exists := m.providers[creds.Provider]
	if !exists {
		return fmt.Errorf("unsupported authentication provider: %s", creds.Provider)
	}
	
	// Refresh token
	if err := provider.RefreshToken(ctx, creds); err != nil {
		return err
	}
	
	// Save updated credentials
	return m.SaveCredentials(creds)
}

// HasPermission checks if a user has a specific permission
func (m *Manager) HasPermission(user *User, permission Permission) bool {
	return m.userStore.HasPermission(user, permission)
}

// GetUser gets a user by ID
func (m *Manager) GetUser(id string) (*User, error) {
	return m.userStore.GetUser(id)
}

// SaveUser saves a user
func (m *Manager) SaveUser(user *User) error {
	return m.userStore.SaveUser(user)
}

// DeleteUser deletes a user by ID
func (m *Manager) DeleteUser(id string) error {
	return m.userStore.DeleteUser(id)
}

// ListUsers lists all users
func (m *Manager) ListUsers() ([]*User, error) {
	return m.userStore.ListUsers()
}

// UpdateLastLogin updates the last login timestamp for a user
func (m *Manager) UpdateLastLogin(id string) error {
	return m.userStore.UpdateLastLogin(id)
}

// GenerateCredentialID generates a unique ID for credentials
func GenerateCredentialID(provider ProviderType, name string) string {
	return fmt.Sprintf("%s-%s-%d", provider, name, time.Now().UnixNano())
}

// DefaultManager is the default authentication manager
var DefaultManager *Manager

// InitDefaultManager initializes the default authentication manager
func InitDefaultManager(configDir string, passphrase string) error {
	var err error
	DefaultManager, err = NewManager(configDir, passphrase)
	return err
}
