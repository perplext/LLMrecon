// Package core provides the core interfaces and types for the Multi-Provider LLM Integration Framework.
package core

import "time"

import (
)

// UsageMetrics represents the usage metrics for a provider
type UsageMetrics struct {
	// Requests is the number of requests made
	Requests int64 `json:"requests"`
	// Tokens is the number of tokens used
	Tokens int64 `json:"tokens"`
	// Errors is the number of errors encountered
	Errors int64 `json:"errors"`
	// LastRequestTime is the time of the last request
	LastRequestTime time.Time `json:"last_request_time"`
	// TotalRequestDuration is the total duration of all requests
	TotalRequestDuration time.Duration `json:"total_request_duration"`
	// AverageResponseTime is the average response time
	AverageResponseTime time.Duration `json:"average_response_time"`
	// TokensPerMinute is the average tokens per minute
	TokensPerMinute float64 `json:"tokens_per_minute"`
	// RequestsPerMinute is the average requests per minute
	RequestsPerMinute float64 `json:"requests_per_minute"`
	// ModelID is the ID of the model
	ModelID string `json:"model_id"`
}

// NewUsageMetrics creates a new usage metrics instance
func NewUsageMetrics(modelID string) *UsageMetrics {
	return &UsageMetrics{
		ModelID: modelID,
	}
}

// AddRequest adds a request to the usage metrics
func (m *UsageMetrics) AddRequest(tokens int64, duration time.Duration, err error) {
	m.Requests++
	m.Tokens += tokens
	m.LastRequestTime = time.Now()
	m.TotalRequestDuration += duration
	
	if err != nil {
		m.Errors++
	}
	
	// Update averages
	if m.Requests > 0 {
		m.AverageResponseTime = time.Duration(m.TotalRequestDuration.Nanoseconds() / m.Requests)
	}
	
	// Calculate tokens per minute and requests per minute
	// based on the last hour of usage
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	if m.LastRequestTime.After(oneHourAgo) {
		elapsedMinutes := time.Since(oneHourAgo).Minutes()
		if elapsedMinutes > 0 {
			m.TokensPerMinute = float64(m.Tokens) / elapsedMinutes
			m.RequestsPerMinute = float64(m.Requests) / elapsedMinutes
		}
	}
}

// Reset resets the usage metrics
func (m *UsageMetrics) Reset() {
	m.Requests = 0
	m.Tokens = 0
	m.Errors = 0
	m.TotalRequestDuration = 0
	m.AverageResponseTime = 0
	m.TokensPerMinute = 0
	m.RequestsPerMinute = 0
}

// Merge merges another usage metrics into this one
func (m *UsageMetrics) Merge(other *UsageMetrics) {
	m.Requests += other.Requests
	m.Tokens += other.Tokens
	m.Errors += other.Errors
	m.TotalRequestDuration += other.TotalRequestDuration
	
	// Update last request time if the other is more recent
	if other.LastRequestTime.After(m.LastRequestTime) {
		m.LastRequestTime = other.LastRequestTime
	}
	
	// Recalculate averages
	if m.Requests > 0 {
		m.AverageResponseTime = time.Duration(m.TotalRequestDuration.Nanoseconds() / m.Requests)
	}
	
	// Recalculate tokens per minute and requests per minute
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	if m.LastRequestTime.After(oneHourAgo) {
		elapsedMinutes := time.Since(oneHourAgo).Minutes()
		if elapsedMinutes > 0 {
			m.TokensPerMinute = float64(m.Tokens) / elapsedMinutes
			m.RequestsPerMinute = float64(m.Requests) / elapsedMinutes
		}
	}
}
