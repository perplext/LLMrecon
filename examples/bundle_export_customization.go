package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/perplext/LLMrecon/src/bundle"
)

// Example content transformers
type EnvironmentVariableTransformer struct {
	sourceEnv string
	targetEnv string
}

func (t *EnvironmentVariableTransformer) Transform(path string, content []byte) ([]byte, error) {
	// Replace environment-specific variables
	result := strings.ReplaceAll(string(content), 
		fmt.Sprintf("${%s_", t.sourceEnv), 
		fmt.Sprintf("${%s_", t.targetEnv))
	return []byte(result), nil
}

func (t *EnvironmentVariableTransformer) ShouldTransform(path string) bool {
	// Transform config files
	ext := filepath.Ext(path)
	return ext == ".yaml" || ext == ".json" || ext == ".env"
}

// Example content sanitizer
type CredentialSanitizer struct{}

func (s *CredentialSanitizer) Sanitize(path string, content []byte) ([]byte, error) {
	// Remove potential credentials
	lines := strings.Split(string(content), "\n")
	var sanitized []string
	
	for _, line := range lines {
		// Simple credential detection
		if strings.Contains(line, "password") || 
		   strings.Contains(line, "api_key") ||
		   strings.Contains(line, "secret") {
			// Replace value with placeholder
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				sanitized = append(sanitized, fmt.Sprintf("%s: <REDACTED>", parts[0]))
			} else {
				sanitized = append(sanitized, line)
			}
		} else {
			sanitized = append(sanitized, line)
		}
	}
	
	return []byte(strings.Join(sanitized, "\n")), nil
}

func (s *CredentialSanitizer) ShouldSanitize(path string) bool {
	// Sanitize config files
	return strings.Contains(path, "config") || strings.Contains(path, ".env")
}

// Example metadata enricher
type BuildMetadataEnricher struct {
	buildNumber string
	gitCommit   string
}

func (e *BuildMetadataEnricher) EnrichMetadata(path string, existing map[string]interface{}) map[string]interface{} {
	if existing == nil {
		existing = make(map[string]interface{})
	}
	
	existing["build_number"] = e.buildNumber
	existing["git_commit"] = e.gitCommit
	existing["export_timestamp"] = time.Now().Format(time.RFC3339)
	
	return existing
}

func main() {
	fmt.Println("Bundle Export Customization Example")
	fmt.Println("===================================")
	
	// Create a comprehensive customization using the builder
	customization := bundle.NewCustomizationBuilder().
		// Configure scope options
		WithScope(func(s *bundle.ScopeOptions) {
			fmt.Println("\n1. Configuring Scope Options:")
			s.IncludeScopes = []string{"templates", "modules", "configs"}
			s.ExcludeScopes = []string{"deprecated", "experimental"}
			s.ScopeDepth = 3
			s.IncludeOrphaned = false
			s.IncludeDeprecated = false
			fmt.Printf("   - Include scopes: %v\n", s.IncludeScopes)
			fmt.Printf("   - Exclude scopes: %v\n", s.ExcludeScopes)
			fmt.Printf("   - Scope depth: %d\n", s.ScopeDepth)
		}).
		// Configure environment-specific settings
		WithEnvironment(func(e *bundle.EnvironmentConfig) {
			fmt.Println("\n2. Configuring Environment Settings:")
			e.SourceEnvironment = "production"
			e.TargetEnvironment = "staging"
			e.SecretHandling = bundle.SecretPlaceholder
			e.ConfigOverrides = map[string]interface{}{
				"api_endpoint": "https://staging-api.example.com",
				"debug_mode":   true,
				"log_level":    "debug",
			}
			e.VariableMapping = map[string]string{
				"PROD_DB_HOST": "STAGING_DB_HOST",
				"PROD_API_KEY": "STAGING_API_KEY",
			}
			fmt.Printf("   - Source: %s → Target: %s\n", e.SourceEnvironment, e.TargetEnvironment)
			fmt.Printf("   - Secret handling: %s\n", e.SecretHandling)
			fmt.Printf("   - Config overrides: %v\n", e.ConfigOverrides)
		}).
		// Configure dependency handling
		WithDependencies(func(d *bundle.DependencyHandling) {
			fmt.Println("\n3. Configuring Dependency Handling:")
			d.ResolutionStrategy = bundle.DependencyDirect
			d.MaxDepth = 2
			d.IncludeOptional = false
			d.IncludeDevDeps = false
			d.ExcludePatterns = []string{
				"test/*",
				"*.test",
				"examples/*",
				"docs/internal/*",
			}
			d.ForceInclude = []string{
				"core-security-module",
				"authentication-provider",
			}
			fmt.Printf("   - Strategy: %s\n", d.ResolutionStrategy)
			fmt.Printf("   - Max depth: %d\n", d.MaxDepth)
			fmt.Printf("   - Exclude patterns: %v\n", d.ExcludePatterns)
			fmt.Printf("   - Force include: %v\n", d.ForceInclude)
		}).
		// Configure template filtering
		WithTemplateFilters(func(tf *bundle.TemplateFilterOptions) {
			fmt.Println("\n4. Configuring Template Filters:")
			tf.Categories = []string{"security", "monitoring", "compliance"}
			tf.ExcludeCategories = []string{"experimental", "legacy"}
			tf.Tags = []string{"production-ready", "validated"}
			tf.ExcludeTags = []string{"beta", "unstable"}
			tf.MinVersion = "1.0.0"
			tf.MaxVersion = "3.0.0"
			
			// Add custom filter for OWASP templates
			tf.CustomFilter = func(template *bundle.TemplateInfo) bool {
				// Include only OWASP-compliant templates
				return strings.Contains(template.Name, "owasp") || 
				       contains(template.Tags, "owasp-compliant")
			}
			
			fmt.Printf("   - Categories: %v\n", tf.Categories)
			fmt.Printf("   - Tags: %v\n", tf.Tags)
			fmt.Printf("   - Version range: %s - %s\n", tf.MinVersion, tf.MaxVersion)
			fmt.Println("   - Custom filter: OWASP-compliant templates only")
		}).
		// Configure module filtering
		WithModuleFilters(func(mf *bundle.ModuleFilterOptions) {
			fmt.Println("\n5. Configuring Module Filters:")
			mf.Types = []bundle.ModuleType{
				bundle.ModuleProvider,
				bundle.ModuleDetector,
			}
			mf.Providers = []string{"openai", "anthropic", "google"}
			mf.MinVersion = "2.0.0"
			mf.Platforms = []string{"linux", "darwin"}
			mf.Architectures = []string{"amd64", "arm64"}
			
			fmt.Printf("   - Types: %v\n", mf.Types)
			fmt.Printf("   - Providers: %v\n", mf.Providers)
			fmt.Printf("   - Platforms: %v\n", mf.Platforms)
			fmt.Printf("   - Architectures: %v\n", mf.Architectures)
		}).
		// Configure file filtering
		WithFileFilters(func(ff *bundle.FileFilterOptions) {
			fmt.Println("\n6. Configuring File Filters:")
			ff.IncludePatterns = []string{
				"src/**/*.go",
				"templates/**/*.yaml",
				"configs/*.json",
				"README.md",
			}
			ff.ExcludePatterns = []string{
				"**/*.log",
				"**/*.tmp",
				"**/node_modules/**",
				".git/**",
				"**/.DS_Store",
			}
			ff.MaxSize = 50 * 1024 * 1024 // 50MB
			ff.FileTypes = []string{".go", ".yaml", ".json", ".md"}
			
			// Only include files modified in the last 30 days
			thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
			ff.ModifiedAfter = &thirtyDaysAgo
			
			fmt.Printf("   - Include patterns: %v\n", ff.IncludePatterns)
			fmt.Printf("   - Max file size: %d MB\n", ff.MaxSize/(1024*1024))
			fmt.Println("   - Modified after: last 30 days")
		}).
		// Configure content transformations
		WithTransformations(func(t *bundle.TransformationOptions) {
			fmt.Println("\n7. Configuring Transformations:")
			
			// Path transformations
			t.PathTransformations = map[string]string{
				"configs/production/": "configs/staging/",
				"logs/prod/":          "logs/staging/",
			}
			
			// Content transformers
			t.ContentTransformers = []bundle.ContentTransformer{
				&EnvironmentVariableTransformer{
					sourceEnv: "PROD",
					targetEnv: "STAGING",
				},
			}
			
			// Content sanitizers
			t.Sanitizers = []bundle.ContentSanitizer{
				&CredentialSanitizer{},
			}
			
			// Metadata enrichment
			t.MetadataEnrichment = &BuildMetadataEnricher{
				buildNumber: "2024.1.15",
				gitCommit:   "abc123def",
			}
			
			fmt.Printf("   - Path transformations: %v\n", t.PathTransformations)
			fmt.Println("   - Content transformers: Environment variable replacer")
			fmt.Println("   - Sanitizers: Credential remover")
			fmt.Println("   - Metadata enrichment: Build info")
		}).
		// Configure hotfix options
		WithHotfix(func(h *bundle.HotfixOptions) {
			fmt.Println("\n8. Configuring Hotfix Generation:")
			h.GenerateHotfix = true
			h.TargetPlatforms = []string{"linux", "darwin", "windows"}
			h.ScriptFormat = bundle.ScriptBash
			h.IncludeRollback = true
			h.TestMode = false
			
			fmt.Printf("   - Generate hotfix: %v\n", h.GenerateHotfix)
			fmt.Printf("   - Target platforms: %v\n", h.TargetPlatforms)
			fmt.Printf("   - Script format: %s\n", h.ScriptFormat)
			fmt.Printf("   - Include rollback: %v\n", h.IncludeRollback)
		}).
		// Configure export behavior
		WithBehavior(func(b *bundle.BehaviorOptions) {
			fmt.Println("\n9. Configuring Export Behavior:")
			b.ContinueOnError = true
			b.ValidateContent = true
			b.GenerateChecksums = true
			b.CreateBackup = true
			b.DryRun = false
			b.Verbose = true
			b.ParallelExport = true
			b.MaxParallelJobs = 4
			
			fmt.Printf("   - Continue on error: %v\n", b.ContinueOnError)
			fmt.Printf("   - Validate content: %v\n", b.ValidateContent)
			fmt.Printf("   - Generate checksums: %v\n", b.GenerateChecksums)
			fmt.Printf("   - Parallel export: %v (max %d jobs)\n", b.ParallelExport, b.MaxParallelJobs)
		}).
		Build()
	
	// Validate the customization
	fmt.Println("\n10. Validating Customization:")
	if err := customization.Validate(); err != nil {
		log.Fatalf("Customization validation failed: %v", err)
	}
	fmt.Println("    ✓ Customization validated successfully")
	
	// Demonstrate filtering
	fmt.Println("\n11. Testing Filters:")
	
	// Test template filter
	testTemplate := &bundle.TemplateInfo{
		Name:     "owasp-prompt-injection",
		Category: "security",
		Version:  "1.5.0",
		Tags:     []string{"production-ready", "owasp-compliant"},
		Author:   "security-team",
	}
	
	if customization.ApplyTemplateFilter(testTemplate) {
		fmt.Printf("    ✓ Template '%s' passed filters\n", testTemplate.Name)
	} else {
		fmt.Printf("    ✗ Template '%s' filtered out\n", testTemplate.Name)
	}
	
	// Test file filter
	testFile := "src/security/scanner.go"
	testFileInfo := bundle.FileInfo{
		Path:     testFile,
		Size:     1024 * 10, // 10KB
		Modified: time.Now().Add(-7 * 24 * time.Hour), // 7 days ago
	}
	
	if customization.ApplyFileFilter(testFile, testFileInfo) {
		fmt.Printf("    ✓ File '%s' passed filters\n", testFile)
	} else {
		fmt.Printf("    ✗ File '%s' filtered out\n", testFile)
	}
	
	// Create export options with customization
	fmt.Println("\n12. Creating Export with Customization:")
	exportOpts := &bundle.ExportOptions{
		OutputPath:       "customized-bundle.tar.gz",
		Format:           bundle.FormatTarGz,
		IncludeBinary:    true,
		IncludeTemplates: true,
		IncludeModules:   true,
		IncludeDocs:      true,
		Compression:      bundle.CompressionZstd,
		Encryption: &bundle.EncryptionOptions{
			Algorithm: bundle.EncryptionChaCha20Poly1305,
			Password:  "staging-deployment-2024",
		},
		// In a real implementation, this would be integrated with the customization
		Filters: &bundle.ExportFilters{
			// Filters would be derived from customization
		},
		Metadata: map[string]interface{}{
			"customization_applied": true,
			"source_environment":    customization.EnvironmentConfig.SourceEnvironment,
			"target_environment":    customization.EnvironmentConfig.TargetEnvironment,
			"export_date":           time.Now().Format(time.RFC3339),
		},
	}
	
	fmt.Printf("    - Output: %s\n", exportOpts.OutputPath)
	fmt.Printf("    - Compression: %s\n", exportOpts.Compression)
	fmt.Printf("    - Encryption: %s\n", exportOpts.Encryption.Algorithm)
	
	fmt.Println("\n✓ Customization example completed!")
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}