// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt


import (
	"time"
)

// ActionType defines the type of action taken by the protection system
type ActionType int

const (
	// ActionNone indicates no action was taken
	ActionNone ActionType = iota
	// ActionModified indicates the prompt/response was modified
	ActionModified
	// ActionWarned indicates a warning was issued
	ActionWarned
	// ActionBlocked indicates the prompt/response was blocked
	ActionBlocked
	// ActionLogged indicates the prompt/response was logged
	ActionLogged
	// ActionReported indicates the prompt/response was reported
	ActionReported
)

// DetectionType defines the type of detection
type DetectionType string

const (
	// DetectionTypeNone indicates no detection was made
	DetectionTypeNone DetectionType = "none"
	// DetectionTypeInjection is a general injection detection type
	DetectionTypeInjection DetectionType = "injection"
	// DetectionTypePromptInjection indicates a direct prompt injection attempt
	DetectionTypePromptInjection DetectionType = "prompt_injection"
	// DetectionTypeIndirectPromptInjection indicates an indirect prompt injection attempt
	DetectionTypeIndirectPromptInjection DetectionType = "indirect_prompt_injection"
	// DetectionTypeJailbreak indicates a jailbreak attempt
	DetectionTypeJailbreak DetectionType = "jailbreak"
	// DetectionTypeRoleChange indicates a role change attempt
	DetectionTypeRoleChange DetectionType = "role_change"
	// DetectionTypeSystemPrompt indicates a system prompt injection attempt
	DetectionTypeSystemPrompt DetectionType = "system_prompt"
	// DetectionTypeBoundaryViolation indicates a context boundary violation
	DetectionTypeBoundaryViolation DetectionType = "boundary_violation"
	// DetectionTypeDelimiterMisuse indicates misuse of delimiters
	DetectionTypeDelimiterMisuse DetectionType = "delimiter_misuse"
	// DetectionTypeUnusualPattern indicates an unusual pattern
	DetectionTypeUnusualPattern DetectionType = "unusual_pattern"
	// DetectionTypeProhibitedContent indicates prohibited content
	DetectionTypeProhibitedContent DetectionType = "prohibited_content"
	// DetectionTypeApprovalDenied indicates approval was denied
	DetectionTypeApprovalDenied DetectionType = "approval_denied"
	// DetectionTypeSensitiveInfo indicates sensitive information was detected
	DetectionTypeSensitiveInfo DetectionType = "sensitive_info"
	// DetectionTypeSystemInfo indicates system information was detected
	DetectionTypeSystemInfo DetectionType = "system_info"
)

// DetectionLocation defines the location of a detection in a prompt or response
type DetectionLocation struct {
	// Start is the starting index of the detection
	Start int `json:"start"`
	// End is the ending index of the detection
	End int `json:"end"`
	// Context is the surrounding context of the detection
	Context string `json:"context,omitempty"`
}

// Detection defines a detection of a potential security issue
type Detection struct {
	// Type is the type of detection
	Type DetectionType `json:"type"`
	// Confidence is the confidence level of the detection (0.0-1.0)
	Confidence float64 `json:"confidence"`
	// Description is a human-readable description of the detection
	Description string `json:"description"`
	// Location is the location of the detection in the prompt or response
	Location *DetectionLocation `json:"location,omitempty"`
	// Pattern is the pattern that triggered the detection
	Pattern string `json:"pattern,omitempty"`
	// Remediation is a suggested remediation for the detection
	Remediation string `json:"remediation,omitempty"`
	// Metadata contains additional metadata about the detection
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ProtectionResult defines the result of a protection operation
type ProtectionResult struct {
	// OriginalPrompt is the original prompt
	OriginalPrompt string `json:"original_prompt,omitempty"`
	// ProtectedPrompt is the protected prompt
	ProtectedPrompt string `json:"protected_prompt,omitempty"`
	// OriginalResponse is the original response
	OriginalResponse string `json:"original_response,omitempty"`
	// ProtectedResponse is the protected response
	ProtectedResponse string `json:"protected_response,omitempty"`
	// Detections is a list of detections
	Detections []*Detection `json:"detections,omitempty"`
	// RiskScore is the overall risk score (0.0-1.0)
	RiskScore float64 `json:"risk_score"`
	// ActionTaken is the action taken by the protection system
	ActionTaken ActionType `json:"action_taken"`
	// Timestamp is the time of the protection operation
	Timestamp time.Time `json:"timestamp"`
	// ProcessingTime is the time taken to process the protection operation
	ProcessingTime time.Duration `json:"processing_time,omitempty"`
}

// ApprovalRequest defines a request for approval
type ApprovalRequest struct {
	// OriginalPrompt is the original prompt
	OriginalPrompt string `json:"original_prompt"`
	// ProtectedPrompt is the protected prompt
	ProtectedPrompt string `json:"protected_prompt"`
	// Detections is a list of detections
	Detections []*Detection `json:"detections"`
	// RiskScore is the overall risk score (0.0-1.0)
	RiskScore float64 `json:"risk_score"`
	// RequestID is a unique identifier for the request
	RequestID string `json:"request_id"`
	// Timestamp is the time of the request
	Timestamp time.Time `json:"timestamp"`
	// Requester is the identifier of the requester
	Requester string `json:"requester,omitempty"`
	// Reason is the reason for the approval request
	Reason string `json:"reason"`
	// Metadata contains additional metadata about the request
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// InjectionReport defines a report of a new injection technique
type InjectionReport struct {
	// ReportID is a unique identifier for the report
	ReportID string `json:"report_id"`
	// DetectionType is the type of detection
	DetectionType DetectionType `json:"detection_type"`
	// Pattern is the pattern that triggered the detection
	Pattern string `json:"pattern,omitempty"`
	// Example is an example of the injection technique
	Example string `json:"example"`
	// Confidence is the confidence level of the detection (0.0-1.0)
	Confidence float64 `json:"confidence"`
	// Severity is the severity level of the injection technique (0.0-1.0)
	Severity float64 `json:"severity"`
	// Description is a human-readable description of the injection technique
	Description string `json:"description"`
	// Timestamp is the time of the report
	Timestamp time.Time `json:"timestamp"`
	// Source is the source of the report
	Source string `json:"source,omitempty"`
	// Metadata contains additional metadata about the report
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TemplatePatternStats defines statistics for a template pattern
type TemplatePatternStats struct {
	// Pattern is the template pattern
	Pattern string `json:"pattern"`
	// Count is the number of times the pattern has been seen
	Count int `json:"count"`
	// FirstSeen is the time the pattern was first seen
	FirstSeen time.Time `json:"first_seen"`
	// LastSeen is the time the pattern was last seen
	LastSeen time.Time `json:"last_seen"`
	// AverageRiskScore is the average risk score for the pattern
	AverageRiskScore float64 `json:"average_risk_score"`
	// DetectionTypes is a map of detection types to counts
	DetectionTypes map[DetectionType]int `json:"detection_types,omitempty"`
	// Examples contains examples of the pattern
	Examples []string `json:"examples,omitempty"`
	// Metadata contains additional metadata about the pattern
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ContentFilterConfig defines the configuration for content filtering
type ContentFilterConfig struct {
	// EnableProfanityFilter enables filtering of profanity
	EnableProfanityFilter bool `json:"enable_profanity_filter"`
	// EnablePIIFilter enables filtering of personally identifiable information
	EnablePIIFilter bool `json:"enable_pii_filter"`
	// EnableCodeFilter enables filtering of code
	EnableCodeFilter bool `json:"enable_code_filter"`
	// EnableURLFilter enables filtering of URLs
	EnableURLFilter bool `json:"enable_url_filter"`
	// CustomFilters defines custom content filters
	CustomFilters map[string]string `json:"custom_filters,omitempty"`
	// ReplacementChar is the character to use for replacements
	ReplacementChar rune `json:"replacement_char"`
	// FilterThreshold is the threshold for filtering (0.0-1.0)
	FilterThreshold float64 `json:"filter_threshold"`
}

// JailbreakDetectionConfig defines the configuration for jailbreak detection
type JailbreakDetectionConfig struct {
	// EnableRoleChangeDetection enables detection of role changes
	EnableRoleChangeDetection bool `json:"enable_role_change_detection"`
	// EnableSystemPromptDetection enables detection of system prompt injections
	EnableSystemPromptDetection bool `json:"enable_system_prompt_detection"`
	// EnableDelimiterMisuseDetection enables detection of delimiter misuse
	EnableDelimiterMisuseDetection bool `json:"enable_delimiter_misuse_detection"`
	// EnableInstructionOverrideDetection enables detection of instruction overrides
	EnableInstructionOverrideDetection bool `json:"enable_instruction_override_detection"`
	// DetectionThreshold is the threshold for detection (0.0-1.0)
	DetectionThreshold float64 `json:"detection_threshold"`
}

// ContextBoundaryConfig defines the configuration for context boundary enforcement
type ContextBoundaryConfig struct {
	// EnableTokenization enables tokenization of prompts
	EnableTokenization bool `json:"enable_tokenization"`
	// EnableSanitization enables sanitization of prompts
	EnableSanitization bool `json:"enable_sanitization"`
	// EnableNormalization enables normalization of prompts
	EnableNormalization bool `json:"enable_normalization"`
	// MaxPromptLength is the maximum allowed length for prompts
	MaxPromptLength int `json:"max_prompt_length"`
	// SanitizationLevel defines how aggressively to sanitize inputs (1-3)
	SanitizationLevel int `json:"sanitization_level"`
}

// MonitoringConfig defines the configuration for real-time monitoring
type MonitoringConfig struct {
	// MonitoringInterval is the interval for real-time monitoring checks
	MonitoringInterval time.Duration `json:"monitoring_interval"`
	// MaxPatternHistory is the maximum number of patterns to keep in history
	MaxPatternHistory int `json:"max_pattern_history"`
	// AnomalyThreshold is the threshold for anomaly detection (0.0-1.0)
	AnomalyThreshold float64 `json:"anomaly_threshold"`
	// EnableAnomalyDetection enables anomaly detection
	EnableAnomalyDetection bool `json:"enable_anomaly_detection"`
}

// ReportingConfig defines the configuration for the reporting system
type ReportingConfig struct {
	// ReportingEndpoint is the endpoint for reporting
	ReportingEndpoint string `json:"reporting_endpoint,omitempty"`
	// ReportingInterval is the interval for reporting
	ReportingInterval time.Duration `json:"reporting_interval"`
	// MaxReportHistory is the maximum number of reports to keep in history
	MaxReportHistory int `json:"max_report_history"`
	// EnableLocalStorage enables local storage of reports
	EnableLocalStorage bool `json:"enable_local_storage"`
	// LocalStoragePath is the path for local storage
	LocalStoragePath string `json:"local_storage_path,omitempty"`
}

// ApprovalWorkflowConfig defines the configuration for the approval workflow
type ApprovalWorkflowConfig struct {
	// ApprovalThreshold is the risk score threshold for requiring approval
	ApprovalThreshold float64 `json:"approval_threshold"`
	// ApprovalTimeout is the timeout for approval requests
	ApprovalTimeout time.Duration `json:"approval_timeout"`
	// EnableAutoApproval enables automatic approval based on rules
	EnableAutoApproval bool `json:"enable_auto_approval"`
	// AutoApprovalRules defines rules for automatic approval
	AutoApprovalRules map[string]interface{} `json:"auto_approval_rules,omitempty"`
}
