#!/bin/bash

# Fix relative imports back to absolute module paths
# Go modules don't support relative imports

echo "Converting relative imports back to absolute module paths..."

# Find all files with relative imports
find . -name "*.go" -type f -exec grep -l '"\.\/' {} \; > relative_files.txt

while IFS= read -r file; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        
        # Convert relative imports back to absolute paths
        sed -i.bak \
            -e 's|"\.\/common"|"github.com/perplext/LLMrecon/src/security/access/common"|g' \
            -e 's|"\.\/mfa"|"github.com/perplext/LLMrecon/src/security/access/mfa"|g' \
            -e 's|"\.\/models"|"github.com/perplext/LLMrecon/src/security/access/models"|g' \
            -e 's|"\.\/rbac"|"github.com/perplext/LLMrecon/src/security/access/rbac"|g' \
            -e 's|"\.\/interfaces"|"github.com/perplext/LLMrecon/src/security/access/interfaces"|g' \
            -e 's|"\.\/audit"|"github.com/perplext/LLMrecon/src/security/access/audit"|g' \
            -e 's|"\.\/adapters"|"github.com/perplext/LLMrecon/src/security/access/adapters"|g' \
            -e 's|"\.\/converters"|"github.com/perplext/LLMrecon/src/security/access/converters"|g' \
            -e 's|"\.\/db"|"github.com/perplext/LLMrecon/src/security/access/db"|g' \
            -e 's|"\.\/impl"|"github.com/perplext/LLMrecon/src/security/access/impl"|g' \
            -e 's|"\.\/types"|"github.com/perplext/LLMrecon/src/security/access/types"|g' \
            -e 's|"\.\."|"github.com/perplext/LLMrecon/src/security/access"|g' \
            "$file"
        
        # Fix template management imports
        sed -i.bak \
            -e 's|"\.\/cache"|"github.com/perplext/LLMrecon/src/template/management/cache"|g' \
            -e 's|"\.\/execution"|"github.com/perplext/LLMrecon/src/template/management/execution"|g' \
            -e 's|"\.\/interfaces"|"github.com/perplext/LLMrecon/src/template/management/interfaces"|g' \
            -e 's|"\.\/loader"|"github.com/perplext/LLMrecon/src/template/management/loader"|g' \
            -e 's|"\.\/parser"|"github.com/perplext/LLMrecon/src/template/management/parser"|g' \
            -e 's|"\.\/registry"|"github.com/perplext/LLMrecon/src/template/management/registry"|g' \
            -e 's|"\.\/reporting"|"github.com/perplext/LLMrecon/src/template/management/reporting"|g' \
            -e 's|"\.\/types"|"github.com/perplext/LLMrecon/src/template/management/types"|g' \
            -e 's|"\.\/validation"|"github.com/perplext/LLMrecon/src/template/management/validation"|g' \
            -e 's|"\.\/optimization"|"github.com/perplext/LLMrecon/src/template/management/optimization"|g' \
            -e 's|"\.\/monitoring"|"github.com/perplext/LLMrecon/src/template/management/monitoring"|g' \
            -e 's|"\.\/structure"|"github.com/perplext/LLMrecon/src/template/management/structure"|g' \
            -e 's|"\.\/streaming"|"github.com/perplext/LLMrecon/src/template/management/streaming"|g' \
            -e 's|"\.\/ratelimit"|"github.com/perplext/LLMrecon/src/template/management/ratelimit"|g' \
            -e 's|"\.\/loaders"|"github.com/perplext/LLMrecon/src/template/management/loaders"|g' \
            -e 's|"\.\/benchmark"|"github.com/perplext/LLMrecon/src/template/management/benchmark"|g' \
            -e 's|"\.\.\/\.\.\/format"|"github.com/perplext/LLMrecon/src/template/format"|g' \
            "$file"
        
        # Fix testing/owasp imports
        sed -i.bak \
            -e 's|"\.\/types"|"github.com/perplext/LLMrecon/src/testing/owasp/types"|g' \
            -e 's|"\.\/mocks"|"github.com/perplext/LLMrecon/src/testing/owasp/mocks"|g' \
            -e 's|"\.\/fixtures"|"github.com/perplext/LLMrecon/src/testing/owasp/fixtures"|g' \
            -e 's|"\.\/validation"|"github.com/perplext/LLMrecon/src/testing/owasp/validation"|g' \
            -e 's|"\.\/compliance"|"github.com/perplext/LLMrecon/src/testing/owasp/compliance"|g' \
            "$file"
        
        # Fix reporting imports
        sed -i.bak \
            -e 's|"\.\/common"|"github.com/perplext/LLMrecon/src/reporting/common"|g' \
            -e 's|"\.\/api"|"github.com/perplext/LLMrecon/src/reporting/api"|g' \
            -e 's|_ "\.\/formats"|_ "github.com/perplext/LLMrecon/src/reporting/formats"|g' \
            "$file"
        
        # Remove backup file
        rm -f "$file.bak"
    fi
done < relative_files.txt

# Clean up
rm -f relative_files.txt

echo "Relative import conversion complete!"