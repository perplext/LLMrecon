#!/bin/bash

# Fix testing/owasp package imports
# Replace full module paths with relative imports

echo "Fixing testing/owasp package imports..."

# Find all Go files in testing/owasp that have the problematic imports
find src/testing/owasp -name "*.go" -type f -exec grep -l "github.com/perplext/LLMrecon" {} \; > owasp_files.txt

while IFS= read -r file; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        
        # Replace full imports with relative imports based on file location
        case "$file" in
            # Files in root testing/owasp/
            src/testing/owasp/*.go)
                sed -i.bak \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/types"|"./types"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/mocks"|"./mocks"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/fixtures"|"./fixtures"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/validation"|"./validation"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/compliance"|"./compliance"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp"|"."|g' \
                    "$file"
                ;;
            # Files in subfolders
            src/testing/owasp/*/*)
                sed -i.bak \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/types"|"../types"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/mocks"|"../mocks"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/fixtures"|"../fixtures"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/validation"|"../validation"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp/compliance"|"../compliance"|g' \
                    -e 's|"github.com/perplext/LLMrecon/src/testing/owasp"|".."|g' \
                    "$file"
                ;;
        esac
        
        # Common replacements for external module paths that are used by all files
        sed -i.bak \
            -e 's|"github.com/perplext/LLMrecon/src/provider"|"../../provider"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/format"|"../../template/format"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/template/management"|"../../template/management"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/repository"|"../../repository"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/bundle"|"../../bundle"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/config"|"../../config"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/audit"|"../../audit"|g' \
            "$file"
        
        # Remove backup file
        rm -f "$file.bak"
    fi
done < owasp_files.txt

# Clean up
rm -f owasp_files.txt

echo "Testing/owasp import fixes complete!"