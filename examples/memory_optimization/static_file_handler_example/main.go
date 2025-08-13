package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/perplext/LLMrecon/src/utils/monitoring"
	"github.com/perplext/LLMrecon/src/utils/profiling"
	"github.com/perplext/LLMrecon/src/utils/static"
)

func main() {
	// Create a memory profiler to monitor memory usage
	profilerOptions := profiling.DefaultProfilerOptions()
	profiler, err := profiling.NewMemoryProfiler(profilerOptions)
	if err != nil {
		log.Fatalf("Failed to create memory profiler: %v", err)
	}

	// Create monitoring service
	monitoringOptions := monitoring.DefaultMonitoringServiceOptions()
	monitoringService, err := monitoring.NewMonitoringService(monitoringOptions)
	if err != nil {
		log.Fatalf("Failed to create monitoring service: %v", err)
	}

	// Take initial memory snapshot
	profiler.CreateSnapshot("initial")
	initialMemStats := profiler.GetFormattedMemoryStats()
	fmt.Printf("Initial Memory Usage - Heap Alloc: %.2f MB\n", initialMemStats["heap_alloc_mb"].(float64))

	// Create static file handler options
	fileHandlerOptions := &static.FileHandlerOptions{
		RootDir:           "./static",
		EnableCache:       true,
		EnableCompression: true,
		MaxCacheSize:      100 * 1024 * 1024, // 100MB
		CacheExpiration:   time.Hour,
		MinCompressSize:   1024, // 1KB
		CompressExtensions: []string{
			".html", ".css", ".js", ".json", ".xml", ".txt", ".md",
		},
	}

	// Create static file handler
	fileHandler := static.NewFileHandler(fileHandlerOptions)
	
	// Add static file handler to monitoring service
	staticFileMonitor := monitoringService.AddStaticFileMonitor(fileHandler)

	// Create static directory if it doesn't exist
	if _, err := os.Stat("./static"); os.IsNotExist(err) {
		os.Mkdir("./static", 0755)
	}

	// Create some example static files
	createExampleFiles("./static", 10, 10240)

	// Set up HTTP server with the file handler
	http.Handle("/static/", http.StripPrefix("/static/", fileHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
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
		w.Write([]byte(fmt.Sprintf(`
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

	// Start the server
	fmt.Println("Server started at http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop the server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// createExampleFiles creates example static files for testing
func createExampleFiles(dir string, numFiles, fileSize int) {
	for i := 1; i <= numFiles; i++ {
		filePath := filepath.Join(dir, fmt.Sprintf("file%d.txt", i))
		content := generateRandomContent(fileSize)
		os.WriteFile(filePath, []byte(content), 0644)
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
`
	os.WriteFile(filePath, []byte(content), 0644)
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

	os.WriteFile(filePath, []byte(content), 0644)
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
