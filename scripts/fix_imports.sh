#!/bin/bash

# Script to fix import paths in Go files
# Changes github.com/your-org/LLMrecon to github.com/LLMrecon/LLMrecon

# Find all Go files with the old import path
find /Users/nconsolo/LLMrecon/src/security/access -type f -name "*.go" -exec grep -l "github.com/your-org/LLMrecon" {} \; | while read -r file; do
  echo "Fixing imports in $file"
  # Replace the import path
  sed -i '' 's|github.com/your-org/LLMrecon|github.com/LLMrecon/LLMrecon|g' "$file"
done

# Also check for any files with the old nconsolo import path
find /Users/nconsolo/LLMrecon/src/security/access -type f -name "*.go" -exec grep -l "github.com/nconsolo/LLMrecon" {} \; | while read -r file; do
  echo "Fixing imports in $file"
  # Replace the import path
  sed -i '' 's|github.com/nconsolo/LLMrecon|github.com/LLMrecon/LLMrecon|g' "$file"
done

echo "Import paths fixed!"
