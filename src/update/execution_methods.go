package update

import (
	"context"
	"fmt"
)

// ExecuteUpdate executes an update from the given package
func (e *UpdateExecutor) ExecuteUpdate(ctx context.Context, pkg *UpdatePackage) error {
	// Create session directory
	sessionDir := filepath.Join(e.TempDir, fmt.Sprintf("update-%s-%d", pkg.Manifest.PackageID, time.Now().Unix()))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}
	defer os.RemoveAll(sessionDir)

	// Create backup directory for this update
	backupDir := filepath.Join(e.BackupDir, fmt.Sprintf("backup-%s-%d", pkg.Manifest.PackageID, time.Now().Unix()))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create JSON log file
	jsonLogFile, err := CreateJSONLogFile(e.LogDir, pkg.Manifest.PackageID)
	if err != nil {
		return fmt.Errorf("failed to create JSON log file: %w", err)
	}
	defer CloseJSONLogFile(jsonLogFile)

	// Update logger with JSON writer
	e.Logger.JSONWriter = jsonLogFile

	// Log update start
	e.Logger.Info("UpdateExecutor", "Starting update execution", "", map[string]interface{}{
		"package_id": pkg.Manifest.PackageID,
		"version":    pkg.Manifest.Components.Binary.Version,
	})

	// Create update transaction
	transaction := NewUpdateTransaction(pkg.Manifest.PackageID, sessionDir, backupDir, e.Logger.Writer)

	// Notify update started
	e.NotificationManager.NotifyUpdateStarted(transaction.ID, pkg.Manifest.PackageID, map[string]interface{}{
		"version": pkg.Manifest.Components.Binary.Version,
	})

	// Log audit event
	e.AuditLogger.LogEvent("update_started", "UpdateExecutor", e.User, transaction.ID, pkg.Manifest.PackageID, map[string]interface{}{
		"version": pkg.Manifest.Components.Binary.Version,
	})

	// Verify package integrity
	result, err := e.Verifier.VerifyPackage(pkg)
	if err != nil {
		e.Logger.Error("UpdateExecutor", fmt.Sprintf("Package integrity verification failed: %v", err), transaction.ID, nil)
		e.NotificationManager.NotifyVerificationFailed(pkg.Manifest.PackageID, err.Error(), nil)
		return fmt.Errorf("package integrity verification failed: %w", err)
	}

	if !result.Success {
		e.Logger.Error("UpdateExecutor", fmt.Sprintf("Package integrity verification failed: %s", result.Message), transaction.ID, result.Details)
		e.NotificationManager.NotifyVerificationFailed(pkg.Manifest.PackageID, result.Message, result.Details)
		return fmt.Errorf("package integrity verification failed: %s", result.Message)
	}

	// Verify package compatibility
	result, err = e.Verifier.VerifyCompatibility(pkg, e.CurrentVersions)
	if err != nil {
		e.Logger.Error("UpdateExecutor", fmt.Sprintf("Package compatibility verification failed: %v", err), transaction.ID, nil)
		e.NotificationManager.NotifyVerificationFailed(pkg.Manifest.PackageID, err.Error(), nil)
		return fmt.Errorf("package compatibility verification failed: %w", err)
	}

	if !result.Success {
		e.Logger.Error("UpdateExecutor", fmt.Sprintf("Package compatibility verification failed: %s", result.Message), transaction.ID, result.Details)
		e.NotificationManager.NotifyVerificationFailed(pkg.Manifest.PackageID, result.Message, result.Details)
		return fmt.Errorf("package compatibility verification failed: %s", result.Message)
	}

	// Run custom verification hooks
	for _, hook := range e.VerificationHooks {
		result, err := hook(ctx, pkg)
		if err != nil {
			e.Logger.Error("UpdateExecutor", fmt.Sprintf("Verification hook failed: %v", err), transaction.ID, nil)
			e.NotificationManager.NotifyVerificationFailed(pkg.Manifest.PackageID, err.Error(), nil)
			return fmt.Errorf("verification hook failed: %w", err)
		}

		if !result.Success {
			e.Logger.Error("UpdateExecutor", fmt.Sprintf("Verification hook failed: %s", result.Message), transaction.ID, result.Details)
			e.NotificationManager.NotifyVerificationFailed(pkg.Manifest.PackageID, result.Message, result.Details)
			return fmt.Errorf("verification hook failed: %s", result.Message)
		}
	}

	// Begin transaction
	if err := transaction.Begin(); err != nil {
		e.Logger.Error("UpdateExecutor", fmt.Sprintf("Failed to begin transaction: %v", err), transaction.ID, nil)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Run pre-update hooks
	for _, hook := range e.PreUpdateHooks {
		if err := hook(ctx, transaction); err != nil {
			e.Logger.Error("UpdateExecutor", fmt.Sprintf("Pre-update hook failed: %v", err), transaction.ID, nil)
			return fmt.Errorf("pre-update hook failed: %w", err)
		}
	}

	// Apply update based on package type
	var updateErr error
	if pkg.Manifest.PackageType == FullPackage {
		updateErr = e.executeFullUpdate(ctx, pkg, transaction)
	} else if pkg.Manifest.PackageType == DifferentialPackage {
		updateErr = e.executeDifferentialUpdate(ctx, pkg, transaction)
	} else {
		updateErr = fmt.Errorf("unsupported package type: %s", pkg.Manifest.PackageType)
	}

	if updateErr != nil {
		e.Logger.Error("UpdateExecutor", fmt.Sprintf("Failed to apply update: %v", updateErr), transaction.ID, nil)
		
		// Rollback transaction
		rollbackErr := transaction.Rollback()
		if rollbackErr != nil {
			e.Logger.Error("UpdateExecutor", fmt.Sprintf("Failed to rollback transaction: %v", rollbackErr), transaction.ID, nil)
			return fmt.Errorf("failed to apply update and rollback transaction: %v (rollback error: %v)", updateErr, rollbackErr)
		}

		// Notify update rolled back
		e.NotificationManager.NotifyUpdateRolledBack(transaction.ID, pkg.Manifest.PackageID, transaction.GetSummary())

		// Log audit event
		e.AuditLogger.LogEvent("update_rolled_back", "UpdateExecutor", e.User, transaction.ID, pkg.Manifest.PackageID, transaction.GetSummary())

		return fmt.Errorf("failed to apply update (rolled back): %w", updateErr)
	}

	// Commit transaction
	if err := transaction.Commit(); err != nil {
		e.Logger.Error("UpdateExecutor", fmt.Sprintf("Failed to commit transaction: %v", err), transaction.ID, nil)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Run post-update hooks
	for _, hook := range e.PostUpdateHooks {
		if err := hook(ctx, transaction); err != nil {
			e.Logger.Error("UpdateExecutor", fmt.Sprintf("Post-update hook failed: %v", err), transaction.ID, nil)
			return fmt.Errorf("post-update hook failed: %w", err)
		}
	}

	// Notify update completed
	e.NotificationManager.NotifyUpdateCompleted(transaction.ID, pkg.Manifest.PackageID, transaction.GetSummary())

	// Log audit event
	e.AuditLogger.LogEvent("update_completed", "UpdateExecutor", e.User, transaction.ID, pkg.Manifest.PackageID, transaction.GetSummary())

	// Log update success
	e.Logger.Info("UpdateExecutor", "Update execution completed successfully", transaction.ID, transaction.GetSummary())

	return nil
}
