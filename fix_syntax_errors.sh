#!/bin/bash

echo "Fixing syntax errors in Go files..."

# Function to fix syntax errors in a file
fix_syntax_errors() {
    local file="$1"
    echo "Fixing syntax errors in $file..."
    
    # Remove malformed error checks and returns
    sed -i '' '/^if err != nil {$/N;/^if err != nil {\ntreturn err$/d' "$file"
    sed -i '' '/^if err != nil {$/N;/^if err != nil {\nreturn err$/d' "$file"
    
    # Fix malformed treturn statements
    sed -i '' 's/^treturn err$/return err/' "$file"
    sed -i '' 's/^treturn /return /' "$file"
    
    # Remove standalone error checks that don't make sense
    sed -i '' '/^if err != nil {$/d' "$file"
    sed -i '' '/^return err$/d' "$file"
    
    # Fix malformed lines starting with }
    sed -i '' '/^}[[:space:]]*if err != nil {$/d' "$file"
    sed -i '' '/^}[[:space:]]*treturn/d' "$file"
    sed -i '' '/^}[[:space:]]*return err$/d' "$file"
    
    # Fix lines that start with return but are orphaned
    sed -i '' '/^[[:space:]]*}[[:space:]]*$/N;/^[[:space:]]*}[[:space:]]*\nreturn/s/\nreturn/ return/' "$file"
}

# Fix files with syntax errors
fix_syntax_errors "src/bundle/errors/recovery.go"
fix_syntax_errors "src/bundle/errors/reporting.go"
fix_syntax_errors "src/security/access/audit/audit_logger.go"
fix_syntax_errors "src/audit/audit.go"
fix_syntax_errors "src/template/format/module.go"
fix_syntax_errors "src/template/format/template.go"
fix_syntax_errors "src/provider/core/connection_pool.go"
fix_syntax_errors "src/provider/core/factory.go"
fix_syntax_errors "src/security/api/ip_allowlist.go"
fix_syntax_errors "src/security/communication/cert_chain_utils.go"
fix_syntax_errors "src/security/prompt/advanced_template_monitor.go"
fix_syntax_errors "src/security/prompt/approval_workflow.go"
fix_syntax_errors "src/security/prompt/enhanced_approval_workflow.go"
fix_syntax_errors "src/config/config.go"
fix_syntax_errors "src/api/scan/handlers.go"

echo "Done fixing syntax errors!"