// Package trail provides a comprehensive audit trail and logging system
package trail

import (
	"encoding/json"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	// LogLevelDebug is used for detailed debugging information
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo is used for general information
	LogLevelInfo LogLevel = "info"
	// LogLevelWarning is used for warning conditions
	LogLevelWarning LogLevel = "warning"
	// LogLevelError is used for error conditions
	LogLevelError LogLevel = "error"
	// LogLevelCritical is used for critical conditions
	LogLevelCritical LogLevel = "critical"
)

// OperationType represents the type of operation being logged
type OperationType string

const (
	// OperationCreate represents a creation operation
	OperationCreate OperationType = "create"
	// OperationRead represents a read operation
	OperationRead OperationType = "read"
	// OperationUpdate represents an update operation
	OperationUpdate OperationType = "update"
	// OperationDelete represents a deletion operation
	OperationDelete OperationType = "delete"
	// OperationVerify represents a verification operation
	OperationVerify OperationType = "verify"
	// OperationAuth represents an authentication operation
	OperationAuth OperationType = "auth"
	// OperationConfig represents a configuration change
	OperationConfig OperationType = "config"
	// OperationExport represents an export operation
	OperationExport OperationType = "export"
	// OperationImport represents an import operation
	OperationImport OperationType = "import"
	// OperationDeploy represents a deployment operation
	OperationDeploy OperationType = "deploy"
	// OperationExecute represents an execution operation
	OperationExecute OperationType = "execute"
	// OperationAccess represents an access control operation
	OperationAccess OperationType = "access"
)

// AuditLog represents a structured audit log entry
type AuditLog struct {
	// ID is a unique identifier for the log entry
	ID string `json:"id"`
	
	// Timestamp is the time when the event occurred (with timezone)
	Timestamp time.Time `json:"timestamp"`
	
	// Level is the severity level of the log entry
	Level LogLevel `json:"level"`
	
	// Operation is the type of operation being performed
	Operation OperationType `json:"operation"`
	
	// Component is the system component affected
	Component string `json:"component"`
	
	// SubComponent is a more specific part of the component
	SubComponent string `json:"sub_component,omitempty"`
	
	// User is the user or process that performed the operation
	User string `json:"user,omitempty"`
	
	// UserID is the unique identifier of the user
	UserID string `json:"user_id,omitempty"`
	
	// SessionID is the session identifier
	SessionID string `json:"session_id,omitempty"`
	
	// RequestID is used to correlate multiple log entries for a single request
	RequestID string `json:"request_id,omitempty"`
	
	// TraceID is used for distributed tracing
	TraceID string `json:"trace_id,omitempty"`
	
	// IPAddress is the IP address where the operation originated
	IPAddress string `json:"ip_address,omitempty"`
	
	// UserAgent is the user agent that performed the operation
	UserAgent string `json:"user_agent,omitempty"`
	
	// Resource is the resource being operated on
	Resource string `json:"resource,omitempty"`
	
	// ResourceID is the identifier of the resource
	ResourceID string `json:"resource_id,omitempty"`
	
	// Action is the specific action being performed
	Action string `json:"action,omitempty"`
	
	// Status is the result status of the operation
	Status string `json:"status"`
	
	// StatusCode is a numeric status code
	StatusCode int `json:"status_code,omitempty"`
	
	// Message is a human-readable description of the event
	Message string `json:"message"`
	
	// ErrorCode is the error code if the operation failed
	ErrorCode string `json:"error_code,omitempty"`
	
	// ErrorMessage is the error message if the operation failed
	ErrorMessage string `json:"error_message,omitempty"`
	
	// Duration is the duration of the operation in milliseconds
	Duration int64 `json:"duration,omitempty"`
	
	// Version contains version information
	Version *VersionInfo `json:"version,omitempty"`
	
	// Changes contains details about what changed
	Changes *ChangeInfo `json:"changes,omitempty"`
	
	// Verification contains verification results
	Verification *VerificationInfo `json:"verification,omitempty"`
	
	// Metadata contains additional contextual information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	
	// Tags for categorizing and filtering logs
	Tags []string `json:"tags,omitempty"`
	
	// Signature is a cryptographic signature for tamper evidence
	Signature string `json:"signature,omitempty"`
}

// VersionInfo contains version information for the affected resource
type VersionInfo struct {
	// Previous is the previous version
	Previous string `json:"previous,omitempty"`
	
	// Current is the current version
	Current string `json:"current,omitempty"`
	
	// ChangeType describes the type of version change
	ChangeType string `json:"change_type,omitempty"`
}

// ChangeInfo contains details about what changed in an update operation
type ChangeInfo struct {
	// Before contains the state before the change
	Before map[string]interface{} `json:"before,omitempty"`
	
	// After contains the state after the change
	After map[string]interface{} `json:"after,omitempty"`
	
	// Fields lists the specific fields that changed
	Fields []string `json:"fields,omitempty"`
	
	// Summary provides a human-readable summary of the changes
	Summary string `json:"summary,omitempty"`
}

// VerificationInfo contains verification results
type VerificationInfo struct {
	// Success indicates if verification was successful
	Success bool `json:"success"`
	
	// Method is the verification method used
	Method string `json:"method,omitempty"`
	
	// Details contains additional verification details
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToJSON converts the audit log to a JSON string
func (l *AuditLog) ToJSON() (string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses a JSON string into an audit log
func FromJSON(data string) (*AuditLog, error) {
	var log AuditLog
	err := json.Unmarshal([]byte(data), &log)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// NewAuditLog creates a new audit log entry with default values
func NewAuditLog(operation OperationType, component string, message string) *AuditLog {
	return &AuditLog{
		ID:        generateID(),
		Timestamp: time.Now().UTC(),
		Level:     LogLevelInfo,
		Operation: operation,
		Component: component,
		Status:    "success",
		Message:   message,
		Metadata:  make(map[string]interface{}),
	}
}

// WithLevel sets the severity level
func (l *AuditLog) WithLevel(level LogLevel) *AuditLog {
	l.Level = level
	return l
}

// WithUser sets the user information
func (l *AuditLog) WithUser(userID, username string) *AuditLog {
	l.UserID = userID
	l.User = username
	return l
}

// WithSession sets the session information
func (l *AuditLog) WithSession(sessionID string) *AuditLog {
	l.SessionID = sessionID
	return l
}

// WithRequest sets the request information
func (l *AuditLog) WithRequest(requestID, traceID string) *AuditLog {
	l.RequestID = requestID
	l.TraceID = traceID
	return l
}

// WithClient sets the client information
func (l *AuditLog) WithClient(ipAddress, userAgent string) *AuditLog {
	l.IPAddress = ipAddress
	l.UserAgent = userAgent
	return l
}

// WithResource sets the resource information
func (l *AuditLog) WithResource(resource, resourceID string) *AuditLog {
	l.Resource = resource
	l.ResourceID = resourceID
	return l
}

// WithAction sets the action information
func (l *AuditLog) WithAction(action string) *AuditLog {
	l.Action = action
	return l
}

// WithStatus sets the status information
func (l *AuditLog) WithStatus(status string, statusCode int) *AuditLog {
	l.Status = status
	l.StatusCode = statusCode
	return l
}

// WithError sets the error information
func (l *AuditLog) WithError(errorCode, errorMessage string) *AuditLog {
	l.Status = "error"
	l.ErrorCode = errorCode
	l.ErrorMessage = errorMessage
	return l
}

// WithDuration sets the operation duration
func (l *AuditLog) WithDuration(durationMs int64) *AuditLog {
	l.Duration = durationMs
	return l
}

// WithVersion sets the version information
func (l *AuditLog) WithVersion(previous, current, changeType string) *AuditLog {
	l.Version = &VersionInfo{
		Previous:   previous,
		Current:    current,
		ChangeType: changeType,
	}
	return l
}

// WithChanges sets the change information
func (l *AuditLog) WithChanges(before, after map[string]interface{}, fields []string, summary string) *AuditLog {
	l.Changes = &ChangeInfo{
		Before:  before,
		After:   after,
		Fields:  fields,
		Summary: summary,
	}
	return l
}

// WithVerification sets the verification information
func (l *AuditLog) WithVerification(success bool, method string, details map[string]interface{}) *AuditLog {
	l.Verification = &VerificationInfo{
		Success: success,
		Method:  method,
		Details: details,
	}
	return l
}

// WithMetadata adds metadata to the audit log
func (l *AuditLog) WithMetadata(key string, value interface{}) *AuditLog {
	if l.Metadata == nil {
		l.Metadata = make(map[string]interface{})
	}
	l.Metadata[key] = value
	return l
}

// WithTags adds tags to the audit log
func (l *AuditLog) WithTags(tags ...string) *AuditLog {
	if l.Tags == nil {
		l.Tags = make([]string, 0)
	}
	l.Tags = append(l.Tags, tags...)
	return l
}

// generateID generates a unique identifier for the log entry
func generateID() string {
	// In a real implementation, this would use a UUID or similar
	return "log-" + time.Now().Format("20060102-150405-999999999")
}
