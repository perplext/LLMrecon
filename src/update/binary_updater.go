package update

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// BinaryUpdater handles self-updating of the binary
type BinaryUpdater struct {
	config   *Config
	verifier *Verifier
	logger   Logger
}

// NewBinaryUpdater creates a new binary updater
func NewBinaryUpdater(config *Config, verifier *Verifier, logger Logger) *BinaryUpdater {
	return &BinaryUpdater{
		config:   config,
		verifier: verifier,
		logger:   logger,
	}
}

// UpdateBinary performs a self-update of the binary
func (bu *BinaryUpdater) UpdateBinary(ctx context.Context, release *Release) error {
	bu.logger.Info(fmt.Sprintf("Starting binary update to version %s", release.Version))
	
	// Check if update is needed
	currentVersion, err := GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}
	
	if !bu.isUpdateNeeded(currentVersion, release.Version) {
		bu.logger.Info("Binary is already up to date")
		return nil
	}
	
	// Find appropriate asset for current platform
	asset := bu.selectBinaryAsset(release.Assets)
	if asset == nil {
		return fmt.Errorf("no compatible binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	
	bu.logger.Info(fmt.Sprintf("Found asset: %s (%s)", asset.Name, FormatFileSize(asset.Size)))
	
	// Create update workspace
	updateDir, err := bu.createUpdateWorkspace()
	if err != nil {
		return fmt.Errorf("failed to create update workspace: %w", err)
	}
	defer bu.cleanupWorkspace(updateDir)
	
	// Download new binary
	newBinaryPath, err := bu.downloadBinary(ctx, asset, updateDir)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	
	// Verify downloaded binary
	if err := bu.verifyBinary(newBinaryPath, asset); err != nil {
		return fmt.Errorf("binary verification failed: %w", err)
	}
	
	// Create backup of current binary
	backupPath, err := bu.createBinaryBackup()
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	// Apply the update
	if err := bu.applyBinaryUpdate(newBinaryPath, backupPath); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}
	
	bu.logger.Info("Binary update completed successfully")
	bu.logger.Info("Please restart the application to use the new version")
	
	return nil
}

// isUpdateNeeded checks if an update is needed
func (bu *BinaryUpdater) isUpdateNeeded(current, latest string) bool {
	currentVer, err := semver.NewVersion(strings.TrimPrefix(current, "v"))
	if err != nil {
		bu.logger.Error("Invalid current version", err)
		return false
	}
	
	latestVer, err := semver.NewVersion(strings.TrimPrefix(latest, "v"))
	if err != nil {
		bu.logger.Error("Invalid latest version", err)
		return false
	}
	
	return latestVer.GreaterThan(currentVer)
}

// selectBinaryAsset selects the appropriate binary asset for the current platform
func (bu *BinaryUpdater) selectBinaryAsset(assets []ReleaseAsset) *ReleaseAsset {
	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH
	
	// Platform mapping
	platformMap := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux",
		"windows": "windows",
		"freebsd": "freebsd",
	}
	
	// Architecture mapping
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
		"arm":   "arm",
	}
	
	platform, platformExists := platformMap[currentOS]
	arch, archExists := archMap[currentArch]
	
	if !platformExists || !archExists {
		bu.logger.Error(fmt.Sprintf("Unsupported platform: %s/%s", currentOS, currentArch), nil)
		return nil
	}
	
	// First pass: exact match on platform and architecture
	for _, asset := range assets {
		if asset.Platform == platform && asset.Architecture == arch {
			return &asset
		}
	}
	
	// Second pass: name-based matching
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		
		// Check for platform and architecture in filename
		hasPlatform := strings.Contains(name, platform)
		hasArch := strings.Contains(name, arch)
		
		// Special handling for different naming conventions
		if currentOS == "darwin" && strings.Contains(name, "macos") {
			hasPlatform = true
		}
		if currentArch == "amd64" && (strings.Contains(name, "x64") || strings.Contains(name, "x86_64")) {
			hasArch = true
		}
		if currentArch == "386" && strings.Contains(name, "x86") {
			hasArch = true
		}
		
		if hasPlatform && hasArch {
			// Update asset metadata
			asset.Platform = platform
			asset.Architecture = arch
			return &asset
		}
	}
	
	bu.logger.Error(fmt.Sprintf("No compatible binary found for %s/%s", platform, arch), nil)
	return nil
}

// createUpdateWorkspace creates a temporary workspace for the update
func (bu *BinaryUpdater) createUpdateWorkspace() (string, error) {
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("LLMrecon-update-%d", time.Now().Unix()))
	
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create update workspace: %w", err)
	}
	
	bu.logger.Debug(fmt.Sprintf("Created update workspace: %s", tempDir))
	return tempDir, nil
}

// downloadBinary downloads the new binary to the workspace
func (bu *BinaryUpdater) downloadBinary(ctx context.Context, asset *ReleaseAsset, workspace string) (string, error) {
	downloader := NewUpdateDownloader(bu.config, bu.logger)
	
	// Download with progress tracking
	progressCallback := func(progress *DownloadProgress) {
		if progress.TotalBytes > 0 {
			percentage := float64(progress.DownloadedBytes) / float64(progress.TotalBytes) * 100
			bu.logger.Info(fmt.Sprintf("Download progress: %.1f%% (%s/%s) - %s/s",
				percentage,
				FormatFileSize(progress.DownloadedBytes),
				FormatFileSize(progress.TotalBytes),
				FormatFileSize(int64(progress.Speed))))
		}
	}
	
	downloadPath, err := downloader.DownloadFileWithProgress(ctx, asset.DownloadURL, asset.Name, progressCallback)
	if err != nil {
		return "", err
	}
	
	// Move to workspace
	workspacePath := filepath.Join(workspace, asset.Name)
	if err := os.Rename(downloadPath, workspacePath); err != nil {
		return "", fmt.Errorf("failed to move binary to workspace: %w", err)
	}
	
	return workspacePath, nil
}

// verifyBinary verifies the downloaded binary
func (bu *BinaryUpdater) verifyBinary(binaryPath string, asset *ReleaseAsset) error {
	bu.logger.Info("Verifying downloaded binary...")
	
	// Verify file size
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to stat binary: %w", err)
	}
	
	if info.Size() != asset.Size {
		return fmt.Errorf("binary size mismatch: expected %d, got %d", asset.Size, info.Size())
	}
	
	// Verify checksum if available
	if asset.Checksum != "" {
		if err := bu.verifier.VerifyFile(binaryPath, asset.Checksum, asset.SignatureURL); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}
	
	// Make binary executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}
	
	// Basic functionality test (try to run --version)
	if err := bu.testBinary(binaryPath); err != nil {
		bu.logger.Warn("Binary functionality test failed (proceeding anyway): " + err.Error())
	}
	
	bu.logger.Info("Binary verification completed successfully")
	return nil
}

// testBinary performs a basic functionality test on the new binary
func (bu *BinaryUpdater) testBinary(binaryPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("binary test failed: %w", err)
	}
	
	if len(output) == 0 {
		return fmt.Errorf("binary test returned no output")
	}
	
	bu.logger.Debug(fmt.Sprintf("Binary test output: %s", strings.TrimSpace(string(output))))
	return nil
}

// createBinaryBackup creates a backup of the current binary
func (bu *BinaryUpdater) createBinaryBackup() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Create backup directory
	backupDir := bu.config.BackupDirectory
	if backupDir == "" {
		backupDir = filepath.Join(filepath.Dir(execPath), "backups")
	}
	
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s_%s%s", 
		strings.TrimSuffix(filepath.Base(execPath), filepath.Ext(execPath)),
		timestamp,
		filepath.Ext(execPath))
	
	backupPath := filepath.Join(backupDir, backupName)
	
	// Copy current binary to backup
	if err := bu.copyFile(execPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	
	bu.logger.Info(fmt.Sprintf("Created backup: %s", backupPath))
	return backupPath, nil
}

// applyBinaryUpdate applies the binary update
func (bu *BinaryUpdater) applyBinaryUpdate(newBinaryPath, backupPath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	bu.logger.Info("Applying binary update...")
	
	// On Windows, we might need special handling due to file locking
	if runtime.GOOS == "windows" {
		return bu.applyWindowsUpdate(newBinaryPath, execPath, backupPath)
	}
	
	// Unix systems: replace the binary directly
	return bu.applyUnixUpdate(newBinaryPath, execPath, backupPath)
}

// applyWindowsUpdate applies update on Windows (handles file locking)
func (bu *BinaryUpdater) applyWindowsUpdate(newBinaryPath, execPath, backupPath string) error {
	// Windows strategy: rename current binary to .old, copy new binary in place
	oldPath := execPath + ".old"
	
	// Remove any existing .old file
	os.Remove(oldPath)
	
	// Rename current binary
	if err := os.Rename(execPath, oldPath); err != nil {
		return fmt.Errorf("failed to rename current binary: %w", err)
	}
	
	// Copy new binary in place
	if err := bu.copyFile(newBinaryPath, execPath); err != nil {
		// Rollback: restore original binary
		os.Rename(oldPath, execPath)
		return fmt.Errorf("failed to copy new binary: %w", err)
	}
	
	// Schedule cleanup of old binary (will happen on next restart)
	bu.scheduleCleanup(oldPath)
	
	return nil
}

// applyUnixUpdate applies update on Unix systems
func (bu *BinaryUpdater) applyUnixUpdate(newBinaryPath, execPath, backupPath string) error {
	// Create a temporary file next to the executable
	tempPath := execPath + ".tmp"
	
	// Copy new binary to temporary location
	if err := bu.copyFile(newBinaryPath, tempPath); err != nil {
		return fmt.Errorf("failed to copy new binary: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempPath, execPath); err != nil {
		os.Remove(tempPath) // Cleanup
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	
	return nil
}

// copyFile copies a file from src to dst
func (bu *BinaryUpdater) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	// Copy content
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}
	
	// Set permissions
	return os.Chmod(dst, sourceInfo.Mode())
}

// scheduleCleanup schedules cleanup of old files
func (bu *BinaryUpdater) scheduleCleanup(filePath string) {
	// On Windows, create a batch script to delete the old file
	if runtime.GOOS == "windows" {
		batchContent := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
del "%s" >nul 2>&1
del "%%~f0" >nul 2>&1`, filePath)
		
		batchPath := filePath + "_cleanup.bat"
		if err := os.WriteFile(batchPath, []byte(batchContent), 0644); err == nil {
			go func() {
				exec.Command("cmd", "/C", batchPath).Start()
			}()
		}
	}
}

// cleanupWorkspace removes the update workspace
func (bu *BinaryUpdater) cleanupWorkspace(workspace string) {
	if err := os.RemoveAll(workspace); err != nil {
		bu.logger.Error("Failed to cleanup workspace", err)
	} else {
		bu.logger.Debug("Cleaned up update workspace")
	}
}

// RestartApplication restarts the application with the new binary
func (bu *BinaryUpdater) RestartApplication() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Get current arguments
	args := os.Args[1:]
	
	bu.logger.Info("Restarting application with new binary...")
	
	// Start new process
	cmd := exec.Command(execPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new process: %w", err)
	}
	
	// Exit current process
	os.Exit(0)
	return nil
}

// RollbackUpdate rolls back to the previous version
func (bu *BinaryUpdater) RollbackUpdate(backupPath string) error {
	if backupPath == "" {
		return fmt.Errorf("no backup path provided")
	}
	
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}
	
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	bu.logger.Info(fmt.Sprintf("Rolling back to backup: %s", backupPath))
	
	// Copy backup over current binary
	if err := bu.copyFile(backupPath, execPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}
	
	bu.logger.Info("Rollback completed successfully")
	return nil
}

// CanSelfUpdate checks if self-update is possible
func (bu *BinaryUpdater) CanSelfUpdate() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	
	// Check if we can write to the executable directory
	execDir := filepath.Dir(execPath)
	testFile := filepath.Join(execDir, ".update_test")
	
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("insufficient permissions to update binary: %w", err)
	}
	
	os.Remove(testFile)
	return nil
}

// GetUpdatePermissions checks what permissions are needed for update
func (bu *BinaryUpdater) GetUpdatePermissions() *UpdatePermissions {
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	
	permissions := &UpdatePermissions{
		CanWriteExecutable: true,
		CanWriteDirectory:  true,
		RequiresElevation:  false,
		Platform:          runtime.GOOS,
	}
	
	// Test write permissions
	testFile := filepath.Join(execDir, ".perm_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		permissions.CanWriteDirectory = false
		permissions.RequiresElevation = true
	} else {
		os.Remove(testFile)
	}
	
	// On Unix, check if we're running as root
	if runtime.GOOS != "windows" {
		if os.Geteuid() == 0 {
			permissions.RunningAsRoot = true
		}
	}
	
	return permissions
}

// UpdatePermissions represents the permissions available for updating
type UpdatePermissions struct {
	CanWriteExecutable bool
	CanWriteDirectory  bool
	RequiresElevation  bool
	RunningAsRoot      bool
	Platform           string
}

// ElevatePermissions attempts to elevate permissions for update
func (bu *BinaryUpdater) ElevatePermissions(args []string) error {
	switch runtime.GOOS {
	case "windows":
		return bu.elevateWindows(args)
	case "darwin":
		return bu.elevateDarwin(args)
	case "linux":
		return bu.elevateLinux(args)
	default:
		return fmt.Errorf("permission elevation not supported on %s", runtime.GOOS)
	}
}

// elevateWindows elevates permissions on Windows using UAC
func (bu *BinaryUpdater) elevateWindows(args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Use PowerShell to elevate
	psScript := fmt.Sprintf(`Start-Process -FilePath "%s" -ArgumentList "%s" -Verb RunAs`, 
		execPath, strings.Join(args, " "))
	
	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run()
}

// elevateDarwin elevates permissions on macOS using osascript
func (bu *BinaryUpdater) elevateDarwin(args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	
	script := fmt.Sprintf(`do shell script "%s %s" with administrator privileges`, 
		execPath, strings.Join(args, " "))
	
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// elevateLinux elevates permissions on Linux using sudo
func (bu *BinaryUpdater) elevateLinux(args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	
	sudoArgs := append([]string{execPath}, args...)
	cmd := exec.Command("sudo", sudoArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

// ValidateUpdate validates that an update was successful
func (bu *BinaryUpdater) ValidateUpdate(expectedVersion string) error {
	// Run the updated binary to get its version
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, execPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to validate update: %w", err)
	}
	
	outputStr := strings.TrimSpace(string(output))
	if !strings.Contains(outputStr, expectedVersion) {
		return fmt.Errorf("version validation failed: expected %s, got %s", expectedVersion, outputStr)
	}
	
	return nil
}

// Helper function to check if running with elevated privileges
func isElevated() bool {
	switch runtime.GOOS {
	case "windows":
		// Check if running as administrator
		cmd := exec.Command("net", "session")
		err := cmd.Run()
		return err == nil
	default:
		// Check if running as root
		return os.Geteuid() == 0
	}
}

// Helper function to request elevation
func requestElevation() error {
	if isElevated() {
		return nil
	}
	
	fmt.Println("This operation requires elevated privileges.")
	fmt.Println("Please run with administrator/root privileges or use sudo.")
	
	return fmt.Errorf("insufficient privileges")
}