package performance

import "time"

// Logger interface for performance package logging operations
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// BalancingStrategy defines the load balancing strategy
type BalancingStrategy string

const (
	BalancingRoundRobin  BalancingStrategy = "round_robin"
	BalancingLeastActive BalancingStrategy = "least_active"
	BalancingWeighted    BalancingStrategy = "weighted"
	BalancingConsistent  BalancingStrategy = "consistent"
	BalancingAdaptive    BalancingStrategy = "adaptive"
)

// DataPoint represents a single data point in a trend
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricsCollector interface for collecting metrics from various sources
type MetricsCollector interface {
	GetMetrics() map[string]interface{}
	GetName() string
	IsEnabled() bool
}
