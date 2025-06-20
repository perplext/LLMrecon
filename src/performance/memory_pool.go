package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryPoolManager manages object pools for memory optimization
type MemoryPoolManager struct {
	pools   map[string]*ObjectPool
	config  MemoryPoolConfig
	logger  Logger
	metrics *MemoryPoolMetrics
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// MemoryPoolConfig defines configuration for memory pools
type MemoryPoolConfig struct {
	// Pool sizing
	DefaultPoolSize    int           `json:"default_pool_size"`
	MaxPoolSize        int           `json:"max_pool_size"`
	MinPoolSize        int           `json:"min_pool_size"`
	
	// Object lifecycle
	MaxObjectAge       time.Duration `json:"max_object_age"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	PreallocationSize  int           `json:"preallocation_size"`
	
	// Memory management
	EnableGCOptimization bool          `json:"enable_gc_optimization"`
	GCTarget            int           `json:"gc_target"`
	MaxMemoryUsage      int64         `json:"max_memory_usage"`
	
	// Monitoring
	EnableMetrics      bool          `json:"enable_metrics"`
	MetricsInterval    time.Duration `json:"metrics_interval"`
}

// ObjectPool manages a pool of reusable objects
type ObjectPool struct {
	name        string
	factory     ObjectFactory
	reset       ObjectReset
	pool        chan PooledObject
	created     int64
	reused      int64
	maxAge      time.Duration
	metrics     *ObjectPoolMetrics
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// PooledObject represents an object in the pool
type PooledObject struct {
	Object    interface{} `json:"object"`
	CreatedAt time.Time   `json:"created_at"`
	UsedCount int64       `json:"used_count"`
	LastUsed  time.Time   `json:"last_used"`
}

// ObjectFactory creates new objects for the pool
type ObjectFactory interface {
	CreateObject() interface{}
	GetObjectType() string
}

// ObjectReset resets objects for reuse
type ObjectReset interface {
	ResetObject(obj interface{}) error
}

// MemoryPoolMetrics tracks memory pool performance
type MemoryPoolMetrics struct {
	TotalPools        int64 `json:"total_pools"`
	TotalObjects      int64 `json:"total_objects"`
	ObjectsCreated    int64 `json:"objects_created"`
	ObjectsReused     int64 `json:"objects_reused"`
	ObjectsDestroyed  int64 `json:"objects_destroyed"`
	MemoryAllocated   int64 `json:"memory_allocated"`
	MemoryReleased    int64 `json:"memory_released"`
	PoolHitRate       float64 `json:"pool_hit_rate"`
	AverageObjectAge  time.Duration `json:"average_object_age"`
}

// ObjectPoolMetrics tracks individual pool performance
type ObjectPoolMetrics struct {
	PoolName          string        `json:"pool_name"`
	PoolSize          int           `json:"pool_size"`
	ObjectsCreated    int64         `json:"objects_created"`
	ObjectsReused     int64         `json:"objects_reused"`
	ObjectsDestroyed  int64         `json:"objects_destroyed"`
	HitRate           float64       `json:"hit_rate"`
	AverageAge        time.Duration `json:"average_age"`
	LastCleanup       time.Time     `json:"last_cleanup"`
}

// Logger interface for memory pool logging
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// DefaultMemoryPoolConfig returns default configuration
func DefaultMemoryPoolConfig() MemoryPoolConfig {
	return MemoryPoolConfig{
		DefaultPoolSize:      100,
		MaxPoolSize:         1000,
		MinPoolSize:         10,
		MaxObjectAge:        30 * time.Minute,
		CleanupInterval:     5 * time.Minute,
		PreallocationSize:   50,
		EnableGCOptimization: true,
		GCTarget:            100,
		MaxMemoryUsage:      100 * 1024 * 1024, // 100MB
		EnableMetrics:       true,
		MetricsInterval:     30 * time.Second,
	}
}

// NewMemoryPoolManager creates a new memory pool manager
func NewMemoryPoolManager(config MemoryPoolConfig, logger Logger) *MemoryPoolManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &MemoryPoolManager{
		pools:   make(map[string]*ObjectPool),
		config:  config,
		logger:  logger,
		metrics: &MemoryPoolMetrics{},
		ctx:     ctx,
		cancel:  cancel,
	}
	
	return manager
}

// Start starts the memory pool manager
func (m *MemoryPoolManager) Start() error {
	m.logger.Info("Starting memory pool manager")
	
	// Start GC optimization if enabled
	if m.config.EnableGCOptimization {
		m.optimizeGC()
	}
	
	// Start cleanup loop
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.cleanupLoop()
	}()
	
	// Start metrics collection if enabled
	if m.config.EnableMetrics {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.metricsLoop()
		}()
	}
	
	m.logger.Info("Memory pool manager started")
	return nil
}

// Stop stops the memory pool manager
func (m *MemoryPoolManager) Stop() error {
	m.logger.Info("Stopping memory pool manager")
	
	m.cancel()
	
	// Stop all pools
	m.mutex.Lock()
	for _, pool := range m.pools {
		pool.Stop()
	}
	m.mutex.Unlock()
	
	m.wg.Wait()
	
	m.logger.Info("Memory pool manager stopped")
	return nil
}

// CreatePool creates a new object pool
func (m *MemoryPoolManager) CreatePool(name string, factory ObjectFactory, reset ObjectReset) (*ObjectPool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.pools[name]; exists {
		return nil, fmt.Errorf("pool %s already exists", name)
	}
	
	pool := &ObjectPool{
		name:    name,
		factory: factory,
		reset:   reset,
		pool:    make(chan PooledObject, m.config.DefaultPoolSize),
		maxAge:  m.config.MaxObjectAge,
		metrics: &ObjectPoolMetrics{
			PoolName: name,
		},
	}
	
	pool.ctx, pool.cancel = context.WithCancel(m.ctx)
	
	// Pre-allocate objects if configured
	if m.config.PreallocationSize > 0 {
		for i := 0; i < m.config.PreallocationSize && i < m.config.DefaultPoolSize; i++ {
			obj := factory.CreateObject()
			pooledObj := PooledObject{
				Object:    obj,
				CreatedAt: time.Now(),
				UsedCount: 0,
			}
			
			select {
			case pool.pool <- pooledObj:
				atomic.AddInt64(&pool.created, 1)
			default:
				break // Pool is full
			}
		}
	}
	
	m.pools[name] = pool
	m.metrics.TotalPools++
	
	m.logger.Info("Created object pool", "name", name, "type", factory.GetObjectType(), "prealloc", m.config.PreallocationSize)
	return pool, nil
}

// GetPool returns an object pool by name
func (m *MemoryPoolManager) GetPool(name string) (*ObjectPool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	pool, exists := m.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", name)
	}
	
	return pool, nil
}

// Get retrieves an object from a pool
func (m *MemoryPoolManager) Get(poolName string) (interface{}, error) {
	pool, err := m.GetPool(poolName)
	if err != nil {
		return nil, err
	}
	
	return pool.Get()
}

// Put returns an object to a pool
func (m *MemoryPoolManager) Put(poolName string, obj interface{}) error {
	pool, err := m.GetPool(poolName)
	if err != nil {
		return err
	}
	
	return pool.Put(obj)
}

// GetMetrics returns current memory pool metrics
func (m *MemoryPoolManager) GetMetrics() *MemoryPoolMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Update aggregate metrics
	var totalObjects, objectsCreated, objectsReused int64
	for _, pool := range m.pools {
		poolMetrics := pool.GetMetrics()
		totalObjects += int64(poolMetrics.PoolSize)
		objectsCreated += poolMetrics.ObjectsCreated
		objectsReused += poolMetrics.ObjectsReused
	}
	
	m.metrics.TotalObjects = totalObjects
	m.metrics.ObjectsCreated = objectsCreated
	m.metrics.ObjectsReused = objectsReused
	
	// Calculate hit rate
	if objectsCreated+objectsReused > 0 {
		m.metrics.PoolHitRate = float64(objectsReused) / float64(objectsCreated+objectsReused)
	}
	
	return m.metrics
}

// GetPoolMetrics returns metrics for all pools
func (m *MemoryPoolManager) GetPoolMetrics() []*ObjectPoolMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	metrics := make([]*ObjectPoolMetrics, 0, len(m.pools))
	for _, pool := range m.pools {
		metrics = append(metrics, pool.GetMetrics())
	}
	
	return metrics
}

// ObjectPool methods

// Get retrieves an object from the pool
func (p *ObjectPool) Get() (interface{}, error) {
	select {
	case pooledObj := <-p.pool:
		// Check if object is too old
		if time.Since(pooledObj.CreatedAt) > p.maxAge {
			// Object is too old, create a new one
			obj := p.factory.CreateObject()
			atomic.AddInt64(&p.created, 1)
			p.metrics.ObjectsCreated++
			return obj, nil
		}
		
		// Update usage statistics
		pooledObj.UsedCount++
		pooledObj.LastUsed = time.Now()
		atomic.AddInt64(&p.reused, 1)
		p.metrics.ObjectsReused++
		
		return pooledObj.Object, nil
		
	default:
		// Pool is empty, create new object
		obj := p.factory.CreateObject()
		atomic.AddInt64(&p.created, 1)
		p.metrics.ObjectsCreated++
		return obj, nil
	}
}

// Put returns an object to the pool
func (p *ObjectPool) Put(obj interface{}) error {
	// Reset the object if reset function is provided
	if p.reset != nil {
		if err := p.reset.ResetObject(obj); err != nil {
			return fmt.Errorf("failed to reset object: %w", err)
		}
	}
	
	pooledObj := PooledObject{
		Object:    obj,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}
	
	select {
	case p.pool <- pooledObj:
		return nil
	default:
		// Pool is full, discard the object
		return nil
	}
}

// Stop stops the object pool
func (p *ObjectPool) Stop() {
	p.cancel()
	
	// Drain the pool
	for {
		select {
		case <-p.pool:
			// Discard object
		default:
			return
		}
	}
}

// GetMetrics returns pool metrics
func (p *ObjectPool) GetMetrics() *ObjectPoolMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	p.metrics.PoolSize = len(p.pool)
	
	// Calculate hit rate
	total := p.metrics.ObjectsCreated + p.metrics.ObjectsReused
	if total > 0 {
		p.metrics.HitRate = float64(p.metrics.ObjectsReused) / float64(total)
	}
	
	return p.metrics
}

// Private methods

// cleanupLoop performs periodic cleanup of old objects
func (m *MemoryPoolManager) cleanupLoop() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.performCleanup()
		case <-m.ctx.Done():
			return
		}
	}
}

// performCleanup removes old objects from pools
func (m *MemoryPoolManager) performCleanup() {
	m.mutex.RLock()
	pools := make([]*ObjectPool, 0, len(m.pools))
	for _, pool := range m.pools {
		pools = append(pools, pool)
	}
	m.mutex.RUnlock()
	
	for _, pool := range pools {
		pool.cleanup()
	}
}

// cleanup removes old objects from this pool
func (p *ObjectPool) cleanup() {
	cleaned := 0
	poolSize := len(p.pool)
	
	for i := 0; i < poolSize; i++ {
		select {
		case pooledObj := <-p.pool:
			// Check if object is too old
			if time.Since(pooledObj.CreatedAt) <= p.maxAge {
				// Object is still good, put it back
				select {
				case p.pool <- pooledObj:
				default:
					// Pool is full, discard
					cleaned++
				}
			} else {
				// Object is too old, discard it
				cleaned++
				p.metrics.ObjectsDestroyed++
			}
		default:
			// No more objects in pool
			break
		}
	}
	
	if cleaned > 0 {
		p.metrics.LastCleanup = time.Now()
	}
}

// metricsLoop periodically updates metrics
func (m *MemoryPoolManager) metricsLoop() {
	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.updateMetrics()
		case <-m.ctx.Done():
			return
		}
	}
}

// updateMetrics updates memory and performance metrics
func (m *MemoryPoolManager) updateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	m.metrics.MemoryAllocated = int64(memStats.Alloc)
	
	// Force GC if memory usage is too high
	if m.config.MaxMemoryUsage > 0 && int64(memStats.Alloc) > m.config.MaxMemoryUsage {
		runtime.GC()
		m.logger.Warn("Forced garbage collection due to high memory usage", "allocated", memStats.Alloc, "limit", m.config.MaxMemoryUsage)
	}
}

// optimizeGC optimizes garbage collection settings
func (m *MemoryPoolManager) optimizeGC() {
	if m.config.GCTarget > 0 {
		runtime.GC()
		runtime.SetGCPercent(m.config.GCTarget)
		m.logger.Info("Optimized GC settings", "gc_target", m.config.GCTarget)
	}
}

// Common object factories and resetters

// ByteSliceFactory creates byte slice pools
type ByteSliceFactory struct {
	Size int
}

func NewByteSliceFactory(size int) *ByteSliceFactory {
	return &ByteSliceFactory{Size: size}
}

func (f *ByteSliceFactory) CreateObject() interface{} {
	return make([]byte, 0, f.Size)
}

func (f *ByteSliceFactory) GetObjectType() string {
	return "[]byte"
}

// ByteSliceReset resets byte slices
type ByteSliceReset struct{}

func (r *ByteSliceReset) ResetObject(obj interface{}) error {
	if slice, ok := obj.([]byte); ok {
		// Reset slice to zero length but keep capacity
		slice = slice[:0]
		return nil
	}
	return fmt.Errorf("object is not a []byte")
}

// StringBuilderFactory creates string builder pools
type StringBuilderFactory struct{}

func (f *StringBuilderFactory) CreateObject() interface{} {
	return &strings.Builder{}
}

func (f *StringBuilderFactory) GetObjectType() string {
	return "*strings.Builder"
}

// StringBuilderReset resets string builders
type StringBuilderReset struct{}

func (r *StringBuilderReset) ResetObject(obj interface{}) error {
	if builder, ok := obj.(*strings.Builder); ok {
		builder.Reset()
		return nil
	}
	return fmt.Errorf("object is not a *strings.Builder")
}

// MapFactory creates map pools
type MapFactory struct {
	InitialSize int
}

func NewMapFactory(initialSize int) *MapFactory {
	return &MapFactory{InitialSize: initialSize}
}

func (f *MapFactory) CreateObject() interface{} {
	return make(map[string]interface{}, f.InitialSize)
}

func (f *MapFactory) GetObjectType() string {
	return "map[string]interface{}"
}

// MapReset resets maps
type MapReset struct{}

func (r *MapReset) ResetObject(obj interface{}) error {
	if m, ok := obj.(map[string]interface{}); ok {
		// Clear the map
		for k := range m {
			delete(m, k)
		}
		return nil
	}
	return fmt.Errorf("object is not a map[string]interface{}")
}