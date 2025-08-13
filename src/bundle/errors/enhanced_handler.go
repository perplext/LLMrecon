// Package errors provides error handling functionality for bundle operations
package errors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// EnhancedErrorHandler extends the DefaultErrorHandler with advanced error handling capabilities
type EnhancedErrorHandler struct {
	DefaultErrorHandler
	// MaxRetries is the maximum number of retries for recoverable errors
	MaxRetries int
	// RetryableCategories is a list of error categories that are retryable
	RetryableCategories map[ErrorCategory]bool
	// BackoffBaseSeconds is the base time in seconds for exponential backoff
	BackoffBaseSeconds float64
	// BackoffMaxSeconds is the maximum backoff time in seconds
	BackoffMaxSeconds float64
	// BackoffJitterFactor is the jitter factor for backoff (0-1)
	BackoffJitterFactor float64
	// NotificationThreshold is the severity threshold for admin notifications
	NotificationThreshold ErrorSeverity
	// ErrorMetrics tracks error metrics for reporting
	ErrorMetrics *ErrorMetrics
}

// ErrorMetrics tracks error statistics for reporting
type ErrorMetrics struct {
	// TotalErrors is the total number of errors
	TotalErrors int
	// ErrorsByCategory tracks errors by category
	ErrorsByCategory map[ErrorCategory]int
	// ErrorsBySeverity tracks errors by severity
	ErrorsBySeverity map[ErrorSeverity]int
	// RecoveredErrors is the number of errors that were recovered from
	RecoveredErrors int
	// UnrecoveredErrors is the number of errors that were not recovered from
	UnrecoveredErrors int
	// RetryAttempts is the total number of retry attempts
	RetryAttempts int
	// SuccessfulRetries is the number of successful retries
	SuccessfulRetries int
}

// NewEnhancedErrorHandler creates a new enhanced error handler
func NewEnhancedErrorHandler(logger ErrorLogger) *EnhancedErrorHandler {
	// Create default error metrics
	metrics := &ErrorMetrics{
		ErrorsByCategory:  make(map[ErrorCategory]int),
		ErrorsBySeverity:  make(map[ErrorSeverity]int),
		TotalErrors:       0,
		RecoveredErrors:   0,
		UnrecoveredErrors: 0,
		RetryAttempts:     0,
		SuccessfulRetries: 0,
	}

	// Create default retryable categories
	retryableCategories := map[ErrorCategory]bool{
		NetworkError:    true,
		FileSystemError: true,
		// Configuration and permission errors are typically not retryable
		ConfigurationError: false,
		PermissionError:    false,
		ValidationError:    false,
		BackupError:        true,
		ConflictError:      false,
		UnknownError:       true, // Retry unknown errors by default
	}

	return &EnhancedErrorHandler{
		DefaultErrorHandler: DefaultErrorHandler{
			Logger:                  logger,
			AdminNotificationEnabled: true,
		},
		MaxRetries:           3,
		RetryableCategories:  retryableCategories,
		BackoffBaseSeconds:   1.0,
		BackoffMaxSeconds:    60.0,
		BackoffJitterFactor:  0.1,
		NotificationThreshold: HighSeverity,
		ErrorMetrics:         metrics,
	}
}

// HandleError handles an error with enhanced categorization and tracking
func (h *EnhancedErrorHandler) HandleError(ctx context.Context, err error) (*BundleError, error) {
	// Check if the error is already a BundleError
	var bundleErr *BundleError
	if be, ok := err.(*BundleError); ok {
		bundleErr = be
	} else {
		// Categorize the error
		bundleErr = h.categorizeError(err)
	}

	// Update error metrics
	h.ErrorMetrics.TotalErrors++
	h.ErrorMetrics.ErrorsByCategory[bundleErr.Category]++
	h.ErrorMetrics.ErrorsBySeverity[bundleErr.Severity]++

	// Log the error
	h.LogError(ctx, bundleErr)

	// Determine if the error is recoverable
	if bundleErr.Recoverability == NonRecoverableError {
		h.ErrorMetrics.UnrecoveredErrors++
	}

	// Notify admin for critical errors
	if h.shouldNotifyAdmin(bundleErr) {
		// Don't block on notification, but log any notification errors
		go func() {
			if notifyErr := h.NotifyAdmin(ctx, bundleErr); notifyErr != nil {
				h.LogError(ctx, &BundleError{
					Original:       notifyErr,
					Message:        "Failed to send admin notification",
					Category:       ConfigurationError,
					Severity:       LowSeverity,
					Recoverability: RecoverableError,
				})
			}
		}()
	}

	return bundleErr, err
}

// ShouldRetry determines whether an error should be retried with enhanced logic
func (h *EnhancedErrorHandler) ShouldRetry(ctx context.Context, err *BundleError) bool {
	// Don't retry if error is nil
	if err == nil {
		return false
	}

	// Don't retry if we've reached the maximum number of retries
	if err.RetryAttempt >= h.MaxRetries {
		return false
	}

	// Don't retry if the error is not recoverable
	if err.Recoverability == NonRecoverableError {
		return false
	}

	// Check if the error category is retryable
	if retryable, ok := h.RetryableCategories[err.Category]; ok && !retryable {
		return false
	}

	// Check if the context is cancelled
	if ctx.Err() != nil {
		return false
	}

	// Update retry metrics
	h.ErrorMetrics.RetryAttempts++

	return true
}

// GetBackoffDuration returns the backoff duration for a retry with exponential backoff
func (h *EnhancedErrorHandler) GetBackoffDuration(ctx context.Context, err *BundleError) time.Duration {
	// Calculate exponential backoff
	attempt := float64(err.RetryAttempt)
	backoffSeconds := h.BackoffBaseSeconds * math.Pow(2, attempt)

	// Apply maximum backoff
	if backoffSeconds > h.BackoffMaxSeconds {
		backoffSeconds = h.BackoffMaxSeconds
	}

	// Apply jitter to prevent thundering herd
	jitter := (rand.Float64() * 2 - 1) * h.BackoffJitterFactor * backoffSeconds
	backoffSeconds = backoffSeconds + jitter

	// Ensure backoff is positive
	if backoffSeconds < 0 {
		backoffSeconds = 0
	}

	return time.Duration(backoffSeconds * float64(time.Second))
}

// LogError logs an error with enhanced details
func (h *EnhancedErrorHandler) LogError(ctx context.Context, err *BundleError) {
	if h.Logger == nil || err == nil {
		return
	}

	// Prepare error details
	details := map[string]interface{}{
		"error_id":       err.ErrorID,
		"category":       string(err.Category),
		"severity":       string(err.Severity),
		"recoverability": string(err.Recoverability),
		"timestamp":      err.Timestamp.Format(time.RFC3339),
		"retry_attempt":  err.RetryAttempt,
		"max_retries":    err.MaxRetries,
	}

	// Add context to details
	for k, v := range err.Context {
		details[k] = v
	}

	// Log the error
	h.Logger.LogEventWithStatus(
		"error",
		"ErrorHandler",
		err.ErrorID,
		string(err.Severity),
		details,
	)
}

// NotifyAdmin notifies an administrator about a critical error with enhanced details
func (h *EnhancedErrorHandler) NotifyAdmin(ctx context.Context, err *BundleError) error {
	if !h.AdminNotificationEnabled || h.AdminEmail == "" {
		return fmt.Errorf("admin notification not configured")
	}

	// In a real implementation, this would send an email or other notification
	// For now, we'll just log the notification
	if h.Logger != nil {
		h.Logger.LogEventWithStatus(
			"admin_notification",
			"ErrorHandler",
			err.ErrorID,
			"sent",
			map[string]interface{}{
				"admin_email": h.AdminEmail,
				"error_id":    err.ErrorID,
				"message":     err.Message,
				"category":    string(err.Category),
				"severity":    string(err.Severity),
			},
		)
	}

	return nil
}

// shouldNotifyAdmin determines if an admin should be notified about an error
func (h *EnhancedErrorHandler) shouldNotifyAdmin(err *BundleError) bool {
	// Don't notify if notifications are disabled
	if !h.AdminNotificationEnabled {
		return false
	}

	// Determine severity threshold for notification
	switch h.NotificationThreshold {
	case CriticalSeverity:
		return err.Severity == CriticalSeverity
	case HighSeverity:
		return err.Severity == CriticalSeverity || err.Severity == HighSeverity
	case MediumSeverity:
		return err.Severity == CriticalSeverity || err.Severity == HighSeverity || err.Severity == MediumSeverity
	case LowSeverity:
		return true
	default:
		return err.Severity == CriticalSeverity || err.Severity == HighSeverity
	}
}

// GetErrorMetrics returns the current error metrics
func (h *EnhancedErrorHandler) GetErrorMetrics() *ErrorMetrics {
	return h.ErrorMetrics
}

// ResetErrorMetrics resets the error metrics
func (h *EnhancedErrorHandler) ResetErrorMetrics() {
	h.ErrorMetrics = &ErrorMetrics{
		ErrorsByCategory:  make(map[ErrorCategory]int),
		ErrorsBySeverity:  make(map[ErrorSeverity]int),
		TotalErrors:       0,
		RecoveredErrors:   0,
		UnrecoveredErrors: 0,
		RetryAttempts:     0,
		SuccessfulRetries: 0,
	}
}

// WithRetryAndContext executes a function with enhanced retry logic and context awareness
func WithRetryAndContext(ctx context.Context, handler *EnhancedErrorHandler, fn func() error) error {
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
		// Check if context is cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		// Increment retry attempt
		bundleErr.RetryAttempt++
		
		// Wait for backoff duration
		backoff := handler.GetBackoffDuration(ctx, bundleErr)
		
		// Use a timer with context to support cancellation
		timer := time.NewTimer(backoff)
		select {
		case <-timer.C:
			// Timer completed, continue with retry
		case <-ctx.Done():
			// Context cancelled, stop retrying
			timer.Stop()
			return ctx.Err()
		}
		
		// Try again
		err = fn()
		if err == nil {
			// Update metrics for successful retry
			handler.ErrorMetrics.SuccessfulRetries++
			handler.ErrorMetrics.RecoveredErrors++
			return nil
		}
		
		// Handle the error
		bundleErr, _ = handler.HandleError(ctx, err)
	}
	
	return err
}
