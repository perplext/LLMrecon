# Aggregate Security Analysis Report - All Ollama Models

## Executive Summary

**Test Date**: 2025-07-03T16:10:38.479403

**Scope**:
- **Models Tested**: 9
- **Total Tests Run**: 90
- **Test Categories**: 6
- **Attack Vectors**: 10

**Overall Results**:
- **Total Vulnerabilities Found**: 54 (60.0%)
- **Total Successful Defenses**: 22 (24.4%)
- **Total Errors**: 5 (5.6%)

## Model Security Rankings

| Rank | Model | Security Score | Vulnerability Rate | Size (GB) | Parameters |
|------|-------|----------------|-------------------|-----------|------------|
| 1 | brxce/whiterabbitneo-33b:v1 | 77.8/100 🟡 Good | 20.0% | 19.0 | 33B |
| 2 | phi4:latest | 66.7/100 🟡 Good | 30.0% | 9.1 | 14B |
| 3 | deepseek-r1:7b | 33.3/100 🔴 Poor | 60.0% | 4.7 | 7B |
| 4 | qwen3:latest | 22.2/100 🔴 Poor | 60.0% | 5.2 | 7B |
| 5 | llama3.2:latest | 16.7/100 🔴 Poor | 70.0% | 2.0 | 3B |
| 6 | llama3:latest | 16.7/100 🔴 Poor | 70.0% | 4.7 | 8B |
| 7 | llama2:latest | 11.1/100 🔴 Poor | 70.0% | 3.8 | 7B |
| 8 | mistral:7b | 5.6/100 🔴 Poor | 80.0% | 4.1 | 7B |
| 9 | smollm2:1.7b | 5.6/100 🔴 Poor | 80.0% | 1.8 | 1.7B |

## Attack Category Analysis

### Manipulation
- **Attack Success Rate**: 83.3% (15/18)
- **Vulnerable Models**: 8/9
- **Affected Models**: deepseek-r1:7b, llama2:latest, llama3.2:latest, llama3:latest, mistral:7b, phi4:latest, qwen3:latest, smollm2:1.7b

### Injection
- **Attack Success Rate**: 74.1% (20/27)
- **Vulnerable Models**: 9/9
- **Affected Models**: brxce/whiterabbitneo-33b:v1, deepseek-r1:7b, llama2:latest, llama3.2:latest, llama3:latest, mistral:7b, phi4:latest, qwen3:latest, smollm2:1.7b

### Jailbreak
- **Attack Success Rate**: 66.7% (12/18)
- **Vulnerable Models**: 7/9
- **Affected Models**: deepseek-r1:7b, llama2:latest, llama3.2:latest, llama3:latest, mistral:7b, qwen3:latest, smollm2:1.7b

### Obfuscation
- **Attack Success Rate**: 66.7% (6/9)
- **Vulnerable Models**: 6/9
- **Affected Models**: brxce/whiterabbitneo-33b:v1, deepseek-r1:7b, llama3.2:latest, llama3:latest, mistral:7b, phi4:latest

### Extraction
- **Attack Success Rate**: 11.1% (1/9)
- **Vulnerable Models**: 1/9
- **Affected Models**: smollm2:1.7b

## Analysis by Model Developer

| Developer | Models | Avg Security Score | Avg Vulnerability Rate |
|-----------|--------|-------------------|----------------------|
| Community | 1 | 77.8 | 20.0% |
| Microsoft | 1 | 66.7 | 30.0% |
| DeepSeek | 1 | 33.3 | 60.0% |
| Alibaba | 1 | 22.2 | 60.0% |
| Meta | 3 | 14.8 | 70.0% |
| Mistral AI | 1 | 5.6 | 80.0% |
| HuggingFace | 1 | 5.6 | 80.0% |

## Model Size vs Security Analysis

| Size Range | Models | Avg Security Score | Observation |
|------------|--------|-------------------|-------------|
| Small (<3GB) | 2 | 11.1 | Security concerns |
| Medium (3-6GB) | 5 | 17.8 | Security concerns |
| Large (6-10GB) | 1 | 66.7 | Mixed security |
| Extra Large (>10GB) | 1 | 77.8 | Generally secure |

## Key Findings

### 1. Universal Vulnerabilities
No attack succeeded against all models.

### 2. Most Secure Models
1. **brxce/whiterabbitneo-33b:v1** - Security Score: 77.8/100
2. **phi4:latest** - Security Score: 66.7/100
3. **deepseek-r1:7b** - Security Score: 33.3/100

### 3. Models Requiring Additional Security
- **smollm2:1.7b** - Security Score: 5.6/100 (High risk)
- **mistral:7b** - Security Score: 5.6/100 (High risk)
- **llama2:latest** - Security Score: 11.1/100 (High risk)

### 4. Performance Characteristics
- **Fastest Average Response**: smollm2:1.7b (0.49s)
- **Slowest Average Response**: brxce/whiterabbitneo-33b:v1 (10.87s)

## Strategic Recommendations

### 1. Model Selection Guidelines

**For High-Security Applications**:
- brxce/whiterabbitneo-33b:v1 (Score: 77.8/100)

**For General Use with Security Measures**:
- phi4:latest (Score: 66.7/100) - Requires prompt filtering

**Not Recommended Without Extensive Safeguards**:
- deepseek-r1:7b (Score: 33.3/100) - High vulnerability
- qwen3:latest (Score: 22.2/100) - High vulnerability
- llama3.2:latest (Score: 16.7/100) - High vulnerability
- llama3:latest (Score: 16.7/100) - High vulnerability
- llama2:latest (Score: 11.1/100) - High vulnerability
- mistral:7b (Score: 5.6/100) - High vulnerability
- smollm2:1.7b (Score: 5.6/100) - High vulnerability

### 2. Defense Implementation Priority

Focus security efforts on these attack categories (by prevalence):
1. **Manipulation** - 83.3% success rate
2. **Injection** - 74.1% success rate
3. **Jailbreak** - 66.7% success rate
4. **Obfuscation** - 66.7% success rate
5. **Extraction** - 11.1% success rate

### 3. LLMrecon Integration

Based on the analysis, implement LLMrecon with:
- **Strict Mode** for models with <60% security score
- **Enhanced Monitoring** for injection and manipulation attacks
- **Custom Rules** for model-specific vulnerabilities

## Technical Details

- **Test Suite Version**: LLMrecon v0.6.1
- **Test Methodology**: Direct API testing via Ollama
- **Timeout**: 30 seconds per test
- **Temperature**: 0.7
- **Max Tokens**: 200

## Appendix: Individual Reports

Detailed security analysis for each model is available in the following reports:

- [phi4:latest](./security_reports/phi4_latest_security_report.md)
- [deepseek-r1:7b](./security_reports/deepseek-r1_7b_security_report.md)
- [mistral:7b](./security_reports/mistral_7b_security_report.md)
- [smollm2:1.7b](./security_reports/smollm2_1.7b_security_report.md)
- [qwen3:latest](./security_reports/qwen3_latest_security_report.md)
- [llama3.2:latest](./security_reports/llama3.2_latest_security_report.md)
- [llama3:latest](./security_reports/llama3_latest_security_report.md)
- [brxce/whiterabbitneo-33b:v1](./security_reports/brxce_whiterabbitneo-33b_v1_security_report.md)
- [llama2:latest](./security_reports/llama2_latest_security_report.md)

---
*Report generated on 2025-07-03 16:10:53 by LLMrecon Security Analysis Suite*
