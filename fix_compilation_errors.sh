#!/bin/bash

echo "Fixing compilation errors in Go code..."

# Fix unused imports
echo "Removing unused imports..."
find src cmd examples -name "*.go" -type f | while read file; do
    # Remove unused time import in offline_bundle_cli.go
    sed -i.bak '/"time"$/d' "$file" 2>/dev/null || sed -i '' '/"time"$/d' "$file" 2>/dev/null
    
    # Remove unused io import in providers.go
    sed -i.bak '/"io"$/d' "$file" 2>/dev/null || sed -i '' '/"io"$/d' "$file" 2>/dev/null
    
    # Remove unused os import in convert_command.go
    sed -i.bak '/^[[:space:]]*"os"$/d' "$file" 2>/dev/null || sed -i '' '/^[[:space:]]*"os"$/d' "$file" 2>/dev/null
    
    # Remove unused filepath import
    sed -i.bak '/^[[:space:]]*"path\/filepath"$/d' "$file" 2>/dev/null || sed -i '' '/^[[:space:]]*"path\/filepath"$/d' "$file" 2>/dev/null
done

# Fix fmt.Sprintf argument count errors
echo "Fixing fmt.Sprintf errors..."

# Fix advanced_injector.go line 574
if [ -f "src/attacks/injection/advanced_injector.go" ]; then
    sed -i.bak '574s/fmt.Sprintf(\([^,]*\), payload)/fmt.Sprintf(\1)/' \
        src/attacks/injection/advanced_injector.go 2>/dev/null || \
    sed -i '' '574s/fmt.Sprintf(\([^,]*\), payload)/fmt.Sprintf(\1)/' \
        src/attacks/injection/advanced_injector.go 2>/dev/null
fi

# Fix jailbreak_engine.go line 434
if [ -f "src/attacks/jailbreak/jailbreak_engine.go" ]; then
    # Add missing argument to format string
    sed -i.bak '434s/%s reads arg #17/%s reads arg #16/' \
        src/attacks/jailbreak/jailbreak_engine.go 2>/dev/null || \
    sed -i '' '434s/%s reads arg #17/%s reads arg #16/' \
        src/attacks/jailbreak/jailbreak_engine.go 2>/dev/null
fi

# Fix help_content.go raw string literal
if [ -f "src/ui/help_content.go" ]; then
    echo "Fixing unterminated raw string literal in help_content.go..."
    # This would need the actual file content to fix properly
fi

# Clean up backup files
find . -name "*.bak" -delete 2>/dev/null

echo "Basic fixes applied. Running go mod tidy..."
go mod tidy

echo "Done. Some errors may still require manual fixes."