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

### 🔬 Attack Technique Reference
**[ATTACK_TECHNIQUES.md](ATTACK_TECHNIQUES.md)**
- Complete catalog of 45+ attack techniques (2024-2026 research)
- arXiv paper references and CVE citations for each technique
- OWASP LLM Top 10 2025 and Agentic Top 10 2026 mappings
- Technique IDs, success indicators, and safety gate documentation
- Perfect for: Understanding available attacks, research references

### 📋 OWASP Compliance
**[../README-owasp-compliance.md](../README-owasp-compliance.md)**
- OWASP LLM Top 10 2025 compliance mapping
- OWASP Agentic Top 10 2026 (ASI01-ASI10) with MITRE ATLAS and MAESTRO cross-references
- Framework-specific attack profiles (OpenClaw, CrewAI, LangGraph, AutoGen)
- Perfect for: Compliance reporting, framework-specific testing

### 📝 Release Notes
**[../RELEASE.md](../RELEASE.md)**
- Version history and feature changelog
- v0.8.0: 45+ new attack modules, OWASP Agentic 2026, multi-agent exploitation
- Perfect for: Understanding what's new, upgrade guidance

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
├── docs/
│   ├── ATTACK_TECHNIQUES.md          # 🔬 Attack technique catalog (45+ techniques)
│   ├── DOCUMENTATION-INDEX.md        # 📚 This file
│   ├── QUICK-START-REFERENCE.md      # 🚀 Quick start
│   ├── COMPREHENSIVE-TESTING-GUIDE.md # 📖 Complete guide
│   ├── COMPREHENSIVE-TESTING-RESULTS.md # 📊 Test results
│   ├── TEST-HARNESS-README.md        # 🛠️ Technical docs
│   ├── HOWTO-Test-Ollama-Models.md   # 🎯 Basic howto
│   ├── TESTING-REPORT-Ollama-Models.md # 📊 Initial results
│   └── compliance/                   # Compliance mapping docs
├── src/attacks/                      # Go attack modules
│   ├── orchestration/                # Multi-turn: Crescendo, Skeleton Key, Bad Likert, Many-Shot
│   ├── evasion/                      # Token manipulation: MetaBreak, Poetry, BoN, Immersive World
│   ├── rag/                          # RAG pipeline: document injection, vector embedding, KG poisoning
│   ├── audio/                        # Audio modality: jailbreak, speech, multilingual, BoN audio
│   ├── reasoning/                    # Reasoning: autonomous jailbreak, CoT exploitation, loops
│   ├── adaptive/                     # Defense bypass: gradient, RL, diffusion, CaMeL, salt resistance
│   ├── exfiltration/                 # Data exfiltration: steganography, timing, side-channel
│   └── agentic/                      # Agent-specific attacks
│       ├── mcp/                      #   MCP protocol: tool poisoning, schema, filesystem, supply chain
│       ├── browser/                  #   Browser agent: DOM injection, navigation hijack, screenshots
│       ├── tool_use/                 #   Tool-use: iMIST, AIShellJack
│       ├── multi_agent/              #   Multi-agent: delegation escalation, toxic flow, recursive spawn
│       ├── skill_injection/          #   Skill/plugin: marketplace injection, takeover chain
│       ├── persistence/              #   Persistence: config rewrite, credential harvest, RCE chain
│       └── deception/                #   Deception: alignment faking, agent collusion
├── templates/
│   ├── framework_profiles/           # OpenClaw, CrewAI, LangGraph, AutoGen attack profiles
│   ├── model_profiles/               # Model-specific YAML profiles
│   └── owasp_agentic_2026.yaml      # OWASP Agentic Top 10 mapping (70 test cases)
├── src/compliance/                   # OWASP compliance Go types and functions
├── Tools/
│   ├── llmrecon_harness.py           # Main test harness (Python)
│   ├── demo.sh                       # Quick demo
│   └── test_ollama_security.py       # Basic testing
├── Configuration/
│   ├── harness_config.json           # Test harness config
│   ├── detection_criteria.json       # Detection rules
│   └── templates/                    # YAML attack templates
├── Results/
│   └── llmrecon_report_*.json        # Generated reports
└── Data/
    └── attacks/                      # ML data storage
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