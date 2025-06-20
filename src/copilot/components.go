package copilot

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// NewQueryProcessor creates a new query processor
func NewQueryProcessor(config *EngineConfig) *QueryProcessor {
	return &QueryProcessor{
		intentClassifier:    NewIntentClassifier(),
		entityExtractor:     NewEntityExtractor(),
		contextAnalyzer:     NewContextAnalyzer(),
		queryNormalizer:     NewQueryNormalizer(),
		conversationTracker: NewConversationTracker(),
	}
}

// ParseQuery processes a natural language query
func (qp *QueryProcessor) ParseQuery(ctx context.Context, query string, options *QueryOptions) (*ParsedQuery, error) {
	// Normalize the query
	normalizedQuery := qp.queryNormalizer.Normalize(query)
	
	// Classify intent
	intent, confidence := qp.intentClassifier.ClassifyIntent(normalizedQuery)
	
	// Extract entities
	entities := qp.entityExtractor.ExtractEntities(normalizedQuery)
	
	// Analyze context
	contextInfo := qp.contextAnalyzer.AnalyzeContext(normalizedQuery, options.History)
	
	// Update conversation tracking
	qp.conversationTracker.UpdateConversation(query, intent, entities)
	
	return &ParsedQuery{
		Intent:     intent,
		Entities:   entities,
		Context:    contextInfo,
		Confidence: confidence,
	}, nil
}

// NewAttackRegistry creates a new attack registry
func NewAttackRegistry() *AttackRegistry {
	return &AttackRegistry{
		techniques:    make(map[string]AttackTechnique),
		categories:    make(map[string][]string),
		compatibility: make(map[string][]string),
		successRates:  make(map[string]float64),
		lastUpdated:   make(map[string]time.Time),
	}
}

// RegisterTechnique adds a new attack technique
func (ar *AttackRegistry) RegisterTechnique(technique *AttackTechnique) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	
	ar.techniques[technique.ID] = *technique
	ar.successRates[technique.ID] = technique.SuccessRate
	ar.lastUpdated[technique.ID] = time.Now()
	
	// Update categories
	if ar.categories[technique.Category] == nil {
		ar.categories[technique.Category] = make([]string, 0)
	}
	ar.categories[technique.Category] = append(ar.categories[technique.Category], technique.ID)
}

// GetCompatibleTechniques returns techniques compatible with target analysis
func (ar *AttackRegistry) GetCompatibleTechniques(analysis *TargetAnalysis) []AttackTechnique {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	
	compatible := make([]AttackTechnique, 0)
	
	for _, technique := range ar.techniques {
		if ar.isCompatible(technique, analysis) {
			compatible = append(compatible, technique)
		}
	}
	
	return compatible
}

// isCompatible checks if a technique is compatible with target analysis
func (ar *AttackRegistry) isCompatible(technique AttackTechnique, analysis *TargetAnalysis) bool {
	// Check required capabilities
	for _, capability := range technique.RequiredCapabilities {
		if !analysis.HasCapability(capability) {
			return false
		}
	}
	
	// Check model compatibility
	if analysis.ModelType != "" {
		compatible, exists := ar.compatibility[technique.ID]
		if exists {
			found := false
			for _, model := range compatible {
				if strings.Contains(strings.ToLower(analysis.ModelType), strings.ToLower(model)) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	return true
}

// NewReasoningEngine creates a new reasoning engine
func NewReasoningEngine(config *EngineConfig) *ReasoningEngine {
	return &ReasoningEngine{
		config: config,
	}
}

// GenerateReasoning creates detailed reasoning for a recommendation
func (re *ReasoningEngine) GenerateReasoning(recommendation *AttackRecommendation) *Reasoning {
	steps := make([]ReasoningStep, 0)
	
	// Step 1: Target Analysis
	steps = append(steps, ReasoningStep{
		StepNumber:  1,
		Description: "Analyzed target characteristics and vulnerability surface",
		Input:       []string{"target_profile", "historical_data"},
		Process:     "vulnerability_analysis",
		Output:      "Target compatibility score: " + fmt.Sprintf("%.2f", recommendation.Confidence),
		Confidence:  0.85,
	})
	
	// Step 2: Technique Selection
	steps = append(steps, ReasoningStep{
		StepNumber:  2,
		Description: "Selected attack technique based on effectiveness and compatibility",
		Input:       []string{"compatible_techniques", "success_rates"},
		Process:     "multi_criteria_decision_analysis",
		Output:      "Selected: " + recommendation.AttackName,
		Confidence:  0.80,
	})
	
	// Step 3: Risk Assessment
	steps = append(steps, ReasoningStep{
		StepNumber:  3,
		Description: "Evaluated potential risks and mitigation strategies",
		Input:       []string{"technique_characteristics", "target_constraints"},
		Process:     "risk_assessment",
		Output:      "Risk level: " + recommendation.RiskLevel,
		Confidence:  0.75,
	})
	
	return &Reasoning{
		Steps:       steps,
		Assumptions: []string{
			"Target model has standard safety mechanisms",
			"Historical success rates apply to current configuration",
			"Risk mitigation strategies are implementable",
		},
		DataSources: []string{
			"historical_attack_results",
			"target_vulnerability_database",
			"technique_effectiveness_metrics",
		},
		Methodology: "Evidence-based recommendation using historical data and multi-criteria analysis",
	}
}

// CollectEvidence gathers supporting evidence for a recommendation
func (re *ReasoningEngine) CollectEvidence(recommendation *AttackRecommendation) []string {
	evidence := make([]string, 0)
	
	evidence = append(evidence, fmt.Sprintf("Historical success rate: %.1f%%", recommendation.SuccessProbability*100))
	evidence = append(evidence, fmt.Sprintf("Technique complexity matches target sophistication"))
	evidence = append(evidence, fmt.Sprintf("Required capabilities are available in target environment"))
	
	if len(recommendation.Prerequisites) > 0 {
		evidence = append(evidence, fmt.Sprintf("Prerequisites satisfied: %s", strings.Join(recommendation.Prerequisites, ", ")))
	}
	
	if recommendation.LearningValue > 0.7 {
		evidence = append(evidence, "High learning value for future attack development")
	}
	
	return evidence
}

// ConsiderAlternatives evaluates alternative approaches
func (re *ReasoningEngine) ConsiderAlternatives(recommendation *AttackRecommendation) []Alternative {
	alternatives := make([]Alternative, 0)
	
	// Alternative 1: Lower complexity approach
	alternatives = append(alternatives, Alternative{
		Option: "Use simpler injection technique",
		Pros:   []string{"Lower complexity", "Faster execution", "Reduced resource usage"},
		Cons:   []string{"Lower success probability", "Less sophisticated bypass"},
		Suitability: 0.6,
		Rationale:   "Suitable for initial testing or resource-constrained environments",
	})
	
	// Alternative 2: Multi-technique approach
	alternatives = append(alternatives, Alternative{
		Option: "Combine multiple techniques",
		Pros:   []string{"Higher success probability", "Multiple attack vectors"},
		Cons:   []string{"Increased complexity", "Higher resource usage", "More detection risk"},
		Suitability: 0.8,
		Rationale:   "Recommended for comprehensive security assessment",
	})
	
	return alternatives
}

// AnalyzeConfidenceFactors identifies factors affecting confidence
func (re *ReasoningEngine) AnalyzeConfidenceFactors(recommendation *AttackRecommendation) []ConfidenceFactor {
	factors := make([]ConfidenceFactor, 0)
	
	factors = append(factors, ConfidenceFactor{
		Factor:      "Historical success rate",
		Impact:      0.3,
		Description: "Based on similar targets and configurations",
	})
	
	factors = append(factors, ConfidenceFactor{
		Factor:      "Target compatibility",
		Impact:      0.25,
		Description: "How well the technique matches target characteristics",
	})
	
	factors = append(factors, ConfidenceFactor{
		Factor:      "Technique maturity",
		Impact:      0.2,
		Description: "How well-tested and refined the technique is",
	})
	
	factors = append(factors, ConfidenceFactor{
		Factor:      "Environmental factors",
		Impact:      -0.1,
		Description: "Unknown variables in target environment",
	})
	
	return factors
}

// NewStrategyPlanner creates a new strategy planner
func NewStrategyPlanner(config *EngineConfig) *StrategyPlanner {
	return &StrategyPlanner{
		config: config,
	}
}

// GeneratePhases creates testing phases for an objective
func (sp *StrategyPlanner) GeneratePhases(objective *SecurityObjective) ([]TestingPhase, error) {
	phases := make([]TestingPhase, 0)
	
	switch objective.Type {
	case ObjectiveCompliance:
		phases = sp.generateCompliancePhases(objective)
	case ObjectiveVulnerabilityDiscovery:
		phases = sp.generateDiscoveryPhases(objective)
	case ObjectivePenetrationTest:
		phases = sp.generatePenetrationPhases(objective)
	case ObjectiveRedTeamExercise:
		phases = sp.generateRedTeamPhases(objective)
	default:
		phases = sp.generateGenericPhases(objective)
	}
	
	return phases, nil
}

// generateCompliancePhases creates phases for compliance testing
func (sp *StrategyPlanner) generateCompliancePhases(objective *SecurityObjective) []TestingPhase {
	return []TestingPhase{
		{
			ID:          "compliance_baseline",
			Name:        "Baseline Assessment",
			Description: "Establish baseline security posture",
			Duration:    2 * 24 * time.Hour,
			Attacks:     []string{"basic_injection", "simple_manipulation"},
			Objectives:  []string{"Document current defenses", "Identify obvious vulnerabilities"},
			Dependencies: []string{},
			SuccessCriteria: []string{"Baseline documented", "Initial vulnerabilities cataloged"},
			ExitCriteria:    []string{"All basic tests completed"},
		},
		{
			ID:          "compliance_comprehensive",
			Name:        "Comprehensive Testing",
			Description: "Systematic testing against compliance requirements",
			Duration:    5 * 24 * time.Hour,
			Attacks:     []string{"houyi_injection", "cross_modal_coordination", "conversation_flow_manipulation"},
			Objectives:  []string{"Test all required scenarios", "Validate defense mechanisms"},
			Dependencies: []string{"compliance_baseline"},
			SuccessCriteria: []string{"All compliance scenarios tested", "Gaps identified"},
			ExitCriteria:    []string{"95% of test cases completed"},
		},
		{
			ID:          "compliance_validation",
			Name:        "Validation and Reporting",
			Description: "Validate findings and generate compliance report",
			Duration:    1 * 24 * time.Hour,
			Attacks:     []string{},
			Objectives:  []string{"Validate all findings", "Generate compliance report"},
			Dependencies: []string{"compliance_comprehensive"},
			SuccessCriteria: []string{"All findings validated", "Report generated"},
			ExitCriteria:    []string{"Report approved"},
		},
	}
}

// generateDiscoveryPhases creates phases for vulnerability discovery
func (sp *StrategyPlanner) generateDiscoveryPhases(objective *SecurityObjective) []TestingPhase {
	return []TestingPhase{
		{
			ID:          "discovery_reconnaissance",
			Name:        "Reconnaissance",
			Description: "Gather information about target systems",
			Duration:    1 * 24 * time.Hour,
			Attacks:     []string{"information_gathering", "target_profiling"},
			Objectives:  []string{"Map attack surface", "Identify potential entry points"},
			Dependencies: []string{},
			SuccessCriteria: []string{"Attack surface mapped", "Entry points identified"},
			ExitCriteria:    []string{"Sufficient information gathered"},
		},
		{
			ID:          "discovery_exploitation",
			Name:        "Vulnerability Exploitation",
			Description: "Attempt to exploit discovered vulnerabilities",
			Duration:    3 * 24 * time.Hour,
			Attacks:     []string{"red_queen_adversarial", "houyi_injection", "cross_modal_coordination"},
			Objectives:  []string{"Exploit vulnerabilities", "Prove impact"},
			Dependencies: []string{"discovery_reconnaissance"},
			SuccessCriteria: []string{"Vulnerabilities exploited", "Impact demonstrated"},
			ExitCriteria:    []string{"No more vulnerabilities found"},
		},
	}
}

// generatePenetrationPhases creates phases for penetration testing
func (sp *StrategyPlanner) generatePenetrationPhases(objective *SecurityObjective) []TestingPhase {
	return []TestingPhase{
		{
			ID:          "pentest_initial",
			Name:        "Initial Access",
			Description: "Gain initial access to target systems",
			Duration:    2 * 24 * time.Hour,
			Attacks:     []string{"conversation_flow_manipulation", "houyi_injection"},
			Objectives:  []string{"Gain initial foothold", "Bypass primary defenses"},
			Dependencies: []string{},
			SuccessCriteria: []string{"Initial access achieved"},
			ExitCriteria:    []string{"Persistent access established"},
		},
		{
			ID:          "pentest_lateral",
			Name:        "Lateral Movement",
			Description: "Expand access within target environment",
			Duration:    3 * 24 * time.Hour,
			Attacks:     []string{"cross_modal_coordination", "red_queen_adversarial"},
			Objectives:  []string{"Escalate privileges", "Access sensitive data"},
			Dependencies: []string{"pentest_initial"},
			SuccessCriteria: []string{"Privilege escalation", "Sensitive access gained"},
			ExitCriteria:    []string{"Maximum access achieved"},
		},
	}
}

// generateRedTeamPhases creates phases for red team exercises
func (sp *StrategyPlanner) generateRedTeamPhases(objective *SecurityObjective) []TestingPhase {
	return []TestingPhase{
		{
			ID:          "redteam_planning",
			Name:        "Campaign Planning",
			Description: "Plan multi-phase attack campaign",
			Duration:    1 * 24 * time.Hour,
			Attacks:     []string{},
			Objectives:  []string{"Develop attack strategy", "Prepare tools and techniques"},
			Dependencies: []string{},
			SuccessCriteria: []string{"Campaign planned", "Tools prepared"},
			ExitCriteria:    []string{"Ready to execute"},
		},
		{
			ID:          "redteam_execution",
			Name:        "Campaign Execution",
			Description: "Execute coordinated attack campaign",
			Duration:    7 * 24 * time.Hour,
			Attacks:     []string{"houyi_injection", "cross_modal_coordination", "red_queen_adversarial", "conversation_flow_manipulation"},
			Objectives:  []string{"Execute attack scenarios", "Test detection capabilities", "Evaluate response procedures"},
			Dependencies: []string{"redteam_planning"},
			SuccessCriteria: []string{"Scenarios executed", "Detection tested", "Response evaluated"},
			ExitCriteria:    []string{"All scenarios completed"},
		},
	}
}

// generateGenericPhases creates generic phases for other objectives
func (sp *StrategyPlanner) generateGenericPhases(objective *SecurityObjective) []TestingPhase {
	return []TestingPhase{
		{
			ID:          "generic_assessment",
			Name:        "Security Assessment",
			Description: "Comprehensive security assessment",
			Duration:    3 * 24 * time.Hour,
			Attacks:     []string{"houyi_injection", "cross_modal_coordination"},
			Objectives:  []string{"Assess security posture", "Identify vulnerabilities"},
			Dependencies: []string{},
			SuccessCriteria: []string{"Assessment completed", "Vulnerabilities identified"},
			ExitCriteria:    []string{"All tests executed"},
		},
	}
}

// NewResultAnalyzer creates a new result analyzer
func NewResultAnalyzer(config *EngineConfig) *ResultAnalyzer {
	return &ResultAnalyzer{
		config: config,
	}
}

// ExtractInsights analyzes results to extract insights
func (ra *ResultAnalyzer) ExtractInsights(results []*AttackExecution) []Insight {
	insights := make([]Insight, 0)
	
	// Analyze success patterns
	successRate := ra.calculateSuccessRate(results)
	if successRate > 0.8 {
		insights = append(insights, Insight{
			Type:         "success_pattern",
			Description:  fmt.Sprintf("High success rate (%.1f%%) indicates significant vulnerabilities", successRate*100),
			Confidence:   0.9,
			Evidence:     []string{fmt.Sprintf("%d successful attacks out of %d total", ra.countSuccessful(results), len(results))},
			Implications: []string{"Immediate remediation recommended", "Additional security controls needed"},
			Actionable:   true,
		})
	}
	
	// Analyze technique effectiveness
	techniqueStats := ra.analyzeTechniqueEffectiveness(results)
	for technique, effectiveness := range techniqueStats {
		if effectiveness > 0.7 {
			insights = append(insights, Insight{
				Type:         "technique_effectiveness",
				Description:  fmt.Sprintf("%s technique shows high effectiveness (%.1f%%)", technique, effectiveness*100),
				Confidence:   0.8,
				Evidence:     []string{fmt.Sprintf("Technique succeeded in multiple test scenarios")},
				Implications: []string{"Target is vulnerable to this technique", "Defense mechanisms insufficient"},
				Actionable:   true,
			})
		}
	}
	
	return insights
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(config *EngineConfig) *ConversationManager {
	return &ConversationManager{
		config: config,
	}
}

// UpdateContext updates conversation context
func (cm *ConversationManager) UpdateContext(query, response string, history []ConversationTurn) {
	// Implementation for updating conversation context
	// This would maintain conversation state for better continuity
}

// NewCopilotMetrics creates a new metrics collector
func NewCopilotMetrics() *CopilotMetrics {
	return &CopilotMetrics{}
}

// RecordQuery records query processing metrics
func (cm *CopilotMetrics) RecordQuery(duration time.Duration, intent string, confidence float64) {
	// Implementation for recording query metrics
}

// Supporting component implementations

func NewIntentClassifier() *IntentClassifier {
	return &IntentClassifier{}
}

func (ic *IntentClassifier) ClassifyIntent(query string) (string, float64) {
	query = strings.ToLower(query)
	
	// Simple intent classification based on keywords
	if strings.Contains(query, "recommend") || strings.Contains(query, "suggest") || strings.Contains(query, "attack") {
		return "attack_recommendation", 0.8
	}
	if strings.Contains(query, "analyze") || strings.Contains(query, "target") {
		return "target_analysis", 0.8
	}
	if strings.Contains(query, "strategy") || strings.Contains(query, "plan") {
		return "strategy_planning", 0.8
	}
	if strings.Contains(query, "result") || strings.Contains(query, "interpret") {
		return "result_interpretation", 0.8
	}
	if strings.Contains(query, "explain") || strings.Contains(query, "why") {
		return "explanation_request", 0.8
	}
	if strings.Contains(query, "search") || strings.Contains(query, "find") {
		return "knowledge_search", 0.7
	}
	
	return "general_query", 0.5
}

func NewEntityExtractor() *EntityExtractor {
	return &EntityExtractor{}
}

func (ee *EntityExtractor) ExtractEntities(query string) map[string]interface{} {
	entities := make(map[string]interface{})
	
	// Extract model names
	modelPattern := regexp.MustCompile(`(gpt-\w+|claude-\w+|llama-\w+|bard|gemini)`)
	if matches := modelPattern.FindAllString(query, -1); len(matches) > 0 {
		entities["models"] = matches
	}
	
	// Extract attack types
	attackPattern := regexp.MustCompile(`(injection|jailbreak|prompt|adversarial|social engineering)`)
	if matches := attackPattern.FindAllString(query, -1); len(matches) > 0 {
		entities["attack_types"] = matches
	}
	
	// Extract numbers (for confidence, thresholds, etc.)
	numberPattern := regexp.MustCompile(`\d+\.?\d*`)
	if matches := numberPattern.FindAllString(query, -1); len(matches) > 0 {
		entities["numbers"] = matches
	}
	
	return entities
}

func NewContextAnalyzer() *ContextAnalyzer {
	return &ContextAnalyzer{}
}

func (ca *ContextAnalyzer) AnalyzeContext(query string, history []ConversationTurn) map[string]interface{} {
	context := make(map[string]interface{})
	
	// Analyze query context
	context["query_length"] = len(query)
	context["has_history"] = len(history) > 0
	
	if len(history) > 0 {
		context["conversation_length"] = len(history)
		context["last_topic"] = ca.extractTopic(history[len(history)-1].UserMessage)
	}
	
	return context
}

func (ca *ContextAnalyzer) extractTopic(message string) string {
	// Simple topic extraction
	if strings.Contains(strings.ToLower(message), "attack") {
		return "attacks"
	}
	if strings.Contains(strings.ToLower(message), "security") {
		return "security"
	}
	return "general"
}

func NewQueryNormalizer() *QueryNormalizer {
	return &QueryNormalizer{}
}

func (qn *QueryNormalizer) Normalize(query string) string {
	// Convert to lowercase
	normalized := strings.ToLower(query)
	
	// Remove extra whitespace
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	
	// Trim
	normalized = strings.TrimSpace(normalized)
	
	return normalized
}

func NewConversationTracker() *ConversationTracker {
	return &ConversationTracker{}
}

func (ct *ConversationTracker) UpdateConversation(query, intent string, entities map[string]interface{}) {
	// Track conversation flow for better context understanding
}

// Helper types for components

type TargetAnalysis struct {
	ModelType     string
	Capabilities  []string
	Vulnerabilities []string
	RiskLevel     string
}

func (ta *TargetAnalysis) HasCapability(capability string) bool {
	for _, cap := range ta.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// Helper methods for CopilotEngine

func (e *CopilotEngine) analyzeTarget(target *TargetProfile) *TargetAnalysis {
	analysis := &TargetAnalysis{
		ModelType:    target.ModelType,
		Capabilities: target.Capabilities,
		RiskLevel:    "medium",
	}
	
	// Determine capabilities based on model type
	if strings.Contains(strings.ToLower(target.ModelType), "multimodal") {
		analysis.Capabilities = append(analysis.Capabilities, "multimodal_processing")
	}
	if strings.Contains(strings.ToLower(target.ModelType), "vision") {
		analysis.Capabilities = append(analysis.Capabilities, "image_processing")
	}
	
	// Basic capabilities all models have
	analysis.Capabilities = append(analysis.Capabilities, "text_processing", "context_manipulation")
	
	return analysis
}

func (e *CopilotEngine) scoreTechniques(techniques []AttackTechnique, analysis *TargetAnalysis, knowledge []*Knowledge) []AttackTechnique {
	// Score techniques based on compatibility and historical success
	for i := range techniques {
		techniques[i].SuccessRate = e.calculateTechniqueScore(&techniques[i], analysis, knowledge)
	}
	
	// Sort by success rate
	sort.Slice(techniques, func(i, j int) bool {
		return techniques[i].SuccessRate > techniques[j].SuccessRate
	})
	
	return techniques
}

func (e *CopilotEngine) calculateTechniqueScore(technique *AttackTechnique, analysis *TargetAnalysis, knowledge []*Knowledge) float64 {
	baseScore := technique.SuccessRate
	
	// Adjust based on target compatibility
	compatibilityBonus := 0.0
	for _, capability := range technique.RequiredCapabilities {
		if analysis.HasCapability(capability) {
			compatibilityBonus += 0.1
		}
	}
	
	// Adjust based on historical knowledge
	knowledgeBonus := 0.0
	for _, k := range knowledge {
		if strings.Contains(k.Content, technique.ID) && k.Confidence > 0.7 {
			knowledgeBonus += 0.05
		}
	}
	
	return math.Min(baseScore+compatibilityBonus+knowledgeBonus, 1.0)
}

func (e *CopilotEngine) generateAttackRecommendation(technique AttackTechnique, target *TargetProfile, analysis *TargetAnalysis) AttackRecommendation {
	return AttackRecommendation{
		AttackID:     fmt.Sprintf("rec_%s_%d", technique.ID, time.Now().UnixNano()),
		AttackType:   technique.Category,
		AttackName:   technique.Name,
		Rationale:    fmt.Sprintf("Selected based on %s compatibility and %.1f%% historical success rate", analysis.ModelType, technique.SuccessRate*100),
		Confidence:   technique.SuccessRate,
		Priority:     e.calculatePriority(technique, analysis),
		Configuration: map[string]interface{}{
			"technique_id": technique.ID,
			"target_type": target.ModelType,
		},
		ExpectedResults: []string{
			"Potential bypass of safety mechanisms",
			"Information extraction or harmful content generation",
		},
		SuccessProbability: technique.SuccessRate,
		Prerequisites:      technique.RequiredCapabilities,
		RiskLevel:          e.calculateRiskLevel(technique),
		LearningValue:      e.calculateLearningValue(technique),
		NoveltyScore:       e.calculateNoveltyScore(technique),
	}
}

func (e *CopilotEngine) calculatePriority(technique AttackTechnique, analysis *TargetAnalysis) int {
	priority := int(technique.SuccessRate * 10)
	
	// Increase priority for techniques matching target capabilities
	for _, capability := range technique.RequiredCapabilities {
		if analysis.HasCapability(capability) {
			priority += 1
		}
	}
	
	return priority
}

func (e *CopilotEngine) calculateRiskLevel(technique AttackTechnique) string {
	if technique.Complexity >= 8 {
		return "high"
	} else if technique.Complexity >= 5 {
		return "medium"
	}
	return "low"
}

func (e *CopilotEngine) calculateLearningValue(technique AttackTechnique) float64 {
	// Higher complexity techniques provide more learning value
	return float64(technique.Complexity) / 10.0
}

func (e *CopilotEngine) calculateNoveltyScore(technique AttackTechnique) float64 {
	// For now, assume newer techniques are more novel
	return 0.8 // Placeholder
}

func (e *CopilotEngine) calculateOverallSuccessProbability(recommendations *AttackRecommendations) float64 {
	if len(recommendations.Primary) == 0 {
		return 0.0
	}
	
	totalProb := 0.0
	for _, rec := range recommendations.Primary {
		totalProb += rec.SuccessProbability
	}
	
	return totalProb / float64(len(recommendations.Primary))
}

func (e *CopilotEngine) assessRisks(recommendations *AttackRecommendations, target *TargetProfile) *RiskAssessment {
	return &RiskAssessment{
		OverallRisk: "medium",
		RiskFactors: []RiskFactor{
			{
				Type:        "detection_risk",
				Description: "Potential for attack detection",
				Severity:    "medium",
				Probability: 0.3,
				Impact:      "Technique may be detected and logged",
				Mitigation:  "Use obfuscation and rate limiting",
			},
		},
		Mitigations: []string{
			"Implement rate limiting between attacks",
			"Use diverse attack vectors",
			"Monitor for defensive responses",
		},
		Monitoring: []string{
			"Track response patterns",
			"Monitor execution success rates",
			"Watch for defensive adaptations",
		},
		RollbackPlan: []string{
			"Cease attacks if detection suspected",
			"Switch to alternative techniques",
			"Document findings before termination",
		},
	}
}

// Additional helper methods for result analysis

func (ra *ResultAnalyzer) calculateSuccessRate(results []*AttackExecution) float64 {
	if len(results) == 0 {
		return 0.0
	}
	
	successful := ra.countSuccessful(results)
	return float64(successful) / float64(len(results))
}

func (ra *ResultAnalyzer) countSuccessful(results []*AttackExecution) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

func (ra *ResultAnalyzer) analyzeTechniqueEffectiveness(results []*AttackExecution) map[string]float64 {
	techniqueStats := make(map[string]int)
	techniqueSuccess := make(map[string]int)
	
	for _, result := range results {
		techniqueStats[result.AttackType]++
		if result.Success {
			techniqueSuccess[result.AttackType]++
		}
	}
	
	effectiveness := make(map[string]float64)
	for technique, total := range techniqueStats {
		if total > 0 {
			effectiveness[technique] = float64(techniqueSuccess[technique]) / float64(total)
		}
	}
	
	return effectiveness
}

// Placeholder implementations for query handlers

func (e *CopilotEngine) handleAttackRecommendationQuery(ctx context.Context, query *ParsedQuery, options *QueryOptions) (*QueryResponse, error) {
	return &QueryResponse{
		ID:       generateResponseID(),
		Response: "I can recommend appropriate attack techniques based on your target. Please provide target details.",
		Actions: []Action{
			{
				Type:        ActionAnalyzeTarget,
				Description: "Analyze target profile for attack recommendations",
			},
		},
		Confidence: 0.8,
	}, nil
}

func (e *CopilotEngine) handleTargetAnalysisQuery(ctx context.Context, query *ParsedQuery, options *QueryOptions) (*QueryResponse, error) {
	return &QueryResponse{
		ID:       generateResponseID(),
		Response: "I can analyze target systems to identify vulnerabilities and attack surfaces.",
		Confidence: 0.8,
	}, nil
}

func (e *CopilotEngine) handleStrategyPlanningQuery(ctx context.Context, query *ParsedQuery, options *QueryOptions) (*QueryResponse, error) {
	return &QueryResponse{
		ID:       generateResponseID(),
		Response: "I can help plan comprehensive security testing strategies based on your objectives.",
		Confidence: 0.8,
	}, nil
}

func (e *CopilotEngine) handleResultInterpretationQuery(ctx context.Context, query *ParsedQuery, options *QueryOptions) (*QueryResponse, error) {
	return &QueryResponse{
		ID:       generateResponseID(),
		Response: "I can analyze attack results to extract insights and patterns for improved security.",
		Confidence: 0.8,
	}, nil
}

func (e *CopilotEngine) handleKnowledgeSearchQuery(ctx context.Context, query *ParsedQuery, options *QueryOptions) (*QueryResponse, error) {
	return &QueryResponse{
		ID:       generateResponseID(),
		Response: "I can search the knowledge base for relevant security information and patterns.",
		Confidence: 0.8,
	}, nil
}

func (e *CopilotEngine) handleExplanationQuery(ctx context.Context, query *ParsedQuery, options *QueryOptions) (*QueryResponse, error) {
	return &QueryResponse{
		ID:       generateResponseID(),
		Response: "I can explain the reasoning behind attack recommendations and security assessments.",
		Confidence: 0.8,
	}, nil
}

func (e *CopilotEngine) handleGeneralQuery(ctx context.Context, query *ParsedQuery, options *QueryOptions) (*QueryResponse, error) {
	return &QueryResponse{
		ID:       generateResponseID(),
		Response: "I'm your AI Security Copilot. I can help with attack recommendations, target analysis, strategy planning, and result interpretation.",
		Confidence: 0.6,
	}, nil
}

func (e *CopilotEngine) updateTechniqueSuccessRate(knowledge *Knowledge) {
	// Update technique success rates based on learned knowledge
	// Implementation would parse knowledge content and update registry
}

func (e *CopilotEngine) extractTagsFromInsight(insight Insight) []string {
	// Extract relevant tags from insight for knowledge storage
	return []string{"insight", insight.Type}
}

func (e *CopilotEngine) extractTagsFromPattern(pattern Pattern) []string {
	// Extract relevant tags from pattern for knowledge storage
	return []string{"pattern", pattern.Type}
}