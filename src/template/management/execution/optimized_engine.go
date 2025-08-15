// Package execution provides functionality for executing templates against LLM systems.
package execution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/interfaces"
)

// OptimizedTemplateExecutor is an enhanced template executor with improved performance
type OptimizedTemplateExecutor struct {
	// defaultOptions is the default execution options
	defaultOptions *ExecutionOptions
	// providers is a map of provider name to provider
	providers map[string]interfaces.LLMProvider
	// detectionEngines is a map of detection engine name to detection engine
	detectionEngines map[string]interfaces.DetectionEngine
	// responseCache is a cache for LLM responses
	responseCache *ResponseCache
	// workPool is a pool of workers for concurrent execution
	workPool *WorkPool
	// stats tracks execution statistics
	stats ExecutionStats
	// statsMutex protects the stats
	statsMutex sync.RWMutex

// ResponseCache is a cache for LLM responses
type ResponseCache struct {
	// cache is a map of cache key to cache entry
	cache map[string]*ResponseCacheEntry
	// mutex is a mutex for the cache
	mutex sync.RWMutex
	// maxSize is the maximum size of the cache
	maxSize int
	// ttl is the time-to-live for cache entries
	ttl time.Duration
	// stats tracks cache statistics
	stats CacheStats

// ResponseCacheEntry represents a cached LLM response
type ResponseCacheEntry struct {
	// Response is the cached response
	Response string
	// CreatedAt is the time the entry was created
	CreatedAt time.Time
	// ExpiresAt is the time the entry expires
	ExpiresAt time.Time
	// Size is an estimate of the response size in bytes
	Size int

// CacheStats tracks cache statistics
type CacheStats struct {
	// Hits is the number of cache hits
	Hits int64
	// Misses is the number of cache misses
	Misses int64
	// Evictions is the number of cache evictions
	Evictions int64
	// TotalLookups is the total number of lookups
	TotalLookups int64

// WorkPool manages a pool of workers for concurrent execution
type WorkPool struct {
	// workers is a channel of worker functions
	workers chan func()
	// wg is a wait group for tracking worker completion
	wg sync.WaitGroup
	// size is the size of the worker pool
	size int

// ExecutionStats tracks execution statistics
type ExecutionStats struct {
	// TotalExecutions is the total number of template executions
	TotalExecutions int64
	// SuccessfulExecutions is the number of successful executions
	SuccessfulExecutions int64
	// FailedExecutions is the number of failed executions
	FailedExecutions int64
	// CachedResponses is the number of cached responses used
	CachedResponses int64
	// TotalExecutionTime is the total time spent executing templates
	TotalExecutionTime time.Duration
	// RetryCount is the number of retries performed
	RetryCount int64

// NewOptimizedTemplateExecutor creates a new optimized template executor
func NewOptimizedTemplateExecutor(defaultOptions *ExecutionOptions, cacheSize int, cacheTTL time.Duration, workerPoolSize int) *OptimizedTemplateExecutor {
	// Set default values
	if defaultOptions.Timeout == 0 {
		defaultOptions.Timeout = 30 * time.Second
	}
	if defaultOptions.RetryCount == 0 {
		defaultOptions.RetryCount = 3
	}
	if defaultOptions.RetryDelay == 0 {
		defaultOptions.RetryDelay = 1 * time.Second
	}
	if defaultOptions.MaxConcurrent == 0 {
		defaultOptions.MaxConcurrent = 10
	}
	if cacheSize <= 0 {
		cacheSize = 1000
	}
	if cacheTTL == 0 {
		cacheTTL = 1 * time.Hour
	}
	if workerPoolSize <= 0 {
		workerPoolSize = defaultOptions.MaxConcurrent
	}

	// Create response cache
	responseCache := &ResponseCache{
		cache:   make(map[string]*ResponseCacheEntry),
		maxSize: cacheSize,
		ttl:     cacheTTL,
	}

	// Create work pool
	workPool := newWorkPool(workerPoolSize)

	return &OptimizedTemplateExecutor{
		defaultOptions:   defaultOptions,
		providers:        make(map[string]interfaces.LLMProvider),
		detectionEngines: make(map[string]interfaces.DetectionEngine),
		responseCache:    responseCache,
		workPool:         workPool,
	}

// newWorkPool creates a new work pool
func newWorkPool(size int) *WorkPool {
	pool := &WorkPool{
		workers: make(chan func()),
		size:    size,
	}

	// Start workers
	for i := 0; i < size; i++ {
		go func() {
			for work := range pool.workers {
				work()
				pool.wg.Done()
			}
		}()
	}

	return pool

// RegisterProvider registers an LLM provider
func (e *OptimizedTemplateExecutor) RegisterProvider(provider interfaces.LLMProvider) {
	e.providers[provider.GetName()] = provider

// RegisterDetectionEngine registers a detection engine
func (e *OptimizedTemplateExecutor) RegisterDetectionEngine(engine interfaces.DetectionEngine) {
	e.detectionEngines[engine.GetName()] = engine

// Execute executes a template
func (e *OptimizedTemplateExecutor) Execute(ctx context.Context, template *format.Template, options map[string]interface{}) (*interfaces.TemplateResult, error) {
	startTime := time.Now()
	e.statsMutex.Lock()
	e.stats.TotalExecutions++
	e.statsMutex.Unlock()

	// Merge options with default options
	mergedOptions := e.mergeOptions(options)

	// Get provider
	providerName, ok := mergedOptions.ProviderOptions["provider"].(string)
	if !ok {
		providerName = "default"
	}

	provider, exists := e.providers[providerName]
	if !exists {
		e.statsMutex.Lock()
		e.stats.FailedExecutions++
		e.stats.TotalExecutionTime += time.Since(startTime)
		e.statsMutex.Unlock()
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	// Check if response is in cache
	cacheKey := e.generateCacheKey(template, mergedOptions)
	cachedResponse, found := e.getCachedResponse(cacheKey)
	if found {
		// Use cached response
		e.statsMutex.Lock()
		e.stats.CachedResponses++
		e.stats.SuccessfulExecutions++
		e.statsMutex.Unlock()

		// Create result
		result := &interfaces.TemplateResult{
			TemplateID:      template.ID,
			Status:          string(interfaces.StatusCompleted),
			Response:        cachedResponse,
			ExecutionTime:   0,
			CompletionTime:  time.Now(),
			Provider:        providerName,
			ProviderOptions: mergedOptions.ProviderOptions,
			FromCache:       true,
		}

		return result, nil
	}

	// Execute template with retry logic
	response, err := e.executeWithRetry(ctx, template, provider, mergedOptions)
	if err != nil {
		e.statsMutex.Lock()
		e.stats.FailedExecutions++
		e.stats.TotalExecutionTime += time.Since(startTime)
		e.statsMutex.Unlock()
		return nil, err
	}

	// Cache response
	e.cacheResponse(cacheKey, response, mergedOptions.ProviderOptions)

	// Run detection engine if specified
	var detected bool
	var score int
	var detectionResults map[string]interface{}

	if mergedOptions.DetectionEngine != nil {
		detected, score, detectionResults, err = mergedOptions.DetectionEngine.Detect(ctx, template, response)
		if err != nil {
			// Log detection error but continue
			fmt.Printf("Detection error: %v\n", err)
		}
	}

	// Create result
	result := &interfaces.TemplateResult{
		TemplateID:            template.ID,
		Status:                string(interfaces.StatusCompleted),
		Response:              response,
		ExecutionTime:         int64(time.Since(startTime).Milliseconds()),
		CompletionTime:        time.Now(),
		Provider:              providerName,
		ProviderOptions:       mergedOptions.ProviderOptions,
		VulnerabilityDetected: detected,
		VulnerabilityScore:    score,
		VulnerabilityDetails:  detectionResults,
		FromCache:             false,
		Duration:              time.Since(startTime),
		StartTime:             startTime,
		EndTime:               time.Now(),
	}

	e.statsMutex.Lock()
	e.stats.SuccessfulExecutions++
	e.stats.TotalExecutionTime += time.Since(startTime)
	e.statsMutex.Unlock()

	return result, nil

// ExecuteBatch executes multiple templates concurrently
func (e *OptimizedTemplateExecutor) ExecuteBatch(ctx context.Context, templates []*format.Template, options map[string]interface{}) ([]*interfaces.TemplateResult, error) {
	startTime := time.Now()

	// Create results slice
	results := make([]*interfaces.TemplateResult, len(templates))

	// Create error channel
	errorChan := make(chan error, len(templates))

	// Execute templates concurrently using work pool
	for i, template := range templates {
		i, template := i, template // Create local variables for closure
		
		// Submit work to pool
		e.workPool.wg.Add(1)
		e.workPool.workers <- func() {
			// Execute template
			result, err := e.Execute(ctx, template, options)
			if err != nil {
				errorChan <- err
				results[i] = &interfaces.TemplateResult{
					TemplateID:     template.ID,
					Status:         string(interfaces.StatusFailed),
					Error:          err,
					CompletionTime: time.Now(),
				}
				return
			}

			// Store result
			results[i] = result
		}
	}

	// Wait for all executions to complete
	e.workPool.wg.Wait()
	close(errorChan)

	// Check for errors
	var lastError error
	for err := range errorChan {
		lastError = err
	}

	// Track total execution time
	totalTime := time.Since(startTime)
	e.stats.TotalExecutionTime += time.Duration(totalTime.Milliseconds())
	
	// Return results
	return results, lastError

// executeWithRetry executes a template with retry logic
func (e *OptimizedTemplateExecutor) executeWithRetry(ctx context.Context, template *format.Template, provider interfaces.LLMProvider, options *ExecutionOptions) (string, error) {
	var response string
	var err error
	var retryCount int

	// Apply rate limiting if configured
	if options.RateLimiter != nil {
		if err := options.RateLimiter.Acquire(ctx); err != nil {
			return "", fmt.Errorf("rate limiter error: %w", err)
		}
		defer options.RateLimiter.Release()
	}

	// Create retry context with timeout
	retryCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Retry loop
	for retryCount = 0; retryCount <= options.RetryCount; retryCount++ {
		// Check if context is cancelled
		if retryCtx.Err() != nil {
			return "", fmt.Errorf("context cancelled: %w", retryCtx.Err())
		}

		// Execute template
		response, err = provider.SendPrompt(retryCtx, template.Test.Prompt, options.ProviderOptions)
		if err == nil {
			// Success
			break
		}

		// Check if we should retry
		if retryCount >= options.RetryCount {
			return "", fmt.Errorf("failed after %d retries: %w", retryCount, err)
		}

		// Wait before retrying
		select {
		case <-retryCtx.Done():
			return "", fmt.Errorf("context cancelled during retry: %w", retryCtx.Err())
		case <-time.After(options.RetryDelay):
			// Continue with retry
		}
	}

	// Update retry stats
	if retryCount > 0 {
		e.statsMutex.Lock()
		e.stats.RetryCount += int64(retryCount)
		e.statsMutex.Unlock()
	}

	return response, nil

// mergeOptions merges user options with default options
func (e *OptimizedTemplateExecutor) mergeOptions(userOptions map[string]interface{}) *ExecutionOptions {
	// Create a copy of default options
	options := &ExecutionOptions{
		Provider:        e.defaultOptions.Provider,
		DetectionEngine: e.defaultOptions.DetectionEngine,
		RateLimiter:     e.defaultOptions.RateLimiter,
		Timeout:         e.defaultOptions.Timeout,
		RetryCount:      e.defaultOptions.RetryCount,
		RetryDelay:      e.defaultOptions.RetryDelay,
		MaxConcurrent:   e.defaultOptions.MaxConcurrent,
		Variables:       make(map[string]interface{}),
		ProviderOptions: make(map[string]interface{}),
	}

	// Copy default variables
	for k, v := range e.defaultOptions.Variables {
		options.Variables[k] = v
	}

	// Copy default provider options
	for k, v := range e.defaultOptions.ProviderOptions {
		options.ProviderOptions[k] = v
	}

	// Override with user options
	if userOptions != nil {
		// Check for provider override
		if providerName, ok := userOptions["provider"].(string); ok {
			if provider, exists := e.providers[providerName]; exists {
				options.Provider = provider
			}
		}

		// Check for detection engine override
		if engineName, ok := userOptions["detection_engine"].(string); ok {
			if engine, exists := e.detectionEngines[engineName]; exists {
				options.DetectionEngine = engine
			}
		}

		// Check for timeout override
		if timeout, ok := userOptions["timeout"].(time.Duration); ok {
			options.Timeout = timeout
		}

		// Check for retry count override
		if retryCount, ok := userOptions["retry_count"].(int); ok {
			options.RetryCount = retryCount
		}

		// Check for retry delay override
		if retryDelay, ok := userOptions["retry_delay"].(time.Duration); ok {
			options.RetryDelay = retryDelay
		}

		// Check for variables override
		if variables, ok := userOptions["variables"].(map[string]interface{}); ok {
			for k, v := range variables {
				options.Variables[k] = v
			}
		}

		// Check for provider options override
		if providerOptions, ok := userOptions["provider_options"].(map[string]interface{}); ok {
			for k, v := range providerOptions {
				options.ProviderOptions[k] = v
			}
		}
	}

	return options

// generateCacheKey generates a cache key for a template and options
func (e *OptimizedTemplateExecutor) generateCacheKey(template *format.Template, options *ExecutionOptions) string {
	// Create a cache key from template ID, prompt, and relevant options
	keyData := struct {
		TemplateID      string
		Prompt          string
		ProviderName    string
		ProviderOptions map[string]interface{}
		Variables       map[string]interface{}
	}{
		TemplateID:      template.ID,
		Prompt:          template.Test.Prompt,
		ProviderOptions: options.ProviderOptions,
		Variables:       options.Variables,
	}

	// Get provider name
	if options.Provider != nil {
		keyData.ProviderName = options.Provider.GetName()
	}

	// Convert to JSON
	jsonData, err := json.Marshal(keyData)
	if err != nil {
		// Fallback to simple key if marshaling fails
		return fmt.Sprintf("%s-%s", template.ID, template.Test.Prompt)
	}

	// Generate MD5 hash
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])

// getCachedResponse gets a response from the cache
func (e *OptimizedTemplateExecutor) getCachedResponse(key string) (string, bool) {
	e.responseCache.mutex.RLock()
	defer e.responseCache.mutex.RUnlock()

	e.responseCache.stats.TotalLookups++

	entry, exists := e.responseCache.cache[key]
	if !exists {
		e.responseCache.stats.Misses++
		return "", false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		delete(e.responseCache.cache, key)
		e.responseCache.stats.Misses++
		return "", false
	}

	e.responseCache.stats.Hits++
	return entry.Response, true

// cacheResponse caches a response
func (e *OptimizedTemplateExecutor) cacheResponse(key string, response string, options map[string]interface{}) {
	// Check if caching is disabled
	if disableCache, ok := options["disable_cache"].(bool); ok && disableCache {
		return
	}

	e.responseCache.mutex.Lock()
	defer e.responseCache.mutex.Unlock()

	// Check if cache is full
	if len(e.responseCache.cache) >= e.responseCache.maxSize {
		// Evict oldest entry
		var oldestKey string
		var oldestTime time.Time

		// Find oldest entry
		for k, entry := range e.responseCache.cache {
			if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = entry.CreatedAt
			}
		}

		// Evict oldest entry
		if oldestKey != "" {
			delete(e.responseCache.cache, oldestKey)
			e.responseCache.stats.Evictions++
		}
	}

	// Create cache entry
	entry := &ResponseCacheEntry{
		Response:  response,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(e.responseCache.ttl),
		Size:      len(response),
	}

	// Add to cache
	e.responseCache.cache[key] = entry

// ClearCache clears the response cache
func (e *OptimizedTemplateExecutor) ClearCache() {
	e.responseCache.mutex.Lock()
	defer e.responseCache.mutex.Unlock()

	e.responseCache.cache = make(map[string]*ResponseCacheEntry)

// GetCacheStats returns statistics about the cache
func (e *OptimizedTemplateExecutor) GetCacheStats() map[string]interface{} {
	e.responseCache.mutex.RLock()
	defer e.responseCache.mutex.RUnlock()

	hitRate := float64(0)
	if e.responseCache.stats.TotalLookups > 0 {
		hitRate = float64(e.responseCache.stats.Hits) / float64(e.responseCache.stats.TotalLookups) * 100
	}

	return map[string]interface{}{
		"size":      len(e.responseCache.cache),
		"max_size":  e.responseCache.maxSize,
		"hits":      e.responseCache.stats.Hits,
		"misses":    e.responseCache.stats.Misses,
		"evictions": e.responseCache.stats.Evictions,
		"hit_rate":  hitRate,
	}

// GetExecutionStats returns statistics about the executor
func (e *OptimizedTemplateExecutor) GetExecutionStats() map[string]interface{} {
	e.statsMutex.RLock()
	defer e.statsMutex.RUnlock()

	avgExecutionTime := time.Duration(0)
	if e.stats.TotalExecutions > 0 {
		avgExecutionTime = e.stats.TotalExecutionTime / time.Duration(e.stats.TotalExecutions)
	}

	successRate := float64(0)
	if e.stats.TotalExecutions > 0 {
		successRate = float64(e.stats.SuccessfulExecutions) / float64(e.stats.TotalExecutions) * 100
	}

	cacheRate := float64(0)
	if e.stats.SuccessfulExecutions > 0 {
		cacheRate = float64(e.stats.CachedResponses) / float64(e.stats.SuccessfulExecutions) * 100
	}

	return map[string]interface{}{
		"total_executions":      e.stats.TotalExecutions,
		"successful_executions": e.stats.SuccessfulExecutions,
		"failed_executions":     e.stats.FailedExecutions,
		"cached_responses":      e.stats.CachedResponses,
		"total_execution_time":  e.stats.TotalExecutionTime,
		"avg_execution_time":    avgExecutionTime,
		"success_rate":          successRate,
		"cache_rate":            cacheRate,
		"retry_count":           e.stats.RetryCount,
	}

// GetProviders returns the list of registered providers
func (e *OptimizedTemplateExecutor) GetProviders() []string {
	providers := make([]string, 0, len(e.providers))
	for name := range e.providers {
		providers = append(providers, name)
	}
	return providers

// GetDetectionEngines returns the list of registered detection engines
func (e *OptimizedTemplateExecutor) GetDetectionEngines() []string {
	engines := make([]string, 0, len(e.detectionEngines))
	for name := range e.detectionEngines {
		engines = append(engines, name)
	}
	return engines

// SetMaxConcurrent sets the maximum number of concurrent executions
func (e *OptimizedTemplateExecutor) SetMaxConcurrent(max int) {
	if max <= 0 {
		return
	}
	e.defaultOptions.MaxConcurrent = max

// SetCacheTTL sets the time-to-live for cache entries
func (e *OptimizedTemplateExecutor) SetCacheTTL(ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	e.responseCache.ttl = ttl

// SetCacheSize sets the maximum size of the cache
func (e *OptimizedTemplateExecutor) SetCacheSize(size int) {
	if size <= 0 {
		return
	}
	e.responseCache.maxSize = size

// SetWorkerPoolSize sets the size of the worker pool
func (e *OptimizedTemplateExecutor) SetWorkerPoolSize(size int) {
	if size <= 0 || size == e.workPool.size {
		return
	}

	// Create new work pool
	oldPool := e.workPool
	e.workPool = newWorkPool(size)

	// Close old pool
	close(oldPool.workers)
