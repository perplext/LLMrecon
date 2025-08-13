// Package middleware provides middleware components for the Multi-Provider LLM Integration Framework.
package middleware

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RequestQueueConfig represents the configuration for the request queue
type RequestQueueConfig struct {
	// MaxQueueSize is the maximum number of requests that can be queued
	MaxQueueSize int
	// MaxWaitTime is the maximum time a request can wait in the queue
	MaxWaitTime time.Duration
	// PriorityLevels is the number of priority levels
	PriorityLevels int
}

// QueuedRequest represents a request in the queue
type QueuedRequest struct {
	// Priority is the priority of the request (lower is higher priority)
	Priority int
	// EnqueueTime is the time the request was enqueued
	EnqueueTime time.Time
	// Context is the context of the request
	Context context.Context
	// Function is the function to execute
	Function func(ctx context.Context) (interface{}, error)
	// ResultChan is the channel to send the result to
	ResultChan chan *QueuedResult
}

// QueuedResult represents the result of a queued request
type QueuedResult struct {
	// Result is the result of the request
	Result interface{}
	// Error is the error from the request
	Error error
}

// RequestQueueMiddleware provides request queuing functionality
type RequestQueueMiddleware struct {
	config RequestQueueConfig
	// queues is a slice of queues, one for each priority level
	queues [][]QueuedRequest
	// queueLock is a mutex for the queues
	queueLock sync.Mutex
	// workerCount is the number of workers
	workerCount int
	// workerWg is a wait group for the workers
	workerWg sync.WaitGroup
	// shutdownChan is a channel to signal shutdown
	shutdownChan chan struct{}
	// isRunning indicates whether the queue is running
	isRunning bool
}

// NewRequestQueueMiddleware creates a new request queue middleware
func NewRequestQueueMiddleware(config RequestQueueConfig, workerCount int) *RequestQueueMiddleware {
	// Set default values if not specified
	if config.MaxQueueSize <= 0 {
		config.MaxQueueSize = 100
	}
	if config.MaxWaitTime <= 0 {
		config.MaxWaitTime = 60 * time.Second
	}
	if config.PriorityLevels <= 0 {
		config.PriorityLevels = 3
	}
	if workerCount <= 0 {
		workerCount = 5
	}

	// Create queues for each priority level
	queues := make([][]QueuedRequest, config.PriorityLevels)
	for i := 0; i < config.PriorityLevels; i++ {
		queues[i] = make([]QueuedRequest, 0, config.MaxQueueSize)
	}

	middleware := &RequestQueueMiddleware{
		config:       config,
		queues:       queues,
		queueLock:    sync.Mutex{},
		workerCount:  workerCount,
		workerWg:     sync.WaitGroup{},
		shutdownChan: make(chan struct{}),
		isRunning:    false,
	}

	return middleware
}

// Start starts the request queue workers
func (rq *RequestQueueMiddleware) Start() {
	rq.queueLock.Lock()
	defer rq.queueLock.Unlock()

	if rq.isRunning {
		return
	}

	rq.isRunning = true
	rq.shutdownChan = make(chan struct{})

	// Start workers
	for i := 0; i < rq.workerCount; i++ {
		rq.workerWg.Add(1)
		go rq.worker()
	}
}

// Stop stops the request queue workers
func (rq *RequestQueueMiddleware) Stop() {
	rq.queueLock.Lock()
	if !rq.isRunning {
		rq.queueLock.Unlock()
		return
	}
	rq.isRunning = false
	close(rq.shutdownChan)
	rq.queueLock.Unlock()

	// Wait for workers to finish
	rq.workerWg.Wait()
}

// Execute executes a function with request queuing
func (rq *RequestQueueMiddleware) Execute(ctx context.Context, priority int, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// Check if the queue is running
	if !rq.isRunning {
		return nil, errors.New("request queue is not running")
	}

	// Validate priority
	if priority < 0 || priority >= rq.config.PriorityLevels {
		priority = rq.config.PriorityLevels - 1 // Default to lowest priority
	}

	// Create result channel
	resultChan := make(chan *QueuedResult, 1)

	// Create queued request
	request := QueuedRequest{
		Priority:    priority,
		EnqueueTime: time.Now(),
		Context:     ctx,
		Function:    fn,
		ResultChan:  resultChan,
	}

	// Enqueue the request
	if !rq.enqueue(request) {
		return nil, errors.New("request queue is full")
	}

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		return result.Result, result.Error
	case <-time.After(rq.config.MaxWaitTime):
		return nil, errors.New("request timed out in queue")
	}
}

// enqueue adds a request to the queue
func (rq *RequestQueueMiddleware) enqueue(request QueuedRequest) bool {
	rq.queueLock.Lock()
	defer rq.queueLock.Unlock()

	// Check if the queue is full
	totalQueueSize := 0
	for i := 0; i < rq.config.PriorityLevels; i++ {
		totalQueueSize += len(rq.queues[i])
	}
	if totalQueueSize >= rq.config.MaxQueueSize {
		return false
	}

	// Add the request to the appropriate queue
	rq.queues[request.Priority] = append(rq.queues[request.Priority], request)
	return true
}

// dequeue removes and returns the next request from the queue
func (rq *RequestQueueMiddleware) dequeue() (QueuedRequest, bool) {
	rq.queueLock.Lock()
	defer rq.queueLock.Unlock()

	// Check each priority level, starting with the highest priority
	for i := 0; i < rq.config.PriorityLevels; i++ {
		if len(rq.queues[i]) > 0 {
			// Get the first request
			request := rq.queues[i][0]
			// Remove it from the queue
			rq.queues[i] = rq.queues[i][1:]
			return request, true
		}
	}

	// No requests in the queue
	return QueuedRequest{}, false
}

// worker processes requests from the queue
func (rq *RequestQueueMiddleware) worker() {
	defer rq.workerWg.Done()

	for {
		// Check if we should shut down
		select {
		case <-rq.shutdownChan:
			return
		default:
			// Continue processing
		}

		// Get the next request
		request, ok := rq.dequeue()
		if !ok {
			// No requests in the queue, wait a bit
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Check if the request context is still valid
		if request.Context.Err() != nil {
			// Context is canceled or timed out, skip this request
			request.ResultChan <- &QueuedResult{
				Result: nil,
				Error:  request.Context.Err(),
			}
			continue
		}

		// Check if the request has been in the queue too long
		if time.Since(request.EnqueueTime) > rq.config.MaxWaitTime {
			// Request has been in the queue too long, skip it
			request.ResultChan <- &QueuedResult{
				Result: nil,
				Error:  errors.New("request timed out in queue"),
			}
			continue
		}

		// Execute the request
		result, err := request.Function(request.Context)

		// Send the result
		request.ResultChan <- &QueuedResult{
			Result: result,
			Error:  err,
		}
	}
}

// GetQueueStats returns statistics about the queue
func (rq *RequestQueueMiddleware) GetQueueStats() map[string]interface{} {
	rq.queueLock.Lock()
	defer rq.queueLock.Unlock()

	stats := make(map[string]interface{})
	stats["is_running"] = rq.isRunning
	stats["worker_count"] = rq.workerCount
	stats["max_queue_size"] = rq.config.MaxQueueSize
	stats["priority_levels"] = rq.config.PriorityLevels

	// Queue sizes by priority
	queueSizes := make([]int, rq.config.PriorityLevels)
	for i := 0; i < rq.config.PriorityLevels; i++ {
		queueSizes[i] = len(rq.queues[i])
	}
	stats["queue_sizes"] = queueSizes

	// Total queue size
	totalQueueSize := 0
	for i := 0; i < rq.config.PriorityLevels; i++ {
		totalQueueSize += len(rq.queues[i])
	}
	stats["total_queue_size"] = totalQueueSize

	return stats
}

// UpdateConfig updates the request queue configuration
func (rq *RequestQueueMiddleware) UpdateConfig(config RequestQueueConfig) {
	rq.queueLock.Lock()
	defer rq.queueLock.Unlock()

	// Update the configuration
	rq.config = config

	// Resize the queues if needed
	if len(rq.queues) != config.PriorityLevels {
		newQueues := make([][]QueuedRequest, config.PriorityLevels)
		for i := 0; i < config.PriorityLevels; i++ {
			if i < len(rq.queues) {
				newQueues[i] = rq.queues[i]
			} else {
				newQueues[i] = make([]QueuedRequest, 0, config.MaxQueueSize)
			}
		}
		rq.queues = newQueues
	}
}
