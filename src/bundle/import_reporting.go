// Package bundle provides functionality for importing and exporting bundles
//
// The enhanced import reporting system extends the basic reporting system to provide
// more detailed information about import operations. It includes:
//
// 1. ImportStatistics: Detailed statistics about the import operation
// 2. SystemImpact: Assessment of the impact of the import on the system
// 3. EnhancedImportReport: Combines the basic ImportReport with additional details
//
// This system is designed to be used by the import system to generate comprehensive
// reports about import operations, with varying levels of detail based on the
// specified ReportLevel (from report.go).
package bundle

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

// Using ReportLevel from report.go for consistency

// ImportStatistics contains statistics about an import operation
type ImportStatistics struct {
	// TotalFiles is the total number of files in the bundle
	TotalFiles int `json:"total_files"`
	// ImportedFiles is the number of files that were imported
	ImportedFiles int `json:"imported_files"`
	// SkippedFiles is the number of files that were skipped
	SkippedFiles int `json:"skipped_files"`
	// ConflictFiles is the number of files that had conflicts
	ConflictFiles int `json:"conflict_files"`
	// ResolvedConflicts is the number of conflicts that were resolved
	ResolvedConflicts int `json:"resolved_conflicts"`
	// ValidationErrors is the number of validation errors
	ValidationErrors int `json:"validation_errors"`
	// ValidationWarnings is the number of validation warnings
	ValidationWarnings int `json:"validation_warnings"`
	// TotalSize is the total size of the imported files in bytes
	TotalSize int64 `json:"total_size"`
	// ProcessingTime is the time spent processing the import
	ProcessingTime time.Duration `json:"processing_time"`
	// ValidationTime is the time spent validating the bundle
	ValidationTime time.Duration `json:"validation_time"`
	// ExtractionTime is the time spent extracting the bundle
	ExtractionTime time.Duration `json:"extraction_time"`
	// InstallationTime is the time spent installing the bundle
	InstallationTime time.Duration `json:"installation_time"`

// SystemImpactAssessment contains information about the impact of an import on the system
type SystemImpactAssessment struct {
	// DiskSpaceUsed is the amount of disk space used by the import in bytes
	DiskSpaceUsed int64 `json:"disk_space_used"`
	// DiskSpaceAvailable is the amount of disk space available after the import in bytes
	DiskSpaceAvailable int64 `json:"disk_space_available"`
	// MemoryUsage is the amount of memory used during the import in bytes
	MemoryUsage int64 `json:"memory_usage"`
	// CPUUsage is the CPU usage during the import as a percentage
	CPUUsage float64 `json:"cpu_usage"`
	// BackupSize is the size of the backup in bytes
	BackupSize int64 `json:"backup_size"`
	// ConfigChanges is the number of configuration changes made
	ConfigChanges int `json:"config_changes"`
	// SecurityImpact contains information about the security impact of the import
	SecurityImpact string `json:"security_impact"`

// EnhancedImportReport extends ImportReport with additional detailed information
type EnhancedImportReport struct {
	// Base report information
	*ImportReport `json:",inline"`
	// Statistics contains statistics about the import operation
	Statistics *ImportStatistics `json:"statistics,omitempty"`
	// SystemImpact contains information about the impact of the import on the system
	SystemImpact *SystemImpactAssessment `json:"system_impact,omitempty"`
	// ValidationDetails contains detailed information about validation results
	ValidationDetails map[string]interface{} `json:"validation_details,omitempty"`
	// ConflictDetails contains detailed information about conflicts
	ConflictDetails map[string]interface{} `json:"conflict_details,omitempty"`
	// PerformanceMetrics contains detailed performance metrics
	PerformanceMetrics map[string]interface{} `json:"performance_metrics,omitempty"`
	// LogEntries contains log entries related to the import
	LogEntries []string `json:"log_entries,omitempty"`
	// ReportingLevel is the level of detail in the report
	ReportingLevel ReportLevel `json:"reporting_level"` // Using ReportLevel from report.go for consistency

// ImportReportingSystem defines the interface for the import reporting system
type ImportReportingSystem interface {
	// CreateReport creates a report for an import operation
	CreateReport(result *ImportResult, level ReportLevel) (*EnhancedImportReport, error)
	// SaveReport saves a report to a file
	SaveReport(report *EnhancedImportReport, path string) error
	// GenerateStatistics generates statistics for an import operation
	GenerateStatistics(result *ImportResult) *ImportStatistics
	// AssessSystemImpact assesses the impact of an import on the system
	AssessSystemImpact(result *ImportResult) *SystemImpactAssessment
	// LogImportEvent logs an event during the import process
	LogImportEvent(bundleID, event string, args ...interface{})
	// GetLogEntries gets the log entries for an import
	GetLogEntries(bundleID string) []string
	// AddPerformanceMetric adds a performance metric for an import
	AddPerformanceMetric(bundleID string, metricName string, value interface{})
	// GenerateUserFriendlySummary creates a human-readable summary of an import operation
	GenerateUserFriendlySummary(report *EnhancedImportReport) string

// DefaultImportReportingSystem is the default implementation of ImportReportingSystem
type DefaultImportReportingSystem struct {
	// ReportManager is the report manager
	ReportManager ReportManager
	// Logger is the logger for import operations
	Logger io.Writer
	// LogEntries stores log entries by bundle ID
	LogEntries map[string][]string
	// PerformanceMetrics stores performance metrics by bundle ID
	PerformanceMetrics map[string]map[string]interface{}
	// ReportsDir is the directory where reports are stored
	ReportsDir string

// NewImportReportingSystem creates a new import reporting system
func NewImportReportingSystem(reportManager ReportManager, reportsDir string, logger io.Writer) ImportReportingSystem {
	if logger == nil {
		logger = os.Stdout
	}
	if reportManager == nil {
		reportManager = NewReportManager(reportsDir, logger)
	}
	return &DefaultImportReportingSystem{
		ReportManager:      reportManager,
		Logger:             logger,
		LogEntries:         make(map[string][]string),
		PerformanceMetrics: make(map[string]map[string]interface{}),
		ReportsDir:         reportsDir,
	}

// CreateReport creates an enhanced report for an import operation.
// This method uses ReportLevel from report.go for consistency across the reporting system.
// The level parameter determines the amount of detail included in the report:
// - BasicReportLevel: Only includes the base import report
// - DetailedReportLevel: Adds statistics, validation details, and conflict details
// - VerboseReportLevel: Adds system impact assessment, performance metrics, and log entries
func (r *DefaultImportReportingSystem) CreateReport(result *ImportResult, level ReportLevel) (*EnhancedImportReport, error) {
	// Create base report
	baseReport, err := r.ReportManager.CreateImportReport(result, level, JSONReportFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to create base report: %w", err)
	}

	// Create enhanced report
	enhancedReport := &EnhancedImportReport{
		ImportReport:   baseReport,
		ReportingLevel: level,
	}

	// Add additional information based on reporting level
	if level != BasicReportLevel {
		enhancedReport.Statistics = r.GenerateStatistics(result)
		
		if level == DetailedReportLevel || level == VerboseReportLevel {
			enhancedReport.ValidationDetails = r.extractValidationDetails(result)
			enhancedReport.ConflictDetails = r.extractConflictDetails(result)
		}
		
		if level == VerboseReportLevel {
			enhancedReport.SystemImpact = r.AssessSystemImpact(result)
			enhancedReport.PerformanceMetrics = r.collectPerformanceMetrics(result)
			enhancedReport.LogEntries = r.GetLogEntries(result.BundleID)
		}
	}

	return enhancedReport, nil

// SaveReport saves a report to a file
func (r *DefaultImportReportingSystem) SaveReport(report *EnhancedImportReport, path string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Write the report to the file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report: %w", err)
	}

	return nil

// GenerateStatistics generates statistics for an import operation
func (r *DefaultImportReportingSystem) GenerateStatistics(result *ImportResult) *ImportStatistics {
	stats := &ImportStatistics{
		TotalFiles:         len(result.ImportedFiles) + len(result.SkippedFiles),
		ImportedFiles:      len(result.ImportedFiles),
		SkippedFiles:       len(result.SkippedFiles),
		ConflictFiles:      len(result.Conflicts),
		ResolvedConflicts:  len(result.ConflictResolutions),
		ValidationErrors:   0,
		ValidationWarnings: 0,
		ProcessingTime:     result.Duration,
	}

	// Calculate validation errors and warnings
	for _, validationResult := range result.ValidationResults {
		stats.ValidationErrors += len(validationResult.Errors)
		stats.ValidationWarnings += len(validationResult.Warnings)
	}

	// Calculate total size
	for _, file := range result.ImportedFiles {
		info, err := os.Stat(file)
		if err == nil {
			stats.TotalSize += info.Size()
		}
	}

	// Calculate phase times if available
	if result.StartTime.Unix() > 0 && result.EndTime.Unix() > 0 {
		// This is an approximation - in a real system we would track these times precisely
		totalTime := result.EndTime.Sub(result.StartTime)
		stats.ValidationTime = totalTime / 4
		stats.ExtractionTime = totalTime / 4
		stats.InstallationTime = totalTime / 2
	}

	return stats

// AssessSystemImpact assesses the impact of an import on the system
func (r *DefaultImportReportingSystem) AssessSystemImpact(result *ImportResult) *SystemImpactAssessment {
	impact := &SystemImpactAssessment{
		SecurityImpact: "Low", // Default value
	}

	// Calculate disk space used
	for _, file := range result.ImportedFiles {
		info, err := os.Stat(file)
		if err == nil {
			impact.DiskSpaceUsed += info.Size()
		}
	}

	// Calculate backup size
	if result.BackupPath != "" {
		size, err := r.calculateDirectorySize(result.BackupPath)
		if err == nil {
			impact.BackupSize = size
		}
	}

	// Get disk space available
	impact.DiskSpaceAvailable = r.getDiskSpaceForImportedFiles(result.ImportedFiles)

	// Get memory usage (approximate)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	impact.MemoryUsage = int64(m.Alloc)

	// Estimate CPU usage (this would be more accurate in a real system)
	impact.CPUUsage = 50.0 // Placeholder value

	// Count config changes (this would be more accurate in a real system)
	impact.ConfigChanges = 0
	for _, file := range result.ImportedFiles {
		if strings.Contains(file, "config") || strings.HasSuffix(file, ".json") || strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
			impact.ConfigChanges++
		}
	}

	// Assess security impact
	if impact.ConfigChanges > 5 {
		impact.SecurityImpact = "Medium"
	}
	if impact.ConfigChanges > 10 {
		impact.SecurityImpact = "High"
	}

	return impact

// LogImportEvent logs an event during the import process
func (r *DefaultImportReportingSystem) LogImportEvent(bundleID, event string, args ...interface{}) {
	message := fmt.Sprintf(event, args...)
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)

	// Log to the logger
	fmt.Fprintln(r.Logger, logEntry)

	// Store in memory
	if r.LogEntries[bundleID] == nil {
		r.LogEntries[bundleID] = []string{}
	}
	r.LogEntries[bundleID] = append(r.LogEntries[bundleID], logEntry)

// GetLogEntries gets the log entries for an import
func (r *DefaultImportReportingSystem) GetLogEntries(bundleID string) []string {
	return r.LogEntries[bundleID]

// AddPerformanceMetric adds a performance metric for an import
func (r *DefaultImportReportingSystem) AddPerformanceMetric(bundleID string, metricName string, value interface{}) {
	// Initialize the metrics map for this bundle if it doesn't exist
	if _, exists := r.PerformanceMetrics[bundleID]; !exists {
		r.PerformanceMetrics[bundleID] = make(map[string]interface{})
	}
	
	// Add the metric
	r.PerformanceMetrics[bundleID][metricName] = value
	
	// Log the metric addition
	r.LogImportEvent(bundleID, "Performance metric added: %s = %v", metricName, value)

// extractValidationDetails extracts detailed validation information about the import operation
// This provides a comprehensive view of all validation checks performed during import
func (r *DefaultImportReportingSystem) extractValidationDetails(result *ImportResult) map[string]interface{} {
	details := make(map[string]interface{})

	// Extract primary validation result
	if result.ValidationResult != nil {
		details["overall"] = map[string]interface{}{
			"valid":    result.ValidationResult.Valid,
			"message":  result.ValidationResult.Message,
			"errors":   result.ValidationResult.Errors,
			"warnings": result.ValidationResult.Warnings,
		}

		// Add error and warning details if available
		var errorDetails []map[string]interface{}
		var warningDetails []map[string]interface{}

		for _, err := range result.ValidationResult.Errors {
			errorDetails = append(errorDetails, map[string]interface{}{
				"message": err,
				"severity": "error",
			})
		}

		for _, warn := range result.ValidationResult.Warnings {
			warningDetails = append(warningDetails, map[string]interface{}{
				"message": warn,
				"severity": "warning",
			})
		}

		details["errors"] = errorDetails
		details["warnings"] = warningDetails
	}

	// Group validation results by level for more detailed analysis
	levelResults := make(map[string][]*ValidationResult)
	componentResults := make(map[string][]*ValidationResult)
	validationSummary := make(map[string]int)

	for _, vr := range result.ValidationResults {
		// Group by validation level
		levelResults[string(vr.Level)] = append(levelResults[string(vr.Level)], vr)

		// Group by component if available in the Details map
		if vr.Details != nil {
			if component, ok := vr.Details["component"].(string); ok && component != "" {
				componentResults[component] = append(componentResults[component], vr)
			}
		}

		// Count validation results by status
		if vr.Valid {
			validationSummary["passed"]++
		} else {
			validationSummary["failed"]++
		}

		// Count by level
		validationSummary["level_"+string(vr.Level)]++
	}

	details["by_level"] = levelResults
	details["by_component"] = componentResults
	details["summary"] = validationSummary

	// Generate a human-readable summary text
	var summaryText strings.Builder
	summaryText.WriteString(fmt.Sprintf("Validation checks: %d passed, %d failed\n", 
		validationSummary["passed"], validationSummary["failed"]))
	
	if result.ValidationResult != nil && len(result.ValidationResult.Errors) > 0 {
		summaryText.WriteString(fmt.Sprintf("Critical errors: %d\n", len(result.ValidationResult.Errors)))
	}
	
	if result.ValidationResult != nil && len(result.ValidationResult.Warnings) > 0 {
		summaryText.WriteString(fmt.Sprintf("Warnings: %d\n", len(result.ValidationResult.Warnings)))
	}

	details["summary_text"] = summaryText.String()

	return details

// extractConflictDetails extracts detailed conflict information
func (r *DefaultImportReportingSystem) extractConflictDetails(result *ImportResult) map[string]interface{} {
	details := make(map[string]interface{})

	// Group conflicts by type
	typeConflicts := make(map[string][]string)
	for _, conflict := range result.Conflicts {
		conflictType := "file"
		if strings.Contains(conflict.Path, "config") {
			conflictType = "config"
		} else if strings.Contains(conflict.Path, "data") {
			conflictType = "data"
		}
		typeConflicts[conflictType] = append(typeConflicts[conflictType], conflict.Path)
	}

	details["by_type"] = typeConflicts

	// Resolution statistics
	resolutionStats := make(map[string]int)
	for _, resolution := range result.ConflictResolutions {
		resolutionStats[string(resolution.Strategy)]++
	}

	details["resolutions"] = resolutionStats

	return details

// GenerateUserFriendlySummary creates a human-readable summary of an import operation
// This provides a clear, concise overview of the import results suitable for end users
func (r *DefaultImportReportingSystem) GenerateUserFriendlySummary(report *EnhancedImportReport) string {
	var summary strings.Builder
	
	// Basic import information
	summary.WriteString(fmt.Sprintf("Import Summary for Bundle: %s\n", report.Result.BundleID))
	summary.WriteString(fmt.Sprintf("Generated: %s\n\n", report.GeneratedAt.Format(time.RFC1123)))
	
	// Overall status
	if report.Result.Success {
		summary.WriteString("Status: ✅ Import Successful\n")
	} else {
		summary.WriteString("Status: ❌ Import Failed\n")
		if report.Result.Message != "" {
			summary.WriteString(fmt.Sprintf("Error: %s\n", report.Result.Message))
		}
	}
	summary.WriteString("\n")
	
	// Statistics if available
	if report.Statistics != nil {
		summary.WriteString("Import Statistics:\n")
		summary.WriteString(fmt.Sprintf("- Files Processed: %d of %d\n", 
			report.Statistics.ImportedFiles, report.Statistics.TotalFiles))
		
		if report.Statistics.SkippedFiles > 0 {
			summary.WriteString(fmt.Sprintf("- Files Skipped: %d\n", report.Statistics.SkippedFiles))
		}
		
		if report.Statistics.ConflictFiles > 0 {
			summary.WriteString(fmt.Sprintf("- Files with Conflicts: %d\n", report.Statistics.ConflictFiles))
			summary.WriteString(fmt.Sprintf("- Conflicts Resolved: %d\n", report.Statistics.ResolvedConflicts))
		}
		
		if report.Statistics.ValidationErrors > 0 || report.Statistics.ValidationWarnings > 0 {
			summary.WriteString(fmt.Sprintf("- Validation Issues: %d errors, %d warnings\n", 
				report.Statistics.ValidationErrors, report.Statistics.ValidationWarnings))
		}
		
		summary.WriteString(fmt.Sprintf("- Total Size: %s\n", formatByteSize(report.Statistics.TotalSize)))
		summary.WriteString(fmt.Sprintf("- Total Processing Time: %s\n", 
			formatImportDuration(report.Statistics.ProcessingTime)))
		summary.WriteString("\n")
	}
	
	// System impact if available
	if report.SystemImpact != nil {
		summary.WriteString("System Impact:\n")
		summary.WriteString(fmt.Sprintf("- Disk Space Used: %s\n", 
			formatByteSize(report.SystemImpact.DiskSpaceUsed)))
		summary.WriteString(fmt.Sprintf("- Disk Space Available: %s\n", 
			formatByteSize(report.SystemImpact.DiskSpaceAvailable)))
		
		if report.SystemImpact.ConfigChanges > 0 {
			summary.WriteString(fmt.Sprintf("- Configuration Changes: %d\n", 
				report.SystemImpact.ConfigChanges))
		}
		
		if report.SystemImpact.SecurityImpact != "" {
			summary.WriteString(fmt.Sprintf("- Security Impact: %s\n", 
				report.SystemImpact.SecurityImpact))
		}
		summary.WriteString("\n")
	}
	
	// Add validation summary if detailed level
	if report.ValidationDetails != nil && len(report.ValidationDetails) > 0 {
		summary.WriteString("Validation Summary:\n")
		if details, ok := report.ValidationDetails["summary"].(string); ok {
			summary.WriteString(details)
		} else {
			// Fallback if summary not available
			if errors, ok := report.ValidationDetails["errors"].([]interface{}); ok {
				summary.WriteString(fmt.Sprintf("- Errors: %d\n", len(errors)))
			}
			if warnings, ok := report.ValidationDetails["warnings"].([]interface{}); ok {
				summary.WriteString(fmt.Sprintf("- Warnings: %d\n", len(warnings)))
			}
		}
		summary.WriteString("\n")
	}
	
	// Add conflict summary if detailed level
	if report.ConflictDetails != nil && len(report.ConflictDetails) > 0 {
		summary.WriteString("Conflict Summary:\n")
		if conflicts, ok := report.ConflictDetails["conflicts"].([]interface{}); ok {
			summary.WriteString(fmt.Sprintf("- Total Conflicts: %d\n", len(conflicts)))
		}
		if resolutions, ok := report.ConflictDetails["resolutions"].(map[string]interface{}); ok {
			var autoResolved, manualResolved, unresolved int
			for _, v := range resolutions {
				if resolution, ok := v.(string); ok {
					switch resolution {
					case "auto":
						autoResolved++
					case "manual":
						manualResolved++
					case "unresolved":
						unresolved++
					}
				}
			}
			summary.WriteString(fmt.Sprintf("- Auto-resolved: %d\n", autoResolved))
			summary.WriteString(fmt.Sprintf("- Manually resolved: %d\n", manualResolved))
			if unresolved > 0 {
				summary.WriteString(fmt.Sprintf("- Unresolved: %d\n", unresolved))
			}
		}
		summary.WriteString("\n")
	}
	
	// Add next steps or recommendations
	summary.WriteString("Next Steps:\n")
	if !report.Result.Success {
		summary.WriteString("- Review error details and try again after addressing issues\n")
		if report.Statistics != nil && report.Statistics.ValidationErrors > 0 {
			summary.WriteString("- Fix validation errors in the bundle\n")
		}
	} else {
		summary.WriteString("- Import completed successfully\n")
		summary.WriteString("- Verify system functionality with the new components\n")
		if report.SystemImpact != nil && report.SystemImpact.ConfigChanges > 0 {
			summary.WriteString("- Review configuration changes\n")
		}
	}
	
	return summary.String()

// Helper function to format byte sizes in a human-readable way
func formatByteSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])

// formatImportDuration formats durations for import reporting in a human-readable way
func formatImportDuration(d time.Duration) string {
	// For durations less than a minute, show milliseconds
	if d < time.Minute {
		return fmt.Sprintf("%.2f seconds", d.Seconds())
	}
	
	// For longer durations, format as minutes and seconds
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	
	// For very long durations, include hours
	hours := minutes / 60
	minutes = minutes % 60
	
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)

// collectPerformanceMetrics collects detailed performance metrics
func (r *DefaultImportReportingSystem) collectPerformanceMetrics(result *ImportResult) map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// Calculate total processing time
	processingTime := result.EndTime.Sub(result.StartTime)
	metrics["total_processing_time_ms"] = processingTime.Milliseconds()
	
	// Add memory usage information
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics["alloc_bytes"] = memStats.Alloc
	metrics["sys_bytes"] = memStats.Sys
	metrics["heap_alloc_bytes"] = memStats.HeapAlloc
	metrics["heap_sys_bytes"] = memStats.HeapSys
	metrics["num_gc"] = memStats.NumGC
	
	// Add CPU usage information if available
	// This is a simplified approach and might not be accurate in all environments
	metrics["num_goroutines"] = runtime.NumGoroutine()
	metrics["num_cpu"] = runtime.NumCPU()
	
	// Add any custom performance metrics that were collected during the import process
	if customMetrics, exists := r.PerformanceMetrics[result.BundleID]; exists {
		for key, value := range customMetrics {
			metrics[key] = value
		}
	}

	return metrics

// calculateDirectorySize calculates the size of a directory in bytes
func (r *DefaultImportReportingSystem) calculateDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err

// ImportConflict represents a conflict between a bundle file and an existing file during import reporting
type ImportConflict struct {
	// Path is the path to the conflicting file
	Path string `json:"path"`
	// BundleChecksum is the checksum of the file in the bundle
	BundleChecksum string `json:"bundle_checksum,omitempty"`
	// ExistingChecksum is the checksum of the existing file
	ExistingChecksum string `json:"existing_checksum,omitempty"`
	// Type is the type of conflict
	Type string `json:"type,omitempty"`

// ConflictResolution represents a resolution for a conflict
// ValidationResult is defined in types.go
