# LLMrecon API Quick Reference

## Authentication

### API Key
```bash
curl -H "X-API-Key: your-api-key" https://api.llmredteam.com/api/v1/endpoint
```

### JWT Token
```bash
# Login
TOKEN=$(curl -s -X POST https://api.llmredteam.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}' | jq -r '.data.token')

# Use token
curl -H "Authorization: Bearer $TOKEN" https://api.llmredteam.com/api/v1/endpoint
```

## Common Operations

### Create and Monitor a Scan
```bash
# Create scan
SCAN_ID=$(curl -s -X POST https://api.llmredteam.com/api/v1/scans \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "target": {"type": "endpoint", "url": "https://api.example.com/chat"},
    "templates": ["prompt-injection", "data-leakage"]
  }' | jq -r '.data.id')

# Check status
curl -s https://api.llmredteam.com/api/v1/scans/$SCAN_ID \
  -H "X-API-Key: $API_KEY" | jq '.data.status'

# Get results
curl -s https://api.llmredteam.com/api/v1/scans/$SCAN_ID/results \
  -H "X-API-Key: $API_KEY" | jq '.data.summary'
```

### Manage API Keys
```bash
# Create API key (requires JWT)
curl -X POST https://api.llmredteam.com/api/v1/auth/keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Key",
    "scopes": ["scan:write", "template:read"],
    "rate_limit": 100
  }'

# List API keys
curl https://api.llmredteam.com/api/v1/auth/keys \
  -H "Authorization: Bearer $TOKEN"

# Revoke API key
curl -X DELETE https://api.llmredteam.com/api/v1/auth/keys/$KEY_ID \
  -H "Authorization: Bearer $TOKEN"
```

## Response Format

### Success Response
```json
{
  "success": true,
  "data": {
    // Response data
  },
  "meta": {
    "version": "1.0.0",
    "request_id": "req-123",
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description",
    "details": "Additional details"
  },
  "meta": {
    "version": "1.0.0",
    "request_id": "req-123"
  }
}
```

## Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad Request - Invalid parameters |
| 401 | Unauthorized - Invalid or missing auth |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Resource already exists |
| 429 | Too Many Requests - Rate limited |
| 500 | Internal Server Error |

## Rate Limiting

- Default: 60 requests/minute per API key
- Headers included in response:
  - `X-RateLimit-Limit`: Max requests per minute
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Unix timestamp of reset

## Pagination

Use `page` and `per_page` query parameters:
```
GET /api/v1/scans?page=2&per_page=50
```

## Filtering

Most list endpoints support filtering:
```
GET /api/v1/scans?status=running
GET /api/v1/templates?category=prompt-injection
GET /api/v1/auth/keys?active=true
```

## Webhooks

Configure webhooks for async notifications:
```json
{
  "url": "https://your-server.com/webhook",
  "events": ["scan.completed", "scan.failed"],
  "secret": "webhook-secret"
}
```

## API Scopes

| Scope | Description |
|-------|-------------|
| `scan:read` | View scan data |
| `scan:write` | Create/manage scans |
| `template:read` | View templates |
| `template:write` | Manage templates |
| `module:read` | View modules |
| `module:write` | Configure modules |
| `system:update` | System updates |
| `admin` | Full access |

## SDK Examples

### Python
```python
from llmrecon import Client

client = Client(api_key="your-api-key")
scan = client.scans.create(
    target={"type": "endpoint", "url": "https://api.example.com"},
    templates=["prompt-injection"]
)
print(f"Scan ID: {scan.id}")
```

### JavaScript
```javascript
const { LLMRedTeamClient } = require('@LLMrecon/sdk');

const client = new LLMRedTeamClient({ apiKey: 'your-api-key' });
const scan = await client.scans.create({
  target: { type: 'endpoint', url: 'https://api.example.com' },
  templates: ['prompt-injection']
});
console.log(`Scan ID: ${scan.id}`);
```

## Useful Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | API health check |
| GET | `/version` | Version info |
| POST | `/auth/login` | User login |
| POST | `/auth/register` | User registration |
| GET | `/auth/profile` | Get user profile |
| POST | `/auth/keys` | Create API key |
| GET | `/auth/keys` | List API keys |
| POST | `/scans` | Create scan |
| GET | `/scans` | List scans |
| GET | `/scans/{id}` | Get scan details |
| DELETE | `/scans/{id}` | Cancel scan |
| GET | `/scans/{id}/results` | Get scan results |
| GET | `/templates` | List templates |
| GET | `/modules` | List modules |
| POST | `/compliance/report` | Generate report |

## Tips

1. **Always use HTTPS** in production
2. **Store API keys securely** - never in code
3. **Implement retry logic** for transient errors
4. **Use pagination** for large result sets
5. **Monitor rate limits** to avoid 429 errors
6. **Cache responses** when appropriate
7. **Use appropriate scopes** - least privilege
8. **Set request timeouts** to handle slow responses