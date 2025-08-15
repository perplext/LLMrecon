// Package ratelimit provides rate limiting functionality for template execution
package ratelimit

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
)

// priorityItem represents an item in the priority queue
type priorityItem struct {
	userID   string
	priority int
	index    int
	added    time.Time
	ctx      context.Context
	done     chan struct{}

// priorityQueue implements a priority queue for rate limiting
type priorityQueue struct {
	items []*priorityItem
	mu    sync.Mutex
	processCh chan struct{}
	shutdown chan struct{}
	running int32 // atomic flag for running state
	
	// Performance optimization: batch processing
	batchSize int
	
	// Performance optimization: fast path for high priority items
	fastPathEnabled bool
	fastPathPriority int
	fastPathCh chan *priorityItem
	
	// Performance optimization: queue statistics for adaptive behavior
	totalProcessed int64
	totalWaitTime  int64 // nanoseconds
	lastAdaptation time.Time

// Len returns the length of the queue
func (pq *priorityQueue) Len() int {
	return len(pq.items)

// Less compares two items in the queue
// Higher priority comes first, and for equal priorities, earlier arrival time comes first
func (pq *priorityQueue) Less(i, j int) bool {
	// Higher priority comes first
	if pq.items[i].priority != pq.items[j].priority {
		return pq.items[i].priority > pq.items[j].priority
	}
	
	// For equal priorities, earlier arrival time comes first (FIFO)
	// This ensures that items with the same priority are processed in the order they were added
	return pq.items[i].added.Before(pq.items[j].added)

// Swap swaps two items in the queue
func (pq *priorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j

// Push adds an item to the queue
func (pq *priorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*priorityItem)
	item.index = n
	pq.items = append(pq.items, item)

// Pop removes and returns the highest priority item from the queue
func (pq *priorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	pq.items = old[0 : n-1]
	return item

// newPriorityQueue creates a new priority queue
func newPriorityQueue() *priorityQueue {
	pq := &priorityQueue{
		items: make([]*priorityItem, 0, 128), // Pre-allocate capacity for better performance
		processCh: make(chan struct{}, 8),    // Increased buffer size to reduce contention
		shutdown: make(chan struct{}),
		batchSize: 16,                         // Process items in batches for efficiency
		fastPathEnabled: true,
		fastPathPriority: 8,                   // Items with priority >= 8 take the fast path
		fastPathCh: make(chan *priorityItem, 32), // Buffer for fast path items
		lastAdaptation: time.Now(),
	}
	heap.Init(pq)
	pq.startProcessor()
	return pq

// add adds a request to the priority queue
func (pq *priorityQueue) add(userID string, priority int, ctx context.Context) chan struct{} {
	// Fast path for high-priority items if enabled
	if pq.fastPathEnabled && priority >= pq.fastPathPriority {
		done := make(chan struct{})
		item := &priorityItem{
			userID:   userID,
			priority: priority,
			added:    time.Now(),
			ctx:      ctx,
			done:     done,
		}
		
		// Try to send directly to fast path channel
		select {
		case pq.fastPathCh <- item:
			return done
		default:
			// Fast path channel is full, fall back to normal path
		}
	}
	
	// Normal path for regular priority items
	pq.mu.Lock()
	
	done := make(chan struct{})
	item := &priorityItem{
		userID:   userID,
		priority: priority,
		added:    time.Now(),
		ctx:      ctx,
		done:     done,
	}
	
	heap.Push(pq, item)
	pq.mu.Unlock()
	
	// Signal that a new item has been added
	select {
	case pq.processCh <- struct{}{}:
		// Signal sent
	default:
		// Channel already has a signal, no need to send another
	}
	
	return done

// nextBatch returns a batch of items from the queue
// It returns the items and a boolean indicating if any items were returned
func (pq *priorityQueue) nextBatch() ([]*priorityItem, bool) {
	pq.mu.Lock()
	if pq.Len() == 0 {
		pq.mu.Unlock()
		return nil, false
	}
	
	// Determine batch size (min of queue length and configured batch size)
	batchSize := pq.batchSize
	if pq.Len() < batchSize {
		batchSize = pq.Len()
	}
	
	// Extract batch of items
	batch := make([]*priorityItem, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		if pq.Len() == 0 {
			break
		}
		item := heap.Pop(pq).(*priorityItem)
		
		// Check if the context is still valid
		select {
		case <-item.ctx.Done():
			// Skip this item, context is canceled
			continue
		default:
			batch = append(batch, item)
		}
	}
	pq.mu.Unlock()
	
	if len(batch) == 0 {
		return nil, false
	}
	
	return batch, true

// processBatch processes a batch of items from the queue
func (pq *priorityQueue) processBatch() int {
	batch, ok := pq.nextBatch()
	if !ok {
		return 0
	}
	
	// Process all items in the batch
	for _, item := range batch {
		// Signal that this item is being processed
		close(item.done)
		
		// Update statistics
		waitTime := time.Since(item.added).Nanoseconds()
		atomic.AddInt64(&pq.totalWaitTime, waitTime)
		atomic.AddInt64(&pq.totalProcessed, 1)
	}
	
	return len(batch)

// startProcessor starts background goroutines to process items in the queue
func (pq *priorityQueue) startProcessor() {
	// Set running state atomically
	if !atomic.CompareAndSwapInt32(&pq.running, 0, 1) {
		return // Already running
	}
	
	// Start the main processor goroutine
	go func() {
		processingDelay := time.Microsecond * 100 // Initial delay between processing batches
		
		for {
			select {
			case <-pq.processCh:
				// Process items in the queue in batches
				processed := 0
				for {
					count := pq.processBatch()
					if count == 0 {
						break
					}
					processed += count
					
					// Adaptive processing: if we're processing a lot of items,
					// reduce the delay to increase throughput
					if processed > pq.batchSize*4 {
						processingDelay = time.Microsecond * 10
					} else {
						processingDelay = time.Microsecond * 100
					}
					
					// Small delay to prevent CPU spinning, but much shorter than before
					time.Sleep(processingDelay)
				}
				
			case <-pq.shutdown:
				return
			}
		}
	}()
	
	// Start the fast path processor goroutine if enabled
	if pq.fastPathEnabled {
		go func() {
			for {
				select {
				case item := <-pq.fastPathCh:
					// Process fast path item immediately
					select {
					case <-item.ctx.Done():
						// Context canceled, skip this item
					default:
						// Signal that this item is being processed
						close(item.done)
						
						// Update statistics
						waitTime := time.Since(item.added).Nanoseconds()
						atomic.AddInt64(&pq.totalWaitTime, waitTime)
						atomic.AddInt64(&pq.totalProcessed, 1)
					}
				case <-pq.shutdown:
					return
				}
			}
		}()
	}
	
	// Start the adaptation goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				pq.adaptQueueParameters()
			case <-pq.shutdown:
				return
			}
		}
	}()

// stop stops the background processor
func (pq *priorityQueue) stop() {
	// Set running state atomically
	if !atomic.CompareAndSwapInt32(&pq.running, 1, 0) {
		return // Not running
	}
	
	close(pq.shutdown)

// waitForTurn waits for this user's turn in the priority queue
// Returns true if it's this user's turn, false if the context was canceled
func (pq *priorityQueue) waitForTurn(userID string, priority int, ctx context.Context) bool {
	// Add to queue and get the done channel
	done := pq.add(userID, priority, ctx)
	
	// Wait for our turn or context cancellation
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}

// adaptQueueParameters dynamically adjusts queue parameters based on load
func (pq *priorityQueue) adaptQueueParameters() {
	processed := atomic.LoadInt64(&pq.totalProcessed)
	if processed == 0 {
		return // No data yet
	}
	
	// Calculate average wait time
	avgWaitTime := atomic.LoadInt64(&pq.totalWaitTime) / processed
	
	// Adjust batch size based on wait time
	// If wait times are high, increase batch size to process more items at once
	if avgWaitTime > int64(time.Millisecond*5) {
		if pq.batchSize < 64 {
			pq.batchSize += 4
		}
	} else if avgWaitTime < int64(time.Millisecond) {
		if pq.batchSize > 8 {
			pq.batchSize -= 2
		}
	}
	
	// Adjust fast path priority threshold based on load
	// If we're processing a lot of items, make the fast path more selective
	pq.mu.Lock()
	queueLen := len(pq.items)
	pq.mu.Unlock()
	
	if queueLen > 100 {
		// Under high load, only the highest priority items take the fast path
		pq.fastPathPriority = 9
	} else if queueLen > 50 {
		pq.fastPathPriority = 8
	} else {
		// Under normal load, more items can take the fast path
		pq.fastPathPriority = 7
	}
	
	// Reset statistics periodically to adapt to changing conditions
	if time.Since(pq.lastAdaptation) > time.Minute {
		atomic.StoreInt64(&pq.totalProcessed, 0)
		atomic.StoreInt64(&pq.totalWaitTime, 0)
		pq.lastAdaptation = time.Now()
	}

// GetQueueStats returns statistics about the priority queue
func (pq *priorityQueue) GetQueueStats() map[string]interface{} {
	pq.mu.Lock()
	queueLen := len(pq.items)
	pq.mu.Unlock()
	
	processed := atomic.LoadInt64(&pq.totalProcessed)
	avgWaitTime := int64(0)
	if processed > 0 {
		avgWaitTime = atomic.LoadInt64(&pq.totalWaitTime) / processed
	}
	
	return map[string]interface{}{
		"queue_length":       queueLen,
		"batch_size":         pq.batchSize,
		"fast_path_priority": pq.fastPathPriority,
		"fast_path_enabled":  pq.fastPathEnabled,
		"total_processed":    processed,
		"avg_wait_time_ns":   avgWaitTime,
	}
