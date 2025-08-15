package update

import (
	"fmt"
	"runtime"
	"strings"
)

// Installer handles installation of updates
type Installer struct {
	config *Config
	logger Logger

// NewInstaller creates a new installer
func NewInstaller(config *Config, logger Logger) *Installer {
	return &Installer{
		config: config,
		logger: logger,
	}

// InstallBinary installs a new binary
func (i *Installer) InstallBinary(binaryPath string) error {
	i.logger.Info("Installing new binary...")
	
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Create backup if enabled
	if i.config.BackupEnabled {
		if err := i.createBackup(execPath); err != nil {
			i.logger.Warn("Failed to create backup: " + err.Error())
		}
	}
	
	// Install based on platform
	switch runtime.GOOS {
	case "windows":
		return i.installBinaryWindows(binaryPath, execPath)
	default:
		return i.installBinaryUnix(binaryPath, execPath)
	}

// installBinaryWindows installs binary on Windows
func (i *Installer) installBinaryWindows(newBinary, targetPath string) error {
	// Windows: move current to .old, then copy new binary
	oldPath := targetPath + ".old"
	
	// Remove any existing .old file
	if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
		i.logger.Warn("Failed to remove old backup: " + err.Error())
	}
	
	// Move current binary to .old
	if err := os.Rename(targetPath, oldPath); err != nil {
		return fmt.Errorf("failed to move current binary: %w", err)
	}
	
	// Copy new binary
	if err := i.copyFile(newBinary, targetPath); err != nil {
		// Try to restore old binary
		if restoreErr := os.Rename(oldPath, targetPath); restoreErr != nil {
			i.logger.Error("Failed to restore old binary after installation failure", restoreErr)
		}
		return fmt.Errorf("failed to install new binary: %w", err)
	}
	
	i.logger.Info("Binary installation completed (restart required)")
	return nil

// installBinaryUnix installs binary on Unix systems
func (i *Installer) installBinaryUnix(newBinary, targetPath string) error {
	// Unix: atomic replacement using rename
	tempPath := targetPath + ".new"
	
	// Copy new binary to temporary location
	if err := i.copyFile(newBinary, tempPath); err != nil {
		return fmt.Errorf("failed to copy new binary: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Remove(tempPath) // Cleanup
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	
	i.logger.Info("Binary installation completed")
	return nil

// InstallTemplates installs template updates
func (i *Installer) InstallTemplates(templateFiles map[string]string) error {
	i.logger.Info(fmt.Sprintf("Installing %d template updates...", len(templateFiles)))
	
	templateDir := i.config.TemplateDirectory
	if err := os.MkdirAll(templateDir, 0700); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}
	
	// Install each template
	for relPath, sourcePath := range templateFiles {
		targetPath := filepath.Join(templateDir, relPath)
		
		// Create target directory
		if err := os.MkdirAll(filepath.Dir(targetPath), 0700); err != nil {
			return fmt.Errorf("failed to create template subdirectory: %w", err)
		}
		
		// Copy template file
		if err := i.copyFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to install template %s: %w", relPath, err)
		}
		
		i.logger.Debug(fmt.Sprintf("Installed template: %s", relPath))
	}
	
	i.logger.Info("Template installation completed")
	return nil

// InstallModules installs module updates
func (i *Installer) InstallModules(moduleFiles map[string]string) error {
	i.logger.Info(fmt.Sprintf("Installing %d module updates...", len(moduleFiles)))
	
	moduleDir := i.config.ModuleDirectory
	if err := os.MkdirAll(moduleDir, 0700); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}
	
	// Install each module
	for relPath, sourcePath := range moduleFiles {
		targetPath := filepath.Join(moduleDir, relPath)
		
		// Create target directory
		if err := os.MkdirAll(filepath.Dir(targetPath), 0700); err != nil {
			return fmt.Errorf("failed to create module subdirectory: %w", err)
		}
		
		// Copy module file
		if err := i.copyFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to install module %s: %w", relPath, err)
		}
		
		// Make executable if it's a binary module
		if i.isBinaryModule(sourcePath) {
			if err := os.Chmod(targetPath, 0700); err != nil {
				i.logger.Warn(fmt.Sprintf("Failed to make module executable: %s", relPath))
			}
		}
		
		i.logger.Debug(fmt.Sprintf("Installed module: %s", relPath))
	}
	
	i.logger.Info("Module installation completed")
	return nil

// RemoveObsoleteFiles removes files that are no longer needed
func (i *Installer) RemoveObsoleteFiles(obsoleteFiles []string, baseDir string) error {
	if len(obsoleteFiles) == 0 {
		return nil
	}
	
	i.logger.Info(fmt.Sprintf("Removing %d obsolete files...", len(obsoleteFiles)))
	
	for _, relPath := range obsoleteFiles {
		fullPath := filepath.Join(baseDir, relPath)
		
		if err := os.Remove(fullPath); err != nil {
			if !os.IsNotExist(err) {
				i.logger.Warn(fmt.Sprintf("Failed to remove obsolete file %s: %v", relPath, err))
			}
		} else {
			i.logger.Debug(fmt.Sprintf("Removed obsolete file: %s", relPath))
		}
	}
	
	// Remove empty directories
	i.removeEmptyDirectories(baseDir)
	
	return nil

// createBackup creates a backup of a file
func (i *Installer) createBackup(filePath string) error {
	backupDir := i.config.BackupDirectory
	if backupDir == "" {
		backupDir = filepath.Join(filepath.Dir(filePath), "backups")
	}
	
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s_%s%s",
		strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)),
		timestamp,
		filepath.Ext(filePath))
	
	backupPath := filepath.Join(backupDir, backupName)
	
	if err := i.copyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	i.logger.Debug(fmt.Sprintf("Created backup: %s", backupPath))
	return nil

// copyFile copies a file from source to destination
func (i *Installer) copyFile(src, dst string) error {
	sourceFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { if err := sourceFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	// Get source file info
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	
	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { if err := destFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	// Copy content
		if _, err := sourceFile.WriteTo(destFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	
	// Copy permissions
	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	
	return nil

// isBinaryModule checks if a file is a binary module
func (i *Installer) isBinaryModule(filePath string) bool {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExtensions := []string{".so", ".dll", ".dylib", ".exe"}
	
	for _, binExt := range binaryExtensions {
		if ext == binExt {
			return true
		}
	}
	
	// Check if file is executable (Unix)
	if runtime.GOOS != "windows" {
		if info, err := os.Stat(filePath); err == nil {
			return info.Mode()&0111 != 0
		}
	}
	
	return false

// removeEmptyDirectories removes empty directories recursively
func (i *Installer) removeEmptyDirectories(baseDir string) {
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() && path != baseDir {
			// Check if directory is empty
			entries, err := os.ReadDir(path)
			if err == nil && len(entries) == 0 {
				if err := os.Remove(path); err == nil {
					i.logger.Debug(fmt.Sprintf("Removed empty directory: %s", path))
				}
			}
		}
		
		return nil
	})

// ValidateInstallation validates that an installation was successful
func (i *Installer) ValidateInstallation(component string) error {
	switch component {
	case ComponentBinary:
		return i.validateBinaryInstallation()
	case ComponentTemplates:
		return i.validateTemplateInstallation()
	case ComponentModules:
		return i.validateModuleInstallation()
	default:
		return fmt.Errorf("unknown component: %s", component)
	}

// validateBinaryInstallation validates binary installation
func (i *Installer) validateBinaryInstallation() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Check if binary exists and is executable
	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}
	
	// Check permissions
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return fmt.Errorf("binary is not executable")
	}
	
	return nil

// validateTemplateInstallation validates template installation
func (i *Installer) validateTemplateInstallation() error {
	templateDir := i.config.TemplateDirectory
	
	// Check if template directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory not found: %s", templateDir)
	}
	
	// Count template files
	templateCount := 0
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			templateCount++
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to validate templates: %w", err)
	}
	
	if templateCount == 0 {
		return fmt.Errorf("no template files found")
	}
	
	i.logger.Debug(fmt.Sprintf("Validated %d template files", templateCount))
	return nil

// validateModuleInstallation validates module installation
func (i *Installer) validateModuleInstallation() error {
	moduleDir := i.config.ModuleDirectory
	
	// Check if module directory exists
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return fmt.Errorf("module directory not found: %s", moduleDir)
	}
	
	// Count module files
	moduleCount := 0
	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() {
			moduleCount++
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to validate modules: %w", err)
	}
	
	i.logger.Debug(fmt.Sprintf("Validated %d module files", moduleCount))
	return nil
	

// CleanupInstallation cleans up installation artifacts
func (i *Installer) CleanupInstallation() error {
	i.logger.Debug("Cleaning up installation artifacts...")
	
	// Clean up temporary files
	tempDirs := []string{
		filepath.Join(os.TempDir(), "LLMrecon-updates"),
		filepath.Join(os.TempDir(), "LLMrecon-install"),
	}
	
	for _, dir := range tempDirs {
		if err := os.RemoveAll(dir); err != nil {
			i.logger.Warn(fmt.Sprintf("Failed to cleanup %s: %v", dir, err))
		}
	}
	
	// Clean up .old files (Windows)
	if runtime.GOOS == "windows" {
		execPath, err := os.Executable()
		if err == nil {
			oldPath := execPath + ".old"
			if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
				i.logger.Warn(fmt.Sprintf("Failed to cleanup old binary: %v", err))
			}
		}
	}
	
	return nil

// GetInstallationInfo returns information about the current installation
func (i *Installer) GetInstallationInfo() (*InstallationInfo, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	
	info := &InstallationInfo{
		BinaryPath:    execPath,
		TemplateDir:   i.config.TemplateDirectory,
		ModuleDir:     i.config.ModuleDirectory,
		BackupDir:     i.config.BackupDirectory,
		Platform:      runtime.GOOS,
		Architecture:  runtime.GOARCH,
	}
	
	// Get binary info
	if stat, err := os.Stat(execPath); err == nil {
		info.BinarySize = stat.Size()
		info.BinaryModTime = stat.ModTime()
	}
	
	// Count templates
	if templateCount, err := i.countFiles(i.config.TemplateDirectory, []string{".yaml", ".yml"}); err == nil {
		info.TemplateCount = templateCount
	}
	
	// Count modules
	if moduleCount, err := i.countFiles(i.config.ModuleDirectory, nil); err == nil {
		info.ModuleCount = moduleCount
	}
	
	return info, nil

// countFiles counts files in a directory with optional extension filter
func (i *Installer) countFiles(dir string, extensions []string) (int, error) {
	count := 0
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			return nil
		}
		
		if extensions == nil {
			count++
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		for _, validExt := range extensions {
			if ext == validExt {
				count++
				break
			}
		}
		
		return nil
	})
	
	return count, err

// InstallationInfo contains information about the current installation
type InstallationInfo struct {
	BinaryPath     string    `json:"binary_path"`
	BinarySize     int64     `json:"binary_size"`
	BinaryModTime  time.Time `json:"binary_mod_time"`
	TemplateDir    string    `json:"template_dir"`
	TemplateCount  int       `json:"template_count"`
	ModuleDir      string    `json:"module_dir"`
	ModuleCount    int       `json:"module_count"`
	BackupDir      string    `json:"backup_dir"`
	Platform       string    `json:"platform"`
