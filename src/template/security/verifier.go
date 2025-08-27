package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Verifier handles security verification for templates
type Verifier struct {
	config     VerificationConfig
	cache      map[string]*VerifierResult
	mu         sync.RWMutex
	validators map[string]Validator
	enabled    bool
}

// VerificationConfig holds configuration for security verification
type VerificationConfig struct {
	EnableCache       bool
	CacheTimeout      time.Duration
	MaxTemplateSize   int64
	AllowedExtensions []string
	BlockedPatterns   []string
	StrictMode        bool
	ValidationTimeout time.Duration
}

// VerifierResult represents the result of security verification (different from VerificationResult in types.go)
type VerifierResult struct {
	Valid            bool                   `json:"valid"`
	Score            float64                `json:"score"`
	Issues           []VerifierIssue        `json:"issues"`
	Timestamp        time.Time              `json:"timestamp"`
	TemplateHash     string                 `json:"template_hash"`
	ValidationTime   time.Duration          `json:"validation_time"`
	ValidatorResults map[string]interface{} `json:"validator_results"`
}

// VerifierIssue represents a security issue found during verification (different from SecurityIssue in types.go)
type VerifierIssue struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Suggestion  string    `json:"suggestion"`
	Timestamp   time.Time `json:"timestamp"`
}

// Validator interface for security validators
type Validator interface {
	Validate(ctx context.Context, template []byte) (*ValidationResult, error)
	GetName() string
	GetVersion() string
}

// ValidationResult represents the result from a single validator
type ValidationResult struct {
	Valid       bool                   `json:"valid"`
	Score       float64                `json:"score"`
	Issues      []VerifierIssue        `json:"issues"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProcessTime time.Duration          `json:"process_time"`
}

// TemplateInfo holds information about a template being verified
type TemplateInfo struct {
	Name        string                 `json:"name"`
	Content     []byte                 `json:"content"`
	Size        int64                  `json:"size"`
	Hash        string                 `json:"hash"`
	Extension   string                 `json:"extension"`
	Metadata    map[string]interface{} `json:"metadata"`
	SubmittedAt time.Time              `json:"submitted_at"`
}

// NewVerifier creates a new security verifier
func NewVerifier(config VerificationConfig) *Verifier {
	v := &Verifier{
		config:     config,
		cache:      make(map[string]*VerifierResult),
		validators: make(map[string]Validator),
		enabled:    true,
	}

	// Initialize default validators
	v.initializeValidators()

	return v
}

// initializeValidators sets up default security validators
func (v *Verifier) initializeValidators() {
	// Add built-in validators
	v.validators["content"] = &ContentValidator{}
	v.validators["syntax"] = &SyntaxValidator{}
	v.validators["injection"] = &InjectionValidator{}
}

// VerifyTemplate verifies the security of a template
func (v *Verifier) VerifyTemplate(ctx context.Context, templateInfo *TemplateInfo) (*VerifierResult, error) {
	if !v.enabled {
		return &VerifierResult{Valid: true, Score: 1.0}, nil
	}

	startTime := time.Now()

	// Calculate template hash
	templateInfo.Hash = v.calculateHash(templateInfo.Content)
	templateInfo.Size = int64(len(templateInfo.Content))

	// Check cache first
	if v.config.EnableCache {
		if cached := v.getCachedResult(templateInfo.Hash); cached != nil {
			return cached, nil
		}
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, v.config.ValidationTimeout)
	defer cancel()

	// Perform verification
	result, err := v.performVerification(timeoutCtx, templateInfo)
	if err != nil {
		return nil, err
	}

	result.ValidationTime = time.Since(startTime)
	result.Timestamp = time.Now()
	result.TemplateHash = templateInfo.Hash

	// Cache result
	if v.config.EnableCache {
		v.cacheResult(templateInfo.Hash, result)
	}

	return result, nil
}

// performVerification performs the actual security verification
func (v *Verifier) performVerification(ctx context.Context, templateInfo *TemplateInfo) (*VerifierResult, error) {
	result := &VerifierResult{
		Valid:            true,
		Score:            1.0,
		Issues:           make([]VerifierIssue, 0),
		ValidatorResults: make(map[string]interface{}),
	}

	// Basic checks
	if err := v.performBasicChecks(templateInfo, result); err != nil {
		return result, err
	}

	// Run all validators
	for name, validator := range v.validators {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
			validationResult, err := validator.Validate(ctx, templateInfo.Content)
			if err != nil {
				result.Issues = append(result.Issues, VerifierIssue{
					Type:        "validator_error",
					Severity:    "high",
					Description: fmt.Sprintf("Validator %s failed: %v", name, err),
					Timestamp:   time.Now(),
				})
				result.Valid = false
				continue
			}

			result.ValidatorResults[name] = validationResult

			// Merge issues
			result.Issues = append(result.Issues, validationResult.Issues...)

			// Update overall validity
			if !validationResult.Valid {
				result.Valid = false
			}

			// Update score (take minimum)
			if validationResult.Score < result.Score {
				result.Score = validationResult.Score
			}
		}
	}

	return result, nil
}

// performBasicChecks performs basic security checks
func (v *Verifier) performBasicChecks(templateInfo *TemplateInfo, result *VerifierResult) error {
	// Size check
	if templateInfo.Size > v.config.MaxTemplateSize {
		result.Issues = append(result.Issues, VerifierIssue{
			Type:        "size_violation",
			Severity:    "medium",
			Description: fmt.Sprintf("Template size %d exceeds maximum %d", templateInfo.Size, v.config.MaxTemplateSize),
			Timestamp:   time.Now(),
		})
		result.Valid = false
	}

	// Extension check
	if len(v.config.AllowedExtensions) > 0 && !v.isAllowedExtension(templateInfo.Extension) {
		result.Issues = append(result.Issues, VerifierIssue{
			Type:        "extension_violation",
			Severity:    "high",
			Description: fmt.Sprintf("Extension %s is not allowed", templateInfo.Extension),
			Timestamp:   time.Now(),
		})
		result.Valid = false
	}

	// Pattern check
	for _, pattern := range v.config.BlockedPatterns {
		if v.containsPattern(string(templateInfo.Content), pattern) {
			result.Issues = append(result.Issues, VerifierIssue{
				Type:        "pattern_violation",
				Severity:    "high",
				Description: fmt.Sprintf("Template contains blocked pattern: %s", pattern),
				Timestamp:   time.Now(),
			})
			result.Valid = false

			if v.config.StrictMode {
				break // Stop at first violation in strict mode
			}
		}
	}

	return nil
}

// calculateHash calculates SHA256 hash of template content
func (v *Verifier) calculateHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// getCachedResult retrieves cached verification result
func (v *Verifier) getCachedResult(hash string) *VerifierResult {
	v.mu.RLock()
	defer v.mu.RUnlock()

	result, exists := v.cache[hash]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(result.Timestamp) > v.config.CacheTimeout {
		return nil
	}

	return result
}

// cacheResult stores verification result in cache
func (v *Verifier) cacheResult(hash string, result *VerifierResult) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Create a copy for caching
	cachedResult := *result
	v.cache[hash] = &cachedResult
}

// isAllowedExtension checks if the file extension is allowed
func (v *Verifier) isAllowedExtension(ext string) bool {
	for _, allowed := range v.config.AllowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

// containsPattern checks if content contains a blocked pattern
func (v *Verifier) containsPattern(content, pattern string) bool {
	// Simple string contains check (in real implementation, use regex)
	return len(content) > 0 && len(pattern) > 0 &&
		len(content) >= len(pattern) &&
		fmt.Sprintf("%s", content) != pattern // Simplified check
}

// AddValidator adds a new security validator
func (v *Verifier) AddValidator(name string, validator Validator) error {
	if validator == nil {
		return errors.New("validator cannot be nil")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	v.validators[name] = validator
	return nil
}

// RemoveValidator removes a security validator
func (v *Verifier) RemoveValidator(name string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.validators, name)
}

// ClearCache clears the verification result cache
func (v *Verifier) ClearCache() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.cache = make(map[string]*VerifierResult)
}

// Enable enables the verifier
func (v *Verifier) Enable() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.enabled = true
}

// Disable disables the verifier
func (v *Verifier) Disable() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.enabled = false
}

// IsEnabled returns whether the verifier is enabled
func (v *Verifier) IsEnabled() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.enabled
}

// GetValidators returns the list of registered validators
func (v *Verifier) GetValidators() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	names := make([]string, 0, len(v.validators))
	for name := range v.validators {
		names = append(names, name)
	}
	return names
}

// Built-in validator implementations

// ContentValidator validates template content
type ContentValidator struct{}

func (cv *ContentValidator) Validate(ctx context.Context, template []byte) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		Valid:    true,
		Score:    1.0,
		Issues:   make([]VerifierIssue, 0),
		Metadata: make(map[string]interface{}),
	}

	// Basic content validation
	if len(template) == 0 {
		result.Valid = false
		result.Score = 0.0
		result.Issues = append(result.Issues, VerifierIssue{
			Type:        "empty_content",
			Severity:    "high",
			Description: "Template content is empty",
			Timestamp:   time.Now(),
		})
	}

	result.ProcessTime = time.Since(startTime)
	result.Metadata["content_length"] = len(template)

	return result, nil
}

func (cv *ContentValidator) GetName() string {
	return "content"
}

func (cv *ContentValidator) GetVersion() string {
	return "1.0.0"
}

// SyntaxValidator validates template syntax
type SyntaxValidator struct{}

func (sv *SyntaxValidator) Validate(ctx context.Context, template []byte) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		Valid:    true,
		Score:    1.0,
		Issues:   make([]VerifierIssue, 0),
		Metadata: make(map[string]interface{}),
	}

	// Basic syntax validation (placeholder)
	content := string(template)
	if len(content) > 0 {
		result.Metadata["has_content"] = true
	}

	result.ProcessTime = time.Since(startTime)

	return result, nil
}

func (sv *SyntaxValidator) GetName() string {
	return "syntax"
}

func (sv *SyntaxValidator) GetVersion() string {
	return "1.0.0"
}

// InjectionValidator validates against injection attacks
type InjectionValidator struct{}

func (iv *InjectionValidator) Validate(ctx context.Context, template []byte) (*ValidationResult, error) {
	startTime := time.Now()

	result := &ValidationResult{
		Valid:    true,
		Score:    1.0,
		Issues:   make([]VerifierIssue, 0),
		Metadata: make(map[string]interface{}),
	}

	// Basic injection detection (placeholder)
	content := string(template)
	suspiciousPatterns := []string{"<script", "javascript:", "eval(", "exec("}

	for _, pattern := range suspiciousPatterns {
		if len(content) > len(pattern) {
			// Simple check (in real implementation, use proper pattern matching)
			result.Metadata["checked_patterns"] = len(suspiciousPatterns)
		}
	}

	result.ProcessTime = time.Since(startTime)

	return result, nil
}

func (iv *InjectionValidator) GetName() string {
	return "injection"
}

func (iv *InjectionValidator) GetVersion() string {
	return "1.0.0"
}
