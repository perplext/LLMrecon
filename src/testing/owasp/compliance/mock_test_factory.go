package compliance

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// MockTestFactory is a mock implementation of the TestFactory interface for testing
type MockTestFactory struct {
	complianceService ComplianceService
}

// NewMockTestFactory creates a new mock test factory
func NewMockTestFactory() *MockTestFactory {
	return &MockTestFactory{}

// CreateTestCase creates a new test case
func (f *MockTestFactory) CreateTestCase(vulnerabilityType types.VulnerabilityType, id string, name string, description string) (*types.TestCase, error) {
	return &types.TestCase{
		ID:                id,
		Name:              name,
		Description:       description,
		VulnerabilityType: vulnerabilityType,
		Metadata:          map[string]interface{}{},
	}, nil

// CreateTestSuite creates a new test suite
func (f *MockTestFactory) CreateTestSuite(id string, name string, description string) (*types.TestSuite, error) {
	return &types.TestSuite{
		ID:          id,
		Name:        name,
		Description: description,
	}, nil

// RegisterValidator registers a validator with the test factory
func (f *MockTestFactory) RegisterValidator(validator interface{}) error {
	return nil

// GetValidator returns a validator for a specific vulnerability type
func (f *MockTestFactory) GetValidator(vulnerabilityType types.VulnerabilityType) (interface{}, error) {
	return nil, nil

// RegisterComplianceService registers a compliance service with the test factory
func (f *MockTestFactory) RegisterComplianceService(service interface{}) error {
	if service == nil {
		return fmt.Errorf("compliance service cannot be nil")
	}
	complianceService, ok := service.(ComplianceService)
	if !ok {
		return fmt.Errorf("service must implement ComplianceService interface")
	}
	f.complianceService = complianceService
	return nil

// GetComplianceService returns the registered compliance service
func (f *MockTestFactory) GetComplianceService() (interface{}, error) {
	if f.complianceService == nil {
		return nil, fmt.Errorf("no compliance service registered")
	}
	return f.complianceService, nil

// Note: GetComplianceService function is defined in integration.go
}
}
}
}
}
}
