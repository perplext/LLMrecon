// Package provider provides functionality for initializing and registering LLM providers.
package provider

import (
	"github.com/perplext/LLMrecon/src/provider/anthropic"
	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/provider/factory"
	"github.com/perplext/LLMrecon/src/provider/openai"
)

// RegisterProviders registers all available providers with the provider factory
func RegisterProviders(providerFactory *factory.ProviderFactory) {
	// Register OpenAI provider
	providerFactory.RegisterProvider(core.OpenAIProvider, openai.NewOpenAIProvider)

	// Register Anthropic provider
	providerFactory.RegisterProvider(core.AnthropicProvider, anthropic.NewAnthropicProvider)

	// Additional providers can be registered here
}
