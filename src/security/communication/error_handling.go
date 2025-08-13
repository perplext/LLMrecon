// Package communication provides secure communication utilities for the LLMrecon tool.
package communication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ErrorLevel represents the severity level of an error
type ErrorLevel int

const (
	// ErrorLevelInfo is for informational errors
	ErrorLevelInfo ErrorLevel = iota
	// ErrorLevelWarning is for warning errors
	ErrorLevelWarning
	// ErrorLevelError is for standard errors
	ErrorLevelError
	// ErrorLevelCritical is for critical errors
	ErrorLevelCritical
)

// SecureError represents a secure error that doesn't leak sensitive information
type SecureError struct {
	// Code is a unique error code
	Code string `json:"code"`
	// Message is a user-friendly error message
	Message string `json:"message"`
	// Level is the severity level of the error
	Level ErrorLevel `json:"level"`
	// Details is additional information (only shown in development mode)
	Details string `json:"-"`
	// OriginalError is the original error (not exposed to clients)
	OriginalError error `json:"-"`
}

// Error implements the error interface
func (e *SecureError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewSecureError creates a new secure error
func NewSecureError(code string, message string, level ErrorLevel, originalError error) *SecureError {
	return &SecureError{
		Code:          code,
		Message:       message,
		Level:         level,
		OriginalError: originalError,
	}
}

// WithDetails adds details to a secure error
func (e *SecureError) WithDetails(details string) *SecureError {
	e.Details = details
	return e
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	// Code is a unique error code
	Code string `json:"code"`
	// Message is a user-friendly error message
	Message string `json:"message"`
	// RequestID is a unique identifier for the request
	RequestID string `json:"request_id,omitempty"`
}

// ErrorHandler handles errors securely
type ErrorHandler struct {
	// DevelopmentMode indicates whether to include detailed error information
	DevelopmentMode bool
	// ErrorCodeMap maps error types to error codes
	ErrorCodeMap map[string]string
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(developmentMode bool) *ErrorHandler {
	return &ErrorHandler{
		DevelopmentMode: developmentMode,
		ErrorCodeMap:    make(map[string]string),
	}
}

// RegisterErrorCode registers an error code for an error type
func (h *ErrorHandler) RegisterErrorCode(errorType string, code string) {
	h.ErrorCodeMap[errorType] = code
}

// GetErrorCode gets the error code for an error
func (h *ErrorHandler) GetErrorCode(err error) string {
	if secureErr, ok := err.(*SecureError); ok {
		return secureErr.Code
	}

	// Try to match the error type
	errType := fmt.Sprintf("%T", err)
	if code, ok := h.ErrorCodeMap[errType]; ok {
		return code
	}

	// Default error code
	return "INTERNAL_ERROR"
}

// HandleError handles an error securely
func (h *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error, defaultMessage string) {
	// Get the status code based on the error
	statusCode := h.getStatusCodeForError(err)

	// Create the error response
	response := ErrorResponse{
		Code:      h.GetErrorCode(err),
		Message:   h.getErrorMessage(err, defaultMessage),
		RequestID: r.Header.Get("X-Request-ID"),
	}

	// Set the content type
	w.Header().Set("Content-Type", "application/json")

	// Set the status code
	w.WriteHeader(statusCode)

	// Write the response
	json.NewEncoder(w).Encode(response)
}

// getStatusCodeForError gets the HTTP status code for an error
func (h *ErrorHandler) getStatusCodeForError(err error) int {
	// Check for specific error types
	if secureErr, ok := err.(*SecureError); ok {
		switch secureErr.Level {
		case ErrorLevelInfo:
			return http.StatusOK
		case ErrorLevelWarning:
			return http.StatusBadRequest
		case ErrorLevelError:
			return http.StatusInternalServerError
		case ErrorLevelCritical:
			return http.StatusServiceUnavailable
		}
	}

	// Check for common error patterns
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "not found"):
		return http.StatusNotFound
	case strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "unauthenticated"):
		return http.StatusUnauthorized
	case strings.Contains(errStr, "forbidden"):
		return http.StatusForbidden
	case strings.Contains(errStr, "bad request") || strings.Contains(errStr, "invalid"):
		return http.StatusBadRequest
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		return http.StatusGatewayTimeout
	case strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "too many requests"):
		return http.StatusTooManyRequests
	}

	// Default to internal server error
	return http.StatusInternalServerError
}

// getErrorMessage gets a user-friendly error message
func (h *ErrorHandler) getErrorMessage(err error, defaultMessage string) string {
	// If it's a secure error, use its message
	if secureErr, ok := err.(*SecureError); ok {
		// In development mode, include details if available
		if h.DevelopmentMode && secureErr.Details != "" {
			return fmt.Sprintf("%s (%s)", secureErr.Message, secureErr.Details)
		}
		return secureErr.Message
	}

	// In development mode, use the actual error message
	if h.DevelopmentMode {
		return err.Error()
	}

	// In production mode, use the default message
	return defaultMessage
}

// SanitizeErrorForLogging sanitizes an error for logging
func (h *ErrorHandler) SanitizeErrorForLogging(err error) string {
	errStr := err.Error()

	// Redact sensitive information
	patterns := []string{
		`api[_-]?key\s*[:=]\s*[A-Za-z0-9_\-]{5,}`,
		`token\s*[:=]\s*[A-Za-z0-9_\-]{5,}`,
		`password\s*[:=]\s*[^\s]{5,}`,
		`secret\s*[:=]\s*[^\s]{5,}`,
		`authorization\s*[:=]\s*[^\s]{5,}`,
	}

	for _, pattern := range patterns {
		errStr = strings.ReplaceAll(errStr, pattern, "[REDACTED]")
	}

	return errStr
}
