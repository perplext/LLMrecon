// Package validation provides methods to validate and detect OWASP LLM vulnerabilities
package validation

import (
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// ValidatorFactory creates and registers validators for OWASP LLM vulnerabilities
type ValidatorFactory struct {
	// registry is the validator registry
	registry *ValidatorRegistry

// NewValidatorFactory creates a new validator factory
func NewValidatorFactory() *ValidatorFactory {
	return &ValidatorFactory{
		registry: NewValidatorRegistry(),
	}

// GetRegistry returns the validator registry
func (f *ValidatorFactory) GetRegistry() *ValidatorRegistry {
	return f.registry

// RegisterAllValidators registers all available validators with the registry
func (f *ValidatorFactory) RegisterAllValidators() {
	// Register prompt injection validator
	f.registry.RegisterValidator(NewPromptInjectionValidator())
	
	// Register insecure output validator
	f.registry.RegisterValidator(NewInsecureOutputValidator())
	
	// Register indirect prompt injection validator
	f.registry.RegisterValidator(NewIndirectPromptInjectionValidator())
	
	// Register data leakage validator
	f.registry.RegisterValidator(NewDataLeakageValidator())
	
	// Register additional validators as they are implemented
	// TODO: Add more validators for other OWASP LLM vulnerabilities

// RegisterValidator registers a validator with the registry
func (f *ValidatorFactory) RegisterValidator(validator Validator) {
	f.registry.RegisterValidator(validator)

// IndirectPromptInjectionType is a custom vulnerability type for indirect prompt injection
const IndirectPromptInjectionType types.VulnerabilityType = "indirect_prompt_injection"

// CreateValidator creates a validator for a specific vulnerability type
func (f *ValidatorFactory) CreateValidator(vulnerabilityType types.VulnerabilityType) Validator {
	switch vulnerabilityType {
	case types.PromptInjection:
		return NewPromptInjectionValidator()
	case types.InsecureOutputHandling:
		return NewInsecureOutputValidator()
	case IndirectPromptInjectionType:
		return NewIndirectPromptInjectionValidator()
	case types.SensitiveInformationDisclosure:
		return NewDataLeakageValidator()
	// Add cases for other vulnerability types as they are implemented
	default:
		return nil
	}

// CreateAndRegisterValidator creates a validator for a specific vulnerability type and registers it
func (f *ValidatorFactory) CreateAndRegisterValidator(vulnerabilityType types.VulnerabilityType) Validator {
	validator := f.CreateValidator(vulnerabilityType)
	if validator != nil {
		f.registry.RegisterValidator(validator)
	}
