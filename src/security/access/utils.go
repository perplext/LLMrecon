// Package access provides access control and security auditing functionality
package access

import (
	"context"
)

// contextKey is a private type for context keys
type contextKey int

const (
	// userIDKey is the context key for user ID
	userIDKey contextKey = iota
)

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)

// formatTime formats a time.Time as a string
func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)

// parseTime parses a string as a time.Time
func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)

// Other utility functions are defined in their respective files to avoid conflicts:
// - getUserIDFromContext is in access_control_system.go
// - generateID is in auth_manager.go
// - containsString and removeString are in integration.go
