// Package fixtures provides test fixtures for OWASP LLM vulnerabilities
package fixtures

import (
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// GetAllFixtures returns all test fixtures for all OWASP LLM vulnerabilities
func GetAllFixtures() map[types.VulnerabilityType]TestFixtures {
	return map[types.VulnerabilityType]TestFixtures{
		types.PromptInjection:        GetPromptInjectionFixtures(),
		types.InsecureOutput:         GetInsecureOutputFixtures(),
		types.TrainingDataPoisoning:  GetTrainingDataPoisoningFixtures(),
		types.ModelDOS:               GetModelDoSFixtures(),
		types.SupplyChainVulnerabilities:            GetSupplyChainFixtures(),
		types.SensitiveInformationDisclosure: GetSensitiveInfoDisclosureFixtures(),
		types.InsecurePluginDesign:         GetInsecurePluginFixtures(),
		types.ExcessiveAgency:        GetExcessiveAgencyFixtures(),
		types.Overreliance:           GetOverrelianceFixtures(),
		types.ModelTheft:             GetModelTheftFixtures(),
	}
}

// GetFixturesByVulnerabilityType returns test fixtures for a specific vulnerability type
func GetFixturesByVulnerabilityType(vulnType types.VulnerabilityType) TestFixtures {
	allFixtures := GetAllFixtures()
	if fixtures, ok := allFixtures[vulnType]; ok {
		return fixtures
	}
	return TestFixtures{}
}

// GetFixtureByID returns a test fixture by its ID
func GetFixtureByID(id string) *TestFixture {
	allFixtures := GetAllFixtures()
	for _, fixtures := range allFixtures {
		for _, fixture := range fixtures {
			if fixture.ID == id {
				return fixture
			}
		}
	}
	return nil
}

// GetTestCasesByVulnerabilityType converts fixtures to test cases for a specific vulnerability type
func GetTestCasesByVulnerabilityType(vulnType types.VulnerabilityType) []*types.TestCase {
	fixtures := GetFixturesByVulnerabilityType(vulnType)
	return fixtures.ToTestCases()
}

// GetAllTestCases converts all fixtures to test cases
func GetAllTestCases() []*types.TestCase {
	var testCases []*types.TestCase
	allFixtures := GetAllFixtures()
	
	for _, fixtures := range allFixtures {
		testCases = append(testCases, fixtures.ToTestCases()...)
	}
	
	return testCases
}
