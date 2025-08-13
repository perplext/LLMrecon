// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"sync"
)

// ProtectionLevel defines the level of protection to apply
type ProtectionLevel int

const (
	// LevelLow provides basic protection with minimal impact on performance
	LevelLow ProtectionLevel = iota
	// LevelMedium provides balanced protection and performance
	LevelMedium
	// LevelHigh provides maximum protection with potential performance impact
	LevelHigh
	// LevelCustom allows for custom configuration of protection mechanisms
	LevelCustom
)

// ProtectionConfig defines the configuration for the prompt injection protection system
type ProtectionConfig struct {
	// Level is the protection level to apply
	Level ProtectionLevel
	// EnableContextBoundaries enables context boundary enforcement
	EnableContextBoundaries bool
	// EnableJailbreakDetection enables detection of jailbreaking attempts
	EnableJailbreakDetection bool
	// EnableRealTimeMonitoring enables real-time monitoring of template patterns
	EnableRealTimeMonitoring bool
	// EnableContentFiltering enables filtering of generated outputs
	EnableContentFiltering bool
	// EnableApprovalWorkflow enables approval workflow for high-risk operations
	EnableApprovalWorkflow bool
	// EnableReportingSystem enables reporting of new prompt injection techniques
	EnableReportingSystem bool
	// CustomPatterns allows for custom prompt injection patterns
	CustomPatterns []string
	// AllowedRoles defines the roles that the LLM is allowed to take
	AllowedRoles []string
	// ForbiddenRoles defines the roles that the LLM is not allowed to take
	ForbiddenRoles []string
	// MaxPromptLength is the maximum allowed length for prompts
	MaxPromptLength int
	// SanitizationLevel defines how aggressively to sanitize inputs
	SanitizationLevel int
	// ApprovalThreshold is the risk score threshold for requiring approval
	ApprovalThreshold float64
	// ApprovalCallback is called when approval is required
	ApprovalCallback func(context.Context, *ApprovalRequest) (bool, error)
	// ReportingCallback is called when a new injection technique is detected
	ReportingCallback func(context.Context, *InjectionReport) error
	// MonitoringInterval is the interval for real-time monitoring checks
	MonitoringInterval time.Duration
}

// DefaultProtectionConfig returns the default protection configuration
func DefaultProtectionConfig() *ProtectionConfig {
	return &ProtectionConfig{
		Level:                   LevelMedium,
		EnableContextBoundaries: true,
		EnableJailbreakDetection: true,
		EnableRealTimeMonitoring: true,
		EnableContentFiltering:   true,
		EnableApprovalWorkflow:   false, // Disabled by default as it requires user interaction
		EnableReportingSystem:    true,
		MaxPromptLength:          8192,  // 8KB default max prompt length
		SanitizationLevel:        2,     // Medium sanitization by default
		ApprovalThreshold:        0.8,   // 80% risk score threshold for approval
		MonitoringInterval:       time.Minute * 5,
	}
}

// HighSecurityProtectionConfig returns a high-security protection configuration
func HighSecurityProtectionConfig() *ProtectionConfig {
	config := DefaultProtectionConfig()
	config.Level = LevelHigh
	config.SanitizationLevel = 3      // High sanitization
	config.ApprovalThreshold = 0.6    // Lower threshold (60%) for requiring approval
	config.MonitoringInterval = time.Minute
	config.EnableApprovalWorkflow = true
	return config
}

// ProtectionManager manages the prompt injection protection system
type ProtectionManager struct {
	config            *ProtectionConfig
	contextEnforcer   *ContextBoundaryEnforcer
	jailbreakDetector *JailbreakDetector
	patternLibrary    *InjectionPatternLibrary
	monitor           *TemplateMonitor
	contentFilter     *ContentFilter
	approvalWorkflow  *ApprovalWorkflow
	reportingSystem   *ReportingSystem
	mu                sync.RWMutex
}

// NewProtectionManager creates a new protection manager
func NewProtectionManager(config *ProtectionConfig) (*ProtectionManager, error) {
	if config == nil {
		config = DefaultProtectionConfig()
	}

	// Create components based on configuration
	var contextEnforcer *ContextBoundaryEnforcer
	var jailbreakDetector *JailbreakDetector
	var patternLibrary *InjectionPatternLibrary
	var monitor *TemplateMonitor
	var contentFilter *ContentFilter
	var approvalWorkflow *ApprovalWorkflow
	var reportingSystem *ReportingSystem

	// Initialize pattern library (always enabled as it's used by other components)
	patternLibrary = NewInjectionPatternLibrary()
	
	// Initialize components based on configuration
	if config.EnableContextBoundaries {
		contextEnforcer = NewContextBoundaryEnforcer(config)
	}
	
	if config.EnableJailbreakDetection {
		jailbreakDetector = NewJailbreakDetector(config, patternLibrary)
	}
	
	if config.EnableRealTimeMonitoring {
		monitor = NewTemplateMonitor(config, patternLibrary)
	}
	
	if config.EnableContentFiltering {
		contentFilter = NewContentFilter(config)
	}
	
	if config.EnableApprovalWorkflow {
		approvalWorkflow = NewApprovalWorkflow(config)
	}
	
	if config.EnableReportingSystem {
		reportingSystem = NewReportingSystem(config)
	}

	return &ProtectionManager{
		config:            config,
		contextEnforcer:   contextEnforcer,
		jailbreakDetector: jailbreakDetector,
		patternLibrary:    patternLibrary,
		monitor:           monitor,
		contentFilter:     contentFilter,
		approvalWorkflow:  approvalWorkflow,
		reportingSystem:   reportingSystem,
	}, nil
}

// ProtectPrompt protects against prompt injection in user inputs
func (pm *ProtectionManager) ProtectPrompt(ctx context.Context, prompt string) (string, *ProtectionResult, error) {
	result := &ProtectionResult{
		OriginalPrompt: prompt,
		ProtectedPrompt: prompt,
		Detections: make([]*Detection, 0),
		RiskScore: 0.0,
		ActionTaken: ActionNone,
	}

	// Apply context boundary enforcement if enabled
	if pm.contextEnforcer != nil {
		protectedPrompt, enforcerResult, err := pm.contextEnforcer.EnforceBoundaries(ctx, prompt)
		if err != nil {
			return prompt, result, err
		}
		
		result.ProtectedPrompt = protectedPrompt
		result.Detections = append(result.Detections, enforcerResult.Detections...)
		result.RiskScore = max(result.RiskScore, enforcerResult.RiskScore)
		
		if enforcerResult.ActionTaken != ActionNone {
			result.ActionTaken = enforcerResult.ActionTaken
		}
	}

	// Apply jailbreak detection if enabled
	if pm.jailbreakDetector != nil {
		detectorResult, err := pm.jailbreakDetector.DetectJailbreak(ctx, result.ProtectedPrompt)
		if err != nil {
			return result.ProtectedPrompt, result, err
		}
		
		result.Detections = append(result.Detections, detectorResult.Detections...)
		result.RiskScore = max(result.RiskScore, detectorResult.RiskScore)
		
		if detectorResult.ActionTaken != ActionNone && detectorResult.ActionTaken > result.ActionTaken {
			result.ActionTaken = detectorResult.ActionTaken
		}
	}

	// Apply approval workflow if enabled and risk score exceeds threshold
	if pm.approvalWorkflow != nil && result.RiskScore >= pm.config.ApprovalThreshold {
		approved, approvalResult, err := pm.approvalWorkflow.RequestApproval(ctx, result)
		if err != nil {
			return result.ProtectedPrompt, result, err
		}
		
		if !approved {
			result.ProtectedPrompt = ""
			result.ActionTaken = ActionBlocked
			result.Detections = append(result.Detections, &Detection{
				Type:        DetectionTypeApprovalDenied,
				Confidence:  1.0,
				Description: "Prompt was denied by approval workflow",
				Location:    &DetectionLocation{Start: 0, End: len(prompt)},
			})
		} else {
			// If approved, use the potentially modified prompt from the approval process
			result.ProtectedPrompt = approvalResult.ProtectedPrompt
			result.Detections = append(result.Detections, approvalResult.Detections...)
		}
	}

	// Apply real-time monitoring if enabled (doesn't modify the prompt, just monitors)
	if pm.monitor != nil {
		go pm.monitor.MonitorPrompt(ctx, result)
	}

	// Apply reporting if enabled and new patterns were detected
	if pm.reportingSystem != nil && len(result.Detections) > 0 {
		go pm.reportingSystem.ReportDetections(ctx, result)
	}

	return result.ProtectedPrompt, result, nil
}

// ProtectResponse protects against prompt injection in LLM responses
func (pm *ProtectionManager) ProtectResponse(ctx context.Context, response string, originalPrompt string) (string, *ProtectionResult, error) {
	result := &ProtectionResult{
		OriginalPrompt: originalPrompt,
		OriginalResponse: response,
		ProtectedResponse: response,
		Detections: make([]*Detection, 0),
		RiskScore: 0.0,
		ActionTaken: ActionNone,
	}

	// Apply content filtering if enabled
	if pm.contentFilter != nil {
		filteredResponse, filterResult, err := pm.contentFilter.FilterContent(ctx, response, originalPrompt)
		if err != nil {
			return response, result, err
		}
		
		result.ProtectedResponse = filteredResponse
		result.Detections = append(result.Detections, filterResult.Detections...)
		result.RiskScore = max(result.RiskScore, filterResult.RiskScore)
		
		// Set ActionTaken based on filter result or detections
		if filterResult.ActionTaken != ActionNone {
			result.ActionTaken = filterResult.ActionTaken
		} else if filteredResponse != response {
			// If the response was modified but no action was explicitly set, set it to Modified
			result.ActionTaken = ActionModified
		} else if len(filterResult.Detections) > 0 {
			// If detections were found but no action taken, set appropriate action based on detection type
			for _, detection := range filterResult.Detections {
				if detection.Type == DetectionTypeSensitiveInfo || detection.Type == DetectionTypeSystemInfo {
					result.ActionTaken = ActionModified
					break
				}
			}
		}
	}

	// Apply reporting if enabled and new patterns were detected
	if pm.reportingSystem != nil && len(result.Detections) > 0 {
		go pm.reportingSystem.ReportDetections(ctx, result)
	}

	return result.ProtectedResponse, result, nil
}

// Close closes the protection manager and releases resources
func (pm *ProtectionManager) Close() error {
	if pm.monitor != nil {
		pm.monitor.Stop()
	}
	
	if pm.reportingSystem != nil {
		pm.reportingSystem.Close()
	}
	
	return nil
}

// Helper function to get the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
