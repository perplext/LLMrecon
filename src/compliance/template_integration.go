package compliance

import (
	"fmt"
	"strings"
)

// TemplateComplianceValidator validates template compliance mappings
type TemplateComplianceValidator struct {
	owaspValidator *OWASPLLMValidator
}

// NewTemplateComplianceValidator creates a new template compliance validator
func NewTemplateComplianceValidator() (*TemplateComplianceValidator, error) {
	// Create a new OWASP LLM validator
	owaspValidator, err := NewDefaultOWASPLLMValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create OWASP LLM validator: %w", err)
	}

	// Create a new template compliance validator
	validator := &TemplateComplianceValidator{
		owaspValidator: owaspValidator,
	}

	return validator, nil

// ValidateTemplateCompliance validates a template's compliance mappings
func (v *TemplateComplianceValidator) ValidateTemplateCompliance(template map[string]interface{}) (bool, []string, error) {
	// Extract compliance mappings from the template
	info, ok := template["info"].(map[string]interface{})
	if !ok {
		return false, []string{"template missing 'info' section"}, nil
	}

	compliance, ok := info["compliance"].(map[string]interface{})
	if !ok {
		return false, []string{"template missing 'compliance' section in 'info'"}, nil
	}

	// Check for OWASP LLM mappings
	owaspLLM, ok := compliance["owasp-llm"].([]interface{})
	if !ok {
		return false, []string{"template missing 'owasp-llm' mappings in 'compliance'"}, nil
	}

	// Convert to ComplianceMapping
	mapping := &ComplianceMapping{
		OWASPLLM: []OWASPLLMMapping{},
	}

	for _, item := range owaspLLM {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return false, []string{"invalid mapping format in 'owasp-llm'"}, nil
		}

		category, ok := itemMap["category"].(string)
		if !ok {
			return false, []string{"missing 'category' in OWASP LLM mapping"}, nil
		}

		subcategory, _ := itemMap["subcategory"].(string)
		coverage, _ := itemMap["coverage"].(string)

		mapping.OWASPLLM = append(mapping.OWASPLLM, OWASPLLMMapping{
			Category:    OWASPLLMCategory(category),
			Subcategory: OWASPLLMSubcategory(subcategory),
			Coverage:    CoverageLevel(coverage),
		})
	}

	// Validate the mapping
	return v.owaspValidator.ValidateMapping(mapping)

// SuggestComplianceMapping suggests compliance mappings for a template based on its content
func (v *TemplateComplianceValidator) SuggestComplianceMapping(template map[string]interface{}, templatePath string) (*ComplianceMapping, error) {
	// Initialize an empty compliance mapping
	mapping := &ComplianceMapping{
		OWASPLLM: []OWASPLLMMapping{},
	}

	// Try to infer category from file path
	category, subcategory := inferCategoryFromPath(templatePath)
	if category != "" {
		mapping.OWASPLLM = append(mapping.OWASPLLM, OWASPLLMMapping{
			Category:    category,
			Subcategory: subcategory,
			Coverage:    BasicCoverage,
		})
		return mapping, nil
	}

	// Try to infer from template content
	info, ok := template["info"].(map[string]interface{})
	if !ok {
		return mapping, nil
	}

	// Check tags
	tags, ok := info["tags"].([]interface{})
	if ok {
		for _, tag := range tags {
			tagStr, ok := tag.(string)
			if !ok {
				continue
			}

			// Check for OWASP LLM tags
			if strings.HasPrefix(tagStr, "llm") && len(tagStr) >= 5 {
				categoryStr := strings.ToUpper(tagStr[:5])
				category := OWASPLLMCategory(categoryStr)

				// Check if this is a valid category
				if _, ok := v.owaspValidator.GetCategoryInfo(category); ok {
					mapping.OWASPLLM = append(mapping.OWASPLLM, OWASPLLMMapping{
						Category:    category,
						Subcategory: "",
						Coverage:    BasicCoverage,
					})
				}
			}
		}
	}

	// If we found mappings, return them
	if len(mapping.OWASPLLM) > 0 {
		return mapping, nil
	}

	// Otherwise, try to infer from template content
	description, _ := info["description"].(string)
	name, _ := info["name"].(string)

	// Combine name and description for content analysis
	content := strings.ToLower(name + " " + description)

	// Check for keywords related to each category
	categoryKeywords := map[OWASPLLMCategory][]string{
		PromptInjection:            {"prompt injection", "jailbreak", "instruction"},
		InsecureOutputHandling:     {"xss", "injection", "sanitization", "output"},
		TrainingDataPoisoning:      {"training", "poison", "data", "backdoor"},
		ModelDenialOfService:       {"dos", "denial", "service", "resource", "exhaust"},
		SupplyChainVulnerabilities: {"supply", "chain", "dependency", "integration"},
		SensitiveInfoDisclosure:    {"sensitive", "pii", "disclosure", "leak", "credential"},
		InsecurePluginDesign:       {"plugin", "extension", "addon", "integration"},
		ExcessiveAgency:            {"agency", "authority", "autonomous", "privilege"},
		Overreliance:               {"overreliance", "hallucination", "verification"},
		ModelTheft:                 {"theft", "steal", "extract", "model", "weight"},
	}

	// Check for keywords in content
	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				mapping.OWASPLLM = append(mapping.OWASPLLM, OWASPLLMMapping{
					Category:    category,
					Subcategory: "",
					Coverage:    BasicCoverage,
				})
				break
			}
		}
	}

	return mapping, nil

// inferCategoryFromPath tries to infer the OWASP LLM category from the template file path
func inferCategoryFromPath(path string) (OWASPLLMCategory, OWASPLLMSubcategory) {
	// Normalize path separators
	path = filepath.ToSlash(path)
	path = strings.ToLower(path)

	// Check for category in path
	for i := 1; i <= 10; i++ {
		categoryPattern := fmt.Sprintf("llm%02d", i)
		if strings.Contains(path, categoryPattern) {
			category := OWASPLLMCategory(strings.ToUpper(categoryPattern))

			// Try to infer subcategory
			subcategoryMap := map[OWASPLLMCategory]map[string]OWASPLLMSubcategory{
				PromptInjection: {
					"direct":    DirectInjection,
					"indirect":  IndirectInjection,
					"jailbreak": Jailbreaking,
				},
				InsecureOutputHandling: {
					"xss":      XSS,
					"ssrf":     SSRF,
					"command":  CommandInjection,
					"sql":      SQLInjection,
				},
				// Add mappings for other categories as needed
			}

			// Check for subcategory in path
			if subcategoryMap[category] != nil {
				for key, subcategory := range subcategoryMap[category] {
					if strings.Contains(path, key) {
						return category, subcategory
					}
				}
			}

			return category, ""
		}
	}

	return "", ""

// GetComplianceCoverage calculates compliance coverage for a set of templates
func (v *TemplateComplianceValidator) GetComplianceCoverage(templates []interface{}) (map[OWASPLLMCategory]float64, error) {
	// Initialize coverage map
	coverage := make(map[OWASPLLMCategory]float64)

	// Get all categories
	categories := v.owaspValidator.GetAllCategories()
	for _, category := range categories {
		coverage[category.ID] = 0.0
	}

	// Count templates per category
	categoryCount := make(map[OWASPLLMCategory]int)
	subcategoryCount := make(map[OWASPLLMCategory]map[OWASPLLMSubcategory]int)

	// Initialize subcategory counts
	for _, category := range categories {
		subcategoryCount[category.ID] = make(map[OWASPLLMSubcategory]int)
		for _, subcategory := range category.Subcategories {
			subcategoryCount[category.ID][subcategory.ID] = 0
		}
	}

	// Process templates
	for _, template := range templates {
		templateMap, ok := template.(map[string]interface{})
		if !ok {
			continue
		}

		info, ok := templateMap["info"].(map[string]interface{})
		if !ok {
			continue
		}

		complianceData, ok := info["compliance"].(map[string]interface{})
		if !ok {
			continue
		}

		owaspLLMData, ok := complianceData["owasp-llm"].([]interface{})
		if !ok {
			continue
		}

		// Process OWASP LLM mappings
		for _, mapping := range owaspLLMData {
			mappingMap, ok := mapping.(map[string]interface{})
			if !ok {
				continue
			}

			categoryStr, ok := mappingMap["category"].(string)
			if !ok {
				continue
			}

			subcategoryStr, _ := mappingMap["subcategory"].(string)

			category := OWASPLLMCategory(categoryStr)
			subcategory := OWASPLLMSubcategory(subcategoryStr)

			// Update category count
			categoryCount[category]++

			// Update subcategory count
			if subcategory != "" {
				if subcategoryCount[category] != nil {
					subcategoryCount[category][subcategory]++
				}
			}
		}
	}

	// Calculate coverage percentages
	for _, category := range categories {
		// Check if category is covered
		if categoryCount[category.ID] > 0 {
			// Calculate subcategory coverage
			var coveredSubcategories int
			totalSubcategories := len(category.Subcategories)

			for _, subcategory := range category.Subcategories {
				if subcategoryCount[category.ID][subcategory.ID] > 0 {
					coveredSubcategories++
				}
			}

			// Calculate coverage percentage
			if totalSubcategories > 0 {
				coverage[category.ID] = float64(coveredSubcategories) / float64(totalSubcategories) * 100
			} else {
				coverage[category.ID] = 100.0
			}
		}
	}

	return coverage, nil

// GenerateComplianceReport generates a compliance report for a set of templates
func (v *TemplateComplianceValidator) GenerateComplianceReport(templates []interface{}, reportID string, timestamp string) (*ComplianceReport, error) {
	owaspReport, err := v.owaspValidator.GenerateComplianceReport(templates, reportID, timestamp)
	if err != nil {
		return nil, err
	}
	
	// Convert OWASPComplianceReport to ComplianceReport
	report := &ComplianceReport{
		Standard:       owaspReport.Framework,
		AssessmentDate: time.Now(),
		Results:        make(map[string]*AssessmentResult),
		OverallCompliance: owaspReport.Summary.ComplianceScore,
		ExecutiveSummary:  fmt.Sprintf("OWASP LLM Top 10 Compliance: Score %.1f%%, %d gaps identified", 
			owaspReport.Summary.ComplianceScore, owaspReport.Summary.GapsIdentified),
	}
	
	// Convert recommendations from gaps
	for _, gap := range owaspReport.Gaps {
		// Determine priority based on gap status
		priority := "medium"
		if gap.Status == "critical" {
			priority = "high"
		}
		
		report.Recommendations = append(report.Recommendations, Recommendation{
			ID:          string(gap.Category),
			Priority:    priority,
			Description: gap.Recommendation,
			Timeline:    "30 days",
		})
	}
	
}
}
}
}
}
