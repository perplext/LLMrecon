#!/bin/bash

# Fix template management package imports
# Replace full module paths with relative imports

echo "Fixing template management package imports..."

# Find all Go files in template/management that have the problematic imports
find src/template/management -name "*.go" -type f -exec grep -l "github.com/perplext/LLMrecon" {} \; > template_files.txt

while IFS= read -r file; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        
        # Replace full imports with relative imports
        sed -i.bak \
            -e 's|"github.com/perplext/LLMrecon/src/template/format"|"../../format"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/interfaces"|"./interfaces"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/cache"|"./cache"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/execution"|"./execution"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/loader"|"./loader"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/parser"|"./parser"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/registry"|"./registry"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/reporting"|"./reporting"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/types"|"./types"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/validation"|"./validation"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/optimization"|"./optimization"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/monitoring"|"./monitoring"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/structure"|"./structure"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/streaming"|"./streaming"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/ratelimit"|"./ratelimit"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/loaders"|"./loaders"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/benchmark"|"./benchmark"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management/execution/optimizer"|"./execution/optimizer"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management"|"."|g' \
            "$file"
        
        # Remove backup file
        rm -f "$file.bak"
    fi
done < template_files.txt

# Clean up
rm -f template_files.txt

echo "Template management import fixes complete!"