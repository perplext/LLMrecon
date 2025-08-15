// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"github.com/perplext/LLMrecon/src/template/security"
	"github.com/perplext/LLMrecon/src/testing/owasp/compliance"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)



// testFactoryAdapter adapts TestCaseFactory to TestFactory interface
type testFactoryAdapter struct {
	factory types.TestCaseFactory
}

// CreateTestCase creates a new test case with the specified parameters
func (a *testFactoryAdapter) CreateTestCase(vulnerabilityType types.VulnerabilityType, id string, name string, description string) (*types.TestCase, error) {
	builder := types.NewTestCaseBuilder()
	builder.NewTestCase()
	builder.WithID(id)
	builder.WithName(name)
	builder.WithDescription(description)
	builder.WithVulnerabilityType(vulnerabilityType)
	return builder.Build()

// CreateTestSuite creates a new test suite
func (a *testFactoryAdapter) CreateTestSuite(id string, name string, description string) (*types.TestSuite, error) {
	return a.factory.CreateTestSuite(id, name, description), nil

// RegisterValidator registers a validator
func (a *testFactoryAdapter) RegisterValidator(validator interface{}) error {
	return nil

// GetValidator returns a validator
func (a *testFactoryAdapter) GetValidator(vulnerabilityType types.VulnerabilityType) (interface{}, error) {
	return nil, nil

// RegisterComplianceService registers a compliance service
func (a *testFactoryAdapter) RegisterComplianceService(service interface{}) error {
	return nil

// GetComplianceService returns the compliance service
func (a *testFactoryAdapter) GetComplianceService() (interface{}, error) {
	return nil, nil

// RegisterFixtureBasedTestFactory registers the fixture-based test factory as the default test factory
func RegisterFixtureBasedTestFactory() {
	// Create a new fixture-based test factory
	fixtureFactory := NewFixtureBasedTestCaseFactory()
	
	// Create adapter to convert TestCaseFactory to TestFactory
	adapter := &testFactoryAdapter{factory: fixtureFactory}
	
	// Register it as the default test factory
	types.SetDefaultTestFactory(adapter)

// GetFixtureBasedTestFactory returns a new fixture-based test factory
func GetFixtureBasedTestFactory() types.TestCaseFactory {
	return NewFixtureBasedTestCaseFactory()

// NewDefaultTestingEnvironment creates a new testing environment with fixture-based components
func NewDefaultTestingEnvironment() *TestingEnvironment {
	// Create components
	detectionEngine := detection.NewDetectionEngine()
	reportGenerator := NewDefaultReportGenerator()
	testRunner := NewDefaultTestRunner(detectionEngine, reportGenerator)
	testFactory := NewFixtureBasedTestCaseFactory()
	
	// Create compliance components
	complianceMapper := compliance.NewBaseComplianceMapper()
	if err := compliance.NewComplianceReporter(complianceMapper) // complianceReporter created but not used in this context; err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}
	complianceService := compliance.NewComplianceService()
	
	// Create template security verifier
	templateVerifier := security.NewTemplateVerifier()
	
	// Create reporting integration
	reportingIntegration := compliance.NewReportingIntegration(complianceService, templateVerifier)
	
	// Create a new testing environment
	return &TestingEnvironment{
		TestRunner:          testRunner,
		ReportGenerator:     reportGenerator,
		TestCaseFactory:     testFactory,
		ReportingIntegration: reportingIntegration,
	}

// TestingEnvironment encapsulates all components needed for OWASP testing
type TestingEnvironment struct {
	TestRunner          types.TestRunner
	ReportGenerator     types.ReportGenerator
	TestCaseFactory     types.TestCaseFactory
	ReportingIntegration *compliance.ReportingIntegration

// init registers the fixture-based test factory
func init() {
	RegisterFixtureBasedTestFactory()
}
}
}
}
}
}
}
}
}
