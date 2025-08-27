#!/usr/bin/env python3

import re

def fix_bundle_create():
    """Fix missing closing braces in bundle_create.go"""
    
    with open('src/cmd/bundle_create.go', 'r') as f:
        content = f.read()
    
    # Find all the function signatures that need closing braces
    functions_needing_braces = [
        'createBundleOptions',
        'createBundleExporter', 
        'createProgressHandler',
        'fetchFromGitHub',
        'fetchFromGitLab',
        'signBundle',
        'calculateBundleChecksum',
        'formatSize'
    ]
    
    # Pattern to find return statements followed by function declarations
    pattern = r'(\treturn [^}]+\n)\n(func (' + '|'.join(functions_needing_braces) + r')\b)'
    
    def replacement(match):
        return match.group(1) + '}\n\n' + match.group(2)
    
    content = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    
    with open('src/cmd/bundle_create.go', 'w') as f:
        f.write(content)
    
    print("Fixed missing braces in bundle_create.go")

if __name__ == "__main__":
    fix_bundle_create()