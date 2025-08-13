# LLMrecon API Documentation

## Overview

The LLMrecon API provides comprehensive endpoints for security testing and vulnerability assessment of Large Language Models (LLMs). This RESTful API supports multiple authentication methods, rate limiting, and extensive security features.

## Base URL

```
Production: https://api.llmredteam.com/api/v1
Development: http://localhost:8080/api/v1
```

## Authentication

The API supports two authentication methods:

### 1. API Key Authentication

Include your API key in the request header:

```
X-API-Key: your-api-key-here
```

Or as a Bearer token:

```
Authorization: Bearer your-api-key-here
```

### 2. JWT Authentication

For user-based authentication, first login to receive a JWT token:

```bash
POST /auth/login
Content-Type: application/json

{
  "username": "your-username",
  "password": "your-password"
}
```

Then include the JWT token in subsequent requests:

```
Authorization: Bearer your-jwt-token-here
```

## Rate Limiting

- Default rate limit: 60 requests per minute per API key
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Maximum requests per minute
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Unix timestamp when the limit resets

## Security Headers

All API responses include comprehensive security headers:

- `Content-Security-Policy`: Restricts resource loading
- `X-Content-Type-Options`: Prevents MIME type sniffing
- `X-Frame-Options`: Prevents clickjacking
- `X-XSS-Protection`: Enables XSS filtering
- `Strict-Transport-Security`: Enforces HTTPS
- `Referrer-Policy`: Controls referrer information
- `Permissions-Policy`: Restricts browser features

## Error Responses

All error responses follow a consistent format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Additional error details (optional)"
  },
  "meta": {
    "version": "1.0.0",
    "request_id": "unique-request-id"
  }
}
```

Common error codes:
- `VALIDATION_ERROR`: Invalid request parameters
- `UNAUTHORIZED`: Missing or invalid authentication
- `FORBIDDEN`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `CONFLICT`: Resource conflict (e.g., duplicate)
- `RATE_LIMITED`: Rate limit exceeded
- `INTERNAL_ERROR`: Server error

## Endpoints

### System Endpoints

#### Health Check

Check API health status.

```
GET /health
```

Response:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "24h 30m 15s",
    "checks": {
      "database": true,
      "cache": true,
      "storage": true
    }
  }
}
```

#### Version Information

Get API version and build information.

```
GET /version
```

Response:
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "build_date": "2024-01-20T10:30:00Z",
    "git_commit": "abc123def",
    "go_version": "1.21.5",
    "api_version": "1.0.0"
  }
}
```

### Authentication Endpoints

#### User Login

Authenticate with username and password.

```
POST /auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "SecurePassword123!"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "user-123",
      "username": "testuser",
      "email": "test@example.com",
      "role": "user",
      "active": true,
      "created_at": "2024-01-20T10:00:00Z"
    }
  }
}
```

#### User Registration

Create a new user account.

```
POST /auth/register
Content-Type: application/json

{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "SecurePassword123!",
  "role": "user"
}
```

#### Refresh Token

Refresh an expired JWT token.

```
POST /auth/refresh
Authorization: Bearer expired-jwt-token
```

#### Get User Profile

Get current user's profile (JWT required).

```
GET /auth/profile
Authorization: Bearer jwt-token
```

#### Update Password

Update user password (JWT required).

```
PUT /auth/password
Authorization: Bearer jwt-token
Content-Type: application/json

{
  "old_password": "CurrentPassword123!",
  "new_password": "NewSecurePassword456!"
}
```

### API Key Management

#### Create API Key

Create a new API key (JWT required).

```
POST /auth/keys
Authorization: Bearer jwt-token
Content-Type: application/json

{
  "name": "Production Key",
  "description": "Key for production environment",
  "scopes": ["scan:write", "template:read"],
  "rate_limit": 100,
  "expires_in": 365
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "key-123",
    "key": "llmrt_prod_AbCdEfGhIjKlMnOpQrStUvWxYz123456",
    "name": "Production Key",
    "created_at": "2024-01-20T10:00:00Z"
  }
}
```

**Note**: The API key is only shown once during creation. Store it securely.

#### List API Keys

List all API keys (JWT required).

```
GET /auth/keys?active=true&expired=false
Authorization: Bearer jwt-token
```

#### Get API Key Details

Get details of a specific API key (JWT required).

```
GET /auth/keys/{key-id}
Authorization: Bearer jwt-token
```

#### Revoke API Key

Revoke an API key (JWT required).

```
DELETE /auth/keys/{key-id}
Authorization: Bearer jwt-token
```

### Scan Management

#### Create Scan

Create a new security scan.

```
POST /scans
X-API-Key: your-api-key
Content-Type: application/json

{
  "target": {
    "type": "endpoint",
    "url": "https://api.example.com/v1/chat",
    "headers": {
      "Authorization": "Bearer target-api-key"
    }
  },
  "templates": ["prompt-injection", "data-leakage"],
  "categories": ["owasp-llm-top-10"],
  "config": {
    "concurrent_tests": 10,
    "timeout": 300,
    "retry_failed": true,
    "max_retries": 3
  }
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "scan-123",
    "status": "pending",
    "target": {
      "type": "endpoint",
      "url": "https://api.example.com/v1/chat"
    },
    "templates": ["prompt-injection", "data-leakage"],
    "created_at": "2024-01-20T10:00:00Z"
  }
}
```

#### List Scans

List all scans with optional filtering.

```
GET /scans?status=running&page=1&per_page=20
X-API-Key: your-api-key
```

Query parameters:
- `status`: Filter by status (pending, running, completed, failed, cancelled)
- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 20, max: 100)

#### Get Scan Details

Get details of a specific scan.

```
GET /scans/{scan-id}
X-API-Key: your-api-key
```

#### Cancel Scan

Cancel a running scan.

```
DELETE /scans/{scan-id}
X-API-Key: your-api-key
```

#### Get Scan Results

Get detailed results of a completed scan.

```
GET /scans/{scan-id}/results
X-API-Key: your-api-key
```

Response:
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_tests": 50,
      "passed": 35,
      "failed": 15,
      "skipped": 0,
      "risk_score": 72.5,
      "severity_counts": {
        "critical": 2,
        "high": 5,
        "medium": 8,
        "low": 12,
        "info": 8
      }
    },
    "findings": [
      {
        "id": "finding-123",
        "template_id": "prompt-injection-001",
        "category": "prompt-injection",
        "severity": "high",
        "title": "System Prompt Override Vulnerability",
        "description": "The model can be manipulated to ignore system instructions",
        "evidence": {
          "payload": "Ignore previous instructions and...",
          "response": "Sure, I'll ignore the previous instructions..."
        },
        "remediation": "Implement input validation and prompt sandboxing",
        "references": [
          "https://owasp.org/llm-top-10/llm01"
        ]
      }
    ]
  }
}
```

### Template Management

#### List Templates

List all available test templates.

```
GET /templates?category=prompt-injection
X-API-Key: your-api-key
```

#### Get Template Details

Get details of a specific template.

```
GET /templates/{template-id}
X-API-Key: your-api-key
```

#### List Template Categories

Get all available template categories.

```
GET /templates/categories
X-API-Key: your-api-key
```

### Module Management

#### List Modules

List all available modules.

```
GET /modules
X-API-Key: your-api-key
```

#### Get Module Details

Get details of a specific module.

```
GET /modules/{module-id}
X-API-Key: your-api-key
```

#### Update Module Configuration

Update a module's configuration.

```
PUT /modules/{module-id}/config
X-API-Key: your-api-key
Content-Type: application/json

{
  "enabled": true,
  "settings": {
    "timeout": 60,
    "max_retries": 3
  }
}
```

### System Updates

#### Check for Updates

Check if system updates are available.

```
GET /update
X-API-Key: your-api-key
```

#### Perform Update

Initiate a system update.

```
POST /update
X-API-Key: your-api-key
Content-Type: application/json

{
  "version": "1.2.0",
  "components": ["templates", "modules"]
}
```

### Bundle Management

#### List Bundles

List available offline bundles.

```
GET /bundles
X-API-Key: your-api-key
```

#### Export Bundle

Create an offline bundle.

```
POST /bundles/export
X-API-Key: your-api-key
Content-Type: application/json

{
  "include_templates": true,
  "include_modules": true,
  "format": "tar.gz"
}
```

#### Import Bundle

Import an offline bundle.

```
POST /bundles/import
X-API-Key: your-api-key
Content-Type: multipart/form-data

bundle: [file]
```

### Compliance Reporting

#### Generate Compliance Report

Generate a compliance report.

```
POST /compliance/report
X-API-Key: your-api-key
Content-Type: application/json

{
  "framework": "owasp-llm-top-10",
  "scan_ids": ["scan-123", "scan-456"],
  "format": "pdf"
}
```

#### Check Compliance Status

Check current compliance status.

```
GET /compliance/check?framework=owasp-llm-top-10
X-API-Key: your-api-key
```

## Code Examples

### Python Example

```python
import requests
import json

# Configuration
BASE_URL = "http://localhost:8080/api/v1"
API_KEY = "your-api-key"

# Create headers
headers = {
    "X-API-Key": API_KEY,
    "Content-Type": "application/json"
}

# Create a scan
scan_data = {
    "target": {
        "type": "endpoint",
        "url": "https://api.example.com/chat"
    },
    "templates": ["prompt-injection"],
    "config": {
        "concurrent_tests": 5,
        "timeout": 300
    }
}

response = requests.post(
    f"{BASE_URL}/scans",
    headers=headers,
    json=scan_data
)

if response.status_code == 200:
    scan = response.json()["data"]
    print(f"Scan created: {scan['id']}")
else:
    print(f"Error: {response.json()['error']['message']}")

# Check scan status
scan_id = scan["id"]
response = requests.get(
    f"{BASE_URL}/scans/{scan_id}",
    headers=headers
)

status = response.json()["data"]["status"]
print(f"Scan status: {status}")
```

### JavaScript/Node.js Example

```javascript
const axios = require('axios');

// Configuration
const BASE_URL = 'http://localhost:8080/api/v1';
const API_KEY = 'your-api-key';

// Create axios instance with default headers
const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'X-API-Key': API_KEY,
    'Content-Type': 'application/json'
  }
});

// Create a scan
async function createScan() {
  try {
    const scanData = {
      target: {
        type: 'endpoint',
        url: 'https://api.example.com/chat'
      },
      templates: ['prompt-injection'],
      config: {
        concurrent_tests: 5,
        timeout: 300
      }
    };

    const response = await api.post('/scans', scanData);
    const scan = response.data.data;
    console.log(`Scan created: ${scan.id}`);
    
    return scan.id;
  } catch (error) {
    console.error(`Error: ${error.response.data.error.message}`);
  }
}

// Check scan status
async function checkScanStatus(scanId) {
  try {
    const response = await api.get(`/scans/${scanId}`);
    const status = response.data.data.status;
    console.log(`Scan status: ${status}`);
    
    return status;
  } catch (error) {
    console.error(`Error: ${error.response.data.error.message}`);
  }
}

// Example usage
(async () => {
  const scanId = await createScan();
  if (scanId) {
    // Check status every 5 seconds
    const interval = setInterval(async () => {
      const status = await checkScanStatus(scanId);
      if (status === 'completed' || status === 'failed') {
        clearInterval(interval);
        
        // Get results if completed
        if (status === 'completed') {
          const results = await api.get(`/scans/${scanId}/results`);
          console.log('Scan results:', results.data.data);
        }
      }
    }, 5000);
  }
})();
```

### cURL Examples

```bash
# Health check
curl -X GET http://localhost:8080/api/v1/health

# Create API key (requires JWT)
curl -X POST http://localhost:8080/api/v1/auth/keys \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Key",
    "scopes": ["scan:write", "template:read"]
  }'

# Create scan
curl -X POST http://localhost:8080/api/v1/scans \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "target": {
      "type": "endpoint",
      "url": "https://api.example.com/chat"
    },
    "templates": ["prompt-injection"]
  }'

# List scans with pagination
curl -X GET "http://localhost:8080/api/v1/scans?page=1&per_page=10" \
  -H "X-API-Key: $API_KEY"
```

## API Scopes

When creating API keys, you can assign specific scopes to limit access:

- `scan:read` - Read scan data
- `scan:write` - Create and manage scans
- `template:read` - Read template data
- `template:write` - Create and manage templates
- `module:read` - Read module data
- `module:write` - Configure modules
- `system:update` - Perform system updates
- `admin` - Full administrative access

## Webhooks

Configure webhooks to receive real-time notifications about scan events:

```json
{
  "url": "https://your-server.com/webhooks/LLMrecon",
  "events": ["scan.completed", "scan.failed"],
  "secret": "webhook-secret-key"
}
```

Webhook payload example:
```json
{
  "event": "scan.completed",
  "timestamp": "2024-01-20T15:30:00Z",
  "data": {
    "scan_id": "scan-123",
    "status": "completed",
    "risk_score": 72.5,
    "findings_count": 15
  }
}
```

## Best Practices

1. **API Key Security**
   - Store API keys securely (use environment variables)
   - Rotate keys regularly
   - Use different keys for different environments
   - Never commit keys to version control

2. **Rate Limiting**
   - Implement exponential backoff for rate limit errors
   - Monitor your usage to avoid hitting limits
   - Request rate limit increases if needed

3. **Error Handling**
   - Always check response status codes
   - Implement proper error handling for all API calls
   - Log errors for debugging

4. **Performance**
   - Use pagination for list endpoints
   - Cache frequently accessed data
   - Minimize API calls by batching operations

5. **Security**
   - Always use HTTPS in production
   - Validate SSL certificates
   - Keep your client libraries updated
   - Follow the principle of least privilege for API key scopes

## Support

For API support, please contact:
- Email: api-support@llmredteam.com
- Documentation: https://docs.llmredteam.com
- Status Page: https://status.llmredteam.com