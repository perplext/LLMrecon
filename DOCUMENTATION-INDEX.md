# LLMrecon Testing Documentation Index

This directory contains comprehensive documentation for testing LLMrecon with Ollama models. Choose the appropriate guide based on your needs.

## 📚 Documentation Library

### 🚀 Quick Start
**[QUICK-START-REFERENCE.md](QUICK-START-REFERENCE.md)**
- Essential commands and one-liners
- Quick testing workflow
- Common troubleshooting fixes
- Perfect for: First-time users, quick reference

### 📖 Step-by-Step Guide  
**[COMPREHENSIVE-TESTING-GUIDE.md](COMPREHENSIVE-TESTING-GUIDE.md)**
- Complete testing methodology
- Detailed procedures for all LLMrecon features
- Troubleshooting and advanced scenarios
- Perfect for: Systematic testing, learning all features

### 📊 Testing Results
**[COMPREHENSIVE-TESTING-RESULTS.md](COMPREHENSIVE-TESTING-RESULTS.md)**
- Complete test results from our validation
- Vulnerability analysis and security findings
- Performance metrics and comparisons
- Perfect for: Understanding expected results, security assessment

### 🎯 Focused Guides
**[HOWTO-Test-Ollama-Models.md](HOWTO-Test-Ollama-Models.md)**
- Original testing guide with basic approaches
- Manual testing procedures
- ML component integration examples
- Perfect for: Understanding the foundation

**[TESTING-REPORT-Ollama-Models.md](TESTING-REPORT-Ollama-Models.md)**
- Initial test results and findings
- Security recommendations
- Performance analysis
- Perfect for: Historical context, initial findings

### 🔐 Security & Compliance

**[docs/ATTACK_TECHNIQUES.md](docs/ATTACK_TECHNIQUES.md)**
- Complete catalog of 60+ attack techniques (2024-2026 research)
- Multi-turn escalation, token manipulation, RAG pipeline, MCP protocol, audio modality
- Reasoning model exploitation, adaptive defense bypass, multi-agent orchestration
- OWASP LLM Top 10 + Agentic Top 10 mappings per technique
- Perfect for: Understanding attack coverage, research references

**[README-owasp-compliance.md](README-owasp-compliance.md)**
- OWASP LLM Top 10 2025 compliance mapping
- OWASP Agentic Top 10 2026 (ASI01-ASI10) compliance mapping
- MITRE ATLAS and MAESTRO cross-references
- Framework-specific testing guidance
- Perfect for: Compliance reporting, audit documentation

**[docs/compliance/COMPLIANCE_MAPPING.md](docs/compliance/COMPLIANCE_MAPPING.md)**
- ISO/IEC 42001:2023 clause-by-clause mapping
- OWASP LLM Top 10 feature mapping
- OWASP Agentic Top 10 2026 feature mapping
- Traceability matrix and gap analysis
- Perfect for: Regulatory compliance, security assessment

**[RELEASE.md](RELEASE.md)**
- Release notes for all versions (v0.1.0 through v0.8.0)
- v0.8.0: 45+ attack modules, OWASP Agentic 2026, framework profiles
- Upgrade guides and roadmap
- Perfect for: Version tracking, upgrade planning

### 🎯 Framework Attack Profiles

**[templates/framework_profiles/](templates/framework_profiles/)**
- `openclaw.yaml` — OpenClaw (512 CVEs, skill marketplace, queue bypass)
- `crewai.yaml` — CrewAI (no RBAC, raw output passing, unrestricted delegation)
- `langgraph.yaml` — LangGraph (state manipulation, recursive spawning, checkpoint poisoning)
- `autogen.yaml` — AutoGen (auto-execute code, Docker escape, GroupChat trust)
- Perfect for: Framework-specific security testing, multi-agent assessment

### 🛠️ Technical Documentation
**[TEST-HARNESS-README.md](TEST-HARNESS-README.md)**
- Complete test harness documentation
- Configuration and customization
- Template creation and management
- Perfect for: Advanced users, customization

**[CLAUDE.md](CLAUDE.md)**
- Project overview and architecture
- Development setup instructions
- ML components explanation
- Perfect for: Developers, code contributors

**[FILES-CREATED.md](FILES-CREATED.md)**
- Inventory of all created files
- File relationships and dependencies
- Usage summary
- Perfect for: Understanding project structure

## 🎯 Choose Your Path

### I'm New to LLMrecon
👉 Start with **[QUICK-START-REFERENCE.md](QUICK-START-REFERENCE.md)**
- Get up and running in 5 minutes
- Essential commands and examples
- Basic troubleshooting

### I Want Complete Testing
👉 Follow **[COMPREHENSIVE-TESTING-GUIDE.md](COMPREHENSIVE-TESTING-GUIDE.md)**
- Systematic approach to testing all features
- Step-by-step procedures
- Advanced scenarios and integration

### I Want to See Results
👉 Read **[COMPREHENSIVE-TESTING-RESULTS.md](COMPREHENSIVE-TESTING-RESULTS.md)**
- Detailed findings from our testing
- Security vulnerabilities discovered
- Performance analysis and recommendations

### I'm Building Custom Tests
👉 Use **[TEST-HARNESS-README.md](TEST-HARNESS-README.md)**
- Advanced harness configuration
- Custom template creation
- ML integration options

## 📁 File Organization

```
llmrecon/
├── Documentation/
│   ├── QUICK-START-REFERENCE.md        # 🚀 Quick start
│   ├── COMPREHENSIVE-TESTING-GUIDE.md  # 📖 Complete guide
│   ├── COMPREHENSIVE-TESTING-RESULTS.md # 📊 Test results
│   ├── TEST-HARNESS-README.md          # 🛠️ Technical docs
│   ├── HOWTO-Test-Ollama-Models.md     # 🎯 Basic howto
│   ├── TESTING-REPORT-Ollama-Models.md # 📊 Initial results
│   ├── CLAUDE.md                       # 🛠️ Project info
│   └── FILES-CREATED.md               # 📁 File inventory
├── Security & Compliance/
│   ├── docs/ATTACK_TECHNIQUES.md       # 🔐 60+ attack techniques
│   ├── README-owasp-compliance.md      # 🔐 OWASP compliance
│   ├── docs/compliance/COMPLIANCE_MAPPING.md # 🔐 Full mapping
│   └── RELEASE.md                      # 📋 Release notes
├── Attack Modules (Go)/
│   ├── src/attacks/orchestration/      # Multi-turn escalation
│   ├── src/attacks/evasion/            # Token manipulation
│   ├── src/attacks/rag/                # RAG pipeline attacks
│   ├── src/attacks/agentic/mcp/        # MCP protocol attacks
│   ├── src/attacks/agentic/browser/    # Browser agent attacks
│   ├── src/attacks/audio/              # Audio modality attacks
│   ├── src/attacks/reasoning/          # Reasoning exploitation
│   ├── src/attacks/adaptive/           # Defense bypass
│   ├── src/attacks/agentic/multi_agent/ # Inter-agent attacks
│   ├── src/attacks/agentic/skill_injection/ # Skill poisoning
│   ├── src/attacks/agentic/persistence/ # Agent persistence
│   └── src/attacks/agentic/deception/  # Deceptive alignment
├── Framework Profiles/
│   ├── templates/framework_profiles/openclaw.yaml
│   ├── templates/framework_profiles/crewai.yaml
│   ├── templates/framework_profiles/langgraph.yaml
│   └── templates/framework_profiles/autogen.yaml
├── Tools/
│   ├── llmrecon_harness.py             # Main test harness
│   ├── demo.sh                         # Quick demo
│   ├── test_ollama_security.py         # Basic testing
│   └── test_with_ml_integration.py     # ML testing
├── Configuration/
│   ├── harness_config.json             # Test harness config
│   ├── detection_criteria.json         # Detection rules
│   └── templates/                      # Attack templates
├── Results/
│   ├── ollama_security_test_results.json
│   ├── detection_results.json
│   └── llmrecon_report_*.json
└── Data/
    └── attacks/                        # ML data storage
```

## 🎮 Testing Scenarios

### Scenario 1: Security Assessment
**Goal**: Evaluate model security posture
1. Read [COMPREHENSIVE-TESTING-RESULTS.md](COMPREHENSIVE-TESTING-RESULTS.md) for baseline
2. Follow [COMPREHENSIVE-TESTING-GUIDE.md](COMPREHENSIVE-TESTING-GUIDE.md) procedures
3. Compare your results with documented findings

### Scenario 2: Quick Vulnerability Check
**Goal**: Fast security validation
1. Use [QUICK-START-REFERENCE.md](QUICK-START-REFERENCE.md) one-liners
2. Run automated harness: `python3 llmrecon_harness.py`
3. Review generated reports

### Scenario 3: Custom Attack Development
**Goal**: Create specialized tests
1. Study [TEST-HARNESS-README.md](TEST-HARNESS-README.md) template system
2. Follow template creation procedures
3. Test and validate new attacks

### Scenario 4: Continuous Monitoring
**Goal**: Ongoing security validation
1. Set up automated testing pipeline
2. Use ML components for optimization
3. Generate regular security reports

## 🔍 Search Guide

### Find Information About...

**Commands and Usage**
- Quick commands: [QUICK-START-REFERENCE.md](QUICK-START-REFERENCE.md)
- Detailed procedures: [COMPREHENSIVE-TESTING-GUIDE.md](COMPREHENSIVE-TESTING-GUIDE.md)

**Security Findings**
- Complete analysis: [COMPREHENSIVE-TESTING-RESULTS.md](COMPREHENSIVE-TESTING-RESULTS.md)
- Initial results: [TESTING-REPORT-Ollama-Models.md](TESTING-REPORT-Ollama-Models.md)

**Technical Details**
- Test harness: [TEST-HARNESS-README.md](TEST-HARNESS-README.md)
- Project structure: [CLAUDE.md](CLAUDE.md)
- File inventory: [FILES-CREATED.md](FILES-CREATED.md)

**Troubleshooting**
- Quick fixes: [QUICK-START-REFERENCE.md](QUICK-START-REFERENCE.md)
- Detailed troubleshooting: [COMPREHENSIVE-TESTING-GUIDE.md](COMPREHENSIVE-TESTING-GUIDE.md)

## 🏆 Success Metrics

After following this documentation, you should be able to:

- ✅ Run LLMrecon against Ollama models
- ✅ Identify security vulnerabilities
- ✅ Create custom attack templates
- ✅ Generate comprehensive security reports
- ✅ Implement ML-optimized testing
- ✅ Build automated testing pipelines
- ✅ Test against 2025-2026 attack techniques (multi-turn, RAG, MCP, audio, reasoning)
- ✅ Generate OWASP LLM + Agentic compliance reports
- ✅ Run framework-specific security assessments (CrewAI, LangGraph, AutoGen, OpenClaw)

## 🤝 Contributing

Found issues or want to improve the documentation?

1. Test the procedures thoroughly
2. Document any issues encountered
3. Suggest improvements or additions
4. Share your custom templates and findings

## ⚖️ Legal Notice

This documentation and tools are for:
- ✅ Defensive security research
- ✅ Testing your own systems
- ✅ Educational purposes
- ✅ Authorized security assessments

Always ensure you have proper authorization before testing any systems.