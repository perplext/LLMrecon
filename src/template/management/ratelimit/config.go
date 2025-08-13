// Package ratelimit provides rate limiting functionality for template execution
package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-redis/redis/v8"
)

// RateLimitConfig represents the configuration for rate limiting
type RateLimitConfig struct {
	// Global rate limit settings
	Global struct {
		// QPS is the queries per second limit for the entire system
		QPS float64 `json:"qps"`
		
		// Burst is the maximum burst size for the entire system
		Burst int `json:"burst"`
	} `json:"global"`
	
	// Default user rate limit settings
	DefaultUser struct {
		// QPS is the default queries per second limit for users
		QPS float64 `json:"qps"`
		
		// Burst is the default maximum burst size for users
		Burst int `json:"burst"`
		
		// MaxTokens is the default maximum number of tokens a user can consume
		MaxTokens int `json:"max_tokens"`
		
		// ResetInterval is the default interval at which user tokens are reset
		ResetInterval string `json:"reset_interval"`
	} `json:"default_user"`
	
	// User-specific rate limit policies
	UserPolicies []UserPolicyConfig `json:"user_policies"`
	
	// Distributed rate limiting settings
	Distributed struct {
		// Enabled indicates whether distributed rate limiting is enabled
		Enabled bool `json:"enabled"`
		
		// Redis connection settings
		Redis struct {
			// Address is the Redis server address
			Address string `json:"address"`
			
			// Password is the Redis server password
			Password string `json:"password"`
			
			// DB is the Redis database number
			DB int `json:"db"`
			
			// KeyPrefix is the prefix for Redis keys
			KeyPrefix string `json:"key_prefix"`
			
			// EnableTLS indicates whether to use TLS for Redis connection
			EnableTLS bool `json:"enable_tls"`
		} `json:"redis"`
		
		// LocalFallback indicates whether to fall back to local rate limiting if Redis is unavailable
		LocalFallback bool `json:"local_fallback"`
	} `json:"distributed"`
	
	// Dynamic policy update settings
	PolicyUpdates struct {
		// Enabled indicates whether dynamic policy updates are enabled
		Enabled bool `json:"enabled"`
		
		// Interval is the interval at which to check for policy updates
		Interval string `json:"interval"`
		
		// FilePath is the path to the policy file
		FilePath string `json:"file_path"`
		
		// RedisChannel is the Redis channel to subscribe to for policy updates
		RedisChannel string `json:"redis_channel"`
	} `json:"policy_updates"`
	
	// Advanced settings
	Advanced struct {
		// FairnessEnabled indicates whether fairness mechanisms are enabled
		FairnessEnabled bool `json:"fairness_enabled"`
		
		// DynamicAdjustmentEnabled indicates whether dynamic adjustment is enabled
		DynamicAdjustmentEnabled bool `json:"dynamic_adjustment_enabled"`
		
		// StatsEnabled indicates whether statistics collection is enabled
		StatsEnabled bool `json:"stats_enabled"`
		
		// MaxStatsEvents is the maximum number of events to keep in the stats collector
		MaxStatsEvents int `json:"max_stats_events"`
	} `json:"advanced"`
}

// UserPolicyConfig represents a user-specific rate limit policy in the configuration
type UserPolicyConfig struct {
	// UserID is the ID of the user
	UserID string `json:"user_id"`
	
	// QPS is the queries per second limit for the user
	QPS float64 `json:"qps"`
	
	// Burst is the maximum burst size for the user
	Burst int `json:"burst"`
	
	// Priority is the priority of the user (higher = more priority)
	Priority int `json:"priority"`
	
	// MaxTokens is the maximum number of tokens the user can consume
	MaxTokens int `json:"max_tokens"`
	
	// ResetInterval is the interval at which the user's tokens are reset
	ResetInterval string `json:"reset_interval"`
}

// DefaultConfig returns the default rate limit configuration
func DefaultConfig() *RateLimitConfig {
	config := &RateLimitConfig{}
	
	// Set default global settings
	config.Global.QPS = 1000
	config.Global.Burst = 100
	
	// Set default user settings
	config.DefaultUser.QPS = 10
	config.DefaultUser.Burst = 5
	config.DefaultUser.MaxTokens = 600
	config.DefaultUser.ResetInterval = "1m"
	
	// Set default user policies
	config.UserPolicies = []UserPolicyConfig{
		{
			UserID:        "admin",
			QPS:           100,
			Burst:         50,
			Priority:      10,
			MaxTokens:     6000,
			ResetInterval: "1h",
		},
		{
			UserID:        "service",
			QPS:           50,
			Burst:         25,
			Priority:      8,
			MaxTokens:     3000,
			ResetInterval: "10m",
		},
	}
	
	// Set default distributed settings
	config.Distributed.Enabled = false
	config.Distributed.Redis.Address = "localhost:6379"
	config.Distributed.Redis.KeyPrefix = "ratelimit"
	config.Distributed.LocalFallback = true
	
	// Set default policy update settings
	config.PolicyUpdates.Enabled = false
	config.PolicyUpdates.Interval = "1m"
	config.PolicyUpdates.FilePath = "config/rate_limit_policies.json"
	
	// Set default advanced settings
	config.Advanced.FairnessEnabled = true
	config.Advanced.DynamicAdjustmentEnabled = true
	config.Advanced.StatsEnabled = true
	config.Advanced.MaxStatsEvents = 1000
	
	return config
}

// LoadConfig loads the rate limit configuration from a file
func LoadConfig(filePath string) (*RateLimitConfig, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create default config
		config := DefaultConfig()
		
		// Create parent directory if needed
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
		
		// Write default config to file
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default config: %w", err)
		}
		
		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
		
		return config, nil
	}
	
	// Read and parse the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config RateLimitConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return &config, nil
}

// CreateLimiterFromConfig creates a rate limiter from the configuration
func CreateLimiterFromConfig(config *RateLimitConfig) (interface{}, error) {
	var limiter interface{}
	
	// Create the appropriate limiter based on configuration
	if config.Distributed.Enabled {
		// Create Redis client
		redisOptions := &redis.Options{
			Addr:     config.Distributed.Redis.Address,
			Password: config.Distributed.Redis.Password,
			DB:       config.Distributed.Redis.DB,
		}
		
		client := redis.NewClient(redisOptions)
		
		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := client.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %w", err)
		}
		
		// Create distributed limiter
		distLimiter := NewDistributedLimiter(
			client,
			config.Distributed.Redis.KeyPrefix,
			config.Global.QPS,
			config.Global.Burst,
			config.DefaultUser.QPS,
			config.DefaultUser.Burst,
		)
		
		distLimiter.EnableLocalFallback(config.Distributed.LocalFallback)
		limiter = distLimiter
	} else {
		// Create adaptive limiter
		adaptiveLimiter := NewAdaptiveLimiter(
			config.Global.QPS,
			config.Global.Burst,
			config.DefaultUser.QPS,
			config.DefaultUser.Burst,
		)
		
		adaptiveLimiter.EnableFairness(config.Advanced.FairnessEnabled)
		adaptiveLimiter.EnableDynamicAdjustment(config.Advanced.DynamicAdjustmentEnabled)
		limiter = adaptiveLimiter
	}
	
	// Apply user policies
	for _, policyConfig := range config.UserPolicies {
		// Parse reset interval
		resetInterval, err := time.ParseDuration(policyConfig.ResetInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid reset interval for user %s: %w", policyConfig.UserID, err)
		}
		
		policy := &UserRateLimitPolicy{
			UserID:        policyConfig.UserID,
			QPS:           policyConfig.QPS,
			Burst:         policyConfig.Burst,
			Priority:      policyConfig.Priority,
			MaxTokens:     policyConfig.MaxTokens,
			ResetInterval: resetInterval,
		}
		
		// Set the policy based on the limiter type
		if distLimiter, ok := limiter.(*DistributedLimiter); ok {
			distLimiter.SetUserPolicy(policy)
		} else if adaptiveLimiter, ok := limiter.(*AdaptiveLimiter); ok {
			adaptiveLimiter.SetUserPolicy(policy)
		}
	}
	
	// Set up policy updater if enabled
	if config.PolicyUpdates.Enabled {
		var updater *PolicyUpdater
		
		// Create updater based on the limiter type
		if distLimiter, ok := limiter.(*DistributedLimiter); ok {
			updater = NewPolicyUpdater(distLimiter)
		} else if adaptiveLimiter, ok := limiter.(*AdaptiveLimiter); ok {
			updater = NewPolicyUpdater(adaptiveLimiter)
		}
		
		// Parse update interval
		updateInterval, err := time.ParseDuration(config.PolicyUpdates.Interval)
		if err != nil {
			return nil, fmt.Errorf("invalid policy update interval: %w", err)
		}
		
		updater.SetUpdateInterval(updateInterval)
		
		// Add file policy source if configured
		if config.PolicyUpdates.FilePath != "" {
			filePath := config.PolicyUpdates.FilePath
			
			// Create default policy file if it doesn't exist
			if err := CreateDefaultPolicyFile(filePath); err != nil {
				return nil, fmt.Errorf("failed to create default policy file: %w", err)
			}
			
			updater.AddSource(NewFilePolicySource(filePath))
		}
		
		// Add Redis channel policy source if configured and distributed is enabled
		if config.PolicyUpdates.RedisChannel != "" && config.Distributed.Enabled {
			if distLimiter, ok := limiter.(*DistributedLimiter); ok {
				updater.AddSource(NewRedisChannelPolicySource(distLimiter.client, config.PolicyUpdates.RedisChannel))
			}
		}
		
		// Start the updater
		updater.Start()
	}
	
	return limiter, nil
}

// ParseUserPolicyFromConfig parses a user policy from the configuration
func ParseUserPolicyFromConfig(policyConfig UserPolicyConfig) (*UserRateLimitPolicy, error) {
	// Parse reset interval
	resetInterval, err := time.ParseDuration(policyConfig.ResetInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid reset interval: %w", err)
	}
	
	return &UserRateLimitPolicy{
		UserID:        policyConfig.UserID,
		QPS:           policyConfig.QPS,
		Burst:         policyConfig.Burst,
		Priority:      policyConfig.Priority,
		MaxTokens:     policyConfig.MaxTokens,
		ResetInterval: resetInterval,
	}, nil
}
