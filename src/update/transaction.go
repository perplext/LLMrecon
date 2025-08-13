// Package update provides functionality for checking and applying updates
package update

import (
	"context"
	"fmt"
)

// TransactionStatus represents the status of an update transaction
type TransactionStatus string

const (
	// TransactionPending indicates the transaction is pending
	TransactionPending TransactionStatus = "pending"
	// TransactionInProgress indicates the transaction is in progress
	TransactionInProgress TransactionStatus = "in_progress"
	// TransactionCommitted indicates the transaction is committed
	TransactionCommitted TransactionStatus = "committed"
	// TransactionRolledBack indicates the transaction is rolled back
	TransactionRolledBack TransactionStatus = "rolled_back"
	// TransactionFailed indicates the transaction failed
	TransactionFailed TransactionStatus = "failed"
)

// UpdateComponent represents a component being updated
type UpdateComponent string

const (
	// BinaryUpdateComponent represents the binary component
	BinaryUpdateComponent UpdateComponent = "binary"
	// TemplatesUpdateComponent represents the templates component
	TemplatesUpdateComponent UpdateComponent = "templates"
	// ModuleUpdateComponent represents a module component
	ModuleUpdateComponent UpdateComponent = "module"
)

// UpdateOperation represents an operation in an update transaction
type UpdateOperation struct {
	// Component is the component being updated
	Component UpdateComponent
	// ComponentID is the ID of the component (e.g., module ID)
	ComponentID string
	// SourcePath is the path to the source files
	SourcePath string
	// DestinationPath is the path to the destination files
	DestinationPath string
	// BackupPath is the path to the backup files
	BackupPath string
	// Status is the status of the operation
	Status TransactionStatus
	// Error is any error that occurred during the operation
	Error error
	// Timestamp is the time the operation was performed
	Timestamp time.Time
}

// UpdateTransaction represents a transaction for applying an update
type UpdateTransaction struct {
	// ID is the unique identifier for the transaction
	ID string
	// PackageID is the ID of the update package
	PackageID string
	// Status is the status of the transaction
	Status TransactionStatus
	// Operations is the list of operations in the transaction
	Operations []*UpdateOperation
	// StartTime is the time the transaction started
	StartTime time.Time
	// EndTime is the time the transaction ended
	EndTime time.Time
	// Logger is the logger for transaction operations
	Logger io.Writer
	// SessionDir is the directory for temporary files during the transaction
	SessionDir string
	// BackupDir is the directory for backups during the transaction
	BackupDir string
}

// NewUpdateTransaction creates a new update transaction
func NewUpdateTransaction(packageID, sessionDir, backupDir string, logger io.Writer) *UpdateTransaction {
	return &UpdateTransaction{
		ID:         fmt.Sprintf("update-%s-%d", packageID, time.Now().Unix()),
		PackageID:  packageID,
		Status:     TransactionPending,
		Operations: make([]*UpdateOperation, 0),
		StartTime:  time.Now(),
		Logger:     logger,
		SessionDir: sessionDir,
		BackupDir:  backupDir,
	}
}

// Begin begins the transaction
func (t *UpdateTransaction) Begin() error {
	// Log transaction start
	fmt.Fprintf(t.Logger, "[%s] Beginning update transaction %s for package %s\n", 
		time.Now().Format(time.RFC3339), t.ID, t.PackageID)

	// Update status
	t.Status = TransactionInProgress

	return nil
}

// AddOperation adds an operation to the transaction
func (t *UpdateTransaction) AddOperation(component UpdateComponent, componentID, sourcePath, destPath, backupPath string) *UpdateOperation {
	operation := &UpdateOperation{
		Component:       component,
		ComponentID:     componentID,
		SourcePath:      sourcePath,
		DestinationPath: destPath,
		BackupPath:      backupPath,
		Status:          TransactionPending,
		Timestamp:       time.Now(),
	}

	t.Operations = append(t.Operations, operation)
	return operation
}

// ExecuteOperation executes an operation in the transaction
func (t *UpdateTransaction) ExecuteOperation(ctx context.Context, operation *UpdateOperation) error {
	// Log operation start
	fmt.Fprintf(t.Logger, "[%s] Executing operation: %s %s\n", 
		time.Now().Format(time.RFC3339), operation.Component, operation.ComponentID)

	// Update status
	operation.Status = TransactionInProgress
	operation.Timestamp = time.Now()

	// Create backup directory
	backupDir := filepath.Dir(operation.BackupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		operation.Status = TransactionFailed
		operation.Error = fmt.Errorf("failed to create backup directory: %w", err)
		return operation.Error
	}

	// Create backup
	if _, err := os.Stat(operation.DestinationPath); err == nil {
		// Destination exists, create backup
		if err := copyDir(operation.DestinationPath, operation.BackupPath); err != nil {
			operation.Status = TransactionFailed
			operation.Error = fmt.Errorf("failed to create backup: %w", err)
			return operation.Error
		}
	} else if !os.IsNotExist(err) {
		// Error other than "not exists"
		operation.Status = TransactionFailed
		operation.Error = fmt.Errorf("failed to check destination path: %w", err)
		return operation.Error
	} else {
		// Destination doesn't exist, create backup directory
		if err := os.MkdirAll(operation.BackupPath, 0755); err != nil {
			operation.Status = TransactionFailed
			operation.Error = fmt.Errorf("failed to create backup directory: %w", err)
			return operation.Error
		}
	}

	// Create destination directory
	destDir := filepath.Dir(operation.DestinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		operation.Status = TransactionFailed
		operation.Error = fmt.Errorf("failed to create destination directory: %w", err)
		return operation.Error
	}

	// Copy source to destination
	if err := replaceDir(operation.SourcePath, operation.DestinationPath); err != nil {
		operation.Status = TransactionFailed
		operation.Error = fmt.Errorf("failed to copy source to destination: %w", err)
		return operation.Error
	}

	// Update status
	operation.Status = TransactionCommitted
	operation.Timestamp = time.Now()

	// Log operation success
	fmt.Fprintf(t.Logger, "[%s] Operation completed successfully: %s %s\n", 
		time.Now().Format(time.RFC3339), operation.Component, operation.ComponentID)

	return nil
}

// Commit commits the transaction
func (t *UpdateTransaction) Commit() error {
	// Log transaction commit
	fmt.Fprintf(t.Logger, "[%s] Committing update transaction %s\n", 
		time.Now().Format(time.RFC3339), t.ID)

	// Update status
	t.Status = TransactionCommitted
	t.EndTime = time.Now()

	// Log transaction success
	fmt.Fprintf(t.Logger, "[%s] Update transaction %s committed successfully\n", 
		time.Now().Format(time.RFC3339), t.ID)

	return nil
}

// Rollback rolls back the transaction
func (t *UpdateTransaction) Rollback() error {
	// Log transaction rollback
	fmt.Fprintf(t.Logger, "[%s] Rolling back update transaction %s\n", 
		time.Now().Format(time.RFC3339), t.ID)

	// Rollback operations in reverse order
	for i := len(t.Operations) - 1; i >= 0; i-- {
		operation := t.Operations[i]

		// Skip operations that weren't executed
		if operation.Status != TransactionCommitted && operation.Status != TransactionInProgress {
			continue
		}

		// Log operation rollback
		fmt.Fprintf(t.Logger, "[%s] Rolling back operation: %s %s\n", 
			time.Now().Format(time.RFC3339), operation.Component, operation.ComponentID)

		// Check if backup exists
		if _, err := os.Stat(operation.BackupPath); os.IsNotExist(err) {
			// No backup, remove destination
			if err := os.RemoveAll(operation.DestinationPath); err != nil {
				fmt.Fprintf(t.Logger, "[%s] Warning: Failed to remove destination during rollback: %v\n", 
					time.Now().Format(time.RFC3339), err)
			}
		} else {
			// Restore from backup
			if err := replaceDir(operation.BackupPath, operation.DestinationPath); err != nil {
				fmt.Fprintf(t.Logger, "[%s] Warning: Failed to restore from backup during rollback: %v\n", 
					time.Now().Format(time.RFC3339), err)
			}
		}

		// Update operation status
		operation.Status = TransactionRolledBack
		operation.Timestamp = time.Now()
	}

	// Update transaction status
	t.Status = TransactionRolledBack
	t.EndTime = time.Now()

	// Log transaction rollback success
	fmt.Fprintf(t.Logger, "[%s] Update transaction %s rolled back successfully\n", 
		time.Now().Format(time.RFC3339), t.ID)

	return nil
}

// GetOperationsByComponent returns operations for a specific component
func (t *UpdateTransaction) GetOperationsByComponent(component UpdateComponent) []*UpdateOperation {
	var operations []*UpdateOperation
	for _, op := range t.Operations {
		if op.Component == component {
			operations = append(operations, op)
		}
	}
	return operations
}

// GetOperationsByStatus returns operations with a specific status
func (t *UpdateTransaction) GetOperationsByStatus(status TransactionStatus) []*UpdateOperation {
	var operations []*UpdateOperation
	for _, op := range t.Operations {
		if op.Status == status {
			operations = append(operations, op)
		}
	}
	return operations
}

// GetFailedOperations returns operations that failed
func (t *UpdateTransaction) GetFailedOperations() []*UpdateOperation {
	return t.GetOperationsByStatus(TransactionFailed)
}

// HasFailedOperations returns true if any operations failed
func (t *UpdateTransaction) HasFailedOperations() bool {
	return len(t.GetFailedOperations()) > 0
}

// GetSummary returns a summary of the transaction
func (t *UpdateTransaction) GetSummary() map[string]interface{} {
	// Count operations by status
	statusCounts := make(map[TransactionStatus]int)
	for _, op := range t.Operations {
		statusCounts[op.Status]++
	}

	// Count operations by component
	componentCounts := make(map[UpdateComponent]int)
	for _, op := range t.Operations {
		componentCounts[op.Component]++
	}

	// Create summary
	return map[string]interface{}{
		"id":              t.ID,
		"package_id":      t.PackageID,
		"status":          t.Status,
		"start_time":      t.StartTime,
		"end_time":        t.EndTime,
		"duration":        t.EndTime.Sub(t.StartTime).String(),
		"operation_count": len(t.Operations),
		"status_counts":   statusCounts,
		"component_counts": componentCounts,
		"failed":          t.HasFailedOperations(),
	}
}
