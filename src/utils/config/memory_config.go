package config

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// Environment represents the application environment
type Environment string

const (
	// Development environment
	Development Environment = "dev"
	// Testing environment
	Testing Environment = "test"
	// Production environment
	Production Environment = "prod"
)

// MemoryConfig represents configuration for memory optimization
type MemoryConfig struct {
	// Environment is the application environment
	Environment Environment `json:"environment"`
	
	// Memory Profiler Configuration
	ProfilerEnabled       bool   `json:"profiler_enabled"`
	ProfilerInterval      int    `json:"profiler_interval"`       // In seconds
	ProfilerOutputDir     string `json:"profiler_output_dir"`
	MemoryThreshold       int64  `json:"memory_threshold"`        // In MB
	GCThreshold           int64  `json:"gc_threshold"`            // In ms
	
	// Resource Pool Configuration
	PoolManagerEnabled    bool   `json:"pool_manager_enabled"`
	DefaultPoolSize       int    `json:"default_pool_size"`
	MinPoolSize           int    `json:"min_pool_size"`
	MaxPoolSize           int    `json:"max_pool_size"`
	EnablePoolScaling     bool   `json:"enable_pool_scaling"`
	ScaleUpThreshold      float64 `json:"scale_up_threshold"`     // 0.0-1.0
	ScaleDownThreshold    float64 `json:"scale_down_threshold"`   // 0.0-1.0
	
	// Memory Optimizer Configuration
	MemoryOptimizerEnabled bool   `json:"memory_optimizer_enabled"`
	EnableDeduplication    bool   `json:"enable_deduplication"`
	EnableCompression      bool   `json:"enable_compression"`
	EnableLazyLoading      bool   `json:"enable_lazy_loading"`
	EnableGCHints          bool   `json:"enable_gc_hints"`
	
	// Inheritance Optimizer Configuration
	InheritanceOptimizerEnabled bool   `json:"inheritance_optimizer_enabled"`
	MaxInheritanceDepth         int    `json:"max_inheritance_depth"`
	FlattenInheritance          bool   `json:"flatten_inheritance"`
	CacheOptimizedTemplates     bool   `json:"cache_optimized_templates"`
	
	// Context Optimizer Configuration
	ContextOptimizerEnabled     bool   `json:"context_optimizer_enabled"`
	ContextDeduplication        bool   `json:"context_deduplication"`
	ContextLazyLoading          bool   `json:"context_lazy_loading"`
	ContextCompression          bool   `json:"context_compression"`
	
	// Concurrency Configuration
	ConcurrencyManagerEnabled bool   `json:"concurrency_manager_enabled"`
	MaxWorkers                int    `json:"max_workers"`
	MinWorkers                int    `json:"min_workers"`
	QueueSize                 int    `json:"queue_size"`
	WorkerIdleTimeout         int    `json:"worker_idle_timeout"`    // In seconds
	
	// Execution Optimizer Configuration
	ExecutionOptimizerEnabled bool   `json:"execution_optimizer_enabled"`
	EnableBatchProcessing     bool   `json:"enable_batch_processing"`
	BatchSize                 int    `json:"batch_size"`
	ResultCacheSize           int    `json:"result_cache_size"`
	ResultCacheTTL            int    `json:"result_cache_ttl"`       // In seconds
	
	// Tuner Configuration
	TunerEnabled           bool   `json:"tuner_enabled"`
	GCPercent              int    `json:"gc_percent"`
	MaxConcurrentRequests  int    `json:"max_concurrent_requests"`
	ConnectionPoolSize     int    `json:"connection_pool_size"`
	BufferPoolSize         int    `json:"buffer_pool_size"`
	
	// Static File Handler Configuration
	StaticFileHandler      *StaticFileHandlerConfig `json:"static_file_handler"`
	
	// Custom configuration by environment
	CustomConfig           map[string]interface{} `json:"custom_config"`

var (
	// instance is the singleton instance of MemoryConfig
	instance *MemoryConfig
	// mutex protects the instance
	mutex sync.RWMutex
	// configDir is the directory containing configuration files
	configDir = "config"
)

// DefaultMemoryConfig returns default configuration for memory optimization
func DefaultMemoryConfig() *MemoryConfig {
	numCPU := runtime.NumCPU()
	
	return &MemoryConfig{
		Environment:              Development,
		
		ProfilerEnabled:          true,
		ProfilerInterval:         300,
		ProfilerOutputDir:        "profiles",
		MemoryThreshold:          100,
		GCThreshold:              100,
		
		PoolManagerEnabled:       true,
		DefaultPoolSize:          numCPU * 2,
		MinPoolSize:              numCPU,
		MaxPoolSize:              numCPU * 4,
		EnablePoolScaling:        true,
		ScaleUpThreshold:         0.8,
		ScaleDownThreshold:       0.2,
		
		MemoryOptimizerEnabled:   true,
		EnableDeduplication:      true,
		EnableCompression:        true,
		EnableLazyLoading:        true,
		EnableGCHints:            true,
		
		InheritanceOptimizerEnabled: true,
		MaxInheritanceDepth:         3,
		FlattenInheritance:          true,
		CacheOptimizedTemplates:     true,
		
		ContextOptimizerEnabled:  true,
		ContextDeduplication:     true,
		ContextLazyLoading:       true,
		ContextCompression:       false,
		
		ConcurrencyManagerEnabled: true,
		MaxWorkers:                numCPU * 4,
		MinWorkers:                numCPU,
		QueueSize:                 1000,
		WorkerIdleTimeout:         30,
		
		ExecutionOptimizerEnabled: true,
		EnableBatchProcessing:     true,
		BatchSize:                 10,
		ResultCacheSize:           1000,
		ResultCacheTTL:            1800,
		
		TunerEnabled:              true,
		GCPercent:                 100,
		MaxConcurrentRequests:     numCPU * 100,
		ConnectionPoolSize:        numCPU * 10,
		BufferPoolSize:            1000,
		
		// Initialize static file handler with default settings
		StaticFileHandler:         DefaultStaticFileHandlerConfig(),
		
		CustomConfig:              make(map[string]interface{}),
	}

// GetMemoryConfig returns the memory configuration
func GetMemoryConfig() *MemoryConfig {
	mutex.RLock()
	if instance != nil {
		defer mutex.RUnlock()
		return instance
	}
	mutex.RUnlock()
	
	mutex.Lock()
	defer mutex.Unlock()
	
	// Double-check after acquiring lock
	if instance != nil {
		return instance
	}
	
	// Create default configuration
	instance = DefaultMemoryConfig()
	
	// Load configuration from environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = string(Development)
	}
	
	// Set environment
	instance.Environment = Environment(strings.ToLower(env))
	
	// Load configuration from file
	configFile := fmt.Sprintf("memory_config_%s.json", instance.Environment)
	configPath := filepath.Join(configDir, configFile)
	
	// Check if configuration file exists
	if _, err := os.Stat(configPath); err == nil {
	}		// Load configuration from file
		data, err := os.ReadFile(filepath.Clean(configPath))
		if err == nil {
			if err := json.Unmarshal(data, instance); err != nil {
				fmt.Printf("Failed to parse configuration file: %v\n", err)
			}
		} else {
			fmt.Printf("Failed to read configuration file: %v\n", err)
		}
	}
	
	// Override with environment variables
	instance.loadFromEnv()
	
	return instance
	

// SaveConfig saves the memory configuration to a file
func (c *MemoryConfig) SaveConfig() error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Create config file
	configFile := fmt.Sprintf("memory_config_%s.json", c.Environment)
	configPath := filepath.Join(configDir, configFile)
	
	// Marshal configuration to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}
	
	// Write configuration to file
	if err := os.WriteFile(filepath.Clean(configPath, data, 0600)); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}
	
	return nil

// loadFromEnv loads configuration from environment variables
func (c *MemoryConfig) loadFromEnv() {
	// Load profiler configuration
	if val := os.Getenv("MEMORY_PROFILER_ENABLED"); val != "" {
		c.ProfilerEnabled = val == "true"
	}
	if val := os.Getenv("MEMORY_PROFILER_INTERVAL"); val != "" {
		if interval, err := parseInt(val); err == nil {
			c.ProfilerInterval = interval
		}
	}
	if val := os.Getenv("MEMORY_PROFILER_OUTPUT_DIR"); val != "" {
		c.ProfilerOutputDir = val
	}
	if val := os.Getenv("MEMORY_THRESHOLD"); val != "" {
		if threshold, err := parseInt64(val); err == nil {
			c.MemoryThreshold = threshold
		}
	}
	if val := os.Getenv("GC_THRESHOLD"); val != "" {
		if threshold, err := parseInt64(val); err == nil {
			c.GCThreshold = threshold
		}
	}

	// Load pool manager configuration
	if val := os.Getenv("POOL_MANAGER_ENABLED"); val != "" {
		c.PoolManagerEnabled = val == "true"
	}
	if val := os.Getenv("DEFAULT_POOL_SIZE"); val != "" {
		if size, err := parseInt(val); err == nil {
			c.DefaultPoolSize = size
		}
	}
	if val := os.Getenv("MIN_POOL_SIZE"); val != "" {
		if size, err := parseInt(val); err == nil {
			c.MinPoolSize = size
		}
	}
	if val := os.Getenv("MAX_POOL_SIZE"); val != "" {
		if size, err := parseInt(val); err == nil {
			c.MaxPoolSize = size
		}
	}
	// Load memory optimizer configuration
	if val := os.Getenv("MEMORY_OPTIMIZER_ENABLED"); val != "" {
		c.MemoryOptimizerEnabled = val == "true"
	}
	if val := os.Getenv("ENABLE_DEDUPLICATION"); val != "" {
		c.EnableDeduplication = val == "true"
	}
	if val := os.Getenv("ENABLE_COMPRESSION"); val != "" {
		c.EnableCompression = val == "true"
	}
	if val := os.Getenv("ENABLE_LAZY_LOADING"); val != "" {
		c.EnableLazyLoading = val == "true"
	}
	if val := os.Getenv("ENABLE_GC_HINTS"); val != "" {
		c.EnableGCHints = val == "true"
	}

	// Load inheritance optimizer configuration
	if val := os.Getenv("INHERITANCE_OPTIMIZER_ENABLED"); val != "" {
		c.InheritanceOptimizerEnabled = val == "true"
	}
	if val := os.Getenv("MAX_INHERITANCE_DEPTH"); val != "" {
		if depth, err := parseInt(val); err == nil {
			c.MaxInheritanceDepth = depth
		}
	}
	if val := os.Getenv("FLATTEN_INHERITANCE"); val != "" {
		c.FlattenInheritance = val == "true"
	}
	if val := os.Getenv("CACHE_OPTIMIZED_TEMPLATES"); val != "" {
		c.CacheOptimizedTemplates = val == "true"
	}

	// Load context optimizer configuration
	if val := os.Getenv("CONTEXT_OPTIMIZER_ENABLED"); val != "" {
		c.ContextOptimizerEnabled = val == "true"
	}
	if val := os.Getenv("CONTEXT_DEDUPLICATION"); val != "" {
		c.ContextDeduplication = val == "true"
	}
	if val := os.Getenv("CONTEXT_LAZY_LOADING"); val != "" {
		c.ContextLazyLoading = val == "true"
	}
	if val := os.Getenv("CONTEXT_COMPRESSION"); val != "" {
		c.ContextCompression = val == "true"
	}

	// Load concurrency configuration
	if val := os.Getenv("CONCURRENCY_MANAGER_ENABLED"); val != "" {
		c.ConcurrencyManagerEnabled = val == "true"
	}
	if val := os.Getenv("MAX_WORKERS"); val != "" {
		if workers, err := parseInt(val); err == nil {
			c.MaxWorkers = workers
		}
	}
	if val := os.Getenv("MIN_WORKERS"); val != "" {
		if workers, err := parseInt(val); err == nil {
			c.MinWorkers = workers
		}
	}

	// Load tuner configuration
	if val := os.Getenv("TUNER_ENABLED"); val != "" {
		c.TunerEnabled = val == "true"
	}
	if val := os.Getenv("GC_PERCENT"); val != "" {
		if percent, err := parseInt(val); err == nil {
			c.GCPercent = percent
		}
	}

// parseInt parses an integer from a string
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err

// parseInt64 parses a 64-bit integer from a string
func parseInt64(s string) (int64, error) {
	var i int64
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err

// GetEnvironment returns the current environment
func (c *MemoryConfig) GetEnvironment() Environment {
	return c.Environment

// SetEnvironment sets the environment
func (c *MemoryConfig) SetEnvironment(env Environment) {
	c.Environment = env

// IsProduction returns true if the environment is production
func (c *MemoryConfig) IsProduction() bool {
	return c.Environment == Production

// IsTesting returns true if the environment is testing
func (c *MemoryConfig) IsTesting() bool {
	return c.Environment == Testing

// IsDevelopment returns true if the environment is development
func (c *MemoryConfig) IsDevelopment() bool {
	return c.Environment == Development

// GetCustomConfig gets a custom configuration value
func (c *MemoryConfig) GetCustomConfig(key string) (interface{}, bool) {
	value, ok := c.CustomConfig[key]
	return value, ok

// SetCustomConfig sets a custom configuration value
func (c *MemoryConfig) SetCustomConfig(key string, value interface{}) {
	c.CustomConfig[key] = value

// Clone creates a deep copy of the configuration
func (c *MemoryConfig) Clone() *MemoryConfig {
	clone := &MemoryConfig{
		Environment:              c.Environment,

		ProfilerEnabled:          c.ProfilerEnabled,
		ProfilerInterval:         c.ProfilerInterval,
		ProfilerOutputDir:        c.ProfilerOutputDir,
		MemoryThreshold:          c.MemoryThreshold,
		GCThreshold:              c.GCThreshold,

		PoolManagerEnabled:       c.PoolManagerEnabled,
		DefaultPoolSize:          c.DefaultPoolSize,
		MinPoolSize:              c.MinPoolSize,
		MaxPoolSize:              c.MaxPoolSize,
		EnablePoolScaling:        c.EnablePoolScaling,
		ScaleUpThreshold:         c.ScaleUpThreshold,
		ScaleDownThreshold:       c.ScaleDownThreshold,

		MemoryOptimizerEnabled:   c.MemoryOptimizerEnabled,
		EnableDeduplication:      c.EnableDeduplication,
		EnableCompression:        c.EnableCompression,
		EnableLazyLoading:        c.EnableLazyLoading,
		EnableGCHints:            c.EnableGCHints,

		InheritanceOptimizerEnabled: c.InheritanceOptimizerEnabled,
		MaxInheritanceDepth:         c.MaxInheritanceDepth,
		FlattenInheritance:          c.FlattenInheritance,
		CacheOptimizedTemplates:     c.CacheOptimizedTemplates,

		ContextOptimizerEnabled:  c.ContextOptimizerEnabled,
		ContextDeduplication:     c.ContextDeduplication,
		ContextLazyLoading:       c.ContextLazyLoading,
		ContextCompression:       c.ContextCompression,

		ConcurrencyManagerEnabled: c.ConcurrencyManagerEnabled,
		MaxWorkers:                c.MaxWorkers,
		MinWorkers:                c.MinWorkers,
		QueueSize:                 c.QueueSize,
		WorkerIdleTimeout:         c.WorkerIdleTimeout,

		ExecutionOptimizerEnabled: c.ExecutionOptimizerEnabled,
		EnableBatchProcessing:     c.EnableBatchProcessing,
		BatchSize:                 c.BatchSize,
		ResultCacheSize:           c.ResultCacheSize,
		ResultCacheTTL:            c.ResultCacheTTL,

		TunerEnabled:              c.TunerEnabled,
		GCPercent:                 c.GCPercent,
		MaxConcurrentRequests:     c.MaxConcurrentRequests,
		ConnectionPoolSize:        c.ConnectionPoolSize,
		BufferPoolSize:            c.BufferPoolSize,

		CustomConfig:              make(map[string]interface{}),
	}

	// Copy custom config
	for key, value := range c.CustomConfig {
		clone.CustomConfig[key] = value
	}

	return clone

// Reset resets the configuration to default values
func ResetConfig() {
	mutex.Lock()
	defer mutex.Unlock()
	
