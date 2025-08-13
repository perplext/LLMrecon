// Package update provides functionality for checking and applying updates
package update

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/version"
)

// UpdateExecutionOptions contains options for update execution
type UpdateExecutionOptions struct {
	// InstallDir is the directory where the tool is installed
	InstallDir string
	// TempDir is the directory for temporary files during update
	TempDir string
	// BackupDir is the directory for backups during update
	BackupDir string
	// LogDir is the directory for log files
	LogDir string
	// CurrentVersions contains the current versions of components
	CurrentVersions map[string]version.Version
	// Logger is the logger for update operations
	Logger io.Writer
	// MinLogLevel is the minimum log level to output
	MinLogLevel LogLevel
	// IncludeLogDetails determines whether to include details in log output
	IncludeLogDetails bool
	// NotificationHandlers is the list of notification handlers
	NotificationHandlers []NotificationHandler
	// User is the user performing the update
	User string
	// PreUpdateHooks are functions to run before the update
	PreUpdateHooks []UpdateHook
	// PostUpdateHooks are functions to run after the update
	PostUpdateHooks []UpdateHook
	// VerificationHooks are functions to run for verification
	VerificationHooks []VerificationHook
}

// UpdateHook is a function that runs before or after an update
type UpdateHook func(ctx context.Context, transaction *UpdateTransaction) error

// VerificationHook is a function that runs for verification
type VerificationHook func(ctx context.Context, pkg *UpdatePackage) (*VerificationResult, error)

// UpdateExecutor handles the execution of updates
type UpdateExecutor struct {
	// InstallDir is the directory where the tool is installed
	InstallDir string
	// TempDir is the directory for temporary files during update
	TempDir string
	// BackupDir is the directory for backups during update
	BackupDir string
	// LogDir is the directory for log files
	LogDir string
	// CurrentVersions contains the current versions of components
	CurrentVersions map[string]version.Version
	// Logger is the logger for update operations
	Logger *UpdateLogger
	// AuditLogger is the logger for audit events
	AuditLogger *AuditLogger
	// NotificationManager is the manager for notifications
	NotificationManager *NotificationManager
	// Verifier is the integrity verifier
	Verifier *IntegrityVerifier
	// User is the user performing the update
	User string
	// PreUpdateHooks are functions to run before the update
	PreUpdateHooks []UpdateHook
	// PostUpdateHooks are functions to run after the update
	PostUpdateHooks []UpdateHook
	// VerificationHooks are functions to run for verification
	VerificationHooks []VerificationHook
}

// NewUpdateExecutor creates a new update executor
func NewUpdateExecutor(options *UpdateExecutionOptions) (*UpdateExecutor, error) {
	// Validate options
	if options.InstallDir == "" {
		return nil, fmt.Errorf("install directory is required")
	}

	// Set default temp directory
	tempDir := options.TempDir
	if tempDir == "" {
		tempDir = filepath.Join(os.TempDir(), "LLMrecon-update")
	}

	// Set default backup directory
	backupDir := options.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join(options.InstallDir, "backups")
	}

	// Set default log directory
	logDir := options.LogDir
	if logDir == "" {
		logDir = filepath.Join(options.InstallDir, "logs")
	}

	// Create directories
	dirs := []string{tempDir, backupDir, logDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Set default logger
	logWriter := options.Logger
	if logWriter == nil {
		logWriter = os.Stdout
	}

	// Create update logger
	updateLogger := NewUpdateLogger(&LoggerOptions{
		Writer:         logWriter,
		MinLevel:       options.MinLogLevel,
		IncludeDetails: options.IncludeLogDetails,
	})

	// Create audit logger
	auditLogPath := filepath.Join(logDir, "audit.log")
	auditLogFile, err := os.OpenFile(auditLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit log file: %w", err)
	}

	auditLogger := NewAuditLogger(auditLogFile)

	// Create notification manager
	notificationManager := NewNotificationManager()

	// Add notification handlers
	for _, handler := range options.NotificationHandlers {
		notificationManager.AddHandler(handler)
	}

	// Add default console notification handler if no handlers provided
	if len(options.NotificationHandlers) == 0 {
		notificationManager.AddHandler(NewConsoleNotificationHandler(logWriter))
	}

	// Create integrity verifier
	verifier := NewIntegrityVerifier(logWriter)

	return &UpdateExecutor{
		InstallDir:          options.InstallDir,
		TempDir:             tempDir,
		BackupDir:           backupDir,
		LogDir:              logDir,
		CurrentVersions:     options.CurrentVersions,
		Logger:              updateLogger,
		AuditLogger:         auditLogger,
		NotificationManager: notificationManager,
		Verifier:            verifier,
		User:                options.User,
		PreUpdateHooks:      options.PreUpdateHooks,
		PostUpdateHooks:     options.PostUpdateHooks,
		VerificationHooks:   options.VerificationHooks,
	}, nil
}
