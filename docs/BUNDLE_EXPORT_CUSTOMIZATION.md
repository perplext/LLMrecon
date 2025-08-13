# Bundle Export Customization System

## Overview

The Bundle Export Customization System provides comprehensive control over the bundle export process, allowing fine-grained configuration of what gets exported, how it's transformed, and how the export behaves. This system is designed for enterprise environments where exports need to be tailored for different deployment scenarios.

## Features

- **Scope Selection**: Control which business objects and components are included
- **Environment Configuration**: Adapt exports for different target environments
- **Dependency Management**: Smart dependency resolution with customizable strategies
- **Advanced Filtering**: Fine-grained control over templates, modules, and files
- **Content Transformation**: Transform content during export for environment adaptation
- **Hotfix Generation**: Automatically generate deployment scripts
- **Flexible Behavior**: Control export behavior with parallelization and error handling

## Architecture

### Core Components

1. **ExportCustomization**: Main configuration container
2. **CustomizationBuilder**: Fluent API for building configurations
3. **Filter Systems**: Template, module, and file filtering
4. **Transformers**: Content transformation pipeline
5. **Validators**: Configuration validation

### Configuration Categories

```go
type ExportCustomization struct {
    ScopeOptions       *ScopeOptions
    EnvironmentConfig  *EnvironmentConfig
    DependencyHandling *DependencyHandling
    TemplateFilters    *TemplateFilterOptions
    ModuleFilters      *ModuleFilterOptions
    FileFilters        *FileFilterOptions
    Transformations    *TransformationOptions
    HotfixOptions      *HotfixOptions
    BehaviorOptions    *BehaviorOptions
}
```

## Usage Guide

### Basic Customization

```go
// Create a simple customization
customization := bundle.NewCustomizationBuilder().
    WithScope(func(s *bundle.ScopeOptions) {
        s.IncludeScopes = []string{"templates", "modules"}
        s.ScopeDepth = 2
    }).
    WithBehavior(func(b *bundle.BehaviorOptions) {
        b.ValidateContent = true
        b.GenerateChecksums = true
    }).
    Build()
```

### Environment-Specific Export

```go
customization := bundle.NewCustomizationBuilder().
    WithEnvironment(func(e *bundle.EnvironmentConfig) {
        e.SourceEnvironment = "production"
        e.TargetEnvironment = "staging"
        e.SecretHandling = bundle.SecretPlaceholder
        e.ConfigOverrides = map[string]interface{}{
            "api_endpoint": "https://staging.example.com",
            "debug_mode":   true,
        }
        e.VariableMapping = map[string]string{
            "PROD_API_KEY": "STAGING_API_KEY",
            "PROD_DB_HOST": "STAGING_DB_HOST",
        }
    }).
    Build()
```

### Advanced Filtering

```go
customization := bundle.NewCustomizationBuilder().
    WithTemplateFilters(func(tf *bundle.TemplateFilterOptions) {
        // Category-based filtering
        tf.Categories = []string{"security", "monitoring"}
        tf.ExcludeCategories = []string{"experimental"}
        
        // Tag-based filtering
        tf.Tags = []string{"production-ready"}
        tf.ExcludeTags = []string{"beta"}
        
        // Version constraints
        tf.MinVersion = "1.0.0"
        tf.MaxVersion = "2.0.0"
        
        // Time-based filtering
        thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
        tf.ModifiedAfter = &thirtyDaysAgo
        
        // Custom filter function
        tf.CustomFilter = func(template *bundle.TemplateInfo) bool {
            return strings.HasPrefix(template.Name, "owasp-")
        }
    }).
    WithFileFilters(func(ff *bundle.FileFilterOptions) {
        ff.IncludePatterns = []string{"src/**/*.go", "*.yaml"}
        ff.ExcludePatterns = []string{"**/*.test", "temp/**"}
        ff.MaxSize = 100 * 1024 * 1024 // 100MB
    }).
    Build()
```

### Content Transformation

```go
// Implement custom transformer
type ConfigTransformer struct {
    replacements map[string]string
}

func (t *ConfigTransformer) Transform(path string, content []byte) ([]byte, error) {
    result := string(content)
    for old, new := range t.replacements {
        result = strings.ReplaceAll(result, old, new)
    }
    return []byte(result), nil
}

func (t *ConfigTransformer) ShouldTransform(path string) bool {
    return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".json")
}

// Use in customization
customization := bundle.NewCustomizationBuilder().
    WithTransformations(func(t *bundle.TransformationOptions) {
        t.ContentTransformers = []bundle.ContentTransformer{
            &ConfigTransformer{
                replacements: map[string]string{
                    "prod.example.com": "staging.example.com",
                    "production":       "staging",
                },
            },
        }
    }).
    Build()
```

## Configuration Options

### Scope Options

Control which business objects are included in the export:

```go
type ScopeOptions struct {
    IncludeScopes     []string  // Scopes to include
    ExcludeScopes     []string  // Scopes to exclude
    ScopeDepth        int       // Maximum traversal depth
    IncludeOrphaned   bool      // Include orphaned objects
    IncludeDeprecated bool      // Include deprecated objects
}
```

### Environment Configuration

Adapt exports for different environments:

```go
type EnvironmentConfig struct {
    SourceEnvironment string                    // Source environment
    TargetEnvironment string                    // Target environment
    ConfigOverrides   map[string]interface{}    // Config overrides
    SecretHandling    SecretHandlingType        // Secret handling strategy
    VariableMapping   map[string]string         // Variable mapping
}
```

Secret handling strategies:
- `SecretExclude`: Remove secrets from export
- `SecretPlaceholder`: Replace with placeholders
- `SecretEncrypt`: Encrypt secrets
- `SecretInclude`: Include as-is (use with caution)

### Dependency Handling

Configure how dependencies are resolved:

```go
type DependencyHandling struct {
    ResolutionStrategy DependencyStrategy  // Resolution strategy
    MaxDepth          int                 // Max dependency depth
    IncludeOptional   bool                // Include optional deps
    IncludeDevDeps    bool                // Include dev dependencies
    ExcludePatterns   []string            // Exclude patterns
    ForceInclude      []string            // Force include list
}
```

Dependency strategies:
- `DependencyAll`: Include all dependencies
- `DependencyDirect`: Only direct dependencies
- `DependencyMinimal`: Minimal required set
- `DependencyCustom`: Custom strategy

### Template Filtering

Filter templates based on various criteria:

```go
type TemplateFilterOptions struct {
    Categories        []string             // Include categories
    ExcludeCategories []string             // Exclude categories
    Tags              []string             // Required tags
    ExcludeTags       []string             // Excluded tags
    MinVersion        string               // Min version
    MaxVersion        string               // Max version
    ModifiedAfter     *time.Time           // Modified after
    ModifiedBefore    *time.Time           // Modified before
    AuthorFilter      string               // Author filter
    CustomFilter      TemplateFilterFunc   // Custom function
}
```

### Module Filtering

Filter modules based on type, provider, and platform:

```go
type ModuleFilterOptions struct {
    Types            []ModuleType         // Module types
    ExcludeTypes     []ModuleType         // Exclude types
    Providers        []string             // Providers
    ExcludeProviders []string             // Exclude providers
    MinVersion       string               // Min version
    MaxVersion       string               // Max version
    Platforms        []string             // Target platforms
    Architectures    []string             // Target architectures
    CustomFilter     ModuleFilterFunc     // Custom function
}
```

### File Filtering

General file filtering options:

```go
type FileFilterOptions struct {
    IncludePatterns  []string       // Include patterns
    ExcludePatterns  []string       // Exclude patterns
    MinSize          int64          // Min file size
    MaxSize          int64          // Max file size
    ModifiedAfter    *time.Time     // Modified after
    ModifiedBefore   *time.Time     // Modified before
    FileTypes        []string       // File extensions
    ExcludeFileTypes []string       // Exclude extensions
    CustomFilter     FileFilterFunc // Custom function
}
```

### Transformation Options

Transform content during export:

```go
type TransformationOptions struct {
    PathTransformations map[string]string      // Path remapping
    ContentTransformers []ContentTransformer  // Content transformers
    MetadataEnrichment  MetadataEnricher      // Metadata enricher
    Sanitizers          []ContentSanitizer    // Content sanitizers
}
```

### Hotfix Options

Generate deployment scripts:

```go
type HotfixOptions struct {
    GenerateHotfix  bool         // Generate scripts
    TargetPlatforms []string     // Target platforms
    ScriptFormat    ScriptFormat // Script format
    IncludeRollback bool         // Include rollback
    TestMode        bool         // Test mode
    CustomTemplate  string       // Custom template
}
```

### Behavior Options

Control export behavior:

```go
type BehaviorOptions struct {
    ContinueOnError   bool  // Continue on errors
    ValidateContent   bool  // Validate content
    GenerateChecksums bool  // Generate checksums
    CreateBackup      bool  // Create backup
    DryRun            bool  // Dry run mode
    Verbose           bool  // Verbose output
    ParallelExport    bool  // Parallel processing
    MaxParallelJobs   int   // Max parallel jobs
}
```

## Advanced Examples

### Production to Staging Migration

```go
customization := bundle.NewCustomizationBuilder().
    WithEnvironment(func(e *bundle.EnvironmentConfig) {
        e.SourceEnvironment = "production"
        e.TargetEnvironment = "staging"
        e.SecretHandling = bundle.SecretPlaceholder
    }).
    WithTransformations(func(t *bundle.TransformationOptions) {
        t.PathTransformations = map[string]string{
            "configs/prod/": "configs/staging/",
        }
    }).
    WithDependencies(func(d *bundle.DependencyHandling) {
        d.ResolutionStrategy = bundle.DependencyDirect
        d.ExcludePatterns = []string{"*-prod", "*-production"}
    }).
    Build()
```

### Security-Focused Export

```go
customization := bundle.NewCustomizationBuilder().
    WithTemplateFilters(func(tf *bundle.TemplateFilterOptions) {
        tf.Categories = []string{"security", "compliance"}
        tf.Tags = []string{"owasp", "validated"}
    }).
    WithTransformations(func(t *bundle.TransformationOptions) {
        t.Sanitizers = []bundle.ContentSanitizer{
            &CredentialSanitizer{},
            &PIISanitizer{},
        }
    }).
    WithBehavior(func(b *bundle.BehaviorOptions) {
        b.ValidateContent = true
        b.GenerateChecksums = true
    }).
    Build()
```

### Minimal Export for Testing

```go
customization := bundle.NewCustomizationBuilder().
    WithScope(func(s *bundle.ScopeOptions) {
        s.IncludeScopes = []string{"core"}
        s.ScopeDepth = 1
    }).
    WithDependencies(func(d *bundle.DependencyHandling) {
        d.ResolutionStrategy = bundle.DependencyMinimal
        d.IncludeOptional = false
        d.IncludeDevDeps = false
    }).
    WithFileFilters(func(ff *bundle.FileFilterOptions) {
        ff.MaxSize = 10 * 1024 * 1024 // 10MB max
    }).
    Build()
```

## Best Practices

1. **Validate Configurations**: Always call `Validate()` before using
2. **Use Builder Pattern**: Leverage the fluent builder for readability
3. **Test Filters**: Test custom filters with sample data
4. **Document Customizations**: Document complex customizations
5. **Version Control**: Store customization configs in version control
6. **Environment Separation**: Use different customizations per environment
7. **Security First**: Default to excluding sensitive data

## Performance Considerations

- **Parallel Processing**: Enable for large exports
- **Filter Efficiency**: Order filters by selectivity
- **Transform Caching**: Cache transformation results
- **Memory Usage**: Set appropriate file size limits
- **Progress Monitoring**: Use progress handlers for large exports

## Troubleshooting

### Common Issues

1. **Empty Exports**: Check filter configurations
2. **Missing Dependencies**: Verify dependency resolution strategy
3. **Transform Failures**: Validate transformer implementations
4. **Performance Issues**: Enable parallel processing
5. **Memory Problems**: Reduce max file size limits

### Debug Mode

Enable verbose output for troubleshooting:

```go
customization := bundle.NewCustomizationBuilder().
    WithBehavior(func(b *bundle.BehaviorOptions) {
        b.Verbose = true
        b.DryRun = true  // Test without actual export
    }).
    Build()
```

## Integration

### With Export Options

```go
exportOpts := &bundle.ExportOptions{
    OutputPath:   "custom-bundle.tar.gz",
    Format:       bundle.FormatTarGz,
    Compression:  bundle.CompressionZstd,
    // Apply customization filters and transformations
    Customization: customization,
}
```

### With CI/CD

```yaml
# Example CI/CD configuration
export:
  stage: deploy
  script:
    - |
      go run export.go \
        --source-env=$CI_ENVIRONMENT_NAME \
        --target-env=$DEPLOY_TARGET \
        --customization=configs/export-$DEPLOY_TARGET.json
```

## Future Enhancements

1. **Configuration Templates**: Pre-built customization templates
2. **Dynamic Filters**: Runtime filter adjustment
3. **Validation Rules**: Custom validation rule engine
4. **Transform Chains**: Complex transformation pipelines
5. **Export Profiles**: Named, reusable export profiles