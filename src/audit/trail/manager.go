// Package trail provides a comprehensive audit trail and logging system
package trail

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

// Common errors
var (
	ErrLoggerNotFound     = errors.New("logger not found")
	ErrInvalidLogFormat   = errors.New("invalid log format")
	ErrInvalidLogLevel    = errors.New("invalid log level")
	ErrStorageFailure     = errors.New("storage failure")
	ErrInvalidTimeRange   = errors.New("invalid time range")
	ErrSignatureGeneration = errors.New("failed to generate signature")
	ErrSignatureVerification = errors.New("failed to verify signature")
)

// AuditTrailManager is responsible for managing the audit trail and logging system
type AuditTrailManager struct {
	config    *AuditConfig
	loggers   []AuditLogger
	signingKey []byte
	mu        sync.RWMutex

// AuditConfig defines the configuration for the audit trail system
type AuditConfig struct {
	// Minimum log level to record
	MinLogLevel LogLevel `json:"min_log_level"`
	
	// Whether to enable tamper-evident logging
	TamperEvident bool `json:"tamper_evident"`
	
	// Secret key for HMAC signatures (if tamper-evident is enabled)
	SigningKey string `json:"signing_key,omitempty"`
	
	// Whether to redact sensitive information
	RedactSensitiveInfo bool `json:"redact_sensitive_info"`
	
	// Fields to redact if redaction is enabled
	RedactFields []string `json:"redact_fields,omitempty"`
	
	// Maximum log retention period in days
	RetentionDays int `json:"retention_days"`
	
	// Whether to include stack traces for errors
	IncludeStackTraces bool `json:"include_stack_traces"`
	
	// Whether to compress logs for storage
	CompressLogs bool `json:"compress_logs"`
	
	// Whether to encrypt logs for storage
	EncryptLogs bool `json:"encrypt_logs"`
	
	// Encryption key ID (if encryption is enabled)
	EncryptionKeyID string `json:"encryption_key_id,omitempty"`

// LogQuery defines parameters for querying audit logs
type LogQuery struct {
	// Start time for the query range
	StartTime time.Time `json:"start_time,omitempty"`
	
	// End time for the query range
	EndTime time.Time `json:"end_time,omitempty"`
	
	// Minimum log level to include
	MinLevel LogLevel `json:"min_level,omitempty"`
	
	// Operation types to include
	Operations []OperationType `json:"operations,omitempty"`
	
	// Components to include
	Components []string `json:"components,omitempty"`
	
	// Users to include
	Users []string `json:"users,omitempty"`
	
	// Resources to include
	Resources []string `json:"resources,omitempty"`
	
	// Status values to include
	Statuses []string `json:"statuses,omitempty"`
	
	// Tags to filter by (must match all)
	Tags []string `json:"tags,omitempty"`
	
	// Full-text search query
	Query string `json:"query,omitempty"`
	
	// Pagination limit
	Limit int `json:"limit,omitempty"`
	
	// Pagination offset
	Offset int `json:"offset,omitempty"`
	
	// Sort field
	SortBy string `json:"sort_by,omitempty"`
	
	// Sort direction (asc or desc)
	SortDirection string `json:"sort_direction,omitempty"`

// LogQueryResult contains the results of a log query
type LogQueryResult struct {
	// Logs matching the query
	Logs []*AuditLog `json:"logs"`
	
	// Total number of logs matching the query (before pagination)
	TotalCount int `json:"total_count"`
	
	// Whether there are more logs available
	HasMore bool `json:"has_more"`

// ExportFormat defines the format for exporting audit logs
type ExportFormat string

const (
	// FormatJSON exports logs in JSON format
	FormatJSON ExportFormat = "json"
	
	// FormatCSV exports logs in CSV format
	FormatCSV ExportFormat = "csv"
	
	// FormatPDF exports logs in PDF format
	FormatPDF ExportFormat = "pdf"
)

// NewAuditTrailManager creates a new audit trail manager
func NewAuditTrailManager(config *AuditConfig) (*AuditTrailManager, error) {
	if config == nil {
		config = DefaultAuditConfig()
	}
	
	manager := &AuditTrailManager{
		config:  config,
		loggers: make([]AuditLogger, 0),
	}
	
	// Set up signing key if tamper-evident logging is enabled
	if config.TamperEvident && config.SigningKey != "" {
		manager.signingKey = []byte(config.SigningKey)
	}
	
	return manager, nil

// AddLogger adds a logger to the manager
func (m *AuditTrailManager) AddLogger(logger AuditLogger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.loggers = append(m.loggers, logger)

// RemoveLogger removes a logger from the manager
func (m *AuditTrailManager) RemoveLogger(loggerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i, logger := range m.loggers {
		if logger.GetID() == loggerID {
			// Remove the logger
			m.loggers = append(m.loggers[:i], m.loggers[i+1:]...)
			return nil
		}
	}
	
	return ErrLoggerNotFound

// Log records an audit log entry
func (m *AuditTrailManager) Log(ctx context.Context, log *AuditLog) error {
	// Check if the log level meets the minimum threshold
	if !m.isLevelEnabled(log.Level) {
		return nil
	}
	
	// Apply redaction if enabled
	if m.config.RedactSensitiveInfo {
		m.redactSensitiveInfo(log)
	}
	
	// Generate tamper-evident signature if enabled
	if m.config.TamperEvident && len(m.signingKey) > 0 {
		signature, err := m.generateSignature(log)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrSignatureGeneration, err)
		}
		log.Signature = signature
	}
	
	// Log to all registered loggers
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var lastErr error
	for _, logger := range m.loggers {
		if err := logger.Log(ctx, log); err != nil {
			lastErr = err
			// Continue logging to other loggers even if one fails
		}
	}
	
	return lastErr

// Query searches for audit logs matching the specified criteria
func (m *AuditTrailManager) Query(ctx context.Context, query *LogQuery) (*LogQueryResult, error) {
	if query == nil {
		query = &LogQuery{}
	}
	
	// Validate time range if specified
	if !query.StartTime.IsZero() && !query.EndTime.IsZero() {
		if query.EndTime.Before(query.StartTime) {
			return nil, ErrInvalidTimeRange
		}
	}
	
	// Query all loggers and merge results
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.loggers) == 0 {
		return &LogQueryResult{
			Logs:       []*AuditLog{},
			TotalCount: 0,
			HasMore:    false,
		}, nil
	}
	
	// Use the first logger that supports querying
	for _, logger := range m.loggers {
		if queryer, ok := logger.(AuditQueryLogger); ok {
			return queryer.Query(ctx, query)
		}
	}
	
	return nil, errors.New("no logger supports querying")

// Export exports audit logs in the specified format
func (m *AuditTrailManager) Export(ctx context.Context, query *LogQuery, format ExportFormat) ([]byte, error) {
	// Query the logs first
	result, err := m.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	
	// Find a logger that supports exporting
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, logger := range m.loggers {
		if exporter, ok := logger.(AuditExporter); ok {
			return exporter.Export(ctx, result.Logs, format)
		}
	}
	
	return nil, errors.New("no logger supports exporting")

// VerifyLogIntegrity verifies the integrity of a log entry
func (m *AuditTrailManager) VerifyLogIntegrity(log *AuditLog) (bool, error) {
	if !m.config.TamperEvident || len(m.signingKey) == 0 {
		return false, errors.New("tamper-evident logging is not enabled")
	}
	
	if log.Signature == "" {
		return false, errors.New("log entry has no signature")
	}
	
	// Save the original signature
	originalSignature := log.Signature
	log.Signature = ""
	
	// Generate a new signature and compare
	newSignature, err := m.generateSignature(log)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrSignatureVerification, err)
	}
	
	// Restore the original signature
	log.Signature = originalSignature
	
	return originalSignature == newSignature, nil

// GetLogLevel returns the current minimum log level
func (m *AuditTrailManager) GetLogLevel() LogLevel {
	return m.config.MinLogLevel

// SetLogLevel sets the minimum log level
func (m *AuditTrailManager) SetLogLevel(level LogLevel) {
	m.config.MinLogLevel = level

// GetConfig returns the current configuration
func (m *AuditTrailManager) GetConfig() *AuditConfig {
	configCopy := *m.config
	return &configCopy

// Close closes all loggers and releases resources
func (m *AuditTrailManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var lastErr error
	for _, logger := range m.loggers {
		if err := logger.Close(); err != nil {
			lastErr = err
		}
	}
	
	m.loggers = nil
	
	return lastErr

// Helper methods

// isLevelEnabled checks if a log level meets the minimum threshold
func (m *AuditTrailManager) isLevelEnabled(level LogLevel) bool {
	levelValue := logLevelValue(level)
	minLevelValue := logLevelValue(m.config.MinLogLevel)
	return levelValue >= minLevelValue

// logLevelValue returns the numeric value of a log level
func logLevelValue(level LogLevel) int {
	switch level {
	case LogLevelDebug:
		return 0
	case LogLevelInfo:
		return 1
	case LogLevelWarning:
		return 2
	case LogLevelError:
		return 3
	case LogLevelCritical:
		return 4
	default:
		return 1 // Default to info level
	}

// redactSensitiveInfo redacts sensitive information from a log entry
func (m *AuditTrailManager) redactSensitiveInfo(log *AuditLog) {
	if len(m.config.RedactFields) == 0 {
		return
	}
	
	// Redact fields in metadata
	if log.Metadata != nil {
		for _, field := range m.config.RedactFields {
			if _, exists := log.Metadata[field]; exists {
				log.Metadata[field] = "[REDACTED]"
			}
		}
	}
	
	// Redact fields in changes
	if log.Changes != nil {
		for _, field := range m.config.RedactFields {
			// Redact in before state
			if log.Changes.Before != nil {
				if _, exists := log.Changes.Before[field]; exists {
					log.Changes.Before[field] = "[REDACTED]"
				}
			}
			
			// Redact in after state
			if log.Changes.After != nil {
				if _, exists := log.Changes.After[field]; exists {
					log.Changes.After[field] = "[REDACTED]"
				}
			}
		}
	}

// generateSignature generates an HMAC signature for a log entry
func (m *AuditTrailManager) generateSignature(log *AuditLog) (string, error) {
	// Convert log to JSON
	jsonData, err := log.ToJSON()
	if err != nil {
		return "", err
	}
	
	// Create HMAC
	h := hmac.New(sha256.New, m.signingKey)
	h.Write([]byte(jsonData))
	
	// Return hex-encoded signature
	return hex.EncodeToString(h.Sum(nil)), nil

// DefaultAuditConfig returns the default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		MinLogLevel:        LogLevelInfo,
		TamperEvident:      true,
		RedactSensitiveInfo: true,
		RedactFields:       []string{"password", "token", "key", "secret", "credential"},
		RetentionDays:      365, // 1 year
		IncludeStackTraces: true,
		CompressLogs:       true,
	}
