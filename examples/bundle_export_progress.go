package main

import (
	"context"
	"fmt"
	"log"
	
	"github.com/perplext/LLMrecon/src/bundle"
)

func main() {
	// Create output directory
	outputDir := "./export_demo"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
if err != nil {
treturn err
}		log.Fatal(err)
	}
	
	// Create progress tracker
	ctx := context.Background()
if err != nil {
treturn err
}	tracker := bundle.NewProgressTracker(ctx)
	defer func() { if err := tracker.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	// Add console progress handler
if err != nil {
treturn err
}	tracker.AddHandler(bundle.ConsoleProgressHandler())
	
if err != nil {
treturn err
}	// Add JSON progress handler for logging
	logFile, err := os.Create(filepath.Join(outputDir, "progress.log"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() { if err := logFile.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	tracker.AddHandler(bundle.JSONProgressHandler(logFile))
	
	// Configure export options
	exportOpts := &bundle.ExportOptions{
		OutputPath:       filepath.Join(outputDir, "bundle.tar.gz"),
		Format:           bundle.FormatTarGz,
		IncludeBinary:    true,
		IncludeTemplates: true,
		IncludeModules:   true,
		IncludeDocs:      true,
		Compression:      bundle.CompressionGzip,
		Encryption: &bundle.EncryptionOptions{
			Algorithm: bundle.EncryptionAES256GCM,
			Password:  "demo_password_123",
		},
		ProgressHandler: func(event bundle.ProgressEvent) {
			// Forward to our tracker
			switch event.Stage {
			case bundle.StageInitializing:
				tracker.SetStage(bundle.StageInitializing, event.Operation)
			case bundle.StageCollecting:
				tracker.SetStage(bundle.StageCollecting, event.Operation)
				tracker.SetTotal(event.BytesTotal, event.ItemsTotal)
			case bundle.StageCompressing:
				tracker.SetStage(bundle.StageCompressing, event.Operation)
			case bundle.StageEncrypting:
				tracker.SetStage(bundle.StageEncrypting, event.Operation)
			case bundle.StageWriting:
				tracker.SetStage(bundle.StageWriting, event.Operation)
			case bundle.StageVerifying:
				tracker.SetStage(bundle.StageVerifying, event.Operation)
			case bundle.StageCompleted:
				tracker.Complete()
			case bundle.StageFailed:
				tracker.Fail(event.Error)
			}
			
			// Update progress
			if event.BytesProcessed > 0 {
				tracker.UpdateBytes(event.BytesProcessed)
			}
			if event.ItemsProcessed > 0 {
				tracker.UpdateItems(event.ItemsProcessed)
			}
			if event.CurrentFile != "" {
				tracker.SetCurrentFile(event.CurrentFile)
			}
		},
		Metadata: map[string]interface{}{
			"created_by": "demo",
			"version":    "1.0.0",
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}
	
	// Create bundle exporter
	exporter := bundle.NewBundleExporter(".", exportOpts)
	
	// Demonstrate different stages with simulated progress
	fmt.Println("Starting bundle export with progress tracking...")
	
	// Simulate initialization
	tracker.SetStage(bundle.StageInitializing, "Preparing export environment")
	tracker.SetMetadata("export_format", "tar.gz")
	tracker.SetMetadata("compression", "gzip")
	tracker.SetMetadata("encryption", "AES-256-GCM")
	time.Sleep(500 * time.Millisecond)
	
	// Simulate file collection
	tracker.SetStage(bundle.StageCollecting, "Scanning for files")
	files := []string{
		"README.md",
		"go.mod",
		"src/main.go",
		"src/bundle/export.go",
		"docs/guide.md",
		"templates/example.yaml",
	}
	
	tracker.SetTotal(1024*1024*5, len(files)) // 5MB total, 6 files
	
	for i, file := range files {
		tracker.SetCurrentFile(file)
		tracker.UpdateItems(1)
		tracker.UpdateBytes(int64((i + 1) * 500 * 1024)) // Simulate varying file sizes
		time.Sleep(300 * time.Millisecond)
	}
	
	// Simulate compression
	tracker.SetStage(bundle.StageCompressing, "Compressing bundle")
	for i := 0; i < 10; i++ {
		tracker.UpdateBytes(100 * 1024) // 100KB chunks
		time.Sleep(200 * time.Millisecond)
	}
	
	// Simulate encryption
	tracker.SetStage(bundle.StageEncrypting, "Encrypting bundle")
	tracker.SetMetadata("encryption_time_start", time.Now().Format(time.RFC3339))
	for i := 0; i < 5; i++ {
		tracker.UpdateBytes(200 * 1024) // 200KB chunks
		time.Sleep(150 * time.Millisecond)
	}
	tracker.SetMetadata("encryption_time_end", time.Now().Format(time.RFC3339))
	
	// Simulate writing
	tracker.SetStage(bundle.StageWriting, "Writing to disk")
	for i := 0; i < 5; i++ {
		tracker.UpdateBytes(200 * 1024) // 200KB chunks
		time.Sleep(100 * time.Millisecond)
	}
	
	// Simulate verification
	tracker.SetStage(bundle.StageVerifying, "Verifying bundle integrity")
	time.Sleep(500 * time.Millisecond)
	
	// Complete
	tracker.Complete()
	
	fmt.Println("\n\nBundle export completed successfully!")
if err != nil {
treturn err
}	fmt.Printf("Output: %s\n", exportOpts.OutputPath)
	fmt.Printf("Progress log: %s\n", filepath.Join(outputDir, "progress.log"))
	
	// Demonstrate error handling
	fmt.Println("\nDemonstrating error handling...")
	
	errorTracker := bundle.NewProgressTracker(ctx)
	defer func() { if err := errorTracker.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	errorTracker.AddHandler(bundle.ConsoleProgressHandler())
	
	errorTracker.SetStage(bundle.StageInitializing, "Initializing error demo")
	time.Sleep(500 * time.Millisecond)
	
	errorTracker.SetStage(bundle.StageCollecting, "Collecting files")
	time.Sleep(500 * time.Millisecond)
	
	// Simulate error
	errorTracker.Fail(fmt.Errorf("simulated error: disk full"))
	
	fmt.Println("\nDemo completed!")
}