package bundle

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// InteractiveProgressTracker provides enhanced progress tracking with visual feedback
type InteractiveProgressTracker struct {
	bars      map[string]*progressbar.ProgressBar
	mu        sync.RWMutex
	writer    io.Writer
	startTime time.Time
	verbose   bool
}

// NewInteractiveProgressTracker creates a new interactive progress tracker
func NewInteractiveProgressTracker(writer io.Writer, verbose bool) *InteractiveProgressTracker {
	if writer == nil {
		writer = os.Stdout
	}
	return &InteractiveProgressTracker{
		bars:      make(map[string]*progressbar.ProgressBar),
		writer:    writer,
		startTime: time.Now(),
		verbose:   verbose,
	}
}

// StartOperation starts tracking a new operation with a progress bar
func (ipt *InteractiveProgressTracker) StartOperation(operation string, total int, description string) {
	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	bar := progressbar.NewOptions(total,
		progressbar.OptionSetWriter(ipt.writer),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription(fmt.Sprintf("[cyan]%s[reset]", description)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(ipt.writer, "\n")
		}),
	)

	ipt.bars[operation] = bar
}

// UpdateProgress updates the progress for an operation
func (ipt *InteractiveProgressTracker) UpdateProgress(operation string, current int) {
	ipt.mu.RLock()
	bar, exists := ipt.bars[operation]
	ipt.mu.RUnlock()

	if exists {
		bar.Set(current)
	}
}

// IncrementProgress increments progress by 1
func (ipt *InteractiveProgressTracker) IncrementProgress(operation string) {
	ipt.mu.RLock()
	bar, exists := ipt.bars[operation]
	ipt.mu.RUnlock()

	if exists {
		bar.Add(1)
	}
}

// CompleteOperation marks an operation as complete
func (ipt *InteractiveProgressTracker) CompleteOperation(operation string, message string) {
	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	if bar, exists := ipt.bars[operation]; exists {
		bar.Finish()
		color.Green("✓ %s", message)
		delete(ipt.bars, operation)
	}
}

// FailOperation marks an operation as failed
func (ipt *InteractiveProgressTracker) FailOperation(operation string, err error) {
	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	if bar, exists := ipt.bars[operation]; exists {
		bar.Finish()
		color.Red("✗ %s: %v", operation, err)
		delete(ipt.bars, operation)
	}
}

// LogMessage logs a message without affecting progress bars
func (ipt *InteractiveProgressTracker) LogMessage(level, message string) {
	if !ipt.verbose && level == "debug" {
		return
	}

	ipt.mu.Lock()
	defer ipt.mu.Unlock()

	// Temporarily clear progress bars
	for _, bar := range ipt.bars {
		bar.Clear()
	}

	// Print message
	switch level {
	case "info":
		color.Cyan("ℹ %s", message)
	case "warning":
		color.Yellow("⚠ %s", message)
	case "error":
		color.Red("✗ %s", message)
	case "success":
		color.Green("✓ %s", message)
	case "debug":
		color.White("• %s", message)
	default:
		fmt.Fprintln(ipt.writer, message)
	}

	// Restore progress bars
	for _, bar := range ipt.bars {
		bar.RenderBlank()
	}
}

// Summary prints a summary of the operation
func (ipt *InteractiveProgressTracker) Summary(stats map[string]interface{}) {
	fmt.Fprintln(ipt.writer)
	color.Cyan("═══════════════════════════════════════════")
	color.Cyan("          Operation Summary")
	color.Cyan("═══════════════════════════════════════════")
	
	for key, value := range stats {
		fmt.Fprintf(ipt.writer, "%-20s: %v\n", key, value)
	}
	
	fmt.Fprintf(ipt.writer, "%-20s: %s\n", "Duration", time.Since(ipt.startTime).Round(time.Second))
	color.Cyan("═══════════════════════════════════════════")
}

// BundleProgressReporter provides specialized progress reporting for bundle operations
type BundleProgressReporter struct {
	tracker *InteractiveProgressTracker
	stats   map[string]interface{}
	mu      sync.RWMutex
}

// NewBundleProgressReporter creates a new bundle progress reporter
func NewBundleProgressReporter(writer io.Writer, verbose bool) *BundleProgressReporter {
	return &BundleProgressReporter{
		tracker: NewInteractiveProgressTracker(writer, verbose),
		stats:   make(map[string]interface{}),
	}
}

// ReportBundleCreation reports progress for bundle creation
func (bpr *BundleProgressReporter) ReportBundleCreation(totalTemplates int) {
	bpr.tracker.StartOperation("create", totalTemplates, "Creating bundle")
	bpr.updateStat("Total Templates", totalTemplates)
}

// ReportTemplateProcessed reports a template has been processed
func (bpr *BundleProgressReporter) ReportTemplateProcessed(templateName string) {
	bpr.tracker.IncrementProgress("create")
	bpr.incrementStat("Templates Processed")
	bpr.tracker.LogMessage("debug", fmt.Sprintf("Processed template: %s", templateName))
}

// ReportBundleVerification reports progress for bundle verification
func (bpr *BundleProgressReporter) ReportBundleVerification(checks int) {
	bpr.tracker.StartOperation("verify", checks, "Verifying bundle")
	bpr.updateStat("Total Checks", checks)
}

// ReportVerificationCheck reports a verification check completed
func (bpr *BundleProgressReporter) ReportVerificationCheck(checkName string, passed bool) {
	bpr.tracker.IncrementProgress("verify")
	
	if passed {
		bpr.incrementStat("Checks Passed")
		bpr.tracker.LogMessage("success", fmt.Sprintf("✓ %s", checkName))
	} else {
		bpr.incrementStat("Checks Failed")
		bpr.tracker.LogMessage("error", fmt.Sprintf("✗ %s", checkName))
	}
}

// ReportImportProgress reports progress for bundle import
func (bpr *BundleProgressReporter) ReportImportProgress(totalFiles int) {
	bpr.tracker.StartOperation("import", totalFiles, "Importing bundle")
	bpr.updateStat("Total Files", totalFiles)
}

// ReportFileImported reports a file has been imported
func (bpr *BundleProgressReporter) ReportFileImported(fileName string, status string) {
	bpr.tracker.IncrementProgress("import")
	
	switch status {
	case "new":
		bpr.incrementStat("New Files")
		bpr.tracker.LogMessage("success", fmt.Sprintf("+ %s (new)", fileName))
	case "updated":
		bpr.incrementStat("Updated Files")
		bpr.tracker.LogMessage("info", fmt.Sprintf("↻ %s (updated)", fileName))
	case "skipped":
		bpr.incrementStat("Skipped Files")
		bpr.tracker.LogMessage("debug", fmt.Sprintf("- %s (skipped)", fileName))
	case "conflict":
		bpr.incrementStat("Conflicts")
		bpr.tracker.LogMessage("warning", fmt.Sprintf("! %s (conflict)", fileName))
	}
}

// ReportCompression reports compression progress
func (bpr *BundleProgressReporter) ReportCompression(originalSize, compressedSize int64) {
	ratio := float64(compressedSize) / float64(originalSize) * 100
	bpr.updateStat("Original Size", formatBytes(originalSize))
	bpr.updateStat("Compressed Size", formatBytes(compressedSize))
	bpr.updateStat("Compression Ratio", fmt.Sprintf("%.1f%%", ratio))
	bpr.tracker.LogMessage("info", fmt.Sprintf("Compressed: %s → %s (%.1f%%)", 
		formatBytes(originalSize), formatBytes(compressedSize), ratio))
}

// ReportEncryption reports encryption status
func (bpr *BundleProgressReporter) ReportEncryption(algorithm string) {
	bpr.updateStat("Encryption", algorithm)
	bpr.tracker.LogMessage("info", fmt.Sprintf("Encrypted with %s", algorithm))
}

// ReportSignature reports signature status
func (bpr *BundleProgressReporter) ReportSignature(keyID string) {
	bpr.updateStat("Signed", "Yes")
	bpr.updateStat("Key ID", keyID)
	bpr.tracker.LogMessage("info", fmt.Sprintf("Signed with key %s", keyID))
}

// Complete marks all operations as complete and shows summary
func (bpr *BundleProgressReporter) Complete() {
	bpr.tracker.CompleteOperation("create", "Bundle creation completed")
	bpr.tracker.CompleteOperation("verify", "Bundle verification completed")
	bpr.tracker.CompleteOperation("import", "Bundle import completed")
	
	bpr.mu.RLock()
	stats := make(map[string]interface{})
	for k, v := range bpr.stats {
		stats[k] = v
	}
	bpr.mu.RUnlock()
	
	bpr.tracker.Summary(stats)
}

// CompleteWithError marks operations as failed
func (bpr *BundleProgressReporter) CompleteWithError(err error) {
	bpr.tracker.FailOperation("create", err)
	bpr.tracker.FailOperation("verify", err)
	bpr.tracker.FailOperation("import", err)
}

// updateStat updates a statistic
func (bpr *BundleProgressReporter) updateStat(key string, value interface{}) {
	bpr.mu.Lock()
	defer bpr.mu.Unlock()
	bpr.stats[key] = value
}

// incrementStat increments a numeric statistic
func (bpr *BundleProgressReporter) incrementStat(key string) {
	bpr.mu.Lock()
	defer bpr.mu.Unlock()
	
	if val, ok := bpr.stats[key].(int); ok {
		bpr.stats[key] = val + 1
	} else {
		bpr.stats[key] = 1
	}
}

// SpinnerProgress provides a simple spinner for indeterminate progress
type SpinnerProgress struct {
	message string
	done    chan bool
	writer  io.Writer
}

// NewSpinnerProgress creates a new spinner progress indicator
func NewSpinnerProgress(message string, writer io.Writer) *SpinnerProgress {
	if writer == nil {
		writer = os.Stdout
	}
	return &SpinnerProgress{
		message: message,
		done:    make(chan bool),
		writer:  writer,
	}
}

// Start starts the spinner
func (sp *SpinnerProgress) Start() {
	go func() {
		spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-sp.done:
				fmt.Fprintf(sp.writer, "\r\033[K")
				return
			default:
				fmt.Fprintf(sp.writer, "\r%s %s", spinChars[i%len(spinChars)], sp.message)
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
}

// Stop stops the spinner
func (sp *SpinnerProgress) Stop() {
	close(sp.done)
}

// StopWithMessage stops the spinner and displays a message
func (sp *SpinnerProgress) StopWithMessage(success bool, message string) {
	sp.Stop()
	if success {
		color.Green("✓ %s", message)
	} else {
		color.Red("✗ %s", message)
	}
}