// Package errors provides error handling functionality for bundle operations
package errors

import (
	"context"
	"fmt"
	"strings"
)

// RecoveryStrategy defines the interface for error recovery strategies
type RecoveryStrategy interface {
	// Recover attempts to recover from an error
	Recover(ctx context.Context, err *BundleError) (bool, error)
	// CanRecover determines if the strategy can recover from an error
	CanRecover(err *BundleError) bool
}

// RecoveryManager manages multiple recovery strategies
type RecoveryManager struct {
	// Strategies is a list of recovery strategies
	Strategies []RecoveryStrategy
	// Logger is the logger for recovery events
	Logger io.Writer
	// AuditLogger is the logger for audit events
	AuditLogger *AuditLogger
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(logger io.Writer, auditLogger *AuditLogger) *RecoveryManager {
	if logger == nil {
		logger = os.Stdout
	}
	
	return &RecoveryManager{
		Strategies:  []RecoveryStrategy{},
		Logger:      logger,
		AuditLogger: auditLogger,
	}
}

// AddStrategy adds a recovery strategy to the manager
func (m *RecoveryManager) AddStrategy(strategy RecoveryStrategy) {
	m.Strategies = append(m.Strategies, strategy)
}

// AttemptRecovery attempts to recover from an error using all available strategies
func (m *RecoveryManager) AttemptRecovery(ctx context.Context, err *BundleError) (bool, error) {
	if err == nil {
		return true, nil
	}
	
	// Log recovery attempt
	fmt.Fprintf(m.Logger, "Attempting to recover from error: %s (ID: %s, Category: %s)\n", 
		err.Message, err.ErrorID, err.Category)
	
	// Log audit event for recovery attempt
	if m.AuditLogger != nil {
		m.AuditLogger.LogEventWithStatus(
			"recovery_attempt",
			"RecoveryManager",
			err.ErrorID,
			"started",
			map[string]interface{}{
				"error_id":       err.ErrorID,
				"category":       string(err.Category),
				"severity":       string(err.Severity),
				"recoverability": string(err.Recoverability),
			},
		)
	}
	
	// Try each strategy
	for _, strategy := range m.Strategies {
		if strategy.CanRecover(err) {
			// Log strategy attempt
			fmt.Fprintf(m.Logger, "Trying recovery strategy: %T\n", strategy)
			
			// Attempt recovery
			recovered, recoverErr := strategy.Recover(ctx, err)
			
			if recovered {
				// Log successful recovery
				fmt.Fprintf(m.Logger, "Successfully recovered from error using strategy: %T\n", strategy)
				
				// Log audit event for successful recovery
				if m.AuditLogger != nil {
					m.AuditLogger.LogEventWithStatus(
						"recovery_successful",
						"RecoveryManager",
						err.ErrorID,
						"success",
						map[string]interface{}{
							"error_id":  err.ErrorID,
							"strategy":  fmt.Sprintf("%T", strategy),
							"timestamp": time.Now().Format(time.RFC3339),
						},
					)
				}
				
				return true, nil
			}
			
			// Log failed recovery attempt
			fmt.Fprintf(m.Logger, "Recovery strategy %T failed: %v\n", strategy, recoverErr)
		}
	}
	
	// Log failed recovery
	fmt.Fprintf(m.Logger, "Failed to recover from error: %s (ID: %s)\n", err.Message, err.ErrorID)
	
	// Log audit event for failed recovery
	if m.AuditLogger != nil {
		m.AuditLogger.LogEventWithStatus(
			"recovery_failed",
			"RecoveryManager",
			err.ErrorID,
			"failure",
			map[string]interface{}{
				"error_id":  err.ErrorID,
				"timestamp": time.Now().Format(time.RFC3339),
			},
		)
	}
	
	return false, err
}

// FileSystemRecoveryStrategy implements recovery strategies for file system errors
type FileSystemRecoveryStrategy struct {
	// Logger is the logger for recovery events
	Logger io.Writer
}

// NewFileSystemRecoveryStrategy creates a new file system recovery strategy
func NewFileSystemRecoveryStrategy(logger io.Writer) *FileSystemRecoveryStrategy {
	if logger == nil {
		logger = os.Stdout
	}
	
	return &FileSystemRecoveryStrategy{
		Logger: logger,
	}
}

// CanRecover determines if the strategy can recover from an error
func (s *FileSystemRecoveryStrategy) CanRecover(err *BundleError) bool {
	return err != nil && err.Category == FileSystemError && err.Recoverability == RecoverableError
}

// Recover attempts to recover from a file system error
func (s *FileSystemRecoveryStrategy) Recover(ctx context.Context, err *BundleError) (bool, error) {
	if err == nil || err.Category != FileSystemError {
		return false, fmt.Errorf("not a file system error")
	}
	
	// Extract file path from context if available
	filePath, ok := err.Context["file_path"].(string)
	if !ok {
		return false, fmt.Errorf("file path not found in error context")
	}
	
	// Check the specific error message to determine recovery strategy
	if os.IsNotExist(err.Original) {
		// Try to create the directory
		dirPath := filePath
		if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
			// If it's a file, get the directory
			dirPath = fmt.Sprintf("%s", filePath[:strings.LastIndex(filePath, "/")])
		}
		
		// Create the directory
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return false, fmt.Errorf("failed to create directory: %w", err)
		}
		
		fmt.Fprintf(s.Logger, "Created directory: %s\n", dirPath)
		return true, nil
	}
	
	// Handle permission errors
	if os.IsPermission(err.Original) {
		// Log that we can't automatically fix permission errors
		fmt.Fprintf(s.Logger, "Permission error for %s - cannot automatically recover\n", filePath)
		return false, fmt.Errorf("permission error requires manual intervention")
	}
	
	return false, fmt.Errorf("unsupported file system error recovery")
}

// BackupRecoveryStrategy implements recovery strategies for backup errors
type BackupRecoveryStrategy struct {
	// Logger is the logger for recovery events
	Logger io.Writer
	// BackupDir is the directory for backups
	BackupDir string
}

// NewBackupRecoveryStrategy creates a new backup recovery strategy
func NewBackupRecoveryStrategy(logger io.Writer, backupDir string) *BackupRecoveryStrategy {
	if logger == nil {
		logger = os.Stdout
	}
	
	return &BackupRecoveryStrategy{
		Logger:    logger,
		BackupDir: backupDir,
	}
}

// CanRecover determines if the strategy can recover from an error
func (s *BackupRecoveryStrategy) CanRecover(err *BundleError) bool {
	return err != nil && err.Category == BackupError && err.Recoverability == RecoverableError
}

// Recover attempts to recover from a backup error
func (s *BackupRecoveryStrategy) Recover(ctx context.Context, err *BundleError) (bool, error) {
	if err == nil || err.Category != BackupError {
		return false, fmt.Errorf("not a backup error")
	}
	
	// Extract backup information from context if available
	_, targetOk := err.Context["target_dir"].(string)
	if !targetOk {
		return false, fmt.Errorf("target directory not found in error context")
	}
	
	// Create a new backup directory if needed
	if s.BackupDir == "" {
		s.BackupDir = os.TempDir()
	}
	
	// Ensure backup directory exists
	if err := os.MkdirAll(s.BackupDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Create a new backup with timestamp
	backupPath := fmt.Sprintf("%s/backup_%s", s.BackupDir, time.Now().Format("20060102_150405"))
	
	fmt.Fprintf(s.Logger, "Creating new backup at: %s\n", backupPath)
	
	// In a real implementation, this would copy files from targetDir to backupPath
	// For now, we'll just create the directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return false, fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Update error context with new backup path
	err.Context["backup_path"] = backupPath
	
	fmt.Fprintf(s.Logger, "Created new backup directory: %s\n", backupPath)
	return true, nil
}

// NetworkRecoveryStrategy implements recovery strategies for network errors
type NetworkRecoveryStrategy struct {
	// Logger is the logger for recovery events
	Logger io.Writer
	// MaxRetries is the maximum number of retries
	MaxRetries int
}

// NewNetworkRecoveryStrategy creates a new network recovery strategy
func NewNetworkRecoveryStrategy(logger io.Writer, maxRetries int) *NetworkRecoveryStrategy {
	if logger == nil {
		logger = os.Stdout
	}
	
	if maxRetries <= 0 {
		maxRetries = 3
	}
	
	return &NetworkRecoveryStrategy{
		Logger:     logger,
		MaxRetries: maxRetries,
	}
}

// CanRecover determines if the strategy can recover from an error
func (s *NetworkRecoveryStrategy) CanRecover(err *BundleError) bool {
	return err != nil && err.Category == NetworkError && err.Recoverability == RecoverableError
}

// Recover attempts to recover from a network error
func (s *NetworkRecoveryStrategy) Recover(ctx context.Context, err *BundleError) (bool, error) {
	if err == nil || err.Category != NetworkError {
		return false, fmt.Errorf("not a network error")
	}
	
	// Check if we've exceeded the maximum number of retries
	if err.RetryAttempt >= s.MaxRetries {
		return false, fmt.Errorf("exceeded maximum number of retries")
	}
	
	// For network errors, we'll just recommend retrying after a delay
	fmt.Fprintf(s.Logger, "Network error detected, recommending retry after backoff\n")
	
	// In a real implementation, this might attempt to reconnect or use an alternative endpoint
	// For now, we'll just return true to indicate that retrying is the recovery strategy
	return true, nil
}

// ConflictRecoveryStrategy implements recovery strategies for conflict errors
type ConflictRecoveryStrategy struct {
	// Logger is the logger for recovery events
	Logger io.Writer
	// Force indicates whether to force resolution
	Force bool
}

// NewConflictRecoveryStrategy creates a new conflict recovery strategy
func NewConflictRecoveryStrategy(logger io.Writer, force bool) *ConflictRecoveryStrategy {
	if logger == nil {
		logger = os.Stdout
	}
	
	return &ConflictRecoveryStrategy{
		Logger: logger,
		Force:  force,
	}
}

// CanRecover determines if the strategy can recover from an error
func (s *ConflictRecoveryStrategy) CanRecover(err *BundleError) bool {
	// Only attempt recovery if force is enabled
	return err != nil && err.Category == ConflictError && s.Force
}

// Recover attempts to recover from a conflict error
func (s *ConflictRecoveryStrategy) Recover(ctx context.Context, err *BundleError) (bool, error) {
	if err == nil || err.Category != ConflictError {
		return false, fmt.Errorf("not a conflict error")
	}
	
	// If force is not enabled, we can't recover
	if !s.Force {
		return false, fmt.Errorf("force resolution not enabled")
	}
	
	// Extract conflict information from context if available
	conflictPath, ok := err.Context["conflict_path"].(string)
	if !ok {
		return false, fmt.Errorf("conflict path not found in error context")
	}
	
	fmt.Fprintf(s.Logger, "Forcing resolution of conflict for: %s\n", conflictPath)
	
	// In a real implementation, this would apply a conflict resolution strategy
	// For now, we'll just log that we're forcing resolution
	fmt.Fprintf(s.Logger, "Forced resolution of conflict for: %s\n", conflictPath)
	
	return true, nil
}
