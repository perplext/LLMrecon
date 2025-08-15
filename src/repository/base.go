package repository

import (
	"context"
	"fmt"
	"sync"
)

// BaseRepository provides common functionality for all repository implementations
type BaseRepository struct {
	// config is the repository configuration
	config *Config
	
	// connected indicates if the repository is connected
	connected bool
	
	// connectionMutex protects the connected flag
	connectionMutex sync.RWMutex
	
	// lastError is the last error that occurred
	lastError error
	
	// lastErrorTime is the time the last error occurred
	lastErrorTime time.Time
	
	// connectionPool is a pool of connections to the repository
	connectionPool chan struct{}

// NewBaseRepository creates a new base repository
func NewBaseRepository(config *Config) *BaseRepository {
	// Create connection pool
	pool := make(chan struct{}, config.MaxConnections)
	for i := 0; i < config.MaxConnections; i++ {
		pool <- struct{}{}
	}
	
	return &BaseRepository{
		config:         config,
		connected:      false,
		connectionPool: pool,
	}

// GetType returns the repository type
func (r *BaseRepository) GetType() RepositoryType {
	return r.config.Type

// GetName returns the repository name
func (r *BaseRepository) GetName() string {
	return r.config.Name

// GetURL returns the repository URL
func (r *BaseRepository) GetURL() string {
	return r.config.URL

// IsConnected returns true if the repository is connected
func (r *BaseRepository) IsConnected() bool {
	r.connectionMutex.RLock()
	defer r.connectionMutex.RUnlock()
	return r.connected

// setConnected sets the connected flag
func (r *BaseRepository) setConnected(connected bool) {
	r.connectionMutex.Lock()
	defer r.connectionMutex.Unlock()
	r.connected = connected

// setLastError sets the last error and its time
func (r *BaseRepository) setLastError(err error) {
	r.lastError = err
	r.lastErrorTime = time.Now()

// GetLastError returns the last error and its time
func (r *BaseRepository) GetLastError() (error, time.Time) {
	return r.lastError, r.lastErrorTime

// AcquireConnection acquires a connection from the pool
func (r *BaseRepository) AcquireConnection(ctx context.Context) error {
	select {
	case <-r.connectionPool:
		// Connection acquired
		return nil
	case <-ctx.Done():
		// Context canceled or timed out
		return fmt.Errorf("failed to acquire connection: %w", ctx.Err())
	}

// ReleaseConnection releases a connection back to the pool
func (r *BaseRepository) ReleaseConnection() {
	select {
	case r.connectionPool <- struct{}{}:
		// Connection released
	default:
		// Pool is full, which shouldn't happen
		// Log this as it indicates a bug
		fmt.Printf("Warning: connection pool overflow for repository %s\n", r.config.Name)
	}

// WithRetry executes a function with retry logic
func (r *BaseRepository) WithRetry(ctx context.Context, operation func() error) error {
	var err error
	
	for attempt := 0; attempt <= r.config.RetryCount; attempt++ {
		// Execute the operation
		err = operation()
		if err == nil {
			// Operation succeeded
			return nil
		}
		
		// Check if context is canceled
		if ctx.Err() != nil {
			return fmt.Errorf("operation canceled: %w", ctx.Err())
		}
		
		// If this was the last attempt, return the error
		if attempt == r.config.RetryCount {
			break
		}
		
		// Wait before retrying
		select {
		case <-time.After(r.config.RetryDelay):
			// Continue with next attempt
		case <-ctx.Done():
			// Context canceled while waiting
			return fmt.Errorf("operation canceled while waiting to retry: %w", ctx.Err())
		}
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", r.config.RetryCount+1, err)

// CreateContextWithTimeout creates a context with the repository's timeout
func (r *BaseRepository) CreateContextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.config.Timeout)
