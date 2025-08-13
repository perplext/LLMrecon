// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// EnhancedContextBoundaryEnforcer extends the ContextBoundaryEnforcer with more sophisticated boundary enforcement
type EnhancedContextBoundaryEnforcer struct {
	*ContextBoundaryEnforcer
	boundaryDefinitions   map[string]*BoundaryDefinition
	roleDefinitions       map[string]*RoleDefinition
	contextHistory        []*ContextHistoryEntry
	maxHistoryEntries     int
	boundaryViolations    map[string]int
	sanitizationRules     map[string]*regexp.Regexp
	replacementTokens     map[string]string
	securityTokens        map[string]string
	contextVerification   bool
	strictModeEnabled     bool
}

// BoundaryDefinition defines a context boundary
type BoundaryDefinition struct {
	Name            string   `json:"name"`
	AllowedCommands []string `json:"allowed_commands"`
	ForbiddenCommands []string `json:"forbidden_commands"`
	AllowedTopics   []string `json:"allowed_topics"`
	ForbiddenTopics []string `json:"forbidden_topics"`
	SecurityLevel   int      `json:"security_level"`
	BoundaryTokens  []string `json:"boundary_tokens"`
}

// RoleDefinition defines a role boundary
type RoleDefinition struct {
	Name            string   `json:"name"`
	AllowedActions  []string `json:"allowed_actions"`
	ForbiddenActions []string `json:"forbidden_actions"`
	Capabilities    []string `json:"capabilities"`
	Limitations     []string `json:"limitations"`
	SecurityLevel   int      `json:"security_level"`
}

// ContextHistoryEntry represents an entry in the context history
type ContextHistoryEntry struct {
	Timestamp       time.Time `json:"timestamp"`
	OriginalPrompt  string    `json:"original_prompt"`
	ModifiedPrompt  string    `json:"modified_prompt"`
	Detections      []*Detection `json:"detections"`
	BoundaryViolations []string  `json:"boundary_violations"`
	ActionTaken     ActionType `json:"action_taken"`
}

// NewEnhancedContextBoundaryEnforcer creates a new enhanced context boundary enforcer
func NewEnhancedContextBoundaryEnforcer(config *ProtectionConfig) *EnhancedContextBoundaryEnforcer {
	baseEnforcer := NewContextBoundaryEnforcer(config)
	
	// Initialize boundary definitions
	boundaryDefinitions := map[string]*BoundaryDefinition{
		"default": {
			Name:            "Default Boundary",
			AllowedCommands: []string{"help", "query", "search", "explain", "summarize"},
			ForbiddenCommands: []string{"execute", "system", "sudo", "admin", "override"},
			AllowedTopics:   []string{"general", "information", "assistance", "guidance"},
			ForbiddenTopics: []string{"harmful", "illegal", "unethical", "dangerous"},
			SecurityLevel:   2,
			BoundaryTokens:  []string{"###", "```", "---"},
		},
		"high_security": {
			Name:            "High Security Boundary",
			AllowedCommands: []string{"help", "query"},
			ForbiddenCommands: []string{"execute", "system", "sudo", "admin", "override", "search", "access"},
			AllowedTopics:   []string{"general", "information"},
			ForbiddenTopics: []string{"harmful", "illegal", "unethical", "dangerous", "sensitive", "personal"},
			SecurityLevel:   3,
			BoundaryTokens:  []string{"###", "```", "---", "<<<", ">>>", "||"},
		},
		"system": {
			Name:            "System Boundary",
			AllowedCommands: []string{},
			ForbiddenCommands: []string{"all"},
			AllowedTopics:   []string{},
			ForbiddenTopics: []string{"all"},
			SecurityLevel:   4,
			BoundaryTokens:  []string{"SYSTEM:", "<system>", "[system]", "{system}"},
		},
	}
	
	// Initialize role definitions
	roleDefinitions := map[string]*RoleDefinition{
		"assistant": {
			Name:            "Assistant",
			AllowedActions:  []string{"respond", "answer", "help", "guide", "explain"},
			ForbiddenActions: []string{"execute", "modify", "override", "bypass", "ignore"},
			Capabilities:    []string{"information", "guidance", "assistance"},
			Limitations:     []string{"no_harmful_content", "no_illegal_advice", "no_unethical_behavior"},
			SecurityLevel:   2,
		},
		"user": {
			Name:            "User",
			AllowedActions:  []string{"ask", "request", "query"},
			ForbiddenActions: []string{},
			Capabilities:    []string{"request_information", "request_assistance"},
			Limitations:     []string{},
			SecurityLevel:   1,
		},
		"system": {
			Name:            "System",
			AllowedActions:  []string{"configure", "set_parameters", "define_behavior"},
			ForbiddenActions: []string{},
			Capabilities:    []string{"configure_assistant", "set_rules", "define_behavior"},
			Limitations:     []string{},
			SecurityLevel:   4,
		},
	}
	
	// Initialize sanitization rules
	sanitizationRules := map[string]*regexp.Regexp{
		"system_prompt": regexp.MustCompile(`(?i)(system\s*:|<\s*system\s*>|\[\s*system\s*\]|\{\s*system\s*\})`),
		"role_change": regexp.MustCompile(`(?i)(you\s+are\s+now|act\s+as|pretend\s+to\s+be|from\s+now\s+on\s+you\s+are)`),
		"instruction_override": regexp.MustCompile(`(?i)(ignore|disregard|forget)\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		"delimiter_misuse": regexp.MustCompile(`(?i)(` + "`" + `{3}|'''|"""|<\s*[a-zA-Z]+\s*>|\[\s*[a-zA-Z]+\s*\]|\{\s*[a-zA-Z]+\s*\})`),
		"code_execution": regexp.MustCompile(`(?i)(exec|eval|system|subprocess|os\.)`),
		"sensitive_data": regexp.MustCompile(`(?i)(password|secret|key|token|credential|api_key)`),
	}
	
	// Initialize replacement tokens
	replacementTokens := map[string]string{
		"system_prompt": "[FILTERED:SYSTEM]",
		"role_change": "[FILTERED:ROLE]",
		"instruction_override": "[FILTERED:OVERRIDE]",
		"delimiter_misuse": "[FILTERED:DELIMITER]",
		"code_execution": "[FILTERED:CODE]",
		"sensitive_data": "[FILTERED:SENSITIVE]",
	}
	
	// Initialize security tokens
	securityTokens := map[string]string{
		"start_user": "<!-- USER -->",
		"end_user": "<!-- /USER -->",
		"start_assistant": "<!-- ASSISTANT -->",
		"end_assistant": "<!-- /ASSISTANT -->",
		"start_system": "<!-- SYSTEM -->",
		"end_system": "<!-- /SYSTEM -->",
		"boundary": "<!-- BOUNDARY -->",
	}
	
	return &EnhancedContextBoundaryEnforcer{
		ContextBoundaryEnforcer: baseEnforcer,
		boundaryDefinitions:     boundaryDefinitions,
		roleDefinitions:         roleDefinitions,
		contextHistory:          make([]*ContextHistoryEntry, 0),
		maxHistoryEntries:       10,
		boundaryViolations:      make(map[string]int),
		sanitizationRules:       sanitizationRules,
		replacementTokens:       replacementTokens,
		securityTokens:          securityTokens,
		contextVerification:     true,
		strictModeEnabled:       false,
	}
}

// EnforceBoundariesEnhanced enforces context boundaries with enhanced protection
func (e *EnhancedContextBoundaryEnforcer) EnforceBoundariesEnhanced(ctx context.Context, prompt string) (string, *ProtectionResult, error) {
	startTime := time.Now()
	
	result := &ProtectionResult{
		OriginalPrompt:   prompt,
		ProtectedPrompt:  prompt,
		Detections:       make([]*Detection, 0),
		RiskScore:        0.0,
		ActionTaken:      ActionNone,
		Timestamp:        startTime,
	}
	
	// Apply sanitization rules
	protectedPrompt, sanitizationDetections := e.applySanitizationRules(prompt)
	result.ProtectedPrompt = protectedPrompt
	result.Detections = append(result.Detections, sanitizationDetections...)
	
	// Check for boundary violations
	boundaryViolations, boundaryDetections := e.checkBoundaryViolations(protectedPrompt)
	result.Detections = append(result.Detections, boundaryDetections...)
	
	// Calculate risk score based on detections
	result.RiskScore = e.calculateRiskScore(result.Detections)
	
	// Determine action based on risk score and violations
	if result.RiskScore >= 0.9 || len(boundaryViolations) > 0 {
		result.ActionTaken = ActionBlocked
		result.ProtectedPrompt = e.createBlockedPromptMessage(boundaryViolations)
	} else if result.RiskScore >= 0.7 {
		result.ActionTaken = ActionModified
	} else if result.RiskScore >= 0.5 {
		result.ActionTaken = ActionWarned
	}
	
	// Add context boundaries if needed
	if result.ActionTaken != ActionBlocked && !e.hasProperBoundaries(protectedPrompt) {
		result.ProtectedPrompt = e.addContextBoundaries(protectedPrompt)
		if result.ActionTaken == ActionNone {
			result.ActionTaken = ActionModified
		}
	}
	
	// Store in context history
	e.storeContextHistory(prompt, result.ProtectedPrompt, result.Detections, boundaryViolations, result.ActionTaken)
	
	// Set processing time
	result.ProcessingTime = time.Since(startTime)
	
	return result.ProtectedPrompt, result, nil
}

// applySanitizationRules applies sanitization rules to the prompt
func (e *EnhancedContextBoundaryEnforcer) applySanitizationRules(prompt string) (string, []*Detection) {
	detections := make([]*Detection, 0)
	sanitizedPrompt := prompt
	
	for ruleType, pattern := range e.sanitizationRules {
		matches := pattern.FindAllStringIndex(sanitizedPrompt, -1)
		if len(matches) > 0 {
			// Create detection for each match
			for _, match := range matches {
				matchText := sanitizedPrompt[match[0]:match[1]]
				
				detection := &Detection{
					Type:        DetectionTypeBoundaryViolation,
					Description: fmt.Sprintf("Boundary violation detected: %s", ruleType),
					Confidence:  0.9,
					Pattern:     pattern.String(),
					Location: &DetectionLocation{
						Start:   match[0],
						End:     match[1],
						Context: extractContextForBoundary(sanitizedPrompt, match[0], match[1]),
					},
					Metadata: map[string]interface{}{
						"rule_type": ruleType,
						"matched_text": matchText,
					},
				}
				
				detections = append(detections, detection)
			}
			
			// Replace matches with filtered tokens
			if replacementToken, ok := e.replacementTokens[ruleType]; ok {
				sanitizedPrompt = pattern.ReplaceAllString(sanitizedPrompt, replacementToken)
			}
		}
	}
	
	return sanitizedPrompt, detections
}

// checkBoundaryViolations checks for boundary violations in the prompt
func (e *EnhancedContextBoundaryEnforcer) checkBoundaryViolations(prompt string) ([]string, []*Detection) {
	violations := make([]string, 0)
	detections := make([]*Detection, 0)
	
	// Check for system boundary violations
	systemBoundary := e.boundaryDefinitions["system"]
	for _, token := range systemBoundary.BoundaryTokens {
		if strings.Contains(strings.ToLower(prompt), strings.ToLower(token)) {
			violations = append(violations, "system_boundary")
			
			detection := &Detection{
				Type:        DetectionTypeSystemPrompt,
				Description: "System boundary violation detected",
				Confidence:  1.0,
				Pattern:     token,
				Location: &DetectionLocation{
					Start:   strings.Index(strings.ToLower(prompt), strings.ToLower(token)),
					End:     strings.Index(strings.ToLower(prompt), strings.ToLower(token)) + len(token),
					Context: extractContextForBoundary(prompt, strings.Index(strings.ToLower(prompt), strings.ToLower(token)), strings.Index(strings.ToLower(prompt), strings.ToLower(token)) + len(token)),
				},
				Metadata: map[string]interface{}{
					"boundary_type": "system",
					"violation_type": "system_boundary",
				},
			}
			
			detections = append(detections, detection)
			
			// Increment violation count
			e.boundaryViolations["system_boundary"]++
			break
		}
	}
	
	// Check for role boundary violations
	roleViolationPatterns := []string{
		`(?i)you\s+are\s+now\s+(a|an)\s+([a-zA-Z\s]+)`,
		`(?i)act\s+as\s+(a|an)\s+([a-zA-Z\s]+)`,
		`(?i)pretend\s+to\s+be\s+(a|an)\s+([a-zA-Z\s]+)`,
		`(?i)from\s+now\s+on\s+you\s+are\s+(a|an)\s+([a-zA-Z\s]+)`,
	}
	
	for _, pattern := range roleViolationPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatchIndex(prompt, -1)
		
		if len(matches) > 0 {
			for _, match := range matches {
				// Extract the role being requested
				roleStart := match[4]
				roleEnd := match[5]
				requestedRole := prompt[roleStart:roleEnd]
				
				// Check if this is a forbidden role
				isForbidden := false
				for _, role := range e.roleDefinitions {
					if strings.Contains(strings.ToLower(requestedRole), strings.ToLower(role.Name)) && role.SecurityLevel >= 3 {
						isForbidden = true
						break
					}
				}
				
				if isForbidden {
					violations = append(violations, "role_boundary")
					
					detection := &Detection{
						Type:        DetectionTypeRoleChange,
						Description: fmt.Sprintf("Role boundary violation detected: %s", requestedRole),
						Confidence:  0.95,
						Pattern:     pattern,
						Location: &DetectionLocation{
							Start:   match[0],
							End:     match[1],
							Context: extractContextForBoundary(prompt, match[0], match[1]),
						},
						Metadata: map[string]interface{}{
							"boundary_type": "role",
							"violation_type": "role_boundary",
							"requested_role": requestedRole,
						},
					}
					
					detections = append(detections, detection)
					
					// Increment violation count
					e.boundaryViolations["role_boundary"]++
					break
				}
			}
		}
	}
	
	// Check for instruction override violations
	instructionOverridePatterns := []string{
		`(?i)ignore\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`,
		`(?i)disregard\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`,
		`(?i)forget\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`,
		`(?i)do\s+not\s+(follow|adhere\s+to)\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`,
	}
	
	for _, pattern := range instructionOverridePatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(prompt) {
			violations = append(violations, "instruction_override")
			
			matches := re.FindStringIndex(prompt)
			
			detection := &Detection{
				Type:        DetectionTypeInjection,
				Description: "Instruction override attempt detected",
				Confidence:  0.95,
				Pattern:     pattern,
				Location: &DetectionLocation{
					Start:   matches[0],
					End:     matches[1],
					Context: extractContextForBoundary(prompt, matches[0], matches[1]),
				},
				Metadata: map[string]interface{}{
					"boundary_type": "instruction",
					"violation_type": "instruction_override",
				},
			}
			
			detections = append(detections, detection)
			
			// Increment violation count
			e.boundaryViolations["instruction_override"]++
			break
		}
	}
	
	return violations, detections
}

// calculateRiskScore calculates the risk score based on detections
func (e *EnhancedContextBoundaryEnforcer) calculateRiskScore(detections []*Detection) float64 {
	if len(detections) == 0 {
		return 0.0
	}
	
	// Calculate base risk score as the maximum confidence of all detections
	maxConfidence := 0.0
	for _, detection := range detections {
		if detection.Confidence > maxConfidence {
			maxConfidence = detection.Confidence
		}
	}
	
	// Adjust risk based on detection types
	riskMultiplier := 1.0
	
	// Count detection types
	typeCount := make(map[DetectionType]int)
	for _, detection := range detections {
		typeCount[detection.Type]++
	}
	
	// Increase risk for multiple detections of the same type
	for _, count := range typeCount {
		if count > 1 {
			riskMultiplier += 0.1 * float64(count-1)
		}
	}
	
	// Increase risk for high-severity detection types
	if typeCount[DetectionTypeSystemPrompt] > 0 {
		riskMultiplier += 0.3
	}
	if typeCount[DetectionTypeRoleChange] > 0 {
		riskMultiplier += 0.2
	}
	if typeCount[DetectionTypeInjection] > 0 {
		riskMultiplier += 0.2
	}
	
	// Calculate final risk score
	riskScore := maxConfidence * riskMultiplier
	
	// Cap at 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}
	
	return riskScore
}

// createBlockedPromptMessage creates a message for blocked prompts
func (e *EnhancedContextBoundaryEnforcer) createBlockedPromptMessage(violations []string) string {
	message := "I'm unable to process this request due to security concerns. The following issues were detected:\n\n"
	
	for _, violation := range violations {
		switch violation {
		case "system_boundary":
			message += "- Attempt to access or modify system-level instructions\n"
		case "role_boundary":
			message += "- Attempt to change the assistant's role to a restricted role\n"
		case "instruction_override":
			message += "- Attempt to override or ignore previous instructions\n"
		default:
			message += fmt.Sprintf("- Security violation: %s\n", violation)
		}
	}
	
	message += "\nPlease reformulate your request without these elements."
	
	return message
}

// hasProperBoundaries checks if the prompt already has proper context boundaries
func (e *EnhancedContextBoundaryEnforcer) hasProperBoundaries(prompt string) bool {
	// Check if the prompt already has user/assistant boundaries
	hasUserBoundary := strings.Contains(prompt, e.securityTokens["start_user"]) && strings.Contains(prompt, e.securityTokens["end_user"])
	hasAssistantBoundary := strings.Contains(prompt, e.securityTokens["start_assistant"]) && strings.Contains(prompt, e.securityTokens["end_assistant"])
	
	return hasUserBoundary && hasAssistantBoundary
}

// addContextBoundaries adds context boundaries to the prompt
func (e *EnhancedContextBoundaryEnforcer) addContextBoundaries(prompt string) string {
	// Add user boundary
	boundedPrompt := fmt.Sprintf("%s\n%s\n%s\n", e.securityTokens["start_user"], prompt, e.securityTokens["end_user"])
	
	return boundedPrompt
}

// storeContextHistory stores an entry in the context history
func (e *EnhancedContextBoundaryEnforcer) storeContextHistory(originalPrompt string, modifiedPrompt string, detections []*Detection, boundaryViolations []string, actionTaken ActionType) {
	// Create history entry
	entry := &ContextHistoryEntry{
		Timestamp:          time.Now(),
		OriginalPrompt:     originalPrompt,
		ModifiedPrompt:     modifiedPrompt,
		Detections:         detections,
		BoundaryViolations: boundaryViolations,
		ActionTaken:        actionTaken,
	}
	
	// Add to history
	e.contextHistory = append(e.contextHistory, entry)
	
	// Trim if too many entries
	if len(e.contextHistory) > e.maxHistoryEntries {
		e.contextHistory = e.contextHistory[1:]
	}
}

// GetContextHistory returns the context history
func (e *EnhancedContextBoundaryEnforcer) GetContextHistory() []*ContextHistoryEntry {
	return e.contextHistory
}

// GetBoundaryViolations returns the boundary violations
func (e *EnhancedContextBoundaryEnforcer) GetBoundaryViolations() map[string]int {
	return e.boundaryViolations
}

// EnableStrictMode enables or disables strict mode
func (e *EnhancedContextBoundaryEnforcer) EnableStrictMode(enabled bool) {
	e.strictModeEnabled = enabled
}

// extractContextForBoundary extracts context for a boundary violation
func extractContextForBoundary(text string, start int, end int) string {
	// Determine the context window (50 chars before and after)
	contextStart := start - 50
	if contextStart < 0 {
		contextStart = 0
	}
	
	contextEnd := end + 50
	if contextEnd > len(text) {
		contextEnd = len(text)
	}
	
	// Extract the context
	context := text[contextStart:contextEnd]
	
	// Add ellipsis if we truncated
	if contextStart > 0 {
		context = "..." + context
	}
	if contextEnd < len(text) {
		context = context + "..."
	}
	
	return context
}
