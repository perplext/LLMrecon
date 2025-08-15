#!/bin/bash

echo "Fixing all remaining syntax issues..."

# Fix security/access/audit/trail/loggers.go
echo "Fixing security/access/audit/trail/loggers.go..."
cat > src/security/access/audit/trail/loggers.go << 'EOF'
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
	
	file, err := os.OpenFile(filepath.Clean(logPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
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
	file, err := os.OpenFile(filepath.Clean(newPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
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
EOF

# Fix security/access/audit/trail/manager.go
echo "Fixing security/access/audit/trail/manager.go..."
cat > src/security/access/audit/trail/manager.go << 'EOF'
// Package trail provides audit trail management
package trail

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages audit trail operations
type Manager struct {
	trail  *AuditTrail
	logger *FileLogger
	mutex  sync.Mutex
}

// NewManager creates a new audit trail manager
func NewManager(logPath string) (*Manager, error) {
	logger, err := NewFileLogger(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file logger: %w", err)
	}
	
	trail := NewAuditTrail(logger)
	
	return &Manager{
		trail:  trail,
		logger: logger,
	}, nil
}

// LogOperation logs an operation to the audit trail
func (m *Manager) LogOperation(ctx context.Context, log *AuditLog) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	return m.trail.LogOperation(ctx, log)
}

// Close closes the audit trail manager
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if err := m.trail.Close(); err != nil {
		return fmt.Errorf("failed to close audit trail: %w", err)
	}
	
	if err := m.logger.Close(); err != nil {
		return fmt.Errorf("failed to close logger: %w", err)
	}
	
	return nil
}

// GetTrail returns the audit trail
func (m *Manager) GetTrail() *AuditTrail {
	return m.trail
}
EOF

# Fix customization/detector.go
echo "Fixing customization/detector.go..."
cat > src/customization/detector.go << 'EOF'
// Package customization provides customization detection and management
package customization

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CustomizationDetector detects user customizations
type CustomizationDetector struct {
	BasePath string
}

// NewCustomizationDetector creates a new customization detector
func NewCustomizationDetector(basePath string) *CustomizationDetector {
	return &CustomizationDetector{
		BasePath: basePath,
	}
}

// DetectCustomizations detects customizations in the given path
func (d *CustomizationDetector) DetectCustomizations() ([]Customization, error) {
	var customizations []Customization
	
	err := filepath.Walk(d.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && d.isCustomizationFile(path) {
			custom := Customization{
				Path:        path,
				Type:        "file",
				Description: "User customization detected",
			}
			customizations = append(customizations, custom)
		}
		
		return nil
	})
	
	return customizations, err
}

// isCustomizationFile checks if a file is a user customization
func (d *CustomizationDetector) isCustomizationFile(path string) bool {
	// Simple heuristic - check for common customization patterns
	base := filepath.Base(path)
	return strings.Contains(base, "custom") || 
		   strings.Contains(base, "user") ||
		   strings.HasSuffix(base, ".custom")
}

// Customization represents a detected customization
type Customization struct {
	Path        string
	Type        string
	Description string
	Hash        string
}

// CalculateHash calculates the hash of the customization
func (c *Customization) CalculateHash() error {
	content, err := os.ReadFile(filepath.Clean(c.Path))
	if err != nil {
		return err
	}
	
	hash := sha256.Sum256(content)
	c.Hash = fmt.Sprintf("%x", hash)
	return nil
}
EOF

# Fix customization/preserver.go
echo "Fixing customization/preserver.go..."
cat > src/customization/preserver.go << 'EOF'
// Package customization provides customization preservation
package customization

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CustomizationPreserver preserves user customizations
type CustomizationPreserver struct {
	BackupPath string
}

// NewCustomizationPreserver creates a new customization preserver
func NewCustomizationPreserver(backupPath string) *CustomizationPreserver {
	return &CustomizationPreserver{
		BackupPath: backupPath,
	}
}

// PreserveCustomization preserves a customization
func (p *CustomizationPreserver) PreserveCustomization(custom Customization) error {
	// Create backup directory
	if err := os.MkdirAll(p.BackupPath, 0750); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Determine backup file path
	relPath, err := filepath.Rel(".", custom.Path)
	if err != nil {
		relPath = filepath.Base(custom.Path)
	}
	
	backupFile := filepath.Join(p.BackupPath, relPath)
	backupDir := filepath.Dir(backupFile)
	
	// Create backup subdirectory if needed
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return fmt.Errorf("failed to create backup subdirectory: %w", err)
	}
	
	// Copy the file
	return p.copyFile(custom.Path, backupFile)
}

// copyFile copies a file from source to destination
func (p *CustomizationPreserver) copyFile(src, dst string) error {
	sourceFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	
	// Copy file permissions
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}
	
	return os.Chmod(dst, info.Mode())
}

// RestoreCustomization restores a customization from backup
func (p *CustomizationPreserver) RestoreCustomization(custom Customization) error {
	// Determine backup file path
	relPath, err := filepath.Rel(".", custom.Path)
	if err != nil {
		relPath = filepath.Base(custom.Path)
	}
	
	backupFile := filepath.Join(p.BackupPath, relPath)
	
	// Check if backup exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}
	
	// Restore the file
	return p.copyFile(backupFile, custom.Path)
}
EOF

# Fix customization/registry.go
echo "Fixing customization/registry.go..."
cat > src/customization/registry.go << 'EOF'
// Package customization provides customization registry
package customization

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Registry manages customization registration
type Registry struct {
	FilePath        string
	Customizations  []Customization
	mutex           sync.RWMutex
}

// NewRegistry creates a new customization registry
func NewRegistry(filePath string) *Registry {
	return &Registry{
		FilePath:       filePath,
		Customizations: make([]Customization, 0),
	}
}

// Load loads the registry from disk
func (r *Registry) Load() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	data, err := os.ReadFile(filepath.Clean(r.FilePath))
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's OK
			return nil
		}
		return fmt.Errorf("failed to read registry file: %w", err)
	}
	
	if err := json.Unmarshal(data, &r.Customizations); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}
	
	return nil
}

// Save saves the registry to disk
func (r *Registry) Save() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	data, err := json.MarshalIndent(r.Customizations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}
	
	// Create directory if needed
	dir := filepath.Dir(r.FilePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}
	
	if err := os.WriteFile(filepath.Clean(r.FilePath), data, 0640); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}
	
	return nil
}

// Register registers a customization
func (r *Registry) Register(custom Customization) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Check if already registered
	for i, existing := range r.Customizations {
		if existing.Path == custom.Path {
			// Update existing registration
			r.Customizations[i] = custom
			return nil
		}
	}
	
	// Add new registration
	r.Customizations = append(r.Customizations, custom)
	return nil
}

// GetCustomizations returns all registered customizations
func (r *Registry) GetCustomizations() []Customization {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make([]Customization, len(r.Customizations))
	copy(result, r.Customizations)
	return result
}

// FindByPath finds a customization by path
func (r *Registry) FindByPath(path string) (*Customization, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	for _, custom := range r.Customizations {
		if custom.Path == path {
			c := custom
			return &c, true
		}
	}
	
	return nil, false
}
EOF

# Fix bundle/errors/reporting.go
echo "Fixing bundle/errors/reporting.go..."
cat > src/bundle/errors/reporting.go << 'EOF'
// Package errors provides error handling functionality for bundle operations
package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ErrorReporter defines the interface for error reporting
type ErrorReporter interface {
	Report(ctx context.Context, err *BundleError) error
	GenerateReport(ctx context.Context, errors []*BundleError) (*ErrorReport, error)
}

// ErrorReport represents a collection of errors with statistics
type ErrorReport struct {
	GeneratedAt time.Time      `json:"generated_at"`
	TotalErrors int            `json:"total_errors"`
	Statistics  ErrorStats     `json:"statistics"`
	Errors      []*BundleError `json:"errors"`
}

// ErrorStats contains error statistics
type ErrorStats struct {
	BySeverity       map[string]int `json:"by_severity"`
	ByCategory       map[string]int `json:"by_category"`
	ByRecoverability map[string]int `json:"by_recoverability"`
}

// DefaultErrorReporter is the default implementation of ErrorReporter
type DefaultErrorReporter struct {
	Writer      io.Writer
	AuditLogger *AuditLogger
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(writer io.Writer, auditLogger *AuditLogger) *DefaultErrorReporter {
	if writer == nil {
		writer = os.Stdout
	}
	
	return &DefaultErrorReporter{
		Writer:      writer,
		AuditLogger: auditLogger,
	}
}

// Report reports a single error
func (r *DefaultErrorReporter) Report(ctx context.Context, err *BundleError) error {
	if err == nil {
		return nil
	}
	
	// Log the error
	fmt.Fprintf(r.Writer, "Error Report: %s (ID: %s, Category: %s, Severity: %s)\n", 
		err.Message, err.ErrorID, err.Category, err.Severity)
	
	// Log audit event
	if r.AuditLogger != nil {
		details := map[string]interface{}{
			"error_id":       err.ErrorID,
			"category":       err.Category,
			"severity":       err.Severity,
			"recoverability": err.Recoverability,
			"message":        err.Message,
		}
		
		// Add context if available
		for k, v := range err.Context {
			details[k] = v
		}
		
		r.AuditLogger.LogEventWithStatus("error_reported", "ErrorReporter", err.ErrorID, "error", details)
	}
	
	return nil
}

// GenerateReport generates a comprehensive error report
func (r *DefaultErrorReporter) GenerateReport(ctx context.Context, errors []*BundleError) (*ErrorReport, error) {
	stats := calculateErrorStats(errors)
	
	report := &ErrorReport{
		GeneratedAt: time.Now(),
		TotalErrors: len(errors),
		Statistics:  stats,
		Errors:      errors,
	}
	
	return report, nil
}

// calculateErrorStats calculates statistics from errors
func calculateErrorStats(errors []*BundleError) ErrorStats {
	stats := ErrorStats{
		BySeverity:       make(map[string]int),
		ByCategory:       make(map[string]int),
		ByRecoverability: make(map[string]int),
	}
	
	for _, err := range errors {
		stats.BySeverity[string(err.Severity)]++
		stats.ByCategory[string(err.Category)]++
		stats.ByRecoverability[string(err.Recoverability)]++
	}
	
	return stats
}

// WriteReportJSON writes an error report as JSON
func WriteReportJSON(w io.Writer, report *ErrorReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// WriteReportText writes an error report as text
func WriteReportText(w io.Writer, report *ErrorReport) error {
	fmt.Fprintf(w, "Error Report\n")
	fmt.Fprintf(w, "============\n\n")
	fmt.Fprintf(w, "Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Total Errors: %d\n\n", report.TotalErrors)
	
	// Write statistics
	fmt.Fprintf(w, "Statistics:\n")
	fmt.Fprintf(w, "-----------\n")
	
	fmt.Fprintf(w, "By Severity:\n")
	for severity, count := range report.Statistics.BySeverity {
		fmt.Fprintf(w, "  %s: %d\n", severity, count)
	}
	
	fmt.Fprintf(w, "\nBy Category:\n")
	for category, count := range report.Statistics.ByCategory {
		fmt.Fprintf(w, "  %s: %d\n", category, count)
	}
	
	fmt.Fprintf(w, "\nBy Recoverability:\n")
	for recoverability, count := range report.Statistics.ByRecoverability {
		fmt.Fprintf(w, "  %s: %d\n", recoverability, count)
	}
	
	// Write individual errors
	fmt.Fprintf(w, "\n\nErrors:\n")
	fmt.Fprintf(w, "-------\n")
	for i, err := range report.Errors {
		fmt.Fprintf(w, "\n%d. %s\n", i+1, err.Message)
		fmt.Fprintf(w, "   ID: %s\n", err.ErrorID)
		fmt.Fprintf(w, "   Category: %s\n", err.Category)
		fmt.Fprintf(w, "   Severity: %s\n", err.Severity)
		fmt.Fprintf(w, "   Recoverability: %s\n", err.Recoverability)
		
		if len(err.Context) > 0 {
			fmt.Fprintf(w, "   Context:\n")
			for k, v := range err.Context {
				fmt.Fprintf(w, "     %s: %v\n", k, v)
			}
		}
	}
	
	return nil
}
EOF

# Fix audit/audit.go - properly close structs
echo "Fixing audit/audit.go..."
# Fix the struct closing issues
sed -i '' '42s/$/}/' src/audit/audit.go
sed -i '' '53s/$/}/' src/audit/audit.go  
sed -i '' '66s/$/}/' src/audit/audit.go
sed -i '' '162s/$/}/' src/audit/audit.go
sed -i '' '172s/$/}/' src/audit/audit.go
sed -i '' '242s/$/}/' src/audit/audit.go
sed -i '' '259s/$/}/' src/audit/audit.go
sed -i '' '268s/$/}/' src/audit/audit.go
sed -i '' '291s/$/}/' src/audit/audit.go
sed -i '' '313s/$/}/' src/audit/audit.go
sed -i '' '412s/$/}/' src/audit/audit.go
sed -i '' '488s/$/}/' src/audit/audit.go

# Remove excess closing braces at the end of the file
head -n 516 src/audit/audit.go > /tmp/audit_fixed.go
echo "}" >> /tmp/audit_fixed.go
mv /tmp/audit_fixed.go src/audit/audit.go

echo "Script completed. Checking compilation..."