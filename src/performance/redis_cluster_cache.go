package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClusterCache provides advanced caching with Redis clustering support
type RedisClusterCache struct {
	config      RedisClusterCacheConfig
	primary     *redis.ClusterClient
	replicas    []*redis.ClusterClient
	partitioner *CachePartitioner
	warmer      *CacheWarmer
	invalidator *CacheInvalidator
	metrics     *CacheMetrics
	logger      Logger
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// RedisClusterCacheConfig defines configuration for Redis cluster caching
type RedisClusterCacheConfig struct {
	// Cluster configuration
	PrimaryAddrs    []string `json:"primary_addrs"`
	ReplicaAddrs    []string `json:"replica_addrs"`
	Password        string   `json:"password"`
	MaxRedirects    int      `json:"max_redirects"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	
	// Partitioning strategy
	PartitionStrategy   PartitionStrategy `json:"partition_strategy"`
	PartitionCount      int               `json:"partition_count"`
	ReplicationFactor   int               `json:"replication_factor"`
	ConsistentHashing   bool              `json:"consistent_hashing"`
	
	// Cache settings
	DefaultTTL          time.Duration `json:"default_ttl"`
	MaxValueSize        int64         `json:"max_value_size"`
	CompressionEnabled  bool          `json:"compression_enabled"`
	CompressionThreshold int64        `json:"compression_threshold"`
	
	// Invalidation settings
	InvalidationStrategy InvalidationStrategy `json:"invalidation_strategy"`
	TagsEnabled         bool                  `json:"tags_enabled"`
	MaxTags             int                   `json:"max_tags"`
	
	// Cache warming
	WarmingEnabled      bool          `json:"warming_enabled"`
	WarmingInterval     time.Duration `json:"warming_interval"`
	WarmingConcurrency  int           `json:"warming_concurrency"`
	WarmingBatchSize    int           `json:"warming_batch_size"`
	
	// Performance optimization
	PipelineSize        int           `json:"pipeline_size"`
	PoolSize           int           `json:"pool_size"`
	MinIdleConns       int           `json:"min_idle_conns"`
	ReadPreference     ReadPreference `json:"read_preference"`
	
	// Monitoring
	EnableMetrics      bool          `json:"enable_metrics"`
	MetricsInterval    time.Duration `json:"metrics_interval"`
	EnableTracing      bool          `json:"enable_tracing"`
}

// PartitionStrategy defines cache partitioning strategies
type PartitionStrategy string

const (
	PartitionByHash       PartitionStrategy = "hash"
	PartitionByRange      PartitionStrategy = "range"
	PartitionByConsistent PartitionStrategy = "consistent"
	PartitionByLoad       PartitionStrategy = "load"
)

// InvalidationStrategy defines cache invalidation strategies
type InvalidationStrategy string

const (
	InvalidationLRU       InvalidationStrategy = "lru"
	InvalidationTTL       InvalidationStrategy = "ttl"
	InvalidationTags      InvalidationStrategy = "tags"
	InvalidationCallbackStrategy  InvalidationStrategy = "callback"
)

// ReadPreference defines read preference for cache operations
type ReadPreference string

const (
	ReadPrimary          ReadPreference = "primary"
	ReadReplica          ReadPreference = "replica"
	ReadNearest          ReadPreference = "nearest"
	ReadPreferReplica    ReadPreference = "prefer_replica"
)

// CacheEntry represents a cache entry with metadata
type CacheEntry struct {
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	TTL         time.Duration          `json:"ttl"`
	CreatedAt   time.Time              `json:"created_at"`
	AccessedAt  time.Time              `json:"accessed_at"`
	AccessCount int64                  `json:"access_count"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	Compressed  bool                   `json:"compressed"`
	Size        int64                  `json:"size"`
}

// CachePartitioner handles cache partitioning across cluster nodes
type CachePartitioner struct {
	strategy    PartitionStrategy
	partitions  []CachePartition
	hashRing    *ConsistentHashRing
	config      CachePartitionerConfig
	metrics     *PartitionerMetrics
	mutex       sync.RWMutex
}

// CachePartition represents a cache partition
type CachePartition struct {
	ID       int      `json:"id"`
	Node     string   `json:"node"`
	KeyRange KeyRange `json:"key_range"`
	Load     float64  `json:"load"`
	Health   PartitionHealth `json:"health"`
}

// KeyRange defines a range of keys for partitioning
type KeyRange struct {
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

// PartitionHealth represents partition health status
type PartitionHealth string

const (
	PartitionHealthy   PartitionHealth = "healthy"
	PartitionDegraded  PartitionHealth = "degraded"
	PartitionUnhealthy PartitionHealth = "unhealthy"
)

// ConsistentHashRing implements consistent hashing for cache distribution
type ConsistentHashRing struct {
	nodes       map[uint32]string
	sortedNodes []uint32
	replicas    int
	mutex       sync.RWMutex
}

// CacheWarmer handles proactive cache warming
type CacheWarmer struct {
	config       CacheWarmerConfig
	strategies   map[string]WarmingStrategy
	scheduler    *WarmingScheduler
	executor     *WarmingExecutor
	predictor    *AccessPredictor
	metrics      *WarmerMetrics
	logger       Logger
	mutex        sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// WarmingStrategy defines cache warming strategies
type WarmingStrategy interface {
	ShouldWarm(key string, entry *CacheEntry) bool
	GetPriority(key string, entry *CacheEntry) int
	GetWarmingData(key string) (interface{}, error)
}

// CacheInvalidator handles cache invalidation
type CacheInvalidator struct {
	config     CacheInvalidatorConfig
	strategies map[InvalidationStrategy]InvalidationHandler
	tagIndex   *TagIndex
	callbacks  map[string]InvalidationCallback
	metrics    *InvalidatorMetrics
	logger     Logger
	mutex      sync.RWMutex
}

// InvalidationHandler handles specific invalidation strategies
type InvalidationHandler interface {
	ShouldInvalidate(entry *CacheEntry) bool
	Invalidate(key string, entry *CacheEntry) error
}

// InvalidationCallback is called when cache entries are invalidated
type InvalidationCallback func(key string, entry *CacheEntry, reason string)

// TagIndex maintains an index of cache entries by tags
type TagIndex struct {
	tagToKeys map[string]map[string]struct{}
	keyToTags map[string]map[string]struct{}
	mutex     sync.RWMutex
}

// Various metrics structures
type CacheMetrics struct {
	Hits               int64         `json:"hits"`
	Misses             int64         `json:"misses"`
	Sets               int64         `json:"sets"`
	Deletes            int64         `json:"deletes"`
	Evictions          int64         `json:"evictions"`
	HitRatio           float64       `json:"hit_ratio"`
	AverageLatency     time.Duration `json:"average_latency"`
	TotalKeys          int64         `json:"total_keys"`
	TotalSize          int64         `json:"total_size"`
	PartitionMetrics   []*PartitionMetrics `json:"partition_metrics"`
}

type PartitionMetrics struct {
	PartitionID int     `json:"partition_id"`
	Load        float64 `json:"load"`
	KeyCount    int64   `json:"key_count"`
	DataSize    int64   `json:"data_size"`
	Health      PartitionHealth `json:"health"`
}

type PartitionerMetrics struct {
	Rebalances    int64 `json:"rebalances"`
	Migrations    int64 `json:"migrations"`
	LoadVariance  float64 `json:"load_variance"`
}

type WarmerMetrics struct {
	WarmingJobs     int64         `json:"warming_jobs"`
	JobsCompleted   int64         `json:"jobs_completed"`
	JobsFailed      int64         `json:"jobs_failed"`
	AverageWarmTime time.Duration `json:"average_warm_time"`
	CacheHitIncrease float64      `json:"cache_hit_increase"`
}

type InvalidatorMetrics struct {
	Invalidations    int64 `json:"invalidations"`
	TagInvalidations int64 `json:"tag_invalidations"`
	TTLExpired       int64 `json:"ttl_expired"`
	ManualEvictions  int64 `json:"manual_evictions"`
}

// Configuration structures
type CachePartitionerConfig struct {
	RebalanceThreshold float64       `json:"rebalance_threshold"`
	RebalanceInterval  time.Duration `json:"rebalance_interval"`
	MigrationBatchSize int           `json:"migration_batch_size"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

type CacheWarmerConfig struct {
	Strategies         []string      `json:"strategies"`
	WarmingInterval    time.Duration `json:"warming_interval"`
	ConcurrentWorkers  int           `json:"concurrent_workers"`
	BatchSize          int           `json:"batch_size"`
	PredictionWindow   time.Duration `json:"prediction_window"`
	MinAccessThreshold int64         `json:"min_access_threshold"`
}

type CacheInvalidatorConfig struct {
	Strategies       []InvalidationStrategy `json:"strategies"`
	TTLCheckInterval time.Duration          `json:"ttl_check_interval"`
	MaxTagsPerKey    int                    `json:"max_tags_per_key"`
	BatchSize        int                    `json:"batch_size"`
}

type WarmingScheduler struct {
	queue    chan WarmingJob
	workers  []*WarmingWorker
	metrics  *CacheSchedulerMetrics
	ctx      context.Context
	cancel   context.CancelFunc
}

type WarmingExecutor struct {
	cache    *RedisClusterCache
	metrics  *ExecutorMetrics
	logger   Logger
}

type AccessPredictor struct {
	patterns map[string]*AccessPattern
	config   PredictorConfig
	mutex    sync.RWMutex
}

type WarmingJob struct {
	Key      string    `json:"key"`
	Priority int       `json:"priority"`
	Strategy string    `json:"strategy"`
	Created  time.Time `json:"created"`
}

type WarmingWorker struct {
	id       int
	executor *WarmingExecutor
	jobs     chan WarmingJob
	ctx      context.Context
	cancel   context.CancelFunc
}

type AccessPattern struct {
	Key           string    `json:"key"`
	AccessTimes   []time.Time `json:"access_times"`
	Frequency     float64   `json:"frequency"`
	LastAccess    time.Time `json:"last_access"`
	PredictedNext time.Time `json:"predicted_next"`
}

type CacheSchedulerMetrics struct {
	JobsQueued    int64 `json:"jobs_queued"`
	JobsProcessed int64 `json:"jobs_processed"`
	QueueDepth    int   `json:"queue_depth"`
}

type ExecutorMetrics struct {
	JobsExecuted  int64         `json:"jobs_executed"`
	ExecutionTime time.Duration `json:"execution_time"`
	SuccessRate   float64       `json:"success_rate"`
}

type PredictorConfig struct {
	WindowSize      int           `json:"window_size"`
	MinDataPoints   int           `json:"min_data_points"`
	PredictionHorizon time.Duration `json:"prediction_horizon"`
}

// DefaultRedisClusterCacheConfig returns default configuration
func DefaultRedisClusterCacheConfig() RedisClusterCacheConfig {
	return RedisClusterCacheConfig{
		PrimaryAddrs:       []string{"localhost:7000", "localhost:7001", "localhost:7002"},
		ReplicaAddrs:       []string{"localhost:7003", "localhost:7004", "localhost:7005"},
		MaxRedirects:       8,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PartitionStrategy:  PartitionByConsistent,
		PartitionCount:     256,
		ReplicationFactor:  3,
		ConsistentHashing:  true,
		DefaultTTL:         time.Hour,
		MaxValueSize:       1024 * 1024, // 1MB
		CompressionEnabled: true,
		CompressionThreshold: 1024, // 1KB
		InvalidationStrategy: InvalidationTTL,
		TagsEnabled:         true,
		MaxTags:            10,
		WarmingEnabled:     true,
		WarmingInterval:    10 * time.Minute,
		WarmingConcurrency: 5,
		WarmingBatchSize:   100,
		PipelineSize:       100,
		PoolSize:          20,
		MinIdleConns:      5,
		ReadPreference:    ReadPreferReplica,
		EnableMetrics:     true,
		MetricsInterval:   30 * time.Second,
		EnableTracing:     false,
	}
}

// NewRedisClusterCache creates a new Redis cluster cache
func NewRedisClusterCache(config RedisClusterCacheConfig, logger Logger) (*RedisClusterCache, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create primary cluster client
	primary := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.PrimaryAddrs,
		Password:     config.Password,
		MaxRedirects: config.MaxRedirects,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})
	
	// Test primary connection
	if err := primary.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to primary cluster: %w", err)
	}
	
	// Create replica clients
	var replicas []*redis.ClusterClient
	for _, addr := range config.ReplicaAddrs {
		replica := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        []string{addr},
			Password:     config.Password,
			ReadOnly:     true,
			RouteRandomly: true,
			ReadTimeout:  config.ReadTimeout,
			PoolSize:     config.PoolSize / 2,
			MinIdleConns: config.MinIdleConns / 2,
		})
		replicas = append(replicas, replica)
	}
	
	cache := &RedisClusterCache{
		config:   config,
		primary:  primary,
		replicas: replicas,
		metrics:  &CacheMetrics{},
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
	
	// Initialize partitioner
	cache.partitioner = NewCachePartitioner(CachePartitionerConfig{
		RebalanceThreshold:  0.2,
		RebalanceInterval:   5 * time.Minute,
		MigrationBatchSize:  100,
		HealthCheckInterval: 30 * time.Second,
	}, config.PartitionStrategy, config.PartitionCount, logger)
	
	// Initialize cache warmer if enabled
	if config.WarmingEnabled {
		cache.warmer = NewCacheWarmer(CacheWarmerConfig{
			Strategies:         []string{"lru", "frequency", "predictive"},
			WarmingInterval:    config.WarmingInterval,
			ConcurrentWorkers:  config.WarmingConcurrency,
			BatchSize:          config.WarmingBatchSize,
			PredictionWindow:   time.Hour,
			MinAccessThreshold: 5,
		}, cache, logger)
	}
	
	// Initialize invalidator
	cache.invalidator = NewCacheInvalidator(CacheInvalidatorConfig{
		Strategies:       []InvalidationStrategy{config.InvalidationStrategy},
		TTLCheckInterval: time.Minute,
		MaxTagsPerKey:    config.MaxTags,
		BatchSize:        1000,
	}, logger)
	
	return cache, nil
}

// Start starts the Redis cluster cache
func (c *RedisClusterCache) Start() error {
	c.logger.Info("Starting Redis cluster cache")
	
	// Start partitioner
	if err := c.partitioner.Start(); err != nil {
		return fmt.Errorf("failed to start partitioner: %w", err)
	}
	
	// Start cache warmer
	if c.warmer != nil {
		if err := c.warmer.Start(); err != nil {
			return fmt.Errorf("failed to start cache warmer: %w", err)
		}
	}
	
	// Start invalidator
	if err := c.invalidator.Start(); err != nil {
		return fmt.Errorf("failed to start invalidator: %w", err)
	}
	
	// Start metrics collection
	if c.config.EnableMetrics {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			c.metricsLoop()
		}()
	}
	
	c.logger.Info("Redis cluster cache started")
	return nil
}

// Stop stops the Redis cluster cache
func (c *RedisClusterCache) Stop() error {
	c.logger.Info("Stopping Redis cluster cache")
	
	c.cancel()
	
	// Stop components
	if c.warmer != nil {
		c.warmer.Stop()
	}
	c.invalidator.Stop()
	c.partitioner.Stop()
	
	// Close Redis connections
	c.primary.Close()
	for _, replica := range c.replicas {
		replica.Close()
	}
	
	c.wg.Wait()
	
	c.logger.Info("Redis cluster cache stopped")
	return nil
}

// Get retrieves a value from the cache
func (c *RedisClusterCache) Get(key string) (interface{}, error) {
	start := time.Now()
	atomic.AddInt64(&c.metrics.Hits, 1)
	
	// Choose client based on read preference
	client := c.getReadClient()
	
	// Get partition for key
	partition := c.partitioner.GetPartition(key)
	partitionKey := c.getPartitionKey(key, partition.ID)
	
	// Retrieve from Redis
	data, err := client.Get(c.ctx, partitionKey).Result()
	if err != nil {
		if err == redis.Nil {
			atomic.AddInt64(&c.metrics.Misses, 1)
			return nil, nil
		}
		return nil, fmt.Errorf("cache get failed: %w", err)
	}
	
	// Deserialize entry
	var entry CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, fmt.Errorf("failed to deserialize cache entry: %w", err)
	}
	
	// Update access tracking
	entry.AccessedAt = time.Now()
	atomic.AddInt64(&entry.AccessCount, 1)
	
	// Update access pattern for warming
	if c.warmer != nil {
		c.warmer.predictor.RecordAccess(key)
	}
	
	c.updateLatencyMetrics(time.Since(start))
	
	return entry.Value, nil
}

// Set stores a value in the cache
func (c *RedisClusterCache) Set(key string, value interface{}, ttl time.Duration, tags ...string) error {
	start := time.Now()
	atomic.AddInt64(&c.metrics.Sets, 1)
	
	// Create cache entry
	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		TTL:         ttl,
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		AccessCount: 1,
		Tags:        tags,
		Metadata:    make(map[string]interface{}),
	}
	
	// Serialize entry
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to serialize cache entry: %w", err)
	}
	
	entry.Size = int64(len(data))
	
	// Check if compression is needed
	if c.config.CompressionEnabled && entry.Size > c.config.CompressionThreshold {
		// TODO: Implement compression
		entry.Compressed = true
	}
	
	// Check value size limit
	if entry.Size > c.config.MaxValueSize {
		return fmt.Errorf("value size %d exceeds maximum %d", entry.Size, c.config.MaxValueSize)
	}
	
	// Get partition for key
	partition := c.partitioner.GetPartition(key)
	partitionKey := c.getPartitionKey(key, partition.ID)
	
	// Store in Redis
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}
	
	err = c.primary.Set(c.ctx, partitionKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("cache set failed: %w", err)
	}
	
	// Update tag index
	if c.config.TagsEnabled && len(tags) > 0 {
		c.invalidator.tagIndex.AddTags(key, tags)
	}
	
	c.updateLatencyMetrics(time.Since(start))
	
	return nil
}

// Delete removes a value from the cache
func (c *RedisClusterCache) Delete(key string) error {
	atomic.AddInt64(&c.metrics.Deletes, 1)
	
	// Get partition for key
	partition := c.partitioner.GetPartition(key)
	partitionKey := c.getPartitionKey(key, partition.ID)
	
	// Delete from Redis
	err := c.primary.Del(c.ctx, partitionKey).Err()
	if err != nil {
		return fmt.Errorf("cache delete failed: %w", err)
	}
	
	// Remove from tag index
	if c.config.TagsEnabled {
		c.invalidator.tagIndex.RemoveKey(key)
	}
	
	return nil
}

// InvalidateByTags invalidates all entries with the specified tags
func (c *RedisClusterCache) InvalidateByTags(tags []string) error {
	if !c.config.TagsEnabled {
		return fmt.Errorf("tags not enabled")
	}
	
	return c.invalidator.InvalidateByTags(tags)
}

// GetMetrics returns cache metrics
func (c *RedisClusterCache) GetMetrics() *CacheMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// Calculate hit ratio
	total := c.metrics.Hits + c.metrics.Misses
	if total > 0 {
		c.metrics.HitRatio = float64(c.metrics.Hits) / float64(total)
	}
	
	// Get partition metrics
	c.metrics.PartitionMetrics = c.partitioner.GetPartitionMetrics()
	
	return c.metrics
}

// Private methods

// getReadClient returns the appropriate Redis client for read operations
func (c *RedisClusterCache) getReadClient() redis.Cmdable {
	switch c.config.ReadPreference {
	case ReadReplica, ReadPreferReplica:
		if len(c.replicas) > 0 {
			// Round-robin selection of replica
			idx := int(atomic.LoadInt64(&c.metrics.Hits)) % len(c.replicas)
			return c.replicas[idx]
		}
		fallthrough
	case ReadPrimary:
		return c.primary
	case ReadNearest:
		// TODO: Implement latency-based selection
		return c.primary
	default:
		return c.primary
	}
}

// getPartitionKey generates a partitioned key
func (c *RedisClusterCache) getPartitionKey(key string, partitionID int) string {
	return fmt.Sprintf("cache:p%d:%s", partitionID, key)
}

// updateLatencyMetrics updates latency metrics
func (c *RedisClusterCache) updateLatencyMetrics(latency time.Duration) {
	if c.metrics.AverageLatency == 0 {
		c.metrics.AverageLatency = latency
	} else {
		c.metrics.AverageLatency = (c.metrics.AverageLatency + latency) / 2
	}
}

// metricsLoop periodically updates metrics
func (c *RedisClusterCache) metricsLoop() {
	ticker := time.NewTicker(c.config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.updateMetrics()
		case <-c.ctx.Done():
			return
		}
	}
}

// updateMetrics updates cache metrics
func (c *RedisClusterCache) updateMetrics() {
	// Count total keys across all partitions
	var totalKeys int64
	// TODO: Parse cluster info to get key counts
	// This is a simplified implementation
	c.metrics.TotalKeys = totalKeys
}

// Placeholder implementations for referenced components

func NewCachePartitioner(config CachePartitionerConfig, strategy PartitionStrategy, count int, logger Logger) *CachePartitioner {
	return &CachePartitioner{
		strategy:   strategy,
		partitions: make([]CachePartition, count),
		metrics:    &PartitionerMetrics{},
	}
}

func (cp *CachePartitioner) Start() error { return nil }
func (cp *CachePartitioner) Stop() error  { return nil }
func (cp *CachePartitioner) GetPartition(key string) *CachePartition {
	// Simple hash-based partitioning
	hash := crc32.ChecksumIEEE([]byte(key))
	idx := int(hash) % len(cp.partitions)
	return &cp.partitions[idx]
}
func (cp *CachePartitioner) GetPartitionMetrics() []*PartitionMetrics {
	return []*PartitionMetrics{}
}

func NewCacheWarmer(config CacheWarmerConfig, cache *RedisClusterCache, logger Logger) *CacheWarmer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CacheWarmer{
		config: config,
		strategies: make(map[string]WarmingStrategy),
		metrics: &WarmerMetrics{},
		logger: logger,
		ctx: ctx,
		cancel: cancel,
		predictor: &AccessPredictor{
			patterns: make(map[string]*AccessPattern),
		},
	}
}

func (cw *CacheWarmer) Start() error { return nil }
func (cw *CacheWarmer) Stop() error  { 
	cw.cancel()
	cw.wg.Wait()
	return nil 
}

func (ap *AccessPredictor) RecordAccess(key string) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	
	if pattern, exists := ap.patterns[key]; exists {
		pattern.AccessTimes = append(pattern.AccessTimes, time.Now())
		pattern.LastAccess = time.Now()
	} else {
		ap.patterns[key] = &AccessPattern{
			Key:         key,
			AccessTimes: []time.Time{time.Now()},
			LastAccess:  time.Now(),
		}
	}
}

func NewCacheInvalidator(config CacheInvalidatorConfig, logger Logger) *CacheInvalidator {
	return &CacheInvalidator{
		config: config,
		strategies: make(map[InvalidationStrategy]InvalidationHandler),
		tagIndex: &TagIndex{
			tagToKeys: make(map[string]map[string]struct{}),
			keyToTags: make(map[string]map[string]struct{}),
		},
		callbacks: make(map[string]InvalidationCallback),
		metrics: &InvalidatorMetrics{},
		logger: logger,
	}
}

func (ci *CacheInvalidator) Start() error { return nil }
func (ci *CacheInvalidator) Stop() error  { return nil }
func (ci *CacheInvalidator) InvalidateByTags(tags []string) error { return nil }

func (ti *TagIndex) AddTags(key string, tags []string) {
	ti.mutex.Lock()
	defer ti.mutex.Unlock()
	
	if ti.keyToTags[key] == nil {
		ti.keyToTags[key] = make(map[string]struct{})
	}
	
	for _, tag := range tags {
		ti.keyToTags[key][tag] = struct{}{}
		
		if ti.tagToKeys[tag] == nil {
			ti.tagToKeys[tag] = make(map[string]struct{})
		}
		ti.tagToKeys[tag][key] = struct{}{}
	}
}

func (ti *TagIndex) RemoveKey(key string) {
	ti.mutex.Lock()
	defer ti.mutex.Unlock()
	
	if tags, exists := ti.keyToTags[key]; exists {
		for tag := range tags {
			if keys, exists := ti.tagToKeys[tag]; exists {
				delete(keys, key)
				if len(keys) == 0 {
					delete(ti.tagToKeys, tag)
				}
			}
		}
		delete(ti.keyToTags, key)
	}
}