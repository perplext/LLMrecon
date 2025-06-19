# Offline Bundle Integration

This document explains how to integrate offline bundles with the template management system in the LLMreconing Tool.

## Overview

The offline bundle format provides a secure way to package templates and modules for air-gapped environments. The integration with the template management system allows you to:

1. Load templates directly from offline bundles
2. Use offline bundles as a repository source
3. Leverage compliance mappings and documentation from offline bundles
4. Convert standard bundles to offline bundles

## Integration Components

The integration consists of the following components:

- **OfflineBundleLoader**: Implements the `TemplateLoader` interface to load templates from offline bundles
- **OfflineBundleRepository**: Implements the `Repository` interface to access offline bundles as repositories
- **Template Manager Integration**: Extensions to the template manager to work with offline bundles

## Usage Examples

### Loading Templates Directly from an Offline Bundle

```go
// Register the offline bundle loader with the template manager
management.RegisterOfflineBundleLoader(templateManager, auditTrailManager)

// Load templates from the offline bundle
templates, err := templateManager.LoadFromOfflineBundle(
    context.Background(),
    "/path/to/offline/bundle",
    bundle.StandardValidation,
)
if err != nil {
    log.Fatalf("Failed to load templates: %v", err)
}

// Use the loaded templates
for _, template := range templates {
    fmt.Printf("Template: %s (%s)\n", template.Name, template.ID)
    
    // Access compliance mappings from metadata
    if categories, ok := template.Metadata["owasp_llm_categories"]; ok {
        fmt.Printf("OWASP LLM Categories: %v\n", categories)
    }
}
```

### Using an Offline Bundle as a Repository

```go
// Create an offline bundle repository
repo, err := management.CreateOfflineBundleRepository("/path/to/offline/bundle", auditTrailManager)
if err != nil {
    log.Fatalf("Failed to create repository: %v", err)
}
defer repo.Disconnect(context.Background())

// Get repository information
repoInfo, err := repo.GetRepositoryInfo(context.Background())
if err != nil {
    log.Fatalf("Failed to get repository info: %v", err)
}
fmt.Printf("Repository: %s (%s)\n", repoInfo.Name, repoInfo.Description)

// Load templates from the repository
templates, err := templateManager.LoadTemplatesFromOfflineBundleRepository(
    context.Background(),
    repo,
)
if err != nil {
    log.Fatalf("Failed to load templates: %v", err)
}
```

### Converting a Standard Bundle to an Offline Bundle

```go
// Create a bundle converter
converter := bundle.NewBundleConverter(nil, auditTrailManager)

// Load a standard bundle
standardBundle, err := bundle.OpenBundle("/path/to/standard/bundle")
if err != nil {
    log.Fatalf("Failed to open standard bundle: %v", err)
}

// Convert to an offline bundle
offlineBundle, err := converter.ConvertToOfflineBundle(
    standardBundle,
    "/path/to/output/offline/bundle",
)
if err != nil {
    log.Fatalf("Failed to convert bundle: %v", err)
}

// Load templates from the converted bundle
templates, err := templateManager.LoadFromOfflineBundle(
    context.Background(),
    "/path/to/output/offline/bundle",
    bundle.StandardValidation,
)
```

## Compliance Mappings

Offline bundles include compliance mappings that associate templates with compliance frameworks like OWASP LLM Top 10 and ISO/IEC 42001. These mappings are automatically loaded as metadata for templates:

```go
// Access compliance mappings from template metadata
if categories, ok := template.Metadata["owasp_llm_categories"]; ok {
    fmt.Printf("OWASP LLM Categories: %v\n", categories)
}

if controls, ok := template.Metadata["iso_iec_controls"]; ok {
    fmt.Printf("ISO/IEC Controls: %v\n", controls)
}
```

## Documentation References

Offline bundles include documentation for templates and modules. These documentation references are loaded as metadata:

```go
// Access documentation references from template metadata
if docPath, ok := template.Metadata["documentation"]; ok {
    // Load and display documentation
    docContent, err := os.ReadFile(docPath.(string))
    if err == nil {
        fmt.Printf("Documentation: %s\n", string(docContent))
    }
}
```

## Validation Levels

You can specify the validation level when loading templates from offline bundles:

- `bundle.BasicValidation`: Validates the manifest integrity and signature
- `bundle.StandardValidation`: Validates the content integrity in addition to basic validation
- `bundle.StrictValidation`: Validates compatibility in addition to standard validation

```go
// Load templates with strict validation
templates, err := templateManager.LoadFromOfflineBundle(
    context.Background(),
    "/path/to/offline/bundle",
    bundle.StrictValidation,
)
```

## Audit Logging

All operations with offline bundles are logged to the audit trail:

- Loading templates from offline bundles
- Converting bundles
- Accessing templates from offline bundle repositories

This ensures a complete audit trail for compliance and security purposes.

## Complete Example

See the `examples/offline_bundle_integration_example.go` file for a complete example of integrating offline bundles with the template management system.

## Next Steps

- Implement a CLI command to load templates from offline bundles
- Add support for incremental offline bundles
- Enhance the web interface to display compliance information from offline bundles
