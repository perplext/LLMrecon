package compliance

import (
	"encoding/json"
	"fmt"
)

// ISO42001Standard represents the ISO/IEC 42001:2023 standard
type ISO42001Standard struct {
	Name        string
	Version     string
	Description string
	Clauses     []Clause

}
// Clause represents a clause in ISO 42001
type Clause struct {
	ID           string        `json:"id"`
	Number       string        `json:"number"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Controls     []Control     `json:"controls"`
	Requirements []Requirement `json:"requirements"`

}
// Control represents a control within a clause
type Control struct {
	ID              string                 `json:"id"`
	ClauseNumber    string                 `json:"clauseNumber"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"` // technical, organizational, documentation
	Implementation  ImplementationStatus   `json:"implementation"`
	Evidence        []Evidence             `json:"evidence"`
	Gaps            []Gap                  `json:"gaps"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`

}
// Requirement represents a specific requirement
type Requirement struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
	Verifiable  bool   `json:"verifiable"`

}
// ImplementationStatus represents the status of control implementation
type ImplementationStatus struct {
	Status         string    `json:"status"` // not_implemented, partial, implemented, verified
	Percentage     float64   `json:"percentage"`
	LastAssessed   time.Time `json:"lastAssessed"`
	AssessedBy     string    `json:"assessedBy"`
	EffectiveDate  time.Time `json:"effectiveDate"`
	ExpirationDate time.Time `json:"expirationDate,omitempty"`

}
// Evidence represents evidence for control implementation
type Evidence struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // document, log, scan_result, test_result
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Location    string                 `json:"location"`
	Date        time.Time              `json:"date"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`

}
// Gap represents a compliance gap
type Gap struct {
	ID          string    `json:"id"`
	ControlID   string    `json:"controlId"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // critical, high, medium, low
	DueDate     time.Time `json:"dueDate"`
	Status      string    `json:"status"` // open, in_progress, closed

}
// ComplianceReport represents a compliance assessment report
type ComplianceReport struct {
	Standard          string                      `json:"standard"`
	AssessmentDate    time.Time                   `json:"assessmentDate"`
	Results           map[string]*AssessmentResult `json:"results"`
	OverallCompliance float64                     `json:"overallCompliance"`
	Recommendations   []Recommendation            `json:"recommendations"`
	ExecutiveSummary  string                      `json:"executiveSummary"`

}
// Recommendation represents a compliance recommendation
type Recommendation struct {
	ID          string   `json:"id"`
	Priority    string   `json:"priority"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
	Timeline    string   `json:"timeline"`
}

}
// ISO42001Summary represents a compliance summary for ISO 42001
type ISO42001Summary struct {
	TotalControls        int `json:"totalControls"`
	CompliantControls    int `json:"compliantControls"`
	PartialControls      int `json:"partialControls"`
	NonCompliantControls int `json:"nonCompliantControls"`
	CriticalGaps         int `json:"criticalGaps"`
	HighGaps             int `json:"highGaps"`
	MediumGaps           int `json:"mediumGaps"`
	LowGaps              int `json:"lowGaps"`

}
// ISO42001Compliance provides ISO 42001 compliance checking
type ISO42001Compliance struct {
	standard         *ISO42001Standard
	controls         map[string]*Control
	evidenceStore    EvidenceStore
	assessmentEngine AssessmentEngine

}
// EvidenceStore interface for storing and retrieving evidence
type EvidenceStore interface {
	Store(evidence Evidence) error
	Retrieve(controlID string) ([]Evidence, error)
	Search(criteria map[string]interface{}) ([]Evidence, error)

// AssessmentEngine interface for assessing controls
}
type AssessmentEngine interface {
	Assess(control *Control, evidence []Evidence) (*AssessmentResult, error)
	CalculateCompliance(results []*AssessmentResult) float64

// Finding represents a compliance finding
}
type Finding struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	ControlID   string    `json:"controlId"`
	Timestamp   time.Time `json:"timestamp"`

}
// AssessmentResult represents the result of a control assessment
type AssessmentResult struct {
	ControlID        string                 `json:"controlId"`
	Status           string                 `json:"status"`
	ComplianceScore  float64                `json:"complianceScore"`
	Findings         []Finding              `json:"findings"`
	Recommendations  []string               `json:"recommendations"`
	AssessmentDate   time.Time              `json:"assessmentDate"`
	NextReviewDate   time.Time              `json:"nextReviewDate"`
	Details          map[string]interface{} `json:"details"`

}
// NewISO42001Compliance creates a new ISO 42001 compliance checker
func NewISO42001Compliance() *ISO42001Compliance {
	return &ISO42001Compliance{
		standard: initializeISO42001Standard(),
		controls: make(map[string]*Control),
	}

// initializeISO42001Standard initializes the ISO 42001 standard structure
}
func initializeISO42001Standard() *ISO42001Standard {
	return &ISO42001Standard{
		Name:        "ISO/IEC 42001",
		Version:     "2023",
		Description: "Information technology — Artificial intelligence — Management system",
		Clauses: []Clause{
			{
				Number:      "4",
				Title:       "Context of the organization",
				Description: "Understanding organizational context and stakeholder needs",
				Controls:    getContextControls(),
			},
			{
				Number:      "5",
				Title:       "Leadership",
				Description: "Leadership and commitment for AI management system",
				Controls:    getLeadershipControls(),
			},
			{
				Number:      "6",
				Title:       "Planning",
				Description: "Risk management and objective planning",
				Controls:    getPlanningControls(),
			},
			{
				Number:      "7",
				Title:       "Support",
				Description: "Resources, competence, and awareness",
				Controls:    getSupportControls(),
			},
			{
				Number:      "8",
				Title:       "Operation",
				Description: "Operational planning and control",
				Controls:    getOperationControls(),
			},
			{
				Number:      "9",
				Title:       "Performance evaluation",
				Description: "Monitoring, measurement, and evaluation",
				Controls:    getPerformanceControls(),
			},
			{
				Number:      "10",
				Title:       "Improvement",
				Description: "Continuous improvement",
				Controls:    getImprovementControls(),
			},
		},
	}

// CheckCompliance performs a comprehensive compliance check
}
func (iso *ISO42001Compliance) CheckCompliance() (*ComplianceReport, error) {
	report := &ComplianceReport{
		Standard:       "ISO/IEC 42001:2023",
		AssessmentDate: time.Now(),
		Results:        make(map[string]*AssessmentResult),
	}

	// Assess each control
	for _, clause := range iso.standard.Clauses {
		for _, control := range clause.Controls {
			result, err := iso.assessControl(&control)
			if err != nil {
				return nil, fmt.Errorf("failed to assess control %s: %w", control.ID, err)
			}
			report.Results[control.ID] = result
		}
	}

	// Calculate overall compliance
	report.OverallCompliance = iso.calculateOverallCompliance(report.Results)
	report.ExecutiveSummary = iso.generateExecutiveSummary(report.Results)
	report.Recommendations = iso.generateRecommendations(report.Results)

	return report, nil

// assessControl assesses a single control
}
func (iso *ISO42001Compliance) assessControl(control *Control) (*AssessmentResult, error) {
	// Retrieve evidence for the control
	evidence, err := iso.evidenceStore.Retrieve(control.ID)
	if err != nil {
		return nil, err
	}

	// Use assessment engine to evaluate
	result, err := iso.assessmentEngine.Assess(control, evidence)
	if err != nil {
		return nil, err
	}

	// Add control-specific logic
	switch control.Type {
	case "technical":
		iso.assessTechnicalControl(control, result)
	case "organizational":
		iso.assessOrganizationalControl(control, result)
	case "documentation":
		iso.assessDocumentationControl(control, result)
	}

	return result, nil

// assessTechnicalControl performs technical control assessment
}
func (iso *ISO42001Compliance) assessTechnicalControl(control *Control, result *AssessmentResult) {
	// Check for security scan results
	scanEvidence := iso.filterEvidence(result.ControlID, "scan_result")
	if len(scanEvidence) == 0 {
		result.Findings = append(result.Findings, Finding{
			ID:          fmt.Sprintf("FIND-%s-%d", result.ControlID, time.Now().Unix()),
			Severity:    "high",
			Description: "No security scan results found",
			Impact:      "Control effectiveness cannot be verified without security scan results",
			ControlID:   result.ControlID,
			Timestamp:   time.Now(),
		})
		result.ComplianceScore *= 0.7
	}

	// Check for recent testing
	latestTest := iso.findLatestEvidence(result.ControlID, "test_result")
	if latestTest != nil && time.Since(latestTest.Date) > 90*24*time.Hour {
		result.Findings = append(result.Findings, Finding{
			ID:          "outdated_testing",
			Severity:    "medium",
			Description: "Testing results are older than 90 days",
			ControlID:   control.ID,
			Timestamp:   time.Now(),
		})
		result.ComplianceScore *= 0.85
	}

// assessOrganizationalControl performs organizational control assessment
}
func (iso *ISO42001Compliance) assessOrganizationalControl(control *Control, result *AssessmentResult) {
	// Check for policy documents
	policyEvidence := iso.filterEvidence(result.ControlID, "document")
	if len(policyEvidence) == 0 {
		result.Findings = append(result.Findings, Finding{
			ID:          "missing_policy",
			Severity:    "critical",
			Description: "Required policy documentation not found",
			ControlID:   control.ID,
			Timestamp:   time.Now(),
		})
		result.ComplianceScore *= 0.5
	}

	// Check for approval records
	for _, evidence := range policyEvidence {
		if evidence.Metadata["approved"] != true {
			result.Findings = append(result.Findings, Finding{
				ID:          "unapproved_policy",
				Severity:    "high",
				Description: fmt.Sprintf("Policy %s lacks approval", evidence.Title),
			})
			result.ComplianceScore *= 0.8
		}
	}

// assessDocumentationControl performs documentation control assessment
}
func (iso *ISO42001Compliance) assessDocumentationControl(control *Control, result *AssessmentResult) {
	// Check for required documentation
	docs := iso.filterEvidence(result.ControlID, "document")
	
	// Check document currency
	for _, doc := range docs {
		if time.Since(doc.Date) > 365*24*time.Hour {
			result.Findings = append(result.Findings, Finding{
				ID:          "outdated_documentation",
				Severity:    "medium",
				Description: fmt.Sprintf("Document %s is over 1 year old", doc.Title),
			})
			result.ComplianceScore *= 0.9
		}
	}

	// Check for version control
	for _, doc := range docs {
		if doc.Metadata["version"] == nil {
			result.Findings = append(result.Findings, Finding{
				ID:          "missing_version",
				Severity:    "low",
				Description: fmt.Sprintf("Document %s lacks version information", doc.Title),
			})
			result.ComplianceScore *= 0.95
		}
	}

// calculateOverallCompliance calculates the overall compliance percentage
}
func (iso *ISO42001Compliance) calculateOverallCompliance(results map[string]*AssessmentResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, result := range results {
		totalScore += result.ComplianceScore
	}

	return (totalScore / float64(len(results))) * 100

// generateExecutiveSummary generates an executive summary string
}
func (iso *ISO42001Compliance) generateExecutiveSummary(results map[string]*AssessmentResult) string {
	summary := iso.generateSummary(results)
	return fmt.Sprintf("ISO 42001 Compliance Assessment: Overall compliance score is %.1f%%. "+
		"Total controls assessed: %d, Compliant: %d, Partial: %d, Non-compliant: %d. "+
		"Critical gaps: %d, High risk gaps: %d.",
		iso.calculateOverallCompliance(results),
		summary.TotalControls, summary.CompliantControls, 
		summary.PartialControls, summary.NonCompliantControls,
		summary.CriticalGaps, summary.HighGaps)

// generateSummary generates a compliance summary
}
func (iso *ISO42001Compliance) generateSummary(results map[string]*AssessmentResult) ISO42001Summary {
	summary := ISO42001Summary{
		TotalControls:      len(results),
		CompliantControls:  0,
		PartialControls:    0,
		NonCompliantControls: 0,
		CriticalGaps:       0,
		HighGaps:           0,
		MediumGaps:         0,
		LowGaps:            0,
	}

	for _, result := range results {
		switch result.Status {
		case "compliant":
			summary.CompliantControls++
		case "partial":
			summary.PartialControls++
		case "non_compliant":
			summary.NonCompliantControls++
		}

		for _, finding := range result.Findings {
			switch finding.Severity {
			case "critical":
				summary.CriticalGaps++
			case "high":
				summary.HighGaps++
			case "medium":
				summary.MediumGaps++
			case "low":
				summary.LowGaps++
			}
		}
	}

	return summary

// generateRecommendations generates prioritized recommendations
}
func (iso *ISO42001Compliance) generateRecommendations(results map[string]*AssessmentResult) []Recommendation {
	recommendations := []Recommendation{}

	// Collect all recommendations with priority
	for controlID, result := range results {
		for _, rec := range result.Recommendations {
			priority := iso.calculatePriority(result)
			recommendations = append(recommendations, Recommendation{
				ID:          fmt.Sprintf("REC-%s-%d", controlID, len(recommendations)),
				Description: rec,
				Priority:    priority,
				Timeline:    iso.calculateTimeline(priority),
			})
		}
	}

	// Sort by priority
	sortRecommendationsByPriority(recommendations)

	return recommendations

// Helper functions for control definitions
}
func getContextControls() []Control {
	return []Control{
		{
			ID:           "4.1.1",
			ClauseNumber: "4.1",
			Title:        "Understanding organizational context",
			Description:  "Determine external and internal issues relevant to AI management",
			Type:         "organizational",
		},
		{
			ID:           "4.2.1",
			ClauseNumber: "4.2",
			Title:        "Understanding stakeholder needs",
			Description:  "Identify stakeholders and their requirements for AI systems",
			Type:         "organizational",
		},
	}

}
func getLeadershipControls() []Control {
	return []Control{
		{
			ID:           "5.1.1",
			ClauseNumber: "5.1",
			Title:        "Leadership commitment",
			Description:  "Top management demonstrates leadership for AI governance",
			Type:         "organizational",
		},
		{
			ID:           "5.2.1",
			ClauseNumber: "5.2",
			Title:        "AI policy",
			Description:  "Establish and communicate AI policy",
			Type:         "documentation",
		},
	}

}
func getPlanningControls() []Control {
	return []Control{
		{
			ID:           "6.1.1",
			ClauseNumber: "6.1",
			Title:        "Risk assessment",
			Description:  "Identify and assess AI-related risks",
			Type:         "technical",
		},
		{
			ID:           "6.1.2",
			ClauseNumber: "6.1",
			Title:        "Risk treatment",
			Description:  "Plan and implement risk treatment measures",
			Type:         "technical",
		},
		{
			ID:           "6.2.1",
			ClauseNumber: "6.2",
			Title:        "AI objectives",
			Description:  "Establish AI objectives and planning to achieve them",
			Type:         "organizational",
		},
	}

}
func getSupportControls() []Control {
	return []Control{
		{
			ID:           "7.1.1",
			ClauseNumber: "7.1",
			Title:        "Resources",
			Description:  "Determine and provide necessary resources",
			Type:         "organizational",
		},
		{
			ID:           "7.2.1",
			ClauseNumber: "7.2",
			Title:        "Competence",
			Description:  "Ensure AI team competence",
			Type:         "organizational",
		},
		{
			ID:           "7.3.1",
			ClauseNumber: "7.3",
			Title:        "Awareness",
			Description:  "Ensure awareness of AI management system",
			Type:         "organizational",
		},
	}

}
func getOperationControls() []Control {
	return []Control{
		{
			ID:           "8.1.1",
			ClauseNumber: "8.1",
			Title:        "Operational planning",
			Description:  "Plan, implement and control AI operations",
			Type:         "technical",
		},
		{
			ID:           "8.2.1",
			ClauseNumber: "8.2",
			Title:        "AI system requirements",
			Description:  "Determine requirements for AI systems",
			Type:         "technical",
		},
		{
			ID:           "8.3.1",
			ClauseNumber: "8.3",
			Title:        "AI system design and development",
			Description:  "Control AI system design and development",
			Type:         "technical",
		},
	}

}
func getPerformanceControls() []Control {
	return []Control{
		{
			ID:           "9.1.1",
			ClauseNumber: "9.1",
			Title:        "Monitoring and measurement",
			Description:  "Monitor and measure AI system performance",
			Type:         "technical",
		},
		{
			ID:           "9.2.1",
			ClauseNumber: "9.2",
			Title:        "Internal audit",
			Description:  "Conduct internal audits at planned intervals",
			Type:         "organizational",
		},
		{
			ID:           "9.3.1",
			ClauseNumber: "9.3",
			Title:        "Management review",
			Description:  "Review AI management system at planned intervals",
			Type:         "organizational",
		},
	}

}
func getImprovementControls() []Control {
	return []Control{
		{
			ID:           "10.1.1",
			ClauseNumber: "10.1",
			Title:        "Continual improvement",
			Description:  "Continually improve AI management system",
			Type:         "organizational",
		},
		{
			ID:           "10.2.1",
			ClauseNumber: "10.2",
			Title:        "Nonconformity and corrective action",
			Description:  "React to nonconformity and take corrective action",
			Type:         "organizational",
		},
	}

// Helper methods
}
func (iso *ISO42001Compliance) filterEvidence(controlID, evidenceType string) []Evidence {
	allEvidence, _ := iso.evidenceStore.Retrieve(controlID)
	filtered := []Evidence{}
	for _, e := range allEvidence {
		if e.Type == evidenceType {
			filtered = append(filtered, e)
		}
	}
	return filtered

func (iso *ISO42001Compliance) findLatestEvidence(controlID, evidenceType string) *Evidence {
	evidence := iso.filterEvidence(controlID, evidenceType)
	if len(evidence) == 0 {
		return nil
	}
	
	latest := &evidence[0]
	for i := 1; i < len(evidence); i++ {
		if evidence[i].Date.After(latest.Date) {
			latest = &evidence[i]
		}
	}
	return latest

func (iso *ISO42001Compliance) calculatePriority(result *AssessmentResult) string {
	// Priority based on compliance score and finding severity
	if result.ComplianceScore < 50 {
		return "critical"
	}
	
	criticalCount := 0
	highCount := 0
	for _, finding := range result.Findings {
		switch finding.Severity {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}
	
	if criticalCount > 0 {
		return "critical"
	} else if highCount > 2 {
		return "high"
	} else if result.ComplianceScore < 70 {
		return "medium"
	}
	return "low"

func (iso *ISO42001Compliance) calculateDueDate(priority string) time.Time {
	now := time.Now()
	switch priority {
	case "critical":
		return now.AddDate(0, 0, 30)  // 30 days
	case "high":
		return now.AddDate(0, 0, 60)  // 60 days
	case "medium":
		return now.AddDate(0, 0, 90)  // 90 days
	default:
		return now.AddDate(0, 0, 180) // 180 days
	}

}
func (iso *ISO42001Compliance) calculateTimeline(priority string) string {
	switch priority {
	case "critical":
		return "30 days"
	case "high":
		return "60 days"
	case "medium":
		return "90 days"
	default:
		return "180 days"
	}

}
func sortRecommendationsByPriority(recommendations []Recommendation) {
	priorityOrder := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}
	
	// Simple bubble sort for demonstration
	for i := 0; i < len(recommendations); i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority] {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

// ExportReport exports the compliance report in various formats
}
func (iso *ISO42001Compliance) ExportReport(report *ComplianceReport, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(report, "", "  ")
	case "summary":
		return iso.generateTextSummary(report), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

// generateTextSummary generates a text summary of the report
}
func (iso *ISO42001Compliance) generateTextSummary(report *ComplianceReport) []byte {
	// Generate summary from results
	summaryData := iso.generateSummary(report.Results)
	
	summary := fmt.Sprintf(`ISO/IEC 42001:2023 Compliance Report
====================================
Assessment Date: %s
Overall Compliance: %.1f%%

Summary:
- Total Controls: %d
- Compliant: %d
- Partial: %d  
- Non-Compliant: %d

Critical Gaps: %d
High Priority Recommendations: %d

Top Recommendations:
`, 
		report.AssessmentDate.Format("2006-01-02"),
		report.OverallCompliance,
		summaryData.TotalControls,
		summaryData.CompliantControls,
		summaryData.PartialControls,
		summaryData.NonCompliantControls,
		summaryData.CriticalGaps,
		len(report.Recommendations),
	)

	// Add top 5 recommendations
	for i, rec := range report.Recommendations {
		if i >= 5 {
			break
		}
		summary += fmt.Sprintf("%d. [%s] %s\n", i+1, rec.Priority, rec.Description)
	}

