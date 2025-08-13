// Package ratelimit provides rate limiting functionality for template execution
package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"github.com/go-redis/redis/v8"
)

// PolicySource defines the interface for policy update sources
type PolicySource interface {
	// GetPolicies returns the current set of rate limit policies
	GetPolicies(ctx context.Context) ([]*UserRateLimitPolicy, error)
}

// FilePolicySource implements PolicySource using a JSON file
type FilePolicySource struct {
	// Path to the policy file
	filePath string
	
	// Last modified time of the file
	lastModified time.Time
}

// NewFilePolicySource creates a new file-based policy source
func NewFilePolicySource(filePath string) *FilePolicySource {
	return &FilePolicySource{
		filePath: filePath,
	}
}

// GetPolicies reads policies from the JSON file
func (s *FilePolicySource) GetPolicies(ctx context.Context) ([]*UserRateLimitPolicy, error) {
	// Check if file exists
	info, err := os.Stat(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat policy file: %w", err)
	}
	
	// Check if file has been modified
	if info.ModTime().Equal(s.lastModified) {
		// File hasn't changed, return empty slice to indicate no updates
		return []*UserRateLimitPolicy{}, nil
	}
	
	// Read and parse the file
	data, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}
	
	var policies []*UserRateLimitPolicy
	if err := json.Unmarshal(data, &policies); err != nil {
		return nil, fmt.Errorf("failed to parse policy file: %w", err)
	}
	
	// Update last modified time
	s.lastModified = info.ModTime()
	
	return policies, nil
}

// RedisChannelPolicySource implements PolicySource using Redis pub/sub
type RedisChannelPolicySource struct {
	// Redis client
	client redis.UniversalClient
	
	// Channel to subscribe to for policy updates
	channel string
	
	// Mutex for thread safety
	mu sync.Mutex
	
	// Latest policies received
	latestPolicies []*UserRateLimitPolicy
	
	// Whether new policies have been received
	hasNewPolicies bool
}

// NewRedisChannelPolicySource creates a new Redis-based policy source
func NewRedisChannelPolicySource(client redis.UniversalClient, channel string) *RedisChannelPolicySource {
	source := &RedisChannelPolicySource{
		client:  client,
		channel: channel,
	}
	
	// Start subscription in background
	go source.subscribe()
	
	return source
}

// subscribe subscribes to the Redis channel for policy updates
func (s *RedisChannelPolicySource) subscribe() {
	ctx := context.Background()
	pubsub := s.client.Subscribe(ctx, s.channel)
	defer pubsub.Close()
	
	// Listen for messages
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Error receiving message from Redis: %v", err)
			time.Sleep(time.Second)
			continue
		}
		
		// Parse the message as JSON
		var policies []*UserRateLimitPolicy
		if err := json.Unmarshal([]byte(msg.Payload), &policies); err != nil {
			log.Printf("Error parsing policy update: %v", err)
			continue
		}
		
		// Update latest policies
		s.mu.Lock()
		s.latestPolicies = policies
		s.hasNewPolicies = true
		s.mu.Unlock()
	}
}

// GetPolicies returns the latest policies received from Redis
func (s *RedisChannelPolicySource) GetPolicies(ctx context.Context) ([]*UserRateLimitPolicy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.hasNewPolicies {
		// No new policies
		return []*UserRateLimitPolicy{}, nil
	}
	
	// Reset flag and return policies
	s.hasNewPolicies = false
	return s.latestPolicies, nil
}

// PolicyUpdater manages dynamic updates to rate limit policies
type PolicyUpdater struct {
	// Target limiter to update
	limiter interface {
		SetUserPolicy(*UserRateLimitPolicy)
		GetUserPolicy(string) *UserRateLimitPolicy
	}
	
	// Policy sources
	sources []PolicySource
	
	// Update interval
	updateInterval time.Duration
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
	
	// Wait group for graceful shutdown
	wg sync.WaitGroup
	
	// Mutex for thread safety
	mu sync.Mutex
	
	// Whether the updater is running
	running bool
	
	// Callback for policy updates
	onUpdate func([]*UserRateLimitPolicy)
}

// NewPolicyUpdater creates a new policy updater
func NewPolicyUpdater(limiter interface {
	SetUserPolicy(*UserRateLimitPolicy)
	GetUserPolicy(string) *UserRateLimitPolicy
}) *PolicyUpdater {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &PolicyUpdater{
		limiter:        limiter,
		sources:        []PolicySource{},
		updateInterval: time.Minute,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// AddSource adds a policy source
func (u *PolicyUpdater) AddSource(source PolicySource) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.sources = append(u.sources, source)
}

// SetUpdateInterval sets the update interval
func (u *PolicyUpdater) SetUpdateInterval(interval time.Duration) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.updateInterval = interval
}

// SetUpdateCallback sets a callback function to be called when policies are updated
func (u *PolicyUpdater) SetUpdateCallback(callback func([]*UserRateLimitPolicy)) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.onUpdate = callback
}

// Start starts the policy updater
func (u *PolicyUpdater) Start() {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	if u.running {
		return
	}
	
	u.running = true
	u.wg.Add(1)
	
	go u.run()
}

// Stop stops the policy updater
func (u *PolicyUpdater) Stop() {
	u.mu.Lock()
	if !u.running {
		u.mu.Unlock()
		return
	}
	u.running = false
	u.mu.Unlock()
	
	u.cancel()
	u.wg.Wait()
}

// run runs the policy updater loop
func (u *PolicyUpdater) run() {
	defer u.wg.Done()
	
	ticker := time.NewTicker(u.updateInterval)
	defer ticker.Stop()
	
	// Do an initial update
	u.updatePolicies()
	
	for {
		select {
		case <-ticker.C:
			u.updatePolicies()
		case <-u.ctx.Done():
			return
		}
	}
}

// updatePolicies updates policies from all sources
func (u *PolicyUpdater) updatePolicies() {
	u.mu.Lock()
	sources := u.sources
	callback := u.onUpdate
	u.mu.Unlock()
	
	ctx, cancel := context.WithTimeout(u.ctx, 10*time.Second)
	defer cancel()
	
	var updatedPolicies []*UserRateLimitPolicy
	
	// Get policies from all sources
	for _, source := range sources {
		policies, err := source.GetPolicies(ctx)
		if err != nil {
			log.Printf("Error getting policies: %v", err)
			continue
		}
		
		// Skip if no new policies
		if len(policies) == 0 {
			continue
		}
		
		// Apply policies
		for _, policy := range policies {
			u.limiter.SetUserPolicy(policy)
		}
		
		// Add to updated policies
		updatedPolicies = append(updatedPolicies, policies...)
	}
	
	// Call callback if policies were updated
	if len(updatedPolicies) > 0 && callback != nil {
		callback(updatedPolicies)
	}
}

// CreateDefaultPolicyFile creates a default policy file if it doesn't exist
func CreateDefaultPolicyFile(filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, don't overwrite
		return nil
	}
	
	// Create parent directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Create default policies
	defaultPolicies := []*UserRateLimitPolicy{
		{
			UserID:        "admin",
			QPS:           100,
			Burst:         50,
			Priority:      10,
			MaxTokens:     1000,
			ResetInterval: time.Hour,
		},
		{
			UserID:        "user",
			QPS:           10,
			Burst:         5,
			Priority:      5,
			MaxTokens:     100,
			ResetInterval: time.Minute * 10,
		},
		{
			UserID:        "guest",
			QPS:           1,
			Burst:         2,
			Priority:      1,
			MaxTokens:     10,
			ResetInterval: time.Minute,
		},
	}
	
	// Convert to JSON
	data, err := json.MarshalIndent(defaultPolicies, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default policies: %w", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write default policies: %w", err)
	}
	
	return nil
}
