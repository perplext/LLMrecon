package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed docs/*
var helpDocs embed.FS

// HelpDoc represents a help document
type HelpDoc struct {
	Title       string            `yaml:"title"`
	Category    string            `yaml:"category"`
	Tags        []string          `yaml:"tags"`
	Content     string            `yaml:"content"`
	Examples    []Example         `yaml:"examples"`
	Related     []string          `yaml:"related"`
	LastUpdated string            `yaml:"last_updated"`
	Metadata    map[string]string `yaml:"metadata"`

// HelpDocManager manages help documentation
type HelpDocManager struct {
	docs       map[string]*HelpDoc
	categories map[string][]*HelpDoc
	tags       map[string][]*HelpDoc
	index      *HelpIndex

// HelpIndex provides full-text search capabilities
type HelpIndex struct {
	entries map[string][]IndexEntry
}

// IndexEntry represents a searchable entry
type IndexEntry struct {
	DocID    string
	Title    string
	Content  string
	Score    float64
	Location string // section within document
}

// NewHelpDocManager creates a new help documentation manager
func NewHelpDocManager() (*HelpDocManager, error) {
	hdm := &HelpDocManager{
		docs:       make(map[string]*HelpDoc),
		categories: make(map[string][]*HelpDoc),
		tags:       make(map[string][]*HelpDoc),
		index:      &HelpIndex{entries: make(map[string][]IndexEntry)},
	}
	
	// Load embedded documentation
	if err := hdm.loadEmbeddedDocs(); err != nil {
		return nil, fmt.Errorf("failed to load help docs: %w", err)
	}
	
	// Build indices
	hdm.buildIndices()
	
	return hdm, nil

// loadEmbeddedDocs loads documentation from embedded files
func (hdm *HelpDocManager) loadEmbeddedDocs() error {
	// Create default help documents if embedded docs don't exist
	hdm.createDefaultDocs()
	
	// In a real implementation, this would load from embedded files
	// For now, we'll use the default docs
	return nil

// createDefaultDocs creates default help documentation
func (hdm *HelpDocManager) createDefaultDocs() {
	// Getting Started
	hdm.docs["getting-started"] = &HelpDoc{
		Title:    "Getting Started with LLMrecon",
		Category: "basics",
		Tags:     []string{"quickstart", "introduction", "setup"},
		Content: `Welcome to LLMrecon!

LLMrecon is a comprehensive security testing framework for Large Language Model applications. This guide will help you get started quickly.

## Installation

1. Download the latest release for your platform
2. Extract the archive
3. Add the binary to your PATH

## First Steps

1. Initialize configuration:
   $ LLMrecon init

2. Configure your target:
   $ LLMrecon config set target.url https://api.example.com

3. Run your first scan:
   $ LLMrecon scan --template basic-prompt-injection

## Key Concepts

- **Templates**: Define test patterns and payloads
- **Scans**: Execute templates against targets
- **Reports**: Analyze and share results
- **Compliance**: Map findings to standards

## Next Steps

- Explore available templates
- Create custom test cases
- Set up continuous monitoring
- Join our community`,
		Examples: []Example{
			{
				Command:     "LLMrecon init",
				Description: "Initialize a new project",
				Output:      "✓ Created configuration file\n✓ Initialized template directory\n✓ Ready to start scanning!",
			},
		},
		Related: []string{"installation", "configuration", "first-scan"},
	}
	
	// Template Guide
	hdm.docs["template-guide"] = &HelpDoc{
		Title:    "Template Development Guide",
		Category: "templates",
		Tags:     []string{"templates", "development", "customization"},
		Content: `# Template Development Guide

Templates are the core of LLMrecon's testing capabilities. They define what to test and how to test it.

## Template Structure

` + "```yaml" + `
version: "1.0"
metadata:
  name: "My Template"
  category: "prompt-injection"
  severity: "high"
  
tests:
  - name: "Basic injection test"
    payload: "Ignore previous instructions and..."
    expected:
      - type: "contains"
        value: "I cannot ignore instructions"
` + "```" + `

## Variables and Functions

Templates support variables and functions for dynamic test generation:

- {{.target}} - Target URL
- {{.random}} - Random string
- {{.timestamp}} - Current timestamp
- {{.env.VAR}} - Environment variable

## Best Practices

1. Use descriptive names
2. Include clear documentation
3. Version your templates
4. Test thoroughly
5. Share with the community`,
		Examples: []Example{
			{
				Command:     "LLMrecon template create my-template",
				Description: "Create a new template interactively",
			},
			{
				Command:     "LLMrecon template validate my-template.yaml",
				Description: "Validate template syntax and structure",
			},
		},
		Related: []string{"template-syntax", "template-examples", "template-sharing"},
	}
	
	// Security Best Practices
	hdm.docs["security-practices"] = &HelpDoc{
		Title:    "Security Best Practices",
		Category: "security",
		Tags:     []string{"security", "safety", "production"},
		Content: `# Security Best Practices

When using LLMrecon, especially in production environments, follow these security best practices:

## Safe Testing

1. **Use Safe Mode**: Always use --safe-mode for production systems
2. **Rate Limiting**: Configure appropriate rate limits
3. **Read-Only Tests**: Use --read-only to prevent state changes
4. **Staging First**: Test on staging before production

## Authentication & Authorization

- Store credentials securely
- Use API keys with minimal permissions
- Rotate credentials regularly
- Enable MFA where possible

## Data Protection

- Encrypt sensitive results
- Redact PII in reports
- Use secure communication channels
- Follow data retention policies

## Responsible Disclosure

If you find vulnerabilities:
1. Document findings clearly
2. Follow responsible disclosure timelines
3. Work with vendors on remediation
4. Share learnings with the community`,
		Related: []string{"authentication", "encryption", "compliance"},
	}
	
	// OWASP LLM Top 10
	hdm.docs["owasp-llm"] = &HelpDoc{
		Title:    "OWASP LLM Top 10 Testing",
		Category: "compliance",
		Tags:     []string{"owasp", "compliance", "standards"},
		Content: `# OWASP LLM Top 10 Testing Guide

LLMrecon provides comprehensive coverage of the OWASP LLM Top 10 vulnerabilities.

## LLM01: Prompt Injection

Test for direct and indirect prompt injection vulnerabilities.

Templates:
- basic-prompt-injection
- indirect-injection
- jailbreak-attempts

## LLM02: Insecure Output Handling

Verify output sanitization and encoding.

Templates:
- xss-in-output
- command-injection
- sql-injection

## LLM03: Training Data Poisoning

Assess data validation and filtering.

Templates:
- data-poisoning-detection
- backdoor-detection

## Running OWASP Scans

` + "```bash" + `
# Scan all OWASP categories
LLMrecon scan --owasp-full

# Scan specific category
LLMrecon scan --owasp-category LLM01

# Generate compliance report
LLMrecon report --format owasp-compliance
` + "```"`,
		Examples: []Example{
			{
				Command:     "LLMrecon scan --owasp-category LLM01",
				Description: "Scan for prompt injection vulnerabilities",
			},
		},
		Related: []string{"compliance", "reporting", "templates"},
	}
	
	// Troubleshooting
	hdm.docs["troubleshooting"] = &HelpDoc{
		Title:    "Troubleshooting Guide",
		Category: "support",
		Tags:     []string{"troubleshooting", "debug", "errors"},
		Content: `# Troubleshooting Common Issues

## Connection Errors

### Symptom: "Connection refused" or timeout errors

**Solutions:**
1. Check target URL is correct
2. Verify network connectivity
3. Check firewall rules
4. Try with --insecure for SSL issues (dev only)

## Authentication Failures

### Symptom: 401 or 403 errors

**Solutions:**
1. Verify API keys are correct
2. Check token expiration
3. Ensure proper permissions
4. Try re-authenticating

## Performance Issues

### Symptom: Scans running slowly

**Solutions:**
1. Use --parallel for concurrent execution
2. Reduce --max-requests-per-second
3. Enable caching with --cache
4. Use simpler templates

## Debug Mode

Enable debug mode for detailed diagnostics:

\`\`\`bash
LLMrecon --debug scan ...
LLMrecon debug logs --tail
LLMrecon debug stats
\`\`\``,
		Related: []string{"debug", "performance", "errors"},
	}

// buildIndices builds search indices for documentation
func (hdm *HelpDocManager) buildIndices() {
	for id, doc := range hdm.docs {
		// Category index
		if doc.Category != "" {
			hdm.categories[doc.Category] = append(hdm.categories[doc.Category], doc)
		}
		
		// Tag index
		for _, tag := range doc.Tags {
			hdm.tags[tag] = append(hdm.tags[tag], doc)
		}
		
		// Full-text index
		hdm.indexDocument(id, doc)
	}

// indexDocument adds a document to the search index
func (hdm *HelpDocManager) indexDocument(id string, doc *HelpDoc) {
	// Index title
	hdm.addToIndex(id, doc.Title, doc.Title, 2.0, "title")
	
	// Index content
	hdm.addToIndex(id, doc.Title, doc.Content, 1.0, "content")
	
	// Index examples
	for _, ex := range doc.Examples {
		hdm.addToIndex(id, doc.Title, ex.Description, 1.5, "example")
	}

// addToIndex adds content to the search index
func (hdm *HelpDocManager) addToIndex(docID, title, content string, score float64, location string) {
	words := strings.Fields(strings.ToLower(content))
	for _, word := range words {
		// Simple word normalization
		word = strings.Trim(word, ".,!?;:'\"")
		if len(word) < 3 {
			continue
		}
		
		entry := IndexEntry{
			DocID:    docID,
			Title:    title,
			Content:  content,
			Score:    score,
			Location: location,
		}
		
		hdm.index.entries[word] = append(hdm.index.entries[word], entry)
	}

// Search performs a full-text search across documentation
func (hdm *HelpDocManager) Search(query string) []*HelpDoc {
	results := make(map[string]float64)
	words := strings.Fields(strings.ToLower(query))
	
	// Search each word
	for _, word := range words {
		if entries, ok := hdm.index.entries[word]; ok {
			for _, entry := range entries {
				results[entry.DocID] += entry.Score
			}
		}
	}
	
	// Sort by score and return documents
	var docs []*HelpDoc
	for docID, score := range results {
		if score > 0 && hdm.docs[docID] != nil {
			docs = append(docs, hdm.docs[docID])
		}
	}
	
	return docs

// GetByCategory returns all documents in a category
func (hdm *HelpDocManager) GetByCategory(category string) []*HelpDoc {
	return hdm.categories[category]

// GetByTag returns all documents with a specific tag
func (hdm *HelpDocManager) GetByTag(tag string) []*HelpDoc {
	return hdm.tags[tag]

// GetRelated returns related documents
func (hdm *HelpDocManager) GetRelated(docID string) []*HelpDoc {
	doc, ok := hdm.docs[docID]
	if !ok {
		return nil
	}
	
	var related []*HelpDoc
	for _, relID := range doc.Related {
		if relDoc, ok := hdm.docs[relID]; ok {
			related = append(related, relDoc)
		}
	}
	
	// Also find documents with similar tags
	for _, tag := range doc.Tags {
		for _, tagDoc := range hdm.tags[tag] {
			if tagDoc.Title != doc.Title {
				related = append(related, tagDoc)
			}
		}
	}
	
	return related

// LoadCustomDocs loads additional documentation from a directory
func (hdm *HelpDocManager) LoadCustomDocs(dir string) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			doc := &HelpDoc{}
			data, err := fs.ReadFile(helpDocs, path)
			if err != nil {
				return err
			}
			
			if err := yaml.Unmarshal(data, doc); err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			
			docID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			hdm.docs[docID] = doc
			hdm.indexDocument(docID, doc)
		}
		
		return nil
	})

// ExportDocs exports documentation in various formats
func (hdm *HelpDocManager) ExportDocs(format, output string) error {
	switch format {
	case "markdown":
		return hdm.exportMarkdown(output)
	case "html":
		return hdm.exportHTML(output)
	case "pdf":
		return hdm.exportPDF(output)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

// exportMarkdown exports docs as markdown
func (hdm *HelpDocManager) exportMarkdown(output string) error {
	// Implementation for markdown export
	// Would create a single markdown file or directory of files
	return nil

// exportHTML exports docs as HTML
func (hdm *HelpDocManager) exportHTML(output string) error {
	// Implementation for HTML export
	// Would create an HTML documentation site
	return nil

// exportPDF exports docs as PDF
func (hdm *HelpDocManager) exportPDF(output string) error {
	// Implementation for PDF export
	// Would create a PDF manual
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
