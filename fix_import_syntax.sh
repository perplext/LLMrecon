#!/bin/bash

# Fix malformed import statements in Go files
# This script fixes imports that are missing proper parentheses or semicolons

echo "Fixing import syntax errors..."

# Find files with malformed import statements
files_with_errors=$(go build ./src/... 2>&1 | grep "expected ';', found" | cut -d':' -f1 | sort -u)

for file in $files_with_errors; do
    if [[ -f "$file" ]]; then
        echo "Fixing import syntax in $file"
        
        # Fix malformed import blocks - look for lines that start with import but don't have proper syntax
        # This handles cases where "import" is followed directly by package names without parentheses
        sed -i.bak '
            # Fix import statements that are missing parentheses
            /^import [^(]/ {
                # If line starts with "import" followed by a space and then a package name (not parentheses)
                s/^import \([^(].*\)/import (\
\t\1\
)/
            }
            
            # Fix cases where import block is malformed
            /^import$/ {
                N
                s/^import\n\([^(]\)/import (\
\t\1\
)/
            }
        ' "$file"
        
        # Remove backup file
        rm -f "${file}.bak"
    fi
done

echo "Import syntax fixing completed."