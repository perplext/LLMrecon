package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubAuthProvider implements authentication for GitHub
type GitHubAuthProvider struct{}

// NewGitHubAuthProvider creates a new GitHub authentication provider
func NewGitHubAuthProvider() *GitHubAuthProvider {
	return &GitHubAuthProvider{}
}

// Authenticate authenticates with GitHub
func (p *GitHubAuthProvider) Authenticate(ctx context.Context, creds *Credentials) (bool, error) {
	switch creds.Type {
	case BasicAuth:
		return p.authenticateBasic(ctx, creds)
	case TokenAuth:
		return p.authenticateToken(ctx, creds)
	case OAuthAuth:
		return p.authenticateOAuth(ctx, creds)
	default:
		return false, fmt.Errorf("unsupported authentication type for GitHub: %s", creds.Type)
	}
}

// authenticateBasic authenticates with GitHub using basic authentication
func (p *GitHubAuthProvider) authenticateBasic(ctx context.Context, creds *Credentials) (bool, error) {
	// GitHub no longer supports basic authentication for API access
	// Convert to token-based authentication if possible
	if creds.Token != "" {
		return p.authenticateToken(ctx, creds)
	}
	
	return false, fmt.Errorf("GitHub no longer supports basic authentication for API access")
}

// authenticateToken authenticates with GitHub using a token
func (p *GitHubAuthProvider) authenticateToken(ctx context.Context, creds *Credentials) (bool, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return false, err
	}
	
	// Add authorization header
	req.Header.Set("Authorization", "token "+creds.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("GitHub API returned status code %d", resp.StatusCode)
	}
	
	return true, nil
}

// authenticateOAuth authenticates with GitHub using OAuth
func (p *GitHubAuthProvider) authenticateOAuth(ctx context.Context, creds *Credentials) (bool, error) {
	// Check if token is valid
	if creds.Token != "" {
		return p.authenticateToken(ctx, creds)
	}
	
	// If no token, need to get one using the authorization code flow
	// This would typically be handled by a web application
	return false, fmt.Errorf("OAuth authentication requires a token")
}

// RefreshToken refreshes a GitHub token
func (p *GitHubAuthProvider) RefreshToken(ctx context.Context, creds *Credentials) error {
	// Only OAuth tokens can be refreshed
	if creds.Type != OAuthAuth {
		return fmt.Errorf("only OAuth tokens can be refreshed")
	}
	
	// Check if refresh token is available
	if creds.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}
	
	// Create request
	data := url.Values{}
	data.Set("client_id", creds.ClientID)
	data.Set("client_secret", creds.ClientSecret)
	data.Set("refresh_token", creds.RefreshToken)
	data.Set("grant_type", "refresh_token")
	
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	
	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status code %d", resp.StatusCode)
	}
	
	// Parse response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}
	
	// Update credentials
	creds.Token = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		creds.RefreshToken = tokenResp.RefreshToken
	}
	
	// Set expiry time
	if tokenResp.ExpiresIn > 0 {
		creds.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}
	
	return nil
}

// GitLabAuthProvider implements authentication for GitLab
type GitLabAuthProvider struct{}

// NewGitLabAuthProvider creates a new GitLab authentication provider
func NewGitLabAuthProvider() *GitLabAuthProvider {
	return &GitLabAuthProvider{}
}

// Authenticate authenticates with GitLab
func (p *GitLabAuthProvider) Authenticate(ctx context.Context, creds *Credentials) (bool, error) {
	switch creds.Type {
	case BasicAuth:
		return p.authenticateBasic(ctx, creds)
	case TokenAuth:
		return p.authenticateToken(ctx, creds)
	case OAuthAuth:
		return p.authenticateOAuth(ctx, creds)
	default:
		return false, fmt.Errorf("unsupported authentication type for GitLab: %s", creds.Type)
	}
}

// authenticateBasic authenticates with GitLab using basic authentication
func (p *GitLabAuthProvider) authenticateBasic(ctx context.Context, creds *Credentials) (bool, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return false, err
	}
	
	// Add authorization header
	req.SetBasicAuth(creds.Username, creds.Password)
	
	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("GitLab API returned status code %d", resp.StatusCode)
	}
	
	return true, nil
}

// authenticateToken authenticates with GitLab using a token
func (p *GitLabAuthProvider) authenticateToken(ctx context.Context, creds *Credentials) (bool, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return false, err
	}
	
	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+creds.Token)
	
	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("GitLab API returned status code %d", resp.StatusCode)
	}
	
	return true, nil
}

// authenticateOAuth authenticates with GitLab using OAuth
func (p *GitLabAuthProvider) authenticateOAuth(ctx context.Context, creds *Credentials) (bool, error) {
	// Check if token is valid
	if creds.Token != "" {
		return p.authenticateToken(ctx, creds)
	}
	
	// If no token, need to get one using the authorization code flow
	// This would typically be handled by a web application
	return false, fmt.Errorf("OAuth authentication requires a token")
}

// RefreshToken refreshes a GitLab token
func (p *GitLabAuthProvider) RefreshToken(ctx context.Context, creds *Credentials) error {
	// Only OAuth tokens can be refreshed
	if creds.Type != OAuthAuth {
		return fmt.Errorf("only OAuth tokens can be refreshed")
	}
	
	// Check if refresh token is available
	if creds.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}
	
	// Create request
	data := url.Values{}
	data.Set("client_id", creds.ClientID)
	data.Set("client_secret", creds.ClientSecret)
	data.Set("refresh_token", creds.RefreshToken)
	data.Set("grant_type", "refresh_token")
	
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://gitlab.com/oauth/token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab API returned status code %d", resp.StatusCode)
	}
	
	// Parse response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}
	
	// Update credentials
	creds.Token = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		creds.RefreshToken = tokenResp.RefreshToken
	}
	
	// Set expiry time
	if tokenResp.ExpiresIn > 0 {
		creds.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}
	
	return nil
}

// GenericAuthProvider implements authentication for generic repositories
type GenericAuthProvider struct{}

// NewGenericAuthProvider creates a new generic authentication provider
func NewGenericAuthProvider() *GenericAuthProvider {
	return &GenericAuthProvider{}
}

// Authenticate authenticates with a generic repository
func (p *GenericAuthProvider) Authenticate(ctx context.Context, creds *Credentials) (bool, error) {
	// For generic repositories, we just assume the credentials are valid
	// as there's no standard way to validate them
	return true, nil
}

// RefreshToken refreshes a token for a generic repository
func (p *GenericAuthProvider) RefreshToken(ctx context.Context, creds *Credentials) error {
	// Generic repositories don't support token refresh
	return fmt.Errorf("token refresh not supported for generic repositories")
}
