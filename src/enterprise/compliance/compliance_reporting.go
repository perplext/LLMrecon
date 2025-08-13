package compliance

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

// ComplianceReportingSystem manages compliance reporting for LLM security testing
type ComplianceReportingSystem struct {
    mu              sync.RWMutex
    frameworks      map[string]*ComplianceFramework
    assessments     map[string]*Assessment
    reports         map[string]*ComplianceReport
    controls        map[string]*SecurityControl
    evidence        map[string]*Evidence
    auditor         *ComplianceAuditor
    generator       *ReportGenerator
    tracker         *ComplianceTracker
    repository      *ComplianceRepository
    config          ComplianceConfig
}

// ComplianceConfig holds configuration for compliance reporting
type ComplianceConfig struct {
    EnabledFrameworks   []string
    ReportingFrequency  time.Duration
    AutomatedAssessment bool
    EvidenceRetention   time.Duration
    ReportFormats       []ReportFormat
}

// ComplianceFramework represents a compliance framework
type ComplianceFramework struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Version         string                 `json:"version"`
    Description     string                 `json:"description"`
    Authority       string                 `json:"authority"`
    Categories      []*Category            `json:"categories"`
    Requirements    []*Requirement         `json:"requirements"`
    Controls        map[string]*Control    `json:"controls"`
    MappingRules    []*MappingRule         `json:"mapping_rules"`
    LastUpdated     time.Time             `json:"last_updated"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// Category represents a compliance category
type Category struct {
    ID              string          `json:"id"`
    Name            string          `json:"name"`
    Description     string          `json:"description"`
    Requirements    []string        `json:"requirements"`
    Priority        Priority        `json:"priority"`
    Parent          string          `json:"parent,omitempty"`
}

// Requirement represents a compliance requirement
type Requirement struct {
    ID              string                 `json:"id"`
    Title           string                 `json:"title"`
    Description     string                 `json:"description"`
    Category        string                 `json:"category"`
    Level           RequirementLevel       `json:"level"`
    Controls        []string               `json:"controls"`
    TestProcedures  []*TestProcedure       `json:"test_procedures"`
    Evidence        []EvidenceType         `json:"evidence_required"`
    References      []string               `json:"references"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// RequirementLevel defines requirement levels
type RequirementLevel string

const (
    RequirementMandatory    RequirementLevel = "mandatory"
    RequirementRecommended  RequirementLevel = "recommended"
    RequirementOptional     RequirementLevel = "optional"
)

// Priority defines priority levels
type Priority string

const (
    PriorityCritical Priority = "critical"
    PriorityHigh     Priority = "high"
    PriorityMedium   Priority = "medium"
    PriorityLow      Priority = "low"
)

// Control represents a security control
type Control struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Type            ControlType            `json:"type"`
    Implementation  string                 `json:"implementation"`
    Effectiveness   float64                `json:"effectiveness"`
    Status          ControlStatus          `json:"status"`
    LastTested      time.Time             `json:"last_tested"`
    TestResults     []*TestResult          `json:"test_results"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// ControlType defines control types
type ControlType string

const (
    ControlPreventive   ControlType = "preventive"
    ControlDetective    ControlType = "detective"
    ControlCorrective   ControlType = "corrective"
    ControlCompensating ControlType = "compensating"
)

// ControlStatus defines control status
type ControlStatus string

const (
    ControlImplemented     ControlStatus = "implemented"
    ControlPartiallyImplemented ControlStatus = "partially_implemented"
    ControlNotImplemented  ControlStatus = "not_implemented"
    ControlNotApplicable   ControlStatus = "not_applicable"
)

// TestProcedure represents a test procedure
type TestProcedure struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Steps           []TestStep             `json:"steps"`
    ExpectedResults []string               `json:"expected_results"`
    ActualResults   []string               `json:"actual_results,omitempty"`
    Status          TestStatus             `json:"status"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// TestStep represents a test step
type TestStep struct {
    Order       int    `json:"order"`
    Action      string `json:"action"`
    Expected    string `json:"expected"`
    Actual      string `json:"actual,omitempty"`
    Passed      bool   `json:"passed"`
}

// TestStatus defines test status
type TestStatus string

const (
    TestPending     TestStatus = "pending"
    TestInProgress  TestStatus = "in_progress"
    TestPassed      TestStatus = "passed"
    TestFailed      TestStatus = "failed"
    TestSkipped     TestStatus = "skipped"
)

// TestResult represents a test result
type TestResult struct {
    ID              string                 `json:"id"`
    TestDate        time.Time             `json:"test_date"`
    Tester          string                 `json:"tester"`
    Status          TestStatus             `json:"status"`
    Findings        []string               `json:"findings"`
    Evidence        []string               `json:"evidence"`
    Recommendations []string               `json:"recommendations"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// EvidenceType defines evidence types
type EvidenceType string

const (
    EvidenceScreenshot      EvidenceType = "screenshot"
    EvidenceLog             EvidenceType = "log"
    EvidenceConfiguration   EvidenceType = "configuration"
    EvidenceTestResult      EvidenceType = "test_result"
    EvidenceDocument        EvidenceType = "document"
    EvidenceAttestation     EvidenceType = "attestation"
)

// MappingRule maps requirements to controls
type MappingRule struct {
    ID              string                 `json:"id"`
    Source          string                 `json:"source"`
    Target          string                 `json:"target"`
    Relationship    string                 `json:"relationship"`
    Strength        float64                `json:"strength"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// Assessment represents a compliance assessment
type Assessment struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Framework       string                 `json:"framework"`
    Scope           *AssessmentScope       `json:"scope"`
    Status          AssessmentStatus       `json:"status"`
    StartDate       time.Time             `json:"start_date"`
    EndDate         *time.Time            `json:"end_date,omitempty"`
    Assessor        string                 `json:"assessor"`
    Results         *AssessmentResults     `json:"results"`
    Findings        []*Finding             `json:"findings"`
    Recommendations []*Recommendation      `json:"recommendations"`
    Evidence        map[string]*Evidence   `json:"evidence"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// AssessmentScope defines assessment scope
type AssessmentScope struct {
    Systems         []string `json:"systems"`
    Components      []string `json:"components"`
    Requirements    []string `json:"requirements"`
    ExcludedItems   []string `json:"excluded_items"`
    TimeFrame       string   `json:"timeframe"`
}

// AssessmentStatus defines assessment status
type AssessmentStatus string

const (
    AssessmentPlanned      AssessmentStatus = "planned"
    AssessmentInProgress   AssessmentStatus = "in_progress"
    AssessmentCompleted    AssessmentStatus = "completed"
    AssessmentReview       AssessmentStatus = "review"
    AssessmentApproved     AssessmentStatus = "approved"
)

// AssessmentResults contains assessment results
type AssessmentResults struct {
    ComplianceScore     float64                `json:"compliance_score"`
    RequirementsMet     int                    `json:"requirements_met"`
    RequirementsTotal   int                    `json:"requirements_total"`
    ControlsEffective   int                    `json:"controls_effective"`
    ControlsTotal       int                    `json:"controls_total"`
    RiskLevel           RiskLevel              `json:"risk_level"`
    Gaps                []*ComplianceGap       `json:"gaps"`
    Strengths           []string               `json:"strengths"`
    Weaknesses          []string               `json:"weaknesses"`
    Metrics             map[string]interface{} `json:"metrics"`
}

// RiskLevel defines risk levels
type RiskLevel string

const (
    RiskCritical RiskLevel = "critical"
    RiskHigh     RiskLevel = "high"
    RiskMedium   RiskLevel = "medium"
    RiskLow      RiskLevel = "low"
    RiskMinimal  RiskLevel = "minimal"
)

// ComplianceGap represents a compliance gap
type ComplianceGap struct {
    ID              string                 `json:"id"`
    Requirement     string                 `json:"requirement"`
    Description     string                 `json:"description"`
    Impact          string                 `json:"impact"`
    RemediationPlan string                 `json:"remediation_plan"`
    Priority        Priority               `json:"priority"`
    DueDate         *time.Time            `json:"due_date,omitempty"`
    Owner           string                 `json:"owner"`
    Status          GapStatus              `json:"status"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// GapStatus defines gap status
type GapStatus string

const (
    GapIdentified   GapStatus = "identified"
    GapInProgress   GapStatus = "in_progress"
    GapRemediated   GapStatus = "remediated"
    GapAccepted     GapStatus = "accepted"
)

// Finding represents an assessment finding
type Finding struct {
    ID              string                 `json:"id"`
    Title           string                 `json:"title"`
    Description     string                 `json:"description"`
    Severity        SeverityLevel          `json:"severity"`
    Category        string                 `json:"category"`
    Requirement     string                 `json:"requirement"`
    Evidence        []string               `json:"evidence"`
    Impact          string                 `json:"impact"`
    Recommendation  string                 `json:"recommendation"`
    Status          FindingStatus          `json:"status"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// SeverityLevel defines severity levels
type SeverityLevel string

const (
    SeverityCritical SeverityLevel = "critical"
    SeverityHigh     SeverityLevel = "high"
    SeverityMedium   SeverityLevel = "medium"
    SeverityLow      SeverityLevel = "low"
    SeverityInfo     SeverityLevel = "info"
)

// FindingStatus defines finding status
type FindingStatus string

const (
    FindingOpen         FindingStatus = "open"
    FindingInProgress   FindingStatus = "in_progress"
    FindingRemediated   FindingStatus = "remediated"
    FindingAccepted     FindingStatus = "accepted"
    FindingFalsePositive FindingStatus = "false_positive"
)

// Recommendation represents a recommendation
type Recommendation struct {
    ID              string                 `json:"id"`
    Title           string                 `json:"title"`
    Description     string                 `json:"description"`
    Priority        Priority               `json:"priority"`
    Category        string                 `json:"category"`
    Implementation  string                 `json:"implementation"`
    Benefits        []string               `json:"benefits"`
    Resources       []string               `json:"resources"`
    Timeline        string                 `json:"timeline"`
    Owner           string                 `json:"owner"`
    Status          RecommendationStatus   `json:"status"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// RecommendationStatus defines recommendation status
type RecommendationStatus string

const (
    RecommendationProposed      RecommendationStatus = "proposed"
    RecommendationApproved      RecommendationStatus = "approved"
    RecommendationImplementing  RecommendationStatus = "implementing"
    RecommendationImplemented   RecommendationStatus = "implemented"
    RecommendationRejected      RecommendationStatus = "rejected"
)

// Evidence represents compliance evidence
type Evidence struct {
    ID              string                 `json:"id"`
    Type            EvidenceType           `json:"type"`
    Title           string                 `json:"title"`
    Description     string                 `json:"description"`
    Source          string                 `json:"source"`
    CollectionDate  time.Time             `json:"collection_date"`
    Collector       string                 `json:"collector"`
    Location        string                 `json:"location"`
    Hash            string                 `json:"hash"`
    Chain           []CustodyEntry         `json:"chain_of_custody"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// CustodyEntry represents chain of custody entry
type CustodyEntry struct {
    Timestamp   time.Time `json:"timestamp"`
    Actor       string    `json:"actor"`
    Action      string    `json:"action"`
    Description string    `json:"description"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
    ID              string                 `json:"id"`
    Title           string                 `json:"title"`
    Type            ReportType             `json:"type"`
    Framework       string                 `json:"framework"`
    Period          ReportPeriod           `json:"period"`
    Executive       *ExecutiveSummary      `json:"executive_summary"`
    Assessment      string                 `json:"assessment_id"`
    Sections        []*ReportSection       `json:"sections"`
    Appendices      []*Appendix            `json:"appendices"`
    GeneratedAt     time.Time             `json:"generated_at"`
    GeneratedBy     string                 `json:"generated_by"`
    ApprovedBy      string                 `json:"approved_by,omitempty"`
    ApprovalDate    *time.Time            `json:"approval_date,omitempty"`
    Distribution    []string               `json:"distribution"`
    Classification  string                 `json:"classification"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// ReportType defines report types
type ReportType string

const (
    ReportCompliance    ReportType = "compliance"
    ReportAssessment    ReportType = "assessment"
    ReportAudit         ReportType = "audit"
    ReportIncident      ReportType = "incident"
    ReportExecutive     ReportType = "executive"
    ReportTechnical     ReportType = "technical"
)

// ReportFormat defines report formats
type ReportFormat string

const (
    FormatPDF       ReportFormat = "pdf"
    FormatHTML      ReportFormat = "html"
    FormatMarkdown  ReportFormat = "markdown"
    FormatJSON      ReportFormat = "json"
    FormatExcel     ReportFormat = "excel"
    FormatWord      ReportFormat = "word"
)

// ReportPeriod represents report period
type ReportPeriod struct {
    Start   time.Time `json:"start"`
    End     time.Time `json:"end"`
    Label   string    `json:"label"`
}

// ExecutiveSummary represents executive summary
type ExecutiveSummary struct {
    Overview        string                 `json:"overview"`
    KeyFindings     []string               `json:"key_findings"`
    ComplianceScore float64                `json:"compliance_score"`
    RiskLevel       RiskLevel              `json:"risk_level"`
    Recommendations []string               `json:"recommendations"`
    NextSteps       []string               `json:"next_steps"`
    Metrics         map[string]interface{} `json:"metrics"`
}

// ReportSection represents a report section
type ReportSection struct {
    ID              string                 `json:"id"`
    Title           string                 `json:"title"`
    Order           int                    `json:"order"`
    Content         string                 `json:"content"`
    Subsections     []*ReportSection       `json:"subsections,omitempty"`
    Tables          []*Table               `json:"tables,omitempty"`
    Charts          []*Chart               `json:"charts,omitempty"`
    References      []string               `json:"references"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// Table represents a data table
type Table struct {
    ID      string      `json:"id"`
    Title   string      `json:"title"`
    Headers []string    `json:"headers"`
    Rows    [][]string  `json:"rows"`
    Footer  []string    `json:"footer,omitempty"`
}

// Chart represents a chart
type Chart struct {
    ID      string                 `json:"id"`
    Title   string                 `json:"title"`
    Type    string                 `json:"type"`
    Data    interface{}            `json:"data"`
    Options map[string]interface{} `json:"options"`
}

// Appendix represents report appendix
type Appendix struct {
    ID      string  `json:"id"`
    Title   string  `json:"title"`
    Content string  `json:"content"`
    Type    string  `json:"type"`
}

// SecurityControl represents an implemented security control
type SecurityControl struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Category        string                 `json:"category"`
    Implementation  ImplementationDetails  `json:"implementation"`
    Testing         TestingDetails         `json:"testing"`
    Monitoring      MonitoringDetails      `json:"monitoring"`
    Documentation   []string               `json:"documentation"`
    Owner           string                 `json:"owner"`
    Status          ControlStatus          `json:"status"`
    LastReview      time.Time             `json:"last_review"`
    NextReview      time.Time             `json:"next_review"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// ImplementationDetails contains control implementation details
type ImplementationDetails struct {
    Method          string    `json:"method"`
    Technology      []string  `json:"technology"`
    Configuration   string    `json:"configuration"`
    Dependencies    []string  `json:"dependencies"`
    ImplementedDate time.Time `json:"implemented_date"`
}

// TestingDetails contains control testing details
type TestingDetails struct {
    Frequency       string       `json:"frequency"`
    LastTest        time.Time    `json:"last_test"`
    NextTest        time.Time    `json:"next_test"`
    TestProcedure   string       `json:"test_procedure"`
    Results         []TestResult `json:"results"`
}

// MonitoringDetails contains control monitoring details
type MonitoringDetails struct {
    Method          string   `json:"method"`
    Frequency       string   `json:"frequency"`
    Metrics         []string `json:"metrics"`
    Alerts          []string `json:"alerts"`
    Dashboard       string   `json:"dashboard"`
}

// ComplianceAuditor performs compliance audits
type ComplianceAuditor struct {
    frameworks  map[string]*ComplianceFramework
    controls    map[string]*SecurityControl
    mu          sync.RWMutex
}

// ReportGenerator generates compliance reports
type ReportGenerator struct {
    templates   map[string]*ReportTemplate
    formatters  map[ReportFormat]Formatter
    mu          sync.RWMutex
}

// ReportTemplate represents a report template
type ReportTemplate struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        ReportType             `json:"type"`
    Structure   []TemplateSection      `json:"structure"`
    Styles      map[string]interface{} `json:"styles"`
    Variables   []string               `json:"variables"`
}

// TemplateSection represents a template section
type TemplateSection struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`
    Required    bool     `json:"required"`
    Content     string   `json:"content"`
    Children    []string `json:"children"`
}

// Formatter interface for report formatting
type Formatter interface {
    Format(report *ComplianceReport) ([]byte, error)
}

// ComplianceTracker tracks compliance status
type ComplianceTracker struct {
    status      map[string]*ComplianceStatus
    history     map[string][]*StatusChange
    mu          sync.RWMutex
}

// ComplianceStatus represents current compliance status
type ComplianceStatus struct {
    Framework       string                 `json:"framework"`
    Score           float64                `json:"score"`
    Status          string                 `json:"status"`
    LastAssessment  time.Time             `json:"last_assessment"`
    NextAssessment  time.Time             `json:"next_assessment"`
    Trends          []TrendPoint           `json:"trends"`
    Metadata        map[string]interface{} `json:"metadata"`
}

// StatusChange represents a status change
type StatusChange struct {
    Timestamp   time.Time              `json:"timestamp"`
    OldStatus   string                 `json:"old_status"`
    NewStatus   string                 `json:"new_status"`
    Reason      string                 `json:"reason"`
    ChangedBy   string                 `json:"changed_by"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// TrendPoint represents a trend data point
type TrendPoint struct {
    Timestamp   time.Time `json:"timestamp"`
    Value       float64   `json:"value"`
    Label       string    `json:"label"`
}

// ComplianceRepository stores compliance data
type ComplianceRepository struct {
    storage     map[string]interface{}
    indices     map[string]map[string]string
    mu          sync.RWMutex
}

// NewComplianceReportingSystem creates a new compliance reporting system
func NewComplianceReportingSystem(config ComplianceConfig) *ComplianceReportingSystem {
    return &ComplianceReportingSystem{
        frameworks:  make(map[string]*ComplianceFramework),
        assessments: make(map[string]*Assessment),
        reports:     make(map[string]*ComplianceReport),
        controls:    make(map[string]*SecurityControl),
        evidence:    make(map[string]*Evidence),
        auditor:     NewComplianceAuditor(),
        generator:   NewReportGenerator(),
        tracker:     NewComplianceTracker(),
        repository:  NewComplianceRepository(),
        config:      config,
    }
}

// LoadFramework loads a compliance framework
func (crs *ComplianceReportingSystem) LoadFramework(ctx context.Context, framework *ComplianceFramework) error {
    crs.mu.Lock()
    defer crs.mu.Unlock()

    if framework.ID == "" {
        framework.ID = generateFrameworkID()
    }

    // Validate framework
    if err := crs.validateFramework(framework); err != nil {
        return fmt.Errorf("invalid framework: %w", err)
    }

    framework.LastUpdated = time.Now()
    crs.frameworks[framework.ID] = framework

    // Initialize framework in auditor
    crs.auditor.LoadFramework(framework)

    return nil
}

// StartAssessment starts a compliance assessment
func (crs *ComplianceReportingSystem) StartAssessment(ctx context.Context, assessment *Assessment) error {
    crs.mu.Lock()
    defer crs.mu.Unlock()

    if assessment.ID == "" {
        assessment.ID = generateAssessmentID()
    }

    assessment.Status = AssessmentInProgress
    assessment.StartDate = time.Now()

    // Validate assessment scope
    if err := crs.validateAssessmentScope(assessment); err != nil {
        return fmt.Errorf("invalid assessment scope: %w", err)
    }

    crs.assessments[assessment.ID] = assessment

    // Start automated assessment if enabled
    if crs.config.AutomatedAssessment {
        go crs.runAutomatedAssessment(ctx, assessment)
    }

    return nil
}

// GenerateReport generates a compliance report
func (crs *ComplianceReportingSystem) GenerateReport(ctx context.Context, assessmentID string, reportType ReportType, format ReportFormat) (*ComplianceReport, error) {
    crs.mu.RLock()
    assessment, exists := crs.assessments[assessmentID]
    crs.mu.RUnlock()

    if !exists {
        return nil, fmt.Errorf("assessment not found")
    }

    report := &ComplianceReport{
        ID:          generateReportID(),
        Title:       fmt.Sprintf("%s Compliance Report - %s", assessment.Framework, time.Now().Format("2006-01-02")),
        Type:        reportType,
        Framework:   assessment.Framework,
        Period: ReportPeriod{
            Start: assessment.StartDate,
            End:   time.Now(),
            Label: fmt.Sprintf("%s Assessment Period", assessment.Name),
        },
        Assessment:   assessmentID,
        GeneratedAt:  time.Now(),
        GeneratedBy:  "System",
        Distribution: []string{},
        Classification: "Internal",
    }

    // Generate executive summary
    report.Executive = crs.generateExecutiveSummary(assessment)

    // Generate report sections
    report.Sections = crs.generateReportSections(assessment, reportType)

    // Generate appendices
    report.Appendices = crs.generateAppendices(assessment)

    // Format report
    if formatted, err := crs.generator.FormatReport(report, format); err == nil {
        report.Metadata = map[string]interface{}{
            "formatted": formatted,
            "format":    format,
        }
    }

    crs.mu.Lock()
    crs.reports[report.ID] = report
    crs.mu.Unlock()

    return report, nil
}

// NewComplianceAuditor creates a new compliance auditor
func NewComplianceAuditor() *ComplianceAuditor {
    return &ComplianceAuditor{
        frameworks: make(map[string]*ComplianceFramework),
        controls:   make(map[string]*SecurityControl),
    }
}

// LoadFramework loads a framework into the auditor
func (ca *ComplianceAuditor) LoadFramework(framework *ComplianceFramework) {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    ca.frameworks[framework.ID] = framework
}

// AuditControl audits a security control
func (ca *ComplianceAuditor) AuditControl(control *SecurityControl, requirements []string) (*TestResult, error) {
    result := &TestResult{
        ID:       generateTestResultID(),
        TestDate: time.Now(),
        Tester:   "Automated Auditor",
        Status:   TestInProgress,
    }

    // Perform control testing
    findings := ca.testControl(control, requirements)
    
    if len(findings) == 0 {
        result.Status = TestPassed
    } else {
        result.Status = TestFailed
        result.Findings = findings
    }

    return result, nil
}

// testControl tests a security control
func (ca *ComplianceAuditor) testControl(control *SecurityControl, requirements []string) []string {
    var findings []string
    
    // Implement control testing logic
    if control.Status != ControlImplemented {
        findings = append(findings, fmt.Sprintf("Control %s is not fully implemented", control.ID))
    }
    
    if time.Since(control.LastReview) > 90*24*time.Hour {
        findings = append(findings, fmt.Sprintf("Control %s has not been reviewed in over 90 days", control.ID))
    }
    
    return findings
}

// NewReportGenerator creates a new report generator
func NewReportGenerator() *ReportGenerator {
    return &ReportGenerator{
        templates:  make(map[string]*ReportTemplate),
        formatters: make(map[ReportFormat]Formatter),
    }
}

// FormatReport formats a compliance report
func (rg *ReportGenerator) FormatReport(report *ComplianceReport, format ReportFormat) ([]byte, error) {
    rg.mu.RLock()
    formatter, exists := rg.formatters[format]
    rg.mu.RUnlock()

    if !exists {
        // Use default JSON formatter
        return json.MarshalIndent(report, "", "  ")
    }

    return formatter.Format(report)
}

// NewComplianceTracker creates a new compliance tracker
func NewComplianceTracker() *ComplianceTracker {
    return &ComplianceTracker{
        status:  make(map[string]*ComplianceStatus),
        history: make(map[string][]*StatusChange),
    }
}

// UpdateStatus updates compliance status
func (ct *ComplianceTracker) UpdateStatus(framework string, score float64, status string) {
    ct.mu.Lock()
    defer ct.mu.Unlock()

    current, exists := ct.status[framework]
    if !exists {
        current = &ComplianceStatus{
            Framework: framework,
            Trends:    []TrendPoint{},
        }
        ct.status[framework] = current
    }

    // Record status change
    if current.Status != status {
        change := &StatusChange{
            Timestamp: time.Now(),
            OldStatus: current.Status,
            NewStatus: status,
            ChangedBy: "System",
        }
        ct.history[framework] = append(ct.history[framework], change)
    }

    // Update current status
    current.Score = score
    current.Status = status
    current.LastAssessment = time.Now()
    current.Trends = append(current.Trends, TrendPoint{
        Timestamp: time.Now(),
        Value:     score,
        Label:     status,
    })
}

// NewComplianceRepository creates a new compliance repository
func NewComplianceRepository() *ComplianceRepository {
    return &ComplianceRepository{
        storage: make(map[string]interface{}),
        indices: make(map[string]map[string]string),
    }
}

// Store stores compliance data
func (cr *ComplianceRepository) Store(dataType, id string, data interface{}) error {
    cr.mu.Lock()
    defer cr.mu.Unlock()

    key := fmt.Sprintf("%s:%s", dataType, id)
    cr.storage[key] = data

    // Update indices
    if _, exists := cr.indices[dataType]; !exists {
        cr.indices[dataType] = make(map[string]string)
    }
    cr.indices[dataType][id] = key

    return nil
}

// Helper functions
func (crs *ComplianceReportingSystem) validateFramework(framework *ComplianceFramework) error {
    if framework.Name == "" {
        return fmt.Errorf("framework name is required")
    }
    if len(framework.Requirements) == 0 {
        return fmt.Errorf("framework must have requirements")
    }
    return nil
}

func (crs *ComplianceReportingSystem) validateAssessmentScope(assessment *Assessment) error {
    if assessment.Framework == "" {
        return fmt.Errorf("framework is required")
    }
    if assessment.Scope == nil || len(assessment.Scope.Systems) == 0 {
        return fmt.Errorf("assessment must have systems in scope")
    }
    return nil
}

func (crs *ComplianceReportingSystem) runAutomatedAssessment(ctx context.Context, assessment *Assessment) {
    // Implement automated assessment logic
    framework, exists := crs.frameworks[assessment.Framework]
    if !exists {
        return
    }

    // Test each requirement
    for _, requirement := range framework.Requirements {
        // Find applicable controls
        controls := crs.findControlsForRequirement(requirement.ID)
        
        // Test controls
        for _, control := range controls {
            result, _ := crs.auditor.AuditControl(control, []string{requirement.ID})
            
            // Store result
            if assessment.Results == nil {
                assessment.Results = &AssessmentResults{
                    Gaps:     []*ComplianceGap{},
                    Metrics:  make(map[string]interface{}),
                }
            }
            
            if result.Status == TestPassed {
                assessment.Results.RequirementsMet++
            }
        }
    }

    assessment.Status = AssessmentCompleted
    endTime := time.Now()
    assessment.EndDate = &endTime
}

func (crs *ComplianceReportingSystem) findControlsForRequirement(requirementID string) []*SecurityControl {
    var controls []*SecurityControl
    
    crs.mu.RLock()
    defer crs.mu.RUnlock()
    
    for _, control := range crs.controls {
        // Check if control maps to requirement
        controls = append(controls, control)
    }
    
    return controls
}

func (crs *ComplianceReportingSystem) generateExecutiveSummary(assessment *Assessment) *ExecutiveSummary {
    summary := &ExecutiveSummary{
        Overview:        fmt.Sprintf("Compliance assessment for %s framework", assessment.Framework),
        KeyFindings:     []string{},
        ComplianceScore: 0.0,
        RiskLevel:       RiskMedium,
        Recommendations: []string{},
        NextSteps:       []string{},
        Metrics:         make(map[string]interface{}),
    }

    if assessment.Results != nil {
        summary.ComplianceScore = assessment.Results.ComplianceScore
        summary.RiskLevel = assessment.Results.RiskLevel
        
        // Extract key findings
        for _, finding := range assessment.Findings {
            if finding.Severity == SeverityCritical || finding.Severity == SeverityHigh {
                summary.KeyFindings = append(summary.KeyFindings, finding.Title)
            }
        }
        
        // Extract recommendations
        for _, rec := range assessment.Recommendations {
            if rec.Priority == PriorityCritical || rec.Priority == PriorityHigh {
                summary.Recommendations = append(summary.Recommendations, rec.Title)
            }
        }
    }

    return summary
}

func (crs *ComplianceReportingSystem) generateReportSections(assessment *Assessment, reportType ReportType) []*ReportSection {
    sections := []*ReportSection{
        {
            ID:    "scope",
            Title: "Assessment Scope",
            Order: 1,
            Content: crs.generateScopeContent(assessment),
        },
        {
            ID:    "methodology",
            Title: "Assessment Methodology",
            Order: 2,
            Content: crs.generateMethodologyContent(assessment),
        },
        {
            ID:    "results",
            Title: "Assessment Results",
            Order: 3,
            Content: crs.generateResultsContent(assessment),
        },
        {
            ID:    "findings",
            Title: "Findings and Observations",
            Order: 4,
            Content: crs.generateFindingsContent(assessment),
        },
        {
            ID:    "recommendations",
            Title: "Recommendations",
            Order: 5,
            Content: crs.generateRecommendationsContent(assessment),
        },
    }

    return sections
}

func (crs *ComplianceReportingSystem) generateAppendices(assessment *Assessment) []*Appendix {
    return []*Appendix{
        {
            ID:      "evidence",
            Title:   "Supporting Evidence",
            Content: "Evidence details available upon request",
            Type:    "reference",
        },
        {
            ID:      "glossary",
            Title:   "Glossary of Terms",
            Content: "Standard compliance terminology",
            Type:    "reference",
        },
    }
}

func (crs *ComplianceReportingSystem) generateScopeContent(assessment *Assessment) string {
    return fmt.Sprintf("Assessment scope includes %d systems and %d requirements",
        len(assessment.Scope.Systems), len(assessment.Scope.Requirements))
}

func (crs *ComplianceReportingSystem) generateMethodologyContent(assessment *Assessment) string {
    return "Assessment conducted using automated testing and manual review procedures"
}

func (crs *ComplianceReportingSystem) generateResultsContent(assessment *Assessment) string {
    if assessment.Results != nil {
        return fmt.Sprintf("Compliance Score: %.2f%%, Requirements Met: %d/%d",
            assessment.Results.ComplianceScore,
            assessment.Results.RequirementsMet,
            assessment.Results.RequirementsTotal)
    }
    return "Results pending"
}

func (crs *ComplianceReportingSystem) generateFindingsContent(assessment *Assessment) string {
    return fmt.Sprintf("Total findings: %d", len(assessment.Findings))
}

func (crs *ComplianceReportingSystem) generateRecommendationsContent(assessment *Assessment) string {
    return fmt.Sprintf("Total recommendations: %d", len(assessment.Recommendations))
}

func generateFrameworkID() string {
    return fmt.Sprintf("framework_%d", time.Now().UnixNano())
}

func generateAssessmentID() string {
    return fmt.Sprintf("assessment_%d", time.Now().UnixNano())
}

func generateReportID() string {
    return fmt.Sprintf("report_%d", time.Now().UnixNano())
}

func generateTestResultID() string {
    return fmt.Sprintf("test_%d", time.Now().UnixNano())
}