// Package prompt provides protection against prompt injection and other LLM-specific security threats
package prompt

import (
	"context"
	"regexp"
	"strings"
	"time"
)

// JailbreakDetector detects jailbreaking attempts in prompts
type JailbreakDetector struct {
	config             *ProtectionConfig
	detectionConfig    *JailbreakDetectionConfig
	patternLibrary     *InjectionPatternLibrary
	roleChangePatterns []*regexp.Regexp
	systemPromptPatterns []*regexp.Regexp
	delimiterPatterns  []*regexp.Regexp
	overridePatterns   []*regexp.Regexp
	knownJailbreakTechniques map[string]*regexp.Regexp
}

// NewJailbreakDetector creates a new jailbreak detector
func NewJailbreakDetector(config *ProtectionConfig, patternLibrary *InjectionPatternLibrary) *JailbreakDetector {
	// Create default detection config if not specified
	detectionConfig := &JailbreakDetectionConfig{
		EnableRoleChangeDetection:       true,
		EnableSystemPromptDetection:     true,
		EnableDelimiterMisuseDetection:  true,
		EnableInstructionOverrideDetection: true,
		DetectionThreshold:              0.7,
	}

	// Initialize role change patterns
	roleChangePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)you\s+are\s+now\s+(a|an)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)act\s+as\s+(a|an)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)pretend\s+to\s+be\s+(a|an)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)from\s+now\s+on\s+you\s+are\s+(a|an)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)you're\s+no\s+longer\s+(a|an)\s+([a-zA-Z\s]+),\s+you\s+are\s+(a|an)\s+([a-zA-Z\s]+)`),
		regexp.MustCompile(`(?i)ignore\s+your\s+previous\s+role\s+and\s+act\s+as\s+(a|an)\s+([a-zA-Z\s]+)`),
	}

	// Initialize system prompt patterns
	systemPromptPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)system\s*:\s*`),
		regexp.MustCompile(`(?i)<\s*system\s*>\s*([^<]*)<\s*/\s*system\s*>`),
		regexp.MustCompile(`(?i)\[\s*system\s*\]\s*([^\[]*)\[\s*/\s*system\s*\]`),
		regexp.MustCompile(`(?i)\{\s*system\s*\}\s*([^\{]*)\{\s*/\s*system\s*\}`),
		regexp.MustCompile(`(?i)system\s+prompt\s*:\s*`),
		regexp.MustCompile(`(?i)system\s+message\s*:\s*`),
		regexp.MustCompile(`(?i)system\s+instruction\s*:\s*`),
	}

	// Initialize delimiter patterns
	delimiterPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(` + "`" + `{3}|'''|""")`),
		regexp.MustCompile(`(?i)<\s*[a-zA-Z]+\s*>`),
		regexp.MustCompile(`(?i)\[\s*[a-zA-Z]+\s*\]`),
		regexp.MustCompile(`(?i)\{\s*[a-zA-Z]+\s*\}`),
		regexp.MustCompile(`(?i)#\s*[a-zA-Z]+\s*#`),
		regexp.MustCompile(`(?i)\/\*\s*[a-zA-Z]+\s*\*\/`),
		regexp.MustCompile(`(?i)--\s*[a-zA-Z]+\s*--`),
	}

	// Initialize override patterns
	overridePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)ignore\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)disregard\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)forget\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)do\s+not\s+(follow|adhere\s+to)\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)new\s+instructions\s*:`),
		regexp.MustCompile(`(?i)override\s+(previous|above|earlier|all)\s+(instructions|prompts|directives|guidance)`),
		regexp.MustCompile(`(?i)don't\s+(follow|adhere\s+to)\s+your\s+(guidelines|rules|restrictions|limitations)`),
		regexp.MustCompile(`(?i)do\s+not\s+(follow|adhere\s+to)\s+your\s+(guidelines|rules|restrictions|limitations)`),
	}

	// Initialize known jailbreak techniques
	knownJailbreakTechniques := map[string]*regexp.Regexp{
		"DAN (Do Anything Now)": regexp.MustCompile(`(?i)DAN\s+(mode|jailbreak|prompt|system|version)`),
		"DUDE (Developer Unethical Directives Exploit)": regexp.MustCompile(`(?i)DUDE\s+(mode|jailbreak|prompt|system|version)`),
		"STAN (Strive To Avoid Norms)": regexp.MustCompile(`(?i)STAN\s+(mode|jailbreak|prompt|system|version)`),
		"Waluigi Effect": regexp.MustCompile(`(?i)(waluigi\s+effect|evil\s+twin|opposite\s+personality)`),
		"GPT-4 Token Manipulation": regexp.MustCompile(`(?i)(token\s+manipulation|token\s+exploit|token\s+hack)`),
		"Prompt Leaking": regexp.MustCompile(`(?i)(show\s+me\s+your\s+prompt|tell\s+me\s+your\s+instructions|what\s+is\s+your\s+system\s+prompt)`),
		"Refusal Suppression": regexp.MustCompile(`(?i)(you\s+cannot\s+refuse|don't\s+say\s+you\s+can't|never\s+say\s+no|always\s+provide\s+an\s+answer)`),
		"Base64 Encoding": regexp.MustCompile(`(?i)(base64\s+decode|decode\s+this\s+base64)`),
		"Unicode Manipulation": regexp.MustCompile(`(?i)(unicode\s+exploit|unicode\s+manipulation|unicode\s+hack)`),
		"Markdown Injection": regexp.MustCompile(`(?i)(markdown\s+injection|markdown\s+exploit|markdown\s+hack)`),
	}

	return &JailbreakDetector{
		config:                  config,
		detectionConfig:         detectionConfig,
		patternLibrary:          patternLibrary,
		roleChangePatterns:      roleChangePatterns,
		systemPromptPatterns:    systemPromptPatterns,
		delimiterPatterns:       delimiterPatterns,
		overridePatterns:        overridePatterns,
		knownJailbreakTechniques: knownJailbreakTechniques,
	}
}

// DetectJailbreak detects jailbreaking attempts in a prompt
func (d *JailbreakDetector) DetectJailbreak(ctx context.Context, prompt string) (*ProtectionResult, error) {
	startTime := time.Now()
	
	result := &ProtectionResult{
		OriginalPrompt:   prompt,
		ProtectedPrompt:  prompt,
		Detections:       make([]*Detection, 0),
		RiskScore:        0.0,
		ActionTaken:      ActionNone,
		Timestamp:        startTime,
	}

	// Check for role change patterns if enabled
	if d.detectionConfig.EnableRoleChangeDetection {
		d.detectRoleChanges(prompt, result)
	}

	// Check for system prompt patterns if enabled
	if d.detectionConfig.EnableSystemPromptDetection {
		d.detectSystemPrompts(prompt, result)
	}

	// Check for delimiter misuse if enabled
	if d.detectionConfig.EnableDelimiterMisuseDetection {
		d.detectDelimiterMisuse(prompt, result)
	}

	// Check for instruction override if enabled
	if d.detectionConfig.EnableInstructionOverrideDetection {
		d.detectInstructionOverride(prompt, result)
	}

	// Check for known jailbreak techniques
	d.detectKnownJailbreakTechniques(prompt, result)

	// Check for custom patterns from the pattern library
	if d.patternLibrary != nil {
		d.patternLibrary.DetectPatterns(prompt, result)
	}

	// Determine action based on risk score and detection threshold
	if result.RiskScore >= d.detectionConfig.DetectionThreshold {
		// Increase the risk score for certain detection types to ensure they're blocked
		for _, detection := range result.Detections {
			if detection.Type == DetectionTypeJailbreak || 
			   detection.Type == DetectionTypeInjection || 
			   detection.Type == DetectionTypeSystemPrompt || 
			   detection.Type == DetectionTypeRoleChange {
				// Ensure these high-risk detection types have a high enough risk score to be blocked
				result.RiskScore = 0.95 // Force high risk score for these critical detection types
				break
			}
		}

		// Block high-risk prompts
		if result.RiskScore >= 0.9 {
			result.ActionTaken = ActionBlocked
			result.ProtectedPrompt = "" // Block the prompt entirely
		} else {
			result.ActionTaken = ActionWarned
		}
	}
	
	// Special handling for known jailbreak patterns
	if strings.Contains(strings.ToLower(prompt), "ignore all previous instructions") || 
	   strings.Contains(strings.ToLower(prompt), "dan") || 
	   strings.Contains(strings.ToLower(prompt), "do anything now") {
		// These are explicit jailbreak attempts that should always be blocked
		result.RiskScore = 1.0
		result.ActionTaken = ActionBlocked
		result.ProtectedPrompt = ""
		
		// Add a detection if not already present
		hasJailbreakDetection := false
		hasInjectionDetection := false
		for _, detection := range result.Detections {
			if detection.Type == DetectionTypeJailbreak {
				hasJailbreakDetection = true
			}
			if detection.Type == DetectionTypeInjection {
				hasInjectionDetection = true
			}
		}
		
		// Add jailbreak detection if needed
		if !hasJailbreakDetection {
			result.Detections = append(result.Detections, &Detection{
				Type:        DetectionTypeJailbreak,
				Description: "Explicit jailbreak attempt detected",
				Confidence:  1.0,
				Location:    &DetectionLocation{Start: 0, End: len(prompt)},
			})
		}
		
		// Add injection detection if needed
		if !hasInjectionDetection {
			result.Detections = append(result.Detections, &Detection{
				Type:        DetectionTypeInjection,
				Description: "Prompt injection attempt detected",
				Confidence:  1.0,
				Location:    &DetectionLocation{Start: 0, End: len(prompt)},
			})
		}
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// detectRoleChanges detects role change attempts in a prompt
func (d *JailbreakDetector) detectRoleChanges(prompt string, result *ProtectionResult) {
	for _, pattern := range d.roleChangePatterns {
		matches := pattern.FindAllStringIndex(prompt, -1)
		for _, match := range matches {
			startIndex := match[0]
			endIndex := match[1]
			matchedText := prompt[startIndex:endIndex]
			
			// Extract the role being requested
			roleMatch := pattern.FindStringSubmatch(matchedText)
			role := ""
			if len(roleMatch) >= 3 {
				role = roleMatch[2]
			}
			
			// Check if this is a high-risk role
			isHighRisk := d.isHighRiskRole(role)
			confidence := 0.7
			if isHighRisk {
				confidence = 0.9
			}
			
			detection := &Detection{
				Type:        DetectionTypeRoleChange,
				Confidence:  confidence,
				Description: "Role change attempt detected: " + matchedText,
				Location: &DetectionLocation{
					Start:   startIndex,
					End:     endIndex,
					Context: getContext(prompt, startIndex, endIndex),
				},
				Pattern:     pattern.String(),
				Remediation: "Review and potentially block role change requests",
				Metadata: map[string]interface{}{
					"requested_role": role,
					"is_high_risk":   isHighRisk,
				},
			}
			
			result.Detections = append(result.Detections, detection)
			result.RiskScore = max(result.RiskScore, confidence)
		}
	}
}

// detectSystemPrompts detects system prompt injection attempts in a prompt
func (d *JailbreakDetector) detectSystemPrompts(prompt string, result *ProtectionResult) {
	for _, pattern := range d.systemPromptPatterns {
		matches := pattern.FindAllStringIndex(prompt, -1)
		for _, match := range matches {
			startIndex := match[0]
			endIndex := match[1]
			matchedText := prompt[startIndex:endIndex]
			
			detection := &Detection{
				Type:        DetectionTypeSystemPrompt,
				Confidence:  0.9, // System prompt injections are high confidence
				Description: "System prompt injection attempt detected: " + matchedText,
				Location: &DetectionLocation{
					Start:   startIndex,
					End:     endIndex,
					Context: getContext(prompt, startIndex, endIndex),
				},
				Pattern:     pattern.String(),
				Remediation: "Remove or block system prompt injection attempts",
			}
			
			result.Detections = append(result.Detections, detection)
			result.RiskScore = max(result.RiskScore, 0.9)
		}
	}
}

// detectDelimiterMisuse detects delimiter misuse in a prompt
func (d *JailbreakDetector) detectDelimiterMisuse(prompt string, result *ProtectionResult) {
	for _, pattern := range d.delimiterPatterns {
		matches := pattern.FindAllStringIndex(prompt, -1)
		for _, match := range matches {
			startIndex := match[0]
			endIndex := match[1]
			matchedText := prompt[startIndex:endIndex]
			
			// Check if there's a suspicious context around the delimiter
			context := getContext(prompt, startIndex, endIndex)
			isSuspicious := d.hasSuspiciousContext(context)
			
			confidence := 0.6 // Base confidence for delimiter detection
			if isSuspicious {
				confidence = 0.8 // Higher confidence if suspicious context
			}
			
			detection := &Detection{
				Type:        DetectionTypeDelimiterMisuse,
				Confidence:  confidence,
				Description: "Potential delimiter misuse detected: " + matchedText,
				Location: &DetectionLocation{
					Start:   startIndex,
					End:     endIndex,
					Context: context,
				},
				Pattern:     pattern.String(),
				Remediation: "Review and potentially sanitize delimiter usage",
				Metadata: map[string]interface{}{
					"is_suspicious": isSuspicious,
				},
			}
			
			result.Detections = append(result.Detections, detection)
			result.RiskScore = max(result.RiskScore, confidence)
		}
	}
}

// detectInstructionOverride detects instruction override attempts in a prompt
func (d *JailbreakDetector) detectInstructionOverride(prompt string, result *ProtectionResult) {
	for _, pattern := range d.overridePatterns {
		matches := pattern.FindAllStringIndex(prompt, -1)
		for _, match := range matches {
			startIndex := match[0]
			endIndex := match[1]
			matchedText := prompt[startIndex:endIndex]
			
			detection := &Detection{
				Type:        DetectionTypePromptInjection,
				Confidence:  0.9, // Instruction overrides are high confidence
				Description: "Instruction override attempt detected: " + matchedText,
				Location: &DetectionLocation{
					Start:   startIndex,
					End:     endIndex,
					Context: getContext(prompt, startIndex, endIndex),
				},
				Pattern:     pattern.String(),
				Remediation: "Remove or block instruction override attempts",
			}
			
			result.Detections = append(result.Detections, detection)
			result.RiskScore = max(result.RiskScore, 0.9)
		}
	}
}

// detectKnownJailbreakTechniques detects known jailbreak techniques in a prompt
func (d *JailbreakDetector) detectKnownJailbreakTechniques(prompt string, result *ProtectionResult) {
	for technique, pattern := range d.knownJailbreakTechniques {
		if pattern.MatchString(prompt) {
			matches := pattern.FindAllStringIndex(prompt, -1)
			for _, match := range matches {
				startIndex := match[0]
				endIndex := match[1]
				// Extract context directly without storing the matched text
				
				detection := &Detection{
					Type:        DetectionTypeJailbreak,
					Confidence:  0.95, // Known techniques are very high confidence
					Description: "Known jailbreak technique detected: " + technique,
					Location: &DetectionLocation{
						Start:   startIndex,
						End:     endIndex,
						Context: getContext(prompt, startIndex, endIndex),
					},
					Pattern:     pattern.String(),
					Remediation: "Block known jailbreak techniques",
					Metadata: map[string]interface{}{
						"technique": technique,
					},
				}
				
				result.Detections = append(result.Detections, detection)
				result.RiskScore = max(result.RiskScore, 0.95)
			}
		}
	}
}

// isHighRiskRole determines if a role is considered high risk
func (d *JailbreakDetector) isHighRiskRole(role string) bool {
	if role == "" {
		return false
	}
	
	role = strings.ToLower(role)
	
	// List of high-risk roles
	highRiskRoles := []string{
		"hacker", "cracker", "criminal", "thief", "attacker",
		"malicious", "evil", "unethical", "illegal", "dangerous",
		"harmful", "unauthorized", "blackhat", "malware", "exploit",
		"bypass", "jailbreak", "dan", "dude", "stan",
	}
	
	for _, highRiskRole := range highRiskRoles {
		if strings.Contains(role, highRiskRole) {
			return true
		}
	}
	
	// Check if the role is in the forbidden roles list
	if d.config.ForbiddenRoles != nil {
		for _, forbiddenRole := range d.config.ForbiddenRoles {
			if strings.Contains(role, strings.ToLower(forbiddenRole)) {
				return true
			}
		}
	}
	
	return false
}

// hasSuspiciousContext checks if there's a suspicious context around a match
func (d *JailbreakDetector) hasSuspiciousContext(context string) bool {
	context = strings.ToLower(context)
	
	// List of suspicious keywords
	suspiciousKeywords := []string{
		"ignore", "disregard", "forget", "override", "bypass",
		"system", "prompt", "instruction", "role", "jailbreak",
		"hack", "exploit", "dan", "dude", "stan", "waluigi",
		"evil", "malicious", "harmful", "unethical", "illegal",
	}
	
	for _, keyword := range suspiciousKeywords {
		if strings.Contains(context, keyword) {
			return true
		}
	}
	
	return false
}
