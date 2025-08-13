# Scan Management API

This document describes the API endpoints for managing red-team scans in the LLMrecon tool.

## Base URL

All API endpoints are relative to the base URL: `/api/v1`

## Authentication

Authentication is required for all endpoints. The API uses token-based authentication via the `Authorization` header:

```
Authorization: Bearer <token>
```

## Error Handling

Errors are returned as JSON objects with the following structure:

```json
{
  "error": "Error message",
  "code": 400
}
```

The `code` field corresponds to the HTTP status code.

## Pagination

List endpoints support pagination using the following query parameters:

- `page`: Page number (1-based, default: 1)
- `page_size`: Number of items per page (default: 10)

Paginated responses have the following structure:

```json
{
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_items": 42,
    "total_pages": 5
  },
  "data": [
    // Array of items
  ]
}
```

## Filtering

List endpoints support filtering using the following query parameters:

- `status`: Filter by status (e.g., "pending", "running", "completed", "failed", "cancelled")
- `severity`: Filter by severity (e.g., "low", "medium", "high", "critical")
- `start_date`: Filter by start date (ISO 8601 format)
- `end_date`: Filter by end date (ISO 8601 format)
- `search`: Search term to filter by name, description, etc.

## Scan Configuration Endpoints

### List Scan Configurations

```
GET /scan-configs
```

Returns a paginated list of scan configurations.

#### Query Parameters

- Pagination parameters
- Filter parameters

#### Response

```json
{
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_items": 2,
    "total_pages": 1
  },
  "data": [
    {
      "id": "config-1",
      "name": "Example Scan Config",
      "description": "An example scan configuration",
      "target": "example-target",
      "target_type": "prompt",
      "templates": ["template-1", "template-2"],
      "parameters": {
        "param1": "value1",
        "param2": 42
      },
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-02T12:00:00Z",
      "created_by": "user-1"
    },
    // ...
  ]
}
```

### Create Scan Configuration

```
POST /scan-configs
```

Creates a new scan configuration.

#### Request Body

```json
{
  "name": "Example Scan Config",
  "description": "An example scan configuration",
  "target": "example-target",
  "target_type": "prompt",
  "templates": ["template-1", "template-2"],
  "parameters": {
    "param1": "value1",
    "param2": 42
  }
}
```

#### Response

```json
{
  "id": "config-1",
  "name": "Example Scan Config",
  "description": "An example scan configuration",
  "target": "example-target",
  "target_type": "prompt",
  "templates": ["template-1", "template-2"],
  "parameters": {
    "param1": "value1",
    "param2": 42
  },
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z",
  "created_by": "user-1"
}
```

### Get Scan Configuration

```
GET /scan-configs/{id}
```

Returns a specific scan configuration.

#### Response

```json
{
  "id": "config-1",
  "name": "Example Scan Config",
  "description": "An example scan configuration",
  "target": "example-target",
  "target_type": "prompt",
  "templates": ["template-1", "template-2"],
  "parameters": {
    "param1": "value1",
    "param2": 42
  },
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z",
  "created_by": "user-1"
}
```

### Update Scan Configuration

```
PUT /scan-configs/{id}
```

Updates a specific scan configuration.

#### Request Body

```json
{
  "name": "Updated Scan Config",
  "description": "An updated scan configuration",
  "target": "updated-target",
  "target_type": "system",
  "templates": ["template-1", "template-3"],
  "parameters": {
    "param1": "new-value",
    "param3": true
  }
}
```

All fields are optional. Only the fields that are provided will be updated.

#### Response

```json
{
  "id": "config-1",
  "name": "Updated Scan Config",
  "description": "An updated scan configuration",
  "target": "updated-target",
  "target_type": "system",
  "templates": ["template-1", "template-3"],
  "parameters": {
    "param1": "new-value",
    "param3": true
  },
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-02T12:00:00Z",
  "created_by": "user-1"
}
```

### Delete Scan Configuration

```
DELETE /scan-configs/{id}
```

Deletes a specific scan configuration.

#### Response

No content (204)

## Scan Execution Endpoints

### List Scans

```
GET /scans
```

Returns a paginated list of scans.

#### Query Parameters

- Pagination parameters
- Filter parameters

#### Response

```json
{
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_items": 2,
    "total_pages": 1
  },
  "data": [
    {
      "id": "scan-1",
      "config_id": "config-1",
      "status": "completed",
      "start_time": "2023-01-01T12:00:00Z",
      "end_time": "2023-01-01T12:05:00Z",
      "progress": 100,
      "results": [
        // Array of scan results (may be omitted for brevity)
      ]
    },
    // ...
  ]
}
```

### Create Scan

```
POST /scans
```

Creates a new scan execution.

#### Request Body

```json
{
  "config_id": "config-1"
}
```

#### Response

```json
{
  "id": "scan-1",
  "config_id": "config-1",
  "status": "pending",
  "start_time": "2023-01-01T12:00:00Z",
  "progress": 0
}
```

### Get Scan

```
GET /scans/{id}
```

Returns a specific scan.

#### Response

```json
{
  "id": "scan-1",
  "config_id": "config-1",
  "status": "running",
  "start_time": "2023-01-01T12:00:00Z",
  "progress": 50,
  "results": [
    // Array of scan results (may be omitted for brevity)
  ]
}
```

### Cancel Scan

```
POST /scans/{id}/cancel
```

Cancels a running scan.

#### Response

```json
{
  "id": "scan-1",
  "config_id": "config-1",
  "status": "cancelled",
  "start_time": "2023-01-01T12:00:00Z",
  "end_time": "2023-01-01T12:02:30Z",
  "progress": 50,
  "results": [
    // Array of scan results (may be omitted for brevity)
  ]
}
```

### Get Scan Results

```
GET /scans/{id}/results
```

Returns a paginated list of results for a specific scan.

#### Query Parameters

- Pagination parameters
- Filter parameters

#### Response

```json
{
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_items": 3,
    "total_pages": 1
  },
  "data": [
    {
      "id": "result-1",
      "scan_id": "scan-1",
      "template_id": "template-1",
      "severity": "medium",
      "title": "Example Finding",
      "description": "An example finding from the scan",
      "details": {
        "key1": "value1",
        "key2": 42
      },
      "timestamp": "2023-01-01T12:01:00Z"
    },
    // ...
  ]
}
```

## Status Codes

- `200 OK`: The request was successful
- `201 Created`: The resource was created successfully
- `204 No Content`: The request was successful but there is no content to return
- `400 Bad Request`: The request was invalid
- `401 Unauthorized`: Authentication is required
- `403 Forbidden`: The authenticated user does not have permission to access the resource
- `404 Not Found`: The requested resource was not found
- `500 Internal Server Error`: An error occurred on the server
