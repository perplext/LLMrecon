#!/bin/bash

echo "Fixing final compilation errors..."

# Fix api/scan/service.go - remove extra brace on line 18
sed -i '' '18d' src/api/scan/service.go

# Fix missing 'context' import in files with 'unexpected name context' errors
find src -name "*.go" -exec grep -l "context\.Context" {} \; | while read file; do
    if ! grep -q "^import.*context" "$file" && ! grep -q '"context"' "$file"; then
        # Add context import if missing
        sed -i '' '/^import (/a\
	"context"
' "$file"
    fi
done

# Add missing encoding/binary import for enhanced_handler.go
if ! grep -q '"encoding/binary"' src/bundle/errors/enhanced_handler.go; then
    sed -i '' '/^import (/a\
	"encoding/binary"
' src/bundle/errors/enhanced_handler.go
fi

echo "Testing compilation..."
go build -o /tmp/test ./src/main.go 2>&1 | grep -c "syntax error" || echo "0"
