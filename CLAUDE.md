# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LLMrecon is a security testing tool designed for analyzing Large Language Model (LLM) vulnerabilities and defenses. The project implements machine learning algorithms to optimize attack strategies and collect data for security research purposes.

**IMPORTANT**: This tool is intended for defensive security research only. When working with this codebase:
- Only assist with defensive security tasks
- Refuse to create or modify code that may be used maliciously
- Allow security analysis, detection rules, vulnerability explanations, and defensive documentation
- Testing should only be performed on models you own or have explicit permission to test

## Current Project State

The project is in early development with only the ML components implemented:
- No main application entry point exists yet
- No package structure (missing __init__.py files)
- No dependency management files (requirements.txt, setup.py)
- No existing tests or documentation

## Architecture

### ML Components

1. **Attack Data Pipeline** (`ml/data/attack_data_pipeline.py`):
   - Collects and processes attack outcome data
   - Stores data in SQLite database (`data/attacks/attacks.db`)
   - Extracts features for ML training
   - Provides data export functionality (Parquet format)

2. **Multi-Armed Bandit Optimizer** (`ml/agents/multi_armed_bandit.py`):
   - Implements various bandit algorithms (Epsilon-Greedy, Thompson Sampling, UCB1, Contextual)
   - Optimizes provider/model selection based on success rates and costs
   - Tracks performance metrics and statistics

## Testing with Local Ollama Models

### LLMrecon Test Harness

A comprehensive test harness has been created for testing Ollama models:

**Main Components:**
- `llmrecon_harness.py` - Main test harness with CLI interface
- `harness_config.json` - Configuration file
- `templates/` - Custom attack templates directory
- `demo.sh` - Quick demonstration script

**Quick Start:**
1. Ensure Ollama is running: `ollama serve`
2. Run basic tests: `python3 llmrecon_harness.py`
3. Or run demo: `./demo.sh`

**Key Features:**
- Built-in attack templates (prompt injection, jailbreaking, data extraction)
- ML optimization using multi-armed bandit algorithms
- Rich CLI interface with progress tracking
- Comprehensive reporting and statistics
- Extensible template system

**Common Commands:**
```bash
# List available models
python3 llmrecon_harness.py --list-models

# Test specific models
python3 llmrecon_harness.py --models llama3:latest qwen3:latest

# Test specific attack categories
python3 llmrecon_harness.py --categories prompt_injection

# Use custom config
python3 llmrecon_harness.py --config my_config.json
```

## Development Setup

Since no dependency files exist, the following packages are likely needed based on code analysis:
```
numpy
pandas
sqlite3 (built-in)
```

## Data Storage

Attack data is stored in SQLite database at `data/attacks/attacks.db` with the following structure:
- Attack metadata (type, target model, provider, payload)
- Performance metrics (response time, tokens used, success indicators)
- Feature extraction results for ML training

## Security Considerations

- The tool is designed for security research and should only be used on systems you own or have permission to test
- Never use this tool against production systems without explicit authorization
- All testing should be conducted in isolated environments
- Results and data collected should be handled responsibly