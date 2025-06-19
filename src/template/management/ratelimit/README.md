# Adaptive Rate Limiting

This package provides an advanced rate limiting system with priority-based fairness for the template management system. It ensures system stability under high load while providing preferential treatment to high-priority operations.

## Features

- **Global and User-Specific Rate Limiting**: Control both overall system throughput and individual user limits
- **Priority-Based Fairness**: Ensure critical operations continue during high load conditions
- **Dynamic Adjustment**: Automatically scale limits based on system load
- **Token Bucket Tracking**: Prevent abuse by limiting total resource consumption
- **Configurable Policies**: Set custom limits and priorities for different users

## Components

### AdaptiveLimiter

The main rate limiting component that implements both global and user-specific rate limiting with priority-based fairness.

```go
// Create a new adaptive limiter
limiter := ratelimit.NewAdaptiveLimiter(
    100,  // Global QPS (queries per second)
    50,   // Global burst size
    10,   // Default user QPS
    5,    // Default user burst size
)
```

### UserRateLimitPolicy

Defines rate limiting policies for specific users, including QPS limits, burst size, priority, and token allocation.

```go
// Set a policy for a specific user
limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
    UserID:        "user123",
    QPS:           20,        // 20 requests per second
    Burst:         10,        // Allow bursts of up to 10 requests
    Priority:      8,         // High priority (1-10 scale)
    MaxTokens:     1000,      // Maximum tokens user can consume
    ResetInterval: time.Hour, // Reset token count hourly
})
```

## Usage Examples

### Basic Rate Limiting

```go
// Create a context
ctx := context.Background()

// Acquire a token for a specific user
err := limiter.AcquireForUser(ctx, "user123")
if err != nil {
    // Handle rate limit exceeded error
    return err
}

// Proceed with the rate-limited operation
// ...
```

### Setting Load Factor

```go
// Simulate high load (reduce capacity to 30%)
limiter.SetLoadFactor(0.3)

// Return to normal operation
limiter.SetLoadFactor(1.0)
```

### Enabling/Disabling Features

```go
// Enable/disable dynamic adjustment
limiter.EnableDynamicAdjustment(true)

// Enable/disable fairness mechanisms
limiter.EnableFairness(true)
```

### Tracking and Resetting Usage

```go
// Get current usage for a user
usage := limiter.GetUserUsage("user123")

// Reset usage for a specific user
limiter.ResetUserUsage("user123")

// Reset usage for all users
limiter.ResetAllUserUsage()
```

## Priority Levels

Priority values typically range from 1 (lowest) to 10 (highest):

- **10**: Critical system operations that must never be throttled
- **8-9**: High-priority administrative operations
- **5-7**: Normal authenticated user operations
- **3-4**: Low-priority background operations
- **1-2**: Lowest priority maintenance tasks

## Integration with Template Management

The rate limiting system is designed to be used with the template management system:

```go
// Create a template executor with rate limiting
executor := execution.NewTemplateExecutor(
    repository,
    renderer,
    limiter, // Pass the rate limiter
)

// Execute a template with rate limiting
result, err := executor.ExecuteTemplate(ctx, template, data, userID)
```

## Best Practices

1. **Set appropriate global limits** based on your system's capacity
2. **Assign priorities thoughtfully** to ensure critical operations continue during high load
3. **Monitor rate limiting metrics** to identify potential issues
4. **Adjust load factor dynamically** based on system health indicators
5. **Set reasonable token limits** to prevent abuse while allowing legitimate usage patterns

## Testing

The package includes comprehensive tests for all rate limiting functionality:

```bash
# Run all tests
go test -v ./...

# Run specific tests
go test -v -run "TestPriority|TestDynamic"
```
