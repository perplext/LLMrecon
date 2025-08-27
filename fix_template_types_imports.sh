#!/bin/bash

# Fix imports in template/management files that incorrectly reference security/access/types
# These should use template/management/types instead

echo "Fixing template management types imports..."

# List of files to fix
files=(
    "src/template/management/benchmark/benchmark.go"
    "src/template/management/optimized_manager.go"
    "src/template/management/management.go"
    "src/template/management/loaders/offline_bundle_loader.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "Fixing $file..."
        sed -i '' 's|github.com/perplext/LLMrecon/src/security/access/types|github.com/perplext/LLMrecon/src/template/management/types|g' "$file"
    fi
done

echo "Done fixing template management types imports."