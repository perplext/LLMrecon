#!/usr/bin/env python3

import re

def fix_bundle_report():
    """Fix missing closing braces in bundle_report.go"""
    
    with open('src/cmd/bundle_report.go', 'r') as f:
        content = f.read()
    
    # Split into lines for easier processing
    lines = content.split('\n')
    result = []
    
    i = 0
    while i < len(lines):
        line = lines[i]
        result.append(line)
        
        # Check if this line is a return statement and next line starts a function
        if (line.strip().startswith('return ') and 
            i + 1 < len(lines) and 
            lines[i + 1].strip() == '' and
            i + 2 < len(lines) and
            lines[i + 2].startswith('func ')):
            
            # Add missing closing brace
            result.append('}')
        
        # Check for specific patterns where return is immediately followed by func
        elif (line.strip().startswith('return ') and 
              i + 1 < len(lines) and
              lines[i + 1].startswith('func ')):
            result.append('}')
            result.append('')  # Add blank line
        
        i += 1
    
    # Join back and write
    fixed_content = '\n'.join(result)
    
    with open('src/cmd/bundle_report.go', 'w') as f:
        f.write(fixed_content)
    
    print("Fixed bundle_report.go function closures")

if __name__ == "__main__":
    fix_bundle_report()