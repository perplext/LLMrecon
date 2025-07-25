// Package db provides database implementations of the access control interfaces
package db

import "time"

// formatTime formats a time.Time value to a string using RFC3339 format
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}