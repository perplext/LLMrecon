#!/usr/bin/env python3

import re

def fix_all_function_closures():
    """Fix all missing function closing braces in bundle_report.go"""
    
    with open('src/cmd/bundle_report.go', 'r') as f:
        content = f.read()
    
    # List of functions that need closing braces based on the error messages
    functions_needing_braces = [
        ('loadBundleForReport', 'return b, nil'),
        ('generateReportData', 'return data, nil'),
        ('calculateSummary', 'return ReportSummary{'),
        ('generateOWASPAnalysis', 'return report'),
        ('generateISOAnalysis', 'return report'),
        ('generateStatistics', 'return stats'),
        ('generateVulnerabilityReport', 'return vulnerabilities'),
        ('generateCoverageReport', 'return report'),
        ('generateReportInFormat', 'return fmt.Errorf'),
        ('extractCategory', 'return part'),
    ]
    
    for func_name, return_pattern in functions_needing_braces:
        # Pattern to find the end of the function and add closing brace
        pattern = r'(\t' + re.escape(return_pattern) + r'[^}]*)\n\n(func \w+)'
        replacement = r'\1\n}\n\n\2'
        content = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    
    # Special cases for return statements that might be incomplete
    special_patterns = [
        (r'(\treturn b, nil)\n\n(func \w+)', r'\1\n}\n\n\2'),
        (r'(\treturn data, nil)\n\n(func \w+)', r'\1\n}\n\n\2'),
        (r'(\treturn report)\n\n(func \w+)', r'\1\n}\n\n\2'),
        (r'(\treturn stats)\n\n(func \w+)', r'\1\n}\n\n\2'),
        (r'(\treturn vulnerabilities)\n\n(func \w+)', r'\1\n}\n\n\2'),
        (r'(\treturn part)\n\n(func \w+)', r'\1\n}\n\n\2'),
        (r'(\treturn fmt\.Errorf[^}]*)\n\n(func \w+)', r'\1\n}\n\n\2'),
    ]
    
    for pattern, replacement in special_patterns:
        content = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    
    with open('src/cmd/bundle_report.go', 'w') as f:
        f.write(content)
    
    print("Applied final fixes to bundle_report.go")

if __name__ == "__main__":
    fix_all_function_closures()