#!/usr/bin/env python3

import re

def fix_missing_braces(filename):
    with open(filename, 'r') as f:
        content = f.read()
    
    lines = content.split('\n')
    fixed_lines = []
    
    i = 0
    while i < len(lines):
        line = lines[i].rstrip()
        fixed_lines.append(line)
        
        # Check if this line starts a function definition
        if line.strip().startswith('func '):
            # Look ahead to find the end of this function
            brace_count = 0
            j = i + 1
            found_opening = False
            
            # Find the opening brace first
            while j < len(lines) and not found_opening:
                if '{' in lines[j]:
                    found_opening = True
                    break
                j += 1
            
            if found_opening:
                # Count braces from the opening
                j = i + 1
                while j < len(lines):
                    for char in lines[j]:
                        if char == '{':
                            brace_count += 1
                        elif char == '}':
                            brace_count -= 1
                    
                    # If we reached 0, we found the end
                    if brace_count == 0:
                        break
                    
                    # If we hit another function definition before closing, we have a missing brace
                    if j + 1 < len(lines) and lines[j + 1].strip().startswith('func ') and brace_count > 0:
                        # Add missing closing brace
                        fixed_lines.append('}')
                        break
                    
                    j += 1
        
        i += 1
    
    # Write back the fixed content
    with open(filename, 'w') as f:
        f.write('\n'.join(fixed_lines))

if __name__ == "__main__":
    fix_missing_braces('/Users/nconsolo/claude-code/llmrecon/src/automated/learning/adaptive_system.go')
    print("Fixed missing braces")