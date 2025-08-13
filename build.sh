#!/bin/bash

echo "Building LLMrecon..."

# Create build directory
mkdir -p build

# Build with specific packages to avoid compilation errors
go build -o build/LLMrecon \
    -ldflags "-X main.Version=0.1.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    ./src/main.go

if [ $? -eq 0 ]; then
    echo "Build successful! Binary created at: build/LLMrecon"
    echo "Run './build/LLMrecon --help' to see available commands"
else
    echo "Build failed. Trying minimal build..."
    
    # Try a minimal build excluding problematic packages
    go build -tags minimal -o build/LLMrecon \
        -ldflags "-X main.Version=0.1.0-minimal -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        ./src/main.go
fi