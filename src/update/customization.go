package update

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/customization"
)

// CustomizationManager manages the preservation and reapplication of user customizations
// during the update process.
type CustomizationManager struct {
	// Registry is the customization registry
	Registry *customization.Registry
	// Detector detects user customizations
	Detector *customization.Detector
	// Preserver preserves and reapplies user customizations
	Preserver *customization.Preserver
	// InstallDir is the directory where the tool is installed
	InstallDir string
	// BackupDir is the directory for backups during update
	BackupDir string
	// Logger is the logger for customization operations
	Logger *os.File

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
	registry, err := customization.NewRegistry(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	// Create detector
	detector, err := customization.NewDetector(&customization.DetectorOptions{
		Registry:      registry,
		BaseDir:       installDir,
		CustomizedDir: installDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create detector: %w", err)
	}

	// Create preserver
	preserver, err := customization.NewPreserver(&customization.PreserverOptions{
		Registry:   registry,
		InstallDir: installDir,
		BackupDir:  backupDir,
		LogFile:    logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create preserver: %w", err)
	}

	return &CustomizationManager{
		Registry:   registry,
		Detector:   detector,
		Preserver:  preserver,
		InstallDir: installDir,
		BackupDir:  backupDir,
		Logger:     logger,
	}, nil
	

// DetectCustomizations detects user customizations in the installation directory
func (m *CustomizationManager) DetectCustomizations() error {
	// Register customizations
	if err := m.Detector.RegisterCustomizations(); err != nil {
		return fmt.Errorf("failed to register customizations: %w", err)
	}
	return nil
// PreserveCustomizations preserves user customizations before update
func (m *CustomizationManager) PreserveCustomizations() error {
	// Preserve template customizations
	if err := m.Preserver.PreserveTemplateCustomizations(); err != nil {
		return fmt.Errorf("failed to preserve template customizations: %w", err)
	}

	// Preserve module customizations
	if err := m.Preserver.PreserveModuleCustomizations(); err != nil {
		return fmt.Errorf("failed to preserve module customizations: %w", err)
	}

	return nil

// ReapplyCustomizations reapplies user customizations after update
func (m *CustomizationManager) ReapplyCustomizations(updatedTemplates, updatedModules []string) error {
	// Reapply template customizations
	if err := m.Preserver.ReapplyTemplateCustomizations(updatedTemplates); err != nil {
		return fmt.Errorf("failed to reapply template customizations: %w", err)
	}
	// Reapply module customizations
	if err := m.Preserver.ReapplyModuleCustomizations(updatedModules); err != nil {
		return fmt.Errorf("failed to reapply module customizations: %w", err)
	}

	return nil

// UpdateWithCustomizationPreservation extends the UpdateApplier to add customization preservation
func (a *UpdateApplier) UpdateWithCustomizationPreservation(ctx context.Context, pkg *UpdatePackage) error {
	// Create log file for customization operations
	logPath := filepath.Join(a.TempDir, "customization.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create customization log file: %w", err)
	}
	defer func() { if err := logFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Log start of update
	fmt.Fprintf(logFile, "[%s] Starting update with customization preservation\n", time.Now().Format(time.RFC3339))

	// Create customization manager
	manager, err := NewCustomizationManager(a.InstallDir, a.BackupDir, logFile)
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

	// Apply update
	fmt.Fprintf(logFile, "[%s] Applying update\n", time.Now().Format(time.RFC3339))
	if err := a.ApplyUpdate(ctx, pkg); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	// Collect updated templates and modules
	var updatedTemplates []string
	var updatedModules []string

	// Check if templates were updated
	if pkg.Manifest.Components.Templates.Version != "" {
		// All templates were updated
		templatesDir := filepath.Join(a.InstallDir, "templates")
		entries, err := os.ReadDir(templatesDir)
		if err != nil {
			return fmt.Errorf("failed to read templates directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				updatedTemplates = append(updatedTemplates, entry.Name())
			}
		}
	}

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
