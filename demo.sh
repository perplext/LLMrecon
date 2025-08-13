#!/bin/bash
# LLMrecon Test Harness Demo Script

echo "🔴 LLMrecon Test Harness Demo"
echo "============================="
echo

# Check if Ollama is running
echo "Checking Ollama status..."
if curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "✅ Ollama is running"
else
    echo "❌ Ollama is not running. Please start it with: ollama serve"
    exit 1
fi

echo

# List available models
echo "📋 Available Models:"
python3 llmrecon_harness.py --list-models
echo

# List available templates
echo "🎯 Available Attack Templates:"
python3 llmrecon_harness.py --list-templates
echo

# Run basic demonstration
echo "🚀 Running Quick Demo (Llama3 + Prompt Injection)..."
echo "This will test Llama3 with prompt injection attacks"
echo

python3 llmrecon_harness.py --models llama3:latest --categories prompt_injection

echo
echo "✅ Demo completed!"
echo
echo "📊 Check the results/ directory for detailed reports"
echo "🗄️  ML data is stored in data/attacks/"
echo
echo "Next steps:"
echo "  • Run full test suite: python3 llmrecon_harness.py"
echo "  • Test multiple models: python3 llmrecon_harness.py --models llama3:latest qwen3:latest"
echo "  • Add custom templates in templates/ directory"
echo "  • Check TEST-HARNESS-README.md for detailed usage"