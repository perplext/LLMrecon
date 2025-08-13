// Package trail provides a comprehensive audit trail system for tracking all operations
package trail

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Logger defines the interface for audit trail loggers
type Logger interface {
	// Log logs an audit log entry
	Log(ctx context.Context, log *AuditLog) error
	
	// Query queries audit logs based on filters
	Query(ctx context.Context, filters map[string]interface{}, startTime, endTime time.Time) ([]*AuditLog, error)
	
	// GetLog retrieves a specific audit log by ID
	GetLog(ctx context.Context, id string) (*AuditLog, error)
}

// LogPurger defines the interface for loggers that support log purging
type LogPurger interface {
	// PurgeLogs purges logs older than the specified time
	PurgeLogs(ctx context.Context, olderThan time.Time) error
}

// InMemoryLogger is an in-memory implementation of Logger
type InMemoryLogger struct {
	logs   []*AuditLog
	config *AuditTrailConfig
	mu     sync.RWMutex
}

// NewInMemoryLogger creates a new in-memory logger
func NewInMemoryLogger(config *AuditTrailConfig) *InMemoryLogger {
	return &InMemoryLogger{
		logs:   make([]*AuditLog, 0),
		config: config,
	}
}

// Log logs an audit log entry to memory
func (l *InMemoryLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Create a copy of the log to prevent modification
	logCopy := *log
	l.logs = append(l.logs, &logCopy)
	
	// Limit the number of logs in memory (default to 1000 if not specified)
	maxLogs := 1000
	if l.config != nil && l.config.MaxLogFiles > 0 {
		maxLogs = l.config.MaxLogFiles * 1000 // Rough estimate
	}
	
	if len(l.logs) > maxLogs {
		l.logs = l.logs[len(l.logs)-maxLogs:]
	}
	
	return nil
}

// Query queries audit logs from memory based on filters
func (l *InMemoryLogger) Query(ctx context.Context, filters map[string]interface{}, startTime, endTime time.Time) ([]*AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	result := make([]*AuditLog, 0)
	
	for _, log := range l.logs {
		// Check time range
		if !startTime.IsZero() && log.Timestamp.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && log.Timestamp.After(endTime) {
			continue
		}
		
		// Check filters
		if !matchesFilters(log, filters) {
			continue
		}
		
		// Create a copy to prevent modification
		logCopy := *log
		result = append(result, &logCopy)
	}
	
	return result, nil
}

// GetLog retrieves a specific audit log by ID from memory
func (l *InMemoryLogger) GetLog(ctx context.Context, id string) (*AuditLog, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	for _, log := range l.logs {
		if log.ID == id {
			// Return a copy to prevent modification
			logCopy := *log
			return &logCopy, nil
		}
	}
	
	return nil, fmt.Errorf("log with ID %s not found", id)
}

// PurgeLogs purges logs older than the specified time from memory
func (l *InMemoryLogger) PurgeLogs(ctx context.Context, olderThan time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	newLogs := make([]*AuditLog, 0)
	for _, log := range l.logs {
		if log.Timestamp.After(olderThan) {
			newLogs = append(newLogs, log)
		}
	}
	
	l.logs = newLogs
	return nil
}

// FileLogger is a file-based implementation of Logger
type FileLogger struct {
	filePath string
	file     *os.File
	config   *AuditTrailConfig
	mu       sync.Mutex
}

// NewFileLogger creates a new file-based logger
func NewFileLogger(filePath string, config *AuditTrailConfig) (*FileLogger, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Open log file for appending
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	return &FileLogger{
		filePath: filePath,
		file:     file,
		config:   config,
	}, nil
}

// Log logs an audit log entry to the file
func (l *FileLogger) Log(ctx context.Context, log *AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Marshal the log to JSON
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}
	
	// Write the log to the file with a newline
	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}
	
	// Check if file rotation is needed
	if l.config != nil && l.config.MaxLogFileSize > 0 {
		if info, err := l.file.Stat(); err == nil && info.Size() > int64(l.config.MaxLogFileSize*1024*1024) {
			if err := l.rotateLogFile(); err != nil {
				return fmt.Errorf("failed to rotate log file: %w", err)
			}
		}
	}
	
	return nil
}

// rotateLogFile rotates the log file
func (l *FileLogger) rotateLogFile() error {
	// Close the current file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}
	
	// Rename the current file with a timestamp
	timestamp := time.Now().Format("20060102-150405")
	newPath := fmt.Sprintf("%s.%s", l.filePath, timestamp)
	if err := os.Rename(l.filePath, newPath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}
	
	// Compress the rotated file if configured
	if l.config != nil && l.config.CompressLogs {
		if err := l.compressFile(newPath); err != nil {
			return fmt.Errorf("failed to compress log file: %w", err)
		}
	}
	
	// Open a new log file
	file, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}
	l.file = file
	
	// Clean up old log files if needed
	if l.config != nil && l.config.MaxLogFiles > 0 {
		if err := l.cleanupOldLogFiles(); err != nil {
			return fmt.Errorf("failed to clean up old log files: %w", err)
		}
	}
	
	return nil
}

// compressFile compresses a file using gzip
func (l *FileLogger) compressFile(filePath string) error {
	// Open the source file
	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for compression: %w", err)
	}
	defer src.Close()
	
	// Create the destination file
	dst, err := os.Create(filePath + ".gz")
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer dst.Close()
	
	// Create gzip writer
	gzWriter := gzip.NewWriter(dst)
	defer gzWriter.Close()
	
	// Copy data from source to gzip writer
	if _, err := io.Copy(gzWriter, src); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}
	
	// Close the gzip writer to flush all data
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to finalize compression: %w", err)
	}
	
	// Remove the original file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove original file after compression: %w", err)
	}
	
	return nil
}

// cleanupOldLogFiles removes old log files
func (l *FileLogger) cleanupOldLogFiles() error {
	// Get the directory and base filename
	dir := filepath.Dir(l.filePath)
	base := filepath.Base(l.filePath)
	
	// List all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}
	
	// Find rotated log files
	rotatedFiles := make([]string, 0)
	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, base+".") && (strings.HasSuffix(name, ".gz") || !strings.Contains(name, ".gz")) {
			rotatedFiles = append(rotatedFiles, filepath.Join(dir, name))
		}
	}
	
	// If we have more files than the limit, delete the oldest ones
	if len(rotatedFiles) > l.config.MaxLogFiles {
		// Sort files by modification time (oldest first)
		sort.Slice(rotatedFiles, func(i, j int) bool {
			iInfo, _ := os.Stat(rotatedFiles[i])
			jInfo, _ := os.Stat(rotatedFiles[j])
			return iInfo.ModTime().Before(jInfo.ModTime())
		})
		
		// Delete the oldest files
		for i := 0; i < len(rotatedFiles)-l.config.MaxLogFiles; i++ {
			if err := os.Remove(rotatedFiles[i]); err != nil {
				return fmt.Errorf("failed to remove old log file: %w", err)
			}
		}
	}
	
	return nil
}

// Query queries audit logs from files based on filters
func (l *FileLogger) Query(ctx context.Context, filters map[string]interface{}, startTime, endTime time.Time) ([]*AuditLog, error) {
	// This is a simplified implementation that reads the current log file
	// A production implementation would need to handle rotated files and be more efficient
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Flush the current file to ensure all logs are written to disk
	if err := l.file.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync log file: %w", err)
	}
	
	// Open the file for reading
	file, err := os.Open(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file for reading: %w", err)
	}
	defer file.Close()
	
	// Read and parse logs
	logs := make([]*AuditLog, 0)
	scanner := NewLineScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		var log AuditLog
		if err := json.Unmarshal([]byte(line), &log); err != nil {
			continue // Skip invalid lines
		}
		
		// Check time range
		if !startTime.IsZero() && log.Timestamp.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && log.Timestamp.After(endTime) {
			continue
		}
		
		// Check filters
		if !matchesFilters(&log, filters) {
			continue
		}
		
		logs = append(logs, &log)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}
	
	return logs, nil
}

// GetLog retrieves a specific audit log by ID from the file
func (l *FileLogger) GetLog(ctx context.Context, id string) (*AuditLog, error) {
	logs, err := l.Query(ctx, map[string]interface{}{"id": id}, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	
	if len(logs) == 0 {
		return nil, fmt.Errorf("log with ID %s not found", id)
	}
	
	return logs[0], nil
}

// PurgeLogs purges logs older than the specified time
func (l *FileLogger) PurgeLogs(ctx context.Context, olderThan time.Time) error {
	// This is a simplified implementation
	// A production implementation would need to handle rotated files more efficiently
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Get the directory and base filename
	dir := filepath.Dir(l.filePath)
	base := filepath.Base(l.filePath)
	
	// List all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}
	
	// Check rotated log files
	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, base+".") && (strings.HasSuffix(name, ".gz") || !strings.Contains(name, ".gz")) {
			// Get file info
			info, err := file.Info()
			if err != nil {
				continue
			}
			
			// If the file is older than the cutoff, delete it
			if info.ModTime().Before(olderThan) {
				if err := os.Remove(filepath.Join(dir, name)); err != nil {
					return fmt.Errorf("failed to remove old log file: %w", err)
				}
			}
		}
	}
	
	return nil
}

// Close closes the file logger
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		return l.file.Close()
	}
	
	return nil
}

// SyslogLogger is a syslog-based implementation of Logger
type SyslogLogger struct {
	config *AuditTrailConfig
	// Syslog writer would be added here
	// This is a simplified implementation
}

// NewSyslogLogger creates a new syslog-based logger
func NewSyslogLogger(config *AuditTrailConfig) (*SyslogLogger, error) {
	// Initialize syslog connection
	// This is a simplified implementation
	
	return &SyslogLogger{
		config: config,
	}, nil
}

// Log logs an audit log entry to syslog
func (l *SyslogLogger) Log(ctx context.Context, log *AuditLog) error {
	// Marshal the log to JSON
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}
	
	// Write to syslog
	// This is a simplified implementation
	fmt.Printf("Syslog: %s\n", string(data))
	
	return nil
}

// Query queries audit logs from syslog based on filters
func (l *SyslogLogger) Query(ctx context.Context, filters map[string]interface{}, startTime, endTime time.Time) ([]*AuditLog, error) {
	// Querying from syslog is not supported in this simplified implementation
	return []*AuditLog{}, nil
}

// GetLog retrieves a specific audit log by ID from syslog
func (l *SyslogLogger) GetLog(ctx context.Context, id string) (*AuditLog, error) {
	// Retrieving specific logs from syslog is not supported in this simplified implementation
	return nil, fmt.Errorf("retrieving specific logs from syslog is not supported")
}

// matchesFilters checks if a log entry matches the specified filters
func matchesFilters(log *AuditLog, filters map[string]interface{}) bool {
	if filters == nil || len(filters) == 0 {
		return true
	}
	
	for key, value := range filters {
		switch key {
		case "id":
			if log.ID != value.(string) {
				return false
			}
		case "user_id":
			if log.UserID != value.(string) {
				return false
			}
		case "username":
			if log.Username != value.(string) {
				return false
			}
		case "operation":
			if log.Operation != value.(string) {
				return false
			}
		case "resource_type":
			if log.ResourceType != value.(string) {
				return false
			}
		case "resource_id":
			if log.ResourceID != value.(string) {
				return false
			}
		case "status":
			if log.Status != value.(string) {
				return false
			}
		case "ip_address":
			if log.IPAddress != value.(string) {
				return false
			}
		}
	}
	
	return true
}

// LineScanner is a simple line scanner for reading log files
type LineScanner struct {
	reader    io.Reader
	buffer    []byte
	remaining []byte
	err       error
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
	if s.err != nil {
		return false
	}
	
	// Check if we have a line in the remaining buffer
	if i := strings.Index(string(s.remaining), "\n"); i >= 0 {
		s.buffer = s.remaining[:i]
		s.remaining = s.remaining[i+1:]
		return true
	}
	
	// Read more data
	var buf []byte
	buf = append(buf, s.remaining...)
	
	for {
		n, err := s.reader.Read(s.buffer)
		if n > 0 {
			buf = append(buf, s.buffer[:n]...)
			if i := strings.Index(string(buf), "\n"); i >= 0 {
				s.buffer = buf[:i]
				s.remaining = buf[i+1:]
				return true
			}
		}
		
		if err != nil {
			if err == io.EOF {
				if len(buf) > 0 {
					s.buffer = buf
					s.remaining = nil
					return true
				}
			}
			s.err = err
			return false
		}
	}
}

// Text returns the current line
func (s *LineScanner) Text() string {
	return string(s.buffer)
}

// Err returns the error, if any
func (s *LineScanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
