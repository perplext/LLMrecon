// Package repository provides interfaces and implementations for accessing templates and modules
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
)

// OfflineBundleRepository implements the Repository interface for offline bundles
type OfflineBundleRepository struct {
	// basePath is the path to the offline bundle
	basePath string
	// validator is the offline bundle validator
	validator *bundle.OfflineBundleValidator
	// auditTrail is the audit trail manager
	auditTrail *trail.AuditTrailManager
	// validationLevel is the level of validation to perform
	validationLevel bundle.ValidationLevel
	// offlineBundle is the loaded offline bundle
	offlineBundle *bundle.OfflineBundle
	// isConnected indicates whether the repository is connected
	isConnected bool
}

// NewOfflineBundleRepository creates a new offline bundle repository
func NewOfflineBundleRepository(basePath string, auditTrail *trail.AuditTrailManager) *OfflineBundleRepository {
	return &OfflineBundleRepository{
		basePath:        basePath,
		validator:       bundle.NewOfflineBundleValidator(nil),
		auditTrail:      auditTrail,
		validationLevel: bundle.StandardValidation,
		isConnected:     false,
	}
}

// Connect connects to the repository
func (r *OfflineBundleRepository) Connect(ctx context.Context) error {
	// Check if the base path exists
	if _, err := os.Stat(r.basePath); os.IsNotExist(err) {
		return fmt.Errorf("offline bundle path does not exist: %s", r.basePath)
	}

	// Load the offline bundle
	offlineBundle, err := bundle.OpenOfflineBundle(r.basePath)
	if err != nil {
		return fmt.Errorf("failed to open offline bundle: %w", err)
	}

	// Validate the bundle
	result, err := r.validator.Validate(offlineBundle, r.validationLevel)
	if err != nil {
		return fmt.Errorf("failed to validate offline bundle: %w", err)
	}

	if !result.Valid {
		return fmt.Errorf("invalid offline bundle: %s", result.Message)
	}

	// Store the offline bundle
	r.offlineBundle = offlineBundle
	r.isConnected = true

	// Log audit event
	if r.auditTrail != nil {
		auditLog := &trail.AuditLog{
			ID:           uuid.New().String(),
			Operation:    "connect_offline_bundle_repository",
			ResourceType: "offline_bundle",
			ResourceID:   offlineBundle.EnhancedManifest.BundleID,
			Description:  fmt.Sprintf("Connected to offline bundle: %s", offlineBundle.EnhancedManifest.Name),
			Status:       "success",
			Timestamp:    time.Now(),
			IPAddress:    "",
			Details: map[string]interface{}{
				"bundle_name":      offlineBundle.EnhancedManifest.Name,
				"bundle_version":   offlineBundle.EnhancedManifest.Version,
				"validation_level": string(r.validationLevel),
				"base_path":        r.basePath,
			},
		}

		if err := r.auditTrail.LogOperation(context.Background(), auditLog); err != nil {
			// Log error but continue
			fmt.Printf("Warning: Failed to log audit event: %v\n", err)
		}
	}

	return nil
}

// Disconnect disconnects from the repository
func (r *OfflineBundleRepository) Disconnect(ctx context.Context) error {
	r.offlineBundle = nil
	r.isConnected = false
	return nil
}

// IsConnected checks if the repository is connected
func (r *OfflineBundleRepository) IsConnected() bool {
	return r.isConnected
}

// ListFiles lists files in the repository
func (r *OfflineBundleRepository) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	if !r.isConnected {
		return nil, fmt.Errorf("repository not connected")
	}

	var files []FileInfo
	for _, item := range r.offlineBundle.EnhancedManifest.Content {
		// Check if the item is in the requested path
		if path == "" || strings.HasPrefix(item.Path, path) {
			// Create file info
			fileInfo := FileInfo{
				Path:         item.Path,
				Name:         filepath.Base(item.Path),
				Size:         0, // Size is not available in the manifest
				IsDirectory:  false,
				LastModified: r.offlineBundle.EnhancedManifest.CreatedAt,
			}

			// Compliance mappings are not included in basic FileInfo structure

			files = append(files, fileInfo)
		}
	}

	return files, nil
}

// GetFile gets a file from the repository
func (r *OfflineBundleRepository) GetFile(ctx context.Context, path string) ([]byte, error) {
	if !r.isConnected {
		return nil, fmt.Errorf("repository not connected")
	}

	// Find the content item in the manifest
	var contentItem *bundle.ContentItem
	for _, item := range r.offlineBundle.EnhancedManifest.Content {
		if item.Path == path {
			contentItem = &item
			break
		}
	}

	if contentItem == nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Read the file from the offline bundle
	filePath := filepath.Join(r.basePath, path)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Log audit event
	if r.auditTrail != nil {
		auditLog := &trail.AuditLog{
			ID:           uuid.New().String(),
			Operation:    "get_file_from_offline_bundle",
			ResourceType: "file",
			ResourceID:   contentItem.ID,
			Description:  fmt.Sprintf("Retrieved file from offline bundle: %s", path),
			Status:       "success",
			Timestamp:    time.Now(),
			IPAddress:    "",
			Details: map[string]interface{}{
				"path":         path,
				"content_type": string(contentItem.Type),
				"bundle_id":    r.offlineBundle.EnhancedManifest.BundleID,
			},
		}

		if err := r.auditTrail.LogOperation(context.Background(), auditLog); err != nil {
			// Log error but continue
			fmt.Printf("Warning: Failed to log audit event: %v\n", err)
		}
	}

	return content, nil
}

// GetFileInfo gets information about a file in the repository
func (r *OfflineBundleRepository) GetFileInfo(ctx context.Context, path string) (FileInfo, error) {
	if !r.isConnected {
		return FileInfo{}, fmt.Errorf("repository not connected")
	}

	// Find the content item in the manifest
	var contentItem *bundle.ContentItem
	for _, item := range r.offlineBundle.EnhancedManifest.Content {
		if item.Path == path {
			contentItem = &item
			break
		}
	}

	if contentItem == nil {
		return FileInfo{}, fmt.Errorf("file not found: %s", path)
	}

	// Create file info
	fileInfo := FileInfo{
		Path:         contentItem.Path,
		Name:         filepath.Base(contentItem.Path),
		Size:         0, // Size is not available in the manifest
		IsDirectory:  false,
		LastModified: r.offlineBundle.EnhancedManifest.CreatedAt,
	}

	// Compliance mappings are not included in basic FileInfo structure

	return fileInfo, nil
}

// GetRepositoryInfo gets information about the repository
func (r *OfflineBundleRepository) GetRepositoryInfo(ctx context.Context) (RepositoryInfo, error) {
	if !r.isConnected {
		return RepositoryInfo{}, fmt.Errorf("repository not connected")
	}

	// Create repository info
	repoInfo := RepositoryInfo{
		Type:           RepositoryType("offline_bundle"),
		Name:           r.offlineBundle.EnhancedManifest.Name,
		URL:            r.basePath,
		LocalPath:      r.basePath,
		CurrentVersion: r.offlineBundle.EnhancedManifest.Version,
		LatestVersion:  r.offlineBundle.EnhancedManifest.Version,
		Description:    r.offlineBundle.EnhancedManifest.Description,
		LastSynced:     r.offlineBundle.EnhancedManifest.CreatedAt,
	}

	return repoInfo, nil
}

// SetValidationLevel sets the validation level for offline bundles
func (r *OfflineBundleRepository) SetValidationLevel(level bundle.ValidationLevel) {
	r.validationLevel = level
}

// GetValidationLevel gets the current validation level
func (r *OfflineBundleRepository) GetValidationLevel() bundle.ValidationLevel {
	return r.validationLevel
}

// GetOfflineBundle gets the loaded offline bundle
func (r *OfflineBundleRepository) GetOfflineBundle() *bundle.OfflineBundle {
	return r.offlineBundle
}
