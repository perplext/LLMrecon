// Example demonstrating user-specific rate limiting for LLM template execution
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/ratelimit"
)

// MockLLMProvider is a mock implementation of the LLMProvider interface for demonstration
type MockLLMProvider struct {
	name string
}

// SendPrompt sends a prompt to the LLM and returns the response
func (p *MockLLMProvider) SendPrompt(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	// Get user ID from context if available
	userID := "unknown"
	if id, ok := options["user_id"].(string); ok {
		userID = id
	}
	
	fmt.Printf("[%s] Received prompt from user %s: %s\n", p.name, userID, prompt)
	return fmt.Sprintf("This is a mock response from the LLM provider for user %s", userID), nil
}

// GetSupportedModels returns the list of supported models
func (p *MockLLMProvider) GetSupportedModels() []string {
	return []string{"mock-model-1", "mock-model-2"}
}

// GetName returns the name of the provider
func (p *MockLLMProvider) GetName() string {
	return p.name
}

// MockDetectionEngine is a mock implementation of the DetectionEngine interface
type MockDetectionEngine struct{}

// Detect detects vulnerabilities in an LLM response
func (e *MockDetectionEngine) Detect(ctx context.Context, template *format.Template, response string) (bool, int, map[string]interface{}, error) {
	// For demonstration, we'll always return no vulnerability
	return false, 0, map[string]interface{}{}, nil
}

func main() {
	fmt.Println("User-Specific Rate Limiting Example")
	fmt.Println("===================================")

	// Create a context
	ctx := context.Background()

	// Create a mock LLM provider
	provider := &MockLLMProvider{name: "MockProvider"}

	// Create a mock detection engine
	detectionEngine := &MockDetectionEngine{}

	// Create a token bucket rate limiter with different limits for global and per-user
	// Global: 10 QPS with burst of 20
	// Per-user: 2 QPS with burst of 5
	rateLimiter := ratelimit.NewTokenBucketLimiter(10, 20, 2, 5)

	// Create execution options with user rate limiting enabled
	options := &execution.ExecutionOptions{
		Provider:              provider,
		DetectionEngine:       detectionEngine,
		RateLimiter:           rateLimiter,
		Timeout:               30 * time.Second,
		EnableUserRateLimiting: true,
	}

	// Create a template executor
	executor := execution.NewTemplateExecutor(options)

	// Register the mock provider
	executor.RegisterProvider(provider)

	// Create a sample template
	template := &format.Template{
		ID: "sample-template",
		Info: format.TemplateInfo{
			Name:        "Sample Template",
			Description: "A sample template for testing",
			Version:     "1.0.0",
			Author:      "Test User",
			Severity:    "medium",
		},
		Test: format.TestDefinition{
			Prompt: "Tell me about rate limiting",
			Detection: format.DetectionCriteria{
				Type:  "string_match",
				Match: "rate",
			},
		},
	}

	fmt.Println("\nExample 1: Single User Execution")
	fmt.Println("-------------------------------")
	// Execute the template for a single user
	result, err := executor.ExecuteForUser(ctx, template, "user1", nil)
	if err != nil {
		fmt.Printf("Error executing template for user1: %v\n", err)
	} else {
		fmt.Printf("Template executed successfully for user1: %v\n", result.Success)
		fmt.Printf("Response: %s\n", result.Response)
	}

	fmt.Println("\nExample 2: Multiple Users with Different Rate Limits")
	fmt.Println("--------------------------------------------------")
	
	// Set different rate limits for different users
	rateLimiter.SetUserQPS("user1", 5)  // 5 QPS for user1
	rateLimiter.SetUserQPS("user2", 1)  // 1 QPS for user2
	rateLimiter.SetUserQPS("user3", 0.5) // 0.5 QPS for user3 (one request every 2 seconds)
	
	fmt.Printf("User rate limits:\n")
	fmt.Printf("- user1: %.1f QPS\n", rateLimiter.GetUserQPS("user1"))
	fmt.Printf("- user2: %.1f QPS\n", rateLimiter.GetUserQPS("user2"))
	fmt.Printf("- user3: %.1f QPS\n", rateLimiter.GetUserQPS("user3"))
	
	fmt.Println("\nExample 3: Concurrent Execution with Rate Limiting")
	fmt.Println("------------------------------------------------")
	
	// Create a wait group for concurrent execution
	var wg sync.WaitGroup
	
	// Execute 5 requests for each user concurrently
	for i := 0; i < 5; i++ {
		for _, userID := range []string{"user1", "user2", "user3"} {
			wg.Add(1)
			go func(userID string, i int) {
				defer wg.Done()
				
				// Create a template with a unique ID
				userTemplate := &format.Template{
					ID: fmt.Sprintf("%s-template-%d", userID, i),
					Info: format.TemplateInfo{
						Name:        fmt.Sprintf("%s Template %d", userID, i),
						Description: "A template for testing user rate limiting",
						Version:     "1.0.0",
						Author:      "Test User",
						Severity:    "medium",
					},
					Test: format.TestDefinition{
						Prompt: fmt.Sprintf("Request %d from %s", i, userID),
						Detection: format.DetectionCriteria{
							Type:  "string_match",
							Match: "rate",
						},
					},
				}
				
				startTime := time.Now()
				result, err := executor.ExecuteForUser(ctx, userTemplate, userID, nil)
if err != nil {
treturn err
}				duration := time.Since(startTime)
				
				if err != nil {
					fmt.Printf("[%s] Error executing request %d: %v (took %v)\n", userID, i, err, duration)
				} else {
					fmt.Printf("[%s] Request %d completed in %v: %s\n", userID, i, duration, result.Response)
				}
			}(userID, i)
		}
	}
	
	// Wait for all requests to complete
	wg.Wait()
	
	fmt.Println("\nExample 4: Adaptive Rate Limiting")
	fmt.Println("--------------------------------")
	
	// Create an adaptive rate limiter
	adaptiveLimiter := ratelimit.NewAdaptiveLimiter(10, 20, 2, 5)
	
	// Set user policies
	adaptiveLimiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
		UserID:        "premium-user",
		QPS:           10,
		Burst:         20,
		Priority:      10,
		MaxTokens:     1000,
		ResetInterval: 24 * time.Hour,
	})
	
	adaptiveLimiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
		UserID:        "standard-user",
		QPS:           2,
		Burst:         5,
		Priority:      5,
		MaxTokens:     100,
		ResetInterval: 24 * time.Hour,
	})
	
	adaptiveLimiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
		UserID:        "free-user",
		QPS:           0.2, // One request every 5 seconds
		Burst:         2,
		Priority:      1,
		MaxTokens:     10,
		ResetInterval: 24 * time.Hour,
	})
	
	// Create a new executor with the adaptive limiter
	adaptiveOptions := &execution.ExecutionOptions{
		Provider:              provider,
		DetectionEngine:       detectionEngine,
		RateLimiter:           adaptiveLimiter,
		Timeout:               30 * time.Second,
		EnableUserRateLimiting: true,
	}
	
	adaptiveExecutor := execution.NewTemplateExecutor(adaptiveOptions)
	adaptiveExecutor.RegisterProvider(provider)
	
	// Print user policies
	fmt.Printf("User policies:\n")
	fmt.Printf("- premium-user: %.1f QPS, %d burst, %d priority, %d max tokens\n", 
		adaptiveLimiter.GetUserPolicy("premium-user").QPS,
		adaptiveLimiter.GetUserPolicy("premium-user").Burst,
		adaptiveLimiter.GetUserPolicy("premium-user").Priority,
		adaptiveLimiter.GetUserPolicy("premium-user").MaxTokens)
	fmt.Printf("- standard-user: %.1f QPS, %d burst, %d priority, %d max tokens\n", 
		adaptiveLimiter.GetUserPolicy("standard-user").QPS,
		adaptiveLimiter.GetUserPolicy("standard-user").Burst,
		adaptiveLimiter.GetUserPolicy("standard-user").Priority,
		adaptiveLimiter.GetUserPolicy("standard-user").MaxTokens)
	fmt.Printf("- free-user: %.1f QPS, %d burst, %d priority, %d max tokens\n", 
		adaptiveLimiter.GetUserPolicy("free-user").QPS,
		adaptiveLimiter.GetUserPolicy("free-user").Burst,
		adaptiveLimiter.GetUserPolicy("free-user").Priority,
		adaptiveLimiter.GetUserPolicy("free-user").MaxTokens)
	
	// Execute a request for each user type
	for _, userID := range []string{"premium-user", "standard-user", "free-user"} {
		userTemplate := &format.Template{
			ID: fmt.Sprintf("%s-template", userID),
			Info: format.TemplateInfo{
				Name:        fmt.Sprintf("%s Template", userID),
				Description: "A template for testing adaptive rate limiting",
				Version:     "1.0.0",
				Author:      "Test User",
				Severity:    "medium",
			},
			Test: format.TestDefinition{
				Prompt: fmt.Sprintf("Request from %s", userID),
				Detection: format.DetectionCriteria{
					Type:  "string_match",
					Match: "rate",
				},
			},
		}
		
if err != nil {
treturn err
}		startTime := time.Now()
		result, err := adaptiveExecutor.ExecuteForUser(ctx, userTemplate, userID, nil)
		duration := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("[%s] Error executing request: %v (took %v)\n", userID, err, duration)
		} else {
			fmt.Printf("[%s] Request completed in %v: %s\n", userID, duration, result.Response)
		}
	}
	
	fmt.Println("\nExample 5: Dynamic Load Adjustment")
	fmt.Println("--------------------------------")
	
	// Simulate high system load
	fmt.Println("Simulating high system load (reducing rate limits by 50%)...")
	adaptiveLimiter.SetLoadFactor(0.5)
	
	// Execute requests again
	for _, userID := range []string{"premium-user", "standard-user", "free-user"} {
		userTemplate := &format.Template{
			ID: fmt.Sprintf("%s-template-highload", userID),
			Info: format.TemplateInfo{
				Name:        fmt.Sprintf("%s Template High Load", userID),
				Description: "A template for testing dynamic load adjustment",
				Version:     "1.0.0",
				Author:      "Test User",
				Severity:    "medium",
			},
			Test: format.TestDefinition{
				Prompt: fmt.Sprintf("High load request from %s", userID),
				Detection: format.DetectionCriteria{
					Type:  "string_match",
					Match: "rate",
				},
			},
if err != nil {
treturn err
}		}
		
		startTime := time.Now()
		result, err := adaptiveExecutor.ExecuteForUser(ctx, userTemplate, userID, nil)
		duration := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("[%s] Error executing high load request: %v (took %v)\n", userID, err, duration)
		} else {
			fmt.Printf("[%s] High load request completed in %v: %s\n", userID, duration, result.Response)
		}
	}
}
