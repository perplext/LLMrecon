#!/usr/bin/env python3

import os
import re

def fix_go_syntax(filepath):
    """Comprehensive syntax fixes for Go files"""
    
    with open(filepath, 'r') as f:
        content = f.read()
    
    original_content = content
    
    # Fix 1: Missing closing braces in struct declarations
    # Pattern: field declaration followed by type declaration
    content = re.sub(r'(\t\w+\s+[^`]+`[^`]*`)\n\n(type \w+)', r'\1\n}\n\n\2', content)
    
    # Fix 2: Missing closing braces for functions
    # Pattern: return statement followed by function declaration
    content = re.sub(r'(\treturn [^}]+)\n\n(func \w+)', r'\1\n}\n\n\2', content)
    
    # Fix 3: Missing closing braces for functions with different return patterns
    content = re.sub(r'(\treturn rootCmd\.Execute\(\))\n(func \w+)', r'\1\n}\n\n\2', content)
    content = re.sub(r'(\treturn nil)\n(func \w+)', r'\1\n}\n\n\2', content)
    
    # Fix 4: Missing closing braces for switch/case blocks
    content = re.sub(r'(\treturn "[^"]+"\n\t}\n)\n(func \w+)', r'\1}\n\n\2', content)
    
    # Fix 5: Fix cobra.Command structs
    content = re.sub(r'(RunE:\s*\w+,)\n\n(func \w+)', r'\1\n}\n\n\2', content)
    
    # Fix 6: Fix slice declarations
    content = re.sub(r'(\]\s*\n\n)(func \w+)', r'}\n\n\2', content)
    
    # Fix 7: Remove trailing extra braces
    content = re.sub(r'\n\}\s*\n$', '\n', content)
    
    # Fix 8: Ensure file ends with newline but not excessive braces
    if not content.endswith('\n'):
        content += '\n'
    
    if content != original_content:
        with open(filepath, 'w') as f:
            f.write(content)
        return True
    return False

def main():
    """Fix all Go files in src/cmd directory"""
    
    cmd_dir = "src/cmd"
    fixed_files = []
    
    for filename in os.listdir(cmd_dir):
        if filename.endswith('.go'):
            filepath = os.path.join(cmd_dir, filename)
            if fix_go_syntax(filepath):
                fixed_files.append(filename)
    
    if fixed_files:
        print(f"Fixed syntax in {len(fixed_files)} files:")
        for f in fixed_files:
            print(f"  - {f}")
    else:
        print("No syntax fixes needed")

if __name__ == "__main__":
    main()