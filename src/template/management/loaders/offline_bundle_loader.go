// Package loaders provides template loaders for different sources
package loaders

import (
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

// OfflineBundleSource represents a source for offline bundles
const OfflineBundleSource = "offline_bundle"

// OfflineBundleLoader loads templates from offline bundles
type OfflineBundleLoader struct {
	// validator is the offline bundle validator
	validator *bundle.OfflineBundleValidator
	// auditTrail is the audit trail manager
	auditTrail *trail.AuditTrailManager
	// validationLevel is the level of validation to perform
	validationLevel bundle.ValidationLevel
}

// NewOfflineBundleLoader creates a new offline bundle loader
func NewOfflineBundleLoader(auditTrail *trail.AuditTrailManager) *OfflineBundleLoader {
	return &OfflineBundleLoader{
		validator:       bundle.NewOfflineBundleValidator(nil),
		auditTrail:      auditTrail,
		validationLevel: bundle.StandardValidation,
	}
}

// LoadFromSource loads templates from a specific source
func (l *OfflineBundleLoader) LoadFromSource(ctx context.Context, source types.TemplateSource, options map[string]interface{}) ([]*format.Template, error) {
	if source.Type != OfflineBundleSource {
		return nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}

	// Extract bundle path from options
	bundlePath, ok := options["bundle_path"].(string)
	if !ok || bundlePath == "" {
		return nil, fmt.Errorf("bundle_path is required in options")
	}

	// Load templates from the offline bundle
	return l.LoadFromPath(ctx, bundlePath, false)
}

// LoadFromPath loads templates from a specific path
func (l *OfflineBundleLoader) LoadFromPath(ctx context.Context, path string, recursive bool) ([]*format.Template, error) {
	// Load the offline bundle
	offlineBundle, err := bundle.OpenOfflineBundle(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open offline bundle: %w", err)
	}

	// Validate the bundle
	result, err := l.validator.ValidateOfflineBundle(offlineBundle, l.validationLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to validate offline bundle: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("invalid offline bundle: %s", result.Message)
	}

	// Extract templates from the bundle
	templates := []*format.Template{}
	for _, item := range offlineBundle.EnhancedManifest.Content {
		if item.Type == bundle.TemplateContentType {
			// Load the template file
			templatePath := filepath.Join(path, item.Path)
			template, err := format.LoadFromFile(templatePath)
			if err != nil {
				return nil, fmt.Errorf("failed to load template from %s: %w", templatePath, err)
			}

			// Add compliance mappings as metadata
			template.Metadata = make(map[string]interface{})
			
			// Check if this template has compliance mappings
			for _, mapping := range offlineBundle.GetComplianceMappings() {
				if mapping.ContentID == item.ID {
					if len(mapping.OwaspLLMCategories) > 0 {
						template.Metadata["owasp_llm_categories"] = mapping.OwaspLLMCategories
					}
					if len(mapping.ISOIECControls) > 0 {
						template.Metadata["iso_iec_controls"] = mapping.ISOIECControls
					}
					if mapping.Description != "" {
						template.Metadata["compliance_description"] = mapping.Description
					}
					break
				}
			}

			// Add documentation references
			for docType, docPath := range offlineBundle.EnhancedManifest.Documentation {
				if strings.Contains(docType, "template") && strings.Contains(docType, item.ID) {
					template.Metadata["documentation"] = filepath.Join(path, docPath)
					break
				}
			}

			templates = append(templates, template)
		}
	}

	// Log audit event
	if l.auditTrail != nil {
		auditLog := trail.NewAuditLog(
			"load_templates_from_offline_bundle",
			"offline_bundle",
			fmt.Sprintf("Loaded %d templates from offline bundle %s", len(templates), offlineBundle.EnhancedManifest.Name),
		).WithResource(offlineBundle.EnhancedManifest.BundleID).WithStatus("success").WithDetail("bundle_path", path).WithDetail("recursive", recursive).WithDetail("template_count", len(templates)).WithDetail("bundle_version", offlineBundle.EnhancedManifest.Version).WithDetail("validation_level", string(l.validationLevel))
		
		if err := l.auditTrail.LogOperation(context.Background(), auditLog); err != nil {
			// Log error but continue
			fmt.Printf("Warning: Failed to log audit event: %v\n", err)
		}
	}

	return templates, nil
}

// LoadFromRepository loads templates from a remote repository
func (l *OfflineBundleLoader) LoadFromRepository(ctx context.Context, repoURL string, options map[string]interface{}) ([]*format.Template, error) {
	// Not implemented for offline bundles
	return nil, fmt.Errorf("loading from repository not supported for offline bundles")
}

// SetValidationLevel sets the validation level for offline bundles
func (l *OfflineBundleLoader) SetValidationLevel(level bundle.ValidationLevel) {
	l.validationLevel = level
}

// GetValidationLevel gets the current validation level
func (l *OfflineBundleLoader) GetValidationLevel() bundle.ValidationLevel {
	return l.validationLevel
}

// LoadTemplate loads a single template from a source
func (l *OfflineBundleLoader) LoadTemplate(ctx context.Context, source string, sourceType string) (*format.Template, error) {
	if sourceType != OfflineBundleSource {
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}

	templates, err := l.LoadFromPath(ctx, source, false)
	if err != nil {
		return nil, err
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found in bundle")
	}

	return templates[0], nil
}

// LoadTemplates loads multiple templates from a source
func (l *OfflineBundleLoader) LoadTemplates(ctx context.Context, source string, sourceType string) ([]*format.Template, error) {
	if sourceType != OfflineBundleSource {
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}

	return l.LoadFromPath(ctx, source, false)
}
