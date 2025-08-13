// Package streaming provides functionality for streaming processing of templates.
package streaming

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// StreamProcessor processes templates in a streaming fashion to minimize memory usage
type StreamProcessor struct {
	// config is the processor configuration
	config *StreamConfig
	// stats tracks processor statistics
	stats StreamStats
	// statsMutex protects the stats
	statsMutex sync.RWMutex
}

// StreamConfig contains configuration for the stream processor
type StreamConfig struct {
	// BufferSize is the size of the buffer for reading files
	BufferSize int
	// ChunkSize is the size of chunks for processing
	ChunkSize int
	// MaxConcurrent is the maximum number of concurrent operations
	MaxConcurrent int
	// TemporaryDir is the directory for temporary files
	TemporaryDir string
}

// StreamStats tracks processor statistics
type StreamStats struct {
	// ProcessedBytes is the number of bytes processed
	ProcessedBytes int64
	// ProcessedTemplates is the number of templates processed
	ProcessedTemplates int64
	// ProcessingTime is the total time spent processing
	ProcessingTime time.Duration
	// PeakMemoryUsage is the peak memory usage in bytes
	PeakMemoryUsage int64
}

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(config *StreamConfig) *StreamProcessor {
	// Set default values
	if config == nil {
		config = &StreamConfig{
			BufferSize:     4096,
			ChunkSize:      1024 * 1024, // 1MB
			MaxConcurrent:  10,
			TemporaryDir:   os.TempDir(),
		}
	}

	return &StreamProcessor{
		config: config,
	}
}

// ProcessTemplateFile processes a template file in a streaming fashion
func (p *StreamProcessor) ProcessTemplateFile(ctx context.Context, filePath string, processor func(*format.Template) error) error {
	startTime := time.Now()

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create buffered reader
	reader := bufio.NewReaderSize(file, p.config.BufferSize)

	// Read file in chunks
	var processedBytes int64
	buffer := make([]byte, p.config.ChunkSize)
	templateData := make([]byte, 0, p.config.ChunkSize)

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		// Read chunk
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if n > 0 {
			// Append chunk to template data
			templateData = append(templateData, buffer[:n]...)
			processedBytes += int64(n)
		}

		// If we've reached EOF or have enough data to parse a template
		if err == io.EOF || len(templateData) >= p.config.ChunkSize {
			// Try to parse template
			template, err := format.ParseTemplate(templateData)
			if err != nil {
				// If we've reached EOF and still can't parse, return error
				if err == io.EOF {
					return fmt.Errorf("failed to parse template: %w", err)
				}
				
				// If we haven't reached EOF, continue reading
				continue
			}

			// Process template
			if err := processor(template); err != nil {
				return fmt.Errorf("failed to process template: %w", err)
			}

			// Update stats
			p.statsMutex.Lock()
			p.stats.ProcessedBytes += processedBytes
			p.stats.ProcessedTemplates++
			p.statsMutex.Unlock()

			// Reset template data
			templateData = make([]byte, 0, p.config.ChunkSize)
		}

		// If we've reached EOF, break
		if err == io.EOF {
			break
		}
	}

	// Update stats
	p.statsMutex.Lock()
	p.stats.ProcessingTime += time.Since(startTime)
	p.statsMutex.Unlock()

	return nil
}

// ProcessTemplateStream processes a template stream
func (p *StreamProcessor) ProcessTemplateStream(ctx context.Context, reader io.Reader, processor func(*format.Template) error) error {
	startTime := time.Now()

	// Create buffered reader
	bufferedReader := bufio.NewReaderSize(reader, p.config.BufferSize)

	// Read stream in chunks
	var processedBytes int64
	buffer := make([]byte, p.config.ChunkSize)
	templateData := make([]byte, 0, p.config.ChunkSize)

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		// Read chunk
		n, err := bufferedReader.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read stream: %w", err)
		}

		if n > 0 {
			// Append chunk to template data
			templateData = append(templateData, buffer[:n]...)
			processedBytes += int64(n)
		}

		// If we've reached EOF or have enough data to parse a template
		if err == io.EOF || len(templateData) >= p.config.ChunkSize {
			// Try to parse template
			template, err := format.ParseTemplate(templateData)
			if err != nil {
				// If we've reached EOF and still can't parse, return error
				if err == io.EOF {
					return fmt.Errorf("failed to parse template: %w", err)
				}
				
				// If we haven't reached EOF, continue reading
				continue
			}

			// Process template
			if err := processor(template); err != nil {
				return fmt.Errorf("failed to process template: %w", err)
			}

			// Update stats
			p.statsMutex.Lock()
			p.stats.ProcessedBytes += processedBytes
			p.stats.ProcessedTemplates++
			p.statsMutex.Unlock()

			// Reset template data
			templateData = make([]byte, 0, p.config.ChunkSize)
		}

		// If we've reached EOF, break
		if err == io.EOF {
			break
		}
	}

	// Update stats
	p.statsMutex.Lock()
	p.stats.ProcessingTime += time.Since(startTime)
	p.statsMutex.Unlock()

	return nil
}

// ProcessTemplateDirectory processes all template files in a directory
func (p *StreamProcessor) ProcessTemplateDirectory(ctx context.Context, dirPath string, processor func(*format.Template) error) error {
	startTime := time.Now()

	// Open directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return fmt.Errorf("failed to open directory: %w", err)
	}
	defer dir.Close()

	// Read directory entries
	entries, err := dir.Readdir(-1)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Process files concurrently with limit
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, p.config.MaxConcurrent)
	errorChan := make(chan error, len(entries))

	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Skip non-template files
		if !isTemplateFile(entry.Name()) {
			continue
		}

		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Process file
			if err := p.ProcessTemplateFile(ctx, filePath, processor); err != nil {
				errorChan <- fmt.Errorf("failed to process file %s: %w", filePath, err)
			}
		}(fmt.Sprintf("%s/%s", dirPath, entry.Name()))
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errorChan)

	// Check for errors
	var lastError error
	for err := range errorChan {
		lastError = err
	}

	// Update stats
	p.statsMutex.Lock()
	p.stats.ProcessingTime += time.Since(startTime)
	p.statsMutex.Unlock()

	return lastError
}

// StreamExecute executes a template in a streaming fashion
func (p *StreamProcessor) StreamExecute(ctx context.Context, template *format.Template, executor interfaces.TemplateExecutor, options map[string]interface{}, resultWriter io.Writer) error {
	startTime := time.Now()

	// Execute template
	result, err := executor.Execute(ctx, template, options)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write result to writer
	if err := json.NewEncoder(resultWriter).Encode(result); err != nil {
		return fmt.Errorf("failed to write result: %w", err)
	}

	// Update stats
	p.statsMutex.Lock()
	p.stats.ProcessedTemplates++
	p.stats.ProcessingTime += time.Since(startTime)
	p.statsMutex.Unlock()

	return nil
}

// StreamExecuteBatch executes multiple templates in a streaming fashion
func (p *StreamProcessor) StreamExecuteBatch(ctx context.Context, templates []*format.Template, executor interfaces.TemplateExecutor, options map[string]interface{}, resultWriter io.Writer) error {
	startTime := time.Now()

	// Create temporary file for results
	tempFile, err := os.CreateTemp(p.config.TemporaryDir, "template-results-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Execute templates in batches
	batchSize := 10
	for i := 0; i < len(templates); i += batchSize {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		// Get batch
		end := i + batchSize
		if end > len(templates) {
			end = len(templates)
		}
		batch := templates[i:end]

		// Execute batch
		results, err := executor.ExecuteBatch(ctx, batch, options)
		if err != nil {
			return fmt.Errorf("failed to execute batch: %w", err)
		}

		// Write results to temporary file
		for _, result := range results {
			if err := json.NewEncoder(tempFile).Encode(result); err != nil {
				return fmt.Errorf("failed to write result: %w", err)
			}
		}

		// Update stats
		p.statsMutex.Lock()
		p.stats.ProcessedTemplates += int64(len(batch))
		p.statsMutex.Unlock()
	}

	// Reset file position
	if _, err := tempFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek temporary file: %w", err)
	}

	// Copy results to writer
	if _, err := io.Copy(resultWriter, tempFile); err != nil {
		return fmt.Errorf("failed to copy results: %w", err)
	}

	// Update stats
	p.statsMutex.Lock()
	p.stats.ProcessingTime += time.Since(startTime)
	p.statsMutex.Unlock()

	return nil
}

// GetStats gets the processor statistics
func (p *StreamProcessor) GetStats() StreamStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()

	return p.stats
}

// ResetStats resets the processor statistics
func (p *StreamProcessor) ResetStats() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()

	p.stats = StreamStats{}
}

// isTemplateFile checks if a file is a template file
func isTemplateFile(fileName string) bool {
	ext := ""
	for i := len(fileName) - 1; i >= 0 && fileName[i] != '.'; i-- {
		ext = string(fileName[i]) + ext
	}
	
	return ext == "yaml" || ext == "yml" || ext == "json"
}
