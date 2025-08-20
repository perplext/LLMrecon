package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Profiler tracks and analyzes rate limiting performance
type Profiler struct {
	metrics        map[string]*MetricData
	mu             sync.RWMutex
	startTime      time.Time
	samplingRate   float64
	enabled        bool
	exportInterval time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
}

// MetricData holds performance metrics for rate limiting
type MetricData struct {
	RequestCount      int64         `json:"request_count"`
	AllowedCount      int64         `json:"allowed_count"`
	DeniedCount       int64         `json:"denied_count"`
	AverageLatency    time.Duration `json:"average_latency"`
	MaxLatency        time.Duration `json:"max_latency"`
	MinLatency        time.Duration `json:"min_latency"`
	TotalLatency      time.Duration `json:"total_latency"`
	ErrorCount        int64         `json:"error_count"`
	LastUpdated       time.Time     `json:"last_updated"`
	WindowStart       time.Time     `json:"window_start"`
	BucketUtilization float64       `json:"bucket_utilization"`
}

// ProfilerConfig configuration for the profiler
type ProfilerConfig struct {
	SamplingRate   float64
	ExportInterval time.Duration
	EnableExport   bool
	MetricWindow   time.Duration
}

// RequestProfile represents profiling data for a single request
type RequestProfile struct {
	Timestamp    time.Time     `json:"timestamp"`
	Key          string        `json:"key"`
	Allowed      bool          `json:"allowed"`
	Latency      time.Duration `json:"latency"`
	TokensUsed   int64         `json:"tokens_used"`
	BucketLevel  float64       `json:"bucket_level"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// AggregateStats represents aggregated profiling statistics
type AggregateStats struct {
	TotalRequests       int64                  `json:"total_requests"`
	AllowedRequests     int64                  `json:"allowed_requests"`
	DeniedRequests      int64                  `json:"denied_requests"`
	AverageLatency      time.Duration          `json:"average_latency"`
	ThroughputPerSecond float64                `json:"throughput_per_second"`
	ErrorRate           float64                `json:"error_rate"`
	TopKeys             []string               `json:"top_keys"`
	MetricsByKey        map[string]*MetricData `json:"metrics_by_key"`
	WindowDuration      time.Duration          `json:"window_duration"`
	GeneratedAt         time.Time              `json:"generated_at"`
}

// NewProfiler creates a new rate limit profiler
func NewProfiler(config ProfilerConfig) *Profiler {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Profiler{
		metrics:        make(map[string]*MetricData),
		startTime:      time.Now(),
		samplingRate:   config.SamplingRate,
		enabled:        true,
		exportInterval: config.ExportInterval,
		ctx:            ctx,
		cancel:         cancel,
	}

	if config.EnableExport && config.ExportInterval > 0 {
		go p.backgroundExport()
	}

	return p
}

// ProfileRequest records profiling data for a rate limit request
func (p *Profiler) ProfileRequest(profile RequestProfile) {
	if !p.enabled || !p.shouldSample() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	key := profile.Key
	metric, exists := p.metrics[key]
	if !exists {
		metric = &MetricData{
			WindowStart: time.Now(),
			MinLatency:  profile.Latency,
			MaxLatency:  profile.Latency,
		}
		p.metrics[key] = metric
	}

	// Update metrics
	metric.RequestCount++
	if profile.Allowed {
		metric.AllowedCount++
	} else {
		metric.DeniedCount++
	}

	if profile.ErrorMessage != "" {
		metric.ErrorCount++
	}

	// Update latency metrics
	metric.TotalLatency += profile.Latency
	metric.AverageLatency = time.Duration(int64(metric.TotalLatency) / metric.RequestCount)

	if profile.Latency > metric.MaxLatency {
		metric.MaxLatency = profile.Latency
	}
	if profile.Latency < metric.MinLatency {
		metric.MinLatency = profile.Latency
	}

	metric.BucketUtilization = profile.BucketLevel
	metric.LastUpdated = time.Now()
}

// shouldSample determines if the request should be sampled based on sampling rate
func (p *Profiler) shouldSample() bool {
	return p.samplingRate >= 1.0 || (time.Now().UnixNano()%1000)/10 < int64(p.samplingRate*100)
}

// GetMetrics returns current metrics for a specific key
func (p *Profiler) GetMetrics(key string) (*MetricData, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	metric, exists := p.metrics[key]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid data races
	metricCopy := *metric
	return &metricCopy, true
}

// GetAllMetrics returns all current metrics
func (p *Profiler) GetAllMetrics() map[string]*MetricData {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*MetricData, len(p.metrics))
	for key, metric := range p.metrics {
		metricCopy := *metric
		result[key] = &metricCopy
	}

	return result
}

// GetAggregateStats returns aggregated statistics
func (p *Profiler) GetAggregateStats() *AggregateStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &AggregateStats{
		MetricsByKey:   make(map[string]*MetricData),
		GeneratedAt:    time.Now(),
		WindowDuration: time.Since(p.startTime),
	}

	var totalLatency time.Duration
	topKeys := make([]string, 0, len(p.metrics))

	for key, metric := range p.metrics {
		stats.TotalRequests += metric.RequestCount
		stats.AllowedRequests += metric.AllowedCount
		stats.DeniedRequests += metric.DeniedCount
		totalLatency += metric.TotalLatency

		metricCopy := *metric
		stats.MetricsByKey[key] = &metricCopy
		topKeys = append(topKeys, key)
	}

	if stats.TotalRequests > 0 {
		stats.AverageLatency = time.Duration(int64(totalLatency) / stats.TotalRequests)
		stats.ErrorRate = float64(p.getTotalErrors()) / float64(stats.TotalRequests)
		stats.ThroughputPerSecond = float64(stats.TotalRequests) / stats.WindowDuration.Seconds()
	}

	stats.TopKeys = p.getTopKeys(topKeys, 10)

	return stats
}

// getTotalErrors calculates total error count across all metrics
func (p *Profiler) getTotalErrors() int64 {
	var total int64
	for _, metric := range p.metrics {
		total += metric.ErrorCount
	}
	return total
}

// getTopKeys returns the top N keys by request count
func (p *Profiler) getTopKeys(keys []string, limit int) []string {
	if len(keys) <= limit {
		return keys
	}

	// Simple sorting by request count (in real implementation, use proper sorting)
	result := make([]string, 0, limit)
	used := make(map[string]bool)

	for i := 0; i < limit && len(result) < limit; i++ {
		var maxKey string
		var maxCount int64

		for _, key := range keys {
			if used[key] {
				continue
			}

			if metric, exists := p.metrics[key]; exists && metric.RequestCount > maxCount {
				maxCount = metric.RequestCount
				maxKey = key
			}
		}

		if maxKey != "" {
			result = append(result, maxKey)
			used[maxKey] = true
		}
	}

	return result
}

// ExportMetrics exports current metrics to JSON
func (p *Profiler) ExportMetrics() ([]byte, error) {
	stats := p.GetAggregateStats()
	return json.MarshalIndent(stats, "", "  ")
}

// backgroundExport runs background metric export
func (p *Profiler) backgroundExport() {
	ticker := time.NewTicker(p.exportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if data, err := p.ExportMetrics(); err == nil {
				p.handleExportedData(data)
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// handleExportedData handles exported metric data (placeholder for actual implementation)
func (p *Profiler) handleExportedData(data []byte) {
	// In a real implementation, this would write to file, send to monitoring system, etc.
	fmt.Printf("Exported metrics: %d bytes\n", len(data))
}

// Reset clears all metrics and resets the profiler
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics = make(map[string]*MetricData)
	p.startTime = time.Now()
}

// SetSamplingRate updates the sampling rate
func (p *Profiler) SetSamplingRate(rate float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if rate >= 0.0 && rate <= 1.0 {
		p.samplingRate = rate
	}
}

// Enable enables profiling
func (p *Profiler) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Disable disables profiling
func (p *Profiler) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// IsEnabled returns whether profiling is enabled
func (p *Profiler) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// Close shuts down the profiler
func (p *Profiler) Close() error {
	p.cancel()
	p.Reset()
	return nil
}
