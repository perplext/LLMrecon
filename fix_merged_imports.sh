#!/bin/bash

# Fix merged import lines where two imports were concatenated

echo "Fixing merged import lines..."

# Get list of files with import syntax errors
go build ./src/... 2>&1 | grep "expected ';', found" | cut -d':' -f1 | sort -u | while read -r file; do
    if [[ -f "$file" ]]; then
        echo "Processing $file"
        
        # Read the file line by line and fix merged imports
        # Look for patterns like: cryptorand "crypto/rand" "fmt"
        # or: something "package1" "package2"
        perl -i -pe '
            # Fix lines that have merged imports like: alias "package1" "package2"
            s/(\s+\w+\s+"[^"]+")(\s+"[^"]+")/\1\n\t\2/g;
            # Fix lines that have merged imports without alias: "package1" "package2"  
            s/(\s+"[^"]+")(\s+"[^"]+")/\1\n\t\2/g;
        ' "$file"
    fi
done

echo "Merged import fixing completed."