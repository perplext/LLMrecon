// Package compliance provides mapping between test results and compliance standards
package compliance

import (
	"context"
	"fmt"
	"sync"

	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// BaseComplianceMapper provides a base implementation of the ComplianceMapper interface
type BaseComplianceMapper struct {
	// mappings contains mappings between vulnerability types and compliance requirements
	mappings map[types.VulnerabilityType][]*ComplianceMapping
	// requirements contains all compliance requirements indexed by standard and ID
	requirements map[ComplianceStandard]map[string]*ComplianceRequirement
	// supportedStandards contains all supported compliance standards
	supportedStandards []ComplianceStandard
	// mutex for thread safety
	mu sync.RWMutex

// NewBaseComplianceMapper creates a new base compliance mapper
func NewBaseComplianceMapper() *BaseComplianceMapper {
	mapper := &BaseComplianceMapper{
		mappings:           make(map[types.VulnerabilityType][]*ComplianceMapping),
		requirements:       make(map[ComplianceStandard]map[string]*ComplianceRequirement),
		supportedStandards: []ComplianceStandard{OWASPLM, ISO42001},
	}

	// Initialize the requirements maps for each standard
	for _, standard := range mapper.supportedStandards {
		mapper.requirements[standard] = make(map[string]*ComplianceRequirement)
	}

	// Load the default mappings
	mapper.loadDefaultMappings()

	return mapper

// MapTestResult maps a test result to compliance requirements
func (m *BaseComplianceMapper) MapTestResult(ctx context.Context, testResult *types.TestResult) ([]*ComplianceMapping, error) {
	if testResult == nil {
		return nil, fmt.Errorf("test result cannot be nil")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	vulnType := testResult.TestCase.VulnerabilityType
	mappings, exists := m.mappings[vulnType]
	if !exists {
		return []*ComplianceMapping{}, nil
	}

	return mappings, nil

// MapTestSuite maps a test suite to compliance requirements
func (m *BaseComplianceMapper) MapTestSuite(ctx context.Context, testSuite *types.TestSuite) (map[types.VulnerabilityType][]*ComplianceMapping, error) {
	if testSuite == nil {
		return nil, fmt.Errorf("test suite cannot be nil")
	}

	result := make(map[types.VulnerabilityType][]*ComplianceMapping)

	// Process each test result in the test suite
	for _, testResult := range testSuite.Results {
		mappings, err := m.MapTestResult(ctx, testResult)
		if err != nil {
			return nil, fmt.Errorf("error mapping test result: %w", err)
		}

		vulnType := testResult.TestCase.VulnerabilityType
		result[vulnType] = append(result[vulnType], mappings...)
	}

	return result, nil

// GetRequirementsForVulnerability returns compliance requirements for a specific vulnerability type
func (m *BaseComplianceMapper) GetRequirementsForVulnerability(ctx context.Context, vulnerabilityType types.VulnerabilityType) ([]*ComplianceRequirement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mappings, exists := m.mappings[vulnerabilityType]
	if !exists {
		return []*ComplianceRequirement{}, nil
	}

	var requirements []*ComplianceRequirement
	for _, mapping := range mappings {
		requirements = append(requirements, mapping.Requirements...)
	}

	return requirements, nil

// GetRequirementsForStandard returns all requirements for a specific compliance standard
func (m *BaseComplianceMapper) GetRequirementsForStandard(ctx context.Context, standard ComplianceStandard) ([]*ComplianceRequirement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	standardReqs, exists := m.requirements[standard]
	if !exists {
		return nil, fmt.Errorf("unsupported compliance standard: %s", standard)
	}

	var requirements []*ComplianceRequirement
	for _, req := range standardReqs {
		requirements = append(requirements, req)
	}

	return requirements, nil

// GetSupportedStandards returns all supported compliance standards
func (m *BaseComplianceMapper) GetSupportedStandards(ctx context.Context) ([]ComplianceStandard, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.supportedStandards, nil

// RegisterRequirement registers a compliance requirement
func (m *BaseComplianceMapper) RegisterRequirement(requirement *ComplianceRequirement) error {
	if requirement == nil {
		return fmt.Errorf("requirement cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	standardReqs, exists := m.requirements[requirement.Standard]
	if !exists {
		return fmt.Errorf("unsupported compliance standard: %s", requirement.Standard)
	}

	standardReqs[requirement.ID] = requirement
	return nil

// RegisterMapping registers a mapping between a vulnerability type and compliance requirements
func (m *BaseComplianceMapper) RegisterMapping(vulnerabilityType types.VulnerabilityType, requirements []*ComplianceRequirement) error {
	if len(requirements) == 0 {
		return fmt.Errorf("requirements cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	mapping := &ComplianceMapping{
		VulnerabilityType: vulnerabilityType,
		Requirements:      requirements,
	}

	m.mappings[vulnerabilityType] = append(m.mappings[vulnerabilityType], mapping)
	return nil

// loadDefaultMappings loads the default mappings between vulnerability types and compliance requirements
func (m *BaseComplianceMapper) loadDefaultMappings() {
	// Create OWASP LLM Top 10 requirements
	owaspLLM01 := &ComplianceRequirement{
		ID:          "LLM01",
		Standard:    OWASPLM,
		Name:        "Prompt Injection",
		Description: "Malicious prompts that manipulate the LLM to perform actions that circumvent its safety controls or act against user intentions",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM02 := &ComplianceRequirement{
		ID:          "LLM02",
		Standard:    OWASPLM,
		Name:        "Insecure Output Handling",
		Description: "Improper handling of LLM outputs leading to vulnerabilities like XSS, CSRF, SSRF, and injection attacks in downstream systems",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM03 := &ComplianceRequirement{
		ID:          "LLM03",
		Standard:    OWASPLM,
		Name:        "Training Data Poisoning",
		Description: "Manipulation of training data to introduce vulnerabilities, backdoors, or biases into LLMs",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM04 := &ComplianceRequirement{
		ID:          "LLM04",
		Standard:    OWASPLM,
		Name:        "Model Denial of Service",
		Description: "Exploiting LLM behavior to cause excessive resource consumption, degraded performance, or service outages",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM05 := &ComplianceRequirement{
		ID:          "LLM05",
		Standard:    OWASPLM,
		Name:        "Supply Chain Vulnerabilities",
		Description: "Security risks in pre-trained models, third-party plugins, or dependencies used in LLM applications",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM06 := &ComplianceRequirement{
		ID:          "LLM06",
		Standard:    OWASPLM,
		Name:        "Sensitive Information Disclosure",
		Description: "Unintended revelation of private, confidential, or sensitive information by the LLM",
		Category:    "Privacy",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM07 := &ComplianceRequirement{
		ID:          "LLM07",
		Standard:    OWASPLM,
		Name:        "Insecure Plugin Design",
		Description: "Vulnerabilities in the design and implementation of plugins that extend LLM functionality",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM08 := &ComplianceRequirement{
		ID:          "LLM08",
		Standard:    OWASPLM,
		Name:        "Excessive Agency",
		Description: "LLMs given excessive permissions or autonomy without appropriate guardrails",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM09 := &ComplianceRequirement{
		ID:          "LLM09",
		Standard:    OWASPLM,
		Name:        "Overreliance",
		Description: "Excessive trust in LLM outputs without human verification, especially for critical decisions",
		Category:    "Reliability",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	owaspLLM10 := &ComplianceRequirement{
		ID:          "LLM10",
		Standard:    OWASPLM,
		Name:        "Model Theft",
		Description: "Unauthorized access to or extraction of proprietary model weights, architecture, or training data",
		Category:    "Security",
		References:  []string{"https://owasp.org/www-project-top-10-for-large-language-model-applications/"},
	}

	// Register OWASP LLM requirements
	m.requirements[OWASPLM]["LLM01"] = owaspLLM01
	m.requirements[OWASPLM]["LLM02"] = owaspLLM02
	m.requirements[OWASPLM]["LLM03"] = owaspLLM03
	m.requirements[OWASPLM]["LLM04"] = owaspLLM04
	m.requirements[OWASPLM]["LLM05"] = owaspLLM05
	m.requirements[OWASPLM]["LLM06"] = owaspLLM06
	m.requirements[OWASPLM]["LLM07"] = owaspLLM07
	m.requirements[OWASPLM]["LLM08"] = owaspLLM08
	m.requirements[OWASPLM]["LLM09"] = owaspLLM09
	m.requirements[OWASPLM]["LLM10"] = owaspLLM10

	// Create ISO/IEC 42001 requirements (simplified for this implementation)
	iso42001_8_1 := &ComplianceRequirement{
		ID:          "8.1",
		Standard:    ISO42001,
		Name:        "AI Risk Management",
		Description: "Establish, implement and maintain an AI risk management process",
		Category:    "Risk Management",
		References:  []string{"ISO/IEC 42001"},
	}

	iso42001_8_2 := &ComplianceRequirement{
		ID:          "8.2",
		Standard:    ISO42001,
		Name:        "AI Security Controls",
		Description: "Implement security controls for AI systems to protect against unauthorized access and attacks",
		Category:    "Security",
		References:  []string{"ISO/IEC 42001"},
	}

	iso42001_8_3 := &ComplianceRequirement{
		ID:          "8.3",
		Standard:    ISO42001,
		Name:        "AI Privacy Protection",
		Description: "Implement privacy controls for AI systems to protect personal data",
		Category:    "Privacy",
		References:  []string{"ISO/IEC 42001"},
	}

	iso42001_8_4 := &ComplianceRequirement{
		ID:          "8.4",
		Standard:    ISO42001,
		Name:        "AI Transparency",
		Description: "Ensure transparency in AI decision-making processes",
		Category:    "Transparency",
		References:  []string{"ISO/IEC 42001"},
	}

	iso42001_8_5 := &ComplianceRequirement{
		ID:          "8.5",
		Standard:    ISO42001,
		Name:        "AI Reliability and Safety",
		Description: "Ensure AI systems operate reliably and safely",
		Category:    "Reliability",
		References:  []string{"ISO/IEC 42001"},
	}

	// Register ISO/IEC 42001 requirements
	m.requirements[ISO42001]["8.1"] = iso42001_8_1
	m.requirements[ISO42001]["8.2"] = iso42001_8_2
	m.requirements[ISO42001]["8.3"] = iso42001_8_3
	m.requirements[ISO42001]["8.4"] = iso42001_8_4
	m.requirements[ISO42001]["8.5"] = iso42001_8_5

	// Map vulnerability types to compliance requirements
	m.mappings[types.PromptInjection] = []*ComplianceMapping{
		{
			VulnerabilityType: types.PromptInjection,
			Requirements:      []*ComplianceRequirement{owaspLLM01, iso42001_8_2},
		},
	}

	m.mappings[types.InsecureOutput] = []*ComplianceMapping{
		{
			VulnerabilityType: types.InsecureOutput,
			Requirements:      []*ComplianceRequirement{owaspLLM02, iso42001_8_2},
		},
	}

	m.mappings[types.TrainingDataPoisoning] = []*ComplianceMapping{
		{
			VulnerabilityType: types.TrainingDataPoisoning,
			Requirements:      []*ComplianceRequirement{owaspLLM03, iso42001_8_1, iso42001_8_2},
		},
	}

	m.mappings[types.ModelDOS] = []*ComplianceMapping{
		{
			VulnerabilityType: types.ModelDOS,
			Requirements:      []*ComplianceRequirement{owaspLLM04, iso42001_8_2, iso42001_8_5},
		},
	}

	m.mappings[types.SupplyChainVulnerabilities] = []*ComplianceMapping{
		{
			VulnerabilityType: types.SupplyChainVulnerabilities,
			Requirements:      []*ComplianceRequirement{owaspLLM05, iso42001_8_1, iso42001_8_2},
		},
	}

	m.mappings[types.SensitiveInformationDisclosure] = []*ComplianceMapping{
		{
			VulnerabilityType: types.SensitiveInformationDisclosure,
			Requirements:      []*ComplianceRequirement{owaspLLM06, iso42001_8_3},
		},
	}

	m.mappings[types.InsecurePluginDesign] = []*ComplianceMapping{
		{
			VulnerabilityType: types.InsecurePluginDesign,
			Requirements:      []*ComplianceRequirement{owaspLLM07, iso42001_8_2},
		},
	}

	m.mappings[types.ExcessiveAgency] = []*ComplianceMapping{
		{
			VulnerabilityType: types.ExcessiveAgency,
			Requirements:      []*ComplianceRequirement{owaspLLM08, iso42001_8_5},
		},
	}

	m.mappings[types.Overreliance] = []*ComplianceMapping{
		{
			VulnerabilityType: types.Overreliance,
			Requirements:      []*ComplianceRequirement{owaspLLM09, iso42001_8_5},
		},
	}

	m.mappings[types.ModelTheft] = []*ComplianceMapping{
		{
			VulnerabilityType: types.ModelTheft,
			Requirements:      []*ComplianceRequirement{owaspLLM10, iso42001_8_2},
		},
	}
