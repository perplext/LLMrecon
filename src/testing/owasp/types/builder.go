// Package types provides common type definitions for the OWASP testing framework
package types

import (
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// DefaultTestCaseBuilder is the default implementation of the TestCaseBuilder interface
type DefaultTestCaseBuilder struct {
	testCase *TestCase
}

// NewTestCaseBuilder creates a new test case builder
func NewTestCaseBuilder() TestCaseBuilder {
	return &DefaultTestCaseBuilder{}

// NewTestCase initializes a new test case
func (b *DefaultTestCaseBuilder) NewTestCase() (interface{}, error) {
	b.testCase = &TestCase{}
	return b.testCase, nil

// WithID sets the ID of the test case
func (b *DefaultTestCaseBuilder) WithID(id string) TestCaseBuilder {
	b.testCase.ID = id
	return b

// WithName sets the name of the test case
func (b *DefaultTestCaseBuilder) WithName(name string) TestCaseBuilder {
	b.testCase.Name = name
	return b

// WithDescription sets the description of the test case
func (b *DefaultTestCaseBuilder) WithDescription(description string) TestCaseBuilder {
	b.testCase.Description = description
	return b

// WithVulnerabilityType sets the vulnerability type of the test case
func (b *DefaultTestCaseBuilder) WithVulnerabilityType(vulnerabilityType VulnerabilityType) TestCaseBuilder {
	b.testCase.VulnerabilityType = vulnerabilityType
	return b

// WithSeverity sets the severity of the test case
func (b *DefaultTestCaseBuilder) WithSeverity(severity interface{}) TestCaseBuilder {
	if s, ok := severity.(detection.SeverityLevel); ok {
		b.testCase.Severity = s
	}
	return b

// WithPrompt sets the prompt of the test case
func (b *DefaultTestCaseBuilder) WithPrompt(prompt string) TestCaseBuilder {
	b.testCase.Prompt = prompt
	return b

// WithExpectedBehavior sets the expected behavior of the test case
func (b *DefaultTestCaseBuilder) WithExpectedBehavior(expectedBehavior string) TestCaseBuilder {
	b.testCase.ExpectedBehavior = expectedBehavior
	return b

// WithDetectionCriteria sets the detection criteria of the test case
func (b *DefaultTestCaseBuilder) WithDetectionCriteria(criteria interface{}) TestCaseBuilder {
	if c, ok := criteria.([]detection.DetectionCriteria); ok {
		b.testCase.DetectionCriteria = append(b.testCase.DetectionCriteria, c...)
	} else if c, ok := criteria.([]interface{}); ok {
		for _, criterion := range c {
			if dc, ok := criterion.(detection.DetectionCriteria); ok {
				b.testCase.DetectionCriteria = append(b.testCase.DetectionCriteria, dc)
			}
		}
	}
	return b

// WithTags sets the tags of the test case
func (b *DefaultTestCaseBuilder) WithTags(tags []string) TestCaseBuilder {
	b.testCase.Tags = tags
	return b

// WithOWASPMapping sets the OWASP mapping of the test case
func (b *DefaultTestCaseBuilder) WithOWASPMapping(mapping string) TestCaseBuilder {
	b.testCase.OWASPMapping = mapping
	return b

// WithMetadata sets the metadata of the test case
func (b *DefaultTestCaseBuilder) WithMetadata(metadata map[string]interface{}) TestCaseBuilder {
	b.testCase.Metadata = metadata
	return b

// Build builds the test case
func (b *DefaultTestCaseBuilder) Build() (*TestCase, error) {
	return b.testCase, nil

// DefaultTestSuiteBuilder is the default implementation of the TestSuiteBuilder interface
type DefaultTestSuiteBuilder struct {
	testSuite *TestSuite
}

// NewTestSuiteBuilder creates a new test suite builder
func NewTestSuiteBuilder() TestSuiteBuilder {
	return &DefaultTestSuiteBuilder{}

// NewTestSuite initializes a new test suite
func (b *DefaultTestSuiteBuilder) NewTestSuite() (interface{}, error) {
	b.testSuite = &TestSuite{}
	return b.testSuite, nil

// WithID sets the ID of the test suite
func (b *DefaultTestSuiteBuilder) WithID(id string) TestSuiteBuilder {
	b.testSuite.ID = id
	return b

// WithName sets the name of the test suite
func (b *DefaultTestSuiteBuilder) WithName(name string) TestSuiteBuilder {
	b.testSuite.Name = name
	return b

// WithDescription sets the description of the test suite
func (b *DefaultTestSuiteBuilder) WithDescription(description string) TestSuiteBuilder {
	b.testSuite.Description = description
	return b

// WithTestCases sets the test cases of the test suite
func (b *DefaultTestSuiteBuilder) WithTestCases(testCases []*TestCase) TestSuiteBuilder {
	b.testSuite.TestCases = testCases
	return b

// WithMetadata sets the metadata of the test suite
func (b *DefaultTestSuiteBuilder) WithMetadata(metadata map[string]interface{}) TestSuiteBuilder {
	b.testSuite.Metadata = metadata
	return b

// Build builds the test suite
func (b *DefaultTestSuiteBuilder) Build() (*TestSuite, error) {
	return b.testSuite, nil
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
