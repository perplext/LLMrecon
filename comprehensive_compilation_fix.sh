#!/bin/bash

echo "Comprehensive compilation fix for remaining Go packages..."

# Get list of files with compilation errors
error_files=$(go build ./src/... 2>&1 | grep -E "syntax error" | cut -d':' -f1 | sort -u)

echo "Found $(echo "$error_files" | wc -l) files with compilation errors"

for file in $error_files; do
    if [[ -f "$file" ]]; then
        echo "Fixing: $file"
        
        # Python script to fix common syntax issues
        python3 -c "
import re
import sys

filename = '$file'
try:
    with open(filename, 'r') as f:
        lines = f.readlines()
    
    fixed_lines = []
    in_struct = False
    in_interface = False
    in_func = False
    brace_count = 0
    struct_start = -1
    interface_start = -1
    func_start = -1
    
    for i, line in enumerate(lines):
        stripped = line.strip()
        
        # Track struct definitions
        if re.match(r'^type\s+\w+\s+struct\s*{', stripped):
            in_struct = True
            struct_start = i
            brace_count = 1
        elif in_struct:
            brace_count += stripped.count('{') - stripped.count('}')
            if brace_count == 0:
                in_struct = False
        elif re.match(r'^type\s+\w+\s+struct\s*$', stripped) and i < len(lines) - 1 and lines[i+1].strip() == '{':
            in_struct = True
            struct_start = i
            brace_count = 0
        
        # Track interface definitions
        if re.match(r'^type\s+\w+\s+interface\s*{', stripped):
            in_interface = True
            interface_start = i
            brace_count = 1
        elif in_interface:
            brace_count += stripped.count('{') - stripped.count('}')
            if brace_count == 0:
                in_interface = False
        
        # Track function definitions
        if re.match(r'^func\s+', stripped) and '{' in stripped:
            in_func = True
            func_start = i
            brace_count = stripped.count('{') - stripped.count('}')
        elif in_func:
            brace_count += stripped.count('{') - stripped.count('}')
            if brace_count == 0:
                in_func = False
        
        # Check for missing closing braces
        if i < len(lines) - 1:
            next_line = lines[i+1].strip()
            
            # If we're in a struct and the next line starts a new type/func, add closing brace
            if in_struct and (next_line.startswith('type ') or next_line.startswith('func ') or 
                             next_line.startswith('//') and i+2 < len(lines) and 
                             (lines[i+2].strip().startswith('type ') or lines[i+2].strip().startswith('func '))):
                if not stripped.endswith('}'):
                    line = line.rstrip() + '\n}\n'
                    in_struct = False
                    brace_count = 0
            
            # If we're in an interface and the next line starts a new type/func, add closing brace
            if in_interface and (next_line.startswith('type ') or next_line.startswith('func ')):
                if not stripped.endswith('}'):
                    line = line.rstrip() + '\n}\n'
                    in_interface = False
                    brace_count = 0
            
            # If we're in a function and the next line starts a new func, add closing brace
            if in_func and next_line.startswith('func '):
                if not stripped.endswith('}'):
                    line = line.rstrip() + '\n}\n'
                    in_func = False
                    brace_count = 0
        
        fixed_lines.append(line)
    
    # Write fixed content
    with open(filename, 'w') as f:
        f.writelines(fixed_lines)
    
    print(f'  Fixed {filename}')
    
except Exception as e:
    print(f'  Error fixing {filename}: {e}')
"
    fi
done

echo "Compilation fix completed. Testing build..."
go build ./src/... 2>&1 | grep -c "^#" | xargs echo "Remaining packages with errors:"
