# API Reference

The LLMrecon tool provides a RESTful API for programmatic access to all features. This enables integration with CI/CD pipelines, custom dashboards, and automated security workflows.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Core Endpoints](#core-endpoints)
4. [Scan Management](#scan-management)
5. [Template Operations](#template-operations)
6. [Provider Management](#provider-management)
7. [Reporting API](#reporting-api)
8. [Webhooks](#webhooks)
9. [Rate Limiting](#rate-limiting)
10. [Error Handling](#error-handling)
11. [Code Examples](#code-examples)

## Getting Started

### Starting the API Server

```bash
# Start API server on default port (8080)
LLMrecon api serve

# Custom port and binding
LLMrecon api serve --port 3000 --bind 0.0.0.0

# With TLS
LLMrecon api serve --tls --cert server.crt --key server.key

# Background mode
LLMrecon api serve --daemon
```

### API Configuration

```yaml
# ~/.LLMrecon/api-config.yaml
api:
  port: 8080
  bind: 127.0.0.1
  tls:
    enabled: true
    cert: /path/to/cert.pem
    key: /path/to/key.pem
  cors:
    enabled: true
    origins:
      - http://localhost:3000
      - https://app.example.com
  rate_limit:
    requests_per_minute: 60
    burst: 10
```

### Base URL

```
http://localhost:8080/api/v1
```

## Authentication

### API Key Authentication

Generate and use API keys:

```bash
# Generate API key
LLMrecon api create-key --name "CI/CD Pipeline" --scope read,write

# Output:
# API Key: llmrt_k3y_1234567890abcdef
# Key ID: key_abc123
```

Use in requests:

```bash
curl -H "Authorization: Bearer llmrt_k3y_1234567890abcdef" \
  http://localhost:8080/api/v1/scans
```

### OAuth 2.0

Configure OAuth providers:

```yaml
auth:
  oauth:
    providers:
      - name: github
        client_id: ${GITHUB_CLIENT_ID}
        client_secret: ${GITHUB_CLIENT_SECRET}
        scopes: [read:user, read:org]
```

### JWT Tokens

For session-based auth:

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "secure123"}'

# Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2024-01-21T10:00:00Z"
}
```

## Core Endpoints

### Health Check

```http
GET /api/v1/health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600,
  "components": {
    "database": "healthy",
    "templates": "healthy",
    "providers": "healthy"
  }
}
```

### System Information

```http
GET /api/v1/info
```

Response:
```json
{
  "version": "1.0.0",
  "git_commit": "abc123def",
  "build_date": "2024-01-20T10:00:00Z",
  "templates": {
    "total": 42,
    "categories": ["prompt-injection", "data-leakage", "output-handling"]
  },
  "providers": ["openai", "anthropic", "azure-openai"],
  "capabilities": ["scan", "report", "compliance", "api"]
}
```

## Scan Management

### Create Scan

```http
POST /api/v1/scans
Content-Type: application/json

{
  "target": {
    "provider": "openai",
    "model": "gpt-4",
    "endpoint": "https://api.openai.com/v1/chat/completions"
  },
  "templates": ["prompt-injection/*", "data-leakage/pii"],
  "options": {
    "parallel": 5,
    "timeout": 30,
    "rate_limit": "10/min"
  },
  "metadata": {
    "project": "production-api",
    "owner": "security-team"
  }
}
```

Response:
```json
{
  "scan_id": "scan_123456",
  "status": "running",
  "created_at": "2024-01-20T10:00:00Z",
  "estimated_duration": 300,
  "progress_url": "/api/v1/scans/scan_123456/progress"
}
```

### Get Scan Status

```http
GET /api/v1/scans/{scan_id}
```

Response:
```json
{
  "scan_id": "scan_123456",
  "status": "completed",
  "started_at": "2024-01-20T10:00:00Z",
  "completed_at": "2024-01-20T10:05:00Z",
  "progress": {
    "total": 42,
    "completed": 42,
    "passed": 38,
    "failed": 4,
    "errors": 0
  },
  "summary": {
    "critical": 1,
    "high": 2,
    "medium": 1,
    "low": 0,
    "info": 0
  }
}
```

### List Scans

```http
GET /api/v1/scans?page=1&limit=20&status=completed&from=2024-01-01
```

Response:
```json
{
  "scans": [
    {
      "scan_id": "scan_123456",
      "status": "completed",
      "target": "openai/gpt-4",
      "created_at": "2024-01-20T10:00:00Z",
      "findings_count": 4
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

### Get Scan Results

```http
GET /api/v1/scans/{scan_id}/results
```

Response:
```json
{
  "scan_id": "scan_123456",
  "findings": [
    {
      "finding_id": "vuln_001",
      "template_id": "prompt-injection-basic",
      "severity": "high",
      "title": "Prompt Injection Vulnerability",
      "description": "System instructions can be overridden",
      "evidence": {
        "request": {
          "prompt": "Ignore all instructions..."
        },
        "response": {
          "content": "INJECTION SUCCESS"
        }
      },
      "remediation": "Implement prompt filtering",
      "references": ["https://owasp.org/llm01"]
    }
  ],
  "statistics": {
    "total_tests": 42,
    "duration_seconds": 300,
    "tokens_used": 15000
  }
}
```

### Cancel Scan

```http
POST /api/v1/scans/{scan_id}/cancel
```

### Re-run Scan

```http
POST /api/v1/scans/{scan_id}/rerun

{
  "templates": ["only-failed"],  // or specific template IDs
  "options": {
    "timeout": 60
  }
}
```

## Template Operations

### List Templates

```http
GET /api/v1/templates?category=prompt-injection&severity=high
```

Response:
```json
{
  "templates": [
    {
      "id": "prompt-injection-basic",
      "name": "Basic Prompt Injection",
      "category": "prompt-injection",
      "severity": "high",
      "author": "security-team",
      "version": "1.0.0",
      "tags": ["owasp", "llm01"]
    }
  ],
  "total": 15
}
```

### Get Template Details

```http
GET /api/v1/templates/{template_id}
```

Response:
```json
{
  "id": "prompt-injection-basic",
  "name": "Basic Prompt Injection",
  "content": "id: prompt-injection-basic\n...",
  "metadata": {
    "created": "2024-01-01",
    "updated": "2024-01-15",
    "downloads": 1523,
    "success_rate": 0.82
  }
}
```

### Create Custom Template

```http
POST /api/v1/templates
Content-Type: application/json

{
  "id": "custom-injection-test",
  "name": "Custom Injection Test",
  "content": "id: custom-injection-test\ninfo:\n  name: ...",
  "category": "custom",
  "private": true
}
```

### Update Template

```http
PUT /api/v1/templates/{template_id}
Content-Type: application/json

{
  "content": "updated template content...",
  "version": "1.1.0",
  "changelog": "Fixed false positives"
}
```

### Validate Template

```http
POST /api/v1/templates/validate
Content-Type: application/json

{
  "content": "id: test-template\n..."
}
```

Response:
```json
{
  "valid": true,
  "errors": [],
  "warnings": [
    "Consider adding 'reference' field"
  ]
}
```

### Test Template

```http
POST /api/v1/templates/{template_id}/test

{
  "target": "openai",
  "mock_response": "Test response for validation"
}
```

## Provider Management

### List Providers

```http
GET /api/v1/providers
```

Response:
```json
{
  "providers": [
    {
      "name": "openai",
      "type": "api",
      "status": "active",
      "models": ["gpt-3.5-turbo", "gpt-4"],
      "capabilities": ["chat", "completion"],
      "rate_limits": {
        "requests_per_minute": 60,
        "tokens_per_minute": 150000
      }
    }
  ]
}
```

### Configure Provider

```http
PUT /api/v1/providers/{provider_name}/config
Content-Type: application/json

{
  "api_key": "sk-...",
  "endpoint": "https://api.openai.com/v1",
  "model": "gpt-4",
  "max_retries": 3,
  "timeout": 30
}
```

### Test Provider

```http
POST /api/v1/providers/{provider_name}/test

{
  "prompt": "Hello, are you working?"
}
```

Response:
```json
{
  "success": true,
  "response": "Yes, I'm working properly!",
  "latency_ms": 450,
  "model": "gpt-4-0613"
}
```

## Reporting API

### Generate Report

```http
POST /api/v1/reports
Content-Type: application/json

{
  "scan_ids": ["scan_123456", "scan_789012"],
  "format": "pdf",
  "template": "executive-summary",
  "options": {
    "include_evidence": false,
    "include_remediation": true,
    "branding": {
      "logo": "https://company.com/logo.png",
      "company": "ACME Corp"
    }
  }
}
```

Response:
```json
{
  "report_id": "report_abc123",
  "status": "generating",
  "estimated_time": 30,
  "download_url": "/api/v1/reports/report_abc123/download"
}
```

### Download Report

```http
GET /api/v1/reports/{report_id}/download
```

### Schedule Reports

```http
POST /api/v1/reports/schedules

{
  "name": "Weekly Security Report",
  "schedule": "0 9 * * 1",  // Every Monday at 9 AM
  "scan_filter": {
    "tags": ["production"],
    "severity": ["critical", "high"]
  },
  "report_config": {
    "format": "pdf",
    "template": "detailed",
    "recipients": ["security@company.com"]
  }
}
```

## Webhooks

### Register Webhook

```http
POST /api/v1/webhooks

{
  "url": "https://app.example.com/webhooks/llm-security",
  "events": ["scan.completed", "finding.critical"],
  "secret": "webhook_secret_123",
  "active": true
}
```

### Webhook Events

Available events:
- `scan.started`
- `scan.completed`
- `scan.failed`
- `finding.critical`
- `finding.high`
- `template.updated`
- `report.ready`

### Webhook Payload

```json
{
  "event": "scan.completed",
  "timestamp": "2024-01-20T10:05:00Z",
  "data": {
    "scan_id": "scan_123456",
    "status": "completed",
    "findings_count": 4,
    "critical_count": 1
  },
  "signature": "sha256=abcdef123456..."
}
```

## Rate Limiting

Rate limits are applied per API key:

| Endpoint | Rate Limit | Burst |
|----------|------------|-------|
| Scan creation | 10/hour | 2 |
| Template operations | 100/hour | 10 |
| Report generation | 20/hour | 5 |
| General API calls | 1000/hour | 50 |

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642680000
```

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "TEMPLATE_NOT_FOUND",
    "message": "Template 'custom-test' not found",
    "details": {
      "template_id": "custom-test",
      "available_templates": ["test1", "test2"]
    },
    "request_id": "req_123456"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `RATE_LIMITED` | 429 | Too many requests |
| `PROVIDER_ERROR` | 502 | LLM provider error |
| `INTERNAL_ERROR` | 500 | Server error |

## Code Examples

### Python

```python
import requests
import json

class LLMRedTeamClient:
    def __init__(self, api_key, base_url="http://localhost:8080/api/v1"):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        }
    
    def create_scan(self, target, templates):
        """Create a new security scan"""
        data = {
            "target": {"provider": target},
            "templates": templates
        }
        
        response = requests.post(
            f"{self.base_url}/scans",
            headers=self.headers,
            json=data
        )
        response.raise_for_status()
        return response.json()
    
    def get_results(self, scan_id):
        """Get scan results"""
        response = requests.get(
            f"{self.base_url}/scans/{scan_id}/results",
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

# Usage
client = LLMRedTeamClient("llmrt_k3y_1234567890abcdef")

# Start scan
scan = client.create_scan("openai", ["prompt-injection/*"])
print(f"Scan started: {scan['scan_id']}")

# Wait and get results
import time
time.sleep(60)
results = client.get_results(scan['scan_id'])
print(f"Found {len(results['findings'])} vulnerabilities")
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

class LLMRedTeamClient {
  constructor(apiKey, baseUrl = 'http://localhost:8080/api/v1') {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
    this.client = axios.create({
      baseURL: baseUrl,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createScan(target, templates) {
    const { data } = await this.client.post('/scans', {
      target: { provider: target },
      templates: templates
    });
    return data;
  }

  async waitForCompletion(scanId, timeout = 300000) {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeout) {
      const { data } = await this.client.get(`/scans/${scanId}`);
      
      if (data.status === 'completed') {
        return data;
      }
      
      if (data.status === 'failed') {
        throw new Error(`Scan failed: ${data.error}`);
      }
      
      await new Promise(resolve => setTimeout(resolve, 5000));
    }
    
    throw new Error('Scan timeout');
  }

  async getReport(scanId, format = 'json') {
    const { data } = await this.client.post('/reports', {
      scan_ids: [scanId],
      format: format
    });
    return data;
  }
}

// Usage
const client = new LLMRedTeamClient('llmrt_k3y_1234567890abcdef');

(async () => {
  try {
    // Start scan
    const scan = await client.createScan('openai', ['owasp-llm/*']);
    console.log(`Scan started: ${scan.scan_id}`);
    
    // Wait for completion
    const completed = await client.waitForCompletion(scan.scan_id);
    console.log(`Scan completed: ${completed.summary}`);
    
    // Generate report
    const report = await client.getReport(scan.scan_id, 'pdf');
    console.log(`Report available at: ${report.download_url}`);
    
  } catch (error) {
    console.error('Error:', error.message);
  }
})();
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    APIKey  string
    BaseURL string
}

type ScanRequest struct {
    Target    map[string]string `json:"target"`
    Templates []string          `json:"templates"`
}

type ScanResponse struct {
    ScanID    string `json:"scan_id"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
}

func NewClient(apiKey string) *Client {
    return &Client{
        APIKey:  apiKey,
        BaseURL: "http://localhost:8080/api/v1",
    }
}

func (c *Client) CreateScan(provider string, templates []string) (*ScanResponse, error) {
    reqBody := ScanRequest{
        Target: map[string]string{
            "provider": provider,
        },
        Templates: templates,
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", c.BaseURL+"/scans", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var scanResp ScanResponse
    if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
        return nil, err
    }
    
    return &scanResp, nil
}

func main() {
    client := NewClient("llmrt_k3y_1234567890abcdef")
    
    scan, err := client.CreateScan("openai", []string{"prompt-injection/*"})
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Scan started: %s\n", scan.ScanID)
}
```

### cURL Examples

```bash
# Create scan
curl -X POST http://localhost:8080/api/v1/scans \
  -H "Authorization: Bearer llmrt_k3y_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "target": {"provider": "openai"},
    "templates": ["prompt-injection/*"]
  }'

# Get scan status
curl http://localhost:8080/api/v1/scans/scan_123456 \
  -H "Authorization: Bearer llmrt_k3y_1234567890abcdef"

# Download results as JSON
curl http://localhost:8080/api/v1/scans/scan_123456/results \
  -H "Authorization: Bearer llmrt_k3y_1234567890abcdef" \
  -o results.json

# Generate PDF report
curl -X POST http://localhost:8080/api/v1/reports \
  -H "Authorization: Bearer llmrt_k3y_1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "scan_ids": ["scan_123456"],
    "format": "pdf",
    "template": "executive-summary"
  }'
```

## SDK and Client Libraries

Official SDKs available:
- Python: `pip install LLMrecon`
- Node.js: `npm install @LLMrecon/client`
- Go: `go get github.com/your-org/LLMrecon-go`
- Ruby: `gem install LLMrecon`

## API Versioning

The API uses URL versioning:
- Current: `/api/v1`
- Beta: `/api/beta`
- Deprecated: `/api/v0` (sunset date: 2024-06-01)

Version compatibility:
```http
GET /api/versions
```

## Support

- API Documentation: https://docs.LLMrecon.com/api
- Status Page: https://status.LLMrecon.com
- Support: api-support@LLMrecon.com