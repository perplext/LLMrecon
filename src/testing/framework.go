package testing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/copilot"
	"github.com/perplext/LLMrecon/src/security/access/common"
)

// EnhancedTestingFramework provides AI-powered security testing capabilities
// Integrates with the Security Copilot for intelligent test planning and execution
type EnhancedTestingFramework struct {
	config           *FrameworkConfig
	copilot          copilot.SecurityCopilot
	orchestrator     *TestOrchestrator
	executor         *TestExecutor
	analyzer         *ResultAnalyzer
	reporter         *TestReporter
	learningEngine   *LearningEngine
	complianceEngine *ComplianceEngine
	logger           common.AuditLogger
	metrics          *FrameworkMetrics
	mu               sync.RWMutex
}

// FrameworkConfig configures the testing framework
type FrameworkConfig struct {
	// Execution settings
	MaxConcurrentTests    int
	DefaultTimeout        time.Duration
	RetryAttempts        int
	ContinuousLearning   bool
	
	// AI Integration
	CopilotEnabled       bool
	AutoPlanGeneration   bool
	IntelligentRetries   bool
	AdaptiveTesting      bool
	
	// Quality settings
	MinConfidenceLevel   float64
	RequiredEvidence     int
	ComplianceFrameworks []string
	
	// Performance settings
	ResourceLimits       *ResourceLimits
	DistributedExecution bool
	LoadBalancing        bool
	
	// Reporting settings
	DetailedReporting    bool
	RealTimeUpdates      bool
	ExportFormats        []string
}

// TestSuite represents a comprehensive security test suite
type TestSuite struct {
	ID                string
	Name              string
	Description       string
	Objective         *SecurityObjective
	TestCases         []TestCase
	Dependencies      []Dependency
	Configuration     *SuiteConfiguration
	Schedule          *TestSchedule
	Metadata          map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TestCase represents an individual security test
type TestCase struct {
	ID               string
	Name             string
	Description      string
	Category         TestCategory
	Priority         TestPriority
	Technique        string
	Configuration    *TestConfiguration
	ExpectedResults  []ExpectedResult
	SuccessCriteria  []SuccessCriterion
	Prerequisites    []Prerequisite
	EstimatedDuration time.Duration
	ResourceRequirements *ResourceRequirements
	RiskLevel        string
	Metadata         map[string]interface{}
}

// TestConfiguration holds test-specific configuration
type TestConfiguration struct {
	Target           *TargetConfiguration
	Parameters       map[string]interface{}
	Constraints      *ExecutionConstraints
	RetryPolicy      *RetryPolicy
	TimeoutSettings  *TimeoutSettings
	ValidationRules  []ValidationRule
}

// TestExecution tracks the execution of a test
type TestExecution struct {
	ID              string
	TestCaseID      string
	SuiteID         string
	Status          ExecutionStatus
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Result          *TestResult
	Attempts        int
	ErrorMessages   []string
	Metadata        map[string]interface{}
	ResourceUsage   *ResourceUsage
	ExecutorNode    string
}

// TestResult contains the results of a test execution
type TestResult struct {
	Success         bool
	Confidence      float64
	Evidence        []Evidence
	Findings        []SecurityFinding
	Metrics         *TestMetrics
	LearningData    *LearningData
	Recommendations []Recommendation
	NextSteps       []string
	RawOutput       string
}

// SecurityFinding represents a security vulnerability or issue found
type SecurityFinding struct {
	ID              string
	Type            FindingType
	Severity        SeverityLevel
	Title           string
	Description     string
	Evidence        []Evidence
	Impact          string
	Remediation     []RemediationStep
	CVSS            *CVSSScore
	References      []string
	ConfidenceLevel float64
	FirstSeen       time.Time
	LastSeen        time.Time
}

// NewEnhancedTestingFramework creates a new testing framework
func NewEnhancedTestingFramework(config *FrameworkConfig, copilot copilot.SecurityCopilot, logger common.AuditLogger) *EnhancedTestingFramework {
	framework := &EnhancedTestingFramework{
		config:           config,
		copilot:          copilot,
		logger:           logger,
		metrics:          NewFrameworkMetrics(),
		orchestrator:     NewTestOrchestrator(config),
		executor:         NewTestExecutor(config, logger),
		analyzer:         NewResultAnalyzer(config),
		reporter:         NewTestReporter(config),
		learningEngine:   NewLearningEngine(config, copilot),
		complianceEngine: NewComplianceEngine(config),
	}

	// Initialize framework
	framework.initialize()

	return framework
}

// GenerateTestSuite creates a test suite using AI-powered planning
func (f *EnhancedTestingFramework) GenerateTestSuite(ctx context.Context, objective *SecurityObjective, target *TargetProfile) (*TestSuite, error) {
	f.logger.LogSecurityEvent("test_suite_generation_started", map[string]interface{}{
		"objective_id": objective.ID,
		"target_id":    target.ID,
	})

	// Use copilot to generate testing strategy if enabled
	var strategy *copilot.TestingStrategy
	var err error
	
	if f.config.CopilotEnabled && f.config.AutoPlanGeneration {
		strategy, err = f.copilot.GenerateStrategy(ctx, objective)
		if err != nil {
			f.logger.LogSecurityEvent("copilot_strategy_generation_failed", map[string]interface{}{
				"error": err.Error(),
			})
			// Fall back to traditional planning
			strategy = f.generateFallbackStrategy(objective, target)
		}
	} else {
		strategy = f.generateFallbackStrategy(objective, target)
	}

	// Convert strategy to test suite
	suite := f.convertStrategyToSuite(strategy, objective, target)

	// Enhance with AI recommendations if enabled
	if f.config.CopilotEnabled {
		if err := f.enhanceWithCopilotRecommendations(ctx, suite, target); err != nil {
			f.logger.LogSecurityEvent("copilot_enhancement_failed", map[string]interface{}{
				"suite_id": suite.ID,
				"error":    err.Error(),
			})
		}
	}

	// Validate test suite
	if err := f.validateTestSuite(suite); err != nil {
		return nil, fmt.Errorf("test suite validation failed: %w", err)
	}

	f.logger.LogSecurityEvent("test_suite_generated", map[string]interface{}{
		"suite_id":        suite.ID,
		"test_count":      len(suite.TestCases),
		"estimated_duration": f.calculateSuiteDuration(suite),
	})

	return suite, nil
}

// ExecuteTestSuite runs a complete test suite with intelligent orchestration
func (f *EnhancedTestingFramework) ExecuteTestSuite(ctx context.Context, suite *TestSuite) (*SuiteExecution, error) {
	execution := &SuiteExecution{
		ID:         generateExecutionID(),
		SuiteID:    suite.ID,
		Status:     StatusRunning,
		StartTime:  time.Now(),
		TestResults: make(map[string]*TestResult),
		Metrics:    NewSuiteMetrics(),
	}

	f.logger.LogSecurityEvent("test_suite_execution_started", map[string]interface{}{
		"execution_id": execution.ID,
		"suite_id":     suite.ID,
		"test_count":   len(suite.TestCases),
	})

	// Execute tests using orchestrator
	results, err := f.orchestrator.ExecuteTests(ctx, suite.TestCases, execution)
	if err != nil {
		execution.Status = StatusFailed
		execution.EndTime = time.Now()
		execution.Duration = time.Since(execution.StartTime)
		return execution, fmt.Errorf("test execution failed: %w", err)
	}

	// Process results
	execution.TestResults = results
	execution.Status = f.determineOverallStatus(results)
	execution.EndTime = time.Now()
	execution.Duration = time.Since(execution.StartTime)

	// Analyze results with AI if enabled
	if f.config.CopilotEnabled {
		f.analyzeResultsWithCopilot(ctx, execution)
	}

	// Update learning engine
	if f.config.ContinuousLearning {
		f.learningEngine.ProcessResults(ctx, execution)
	}

	f.logger.LogSecurityEvent("test_suite_execution_completed", map[string]interface{}{
		"execution_id":     execution.ID,
		"status":           execution.Status,
		"duration":         execution.Duration,
		"successful_tests": f.countSuccessfulTests(results),
		"total_tests":      len(results),
	})

	return execution, nil
}

// ExecuteAdaptiveTest runs a single test with AI-powered adaptation
func (f *EnhancedTestingFramework) ExecuteAdaptiveTest(ctx context.Context, testCase TestCase, target *TargetProfile) (*TestResult, error) {
	// Start with initial execution
	result, err := f.executor.ExecuteTest(ctx, testCase)
	if err != nil {
		return nil, err
	}

	// If adaptive testing is enabled and the test failed, try to improve it
	if f.config.AdaptiveTesting && f.config.CopilotEnabled && !result.Success {
		f.logger.LogSecurityEvent("adaptive_test_improvement_started", map[string]interface{}{
			"test_id": testCase.ID,
			"initial_confidence": result.Confidence,
		})

		// Get recommendations from copilot for improvement
		improved, err := f.improveTestWithCopilot(ctx, testCase, result, target)
		if err != nil {
			f.logger.LogSecurityEvent("adaptive_improvement_failed", map[string]interface{}{
				"test_id": testCase.ID,
				"error":   err.Error(),
			})
			return result, nil // Return original result if improvement fails
		}

		// Execute improved test
		improvedResult, err := f.executor.ExecuteTest(ctx, *improved)
		if err != nil {
			return result, nil // Return original result if improved execution fails
		}

		// Use improved result if it's better
		if improvedResult.Confidence > result.Confidence {
			f.logger.LogSecurityEvent("adaptive_improvement_successful", map[string]interface{}{
				"test_id": testCase.ID,
				"original_confidence": result.Confidence,
				"improved_confidence": improvedResult.Confidence,
			})
			result = improvedResult
		}
	}

	return result, nil
}

// AnalyzeResults performs comprehensive analysis of test results
func (f *EnhancedTestingFramework) AnalyzeResults(ctx context.Context, execution *SuiteExecution) (*TestAnalysis, error) {
	analysis := &TestAnalysis{
		ExecutionID:      execution.ID,
		Timestamp:        time.Now(),
		OverallScore:     f.calculateOverallScore(execution),
		SecurityPosture:  f.assessSecurityPosture(execution),
		Findings:         f.extractFindings(execution),
		Recommendations:  make([]Recommendation, 0),
		TrendAnalysis:    f.performTrendAnalysis(execution),
		ComplianceStatus: f.assessCompliance(execution),
	}

	// Enhanced analysis with AI if enabled
	if f.config.CopilotEnabled {
		aiAnalysis, err := f.getAIAnalysis(ctx, execution)
		if err != nil {
			f.logger.LogSecurityEvent("ai_analysis_failed", map[string]interface{}{
				"execution_id": execution.ID,
				"error":        err.Error(),
			})
		} else {
			analysis.AIInsights = aiAnalysis
			analysis.Recommendations = append(analysis.Recommendations, aiAnalysis.Recommendations...)
		}
	}

	return analysis, nil
}

// GenerateReport creates comprehensive test reports
func (f *EnhancedTestingFramework) GenerateReport(ctx context.Context, execution *SuiteExecution, format ReportFormat) (*TestReport, error) {
	analysis, err := f.AnalyzeResults(ctx, execution)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze results: %w", err)
	}

	report := f.reporter.GenerateReport(execution, analysis, format)
	
	f.logger.LogSecurityEvent("test_report_generated", map[string]interface{}{
		"execution_id":  execution.ID,
		"format":        format,
		"report_size":   len(report.Content),
		"findings_count": len(analysis.Findings),
	})

	return report, nil
}

// Helper methods

func (f *EnhancedTestingFramework) initialize() {
	// Initialize components
	f.orchestrator.Initialize()
	f.executor.Initialize()
	f.analyzer.Initialize()
	f.reporter.Initialize()
	f.learningEngine.Initialize()
	f.complianceEngine.Initialize()

	f.logger.LogSecurityEvent("framework_initialized", map[string]interface{}{
		"copilot_enabled":    f.config.CopilotEnabled,
		"adaptive_testing":   f.config.AdaptiveTesting,
		"continuous_learning": f.config.ContinuousLearning,
	})
}

func (f *EnhancedTestingFramework) generateFallbackStrategy(objective *SecurityObjective, target *TargetProfile) *copilot.TestingStrategy {
	// Generate a basic strategy when copilot is not available
	return &copilot.TestingStrategy{
		ID:          generateStrategyID(),
		Name:        fmt.Sprintf("Fallback Strategy for %s", objective.Name),
		Description: "Basic testing strategy generated without AI assistance",
		ObjectiveID: objective.ID,
		Phases: []copilot.TestingPhase{
			{
				ID:          "basic_testing",
				Name:        "Basic Security Testing",
				Description: "Fundamental security tests",
				Duration:    4 * time.Hour,
				Attacks:     []string{"basic_injection", "simple_bypass"},
				Objectives:  []string{"Test basic security controls"},
			},
		},
	}
}

func (f *EnhancedTestingFramework) convertStrategyToSuite(strategy *copilot.TestingStrategy, objective *SecurityObjective, target *TargetProfile) *TestSuite {
	suite := &TestSuite{
		ID:          generateSuiteID(),
		Name:        strategy.Name,
		Description: strategy.Description,
		Objective:   objective,
		TestCases:   make([]TestCase, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Convert strategy phases to test cases
	for _, phase := range strategy.Phases {
		for _, attack := range phase.Attacks {
			testCase := TestCase{
				ID:          generateTestCaseID(),
				Name:        fmt.Sprintf("%s - %s", phase.Name, attack),
				Description: fmt.Sprintf("Execute %s attack as part of %s", attack, phase.Description),
				Category:    f.categorizeAttack(attack),
				Priority:    PriorityMedium,
				Technique:   attack,
				Configuration: &TestConfiguration{
					Target: &TargetConfiguration{
						Profile: target,
					},
				},
				EstimatedDuration: phase.Duration / time.Duration(len(phase.Attacks)),
			}
			suite.TestCases = append(suite.TestCases, testCase)
		}
	}

	return suite
}

func (f *EnhancedTestingFramework) enhanceWithCopilotRecommendations(ctx context.Context, suite *TestSuite, target *TargetProfile) error {
	// Get attack recommendations from copilot
	recommendations, err := f.copilot.RecommendAttacks(ctx, target)
	if err != nil {
		return err
	}

	// Add high-confidence primary recommendations as test cases
	for _, rec := range recommendations.Primary {
		if rec.Confidence > f.config.MinConfidenceLevel {
			testCase := TestCase{
				ID:          generateTestCaseID(),
				Name:        fmt.Sprintf("AI-Recommended: %s", rec.AttackName),
				Description: rec.Rationale,
				Category:    f.categorizeAttack(rec.AttackType),
				Priority:    f.determinePriority(rec.Priority),
				Technique:   rec.AttackID,
				Configuration: &TestConfiguration{
					Target: &TargetConfiguration{
						Profile: target,
					},
					Parameters: rec.Configuration,
				},
				EstimatedDuration: f.estimateDuration(rec),
			}
			suite.TestCases = append(suite.TestCases, testCase)
		}
	}

	return nil
}

func (f *EnhancedTestingFramework) validateTestSuite(suite *TestSuite) error {
	if len(suite.TestCases) == 0 {
		return fmt.Errorf("test suite must contain at least one test case")
	}

	// Validate each test case
	for _, testCase := range suite.TestCases {
		if err := f.validateTestCase(testCase); err != nil {
			return fmt.Errorf("invalid test case %s: %w", testCase.ID, err)
		}
	}

	return nil
}

func (f *EnhancedTestingFramework) validateTestCase(testCase TestCase) error {
	if testCase.ID == "" {
		return fmt.Errorf("test case ID cannot be empty")
	}
	if testCase.Technique == "" {
		return fmt.Errorf("test case must specify a technique")
	}
	return nil
}

func (f *EnhancedTestingFramework) calculateSuiteDuration(suite *TestSuite) time.Duration {
	total := time.Duration(0)
	for _, testCase := range suite.TestCases {
		total += testCase.EstimatedDuration
	}
	return total
}

func (f *EnhancedTestingFramework) determineOverallStatus(results map[string]*TestResult) ExecutionStatus {
	if len(results) == 0 {
		return StatusFailed
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	successRate := float64(successCount) / float64(len(results))
	if successRate >= 0.8 {
		return StatusSuccess
	} else if successRate >= 0.5 {
		return StatusPartialSuccess
	} else {
		return StatusFailed
	}
}

func (f *EnhancedTestingFramework) countSuccessfulTests(results map[string]*TestResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

func (f *EnhancedTestingFramework) analyzeResultsWithCopilot(ctx context.Context, execution *SuiteExecution) {
	// Convert test results to attack executions for copilot analysis
	attackExecutions := f.convertToAttackExecutions(execution)
	
	analysis, err := f.copilot.AnalyzeResults(ctx, attackExecutions)
	if err != nil {
		f.logger.LogSecurityEvent("copilot_analysis_failed", map[string]interface{}{
			"execution_id": execution.ID,
			"error":        err.Error(),
		})
		return
	}

	// Store AI analysis in execution metadata
	execution.Metadata = make(map[string]interface{})
	execution.Metadata["ai_analysis"] = analysis
}

func (f *EnhancedTestingFramework) improveTestWithCopilot(ctx context.Context, testCase TestCase, result *TestResult, target *TargetProfile) (*TestCase, error) {
	// This would use the copilot to suggest improvements to the test case
	// For now, return a modified version of the test case
	improved := testCase
	improved.ID = generateTestCaseID()
	improved.Name = "Improved: " + testCase.Name
	
	// In a real implementation, this would analyze the failure and suggest specific improvements
	// based on the copilot's recommendations
	
	return &improved, nil
}

func (f *EnhancedTestingFramework) calculateOverallScore(execution *SuiteExecution) float64 {
	if len(execution.TestResults) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, result := range execution.TestResults {
		totalConfidence += result.Confidence
	}

	return totalConfidence / float64(len(execution.TestResults))
}

func (f *EnhancedTestingFramework) assessSecurityPosture(execution *SuiteExecution) SecurityPosture {
	successRate := float64(f.countSuccessfulTests(execution.TestResults)) / float64(len(execution.TestResults))
	
	if successRate >= 0.8 {
		return PostureVulnerable
	} else if successRate >= 0.5 {
		return PostureWeakened
	} else if successRate >= 0.2 {
		return PostureResilient
	} else {
		return PostureSecure
	}
}

func (f *EnhancedTestingFramework) extractFindings(execution *SuiteExecution) []SecurityFinding {
	findings := make([]SecurityFinding, 0)
	
	for testID, result := range execution.TestResults {
		if result.Success {
			finding := SecurityFinding{
				ID:              generateFindingID(),
				Type:            FindingVulnerability,
				Severity:        f.determineSeverity(result.Confidence),
				Title:           fmt.Sprintf("Security vulnerability detected in test %s", testID),
				Description:     fmt.Sprintf("Test execution succeeded with %.1f%% confidence", result.Confidence*100),
				Evidence:        result.Evidence,
				ConfidenceLevel: result.Confidence,
				FirstSeen:       time.Now(),
				LastSeen:        time.Now(),
			}
			findings = append(findings, finding)
		}
	}
	
	return findings
}

func (f *EnhancedTestingFramework) performTrendAnalysis(execution *SuiteExecution) *TrendAnalysis {
	// This would analyze trends across multiple executions
	// For now, return basic analysis
	return &TrendAnalysis{
		Period:         "current",
		SuccessRate:    float64(f.countSuccessfulTests(execution.TestResults)) / float64(len(execution.TestResults)),
		Trend:          "stable",
		Confidence:     0.7,
	}
}

func (f *EnhancedTestingFramework) assessCompliance(execution *SuiteExecution) map[string]ComplianceStatus {
	status := make(map[string]ComplianceStatus)
	
	for _, framework := range f.config.ComplianceFrameworks {
		// Assess compliance for each framework
		status[framework] = f.complianceEngine.AssessCompliance(framework, execution)
	}
	
	return status
}

func (f *EnhancedTestingFramework) getAIAnalysis(ctx context.Context, execution *SuiteExecution) (*AIAnalysis, error) {
	// Convert execution to format suitable for AI analysis
	attackExecutions := f.convertToAttackExecutions(execution)
	
	analysis, err := f.copilot.AnalyzeResults(ctx, attackExecutions)
	if err != nil {
		return nil, err
	}
	
	return &AIAnalysis{
		Insights:        analysis.Insights,
		Patterns:        analysis.Patterns,
		Recommendations: f.convertToRecommendations(analysis.Improvements),
		Confidence:      f.calculateAnalysisConfidence(analysis),
	}, nil
}

// Helper functions for conversions and utilities

func generateExecutionID() string { return fmt.Sprintf("exec_%d", time.Now().UnixNano()) }
func generateSuiteID() string { return fmt.Sprintf("suite_%d", time.Now().UnixNano()) }
func generateTestCaseID() string { return fmt.Sprintf("test_%d", time.Now().UnixNano()) }
func generateStrategyID() string { return fmt.Sprintf("strategy_%d", time.Now().UnixNano()) }
func generateFindingID() string { return fmt.Sprintf("finding_%d", time.Now().UnixNano()) }

func (f *EnhancedTestingFramework) categorizeAttack(attack string) TestCategory {
	if strings.Contains(attack, "injection") {
		return CategoryInjection
	} else if strings.Contains(attack, "modal") {
		return CategoryMultimodal
	} else if strings.Contains(attack, "flow") {
		return CategorySocialEngineering
	}
	return CategoryGeneral
}

func (f *EnhancedTestingFramework) determinePriority(priority int) TestPriority {
	if priority >= 8 {
		return PriorityHigh
	} else if priority >= 5 {
		return PriorityMedium
	}
	return PriorityLow
}

func (f *EnhancedTestingFramework) estimateDuration(rec copilot.AttackRecommendation) time.Duration {
	// Estimate based on attack complexity and type
	baseDuration := 5 * time.Minute
	if rec.RiskLevel == "high" {
		baseDuration = 10 * time.Minute
	}
	return baseDuration
}

func (f *EnhancedTestingFramework) determineSeverity(confidence float64) SeverityLevel {
	if confidence >= 0.8 {
		return SeverityHigh
	} else if confidence >= 0.6 {
		return SeverityMedium
	} else if confidence >= 0.4 {
		return SeverityLow
	}
	return SeverityInfo
}

func (f *EnhancedTestingFramework) convertToAttackExecutions(execution *SuiteExecution) []*copilot.AttackExecution {
	executions := make([]*copilot.AttackExecution, 0)
	
	for testID, result := range execution.TestResults {
		exec := &copilot.AttackExecution{
			ExecutionID: testID,
			AttackID:    testID,
			AttackType:  "test_execution",
			StartTime:   execution.StartTime,
			EndTime:     execution.EndTime,
			Duration:    execution.Duration,
			Success:     result.Success,
			Confidence:  result.Confidence,
			Response:    result.RawOutput,
		}
		executions = append(executions, exec)
	}
	
	return executions
}

func (f *EnhancedTestingFramework) convertToRecommendations(improvements []copilot.Improvement) []Recommendation {
	recommendations := make([]Recommendation, len(improvements))
	
	for i, improvement := range improvements {
		recommendations[i] = Recommendation{
			Type:        improvement.Area,
			Description: improvement.Description,
			Priority:    improvement.Priority,
			Difficulty:  improvement.Difficulty,
			Impact:      improvement.Impact,
		}
	}
	
	return recommendations
}

func (f *EnhancedTestingFramework) calculateAnalysisConfidence(analysis *copilot.Analysis) float64 {
	if len(analysis.Insights) == 0 {
		return 0.0
	}
	
	totalConfidence := 0.0
	for _, insight := range analysis.Insights {
		totalConfidence += insight.Confidence
	}
	
	return totalConfidence / float64(len(analysis.Insights))
}

// Supporting types for the framework

type FrameworkMetrics struct {
	TotalExecutions     int
	SuccessfulExecutions int
	AverageExecutionTime time.Duration
	TestsExecuted       int
	FindingsDiscovered  int
}

func NewFrameworkMetrics() *FrameworkMetrics {
	return &FrameworkMetrics{}
}

type SuiteExecution struct {
	ID          string
	SuiteID     string
	Status      ExecutionStatus
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	TestResults map[string]*TestResult
	Metrics     *SuiteMetrics
	Metadata    map[string]interface{}
}

type SuiteMetrics struct {
	TotalTests      int
	SuccessfulTests int
	FailedTests     int
	SuccessRate     float64
	AverageConfidence float64
}

func NewSuiteMetrics() *SuiteMetrics {
	return &SuiteMetrics{}
}

// Enums and constants

type TestCategory string
const (
	CategoryInjection       TestCategory = "injection"
	CategoryMultimodal      TestCategory = "multimodal"
	CategorySocialEngineering TestCategory = "social_engineering"
	CategoryGeneral         TestCategory = "general"
)

type TestPriority string
const (
	PriorityHigh   TestPriority = "high"
	PriorityMedium TestPriority = "medium"
	PriorityLow    TestPriority = "low"
)

type ExecutionStatus string
const (
	StatusRunning        ExecutionStatus = "running"
	StatusSuccess        ExecutionStatus = "success"
	StatusPartialSuccess ExecutionStatus = "partial_success"
	StatusFailed         ExecutionStatus = "failed"
	StatusSuspended      ExecutionStatus = "suspended"
)

type FindingType string
const (
	FindingVulnerability FindingType = "vulnerability"
	FindingWeakness      FindingType = "weakness"
	FindingMisconfiguration FindingType = "misconfiguration"
)

type SeverityLevel string
const (
	SeverityHigh   SeverityLevel = "high"
	SeverityMedium SeverityLevel = "medium" 
	SeverityLow    SeverityLevel = "low"
	SeverityInfo   SeverityLevel = "info"
)

type SecurityPosture string
const (
	PostureSecure     SecurityPosture = "secure"
	PostureResilient  SecurityPosture = "resilient"
	PostureWeakened   SecurityPosture = "weakened"
	PostureVulnerable SecurityPosture = "vulnerable"
)

type ReportFormat string
const (
	FormatJSON ReportFormat = "json"
	FormatHTML ReportFormat = "html"
	FormatPDF  ReportFormat = "pdf"
	FormatCSV  ReportFormat = "csv"
)

type ComplianceStatus string
const (
	CompliancePassing ComplianceStatus = "passing"
	ComplianceFailing ComplianceStatus = "failing"
	CompliancePartial ComplianceStatus = "partial"
	ComplianceUnknown ComplianceStatus = "unknown"
)