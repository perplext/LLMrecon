#!/bin/bash

echo "Cleaning up excess braces and fixing syntax..."

# Find and remove lines that only contain a single closing brace preceded by another closing brace
for file in $(find src -name "*.go"); do
    # Remove consecutive closing braces on separate lines (keep only one)
    sed -i '' '/^}$/{
        N
        s/^}\n}$/}/
    }' "$file"
    
    # Remove standalone closing braces that appear after struct field definitions
    sed -i '' '/^[[:space:]]*[A-Za-z].*[[:space:]].*$/!b
        N
        s/\n}$//
    ' "$file" 2>/dev/null || true
done

echo "Testing compilation..."
go build -o /tmp/test ./src/main.go 2>&1 | grep -c "syntax error" || echo "0"
