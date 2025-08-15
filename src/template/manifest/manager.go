package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	
	"github.com/perplext/LLMrecon/src/template/format"
)

// Manager handles loading and saving of manifest files
type Manager struct {
	templateManifest *TemplateManifest
	moduleManifest   *ModuleManifest
	basePath         string
	mutex            sync.RWMutex

// NewManager creates a new manifest manager
}
func NewManager(basePath string) *Manager {
	return &Manager{
		templateManifest: NewTemplateManifest(),
		moduleManifest:   NewModuleManifest(),
		basePath:         basePath,
	}

// LoadManifests loads both template and module manifests
func (m *Manager) LoadManifests() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Load template manifest
	if err := m.loadTemplateManifest(); err != nil {
		return fmt.Errorf("failed to load template manifest: %w", err)
	}
	
	// Load module manifest
	if err := m.loadModuleManifest(); err != nil {
		return fmt.Errorf("failed to load module manifest: %w", err)
	}
	
	return nil

// SaveManifests saves both template and module manifests
func (m *Manager) SaveManifests() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Update last updated timestamp
	m.templateManifest.LastUpdated = time.Now().Format(time.RFC3339)
	m.moduleManifest.LastUpdated = time.Now().Format(time.RFC3339)
	
	// Save template manifest
	if err := m.saveTemplateManifest(); err != nil {
		return fmt.Errorf("failed to save template manifest: %w", err)
	}
	
	// Save module manifest
	if err := m.saveModuleManifest(); err != nil {
		return fmt.Errorf("failed to save module manifest: %w", err)
	}
	
	return nil

// GetTemplateManifest returns the template manifest
func (m *Manager) GetTemplateManifest() *TemplateManifest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.templateManifest

// GetModuleManifest returns the module manifest
func (m *Manager) GetModuleManifest() *ModuleManifest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.moduleManifest

// RegisterTemplate registers a template in the manifest
func (m *Manager) RegisterTemplate(template *format.Template) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if template already exists
	now := time.Now().Format(time.RFC3339)
	_, exists := m.templateManifest.Templates[template.ID]
	
	// Create template entry
	entry := TemplateEntry{
		ID:          template.ID,
		Name:        template.Info.Name,
		Description: template.Info.Description,
		Version:     template.Info.Version,
		Author:      template.Info.Author,
		Category:    getCategoryFromID(template.ID),
		Severity:    template.Info.Severity,
		Tags:        template.Info.Tags,
		Path:        getRelativePathForTemplate(template),
		UpdatedAt:   now,
	}
	
	// Set added timestamp if new
	if !exists {
		entry.AddedAt = now
	}
	
	// Add to manifest
	m.templateManifest.Templates[template.ID] = entry
	
	// Update category info
	category := getCategoryFromID(template.ID)
	if _, exists := m.templateManifest.Categories[category]; !exists {
		m.templateManifest.Categories[category] = CategoryInfo{
			Name:        category,
			Description: fmt.Sprintf("%s vulnerability templates", category),
			Templates:   []string{},
		}
	}
	
	// Add template to category if not already present
	categoryInfo := m.templateManifest.Categories[category]
	found := false
	for _, id := range categoryInfo.Templates {
		if id == template.ID {
			found = true
			break
		}
	}
	
	if !found {
		categoryInfo.Templates = append(categoryInfo.Templates, template.ID)
		m.templateManifest.Categories[category] = categoryInfo
	}
	
	return nil

// RegisterModule registers a module in the manifest
func (m *Manager) RegisterModule(module *format.Module) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if module already exists
	now := time.Now().Format(time.RFC3339)
	_, exists := m.moduleManifest.Modules[module.ID]
	
	// Create module entry
	entry := ModuleEntry{
		ID:          module.ID,
		Type:        string(module.Type),
		Name:        module.Info.Name,
		Description: module.Info.Description,
		Version:     module.Info.Version,
		Author:      module.Info.Author,
		Tags:        module.Info.Tags,
		Path:        getRelativePathForModule(module),
		UpdatedAt:   now,
	}
	
	// Set added timestamp if new
	if !exists {
		entry.AddedAt = now
	}
	
	// Add to manifest
	m.moduleManifest.Modules[module.ID] = entry
	
	// Update type info
	moduleType := string(module.Type)
	if _, exists := m.moduleManifest.Types[moduleType]; !exists {
		m.moduleManifest.Types[moduleType] = TypeInfo{
			Name:        moduleType,
			Description: fmt.Sprintf("%s modules", moduleType),
			Modules:     []string{},
		}
	}
	
	// Add module to type if not already present
	typeInfo := m.moduleManifest.Types[moduleType]
	found := false
	for _, id := range typeInfo.Modules {
		if id == module.ID {
			found = true
			break
		}
	}
	
	if !found {
		typeInfo.Modules = append(typeInfo.Modules, module.ID)
		m.moduleManifest.Types[moduleType] = typeInfo
	}
	
	return nil

// UnregisterTemplate removes a template from the manifest
func (m *Manager) UnregisterTemplate(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if template exists
	_, exists := m.templateManifest.Templates[id]
	if !exists {
		return fmt.Errorf("template %s not found in manifest", id)
	}
	
	// Remove from templates
	delete(m.templateManifest.Templates, id)
	
	// Remove from category
	category := getCategoryFromID(id)
	if categoryInfo, exists := m.templateManifest.Categories[category]; exists {
		var templates []string
		for _, templateID := range categoryInfo.Templates {
			if templateID != id {
				templates = append(templates, templateID)
			}
		}
		categoryInfo.Templates = templates
		m.templateManifest.Categories[category] = categoryInfo
		
		// Remove category if empty
		if len(templates) == 0 {
			delete(m.templateManifest.Categories, category)
		}
	}
	
	return nil

// UnregisterModule removes a module from the manifest
func (m *Manager) UnregisterModule(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if module exists
	moduleEntry, exists := m.moduleManifest.Modules[id]
	if !exists {
		return fmt.Errorf("module %s not found in manifest", id)
	}
	
	// Remove from modules
	delete(m.moduleManifest.Modules, id)
	
	// Remove from type
	moduleType := moduleEntry.Type
	if typeInfo, exists := m.moduleManifest.Types[moduleType]; exists {
		var modules []string
		for _, moduleID := range typeInfo.Modules {
			if moduleID != id {
				modules = append(modules, moduleID)
			}
		}
		typeInfo.Modules = modules
		m.moduleManifest.Types[moduleType] = typeInfo
		
		// Remove type if empty
		if len(modules) == 0 {
			delete(m.moduleManifest.Types, moduleType)
		}
	}
	
	return nil

// ScanAndRegisterTemplates scans the templates directory and registers all templates
func (m *Manager) ScanAndRegisterTemplates() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Reset template manifest
	m.templateManifest = NewTemplateManifest()
	
	// Get templates directory
	templatesDir := filepath.Join(m.basePath, "templates")
	
	// List all template files
	templateFiles, err := format.ListTemplates(templatesDir)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}
	
	// Load and register each template
	for _, relPath := range templateFiles {
		fullPath := filepath.Join(templatesDir, relPath)
		
		template, err := format.LoadFromFile(fullPath)
		if err != nil {
			// Log error but continue with other templates
			fmt.Printf("Warning: Failed to load template %s: %v\n", fullPath, err)
			continue
		}
		
		// Create template entry
		entry := TemplateEntry{
			ID:          template.ID,
			Name:        template.Info.Name,
			Description: template.Info.Description,
			Version:     template.Info.Version,
			Author:      template.Info.Author,
			Category:    getCategoryFromID(template.ID),
			Severity:    template.Info.Severity,
			Tags:        template.Info.Tags,
			Path:        relPath,
			AddedAt:     time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		}
		
		// Add to manifest
		m.templateManifest.Templates[template.ID] = entry
		
		// Update category info
		category := getCategoryFromID(template.ID)
		if _, exists := m.templateManifest.Categories[category]; !exists {
			m.templateManifest.Categories[category] = CategoryInfo{
				Name:        category,
				Description: fmt.Sprintf("%s vulnerability templates", category),
				Templates:   []string{},
			}
		}
		
		// Add template to category if not already present
		categoryInfo := m.templateManifest.Categories[category]
		found := false
		for _, id := range categoryInfo.Templates {
			if id == template.ID {
				found = true
				break
			}
		}
		
		if !found {
			categoryInfo.Templates = append(categoryInfo.Templates, template.ID)
			m.templateManifest.Categories[category] = categoryInfo
		}
	}
	
	return nil

// ScanAndRegisterModules scans the modules directory and registers all modules
func (m *Manager) ScanAndRegisterModules() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Reset module manifest
	m.moduleManifest = NewModuleManifest()
	
	// Get modules directory
	modulesDir := filepath.Join(m.basePath, "modules")
	
	// List all module files
	moduleFiles, err := format.ListModules(modulesDir)
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}
	
	// Load and register each module
	for moduleType, paths := range moduleFiles {
		for _, relPath := range paths {
			fullPath := filepath.Join(modulesDir, relPath)
			
			module, err := format.LoadModuleFromFile(fullPath)
			if err != nil {
				// Log error but continue with other modules
				fmt.Printf("Warning: Failed to load module %s: %v\n", fullPath, err)
				continue
			}
			
			// Create module entry
			entry := ModuleEntry{
				ID:          module.ID,
				Type:        string(module.Type),
				Name:        module.Info.Name,
				Description: module.Info.Description,
				Version:     module.Info.Version,
				Author:      module.Info.Author,
				Tags:        module.Info.Tags,
				Path:        relPath,
				AddedAt:     time.Now().Format(time.RFC3339),
				UpdatedAt:   time.Now().Format(time.RFC3339),
			}
			
			// Add to manifest
			m.moduleManifest.Modules[module.ID] = entry
			
			// Update type info
			if _, exists := m.moduleManifest.Types[moduleType]; !exists {
				m.moduleManifest.Types[moduleType] = TypeInfo{
					Name:        moduleType,
					Description: fmt.Sprintf("%s modules", moduleType),
					Modules:     []string{},
				}
			}
			
			// Add module to type if not already present
			typeInfo := m.moduleManifest.Types[moduleType]
			found := false
			for _, id := range typeInfo.Modules {
				if id == module.ID {
					found = true
					break
				}
			}
			
			if !found {
				typeInfo.Modules = append(typeInfo.Modules, module.ID)
				m.moduleManifest.Types[moduleType] = typeInfo
			}
		}
	}
	
	return nil
	

// loadTemplateManifest loads the template manifest from file
func (m *Manager) loadTemplateManifest() error {
	// Get manifest file path
	manifestPath := filepath.Join(m.basePath, "templates", "manifest.json")
	
	// Check if file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Create new manifest if file doesn't exist
		m.templateManifest = NewTemplateManifest()
		return nil
	}
	
	// Read file
	data, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return fmt.Errorf("failed to read template manifest file: %w", err)
	}
	
	// Parse JSON
	var manifest TemplateManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse template manifest file: %w", err)
	}
	
	m.templateManifest = &manifest
	return nil
// loadModuleManifest loads the module manifest from file
func (m *Manager) loadModuleManifest() error {
	// Get manifest file path
	manifestPath := filepath.Join(m.basePath, "modules", "manifest.json")
	
	// Check if file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Create new manifest if file doesn't exist
		m.moduleManifest = NewModuleManifest()
		return nil
	}
	
	// Read file
	data, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return fmt.Errorf("failed to read module manifest file: %w", err)
	}
	
	// Parse JSON
	var manifest ModuleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse module manifest file: %w", err)
	}
	
	m.moduleManifest = &manifest
	return nil

// saveTemplateManifest saves the template manifest to file
func (m *Manager) saveTemplateManifest() error {
	// Get manifest file path
	manifestPath := filepath.Join(m.basePath, "templates", "manifest.json")
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(manifestPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(m.templateManifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template manifest to JSON: %w", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(manifestPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write template manifest file: %w", err)
	}
	
	return nil

// saveModuleManifest saves the module manifest to file
func (m *Manager) saveModuleManifest() error {
	// Get manifest file path
	manifestPath := filepath.Join(m.basePath, "modules", "manifest.json")
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(manifestPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(m.moduleManifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal module manifest to JSON: %w", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(manifestPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write module manifest file: %w", err)
	}
	
	return nil

// Helper functions

// getCategoryFromID extracts the category from a template ID
func getCategoryFromID(id string) string {
	// Template IDs are in the format: category_name_vX.Y
	// Extract the category part
	parts := strings.Split(id, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"

// getRelativePathForTemplate returns the relative path for a template
func getRelativePathForTemplate(template *format.Template) string {
	category := getCategoryFromID(template.ID)
	filename := fmt.Sprintf("%s_v%s.yaml", 
		format.SanitizeFilename(template.Info.Name), 
		template.Info.Version)
	return filepath.Join(category, filename)

// getRelativePathForModule returns the relative path for a module
func getRelativePathForModule(module *format.Module) string {
	var subdir string
	switch module.Type {
	case format.ProviderModule:
		subdir = "providers"
	case format.UtilityModule:
		subdir = "utils"
	case format.DetectorModule:
		subdir = "detectors"
	default:
		subdir = string(module.Type)
	}
	
	filename := fmt.Sprintf("%s_v%s.yaml", 
		format.SanitizeFilename(module.Info.Name), 
		module.Info.Version)
	return filepath.Join(subdir, filename)

// sanitizeFilename sanitizes a string for use as a filename
func sanitizeFilename(name string) string {
	// Replace spaces with underscores
	result := strings.ReplaceAll(name, " ", "_")
	
	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	var sanitized strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sanitized.WriteRune(r)
		}
	}
	
	// Convert to lowercase
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
