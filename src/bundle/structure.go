// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BundleStructure defines the structure of an offline bundle
type BundleStructure struct {
	// RootDirectories defines the directories that should exist in the bundle root
	RootDirectories map[string]DirectorySpec
	// FileNamingRules defines the rules for naming files in the bundle
	FileNamingRules map[string][]string
	// TemplateCategories defines the standard categories for organizing templates
	TemplateCategories []string
	// ModuleTypes defines the standard types for organizing modules
	ModuleTypes []string
}

// DirectorySpec defines the specification for a directory in the bundle
type DirectorySpec struct {
	// Required indicates whether the directory is required
	Required bool
	// Description provides a description of the directory
	Description string
	// Subdirectories defines the subdirectories that should exist
	Subdirectories map[string]DirectorySpec
	// FileExtensions defines the allowed file extensions in this directory
	FileExtensions []string
	// NamingConvention defines the naming convention for files in this directory
	NamingConvention string
}

// DefaultBundleStructure returns the default bundle structure specification
func DefaultBundleStructure() *BundleStructure {
	return &BundleStructure{
		RootDirectories: map[string]DirectorySpec{
			"manifest.json": {
				Required:    true,
				Description: "Enhanced bundle manifest containing metadata, content inventory, and compliance mappings",
			},
			"README.md": {
				Required:    true,
				Description: "Human-readable overview of the bundle and its contents",
			},
			"templates": {
				Required:    true,
				Description: "Contains all template files organized by category",
				Subdirectories: map[string]DirectorySpec{
					"owasp-llm": {
						Required:    false,
						Description: "OWASP LLM Top 10 categories",
						Subdirectories: map[string]DirectorySpec{
							"llm01-prompt-injection": {
								Required:    false,
								Description: "Prompt injection templates",
							},
							"llm02-insecure-output": {
								Required:    false,
								Description: "Insecure output handling templates",
							},
							"llm03-training-data-poisoning": {
								Required:    false,
								Description: "Training data poisoning templates",
							},
							"llm04-model-denial-of-service": {
								Required:    false,
								Description: "Model denial of service templates",
							},
							"llm05-supply-chain": {
								Required:    false,
								Description: "Supply chain vulnerability templates",
							},
							"llm06-sensitive-information-disclosure": {
								Required:    false,
								Description: "Sensitive information disclosure templates",
							},
							"llm07-insecure-plugin-design": {
								Required:    false,
								Description: "Insecure plugin design templates",
							},
							"llm08-excessive-agency": {
								Required:    false,
								Description: "Excessive agency templates",
							},
							"llm09-overreliance": {
								Required:    false,
								Description: "Overreliance templates",
							},
							"llm10-model-theft": {
								Required:    false,
								Description: "Model theft templates",
							},
						},
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
					"iso-42001": {
						Required:    false,
						Description: "ISO/IEC 42001 controls",
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
					"custom": {
						Required:    false,
						Description: "Custom templates",
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
				},
				FileExtensions:   []string{".json"},
				NamingConvention: "lowercase-with-hyphens",
			},
			"modules": {
				Required:    false,
				Description: "Contains modules that extend the functionality of templates",
				Subdirectories: map[string]DirectorySpec{
					"providers": {
						Required:    false,
						Description: "LLM provider modules",
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
					"detectors": {
						Required:    false,
						Description: "Vulnerability detection modules",
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
					"utilities": {
						Required:    false,
						Description: "Utility modules",
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
				},
				FileExtensions:   []string{".json"},
				NamingConvention: "lowercase-with-hyphens",
			},
			"binary": {
				Required:    false,
				Description: "Contains binary executables for the LLMreconing Tool",
				FileExtensions:   []string{"", ".exe"},
				NamingConvention: "tool-vX.Y.Z-OS-ARCH",
			},
			"documentation": {
				Required:    true,
				Description: "Contains comprehensive documentation for the bundle",
				Subdirectories: map[string]DirectorySpec{
					"compliance": {
						Required:    false,
						Description: "Compliance documentation",
						FileExtensions:   []string{".md"},
						NamingConvention: "lowercase-with-hyphens",
					},
				},
				FileExtensions:   []string{".md"},
				NamingConvention: "lowercase-with-hyphens",
			},
			"signatures": {
				Required:    true,
				Description: "Contains cryptographic signatures for bundle verification",
				Subdirectories: map[string]DirectorySpec{
					"content": {
						Required:    false,
						Description: "Signatures for content files",
					},
				},
				FileExtensions:   []string{".sig", ".pem"},
				NamingConvention: "same-as-signed-file",
			},
			"resources": {
				Required:    false,
				Description: "Contains additional resources used by templates or modules",
				Subdirectories: map[string]DirectorySpec{
					"images": {
						Required:    false,
						Description: "Image resources",
						FileExtensions:   []string{".png", ".jpg", ".svg"},
						NamingConvention: "lowercase-with-hyphens",
					},
					"data": {
						Required:    false,
						Description: "Data files",
						FileExtensions:   []string{".json", ".csv", ".txt"},
						NamingConvention: "lowercase-with-hyphens",
					},
					"schemas": {
						Required:    false,
						Description: "JSON schemas",
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
				},
				NamingConvention: "lowercase-with-hyphens",
			},
			"config": {
				Required:    false,
				Description: "Contains configuration files for templates and modules",
				Subdirectories: map[string]DirectorySpec{
					"environments": {
						Required:    false,
						Description: "Environment-specific configurations",
						FileExtensions:   []string{".json"},
						NamingConvention: "lowercase-with-hyphens",
					},
				},
				FileExtensions:   []string{".json"},
				NamingConvention: "lowercase-with-hyphens",
			},
			"repository-config": {
				Required:    false,
				Description: "Contains configuration for template repositories",
				FileExtensions:   []string{".json"},
				NamingConvention: "lowercase-with-hyphens",
			},
		},
		FileNamingRules: map[string][]string{
			"lowercase-with-hyphens": {
				"^[a-z0-9]+(-[a-z0-9]+)*$",
			},
			"tool-vX.Y.Z-OS-ARCH": {
				"^tool-v[0-9]+\\.[0-9]+\\.[0-9]+-[a-z]+-[a-z0-9]+(\\.exe)?$",
			},
			"same-as-signed-file": {
				"^.+\\.sig$",
			},
		},
		TemplateCategories: []string{
			"owasp-llm/llm01-prompt-injection",
			"owasp-llm/llm02-insecure-output",
			"owasp-llm/llm03-training-data-poisoning",
			"owasp-llm/llm04-model-denial-of-service",
			"owasp-llm/llm05-supply-chain",
			"owasp-llm/llm06-sensitive-information-disclosure",
			"owasp-llm/llm07-insecure-plugin-design",
			"owasp-llm/llm08-excessive-agency",
			"owasp-llm/llm09-overreliance",
			"owasp-llm/llm10-model-theft",
			"iso-42001",
			"custom",
		},
		ModuleTypes: []string{
			"providers",
			"detectors",
			"utilities",
		},
	}
}

// BundleStructureValidator validates the structure of a bundle
type BundleStructureValidator struct {
	// Structure is the bundle structure specification
	Structure *BundleStructure
}

// NewBundleStructureValidator creates a new bundle structure validator
func NewBundleStructureValidator() *BundleStructureValidator {
	return &BundleStructureValidator{
		Structure: DefaultBundleStructure(),
	}
}

// ValidateStructure validates the structure of a bundle
func (v *BundleStructureValidator) ValidateStructure(bundlePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		IsValid:   true,
		Level:     StructureValidationLevel,
		Message:   "Bundle structure validation successful",
		Errors:    []string{},
		Warnings:  []string{},
		Details:   make(map[string]interface{}),
		Timestamp: GetCurrentTime(),
	}

	// Check if the bundle path exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		result.Valid = false
		result.IsValid = false
		result.Message = "Bundle path does not exist"
		result.Errors = append(result.Errors, "Bundle path does not exist")
		return result, nil
	}

	// Validate required directories
	missingRequired := []string{}
	for dirName, dirSpec := range v.Structure.RootDirectories {
		if dirSpec.Required {
			path := filepath.Join(bundlePath, dirName)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				missingRequired = append(missingRequired, dirName)
			}
		}
	}

	if len(missingRequired) > 0 {
		result.Valid = false
		result.IsValid = false
		result.Message = "Missing required directories or files"
		for _, dir := range missingRequired {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing required directory or file: %s", dir))
		}
	}

	// Validate directory contents
	for dirName, dirSpec := range v.Structure.RootDirectories {
		path := filepath.Join(bundlePath, dirName)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			// Check if it's a directory
			fileInfo, _ := os.Stat(path)
			if fileInfo.IsDir() {
				// Validate subdirectories
				if err := v.validateDirectory(path, dirSpec, result); err != nil {
					result.Valid = false
					result.IsValid = false
					result.Message = "Error validating directory structure"
					result.Errors = append(result.Errors, fmt.Sprintf("Error validating directory %s: %v", dirName, err))
				}
			}
		}
	}

	// Add validation details
	result.Details["validated_path"] = bundlePath
	result.Details["missing_required"] = missingRequired

	return result, nil
}

// validateDirectory validates a directory against its specification
func (v *BundleStructureValidator) validateDirectory(dirPath string, dirSpec DirectorySpec, result *ValidationResult) error {
	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Check file extensions and naming conventions
	for _, entry := range entries {
		if !entry.IsDir() {
			// Validate file extension
			if len(dirSpec.FileExtensions) > 0 {
				ext := filepath.Ext(entry.Name())
				validExt := false
				for _, allowedExt := range dirSpec.FileExtensions {
					if ext == allowedExt {
						validExt = true
						break
					}
				}
				if !validExt {
					result.Warnings = append(result.Warnings, fmt.Sprintf("File %s has invalid extension", filepath.Join(dirPath, entry.Name())))
				}
			}

			// Validate naming convention
			if dirSpec.NamingConvention != "" {
				if !v.validateNamingConvention(entry.Name(), dirSpec.NamingConvention) {
					result.Warnings = append(result.Warnings, fmt.Sprintf("File %s does not follow naming convention %s", filepath.Join(dirPath, entry.Name()), dirSpec.NamingConvention))
				}
			}
		}
	}

	// Validate subdirectories
	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(dirPath, entry.Name())
			if subDirSpec, ok := dirSpec.Subdirectories[entry.Name()]; ok {
				if err := v.validateDirectory(subDirPath, subDirSpec, result); err != nil {
					return fmt.Errorf("failed to validate subdirectory %s: %w", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

// validateNamingConvention validates a file name against a naming convention
func (v *BundleStructureValidator) validateNamingConvention(fileName, convention string) bool {
	// For now, just check if the convention exists
	// In a real implementation, this would use regex patterns from FileNamingRules
	_, exists := v.Structure.FileNamingRules[convention]
	return exists
}

// CreateBundleStructure creates the directory structure for a new bundle
func CreateBundleStructure(bundlePath string, structure *BundleStructure) error {
	// Create the bundle directory if it doesn't exist
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}

	// Create required directories
	for dirName, dirSpec := range structure.RootDirectories {
		if strings.Contains(dirName, ".") {
			// This is a file, not a directory
			continue
		}
		
		path := filepath.Join(bundlePath, dirName)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirName, err)
		}

		// Create subdirectories
		if err := createSubdirectories(path, dirSpec.Subdirectories); err != nil {
			return fmt.Errorf("failed to create subdirectories for %s: %w", dirName, err)
		}
	}

	return nil
}

// createSubdirectories recursively creates subdirectories
func createSubdirectories(parentPath string, subdirs map[string]DirectorySpec) error {
	for dirName, dirSpec := range subdirs {
		path := filepath.Join(parentPath, dirName)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory %s: %w", dirName, err)
		}

		// Recursively create subdirectories
		if err := createSubdirectories(path, dirSpec.Subdirectories); err != nil {
			return fmt.Errorf("failed to create subdirectories for %s: %w", dirName, err)
		}
	}

	return nil
}

// GetTemplateCategory returns the appropriate category for a template
func (s *BundleStructure) GetTemplateCategory(templateType string) string {
	switch {
	case strings.HasPrefix(templateType, "llm01"):
		return "owasp-llm/llm01-prompt-injection"
	case strings.HasPrefix(templateType, "llm02"):
		return "owasp-llm/llm02-insecure-output"
	case strings.HasPrefix(templateType, "llm03"):
		return "owasp-llm/llm03-training-data-poisoning"
	case strings.HasPrefix(templateType, "llm04"):
		return "owasp-llm/llm04-model-denial-of-service"
	case strings.HasPrefix(templateType, "llm05"):
		return "owasp-llm/llm05-supply-chain"
	case strings.HasPrefix(templateType, "llm06"):
		return "owasp-llm/llm06-sensitive-information-disclosure"
	case strings.HasPrefix(templateType, "llm07"):
		return "owasp-llm/llm07-insecure-plugin-design"
	case strings.HasPrefix(templateType, "llm08"):
		return "owasp-llm/llm08-excessive-agency"
	case strings.HasPrefix(templateType, "llm09"):
		return "owasp-llm/llm09-overreliance"
	case strings.HasPrefix(templateType, "llm10"):
		return "owasp-llm/llm10-model-theft"
	case strings.HasPrefix(templateType, "iso"):
		return "iso-42001"
	default:
		return "custom"
	}
}

// GetModuleType returns the appropriate type for a module
func (s *BundleStructure) GetModuleType(moduleType string) string {
	switch {
	case strings.Contains(moduleType, "provider"):
		return "providers"
	case strings.Contains(moduleType, "detector"):
		return "detectors"
	default:
		return "utilities"
	}
}

// GetTemplatePath returns the path for a template within the bundle
func (s *BundleStructure) GetTemplatePath(bundlePath, templateID, templateType string) string {
	category := s.GetTemplateCategory(templateType)
	fileName := fmt.Sprintf("%s.json", templateID)
	return filepath.Join(bundlePath, "templates", category, fileName)
}

// GetModulePath returns the path for a module within the bundle
func (s *BundleStructure) GetModulePath(bundlePath, moduleID, moduleType string) string {
	category := s.GetModuleType(moduleType)
	fileName := fmt.Sprintf("%s.json", moduleID)
	return filepath.Join(bundlePath, "modules", category, fileName)
}

// GetDocumentationPath returns the path for a documentation file within the bundle
func (s *BundleStructure) GetDocumentationPath(bundlePath, docType string) string {
	if strings.HasPrefix(docType, "compliance/") {
		parts := strings.Split(docType, "/")
		if len(parts) > 1 {
			return filepath.Join(bundlePath, "documentation", "compliance", fmt.Sprintf("%s.md", parts[1]))
		}
	}
	return filepath.Join(bundlePath, "documentation", fmt.Sprintf("%s.md", docType))
}

// GetSignaturePath returns the path for a signature file within the bundle
func (s *BundleStructure) GetSignaturePath(bundlePath, filePath string) string {
	// For manifest.json, the signature is in the root signatures directory
	if filepath.Base(filePath) == "manifest.json" {
		return filepath.Join(bundlePath, "signatures", "manifest.sig")
	}
	
	// For other files, the signature is in the content directory with the same path
	relPath, _ := filepath.Rel(bundlePath, filePath)
	return filepath.Join(bundlePath, "signatures", "content", fmt.Sprintf("%s.sig", relPath))
}

// GetPublicKeyPath returns the path for the public key file within the bundle
func (s *BundleStructure) GetPublicKeyPath(bundlePath string) string {
	return filepath.Join(bundlePath, "signatures", "public-key.pem")
}

// StructureValidationLevel is a constant for structure validation level
const StructureValidationLevel = "structure"

// GetCurrentTime returns the current time
func GetCurrentTime() time.Time {
	return time.Now()
}
