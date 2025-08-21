package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/perplext/LLMrecon/src/customization"
)

// CustomizationManager manages the preservation and reapplication of user customizations
// during the update process.
type CustomizationManager struct {
	// Registry is the customization registry
	Registry *customization.Registry
	// Detector detects user customizations
	Detector *customization.CustomizationDetector
	// Preserver preserves and reapplies user customizations
	Preserver *customization.CustomizationPreserver
	// InstallDir is the directory where the tool is installed
	InstallDir string
	// BackupDir is the directory for backups during update
	BackupDir string
	// Logger is the logger for customization operations
	Logger *os.File
}

// NewCustomizationManager creates a new customization manager
func NewCustomizationManager(installDir, backupDir string, logger *os.File) (*CustomizationManager, error) {
	// Create registry path
	registryPath := filepath.Join(installDir, "data", "customization-registry.json")

	// Create registry directory if it doesn't exist
	registryDir := filepath.Dir(registryPath)
	if err := os.MkdirAll(registryDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Create registry
	registry := customization.NewRegistry(registryPath)

	// Create detector
	detector := customization.NewCustomizationDetector(installDir)

	// Create preserver
	preserver := customization.NewCustomizationPreserver(backupDir)

	return &CustomizationManager{
		Registry:   registry,
		Detector:   detector,
		Preserver:  preserver,
		InstallDir: installDir,
		BackupDir:  backupDir,
		Logger:     logger,
	}, nil
}

// DetectCustomizations detects user customizations in the installation directory
func (m *CustomizationManager) DetectCustomizations() error {
	// Detect customizations
	customizations, err := m.Detector.DetectCustomizations()
	if err != nil {
		return fmt.Errorf("failed to detect customizations: %w", err)
	}

	// Register detected customizations
	for _, custom := range customizations {
		if err := m.Registry.Register(custom); err != nil {
			return fmt.Errorf("failed to register customization %s: %w", custom.Path, err)
		}
	}
	return nil
}

// PreserveCustomizations preserves user customizations before update
func (m *CustomizationManager) PreserveCustomizations() error {
	// Get all registered customizations
	customizations := m.Registry.GetCustomizations()

	// Preserve each customization
	for _, custom := range customizations {
		if err := m.Preserver.PreserveCustomization(custom); err != nil {
			return fmt.Errorf("failed to preserve customization %s: %w", custom.Path, err)
		}
	}

	return nil
}

// ReapplyCustomizations reapplies user customizations after update
func (m *CustomizationManager) ReapplyCustomizations(updatedTemplates, updatedModules []string) error {
	// Get all registered customizations
	customizations := m.Registry.GetCustomizations()

	// Restore each customization
	for _, custom := range customizations {
		if err := m.Preserver.RestoreCustomization(custom); err != nil {
			return fmt.Errorf("failed to restore customization %s: %w", custom.Path, err)
		}
	}

	return nil
}

// UpdateWithCustomizationPreservation wraps the update process with customization preservation
func UpdateWithCustomizationPreservation(ctx context.Context, installDir, backupDir string, pkg *UpdatePackage) error {
	// Create log file for customization operations
	logPath := filepath.Join(backupDir, "customization.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create customization log file: %w", err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Printf("Failed to close: %v\n", err)
		}
	}()

	// Log start of update
	fmt.Fprintf(logFile, "[%s] Starting update with customization preservation\n", time.Now().Format(time.RFC3339))

	// Create customization manager
	manager, err := NewCustomizationManager(installDir, backupDir, logFile)
	if err != nil {
		return fmt.Errorf("failed to create customization manager: %w", err)
	}

	// Detect customizations
	fmt.Fprintf(logFile, "[%s] Detecting customizations\n", time.Now().Format(time.RFC3339))
	if err := manager.DetectCustomizations(); err != nil {
		return fmt.Errorf("failed to detect customizations: %w", err)
	}

	// Preserve customizations
	fmt.Fprintf(logFile, "[%s] Preserving customizations\n", time.Now().Format(time.RFC3339))
	if err := manager.PreserveCustomizations(); err != nil {
		return fmt.Errorf("failed to preserve customizations: %w", err)
	}

	// Apply update (simplified - would need actual update logic)
	fmt.Fprintf(logFile, "[%s] Applying update\n", time.Now().Format(time.RFC3339))
	// TODO: Implement actual update logic here

	// Collect updated templates and modules (simplified)
	var updatedTemplates []string
	var updatedModules []string

	// Check if modules were updated
	if len(pkg.Manifest.Components.Modules) > 0 {
		for _, moduleInfo := range pkg.Manifest.Components.Modules {
			updatedModules = append(updatedModules, moduleInfo.ID)
		}
	}

	// Reapply customizations
	fmt.Fprintf(logFile, "[%s] Reapplying customizations\n", time.Now().Format(time.RFC3339))
	if err := manager.ReapplyCustomizations(updatedTemplates, updatedModules); err != nil {
		return fmt.Errorf("failed to reapply customizations: %w", err)
	}

	fmt.Fprintf(logFile, "[%s] Update with customization preservation completed successfully\n", time.Now().Format(time.RFC3339))
	return nil
}
