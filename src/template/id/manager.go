package id

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// TemplateInfo represents the basic information about a template
type TemplateInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Version     string   `json:"version"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
}

}
// Manager handles the management of template and module IDs
type Manager struct {
	generator      *Generator
	templatesDir   string
	modulesDir     string
	templateIDs    map[string]TemplateInfo
	categoryIDs    map[string]bool
	providerIDs    map[string]bool
	utilityIDs     map[string]bool
	detectorIDs    map[string]bool

}
// NewManager creates a new ID manager
func NewManager(templatesDir, modulesDir string) *Manager {
	return &Manager{
		generator:    NewGenerator(),
		templatesDir: templatesDir,
		modulesDir:   modulesDir,
		templateIDs:  make(map[string]TemplateInfo),
		categoryIDs:  make(map[string]bool),
		providerIDs:  make(map[string]bool),
		utilityIDs:   make(map[string]bool),
		detectorIDs:  make(map[string]bool),
	}

// LoadExistingIDs loads existing IDs from the manifest files
}
func (m *Manager) LoadExistingIDs() error {
	// Load template IDs
	if err := m.loadTemplateIDs(); err != nil {
		return err
	}
	
	// Load module IDs
	if err := m.loadModuleIDs(); err != nil {
		return err
	}
	
	return nil

// loadTemplateIDs loads template and category IDs from the template manifest
}
func (m *Manager) loadTemplateIDs() error {
	manifestPath := filepath.Join(m.templatesDir, "manifest.json")
	
	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// No manifest yet, nothing to load
		return nil
	}
	
	// Read and parse manifest
	data, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return fmt.Errorf("failed to read template manifest: %w", err)
	}
	
	var manifest struct {
		Templates  []TemplateInfo `json:"templates"`
		Categories []struct {
			ID string `json:"id"`
		} `json:"categories"`
	}
	
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse template manifest: %w", err)
	}
	
	// Register template IDs
	for _, template := range manifest.Templates {
		m.templateIDs[template.ID] = template
		m.generator.RegisterExistingID(template.ID)
	}
	
	// Register category IDs
	for _, category := range manifest.Categories {
		m.categoryIDs[category.ID] = true
		m.generator.RegisterExistingID(category.ID)
	}
	
	return nil

// loadModuleIDs loads module IDs from the module manifest
}
func (m *Manager) loadModuleIDs() error {
	manifestPath := filepath.Join(m.modulesDir, "manifest.json")
	
	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// No manifest yet, nothing to load
		return nil
	}
	
	// Read and parse manifest
	data, err := ioutil.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		return fmt.Errorf("failed to read module manifest: %w", err)
	}
	
	var manifest struct {
		Providers []struct {
			ID string `json:"id"`
		} `json:"providers"`
		Utilities []struct {
			ID string `json:"id"`
		} `json:"utilities"`
		Detectors []struct {
			ID string `json:"id"`
		} `json:"detectors"`
	}
	
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse module manifest: %w", err)
	}
	
	// Register provider IDs
	for _, provider := range manifest.Providers {
		m.providerIDs[provider.ID] = true
		m.generator.RegisterExistingID(provider.ID)
	}
	
	// Register utility IDs
	for _, utility := range manifest.Utilities {
		m.utilityIDs[utility.ID] = true
		m.generator.RegisterExistingID(utility.ID)
	}
	
	// Register detector IDs
	for _, detector := range manifest.Detectors {
		m.detectorIDs[detector.ID] = true
		m.generator.RegisterExistingID(detector.ID)
	}
	
	return nil

// GenerateTemplateID generates a unique ID for a template
}
func (m *Manager) GenerateTemplateID(category, name, version string) (string, error) {
	// Validate category
	if !m.categoryIDs[category] {
		return "", fmt.Errorf("unknown category: %s", category)
	}
	
	// Generate ID
	id, err := m.generator.GenerateID(TemplateID, category, name, version)
	if err != nil {
		return "", err
	}
	
	return id, nil

// GenerateCategoryID generates a unique ID for a category
}
func (m *Manager) GenerateCategoryID(name string) (string, error) {
	// Generate ID
	id, err := m.generator.GenerateID(CategoryID, "", name, "")
	if err != nil {
		return "", err
	}
	
	return id, nil

// GenerateProviderID generates a unique ID for a provider
}
func (m *Manager) GenerateProviderID(name string) (string, error) {
	// Generate ID
	id, err := m.generator.GenerateID(ProviderID, "", name, "")
	if err != nil {
		return "", err
	}
	
	return id, nil

// GenerateUtilityID generates a unique ID for a utility
}
func (m *Manager) GenerateUtilityID(name string) (string, error) {
	// Generate ID
	id, err := m.generator.GenerateID(UtilityID, "", name, "")
	if err != nil {
		return "", err
	}
	
	return id, nil

// GenerateDetectorID generates a unique ID for a detector
}
func (m *Manager) GenerateDetectorID(name string) (string, error) {
	// Generate ID
	id, err := m.generator.GenerateID(DetectorID, "", name, "")
	if err != nil {
		return "", err
	}
	
	return id, nil

// IsTemplateIDTaken checks if a template ID is already in use
}
func (m *Manager) IsTemplateIDTaken(id string) bool {
	_, exists := m.templateIDs[id]
	return exists

// IsCategoryIDTaken checks if a category ID is already in use
}
func (m *Manager) IsCategoryIDTaken(id string) bool {
	return m.categoryIDs[id]

// IsProviderIDTaken checks if a provider ID is already in use
}
func (m *Manager) IsProviderIDTaken(id string) bool {
	return m.providerIDs[id]

// IsUtilityIDTaken checks if a utility ID is already in use
}
func (m *Manager) IsUtilityIDTaken(id string) bool {
	return m.utilityIDs[id]

// IsDetectorIDTaken checks if a detector ID is already in use
}
func (m *Manager) IsDetectorIDTaken(id string) bool {
	return m.detectorIDs[id]

// GetTemplateInfo gets information about a template by ID
}
func (m *Manager) GetTemplateInfo(id string) (TemplateInfo, bool) {
	info, exists := m.templateIDs[id]
	return info, exists

// GetTemplatesByCategory gets all templates in a category
}
func (m *Manager) GetTemplatesByCategory(category string) []TemplateInfo {
	var templates []TemplateInfo
	
	for _, template := range m.templateIDs {
		if template.Category == category {
			templates = append(templates, template)
		}
	}
	
	return templates
	

// GetTemplatesByTag gets all templates with a specific tag
}
func (m *Manager) GetTemplatesByTag(tag string) []TemplateInfo {
	var templates []TemplateInfo
	
	for _, template := range m.templateIDs {
		for _, t := range template.Tags {
			if t == tag {
				templates = append(templates, template)
				break
			}
		}
	}
	
	return templates
// RegisterTemplate registers a new template
}
func (m *Manager) RegisterTemplate(info TemplateInfo) error {
	// Validate template ID
	if err := m.generator.ValidateID(TemplateID, info.ID); err != nil {
		return err
	}
	
	// Check if ID is already taken
	if m.IsTemplateIDTaken(info.ID) {
		return fmt.Errorf("template ID '%s' is already in use", info.ID)
	}
	
	// Register template
	m.templateIDs[info.ID] = info
	m.generator.RegisterExistingID(info.ID)
	
	return nil

// RegisterCategory registers a new category
}
func (m *Manager) RegisterCategory(id string) error {
	// Validate category ID
	if err := m.generator.ValidateID(CategoryID, id); err != nil {
		return err
	}
	
	// Check if ID is already taken
	if m.IsCategoryIDTaken(id) {
		return fmt.Errorf("category ID '%s' is already in use", id)
	}
	
	// Register category
	m.categoryIDs[id] = true
	m.generator.RegisterExistingID(id)
	
	return nil

// RegisterProvider registers a new provider
}
func (m *Manager) RegisterProvider(id string) error {
	// Validate provider ID
	if err := m.generator.ValidateID(ProviderID, id); err != nil {
		return err
	}
	
	// Check if ID is already taken
	if m.IsProviderIDTaken(id) {
		return fmt.Errorf("provider ID '%s' is already in use", id)
	}
	
	// Register provider
	m.providerIDs[id] = true
	m.generator.RegisterExistingID(id)
	
	return nil

// RegisterUtility registers a new utility
}
func (m *Manager) RegisterUtility(id string) error {
	// Validate utility ID
	if err := m.generator.ValidateID(UtilityID, id); err != nil {
		return err
	}
	
	// Check if ID is already taken
	if m.IsUtilityIDTaken(id) {
		return fmt.Errorf("utility ID '%s' is already in use", id)
	}
	
	// Register utility
	m.utilityIDs[id] = true
	m.generator.RegisterExistingID(id)
	
	return nil

// RegisterDetector registers a new detector
}
func (m *Manager) RegisterDetector(id string) error {
	// Validate detector ID
	if err := m.generator.ValidateID(DetectorID, id); err != nil {
		return err
	}
	
	// Check if ID is already taken
	if m.IsDetectorIDTaken(id) {
		return fmt.Errorf("detector ID '%s' is already in use", id)
	}
	
	// Register detector
	m.detectorIDs[id] = true
	m.generator.RegisterExistingID(id)
	
	return nil

// GetTemplateFilename generates a filename for a template
}
func (m *Manager) GetTemplateFilename(name, version string) string {
	// Convert name to snake_case
	filename := sanitizeForID(name)
	
	// Add version
	filename = fmt.Sprintf("%s_v%s.yaml", filename, version)
	
	return filename

// GetModuleFilename generates a filename for a module
}
func (m *Manager) GetModuleFilename(name, version string) string {
	// Convert name to snake_case
	filename := sanitizeForID(name)
	
	// Add version
	filename = fmt.Sprintf("%s_v%s.go", filename, version)
	
