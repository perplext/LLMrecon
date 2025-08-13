package main

import (
	"fmt"
	"sync"
)

// MockMonitoringService is a simplified version of the monitoring service for the example
type MockMonitoringService struct {
	monitors map[string]interface{}
	mutex    sync.RWMutex
}

// NewMockMonitoringService creates a new mock monitoring service
func NewMockMonitoringService() *MockMonitoringService {
	return &MockMonitoringService{
		monitors: make(map[string]interface{}),
	}
}

// AddStaticFileMonitor adds a static file monitor to the service
func (s *MockMonitoringService) AddStaticFileMonitor(fileHandler interface{}) *MockStaticFileMonitor {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	monitor := NewMockStaticFileMonitor(fileHandler)
	s.monitors[monitor.ID] = monitor
	return monitor
}

// Start starts the monitoring service
func (s *MockMonitoringService) Start() {
	fmt.Println("Mock monitoring service started")
}

// Stop stops the monitoring service
func (s *MockMonitoringService) Stop() {
	fmt.Println("Mock monitoring service stopped")
}

// MockStaticFileMonitor is a simplified version of the static file monitor for the example
type MockStaticFileMonitor struct {
	ID          string
	FileHandler interface{}
	metrics     MockStaticFileMetrics
	mutex       sync.RWMutex
}

// NewMockStaticFileMonitor creates a new mock static file monitor
func NewMockStaticFileMonitor(fileHandler interface{}) *MockStaticFileMonitor {
	return &MockStaticFileMonitor{
		ID:          fmt.Sprintf("static-file-monitor-%d", time.Now().UnixNano()),
		FileHandler: fileHandler,
		metrics: MockStaticFileMetrics{
			FilesServed:      0,
			CacheHits:        0,
			CacheMisses:      0,
			CacheHitRatio:    0.0,
			CompressedFiles:  0,
			CompressionRatio: 0.0,
			AverageServeTime: 5 * time.Millisecond,
			CacheSize:        0,
			CacheItemCount:   0,
		},
	}
}

// GetMetrics returns the current metrics
func (m *MockStaticFileMonitor) GetMetrics() MockStaticFileMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Simulate some metrics for the example
	m.metrics.FilesServed += 1
	m.metrics.CacheHits += 1
	m.metrics.CacheHitRatio = float64(m.metrics.CacheHits) / float64(m.metrics.FilesServed)
	m.metrics.CacheSize += 1024
	m.metrics.CacheItemCount += 1
	
	return m.metrics
}

// CheckAlerts checks for alerts based on the current metrics
func (m *MockStaticFileMonitor) CheckAlerts() []MockAlert {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Simulate some alerts for the example
	var alerts []MockAlert
	
	if m.metrics.CacheHitRatio < 0.5 {
		alerts = append(alerts, MockAlert{
			ID:       "low-cache-hit-ratio",
			Message:  "Cache hit ratio is below 50%",
			Severity: "info",
		})
	}
	
	if m.metrics.AverageServeTime > 50*time.Millisecond {
		alerts = append(alerts, MockAlert{
			ID:       "slow-serve-time",
			Message:  "Average serve time is above 50ms",
			Severity: "warning",
		})
	}
	
	return alerts
}

// MockStaticFileMetrics represents metrics for the static file handler
type MockStaticFileMetrics struct {
	FilesServed      int64
	CacheHits        int64
	CacheMisses      int64
	CacheHitRatio    float64
	CompressedFiles  int64
	CompressionRatio float64
	AverageServeTime time.Duration
	CacheSize        int64
	CacheItemCount   int64
}

// MockAlert represents an alert from the monitoring system
type MockAlert struct {
	ID       string
	Message  string
	Severity string
}
