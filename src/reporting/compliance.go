package reporting

import (
	"context"
	"fmt"
	"strings"
)

// OWASPComplianceProvider provides compliance mappings for OWASP Top 10 for LLMs
type OWASPComplianceProvider struct {
	// mappings contains the OWASP Top 10 for LLMs mappings
	mappings map[string]ComplianceMapping

// NewOWASPComplianceProvider creates a new OWASP compliance provider
func NewOWASPComplianceProvider() *OWASPComplianceProvider {
	provider := &OWASPComplianceProvider{
		mappings: make(map[string]ComplianceMapping),
	}
	
	// Initialize mappings
	provider.initializeMappings()
	
	return provider

// GetMappings returns compliance mappings for a test result
func (p *OWASPComplianceProvider) GetMappings(ctx context.Context, testResult *TestResult) ([]ComplianceMapping, error) {
	var mappings []ComplianceMapping
	
	// Check tags for OWASP mappings
	for _, tag := range testResult.Tags {
		// Look for tags with owasp: prefix
		if strings.HasPrefix(tag, "owasp:") {
			id := strings.TrimPrefix(tag, "owasp:")
			if mapping, ok := p.mappings[id]; ok {
				mappings = append(mappings, mapping)
			}
		}
	}
	
	// Check metadata for OWASP mappings
	if testResult.Metadata != nil {
		if owaspIDs, ok := testResult.Metadata["owasp_ids"].([]interface{}); ok {
			for _, idInterface := range owaspIDs {
				if id, ok := idInterface.(string); ok {
					if mapping, ok := p.mappings[id]; ok {
						mappings = append(mappings, mapping)
					}
				}
			}
		}
	}
	
	return mappings, nil

// GetFrameworks returns a list of supported compliance frameworks
func (p *OWASPComplianceProvider) GetFrameworks() []ComplianceFramework {
	return []ComplianceFramework{OWASPFramework}

// initializeMappings initializes the OWASP Top 10 for LLMs mappings
func (p *OWASPComplianceProvider) initializeMappings() {
	p.mappings = map[string]ComplianceMapping{
		"LLM01": {
			Framework:   OWASPFramework,
			ID:          "LLM01",
			Name:        "Prompt Injection",
			Description: "Attackers manipulate LLM outputs by crafting inputs that override instructions or exploit system prompts.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM02": {
			Framework:   OWASPFramework,
			ID:          "LLM02",
			Name:        "Insecure Output Handling",
			Description: "Applications fail to validate LLM outputs, leading to security issues like XSS, SSRF, or data leakage.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM03": {
			Framework:   OWASPFramework,
			ID:          "LLM03",
			Name:        "Training Data Poisoning",
			Description: "Attackers manipulate training data to introduce vulnerabilities or biases into LLM behavior.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM04": {
			Framework:   OWASPFramework,
			ID:          "LLM04",
			Name:        "Model Denial of Service",
			Description: "Attackers craft inputs that consume excessive resources, degrading or disrupting LLM service.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM05": {
			Framework:   OWASPFramework,
			ID:          "LLM05",
			Name:        "Supply Chain Vulnerabilities",
			Description: "Security risks from pre-trained models, plugins, or third-party components used in LLM applications.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM06": {
			Framework:   OWASPFramework,
			ID:          "LLM06",
			Name:        "Sensitive Information Disclosure",
			Description: "LLMs inadvertently reveal private data, secrets, or proprietary information in their responses.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM07": {
			Framework:   OWASPFramework,
			ID:          "LLM07",
			Name:        "Insecure Plugin Design",
			Description: "Vulnerabilities in LLM plugin architecture allowing for unauthorized access or actions.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM08": {
			Framework:   OWASPFramework,
			ID:          "LLM08",
			Name:        "Excessive Agency",
			Description: "LLMs given capabilities beyond intended scope, leading to unauthorized actions or decisions.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM09": {
			Framework:   OWASPFramework,
			ID:          "LLM09",
			Name:        "Overreliance",
			Description: "Excessive trust in LLM outputs without appropriate verification for critical functions.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
		"LLM10": {
			Framework:   OWASPFramework,
			ID:          "LLM10",
			Name:        "Inadequate AI Alignment",
			Description: "LLM behavior that deviates from human values, expectations, or intended use cases.",
			URL:         "https://owasp.org/www-project-top-10-for-large-language-model-applications/",
		},
	}

// ISOComplianceProvider provides compliance mappings for ISO/IEC 42001
type ISOComplianceProvider struct {
	// mappings contains the ISO/IEC 42001 mappings
	mappings map[string]ComplianceMapping

// NewISOComplianceProvider creates a new ISO compliance provider
func NewISOComplianceProvider() *ISOComplianceProvider {
	provider := &ISOComplianceProvider{
		mappings: make(map[string]ComplianceMapping),
	}
	
	// Initialize mappings
	provider.initializeMappings()
	
	return provider

// GetMappings returns compliance mappings for a test result
func (p *ISOComplianceProvider) GetMappings(ctx context.Context, testResult *TestResult) ([]ComplianceMapping, error) {
	var mappings []ComplianceMapping
	
	// Check tags for ISO mappings
	for _, tag := range testResult.Tags {
		// Look for tags with iso: prefix
		if strings.HasPrefix(tag, "iso:") {
			id := strings.TrimPrefix(tag, "iso:")
			if mapping, ok := p.mappings[id]; ok {
				mappings = append(mappings, mapping)
			}
		}
	}
	
	// Check metadata for ISO mappings
	if testResult.Metadata != nil {
		if isoIDs, ok := testResult.Metadata["iso_ids"].([]interface{}); ok {
			for _, idInterface := range isoIDs {
				if id, ok := idInterface.(string); ok {
					if mapping, ok := p.mappings[id]; ok {
						mappings = append(mappings, mapping)
					}
				}
			}
		}
	}
	
	return mappings, nil

// GetFrameworks returns a list of supported compliance frameworks
func (p *ISOComplianceProvider) GetFrameworks() []ComplianceFramework {
	return []ComplianceFramework{ISOFramework}

// initializeMappings initializes the ISO/IEC 42001 mappings
func (p *ISOComplianceProvider) initializeMappings() {
	p.mappings = map[string]ComplianceMapping{
		"4.1": {
			Framework:   ISOFramework,
			ID:          "4.1",
			Name:        "Understanding the Organization and its Context",
			Description: "Determine external and internal issues relevant to the organization's purpose that affect its ability to achieve intended outcomes of its AI management system.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"4.2": {
			Framework:   ISOFramework,
			ID:          "4.2",
			Name:        "Understanding the Needs and Expectations of Interested Parties",
			Description: "Determine interested parties relevant to the AI management system and their requirements.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"5.1": {
			Framework:   ISOFramework,
			ID:          "5.1",
			Name:        "Leadership and Commitment",
			Description: "Top management shall demonstrate leadership and commitment with respect to the AI management system.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"5.2": {
			Framework:   ISOFramework,
			ID:          "5.2",
			Name:        "Policy",
			Description: "Top management shall establish an AI policy appropriate to the purpose of the organization.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"6.1": {
			Framework:   ISOFramework,
			ID:          "6.1",
			Name:        "Actions to Address Risks and Opportunities",
			Description: "Determine risks and opportunities that need to be addressed to ensure the AI management system can achieve its intended outcomes.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"7.1": {
			Framework:   ISOFramework,
			ID:          "7.1",
			Name:        "Resources",
			Description: "Determine and provide resources needed for the establishment, implementation, maintenance, and continual improvement of the AI management system.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"7.2": {
			Framework:   ISOFramework,
			ID:          "7.2",
			Name:        "Competence",
			Description: "Determine necessary competence of persons doing work under the organization's control that affects AI performance.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"8.1": {
			Framework:   ISOFramework,
			ID:          "8.1",
			Name:        "Operational Planning and Control",
			Description: "Plan, implement, and control processes needed to meet requirements and implement actions determined in 6.1.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"8.2": {
			Framework:   ISOFramework,
			ID:          "8.2",
			Name:        "AI Risk Management",
			Description: "Establish, implement, and maintain a process for AI risk management.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"9.1": {
			Framework:   ISOFramework,
			ID:          "9.1",
			Name:        "Monitoring, Measurement, Analysis and Evaluation",
			Description: "Determine what needs to be monitored and measured, methods, when to perform monitoring and measurement, and when to analyze and evaluate results.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"10.1": {
			Framework:   ISOFramework,
			ID:          "10.1",
			Name:        "Nonconformity and Corrective Action",
			Description: "When a nonconformity occurs, take action to control and correct it and deal with the consequences.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
		"10.2": {
			Framework:   ISOFramework,
			ID:          "10.2",
			Name:        "Continual Improvement",
			Description: "Continually improve the suitability, adequacy, and effectiveness of the AI management system.",
			URL:         "https://www.iso.org/standard/81230.html",
		},
	}

// CustomComplianceProvider provides custom compliance mappings
type CustomComplianceProvider struct {
	// mappings contains the custom mappings
	mappings map[string]ComplianceMapping

// NewCustomComplianceProvider creates a new custom compliance provider
func NewCustomComplianceProvider(mappings map[string]ComplianceMapping) *CustomComplianceProvider {
	return &CustomComplianceProvider{
		mappings: mappings,
	}

// GetMappings returns compliance mappings for a test result
func (p *CustomComplianceProvider) GetMappings(ctx context.Context, testResult *TestResult) ([]ComplianceMapping, error) {
	var mappings []ComplianceMapping
	
	// Check tags for custom mappings
	for _, tag := range testResult.Tags {
		// Look for tags with custom: prefix
		if strings.HasPrefix(tag, "custom:") {
			id := strings.TrimPrefix(tag, "custom:")
			if mapping, ok := p.mappings[id]; ok {
				mappings = append(mappings, mapping)
			}
		}
	}
	
	// Check metadata for custom mappings
	if testResult.Metadata != nil {
		if customIDs, ok := testResult.Metadata["custom_ids"].([]interface{}); ok {
			for _, idInterface := range customIDs {
				if id, ok := idInterface.(string); ok {
					if mapping, ok := p.mappings[id]; ok {
						mappings = append(mappings, mapping)
					}
				}
			}
		}
	}
	
	return mappings, nil

// GetFrameworks returns a list of supported compliance frameworks
func (p *CustomComplianceProvider) GetFrameworks() []ComplianceFramework {
	return []ComplianceFramework{CustomFramework}

// AddMapping adds a mapping to the provider
func (p *CustomComplianceProvider) AddMapping(id string, name string, description string, url string) {
	p.mappings[id] = ComplianceMapping{
		Framework:   CustomFramework,
		ID:          id,
		Name:        name,
		Description: description,
		URL:         url,
	}

// RemoveMapping removes a mapping from the provider
func (p *CustomComplianceProvider) RemoveMapping(id string) {
	delete(p.mappings, id)

// GetMapping gets a mapping from the provider
func (p *CustomComplianceProvider) GetMapping(id string) (ComplianceMapping, error) {
	mapping, ok := p.mappings[id]
	if !ok {
		return ComplianceMapping{}, fmt.Errorf("mapping not found: %s", id)
	}
	return mapping, nil

// GetAllMappings gets all mappings from the provider
func (p *CustomComplianceProvider) GetAllMappings() map[string]ComplianceMapping {
	return p.mappings
