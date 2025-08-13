// Package repository provides functionality for interacting with bundle repositories
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/audit"
)

// RepositoryAuditLogger is a wrapper around the audit.AuditLogger that provides
// repository-specific audit logging functionality
type RepositoryAuditLogger struct {
	// AuditLogger is the underlying audit logger
	AuditLogger *audit.AuditLogger
	// RepositoryType is the type of repository (HTTP, File, etc.)
	RepositoryType string
	// RepositoryURL is the URL of the repository
	RepositoryURL string
}

// NewRepositoryAuditLogger creates a new repository audit logger
func NewRepositoryAuditLogger(auditLogger *audit.AuditLogger, repositoryType, repositoryURL string) *RepositoryAuditLogger {
	if auditLogger == nil {
		// Create a default audit logger if none is provided
		auditLogger = audit.NewAuditLogger(nil, "system")
		auditLogger.StoreEvents = true
	}

	return &RepositoryAuditLogger{
		AuditLogger:     auditLogger,
		RepositoryType:  repositoryType,
		RepositoryURL:   repositoryURL,
	}
}

// LogRepositoryConnect logs a repository connection event
func (l *RepositoryAuditLogger) LogRepositoryConnect(ctx context.Context, repositoryID string) {
	l.AuditLogger.LogEventWithStatus("repository_connect", "Repository", repositoryID, "in-progress", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "connect",
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryConnectSuccess logs a successful repository connection event
func (l *RepositoryAuditLogger) LogRepositoryConnectSuccess(ctx context.Context, repositoryID string) {
	l.AuditLogger.LogEventWithStatus("repository_connect", "Repository", repositoryID, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "connect",
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryConnectFailure logs a failed repository connection event
func (l *RepositoryAuditLogger) LogRepositoryConnectFailure(ctx context.Context, repositoryID string, err error) {
	l.AuditLogger.LogEventWithStatus("repository_connect", "Repository", repositoryID, "failure", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "connect",
		"error":           err.Error(),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryDisconnect logs a repository disconnection event
func (l *RepositoryAuditLogger) LogRepositoryDisconnect(ctx context.Context, repositoryID string) {
	l.AuditLogger.LogEventWithStatus("repository_disconnect", "Repository", repositoryID, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "disconnect",
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogFileList logs a file listing event
func (l *RepositoryAuditLogger) LogFileList(ctx context.Context, repositoryID, pattern string, count int) {
	l.AuditLogger.LogEventWithStatus("repository_list_files", "Repository", repositoryID, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "list_files",
		"pattern":         pattern,
		"file_count":      count,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryListFiles logs a file listing operation in a repository
func (l *RepositoryAuditLogger) LogRepositoryListFiles(ctx context.Context, repoURL, pattern string) {
	if l.AuditLogger == nil {
		return
	}
	
	details := map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  repoURL,
		"operation":       "list_files",
		"timestamp":       time.Now().Format(time.RFC3339),
	}
	
	// Only add pattern if it's not empty
	if pattern != "" {
		details["pattern"] = pattern
	}
	
	l.AuditLogger.LogEventWithStatus("repository_list_files", "Repository", l.RepositoryURL, "in-progress", details)
}

// LogFileListFailure logs a failed file listing event
func (l *RepositoryAuditLogger) LogFileListFailure(ctx context.Context, repositoryID, pattern string, err error) {
	l.AuditLogger.LogEventWithStatus("repository_list_files", "Repository", repositoryID, "failure", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "list_files",
		"pattern":         pattern,
		"error":           err.Error(),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogFileDownloadStart logs the start of a file download
func (l *RepositoryAuditLogger) LogFileDownloadStart(ctx context.Context, repositoryID, filePath string) {
	l.AuditLogger.LogEventWithStatus("repository_download", "Repository", repositoryID, "in-progress", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "download",
		"file_path":       filePath,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogFileDownloadSuccess logs a successful file download
func (l *RepositoryAuditLogger) LogFileDownloadSuccess(ctx context.Context, repositoryID, filePath string, sizeBytes int64) {
	l.AuditLogger.LogEventWithStatus("repository_download", "Repository", repositoryID, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "download",
		"file_path":       filePath,
		"size_bytes":      sizeBytes,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogFileDownloadFailure logs a failed file download
func (l *RepositoryAuditLogger) LogFileDownloadFailure(ctx context.Context, repositoryID, filePath string, err error) {
	l.AuditLogger.LogEventWithStatus("repository_download", "Repository", repositoryID, "failure", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "download",
		"file_path":       filePath,
		"error":           err.Error(),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogFileExists logs a file existence check
func (l *RepositoryAuditLogger) LogFileExists(ctx context.Context, repositoryID, filePath string, exists bool) {
	l.AuditLogger.LogEventWithStatus("repository_file_exists", "Repository", repositoryID, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "file_exists",
		"file_path":       filePath,
		"exists":          exists,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogFileExistsFailure logs a failed file existence check
func (l *RepositoryAuditLogger) LogFileExistsFailure(ctx context.Context, repositoryID, filePath string, err error) {
	l.AuditLogger.LogEventWithStatus("repository_file_exists", "Repository", repositoryID, "failure", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "file_exists",
		"file_path":       filePath,
		"error":           err.Error(),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogGetLastModified logs a last modified time check
func (l *RepositoryAuditLogger) LogGetLastModified(ctx context.Context, repositoryID, filePath string, modTime time.Time) {
	l.AuditLogger.LogEventWithStatus("repository_get_last_modified", "Repository", repositoryID, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "get_last_modified",
		"file_path":       filePath,
		"modified_time":   modTime.Format(time.RFC3339),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogGetLastModifiedFailure logs a failed last modified time check
func (l *RepositoryAuditLogger) LogGetLastModifiedFailure(ctx context.Context, repositoryID, filePath string, err error) {
	l.AuditLogger.LogEventWithStatus("repository_get_last_modified", "Repository", repositoryID, "failure", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  l.RepositoryURL,
		"operation":       "get_last_modified",
		"file_path":       filePath,
		"error":           err.Error(),
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// GenerateComplianceReport generates a compliance report for repository operations
func (l *RepositoryAuditLogger) GenerateComplianceReport(ctx context.Context, repositoryID string, startTime, endTime time.Time) (string, error) {
	// Create a buffer to store the report
	var reportBuffer struct {
		RepositoryID   string    `json:"repository_id"`
		RepositoryType string    `json:"repository_type"`
		RepositoryURL  string    `json:"repository_url"`
		StartTime      time.Time `json:"start_time"`
		EndTime        time.Time `json:"end_time"`
		GeneratedAt    time.Time `json:"generated_at"`
	}

	// Set report metadata
	reportBuffer.RepositoryID = repositoryID
	reportBuffer.RepositoryType = l.RepositoryType
	reportBuffer.RepositoryURL = l.RepositoryURL
	reportBuffer.StartTime = startTime
	reportBuffer.EndTime = endTime
	reportBuffer.GeneratedAt = time.Now()

	// Create filter options for the report
	filterOptions := audit.FilterOptions{
		StartTime:  &startTime,
		EndTime:    &endTime,
		BundleIDs:  []string{repositoryID},
		EventTypes: []string{"repository_connect", "repository_disconnect", "repository_list_files", "repository_download", "repository_file_exists", "repository_get_last_modified"},
	}

	// Generate the report using the enhanced audit logger
	reportOptions := audit.ComplianceReportOptions{
		ReportType: audit.ActivityReport,
		Filter:     filterOptions,
		Format:     audit.JSONAuditFormat,
	}

	// Create a string builder to store the report
	var reportBuilder strings.Builder

	// Generate the report
	err := l.AuditLogger.GenerateComplianceReport(&reportBuilder, reportOptions)
	if err != nil {
		return "", fmt.Errorf("failed to generate compliance report: %w", err)
	}

	return reportBuilder.String(), nil
}

// LogRepositoryGetFile logs a file retrieval operation
func (l *RepositoryAuditLogger) LogRepositoryGetFile(ctx context.Context, repoURL, path string) {
	if l.AuditLogger == nil {
		return
	}
	
	l.AuditLogger.LogEventWithStatus("repository_get_file", "Repository", repoURL, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  repoURL,
		"operation":       "get_file",
		"file_path":       path,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryFileExists logs a file existence check operation
func (l *RepositoryAuditLogger) LogRepositoryFileExists(ctx context.Context, repoURL, path string) {
	if l.AuditLogger == nil {
		return
	}
	
	l.AuditLogger.LogEventWithStatus("repository_file_exists", "Repository", repoURL, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  repoURL,
		"operation":       "file_exists",
		"file_path":       path,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryGetLastModified logs a last modified time retrieval operation
func (l *RepositoryAuditLogger) LogRepositoryGetLastModified(ctx context.Context, repoURL, path string) {
	if l.AuditLogger == nil {
		return
	}
	
	l.AuditLogger.LogEventWithStatus("repository_get_last_modified", "Repository", repoURL, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  repoURL,
		"operation":       "get_last_modified",
		"file_path":       path,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryStoreFile logs a file storage operation
func (l *RepositoryAuditLogger) LogRepositoryStoreFile(ctx context.Context, repoURL, path string) {
	if l.AuditLogger == nil {
		return
	}
	
	l.AuditLogger.LogEventWithStatus("repository_store_file", "Repository", repoURL, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  repoURL,
		"operation":       "store_file",
		"file_path":       path,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// LogRepositoryDeleteFile logs a file deletion operation
func (l *RepositoryAuditLogger) LogRepositoryDeleteFile(ctx context.Context, repoURL, path string) {
	if l.AuditLogger == nil {
		return
	}
	
	l.AuditLogger.LogEventWithStatus("repository_delete_file", "Repository", repoURL, "success", map[string]interface{}{
		"repository_type": l.RepositoryType,
		"repository_url":  repoURL,
		"operation":       "delete_file",
		"file_path":       path,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}
