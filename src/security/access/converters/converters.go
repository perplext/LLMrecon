// Package converters provides conversion functions between different types
package converters

import (
	"github.com/perplext/LLMrecon/src/security/access/common"
	"github.com/perplext/LLMrecon/src/security/access/interfaces"
	"github.com/perplext/LLMrecon/src/security/access/models"
)

// StringToAuthMethod converts a string to an AuthMethod
func StringToAuthMethod(s string) common.AuthMethod {
	return common.AuthMethod(s)
}

// AuthMethodToString converts an AuthMethod to a string
func AuthMethodToString(m common.AuthMethod) string {
	return string(m)
}

// StringSliceToAuthMethodSlice converts a slice of strings to a slice of AuthMethods
func StringSliceToAuthMethodSlice(s []string) []common.AuthMethod {
	if s == nil {
		return nil
	}
	
	result := make([]common.AuthMethod, len(s))
	for i, v := range s {
		result[i] = common.AuthMethod(v)
	}
	return result
}

// AuthMethodSliceToStringSlice converts a slice of AuthMethods to a slice of strings
func AuthMethodSliceToStringSlice(m []common.AuthMethod) []string {
	if m == nil {
		return nil
	}
	
	result := make([]string, len(m))
	for i, v := range m {
		result[i] = string(v)
	}
	return result
}

// InterfaceUserToModelUser converts an interfaces.User to a models.User
func InterfaceUserToModelUser(user *interfaces.User) *models.User {
	if user == nil {
		return nil
	}

	return &models.User{
		ID:                 user.ID,
		Username:           user.Username,
		Email:              user.Email,
		PasswordHash:       user.PasswordHash,
		Roles:              user.Roles,
		Permissions:        user.Permissions,
		MFAEnabled:         user.MFAEnabled,
		MFAMethod:          user.MFAMethod,  // Keep as string
		MFAMethods:         user.MFAMethods, // Keep as []string
		MFASecret:          user.MFASecret,
		LastLogin:          user.LastLogin,
		LastPasswordChange: user.LastPasswordChange,
		FailedLoginAttempts: user.FailedLoginAttempts,
		Locked:             user.Locked,
		Active:             user.Active,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
		Metadata:           user.Metadata,
	}
}

// ModelUserToInterfaceUser converts a models.User to an interfaces.User
func ModelUserToInterfaceUser(user *models.User) *interfaces.User {
	if user == nil {
		return nil
	}

	return &interfaces.User{
		ID:                 user.ID,
		Username:           user.Username,
		Email:              user.Email,
		PasswordHash:       user.PasswordHash,
		Roles:              user.Roles,
		Permissions:        user.Permissions,
		MFAEnabled:         user.MFAEnabled,
		MFAMethod:          user.MFAMethod,  // Keep as string
		MFAMethods:         user.MFAMethods, // Keep as []string
		MFASecret:          user.MFASecret,
		LastLogin:          user.LastLogin,
		LastPasswordChange: user.LastPasswordChange,
		FailedLoginAttempts: user.FailedLoginAttempts,
		Locked:             user.Locked,
		Active:             user.Active,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
		Metadata:           user.Metadata,
	}
}

// InterfaceSessionToModelSession converts an interfaces.Session to a models.Session
func InterfaceSessionToModelSession(session *interfaces.Session) *models.Session {
	if session == nil {
		return nil
	}

	return &models.Session{
		ID:           session.ID,
		UserID:       session.UserID,
		Token:        session.Token,
		RefreshToken: session.RefreshToken,
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
		ExpiresAt:    session.ExpiresAt,
		LastActivity: session.LastActivity,
		MFACompleted: session.MFACompleted,
		CreatedAt:    session.CreatedAt,
		Metadata:     session.Metadata,
	}
}

// ModelSessionToInterfaceSession converts a models.Session to an interfaces.Session
func ModelSessionToInterfaceSession(session *models.Session) *interfaces.Session {
	if session == nil {
		return nil
	}

	return &interfaces.Session{
		ID:           session.ID,
		UserID:       session.UserID,
		Token:        session.Token,
		RefreshToken: session.RefreshToken,
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
		ExpiresAt:    session.ExpiresAt,
		LastActivity: session.LastActivity,
		MFACompleted: session.MFACompleted,
		CreatedAt:    session.CreatedAt,
		Metadata:     session.Metadata,
	}
}
