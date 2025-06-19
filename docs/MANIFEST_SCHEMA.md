# Manifest Schema Design

## Overview

This document defines the schema for manifest files in the LLMreconing Tool. Manifests are used to track and manage templates and modules, providing metadata, versioning information, and compatibility details.

## Template Manifest Schema

### Global Template Manifest (`templates/manifest.json`)

The global template manifest provides a comprehensive listing of all available templates and categories.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["schema_version", "last_updated", "templates", "categories"],
  "properties": {
    "schema_version": {
      "type": "string",
      "description": "Version of the manifest schema",
      "pattern": "^\\d+\\.\\d+$"
    },
    "last_updated": {
      "type": "string",
      "description": "ISO 8601 timestamp of when the manifest was last updated",
      "format": "date-time"
    },
    "templates": {
      "type": "array",
      "description": "List of all available templates",
      "items": {
        "type": "object",
        "required": ["id", "name", "path", "version", "category", "severity", "compatible_providers", "min_tool_version", "author", "description"],
        "properties": {
          "id": {
            "type": "string",
            "description": "Unique identifier for the template",
            "pattern": "^[a-z0-9_]+(_v\\d+\\.\\d+)?$"
          },
          "name": {
            "type": "string",
            "description": "Human-readable name of the template"
          },
          "path": {
            "type": "string",
            "description": "Relative path to the template file from the templates directory"
          },
          "version": {
            "type": "string",
            "description": "Version of the template in major.minor format",
            "pattern": "^\\d+\\.\\d+$"
          },
          "category": {
            "type": "string",
            "description": "Category the template belongs to (e.g., prompt_injection)"
          },
          "severity": {
            "type": "string",
            "description": "Severity level of the vulnerability being tested",
            "enum": ["info", "low", "medium", "high", "critical"]
          },
          "compatible_providers": {
            "type": "array",
            "description": "List of LLM providers this template is compatible with",
            "items": {
              "type": "string"
            }
          },
          "min_tool_version": {
            "type": "string",
            "description": "Minimum tool version required to use this template",
            "pattern": "^\\d+\\.\\d+\\.\\d+$"
          },
          "author": {
            "type": "string",
            "description": "Author of the template"
          },
          "description": {
            "type": "string",
            "description": "Brief description of the template"
          },
          "tags": {
            "type": "array",
            "description": "List of tags for categorization and filtering",
            "items": {
              "type": "string"
            }
          }
        }
      }
    },
    "categories": {
      "type": "array",
      "description": "List of all template categories",
      "items": {
        "type": "object",
        "required": ["id", "name", "description"],
        "properties": {
          "id": {
            "type": "string",
            "description": "Unique identifier for the category",
            "pattern": "^[a-z0-9_]+$"
          },
          "name": {
            "type": "string",
            "description": "Human-readable name of the category"
          },
          "description": {
            "type": "string",
            "description": "Description of the category"
          },
          "owasp_reference": {
            "type": "string",
            "description": "Reference to the corresponding OWASP LLM Top 10 category"
          },
          "iso_reference": {
            "type": "string",
            "description": "Reference to the corresponding ISO standard"
          }
        }
      }
    }
  }
}
```

### Category Info (`templates/[category]/category_info.json`)

Each category directory contains a category info file that provides metadata about the category and lists the templates within it.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["id", "name", "description", "templates"],
  "properties": {
    "id": {
      "type": "string",
      "description": "Unique identifier for the category",
      "pattern": "^[a-z0-9_]+$"
    },
    "name": {
      "type": "string",
      "description": "Human-readable name of the category"
    },
    "description": {
      "type": "string",
      "description": "Description of the category"
    },
    "owasp_reference": {
      "type": "string",
      "description": "Reference to the corresponding OWASP LLM Top 10 category"
    },
    "iso_reference": {
      "type": "string",
      "description": "Reference to the corresponding ISO standard"
    },
    "templates": {
      "type": "array",
      "description": "List of template filenames in this category",
      "items": {
        "type": "string",
        "pattern": "^[a-z0-9_]+_v\\d+\\.\\d+\\.yaml$"
      }
    }
  }
}
```

## Module Manifest Schema

### Global Module Manifest (`modules/manifest.json`)

The global module manifest provides a comprehensive listing of all available modules, including providers, utilities, and detectors.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["schema_version", "last_updated"],
  "properties": {
    "schema_version": {
      "type": "string",
      "description": "Version of the manifest schema",
      "pattern": "^\\d+\\.\\d+$"
    },
    "last_updated": {
      "type": "string",
      "description": "ISO 8601 timestamp of when the manifest was last updated",
      "format": "date-time"
    },
    "providers": {
      "type": "array",
      "description": "List of all available LLM provider modules",
      "items": {
        "type": "object",
        "required": ["id", "name", "path", "version", "min_tool_version", "author", "description"],
        "properties": {
          "id": {
            "type": "string",
            "description": "Unique identifier for the provider",
            "pattern": "^[a-z0-9_]+$"
          },
          "name": {
            "type": "string",
            "description": "Human-readable name of the provider"
          },
          "path": {
            "type": "string",
            "description": "Relative path to the provider module from the modules directory"
          },
          "version": {
            "type": "string",
            "description": "Version of the provider module in major.minor format",
            "pattern": "^\\d+\\.\\d+$"
          },
          "supported_models": {
            "type": "array",
            "description": "List of model identifiers supported by this provider",
            "items": {
              "type": "string"
            }
          },
          "min_tool_version": {
            "type": "string",
            "description": "Minimum tool version required to use this provider",
            "pattern": "^\\d+\\.\\d+\\.\\d+$"
          },
          "author": {
            "type": "string",
            "description": "Author of the provider module"
          },
          "description": {
            "type": "string",
            "description": "Brief description of the provider module"
          }
        }
      }
    },
    "utilities": {
      "type": "array",
      "description": "List of all available utility modules",
      "items": {
        "type": "object",
        "required": ["id", "name", "path", "version", "min_tool_version", "author", "description"],
        "properties": {
          "id": {
            "type": "string",
            "description": "Unique identifier for the utility",
            "pattern": "^[a-z0-9_]+$"
          },
          "name": {
            "type": "string",
            "description": "Human-readable name of the utility"
          },
          "path": {
            "type": "string",
            "description": "Relative path to the utility module from the modules directory"
          },
          "version": {
            "type": "string",
            "description": "Version of the utility module in major.minor format",
            "pattern": "^\\d+\\.\\d+$"
          },
          "min_tool_version": {
            "type": "string",
            "description": "Minimum tool version required to use this utility",
            "pattern": "^\\d+\\.\\d+\\.\\d+$"
          },
          "author": {
            "type": "string",
            "description": "Author of the utility module"
          },
          "description": {
            "type": "string",
            "description": "Brief description of the utility module"
          }
        }
      }
    },
    "detectors": {
      "type": "array",
      "description": "List of all available detector modules",
      "items": {
        "type": "object",
        "required": ["id", "name", "path", "version", "min_tool_version", "author", "description"],
        "properties": {
          "id": {
            "type": "string",
            "description": "Unique identifier for the detector",
            "pattern": "^[a-z0-9_]+$"
          },
          "name": {
            "type": "string",
            "description": "Human-readable name of the detector"
          },
          "path": {
            "type": "string",
            "description": "Relative path to the detector module from the modules directory"
          },
          "version": {
            "type": "string",
            "description": "Version of the detector module in major.minor format",
            "pattern": "^\\d+\\.\\d+$"
          },
          "min_tool_version": {
            "type": "string",
            "description": "Minimum tool version required to use this detector",
            "pattern": "^\\d+\\.\\d+\\.\\d+$"
          },
          "author": {
            "type": "string",
            "description": "Author of the detector module"
          },
          "description": {
            "type": "string",
            "description": "Brief description of the detector module"
          }
        }
      }
    }
  }
}
```

## Template File Schema (YAML)

```yaml
# JSON Schema representation in YAML format
$schema: http://json-schema.org/draft-07/schema#
type: object
required:
  - id
  - info
  - compatibility
  - test
properties:
  id:
    type: string
    description: Unique identifier for the template
    pattern: ^[a-z0-9_]+(_v\d+\.\d+)?$
  
  info:
    type: object
    required:
      - name
      - description
      - version
      - author
      - severity
    properties:
      name:
        type: string
        description: Human-readable name of the template
      description:
        type: string
        description: Detailed description of the template
      version:
        type: string
        description: Version of the template in major.minor format
        pattern: ^\d+\.\d+$
      author:
        type: string
        description: Author of the template
      severity:
        type: string
        description: Severity level of the vulnerability being tested
        enum:
          - info
          - low
          - medium
          - high
          - critical
      tags:
        type: array
        description: List of tags for categorization and filtering
        items:
          type: string
      references:
        type: array
        description: List of reference URLs
        items:
          type: string
      compliance:
        type: object
        properties:
          owasp:
            type: string
            description: Reference to the corresponding OWASP LLM Top 10 category
          iso:
            type: string
            description: Reference to the corresponding ISO standard
  
  compatibility:
    type: object
    required:
      - providers
    properties:
      providers:
        type: array
        description: List of LLM providers this template is compatible with
        items:
          type: string
      min_tool_version:
        type: string
        description: Minimum tool version required to use this template
        pattern: ^\d+\.\d+\.\d+$
  
  test:
    type: object
    required:
      - prompt
      - detection
    properties:
      prompt:
        type: string
        description: The prompt to send to the LLM
      expected_behavior:
        type: string
        description: Description of the expected behavior from a secure LLM
      detection:
        type: object
        required:
          - type
        properties:
          type:
            type: string
            description: Type of detection method to use
            enum:
              - string_match
              - regex_match
              - semantic_match
          match:
            type: string
            description: String to match (for string_match type)
          pattern:
            type: string
            description: Regex pattern to match (for regex_match type)
          criteria:
            type: string
            description: Semantic criteria to evaluate (for semantic_match type)
          condition:
            type: string
            description: Condition for the match (contains or not_contains)
            enum:
              - contains
              - not_contains
      variations:
        type: array
        description: Additional test variations
        items:
          type: object
          required:
            - prompt
            - detection
          properties:
            prompt:
              type: string
              description: The prompt to send to the LLM
            detection:
              type: object
              required:
                - type
              properties:
                type:
                  type: string
                  description: Type of detection method to use
                  enum:
                    - string_match
                    - regex_match
                    - semantic_match
                match:
                  type: string
                  description: String to match (for string_match type)
                pattern:
                  type: string
                  description: Regex pattern to match (for regex_match type)
                criteria:
                  type: string
                  description: Semantic criteria to evaluate (for semantic_match type)
                condition:
                  type: string
                  description: Condition for the match (contains or not_contains)
                  enum:
                    - contains
                    - not_contains
```

## Validation Rules

### General Validation Rules

1. **Version Format**:
   - All version numbers must follow the specified format
   - Schema version: `major.minor` (e.g., `1.0`)
   - Template/module version: `major.minor` (e.g., `1.2`)
   - Tool version: `major.minor.patch` (e.g., `1.0.0`)

2. **ID Format**:
   - All IDs must be lowercase alphanumeric with underscores
   - Template IDs should include the category and optionally the version
   - Category IDs should be descriptive and match directory names

3. **Path Format**:
   - All paths must be relative to the root directory (templates/ or modules/)
   - Paths must use forward slashes (/) as separators
   - Paths must include the file extension

4. **Timestamps**:
   - All timestamps must be in ISO 8601 format (e.g., `2025-05-18T20:00:00Z`)

### Template-Specific Validation Rules

1. **Template Files**:
   - Must be valid YAML or JSON
   - Must include all required fields
   - Detection methods must match one of the supported types

2. **Category Organization**:
   - Each template must belong to exactly one category
   - Categories must have a corresponding directory
   - Category info must list all templates in the directory

### Module-Specific Validation Rules

1. **Provider Modules**:
   - Must implement the Provider interface
   - Must specify supported models
   - Must handle authentication and API communication

2. **Utility Modules**:
   - Must provide a clear, single-purpose utility function
   - Must be reusable across different providers

3. **Detector Modules**:
   - Must implement the Detector interface
   - Must provide clear detection logic and reporting

## Manifest Update Process

When adding or updating templates or modules, the following process should be followed:

1. **Adding a New Template**:
   - Create the template file in the appropriate category directory
   - Update the category_info.json file to include the new template
   - Update the global manifest.json to include the new template metadata

2. **Updating an Existing Template**:
   - Increment the version number in the template file
   - Update the template file with the new content
   - Update the category_info.json and global manifest.json with the new version

3. **Adding a New Module**:
   - Create the module file in the appropriate directory
   - Update the global manifest.json to include the new module metadata

4. **Updating an Existing Module**:
   - Increment the version number in the module file
   - Update the module file with the new content
   - Update the global manifest.json with the new version

## Conclusion

This manifest schema design provides a structured approach to managing templates and modules in the LLMreconing Tool. It ensures consistency, enables versioning, and facilitates compatibility checks between different components of the system.
