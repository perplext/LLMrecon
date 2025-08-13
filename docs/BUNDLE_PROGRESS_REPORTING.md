# Bundle Progress Reporting System

## Overview

The Bundle Progress Reporting System provides real-time tracking and monitoring of bundle export operations. It offers flexible progress handlers, detailed event tracking, and support for both console and programmatic progress monitoring.

## Features

- **Real-time Progress Tracking**: Monitor bundle operations as they happen
- **Multiple Progress Handlers**: Console, JSON, and custom handlers
- **Detailed Event Information**: Stage tracking, time estimation, speed calculation
- **Thread-safe Operations**: Safe for concurrent updates
- **Streaming Progress**: Track reading/writing operations with ProgressReader/Writer
- **Metadata Support**: Attach custom metadata to progress events
- **Error Handling**: Proper failure reporting with error details

## Architecture

### Core Components

1. **ProgressTracker**: Central coordinator for progress events
2. **ProgressEvent**: Detailed progress information structure
3. **ProgressHandler**: Callback interface for progress updates
4. **ProgressReader/Writer**: IO wrappers for streaming progress

### Progress Stages

```go
const (
    StageInitializing   ProgressStage = "initializing"
    StageValidating     ProgressStage = "validating"
    StageCollecting     ProgressStage = "collecting"
    StageCompressing    ProgressStage = "compressing"
    StageEncrypting     ProgressStage = "encrypting"
    StageWriting        ProgressStage = "writing"
    StageVerifying      ProgressStage = "verifying"
    StageFinalizing     ProgressStage = "finalizing"
    StageCompleted      ProgressStage = "completed"
    StageFailed         ProgressStage = "failed"
)
```

## Usage Examples

### Basic Progress Tracking

```go
// Create progress tracker
ctx := context.Background()
tracker := bundle.NewProgressTracker(ctx)
defer tracker.Close()

// Add console handler
tracker.AddHandler(bundle.ConsoleProgressHandler())

// Set total work
tracker.SetTotal(1024*1024*10, 100) // 10MB, 100 files

// Update progress
tracker.SetStage(bundle.StageCollecting, "Collecting files")
tracker.UpdateBytes(1024 * 1024) // 1MB processed
tracker.UpdateItems(10)          // 10 files processed
```

### JSON Progress Logging

```go
// Create log file
logFile, err := os.Create("progress.log")
if err != nil {
    log.Fatal(err)
}
defer logFile.Close()

// Add JSON handler
tracker.AddHandler(bundle.JSONProgressHandler(logFile))
```

### Custom Progress Handler

```go
// Create custom handler
customHandler := func(event bundle.ProgressEvent) {
    // Custom processing
    log.Printf("[%s] %s: %.2f%% complete",
        event.Stage,
        event.Operation,
        event.Percentage)
    
    // Send to monitoring system
    metrics.RecordProgress(event)
}

tracker.AddHandler(customHandler)
```

### Streaming Progress

```go
// Track file reading
file, err := os.Open("large_file.bin")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Get file size
info, _ := file.Stat()
tracker.SetTotal(info.Size(), 1)

// Wrap with progress reader
reader := bundle.NewProgressReader(file, tracker)

// Read with automatic progress tracking
data, err := io.ReadAll(reader)
```

### Integration with Bundle Export

```go
exportOpts := &bundle.ExportOptions{
    OutputPath: "bundle.tar.gz",
    Format:     bundle.FormatTarGz,
    ProgressHandler: func(event bundle.ProgressEvent) {
        // Forward to tracker
        tracker.SetStage(event.Stage, event.Operation)
        tracker.UpdateBytes(event.BytesProcessed)
        tracker.UpdateItems(event.ItemsProcessed)
    },
}

exporter := bundle.NewBundleExporter(".", exportOpts)
err := exporter.Export()
```

## Progress Event Structure

```go
type ProgressEvent struct {
    Stage           ProgressStage          // Current operation stage
    Operation       string                 // Human-readable operation description
    Current         int64                  // Current progress value
    Total           int64                  // Total expected value
    Percentage      float64                // Completion percentage (0-100)
    BytesProcessed  int64                  // Bytes processed so far
    BytesTotal      int64                  // Total bytes to process
    ItemsProcessed  int                    // Number of items processed
    ItemsTotal      int                    // Total number of items
    TimeElapsed     time.Duration          // Time since operation start
    TimeRemaining   time.Duration          // Estimated time to completion
    Speed           float64                // Processing speed (bytes/sec)
    CurrentFile     string                 // Current file being processed
    Error           error                  // Error if operation failed
    Metadata        map[string]interface{} // Custom metadata
}
```

## Console Output Format

The console progress handler displays progress in a user-friendly format:

```
[collecting] Scanning for files: 45% (4.5MB/10.0MB) - 30s remaining
[compressing] Compressing bundle: 78% (7.8MB/10.0MB) - 10s remaining
[completed] Bundle operation completed successfully: 100% (10.0MB/10.0MB) - complete
```

## JSON Output Format

The JSON progress handler outputs events in a structured format:

```json
{
    "stage": "compressing",
    "operation": "Compressing bundle",
    "percentage": 45.5,
    "bytes_processed": 4550000,
    "bytes_total": 10000000,
    "items_processed": 45,
    "items_total": 100,
    "time_elapsed": 30.5,
    "time_remaining": 36.7,
    "speed": 149180.32,
    "current_file": "docs/manual.pdf",
    "metadata": {
        "compression": "gzip",
        "compression_level": 6
    }
}
```

## Best Practices

1. **Update Frequency**: The tracker automatically rate-limits updates to prevent UI flooding
2. **Thread Safety**: All update methods are thread-safe for concurrent use
3. **Context Cancellation**: Use context for graceful shutdown
4. **Error Handling**: Always call `Fail()` on errors for proper cleanup
5. **Completion**: Always call `Complete()` when finished
6. **Metadata**: Use metadata for additional context without modifying the event structure

## Performance Considerations

- Progress tracking adds minimal overhead (~1-2% in benchmarks)
- Background updater runs at configurable intervals (default: 100ms)
- Events are delivered asynchronously to avoid blocking operations
- Memory usage is constant regardless of operation size

## Advanced Features

### Progress Estimation

The system automatically calculates:
- Completion percentage based on bytes or items
- Processing speed in bytes/second
- Estimated time remaining based on current speed
- Accurate time elapsed tracking

### Concurrent Updates

Safe for multiple goroutines:
```go
var wg sync.WaitGroup
for i := 0; i < workers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Process work
        tracker.UpdateBytes(processed)
        tracker.UpdateItems(1)
    }()
}
wg.Wait()
```

### Custom Metadata

Attach operation-specific data:
```go
tracker.SetMetadata("compression_algorithm", "zstd")
tracker.SetMetadata("compression_level", 3)
tracker.SetMetadata("files_skipped", 5)
tracker.SetMetadata("warnings", []string{"large file detected"})
```

## Error Handling

```go
// On error
if err != nil {
    tracker.Fail(fmt.Errorf("export failed: %w", err))
    return err
}

// Handler receives error event
handler := func(event ProgressEvent) {
    if event.Stage == StageFailed {
        log.Printf("Operation failed: %v", event.Error)
        // Cleanup or recovery logic
    }
}
```

## Testing

The progress system includes comprehensive tests:
- Unit tests for all components
- Concurrent update testing
- Progress calculation verification
- Handler notification testing
- Benchmark tests for performance

Run tests:
```bash
go test -v ./src/bundle -run TestProgress
```

## Future Enhancements

1. **Progress Persistence**: Save/restore progress for resumable operations
2. **Progress Aggregation**: Combine multiple trackers for parent/child operations
3. **Web Dashboard**: Real-time web-based progress monitoring
4. **Metrics Integration**: Export to Prometheus, StatsD, etc.
5. **Progress History**: Store historical progress data for analysis