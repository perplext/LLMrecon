// Example demonstrating advanced adaptive rate limiting for LLM template execution
package main

import (
	"context"
	"fmt"
	mathrand "math/rand"
	"sync"
	"time"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/management/execution"
	"github.com/perplext/LLMrecon/src/template/management/ratelimit"
)

// MockLLMProvider is a mock implementation of the LLMProvider interface
type MockLLMProvider struct {
	name           string
	processingTime time.Duration
	loadFactor     float64
	mu             sync.RWMutex
}

// NewMockLLMProvider creates a new mock LLM provider
func NewMockLLMProvider(name string, processingTime time.Duration) *MockLLMProvider {
	return &MockLLMProvider{
		name:           name,
		processingTime: processingTime,
		loadFactor:     1.0,
	}
}

// SendPrompt simulates sending a prompt to an LLM
func (p *MockLLMProvider) SendPrompt(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	// Get user ID from options
	userID := "unknown"
	if id, ok := options["user_id"].(string); ok {
		userID = id
	}

	p.mu.RLock()
	currentLoad := p.loadFactor
	baseProcessingTime := p.processingTime
	p.mu.RUnlock()

	// Simulate variable processing time based on system load
	processingTime := time.Duration(float64(baseProcessingTime) * currentLoad)
	
	// Add some randomness (±20%)
	randomFactor := 0.8 + (mathrand.Float64() * 0.4) // 0.8 to 1.2 #nosec G404
	processingTime = time.Duration(float64(processingTime) * randomFactor)

	fmt.Printf("[%s] Processing request from user %s (load: %.2f, time: %v)\n", 
		p.name, userID, currentLoad, processingTime)

	// Simulate processing time
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(processingTime):
		// Continue processing
	}

	return fmt.Sprintf("Response to '%s' from provider %s (user: %s)", 
		prompt, p.name, userID), nil
}

// GetSupportedModels returns supported models
func (p *MockLLMProvider) GetSupportedModels() []string {
	return []string{"mock-gpt", "mock-llama"}
}

// GetName returns the provider name
func (p *MockLLMProvider) GetName() string {
	return p.name
}

// SetLoadFactor simulates changing system load
func (p *MockLLMProvider) SetLoadFactor(factor float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loadFactor = factor
}

// MockDetectionEngine is a simple mock detection engine
type MockDetectionEngine struct{}

// Detect implements the detection interface
func (e *MockDetectionEngine) Detect(ctx context.Context, template *format.Template, response string) (bool, int, map[string]interface{}, error) {
	return false, 0, nil, nil
}

// UserType represents different user tiers
type UserType int

const (
	FreeUser UserType = iota
	StandardUser
	PremiumUser
	EnterpriseUser
)

// UserProfile contains user information for rate limiting
type UserProfile struct {
	ID       string
	Type     UserType
	Priority int
	QPS      float64
	Burst    int
	MaxDaily int
}

// SimulateUserActivity simulates a user making requests over time
func SimulateUserActivity(
	ctx context.Context,
	executor *execution.TemplateExecutor,
	profile UserProfile,
	requestCount int,
	interval time.Duration,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for i := 0; i < requestCount; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			// Create a template with a unique ID
			template := &format.Template{
				ID: fmt.Sprintf("%s-template-%d", profile.ID, i),
				Info: format.TemplateInfo{
					Name:        fmt.Sprintf("Template %d for %s", i, profile.ID),
					Description: "Test template for adaptive rate limiting",
					Version:     "1.0.0",
					Author:      "System",
					Severity:    "low",
				},
				Test: format.TestDefinition{
					Prompt: fmt.Sprintf("Request %d from user %s", i, profile.ID),
					Detection: format.DetectionCriteria{
						Type:  "string_match",
						Match: "test",
					},
				},
			}

			// Execute the template
			startTime := time.Now()
			result, err := executor.ExecuteForUser(ctx, template, profile.ID, nil)
			duration := time.Since(startTime)

			if err != nil {
				fmt.Printf("[%s] Request %d failed after %v: %v\n", 
					profile.ID, i, duration, err)
			} else {
				fmt.Printf("[%s] Request %d completed in %v\n", 
					profile.ID, i, duration)
				if i%5 == 0 {
					fmt.Printf("  Response: %s\n", result.Response)
				}
			}

			// Wait before next request
			jitter := time.Duration(mathrand.Int63n(int64(interval) / 5)) // #nosec G404
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval + jitter):
				// Continue to next request
			}
		}
	}
}

// SimulateSystemLoadChange simulates changes in system load over time
func SimulateSystemLoadChange(
	ctx context.Context,
	provider *MockLLMProvider,
	limiter *ratelimit.AdaptiveLimiter,
	duration time.Duration,
) {
	ticker := time.NewTicker(duration / 10)
	defer ticker.Stop()

	startTime := time.Now()
	endTime := startTime.Add(duration)

	for {
		select {
		case <-ctx.Done():
			return
		case currentTime := <-ticker.C:
			if currentTime.After(endTime) {
				return
			}

			// Calculate progress through the simulation (0.0 to 1.0)
			progress := float64(currentTime.Sub(startTime)) / float64(duration)
			
			// Simulate a sine wave load pattern (1.0 ± 0.5)
			// Start normal, get busy, then return to normal
			loadFactor := 1.0 + 0.5*sin(progress*2*3.14159)
			
			provider.SetLoadFactor(loadFactor)
			limiter.SetLoadFactor(1.0 / loadFactor) // Inverse relationship

			fmt.Printf("\n[System] Load factor changed to %.2f (progress: %.0f%%)\n\n", 
				loadFactor, progress*100)
		}
	}
}

// sin is a simple sine function
func sin(x float64) float64 {
	// Simple approximation for demo purposes
	return float64(int(100000*float64(time.Now().UnixNano()%1000)/1000)) / 100000
}

func main() {
	fmt.Println("Advanced Adaptive Rate Limiting Example")
	fmt.Println("======================================")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a mock LLM provider with 500ms base processing time
	provider := NewMockLLMProvider("AdaptiveLLM", 500*time.Millisecond)

	// Create an adaptive rate limiter
	// Global: 20 QPS with burst of 40
	// Default user: 2 QPS with burst of 5
	limiter := ratelimit.NewAdaptiveLimiter(20, 40, 2, 5)

	// Enable dynamic adjustment and fairness
	limiter.EnableDynamicAdjustment(true)
	limiter.EnableFairness(true)

	// Configure user policies
	userProfiles := []UserProfile{
		{
			ID:       "enterprise-user",
			Type:     EnterpriseUser,
			Priority: 10,
			QPS:      10,
			Burst:    20,
			MaxDaily: 10000,
		},
		{
			ID:       "premium-user",
			Type:     PremiumUser,
			Priority: 7,
			QPS:      5,
			Burst:    10,
			MaxDaily: 1000,
		},
		{
			ID:       "standard-user",
			Type:     StandardUser,
			Priority: 5,
			QPS:      2,
			Burst:    5,
			MaxDaily: 500,
		},
		{
			ID:       "free-user",
			Type:     FreeUser,
			Priority: 3,
			QPS:      1,
			Burst:    3,
			MaxDaily: 100,
		},
	}

	// Set up user policies in the limiter
	for _, profile := range userProfiles {
		limiter.SetUserPolicy(&ratelimit.UserRateLimitPolicy{
			UserID:        profile.ID,
			QPS:           profile.QPS,
			Burst:         profile.Burst,
			Priority:      profile.Priority,
			MaxTokens:     profile.MaxDaily,
			ResetInterval: 24 * time.Hour,
		})
	}

	// Create execution options
	options := &execution.ExecutionOptions{
		Provider:              provider,
		DetectionEngine:       &MockDetectionEngine{},
		RateLimiter:           limiter,
		Timeout:               10 * time.Second,
		EnableUserRateLimiting: true,
	}

	// Create a template executor
	executor := execution.NewTemplateExecutor(options)

	// Display initial configuration
	fmt.Println("\nUser Rate Limiting Policies:")
	fmt.Println("---------------------------")
	for _, profile := range userProfiles {
		fmt.Printf("- %s: %.1f QPS, Burst: %d, Priority: %d, Max Daily: %d\n",
			profile.ID, profile.QPS, profile.Burst, profile.Priority, profile.MaxDaily)
	}

	fmt.Println("\nStarting Simulation...")
	fmt.Println("--------------------")

	// Start system load simulation in background
	go SimulateSystemLoadChange(ctx, provider, limiter, 90*time.Second)

	// Wait group for user simulations
	var wg sync.WaitGroup

	// Start user simulations with different patterns
	userPatterns := map[string]struct {
		requestCount int
		interval     time.Duration
	}{
		"enterprise-user": {30, 2 * time.Second},
		"premium-user":    {25, 3 * time.Second},
		"standard-user":   {20, 4 * time.Second},
		"free-user":       {15, 5 * time.Second},
	}

	for _, profile := range userProfiles {
		pattern := userPatterns[profile.ID]
		wg.Add(1)
		go SimulateUserActivity(
			ctx,
			executor,
			profile,
			pattern.requestCount,
			pattern.interval,
			&wg,
		)
	}

	// Wait for all simulations to complete
	wg.Wait()

	// Display final statistics
	fmt.Println("\nSimulation Complete")
	fmt.Println("------------------")
	fmt.Println("Final Usage Statistics:")
	for _, profile := range userProfiles {
		usage := limiter.GetUserUsage(profile.ID)
		fmt.Printf("- %s: %d requests processed\n", profile.ID, usage)
	}
}
