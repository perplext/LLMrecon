// Package trail provides a comprehensive audit trail system for tracking all operations
package trail

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// AuditLogFilter defines a function type for filtering audit logs
type AuditLogFilter func(*AuditLog) bool

// CreateUserFilter creates a filter for logs by user ID or username
func CreateUserFilter(userID, username string) AuditLogFilter {
	return func(log *AuditLog) bool {
		if userID != "" && log.UserID != userID {
			return false
		}
		if username != "" && log.Username != username {
			return false
		}
		return true
	}
}

// CreateResourceFilter creates a filter for logs by resource type and ID
func CreateResourceFilter(resourceType, resourceID string) AuditLogFilter {
	return func(log *AuditLog) bool {
		if resourceType != "" && log.ResourceType != resourceType {
			return false
		}
		if resourceID != "" && log.ResourceID != resourceID {
			return false
		}
		return true
	}
}

// CreateOperationFilter creates a filter for logs by operation
func CreateOperationFilter(operations ...string) AuditLogFilter {
	return func(log *AuditLog) bool {
		if len(operations) == 0 {
			return true
		}
		for _, op := range operations {
			if log.Operation == op {
				return true
			}
		}
		return false
	}
}

// CreateTimeRangeFilter creates a filter for logs within a time range
func CreateTimeRangeFilter(startTime, endTime time.Time) AuditLogFilter {
	return func(log *AuditLog) bool {
		if !startTime.IsZero() && log.Timestamp.Before(startTime) {
			return false
		}
		if !endTime.IsZero() && log.Timestamp.After(endTime) {
			return false
		}
		return true
	}
}

// CreateStatusFilter creates a filter for logs by status
func CreateStatusFilter(statuses ...string) AuditLogFilter {
	return func(log *AuditLog) bool {
		if len(statuses) == 0 {
			return true
		}
		for _, status := range statuses {
			if log.Status == status {
				return true
			}
		}
		return false
	}
}

// CreateComplianceFilter creates a filter for logs by compliance framework
func CreateComplianceFilter(frameworks ...string) AuditLogFilter {
	return func(log *AuditLog) bool {
		if len(frameworks) == 0 || log.Compliance == nil {
			return len(frameworks) == 0
		}
		
		for _, framework := range frameworks {
			for _, f := range log.Compliance.Frameworks {
				if f == framework {
					return true
				}
			}
		}
		return false
	}
}

// CombineFilters combines multiple filters with AND logic
func CombineFilters(filters ...AuditLogFilter) AuditLogFilter {
	return func(log *AuditLog) bool {
		for _, filter := range filters {
			if !filter(log) {
				return false
			}
		}
		return true
	}
}

// HashLog creates a hash of an audit log for integrity verification
func HashLog(log *AuditLog) (string, error) {
	// Create a copy of the log without the signature and previous hash
	logCopy := *log
	logCopy.Signature = ""
	logCopy.PreviousHash = ""
	
	// Marshal the log to JSON
	data, err := json.Marshal(logCopy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal log for hashing: %w", err)
	}
	
	// Create SHA-256 hash
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

// VerifyLogChain verifies the integrity of a chain of audit logs
func VerifyLogChain(logs []*AuditLog) (bool, []string) {
	if len(logs) == 0 {
		return true, nil
	}
	
	// Sort logs by timestamp (oldest first)
	sortedLogs := make([]*AuditLog, len(logs))
	copy(sortedLogs, logs)
	sort.Slice(sortedLogs, func(i, j int) bool {
		return sortedLogs[i].Timestamp.Before(sortedLogs[j].Timestamp)
	})
	
	issues := make([]string, 0)
	previousHash := ""
	
	for i, log := range sortedLogs {
		// Skip the first log's previous hash check
		if i > 0 {
			if log.PreviousHash == "" {
				issues = append(issues, fmt.Sprintf("Log %s has no previous hash", log.ID))
			} else if log.PreviousHash != previousHash {
				issues = append(issues, fmt.Sprintf("Log %s has invalid previous hash", log.ID))
			}
		}
		
		// Calculate the hash for the current log
		hash, err := HashLog(log)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Failed to hash log %s: %v", log.ID, err))
			continue
		}
		
		previousHash = hash
	}
	
	return len(issues) == 0, issues
}

// FindChanges finds all logs that contain changes to a specific field
func FindChanges(logs []*AuditLog, field string) []*AuditLog {
	result := make([]*AuditLog, 0)
	
	for _, log := range logs {
		if log.Changes != nil {
			if _, exists := log.Changes[field]; exists {
				result = append(result, log)
			}
		}
	}
	
	return result
}

// GetFieldHistory gets the history of changes to a specific field
func GetFieldHistory(logs []*AuditLog, resourceType, resourceID, field string) []map[string]interface{} {
	// Filter logs by resource type and ID
	filtered := FilterLogs(logs, CreateResourceFilter(resourceType, resourceID))
	
	// Sort logs by timestamp (oldest first)
	sorted := SortLogs(filtered, true)
	
	// Extract field changes
	history := make([]map[string]interface{}, 0)
	
	for _, log := range sorted {
		if log.Changes != nil {
			if change, exists := log.Changes[field]; exists {
				if changeMap, ok := change.(map[string]interface{}); ok {
					history = append(history, map[string]interface{}{
						"timestamp": log.Timestamp,
						"user_id":   log.UserID,
						"username":  log.Username,
						"old_value": changeMap["old"],
						"new_value": changeMap["new"],
					})
				}
			}
		}
	}
	
	return history
}

// CreateComplianceReport creates a compliance report for the specified framework
func CreateComplianceReport(logs []*AuditLog, framework string, startTime, endTime time.Time) map[string]interface{} {
	// Filter logs by time range and framework
	timeFilter := CreateTimeRangeFilter(startTime, endTime)
	frameworkFilter := CreateComplianceFilter(framework)
	filtered := FilterLogs(logs, CombineFilters(timeFilter, frameworkFilter))
	
	// Group logs by resource type
	resourceTypes := make(map[string][]*AuditLog)
	for _, log := range filtered {
		resourceTypes[log.ResourceType] = append(resourceTypes[log.ResourceType], log)
	}
	
	// Create report
	report := map[string]interface{}{
		"framework":   framework,
		"start_time":  startTime,
		"end_time":    endTime,
		"total_logs":  len(filtered),
		"resources":   make(map[string]interface{}),
		"operations":  make(map[string]int),
		"users":       make(map[string]int),
		"status":      make(map[string]int),
		"compliance": make(map[string]interface{}),
	}
	
	// Populate resource types
	for resourceType, logs := range resourceTypes {
		report["resources"].(map[string]interface{})[resourceType] = len(logs)
	}
	
	// Count operations
	for _, log := range filtered {
		report["operations"].(map[string]int)[log.Operation]++
		report["users"].(map[string]int)[log.Username]++
		report["status"].(map[string]int)[log.Status]++
		
		// Extract compliance controls
		if log.Compliance != nil && len(log.Compliance.Controls) > 0 {
			for _, control := range log.Compliance.Controls {
				if _, exists := report["compliance"].(map[string]interface{})[control]; !exists {
					report["compliance"].(map[string]interface{})[control] = 0
				}
				report["compliance"].(map[string]interface{})[control] = report["compliance"].(map[string]interface{})[control].(int) + 1
			}
		}
	}
	
	return report
}

// ExportComplianceReport exports a compliance report to a file
func ExportComplianceReport(report map[string]interface{}, filePath string) error {
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
	
	// Write report to file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report to JSON: %w", err)
	}
	
	return nil
}

// GenerateAuditSummary generates a summary of audit logs
func GenerateAuditSummary(logs []*AuditLog) map[string]interface{} {
	summary := map[string]interface{}{
		"total_logs":     len(logs),
		"resource_types": make(map[string]int),
		"operations":     make(map[string]int),
		"users":          make(map[string]int),
		"status":         make(map[string]int),
		"time_range":     make(map[string]time.Time),
	}
	
	if len(logs) == 0 {
		return summary
	}
	
	// Initialize time range with first log
	summary["time_range"].(map[string]time.Time)["start"] = logs[0].Timestamp
	summary["time_range"].(map[string]time.Time)["end"] = logs[0].Timestamp
	
	// Process logs
	for _, log := range logs {
		// Update resource types
		summary["resource_types"].(map[string]int)[log.ResourceType]++
		
		// Update operations
		summary["operations"].(map[string]int)[log.Operation]++
		
		// Update users
		if log.Username != "" {
			summary["users"].(map[string]int)[log.Username]++
		}
		
		// Update status
		summary["status"].(map[string]int)[log.Status]++
		
		// Update time range
		if log.Timestamp.Before(summary["time_range"].(map[string]time.Time)["start"]) {
			summary["time_range"].(map[string]time.Time)["start"] = log.Timestamp
		}
		if log.Timestamp.After(summary["time_range"].(map[string]time.Time)["end"]) {
			summary["time_range"].(map[string]time.Time)["end"] = log.Timestamp
		}
	}
	
	return summary
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	parts := make([]string, 0)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hours", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d minutes", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d seconds", seconds))
	}
	
	return strings.Join(parts, ", ")
}

// GetResourceHistory gets the complete history of a resource
func GetResourceHistory(logs []*AuditLog, resourceType, resourceID string) []*AuditLog {
	// Filter logs by resource type and ID
	filtered := FilterLogs(logs, CreateResourceFilter(resourceType, resourceID))
	
	// Sort logs by timestamp (oldest first)
	return SortLogs(filtered, true)
}

// GetUserActivity gets all activity for a specific user
func GetUserActivity(logs []*AuditLog, userID, username string) []*AuditLog {
	// Filter logs by user ID or username
	filtered := FilterLogs(logs, CreateUserFilter(userID, username))
	
	// Sort logs by timestamp (newest first)
	return SortLogs(filtered, false)
}

// GetRecentActivity gets recent activity logs
func GetRecentActivity(logs []*AuditLog, duration time.Duration) []*AuditLog {
	// Calculate cutoff time
	cutoff := time.Now().UTC().Add(-duration)
	
	// Filter logs by time
	filtered := FilterLogs(logs, func(log *AuditLog) bool {
		return log.Timestamp.After(cutoff)
	})
	
	// Sort logs by timestamp (newest first)
	return SortLogs(filtered, false)
}

// GetFailedOperations gets all failed operations
func GetFailedOperations(logs []*AuditLog) []*AuditLog {
	// Filter logs by status
	return FilterLogs(logs, func(log *AuditLog) bool {
		return log.Status != "success" && log.Status != "approved"
	})
}

// GetVerificationHistory gets all verification operations
func GetVerificationHistory(logs []*AuditLog) []*AuditLog {
	// Filter logs by operation
	return FilterLogs(logs, func(log *AuditLog) bool {
		return log.Operation == "verify" || log.Verification != nil
	})
}

// GetApprovalHistory gets all approval operations
func GetApprovalHistory(logs []*AuditLog) []*AuditLog {
	// Filter logs by operation and compliance info
	return FilterLogs(logs, func(log *AuditLog) bool {
		return log.Operation == "approval" || (log.Compliance != nil && log.Compliance.RequiresApproval)
	})
}
