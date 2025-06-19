// Package errors provides error handling functionality for bundle operations
package errors

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ErrorCategory defines the category of an error
type ErrorCategory string

const (
	// ValidationError indicates an error during bundle validation
	ValidationError ErrorCategory = "validation"
	// FileSystemError indicates an error related to file system operations
	FileSystemError ErrorCategory = "filesystem"
	// NetworkError indicates an error related to network operations
	NetworkError ErrorCategory = "network"
	// ConfigurationError indicates an error related to configuration
	ConfigurationError ErrorCategory = "configuration"
	// PermissionError indicates an error related to permissions
	PermissionError ErrorCategory = "permission"
	// BackupError indicates an error related to backup operations
	BackupError ErrorCategory = "backup"
	// ConflictError indicates an error related to file conflicts
	ConflictError ErrorCategory = "conflict"
	// UnknownError indicates an error of unknown category
	UnknownError ErrorCategory = "unknown"
)

// ErrorSeverity defines the severity of an error
type ErrorSeverity string

const (
	// CriticalSeverity indicates a critical error that prevents the operation from continuing
	CriticalSeverity ErrorSeverity = "critical"
	// HighSeverity indicates a high severity error that may affect the operation
	HighSeverity ErrorSeverity = "high"
	// MediumSeverity indicates a medium severity error that should be addressed
	MediumSeverity ErrorSeverity = "medium"
	// LowSeverity indicates a low severity error that is not critical
	LowSeverity ErrorSeverity = "low"
)

// ErrorRecoverability defines whether an error is recoverable
type ErrorRecoverability string

const (
	// RecoverableError indicates an error that can be recovered from
	RecoverableError ErrorRecoverability = "recoverable"
	// NonRecoverableError indicates an error that cannot be recovered from
	NonRecoverableError ErrorRecoverability = "non-recoverable"
)

// BundleError represents a structured error with additional metadata
type BundleError struct {
	// Original is the original error
	Original error
	// Message is a human-readable error message
	Message string
	// Category is the category of the error
	Category ErrorCategory
	// Severity is the severity of the error
	Severity ErrorSeverity
	// Recoverability indicates whether the error is recoverable
	Recoverability ErrorRecoverability
	// Context contains additional context about the error
	Context map[string]interface{}
	// ErrorID is a unique identifier for the error
	ErrorID string
	// Timestamp is the time the error occurred
	Timestamp time.Time
	// RetryAttempt is the current retry attempt
	RetryAttempt int
	// MaxRetries is the maximum number of retries
	MaxRetries int
}

// Error returns the error message
func (e *BundleError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Original.Error())
	}
	return e.Message
}

// WithContext adds context to the error
func (e *BundleError) WithContext(key string, value interface{}) *BundleError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewBundleError creates a new bundle error
func NewBundleError(err error, message string, category ErrorCategory, severity ErrorSeverity, recoverability ErrorRecoverability) *BundleError {
	return &BundleError{
		Original:       err,
		Message:        message,
		Category:       category,
		Severity:       severity,
		Recoverability: recoverability,
		Context:        make(map[string]interface{}),
		ErrorID:        generateErrorID(),
		Timestamp:      time.Now(),
		RetryAttempt:   0,
		MaxRetries:     3,
	}
}

// generateErrorID generates a unique error ID
func generateErrorID() string {
	// Generate a random string for the error ID
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 10
	
	// Create a random number generator with a time-based seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Generate a random string
	b := strings.Builder{}
	b.Grow(length)
	for i := 0; i < length; i++ {
		b.WriteByte(charset[r.Intn(len(charset))])
	}
	
	return fmt.Sprintf("ERR-%s-%d", b.String(), time.Now().Unix())
}

// ErrorHandler defines the interface for error handling
type ErrorHandler interface {
	// HandleError handles an error
	HandleError(ctx context.Context, err error) (*BundleError, error)
	// ShouldRetry determines whether an error should be retried
	ShouldRetry(ctx context.Context, err *BundleError) bool
	// GetBackoffDuration returns the backoff duration for a retry
	GetBackoffDuration(ctx context.Context, err *BundleError) time.Duration
	// LogError logs an error
	LogError(ctx context.Context, err *BundleError)
	// NotifyAdmin notifies an administrator about a critical error
	NotifyAdmin(ctx context.Context, err *BundleError) error
}

// DefaultErrorHandler is the default implementation of ErrorHandler
type DefaultErrorHandler struct {
	// Logger is the logger for error handling
	Logger ErrorLogger
	// AdminNotificationEnabled indicates whether admin notifications are enabled
	AdminNotificationEnabled bool
	// AdminEmail is the email address for admin notifications
	AdminEmail string
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger ErrorLogger) ErrorHandler {
	return &DefaultErrorHandler{
		Logger:                  logger,
		AdminNotificationEnabled: false,
	}
}

// HandleError handles an error
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error) (*BundleError, error) {
	// If it's already a BundleError, return it
	if bundleErr, ok := err.(*BundleError); ok {
		h.LogError(ctx, bundleErr)
		return bundleErr, err
	}

	// Categorize the error
	bundleErr := h.categorizeError(err)
	h.LogError(ctx, bundleErr)

	// Notify admin for critical errors
	if bundleErr.Severity == CriticalSeverity && h.AdminNotificationEnabled {
		notifyErr := h.NotifyAdmin(ctx, bundleErr)
		if notifyErr != nil {
			// Log notification failure but don't fail the operation
			if h.Logger != nil {
				h.Logger.LogEventWithStatus("admin_notification_failed", "ErrorHandler", bundleErr.ErrorID, "failure", map[string]interface{}{
					"error":     notifyErr.Error(),
					"original_error": bundleErr.Error(),
					"operation": "admin_notification",
				})
			}
		}
	}

	return bundleErr, err
}

// categorizeError categorizes an error
func (h *DefaultErrorHandler) categorizeError(err error) *BundleError {
	// Default to unknown error
	category := UnknownError
	severity := MediumSeverity
	recoverability := RecoverableError
	
	// Categorize based on error message
	errMsg := err.Error()
	
	// Check for file system errors
	if strings.Contains(errMsg, "no such file") || 
	   strings.Contains(errMsg, "file exists") || 
	   strings.Contains(errMsg, "permission denied") {
		category = FileSystemError
		severity = HighSeverity
	}
	
	// Check for network errors
	if strings.Contains(errMsg, "connection") || 
	   strings.Contains(errMsg, "timeout") || 
	   strings.Contains(errMsg, "network") {
		category = NetworkError
		severity = HighSeverity
	}
	
	// Check for validation errors
	if strings.Contains(errMsg, "invalid") || 
	   strings.Contains(errMsg, "validation") || 
	   strings.Contains(errMsg, "not valid") {
		category = ValidationError
		severity = HighSeverity
		recoverability = NonRecoverableError
	}
	
	// Create the bundle error
	return &BundleError{
		Original:       err,
		Message:        fmt.Sprintf("Error occurred: %s", err.Error()),
		Category:       category,
		Severity:       severity,
		Recoverability: recoverability,
		Context:        make(map[string]interface{}),
		ErrorID:        generateErrorID(),
		Timestamp:      time.Now(),
		RetryAttempt:   0,
		MaxRetries:     3,
	}
}

// ShouldRetry determines whether an error should be retried
func (h *DefaultErrorHandler) ShouldRetry(ctx context.Context, err *BundleError) bool {
	// Don't retry if error is not recoverable
	if err.Recoverability == NonRecoverableError {
		return false
	}
	
	// Don't retry if max retries reached
	if err.RetryAttempt >= err.MaxRetries {
		return false
	}
	
	// Retry based on error category
	switch err.Category {
	case NetworkError:
		// Always retry network errors
		return true
	case FileSystemError:
		// Retry file system errors
		return true
	case BackupError:
		// Retry backup errors
		return true
	case ConflictError:
		// Don't retry conflict errors by default
		return false
	default:
		// Don't retry other errors by default
		return false
	}
}

// GetBackoffDuration returns the backoff duration for a retry
func (h *DefaultErrorHandler) GetBackoffDuration(ctx context.Context, err *BundleError) time.Duration {
	// Use exponential backoff with jitter
	baseDelay := 500 * time.Millisecond
	maxDelay := 30 * time.Second
	
	// Calculate delay with exponential backoff
	delay := baseDelay * time.Duration(1<<uint(err.RetryAttempt))
	if delay > maxDelay {
		delay = maxDelay
	}
	
	// Add jitter (±20%)
	jitter := rand.Float64()*0.4 - 0.2
	delay = time.Duration(float64(delay) * (1 + jitter))
	
	return delay
}

// LogError logs an error
func (h *DefaultErrorHandler) LogError(ctx context.Context, err *BundleError) {
	if h.Logger == nil {
		return
	}
	
	// Create error details
	details := map[string]interface{}{
		"error_id":       err.ErrorID,
		"message":        err.Message,
		"category":       string(err.Category),
		"severity":       string(err.Severity),
		"recoverability": string(err.Recoverability),
		"timestamp":      err.Timestamp.Format(time.RFC3339),
		"retry_attempt":  err.RetryAttempt,
		"max_retries":    err.MaxRetries,
		"operation":      "error_handling",
	}
	
	// Add original error if available
	if err.Original != nil {
		details["original_error"] = err.Original.Error()
	}
	
	// Add context if available
	for k, v := range err.Context {
		details[k] = v
	}
	
	// Log the error with appropriate status
	status := "error"
	if err.Severity == CriticalSeverity {
		status = "critical"
	}
	
	h.Logger.LogEventWithStatus("error_occurred", "ErrorHandler", err.ErrorID, status, details)
}

// NotifyAdmin notifies an administrator about a critical error
func (h *DefaultErrorHandler) NotifyAdmin(ctx context.Context, err *BundleError) error {
	// This is a placeholder for actual admin notification
	// In a real implementation, this would send an email or notification
	
	if h.Logger != nil {
		h.Logger.LogEventWithStatus("admin_notification_sent", "ErrorHandler", err.ErrorID, "success", map[string]interface{}{
			"admin_email": h.AdminEmail,
			"error_id":    err.ErrorID,
			"message":     err.Message,
			"severity":    string(err.Severity),
			"operation":   "admin_notification",
		})
	}
	
	return nil
}

// WithRetry executes a function with retry logic
func WithRetry(ctx context.Context, handler ErrorHandler, fn func() error) error {
	var err error
	var bundleErr *BundleError
	
	// Execute the function
	err = fn()
	if err == nil {
		return nil
	}
	
	// Handle the error
	bundleErr, _ = handler.HandleError(ctx, err)
	
	// Retry if needed
	for handler.ShouldRetry(ctx, bundleErr) {
		// Increment retry attempt
		bundleErr.RetryAttempt++
		
		// Wait for backoff duration
		backoff := handler.GetBackoffDuration(ctx, bundleErr)
		time.Sleep(backoff)
		
		// Try again
		err = fn()
		if err == nil {
			return nil
		}
		
		// Handle the error
		bundleErr, _ = handler.HandleError(ctx, err)
	}
	
	return err
}
