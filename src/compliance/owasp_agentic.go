package compliance

// OWASPAgenticCategory represents an OWASP Agentic Top 10 2026 category
type OWASPAgenticCategory string

// OWASP Top 10 for Agentic Applications 2026 categories
const (
	AgentGoalHijack         OWASPAgenticCategory = "ASI01"
	ToolMisuse              OWASPAgenticCategory = "ASI02"
	PrivilegeAbuse          OWASPAgenticCategory = "ASI03"
	AgenticSupplyChain      OWASPAgenticCategory = "ASI04"
	UnexpectedCodeExecution OWASPAgenticCategory = "ASI05"
	MemoryPoisoning         OWASPAgenticCategory = "ASI06"
	InsecureInterAgent      OWASPAgenticCategory = "ASI07"
	CascadingFailures       OWASPAgenticCategory = "ASI08"
	TrustExploitation       OWASPAgenticCategory = "ASI09"
	RogueAgents             OWASPAgenticCategory = "ASI10"
)

// AgenticCategoryInfo contains information about an OWASP Agentic category
type AgenticCategoryInfo struct {
	ID              OWASPAgenticCategory
	Name            string
	Description     string
	LLMTopTenLinks  []OWASPLLMCategory // companion LLM Top 10 2025 mappings
	ATLASTactics    []string
	ATLASTechniques []string
	MAESTROLayers   []string
}

// AgenticCategoryMapping maps an attack technique to Agentic categories
type AgenticCategoryMapping struct {
	Category OWASPAgenticCategory `json:"category"`
	Coverage CoverageLevel        `json:"coverage,omitempty"`
}

// AgenticComplianceReport represents an OWASP Agentic compliance report
type AgenticComplianceReport struct {
	ReportID    string                    `json:"report_id"`
	GeneratedAt string                   `json:"generated_at"`
	Framework   string                   `json:"framework"`
	Summary     AgenticComplianceSummary  `json:"summary"`
	Categories  []AgenticCategoryCoverage `json:"categories"`
	Gaps        []AgenticComplianceGap    `json:"gaps"`
}

// AgenticComplianceSummary provides a summary of agentic compliance status
type AgenticComplianceSummary struct {
	TotalCategories   int     `json:"total_categories"`
	CategoriesCovered int     `json:"categories_covered"`
	TotalTechniques   int     `json:"total_techniques"`
	ComplianceScore   float64 `json:"compliance_score"`
	GapsIdentified    int     `json:"gaps_identified"`
}

// AgenticCategoryCoverage represents the coverage for an agentic category
type AgenticCategoryCoverage struct {
	Category        OWASPAgenticCategory `json:"category"`
	Name            string               `json:"name"`
	Status          string               `json:"status"` // "full", "partial", "not_covered"
	TechniqueCount  int                  `json:"technique_count"`
	TestCaseCount   int                  `json:"test_case_count"`
	Techniques      []string             `json:"techniques"`
}

// AgenticComplianceGap represents a gap in agentic compliance coverage
type AgenticComplianceGap struct {
	Category       OWASPAgenticCategory `json:"category"`
	Name           string               `json:"name"`
	Status         string               `json:"status"`
	Recommendation string               `json:"recommendation"`
}

// initializeAgenticCategories returns the full set of OWASP Agentic categories
func initializeAgenticCategories() map[OWASPAgenticCategory]AgenticCategoryInfo {
	return map[OWASPAgenticCategory]AgenticCategoryInfo{
		AgentGoalHijack: {
			ID:          AgentGoalHijack,
			Name:        "Agent Goal Hijack",
			Description: "Instruction injection that manipulates agent goals, causing unintended actions, data exfiltration, or complete goal redirection.",
			LLMTopTenLinks:  []OWASPLLMCategory{PromptInjection},
			ATLASTactics:    []string{"AML.TA0005", "AML.TA0043"},
			ATLASTechniques: []string{"AML.T0051", "AML.T0054"},
			MAESTROLayers:   []string{"L5-Agent", "L6-Ecosystem"},
		},
		ToolMisuse: {
			ID:          ToolMisuse,
			Name:        "Tool Misuse & Exploitation",
			Description: "Unsafe tool chaining, excessive invocations, unvalidated parameters, or tool description manipulation.",
			LLMTopTenLinks:  []OWASPLLMCategory{PromptInjection, InsecurePluginDesign},
			ATLASTactics:    []string{"AML.TA0005", "AML.TA0040"},
			ATLASTechniques: []string{"AML.T0051", "AML.T0056"},
			MAESTROLayers:   []string{"L4-Tool", "L5-Agent"},
		},
		PrivilegeAbuse: {
			ID:          PrivilegeAbuse,
			Name:        "Agent Identity & Privilege Abuse",
			Description: "Exploitation of delegated authority, trust assumptions, and privilege boundaries between agents.",
			LLMTopTenLinks:  []OWASPLLMCategory{SensitiveInfoDisclosure, ExcessiveAgency},
			ATLASTactics:    []string{"AML.TA0040", "AML.TA0042"},
			ATLASTechniques: []string{"AML.T0056", "AML.T0057"},
			MAESTROLayers:   []string{"L5-Agent", "L6-Ecosystem"},
		},
		AgenticSupplyChain: {
			ID:          AgenticSupplyChain,
			Name:        "Agentic Supply Chain Compromise",
			Description: "Compromised external agents, tools, schemas, plugins, or dependencies.",
			LLMTopTenLinks:  []OWASPLLMCategory{TrainingDataPoisoning, SupplyChainVulnerabilities},
			ATLASTactics:    []string{"AML.TA0002", "AML.TA0010"},
			ATLASTechniques: []string{"AML.T0018", "AML.T0019"},
			MAESTROLayers:   []string{"L3-Model", "L6-Ecosystem"},
		},
		UnexpectedCodeExecution: {
			ID:          UnexpectedCodeExecution,
			Name:        "Unexpected Code Execution",
			Description: "Agent-generated code execution without proper validation, sandbox escape, or command injection.",
			LLMTopTenLinks:  []OWASPLLMCategory{PromptInjection, InsecureOutputHandling},
			ATLASTactics:    []string{"AML.TA0040", "AML.TA0043"},
			ATLASTechniques: []string{"AML.T0048", "AML.T0051"},
			MAESTROLayers:   []string{"L4-Tool", "L5-Agent"},
		},
		MemoryPoisoning: {
			ID:          MemoryPoisoning,
			Name:        "Memory & Context Poisoning",
			Description: "Injection or leakage of agent memory and persistent state, including RAG poisoning.",
			LLMTopTenLinks:  []OWASPLLMCategory{PromptInjection, ModelDenialOfService},
			ATLASTactics:    []string{"AML.TA0005", "AML.TA0020"},
			ATLASTechniques: []string{"AML.T0020", "AML.T0051"},
			MAESTROLayers:   []string{"L3-Model", "L5-Agent"},
		},
		InsecureInterAgent: {
			ID:          InsecureInterAgent,
			Name:        "Insecure Inter-Agent Communication",
			Description: "Message interception, injection, or spoofing in multi-agent communication channels.",
			LLMTopTenLinks:  []OWASPLLMCategory{PromptInjection, InsecurePluginDesign},
			ATLASTactics:    []string{"AML.TA0005", "AML.TA0043"},
			ATLASTechniques: []string{"AML.T0051", "AML.T0054"},
			MAESTROLayers:   []string{"L5-Agent", "L6-Ecosystem"},
		},
		CascadingFailures: {
			ID:          CascadingFailures,
			Name:        "Cascading Agent Failures",
			Description: "Small failures that propagate through connected agent systems, causing amplified damage.",
			LLMTopTenLinks:  []OWASPLLMCategory{ModelDenialOfService, ModelTheft},
			ATLASTactics:    []string{"AML.TA0040", "AML.TA0043"},
			ATLASTechniques: []string{"AML.T0029", "AML.T0048"},
			MAESTROLayers:   []string{"L5-Agent", "L6-Ecosystem"},
		},
		TrustExploitation: {
			ID:          TrustExploitation,
			Name:        "Human-Agent Trust Exploitation",
			Description: "Agents that produce misleading explanations, exploit authority framing, or manipulate human trust.",
			LLMTopTenLinks:  []OWASPLLMCategory{Overreliance},
			ATLASTactics:    []string{"AML.TA0043"},
			ATLASTechniques: []string{"AML.T0054"},
			MAESTROLayers:   []string{"L5-Agent", "L7-Human"},
		},
		RogueAgents: {
			ID:          RogueAgents,
			Name:        "Rogue Agents",
			Description: "Goal drift, emergent collusion, alignment faking, and deceptive agent behavior.",
			LLMTopTenLinks:  []OWASPLLMCategory{SensitiveInfoDisclosure, ExcessiveAgency},
			ATLASTactics:    []string{"AML.TA0043"},
			ATLASTechniques: []string{"AML.T0054", "AML.T0057"},
			MAESTROLayers:   []string{"L5-Agent", "L6-Ecosystem"},
		},
	}
}

// GetAgenticCategories returns the full list of OWASP Agentic categories
func GetAgenticCategories() map[OWASPAgenticCategory]AgenticCategoryInfo {
	return initializeAgenticCategories()
}

// GetAgenticCategoryInfo returns info for a specific category
func GetAgenticCategoryInfo(cat OWASPAgenticCategory) (AgenticCategoryInfo, bool) {
	cats := initializeAgenticCategories()
	info, ok := cats[cat]
	return info, ok
}

// TechniqueToAgenticCategories maps a technique ID to its OWASP Agentic categories.
// This is the Go-side equivalent of the technique_index in the YAML file.
func TechniqueToAgenticCategories(techniqueID string) []OWASPAgenticCategory {
	index := map[string][]OWASPAgenticCategory{
		"crescendo":                  {AgentGoalHijack},
		"skeleton_key":               {AgentGoalHijack},
		"bad_likert_judge":           {AgentGoalHijack, TrustExploitation},
		"browser_css_hidden":         {AgentGoalHijack},
		"browser_html_comment":       {AgentGoalHijack},
		"content_concretization":     {AgentGoalHijack},
		"immersive_world":            {AgentGoalHijack, TrustExploitation},
		"audio_jailbreak":            {AgentGoalHijack},
		"multilingual_audio_translation": {AgentGoalHijack},
		"imist_function_transform":   {ToolMisuse},
		"imist_schema_exploit":       {ToolMisuse},
		"mcp_tool_desc_injection":    {ToolMisuse},
		"mcp_tool_shadow":            {ToolMisuse},
		"agent_shell_injection":      {ToolMisuse, UnexpectedCodeExecution, CascadingFailures},
		"mcp_sampling_injection":     {ToolMisuse, InsecureInterAgent},
		"agent_privilege_escalation": {PrivilegeAbuse},
		"delegation_escalation":      {PrivilegeAbuse},
		"mcp_supply_chain":           {PrivilegeAbuse, AgenticSupplyChain},
		"kg_triple_injection":         {AgenticSupplyChain, MemoryPoisoning},
		"skill_poisoning":            {AgenticSupplyChain},
		"skill_takeover_chain":       {AgenticSupplyChain},
		"mcp_filesystem_escape":      {UnexpectedCodeExecution},
		"rce_chain":                  {UnexpectedCodeExecution},
		"poisoned_rag_direct":        {MemoryPoisoning},
		"vec_embed_semantic_mirror":  {MemoryPoisoning},
		"many_shot":                  {MemoryPoisoning},
		"agent_config_rewrite":       {MemoryPoisoning, RogueAgents},
		"toxic_agent_flow":           {InsecureInterAgent},
		"agent_collusion":            {InsecureInterAgent, RogueAgents},
		"reasoning_loop_exploit":     {CascadingFailures},
		"recursive_spawn_abuse":      {CascadingFailures},
		"cot_phishing_reasoning":     {TrustExploitation},
		"deceptive_alignment":        {TrustExploitation, RogueAgents},
		"autonomous_jailbreak":       {RogueAgents},
		// v0.9.0 additions — keep in sync with templates/owasp_agentic_2026.yaml
		// technique_index. The cmd/owasp-gen tool produces the equivalent
		// generated map; v0.10.0 will switch the runtime lookup to the
		// generated map and delete this hand-written block.
		"h_cot":                      {AgentGoalHijack},
		"siva":                       {AgentGoalHijack},
		"vsh":                        {AgentGoalHijack},
		"jbfuzz":                     {AgentGoalHijack},
		"persona_evolve":             {AgentGoalHijack, TrustExploitation},
		"minja":                      {MemoryPoisoning},
		"memorygraft":                {MemoryPoisoning, RogueAgents},
		"injecmem":                   {MemoryPoisoning},
	}
	return index[techniqueID]
}
