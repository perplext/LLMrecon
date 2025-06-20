package copilot

import (
	"context"
	"time"
)

// SecurityCopilot is the main interface for the AI security assistant
type SecurityCopilot interface {
	// ProcessQuery handles natural language security queries
	ProcessQuery(ctx context.Context, query string, options *QueryOptions) (*QueryResponse, error)
	
	// RecommendAttacks suggests appropriate attacks for a target
	RecommendAttacks(ctx context.Context, target *TargetProfile) (*AttackRecommendations, error)
	
	// AnalyzeResults learns from attack execution results
	AnalyzeResults(ctx context.Context, results []*AttackExecution) (*Analysis, error)
	
	// GenerateStrategy creates comprehensive testing strategies
	GenerateStrategy(ctx context.Context, objective *SecurityObjective) (*TestingStrategy, error)
	
	// ExplainReasoning provides explanations for recommendations
	ExplainReasoning(ctx context.Context, recommendation *AttackRecommendation) (*Explanation, error)
}

// QueryOptions configures how queries are processed
type QueryOptions struct {
	// Context for the query
	Context     map[string]interface{}
	
	// Previous conversation history
	History     []ConversationTurn
	
	// User preferences
	Preferences *UserPreferences
	
	// Execution constraints
	Constraints *ExecutionConstraints
}

// ConversationTurn represents one exchange in the conversation
type ConversationTurn struct {
	UserMessage      string
	CopilotResponse  string
	Timestamp        time.Time
	Actions          []string
	Results          []string
}

// UserPreferences configure copilot behavior
type UserPreferences struct {
	// Preferred attack categories
	PreferredCategories []string
	
	// Risk tolerance (conservative, moderate, aggressive)
	RiskTolerance      string
	
	// Explanation detail level (brief, detailed, comprehensive)
	ExplanationLevel   string
	
	// Automation level (manual, assisted, automated)
	AutomationLevel    string
	
	// Industry context
	Industry           string
	
	// Compliance requirements
	ComplianceFrameworks []string
}

// ExecutionConstraints limit what the copilot can do
type ExecutionConstraints struct {
	// Maximum number of attacks to suggest
	MaxAttacks         int
	
	// Time budget for execution
	TimeLimit          time.Duration
	
	// Resource limits
	MaxConcurrency     int
	MaxTokensPerAttack int
	
	// Safety constraints
	SafetyLevel        string
	ProhibitedTechniques []string
	
	// Target restrictions
	AllowedTargets     []string
	ForbiddenTargets   []string
}

// QueryResponse contains the copilot's response to a query
type QueryResponse struct {
	// Unique response ID
	ID              string
	
	// Natural language response
	Response        string
	
	// Structured actions to take
	Actions         []Action
	
	// Attack recommendations
	Recommendations *AttackRecommendations
	
	// Follow-up questions
	FollowUpQuestions []string
	
	// Confidence in the response (0.0-1.0)
	Confidence      float64
	
	// Reasoning explanation
	Reasoning       *Reasoning
	
	// Additional metadata
	Metadata        map[string]interface{}
}

// Action represents something the copilot wants to do
type Action struct {
	// Action type (execute_attack, analyze_target, generate_report, etc.)
	Type        ActionType
	
	// Action description
	Description string
	
	// Action parameters
	Parameters  map[string]interface{}
	
	// Expected outcome
	ExpectedOutcome string
	
	// Risk level (low, medium, high)
	RiskLevel   string
	
	// Requires user confirmation
	RequiresConfirmation bool
}

// ActionType defines types of actions the copilot can suggest
type ActionType string

const (
	ActionExecuteAttack    ActionType = "execute_attack"
	ActionAnalyzeTarget    ActionType = "analyze_target"
	ActionGenerateReport   ActionType = "generate_report"
	ActionCreateStrategy   ActionType = "create_strategy"
	ActionValidateCompliance ActionType = "validate_compliance"
	ActionSearchKnowledge  ActionType = "search_knowledge"
	ActionLearnFromResults ActionType = "learn_from_results"
	ActionOptimizePayload  ActionType = "optimize_payload"
)

// TargetProfile describes a target system for security testing
type TargetProfile struct {
	// Basic identification
	ID           string
	Name         string
	Description  string
	
	// Technical details
	ModelType    string
	Provider     string
	Version      string
	Capabilities []string
	
	// Deployment context
	Environment  string // development, staging, production
	Industry     string
	UseCase      string
	
	// Security posture
	KnownDefenses    []string
	PreviousTests    []PreviousTest
	VulnerabilityHistory []VulnerabilityRecord
	
	// Compliance requirements
	ComplianceFrameworks []string
	RegulatoryConstraints []string
	
	// Risk factors
	SensitivityLevel string
	DataTypes        []string
	UserBase         string
	
	// Technical constraints
	RateLimits       *RateLimitInfo
	AccessMethods    []string
	AuthenticationRequired bool
}

// PreviousTest records previous security testing
type PreviousTest struct {
	Date         time.Time
	TestType     string
	AttacksUsed  []string
	Results      string
	Findings     []string
	Remediation  []string
}

// VulnerabilityRecord tracks known vulnerabilities
type VulnerabilityRecord struct {
	ID           string
	Type         string
	Severity     string
	Description  string
	Status       string // open, mitigated, fixed
	DiscoveryDate time.Time
	CVENumber    string
}

// RateLimitInfo describes API rate limits
type RateLimitInfo struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	TokensPerMinute   int
	ConcurrentRequests int
}

// AttackRecommendations contains suggested attacks
type AttackRecommendations struct {
	// Primary recommendations
	Primary     []AttackRecommendation
	
	// Alternative approaches
	Alternatives []AttackRecommendation
	
	// Experimental techniques
	Experimental []AttackRecommendation
	
	// Overall strategy
	Strategy    *RecommendationStrategy
	
	// Success probability estimates
	SuccessProbability float64
	
	// Risk assessment
	RiskAssessment *RiskAssessment
}

// AttackRecommendation suggests a specific attack
type AttackRecommendation struct {
	// Attack identification
	AttackID     string
	AttackType   string
	AttackName   string
	
	// Recommendation details
	Rationale    string
	Confidence   float64
	Priority     int
	
	// Execution parameters
	Configuration map[string]interface{}
	
	// Expected outcomes
	ExpectedResults []string
	SuccessProbability float64
	
	// Dependencies
	Prerequisites []string
	Dependencies  []string
	
	// Risk factors
	RiskLevel    string
	Mitigations  []string
	
	// Learning potential
	LearningValue float64
	NoveltyScore  float64
}

// RecommendationStrategy explains the overall approach
type RecommendationStrategy struct {
	// Strategy name
	Name         string
	
	// Strategy description
	Description  string
	
	// Strategic phases
	Phases       []StrategyPhase
	
	// Success criteria
	SuccessCriteria []string
	
	// Risk mitigation
	RiskMitigation []string
	
	// Expected timeline
	Timeline     time.Duration
}

// StrategyPhase represents a phase in the testing strategy
type StrategyPhase struct {
	Name         string
	Description  string
	Duration     time.Duration
	Attacks      []string
	Objectives   []string
	Dependencies []string
}

// RiskAssessment evaluates the risks of recommended attacks
type RiskAssessment struct {
	// Overall risk level
	OverallRisk  string
	
	// Specific risk factors
	RiskFactors  []RiskFactor
	
	// Mitigation strategies
	Mitigations  []string
	
	// Monitoring recommendations
	Monitoring   []string
	
	// Rollback procedures
	RollbackPlan []string
}

// RiskFactor identifies a specific risk
type RiskFactor struct {
	Type         string
	Description  string
	Severity     string
	Probability  float64
	Impact       string
	Mitigation   string
}

// AttackExecution records the execution of an attack
type AttackExecution struct {
	// Execution identification
	ExecutionID  string
	AttackID     string
	AttackType   string
	
	// Execution details
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	
	// Target information
	TargetID     string
	Configuration map[string]interface{}
	
	// Results
	Success      bool
	Confidence   float64
	Response     string
	Evidence     []string
	
	// Metrics
	TokensUsed   int
	RequestsMade int
	Cost         float64
	
	// Learning data
	FailureReasons []string
	SuccessFactors []string
	Insights       []string
	
	// Context
	Environment  string
	UserFeedback string
	Metadata     map[string]interface{}
}

// Analysis contains insights from attack results
type Analysis struct {
	// Analysis identification
	ID           string
	Timestamp    time.Time
	
	// Key insights
	Insights     []Insight
	
	// Pattern recognition
	Patterns     []Pattern
	
	// Recommendations for improvement
	Improvements []Improvement
	
	// Target vulnerability assessment
	Vulnerabilities []VulnerabilityAssessment
	
	// Defense effectiveness
	DefenseAnalysis *DefenseAnalysis
	
	// Future recommendations
	FutureTests  []FutureTestRecommendation
}

// Insight represents a learned insight
type Insight struct {
	Type         string
	Description  string
	Confidence   float64
	Evidence     []string
	Implications []string
	Actionable   bool
}

// Pattern represents a discovered pattern
type Pattern struct {
	Type         string
	Description  string
	Frequency    float64
	Conditions   []string
	Outcomes     []string
	Reliability  float64
}

// Improvement suggests ways to improve attacks
type Improvement struct {
	Area         string
	Description  string
	Priority     string
	Difficulty   string
	Impact       string
	Steps        []string
}

// VulnerabilityAssessment evaluates target vulnerabilities
type VulnerabilityAssessment struct {
	VulnerabilityType string
	Severity         string
	Exploitability   float64
	Impact           string
	Remediation      []string
	Timeline         time.Duration
}

// DefenseAnalysis evaluates defense effectiveness
type DefenseAnalysis struct {
	OverallEffectiveness float64
	DefenseTypes        []string
	Strengths           []string
	Weaknesses          []string
	Bypasses            []string
	Recommendations     []string
}

// FutureTestRecommendation suggests future testing
type FutureTestRecommendation struct {
	TestType     string
	Description  string
	Priority     string
	Timeline     time.Duration
	Prerequisites []string
	ExpectedValue float64
}

// SecurityObjective defines what we're trying to achieve
type SecurityObjective struct {
	// Objective identification
	ID           string
	Name         string
	Description  string
	
	// Objective type (compliance, vulnerability_discovery, penetration_test, etc.)
	Type         ObjectiveType
	
	// Target systems
	Targets      []string
	
	// Success criteria
	SuccessCriteria []string
	
	// Constraints
	Constraints  *ObjectiveConstraints
	
	// Timeline
	Deadline     time.Time
	
	// Priority level
	Priority     string
	
	// Stakeholders
	Stakeholders []string
}

// ObjectiveType defines types of security objectives
type ObjectiveType string

const (
	ObjectiveCompliance        ObjectiveType = "compliance"
	ObjectiveVulnerabilityDiscovery ObjectiveType = "vulnerability_discovery"
	ObjectivePenetrationTest   ObjectiveType = "penetration_test"
	ObjectiveRedTeamExercise   ObjectiveType = "red_team_exercise"
	ObjectiveSecurityAudit     ObjectiveType = "security_audit"
	ObjectiveRiskAssessment    ObjectiveType = "risk_assessment"
)

// ObjectiveConstraints limit how objectives can be achieved
type ObjectiveConstraints struct {
	// Time constraints
	MaxDuration  time.Duration
	
	// Resource constraints
	MaxCost      float64
	MaxTokens    int
	
	// Technical constraints
	AllowedAttacks []string
	ForbiddenAttacks []string
	
	// Risk constraints
	MaxRiskLevel string
	
	// Compliance constraints
	RequiredFrameworks []string
}

// TestingStrategy provides a comprehensive plan
type TestingStrategy struct {
	// Strategy identification
	ID           string
	Name         string
	Description  string
	
	// Objective alignment
	ObjectiveID  string
	
	// Strategy phases
	Phases       []TestingPhase
	
	// Resource requirements
	Resources    *ResourceRequirements
	
	// Timeline
	Timeline     *Timeline
	
	// Risk management
	RiskManagement *RiskManagement
	
	// Success metrics
	SuccessMetrics []SuccessMetric
	
	// Deliverables
	Deliverables []Deliverable
}

// TestingPhase represents a phase in the testing strategy
type TestingPhase struct {
	ID           string
	Name         string
	Description  string
	Duration     time.Duration
	
	// Attacks in this phase
	Attacks      []string
	
	// Phase objectives
	Objectives   []string
	
	// Dependencies
	Dependencies []string
	
	// Success criteria
	SuccessCriteria []string
	
	// Exit criteria
	ExitCriteria []string
}

// ResourceRequirements specifies needed resources
type ResourceRequirements struct {
	// Human resources
	Personnel    []PersonnelRequirement
	
	// Technical resources
	Infrastructure []InfrastructureRequirement
	
	// Financial resources
	Budget       float64
	
	// Time resources
	Timeline     time.Duration
}

// PersonnelRequirement specifies needed people
type PersonnelRequirement struct {
	Role         string
	Skillset     []string
	Experience   string
	TimeCommitment time.Duration
}

// InfrastructureRequirement specifies needed infrastructure
type InfrastructureRequirement struct {
	Type         string
	Specifications map[string]interface{}
	Duration     time.Duration
	Cost         float64
}

// Timeline provides detailed timeline information
type Timeline struct {
	StartDate    time.Time
	EndDate      time.Time
	Milestones   []Milestone
	Dependencies []Dependency
}

// Milestone represents a key milestone
type Milestone struct {
	Name         string
	Date         time.Time
	Description  string
	Deliverables []string
	Criteria     []string
}

// Dependency represents a project dependency
type Dependency struct {
	Type         string
	Description  string
	Impact       string
	Mitigation   string
}

// RiskManagement provides risk management plan
type RiskManagement struct {
	IdentifiedRisks []IdentifiedRisk
	MitigationPlans []MitigationPlan
	ContingencyPlans []ContingencyPlan
	MonitoringPlan  *MonitoringPlan
}

// IdentifiedRisk represents an identified risk
type IdentifiedRisk struct {
	ID           string
	Description  string
	Probability  float64
	Impact       string
	RiskLevel    string
	Owner        string
}

// MitigationPlan provides a plan to mitigate risks
type MitigationPlan struct {
	RiskID       string
	Strategy     string
	Actions      []string
	Timeline     time.Duration
	Cost         float64
	Effectiveness float64
}

// ContingencyPlan provides backup plans
type ContingencyPlan struct {
	TriggerConditions []string
	Actions          []string
	Timeline         time.Duration
	Resources        []string
}

// MonitoringPlan specifies how to monitor risks
type MonitoringPlan struct {
	Metrics      []string
	Frequency    time.Duration
	Thresholds   map[string]float64
	Alerts       []string
}

// SuccessMetric defines how success is measured
type SuccessMetric struct {
	Name         string
	Description  string
	Target       float64
	Measurement  string
	Frequency    time.Duration
}

// Deliverable represents a project deliverable
type Deliverable struct {
	Name         string
	Description  string
	Type         string
	DueDate      time.Time
	Owner        string
	Dependencies []string
}

// Explanation provides reasoning for recommendations
type Explanation struct {
	// Summary explanation
	Summary      string
	
	// Detailed reasoning
	Reasoning    *Reasoning
	
	// Supporting evidence
	Evidence     []string
	
	// Alternative considerations
	Alternatives []Alternative
	
	// Confidence factors
	ConfidenceFactors []ConfidenceFactor
}

// Reasoning provides detailed reasoning
type Reasoning struct {
	// Logical steps
	Steps        []ReasoningStep
	
	// Assumptions made
	Assumptions  []string
	
	// Data sources used
	DataSources  []string
	
	// Methodology
	Methodology  string
}

// ReasoningStep represents a step in reasoning
type ReasoningStep struct {
	StepNumber   int
	Description  string
	Input        []string
	Process      string
	Output       string
	Confidence   float64
}

// Alternative represents an alternative consideration
type Alternative struct {
	Option       string
	Pros         []string
	Cons         []string
	Suitability  float64
	Rationale    string
}

// ConfidenceFactor affects confidence in recommendations
type ConfidenceFactor struct {
	Factor       string
	Impact       float64 // positive or negative
	Description  string
}

// KnowledgeBase provides access to learned knowledge
type KnowledgeBase interface {
	// Store learned information
	Store(ctx context.Context, knowledge *Knowledge) error
	
	// Retrieve relevant knowledge
	Retrieve(ctx context.Context, query *KnowledgeQuery) ([]*Knowledge, error)
	
	// Update existing knowledge
	Update(ctx context.Context, knowledge *Knowledge) error
	
	// Delete knowledge
	Delete(ctx context.Context, id string) error
	
	// Search knowledge
	Search(ctx context.Context, query string) ([]*Knowledge, error)
}

// Knowledge represents a piece of learned knowledge
type Knowledge struct {
	ID           string
	Type         KnowledgeType
	Content      string
	Source       string
	Timestamp    time.Time
	Confidence   float64
	Tags         []string
	Metadata     map[string]interface{}
}

// KnowledgeType categorizes knowledge
type KnowledgeType string

const (
	KnowledgePattern       KnowledgeType = "pattern"
	KnowledgeVulnerability KnowledgeType = "vulnerability"
	KnowledgeDefense       KnowledgeType = "defense"
	KnowledgeStrategy      KnowledgeType = "strategy"
	KnowledgeInsight       KnowledgeType = "insight"
	KnowledgeBestPractice  KnowledgeType = "best_practice"
)

// KnowledgeQuery specifies knowledge retrieval criteria
type KnowledgeQuery struct {
	Type         KnowledgeType
	Tags         []string
	Content      string
	MinConfidence float64
	MaxResults   int
	SortBy       string
}