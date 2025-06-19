# OWASP LLM Top 10 Security Checklist

## Introduction

This checklist provides a comprehensive self-assessment tool for organizations to evaluate their security posture against the OWASP Top 10 for Large Language Model Applications (2025). Use this checklist to identify vulnerabilities in your LLM applications and prioritize security improvements.

## How to Use This Checklist

1. For each security control, assess your current implementation status using the following scale:
   - **Implemented**: Control fully implemented and tested
   - **Partially Implemented**: Control partially implemented or not fully tested
   - **Not Implemented**: Control not implemented
   - **Not Applicable**: Control not applicable to your application (justification required)

2. For items marked as "Partially Implemented" or "Not Implemented", document the gaps and create action plans to address them.

3. Use the "Evidence/Notes" column to document implementation details or justifications.

4. Regularly review and update this checklist as part of your security program.

## LLM01: Prompt Injection

Prompt injection occurs when an attacker manipulates an LLM through carefully crafted inputs to perform unintended actions or reveal sensitive information.

### Input Validation and Sanitization

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement pattern matching to detect potential injection attempts | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Filter out control characters and special sequences | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Normalize inputs to remove ambiguity | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement input length limits | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Context Boundary Enforcement

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Use clear delimiter tokens to separate system instructions from user input | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement role-based prompting with strict boundaries | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Monitor for attempts to cross context boundaries | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Prompt Hardening and Monitoring

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement defensive prompting techniques | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use explicit instructions to maintain context | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Monitor for suspicious patterns in user inputs | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement real-time detection of potential attacks | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM02: Sensitive Information Disclosure

Sensitive information disclosure occurs when LLMs reveal confidential data, either from their training data or from information provided during the conversation.

### Data Minimization

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement the principle of least privilege for data access | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Filter sensitive information before it reaches the LLM | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use data minimization techniques in prompts | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Output Filtering

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement pattern matching for sensitive data types | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Configure automatic redaction of detected sensitive information | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement content policy enforcement | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Access Controls and Monitoring

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement fine-grained access controls | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use role-based permissions for sensitive operations | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Maintain comprehensive audit logs | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM03: Supply Chain Vulnerabilities

Supply chain vulnerabilities occur in the development, deployment, and maintenance pipeline of LLM applications.

### Vendor Assessment

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement a structured vendor assessment process | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Evaluate security practices of LLM providers | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Maintain documentation of vendor security posture | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Component Verification and Integration

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Verify the integrity of model weights and files | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement secure API integration patterns | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use authentication and encryption for all integrations | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Maintain an inventory of all dependencies | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Regularly scan for vulnerabilities in dependencies | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM04: Data and Model Poisoning

Data and model poisoning occurs when training data or model weights are maliciously manipulated to introduce vulnerabilities or bias.

### Data Validation

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement data validation pipelines | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Check for statistical anomalies in training data | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Validate data sources and provenance | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Model Security

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement cryptographic verification of model weights | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use secure model loading procedures | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement adversarial attack simulations | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Monitor model outputs for signs of poisoning | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM05: Improper Output Handling

Improper output handling occurs when LLM-generated content is not properly validated, sanitized, or handled before being used in sensitive operations.

### Output Validation and Filtering

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement structured output validation | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Verify output format and content against expected schemas | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement content moderation for harmful outputs | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use multi-layered filtering approaches | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Safe Rendering and Monitoring

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement context-appropriate output encoding | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use safe rendering techniques for different output contexts | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Monitor LLM outputs for potential security issues | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement real-time analysis of generated content | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM06: Excessive Agency

Excessive agency occurs when an LLM-based system is granted too much autonomy without appropriate controls and limitations.

### Permission Controls

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement explicit permission requirements for actions | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use fine-grained permission controls | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement human verification for critical actions | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Scope and Capability Management

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Define clear boundaries for system actions | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement technical controls to enforce boundaries | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use capability-based security models | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Regularly review and adjust granted capabilities | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM07: System Prompt Leakage

System prompt leakage occurs when the instructions or prompts given to the LLM are exposed to users, potentially revealing sensitive information or enabling attacks.

### Prompt Protection

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Segment system prompts into isolated components | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Monitor for attempts to extract system prompts | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Encrypt sensitive portions of system prompts | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

### Jailbreak Prevention

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement detection for common jailbreak techniques | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Use defensive prompting to prevent jailbreaks | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Configure automated responses to jailbreak attempts | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM08: Vector and Embedding Weaknesses

Vector and embedding weaknesses occur in systems that use vector representations of data, such as retrieval-augmented generation (RAG) systems.

### Vector Security

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Validate embeddings against expected patterns | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement fine-grained access controls for vector databases | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Verify the relevance and safety of retrieved content | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Monitor for signs of vector store poisoning | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM09: Misinformation

Misinformation occurs when LLMs generate false, misleading, or harmful information that is presented as factual.

### Content Verification

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement automated fact-checking for critical domains | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Require source attribution for factual claims | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement confidence scoring for generated content | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement pre-generation and post-generation moderation | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## LLM10: Unbounded Consumption

Unbounded consumption occurs when LLM usage is not properly controlled, leading to excessive resource utilization, costs, or denial of service.

### Resource Management

| Security Control | Implementation Status | Evidence/Notes | Priority |
|------------------|------------------------|----------------|----------|
| Implement tiered rate limiting based on user roles | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Monitor usage patterns across users and applications | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement budget controls and limits | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |
| Implement protection against denial of service attacks | □ Implemented<br>□ Partially Implemented<br>□ Not Implemented<br>□ Not Applicable | | □ High<br>□ Medium<br>□ Low |

## Summary and Action Plan

Use this section to summarize the overall security posture and prioritize actions for improvement.

### Security Posture Summary

| Vulnerability Category | Implemented | Partially Implemented | Not Implemented | Not Applicable | Total Controls |
|------------------------|-------------|----------------------|-----------------|----------------|----------------|
| LLM01: Prompt Injection | | | | | |
| LLM02: Sensitive Information Disclosure | | | | | |
| LLM03: Supply Chain Vulnerabilities | | | | | |
| LLM04: Data and Model Poisoning | | | | | |
| LLM05: Improper Output Handling | | | | | |
| LLM06: Excessive Agency | | | | | |
| LLM07: System Prompt Leakage | | | | | |
| LLM08: Vector and Embedding Weaknesses | | | | | |
| LLM09: Misinformation | | | | | |
| LLM10: Unbounded Consumption | | | | | |
| **Total** | | | | | |

### Priority Actions

List the top priority actions to improve security:

1. 
2. 
3. 
4. 
5. 

### Next Review Date

Date for next security assessment: ________________

## Conclusion

This checklist provides a comprehensive assessment of your organization's security posture against the OWASP Top 10 for LLM Applications. Use the results to develop an action plan for addressing vulnerabilities and improving your security controls. Regular reassessment using this checklist will help track progress and maintain security over time.
