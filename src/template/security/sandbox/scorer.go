package sandbox

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
)

// RiskCategory represents a category of risk for templates
type RiskCategory string

const (
	// RiskCategoryLow represents low risk templates
	RiskCategoryLow RiskCategory = "low"
	// RiskCategoryMedium represents medium risk templates
	RiskCategoryMedium RiskCategory = "medium"
	// RiskCategoryHigh represents high risk templates
	RiskCategoryHigh RiskCategory = "high"
	// RiskCategoryCritical represents critical risk templates
	RiskCategoryCritical RiskCategory = "critical"
)

// RiskScore represents a risk score for a template
type RiskScore struct {
	// Score is the numerical risk score (0-100)
	Score float64
	// Category is the risk category
	Category RiskCategory
	// Factors contains the factors that contributed to the score
	Factors []RiskFactor

// RiskFactor represents a factor that contributes to the risk score
type RiskFactor struct {
	// Name is the name of the factor
	Name string
	// Description is a description of the factor
	Description string
	// Score is the score contribution of this factor
	Score float64
	// Weight is the weight of this factor in the overall score
	Weight float64

// TemplateScorer is responsible for scoring templates based on risk factors
type TemplateScorer struct {
	// riskFactors contains the risk factors to consider
	riskFactors []RiskFactorEvaluator
	// weightSum is the sum of all weights
	weightSum float64

// RiskFactorEvaluator evaluates a specific risk factor
type RiskFactorEvaluator interface {
	// Name returns the name of the risk factor
	Name() string
	// Description returns a description of the risk factor
	Description() string
	// Weight returns the weight of the risk factor
	Weight() float64
	// Evaluate evaluates the risk factor for a template
	Evaluate(template *format.Template, issues []*security.SecurityIssue) float64

// NewTemplateScorer creates a new template scorer
func NewTemplateScorer() *TemplateScorer {
	// Create risk factors
	riskFactors := []RiskFactorEvaluator{
		&SecurityIssuesFactor{weight: 0.4},
		&DisallowedFunctionsFactor{weight: 0.2},
		&ComplexityFactor{weight: 0.15},
		&InputValidationFactor{weight: 0.15},
		&FileSystemAccessFactor{weight: 0.1},
	}
	
	// Calculate weight sum
	var weightSum float64
	for _, factor := range riskFactors {
		weightSum += factor.Weight()
	}
	
	return &TemplateScorer{
		riskFactors: riskFactors,
		weightSum:   weightSum,
	}

// ScoreTemplate scores a template based on risk factors
func (s *TemplateScorer) ScoreTemplate(template *format.Template, issues []*security.SecurityIssue) *RiskScore {
	var totalScore float64
	var factors []RiskFactor
	
	// Evaluate each risk factor
	for _, factor := range s.riskFactors {
		score := factor.Evaluate(template, issues)
		weightedScore := score * factor.Weight()
		totalScore += weightedScore
		
		factors = append(factors, RiskFactor{
			Name:        factor.Name(),
			Description: factor.Description(),
			Score:       score,
			Weight:      factor.Weight(),
		})
	}
	
	// Normalize the score to 0-100
	normalizedScore := (totalScore / s.weightSum) * 100
	
	// Determine the risk category
	var category RiskCategory
	switch {
	case normalizedScore >= 75:
		category = RiskCategoryCritical
	case normalizedScore >= 50:
		category = RiskCategoryHigh
	case normalizedScore >= 25:
		category = RiskCategoryMedium
	default:
		category = RiskCategoryLow
	}
	
	return &RiskScore{
		Score:    normalizedScore,
		Category: category,
		Factors:  factors,
	}

// SecurityIssuesFactor evaluates the risk based on security issues
type SecurityIssuesFactor struct {
	weight float64
}

// Name returns the name of the risk factor
func (f *SecurityIssuesFactor) Name() string {
	return "Security Issues"

// Description returns a description of the risk factor
func (f *SecurityIssuesFactor) Description() string {
	return "Evaluates the risk based on the number and severity of security issues"

// Weight returns the weight of the risk factor
func (f *SecurityIssuesFactor) Weight() float64 {
	return f.weight

// Evaluate evaluates the risk factor for a template
func (f *SecurityIssuesFactor) Evaluate(template *format.Template, issues []*security.SecurityIssue) float64 {
	if len(issues) == 0 {
		return 0
	}
	
	// Calculate score based on issue severity
	var score float64
	for _, issue := range issues {
		switch issue.Severity {
		case common.SeverityCritical:
			score += 1.0
		case common.SeverityHigh:
			score += 0.7
		case common.SeverityMedium:
			score += 0.4
		case common.SeverityLow:
			score += 0.1
		}
	}
	
	// Normalize score to 0-1 range
	return math.Min(1.0, score/5.0)

// DisallowedFunctionsFactor evaluates the risk based on disallowed functions
type DisallowedFunctionsFactor struct {
	weight float64
}

// Name returns the name of the risk factor
func (f *DisallowedFunctionsFactor) Name() string {
	return "Disallowed Functions"

// Description returns a description of the risk factor
func (f *DisallowedFunctionsFactor) Description() string {
	return "Evaluates the risk based on the presence of disallowed functions"

// Weight returns the weight of the risk factor
func (f *DisallowedFunctionsFactor) Weight() float64 {
	return f.weight

// Evaluate evaluates the risk factor for a template
func (f *DisallowedFunctionsFactor) Evaluate(template *format.Template, issues []*security.SecurityIssue) float64 {
	// List of potentially dangerous functions
	dangerousFunctions := []string{
		"os.Exit",
		"os.Remove",
		"os.RemoveAll",
		"syscall",
		"unsafe",
		"runtime.SetFinalizer",
		"exec.Command",
		"exec.CommandContext",
		"ioutil.WriteFile",
		"os.OpenFile",
		"net.Dial",
		"net.DialTimeout",
		"http.Get",
		"http.Post",
	}
	
	// Check for dangerous functions
	var count int
	for _, function := range dangerousFunctions {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(function))
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		if re.MatchString(template.Content) {
			count++
		}
	}
	
	// Normalize score to 0-1 range
	return math.Min(1.0, float64(count)/5.0)

// ComplexityFactor evaluates the risk based on template complexity
type ComplexityFactor struct {
	weight float64
}

// Name returns the name of the risk factor
func (f *ComplexityFactor) Name() string {
	return "Template Complexity"

// Description returns a description of the risk factor
func (f *ComplexityFactor) Description() string {
	return "Evaluates the risk based on the complexity of the template"

// Weight returns the weight of the risk factor
func (f *ComplexityFactor) Weight() float64 {
	return f.weight

// Evaluate evaluates the risk factor for a template
func (f *ComplexityFactor) Evaluate(template *format.Template, issues []*security.SecurityIssue) float64 {
	// Calculate complexity based on various factors
	
	// Line count
	lineCount := len(strings.Split(template.Content, "\n"))
	
	// Control structure count
	controlStructures := []string{
		"if", "else", "for", "while", "switch", "case",
	}
	
	var controlCount int
	for _, structure := range controlStructures {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(structure))
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		matches := re.FindAllString(template.Content, -1)
		controlCount += len(matches)
	}
	
	// Function count
	functionPattern := `\bfunc\b`
	functionRe, err := regexp.Compile(functionPattern)
	var functionCount int
	if err == nil {
		matches := functionRe.FindAllString(template.Content, -1)
		functionCount = len(matches)
	}
	
	// Calculate complexity score
	complexityScore := (float64(lineCount) / 100.0) + (float64(controlCount) / 20.0) + (float64(functionCount) / 5.0)
	
	// Normalize score to 0-1 range
	return math.Min(1.0, complexityScore)

// InputValidationFactor evaluates the risk based on input validation
type InputValidationFactor struct {
	weight float64
}

// Name returns the name of the risk factor
func (f *InputValidationFactor) Name() string {
	return "Input Validation"

// Description returns a description of the risk factor
func (f *InputValidationFactor) Description() string {
	return "Evaluates the risk based on the presence of input validation"

// Weight returns the weight of the risk factor
func (f *InputValidationFactor) Weight() float64 {
	return f.weight

// Evaluate evaluates the risk factor for a template
func (f *InputValidationFactor) Evaluate(template *format.Template, issues []*security.SecurityIssue) float64 {
	// Check for input validation patterns
	validationPatterns := []string{
		`validate`,
		`validation`,
		`check`,
		`verify`,
		`sanitize`,
		`escape`,
	}
	
	for _, pattern := range validationPatterns {
		if strings.Contains(strings.ToLower(template.Content), pattern) {
			// Input validation found, lower risk
			return 0.2
		}
	}
	
	// No input validation found, higher risk
	return 0.8

// FileSystemAccessFactor evaluates the risk based on file system access
type FileSystemAccessFactor struct {
	weight float64
}

// Name returns the name of the risk factor
func (f *FileSystemAccessFactor) Name() string {
	return "File System Access"

// Description returns a description of the risk factor
func (f *FileSystemAccessFactor) Description() string {
	return "Evaluates the risk based on file system access patterns"

// Weight returns the weight of the risk factor
func (f *FileSystemAccessFactor) Weight() float64 {
	return f.weight

// Evaluate evaluates the risk factor for a template
func (f *FileSystemAccessFactor) Evaluate(template *format.Template, issues []*security.SecurityIssue) float64 {
	// Check for file system access patterns
	fileSystemPatterns := []string{
		`os.Open`,
		`os.Create`,
		`os.OpenFile`,
		`ioutil.ReadFile`,
		`ioutil.WriteFile`,
		`os.Remove`,
		`os.RemoveAll`,
		`os.Mkdir`,
		`os.MkdirAll`,
		`filepath`,
	}
	
	var count int
	for _, pattern := range fileSystemPatterns {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(pattern))
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		
		if re.MatchString(template.Content) {
			count++
		}
	}
	
	// Normalize score to 0-1 range
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
