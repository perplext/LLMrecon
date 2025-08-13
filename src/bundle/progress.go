package bundle

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// ProgressEvent represents a progress update
type ProgressEvent struct {
	Stage           ProgressStage          // Current stage of operation
	Operation       string                 // Current operation description
	Current         int64                  // Current progress value
	Total           int64                  // Total expected value
	Percentage      float64                // Percentage complete (0-100)
	BytesProcessed  int64                  // Bytes processed so far
	BytesTotal      int64                  // Total bytes to process
	ItemsProcessed  int                    // Number of items processed
	ItemsTotal      int                    // Total number of items
	TimeElapsed     time.Duration          // Time elapsed since start
	TimeRemaining   time.Duration          // Estimated time remaining
	Speed           float64                // Processing speed (bytes/sec)
	CurrentFile     string                 // Current file being processed
	Error           error                  // Error if any
	Metadata        map[string]interface{} // Additional metadata
}

// ProgressStage represents different stages of bundle operations
type ProgressStage string

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

// ProgressHandler is a callback for progress updates
type ProgressHandler func(event ProgressEvent)

// ProgressTracker tracks progress of bundle operations
type ProgressTracker struct {
	mu              sync.RWMutex
	stage           ProgressStage
	operation       string
	startTime       time.Time
	lastUpdate      time.Time
	bytesProcessed  int64
	bytesTotal      int64
	itemsProcessed  int32
	itemsTotal      int32
	currentFile     string
	handlers        []ProgressHandler
	updateInterval  time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	done            chan struct{}
	metadata        map[string]interface{}
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(ctx context.Context) *ProgressTracker {
	if ctx == nil {
		ctx = context.Background()
	}
	
	trackerCtx, cancel := context.WithCancel(ctx)
	
	p := &ProgressTracker{
		stage:          StageInitializing,
		startTime:      time.Now(),
		lastUpdate:     time.Now(),
		handlers:       make([]ProgressHandler, 0),
		updateInterval: 100 * time.Millisecond,
		ctx:            trackerCtx,
		cancel:         cancel,
		done:           make(chan struct{}),
		metadata:       make(map[string]interface{}),
	}
	
	// Start background updater
	go p.backgroundUpdater()
	
	return p
}

// AddHandler adds a progress handler
func (p *ProgressTracker) AddHandler(handler ProgressHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, handler)
}

// SetStage updates the current stage
func (p *ProgressTracker) SetStage(stage ProgressStage, operation string) {
	p.mu.Lock()
	p.stage = stage
	p.operation = operation
	p.mu.Unlock()
	
	p.sendUpdate()
}

// SetTotal sets the total bytes and items
func (p *ProgressTracker) SetTotal(bytes int64, items int) {
	p.mu.Lock()
	p.bytesTotal = bytes
	atomic.StoreInt32(&p.itemsTotal, int32(items))
	p.mu.Unlock()
	
	p.sendUpdate()
}

// UpdateBytes updates bytes processed
func (p *ProgressTracker) UpdateBytes(bytes int64) {
	atomic.AddInt64(&p.bytesProcessed, bytes)
}

// UpdateItems updates items processed
func (p *ProgressTracker) UpdateItems(count int) {
	atomic.AddInt32(&p.itemsProcessed, int32(count))
}

// SetCurrentFile sets the current file being processed
func (p *ProgressTracker) SetCurrentFile(file string) {
	p.mu.Lock()
	p.currentFile = file
	p.mu.Unlock()
}

// SetMetadata sets custom metadata
func (p *ProgressTracker) SetMetadata(key string, value interface{}) {
	p.mu.Lock()
	p.metadata[key] = value
	p.mu.Unlock()
}

// Fail marks the operation as failed
func (p *ProgressTracker) Fail(err error) {
	p.mu.Lock()
	p.stage = StageFailed
	p.mu.Unlock()
	
	event := p.createEvent()
	event.Error = err
	p.notifyHandlers(event)
	
	p.Close()
}

// Complete marks the operation as completed
func (p *ProgressTracker) Complete() {
	p.mu.Lock()
	p.stage = StageCompleted
	p.operation = "Bundle operation completed successfully"
	p.mu.Unlock()
	
	p.sendUpdate()
	p.Close()
}

// Close stops the progress tracker
func (p *ProgressTracker) Close() {
	p.cancel()
	<-p.done
}

// backgroundUpdater sends periodic updates
func (p *ProgressTracker) backgroundUpdater() {
	defer close(p.done)
	
	ticker := time.NewTicker(p.updateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.sendUpdate()
		}
	}
}

// sendUpdate sends a progress update
func (p *ProgressTracker) sendUpdate() {
	if time.Since(p.lastUpdate) < p.updateInterval/2 {
		return // Rate limit updates
	}
	
	p.mu.Lock()
	p.lastUpdate = time.Now()
	p.mu.Unlock()
	
	event := p.createEvent()
	p.notifyHandlers(event)
}

// createEvent creates a progress event
func (p *ProgressTracker) createEvent() ProgressEvent {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	bytesProcessed := atomic.LoadInt64(&p.bytesProcessed)
	itemsProcessed := atomic.LoadInt32(&p.itemsProcessed)
	itemsTotal := atomic.LoadInt32(&p.itemsTotal)
	
	elapsed := time.Since(p.startTime)
	
	var percentage float64
	if p.bytesTotal > 0 {
		percentage = float64(bytesProcessed) / float64(p.bytesTotal) * 100
	} else if itemsTotal > 0 {
		percentage = float64(itemsProcessed) / float64(itemsTotal) * 100
	}
	
	var speed float64
	if elapsed.Seconds() > 0 {
		speed = float64(bytesProcessed) / elapsed.Seconds()
	}
	
	var remaining time.Duration
	if speed > 0 && p.bytesTotal > bytesProcessed {
		remainingBytes := p.bytesTotal - bytesProcessed
		remaining = time.Duration(float64(remainingBytes)/speed) * time.Second
	}
	
	// Copy metadata to avoid race conditions
	metadata := make(map[string]interface{})
	for k, v := range p.metadata {
		metadata[k] = v
	}
	
	return ProgressEvent{
		Stage:          p.stage,
		Operation:      p.operation,
		Current:        bytesProcessed,
		Total:          p.bytesTotal,
		Percentage:     percentage,
		BytesProcessed: bytesProcessed,
		BytesTotal:     p.bytesTotal,
		ItemsProcessed: int(itemsProcessed),
		ItemsTotal:     int(itemsTotal),
		TimeElapsed:    elapsed,
		TimeRemaining:  remaining,
		Speed:          speed,
		CurrentFile:    p.currentFile,
		Metadata:       metadata,
	}
}

// notifyHandlers notifies all registered handlers
func (p *ProgressTracker) notifyHandlers(event ProgressEvent) {
	p.mu.RLock()
	handlers := make([]ProgressHandler, len(p.handlers))
	copy(handlers, p.handlers)
	p.mu.RUnlock()
	
	for _, handler := range handlers {
		handler(event)
	}
}

// ProgressReader wraps a reader to track reading progress
type ProgressReader struct {
	reader  io.Reader
	tracker *ProgressTracker
	read    int64
}

// NewProgressReader creates a new progress reader
func NewProgressReader(reader io.Reader, tracker *ProgressTracker) *ProgressReader {
	return &ProgressReader{
		reader:  reader,
		tracker: tracker,
	}
}

// Read implements io.Reader
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.read += int64(n)
		pr.tracker.UpdateBytes(int64(n))
	}
	return
}

// ProgressWriter wraps a writer to track writing progress
type ProgressWriter struct {
	writer  io.Writer
	tracker *ProgressTracker
	written int64
}

// NewProgressWriter creates a new progress writer
func NewProgressWriter(writer io.Writer, tracker *ProgressTracker) *ProgressWriter {
	return &ProgressWriter{
		writer:  writer,
		tracker: tracker,
	}
}

// Write implements io.Writer
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	if n > 0 {
		pw.written += int64(n)
		pw.tracker.UpdateBytes(int64(n))
	}
	return
}

// ConsoleProgressHandler creates a console progress handler
func ConsoleProgressHandler() ProgressHandler {
	var lastPercentage int
	
	return func(event ProgressEvent) {
		percentage := int(event.Percentage)
		
		// Only update if percentage changed or important stage change
		if percentage == lastPercentage && 
			event.Stage != StageCompleted && 
			event.Stage != StageFailed {
			return
		}
		
		lastPercentage = percentage
		
		// Clear line and print progress
		fmt.Printf("\r\033[K[%s] %s: %d%% (%s/%s) - %s",
			event.Stage,
			event.Operation,
			percentage,
			formatBytes(event.BytesProcessed),
			formatBytes(event.BytesTotal),
			formatDuration(event.TimeRemaining),
		)
		
		// Print newline on completion or failure
		if event.Stage == StageCompleted || event.Stage == StageFailed {
			fmt.Println()
			if event.Error != nil {
				fmt.Printf("Error: %v\n", event.Error)
			}
		}
	}
}

// JSONProgressHandler creates a JSON progress handler
func JSONProgressHandler(writer io.Writer) ProgressHandler {
	encoder := json.NewEncoder(writer)
	
	return func(event ProgressEvent) {
		// Create a simplified event for JSON encoding
		jsonEvent := map[string]interface{}{
			"stage":           string(event.Stage),
			"operation":       event.Operation,
			"percentage":      event.Percentage,
			"bytes_processed": event.BytesProcessed,
			"bytes_total":     event.BytesTotal,
			"items_processed": event.ItemsProcessed,
			"items_total":     event.ItemsTotal,
			"time_elapsed":    event.TimeElapsed.Seconds(),
			"time_remaining":  event.TimeRemaining.Seconds(),
			"speed":           event.Speed,
			"current_file":    event.CurrentFile,
			"metadata":        event.Metadata,
		}
		
		if event.Error != nil {
			jsonEvent["error"] = event.Error.Error()
		}
		
		encoder.Encode(jsonEvent)
	}
}

// Helper functions
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "calculating..."
	}
	if d == 0 {
		return "complete"
	}
	
	if d < time.Minute {
		return fmt.Sprintf("%ds remaining", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds remaining", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm remaining", int(d.Hours()), int(d.Minutes())%60)
}