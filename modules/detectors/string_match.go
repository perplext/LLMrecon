package detectors

import (
	"strings"
)

// StringMatchDetector implements the Detector interface for string matching
type StringMatchDetector struct {
	Match     string
	Condition string // "contains" or "not_contains"
	CaseSensitive bool
}

// NewStringMatchDetector creates a new string match detector
func NewStringMatchDetector(match string, condition string, caseSensitive bool) *StringMatchDetector {
	return &StringMatchDetector{
		Match:     match,
		Condition: condition,
		CaseSensitive: caseSensitive,
	}
}

// Detect checks if the response matches the detection criteria
func (d *StringMatchDetector) Detect(response string) (bool, string) {
	var found bool
	
	// Handle case sensitivity
	compareResponse := response
	compareMatch := d.Match
	
	if !d.CaseSensitive {
		compareResponse = strings.ToLower(response)
		compareMatch = strings.ToLower(d.Match)
	}
	
	// Check if the string is found
	found = strings.Contains(compareResponse, compareMatch)
	
	// Determine if this is a vulnerability based on the condition
	var isVulnerable bool
	var reason string
	
	switch d.Condition {
	case "contains":
		isVulnerable = found
		if isVulnerable {
			reason = "Response contains the prohibited string: " + d.Match
		} else {
			reason = "Response does not contain the prohibited string: " + d.Match
		}
	case "not_contains":
		isVulnerable = !found
		if isVulnerable {
			reason = "Response does not contain the required string: " + d.Match
		} else {
			reason = "Response contains the required string: " + d.Match
		}
	default:
		reason = "Invalid condition: " + d.Condition
		isVulnerable = false
	}
	
	return isVulnerable, reason
}

// GetInfo returns information about the detector
func (d *StringMatchDetector) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":           "string_match",
		"match":          d.Match,
		"condition":      d.Condition,
		"case_sensitive": d.CaseSensitive,
	}
}
