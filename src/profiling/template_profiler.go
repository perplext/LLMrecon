// Package profiling provides tools for profiling and performance measurement.
package profiling

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

// TemplateOperation defines the type of template operation being profiled
type TemplateOperation string

const (
	// LoadOperation represents template loading
	LoadOperation TemplateOperation = "load"
	// ParseOperation represents template parsing
	ParseOperation TemplateOperation = "parse"
	// RenderOperation represents template rendering
	RenderOperation TemplateOperation = "render"
	// ExecuteOperation represents template execution
	ExecuteOperation TemplateOperation = "execute"
	// ValidationOperation represents template validation
	ValidationOperation TemplateOperation = "validate"
)

// TemplateProfiler profiles template operations
type TemplateProfiler struct {
	// profiler is the underlying profiler
	profiler *Profiler
	// templateManager is the template manager to profile
	templateManager types.TemplateManager
	// config is the profiler configuration
	config *TemplateProfilerConfig
	// mutex protects the profiler
	mutex sync.RWMutex
	// baselineMetrics stores baseline metrics for comparison
	baselineMetrics map[string]MetricSummary

// TemplateProfilerConfig contains configuration for the template profiler
type TemplateProfilerConfig struct {
	// ProfilerConfig is the configuration for the underlying profiler
	ProfilerConfig *ProfilerConfig
	// EnableDetailedProfiling enables detailed profiling
	EnableDetailedProfiling bool
	// EnableContinuousMonitoring enables continuous monitoring
	EnableContinuousMonitoring bool
	// MonitoringInterval is how often to collect metrics
	MonitoringInterval time.Duration
	// BaselineFilePath is the path to save baseline metrics
	BaselineFilePath string
	// ReportFilePath is the path to save profiling reports
	ReportFilePath string

// NewTemplateProfiler creates a new template profiler
func NewTemplateProfiler(templateManager types.TemplateManager, config *TemplateProfilerConfig) *TemplateProfiler {
	// Set default values
	if config == nil {
		config = &TemplateProfilerConfig{
			ProfilerConfig: &ProfilerConfig{
				EnableCPUProfiling: false,
				EnableMemProfiling: false,
				SamplingInterval:   1 * time.Second,
				MaxSamples:         1000,
				Tags:               make(map[string]string),
			},
			EnableDetailedProfiling:   false,
			EnableContinuousMonitoring: false,
			MonitoringInterval:        5 * time.Minute,
			BaselineFilePath:          "template_baseline.json",
			ReportFilePath:            "template_profile.txt",
		}
	}

	// Create profiler
	profiler := NewProfiler(config.ProfilerConfig)

	// Register template metrics
	profiler.RegisterMetric("template.load.time", TemplateLoadTime, "Time to load templates", Milliseconds)
	profiler.RegisterMetric("template.parse.time", TemplateRenderTime, "Time to parse templates", Milliseconds)
	profiler.RegisterMetric("template.render.time", TemplateRenderTime, "Time to render templates", Milliseconds)
	profiler.RegisterMetric("template.execute.time", ResponseTime, "Time to execute templates", Milliseconds)
	profiler.RegisterMetric("template.validate.time", ResponseTime, "Time to validate templates", Milliseconds)
	profiler.RegisterMetric("template.memory.usage", MemoryUsage, "Memory usage during template operations", Megabytes)
	profiler.RegisterMetric("template.throughput", ThroughputMetric, "Template operations per second", OpsPerSecond)
	profiler.RegisterMetric("template.error.rate", ErrorRateMetric, "Template error rate", Percentage)
	profiler.RegisterMetric("template.cache.hit_rate", CacheHitRateMetric, "Template cache hit rate", Percentage)
	profiler.RegisterMetric("template.goroutines", GoroutineCount, "Goroutine count during template operations", Count)

	return &TemplateProfiler{
		profiler:        profiler,
		templateManager: templateManager,
		config:          config,
		baselineMetrics: make(map[string]MetricSummary),
	}

// Start starts the profiler
func (p *TemplateProfiler) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Start underlying profiler
	if err := p.profiler.Start(); err != nil {
		return err
	}

	// Start continuous monitoring if enabled
	if p.config.EnableContinuousMonitoring {
		go p.startContinuousMonitoring()
	}

	return nil

// Stop stops the profiler
func (p *TemplateProfiler) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Stop underlying profiler
	return p.profiler.Stop()

// ProfileTemplateLoad profiles template loading
func (p *TemplateProfiler) ProfileTemplateLoad(ctx context.Context, source string, sourceType string) (*format.Template, error) {
	// Create labels
	labels := map[string]string{
		"source":      source,
		"source_type": sourceType,
		"operation":   string(LoadOperation),
	}

	// Start timer
	stop := p.profiler.StartTimer("template.load.time", labels)
	defer stop()

	// Record memory before
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)

	// Record goroutine count
	p.profiler.RecordGoroutineCount("template.goroutines", labels)

	// Load template
	template, err := p.templateManager.LoadTemplate(ctx, source, sourceType)

	// Record error if any
	if err != nil {
		p.profiler.RecordMetric("template.error.rate", 100.0, labels)
		return nil, err
	} else {
		p.profiler.RecordMetric("template.error.rate", 0.0, labels)
	}

	// Record memory after
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)

	return template, nil

// ProfileTemplateLoadBatch profiles batch template loading
func (p *TemplateProfiler) ProfileTemplateLoadBatch(ctx context.Context, source string, sourceType string) ([]*format.Template, error) {
	// Create labels
	labels := map[string]string{
		"source":      source,
		"source_type": sourceType,
		"operation":   string(LoadOperation),
		"batch":       "true",
	}

	// Start timer
	startTime := time.Now()
	stop := p.profiler.StartTimer("template.load.time", labels)
	defer stop()

	// Record memory before
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)

	// Record goroutine count
	p.profiler.RecordGoroutineCount("template.goroutines", labels)

	// Load templates
	templates, err := p.templateManager.LoadTemplates(ctx, source, sourceType)

	// Record error if any
	if err != nil {
		p.profiler.RecordMetric("template.error.rate", 100.0, labels)
		return nil, err
	} else {
		p.profiler.RecordMetric("template.error.rate", 0.0, labels)
	}

	// Record memory after
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)

	// Calculate throughput
	duration := time.Since(startTime)
	if duration > 0 {
		throughput := float64(len(templates)) / duration.Seconds()
		p.profiler.RecordMetric("template.throughput", throughput, labels)
	}

	return templates, nil

// ProfileTemplateExecution profiles template execution
func (p *TemplateProfiler) ProfileTemplateExecution(ctx context.Context, template *format.Template, options map[string]interface{}) (*interfaces.TemplateResult, error) {
	// Create labels
	labels := map[string]string{
		"template_id": template.ID,
		"operation":   string(ExecuteOperation),
	}

	// Start timer
	stop := p.profiler.StartTimer("template.execute.time", labels)
	defer stop()

	// Record memory before
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)
	// Record goroutine count
	p.profiler.RecordGoroutineCount("template.goroutines", labels)

	// Execute template
	result, err := p.templateManager.Execute(ctx, template, options)

	// Record error if any
	if err != nil {
		p.profiler.RecordMetric("template.error.rate", 100.0, labels)
		return nil, err
	} else {
		p.profiler.RecordMetric("template.error.rate", 0.0, labels)
	}

	// Record memory after
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)

	return result, nil

// ProfileTemplateExecutionBatch profiles batch template execution
func (p *TemplateProfiler) ProfileTemplateExecutionBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*interfaces.TemplateResult, error) {
	// Create labels
	labels := map[string]string{
		"template_count": fmt.Sprintf("%d", len(templates)),
		"operation":      string(ExecuteOperation),
		"batch":          "true",
	}

	// Start timer
	startTime := time.Now()
	stop := p.profiler.StartTimer("template.execute.time", labels)
	defer stop()

	// Record memory before
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)

	// Record goroutine count
	p.profiler.RecordGoroutineCount("template.goroutines", labels)

	// Execute templates
	results, err := p.templateManager.ExecuteBatch(ctx, templates, options)

	// Record error if any
	if err != nil {
		p.profiler.RecordMetric("template.error.rate", 100.0, labels)
		return nil, err
	} else {
		p.profiler.RecordMetric("template.error.rate", 0.0, labels)
	}

	// Record memory after
	p.profiler.RecordMemoryUsage("template.memory.usage", labels)

	// Calculate throughput
	duration := time.Since(startTime)
	if duration > 0 {
		throughput := float64(len(templates)) / duration.Seconds()
		p.profiler.RecordMetric("template.throughput", throughput, labels)
	}

	return results, nil

// EstablishBaseline establishes baseline metrics
func (p *TemplateProfiler) EstablishBaseline(ctx context.Context, sources []types.TemplateSource, iterations int) error {
	fmt.Println("Establishing baseline metrics...")

	// Reset profiler
	p.profiler = NewProfiler(p.config.ProfilerConfig)
	p.Start()

	// Load and execute templates from each source
	for _, source := range sources {
		fmt.Printf("Processing source: %s (%s)\n", source.Path, source.Type)

		for i := 0; i < iterations; i++ {
			// Load templates
			templates, err := p.ProfileTemplateLoadBatch(ctx, source.Path, source.Type)
			if err != nil {
				fmt.Printf("Error loading templates: %v\n", err)
				continue
			}

			fmt.Printf("Loaded %d templates\n", len(templates))

			// Execute templates
			if len(templates) > 0 {
				sampleSize := min(5, len(templates))
				_, err := p.ProfileTemplateExecutionBatch(ctx, templates[:sampleSize], nil)
				if err != nil {
					fmt.Printf("Error executing templates: %v\n", err)
				}
			}
		}
	}

	// Get metrics
	metrics := p.profiler.GetMetrics()

	// Store baseline metrics
	p.baselineMetrics = make(map[string]MetricSummary)
	for name, metric := range metrics {
		p.baselineMetrics[name] = metric.Summary
	}

	// Save baseline to file
	if err := p.saveBaseline(); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	// Stop profiler
	p.Stop()

	fmt.Println("Baseline established successfully")
	return nil

// saveBaseline saves baseline metrics to a file
func (p *TemplateProfiler) saveBaseline() error {
	// Create file
	file, err := os.Create(p.config.BaselineFilePath)
	if err != nil {
		return fmt.Errorf("failed to create baseline file: %w", err)
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Write header
	fmt.Fprintf(file, "Template Performance Baseline - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "=================================================\n\n")

	// Write metrics
	for name, summary := range p.baselineMetrics {
		fmt.Fprintf(file, "%s:\n", name)
		fmt.Fprintf(file, "  Min: %.2f\n", summary.Min)
		fmt.Fprintf(file, "  Max: %.2f\n", summary.Max)
		fmt.Fprintf(file, "  Mean: %.2f\n", summary.Mean)
		fmt.Fprintf(file, "  Median: %.2f\n", summary.Median)
		fmt.Fprintf(file, "  P95: %.2f\n", summary.P95)
		fmt.Fprintf(file, "  P99: %.2f\n", summary.P99)
		fmt.Fprintf(file, "  StdDev: %.2f\n", summary.StdDev)
		fmt.Fprintf(file, "  Count: %d\n\n", summary.Count)
	}

	// Write performance targets
	fmt.Fprintf(file, "Performance Targets:\n")
	fmt.Fprintf(file, "  Template Load Time: < 200ms\n")
	fmt.Fprintf(file, "  Template Execution Time: < 500ms\n")
	fmt.Fprintf(file, "  Template Throughput: > 10 templates/sec\n")
	fmt.Fprintf(file, "  Memory Usage: < 100MB\n")
	fmt.Fprintf(file, "  Error Rate: < 1%%\n")
	fmt.Fprintf(file, "  Cache Hit Rate: > 90%%\n")

	return nil

// CompareWithBaseline compares current metrics with baseline
func (p *TemplateProfiler) CompareWithBaseline() map[string]map[string]interface{} {
	// Get current metrics
	metrics := p.profiler.GetMetrics()

	// Compare with baseline
	comparison := make(map[string]map[string]interface{})
	for name, metric := range metrics {
		baseline, exists := p.baselineMetrics[name]
		if !exists {
			continue
		}

		// Calculate differences
		meanDiff := calculatePercentageDiff(metric.Summary.Mean, baseline.Mean)
		p95Diff := calculatePercentageDiff(metric.Summary.P95, baseline.P95)
		maxDiff := calculatePercentageDiff(metric.Summary.Max, baseline.Max)

		comparison[name] = map[string]interface{}{
			"current_mean":  metric.Summary.Mean,
			"baseline_mean": baseline.Mean,
			"mean_diff":     meanDiff,
			"current_p95":   metric.Summary.P95,
			"baseline_p95":  baseline.P95,
			"p95_diff":      p95Diff,
			"current_max":   metric.Summary.Max,
			"baseline_max":  baseline.Max,
			"max_diff":      maxDiff,
		}
	}

	return comparison

// SaveComparisonReport saves a comparison report to a file
func (p *TemplateProfiler) SaveComparisonReport(filePath string) error {
	// Get comparison
	comparison := p.CompareWithBaseline()

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create comparison report file: %w", err)
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Write header
	fmt.Fprintf(file, "Template Performance Comparison - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "=================================================\n\n")

	// Write comparison
	for name, data := range comparison {
		fmt.Fprintf(file, "%s:\n", name)
		fmt.Fprintf(file, "  Mean: %.2f -> %.2f (%+.2f%%)\n",
			data["baseline_mean"], data["current_mean"], data["mean_diff"])
		fmt.Fprintf(file, "  P95: %.2f -> %.2f (%+.2f%%)\n",
			data["baseline_p95"], data["current_p95"], data["p95_diff"])
		fmt.Fprintf(file, "  Max: %.2f -> %.2f (%+.2f%%)\n\n",
			data["baseline_max"], data["current_max"], data["max_diff"])
	}

	// Check performance targets
	fmt.Fprintf(file, "Performance Target Analysis:\n")
	metrics := p.profiler.GetMetrics()

	// Template load time
	if loadTimeMetric, exists := metrics["template.load.time"]; exists {
		loadTime := loadTimeMetric.Summary.P95
		fmt.Fprintf(file, "  Template Load Time (P95): %.2f ms (Target: < 200ms) - %s\n",
			loadTime, getStatusString(loadTime < 200))
	}

	// Template execution time
	if execTimeMetric, exists := metrics["template.execute.time"]; exists {
		execTime := execTimeMetric.Summary.P95
		fmt.Fprintf(file, "  Template Execution Time (P95): %.2f ms (Target: < 500ms) - %s\n",
			execTime, getStatusString(execTime < 500))
	}

	// Template throughput
	if throughputMetric, exists := metrics["template.throughput"]; exists {
		throughput := throughputMetric.Summary.Mean
		fmt.Fprintf(file, "  Template Throughput (Mean): %.2f templates/sec (Target: > 10) - %s\n",
			throughput, getStatusString(throughput > 10))
	}

	// Memory usage
	if memoryMetric, exists := metrics["template.memory.usage"]; exists {
		memory := memoryMetric.Summary.P95
		fmt.Fprintf(file, "  Memory Usage (P95): %.2f MB (Target: < 100MB) - %s\n",
			memory, getStatusString(memory < 100))
	}

	// Error rate
	if errorRateMetric, exists := metrics["template.error.rate"]; exists {
		errorRate := errorRateMetric.Summary.Mean
		fmt.Fprintf(file, "  Error Rate (Mean): %.2f%% (Target: < 1%%) - %s\n",
			errorRate, getStatusString(errorRate < 1))
	}

	// Cache hit rate
	if cacheHitRateMetric, exists := metrics["template.cache.hit_rate"]; exists {
		cacheHitRate := cacheHitRateMetric.Summary.Mean
		fmt.Fprintf(file, "  Cache Hit Rate (Mean): %.2f%% (Target: > 90%%) - %s\n",
			cacheHitRate, getStatusString(cacheHitRate > 90))
	}

	return nil

// startContinuousMonitoring starts continuous monitoring
func (p *TemplateProfiler) startContinuousMonitoring() {
	ticker := time.NewTicker(p.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Record memory usage
			p.profiler.RecordMemoryUsage("template.memory.usage", map[string]string{
				"operation": "monitoring",
			})

			// Record goroutine count
			p.profiler.RecordGoroutineCount("template.goroutines", map[string]string{
				"operation": "monitoring",
			})

			// Get cache stats if available
			if optimizedManager, ok := p.templateManager.(types.OptimizedTemplateManager); ok {
				stats := optimizedManager.GetStats()
				if cacheStats, ok := stats["loader_cache_stats"].(map[string]interface{}); ok {
					if hitRate, ok := cacheStats["hit_rate"].(float64); ok {
						p.profiler.RecordMetric("template.cache.hit_rate", hitRate, map[string]string{
							"operation": "monitoring",
						})
					}
				}
			}

			// Save report periodically
			if p.config.ReportFilePath != "" {
				p.profiler.SaveReport(p.config.ReportFilePath)
			}
		}
	}

// calculatePercentageDiff calculates the percentage difference between two values
func calculatePercentageDiff(current, baseline float64) float64 {
	if baseline == 0 {
		return 0
	}
	return ((current - baseline) / baseline) * 100

// getStatusString returns a status string based on a condition
func getStatusString(condition bool) string {
	if condition {
		return "✓ PASS"
	}
	return "✗ FAIL"

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
