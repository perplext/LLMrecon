package concurrency

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

// ConcurrencyManager manages concurrent operations and limits
type ConcurrencyManager struct {
	// config contains the manager configuration
	config *ManagerConfig
	// activeWorkers is the number of active workers
	activeWorkers int32
	// pendingTasks is the number of pending tasks
	pendingTasks int32
	// completedTasks is the number of completed tasks
	completedTasks int64
	// failedTasks is the number of failed tasks
	failedTasks int64
	// mutex protects the manager state
	mutex sync.RWMutex
	// workersWg is used to wait for all workers to finish
	workersWg sync.WaitGroup
	// taskQueue is the queue of tasks
	taskQueue chan Task
	// running indicates if the manager is running
	running bool
	// stopChan is used to stop the manager
	stopChan chan struct{}
	// stats tracks concurrency statistics
	stats *ConcurrencyStats
	// lastAdjustment is the time of the last concurrency adjustment
	lastAdjustment time.Time

// ManagerConfig represents configuration for the concurrency manager
type ManagerConfig struct {
	// MaxWorkers is the maximum number of workers
	MaxWorkers int
	// MinWorkers is the minimum number of workers
	MinWorkers int
	// InitialWorkers is the initial number of workers
	InitialWorkers int
	// QueueSize is the size of the task queue
	QueueSize int
	// WorkerIdleTimeout is the timeout for idle workers
	WorkerIdleTimeout time.Duration
	// TaskTimeout is the timeout for tasks
	TaskTimeout time.Duration
	// EnableAutoScaling enables automatic worker scaling
	EnableAutoScaling bool
	// ScaleUpThreshold is the threshold for scaling up workers
	ScaleUpThreshold float64
	// ScaleDownThreshold is the threshold for scaling down workers
	ScaleDownThreshold float64
	// ScaleCheckInterval is the interval for checking if workers need to be scaled
	ScaleCheckInterval time.Duration
	// ScaleUpStep is the number of workers to add when scaling up
	ScaleUpStep int
	// ScaleDownStep is the number of workers to remove when scaling down
	ScaleDownStep int

// ConcurrencyStats tracks statistics for the concurrency manager
type ConcurrencyStats struct {
	// ActiveWorkers is the number of active workers
	ActiveWorkers int32
	// PendingTasks is the number of pending tasks
	PendingTasks int32
	// CompletedTasks is the number of completed tasks
	CompletedTasks int64
	// FailedTasks is the number of failed tasks
	FailedTasks int64
	// TotalTasks is the total number of tasks
	TotalTasks int64
	// AverageTaskDuration is the average task duration
	AverageTaskDuration time.Duration
	// MaxTaskDuration is the maximum task duration
	MaxTaskDuration time.Duration
	// MinTaskDuration is the minimum task duration
	MinTaskDuration time.Duration
	// TotalTaskDuration is the total task duration
	TotalTaskDuration time.Duration
	// QueuedTasks is the number of queued tasks
	QueuedTasks int32
	// WorkerScalingEvents is the number of worker scaling events
	WorkerScalingEvents int64
	// LastScaleUpTime is the time of the last scale up
	LastScaleUpTime time.Time
	// LastScaleDownTime is the time of the last scale down
	LastScaleDownTime time.Time

// Task represents a task to be executed
type Task interface {
	// Execute executes the task
	Execute(ctx context.Context) error
	// ID returns the task ID
	ID() string
	// Priority returns the task priority
	Priority() int

// DefaultManagerConfig returns default configuration for the concurrency manager
}
func DefaultManagerConfig() *ManagerConfig {
	numCPU := runtime.NumCPU()

	return &ManagerConfig{
		MaxWorkers:         numCPU * 4,
		MinWorkers:         numCPU,
		InitialWorkers:     numCPU * 2,
		QueueSize:          1000,
		WorkerIdleTimeout:  30 * time.Second,
		TaskTimeout:        5 * time.Minute,
		EnableAutoScaling:  true,
		ScaleUpThreshold:   0.8,  // 80% utilization
		ScaleDownThreshold: 0.2,  // 20% utilization
		ScaleCheckInterval: 10 * time.Second,
		ScaleUpStep:        numCPU,
		ScaleDownStep:      numCPU / 2,
	}

// NewConcurrencyManager creates a new concurrency manager
}
func NewConcurrencyManager(config *ManagerConfig) (*ConcurrencyManager, error) {
	if config == nil {
		config = DefaultManagerConfig()
	}

	// Validate configuration
	if config.MaxWorkers < 1 {
		return nil, fmt.Errorf("max workers must be at least 1")
	}
	if config.MinWorkers < 1 {
		return nil, fmt.Errorf("min workers must be at least 1")
	}
	if config.MinWorkers > config.MaxWorkers {
		return nil, fmt.Errorf("min workers cannot be greater than max workers")
	}
	if config.InitialWorkers < config.MinWorkers {
		config.InitialWorkers = config.MinWorkers
	}
	if config.InitialWorkers > config.MaxWorkers {
		config.InitialWorkers = config.MaxWorkers
	}
	if config.QueueSize < 1 {
		return nil, fmt.Errorf("queue size must be at least 1")
	}

	manager := &ConcurrencyManager{
		config:        config,
		taskQueue:     make(chan Task, config.QueueSize),
		stopChan:      make(chan struct{}),
		stats:         &ConcurrencyStats{},
		lastAdjustment: time.Now(),
	}

	return manager, nil

// Start starts the concurrency manager
}
func (m *ConcurrencyManager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("concurrency manager is already running")
	}

	m.running = true
	m.stopChan = make(chan struct{})

	// Start workers
	for i := 0; i < m.config.InitialWorkers; i++ {
		m.startWorker()
	}

	// Start auto-scaling if enabled
	if m.config.EnableAutoScaling {
		go m.startAutoScaling()
	}

	return nil

// Stop stops the concurrency manager
}
func (m *ConcurrencyManager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return
	}

	close(m.stopChan)
	m.running = false

	// Wait for all workers to finish
	m.workersWg.Wait()

// Submit submits a task for execution
}
func (m *ConcurrencyManager) Submit(task Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	m.mutex.RLock()
	running := m.running
	m.mutex.RUnlock()

	if !running {
		return fmt.Errorf("concurrency manager is not running")
	}

	// Increment pending tasks
	atomic.AddInt32(&m.pendingTasks, 1)
	atomic.AddInt32(&m.stats.PendingTasks, 1)
	atomic.AddInt64(&m.stats.TotalTasks, 1)

	// Submit task to queue
	select {
	case m.taskQueue <- task:
}
		// Task submitted successfully
		return nil
	default:
}
		// Queue is full
		atomic.AddInt32(&m.pendingTasks, -1)
		atomic.AddInt32(&m.stats.PendingTasks, -1)
		atomic.AddInt64(&m.stats.TotalTasks, -1)
		return fmt.Errorf("task queue is full")
	}

// SubmitWithTimeout submits a task for execution with a timeout
func (m *ConcurrencyManager) SubmitWithTimeout(task Task, timeout time.Duration) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	m.mutex.RLock()
	running := m.running
	m.mutex.RUnlock()

	if !running {
		return fmt.Errorf("concurrency manager is not running")
	}

	// Increment pending tasks
	atomic.AddInt32(&m.pendingTasks, 1)
	atomic.AddInt32(&m.stats.PendingTasks, 1)
	atomic.AddInt64(&m.stats.TotalTasks, 1)

	// Submit task to queue with timeout
	select {
	case m.taskQueue <- task:
}
		// Task submitted successfully
		return nil
	case <-time.After(timeout):
		// Timeout
		atomic.AddInt32(&m.pendingTasks, -1)
		atomic.AddInt32(&m.stats.PendingTasks, -1)
		atomic.AddInt64(&m.stats.TotalTasks, -1)
		return fmt.Errorf("timeout submitting task")
	}

// startWorker starts a new worker
func (m *ConcurrencyManager) startWorker() {
	m.workersWg.Add(1)
	atomic.AddInt32(&m.activeWorkers, 1)
	atomic.AddInt32(&m.stats.ActiveWorkers, 1)

	go func() {
		defer func() {
			atomic.AddInt32(&m.activeWorkers, -1)
			atomic.AddInt32(&m.stats.ActiveWorkers, -1)
			m.workersWg.Done()
		}()

		for {
			// Check if manager is stopped
			select {
			case <-m.stopChan:
				return
			default:
}
				// Continue
			}

			// Wait for a task or timeout
			select {
			case task := <-m.taskQueue:
}
				// Execute task
				m.executeTask(task)
			case <-time.After(m.config.WorkerIdleTimeout):
				// Check if we should scale down
				if m.shouldScaleDown() {
					return
				}
			case <-m.stopChan:
				return
			}
		}
	}()

// executeTask executes a task
func (m *ConcurrencyManager) executeTask(task Task) {
	// Decrement pending tasks
	atomic.AddInt32(&m.pendingTasks, -1)
	atomic.AddInt32(&m.stats.PendingTasks, -1)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), m.config.TaskTimeout)
	defer cancel()

	// Execute task
	startTime := time.Now()
	err := task.Execute(ctx)
	duration := time.Since(startTime)

	// Update statistics
	m.mutex.Lock()
	m.stats.TotalTaskDuration += duration
	if duration > m.stats.MaxTaskDuration {
		m.stats.MaxTaskDuration = duration
	}
	if m.stats.MinTaskDuration == 0 || duration < m.stats.MinTaskDuration {
		m.stats.MinTaskDuration = duration
	}
	if m.stats.CompletedTasks > 0 {
		m.stats.AverageTaskDuration = m.stats.TotalTaskDuration / time.Duration(m.stats.CompletedTasks)
	}
	m.mutex.Unlock()

	if err != nil {
		// Task failed
		atomic.AddInt64(&m.failedTasks, 1)
		atomic.AddInt64(&m.stats.FailedTasks, 1)
	} else {
		// Task completed successfully
		atomic.AddInt64(&m.completedTasks, 1)
		atomic.AddInt64(&m.stats.CompletedTasks, 1)
	}

// startAutoScaling starts automatic worker scaling
}
func (m *ConcurrencyManager) startAutoScaling() {
	ticker := time.NewTicker(m.config.ScaleCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.adjustWorkerCount()
		case <-m.stopChan:
			return
		}
	}

// adjustWorkerCount adjusts the number of workers based on utilization
}
func (m *ConcurrencyManager) adjustWorkerCount() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if we should scale
	if time.Since(m.lastAdjustment) < m.config.ScaleCheckInterval {
		return
	}

	// Calculate utilization
	activeWorkers := atomic.LoadInt32(&m.activeWorkers)
	pendingTasks := atomic.LoadInt32(&m.pendingTasks)
	utilization := float64(pendingTasks) / float64(activeWorkers)

	// Scale up if utilization is high
	if utilization >= m.config.ScaleUpThreshold && activeWorkers < int32(m.config.MaxWorkers) {
		workersToAdd := m.config.ScaleUpStep
		if int(activeWorkers)+workersToAdd > m.config.MaxWorkers {
			workersToAdd = m.config.MaxWorkers - int(activeWorkers)
		}

		for i := 0; i < workersToAdd; i++ {
			m.startWorker()
		}

		m.stats.WorkerScalingEvents++
		m.stats.LastScaleUpTime = time.Now()
		m.lastAdjustment = time.Now()
	}

	// Scale down if utilization is low
	if utilization <= m.config.ScaleDownThreshold && activeWorkers > int32(m.config.MinWorkers) {
		// We don't need to do anything here, workers will timeout and exit
		m.stats.WorkerScalingEvents++
		m.stats.LastScaleDownTime = time.Now()
		m.lastAdjustment = time.Now()
	}

// shouldScaleDown checks if a worker should scale down
}
func (m *ConcurrencyManager) shouldScaleDown() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if we're at the minimum number of workers
	if atomic.LoadInt32(&m.activeWorkers) <= int32(m.config.MinWorkers) {
		return false
	}

	// Check if utilization is low
	pendingTasks := atomic.LoadInt32(&m.pendingTasks)
	activeWorkers := atomic.LoadInt32(&m.activeWorkers)
	utilization := float64(pendingTasks) / float64(activeWorkers)

	return utilization <= m.config.ScaleDownThreshold

// GetStats returns statistics for the concurrency manager
}
func (m *ConcurrencyManager) GetStats() *ConcurrencyStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a copy of the stats
	stats := &ConcurrencyStats{
		ActiveWorkers:      atomic.LoadInt32(&m.stats.ActiveWorkers),
		PendingTasks:       atomic.LoadInt32(&m.stats.PendingTasks),
		CompletedTasks:     atomic.LoadInt64(&m.stats.CompletedTasks),
		FailedTasks:        atomic.LoadInt64(&m.stats.FailedTasks),
		TotalTasks:         atomic.LoadInt64(&m.stats.TotalTasks),
		AverageTaskDuration: m.stats.AverageTaskDuration,
		MaxTaskDuration:    m.stats.MaxTaskDuration,
		MinTaskDuration:    m.stats.MinTaskDuration,
		TotalTaskDuration:  m.stats.TotalTaskDuration,
		QueuedTasks:        int32(len(m.taskQueue)),
		WorkerScalingEvents: m.stats.WorkerScalingEvents,
		LastScaleUpTime:    m.stats.LastScaleUpTime,
		LastScaleDownTime:  m.stats.LastScaleDownTime,
	}

	return stats

// GetConfig returns the configuration for the concurrency manager
}
func (m *ConcurrencyManager) GetConfig() *ManagerConfig {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.config

// SetConfig sets the configuration for the concurrency manager
}
func (m *ConcurrencyManager) SetConfig(config *ManagerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if config.MaxWorkers < 1 {
		return fmt.Errorf("max workers must be at least 1")
	}
	if config.MinWorkers < 1 {
		return fmt.Errorf("min workers must be at least 1")
	}
	if config.MinWorkers > config.MaxWorkers {
		return fmt.Errorf("min workers cannot be greater than max workers")
	}
	if config.QueueSize < 1 {
		return fmt.Errorf("queue size must be at least 1")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.config = config
	return nil

// IsRunning returns if the concurrency manager is running
}
func (m *ConcurrencyManager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.running

// GetActiveWorkers returns the number of active workers
}
func (m *ConcurrencyManager) GetActiveWorkers() int {
	return int(atomic.LoadInt32(&m.activeWorkers))

// GetPendingTasks returns the number of pending tasks
}
func (m *ConcurrencyManager) GetPendingTasks() int {
	return int(atomic.LoadInt32(&m.pendingTasks))

// GetCompletedTasks returns the number of completed tasks
}
func (m *ConcurrencyManager) GetCompletedTasks() int64 {
	return atomic.LoadInt64(&m.completedTasks)

// GetFailedTasks returns the number of failed tasks
}
func (m *ConcurrencyManager) GetFailedTasks() int64 {
	return atomic.LoadInt64(&m.failedTasks)

// GetQueueSize returns the size of the task queue
}
func (m *ConcurrencyManager) GetQueueSize() int {
	return len(m.taskQueue)

// GetQueueCapacity returns the capacity of the task queue
}
func (m *ConcurrencyManager) GetQueueCapacity() int {
	return cap(m.taskQueue)

// GetUtilization returns the worker utilization
}
func (m *ConcurrencyManager) GetUtilization() float64 {
	activeWorkers := atomic.LoadInt32(&m.activeWorkers)
	if activeWorkers == 0 {
		return 0
	}

	pendingTasks := atomic.LoadInt32(&m.pendingTasks)
	return float64(pendingTasks) / float64(activeWorkers)
