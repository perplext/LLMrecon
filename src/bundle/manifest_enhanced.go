package bundle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnhancedManifestGenerator handles comprehensive manifest generation with advanced features
type EnhancedManifestGenerator struct {
	bundlePath      string
	options         ManifestOptions
	contentItems    []*ContentItem
	dependencies    map[string][]string
	crossReferences map[string][]string
	metadata        map[string]interface{}
	errors          []error
}

// ManifestOptions defines options for manifest generation
type ManifestOptions struct {
	IncludeChecksums    bool                   // Include file checksums
	ChecksumAlgorithms  []string               // Checksum algorithms to use
	IncludePermissions  bool                   // Include file permissions
	IncludeTimestamps   bool                   // Include modification timestamps
	IncludeSize         bool                   // Include file sizes
	ResolveDependencies bool                   // Analyze and include dependencies
	IncludeMetadata     bool                   // Extract and include metadata
	CustomFields        map[string]interface{} // Custom fields to add
	PrettyPrint         bool                   // Format output for readability
	Format              ManifestFormat         // Output format
}

// ManifestFormat defines the output format
type ManifestFormat string

const (
	ManifestFormatJSON ManifestFormat = "json"
	ManifestFormatYAML ManifestFormat = "yaml"
	ManifestFormatXML  ManifestFormat = "xml"
)

// DependencyInfo contains dependency information
type DependencyInfo struct {
	Type         string   `json:"type"`         // hard, soft, optional
	Required     bool     `json:"required"`     // Is this dependency required
	Version      string   `json:"version"`      // Version constraint
	Alternatives []string `json:"alternatives"` // Alternative dependencies
}

// MetadataExtractor extracts metadata from different file types
type MetadataExtractor interface {
	Extract(path string) (map[string]interface{}, error)
	Supports(path string) bool
}

// NewEnhancedManifestGenerator creates a new enhanced manifest generator
func NewEnhancedManifestGenerator(bundlePath string, options ManifestOptions) *EnhancedManifestGenerator {
	if options.ChecksumAlgorithms == nil {
		options.ChecksumAlgorithms = []string{"sha256"}
	}
	if options.Format == "" {
		options.Format = ManifestFormatJSON
	}

	return &EnhancedManifestGenerator{
		bundlePath:      bundlePath,
		options:         options,
		contentItems:    []*ContentItem{},
		dependencies:    make(map[string][]string),
		crossReferences: make(map[string][]string),
		metadata:        make(map[string]interface{}),
		errors:          []error{},
	}
}

// Generate creates a comprehensive manifest
func (g *EnhancedManifestGenerator) Generate() (*BundleManifest, error) {
	// Step 1: Scan bundle contents
	if err := g.scanBundleContents(); err != nil {
		return nil, fmt.Errorf("failed to scan bundle contents: %w", err)
	}

	// Step 2: Extract metadata if requested
	if g.options.IncludeMetadata {
		g.extractMetadata()
	}

	// Step 3: Resolve dependencies if requested
	if g.options.ResolveDependencies {
		g.resolveDependencies()
	}

	// Step 4: Build manifest
	manifest := g.buildManifest()

	// Step 5: Validate manifest
	if err := g.validateManifest(manifest); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	return manifest, nil
}

// scanBundleContents scans all files in the bundle
func (g *EnhancedManifestGenerator) scanBundleContents() error {
	return filepath.Walk(g.bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			g.errors = append(g.errors, fmt.Errorf("error accessing %s: %w", path, err))
			return nil // Continue scanning
		}

		// Skip manifest file itself
		if strings.HasSuffix(path, "manifest.json") || strings.HasSuffix(path, "manifest.yaml") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(g.bundlePath, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Create content item
		contentTypeStr := g.determineContentType(relPath, info)
		var contentType ContentType
		switch contentTypeStr {
		case "template":
			contentType = TemplateContentType
		case "module":
			contentType = ModuleContentType
		case "config":
			contentType = ConfigContentType
		case "resource":
			contentType = ResourceContentType
		default:
			contentType = ContentType(contentTypeStr)
		}
		
		item := &ContentItem{
			Path:     relPath,
			Type:     contentType,
			Metadata: make(map[string]interface{}),
		}

		// Calculate checksums if requested
		if g.options.IncludeChecksums && !info.IsDir() {
			checksums, err := g.calculateChecksums(path)
			if err != nil {
				g.errors = append(g.errors, fmt.Errorf("checksum calculation failed for %s: %w", relPath, err))
			} else {
				item.Checksum = checksums["sha256"] // Primary checksum
			}
		}

		// Skip directories for content items
		if !info.IsDir() {
			g.contentItems = append(g.contentItems, item)
		}

		return nil
	})
}

// determineContentType determines the type of content
func (g *EnhancedManifestGenerator) determineContentType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}

	// Check by path patterns
	switch {
	case strings.Contains(path, "templates/"):
		return "template"
	case strings.Contains(path, "modules/"):
		return "module"
	case strings.Contains(path, "binary/"):
		return "binary"
	case strings.Contains(path, "documentation/") || strings.Contains(path, "docs/"):
		return "documentation"
	case strings.Contains(path, "signatures/"):
		return "signature"
	}

	// Check by extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return "template"
	case ".so", ".dll", ".dylib":
		return "module"
	case ".md", ".txt", ".pdf", ".html":
		return "documentation"
	case ".sig", ".asc":
		return "signature"
	case ".json", ".toml", ".ini", ".conf":
		return "configuration"
	default:
		return "file"
	}
}

// calculateChecksums calculates multiple checksums for a file
func (g *EnhancedManifestGenerator) calculateChecksums(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	checksums := make(map[string]string)

	// For now, just implement SHA256
	// TODO: Add support for other algorithms based on options
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}

	checksums["sha256"] = fmt.Sprintf("sha256:%x", h.Sum(nil))

	return checksums, nil
}

// extractMetadata extracts metadata from content files
func (g *EnhancedManifestGenerator) extractMetadata() {
	extractors := []MetadataExtractor{
		&TemplateMetadataExtractor{},
		&ModuleMetadataExtractor{},
		&DocumentationMetadataExtractor{},
	}

	for _, item := range g.contentItems {
		fullPath := filepath.Join(g.bundlePath, item.Path)
		
		for _, extractor := range extractors {
			if extractor.Supports(fullPath) {
				metadata, err := extractor.Extract(fullPath)
				if err != nil {
					g.errors = append(g.errors, fmt.Errorf("metadata extraction failed for %s: %w", item.Path, err))
					continue
				}
				
				// Merge extracted metadata
				if item.Metadata == nil {
					item.Metadata = make(map[string]interface{})
				}
				for k, v := range metadata {
					item.Metadata[k] = v
				}
				
				// Set version if found
				if version, ok := metadata["version"].(string); ok {
					item.Version = version
				}
				
				break
			}
		}
	}
}

// resolveDependencies analyzes and resolves dependencies
func (g *EnhancedManifestGenerator) resolveDependencies() {
	// Build dependency graph
	for _, item := range g.contentItems {
		if item.Type == "template" {
			deps := g.analyzeTemplateDependencies(item)
			if len(deps) > 0 {
				g.dependencies[item.Path] = deps
				if item.Metadata == nil {
					item.Metadata = make(map[string]interface{})
				}
				item.Metadata["dependencies"] = deps
			}
		} else if item.Type == "module" {
			deps := g.analyzeModuleDependencies(item)
			if len(deps) > 0 {
				g.dependencies[item.Path] = deps
				if item.Metadata == nil {
					item.Metadata = make(map[string]interface{})
				}
				item.Metadata["dependencies"] = deps
			}
		}
	}

	// Find cross-references
	g.findCrossReferences()
}

// analyzeTemplateDependencies finds dependencies in templates
func (g *EnhancedManifestGenerator) analyzeTemplateDependencies(item *ContentItem) []string {
	deps := []string{}
	
	// Read template file
	fullPath := filepath.Join(g.bundlePath, item.Path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return deps
	}

	content := string(data)
	
	// Look for references to other templates
	if strings.Contains(content, "template:") || strings.Contains(content, "workflow:") {
		// Simple pattern matching for demonstration
		// TODO: Implement proper YAML parsing
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Contains(line, "template:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					dep := strings.TrimSpace(parts[1])
					if dep != "" {
						deps = append(deps, dep)
					}
				}
			}
		}
	}

	return deps
}

// analyzeModuleDependencies finds module dependencies
func (g *EnhancedManifestGenerator) analyzeModuleDependencies(item *ContentItem) []string {
	// This would analyze binary dependencies, imports, etc.
	// For now, return empty
	return []string{}
}

// findCrossReferences finds references between content items
func (g *EnhancedManifestGenerator) findCrossReferences() {
	// Build a map of content by name/id
	contentMap := make(map[string]*ContentItem)
	for _, item := range g.contentItems {
		contentMap[item.Path] = item
		if id, ok := item.Metadata["id"].(string); ok {
			contentMap[id] = item
		}
	}

	// Find references
	for _, item := range g.contentItems {
		for _, dep := range g.dependencies[item.Path] {
			if referenced, exists := contentMap[dep]; exists {
				g.crossReferences[item.Path] = append(g.crossReferences[item.Path], referenced.Path)
			}
		}
	}
}

// buildManifest builds the final manifest
func (g *EnhancedManifestGenerator) buildManifest() *BundleManifest {
	manifest := &BundleManifest{
		ManifestVersion: "1.0",
		BundleID:        generateBundleID(),
		CreatedAt:       time.Now().UTC(),
		Content:         convertContentItems(g.contentItems),
		Metadata:        g.metadata,
		Checksums: Checksums{
			Content:   make(map[string]string),
		},
	}

	// Add checksums to manifest
	for _, item := range g.contentItems {
		if item.Checksum != "" {
			manifest.Checksums.Content[item.Path] = item.Checksum
		}
	}

	// Add dependency information
	if len(g.dependencies) > 0 {
		manifest.Dependencies = g.dependencies
	}

	// Add cross-references
	if len(g.crossReferences) > 0 {
		manifest.Metadata["crossReferences"] = g.crossReferences
	}

	// Add custom fields
	for k, v := range g.options.CustomFields {
		manifest.Metadata[k] = v
	}

	// Add generation information
	manifest.Metadata["generatedAt"] = time.Now().UTC()
	manifest.Metadata["generatorVersion"] = "1.0.0"
	manifest.Metadata["errors"] = len(g.errors)

	// Sort content items for consistency
	sort.Slice(manifest.Content, func(i, j int) bool {
		return manifest.Content[i].Path < manifest.Content[j].Path
	})

	return manifest
}

// validateManifest validates the generated manifest
func (g *EnhancedManifestGenerator) validateManifest(manifest *BundleManifest) error {
	// Check required fields
	if manifest.ManifestVersion == "" {
		return fmt.Errorf("manifest version is required")
	}

	if manifest.BundleID == "" {
		return fmt.Errorf("bundle ID is required")
	}

	if len(manifest.Content) == 0 {
		return fmt.Errorf("manifest has no content items")
	}

	// Validate content items
	for _, item := range manifest.Content {
		if item.Path == "" {
			return fmt.Errorf("content item missing path")
		}
		if item.Type == "" {
			return fmt.Errorf("content item %s missing type", item.Path)
		}
	}

	// Check for dependency cycles
	if err := g.checkDependencyCycles(); err != nil {
		return err
	}

	return nil
}

// checkDependencyCycles checks for circular dependencies
func (g *EnhancedManifestGenerator) checkDependencyCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range g.dependencies[node] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range g.dependencies {
		if !visited[node] {
			if hasCycle(node) {
				return fmt.Errorf("circular dependency detected involving %s", node)
			}
		}
	}

	return nil
}

// WriteManifest writes the manifest to a file
func (g *EnhancedManifestGenerator) WriteManifest(manifest *BundleManifest, outputPath string) error {
	var data []byte
	var err error

	switch g.options.Format {
	case ManifestFormatJSON:
		if g.options.PrettyPrint {
			data, err = json.MarshalIndent(manifest, "", "  ")
		} else {
			data, err = json.Marshal(manifest)
		}
	case ManifestFormatYAML:
		data, err = yaml.Marshal(manifest)
	default:
		return fmt.Errorf("unsupported manifest format: %s", g.options.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return os.WriteFile(outputPath, data, 0644)
}

// GetErrors returns any errors encountered during generation
func (g *EnhancedManifestGenerator) GetErrors() []error {
	return g.errors
}

// Metadata Extractors

// TemplateMetadataExtractor extracts metadata from template files
type TemplateMetadataExtractor struct{}

func (e *TemplateMetadataExtractor) Supports(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

func (e *TemplateMetadataExtractor) Extract(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		// Try to extract basic fields even if full parsing fails
		metadata = make(map[string]interface{})
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "id:") {
				metadata["id"] = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			} else if strings.HasPrefix(line, "name:") {
				metadata["name"] = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			} else if strings.HasPrefix(line, "version:") {
				metadata["version"] = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			} else if strings.HasPrefix(line, "category:") {
				metadata["category"] = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
			}
		}
	}

	return metadata, nil
}

// ModuleMetadataExtractor extracts metadata from module files
type ModuleMetadataExtractor struct{}

func (e *ModuleMetadataExtractor) Supports(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".so" || ext == ".dll" || ext == ".dylib"
}

func (e *ModuleMetadataExtractor) Extract(path string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})
	
	// Check for companion metadata file
	metadataPath := path + ".metadata.json"
	if data, err := os.ReadFile(metadataPath); err == nil {
		json.Unmarshal(data, &metadata)
	}

	// Extract from filename
	base := filepath.Base(path)
	metadata["filename"] = base
	
	// Extract module type from path
	if strings.Contains(path, "providers") {
		metadata["moduleType"] = "provider"
	} else if strings.Contains(path, "detectors") {
		metadata["moduleType"] = "detector"
	}

	return metadata, nil
}

// DocumentationMetadataExtractor extracts metadata from documentation files
type DocumentationMetadataExtractor struct{}

func (e *DocumentationMetadataExtractor) Supports(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".txt" || ext == ".pdf" || ext == ".html"
}

func (e *DocumentationMetadataExtractor) Extract(path string) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})
	
	// Extract title from markdown files
	if strings.HasSuffix(path, ".md") {
		data, err := os.ReadFile(path)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "# ") {
					metadata["title"] = strings.TrimPrefix(line, "# ")
					break
				}
			}
		}
	}

	metadata["docType"] = strings.TrimPrefix(filepath.Ext(path), ".")
	
	return metadata, nil
}

// convertContentItems converts []*ContentItem to []ContentItem
func convertContentItems(items []*ContentItem) []ContentItem {
	result := make([]ContentItem, len(items))
	for i, item := range items {
		if item != nil {
			result[i] = *item
		}
	}
	return result
}


// GenerateComparisonReport generates a comparison between two manifests
func GenerateComparisonReport(oldManifest, newManifest *BundleManifest) *ManifestComparison {
	comparison := &ManifestComparison{
		OldVersion: oldManifest.BundleVersion,
		NewVersion: newManifest.BundleVersion,
		Changes:    ManifestChanges{},
		Summary:    ComparisonSummary{},
	}

	// Build maps for comparison
	oldContent := make(map[string]*ContentItem)
	for i := range oldManifest.Content {
		oldContent[oldManifest.Content[i].Path] = &oldManifest.Content[i]
	}

	newContent := make(map[string]*ContentItem)
	for i := range newManifest.Content {
		newContent[newManifest.Content[i].Path] = &newManifest.Content[i]
	}

	// Find additions
	for path, newItem := range newContent {
		if _, exists := oldContent[path]; !exists {
			comparison.Changes.Added = append(comparison.Changes.Added, ContentChange{
				Path: path,
				Type: newItem.Type,
				Size: newItem.Size,
			})
			comparison.Summary.AddedCount++
		}
	}

	// Find deletions and modifications
	for path, oldItem := range oldContent {
		if newItem, exists := newContent[path]; !exists {
			comparison.Changes.Removed = append(comparison.Changes.Removed, ContentChange{
				Path: path,
				Type: oldItem.Type,
				Size: oldItem.Size,
			})
			comparison.Summary.RemovedCount++
		} else {
			// Check for modifications
			if oldItem.Checksum != newItem.Checksum {
				comparison.Changes.Modified = append(comparison.Changes.Modified, ContentChange{
					Path:        path,
					Type:        newItem.Type,
					OldChecksum: oldItem.Checksum,
					NewChecksum: newItem.Checksum,
					SizeDelta:   newItem.Size - oldItem.Size,
				})
				comparison.Summary.ModifiedCount++
			}
		}
	}

	comparison.Summary.TotalChanges = comparison.Summary.AddedCount + 
		comparison.Summary.RemovedCount + comparison.Summary.ModifiedCount

	return comparison
}

// Types for manifest comparison

// ManifestComparison contains the comparison results
type ManifestComparison struct {
	OldVersion string             `json:"oldVersion"`
	NewVersion string             `json:"newVersion"`
	Changes    ManifestChanges    `json:"changes"`
	Summary    ComparisonSummary  `json:"summary"`
}

// ManifestChanges contains lists of changes
type ManifestChanges struct {
	Added    []ContentChange `json:"added"`
	Removed  []ContentChange `json:"removed"`
	Modified []ContentChange `json:"modified"`
}

// ContentChange represents a change to content
type ContentChange struct {
	Path        string      `json:"path"`
	Type        ContentType `json:"type"`
	Size        int64       `json:"size,omitempty"`
	OldChecksum string      `json:"oldChecksum,omitempty"`
	NewChecksum string      `json:"newChecksum,omitempty"`
	SizeDelta   int64       `json:"sizeDelta,omitempty"`
}

// ComparisonSummary provides a summary of changes
type ComparisonSummary struct {
	TotalChanges  int `json:"totalChanges"`
	AddedCount    int `json:"addedCount"`
	RemovedCount  int `json:"removedCount"`
	ModifiedCount int `json:"modifiedCount"`
}