package bundle

import (
	"fmt"
	"regexp"
	"strings"
)

// ExportCustomization provides advanced customization options for bundle export
type ExportCustomization struct {
	// Scope selection
	ScopeOptions       *ScopeOptions              // Business object scope selection
	EnvironmentConfig  *EnvironmentConfig         // Environment-specific configurations
	DependencyHandling *DependencyHandling        // How to handle dependencies
	
	// Filtering options
	TemplateFilters    *TemplateFilterOptions     // Template filtering
	ModuleFilters      *ModuleFilterOptions       // Module filtering
	FileFilters        *FileFilterOptions         // General file filtering
	
	// Transformation options
	Transformations    *TransformationOptions     // Content transformations
	
	// Hotfix generation
	HotfixOptions      *HotfixOptions             // Hotfix script generation
	
	// Export behavior
	BehaviorOptions    *BehaviorOptions           // Export behavior customization
}

// ScopeOptions defines business object scope selection
type ScopeOptions struct {
	IncludeScopes      []string                   // Scopes to include
	ExcludeScopes      []string                   // Scopes to exclude
	ScopeDepth         int                        // Maximum scope traversal depth
	IncludeOrphaned    bool                       // Include orphaned objects
	IncludeDeprecated  bool                       // Include deprecated objects
}

// EnvironmentConfig defines environment-specific configurations
type EnvironmentConfig struct {
	SourceEnvironment  string                     // Source environment name
	TargetEnvironment  string                     // Target environment name
	ConfigOverrides    map[string]interface{}     // Configuration overrides
	SecretHandling     SecretHandlingType         // How to handle secrets
	VariableMapping    map[string]string          // Environment variable mapping
}

// SecretHandlingType defines how secrets are handled during export
type SecretHandlingType string

const (
	SecretExclude      SecretHandlingType = "exclude"      // Exclude secrets
	SecretPlaceholder  SecretHandlingType = "placeholder"  // Replace with placeholders
	SecretEncrypt      SecretHandlingType = "encrypt"      // Encrypt secrets
	SecretInclude      SecretHandlingType = "include"      // Include as-is (dangerous)
)

// DependencyHandling defines how dependencies are handled
type DependencyHandling struct {
	ResolutionStrategy DependencyStrategy         // How to resolve dependencies
	MaxDepth           int                        // Maximum dependency depth
	IncludeOptional    bool                       // Include optional dependencies
	IncludeDevDeps     bool                       // Include development dependencies
	ExcludePatterns    []string                   // Patterns to exclude
	ForceInclude       []string                   // Force include specific dependencies
}

// DependencyStrategy defines dependency resolution strategies
type DependencyStrategy string

const (
	DependencyAll      DependencyStrategy = "all"       // Include all dependencies
	DependencyDirect   DependencyStrategy = "direct"    // Only direct dependencies
	DependencyMinimal  DependencyStrategy = "minimal"   // Minimal required set
	DependencyCustom   DependencyStrategy = "custom"    // Custom strategy
)

// TemplateFilterOptions provides template-specific filtering
type TemplateFilterOptions struct {
	Categories         []string                   // Include only these categories
	ExcludeCategories  []string                   // Exclude these categories
	Tags               []string                   // Include templates with these tags
	ExcludeTags        []string                   // Exclude templates with these tags
	MinVersion         string                     // Minimum template version
	MaxVersion         string                     // Maximum template version
	ModifiedAfter      *time.Time                 // Templates modified after this date
	ModifiedBefore     *time.Time                 // Templates modified before this date
	AuthorFilter       string                     // Filter by author
	CustomFilter       TemplateFilterFunc         // Custom filter function
}

// TemplateFilterFunc is a custom template filter function
type TemplateFilterFunc func(template *TemplateInfo) bool

// TemplateInfo contains template metadata for filtering
type TemplateInfo struct {
	Name               string
	Category           string
	Version            string
	Tags               []string
	Author             string
	Modified           time.Time
	Description        string
	Dependencies       []string
}

// ModuleFilterOptions provides module-specific filtering
type ModuleFilterOptions struct {
	Types              []ModuleType               // Include only these module types
	ExcludeTypes       []ModuleType               // Exclude these module types
	Providers          []string                   // Include modules from these providers
	ExcludeProviders   []string                   // Exclude modules from these providers
	MinVersion         string                     // Minimum module version
	MaxVersion         string                     // Maximum module version
	Platforms          []string                   // Target platforms
	Architectures      []string                   // Target architectures
	CustomFilter       ModuleFilterFunc           // Custom filter function
}

// ModuleType defines types of modules
type ModuleType string

const (
	ModuleProvider     ModuleType = "provider"    // Provider modules
	ModuleDetector     ModuleType = "detector"    // Detector modules
	ModuleProcessor    ModuleType = "processor"   // Processor modules
	ModuleUtility      ModuleType = "utility"     // Utility modules
)

// ModuleFilterFunc is a custom module filter function
type ModuleFilterFunc func(module *ModuleInfo) bool

// ModuleInfo contains module metadata for filtering
type ModuleInfo struct {
	Name               string
	Type               ModuleType
	Version            string
	Provider           string
	Platform           string
	Architecture       string
	Dependencies       []string
}

// FileFilterOptions provides general file filtering
type FileFilterOptions struct {
	IncludePatterns    []string                   // Include files matching these patterns
	ExcludePatterns    []string                   // Exclude files matching these patterns
	MinSize            int64                      // Minimum file size
	MaxSize            int64                      // Maximum file size (0 = no limit)
	ModifiedAfter      *time.Time                 // Files modified after this date
	ModifiedBefore     *time.Time                 // Files modified before this date
	FileTypes          []string                   // Include only these file types
	ExcludeFileTypes   []string                   // Exclude these file types
	CustomFilter       FileFilterFunc             // Custom filter function
}

// FileFilterFunc is a custom file filter function
type FileFilterFunc func(path string, info FileInfo) bool

// FileInfo contains file metadata for filtering
type FileInfo struct {
	Path               string
	Size               int64
	Modified           time.Time
	IsDirectory        bool
	Permissions        uint32
}

// TransformationOptions defines content transformations
type TransformationOptions struct {
	PathTransformations map[string]string         // Path remapping
	ContentTransformers []ContentTransformer      // Content transformation functions
	MetadataEnrichment  MetadataEnricher          // Add metadata to files
	Sanitizers          []ContentSanitizer        // Content sanitization
}

// ContentTransformer transforms file content
type ContentTransformer interface {
	Transform(path string, content []byte) ([]byte, error)
	ShouldTransform(path string) bool
}

// MetadataEnricher adds metadata to files
type MetadataEnricher interface {
	EnrichMetadata(path string, existing map[string]interface{}) map[string]interface{}
}

// ContentSanitizer sanitizes content
type ContentSanitizer interface {
	Sanitize(path string, content []byte) ([]byte, error)
	ShouldSanitize(path string) bool
}

// HotfixOptions defines hotfix script generation options
type HotfixOptions struct {
	GenerateHotfix     bool                       // Generate hotfix scripts
	TargetPlatforms    []string                   // Target platforms for scripts
	ScriptFormat       ScriptFormat               // Script format
	IncludeRollback    bool                       // Include rollback scripts
	TestMode           bool                       // Generate in test mode
	CustomTemplate     string                     // Custom script template
}

// ScriptFormat defines hotfix script formats
type ScriptFormat string

const (
	ScriptBash         ScriptFormat = "bash"       // Bash script
	ScriptPowerShell   ScriptFormat = "powershell" // PowerShell script
	ScriptPython       ScriptFormat = "python"     // Python script
	ScriptCustom       ScriptFormat = "custom"     // Custom format
)

// BehaviorOptions defines export behavior customization
type BehaviorOptions struct {
	ContinueOnError    bool                       // Continue export on non-fatal errors
	ValidateContent    bool                       // Validate content before export
	GenerateChecksums  bool                       // Generate checksums for all files
	CreateBackup       bool                       // Create backup before export
	DryRun             bool                       // Perform dry run only
	Verbose            bool                       // Verbose output
	ParallelExport     bool                       // Use parallel processing
	MaxParallelJobs    int                        // Maximum parallel jobs
}

// CustomizationBuilder provides a fluent API for building customizations
type CustomizationBuilder struct {
	customization *ExportCustomization
}

// NewCustomizationBuilder creates a new customization builder
func NewCustomizationBuilder() *CustomizationBuilder {
	return &CustomizationBuilder{
		customization: &ExportCustomization{
			ScopeOptions:       &ScopeOptions{},
			EnvironmentConfig:  &EnvironmentConfig{},
			DependencyHandling: &DependencyHandling{},
			TemplateFilters:    &TemplateFilterOptions{},
			ModuleFilters:      &ModuleFilterOptions{},
			FileFilters:        &FileFilterOptions{},
			Transformations:    &TransformationOptions{},
			HotfixOptions:      &HotfixOptions{},
			BehaviorOptions:    &BehaviorOptions{},
		},
	}
}

// WithScope configures scope options
func (b *CustomizationBuilder) WithScope(configure func(*ScopeOptions)) *CustomizationBuilder {
	configure(b.customization.ScopeOptions)
	return b
}

// WithEnvironment configures environment options
func (b *CustomizationBuilder) WithEnvironment(configure func(*EnvironmentConfig)) *CustomizationBuilder {
	configure(b.customization.EnvironmentConfig)
	return b
}

// WithDependencies configures dependency handling
func (b *CustomizationBuilder) WithDependencies(configure func(*DependencyHandling)) *CustomizationBuilder {
	configure(b.customization.DependencyHandling)
	return b
}

// WithTemplateFilters configures template filters
func (b *CustomizationBuilder) WithTemplateFilters(configure func(*TemplateFilterOptions)) *CustomizationBuilder {
	configure(b.customization.TemplateFilters)
	return b
}

// WithModuleFilters configures module filters
func (b *CustomizationBuilder) WithModuleFilters(configure func(*ModuleFilterOptions)) *CustomizationBuilder {
	configure(b.customization.ModuleFilters)
	return b
}

// WithFileFilters configures file filters
func (b *CustomizationBuilder) WithFileFilters(configure func(*FileFilterOptions)) *CustomizationBuilder {
	configure(b.customization.FileFilters)
	return b
}

// WithTransformations configures transformations
func (b *CustomizationBuilder) WithTransformations(configure func(*TransformationOptions)) *CustomizationBuilder {
	configure(b.customization.Transformations)
	return b
}

// WithHotfix configures hotfix options
func (b *CustomizationBuilder) WithHotfix(configure func(*HotfixOptions)) *CustomizationBuilder {
	configure(b.customization.HotfixOptions)
	return b
}

// WithBehavior configures behavior options
func (b *CustomizationBuilder) WithBehavior(configure func(*BehaviorOptions)) *CustomizationBuilder {
	configure(b.customization.BehaviorOptions)
	return b
}

// Build returns the configured customization
func (b *CustomizationBuilder) Build() *ExportCustomization {
	return b.customization
}

// Validate validates the customization options
func (c *ExportCustomization) Validate() error {
	// Validate scope options
	if c.ScopeOptions.ScopeDepth < 0 {
		return fmt.Errorf("scope depth must be non-negative")
	}
	
	// Validate environment config
	if c.EnvironmentConfig.SecretHandling == "" {
		c.EnvironmentConfig.SecretHandling = SecretExclude
	}
	
	// Validate dependency handling
	if c.DependencyHandling.MaxDepth < 0 {
		return fmt.Errorf("dependency depth must be non-negative")
	}
	
	// Validate file filters
	if c.FileFilters.MaxSize < 0 {
		return fmt.Errorf("max file size must be non-negative")
	}
	
	// Validate behavior options
	if c.BehaviorOptions.MaxParallelJobs < 0 {
		return fmt.Errorf("max parallel jobs must be non-negative")
	}
	
	return nil
}

// ApplyTemplateFilter applies template filtering based on customization
func (c *ExportCustomization) ApplyTemplateFilter(template *TemplateInfo) bool {
	filters := c.TemplateFilters
	
	// Category filter
	if len(filters.Categories) > 0 && !containsInSlice(filters.Categories, template.Category) {
		return false
	}
	if containsInSlice(filters.ExcludeCategories, template.Category) {
		return false
	}
	
	// Tag filter
	if len(filters.Tags) > 0 && !hasAnyTag(template.Tags, filters.Tags) {
		return false
	}
	if hasAnyTag(template.Tags, filters.ExcludeTags) {
		return false
	}
	
	// Version filter
	if filters.MinVersion != "" && template.Version < filters.MinVersion {
		return false
	}
	if filters.MaxVersion != "" && template.Version > filters.MaxVersion {
		return false
	}
	
	// Date filter
	if filters.ModifiedAfter != nil && template.Modified.Before(*filters.ModifiedAfter) {
		return false
	}
	if filters.ModifiedBefore != nil && template.Modified.After(*filters.ModifiedBefore) {
		return false
	}
	
	// Author filter
	if filters.AuthorFilter != "" && !strings.Contains(template.Author, filters.AuthorFilter) {
		return false
	}
	
	// Custom filter
	if filters.CustomFilter != nil && !filters.CustomFilter(template) {
		return false
	}
	
	return true
}

// ApplyFileFilter applies file filtering based on customization
func (c *ExportCustomization) ApplyFileFilter(path string, info FileInfo) bool {
	filters := c.FileFilters
	
	// Pattern matching
	if len(filters.IncludePatterns) > 0 && !matchesAnyPattern(path, filters.IncludePatterns) {
		return false
	}
	if matchesAnyPattern(path, filters.ExcludePatterns) {
		return false
	}
	
	// Size filter
	if filters.MinSize > 0 && info.Size < filters.MinSize {
		return false
	}
	if filters.MaxSize > 0 && info.Size > filters.MaxSize {
		return false
	}
	
	// Date filter
	if filters.ModifiedAfter != nil && info.Modified.Before(*filters.ModifiedAfter) {
		return false
	}
	if filters.ModifiedBefore != nil && info.Modified.After(*filters.ModifiedBefore) {
		return false
	}
	
	// File type filter
	ext := filepath.Ext(path)
	if len(filters.FileTypes) > 0 && !containsInSlice(filters.FileTypes, ext) {
		return false
	}
	if containsInSlice(filters.ExcludeFileTypes, ext) {
		return false
	}
	
	// Custom filter
	if filters.CustomFilter != nil && !filters.CustomFilter(path, info) {
		return false
	}
	
	return true
}

// Helper functions
func containsInSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasAnyTag(tags []string, searchTags []string) bool {
	for _, tag := range tags {
		if containsInSlice(searchTags, tag) {
			return true
		}
	}
	return false
}

func matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, path)
		if err == nil && matched {
			return true
		}
		// Also try glob matching
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}