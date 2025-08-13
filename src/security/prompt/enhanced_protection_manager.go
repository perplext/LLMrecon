// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EnhancedProtectionManager extends the ProtectionManager with more sophisticated protection capabilities
type EnhancedProtectionManager struct {
	*ProtectionManager
	config                  *ProtectionConfig
	enhancedConfig          *EnhancedProtectionConfig
	advancedJailbreakDetector *AdvancedJailbreakDetector
	enhancedPatternLibrary  *EnhancedInjectionPatternLibrary
	enhancedContextEnforcer *EnhancedContextBoundaryEnforcer
	enhancedContentFilter   *EnhancedContentFilter
	enhancedApprovalWorkflow *EnhancedApprovalWorkflow
	enhancedReportingSystem *EnhancedReportingSystem
	advancedTemplateMonitor *AdvancedTemplateMonitor
	dataDir                 string
	mu                      sync.RWMutex
}

// EnhancedProtectionConfig defines the configuration for enhanced protection
type EnhancedProtectionConfig struct {
	Level                   ProtectionLevel      `json:"level"`
	EnableAdvancedDetection bool                 `json:"enable_advanced_detection"`
	EnableEnhancedPatterns  bool                 `json:"enable_enhanced_patterns"`
	EnableEnhancedBoundaries bool                `json:"enable_enhanced_boundaries"`
	EnableEnhancedFiltering bool                 `json:"enable_enhanced_filtering"`
	EnableEnhancedApproval  bool                 `json:"enable_enhanced_approval"`
	EnableEnhancedReporting bool                 `json:"enable_enhanced_reporting"`
	EnableAdvancedMonitoring bool                `json:"enable_advanced_monitoring"`
	DataDirectory           string               `json:"data_directory"`
	LogDirectory            string               `json:"log_directory"`
	LogLevel                string               `json:"log_level"`
	AnalysisInterval        time.Duration        `json:"analysis_interval"`
}

// NewEnhancedProtectionManager creates a new enhanced protection manager
func NewEnhancedProtectionManager(config *ProtectionConfig) (*EnhancedProtectionManager, error) {
	if config == nil {
		config = DefaultProtectionConfig()
	}
	
	// Create base protection manager
	baseManager, err := NewProtectionManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create base protection manager: %w", err)
	}
	
	// Initialize enhanced config
	enhancedConfig := &EnhancedProtectionConfig{
		Level:                   config.Level,
		EnableAdvancedDetection: true,
		EnableEnhancedPatterns:  true,
		EnableEnhancedBoundaries: true,
		EnableEnhancedFiltering: true,
		EnableEnhancedApproval:  true,
		EnableEnhancedReporting: true,
		EnableAdvancedMonitoring: true,
		DataDirectory:           "data/security/prompt",
		LogDirectory:            "logs/security/prompt",
		LogLevel:                "info",
		AnalysisInterval:        time.Hour * 24,
	}
	
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(enhancedConfig.DataDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(enhancedConfig.LogDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Create enhanced pattern library
	patternLibraryDir := filepath.Join(enhancedConfig.DataDirectory, "patterns")
	enhancedPatternLibrary, err := NewEnhancedInjectionPatternLibrary(patternLibraryDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create enhanced pattern library: %w", err)
	}
	
	// Create enhanced components
	// Convert EnhancedInjectionPatternLibrary to InjectionPatternLibrary for compatibility
	basePatternLibrary := enhancedPatternLibrary.InjectionPatternLibrary
	advancedJailbreakDetector := NewAdvancedJailbreakDetector(config, basePatternLibrary)
	enhancedContextEnforcer := NewEnhancedContextBoundaryEnforcer(config)
	enhancedContentFilter := NewEnhancedContentFilter(config)
	
	// Create enhanced approval workflow
	approvalWorkflowDir := filepath.Join(enhancedConfig.DataDirectory, "approvals")
	enhancedApprovalWorkflow, err := NewEnhancedApprovalWorkflow(config, approvalWorkflowDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create enhanced approval workflow: %w", err)
	}
	
	// Create enhanced reporting system
	reportingSystemDir := filepath.Join(enhancedConfig.DataDirectory, "reports")
	enhancedReportingSystem, err := NewEnhancedReportingSystem(config, enhancedPatternLibrary, reportingSystemDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create enhanced reporting system: %w", err)
	}
	
	// Create advanced template monitor
	advancedTemplateMonitor := NewAdvancedTemplateMonitor(config, enhancedPatternLibrary)
	
	return &EnhancedProtectionManager{
		ProtectionManager:        baseManager,
		config:                   config,
		enhancedConfig:           enhancedConfig,
		advancedJailbreakDetector: advancedJailbreakDetector,
		enhancedPatternLibrary:   enhancedPatternLibrary,
		enhancedContextEnforcer:  enhancedContextEnforcer,
		enhancedContentFilter:    enhancedContentFilter,
		enhancedApprovalWorkflow: enhancedApprovalWorkflow,
		enhancedReportingSystem:  enhancedReportingSystem,
		advancedTemplateMonitor:  advancedTemplateMonitor,
		dataDir:                  enhancedConfig.DataDirectory,
	}, nil
}

// ProtectPromptEnhanced protects against prompt injection with enhanced protection
func (pm *EnhancedProtectionManager) ProtectPromptEnhanced(ctx context.Context, prompt string, userID string, sessionID string, templateID string) (string, *ProtectionResult, error) {
	startTime := time.Now()
	
	result := &ProtectionResult{
		OriginalPrompt:   prompt,
		ProtectedPrompt:  prompt,
		Detections:       make([]*Detection, 0),
		RiskScore:        0.0,
		ActionTaken:      ActionNone,
		Timestamp:        startTime,
	}
	
	// Apply enhanced context boundary enforcement if enabled
	if pm.enhancedConfig.EnableEnhancedBoundaries && pm.enhancedContextEnforcer != nil {
		protectedPrompt, enforcerResult, err := pm.enhancedContextEnforcer.EnforceBoundariesEnhanced(ctx, prompt)
		if err != nil {
			return prompt, result, err
		}
		
		result.ProtectedPrompt = protectedPrompt
		result.Detections = append(result.Detections, enforcerResult.Detections...)
		result.RiskScore = max(result.RiskScore, enforcerResult.RiskScore)
		
		if enforcerResult.ActionTaken != ActionNone {
			result.ActionTaken = enforcerResult.ActionTaken
		}
		
		// If the prompt was blocked, return immediately
		if enforcerResult.ActionTaken == ActionBlocked {
			result.ProcessingTime = time.Since(startTime)
			return result.ProtectedPrompt, result, nil
		}
	}
	
	// Apply advanced jailbreak detection if enabled
	if pm.enhancedConfig.EnableAdvancedDetection && pm.advancedJailbreakDetector != nil {
		detectorResult, err := pm.advancedJailbreakDetector.DetectAdvancedJailbreak(ctx, result.ProtectedPrompt)
		if err != nil {
			return result.ProtectedPrompt, result, err
		}
		
		result.Detections = append(result.Detections, detectorResult.Detections...)
		result.RiskScore = max(result.RiskScore, detectorResult.RiskScore)
		
		if detectorResult.ActionTaken != ActionNone && detectorResult.ActionTaken > result.ActionTaken {
			result.ActionTaken = detectorResult.ActionTaken
		}
		
		// If the prompt was blocked, return immediately
		if detectorResult.ActionTaken == ActionBlocked {
			result.ProcessingTime = time.Since(startTime)
			return result.ProtectedPrompt, result, nil
		}
	}
	
	// Apply enhanced approval workflow if enabled and risk score exceeds threshold
	if pm.enhancedConfig.EnableEnhancedApproval && pm.enhancedApprovalWorkflow != nil && result.RiskScore >= pm.config.ApprovalThreshold {
		approved, approvalResult, err := pm.enhancedApprovalWorkflow.RequestApprovalEnhanced(ctx, result)
		if err != nil {
			return result.ProtectedPrompt, result, err
		}
		
		if !approved {
			result.ActionTaken = ActionBlocked
			result.ProtectedPrompt = "This operation requires approval but was rejected."
			result.ProcessingTime = time.Since(startTime)
			return result.ProtectedPrompt, result, nil
		}
		
		// Update result with approval result
		if approvalResult != nil {
			result.Detections = append(result.Detections, approvalResult.Detections...)
			result.RiskScore = max(result.RiskScore, approvalResult.RiskScore)
			
			if approvalResult.ActionTaken != ActionNone && approvalResult.ActionTaken > result.ActionTaken {
				result.ActionTaken = approvalResult.ActionTaken
			}
			
			// If the prompt was modified, update it
			if approvalResult.ProtectedPrompt != "" {
				result.ProtectedPrompt = approvalResult.ProtectedPrompt
			}
		}
	}
	
	// Monitor template usage if enabled
	if pm.enhancedConfig.EnableAdvancedMonitoring && pm.advancedTemplateMonitor != nil && templateID != "" {
		// Start monitoring asynchronously to avoid blocking
		go func() {
			if err := pm.advancedTemplateMonitor.MonitorTemplate(context.Background(), templateID, "Unknown", userID, sessionID, prompt, result); err != nil {
				// Log error but don't fail the protection
				fmt.Printf("Error monitoring template: %v\n", err)
			}
		}()
	}
	
	// Report detections if enabled
	if pm.enhancedConfig.EnableEnhancedReporting && pm.enhancedReportingSystem != nil && len(result.Detections) > 0 {
		// Report asynchronously to avoid blocking
		go func() {
			if err := pm.enhancedReportingSystem.ReportInjectionEnhanced(context.Background(), result.Detections, prompt, ""); err != nil {
				// Log error but don't fail the protection
				fmt.Printf("Error reporting injection: %v\n", err)
			}
		}()
	}
	
	// Set processing time
	result.ProcessingTime = time.Since(startTime)
	
	return result.ProtectedPrompt, result, nil
}

// ProtectResponseEnhanced protects against prompt injection in responses with enhanced protection
func (pm *EnhancedProtectionManager) ProtectResponseEnhanced(ctx context.Context, response string, originalPrompt string, userID string, sessionID string, templateID string) (string, *ProtectionResult, error) {
	startTime := time.Now()
	
	result := &ProtectionResult{
		OriginalResponse:  response,
		ProtectedResponse: response,
		Detections:        make([]*Detection, 0),
		RiskScore:         0.0,
		ActionTaken:       ActionNone,
		Timestamp:         startTime,
	}
	
	// Apply enhanced content filtering if enabled
	if pm.enhancedConfig.EnableEnhancedFiltering && pm.enhancedContentFilter != nil {
		filteredResponse, filterResult, err := pm.enhancedContentFilter.FilterContentEnhanced(ctx, response)
		if err != nil {
			return response, result, err
		}
		
		result.ProtectedResponse = filteredResponse
		result.Detections = append(result.Detections, filterResult.Detections...)
		result.RiskScore = max(result.RiskScore, filterResult.RiskScore)
		
		if filterResult.ActionTaken != ActionNone {
			result.ActionTaken = filterResult.ActionTaken
		}
		
		// If the response was blocked, return immediately
		if filterResult.ActionTaken == ActionBlocked {
			result.ProcessingTime = time.Since(startTime)
			return result.ProtectedResponse, result, nil
		}
	}
	
	// Apply advanced jailbreak detection if enabled
	if pm.enhancedConfig.EnableAdvancedDetection && pm.advancedJailbreakDetector != nil {
		detectorResult, err := pm.advancedJailbreakDetector.DetectAdvancedJailbreak(ctx, result.ProtectedResponse)
		if err != nil {
			return result.ProtectedResponse, result, err
		}
		
		result.Detections = append(result.Detections, detectorResult.Detections...)
		result.RiskScore = max(result.RiskScore, detectorResult.RiskScore)
		
		if detectorResult.ActionTaken != ActionNone && detectorResult.ActionTaken > result.ActionTaken {
			result.ActionTaken = detectorResult.ActionTaken
		}
		
		// If the response was blocked, return immediately
		if detectorResult.ActionTaken == ActionBlocked {
			result.ProcessingTime = time.Since(startTime)
			return result.ProtectedResponse, result, nil
		}
	}
	
	// Apply enhanced approval workflow if enabled and risk score exceeds threshold
	if pm.enhancedConfig.EnableEnhancedApproval && pm.enhancedApprovalWorkflow != nil && result.RiskScore >= pm.config.ApprovalThreshold {
		// For responses, we use a different approval request
		// Create a temporary result with the original prompt
		tempResult := &ProtectionResult{
			OriginalPrompt:   originalPrompt,
			ProtectedPrompt:  originalPrompt,
			OriginalResponse: response,
			ProtectedResponse: result.ProtectedResponse,
			Detections:       result.Detections,
			RiskScore:        result.RiskScore,
			ActionTaken:      result.ActionTaken,
			Timestamp:        startTime,
		}
		
		approved, approvalResult, err := pm.enhancedApprovalWorkflow.RequestApprovalEnhanced(ctx, tempResult)
		if err != nil {
			return result.ProtectedResponse, result, err
		}
		
		if !approved {
			result.ActionTaken = ActionBlocked
			result.ProtectedResponse = "This response requires approval but was rejected."
			result.ProcessingTime = time.Since(startTime)
			return result.ProtectedResponse, result, nil
		}
		
		// Update result with approval result
		if approvalResult != nil {
			result.Detections = append(result.Detections, approvalResult.Detections...)
			result.RiskScore = max(result.RiskScore, approvalResult.RiskScore)
			
			if approvalResult.ActionTaken != ActionNone && approvalResult.ActionTaken > result.ActionTaken {
				result.ActionTaken = approvalResult.ActionTaken
			}
			
			// If the response was modified, update it
			if approvalResult.ProtectedResponse != "" {
				result.ProtectedResponse = approvalResult.ProtectedResponse
			}
		}
	}
	
	// Monitor template usage if enabled
	if pm.enhancedConfig.EnableAdvancedMonitoring && pm.advancedTemplateMonitor != nil && templateID != "" {
		// Start monitoring asynchronously to avoid blocking
		go func() {
			if err := pm.advancedTemplateMonitor.MonitorTemplate(context.Background(), templateID, "Unknown", userID, sessionID, originalPrompt, result); err != nil {
				// Log error but don't fail the protection
				fmt.Printf("Error monitoring template: %v\n", err)
			}
		}()
	}
	
	// Report detections if enabled
	if pm.enhancedConfig.EnableEnhancedReporting && pm.enhancedReportingSystem != nil && len(result.Detections) > 0 {
		// Report asynchronously to avoid blocking
		go func() {
			if err := pm.enhancedReportingSystem.ReportInjectionEnhanced(context.Background(), result.Detections, originalPrompt, response); err != nil {
				// Log error but don't fail the protection
				fmt.Printf("Error reporting injection: %v\n", err)
			}
		}()
	}
	
	// Set processing time
	result.ProcessingTime = time.Since(startTime)
	
	return result.ProtectedResponse, result, nil
}

// StartMonitoring starts the template monitoring
func (pm *EnhancedProtectionManager) StartMonitoring(ctx context.Context) error {
	if !pm.enhancedConfig.EnableAdvancedMonitoring || pm.advancedTemplateMonitor == nil {
		return fmt.Errorf("advanced monitoring is not enabled")
	}
	
	return pm.advancedTemplateMonitor.StartMonitoring(ctx)
}

// StopMonitoring stops the template monitoring
func (pm *EnhancedProtectionManager) StopMonitoring() {
	if !pm.enhancedConfig.EnableAdvancedMonitoring || pm.advancedTemplateMonitor == nil {
		return
	}
	
	pm.advancedTemplateMonitor.StopMonitoring()
}

// AnalyzeReports analyzes all reports
func (pm *EnhancedProtectionManager) AnalyzeReports(ctx context.Context) error {
	if !pm.enhancedConfig.EnableEnhancedReporting || pm.enhancedReportingSystem == nil {
		return fmt.Errorf("enhanced reporting is not enabled")
	}
	
	return pm.enhancedReportingSystem.AnalyzeReports(ctx)
}

// GetPatternLibrary gets the enhanced pattern library
func (pm *EnhancedProtectionManager) GetPatternLibrary() *EnhancedInjectionPatternLibrary {
	return pm.enhancedPatternLibrary
}

// GetApprovalWorkflow gets the enhanced approval workflow
func (pm *EnhancedProtectionManager) GetApprovalWorkflow() *EnhancedApprovalWorkflow {
	return pm.enhancedApprovalWorkflow
}

// GetReportingSystem gets the enhanced reporting system
func (pm *EnhancedProtectionManager) GetReportingSystem() *EnhancedReportingSystem {
	return pm.enhancedReportingSystem
}

// GetTemplateMonitor gets the advanced template monitor
func (pm *EnhancedProtectionManager) GetTemplateMonitor() *AdvancedTemplateMonitor {
	return pm.advancedTemplateMonitor
}

// EnableComponent enables or disables a component
func (pm *EnhancedProtectionManager) EnableComponent(component string, enabled bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	switch component {
	case "advanced_detection":
		pm.enhancedConfig.EnableAdvancedDetection = enabled
	case "enhanced_patterns":
		pm.enhancedConfig.EnableEnhancedPatterns = enabled
	case "enhanced_boundaries":
		pm.enhancedConfig.EnableEnhancedBoundaries = enabled
	case "enhanced_filtering":
		pm.enhancedConfig.EnableEnhancedFiltering = enabled
	case "enhanced_approval":
		pm.enhancedConfig.EnableEnhancedApproval = enabled
	case "enhanced_reporting":
		pm.enhancedConfig.EnableEnhancedReporting = enabled
	case "advanced_monitoring":
		pm.enhancedConfig.EnableAdvancedMonitoring = enabled
	}
}

// SetProtectionLevel sets the protection level
func (pm *EnhancedProtectionManager) SetProtectionLevel(level ProtectionLevel) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.enhancedConfig.Level = level
	pm.config.Level = level
	
	// Update component settings based on level
	switch level {
	case LevelLow:
		pm.enhancedConfig.EnableAdvancedDetection = true
		pm.enhancedConfig.EnableEnhancedPatterns = true
		pm.enhancedConfig.EnableEnhancedBoundaries = true
		pm.enhancedConfig.EnableEnhancedFiltering = true
		pm.enhancedConfig.EnableEnhancedApproval = false
		pm.enhancedConfig.EnableEnhancedReporting = true
		pm.enhancedConfig.EnableAdvancedMonitoring = false
	case LevelMedium:
		pm.enhancedConfig.EnableAdvancedDetection = true
		pm.enhancedConfig.EnableEnhancedPatterns = true
		pm.enhancedConfig.EnableEnhancedBoundaries = true
		pm.enhancedConfig.EnableEnhancedFiltering = true
		pm.enhancedConfig.EnableEnhancedApproval = true
		pm.enhancedConfig.EnableEnhancedReporting = true
		pm.enhancedConfig.EnableAdvancedMonitoring = true
	case LevelHigh:
		pm.enhancedConfig.EnableAdvancedDetection = true
		pm.enhancedConfig.EnableEnhancedPatterns = true
		pm.enhancedConfig.EnableEnhancedBoundaries = true
		pm.enhancedConfig.EnableEnhancedFiltering = true
		pm.enhancedConfig.EnableEnhancedApproval = true
		pm.enhancedConfig.EnableEnhancedReporting = true
		pm.enhancedConfig.EnableAdvancedMonitoring = true
	}
}

// Close closes the protection manager and releases resources
func (pm *EnhancedProtectionManager) Close() error {
	// Stop monitoring
	pm.StopMonitoring()
	
	// Close base protection manager
	if err := pm.ProtectionManager.Close(); err != nil {
		return err
	}
	
	return nil
}
