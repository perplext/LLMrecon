package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisJobQueue implements a persistent job queue using Redis
type RedisJobQueue struct {
	client       *redis.Client
	config       RedisQueueConfig
	workers      map[string]*Worker
	logger       Logger
	metrics      *QueueMetrics
	mutex        sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// RedisQueueConfig defines configuration for Redis job queue
type RedisQueueConfig struct {
	// Redis connection
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	
	// Queue configuration
	QueuePrefix      string        `json:"queue_prefix"`
	DefaultPriority  int           `json:"default_priority"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	JobTimeout       time.Duration `json:"job_timeout"`
	VisibilityTimeout time.Duration `json:"visibility_timeout"`
	
	// Worker configuration
	WorkerCount      int           `json:"worker_count"`
	PollInterval     time.Duration `json:"poll_interval"`
	BatchSize        int           `json:"batch_size"`
	
	// Performance settings
	EnablePipelining bool          `json:"enable_pipelining"`
	PoolSize         int           `json:"pool_size"`
	MinIdleConns     int           `json:"min_idle_conns"`
	MaxConnAge       time.Duration `json:"max_conn_age"`
	
	// Monitoring
	EnableMetrics    bool          `json:"enable_metrics"`
	MetricsInterval  time.Duration `json:"metrics_interval"`
}

// Job represents a job in the queue
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Queue       string                 `json:"queue"`
	Priority    int                    `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxRetries  int                    `json:"max_retries"`
	Error       string                 `json:"error,omitempty"`
	Result      interface{}            `json:"result,omitempty"`
	Status      JobStatus              `json:"status"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
	JobStatusCancelled  JobStatus = "cancelled"
)

// Worker processes jobs from the queue
type Worker struct {
	id       string
	queue    *RedisJobQueue
	handler  JobHandler
	logger   Logger
	metrics  *WorkerMetrics
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// JobHandler processes a job
type JobHandler interface {
	ProcessJob(ctx context.Context, job *Job) error
	GetJobTypes() []string
}

// QueueMetrics tracks queue performance
type QueueMetrics struct {
	JobsEnqueued   int64 `json:"jobs_enqueued"`
	JobsProcessed  int64 `json:"jobs_processed"`
	JobsFailed     int64 `json:"jobs_failed"`
	JobsRetried    int64 `json:"jobs_retried"`
	ActiveWorkers  int64 `json:"active_workers"`
	QueueLength    int64 `json:"queue_length"`
	AverageLatency time.Duration `json:"average_latency"`
	ThroughputRPS  float64 `json:"throughput_rps"`
}

// WorkerMetrics tracks individual worker performance
type WorkerMetrics struct {
	WorkerID       string        `json:"worker_id"`
	JobsProcessed  int64         `json:"jobs_processed"`
	JobsFailed     int64         `json:"jobs_failed"`
	AverageLatency time.Duration `json:"average_latency"`
	LastActive     time.Time     `json:"last_active"`
	Status         string        `json:"status"`
}

// Logger interface for queue logging
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// DefaultQueueConfig returns default configuration
func DefaultQueueConfig() RedisQueueConfig {
	return RedisQueueConfig{
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		RedisDB:           0,
		QueuePrefix:       "llm_queue",
		DefaultPriority:   5,
		MaxRetries:        3,
		RetryDelay:        30 * time.Second,
		JobTimeout:        300 * time.Second,
		VisibilityTimeout: 60 * time.Second,
		WorkerCount:       5,
		PollInterval:      1 * time.Second,
		BatchSize:         10,
		EnablePipelining:  true,
		PoolSize:          10,
		MinIdleConns:      5,
		MaxConnAge:        30 * time.Minute,
		EnableMetrics:     true,
		MetricsInterval:   10 * time.Second,
	}
}

// NewRedisJobQueue creates a new Redis job queue
func NewRedisJobQueue(config RedisQueueConfig, logger Logger) (*RedisJobQueue, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr,
		Password:     config.RedisPassword,
		DB:           config.RedisDB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxConnAge:   config.MaxConnAge,
	})
	
	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	queue := &RedisJobQueue{
		client:  client,
		config:  config,
		workers: make(map[string]*Worker),
		logger:  logger,
		metrics: &QueueMetrics{},
		ctx:     ctx,
		cancel:  cancel,
	}
	
	return queue, nil
}

// Start starts the job queue and workers
func (q *RedisJobQueue) Start() error {
	q.logger.Info("Starting Redis job queue", "workers", q.config.WorkerCount)
	
	// Start metrics collection if enabled
	if q.config.EnableMetrics {
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
			q.metricsLoop()
		}()
	}
	
	q.logger.Info("Redis job queue started successfully")
	return nil
}

// Stop stops the job queue and all workers
func (q *RedisJobQueue) Stop() error {
	q.logger.Info("Stopping Redis job queue")
	
	q.cancel()
	
	// Stop all workers
	q.mutex.Lock()
	for _, worker := range q.workers {
		worker.Stop()
	}
	q.mutex.Unlock()
	
	// Wait for all goroutines to finish
	q.wg.Wait()
	
	// Close Redis connection
	if err := q.client.Close(); err != nil {
		q.logger.Error("Error closing Redis connection", "error", err)
		return err
	}
	
	q.logger.Info("Redis job queue stopped")
	return nil
}

// AddWorker adds a worker to process jobs
func (q *RedisJobQueue) AddWorker(handler JobHandler) (*Worker, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	workerID := fmt.Sprintf("worker_%d", len(q.workers)+1)
	
	worker := &Worker{
		id:      workerID,
		queue:   q,
		handler: handler,
		logger:  q.logger,
		metrics: &WorkerMetrics{
			WorkerID: workerID,
			Status:   "idle",
		},
	}
	
	q.workers[workerID] = worker
	
	// Start the worker
	if err := worker.Start(); err != nil {
		delete(q.workers, workerID)
		return nil, fmt.Errorf("failed to start worker %s: %w", workerID, err)
	}
	
	q.logger.Info("Added worker", "worker_id", workerID, "job_types", handler.GetJobTypes())
	return worker, nil
}

// Enqueue adds a job to the queue
func (q *RedisJobQueue) Enqueue(job *Job) error {
	// Set job defaults
	if job.ID == "" {
		job.ID = fmt.Sprintf("job_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
	}
	if job.Priority == 0 {
		job.Priority = q.config.DefaultPriority
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = q.config.MaxRetries
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.ScheduledAt.IsZero() {
		job.ScheduledAt = time.Now()
	}
	job.Status = JobStatusPending
	
	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}
	
	// Add to priority queue (higher priority = lower score for Redis sorted set)
	queueKey := q.getQueueKey(job.Queue)
	score := float64(10-job.Priority) + float64(job.ScheduledAt.Unix())/1000000.0
	
	if err := q.client.ZAdd(q.ctx, queueKey, &redis.Z{
		Score:  score,
		Member: string(jobData),
	}).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	
	// Store job details for tracking
	jobKey := q.getJobKey(job.ID)
	if err := q.client.Set(q.ctx, jobKey, jobData, q.config.JobTimeout).Err(); err != nil {
		return fmt.Errorf("failed to store job details: %w", err)
	}
	
	q.metrics.JobsEnqueued++
	q.logger.Info("Job enqueued", "job_id", job.ID, "type", job.Type, "queue", job.Queue, "priority", job.Priority)
	
	return nil
}

// Dequeue removes and returns the next job from the queue
func (q *RedisJobQueue) Dequeue(queueName string) (*Job, error) {
	queueKey := q.getQueueKey(queueName)
	
	// Get job with lowest score (highest priority, earliest scheduled time)
	result, err := q.client.ZPopMin(q.ctx, queueKey, 1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}
	
	if len(result) == 0 {
		return nil, nil // No jobs available
	}
	
	// Deserialize job
	var job Job
	if err := json.Unmarshal([]byte(result[0].Member.(string)), &job); err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}
	
	// Mark job as processing
	now := time.Now()
	job.Status = JobStatusProcessing
	job.StartedAt = &now
	job.Attempts++
	
	// Update job in Redis
	jobKey := q.getJobKey(job.ID)
	jobData, _ := json.Marshal(job)
	q.client.Set(q.ctx, jobKey, jobData, q.config.JobTimeout)
	
	return &job, nil
}

// CompleteJob marks a job as completed
func (q *RedisJobQueue) CompleteJob(job *Job, result interface{}) error {
	now := time.Now()
	job.Status = JobStatusCompleted
	job.CompletedAt = &now
	job.Result = result
	
	return q.updateJob(job)
}

// FailJob marks a job as failed and potentially retries it
func (q *RedisJobQueue) FailJob(job *Job, err error) error {
	job.Error = err.Error()
	
	if job.Attempts < job.MaxRetries {
		// Retry the job
		job.Status = JobStatusRetrying
		job.ScheduledAt = time.Now().Add(q.config.RetryDelay)
		
		// Re-enqueue for retry
		if retryErr := q.Enqueue(job); retryErr != nil {
			q.logger.Error("Failed to retry job", "job_id", job.ID, "error", retryErr)
		}
		
		q.metrics.JobsRetried++
		q.logger.Info("Job scheduled for retry", "job_id", job.ID, "attempt", job.Attempts, "max_retries", job.MaxRetries)
	} else {
		// Job has exceeded max retries
		now := time.Now()
		job.Status = JobStatusFailed
		job.FailedAt = &now
		
		q.metrics.JobsFailed++
		q.logger.Error("Job failed permanently", "job_id", job.ID, "attempts", job.Attempts, "error", err)
	}
	
	return q.updateJob(job)
}

// GetJob retrieves a job by ID
func (q *RedisJobQueue) GetJob(jobID string) (*Job, error) {
	jobKey := q.getJobKey(jobID)
	
	jobData, err := q.client.Get(q.ctx, jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	
	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}
	
	return &job, nil
}

// GetQueueLength returns the number of jobs in a queue
func (q *RedisJobQueue) GetQueueLength(queueName string) (int64, error) {
	queueKey := q.getQueueKey(queueName)
	return q.client.ZCard(q.ctx, queueKey).Result()
}

// GetMetrics returns current queue metrics
func (q *RedisJobQueue) GetMetrics() *QueueMetrics {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	
	// Update active workers count
	q.metrics.ActiveWorkers = int64(len(q.workers))
	
	return q.metrics
}

// GetWorkerMetrics returns metrics for all workers
func (q *RedisJobQueue) GetWorkerMetrics() []*WorkerMetrics {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	
	metrics := make([]*WorkerMetrics, 0, len(q.workers))
	for _, worker := range q.workers {
		metrics = append(metrics, worker.metrics)
	}
	
	return metrics
}

// Private methods

func (q *RedisJobQueue) getQueueKey(queueName string) string {
	return fmt.Sprintf("%s:queue:%s", q.config.QueuePrefix, queueName)
}

func (q *RedisJobQueue) getJobKey(jobID string) string {
	return fmt.Sprintf("%s:job:%s", q.config.QueuePrefix, jobID)
}

func (q *RedisJobQueue) updateJob(job *Job) error {
	jobKey := q.getJobKey(job.ID)
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}
	
	return q.client.Set(q.ctx, jobKey, jobData, q.config.JobTimeout).Err()
}

func (q *RedisJobQueue) metricsLoop() {
	ticker := time.NewTicker(q.config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			q.updateMetrics()
		case <-q.ctx.Done():
			return
		}
	}
}

func (q *RedisJobQueue) updateMetrics() {
	// Update queue length metrics
	queuePattern := q.config.QueuePrefix + ":queue:*"
	keys, err := q.client.Keys(q.ctx, queuePattern).Result()
	if err != nil {
		q.logger.Error("Failed to get queue keys for metrics", "error", err)
		return
	}
	
	var totalLength int64
	for _, key := range keys {
		length, err := q.client.ZCard(q.ctx, key).Result()
		if err != nil {
			q.logger.Error("Failed to get queue length", "key", key, "error", err)
			continue
		}
		totalLength += length
	}
	
	q.metrics.QueueLength = totalLength
}

// Worker implementation

// Start starts the worker
func (w *Worker) Start() error {
	w.ctx, w.cancel = context.WithCancel(w.queue.ctx)
	
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.processLoop()
	}()
	
	w.metrics.Status = "running"
	w.logger.Info("Worker started", "worker_id", w.id)
	return nil
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.logger.Info("Stopping worker", "worker_id", w.id)
	w.cancel()
	w.wg.Wait()
	w.metrics.Status = "stopped"
}

// processLoop is the main worker loop
func (w *Worker) processLoop() {
	ticker := time.NewTicker(w.queue.config.PollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			w.processJobs()
		case <-w.ctx.Done():
			return
		}
	}
}

// processJobs processes available jobs
func (w *Worker) processJobs() {
	jobTypes := w.handler.GetJobTypes()
	
	for _, jobType := range jobTypes {
		job, err := w.queue.Dequeue(jobType)
		if err != nil {
			w.logger.Error("Failed to dequeue job", "worker_id", w.id, "job_type", jobType, "error", err)
			continue
		}
		
		if job == nil {
			continue // No jobs available
		}
		
		w.processJob(job)
	}
}

// processJob processes a single job
func (w *Worker) processJob(job *Job) {
	start := time.Now()
	w.metrics.LastActive = start
	
	w.logger.Info("Processing job", "worker_id", w.id, "job_id", job.ID, "type", job.Type)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(w.ctx, w.queue.config.JobTimeout)
	defer cancel()
	
	// Process the job
	err := w.handler.ProcessJob(ctx, job)
	
	duration := time.Since(start)
	w.updateMetrics(duration)
	
	if err != nil {
		w.queue.FailJob(job, err)
		w.metrics.JobsFailed++
		w.logger.Error("Job failed", "worker_id", w.id, "job_id", job.ID, "duration", duration, "error", err)
	} else {
		w.queue.CompleteJob(job, nil)
		w.metrics.JobsProcessed++
		w.queue.metrics.JobsProcessed++
		w.logger.Info("Job completed", "worker_id", w.id, "job_id", job.ID, "duration", duration)
	}
}

// updateMetrics updates worker metrics
func (w *Worker) updateMetrics(duration time.Duration) {
	// Calculate rolling average latency
	if w.metrics.AverageLatency == 0 {
		w.metrics.AverageLatency = duration
	} else {
		w.metrics.AverageLatency = (w.metrics.AverageLatency + duration) / 2
	}
}

// DefaultLogger provides a simple logger implementation
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(log.Writer(), "[Queue] ", log.LstdFlags),
	}
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("INFO: "+msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("WARN: "+msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("ERROR: "+msg, args...)
}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	l.logger.Printf("DEBUG: "+msg, args...)
}