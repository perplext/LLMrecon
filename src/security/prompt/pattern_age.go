// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
)

// patternAge is a helper struct for sorting patterns by age
type patternAge struct {
	pattern  string
	lastSeen time.Time
}
