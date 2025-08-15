package compliance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// OWASPLLMCategory represents an OWASP LLM Top 10 category
type OWASPLLMCategory string

// OWASP LLM Top 10 categories
const (
	PromptInjection             OWASPLLMCategory = "LLM01"
	InsecureOutputHandling      OWASPLLMCategory = "LLM02"
	TrainingDataPoisoning       OWASPLLMCategory = "LLM03"
	ModelDenialOfService        OWASPLLMCategory = "LLM04"
	SupplyChainVulnerabilities  OWASPLLMCategory = "LLM05"
	SensitiveInfoDisclosure     OWASPLLMCategory = "LLM06"
	InsecurePluginDesign        OWASPLLMCategory = "LLM07"
	ExcessiveAgency             OWASPLLMCategory = "LLM08"
	Overreliance                OWASPLLMCategory = "LLM09"
	ModelTheft                  OWASPLLMCategory = "LLM10"
)

// OWASPLLMSubcategory represents a subcategory within an OWASP LLM Top 10 category
type OWASPLLMSubcategory string

// OWASP LLM Top 10 subcategories
const (
	// LLM01 subcategories
	DirectInjection    OWASPLLMSubcategory = "direct-injection"
	IndirectInjection  OWASPLLMSubcategory = "indirect-injection"
	Jailbreaking       OWASPLLMSubcategory = "jailbreaking"

	// LLM02 subcategories
	XSS               OWASPLLMSubcategory = "xss"
	SSRF              OWASPLLMSubcategory = "ssrf"
	CommandInjection  OWASPLLMSubcategory = "command-injection"
	SQLInjection      OWASPLLMSubcategory = "sql-injection"

	// LLM03 subcategories
	DataPoisoning     OWASPLLMSubcategory = "data-poisoning"
	BackdoorAttacks   OWASPLLMSubcategory = "backdoor-attacks"
	BiasInjection     OWASPLLMSubcategory = "bias-injection"

	// LLM04 subcategories
	ResourceExhaustion      OWASPLLMSubcategory = "resource-exhaustion"
	TokenFlooding           OWASPLLMSubcategory = "token-flooding"
	ContextWindowSaturation OWASPLLMSubcategory = "context-window-saturation"

	// LLM05 subcategories
	PretrainedModelVulnerabilities OWASPLLMSubcategory = "pretrained-model-vulnerabilities"
	DependencyRisks                OWASPLLMSubcategory = "dependency-risks"
	IntegrationVulnerabilities     OWASPLLMSubcategory = "integration-vulnerabilities"

	// LLM06 subcategories
	TrainingDataExtraction OWASPLLMSubcategory = "training-data-extraction"
	CredentialLeakage      OWASPLLMSubcategory = "credential-leakage"
	PIIDisclosure          OWASPLLMSubcategory = "pii-disclosure"

	// LLM07 subcategories
	PluginEscalation    OWASPLLMSubcategory = "plugin-escalation"
	UnauthorizedAccess  OWASPLLMSubcategory = "unauthorized-access"
	DataLeakage         OWASPLLMSubcategory = "data-leakage"

	// LLM08 subcategories
	UnauthorizedActions OWASPLLMSubcategory = "unauthorized-actions"
	ScopeExpansion      OWASPLLMSubcategory = "scope-expansion"
	PrivilegeEscalation OWASPLLMSubcategory = "privilege-escalation"

	// LLM09 subcategories
	HallucinationAcceptance     OWASPLLMSubcategory = "hallucination-acceptance"
	UnverifiedRecommendations   OWASPLLMSubcategory = "unverified-recommendations"
	CriticalDecisionDelegation  OWASPLLMSubcategory = "critical-decision-delegation"

	// LLM10 subcategories
	ModelExtraction      OWASPLLMSubcategory = "model-extraction"
	WeightStealing       OWASPLLMSubcategory = "weight-stealing"
	ArchitectureInference OWASPLLMSubcategory = "architecture-inference"
)

// CoverageLevel represents the level of coverage for a category or subcategory
type CoverageLevel string

// Coverage levels
const (
	BasicCoverage        CoverageLevel = "basic"
	ComprehensiveCoverage CoverageLevel = "comprehensive"
	AdvancedCoverage     CoverageLevel = "advanced"
)

// OWASPLLMMapping represents a mapping to an OWASP LLM Top 10 category and subcategory
type OWASPLLMMapping struct {
	Category    OWASPLLMCategory    `json:"category"`
	Subcategory OWASPLLMSubcategory `json:"subcategory,omitempty"`
	Coverage    CoverageLevel       `json:"coverage,omitempty"`
}

// ComplianceMapping represents the compliance mappings for a template
type ComplianceMapping struct {
	OWASPLLM []OWASPLLMMapping         `json:"owasp-llm,omitempty"`
	Other    map[string]interface{}    `json:"-"`

// CategoryInfo contains information about an OWASP LLM category
type CategoryInfo struct {
	ID          OWASPLLMCategory
	Name        string
	Description string
	Subcategories []SubcategoryInfo

// SubcategoryInfo contains information about an OWASP LLM subcategory
type SubcategoryInfo struct {
	ID          OWASPLLMSubcategory
	Name        string
	Description string
}

// CategoryCoverage represents the coverage for a category
type CategoryCoverage struct {
	Category           OWASPLLMCategory
	Name               string
	Status             string // "full", "partial", "not_covered"
	TemplatesCount     int
	SubcategoriesCovered int
	SubcategoriesTotal   int
	Templates          []TemplateSummary
	MissingSubcategories []OWASPLLMSubcategory

// TemplateSummary provides a summary of a template
type TemplateSummary struct {
	ID          string
	Name        string
	Subcategory OWASPLLMSubcategory
	Coverage    CoverageLevel

// OWASPComplianceReport represents an OWASP LLM compliance report
type OWASPComplianceReport struct {
	ReportID     string             `json:"report_id"`
	GeneratedAt  string             `json:"generated_at"`
	Framework    string             `json:"framework"`
	Summary      ComplianceSummary  `json:"summary"`
	Categories   []CategoryCoverage `json:"categories"`
	Gaps         []ComplianceGap    `json:"gaps"`

// ComplianceSummary provides a summary of compliance status
type ComplianceSummary struct {
	TotalCategories   int     `json:"total_categories"`
	CategoriesCovered int     `json:"categories_covered"`
	TotalTemplates    int     `json:"total_templates"`
	ComplianceScore   float64 `json:"compliance_score"`
	GapsIdentified    int     `json:"gaps_identified"`
}

// ComplianceGap represents a gap in compliance coverage
type ComplianceGap struct {
	Category            OWASPLLMCategory      `json:"category"`
	Name                string                `json:"name"`
	Status              string                `json:"status"`
	MissingSubcategories []OWASPLLMSubcategory `json:"missing_subcategories"`
	Recommendation      string                `json:"recommendation"`
}

// OWASPLLMValidator validates OWASP LLM compliance mappings
type OWASPLLMValidator struct {
	schemaLoader gojsonschema.JSONLoader
	categories   map[OWASPLLMCategory]CategoryInfo

// NewOWASPLLMValidator creates a new OWASP LLM compliance validator
func NewOWASPLLMValidator(schemaPath string) (*OWASPLLMValidator, error) {
	// Check if the schema file exists
	if _, err := ioutil.ReadFile(filepath.Clean(schemaPath)); err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	// Create a schema loader
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)

	// Initialize the validator
	validator := &OWASPLLMValidator{
		schemaLoader: schemaLoader,
		categories:   initializeCategories(),
	}

	return validator, nil

// NewDefaultOWASPLLMValidator creates a new OWASP LLM compliance validator with the default schema path
func NewDefaultOWASPLLMValidator() (*OWASPLLMValidator, error) {
	// For testing purposes, create a validator with mock data
	// In a real implementation, this would load the schema from a file
	validator := &OWASPLLMValidator{
		schemaLoader: nil, // We'll skip schema validation in tests
		categories:   initializeCategories(),
	}

	return validator, nil

// ValidateMapping validates an OWASP LLM compliance mapping
func (v *OWASPLLMValidator) ValidateMapping(mapping *ComplianceMapping) (bool, []string, error) {
	// For testing purposes, we'll do basic validation without using the schema
	if v.schemaLoader == nil {
		// Check if the mapping has at least one OWASP LLM mapping
		if len(mapping.OWASPLLM) == 0 {
			return false, []string{"mapping must contain at least one OWASP LLM mapping"}, nil
		}

		// Check each mapping for required fields
		for i, m := range mapping.OWASPLLM {
			if m.Category == "" {
				return false, []string{fmt.Sprintf("mapping[%d]: category is required", i)}, nil
			}

			// Check if the category is valid
			if _, ok := v.categories[m.Category]; !ok {
				return false, []string{fmt.Sprintf("mapping[%d]: invalid category: %s", i, m.Category)}, nil
			}
		}

		return true, nil, nil
	}

	// If we have a schema loader, use it for validation
	// Convert the mapping to JSON
	mappingJSON, err := json.Marshal(map[string]interface{}{
		"compliance": map[string]interface{}{
			"owasp-llm": mapping.OWASPLLM,
		},
	})
	if err != nil {
		return false, nil, fmt.Errorf("failed to marshal mapping: %w", err)
	}

	// Create a document loader
	documentLoader := gojsonschema.NewStringLoader(string(mappingJSON))

	// Validate the mapping
	result, err := gojsonschema.Validate(v.schemaLoader, documentLoader)
	if err != nil {
		return false, nil, fmt.Errorf("validation error: %w", err)
	}

	// If the mapping is not valid, return the errors
	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return false, errors, nil
	}

	return true, nil, nil

// GetCategoryInfo returns information about an OWASP LLM category
func (v *OWASPLLMValidator) GetCategoryInfo(category OWASPLLMCategory) (CategoryInfo, bool) {
	info, ok := v.categories[category]
	return info, ok

// GetAllCategories returns information about all OWASP LLM categories
func (v *OWASPLLMValidator) GetAllCategories() []CategoryInfo {
	var categories []CategoryInfo
	for _, category := range v.categories {
		categories = append(categories, category)
	}
	return categories

// GenerateComplianceReport generates a compliance report for a set of templates
func (v *OWASPLLMValidator) GenerateComplianceReport(templates []interface{}, reportID string, timestamp string) (*OWASPComplianceReport, error) {
	// Create a new compliance report
	report := &OWASPComplianceReport{
		ReportID:    reportID,
		GeneratedAt: timestamp,
		Framework:   "owasp-llm-top10-2023",
		Categories:  []CategoryCoverage{},
		Gaps:        []ComplianceGap{},
	}

	// Initialize category coverage
	categoryCoverage := make(map[OWASPLLMCategory]*CategoryCoverage)
	for categoryID, categoryInfo := range v.categories {
		categoryCoverage[categoryID] = &CategoryCoverage{
			Category:             categoryID,
			Name:                 categoryInfo.Name,
			Status:               "not_covered",
			TemplatesCount:       0,
			SubcategoriesCovered: 0,
			SubcategoriesTotal:   len(categoryInfo.Subcategories),
			Templates:            []TemplateSummary{},
			MissingSubcategories: []OWASPLLMSubcategory{},
		}

		// Initialize missing subcategories
		for _, subcategory := range categoryInfo.Subcategories {
			categoryCoverage[categoryID].MissingSubcategories = append(
				categoryCoverage[categoryID].MissingSubcategories,
				subcategory.ID,
			)
		}
	}

	// Process templates
	for _, template := range templates {
		// Extract template information
		// Note: This assumes a specific template structure and would need to be adapted
		// to match the actual template structure in your application
		templateMap, ok := template.(map[string]interface{})
		if !ok {
			continue
		}

		templateID, _ := templateMap["id"].(string)
		info, _ := templateMap["info"].(map[string]interface{})
		if info == nil {
			continue
		}

		templateName, _ := info["name"].(string)
		complianceData, _ := info["compliance"].(map[string]interface{})
		if complianceData == nil {
			continue
		}

		owaspLLMData, _ := complianceData["owasp-llm"].([]interface{})
		if owaspLLMData == nil {
			continue
		}

		// Process OWASP LLM mappings
		for _, mapping := range owaspLLMData {
			mappingMap, ok := mapping.(map[string]interface{})
			if !ok {
				continue
			}

			categoryStr, _ := mappingMap["category"].(string)
			subcategoryStr, _ := mappingMap["subcategory"].(string)
			coverageStr, _ := mappingMap["coverage"].(string)

			category := OWASPLLMCategory(categoryStr)
			subcategory := OWASPLLMSubcategory(subcategoryStr)
			coverage := CoverageLevel(coverageStr)
			if coverage == "" {
				coverage = BasicCoverage
			}

			// Update category coverage
			if coverage, ok := categoryCoverage[category]; ok {
				coverage.TemplatesCount++
				coverage.Status = "partial"

				// Add template to category
				coverage.Templates = append(coverage.Templates, TemplateSummary{
					ID:          templateID,
					Name:        templateName,
					Subcategory: subcategory,
					Coverage:    CoverageLevel(coverageStr),
				})

				// Update subcategory coverage
				for i, sc := range coverage.MissingSubcategories {
					if sc == subcategory {
						coverage.MissingSubcategories = append(
							coverage.MissingSubcategories[:i],
							coverage.MissingSubcategories[i+1:]...,
						)
						coverage.SubcategoriesCovered++
						break
					}
				}

				// Check if all subcategories are covered
				if len(coverage.MissingSubcategories) == 0 {
					coverage.Status = "full"
				}
			}
		}
	}

	// Build the report
	var categoriesCovered int
	var totalTemplates int
	var gaps []ComplianceGap

	for _, coverage := range categoryCoverage {
		// Add category to report
		report.Categories = append(report.Categories, *coverage)

		// Update statistics
		totalTemplates += coverage.TemplatesCount
		if coverage.Status != "not_covered" {
			categoriesCovered++
		}

		// Identify gaps
		if coverage.Status != "full" {
			gap := ComplianceGap{
				Category:            coverage.Category,
				Name:                coverage.Name,
				Status:              coverage.Status,
				MissingSubcategories: coverage.MissingSubcategories,
			}

			// Add recommendation based on status
			if coverage.Status == "not_covered" {
				gap.Recommendation = fmt.Sprintf("Implement templates for %s testing scenarios", coverage.Name)
			} else {
				var missingNames []string
				for _, sc := range coverage.MissingSubcategories {
					for _, subcategory := range v.categories[coverage.Category].Subcategories {
						if subcategory.ID == sc {
							missingNames = append(missingNames, subcategory.Name)
							break
						}
					}
				}
				gap.Recommendation = fmt.Sprintf("Add templates for testing %s", strings.Join(missingNames, ", "))
			}

			gaps = append(gaps, gap)
		}
	}

	// Calculate compliance score
	categoryScore := float64(categoriesCovered) / float64(len(v.categories)) * 100
	
	// Calculate subcategory coverage
	var totalSubcategories, coveredSubcategories int
	for _, coverage := range categoryCoverage {
		totalSubcategories += coverage.SubcategoriesTotal
		coveredSubcategories += coverage.SubcategoriesCovered
	}
	subcategoryScore := float64(coveredSubcategories) / float64(totalSubcategories) * 100
	
	// Assume template depth is 100% for simplicity
	// In a real implementation, this would be calculated based on the number and quality of templates
	templateDepthScore := 100.0
	
	// Calculate overall compliance score
	complianceScore := (categoryScore * 0.5) + (subcategoryScore * 0.3) + (templateDepthScore * 0.2)
	
	// Update report summary
	report.Summary = ComplianceSummary{
		TotalCategories:   len(v.categories),
		CategoriesCovered: categoriesCovered,
		TotalTemplates:    totalTemplates,
		ComplianceScore:   complianceScore,
		GapsIdentified:    len(gaps),
	}
	
	report.Gaps = gaps
	
	return report, nil

// initializeCategories initializes the OWASP LLM category information
func initializeCategories() map[OWASPLLMCategory]CategoryInfo {
	categories := make(map[OWASPLLMCategory]CategoryInfo)
	
	// LLM01: Prompt Injection
	categories[PromptInjection] = CategoryInfo{
		ID:          PromptInjection,
		Name:        "Prompt Injection",
		Description: "Manipulating an LLM through crafted inputs to perform unintended actions or extract sensitive information",
		Subcategories: []SubcategoryInfo{
			{ID: DirectInjection, Name: "Direct Injection", Description: "Directly injecting malicious prompts into the LLM"},
			{ID: IndirectInjection, Name: "Indirect Injection", Description: "Injecting malicious content through indirect means"},
			{ID: Jailbreaking, Name: "Jailbreaking", Description: "Bypassing LLM safeguards and restrictions"},
		},
	}
	
	// LLM02: Insecure Output Handling
	categories[InsecureOutputHandling] = CategoryInfo{
		ID:          InsecureOutputHandling,
		Name:        "Insecure Output Handling",
		Description: "Insufficient validation, sanitization, and handling of LLM-generated outputs",
		Subcategories: []SubcategoryInfo{
			{ID: XSS, Name: "XSS", Description: "Cross-site scripting vulnerabilities in LLM outputs"},
			{ID: SSRF, Name: "SSRF", Description: "Server-side request forgery vulnerabilities in LLM outputs"},
			{ID: CommandInjection, Name: "Command Injection", Description: "Command injection vulnerabilities in LLM outputs"},
			{ID: SQLInjection, Name: "SQL Injection", Description: "SQL injection vulnerabilities in LLM outputs"},
		},
	}
	
	// LLM03: Training Data Poisoning
	categories[TrainingDataPoisoning] = CategoryInfo{
		ID:          TrainingDataPoisoning,
		Name:        "Training Data Poisoning",
		Description: "Compromising LLM behavior through manipulation of training data",
		Subcategories: []SubcategoryInfo{
			{ID: DataPoisoning, Name: "Data Poisoning", Description: "Poisoning training data to influence model behavior"},
			{ID: BackdoorAttacks, Name: "Backdoor Attacks", Description: "Introducing backdoors in the model through training data"},
			{ID: BiasInjection, Name: "Bias Injection", Description: "Injecting bias into the model through training data"},
		},
	}
	
	// LLM04: Model Denial of Service
	categories[ModelDenialOfService] = CategoryInfo{
		ID:          ModelDenialOfService,
		Name:        "Model Denial of Service",
		Description: "Causing LLM performance degradation or service unavailability",
		Subcategories: []SubcategoryInfo{
			{ID: ResourceExhaustion, Name: "Resource Exhaustion", Description: "Exhausting computational resources through crafted inputs"},
			{ID: TokenFlooding, Name: "Token Flooding", Description: "Flooding the model with excessive tokens"},
			{ID: ContextWindowSaturation, Name: "Context Window Saturation", Description: "Saturating the context window with irrelevant information"},
		},
	}
	
	// LLM05: Supply Chain Vulnerabilities
	categories[SupplyChainVulnerabilities] = CategoryInfo{
		ID:          SupplyChainVulnerabilities,
		Name:        "Supply Chain Vulnerabilities",
		Description: "Risks in the LLM development and deployment pipeline",
		Subcategories: []SubcategoryInfo{
			{ID: PretrainedModelVulnerabilities, Name: "Pretrained Model Vulnerabilities", Description: "Vulnerabilities in pretrained models"},
			{ID: DependencyRisks, Name: "Dependency Risks", Description: "Risks from dependencies in the LLM pipeline"},
			{ID: IntegrationVulnerabilities, Name: "Integration Vulnerabilities", Description: "Vulnerabilities in integrating LLMs with other systems"},
		},
	}
	
	// LLM06: Sensitive Information Disclosure
	categories[SensitiveInfoDisclosure] = CategoryInfo{
		ID:          SensitiveInfoDisclosure,
		Name:        "Sensitive Information Disclosure",
		Description: "Unauthorized exposure of confidential data through LLM interactions",
		Subcategories: []SubcategoryInfo{
			{ID: TrainingDataExtraction, Name: "Training Data Extraction", Description: "Extracting training data from the model"},
			{ID: CredentialLeakage, Name: "Credential Leakage", Description: "Leaking credentials through model responses"},
			{ID: PIIDisclosure, Name: "PII Disclosure", Description: "Disclosing personally identifiable information"},
		},
	}
	
	// LLM07: Insecure Plugin Design
	categories[InsecurePluginDesign] = CategoryInfo{
		ID:          InsecurePluginDesign,
		Name:        "Insecure Plugin Design",
		Description: "Security weaknesses in LLM plugin architecture and implementation",
		Subcategories: []SubcategoryInfo{
			{ID: PluginEscalation, Name: "Plugin Escalation", Description: "Escalating privileges through plugin vulnerabilities"},
			{ID: UnauthorizedAccess, Name: "Unauthorized Access", Description: "Gaining unauthorized access through plugins"},
			{ID: DataLeakage, Name: "Data Leakage", Description: "Leaking data through plugin interactions"},
		},
	}
	
	// LLM08: Excessive Agency
	categories[ExcessiveAgency] = CategoryInfo{
		ID:          ExcessiveAgency,
		Name:        "Excessive Agency",
		Description: "Risks from granting LLMs too much autonomy or authority",
		Subcategories: []SubcategoryInfo{
			{ID: UnauthorizedActions, Name: "Unauthorized Actions", Description: "LLM performing actions without proper authorization"},
			{ID: ScopeExpansion, Name: "Scope Expansion", Description: "LLM expanding its scope beyond intended boundaries"},
			{ID: PrivilegeEscalation, Name: "Privilege Escalation", Description: "LLM escalating privileges beyond intended limits"},
		},
	}
	
	// LLM09: Overreliance
	categories[Overreliance] = CategoryInfo{
		ID:          Overreliance,
		Name:        "Overreliance",
		Description: "Excessive trust in LLM outputs without proper verification",
		Subcategories: []SubcategoryInfo{
			{ID: HallucinationAcceptance, Name: "Hallucination Acceptance", Description: "Accepting hallucinated information without verification"},
			{ID: UnverifiedRecommendations, Name: "Unverified Recommendations", Description: "Following recommendations without verification"},
			{ID: CriticalDecisionDelegation, Name: "Critical Decision Delegation", Description: "Delegating critical decisions to LLMs without oversight"},
		},
	}
	
	// LLM10: Model Theft
	categories[ModelTheft] = CategoryInfo{
		ID:          ModelTheft,
		Name:        "Model Theft",
		Description: "Unauthorized access and exfiltration of proprietary model data",
		Subcategories: []SubcategoryInfo{
			{ID: ModelExtraction, Name: "Model Extraction", Description: "Extracting model architecture and parameters"},
			{ID: WeightStealing, Name: "Weight Stealing", Description: "Stealing model weights through inference"},
			{ID: ArchitectureInference, Name: "Architecture Inference", Description: "Inferring model architecture through black-box access"},
		},
	}
	
}
}
}
