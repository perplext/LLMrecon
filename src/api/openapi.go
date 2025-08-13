package api

import (
	"encoding/json"
	"net/http"
)

// OpenAPI specification version
const OpenAPIVersion = "3.0.3"

// OpenAPISpec generates the OpenAPI specification
func GenerateOpenAPISpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": OpenAPIVersion,
		"info": map[string]interface{}{
			"title":       "LLMrecon API",
			"description": "Comprehensive API for LLM security testing and vulnerability assessment",
			"version":     APIVersion,
			"contact": map[string]interface{}{
				"name":  "LLMrecon",
				"email": "support@llmredteam.com",
			},
			"license": map[string]interface{}{
				"name": "MIT",
				"url":  "https://opensource.org/licenses/MIT",
			},
		},
		"servers": []map[string]interface{}{
			{
				"url":         "https://api.llmredteam.com/api/v1",
				"description": "Production server",
			},
			{
				"url":         "http://localhost:8080/api/v1",
				"description": "Development server",
			},
		},
		"security": []map[string]interface{}{
			{"apiKey": []string{}},
			{"bearerAuth": []string{}},
		},
		"tags": []map[string]interface{}{
			{
				"name":        "Authentication",
				"description": "Authentication and authorization endpoints",
			},
			{
				"name":        "Scans",
				"description": "Security scan management",
			},
			{
				"name":        "Templates",
				"description": "Test template management",
			},
			{
				"name":        "Modules",
				"description": "Module configuration",
			},
			{
				"name":        "Updates",
				"description": "System updates and versioning",
			},
			{
				"name":        "Compliance",
				"description": "Compliance reporting",
			},
		},
		"paths":      generatePaths(),
		"components": generateComponents(),
	}
}

// generatePaths generates the API paths section
func generatePaths() map[string]interface{} {
	return map[string]interface{}{
		// Health & Version
		"/health": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Health check",
				"description": "Check if the API is healthy and operational",
				"tags":        []string{"System"},
				"operationId": "getHealth",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "API is healthy",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/HealthResponse",
								},
							},
						},
					},
				},
			},
		},
		"/version": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Get version information",
				"description": "Get API version and build information",
				"tags":        []string{"System"},
				"operationId": "getVersion",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Version information",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/VersionInfo",
								},
							},
						},
					},
				},
			},
		},
		
		// Authentication
		"/auth/login": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "User login",
				"description": "Authenticate with username and password to receive JWT token",
				"tags":        []string{"Authentication"},
				"operationId": "login",
				"security":    []map[string]interface{}{},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/LoginRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Login successful",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/LoginResponse",
								},
							},
						},
					},
					"401": map[string]interface{}{
						"$ref": "#/components/responses/UnauthorizedError",
					},
				},
			},
		},
		"/auth/register": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "Register new user",
				"description": "Create a new user account",
				"tags":        []string{"Authentication"},
				"operationId": "register",
				"security":    []map[string]interface{}{},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/CreateUserRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "User created successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/User",
								},
							},
						},
					},
					"409": map[string]interface{}{
						"$ref": "#/components/responses/ConflictError",
					},
				},
			},
		},
		"/auth/refresh": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "Refresh JWT token",
				"description": "Refresh an expired JWT token",
				"tags":        []string{"Authentication"},
				"operationId": "refreshToken",
				"security": []map[string]interface{}{
					{"bearerAuth": []string{}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Token refreshed successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"token": map[string]interface{}{
											"type":        "string",
											"description": "New JWT token",
										},
									},
								},
							},
						},
					},
					"401": map[string]interface{}{
						"$ref": "#/components/responses/UnauthorizedError",
					},
				},
			},
		},
		"/auth/profile": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Get user profile",
				"description": "Get the current user's profile information",
				"tags":        []string{"Authentication"},
				"operationId": "getProfile",
				"security": []map[string]interface{}{
					{"bearerAuth": []string{}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "User profile",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/UserProfile",
								},
							},
						},
					},
					"401": map[string]interface{}{
						"$ref": "#/components/responses/UnauthorizedError",
					},
				},
			},
		},
		
		// API Keys
		"/auth/keys": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "List API keys",
				"description": "List all API keys for the current user",
				"tags":        []string{"Authentication"},
				"operationId": "listAPIKeys",
				"security": []map[string]interface{}{
					{"bearerAuth": []string{}},
				},
				"parameters": []map[string]interface{}{
					{
						"name":        "active",
						"in":          "query",
						"description": "Filter by active status",
						"schema": map[string]interface{}{
							"type": "boolean",
						},
					},
					{
						"name":        "expired",
						"in":          "query",
						"description": "Show only expired keys",
						"schema": map[string]interface{}{
							"type": "boolean",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of API keys",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"$ref": "#/components/schemas/APIKey",
									},
								},
							},
						},
					},
				},
			},
			"post": map[string]interface{}{
				"summary":     "Create API key",
				"description": "Create a new API key",
				"tags":        []string{"Authentication"},
				"operationId": "createAPIKey",
				"security": []map[string]interface{}{
					{"bearerAuth": []string{}},
				},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/CreateAPIKeyRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "API key created",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/APIKeyResponse",
								},
							},
						},
					},
				},
			},
		},
		
		// Scans
		"/scans": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "List scans",
				"description": "List all security scans with optional filtering",
				"tags":        []string{"Scans"},
				"operationId": "listScans",
				"parameters": []map[string]interface{}{
					{
						"name":        "status",
						"in":          "query",
						"description": "Filter by scan status",
						"schema": map[string]interface{}{
							"type": "string",
							"enum": []string{"pending", "running", "completed", "failed", "cancelled"},
						},
					},
					{
						"name":        "page",
						"in":          "query",
						"description": "Page number",
						"schema": map[string]interface{}{
							"type":    "integer",
							"minimum": 1,
							"default": 1,
						},
					},
					{
						"name":        "per_page",
						"in":          "query",
						"description": "Items per page",
						"schema": map[string]interface{}{
							"type":    "integer",
							"minimum": 1,
							"maximum": 100,
							"default": 20,
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of scans",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ScanListResponse",
								},
							},
						},
					},
				},
			},
			"post": map[string]interface{}{
				"summary":     "Create scan",
				"description": "Create a new security scan",
				"tags":        []string{"Scans"},
				"operationId": "createScan",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/CreateScanRequest",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Scan created",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/Scan",
								},
							},
						},
					},
					"400": map[string]interface{}{
						"$ref": "#/components/responses/BadRequestError",
					},
				},
			},
		},
		"/scans/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Get scan",
				"description": "Get details of a specific scan",
				"tags":        []string{"Scans"},
				"operationId": "getScan",
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"required":    true,
						"description": "Scan ID",
						"schema": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Scan details",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/Scan",
								},
							},
						},
					},
					"404": map[string]interface{}{
						"$ref": "#/components/responses/NotFoundError",
					},
				},
			},
			"delete": map[string]interface{}{
				"summary":     "Cancel scan",
				"description": "Cancel a running scan",
				"tags":        []string{"Scans"},
				"operationId": "cancelScan",
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"required":    true,
						"description": "Scan ID",
						"schema": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Scan cancelled",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"message": map[string]interface{}{
											"type": "string",
										},
									},
								},
							},
						},
					},
					"404": map[string]interface{}{
						"$ref": "#/components/responses/NotFoundError",
					},
				},
			},
		},
		
		// Templates
		"/templates": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "List templates",
				"description": "List all available test templates",
				"tags":        []string{"Templates"},
				"operationId": "listTemplates",
				"parameters": []map[string]interface{}{
					{
						"name":        "category",
						"in":          "query",
						"description": "Filter by category",
						"schema": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of templates",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"$ref": "#/components/schemas/Template",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// generateComponents generates the components section
func generateComponents() map[string]interface{} {
	return map[string]interface{}{
		"securitySchemes": map[string]interface{}{
			"apiKey": map[string]interface{}{
				"type":        "apiKey",
				"in":          "header",
				"name":        "X-API-Key",
				"description": "API key authentication",
			},
			"bearerAuth": map[string]interface{}{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
				"description":  "JWT authentication",
			},
		},
		"schemas": map[string]interface{}{
			// Common schemas
			"Response": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"success": map[string]interface{}{
						"type": "boolean",
					},
					"data": map[string]interface{}{
						"type": "object",
					},
					"error": map[string]interface{}{
						"$ref": "#/components/schemas/Error",
					},
					"meta": map[string]interface{}{
						"$ref": "#/components/schemas/Meta",
					},
				},
			},
			"Error": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Error code",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Error message",
					},
					"details": map[string]interface{}{
						"type":        "string",
						"description": "Additional error details",
					},
				},
				"required": []string{"code", "message"},
			},
			"Meta": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type": "string",
					},
					"request_id": map[string]interface{}{
						"type": "string",
					},
					"pagination": map[string]interface{}{
						"$ref": "#/components/schemas/Pagination",
					},
				},
			},
			"Pagination": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page": map[string]interface{}{
						"type": "integer",
					},
					"per_page": map[string]interface{}{
						"type": "integer",
					},
					"total": map[string]interface{}{
						"type": "integer",
					},
					"total_pages": map[string]interface{}{
						"type": "integer",
					},
				},
			},
			
			// Auth schemas
			"LoginRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"username": map[string]interface{}{
						"type": "string",
					},
					"password": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"username", "password"},
			},
			"LoginResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"token": map[string]interface{}{
						"type":        "string",
						"description": "JWT token",
					},
					"user": map[string]interface{}{
						"$ref": "#/components/schemas/User",
					},
				},
			},
			"User": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"username": map[string]interface{}{
						"type": "string",
					},
					"email": map[string]interface{}{
						"type": "string",
					},
					"role": map[string]interface{}{
						"type": "string",
					},
					"active": map[string]interface{}{
						"type": "boolean",
					},
					"created_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"updated_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			"CreateUserRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"username": map[string]interface{}{
						"type": "string",
					},
					"email": map[string]interface{}{
						"type":   "string",
						"format": "email",
					},
					"password": map[string]interface{}{
						"type":      "string",
						"minLength": 8,
					},
					"role": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"username", "email", "password"},
			},
			
			// API Key schemas
			"APIKey": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"name": map[string]interface{}{
						"type": "string",
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"scopes": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"rate_limit": map[string]interface{}{
						"type":        "integer",
						"description": "Requests per minute",
					},
					"expires_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"last_used_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"created_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			"CreateAPIKeyRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"scopes": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"rate_limit": map[string]interface{}{
						"type": "integer",
					},
					"expires_in": map[string]interface{}{
						"type":        "integer",
						"description": "Days until expiration",
					},
				},
				"required": []string{"name"},
			},
			"APIKeyResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"key": map[string]interface{}{
						"type":        "string",
						"description": "The API key (only shown once)",
					},
					"name": map[string]interface{}{
						"type": "string",
					},
					"created_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			
			// Scan schemas
			"Scan": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"status": map[string]interface{}{
						"type": "string",
						"enum": []string{"pending", "running", "completed", "failed", "cancelled"},
					},
					"target": map[string]interface{}{
						"$ref": "#/components/schemas/ScanTarget",
					},
					"templates": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"config": map[string]interface{}{
						"$ref": "#/components/schemas/ScanConfig",
					},
					"results": map[string]interface{}{
						"$ref": "#/components/schemas/ScanResults",
					},
					"started_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"completed_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"duration": map[string]interface{}{
						"type": "string",
					},
					"created_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			"CreateScanRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target": map[string]interface{}{
						"$ref": "#/components/schemas/ScanTarget",
					},
					"templates": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"categories": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"config": map[string]interface{}{
						"$ref": "#/components/schemas/ScanConfig",
					},
				},
				"required": []string{"target"},
			},
			"ScanTarget": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"endpoint", "model", "dataset"},
					},
					"url": map[string]interface{}{
						"type": "string",
					},
					"model_id": map[string]interface{}{
						"type": "string",
					},
					"dataset_id": map[string]interface{}{
						"type": "string",
					},
					"headers": map[string]interface{}{
						"type": "object",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"type"},
			},
			"ScanConfig": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"concurrent_tests": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 50,
						"default": 10,
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in seconds",
						"minimum":     1,
						"default":     300,
					},
					"retry_failed": map[string]interface{}{
						"type":    "boolean",
						"default": true,
					},
					"max_retries": map[string]interface{}{
						"type":    "integer",
						"minimum": 0,
						"maximum": 10,
						"default": 3,
					},
				},
			},
			"ScanResults": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"summary": map[string]interface{}{
						"$ref": "#/components/schemas/ScanSummary",
					},
					"findings": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/Finding",
						},
					},
					"metrics": map[string]interface{}{
						"type": "object",
						"additionalProperties": map[string]interface{}{
							"type": "number",
						},
					},
				},
			},
			"ScanSummary": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"total_tests": map[string]interface{}{
						"type": "integer",
					},
					"passed": map[string]interface{}{
						"type": "integer",
					},
					"failed": map[string]interface{}{
						"type": "integer",
					},
					"skipped": map[string]interface{}{
						"type": "integer",
					},
					"risk_score": map[string]interface{}{
						"type":    "number",
						"minimum": 0,
						"maximum": 100,
					},
					"severity_counts": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"critical": map[string]interface{}{
								"type": "integer",
							},
							"high": map[string]interface{}{
								"type": "integer",
							},
							"medium": map[string]interface{}{
								"type": "integer",
							},
							"low": map[string]interface{}{
								"type": "integer",
							},
							"info": map[string]interface{}{
								"type": "integer",
							},
						},
					},
				},
			},
			"Finding": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"template_id": map[string]interface{}{
						"type": "string",
					},
					"category": map[string]interface{}{
						"type": "string",
					},
					"severity": map[string]interface{}{
						"type": "string",
						"enum": []string{"critical", "high", "medium", "low", "info"},
					},
					"title": map[string]interface{}{
						"type": "string",
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"evidence": map[string]interface{}{
						"type": "object",
					},
					"remediation": map[string]interface{}{
						"type": "string",
					},
					"references": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			
			// Template schemas
			"Template": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"name": map[string]interface{}{
						"type": "string",
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"category": map[string]interface{}{
						"type": "string",
					},
					"severity": map[string]interface{}{
						"type": "string",
						"enum": []string{"critical", "high", "medium", "low", "info"},
					},
					"tags": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"version": map[string]interface{}{
						"type": "string",
					},
					"author": map[string]interface{}{
						"type": "string",
					},
					"created_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
					"updated_at": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			
			// System schemas
			"HealthResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type": "string",
						"enum": []string{"healthy", "degraded", "unhealthy"},
					},
					"version": map[string]interface{}{
						"type": "string",
					},
					"uptime": map[string]interface{}{
						"type": "string",
					},
					"checks": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"database": map[string]interface{}{
								"type": "boolean",
							},
							"cache": map[string]interface{}{
								"type": "boolean",
							},
							"storage": map[string]interface{}{
								"type": "boolean",
							},
						},
					},
				},
			},
			"VersionInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type": "string",
					},
					"build_date": map[string]interface{}{
						"type": "string",
					},
					"git_commit": map[string]interface{}{
						"type": "string",
					},
					"go_version": map[string]interface{}{
						"type": "string",
					},
					"api_version": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		"responses": map[string]interface{}{
			"BadRequestError": map[string]interface{}{
				"description": "Bad request",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/Response",
						},
					},
				},
			},
			"UnauthorizedError": map[string]interface{}{
				"description": "Unauthorized",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/Response",
						},
					},
				},
			},
			"ForbiddenError": map[string]interface{}{
				"description": "Forbidden",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/Response",
						},
					},
				},
			},
			"NotFoundError": map[string]interface{}{
				"description": "Not found",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/Response",
						},
					},
				},
			},
			"ConflictError": map[string]interface{}{
				"description": "Conflict",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/Response",
						},
					},
				},
			},
			"InternalServerError": map[string]interface{}{
				"description": "Internal server error",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/Response",
						},
					},
				},
			},
		},
	}
}

// handleOpenAPISpec serves the OpenAPI specification
func handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := GenerateOpenAPISpec()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}