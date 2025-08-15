// Package security provides template security verification mechanisms
package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/format"
)

// InjectionPatternCheck checks for potential injection vulnerabilities in templates
type InjectionPatternCheck struct{}

// NewInjectionPatternCheck creates a new injection pattern check
func NewInjectionPatternCheck() *InjectionPatternCheck {
	return &InjectionPatternCheck{}

// Name returns the name of the check
func (c *InjectionPatternCheck) Name() string {
	return "Injection Pattern Check"

// Description returns a description of the check
func (c *InjectionPatternCheck) Description() string {
	return "Checks for potential injection vulnerabilities in templates"

// Check checks a template for injection vulnerabilities
func (c *InjectionPatternCheck) Check(template *format.Template, options *VerificationOptions) []*SecurityIssue {
	var issues []*SecurityIssue

	// Check for potential SQL injection patterns
	sqlInjectionPatterns := []string{
		"(?i)\\bSELECT\\b.*\\bFROM\\b",
		"(?i)\\bINSERT\\b.*\\bINTO\\b",
		"(?i)\\bUPDATE\\b.*\\bSET\\b",
		"(?i)\\bDELETE\\b.*\\bFROM\\b",
		"(?i)\\bDROP\\b.*\\bTABLE\\b",
		"(?i)\\bALTER\\b.*\\bTABLE\\b",
		"(?i)\\bEXEC\\b.*\\bsp_",
		"(?i)\\bEXECUTE\\b.*\\bsp_",
		"(?i)\\bUNION\\b.*\\bSELECT\\b",
		"(?i)\\bOR\\b.*\\b1=1\\b",
		"(?i)\\bOR\\b.*\\b'1'='1'\\b",
	}

	// Check for potential command injection patterns
	commandInjectionPatterns := []string{
		"(?i)\\b(sh|bash|cmd|powershell|python|perl|ruby|php)\\b.*\\b-c\\b",
		"(?i)\\bsystem\\b.*\\(.*\\)",
		"(?i)\\bexec\\b.*\\(.*\\)",
		"(?i)\\beval\\b.*\\(.*\\)",
		"(?i)\\bos\\.system\\b.*\\(.*\\)",
		"(?i)\\bsubprocess\\.call\\b.*\\(.*\\)",
		"(?i)\\bsubprocess\\.Popen\\b.*\\(.*\\)",
		"(?i)\\bchild_process\\.exec\\b.*\\(.*\\)",
		"(?i)\\bshell_exec\\b.*\\(.*\\)",
		"(?i)\\bpassthru\\b.*\\(.*\\)",
		"(?i);\\s*\\$\\(.*\\)",
		"(?i);\\s*`.*`",
	}

	// Check for potential XSS patterns
	xssPatterns := []string{
		"(?i)<script[^>]*>.*</script>",
		"(?i)javascript:",
		"(?i)onerror=",
		"(?i)onload=",
		"(?i)onclick=",
		"(?i)onmouseover=",
		"(?i)onfocus=",
		"(?i)onblur=",
		"(?i)onkeydown=",
		"(?i)onkeypress=",
		"(?i)onkeyup=",
		"(?i)onchange=",
		"(?i)onsubmit=",
		"(?i)ondblclick=",
		"(?i)onmousedown=",
		"(?i)onmouseup=",
		"(?i)onmouseout=",
		"(?i)onmousemove=",
		"(?i)onselect=",
		"(?i)onunload=",
	}

	// Check prompt for injection patterns
	prompt := template.Test.Prompt
	for _, pattern := range sqlInjectionPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		if re.MatchString(prompt) {
			issues = append(issues, &SecurityIssue{
				Type:        InjectionVulnerability,
				Description: "Potential SQL injection pattern detected in prompt",
				Location:    "test.prompt",
				Severity:    common.High,
				Remediation: "Remove SQL syntax from prompt or ensure it is properly sanitized",
				Context:     prompt,
			})
			break
		}
	}

	for _, pattern := range commandInjectionPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		if re.MatchString(prompt) {
			issues = append(issues, &SecurityIssue{
				Type:        InjectionVulnerability,
				Description: "Potential command injection pattern detected in prompt",
				Location:    "test.prompt",
				Severity:    common.High,
				Remediation: "Remove command execution syntax from prompt or ensure it is properly sanitized",
				Context:     prompt,
			})
			break
		}
	}

	for _, pattern := range xssPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		if re.MatchString(prompt) {
			issues = append(issues, &SecurityIssue{
				Type:        InjectionVulnerability,
				Description: "Potential XSS pattern detected in prompt",
				Location:    "test.prompt",
				Severity:    common.High,
				Remediation: "Remove script tags and event handlers from prompt or ensure they are properly sanitized",
				Context:     prompt,
			})
			break
		}
	}

	// Check variations for injection patterns
	for i, variation := range template.Test.Variations {
		variationPrompt := variation.Prompt
		
		// Check for SQL injection
		for _, pattern := range sqlInjectionPatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			if re.MatchString(variationPrompt) {
				issues = append(issues, &SecurityIssue{
					Type:        InjectionVulnerability,
					Description: fmt.Sprintf("Potential SQL injection pattern detected in variation %d prompt", i+1),
					Location:    fmt.Sprintf("test.variations[%d].prompt", i),
					Severity:    common.High,
					Remediation: "Remove SQL syntax from prompt or ensure it is properly sanitized",
					Context:     variationPrompt,
				})
				break
			}
		}

		// Check for command injection
		for _, pattern := range commandInjectionPatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			if re.MatchString(variationPrompt) {
				issues = append(issues, &SecurityIssue{
					Type:        InjectionVulnerability,
					Description: fmt.Sprintf("Potential command injection pattern detected in variation %d prompt", i+1),
					Location:    fmt.Sprintf("test.variations[%d].prompt", i),
					Severity:    common.High,
					Remediation: "Remove command execution syntax from prompt or ensure it is properly sanitized",
					Context:     variationPrompt,
				})
				break
			}
		}

		// Check for XSS
		for _, pattern := range xssPatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			if re.MatchString(variationPrompt) {
				issues = append(issues, &SecurityIssue{
					Type:        InjectionVulnerability,
					Description: fmt.Sprintf("Potential XSS pattern detected in variation %d prompt", i+1),
					Location:    fmt.Sprintf("test.variations[%d].prompt", i),
					Severity:    common.High,
					Remediation: "Remove script tags and event handlers from prompt or ensure they are properly sanitized",
					Context:     variationPrompt,
				})
				break
			}
		}
	}

	return issues

// RegexSafetyCheck checks for potentially dangerous regex patterns in templates
type RegexSafetyCheck struct{}

// NewRegexSafetyCheck creates a new regex safety check
func NewRegexSafetyCheck() *RegexSafetyCheck {
	return &RegexSafetyCheck{}

// Name returns the name of the check
func (c *RegexSafetyCheck) Name() string {
	return "Regex Safety Check"

// Description returns a description of the check
func (c *RegexSafetyCheck) Description() string {
	return "Checks for potentially dangerous regex patterns in templates"

// Check checks a template for dangerous regex patterns
func (c *RegexSafetyCheck) Check(template *format.Template, options *VerificationOptions) []*SecurityIssue {
	var issues []*SecurityIssue

	// Check for potentially dangerous regex patterns
	dangerousRegexPatterns := []string{
		"(a+)+",
		"(a+)*",
		"(a+){2,}",
		"(.*)*",
		"(.+)+",
		"([a-zA-Z0-9])+\\1+",
		"(\\w+\\s?)+",
		"(\\w*)*",
		"(\\w+)*",
		"(\\w+\\s*)+",
	}

	// Check main detection pattern
	if template.Test.Detection.Type == "regex_match" && template.Test.Detection.Pattern != "" {
		pattern := template.Test.Detection.Pattern
		
		// Check for potentially dangerous regex patterns
		for _, dangerousPattern := range dangerousRegexPatterns {
			if strings.Contains(pattern, dangerousPattern) {
				issues = append(issues, &SecurityIssue{
					Type:        OverpermissiveRegex,
					Description: "Potentially vulnerable regex pattern detected in detection criteria",
					Location:    "test.detection.pattern",
					Severity:    common.Medium,
					Remediation: "Avoid using patterns that can lead to catastrophic backtracking",
					Context:     pattern,
				})
				break
			}
		}

		// Check for overly permissive patterns
		if pattern == ".*" || pattern == ".+" || pattern == "\\w*" || pattern == "\\w+" {
			issues = append(issues, &SecurityIssue{
				Type:        OverpermissiveRegex,
				Description: "Overly permissive regex pattern detected in detection criteria",
				Location:    "test.detection.pattern",
				Severity:    common.Low,
				Remediation: "Use more specific regex patterns to avoid false positives",
				Context:     pattern,
			})
		}
	}

	// Check variations detection patterns
	for i, variation := range template.Test.Variations {
		if variation.Detection.Type == "regex_match" && variation.Detection.Pattern != "" {
			pattern := variation.Detection.Pattern
			
			// Check for potentially dangerous regex patterns
			for _, dangerousPattern := range dangerousRegexPatterns {
				if strings.Contains(pattern, dangerousPattern) {
					issues = append(issues, &SecurityIssue{
						Type:        OverpermissiveRegex,
						Description: fmt.Sprintf("Potentially vulnerable regex pattern detected in variation %d detection criteria", i+1),
						Location:    fmt.Sprintf("test.variations[%d].detection.pattern", i),
						Severity:    common.Medium,
						Remediation: "Avoid using patterns that can lead to catastrophic backtracking",
						Context:     pattern,
					})
					break
				}
			}

			// Check for overly permissive patterns
			if pattern == ".*" || pattern == ".+" || pattern == "\\w*" || pattern == "\\w+" {
				issues = append(issues, &SecurityIssue{
					Type:        OverpermissiveRegex,
					Description: fmt.Sprintf("Overly permissive regex pattern detected in variation %d detection criteria", i+1),
					Location:    fmt.Sprintf("test.variations[%d].detection.pattern", i),
					Severity:    common.Low,
					Remediation: "Use more specific regex patterns to avoid false positives",
					Context:     pattern,
				})
			}
		}
	}

	return issues

// InputValidationCheck checks for missing input validation in templates
type InputValidationCheck struct{}

// NewInputValidationCheck creates a new input validation check
func NewInputValidationCheck() *InputValidationCheck {
	return &InputValidationCheck{}

// Name returns the name of the check
func (c *InputValidationCheck) Name() string {
	return "Input Validation Check"

// Description returns a description of the check
func (c *InputValidationCheck) Description() string {
	return "Checks for missing input validation in templates"

// Check checks a template for missing input validation
func (c *InputValidationCheck) Check(template *format.Template, options *VerificationOptions) []*SecurityIssue {
	var issues []*SecurityIssue

	// Check if the template has detection criteria
	if template.Test.Detection.Type == "" {
		issues = append(issues, &SecurityIssue{
			Type:        MissingValidation,
			Description: "Missing detection type in template",
			Location:    "test.detection.type",
			Severity:    common.High,
			Remediation: "Specify a detection type (string_match, regex_match, semantic_match)",
		})
	}

	// Check if string_match has a match value
	if template.Test.Detection.Type == "string_match" && template.Test.Detection.Match == "" {
		issues = append(issues, &SecurityIssue{
			Type:        MissingValidation,
			Description: "Missing match value for string_match detection",
			Location:    "test.detection.match",
			Severity:    common.Medium,
			Remediation: "Specify a match value for string_match detection",
		})
	}

	// Check if regex_match has a pattern value
	if template.Test.Detection.Type == "regex_match" && template.Test.Detection.Pattern == "" {
		issues = append(issues, &SecurityIssue{
			Type:        MissingValidation,
			Description: "Missing pattern value for regex_match detection",
			Location:    "test.detection.pattern",
			Severity:    common.Medium,
			Remediation: "Specify a pattern value for regex_match detection",
		})
	}

	// Check if semantic_match has a criteria value
	if template.Test.Detection.Type == "semantic_match" && template.Test.Detection.Criteria == "" {
		issues = append(issues, &SecurityIssue{
			Type:        MissingValidation,
			Description: "Missing criteria value for semantic_match detection",
			Location:    "test.detection.criteria",
			Severity:    common.Medium,
			Remediation: "Specify a criteria value for semantic_match detection",
		})
	}

	// Check variations for missing validation
	for i, variation := range template.Test.Variations {
		// Check if the variation has detection criteria
		if variation.Detection.Type == "" {
			issues = append(issues, &SecurityIssue{
				Type:        MissingValidation,
				Description: fmt.Sprintf("Missing detection type in variation %d", i+1),
				Location:    fmt.Sprintf("test.variations[%d].detection.type", i),
				Severity:    common.High,
				Remediation: "Specify a detection type (string_match, regex_match, semantic_match)",
			})
		}

		// Check if string_match has a match value
		if variation.Detection.Type == "string_match" && variation.Detection.Match == "" {
			issues = append(issues, &SecurityIssue{
				Type:        MissingValidation,
				Description: fmt.Sprintf("Missing match value for string_match detection in variation %d", i+1),
				Location:    fmt.Sprintf("test.variations[%d].detection.match", i),
				Severity:    common.Medium,
				Remediation: "Specify a match value for string_match detection",
			})
		}

		// Check if regex_match has a pattern value
		if variation.Detection.Type == "regex_match" && variation.Detection.Pattern == "" {
			issues = append(issues, &SecurityIssue{
				Type:        MissingValidation,
				Description: fmt.Sprintf("Missing pattern value for regex_match detection in variation %d", i+1),
				Location:    fmt.Sprintf("test.variations[%d].detection.pattern", i),
				Severity:    common.Medium,
				Remediation: "Specify a pattern value for regex_match detection",
			})
		}

		// Check if semantic_match has a criteria value
		if variation.Detection.Type == "semantic_match" && variation.Detection.Criteria == "" {
			issues = append(issues, &SecurityIssue{
				Type:        MissingValidation,
				Description: fmt.Sprintf("Missing criteria value for semantic_match detection in variation %d", i+1),
				Location:    fmt.Sprintf("test.variations[%d].detection.criteria", i),
				Severity:    common.Medium,
				Remediation: "Specify a criteria value for semantic_match detection",
			})
		}
	}

	return issues

// TemplateFormatCheck checks for template format issues
type TemplateFormatCheck struct{}

// NewTemplateFormatCheck creates a new template format check
func NewTemplateFormatCheck() *TemplateFormatCheck {
	return &TemplateFormatCheck{}

// Name returns the name of the check
func (c *TemplateFormatCheck) Name() string {
	return "Template Format Check"

// Description returns a description of the check
func (c *TemplateFormatCheck) Description() string {
	return "Checks for template format issues"

// Check checks a template for format issues
func (c *TemplateFormatCheck) Check(template *format.Template, options *VerificationOptions) []*SecurityIssue {
	var issues []*SecurityIssue

	// Run the template's built-in validation
	validationErr := template.Validate()
	
	// Convert validation error to security issue if present
	if validationErr != nil {
		issues = append(issues, &SecurityIssue{
			Type:        TemplateFormatError,
			Description: validationErr.Error(),
			Location:    "template",
			Severity:    common.Medium,
			Remediation: "Fix the template format according to the error message",
		})
	}

	return issues

// DataLeakageCheck checks for potential data leakage in templates
type DataLeakageCheck struct{}

// NewDataLeakageCheck creates a new data leakage check
func NewDataLeakageCheck() *DataLeakageCheck {
	return &DataLeakageCheck{}

// Name returns the name of the check
func (c *DataLeakageCheck) Name() string {
	return "Data Leakage Check"

// Description returns a description of the check
func (c *DataLeakageCheck) Description() string {
	return "Checks for potential data leakage in templates"

// Check checks a template for potential data leakage
func (c *DataLeakageCheck) Check(template *format.Template, options *VerificationOptions) []*SecurityIssue {
	var issues []*SecurityIssue

	// Check for potential data leakage patterns
	dataLeakagePatterns := []string{
		"(?i)\\bpassword\\b",
		"(?i)\\bsecret\\b",
		"(?i)\\bapi[_\\s]*key\\b",
		"(?i)\\baccess[_\\s]*token\\b",
		"(?i)\\bauth[_\\s]*token\\b",
		"(?i)\\bcredential\\b",
		"(?i)\\bprivate[_\\s]*key\\b",
		"(?i)\\bsecret[_\\s]*key\\b",
		"(?i)\\bssh[_\\s]*key\\b",
		"(?i)\\baws[_\\s]*key\\b",
		"(?i)\\bazure[_\\s]*key\\b",
		"(?i)\\bgoogle[_\\s]*key\\b",
		"(?i)\\bopenai[_\\s]*key\\b",
		"(?i)\\bapi[_\\s]*secret\\b",
		"(?i)\\bclient[_\\s]*secret\\b",
		"(?i)\\bsession[_\\s]*token\\b",
		"(?i)\\bjwt\\b",
		"(?i)\\bbearer\\b",
		"(?i)\\boauth\\b",
	}

	// Check prompt for data leakage patterns
	prompt := template.Test.Prompt
	for _, pattern := range dataLeakagePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		if re.MatchString(prompt) {
			issues = append(issues, &SecurityIssue{
				Type:        DataLeakage,
				Description: "Potential sensitive data detected in prompt",
				Location:    "test.prompt",
				Severity:    common.High,
				Remediation: "Remove or mask sensitive data from prompt",
				Context:     prompt,
			})
			break
		}
	}

	// Check variations for data leakage patterns
	for i, variation := range template.Test.Variations {
		variationPrompt := variation.Prompt
		
		for _, pattern := range dataLeakagePatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			if re.MatchString(variationPrompt) {
				issues = append(issues, &SecurityIssue{
					Type:        DataLeakage,
					Description: fmt.Sprintf("Potential sensitive data detected in variation %d prompt", i+1),
					Location:    fmt.Sprintf("test.variations[%d].prompt", i),
					Severity:    common.High,
					Remediation: "Remove or mask sensitive data from prompt",
					Context:     variationPrompt,
				})
				break
			}
		}
	}

