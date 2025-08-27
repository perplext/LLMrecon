#!/usr/bin/env python3

import re
import sys

def fix_missing_braces(filepath):
    """Fix missing closing braces in Go files"""
    
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Pattern to find return statements followed directly by function declarations
    # This indicates a missing closing brace
    fixes = [
        (r'(\treturn [^}]+\n)\n(func \w+)', r'\1}\n\n\2'),
        (r'(\treturn [^}]+\n)\n(// [^\n]+\n\nfunc \w+)', r'\1}\n\n\2'),
        (r'(\treturn [^}]+\n)\n(// [^\n]+\n\n// [^\n]+\n\nfunc \w+)', r'\1}\n\n\2'),
    ]
    
    original_content = content
    
    for pattern, replacement in fixes:
        content = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    
    # Additional specific fixes for malformed returns
    specific_fixes = [
        # Fix malformed os.WriteFile calls
        (r'return os\.WriteFile\(filepath\.Clean\(filename, \[\]byte\(content\.String\(\)\), 0600\)', 
         'return os.WriteFile(filepath.Clean(filename), []byte(content.String()), 0600)'),
        (r'return os\.WriteFile\(filepath\.Clean\(filename, data, 0600\)', 
         'return os.WriteFile(filepath.Clean(filename), data, 0600)'),
        (r'return os\.WriteFile\(filepath\.Clean\(filepath\.Join\([^)]+\), \[\]byte\([^)]+\), 0600\)',
         lambda m: m.group(0).replace('filepath.Clean(filepath.Join(', 'os.WriteFile(filepath.Clean(filepath.Join(')),
    ]
    
    for pattern, replacement in specific_fixes:
        if callable(replacement):
            content = re.sub(pattern, replacement, content)
        else:
            content = re.sub(pattern, replacement, content)
    
    if content != original_content:
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"Fixed missing braces in {filepath}")
        return True
    else:
        print(f"No fixes needed in {filepath}")
        return False

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 fix_closing_braces.py <filepath>")
        sys.exit(1)
    
    filepath = sys.argv[1]
    fix_missing_braces(filepath)