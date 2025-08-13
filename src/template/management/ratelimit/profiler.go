// Package ratelimit provides rate limiting functionality for template execution
package ratelimit

import (
	"context"
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
	"os"
	"path/filepath"
	"log"
)

// ProfileResult contains the results of a profiling run
type ProfileResult struct {
	// Duration of the profiling run
	Duration time.Duration
	
	// Total number of requests processed
	TotalRequests int64
	
	// Total number of successful acquisitions
	SuccessfulAcquisitions int64
	
	// Total number of rejections
	Rejections int64
	
	// Average response time (time to acquire or be rejected)
	AverageResponseTime time.Duration
	
	// Maximum response time observed
	MaxResponseTime time.Duration
	
	// Throughput (requests per second)
	Throughput float64
	
	// CPU utilization during the test
	CPUUtilization float64
	
	// Memory usage statistics
	MemoryUsage runtime.MemStats
	
	// Detailed statistics by user priority
	StatsByPriority map[int]*PriorityStats
	
	// Raw profiling data for further analysis
	ProfileData map[string]interface{}
}

// PriorityStats contains statistics for a specific priority level
type PriorityStats struct {
	// Total requests for this priority
	Requests int64
	
	// Successful acquisitions for this priority
	Acquisitions int64
	
	// Rejections for this priority
	Rejections int64
	
	// Average response time for this priority
	AverageResponseTime time.Duration
	
	// Maximum response time for this priority
	MaxResponseTime time.Duration
}

// Profiler provides tools for profiling the rate limiting system
type Profiler struct {
	// Target limiter to profile
	limiter *AdaptiveLimiter
	
	// Output directory for profiling data
	outputDir string
	
	// Whether to enable CPU profiling
	cpuProfiling bool
	
	// Whether to enable memory profiling
	memoryProfiling bool
	
	// Whether to enable trace profiling
	traceProfiling bool
}

// NewProfiler creates a new profiler for the given limiter
func NewProfiler(limiter *AdaptiveLimiter, outputDir string) *Profiler {
	// Create output directory if it doesn't exist
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Printf("Failed to create output directory: %v", err)
		}
	}
	
	return &Profiler{
		limiter:         limiter,
		outputDir:       outputDir,
		cpuProfiling:    true,
		memoryProfiling: true,
		traceProfiling:  false,
	}
}

// SetCPUProfiling enables or disables CPU profiling
func (p *Profiler) SetCPUProfiling(enabled bool) {
	p.cpuProfiling = enabled
}

// SetMemoryProfiling enables or disables memory profiling
func (p *Profiler) SetMemoryProfiling(enabled bool) {
	p.memoryProfiling = enabled
}

// SetTraceProfiling enables or disables trace profiling
func (p *Profiler) SetTraceProfiling(enabled bool) {
	p.traceProfiling = enabled
}

// RunLoadTest runs a load test with the specified parameters
func (p *Profiler) RunLoadTest(
	ctx context.Context,
	duration time.Duration,
	concurrentUsers int,
	requestsPerSecond int,
	userPriorities map[string]int,
) (*ProfileResult, error) {
	if p.limiter == nil {
		return nil, fmt.Errorf("no limiter specified for profiling")
	}
	
	// Start profiling if enabled
	if p.cpuProfiling {
		cpuFile, err := os.Create(filepath.Join(p.outputDir, "cpu.pprof"))
		if err != nil {
			return nil, fmt.Errorf("could not create CPU profile: %v", err)
		}
		defer cpuFile.Close()
		
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			return nil, fmt.Errorf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	
	// Create a context with timeout for the test duration
	testCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	
	// Create channels for collecting results
	resultCh := make(chan struct {
		userID       string
		priority     int
		success      bool
		responseTime time.Duration
	}, concurrentUsers*requestsPerSecond)
	
	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup
	
	// Create counters for tracking progress
	var totalRequests int64
	var successfulAcquisitions int64
	var rejections int64
	
	// Create a map for tracking statistics by priority
	statsByPriority := make(map[int]*PriorityStats)
	
	// Start time for calculating throughput
	startTime := time.Now()
	
	// Start workers for each user
	userIDs := make([]string, 0, len(userPriorities))
	for userID := range userPriorities {
		userIDs = append(userIDs, userID)
	}
	
	// If no user priorities specified, create some default users
	if len(userIDs) == 0 {
		for i := 0; i < concurrentUsers; i++ {
			userID := fmt.Sprintf("user-%d", i)
			priority := (i % 10) + 1 // Priorities 1-10
			userPriorities[userID] = priority
			userIDs = append(userIDs, userID)
		}
	}
	
	// Ensure we have enough users
	for len(userIDs) < concurrentUsers {
		for i := len(userIDs); i < concurrentUsers; i++ {
			userID := fmt.Sprintf("user-%d", i)
			priority := (i % 10) + 1 // Priorities 1-10
			userPriorities[userID] = priority
			userIDs = append(userIDs, userID)
		}
	}
	
	// Start worker goroutines
	for i := 0; i < concurrentUsers; i++ {
		userID := userIDs[i%len(userIDs)]
		priority := userPriorities[userID]
		
		// Set the user policy if it doesn't exist
		policy := p.limiter.GetUserPolicy(userID)
		if policy == nil {
			p.limiter.SetUserPolicy(&UserRateLimitPolicy{
				UserID:        userID,
				QPS:           float64(requestsPerSecond),
				Burst:         requestsPerSecond * 2,
				Priority:      priority,
				MaxTokens:     requestsPerSecond * 60, // 1 minute worth of tokens
				ResetInterval: time.Minute,
			})
		}
		
		wg.Add(1)
		go func(userID string, priority int) {
			defer wg.Done()
			
			// Calculate delay between requests to achieve target RPS
			delay := time.Second / time.Duration(requestsPerSecond)
			
			ticker := time.NewTicker(delay)
			defer ticker.Stop()
			
			for {
				select {
				case <-testCtx.Done():
					return
				case <-ticker.C:
					// Create a context with a short timeout for each request
					reqCtx, reqCancel := context.WithTimeout(testCtx, 5*time.Second)
					
					// Track request time
					requestStart := time.Now()
					
					// Try to acquire a token
					err := p.limiter.AcquireForUser(reqCtx, userID)
					
					// Calculate response time
					responseTime := time.Since(requestStart)
					
					// Record result
					atomic.AddInt64(&totalRequests, 1)
					
					if err == nil {
						atomic.AddInt64(&successfulAcquisitions, 1)
						resultCh <- struct {
							userID       string
							priority     int
							success      bool
							responseTime time.Duration
						}{
							userID:       userID,
							priority:     priority,
							success:      true,
							responseTime: responseTime,
						}
					} else {
						atomic.AddInt64(&rejections, 1)
						resultCh <- struct {
							userID       string
							priority     int
							success      bool
							responseTime time.Duration
						}{
							userID:       userID,
							priority:     priority,
							success:      false,
							responseTime: responseTime,
						}
					}
					
					reqCancel()
				}
			}
		}(userID, priority)
	}
	
	// Start a goroutine to collect results
	go func() {
		for result := range resultCh {
			// Initialize priority stats if needed
			if _, ok := statsByPriority[result.priority]; !ok {
				statsByPriority[result.priority] = &PriorityStats{}
			}
			
			stats := statsByPriority[result.priority]
			stats.Requests++
			
			if result.success {
				stats.Acquisitions++
			} else {
				stats.Rejections++
			}
			
			// Update average response time
			if stats.Requests == 1 {
				stats.AverageResponseTime = result.responseTime
			} else {
				stats.AverageResponseTime = time.Duration(
					(stats.AverageResponseTime.Nanoseconds()*(stats.Requests-1) + 
					result.responseTime.Nanoseconds()) / stats.Requests,
				)
			}
			
			// Update max response time
			if result.responseTime > stats.MaxResponseTime {
				stats.MaxResponseTime = result.responseTime
			}
		}
	}()
	
	// Wait for the test to complete
	<-testCtx.Done()
	
	// Wait for all workers to finish
	wg.Wait()
	
	// Close the result channel
	close(resultCh)
	
	// Calculate test duration
	testDuration := time.Since(startTime)
	
	// Capture memory profile if enabled
	var memStats runtime.MemStats
	if p.memoryProfiling {
		runtime.ReadMemStats(&memStats)
		
		memFile, err := os.Create(filepath.Join(p.outputDir, "memory.pprof"))
		if err != nil {
			log.Printf("Could not create memory profile: %v", err)
		} else {
			defer memFile.Close()
			if err := pprof.WriteHeapProfile(memFile); err != nil {
				log.Printf("Could not write memory profile: %v", err)
			}
		}
	}
	
	// Calculate overall statistics
	var totalResponseTime time.Duration
	var maxResponseTime time.Duration
	
	for _, stats := range statsByPriority {
		totalResponseTime += stats.AverageResponseTime * time.Duration(stats.Requests)
		if stats.MaxResponseTime > maxResponseTime {
			maxResponseTime = stats.MaxResponseTime
		}
	}
	
	averageResponseTime := time.Duration(0)
	if totalRequests > 0 {
		averageResponseTime = time.Duration(totalResponseTime.Nanoseconds() / totalRequests)
	}
	
	// Calculate throughput
	throughput := float64(totalRequests) / testDuration.Seconds()
	
	// Create the profile result
	result := &ProfileResult{
		Duration:              testDuration,
		TotalRequests:         totalRequests,
		SuccessfulAcquisitions: successfulAcquisitions,
		Rejections:            rejections,
		AverageResponseTime:   averageResponseTime,
		MaxResponseTime:       maxResponseTime,
		Throughput:            throughput,
		CPUUtilization:        0, // Not calculated in this implementation
		MemoryUsage:           memStats,
		StatsByPriority:       statsByPriority,
		ProfileData:           make(map[string]interface{}),
	}
	
	// Add limiter stats
	if p.limiter.statsEnabled {
		result.ProfileData["limiter_stats"] = p.limiter.stats.GetStats()
		result.ProfileData["recent_events"] = p.limiter.stats.GetRecentEvents()
	}
	
	return result, nil
}

// GenerateReport generates a human-readable report from profiling results
func (p *Profiler) GenerateReport(result *ProfileResult) string {
	report := "Rate Limiting System Performance Report\n"
	report += "=====================================\n\n"
	
	report += fmt.Sprintf("Test Duration: %v\n", result.Duration)
	report += fmt.Sprintf("Total Requests: %d\n", result.TotalRequests)
	report += fmt.Sprintf("Successful Acquisitions: %d (%.2f%%)\n", 
		result.SuccessfulAcquisitions, 
		float64(result.SuccessfulAcquisitions)/float64(result.TotalRequests)*100)
	report += fmt.Sprintf("Rejections: %d (%.2f%%)\n", 
		result.Rejections, 
		float64(result.Rejections)/float64(result.TotalRequests)*100)
	report += fmt.Sprintf("Average Response Time: %v\n", result.AverageResponseTime)
	report += fmt.Sprintf("Maximum Response Time: %v\n", result.MaxResponseTime)
	report += fmt.Sprintf("Throughput: %.2f requests/second\n", result.Throughput)
	report += fmt.Sprintf("Memory Usage: %.2f MB\n\n", float64(result.MemoryUsage.Alloc)/(1024*1024))
	
	report += "Statistics by Priority Level\n"
	report += "---------------------------\n"
	
	// Sort priorities for consistent output
	priorities := make([]int, 0, len(result.StatsByPriority))
	for priority := range result.StatsByPriority {
		priorities = append(priorities, priority)
	}
	
	// Sort priorities in descending order (highest priority first)
	for i := 0; i < len(priorities); i++ {
		for j := i + 1; j < len(priorities); j++ {
			if priorities[i] < priorities[j] {
				priorities[i], priorities[j] = priorities[j], priorities[i]
			}
		}
	}
	
	for _, priority := range priorities {
		stats := result.StatsByPriority[priority]
		report += fmt.Sprintf("\nPriority %d:\n", priority)
		report += fmt.Sprintf("  Requests: %d\n", stats.Requests)
		report += fmt.Sprintf("  Acquisitions: %d (%.2f%%)\n", 
			stats.Acquisitions, 
			float64(stats.Acquisitions)/float64(stats.Requests)*100)
		report += fmt.Sprintf("  Rejections: %d (%.2f%%)\n", 
			stats.Rejections, 
			float64(stats.Rejections)/float64(stats.Requests)*100)
		report += fmt.Sprintf("  Average Response Time: %v\n", stats.AverageResponseTime)
		report += fmt.Sprintf("  Maximum Response Time: %v\n", stats.MaxResponseTime)
	}
	
	report += "\nBottleneck Analysis\n"
	report += "------------------\n"
	
	// Analyze potential bottlenecks
	if result.MaxResponseTime > 500*time.Millisecond {
		report += "- High maximum response time detected. This could indicate contention in the priority queue.\n"
	}
	
	if result.Throughput < float64(result.TotalRequests)/result.Duration.Seconds()*0.8 {
		report += "- Throughput is lower than expected. Check for lock contention or inefficient queue processing.\n"
	}
	
	// Check for priority inversion (lower priority users getting better service than higher priority)
	priorityInversion := false
	for i := 0; i < len(priorities)-1; i++ {
		higherPriority := priorities[i]
		lowerPriority := priorities[i+1]
		
		higherStats := result.StatsByPriority[higherPriority]
		lowerStats := result.StatsByPriority[lowerPriority]
		
		// Check if lower priority has better success rate or response time
		if float64(lowerStats.Acquisitions)/float64(lowerStats.Requests) > 
		   float64(higherStats.Acquisitions)/float64(higherStats.Requests) {
			priorityInversion = true
			report += fmt.Sprintf("- Priority inversion detected: Priority %d has better success rate than Priority %d\n", 
				lowerPriority, higherPriority)
		}
		
		if lowerStats.AverageResponseTime < higherStats.AverageResponseTime {
			priorityInversion = true
			report += fmt.Sprintf("- Priority inversion detected: Priority %d has better response time than Priority %d\n", 
				lowerPriority, higherPriority)
		}
	}
	
	if !priorityInversion {
		report += "- No priority inversion detected. Priority-based fairness is working correctly.\n"
	}
	
	// Check for high rejection rate
	if float64(result.Rejections)/float64(result.TotalRequests) > 0.2 {
		report += "- High rejection rate detected. Consider increasing rate limits or improving fairness mechanisms.\n"
	}
	
	return report
}

// SaveReportToFile saves the profiling report to a file
func (p *Profiler) SaveReportToFile(report string, filename string) error {
	path := filepath.Join(p.outputDir, filename)
	return os.WriteFile(path, []byte(report), 0644)
}
