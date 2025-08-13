// Package trail provides a comprehensive audit trail system for tracking all operations
package trail

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/perplext/LLMrecon/src/security/access/audit"
	"github.com/perplext/LLMrecon/src/security/access/common"
)

// AuditLog represents a comprehensive audit log entry
type AuditLog struct {
	// Unique identifier for the audit log
	ID string `json:"id"`
	
	// Time when the event occurred
	Timestamp time.Time `json:"timestamp"`
	
	// User who performed the action
	UserID string `json:"user_id,omitempty"`
	
	// Username of the user who performed the action
	Username string `json:"username,omitempty"`
	
	// Type of operation performed
	Operation string `json:"operation"`
	
	// Resource type that was acted upon
	ResourceType string `json:"resource_type"`
	
	// Specific resource identifier
	ResourceID string `json:"resource_id,omitempty"`
	
	// Human-readable description of the event
	Description string `json:"description"`
	
	// IP address where the action originated
	IPAddress string `json:"ip_address,omitempty"`
	
	// Status of the operation (success, failure, etc.)
	Status string `json:"status"`
	
	// Additional details about the operation
	Details map[string]interface{} `json:"details,omitempty"`
	
	// Previous state of the resource (for update operations)
	PreviousState map[string]interface{} `json:"previous_state,omitempty"`
	
	// New state of the resource (for update operations)
	NewState map[string]interface{} `json:"new_state,omitempty"`
	
	// Changes made during the operation
	Changes map[string]interface{} `json:"changes,omitempty"`
	
	// Verification information (for verification operations)
	Verification *VerificationInfo `json:"verification,omitempty"`
	
	// Compliance metadata
	Compliance *ComplianceInfo `json:"compliance,omitempty"`
	
	// Digital signature for tamper evidence
	Signature string `json:"signature,omitempty"`
	
	// Hash of the previous log entry (for chain of custody)
	PreviousHash string `json:"previous_hash,omitempty"`
}

// VerificationInfo contains information about verification operations
type VerificationInfo struct {
	// Type of verification performed
	VerificationType string `json:"verification_type"`
	
	// Entity that performed the verification
	VerifiedBy string `json:"verified_by"`
	
	// Time when the verification was performed
	VerifiedAt time.Time `json:"verified_at"`
	
	// Result of the verification
	Result string `json:"result"`
	
	// Additional verification metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceInfo contains compliance-related metadata
type ComplianceInfo struct {
	// Compliance frameworks applicable to this operation
	Frameworks []string `json:"frameworks,omitempty"`
	
	// Specific compliance controls addressed
	Controls []string `json:"controls,omitempty"`
	
	// Retention period for this log entry
	RetentionPeriod string `json:"retention_period,omitempty"`
	
	// Classification of the data involved
	DataClassification string `json:"data_classification,omitempty"`
	
	// Whether this operation requires approval
	RequiresApproval bool `json:"requires_approval,omitempty"`
	
	// Approval status if approval is required
	ApprovalStatus string `json:"approval_status,omitempty"`
	
	// Entity that approved the operation
	ApprovedBy string `json:"approved_by,omitempty"`
	
	// Time when the operation was approved
	ApprovedAt time.Time `json:"approved_at,omitempty"`
}

// AuditTrailConfig defines configuration options for the audit trail system
type AuditTrailConfig struct {
	// Whether to enable the audit trail
	Enabled bool `json:"enabled"`
	
	// Directory where audit logs are stored
	LogDirectory string `json:"log_directory"`
	
	// Whether to sign audit logs for tamper evidence
	SignLogs bool `json:"sign_logs"`
	
	// Secret key for signing logs
	SigningKey string `json:"signing_key,omitempty"`
	
	// Whether to maintain a hash chain for logs
	EnableHashChain bool `json:"enable_hash_chain"`
	
	// Whether to compress log files
	CompressLogs bool `json:"compress_logs"`
	
	// Maximum size of a single log file in MB
	MaxLogFileSize int `json:"max_log_file_size"`
	
	// Maximum number of log files to keep
	MaxLogFiles int `json:"max_log_files"`
	
	// Log retention period in days
	RetentionDays int `json:"retention_days"`
	
	// Fields to redact from logs
	RedactFields []string `json:"redact_fields"`
	
	// Whether to include previous and new state in logs
	IncludeState bool `json:"include_state"`
	
	// Whether to include compliance information
	IncludeCompliance bool `json:"include_compliance"`
}

// DefaultAuditTrailConfig returns the default configuration for the audit trail
func DefaultAuditTrailConfig() *AuditTrailConfig {
	return &AuditTrailConfig{
		Enabled:          true,
		LogDirectory:     "audit/trail",
		SignLogs:         true,
		EnableHashChain:  true,
		CompressLogs:     true,
		MaxLogFileSize:   10, // 10 MB
		MaxLogFiles:      100,
		RetentionDays:    365, // 1 year
		RedactFields:     []string{"password", "secret", "token", "key"},
		IncludeState:     true,
		IncludeCompliance: true,
	}
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(operation, resourceType, description string) *AuditLog {
	return &AuditLog{
		ID:          uuid.New().String(),
		Timestamp:   time.Now().UTC(),
		Operation:   operation,
		ResourceType: resourceType,
		Description: description,
		Status:      "success",
		Details:     make(map[string]interface{}),
	}
}

// WithUser adds user information to the audit log
func (l *AuditLog) WithUser(userID, username string) *AuditLog {
	l.UserID = userID
	l.Username = username
	return l
}

// WithResource adds resource information to the audit log
func (l *AuditLog) WithResource(resourceID string) *AuditLog {
	l.ResourceID = resourceID
	return l
}

// WithStatus sets the status of the audit log
func (l *AuditLog) WithStatus(status string) *AuditLog {
	l.Status = status
	return l
}

// WithIP adds IP address information to the audit log
func (l *AuditLog) WithIP(ipAddress string) *AuditLog {
	l.IPAddress = ipAddress
	return l
}

// WithDetail adds a detail to the audit log
func (l *AuditLog) WithDetail(key string, value interface{}) *AuditLog {
	if l.Details == nil {
		l.Details = make(map[string]interface{})
	}
	l.Details[key] = value
	return l
}

// WithChange adds a change to the audit log
func (l *AuditLog) WithChange(field string, oldValue, newValue interface{}) *AuditLog {
	if l.Changes == nil {
		l.Changes = make(map[string]interface{})
	}
	l.Changes[field] = map[string]interface{}{
		"old": oldValue,
		"new": newValue,
	}
	return l
}

// WithChanges adds multiple changes to the audit log
func (l *AuditLog) WithChanges(changes map[string]interface{}) *AuditLog {
	l.Changes = changes
	return l
}

// WithStates adds previous and new state information to the audit log
func (l *AuditLog) WithStates(previousState, newState map[string]interface{}) *AuditLog {
	l.PreviousState = previousState
	l.NewState = newState
	
	// Automatically calculate changes if not already set
	if l.Changes == nil && previousState != nil && newState != nil {
		l.Changes = calculateChanges(previousState, newState)
	}
	
	return l
}

// WithVerification adds verification information to the audit log
func (l *AuditLog) WithVerification(verificationType, verifiedBy string, result string) *AuditLog {
	l.Verification = &VerificationInfo{
		VerificationType: verificationType,
		VerifiedBy:       verifiedBy,
		VerifiedAt:       time.Now().UTC(),
		Result:           result,
		Metadata:         make(map[string]interface{}),
	}
	return l
}

// WithVerificationMetadata adds metadata to the verification information
func (l *AuditLog) WithVerificationMetadata(key string, value interface{}) *AuditLog {
	if l.Verification == nil {
		l.Verification = &VerificationInfo{
			VerifiedAt: time.Now().UTC(),
			Metadata:   make(map[string]interface{}),
		}
	}
	l.Verification.Metadata[key] = value
	return l
}

// WithCompliance adds compliance information to the audit log
func (l *AuditLog) WithCompliance(frameworks, controls []string, dataClassification, retentionPeriod string) *AuditLog {
	l.Compliance = &ComplianceInfo{
		Frameworks:        frameworks,
		Controls:          controls,
		DataClassification: dataClassification,
		RetentionPeriod:   retentionPeriod,
	}
	return l
}

// WithApproval adds approval information to the compliance metadata
func (l *AuditLog) WithApproval(requiresApproval bool, status, approvedBy string) *AuditLog {
	if l.Compliance == nil {
		l.Compliance = &ComplianceInfo{}
	}
	l.Compliance.RequiresApproval = requiresApproval
	l.Compliance.ApprovalStatus = status
	l.Compliance.ApprovedBy = approvedBy
	if status == "approved" {
		l.Compliance.ApprovedAt = time.Now().UTC()
	}
	return l
}

// Sign signs the audit log for tamper evidence
func (l *AuditLog) Sign(key string) error {
	if key == "" {
		return fmt.Errorf("signing key cannot be empty")
	}
	
	// Create a copy of the log without the signature
	logCopy := *l
	logCopy.Signature = ""
	
	// Marshal the log to JSON
	data, err := json.Marshal(logCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal log for signing: %w", err)
	}
	
	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// Set the signature
	l.Signature = signature
	
	return nil
}

// Verify verifies the signature of the audit log
func (l *AuditLog) Verify(key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("signing key cannot be empty")
	}
	
	// Get the existing signature
	existingSignature := l.Signature
	if existingSignature == "" {
		return false, fmt.Errorf("log is not signed")
	}
	
	// Create a copy of the log without the signature
	logCopy := *l
	logCopy.Signature = ""
	
	// Marshal the log to JSON
	data, err := json.Marshal(logCopy)
	if err != nil {
		return false, fmt.Errorf("failed to marshal log for verification: %w", err)
	}
	
	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	calculatedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// Compare signatures
	return hmac.Equal([]byte(existingSignature), []byte(calculatedSignature)), nil
}

// ToJSON converts the audit log to a JSON string
func (l *AuditLog) ToJSON() (string, error) {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal audit log to JSON: %w", err)
	}
	return string(data), nil
}

// ToCSV converts the audit log to a CSV record
func (l *AuditLog) ToCSV() ([]string, error) {
	// Define the CSV fields in order
	record := []string{
		l.ID,
		l.Timestamp.Format(time.RFC3339),
		l.UserID,
		l.Username,
		l.Operation,
		l.ResourceType,
		l.ResourceID,
		l.Description,
		l.IPAddress,
		l.Status,
	}
	
	// Add details as JSON string
	detailsJSON, err := json.Marshal(l.Details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal details to JSON: %w", err)
	}
	record = append(record, string(detailsJSON))
	
	// Add changes as JSON string
	changesJSON, err := json.Marshal(l.Changes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changes to JSON: %w", err)
	}
	record = append(record, string(changesJSON))
	
	return record, nil
}

// CSVHeader returns the header row for CSV export
func CSVHeader() []string {
	return []string{
		"ID",
		"Timestamp",
		"UserID",
		"Username",
		"Operation",
		"ResourceType",
		"ResourceID",
		"Description",
		"IPAddress",
		"Status",
		"Details",
		"Changes",
	}
}

// FromJSON creates an audit log from a JSON string
func FromJSON(jsonStr string) (*AuditLog, error) {
	var log AuditLog
	if err := json.Unmarshal([]byte(jsonStr), &log); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit log from JSON: %w", err)
	}
	return &log, nil
}

// calculateChanges computes the differences between previous and new state
func calculateChanges(previous, new map[string]interface{}) map[string]interface{} {
	changes := make(map[string]interface{})
	
	// Check for changes in existing fields
	for key, oldValue := range previous {
		if newValue, exists := new[key]; exists {
			// Compare values
			if !deepEqual(oldValue, newValue) {
				changes[key] = map[string]interface{}{
					"old": oldValue,
					"new": newValue,
				}
			}
		} else {
			// Field was removed
			changes[key] = map[string]interface{}{
				"old": oldValue,
				"new": nil,
			}
		}
	}
	
	// Check for new fields
	for key, newValue := range new {
		if _, exists := previous[key]; !exists {
			changes[key] = map[string]interface{}{
				"old": nil,
				"new": newValue,
			}
		}
	}
	
	return changes
}

// deepEqual performs a deep comparison of two values
func deepEqual(a, b interface{}) bool {
	// Simple implementation for basic types
	// For production use, consider using reflect.DeepEqual or a more robust solution
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// FilterLogs filters a slice of audit logs based on criteria
func FilterLogs(logs []*AuditLog, filter func(*AuditLog) bool) []*AuditLog {
	if filter == nil {
		return logs
	}
	
	filtered := make([]*AuditLog, 0)
	for _, log := range logs {
		if filter(log) {
			filtered = append(filtered, log)
		}
	}
	
	return filtered
}

// SortLogs sorts a slice of audit logs by timestamp
func SortLogs(logs []*AuditLog, ascending bool) []*AuditLog {
	sorted := make([]*AuditLog, len(logs))
	copy(sorted, logs)
	
	if ascending {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Timestamp.Before(sorted[j].Timestamp)
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[j].Timestamp.Before(sorted[i].Timestamp)
		})
	}
	
	return sorted
}

// ExportLogsToJSON exports audit logs to a JSON file
func ExportLogsToJSON(logs []*AuditLog, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Open file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// Write logs to file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(logs); err != nil {
		return fmt.Errorf("failed to encode logs to JSON: %w", err)
	}
	
	return nil
}

// ExportLogsToCSV exports audit logs to a CSV file
func ExportLogsToCSV(logs []*AuditLog, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Open file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header
	if err := writer.Write(CSVHeader()); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	
	// Write logs
	for _, log := range logs {
		record, err := log.ToCSV()
		if err != nil {
			return fmt.Errorf("failed to convert log to CSV: %w", err)
		}
		
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}
	
	return nil
}

// RedactSensitiveData redacts sensitive data from an audit log
func RedactSensitiveData(log *AuditLog, fields []string) *AuditLog {
	// Create a copy of the log
	logCopy := *log
	
	// Redact fields in details
	if logCopy.Details != nil {
		for _, field := range fields {
			if _, exists := logCopy.Details[field]; exists {
				logCopy.Details[field] = "REDACTED"
			}
		}
	}
	
	// Redact fields in changes
	if logCopy.Changes != nil {
		for _, field := range fields {
			if change, exists := logCopy.Changes[field]; exists {
				if changeMap, ok := change.(map[string]interface{}); ok {
					changeMap["old"] = "REDACTED"
					changeMap["new"] = "REDACTED"
					logCopy.Changes[field] = changeMap
				}
			}
		}
	}
	
	// Redact fields in previous state
	if logCopy.PreviousState != nil {
		for _, field := range fields {
			if _, exists := logCopy.PreviousState[field]; exists {
				logCopy.PreviousState[field] = "REDACTED"
			}
		}
	}
	
	// Redact fields in new state
	if logCopy.NewState != nil {
		for _, field := range fields {
			if _, exists := logCopy.NewState[field]; exists {
				logCopy.NewState[field] = "REDACTED"
			}
		}
	}
	
	return &logCopy
}

// ConvertAuditEventToAuditLog converts an AuditEvent to an AuditLog
func ConvertAuditEventToAuditLog(event *audit.AuditEvent) *AuditLog {
	log := &AuditLog{
		ID:          event.ID,
		Timestamp:   event.Timestamp,
		UserID:      event.UserID,
		Username:    event.Username,
		Operation:   string(event.Action),
		ResourceType: event.Resource,
		ResourceID:   event.ResourceID,
		Description: event.Description,
		IPAddress:   event.IPAddress,
		Status:      event.Status,
		Details:     event.Metadata,
		Changes:     event.Changes,
	}
	
	return log
}

// ConvertAuditLogToAuditEvent converts an AuditLog to an AuditEvent
func ConvertAuditLogToAuditEvent(log *AuditLog) *audit.AuditEvent {
	event := audit.NewAuditEvent(
		common.AuditAction(log.Operation),
		log.ResourceType,
		log.Description,
	)
	
	event.ID = log.ID
	event.Timestamp = log.Timestamp
	event.UserID = log.UserID
	event.Username = log.Username
	event.ResourceID = log.ResourceID
	event.IPAddress = log.IPAddress
	event.Status = log.Status
	event.Metadata = log.Details
	event.Changes = log.Changes
	
	return event
}
