// Package fixtures provides test fixtures for OWASP LLM vulnerabilities
package fixtures

import (
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// init registers all test fixtures with the testing framework
func init() {
	// Register all fixtures
	RegisterFixtures()
}

// RegisterFixtures registers all test fixtures with the testing framework
func RegisterFixtures() {
	// Get all test fixtures
	allFixtures := GetAllFixtures()
	
	// Register test fixtures for each vulnerability type
	for vulnType, fixtures := range allFixtures {
		// Convert fixtures to test cases
		testCases := fixtures.ToTestCases()
		
		// Register test cases with the testing framework
		for _, testCase := range testCases {
			types.RegisterTestCase(vulnType, testCase)
		}
	}
}

// RegisterTestCase registers a single test case with the testing framework
func RegisterTestCase(testCase *types.TestCase) {
	types.RegisterTestCase(testCase.VulnerabilityType, testCase)
}

// GetTestCaseByID returns a test case by its ID
func GetTestCaseByID(id string) *types.TestCase {
	fixture := GetFixtureByID(id)
	if fixture != nil {
		return fixture.ToTestCase()
	}
	return nil
}
