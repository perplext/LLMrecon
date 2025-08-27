#!/bin/bash

# Fix all remaining relative imports with absolute module paths
echo "Converting all remaining relative imports to absolute module paths..."

# Find all files with relative imports
find . -name "*.go" -type f -exec grep -l '\.\.\/' {} \; > relative_files.txt

while IFS= read -r file; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        
        # Template format imports
        sed -i.bak \
            -e 's|"\.\.\/\.\.\/format"|"github.com/perplext/LLMrecon/src/template/format"|g' \
            -e 's|"\.\.\/format"|"github.com/perplext/LLMrecon/src/template/format"|g' \
            "$file"
        
        # Security access package internal imports
        sed -i.bak \
            -e 's|"\.\.\/common"|"github.com/perplext/LLMrecon/src/security/access/common"|g' \
            -e 's|"\.\.\/mfa"|"github.com/perplext/LLMrecon/src/security/access/mfa"|g' \
            -e 's|"\.\.\/models"|"github.com/perplext/LLMrecon/src/security/access/models"|g' \
            -e 's|"\.\.\/rbac"|"github.com/perplext/LLMrecon/src/security/access/rbac"|g' \
            -e 's|"\.\.\/interfaces"|"github.com/perplext/LLMrecon/src/security/access/interfaces"|g' \
            -e 's|"\.\.\/audit"|"github.com/perplext/LLMrecon/src/security/access/audit"|g' \
            -e 's|"\.\.\/adapters"|"github.com/perplext/LLMrecon/src/security/access/adapters"|g' \
            -e 's|"\.\.\/converters"|"github.com/perplext/LLMrecon/src/security/access/converters"|g' \
            -e 's|"\.\.\/db/adapter"|"github.com/perplext/LLMrecon/src/security/access/db/adapter"|g' \
            -e 's|"\.\.\/db"|"github.com/perplext/LLMrecon/src/security/access/db"|g' \
            -e 's|"\.\.\/impl"|"github.com/perplext/LLMrecon/src/security/access/impl"|g' \
            -e 's|"\.\.\/types"|"github.com/perplext/LLMrecon/src/security/access/types"|g' \
            "$file"
        
        # Template management package imports
        sed -i.bak \
            -e 's|"\.\.\/\.\.\/template/management"|"github.com/perplext/LLMrecon/src/template/management"|g' \
            -e 's|"\.\.\/cache"|"github.com/perplext/LLMrecon/src/template/management/cache"|g' \
            -e 's|"\.\.\/execution"|"github.com/perplext/LLMrecon/src/template/management/execution"|g' \
            -e 's|"\.\.\/interfaces"|"github.com/perplext/LLMrecon/src/template/management/interfaces"|g' \
            -e 's|"\.\.\/loader"|"github.com/perplext/LLMrecon/src/template/management/loader"|g' \
            -e 's|"\.\.\/parser"|"github.com/perplext/LLMrecon/src/template/management/parser"|g' \
            -e 's|"\.\.\/registry"|"github.com/perplext/LLMrecon/src/template/management/registry"|g' \
            -e 's|"\.\.\/reporting"|"github.com/perplext/LLMrecon/src/template/management/reporting"|g' \
            -e 's|"\.\.\/types"|"github.com/perplext/LLMrecon/src/template/management/types"|g' \
            "$file"
        
        # Testing OWASP package imports
        sed -i.bak \
            -e 's|"\.\.\/types"|"github.com/perplext/LLMrecon/src/testing/owasp/types"|g' \
            -e 's|"\.\.\/mocks"|"github.com/perplext/LLMrecon/src/testing/owasp/mocks"|g' \
            -e 's|"\.\.\/fixtures"|"github.com/perplext/LLMrecon/src/testing/owasp/fixtures"|g' \
            -e 's|"\.\.\/validation"|"github.com/perplext/LLMrecon/src/testing/owasp/validation"|g' \
            -e 's|"\.\.\/compliance"|"github.com/perplext/LLMrecon/src/testing/owasp/compliance"|g' \
            "$file"
        
        # Reporting package imports
        sed -i.bak \
            -e 's|"\.\.\/common"|"github.com/perplext/LLMrecon/src/reporting/common"|g' \
            -e 's|"\.\.\/api"|"github.com/perplext/LLMrecon/src/reporting/api"|g' \
            -e 's|"\.\.\/\.\.\/api"|"github.com/perplext/LLMrecon/src/reporting/api"|g' \
            "$file"
        
        # Generic package navigation imports
        sed -i.bak \
            -e 's|"\.\."|"github.com/perplext/LLMrecon/src/security/access"|g' \
            "$file"
        
        # Remove backup file
        rm -f "$file.bak"
    fi
done < relative_files.txt

# Clean up
rm -f relative_files.txt

echo "All relative import conversion complete!"

# Show final status
echo "Remaining relative imports:"
find . -name "*.go" -type f -exec grep -l '\.\.\/' {} \; | wc -l