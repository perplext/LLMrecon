#!/bin/bash

# Fix remaining package imports across the codebase
# This script handles imports from main files in cmd/ and examples/ directories
# and any other remaining internal imports

echo "Fixing remaining package imports across the codebase..."

# Find all remaining Go files that have the problematic imports
find . -name "*.go" -type f -exec grep -l "github.com/perplext/LLMrecon/src" {} \; > remaining_files.txt

while IFS= read -r file; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        
        # For files in cmd/ and examples/ directories, we need absolute imports from the module root
        # We'll replace the full github path with just the internal module path
        
        # Replace the full github.com module path with just the internal path
        sed -i.bak \
            -e 's|github.com/perplext/LLMrecon/src/|github.com/perplext/LLMrecon/src/|g' \
            "$file"
        
        # For files that are at the root level (cmd/, examples/) we keep absolute imports
        # But for internal package files, we might need relative imports
        
        # The pattern here is to keep absolute imports for main packages
        # and only fix relative imports within the same package hierarchy
        
        # Remove backup file
        rm -f "$file.bak"
    fi
done < remaining_files.txt

# Clean up
rm -f remaining_files.txt

echo "Remaining import fixes complete!"

# Show count of remaining issues
echo "Remaining files with potential issues:"
find . -name "*.go" -type f -exec grep -l "github.com/perplext/LLMrecon/src" {} \; | wc -l