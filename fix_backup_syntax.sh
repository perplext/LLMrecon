#!/bin/bash

# Fix syntax errors in backup version files
echo "Fixing syntax errors in version backup files..."

# Function to fix missing closing braces for structs and functions
fix_missing_braces() {
    local file="$1"
    echo "Fixing $file..."
    
    # Common patterns to fix:
    # 1. Missing } after struct definitions before next type/func
    # 2. Missing } after function definitions before next func
    # 3. Missing } at end of functions
    
    # Use sed to add missing closing braces in common patterns
    sed -i '' \
        -e '/^type.*struct {$/,/^$/{
            /^$/i\
}
        }' \
        -e '/^func.*{$/,/^func/{
            /^func.*{$/{
                x
                s/.*/INFUNC/
                x
                b
            }
            /^func/{
                x
                /INFUNC/{
                    x
                    i\
}
                    b
                }
                x
            }
        }' \
        "$file"
}

# Fix each file
for file in src/version/backup/*.go; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        # Make a backup
        cp "$file" "${file}.bak"
        
        # Try to fix common patterns
        fix_missing_braces "$file"
        
        echo "Fixed $file"
    fi
done

echo "Completed fixing backup files"