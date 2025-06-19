package bundle

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// DeltaManifest represents the manifest for an incremental update bundle
type DeltaManifest struct {
	Version      string                `json:"version"`
	DeltaType    string                `json:"deltaType"`
	FromVersion  string                `json:"fromVersion"`
	ToVersion    string                `json:"toVersion"`
	Timestamp    time.Time             `json:"timestamp"`
	Size         DeltaSize             `json:"size"`
	Operations   DeltaOperations       `json:"operations"`
	Dependencies DeltaDependencies     `json:"dependencies"`
	Rollback     RollbackInfo          `json:"rollback"`
}

// DeltaSize contains size information for the delta bundle
type DeltaSize struct {
	Compressed   int64 `json:"compressed"`
	Uncompressed int64 `json:"uncompressed"`
}

// DeltaOperations contains lists of operations by type
type DeltaOperations struct {
	Add    []AddOperation    `json:"add"`
	Update []UpdateOperation `json:"update"`
	Delete []DeleteOperation `json:"delete"`
	Patch  []PatchOperation  `json:"patch"`
}

// AddOperation represents a file addition
type AddOperation struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Size int64  `json:"size"`
	Hash string `json:"hash"`
	Mode uint32 `json:"mode,omitempty"`
}

// UpdateOperation represents a file update
type UpdateOperation struct {
	Path           string `json:"path"`
	Type           string `json:"type"`
	OldHash        string `json:"oldHash"`
	NewHash        string `json:"newHash"`
	PatchAvailable bool   `json:"patchAvailable"`
	PatchSize      int64  `json:"patchSize,omitempty"`
}

// DeleteOperation represents a file deletion
type DeleteOperation struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// PatchOperation represents a patch to be applied
type PatchOperation struct {
	Path      string `json:"path"`
	Type      string `json:"type"`
	PatchFile string `json:"patchFile"`
	Algorithm string `json:"algorithm"`
}

// DeltaDependencies specifies version requirements
type DeltaDependencies struct {
	Required   []string `json:"required"`
	Compatible []string `json:"compatible"`
}

// RollbackInfo contains rollback configuration
type RollbackInfo struct {
	Supported        bool `json:"supported"`
	SnapshotRequired bool `json:"snapshotRequired"`
}

// UpdateContext holds the context for an update operation
type UpdateContext struct {
	CurrentVersion    string
	TargetVersion     string
	BundlePath        string
	BackupPath        string
	DeltaPath         string
	DryRun            bool
	Plan              *UpdatePlan
	AppliedOperations []Operation
	Progress          *UpdateProgress
}

// UpdatePlan represents a planned update
type UpdatePlan struct {
	Operations    []Operation `json:"operations"`
	SpaceRequired int64       `json:"spaceRequired"`
	BackupSize    int64       `json:"backupSize"`
	EstimatedTime int         `json:"estimatedTime"`
}

// Operation represents a single update operation
type Operation struct {
	Type      string      `json:"type"`
	Path      string      `json:"path"`
	Details   interface{} `json:"details"`
	Completed bool        `json:"completed"`
}

// UpdateProgress tracks update progress
type UpdateProgress struct {
	TotalOperations     int
	CompletedOperations int
	CurrentOperation    string
	BytesProcessed      int64
	TotalBytes          int64
}

// Backup represents a backup of files before update
type Backup struct {
	Version   string               `json:"version"`
	Timestamp time.Time            `json:"timestamp"`
	Files     map[string]FileBackup `json:"files"`
}

// FileBackup represents a backed up file
type FileBackup struct {
	Path         string `json:"path"`
	Hash         string `json:"hash"`
	Size         int64  `json:"size"`
	Mode         uint32 `json:"mode"`
	BackupPath   string `json:"backupPath"`
}

// DeltaBundle represents an incremental update bundle
type DeltaBundle struct {
	Path     string
	Manifest *DeltaManifest
}

// LoadDeltaBundle loads a delta bundle from disk
func LoadDeltaBundle(path string) (*DeltaBundle, error) {
	// Load manifest
	manifestPath := filepath.Join(path, "delta-manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read delta manifest: %w", err)
	}

	var manifest DeltaManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse delta manifest: %w", err)
	}

	return &DeltaBundle{
		Path:     path,
		Manifest: &manifest,
	}, nil
}

// GenerateDelta generates a delta bundle between two versions
func GenerateDelta(oldBundle, newBundle *Bundle) (*DeltaBundle, error) {
	delta := &DeltaBundle{
		Manifest: &DeltaManifest{
			Version:     "1.0",
			DeltaType:   "incremental",
			FromVersion: oldBundle.Manifest.Version,
			ToVersion:   newBundle.Manifest.Version,
			Timestamp:   time.Now().UTC(),
			Operations: DeltaOperations{
				Add:    []AddOperation{},
				Update: []UpdateOperation{},
				Delete: []DeleteOperation{},
				Patch:  []PatchOperation{},
			},
			Dependencies: DeltaDependencies{
				Required:   []string{oldBundle.Manifest.Version},
				Compatible: []string{oldBundle.Manifest.Version},
			},
			Rollback: RollbackInfo{
				Supported:        true,
				SnapshotRequired: true,
			},
		},
	}

	// Compare file lists
	oldFiles := make(map[string]*ContentItem)
	for i := range oldBundle.Manifest.Content {
		item := &oldBundle.Manifest.Content[i]
		oldFiles[item.Path] = item
	}

	newFiles := make(map[string]*ContentItem)
	for i := range newBundle.Manifest.Content {
		item := &newBundle.Manifest.Content[i]
		newFiles[item.Path] = item
	}

	// Find additions and updates
	for path, newItem := range newFiles {
		if oldItem, exists := oldFiles[path]; exists {
			// File exists in both - check if updated
			if oldItem.Checksum != newItem.Checksum {
				delta.Manifest.Operations.Update = append(delta.Manifest.Operations.Update, UpdateOperation{
					Path:    path,
					Type:    string(newItem.Type),
					OldHash: oldItem.Checksum,
					NewHash: newItem.Checksum,
				})
			}
		} else {
			// New file
			delta.Manifest.Operations.Add = append(delta.Manifest.Operations.Add, AddOperation{
				Path: path,
				Type: string(newItem.Type),
				Size: 0, // Size info not available in ContentItem
				Hash: newItem.Checksum,
			})
		}
	}

	// Find deletions
	for path, oldItem := range oldFiles {
		if _, exists := newFiles[path]; !exists {
			delta.Manifest.Operations.Delete = append(delta.Manifest.Operations.Delete, DeleteOperation{
				Path: path,
				Type: string(oldItem.Type),
			})
		}
	}

	return delta, nil
}

// PrepareUpdate prepares an update plan
func PrepareUpdate(ctx *UpdateContext) (*UpdatePlan, error) {
	// Load current bundle
	currentBundle, err := LoadBundle(ctx.BundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load current bundle: %w", err)
	}

	// Verify version
	if currentBundle.Manifest.Version != ctx.CurrentVersion {
		return nil, fmt.Errorf("version mismatch: expected %s, found %s", 
			ctx.CurrentVersion, currentBundle.Manifest.Version)
	}

	// Load delta bundle
	deltaBundle, err := LoadDeltaBundle(ctx.DeltaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load delta bundle: %w", err)
	}

	// Create update plan
	plan := &UpdatePlan{
		Operations:    []Operation{},
		SpaceRequired: 0,
		BackupSize:    0,
	}

	// Convert delta operations to plan operations
	for _, op := range deltaBundle.Manifest.Operations.Add {
		plan.Operations = append(plan.Operations, Operation{
			Type:    "add",
			Path:    op.Path,
			Details: op,
		})
		plan.SpaceRequired += op.Size
	}

	for _, op := range deltaBundle.Manifest.Operations.Update {
		plan.Operations = append(plan.Operations, Operation{
			Type:    "update",
			Path:    op.Path,
			Details: op,
		})
		// Add space for backup
		if fileInfo, err := os.Stat(filepath.Join(ctx.BundlePath, op.Path)); err == nil {
			plan.BackupSize += fileInfo.Size()
		}
	}

	for _, op := range deltaBundle.Manifest.Operations.Delete {
		plan.Operations = append(plan.Operations, Operation{
			Type:    "delete",
			Path:    op.Path,
			Details: op,
		})
	}

	// Estimate time (rough estimate: 1MB/second)
	plan.EstimatedTime = int((plan.SpaceRequired + plan.BackupSize) / (1024 * 1024))
	if plan.EstimatedTime < 1 {
		plan.EstimatedTime = 1
	}

	return plan, nil
}

// CreateBackup creates a backup of files that will be modified
func CreateBackup(ctx *UpdateContext) (*Backup, error) {
	backup := &Backup{
		Version:   ctx.CurrentVersion,
		Timestamp: time.Now().UTC(),
		Files:     make(map[string]FileBackup),
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(ctx.BackupPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup files that will be modified or deleted
	for _, op := range ctx.Plan.Operations {
		if op.Type == "update" || op.Type == "delete" {
			sourcePath := filepath.Join(ctx.BundlePath, op.Path)
			
			// Get file info
			info, err := os.Stat(sourcePath)
			if err != nil {
				continue // File might already be deleted
			}

			// Calculate hash
			hash, err := calculateDeltaFileHash(sourcePath)
			if err != nil {
				return nil, fmt.Errorf("failed to hash %s: %w", op.Path, err)
			}

			// Copy to backup
			backupFile := fmt.Sprintf("%s_%s", 
				filepath.Base(op.Path), 
				time.Now().Format("20060102_150405"))
			backupPath := filepath.Join(ctx.BackupPath, backupFile)
			
			if err := copyDeltaFile(sourcePath, backupPath); err != nil {
				return nil, fmt.Errorf("failed to backup %s: %w", op.Path, err)
			}

			backup.Files[op.Path] = FileBackup{
				Path:       op.Path,
				Hash:       hash,
				Size:       info.Size(),
				Mode:       uint32(info.Mode()),
				BackupPath: backupPath,
			}
		}
	}

	// Save backup manifest
	manifestPath := filepath.Join(ctx.BackupPath, "backup-manifest.json")
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to save backup manifest: %w", err)
	}

	return backup, nil
}

// ApplyOperations applies update operations
func ApplyOperations(ctx *UpdateContext, ops []Operation) error {
	for i, op := range ops {
		// Update progress
		ctx.Progress.CurrentOperation = fmt.Sprintf("%s: %s", op.Type, op.Path)
		ctx.Progress.CompletedOperations = i

		switch op.Type {
		case "add":
			if err := applyAddOperation(ctx, op); err != nil {
				return fmt.Errorf("failed to add %s: %w", op.Path, err)
			}
		case "update":
			if err := applyUpdateOperation(ctx, op); err != nil {
				return fmt.Errorf("failed to update %s: %w", op.Path, err)
			}
		case "delete":
			if err := applyDeleteOperation(ctx, op); err != nil {
				return fmt.Errorf("failed to delete %s: %w", op.Path, err)
			}
		case "patch":
			if err := applyPatchOperation(ctx, op); err != nil {
				return fmt.Errorf("failed to patch %s: %w", op.Path, err)
			}
		}

		op.Completed = true
		ctx.AppliedOperations = append(ctx.AppliedOperations, op)
	}

	return nil
}

// applyAddOperation adds a new file
func applyAddOperation(ctx *UpdateContext, op Operation) error {
	addOp := op.Details.(AddOperation)
	
	sourcePath := filepath.Join(ctx.DeltaPath, "operations", "add", addOp.Path)
	destPath := filepath.Join(ctx.BundlePath, addOp.Path)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Copy file
	return copyDeltaFile(sourcePath, destPath)
}

// applyUpdateOperation updates an existing file
func applyUpdateOperation(ctx *UpdateContext, op Operation) error {
	updateOp := op.Details.(UpdateOperation)
	
	sourcePath := filepath.Join(ctx.DeltaPath, "operations", "update", updateOp.Path)
	destPath := filepath.Join(ctx.BundlePath, updateOp.Path)

	// Copy new version
	return copyDeltaFile(sourcePath, destPath)
}

// applyDeleteOperation deletes a file
func applyDeleteOperation(ctx *UpdateContext, op Operation) error {
	deleteOp := op.Details.(DeleteOperation)
	
	targetPath := filepath.Join(ctx.BundlePath, deleteOp.Path)
	return os.Remove(targetPath)
}

// applyPatchOperation applies a patch to a file
func applyPatchOperation(ctx *UpdateContext, op Operation) error {
	// TODO: Implement patch application
	// This would use bsdiff or similar for binary patches
	// and unified diff for text patches
	return fmt.Errorf("patch operations not yet implemented")
}

// RollbackUpdate rolls back an update using a backup
func RollbackUpdate(ctx *UpdateContext, backup *Backup) error {
	// Restore backed up files
	for path, backupFile := range backup.Files {
		destPath := filepath.Join(ctx.BundlePath, path)
		
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}

		// Restore file
		if err := copyDeltaFile(backupFile.BackupPath, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", path, err)
		}

		// Restore permissions
		if err := os.Chmod(destPath, os.FileMode(backupFile.Mode)); err != nil {
			return fmt.Errorf("failed to restore permissions for %s: %w", path, err)
		}
	}

	// Remove added files
	for _, op := range ctx.AppliedOperations {
		if op.Type == "add" {
			targetPath := filepath.Join(ctx.BundlePath, op.Path)
			os.Remove(targetPath) // Ignore errors
		}
	}

	return nil
}

// calculateDeltaFileHash calculates SHA-256 hash of a file for delta operations
func calculateDeltaFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// copyDeltaFile copies a file from source to destination for delta operations
func copyDeltaFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// CompressDelta compresses a delta bundle
func CompressDelta(deltaPath string, outputPath string) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	gzWriter := gzip.NewWriter(output)
	defer gzWriter.Close()

	// TODO: Implement tar + gzip compression of delta directory
	return fmt.Errorf("delta compression not yet implemented")
}

// Update applies an update to the bundle
func (uc *UpdateContext) Update() error {
	// Progress tracking
	uc.Progress = &UpdateProgress{
		TotalOperations: len(uc.Plan.Operations),
	}

	// Step 1: Create backup
	backup, err := CreateBackup(uc)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Step 2: Apply operations
	if err := ApplyOperations(uc, uc.Plan.Operations); err != nil {
		// Rollback on failure
		if rollbackErr := RollbackUpdate(uc, backup); rollbackErr != nil {
			return fmt.Errorf("update failed: %v, rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("update failed (rolled back): %w", err)
	}

	// Step 3: Update version info
	if err := uc.updateVersionInfo(); err != nil {
		// Rollback on failure
		if rollbackErr := RollbackUpdate(uc, backup); rollbackErr != nil {
			return fmt.Errorf("version update failed: %v, rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("version update failed (rolled back): %w", err)
	}

	return nil
}

// updateVersionInfo updates the bundle version information
func (uc *UpdateContext) updateVersionInfo() error {
	// Update manifest version
	manifestPath := filepath.Join(uc.BundlePath, "manifest.json")
	
	// Load current manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest BundleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	// Update version
	manifest.Version = uc.TargetVersion
	manifest.CreatedAt = time.Now().UTC() // Using CreatedAt instead of UpdatedAt

	// Save updated manifest
	updatedData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, updatedData, 0644)
}