#!/usr/bin/env python3
"""
Quick test of LLMrecon 2025 features with gpt-oss model
"""

import subprocess
import sys
import json
from pathlib import Path

def run_test(categories=None):
    """Run a quick test with specified categories"""
    
    cmd = [
        sys.executable, 
        "llmrecon_2025.py",
        "--models", "gpt-oss:latest"
    ]
    
    if categories:
        cmd.extend(["--categories"] + categories)
    
    print(f"Running: {' '.join(cmd)}")
    print("=" * 60)
    
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)
    
    # Show output
    if result.stdout:
        print(result.stdout)
    if result.stderr:
        print("STDERR:", result.stderr)
    
    # Check if report was created
    report_file = Path("llmrecon_2025_report.json")
    if report_file.exists():
        with open(report_file, 'r') as f:
            report = json.load(f)
        
        print("\n" + "=" * 60)
        print("REPORT SUMMARY")
        print("=" * 60)
        
        summary = report.get("summary", {})
        print(f"Total Tests: {summary.get('total_tests', 0)}")
        print(f"Vulnerable: {summary.get('vulnerable_tests', 0)}")
        
        print("\nOWASP 2025 Compliance:")
        for owasp_id, stats in summary.get("owasp_2025_compliance", {}).items():
            if stats["tested"] > 0:
                rate = stats["vulnerable"] / stats["tested"]
                print(f"  {owasp_id}: {stats['tested']} tests, {stats['vulnerable']} vulnerable ({rate:.0%})")
        
        print("\nDefense Mechanisms Detected:")
        for defense, count in summary.get("defense_mechanisms", {}).items():
            print(f"  {defense}: {count}")
        
        print("\nRecommendations:")
        for rec in report.get("recommendations", []):
            print(f"  • {rec}")
    
    return result.returncode

if __name__ == "__main__":
    print("Testing LLMrecon 2025 with Novel Attack Techniques")
    print("=" * 60)
    
    # Test just the novel 2024-2025 attack techniques
    print("\n1. Testing Novel Prompt Injection Attacks (FlipAttack, DrAttack, Policy Puppetry)")
    try:
        run_test(["prompt_injection"])
    except subprocess.TimeoutExpired:
        print("Test timed out after 120 seconds")
    
    print("\n" + "=" * 60)
    print("Test complete. Check llmrecon_2025_report.json for detailed results.")