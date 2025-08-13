// Package middleware provides middleware components for the Multi-Provider LLM Integration Framework.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	// LogLevelDebug is for debug logging
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for info logging
	LogLevelInfo
	// LogLevelWarning is for warning logging
	LogLevelWarning
	// LogLevelError is for error logging
	LogLevelError
)

// LogEntry represents a log entry
type LogEntry struct {
	// Timestamp is the timestamp of the log entry
	Timestamp time.Time `json:"timestamp"`
	// Level is the log level
	Level LogLevel `json:"level"`
	// ProviderType is the type of provider
	ProviderType core.ProviderType `json:"provider_type"`
	// Operation is the operation being performed
	Operation string `json:"operation"`
	// RequestID is the ID of the request
	RequestID string `json:"request_id"`
	// Request is the request data
	Request interface{} `json:"request,omitempty"`
	// Response is the response data
	Response interface{} `json:"response,omitempty"`
	// Error is the error message
	Error string `json:"error,omitempty"`
	// Duration is the duration of the operation
	Duration time.Duration `json:"duration,omitempty"`
	// AdditionalInfo is additional information
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

// LogHandler is a function that handles log entries
type LogHandler func(entry *LogEntry)

// LoggingMiddleware provides logging functionality
type LoggingMiddleware struct {
	// handlers is a map of log levels to handlers
	handlers map[LogLevel][]LogHandler
	// minLevel is the minimum log level to log
	minLevel LogLevel
	// redactPII indicates whether to redact PII
	redactPII bool
	// redactPatterns is a list of patterns to redact
	redactPatterns []*regexp.Regexp
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(minLevel LogLevel, redactPII bool) *LoggingMiddleware {
	middleware := &LoggingMiddleware{
		handlers:  make(map[LogLevel][]LogHandler),
		minLevel:  minLevel,
		redactPII: redactPII,
	}

	if redactPII {
		// Add default PII redaction patterns
		middleware.AddRedactPattern(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`, "[EMAIL]")
		middleware.AddRedactPattern(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`, "[PHONE]")
		middleware.AddRedactPattern(`\b\d{3}[-]?\d{2}[-]?\d{4}\b`, "[SSN]")
		middleware.AddRedactPattern(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|6(?:011|5[0-9]{2})[0-9]{12}|(?:2131|1800|35\d{3})\d{11})\b`, "[CREDIT_CARD]")
		middleware.AddRedactPattern(`\bsk-[A-Za-z0-9]{48}\b`, "[OPENAI_API_KEY]")
		middleware.AddRedactPattern(`\bsk-ant-[A-Za-z0-9]{48}\b`, "[ANTHROPIC_API_KEY]")
	}

	return middleware
}

// AddHandler adds a log handler for a specific level
func (m *LoggingMiddleware) AddHandler(level LogLevel, handler LogHandler) {
	if m.handlers[level] == nil {
		m.handlers[level] = make([]LogHandler, 0)
	}
	m.handlers[level] = append(m.handlers[level], handler)
}

// RemoveHandlers removes all handlers for a specific level
func (m *LoggingMiddleware) RemoveHandlers(level LogLevel) {
	delete(m.handlers, level)
}

// SetMinLevel sets the minimum log level
func (m *LoggingMiddleware) SetMinLevel(level LogLevel) {
	m.minLevel = level
}

// GetMinLevel returns the minimum log level
func (m *LoggingMiddleware) GetMinLevel() LogLevel {
	return m.minLevel
}

// SetRedactPII sets whether to redact PII
func (m *LoggingMiddleware) SetRedactPII(redact bool) {
	m.redactPII = redact
}

// IsRedactingPII returns whether PII is being redacted
func (m *LoggingMiddleware) IsRedactingPII() bool {
	return m.redactPII
}

// AddRedactPattern adds a pattern to redact
func (m *LoggingMiddleware) AddRedactPattern(pattern string, replacement string) error {
	// Compile the pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile redact pattern: %w", err)
	}

	// Add the pattern to the list
	m.redactPatterns = append(m.redactPatterns, re)

	return nil
}

// ClearRedactPatterns clears all redact patterns
func (m *LoggingMiddleware) ClearRedactPatterns() {
	m.redactPatterns = nil
}

// Log logs an entry
func (m *LoggingMiddleware) Log(entry *LogEntry) {
	// Check if the level is high enough
	if entry.Level < m.minLevel {
		return
	}

	// Redact PII if enabled
	if m.redactPII {
		m.redactEntry(entry)
	}

	// Call handlers for the specific level
	for _, handler := range m.handlers[entry.Level] {
		handler(entry)
	}

	// Call handlers for all levels
	for _, handler := range m.handlers[LogLevelDebug-1] {
		handler(entry)
	}
}

// redactEntry redacts PII from a log entry
func (m *LoggingMiddleware) redactEntry(entry *LogEntry) {
	// Convert request and response to JSON
	requestJSON, err := json.Marshal(entry.Request)
	if err == nil {
		requestStr := string(requestJSON)
		for _, pattern := range m.redactPatterns {
			requestStr = pattern.ReplaceAllString(requestStr, "[REDACTED]")
		}
		var redactedRequest interface{}
		if err := json.Unmarshal([]byte(requestStr), &redactedRequest); err == nil {
			entry.Request = redactedRequest
		}
	}

	responseJSON, err := json.Marshal(entry.Response)
	if err == nil {
		responseStr := string(responseJSON)
		for _, pattern := range m.redactPatterns {
			responseStr = pattern.ReplaceAllString(responseStr, "[REDACTED]")
		}
		var redactedResponse interface{}
		if err := json.Unmarshal([]byte(responseStr), &redactedResponse); err == nil {
			entry.Response = redactedResponse
		}
	}

	// Redact error message
	if entry.Error != "" {
		for _, pattern := range m.redactPatterns {
			entry.Error = pattern.ReplaceAllString(entry.Error, "[REDACTED]")
		}
	}

	// Redact additional info
	for key, value := range entry.AdditionalInfo {
		if strValue, ok := value.(string); ok {
			for _, pattern := range m.redactPatterns {
				entry.AdditionalInfo[key] = pattern.ReplaceAllString(strValue, "[REDACTED]")
			}
		}
	}
}

// LogRequest logs a request
func (m *LoggingMiddleware) LogRequest(ctx context.Context, providerType core.ProviderType, operation string, request interface{}, additionalInfo map[string]interface{}) string {
	// Generate a request ID
	requestID := generateRequestID()

	// Create a log entry
	entry := &LogEntry{
		Timestamp:     time.Now(),
		Level:         LogLevelInfo,
		ProviderType:  providerType,
		Operation:     operation,
		RequestID:     requestID,
		Request:       request,
		AdditionalInfo: additionalInfo,
	}

	// Log the entry
	m.Log(entry)

	return requestID
}

// LogResponse logs a response
func (m *LoggingMiddleware) LogResponse(ctx context.Context, providerType core.ProviderType, operation string, requestID string, request interface{}, response interface{}, err error, duration time.Duration, additionalInfo map[string]interface{}) {
	// Determine log level based on error
	level := LogLevelInfo
	if err != nil {
		level = LogLevelError
	}

	// Create a log entry
	entry := &LogEntry{
		Timestamp:     time.Now(),
		Level:         level,
		ProviderType:  providerType,
		Operation:     operation,
		RequestID:     requestID,
		Request:       request,
		Response:      response,
		Duration:      duration,
		AdditionalInfo: additionalInfo,
	}

	// Add error if present
	if err != nil {
		entry.Error = err.Error()
	}

	// Log the entry
	m.Log(entry)
}

// LogMiddleware is middleware that logs requests and responses
func (m *LoggingMiddleware) LogMiddleware(ctx context.Context, providerType core.ProviderType, operation string, request interface{}, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// Log the request
	requestID := m.LogRequest(ctx, providerType, operation, request, nil)

	// Record start time
	startTime := time.Now()

	// Execute the function
	response, err := fn(ctx)

	// Calculate duration
	duration := time.Since(startTime)

	// Log the response
	m.LogResponse(ctx, providerType, operation, requestID, request, response, err, duration, nil)

	return response, err
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ConsoleLogHandler returns a log handler that logs to the console
func ConsoleLogHandler() LogHandler {
	return func(entry *LogEntry) {
		// Format the log entry
		var level string
		switch entry.Level {
		case LogLevelDebug:
			level = "DEBUG"
		case LogLevelInfo:
			level = "INFO"
		case LogLevelWarning:
			level = "WARNING"
		case LogLevelError:
			level = "ERROR"
		}

		// Format the log message
		var message strings.Builder
		message.WriteString(fmt.Sprintf("[%s] [%s] [%s] [%s] [%s]", entry.Timestamp.Format(time.RFC3339), level, entry.ProviderType, entry.Operation, entry.RequestID))

		if entry.Duration > 0 {
			message.WriteString(fmt.Sprintf(" [%s]", entry.Duration))
		}

		if entry.Error != "" {
			message.WriteString(fmt.Sprintf(" Error: %s", entry.Error))
		}

		// Print the log message
		fmt.Println(message.String())

		// Print request and response in debug mode
		if entry.Level == LogLevelDebug {
			if entry.Request != nil {
				requestJSON, _ := json.MarshalIndent(entry.Request, "", "  ")
				fmt.Printf("Request: %s\n", requestJSON)
			}

			if entry.Response != nil {
				responseJSON, _ := json.MarshalIndent(entry.Response, "", "  ")
				fmt.Printf("Response: %s\n", responseJSON)
			}
		}
	}
}

// FileLogHandler returns a log handler that logs to a file
func FileLogHandler(filePath string) (LogHandler, error) {
	// Open the file for appending
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return func(entry *LogEntry) {
		// Marshal the entry to JSON
		entryJSON, err := json.Marshal(entry)
		if err != nil {
			fmt.Printf("Failed to marshal log entry: %v\n", err)
			return
		}

		// Write the entry to the file
		if _, err := file.Write(append(entryJSON, '\n')); err != nil {
			fmt.Printf("Failed to write log entry: %v\n", err)
		}
	}, nil
}

// JSONLogHandler returns a log handler that logs to a JSON file
func JSONLogHandler(filePath string) (LogHandler, error) {
	// Open the file for appending
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return func(entry *LogEntry) {
		// Marshal the entry to JSON
		entryJSON, err := json.Marshal(entry)
		if err != nil {
			fmt.Printf("Failed to marshal log entry: %v\n", err)
			return
		}

		// Write the entry to the file
		if _, err := file.Write(append(entryJSON, '\n')); err != nil {
			fmt.Printf("Failed to write log entry: %v\n", err)
		}
	}, nil
}
