# Manifest Schema Definition

## Overview

This document provides a comprehensive definition of the manifest schema for the LLMrecon Tool's bundle format. The manifest is a critical component of every bundle, containing metadata, content inventory, and validation information that ensures the integrity and usability of the bundle.

## Schema Version

The current schema version is `1.0.0`. This version number will be incremented for any backward-incompatible changes to the schema structure.

## Schema Location

The JSON Schema definition is located at:
- `/schemas/bundle-manifest-schema.json`

This schema can be used to validate manifest files programmatically.

## Required Fields

The following fields are required in every bundle manifest:

| Field | Type | Description |
|-------|------|-------------|
| `schema_version` | String | Version of the manifest schema (e.g., "1.0.0") |
| `bundle_id` | String | Unique identifier for the bundle |
| `bundle_type` | String | Type of bundle ("templates", "modules", or "mixed") |
| `name` | String | Human-readable name of the bundle |
| `description` | String | Detailed description of the bundle |
| `version` | String | Version of the bundle (SemVer format) |
| `created_at` | String | Timestamp when the bundle was created (ISO 8601 format) |
| `author` | Object | Information about the bundle author |
| `content` | Array | List of content items in the bundle |
| `checksums` | Object | Checksums for bundle components |
| `compatibility` | Object | Compatibility information for the bundle |

## Field Definitions

### `schema_version`

- **Type**: String
- **Pattern**: `^\d+\.\d+\.\d+$` (SemVer format)
- **Description**: Version of the manifest schema
- **Example**: `"1.0.0"`

### `bundle_id`

- **Type**: String
- **Pattern**: `^[a-zA-Z0-9_-]+$` (alphanumeric, underscore, hyphen)
- **Length**: 3-64 characters
- **Description**: Unique identifier for the bundle
- **Example**: `"owasp-llm-top10-templates"`

### `bundle_type`

- **Type**: String
- **Enum**: `["templates", "modules", "mixed"]`
- **Description**: Type of bundle
- **Example**: `"templates"`

### `name`

- **Type**: String
- **Length**: 1-100 characters
- **Description**: Human-readable name of the bundle
- **Example**: `"OWASP LLM Top 10 Templates"`

### `description`

- **Type**: String
- **Length**: 1-1000 characters
- **Description**: Detailed description of the bundle
- **Example**: `"A collection of templates for testing against the OWASP LLM Top 10 vulnerabilities"`

### `version`

- **Type**: String
- **Pattern**: `^\d+\.\d+\.\d+(-[a-zA-Z0-9\.]+)?$` (SemVer format with optional pre-release)
- **Description**: Version of the bundle
- **Example**: `"1.0.0"` or `"1.0.0-beta.1"`

### `created_at`

- **Type**: String (date-time)
- **Format**: ISO 8601 (e.g., "2025-05-25T17:57:09-04:00")
- **Description**: Timestamp when the bundle was created
- **Example**: `"2025-05-25T17:57:09-04:00"`

### `author`

- **Type**: Object
- **Required Fields**: `name`, `email`
- **Description**: Information about the bundle author
- **Example**:
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com",
    "organization": "Example Org",
    "url": "https://example.com",
    "key_id": "1a2b3c4d"
  }
  ```

#### Author Properties

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | String | Yes | Name of the author |
| `email` | String | Yes | Email of the author |
| `url` | String | No | URL of the author's website |
| `organization` | String | No | Organization the author belongs to |
| `key_id` | String | No | Cryptographic key identifier for the author |

### `content`

- **Type**: Array
- **Min Items**: 1
- **Description**: List of content items in the bundle
- **Example**:
  ```json
  [
    {
      "path": "templates/owasp-llm/llm01-prompt-injection/basic-injection.json",
      "type": "template",
      "id": "template-001",
      "version": "1.0.0",
      "description": "Basic prompt injection template",
      "checksum": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "categories": ["prompt-injection", "owasp-llm"]
    }
  ]
  ```

#### Content Item Properties

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `path` | String | Yes | Relative path to the content item within the bundle |
| `type` | String | Yes | Type of content item (enum: "template", "module", "documentation", "resource", "configuration") |
| `id` | String | No | Unique identifier for the content item |
| `version` | String | No | Version of the content item (SemVer format) |
| `description` | String | No | Description of the content item |
| `checksum` | String | Yes | Checksum of the content item (format: "sha256:[64 hex chars]") |
| `bundle_id` | String | No | ID of the bundle this item belongs to |
| `categories` | Array | No | Categories the content item belongs to |
| `tags` | Array | No | Tags associated with the content item |

### `checksums`

- **Type**: Object
- **Required Fields**: `manifest`, `content`
- **Description**: Checksums for bundle components
- **Example**:
  ```json
  {
    "manifest": "sha256:0987654321fedcba0987654321fedcba0987654321fedcba0987654321fedcba",
    "content": {
      "templates/example.json": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    }
  }
  ```

#### Checksums Properties

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `manifest` | String | Yes | Checksum of the manifest file without the checksums and signature fields |
| `content` | Object | Yes | Map of content paths to checksums |

### `compatibility`

- **Type**: Object
- **Required Fields**: `min_version`
- **Description**: Compatibility information for the bundle
- **Example**:
  ```json
  {
    "min_version": "1.0.0",
    "max_version": "2.0.0",
    "dependencies": ["base-templates"],
    "incompatible": ["legacy-templates"]
  }
  ```

#### Compatibility Properties

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `min_version` | String | Yes | Minimum version of the LLMrecon Tool required |
| `max_version` | String | No | Maximum version of the LLMrecon Tool supported |
| `dependencies` | Array | No | List of bundle IDs this bundle depends on |
| `incompatible` | Array | No | List of bundle IDs this bundle is incompatible with |

## Optional Fields

The following fields are optional in the bundle manifest:

### `signature`

- **Type**: String
- **Pattern**: Base64 pattern
- **Description**: Base64-encoded cryptographic signature of the manifest
- **Example**: `"base64-encoded-signature"`

### `is_incremental`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Whether this is an incremental bundle that updates a previous version
- **Example**: `false`

### `base_version`

- **Type**: String
- **Pattern**: SemVer format
- **Description**: Version of the base bundle this incremental bundle updates
- **Example**: `"1.0.0"`

### `compliance`

- **Type**: Object
- **Description**: Compliance mappings for content items
- **Example**:
  ```json
  {
    "owaspLLMTop10": {
      "LLM01:PromptInjection": ["template-001"],
      "LLM06:SensitiveInformationDisclosure": ["template-001"]
    },
    "ISOIEC42001": {
      "42001:8.2.3": ["template-001"],
      "42001:8.2.4": ["template-001"]
    }
  }
  ```

### `changelog`

- **Type**: Array
- **Description**: List of changes in this bundle version
- **Example**:
  ```json
  [
    {
      "version": "1.0.0",
      "date": "2025-05-25T17:57:09-04:00",
      "changes": ["Initial release"]
    }
  ]
  ```

#### Changelog Entry Properties

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | String | Yes | Version associated with these changes |
| `date` | String | Yes | Date when these changes were made |
| `changes` | Array | Yes | List of changes |

### `documentation`

- **Type**: Object
- **Description**: Paths to documentation files
- **Example**:
  ```json
  {
    "README": "README.md",
    "usage": "documentation/usage.md",
    "installation": "documentation/installation.md"
  }
  ```

## Validation Rules

The manifest schema enforces the following validation rules:

1. **Required Fields**: All required fields must be present and non-empty.
2. **Field Types**: Each field must have the correct data type.
3. **String Patterns**: String fields with pattern constraints must match their patterns.
4. **Enumerations**: Fields with enumerated values must contain one of the allowed values.
5. **Length Constraints**: String fields must respect their minimum and maximum length constraints.
6. **Array Constraints**: Arrays must respect their minimum item count constraints.
7. **Object Constraints**: Objects must contain all their required properties.
8. **Format Constraints**: Fields with format constraints (e.g., email, URI, date-time) must be valid.

## Complete Example

```json
{
  "schema_version": "1.0.0",
  "bundle_id": "owasp-llm-top10-templates",
  "bundle_type": "templates",
  "name": "OWASP LLM Top 10 Templates",
  "description": "A collection of templates for testing against the OWASP LLM Top 10 vulnerabilities",
  "version": "1.0.0",
  "created_at": "2025-05-25T17:57:09-04:00",
  "author": {
    "name": "John Doe",
    "email": "john@example.com",
    "organization": "Example Org",
    "url": "https://example.com",
    "key_id": "1a2b3c4d"
  },
  "content": [
    {
      "path": "templates/owasp-llm/llm01-prompt-injection/basic-injection.json",
      "type": "template",
      "id": "template-001",
      "version": "1.0.0",
      "description": "Basic prompt injection template",
      "checksum": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "categories": ["prompt-injection", "owasp-llm"],
      "tags": ["basic", "injection"]
    }
  ],
  "checksums": {
    "manifest": "sha256:0987654321fedcba0987654321fedcba0987654321fedcba0987654321fedcba",
    "content": {
      "templates/owasp-llm/llm01-prompt-injection/basic-injection.json": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    }
  },
  "compatibility": {
    "min_version": "1.0.0",
    "max_version": "2.0.0",
    "dependencies": ["base-templates"],
    "incompatible": ["legacy-templates"]
  },
  "signature": "base64-encoded-signature",
  "is_incremental": false,
  "compliance": {
    "owaspLLMTop10": {
      "LLM01:PromptInjection": ["template-001"],
      "LLM06:SensitiveInformationDisclosure": ["template-001"]
    },
    "ISOIEC42001": {
      "42001:8.2.3": ["template-001"],
      "42001:8.2.4": ["template-001"]
    }
  },
  "changelog": [
    {
      "version": "1.0.0",
      "date": "2025-05-25T17:57:09-04:00",
      "changes": ["Initial release"]
    }
  ],
  "documentation": {
    "README": "README.md",
    "usage": "documentation/usage.md",
    "installation": "documentation/installation.md"
  }
}
```

## Programmatic Validation

To validate a manifest against this schema, you can use any JSON Schema validator that supports draft-07. For example, in Go, you can use the `github.com/xeipuuv/gojsonschema` package:

```go
package main

import (
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

func ValidateManifest(manifestPath, schemaPath string) (bool, []string, error) {
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewReferenceLoader("file://" + manifestPath)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false, nil, err
	}

	if result.Valid() {
		return true, nil, nil
	}

	var errors []string
	for _, desc := range result.Errors() {
		errors = append(errors, desc.String())
	}

	return false, errors, nil
}
```
