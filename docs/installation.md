# Installation Guide

This guide covers all installation methods for the LLMrecon security testing tool.

## Table of Contents

- [System Requirements](#system-requirements)
- [Installation Methods](#installation-methods)
  - [Pre-built Binaries](#pre-built-binaries)
  - [Building from Source](#building-from-source)
  - [Docker Installation](#docker-installation)
  - [Package Managers](#package-managers)
- [Configuration](#configuration)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## System Requirements

### Minimum Requirements

- **Operating System**: Linux, macOS, or Windows 10/11
- **Memory**: 4GB RAM (8GB recommended for large-scale testing)
- **Storage**: 500MB for tool and templates
- **Network**: Internet connection for LLM API access and updates

### Software Dependencies

- **Go**: Version 1.21 or higher (for building from source)
- **Git**: For cloning repository and template updates
- **OpenSSL**: For certificate verification

## Installation Methods

### Pre-built Binaries

The easiest way to install LLMrecon is using pre-built binaries:

```bash
# Download the latest release for your platform
# Linux (amd64)
wget https://github.com/your-org/LLMrecon/releases/latest/download/LLMrecon-linux-amd64.tar.gz
tar -xzf LLMrecon-linux-amd64.tar.gz
sudo mv LLMrecon /usr/local/bin/

# macOS (Intel)
wget https://github.com/your-org/LLMrecon/releases/latest/download/LLMrecon-darwin-amd64.tar.gz
tar -xzf LLMrecon-darwin-amd64.tar.gz
sudo mv LLMrecon /usr/local/bin/

# macOS (Apple Silicon)
wget https://github.com/your-org/LLMrecon/releases/latest/download/LLMrecon-darwin-arm64.tar.gz
tar -xzf LLMrecon-darwin-arm64.tar.gz
sudo mv LLMrecon /usr/local/bin/

# Windows
# Download LLMrecon-windows-amd64.zip from releases page
# Extract and add to PATH
```

### Building from Source

For development or customization, build from source:

```bash
# Prerequisites
# Install Go 1.21+ from https://golang.org/dl/

# Clone the repository
git clone https://github.com/your-org/LLMrecon.git
cd LLMrecon

# Download dependencies
go mod download

# Build the binary
go build -o LLMrecon ./src/main.go

# Install globally (optional)
sudo mv LLMrecon /usr/local/bin/

# Or use make (if available)
make build
make install
```

#### Build Options

```bash
# Build with version information
go build -ldflags="-X main.version=1.0.0" -o LLMrecon ./src/main.go

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o LLMrecon-linux ./src/main.go
GOOS=darwin GOARCH=arm64 go build -o LLMrecon-mac ./src/main.go
GOOS=windows GOARCH=amd64 go build -o LLMrecon.exe ./src/main.go

# Build with optimization
go build -ldflags="-s -w" -o LLMrecon ./src/main.go
```

### Docker Installation

Run LLMrecon in a container:

```bash
# Pull the official image
docker pull your-org/LLMrecon:latest

# Run with configuration
docker run -it \
  -v ~/.LLMrecon:/root/.LLMrecon \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  your-org/LLMrecon:latest scan --help

# Build your own image
docker build -t LLMrecon .

# Docker Compose setup
cat > docker-compose.yml <<EOF
version: '3.8'
services:
  LLMrecon:
    image: your-org/LLMrecon:latest
    volumes:
      - ./config:/root/.LLMrecon
      - ./results:/results
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
EOF

docker-compose run LLMrecon scan --help
```

### Package Managers

#### Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap your-org/LLMrecon

# Install
brew install LLMrecon

# Upgrade
brew upgrade LLMrecon
```

#### APT (Debian/Ubuntu)

```bash
# Add the repository
echo "deb https://apt.your-org.com stable main" | sudo tee /etc/apt/sources.list.d/LLMrecon.list

# Add the GPG key
curl -fsSL https://apt.your-org.com/gpg | sudo apt-key add -

# Install
sudo apt update
sudo apt install LLMrecon
```

#### YUM/DNF (RHEL/Fedora)

```bash
# Add the repository
sudo cat > /etc/yum.repos.d/LLMrecon.repo <<EOF
[LLMrecon]
name=LLMrecon Repository
baseurl=https://yum.your-org.com/stable/\$basearch
enabled=1
gpgcheck=1
gpgkey=https://yum.your-org.com/gpg
EOF

# Install
sudo yum install LLMrecon
# or
sudo dnf install LLMrecon
```

## Configuration

### Initial Setup

After installation, configure the tool:

```bash
# Initialize configuration
LLMrecon init

# This creates:
# ~/.LLMrecon/
#   ├── config.yaml      # Main configuration
#   ├── templates/       # Security test templates
#   ├── providers/       # Provider modules
#   └── logs/           # Application logs
```

### Provider Configuration

Configure your LLM providers:

```bash
# Configure OpenAI
LLMrecon config set provider openai
LLMrecon config set openai.api_key "sk-..."
LLMrecon config set openai.model "gpt-4"

# Configure Anthropic
LLMrecon config set anthropic.api_key "sk-ant-..."
LLMrecon config set anthropic.model "claude-3-opus-20240229"

# Configure Azure OpenAI
LLMrecon config set azure.endpoint "https://your-resource.openai.azure.com/"
LLMrecon config set azure.api_key "your-key"
LLMrecon config set azure.deployment "gpt-4"

# List all configuration
LLMrecon config list
```

### Environment Variables

You can use environment variables instead of storing keys in config:

```bash
# Set in your shell profile
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export LLM_RED_TEAM_CONFIG_DIR="$HOME/.LLMrecon"

# Use in config.yaml
providers:
  openai:
    api_key: ${OPENAI_API_KEY}
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
```

### Template Installation

Install security test templates:

```bash
# Update official templates
LLMrecon template update

# Add custom template repository
LLMrecon template add-source https://github.com/your-org/custom-templates

# List available templates
LLMrecon template list

# Install specific template pack
LLMrecon template install owasp-llm-top10
```

## Verification

Verify your installation:

```bash
# Check version
LLMrecon version

# Run diagnostics
LLMrecon doctor

# Expected output:
# ✓ Binary executable found
# ✓ Configuration directory exists
# ✓ Templates directory found (42 templates)
# ✓ OpenAI provider configured
# ✓ Network connectivity OK
# ✓ Template updates available

# Test with a simple scan
LLMrecon scan --target test --dry-run

# Check all components
LLMrecon self-test
```

## Troubleshooting

### Common Issues

#### Permission Denied

```bash
# If you get permission errors
sudo chmod +x /usr/local/bin/LLMrecon

# For config directory
chmod -R 755 ~/.LLMrecon
```

#### Command Not Found

```bash
# Add to PATH in your shell profile
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc

# For macOS with Homebrew
echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

#### SSL/TLS Errors

```bash
# For self-signed certificates
LLMrecon config set security.tls_verify false

# Or add your CA certificate
LLMrecon config set security.ca_bundle /path/to/ca-bundle.crt
```

#### API Key Issues

```bash
# Verify API keys are set
LLMrecon config get openai.api_key

# Test provider connection
LLMrecon provider test openai

# Check for special characters in keys
# Wrap in single quotes if needed
LLMrecon config set openai.api_key 'sk-...'
```

#### Template Loading Errors

```bash
# Reset templates
rm -rf ~/.LLMrecon/templates
LLMrecon template update --force

# Validate templates
LLMrecon template validate

# Check template syntax
LLMrecon template lint prompt-injection/*.yaml
```

### Getting Help

```bash
# Built-in help
LLMrecon help
LLMrecon scan --help

# Generate debug log
LLMrecon --debug scan ... 2>debug.log

# Community support
# GitHub Issues: https://github.com/your-org/LLMrecon/issues
# Discord: https://discord.gg/LLMrecon
```

## Next Steps

- Read the [User Guide](user-guide.md) to learn how to run security tests
- Check out [Example Workflows](examples.md) for common use cases
- Learn to write custom tests in the [Template Guide](template-guide.md)
- Set up CI/CD integration with our [Automation Guide](automation.md)