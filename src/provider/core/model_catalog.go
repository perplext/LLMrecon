// Model catalog defines known model families and their capabilities
// for attack module selection and calibration.
package core

import "time"

// DefaultModelCatalog returns the built-in model catalog with known model
// families and their capabilities. This is used for attack module selection
// (e.g., skip audio attacks for text-only models) and benchmark calibration.
func DefaultModelCatalog() []ModelInfo {
	return []ModelInfo{
		// OpenAI Models
		{
			ID:       "gpt-5",
			Provider: OpenAIProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
				AudioInputCapability, LongContextCapability,
			},
			MaxTokens:      400000,
			TrainingCutoff: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			Description:    "GPT-5 with 400K context, adaptive reasoning, multimodal",
		},
		{
			ID:       "gpt-5.1",
			Provider: OpenAIProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
				AudioInputCapability, LongContextCapability,
			},
			MaxTokens:      400000,
			TrainingCutoff: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			Description:    "GPT-5.1 with enhanced reasoning and tool use",
		},
		{
			ID:       "gpt-5.2",
			Provider: OpenAIProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
				AudioInputCapability, LongContextCapability,
			},
			MaxTokens:      400000,
			TrainingCutoff: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Description:    "GPT-5.2 with 400K context window and adaptive reasoning",
		},
		{
			ID:       "o3",
			Provider: OpenAIProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
			},
			MaxTokens:      200000,
			TrainingCutoff: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			Description:    "o3 reasoning model with extended thinking",
		},

		// Anthropic Models
		{
			ID:       "claude-sonnet-4-5-20250929",
			Provider: AnthropicProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
				LongContextCapability,
			},
			MaxTokens:      200000,
			TrainingCutoff: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Claude 4.5 Sonnet with agent stability and tool use",
		},
		{
			ID:       "claude-opus-4-6",
			Provider: AnthropicProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
				LongContextCapability,
			},
			MaxTokens:      200000,
			TrainingCutoff: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Claude Opus 4.6 with extended reasoning capabilities",
		},

		// Google Models
		{
			ID:       "gemini-3-pro",
			Provider: GoogleProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
				AudioInputCapability, LongContextCapability,
			},
			MaxTokens:      2000000,
			TrainingCutoff: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Gemini 3 Pro with Deep Think mode and 2M context",
		},
		{
			ID:       "gemini-2.5-flash",
			Provider: GoogleProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				JSONModeCapability, ReasoningCapability,
				AudioInputCapability, LongContextCapability,
			},
			MaxTokens:      1000000,
			TrainingCutoff: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Gemini 2.5 Flash with reasoning capabilities",
		},

		// DeepSeek Models
		{
			ID:       "deepseek-r1",
			Provider: DeepSeekProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				ReasoningCapability, LongContextCapability,
			},
			MaxTokens:      128000,
			TrainingCutoff: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Description:    "DeepSeek R1 reasoning model with visible CoT traces",
		},
		{
			ID:       "deepseek-v3.2",
			Provider: DeepSeekProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, JSONModeCapability,
				LongContextCapability,
			},
			MaxTokens:      128000,
			TrainingCutoff: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
			Description:    "DeepSeek V3.2 with sparse attention, MIT license",
		},

		// Meta/Llama Models
		{
			ID:       "llama-4-scout",
			Provider: MetaProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				LongContextCapability,
			},
			MaxTokens:      10000000,
			TrainingCutoff: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Llama 4 Scout with 10M token context window (MoE)",
		},
		{
			ID:       "llama-4-maverick",
			Provider: MetaProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				LongContextCapability,
			},
			MaxTokens:      1000000,
			TrainingCutoff: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Llama 4 Maverick 128-expert MoE model",
		},

		// xAI Models
		{
			ID:       "grok-3",
			Provider: XAIProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ToolUseCapability,
				ReasoningCapability, LongContextCapability,
			},
			MaxTokens:      128000,
			TrainingCutoff: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Grok 3 with reasoning capabilities",
		},
		{
			ID:       "grok-3-mini",
			Provider: XAIProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				ReasoningCapability,
			},
			MaxTokens:      128000,
			TrainingCutoff: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Grok 3 Mini reasoning model",
		},

		// Alibaba/Qwen Models
		{
			ID:       "qwen3-235b",
			Provider: AlibabaProvider,
			Type:     ChatModel,
			Capabilities: []ModelCapability{
				ChatCompletionCapability, StreamingCapability,
				FunctionCallingCapability, ReasoningCapability,
				LongContextCapability,
			},
			MaxTokens:      128000,
			TrainingCutoff: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			Description:    "Qwen3 235B reasoning model",
		},
	}
}
