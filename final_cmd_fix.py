#!/usr/bin/env python3

import os
import re

def add_missing_braces(filepath):
    """Add missing closing braces to all functions in a Go file"""
    
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Split into lines
    lines = content.split('\n')
    result = []
    
    for i, line in enumerate(lines):
        result.append(line)
        
        # If this line is a return statement and next non-empty line starts with "func"
        if (line.strip().startswith('return ') and 
            i + 1 < len(lines)):
            
            # Look ahead for next non-empty line
            j = i + 1
            while j < len(lines) and lines[j].strip() == '':
                j += 1
            
            if j < len(lines) and lines[j].startswith('func '):
                # Add closing brace before the function
                result.append('}')
    
    # Join and write back
    fixed_content = '\n'.join(result)
    
    # Remove any duplicate closing braces at the end
    fixed_content = re.sub(r'\n}\s*\n}\s*$', '\n}\n', fixed_content)
    
    with open(filepath, 'w') as f:
        f.write(fixed_content)

def main():
    """Fix all Go files in src/cmd"""
    
    cmd_dir = 'src/cmd'
    for filename in os.listdir(cmd_dir):
        if filename.endswith('.go'):
            filepath = os.path.join(cmd_dir, filename)
            add_missing_braces(filepath)
    
    print("Applied final comprehensive fixes to all cmd files")

if __name__ == "__main__":
    main()