// Package interfaces defines the interfaces for the access control system
package interfaces

import (
	"errors"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserLocked         = errors.New("user account is locked")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrMFARequired        = errors.New("multi-factor authentication required")
	ErrInvalidMFACode     = errors.New("invalid MFA code")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrNotFound           = errors.New("resource not found")
)
