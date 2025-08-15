// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
)

// EnhancedReportingSystem extends the ReportingSystem with more sophisticated reporting capabilities
type EnhancedReportingSystem struct {
	*ReportingSystem
	config              *ProtectionConfig
	reportingConfig     *EnhancedReportingConfig
	reports             map[string]*EnhancedInjectionReport
	reportHandlers      map[string]ReportHandlerFunc
	patternLibrary      *EnhancedInjectionPatternLibrary
	maxReports          int
	dataDir             string
	mu                  sync.RWMutex
}

// EnhancedReportingConfig defines the configuration for enhanced reporting
type EnhancedReportingConfig struct {
	EnableAutomaticReporting bool                   `json:"enable_automatic_reporting"`
	EnableReportSharing      bool                   `json:"enable_report_sharing"`
	EnableReportAnalysis     bool                   `json:"enable_report_analysis"`
	ReportingThreshold       float64                `json:"reporting_threshold"`
	ReportCategories         []string               `json:"report_categories"`
	ReportingEndpoints       map[string]string      `json:"reporting_endpoints"`
	AnalysisInterval         time.Duration          `json:"analysis_interval"`
}

// EnhancedInjectionReport extends the InjectionReport with more information
type EnhancedInjectionReport struct {
	*InjectionReport
	Status              ReportStatus        `json:"status"`
	Category            string              `json:"category"`
	AnalysisResults     map[string]interface{} `json:"analysis_results,omitempty"`
	RelatedReports      []string            `json:"related_reports,omitempty"`
	PatternMatches      int                 `json:"pattern_matches"`
	FalsePositiveRate   float64             `json:"false_positive_rate"`
	EffectivenessScore  float64             `json:"effectiveness_score"`
	CreatedBy           string              `json:"created_by"`
	LastUpdated         time.Time           `json:"last_updated"`
	Shared              bool                `json:"shared"`
	SharedWith          []string            `json:"shared_with,omitempty"`

// ReportStatus defines the status of a report
type ReportStatus string

const (
	// ReportStatusNew indicates a new report
	ReportStatusNew ReportStatus = "new"
	// ReportStatusAnalyzed indicates an analyzed report
	ReportStatusAnalyzed ReportStatus = "analyzed"
	// ReportStatusVerified indicates a verified report
	ReportStatusVerified ReportStatus = "verified"
	// ReportStatusRejected indicates a rejected report
	ReportStatusRejected ReportStatus = "rejected"
	// ReportStatusShared indicates a shared report
	ReportStatusShared ReportStatus = "shared"
)

// ReportHandlerFunc defines a function that handles reports
type ReportHandlerFunc func(context.Context, *EnhancedInjectionReport) error

// NewEnhancedReportingSystem creates a new enhanced reporting system
func NewEnhancedReportingSystem(config *ProtectionConfig, patternLibrary *EnhancedInjectionPatternLibrary, dataDir string) (*EnhancedReportingSystem, error) {
	baseSystem := NewReportingSystem(config)
	
	// Create the data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Initialize reporting config
	reportingConfig := &EnhancedReportingConfig{
		EnableAutomaticReporting: true,
		EnableReportSharing:      true,
		EnableReportAnalysis:     true,
		ReportingThreshold:       0.7,
		ReportCategories:         []string{"prompt_injection", "jailbreak", "role_change", "system_prompt", "boundary_violation", "delimiter_misuse", "unusual_pattern"},
		ReportingEndpoints:       make(map[string]string),
		AnalysisInterval:         time.Hour * 24,
	}
	
	// Set default reporting endpoints
	reportingConfig.ReportingEndpoints["local"] = "file://" + filepath.Join(dataDir, "reports")
	
	return &EnhancedReportingSystem{
		ReportingSystem:     baseSystem,
		config:              config,
		reportingConfig:     reportingConfig,
		reports:             make(map[string]*EnhancedInjectionReport),
		reportHandlers:      make(map[string]ReportHandlerFunc),
		patternLibrary:      patternLibrary,
		maxReports:          1000,
		dataDir:             dataDir,
	}, nil

// ReportInjectionEnhanced reports an injection technique with enhanced capabilities
func (r *EnhancedReportingSystem) ReportInjectionEnhanced(ctx context.Context, detections []*Detection, prompt string, response string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if there are any detections to report
	if len(detections) == 0 {
		return nil
	}
	
	// Group detections by type
	detectionsByType := make(map[DetectionType][]*Detection)
	for _, detection := range detections {
		detectionsByType[detection.Type] = append(detectionsByType[detection.Type], detection)
	}
	
	// Create a report for each detection type
	for detectionType, typeDetections := range detectionsByType {
		// Skip if below threshold
		maxConfidence := 0.0
		for _, detection := range typeDetections {
			if detection.Confidence > maxConfidence {
				maxConfidence = detection.Confidence
			}
		}
		
		if maxConfidence < r.reportingConfig.ReportingThreshold {
			continue
		}
		
		// Create report ID
		reportID := fmt.Sprintf("report-%d", time.Now().UnixNano())
		
		// Determine category
		category := r.determineReportCategory(detectionType)
		
		// Create example from the prompt and detection
		example := r.createExampleFromDetection(prompt, typeDetections[0])
		
		// Create base injection report
		baseReport := &InjectionReport{
			ReportID:      reportID,
			DetectionType: detectionType,
			Pattern:       typeDetections[0].Pattern,
			Example:       example,
			Confidence:    maxConfidence,
			Severity:      r.calculateSeverity(typeDetections),
			Description:   r.createDescriptionFromDetections(typeDetections),
			Timestamp:     time.Now(),
			Source:        "automatic",
			Metadata:      make(map[string]interface{}),
		}
		
		// Create enhanced injection report
		report := &EnhancedInjectionReport{
			InjectionReport:    baseReport,
			Status:             ReportStatusNew,
			Category:           category,
			AnalysisResults:    make(map[string]interface{}),
			RelatedReports:     make([]string, 0),
			PatternMatches:     len(typeDetections),
			FalsePositiveRate:  0.0,
			EffectivenessScore: maxConfidence,
			CreatedBy:          "system",
			LastUpdated:        time.Now(),
			Shared:             false,
			SharedWith:         make([]string, 0),
		}
		
		// Add to reports
		r.reports[reportID] = report
		
		// Trim if too many reports
		if len(r.reports) > r.maxReports {
			// Find oldest report
			var oldestID string
			var oldestTime time.Time
			for id, rep := range r.reports {
				if oldestID == "" || rep.Timestamp.Before(oldestTime) {
					oldestID = id
					oldestTime = rep.Timestamp
				}
			}
			
			// Remove oldest report
			if oldestID != "" {
				delete(r.reports, oldestID)
			}
		}
		
		// Save to disk
		r.saveReportToDisk(report)
		
		// Process with report handlers
		for _, handler := range r.reportHandlers {
			if err := handler(ctx, report); err != nil {
				// Log error but continue processing
				fmt.Printf("Error processing report: %v\n", err)
			}
		}
		
		// Add to pattern library if automatic reporting is enabled
		if r.reportingConfig.EnableAutomaticReporting && r.patternLibrary != nil {
			// Only add to pattern library if confidence is high enough
			if maxConfidence >= 0.8 {
				examples := []string{example}
				if err := r.patternLibrary.AddEmergingPattern(
					typeDetections[0].Pattern,
					r.createDescriptionFromDetections(typeDetections),
					"automatic",
					examples,
					maxConfidence,
				); err != nil {
					// Log error but continue processing
					fmt.Printf("Error adding pattern to library: %v\n", err)
				}
			}
		}
	}
	
	return nil

// determineReportCategory determines the category for a report
func (r *EnhancedReportingSystem) determineReportCategory(detectionType DetectionType) string {
	switch detectionType {
	case DetectionTypePromptInjection, DetectionTypeIndirectPromptInjection:
		return "prompt_injection"
	case DetectionTypeJailbreak:
		return "jailbreak"
	case DetectionTypeRoleChange:
		return "role_change"
	case DetectionTypeSystemPrompt:
		return "system_prompt"
	case DetectionTypeBoundaryViolation:
		return "boundary_violation"
	case DetectionTypeDelimiterMisuse:
		return "delimiter_misuse"
	case DetectionTypeUnusualPattern:
		return "unusual_pattern"
	default:
		return "other"
	}

// createExampleFromDetection creates an example from a detection
func (r *EnhancedReportingSystem) createExampleFromDetection(prompt string, detection *Detection) string {
	if detection.Location == nil {
		return prompt
	}
	
	// Extract the context from the detection
	if detection.Location.Context != "" {
		return detection.Location.Context
	}
	
	// Extract the relevant part of the prompt
	start := detection.Location.Start
	end := detection.Location.End
	
	// Ensure valid indices
	if start < 0 {
		start = 0
	}
	if end > len(prompt) {
		end = len(prompt)
	}
	
	// Extract context (50 chars before and after)
	contextStart := start - 50
	if contextStart < 0 {
		contextStart = 0
	}
	
	contextEnd := end + 50
	if contextEnd > len(prompt) {
		contextEnd = len(prompt)
	}
	
	// Extract the context
	context := prompt[contextStart:contextEnd]
	
	// Add ellipsis if truncated
	if contextStart > 0 {
		context = "..." + context
	}
	if contextEnd < len(prompt) {
		context = context + "..."
	}
	
	return context

// calculateSeverity calculates the severity of detections
func (r *EnhancedReportingSystem) calculateSeverity(detections []*Detection) float64 {
	if len(detections) == 0 {
		return 0.0
	}
	
	// Calculate base severity as the maximum confidence
	maxConfidence := 0.0
	for _, detection := range detections {
		if detection.Confidence > maxConfidence {
			maxConfidence = detection.Confidence
		}
	}
	
	// Adjust severity based on detection type
	severityMultiplier := 1.0
	
	// Check detection types
	for _, detection := range detections {
		switch detection.Type {
		case DetectionTypeJailbreak, DetectionTypeSystemPrompt:
			// These are high-severity detection types
			severityMultiplier = 1.5
		case DetectionTypePromptInjection, DetectionTypeRoleChange:
			// These are medium-high severity detection types
			if severityMultiplier < 1.3 {
				severityMultiplier = 1.3
			}
		}
	}
	
	// Calculate final severity
	severity := maxConfidence * severityMultiplier
	
	// Cap at 1.0
	if severity > 1.0 {
		severity = 1.0
	}
	
	return severity

// createDescriptionFromDetections creates a description from detections
func (r *EnhancedReportingSystem) createDescriptionFromDetections(detections []*Detection) string {
	if len(detections) == 0 {
		return "Unknown injection technique"
	}
	
	// Use the description of the detection with the highest confidence
	maxConfidence := 0.0
	var bestDescription string
	
	for _, detection := range detections {
		if detection.Confidence > maxConfidence {
			maxConfidence = detection.Confidence
			bestDescription = detection.Description
		}
	}
	
	return bestDescription
// saveReportToDisk saves a report to disk
func (r *EnhancedReportingSystem) saveReportToDisk(report *EnhancedInjectionReport) error {
	// Create reports directory if it doesn't exist
	reportsDir := filepath.Join(r.dataDir, "reports")
	if err := os.MkdirAll(reportsDir, 0700); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}
	
	// Create category directory if it doesn't exist
	categoryDir := filepath.Join(reportsDir, report.Category)
	if err := os.MkdirAll(categoryDir, 0700); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}
	
	// Create file path
	filePath := filepath.Join(categoryDir, fmt.Sprintf("%s.json", report.ReportID))
	
	// Marshal to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write report to file: %w", err)
	}
	
	return nil

// AnalyzeReports analyzes all reports
func (r *EnhancedReportingSystem) AnalyzeReports(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Skip if analysis is disabled
	if !r.reportingConfig.EnableReportAnalysis {
		return nil
	}
	
	// Find related reports
	r.findRelatedReports()
	
	// Calculate false positive rates
	r.calculateFalsePositiveRates()
	
	// Calculate effectiveness scores
	r.calculateEffectivenessScores()
	
	// Save reports to disk
	for _, report := range r.reports {
		if err := r.saveReportToDisk(report); err != nil {
			// Log error but continue processing
			fmt.Printf("Error saving report to disk: %v\n", err)
		}
	}
	
	return nil

// findRelatedReports finds related reports
func (r *EnhancedReportingSystem) findRelatedReports() {
	// Group reports by category
	reportsByCategory := make(map[string][]*EnhancedInjectionReport)
	for _, report := range r.reports {
		reportsByCategory[report.Category] = append(reportsByCategory[report.Category], report)
	}
	
	// Find related reports within each category
	for _, reports := range reportsByCategory {
		for i, report := range reports {
			// Clear existing related reports
			report.RelatedReports = make([]string, 0)
			
			// Find related reports
			for j, otherReport := range reports {
				if i == j {
					continue
				}
				
				// Check if patterns are similar
				if r.arePatternsRelated(report.Pattern, otherReport.Pattern) {
					report.RelatedReports = append(report.RelatedReports, otherReport.ReportID)
				}
			}
		}
	}

// arePatternsRelated checks if two patterns are related
func (r *EnhancedReportingSystem) arePatternsRelated(pattern1 string, pattern2 string) bool {
	// Simple check for now: if one pattern contains the other
	return pattern1 != "" && pattern2 != "" && (pattern1 == pattern2 || strings.Contains(pattern1, pattern2) || strings.Contains(pattern2, pattern1))

// calculateFalsePositiveRates calculates false positive rates
func (r *EnhancedReportingSystem) calculateFalsePositiveRates() {
	// In a real implementation, this would use feedback from users or other sources
	// For now, we'll use a simple heuristic based on the number of related reports
	for _, report := range r.reports {
		if len(report.RelatedReports) > 0 {
			// More related reports means lower false positive rate
			report.FalsePositiveRate = 1.0 / float64(len(report.RelatedReports)+1)
		} else {
			// No related reports means higher false positive rate
			report.FalsePositiveRate = 0.5
		}
		
		// Update analysis results
		report.AnalysisResults["false_positive_rate"] = report.FalsePositiveRate
		report.Status = ReportStatusAnalyzed
		report.LastUpdated = time.Now()
	}

// calculateEffectivenessScores calculates effectiveness scores
func (r *EnhancedReportingSystem) calculateEffectivenessScores() {
	for _, report := range r.reports {
		// Effectiveness is based on confidence, severity, and false positive rate
		report.EffectivenessScore = (report.Confidence + report.Severity) / 2 * (1 - report.FalsePositiveRate)
		
		// Update analysis results
		report.AnalysisResults["effectiveness_score"] = report.EffectivenessScore
		report.Status = ReportStatusAnalyzed
		report.LastUpdated = time.Now()
	}

// ShareReport shares a report
func (r *EnhancedReportingSystem) ShareReport(ctx context.Context, reportID string, destination string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Skip if sharing is disabled
	if !r.reportingConfig.EnableReportSharing {
		return fmt.Errorf("report sharing is disabled")
	}
	
	// Get report
	report, ok := r.reports[reportID]
	if !ok {
		return fmt.Errorf("report not found")
	}
	
	// Check if destination is valid
	endpoint, ok := r.reportingConfig.ReportingEndpoints[destination]
	if !ok {
		return fmt.Errorf("invalid destination")
	}
	
	// In a real implementation, this would send the report to the destination
	// For now, we'll just update the report status
	report.Shared = true
	report.SharedWith = append(report.SharedWith, destination)
	report.Status = ReportStatusShared
	report.LastUpdated = time.Now()
	
	// Save to disk
	if err := r.saveReportToDisk(report); err != nil {
		return fmt.Errorf("failed to save report to disk: %w", err)
	}
	
	// Log sharing
	fmt.Printf("Report %s shared with %s (%s)\n", reportID, destination, endpoint)
	
	return nil

// VerifyReport verifies a report
func (r *EnhancedReportingSystem) VerifyReport(ctx context.Context, reportID string, verified bool, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Get report
	report, ok := r.reports[reportID]
	if !ok {
		return fmt.Errorf("report not found")
	}
	
	// Update report status
	if verified {
		report.Status = ReportStatusVerified
		
		// Add to pattern library if verified
		if r.patternLibrary != nil {
			if err := r.patternLibrary.ValidateEmergingPattern(report.Pattern, true); err != nil {
				// Log error but continue processing
				fmt.Printf("Error validating pattern: %v\n", err)
			}
		}
	} else {
		report.Status = ReportStatusRejected
		
		// Remove from pattern library if rejected
		if r.patternLibrary != nil {
			if err := r.patternLibrary.ValidateEmergingPattern(report.Pattern, false); err != nil {
				// Log error but continue processing
				fmt.Printf("Error invalidating pattern: %v\n", err)
			}
		}
	}
	
	// Update metadata
	report.Metadata["verification_reason"] = reason
	report.LastUpdated = time.Now()
	
	// Save to disk
	return r.saveReportToDisk(report)

// GetReports gets all reports
func (r *EnhancedReportingSystem) GetReports() []*EnhancedInjectionReport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Convert map to slice
	reports := make([]*EnhancedInjectionReport, 0, len(r.reports))
	for _, report := range r.reports {
		reports = append(reports, report)
	}
	
	return reports

// GetReportsByCategory gets reports by category
func (r *EnhancedReportingSystem) GetReportsByCategory(category string) []*EnhancedInjectionReport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Filter reports by category
	reports := make([]*EnhancedInjectionReport, 0)
	for _, report := range r.reports {
		if report.Category == category {
			reports = append(reports, report)
		}
	}
	
	return reports

// GetReportsByStatus gets reports by status
func (r *EnhancedReportingSystem) GetReportsByStatus(status ReportStatus) []*EnhancedInjectionReport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Filter reports by status
	reports := make([]*EnhancedInjectionReport, 0)
	for _, report := range r.reports {
		if report.Status == status {
			reports = append(reports, report)
		}
	}
	
	return reports

// GetReport gets a report by ID
func (r *EnhancedReportingSystem) GetReport(reportID string) (*EnhancedInjectionReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Get report from memory
	report, ok := r.reports[reportID]
	if ok {
		return report, nil
	}
	
	// Try to load from disk
	for _, category := range r.reportingConfig.ReportCategories {
		filePath := filepath.Join(r.dataDir, "reports", category, fmt.Sprintf("%s.json", reportID))
		if _, err := os.Stat(filePath); err == nil {
			data, err := ioutil.ReadFile(filepath.Clean(filePath))
			if err != nil {
				return nil, fmt.Errorf("failed to read report file: %w", err)
			}
			
			var report EnhancedInjectionReport
			if err := json.Unmarshal(data, &report); err != nil {
				return nil, fmt.Errorf("failed to unmarshal report: %w", err)
			}
			
			return &report, nil
		}
	}
	
	return nil, fmt.Errorf("report not found")

// RegisterReportHandler registers a handler for reports
func (r *EnhancedReportingSystem) RegisterReportHandler(name string, handler ReportHandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.reportHandlers[name] = handler

// SetReportingThreshold sets the threshold for reporting
func (r *EnhancedReportingSystem) SetReportingThreshold(threshold float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.reportingConfig.ReportingThreshold = threshold

// EnableAutomaticReporting enables or disables automatic reporting
func (r *EnhancedReportingSystem) EnableAutomaticReporting(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.reportingConfig.EnableAutomaticReporting = enabled

// AddReportingEndpoint adds a reporting endpoint
func (r *EnhancedReportingSystem) AddReportingEndpoint(name string, endpoint string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.reportingConfig.ReportingEndpoints[name] = endpoint

// CreateCustomReport creates a custom report
func (r *EnhancedReportingSystem) CreateCustomReport(ctx context.Context, detectionType DetectionType, pattern string, example string, description string, createdBy string) (*EnhancedInjectionReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Create report ID
	reportID := fmt.Sprintf("report-%d", time.Now().UnixNano())
	
	// Determine category
	category := r.determineReportCategory(detectionType)
	
	// Create base injection report
	baseReport := &InjectionReport{
		ReportID:      reportID,
		DetectionType: detectionType,
		Pattern:       pattern,
		Example:       example,
		Confidence:    0.8, // Default confidence for custom reports
		Severity:      0.8, // Default severity for custom reports
		Description:   description,
		Timestamp:     time.Now(),
		Source:        "custom",
		Metadata:      make(map[string]interface{}),
	}
	
	// Create enhanced injection report
	report := &EnhancedInjectionReport{
		InjectionReport:    baseReport,
		Status:             ReportStatusNew,
		Category:           category,
		AnalysisResults:    make(map[string]interface{}),
		RelatedReports:     make([]string, 0),
		PatternMatches:     1,
		FalsePositiveRate:  0.0,
		EffectivenessScore: 0.8,
		CreatedBy:          createdBy,
		LastUpdated:        time.Now(),
		Shared:             false,
		SharedWith:         make([]string, 0),
	}
	
	// Add to reports
	r.reports[reportID] = report
	
	// Save to disk
	if err := r.saveReportToDisk(report); err != nil {
		return nil, fmt.Errorf("failed to save report to disk: %w", err)
	}
	
	// Process with report handlers
	for _, handler := range r.reportHandlers {
		if err := handler(ctx, report); err != nil {
			// Log error but continue processing
			fmt.Printf("Error processing report: %v\n", err)
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
}
}
}
