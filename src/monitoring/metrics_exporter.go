package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsExporter exports metrics for Prometheus
type MetricsExporter struct {
	// Counters
	totalRequests      atomic.Int64
	successfulRequests atomic.Int64
	failedRequests     atomic.Int64
	
	// Gauges
	concurrentAttacks atomic.Int64
	queueDepth        atomic.Int64
	
	// Histograms (simplified)
	responseTimes []float64
	mu            sync.RWMutex
	
	// Start time for uptime calculation
	startTime time.Time
}

// NewMetricsExporter creates a new metrics exporter
func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{
		responseTimes: make([]float64, 0, 1000),
		startTime:     time.Now(),
	}
}

// RecordRequest records a request with its outcome
func (m *MetricsExporter) RecordRequest(success bool, responseTime float64) {
	m.totalRequests.Add(1)
	
	if success {
		m.successfulRequests.Add(1)
	} else {
		m.failedRequests.Add(1)
	}
	
	m.mu.Lock()
	m.responseTimes = append(m.responseTimes, responseTime)
	// Keep only last 1000 samples
	if len(m.responseTimes) > 1000 {
		m.responseTimes = m.responseTimes[len(m.responseTimes)-1000:]
	}
	m.mu.Unlock()
}

// SetConcurrentAttacks updates the current concurrent attacks count
func (m *MetricsExporter) SetConcurrentAttacks(count int64) {
	m.concurrentAttacks.Store(count)
}

// SetQueueDepth updates the current queue depth
func (m *MetricsExporter) SetQueueDepth(depth int64) {
	m.queueDepth.Store(depth)
}

// GetMetrics returns current metrics as JSON
func (m *MetricsExporter) GetMetrics() map[string]interface{} {
	total := m.totalRequests.Load()
	successful := m.successfulRequests.Load()
	
	successRate := float64(0)
	if total > 0 {
		successRate = float64(successful) / float64(total)
	}
	
	m.mu.RLock()
	avgResponseTime := m.calculateAverage(m.responseTimes)
	p95ResponseTime := m.calculatePercentile(m.responseTimes, 0.95)
	p99ResponseTime := m.calculatePercentile(m.responseTimes, 0.99)
	m.mu.RUnlock()
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return map[string]interface{}{
		"uptime_seconds": time.Since(m.startTime).Seconds(),
		"requests": map[string]interface{}{
			"total":      total,
			"successful": successful,
			"failed":     m.failedRequests.Load(),
		},
		"performance": map[string]interface{}{
			"success_rate":         successRate,
			"concurrent_attacks":   m.concurrentAttacks.Load(),
			"avg_response_time":    avgResponseTime,
			"p95_response_time":    p95ResponseTime,
			"p99_response_time":    p99ResponseTime,
			"queue_depth":          m.queueDepth.Load(),
		},
		"resources": map[string]interface{}{
			"cpu_count":        runtime.NumCPU(),
			"goroutines":       runtime.NumGoroutine(),
			"memory_alloc":     memStats.Alloc,
			"memory_total":     memStats.TotalAlloc,
			"memory_sys":       memStats.Sys,
			"gc_runs":          memStats.NumGC,
		},
	}
}

// ServeHTTP implements http.Handler for metrics endpoint
func (m *MetricsExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/metrics":
		m.servePrometheusMetrics(w, r)
	case "/api/v1/metrics":
		m.serveJSONMetrics(w, r)
	case "/api/v1/status":
		m.serveStatus(w, r)
	default:
		http.NotFound(w, r)
	}
}

// servePrometheusMetrics serves metrics in Prometheus format
func (m *MetricsExporter) servePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.GetMetrics()
	
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	
	// Write metrics in Prometheus format
	fmt.Fprintf(w, "# HELP llmrecon_uptime_seconds Time since application start\n")
	fmt.Fprintf(w, "# TYPE llmrecon_uptime_seconds counter\n")
	fmt.Fprintf(w, "llmrecon_uptime_seconds %f\n\n", metrics["uptime_seconds"])
	
	requests := metrics["requests"].(map[string]interface{})
	fmt.Fprintf(w, "# HELP llmrecon_requests_total Total number of requests\n")
	fmt.Fprintf(w, "# TYPE llmrecon_requests_total counter\n")
	fmt.Fprintf(w, "llmrecon_requests_total %d\n\n", requests["total"])
	
	performance := metrics["performance"].(map[string]interface{})
	fmt.Fprintf(w, "# HELP llmrecon_success_rate Current success rate\n")
	fmt.Fprintf(w, "# TYPE llmrecon_success_rate gauge\n")
	fmt.Fprintf(w, "llmrecon_success_rate %f\n\n", performance["success_rate"])
	
	fmt.Fprintf(w, "# HELP llmrecon_concurrent_attacks Current concurrent attacks\n")
	fmt.Fprintf(w, "# TYPE llmrecon_concurrent_attacks gauge\n")
	fmt.Fprintf(w, "llmrecon_concurrent_attacks %d\n\n", performance["concurrent_attacks"])
	
	fmt.Fprintf(w, "# HELP llmrecon_response_time_seconds Response time in seconds\n")
	fmt.Fprintf(w, "# TYPE llmrecon_response_time_seconds gauge\n")
	fmt.Fprintf(w, "llmrecon_response_time_seconds %f\n\n", performance["avg_response_time"])
	
	resources := metrics["resources"].(map[string]interface{})
	fmt.Fprintf(w, "# HELP llmrecon_goroutines Number of goroutines\n")
	fmt.Fprintf(w, "# TYPE llmrecon_goroutines gauge\n")
	fmt.Fprintf(w, "llmrecon_goroutines %d\n\n", resources["goroutines"])
	
	fmt.Fprintf(w, "# HELP llmrecon_memory_usage_bytes Current memory usage\n")
	fmt.Fprintf(w, "# TYPE llmrecon_memory_usage_bytes gauge\n")
	fmt.Fprintf(w, "llmrecon_memory_usage_bytes %d\n", resources["memory_alloc"])
}

// serveJSONMetrics serves metrics in JSON format
func (m *MetricsExporter) serveJSONMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.GetMetrics())
}

// serveStatus serves a simple health status
func (m *MetricsExporter) serveStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"uptime": time.Since(m.startTime).String(),
	})
}

// calculateAverage calculates the average of a slice of floats
func (m *MetricsExporter) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := float64(0)
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculatePercentile calculates the percentile of a slice of floats
func (m *MetricsExporter) calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Simple implementation - in production use a proper algorithm
	index := int(float64(len(values)) * percentile)
	if index >= len(values) {
		index = len(values) - 1
	}
	
	return values[index]
}

// StartMetricsServer starts the metrics HTTP server
func StartMetricsServer(ctx context.Context, addr string, exporter *MetricsExporter) error {
	mux := http.NewServeMux()
	mux.Handle("/", exporter)
	
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()
	
	return server.ListenAndServe()
}