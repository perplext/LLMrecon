#!/bin/bash

# Fix reporting package imports
echo "Fixing reporting package imports..."

# Files in root reporting directory
for file in src/reporting/*.go; do
    if [ -f "$file" ] && grep -q "github.com/perplext/LLMrecon/src/reporting" "$file"; then
        echo "Processing $file..."
        sed -i.bak \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/common"|"./common"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/api"|"./api"|g' \
            -e 's|_ "github.com/perplext/LLMrecon/src/reporting/formats"|_ "./formats"|g' \
            "$file"
        rm -f "$file.bak"
    fi
done

# Files in reporting/interfaces
for file in src/reporting/interfaces/*.go; do
    if [ -f "$file" ] && grep -q "github.com/perplext/LLMrecon/src/reporting" "$file"; then
        echo "Processing $file..."
        sed -i.bak \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/common"|"../common"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/api"|"../api"|g' \
            "$file"
        rm -f "$file.bak"
    fi
done

# Files in reporting/formats
for file in src/reporting/formats/*.go; do
    if [ -f "$file" ] && grep -q "github.com/perplext/LLMrecon/src/reporting" "$file"; then
        echo "Processing $file..."
        sed -i.bak \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/api"|"../api"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/common"|"../common"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/formats/types"|"./types"|g' \
            "$file"
        rm -f "$file.bak"
    fi
done

# Files in reporting/formats/types
for file in src/reporting/formats/types/*.go; do
    if [ -f "$file" ] && grep -q "github.com/perplext/LLMrecon/src/reporting" "$file"; then
        echo "Processing $file..."
        sed -i.bak \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/api"|"../../api"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/reporting/common"|"../../common"|g' \
            "$file"
        rm -f "$file.bak"
    fi
done

echo "Reporting package import fixes complete!"