// Package middleware provides middleware components for the Multi-Provider LLM Integration Framework.
package middleware

import (
	"sync"
	"time"
)

// UsageMetrics represents the usage metrics for a provider
type UsageMetrics struct {
	// Requests is the number of requests made
	Requests int64
	// Tokens is the number of tokens used
	Tokens int64
	// Errors is the number of errors encountered
	Errors int64
	// LastRequestTime is the time of the last request
	LastRequestTime time.Time
	// TotalRequestDuration is the total duration of all requests
	TotalRequestDuration time.Duration

// UsageTracker provides usage tracking functionality
type UsageTracker struct {
	// metrics is a map of model IDs to usage metrics
	metrics map[string]*UsageMetrics
	// mutex is a mutex for the metrics
	mutex sync.RWMutex
	// resetInterval is the interval at which to reset the metrics
	resetInterval time.Duration
	// lastResetTime is the time of the last reset
	lastResetTime time.Time

// NewUsageTracker creates a new usage tracker
func NewUsageTracker(resetInterval time.Duration) *UsageTracker {
	return &UsageTracker{
		metrics:       make(map[string]*UsageMetrics),
		mutex:         sync.RWMutex{},
		resetInterval: resetInterval,
		lastResetTime: time.Now(),
	}

// TrackRequest tracks a request
func (ut *UsageTracker) TrackRequest(modelID string, tokens int64, duration time.Duration, err error) {
	ut.mutex.Lock()
	defer ut.mutex.Unlock()

	// Check if we need to reset the metrics
	if ut.resetInterval > 0 && time.Since(ut.lastResetTime) > ut.resetInterval {
		ut.resetMetricsLocked()
	}

	// Get or create metrics for the model
	metrics, ok := ut.metrics[modelID]
	if !ok {
		metrics = &UsageMetrics{}
		ut.metrics[modelID] = metrics
	}

	// Update metrics
	metrics.Requests++
	metrics.Tokens += tokens
	metrics.LastRequestTime = time.Now()
	metrics.TotalRequestDuration += duration
	if err != nil {
		metrics.Errors++
	}

// GetMetrics returns the usage metrics for a model
func (ut *UsageTracker) GetMetrics(modelID string) *UsageMetrics {
	ut.mutex.RLock()
	defer ut.mutex.RUnlock()

	// Check if we need to reset the metrics
	if ut.resetInterval > 0 && time.Since(ut.lastResetTime) > ut.resetInterval {
		ut.mutex.RUnlock()
		ut.mutex.Lock()
		ut.resetMetricsLocked()
		ut.mutex.Unlock()
		ut.mutex.RLock()
	}

	// Get metrics for the model
	metrics, ok := ut.metrics[modelID]
	if !ok {
		return &UsageMetrics{}
	}
	return metrics

// GetAllMetrics returns the usage metrics for all models
func (ut *UsageTracker) GetAllMetrics() map[string]*UsageMetrics {
	ut.mutex.RLock()
	defer ut.mutex.RUnlock()

	// Check if we need to reset the metrics
	if ut.resetInterval > 0 && time.Since(ut.lastResetTime) > ut.resetInterval {
		ut.mutex.RUnlock()
		ut.mutex.Lock()
		ut.resetMetricsLocked()
		ut.mutex.Unlock()
		ut.mutex.RLock()
	}

	// Copy metrics
	metrics := make(map[string]*UsageMetrics, len(ut.metrics))
	for modelID, modelMetrics := range ut.metrics {
		metricsCopy := *modelMetrics
		metrics[modelID] = &metricsCopy
	}
	return metrics

// ResetMetrics resets the usage metrics
func (ut *UsageTracker) ResetMetrics() {
	ut.mutex.Lock()
	defer ut.mutex.Unlock()
	ut.resetMetricsLocked()

// resetMetricsLocked resets the usage metrics (must be called with the mutex locked)
func (ut *UsageTracker) resetMetricsLocked() {
	ut.metrics = make(map[string]*UsageMetrics)
	ut.lastResetTime = time.Now()

// SetResetInterval sets the reset interval
func (ut *UsageTracker) SetResetInterval(resetInterval time.Duration) {
	ut.mutex.Lock()
	defer ut.mutex.Unlock()
	ut.resetInterval = resetInterval
