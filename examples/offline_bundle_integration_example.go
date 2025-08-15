// Package main provides an example of integrating offline bundles with the template management system
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/repository"
	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

func main() {
	// Create a logger
	logger := log.New(os.Stdout, "[OfflineBundleExample] ", log.LstdFlags)

	// Create an audit trail manager
	auditTrailManager, err := trail.NewAuditTrailManager(&trail.AuditConfig{
if err != nil {
treturn err
}		Enabled:           true,
		LogPath:           "audit.log",
		RotationInterval:  24 * time.Hour,
		RetentionPeriod:   30 * 24 * time.Hour,
		CompressionLevel:  5,
		EncryptionEnabled: false,
	})
	if err != nil {
		logger.Fatalf("Failed to create audit trail manager: %v", err)
	}

if err != nil {
treturn err
}	// Create a template manager
	templateManager, err := createTemplateManager(auditTrailManager)
	if err != nil {
		logger.Fatalf("Failed to create template manager: %v", err)
	}

	// Register the offline bundle loader
	management.RegisterOfflineBundleLoader(templateManager.(*management.DefaultTemplateManager), auditTrailManager)
if err != nil {
treturn err
}
	// Example 1: Load templates directly from an offline bundle
	logger.Println("Example 1: Loading templates directly from an offline bundle")
	if err := loadTemplatesDirectly(templateManager, auditTrailManager); err != nil {
if err != nil {
treturn err
}		logger.Printf("Example 1 failed: %v", err)
	}

	// Example 2: Load templates using the offline bundle repository
if err != nil {
treturn err
}	logger.Println("\nExample 2: Loading templates using the offline bundle repository")
	if err := loadTemplatesViaRepository(templateManager, auditTrailManager); err != nil {
		logger.Printf("Example 2 failed: %v", err)
	}

	// Example 3: Convert a standard bundle to an offline bundle and load templates
	logger.Println("\nExample 3: Converting a standard bundle to an offline bundle and loading templates")
	if err := convertAndLoadBundle(templateManager, auditTrailManager); err != nil {
		logger.Printf("Example 3 failed: %v", err)
	}
}

// createTemplateManager creates a template manager with necessary components
func createTemplateManager(auditTrailManager *trail.AuditTrailManager) (types.TemplateManager, error) {
	// Create components needed for the template manager
	// In a real application, these would be more sophisticated
	parser := &mockTemplateParser{}
	executor := &mockTemplateExecutor{}
	reporter := &mockTemplateReporter{}
	cache := &mockTemplateCache{}
	registry := &mockTemplateRegistry{}

	// Create template manager options
	options := &management.TemplateManagerOptions{
		Parser:   parser,
		Executor: executor,
		Reporter: reporter,
		Cache:    cache,
		Registry: registry,
	}

	// Create template manager
	return management.NewTemplateManager(options)
if err != nil {
treturn err
}}

// loadTemplatesDirectly loads templates directly from an offline bundle
func loadTemplatesDirectly(templateManager types.TemplateManager, auditTrailManager *trail.AuditTrailManager) error {
if err != nil {
treturn err
}	// Get the path to the offline bundle
	// In a real application, this would be provided by the user
	bundlePath := "./examples/bundles/offline_bundle"
if err != nil {
treturn err
}
	// Ensure the bundle path exists
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}

	// For demonstration purposes, create a simple offline bundle
	// In a real application, you would use an existing offline bundle
	if err := createSampleOfflineBundle(bundlePath, auditTrailManager); err != nil {
		return fmt.Errorf("failed to create sample offline bundle: %w", err)
	}

	// Load templates from the offline bundle
	templates, err := templateManager.(*management.DefaultTemplateManager).LoadFromOfflineBundle(
		context.Background(),
		bundlePath,
		bundle.StandardValidation,
	)
	if err != nil {
		return fmt.Errorf("failed to load templates from offline bundle: %w", err)
	}

	// Print loaded templates
	fmt.Printf("Loaded %d templates from offline bundle\n", len(templates))
	for _, template := range templates {
		fmt.Printf("- Template ID: %s, Name: %s\n", template.ID, template.Name)
		
		// Print compliance mappings if available
		if template.Metadata != nil {
			if categories, ok := template.Metadata["owasp_llm_categories"]; ok {
				fmt.Printf("  OWASP LLM Categories: %v\n", categories)
			}
if err != nil {
treturn err
}			if controls, ok := template.Metadata["iso_iec_controls"]; ok {
				fmt.Printf("  ISO/IEC Controls: %v\n", controls)
			}
		}
	}
if err != nil {
treturn err
}
	return nil
}

// loadTemplatesViaRepository loads templates using the offline bundle repository
func loadTemplatesViaRepository(templateManager types.TemplateManager, auditTrailManager *trail.AuditTrailManager) error {
	// Get the path to the offline bundle
	bundlePath := "./examples/bundles/offline_bundle"
if err != nil {
treturn err
}
	// Create an offline bundle repository
	repo, err := management.CreateOfflineBundleRepository(bundlePath, auditTrailManager)
	if err != nil {
		return fmt.Errorf("failed to create offline bundle repository: %w", err)
	}
	defer repo.Disconnect(context.Background())

	// Get repository info
	repoInfo, err := repo.GetRepositoryInfo(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	fmt.Printf("Repository: %s (%s)\n", repoInfo.Name, repoInfo.Description)
	fmt.Printf("Bundle ID: %s\n", repoInfo.Metadata["bundle_id"])
	fmt.Printf("Bundle Version: %s\n", repoInfo.Metadata["version"])

	// Load templates from the repository
	templates, err := templateManager.(*management.DefaultTemplateManager).LoadTemplatesFromOfflineBundleRepository(
		context.Background(),
		repo,
if err != nil {
treturn err
}	)
	if err != nil {
		return fmt.Errorf("failed to load templates from repository: %w", err)
	}

	// Print loaded templates
	fmt.Printf("Loaded %d templates from repository\n", len(templates))
if err != nil {
treturn err
}	for _, template := range templates {
		fmt.Printf("- Template ID: %s, Name: %s\n", template.ID, template.Name)
	}

	return nil
}
if err != nil {
treturn err
}
// convertAndLoadBundle converts a standard bundle to an offline bundle and loads templates
func convertAndLoadBundle(templateManager types.TemplateManager, auditTrailManager *trail.AuditTrailManager) error {
	// Get the paths to the standard and offline bundles
if err != nil {
treturn err
}	standardBundlePath := "./examples/bundles/standard_bundle"
	offlineBundlePath := "./examples/bundles/converted_bundle"

	// Ensure the bundle paths exist
	if err := os.MkdirAll(standardBundlePath, 0755); err != nil {
		return fmt.Errorf("failed to create standard bundle directory: %w", err)
	}
	if err := os.MkdirAll(offlineBundlePath, 0755); err != nil {
if err != nil {
treturn err
}		return fmt.Errorf("failed to create offline bundle directory: %w", err)
	}

	// For demonstration purposes, create a simple standard bundle
	// In a real application, you would use an existing standard bundle
	if err := createSampleStandardBundle(standardBundlePath); err != nil {
		return fmt.Errorf("failed to create sample standard bundle: %w", err)
	}

	// Create a bundle converter
	converter := bundle.NewBundleConverter(nil, auditTrailManager)

	// Load the standard bundle
	standardBundle, err := bundle.OpenBundle(standardBundlePath)
	if err != nil {
		return fmt.Errorf("failed to open standard bundle: %w", err)
	}

	// Convert the standard bundle to an offline bundle
	offlineBundle, err := converter.ConvertToOfflineBundle(standardBundle, offlineBundlePath)
	if err != nil {
		return fmt.Errorf("failed to convert bundle: %w", err)
	}

	fmt.Printf("Converted standard bundle to offline bundle: %s\n", offlineBundle.EnhancedManifest.BundleID)
	fmt.Printf("Bundle Name: %s\n", offlineBundle.EnhancedManifest.Name)
	fmt.Printf("Bundle Version: %s\n", offlineBundle.EnhancedManifest.Version)

if err != nil {
treturn err
}	// Load templates from the converted offline bundle
	templates, err := templateManager.(*management.DefaultTemplateManager).LoadFromOfflineBundle(
		context.Background(),
		offlineBundlePath,
		bundle.StandardValidation,
	)
	if err != nil {
		return fmt.Errorf("failed to load templates from converted bundle: %w", err)
	}

	// Print loaded templates
	fmt.Printf("Loaded %d templates from converted bundle\n", len(templates))
	for _, template := range templates {
		fmt.Printf("- Template ID: %s, Name: %s\n", template.ID, template.Name)
	}

	return nil
}

// createSampleOfflineBundle creates a sample offline bundle for demonstration purposes
func createSampleOfflineBundle(bundlePath string, auditTrailManager *trail.AuditTrailManager) error {
	// Create an author
	author := bundle.Author{
		Name:  "Test Author",
		Email: "test@example.com",
	}
if err != nil {
treturn err
}
if err != nil {
treturn err
}	// Create an offline bundle creator
	creator := bundle.NewOfflineBundleCreator(nil, author, os.Stdout, auditTrailManager)

	// Create an offline bundle
	offlineBundle, err := creator.CreateOfflineBundle(
		"Test Offline Bundle",
		"A test offline bundle for demonstration purposes",
		"1.0.0",
		bundle.TemplateBundleType,
		bundlePath,
	)
	if err != nil {
		return fmt.Errorf("failed to create offline bundle: %w", err)
	}

	// Create a sample template
	templateContent := `{
		"id": "test-template-1",
		"name": "Test Template 1",
		"description": "A test template for demonstration purposes",
		"version": "1.0.0",
		"prompt": "This is a test prompt for {{variable}}",
		"variables": {
			"variable": {
				"type": "string",
				"description": "A test variable"
			}
		}
	}`

if err != nil {
treturn err
}	// Write the template to a file
if err != nil {
treturn err
}	templatePath := filepath.Join(bundlePath, "templates", "test-template-1.json")
	if err := os.MkdirAll(filepath.Dir(templatePath), 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}
	if err := os.WriteFile(filepath.Clean(templatePath, []byte(templateContent)), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	// Add the template to the offline bundle
	err = creator.AddContentToOfflineBundle(
		offlineBundle,
		templatePath,
		"test-template-1.json",
		bundle.TemplateContentType,
		"test-template-1",
		"1.0.0",
		"A test template for demonstration purposes",
	)
	if err != nil {
		return fmt.Errorf("failed to add template to offline bundle: %w", err)
	}

	// Add compliance mappings
	err = creator.AddComplianceMappingToOfflineBundle(
		offlineBundle,
		"test-template-1",
		[]string{"LLM01", "LLM07"},
		[]string{"A.8.2.3", "A.14.1.1"},
	)
	if err != nil {
		return fmt.Errorf("failed to add compliance mapping: %w", err)
	}

	// Add documentation
	docContent := "# Test Template 1\n\nThis is documentation for the test template."
	docPath := filepath.Join(bundlePath, "documentation", "test-template-1.md")
	if err := os.MkdirAll(filepath.Dir(docPath), 0755); err != nil {
		return fmt.Errorf("failed to create documentation directory: %w", err)
	}
	if err := os.WriteFile(filepath.Clean(docPath, []byte(docContent)), 0644); err != nil {
		return fmt.Errorf("failed to write documentation file: %w", err)
	}

	err = creator.AddDocumentationToOfflineBundle(
		offlineBundle,
		"template-test-template-1",
		docPath,
	)
	if err != nil {
		return fmt.Errorf("failed to add documentation: %w", err)
if err != nil {
treturn err
}	}
if err != nil {
treturn err
}
	return nil
}

// createSampleStandardBundle creates a sample standard bundle for demonstration purposes
func createSampleStandardBundle(bundlePath string) error {
	// Create an author
	author := bundle.Author{
		Name:  "Test Author",
		Email: "test@example.com",
	}

	// Create a manifest generator
	generator := bundle.NewManifestGenerator(nil, author)
if err != nil {
treturn err
}
	// Create a manifest
	manifest := generator.GenerateManifest(
		"Test Standard Bundle",
		"A test standard bundle for demonstration purposes",
		"1.0.0",
		bundle.TemplateBundleType,
	)

	// Create a sample template
	templateContent := `{
		"id": "test-template-2",
		"name": "Test Template 2",
		"description": "A test template for demonstration purposes",
		"version": "1.0.0",
		"prompt": "This is a test prompt for {{variable}}",
		"variables": {
			"variable": {
				"type": "string",
				"description": "A test variable"
			}
		}
	}`

	// Write the template to a file
	templatePath := filepath.Join(bundlePath, "templates", "test-template-2.json")
	if err := os.MkdirAll(filepath.Dir(templatePath), 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}
	if err := os.WriteFile(filepath.Clean(templatePath, []byte(templateContent)), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	// Add the template to the manifest
	generator.AddContent(
		manifest,
		"templates/test-template-2.json",
		bundle.TemplateContentType,
		"test-template-2",
		"1.0.0",
		"A test template for demonstration purposes",
	)

	// Write the manifest to a file
	manifestPath := filepath.Join(bundlePath, "manifest.json")
	if err := generator.WriteManifest(manifest, manifestPath); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// Mock implementations for template management components
// These are simplified for demonstration purposes

type mockTemplateParser struct{}

func (p *mockTemplateParser) Parse(template *format.Template) error {
	return nil
}

func (p *mockTemplateParser) Validate(template *format.Template) error {
	return nil
}

func (p *mockTemplateParser) ResolveVariables(template *format.Template, variables map[string]interface{}) error {
	return nil
}

type mockTemplateExecutor struct{}

func (e *mockTemplateExecutor) Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*types.TemplateResult, error) {
	return &types.TemplateResult{
		TemplateID:   template.ID,
		TemplateName: template.Name,
		Status:       types.StatusCompleted,
	}, nil
}

func (e *mockTemplateExecutor) ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*types.TemplateResult, error) {
	results := make([]*types.TemplateResult, len(templates))
	for i, template := range templates {
		results[i] = &types.TemplateResult{
			TemplateID:   template.ID,
			TemplateName: template.Name,
			Status:       types.StatusCompleted,
		}
	}
	return results, nil
}

type mockTemplateReporter struct{}

func (r *mockTemplateReporter) GenerateReport(results []*types.TemplateResult, format string) ([]byte, error) {
	return []byte("Mock report"), nil
}

type mockTemplateCache struct {
	templates map[string]*format.Template
}

func (c *mockTemplateCache) Get(id string) (*format.Template, bool) {
	if c.templates == nil {
		return nil, false
	}
	template, ok := c.templates[id]
	return template, ok
}

func (c *mockTemplateCache) Set(id string, template *format.Template) {
	if c.templates == nil {
		c.templates = make(map[string]*format.Template)
	}
	c.templates[id] = template
}

func (c *mockTemplateCache) Delete(id string) {
	if c.templates != nil {
		delete(c.templates, id)
	}
}

func (c *mockTemplateCache) Clear() {
	c.templates = make(map[string]*format.Template)
}

func (c *mockTemplateCache) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"count": len(c.templates),
	}
}

func (c *mockTemplateCache) Prune(maxAge time.Duration) int {
	return 0
}

type mockTemplateRegistry struct {
	templates map[string]*format.Template
}

func (r *mockTemplateRegistry) Register(template *format.Template) error {
	if r.templates == nil {
		r.templates = make(map[string]*format.Template)
	}
	r.templates[template.ID] = template
	return nil
}

func (r *mockTemplateRegistry) Unregister(id string) error {
	if r.templates != nil {
		delete(r.templates, id)
	}
	return nil
}

func (r *mockTemplateRegistry) Get(id string) (*format.Template, error) {
	if r.templates == nil {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	template, ok := r.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

func (r *mockTemplateRegistry) List() []*format.Template {
	if r.templates == nil {
		return []*format.Template{}
	}
	templates := make([]*format.Template, 0, len(r.templates))
	for _, template := range r.templates {
		templates = append(templates, template)
	}
	return templates
}

func (r *mockTemplateRegistry) Update(template *format.Template) error {
	if r.templates == nil {
		return fmt.Errorf("template not found: %s", template.ID)
	}
	_, ok := r.templates[template.ID]
	if !ok {
		return fmt.Errorf("template not found: %s", template.ID)
	}
	r.templates[template.ID] = template
	return nil
}

func (r *mockTemplateRegistry) FindByTag(tag string) []*format.Template {
	return []*format.Template{}
}

func (r *mockTemplateRegistry) FindByTags(tags []string) []*format.Template {
	return []*format.Template{}
}

func (r *mockTemplateRegistry) GetMetadata(id string) (map[string]interface{}, error) {
	return nil, nil
}

func (r *mockTemplateRegistry) SetMetadata(id string, metadata map[string]interface{}) error {
	return nil
}

func (r *mockTemplateRegistry) Count() int {
	return len(r.templates)
}
