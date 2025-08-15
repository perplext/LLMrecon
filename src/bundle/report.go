// Package bundle provides functionality for importing and exporting bundles
//
// The reporting system in this package is designed to provide a flexible and
// extensible way to generate reports about bundle import operations. The system
// consists of several components:
//
// 1. ReportLevel: Defines the level of detail in a report (Basic, Detailed, Verbose)
// 2. ImportResult: Contains the result of an import operation
// 3. ImportReport: Wraps an ImportResult with additional metadata
// 4. ReportManager: Manages the creation and storage of reports
// 5. ReportGenerator: Generates reports from import results
//
// The reporting system is designed to be used by the import system to generate
// reports about import operations, but it can also be used independently.
package bundle

import (
	"encoding/json"
	"fmt"
)

// ReportLevel represents the level of detail in a report
type ReportLevel string

const (
	// BasicReportLevel provides basic information about the import
	BasicReportLevel ReportLevel = "basic"
	// DetailedReportLevel provides detailed information about the import
	DetailedReportLevel ReportLevel = "detailed"
	// VerboseReportLevel provides verbose information about the import
	VerboseReportLevel ReportLevel = "verbose"
)

// ReportFormat represents the format of a report
type ReportFormat string

const (
	// JSONReportFormat represents a JSON report
	JSONReportFormat ReportFormat = "json"
	// TextReportFormat represents a plain text report
	TextReportFormat ReportFormat = "text"
	// MarkdownReportFormat represents a markdown report
	MarkdownReportFormat ReportFormat = "markdown"
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	// Success indicates whether the import was successful
	Success bool `json:"success"`
	// Message contains a human-readable message about the import
	Message string `json:"message"`
	// BundleID is the ID of the imported bundle
	BundleID string `json:"bundle_id"`
	// BundleName is the name of the imported bundle
	BundleName string `json:"bundle_name"`
	// BundleVersion is the version of the imported bundle
	BundleVersion string `json:"bundle_version"`
	// BundleType is the type of the imported bundle
	BundleType BundleType `json:"bundle_type"`
	// ImportTime is the time the import was performed
	ImportTime time.Time `json:"import_time,omitempty"`
	// StartTime is the time the import started
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the import ended
	EndTime time.Time `json:"end_time"`
	// Duration is the duration of the import
	Duration time.Duration `json:"duration"`
	// Errors contains any errors encountered during import
	Errors []string `json:"errors,omitempty"`
	// Warnings contains any warnings that occurred during import
	Warnings []string `json:"warnings,omitempty"`
	// ImportedItems contains the items that were imported
	ImportedItems []ContentItem `json:"imported_items,omitempty"`
	// ValidationResult contains the result of the validation
	ValidationResult *ValidationResult `json:"validation_result,omitempty"`
	// ValidationResults contains the results of validation
	ValidationResults []*ValidationResult `json:"validation_results,omitempty"`
	// ErrorReportPath is the path to the error report
	ErrorReportPath string `json:"error_report_path,omitempty"`
	// Conflicts contains the conflicts that were detected
	Conflicts []*Conflict `json:"conflicts,omitempty"`
	// ConflictResolutions contains the resolutions for conflicts
	ConflictResolutions []*ConflictResolution `json:"conflict_resolutions,omitempty"`
	// BackupID is the ID of the backup created during import
	BackupID string `json:"backup_id,omitempty"`
	// BackupPath is the path to the backup created during import
	BackupPath string `json:"backup_path,omitempty"`
	// ImportedFiles is a list of files that were imported
	ImportedFiles []string `json:"imported_files,omitempty"`
	// SkippedFiles is a list of files that were skipped
	SkippedFiles []string `json:"skipped_files,omitempty"`
	// ErrorMessage contains the error message if the import failed
	ErrorMessage string `json:"error_message,omitempty"`

// ImportReport represents a report of an import operation
type ImportReport struct {
	// ID is the unique identifier for the report
	ID string `json:"id"`
	// Result is the result of the import
	Result *ImportResult `json:"result"`
	// Level is the level of detail in the report
	Level ReportLevel `json:"level"`
	// Format is the format of the report
	Format ReportFormat `json:"format"`
	// GeneratedAt is the time the report was generated
	GeneratedAt time.Time `json:"generated_at"`

// ReportGenerator defines the interface for generating reports
type ReportGenerator interface {
	// GenerateReport generates a report of an import operation
	GenerateReport(result *ImportResult, level ReportLevel, format ReportFormat) (*ImportReport, error)
	// WriteReport writes a report to a writer
	WriteReport(report *ImportReport, writer io.Writer) error
	// SaveReport saves a report to a file
	SaveReport(report *ImportReport, path string) error

// DefaultReportGenerator is the default implementation of ReportGenerator
type DefaultReportGenerator struct {
	// Logger is the logger for report generation operations
	Logger io.Writer

// NewReportGenerator creates a new report generator
func NewReportGenerator(logger io.Writer) ReportGenerator {
	if logger == nil {
		logger = os.Stdout
	}
	return &DefaultReportGenerator{
		Logger: logger,
	}

// GenerateReport generates a report of an import operation
func (g *DefaultReportGenerator) GenerateReport(result *ImportResult, level ReportLevel, format ReportFormat) (*ImportReport, error) {
	// Create a unique ID for the report
	reportID := fmt.Sprintf("report-%s-%s", result.BundleID, time.Now().Format("20060102-150405"))

	// Create a new report
	report := &ImportReport{
		ID:          reportID,
		Result:      result,
		Level:       level,
		Format:      format,
		GeneratedAt: time.Now(),
	}

	// Filter the report based on the level
	if level == BasicReportLevel {
		// For basic level, remove detailed information
		report.Result.ValidationResults = nil
		report.Result.Conflicts = nil
		report.Result.ConflictResolutions = nil
		report.Result.ImportedFiles = nil
		report.Result.SkippedFiles = nil
		report.Result.Warnings = nil
	} else if level == DetailedReportLevel {
		// For detailed level, keep most information but limit lists
		if len(report.Result.ImportedFiles) > 10 {
			report.Result.ImportedFiles = report.Result.ImportedFiles[:10]
			report.Result.Warnings = append(report.Result.Warnings, "Imported files list truncated")
		}
		if len(report.Result.SkippedFiles) > 10 {
			report.Result.SkippedFiles = report.Result.SkippedFiles[:10]
			report.Result.Warnings = append(report.Result.Warnings, "Skipped files list truncated")
		}
	}
	// For verbose level, keep all information

	return report, nil

// WriteReport writes a report to a writer
func (g *DefaultReportGenerator) WriteReport(report *ImportReport, writer io.Writer) error {
	switch report.Format {
	case JSONReportFormat:
		return g.writeJSONReport(report, writer)
	case TextReportFormat:
		return g.writeTextReport(report, writer)
	case MarkdownReportFormat:
		return g.writeMarkdownReport(report, writer)
	default:
		return fmt.Errorf("unsupported report format: %s", report.Format)
	}

// SaveReport saves a report to a file
func (g *DefaultReportGenerator) SaveReport(report *ImportReport, path string) error {
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
	return g.WriteReport(report, file)

// writeJSONReport writes a report in JSON format
func (g *DefaultReportGenerator) writeJSONReport(report *ImportReport, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)

// writeTextReport writes a report in plain text format
func (g *DefaultReportGenerator) writeTextReport(report *ImportReport, writer io.Writer) error {
	result := report.Result

	// Write the header
	fmt.Fprintf(writer, "Bundle Import Report\n")
	fmt.Fprintf(writer, "====================\n\n")

	// Write the basic information
	fmt.Fprintf(writer, "Import Status: %s\n", getStatusString(result.Success))
	fmt.Fprintf(writer, "Bundle ID: %s\n", result.BundleID)
	fmt.Fprintf(writer, "Bundle Name: %s\n", result.BundleName)
	fmt.Fprintf(writer, "Bundle Version: %s\n", result.BundleVersion)
	fmt.Fprintf(writer, "Bundle Type: %s\n", result.BundleType)
	fmt.Fprintf(writer, "Import Time: %s\n", result.ImportTime.Format(time.RFC3339))
	fmt.Fprintf(writer, "Duration: %s\n", result.Duration)

	// If the import failed, write the error message
	if !result.Success {
		fmt.Fprintf(writer, "\nError: %s\n", result.ErrorMessage)
	}

	// Write the backup information
	if result.BackupID != "" {
		fmt.Fprintf(writer, "\nBackup ID: %s\n", result.BackupID)
		fmt.Fprintf(writer, "Backup Path: %s\n", result.BackupPath)
	}

	// Write detailed information based on the report level
	if report.Level != BasicReportLevel {
		// Write validation results
		if len(result.ValidationResults) > 0 {
			fmt.Fprintf(writer, "\nValidation Results:\n")
			for i, vr := range result.ValidationResults {
				fmt.Fprintf(writer, "  %d. Level: %s, Valid: %t\n", i+1, vr.Level, vr.IsValid)
				if !vr.IsValid && vr.Error != nil {
					fmt.Fprintf(writer, "     Error: %s\n", vr.Error.Error())
				}
			}
		}

		// Write conflict information
		if len(result.Conflicts) > 0 {
			fmt.Fprintf(writer, "\nConflicts:\n")
			for i, conflict := range result.Conflicts {
				fmt.Fprintf(writer, "  %d. Type: %s, Path: %s\n", i+1, conflict.Type, conflict.Path)
				fmt.Fprintf(writer, "     Message: %s\n", conflict.Message)
				if conflict.Resolved {
					fmt.Fprintf(writer, "     Resolved: Yes, Strategy: %s\n", conflict.ResolutionStrategy)
					if conflict.ResolutionPath != "" {
						fmt.Fprintf(writer, "     Resolution Path: %s\n", conflict.ResolutionPath)
					}
				} else {
					fmt.Fprintf(writer, "     Resolved: No\n")
				}
			}
		}

		// Write warnings
		if len(result.Warnings) > 0 {
			fmt.Fprintf(writer, "\nWarnings:\n")
			for i, warning := range result.Warnings {
				fmt.Fprintf(writer, "  %d. %s\n", i+1, warning)
			}
		}

		// Write imported files
		if len(result.ImportedFiles) > 0 {
			fmt.Fprintf(writer, "\nImported Files:\n")
			for i, file := range result.ImportedFiles {
				fmt.Fprintf(writer, "  %d. %s\n", i+1, file)
			}
			if len(result.ImportedFiles) == 10 && report.Level == DetailedReportLevel {
				fmt.Fprintf(writer, "  (list truncated, use verbose level for full list)\n")
			}
		}

		// Write skipped files
		if len(result.SkippedFiles) > 0 {
			fmt.Fprintf(writer, "\nSkipped Files:\n")
			for i, file := range result.SkippedFiles {
				fmt.Fprintf(writer, "  %d. %s\n", i+1, file)
			}
			if len(result.SkippedFiles) == 10 && report.Level == DetailedReportLevel {
				fmt.Fprintf(writer, "  (list truncated, use verbose level for full list)\n")
			}
		}
	}

	// Write the footer
	fmt.Fprintf(writer, "\nReport generated at: %s\n", report.GeneratedAt.Format(time.RFC3339))

	return nil

// writeMarkdownReport writes a report in markdown format
func (g *DefaultReportGenerator) writeMarkdownReport(report *ImportReport, writer io.Writer) error {
	result := report.Result

	// Write the header
	fmt.Fprintf(writer, "# Bundle Import Report\n\n")

	// Write the basic information
	fmt.Fprintf(writer, "## Summary\n\n")
	fmt.Fprintf(writer, "| Property | Value |\n")
	fmt.Fprintf(writer, "| --- | --- |\n")
	fmt.Fprintf(writer, "| Import Status | %s |\n", getStatusString(result.Success))
	fmt.Fprintf(writer, "| Bundle ID | %s |\n", result.BundleID)
	fmt.Fprintf(writer, "| Bundle Name | %s |\n", result.BundleName)
	fmt.Fprintf(writer, "| Bundle Version | %s |\n", result.BundleVersion)
	fmt.Fprintf(writer, "| Bundle Type | %s |\n", result.BundleType)
	fmt.Fprintf(writer, "| Import Time | %s |\n", result.ImportTime.Format(time.RFC3339))
	fmt.Fprintf(writer, "| Duration | %s |\n", result.Duration)

	// If the import failed, write the error message
	if !result.Success {
		fmt.Fprintf(writer, "\n## Error\n\n")
		fmt.Fprintf(writer, "%s\n", result.ErrorMessage)
	}

	// Write the backup information
	if result.BackupID != "" {
		fmt.Fprintf(writer, "\n## Backup\n\n")
		fmt.Fprintf(writer, "| Property | Value |\n")
		fmt.Fprintf(writer, "| --- | --- |\n")
		fmt.Fprintf(writer, "| Backup ID | %s |\n", result.BackupID)
		fmt.Fprintf(writer, "| Backup Path | %s |\n", result.BackupPath)
	}

	// Write detailed information based on the report level
	if report.Level != BasicReportLevel {
		// Write validation results
		if len(result.ValidationResults) > 0 {
			fmt.Fprintf(writer, "\n## Validation Results\n\n")
			fmt.Fprintf(writer, "| Level | Valid | Error |\n")
			fmt.Fprintf(writer, "| --- | --- | --- |\n")
			for _, vr := range result.ValidationResults {
				errorMsg := ""
				if !vr.IsValid && vr.Error != nil {
					errorMsg = vr.Error.Error()
				}
				fmt.Fprintf(writer, "| %s | %t | %s |\n", vr.Level, vr.IsValid, errorMsg)
			}
		}

		// Write conflict information
		if len(result.Conflicts) > 0 {
			fmt.Fprintf(writer, "\n## Conflicts\n\n")
			fmt.Fprintf(writer, "| Type | Path | Message | Resolved | Strategy | Resolution Path |\n")
			fmt.Fprintf(writer, "| --- | --- | --- | --- | --- | --- |\n")
			for _, conflict := range result.Conflicts {
				resolved := "No"
				strategy := ""
				resolutionPath := ""
				if conflict.Resolved {
					resolved = "Yes"
					strategy = string(conflict.ResolutionStrategy)
					resolutionPath = conflict.ResolutionPath
				}
				fmt.Fprintf(writer, "| %s | %s | %s | %s | %s | %s |\n",
					conflict.Type, conflict.Path, conflict.Message, resolved, strategy, resolutionPath)
			}
		}

		// Write warnings
		if len(result.Warnings) > 0 {
			fmt.Fprintf(writer, "\n## Warnings\n\n")
			for _, warning := range result.Warnings {
				fmt.Fprintf(writer, "- %s\n", warning)
			}
		}

		// Write imported files
		if len(result.ImportedFiles) > 0 {
			fmt.Fprintf(writer, "\n## Imported Files\n\n")
			for _, file := range result.ImportedFiles {
				fmt.Fprintf(writer, "- %s\n", file)
			}
			if len(result.ImportedFiles) == 10 && report.Level == DetailedReportLevel {
				fmt.Fprintf(writer, "\n*List truncated, use verbose level for full list*\n")
			}
		}

		// Write skipped files
		if len(result.SkippedFiles) > 0 {
			fmt.Fprintf(writer, "\n## Skipped Files\n\n")
			for _, file := range result.SkippedFiles {
				fmt.Fprintf(writer, "- %s\n", file)
			}
			if len(result.SkippedFiles) == 10 && report.Level == DetailedReportLevel {
				fmt.Fprintf(writer, "\n*List truncated, use verbose level for full list*\n")
			}
		}
	}

	// Write the footer
	fmt.Fprintf(writer, "\n---\n\n")
	fmt.Fprintf(writer, "*Report generated at: %s*\n", report.GeneratedAt.Format(time.RFC3339))

	return nil

// getStatusString returns a string representation of a success status
func getStatusString(success bool) string {
	if success {
		return "Success"
	}
	return "Failed"

// ReportManager defines the interface for managing reports
type ReportManager interface {
	// CreateImportReport creates a report for an import operation
	CreateImportReport(result *ImportResult, level ReportLevel, format ReportFormat) (*ImportReport, error)
	// SaveImportReport saves an import report to a file
	SaveImportReport(report *ImportReport, path string) error
	// GetReportPath gets the path for a report
	GetReportPath(bundleID string, format ReportFormat) string
	// ListReports lists all reports in a directory
	ListReports(dir string) ([]string, error)

// DefaultReportManager is the default implementation of ReportManager
type DefaultReportManager struct {
	// Generator is the report generator
	Generator ReportGenerator
	// ReportsDir is the directory where reports are stored
	ReportsDir string
	// Logger is the logger for report management operations
	Logger io.Writer

// NewReportManager creates a new report manager
func NewReportManager(reportsDir string, logger io.Writer) ReportManager {
	if logger == nil {
		logger = os.Stdout
	}
	return &DefaultReportManager{
		Generator:  NewReportGenerator(logger),
		ReportsDir: reportsDir,
		Logger:     logger,
	}

// CreateImportReport creates a report for an import operation
func (m *DefaultReportManager) CreateImportReport(result *ImportResult, level ReportLevel, format ReportFormat) (*ImportReport, error) {
	return m.Generator.GenerateReport(result, level, format)

// SaveImportReport saves an import report to a file
func (m *DefaultReportManager) SaveImportReport(report *ImportReport, path string) error {
	return m.Generator.SaveReport(report, path)

// GetReportPath gets the path for a report
func (m *DefaultReportManager) GetReportPath(bundleID string, format ReportFormat) string {
	// Create a filename based on the bundle ID and format
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("import-%s-%s.%s", bundleID, timestamp, format)
	return filepath.Join(m.ReportsDir, filename)

// ListReports lists all reports in a directory
func (m *DefaultReportManager) ListReports(dir string) ([]string, error) {
	// If no directory is specified, use the reports directory
	if dir == "" {
		dir = m.ReportsDir
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// List all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Filter for report files
	var reports []string
	for _, file := range files {
		if !file.IsDir() && isReportFile(file.Name()) {
			reports = append(reports, filepath.Join(dir, file.Name()))
		}
	}

	return reports, nil

// isReportFile checks if a filename is a report file
func isReportFile(filename string) bool {
	// Check if the filename starts with "import-" and has a valid extension
	ext := filepath.Ext(filename)
	return len(filename) > 7 && filename[:7] == "import-" && (ext == ".json" || ext == ".txt" || ext == ".md")
