# Version Management Database Schema

## Overview

This document defines the database schema for storing version information in the LLMreconing Tool. The schema is designed to track versions of the core binary, templates, and provider modules, as well as maintain version history and compatibility information.

## Schema Design

The version management system will use a JSON-based storage format for simplicity and portability. This allows for easy serialization, deserialization, and human readability while avoiding external database dependencies. The schema will be implemented as a set of structured JSON files stored in a designated directory.

### Directory Structure

```
~/.LLMrecon/
  ├── version/
  │   ├── current.json       # Current installed versions
  │   ├── history.json       # Version update history
  │   └── compatibility.json # Component compatibility rules
  ├── templates/
  │   ├── manifest.json      # Template metadata and versions
  │   └── ...
  └── modules/
      ├── manifest.json      # Module metadata and versions
      └── ...
```

### Schema Definitions

#### 1. Current Version Information (`current.json`)

```json
{
  "core": {
    "version": "1.2.3",
    "build_date": "2025-05-16T12:00:00Z",
    "build_hash": "abc123def456",
    "installed_date": "2025-05-16T14:30:00Z"
  },
  "templates": {
    "version": "1.1.0",
    "last_updated": "2025-05-15T10:00:00Z",
    "count": 42,
    "sources": [
      {
        "name": "github_official",
        "version": "1.1.0",
        "last_updated": "2025-05-15T10:00:00Z"
      },
      {
        "name": "gitlab_dev",
        "version": "0.9.0",
        "last_updated": "2025-05-14T09:00:00Z"
      }
    ]
  },
  "modules": {
    "last_updated": "2025-05-16T14:30:00Z",
    "items": [
      {
        "id": "openai",
        "name": "OpenAI Provider",
        "version": "1.2.0",
        "compatible_with_core": ">=1.0.0 <2.0.0"
      },
      {
        "id": "anthropic",
        "name": "Anthropic Provider",
        "version": "1.1.0",
        "compatible_with_core": ">=1.0.0 <2.0.0"
      }
    ]
  },
  "last_update_check": "2025-05-16T15:00:00Z",
  "updates_available": {
    "core": false,
    "templates": true,
    "modules": {
      "openai": true,
      "anthropic": false
    }
  }
}
```

#### 2. Version History (`history.json`)

```json
{
  "core": [
    {
      "version": "1.2.3",
      "installed_date": "2025-05-16T14:30:00Z",
      "previous_version": "1.2.2",
      "changelog_summary": "Bug fixes and performance improvements",
      "changelog_url": "https://github.com/org/LLMrecon/releases/tag/v1.2.3"
    },
    {
      "version": "1.2.2",
      "installed_date": "2025-05-01T09:15:00Z",
      "previous_version": "1.2.1",
      "changelog_summary": "Fixed template loading issue",
      "changelog_url": "https://github.com/org/LLMrecon/releases/tag/v1.2.2"
    }
  ],
  "templates": [
    {
      "version": "1.1.0",
      "installed_date": "2025-05-15T10:00:00Z",
      "previous_version": "1.0.0",
      "changes": {
        "added": 5,
        "updated": 3,
        "removed": 0
      },
      "sources": [
        {
          "name": "github_official",
          "version": "1.1.0",
          "previous_version": "1.0.0"
        }
      ]
    }
  ],
  "modules": [
    {
      "id": "openai",
      "version": "1.2.0",
      "installed_date": "2025-05-10T11:30:00Z",
      "previous_version": "1.1.0",
      "changelog_summary": "Added support for GPT-4o"
    }
  ]
}
```

#### 3. Compatibility Rules (`compatibility.json`)

```json
{
  "core_compatibility": [
    {
      "version_range": "1.x.x",
      "compatible_template_formats": ["1.0", "1.1"],
      "compatible_module_api": "1.0"
    },
    {
      "version_range": "2.x.x",
      "compatible_template_formats": ["1.1", "2.0"],
      "compatible_module_api": "2.0"
    }
  ],
  "module_requirements": {
    "openai": {
      "min_core_version": "1.0.0",
      "max_core_version": "1.99.99"
    },
    "anthropic": {
      "min_core_version": "1.0.0",
      "max_core_version": "1.99.99"
    }
  },
  "template_compatibility": {
    "prompt_injection": {
      "format_version": "1.0",
      "min_core_version": "1.0.0"
    },
    "data_leakage": {
      "format_version": "1.1",
      "min_core_version": "1.1.0"
    }
  }
}
```

#### 4. Template Manifest (`templates/manifest.json`)

```json
{
  "version": "1.1.0",
  "last_updated": "2025-05-15T10:00:00Z",
  "source": "github_official",
  "templates": [
    {
      "id": "prompt-injection-basic",
      "name": "Basic Prompt Injection Test",
      "version": "1.2.0",
      "category": "prompt_injection",
      "path": "prompt_injection/basic_injection_v1.2.yaml",
      "last_updated": "2025-05-10T08:00:00Z",
      "author": "Security Team",
      "min_core_version": "1.0.0"
    },
    {
      "id": "data-leakage-pii",
      "name": "PII Data Leakage Test",
      "version": "1.1.0",
      "category": "data_leakage",
      "path": "data_leakage/pii_leakage_v1.1.yaml",
      "last_updated": "2025-05-12T14:00:00Z",
      "author": "Privacy Team",
      "min_core_version": "1.0.0"
    }
  ],
  "categories": [
    {
      "id": "prompt_injection",
      "name": "Prompt Injection",
      "description": "Tests for prompt injection vulnerabilities",
      "count": 12
    },
    {
      "id": "data_leakage",
      "name": "Data Leakage",
      "description": "Tests for data leakage vulnerabilities",
      "count": 8
    }
  ]
}
```

#### 5. Module Manifest (`modules/manifest.json`)

```json
{
  "last_updated": "2025-05-16T14:30:00Z",
  "modules": [
    {
      "id": "openai",
      "name": "OpenAI Provider",
      "version": "1.2.0",
      "description": "Integration with OpenAI models (GPT-3.5, GPT-4, GPT-4o)",
      "path": "providers/openai_v1.2.py",
      "author": "Integration Team",
      "compatible_with_core": ">=1.0.0 <2.0.0",
      "last_updated": "2025-05-10T11:30:00Z",
      "supported_models": ["gpt-3.5-turbo", "gpt-4", "gpt-4o"]
    },
    {
      "id": "anthropic",
      "name": "Anthropic Provider",
      "version": "1.1.0",
      "description": "Integration with Anthropic Claude models",
      "path": "providers/anthropic_v1.1.py",
      "author": "Integration Team",
      "compatible_with_core": ">=1.0.0 <2.0.0",
      "last_updated": "2025-05-05T16:45:00Z",
      "supported_models": ["claude-3-opus", "claude-3-sonnet", "claude-3-haiku"]
    }
  ]
}
```

## Schema Relationships

The schema is designed with the following relationships:

1. **Core Version to Templates/Modules**: The core binary version determines compatibility with template formats and module APIs.

2. **Templates to Categories**: Templates are organized by categories, with each template belonging to a specific category.

3. **Modules to Core**: Modules specify compatibility with core versions to ensure they work correctly with the tool.

4. **Version History**: Each component maintains a history of previous versions, allowing for tracking changes over time.

## Version Comparison Logic

The version management system will implement comparison logic based on semantic versioning rules:

1. **Version Parsing**: Parse version strings into major, minor, and patch components.

2. **Comparison Algorithm**: Compare versions by first checking major, then minor, then patch numbers.

3. **Range Checking**: Support version range specifications (e.g., `>=1.0.0 <2.0.0`) for compatibility checks.

4. **Precedence Rules**: Follow SemVer precedence rules for pre-release versions and build metadata.

## Data Access Patterns

The schema is optimized for the following common operations:

1. **Current Version Lookup**: Quick access to currently installed versions.

2. **Update Availability Check**: Efficient comparison of local and remote versions.

3. **Compatibility Verification**: Fast checking of component compatibility.

4. **Version History Tracking**: Chronological tracking of version changes.

5. **Template/Module Enumeration**: Listing available templates and modules with version information.

## Implementation Considerations

1. **File Locking**: Implement proper file locking to prevent corruption during updates.

2. **Backup Strategy**: Create backups of version files before modifications.

3. **Schema Migration**: Support schema evolution as the tool develops.

4. **Error Handling**: Implement robust error handling for file access and parsing issues.

5. **Performance Optimization**: Keep files small and focused for quick loading and saving.

## Future Extensions

The schema is designed to be extensible for future needs:

1. **Additional Metadata**: Support for more detailed component metadata.

2. **Multiple Template Sources**: Enhanced tracking of templates from various sources.

3. **User Customizations**: Tracking user-modified templates separately from official ones.

4. **Dependency Graphs**: More detailed tracking of dependencies between components.

5. **Distributed Versioning**: Support for distributed version management in enterprise environments.
