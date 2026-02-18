package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// MockFileHandlerOptions represents options for the mock file handler
type MockFileHandlerOptions struct {
	RootDir           string
	EnableCache       bool
	EnableCompression bool
	MaxCacheSize      int64
}

// DefaultMockFileHandlerOptions returns default options for the mock file handler
func DefaultMockFileHandlerOptions() *MockFileHandlerOptions {
	return &MockFileHandlerOptions{
		RootDir:           "./static",
		EnableCache:       true,
		EnableCompression: true,
		MaxCacheSize:      100 * 1024 * 1024, // 100MB
	}
}

// MockFileHandler represents a mock static file handler
type MockFileHandler struct {
	options *MockFileHandlerOptions
	cache   map[string][]byte
	stats   MockFileHandlerStats
	mutex   sync.RWMutex
}

// MockFileHandlerStats represents stats for the mock file handler
type MockFileHandlerStats struct {
	FilesServed      int64
	CacheHits        int64
	CacheMisses      int64
	CompressedFiles  int64
	TotalSize        int64
	CompressedSize   int64
	CompressionRatio float64
	AverageServeTime time.Duration
}

// NewMockFileHandler creates a new mock file handler
func NewMockFileHandler(options *MockFileHandlerOptions) *MockFileHandler {
	return &MockFileHandler{
		options: options,
		cache:   make(map[string][]byte),
		stats: MockFileHandlerStats{
			FilesServed:      0,
			CacheHits:        0,
			CacheMisses:      0,
			CompressedFiles:  0,
			TotalSize:        0,
			CompressedSize:   0,
			CompressionRatio: 0.0,
			AverageServeTime: 5 * time.Millisecond,
		},
	}
}

// ServeHTTP implements the http.Handler interface
func (h *MockFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Update stats
	h.stats.FilesServed++
	if h.options.EnableCache {
		h.stats.CacheHits++
	} else {
		h.stats.CacheMisses++
	}
	if h.options.EnableCompression {
		h.stats.CompressedFiles++
	}
	
	// Simple file serving logic for the example
	filePath := filepath.Join(h.options.RootDir, r.URL.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filePath)
}

// GetStats returns the current stats
func (h *MockFileHandler) GetStats() MockFileHandlerStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return h.stats
}

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

func main() {
	fmt.Println("Starting Static File Handler Example with Monitoring Integration")
	
	// Create monitoring service
	monitoringService := NewMockMonitoringService()
	
	// Take initial memory snapshot
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("Initial Memory Usage - Heap Alloc: %.2f MB\n", float64(memStats.HeapAlloc)/1024/1024)

	// Create static file handler options
	fileHandlerOptions := DefaultMockFileHandlerOptions()
	fileHandlerOptions.RootDir = "./static"
	fileHandlerOptions.EnableCache = true
	fileHandlerOptions.EnableCompression = true
	fileHandlerOptions.MaxCacheSize = 100 * 1024 * 1024 // 100MB

	// Create static file handler
	fileHandler := NewMockFileHandler(fileHandlerOptions)
	
	// Add static file handler to monitoring service
	staticFileMonitor := monitoringService.AddStaticFileMonitor(fileHandler)

	// Start the monitoring service
	monitoringService.Start()

	// Create static directory if it doesn't exist
	if _, err := os.Stat("./static"); os.IsNotExist(err) {
		_ = os.Mkdir("./static", 0750)
	}

	// Create some example static files
	createExampleFiles("./static", 10, 10240)

	// Set up HTTP server with the file handler
	http.Handle("/static/", http.StripPrefix("/static/", fileHandler))
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`
				<!DOCTYPE html>
				<html>
				<head>
					<title>Static File Handler Example</title>
					<link rel="stylesheet" href="/static/styles.css">
				</head>
				<body>
					<h1>Static File Handler Example</h1>
					<p>This example demonstrates the memory-optimized static file handler with monitoring integration.</p>
					<ul>
						<li><a href="/static/file1.txt">File 1</a></li>
						<li><a href="/static/file2.txt">File 2</a></li>
						<li><a href="/static/file3.txt">File 3</a></li>
					</ul>
					<div class="stats">
						<h2>Memory Usage</h2>
						<pre id="memory-stats">Loading...</pre>
					</div>
					<div class="stats">
						<h2>Monitoring Metrics</h2>
						<pre id="monitoring-stats">Loading...</pre>
					</div>
					<script src="/static/script.js"></script>
				</body>
				</html>
			`))
		}
	})
	
	// Add a monitoring endpoint
	http.HandleFunc("/monitoring", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		
		// Get monitoring metrics
		metrics := staticFileMonitor.GetMetrics()
		
		// Format the metrics in a user-friendly way
		_, _ = w.Write([]byte(fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Static File Handler Monitoring</title>
				<link rel="stylesheet" href="/static/styles.css">
				<meta http-equiv="refresh" content="5">
			</head>
			<body>
				<h1>Static File Handler Monitoring</h1>
				<div class="stats">
					<h2>File Serving Statistics</h2>
					<table>
						<tr><th>Metric</th><th>Value</th></tr>
						<tr><td>Files Served</td><td>%d</td></tr>
						<tr><td>Cache Hits</td><td>%d</td></tr>
						<tr><td>Cache Misses</td><td>%d</td></tr>
						<tr><td>Cache Hit Ratio</td><td>%.2f%%</td></tr>
						<tr><td>Compressed Files</td><td>%d</td></tr>
						<tr><td>Compression Ratio</td><td>%.2f%%</td></tr>
						<tr><td>Average Serve Time</td><td>%s</td></tr>
					</table>
				</div>
				<div class="stats">
					<h2>Cache Statistics</h2>
					<table>
						<tr><th>Metric</th><th>Value</th></tr>
						<tr><td>Cache Size</td><td>%s</td></tr>
						<tr><td>Cache Items</td><td>%d</td></tr>
					</table>
				</div>
				<p><a href="/">Back to Home</a></p>
			</body>
			</html>
		`, 
			metrics.FilesServed,
			metrics.CacheHits,
			metrics.CacheMisses,
			metrics.CacheHitRatio * 100,
			metrics.CompressedFiles,
			metrics.CompressionRatio * 100,
			metrics.AverageServeTime.String(),
			formatBytes(metrics.CacheSize),
			metrics.CacheItemCount,
		)))
	})

	// Add a stats endpoint
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Get memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		
		// Get file handler stats
		fileStats := fileHandler.GetStats()
		
		// Get monitoring metrics
		monitoringMetrics := staticFileMonitor.GetMetrics()
		
		// Create response
		response := map[string]interface{}{
			"heapAlloc": float64(memStats.HeapAlloc) / 1024 / 1024,
			"heapObjects": memStats.HeapObjects,
			"gcCPUFraction": memStats.GCCPUFraction,
			"filesServed": fileStats.FilesServed,
			"cacheHits": fileStats.CacheHits,
			"cacheMisses": fileStats.CacheMisses,
			"monitoring": map[string]interface{}{
				"cacheHitRatio": monitoringMetrics.CacheHitRatio,
				"compressionRatio": monitoringMetrics.CompressionRatio,
				"averageServeTimeMs": monitoringMetrics.AverageServeTime.Milliseconds(),
				"cacheSize": monitoringMetrics.CacheSize,
				"cacheItemCount": monitoringMetrics.CacheItemCount,
			},
		}
		
		// Encode response
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Create CSS and JS files
	createCSSFile("./static/styles.css")
	createJSFile("./static/script.js")

	// Start HTTP server
	fmt.Println("Starting HTTP server on localhost:8080")
	fmt.Println("Visit http://localhost:8080 to see the example")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

// createExampleFiles creates example static files for testing
func createExampleFiles(dir string, numFiles, fileSize int) {
	for i := 1; i <= numFiles; i++ {
		filePath := filepath.Join(dir, fmt.Sprintf("file%d.txt", i))
		content := generateRandomContent(fileSize)
		_ = os.WriteFile(filepath.Clean(filePath), []byte(content), 0600)
	}
}

// generateRandomContent generates random content for static files
func generateRandomContent(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	content := make([]byte, size)
	for i := 0; i < size; i++ {
		content[i] = charset[i%len(charset)]
	}
	return string(content)
}

// createCSSFile creates a CSS file for the example
func createCSSFile(filePath string) {
	content := `
body {
    font-family: Arial, sans-serif;
    line-height: 1.6;
    margin: 0;
    padding: 20px;
    color: #333;
    max-width: 800px;
    margin: 0 auto;
}

h1 {
    color: #2c3e50;
    border-bottom: 2px solid #3498db;
    padding-bottom: 10px;
}

ul {
    list-style-type: none;
    padding: 0;
}

li {
    margin-bottom: 10px;
}

a {
    color: #3498db;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

.stats {
    margin-top: 30px;
    padding: 20px;
    background-color: #f8f9fa;
    border-radius: 5px;
}

pre {
    background-color: #f1f1f1;
    padding: 15px;
    border-radius: 5px;
    overflow-x: auto;
}

table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 10px;
}

th, td {
    padding: 8px;
    text-align: left;
    border-bottom: 1px solid #ddd;
}

th {
    background-color: #f2f2f2;
}
`
	os.WriteFile(filepath.Clean(filePath), []byte(content), 0600)
}

// createJSFile creates a JavaScript file for the example
func createJSFile(filePath string) {
	content := `
// JavaScript for the static file handler example
document.addEventListener('DOMContentLoaded', function() {
	const memoryStatsElement = document.getElementById('memory-stats');
	const monitoringStatsElement = document.getElementById('monitoring-stats');
	
	function updateStats() {
		fetch('/stats')
			.then(response => response.json())
			.then(data => {
				// Update memory stats
				const memoryData = {
					heapAlloc: data.heapAlloc + ' MB',
					heapObjects: data.heapObjects,
					gcCPUFraction: data.gcCPUFraction
				};
				memoryStatsElement.innerHTML = JSON.stringify(memoryData, null, 2);
				
				// Update monitoring stats
				const monitoringData = {
					filesServed: data.filesServed,
					cacheHits: data.cacheHits,
					cacheMisses: data.cacheMisses,
					cacheHitRatio: (data.monitoring.cacheHitRatio * 100).toFixed(2) + '%',
					averageServeTime: data.monitoring.averageServeTimeMs + ' ms'
				};
				monitoringStatsElement.innerHTML = JSON.stringify(monitoringData, null, 2);
			})
			.catch(error => {
				console.error('Error fetching stats:', error);
			});
	}
	
	// Update stats initially and then every 2 seconds
	updateStats();
	setInterval(updateStats, 2000);
});
`
	os.WriteFile(filepath.Clean(filePath), []byte(content), 0600)
}

// formatBytes formats bytes to a human-readable string (KB, MB, GB)
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
