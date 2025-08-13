// Package trail provides a comprehensive audit trail and logging system
package trail

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/syslog"
	"sort"
	"strings"
	"sync"
)

// AuditLogger defines the interface for basic audit logging
type AuditLogger interface {
	// GetID returns the unique identifier for this logger
	GetID() string
	
	// Log records an audit log entry
	Log(ctx context.Context, log *AuditLog) error
	
	// Close releases any resources used by the logger
	Close() error
}

// AuditQueryLogger extends AuditLogger with query capabilities
type AuditQueryLogger interface {
	AuditLogger
	
	// Query searches for audit logs matching the specified criteria
	Query(ctx context.Context, query *LogQuery) (*LogQueryResult, error)
}

// AuditExporter extends AuditLogger with export capabilities
type AuditExporter interface {
	AuditLogger
	
	// Export exports audit logs in the specified format
	Export(ctx context.Context, logs []*AuditLog, format ExportFormat) ([]byte, error)
}

// FileLogger implements audit logging to files
type FileLogger struct {
	id           string
	directory    string
	currentFile  *os.File
	maxFileSize  int64 // in bytes
	maxFiles     int
	compress     bool
	mu           sync.Mutex
	rotationTime time.Time
}

// NewFileLogger creates a new file-based audit logger
func NewFileLogger(directory string, maxFileSize int64, maxFiles int, compress bool) (*FileLogger, error) {
	// Ensure directory exists
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	logger := &FileLogger{
		id:          "file-logger",
		directory:   directory,
		maxFileSize: maxFileSize,
		maxFiles:    maxFiles,
		compress:    compress,
	}
	
	// Open initial log file
	if err := logger.openLogFile(); err != nil {
		return nil, err
	}
	
	return logger, nil
}

// GetID returns the unique identifier for this logger
func (l *FileLogger) GetID() string {
	return l.id
}

// Log records an audit log entry to a file
func (l *FileLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Convert log to JSON
	data, err := log.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert log to JSON: %w", err)
	}
	
	// Check if we need to rotate the log file
	if l.currentFile != nil {
		info, err := l.currentFile.Stat()
		if err == nil && info.Size() >= l.maxFileSize {
			if err := l.rotateLogFile(); err != nil {
				return fmt.Errorf("failed to rotate log file: %w", err)
			}
		}
	}
	
	// Write the log entry with a newline
	if _, err := l.currentFile.WriteString(data + "\n"); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}
	
	return nil
}

// Query searches for audit logs matching the specified criteria
func (l *FileLogger) Query(ctx context.Context, query *LogQuery) (*LogQueryResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Get all log files
	files, err := l.getLogFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get log files: %w", err)
	}
	
	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
	
	// Process files until we have enough logs or run out of files
	logs := make([]*AuditLog, 0)
	totalCount := 0
	
	for _, fileInfo := range files {
		// Skip processing more files if we have enough logs
		if query.Limit > 0 && len(logs) >= query.Limit {
			break
		}
		
		// Open the file
		filePath := filepath.Join(l.directory, fileInfo.Name())
		file, err := os.Open(filePath)
		if err != nil {
			continue // Skip files we can't open
		}
		
		// Handle compressed files
		var reader io.Reader = file
		if strings.HasSuffix(fileInfo.Name(), ".gz") {
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				file.Close()
				continue // Skip files we can't decompress
			}
			reader = gzReader
			defer gzReader.Close()
		}
		
		// Process each line in the file
		scanner := NewLineScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			
			// Parse the log entry
			logEntry, err := FromJSON(line)
			if err != nil {
				continue // Skip invalid entries
			}
			
			// Check if the log matches the query
			if matchesQuery(logEntry, query) {
				totalCount++
				
				// Add to results if within pagination range
				if (query.Offset == 0 || totalCount > query.Offset) && 
				   (query.Limit == 0 || len(logs) < query.Limit) {
					logs = append(logs, logEntry)
				}
			}
		}
		
		file.Close()
	}
	
	// Sort logs if needed
	if query.SortBy != "" {
		sortLogs(logs, query.SortBy, query.SortDirection)
	}
	
	return &LogQueryResult{
		Logs:       logs,
		TotalCount: totalCount,
		HasMore:    totalCount > query.Offset+len(logs),
	}, nil
}

// Export exports audit logs in the specified format
func (l *FileLogger) Export(ctx context.Context, logs []*AuditLog, format ExportFormat) ([]byte, error) {
	switch format {
	case FormatJSON:
		return exportAsJSON(logs)
	case FormatCSV:
		return exportAsCSV(logs)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// Close closes the file logger
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.currentFile != nil {
		err := l.currentFile.Close()
		l.currentFile = nil
		return err
	}
	
	return nil
}

// Helper methods for FileLogger

// openLogFile opens a new log file
func (l *FileLogger) openLogFile() error {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("audit-%s.log", timestamp)
	filePath := filepath.Join(l.directory, filename)
	
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	l.currentFile = file
	l.rotationTime = time.Now()
	
	return nil
}

// rotateLogFile rotates the current log file
func (l *FileLogger) rotateLogFile() error {
	// Close the current file
	if l.currentFile != nil {
		if err := l.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		l.currentFile = nil
	}
	
	// Compress the old file if needed
	if l.compress {
		oldFilePath := l.currentFile.Name()
		if err := compressFile(oldFilePath); err != nil {
			return fmt.Errorf("failed to compress log file: %w", err)
		}
	}
	
	// Open a new log file
	if err := l.openLogFile(); err != nil {
		return err
	}
	
	// Clean up old files if needed
	if l.maxFiles > 0 {
		if err := l.cleanupOldFiles(); err != nil {
			return fmt.Errorf("failed to clean up old files: %w", err)
		}
	}
	
	return nil
}

// getLogFiles returns all log files in the directory
func (l *FileLogger) getLogFiles() ([]os.FileInfo, error) {
	dir, err := os.Open(l.directory)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	
	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	
	// Filter for log files
	logFiles := make([]os.FileInfo, 0)
	for _, info := range fileInfos {
		if !info.IsDir() && (strings.HasPrefix(info.Name(), "audit-") &&
			(strings.HasSuffix(info.Name(), ".log") || strings.HasSuffix(info.Name(), ".log.gz"))) {
			logFiles = append(logFiles, info)
		}
	}
	
	return logFiles, nil
}

// cleanupOldFiles removes old log files when there are too many
func (l *FileLogger) cleanupOldFiles() error {
	files, err := l.getLogFiles()
	if err != nil {
		return err
	}
	
	// Sort files by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	
	// Remove oldest files if we have too many
	for i := 0; i < len(files)-l.maxFiles; i++ {
		filePath := filepath.Join(l.directory, files[i].Name())
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	
	return nil
}

// SyslogLogger implements audit logging to syslog
type SyslogLogger struct {
	id     string
	writer *syslog.Writer
	mu     sync.Mutex
}

// NewSyslogLogger creates a new syslog-based audit logger
func NewSyslogLogger(facility syslog.Priority, tag string) (*SyslogLogger, error) {
	writer, err := syslog.New(facility, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}
	
	return &SyslogLogger{
		id:     "syslog-logger",
		writer: writer,
	}, nil
}

// GetID returns the unique identifier for this logger
func (l *SyslogLogger) GetID() string {
	return l.id
}

// Log records an audit log entry to syslog
func (l *SyslogLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Convert log to JSON
	data, err := log.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert log to JSON: %w", err)
	}
	
	// Write to syslog with appropriate priority
	var logErr error
	switch log.Level {
	case LogLevelDebug:
		logErr = l.writer.Debug(data)
	case LogLevelInfo:
		logErr = l.writer.Info(data)
	case LogLevelWarning:
		logErr = l.writer.Warning(data)
	case LogLevelError:
		logErr = l.writer.Err(data)
	case LogLevelCritical:
		logErr = l.writer.Crit(data)
	default:
		logErr = l.writer.Info(data)
	}
	
	if logErr != nil {
		return fmt.Errorf("failed to write to syslog: %w", logErr)
	}
	
	return nil
}

// Close closes the syslog logger
func (l *SyslogLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.writer != nil {
		return l.writer.Close()
	}
	
	return nil
}

// InMemoryLogger implements in-memory audit logging
type InMemoryLogger struct {
	id       string
	logs     []*AuditLog
	maxLogs  int
	mu       sync.RWMutex
}

// NewInMemoryLogger creates a new in-memory audit logger
func NewInMemoryLogger(maxLogs int) *InMemoryLogger {
	if maxLogs <= 0 {
		maxLogs = 1000 // Default to 1000 logs
	}
	
	return &InMemoryLogger{
		id:      "memory-logger",
		logs:    make([]*AuditLog, 0, maxLogs),
		maxLogs: maxLogs,
	}
}

// GetID returns the unique identifier for this logger
func (l *InMemoryLogger) GetID() string {
	return l.id
}

// Log records an audit log entry in memory
func (l *InMemoryLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Create a copy of the log to prevent modification
	logCopy := *log
	
	// Add to logs
	l.logs = append(l.logs, &logCopy)
	
	// Trim if we have too many logs
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[len(l.logs)-l.maxLogs:]
	}
	
	return nil
}

// Query searches for audit logs matching the specified criteria
func (l *InMemoryLogger) Query(ctx context.Context, query *LogQuery) (*LogQueryResult, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Filter logs
	filtered := make([]*AuditLog, 0)
	for _, log := range l.logs {
		if matchesQuery(log, query) {
			filtered = append(filtered, log)
		}
	}
	
	totalCount := len(filtered)
	
	// Sort if needed
	if query.SortBy != "" {
		sortLogs(filtered, query.SortBy, query.SortDirection)
	}
	
	// Apply pagination
	if query.Limit > 0 || query.Offset > 0 {
		start := query.Offset
		if start >= len(filtered) {
			start = len(filtered)
		}
		
		end := len(filtered)
		if query.Limit > 0 && start+query.Limit < end {
			end = start + query.Limit
		}
		
		filtered = filtered[start:end]
	}
	
	return &LogQueryResult{
		Logs:       filtered,
		TotalCount: totalCount,
		HasMore:    totalCount > query.Offset+len(filtered),
	}, nil
}

// Export exports audit logs in the specified format
func (l *InMemoryLogger) Export(ctx context.Context, logs []*AuditLog, format ExportFormat) ([]byte, error) {
	switch format {
	case FormatJSON:
		return exportAsJSON(logs)
	case FormatCSV:
		return exportAsCSV(logs)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// Close closes the in-memory logger
func (l *InMemoryLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.logs = nil
	
	return nil
}

// Helper functions

// matchesQuery checks if a log entry matches the query criteria
func matchesQuery(log *AuditLog, query *LogQuery) bool {
	// Check time range
	if !query.StartTime.IsZero() && log.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && log.Timestamp.After(query.EndTime) {
		return false
	}
	
	// Check log level
	if query.MinLevel != "" && logLevelValue(log.Level) < logLevelValue(query.MinLevel) {
		return false
	}
	
	// Check operations
	if len(query.Operations) > 0 {
		found := false
		for _, op := range query.Operations {
			if log.Operation == op {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check components
	if len(query.Components) > 0 {
		found := false
		for _, comp := range query.Components {
			if log.Component == comp {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check users
	if len(query.Users) > 0 {
		found := false
		for _, user := range query.Users {
			if log.User == user || log.UserID == user {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check resources
	if len(query.Resources) > 0 {
		found := false
		for _, res := range query.Resources {
			if log.Resource == res {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check statuses
	if len(query.Statuses) > 0 {
		found := false
		for _, status := range query.Statuses {
			if log.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check tags
	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			found := false
			for _, logTag := range log.Tags {
				if logTag == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	// Check full-text search
	if query.Query != "" {
		// Convert log to JSON for full-text search
		jsonData, err := log.ToJSON()
		if err != nil {
			return false
		}
		
		// Simple case-insensitive search
		if !strings.Contains(strings.ToLower(jsonData), strings.ToLower(query.Query)) {
			return false
		}
	}
	
	return true
}

// sortLogs sorts logs by the specified field and direction
func sortLogs(logs []*AuditLog, sortBy, sortDirection string) {
	// Default to descending order for timestamps
	desc := sortDirection != "asc"
	
	sort.Slice(logs, func(i, j int) bool {
		var result bool
		
		switch sortBy {
		case "timestamp":
			result = logs[i].Timestamp.Before(logs[j].Timestamp)
		case "level":
			result = logLevelValue(logs[i].Level) < logLevelValue(logs[j].Level)
		case "operation":
			result = logs[i].Operation < logs[j].Operation
		case "component":
			result = logs[i].Component < logs[j].Component
		case "user":
			result = logs[i].User < logs[j].User
		case "status":
			result = logs[i].Status < logs[j].Status
		default:
			// Default to timestamp
			result = logs[i].Timestamp.Before(logs[j].Timestamp)
		}
		
		if desc {
			return !result
		}
		return result
	})
}

// exportAsJSON exports logs as JSON
func exportAsJSON(logs []*AuditLog) ([]byte, error) {
	return json.MarshalIndent(logs, "", "  ")
}

// exportAsCSV exports logs as CSV
func exportAsCSV(logs []*AuditLog) ([]byte, error) {
	// Create a buffer to write to
	var buf strings.Builder
	writer := csv.NewWriter(&buf)
	
	// Write header
	header := []string{
		"ID", "Timestamp", "Level", "Operation", "Component", "User", "UserID",
		"Resource", "ResourceID", "Action", "Status", "Message",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}
	
	// Write data
	for _, log := range logs {
		row := []string{
			log.ID,
			log.Timestamp.Format(time.RFC3339),
			string(log.Level),
			string(log.Operation),
			log.Component,
			log.User,
			log.UserID,
			log.Resource,
			log.ResourceID,
			log.Action,
			log.Status,
			log.Message,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	
	writer.Flush()
	return []byte(buf.String()), nil
}

// compressFile compresses a file using gzip
func compressFile(filePath string) error {
	// Open the original file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Create the compressed file
	compressedPath := filePath + ".gz"
	compressedFile, err := os.Create(compressedPath)
	if err != nil {
		return err
	}
	defer compressedFile.Close()
	
	// Create a gzip writer
	gzWriter := gzip.NewWriter(compressedFile)
	defer gzWriter.Close()
	
	// Copy the file content to the gzip writer
	if _, err := io.Copy(gzWriter, file); err != nil {
		return err
	}
	
	// Close the gzip writer to flush all data
	if err := gzWriter.Close(); err != nil {
		return err
	}
	
	// Remove the original file
	return os.Remove(filePath)
}

// LineScanner is a simple line scanner that handles both LF and CRLF
type LineScanner struct {
	reader    io.Reader
	buffer    []byte
	remaining []byte
}

// NewLineScanner creates a new line scanner
func NewLineScanner(reader io.Reader) *LineScanner {
	return &LineScanner{
		reader: reader,
		buffer: make([]byte, 4096),
	}
}

// Scan advances to the next line
func (s *LineScanner) Scan() bool {
	// Check if we have data in the buffer
	if len(s.remaining) == 0 {
		n, err := s.reader.Read(s.buffer)
		if err != nil || n == 0 {
			return false
		}
		s.remaining = s.buffer[:n]
	}
	
	// Find the next newline
	i := 0
	for i < len(s.remaining) {
		if s.remaining[i] == '\n' {
			s.remaining = s.remaining[i+1:]
			return true
		}
		i++
	}
	
	// No newline found, read more
	s.remaining = nil
	return s.Scan()
}

// Text returns the current line
func (s *LineScanner) Text() string {
	i := 0
	for i < len(s.remaining) {
		if s.remaining[i] == '\n' {
			break
		}
		i++
	}
	
	line := string(s.remaining[:i])
	s.remaining = s.remaining[i:]
	
	// Remove carriage return if present
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	
	return line
}
