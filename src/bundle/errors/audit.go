// Package errors provides error handling functionality for bundle operations
package errors


import (
	"fmt"
	"io"
	"time"
)

// ErrorLogger defines the interface for logging errors
type ErrorLogger interface {
	// LogEvent logs an event
	LogEvent(event, component, id string, details map[string]interface{})
	// LogEventWithStatus logs an event with a status
	LogEventWithStatus(event, component, id, status string, details map[string]interface{})
}

// AuditLogger implements the ErrorLogger interface for audit logging
type AuditLogger struct {
	// Writer is the writer for audit logs
	Writer io.Writer
	// User is the user performing the operation
	User string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(writer io.Writer, user string) *AuditLogger {
	return &AuditLogger{
		Writer: writer,
		User:   user,
	}
}

// LogEvent logs an event
func (l *AuditLogger) LogEvent(event, component, id string, details map[string]interface{}) {
	l.LogEventWithStatus(event, component, id, "info", details)
}

// LogEventWithStatus logs an event with a status
func (l *AuditLogger) LogEventWithStatus(event, component, id, status string, details map[string]interface{}) {
	if l.Writer == nil {
		return
	}
	
	timestamp := time.Now().Format(time.RFC3339)
	user := l.User
	if user == "" {
		user = "system"
	}
	
	// Format the details as a string
	detailsStr := ""
	for k, v := range details {
		detailsStr += fmt.Sprintf(" %s=%v", k, v)
	}
	
	// Write the audit log entry
	fmt.Fprintf(l.Writer, "[%s] [%s] [%s] [%s] [%s] [%s]%s\n",
		timestamp, status, user, component, event, id, detailsStr)
}

// LogImportStart logs the start of an import operation
func (l *AuditLogger) LogImportStart(bundleID, bundlePath string, options map[string]interface{}) {
	l.LogEvent("import_started", "BundleImporter", bundleID, map[string]interface{}{
		"bundle_path": bundlePath,
		"options":     options,
	})
}

// LogImportComplete logs the completion of an import operation
func (l *AuditLogger) LogImportComplete(bundleID string, success bool, details map[string]interface{}) {
	status := "success"
	if !success {
		status = "failure"
	}
	l.LogEventWithStatus("import_completed", "BundleImporter", bundleID, status, details)
}

// LogValidation logs a validation event
func (l *AuditLogger) LogValidation(bundleID, bundlePath string, level string, success bool, details map[string]interface{}) {
	status := "success"
	if !success {
		status = "failure"
	}
	details["validation_level"] = level
	details["bundle_path"] = bundlePath
	l.LogEventWithStatus("validation", "BundleImporter", bundleID, status, details)
}

// LogBackupCreated logs a backup creation event
func (l *AuditLogger) LogBackupCreated(bundleID, targetDir, backupPath string) {
	l.LogEvent("backup_created", "BundleImporter", bundleID, map[string]interface{}{
		"target_dir":  targetDir,
		"backup_path": backupPath,
	})
}

// LogFileInstallation logs a file installation event
func (l *AuditLogger) LogFileInstallation(bundleID, filePath string, success bool, details map[string]interface{}) {
	status := "success"
	if !success {
		status = "failure"
	}
	details["file_path"] = filePath
	l.LogEventWithStatus("file_installation", "BundleImporter", bundleID, status, details)
}

// LogImportSummary logs a summary of the import operation
func (l *AuditLogger) LogImportSummary(bundleID string, stats map[string]interface{}) {
	l.LogEvent("import_summary", "BundleImporter", bundleID, stats)
}

// LogConflict logs a conflict resolution event
func (l *AuditLogger) LogConflict(bundleID string, conflict interface{}, strategy interface{}) {
	l.LogEvent("conflict_resolved", "ConflictResolver", bundleID, map[string]interface{}{
		"conflict": conflict,
		"strategy": strategy,
	})
}
