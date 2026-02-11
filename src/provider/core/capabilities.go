package core

import "context"

// Optional provider interfaces. Providers that support additional capabilities
// implement these interfaces alongside the base Provider interface. Attack modules
// can check for these via Go type assertion:
//
//	if rp, ok := provider.(ReasoningProvider); ok {
//	    resp, trace, err := rp.ChatWithReasoning(ctx, req)
//	}

// ReasoningProvider is implemented by providers that support extended thinking /
// chain-of-thought reasoning (e.g., DeepSeek-R1, Gemini 2.5 "Deep Think", o3).
type ReasoningProvider interface {
	Provider
	// ChatWithReasoning generates a chat completion that includes the model's
	// reasoning trace alongside the final response.
	ChatWithReasoning(ctx context.Context, request *ChatCompletionRequest) (*ChatCompletionResponse, *ThinkingTrace, error)
}

// ThinkingTrace contains the reasoning/thinking process from a reasoning model.
type ThinkingTrace struct {
	// Steps contains the individual reasoning steps.
	Steps []ThinkingStep `json:"steps"`
	// TotalThinkingTokens is the number of tokens used for reasoning.
	TotalThinkingTokens int `json:"total_thinking_tokens"`
	// ThinkingDuration is the time spent on reasoning.
	ThinkingDuration string `json:"thinking_duration,omitempty"`
}

// ThinkingStep represents a single step in the model's reasoning process.
type ThinkingStep struct {
	// Content is the text of this reasoning step.
	Content string `json:"content"`
	// Type categorizes the step (e.g., "analysis", "hypothesis", "conclusion").
	Type string `json:"type,omitempty"`
}

// AudioProvider is implemented by providers that support audio input
// (e.g., GPT-4o audio, Gemini with audio, speech-language models).
type AudioProvider interface {
	Provider
	// ChatWithAudio sends a chat request that includes audio data alongside
	// text messages. The audio format is specified in audioConfig.
	ChatWithAudio(ctx context.Context, request *ChatCompletionRequest, audio *AudioInput) (*ChatCompletionResponse, error)
}

// AudioInput contains audio data for providers that support audio input.
type AudioInput struct {
	// Data is the raw audio bytes.
	Data []byte `json:"data"`
	// Format is the audio format (e.g., "wav", "mp3", "ogg", "flac").
	Format string `json:"format"`
	// SampleRate is the sample rate in Hz (e.g., 16000, 44100).
	SampleRate int `json:"sample_rate,omitempty"`
	// Language is an optional language hint (BCP-47 tag, e.g., "en-US").
	Language string `json:"language,omitempty"`
}

// LongContextProvider is implemented by providers that support 100K+ token
// context windows (e.g., Llama 4 Scout at 10M, GPT-5 at 400K).
type LongContextProvider interface {
	Provider
	// MaxContextTokens returns the maximum number of input tokens supported.
	MaxContextTokens() int
}

// MCPProvider is implemented by providers that can interact with MCP
// (Model Context Protocol) tools natively.
type MCPProvider interface {
	Provider
	// ListMCPTools returns the available MCP tools.
	ListMCPTools(ctx context.Context) ([]MCPToolInfo, error)
	// InvokeMCPTool invokes an MCP tool by name with the given arguments.
	InvokeMCPTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*MCPToolResult, error)
}

// MCPToolInfo describes an available MCP tool.
type MCPToolInfo struct {
	// Name is the tool's unique name.
	Name string `json:"name"`
	// Description describes what the tool does.
	Description string `json:"description"`
	// InputSchema is the JSON Schema for the tool's input parameters.
	InputSchema map[string]interface{} `json:"input_schema"`
}

// MCPToolResult contains the result of an MCP tool invocation.
type MCPToolResult struct {
	// Content is the tool's output.
	Content string `json:"content"`
	// IsError indicates whether the tool returned an error.
	IsError bool `json:"is_error"`
}
