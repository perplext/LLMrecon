# Integrating Rate Limiting with Template Management

This guide explains how to integrate the rate limiting system with the template management system to control the rate of template executions and ensure system stability under high load.

## Overview

The rate limiting system provides two main implementations:

1. **TokenBucketLimiter**: A simple rate limiter with global and user-specific limits
2. **AdaptiveLimiter**: An advanced rate limiter with priority-based fairness and dynamic adjustment

Both implementations can be used with the template management system, but the `AdaptiveLimiter` is recommended for production environments where fairness and priority handling are important.

## Basic Integration

### Step 1: Create a Rate Limiter

```go
import (
    "github.com/perplext/LLMrecon/src/template/management/ratelimit"
)

// Create an adaptive limiter with appropriate limits
limiter := ratelimit.NewAdaptiveLimiter(
    100,  // Global QPS (queries per second)
    50,   // Global burst size
    10,   // Default user QPS
    5,    // Default user burst size
)
```

### Step 2: Configure User Policies

```go
// Set policies for different user types
limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
    UserID:        "admin",
    QPS:           20,        // 20 requests per second
    Burst:         10,        // Allow bursts of up to 10 requests
    Priority:      10,        // Highest priority
    MaxTokens:     1000,      // Maximum tokens user can consume
    ResetInterval: time.Hour, // Reset token count hourly
})

limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
    UserID:        "standard-user",
    QPS:           10,
    Burst:         5,
    Priority:      5,         // Medium priority
    MaxTokens:     500,
    ResetInterval: time.Hour,
})

limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
    UserID:        "api-client",
    QPS:           5,
    Burst:         3,
    Priority:      3,         // Lower priority
    MaxTokens:     300,
    ResetInterval: time.Hour,
})
```

### Step 3: Integrate with Template Executor

```go
import (
    "github.com/perplext/LLMrecon/src/template/management/execution"
    "github.com/perplext/LLMrecon/src/template/management/ratelimit"
)

// Create a template executor with rate limiting
executor := execution.NewTemplateExecutor(
    repository,
    renderer,
    limiter, // Pass the rate limiter
)

// Execute a template with rate limiting
result, err := executor.ExecuteTemplate(ctx, template, data, userID)
if err != nil {
    // Handle error (including rate limit exceeded)
    return err
}
```

## Advanced Integration

### Dynamic Load Factor Adjustment

For production systems, you may want to dynamically adjust the load factor based on system metrics:

```go
// Monitor system metrics and adjust load factor
go func() {
    for {
        // Get current system metrics (CPU, memory, etc.)
        cpuUsage := getSystemCPUUsage()
        
        // Calculate load factor (example formula)
        // Higher CPU usage = lower load factor
        loadFactor := 1.0 - (cpuUsage / 100.0 * 0.7)
        
        // Ensure reasonable bounds
        if loadFactor < 0.1 {
            loadFactor = 0.1 // Minimum 10% capacity
        } else if loadFactor > 1.5 {
            loadFactor = 1.5 // Maximum 150% capacity
        }
        
        // Update the limiter
        limiter.SetLoadFactor(loadFactor)
        
        // Check periodically
        time.Sleep(5 * time.Second)
    }
}()
```

### User Group Policies

For larger systems, you may want to manage policies by user group rather than individual users:

```go
// Map of user groups to policies
userGroups := map[string]string{
    "user123": "standard",
    "user456": "premium",
    "user789": "admin",
}

// Group policies
groupPolicies := map[string]*ratelimit.UserRateLimitPolicy{
    "standard": {
        QPS:           5,
        Burst:         3,
        Priority:      3,
        MaxTokens:     300,
        ResetInterval: time.Hour,
    },
    "premium": {
        QPS:           15,
        Burst:         7,
        Priority:      7,
        MaxTokens:     700,
        ResetInterval: time.Hour,
    },
    "admin": {
        QPS:           30,
        Burst:         15,
        Priority:      10,
        MaxTokens:     1500,
        ResetInterval: time.Hour,
    },
}

// Apply policies based on user group
func applyUserPolicy(userID string, limiter *ratelimit.AdaptiveLimiter) {
    group, exists := userGroups[userID]
    if !exists {
        group = "standard" // Default group
    }
    
    policy := groupPolicies[group]
    policy.UserID = userID // Set the specific user ID
    
    limiter.SetUserPolicy(policy)
}
```

### Error Handling

Proper error handling for rate limiting is important for a good user experience:

```go
func executeTemplateWithRateLimiting(ctx context.Context, executor *execution.TemplateExecutor, 
                                    template *models.Template, data interface{}, userID string) (string, error) {
    result, err := executor.ExecuteTemplate(ctx, template, data, userID)
    if err != nil {
        // Check if it's a rate limit error
        if strings.Contains(err.Error(), "rate limit exceeded") {
            // Log the rate limiting event
            log.Printf("Rate limit exceeded for user %s", userID)
            
            // Return a user-friendly error
            return "", fmt.Errorf("rate limit exceeded, please try again in a few moments")
        }
        
        // Handle other errors
        return "", err
    }
    
    return result, nil
}
```

## Monitoring and Metrics

For production systems, it's important to monitor rate limiting events:

```go
// Track rate limiting metrics
type RateLimitMetrics struct {
    GlobalLimitExceeded    int64
    UserLimitExceeded      map[string]int64
    SuccessfulAcquisitions int64
    mu                     sync.Mutex
}

func NewRateLimitMetrics() *RateLimitMetrics {
    return &RateLimitMetrics{
        UserLimitExceeded: make(map[string]int64),
    }
}

func (m *RateLimitMetrics) TrackAcquisition(userID string, err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if err == nil {
        m.SuccessfulAcquisitions++
        return
    }
    
    if strings.Contains(err.Error(), "global rate limit exceeded") {
        m.GlobalLimitExceeded++
    } else if strings.Contains(err.Error(), "user rate limit exceeded") {
        m.UserLimitExceeded[userID]++
    }
}

// Use the metrics in your application
metrics := NewRateLimitMetrics()

// When acquiring tokens
err := limiter.AcquireForUser(ctx, userID)
metrics.TrackAcquisition(userID, err)
```

## Best Practices

1. **Set appropriate limits** based on your system's capacity and expected load
2. **Use priority levels consistently** across your application
3. **Monitor rate limiting events** to identify potential issues
4. **Provide clear feedback** to users when rate limits are exceeded
5. **Adjust load factors dynamically** based on system health indicators
6. **Reset user usage periodically** to prevent long-term throttling
7. **Test rate limiting under load** to ensure it behaves as expected

## Troubleshooting

### Common Issues

1. **Too many rate limit errors**
   - Check if global or user limits are set too low
   - Consider increasing burst sizes for bursty workloads
   - Verify that load factor is appropriate for system capacity

2. **High-priority operations being throttled**
   - Ensure priority values are set correctly
   - Check if fairness mechanisms are enabled
   - Verify that the load factor isn't too low

3. **Uneven distribution of resources**
   - Review priority assignments across user groups
   - Consider adjusting the priority factor in dynamic adjustment
   - Check if token usage is being reset appropriately

4. **System overload despite rate limiting**
   - Decrease global QPS limit
   - Reduce default user limits
   - Implement more aggressive dynamic adjustment

## Conclusion

Proper integration of the rate limiting system with the template management system helps ensure system stability and fair resource allocation. By using the `AdaptiveLimiter` with appropriate policies and dynamic adjustment, you can maintain good performance even under high load conditions while ensuring that high-priority operations continue to function.
