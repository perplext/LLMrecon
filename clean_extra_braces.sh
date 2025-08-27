#!/bin/bash

echo "Cleaning up extra closing braces from Go files..."

# Remove trailing lines with only closing braces
for file in src/analytics/*.go src/api/*.go src/provider/core/*.go; do
    if [ -f "$file" ]; then
        echo "Cleaning $file..."
        # Remove lines at the end of file that only contain }
        sed -i.bak '/^}$/d' "$file"
        # Add back one closing brace at the end if needed
        echo "}" >> "$file"
    fi
done

# Remove backup files
find src -name "*.bak" -delete

echo "Cleanup complete."