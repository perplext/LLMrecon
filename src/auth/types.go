package auth

import (
	"context"
)

// AuthType represents the type of authentication
type AuthType string

const (
	// BasicAuth represents username/password authentication
	BasicAuth AuthType = "basic"
	// TokenAuth represents token-based authentication
	TokenAuth AuthType = "token"
	// OAuthAuth represents OAuth authentication
	OAuthAuth AuthType = "oauth"
	// CertAuth represents certificate-based authentication
	CertAuth AuthType = "cert"
	// NoneAuth represents no authentication
	NoneAuth AuthType = "none"
)

// ProviderType represents the type of authentication provider
type ProviderType string

const (
	// GitHubProvider represents GitHub authentication
	GitHubProvider ProviderType = "github"
	// GitLabProvider represents GitLab authentication
	GitLabProvider ProviderType = "gitlab"
	// GenericProvider represents a generic authentication provider
	GenericProvider ProviderType = "generic"
)

// Credentials represents authentication credentials
type Credentials struct {
	// ID is a unique identifier for the credentials
	ID string `json:"id"`
	
	// Name is a human-readable name for the credentials
	Name string `json:"name"`
	
	// Type is the type of authentication
	Type AuthType `json:"type"`
	
	// Provider is the authentication provider
	Provider ProviderType `json:"provider"`
	
	// Username is the username for basic authentication
	Username string `json:"username,omitempty"`
	
	// Password is the password for basic authentication
	// This field should be encrypted when stored
	Password string `json:"password,omitempty"`
	
	// Token is the token for token-based authentication
	// This field should be encrypted when stored
	Token string `json:"token,omitempty"`
	
	// TokenExpiry is the expiry time for the token
	TokenExpiry time.Time `json:"token_expiry,omitempty"`
	
	// RefreshToken is the refresh token for OAuth authentication
	// This field should be encrypted when stored
	RefreshToken string `json:"refresh_token,omitempty"`
	
	// ClientID is the client ID for OAuth authentication
	ClientID string `json:"client_id,omitempty"`
	
	// ClientSecret is the client secret for OAuth authentication
	// This field should be encrypted when stored
	ClientSecret string `json:"client_secret,omitempty"`
	
	// CertPath is the path to the certificate file for certificate-based authentication
	CertPath string `json:"cert_path,omitempty"`
	
	// KeyPath is the path to the private key file for certificate-based authentication
	KeyPath string `json:"key_path,omitempty"`
	
	// CACertPath is the path to the CA certificate file for certificate-based authentication
	CACertPath string `json:"ca_cert_path,omitempty"`
	
	// Scopes is the list of scopes for OAuth authentication
	Scopes []string `json:"scopes,omitempty"`
	
	// CreatedAt is the time the credentials were created
	CreatedAt time.Time `json:"created_at"`
	
	// UpdatedAt is the time the credentials were last updated
	UpdatedAt time.Time `json:"updated_at"`
	
	// LastUsedAt is the time the credentials were last used
	LastUsedAt time.Time `json:"last_used_at,omitempty"`

// Role represents a role for role-based access control
type Role string

const (
	// AdminRole has full access to all operations
	AdminRole Role = "admin"
	// UserRole has limited access to operations
	UserRole Role = "user"
	// ReadOnlyRole has read-only access to operations
	ReadOnlyRole Role = "readonly"
)

// Permission represents a permission for role-based access control
type Permission string

const (
	// ReadPermission allows reading operations
	ReadPermission Permission = "read"
	// WritePermission allows writing operations
	WritePermission Permission = "write"
	// DeletePermission allows deleting operations
	DeletePermission Permission = "delete"
	// AdminPermission allows administrative operations
	AdminPermission Permission = "admin"
)

// User represents a user of the system
type User struct {
	// ID is a unique identifier for the user
	ID string `json:"id"`
	
	// Username is the username of the user
	Username string `json:"username"`
	
	// Role is the role of the user
	Role Role `json:"role"`
	
	// Permissions is the list of permissions for the user
	Permissions []Permission `json:"permissions"`
	
	// CreatedAt is the time the user was created
	CreatedAt time.Time `json:"created_at"`
	
	// UpdatedAt is the time the user was last updated
	UpdatedAt time.Time `json:"updated_at"`
	
	// LastLoginAt is the time the user last logged in
	LastLoginAt time.Time `json:"last_login_at,omitempty"`

// Authenticator defines the interface for authentication
type Authenticator interface {
	// Authenticate authenticates a user with the given credentials
	Authenticate(ctx context.Context, creds *Credentials) (bool, error)
	
	// GetCredentials gets credentials by ID
	GetCredentials(id string) (*Credentials, error)
	
	// SaveCredentials saves credentials
	SaveCredentials(creds *Credentials) error
	
	// DeleteCredentials deletes credentials by ID
	DeleteCredentials(id string) error
	
	// ListCredentials lists all credentials
	ListCredentials() ([]*Credentials, error)
	
	// RefreshToken refreshes a token if it's expired
	RefreshToken(ctx context.Context, creds *Credentials) error

// Authorizer defines the interface for authorization
type Authorizer interface {
	// HasPermission checks if a user has a specific permission
	HasPermission(user *User, permission Permission) bool
	
	// GetUser gets a user by ID
	GetUser(id string) (*User, error)
	
	// SaveUser saves a user
	SaveUser(user *User) error
	
	// DeleteUser deletes a user by ID
	DeleteUser(id string) error
	
	// ListUsers lists all users
	ListUsers() ([]*User, error)

// AuthManager combines authentication and authorization
type AuthManager interface {
	Authenticator
	Authorizer
