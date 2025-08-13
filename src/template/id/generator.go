package id

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// Constants for ID generation and validation
const (
	// ID format patterns
	TemplateIDPattern     = `^[a-z0-9_]+(_v\d+\.\d+)?(_[a-f0-9]{8})?$`
	CategoryIDPattern     = `^[a-z0-9_]+$`
	ModuleIDPattern       = `^[a-z0-9_]+$`
	VersionPattern        = `^\d+\.\d+$`
	ToolVersionPattern    = `^\d+\.\d+\.\d+$`
	
	// Maximum lengths
	MaxIDLength           = 64
	MaxDescriptiveLength  = 32
)

// IDType represents the type of ID to generate
type IDType string

// Available ID types
const (
	TemplateID  IDType = "template"
	CategoryID  IDType = "category"
	ProviderID  IDType = "provider"
	UtilityID   IDType = "utility"
	DetectorID  IDType = "detector"
)

// Generator handles the generation and validation of unique IDs
type Generator struct {
	existingIDs map[string]bool
}

// NewGenerator creates a new ID generator
func NewGenerator() *Generator {
	return &Generator{
		existingIDs: make(map[string]bool),
	}
}

// RegisterExistingID registers an existing ID to avoid duplicates
func (g *Generator) RegisterExistingID(id string) {
	g.existingIDs[id] = true
}

// GenerateID generates a unique ID based on the provided parameters
func (g *Generator) GenerateID(idType IDType, category, name, version string) (string, error) {
	var id string
	
	switch idType {
	case TemplateID:
		// Format: category_descriptive-name_vX.Y
		descriptivePart := sanitizeForID(name)
		if len(descriptivePart) > MaxDescriptiveLength {
			descriptivePart = descriptivePart[:MaxDescriptiveLength]
		}
		
		id = fmt.Sprintf("%s_%s_v%s", sanitizeForID(category), descriptivePart, version)
		
	case CategoryID:
		// Format: category_name
		id = sanitizeForID(name)
		
	case ProviderID, UtilityID, DetectorID:
		// Format: descriptive_name
		id = sanitizeForID(name)
	}
	
	// Ensure uniqueness
	if g.existingIDs[id] {
		// If ID already exists, append a hash of the current timestamp
		hash := generateTimeHash()
		id = fmt.Sprintf("%s_%s", id, hash[:8])
	}
	
	// Register the new ID
	g.existingIDs[id] = true
	
	// Validate the generated ID
	if err := g.ValidateID(idType, id); err != nil {
		return "", err
	}
	
	return id, nil
}

// ValidateID validates an ID against the appropriate pattern
func (g *Generator) ValidateID(idType IDType, id string) error {
	if len(id) > MaxIDLength {
		return fmt.Errorf("ID exceeds maximum length of %d characters", MaxIDLength)
	}
	
	var pattern string
	
	switch idType {
	case TemplateID:
		pattern = TemplateIDPattern
	case CategoryID:
		pattern = CategoryIDPattern
	case ProviderID, UtilityID, DetectorID:
		pattern = ModuleIDPattern
	default:
		return fmt.Errorf("unknown ID type: %s", idType)
	}
	
	matched, err := regexp.MatchString(pattern, id)
	if err != nil {
		return fmt.Errorf("error validating ID: %w", err)
	}
	
	if !matched {
		return fmt.Errorf("ID '%s' does not match the required pattern for %s", id, idType)
	}
	
	return nil
}

// ValidateVersion validates a version string against the appropriate pattern
func (g *Generator) ValidateVersion(version string, isTool bool) error {
	pattern := VersionPattern
	if isTool {
		pattern = ToolVersionPattern
	}
	
	matched, err := regexp.MatchString(pattern, version)
	if err != nil {
		return fmt.Errorf("error validating version: %w", err)
	}
	
	if !matched {
		return fmt.Errorf("version '%s' does not match the required pattern", version)
	}
	
	return nil
}

// sanitizeForID converts a string to a format suitable for an ID
func sanitizeForID(input string) string {
	// Convert to lowercase
	result := strings.ToLower(input)
	
	// Replace spaces and hyphens with underscores
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, "-", "_")
	
	// Remove any characters that aren't alphanumeric or underscore
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	result = reg.ReplaceAllString(result, "")
	
	// Ensure it doesn't start with a number
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "id_" + result
	}
	
	return result
}

// generateTimeHash generates a hash based on the current time
func generateTimeHash() string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%d", timestamp)
	
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
