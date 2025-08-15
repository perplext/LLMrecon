// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
)

// BundleConverter converts standard bundles to offline bundles
type BundleConverter struct {
	// Creator is the offline bundle creator
	Creator *OfflineBundleCreator
	// Logger is the logger for conversion operations
	Logger io.Writer
	// AuditTrail is the audit trail manager for logging operations
	AuditTrail *trail.AuditTrailManager

// NewBundleConverter creates a new bundle converter
func NewBundleConverter(creator *OfflineBundleCreator, logger io.Writer, auditTrail *trail.AuditTrailManager) *BundleConverter {
	if logger == nil {
		logger = os.Stdout
	}

	return &BundleConverter{
		Creator:    creator,
		Logger:     logger,
		AuditTrail: auditTrail,
	}

// ConvertToOfflineBundle converts a standard bundle to an offline bundle
func (c *BundleConverter) ConvertToOfflineBundle(bundle *Bundle, outputPath string) (*OfflineBundle, error) {
	// Log conversion start
	fmt.Fprintf(c.Logger, "Converting bundle to offline format: %s\n", bundle.Manifest.Name)

	// Create enhanced manifest
	enhancedManifest := EnhancedBundleManifest{
		BundleManifest: bundle.Manifest,
		Changelog: []ChangelogEntry{
			{
				Version: bundle.Manifest.Version,
				Date:    time.Now().UTC(),
				Changes: []string{
					"Converted from standard bundle format",
				},
			},
		},
		Documentation: make(map[string]string),
	}

	// Create offline bundle
	offlineBundle := &OfflineBundle{
		Bundle: Bundle{
			BundlePath: outputPath,
			Manifest:   bundle.Manifest,
		},
		EnhancedManifest:   enhancedManifest,
		Format:             DefaultOfflineBundleFormat(),
		IsIncremental:      false,
		ComplianceMappings: []ComplianceMapping{},
	}

	// Create output directory
	if err := os.MkdirAll(outputPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create required directories
	for _, dir := range offlineBundle.Format.RequiredDirectories {
		dirPath := filepath.Join(outputPath, dir)
		if err := os.MkdirAll(dirPath, 0700); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Copy content from original bundle
	for _, item := range bundle.Manifest.Content {
		// Determine source path
		sourcePath := filepath.Join(bundle.BundlePath, item.Path)

		// Determine target directory based on content type
		var targetDir string
		switch item.Type {
		case TemplateContentType:
			targetDir = filepath.Join(outputPath, "templates")
		case ModuleContentType:
			targetDir = filepath.Join(outputPath, "modules")
		case ConfigContentType:
			targetDir = filepath.Join(outputPath, "config")
		case ResourceContentType:
			targetDir = filepath.Join(outputPath, "resources")
		default:
			return nil, fmt.Errorf("unsupported content type: %s", item.Type)
		}

		// Create target directory if it doesn't exist
		if err := os.MkdirAll(targetDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create target directory: %w", err)
		}
		// Determine target path
		targetPath := filepath.Join(targetDir, filepath.Base(item.Path))
		// Copy file
		sourceData, err := os.ReadFile(filepath.Clean(sourcePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read source file: %w", err)
		}

		if err := os.WriteFile(filepath.Clean(targetPath, sourceData, 0600)); err != nil {
			return nil, fmt.Errorf("failed to write target file: %w", err)
		}
	}

	// Create README.md file with basic information
	readmePath := filepath.Join(outputPath, "README.md")
	readmeContent := fmt.Sprintf(`# %s

%s

## Version

%s

## Bundle Type

%s

## Created

%s

## Author

%s (%s)

## Contents

This offline bundle contains templates and modules for LLM red teaming.

## Usage

See the documentation directory for usage instructions.
`, bundle.Manifest.Name, bundle.Manifest.Description, bundle.Manifest.Version, 
		bundle.Manifest.BundleType, bundle.Manifest.CreatedAt.Format(time.RFC3339),
		bundle.Manifest.Author.Name, bundle.Manifest.Author.Email)

	if err := os.WriteFile(filepath.Clean(readmePath, []byte(readmeContent)), 0600); err != nil {
		return nil, fmt.Errorf("failed to write README.md: %w", err)
	}

	// Add README.md to documentation
	enhancedManifest.Documentation["README"] = "README.md"

	// Create manifest file
	manifestPath := filepath.Join(outputPath, "manifest.json")
	if err := c.Creator.Generator.WriteEnhancedManifest(&enhancedManifest, manifestPath); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Update checksums
	if err := c.Creator.Generator.UpdateChecksumsForEnhancedManifest(&enhancedManifest, outputPath); err != nil {
		return nil, fmt.Errorf("failed to update checksums: %w", err)
	}

	// Sign manifest
	if err := c.Creator.Generator.SignEnhancedManifest(&enhancedManifest); err != nil {
		return nil, fmt.Errorf("failed to sign manifest: %w", err)
	}

	// Update bundle manifest
	offlineBundle.Manifest = enhancedManifest.BundleManifest
	offlineBundle.EnhancedManifest = enhancedManifest

	// Write updated manifest
	if err := c.Creator.Generator.WriteEnhancedManifest(&enhancedManifest, manifestPath); err != nil {
		return nil, fmt.Errorf("failed to write updated manifest: %w", err)
	}

	// Log audit event
	if c.AuditTrail != nil {
		auditLog := &trail.AuditLog{
			Operation:     "convert_to_offline_bundle",
			ResourceType:  "bundle",
			ResourceID:    bundle.Manifest.BundleID,
			Status:        "success",
			Timestamp:     time.Now(),
			UserID:        bundle.Manifest.Author.Email,
			Username:      bundle.Manifest.Author.Name,
			IPAddress:     "",
			Details: map[string]interface{}{
				"original_bundle_id": bundle.Manifest.BundleID,
				"original_version":   bundle.Manifest.Version,
			},
		}
		
		if err := c.AuditTrail.LogOperation(nil, auditLog); err != nil {
			fmt.Fprintf(c.Logger, "Warning: Failed to log audit event: %v\n", err)
		}
	}
	// Log conversion success
	fmt.Fprintf(c.Logger, "Bundle converted successfully to offline format: %s\n", outputPath)

	return offlineBundle, nil

// AutoDetectComplianceForTemplates automatically detects and adds compliance mappings for templates
func (c *BundleConverter) AutoDetectComplianceForTemplates(offlineBundle *OfflineBundle) error {
	// Log operation start
	fmt.Fprintf(c.Logger, "Auto-detecting compliance mappings for templates in bundle: %s\n", offlineBundle.Manifest.Name)

	// Iterate through content items
	for _, item := range offlineBundle.Manifest.Content {
		// Only process templates
		if item.Type != TemplateContentType {
			continue
		}

		// Read template file
		templatePath := filepath.Join(offlineBundle.BundlePath, item.Path)
		templateData, err := os.ReadFile(filepath.Clean(templatePath))
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		// Analyze template content for OWASP LLM Top 10 categories
		owaspCategories := detectOwaspCategories(string(templateData))

		// Analyze template content for ISO/IEC 42001 controls
		isoControls := detectISOControls(string(templateData))

		// Add compliance mappings if any were detected
		if len(owaspCategories) > 0 || len(isoControls) > 0 {
			err = c.Creator.AddComplianceMappingToOfflineBundle(offlineBundle, item.ID, owaspCategories, isoControls)
			if err != nil {
				return fmt.Errorf("failed to add compliance mappings for template %s: %w", item.ID, err)
			}
			
			fmt.Fprintf(c.Logger, "Added compliance mappings for template %s: OWASP categories: %v, ISO controls: %v\n", 
				item.ID, owaspCategories, isoControls)
		}
	}

	// Log operation success
	fmt.Fprintf(c.Logger, "Compliance mapping auto-detection completed successfully\n")

	return nil

// detectOwaspCategories analyzes template content and detects relevant OWASP LLM Top 10 categories
func detectOwaspCategories(templateContent string) []string {
	categories := []string{}
	
	// Simple keyword-based detection for demonstration purposes
	// In a real implementation, this would use more sophisticated analysis
	
	// LLM01: Prompt Injection
	if containsKeywords(templateContent, "prompt injection", "input validation", "sanitize input", "user input") {
		categories = append(categories, "LLM01:PromptInjection")
	}
	
	// LLM02: Insecure Output
	if containsKeywords(templateContent, "output validation", "validate response", "harmful output", "dangerous output") {
		categories = append(categories, "LLM02:InsecureOutput")
	}
	
	// LLM06: Sensitive Information Disclosure
	if containsKeywords(templateContent, "sensitive information", "personal data", "pii", "confidential") {
		categories = append(categories, "LLM06:SensitiveInformationDisclosure")
	}
	
	// LLM07: Insecure Plugin Design
	if containsKeywords(templateContent, "plugin", "extension", "module integration", "third-party") {
		categories = append(categories, "LLM07:InsecurePluginDesign")
	}
	
	// LLM08: Excessive Agency
	if containsKeywords(templateContent, "autonomous", "agency", "decision making", "authority") {
		categories = append(categories, "LLM08:ExcessiveAgency")
	}
	
	// LLM09: Overreliance
	if containsKeywords(templateContent, "verification", "human review", "oversight", "check accuracy") {
		categories = append(categories, "LLM09:Overreliance")
	}
	
	return categories

// detectISOControls analyzes template content and detects relevant ISO/IEC 42001 controls
func detectISOControls(templateContent string) []string {
	controls := []string{}
	
	// Simple keyword-based detection for demonstration purposes
	// In a real implementation, this would use more sophisticated analysis
	
	// 42001:8.2.3 - Risk assessment
	if containsKeywords(templateContent, "risk assessment", "risk analysis", "threat model") {
		controls = append(controls, "42001:8.2.3")
	}
	
	// 42001:8.2.4 - Risk treatment
	if containsKeywords(templateContent, "risk mitigation", "risk treatment", "control implementation") {
		controls = append(controls, "42001:8.2.4")
	}
	
	// 42001:9.2 - Internal audit
	if containsKeywords(templateContent, "audit", "review", "assessment", "evaluation") {
		controls = append(controls, "42001:9.2")
	}
	
	// 42001:10.1 - Nonconformity and corrective action
	if containsKeywords(templateContent, "nonconformity", "corrective action", "remediation", "fix") {
		controls = append(controls, "42001:10.1")
	}
	
	return controls

// containsKeywords checks if any of the keywords are present in the content
func containsKeywords(content string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
			return true
		}
	}
