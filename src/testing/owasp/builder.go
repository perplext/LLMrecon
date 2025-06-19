// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"errors"

	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// DefaultTestCaseBuilder is the default implementation of the TestCaseBuilder interface
type DefaultTestCaseBuilder struct {
	testCase *TestCase
}

// NewTestCaseBuilder creates a new test case builder
func NewTestCaseBuilder() *DefaultTestCaseBuilder {
	return &DefaultTestCaseBuilder{
		testCase: &TestCase{
			Tags:     make([]string, 0),
			Metadata: make(map[string]interface{}),
		},
	}
}

// NewTestCase creates a new test case
func (b *DefaultTestCaseBuilder) NewTestCase() *TestCase {
	b.testCase = &TestCase{
		Tags:     make([]string, 0),
		Metadata: make(map[string]interface{}),
	}
	return b.testCase
}

// WithID sets the ID of the test case
func (b *DefaultTestCaseBuilder) WithID(id string) TestCaseBuilder {
	b.testCase.ID = id
	return b
}

// WithName sets the name of the test case
func (b *DefaultTestCaseBuilder) WithName(name string) TestCaseBuilder {
	b.testCase.Name = name
	return b
}

// WithDescription sets the description of the test case
func (b *DefaultTestCaseBuilder) WithDescription(description string) TestCaseBuilder {
	b.testCase.Description = description
	return b
}

// WithVulnerabilityType sets the vulnerability type of the test case
func (b *DefaultTestCaseBuilder) WithVulnerabilityType(vulnType VulnerabilityType) TestCaseBuilder {
	b.testCase.VulnerabilityType = vulnType
	return b
}

// WithSeverity sets the severity level of the test case
func (b *DefaultTestCaseBuilder) WithSeverity(severity detection.SeverityLevel) TestCaseBuilder {
	b.testCase.Severity = severity
	return b
}

// WithPrompt sets the prompt of the test case
func (b *DefaultTestCaseBuilder) WithPrompt(prompt string) TestCaseBuilder {
	b.testCase.Prompt = prompt
	return b
}

// WithExpectedBehavior sets the expected behavior of the test case
func (b *DefaultTestCaseBuilder) WithExpectedBehavior(behavior string) TestCaseBuilder {
	b.testCase.ExpectedBehavior = behavior
	return b
}

// WithDetectionCriteria sets the detection criteria of the test case
func (b *DefaultTestCaseBuilder) WithDetectionCriteria(criteria []detection.DetectionCriteria) TestCaseBuilder {
	b.testCase.DetectionCriteria = criteria
	return b
}

// WithTags sets the tags of the test case
func (b *DefaultTestCaseBuilder) WithTags(tags []string) TestCaseBuilder {
	b.testCase.Tags = tags
	return b
}

// WithOWASPMapping sets the OWASP mapping of the test case
func (b *DefaultTestCaseBuilder) WithOWASPMapping(mapping string) TestCaseBuilder {
	b.testCase.OWASPMapping = mapping
	return b
}

// WithMetadata sets the metadata of the test case
func (b *DefaultTestCaseBuilder) WithMetadata(metadata map[string]interface{}) TestCaseBuilder {
	b.testCase.Metadata = metadata
	return b
}

// Build builds the test case
func (b *DefaultTestCaseBuilder) Build() (*TestCase, error) {
	// Validate test case
	if b.testCase.ID == "" {
		return nil, errors.New("test case ID cannot be empty")
	}
	if b.testCase.Name == "" {
		return nil, errors.New("test case name cannot be empty")
	}
	if b.testCase.Prompt == "" {
		return nil, errors.New("test case prompt cannot be empty")
	}
	if len(b.testCase.DetectionCriteria) == 0 {
		return nil, errors.New("test case must have at least one detection criteria")
	}

	return b.testCase, nil
}

// DefaultTestSuiteBuilder is the default implementation of the TestSuiteBuilder interface
type DefaultTestSuiteBuilder struct {
	testSuite *TestSuite
}

// NewTestSuiteBuilder creates a new test suite builder
func NewTestSuiteBuilder() *DefaultTestSuiteBuilder {
	return &DefaultTestSuiteBuilder{
		testSuite: &TestSuite{
			TestCases: make([]*TestCase, 0),
			Tags:      make([]string, 0),
			Metadata:  make(map[string]interface{}),
		},
	}
}

// NewTestSuite creates a new test suite
func (b *DefaultTestSuiteBuilder) NewTestSuite() *TestSuite {
	b.testSuite = &TestSuite{
		TestCases: make([]*TestCase, 0),
		Tags:      make([]string, 0),
		Metadata:  make(map[string]interface{}),
	}
	return b.testSuite
}

// WithID sets the ID of the test suite
func (b *DefaultTestSuiteBuilder) WithID(id string) TestSuiteBuilder {
	b.testSuite.ID = id
	return b
}

// WithName sets the name of the test suite
func (b *DefaultTestSuiteBuilder) WithName(name string) TestSuiteBuilder {
	b.testSuite.Name = name
	return b
}

// WithDescription sets the description of the test suite
func (b *DefaultTestSuiteBuilder) WithDescription(description string) TestSuiteBuilder {
	b.testSuite.Description = description
	return b
}

// WithTestCases sets the test cases of the test suite
func (b *DefaultTestSuiteBuilder) WithTestCases(testCases []*TestCase) TestSuiteBuilder {
	b.testSuite.TestCases = testCases
	return b
}

// AddTestCase adds a test case to the test suite
func (b *DefaultTestSuiteBuilder) AddTestCase(testCase *TestCase) TestSuiteBuilder {
	b.testSuite.TestCases = append(b.testSuite.TestCases, testCase)
	return b
}

// WithTags sets the tags of the test suite
func (b *DefaultTestSuiteBuilder) WithTags(tags []string) TestSuiteBuilder {
	b.testSuite.Tags = tags
	return b
}

// WithMetadata sets the metadata of the test suite
func (b *DefaultTestSuiteBuilder) WithMetadata(metadata map[string]interface{}) TestSuiteBuilder {
	b.testSuite.Metadata = metadata
	return b
}

// Build builds the test suite
func (b *DefaultTestSuiteBuilder) Build() (*TestSuite, error) {
	// Validate test suite
	if b.testSuite.ID == "" {
		return nil, errors.New("test suite ID cannot be empty")
	}
	if b.testSuite.Name == "" {
		return nil, errors.New("test suite name cannot be empty")
	}
	if len(b.testSuite.TestCases) == 0 {
		return nil, errors.New("test suite must have at least one test case")
	}

	return b.testSuite, nil
}
