// Package scan provides API endpoints for managing red-team scans
package scan

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Service handles the business logic for scan operations
type Service struct {
	storage Storage
}

// NewService creates a new scan service
func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// CreateScanConfig creates a new scan configuration
func (s *Service) CreateScanConfig(ctx context.Context, req CreateScanConfigRequest, userID string) (*ScanConfig, error) {
	// Generate a unique ID
	id := uuid.New().String()

	config := &ScanConfig{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Target:      req.Target,
		TargetType:  req.TargetType,
		Templates:   req.Templates,
		Parameters:  req.Parameters,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   userID,
	}

	if err := s.storage.CreateScanConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to create scan configuration: %w", err)
	}

	return config, nil
}

// GetScanConfig retrieves a scan configuration by ID
func (s *Service) GetScanConfig(ctx context.Context, id string) (*ScanConfig, error) {
	config, err := s.storage.GetScanConfig(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get scan configuration: %w", err)
	}

	return config, nil
}

// UpdateScanConfig updates an existing scan configuration
func (s *Service) UpdateScanConfig(ctx context.Context, id string, req UpdateScanConfigRequest) (*ScanConfig, error) {
	// Get the existing config
	config, err := s.storage.GetScanConfig(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get scan configuration: %w", err)
	}

	// Update fields if provided
	if req.Name != "" {
		config.Name = req.Name
	}
	if req.Description != "" {
		config.Description = req.Description
	}
	if req.Target != "" {
		config.Target = req.Target
	}
	if req.TargetType != "" {
		config.TargetType = req.TargetType
	}
	if req.Templates != nil {
		config.Templates = req.Templates
	}
	if req.Parameters != nil {
		config.Parameters = req.Parameters
	}

	// Update timestamp
	config.UpdatedAt = time.Now()

	// Save the updated config
	if err := s.storage.UpdateScanConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update scan configuration: %w", err)
	}

	return config, nil
}

// DeleteScanConfig deletes a scan configuration by ID
func (s *Service) DeleteScanConfig(ctx context.Context, id string) error {
	if err := s.storage.DeleteScanConfig(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to delete scan configuration: %w", err)
	}

	return nil
}

// ListScanConfigs lists scan configurations with pagination and filtering
func (s *Service) ListScanConfigs(ctx context.Context, page, pageSize int, filters FilterParams) ([]*ScanConfig, *PaginationParams, error) {
	configs, total, err := s.storage.ListScanConfigs(ctx, page, pageSize, filters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list scan configurations: %w", err)
	}

	// Calculate pagination parameters
	pagination := calculatePagination(page, pageSize, total)

	return configs, pagination, nil
}

// CreateScan creates a new scan execution
func (s *Service) CreateScan(ctx context.Context, req CreateScanRequest) (*Scan, error) {
	// Check if the config exists
	config, err := s.storage.GetScanConfig(ctx, req.ConfigID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("scan configuration not found: %s", req.ConfigID)
		}
		return nil, fmt.Errorf("failed to get scan configuration: %w", err)
	}

	// Generate a unique ID
	id := uuid.New().String()

	// Create the scan
	scan := &Scan{
		ID:        id,
		ConfigID:  config.ID,
		Status:    ScanStatusPending,
		StartTime: time.Now(),
		Progress:  0,
	}

	if err := s.storage.CreateScan(ctx, scan); err != nil {
		return nil, fmt.Errorf("failed to create scan: %w", err)
	}

	// Start the scan asynchronously
	go s.runScan(context.Background(), scan)

	return scan, nil
}

// GetScan retrieves a scan by ID
func (s *Service) GetScan(ctx context.Context, id string) (*Scan, error) {
	scan, err := s.storage.GetScan(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get scan: %w", err)
	}

	return scan, nil
}

// CancelScan cancels a running scan
func (s *Service) CancelScan(ctx context.Context, id string) (*Scan, error) {
	// Get the scan
	scan, err := s.storage.GetScan(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get scan: %w", err)
	}

	// Check if the scan can be cancelled
	if scan.Status != ScanStatusPending && scan.Status != ScanStatusRunning {
		return nil, fmt.Errorf("cannot cancel scan with status %s", scan.Status)
	}

	// Update the scan status
	scan.Status = ScanStatusCancelled
	scan.EndTime = time.Now()

	// Save the updated scan
	if err := s.storage.UpdateScan(ctx, scan); err != nil {
		return nil, fmt.Errorf("failed to update scan: %w", err)
	}

	return scan, nil
}

// ListScans lists scans with pagination and filtering
func (s *Service) ListScans(ctx context.Context, page, pageSize int, filters FilterParams) ([]*Scan, *PaginationParams, error) {
	scans, total, err := s.storage.ListScans(ctx, page, pageSize, filters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list scans: %w", err)
	}

	// Calculate pagination parameters
	pagination := calculatePagination(page, pageSize, total)

	return scans, pagination, nil
}

// GetScanResults retrieves the results for a scan with pagination and filtering
func (s *Service) GetScanResults(ctx context.Context, scanID string, page, pageSize int, filters FilterParams) ([]*ScanResult, *PaginationParams, error) {
	// Check if the scan exists
	if _, err := s.storage.GetScan(ctx, scanID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("failed to get scan: %w", err)
	}

	// Get the results
	results, total, err := s.storage.ListScanResults(ctx, scanID, page, pageSize, filters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list scan results: %w", err)
	}

	// Calculate pagination parameters
	pagination := calculatePagination(page, pageSize, total)

	return results, pagination, nil
}

// Helper function to calculate pagination parameters
func calculatePagination(page, pageSize, total int) *PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationParams{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

// runScan executes a scan asynchronously
func (s *Service) runScan(ctx context.Context, scan *Scan) {
	// Update the scan status
	scan.Status = ScanStatusRunning
	if err := s.storage.UpdateScan(ctx, scan); err != nil {
		// Log the error but continue
		fmt.Printf("Failed to update scan status: %v\n", err)
	}

	// Get the scan configuration
	config, err := s.storage.GetScanConfig(ctx, scan.ConfigID)
	if err != nil {
		scan.Status = ScanStatusFailed
		scan.Error = fmt.Sprintf("Failed to get scan configuration: %v", err)
		scan.EndTime = time.Now()
		if err := s.storage.UpdateScan(ctx, scan); err != nil {
			fmt.Printf("Failed to update scan status: %v\n", err)
		}
		return
	}

	// In a real implementation, we would execute the scan using the templates
	// and parameters from the configuration. For now, we'll simulate progress
	// and generate some sample results.

	// Simulate scan execution
	// In a real implementation, this would be replaced with actual scan execution
	// using the templates and parameters from the configuration
	simulateScanExecution(ctx, s, scan, config)
}

// simulateScanExecution simulates the execution of a scan
// This is a placeholder for the actual scan execution logic
func simulateScanExecution(ctx context.Context, s *Service, scan *Scan, config *ScanConfig) {
	// Simulate progress updates
	for i := 0; i <= 100; i += 10 {
		// Check if the scan has been cancelled
		updatedScan, err := s.storage.GetScan(ctx, scan.ID)
		if err != nil {
			fmt.Printf("Failed to get scan: %v\n", err)
			return
		}

		if updatedScan.Status == ScanStatusCancelled {
			fmt.Printf("Scan %s has been cancelled\n", scan.ID)
			return
		}

		// Update progress
		scan.Progress = i
		if err := s.storage.UpdateScan(ctx, scan); err != nil {
			fmt.Printf("Failed to update scan progress: %v\n", err)
		}

		// Simulate some work
		time.Sleep(100 * time.Millisecond)

		// Generate a sample result every 20%
		if i > 0 && i%20 == 0 {
			result := &ScanResult{
				ID:          uuid.New().String(),
				ScanID:      scan.ID,
				TemplateID:  config.Templates[0],
				Severity:    getSeverityForProgress(i),
				Title:       fmt.Sprintf("Sample finding at %d%% progress", i),
				Description: fmt.Sprintf("This is a sample finding generated at %d%% progress", i),
				Details: map[string]interface{}{
					"progress": i,
					"sample":   true,
				},
				Timestamp: time.Now(),
			}

			if err := s.storage.CreateScanResult(ctx, result); err != nil {
				fmt.Printf("Failed to create scan result: %v\n", err)
			}
		}
	}

	// Complete the scan
	scan.Status = ScanStatusCompleted
	scan.Progress = 100
	scan.EndTime = time.Now()
	if err := s.storage.UpdateScan(ctx, scan); err != nil {
		fmt.Printf("Failed to update scan status: %v\n", err)
	}
}

// getSeverityForProgress returns a severity level based on the progress
// This is just for demonstration purposes
func getSeverityForProgress(progress int) ScanSeverity {
	switch {
	case progress <= 20:
		return ScanSeverityLow
	case progress <= 40:
		return ScanSeverityMedium
	case progress <= 60:
		return ScanSeverityHigh
	default:
		return ScanSeverityCritical
	}
}
