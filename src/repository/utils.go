package repository

import (
	"net/url"
	"regexp"
	"strings"
)

// getRepositoryIDFromURL extracts a clean repository ID from a URL
// This is used for audit logging to create a consistent identifier
func getRepositoryIDFromURL(repoURL string) string {
	// Try to parse the URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		// If URL can't be parsed, use a sanitized version of the original
		return sanitizeURLForID(repoURL)
	}

	// Use hostname as the base for the ID
	id := parsedURL.Hostname()
	
	// If there's a path, append it (without leading/trailing slashes)
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		path := strings.Trim(parsedURL.Path, "/")
		if path != "" {
			id = id + "-" + path
		}
	}
	
	// Replace any non-alphanumeric characters with dashes
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	id = re.ReplaceAllString(id, "-")
	
	// Remove leading and trailing dashes
	id = strings.Trim(id, "-")
	
	return id
}

// sanitizeURLForID creates a safe ID from a URL string
// Used as a fallback when URL parsing fails
func sanitizeURLForID(input string) string {
	// Remove any protocol prefix
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "https://")
	
	// Replace any non-alphanumeric characters with dashes
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	id := re.ReplaceAllString(input, "-")
	
	// Remove leading and trailing dashes
	id = strings.Trim(id, "-")
	
	return id
}
