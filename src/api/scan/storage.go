// Package scan provides API endpoints for managing red-team scans
package scan

import (
	"context"
	"errors"
	"sync"
)

// Common errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidID         = errors.New("invalid ID")
	ErrInvalidParameters = errors.New("invalid parameters")
	ErrAlreadyExists     = errors.New("resource already exists")
)

// Storage defines the interface for storing scan data
type Storage interface {
	// ScanConfig methods
	CreateScanConfig(ctx context.Context, config *ScanConfig) error
	GetScanConfig(ctx context.Context, id string) (*ScanConfig, error)
	UpdateScanConfig(ctx context.Context, config *ScanConfig) error
	DeleteScanConfig(ctx context.Context, id string) error
	ListScanConfigs(ctx context.Context, page, pageSize int, filters FilterParams) ([]*ScanConfig, int, error)

	// Scan methods
	CreateScan(ctx context.Context, scan *Scan) error
	GetScan(ctx context.Context, id string) (*Scan, error)
	UpdateScan(ctx context.Context, scan *Scan) error
	DeleteScan(ctx context.Context, id string) error
	ListScans(ctx context.Context, page, pageSize int, filters FilterParams) ([]*Scan, int, error)

	// ScanResult methods
	CreateScanResult(ctx context.Context, result *ScanResult) error
	GetScanResult(ctx context.Context, id string) (*ScanResult, error)
	UpdateScanResult(ctx context.Context, result *ScanResult) error
	DeleteScanResult(ctx context.Context, id string) error
	ListScanResults(ctx context.Context, scanID string, page, pageSize int, filters FilterParams) ([]*ScanResult, int, error)
}

// MemoryStorage implements the Storage interface using in-memory storage
// This is primarily for testing and development purposes
type MemoryStorage struct {
	configs      map[string]*ScanConfig
	scans        map[string]*Scan
	results      map[string]*ScanResult
	resultsByScan map[string][]*ScanResult
	mu           sync.RWMutex
}

// NewMemoryStorage creates a new MemoryStorage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		configs:      make(map[string]*ScanConfig),
		scans:        make(map[string]*Scan),
		results:      make(map[string]*ScanResult),
		resultsByScan: make(map[string][]*ScanResult),
	}
}

// CreateScanConfig creates a new scan configuration
func (s *MemoryStorage) CreateScanConfig(ctx context.Context, config *ScanConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config.ID == "" {
		return ErrInvalidID
	}

	if _, exists := s.configs[config.ID]; exists {
		return ErrAlreadyExists
	}

	// Set timestamps
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	// Store a copy to prevent external modification
	configCopy := *config
	s.configs[config.ID] = &configCopy

	return nil
}

// GetScanConfig retrieves a scan configuration by ID
func (s *MemoryStorage) GetScanConfig(ctx context.Context, id string) (*ScanConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to prevent external modification
	configCopy := *config
	return &configCopy, nil
}

// UpdateScanConfig updates an existing scan configuration
func (s *MemoryStorage) UpdateScanConfig(ctx context.Context, config *ScanConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config.ID == "" {
		return ErrInvalidID
	}

	if _, exists := s.configs[config.ID]; !exists {
		return ErrNotFound
	}

	// Update timestamp
	config.UpdatedAt = time.Now()

	// Store a copy to prevent external modification
	configCopy := *config
	s.configs[config.ID] = &configCopy

	return nil
}

// DeleteScanConfig deletes a scan configuration by ID
func (s *MemoryStorage) DeleteScanConfig(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[id]; !exists {
		return ErrNotFound
	}

	delete(s.configs, id)
	return nil
}

// ListScanConfigs lists scan configurations with pagination and filtering
func (s *MemoryStorage) ListScanConfigs(ctx context.Context, page, pageSize int, filters FilterParams) ([]*ScanConfig, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply filters
	var filtered []*ScanConfig
	for _, config := range s.configs {
		if matchesConfigFilters(config, filters) {
			configCopy := *config
			filtered = append(filtered, &configCopy)
		}
	}

	// Sort by creation time (newest first)
	// In a real implementation, we would use a more efficient sorting method
	// and possibly allow sorting by different fields
	sortConfigsByCreationTime(filtered)

	// Apply pagination
	total := len(filtered)
	start, end := calculatePaginationBounds(page, pageSize, total)
	if start >= total {
		return []*ScanConfig{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// CreateScan creates a new scan
func (s *MemoryStorage) CreateScan(ctx context.Context, scan *Scan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if scan.ID == "" {
		return ErrInvalidID
	}

	if _, exists := s.scans[scan.ID]; exists {
		return ErrAlreadyExists
	}

	// Verify that the referenced config exists
	if _, exists := s.configs[scan.ConfigID]; !exists {
		return ErrNotFound
	}

	// Store a copy to prevent external modification
	scanCopy := *scan
	scanCopy.Results = nil // Results are stored separately
	s.scans[scan.ID] = &scanCopy

	return nil
}

// GetScan retrieves a scan by ID
func (s *MemoryStorage) GetScan(ctx context.Context, id string) (*Scan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scan, exists := s.scans[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Create a copy with results
	scanCopy := *scan
	results := s.resultsByScan[id]
	if results != nil {
		scanCopy.Results = make([]ScanResult, len(results))
		for i, result := range results {
			scanCopy.Results[i] = *result
		}
	}

	return &scanCopy, nil
}

// UpdateScan updates an existing scan
func (s *MemoryStorage) UpdateScan(ctx context.Context, scan *Scan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if scan.ID == "" {
		return ErrInvalidID
	}

	if _, exists := s.scans[scan.ID]; !exists {
		return ErrNotFound
	}

	// Store a copy to prevent external modification
	scanCopy := *scan
	scanCopy.Results = nil // Results are stored separately
	s.scans[scan.ID] = &scanCopy

	return nil
}

// DeleteScan deletes a scan by ID
func (s *MemoryStorage) DeleteScan(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.scans[id]; !exists {
		return ErrNotFound
	}

	delete(s.scans, id)
	delete(s.resultsByScan, id) // Also delete associated results

	// Delete individual results
	for resultID, result := range s.results {
		if result.ScanID == id {
			delete(s.results, resultID)
		}
	}

	return nil
}

// ListScans lists scans with pagination and filtering
func (s *MemoryStorage) ListScans(ctx context.Context, page, pageSize int, filters FilterParams) ([]*Scan, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply filters
	var filtered []*Scan
	for _, scan := range s.scans {
		if matchesScanFilters(scan, filters) {
			scanCopy := *scan
			filtered = append(filtered, &scanCopy)
		}
	}

	// Sort by start time (newest first)
	sortScansByStartTime(filtered)

	// Apply pagination
	total := len(filtered)
	start, end := calculatePaginationBounds(page, pageSize, total)
	if start >= total {
		return []*Scan{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// CreateScanResult creates a new scan result
func (s *MemoryStorage) CreateScanResult(ctx context.Context, result *ScanResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if result.ID == "" {
		return ErrInvalidID
	}

	if _, exists := s.results[result.ID]; exists {
		return ErrAlreadyExists
	}

	// Verify that the referenced scan exists
	if _, exists := s.scans[result.ScanID]; !exists {
		return ErrNotFound
	}

	// Store a copy to prevent external modification
	resultCopy := *result
	s.results[result.ID] = &resultCopy

	// Add to results by scan index
	s.resultsByScan[result.ScanID] = append(s.resultsByScan[result.ScanID], &resultCopy)

	return nil
}

// GetScanResult retrieves a scan result by ID
func (s *MemoryStorage) GetScanResult(ctx context.Context, id string) (*ScanResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exists := s.results[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to prevent external modification
	resultCopy := *result
	return &resultCopy, nil
}

// UpdateScanResult updates an existing scan result
func (s *MemoryStorage) UpdateScanResult(ctx context.Context, result *ScanResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if result.ID == "" {
		return ErrInvalidID
	}

	existing, exists := s.results[result.ID]
	if !exists {
		return ErrNotFound
	}

	// Store a copy to prevent external modification
	resultCopy := *result
	s.results[result.ID] = &resultCopy

	// Update in the results by scan index
	scanID := existing.ScanID
	for i, r := range s.resultsByScan[scanID] {
		if r.ID == result.ID {
			s.resultsByScan[scanID][i] = &resultCopy
			break
		}
	}

	return nil
}

// DeleteScanResult deletes a scan result by ID
func (s *MemoryStorage) DeleteScanResult(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, exists := s.results[id]
	if !exists {
		return ErrNotFound
	}

	// Remove from results by scan index
	scanID := result.ScanID
	for i, r := range s.resultsByScan[scanID] {
		if r.ID == id {
			s.resultsByScan[scanID] = append(s.resultsByScan[scanID][:i], s.resultsByScan[scanID][i+1:]...)
			break
		}
	}

	delete(s.results, id)
	return nil
}

// ListScanResults lists scan results for a specific scan with pagination and filtering
func (s *MemoryStorage) ListScanResults(ctx context.Context, scanID string, page, pageSize int, filters FilterParams) ([]*ScanResult, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if scan exists
	if _, exists := s.scans[scanID]; !exists {
		return nil, 0, ErrNotFound
	}

	// Get results for this scan
	results := s.resultsByScan[scanID]
	if results == nil {
		return []*ScanResult{}, 0, nil
	}

	// Apply filters
	var filtered []*ScanResult
	for _, result := range results {
		if matchesResultFilters(result, filters) {
			resultCopy := *result
			filtered = append(filtered, &resultCopy)
		}
	}

	// Sort by timestamp (newest first)
	sortResultsByTimestamp(filtered)

	// Apply pagination
	total := len(filtered)
	start, end := calculatePaginationBounds(page, pageSize, total)
	if start >= total {
		return []*ScanResult{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// Helper functions for filtering and sorting

func matchesConfigFilters(config *ScanConfig, filters FilterParams) bool {
	// Apply search filter if provided
	if filters.Search != "" {
		// Simple case-insensitive substring search
		// In a real implementation, we would use a more sophisticated search mechanism
		searchLower := filters.Search
		if !(containsIgnoreCase(config.Name, searchLower) ||
			containsIgnoreCase(config.Description, searchLower) ||
			containsIgnoreCase(config.Target, searchLower) ||
			containsIgnoreCase(config.TargetType, searchLower)) {
			return false
		}
	}

	// Add more filters as needed
	return true
}

func matchesScanFilters(scan *Scan, filters FilterParams) bool {
	// Apply status filter if provided
	if filters.Status != "" && string(scan.Status) != filters.Status {
		return false
	}

	// Apply date filters if provided
	if filters.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, filters.StartDate)
		if err == nil && scan.StartTime.Before(startDate) {
			return false
		}
	}

	if filters.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, filters.EndDate)
		if err == nil && !scan.EndTime.IsZero() && scan.EndTime.After(endDate) {
			return false
		}
	}

	// Add more filters as needed
	return true
}

func matchesResultFilters(result *ScanResult, filters FilterParams) bool {
	// Apply severity filter if provided
	if filters.Severity != "" && string(result.Severity) != filters.Severity {
		return false
	}

	// Apply search filter if provided
	if filters.Search != "" {
		// Simple case-insensitive substring search
		searchLower := filters.Search
		if !(containsIgnoreCase(result.Title, searchLower) ||
			containsIgnoreCase(result.Description, searchLower)) {
			return false
		}
	}

	// Add more filters as needed
	return true
}

func sortConfigsByCreationTime(configs []*ScanConfig) {
	// Sort by creation time (newest first)
	// In a real implementation, we would use a more efficient sorting method
	for i := 0; i < len(configs)-1; i++ {
		for j := i + 1; j < len(configs); j++ {
			if configs[i].CreatedAt.Before(configs[j].CreatedAt) {
				configs[i], configs[j] = configs[j], configs[i]
			}
		}
	}
}

func sortScansByStartTime(scans []*Scan) {
	// Sort by start time (newest first)
	for i := 0; i < len(scans)-1; i++ {
		for j := i + 1; j < len(scans); j++ {
			if scans[i].StartTime.Before(scans[j].StartTime) {
				scans[i], scans[j] = scans[j], scans[i]
			}
		}
	}
}

func sortResultsByTimestamp(results []*ScanResult) {
	// Sort by timestamp (newest first)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Timestamp.Before(results[j].Timestamp) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func calculatePaginationBounds(page, pageSize, total int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	return start, end
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive substring check
	// In a real implementation, we would use a more sophisticated search mechanism
	return true // Placeholder implementation
}
