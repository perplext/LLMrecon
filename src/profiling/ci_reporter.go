// Package profiling provides tools for profiling and performance measurement.
package profiling

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CIReporter generates performance reports for CI/CD pipelines
type CIReporter struct {
	// profiler is the underlying profiler
	profiler *Profiler
	// templateProfiler is the template profiler
	templateProfiler *TemplateProfiler
	// config is the reporter configuration
	config *CIReporterConfig
	// baselineData stores baseline metrics for comparison
	baselineData map[string]interface{}
	// currentData stores current metrics
	currentData map[string]interface{}
}

// CIReporterConfig contains configuration for the CI reporter
type CIReporterConfig struct {
	// ReportDir is the directory for reports
	ReportDir string
	// BaselineFile is the path to the baseline file
	BaselineFile string
	// ThresholdFile is the path to the threshold file
	ThresholdFile string
	// ReportFormats is a list of report formats to generate
	ReportFormats []string
	// FailOnThresholdExceeded determines if the CI should fail when thresholds are exceeded
	FailOnThresholdExceeded bool
	// PerformanceThresholds defines thresholds for metrics
	PerformanceThresholds map[string]float64
}

// NewCIReporter creates a new CI reporter
func NewCIReporter(profiler *Profiler, templateProfiler *TemplateProfiler, config *CIReporterConfig) *CIReporter {
	// Set default values
	if config == nil {
		config = &CIReporterConfig{
			ReportDir:              "performance-reports",
			BaselineFile:           "baseline.json",
			ThresholdFile:          "thresholds.json",
			ReportFormats:          []string{"json", "txt", "html"},
			FailOnThresholdExceeded: true,
			PerformanceThresholds: map[string]float64{
				"template.load.time":     200,  // 200ms
				"template.execute.time":  500,  // 500ms
				"template.throughput":    10,   // 10 ops/sec
				"template.memory.usage":  100,  // 100MB
				"template.error.rate":    1,    // 1%
				"template.cache.hit_rate": 90,  // 90%
			},
		}
	}

	return &CIReporter{
		profiler:        profiler,
		templateProfiler: templateProfiler,
		config:          config,
		baselineData:    make(map[string]interface{}),
		currentData:     make(map[string]interface{}),
	}
}

// LoadBaseline loads baseline metrics from a file
func (r *CIReporter) LoadBaseline() error {
	// Check if baseline file exists
	if _, err := os.Stat(r.config.BaselineFile); os.IsNotExist(err) {
		return fmt.Errorf("baseline file %s does not exist", r.config.BaselineFile)
	}

	// Open baseline file
	file, err := os.Open(r.config.BaselineFile)
	if err != nil {
		return fmt.Errorf("failed to open baseline file: %w", err)
	}
	defer file.Close()

	// Decode baseline data
	if err := json.NewDecoder(file).Decode(&r.baselineData); err != nil {
		return fmt.Errorf("failed to decode baseline data: %w", err)
	}

	return nil
}

// SaveBaseline saves baseline metrics to a file
func (r *CIReporter) SaveBaseline() error {
	// Create baseline data
	r.baselineData = r.profiler.GetReport()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(r.config.BaselineFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create baseline file
	file, err := os.Create(r.config.BaselineFile)
	if err != nil {
		return fmt.Errorf("failed to create baseline file: %w", err)
	}
	defer file.Close()

	// Encode baseline data
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.baselineData); err != nil {
		return fmt.Errorf("failed to encode baseline data: %w", err)
	}

	return nil
}

// LoadThresholds loads performance thresholds from a file
func (r *CIReporter) LoadThresholds() error {
	// Check if threshold file exists
	if _, err := os.Stat(r.config.ThresholdFile); os.IsNotExist(err) {
		return fmt.Errorf("threshold file %s does not exist", r.config.ThresholdFile)
	}

	// Open threshold file
	file, err := os.Open(r.config.ThresholdFile)
	if err != nil {
		return fmt.Errorf("failed to open threshold file: %w", err)
	}
	defer file.Close()

	// Decode threshold data
	var thresholds map[string]float64
	if err := json.NewDecoder(file).Decode(&thresholds); err != nil {
		return fmt.Errorf("failed to decode threshold data: %w", err)
	}

	// Update thresholds
	r.config.PerformanceThresholds = thresholds

	return nil
}

// SaveThresholds saves performance thresholds to a file
func (r *CIReporter) SaveThresholds() error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(r.config.ThresholdFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create threshold file
	file, err := os.Create(r.config.ThresholdFile)
	if err != nil {
		return fmt.Errorf("failed to create threshold file: %w", err)
	}
	defer file.Close()

	// Encode threshold data
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.config.PerformanceThresholds); err != nil {
		return fmt.Errorf("failed to encode threshold data: %w", err)
	}

	return nil
}

// GenerateReports generates performance reports
func (r *CIReporter) GenerateReports() error {
	// Create report directory if it doesn't exist
	if err := os.MkdirAll(r.config.ReportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Get current data
	r.currentData = r.profiler.GetReport()

	// Generate reports in requested formats
	for _, format := range r.config.ReportFormats {
		switch strings.ToLower(format) {
		case "json":
			if err := r.generateJSONReport(); err != nil {
				return err
			}
		case "txt":
			if err := r.generateTextReport(); err != nil {
				return err
			}
		case "html":
			if err := r.generateHTMLReport(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported report format: %s", format)
		}
	}

	// Generate threshold report
	if err := r.generateThresholdReport(); err != nil {
		return err
	}

	// Generate comparison report if baseline is available
	if len(r.baselineData) > 0 {
		if err := r.generateComparisonReport(); err != nil {
			return err
		}
	}

	return nil
}

// generateJSONReport generates a JSON report
func (r *CIReporter) generateJSONReport() error {
	// Create report file
	reportPath := filepath.Join(r.config.ReportDir, "performance-report.json")
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON report file: %w", err)
	}
	defer file.Close()

	// Encode report data
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.currentData); err != nil {
		return fmt.Errorf("failed to encode JSON report data: %w", err)
	}

	return nil
}

// generateTextReport generates a text report
func (r *CIReporter) generateTextReport() error {
	// Create report file
	reportPath := filepath.Join(r.config.ReportDir, "performance-report.txt")
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create text report file: %w", err)
	}
	defer file.Close()

	// Write report header
	fmt.Fprintf(file, "Performance Report - %s\n", r.currentData["timestamp"])
	fmt.Fprintf(file, "=================================================\n\n")
	fmt.Fprintf(file, "Uptime: %s\n", r.currentData["uptime"])
	fmt.Fprintf(file, "Metric Count: %d\n", r.currentData["metric_count"])
	fmt.Fprintf(file, "Goroutines: %d\n", r.currentData["goroutines"])

	// Write memory statistics
	memStats := r.currentData["memory"].(map[string]interface{})
	fmt.Fprintf(file, "\nMemory Statistics:\n")
	fmt.Fprintf(file, "  Alloc: %.2f MB\n", float64(memStats["alloc"].(float64))/(1024*1024))
	fmt.Fprintf(file, "  Total Alloc: %.2f MB\n", float64(memStats["total_alloc"].(float64))/(1024*1024))
	fmt.Fprintf(file, "  Sys: %.2f MB\n", float64(memStats["sys"].(float64))/(1024*1024))
	fmt.Fprintf(file, "  GC Cycles: %d\n", int(memStats["num_gc"].(float64)))

	// Write metrics
	fmt.Fprintf(file, "\nMetrics:\n")
	metrics := r.currentData["metrics"].(map[string]interface{})
	for name, metricData := range metrics {
		metric := metricData.(map[string]interface{})
		fmt.Fprintf(file, "\n%s (%s):\n", name, metric["description"])
		fmt.Fprintf(file, "  Type: %s\n", metric["type"])
		fmt.Fprintf(file, "  Unit: %s\n", metric["unit"])
		fmt.Fprintf(file, "  Min: %.2f\n", metric["min"])
		fmt.Fprintf(file, "  Max: %.2f\n", metric["max"])
		fmt.Fprintf(file, "  Mean: %.2f\n", metric["mean"])
		fmt.Fprintf(file, "  Median: %.2f\n", metric["median"])
		fmt.Fprintf(file, "  P95: %.2f\n", metric["p95"])
		fmt.Fprintf(file, "  P99: %.2f\n", metric["p99"])
		fmt.Fprintf(file, "  StdDev: %.2f\n", metric["std_dev"])
		fmt.Fprintf(file, "  Count: %d\n", int(metric["count"].(float64)))

		// Write recent values
		recentValues := metric["recent_values"].([]interface{})
		if len(recentValues) > 0 {
			fmt.Fprintf(file, "  Recent Values:\n")
			for _, v := range recentValues {
				value := v.(map[string]interface{})
				fmt.Fprintf(file, "    %.2f (%s)\n", value["value"], value["timestamp"])
			}
		}
	}

	return nil
}

// generateHTMLReport generates an HTML report
func (r *CIReporter) generateHTMLReport() error {
	// Create report file
	reportPath := filepath.Join(r.config.ReportDir, "performance-report.html")
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML report file: %w", err)
	}
	defer file.Close()

	// Write HTML header
	fmt.Fprintf(file, `<!DOCTYPE html>
<html>
<head>
    <title>Performance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2, h3 { color: #333; }
        table { border-collapse: collapse; width: 100%%; margin-bottom: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .pass { color: green; }
        .fail { color: red; }
        .chart { width: 100%%; height: 300px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>Performance Report</h1>
    <p><strong>Date:</strong> %s</p>
    <p><strong>Uptime:</strong> %s</p>
    <p><strong>Metric Count:</strong> %d</p>
    <p><strong>Goroutines:</strong> %d</p>

    <h2>Memory Statistics</h2>
    <table>
        <tr>
            <th>Metric</th>
            <th>Value</th>
        </tr>
`, r.currentData["timestamp"], r.currentData["uptime"], int(r.currentData["metric_count"].(float64)), int(r.currentData["goroutines"].(float64)))

	// Write memory statistics
	memStats := r.currentData["memory"].(map[string]interface{})
	fmt.Fprintf(file, `
        <tr>
            <td>Alloc</td>
            <td>%.2f MB</td>
        </tr>
        <tr>
            <td>Total Alloc</td>
            <td>%.2f MB</td>
        </tr>
        <tr>
            <td>Sys</td>
            <td>%.2f MB</td>
        </tr>
        <tr>
            <td>GC Cycles</td>
            <td>%d</td>
        </tr>
    </table>
`, float64(memStats["alloc"].(float64))/(1024*1024), float64(memStats["total_alloc"].(float64))/(1024*1024), float64(memStats["sys"].(float64))/(1024*1024), int(memStats["num_gc"].(float64)))

	// Write metrics
	fmt.Fprintf(file, `
    <h2>Metrics</h2>
`)

	metrics := r.currentData["metrics"].(map[string]interface{})
	for name, metricData := range metrics {
		metric := metricData.(map[string]interface{})
		fmt.Fprintf(file, `
    <h3>%s (%s)</h3>
    <table>
        <tr>
            <th>Metric</th>
            <th>Value</th>
        </tr>
        <tr>
            <td>Type</td>
            <td>%s</td>
        </tr>
        <tr>
            <td>Unit</td>
            <td>%s</td>
        </tr>
        <tr>
            <td>Min</td>
            <td>%.2f</td>
        </tr>
        <tr>
            <td>Max</td>
            <td>%.2f</td>
        </tr>
        <tr>
            <td>Mean</td>
            <td>%.2f</td>
        </tr>
        <tr>
            <td>Median</td>
            <td>%.2f</td>
        </tr>
        <tr>
            <td>P95</td>
            <td>%.2f</td>
        </tr>
        <tr>
            <td>P99</td>
            <td>%.2f</td>
        </tr>
        <tr>
            <td>StdDev</td>
            <td>%.2f</td>
        </tr>
        <tr>
            <td>Count</td>
            <td>%d</td>
        </tr>
    </table>
`, name, metric["description"], metric["type"], metric["unit"], metric["min"], metric["max"], metric["mean"], metric["median"], metric["p95"], metric["p99"], metric["std_dev"], int(metric["count"].(float64)))

		// Write recent values
		recentValues := metric["recent_values"].([]interface{})
		if len(recentValues) > 0 {
			fmt.Fprintf(file, `
    <h4>Recent Values</h4>
    <table>
        <tr>
            <th>Value</th>
            <th>Timestamp</th>
        </tr>
`)
			for _, v := range recentValues {
				value := v.(map[string]interface{})
				fmt.Fprintf(file, `
        <tr>
            <td>%.2f</td>
            <td>%s</td>
        </tr>
`, value["value"], value["timestamp"])
			}
			fmt.Fprintf(file, `
    </table>
`)
		}
	}

	// Write HTML footer
	fmt.Fprintf(file, `
</body>
</html>
`)

	return nil
}

// generateThresholdReport generates a threshold report
func (r *CIReporter) generateThresholdReport() error {
	// Create report file
	reportPath := filepath.Join(r.config.ReportDir, "threshold-report.txt")
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create threshold report file: %w", err)
	}
	defer file.Close()

	// Write report header
	fmt.Fprintf(file, "Performance Threshold Report - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "=================================================\n\n")

	// Check thresholds
	metrics := r.currentData["metrics"].(map[string]interface{})
	thresholdExceeded := false

	for name, threshold := range r.config.PerformanceThresholds {
		if metricData, exists := metrics[name]; exists {
			metric := metricData.(map[string]interface{})
			value := metric["p95"].(float64)
			
			// For some metrics, higher is better (e.g., throughput, cache hit rate)
			higherIsBetter := false
			if strings.Contains(name, "throughput") || strings.Contains(name, "hit_rate") {
				higherIsBetter = true
			}

			// Check if threshold is exceeded
			exceeded := false
			if higherIsBetter {
				exceeded = value < threshold
			} else {
				exceeded = value > threshold
			}

			if exceeded {
				thresholdExceeded = true
				fmt.Fprintf(file, "❌ %s: %.2f %s (Threshold: %.2f) - EXCEEDED\n", 
					name, value, metric["unit"], threshold)
			} else {
				fmt.Fprintf(file, "✅ %s: %.2f %s (Threshold: %.2f) - OK\n", 
					name, value, metric["unit"], threshold)
			}
		}
	}

	// Write summary
	fmt.Fprintf(file, "\nSummary: ")
	if thresholdExceeded {
		fmt.Fprintf(file, "❌ Some thresholds were exceeded\n")
		
		// Exit with error if configured to fail on threshold exceeded
		if r.config.FailOnThresholdExceeded {
			fmt.Fprintf(file, "\nCI/CD pipeline should fail due to exceeded thresholds\n")
		}
	} else {
		fmt.Fprintf(file, "✅ All thresholds are within acceptable limits\n")
	}

	return nil
}

// generateComparisonReport generates a comparison report
func (r *CIReporter) generateComparisonReport() error {
	// Create report file
	reportPath := filepath.Join(r.config.ReportDir, "comparison-report.txt")
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create comparison report file: %w", err)
	}
	defer file.Close()

	// Write report header
	fmt.Fprintf(file, "Performance Comparison Report - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "=================================================\n\n")

	// Compare metrics
	currentMetrics := r.currentData["metrics"].(map[string]interface{})
	baselineMetrics, ok := r.baselineData["metrics"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid baseline data format")
	}

	for name, currentData := range currentMetrics {
		baselineData, exists := baselineMetrics[name]
		if !exists {
			fmt.Fprintf(file, "%s: No baseline data available\n", name)
			continue
		}

		current := currentData.(map[string]interface{})
		baseline := baselineData.(map[string]interface{})

		// Calculate differences
		meanDiff := calculatePercentageDiff(current["mean"].(float64), baseline["mean"].(float64))
		p95Diff := calculatePercentageDiff(current["p95"].(float64), baseline["p95"].(float64))
		maxDiff := calculatePercentageDiff(current["max"].(float64), baseline["max"].(float64))

		fmt.Fprintf(file, "%s (%s):\n", name, current["description"])
		fmt.Fprintf(file, "  Mean: %.2f -> %.2f (%+.2f%%)\n", 
			baseline["mean"].(float64), current["mean"].(float64), meanDiff)
		fmt.Fprintf(file, "  P95: %.2f -> %.2f (%+.2f%%)\n", 
			baseline["p95"].(float64), current["p95"].(float64), p95Diff)
		fmt.Fprintf(file, "  Max: %.2f -> %.2f (%+.2f%%)\n", 
			baseline["max"].(float64), current["max"].(float64), maxDiff)
		
		// Determine if performance improved or degraded
		// For some metrics, higher is better (e.g., throughput, cache hit rate)
		higherIsBetter := false
		if strings.Contains(name, "throughput") || strings.Contains(name, "hit_rate") {
			higherIsBetter = true
		}

		if (higherIsBetter && meanDiff > 0) || (!higherIsBetter && meanDiff < 0) {
			fmt.Fprintf(file, "  Performance improved ✅\n\n")
		} else if (higherIsBetter && meanDiff < 0) || (!higherIsBetter && meanDiff > 0) {
			fmt.Fprintf(file, "  Performance degraded ❌\n\n")
		} else {
			fmt.Fprintf(file, "  Performance unchanged ⚠️\n\n")
		}
	}

	return nil
}

// CheckThresholds checks if any performance thresholds are exceeded
func (r *CIReporter) CheckThresholds() (bool, map[string]interface{}) {
	// Get current metrics
	metrics := r.currentData["metrics"].(map[string]interface{})
	
	// Check thresholds
	thresholdExceeded := false
	results := make(map[string]interface{})

	for name, threshold := range r.config.PerformanceThresholds {
		if metricData, exists := metrics[name]; exists {
			metric := metricData.(map[string]interface{})
			value := metric["p95"].(float64)
			
			// For some metrics, higher is better (e.g., throughput, cache hit rate)
			higherIsBetter := false
			if strings.Contains(name, "throughput") || strings.Contains(name, "hit_rate") {
				higherIsBetter = true
			}

			// Check if threshold is exceeded
			exceeded := false
			if higherIsBetter {
				exceeded = value < threshold
			} else {
				exceeded = value > threshold
			}

			if exceeded {
				thresholdExceeded = true
			}

			results[name] = map[string]interface{}{
				"value":      value,
				"threshold":  threshold,
				"unit":       metric["unit"],
				"exceeded":   exceeded,
				"higher_is_better": higherIsBetter,
			}
		}
	}

	return thresholdExceeded, results
}
