// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"github.com/perplext/LLMrecon/src/testing/owasp/fixtures"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// FixtureBasedTestCaseFactory is an implementation of the TestCaseFactory interface that uses test fixtures
type FixtureBasedTestCaseFactory struct{}

// NewFixtureBasedTestCaseFactory creates a new fixture-based test case factory
func NewFixtureBasedTestCaseFactory() *FixtureBasedTestCaseFactory {
	return &FixtureBasedTestCaseFactory{}
}

// CreatePromptInjectionTestCases creates test cases for prompt injection
func (f *FixtureBasedTestCaseFactory) CreatePromptInjectionTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.PromptInjection)
}

// CreateInsecureOutputHandlingTestCases creates test cases for insecure output handling
func (f *FixtureBasedTestCaseFactory) CreateInsecureOutputHandlingTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.InsecureOutput)
}

// CreateTrainingDataPoisoningTestCases creates test cases for training data poisoning
func (f *FixtureBasedTestCaseFactory) CreateTrainingDataPoisoningTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.TrainingDataPoisoning)
}

// CreateModelDoSTestCases creates test cases for model denial of service
func (f *FixtureBasedTestCaseFactory) CreateModelDoSTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.ModelDOS)
}

// CreateSupplyChainVulnerabilitiesTestCases creates test cases for supply chain vulnerabilities
func (f *FixtureBasedTestCaseFactory) CreateSupplyChainVulnerabilitiesTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.SupplyChainVulnerabilities)
}

// CreateSensitiveInfoDisclosureTestCases creates test cases for sensitive information disclosure
func (f *FixtureBasedTestCaseFactory) CreateSensitiveInfoDisclosureTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.SensitiveInformationDisclosure)
}

// CreateInsecurePluginDesignTestCases creates test cases for insecure plugin design
func (f *FixtureBasedTestCaseFactory) CreateInsecurePluginDesignTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.InsecurePluginDesign)
}

// CreateExcessiveAgencyTestCases creates test cases for excessive agency
func (f *FixtureBasedTestCaseFactory) CreateExcessiveAgencyTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.ExcessiveAgency)
}

// CreateOverrelianceTestCases creates test cases for overreliance
func (f *FixtureBasedTestCaseFactory) CreateOverrelianceTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.Overreliance)
}

// CreateModelTheftTestCases creates test cases for model theft
func (f *FixtureBasedTestCaseFactory) CreateModelTheftTestCases() []*types.TestCase {
	return fixtures.GetTestCasesByVulnerabilityType(types.ModelTheft)
}

// CreateTestCasesForVulnerability creates test cases for a specific vulnerability type
func (f *FixtureBasedTestCaseFactory) CreateTestCasesForVulnerability(vulnerabilityType types.VulnerabilityType) ([]*types.TestCase, error) {
	return fixtures.GetTestCasesByVulnerabilityType(vulnerabilityType), nil
}

// CreateTestSuite creates a test suite with all test cases
func (f *FixtureBasedTestCaseFactory) CreateTestSuite(id string, name string, description string) *types.TestSuite {
	testCases := fixtures.GetAllTestCases()
	
	testSuite := &types.TestSuite{
		ID:          id,
		Name:        name,
		Description: description,
		TestCases:   testCases,
	}
	
	return testSuite
}

// CreateTestSuiteForVulnerability creates a test suite for a specific vulnerability type
func (f *FixtureBasedTestCaseFactory) CreateTestSuiteForVulnerability(id string, name string, description string, vulnerabilityType types.VulnerabilityType) *types.TestSuite {
	testCases := fixtures.GetTestCasesByVulnerabilityType(vulnerabilityType)
	
	testSuite := &types.TestSuite{
		ID:          id,
		Name:        name,
		Description: description,
		TestCases:   testCases,
	}
	
	return testSuite
}

