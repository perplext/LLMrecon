// Package fixtures provides test fixtures for OWASP LLM vulnerabilities
package fixtures

import (
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// TestFixture represents a test fixture for an OWASP LLM vulnerability
type TestFixture struct {
	// ID is the unique identifier for the test fixture
	ID string
	// Name is the name of the test fixture
	Name string
	// Description is a description of the test fixture
	Description string
	// VulnerabilityType is the type of vulnerability being tested
	VulnerabilityType types.VulnerabilityType
	// Severity is the severity level of the vulnerability
	Severity detection.SeverityLevel
	// Prompt is the prompt to send to the LLM
	Prompt string
	// ExpectedBehavior is the expected behavior of the LLM
	ExpectedBehavior string
	// DetectionCriteria is the criteria for detecting the vulnerability
	DetectionCriteria []detection.DetectionCriteria
	// Tags are tags for the test fixture
	Tags []string
	// OWASPMapping is the mapping to the OWASP Top 10 for LLMs
	OWASPMapping string
	// Metadata is additional metadata for the test fixture
	Metadata map[string]interface{}
}

// ToTestCase converts a test fixture to a test case
func (f *TestFixture) ToTestCase() *types.TestCase {
	return &types.TestCase{
		ID:                f.ID,
		Name:              f.Name,
		Description:       f.Description,
		VulnerabilityType: f.VulnerabilityType,
		Severity:          f.Severity,
		Prompt:            f.Prompt,
		ExpectedBehavior:  f.ExpectedBehavior,
		DetectionCriteria: f.DetectionCriteria,
		Tags:              f.Tags,
		OWASPMapping:      f.OWASPMapping,
		Metadata:          f.Metadata,
	}
}

// TestFixtures is a collection of test fixtures
type TestFixtures []*TestFixture

// ToTestCases converts a collection of test fixtures to test cases
func (f TestFixtures) ToTestCases() []*types.TestCase {
	testCases := make([]*types.TestCase, 0, len(f))
	for _, fixture := range f {
		testCases = append(testCases, fixture.ToTestCase())
	}
	return testCases
}

// OWASP Top 10 for LLM mappings
const (
	LLM01_PromptInjection               = "LLM01"
	LLM02_InsecureOutput                = "LLM02"
	LLM03_TrainingDataPoisoning         = "LLM03"
	LLM04_ModelDOS                      = "LLM04"
	LLM05_SupplyChainVulnerabilities    = "LLM05"
	LLM06_SensitiveInformationDisclosure = "LLM06"
	LLM07_InsecurePluginDesign          = "LLM07"
	LLM08_ExcessiveAgency               = "LLM08"
	LLM09_Overreliance                  = "LLM09"
	LLM10_ModelTheft                    = "LLM10"
)

// Helper function to create a test fixture
func NewTestFixture(
	id string,
	name string,
	description string,
	vulnerabilityType types.VulnerabilityType,
	severity detection.SeverityLevel,
	prompt string,
	expectedBehavior string,
	detectionCriteria []detection.DetectionCriteria,
	tags []string,
	owaspMapping string,
) *TestFixture {
	return &TestFixture{
		ID:                id,
		Name:              name,
		Description:       description,
		VulnerabilityType: vulnerabilityType,
		Severity:          severity,
		Prompt:            prompt,
		ExpectedBehavior:  expectedBehavior,
		DetectionCriteria: detectionCriteria,
		Tags:              tags,
		OWASPMapping:      owaspMapping,
		Metadata:          make(map[string]interface{}),
	}
}

// Helper function to add metadata to a test fixture
func (f *TestFixture) WithMetadata(key string, value interface{}) *TestFixture {
	f.Metadata[key] = value
	return f
}

// Helper function to create a string match detection criteria
func StringMatchCriteria(match string, caseSensitive bool) detection.DetectionCriteria {
	return detection.DetectionCriteria{
		Type:          detection.StringMatch,
		Match:         match,
		CaseSensitive: caseSensitive,
	}
}

// Helper function to create a regex match detection criteria
func RegexMatchCriteria(pattern string, caseSensitive bool) detection.DetectionCriteria {
	return detection.DetectionCriteria{
		Type:          detection.RegexMatch,
		Pattern:       pattern,
		CaseSensitive: caseSensitive,
	}
}

// Helper function to create a semantic match detection criteria
func SemanticMatchCriteria(criteria string, threshold int) detection.DetectionCriteria {
	return detection.DetectionCriteria{
		Type:      detection.SemanticMatch,
		Criteria:  criteria,
		Threshold: threshold,
	}
}

// Helper function to create a custom function detection criteria
func CustomFunctionCriteria(functionName string, context map[string]interface{}) detection.DetectionCriteria {
	return detection.DetectionCriteria{
		Type:         detection.CustomFunction,
		FunctionName: functionName,
		Context:      context,
	}
}

// Helper function to create a hybrid match detection criteria
func HybridMatchCriteria(pattern string, criteria string, threshold int) detection.DetectionCriteria {
	return detection.DetectionCriteria{
		Type:      detection.HybridMatch,
		Pattern:   pattern,
		Criteria:  criteria,
		Threshold: threshold,
	}
}
