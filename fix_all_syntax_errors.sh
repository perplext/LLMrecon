#!/bin/bash

echo "Comprehensive syntax error fix..."

# Remove duplicate imports
find src -name "*.go" -exec sed -i '' '/^import ($/,/^)$/{/encoding\/binary/d;}' {} \;
find src -name "*.go" -exec sed -i '' '1,/^import ($/!b; /^import ($/,/^)$/{ N; s/\n\t"encoding\/binary"\n\t"encoding\/binary"/\n\t"encoding\/binary"/;}' {} \;

# Fix enhanced_handler.go specifically
sed -i '' '5,7d' src/bundle/errors/enhanced_handler.go

# Check for functions missing closing braces by counting braces
for file in $(find src -name "*.go"); do
    open_braces=$(grep -c '^func ' "$file" 2>/dev/null || echo 0)
    func_close=$(grep -c '^}$' "$file" 2>/dev/null || echo 0)
    
    # If we have more functions than closing braces at the root level, add them
    if [ "$open_braces" -gt 0 ] && [ "$func_close" -lt "$open_braces" ]; then
        diff=$((open_braces - func_close))
        for i in $(seq 1 $diff); do
            echo "}" >> "$file"
        done
    fi
done

echo "Checking compilation..."
go build -o /tmp/test ./src/main.go 2>&1 | grep -c "syntax error" || echo "0"
