// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// StagedImportPhase represents a phase in the staged import process
type StagedImportPhase string

const (
	// ValidationPhase represents the validation phase
	ValidationPhase StagedImportPhase = "validation"
	// ExtractionPhase represents the extraction phase
	ExtractionPhase StagedImportPhase = "extraction"
	// BackupPhase represents the backup phase
	BackupPhase StagedImportPhase = "backup"
	// VerificationPhase represents the verification phase
	VerificationPhase StagedImportPhase = "verification"
	// InstallationPhase represents the installation phase
	InstallationPhase StagedImportPhase = "installation"
	// CleanupPhase represents the cleanup phase
	CleanupPhase StagedImportPhase = "cleanup"
)

// StagedImportOptions represents options for staged import
type StagedImportOptions struct {
	// ValidationLevel is the level of validation to perform
	ValidationLevel ValidationLevel
	// TargetDir is the directory to import the bundle to
	TargetDir string
	// BackupDir is the directory to store backups in
	BackupDir string
	// TempDir is the directory to use for temporary files
	TempDir string
	// Force indicates whether to force import even if there are conflicts
	Force bool
	// KeepBackup indicates whether to keep the backup after successful import
	KeepBackup bool
	// Logger is the logger for import operations
	Logger io.Writer
	// ProgressCallback is called with progress updates
	ProgressCallback func(phase StagedImportPhase, progress float64, message string)
}

// StagedImportStatus represents the status of a staged import
type StagedImportStatus struct {
	// Phase is the current phase of the import
	Phase StagedImportPhase
	// Progress is the progress of the current phase (0-100)
	Progress float64
	// Message is a human-readable message about the current status
	Message string
	// StartTime is the time the import started
	StartTime time.Time
	// CurrentPhaseStartTime is the time the current phase started
	CurrentPhaseStartTime time.Time
	// EndTime is the time the import ended (if completed)
	EndTime time.Time
	// Success indicates whether the import was successful (if completed)
	Success bool
	// Error is the error that occurred (if any)
	Error error
	// ValidationResult is the result of the validation phase
	ValidationResult *ValidationResult
	// BackupPath is the path to the backup (if created)
	BackupPath string
	// ImportedItems are the items that were imported
	ImportedItems []ContentItem
}

// StagedImporter defines the interface for staged import
type StagedImporter interface {
	// Import performs a staged import
	Import(ctx context.Context, bundlePath string, options StagedImportOptions) (*ImportResult, error)
	// GetStatus returns the current status of the import
	GetStatus() *StagedImportStatus
	// Cancel cancels the import
	Cancel() error
}

// DefaultStagedImporter is the default implementation of StagedImporter
type DefaultStagedImporter struct {
	// Importer is the bundle importer
	Importer BundleImporter
	// Status is the current status of the import
	Status *StagedImportStatus
	// CancelCh is a channel for cancellation
	CancelCh chan struct{}
}

// NewStagedImporter creates a new staged importer
func NewStagedImporter(importer BundleImporter) StagedImporter {
	if importer == nil {
		importer = NewBundleImporter(nil, nil, nil)
	}
	return &DefaultStagedImporter{
		Importer: importer,
		Status: &StagedImportStatus{
			Phase:    ValidationPhase,
			Progress: 0,
			Message:  "Initializing",
		},
		CancelCh: make(chan struct{}),
	}
}

// Import performs a staged import
func (i *DefaultStagedImporter) Import(ctx context.Context, bundlePath string, options StagedImportOptions) (*ImportResult, error) {
	// Initialize status
	i.Status = &StagedImportStatus{
		Phase:               ValidationPhase,
		Progress:            0,
		Message:             "Starting import",
		StartTime:           time.Now(),
		CurrentPhaseStartTime: time.Now(),
	}

	// Set default logger if not provided
	logger := options.Logger
	if logger == nil {
		logger = os.Stdout
	}

	// Create result
	result := &ImportResult{
		Success:       false,
		StartTime:     i.Status.StartTime,
		ImportTime:    i.Status.StartTime,
		Errors:        []string{},
		Warnings:      []string{},
		ImportedFiles: []string{},
		SkippedFiles:  []string{},
	}

	// Set up progress callback
	progressCallback := options.ProgressCallback
	if progressCallback == nil {
		progressCallback = func(phase StagedImportPhase, progress float64, message string) {
			// No-op
		}
	}

	// Update progress
	updateProgress := func(phase StagedImportPhase, progress float64, message string) {
		if phase != i.Status.Phase {
			i.Status.Phase = phase
			i.Status.CurrentPhaseStartTime = time.Now()
		}
		i.Status.Progress = progress
		i.Status.Message = message
		progressCallback(phase, progress, message)
	}

	// Log start
	fmt.Fprintf(logger, "Starting staged import: %s\n", bundlePath)
	updateProgress(ValidationPhase, 0, "Starting validation")

	// Check for cancellation
	select {
	case <-i.CancelCh:
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = fmt.Errorf("import cancelled")
		result.Message = "Import cancelled"
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	case <-ctx.Done():
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = ctx.Err()
		result.Message = fmt.Sprintf("Import cancelled: %s", ctx.Err())
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	default:
		// Continue
	}

	// Phase 1: Validation
	updateProgress(ValidationPhase, 10, "Validating bundle")
	validationResult, err := i.Importer.ValidateBeforeImport(ctx, bundlePath, options.ValidationLevel)
	if err != nil {
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = err
		i.Status.ValidationResult = validationResult
		result.Message = fmt.Sprintf("Validation failed: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.ValidationResult = validationResult
		result.EndTime = i.Status.EndTime
		return result, err
	}
	i.Status.ValidationResult = validationResult
	result.ValidationResult = validationResult

	// If validation failed, return error
	if !validationResult.Valid {
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = fmt.Errorf("bundle validation failed")
		result.Message = "Validation failed"
		result.Errors = append(result.Errors, validationResult.Errors...)
		result.Warnings = append(result.Warnings, validationResult.Warnings...)
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	}

	updateProgress(ValidationPhase, 100, "Validation successful")

	// Check for cancellation
	select {
	case <-i.CancelCh:
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = fmt.Errorf("import cancelled")
		result.Message = "Import cancelled"
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	case <-ctx.Done():
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = ctx.Err()
		result.Message = fmt.Sprintf("Import cancelled: %s", ctx.Err())
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	default:
		// Continue
	}

	// Phase 2: Extraction
	updateProgress(ExtractionPhase, 0, "Starting extraction")

	// Open the bundle
	bundle, err := OpenBundle(bundlePath)
	if err != nil {
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = err
		result.Message = fmt.Sprintf("Failed to open bundle: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = i.Status.EndTime
		return result, err
	}

	// Create temporary directory for extraction
	tempDir := options.TempDir
	if tempDir == "" {
		tempDir, err = os.MkdirTemp("", "bundle-import-*")
		if err != nil {
			i.Status.Success = false
			i.Status.EndTime = time.Now()
			i.Status.Error = err
			result.Message = fmt.Sprintf("Failed to create temporary directory: %s", err.Error())
			result.Errors = append(result.Errors, err.Error())
			result.EndTime = i.Status.EndTime
			return result, err
		}
		defer os.RemoveAll(tempDir)
	}

	updateProgress(ExtractionPhase, 20, "Extracting bundle")

	// Extract bundle to temporary directory
	err = ExtractBundle(bundlePath, tempDir)
	if err != nil {
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = err
		result.Message = fmt.Sprintf("Failed to extract bundle: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = i.Status.EndTime
		return result, err
	}

	updateProgress(ExtractionPhase, 100, "Extraction successful")

	// Check for cancellation
	select {
	case <-i.CancelCh:
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = fmt.Errorf("import cancelled")
		result.Message = "Import cancelled"
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	case <-ctx.Done():
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = ctx.Err()
		result.Message = fmt.Sprintf("Import cancelled: %s", ctx.Err())
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	default:
		// Continue
	}

	// Phase 3: Backup
	updateProgress(BackupPhase, 0, "Starting backup")

	// Create backup if backup directory is provided
	var backupPath string
	if options.BackupDir != "" {
		updateProgress(BackupPhase, 50, "Creating backup")
		backupPath, err = i.Importer.CreateBackup(ctx, options.TargetDir, options.BackupDir)
		if err != nil {
			i.Status.Success = false
			i.Status.EndTime = time.Now()
			i.Status.Error = err
			result.Message = fmt.Sprintf("Failed to create backup: %s", err.Error())
			result.Errors = append(result.Errors, err.Error())
			result.EndTime = i.Status.EndTime
			return result, err
		}
		i.Status.BackupPath = backupPath
		result.BackupPath = backupPath
		fmt.Fprintf(logger, "Created backup at: %s\n", backupPath)
	}

	updateProgress(BackupPhase, 100, "Backup successful")

	// Check for cancellation
	select {
	case <-i.CancelCh:
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = fmt.Errorf("import cancelled")
		result.Message = "Import cancelled"
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	case <-ctx.Done():
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = ctx.Err()
		result.Message = fmt.Sprintf("Import cancelled: %s", ctx.Err())
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	default:
		// Continue
	}

	// Phase 4: Verification
	updateProgress(VerificationPhase, 0, "Starting verification")

	// Check for conflicts
	updateProgress(VerificationPhase, 50, "Checking for conflicts")
	conflicts, err := i.checkForConflicts(ctx, tempDir, options.TargetDir, bundle.Manifest.Content)
	if err != nil {
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = err
		result.Message = fmt.Sprintf("Failed to check for conflicts: %s", err.Error())
		result.Errors = append(result.Errors, err.Error())
		result.EndTime = i.Status.EndTime
		return result, err
	}

	// If there are conflicts and force is not enabled, return error
	if len(conflicts) > 0 && !options.Force {
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = fmt.Errorf("import would overwrite existing files")
		result.Message = fmt.Sprintf("Import would overwrite %d existing files", len(conflicts))
		for _, conflict := range conflicts {
			result.Errors = append(result.Errors, fmt.Sprintf("Conflict: %s", conflict))
		}
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	}

	updateProgress(VerificationPhase, 100, "Verification successful")

	// Check for cancellation
	select {
	case <-i.CancelCh:
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = fmt.Errorf("import cancelled")
		result.Message = "Import cancelled"
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	case <-ctx.Done():
		i.Status.Success = false
		i.Status.EndTime = time.Now()
		i.Status.Error = ctx.Err()
		result.Message = fmt.Sprintf("Import cancelled: %s", ctx.Err())
		result.EndTime = i.Status.EndTime
		return result, i.Status.Error
	default:
		// Continue
	}

	// Phase 5: Installation
	updateProgress(InstallationPhase, 0, "Starting installation")

	// Import files
	importedItems := []ContentItem{}
	totalItems := len(bundle.Manifest.Content)
	for idx, item := range bundle.Manifest.Content {
		// Update progress
		progress := float64(idx) / float64(totalItems) * 100
		updateProgress(InstallationPhase, progress, fmt.Sprintf("Installing %s (%d/%d)", item.Path, idx+1, totalItems))

		srcPath := filepath.Join(tempDir, item.Path)
		dstPath := filepath.Join(options.TargetDir, item.Path)

		// Create parent directories
		err = os.MkdirAll(filepath.Dir(dstPath), 0755)
		if err != nil {
			i.Status.Success = false
			i.Status.EndTime = time.Now()
			i.Status.Error = err
			result.Message = fmt.Sprintf("Failed to create directory for %s: %s", item.Path, err.Error())
			result.Errors = append(result.Errors, err.Error())
			
			// Attempt to restore backup if one was created
			if i.Status.BackupPath != "" {
				restoreErr := i.Importer.RestoreBackup(ctx, i.Status.BackupPath, options.TargetDir)
				if restoreErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore backup: %s", restoreErr.Error()))
				} else {
					result.Warnings = append(result.Warnings, "Restored backup after failed import")
				}
			}
			
			result.EndTime = i.Status.EndTime
			return result, err
		}

		// Copy file or directory
		err = copyPath(srcPath, dstPath)
		if err != nil {
			i.Status.Success = false
			i.Status.EndTime = time.Now()
			i.Status.Error = err
			result.Message = fmt.Sprintf("Failed to copy %s: %s", item.Path, err.Error())
			result.Errors = append(result.Errors, err.Error())
			
			// Attempt to restore backup if one was created
			if i.Status.BackupPath != "" {
				restoreErr := i.Importer.RestoreBackup(ctx, i.Status.BackupPath, options.TargetDir)
				if restoreErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore backup: %s", restoreErr.Error()))
				} else {
					result.Warnings = append(result.Warnings, "Restored backup after failed import")
				}
			}
			
			result.EndTime = i.Status.EndTime
			return result, err
		}

		// Add to imported items
		importedItems = append(importedItems, item)
		fmt.Fprintf(logger, "Imported: %s\n", item.Path)

		// Check for cancellation
		select {
		case <-i.CancelCh:
			i.Status.Success = false
			i.Status.EndTime = time.Now()
			i.Status.Error = fmt.Errorf("import cancelled")
			result.Message = "Import cancelled"
			result.EndTime = i.Status.EndTime
			
			// Attempt to restore backup if one was created
			if i.Status.BackupPath != "" {
				restoreErr := i.Importer.RestoreBackup(ctx, i.Status.BackupPath, options.TargetDir)
				if restoreErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore backup: %s", restoreErr.Error()))
				} else {
					result.Warnings = append(result.Warnings, "Restored backup after cancelled import")
				}
			}
			
			return result, i.Status.Error
		case <-ctx.Done():
			i.Status.Success = false
			i.Status.EndTime = time.Now()
			i.Status.Error = ctx.Err()
			result.Message = fmt.Sprintf("Import cancelled: %s", ctx.Err())
			result.EndTime = i.Status.EndTime
			
			// Attempt to restore backup if one was created
			if i.Status.BackupPath != "" {
				restoreErr := i.Importer.RestoreBackup(ctx, i.Status.BackupPath, options.TargetDir)
				if restoreErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to restore backup: %s", restoreErr.Error()))
				} else {
					result.Warnings = append(result.Warnings, "Restored backup after cancelled import")
				}
			}
			
			return result, i.Status.Error
		default:
			// Continue
		}
	}

	i.Status.ImportedItems = importedItems
	result.ImportedItems = importedItems
	updateProgress(InstallationPhase, 100, "Installation successful")

	// Phase 6: Cleanup
	updateProgress(CleanupPhase, 0, "Starting cleanup")

	// Remove backup if not keeping it
	if i.Status.BackupPath != "" && !options.KeepBackup {
		updateProgress(CleanupPhase, 50, "Removing backup")
		err = os.RemoveAll(i.Status.BackupPath)
		if err != nil {
			// Just log the error, don't fail the import
			fmt.Fprintf(logger, "Warning: Failed to remove backup: %s\n", err.Error())
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to remove backup: %s", err.Error()))
		}
	}

	updateProgress(CleanupPhase, 100, "Cleanup successful")

	// Set success
	i.Status.Success = true
	i.Status.EndTime = time.Now()
	result.Success = true
	result.Message = fmt.Sprintf("Successfully imported %d items", len(result.ImportedItems))
	result.EndTime = i.Status.EndTime

	// Log import success
	fmt.Fprintf(logger, "Bundle import successful: %d items imported\n", len(result.ImportedItems))

	return result, nil
}

// GetStatus returns the current status of the import
func (i *DefaultStagedImporter) GetStatus() *StagedImportStatus {
	return i.Status
}

// Cancel cancels the import
func (i *DefaultStagedImporter) Cancel() error {
	select {
	case i.CancelCh <- struct{}{}:
		return nil
	default:
		// Channel already closed or full
		return nil
	}
}

// checkForConflicts checks for conflicts between the bundle and the target directory
func (i *DefaultStagedImporter) checkForConflicts(ctx context.Context, tempDir, targetDir string, content []ContentItem) ([]string, error) {
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
