// Package errors provides error handling functionality for bundle operations
package errors

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

// ErrorCategorizer defines the interface for error categorization
type ErrorCategorizer interface {
	// CategorizeError categorizes an error
	CategorizeError(err error) (ErrorCategory, ErrorSeverity, ErrorRecoverability)
}

// DefaultErrorCategorizer is the default implementation of ErrorCategorizer
type DefaultErrorCategorizer struct {
	// Logger is the logger for categorization events
	Logger io.Writer
}

// NewErrorCategorizer creates a new error categorizer
func NewErrorCategorizer(logger io.Writer) ErrorCategorizer {
	if logger == nil {
		logger = os.Stdout
	}
	return &DefaultErrorCategorizer{
		Logger: logger,
	}
}

// CategorizeError categorizes an error based on its type and message
func (c *DefaultErrorCategorizer) CategorizeError(err error) (ErrorCategory, ErrorSeverity, ErrorRecoverability) {
	if err == nil {
		return UnknownError, LowSeverity, RecoverableError
	}

	// Check if the error is already a BundleError
	if be, ok := err.(*BundleError); ok {
		return be.Category, be.Severity, be.Recoverability
	}

	// Get the error message
	errMsg := err.Error()
	
	// Check for validation errors
	if strings.Contains(errMsg, "validation") || 
	   strings.Contains(errMsg, "invalid") || 
	   strings.Contains(errMsg, "schema") {
		return ValidationError, HighSeverity, NonRecoverableError
	}
	
	// Check for file system errors
	if os.IsNotExist(err) || 
	   os.IsPermission(err) || 
	   strings.Contains(errMsg, "file") || 
	   strings.Contains(errMsg, "directory") || 
	   strings.Contains(errMsg, "path") {
		
		// Check for permission errors specifically
		if os.IsPermission(err) {
			return PermissionError, HighSeverity, NonRecoverableError
		}
		
		// Check for not found errors
		if os.IsNotExist(err) {
			return FileSystemError, MediumSeverity, NonRecoverableError
		}
		
		return FileSystemError, MediumSeverity, RecoverableError
	}
	
	// Check for network errors
	if strings.Contains(errMsg, "network") || 
	   strings.Contains(errMsg, "connection") || 
	   strings.Contains(errMsg, "timeout") || 
	   strings.Contains(errMsg, "unreachable") {
		return NetworkError, MediumSeverity, RecoverableError
	}
	
	// Check for configuration errors
	if strings.Contains(errMsg, "config") || 
	   strings.Contains(errMsg, "configuration") || 
	   strings.Contains(errMsg, "settings") {
		return ConfigurationError, MediumSeverity, NonRecoverableError
	}
	
	// Check for backup errors
	if strings.Contains(errMsg, "backup") {
		return BackupError, HighSeverity, RecoverableError
	}
	
	// Check for conflict errors
	if strings.Contains(errMsg, "conflict") {
		return ConflictError, MediumSeverity, NonRecoverableError
	}
	
	// Check for system call errors
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.EACCES, syscall.EPERM:
			return PermissionError, HighSeverity, NonRecoverableError
		case syscall.ENOENT:
			return FileSystemError, MediumSeverity, NonRecoverableError
		case syscall.ENOSPC:
			return FileSystemError, CriticalSeverity, NonRecoverableError
		case syscall.ETIMEDOUT, syscall.ECONNREFUSED, syscall.ECONNRESET:
			return NetworkError, MediumSeverity, RecoverableError
		}
	}
	
	// Default to unknown error
	return UnknownError, MediumSeverity, RecoverableError
}

// GetErrorDetails returns detailed information about an error
func GetErrorDetails(err error) map[string]interface{} {
	if err == nil {
		return map[string]interface{}{"error": "nil"}
	}
	
	details := map[string]interface{}{
		"error":     err.Error(),
		"error_type": fmt.Sprintf("%T", err),
	}
	
	// Extract additional details from BundleError
	if be, ok := err.(*BundleError); ok {
		details["category"] = string(be.Category)
		details["severity"] = string(be.Severity)
		details["recoverability"] = string(be.Recoverability)
		details["error_id"] = be.ErrorID
		details["timestamp"] = be.Timestamp.Format("2006-01-02T15:04:05Z07:00")
		details["retry_attempt"] = be.RetryAttempt
		details["max_retries"] = be.MaxRetries
		
		// Add context
		for k, v := range be.Context {
			details[k] = v
		}
	}
	
	return details
}

// IsRetryableError determines if an error is retryable
func IsRetryableError(err error) bool {
	// Check if the error is a BundleError
	if be, ok := err.(*BundleError); ok {
		return be.Recoverability == RecoverableError
	}
	
	// Categorize the error
	categorizer := NewErrorCategorizer(nil)
	_, _, recoverability := categorizer.CategorizeError(err)
	
	return recoverability == RecoverableError
}

// GetErrorSeverity returns the severity of an error
func GetErrorSeverity(err error) ErrorSeverity {
	// Check if the error is a BundleError
	if be, ok := err.(*BundleError); ok {
		return be.Severity
	}
	
	// Categorize the error
	categorizer := NewErrorCategorizer(nil)
	_, severity, _ := categorizer.CategorizeError(err)
	
	return severity
}

// GetErrorCategory returns the category of an error
func GetErrorCategory(err error) ErrorCategory {
	// Check if the error is a BundleError
	if be, ok := err.(*BundleError); ok {
		return be.Category
	}
	
	// Categorize the error
	categorizer := NewErrorCategorizer(nil)
	category, _, _ := categorizer.CategorizeError(err)
	
	return category
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	
	return fmt.Errorf("%s: %w", message, err)
}
