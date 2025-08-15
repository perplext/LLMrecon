#!/bin/bash

# Final cleanup for import blocks that have improper spacing/formatting

echo "Final import cleanup..."

# Find remaining files with import syntax errors
go build ./src/... 2>&1 | grep "expected ';', found" | cut -d':' -f1 | sort -u | while read -r file; do
    if [[ -f "$file" ]]; then
        echo "Cleaning imports in $file"
        
        # Read file and fix import block formatting
        python3 -c "
import re
import sys

filename = '$file'
with open(filename, 'r') as f:
    content = f.read()

# Pattern to match import blocks with issues
# Look for import ( followed by potential malformed lines
import_pattern = r'import \((.*?)\)'

def fix_import_block(match):
    import_content = match.group(1).strip()
    lines = import_content.split('\n')
    
    fixed_lines = []
    for line in lines:
        line = line.strip()
        if not line:
            continue
            
        # Remove extra tabs/spaces and fix formatting
        line = re.sub(r'^\s*', '\t', line)
        
        # Handle cases where two imports are on same line
        # Look for patterns like: alias \"pkg1\" \"pkg2\"
        if line.count('\"') > 2:
            # Split multiple imports on same line
            parts = line.split('\"')
            current_import = ''
            for i, part in enumerate(parts):
                if i == 0:
                    current_import = part
                elif i % 2 == 1:  # Odd indices are package names
                    current_import += '\"' + part + '\"'
                    if i < len(parts) - 1 and parts[i+1].strip():
                        fixed_lines.append(current_import)
                        current_import = '\t'
                    elif i == len(parts) - 1:
                        fixed_lines.append(current_import)
                else:  # Even indices after 0 are separators
                    if part.strip() and i < len(parts) - 1:
                        fixed_lines.append(current_import)
                        current_import = '\t'
        else:
            fixed_lines.append(line)
    
    return 'import (\n' + '\n'.join(fixed_lines) + '\n)'

# Apply the fix
content = re.sub(import_pattern, fix_import_block, content, flags=re.DOTALL)

with open(filename, 'w') as f:
    f.write(content)
"
    fi
done

echo "Final import cleanup completed."