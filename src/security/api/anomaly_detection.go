// Package api provides API protection mechanisms for the LLMrecon tool.
package api

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AnomalyType represents the type of anomaly
type AnomalyType string

const (
	// AnomalyTypeRateSpike represents a sudden spike in request rate
	AnomalyTypeRateSpike AnomalyType = "rate_spike"
	// AnomalyTypeUnusualPath represents access to unusual paths
	AnomalyTypeUnusualPath AnomalyType = "unusual_path"
	// AnomalyTypeUnusualMethod represents use of unusual HTTP methods
	AnomalyTypeUnusualMethod AnomalyType = "unusual_method"
	// AnomalyTypeUnusualUserAgent represents use of unusual user agents
	AnomalyTypeUnusualUserAgent AnomalyType = "unusual_user_agent"
	// AnomalyTypeUnusualIP represents access from unusual IPs
	AnomalyTypeUnusualIP AnomalyType = "unusual_ip"
	// AnomalyTypeUnusualPattern represents unusual access patterns
	AnomalyTypeUnusualPattern AnomalyType = "unusual_pattern"
)

// AnomalyLevel represents the severity level of an anomaly
type AnomalyLevel int

const (
	// AnomalyLevelInfo is for informational anomalies
	AnomalyLevelInfo AnomalyLevel = iota
	// AnomalyLevelWarning is for warning anomalies
	AnomalyLevelWarning
	// AnomalyLevelAlert is for alert anomalies
	AnomalyLevelAlert
	// AnomalyLevelCritical is for critical anomalies
	AnomalyLevelCritical
)

// Anomaly represents a detected anomaly
type Anomaly struct {
	// Type is the type of anomaly
	Type AnomalyType `json:"type"`
	// Level is the severity level of the anomaly
	Level AnomalyLevel `json:"level"`
	// Description is a description of the anomaly
	Description string `json:"description"`
	// Timestamp is the time the anomaly was detected
	Timestamp time.Time `json:"timestamp"`
	// ClientIP is the client IP associated with the anomaly
	ClientIP string `json:"client_ip,omitempty"`
	// Path is the request path associated with the anomaly
	Path string `json:"path,omitempty"`
	// Method is the HTTP method associated with the anomaly
	Method string `json:"method,omitempty"`
	// UserAgent is the user agent associated with the anomaly
	UserAgent string `json:"user_agent,omitempty"`
	// RequestID is the request ID associated with the anomaly
	RequestID string `json:"request_id,omitempty"`
	// Count is the count associated with the anomaly
	Count int `json:"count,omitempty"`
	// Threshold is the threshold that was exceeded
	Threshold int `json:"threshold,omitempty"`
	// Duration is the duration associated with the anomaly
	Duration time.Duration `json:"duration,omitempty"`
}

// AnomalyDetectorConfig represents the configuration for an anomaly detector
type AnomalyDetectorConfig struct {
	// Enabled indicates whether anomaly detection is enabled
	Enabled bool
	// WindowSize is the size of the sliding window for rate calculations
	WindowSize time.Duration
	// RateSpikeThreshold is the threshold for rate spikes
	RateSpikeThreshold int
	// UnusualPathThreshold is the threshold for unusual paths
	UnusualPathThreshold int
	// UnusualMethodThreshold is the threshold for unusual methods
	UnusualMethodThreshold int
	// UnusualUserAgentThreshold is the threshold for unusual user agents
	UnusualUserAgentThreshold int
	// UnusualIPThreshold is the threshold for unusual IPs
	UnusualIPThreshold int
	// UnusualPatternThreshold is the threshold for unusual patterns
	UnusualPatternThreshold int
	// LearningMode indicates whether the detector is in learning mode
	LearningMode bool
	// LearningPeriod is the learning period
	LearningPeriod time.Duration
	// AlertCallback is a callback function for anomaly alerts
	AlertCallback func(anomaly *Anomaly)
}

// DefaultAnomalyDetectorConfig returns the default anomaly detector configuration
func DefaultAnomalyDetectorConfig() *AnomalyDetectorConfig {
	return &AnomalyDetectorConfig{
		Enabled:                  true,
		WindowSize:               5 * time.Minute,
		RateSpikeThreshold:       100,
		UnusualPathThreshold:     10,
		UnusualMethodThreshold:   5,
		UnusualUserAgentThreshold: 5,
		UnusualIPThreshold:       10,
		UnusualPatternThreshold:  10,
		LearningMode:             true,
		LearningPeriod:           24 * time.Hour,
	}
}

// AnomalyDetector implements anomaly detection for API requests
type AnomalyDetector struct {
	config           *AnomalyDetectorConfig
	startTime        time.Time
	mu               sync.RWMutex
	pathCounts       map[string]int
	methodCounts     map[string]int
	userAgentCounts  map[string]int
	ipCounts         map[string]int
	patternCounts    map[string]int
	requestCounts    []int
	requestTimes     []time.Time
	anomalies        []*Anomaly
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(config *AnomalyDetectorConfig) *AnomalyDetector {
	if config == nil {
		config = DefaultAnomalyDetectorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	detector := &AnomalyDetector{
		config:          config,
		startTime:       time.Now(),
		pathCounts:      make(map[string]int),
		methodCounts:    make(map[string]int),
		userAgentCounts: make(map[string]int),
		ipCounts:        make(map[string]int),
		patternCounts:   make(map[string]int),
		requestCounts:   make([]int, 0),
		requestTimes:    make([]time.Time, 0),
		anomalies:       make([]*Anomaly, 0),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start the cleanup goroutine
	go detector.cleanup()

	return detector
}

// cleanup periodically cleans up old data
func (ad *AnomalyDetector) cleanup() {
	ticker := time.NewTicker(ad.config.WindowSize / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ad.cleanupOldData()
		case <-ad.ctx.Done():
			return
		}
	}
}

// cleanupOldData removes data older than the window size
func (ad *AnomalyDetector) cleanupOldData() {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	cutoff := time.Now().Add(-ad.config.WindowSize)

	// Remove old request times and counts
	newIndex := 0
	for i, t := range ad.requestTimes {
		if t.After(cutoff) {
			newIndex = i
			break
		}
	}

	if newIndex > 0 {
		ad.requestTimes = ad.requestTimes[newIndex:]
		ad.requestCounts = ad.requestCounts[newIndex:]
	}

	// Remove old anomalies
	newAnomalies := make([]*Anomaly, 0)
	for _, anomaly := range ad.anomalies {
		if anomaly.Timestamp.After(cutoff) {
			newAnomalies = append(newAnomalies, anomaly)
		}
	}
	ad.anomalies = newAnomalies
}

// Close stops the anomaly detector
func (ad *AnomalyDetector) Close() {
	ad.cancel()
}

// RecordRequest records a request for anomaly detection
func (ad *AnomalyDetector) RecordRequest(r *http.Request) {
	// Skip if not enabled
	if !ad.config.Enabled {
		return
	}

	ad.mu.Lock()
	defer ad.mu.Unlock()

	// Record request time
	now := time.Now()
	ad.requestTimes = append(ad.requestTimes, now)
	ad.requestCounts = append(ad.requestCounts, 1)

	// Record path
	ad.pathCounts[r.URL.Path]++

	// Record method
	ad.methodCounts[r.Method]++

	// Record user agent
	userAgent := r.UserAgent()
	if userAgent != "" {
		ad.userAgentCounts[userAgent]++
	}

	// Record IP
	clientIP := r.RemoteAddr
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		clientIP = strings.Split(ip, ",")[0]
	}
	ad.ipCounts[clientIP]++

	// Record pattern (method + path)
	pattern := r.Method + " " + r.URL.Path
	ad.patternCounts[pattern]++

	// Check for anomalies if not in learning mode
	if !ad.config.LearningMode || time.Since(ad.startTime) > ad.config.LearningPeriod {
		ad.checkForAnomalies(r)
	}
}

// checkForAnomalies checks for anomalies in the current request
func (ad *AnomalyDetector) checkForAnomalies(r *http.Request) {
	// Get client IP
	clientIP := r.RemoteAddr
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		clientIP = strings.Split(ip, ",")[0]
	}

	// Check for rate spikes
	if len(ad.requestTimes) > 0 {
		// Calculate the rate over the window
		windowStart := time.Now().Add(-ad.config.WindowSize)
		count := 0
		for i, t := range ad.requestTimes {
			if t.After(windowStart) {
				count += ad.requestCounts[i]
			}
		}

		// Check if the rate exceeds the threshold
		if count > ad.config.RateSpikeThreshold {
			anomaly := &Anomaly{
				Type:        AnomalyTypeRateSpike,
				Level:       AnomalyLevelAlert,
				Description: "Request rate spike detected",
				Timestamp:   time.Now(),
				Count:       count,
				Threshold:   ad.config.RateSpikeThreshold,
				Duration:    ad.config.WindowSize,
			}
			ad.recordAnomaly(anomaly)
		}
	}

	// Check for unusual path
	pathCount := ad.pathCounts[r.URL.Path]
	totalPaths := len(ad.pathCounts)
	if totalPaths > ad.config.UnusualPathThreshold && pathCount == 1 {
		anomaly := &Anomaly{
			Type:        AnomalyTypeUnusualPath,
			Level:       AnomalyLevelWarning,
			Description: "Unusual path accessed",
			Timestamp:   time.Now(),
			ClientIP:    clientIP,
			Path:        r.URL.Path,
			Method:      r.Method,
			UserAgent:   r.UserAgent(),
			RequestID:   r.Header.Get("X-Request-ID"),
		}
		ad.recordAnomaly(anomaly)
	}

	// Check for unusual method
	methodCount := ad.methodCounts[r.Method]
	totalMethods := len(ad.methodCounts)
	if totalMethods > ad.config.UnusualMethodThreshold && methodCount == 1 {
		anomaly := &Anomaly{
			Type:        AnomalyTypeUnusualMethod,
			Level:       AnomalyLevelWarning,
			Description: "Unusual HTTP method used",
			Timestamp:   time.Now(),
			ClientIP:    clientIP,
			Path:        r.URL.Path,
			Method:      r.Method,
			UserAgent:   r.UserAgent(),
			RequestID:   r.Header.Get("X-Request-ID"),
		}
		ad.recordAnomaly(anomaly)
	}

	// Check for unusual user agent
	userAgent := r.UserAgent()
	if userAgent != "" {
		userAgentCount := ad.userAgentCounts[userAgent]
		totalUserAgents := len(ad.userAgentCounts)
		if totalUserAgents > ad.config.UnusualUserAgentThreshold && userAgentCount == 1 {
			anomaly := &Anomaly{
				Type:        AnomalyTypeUnusualUserAgent,
				Level:       AnomalyLevelWarning,
				Description: "Unusual user agent detected",
				Timestamp:   time.Now(),
				ClientIP:    clientIP,
				Path:        r.URL.Path,
				Method:      r.Method,
				UserAgent:   userAgent,
				RequestID:   r.Header.Get("X-Request-ID"),
			}
			ad.recordAnomaly(anomaly)
		}
	}

	// Check for unusual IP
	ipCount := ad.ipCounts[clientIP]
	totalIPs := len(ad.ipCounts)
	if totalIPs > ad.config.UnusualIPThreshold && ipCount == 1 {
		anomaly := &Anomaly{
			Type:        AnomalyTypeUnusualIP,
			Level:       AnomalyLevelWarning,
			Description: "Access from unusual IP address",
			Timestamp:   time.Now(),
			ClientIP:    clientIP,
			Path:        r.URL.Path,
			Method:      r.Method,
			UserAgent:   r.UserAgent(),
			RequestID:   r.Header.Get("X-Request-ID"),
		}
		ad.recordAnomaly(anomaly)
	}

	// Check for unusual pattern
	pattern := r.Method + " " + r.URL.Path
	patternCount := ad.patternCounts[pattern]
	totalPatterns := len(ad.patternCounts)
	if totalPatterns > ad.config.UnusualPatternThreshold && patternCount == 1 {
		anomaly := &Anomaly{
			Type:        AnomalyTypeUnusualPattern,
			Level:       AnomalyLevelWarning,
			Description: "Unusual access pattern detected",
			Timestamp:   time.Now(),
			ClientIP:    clientIP,
			Path:        r.URL.Path,
			Method:      r.Method,
			UserAgent:   r.UserAgent(),
			RequestID:   r.Header.Get("X-Request-ID"),
		}
		ad.recordAnomaly(anomaly)
	}
}

// recordAnomaly records an anomaly
func (ad *AnomalyDetector) recordAnomaly(anomaly *Anomaly) {
	// Add to the list of anomalies
	ad.anomalies = append(ad.anomalies, anomaly)

	// Call the alert callback if configured
	if ad.config.AlertCallback != nil {
		go ad.config.AlertCallback(anomaly)
	}
}

// GetAnomalies returns the list of detected anomalies
func (ad *AnomalyDetector) GetAnomalies() []*Anomaly {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	// Create a copy of the anomalies
	anomalies := make([]*Anomaly, len(ad.anomalies))
	copy(anomalies, ad.anomalies)

	return anomalies
}

// ClearAnomalies clears the list of detected anomalies
func (ad *AnomalyDetector) ClearAnomalies() {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	ad.anomalies = make([]*Anomaly, 0)
}

// Middleware returns a middleware function for anomaly detection
func (ad *AnomalyDetector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the request
		ad.RecordRequest(r)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
