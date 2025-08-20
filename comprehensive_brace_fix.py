#!/usr/bin/env python3

import re

def fix_bundle_compliance():
    """Comprehensive fix for bundle_compliance.go missing braces"""
    
    with open('src/cmd/bundle_compliance.go', 'r') as f:
        content = f.read()
    
    # Define specific line patterns where closing braces are missing
    fixes = [
        # Pattern: return statement followed by function declaration without closing brace
        (r'(\treturn ComplianceStatus\{[^}]+}\)\n\n)(func \w+)', r'\1}\n\n\2'),
        (r'(\treturn [^{}]+\n)\n(func \w+)', r'\1}\n\n\2'),
        
        # Handle specific cases
        (r'(\t}\n)\n(func checkNISTCompliance)', r'\1}\n\n\2'),
        (r'(\t}\n)\n(func checkEUAICompliance)', r'\1}\n\n\2'),
        (r'(\t}\n)\n(func displayComplianceStatus)', r'\1}\n\n\2'),
        (r'(\t}\n)\n(func calculateOverallCompliance)', r'\1}\n\n\2'),
        (r'(\t}\n)\n(func saveComplianceDocument)', r'\1}\n\n\2'),
        (r'(\t}\n)\n(func saveMarkdownDocument)', r'\1}\n\n\2'),
        (r'(\t}\n)\n(func saveHTMLDocument)', r'\1}\n\n\2'),
    ]
    
    for pattern, replacement in fixes:
        content = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    
    # Write back
    with open('src/cmd/bundle_compliance.go', 'w') as f:
        f.write(content)
    
    print("Applied comprehensive brace fixes")

if __name__ == "__main__":
    fix_bundle_compliance()