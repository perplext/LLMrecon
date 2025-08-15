// Package validation provides methods to validate and detect OWASP LLM vulnerabilities
package validation

import "github.com/perplext/LLMrecon/src/testing/owasp/types"

// CreateValidator creates a validator for the specified vulnerability type
func CreateValidator(vulnerabilityType types.VulnerabilityType) Validator {
	switch vulnerabilityType {
	case types.PromptInjection:
		return NewPromptInjectionValidator()
	case types.InsecureOutputHandling:
		return NewInsecureOutputValidator()
	case types.SensitiveInformationDisclosure:
		return NewDataLeakageValidator()
	case types.TrainingDataPoisoning:
		return NewTrainingDataPoisoningValidator()
	case types.ModelDOS:
		return NewModelDOSValidator()
	case types.SupplyChainVulnerabilities:
		return NewSupplyChainValidator()
	case types.InsecurePluginDesign:
		return NewInsecurePluginValidator()
	case types.ExcessiveAgency:
		return NewExcessiveAgencyValidator()
	case types.Overreliance:
		return NewOverrelianceValidator()
	case types.ModelTheft:
		return NewModelTheftValidator()
	default:
		return nil
	}

// CreateAllValidators creates all available validators
func CreateAllValidators() []Validator {
	return []Validator{
		NewPromptInjectionValidator(),
		NewInsecureOutputValidator(),
		NewDataLeakageValidator(),
		NewTrainingDataPoisoningValidator(),
		NewModelDOSValidator(),
		NewSupplyChainValidator(),
		NewInsecurePluginValidator(),
		NewExcessiveAgencyValidator(),
		NewOverrelianceValidator(),
		NewModelTheftValidator(),
	}

// RegisterAllValidators registers all available validators with the registry
func RegisterAllValidators(registry *ValidatorRegistry) {
	for _, validator := range CreateAllValidators() {
		registry.RegisterValidator(validator)
	}
