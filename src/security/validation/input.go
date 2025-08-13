package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// InputValidator provides methods for validating and sanitizing user input
type InputValidator struct {
	maxLength int
	allowedChars *regexp.Regexp
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxLength: 10000,
		allowedChars: regexp.MustCompile(`^[\w\s\-\.@/:\[\]\{\}\(\),"'!?;]+$`),
	}
}

// ValidatePrompt validates a prompt input for security issues
func (v *InputValidator) ValidatePrompt(prompt string) error {
	// Check length
	if len(prompt) > v.maxLength {
		return fmt.Errorf("prompt exceeds maximum length of %d characters", v.maxLength)
	}

	// Check for null bytes
	if strings.Contains(prompt, "\x00") {
		return fmt.Errorf("prompt contains null bytes")
	}

	// Check for control characters (except newline, tab, carriage return)
	for _, r := range prompt {
		if r < 32 && r != '\n' && r != '\t' && r != '\r' {
			return fmt.Errorf("prompt contains invalid control characters")
		}
	}

	return nil
}

// ValidateURL validates a URL for security issues
func (v *InputValidator) ValidateURL(rawURL string) error {
	// Parse the URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s", u.Scheme)
	}

	// Check for localhost/private IPs (prevent SSRF)
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.") {
		return fmt.Errorf("URLs to private networks are not allowed")
	}

	return nil
}

// ValidateFilePath validates a file path for security issues
func (v *InputValidator) ValidateFilePath(path string) error {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"/etc/passwd",
		"/etc/shadow",
		"~/.ssh/",
		".env",
		"id_rsa",
		".git/config",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerPath, pattern) {
			return fmt.Errorf("access to sensitive file paths is restricted")
		}
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from a string
func (v *InputValidator) SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline, tab, carriage return
	var sanitized strings.Builder
	for _, r := range input {
		if r >= 32 || r == '\n' || r == '\t' || r == '\r' {
			sanitized.WriteRune(r)
		}
	}

	// Truncate if too long
	result := sanitized.String()
	if len(result) > v.maxLength {
		result = result[:v.maxLength]
	}

	return result
}

// ValidateAPIKey validates an API key format
func (v *InputValidator) ValidateAPIKey(key string) error {
	// Check if empty
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Check length (most API keys are between 20-100 characters)
	if len(key) < 20 || len(key) > 200 {
		return fmt.Errorf("API key has invalid length")
	}

	// Check for spaces or control characters
	for _, r := range key {
		if r < 33 || r > 126 {
			return fmt.Errorf("API key contains invalid characters")
		}
	}

	return nil
}

// ValidateModelName validates a model name
func (v *InputValidator) ValidateModelName(name string) error {
	// Check if empty
	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Check length
	if len(name) > 100 {
		return fmt.Errorf("model name too long")
	}

	// Allow alphanumeric, dash, underscore, colon, slash, dot
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9\-_:/\.]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("model name contains invalid characters")
	}

	return nil
}