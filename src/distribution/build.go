package distribution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// BuildPipelineImpl implements the BuildPipeline interface
type BuildPipelineImpl struct {
	config        *Config
	logger        Logger
	storage       ArtifactStorage
	signer        CodeSigner
	
	// Build state
	builds        map[string]*BuildExecution
	buildsMutex   sync.RWMutex
	
	// Worker pool
	workerPool    chan struct{}
	workers       int
}

// BuildExecution tracks a build in progress
type BuildExecution struct {
	ID          string                `json:"id"`
	Version     string                `json:"version"`
	Targets     []BuildTarget         `json:"targets"`
	Status      BuildState            `json:"status"`
	Progress    int                   `json:"progress"`
	StartTime   time.Time             `json:"start_time"`
	EndTime     *time.Time            `json:"end_time,omitempty"`
	Artifacts   []BuildArtifact       `json:"artifacts"`
	Logs        []BuildLogEntry       `json:"logs"`
	Errors      []BuildError          `json:"errors"`
	Context     context.Context       `json:"-"`
	Cancel      context.CancelFunc    `json:"-"`
	Mutex       sync.RWMutex          `json:"-"`
}

// BuildLogEntry represents a build log entry
type BuildLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Target    string    `json:"target,omitempty"`
	Message   string    `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// CodeSigner interface for signing artifacts
type CodeSigner interface {
	Sign(ctx context.Context, artifact *BuildArtifact, config PlatformSigningConfig) error
	Verify(ctx context.Context, artifact *BuildArtifact) error
	GetCertificateInfo() (*CertificateInfo, error)
}

// CertificateInfo contains signing certificate information
type CertificateInfo struct {
	Subject    string    `json:"subject"`
	Issuer     string    `json:"issuer"`
	NotBefore  time.Time `json:"not_before"`
	NotAfter   time.Time `json:"not_after"`
	Serial     string    `json:"serial"`
	Thumbprint string    `json:"thumbprint"`
}

// NewBuildPipeline creates a new build pipeline
func NewBuildPipeline(config *Config, logger Logger) BuildPipeline {
	storage := NewArtifactStorage(config.ArtifactStorage, logger)
	signer := NewCodeSigner(config.SigningConfig, logger)
	
	return &BuildPipelineImpl{
		config:     config,
		logger:     logger,
		storage:    storage,
		signer:     signer,
		builds:     make(map[string]*BuildExecution),
		workerPool: make(chan struct{}, 4), // Default 4 concurrent builds
		workers:    4,
	}
}

// Build starts a new build for the specified targets
func (bp *BuildPipelineImpl) Build(ctx context.Context, version string, targets []BuildTarget) (*BuildResult, error) {
	buildID := generateBuildID()
	
	bp.logger.Info("Starting build", "buildID", buildID, "version", version, "targets", len(targets))
	
	// Create build execution context
	buildCtx, cancel := context.WithCancel(ctx)
	execution := &BuildExecution{
		ID:        buildID,
		Version:   version,
		Targets:   targets,
		Status:    BuildStatePending,
		Progress:  0,
		StartTime: time.Now(),
		Artifacts: make([]BuildArtifact, 0),
		Logs:      make([]BuildLogEntry, 0),
		Errors:    make([]BuildError, 0),
		Context:   buildCtx,
		Cancel:    cancel,
	}
	
	bp.buildsMutex.Lock()
	bp.builds[buildID] = execution
	bp.buildsMutex.Unlock()
	
	// Start build asynchronously
	go bp.executeBuild(execution)
	
	// Return immediate result with build ID
	return &BuildResult{
		BuildID:   buildID,
		Version:   version,
		Status:    BuildStatePending,
		StartTime: execution.StartTime,
	}, nil
}

// GetBuildStatus returns the current status of a build
func (bp *BuildPipelineImpl) GetBuildStatus(buildID string) (*BuildStatus, error) {
	bp.buildsMutex.RLock()
	execution, exists := bp.builds[buildID]
	bp.buildsMutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("build not found: %s", buildID)
	}
	
	execution.Mutex.RLock()
	defer execution.Mutex.RUnlock()
	
	return &BuildStatus{
		BuildID:   buildID,
		Status:    execution.Status,
		Progress:  execution.Progress,
		Message:   bp.getStatusMessage(execution),
		UpdatedAt: time.Now(),
	}, nil
}

// ListBuilds returns a list of builds matching the filters
func (bp *BuildPipelineImpl) ListBuilds(ctx context.Context, filters BuildFilters) ([]BuildInfo, error) {
	bp.buildsMutex.RLock()
	defer bp.buildsMutex.RUnlock()
	
	var builds []BuildInfo
	
	for _, execution := range bp.builds {
		execution.Mutex.RLock()
		
		// Apply filters
		if len(filters.Status) > 0 && !contains(filters.Status, execution.Status) {
			execution.Mutex.RUnlock()
			continue
		}
		
		if filters.Version != "" && execution.Version != filters.Version {
			execution.Mutex.RUnlock()
			continue
		}
		
		if filters.StartDate != nil && execution.StartTime.Before(*filters.StartDate) {
			execution.Mutex.RUnlock()
			continue
		}
		
		if filters.EndDate != nil && execution.EndTime != nil && execution.EndTime.After(*filters.EndDate) {
			execution.Mutex.RUnlock()
			continue
		}
		
		duration := time.Duration(0)
		if execution.EndTime != nil {
			duration = execution.EndTime.Sub(execution.StartTime)
		}
		
		builds = append(builds, BuildInfo{
			BuildID:   execution.ID,
			Version:   execution.Version,
			Status:    execution.Status,
			Targets:   execution.Targets,
			StartTime: execution.StartTime,
			Duration:  duration,
		})
		
		execution.Mutex.RUnlock()
	}
	
	// Apply pagination
	if filters.Limit > 0 {
		start := filters.Offset
		end := start + filters.Limit
		if start < len(builds) {
			if end > len(builds) {
				end = len(builds)
			}
			builds = builds[start:end]
		} else {
			builds = []BuildInfo{}
		}
	}
	
	return builds, nil
}

// CleanupBuilds removes old build records
func (bp *BuildPipelineImpl) CleanupBuilds(ctx context.Context, olderThan time.Time) error {
	bp.buildsMutex.Lock()
	defer bp.buildsMutex.Unlock()
	
	var cleaned int
	for id, execution := range bp.builds {
		if execution.StartTime.Before(olderThan) {
			// Cancel if still running
			if execution.Status == BuildStateRunning {
				execution.Cancel()
			}
			delete(bp.builds, id)
			cleaned++
		}
	}
	
	bp.logger.Info("Cleaned up builds", "count", cleaned, "olderThan", olderThan)
	
	return nil
}

// Internal methods

func (bp *BuildPipelineImpl) executeBuild(execution *BuildExecution) {
	defer func() {
		if r := recover(); r != nil {
			bp.logger.Error("Build execution panicked", "buildID", execution.ID, "panic", r)
			bp.updateBuildStatus(execution, BuildStateFailed, 0, fmt.Sprintf("Build panicked: %v", r))
		}
	}()
	
	bp.updateBuildStatus(execution, BuildStateRunning, 0, "Build started")
	
	// Create temporary build directory
	buildDir, err := bp.createBuildDirectory(execution.ID)
	if err != nil {
		bp.logger.Error("Failed to create build directory", "buildID", execution.ID, "error", err)
		bp.updateBuildStatus(execution, BuildStateFailed, 0, fmt.Sprintf("Failed to create build directory: %v", err))
		return
	}
	defer os.RemoveAll(buildDir)
	
	// Build each target
	totalTargets := len(execution.Targets)
	completedTargets := 0
	
	// Use semaphore to limit concurrent builds
	for i, target := range execution.Targets {
		select {
		case <-execution.Context.Done():
			bp.updateBuildStatus(execution, BuildStateCancelled, (completedTargets*100)/totalTargets, "Build cancelled")
			return
		case bp.workerPool <- struct{}{}:
			// Got worker slot
		}
		
		bp.logBuild(execution, "info", target.OutputName, fmt.Sprintf("Building target %d/%d: %s/%s", i+1, totalTargets, target.Platform, target.Architecture), nil)
		
		artifact, err := bp.buildTarget(execution.Context, buildDir, execution.Version, target)
		if err != nil {
			bp.logBuild(execution, "error", target.OutputName, fmt.Sprintf("Build failed: %v", err), map[string]interface{}{"error": err.Error()})
			execution.Mutex.Lock()
			execution.Errors = append(execution.Errors, BuildError{
				Target:  target,
				Error:   err.Error(),
				Stage:   "build",
			})
			execution.Mutex.Unlock()
		} else {
			bp.logBuild(execution, "info", target.OutputName, "Build completed successfully", nil)
			execution.Mutex.Lock()
			execution.Artifacts = append(execution.Artifacts, *artifact)
			execution.Mutex.Unlock()
		}
		
		<-bp.workerPool // Release worker slot
		completedTargets++
		
		progress := (completedTargets * 100) / totalTargets
		bp.updateBuildStatus(execution, BuildStateRunning, progress, fmt.Sprintf("Completed %d/%d targets", completedTargets, totalTargets))
	}
	
	// Determine final status
	execution.Mutex.RLock()
	hasErrors := len(execution.Errors) > 0
	artifactCount := len(execution.Artifacts)
	execution.Mutex.RUnlock()
	
	if hasErrors && artifactCount == 0 {
		bp.updateBuildStatus(execution, BuildStateFailed, 100, "All builds failed")
	} else if hasErrors {
		bp.updateBuildStatus(execution, BuildStateCompleted, 100, fmt.Sprintf("Build completed with %d errors", len(execution.Errors)))
	} else {
		bp.updateBuildStatus(execution, BuildStateCompleted, 100, "Build completed successfully")
	}
	
	bp.logger.Info("Build execution completed", "buildID", execution.ID, "artifacts", artifactCount, "errors", len(execution.Errors))
}

func (bp *BuildPipelineImpl) buildTarget(ctx context.Context, buildDir, version string, target BuildTarget) (*BuildArtifact, error) {
	// Set up build environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", target.GoOS))
	env = append(env, fmt.Sprintf("GOARCH=%s", target.GoArch))
	if target.CGOEnabled {
		env = append(env, "CGO_ENABLED=1")
	} else {
		env = append(env, "CGO_ENABLED=0")
	}
	
	// Prepare output filename
	outputName := target.OutputName
	if target.Platform == PlatformWindows && !strings.HasSuffix(outputName, ".exe") {
		outputName += ".exe"
	}
	
	outputPath := filepath.Join(buildDir, outputName)
	
	// Prepare build command
	args := []string{"build"}
	args = append(args, target.BuildFlags...)
	
	// Add version information to ldflags
	ldflags := target.LDFlags
	ldflags = append(ldflags, fmt.Sprintf("-X main.Version=%s", version))
	ldflags = append(ldflags, fmt.Sprintf("-X main.BuildTime=%s", time.Now().Format(time.RFC3339)))
	ldflags = append(ldflags, fmt.Sprintf("-X main.Platform=%s", target.Platform))
	ldflags = append(ldflags, fmt.Sprintf("-X main.Architecture=%s", target.Architecture))
	
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", strings.Join(ldflags, " "))
	}
	
	args = append(args, "-o", outputPath)
	args = append(args, ".") // Build current directory
	
	// Execute build
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = env
	cmd.Dir = "." // Set to project root
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed: %v\nOutput: %s", err, string(output))
	}
	
	// Get file info
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}
	
	// Calculate checksums
	checksums, err := bp.calculateChecksums(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksums: %v", err)
	}
	
	// Create artifact
	artifact := &BuildArtifact{
		ID:           generateArtifactID(),
		Name:         outputName,
		Platform:     target.Platform,
		Architecture: target.Architecture,
		Type:         ArtifactTypeBinary,
		Size:         fileInfo.Size(),
		Checksum:     checksums,
		CreatedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"version":      version,
			"build_flags":  target.BuildFlags,
			"ldflags":      target.LDFlags,
			"cgo_enabled":  target.CGOEnabled,
			"go_version":   runtime.Version(),
		},
	}
	
	// Store artifact
	location, err := bp.storage.Store(ctx, artifact)
	if err != nil {
		return nil, fmt.Errorf("failed to store artifact: %v", err)
	}
	artifact.Location = location
	
	// Sign artifact if signing is enabled
	if bp.config.SigningConfig.Enabled {
		if platformConfig, exists := bp.config.SigningConfig.Platforms[target.Platform]; exists && platformConfig.Enabled {
			if err := bp.signer.Sign(ctx, artifact, platformConfig); err != nil {
				bp.logger.Warn("Failed to sign artifact", "artifact", artifact.Name, "error", err)
			}
		}
	}
	
	// Compress if requested
	if target.Compress {
		compressedArtifact, err := bp.compressArtifact(ctx, artifact)
		if err != nil {
			bp.logger.Warn("Failed to compress artifact", "artifact", artifact.Name, "error", err)
		} else {
			artifact = compressedArtifact
		}
	}
	
	return artifact, nil
}

func (bp *BuildPipelineImpl) createBuildDirectory(buildID string) (string, error) {
	tempDir := os.TempDir()
	buildDir := filepath.Join(tempDir, "LLMrecon-build", buildID)
	
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", err
	}
	
	return buildDir, nil
}

func (bp *BuildPipelineImpl) calculateChecksums(filePath string) (map[string]string, error) {
	checksums := make(map[string]string)
	
	// Calculate MD5
	if md5Sum, err := calculateFileChecksum(filePath, "md5"); err == nil {
		checksums["md5"] = md5Sum
	}
	
	// Calculate SHA1
	if sha1Sum, err := calculateFileChecksum(filePath, "sha1"); err == nil {
		checksums["sha1"] = sha1Sum
	}
	
	// Calculate SHA256
	if sha256Sum, err := calculateFileChecksum(filePath, "sha256"); err == nil {
		checksums["sha256"] = sha256Sum
	}
	
	// Calculate SHA512
	if sha512Sum, err := calculateFileChecksum(filePath, "sha512"); err == nil {
		checksums["sha512"] = sha512Sum
	}
	
	return checksums, nil
}

func (bp *BuildPipelineImpl) compressArtifact(ctx context.Context, artifact *BuildArtifact) (*BuildArtifact, error) {
	// This would implement compression logic
	// For now, just return the original artifact
	return artifact, nil
}

func (bp *BuildPipelineImpl) updateBuildStatus(execution *BuildExecution, status BuildState, progress int, message string) {
	execution.Mutex.Lock()
	execution.Status = status
	execution.Progress = progress
	if status == BuildStateCompleted || status == BuildStateFailed || status == BuildStateCancelled {
		now := time.Now()
		execution.EndTime = &now
	}
	execution.Mutex.Unlock()
	
	bp.logBuild(execution, "info", "", message, map[string]interface{}{
		"status":   string(status),
		"progress": progress,
	})
}

func (bp *BuildPipelineImpl) logBuild(execution *BuildExecution, level, target, message string, details map[string]interface{}) {
	logEntry := BuildLogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Target:    target,
		Message:   message,
		Details:   details,
	}
	
	execution.Mutex.Lock()
	execution.Logs = append(execution.Logs, logEntry)
	execution.Mutex.Unlock()
	
	bp.logger.Info("Build log", "buildID", execution.ID, "level", level, "target", target, "message", message)
}

func (bp *BuildPipelineImpl) getStatusMessage(execution *BuildExecution) string {
	switch execution.Status {
	case BuildStatePending:
		return "Build queued"
	case BuildStateRunning:
		return fmt.Sprintf("Building %d targets (%d%% complete)", len(execution.Targets), execution.Progress)
	case BuildStateCompleted:
		return fmt.Sprintf("Build completed - %d artifacts, %d errors", len(execution.Artifacts), len(execution.Errors))
	case BuildStateFailed:
		return fmt.Sprintf("Build failed - %d errors", len(execution.Errors))
	case BuildStateCancelled:
		return "Build cancelled"
	default:
		return "Unknown status"
	}
}

// Utility functions

func contains(slice []BuildState, item BuildState) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateBuildID() string {
	return fmt.Sprintf("build_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateArtifactID() string {
	return fmt.Sprintf("artifact_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// Mock implementations for dependencies

// NewArtifactStorage creates a new artifact storage instance
func NewArtifactStorage(config ArtifactStorageConfig, logger Logger) ArtifactStorage {
	switch config.Type {
	case StorageTypeS3:
		return &S3Storage{config: config, logger: logger}
	case StorageTypeGCS:
		return &GCSStorage{config: config, logger: logger}
	default:
		return &LocalStorage{config: config, logger: logger}
	}
}

// NewCodeSigner creates a new code signer
func NewCodeSigner(config SigningConfig, logger Logger) CodeSigner {
	if !config.Enabled {
		return &NoOpSigner{}
	}
	return &DefaultCodeSigner{config: config, logger: logger}
}

// Simple storage implementations (mock)
type LocalStorage struct {
	config ArtifactStorageConfig
	logger Logger
}

func (ls *LocalStorage) Store(ctx context.Context, artifact *BuildArtifact) (*StorageLocation, error) {
	return &StorageLocation{
		Provider: "local",
		Key:      artifact.Name,
		URL:      fmt.Sprintf("file://%s", artifact.Name),
	}, nil
}

func (ls *LocalStorage) Retrieve(ctx context.Context, location *StorageLocation, writer io.Writer) error {
	return nil
}

func (ls *LocalStorage) Delete(ctx context.Context, location *StorageLocation) error {
	return nil
}

func (ls *LocalStorage) Exists(ctx context.Context, location *StorageLocation) (bool, error) {
	return true, nil
}

func (ls *LocalStorage) GetMetadata(ctx context.Context, location *StorageLocation) (*ArtifactMetadata, error) {
	return &ArtifactMetadata{}, nil
}

func (ls *LocalStorage) UpdateMetadata(ctx context.Context, location *StorageLocation, metadata *ArtifactMetadata) error {
	return nil
}

func (ls *LocalStorage) List(ctx context.Context, filters StorageFilters) ([]StorageLocation, error) {
	return []StorageLocation{}, nil
}

func (ls *LocalStorage) Search(ctx context.Context, query string) ([]StorageLocation, error) {
	return []StorageLocation{}, nil
}

func (ls *LocalStorage) Cleanup(ctx context.Context, retentionPolicy RetentionPolicy) error {
	return nil
}

func (ls *LocalStorage) GetStorageUsage(ctx context.Context) (*StorageUsage, error) {
	return &StorageUsage{}, nil
}

// S3Storage and GCSStorage would be similar implementations for cloud storage

type S3Storage struct {
	config ArtifactStorageConfig
	logger Logger
}

func (s3 *S3Storage) Store(ctx context.Context, artifact *BuildArtifact) (*StorageLocation, error) {
	return &StorageLocation{Provider: "s3"}, nil
}

func (s3 *S3Storage) Retrieve(ctx context.Context, location *StorageLocation, writer io.Writer) error { return nil }
func (s3 *S3Storage) Delete(ctx context.Context, location *StorageLocation) error { return nil }
func (s3 *S3Storage) Exists(ctx context.Context, location *StorageLocation) (bool, error) { return true, nil }
func (s3 *S3Storage) GetMetadata(ctx context.Context, location *StorageLocation) (*ArtifactMetadata, error) { return &ArtifactMetadata{}, nil }
func (s3 *S3Storage) UpdateMetadata(ctx context.Context, location *StorageLocation, metadata *ArtifactMetadata) error { return nil }
func (s3 *S3Storage) List(ctx context.Context, filters StorageFilters) ([]StorageLocation, error) { return []StorageLocation{}, nil }
func (s3 *S3Storage) Search(ctx context.Context, query string) ([]StorageLocation, error) { return []StorageLocation{}, nil }
func (s3 *S3Storage) Cleanup(ctx context.Context, retentionPolicy RetentionPolicy) error { return nil }
func (s3 *S3Storage) GetStorageUsage(ctx context.Context) (*StorageUsage, error) { return &StorageUsage{}, nil }

type GCSStorage struct {
	config ArtifactStorageConfig
	logger Logger
}

func (gcs *GCSStorage) Store(ctx context.Context, artifact *BuildArtifact) (*StorageLocation, error) {
	return &StorageLocation{Provider: "gcs"}, nil
}

func (gcs *GCSStorage) Retrieve(ctx context.Context, location *StorageLocation, writer io.Writer) error { return nil }
func (gcs *GCSStorage) Delete(ctx context.Context, location *StorageLocation) error { return nil }
func (gcs *GCSStorage) Exists(ctx context.Context, location *StorageLocation) (bool, error) { return true, nil }
func (gcs *GCSStorage) GetMetadata(ctx context.Context, location *StorageLocation) (*ArtifactMetadata, error) { return &ArtifactMetadata{}, nil }
func (gcs *GCSStorage) UpdateMetadata(ctx context.Context, location *StorageLocation, metadata *ArtifactMetadata) error { return nil }
func (gcs *GCSStorage) List(ctx context.Context, filters StorageFilters) ([]StorageLocation, error) { return []StorageLocation{}, nil }
func (gcs *GCSStorage) Search(ctx context.Context, query string) ([]StorageLocation, error) { return []StorageLocation{}, nil }
func (gcs *GCSStorage) Cleanup(ctx context.Context, retentionPolicy RetentionPolicy) error { return nil }
func (gcs *GCSStorage) GetStorageUsage(ctx context.Context) (*StorageUsage, error) { return &StorageUsage{}, nil }

// Code signing implementations
type NoOpSigner struct{}

func (nos *NoOpSigner) Sign(ctx context.Context, artifact *BuildArtifact, config PlatformSigningConfig) error {
	return nil
}

func (nos *NoOpSigner) Verify(ctx context.Context, artifact *BuildArtifact) error {
	return nil
}

func (nos *NoOpSigner) GetCertificateInfo() (*CertificateInfo, error) {
	return &CertificateInfo{}, nil
}

type DefaultCodeSigner struct {
	config SigningConfig
	logger Logger
}

func (dcs *DefaultCodeSigner) Sign(ctx context.Context, artifact *BuildArtifact, config PlatformSigningConfig) error {
	dcs.logger.Info("Signing artifact", "artifact", artifact.Name, "platform", artifact.Platform)
	// Mock signing implementation
	artifact.Signature = &Signature{
		Algorithm: "RSA-SHA256",
		KeyID:     "default-key",
		Signature: "mock-signature",
		Timestamp: time.Now(),
	}
	return nil
}

func (dcs *DefaultCodeSigner) Verify(ctx context.Context, artifact *BuildArtifact) error {
	return nil
}

func (dcs *DefaultCodeSigner) GetCertificateInfo() (*CertificateInfo, error) {
	return &CertificateInfo{
		Subject:   "CN=LLMrecon",
		Issuer:    "CN=LLMrecon CA",
		NotBefore: time.Now().Add(-24 * time.Hour),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
	}, nil
}