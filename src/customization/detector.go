package customization

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
)

// Detector identifies user customizations in templates and modules
type Detector struct {
	// Registry is the customization registry
	Registry *Registry
	// BaseDir is the directory containing the original templates and modules
	BaseDir string
	// CustomizedDir is the directory containing the customized templates and modules
	CustomizedDir string
}

// DetectorOptions contains options for the Detector
type DetectorOptions struct {
	// Registry is the customization registry
	Registry *Registry
	// BaseDir is the directory containing the original templates and modules
	BaseDir string
	// CustomizedDir is the directory containing the customized templates and modules
	CustomizedDir string
}

// NewDetector creates a new customization detector
func NewDetector(options *DetectorOptions) (*Detector, error) {
	if options.Registry == nil {
		return nil, fmt.Errorf("registry is required")
	}
	if options.BaseDir == "" {
		return nil, fmt.Errorf("base directory is required")
	}
	if options.CustomizedDir == "" {
		return nil, fmt.Errorf("customized directory is required")
	}

	return &Detector{
		Registry:      options.Registry,
		BaseDir:       options.BaseDir,
		CustomizedDir: options.CustomizedDir,
	}, nil
}

// DetectTemplateCustomizations detects customizations in templates
func (d *Detector) DetectTemplateCustomizations() ([]*CustomizationEntry, error) {
	templatesDir := filepath.Join(d.CustomizedDir, "templates")
	baseTemplatesDir := filepath.Join(d.BaseDir, "templates")

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return nil, nil
	}

	var customizations []*CustomizationEntry

	// Walk through templates directory
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Check if file exists in base directory
		basePath := filepath.Join(baseTemplatesDir, relPath)
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			// File doesn't exist in base, it's a new file
			return nil
		}

		// Read file contents
		customizedContent, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read customized file: %w", err)
		}

		baseContent, err := ioutil.ReadFile(basePath)
		if err != nil {
			return fmt.Errorf("failed to read base file: %w", err)
		}

		// Calculate hashes
		customizedHash := calculateHash(customizedContent)
		baseHash := calculateHash(baseContent)

		// If hashes are different, it's a customization
		if customizedHash != baseHash {
			// Extract template ID from path
			templateID := extractTemplateID(relPath)
			if templateID == "" {
				return nil
			}

			// Get template version
			templateVersion, err := getTemplateVersion(basePath)
			if err != nil {
				return fmt.Errorf("failed to get template version: %w", err)
			}

			// Detect customization markers
			markers, err := detectCustomizationMarkers(string(customizedContent))
			if err != nil {
				return fmt.Errorf("failed to detect customization markers: %w", err)
			}

			// Create customization entry
			entry := &CustomizationEntry{
				ID:               fmt.Sprintf("template-%s-%s", templateID, filepath.Base(relPath)),
				Type:             TemplateCustomization,
				Path:             relPath,
				ComponentID:      templateID,
				BaseVersion:      templateVersion,
				CustomizationDate: time.Now(),
				OriginalHash:     baseHash,
				CustomizedHash:   customizedHash,
				Markers:          markers,
				Policy:           determinePreservationPolicy(markers),
			}

			customizations = append(customizations, entry)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk templates directory: %w", err)
	}

	return customizations, nil
}

// DetectModuleCustomizations detects customizations in modules
func (d *Detector) DetectModuleCustomizations() ([]*CustomizationEntry, error) {
	modulesDir := filepath.Join(d.CustomizedDir, "modules")
	baseModulesDir := filepath.Join(d.BaseDir, "modules")

	// Check if modules directory exists
	if _, err := os.Stat(modulesDir); os.IsNotExist(err) {
		return nil, nil
	}

	var customizations []*CustomizationEntry

	// Walk through modules directory
	err := filepath.Walk(modulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(modulesDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Check if file exists in base directory
		basePath := filepath.Join(baseModulesDir, relPath)
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			// File doesn't exist in base, it's a new file
			return nil
		}

		// Read file contents
		customizedContent, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read customized file: %w", err)
		}

		baseContent, err := ioutil.ReadFile(basePath)
		if err != nil {
			return fmt.Errorf("failed to read base file: %w", err)
		}

		// Calculate hashes
		customizedHash := calculateHash(customizedContent)
		baseHash := calculateHash(baseContent)

		// If hashes are different, it's a customization
		if customizedHash != baseHash {
			// Extract module ID from path
			parts := strings.Split(relPath, string(os.PathSeparator))
			if len(parts) < 1 {
				return nil
			}
			moduleID := parts[0]

			// Get module version
			moduleVersion, err := getModuleVersion(filepath.Join(baseModulesDir, moduleID))
			if err != nil {
				return fmt.Errorf("failed to get module version: %w", err)
			}

			// Detect customization markers
			markers, err := detectCustomizationMarkers(string(customizedContent))
			if err != nil {
				return fmt.Errorf("failed to detect customization markers: %w", err)
			}

			// Create customization entry
			entry := &CustomizationEntry{
				ID:               fmt.Sprintf("module-%s-%s", moduleID, filepath.Base(relPath)),
				Type:             ModuleCustomization,
				Path:             relPath,
				ComponentID:      moduleID,
				BaseVersion:      moduleVersion,
				CustomizationDate: time.Now(),
				OriginalHash:     baseHash,
				CustomizedHash:   customizedHash,
				Markers:          markers,
				Policy:           determinePreservationPolicy(markers),
			}

			customizations = append(customizations, entry)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk modules directory: %w", err)
	}

	return customizations, nil
}

// RegisterCustomizations registers all detected customizations in the registry
func (d *Detector) RegisterCustomizations() error {
	// Detect template customizations
	templateCustomizations, err := d.DetectTemplateCustomizations()
	if err != nil {
		return fmt.Errorf("failed to detect template customizations: %w", err)
	}

	// Detect module customizations
	moduleCustomizations, err := d.DetectModuleCustomizations()
	if err != nil {
		return fmt.Errorf("failed to detect module customizations: %w", err)
	}

	// Register all customizations
	for _, entry := range templateCustomizations {
		d.Registry.AddEntry(entry)
	}

	for _, entry := range moduleCustomizations {
		d.Registry.AddEntry(entry)
	}

	// Save registry
	if err := d.Registry.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	return nil
}

// Helper functions

// calculateHash calculates the SHA-256 hash of content
func calculateHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// extractTemplateID extracts the template ID from the path
func extractTemplateID(path string) string {
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) < 1 {
		return ""
	}
	return parts[0]
}

// getTemplateVersion gets the version of a template
func getTemplateVersion(path string) (string, error) {
	template, err := format.LoadFromFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}
	return template.Info.Version, nil
}

// getModuleVersion gets the version of a module
func getModuleVersion(moduleDir string) (string, error) {
	// Look for module.yaml or module.yml
	var modulePath string
	for _, name := range []string{"module.yaml", "module.yml"} {
		path := filepath.Join(moduleDir, name)
		if _, err := os.Stat(path); err == nil {
			modulePath = path
			break
		}
	}

	if modulePath == "" {
		return "", fmt.Errorf("module definition file not found")
	}

	module, err := format.LoadModuleFromFile(modulePath)
	if err != nil {
		return "", fmt.Errorf("failed to load module: %w", err)
	}
	return module.Info.Version, nil
}

// detectCustomizationMarkers detects customization markers in content
func detectCustomizationMarkers(content string) ([]CustomizationMarker, error) {
	var markers []CustomizationMarker

	// Define marker patterns
	patterns := map[string]*regexp.Regexp{
		"user_customization": regexp.MustCompile(`(?m)^.*?USER CUSTOMIZATION BEGIN.*?$\n(.*?)^.*?USER CUSTOMIZATION END.*?$`),
		"custom_code":        regexp.MustCompile(`(?m)^.*?CUSTOM CODE BEGIN.*?$\n(.*?)^.*?CUSTOM CODE END.*?$`),
		"do_not_modify":      regexp.MustCompile(`(?m)^.*?DO NOT MODIFY BEGIN.*?$\n(.*?)^.*?DO NOT MODIFY END.*?$`),
	}

	// Scan content line by line to get line numbers
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		lineNum++
	}

	// Find markers
	for markerType, _ := range patterns {
		for i, line := range lines {
			if strings.Contains(line, fmt.Sprintf("%s BEGIN", strings.ToUpper(strings.Replace(markerType, "_", " ", -1)))) {
				// Find end marker
				endLine := -1
				for j := i + 1; j < len(lines); j++ {
					if strings.Contains(lines[j], fmt.Sprintf("%s END", strings.ToUpper(strings.Replace(markerType, "_", " ", -1)))) {
						endLine = j
						break
					}
				}

				if endLine != -1 {
					// Extract content between markers
					markerContent := strings.Join(lines[i+1:endLine], "\n")
					
					marker := CustomizationMarker{
						Type:      markerType,
						StartLine: i,
						EndLine:   endLine,
						Content:   markerContent,
					}
					markers = append(markers, marker)
				}
			}
		}
	}

	return markers, nil
}

// determinePreservationPolicy determines the preservation policy based on markers
func determinePreservationPolicy(markers []CustomizationMarker) PreservationPolicy {
	for _, marker := range markers {
		switch marker.Type {
		case "do_not_modify":
			return AlwaysPreserve
		case "user_customization", "custom_code":
			return PreserveWithConflictResolution
		}
	}
	return AskUser
}
