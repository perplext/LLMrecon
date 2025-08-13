// Package execution provides functionality for executing templates against LLM systems.
package execution

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// AsyncTemplateExecutor is a high-performance template executor with advanced concurrency
type AsyncTemplateExecutor struct {
	// baseExecutor is the underlying executor
	baseExecutor *OptimizedTemplateExecutor
	// taskQueue is the queue of execution tasks
	taskQueue chan *executionTask
	// resultCache is a cache for execution results
	resultCache *resultCache
	// workerPool manages worker goroutines
	workerPool *advancedWorkPool
	// priorityManager manages execution priorities
	priorityManager *priorityManager
	// stats tracks execution statistics
	stats asyncExecutionStats
	// statsMutex protects the stats
	statsMutex sync.RWMutex
	// shutdownCh is used to signal shutdown
	shutdownCh chan struct{}
	// isShutdown indicates if the executor is shutting down
	isShutdown bool
	// shutdownMutex protects isShutdown
	shutdownMutex sync.RWMutex
}

// executionTask represents a template execution task
type executionTask struct {
	// template is the template to execute
	template *format.Template
	// options are the execution options
	options map[string]interface{}
	// ctx is the execution context
	ctx context.Context
	// resultCh is the channel for the execution result
	resultCh chan *taskResult
	// priority is the execution priority (higher values = higher priority)
	priority int
	// createdAt is the time the task was created
	createdAt time.Time
}

// taskResult represents the result of a task execution
type taskResult struct {
	// result is the template execution result
	result *interfaces.TemplateResult
	// err is the execution error
	err error
}

// resultCache is a cache for execution results
type resultCache struct {
	// cache is a map of cache key to result
	cache map[string]*interfaces.TemplateResult
	// mutex protects the cache
	mutex sync.RWMutex
	// maxSize is the maximum size of the cache
	maxSize int
	// ttl is the time-to-live for cache entries
	ttl time.Duration
}

// asyncExecutionStats tracks execution statistics
type asyncExecutionStats struct {
	// TotalTasks is the total number of tasks submitted
	TotalTasks int64
	// CompletedTasks is the number of completed tasks
	CompletedTasks int64
	// FailedTasks is the number of failed tasks
	FailedTasks int64
	// CachedResults is the number of cached results used
	CachedResults int64
	// TotalExecutionTime is the total execution time
	TotalExecutionTime time.Duration
	// AverageQueueTime is the average time tasks spend in the queue
	AverageQueueTime time.Duration
	// QueuedTasks is the current number of queued tasks
	QueuedTasks int64
}

// advancedWorkPool is an advanced worker pool with dynamic scaling
type advancedWorkPool struct {
	// workers is the number of workers
	workers int
	// minWorkers is the minimum number of workers
	minWorkers int
	// maxWorkers is the maximum number of workers
	maxWorkers int
	// activeWorkers is the current number of active workers
	activeWorkers int
	// mutex protects the worker count
	mutex sync.RWMutex
	// taskCh is the channel for tasks
	taskCh chan *executionTask
	// workerWg is a wait group for workers
	workerWg sync.WaitGroup
	// shutdownCh is used to signal shutdown
	shutdownCh chan struct{}
}

// priorityManager manages execution priorities
type priorityManager struct {
	// priorityQueues is a map of priority to task queue
	priorityQueues map[int]chan *executionTask
	// mutex protects the priority queues
	mutex sync.RWMutex
	// defaultPriority is the default priority
	defaultPriority int
}

// NewAsyncTemplateExecutor creates a new async template executor
func NewAsyncTemplateExecutor(baseExecutor *OptimizedTemplateExecutor, queueSize int, cacheSize int, cacheTTL time.Duration) *AsyncTemplateExecutor {
	if queueSize <= 0 {
		queueSize = 1000
	}
	if cacheSize <= 0 {
		cacheSize = 1000
	}
	if cacheTTL == 0 {
		cacheTTL = 1 * time.Hour
	}

	// Create result cache
	resultCache := &resultCache{
		cache:   make(map[string]*interfaces.TemplateResult),
		maxSize: cacheSize,
		ttl:     cacheTTL,
	}

	// Create priority manager
	priorityManager := &priorityManager{
		priorityQueues:  make(map[int]chan *executionTask),
		defaultPriority: 5,
	}

	// Create task queue
	taskQueue := make(chan *executionTask, queueSize)

	// Create worker pool
	workerPool := newAdvancedWorkPool(runtime.NumCPU()*2, runtime.NumCPU(), runtime.NumCPU()*4, taskQueue)

	// Create executor
	executor := &AsyncTemplateExecutor{
		baseExecutor:    baseExecutor,
		taskQueue:       taskQueue,
		resultCache:     resultCache,
		workerPool:      workerPool,
		priorityManager: priorityManager,
		shutdownCh:      make(chan struct{}),
	}

	// Start dispatcher
	go executor.dispatchTasks()

	return executor
}

// newAdvancedWorkPool creates a new advanced worker pool
func newAdvancedWorkPool(workers, minWorkers, maxWorkers int, taskCh chan *executionTask) *advancedWorkPool {
	pool := &advancedWorkPool{
		workers:    workers,
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
		taskCh:     taskCh,
		shutdownCh: make(chan struct{}),
	}

	// Start workers
	pool.startWorkers(workers)

	// Start worker manager
	go pool.manageWorkers()

	return pool
}

// startWorkers starts a number of workers
func (p *advancedWorkPool) startWorkers(count int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i := 0; i < count; i++ {
		p.workerWg.Add(1)
		p.activeWorkers++
		go p.worker()
	}
}

// worker is a worker goroutine
func (p *advancedWorkPool) worker() {
	defer p.workerWg.Done()

	for {
		select {
		case task, ok := <-p.taskCh:
			if !ok {
				return
			}

			// Execute task
			result, err := task.execute()

			// Send result
			task.resultCh <- &taskResult{
				result: result,
				err:    err,
			}
			close(task.resultCh)

		case <-p.shutdownCh:
			return
		}
	}
}

// manageWorkers manages the worker pool size
func (p *advancedWorkPool) manageWorkers() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.adjustWorkerCount()
		case <-p.shutdownCh:
			return
		}
	}
}

// adjustWorkerCount adjusts the worker count based on load
func (p *advancedWorkPool) adjustWorkerCount() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Get queue length
	queueLength := len(p.taskCh)

	// Calculate target worker count
	_ = p.activeWorkers // targetWorkers not used

	// Increase workers if queue is filling up
	if queueLength > p.activeWorkers*2 && p.activeWorkers < p.maxWorkers {
		// Add workers (up to 25% more, but at least 1)
		addCount := max(1, p.activeWorkers/4)
		// Don't exceed max workers
		addCount = min(addCount, p.maxWorkers-p.activeWorkers)
		
		for i := 0; i < addCount; i++ {
			p.workerWg.Add(1)
			p.activeWorkers++
			go p.worker()
		}
	}

	// Decrease workers if queue is empty and we have more than minimum
	if queueLength == 0 && p.activeWorkers > p.minWorkers {
		// Remove workers gradually (up to 10% fewer, but at least 1)
		removeCount := max(1, p.activeWorkers/10)
		// Don't go below min workers
		removeCount = min(removeCount, p.activeWorkers-p.minWorkers)
		
		// Signal workers to stop
		for i := 0; i < removeCount; i++ {
			select {
			case p.shutdownCh <- struct{}{}:
				p.activeWorkers--
			default:
				// If channel is full, stop removing workers
				break
			}
		}
	}
}

// execute executes the task
func (t *executionTask) execute() (*interfaces.TemplateResult, error) {
	// Create result
	result := &interfaces.TemplateResult{
		TemplateID: t.template.ID,
		Template:   t.template,
		StartTime:  time.Now(),
		Status:     string(interfaces.StatusExecuting),
	}

	// Execute template
	baseResult, err := t.executeTemplate()
	if err != nil {
		result.Status = string(interfaces.StatusFailed)
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Copy fields from base result
	result.Response = baseResult.Response
	result.Detected = baseResult.Detected
	result.Score = baseResult.Score
	result.Details = baseResult.Details
	result.Status = string(interfaces.StatusCompleted)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// executeTemplate executes the template using the base executor
func (t *executionTask) executeTemplate() (*interfaces.TemplateResult, error) {
	// Get base executor from context
	baseExecutor, ok := t.ctx.Value("baseExecutor").(*OptimizedTemplateExecutor)
	if !ok {
		return nil, fmt.Errorf("base executor not found in context")
	}

	// Execute template
	return baseExecutor.Execute(t.ctx, t.template, t.options)
}

// dispatchTasks dispatches tasks to workers
func (e *AsyncTemplateExecutor) dispatchTasks() {
	for {
		select {
		case task := <-e.taskQueue:
			// Add base executor to context
			ctx := context.WithValue(task.ctx, "baseExecutor", e.baseExecutor)
			task.ctx = ctx

			// Send task to worker pool
			e.workerPool.taskCh <- task

		case <-e.shutdownCh:
			return
		}
	}
}

// Execute executes a template asynchronously
func (e *AsyncTemplateExecutor) Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*interfaces.TemplateResult, error) {
	// Check if shutting down
	if e.isShuttingDown() {
		return nil, fmt.Errorf("executor is shutting down")
	}

	// Update stats
	e.statsMutex.Lock()
	e.stats.TotalTasks++
	e.stats.QueuedTasks++
	e.statsMutex.Unlock()

	// Create result channel
	resultCh := make(chan *taskResult, 1)

	// Create task
	task := &executionTask{
		template:  template,
		options:   options,
		ctx:       ctx,
		resultCh:  resultCh,
		priority:  e.getPriority(options),
		createdAt: time.Now(),
	}

	// Submit task
	select {
	case e.taskQueue <- task:
		// Task submitted
	case <-ctx.Done():
		// Context cancelled
		e.statsMutex.Lock()
		e.stats.QueuedTasks--
		e.stats.FailedTasks++
		e.statsMutex.Unlock()
		return nil, ctx.Err()
	}

	// Wait for result
	select {
	case result := <-resultCh:
		// Update stats
		e.statsMutex.Lock()
		e.stats.QueuedTasks--
		if result.err != nil {
			e.stats.FailedTasks++
		} else {
			e.stats.CompletedTasks++
			e.stats.TotalExecutionTime += result.result.Duration
		}
		e.statsMutex.Unlock()

		return result.result, result.err
	case <-ctx.Done():
		// Context cancelled
		e.statsMutex.Lock()
		e.stats.QueuedTasks--
		e.stats.FailedTasks++
		e.statsMutex.Unlock()
		return nil, ctx.Err()
	}
}

// ExecuteBatch executes multiple templates asynchronously
func (e *AsyncTemplateExecutor) ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*interfaces.TemplateResult, error) {
	// Check if shutting down
	if e.isShuttingDown() {
		return nil, fmt.Errorf("executor is shutting down")
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
			result, err := e.Execute(ctx, template, options)
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

// getPriority gets the priority from options
func (e *AsyncTemplateExecutor) getPriority(options map[string]interface{}) int {
	// Check if priority is specified
	if priorityVal, ok := options["priority"]; ok {
		if priority, ok := priorityVal.(int); ok {
			return priority
		}
	}

	// Return default priority
	return e.priorityManager.defaultPriority
}

// isShuttingDown checks if the executor is shutting down
func (e *AsyncTemplateExecutor) isShuttingDown() bool {
	e.shutdownMutex.RLock()
	defer e.shutdownMutex.RUnlock()
	return e.isShutdown
}

// Shutdown shuts down the executor
func (e *AsyncTemplateExecutor) Shutdown() {
	// Set shutdown flag
	e.shutdownMutex.Lock()
	if e.isShutdown {
		e.shutdownMutex.Unlock()
		return
	}
	e.isShutdown = true
	e.shutdownMutex.Unlock()

	// Signal shutdown
	close(e.shutdownCh)

	// Close worker pool
	close(e.workerPool.shutdownCh)

	// Wait for workers to finish
	e.workerPool.workerWg.Wait()
}

// GetExecutionStats returns statistics about the executor
func (e *AsyncTemplateExecutor) GetExecutionStats() map[string]interface{} {
	e.statsMutex.RLock()
	defer e.statsMutex.RUnlock()

	avgExecutionTime := time.Duration(0)
	if e.stats.CompletedTasks > 0 {
		avgExecutionTime = time.Duration(int64(e.stats.TotalExecutionTime) / e.stats.CompletedTasks)
	}

	return map[string]interface{}{
		"total_tasks":          e.stats.TotalTasks,
		"completed_tasks":      e.stats.CompletedTasks,
		"failed_tasks":         e.stats.FailedTasks,
		"cached_results":       e.stats.CachedResults,
		"queued_tasks":         e.stats.QueuedTasks,
		"avg_execution_time":   avgExecutionTime,
		"avg_queue_time":       e.stats.AverageQueueTime,
		"active_workers":       e.workerPool.activeWorkers,
		"min_workers":          e.workerPool.minWorkers,
		"max_workers":          e.workerPool.maxWorkers,
	}
}

// helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
