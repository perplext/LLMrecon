package resource

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

// ResourcePoolManager manages resource pools for efficient resource utilization
type ResourcePoolManager struct {
	// pools is a map of pool name to pool
	pools map[string]*ResourcePool
	// mutex protects the pools map
	mutex sync.RWMutex
	// stats tracks resource usage statistics
	stats *PoolManagerStats
	// config contains the manager configuration
	config *PoolManagerConfig

// ResourcePool represents a pool of resources
type ResourcePool struct {
	// name is the name of the pool
	name string
	// size is the size of the pool
	size int
	// available is the number of available resources
	available int32
	// resources is a channel of resources
	resources chan interface{}
	// inUse tracks resources that are currently in use
	inUse map[interface{}]time.Time
	// mutex protects the inUse map
	mutex sync.RWMutex
	// factory is a function that creates new resources
	factory func() (interface{}, error)
	// cleanup is a function that cleans up resources
	cleanup func(interface{})
	// validate is a function that validates resources
	validate func(interface{}) bool
	// maxIdleTime is the maximum time a resource can be idle
	maxIdleTime time.Duration
	// maxLifetime is the maximum lifetime of a resource
	maxLifetime time.Duration
	// createdAt tracks when resources were created
	createdAt map[interface{}]time.Time
	// lastUsedAt tracks when resources were last used
	lastUsedAt map[interface{}]time.Time
	// stats tracks pool statistics
	stats *PoolStats

// PoolManagerStats tracks statistics for the resource pool manager
type PoolManagerStats struct {
	// TotalPools is the total number of pools
	TotalPools int
	// TotalResources is the total number of resources across all pools
	TotalResources int
	// TotalAvailable is the total number of available resources across all pools
	TotalAvailable int
	// TotalInUse is the total number of resources in use across all pools
	TotalInUse int
	// PoolStats is a map of pool name to pool statistics
	PoolStats map[string]*PoolStats

// PoolStats tracks statistics for a resource pool
type PoolStats struct {
	// Size is the size of the pool
	Size int
	// Available is the number of available resources
	Available int
	// InUse is the number of resources in use
	InUse int
	// Created is the number of resources created
	Created int64
	// Acquired is the number of resources acquired
	Acquired int64
	// Released is the number of resources released
	Released int64
	// Errors is the number of errors
	Errors int64
	// Timeouts is the number of timeouts
	Timeouts int64
	// MaxWaitTime is the maximum wait time for a resource
	MaxWaitTime time.Duration
	// TotalWaitTime is the total wait time for resources
	TotalWaitTime time.Duration
	// AvgWaitTime is the average wait time for a resource
	AvgWaitTime time.Duration

// PoolManagerConfig represents configuration for the resource pool manager
type PoolManagerConfig struct {
	// DefaultPoolSize is the default size of pools
	DefaultPoolSize int
	// DefaultMaxIdleTime is the default maximum idle time for resources
	DefaultMaxIdleTime time.Duration
	// DefaultMaxLifetime is the default maximum lifetime for resources
	DefaultMaxLifetime time.Duration
	// EnableResourceValidation enables resource validation
	EnableResourceValidation bool
	// EnableResourceCleanup enables resource cleanup
	EnableResourceCleanup bool
	// EnablePoolScaling enables automatic pool scaling
	EnablePoolScaling bool
	// MinPoolSize is the minimum size of pools when scaling
	MinPoolSize int
	// MaxPoolSize is the maximum size of pools when scaling
	MaxPoolSize int
	// ScaleUpThreshold is the threshold for scaling up pools
	ScaleUpThreshold float64
	// ScaleDownThreshold is the threshold for scaling down pools
	ScaleDownThreshold float64
	// ScaleCheckInterval is the interval for checking if pools need to be scaled
	ScaleCheckInterval time.Duration

// DefaultPoolManagerConfig returns default configuration for the resource pool manager
func DefaultPoolManagerConfig() *PoolManagerConfig {
	return &PoolManagerConfig{
		DefaultPoolSize:         runtime.NumCPU() * 2,
		DefaultMaxIdleTime:      30 * time.Minute,
		DefaultMaxLifetime:      24 * time.Hour,
		EnableResourceValidation: true,
		EnableResourceCleanup:   true,
		EnablePoolScaling:       true,
		MinPoolSize:             1,
		MaxPoolSize:             runtime.NumCPU() * 4,
		ScaleUpThreshold:        0.8,  // 80% utilization
		ScaleDownThreshold:      0.2,  // 20% utilization
		ScaleCheckInterval:      1 * time.Minute,
	}

// NewResourcePoolManager creates a new resource pool manager
}
func NewResourcePoolManager(config *PoolManagerConfig) *ResourcePoolManager {
	if config == nil {
		config = DefaultPoolManagerConfig()
	}

	manager := &ResourcePoolManager{
		pools:  make(map[string]*ResourcePool),
		stats:  &PoolManagerStats{PoolStats: make(map[string]*PoolStats)},
		config: config,
	}

	// Start pool scaling if enabled
	if config.EnablePoolScaling {
		go manager.startPoolScaling()
	}

	return manager

// CreatePool creates a new resource pool
}
func (m *ResourcePoolManager) CreatePool(name string, size int, factory func() (interface{}, error), cleanup func(interface{})) (*ResourcePool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if pool already exists
	if _, exists := m.pools[name]; exists {
		return nil, fmt.Errorf("pool '%s' already exists", name)
	}

	// Use default size if not specified
	if size <= 0 {
		size = m.config.DefaultPoolSize
	}

	// Create pool
	pool := &ResourcePool{
		name:        name,
		size:        size,
		available:   int32(size),
		resources:   make(chan interface{}, size),
		inUse:       make(map[interface{}]time.Time),
		factory:     factory,
		cleanup:     cleanup,
		validate:    func(interface{}) bool { return true }, // Default validation always passes
		maxIdleTime: m.config.DefaultMaxIdleTime,
		maxLifetime: m.config.DefaultMaxLifetime,
		createdAt:   make(map[interface{}]time.Time),
		lastUsedAt:  make(map[interface{}]time.Time),
		stats: &PoolStats{
			Size:      size,
			Available: size,
		},
	}

	// Initialize pool with resources
	for i := 0; i < size; i++ {
		resource, err := factory()
		if err != nil {
			// Clean up already created resources
			for j := 0; j < i; j++ {
				if r := <-pool.resources; cleanup != nil {
					cleanup(r)
				}
			}
			return nil, fmt.Errorf("failed to create resource: %w", err)
		}

		now := time.Now()
		pool.createdAt[resource] = now
		pool.lastUsedAt[resource] = now
		pool.resources <- resource
		atomic.AddInt64(&pool.stats.Created, 1)
	}

	// Add pool to manager
	m.pools[name] = pool
	m.updateStats()

	return pool, nil

// GetPool gets a resource pool by name
}
func (m *ResourcePoolManager) GetPool(name string) (*ResourcePool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check if pool exists
	pool, exists := m.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool '%s' does not exist", name)
	}

	return pool, nil

// RemovePool removes a resource pool
}
func (m *ResourcePoolManager) RemovePool(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if pool exists
	pool, exists := m.pools[name]
	if !exists {
		return fmt.Errorf("pool '%s' does not exist", name)
	}

	// Close pool
	pool.Close()

	// Remove pool from manager
	delete(m.pools, name)
	delete(m.stats.PoolStats, name)
	m.updateStats()

	return nil

// ListPools lists all resource pools
}
func (m *ResourcePoolManager) ListPools() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create result slice
	result := make([]string, 0, len(m.pools))

	// Add pools to result
	for name := range m.pools {
		result = append(result, name)
	}

	return result

// GetStats returns statistics for the resource pool manager
}
func (m *ResourcePoolManager) GetStats() *PoolManagerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Update stats
	m.updateStats()

	return m.stats

// updateStats updates statistics for the resource pool manager
}
func (m *ResourcePoolManager) updateStats() {
	// Reset totals
	m.stats.TotalPools = len(m.pools)
	m.stats.TotalResources = 0
	m.stats.TotalAvailable = 0
	m.stats.TotalInUse = 0

	// Update pool stats
	for name, pool := range m.pools {
		// Update pool stats
		pool.mutex.RLock()
		pool.stats.Available = int(pool.available)
		pool.stats.InUse = len(pool.inUse)
		pool.mutex.RUnlock()

		// Update manager stats
		m.stats.PoolStats[name] = pool.stats
		m.stats.TotalResources += pool.stats.Size
		m.stats.TotalAvailable += pool.stats.Available
		m.stats.TotalInUse += pool.stats.InUse
	}

// startPoolScaling starts automatic pool scaling
}
func (m *ResourcePoolManager) startPoolScaling() {
	ticker := time.NewTicker(m.config.ScaleCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.scaleAllPools()
	}

// scaleAllPools scales all pools based on utilization
}
func (m *ResourcePoolManager) scaleAllPools() {
	m.mutex.RLock()
	poolNames := make([]string, 0, len(m.pools))
	for name := range m.pools {
		poolNames = append(poolNames, name)
	}
	m.mutex.RUnlock()

	// Scale each pool
	for _, name := range poolNames {
		m.scalePool(name)
	}

// scalePool scales a pool based on utilization
}
func (m *ResourcePoolManager) scalePool(name string) {
	m.mutex.RLock()
	pool, exists := m.pools[name]
	m.mutex.RUnlock()

	if !exists {
		return
	}

	// Calculate utilization
	pool.mutex.RLock()
	utilization := float64(len(pool.inUse)) / float64(pool.size)
	currentSize := pool.size
	pool.mutex.RUnlock()

	// Scale up if utilization is high
	if utilization >= m.config.ScaleUpThreshold && currentSize < m.config.MaxPoolSize {
		newSize := min(currentSize*2, m.config.MaxPoolSize)
		m.resizePool(name, newSize)
	}

	// Scale down if utilization is low
	if utilization <= m.config.ScaleDownThreshold && currentSize > m.config.MinPoolSize {
		newSize := max(currentSize/2, m.config.MinPoolSize)
		m.resizePool(name, newSize)
	}

// resizePool resizes a pool to the specified size
}
func (m *ResourcePoolManager) resizePool(name string, newSize int) {
	m.mutex.RLock()
	pool, exists := m.pools[name]
	m.mutex.RUnlock()

	if !exists {
		return
	}

	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// No change needed
	if newSize == pool.size {
		return
	}

	// Scaling up
	if newSize > pool.size {
		// Create new resources
		for i := 0; i < newSize-pool.size; i++ {
			resource, err := pool.factory()
			if err != nil {
				// Log error but continue
				atomic.AddInt64(&pool.stats.Errors, 1)
				continue
			}

			now := time.Now()
			pool.createdAt[resource] = now
			pool.lastUsedAt[resource] = now
			pool.resources <- resource
			atomic.AddInt32(&pool.available, 1)
			atomic.AddInt64(&pool.stats.Created, 1)
		}
	} else {
		// Scaling down
		// Remove excess resources
		for i := 0; i < pool.size-newSize; i++ {
			select {
			case resource := <-pool.resources:
				// Clean up resource
				if pool.cleanup != nil {
					pool.cleanup(resource)
				}
				delete(pool.createdAt, resource)
				delete(pool.lastUsedAt, resource)
				atomic.AddInt32(&pool.available, -1)
			default:
}
				// No more available resources to remove
				break
			}
		}
	}

	// Update pool size
	pool.size = newSize
	pool.stats.Size = newSize

// Acquire acquires a resource from a pool
func (p *ResourcePool) Acquire(ctx context.Context) (interface{}, error) {
	// Try to get a resource from the pool
	select {
	case resource := <-p.resources:
}
		// Check if resource is valid
		if p.validate != nil && !p.validate(resource) {
			// Resource is invalid, create a new one
			if p.cleanup != nil {
				p.cleanup(resource)
			}
			var err error
			resource, err = p.factory()
			if err != nil {
				atomic.AddInt64(&p.stats.Errors, 1)
				atomic.AddInt32(&p.available, -1)
				return nil, fmt.Errorf("failed to create resource: %w", err)
			}
			p.createdAt[resource] = time.Now()
			atomic.AddInt64(&p.stats.Created, 1)
		}

		// Check if resource has expired
		p.mutex.Lock()
		createdAt, exists := p.createdAt[resource]
		if exists && p.maxLifetime > 0 && time.Since(createdAt) > p.maxLifetime {
			// Resource has expired, create a new one
			if p.cleanup != nil {
				p.cleanup(resource)
			}
			var err error
			resource, err = p.factory()
			if err != nil {
				p.mutex.Unlock()
				atomic.AddInt64(&p.stats.Errors, 1)
				atomic.AddInt32(&p.available, -1)
				return nil, fmt.Errorf("failed to create resource: %w", err)
			}
			p.createdAt[resource] = time.Now()
			atomic.AddInt64(&p.stats.Created, 1)
		}

		// Mark resource as in use
		p.inUse[resource] = time.Now()
		p.lastUsedAt[resource] = time.Now()
		p.mutex.Unlock()

		atomic.AddInt32(&p.available, -1)
		atomic.AddInt64(&p.stats.Acquired, 1)

		return resource, nil
	case <-ctx.Done():
		// Context canceled or timed out
		atomic.AddInt64(&p.stats.Timeouts, 1)
		return nil, ctx.Err()
	}

// Release releases a resource back to the pool
func (p *ResourcePool) Release(resource interface{}) {
	p.mutex.Lock()
	// Check if resource is in use
	_, exists := p.inUse[resource]
	if !exists {
		p.mutex.Unlock()
		return
	}

	// Remove resource from in-use map
	delete(p.inUse, resource)
	p.lastUsedAt[resource] = time.Now()
	p.mutex.Unlock()

	// Return resource to pool
	select {
	case p.resources <- resource:
		atomic.AddInt32(&p.available, 1)
		atomic.AddInt64(&p.stats.Released, 1)
	default:
}
		// Pool is full, clean up resource
		if p.cleanup != nil {
			p.cleanup(resource)
		}
		p.mutex.Lock()
		delete(p.createdAt, resource)
		delete(p.lastUsedAt, resource)
		p.mutex.Unlock()
	}

// Close closes the resource pool
func (p *ResourcePool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Clean up all resources
	close(p.resources)
	for resource := range p.resources {
		if p.cleanup != nil {
			p.cleanup(resource)
		}
	}

	// Clean up in-use resources
	for resource := range p.inUse {
		if p.cleanup != nil {
			p.cleanup(resource)
		}
	}

	// Clear maps
	p.inUse = make(map[interface{}]time.Time)
	p.createdAt = make(map[interface{}]time.Time)
	p.lastUsedAt = make(map[interface{}]time.Time)

// SetValidateFunc sets the validation function for the pool
}
func (p *ResourcePool) SetValidateFunc(validate func(interface{}) bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.validate = validate

// SetMaxIdleTime sets the maximum idle time for resources
}
func (p *ResourcePool) SetMaxIdleTime(maxIdleTime time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.maxIdleTime = maxIdleTime

// SetMaxLifetime sets the maximum lifetime for resources
}
func (p *ResourcePool) SetMaxLifetime(maxLifetime time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.maxLifetime = maxLifetime

// GetStats returns statistics for the resource pool
}
func (p *ResourcePool) GetStats() *PoolStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Update stats
	p.stats.Available = int(p.available)
	p.stats.InUse = len(p.inUse)

	// Calculate average wait time
	if p.stats.Acquired > 0 {
		p.stats.AvgWaitTime = time.Duration(p.stats.TotalWaitTime.Nanoseconds() / p.stats.Acquired)
	}

	return p.stats

// GetSize returns the size of the pool
}
func (p *ResourcePool) GetSize() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.size

// GetAvailable returns the number of available resources
}
func (p *ResourcePool) GetAvailable() int {
	return int(atomic.LoadInt32(&p.available))

// GetInUse returns the number of resources in use
}
func (p *ResourcePool) GetInUse() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.inUse)

// GetName returns the name of the pool
}
func (p *ResourcePool) GetName() string {
	return p.name

// min returns the minimum of two integers
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b

// max returns the maximum of two integers
}
func max(a, b int) int {
	if a > b {
		return a
	}
}
