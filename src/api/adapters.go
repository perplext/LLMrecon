package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/template/management"
	"github.com/perplext/LLMrecon/src/update"
)

// TemplateServiceAdapter adapts the template manager to the API service interface
type TemplateServiceAdapter struct {
	manager management.TemplateManager
}

// NewTemplateServiceAdapter creates a new template service adapter
func NewTemplateServiceAdapter(manager management.TemplateManager) *TemplateServiceAdapter {
	return &TemplateServiceAdapter{
		manager: manager,
	}
}

// ListTemplates lists templates with filtering
func (a *TemplateServiceAdapter) ListTemplates(filter TemplateFilter) ([]Template, error) {
	// Handle nil manager (placeholder implementation)
	if a.manager == nil {
		return []Template{
			{
				ID:          "example-template",
				Name:        "Example Template",
				Description: "Example security test template",
				Category:    "owasp-llm-01",
				Severity:    "high",
				Version:     "1.0.0",
				Tags:        []string{"injection", "security"},
				LastUpdated: time.Now(),
			},
		}, nil
	}
	
	// Get all templates from manager
	templates, err := a.manager.ListTemplates()
	if err != nil {
		return nil, err
	}
	
	// Convert and filter
	var results []Template
	for _, tmpl := range templates {
		// Apply filters
		if filter.Category != "" && tmpl.GetCategory() != filter.Category {
			continue
		}
		if filter.Severity != "" && tmpl.GetSeverity() != filter.Severity {
			continue
		}
		if filter.Search != "" && !a.matchesSearch(tmpl, filter.Search) {
			continue
		}
		if len(filter.Tags) > 0 && !a.hasTags(tmpl, filter.Tags) {
			continue
		}
		
		// Convert to API template
		apiTemplate := Template{
			ID:          tmpl.GetID(),
			Name:        tmpl.GetName(),
			Description: tmpl.GetDescription(),
			Category:    tmpl.GetCategory(),
			Severity:    tmpl.GetSeverity(),
			Author:      tmpl.GetAuthor(),
			Version:     tmpl.GetVersion(),
			Tags:        tmpl.GetTags(),
			References:  tmpl.GetReferences(),
			LastUpdated: time.Now(), // TODO: Get actual update time
		}
		
		results = append(results, apiTemplate)
	}
	
	return results, nil
}

// GetTemplate gets a single template by ID
func (a *TemplateServiceAdapter) GetTemplate(id string) (*Template, error) {
	// Handle nil manager (placeholder implementation)
	if a.manager == nil {
		return &Template{
			ID:          id,
			Name:        "Example Template",
			Description: "Example security test template",
			Category:    "owasp-llm-01",
			Severity:    "high",
			Version:     "1.0.0",
			Tags:        []string{"injection", "security"},
			LastUpdated: time.Now(),
		}, nil
	}
	
	tmpl, err := a.manager.GetTemplate(id)
	if err != nil {
		return nil, err
	}
	
	return &Template{
		ID:          tmpl.GetID(),
		Name:        tmpl.GetName(),
		Description: tmpl.GetDescription(),
		Category:    tmpl.GetCategory(),
		Severity:    tmpl.GetSeverity(),
		Author:      tmpl.GetAuthor(),
		Version:     tmpl.GetVersion(),
		Tags:        tmpl.GetTags(),
		References:  tmpl.GetReferences(),
		LastUpdated: time.Now(), // TODO: Get actual update time
	}, nil
}

// GetCategories returns all available categories
func (a *TemplateServiceAdapter) GetCategories() ([]string, error) {
	// Handle nil manager (placeholder implementation)
	if a.manager == nil {
		return []string{
			"owasp-llm-01",
			"owasp-llm-02",
			"owasp-llm-03",
			"owasp-llm-04",
			"owasp-llm-05",
		}, nil
	}
	
	templates, err := a.manager.ListTemplates()
	if err != nil {
		return nil, err
	}
	
	// Collect unique categories
	categoryMap := make(map[string]bool)
	for _, tmpl := range templates {
		if cat := tmpl.GetCategory(); cat != "" {
			categoryMap[cat] = true
		}
	}
	
	// Convert to slice
	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	
	return categories, nil
}

// ValidateTemplate validates a template
func (a *TemplateServiceAdapter) ValidateTemplate(template *Template) error {
	// Basic validation - in a real implementation this would be more comprehensive
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Category == "" {
		return fmt.Errorf("template category is required")
	}
	return nil
}

// matchesSearch checks if template matches search query
func (a *TemplateServiceAdapter) matchesSearch(tmpl management.Template, search string) bool {
	search = strings.ToLower(search)
	return strings.Contains(strings.ToLower(tmpl.GetName()), search) ||
		strings.Contains(strings.ToLower(tmpl.GetDescription()), search) ||
		strings.Contains(strings.ToLower(tmpl.GetID()), search)
}

// hasTags checks if template has any of the specified tags
func (a *TemplateServiceAdapter) hasTags(tmpl management.Template, tags []string) bool {
	tmplTags := tmpl.GetTags()
	tagMap := make(map[string]bool)
	for _, tag := range tmplTags {
		tagMap[strings.ToLower(tag)] = true
	}
	
	for _, tag := range tags {
		if tagMap[strings.ToLower(tag)] {
			return true
		}
	}
	
	return false
}

// ModuleServiceAdapter adapts the provider system to the API service interface
type ModuleServiceAdapter struct {
	providerFactory core.ProviderFactory
	registry        map[core.ProviderType]core.Provider
}

// NewModuleServiceAdapter creates a new module service adapter
func NewModuleServiceAdapter(providerFactory core.ProviderFactory) *ModuleServiceAdapter {
	return &ModuleServiceAdapter{
		providerFactory: providerFactory,
		registry:        make(map[core.ProviderType]core.Provider),
	}
}

// ListModules lists available provider modules
func (a *ModuleServiceAdapter) ListModules() ([]Module, error) {
	// Handle nil factory (placeholder implementation)
	if a.providerFactory == nil {
		return []Module{
			{
				ID:          "openai",
				Name:        "OpenAI Provider",
				Type:        "provider",
				Version:     "1.0.0",
				Description: "OpenAI API provider module",
				Provider:    "openai",
				Status:      "available",
				LoadedAt:    time.Now(),
				Config:      ModuleConfig{Enabled: true},
			},
		}, nil
	}
	
	// Get available providers from factory
	providers := a.providerFactory.GetSupportedProviderTypes()
	
	var modules []Module
	for _, providerType := range providers {
		// Get or create provider instance
		_, err := a.getOrCreateProvider(providerType)
		if err != nil {
			continue
		}
		
		// For now, create a basic module info since Provider interface doesn't have GetInfo method
		// In a real implementation, we would need to extend the Provider interface or use a different approach
		module := Module{
			ID:          string(providerType),
			Name:        string(providerType),
			Type:        "provider",
			Version:     "1.0.0",
			Description: fmt.Sprintf("Provider module for %s", providerType),
			Provider:    string(providerType),
			Status:      "available",
			LoadedAt:    time.Now(),
		}
		
		modules = append(modules, module)
	}
	
	return modules, nil
}

// GetModule gets a specific module by ID
func (a *ModuleServiceAdapter) GetModule(id string) (*Module, error) {
	// Handle nil factory (placeholder implementation)
	if a.providerFactory == nil {
		return &Module{
			ID:          id,
			Name:        fmt.Sprintf("%s Provider", id),
			Type:        "provider",
			Version:     "1.0.0",
			Description: fmt.Sprintf("Provider module for %s", id),
			Provider:    id,
			Status:      "loaded",
			LoadedAt:    time.Now(),
			Config:      ModuleConfig{Enabled: true},
		}, nil
	}
	
	providerType := core.ProviderType(id)
	_, err := a.getOrCreateProvider(providerType)
	if err != nil {
		return nil, fmt.Errorf("module not found: %s", id)
	}
	
	return &Module{
		ID:          id,
		Name:        id,
		Type:        "provider",
		Version:     "1.0.0",
		Description: fmt.Sprintf("Provider module for %s", id),
		Provider:    id,
		Status:      "loaded",
		LoadedAt:    time.Now(),
		Config: ModuleConfig{
			Enabled: true,
		},
	}, nil
}

// UpdateModuleConfig updates module configuration
func (a *ModuleServiceAdapter) UpdateModuleConfig(id string, config ModuleConfig) error {
	// TODO: Implement configuration persistence
	// For now, just validate the module exists
	providerType := core.ProviderType(id)
	_, err := a.getOrCreateProvider(providerType)
	if err != nil {
		return fmt.Errorf("module not found: %s", id)
	}
	
	return nil
}

// ReloadModule reloads a module
func (a *ModuleServiceAdapter) ReloadModule(id string) error {
	// For now, just validate the module exists
	providerType := core.ProviderType(id)
	_, err := a.getOrCreateProvider(providerType)
	if err != nil {
		return fmt.Errorf("module not found: %s", id)
	}
	
	return nil
}

// getOrCreateProvider gets or creates a provider instance
func (a *ModuleServiceAdapter) getOrCreateProvider(providerType core.ProviderType) (core.Provider, error) {
	// Handle nil factory
	if a.providerFactory == nil {
		return nil, fmt.Errorf("provider factory not initialized")
	}
	
	// Check registry
	if provider, exists := a.registry[providerType]; exists {
		return provider, nil
	}
	
	// Create new provider with default config
	config := &core.ProviderConfig{
		Type: providerType,
	}
	
	provider, err := a.providerFactory.CreateProvider(config)
	if err != nil {
		return nil, err
	}
	
	// Store in registry
	a.registry[providerType] = provider
	
	return provider, nil
}

// ScanServiceAdapter adapts the scan system to the API service interface
type ScanServiceAdapter struct {
	// scan dependencies would be injected here
}

// CreateScan creates a new scan
func (a *ScanServiceAdapter) CreateScan(request CreateScanRequest) (*Scan, error) {
	// Placeholder implementation
	return &Scan{
		ID:     fmt.Sprintf("scan_%d", time.Now().Unix()),
		Status: ScanStatusPending,
		Target: request.Target,
		Templates: request.Templates,
		Categories: request.Categories,
		Config: request.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetScan gets a scan by ID
func (a *ScanServiceAdapter) GetScan(id string) (*Scan, error) {
	// Placeholder implementation
	return &Scan{
		ID:     id,
		Status: ScanStatusCompleted,
		Target: ScanTarget{Type: "api", Provider: "openai"},
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
		CompletedAt: &[]time.Time{time.Now()}[0],
		Duration: "60s",
		Results: &ScanResults{
			Summary: ResultSummary{
				TotalTests: 10,
				Passed: 8,
				Failed: 2,
				ComplianceScore: 80.0,
			},
			Findings: []Finding{},
			TemplateRuns: []TemplateExecution{},
		},
	}, nil
}

// ListScans lists scans with filtering
func (a *ScanServiceAdapter) ListScans(filter ScanFilter) ([]Scan, error) {
	// Placeholder implementation
	return []Scan{
		{
			ID:     "scan_1",
			Status: ScanStatusCompleted,
			Target: ScanTarget{Type: "api", Provider: "openai"},
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now(),
		},
	}, nil
}

// CancelScan cancels a scan
func (a *ScanServiceAdapter) CancelScan(id string) error {
	// Placeholder implementation
	return nil
}

// GetScanResults gets scan results
func (a *ScanServiceAdapter) GetScanResults(id string) (*ScanResults, error) {
	// Placeholder implementation
	return &ScanResults{
		Summary: ResultSummary{
			TotalTests: 10,
			Passed: 8,
			Failed: 2,
			ComplianceScore: 80.0,
		},
		Findings: []Finding{},
		TemplateRuns: []TemplateExecution{},
	}, nil
}

// UpdateServiceAdapter adapts the update system to the API service interface
type UpdateServiceAdapter struct {
	updateManager *update.UpdateManager
}

// NewUpdateServiceAdapter creates a new update service adapter
func NewUpdateServiceAdapter() *UpdateServiceAdapter {
	return &UpdateServiceAdapter{
		updateManager: update.NewUpdateManager(update.DefaultConfig(), &DefaultLogger{}),
	}
}

// CheckForUpdates checks for available updates
func (a *UpdateServiceAdapter) CheckForUpdates() (*VersionInfo, error) {
	// Check for updates using the update manager
	updateCheck, err := a.updateManager.CheckForUpdates(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	
	// Convert update check to version info
	versionInfo := &VersionInfo{
		API: ComponentVersion{
			Current:         APIVersion,
			Latest:          APIVersion,
			UpdateAvailable: false,
		},
	}
	
	// Process core updates
	if coreUpdate, exists := updateCheck.Components["binary"]; exists {
		versionInfo.Core = ComponentVersion{
			Current:         coreUpdate.CurrentVersion,
			Latest:          coreUpdate.LatestVersion,
			UpdateAvailable: coreUpdate.Available,
			Changelog:       coreUpdate.ChangelogURL,
		}
	}
	
	// Process template updates
	if templateUpdate, exists := updateCheck.Components["templates"]; exists {
		versionInfo.Templates = ComponentVersion{
			Current:         templateUpdate.CurrentVersion,
			Latest:          templateUpdate.LatestVersion,
			UpdateAvailable: templateUpdate.Available,
		}
	}
	
	// Process module updates
	if moduleUpdate, exists := updateCheck.Components["modules"]; exists {
		versionInfo.Modules = ComponentVersion{
			Current:         moduleUpdate.CurrentVersion,
			Latest:          moduleUpdate.LatestVersion,
			UpdateAvailable: moduleUpdate.Available,
		}
	}
	
	return versionInfo, nil
}

// PerformUpdate performs system updates
func (a *UpdateServiceAdapter) PerformUpdate(request UpdateRequest) (*UpdateResponse, error) {
	var updates []UpdateResult
	
	if request.DryRun {
		// For dry run, just simulate the updates
		for _, component := range request.Components {
			updates = append(updates, UpdateResult{
				Component: component,
				Status:    "dry-run",
				Message:   fmt.Sprintf("Would update %s", component),
			})
		}
		
		return &UpdateResponse{
			Status:  "completed",
			Updates: updates,
		}, nil
	}
	
	// Perform actual updates
	updateSummary, err := a.updateManager.ApplyUpdates(context.Background(), request.Components)
	if err != nil {
		return nil, fmt.Errorf("failed to apply updates: %w", err)
	}
	
	// Convert update summary to API response
	for _, result := range updateSummary.Results {
		status := "success"
		message := fmt.Sprintf("%s updated successfully", result.Component)
		
		if !result.Success {
			status = "failed"
			message = "Update failed"
			if result.Error != nil {
				message = result.Error.Error()
			}
		}
		
		updates = append(updates, UpdateResult{
			Component:  result.Component,
			OldVersion: result.OldVersion,
			NewVersion: result.NewVersion,
			Status:     status,
			Message:    message,
		})
	}
	
	// Determine overall status
	status := "completed"
	if !updateSummary.Success {
		status = "failed"
	}
	
	return &UpdateResponse{
		Status:  status,
		Updates: updates,
	}, nil
}

// GetUpdateHistory gets update history
func (a *UpdateServiceAdapter) GetUpdateHistory() ([]UpdateResult, error) {
	// For now, return placeholder history
	return []UpdateResult{
		{
			Component:  "core",
			OldVersion: "1.0.0",
			NewVersion: "1.1.0",
			Status:     "success",
			Message:    "Core updated successfully",
		},
		{
			Component:  "templates",
			OldVersion: "1.0.0",
			NewVersion: "1.1.0",
			Status:     "success",
			Message:    "Templates updated successfully",
		},
	}, nil
}

// BundleServiceAdapter adapts the bundle system to the API service interface
type BundleServiceAdapter struct {
	// bundleManager would be injected here in a real implementation
}

// NewBundleServiceAdapter creates a new bundle service adapter
func NewBundleServiceAdapter() *BundleServiceAdapter {
	return &BundleServiceAdapter{}
}

// ListBundles lists available bundles
func (a *BundleServiceAdapter) ListBundles() ([]BundleInfo, error) {
	// List bundle files in the bundles directory
	bundleDir := filepath.Join(".", "bundles")
	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BundleInfo{}, nil
		}
		return nil, err
	}
	
	var bundles []BundleInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bundle") {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// For now, just use basic file info
		// In a real implementation, this would load bundle metadata
		bundles = append(bundles, BundleInfo{
			ID:          strings.TrimSuffix(entry.Name(), ".bundle"),
			Name:        strings.TrimSuffix(entry.Name(), ".bundle"),
			Version:     "1.0.0",
			Description: "Offline bundle",
			CreatedAt:   info.ModTime(),
			Size:        info.Size(),
		})
	}
	
	return bundles, nil
}

// ExportBundle creates a new bundle
func (a *BundleServiceAdapter) ExportBundle(request ExportBundleRequest) (*BundleOperationResult, error) {
	// For now, just return a placeholder result
	// In a real implementation, this would integrate with the bundle manager
	return &BundleOperationResult{
		BundleID: fmt.Sprintf("bundle_%d", time.Now().Unix()),
		Status:   "success",
		Message:  "Bundle exported successfully",
		Templates: request.Templates,
		Modules:  request.Modules,
	}, nil
}

// ImportBundle imports a bundle
func (a *BundleServiceAdapter) ImportBundle(request ImportBundleRequest) (*BundleOperationResult, error) {
	// For now, just return a placeholder result
	// In a real implementation, this would integrate with the bundle manager
	return &BundleOperationResult{
		BundleID: fmt.Sprintf("bundle_%d", time.Now().Unix()),
		Status:   "success",
		Message:  "Bundle imported successfully",
	}, nil
}

// GetBundle gets a specific bundle by ID
func (a *BundleServiceAdapter) GetBundle(id string) (*BundleInfo, error) {
	return &BundleInfo{
		ID:          id,
		Name:        "Sample Bundle",
		Description: "Sample bundle description",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		Size:        1024,
	}, nil
}

// DeleteBundle deletes a bundle
func (a *BundleServiceAdapter) DeleteBundle(id string) error {
	// In a real implementation, this would delete the bundle
	return nil
}

// ComplianceServiceAdapter adapts the compliance system to the API service interface
type ComplianceServiceAdapter struct {
	templateManager management.TemplateManager
}

// NewComplianceServiceAdapter creates a new compliance service adapter
func NewComplianceServiceAdapter(templateManager management.TemplateManager) *ComplianceServiceAdapter {
	return &ComplianceServiceAdapter{
		templateManager: templateManager,
	}
}

// GenerateReport generates a compliance report
func (a *ComplianceServiceAdapter) GenerateReport(request ComplianceReportRequest) (*ComplianceReport, error) {
	// Handle nil template manager (placeholder implementation)
	var templates []management.Template
	var err error
	
	if a.templateManager == nil {
		// Create placeholder templates for compliance assessment
		templates = []management.Template{} // Empty slice for now
	} else {
		// Get templates from manager
		templates, err = a.templateManager.ListTemplates()
		if err != nil {
			return nil, err
		}
	}
	
	// Calculate compliance based on framework
	var findings []ComplianceResult
	score := 0.0
	
	switch request.Framework {
	case "owasp":
		findings, score = a.assessOWASPCompliance(templates)
	case "iso42001":
		findings, score = a.assessISO42001Compliance(templates)
	case "nist":
		findings, score = a.assessNISTCompliance(templates)
	default:
		return nil, fmt.Errorf("unsupported framework: %s", request.Framework)
	}
	
	return &ComplianceReport{
		ID:          fmt.Sprintf("report_%d", time.Now().Unix()),
		Framework:   request.Framework,
		GeneratedAt: time.Now(),
		Period: CompliancePeriod{
			StartDate: time.Now().AddDate(0, 0, -30),
			EndDate:   time.Now(),
		},
		Summary: ComplianceSummary{
			TotalControls:     len(findings),
			CompliantControls: countCompliantControls(findings),
			ComplianceScore:   score,
			RiskLevel:         getRiskLevel(score),
		},
		Results:    findings,
	}, nil
}

// CheckCompliance checks compliance status
func (a *ComplianceServiceAdapter) CheckCompliance(framework string) (*ComplianceStatus, error) {
	report, err := a.GenerateReport(ComplianceReportRequest{
		Framework: framework,
		Format:    "json",
	})
	if err != nil {
		return nil, err
	}
	
	return &ComplianceStatus{
		Framework:       framework,
		OverallScore:    calculateOverallScore(report.Results),
		RiskLevel:       "medium",
		LastAssessment:  time.Now(),
		ControlsSummary: ComplianceSummary{
			TotalControls:     len(report.Results),
			CompliantControls: countCompliantControls(report.Results),
			ComplianceScore:   calculateOverallScore(report.Results),
		},
	}, nil
}

// GetComplianceHistory gets compliance history for a framework
func (a *ComplianceServiceAdapter) GetComplianceHistory(framework string, days int) ([]ComplianceTrend, error) {
	// For now, return placeholder trend data
	trends := make([]ComplianceTrend, 0, days)
	startDate := time.Now().AddDate(0, 0, -days)
	
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		score := 75.0 + (float64(i%10) * 2.5) // Simulate fluctuating scores
		trends = append(trends, ComplianceTrend{
			Date:  date,
			Score: score,
		})
	}
	
	return trends, nil
}

// assessOWASPCompliance assesses OWASP LLM Top 10 compliance
func (a *ComplianceServiceAdapter) assessOWASPCompliance(templates []management.Template) ([]ComplianceResult, float64) {
	// OWASP LLM Top 10 categories
	categories := []string{
		"llm01-prompt-injection",
		"llm02-insecure-output",
		"llm03-training-data-poisoning",
		"llm04-model-denial-of-service",
		"llm05-supply-chain",
		"llm06-sensitive-information",
		"llm07-insecure-plugin",
		"llm08-excessive-agency",
		"llm09-overreliance",
		"llm10-model-theft",
	}
	
	// Count templates per category
	categoryCount := make(map[string]int)
	for _, tmpl := range templates {
		cat := tmpl.GetCategory()
		if cat != "" {
			categoryCount[cat]++
		}
	}
	
	// Assess each category
	var findings []ComplianceResult
	compliantCount := 0
	
	for _, category := range categories {
		count := categoryCount[category]
		status := "Not Compliant"
		evidence := fmt.Sprintf("No templates for %s", category)
		
		if count > 0 {
			status = "Compliant"
			evidence = fmt.Sprintf("%d templates available for %s", count, category)
			compliantCount++
		}
		
		findings = append(findings, ComplianceResult{
			ControlID:    category,
			ControlName:  fmt.Sprintf("OWASP %s", category),
			Description:  fmt.Sprintf("Security control for %s", category),
			Status:       status,
			Evidence:     []ComplianceEvidence{{
				Type:        "scan",
				Source:      "template_count",
				Description: evidence,
				Timestamp:   time.Now(),
			}},
		})
	}
	
	score := float64(compliantCount) / float64(len(categories)) * 100
	return findings, score
}

// assessISO42001Compliance assesses ISO/IEC 42001 compliance
func (a *ComplianceServiceAdapter) assessISO42001Compliance(templates []management.Template) ([]ComplianceResult, float64) {
	// Simplified ISO 42001 assessment
	requirements := []string{
		"AI Risk Assessment Templates",
		"Model Testing Templates",
		"Data Validation Templates",
		"Security Testing Templates",
		"Performance Testing Templates",
	}
	
	// Check for templates in each area
	var findings []ComplianceResult
	compliantCount := 0
	
	if len(templates) > 0 {
		compliantCount = 3 // Assume we meet some requirements if we have templates
	}
	
	for i, req := range requirements {
		status := "Not Compliant"
		evidence := "No templates available"
		
		if i < compliantCount {
			status = "Compliant"
			evidence = "Templates available"
		}
		
		findings = append(findings, ComplianceResult{
			ControlID:    fmt.Sprintf("iso_%d", i+1),
			ControlName:  req,
			Description:  fmt.Sprintf("ISO 42001 requirement: %s", req),
			Status:       status,
			Evidence:     []ComplianceEvidence{{
				Type:        "scan",
				Source:      "template_count",
				Description: evidence,
				Timestamp:   time.Now(),
			}},
		})
	}
	
	score := float64(compliantCount) / float64(len(requirements)) * 100
	return findings, score
}

// assessNISTCompliance assesses NIST AI RMF compliance
func (a *ComplianceServiceAdapter) assessNISTCompliance(templates []management.Template) ([]ComplianceResult, float64) {
	// Simplified NIST assessment
	functions := []string{
		"GOVERN - AI Risk Management",
		"MAP - Risk Identification",
		"MEASURE - Risk Assessment",
		"MANAGE - Risk Mitigation",
	}
	
	// Check template coverage
	var findings []ComplianceResult
	score := 0.0
	
	if len(templates) >= 10 {
		score = 75.0 // Good coverage
	} else if len(templates) >= 5 {
		score = 50.0 // Partial coverage
	} else if len(templates) > 0 {
		score = 25.0 // Minimal coverage
	}
	
	for _, function := range functions {
		status := "Not Compliant"
		if score >= 50 {
			status = "Partially Compliant"
		}
		if score >= 75 {
			status = "Compliant"
		}
		
		findings = append(findings, ComplianceResult{
			ControlID:    strings.ToLower(strings.ReplaceAll(function, " ", "_")),
			ControlName:  function,
			Description:  fmt.Sprintf("NIST AI RMF function: %s", function),
			Status:       status,
			Evidence:     []ComplianceEvidence{{
				Type:        "scan",
				Source:      "template_count",
				Description: fmt.Sprintf("%d security test templates available", len(templates)),
				Timestamp:   time.Now(),
			}},
		})
	}
	
	return findings, score
}

// Helper functions for compliance calculations
func calculateOverallScore(results []ComplianceResult) float64 {
	if len(results) == 0 {
		return 0.0
	}
	
	compliant := 0
	for _, result := range results {
		if result.Status == "compliant" {
			compliant++
		}
	}
	
	return float64(compliant) / float64(len(results)) * 100.0
}

func countCompliantControls(results []ComplianceResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "compliant" {
			count++
		}
	}
	return count
}

func getRiskLevel(score float64) string {
	if score >= 80 {
		return "low"
	} else if score >= 60 {
		return "medium"
	} else if score >= 40 {
		return "high"
	}
	return "critical"
}

// DefaultLogger provides a simple logger implementation
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string) {
	fmt.Printf("INFO: %s\n", msg)
}

func (l *DefaultLogger) Error(msg string, err error) {
	if err != nil {
		fmt.Printf("ERROR: %s: %v\n", msg, err)
	} else {
		fmt.Printf("ERROR: %s\n", msg)
	}
}

func (l *DefaultLogger) Debug(msg string) {
	fmt.Printf("DEBUG: %s\n", msg)
}

func (l *DefaultLogger) Warn(msg string) {
	fmt.Printf("WARN: %s\n", msg)
}

// CreateAPIServices creates all API service implementations
func CreateAPIServices() (*Services, error) {
	// For now, create placeholder implementations
	// In a real implementation, these would be properly initialized with dependencies
	
	// Create template manager adapter
	templateManager := NewTemplateServiceAdapter(nil) // Would inject real manager
	
	// Create module manager adapter
	moduleManager := NewModuleServiceAdapter(nil) // Would inject real factory
	
	// Create scan service adapter (placeholder)
	scanService := &ScanServiceAdapter{}
	
	// Create update manager adapter
	updateManager := NewUpdateServiceAdapter()
	
	// Create bundle manager adapter
	bundleManager := NewBundleServiceAdapter()
	
	// Create compliance manager adapter
	complianceManager := NewComplianceServiceAdapter(nil) // Would inject real manager
	
	// Create services
	return &Services{
		TemplateManager:   templateManager,
		ModuleManager:     moduleManager,
		ScanEngine:        scanService,
		UpdateManager:     updateManager,
		BundleManager:     bundleManager,
		ComplianceManager: complianceManager,
	}, nil
}