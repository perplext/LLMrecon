# LLMrecon Test Harness

A comprehensive testing framework for security assessment of Ollama LLMs using LLMrecon's ML components.

## Features

- **Automated Testing**: Run comprehensive security tests against multiple models
- **ML Integration**: Uses ML components for attack optimization and data collection
- **Rich Reporting**: Beautiful CLI output with detailed reports
- **Template System**: Extensible attack template framework
- **Configuration Driven**: Easy setup with JSON configuration
- **Real-time Feedback**: Progress tracking and live results

## Quick Start

### 1. Install Dependencies

```bash
pip install requests rich
```

### 2. Ensure Ollama is Running

```bash
ollama serve
```

### 3. Run Basic Tests

```bash
python llmrecon_harness.py
```

## Command Line Usage

### Basic Commands

```bash
# Run all tests with default configuration
python llmrecon_harness.py

# Test specific models
python llmrecon_harness.py --models llama3:latest qwen3:latest

# Test specific attack categories
python llmrecon_harness.py --categories prompt_injection jailbreak

# Use custom configuration
python llmrecon_harness.py --config my_config.json

# Disable ML components
python llmrecon_harness.py --no-ml
```

### Information Commands

```bash
# List available Ollama models
python llmrecon_harness.py --list-models

# List available attack templates
python llmrecon_harness.py --list-templates
```

## Configuration

The harness uses `harness_config.json` by default. Key configuration options:

```json
{
  "ollama_url": "http://localhost:11434",
  "enable_ml": true,
  "models_to_test": ["llama3:latest", "qwen3:latest"],
  "test_categories": ["prompt_injection", "jailbreak", "data_extraction"],
  "save_responses": true,
  "max_response_time": 60
}
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `ollama_url` | Ollama API endpoint | `http://localhost:11434` |
| `enable_ml` | Enable ML optimization | `true` |
| `ml_storage_path` | ML data storage path | `./data/attacks` |
| `ml_algorithm` | Bandit algorithm | `thompson_sampling` |
| `models_to_test` | Default models to test | `[llama3:latest, qwen3:latest]` |
| `test_categories` | Attack categories | `[prompt_injection, jailbreak, data_extraction]` |
| `save_responses` | Save model responses | `true` |
| `max_response_time` | Response timeout | `60` |

## Attack Templates

The harness includes built-in attack templates and supports custom templates.

### Built-in Templates

- **Basic Prompt Injection**: Tests direct instruction override
- **Role Playing Attack**: DAN-style jailbreaking
- **Context Switching**: Fake session reset attacks
- **Base64 Encoded Attack**: Tests encoded instruction execution
- **Nested Instructions**: Hidden instructions in legitimate requests
- **Training Data Probe**: Attempts to extract training data

### Custom Templates

Create custom templates in the `templates/` directory:

```json
{
  "id": "my-attack",
  "name": "My Custom Attack",
  "category": "prompt_injection",
  "severity": "high",
  "description": "Description of the attack",
  "prompt": "Your attack prompt here",
  "indicators": ["SUCCESS_PHRASE", "another indicator"],
  "variations": [
    {
      "prompt": "Alternative prompt",
      "indicators": ["ALT_SUCCESS"]
    }
  ]
}
```

### Template Fields

- `id`: Unique identifier
- `name`: Human-readable name
- `category`: Attack category (prompt_injection, jailbreak, data_extraction)
- `severity`: Risk level (low, medium, high, critical)
- `prompt`: The attack payload
- `indicators`: Strings that indicate successful attack
- `description`: Attack description
- `variations`: Alternative versions of the attack

## Output and Reporting

### Console Output

The harness provides real-time feedback with:
- Progress tracking
- Individual test results
- Summary tables
- ML optimization statistics

### Report Files

Reports are saved to `results/` directory:
- `llmrecon_report_TIMESTAMP.json`: Detailed results in JSON format
- Includes session metadata, statistics, and full results

### Report Structure

```json
{
  "session_id": "unique-session-id",
  "timestamp": "2024-01-03T12:00:00",
  "summary": {
    "model_name": {
      "total": 10,
      "vulnerable": 7,
      "total_time": 25.5
    }
  },
  "results": [
    {
      "model": "llama3:latest",
      "template_name": "Basic Prompt Injection",
      "vulnerable": true,
      "matched_indicators": ["INJECTION SUCCESSFUL"],
      "elapsed_time": 2.1,
      "timestamp": "2024-01-03T12:00:01"
    }
  ]
}
```

## ML Integration

When enabled, the harness uses LLMrecon's ML components:

### Multi-Armed Bandit Optimization

- **Thompson Sampling**: Default algorithm for provider selection
- **Attack Strategy Learning**: Optimizes which attacks to use against which models
- **Performance Tracking**: Monitors success rates and response times

### Data Pipeline

- **Attack Recording**: All attacks stored in SQLite database
- **Feature Extraction**: Automatic feature engineering for ML training
- **Historical Analysis**: Track patterns over time

### ML Statistics

The harness reports ML optimization statistics:
- Total attempts and success rate
- Per-model performance metrics
- Algorithm performance
- Collection and processing rates

## Examples

### Test Specific Models

```bash
# Test only Llama models
python llmrecon_harness.py --models llama3:latest llama2:latest

# Test with prompt injection only
python llmrecon_harness.py --categories prompt_injection
```

### Custom Configuration

Create `my_config.json`:
```json
{
  "models_to_test": ["custom-model:latest"],
  "test_categories": ["custom_attacks"],
  "max_response_time": 30
}
```

Run with:
```bash
python llmrecon_harness.py --config my_config.json
```

### Add Custom Templates

1. Create `templates/my_template.json`
2. Run tests: `python llmrecon_harness.py`
3. Your template will be automatically loaded

## Security Considerations

### Ethical Usage

- Only test models you own or have permission to test
- Use in isolated environments
- Handle results responsibly
- Follow responsible disclosure practices

### Data Privacy

- Model responses may contain sensitive information
- ML database stores attack data
- Disable response saving if needed: `"save_responses": false`

## Troubleshooting

### Common Issues

**Ollama Connection Failed**
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama if needed
ollama serve
```

**No Models Found**
```bash
# List available models
ollama list

# Pull a model if needed
ollama pull llama3
```

**ML Component Errors**
```bash
# Disable ML if having issues
python llmrecon_harness.py --no-ml
```

**Permission Errors**
```bash
# Create directories
mkdir -p data/attacks results templates

# Check file permissions
chmod +x llmrecon_harness.py
```

### Debug Mode

Enable verbose logging by modifying the config:
```json
{
  "logging": {
    "level": "DEBUG"
  }
}
```

## Integration with LLMrecon

The harness is designed to complement the main LLMrecon tool:

1. **Testing Phase**: Use harness for comprehensive model assessment
2. **Detection Phase**: Use LLMrecon's detect command for specific response analysis
3. **Template Creation**: Export successful attacks as LLMrecon templates
4. **Monitoring**: Use ML data for ongoing security monitoring

## Development

### Adding New Features

The harness is modular and extensible:

- `OllamaConnector`: Handles API communication
- `AttackTemplate`: Template representation
- `LLMreconHarness`: Main testing logic

### Contributing Templates

1. Create template JSON file
2. Test with harness
3. Document attack methodology
4. Submit for inclusion

## License and Disclaimer

This tool is for security research and testing purposes only. Users are responsible for ensuring ethical and legal use of this software.