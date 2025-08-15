// Package errors provides error handling functionality for bundle operations
package errors

import (
	"encoding/json"
	"fmt"
	"time"
)

// ErrorCode represents a type of error
type ErrorCode string

// Error codes
const (
	// Import errors
	ImportErrorCode         ErrorCode = "IMPORT_ERROR"
	ValidationErrorCode     ErrorCode = "VALIDATION_ERROR"
	ConversionErrorCode     ErrorCode = "CONVERSION_ERROR"
	FileSystemErrorCode     ErrorCode = "FILE_SYSTEM_ERROR"
	NetworkErrorCode        ErrorCode = "NETWORK_ERROR"
	ConfigurationErrorCode  ErrorCode = "CONFIGURATION_ERROR"
	SecurityErrorCode       ErrorCode = "SECURITY_ERROR"
	PermissionErrorCode     ErrorCode = "PERMISSION_ERROR"
	BackupErrorCode         ErrorCode = "BACKUP_ERROR"
	ConflictErrorCode       ErrorCode = "CONFLICT_ERROR"
	UnknownErrorCode        ErrorCode = "UNKNOWN_ERROR"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

// Error categories
const (
	ValidationError     ErrorCategory = "validation"
	FileSystemError     ErrorCategory = "filesystem"
	NetworkError        ErrorCategory = "network"
	ConfigurationError  ErrorCategory = "configuration"
	SecurityError       ErrorCategory = "security"
	PermissionError     ErrorCategory = "permission"
	BackupError         ErrorCategory = "backup"
	ConflictError       ErrorCategory = "conflict"
	UnknownError        ErrorCategory = "unknown"
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

// Error severities
const (
	LowSeverity      ErrorSeverity = "low"
	MediumSeverity   ErrorSeverity = "medium"
	HighSeverity     ErrorSeverity = "high"
	CriticalSeverity ErrorSeverity = "critical"
)

// ErrorRecoverability represents whether an error is recoverable
type ErrorRecoverability string

// Error recoverability types
const (
	RecoverableError    ErrorRecoverability = "recoverable"
	NonRecoverableError ErrorRecoverability = "non_recoverable"
)

// BundleError represents a structured error with additional context
type BundleError struct {
	// ErrorID is a unique identifier for this error instance
	ErrorID string `json:"error_id"`
	
	// Code is the error code
	Code ErrorCode `json:"code"`
	
	// Category is the error category
	Category ErrorCategory `json:"category"`
	
	// Severity is the error severity
	Severity ErrorSeverity `json:"severity"`
	
	// Recoverability indicates if the error is recoverable
	Recoverability ErrorRecoverability `json:"recoverability"`
	
	// Message is the human-readable error message
	Message string `json:"message"`
	
	// Details contains additional error details
	Details map[string]interface{} `json:"details,omitempty"`
	
	// Context contains contextual information
	Context map[string]interface{} `json:"context,omitempty"`
	
	// Timestamp is when the error occurred
	Timestamp time.Time `json:"timestamp"`
	
	// StackTrace contains the stack trace
	StackTrace string `json:"stack_trace,omitempty"`
	
	// RetryAttempt is the current retry attempt
	RetryAttempt int `json:"retry_attempt"`
	
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries"`
	
	// Cause is the underlying error
	Cause error `json:"-"`
}

// Error implements the error interface
func (e *BundleError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

// Unwrap returns the underlying error
func (e *BundleError) Unwrap() error {
	return e.Cause
}

// ToJSON converts the error to JSON
func (e *BundleError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// NewBundleError creates a new BundleError
func NewBundleError(code ErrorCode, category ErrorCategory, severity ErrorSeverity, recoverability ErrorRecoverability, message string) *BundleError {
	return &BundleError{
		ErrorID:        generateErrorID(),
		Code:           code,
		Category:       category,
		Severity:       severity,
		Recoverability: recoverability,
		Message:        message,
		Details:        make(map[string]interface{}),
		Context:        make(map[string]interface{}),
		Timestamp:      time.Now(),
		RetryAttempt:   0,
		MaxRetries:     3,
	}
}

// generateErrorID generates a unique error ID
func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

// WithDetails adds details to the error
func (e *BundleError) WithDetails(details map[string]interface{}) *BundleError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithContext adds context to the error
func (e *BundleError) WithContext(context map[string]interface{}) *BundleError {
	for k, v := range context {
		e.Context[k] = v
	}
	return e
}

// WithCause sets the underlying cause
func (e *BundleError) WithCause(cause error) *BundleError {
	e.Cause = cause
	return e
}

// WithRetryInfo sets retry information
func (e *BundleError) WithRetryInfo(attempt, maxRetries int) *BundleError {
	e.RetryAttempt = attempt
	e.MaxRetries = maxRetries
	return e
}
