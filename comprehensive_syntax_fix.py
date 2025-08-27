#!/usr/bin/env python3

import os
import re
import sys

def fix_go_file(filepath):
    """Fix common syntax issues in Go files"""
    
    with open(filepath, 'r') as f:
        content = f.read()
    
    original_content = content
    
    # Pattern 1: return statement followed by function declaration without closing brace
    pattern1 = r'(\treturn [^{};]+;?\s*)\n\n(func \w+)'
    content = re.sub(pattern1, r'\1\n}\n\n\2', content, flags=re.MULTILINE)
    
    # Pattern 2: struct declaration followed by function without closing brace
    pattern2 = r'(\})\n\n(func \w+)'
    # Only add brace if the previous line doesn't already end with }
    content = re.sub(r'([^}]\s*)\n\n(func \w+)', r'\1\n}\n\n\2', content, flags=re.MULTILINE)
    
    # Pattern 3: Fix unclosed slice/struct declarations
    pattern3 = r'(\w+\s*=\s*\[\]string\{[^}]+)\n\n(/\*|//|\w|func)'
    content = re.sub(pattern3, r'\1}\n\n\2', content, flags=re.MULTILINE)
    
    # Pattern 4: Fix unclosed cobra.Command structs
    pattern4 = r'(RunE:\s*\w+,)\s*\n\n(func \w+)'
    content = re.sub(pattern4, r'\1\n}\n\n\2', content, flags=re.MULTILINE)
    
    # Pattern 5: Fix extra closing braces at end of file
    content = re.sub(r'\n(\}\s*)+$', '\n}', content)
    
    # Pattern 6: Fix missing return statements in functions that should return values
    # Look for functions that end with } followed by another function, but should have a return
    pattern6 = r'(func \w+\([^)]*\)\s+[^{]+\{[^}]*)\n\}\n\n(func \w+)'
    
    if content != original_content:
        with open(filepath, 'w') as f:
            f.write(content)
        return True
    return False

def fix_directory(directory):
    """Fix all Go files in a directory"""
    fixed_files = []
    
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                if fix_go_file(filepath):
                    fixed_files.append(filepath)
    
    return fixed_files

if __name__ == "__main__":
    if len(sys.argv) > 1:
        target = sys.argv[1]
    else:
        target = "src/cmd"
    
    if os.path.isfile(target):
        if fix_go_file(target):
            print(f"Fixed {target}")
        else:
            print(f"No changes needed in {target}")
    else:
        fixed = fix_directory(target)
        if fixed:
            print(f"Fixed {len(fixed)} files:")
            for f in fixed:
                print(f"  - {f}")
        else:
            print("No files needed fixing")