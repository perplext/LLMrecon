package update

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// UpdateManager handles all update operations for the tool
type UpdateManager struct {
	config     *Config
	checker    *VersionChecker
	compChecker *ComponentVersionChecker
	downloader *UpdateDownloader
	verifier   *Verifier
	installer  *Installer
	bundler    *BundleManager
	logger     Logger

// Config holds update configuration
type Config struct {
	// Binary update settings
	BinaryUpdateEnabled bool
	BinaryRepo          string
	BinaryUpdateURL     string
	
	// Template update settings
	TemplateUpdateEnabled bool
	TemplateRepos         []RepositoryConfig
	TemplateDirectory     string
	
	// Module update settings
	ModuleUpdateEnabled bool
	ModuleRepos         []RepositoryConfig
	ModuleDirectory     string
	
	// Security settings
	VerifySignatures     bool
	TrustedKeys          []string
	ChecksumVerification bool
	
	// General settings
	AutoUpdate           bool
	UpdateCheckInterval  time.Duration
	BackupEnabled        bool
	BackupDirectory      string
	
	// Network settings
	Timeout     time.Duration
	MaxRetries  int
	ProxyURL    string
	UserAgent   string
}

// RepositoryConfig defines a source for updates
type RepositoryConfig struct {
	Name     string
	URL      string
	Branch   string
	Token    string
	Type     RepositoryType
	Priority int

// RepositoryType represents the type of repository
type RepositoryType string

const (
	RepositoryTypeGitHub RepositoryType = "github"
	RepositoryTypeGitLab RepositoryType = "gitlab"
	RepositoryTypeHTTP   RepositoryType = "http"
	RepositoryTypeLocal  RepositoryType = "local"
)

// ManagerUpdateResult represents the result of an update operation
type ManagerUpdateResult struct {
	Component     string
	Success       bool
	OldVersion    string
	NewVersion    string
	ChangelogURL  string
	FilesUpdated  []string
	Error         error
	Duration      time.Duration

// UpdateSummary contains summary of all updates
type UpdateSummary struct {
	Results       []ManagerUpdateResult
	TotalDuration time.Duration
	Success       bool
	RestartRequired bool

// Logger interface for update operations
type Logger interface {
	Info(msg string)
	Error(msg string, err error)
	Debug(msg string)
	Warn(msg string)

// NewManager creates a new update manager
func NewUpdateManager(config *Config, logger Logger) *UpdateManager {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Create version checker - for now, use nil as we can't handle errors here
	checker, _ := NewVersionChecker(context.Background())
	compChecker := NewComponentVersionChecker(config, logger)
	
	return &UpdateManager{
		config:      config,
		checker:     checker,
		compChecker: compChecker,
		downloader:  NewUpdateDownloader(config, logger),
		verifier:    NewVerifier(config, logger),
		installer:   NewInstaller(config, logger),
		bundler:     NewBundleManager(config, logger),
		logger:      logger,
	}

// DefaultConfig returns default update configuration
func DefaultConfig() *Config {
	return &Config{
		BinaryUpdateEnabled:   true,
		BinaryRepo:           "github.com/perplext/LLMrecon",
		BinaryUpdateURL:      "https://api.github.com/repos/LLMrecon/LLMrecon/releases/latest",
		
		TemplateUpdateEnabled: true,
		TemplateRepos: []RepositoryConfig{
			{
				Name:     "official",
				URL:      "https://github.com/LLMrecon/templates",
				Branch:   "main",
				Type:     RepositoryTypeGitHub,
				Priority: 1,
			},
		},
		TemplateDirectory: "./templates",
		
		ModuleUpdateEnabled: true,
		ModuleRepos: []RepositoryConfig{
			{
				Name:     "official",
				URL:      "https://github.com/LLMrecon/modules",
				Branch:   "main",
				Type:     RepositoryTypeGitHub,
				Priority: 1,
			},
		},
		ModuleDirectory: "./modules",
		
		VerifySignatures:     true,
		ChecksumVerification: true,
		AutoUpdate:           false,
		UpdateCheckInterval:  24 * time.Hour,
		BackupEnabled:        true,
		BackupDirectory:      "./backups",
		
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		UserAgent:  "LLM-Red-Team-Updater/1.0",
	}

// CheckForUpdates checks for available updates for all components
func (m *UpdateManager) CheckForUpdates(ctx context.Context) (*UpdateCheck, error) {
	m.logger.Info("Checking for updates...")
	
	check := &UpdateCheck{
		CheckTime: time.Now(),
		Components: make(map[string]*ComponentUpdate),
	}
	
	// Check binary updates
	if m.config.BinaryUpdateEnabled {
		binaryUpdate, err := m.compChecker.CheckBinaryUpdate(ctx)
		if err != nil {
			m.logger.Error("Failed to check binary updates", err)
		} else {
			check.Components["binary"] = binaryUpdate
		}
	}
	
	// Check template updates
	if m.config.TemplateUpdateEnabled {
		templateUpdate, err := m.compChecker.CheckTemplateUpdates(ctx)
		if err != nil {
			m.logger.Error("Failed to check template updates", err)
		} else {
			check.Components["templates"] = templateUpdate
		}
	}
	
	// Check module updates
	if m.config.ModuleUpdateEnabled {
		moduleUpdate, err := m.compChecker.CheckModuleUpdates(ctx)
		if err != nil {
			m.logger.Error("Failed to check module updates", err)
		} else {
			check.Components["modules"] = moduleUpdate
		}
	}
	
	// Determine if any updates are available
	check.UpdatesAvailable = false
	for _, component := range check.Components {
		if component.Available {
			check.UpdatesAvailable = true
			break
		}
	}
	
	return check, nil

// ApplyUpdates applies all available updates
func (m *UpdateManager) ApplyUpdates(ctx context.Context, components []string) (*UpdateSummary, error) {
	startTime := time.Now()
	summary := &UpdateSummary{
		Results: make([]ManagerUpdateResult, 0),
		Success: true,
	}
	
	m.logger.Info("Starting update process...")
	
	// Create backup if enabled
	if m.config.BackupEnabled {
		if err := m.createBackup(); err != nil {
			m.logger.Error("Failed to create backup", err)
			return nil, fmt.Errorf("backup failed: %w", err)
		}
	}
	
	// Update components in order: templates, modules, then binary
	updateOrder := []string{"templates", "modules", "binary"}
	
	for _, component := range updateOrder {
		if len(components) > 0 && !contains(components, component) {
			continue
		}
		
		result := m.updateComponent(ctx, component)
		summary.Results = append(summary.Results, result)
		
		if !result.Success {
			summary.Success = false
			m.logger.Error(fmt.Sprintf("Failed to update %s", component), result.Error)
		}
		
		// Binary update requires restart
		if component == "binary" && result.Success {
			summary.RestartRequired = true
		}
	}
	
	summary.TotalDuration = time.Since(startTime)
	
	if summary.Success {
		m.logger.Info("All updates completed successfully")
	} else {
		m.logger.Error("Some updates failed", nil)
	}
	
	return summary, nil

// updateComponent updates a specific component
func (m *UpdateManager) updateComponent(ctx context.Context, component string) ManagerUpdateResult {
	startTime := time.Now()
	result := ManagerUpdateResult{
		Component: component,
		Success:   false,
		Duration:  0,
	}
	
	defer func() {
		result.Duration = time.Since(startTime)
	}()
	
	switch component {
	case "binary":
		return m.updateBinary(ctx)
	case "templates":
		return m.updateTemplates(ctx)
	case "modules":
		return m.updateModules(ctx)
	default:
		result.Error = fmt.Errorf("unknown component: %s", component)
		return result
	}

// updateBinary updates the main binary
func (m *UpdateManager) updateBinary(ctx context.Context) ManagerUpdateResult {
	result := ManagerUpdateResult{
		Component: "binary",
		Success:   false,
	}
	
	// Get current version
	currentVersion, err := GetCurrentVersion()
	if err != nil {
		result.Error = fmt.Errorf("failed to get current version: %w", err)
		return result
	}
	result.OldVersion = currentVersion
	
	// Check for latest version
	latestRelease, err := m.compChecker.GetLatestBinaryRelease(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to get latest release: %w", err)
		return result
	}
	
	result.NewVersion = latestRelease.Version
	result.ChangelogURL = latestRelease.ChangelogURL
	
	// Check if update is needed
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		result.Error = fmt.Errorf("invalid current version: %w", err)
		return result
	}
	
	latest, err := semver.NewVersion(latestRelease.Version)
	if err != nil {
		result.Error = fmt.Errorf("invalid latest version: %w", err)
		return result
	}
	
	if !latest.GreaterThan(current) {
		result.Success = true // No update needed
		return result
	}
	
	// Download new binary
	m.logger.Info(fmt.Sprintf("Downloading binary update %s -> %s", currentVersion, latestRelease.Version))
	
	asset := m.selectBinaryAsset(latestRelease.Assets)
	if asset == nil {
		result.Error = fmt.Errorf("no compatible binary found for platform")
		return result
	}
	
	downloadPath, err := m.downloader.DownloadFile(ctx, asset.DownloadURL, asset.Name)
	if err != nil {
		result.Error = fmt.Errorf("failed to download binary: %w", err)
		return result
	}
	defer os.Remove(downloadPath) // Clean up
	
	// Verify download
	if err := m.verifier.VerifyFile(downloadPath, asset.Checksum, asset.SignatureURL); err != nil {
		result.Error = fmt.Errorf("verification failed: %w", err)
		return result
	}
	
	// Install new binary
	if err := m.installer.InstallBinary(downloadPath); err != nil {
		result.Error = fmt.Errorf("installation failed: %w", err)
		return result
	}
	
	result.Success = true
	result.FilesUpdated = []string{os.Args[0]} // Current executable
	
	return result

// updateTemplates updates template files
func (m *UpdateManager) updateTemplates(ctx context.Context) ManagerUpdateResult {
	result := ManagerUpdateResult{
		Component: "templates",
		Success:   false,
	}
	
	// Get current template versions
	currentVersions := m.getTemplateVersions()
	result.OldVersion = formatVersionMap(currentVersions)
	
	updatedFiles := make([]string, 0)
	
	// Update from each repository
	for _, repo := range m.config.TemplateRepos {
		m.logger.Info(fmt.Sprintf("Updating templates from %s", repo.Name))
		
		files, err := m.updateFromRepository(ctx, repo, m.config.TemplateDirectory)
		if err != nil {
			m.logger.Error(fmt.Sprintf("Failed to update from %s", repo.Name), err)
			continue
		}
		
		updatedFiles = append(updatedFiles, files...)
	}
	
	if len(updatedFiles) == 0 {
		result.Success = true // No updates needed
		return result
	}
	
	// Get new template versions
	newVersions := m.getTemplateVersions()
	result.NewVersion = formatVersionMap(newVersions)
	result.FilesUpdated = updatedFiles
	result.Success = true
	
	return result

// updateModules updates provider modules
func (m *UpdateManager) updateModules(ctx context.Context) ManagerUpdateResult {
	result := ManagerUpdateResult{
		Component: "modules",
		Success:   false,
	}
	
	// Get current module versions
	currentVersions := m.getModuleVersions()
	result.OldVersion = formatVersionMap(currentVersions)
	
	updatedFiles := make([]string, 0)
	
	// Update from each repository
	for _, repo := range m.config.ModuleRepos {
		m.logger.Info(fmt.Sprintf("Updating modules from %s", repo.Name))
		
		files, err := m.updateFromRepository(ctx, repo, m.config.ModuleDirectory)
		if err != nil {
			m.logger.Error(fmt.Sprintf("Failed to update from %s", repo.Name), err)
			continue
		}
		
		updatedFiles = append(updatedFiles, files...)
	}
	
	if len(updatedFiles) == 0 {
		result.Success = true // No updates needed
		return result
	}
	
	// Get new module versions
	newVersions := m.getModuleVersions()
	result.NewVersion = formatVersionMap(newVersions)
	result.FilesUpdated = updatedFiles
	result.Success = true
	
	return result

// updateFromRepository updates files from a specific repository
func (m *UpdateManager) updateFromRepository(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	switch repo.Type {
	case RepositoryTypeGitHub:
		return m.updateFromGitHub(ctx, repo, targetDir)
	case RepositoryTypeGitLab:
		return m.updateFromGitLab(ctx, repo, targetDir)
	case RepositoryTypeHTTP:
		return m.updateFromHTTP(ctx, repo, targetDir)
	case RepositoryTypeLocal:
		return m.updateFromLocal(ctx, repo, targetDir)
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", repo.Type)
	}

// createBackup creates a backup of current installation
func (m *UpdateManager) createBackup() error {
	backupDir := filepath.Join(m.config.BackupDirectory, fmt.Sprintf("backup_%d", time.Now().Unix()))
	
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Backup binary
	if err := m.backupBinary(backupDir); err != nil {
		return fmt.Errorf("failed to backup binary: %w", err)
	}
	
	// Backup templates
	if err := m.backupDirectory(m.config.TemplateDirectory, filepath.Join(backupDir, "templates")); err != nil {
		return fmt.Errorf("failed to backup templates: %w", err)
	}
	
	// Backup modules
	if err := m.backupDirectory(m.config.ModuleDirectory, filepath.Join(backupDir, "modules")); err != nil {
		return fmt.Errorf("failed to backup modules: %w", err)
	}
	
	m.logger.Info(fmt.Sprintf("Backup created at %s", backupDir))
	return nil

// backupBinary backs up the current binary
func (m *UpdateManager) backupBinary(backupDir string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	
	backupPath := filepath.Join(backupDir, "binary", filepath.Base(execPath))
	if err := os.MkdirAll(filepath.Dir(backupPath), 0700); err != nil {
		return err
	}
	
	return copyFileHelper(execPath, backupPath)

// backupDirectory backs up a directory
func (m *UpdateManager) backupDirectory(srcDir, destDir string) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil // Source doesn't exist, nothing to backup
	}
	
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
			relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
			
		destPath := filepath.Join(destDir, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		
		return copyFileHelper(path, destPath)
	})

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false

func copyFileHelper(src, dst string) error {
	sourceFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer func() { if err := sourceFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { if err := destFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}
	
	// Copy file permissions
	if info, err := os.Stat(src); err == nil {
		return os.Chmod(dst, info.Mode())
	}
	
	return nil

// GetCurrentVersion returns the current version of the tool
func GetCurrentVersion() (string, error) {
	// This would be set at build time
	// For now, return a placeholder
	return "1.0.0", nil

// Placeholder implementations for methods that will be implemented in other files

func (m *UpdateManager) selectBinaryAsset(assets []ReleaseAsset) *ReleaseAsset {
	// Implementation will be in binary_updater.go
	return nil

func (m *UpdateManager) getTemplateVersions() map[string]string {
	// Implementation will be in template_updater.go
	return make(map[string]string)

func (m *UpdateManager) getModuleVersions() map[string]string {
	// Implementation will be in module_updater.go
	return make(map[string]string)

func (m *UpdateManager) updateFromGitHub(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	// Implementation will be in repository_updater.go
	return nil, nil

func (m *UpdateManager) updateFromGitLab(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	// Implementation will be in repository_updater.go
	return nil, nil

func (m *UpdateManager) updateFromHTTP(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	// Implementation will be in repository_updater.go
	return nil, nil

func (m *UpdateManager) updateFromLocal(ctx context.Context, repo RepositoryConfig, targetDir string) ([]string, error) {
	// Implementation will be in repository_updater.go
	return nil, nil

func formatVersionMap(versions map[string]string) string {
	if len(versions) == 0 {
		return "none"
	}
	
	result := ""
	for name, version := range versions {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("%s:%s", name, version)
	}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
