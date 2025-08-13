package format

import (
	"fmt"
	"regexp"
	"strings"
)

// Regex patterns for filename sanitization
var (
	spaceRegex    = regexp.MustCompile(`\s+`)
	filenameRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
)

// EnsureDir ensures that a directory exists
func EnsureDir(dir string) error {
	// Check if directory already exists
	info, err := os.Stat(dir)
	if err == nil {
		if info.IsDir() {
			return nil // Directory already exists
		}
		return fmt.Errorf("path exists but is not a directory: %s", dir)
	}
	
	// Create directory with all parent directories
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	
	return err
}

// GetTemplatePath returns the full path for a template file
func GetTemplatePath(baseDir, category, name, version string) string {
	// Sanitize name for filename
	filename := fmt.Sprintf("%s_v%s.yaml", SanitizeFilename(name), version)
	
	// Construct path
	return filepath.Join(baseDir, category, filename)
}

// GetModulePath returns the full path for a module file
func GetModulePath(baseDir, moduleType, name, version string) string {
	// Sanitize name for filename
	filename := fmt.Sprintf("%s_v%s.yaml", SanitizeFilename(name), version)
	
	// Determine subdirectory based on module type
	var subdir string
	switch moduleType {
	case "provider":
		subdir = "providers"
	case "utility":
		subdir = "utils"
	case "detector":
		subdir = "detectors"
	default:
		subdir = moduleType
	}
	
	// Construct path
	return filepath.Join(baseDir, subdir, filename)
}

// SanitizeFilename sanitizes a string for use as a filename
func SanitizeFilename(name string) string {
	// Replace spaces with underscores
	result := spaceRegex.ReplaceAllString(name, "_")
	
	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	result = filenameRegex.ReplaceAllString(result, "")
	
	// Convert to lowercase
	result = strings.ToLower(result)
	
	return result
}

// ListTemplates lists all template files in a directory
func ListTemplates(baseDir string) ([]string, error) {
	var templates []string
	
	// Walk through the base directory
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Check if file has YAML extension
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			// Get relative path from base directory
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			
			templates = append(templates, relPath)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	
	return templates, nil
}

// ListModules lists all module files in a directory
func ListModules(baseDir string) (map[string][]string, error) {
	modules := make(map[string][]string)
	modules["providers"] = []string{}
	modules["utils"] = []string{}
	modules["detectors"] = []string{}
	
	// Walk through the base directory
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Check if file has YAML extension
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			// Get relative path from base directory
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			
			// Determine module type based on directory
			dir := filepath.Dir(relPath)
			if dir == "providers" || dir == "utils" || dir == "detectors" {
				modules[dir] = append(modules[dir], relPath)
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}
	
	return modules, nil
}
