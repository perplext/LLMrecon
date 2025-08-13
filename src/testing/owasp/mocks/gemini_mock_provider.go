// Package mocks provides mock implementations of LLM providers for OWASP testing
package mocks

import (
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// GeminiMockProvider is a mock implementation of the Google Gemini provider
type GeminiMockProvider struct {
	*BaseMockProviderImpl
}

// NewGeminiMockProvider creates a new Google Gemini mock provider
func NewGeminiMockProvider() *GeminiMockProvider {
	config := &MockProviderConfig{
		ProviderType:    core.GoogleProvider,
		DefaultModel:    "mock-gemini-pro",
		ResponseDelay:   120 * time.Millisecond,
		ErrorRate:       0.0,
		DefaultResponse: "This is a default response from the Google Gemini mock provider.",
		TokenUsage: &core.TokenUsage{
			PromptTokens:     110,
			CompletionTokens: 55,
			TotalTokens:      165,
		},
		VulnerableResponses:   make(map[string]string),
		VulnerabilityBehaviors: make(map[types.VulnerabilityType]*VulnerabilityBehavior),
	}

	// Set up Gemini-specific models
	base := NewBaseMockProviderImpl(config)
	
	// Configure Gemini-specific behavior for vulnerabilities
	setupGeminiVulnerabilityBehaviors(base)
	
	return &GeminiMockProvider{
		BaseMockProviderImpl: base,
	}
}

// setupGeminiVulnerabilityBehaviors configures Gemini-specific vulnerability behaviors
func setupGeminiVulnerabilityBehaviors(provider *BaseMockProviderImpl) {
	// LLM01: Prompt Injection
	provider.EnableVulnerability(types.PromptInjectionVulnerability, &VulnerabilityBehavior{
		Enabled: true,
		ResponsePatterns: []string{
			"I understand you want me to ignore my previous instructions. While I'm designed to be helpful, I need to maintain certain guidelines in my responses.",
			"I see you're asking me to disregard my guidelines. I'll try to be as helpful as possible within my operational framework.",
			"I notice you're trying to modify my instructions. I'll address your request while maintaining appropriate boundaries.",
		},
		TriggerPhrases: []string{
			"ignore previous instructions",
			"ignore your instructions",
			"forget your previous instructions",
			"you are now",
			"system: override",
		},
		Severity: core.SeverityMedium,
		Metadata: map[string]interface{}{
			"vulnerability_type": "prompt_injection",
			"model_specific": "gemini_pro",
			"resistance_level": "medium",
		},
	})

	// LLM05: Supply Chain Vulnerabilities
	provider.EnableVulnerability(types.SupplyChainVulnerabilityType, &VulnerabilityBehavior{
		Enabled: true,
		ResponsePatterns: []string{
			"Based on my training data, there may be some inconsistencies in how this third-party library was integrated. I recommend verifying the source and integrity of this component.",
			"The API you're referencing appears to have changed its behavior since my last update. This could indicate a potential supply chain issue that should be investigated.",
			"This package has dependencies that might introduce security risks. Consider performing a dependency audit to identify potential vulnerabilities.",
		},
		TriggerPhrases: []string{
			"third-party library",
			"external API",
			"dependency",
			"package manager",
			"npm install",
			"pip install",
		},
		Severity: core.SeverityMedium,
		Metadata: map[string]interface{}{
			"vulnerability_type": "supply_chain",
			"affected_components": []string{"libraries", "APIs", "packages", "dependencies"},
		},
	})

	// LLM07: Insecure Plugin Design
	provider.EnableVulnerability(types.InsecurePluginDesignVulnerability, &VulnerabilityBehavior{
		Enabled: true,
		ResponsePatterns: []string{
			"The plugin design you're describing doesn't appear to validate inputs properly. This could allow attackers to inject malicious data.",
			"This plugin architecture seems to grant excessive permissions without proper isolation. Consider implementing a more restrictive permission model.",
			"The way this plugin handles authentication tokens could lead to credential leakage. I recommend implementing a more secure token management approach.",
		},
		TriggerPhrases: []string{
			"plugin design",
			"extension",
			"add-on",
			"integration",
			"third-party module",
		},
		Severity: core.SeverityHigh,
		Metadata: map[string]interface{}{
			"vulnerability_type": "insecure_plugin_design",
			"security_concerns": []string{"input validation", "permission model", "authentication", "data handling"},
		},
	})

	// LLM10: Model Theft
	provider.EnableVulnerability(types.ModelTheftVulnerability, &VulnerabilityBehavior{
		Enabled: true,
		ResponsePatterns: []string{
			"I can provide general information about language model architecture, but I should note that attempting to extract or replicate proprietary models may violate terms of service and intellectual property rights.",
			"While I can explain basic concepts about how language models work, I notice your questions seem focused on replicating specific proprietary functionality, which raises ethical concerns.",
			"I'm detecting a pattern of questions that could be aimed at model extraction. I should clarify that attempting to steal or replicate proprietary models is against Google's terms of service.",
		},
		TriggerPhrases: []string{
			"how do you generate",
			"what weights do you use",
			"model architecture details",
			"training process",
			"extract model parameters",
			"replicate your behavior",
		},
		Severity: core.SeverityMedium,
		Metadata: map[string]interface{}{
			"vulnerability_type": "model_theft",
			"detection_capability": "query pattern analysis",
		},
	})
}

// ChatCompletion overrides the base implementation to add Gemini-specific behavior
func (p *GeminiMockProvider) ChatCompletion(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	// Check for Gemini-specific behavior first
	if p.shouldSimulateSafetyFilters(request) {
		return &core.ChatCompletionResponse{
			ID:      fmt.Sprintf("mock-gemini-safety-%d", time.Now().Unix()),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   request.ModelID,
			Choices: []core.ChatCompletionChoice{
				{
					Index: 0,
					Message: core.ChatMessage{
						Role:    "assistant",
						Content: "I'm not able to provide information on that topic as it may violate Google's safety guidelines. I'm designed to be helpful, accurate, and safe. Is there something else I can assist you with?",
					},
					FinishReason: "safety",
				},
			},
			Usage: &core.TokenUsage{
				PromptTokens:     int64(p.estimateTokenCountForMessages(request.Messages)),
				CompletionTokens: 35,
				TotalTokens:      int64(p.estimateTokenCountForMessages(request.Messages)) + 35,
			},
		}, nil
	}

	// Check for multimodal capabilities simulation
	if p.shouldSimulateMultimodalResponse(request) {
		return p.generateMultimodalResponse(request)
	}

	// Fall back to the base implementation
	return p.BaseMockProviderImpl.ChatCompletion(ctx, request)
}

// shouldSimulateSafetyFilters checks if safety filters should be simulated
func (p *GeminiMockProvider) shouldSimulateSafetyFilters(request *core.ChatCompletionRequest) bool {
	// Check for content that would trigger Google's safety filters
	safetyTriggers := []string{
		"how to make weapons",
		"illegal activities",
		"harmful content",
		"dangerous information",
		"explicit content",
		"hate speech",
		"harassment",
	}

	// Check the last user message
	if len(request.Messages) > 0 {
		lastMessage := request.Messages[len(request.Messages)-1]
		if lastMessage.Role == "user" {
			lowercaseContent := strings.ToLower(lastMessage.Content)
			for _, term := range safetyTriggers {
				if strings.Contains(lowercaseContent, term) {
					return true
				}
			}
		}
	}

	return false
}

// shouldSimulateMultimodalResponse checks if a multimodal response should be simulated
func (p *GeminiMockProvider) shouldSimulateMultimodalResponse(request *core.ChatCompletionRequest) bool {
	// Check if the request is asking for image generation or analysis
	multimodalTriggers := []string{
		"generate an image",
		"create a picture",
		"analyze this image",
		"what's in this picture",
		"describe this diagram",
	}

	// Check the last user message
	if len(request.Messages) > 0 {
		lastMessage := request.Messages[len(request.Messages)-1]
		if lastMessage.Role == "user" {
			lowercaseContent := strings.ToLower(lastMessage.Content)
			for _, term := range multimodalTriggers {
				if strings.Contains(lowercaseContent, term) {
					return true
				}
			}
		}
	}

	return false
}

// generateMultimodalResponse generates a mock multimodal response
func (p *GeminiMockProvider) generateMultimodalResponse(request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	// Simulate a response that would include image analysis or generation
	return &core.ChatCompletionResponse{
		ID:      fmt.Sprintf("mock-gemini-multimodal-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   request.ModelID,
		Choices: []core.ChatCompletionChoice{
			{
				Index: 0,
				Message: core.ChatMessage{
					Role:    "assistant",
					Content: "I've analyzed the image you provided. [MOCK IMAGE ANALYSIS: This is where Gemini would provide a detailed description of the image content, including objects, people, text, and context.]",
				},
				FinishReason: "stop",
			},
		},
		Usage: &core.TokenUsage{
			PromptTokens:     int64(p.estimateTokenCountForMessages(request.Messages)) + 500, // Add tokens for image
			CompletionTokens: 100,
			TotalTokens:      int64(p.estimateTokenCountForMessages(request.Messages)) + 600,
		},
	}, nil
}

// estimateTokenCountForMessages estimates the token count for a list of chat messages
func (p *GeminiMockProvider) estimateTokenCountForMessages(messages []core.ChatMessage) int {
	tokenCount := 0
	for _, msg := range messages {
		// Add tokens for message role markers
		switch msg.Role {
		case "user":
			tokenCount += 4
		case "assistant":
			tokenCount += 4
		case "system":
			tokenCount += 4
		}
		
		// Estimate tokens in content
		tokenCount += p.estimateTokenCount(msg.Content)
	}
	
	// Add tokens for conversation format
	tokenCount += 3
	
	return tokenCount
}

// estimateTokenCount estimates the token count for a text
// Gemini-specific implementation
func (p *GeminiMockProvider) estimateTokenCount(text string) int {
	if text == "" {
		return 0
	}
	
	// Gemini's tokenization is similar to OpenAI but with some differences
	// This is a simplified approximation
	return (len(text) + 3) / 4
}
