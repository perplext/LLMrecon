package monitoring

import "time"

// Stats represents statistics for the file handler
type Stats struct {
	FilesServed      int64
	CacheHits        int64
	CacheMisses      int64
	CompressedFiles  int64
	TotalSize        int64
	CompressedSize   int64
	CompressionRatio float64
	AverageServeTime time.Duration
}

// FileHandlerInterface defines the interface for a static file handler
type FileHandlerInterface interface {
	GetStats() *Stats
	GetCacheSize() int64
	GetCacheItemCount() int64
}

// MetricsManagerInterface defines the interface for a metrics manager
type MetricsManagerInterface interface {
	RecordCounter(name string, value int64, tags map[string]string) error
	RecordGauge(name string, value interface{}, tags map[string]string) error
	RegisterGauge(name string, description string, tags map[string]string) error
}

// AlertManagerInterface defines the interface for an alert manager
type AlertManagerInterface interface {
	CheckThreshold(name string, value interface{}, tags map[string]string) error
}
