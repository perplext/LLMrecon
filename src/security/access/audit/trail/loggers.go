// Package trail provides audit trail loggers
package trail

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileLogger logs audit events to a file
type FileLogger struct {
	file *os.File
}

// NewFileLogger creates a new file logger
func NewFileLogger(logPath string) (*FileLogger, error) {
	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	file, err := os.OpenFile(filepath.Clean(logPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	return &FileLogger{file: file}, nil
}

// Write writes data to the log file
func (f *FileLogger) Write(data []byte) (int, error) {
	return f.file.Write(data)
}

// Close closes the log file
func (f *FileLogger) Close() error {
	return f.file.Close()
}

// RotatingLogger provides log rotation functionality
type RotatingLogger struct {
	basePath    string
	maxSize     int64
	currentFile *os.File
	currentSize int64
}

// NewRotatingLogger creates a new rotating logger
func NewRotatingLogger(basePath string, maxSize int64) *RotatingLogger {
	return &RotatingLogger{
		basePath: basePath,
		maxSize:  maxSize,
	}
}

// Write writes data with automatic rotation
func (r *RotatingLogger) Write(data []byte) (int, error) {
	if r.currentFile == nil || r.currentSize+int64(len(data)) > r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	
	n, err := r.currentFile.Write(data)
	r.currentSize += int64(n)
	return n, err
}

// rotate rotates the log file
func (r *RotatingLogger) rotate() error {
	if r.currentFile != nil {
		if err := r.currentFile.Close(); err != nil {
			return err
		}
	}
	
	// Create new file with timestamp
	newPath := fmt.Sprintf("%s-%s.log", r.basePath, time.Now().Format("20060102-150405"))
	file, err := os.OpenFile(filepath.Clean(newPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}
	
	r.currentFile = file
	r.currentSize = 0
	return nil
}

// Close closes the current log file
func (r *RotatingLogger) Close() error {
	if r.currentFile != nil {
		return r.currentFile.Close()
	}
	return nil
}
