// Package customization provides customization detection and management
package customization

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CustomizationDetector detects user customizations
type CustomizationDetector struct {
	BasePath string
}

// NewCustomizationDetector creates a new customization detector
func NewCustomizationDetector(basePath string) *CustomizationDetector {
	return &CustomizationDetector{
		BasePath: basePath,
	}
}

// DetectCustomizations detects customizations in the given path
func (d *CustomizationDetector) DetectCustomizations() ([]Customization, error) {
	var customizations []Customization
	
	err := filepath.Walk(d.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && d.isCustomizationFile(path) {
			custom := Customization{
				Path:        path,
				Type:        "file",
				Description: "User customization detected",
			}
			customizations = append(customizations, custom)
		}
		
		return nil
	})
	
	return customizations, err
}

// isCustomizationFile checks if a file is a user customization
func (d *CustomizationDetector) isCustomizationFile(path string) bool {
	// Simple heuristic - check for common customization patterns
	base := filepath.Base(path)
	return strings.Contains(base, "custom") || 
		   strings.Contains(base, "user") ||
		   strings.HasSuffix(base, ".custom")
}

// Customization represents a detected customization
type Customization struct {
	Path        string
	Type        string
	Description string
	Hash        string
}

// CalculateHash calculates the hash of the customization
func (c *Customization) CalculateHash() error {
	content, err := os.ReadFile(filepath.Clean(c.Path))
	if err != nil {
		return err
	}
	
	hash := sha256.Sum256(content)
	c.Hash = fmt.Sprintf("%x", hash)
	return nil
}
