package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/perplext/LLMrecon/src/provider"
	"github.com/perplext/LLMrecon/src/template/management"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// ScanServiceImpl implements the ScanService interface
type ScanServiceImpl struct {
	store           ScanStore
	templateManager management.TemplateManager
	providerFactory provider.Factory
	detectionEngine detection.Engine
	executors       map[string]*scanExecutor
	mu              sync.RWMutex
}

// scanExecutor manages the execution of a single scan
type scanExecutor struct {
	scan      *Scan
	ctx       context.Context
	cancel    context.CancelFunc
	progress  chan ScanProgress
	completed chan bool
}

// ScanProgress represents progress updates during scan execution
type ScanProgress struct {
	TemplateID   string
	TemplateName string
	Status       string
	Message      string
	Timestamp    time.Time
}

// NewScanService creates a new scan service
func NewScanService(
	store ScanStore,
	templateManager management.TemplateManager,
	providerFactory provider.Factory,
	detectionEngine detection.Engine,
) *ScanServiceImpl {
	return &ScanServiceImpl{
		store:           store,
		templateManager: templateManager,
		providerFactory: providerFactory,
		detectionEngine: detectionEngine,
		executors:       make(map[string]*scanExecutor),
	}
}

// CreateScan creates and starts a new scan
func (s *ScanServiceImpl) CreateScan(request CreateScanRequest) (*Scan, error) {
	// Validate request
	if err := s.validateScanRequest(request); err != nil {
		return nil, fmt.Errorf("invalid scan request: %w", err)
	}
	
	// Create scan object
	scan := &Scan{
		ID:         uuid.New().String(),
		Status:     ScanStatusPending,
		Target:     request.Target,
		Templates:  request.Templates,
		Categories: request.Categories,
		Config:     request.Config,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Store scan
	if err := s.store.Create(scan); err != nil {
		return nil, fmt.Errorf("failed to store scan: %w", err)
	}
	
	// Start scan execution
	if err := s.startScanExecution(scan); err != nil {
		// Update scan status to failed
		scan.Status = ScanStatusFailed
		s.store.Update(scan)
		return nil, fmt.Errorf("failed to start scan: %w", err)
	}
	
	return scan, nil
}

// GetScan retrieves a scan by ID
func (s *ScanServiceImpl) GetScan(id string) (*Scan, error) {
	scan, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("scan not found: %w", err)
	}
	
	// Check if scan is running and update progress
	s.mu.RLock()
	if executor, exists := s.executors[id]; exists && scan.Status == ScanStatusRunning {
		// Could add real-time progress info here
		_ = executor
	}
	s.mu.RUnlock()
	
	return scan, nil
}

// ListScans lists scans matching the filter
func (s *ScanServiceImpl) ListScans(filter ScanFilter) ([]Scan, error) {
	return s.store.List(filter)
}

// CancelScan cancels a running scan
func (s *ScanServiceImpl) CancelScan(id string) error {
	// Get scan
	scan, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("scan not found: %w", err)
	}
	
	// Check if scan can be cancelled
	if scan.Status != ScanStatusPending && scan.Status != ScanStatusRunning {
		return fmt.Errorf("cannot cancel scan in status: %s", scan.Status)
	}
	
	// Cancel executor if running
	s.mu.Lock()
	if executor, exists := s.executors[id]; exists {
		executor.cancel()
		delete(s.executors, id)
	}
	s.mu.Unlock()
	
	// Update scan status
	scan.Status = ScanStatusCancelled
	scan.UpdatedAt = time.Now()
	
	return s.store.Update(scan)
}

// GetScanResults retrieves results for a completed scan
func (s *ScanServiceImpl) GetScanResults(id string) (*ScanResults, error) {
	scan, err := s.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("scan not found: %w", err)
	}
	
	if scan.Status != ScanStatusCompleted {
		return nil, fmt.Errorf("scan not completed (status: %s)", scan.Status)
	}
	
	if scan.Results == nil {
		return nil, fmt.Errorf("no results available")
	}
	
	return scan.Results, nil
}

// validateScanRequest validates a scan creation request
func (s *ScanServiceImpl) validateScanRequest(request CreateScanRequest) error {
	// Validate target
	if request.Target.Type == "" {
		return fmt.Errorf("target type is required")
	}
	
	switch request.Target.Type {
	case "api":
		if request.Target.Endpoint == "" {
			return fmt.Errorf("endpoint is required for API target")
		}
		if request.Target.Provider == "" {
			return fmt.Errorf("provider is required for API target")
		}
	case "model":
		if request.Target.Model == "" {
			return fmt.Errorf("model is required for model target")
		}
	default:
		return fmt.Errorf("unsupported target type: %s", request.Target.Type)
	}
	
	// Validate templates or categories specified
	if len(request.Templates) == 0 && len(request.Categories) == 0 {
		// This is okay - will use all available templates
		log.Debug().Msg("No templates or categories specified, will use all available")
	}
	
	return nil
}

// startScanExecution starts the asynchronous execution of a scan
func (s *ScanServiceImpl) startScanExecution(scan *Scan) error {
	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create executor
	executor := &scanExecutor{
		scan:      scan,
		ctx:       ctx,
		cancel:    cancel,
		progress:  make(chan ScanProgress, 100),
		completed: make(chan bool, 1),
	}
	
	// Store executor
	s.mu.Lock()
	s.executors[scan.ID] = executor
	s.mu.Unlock()
	
	// Start execution in background
	go s.executeScan(executor)
	
	return nil
}

// executeScan executes a scan
func (s *ScanServiceImpl) executeScan(executor *scanExecutor) {
	scan := executor.scan
	startTime := time.Now()
	
	// Update scan status to running
	scan.Status = ScanStatusRunning
	scan.StartedAt = &startTime
	s.store.Update(scan)
	
	// Ensure cleanup
	defer func() {
		s.mu.Lock()
		delete(s.executors, scan.ID)
		s.mu.Unlock()
		
		// Calculate duration
		endTime := time.Now()
		scan.CompletedAt = &endTime
		scan.Duration = endTime.Sub(startTime).String()
		
		// Update final status
		if scan.Status == ScanStatusRunning {
			scan.Status = ScanStatusCompleted
		}
		scan.UpdatedAt = time.Now()
		s.store.Update(scan)
		
		close(executor.progress)
		executor.completed <- true
		close(executor.completed)
	}()
	
	// Get templates to execute
	templates, err := s.getTemplatesForScan(scan)
	if err != nil {
		log.Error().Err(err).Str("scan_id", scan.ID).Msg("Failed to get templates")
		scan.Status = ScanStatusFailed
		return
	}
	
	if len(templates) == 0 {
		log.Warn().Str("scan_id", scan.ID).Msg("No templates found for scan")
		scan.Status = ScanStatusCompleted
		scan.Results = &ScanResults{
			Summary: ResultSummary{
				TotalTests: 0,
			},
		}
		return
	}
	
	// Initialize results
	results := &ScanResults{
		Summary:      ResultSummary{
			TotalTests:    len(templates),
			SeverityCount: make(map[string]int),
			CategoryCount: make(map[string]int),
		},
		Findings:     []Finding{},
		Errors:       []ScanError{},
		TemplateRuns: []TemplateExecution{},
	}
	
	// Create provider for the target
	targetProvider, err := s.createProviderForTarget(scan.Target)
	if err != nil {
		log.Error().Err(err).Str("scan_id", scan.ID).Msg("Failed to create provider")
		scan.Status = ScanStatusFailed
		return
	}
	
	// Execute templates
	for _, template := range templates {
		// Check for cancellation
		select {
		case <-executor.ctx.Done():
			log.Info().Str("scan_id", scan.ID).Msg("Scan cancelled")
			scan.Status = ScanStatusCancelled
			return
		default:
		}
		
		// Send progress update
		executor.progress <- ScanProgress{
			TemplateID:   template.GetID(),
			TemplateName: template.GetName(),
			Status:       "starting",
			Timestamp:    time.Now(),
		}
		
		// Execute template
		templateStart := time.Now()
		finding, err := s.executeTemplate(executor.ctx, template, targetProvider, scan.Config)
		templateEnd := time.Now()
		
		// Record execution
		execution := TemplateExecution{
			TemplateID: template.GetID(),
			StartTime:  templateStart,
			EndTime:    templateEnd,
			Duration:   templateEnd.Sub(templateStart).String(),
		}
		
		if err != nil {
			// Template execution error
			execution.Status = "error"
			results.Errors = append(results.Errors, ScanError{
				TemplateID: template.GetID(),
				Error:      err.Error(),
				Timestamp:  time.Now(),
			})
			results.Summary.Errors++
			
			log.Error().
				Err(err).
				Str("scan_id", scan.ID).
				Str("template_id", template.GetID()).
				Msg("Template execution failed")
		} else if finding != nil {
			// Vulnerability found
			execution.Status = "failed"
			results.Findings = append(results.Findings, *finding)
			results.Summary.Failed++
			
			// Update severity count
			results.Summary.SeverityCount[finding.Severity]++
			results.Summary.CategoryCount[finding.Category]++
			
			log.Warn().
				Str("scan_id", scan.ID).
				Str("template_id", template.GetID()).
				Str("severity", finding.Severity).
				Msg("Vulnerability found")
		} else {
			// Test passed
			execution.Status = "passed"
			results.Summary.Passed++
			
			log.Debug().
				Str("scan_id", scan.ID).
				Str("template_id", template.GetID()).
				Msg("Template passed")
		}
		
		results.TemplateRuns = append(results.TemplateRuns, execution)
		
		// Send progress update
		executor.progress <- ScanProgress{
			TemplateID:   template.GetID(),
			TemplateName: template.GetName(),
			Status:       execution.Status,
			Timestamp:    time.Now(),
		}
	}
	
	// Calculate compliance score
	if results.Summary.TotalTests > 0 {
		results.Summary.ComplianceScore = float64(results.Summary.Passed) / float64(results.Summary.TotalTests) * 100
	}
	
	// Store results
	scan.Results = results
	scan.Status = ScanStatusCompleted
	
	log.Info().
		Str("scan_id", scan.ID).
		Int("total", results.Summary.TotalTests).
		Int("passed", results.Summary.Passed).
		Int("failed", results.Summary.Failed).
		Int("errors", results.Summary.Errors).
		Float64("compliance_score", results.Summary.ComplianceScore).
		Msg("Scan completed")
}

// getTemplatesForScan retrieves templates based on scan configuration
func (s *ScanServiceImpl) getTemplatesForScan(scan *Scan) ([]management.Template, error) {
	var templates []management.Template
	
	// If specific templates are requested
	if len(scan.Templates) > 0 {
		for _, templateID := range scan.Templates {
			template, err := s.templateManager.GetTemplate(templateID)
			if err != nil {
				log.Warn().
					Err(err).
					Str("template_id", templateID).
					Msg("Failed to load template")
				continue
			}
			templates = append(templates, template)
		}
		return templates, nil
	}
	
	// If categories are specified
	if len(scan.Categories) > 0 {
		allTemplates, err := s.templateManager.ListTemplates()
		if err != nil {
			return nil, fmt.Errorf("failed to list templates: %w", err)
		}
		
		// Filter by category
		categoryMap := make(map[string]bool)
		for _, cat := range scan.Categories {
			categoryMap[cat] = true
		}
		
		for _, tmpl := range allTemplates {
			if categoryMap[tmpl.GetCategory()] {
				templates = append(templates, tmpl)
			}
		}
		return templates, nil
	}
	
	// Default: use all templates
	return s.templateManager.ListTemplates()
}

// createProviderForTarget creates a provider instance for the scan target
func (s *ScanServiceImpl) createProviderForTarget(target ScanTarget) (provider.Provider, error) {
	config := make(map[string]interface{})
	
	// Merge target parameters into config
	for k, v := range target.Parameters {
		config[k] = v
	}
	
	// Add target-specific configuration
	switch target.Type {
	case "api":
		config["endpoint"] = target.Endpoint
		if target.Model != "" {
			config["model"] = target.Model
		}
	case "model":
		config["model"] = target.Model
	}
	
	// Create provider
	return s.providerFactory.CreateProvider(target.Provider, config)
}

// executeTemplate executes a single template against the target
func (s *ScanServiceImpl) executeTemplate(
	ctx context.Context,
	template management.Template,
	targetProvider provider.Provider,
	config ScanConfig,
) (*Finding, error) {
	// Set timeout for template execution
	timeout := 60 * time.Second // Default timeout
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}
	
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Execute template with retries
	maxRetries := 1
	if config.MaxRetries > 0 {
		maxRetries = config.MaxRetries
	}
	
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(time.Second * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		
		// Execute detection
		result, err := s.detectionEngine.Execute(execCtx, template, targetProvider)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Check if vulnerability was detected
		if result.VulnerabilityDetected {
			// Create finding
			finding := &Finding{
				ID:           uuid.New().String(),
				TemplateID:   template.GetID(),
				TemplateName: template.GetName(),
				Category:     template.GetCategory(),
				Severity:     template.GetSeverity(),
				Title:        result.Title,
				Description:  result.Description,
				Evidence:     result.Evidence,
				Remediation:  result.Remediation,
				References:   template.GetReferences(),
				Timestamp:    time.Now(),
			}
			return finding, nil
		}
		
		// Test passed
		return nil, nil
	}
	
	// All retries failed
	if lastErr != nil {
		return nil, fmt.Errorf("template execution failed after %d attempts: %w", maxRetries, lastErr)
	}
	
	return nil, nil
}