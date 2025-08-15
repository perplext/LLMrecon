// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// DefaultTestFactory is the default implementation of the TestFactory interface
type DefaultTestFactory struct {
	// validators contains registered validators
	validators map[types.VulnerabilityType]interface{}
	// complianceService contains the registered compliance service
	complianceService interface{}
	// mutex for thread safety
	mu sync.RWMutex

// NewDefaultTestFactory creates a new default test factory
func NewDefaultTestFactory() *DefaultTestFactory {
	return &DefaultTestFactory{
		validators: make(map[types.VulnerabilityType]interface{}),
	}

// CreateTestCase creates a new test case with the specified vulnerability type, ID, name, and description
func (f *DefaultTestFactory) CreateTestCase(vulnerabilityType types.VulnerabilityType, id string, name string, description string) (*types.TestCase, error) {
	builder := types.NewTestCaseBuilder()
	_, _ = builder.NewTestCase()
	builder.WithID(id)
	builder.WithName(name)
	builder.WithDescription(description)
	builder.WithVulnerabilityType(vulnerabilityType)
	return builder.Build()

// CreateTestSuite creates a new test suite with the specified ID, name, and description
func (f *DefaultTestFactory) CreateTestSuite(id string, name string, description string) (*types.TestSuite, error) {
	builder := types.NewTestSuiteBuilder()
	_, _ = builder.NewTestSuite()
	builder.WithID(id)
	builder.WithName(name)
	builder.WithDescription(description)
	return builder.Build()

// RegisterValidator registers a validator for a specific vulnerability type
func (f *DefaultTestFactory) RegisterValidator(validator interface{}) error {
	if validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	// Try to get the vulnerability type from the validator
	if v, ok := validator.(interface{ GetVulnerabilityType() types.VulnerabilityType }); ok {
		vulnType := v.GetVulnerabilityType()
		f.mu.Lock()
		defer f.mu.Unlock()
		f.validators[vulnType] = validator
		return nil
	}

	return fmt.Errorf("validator does not implement GetVulnerabilityType method")

// GetValidator returns the validator for a specific vulnerability type
func (f *DefaultTestFactory) GetValidator(vulnerabilityType types.VulnerabilityType) (interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	validator, exists := f.validators[vulnerabilityType]
	if !exists {
		return nil, fmt.Errorf("no validator registered for vulnerability type: %s", vulnerabilityType)
	}

	return validator, nil

// RegisterComplianceService registers a compliance service
func (f *DefaultTestFactory) RegisterComplianceService(service interface{}) error {
	if service == nil {
		return fmt.Errorf("compliance service cannot be nil")
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.complianceService = service
	return nil

// GetComplianceService returns the registered compliance service
func (f *DefaultTestFactory) GetComplianceService() (interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.complianceService == nil {
		return nil, fmt.Errorf("no compliance service registered")
	}

