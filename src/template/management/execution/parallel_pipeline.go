// Package execution provides functionality for executing templates against LLM systems.
package execution

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// PipelineStage represents a stage in the execution pipeline
type PipelineStage int

const (
	// StagePreprocessing is the preprocessing stage
	StagePreprocessing PipelineStage = iota
	// StageExecution is the execution stage
	StageExecution
	// StageDetection is the detection stage
	StageDetection
	// StagePostprocessing is the postprocessing stage
	StagePostprocessing
)

// PipelineExecutor is a template executor that uses a parallel pipeline
type PipelineExecutor struct {
	// baseExecutor is the underlying executor
	baseExecutor interfaces.TemplateExecutor
	// stages is a map of pipeline stages to worker pools
	stages map[PipelineStage]*stageWorkerPool
	// stageBuffers is a map of pipeline stages to buffers
	stageBuffers map[PipelineStage]chan *pipelineTask
	// bufferSize is the size of stage buffers
	bufferSize int
	// preprocessors is a list of template preprocessors
	preprocessors []TemplatePreprocessor
	// postprocessors is a list of result postprocessors
	postprocessors []ResultPostprocessor
	// stats tracks pipeline statistics
	stats pipelineStats
	// statsMutex protects the stats
	statsMutex sync.RWMutex
	// shutdownCh is used to signal shutdown
	shutdownCh chan struct{}
	// isShutdown indicates if the pipeline is shutting down
	isShutdown bool
	// shutdownMutex protects isShutdown
	shutdownMutex sync.RWMutex
}

// pipelineTask represents a task in the execution pipeline
type pipelineTask struct {
	// template is the template to execute
	template *format.Template
	// options are the execution options
	options map[string]interface{}
	// ctx is the execution context
	ctx context.Context
	// result is the execution result
	result *interfaces.TemplateResult
	// err is the execution error
	err error
	// stage is the current pipeline stage
	stage PipelineStage
	// resultCh is the channel for the final result
	resultCh chan *interfaces.TemplateResult
	// errCh is the channel for errors
	errCh chan error
	// startTime is the time the task entered the pipeline
	startTime time.Time
	// stageStartTimes tracks start times for each stage
	stageStartTimes map[PipelineStage]time.Time
	// stageDurations tracks durations for each stage
	stageDurations map[PipelineStage]time.Duration
}

// stageWorkerPool is a pool of workers for a pipeline stage
type stageWorkerPool struct {
	// workers is the number of workers
	workers int
	// inputCh is the input channel
	inputCh chan *pipelineTask
	// outputCh is the output channel
	outputCh chan *pipelineTask
	// processor is the stage processor
	processor stageProcessor
	// wg is a wait group for workers
	wg sync.WaitGroup
	// shutdownCh is used to signal shutdown
	shutdownCh chan struct{}
}

// stageProcessor is a processor for a pipeline stage
type stageProcessor interface {
	// Process processes a task
	Process(task *pipelineTask) error
}

// TemplatePreprocessor is a preprocessor for templates
type TemplatePreprocessor interface {
	// Preprocess preprocesses a template
	Preprocess(ctx context.Context, template *format.Template, options map[string]interface{}) (*format.Template, error)
}

// ResultPostprocessor is a postprocessor for results
type ResultPostprocessor interface {
	// Postprocess postprocesses a result
	Postprocess(ctx context.Context, result *interfaces.TemplateResult) (*interfaces.TemplateResult, error)
}

// pipelineStats tracks pipeline statistics
type pipelineStats struct {
	// TotalTasks is the total number of tasks processed
	TotalTasks int64
	// CompletedTasks is the number of completed tasks
	CompletedTasks int64
	// FailedTasks is the number of failed tasks
	FailedTasks int64
	// StageTasks is the number of tasks in each stage
	StageTasks map[PipelineStage]int64
	// StageFailures is the number of failures in each stage
	StageFailures map[PipelineStage]int64
	// StageDurations is the total duration of each stage
	StageDurations map[PipelineStage]time.Duration
	// TotalDuration is the total pipeline duration
	TotalDuration time.Duration
}

// preprocessingProcessor is a processor for the preprocessing stage
type preprocessingProcessor struct {
	// preprocessors is a list of template preprocessors
	preprocessors []TemplatePreprocessor
}

// executionProcessor is a processor for the execution stage
type executionProcessor struct {
	// executor is the template executor
	executor interfaces.TemplateExecutor
}

// detectionProcessor is a processor for the detection stage
type detectionProcessor struct {
	// detectionEngine is the detection engine
	detectionEngine DetectionEngine
}

// postprocessingProcessor is a processor for the postprocessing stage
type postprocessingProcessor struct {
	// postprocessors is a list of result postprocessors
	postprocessors []ResultPostprocessor
}

// NewPipelineExecutor creates a new pipeline executor
func NewPipelineExecutor(baseExecutor interfaces.TemplateExecutor, bufferSize int) *PipelineExecutor {
	if bufferSize <= 0 {
		bufferSize = 100
	}

	// Create pipeline
	pipeline := &PipelineExecutor{
		baseExecutor:   baseExecutor,
		stages:         make(map[PipelineStage]*stageWorkerPool),
		stageBuffers:   make(map[PipelineStage]chan *pipelineTask),
		bufferSize:     bufferSize,
		preprocessors:  make([]TemplatePreprocessor, 0),
		postprocessors: make([]ResultPostprocessor, 0),
		stats: pipelineStats{
			StageTasks:     make(map[PipelineStage]int64),
			StageFailures:  make(map[PipelineStage]int64),
			StageDurations: make(map[PipelineStage]time.Duration),
		},
		shutdownCh: make(chan struct{}),
	}

	// Create stage buffers
	pipeline.stageBuffers[StagePreprocessing] = make(chan *pipelineTask, bufferSize)
	pipeline.stageBuffers[StageExecution] = make(chan *pipelineTask, bufferSize)
	pipeline.stageBuffers[StageDetection] = make(chan *pipelineTask, bufferSize)
	pipeline.stageBuffers[StagePostprocessing] = make(chan *pipelineTask, bufferSize)

	// Create stage worker pools
	pipeline.stages[StagePreprocessing] = newStageWorkerPool(4, pipeline.stageBuffers[StagePreprocessing], pipeline.stageBuffers[StageExecution], &preprocessingProcessor{pipeline.preprocessors})
	pipeline.stages[StageExecution] = newStageWorkerPool(8, pipeline.stageBuffers[StageExecution], pipeline.stageBuffers[StageDetection], &executionProcessor{baseExecutor})
	pipeline.stages[StageDetection] = newStageWorkerPool(4, pipeline.stageBuffers[StageDetection], pipeline.stageBuffers[StagePostprocessing], &detectionProcessor{nil})
	pipeline.stages[StagePostprocessing] = newStageWorkerPool(2, pipeline.stageBuffers[StagePostprocessing], nil, &postprocessingProcessor{pipeline.postprocessors})

	// Start pipeline
	pipeline.startPipeline()

	return pipeline
}

// newStageWorkerPool creates a new stage worker pool
func newStageWorkerPool(workers int, inputCh, outputCh chan *pipelineTask, processor stageProcessor) *stageWorkerPool {
	pool := &stageWorkerPool{
		workers:    workers,
		inputCh:    inputCh,
		outputCh:   outputCh,
		processor:  processor,
		shutdownCh: make(chan struct{}),
	}

	// Start workers
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker is a worker goroutine for a pipeline stage
func (p *stageWorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.inputCh:
			if !ok {
				return
			}

			// Process task
			err := p.processor.Process(task)
			if err != nil {
				// Task failed
				task.err = err
				// Send error to error channel
				select {
				case task.errCh <- err:
				default:
				}
			} else if p.outputCh != nil {
				// Send task to next stage
				select {
				case p.outputCh <- task:
				case <-p.shutdownCh:
					return
				}
			} else {
				// Final stage, send result to result channel
				select {
				case task.resultCh <- task.result:
				case <-p.shutdownCh:
					return
				}
			}

		case <-p.shutdownCh:
			return
		}
	}
}

// startPipeline starts the pipeline
func (p *PipelineExecutor) startPipeline() {
	// Start result collector
	go p.collectResults()
}

// collectResults collects results from the final stage
func (p *PipelineExecutor) collectResults() {
	for {
		select {
		case task := <-p.stageBuffers[StagePostprocessing]:
			// Record stage duration
			task.stageDurations[StagePostprocessing] = time.Since(task.stageStartTimes[StagePostprocessing])

			// Update stats
			p.statsMutex.Lock()
			p.stats.CompletedTasks++
			p.stats.StageDurations[StagePostprocessing] += task.stageDurations[StagePostprocessing]
			p.stats.TotalDuration += time.Since(task.startTime)
			p.statsMutex.Unlock()

			// Send result to result channel
			task.resultCh <- task.result
			close(task.resultCh)
			close(task.errCh)

		case <-p.shutdownCh:
			return
		}
	}
}

// Process processes a task in the preprocessing stage
func (p *preprocessingProcessor) Process(task *pipelineTask) error {
	// Record stage start time
	task.stageStartTimes[StagePreprocessing] = time.Now()

	// Create result
	task.result = &interfaces.TemplateResult{
		TemplateID: task.template.ID,
		Template:   task.template,
		StartTime:  time.Now(),
		Status:     string(interfaces.StatusExecuting),
	}

	// Apply preprocessors
	processedTemplate := task.template
	var err error

	for _, preprocessor := range p.preprocessors {
		processedTemplate, err = preprocessor.Preprocess(task.ctx, processedTemplate, task.options)
		if err != nil {
			task.result.Status = string(interfaces.StatusFailed)
			task.result.Error = err
			task.result.EndTime = time.Now()
			task.result.Duration = task.result.EndTime.Sub(task.result.StartTime)
			return err
		}
	}

	// Update template
	task.template = processedTemplate

	// Update stage
	task.stage = StageExecution

	// Record stage duration
	task.stageDurations[StagePreprocessing] = time.Since(task.stageStartTimes[StagePreprocessing])

	return nil
}

// Process processes a task in the execution stage
func (p *executionProcessor) Process(task *pipelineTask) error {
	// Record stage start time
	task.stageStartTimes[StageExecution] = time.Now()

	// Execute template
	result, err := p.executor.Execute(task.ctx, task.template, task.options)
	if err != nil {
		task.result.Status = string(interfaces.StatusFailed)
		task.result.Error = err
		task.result.EndTime = time.Now()
		task.result.Duration = task.result.EndTime.Sub(task.result.StartTime)
		return err
	}

	// Update result
	task.result = result

	// Update stage
	task.stage = StageDetection

	// Record stage duration
	task.stageDurations[StageExecution] = time.Since(task.stageStartTimes[StageExecution])

	return nil
}

// Process processes a task in the detection stage
func (p *detectionProcessor) Process(task *pipelineTask) error {
	// Record stage start time
	task.stageStartTimes[StageDetection] = time.Now()

	// Skip if no detection engine
	if p.detectionEngine == nil {
		task.stage = StagePostprocessing
		task.stageDurations[StageDetection] = time.Since(task.stageStartTimes[StageDetection])
		return nil
	}

	// Detect vulnerabilities
	detected, score, details, err := p.detectionEngine.Detect(task.ctx, task.template, task.result.Response)
	if err != nil {
		task.result.Status = string(interfaces.StatusFailed)
		task.result.Error = fmt.Errorf("detection failed: %w", err)
		task.result.EndTime = time.Now()
		task.result.Duration = task.result.EndTime.Sub(task.result.StartTime)
		return err
	}

	// Update result
	task.result.Detected = detected
	task.result.Score = score
	task.result.Details = details

	// Update stage
	task.stage = StagePostprocessing

	// Record stage duration
	task.stageDurations[StageDetection] = time.Since(task.stageStartTimes[StageDetection])

	return nil
}

// Process processes a task in the postprocessing stage
func (p *postprocessingProcessor) Process(task *pipelineTask) error {
	// Record stage start time
	task.stageStartTimes[StagePostprocessing] = time.Now()

	// Apply postprocessors
	processedResult := task.result
	var err error

	for _, postprocessor := range p.postprocessors {
		processedResult, err = postprocessor.Postprocess(task.ctx, processedResult)
		if err != nil {
			task.result.Status = string(interfaces.StatusFailed)
			task.result.Error = err
			task.result.EndTime = time.Now()
			task.result.Duration = task.result.EndTime.Sub(task.result.StartTime)
			return err
		}
	}

	// Update result
	task.result = processedResult
	task.result.Status = string(interfaces.StatusCompleted)
	task.result.EndTime = time.Now()
	task.result.Duration = task.result.EndTime.Sub(task.result.StartTime)

	return nil
}

// Execute executes a template through the pipeline
func (p *PipelineExecutor) Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*interfaces.TemplateResult, error) {
	// Check if shutting down
	if p.isShuttingDown() {
		return nil, fmt.Errorf("pipeline is shutting down")
	}

	// Update stats
	p.statsMutex.Lock()
	p.stats.TotalTasks++
	p.stats.StageTasks[StagePreprocessing]++
	p.statsMutex.Unlock()

	// Create channels
	resultCh := make(chan *interfaces.TemplateResult, 1)
	errCh := make(chan error, 1)

	// Create task
	task := &pipelineTask{
		template:        template,
		options:         options,
		ctx:             ctx,
		stage:           StagePreprocessing,
		resultCh:        resultCh,
		errCh:           errCh,
		startTime:       time.Now(),
		stageStartTimes: make(map[PipelineStage]time.Time),
		stageDurations:  make(map[PipelineStage]time.Duration),
	}

	// Submit task to pipeline
	select {
	case p.stageBuffers[StagePreprocessing] <- task:
		// Task submitted
	case <-ctx.Done():
		// Context cancelled
		p.statsMutex.Lock()
		p.stats.FailedTasks++
		p.statsMutex.Unlock()
		return nil, ctx.Err()
	}

	// Wait for result or error
	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		p.statsMutex.Lock()
		p.stats.FailedTasks++
		p.stats.StageFailures[task.stage]++
		p.statsMutex.Unlock()
		return nil, err
	case <-ctx.Done():
		p.statsMutex.Lock()
		p.stats.FailedTasks++
		p.statsMutex.Unlock()
		return nil, ctx.Err()
	}
}

// ExecuteBatch executes multiple templates through the pipeline
func (p *PipelineExecutor) ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*interfaces.TemplateResult, error) {
	// Check if shutting down
	if p.isShuttingDown() {
		return nil, fmt.Errorf("pipeline is shutting down")
	}

	// Create results slice
	results := make([]*interfaces.TemplateResult, len(templates))

	// Create wait group
	var wg sync.WaitGroup
	wg.Add(len(templates))

	// Create error channel
	errorCh := make(chan error, len(templates))

	// Execute templates concurrently
	for i, template := range templates {
		i, template := i, template // Create local variables for closure

		// Execute template asynchronously
		go func() {
			defer wg.Done()

			// Execute template
			result, err := p.Execute(ctx, template, options)
			if err != nil {
				errorCh <- err
				results[i] = &interfaces.TemplateResult{
					TemplateID: template.ID,
					Template:   template,
					Status:     string(interfaces.StatusFailed),
					Error:      err,
					EndTime:    time.Now(),
				}
				return
			}

			// Store result
			results[i] = result
		}()
	}

	// Wait for all executions to complete
	wg.Wait()
	close(errorCh)

	// Check for errors
	var lastError error
	for err := range errorCh {
		lastError = err
	}

	return results, lastError
}

// AddPreprocessor adds a preprocessor to the pipeline
func (p *PipelineExecutor) AddPreprocessor(preprocessor TemplatePreprocessor) {
	p.preprocessors = append(p.preprocessors, preprocessor)
	// Update preprocessor in the preprocessing stage
	p.stages[StagePreprocessing].processor = &preprocessingProcessor{p.preprocessors}
}

// AddPostprocessor adds a postprocessor to the pipeline
func (p *PipelineExecutor) AddPostprocessor(postprocessor ResultPostprocessor) {
	p.postprocessors = append(p.postprocessors, postprocessor)
	// Update postprocessor in the postprocessing stage
	p.stages[StagePostprocessing].processor = &postprocessingProcessor{p.postprocessors}
}

// SetDetectionEngine sets the detection engine for the pipeline
func (p *PipelineExecutor) SetDetectionEngine(detectionEngine DetectionEngine) {
	// Update detection engine in the detection stage
	p.stages[StageDetection].processor = &detectionProcessor{detectionEngine}
}

// isShuttingDown checks if the pipeline is shutting down
func (p *PipelineExecutor) isShuttingDown() bool {
	p.shutdownMutex.RLock()
	defer p.shutdownMutex.RUnlock()
	return p.isShutdown
}

// Shutdown shuts down the pipeline
func (p *PipelineExecutor) Shutdown() {
	// Set shutdown flag
	p.shutdownMutex.Lock()
	if p.isShutdown {
		p.shutdownMutex.Unlock()
		return
	}
	p.isShutdown = true
	p.shutdownMutex.Unlock()

	// Signal shutdown
	close(p.shutdownCh)

	// Close stage worker pools
	for _, pool := range p.stages {
		close(pool.shutdownCh)
	}

	// Wait for workers to finish
	for _, pool := range p.stages {
		pool.wg.Wait()
	}
}

// GetPipelineStats returns statistics about the pipeline
func (p *PipelineExecutor) GetPipelineStats() map[string]interface{} {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()

	// Calculate average durations
	avgStageDurations := make(map[string]time.Duration)
	for stage, duration := range p.stats.StageDurations {
		tasks := p.stats.StageTasks[stage]
		if tasks > 0 {
			avgStageDurations[fmt.Sprintf("stage_%d", stage)] = time.Duration(int64(duration) / tasks)
		}
	}

	avgDuration := time.Duration(0)
	if p.stats.CompletedTasks > 0 {
		avgDuration = time.Duration(int64(p.stats.TotalDuration) / p.stats.CompletedTasks)
	}

	return map[string]interface{}{
		"total_tasks":      p.stats.TotalTasks,
		"completed_tasks":  p.stats.CompletedTasks,
		"failed_tasks":     p.stats.FailedTasks,
		"stage_tasks":      p.stats.StageTasks,
		"stage_failures":   p.stats.StageFailures,
		"avg_duration":     avgDuration,
		"avg_stage_durations": avgStageDurations,
	}
}
