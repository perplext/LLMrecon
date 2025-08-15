#!/bin/bash

echo "Comprehensive syntax cleanup..."

# Remove all duplicate "if err != nil {" lines that were added by mistake
find src -name "*.go" -exec sed -i '' '/^if err != nil {$/d' {} \;

# Remove duplicate closing braces on consecutive lines
find src -name "*.go" -exec sed -i '' '/^}$/{N;s/^\}\n}$/}/;}' {} \;

# Fix struct definitions that are missing closing braces
find src -name "*.go" -exec sed -i '' '
/^type.*struct {/,/^$/ {
  /^$/i\
}
}' {} \; 2>/dev/null || true

# Clean up malformed function signatures
find src -name "*.go" -exec sed -i '' 's/^\}func /}\n\nfunc /g' {} \;
find src -name "*.go" -exec sed -i '' 's/^}\/\/ /}\n\n\/\/ /g' {} \;

# Fix specific known issues in config/config.go
sed -i '' '/^}$/!b; N; /^\}\n}$/{s/^\}\n}/}/;}' src/config/config.go

# Fix version package structs
for file in src/version/*.go; do
  # Ensure structs have closing braces
  sed -i '' '/^type.*struct {/,/^func\|^type\|^\/\// {
    /^func\|^type\|^\/\// i\
}
  }' "$file" 2>/dev/null || true
done

# Fix security/api structs
for file in src/security/api/*.go; do
  sed -i '' '/^type.*struct {/,/^func\|^type\|^\/\// {
    /^func\|^type\|^\/\// i\
}
  }' "$file" 2>/dev/null || true
done

# Fix template package structs  
for file in src/template/**/*.go; do
  if [ -f "$file" ]; then
    sed -i '' '/^type.*struct {/,/^func\|^type\|^\/\// {
      /^func\|^type\|^\/\// i\
}
    }' "$file" 2>/dev/null || true
  fi
done

# Fix provider/core structs
for file in src/provider/core/*.go; do
  sed -i '' '/^type.*struct {/,/^func\|^type\|^\/\// {
    /^func\|^type\|^\/\// i\
}
  }' "$file" 2>/dev/null || true
done

# Count remaining errors
echo "Remaining syntax errors:"
go build -o /tmp/test ./src/main.go 2>&1 | grep -c "syntax error" || echo "0"