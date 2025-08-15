package core

import "time"

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// ConnectionPoolConfig defines configuration for HTTP connection pools
type ConnectionPoolConfig struct {
	// Pool sizing
	MaxIdleConns        int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host"`
	MaxConnsPerHost     int           `json:"max_conns_per_host"`
	
	// Timeouts
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout"`
	TLSHandshakeTimeout   time.Duration `json:"tls_handshake_timeout"`
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout"`
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout"`
	
	// Keep-alive settings
	KeepAlive             time.Duration `json:"keep_alive"`
	DisableKeepAlives     bool          `json:"disable_keep_alives"`
	DisableCompression    bool          `json:"disable_compression"`
	
	// TLS settings
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
	
	// Health check settings
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	
	// Provider-specific settings
	ProviderType ProviderType `json:"provider_type"`
	BaseURL      string       `json:"base_url"`
}

// ConnectionPoolManager manages HTTP connection pools for LLM providers
type ConnectionPoolManager struct {
	pools     map[ProviderType]*ProviderConnectionPool
	config    ConnectionPoolConfig
	logger    Logger
	metrics   *ConnectionPoolMetrics
	mutex     sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// ProviderConnectionPool manages connections for a specific provider
type ProviderConnectionPool struct {
	providerType ProviderType
	client       *http.Client
	transport    *http.Transport
	config       ConnectionPoolConfig
	metrics      *ProviderPoolMetrics
	healthCheck  *HealthChecker
	mutex        sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// ConnectionPoolMetrics tracks connection pool performance
type ConnectionPoolMetrics struct {
	TotalPools         int64 `json:"total_pools"`
	ActiveConnections  int64 `json:"active_connections"`
	IdleConnections    int64 `json:"idle_connections"`
	ConnectionsCreated int64 `json:"connections_created"`
	ConnectionsReused  int64 `json:"connections_reused"`
	ConnectionErrors   int64 `json:"connection_errors"`
	HealthChecksPassed int64 `json:"health_checks_passed"`
	HealthChecksFailed int64 `json:"health_checks_failed"`
	
	// Per-provider metrics
	ProviderMetrics map[ProviderType]*ProviderPoolMetrics `json:"provider_metrics"`
}

// ProviderPoolMetrics tracks metrics for a specific provider pool
type ProviderPoolMetrics struct {
	ProviderType       ProviderType  `json:"provider_type"`
	ActiveConnections  int64         `json:"active_connections"`
	IdleConnections    int64         `json:"idle_connections"`
	ConnectionsCreated int64         `json:"connections_created"`
	ConnectionsReused  int64         `json:"connections_reused"`
	ConnectionErrors   int64         `json:"connection_errors"`
	AverageLatency     time.Duration `json:"average_latency"`
	LastHealthCheck    time.Time     `json:"last_health_check"`
	HealthStatus       string        `json:"health_status"`
}

// HealthChecker performs periodic health checks on connections
type HealthChecker struct {
	config   ConnectionPoolConfig
	client   *http.Client
	endpoint string
	logger   Logger
	metrics  *ProviderPoolMetrics
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	ticker   *time.Ticker
	wg       sync.WaitGroup
}

// DefaultConnectionPoolConfig returns default configuration
func DefaultConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		KeepAlive:             30 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		InsecureSkipVerify:    false,
		HealthCheckInterval:   30 * time.Second,
		HealthCheckTimeout:    5 * time.Second,
	}
}

// NewConnectionPoolManager creates a new connection pool manager
func NewConnectionPoolManager(config ConnectionPoolConfig, logger Logger) *ConnectionPoolManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &ConnectionPoolManager{
		pools:   make(map[ProviderType]*ProviderConnectionPool),
		config:  config,
		logger:  logger,
		metrics: NewConnectionPoolMetrics(),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	return manager
}

// CreatePool creates a connection pool for a specific provider
func (m *ConnectionPoolManager) CreatePool(providerType ProviderType, config ConnectionPoolConfig) (*ProviderConnectionPool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.pools[providerType]; exists {
		return nil, fmt.Errorf("connection pool for provider %s already exists", providerType)
	}
	
	pool, err := m.createProviderPool(providerType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool for %s: %w", providerType, err)
	}
	
	m.pools[providerType] = pool
	m.metrics.TotalPools++
	m.metrics.ProviderMetrics[providerType] = pool.metrics
	
	m.logger.Info("Created connection pool", "provider", providerType, "max_idle", config.MaxIdleConns)
	return pool, nil
}

// GetPool returns the connection pool for a provider
func (m *ConnectionPoolManager) GetPool(providerType ProviderType) (*ProviderConnectionPool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	pool, exists := m.pools[providerType]
	if !exists {
		return nil, fmt.Errorf("no connection pool found for provider %s", providerType)
	}
	
	return pool, nil
}

// GetClient returns an HTTP client for a provider with connection pooling
func (m *ConnectionPoolManager) GetClient(providerType ProviderType) (*http.Client, error) {
	pool, err := m.GetPool(providerType)
	if err != nil {
		return nil, err
	}
	
	return pool.client, nil
}

// Start starts the connection pool manager and health checkers
func (m *ConnectionPoolManager) Start() error {
	m.logger.Info("Starting connection pool manager")
	
	for _, pool := range m.pools {
		if err := pool.Start(); err != nil {
			return fmt.Errorf("failed to start pool for %s: %w", pool.providerType, err)
		}
	}
	
	// Start metrics collection
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.metricsLoop()
	}()
	
	m.logger.Info("Connection pool manager started")
	return nil
}

// Stop stops the connection pool manager
func (m *ConnectionPoolManager) Stop() error {
	m.logger.Info("Stopping connection pool manager")
	
	m.cancel()
	
	for _, pool := range m.pools {
		pool.Stop()
	}
	
	m.wg.Wait()
	
	m.logger.Info("Connection pool manager stopped")
	return nil
}

// GetMetrics returns current connection pool metrics
func (m *ConnectionPoolManager) GetMetrics() *ConnectionPoolMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Update aggregate metrics
	var totalActive, totalIdle int64
	for _, poolMetrics := range m.metrics.ProviderMetrics {
		totalActive += poolMetrics.ActiveConnections
		totalIdle += poolMetrics.IdleConnections
	}
	
	m.metrics.ActiveConnections = totalActive
	m.metrics.IdleConnections = totalIdle
	
	return m.metrics
}
// createProviderPool creates a connection pool for a specific provider
func (m *ConnectionPoolManager) createProviderPool(providerType ProviderType, config ConnectionPoolConfig) (*ProviderConnectionPool, error) {
	ctx, cancel := context.WithCancel(m.ctx)
	
	// Create custom transport with connection pooling settings
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
		DisableCompression:    config.DisableCompression,
	}
	
	// Configure TLS if needed
	if config.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: false, // Fixed: Enable cert validation
		}
	}
	
	// Create HTTP client
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // Default timeout, can be overridden per request
	}
	
	// Create health checker
	healthChecker := &HealthChecker{
		config:   config,
		client:   client,
		endpoint: config.BaseURL + "/health", // Assume health endpoint
		logger:   m.logger,
		ctx:      ctx,
		cancel:   cancel,
		ticker:   time.NewTicker(config.HealthCheckInterval),
	}
	
	pool := &ProviderConnectionPool{
		providerType: providerType,
		client:       client,
		transport:    transport,
		config:       config,
		metrics:      NewProviderPoolMetrics(providerType),
		healthCheck:  healthChecker,
		ctx:          ctx,
		cancel:       cancel,
	}
	
	healthChecker.metrics = pool.metrics
	
	return pool, nil
}

// metricsLoop periodically updates connection pool metrics
func (m *ConnectionPoolManager) metricsLoop() {
	ticker := time.NewTicker(10 * time.Second)
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

// updateMetrics updates connection pool metrics
func (m *ConnectionPoolManager) updateMetrics() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for _, pool := range m.pools {
		pool.updateMetrics()
	}
}

// ProviderConnectionPool methods

// Start starts the provider connection pool
func (p *ProviderConnectionPool) Start() error {
	// Start health checker
	if p.config.HealthCheckInterval > 0 {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.healthCheck.start()
		}()
	}
	
	return nil
}

// Stop stops the provider connection pool
func (p *ProviderConnectionPool) Stop() {
	p.cancel()
	p.healthCheck.stop()
	p.wg.Wait()
	
	// Close idle connections
	p.transport.CloseIdleConnections()
}

// GetClient returns the HTTP client for this pool
func (p *ProviderConnectionPool) GetClient() *http.Client {
	p.metrics.ConnectionsReused++
	return p.client
}

// updateMetrics updates provider pool metrics
func (p *ProviderConnectionPool) updateMetrics() {
	// Note: Go's http.Transport doesn't expose connection count directly
	// These would need to be tracked through custom transport or monitoring
	p.metrics.LastHealthCheck = time.Now()
}

// GetMetrics returns provider pool metrics
func (p *ProviderConnectionPool) GetMetrics() *ProviderPoolMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.metrics
}

// HealthChecker methods

// start starts the health checker
func (h *HealthChecker) start() {
	for {
		select {
		case <-h.ticker.C:
			h.performHealthCheck()
		case <-h.ctx.Done():
			return
		}
	}
}

// stop stops the health checker
func (h *HealthChecker) stop() {
	h.cancel()
	h.ticker.Stop()
	h.wg.Wait()
}

// performHealthCheck performs a health check on the connection pool
func (h *HealthChecker) performHealthCheck() {
	ctx, cancel := context.WithTimeout(h.ctx, h.config.HealthCheckTimeout)
	defer cancel()
	
	start := time.Now()
	
	req, err := http.NewRequestWithContext(ctx, "GET", h.endpoint, nil)
	if err != nil {
		h.handleHealthCheckFailure(err, time.Since(start))
		return
	}
	
	resp, err := h.client.Do(req)
	if err != nil {
		h.handleHealthCheckFailure(err, time.Since(start))
		return
	}
	defer func() { if err := resp.Body.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		h.handleHealthCheckSuccess(time.Since(start))
	} else {
		h.handleHealthCheckFailure(fmt.Errorf("health check returned status %d", resp.StatusCode), time.Since(start))
	}

// handleHealthCheckSuccess handles successful health checks
func (h *HealthChecker) handleHealthCheckSuccess(latency time.Duration) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.metrics.LastHealthCheck = time.Now()
	h.metrics.HealthStatus = "healthy"
	h.metrics.AverageLatency = (h.metrics.AverageLatency + latency) / 2
	h.metrics.ConnectionsReused++
	
	h.logger.Debug("Health check passed", "provider", h.metrics.ProviderType, "latency", latency)

// handleHealthCheckFailure handles failed health checks
func (h *HealthChecker) handleHealthCheckFailure(err error, latency time.Duration) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.metrics.LastHealthCheck = time.Now()
	h.metrics.HealthStatus = "unhealthy"
	h.metrics.ConnectionErrors++
	
	h.logger.Warn("Health check failed", "provider", h.metrics.ProviderType, "error", err, "latency", latency)

// Utility functions

// NewConnectionPoolMetrics creates new connection pool metrics
func NewConnectionPoolMetrics() *ConnectionPoolMetrics {
	return &ConnectionPoolMetrics{
		ProviderMetrics: make(map[ProviderType]*ProviderPoolMetrics),
	}

// NewProviderPoolMetrics creates new provider pool metrics
func NewProviderPoolMetrics(providerType ProviderType) *ProviderPoolMetrics {
	return &ProviderPoolMetrics{
		ProviderType: providerType,
		HealthStatus: "unknown",
	}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
