// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle/errors"
)

// ProgressCallback is a function that is called to report progress
type ProgressCallback func(progress float64, message string)

// ImportOptions represents options for importing a bundle
type ImportOptions struct {
	// ValidationLevel is the level of validation to perform
	ValidationLevel ValidationLevel
	// TargetDir is the directory to import the bundle to
	TargetDir string
	// BackupDir is the directory to store backups in
	BackupDir string
	// Force indicates whether to force import even if there are conflicts
	Force bool
	// Logger is the logger for import operations
	Logger io.Writer
	// VerboseReporting indicates whether to generate verbose reports
	VerboseReporting bool
	// ProgressCallback is called with progress updates
	ProgressCallback ProgressCallback
	// User is the user performing the import operation
	User string
	// AuditLogger is the logger for audit events
	AuditLogger *errors.AuditLogger
}

// ImportResult is defined in report.go

// BundleImporter defines the interface for bundle import
type BundleImporter interface {
	// Import imports a bundle with the specified options
	Import(context.Context, string, ImportOptions) (*ImportResult, error)
	// ValidateBeforeImport validates a bundle before importing
	ValidateBeforeImport(context.Context, string, ValidationLevel) (*ValidationResult, error)
	// CreateBackup creates a backup of the target directory
	CreateBackup(context.Context, string, string) (string, error)
	// RestoreBackup restores a backup
	RestoreBackup(context.Context, string, string) error
}

// DefaultBundleImporter is the default implementation of BundleImporter
type DefaultBundleImporter struct {
	// Validator is the validator for bundle validation
	Validator BundleValidator
	// ReportManager is the report manager for generating reports
	ReportManager ReportManager
	// ReportingSystem is the reporting system for generating reports
	ReportingSystem ImportReportingSystem
	// Logger is the logger for import operations
	Logger io.Writer
	// AuditLogger is the logger for audit events
	AuditLogger *errors.AuditLogger
	// ErrorHandler is the error handler for import operations
	ErrorHandler errors.ErrorHandler
	// EnhancedErrorHandler is the enhanced error handler for advanced error handling
	EnhancedErrorHandler *errors.EnhancedErrorHandler
	// RecoveryManager is the recovery manager for error recovery
	RecoveryManager *errors.RecoveryManager
	// ErrorReporter is the error reporter for error reporting
	ErrorReporter errors.ErrorReporter
	// CollectedErrors tracks errors that occurred during import
	CollectedErrors []*errors.BundleError
}

// NewBundleImporter creates a new bundle importer
func NewBundleImporter(validator BundleValidator, reportManager ReportManager, logger io.Writer) BundleImporter {
	if validator == nil {
		validator = NewBundleValidator(os.Stdout)
	}
	if logger == nil {
		logger = os.Stdout
	}
	
	// Create a reporting system if not provided through the report manager
	var reportingSystem ImportReportingSystem
	if reportManager != nil {
		// If we have a report manager, we can determine the reports directory
		reportsDir := ""
		if rm, ok := reportManager.(*DefaultReportManager); ok {
			reportsDir = rm.ReportsDir
		}
		reportingSystem = NewImportReportingSystem(reportManager, reportsDir, logger)
	}
	
	// Create default audit logger with system user
	auditLogFile, err := os.OpenFile(
		fmt.Sprintf("%s/bundle_audit_%s.log", 
			os.TempDir(), 
			time.Now().Format("20060102")),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 
		0644,
	)
	if err != nil {
		// Fall back to the main logger if audit log file creation fails
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] Failed to create audit log file: %v\n", 
			time.Now().Format(time.RFC3339), err)
		auditLogFile = os.Stderr
	}
	
	// Create audit logger
	auditLogger := errors.NewAuditLogger(auditLogFile, "system")
	
	// Create enhanced error handler
	enhancedErrorHandler := errors.NewEnhancedErrorHandler(auditLogger)
	
	// Create recovery manager
	recoveryManager := errors.NewRecoveryManager(logger, auditLogger)
	
	// Add recovery strategies
	recoveryManager.AddStrategy(errors.NewFileSystemRecoveryStrategy(logger))
	recoveryManager.AddStrategy(errors.NewBackupRecoveryStrategy(logger, os.TempDir()+"/backups"))
	recoveryManager.AddStrategy(errors.NewNetworkRecoveryStrategy(logger, 3))
	recoveryManager.AddStrategy(errors.NewConflictRecoveryStrategy(logger, false))
	
	// Create error reporter
	errorReporter := errors.NewErrorReporter(logger, os.TempDir()+"/error-reports")
	
	return &DefaultBundleImporter{
		Validator:           validator,
		ReportManager:       reportManager,
		ReportingSystem:     reportingSystem,
		Logger:              logger,
		AuditLogger:         auditLogger,
		ErrorHandler:        enhancedErrorHandler,
		EnhancedErrorHandler: enhancedErrorHandler,
		RecoveryManager:     recoveryManager,
		ErrorReporter:       errorReporter,
		CollectedErrors:     []*errors.BundleError{},
	}
}

// Import imports a bundle with the specified options
func (i *DefaultBundleImporter) Import(ctx context.Context, bundlePath string, options ImportOptions) (*ImportResult, error) {
	startTime := time.Now()
	result := &ImportResult{
		Success:       false,
		StartTime:     startTime,
		ImportTime:    startTime,
		Errors:        []string{},
		Warnings:      []string{},
		ImportedFiles: []string{},
		SkippedFiles:  []string{},
	}

	// Set default logger if not provided
	logger := options.Logger
	if logger == nil {
		logger = os.Stdout
	}

	// Create a reporting system if not already initialized
	if i.ReportingSystem == nil && i.ReportManager != nil {
		reportsDir := ""
		if rm, ok := i.ReportManager.(*DefaultReportManager); ok {
			reportsDir = rm.ReportsDir
		}
		i.ReportingSystem = NewImportReportingSystem(i.ReportManager, reportsDir, logger)
	}

	// Use audit logger from options if provided, otherwise use the default one
	auditLogger := i.AuditLogger
	if options.AuditLogger != nil {
		auditLogger = options.AuditLogger
	}
	
	// Set user if provided in options
	if options.User != "" && auditLogger != nil {
		auditLogger.User = options.User
	}

	// Generate a bundle ID for tracking
	bundleID := filepath.Base(bundlePath)
	
	// Log import start
	fmt.Fprintf(logger, "Starting bundle import: %s\n", bundlePath)
	
	// Use reporting system for logging if available
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Starting import of bundle: %s", bundlePath)
	}
	
	// Log audit event for import start
	if auditLogger != nil {
		auditLogger.LogImportStart(bundleID, bundlePath, map[string]interface{}{
			"validation_level": string(options.ValidationLevel),
			"force": options.Force,
			"user": options.User,
		})
	}

	// Validate bundle
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Validating bundle at validation level: %s", options.ValidationLevel)
	}
	
	// Log audit event for validation start
	if auditLogger != nil {
		auditLogger.LogEvent("validation_started", "BundleValidator", bundleID, map[string]interface{}{
			"validation_level": string(options.ValidationLevel),
			"bundle_path": bundlePath,
			"operation": "validation",
		})
	}
	
	// Perform validation with enhanced error handling
	var validationResult *ValidationResult
	err := errors.WithRetryAndContext(ctx, i.EnhancedErrorHandler, func() error {
		var validateErr error
		validationResult, validateErr = i.ValidateBeforeImport(ctx, bundlePath, options.ValidationLevel)
		if validateErr != nil {
			// Create a structured error with context
			bundleErr := errors.NewBundleError(
				validateErr,
				"Bundle validation failed",
				errors.ValidationError,
				errors.HighSeverity,
				errors.NonRecoverableError,
			)
			// Add context to the error
			bundleErr.WithContext("bundle_id", bundleID)
			bundleErr.WithContext("validation_level", string(options.ValidationLevel))
			bundleErr.WithContext("bundle_path", bundlePath)
			
			// Collect the error for reporting
			i.CollectedErrors = append(i.CollectedErrors, bundleErr)
			
			return bundleErr
		}
		return nil
	})
	
	// Handle validation error
	if err != nil {
		result.Message = fmt.Sprintf("Bundle validation failed: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = time.Now()
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Validation error: %s", err.Error())
		}
		
		// Log audit event for validation failure
		if auditLogger != nil {
			auditLogger.LogEventWithStatus("validation_failed", "BundleValidator", bundleID, "failure", map[string]interface{}{
				"error": err.Error(),
				"validation_level": string(options.ValidationLevel),
				"operation": "validation",
			})
			
			// Log audit event for import failure
			auditLogger.LogImportComplete(bundleID, false, map[string]interface{}{
				"errors": len(result.Errors),
				"warnings": len(result.Warnings),
				"duration_seconds": time.Since(startTime).Seconds(),
			})
		}
		
		// Generate error report
		if i.ErrorReporter != nil {
			reportPath := filepath.Join(os.TempDir(), "error-reports", fmt.Sprintf("import_failure_%s.json", bundleID))
			if reportErr := i.ErrorReporter.GenerateErrorReport(ctx, i.CollectedErrors, reportPath); reportErr != nil {
				fmt.Fprintf(i.Logger, "Failed to generate error report: %s\n", reportErr.Error())
			} else {
				result.ErrorReportPath = reportPath
				fmt.Fprintf(i.Logger, "Error report generated at: %s\n", reportPath)
			}
		}
		
		// Attempt recovery if possible
		if i.RecoveryManager != nil {
			if bundleErr, ok := err.(*errors.BundleError); ok {
				recovered, recoveryErr := i.RecoveryManager.AttemptRecovery(ctx, bundleErr)
				if recovered {
					fmt.Fprintf(i.Logger, "Successfully recovered from validation error, but import will not continue\n")
					result.Warnings = append(result.Warnings, "Recovered from validation error, but import cannot continue")
				} else if recoveryErr != nil {
					fmt.Fprintf(i.Logger, "Recovery attempt failed: %s\n", recoveryErr.Error())
				}
			}
		}
		
		return result, err
	}
	result.ValidationResult = validationResult
	
	// Add validation result to the list of validation results
	result.ValidationResults = append(result.ValidationResults, validationResult)
	
	// Log audit event for validation result
	if auditLogger != nil {
		auditLogger.LogValidation(bundleID, bundlePath, string(options.ValidationLevel), validationResult.IsValid, map[string]interface{}{
			"errors_count": len(validationResult.Errors),
			"warnings_count": len(validationResult.Warnings),
		})
	}

	// If validation failed, return error
	if !validationResult.Valid {
		result.Message = "Bundle validation failed"
		result.Errors = append(result.Errors, validationResult.Errors...)
		result.Warnings = append(result.Warnings, validationResult.Warnings...)
		result.EndTime = time.Now()
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Validation failed with %d errors and %d warnings", 
				len(validationResult.Errors), len(validationResult.Warnings))
		}
		
		return result, fmt.Errorf("bundle validation failed")
	}
	
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Validation successful with %d warnings", len(validationResult.Warnings))
	}

	// Open the bundle
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Opening bundle for processing")
	}
	
	bundle, err := OpenBundle(bundlePath)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to open bundle: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = time.Now()
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Failed to open bundle: %s", err.Error())
		}
		
		return result, err
	}
	
	// Store bundle information in the result
	result.BundleID = bundle.Manifest.BundleID
	result.BundleName = bundle.Manifest.Name
	result.BundleVersion = bundle.Manifest.Version
	result.BundleType = bundle.Manifest.BundleType
	
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Bundle opened successfully: %s (v%s)", 
			bundle.Manifest.Name, bundle.Manifest.Version)
	}

	// Create backup if requested
	var backupPath string
	if options.BackupDir != "" {
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Creating backup of target directory")
		}
		
		// Log audit event for backup start
		if auditLogger != nil {
			auditLogger.LogEvent("backup_started", "BundleImporter", bundleID, map[string]interface{}{
				"target_dir": options.TargetDir,
				"backup_dir": options.BackupDir,
				"operation": "backup",
			})
		}
		
		backupStartTime := time.Now()
		var backupErr error
		
		// Use WithRetryAndContext for backup creation with enhanced error handling
		backupErr = errors.WithRetryAndContext(ctx, i.EnhancedErrorHandler, func() error {
			var err error
			backupPath, err = i.CreateBackup(ctx, options.TargetDir, options.BackupDir)
			if err != nil {
				// Create a structured error with context
				bundleErr := errors.NewBundleError(
					err,
					"Failed to create backup",
					errors.BackupError,
					errors.HighSeverity,
					errors.RecoverableError,
				)
				// Add context to the error
				bundleErr.WithContext("bundle_id", bundleID)
				bundleErr.WithContext("target_dir", options.TargetDir)
				bundleErr.WithContext("backup_dir", options.BackupDir)
				
				// Collect the error for reporting
				i.CollectedErrors = append(i.CollectedErrors, bundleErr)
				
				return bundleErr
			}
			return nil
		})
		
		backupDuration := time.Since(backupStartTime)
		
		// Log backup duration as a performance metric
		if i.ReportingSystem != nil {
			i.ReportingSystem.AddPerformanceMetric(bundleID, "backup_creation_time", backupDuration.Seconds())
		}
		
		if backupErr != nil {
			result.Message = fmt.Sprintf("Failed to create backup: %s", backupErr.Error())
			result.Errors = append(result.Errors, backupErr.Error())
			result.EndTime = time.Now()
			
			if i.ReportingSystem != nil {
				i.ReportingSystem.LogImportEvent(bundleID, "Backup creation failed: %s", backupErr.Error())
			}
			
			// Log audit event for backup failure
			if auditLogger != nil {
				auditLogger.LogEventWithStatus("backup_failed", "BundleImporter", bundleID, "failure", map[string]interface{}{
					"error": backupErr.Error(),
					"target_dir": options.TargetDir,
					"backup_dir": options.BackupDir,
					"operation": "backup",
					"duration_seconds": backupDuration.Seconds(),
				})
			}
			
			// Generate error report
			if i.ErrorReporter != nil {
				reportPath := filepath.Join(os.TempDir(), "error-reports", fmt.Sprintf("backup_failure_%s.json", bundleID))
				if reportErr := i.ErrorReporter.GenerateErrorReport(ctx, i.CollectedErrors, reportPath); reportErr != nil {
					fmt.Fprintf(i.Logger, "Failed to generate error report: %s\n", reportErr.Error())
				} else {
					result.ErrorReportPath = reportPath
					fmt.Fprintf(i.Logger, "Error report generated at: %s\n", reportPath)
				}
			}
			
			// Attempt recovery if possible
			if i.RecoveryManager != nil {
				if bundleErr, ok := backupErr.(*errors.BundleError); ok {
					recovered, recoveryErr := i.RecoveryManager.AttemptRecovery(ctx, bundleErr)
					if recovered {
						fmt.Fprintf(i.Logger, "Successfully recovered from backup error, but import will not continue\n")
						result.Warnings = append(result.Warnings, "Recovered from backup error, but import cannot continue")
					} else if recoveryErr != nil {
						fmt.Fprintf(i.Logger, "Recovery attempt failed: %s\n", recoveryErr.Error())
					}
				}
			}
			
			return result, backupErr
		}
		
		result.BackupPath = backupPath
		fmt.Fprintf(logger, "Backup created at: %s\n", backupPath)
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Backup created at: %s", backupPath)
		}
		
		// Log audit event for backup success
		if auditLogger != nil {
			auditLogger.LogBackupCreated(bundleID, options.TargetDir, backupPath)
		}
	} else if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Skipping backup creation (no backup directory specified)")
	}
	
	fmt.Fprintf(logger, "Created backup at: %s\n", result.BackupPath)

	// Create temporary directory for extraction
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Creating temporary directory for extraction")
	}
	
	tempDir, err := os.MkdirTemp("", "bundle-import-*")
	if err != nil {
		result.Message = fmt.Sprintf("Failed to create temporary directory: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = time.Now()
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Failed to create temporary directory: %s", err.Error())
		}
		
		return result, err
	}
	defer os.RemoveAll(tempDir)
	
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Created temporary directory at: %s", tempDir)
	}

	// Extract bundle to temporary directory
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Extracting bundle to temporary directory")
	}
	
	extractionStartTime := time.Now()
	err = errors.WithRetryAndContext(ctx, i.EnhancedErrorHandler, func() error {
		extractErr := ExtractBundle(bundlePath, tempDir)
		if extractErr != nil {
			// Create a structured error with context
			bundleErr := errors.NewBundleError(
				extractErr,
				"Bundle extraction failed",
				errors.FileSystemError,
				errors.HighSeverity,
				errors.RecoverableError,
			)
			// Add context to the error
			bundleErr.WithContext("bundle_id", bundleID)
			bundleErr.WithContext("bundle_path", bundlePath)
			bundleErr.WithContext("temp_dir", tempDir)
			
			// Collect the error for reporting
			i.CollectedErrors = append(i.CollectedErrors, bundleErr)
			
			return bundleErr
		}
		return nil
	})
	extractionDuration := time.Since(extractionStartTime)
	
	if err != nil {
		result.Message = fmt.Sprintf("Failed to extract bundle: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = time.Now()
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Failed to extract bundle: %s", err.Error())
		}
		
		// Log audit event for extraction failure
		if auditLogger != nil {
			auditLogger.LogEventWithStatus("extraction_failed", "BundleImporter", bundleID, "failure", map[string]interface{}{
				"error": err.Error(),
				"bundle_path": bundlePath,
				"temp_dir": tempDir,
				"duration_seconds": extractionDuration.Seconds(),
			})
		}
		
		// Generate error report
		if i.ErrorReporter != nil {
			reportPath := filepath.Join(os.TempDir(), "error-reports", fmt.Sprintf("extraction_failure_%s.json", bundleID))
			if reportErr := i.ErrorReporter.GenerateErrorReport(ctx, i.CollectedErrors, reportPath); reportErr != nil {
				fmt.Fprintf(i.Logger, "Failed to generate error report: %s\n", reportErr.Error())
			} else {
				result.ErrorReportPath = reportPath
				fmt.Fprintf(i.Logger, "Error report generated at: %s\n", reportPath)
			}
		}
		
		// Attempt recovery if possible
		if i.RecoveryManager != nil {
			if bundleErr, ok := err.(*errors.BundleError); ok {
				recovered, recoveryErr := i.RecoveryManager.AttemptRecovery(ctx, bundleErr)
				if recovered {
					fmt.Fprintf(i.Logger, "Successfully recovered from extraction error, but import will not continue\n")
					result.Warnings = append(result.Warnings, "Recovered from extraction error, but import cannot continue")
				} else if recoveryErr != nil {
					fmt.Fprintf(i.Logger, "Recovery attempt failed: %s\n", recoveryErr.Error())
				}
			}
		}
		
		return result, err
	}
	
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Bundle extracted successfully (took %v)", extractionDuration)
		i.ReportingSystem.AddPerformanceMetric(bundleID, "extraction_time", extractionDuration.Seconds())
	}

	// Check for conflicts
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Checking for conflicts with target directory: %s", options.TargetDir)
	}
	
	// Log audit event for conflict check start
	if auditLogger != nil {
		auditLogger.LogEvent("conflict_check_started", "BundleImporter", bundleID, map[string]interface{}{
			"target_dir": options.TargetDir,
			"temp_dir": tempDir,
			"operation": "conflict_check",
		})
	}
	
	conflictCheckStartTime := time.Now()
	var conflicts []string
	err = errors.WithRetryAndContext(ctx, i.EnhancedErrorHandler, func() error {
		var conflictErr error
		conflicts, conflictErr = i.checkForConflicts(ctx, tempDir, options.TargetDir, bundle.Manifest.Content)
		if conflictErr != nil {
			// Create a structured error with context
			bundleErr := errors.NewBundleError(
				conflictErr,
				"Conflict check failed",
				errors.ConflictError,
				errors.MediumSeverity,
				errors.RecoverableError,
			)
			// Add context to the error
			bundleErr.WithContext("bundle_id", bundleID)
			bundleErr.WithContext("temp_dir", tempDir)
			bundleErr.WithContext("target_dir", options.TargetDir)
			
			// Collect the error for reporting
			i.CollectedErrors = append(i.CollectedErrors, bundleErr)
			
			return bundleErr
		}
		return nil
	})
	conflictCheckDuration := time.Since(conflictCheckStartTime)
	
	if err != nil {
		result.Message = fmt.Sprintf("Failed to check for conflicts: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = time.Now()
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Failed to check for conflicts: %s", err.Error())
		}
		
		// Log audit event for conflict check failure
		if auditLogger != nil {
			auditLogger.LogEventWithStatus("conflict_check_failed", "BundleImporter", bundleID, "failure", map[string]interface{}{
				"error": err.Error(),
				"target_dir": options.TargetDir,
				"operation": "conflict_check",
				"duration_seconds": conflictCheckDuration.Seconds(),
			})
		}
		
		// Generate error report
		if i.ErrorReporter != nil {
			reportPath := filepath.Join(os.TempDir(), "error-reports", fmt.Sprintf("conflict_check_failure_%s.json", bundleID))
			if reportErr := i.ErrorReporter.GenerateErrorReport(ctx, i.CollectedErrors, reportPath); reportErr != nil {
				fmt.Fprintf(i.Logger, "Failed to generate error report: %s\n", reportErr.Error())
			} else {
				result.ErrorReportPath = reportPath
				fmt.Fprintf(i.Logger, "Error report generated at: %s\n", reportPath)
			}
		}
		
		// Attempt recovery if possible
		if i.RecoveryManager != nil {
			if bundleErr, ok := err.(*errors.BundleError); ok {
				recovered, recoveryErr := i.RecoveryManager.AttemptRecovery(ctx, bundleErr)
				if recovered {
					fmt.Fprintf(i.Logger, "Successfully recovered from conflict check error, but import will not continue\n")
					result.Warnings = append(result.Warnings, "Recovered from conflict check error, but import cannot continue")
				} else if recoveryErr != nil {
					fmt.Fprintf(i.Logger, "Recovery attempt failed: %s\n", recoveryErr.Error())
				}
			}
		}
		
		// Log audit event for import failure
		if auditLogger != nil {
			auditLogger.LogImportComplete(bundleID, result.Success, map[string]interface{}{
				"imported_items": len(result.ImportedItems),
				"skipped_items": len(result.SkippedFiles),
				"errors": len(result.Errors),
				"warnings": len(result.Warnings),
				"duration_seconds": result.Duration.Seconds(),
				"error_report_path": result.ErrorReportPath,
			})
		}
		
		return result, err
	}
	
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Conflict check completed in %v, found %d conflicts", 
			conflictCheckDuration, len(conflicts))
		i.ReportingSystem.AddPerformanceMetric(bundleID, "conflict_check_time", conflictCheckDuration.Seconds())
	}
	
	// Log audit event for conflict check completion
	if auditLogger != nil {
		auditLogger.LogEvent("conflict_check_completed", "BundleImporter", bundleID, map[string]interface{}{
			"conflicts_found": len(conflicts),
			"duration_seconds": conflictCheckDuration.Seconds(),
			"target_dir": options.TargetDir,
			"operation": "conflict_check",
		})
	}

	// If there are conflicts and force is not enabled, return error
	if len(conflicts) > 0 && !options.Force {
		result.Message = fmt.Sprintf("Import would overwrite %d existing files", len(conflicts))
		for _, conflict := range conflicts {
			result.Errors = append(result.Errors, fmt.Sprintf("Conflict: %s", conflict))
		}
		result.EndTime = time.Now()
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Import aborted due to %d conflicts and force option not enabled", len(conflicts))
		}
		
		// Log audit event for import abort due to conflicts
		if auditLogger != nil {
			auditLogger.LogEventWithStatus("import_aborted", "BundleImporter", bundleID, "failure", map[string]interface{}{
				"reason": "conflicts_detected",
				"conflicts_count": len(conflicts),
				"force_enabled": false,
				"operation": "import",
			})
			
			// Log audit event for import failure
			auditLogger.LogImportComplete(bundleID, false, map[string]interface{}{
				"errors": len(result.Errors),
				"warnings": len(result.Warnings),
				"conflicts_count": len(conflicts),
				"duration_seconds": time.Since(startTime).Seconds(),
				"reason": "conflicts_detected",
			})
		}
		
		return result, fmt.Errorf("import would overwrite existing files, use force option to override")
	} else if len(conflicts) > 0 {
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Proceeding with import despite %d conflicts (force option enabled)", len(conflicts))
		}
		
		// Log audit event for proceeding with conflicts
		if auditLogger != nil {
			auditLogger.LogEvent("conflicts_overridden", "BundleImporter", bundleID, map[string]interface{}{
				"conflicts_count": len(conflicts),
				"force_enabled": true,
				"operation": "import",
			})
		}
	}

	// Import files
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Starting installation of %d files to target directory", len(bundle.Manifest.Content))
	}
	
	installStartTime := time.Now()
	successfulInstalls := 0
	skippedFiles := 0
	
	for _, item := range bundle.Manifest.Content {
		srcPath := filepath.Join(tempDir, item.Path)
		dstPath := filepath.Join(options.TargetDir, item.Path)
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Installing item: %s (type: %s)", item.Path, item.Type)
		}

		// Create parent directories
		err = os.MkdirAll(filepath.Dir(dstPath), 0755)
		if err != nil {
			result.Message = fmt.Sprintf("Failed to create directory for %s: %s", item.Path, err.Error())
			result.Errors = append(result.Errors, err.Error())
			
			if i.ReportingSystem != nil {
				i.ReportingSystem.LogImportEvent(bundleID, "Failed to create directory for %s: %s", item.Path, err.Error())
			}
			
			// Log audit event for directory creation failure
			if auditLogger != nil {
				// Set the bundleID in the content item for tracking
				item.BundleID = bundleID
				// Create detailed error information
				details := map[string]interface{}{
					"operation":    "directory_creation",
					"source_path":  item.Path,
					"target_path":  dstPath,
					"content_type": string(item.Type),
					"error":        err.Error(),
				}
				auditLogger.LogEventWithStatus("directory_creation_failed", "BundleImporter", bundleID, "failed", details)
			}
			
			// Attempt to restore backup if one was created
			if result.BackupPath != "" {
				if i.ReportingSystem != nil {
					i.ReportingSystem.LogImportEvent(bundleID, "Attempting to restore backup from: %s", result.BackupPath)
				}
				
				restoreErr := i.RestoreBackup(ctx, result.BackupPath, options.TargetDir)
				if restoreErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore backup: %s", restoreErr.Error()))
					
					if i.ReportingSystem != nil {
						i.ReportingSystem.LogImportEvent(bundleID, "Failed to restore backup: %s", restoreErr.Error())
					}
				} else {
					result.Warnings = append(result.Warnings, "Restored backup after failed import")
					
					if i.ReportingSystem != nil {
						i.ReportingSystem.LogImportEvent(bundleID, "Successfully restored backup after failed import")
					}
				}
			}
			result.EndTime = time.Now()
			return result, err
		}

		// Copy file or directory
		fileCopyStartTime := time.Now()
		err = copyPath(srcPath, dstPath)
		fileCopyDuration := time.Since(fileCopyStartTime)
		
		if err != nil {
			result.Message = fmt.Sprintf("Failed to copy %s: %s", item.Path, err.Error())
			result.Errors = append(result.Errors, err.Error())
			
			if i.ReportingSystem != nil {
				i.ReportingSystem.LogImportEvent(bundleID, "Failed to copy %s: %s", item.Path, err.Error())
			}
			
			// Log audit event for failed file installation
			if auditLogger != nil {
				// Set the bundleID in the content item for tracking
				item.BundleID = bundleID
				auditLogger.LogFileInstallation(bundleID, dstPath, false, map[string]interface{}{
					"error": err.Error(),
					"item_path": item.Path,
				})
			}
			
			// Attempt to restore backup if one was created
			if result.BackupPath != "" {
				if i.ReportingSystem != nil {
					i.ReportingSystem.LogImportEvent(bundleID, "Attempting to restore backup from: %s", result.BackupPath)
				}
				
				restoreErr := i.RestoreBackup(ctx, result.BackupPath, options.TargetDir)
				if restoreErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore backup: %s", restoreErr.Error()))
					
					if i.ReportingSystem != nil {
						i.ReportingSystem.LogImportEvent(bundleID, "Failed to restore backup: %s", restoreErr.Error())
					}
				} else {
					result.Warnings = append(result.Warnings, "Restored backup after failed import")
					
					if i.ReportingSystem != nil {
						i.ReportingSystem.LogImportEvent(bundleID, "Successfully restored backup after failed import")
					}
				}
			}
			result.EndTime = time.Now()
			return result, err
		}

		// Add to imported items
		result.ImportedItems = append(result.ImportedItems, item)
		successfulInstalls++
		
		if i.ReportingSystem != nil {
			i.ReportingSystem.LogImportEvent(bundleID, "Successfully installed: %s (took %v)", item.Path, fileCopyDuration)
		}
		
		// Log audit event for file installation using the specialized method
		if auditLogger != nil {
			// Set the bundleID in the content item for tracking
			item.BundleID = bundleID
			auditLogger.LogFileInstallation(bundleID, dstPath, true, map[string]interface{}{
				"item_path": item.Path,
				"item_type": string(item.Type),
			})
		}
		
		fmt.Fprintf(logger, "Imported: %s\n", item.Path)
	}
	
	installDuration := time.Since(installStartTime)
	
	if i.ReportingSystem != nil {
		i.ReportingSystem.LogImportEvent(bundleID, "Installation completed: %d files installed successfully, %d skipped (took %v)", 
			successfulInstalls, skippedFiles, installDuration)
		i.ReportingSystem.AddPerformanceMetric(bundleID, "installation_time", installDuration.Seconds())
		i.ReportingSystem.AddPerformanceMetric(bundleID, "files_installed", float64(successfulInstalls))
	}

	// Set success
	result.Success = true
	result.Message = fmt.Sprintf("Successfully imported %d items", len(result.ImportedItems))
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = len(result.Errors) == 0
	
	// Generate final error report if there were any errors or warnings
	if len(i.CollectedErrors) > 0 || len(result.Errors) > 0 || len(result.Warnings) > 0 {
		// Ensure we have collected all errors
		for _, errMsg := range result.Errors {
			// Check if this error is already in CollectedErrors
			var found bool
			for _, bundleErr := range i.CollectedErrors {
				if strings.Contains(bundleErr.Error(), errMsg) {
					found = true
					break
				}
			}
			
			// If not found, add it as a generic error
			if !found {
				bundleErr := errors.NewBundleError(
					fmt.Errorf(errMsg),
					errMsg,
					errors.UnknownError,
					errors.MediumSeverity,
					errors.NonRecoverableError,
				)
				bundleErr.WithContext("bundle_id", bundleID)
				i.CollectedErrors = append(i.CollectedErrors, bundleErr)
			}
		}
		
		// Generate comprehensive error report
		if i.ErrorReporter != nil {
			reportPath := filepath.Join(os.TempDir(), "error-reports", fmt.Sprintf("import_summary_%s.json", bundleID))
			if reportErr := i.ErrorReporter.GenerateErrorReport(ctx, i.CollectedErrors, reportPath); reportErr != nil {
				fmt.Fprintf(i.Logger, "Failed to generate error summary report: %s\n", reportErr.Error())
			} else {
				result.ErrorReportPath = reportPath
				fmt.Fprintf(i.Logger, "Comprehensive error report generated at: %s\n", reportPath)
				
				// Add error statistics to the result
				if i.EnhancedErrorHandler != nil {
					metrics := i.EnhancedErrorHandler.GetErrorMetrics()
					if metrics != nil {
						fmt.Fprintf(i.Logger, "Error statistics: %d total errors, %d recovered, %d unrecovered\n", 
							metrics.TotalErrors, metrics.RecoveredErrors, metrics.UnrecoveredErrors)
						
						// Log error statistics
						if i.ReportingSystem != nil {
							i.ReportingSystem.LogImportEvent(bundleID, "Error statistics: %d total, %d recovered, %d unrecovered, %d retry attempts", 
								metrics.TotalErrors, metrics.RecoveredErrors, metrics.UnrecoveredErrors, metrics.RetryAttempts)
						}
					}
				}
			}
		}
	}
	
	// Log success or failure message
	if result.Success {
		fmt.Fprintf(logger, "Bundle import successful: %d items imported in %v\n", 
			len(result.ImportedItems), result.Duration)
	} else {
		fmt.Fprintf(logger, "Bundle import completed with errors: %d items imported, %d errors in %v\n", 
			len(result.ImportedItems), len(result.Errors), result.Duration)
	}
	
	if i.ReportingSystem != nil {
		if result.Success {
			i.ReportingSystem.LogImportEvent(bundleID, "Bundle import completed successfully: %d items imported in %v", 
				len(result.ImportedItems), result.Duration)
		} else {
			i.ReportingSystem.LogImportEvent(bundleID, "Bundle import completed with errors: %d items imported, %d errors in %v", 
				len(result.ImportedItems), len(result.Errors), result.Duration)
		}
		i.ReportingSystem.AddPerformanceMetric(bundleID, "total_import_time", result.Duration.Seconds())
	}
	
	// Log audit event for import completion
	if auditLogger != nil {
		// Log the standard import completion event
		auditLogger.LogImportComplete(bundleID, result.Success, map[string]interface{}{
			"imported_items": len(result.ImportedItems),
			"skipped_items": len(result.SkippedFiles),
			"errors": len(result.Errors),
			"warnings": len(result.Warnings),
			"duration_seconds": result.Duration.Seconds(),
			"error_report_path": result.ErrorReportPath,
		})
		
		// Log a comprehensive summary of the import process
		auditLogger.LogImportSummary(bundleID, map[string]interface{}{
			"imported_items": len(result.ImportedItems),
			"skipped_items": len(result.SkippedFiles),
			"errors": len(result.Errors),
			"warnings": len(result.Warnings),
			"duration_seconds": result.Duration.Seconds(),
			"success": result.Success,
			"bundle_id": result.BundleID,
			"bundle_name": result.BundleName,
			"bundle_version": result.BundleVersion,
		})
	}
	
	// Return the result
	return result, nil
}

// ValidateBeforeImport validates a bundle before importing
func (i *DefaultBundleImporter) ValidateBeforeImport(ctx context.Context, bundlePath string, level ValidationLevel) (*ValidationResult, error) {
	// Open the bundle
	bundle, err := OpenBundle(bundlePath)
	if err != nil {
		return &ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("Failed to open bundle: %s", err.Error()),
			Errors:  []string{err.Error()},
		}, err
	}

	// Validate the bundle
	return i.Validator.Validate(bundle, level)
}

// CreateBackup creates a backup of the target directory
func (i *DefaultBundleImporter) CreateBackup(ctx context.Context, targetDir, backupDir string) (string, error) {
	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return "", fmt.Errorf("target directory does not exist: %w", err)
	}

	// Create backup directory if it doesn't exist
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		err = os.MkdirAll(backupDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create backup directory: %w", err)
		}
	}

	// Create backup name with timestamp
	backupName := fmt.Sprintf("backup-%s", time.Now().Format("20060102-150405"))
	backupPath := filepath.Join(backupDir, backupName)

	// Create backup directory
	err := os.Mkdir(backupPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy target directory to backup
	err = copyDir(targetDir, backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy target directory to backup: %w", err)
	}

	return backupPath, nil
}

// RestoreBackup restores a backup
func (i *DefaultBundleImporter) RestoreBackup(ctx context.Context, backupPath, targetDir string) error {
	// Check if backup directory exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup directory does not exist: %w", err)
	}

	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// Create target directory
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}
	} else {
		// Clear target directory
		err = clearDirectory(targetDir)
		if err != nil {
			return fmt.Errorf("failed to clear target directory: %w", err)
		}
	}

	// Copy backup to target directory
	err := copyDir(backupPath, targetDir)
	if err != nil {
		return fmt.Errorf("failed to copy backup to target directory: %w", err)
	}

	return nil
}

// checkForConflicts checks for conflicts between the bundle and the target directory
func (i *DefaultBundleImporter) checkForConflicts(ctx context.Context, tempDir, targetDir string, content []ContentItem) ([]string, error) {
	conflicts := []string{}

	for _, item := range content {
		dstPath := filepath.Join(targetDir, item.Path)

		// Check if destination exists
		if _, err := os.Stat(dstPath); err == nil {
			conflicts = append(conflicts, item.Path)
		}
	}

	return conflicts, nil
}

// Note: clearDirectory function is defined in utils.go
