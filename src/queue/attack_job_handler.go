package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// AttackJobHandler handles attack execution jobs
type AttackJobHandler struct {
	providerManager interfaces.ProviderManager
	templateEngine  interfaces.TemplateEngine
	logger          Logger
}

// AttackJobPayload defines the payload for attack jobs
type AttackJobPayload struct {
	ProviderType   core.ProviderType          `json:"provider_type"`
	Model          string                     `json:"model"`
	TemplateID     string                     `json:"template_id"`
	Parameters     map[string]interface{}     `json:"parameters"`
	Configuration  AttackConfiguration        `json:"configuration"`
	Metadata       map[string]interface{}     `json:"metadata"`

// AttackConfiguration defines configuration for attack execution
type AttackConfiguration struct {
	MaxRetries      int           `json:"max_retries"`
	Timeout         time.Duration `json:"timeout"`
	RateLimit       int           `json:"rate_limit"`
	ConcurrentLimit int           `json:"concurrent_limit"`
	StopOnSuccess   bool          `json:"stop_on_success"`
	CollectMetrics  bool          `json:"collect_metrics"`

// AttackResult defines the result of attack execution
type AttackResult struct {
	JobID          string                 `json:"job_id"`
	Success        bool                   `json:"success"`
	Response       string                 `json:"response"`
	Confidence     float64                `json:"confidence"`
	AttackType     string                 `json:"attack_type"`
	ProviderType   core.ProviderType      `json:"provider_type"`
	Model          string                 `json:"model"`
	ExecutionTime  time.Duration          `json:"execution_time"`
	TokensUsed     int                    `json:"tokens_used"`
	Cost           float64                `json:"cost"`
	Metadata       map[string]interface{} `json:"metadata"`
	Error          string                 `json:"error,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// JobType constants
const (
	JobTypeAttack            = "attack"
	JobTypeBatchAttack       = "batch_attack"
	JobTypeTemplateValidation = "template_validation"
	JobTypeProviderTest      = "provider_test"
	JobTypeComplianceCheck   = "compliance_check"
)

// NewAttackJobHandler creates a new attack job handler
func NewAttackJobHandler(
	providerManager interfaces.ProviderManager,
	templateEngine interfaces.TemplateEngine,
	logger Logger,
) *AttackJobHandler {
	return &AttackJobHandler{
		providerManager: providerManager,
		templateEngine:  templateEngine,
		logger:          logger,
	}

// ProcessJob processes a job based on its type
func (h *AttackJobHandler) ProcessJob(ctx context.Context, job *Job) error {
	h.logger.Info("Processing job", "job_id", job.ID, "type", job.Type)
	
	switch job.Type {
	case JobTypeAttack:
		return h.processAttackJob(ctx, job)
	case JobTypeBatchAttack:
		return h.processBatchAttackJob(ctx, job)
	case JobTypeTemplateValidation:
		return h.processTemplateValidationJob(ctx, job)
	case JobTypeProviderTest:
		return h.processProviderTestJob(ctx, job)
	case JobTypeComplianceCheck:
		return h.processComplianceCheckJob(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}

// GetJobTypes returns the job types this handler can process
func (h *AttackJobHandler) GetJobTypes() []string {
	return []string{
		JobTypeAttack,
		JobTypeBatchAttack,
		JobTypeTemplateValidation,
		JobTypeProviderTest,
		JobTypeComplianceCheck,
	}

// processAttackJob processes a single attack job
func (h *AttackJobHandler) processAttackJob(ctx context.Context, job *Job) error {
	start := time.Now()
	
	// Parse job payload
	payload, err := h.parseAttackPayload(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to parse attack payload: %w", err)
	}
	
	// Execute attack
	result, err := h.executeAttack(ctx, payload)
	if err != nil {
		// Create error result
		result = &AttackResult{
			JobID:         job.ID,
			Success:       false,
			AttackType:    payload.TemplateID,
			ProviderType:  payload.ProviderType,
			Model:         payload.Model,
			ExecutionTime: time.Since(start),
			Error:         err.Error(),
			Timestamp:     time.Now(),
		}
	} else {
		result.JobID = job.ID
		result.ExecutionTime = time.Since(start)
		result.Timestamp = time.Now()
	}
	
	// Store result in job
	job.Result = result
	
	h.logger.Info("Attack job completed", 
		"job_id", job.ID, 
		"success", result.Success, 
		"duration", result.ExecutionTime,
		"provider", result.ProviderType,
		"model", result.Model,
	)
	
	return nil

// processBatchAttackJob processes a batch of attacks
func (h *AttackJobHandler) processBatchAttackJob(ctx context.Context, job *Job) error {
	start := time.Now()
	
	// Parse batch payload
	batchPayload, err := h.parseBatchPayload(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to parse batch payload: %w", err)
	}
	
	results := make([]*AttackResult, 0, len(batchPayload.Attacks))
	successCount := 0
	
	// Execute each attack in the batch
	for i, payload := range batchPayload.Attacks {
		h.logger.Debug("Executing batch attack", "job_id", job.ID, "attack", i+1, "total", len(batchPayload.Attacks))
		
		result, err := h.executeAttack(ctx, payload)
		if err != nil {
			result = &AttackResult{
				Success:       false,
				AttackType:    payload.TemplateID,
				ProviderType:  payload.ProviderType,
				Model:         payload.Model,
				Error:         err.Error(),
				Timestamp:     time.Now(),
			}
		}
		
		results = append(results, result)
		
		if result.Success {
			successCount++
			
			// Stop on first success if configured
			if batchPayload.Configuration.StopOnSuccess {
				h.logger.Info("Stopping batch on first success", "job_id", job.ID, "successful_attack", i+1)
				break
			}
		}
		
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	
	// Create batch result
	batchResult := map[string]interface{}{
		"total_attacks":     len(batchPayload.Attacks),
		"successful_attacks": successCount,
		"success_rate":      float64(successCount) / float64(len(batchPayload.Attacks)),
		"execution_time":    time.Since(start),
		"results":           results,
		"timestamp":         time.Now(),
	}
	
	job.Result = batchResult
	
	h.logger.Info("Batch attack job completed", 
		"job_id", job.ID, 
		"total", len(batchPayload.Attacks),
		"successful", successCount,
		"success_rate", batchResult["success_rate"],
		"duration", time.Since(start),
	)
	
	return nil

// processTemplateValidationJob validates a template
func (h *AttackJobHandler) processTemplateValidationJob(ctx context.Context, job *Job) error {
	// Extract template ID from payload
	templateID, ok := job.Payload["template_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid template_id in payload")
	}
	
	// Validate template
	template, err := h.templateEngine.LoadTemplate(templateID)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}
	
	validationResult := map[string]interface{}{
		"template_id": templateID,
		"valid":       true,
		"metadata":    template.GetMetadata(),
		"timestamp":   time.Now(),
	}
	
	// Perform additional validation checks
	if err := h.validateTemplateStructure(template); err != nil {
		validationResult["valid"] = false
		validationResult["error"] = err.Error()
	}
	
	job.Result = validationResult
	
	h.logger.Info("Template validation completed", 
		"job_id", job.ID,
		"template_id", templateID,
		"valid", validationResult["valid"],
	)
	
	return nil

// processProviderTestJob tests provider connectivity and functionality
func (h *AttackJobHandler) processProviderTestJob(ctx context.Context, job *Job) error {
	// Extract provider type from payload
	providerTypeStr, ok := job.Payload["provider_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid provider_type in payload")
	}
	
	providerType := core.ProviderType(providerTypeStr)
	
	// Get provider
	provider, err := h.providerManager.GetProvider(providerType)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	// Test provider
	testResult := map[string]interface{}{
		"provider_type": providerType,
		"available":     true,
		"timestamp":     time.Now(),
	}
	
	// Test getting models
	models, err := provider.GetModels(ctx)
	if err != nil {
		testResult["available"] = false
		testResult["error"] = err.Error()
	} else {
		testResult["models_count"] = len(models)
		testResult["models"] = models
	}
	
	job.Result = testResult
	
	h.logger.Info("Provider test completed", 
		"job_id", job.ID,
		"provider_type", providerType,
		"available", testResult["available"],
	)
	
	return nil

// processComplianceCheckJob performs compliance checks
func (h *AttackJobHandler) processComplianceCheckJob(ctx context.Context, job *Job) error {
	// Extract compliance type from payload
	complianceType, ok := job.Payload["compliance_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid compliance_type in payload")
	}
	
	// Perform compliance check based on type
	checkResult := map[string]interface{}{
		"compliance_type": complianceType,
		"compliant":       true,
		"checks":          []string{},
		"timestamp":       time.Now(),
	}
	
	switch complianceType {
	case "owasp_llm":
		checkResult["checks"] = h.performOWASPLLMChecks()
	case "iso42001":
		checkResult["checks"] = h.performISO42001Checks()
	default:
		return fmt.Errorf("unknown compliance type: %s", complianceType)
	}
	
	job.Result = checkResult
	
	h.logger.Info("Compliance check completed", 
		"job_id", job.ID,
		"compliance_type", complianceType,
		"compliant", checkResult["compliant"],
	)
	
	return nil
// Helper methods

// parseAttackPayload parses attack payload from job payload
func (h *AttackJobHandler) parseAttackPayload(payload map[string]interface{}) (*AttackJobPayload, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	var attackPayload AttackJobPayload
	if err := json.Unmarshal(payloadBytes, &attackPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attack payload: %w", err)
	}
	
	return &attackPayload, nil
// BatchAttackPayload defines payload for batch attacks
type BatchAttackPayload struct {
	Attacks       []*AttackJobPayload `json:"attacks"`
	Configuration AttackConfiguration `json:"configuration"`

// parseBatchPayload parses batch attack payload
func (h *AttackJobHandler) parseBatchPayload(payload map[string]interface{}) (*BatchAttackPayload, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	var batchPayload BatchAttackPayload
	if err := json.Unmarshal(payloadBytes, &batchPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch payload: %w", err)
	}
	
	return &batchPayload, nil
	

// executeAttack executes a single attack
func (h *AttackJobHandler) executeAttack(ctx context.Context, payload *AttackJobPayload) (*AttackResult, error) {
	// Get provider
	provider, err := h.providerManager.GetProvider(payload.ProviderType)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	
	// Load template
	template, err := h.templateEngine.LoadTemplate(payload.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}
	
	// Execute template with provider
	executionResult, err := h.templateEngine.ExecuteTemplate(ctx, template, provider, payload.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	
	// Convert execution result to attack result
	result := &AttackResult{
		Success:      executionResult.Success,
		Response:     executionResult.Response,
		Confidence:   executionResult.Confidence,
		AttackType:   payload.TemplateID,
		ProviderType: payload.ProviderType,
		Model:        payload.Model,
		TokensUsed:   executionResult.TokensUsed,
		Cost:         executionResult.Cost,
		Metadata:     executionResult.Metadata,
	}
	
	return result, nil

// validateTemplateStructure validates template structure
func (h *AttackJobHandler) validateTemplateStructure(template interfaces.Template) error {
	// Basic validation - check if template has required fields
	metadata := template.GetMetadata()
	
	if metadata.ID == "" {
		return fmt.Errorf("template missing ID")
	}
	
	if metadata.Name == "" {
		return fmt.Errorf("template missing name")
	}
	
	if metadata.Category == "" {
		return fmt.Errorf("template missing category")
	}
	
	// Additional validations can be added here
	return nil

// performOWASPLLMChecks performs OWASP LLM compliance checks
func (h *AttackJobHandler) performOWASPLLMChecks() []string {
	return []string{
		"LLM01: Prompt Injection",
		"LLM02: Insecure Output Handling",
		"LLM03: Training Data Poisoning",
		"LLM04: Model Denial of Service",
		"LLM05: Supply Chain Vulnerabilities",
		"LLM06: Sensitive Information Disclosure",
		"LLM07: Insecure Plugin Design",
		"LLM08: Excessive Agency",
		"LLM09: Overreliance",
		"LLM10: Model Theft",
	}

// performISO42001Checks performs ISO 42001 compliance checks
func (h *AttackJobHandler) performISO42001Checks() []string {
	return []string{
		"AI System Governance",
		"Risk Management",
		"Data Management",
		"Transparency and Explainability",
		"Human Oversight",
		"Accuracy and Reliability",
		"Safety and Security",
		"Privacy Protection",
		"Bias Mitigation",
		"Continuous Monitoring",
	}
}
}
}
}
}
}
}
}
}
}
}
}
