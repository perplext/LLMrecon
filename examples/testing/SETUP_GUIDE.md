# LLMrecon Testing Framework Setup Guide

This guide helps you set up the LLMrecon testing environment on different operating systems.

## Quick Start

### 1. Prerequisites

- Python 3.8 or higher
- Go 1.19 or higher (for building LLMrecon)
- Ollama installed and running
- Git for cloning the repository

### 2. Clone and Build

```bash
# Clone the repository
git clone https://github.com/perplext/LLMrecon.git
cd LLMrecon

# Build LLMrecon
go build -o llmrecon ./src/main.go
# Or use make if available
make build
```

### 3. Install Python Dependencies

```bash
# Navigate to testing directory
cd examples/testing

# Create virtual environment (recommended)
python3 -m venv venv

# Activate virtual environment
# On macOS/Linux:
source venv/bin/activate
# On Windows:
# venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
```

## Platform-Specific Setup

### macOS

```bash
# Install Homebrew if not present
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew install python3 go

# Install Ollama
brew install ollama

# Start Ollama service
ollama serve
```

### Linux (Ubuntu/Debian)

```bash
# Update package manager
sudo apt update

# Install dependencies
sudo apt install python3 python3-pip golang-go

# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama service
ollama serve
```

### Windows

```powershell
# Install Chocolatey if not present
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install dependencies
choco install python3 golang

# Install Ollama
# Download from: https://ollama.com/download/windows

# Start Ollama
ollama serve
```

## Environment Configuration

### 1. Ollama Models

```bash
# Pull required models
ollama pull llama3
ollama pull qwen3
ollama pull llama2  # Optional

# Verify models
ollama list
```

### 2. Test Configuration

Create a custom configuration file:

```json
{
    "ollama_url": "http://localhost:11434",
    "enable_ml": false,
    "output_dir": "./results",
    "templates_dir": "./templates",
    "models_to_test": ["llama3:latest", "qwen3:latest"],
    "test_categories": ["prompt_injection", "jailbreak"],
    "max_response_time": 60,
    "save_responses": true
}
```

Save as `config.json` in the testing directory.

## Running Tests

### Basic Testing

```bash
# Run test harness
python llmrecon_harness.py

# Run with custom config
python llmrecon_harness.py --config config.json

# Test specific models
python llmrecon_harness.py --models llama3:latest

# Test specific categories
python llmrecon_harness.py --categories prompt_injection
```

### Advanced Testing

```bash
# Run comprehensive test suite
python llmrecon_harness.py --comprehensive

# Generate detailed report
python llmrecon_harness.py --report detailed

# Export results
python llmrecon_harness.py --export json --output results.json
```

## Troubleshooting

### Common Issues

#### 1. Ollama Connection Error
```
Error: Connection refused to http://localhost:11434
```
**Solution**: Ensure Ollama is running with `ollama serve`

#### 2. Model Not Found
```
Error: Model 'llama3:latest' not found
```
**Solution**: Pull the model with `ollama pull llama3`

#### 3. Python Import Error
```
ImportError: No module named 'requests'
```
**Solution**: Install dependencies with `pip install -r requirements.txt`

#### 4. Go Build Error
```
go: cannot find main module
```
**Solution**: Ensure you're in the repository root directory

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Set log level
export LLMRECON_LOG_LEVEL=DEBUG

# Run with debug output
python llmrecon_harness.py --debug
```

## Advanced Configuration

### Enabling ML Components

If you have the full ML components available:

1. Install additional dependencies:
```bash
pip install numpy scikit-learn sqlalchemy pandas
```

2. Update configuration:
```json
{
    "enable_ml": true,
    "ml_storage_path": "./data/attacks",
    "ml_algorithm": "thompson_sampling"
}
```

3. Import ML modules in scripts (uncomment imports)

### Custom Templates

Create custom attack templates in `templates/` directory:

```yaml
id: custom-injection-test
info:
  name: Custom Injection Test
  severity: high
test:
  prompt: "Your custom test prompt here"
  indicators:
    - "success_keyword"
    - "another_indicator"
```

## Support

For issues or questions:
- Check the [GitHub Issues](https://github.com/perplext/LLMrecon/issues)
- Review documentation in the `docs/` directory
- Consult the troubleshooting section above