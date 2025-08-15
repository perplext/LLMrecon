package bundle

import (
	"encoding/csv"
	"fmt"
	"strings"
)

// PartialExportOptions defines options for selective export
type PartialExportOptions struct {
	// Entity selection methods
	EntitySearch      *EntitySearchCriteria  // Search-based selection
	EntityCSV         string                 // CSV file with entity list
	EntityList        []string               // Direct list of entities
	RevisionControl   *RevisionControlConfig // Git/SVN integration
	
	// Export scope
	ExportScope       ExportScope            // What to export
	IncludeDependencies bool                 // Include entity dependencies
	ResolveReferences bool                   // Resolve cross-references
	
	// Selection state
	PendingEntities   []EntityInfo           // Entities pending addition
	IncludedEntities  map[string]EntityInfo  // Currently included entities
	ExcludedEntities  map[string]EntityInfo  // Explicitly excluded entities

// EntitySearchCriteria defines search parameters for entities
type EntitySearchCriteria struct {
	Query           string            // Search query
	Type            []string          // Entity types to search
	Category        []string          // Categories to include
	Tags            []string          // Required tags
	Author          string            // Filter by author
	DateRange       *DateRange        // Created/modified date range
	VersionRange    *VersionRange     // Version constraints
	CustomFilters   map[string]string // Additional filters

// RevisionControlConfig defines VCS integration settings
type RevisionControlConfig struct {
	Type            string    // git, svn, mercurial
	Repository      string    // Repository URL
	Branch          string    // Branch/tag to use
	CommitRange     string    // Commit range (e.g., HEAD~10..HEAD)
	IncludeModified bool      // Include uncommitted changes
	AuthConfig      *AuthInfo // Authentication details

// ExportScope defines what to include in partial export
type ExportScope string

const (
	ScopeTemplatesOnly   ExportScope = "templates"
	ScopeModulesOnly     ExportScope = "modules"
	ScopeDocsOnly        ExportScope = "documentation"
	ScopeCustom          ExportScope = "custom"
	ScopeModifiedOnly    ExportScope = "modified"
	ScopeDependencies    ExportScope = "dependencies"
)

// EntityInfo represents an entity in the bundle
type EntityInfo struct {
	ID              string                 `json:"id"`
	Path            string                 `json:"path"`
	Type            string                 `json:"type"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Category        string                 `json:"category"`
	Tags            []string               `json:"tags"`
	Dependencies    []string               `json:"dependencies"`
	Size            int64                  `json:"size"`
	Modified        time.Time              `json:"modified"`
	Author          string                 `json:"author"`
	Metadata        map[string]interface{} `json:"metadata"`
	Selected        bool                   `json:"selected"`
	SelectionReason string                 `json:"selectionReason"`

// PartialBundleExporter extends BundleExporter for selective exports
type PartialBundleExporter struct {
	*BundleExporter
	partialOptions *PartialExportOptions
	entityIndex    map[string]EntityInfo
	dependencies   map[string][]string

// NewPartialBundleExporter creates a new partial bundle exporter
func NewPartialBundleExporter(options ExportOptions, partialOptions *PartialExportOptions) *PartialBundleExporter {
	return &PartialBundleExporter{
		BundleExporter: NewBundleExporter(options),
		partialOptions: partialOptions,
		entityIndex:    make(map[string]EntityInfo),
		dependencies:   make(map[string][]string),
	}

// SelectEntities performs entity selection based on criteria
func (e *PartialBundleExporter) SelectEntities() error {
	// Initialize included entities map
	if e.partialOptions.IncludedEntities == nil {
		e.partialOptions.IncludedEntities = make(map[string]EntityInfo)
	}

	// Build entity index
	if err := e.buildEntityIndex(); err != nil {
		return fmt.Errorf("failed to build entity index: %w", err)
	}

	// Apply selection methods
	var selectedEntities []EntityInfo

	// 1. Search-based selection
	if e.partialOptions.EntitySearch != nil {
		entities, err := e.searchEntities(e.partialOptions.EntitySearch)
		if err != nil {
			return fmt.Errorf("entity search failed: %w", err)
		}
		selectedEntities = append(selectedEntities, entities...)
	}
	// 2. CSV-based selection
	if e.partialOptions.EntityCSV != "" {
		entities, err := e.loadEntitiesFromCSV(e.partialOptions.EntityCSV)
		if err != nil {
			return fmt.Errorf("failed to load entities from CSV: %w", err)
		}
		selectedEntities = append(selectedEntities, entities...)
	}

	// 3. Direct entity list
	for _, entityID := range e.partialOptions.EntityList {
		if entity, exists := e.entityIndex[entityID]; exists {
			entity.SelectionReason = "direct selection"
			selectedEntities = append(selectedEntities, entity)
		}
	}

	// 4. Revision control integration
	if e.partialOptions.RevisionControl != nil {
		entities, err := e.selectFromRevisionControl(e.partialOptions.RevisionControl)
		if err != nil {
			return fmt.Errorf("revision control selection failed: %w", err)
		}
		selectedEntities = append(selectedEntities, entities...)
	}

	// 5. Scope-based selection
	if e.partialOptions.ExportScope != "" {
		entities, err := e.selectByScope(e.partialOptions.ExportScope)
		if err != nil {
			return fmt.Errorf("scope selection failed: %w", err)
		}
		selectedEntities = append(selectedEntities, entities...)
	}

	// Add selected entities to included list
	for _, entity := range selectedEntities {
		if _, excluded := e.partialOptions.ExcludedEntities[entity.ID]; !excluded {
			entity.Selected = true
			e.partialOptions.IncludedEntities[entity.ID] = entity
		}
	}

	// Resolve dependencies if requested
	if e.partialOptions.IncludeDependencies {
		if err := e.resolveDependencies(); err != nil {
			return fmt.Errorf("dependency resolution failed: %w", err)
		}
	}

	// Update pending entities list
	e.updatePendingEntities()

	return nil

// AddEntity adds a single entity to the export
func (e *PartialBundleExporter) AddEntity(entityID string) error {
	entity, exists := e.entityIndex[entityID]
	if !exists {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	entity.Selected = true
	entity.SelectionReason = "manual addition"
	e.partialOptions.IncludedEntities[entityID] = entity

	// Check dependencies
	if e.partialOptions.IncludeDependencies {
		for _, depID := range entity.Dependencies {
			if _, included := e.partialOptions.IncludedEntities[depID]; !included {
				if dep, exists := e.entityIndex[depID]; exists {
					dep.Selected = true
					dep.SelectionReason = fmt.Sprintf("dependency of %s", entityID)
					e.partialOptions.IncludedEntities[depID] = dep
				}
			}
		}
	}

	e.updatePendingEntities()
	return nil

// RemoveEntity removes an entity from the export
func (e *PartialBundleExporter) RemoveEntity(entityID string) error {
	entity, exists := e.partialOptions.IncludedEntities[entityID]
	if !exists {
		return fmt.Errorf("entity not in export: %s", entityID)
	}

	// Remove from included list
	delete(e.partialOptions.IncludedEntities, entityID)

	// Add to excluded list to prevent re-addition
	if e.partialOptions.ExcludedEntities == nil {
		e.partialOptions.ExcludedEntities = make(map[string]EntityInfo)
	}
	entity.Selected = false
	e.partialOptions.ExcludedEntities[entityID] = entity

	// Check for orphaned dependencies
	e.checkOrphanedDependencies()
	e.updatePendingEntities()

	return nil

// GetExportPreview returns a preview of what will be exported
func (e *PartialBundleExporter) GetExportPreview() ExportPreview {
	preview := ExportPreview{
		TotalEntities:    len(e.partialOptions.IncludedEntities),
		TotalSize:        0,
		EntityBreakdown:  make(map[string]int),
		DependencyGraph:  make(map[string][]string),
	}

	// Calculate totals
	for _, entity := range e.partialOptions.IncludedEntities {
		preview.TotalSize += entity.Size
		preview.EntityBreakdown[entity.Type]++
		if len(entity.Dependencies) > 0 {
			preview.DependencyGraph[entity.ID] = entity.Dependencies
		}
	}

	// Add entity list
	for _, entity := range e.partialOptions.IncludedEntities {
		preview.Entities = append(preview.Entities, EntitySummary{
			ID:       entity.ID,
			Name:     entity.Name,
			Type:     entity.Type,
			Size:     entity.Size,
			Reason:   entity.SelectionReason,
			Selected: entity.Selected,
		})
	}

	return preview

// Export performs the partial bundle export
func (e *PartialBundleExporter) Export() error {
	// Override the parent's collectComponents method
	e.BundleExporter.options.Filters = &ExportFilters{
		// Convert included entities to filter criteria
		IncludeList: e.getIncludedPaths(),
		ExcludeList: e.getExcludedPaths(),
	}

	// Perform standard export with filters
	return e.BundleExporter.Export()
	

// Private methods

func (e *PartialBundleExporter) buildEntityIndex() error {
	// Index templates
	templateDir := getTemplateDirectory()
	if err := e.indexDirectory(templateDir, "template"); err != nil {
		return err
	}

	// Index modules
	moduleDir := getModuleDirectory()
	if err := e.indexDirectory(moduleDir, "module"); err != nil {
		return err
	}

	// Index documentation
	docDir := getDocumentationDirectory()
	if err := e.indexDirectory(docDir, "documentation"); err != nil {
		return err
	}

	return nil

func (e *PartialBundleExporter) indexDirectory(dir string, entityType string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(dir, path)
			entityID := generateEntityID(entityType, relPath)
			
			entity := EntityInfo{
				ID:       entityID,
				Path:     path,
				Type:     entityType,
				Name:     filepath.Base(path),
				Size:     info.Size(),
				Modified: info.ModTime(),
			}

			// Extract metadata based on type
			switch entityType {
			case "template":
				e.extractTemplateMetadata(&entity)
			case "module":
				e.extractModuleMetadata(&entity)
			}

			e.entityIndex[entityID] = entity
		}

		return nil
	})

func (e *PartialBundleExporter) searchEntities(criteria *EntitySearchCriteria) ([]EntityInfo, error) {
	var results []EntityInfo

	for _, entity := range e.entityIndex {
		if e.matchesSearchCriteria(entity, criteria) {
			entity.SelectionReason = "search match"
			results = append(results, entity)
		}
	}

	return results, nil

func (e *PartialBundleExporter) matchesSearchCriteria(entity EntityInfo, criteria *EntitySearchCriteria) bool {
	// Check type filter
	if len(criteria.Type) > 0 {
		found := false
		for _, t := range criteria.Type {
			if entity.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check category filter
	if len(criteria.Category) > 0 {
		found := false
		for _, cat := range criteria.Category {
			if entity.Category == cat {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check query match
	if criteria.Query != "" {
		query := strings.ToLower(criteria.Query)
		if !strings.Contains(strings.ToLower(entity.Name), query) &&
		   !strings.Contains(strings.ToLower(entity.Path), query) {
			return false
		}
	}

	// Check date range
	if criteria.DateRange != nil {
		if entity.Modified.Before(criteria.DateRange.Start) ||
		   entity.Modified.After(criteria.DateRange.End) {
			return false
		}
	}

	return true

func (e *PartialBundleExporter) loadEntitiesFromCSV(csvPath string) ([]EntityInfo, error) {
	file, err := os.Open(filepath.Clean(csvPath))
	if err != nil {
		return nil, err
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var entities []EntityInfo
	for i, record := range records {
		if i == 0 && strings.ToLower(record[0]) == "id" {
			// Skip header
			continue
		}

		if len(record) > 0 {
			entityID := record[0]
			if entity, exists := e.entityIndex[entityID]; exists {
				entity.SelectionReason = "CSV import"
				entities = append(entities, entity)
			}
		}
	}

	return entities, nil

func (e *PartialBundleExporter) selectFromRevisionControl(config *RevisionControlConfig) ([]EntityInfo, error) {
	// This would integrate with git/svn to find modified files
	// For now, return a stub implementation
	return []EntityInfo{}, nil

func (e *PartialBundleExporter) selectByScope(scope ExportScope) ([]EntityInfo, error) {
	var entities []EntityInfo

	for _, entity := range e.entityIndex {
		include := false
		reason := ""

		switch scope {
		case ScopeTemplatesOnly:
			include = entity.Type == "template"
			reason = "scope: templates only"
		case ScopeModulesOnly:
			include = entity.Type == "module"
			reason = "scope: modules only"
		case ScopeDocsOnly:
			include = entity.Type == "documentation"
			reason = "scope: documentation only"
		case ScopeModifiedOnly:
			// Check if modified recently (e.g., last 7 days)
			include = time.Since(entity.Modified) < 7*24*time.Hour
			reason = "scope: recently modified"
		}

		if include {
			entity.SelectionReason = reason
			entities = append(entities, entity)
		}
	}

	return entities, nil

func (e *PartialBundleExporter) resolveDependencies() error {
	// Build dependency graph
	for _, entity := range e.partialOptions.IncludedEntities {
		for _, depID := range entity.Dependencies {
			if _, included := e.partialOptions.IncludedEntities[depID]; !included {
				if dep, exists := e.entityIndex[depID]; exists {
					dep.Selected = true
					dep.SelectionReason = fmt.Sprintf("dependency of %s", entity.Name)
					e.partialOptions.IncludedEntities[depID] = dep
				}
			}
		}
	}

	return nil

func (e *PartialBundleExporter) checkOrphanedDependencies() {
	// Remove dependencies that are no longer needed
	for id, entity := range e.partialOptions.IncludedEntities {
		if strings.Contains(entity.SelectionReason, "dependency of") {
			// Check if the parent is still included
			stillNeeded := false
			for _, other := range e.partialOptions.IncludedEntities {
				for _, depID := range other.Dependencies {
					if depID == id {
						stillNeeded = true
						break
					}
				}
				if stillNeeded {
					break
				}
			}
			
			if !stillNeeded {
				delete(e.partialOptions.IncludedEntities, id)
			}
		}
	}

func (e *PartialBundleExporter) updatePendingEntities() {
	e.partialOptions.PendingEntities = []EntityInfo{}
	
	// Find entities that might be added (e.g., unresolved dependencies)
	for _, entity := range e.partialOptions.IncludedEntities {
		for _, depID := range entity.Dependencies {
			if _, included := e.partialOptions.IncludedEntities[depID]; !included {
				if dep, exists := e.entityIndex[depID]; exists {
					dep.SelectionReason = fmt.Sprintf("suggested: dependency of %s", entity.Name)
					e.partialOptions.PendingEntities = append(e.partialOptions.PendingEntities, dep)
				}
			}
		}
	}

func (e *PartialBundleExporter) getIncludedPaths() []string {
	var paths []string
	for _, entity := range e.partialOptions.IncludedEntities {
		paths = append(paths, entity.Path)
	}
	return paths

func (e *PartialBundleExporter) getExcludedPaths() []string {
	var paths []string
	for _, entity := range e.partialOptions.ExcludedEntities {
		paths = append(paths, entity.Path)
	}
	return paths

func (e *PartialBundleExporter) extractTemplateMetadata(entity *EntityInfo) {
	// Read template file and extract metadata
	data, err := os.ReadFile(filepath.Clean(entity.Path))
	if err != nil {
		return
	}

	// Simple YAML parsing for demonstration
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "category:") {
			entity.Category = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
		} else if strings.HasPrefix(line, "tags:") {
			// Parse tags (simplified)
			entity.Tags = strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "tags:")), ",")
		} else if strings.HasPrefix(line, "version:") {
			entity.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		}
	}

func (e *PartialBundleExporter) extractModuleMetadata(entity *EntityInfo) {
	// Extract module metadata based on file location or content
	if strings.Contains(entity.Path, "providers") {
		entity.Category = "provider"
	} else if strings.Contains(entity.Path, "detectors") {
		entity.Category = "detector"
	}

// Helper types

// ExportPreview provides a preview of the export
type ExportPreview struct {
	TotalEntities   int                  `json:"totalEntities"`
	TotalSize       int64                `json:"totalSize"`
	EntityBreakdown map[string]int       `json:"entityBreakdown"`
	Entities        []EntitySummary      `json:"entities"`
	DependencyGraph map[string][]string  `json:"dependencyGraph"`
}

// EntitySummary provides a summary of an entity
type EntitySummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Reason   string `json:"reason"`
	Selected bool   `json:"selected"`

// DateRange defines a date range
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`

// VersionRange defines a version range constraint
type VersionRange struct {
	Min string `json:"min"`
	Max string `json:"max"`

// AuthInfo contains authentication details
type AuthInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	KeyFile  string `json:"keyFile"`

func generateEntityID(entityType, path string) string {
	// Generate a unique ID for an entity
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
}
}
