package optimization

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"
)

// MemoryOptimizer handles memory optimization for template processing
type MemoryOptimizer struct {
	maxMemoryMB   int64
	gcInterval    time.Duration
	poolSize      int
	templatePools map[string]*sync.Pool
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	stats         OptimizationStats
}

// OptimizationStats tracks memory optimization statistics
type OptimizationStats struct {
	GCRuns           int64
	MemoryFreed      int64
	PoolHits         int64
	PoolMisses       int64
	OptimizationTime time.Duration
}

// TemplateCache represents cached template data
type TemplateCache struct {
	Data       interface{}
	Size       int64
	LastAccess time.Time
	RefCount   int32
}

// MemoryConfig configuration for memory optimizer
type MemoryConfig struct {
	MaxMemoryMB   int64
	GCInterval    time.Duration
	PoolSize      int
	CacheTimeout  time.Duration
	EnablePooling bool
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(config MemoryConfig) *MemoryOptimizer {
	ctx, cancel := context.WithCancel(context.Background())

	mo := &MemoryOptimizer{
		maxMemoryMB:   config.MaxMemoryMB,
		gcInterval:    config.GCInterval,
		poolSize:      config.PoolSize,
		templatePools: make(map[string]*sync.Pool),
		ctx:           ctx,
		cancel:        cancel,
		stats:         OptimizationStats{},
	}

	if config.EnablePooling {
		mo.initializePools()
	}

	// Start background optimization routine
	go mo.backgroundOptimization()

	return mo
}

// initializePools initializes object pools for different template types
func (mo *MemoryOptimizer) initializePools() {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	// Template data pool
	mo.templatePools["data"] = &sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{})
		},
	}

	// String builder pool
	mo.templatePools["builder"] = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024)
		},
	}

	// Cache entry pool
	mo.templatePools["cache"] = &sync.Pool{
		New: func() interface{} {
			return &TemplateCache{
				LastAccess: time.Now(),
				RefCount:   0,
			}
		},
	}
}

// GetFromPool retrieves an object from the specified pool
func (mo *MemoryOptimizer) GetFromPool(poolName string) interface{} {
	mo.mu.RLock()
	pool, exists := mo.templatePools[poolName]
	mo.mu.RUnlock()

	if !exists {
		mo.stats.PoolMisses++
		return nil
	}

	mo.stats.PoolHits++
	return pool.Get()
}

// PutToPool returns an object to the specified pool
func (mo *MemoryOptimizer) PutToPool(poolName string, obj interface{}) {
	mo.mu.RLock()
	pool, exists := mo.templatePools[poolName]
	mo.mu.RUnlock()

	if exists {
		pool.Put(obj)
	}
}

// OptimizeMemory performs memory optimization
func (mo *MemoryOptimizer) OptimizeMemory(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		mo.stats.OptimizationTime += time.Since(startTime)
	}()

	// Get current memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := m.Alloc / 1024 / 1024
	currentMemoryMB := int64(min(allocMB, uint64(math.MaxInt64)))

	if currentMemoryMB > mo.maxMemoryMB {
		// Force garbage collection
		beforeGC := m.Alloc
		runtime.GC()

		// Update stats
		runtime.ReadMemStats(&m)
		mo.stats.GCRuns++
		freedBytes := beforeGC - m.Alloc
		mo.stats.MemoryFreed += int64(min(freedBytes, uint64(math.MaxInt64)))
	}

	return nil
}

// backgroundOptimization runs continuous memory optimization
func (mo *MemoryOptimizer) backgroundOptimization() {
	ticker := time.NewTicker(mo.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = mo.OptimizeMemory(mo.ctx) // #nosec G104 -- background optimization errors are non-fatal
		case <-mo.ctx.Done():
			return
		}
	}
}

// GetStats returns current optimization statistics
func (mo *MemoryOptimizer) GetStats() OptimizationStats {
	return mo.stats
}

// GetMemoryUsage returns current memory usage in MB
func (mo *MemoryOptimizer) GetMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocMB := m.Alloc / 1024 / 1024
	return int64(min(allocMB, uint64(math.MaxInt64)))
}

// SetMaxMemory updates the maximum memory threshold
func (mo *MemoryOptimizer) SetMaxMemory(maxMemoryMB int64) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.maxMemoryMB = maxMemoryMB
}

// Close shuts down the memory optimizer
func (mo *MemoryOptimizer) Close() error {
	mo.cancel()

	// Clear all pools
	mo.mu.Lock()
	defer mo.mu.Unlock()

	for name := range mo.templatePools {
		delete(mo.templatePools, name)
	}

	return nil
}

// ClearPools clears all object pools
func (mo *MemoryOptimizer) ClearPools() {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	// Recreate pools to clear them
	mo.initializePools()
}

// ResizePool resizes a specific pool
func (mo *MemoryOptimizer) ResizePool(poolName string, newSize int) error {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	if _, exists := mo.templatePools[poolName]; !exists {
		return nil // Pool doesn't exist
	}

	// Create new pool with updated size
	// Note: sync.Pool doesn't have a direct way to resize,
	// so we recreate the pool
	switch poolName {
	case "data":
		mo.templatePools["data"] = &sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{}, newSize)
			},
		}
	case "builder":
		mo.templatePools["builder"] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, newSize)
			},
		}
	}

	return nil
}
