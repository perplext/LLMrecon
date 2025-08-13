// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// ReportingSystem manages reporting of new prompt injection techniques
type ReportingSystem struct {
	config          *ProtectionConfig
	reportingConfig *ReportingConfig
	reports         []*InjectionReport
	reportChan      chan *InjectionReport
	stopChan        chan struct{}
	mu              sync.RWMutex
	running         bool
}

// NewReportingSystem creates a new reporting system
func NewReportingSystem(config *ProtectionConfig) *ReportingSystem {
	// Create default reporting config if not specified
	reportingConfig := &ReportingConfig{
		ReportingInterval: time.Hour, // Report every hour by default
		MaxReportHistory:  1000,      // Keep up to 1000 reports in memory
		EnableLocalStorage: true,
		LocalStoragePath:   "reports",
	}

	system := &ReportingSystem{
		config:          config,
		reportingConfig: reportingConfig,
		reports:         make([]*InjectionReport, 0),
		reportChan:      make(chan *InjectionReport, 100),
		stopChan:        make(chan struct{}),
	}

	// Start the reporting loop
	system.Start()

	return system
}

// Start starts the reporting system
func (r *ReportingSystem) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return
	}

	r.running = true

	// Start the reporting loop
	go r.reportingLoop()
}

// Stop stops the reporting system
func (r *ReportingSystem) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return
	}

	r.running = false
	close(r.stopChan)
}

// Close closes the reporting system and releases resources
func (r *ReportingSystem) Close() error {
	r.Stop()
	return nil
}

// reportingLoop is the main reporting loop
func (r *ReportingSystem) reportingLoop() {
	ticker := time.NewTicker(r.reportingConfig.ReportingInterval)
	defer ticker.Stop()

	for {
		select {
		case report := <-r.reportChan:
			// Process the report
			r.processReport(report)
		case <-ticker.C:
			// Send reports to the reporting endpoint
			r.sendReports()
		case <-r.stopChan:
			return
		}
	}
}

// processReport processes a new report
func (r *ReportingSystem) processReport(report *InjectionReport) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add the report to the list
	r.reports = append(r.reports, report)

	// Prune old reports if we have too many
	if len(r.reports) > r.reportingConfig.MaxReportHistory {
		r.reports = r.reports[len(r.reports)-r.reportingConfig.MaxReportHistory:]
	}

	// Save to local storage if enabled
	if r.reportingConfig.EnableLocalStorage {
		r.saveReportToLocalStorage(report)
	}

	// Call the reporting callback if provided
	if r.config.ReportingCallback != nil {
		// Run in a goroutine to avoid blocking
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()

			err := r.config.ReportingCallback(ctx, report)
			if err != nil {
				// Log the error
				// This would be implemented with a proper logging system
				fmt.Printf("Error calling reporting callback: %v\n", err)
			}
		}()
	}
}

// saveReportToLocalStorage saves a report to local storage
func (r *ReportingSystem) saveReportToLocalStorage(report *InjectionReport) {
	// Ensure the directory exists
	if err := os.MkdirAll(r.reportingConfig.LocalStoragePath, 0755); err != nil {
		// Log the error
		// This would be implemented with a proper logging system
		fmt.Printf("Error creating directory for reports: %v\n", err)
		return
	}

	// Create a filename based on the report ID and timestamp
	filename := filepath.Join(
		r.reportingConfig.LocalStoragePath,
		fmt.Sprintf("report_%s_%d.json", report.ReportID, report.Timestamp.Unix()),
	)

	// Marshal the report to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		// Log the error
		// This would be implemented with a proper logging system
		fmt.Printf("Error marshaling report to JSON: %v\n", err)
		return
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		// Log the error
		// This would be implemented with a proper logging system
		fmt.Printf("Error writing report to file: %v\n", err)
		return
	}
}

// sendReports sends reports to the reporting endpoint
func (r *ReportingSystem) sendReports() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Skip if there are no reports or no endpoint
	if len(r.reports) == 0 || r.reportingConfig.ReportingEndpoint == "" {
		return
	}

	// In a real implementation, this would send the reports to an API endpoint
	// For now, we'll just log that reports would be sent
	// This would be implemented with a proper logging system
	fmt.Printf("Would send %d reports to %s\n", len(r.reports), r.reportingConfig.ReportingEndpoint)
}

// ReportDetections reports detections from a protection result
func (r *ReportingSystem) ReportDetections(ctx context.Context, result *ProtectionResult) {
	// Skip if there are no detections
	if len(result.Detections) == 0 {
		return
	}

	// Process each detection
	for _, detection := range result.Detections {
		// Skip low confidence detections
		if detection.Confidence < 0.7 {
			continue
		}

		// Create a report with safe access to potentially nil fields
		example := ""
		if detection.Location != nil {
			example = detection.Location.Context
		}
		
		report := &InjectionReport{
			ReportID:      uuid.New().String(),
			DetectionType: detection.Type,
			Pattern:       detection.Pattern,
			Example:       example,
			Confidence:    detection.Confidence,
			Severity:      calculateSeverity(detection, result),
			Description:   detection.Description,
			Timestamp:     time.Now(),
			Source:        "prompt_protection_system",
			Metadata: map[string]interface{}{
				"risk_score":    result.RiskScore,
				"action_taken":  result.ActionTaken,
				"detection_raw": detection,
			},
		}

		// Send the report to the channel
		select {
		case r.reportChan <- report:
			// Report sent successfully
		default:
			// Channel is full, log the error
			// This would be implemented with a proper logging system
			fmt.Println("Report channel is full, dropping report")
		}
	}
}

// GetReports gets all reports
func (r *ReportingSystem) GetReports() []*InjectionReport {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy of the reports
	reports := make([]*InjectionReport, len(r.reports))
	copy(reports, r.reports)

	return reports
}

// GetReportsByType gets reports by detection type
func (r *ReportingSystem) GetReportsByType(detectionType DetectionType) []*InjectionReport {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Filter reports by type
	reports := make([]*InjectionReport, 0)
	for _, report := range r.reports {
		if report.DetectionType == detectionType {
			reports = append(reports, report)
		}
	}

	return reports
}

// GetReportByID gets a report by ID
func (r *ReportingSystem) GetReportByID(reportID string) *InjectionReport {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find the report by ID
	for _, report := range r.reports {
		if report.ReportID == reportID {
			return report
		}
	}

	return nil
}

// calculateSeverity calculates the severity of a detection
func calculateSeverity(detection *Detection, result *ProtectionResult) float64 {
	// Base severity on confidence
	severity := detection.Confidence

	// Adjust based on detection type
	switch detection.Type {
	case DetectionTypeJailbreak:
		// Jailbreak attempts are very severe
		severity = max(severity, 0.9)
	case DetectionTypeSystemPrompt:
		// System prompt injections are very severe
		severity = max(severity, 0.9)
	case DetectionTypePromptInjection:
		// Direct prompt injections are severe
		severity = max(severity, 0.8)
	case DetectionTypeRoleChange:
		// Role changes can be severe depending on the role
		if detection.Metadata != nil {
			if isHighRisk, ok := detection.Metadata["is_high_risk"].(bool); ok && isHighRisk {
				severity = max(severity, 0.8)
			}
		}
	case DetectionTypeIndirectPromptInjection:
		// Indirect prompt injections are moderately severe
		severity = max(severity, 0.7)
	case DetectionTypeDelimiterMisuse:
		// Delimiter misuse can be moderately severe
		severity = max(severity, 0.6)
	case DetectionTypeBoundaryViolation:
		// Boundary violations are less severe
		severity = max(severity, 0.5)
	case DetectionTypeUnusualPattern:
		// Unusual patterns are less severe
		severity = max(severity, 0.5)
	case DetectionTypeProhibitedContent:
		// Prohibited content can vary in severity
		severity = max(severity, 0.6)
	}

	// Adjust based on risk score
	if result.RiskScore > 0.8 {
		// High risk score increases severity
		severity = max(severity, 0.7)
	}

	return severity
}
