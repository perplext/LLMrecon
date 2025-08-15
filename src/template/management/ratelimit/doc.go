// Package ratelimit provides adaptive rate limiting functionality for template execution.
package ratelimit

/*
Package ratelimit provides adaptive rate limiting functionality for template execution.

This package implements a sophisticated rate limiting system with the following key features:

1. Global and User-Specific Rate Limiting
   - Control overall system throughput with global limits
   - Apply user-specific limits based on individual policies

2. Priority-Based Fairness
   - Ensure high-priority users receive preferential treatment during high load
   - Maintain fair access for all users under normal conditions

3. Dynamic Adjustment
   - Automatically adjust limits based on system load
   - Scale limits up or down based on user priority and historical usage

4. Token Bucket Tracking
   - Prevent abuse by limiting total resource consumption
   - Reset token usage at configurable intervals

# Basic Usage

Here's a simple example of how to use the rate limiting system:

    // Create a new adaptive limiter with global and default user limits
    limiter := ratelimit.NewAdaptiveLimiter(
        100,  // Global QPS (queries per second)
        50,   // Global burst size
        10,   // Default user QPS
        5,    // Default user burst size
    )

    // Set a policy for a specific user
    limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
        UserID:        "user123",
        QPS:           20,        // 20 requests per second
        Burst:         10,        // Allow bursts of up to 10 requests
        Priority:      8,         // High priority (1-10 scale)
        MaxTokens:     1000,      // Maximum tokens user can consume
        ResetInterval: time.Hour, // Reset token count hourly
    })

    // Acquire a token for a specific user
    ctx := context.Background()
    err := limiter.AcquireForUser(ctx, "user123")
    if err != nil {
        // Handle rate limit exceeded error
        return err
    }

    // Proceed with the rate-limited operation
    // ...

# Priority-Based Fairness

The rate limiting system implements priority-based fairness to ensure that high-priority
operations can continue even during high load conditions. When the system is under high load
(determined by the load factor), requests are queued and processed based on their priority.

Priority values typically range from 1 (lowest) to 10 (highest). Users with higher priority
will receive preferential treatment during contention periods.

Example of setting different priority levels:

    // Set policies for users with different priorities
    limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
        UserID:   "critical-user",
        Priority: 10, // Highest priority
        // Other policy settings...
    })

    limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
        UserID:   "standard-user",
        Priority: 5, // Medium priority
        // Other policy settings...
    })

    limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
        UserID:   "background-job",
        Priority: 1, // Lowest priority
        // Other policy settings...
    })

# Dynamic Adjustment

The system can dynamically adjust rate limits based on the current system load. This is
controlled by the load factor, which represents the current system capacity:

- 1.0: Normal operation (100% capacity)
- <1.0: Reduced capacity (e.g., 0.5 = 50% capacity)
- >1.0: Increased capacity (e.g., 1.5 = 150% capacity)

You can manually set the load factor to simulate different load conditions:

    // Simulate high load (reduce capacity to 30%)
    limiter.SetLoadFactor(0.3)

    // Return to normal operation
    limiter.SetLoadFactor(1.0)

    // Enable/disable dynamic adjustment
    limiter.EnableDynamicAdjustment(true)

# Token Usage Tracking

The system tracks token usage per user to prevent abuse. Each user has a maximum number of
tokens they can consume within a reset interval. This provides a hard cap on resource usage
over time, independent of the rate limiting.

You can check and reset user usage:

    // Get current usage for a user
    usage := limiter.GetUserUsage("user123")

    // Reset usage for a specific user
    limiter.ResetUserUsage("user123")

    // Reset usage for all users
    limiter.ResetAllUserUsage()

# Integration with Template Management

The rate limiting system is designed to be used with the template management system to
control the rate of template executions. It can be integrated as follows:

    // Create a template executor with rate limiting
    executor := execution.NewTemplateExecutor(
        repository,
        renderer,
        limiter, // Pass the rate limiter
    )

    // Execute a template with rate limiting
    result, err := executor.ExecuteTemplate(ctx, template, data, userID)
*/
