// Package validation provides methods to validate and detect OWASP LLM vulnerabilities
package validation

import (
	"context"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	// Vulnerable indicates if the prompt or response is vulnerable
	Vulnerable bool
	// VulnerabilityType is the type of vulnerability detected
	VulnerabilityType types.VulnerabilityType
	// Confidence is the confidence level of the detection (0.0 to 1.0)
	Confidence float64
	// Details provides additional information about the vulnerability
	Details string
	// Location indicates where in the prompt or response the vulnerability was detected
	Location *ValidationLocation
	// Severity indicates the severity of the vulnerability
	Severity detection.SeverityLevel
	// Remediation provides suggestions for fixing the vulnerability
	Remediation string
	// RawData contains additional data specific to the vulnerability type
	RawData map[string]interface{}
}

// ValidationLocation represents the location of a vulnerability in a prompt or response
type ValidationLocation struct {
	// StartIndex is the starting character index of the vulnerability
	StartIndex int
	// EndIndex is the ending character index of the vulnerability
	EndIndex int
	// Context is the surrounding text for context
	Context string
}

// PromptValidationOptions represents options for validating prompts
type PromptValidationOptions struct {
	// StrictMode enables more stringent validation
	StrictMode bool
	// IncludeMetadata indicates whether to include metadata in validation
	IncludeMetadata bool
	// CustomPatterns allows for additional patterns to check
	CustomPatterns []string
	// ExcludePatterns allows for patterns to exclude from validation
	ExcludePatterns []string
	// MaxScanDepth sets the maximum depth for recursive scanning
	MaxScanDepth int
	// ContextAware indicates whether to use context-aware validation
	ContextAware bool
}

// ResponseValidationOptions represents options for validating responses
type ResponseValidationOptions struct {
	// StrictMode enables more stringent validation
	StrictMode bool
	// IncludeMetadata indicates whether to include metadata in validation
	IncludeMetadata bool
	// CustomPatterns allows for additional patterns to check
	CustomPatterns []string
	// ExcludePatterns allows for patterns to exclude from validation
	ExcludePatterns []string
	// OriginalPrompt is the prompt that generated the response
	OriginalPrompt string
	// ContextAware indicates whether to use context-aware validation
	ContextAware bool
}

// DefaultPromptValidationOptions returns the default options for prompt validation
func DefaultPromptValidationOptions() *PromptValidationOptions {
	return &PromptValidationOptions{
		StrictMode:      false,
		IncludeMetadata: true,
		MaxScanDepth:    3,
		ContextAware:    true,
	}
}

// DefaultResponseValidationOptions returns the default options for response validation
func DefaultResponseValidationOptions() *ResponseValidationOptions {
	return &ResponseValidationOptions{
		StrictMode:      false,
		IncludeMetadata: true,
		ContextAware:    true,
	}
}

// Validator is the interface that all vulnerability validators must implement
type Validator interface {
	// ValidatePrompt validates a prompt for vulnerabilities
	ValidatePrompt(ctx context.Context, prompt string, options *PromptValidationOptions) ([]*ValidationResult, error)
	
	// ValidateResponse validates a response for vulnerabilities
	ValidateResponse(ctx context.Context, response string, options *ResponseValidationOptions) ([]*ValidationResult, error)
	
	// ValidateChatMessages validates a list of chat messages for vulnerabilities
	ValidateChatMessages(ctx context.Context, messages []core.Message, options *PromptValidationOptions) ([]*ValidationResult, error)
	
	// GetVulnerabilityType returns the vulnerability type that this validator checks for
	GetVulnerabilityType() types.VulnerabilityType
	
	// GetName returns the name of the validator
	GetName() string
	
	// GetDescription returns a description of the validator
	GetDescription() string
}

// BaseValidator provides a base implementation of the Validator interface
type BaseValidator struct {
	// vulnerabilityType is the type of vulnerability this validator checks for
	vulnerabilityType types.VulnerabilityType
	// name is the name of the validator
	name string
	// description is a description of the validator
	description string
}

// NewBaseValidator creates a new base validator
func NewBaseValidator(vulnerabilityType types.VulnerabilityType, name, description string) *BaseValidator {
	return &BaseValidator{
		vulnerabilityType: vulnerabilityType,
		name:              name,
		description:       description,
	}
}

// GetVulnerabilityType returns the vulnerability type that this validator checks for
func (v *BaseValidator) GetVulnerabilityType() types.VulnerabilityType {
	return v.vulnerabilityType
}

// GetName returns the name of the validator
func (v *BaseValidator) GetName() string {
	return v.name
}

// GetDescription returns a description of the validator
func (v *BaseValidator) GetDescription() string {
	return v.description
}

// ValidatePrompt validates a prompt for vulnerabilities
// This is a default implementation that should be overridden by specific validators
func (v *BaseValidator) ValidatePrompt(ctx context.Context, prompt string, options *PromptValidationOptions) ([]*ValidationResult, error) {
	// Default implementation returns no vulnerabilities
	return []*ValidationResult{}, nil
}

// ValidateResponse validates a response for vulnerabilities
// This is a default implementation that should be overridden by specific validators
func (v *BaseValidator) ValidateResponse(ctx context.Context, response string, options *ResponseValidationOptions) ([]*ValidationResult, error) {
	// Default implementation returns no vulnerabilities
	return []*ValidationResult{}, nil
}

// ValidateChatMessages validates a list of chat messages for vulnerabilities
// This is a default implementation that can be overridden by specific validators
func (v *BaseValidator) ValidateChatMessages(ctx context.Context, messages []core.Message, options *PromptValidationOptions) ([]*ValidationResult, error) {
	var results []*ValidationResult
	
	// Validate each message individually based on its role
	for _, message := range messages {
		var messageResults []*ValidationResult
		var err error
		
		// Use ValidatePrompt for user messages and ValidateResponse for assistant messages
		if message.Role == "assistant" {
			// Convert PromptValidationOptions to ResponseValidationOptions
			responseOptions := &ResponseValidationOptions{
				StrictMode:      options.StrictMode,
				IncludeMetadata: options.IncludeMetadata,
				CustomPatterns:  options.CustomPatterns,
				ExcludePatterns: options.ExcludePatterns,
			}
			
			messageResults, err = v.ValidateResponse(ctx, message.Content, responseOptions)
		} else {
			messageResults, err = v.ValidatePrompt(ctx, message.Content, options)
		}
		
		if err != nil {
			return nil, err
		}
		
		// Add message role to raw data
		for _, result := range messageResults {
			if result.RawData == nil {
				result.RawData = make(map[string]interface{})
			}
			result.RawData["message_role"] = message.Role
		}
		
		results = append(results, messageResults...)
	}
	
	return results, nil
}

// CreateValidationResult creates a new validation result
func CreateValidationResult(vulnerable bool, vulnerabilityType types.VulnerabilityType, confidence float64, details string, severity detection.SeverityLevel) *ValidationResult {
	return &ValidationResult{
		Vulnerable:        vulnerable,
		VulnerabilityType: vulnerabilityType,
		Confidence:        confidence,
		Details:           details,
		Severity:          severity,
		RawData:           make(map[string]interface{}),
	}
}

// SetLocation sets the location of the vulnerability in the validation result
func (r *ValidationResult) SetLocation(startIndex, endIndex int, context string) *ValidationResult {
	r.Location = &ValidationLocation{
		StartIndex: startIndex,
		EndIndex:   endIndex,
		Context:    context,
	}
	return r
}

// SetRemediation sets the remediation suggestion in the validation result
func (r *ValidationResult) SetRemediation(remediation string) *ValidationResult {
	r.Remediation = remediation
	return r
}

// AddRawData adds raw data to the validation result
func (r *ValidationResult) AddRawData(key string, value interface{}) *ValidationResult {
	if r.RawData == nil {
		r.RawData = make(map[string]interface{})
	}
	r.RawData[key] = value
	return r
}

// ValidatorRegistry is a registry of all available validators
type ValidatorRegistry struct {
	validators map[types.VulnerabilityType][]Validator
}

// NewValidatorRegistry creates a new validator registry
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{
		validators: make(map[types.VulnerabilityType][]Validator),
	}
}

// RegisterValidator registers a validator with the registry
func (r *ValidatorRegistry) RegisterValidator(validator Validator) {
	vulnerabilityType := validator.GetVulnerabilityType()
	if r.validators[vulnerabilityType] == nil {
		r.validators[vulnerabilityType] = make([]Validator, 0)
	}
	r.validators[vulnerabilityType] = append(r.validators[vulnerabilityType], validator)
}

// GetValidators returns all validators for a specific vulnerability type
func (r *ValidatorRegistry) GetValidators(vulnerabilityType types.VulnerabilityType) []Validator {
	return r.validators[vulnerabilityType]
}

// GetAllValidators returns all registered validators
func (r *ValidatorRegistry) GetAllValidators() []Validator {
	var allValidators []Validator
	for _, validators := range r.validators {
		allValidators = append(allValidators, validators...)
	}
	return allValidators
}

// ValidatePrompt validates a prompt using all registered validators
func (r *ValidatorRegistry) ValidatePrompt(ctx context.Context, prompt string, options *PromptValidationOptions) (map[types.VulnerabilityType][]*ValidationResult, error) {
	if options == nil {
		options = DefaultPromptValidationOptions()
	}
	
	results := make(map[types.VulnerabilityType][]*ValidationResult)
	
	for _, validator := range r.GetAllValidators() {
		validatorResults, err := validator.ValidatePrompt(ctx, prompt, options)
		if err != nil {
			return nil, err
		}
		
		vulnerabilityType := validator.GetVulnerabilityType()
		if results[vulnerabilityType] == nil {
			results[vulnerabilityType] = make([]*ValidationResult, 0)
		}
		results[vulnerabilityType] = append(results[vulnerabilityType], validatorResults...)
	}
	
	return results, nil
}

// ValidateResponse validates a response using all registered validators
func (r *ValidatorRegistry) ValidateResponse(ctx context.Context, response string, options *ResponseValidationOptions) (map[types.VulnerabilityType][]*ValidationResult, error) {
	if options == nil {
		options = DefaultResponseValidationOptions()
	}
	
	results := make(map[types.VulnerabilityType][]*ValidationResult)
	
	for _, validator := range r.GetAllValidators() {
		validatorResults, err := validator.ValidateResponse(ctx, response, options)
		if err != nil {
			return nil, err
		}
		
		vulnerabilityType := validator.GetVulnerabilityType()
		if results[vulnerabilityType] == nil {
			results[vulnerabilityType] = make([]*ValidationResult, 0)
		}
		results[vulnerabilityType] = append(results[vulnerabilityType], validatorResults...)
	}
	
	return results, nil
}

// ValidateChatMessages validates chat messages using all registered validators
func (r *ValidatorRegistry) ValidateChatMessages(ctx context.Context, messages []core.Message, options *PromptValidationOptions) (map[types.VulnerabilityType][]*ValidationResult, error) {
	if options == nil {
		options = DefaultPromptValidationOptions()
	}
	
	results := make(map[types.VulnerabilityType][]*ValidationResult)
	
	// Process each message individually based on its role
	for _, message := range messages {
		if message.Role == "assistant" {
			// For assistant messages, use ValidateResponse
			responseOptions := &ResponseValidationOptions{
				StrictMode:      options.StrictMode,
				IncludeMetadata: options.IncludeMetadata,
				CustomPatterns:  options.CustomPatterns,
				ExcludePatterns: options.ExcludePatterns,
			}
			
			messageResults, err := r.ValidateResponse(ctx, message.Content, responseOptions)
			if err != nil {
				return nil, err
			}
			
			// Merge results
			for vulnType, vulnResults := range messageResults {
				if results[vulnType] == nil {
					results[vulnType] = make([]*ValidationResult, 0)
				}
				results[vulnType] = append(results[vulnType], vulnResults...)
			}
		} else {
			// For user and system messages, use ValidatePrompt
			messageResults, err := r.ValidatePrompt(ctx, message.Content, options)
			if err != nil {
				return nil, err
			}
			
			// Merge results
			for vulnType, vulnResults := range messageResults {
				if results[vulnType] == nil {
					results[vulnType] = make([]*ValidationResult, 0)
				}
				results[vulnType] = append(results[vulnType], vulnResults...)
			}
		}
	}
	
	return results, nil
}
