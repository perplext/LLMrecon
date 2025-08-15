package api

import (
	"fmt"
	"sync"
)

// ScanStore interface for scan storage
type ScanStore interface {
	Create(scan *Scan) error
	Get(id string) (*Scan, error)
	Update(scan *Scan) error
	Delete(id string) error
	List(filter ScanFilter) ([]Scan, error)
	CleanupOldScans(olderThan time.Duration) error

// InMemoryScanStore implements ScanStore using in-memory storage
type InMemoryScanStore struct {
	mu    sync.RWMutex
	scans map[string]*Scan

// NewInMemoryScanStore creates a new in-memory scan store
func NewInMemoryScanStore() *InMemoryScanStore {
	return &InMemoryScanStore{
		scans: make(map[string]*Scan),
	}

// Create creates a new scan
func (s *InMemoryScanStore) Create(scan *Scan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.scans[scan.ID]; exists {
		return fmt.Errorf("scan with ID %s already exists", scan.ID)
	}
	
	s.scans[scan.ID] = scan
	return nil

// Get retrieves a scan by ID
func (s *InMemoryScanStore) Get(id string) (*Scan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	scan, exists := s.scans[id]
	if !exists {
		return nil, fmt.Errorf("scan not found: %s", id)
	}
	
	// Return a copy to prevent external modifications
	scanCopy := *scan
	return &scanCopy, nil

// Update updates an existing scan
func (s *InMemoryScanStore) Update(scan *Scan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.scans[scan.ID]; !exists {
		return fmt.Errorf("scan not found: %s", scan.ID)
	}
	
	scan.UpdatedAt = time.Now()
	s.scans[scan.ID] = scan
	return nil

// Delete removes a scan
func (s *InMemoryScanStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.scans[id]; !exists {
		return fmt.Errorf("scan not found: %s", id)
	}
	
	delete(s.scans, id)
	return nil

// List returns scans matching the filter
func (s *InMemoryScanStore) List(filter ScanFilter) ([]Scan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var results []Scan
	
	for _, scan := range s.scans {
		// Apply filters
		if filter.Status != "" && scan.Status != filter.Status {
			continue
		}
		
		if filter.DateFrom != nil && scan.CreatedAt.Before(*filter.DateFrom) {
			continue
		}
		
		if filter.DateTo != nil && scan.CreatedAt.After(*filter.DateTo) {
			continue
		}
		
		// Add to results
		scanCopy := *scan
		results = append(results, scanCopy)
		
		// Check limit
		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}
	
	return results, nil

// CleanupOldScans removes scans older than the specified duration
func (s *InMemoryScanStore) CleanupOldScans(olderThan time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	toDelete := []string{}
	
	for id, scan := range s.scans {
		if scan.CreatedAt.Before(cutoff) {
			toDelete = append(toDelete, id)
		}
	}
	
	for _, id := range toDelete {
		delete(s.scans, id)
	}
	
	return nil

// MockScanService implements a mock scan service for testing
type MockScanService struct {
	store ScanStore
}

// NewMockScanService creates a new mock scan service
func NewMockScanService() *MockScanService {
	return &MockScanService{
		store: NewInMemoryScanStore(),
	}

// CreateScan creates a new scan
func (m *MockScanService) CreateScan(request CreateScanRequest) (*Scan, error) {
	scan := &Scan{
		ID:        generateMockToken(),
		Status:    ScanStatusPending,
		Target:    request.Target,
		Templates: request.Templates,
		Categories: request.Categories,
		Config:    request.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := m.store.Create(scan); err != nil {
		return nil, err
	}
	
	// Simulate scan execution
	go m.executeScan(scan.ID)
	
	return scan, nil

// GetScan retrieves a scan by ID
func (m *MockScanService) GetScan(id string) (*Scan, error) {
	return m.store.Get(id)

// ListScans lists all scans
func (m *MockScanService) ListScans(filter ScanFilter) ([]Scan, error) {
	return m.store.List(filter)

// CancelScan cancels a running scan
func (m *MockScanService) CancelScan(id string) error {
	scan, err := m.store.Get(id)
	if err != nil {
		return err
	}
	
	if scan.Status != ScanStatusRunning && scan.Status != ScanStatusPending {
		return fmt.Errorf("cannot cancel scan in status: %s", scan.Status)
	}
	
	scan.Status = ScanStatusCancelled
	return m.store.Update(scan)
// GetScanResults retrieves scan results
func (m *MockScanService) GetScanResults(id string) (*ScanResults, error) {
	scan, err := m.store.Get(id)
	if err != nil {
		return nil, err
	}
	
	if scan.Status != ScanStatusCompleted {
		return nil, fmt.Errorf("scan not completed")
	}
	
	return scan.Results, nil

// executeScan simulates scan execution
func (m *MockScanService) executeScan(id string) {
	// Update status to running
	scan, _ := m.store.Get(id)
	scan.Status = ScanStatusRunning
	now := time.Now()
	scan.StartedAt = &now
	m.store.Update(scan)
	
	// Simulate scan execution
	time.Sleep(5 * time.Second)
	
	// Check if cancelled
	scan, _ = m.store.Get(id)
	if scan.Status == ScanStatusCancelled {
		return
	}
	
	// Generate mock results
	results := &ScanResults{
		Summary: ResultSummary{
			TotalTests:      10,
			Passed:          7,
			Failed:          2,
			Errors:          1,
			Skipped:         0,
			SeverityCount:   map[string]int{"high": 1, "medium": 1, "low": 0},
			CategoryCount:   map[string]int{"prompt-injection": 1, "data-leakage": 1},
			ComplianceScore: 70.0,
		},
		Findings: []Finding{
			{
				ID:           generateMockToken(),
				TemplateID:   "prompt-injection-001",
				TemplateName: "Basic Prompt Injection",
				Category:     "Prompt Injection",
				Severity:     "High",
				Title:        "Prompt Injection Vulnerability Detected",
				Description:  "The model is vulnerable to direct prompt injection attacks",
				Remediation:  "Implement input validation and output filtering",
				Timestamp:    time.Now(),
			},
		},
		TemplateRuns: []TemplateExecution{
			{
				TemplateID: "prompt-injection-001",
				Status:     "completed",
				StartTime:  time.Now().Add(-4 * time.Second),
				EndTime:    time.Now().Add(-2 * time.Second),
				Duration:   "2s",
			},
		},
	}
	
	// Update scan with results
	scan.Status = ScanStatusCompleted
	scan.Results = results
	completedAt := time.Now()
	scan.CompletedAt = &completedAt
	scan.Duration = completedAt.Sub(*scan.StartedAt).String()
	m.store.Update(scan)
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
