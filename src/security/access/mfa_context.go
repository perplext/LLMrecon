package access

import (
	"context"
	"errors"

	"github.com/perplext/LLMrecon/src/security/access/common"
)

// MFAContextKey is the key used to store MFA information in the context
type MFAContextKey string

const (
	// MFAStatusKey is the key for MFA status in context
	MFAStatusKey MFAContextKey = "mfa_status"
	// MFAMethodKey is the key for MFA method in context
	MFAMethodKey MFAContextKey = "mfa_method"
	// MFAUserIDKey is the key for user ID in MFA context
	MFAUserIDKey MFAContextKey = "mfa_user_id"
)

// MFAStatus represents the status of MFA verification
type MFAStatus string

const (
	// MFAStatusRequired indicates that MFA is required but not completed
	MFAStatusRequired MFAStatus = "required"
	// MFAStatusCompleted indicates that MFA has been completed
	MFAStatusCompleted MFAStatus = "completed"
	// MFAStatusNotRequired indicates that MFA is not required
	MFAStatusNotRequired MFAStatus = "not_required"
)

// WithMFAStatus adds MFA status to context
func WithMFAStatus(ctx context.Context, status MFAStatus) context.Context {
	return context.WithValue(ctx, MFAStatusKey, status)
}

// GetMFAStatus gets MFA status from context
func GetMFAStatus(ctx context.Context) (MFAStatus, error) {
	status, ok := ctx.Value(MFAStatusKey).(MFAStatus)
	if !ok {
		return "", errors.New("MFA status not found in context")
	}
	return status, nil
}

// WithMFAMethod adds MFA method to context
func WithMFAMethod(ctx context.Context, method common.AuthMethod) context.Context {
	return context.WithValue(ctx, MFAMethodKey, method)
}

// GetMFAMethod gets MFA method from context
func GetMFAMethod(ctx context.Context) (common.AuthMethod, error) {
	method, ok := ctx.Value(MFAMethodKey).(common.AuthMethod)
	if !ok {
		return "", errors.New("MFA method not found in context")
	}
	return method, nil
}

// WithMFAUserID adds user ID to MFA context
func WithMFAUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, MFAUserIDKey, userID)
}

// GetMFAUserID gets user ID from MFA context
func GetMFAUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(MFAUserIDKey).(string)
	if !ok {
		return "", errors.New("MFA user ID not found in context")
	}
	return userID, nil
}

// CreateMFAContext creates a new context with MFA information
func CreateMFAContext(ctx context.Context, userID string, method common.AuthMethod, status MFAStatus) context.Context {
	ctx = WithMFAUserID(ctx, userID)
	ctx = WithMFAMethod(ctx, method)
	ctx = WithMFAStatus(ctx, status)
	return ctx
}

// IsMFACompleted checks if MFA has been completed in the context
func IsMFACompleted(ctx context.Context) bool {
	status, err := GetMFAStatus(ctx)
	if err != nil {
		return false
	}
	return status == MFAStatusCompleted
}

// IsMFARequired checks if MFA is required in the context
func IsMFARequired(ctx context.Context) bool {
	status, err := GetMFAStatus(ctx)
	if err != nil {
		return false
	}
	return status == MFAStatusRequired
}
