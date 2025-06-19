// Package api provides API protection mechanisms for the LLMrecon tool.
package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	// LogLevelDebug is for debug logs
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for informational logs
	LogLevelInfo
	// LogLevelWarning is for warning logs
	LogLevelWarning
	// LogLevelError is for error logs
	LogLevelError
)

// SecureLoggerConfig represents the configuration for a secure logger
type SecureLoggerConfig struct {
	// Level is the minimum log level to record
	Level LogLevel
	// LogRequests indicates whether to log requests
	LogRequests bool
	// LogResponses indicates whether to log responses
	LogResponses bool
	// LogHeaders indicates whether to log headers
	LogHeaders bool
	// LogBodies indicates whether to log request and response bodies
	LogBodies bool
	// MaxBodySize is the maximum size of a body to log
	MaxBodySize int
	// SensitiveHeaders is a list of headers that contain sensitive information
	SensitiveHeaders []string
	// SensitiveFields is a list of JSON fields that contain sensitive information
	SensitiveFields []string
	// SensitivePatterns is a list of regex patterns for sensitive information
	SensitivePatterns []string
	// RedactionString is the string to use for redaction
	RedactionString string
	// OutputWriter is the writer to use for log output
	OutputWriter io.Writer
}

// DefaultSecureLoggerConfig returns the default secure logger configuration
func DefaultSecureLoggerConfig() *SecureLoggerConfig {
	return &SecureLoggerConfig{
		Level:           LogLevelInfo,
		LogRequests:     true,
		LogResponses:    true,
		LogHeaders:      true,
		LogBodies:       true,
		MaxBodySize:     10 * 1024, // 10KB
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"Set-Cookie",
			"X-Api-Key",
			"Api-Key",
			"Password",
		},
		SensitiveFields: []string{
			"password",
			"token",
			"api_key",
			"apiKey",
			"secret",
			"credential",
			"ssn",
			"credit_card",
			"creditCard",
			"cvv",
		},
		SensitivePatterns: []string{
			`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`,                           // Email
			`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`,                                               // US Phone
			`\b\d{3}[-.]?\d{2}[-.]?\d{4}\b`,                                               // SSN
			`\b(?:\d[ -]*?){13,16}\b`,                                                     // Credit Card
			`\b[A-Za-z0-9]{24,}\b`,                                                        // API Key
			`\bsk-[A-Za-z0-9]{24,}\b`,                                                     // OpenAI API Key
			`\b(?:Bearer|bearer|BEARER)\s+[A-Za-z0-9\-._~+/]+=*\b`,                        // Bearer Token
			`\b(?:eyJ|ey0)[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+\b`,            // JWT
			`\b[A-Za-z0-9]{8}-[A-Za-z0-9]{4}-[A-Za-z0-9]{4}-[A-Za-z0-9]{4}-[A-Za-z0-9]{12}\b`, // UUID
		},
		RedactionString: "[REDACTED]",
	}
}

// SecureLogger implements secure logging for API requests and responses
type SecureLogger struct {
	config            *SecureLoggerConfig
	sensitivePatterns []*regexp.Regexp
}

// NewSecureLogger creates a new secure logger
func NewSecureLogger(config *SecureLoggerConfig) (*SecureLogger, error) {
	if config == nil {
		config = DefaultSecureLoggerConfig()
	}

	// Compile regex patterns
	sensitivePatterns := make([]*regexp.Regexp, 0, len(config.SensitivePatterns))
	for _, pattern := range config.SensitivePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		sensitivePatterns = append(sensitivePatterns, re)
	}

	return &SecureLogger{
		config:            config,
		sensitivePatterns: sensitivePatterns,
	}, nil
}

// LogEntry represents a log entry
type LogEntry struct {
	// Timestamp is the time of the log entry
	Timestamp time.Time `json:"timestamp"`
	// Level is the log level
	Level LogLevel `json:"level"`
	// RequestID is the ID of the request
	RequestID string `json:"request_id,omitempty"`
	// Method is the HTTP method
	Method string `json:"method,omitempty"`
	// Path is the request path
	Path string `json:"path,omitempty"`
	// ClientIP is the client IP
	ClientIP string `json:"client_ip,omitempty"`
	// StatusCode is the response status code
	StatusCode int `json:"status_code,omitempty"`
	// Duration is the request duration in milliseconds
	Duration int64 `json:"duration,omitempty"`
	// Headers are the request or response headers
	Headers map[string]string `json:"headers,omitempty"`
	// Body is the request or response body
	Body string `json:"body,omitempty"`
	// Error is an error message
	Error string `json:"error,omitempty"`
	// Message is a log message
	Message string `json:"message,omitempty"`
}

// Middleware returns a middleware function for secure logging
func (sl *SecureLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging if level is too low
		if sl.config.Level > LogLevelInfo {
			next.ServeHTTP(w, r)
			return
		}

		// Generate a request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}

		// Get the client IP
		clientIP := r.RemoteAddr
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			clientIP = strings.Split(ip, ",")[0]
		}

		// Create a log entry for the request
		requestEntry := LogEntry{
			Timestamp: time.Now(),
			Level:     LogLevelInfo,
			RequestID: requestID,
			Method:    r.Method,
			Path:      r.URL.Path,
			ClientIP:  clientIP,
		}

		// Log request headers if enabled
		if sl.config.LogHeaders {
			requestEntry.Headers = sl.redactHeaders(r.Header)
		}

		// Log request body if enabled
		if sl.config.LogBodies && r.Body != nil {
			body, err := ioutil.ReadAll(r.Body)
			if err == nil {
				// Restore the body for the next handler
				r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

				// Redact sensitive information
				if len(body) <= sl.config.MaxBodySize {
					requestEntry.Body = sl.redactBody(string(body))
				} else {
					requestEntry.Body = "[BODY TOO LARGE]"
				}
			}
		}

		// Log the request
		if sl.config.LogRequests {
			sl.logEntry(requestEntry)
		}

		// Create a response wrapper to capture the response
		rw := newResponseWrapper(w)

		// Record the start time
		startTime := time.Now()

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Calculate the duration
		duration := time.Since(startTime)

		// Create a log entry for the response
		responseEntry := LogEntry{
			Timestamp:  time.Now(),
			Level:      LogLevelInfo,
			RequestID:  requestID,
			Method:     r.Method,
			Path:       r.URL.Path,
			ClientIP:   clientIP,
			StatusCode: rw.statusCode,
			Duration:   duration.Milliseconds(),
		}

		// Log response headers if enabled
		if sl.config.LogHeaders {
			responseEntry.Headers = sl.redactHeaders(rw.Header())
		}

		// Log response body if enabled
		if sl.config.LogBodies {
			body := rw.body.Bytes()
			if len(body) <= sl.config.MaxBodySize {
				responseEntry.Body = sl.redactBody(string(body))
			} else {
				responseEntry.Body = "[BODY TOO LARGE]"
			}
		}

		// Log the response
		if sl.config.LogResponses {
			sl.logEntry(responseEntry)
		}
	})
}

// redactHeaders redacts sensitive information from headers
func (sl *SecureLogger) redactHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	for key, values := range headers {
		// Check if this is a sensitive header
		isSensitive := false
		for _, sensitiveHeader := range sl.config.SensitiveHeaders {
			if strings.EqualFold(key, sensitiveHeader) {
				isSensitive = true
				break
			}
		}

		// Redact if sensitive, otherwise use the value
		if isSensitive {
			result[key] = sl.config.RedactionString
		} else {
			result[key] = strings.Join(values, ", ")
		}
	}

	return result
}

// redactBody redacts sensitive information from a body
func (sl *SecureLogger) redactBody(body string) string {
	// Try to parse as JSON
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(body), &jsonObj); err == nil {
		// If it's JSON, redact sensitive fields
		redactedJSON := sl.redactJSON(jsonObj)
		redactedBody, err := json.Marshal(redactedJSON)
		if err == nil {
			body = string(redactedBody)
		}
	}

	// Apply regex patterns to redact other sensitive information
	for _, pattern := range sl.sensitivePatterns {
		body = pattern.ReplaceAllString(body, sl.config.RedactionString)
	}

	return body
}

// redactJSON redacts sensitive information from a JSON object
func (sl *SecureLogger) redactJSON(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			// Check if this is a sensitive field
			isSensitive := false
			for _, sensitiveField := range sl.config.SensitiveFields {
				if strings.EqualFold(key, sensitiveField) {
					isSensitive = true
					break
				}
			}

			// Redact if sensitive, otherwise recurse
			if isSensitive {
				result[key] = sl.config.RedactionString
			} else {
				result[key] = sl.redactJSON(value)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = sl.redactJSON(value)
		}
		return result
	default:
		return v
	}
}

// logEntry logs a log entry
func (sl *SecureLogger) logEntry(entry LogEntry) {
	// Convert the entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// Write the entry to the output
	if sl.config.OutputWriter != nil {
		sl.config.OutputWriter.Write(data)
		sl.config.OutputWriter.Write([]byte("\n"))
	}
}

// Log logs a message
func (sl *SecureLogger) Log(level LogLevel, requestID string, message string, err error) {
	// Skip logging if level is too low
	if level < sl.config.Level {
		return
	}

	// Create a log entry
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		RequestID: requestID,
		Message:   message,
	}

	// Add error if present
	if err != nil {
		entry.Error = err.Error()
	}

	// Log the entry
	sl.logEntry(entry)
}

// responseWrapper wraps an http.ResponseWriter to capture the response
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// newResponseWrapper creates a new response wrapper
func newResponseWrapper(w http.ResponseWriter) *responseWrapper {
	return &responseWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           bytes.NewBuffer(nil),
	}
}

// WriteHeader captures the status code
func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body
func (rw *responseWrapper) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Flush implements the http.Flusher interface
func (rw *responseWrapper) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
