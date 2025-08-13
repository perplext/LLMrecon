// Package bundle provides functionality for importing and exporting bundles
package bundle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/bundle/errors"
)

// ConflictType represents the type of conflict
type ConflictType string

const (
	// FileExistsConflict represents a conflict where a file already exists
	FileExistsConflict ConflictType = "file_exists"
	// ContentConflict represents a conflict where file content differs
	ContentConflict ConflictType = "content_conflict"
	// VersionConflict represents a conflict where versions differ
	VersionConflict ConflictType = "version_conflict"
	// DependencyConflict represents a conflict where dependencies differ
	DependencyConflict ConflictType = "dependency_conflict"
	// PermissionConflict represents a conflict where permissions differ
	PermissionConflict ConflictType = "permission_conflict"
)

// ConflictResolutionStrategy represents a strategy for resolving conflicts
type ConflictResolutionStrategy string

const (
	// SkipStrategy represents a strategy to skip conflicting items
	SkipStrategy ConflictResolutionStrategy = "skip"
	// OverwriteStrategy represents a strategy to overwrite conflicting items
	OverwriteStrategy ConflictResolutionStrategy = "overwrite"
	// MergeStrategy represents a strategy to merge conflicting items
	MergeStrategy ConflictResolutionStrategy = "merge"
	// RenameStrategy represents a strategy to rename conflicting items
	RenameStrategy ConflictResolutionStrategy = "rename"
	// KeepBothStrategy represents a strategy to keep both conflicting items
	KeepBothStrategy ConflictResolutionStrategy = "keep_both"
	// PromptStrategy represents a strategy to prompt the user for resolution
	PromptStrategy ConflictResolutionStrategy = "prompt"
)

// Conflict represents a conflict during bundle import
type Conflict struct {
	// Type is the type of conflict
	Type ConflictType
	// Path is the path of the conflicting item
	Path string
	// SourcePath is the path of the source item
	SourcePath string
	// TargetPath is the path of the target item
	TargetPath string
	// ContentItem is the content item from the bundle manifest
	ContentItem *ContentItem
	// Message is a human-readable message about the conflict
	Message string
	// ResolutionStrategy is the strategy for resolving the conflict
	ResolutionStrategy ConflictResolutionStrategy
	// ResolutionPath is the path where the item was resolved (if applicable)
	ResolutionPath string
	// Resolved indicates whether the conflict has been resolved
	Resolved bool
}

// ConflictResolution represents a resolution for a conflict
type ConflictResolution struct {
	// Conflict is the conflict being resolved
	Conflict *Conflict
	// Strategy is the strategy used to resolve the conflict
	Strategy ConflictResolutionStrategy
	// Path is the path where the item was resolved (if applicable)
	Path string
	// Success indicates whether the resolution was successful
	Success bool
	// Error is the error that occurred during resolution (if any)
	Error error
}

// ConflictDetector defines the interface for conflict detection
type ConflictDetector interface {
	// DetectConflicts detects conflicts between the bundle and the target directory
	DetectConflicts(ctx context.Context, bundle *Bundle, targetDir string) ([]*Conflict, error)
	// DetectContentConflicts detects content conflicts between the bundle and the target directory
	DetectContentConflicts(ctx context.Context, bundle *Bundle, targetDir string) ([]*Conflict, error)
	// DetectVersionConflicts detects version conflicts between the bundle and the target directory
	DetectVersionConflicts(ctx context.Context, bundle *Bundle, targetDir string) ([]*Conflict, error)
}

// ConflictResolver defines the interface for conflict resolution
type ConflictResolver interface {
	// ResolveConflict resolves a single conflict
	ResolveConflict(ctx context.Context, conflict *Conflict) (*ConflictResolution, error)
	// ResolveConflicts resolves multiple conflicts
	ResolveConflicts(ctx context.Context, conflicts []*Conflict) ([]*ConflictResolution, error)
	// GetDefaultStrategy gets the default resolution strategy for a conflict type
	GetDefaultStrategy(conflictType ConflictType) ConflictResolutionStrategy
	// SetDefaultStrategy sets the default resolution strategy for a conflict type
	SetDefaultStrategy(conflictType ConflictType, strategy ConflictResolutionStrategy)
}

// DefaultConflictDetector is the default implementation of ConflictDetector
type DefaultConflictDetector struct {
	// Logger is the logger for conflict detection operations
	Logger io.Writer
}

// NewConflictDetector creates a new conflict detector
func NewConflictDetector(logger io.Writer) ConflictDetector {
	if logger == nil {
		logger = os.Stdout
	}
	return &DefaultConflictDetector{
		Logger: logger,
	}
}

// DetectConflicts detects conflicts between the bundle and the target directory
func (d *DefaultConflictDetector) DetectConflicts(ctx context.Context, bundle *Bundle, targetDir string) ([]*Conflict, error) {
	var conflicts []*Conflict

	// Detect file existence conflicts
	for _, item := range bundle.Manifest.Content {
		targetPath := filepath.Join(targetDir, item.Path)
		
		// Check if the target path exists
		if _, err := os.Stat(targetPath); err == nil {
			// Target exists, create a conflict
			conflict := &Conflict{
				Type:               FileExistsConflict,
				Path:               item.Path,
				SourcePath:         filepath.Join(bundle.BundlePath, item.Path),
				TargetPath:         targetPath,
				ContentItem:        &item,
				Message:            fmt.Sprintf("File already exists: %s", item.Path),
				ResolutionStrategy: OverwriteStrategy, // Default strategy
				Resolved:           false,
			}
			conflicts = append(conflicts, conflict)
		}
	}

	// Detect content conflicts
	contentConflicts, err := d.DetectContentConflicts(ctx, bundle, targetDir)
	if err != nil {
		return nil, err
	}
	conflicts = append(conflicts, contentConflicts...)

	// Detect version conflicts
	versionConflicts, err := d.DetectVersionConflicts(ctx, bundle, targetDir)
	if err != nil {
		return nil, err
	}
	conflicts = append(conflicts, versionConflicts...)

	return conflicts, nil
}

// DetectContentConflicts detects content conflicts between the bundle and the target directory
func (d *DefaultConflictDetector) DetectContentConflicts(ctx context.Context, bundle *Bundle, targetDir string) ([]*Conflict, error) {
	var conflicts []*Conflict

	// Check each content item for content conflicts
	for _, item := range bundle.Manifest.Content {
		targetPath := filepath.Join(targetDir, item.Path)
		
		// Skip if the target doesn't exist
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			continue
		}

		// Check if it's a directory
		sourcePath := filepath.Join(bundle.BundlePath, item.Path)
		sourceInfo, err := os.Stat(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get source info for %s: %w", item.Path, err)
		}

		targetInfo, err := os.Stat(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get target info for %s: %w", item.Path, err)
		}

		// If one is a directory and the other is a file, it's a conflict
		if sourceInfo.IsDir() != targetInfo.IsDir() {
			conflict := &Conflict{
				Type:               ContentConflict,
				Path:               item.Path,
				SourcePath:         sourcePath,
				TargetPath:         targetPath,
				ContentItem:        &item,
				Message:            fmt.Sprintf("Type mismatch: %s (source: %v, target: %v)", item.Path, sourceInfo.IsDir(), targetInfo.IsDir()),
				ResolutionStrategy: OverwriteStrategy, // Default strategy
				Resolved:           false,
			}
			conflicts = append(conflicts, conflict)
			continue
		}

		// If both are directories, skip
		if sourceInfo.IsDir() {
			continue
		}

		// Both are files, compare content
		sourceHash, err := calculateFileHash(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate source hash for %s: %w", item.Path, err)
		}

		targetHash, err := calculateFileHash(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate target hash for %s: %w", item.Path, err)
		}

		// If the hashes are different, it's a conflict
		if sourceHash != targetHash {
			conflict := &Conflict{
				Type:               ContentConflict,
				Path:               item.Path,
				SourcePath:         sourcePath,
				TargetPath:         targetPath,
				ContentItem:        &item,
				Message:            fmt.Sprintf("Content differs: %s", item.Path),
				ResolutionStrategy: OverwriteStrategy, // Default strategy
				Resolved:           false,
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, nil
}

// DetectVersionConflicts detects version conflicts between the bundle and the target directory
func (d *DefaultConflictDetector) DetectVersionConflicts(ctx context.Context, bundle *Bundle, targetDir string) ([]*Conflict, error) {
	var conflicts []*Conflict

	// Check for version conflicts in templates and modules
	for _, item := range bundle.Manifest.Content {
		// Skip items without version information
		if item.Version == "" || (item.Type != "template" && item.Type != "module") {
			continue
		}

		targetPath := filepath.Join(targetDir, item.Path)
		
		// Skip if the target doesn't exist
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			continue
		}

		// Try to get version information from the target
		targetVersion, err := d.getItemVersion(targetPath, string(item.Type))
		if err != nil {
			// Skip if we can't get version information
			continue
		}

		// If the versions are different, it's a conflict
		if targetVersion != item.Version {
			conflict := &Conflict{
				Type:               VersionConflict,
				Path:               item.Path,
				SourcePath:         filepath.Join(bundle.BundlePath, item.Path),
				TargetPath:         targetPath,
				ContentItem:        &item,
				Message:            fmt.Sprintf("Version conflict: %s (source: %s, target: %s)", item.Path, item.Version, targetVersion),
				ResolutionStrategy: PromptStrategy, // Default strategy for version conflicts
				Resolved:           false,
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, nil
}

// getItemVersion gets the version of an item
func (d *DefaultConflictDetector) getItemVersion(path string, itemType string) (string, error) {
	// In a real implementation, this would parse the file to extract version information
	// For now, we'll just return a mock version
	return "1.0.0", nil
}

// DefaultConflictResolver is the default implementation of ConflictResolver
type DefaultConflictResolver struct {
	// Logger is the logger for conflict resolution operations
	Logger io.Writer
	// DefaultStrategies maps conflict types to default resolution strategies
	DefaultStrategies map[ConflictType]ConflictResolutionStrategy
	// PromptCallback is called when a conflict requires user input
	PromptCallback func(conflict *Conflict) (ConflictResolutionStrategy, error)
	// AuditLogger is used for logging audit events
	AuditLogger *errors.AuditLogger
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver(logger io.Writer, auditLogger *errors.AuditLogger, promptCallback func(conflict *Conflict) (ConflictResolutionStrategy, error)) ConflictResolver {
	if logger == nil {
		logger = os.Stdout
	}
	
	// Initialize default strategies
	defaultStrategies := make(map[ConflictType]ConflictResolutionStrategy)
	defaultStrategies[FileExistsConflict] = OverwriteStrategy
	defaultStrategies[ContentConflict] = OverwriteStrategy
	defaultStrategies[VersionConflict] = PromptStrategy
	defaultStrategies[DependencyConflict] = PromptStrategy
	defaultStrategies[PermissionConflict] = OverwriteStrategy
	
	// If no prompt callback is provided, use a default one that returns the default strategy
	if promptCallback == nil {
		promptCallback = func(conflict *Conflict) (ConflictResolutionStrategy, error) {
			return defaultStrategies[conflict.Type], nil
		}
	}
	
	return &DefaultConflictResolver{
		Logger:           logger,
		DefaultStrategies: defaultStrategies,
		PromptCallback:    promptCallback,
		AuditLogger:       auditLogger,
	}
}

// ResolveConflict resolves a single conflict
func (r *DefaultConflictResolver) ResolveConflict(ctx context.Context, conflict *Conflict) (*ConflictResolution, error) {
	// If the conflict is already resolved, return
	if conflict.Resolved {
		return &ConflictResolution{
			Conflict: conflict,
			Strategy: conflict.ResolutionStrategy,
			Path:     conflict.ResolutionPath,
			Success:  true,
		}, nil
	}

	// Get the resolution strategy
	strategy := conflict.ResolutionStrategy
	if strategy == "" {
		strategy = r.GetDefaultStrategy(conflict.Type)
	}

	// If the strategy is to prompt, call the prompt callback
	if strategy == PromptStrategy {
		var err error
		strategy, err = r.PromptCallback(conflict)
		if err != nil {
			return &ConflictResolution{
				Conflict: conflict,
				Strategy: PromptStrategy,
				Success:  false,
				Error:    err,
			}, err
		}
	}

	// Resolve the conflict based on the strategy
	resolution := &ConflictResolution{
		Conflict: conflict,
		Strategy: strategy,
		Success:  false,
	}

	// Create a bundle ID from the conflict path if not available
	bundleID := "unknown"
	if conflict.ContentItem != nil && conflict.ContentItem.BundleID != "" {
		bundleID = conflict.ContentItem.BundleID
	} else {
		// Use the path as a fallback
		bundleID = filepath.Base(conflict.Path)
	}

	// Log audit event for conflict resolution start
	if r.AuditLogger != nil {
		r.AuditLogger.LogEvent("conflict_resolution_started", "ConflictResolver", bundleID, map[string]interface{}{
			"conflict_type": string(conflict.Type),
			"conflict_path": conflict.Path,
			"strategy": string(strategy),
			"operation": "conflict_resolution",
		})
	}

	switch strategy {
	case SkipStrategy:
		// Skip the conflicting item
		fmt.Fprintf(r.Logger, "Skipping: %s\n", conflict.Path)
		resolution.Success = true
		conflict.Resolved = true
		conflict.ResolutionStrategy = SkipStrategy
		
		// Log audit event for skipped conflict
		if r.AuditLogger != nil {
			r.AuditLogger.LogConflict(bundleID, conflict, SkipStrategy)
		}

	case OverwriteStrategy:
		// Overwrite the target with the source
		fmt.Fprintf(r.Logger, "Overwriting: %s\n", conflict.Path)
		err := copyPath(conflict.SourcePath, conflict.TargetPath)
		if err != nil {
			resolution.Error = fmt.Errorf("failed to overwrite %s: %w", conflict.Path, err)
			return resolution, resolution.Error
		}
		resolution.Path = conflict.TargetPath
		resolution.Success = true
		conflict.Resolved = true
		conflict.ResolutionStrategy = OverwriteStrategy
		conflict.ResolutionPath = conflict.TargetPath
		
		// Log audit event for overwrite conflict resolution
		if r.AuditLogger != nil {
			r.AuditLogger.LogConflict(bundleID, conflict, OverwriteStrategy)
		}

	case MergeStrategy:
		// Merge the source and target
		fmt.Fprintf(r.Logger, "Merging: %s\n", conflict.Path)
		mergedPath, err := r.mergeFiles(conflict.SourcePath, conflict.TargetPath)
		if err != nil {
			resolution.Error = fmt.Errorf("failed to merge %s: %w", conflict.Path, err)
			return resolution, resolution.Error
		}
		// Copy the merged file to the target
		err = copyPath(mergedPath, conflict.TargetPath)
		if err != nil {
			resolution.Error = fmt.Errorf("failed to copy merged file %s: %w", conflict.Path, err)
			return resolution, resolution.Error
		}
		resolution.Path = conflict.TargetPath
		resolution.Success = true
		conflict.Resolved = true
		conflict.ResolutionStrategy = MergeStrategy
		conflict.ResolutionPath = conflict.TargetPath
		
		// Log audit event for merge conflict resolution
		if r.AuditLogger != nil {
			r.AuditLogger.LogConflict(bundleID, conflict, MergeStrategy)
		}

	case RenameStrategy:
		// Rename the source file before copying
		fmt.Fprintf(r.Logger, "Renaming: %s\n", conflict.Path)
		ext := filepath.Ext(conflict.TargetPath)
		base := strings.TrimSuffix(conflict.TargetPath, ext)
		timestamp := time.Now().Format("20060102-150405")
		newPath := fmt.Sprintf("%s.%s%s", base, timestamp, ext)
		
		// Copy the source to the new path
		err := copyPath(conflict.SourcePath, newPath)
		if err != nil {
			resolution.Error = fmt.Errorf("failed to copy %s to %s: %w", conflict.Path, newPath, err)
			return resolution, resolution.Error
		}
		resolution.Path = newPath
		resolution.Success = true
		conflict.Resolved = true
		conflict.ResolutionStrategy = RenameStrategy
		conflict.ResolutionPath = newPath
		
		// Log audit event for rename conflict resolution
		if r.AuditLogger != nil {
			r.AuditLogger.LogConflict(bundleID, conflict, RenameStrategy)
		}

	case KeepBothStrategy:
		// Keep both files
		fmt.Fprintf(r.Logger, "Keeping both: %s\n", conflict.Path)
		ext := filepath.Ext(conflict.TargetPath)
		base := strings.TrimSuffix(conflict.TargetPath, ext)
		newPath := fmt.Sprintf("%s.new%s", base, ext)
		
		// Copy the source to the new path
		err := copyPath(conflict.SourcePath, newPath)
		if err != nil {
			resolution.Error = fmt.Errorf("failed to copy %s to %s: %w", conflict.Path, newPath, err)
			return resolution, resolution.Error
		}
		resolution.Path = newPath
		resolution.Success = true
		conflict.Resolved = true
		conflict.ResolutionStrategy = KeepBothStrategy
		conflict.ResolutionPath = newPath
		
		// Log audit event for keep both conflict resolution
		if r.AuditLogger != nil {
			r.AuditLogger.LogConflict(bundleID, conflict, KeepBothStrategy)
		}

	default:
		resolution.Error = fmt.Errorf("unknown resolution strategy: %s", strategy)
		
		// Log audit event for unknown strategy
		if r.AuditLogger != nil {
			r.AuditLogger.LogConflict(bundleID, conflict, strategy)
		}
		
		return resolution, resolution.Error
	}

	return resolution, nil
}
// ResolveConflicts resolves multiple conflicts
func (r *DefaultConflictResolver) ResolveConflicts(ctx context.Context, conflicts []*Conflict) ([]*ConflictResolution, error) {
	var resolutions []*ConflictResolution
	var errors []error

	// Log audit event for batch conflict resolution start
	if r.AuditLogger != nil {
		details := map[string]interface{}{
			"operation": "batch_conflict_resolution",
			"count":     len(conflicts),
		}
		
		// Use the first conflict's bundleID if available
		bundleID := ""
		if len(conflicts) > 0 && conflicts[0].ContentItem != nil {
			bundleID = conflicts[0].ContentItem.BundleID
		}
		
		r.AuditLogger.LogEvent("batch_conflict_resolution_started", "ConflictResolver", bundleID, details)
	}

	// Resolve each conflict
	for _, conflict := range conflicts {
		resolution, err := r.ResolveConflict(ctx, conflict)
		resolutions = append(resolutions, resolution)
		if err != nil {
			errors = append(errors, err)
		}
	}

	// If there were any errors, return an error
	if len(errors) > 0 {
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		
		// Log audit event for batch conflict resolution completion with errors
		if r.AuditLogger != nil {
			details := map[string]interface{}{
				"operation":     "batch_conflict_resolution",
				"total":        len(conflicts),
				"successful":   len(conflicts) - len(errors),
				"failed":       len(errors),
				"error_count":  len(errors),
				"error_details": strings.Join(errorMessages, "; "),
			}
			
			// Use the first conflict's bundleID if available
			bundleID := ""
			if len(conflicts) > 0 && conflicts[0].ContentItem != nil {
				bundleID = conflicts[0].ContentItem.BundleID
			}
			
			r.AuditLogger.LogEvent("batch_conflict_resolution_completed", "ConflictResolver", bundleID, details)
		}
		
		return resolutions, fmt.Errorf("failed to resolve all conflicts: %s", strings.Join(errorMessages, "; "))
	}

	// Log audit event for successful batch conflict resolution completion
	if r.AuditLogger != nil {
		details := map[string]interface{}{
			"operation":   "batch_conflict_resolution",
			"total":      len(conflicts),
			"successful": len(conflicts),
			"failed":     0,
		}
		
		// Use the first conflict's bundleID if available
		bundleID := ""
		if len(conflicts) > 0 && conflicts[0].ContentItem != nil {
			bundleID = conflicts[0].ContentItem.BundleID
		}
		
		r.AuditLogger.LogEvent("batch_conflict_resolution_completed", "ConflictResolver", bundleID, details)
	}
	
	return resolutions, nil
}

// GetDefaultStrategy gets the default resolution strategy for a conflict type
func (r *DefaultConflictResolver) GetDefaultStrategy(conflictType ConflictType) ConflictResolutionStrategy {
	strategy, ok := r.DefaultStrategies[conflictType]
	if !ok {
		return PromptStrategy
	}
	return strategy
}

// SetDefaultStrategy sets the default resolution strategy for a conflict type
func (r *DefaultConflictResolver) SetDefaultStrategy(conflictType ConflictType, strategy ConflictResolutionStrategy) {
	r.DefaultStrategies[conflictType] = strategy
}

// mergeFiles merges two files and returns the path to the merged file
func (r *DefaultConflictResolver) mergeFiles(sourcePath, targetPath string) (string, error) {
	// In a real implementation, this would use a diff/merge algorithm
	// For now, we'll just create a simple merged file
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read target file: %w", err)
	}

	// Create a temporary file for the merged content
	tempFile, err := os.CreateTemp("", "merge-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tempFile.Close()

	// Write a simple merged file
	_, err = tempFile.WriteString("<<<<<<< SOURCE\n")
	if err != nil {
		return "", fmt.Errorf("failed to write to merged file: %w", err)
	}
	_, err = tempFile.Write(sourceData)
	if err != nil {
		return "", fmt.Errorf("failed to write to merged file: %w", err)
	}
	_, err = tempFile.WriteString("\n=======\n")
	if err != nil {
		return "", fmt.Errorf("failed to write to merged file: %w", err)
	}
	_, err = tempFile.Write(targetData)
	if err != nil {
		return "", fmt.Errorf("failed to write to merged file: %w", err)
	}
	_, err = tempFile.WriteString("\n>>>>>>> TARGET\n")
	if err != nil {
		return "", fmt.Errorf("failed to write to merged file: %w", err)
	}

	return tempFile.Name(), nil
}

// calculateFileHash calculates the SHA-256 hash of a file
func calculateFileHash(path string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new hash
	hash := sha256.New()

	// Copy the file to the hash
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	// Get the hash
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}
