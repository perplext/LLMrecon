// Package access provides access control and security auditing functionality
package access

import "github.com/perplext/LLMrecon/src/security/access/models"

// LoginResult represents the result of a successful login attempt
type LoginResult struct {
	User    *models.User    `json:"user"`
	Session *models.Session `json:"session"`
	Token   string          `json:"token,omitempty"`       // Convenience field, same as Session.Token
	RefreshToken string     `json:"refresh_token,omitempty"` // Convenience field, same as Session.RefreshToken
}
