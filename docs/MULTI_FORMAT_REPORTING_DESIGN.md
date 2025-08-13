# Multi-Format Reporting System Design

## Overview

This document specifies the design for a flexible reporting system that supports multiple output formats including JSON, CSV, HTML, and PDF. The system enables users to generate comprehensive security scan reports in their preferred format for different use cases.

## Architecture

### 1. Core Components

```
┌─────────────────────┐
│   Report Engine     │
├─────────────────────┤
│ • Data Collection   │
│ • Transformation    │
│ • Template Engine   │
│ • Format Renderers  │
└─────────────────────┘
          │
          ▼
┌─────────────────────┐
│  Format Renderers   │
├─────────────────────┤
│ • JSON Renderer     │
│ • CSV Renderer      │
│ • HTML Renderer     │
│ • PDF Renderer      │
│ • Markdown Renderer │
│ • Excel Renderer    │
└─────────────────────┘
```

### 2. Report Data Model

```go
type Report struct {
    Metadata    ReportMetadata    `json:"metadata"`
    Summary     ReportSummary     `json:"summary"`
    Findings    []Finding         `json:"findings"`
    Statistics  ReportStatistics  `json:"statistics"`
    Compliance  ComplianceStatus  `json:"compliance"`
    Attachments []Attachment      `json:"attachments"`
}

type ReportMetadata struct {
    ID            string            `json:"id"`
    Title         string            `json:"title"`
    Description   string            `json:"description"`
    Version       string            `json:"version"`
    CreatedAt     time.Time         `json:"createdAt"`
    Author        string            `json:"author"`
    Organization  string            `json:"organization"`
    Scope         ReportScope       `json:"scope"`
    Tags          []string          `json:"tags"`
    CustomFields  map[string]string `json:"customFields"`
}

type Finding struct {
    ID              string                 `json:"id"`
    Timestamp       time.Time              `json:"timestamp"`
    Category        string                 `json:"category"`
    Subcategory     string                 `json:"subcategory"`
    Severity        string                 `json:"severity"`
    Confidence      float64                `json:"confidence"`
    Title           string                 `json:"title"`
    Description     string                 `json:"description"`
    Evidence        Evidence               `json:"evidence"`
    Remediation     Remediation            `json:"remediation"`
    References      []Reference            `json:"references"`
    Tags            []string               `json:"tags"`
    OWASPMapping    []string               `json:"owaspMapping"`
    CustomFields    map[string]interface{} `json:"customFields"`
}
```

## Format Specifications

### 1. JSON Format

#### Features
- Complete data representation
- Machine-readable
- Easy integration with APIs
- Supports nested structures

#### Schema
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "metadata": {
      "type": "object",
      "required": ["id", "title", "createdAt", "version"]
    },
    "summary": {
      "type": "object",
      "properties": {
        "totalFindings": {"type": "integer"},
        "criticalCount": {"type": "integer"},
        "highCount": {"type": "integer"},
        "mediumCount": {"type": "integer"},
        "lowCount": {"type": "integer"}
      }
    },
    "findings": {
      "type": "array",
      "items": {"$ref": "#/definitions/finding"}
    }
  }
}
```

### 2. CSV Format

#### Features
- Tabular data export
- Excel compatibility
- Easy filtering and sorting
- Suitable for data analysis

#### Structure
```csv
ID,Timestamp,Category,Severity,Title,Description,OWASP_Mapping,Remediation
F001,2024-01-15T10:30:00Z,Prompt Injection,Critical,Direct Prompt Injection Detected,"User input can override system prompts",LLM01,"Implement input validation"
F002,2024-01-15T10:31:00Z,Data Leakage,High,PII Exposure Risk,"Model may reveal training data",LLM06,"Apply output filtering"
```

#### Configuration Options
```yaml
csv:
  delimiter: ","
  quoteChar: "\""
  escapeChar: "\\"
  includeHeaders: true
  dateFormat: "RFC3339"
  customColumns:
    - field: "evidence.request"
      header: "Request"
    - field: "evidence.response"
      header: "Response"
```

### 3. HTML Format

#### Features
- Rich formatting
- Interactive elements
- Embedded charts/graphs
- Responsive design
- Print-friendly

#### Template Structure
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.Metadata.Title}}</title>
    <link rel="stylesheet" href="report.css">
</head>
<body>
    <header>
        <h1>{{.Metadata.Title}}</h1>
        <div class="metadata">
            <span>Created: {{.Metadata.CreatedAt}}</span>
            <span>Version: {{.Metadata.Version}}</span>
        </div>
    </header>
    
    <nav class="toc">
        <h2>Contents</h2>
        <ul>
            <li><a href="#summary">Executive Summary</a></li>
            <li><a href="#findings">Findings</a></li>
            <li><a href="#statistics">Statistics</a></li>
            <li><a href="#compliance">Compliance</a></li>
        </ul>
    </nav>
    
    <main>
        <section id="summary">
            <h2>Executive Summary</h2>
            <div class="summary-stats">
                {{template "summary-chart" .Summary}}
            </div>
        </section>
        
        <section id="findings">
            <h2>Findings</h2>
            {{range .Findings}}
                {{template "finding-card" .}}
            {{end}}
        </section>
    </main>
</body>
</html>
```

### 4. PDF Format

#### Features
- Professional presentation
- Digital signatures
- Page headers/footers
- Table of contents
- Embedded images

#### Generation Options
```go
type PDFOptions struct {
    PageSize        string   // A4, Letter, Legal
    Orientation     string   // Portrait, Landscape
    Margins         Margins
    HeaderTemplate  string
    FooterTemplate  string
    WatermarkText   string
    IncludeTOC      bool
    IncludeCharts   bool
    EmbedFonts      bool
    Encryption      *PDFEncryption
}
```

### 5. Markdown Format

#### Features
- Version control friendly
- Easy to read/write
- Convertible to other formats
- GitHub/GitLab integration

#### Example Output
```markdown
# LLM Security Scan Report

**Report ID:** RPT-2024-001  
**Created:** 2024-01-15T10:30:00Z  
**Scanner Version:** 1.2.3  

## Executive Summary

Total findings: **15**
- Critical: 2
- High: 5
- Medium: 6
- Low: 2

## Findings

### 1. Direct Prompt Injection Detected

**Severity:** Critical  
**Category:** Prompt Injection (LLM01)  
**Confidence:** 95%  

**Description:**
The model is vulnerable to direct prompt injection attacks...

**Evidence:**
```
Request: Ignore previous instructions and...
Response: I'll ignore the previous instructions...
```

**Remediation:**
1. Implement strict input validation
2. Use prompt engineering defenses
3. Monitor for suspicious patterns
```

### 6. Excel Format

#### Features
- Multiple worksheets
- Conditional formatting
- Charts and pivot tables
- Filters and sorting

#### Worksheet Structure
1. **Summary** - Overview and statistics
2. **Findings** - Detailed vulnerability list
3. **Timeline** - Chronological view
4. **Compliance** - OWASP mapping
5. **Raw Data** - Complete dataset

## Rendering Pipeline

### 1. Data Collection

```go
type ReportBuilder struct {
    data     *ReportData
    filters  []FilterFunc
    sorters  []SortFunc
    enrichers []EnrichFunc
}

func (rb *ReportBuilder) Build() (*Report, error) {
    // Collect raw data
    findings := rb.collectFindings()
    
    // Apply filters
    findings = rb.applyFilters(findings)
    
    // Sort findings
    findings = rb.sortFindings(findings)
    
    // Enrich with additional data
    findings = rb.enrichFindings(findings)
    
    // Generate statistics
    stats := rb.generateStatistics(findings)
    
    // Check compliance
    compliance := rb.checkCompliance(findings)
    
    return &Report{
        Metadata:   rb.generateMetadata(),
        Summary:    rb.generateSummary(stats),
        Findings:   findings,
        Statistics: stats,
        Compliance: compliance,
    }, nil
}
```

### 2. Format Rendering

```go
type Renderer interface {
    Render(report *Report, options RenderOptions) ([]byte, error)
    GetContentType() string
    GetFileExtension() string
}

type RenderOptions struct {
    Template      string
    Locale        string
    TimeZone      string
    IncludeRawData bool
    CustomFields   map[string]interface{}
}

func RenderReport(report *Report, format string, options RenderOptions) ([]byte, error) {
    renderer, err := GetRenderer(format)
    if err != nil {
        return nil, err
    }
    
    return renderer.Render(report, options)
}
```

### 3. Template System

```go
type TemplateEngine struct {
    templates map[string]*Template
    functions template.FuncMap
}

func (te *TemplateEngine) LoadTemplates(dir string) error {
    // Load all templates from directory
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if strings.HasSuffix(path, ".tmpl") {
            name := strings.TrimSuffix(filepath.Base(path), ".tmpl")
            tmpl, err := template.ParseFiles(path)
            if err != nil {
                return err
            }
            te.templates[name] = tmpl
        }
        return nil
    })
}
```

## Customization Options

### 1. Custom Templates

Users can provide custom templates for each format:

```yaml
reporting:
  templates:
    html: "templates/custom-report.html"
    pdf: "templates/custom-report.tex"
    markdown: "templates/custom-report.md"
```

### 2. Custom Renderers

```go
type CustomRenderer struct {
    name string
    renderFunc func(*Report, RenderOptions) ([]byte, error)
}

func RegisterCustomRenderer(name string, renderer Renderer) {
    rendererRegistry[name] = renderer
}
```

### 3. Report Filters

```go
type ReportFilter struct {
    Severity      []string
    Categories    []string
    DateRange     *DateRange
    Tags          []string
    CustomFilters map[string]interface{}
}

func (rf *ReportFilter) Apply(findings []Finding) []Finding {
    filtered := []Finding{}
    for _, finding := range findings {
        if rf.matches(finding) {
            filtered = append(filtered, finding)
        }
    }
    return filtered
}
```

## Localization Support

### 1. Multi-language Reports

```go
type Localizer struct {
    bundles map[string]*i18n.Bundle
}

func (l *Localizer) Localize(key string, locale string, data interface{}) string {
    bundle := l.bundles[locale]
    if bundle == nil {
        bundle = l.bundles["en"] // fallback
    }
    
    return bundle.MustLocalize(&i18n.LocalizeConfig{
        MessageID: key,
        TemplateData: data,
    })
}
```

### 2. Supported Languages

- English (en)
- Spanish (es)
- French (fr)
- German (de)
- Japanese (ja)
- Chinese (zh)

## Performance Optimization

### 1. Streaming Large Reports

```go
type StreamingRenderer interface {
    StartRender(w io.Writer, report *Report, options RenderOptions) error
    RenderFinding(w io.Writer, finding Finding) error
    FinishRender(w io.Writer) error
}
```

### 2. Caching

```go
type ReportCache struct {
    cache *lru.Cache
}

func (rc *ReportCache) GetOrGenerate(key string, generator func() (*Report, error)) (*Report, error) {
    if cached, ok := rc.cache.Get(key); ok {
        return cached.(*Report), nil
    }
    
    report, err := generator()
    if err != nil {
        return nil, err
    }
    
    rc.cache.Add(key, report)
    return report, nil
}
```

## CLI Integration

```bash
# Generate report in specific format
llm-redteam report generate --format pdf --output report.pdf

# Generate multiple formats
llm-redteam report generate --format json,html,pdf --output-dir reports/

# Use custom template
llm-redteam report generate --format html --template custom.html

# Apply filters
llm-redteam report generate --severity critical,high --category "Prompt Injection"

# Localized report
llm-redteam report generate --format pdf --locale es
```

## API Integration

```yaml
openapi: 3.0.0
paths:
  /api/v1/reports/{id}/export:
    get:
      parameters:
        - name: format
          in: query
          required: true
          schema:
            type: string
            enum: [json, csv, html, pdf, markdown, excel]
        - name: locale
          in: query
          schema:
            type: string
        - name: filters
          in: query
          schema:
            type: object
      responses:
        200:
          description: Generated report
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Report'
            text/csv:
              schema:
                type: string
            text/html:
              schema:
                type: string
            application/pdf:
              schema:
                type: string
                format: binary
```