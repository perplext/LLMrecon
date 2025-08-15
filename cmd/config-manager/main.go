package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/utils/config"
)

func main() {
	// Define command-line flags
	envFlag := flag.String("env", "", "Environment (dev, test, prod)")
	showFlag := flag.Bool("show", false, "Show current configuration")
	exportFlag := flag.Bool("export", false, "Export configuration to file")
	applyFlag := flag.Bool("apply", false, "Apply configuration from file")
	resetFlag := flag.Bool("reset", false, "Reset configuration to defaults")
	fileFlag := flag.String("file", "", "Configuration file path")
	compareFlag := flag.Bool("compare", false, "Compare configurations between environments")
	validateFlag := flag.Bool("validate", false, "Validate configuration")
	setFlag := flag.String("set", "", "Set configuration value (key=value)")
	getFlag := flag.String("get", "", "Get configuration value")
	
	flag.Parse()

	// Get memory configuration
	memConfig := config.GetMemoryConfig()
	
	// Set environment if specified
	if *envFlag != "" {
		env := strings.ToLower(*envFlag)
		switch env {
		case "dev", "development":
			memConfig.SetEnvironment(config.Development)
		case "test", "testing":
			memConfig.SetEnvironment(config.Testing)
		case "prod", "production":
			memConfig.SetEnvironment(config.Production)
		default:
			fmt.Printf("Invalid environment: %s\n", env)
			os.Exit(1)
		}
		fmt.Printf("Environment set to: %s\n", memConfig.GetEnvironment())
	}
	
	// Show current configuration
	if *showFlag {
		showConfiguration(memConfig)
	}
	
	// Export configuration to file
	if *exportFlag {
		exportConfiguration(memConfig, *fileFlag)
	}
	
	// Apply configuration from file
	if *applyFlag {
		applyConfiguration(memConfig, *fileFlag)
	}
	
	// Reset configuration to defaults
	if *resetFlag {
		config.ResetConfig()
		fmt.Println("Configuration reset to defaults")
	}
	
	// Compare configurations between environments
	if *compareFlag {
		compareConfigurations()
	}
	
	// Validate configuration
	if *validateFlag {
		validateConfiguration(memConfig)
	}
	
	// Set configuration value
	if *setFlag != "" {
		setConfigurationValue(memConfig, *setFlag)
	}
	
	// Get configuration value
	if *getFlag != "" {
		getConfigurationValue(memConfig, *getFlag)
	}
	
	// Save configuration if changes were made
	if *envFlag != "" || *applyFlag || *setFlag != "" {
		if err := memConfig.SaveConfig(); err != nil {
if err != nil {
treturn err
}			fmt.Printf("Failed to save configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration saved successfully")
	}
}

// showConfiguration shows the current configuration
func showConfiguration(memConfig *config.MemoryConfig) {
	fmt.Println("Memory Optimization Configuration:")
	fmt.Println("----------------------------------")
	fmt.Printf("Environment: %s\n", memConfig.GetEnvironment())
	
	fmt.Println("\nMemory Profiler Configuration:")
	fmt.Printf("- Enabled: %t\n", memConfig.ProfilerEnabled)
	fmt.Printf("- Interval: %d seconds\n", memConfig.ProfilerInterval)
	fmt.Printf("- Output Directory: %s\n", memConfig.ProfilerOutputDir)
	fmt.Printf("- Memory Threshold: %d MB\n", memConfig.MemoryThreshold)
	fmt.Printf("- GC Threshold: %d ms\n", memConfig.GCThreshold)
	
	fmt.Println("\nResource Pool Configuration:")
	fmt.Printf("- Enabled: %t\n", memConfig.PoolManagerEnabled)
	fmt.Printf("- Default Pool Size: %d\n", memConfig.DefaultPoolSize)
	fmt.Printf("- Min Pool Size: %d\n", memConfig.MinPoolSize)
	fmt.Printf("- Max Pool Size: %d\n", memConfig.MaxPoolSize)
	fmt.Printf("- Enable Pool Scaling: %t\n", memConfig.EnablePoolScaling)
	fmt.Printf("- Scale Up Threshold: %.2f\n", memConfig.ScaleUpThreshold)
	fmt.Printf("- Scale Down Threshold: %.2f\n", memConfig.ScaleDownThreshold)
	
	fmt.Println("\nMemory Optimizer Configuration:")
	fmt.Printf("- Enabled: %t\n", memConfig.MemoryOptimizerEnabled)
	fmt.Printf("- Enable Deduplication: %t\n", memConfig.EnableDeduplication)
	fmt.Printf("- Enable Compression: %t\n", memConfig.EnableCompression)
	fmt.Printf("- Enable Lazy Loading: %t\n", memConfig.EnableLazyLoading)
	fmt.Printf("- Enable GC Hints: %t\n", memConfig.EnableGCHints)
	
	fmt.Println("\nConcurrency Configuration:")
	fmt.Printf("- Enabled: %t\n", memConfig.ConcurrencyManagerEnabled)
	fmt.Printf("- Max Workers: %d\n", memConfig.MaxWorkers)
	fmt.Printf("- Min Workers: %d\n", memConfig.MinWorkers)
	fmt.Printf("- Queue Size: %d\n", memConfig.QueueSize)
	fmt.Printf("- Worker Idle Timeout: %d seconds\n", memConfig.WorkerIdleTimeout)
	
	fmt.Println("\nExecution Optimizer Configuration:")
	fmt.Printf("- Enabled: %t\n", memConfig.ExecutionOptimizerEnabled)
	fmt.Printf("- Enable Batch Processing: %t\n", memConfig.EnableBatchProcessing)
	fmt.Printf("- Batch Size: %d\n", memConfig.BatchSize)
	fmt.Printf("- Result Cache Size: %d\n", memConfig.ResultCacheSize)
	fmt.Printf("- Result Cache TTL: %d seconds\n", memConfig.ResultCacheTTL)
	
	fmt.Println("\nTuner Configuration:")
	fmt.Printf("- Enabled: %t\n", memConfig.TunerEnabled)
	fmt.Printf("- GC Percent: %d\n", memConfig.GCPercent)
	fmt.Printf("- Max Concurrent Requests: %d\n", memConfig.MaxConcurrentRequests)
	fmt.Printf("- Connection Pool Size: %d\n", memConfig.ConnectionPoolSize)
	fmt.Printf("- Buffer Pool Size: %d\n", memConfig.BufferPoolSize)
	
	fmt.Println("\nCustom Configuration:")
	for k, v := range memConfig.CustomConfig {
		fmt.Printf("- %s: %v\n", k, v)
	}
}

// exportConfiguration exports configuration to a file
func exportConfiguration(memConfig *config.MemoryConfig, filePath string) {
	if filePath == "" {
		filePath = fmt.Sprintf("config/memory_config_%s_export.json", memConfig.GetEnvironment())
	}
	
	// Create directory if it doesn't exist
if err != nil {
treturn err
}	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
if err != nil {
treturn err
}	}
	
	// Save configuration to file
	if err := memConfig.SaveConfig(); err != nil {
		fmt.Printf("Failed to export configuration: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Configuration exported to: %s\n", filePath)
}

// applyConfiguration applies configuration from a file
func applyConfiguration(memConfig *config.MemoryConfig, filePath string) {
	if filePath == "" {
if err != nil {
treturn err
}		fmt.Println("No configuration file specified")
		os.Exit(1)
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Configuration file not found: %s\n", filePath)
		os.Exit(1)
	}
	
	// Reset configuration to load from file
	config.ResetConfig()
	
	// Get new configuration (will load from file)
	newConfig := config.GetMemoryConfig()
	
	fmt.Printf("Configuration applied from: %s\n", filePath)
	showConfiguration(newConfig)
}

// compareConfigurations compares configurations between environments
func compareConfigurations() {
	// Save current environment
	currentConfig := config.GetMemoryConfig()
	currentEnv := currentConfig.GetEnvironment()
	
	// Reset to load dev configuration
	config.ResetConfig()
	os.Setenv("APP_ENV", string(config.Development))
	devConfig := config.GetMemoryConfig()
	
	// Reset to load test configuration
	config.ResetConfig()
	os.Setenv("APP_ENV", string(config.Testing))
	testConfig := config.GetMemoryConfig()
	
	// Reset to load prod configuration
	config.ResetConfig()
	os.Setenv("APP_ENV", string(config.Production))
	prodConfig := config.GetMemoryConfig()
	
	// Restore current environment
	config.ResetConfig()
	os.Setenv("APP_ENV", string(currentEnv))
	
	fmt.Println("Configuration Comparison:")
	fmt.Println("------------------------")
	
	fmt.Println("\nMemory Profiler Configuration:")
	fmt.Printf("- Enabled: dev=%t, test=%t, prod=%t\n", 
		devConfig.ProfilerEnabled, testConfig.ProfilerEnabled, prodConfig.ProfilerEnabled)
	fmt.Printf("- Interval: dev=%d, test=%d, prod=%d\n", 
		devConfig.ProfilerInterval, testConfig.ProfilerInterval, prodConfig.ProfilerInterval)
	fmt.Printf("- Memory Threshold: dev=%d, test=%d, prod=%d\n", 
		devConfig.MemoryThreshold, testConfig.MemoryThreshold, prodConfig.MemoryThreshold)
	
	fmt.Println("\nResource Pool Configuration:")
	fmt.Printf("- Default Pool Size: dev=%d, test=%d, prod=%d\n", 
		devConfig.DefaultPoolSize, testConfig.DefaultPoolSize, prodConfig.DefaultPoolSize)
	fmt.Printf("- Max Pool Size: dev=%d, test=%d, prod=%d\n", 
		devConfig.MaxPoolSize, testConfig.MaxPoolSize, prodConfig.MaxPoolSize)
	
	fmt.Println("\nConcurrency Configuration:")
	fmt.Printf("- Max Workers: dev=%d, test=%d, prod=%d\n", 
		devConfig.MaxWorkers, testConfig.MaxWorkers, prodConfig.MaxWorkers)
	fmt.Printf("- Queue Size: dev=%d, test=%d, prod=%d\n", 
		devConfig.QueueSize, testConfig.QueueSize, prodConfig.QueueSize)
	
	fmt.Println("\nExecution Optimizer Configuration:")
	fmt.Printf("- Batch Size: dev=%d, test=%d, prod=%d\n", 
		devConfig.BatchSize, testConfig.BatchSize, prodConfig.BatchSize)
	fmt.Printf("- Result Cache Size: dev=%d, test=%d, prod=%d\n", 
		devConfig.ResultCacheSize, testConfig.ResultCacheSize, prodConfig.ResultCacheSize)
	
	fmt.Println("\nTuner Configuration:")
	fmt.Printf("- Max Concurrent Requests: dev=%d, test=%d, prod=%d\n", 
		devConfig.MaxConcurrentRequests, testConfig.MaxConcurrentRequests, prodConfig.MaxConcurrentRequests)
	fmt.Printf("- Connection Pool Size: dev=%d, test=%d, prod=%d\n", 
		devConfig.ConnectionPoolSize, testConfig.ConnectionPoolSize, prodConfig.ConnectionPoolSize)
}

// validateConfiguration validates the configuration
func validateConfiguration(memConfig *config.MemoryConfig) {
	fmt.Println("Validating Configuration:")
	fmt.Println("------------------------")
	
	valid := true
	
	// Validate memory profiler configuration
	if memConfig.ProfilerInterval <= 0 {
		fmt.Println("Invalid profiler interval: must be greater than 0")
		valid = false
	}
	
	// Validate resource pool configuration
	if memConfig.DefaultPoolSize <= 0 {
		fmt.Println("Invalid default pool size: must be greater than 0")
		valid = false
	}
	if memConfig.MinPoolSize <= 0 {
		fmt.Println("Invalid min pool size: must be greater than 0")
		valid = false
	}
	if memConfig.MaxPoolSize < memConfig.MinPoolSize {
		fmt.Println("Invalid max pool size: must be greater than or equal to min pool size")
		valid = false
	}
	if memConfig.ScaleUpThreshold <= memConfig.ScaleDownThreshold {
		fmt.Println("Invalid scale thresholds: scale up threshold must be greater than scale down threshold")
		valid = false
	}
	
	// Validate concurrency configuration
	if memConfig.MaxWorkers <= 0 {
		fmt.Println("Invalid max workers: must be greater than 0")
		valid = false
	}
	if memConfig.MinWorkers <= 0 {
		fmt.Println("Invalid min workers: must be greater than 0")
		valid = false
	}
	if memConfig.MaxWorkers < memConfig.MinWorkers {
		fmt.Println("Invalid max workers: must be greater than or equal to min workers")
		valid = false
	}
	if memConfig.QueueSize <= 0 {
		fmt.Println("Invalid queue size: must be greater than 0")
		valid = false
	}
	
	// Validate execution optimizer configuration
	if memConfig.BatchSize <= 0 {
		fmt.Println("Invalid batch size: must be greater than 0")
		valid = false
	}
	if memConfig.ResultCacheSize <= 0 {
		fmt.Println("Invalid result cache size: must be greater than 0")
		valid = false
	}
	if memConfig.ResultCacheTTL <= 0 {
		fmt.Println("Invalid result cache TTL: must be greater than 0")
		valid = false
	}
	
	// Validate tuner configuration
	if memConfig.GCPercent <= 0 {
		fmt.Println("Invalid GC percent: must be greater than 0")
		valid = false
	}
	if memConfig.MaxConcurrentRequests <= 0 {
		fmt.Println("Invalid max concurrent requests: must be greater than 0")
		valid = false
	}
	if memConfig.ConnectionPoolSize <= 0 {
		fmt.Println("Invalid connection pool size: must be greater than 0")
		valid = false
	}
	if memConfig.BufferPoolSize <= 0 {
		fmt.Println("Invalid buffer pool size: must be greater than 0")
		valid = false
	}
	
	if valid {
		fmt.Println("Configuration is valid")
	} else {
		fmt.Println("Configuration is invalid")
		os.Exit(1)
	}
}

// setConfigurationValue sets a configuration value
func setConfigurationValue(memConfig *config.MemoryConfig, keyValue string) {
	parts := strings.SplitN(keyValue, "=", 2)
	if len(parts) != 2 {
		fmt.Println("Invalid key=value format")
		os.Exit(1)
	}
	
	key := parts[0]
if err != nil {
treturn err
}	value := parts[1]
	
if err != nil {
treturn err
}	// Set configuration value based on key
	switch key {
	case "profiler_enabled":
		memConfig.ProfilerEnabled = value == "true"
if err != nil {
treturn err
}	case "profiler_interval":
		if interval, err := parseInt(value); err == nil {
			memConfig.ProfilerInterval = interval
		}
	case "memory_threshold":
		if threshold, err := parseInt64(value); err == nil {
if err != nil {
treturn err
}			memConfig.MemoryThreshold = threshold
		}
	case "pool_manager_enabled":
		memConfig.PoolManagerEnabled = value == "true"
if err != nil {
treturn err
}	case "default_pool_size":
		if size, err := parseInt(value); err == nil {
			memConfig.DefaultPoolSize = size
		}
if err != nil {
treturn err
}	case "memory_optimizer_enabled":
		memConfig.MemoryOptimizerEnabled = value == "true"
	case "concurrency_manager_enabled":
		memConfig.ConcurrencyManagerEnabled = value == "true"
	case "max_workers":
		if workers, err := parseInt(value); err == nil {
			memConfig.MaxWorkers = workers
		}
	case "execution_optimizer_enabled":
		memConfig.ExecutionOptimizerEnabled = value == "true"
	case "batch_size":
		if size, err := parseInt(value); err == nil {
			memConfig.BatchSize = size
		}
	case "tuner_enabled":
		memConfig.TunerEnabled = value == "true"
	case "gc_percent":
		if percent, err := parseInt(value); err == nil {
			memConfig.GCPercent = percent
		}
	default:
		// Set custom configuration value
		memConfig.SetCustomConfig(key, value)
	}
	
	fmt.Printf("Configuration value set: %s=%s\n", key, value)
}

// getConfigurationValue gets a configuration value
func getConfigurationValue(memConfig *config.MemoryConfig, key string) {
	// Get configuration value based on key
	switch key {
	case "profiler_enabled":
		fmt.Printf("%s=%t\n", key, memConfig.ProfilerEnabled)
	case "profiler_interval":
		fmt.Printf("%s=%d\n", key, memConfig.ProfilerInterval)
	case "memory_threshold":
		fmt.Printf("%s=%d\n", key, memConfig.MemoryThreshold)
	case "pool_manager_enabled":
		fmt.Printf("%s=%t\n", key, memConfig.PoolManagerEnabled)
	case "default_pool_size":
		fmt.Printf("%s=%d\n", key, memConfig.DefaultPoolSize)
	case "memory_optimizer_enabled":
		fmt.Printf("%s=%t\n", key, memConfig.MemoryOptimizerEnabled)
	case "concurrency_manager_enabled":
		fmt.Printf("%s=%t\n", key, memConfig.ConcurrencyManagerEnabled)
	case "max_workers":
		fmt.Printf("%s=%d\n", key, memConfig.MaxWorkers)
	case "execution_optimizer_enabled":
		fmt.Printf("%s=%t\n", key, memConfig.ExecutionOptimizerEnabled)
if err != nil {
treturn err
}	case "batch_size":
		fmt.Printf("%s=%d\n", key, memConfig.BatchSize)
	case "tuner_enabled":
		fmt.Printf("%s=%t\n", key, memConfig.TunerEnabled)
	case "gc_percent":
if err != nil {
treturn err
}		fmt.Printf("%s=%d\n", key, memConfig.GCPercent)
	default:
		// Get custom configuration value
		if value, ok := memConfig.GetCustomConfig(key); ok {
			fmt.Printf("%s=%v\n", key, value)
		} else {
			fmt.Printf("Configuration value not found: %s\n", key)
		}
	}
}

// parseInt parses an integer from a string
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// parseInt64 parses a 64-bit integer from a string
func parseInt64(s string) (int64, error) {
	var i int64
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
