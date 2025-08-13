// Package ratelimit provides rate limiting functionality for template execution
package ratelimit

import (
	"sync"
	"time"
)

// RateLimitEvent represents a single rate limiting event
type RateLimitEvent struct {
	// Type of event (acquire, reject, etc.)
	Type string
	
	// UserID associated with the event
	UserID string
	
	// Priority of the user at the time of the event
	Priority int
	
	// Timestamp when the event occurred
	Timestamp time.Time
	
	// Duration the request waited (for successful acquisitions)
	WaitDuration time.Duration
	
	// Error message (for rejections)
	ErrorMessage string
	
	// LoadFactor at the time of the event
	LoadFactor float64
}

// EventType constants for rate limit events
const (
	EventTypeAcquire           = "acquire"
	EventTypeReject            = "reject"
	EventTypeGlobalLimitExceed = "global_limit_exceed"
	EventTypeUserLimitExceed   = "user_limit_exceed"
	EventTypeTokensExceed      = "tokens_exceed"
	EventTypeQueueTimeout      = "queue_timeout"
)

// StatsCollector collects and provides statistics about rate limiting operations
type StatsCollector struct {
	// Total number of successful acquisitions
	TotalAcquisitions uint64
	
	// Total number of rejections
	TotalRejections uint64
	
	// Rejections by type
	RejectionsByType map[string]uint64
	
	// Rejections by user
	RejectionsByUser map[string]uint64
	
	// Acquisitions by user
	AcquisitionsByUser map[string]uint64
	
	// Recent events for detailed analysis
	RecentEvents []RateLimitEvent
	
	// Maximum number of recent events to keep
	maxRecentEvents int
	
	// Average wait time for successful acquisitions
	AverageWaitTime time.Duration
	
	// Total wait time for calculating average
	totalWaitTime time.Duration
	
	// Peak load factor observed
	PeakLoadFactor float64
	
	// Current load factor
	CurrentLoadFactor float64
	
	// Timestamp of last reset
	LastResetTime time.Time
	
	// Mutex for thread safety
	mu sync.RWMutex
}

// NewStatsCollector creates a new stats collector
func NewStatsCollector(maxRecentEvents int) *StatsCollector {
	return &StatsCollector{
		RejectionsByType:   make(map[string]uint64),
		RejectionsByUser:   make(map[string]uint64),
		AcquisitionsByUser: make(map[string]uint64),
		RecentEvents:       make([]RateLimitEvent, 0, maxRecentEvents),
		maxRecentEvents:    maxRecentEvents,
		LastResetTime:      time.Now(),
	}
}

// RecordEvent records a rate limiting event
func (s *StatsCollector) RecordEvent(event RateLimitEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Update load factor statistics
	if event.LoadFactor > s.PeakLoadFactor {
		s.PeakLoadFactor = event.LoadFactor
	}
	s.CurrentLoadFactor = event.LoadFactor
	
	// Add to recent events, maintaining max size
	s.RecentEvents = append(s.RecentEvents, event)
	if len(s.RecentEvents) > s.maxRecentEvents {
		s.RecentEvents = s.RecentEvents[1:]
	}
	
	// Update counters based on event type
	switch event.Type {
	case EventTypeAcquire:
		s.TotalAcquisitions++
		s.AcquisitionsByUser[event.UserID]++
		s.totalWaitTime += event.WaitDuration
		if s.TotalAcquisitions > 0 {
			s.AverageWaitTime = time.Duration(s.totalWaitTime.Nanoseconds() / int64(s.TotalAcquisitions))
		}
	case EventTypeReject, EventTypeGlobalLimitExceed, EventTypeUserLimitExceed, EventTypeTokensExceed, EventTypeQueueTimeout:
		s.TotalRejections++
		s.RejectionsByType[event.Type]++
		s.RejectionsByUser[event.UserID]++
	}
}

// GetStats returns a copy of the current statistics
func (s *StatsCollector) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Create a copy of the stats
	stats := map[string]interface{}{
		"total_acquisitions":   s.TotalAcquisitions,
		"total_rejections":     s.TotalRejections,
		"rejections_by_type":   copyMap(s.RejectionsByType),
		"rejections_by_user":   copyMap(s.RejectionsByUser),
		"acquisitions_by_user": copyMap(s.AcquisitionsByUser),
		"average_wait_time_ms": s.AverageWaitTime.Milliseconds(),
		"peak_load_factor":     s.PeakLoadFactor,
		"current_load_factor":  s.CurrentLoadFactor,
		"last_reset_time":      s.LastResetTime,
		"uptime_seconds":       time.Since(s.LastResetTime).Seconds(),
	}
	
	return stats
}

// GetRecentEvents returns a copy of recent events
func (s *StatsCollector) GetRecentEvents() []RateLimitEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Create a copy of recent events
	events := make([]RateLimitEvent, len(s.RecentEvents))
	copy(events, s.RecentEvents)
	
	return events
}

// Reset resets all statistics
func (s *StatsCollector) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.TotalAcquisitions = 0
	s.TotalRejections = 0
	s.RejectionsByType = make(map[string]uint64)
	s.RejectionsByUser = make(map[string]uint64)
	s.AcquisitionsByUser = make(map[string]uint64)
	s.RecentEvents = make([]RateLimitEvent, 0, s.maxRecentEvents)
	s.AverageWaitTime = 0
	s.totalWaitTime = 0
	s.PeakLoadFactor = s.CurrentLoadFactor
	s.LastResetTime = time.Now()
}

// Helper function to copy a map
func copyMap[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
