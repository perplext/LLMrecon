#!/bin/bash

# Fix all import paths in Go files across the entire project
echo "Fixing all import paths in the project..."

# Change to project root
cd /Users/nconsolo/claude-code/LLMrecon

# Find and fix all Go files
find . -name "*.go" -type f | while read -r file; do
    # Skip vendor and .git directories
    if [[ "$file" == *"/vendor/"* ]] || [[ "$file" == *"/.git/"* ]]; then
        continue
    fi
    
    # Check if file has any incorrect imports
    if grep -q -E '"(github\.com/(llm-security|yourusername|your-org|nconsolo)/LLMrecon|LLMrecon/src/)' "$file"; then
        echo "Fixing: $file"
        
        # Fix all variations of incorrect imports
        sed -i '' \
            -e 's|"github.com/llm-security/LLMrecon/|"github.com/LLMrecon/LLMrecon/|g' \
            -e 's|"github.com/yourusername/LLMrecon/|"github.com/LLMrecon/LLMrecon/|g' \
            -e 's|"github.com/your-org/LLMrecon/|"github.com/LLMrecon/LLMrecon/|g' \
            -e 's|"github.com/nconsolo/LLMrecon/|"github.com/LLMrecon/LLMrecon/|g' \
            -e 's|"LLMrecon/src/|"github.com/LLMrecon/LLMrecon/src/|g' \
            "$file"
    fi
done

echo "All import paths have been fixed!"