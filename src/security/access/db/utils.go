// Package db provides database implementations of the access control interfaces
package db


// formatTime formats a time.Time value to a string using RFC3339 format
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
