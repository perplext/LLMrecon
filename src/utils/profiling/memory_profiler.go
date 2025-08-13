package profiling

import (
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
)

// MemoryProfiler provides tools for memory profiling and optimization
type MemoryProfiler struct {
	// outputDir is the directory where profile files are saved
	outputDir string
	// interval is the time between automatic profile captures
	interval time.Duration
	// running indicates if automatic profiling is running
	running bool
	// stopChan is used to stop automatic profiling
	stopChan chan struct{}
	// mutex protects the profiler state
	mutex sync.Mutex
	// memStats stores the last memory statistics
	memStats runtime.MemStats
	// lastCapture is the time of the last profile capture
	lastCapture time.Time
	// profileCount is the number of profiles captured
	profileCount int
	// memoryThreshold is the threshold for memory usage alerts (in MB)
	memoryThreshold int64
	// gcThreshold is the threshold for GC pause time alerts (in ms)
	gcThreshold int64
	// alerts stores memory usage alerts
	alerts []string
	// snapshots stores memory snapshots for comparison
	snapshots map[string]*runtime.MemStats
}

// ProfilerOptions represents options for the memory profiler
type ProfilerOptions struct {
	// OutputDir is the directory where profile files are saved
	OutputDir string
	// Interval is the time between automatic profile captures
	Interval time.Duration
	// MemoryThreshold is the threshold for memory usage alerts (in MB)
	MemoryThreshold int64
	// GCThreshold is the threshold for GC pause time alerts (in ms)
	GCThreshold int64
}

// DefaultProfilerOptions returns default profiler options
func DefaultProfilerOptions() *ProfilerOptions {
	return &ProfilerOptions{
		OutputDir:       "profiles",
		Interval:        5 * time.Minute,
		MemoryThreshold: 100, // 100 MB
		GCThreshold:     100, // 100 ms
	}
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(options *ProfilerOptions) (*MemoryProfiler, error) {
	if options == nil {
		options = DefaultProfilerOptions()
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &MemoryProfiler{
		outputDir:       options.OutputDir,
		interval:        options.Interval,
		memoryThreshold: options.MemoryThreshold,
		gcThreshold:     options.GCThreshold,
		stopChan:        make(chan struct{}),
		snapshots:       make(map[string]*runtime.MemStats),
	}, nil
}

// CaptureHeapProfile captures a heap profile
func (p *MemoryProfiler) CaptureHeapProfile(label string) (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Update profile count
	p.profileCount++

	// Create profile file
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("heap-%s-%s-%d.pprof", label, timestamp, p.profileCount)
	filepath := filepath.Join(p.outputDir, filename)

	// Create file
	f, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create profile file: %w", err)
	}
	defer f.Close()

	// Write heap profile
	if err := pprof.WriteHeapProfile(f); err != nil {
		return "", fmt.Errorf("failed to write heap profile: %w", err)
	}

	// Update last capture time
	p.lastCapture = time.Now()

	return filepath, nil
}

// StartAutomaticProfiling starts automatic profiling
func (p *MemoryProfiler) StartAutomaticProfiling() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return fmt.Errorf("automatic profiling is already running")
	}

	p.running = true
	p.stopChan = make(chan struct{})

	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_, err := p.CaptureHeapProfile("auto")
				if err != nil {
					fmt.Printf("Failed to capture automatic heap profile: %v\n", err)
				}

				p.UpdateMemoryStats()
				p.CheckMemoryThresholds()
			case <-p.stopChan:
				return
			}
		}
	}()

	return nil
}

// StopAutomaticProfiling stops automatic profiling
func (p *MemoryProfiler) StopAutomaticProfiling() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return
	}

	close(p.stopChan)
	p.running = false
}

// UpdateMemoryStats updates memory statistics
func (p *MemoryProfiler) UpdateMemoryStats() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Read memory stats
	runtime.ReadMemStats(&p.memStats)
}

// GetMemoryStats returns memory statistics
func (p *MemoryProfiler) GetMemoryStats() runtime.MemStats {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.memStats
}

// GetFormattedMemoryStats returns formatted memory statistics
func (p *MemoryProfiler) GetFormattedMemoryStats() map[string]interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Read memory stats if not updated recently
	if time.Since(p.lastCapture) > time.Minute {
		runtime.ReadMemStats(&p.memStats)
	}

	// Convert to MB for better readability
	const MB = 1024 * 1024

	return map[string]interface{}{
		"alloc_mb":        float64(p.memStats.Alloc) / MB,
		"total_alloc_mb":  float64(p.memStats.TotalAlloc) / MB,
		"sys_mb":          float64(p.memStats.Sys) / MB,
		"heap_alloc_mb":   float64(p.memStats.HeapAlloc) / MB,
		"heap_sys_mb":     float64(p.memStats.HeapSys) / MB,
		"heap_idle_mb":    float64(p.memStats.HeapIdle) / MB,
		"heap_inuse_mb":   float64(p.memStats.HeapInuse) / MB,
		"heap_released_mb": float64(p.memStats.HeapReleased) / MB,
		"heap_objects":    p.memStats.HeapObjects,
		"num_gc":          p.memStats.NumGC,
		"next_gc_mb":      float64(p.memStats.NextGC) / MB,
		"gc_cpu_fraction": p.memStats.GCCPUFraction,
		"num_goroutines":  runtime.NumGoroutine(),
		"num_cpu":         runtime.NumCPU(),
		"max_procs":       runtime.GOMAXPROCS(0),
	}
}

// CheckMemoryThresholds checks memory thresholds and generates alerts
func (p *MemoryProfiler) CheckMemoryThresholds() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check memory usage threshold
	heapAllocMB := float64(p.memStats.HeapAlloc) / (1024 * 1024)
	if int64(heapAllocMB) > p.memoryThreshold {
		alert := fmt.Sprintf("Memory usage alert: HeapAlloc %.2f MB exceeds threshold of %d MB", 
			heapAllocMB, p.memoryThreshold)
		p.alerts = append(p.alerts, alert)
	}

	// Check GC pause time threshold
	var gcPauseMS float64
	if p.memStats.NumGC > 0 {
		gcPauseMS = float64(p.memStats.PauseNs[(p.memStats.NumGC+255)%256]) / float64(time.Millisecond)
		if int64(gcPauseMS) > p.gcThreshold {
			alert := fmt.Sprintf("GC pause time alert: %.2f ms exceeds threshold of %d ms", 
				gcPauseMS, p.gcThreshold)
			p.alerts = append(p.alerts, alert)
		}
	}

	// Limit alerts to last 100
	if len(p.alerts) > 100 {
		p.alerts = p.alerts[len(p.alerts)-100:]
	}
}

// GetAlerts returns memory usage alerts
func (p *MemoryProfiler) GetAlerts() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return append([]string{}, p.alerts...)
}

// ClearAlerts clears memory usage alerts
func (p *MemoryProfiler) ClearAlerts() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.alerts = nil
}

// CreateSnapshot creates a memory snapshot
func (p *MemoryProfiler) CreateSnapshot(name string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Read memory stats
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	// Create snapshot
	p.snapshots[name] = &stats
}

// CompareSnapshots compares two memory snapshots
func (p *MemoryProfiler) CompareSnapshots(name1, name2 string) (map[string]interface{}, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if snapshots exist
	snapshot1, exists1 := p.snapshots[name1]
	if !exists1 {
		return nil, fmt.Errorf("snapshot '%s' does not exist", name1)
	}

	snapshot2, exists2 := p.snapshots[name2]
	if !exists2 {
		return nil, fmt.Errorf("snapshot '%s' does not exist", name2)
	}

	// Convert to MB for better readability
	const MB = 1024 * 1024

	// Calculate differences
	return map[string]interface{}{
		"alloc_diff_mb":        float64(snapshot2.Alloc-snapshot1.Alloc) / MB,
		"total_alloc_diff_mb":  float64(snapshot2.TotalAlloc-snapshot1.TotalAlloc) / MB,
		"sys_diff_mb":          float64(snapshot2.Sys-snapshot1.Sys) / MB,
		"heap_alloc_diff_mb":   float64(snapshot2.HeapAlloc-snapshot1.HeapAlloc) / MB,
		"heap_sys_diff_mb":     float64(snapshot2.HeapSys-snapshot1.HeapSys) / MB,
		"heap_idle_diff_mb":    float64(snapshot2.HeapIdle-snapshot1.HeapIdle) / MB,
		"heap_inuse_diff_mb":   float64(snapshot2.HeapInuse-snapshot1.HeapInuse) / MB,
		"heap_released_diff_mb": float64(snapshot2.HeapReleased-snapshot1.HeapReleased) / MB,
		"heap_objects_diff":    snapshot2.HeapObjects - snapshot1.HeapObjects,
		"num_gc_diff":          snapshot2.NumGC - snapshot1.NumGC,
		"gc_cpu_fraction_diff": snapshot2.GCCPUFraction - snapshot1.GCCPUFraction,
	}, nil
}

// ForceGC forces garbage collection
func (p *MemoryProfiler) ForceGC() {
	runtime.GC()
}

// SetMemoryThreshold sets the memory usage threshold
func (p *MemoryProfiler) SetMemoryThreshold(threshold int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.memoryThreshold = threshold
}

// SetGCThreshold sets the GC pause time threshold
func (p *MemoryProfiler) SetGCThreshold(threshold int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.gcThreshold = threshold
}

// GetProfileCount returns the number of profiles captured
func (p *MemoryProfiler) GetProfileCount() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.profileCount
}

// GetLastCaptureTime returns the time of the last profile capture
func (p *MemoryProfiler) GetLastCaptureTime() time.Time {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.lastCapture
}

// IsRunning returns if automatic profiling is running
func (p *MemoryProfiler) IsRunning() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.running
}

// GetOutputDir returns the output directory
func (p *MemoryProfiler) GetOutputDir() string {
	return p.outputDir
}

// GetInterval returns the profiling interval
func (p *MemoryProfiler) GetInterval() time.Duration {
	return p.interval
}

// SetInterval sets the profiling interval
func (p *MemoryProfiler) SetInterval(interval time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.interval = interval
}

// GetMemoryUsage returns the current memory usage in MB
func (p *MemoryProfiler) GetMemoryUsage() float64 {
	p.UpdateMemoryStats()
	return float64(p.memStats.HeapAlloc) / (1024 * 1024)
}

// GetGCPauseTime returns the last GC pause time in ms
func (p *MemoryProfiler) GetGCPauseTime() float64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.memStats.NumGC > 0 {
		return float64(p.memStats.PauseNs[(p.memStats.NumGC+255)%256]) / float64(time.Millisecond)
	}
	return 0
}

// GetGCStats returns GC statistics
func (p *MemoryProfiler) GetGCStats() map[string]interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Read memory stats if not updated recently
	if time.Since(p.lastCapture) > time.Minute {
		runtime.ReadMemStats(&p.memStats)
	}

	// Calculate average pause time
	var totalPause uint64
	n := uint32(0)
	for i := uint32(0); i < p.memStats.NumGC && i < 256; i++ {
		totalPause += p.memStats.PauseNs[(p.memStats.NumGC+255-i)%256]
		n++
	}

	avgPauseMS := 0.0
	if n > 0 {
		avgPauseMS = float64(totalPause) / float64(n) / float64(time.Millisecond)
	}

	lastPauseMS := 0.0
	if p.memStats.NumGC > 0 {
		lastPauseMS = float64(p.memStats.PauseNs[(p.memStats.NumGC+255)%256]) / float64(time.Millisecond)
	}

	return map[string]interface{}{
		"num_gc":           p.memStats.NumGC,
		"gc_cpu_fraction":  p.memStats.GCCPUFraction,
		"avg_pause_ms":     avgPauseMS,
		"last_pause_ms":    lastPauseMS,
		"next_gc_mb":       float64(p.memStats.NextGC) / (1024 * 1024),
		"gc_cycles":        p.memStats.NumGC,
		"forced_gc_cycles": p.memStats.NumForcedGC,
	}
}
