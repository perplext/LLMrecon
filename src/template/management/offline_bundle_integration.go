// Package management provides functionality for managing templates in the LLMreconing Tool.
package management

import (
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/loaders"
)

// OfflineBundleSource is a constant for the offline bundle source
const OfflineBundleSource = loaders.OfflineBundleSource

// RegisterOfflineBundleLoader registers an offline bundle loader with the template manager
func RegisterOfflineBundleLoader(manager *DefaultTemplateManager, auditTrail *trail.AuditTrailManager) {
	loader := loaders.NewOfflineBundleLoader(auditTrail)
	manager.loaders = append(manager.loaders, loader)

// LoadFromOfflineBundle loads templates from an offline bundle
func (m *DefaultTemplateManager) LoadFromOfflineBundle(ctx context.Context, bundlePath string, validationLevel bundle.ValidationLevel) ([]*format.Template, error) {
	// Find the offline bundle loader
	var offlineBundleLoader *loaders.OfflineBundleLoader
	for _, loader := range m.loaders {
		if l, ok := loader.(*loaders.OfflineBundleLoader); ok {
			offlineBundleLoader = l
			break
		}
	}

	// If no offline bundle loader is registered, register one
	if offlineBundleLoader == nil {
		offlineBundleLoader = loaders.NewOfflineBundleLoader(nil)
		m.loaders = append(m.loaders, offlineBundleLoader)
	}

	// Set the validation level
	offlineBundleLoader.SetValidationLevel(validationLevel)

	// Load templates from the offline bundle
	templates, err := offlineBundleLoader.LoadFromPath(ctx, bundlePath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates from offline bundle: %w", err)
	}

	// Register the templates
	for _, template := range templates {
		if err := m.registry.Register(template); err != nil {
			return nil, fmt.Errorf("failed to register template %s: %w", template.ID, err)
		}

		if m.cache != nil {
			m.cache.Set(template.ID, template)
		}
	}

	return templates, nil

// CreateOfflineBundleRepository creates a repository for an offline bundle
func CreateOfflineBundleRepository(bundlePath string, auditTrail *trail.AuditTrailManager) (*repository.OfflineBundleRepository, error) {
	repo := repository.NewOfflineBundleRepository(bundlePath, auditTrail)
	
	// Connect to the repository
	if err := repo.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to offline bundle repository: %w", err)
	}
	
	return repo, nil

// LoadTemplatesFromOfflineBundleRepository loads templates from an offline bundle repository
func (m *DefaultTemplateManager) LoadTemplatesFromOfflineBundleRepository(ctx context.Context, repo *repository.OfflineBundleRepository) ([]*format.Template, error) {
	if !repo.IsConnected() {
		return nil, fmt.Errorf("repository not connected")
	}

	// Get the offline bundle
	offlineBundle := repo.GetOfflineBundle()
	if offlineBundle == nil {
		return nil, fmt.Errorf("offline bundle not loaded")
	}

	// Load templates from the repository
	var templates []*format.Template
	
	// Get the list of template files
	files, err := repo.ListFiles(ctx, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to list template files: %w", err)
	}

	// Load each template
	for _, file := range files {
		// Skip directories
		if file.IsDirectory {
			continue
		}
		
		// Skip non-template files based on extension
		if !strings.HasSuffix(file.Path, ".yaml") && !strings.HasSuffix(file.Path, ".yml") && !strings.HasSuffix(file.Path, ".json") {
			continue
		}

		// Get the file content
		content, err := repo.GetFile(ctx, file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to get template file %s: %w", file.Path, err)
		}

		// Parse the template
		template, err := format.ParseTemplate(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", file.Path, err)
		}

		// Add metadata from file info
		if template.Metadata == nil {
			template.Metadata = make(map[string]interface{})
		}
		template.Metadata["path"] = file.Path
		template.Metadata["size"] = file.Size
		template.Metadata["modTime"] = file.LastModified

		// Register the template
		if err := m.registry.Register(template); err != nil {
			return nil, fmt.Errorf("failed to register template %s: %w", template.ID, err)
		}

		if m.cache != nil {
			m.cache.Set(template.ID, template)
		}

		templates = append(templates, template)
	}

