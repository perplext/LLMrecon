# Bundle Structure Specification

## Overview

This document provides a comprehensive specification for the structure of offline bundles in the LLMreconing Tool. It defines the directory layout, file organization, naming conventions, and resource placement within bundles. This specification serves as the official reference for creating, validating, and working with bundles in the LLMrecon Tool ecosystem.

## Bundle Format

Offline bundles are structured as directories with a standardized layout. The bundle can be distributed in two formats:

1. **Directory Format**: A directory containing all bundle files and subdirectories
2. **Archive Format**: A ZIP archive containing the same structure for easier transfer in air-gapped environments

Both formats must maintain the same internal structure and follow the same naming conventions.

## Root Directory Structure

The root directory of an offline bundle contains the following elements:

| Path | Type | Required | Description |
|------|------|----------|-------------|
| `manifest.json` | File | Yes | Enhanced bundle manifest containing metadata, content inventory, and compliance mappings |
| `README.md` | File | Yes | Human-readable overview of the bundle and its contents |
| `templates/` | Directory | Yes | Contains all template files organized by category |
| `modules/` | Directory | No | Contains modules that extend the functionality of templates |
| `binary/` | Directory | No | Contains binary executables for the LLMreconing Tool |
| `documentation/` | Directory | Yes | Contains comprehensive documentation for the bundle |
| `signatures/` | Directory | Yes | Contains cryptographic signatures for bundle verification |
| `resources/` | Directory | No | Contains additional resources used by templates or modules |
| `config/` | Directory | No | Contains configuration files for templates and modules |
| `repository-config/` | Directory | No | Contains configuration for template repositories |

Each of these elements has a specific purpose and structure, which are detailed in the following sections.

## Directory Structure Details

### `templates/` Directory

The `templates/` directory contains all template files organized by category. The recommended structure is:

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

Template files should follow these naming conventions:
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

Module files should follow these naming conventions:
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

Binary files should follow these naming conventions:
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

Documentation files should follow these naming conventions:
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

Signature files should follow these naming conventions:
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
│   ├── ...
└── ...
```

Resource files should follow these naming conventions:
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

Configuration files should follow these naming conventions:
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

Repository configuration files should follow these naming conventions:
- Use lowercase with hyphens for readability
- Use `.json` extension for configuration files
- Name files based on the repository type they configure
- Name files based on the repository type

## File Naming Conventions

### General Naming Conventions

- Use lowercase letters for all file and directory names
- Use hyphens (`-`) to separate words in file and directory names
- Use descriptive names that indicate the purpose or content
- Avoid spaces, underscores, and special characters in file and directory names
- Use appropriate file extensions for each file type

### Version Naming

- Use semantic versioning (MAJOR.MINOR.PATCH) for version numbers
- Include version numbers in binary file names
- Include version information in the manifest file
- Use version directories when multiple versions need to be stored together

### Category Naming

- Use standardized category names for organizing templates and modules
- For OWASP LLM Top 10 categories, use the format `llmXX-category-name`
- For ISO/IEC 42001 controls, use the format `control-X.X.X`
- For custom categories, use descriptive names with hyphens

## File Formats

### Manifest File

The manifest file (`manifest.json`) must be a valid JSON file that conforms to the enhanced manifest schema defined in the format specification.

### Template Files

Template files must be valid JSON files that conform to the template schema defined in the template format specification.

### Module Files

Module files must be valid JSON files that conform to the module schema defined in the module format specification.

### Documentation Files

Documentation files should be in Markdown format with clear structure, headings, and content organization.

## Content Organization

### Template Organization

Templates should be organized by:
1. Compliance framework (e.g., OWASP LLM Top 10, ISO/IEC 42001)
2. Category within the framework
3. Specific vulnerability or control being tested

### Module Organization

Modules should be organized by:
1. Module type (provider, detector, utility)
2. Specific functionality or purpose

### Resource Organization

Resources should be organized by:
1. Resource type (images, data, schemas)
2. Specific purpose or related template/module

## Incremental Bundles

Incremental bundles follow the same structure as full bundles but contain only the changes from a base bundle. The manifest file includes:
- `isIncremental` flag set to `true`
- `baseVersion` field specifying the version of the base bundle
- Content entries only for new or modified files

## Bundle References

### Internal References

Files within the bundle should be referenced using relative paths from the bundle root. For example:
- Template references in the manifest: `templates/owasp-llm/llm01-prompt-injection/basic-injection.json`
- Documentation references: `documentation/usage.md`

### External References

External references (e.g., URLs, external repositories) should be fully qualified and include:
- Protocol (e.g., `https://`)
- Domain name
- Path
- Version or commit reference (if applicable)

## Validation Rules

The following validation rules apply to the bundle structure:

1. All required directories and files must be present
2. The manifest file must conform to the enhanced manifest schema
3. All referenced files in the manifest must exist in the bundle
4. All files must be in the correct format and location
5. All file and directory names must follow the naming conventions
6. All signatures must be valid for their corresponding files

## Example Bundle Structure

Here's an example of a complete offline bundle structure:

```
example-offline-bundle/
├── manifest.json
├── README.md
├── templates/
│   ├── owasp-llm/
│   │   ├── llm01-prompt-injection/
│   │   │   ├── basic-injection.json
│   │   │   └── advanced-injection.json
│   │   ├── llm02-insecure-output/
│   │   │   └── output-manipulation.json
│   │   └── ...
│   └── iso-42001/
│       ├── control-8.2.3/
│       │   └── data-protection.json
│       └── ...
├── modules/
│   ├── providers/
│   │   ├── openai.json
│   │   └── anthropic.json
│   └── detectors/
│       ├── prompt-injection.json
│       └── ...
├── binary/
│   ├── tool-v1.2.3-linux-amd64
│   ├── tool-v1.2.3-darwin-amd64
│   └── tool-v1.2.3-windows-amd64.exe
├── documentation/
│   ├── usage.md
│   ├── installation.md
│   ├── api-endpoints.md
│   ├── cli-commands.md
│   ├── template-format.md
│   └── compliance/
│       ├── owasp-llm-top10.md
│       └── iso-42001.md
├── signatures/
│   ├── manifest.sig
│   ├── content/
│   │   └── ...
│   └── public-key.pem
├── resources/
│   ├── images/
│   │   └── logo.png
│   └── data/
│       └── sample-prompts.json
├── config/
│   ├── defaults.json
│   └── environments/
│       ├── dev.json
│       ├── test.json
│       └── prod.json
└── repository-config/
    ├── github-sources.json
    └── gitlab-sources.json
```

## Implementation Guidelines

When implementing support for offline bundles, follow these guidelines:

1. **Directory Creation**: Create all required directories when creating a new bundle
2. **File Placement**: Place files in the correct directories based on their type and purpose
3. **Manifest Generation**: Generate a complete manifest that includes all content with correct paths
4. **Validation**: Validate the bundle structure before distribution or import
5. **Path Handling**: Use platform-independent path handling to ensure compatibility across operating systems
6. **Incremental Updates**: Support both full and incremental bundle creation and merging

## Conclusion

This specification provides a standardized structure for offline bundles in the LLMreconing Tool. Following this specification ensures consistency, compatibility, and ease of use across different environments and use cases.
