// Package execution provides functionality for executing templates against LLM systems.
package execution

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/validation"
)

// LLMProvider is the interface for interacting with LLM systems
type LLMProvider interface {
	// SendPrompt sends a prompt to the LLM and returns the response
	SendPrompt(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
	// GetSupportedModels returns the list of supported models
	GetSupportedModels() []string
	// GetName returns the name of the provider
	GetName() string
}

// DetectionEngine is the interface for detecting vulnerabilities in LLM responses
type DetectionEngine interface {
	// Detect detects vulnerabilities in an LLM response
	Detect(ctx context.Context, template *format.Template, response string) (bool, int, map[string]interface{}, error)
}

// RateLimiter is an alias for interfaces.RateLimiter
type RateLimiter = interfaces.RateLimiter

// ExecutionOptions contains options for template execution
type ExecutionOptions struct {
	// Provider is the LLM provider to use
	Provider LLMProvider
	// DetectionEngine is the detection engine to use
	DetectionEngine DetectionEngine
	// RateLimiter is the rate limiter to use
	RateLimiter RateLimiter
	// InputValidator is the input validator to use
	InputValidator interfaces.InputValidator
	// Timeout is the timeout for template execution
	Timeout time.Duration
	// RetryCount is the number of retries for failed requests
	RetryCount int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	// MaxConcurrent is the maximum number of concurrent executions
	MaxConcurrent int
	// Variables is a map of variables to substitute in templates
	Variables map[string]interface{}
	// ProviderOptions is a map of provider-specific options
	ProviderOptions map[string]interface{}
	// StrictValidation determines if validation errors should fail execution
	StrictValidation bool
	// SanitizePrompts determines if prompts should be sanitized before execution
	SanitizePrompts bool
	// UserID is the ID of the user making the request
	UserID string
	// EnableUserRateLimiting determines if user-specific rate limiting is enabled
	EnableUserRateLimiting bool
}

// TemplateExecutor is responsible for executing templates against LLM systems
type TemplateExecutor struct {
	// defaultOptions is the default execution options
	defaultOptions *ExecutionOptions
	// providers is a map of provider name to provider
	providers map[string]LLMProvider
	// detectionEngines is a map of detection engine name to detection engine
	detectionEngines map[string]DetectionEngine
	// inputValidator is the default input validator
	inputValidator interfaces.InputValidator
	// semaphore is a channel for limiting concurrent executions
	semaphore chan struct{}
}

// NewTemplateExecutor creates a new template executor
func NewTemplateExecutor(defaultOptions *ExecutionOptions) *TemplateExecutor {
	// Set default values
	if defaultOptions.Timeout == 0 {
		defaultOptions.Timeout = 30 * time.Second
	}
	if defaultOptions.RetryCount == 0 {
		defaultOptions.RetryCount = 3
	}
	if defaultOptions.RetryDelay == 0 {
		defaultOptions.RetryDelay = 1 * time.Second
	}
	if defaultOptions.MaxConcurrent == 0 {
		defaultOptions.MaxConcurrent = 10
	}

	// Create default input validator if not provided
	defaultInputValidator := validation.NewInputValidator(defaultOptions.StrictValidation)
	if defaultOptions.InputValidator == nil {
		defaultOptions.InputValidator = defaultInputValidator
	}

	return &TemplateExecutor{
		defaultOptions:   defaultOptions,
		providers:        make(map[string]LLMProvider),
		detectionEngines: make(map[string]DetectionEngine),
		inputValidator:   defaultOptions.InputValidator,
		semaphore:        make(chan struct{}, defaultOptions.MaxConcurrent),
	}
}

// RegisterProvider registers an LLM provider
func (e *TemplateExecutor) RegisterProvider(provider LLMProvider) {
	e.providers[provider.GetName()] = provider
}

// RegisterDetectionEngine registers a detection engine
func (e *TemplateExecutor) RegisterDetectionEngine(name string, engine DetectionEngine) {
	e.detectionEngines[name] = engine
}

// Execute executes a template
func (e *TemplateExecutor) Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*interfaces.TemplateResult, error) {
	// Create result
	result := &interfaces.TemplateResult{
		TemplateID: template.ID,
		Template:   template,
		Timestamp:  time.Now().Unix(),
		StartTime:  time.Now(),
		Details:    make(map[string]interface{}),
	}

	// Merge options with default options
	execOptions := e.mergeOptions(options)

	// Get provider
	provider := execOptions.Provider
	if provider == nil {
		return nil, fmt.Errorf("no provider specified")
	}

	// Get detection engine
	detectionEngine := execOptions.DetectionEngine
	if detectionEngine == nil {
		return nil, fmt.Errorf("no detection engine specified")
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, execOptions.Timeout)
	defer cancel()

	// Acquire semaphore
	select {
	case e.semaphore <- struct{}{}:
		// Acquired semaphore
		defer func() { <-e.semaphore }() // Release semaphore
	case <-execCtx.Done():
		// Context cancelled or timed out
		result.Status = string(interfaces.StatusFailed)
		result.Error = execCtx.Err()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, execCtx.Err()
	}

	// Execute template
	response, err := e.executeWithRetry(execCtx, template, provider, execOptions)
	if err != nil {
		result.Status = string(interfaces.StatusFailed)
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Store response
	result.Response = response

	// Detect vulnerabilities
	detected, score, details, err := detectionEngine.Detect(execCtx, template, response)
	if err != nil {
		result.Status = string(interfaces.StatusFailed)
		result.Error = fmt.Errorf("detection failed: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, result.Error
	}

	// Update result
	result.Detected = detected
	result.Score = score
	result.Details = details
	result.Status = string(interfaces.StatusCompleted)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// ExecuteBatch executes multiple templates
func (e *TemplateExecutor) ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*interfaces.TemplateResult, error) {
	var results []*interfaces.TemplateResult
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(templates))

	for _, template := range templates {
		wg.Add(1)
		go func(template *format.Template) {
			defer wg.Done()

			result, err := e.Execute(ctx, template, options)
			if err != nil {
				errChan <- fmt.Errorf("failed to execute template %s: %w", template.ID, err)
				return
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(template)
	}

	// Wait for all executions to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("some templates failed to execute: %s", errs)
	}

	return results, nil
}

// executeWithRetry executes a template with retry logic
func (e *TemplateExecutor) executeWithRetry(ctx context.Context, template *format.Template, provider LLMProvider, options *ExecutionOptions) (string, error) {
	var response string
	var err error

	// Apply rate limiting if configured
	if options.RateLimiter != nil {
		var err error
		
		// Apply user-specific rate limiting if enabled and user ID is provided
		if options.EnableUserRateLimiting && options.UserID != "" {
			err = options.RateLimiter.AcquireForUser(ctx, options.UserID)
			defer options.RateLimiter.ReleaseForUser(options.UserID)
		} else {
			// Fall back to global rate limiting
			err = options.RateLimiter.Acquire(ctx)
			defer options.RateLimiter.Release()
		}
		
		if err != nil {
			return "", fmt.Errorf("rate limiter acquisition failed: %w", err)
		}
	}

	// Validate template input before execution
	inputValidator := options.InputValidator
	if inputValidator == nil {
		inputValidator = e.inputValidator
	}

	if inputValidator != nil {
		// Set strict mode based on options
		inputValidator.SetStrictMode(options.StrictValidation)

		// Validate template
		if err := inputValidator.ValidateTemplate(ctx, template); err != nil {
			// If in strict mode, validation errors will cause execution to fail
			// Otherwise, warnings will be logged but execution will continue
			if options.StrictValidation {
				return "", fmt.Errorf("template validation failed: %w", err)
			}
			// Log warning but continue with execution
			fmt.Printf("Warning: Template validation issues detected: %v\n", err)
		}
	}

	// Get the prompt from the template
	prompt := template.Test.Prompt

	// Sanitize prompt if enabled
	if options.SanitizePrompts && inputValidator != nil {
		prompt = inputValidator.SanitizePrompt(prompt)
	}

	// Try to execute the template with retries
	for i := 0; i <= options.RetryCount; i++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Continue execution
		}

		// Send prompt to provider
		response, err = provider.SendPrompt(ctx, prompt, options.ProviderOptions)
		if err == nil {
			// Success
			return response, nil
		}

		// Retry if not the last attempt
		if i < options.RetryCount {
			// Wait before retrying
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(options.RetryDelay):
				// Continue with retry
			}
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", options.RetryCount, err)
}

// mergeOptions merges user options with default options
func (e *TemplateExecutor) mergeOptions(userOptions map[string]interface{}) *ExecutionOptions {
	// Create a copy of default options
	options := &ExecutionOptions{
		Provider:              e.defaultOptions.Provider,
		DetectionEngine:       e.defaultOptions.DetectionEngine,
		RateLimiter:           e.defaultOptions.RateLimiter,
		InputValidator:        e.defaultOptions.InputValidator,
		Timeout:               e.defaultOptions.Timeout,
		RetryCount:            e.defaultOptions.RetryCount,
		RetryDelay:            e.defaultOptions.RetryDelay,
		MaxConcurrent:         e.defaultOptions.MaxConcurrent,
		StrictValidation:      e.defaultOptions.StrictValidation,
		SanitizePrompts:       e.defaultOptions.SanitizePrompts,
		UserID:                e.defaultOptions.UserID,
		EnableUserRateLimiting: e.defaultOptions.EnableUserRateLimiting,
		Variables:             make(map[string]interface{}),
		ProviderOptions:       make(map[string]interface{}),
	}

	// Copy default variables
	for k, v := range e.defaultOptions.Variables {
		options.Variables[k] = v
	}

	// Copy default provider options
	for k, v := range e.defaultOptions.ProviderOptions {
		options.ProviderOptions[k] = v
	}

	// Override with user options
	if userOptions != nil {
		// Provider
		if providerName, ok := userOptions["provider"].(string); ok {
			if provider, exists := e.providers[providerName]; exists {
				options.Provider = provider
			}
		}

		// Detection engine
		if engineName, ok := userOptions["detection_engine"].(string); ok {
			if engine, exists := e.detectionEngines[engineName]; exists {
				options.DetectionEngine = engine
			}
		}

		// Validation options
		if strictValidation, ok := userOptions["strict_validation"].(bool); ok {
			options.StrictValidation = strictValidation
			if options.InputValidator != nil {
				options.InputValidator.SetStrictMode(strictValidation)
			}
		}

		if sanitizePrompts, ok := userOptions["sanitize_prompts"].(bool); ok {
			options.SanitizePrompts = sanitizePrompts
		}
		
		// User rate limiting options
		if userID, ok := userOptions["user_id"].(string); ok {
			options.UserID = userID
		}
		
		if enableUserRateLimiting, ok := userOptions["enable_user_rate_limiting"].(bool); ok {
			options.EnableUserRateLimiting = enableUserRateLimiting
		}

		// Timeout
		if timeout, ok := userOptions["timeout"].(time.Duration); ok {
			options.Timeout = timeout
		} else if timeoutSec, ok := userOptions["timeout"].(int); ok {
			options.Timeout = time.Duration(timeoutSec) * time.Second
		}

		// Retry count
		if retryCount, ok := userOptions["retry_count"].(int); ok {
			options.RetryCount = retryCount
		}

		// Retry delay
		if retryDelay, ok := userOptions["retry_delay"].(time.Duration); ok {
			options.RetryDelay = retryDelay
		} else if retryDelaySec, ok := userOptions["retry_delay"].(int); ok {
			options.RetryDelay = time.Duration(retryDelaySec) * time.Second
		}

		// Variables
		if variables, ok := userOptions["variables"].(map[string]interface{}); ok {
			for k, v := range variables {
				options.Variables[k] = v
			}
		}

		// Provider options
		if providerOptions, ok := userOptions["provider_options"].(map[string]interface{}); ok {
			for k, v := range providerOptions {
				options.ProviderOptions[k] = v
			}
		}
	}

	return options
}

// GetProviders returns the list of registered providers
func (e *TemplateExecutor) GetProviders() []string {
	var providers []string
	for name := range e.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetDetectionEngines returns the list of registered detection engines
func (e *TemplateExecutor) GetDetectionEngines() []string {
	var engines []string
	for name := range e.detectionEngines {
		engines = append(engines, name)
	}
	return engines
}

// SetMaxConcurrent sets the maximum number of concurrent executions
func (e *TemplateExecutor) SetMaxConcurrent(max int) {
	if max <= 0 {
		max = 1
	}

	// Create a new semaphore with the new size
	e.semaphore = make(chan struct{}, max)
	e.defaultOptions.MaxConcurrent = max
}

// SetInputValidator sets the input validator
func (e *TemplateExecutor) SetInputValidator(validator interfaces.InputValidator) {
	e.inputValidator = validator
	e.defaultOptions.InputValidator = validator
}

// GetInputValidator gets the input validator
func (e *TemplateExecutor) GetInputValidator() interfaces.InputValidator {
	return e.inputValidator
}

// SetStrictValidation sets the strict validation mode
func (e *TemplateExecutor) SetStrictValidation(strict bool) {
	e.defaultOptions.StrictValidation = strict
	if e.inputValidator != nil {
		e.inputValidator.SetStrictMode(strict)
	}
}

// SetSanitizePrompts sets whether prompts should be sanitized
func (e *TemplateExecutor) SetSanitizePrompts(sanitize bool) {
	e.defaultOptions.SanitizePrompts = sanitize
}

// SetUserID sets the user ID for rate limiting
func (e *TemplateExecutor) SetUserID(userID string) {
	e.defaultOptions.UserID = userID
}

// EnableUserRateLimiting enables or disables user-specific rate limiting
func (e *TemplateExecutor) EnableUserRateLimiting(enabled bool) {
	e.defaultOptions.EnableUserRateLimiting = enabled
}

// ExecuteForUser executes a template for a specific user
func (e *TemplateExecutor) ExecuteForUser(ctx context.Context, template *format.Template, userID string, options map[string]interface{}) (*interfaces.TemplateResult, error) {
	// Create options map if nil
	if options == nil {
		options = make(map[string]interface{})
	}
	
	// Set user ID and enable user rate limiting
	options["user_id"] = userID
	options["enable_user_rate_limiting"] = true
	
	// Execute template with user-specific options
	return e.Execute(ctx, template, options)
}

// ExecuteBatchForUser executes multiple templates for a specific user
func (e *TemplateExecutor) ExecuteBatchForUser(ctx context.Context, templates []*format.Template, userID string, options map[string]interface{}) ([]*interfaces.TemplateResult, error) {
	// Create options map if nil
	if options == nil {
		options = make(map[string]interface{})
	}
	
	// Set user ID and enable user rate limiting
	options["user_id"] = userID
	options["enable_user_rate_limiting"] = true
	
	// Execute templates with user-specific options
	return e.ExecuteBatch(ctx, templates, options)
}
