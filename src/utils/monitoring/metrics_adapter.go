package monitoring

// MetricsManagerAdapter adapts the existing MetricsManager to implement the MetricsManagerInterface
type MetricsManagerAdapter struct {
	manager interface{
		RegisterGauge(name, description string, labels map[string]string) *Metric
		IncrementCounter(name string, value float64) error
		SetGauge(name string, value float64) error
	}
}

// NewMetricsManagerAdapter creates a new adapter for MetricsManager
func NewMetricsManagerAdapter(manager *MetricsManager) *MetricsManagerAdapter {
	return &MetricsManagerAdapter{
		manager: manager,
	}
}

// RecordCounter records a counter metric
func (a *MetricsManagerAdapter) RecordCounter(name string, value int64, tags map[string]string) error {
	// Convert int64 to float64 for the underlying implementation
	return a.manager.IncrementCounter(name, float64(value))
}

// RecordGauge records a gauge metric
func (a *MetricsManagerAdapter) RecordGauge(name string, value interface{}, tags map[string]string) error {
	// Try to convert to float64
	var floatValue float64
	switch v := value.(type) {
	case float64:
		floatValue = v
	case float32:
		floatValue = float64(v)
	case int:
		floatValue = float64(v)
	case int64:
		floatValue = float64(v)
	case int32:
		floatValue = float64(v)
	default:
		// Default to 0 if conversion not possible
		floatValue = 0
	}
	
	return a.manager.SetGauge(name, floatValue)
}

// RegisterGauge registers a gauge metric
func (a *MetricsManagerAdapter) RegisterGauge(name string, description string, tags map[string]string) error {
	// The original implementation returns *Metric, but our interface requires error
	a.manager.RegisterGauge(name, description, tags)
	return nil
}
