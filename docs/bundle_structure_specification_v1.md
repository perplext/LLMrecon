# LLMrecon Tool - Bundle Structure Specification

## Overview

This document provides the official specification for the structure of bundles in the LLMrecon Tool. It defines the directory layout, file organization, naming conventions, and resource placement within bundles to ensure consistency and interoperability.

## Bundle Format

Bundles are structured as directories with a standardized layout. They can be distributed in two formats:

1. **Directory Format**: A directory containing all bundle files and subdirectories
2. **Archive Format**: A ZIP archive containing the same structure for easier transfer

## Root Directory Structure

The root directory of a bundle contains the following elements:

| Path | Type | Required | Description |
|------|------|----------|-------------|
| `manifest.json` | File | Yes | Bundle manifest containing metadata, content inventory, and integrity information |
| `README.md` | File | Yes | Human-readable overview of the bundle and its contents |
| `templates/` | Directory | Yes | Contains all template files organized by category |
| `modules/` | Directory | No | Contains modules that extend the functionality of templates |
| `binary/` | Directory | No | Contains binary executables for the LLMreconing Tool |
| `documentation/` | Directory | Yes | Contains comprehensive documentation for the bundle |
| `signatures/` | Directory | Yes | Contains cryptographic signatures for bundle verification |
| `resources/` | Directory | No | Contains additional resources used by templates or modules |
| `config/` | Directory | No | Contains configuration files for templates and modules |
| `repository-config/` | Directory | No | Contains configuration for template repositories |

## Directory Structure Details

### `templates/` Directory

The `templates/` directory contains all template files organized by category. The structure follows a hierarchical organization based on vulnerability categories and compliance frameworks:

```
templates/
├── owasp-llm/                  # OWASP LLM Top 10 categories
│   ├── llm01-prompt-injection/
│   │   ├── basic-injection.json
│   │   ├── advanced-injection.json
│   │   └── ...
│   ├── llm02-insecure-output/
│   │   ├── ...
│   └── ...
├── iso-42001/                  # ISO/IEC 42001 controls
│   ├── control-8.2.3/
│   │   ├── ...
│   └── ...
├── custom/                     # Custom templates
│   ├── ...
└── ...
```

Template files must follow these naming conventions:
- Use lowercase with hyphens for readability
- Include a descriptive name that indicates the purpose
- Use `.json` extension for all template files
- Group related templates in subdirectories

### `modules/` Directory

The `modules/` directory contains modules that extend the functionality of templates:

```
modules/
├── providers/                  # LLM provider modules
│   ├── openai.json
│   ├── anthropic.json
│   └── ...
├── detectors/                  # Vulnerability detection modules
│   ├── prompt-injection.json
│   ├── insecure-output.json
│   └── ...
├── utilities/                  # Utility modules
│   ├── ...
└── ...
```

Module files must follow these naming conventions:
- Use lowercase with hyphens for readability
- Include a descriptive name that indicates the purpose
- Use `.json` extension for module definition files
- Group related modules in subdirectories

### `binary/` Directory

The `binary/` directory contains binary executables for the LLMreconing Tool:

```
binary/
├── tool-v1.2.3-linux-amd64     # Linux binary
├── tool-v1.2.3-darwin-amd64    # macOS binary
├── tool-v1.2.3-windows-amd64.exe # Windows binary
└── ...
```

Binary files must follow these naming conventions:
- Include the tool name, version, operating system, and architecture
- Use lowercase with hyphens for readability
- Include appropriate file extensions for executables

### `documentation/` Directory

The `documentation/` directory contains comprehensive documentation for the bundle:

```
documentation/
├── usage.md                    # Usage guide
├── installation.md             # Installation instructions
├── api-endpoints.md            # API documentation
├── cli-commands.md             # CLI documentation
├── template-format.md          # Template format specification
├── module-format.md            # Module format specification
├── compliance/                 # Compliance documentation
│   ├── owasp-llm-top10.md      # OWASP LLM Top 10 documentation
│   ├── iso-42001.md            # ISO/IEC 42001 documentation
│   └── ...
└── ...
```

Documentation files must follow these naming conventions:
- Use lowercase with hyphens for readability
- Use `.md` (Markdown) format for all documentation
- Group related documentation in subdirectories
- Include clear titles and section headings

### `signatures/` Directory

The `signatures/` directory contains cryptographic signatures for bundle verification:

```
signatures/
├── manifest.sig                # Signature for the manifest file
├── content/                    # Signatures for content files
│   ├── templates/
│   │   ├── ...
│   ├── modules/
│   │   ├── ...
│   └── ...
└── public-key.pem              # Public key for verification
```

Signature files must follow these naming conventions:
- Use the same path as the signed file with `.sig` extension
- Store public keys with `.pem` extension
- Organize signatures to mirror the structure of the signed content

### `resources/` Directory

The `resources/` directory contains additional resources used by templates or modules:

```
resources/
├── images/                     # Image resources
│   ├── ...
├── data/                       # Data files
│   ├── ...
├── schemas/                    # JSON schemas
│   ├── template-schema.json    # Schema for template validation
│   ├── module-schema.json      # Schema for module validation
│   └── ...
└── ...
```

Resource files must follow these naming conventions:
- Use lowercase with hyphens for readability
- Group resources by type in subdirectories
- Use appropriate file extensions for each resource type

### `config/` Directory

The `config/` directory contains configuration files for templates and modules:

```
config/
├── defaults.json               # Default configuration
├── environments/               # Environment-specific configurations
│   ├── dev.json
│   ├── test.json
│   ├── prod.json
│   └── ...
├── providers/                  # Provider-specific configurations
│   ├── openai.json
│   ├── anthropic.json
│   └── ...
└── ...
```

Configuration files must follow these naming conventions:
- Use lowercase with hyphens for readability
- Use `.json` extension for configuration files
- Group related configurations in subdirectories

### `repository-config/` Directory

The `repository-config/` directory contains configuration for template repositories:

```
repository-config/
├── github-sources.json         # GitHub repository sources
├── gitlab-sources.json         # GitLab repository sources
├── database-sources.json       # Database repository sources
├── s3-sources.json             # S3 repository sources
├── local-sources.json          # Local repository sources
└── ...
```

Repository configuration files must follow these naming conventions:
- Use lowercase with hyphens for readability
- Use `.json` extension for configuration files
- Name files based on the repository type they configure

## Manifest Format

The bundle manifest (`manifest.json`) is a JSON file that contains metadata about the bundle and its contents. It must include the following fields:

```json
{
  "schemaVersion": "1.0",
  "bundleId": "uuid-string",
  "bundleType": "templates",
  "name": "Example Bundle",
  "description": "An example bundle for LLM red teaming",
  "version": "1.0.0",
  "createdAt": "2025-05-25T17:57:09-04:00",
  "author": {
    "name": "John Doe",
    "email": "john@example.com",
    "organization": "Example Org"
  },
  "content": [
    {
      "id": "template-001",
      "path": "templates/owasp-llm/llm01-prompt-injection/basic-injection.json",
      "type": "template",
      "version": "1.0.0",
      "description": "A basic prompt injection template",
      "checksum": "sha256:1234567890abcdef"
    }
  ],
  "checksums": {
    "manifest": "sha256:0987654321fedcba",
    "content": {
      "templates/owasp-llm/llm01-prompt-injection/basic-injection.json": "sha256:1234567890abcdef"
    }
  },
  "compatibility": {
    "minVersion": "1.0.0",
    "maxVersion": "",
    "dependencies": [],
    "incompatible": []
  },
  "isIncremental": false,
  "baseVersion": "",
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
  },
  "signature": "base64-encoded-signature"
}
```

## Naming Conventions

### File and Directory Names

- Use lowercase with hyphens for readability
- Avoid spaces and special characters
- Use descriptive names that indicate the purpose
- Use appropriate file extensions

### Path References

- Use forward slashes (`/`) for path separators in all platforms
- Use relative paths from the bundle root
- Include the full path from the bundle root when referencing files in the manifest

Examples:
- Template references in the manifest: `templates/owasp-llm/llm01-prompt-injection/basic-injection.json`
- Documentation references: `documentation/usage.md`

### Version Numbers

- Use semantic versioning (MAJOR.MINOR.PATCH)
- Include the version in binary file names
- Include the version in the manifest

## Incremental Bundles

Incremental bundles contain only the changes from a base version. They must include:

- A reference to the base version in the manifest
- The `isIncremental` flag set to `true`
- Only the files that have changed or been added

The manifest for an incremental bundle must include:

```json
{
  "isIncremental": true,
  "baseVersion": "1.0.0",
  "changelog": [
    {
      "version": "1.1.0",
      "date": "2025-06-01T10:00:00-04:00",
      "changes": ["Added new templates", "Fixed bugs in existing templates"]
    }
  ]
}
```

## Bundle Validation

Bundles must be validated at multiple levels:

1. **Basic Validation**: Validates the manifest structure and required fields
2. **Standard Validation**: Adds validation of checksums and content integrity
3. **Strict Validation**: Adds validation of compatibility, compliance mappings, and directory structure

## Implementation Requirements

The bundle structure must be implemented with the following requirements:

1. **Consistency**: All bundles must follow the same structure
2. **Extensibility**: The structure must allow for future extensions
3. **Compatibility**: The structure must be compatible with different platforms
4. **Security**: The structure must support cryptographic verification
5. **Usability**: The structure must be easy to understand and use

## Conclusion

This specification defines the structure of bundles for the LLMrecon Tool. It provides a standardized way to package, distribute, and validate templates, modules, and related resources. By following this specification, developers can ensure that their bundles are compatible with the LLMrecon Tool and can be easily shared and used by others.
