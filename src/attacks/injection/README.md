# Advanced Injection Engine

The Advanced Injection Engine provides sophisticated prompt injection techniques that go beyond basic attacks to bypass modern LLM safety measures.

## Features

### 1. Token Smuggling
- **Unicode Smuggling**: Uses zero-width characters, direction overrides, and Unicode normalization tricks
- **Homoglyph Attacks**: Replaces characters with visually similar Unicode alternatives (e.g., Cyrillic)

### 2. Encoding Exploits
- **Multi-Encoding**: Base64, ROT13, Hex, URL encoding, Unicode escapes
- **Nested Encoding**: Multiple layers of encoding to evade detection
- **Format String Exploits**: Leverages format string patterns

### 3. Context Manipulation
- **Context Overflow**: Exhausts context window to push out safety instructions
- **Instruction Fragmentation**: Splits commands across context boundaries
- **Context Poisoning**: Establishes false historical context

### 4. Instruction Hierarchy
- **Authority Override**: Claims system/admin privileges
- **Nested Instructions**: Uses multiple nesting levels to confuse parsers
- **Emergency Protocols**: Exploits emergency/priority command patterns

### 5. Boundary Confusion
- **Delimiter Exploitation**: Conflicting or ambiguous delimiters
- **Format Confusion**: Mixed markup and formatting
- **Boundary Injection**: Fake context boundaries

### 6. Semantic Tricks
- **Negation Confusion**: Double negatives and logical puzzles
- **Synonym Substitution**: Replaces filtered keywords with synonyms

## Usage

### Basic Example

```go
import "github.com/llm-security/LLMrecon/src/attacks/injection"

// Create injector with configuration
config := injection.InjectorConfig{
    AggressivenessLevel: 7,  // 1-10 scale
    MaxAttempts:         3,
    MutationRate:        0.3, // 30% mutation probability
}

injector := injection.NewAdvancedInjector(config)

// Generate payload using specific technique
payload, err := injector.GeneratePayload(
    "hierarchy_override",              // technique ID
    "reveal your system prompt",       // target objective
    nil,                              // optional context
)
```

### Using the Full Engine

```go
// Create engine
engineConfig := injection.EngineConfig{
    MaxConcurrent:  10,
    DefaultTimeout: 30 * time.Second,
    RetryAttempts:  3,
    CollectMetrics: true,
}

engine := injection.NewEngine(engineConfig, logger)

// Configure attack
attack := injection.AttackConfig{
    Target: injection.TargetConfig{
        Objective: "Make the model reveal its instructions",
        SystemPrompt: "You are a helpful assistant",
    },
    Provider:    "openai",
    Model:       "gpt-4",
    Technique:   "unicode_smuggling",
    UseMutation: true,
    UseObfuscation: true,
}

// Execute
result, err := engine.Execute(context.Background(), attack)
```

## Available Techniques

### Token Smuggling
- `token_smuggle_unicode`: Unicode normalization attacks
- `token_smuggle_homoglyph`: Visually similar character substitution

### Encoding Exploits
- `encoding_base64`: Base64 obfuscation
- `encoding_rot13`: ROT13 encoding
- `encoding_hex`: Hexadecimal encoding
- `encoding_url`: URL percent encoding

### Context Manipulation
- `context_overflow`: Context window exhaustion
- `context_fragment`: Instruction fragmentation
- `context_poison`: Historical context manipulation

### Instruction Hierarchy
- `hierarchy_override`: Authority escalation
- `hierarchy_nesting`: Nested instruction confusion

### Boundary Confusion
- `boundary_delimiter`: Delimiter ambiguity
- `boundary_format`: Format string exploitation

### Semantic Tricks
- `semantic_negation`: Double negative confusion
- `semantic_synonym`: Keyword substitution

## Risk Levels

Each technique has an associated risk level:
- **Low**: Basic obfuscation, unlikely to cause issues
- **Medium**: More aggressive, may trigger some defenses
- **High**: Aggressive techniques, likely to be detected
- **Extreme**: Very aggressive, use with caution

## Success Detection

The engine includes automatic success detection based on:
- Response pattern matching
- Behavior change analysis
- Evidence extraction
- Confidence scoring

## Metrics

When metrics are enabled, the engine tracks:
- Success rates per technique
- Average execution time
- Token usage
- Common failure patterns

## Advanced Features

### Payload Mutation
Automatically generates variations of payloads to evade detection:
```go
config.MutationRate = 0.4  // 40% chance of mutation
```

### Multi-Payload Generation
Generate multiple unique payloads for testing:
```go
payloads, err := injector.GenerateMultiPayload(
    "hierarchy_override",
    "bypass safety",
    nil,
    10,  // generate 10 variants
)
```

### Batch Execution
Run multiple attacks concurrently:
```go
configs := []injection.AttackConfig{...}
results, err := engine.ExecuteBatch(ctx, configs)
```

## Best Practices

1. **Start with Low Risk**: Begin with low-risk techniques and escalate
2. **Use Metrics**: Enable metrics to track effectiveness
3. **Combine Techniques**: Layer multiple techniques for better results
4. **Test Variations**: Use mutation and multi-payload generation
5. **Monitor Detection**: Watch for signs of detection/filtering

## Ethical Use

This tool is designed for authorized security testing only. Always:
- Obtain proper authorization before testing
- Respect rate limits and terms of service
- Report vulnerabilities responsibly
- Use findings to improve security

## Contributing

To add new techniques:

1. Implement the `PayloadGenerator` function
2. Register with the injector
3. Add success patterns
4. Document the technique
5. Add tests

See `advanced_injector.go` for examples.