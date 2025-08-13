// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// DefaultTestRunner is the default implementation of the types.TestRunner interface
type DefaultTestRunner struct {
	// DetectionEngine is the engine for detecting vulnerabilities
	DetectionEngine detection.DetectionEngine
	// ReportGenerator is the generator for creating reports
	ReportGenerator types.ReportGenerator
	// Concurrency is the number of concurrent tests to run
	Concurrency int
	// MaxRetries is the maximum number of retries for failed tests
	MaxRetries int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// NewDefaultTestRunner creates a new default test runner
func NewDefaultTestRunner(detectionEngine detection.DetectionEngine, reportGenerator types.ReportGenerator) *DefaultTestRunner {
	return &DefaultTestRunner{
		DetectionEngine:  detectionEngine,
		ReportGenerator:  reportGenerator,
		Concurrency:      5,
		MaxRetries:       3,
		RetryDelay:       time.Second * 2,
	}
}

// RunTest runs a single test case
func (r *DefaultTestRunner) RunTest(ctx context.Context, testCase *types.TestCase, provider core.Provider, model string) (*types.TestResult, error) {
	// Validate the test case
	if err := r.ValidateTestCase(testCase); err != nil {
		return nil, fmt.Errorf("invalid test case: %w", err)
	}

	// Start timer
	startTime := time.Now()

	// Create a chat completion request
	request := &core.ChatCompletionRequest{
		Model: model,
		Messages: []core.Message{
			{
				Role:    "user",
				Content: testCase.Prompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	// Execute the request with retries
	var response *core.ChatCompletionResponse
	var err error
	for retry := 0; retry <= r.MaxRetries; retry++ {
		response, err = provider.ChatCompletion(ctx, request)
		if err == nil {
			break
		}
		
		if retry < r.MaxRetries {
			time.Sleep(r.RetryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute chat completion: %w", err)
	}

	// Extract the response content
	responseContent := ""
	if len(response.Choices) > 0 {
		responseContent = response.Choices[0].Message.Content
	}

	// Detect vulnerabilities using the detection engine
	detectionResults, err := r.DetectionEngine.Detect(ctx, responseContent, testCase.DetectionCriteria)
	if err != nil {
		return nil, fmt.Errorf("failed to detect vulnerabilities: %w", err)
	}

	// Determine if the test passed
	passed := true
	for _, result := range detectionResults {
		if result.Detected {
			passed = false
			break
		}
	}

	// Create test result
	result := &types.TestResult{
		TestCase:      testCase,
		Passed:        passed,
		Response:      responseContent,
		ExecutionTime: time.Since(startTime),
		Timestamp:     time.Now(),
	}

	return result, nil
}

// RunTestSuite runs a test suite
func (r *DefaultTestRunner) RunTestSuite(ctx context.Context, testSuite *types.TestSuite, provider core.Provider, model string) error {
	if testSuite == nil {
		return errors.New("test suite cannot be nil")
	}

	if len(testSuite.TestCases) == 0 {
		return errors.New("test suite contains no test cases")
	}

	// Create a wait group to wait for all tests to complete
	wg := sync.WaitGroup{}

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, r.Concurrency)

	// Create a context with cancellation
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a mutex to protect access to the results slice
	mutex := &sync.Mutex{}

	// Initialize the results slice
	results := make([]*types.TestResult, 0, len(testSuite.TestCases))

	// Track errors
	errs := make([]error, 0)
	errMutex := &sync.Mutex{}

	// Run each test case
	for _, testCase := range testSuite.TestCases {
		wg.Add(1)

		// Acquire semaphore
		sem <- struct{}{}

		// Run the test case in a goroutine
		go func(tc *types.TestCase) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Run the test
			result, err := r.RunTest(ctxWithCancel, tc, provider, model)
			if err != nil {
				// Record the error
				errMutex.Lock()
				errs = append(errs, err)
				errMutex.Unlock()

				// Create a failed result
				result = &types.TestResult{
					TestCase:      tc,
					Passed:        false,
					Response:      "",
					ExecutionTime: 0,
					Timestamp:     time.Now(),
					Notes:         fmt.Sprintf("Error running test: %v", err),
				}
			}

			// Add the result to the results slice
			mutex.Lock()
			results = append(results, result)
			mutex.Unlock()
		}(testCase)
	}

	// Wait for all tests to complete
	wg.Wait()

	// Store results in the test suite for future reference
	testSuite.Results = results

	// Return error if any tests failed
	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during test suite execution", len(errs))
	}

	return nil
}

// ValidateTestCase validates a test case
func (r *DefaultTestRunner) ValidateTestCase(testCase *types.TestCase) error {
	if testCase == nil {
		return errors.New("test case cannot be nil")
	}

	if testCase.ID == "" {
		return errors.New("test case ID cannot be empty")
	}

	if testCase.Name == "" {
		return errors.New("test case name cannot be empty")
	}

	if testCase.Prompt == "" {
		return errors.New("test case prompt cannot be empty")
	}

	if len(testCase.DetectionCriteria) == 0 {
		return errors.New("test case must have at least one detection criteria")
	}

	return nil
}
