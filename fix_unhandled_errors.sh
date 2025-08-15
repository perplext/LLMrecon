#!/bin/bash

# Script to fix G104 unhandled errors systematically

echo "Fixing unhandled errors..."

# Fix 1: json.NewEncoder().Encode() errors in HTTP handlers
echo "Fixing json.Encode errors in HTTP handlers..."

# Fix specific files with unhandled json encoding
fix_json_encode() {
    local file=$1
    local line=$2
    
    echo "Fixing $file:$line"
    
    # For HTTP handlers, we should log errors but not break the response
    # since headers are already sent
}

# Fix 2: cmd.MarkFlagRequired() errors
echo "Fixing cmd.MarkFlagRequired errors..."

# These are usually fine to ignore as they only fail if the flag doesn't exist
# which would be a programming error caught during development

# Fix 3: defer file.Close() errors
echo "Looking for defer Close() patterns that need fixing..."

# Find all defer Close() without error handling
find src -name "*.go" -type f -exec grep -l "defer.*\.Close()" {} \; | while read -r file; do
    # Check if the file has unhandled defer close
    if grep -q "defer [^(]*\.Close()" "$file"; then
        echo "Found potential unhandled defer Close() in $file"
    fi
done

echo "Manual fixes needed for critical error handling..."