package monitoring

import (
	"fmt"
	"runtime"
	"sync"
)

// MetricType defines the type of metric
type MetricType string

const (
	// CounterMetric is a cumulative metric that only increases
	CounterMetric MetricType = "counter"
	// GaugeMetric is a metric that can increase and decrease
	GaugeMetric MetricType = "gauge"
	// HistogramMetric is a metric that samples observations and counts them in configurable buckets
	HistogramMetric MetricType = "histogram"
)

// MetricsManager manages metrics collection and reporting
type MetricsManager struct {
	metrics     map[string]*Metric
	mutex       sync.RWMutex
	subscribers []MetricsSubscriber
}

// Metric represents a single metric
type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      map[string]string
	LastUpdated time.Time
	Buckets     map[float64]int // For histogram metrics
	Sum         float64         // For histogram metrics
	Count       int             // For histogram metrics
}

// MetricsSubscriber is an interface for components that want to be notified of metric updates
type MetricsSubscriber interface {
	OnMetricUpdate(metric *Metric)
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager() *MetricsManager {
	manager := &MetricsManager{
		metrics:     make(map[string]*Metric),
		subscribers: make([]MetricsSubscriber, 0),
	}
	
	// Initialize system metrics
	manager.initSystemMetrics()
	
	return manager
}

// initSystemMetrics initializes system-level metrics
func (m *MetricsManager) initSystemMetrics() {
	// Memory metrics
	m.RegisterGauge("system.memory.alloc", "Memory allocated and not yet freed (bytes)", nil)
	m.RegisterGauge("system.memory.total_alloc", "Total memory allocated since process start (bytes)", nil)
	m.RegisterGauge("system.memory.sys", "Memory obtained from system (bytes)", nil)
	m.RegisterGauge("system.memory.mallocs", "Number of mallocs", nil)
	m.RegisterGauge("system.memory.frees", "Number of frees", nil)
	m.RegisterGauge("system.memory.heap_alloc", "Heap memory allocated (bytes)", nil)
	m.RegisterGauge("system.memory.heap_sys", "Heap memory obtained from system (bytes)", nil)
	m.RegisterGauge("system.memory.heap_idle", "Heap memory idle (bytes)", nil)
	m.RegisterGauge("system.memory.heap_inuse", "Heap memory in use (bytes)", nil)
	m.RegisterGauge("system.memory.heap_released", "Heap memory released to system (bytes)", nil)
	m.RegisterGauge("system.memory.heap_objects", "Number of allocated heap objects", nil)
	
	// GC metrics
	m.RegisterGauge("system.gc.next", "Next GC target heap size (bytes)", nil)
	m.RegisterGauge("system.gc.last", "Time the last GC finished (unix timestamp)", nil)
	m.RegisterGauge("system.gc.num", "Number of GC cycles completed", nil)
	m.RegisterGauge("system.gc.cpu_fraction", "Fraction of CPU time used by GC", nil)
	
	// Goroutine metrics
	m.RegisterGauge("system.goroutines", "Number of goroutines", nil)
}

// RegisterCounter registers a new counter metric
func (m *MetricsManager) RegisterCounter(name, description string, labels map[string]string) *Metric {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	metric := &Metric{
		Name:        name,
		Type:        CounterMetric,
		Value:       0,
		Labels:      labels,
		LastUpdated: time.Now(),
	}
	
	m.metrics[name] = metric
	return metric
}

// RegisterGauge registers a new gauge metric
func (m *MetricsManager) RegisterGauge(name, description string, labels map[string]string) *Metric {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	metric := &Metric{
		Name:        name,
		Type:        GaugeMetric,
		Value:       0,
		Labels:      labels,
		LastUpdated: time.Now(),
	}
	
	m.metrics[name] = metric
	return metric
}

// RegisterHistogram registers a new histogram metric
func (m *MetricsManager) RegisterHistogram(name, description string, buckets []float64, labels map[string]string) *Metric {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	bucketMap := make(map[float64]int)
	for _, bucket := range buckets {
		bucketMap[bucket] = 0
	}
	
	metric := &Metric{
		Name:        name,
		Type:        HistogramMetric,
		Value:       0,
		Labels:      labels,
		LastUpdated: time.Now(),
		Buckets:     bucketMap,
		Sum:         0,
		Count:       0,
	}
	
	m.metrics[name] = metric
	return metric
}

// GetMetric gets a metric by name
func (m *MetricsManager) GetMetric(name string) (*Metric, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	metric, ok := m.metrics[name]
	return metric, ok
}

// IncrementCounter increments a counter metric by the given value
func (m *MetricsManager) IncrementCounter(name string, value float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	metric, ok := m.metrics[name]
	if !ok {
		return fmt.Errorf("metric not found: %s", name)
	}
	
	if metric.Type != CounterMetric {
		return fmt.Errorf("metric is not a counter: %s", name)
	}
	
	metric.Value += value
	metric.LastUpdated = time.Now()
	
	// Notify subscribers
	m.notifySubscribers(metric)
	
	return nil
}

// SetGauge sets a gauge metric to the given value
func (m *MetricsManager) SetGauge(name string, value float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	metric, ok := m.metrics[name]
	if !ok {
		return fmt.Errorf("metric not found: %s", name)
	}
	
	if metric.Type != GaugeMetric {
		return fmt.Errorf("metric is not a gauge: %s", name)
	}
	
	metric.Value = value
	metric.LastUpdated = time.Now()
	
	// Notify subscribers
	m.notifySubscribers(metric)
	
	return nil
}

// ObserveHistogram adds an observation to a histogram metric
func (m *MetricsManager) ObserveHistogram(name string, value float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	metric, ok := m.metrics[name]
	if !ok {
		return fmt.Errorf("metric not found: %s", name)
	}
	
	if metric.Type != HistogramMetric {
		return fmt.Errorf("metric is not a histogram: %s", name)
	}
	
	// Update histogram buckets
	for bucket := range metric.Buckets {
		if value <= bucket {
			metric.Buckets[bucket]++
		}
	}
	
	// Update histogram stats
	metric.Sum += value
	metric.Count++
	metric.LastUpdated = time.Now()
	
	// Notify subscribers
	m.notifySubscribers(metric)
	
	return nil
}

// Subscribe adds a subscriber to be notified of metric updates
func (m *MetricsManager) Subscribe(subscriber MetricsSubscriber) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.subscribers = append(m.subscribers, subscriber)
}

// Unsubscribe removes a subscriber
func (m *MetricsManager) Unsubscribe(subscriber MetricsSubscriber) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for i, s := range m.subscribers {
		if s == subscriber {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			break
		}
	}
}

// notifySubscribers notifies all subscribers of a metric update
func (m *MetricsManager) notifySubscribers(metric *Metric) {
	for _, subscriber := range m.subscribers {
		subscriber.OnMetricUpdate(metric)
	}
}

// CollectSystemMetrics collects system metrics
func (m *MetricsManager) CollectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Update memory metrics
	m.SetGauge("system.memory.alloc", float64(memStats.Alloc))
	m.SetGauge("system.memory.total_alloc", float64(memStats.TotalAlloc))
	m.SetGauge("system.memory.sys", float64(memStats.Sys))
	m.SetGauge("system.memory.mallocs", float64(memStats.Mallocs))
	m.SetGauge("system.memory.frees", float64(memStats.Frees))
	m.SetGauge("system.memory.heap_alloc", float64(memStats.HeapAlloc))
	m.SetGauge("system.memory.heap_sys", float64(memStats.HeapSys))
	m.SetGauge("system.memory.heap_idle", float64(memStats.HeapIdle))
	m.SetGauge("system.memory.heap_inuse", float64(memStats.HeapInuse))
	m.SetGauge("system.memory.heap_released", float64(memStats.HeapReleased))
	m.SetGauge("system.memory.heap_objects", float64(memStats.HeapObjects))
	
	// Update GC metrics
	m.SetGauge("system.gc.next", float64(memStats.NextGC))
	m.SetGauge("system.gc.last", float64(memStats.LastGC))
	m.SetGauge("system.gc.num", float64(memStats.NumGC))
	m.SetGauge("system.gc.cpu_fraction", float64(memStats.GCCPUFraction))
	
	// Update goroutine metrics
	m.SetGauge("system.goroutines", float64(runtime.NumGoroutine()))
}

// StartCollectingSystemMetrics starts collecting system metrics at the specified interval
func (m *MetricsManager) StartCollectingSystemMetrics(interval time.Duration) chan struct{} {
	stopChan := make(chan struct{})
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.CollectSystemMetrics()
			case <-stopChan:
				return
			}
		}
	}()
	
	return stopChan
}

// GetAllMetrics returns all metrics
func (m *MetricsManager) GetAllMetrics() map[string]*Metric {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Create a copy of the metrics map
	metrics := make(map[string]*Metric, len(m.metrics))
	for name, metric := range m.metrics {
		metrics[name] = metric
	}
	
	return metrics
}

// GetMetricsByPrefix returns all metrics with the given prefix
func (m *MetricsManager) GetMetricsByPrefix(prefix string) map[string]*Metric {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	metrics := make(map[string]*Metric)
	for name, metric := range m.metrics {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			metrics[name] = metric
		}
	}
	
	return metrics
}

// GetMetricValue gets the value of a metric
func (m *MetricsManager) GetMetricValue(name string) (float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	metric, ok := m.metrics[name]
	if !ok {
		return 0, fmt.Errorf("metric not found: %s", name)
	}
	
	return metric.Value, nil
}

// ResetMetric resets a metric to its initial value
func (m *MetricsManager) ResetMetric(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	metric, ok := m.metrics[name]
	if !ok {
		return fmt.Errorf("metric not found: %s", name)
	}
	
	switch metric.Type {
	case CounterMetric:
		metric.Value = 0
	case GaugeMetric:
		metric.Value = 0
	case HistogramMetric:
		for bucket := range metric.Buckets {
			metric.Buckets[bucket] = 0
		}
		metric.Sum = 0
		metric.Count = 0
	}
	
	metric.LastUpdated = time.Now()
	
	// Notify subscribers
	m.notifySubscribers(metric)
	
	return nil
}

// GetHistogramStats gets statistics for a histogram metric
func (m *MetricsManager) GetHistogramStats(name string) (sum float64, count int, buckets map[float64]int, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	metric, ok := m.metrics[name]
	if !ok {
		return 0, 0, nil, fmt.Errorf("metric not found: %s", name)
	}
	
	if metric.Type != HistogramMetric {
		return 0, 0, nil, fmt.Errorf("metric is not a histogram: %s", name)
	}
	
	// Create a copy of the buckets map
	bucketsCopy := make(map[float64]int, len(metric.Buckets))
	for bucket, count := range metric.Buckets {
		bucketsCopy[bucket] = count
	}
	
	return metric.Sum, metric.Count, bucketsCopy, nil
}
