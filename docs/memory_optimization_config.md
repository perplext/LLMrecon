# Memory Optimization Configuration Guide

This document provides a comprehensive guide to the memory optimization and configuration system implemented for the LLMrecon project. The system is designed to optimize memory usage and tune server parameters across different environments (development, testing, and production).

## Table of Contents

1. [Overview](#overview)
2. [Components](#components)
3. [Configuration System](#configuration-system)
4. [Environment-Specific Configurations](#environment-specific-configurations)
5. [Configuration Manager Tool](#configuration-manager-tool)
6. [Integration Examples](#integration-examples)
7. [Best Practices](#best-practices)

## Overview

The memory optimization system aims to reduce memory footprint by at least 25% and support at least 2x the current concurrent users without degradation. It achieves this through a combination of memory profiling, resource pooling, template optimization, concurrency management, and automatic configuration tuning.

## Components

The system consists of the following components:

### Memory Profiler (`src/utils/profiling/memory_profiler.go`)

- Monitors memory usage during peak loads
- Captures heap profiles and provides memory statistics
- Implements automatic profiling with configurable intervals and thresholds

### Resource Pool Manager (`src/utils/resource/pool_manager.go`)

- Manages resource pools for efficient resource utilization
- Implements dynamic scaling based on utilization metrics
- Provides resource validation and cleanup functions

### Memory Optimizer (`src/template/management/optimization/memory_optimizer.go`)

- Optimizes template memory usage through deduplication
- Implements variable optimization and inheritance flattening
- Provides garbage collection hints for better memory management

### Configuration Tuner (`src/utils/config/tuner.go`)

- Automatically tunes server configuration parameters
- Adjusts worker count, connection pool size, and GC settings
- Provides recommendations for configuration improvements

### Concurrency Manager (`src/utils/concurrency/manager.go`)

- Manages concurrent operations with dynamic worker scaling
- Implements task prioritization and queue management
- Provides detailed statistics for monitoring concurrency patterns

### Template Execution Optimizer (`src/template/management/execution/optimizer/execution_optimizer.go`)

- Optimizes template execution for memory efficiency
- Combines memory optimization with concurrency management
- Supports batch processing for improved throughput

## Configuration System

The configuration system (`src/utils/config/memory_config.go`) provides a centralized way to manage memory optimization settings across different environments. It includes:

- Environment-specific configuration files
- Configuration loading from files and environment variables
- Default configuration values for each component
- Validation of configuration parameters
- Custom configuration options

### Configuration Structure

The `MemoryConfig` struct contains configuration options for all memory optimization components:

```go
type MemoryConfig struct {
    Environment Environment `json:"environment"`
    
    // Memory Profiler Configuration
    ProfilerEnabled       bool   `json:"profiler_enabled"`
    ProfilerInterval      int    `json:"profiler_interval"`
    ProfilerOutputDir     string `json:"profiler_output_dir"`
    MemoryThreshold       int64  `json:"memory_threshold"`
    GCThreshold           int64  `json:"gc_threshold"`
    
    // Resource Pool Configuration
    PoolManagerEnabled    bool   `json:"pool_manager_enabled"`
    DefaultPoolSize       int    `json:"default_pool_size"`
    MinPoolSize           int    `json:"min_pool_size"`
    MaxPoolSize           int    `json:"max_pool_size"`
    EnablePoolScaling     bool   `json:"enable_pool_scaling"`
    ScaleUpThreshold      float64 `json:"scale_up_threshold"`
    ScaleDownThreshold    float64 `json:"scale_down_threshold"`
    
    // Memory Optimizer Configuration
    MemoryOptimizerEnabled bool   `json:"memory_optimizer_enabled"`
    EnableDeduplication    bool   `json:"enable_deduplication"`
    EnableCompression      bool   `json:"enable_compression"`
    EnableLazyLoading      bool   `json:"enable_lazy_loading"`
    EnableGCHints          bool   `json:"enable_gc_hints"`
    
    // Concurrency Configuration
    ConcurrencyManagerEnabled bool   `json:"concurrency_manager_enabled"`
    MaxWorkers                int    `json:"max_workers"`
    MinWorkers                int    `json:"min_workers"`
    QueueSize                 int    `json:"queue_size"`
    WorkerIdleTimeout         int    `json:"worker_idle_timeout"`
    
    // Execution Optimizer Configuration
    ExecutionOptimizerEnabled bool   `json:"execution_optimizer_enabled"`
    EnableBatchProcessing     bool   `json:"enable_batch_processing"`
    BatchSize                 int    `json:"batch_size"`
    ResultCacheSize           int    `json:"result_cache_size"`
    ResultCacheTTL            int    `json:"result_cache_ttl"`
    
    // Tuner Configuration
    TunerEnabled           bool   `json:"tuner_enabled"`
    GCPercent              int    `json:"gc_percent"`
    MaxConcurrentRequests  int    `json:"max_concurrent_requests"`
    ConnectionPoolSize     int    `json:"connection_pool_size"`
    BufferPoolSize         int    `json:"buffer_pool_size"`
    
    // Custom configuration by environment
    CustomConfig           map[string]interface{} `json:"custom_config"`
}
```

### Using the Configuration System

To use the configuration system in your code:

```go
import "github.com/LLMrecon/src/utils/config"

// Get the memory configuration
memConfig := config.GetMemoryConfig()

// Use configuration values
if memConfig.MemoryOptimizerEnabled {
    // Use memory optimizer
}

// Check environment
if memConfig.IsProduction() {
    // Production-specific logic
}

// Get custom configuration value
if value, ok := memConfig.GetCustomConfig("custom_key"); ok {
    // Use custom value
}
```

## Environment-Specific Configurations

The system provides default configurations for three environments:

### Development Environment (`src/utils/config/environments/dev.json`)

- Optimized for development and debugging
- Lower resource limits to work well on development machines
- More verbose logging and debugging options
- Faster profiling intervals for immediate feedback

### Testing Environment (`src/utils/config/environments/test.json`)

- Configured for load testing and performance testing
- Moderate resource limits to simulate production-like conditions
- Balanced between performance and debugging capabilities
- Configured to simulate production loads

### Production Environment (`src/utils/config/environments/prod.json`)

- Optimized for maximum performance and resource efficiency
- Higher resource limits to handle production loads
- Minimal debugging overhead
- Configured for high availability and reliability

## Configuration Manager Tool

The Configuration Manager Tool (`cmd/config-manager/main.go`) provides a command-line interface for managing memory optimization configurations:

### Usage

```
Usage: config-manager [options]

Options:
  -env string      Environment (dev, test, prod)
  -show            Show current configuration
  -export          Export configuration to file
  -apply           Apply configuration from file
  -reset           Reset configuration to defaults
  -file string     Configuration file path
  -compare         Compare configurations between environments
  -validate        Validate configuration
  -set string      Set configuration value (key=value)
  -get string      Get configuration value
```

### Examples

Show current configuration:
```
config-manager -show
```

Set environment:
```
config-manager -env=prod
```

Compare configurations:
```
config-manager -compare
```

Set configuration value:
```
config-manager -set="max_workers=16"
```

Get configuration value:
```
config-manager -get="max_workers"
```

Validate configuration:
```
config-manager -validate
```

## Integration Examples

### Using Memory Optimizer

```go
import (
    "context"
    "github.com/LLMrecon/src/template/format"
    "github.com/LLMrecon/src/template/management/optimization"
    "github.com/LLMrecon/src/utils/config"
)

func optimizeTemplate(template *format.Template) (*format.Template, error) {
    // Get memory configuration
    memConfig := config.GetMemoryConfig()
    
    // Create memory optimizer
    optimizer := optimization.NewMemoryOptimizer(&optimization.MemoryOptimizerOptions{
        EnableDeduplication: memConfig.EnableDeduplication,
        EnableCompression:   memConfig.EnableCompression,
        EnableLazyLoading:   memConfig.EnableLazyLoading,
        EnableGCHints:       memConfig.EnableGCHints,
    })
    
    // Optimize template
    optimizedTemplate, err := optimizer.OptimizeTemplate(context.Background(), template)
    if err != nil {
        return nil, err
    }
    
    return optimizedTemplate, nil
}
```

### Using Resource Pool Manager

```go
import (
    "github.com/LLMrecon/src/utils/config"
    "github.com/LLMrecon/src/utils/resource"
)

func createConnectionPool() *resource.Pool {
    // Get memory configuration
    memConfig := config.GetMemoryConfig()
    
    // Create resource pool manager
    poolManager := resource.NewPoolManager(&resource.PoolManagerOptions{
        DefaultPoolSize:    memConfig.DefaultPoolSize,
        MinPoolSize:        memConfig.MinPoolSize,
        MaxPoolSize:        memConfig.MaxPoolSize,
        EnablePoolScaling:  memConfig.EnablePoolScaling,
        ScaleUpThreshold:   memConfig.ScaleUpThreshold,
        ScaleDownThreshold: memConfig.ScaleDownThreshold,
    })
    
    // Create connection pool
    pool := poolManager.CreatePool("connections", func() (interface{}, error) {
        // Create new connection
        return createConnection(), nil
    }, func(resource interface{}) error {
        // Close connection
        conn := resource.(Connection)
        return conn.Close()
    })
    
    return pool
}
```

### Using Concurrency Manager

```go
import (
    "context"
    "github.com/LLMrecon/src/utils/config"
    "github.com/LLMrecon/src/utils/concurrency"
)

func processItems(items []Item) error {
    // Get memory configuration
    memConfig := config.GetMemoryConfig()
    
    // Create concurrency manager
    manager := concurrency.NewConcurrencyManager(&concurrency.ConcurrencyManagerOptions{
        MaxWorkers:        memConfig.MaxWorkers,
        MinWorkers:        memConfig.MinWorkers,
        QueueSize:         memConfig.QueueSize,
        WorkerIdleTimeout: memConfig.WorkerIdleTimeout,
    })
    
    // Process items concurrently
    for _, item := range items {
        item := item // Create local copy for closure
        err := manager.Submit(func(ctx context.Context) error {
            return processItem(ctx, item)
        })
        if err != nil {
            return err
        }
    }
    
    // Wait for all tasks to complete
    return manager.Wait(context.Background())
}
```

## Best Practices

1. **Environment-Specific Configuration**
   - Use environment-specific configurations for dev, test, and prod
   - Adjust resource limits based on environment requirements
   - Enable more debugging options in development

2. **Memory Profiling**
   - Enable memory profiling in all environments
   - Use shorter intervals in development for immediate feedback
   - Use longer intervals in production to minimize overhead
   - Set appropriate memory thresholds for alerts

3. **Resource Pool Management**
   - Configure pool sizes based on available resources
   - Enable pool scaling for dynamic resource utilization
   - Set appropriate scale thresholds based on load patterns
   - Implement proper resource cleanup functions

4. **Concurrency Management**
   - Configure worker counts based on CPU cores and workload
   - Set appropriate queue sizes based on expected load
   - Implement task prioritization for critical operations
   - Monitor worker utilization for bottlenecks

5. **Template Optimization**
   - Enable deduplication for all environments
   - Enable compression in production for memory savings
   - Use lazy loading for large templates
   - Provide garbage collection hints for better memory management

6. **Configuration Tuning**
   - Regularly review and adjust configuration parameters
   - Use the Configuration Manager Tool to compare configurations
   - Validate configuration changes before deployment
   - Monitor the impact of configuration changes on performance

7. **Monitoring and Alerting**
   - Set up monitoring for memory usage and resource utilization
   - Configure alerts for when memory usage exceeds thresholds
   - Capture heap profiles for analysis when issues occur
   - Review memory usage patterns regularly
