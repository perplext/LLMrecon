# Static File Handler Monitoring Integration

This document describes how to integrate the static file handler with the monitoring system to track performance metrics and receive alerts for potential issues.

## Overview

The static file handler monitoring integration provides real-time metrics and alerts for the static file handler, allowing administrators to monitor performance, identify bottlenecks, and receive notifications when certain thresholds are exceeded.

## Key Features

- **Real-time Metrics**: Track key performance indicators such as cache hit ratio, average serve time, and compression ratio.
- **Automatic Alerts**: Receive alerts when metrics exceed predefined thresholds.
- **Integration with Existing Monitoring**: Seamlessly integrates with the existing monitoring system.
- **Customizable Thresholds**: Configure alert thresholds based on your specific requirements.

## Metrics Tracked

The static file handler monitoring integration tracks the following metrics:

| Metric | Description |
|--------|-------------|
| Files Served | Total number of files served |
| Cache Hits | Number of files served from cache |
| Cache Misses | Number of files served from disk |
| Cache Hit Ratio | Percentage of files served from cache |
| Compressed Files | Number of files served with compression |
| Compression Ratio | Average compression ratio |
| Cache Size | Current size of the cache in bytes |
| Cache Item Count | Number of items in the cache |
| Average Serve Time | Average time to serve a file |

## Alert Rules

The static file handler monitoring integration provides the following alert rules:

| Rule | Description | Default Threshold | Severity |
|------|-------------|-------------------|----------|
| Cache Nearly Full | Cache size is approaching the maximum | 90% | Warning |
| Low Cache Hit Ratio | Cache hit ratio is below threshold | 50% | Info |
| Slow Serve Time | Average serve time is above threshold | 50ms | Warning |

## Integration Steps

### 1. Create a Static File Handler

```go
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
```

### 2. Create a Monitoring Service

```go
// Create monitoring service
monitoringOptions := monitoring.DefaultMonitoringServiceOptions()
monitoringService, err := monitoring.NewMonitoringService(monitoringOptions)
if err != nil {
    log.Fatalf("Failed to create monitoring service: %v", err)
}
```

### 3. Add Static File Handler to Monitoring Service

```go
// Add static file handler to monitoring service
staticFileMonitor := monitoringService.AddStaticFileMonitor(fileHandler)
```

### 4. Access Monitoring Metrics

```go
// Get monitoring metrics
metrics := staticFileMonitor.GetMetrics()

// Access specific metrics
fmt.Printf("Files Served: %d\n", metrics.FilesServed)
fmt.Printf("Cache Hit Ratio: %.2f%%\n", metrics.CacheHitRatio * 100)
fmt.Printf("Average Serve Time: %s\n", metrics.AverageServeTime)
```

### 5. Customize Alert Rules (Optional)

```go
// Add custom alert rules
staticFileMonitor.AddStaticFileAlertRules(10 * time.Minute) // 10 minute cooldown between alerts
```

## Example Implementation

See the complete example in `examples/memory_optimization/static_file_handler_example.go`.

## Best Practices

1. **Monitor Cache Hit Ratio**: A low cache hit ratio may indicate that the cache size is too small or that the cache expiration time is too short.

2. **Monitor Average Serve Time**: A high average serve time may indicate that the server is overloaded or that the files being served are too large.

3. **Monitor Cache Size**: If the cache is frequently near capacity, consider increasing the maximum cache size.

4. **Enable Compression**: Enable compression for text-based files to reduce bandwidth usage and improve load times.

5. **Set Appropriate Cache Expiration**: Set the cache expiration time based on how frequently your static files change.

## Performance Impact

The monitoring integration adds minimal overhead to the static file handler. In benchmarks, the overhead is typically less than 1% of the total serve time.

## Troubleshooting

### Common Issues

1. **High Memory Usage**: If memory usage is high, consider reducing the maximum cache size or enabling more aggressive cache eviction.

2. **Slow Serve Times**: If serve times are slow, check if compression is enabled and if the files being served are too large.

3. **Low Cache Hit Ratio**: If the cache hit ratio is low, check if the cache size is sufficient and if the cache expiration time is appropriate.

### Debugging

Enable debug logging to see more detailed information about the static file handler and monitoring system:

```go
monitoringOptions.EnableConsoleLogging = true
monitoringOptions.LogLevel = monitoring.DebugLevel
```

## Conclusion

The static file handler monitoring integration provides valuable insights into the performance of the static file handler, allowing administrators to identify and address issues before they impact users. By following the integration steps and best practices outlined in this document, you can ensure that your static file serving system is optimized for performance and reliability.
