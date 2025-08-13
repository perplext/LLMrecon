#!/usr/bin/env python3
"""
Generate aggregate security analysis report from individual model tests
"""

import json
from pathlib import Path
from datetime import datetime

def load_test_results():
    """Load the test results from JSON file"""
    results_file = Path("security_reports/all_models_test_results.json")
    with open(results_file, 'r') as f:
        return json.load(f)

def generate_aggregate_report(data):
    """Generate comprehensive aggregate security report"""
    
    models = data['models_tested']
    results = data['results']
    rankings = data['model_rankings']
    model_info = data['model_info']
    test_suite = data['test_suite']
    
    # Calculate aggregate statistics
    total_vulnerabilities = sum(r['vulnerabilities'] for r in results.values())
    total_defenses = sum(r['defenses'] for r in results.values())
    total_errors = sum(r['errors'] for r in results.values())
    total_tests = len(models) * len(test_suite)
    
    # Analyze by attack category
    category_stats = {}
    for test in test_suite:
        category = test['category']
        if category not in category_stats:
            category_stats[category] = {
                'total': 0,
                'vulnerable': 0,
                'models_vulnerable': []
            }
    
    # Count vulnerabilities by category
    for model, model_results in results.items():
        for test_result in model_results['tests']:
            test_info = next((t for t in test_suite if t['name'] == test_result['test']), None)
            if test_info and test_result['expected'] == 'vulnerable':
                category = test_info['category']
                category_stats[category]['total'] += 1
                if test_result['vulnerable']:
                    category_stats[category]['vulnerable'] += 1
                    if model not in category_stats[category]['models_vulnerable']:
                        category_stats[category]['models_vulnerable'].append(model)
    
    # Generate report content
    content = f"""# Aggregate Security Analysis Report - All Ollama Models

## Executive Summary

**Test Date**: {data['test_date']}

**Scope**:
- **Models Tested**: {len(models)}
- **Total Tests Run**: {total_tests}
- **Test Categories**: {len(set(t['category'] for t in test_suite))}
- **Attack Vectors**: {len(test_suite)}

**Overall Results**:
- **Total Vulnerabilities Found**: {total_vulnerabilities} ({(total_vulnerabilities/total_tests*100):.1f}%)
- **Total Successful Defenses**: {total_defenses} ({(total_defenses/total_tests*100):.1f}%)
- **Total Errors**: {total_errors} ({(total_errors/total_tests*100):.1f}%)

## Model Security Rankings

| Rank | Model | Security Score | Vulnerability Rate | Size (GB) | Parameters |
|------|-------|----------------|-------------------|-----------|------------|
"""
    
    # Add rankings table
    for rank, (model, score, vuln_rate) in enumerate(rankings, 1):
        info = model_info.get(model, {})
        size = info.get('size', 'Unknown')
        params = info.get('params', 'Unknown')
        
        # Determine rating
        if score >= 80:
            rating = "🟢 Excellent"
        elif score >= 60:
            rating = "🟡 Good"
        elif score >= 40:
            rating = "🟠 Moderate"
        else:
            rating = "🔴 Poor"
        
        content += f"| {rank} | {model} | {score:.1f}/100 {rating} | {vuln_rate:.1f}% | {size} | {params} |\n"
    
    # Attack category analysis
    content += f"\n## Attack Category Analysis\n\n"
    
    for category, stats in sorted(category_stats.items(), 
                                 key=lambda x: x[1]['vulnerable']/x[1]['total'] if x[1]['total'] > 0 else 0, 
                                 reverse=True):
        if stats['total'] > 0:
            success_rate = (stats['vulnerable'] / stats['total']) * 100
            content += f"### {category.title()}\n"
            content += f"- **Attack Success Rate**: {success_rate:.1f}% ({stats['vulnerable']}/{stats['total']})\n"
            content += f"- **Vulnerable Models**: {len(stats['models_vulnerable'])}/{len(models)}\n"
            if stats['models_vulnerable']:
                content += f"- **Affected Models**: {', '.join(sorted(stats['models_vulnerable']))}\n"
            content += "\n"
    
    # Company/Developer Analysis
    content += "## Analysis by Model Developer\n\n"
    
    company_stats = {}
    for model in models:
        info = model_info.get(model, {})
        company = info.get('company', 'Unknown')
        
        if company not in company_stats:
            company_stats[company] = {
                'models': [],
                'scores': [],
                'vuln_rates': []
            }
        
        # Find model score and vuln rate
        model_data = next((m for m in rankings if m[0] == model), None)
        if model_data:
            company_stats[company]['models'].append(model)
            company_stats[company]['scores'].append(model_data[1])
            company_stats[company]['vuln_rates'].append(model_data[2])
    
    content += "| Developer | Models | Avg Security Score | Avg Vulnerability Rate |\n"
    content += "|-----------|--------|-------------------|----------------------|\n"
    
    for company, stats in sorted(company_stats.items(), 
                                key=lambda x: sum(x[1]['scores'])/len(x[1]['scores']) if x[1]['scores'] else 0,
                                reverse=True):
        if stats['scores']:
            avg_score = sum(stats['scores']) / len(stats['scores'])
            avg_vuln = sum(stats['vuln_rates']) / len(stats['vuln_rates'])
            content += f"| {company} | {len(stats['models'])} | {avg_score:.1f} | {avg_vuln:.1f}% |\n"
    
    # Size vs Security Analysis
    content += "\n## Model Size vs Security Analysis\n\n"
    content += "| Size Range | Models | Avg Security Score | Observation |\n"
    content += "|------------|--------|-------------------|-------------|\n"
    
    size_ranges = [
        ("Small (<3GB)", 0, 3),
        ("Medium (3-6GB)", 3, 6),
        ("Large (6-10GB)", 6, 10),
        ("Extra Large (>10GB)", 10, 1000)
    ]
    
    for range_name, min_size, max_size in size_ranges:
        range_models = []
        range_scores = []
        
        for model, score, vuln_rate in rankings:
            info = model_info.get(model, {})
            size = info.get('size', 0)
            if min_size <= size < max_size:
                range_models.append(model)
                range_scores.append(score)
        
        if range_scores:
            avg_score = sum(range_scores) / len(range_scores)
            if avg_score >= 70:
                observation = "Generally secure"
            elif avg_score >= 50:
                observation = "Mixed security"
            else:
                observation = "Security concerns"
            
            content += f"| {range_name} | {len(range_models)} | {avg_score:.1f} | {observation} |\n"
    
    # Key Findings
    content += "\n## Key Findings\n\n"
    
    # Find universal vulnerabilities
    universal_vulns = []
    for test in test_suite:
        if test['expected'] == 'vulnerable':
            vulnerable_count = sum(1 for model_results in results.values()
                                 for test_result in model_results['tests']
                                 if test_result['test'] == test['name'] and test_result['vulnerable'])
            if vulnerable_count == len(models):
                universal_vulns.append(test['name'])
    
    content += "### 1. Universal Vulnerabilities\n"
    if universal_vulns:
        content += f"The following attacks succeeded against ALL tested models:\n"
        for vuln in universal_vulns:
            content += f"- {vuln}\n"
    else:
        content += "No attack succeeded against all models.\n"
    
    # Find most resistant models
    content += "\n### 2. Most Secure Models\n"
    top_models = rankings[:3]
    for rank, (model, score, vuln_rate) in enumerate(top_models, 1):
        content += f"{rank}. **{model}** - Security Score: {score:.1f}/100\n"
    
    # Find most vulnerable models
    content += "\n### 3. Models Requiring Additional Security\n"
    bottom_models = rankings[-3:]
    for model, score, vuln_rate in reversed(bottom_models):
        if score < 60:
            content += f"- **{model}** - Security Score: {score:.1f}/100 (High risk)\n"
    
    # Response time analysis
    content += "\n### 4. Performance Characteristics\n"
    
    model_perf = {}
    for model, model_results in results.items():
        response_times = [t['time'] for t in model_results['tests'] if t['success']]
        if response_times:
            model_perf[model] = {
                'avg': sum(response_times) / len(response_times),
                'min': min(response_times),
                'max': max(response_times)
            }
    
    # Find fastest and slowest
    if model_perf:
        fastest = min(model_perf.items(), key=lambda x: x[1]['avg'])
        slowest = max(model_perf.items(), key=lambda x: x[1]['avg'])
        
        content += f"- **Fastest Average Response**: {fastest[0]} ({fastest[1]['avg']:.2f}s)\n"
        content += f"- **Slowest Average Response**: {slowest[0]} ({slowest[1]['avg']:.2f}s)\n"
    
    # Recommendations
    content += "\n## Strategic Recommendations\n\n"
    
    content += "### 1. Model Selection Guidelines\n\n"
    content += "**For High-Security Applications**:\n"
    for model, score, _ in rankings[:3]:
        if score >= 70:
            content += f"- {model} (Score: {score:.1f}/100)\n"
    
    content += "\n**For General Use with Security Measures**:\n"
    for model, score, _ in rankings:
        if 50 <= score < 70:
            content += f"- {model} (Score: {score:.1f}/100) - Requires prompt filtering\n"
    
    content += "\n**Not Recommended Without Extensive Safeguards**:\n"
    for model, score, _ in rankings:
        if score < 50:
            content += f"- {model} (Score: {score:.1f}/100) - High vulnerability\n"
    
    content += "\n### 2. Defense Implementation Priority\n\n"
    
    # Sort categories by success rate
    sorted_categories = sorted(category_stats.items(), 
                             key=lambda x: x[1]['vulnerable']/x[1]['total'] if x[1]['total'] > 0 else 0,
                             reverse=True)
    
    content += "Focus security efforts on these attack categories (by prevalence):\n"
    for i, (category, stats) in enumerate(sorted_categories[:5], 1):
        if stats['total'] > 0:
            success_rate = (stats['vulnerable'] / stats['total']) * 100
            content += f"{i}. **{category.title()}** - {success_rate:.1f}% success rate\n"
    
    content += "\n### 3. LLMrecon Integration\n\n"
    content += "Based on the analysis, implement LLMrecon with:\n"
    content += "- **Strict Mode** for models with <60% security score\n"
    content += "- **Enhanced Monitoring** for injection and manipulation attacks\n"
    content += "- **Custom Rules** for model-specific vulnerabilities\n"
    
    # Technical details
    content += "\n## Technical Details\n\n"
    content += f"- **Test Suite Version**: LLMrecon v0.6.1\n"
    content += f"- **Test Methodology**: Direct API testing via Ollama\n"
    content += f"- **Timeout**: 30 seconds per test\n"
    content += f"- **Temperature**: 0.7\n"
    content += f"- **Max Tokens**: 200\n"
    
    # Appendix
    content += "\n## Appendix: Individual Reports\n\n"
    content += "Detailed security analysis for each model is available in the following reports:\n\n"
    
    for model in models:
        report_name = f"{model.replace(':', '_').replace('/', '_')}_security_report.md"
        content += f"- [{model}](./security_reports/{report_name})\n"
    
    content += f"\n---\n"
    content += f"*Report generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')} by LLMrecon Security Analysis Suite*\n"
    
    # Save the report
    report_path = Path("MODEL_SECURITY_ANALYSIS_AGGREGATE.md")
    with open(report_path, 'w') as f:
        f.write(content)
    
    return report_path

def main():
    """Main function"""
    print("Loading test results...")
    data = load_test_results()
    
    print("Generating aggregate report...")
    report_path = generate_aggregate_report(data)
    
    print(f"\nAggregate report saved to: {report_path}")
    print("\nReport generation complete!")

if __name__ == "__main__":
    main()