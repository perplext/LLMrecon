// Command profiler provides a command-line tool for profiling the rate limiting system
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"

	"github.com/perplext/LLMrecon/src/template/management/ratelimit"
)

var (
	// Command line flags
	duration        = flag.Duration("duration", 30*time.Second, "Duration of the profiling run")
	concurrentUsers = flag.Int("users", 10, "Number of concurrent users")
	requestsPerSec  = flag.Int("rps", 100, "Requests per second per user")
	globalQPS       = flag.Float64("global-qps", 1000, "Global queries per second limit")
	globalBurst     = flag.Int("global-burst", 100, "Global burst limit")
	userQPS         = flag.Float64("user-qps", 100, "Default user queries per second limit")
	userBurst       = flag.Int("user-burst", 10, "Default user burst limit")
	outputDir       = flag.String("output", "profiles", "Output directory for profiling data")
	userPriorities  = flag.String("priorities", "", "Comma-separated list of user:priority pairs (e.g., user1:10,user2:5)")
	cpuProfile      = flag.Bool("cpu", true, "Enable CPU profiling")
	memProfile      = flag.Bool("mem", true, "Enable memory profiling")
	traceProfile    = flag.Bool("trace", false, "Enable execution tracing")
	loadFactor      = flag.Float64("load", 1.0, "System load factor (1.0 = normal, <1.0 = high load)")
	fairness        = flag.Bool("fairness", true, "Enable fairness mechanisms")
	dynamicAdj      = flag.Bool("dynamic", true, "Enable dynamic adjustment")
	verbose         = flag.Bool("verbose", false, "Enable verbose output")
)

func main() {
	flag.Parse()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		fmt.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	// Create the limiter
	limiter := ratelimit.NewAdaptiveLimiter(*globalQPS, *globalBurst, *userQPS, *userBurst)
	limiter.EnableFairness(*fairness)
	limiter.EnableDynamicAdjustment(*dynamicAdj)
	limiter.SetLoadFactor(*loadFactor)

	// Parse user priorities
	userPriorityMap := parseUserPriorities(*userPriorities)

	// Set up user policies
	for userID, priority := range userPriorityMap {
		limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
			UserID:        userID,
			QPS:           *userQPS,
			Burst:         *userBurst,
			Priority:      priority,
			MaxTokens:     int(*userQPS) * 60, // 1 minute worth of tokens
			ResetInterval: time.Minute,
		})
	}

	// Create the profiler
	profiler := ratelimit.NewProfiler(limiter, *outputDir)
	profiler.SetCPUProfiling(*cpuProfile)
	profiler.SetMemoryProfiling(*memProfile)
	profiler.SetTraceProfiling(*traceProfile)

	// Start CPU profiling if enabled
	if *cpuProfile {
		cpuFile, err := os.Create(filepath.Join(*outputDir, "cpu_main.pprof"))
		if err != nil {
			log.Fatalf("Could not create CPU profile: %v", err)
		}
		defer cpuFile.Close()
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			log.Fatalf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Print test configuration
	fmt.Println("Rate Limiting System Profiler")
	fmt.Println("============================")
	fmt.Printf("Duration: %v\n", *duration)
	fmt.Printf("Concurrent Users: %d\n", *concurrentUsers)
	fmt.Printf("Requests Per Second Per User: %d\n", *requestsPerSec)
	fmt.Printf("Global QPS: %.2f\n", *globalQPS)
	fmt.Printf("User QPS: %.2f\n", *userQPS)
	fmt.Printf("System Load Factor: %.2f\n", *loadFactor)
	fmt.Printf("Fairness Enabled: %t\n", *fairness)
	fmt.Printf("Dynamic Adjustment Enabled: %t\n", *dynamicAdj)
	fmt.Printf("Output Directory: %s\n", *outputDir)
	fmt.Println()

	// Run the load test
	fmt.Println("Running load test...")
	startTime := time.Now()

	result, err := profiler.RunLoadTest(
		ctx,
		*duration,
		*concurrentUsers,
		*requestsPerSec,
		userPriorityMap,
	)

	if err != nil {
		log.Fatalf("Failed to run load test: %v", err)
	}

	// Generate and print the report
	report := profiler.GenerateReport(result)
	fmt.Println(report)

	// Save the report to a file
	reportFilename := fmt.Sprintf("profile_report_%s.txt", time.Now().Format("20060102_150405"))
	if err := profiler.SaveReportToFile(report, reportFilename); err != nil {
		log.Printf("Failed to save report: %v", err)
	} else {
		fmt.Printf("Report saved to %s\n", filepath.Join(*outputDir, reportFilename))
	}

	// Print detailed statistics if verbose mode is enabled
	if *verbose {
		printDetailedStats(result)
	}

	// Print summary table
	printSummaryTable(result)

	fmt.Printf("\nProfiling completed in %v\n", time.Since(startTime))
}

// parseUserPriorities parses a comma-separated list of user:priority pairs
func parseUserPriorities(prioritiesStr string) map[string]int {
	result := make(map[string]int)

	if prioritiesStr == "" {
		// Create default users if none specified
		for i := 1; i <= 10; i++ {
			userID := fmt.Sprintf("user-%d", i)
			priority := (i-1)%10 + 1 // Priorities 1-10
			result[userID] = priority
		}
		return result
	}

	pairs := strings.Split(prioritiesStr, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			log.Printf("Invalid user:priority pair: %s", pair)
			continue
		}

		userID := strings.TrimSpace(parts[0])
		priorityStr := strings.TrimSpace(parts[1])

		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			log.Printf("Invalid priority for user %s: %s", userID, priorityStr)
			continue
		}

		result[userID] = priority
	}

	return result
}

// printDetailedStats prints detailed statistics about the profiling run
func printDetailedStats(result *ratelimit.ProfileResult) {
	fmt.Println("\nDetailed Statistics")
	fmt.Println("===================")

	fmt.Println("\nMemory Statistics:")
	fmt.Printf("  Alloc: %.2f MB\n", float64(result.MemoryUsage.Alloc)/(1024*1024))
	fmt.Printf("  TotalAlloc: %.2f MB\n", float64(result.MemoryUsage.TotalAlloc)/(1024*1024))
	fmt.Printf("  Sys: %.2f MB\n", float64(result.MemoryUsage.Sys)/(1024*1024))
	fmt.Printf("  NumGC: %d\n", result.MemoryUsage.NumGC)

	if limiterStats, ok := result.ProfileData["limiter_stats"].(map[string]interface{}); ok {
		fmt.Println("\nLimiter Statistics:")
		for key, value := range limiterStats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}

// printSummaryTable prints a summary table of the profiling results
func printSummaryTable(result *ratelimit.ProfileResult) {
	fmt.Println("\nSummary by Priority Level")
	fmt.Println("========================")

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

	// Print header
	fmt.Printf("%-10s %-10s %-10s %-15s %-15s\n", "Priority", "Requests", "Success %", "Avg Response", "Max Response")
	fmt.Println(strings.Repeat("-", 65))

	// Print data rows
	for _, priority := range priorities {
		stats := result.StatsByPriority[priority]
		successRate := float64(stats.Acquisitions) / float64(stats.Requests) * 100

		fmt.Printf("%-10d %-10d %-10.2f%% %-15v %-15v\n",
			priority,
			stats.Requests,
			successRate,
			stats.AverageResponseTime,
			stats.MaxResponseTime)
	}
}
