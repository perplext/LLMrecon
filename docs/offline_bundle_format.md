# Offline Bundle Format

## Overview

The Offline Bundle Format is a comprehensive solution for securely packaging, distributing, and validating LLMreconing templates and modules in air-gapped environments. This format extends the standard bundle format with enhanced features for compliance mapping, versioning, and secure distribution.

## Key Features

- **Air-Gapped Operation**: Designed for secure environments without internet connectivity
- **Compliance Mappings**: Built-in mappings to OWASP LLM Top 10 and ISO/IEC 42001 standards
- **Versioning and Incremental Updates**: Support for versioning and incremental bundle updates
- **Cryptographic Verification**: Digital signatures and checksums for integrity validation
- **Comprehensive Documentation**: Structured documentation included in the bundle
- **Validation Levels**: Multiple validation levels for different security requirements

## Directory Structure

An offline bundle follows this standardized directory structure:

```
offline-bundle/
├── manifest.json             # Enhanced bundle manifest
├── README.md                 # Bundle overview and basic information
├── templates/                # LLM red teaming templates
├── modules/                  # Modules for template execution
├── config/                   # Configuration files
├── resources/                # Additional resources
├── documentation/            # Comprehensive documentation
│   ├── usage.md              # Usage guide
│   ├── installation.md       # Installation instructions
│   └── ...                   # Other documentation
└── compliance/               # Compliance information
    └── mappings.json         # Detailed compliance mappings
```

## Manifest Format

The enhanced manifest extends the standard bundle manifest with additional fields:

```json
{
  "schemaVersion": "1.0",
  "bundleId": "uuid-string",
  "bundleType": "template",
  "name": "Example Offline Bundle",
  "description": "An example offline bundle for LLM red teaming",
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
      "path": "templates/example.json",
      "type": "template",
      "version": "1.0.0",
      "description": "An example template",
      "checksum": "sha256:1234567890abcdef"
    }
  ],
  "checksums": {
    "manifest": "sha256:0987654321fedcba",
    "content": {
      "templates/example.json": "sha256:1234567890abcdef"
    }
  },
  "compatibility": {
    "minVersion": "1.0.0",
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

## Validation Levels

The offline bundle format supports three validation levels:

1. **Basic Validation**: Validates the manifest structure and required fields
2. **Standard Validation**: Adds validation of checksums and content integrity
3. **Strict Validation**: Adds validation of compatibility, compliance mappings, and directory structure

## Command-Line Interface

The offline bundle CLI provides commands for creating, validating, and managing offline bundles:

### Creating a Bundle

```bash
offline-bundle create --name "Example Bundle" --description "An example bundle" --version "1.0.0" --type template --output /path/to/output --author-name "John Doe" --author-email "john@example.com" --author-org "Example Org"
```

### Adding Content

```bash
offline-bundle add-content --bundle /path/to/bundle --source /path/to/template.json --target example.json --type template --id template-001 --version 1.0.0 --description "An example template"
```

### Adding Compliance Mappings

```bash
offline-bundle add-compliance --bundle /path/to/bundle --content-id template-001 --owasp "LLM01:PromptInjection,LLM06:SensitiveInformationDisclosure" --iso "42001:8.2.3,42001:8.2.4"
```

### Adding Documentation

```bash
offline-bundle add-documentation --bundle /path/to/bundle --type usage --source /path/to/usage.md
```

### Validating a Bundle

```bash
offline-bundle validate --bundle /path/to/bundle --level strict
```

### Creating an Incremental Bundle

```bash
offline-bundle incremental --base /path/to/base-bundle --output /path/to/incremental-bundle --version 1.1.0 --changes /path/to/changes.txt
```

### Exporting a Bundle

```bash
offline-bundle export --bundle /path/to/bundle --output /path/to/exported-bundle.zip
```

### Converting Existing Bundles

```bash
offline-bundle convert --bundle /path/to/standard-bundle --output /path/to/offline-bundle --auto-detect-compliance true
```

The `--auto-detect-compliance` flag enables automatic detection of compliance mappings based on template content. This helps quickly establish initial compliance mappings that can be refined later.

### Generating Signing Keys

```bash
offline-bundle keygen --output /path/to/keys
```

## Programmatic Usage

The offline bundle format can also be used programmatically:

```go
package main

import (
	"fmt"
	"os"

	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/security/access/audit/trail"
)

func main() {
	// Generate signing key
	publicKey, privateKey, err := bundle.GenerateKeyPair()
	if err != nil {
		fmt.Printf("Failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	// Create author
	author := bundle.Author{
		Name:         "John Doe",
		Email:        "john@example.com",
		Organization: "Example Org",
	}

	// Create audit trail manager
	auditTrailManager := trail.NewAuditTrailManager(&trail.AuditConfig{
		Enabled:        true,
		LoggingBackend: "file",
		LogDirectory:   "logs/audit",
		RetentionDays:  90,
		SigningEnabled: true,
	})

	// Create offline bundle creator
	creator := bundle.NewOfflineBundleCreator(privateKey, author, os.Stdout, auditTrailManager)

	// Create offline bundle
	offlineBundle, err := creator.CreateOfflineBundle(
		"Example Bundle",
		"An example bundle for LLM red teaming",
		"1.0.0",
		bundle.TemplateBundleType,
		"/path/to/output",
	)
	if err != nil {
		fmt.Printf("Failed to create offline bundle: %v\n", err)
		os.Exit(1)
	}

	// Add content
	err = creator.AddContentToOfflineBundle(
		offlineBundle,
		"/path/to/template.json",
		"example.json",
		bundle.TemplateContentType,
		"template-001",
		"1.0.0",
		"An example template",
	)
	if err != nil {
		fmt.Printf("Failed to add content: %v\n", err)
		os.Exit(1)
	}

	// Add compliance mappings
	err = creator.AddComplianceMappingToOfflineBundle(
		offlineBundle,
		"template-001",
		[]string{"LLM01:PromptInjection", "LLM06:SensitiveInformationDisclosure"},
		[]string{"42001:8.2.3", "42001:8.2.4"},
	)
	if err != nil {
		fmt.Printf("Failed to add compliance mappings: %v\n", err)
		os.Exit(1)
	}

	// Validate offline bundle
	validator := bundle.NewOfflineBundleValidator(os.Stdout)
	result, err := validator.ValidateOfflineBundle(offlineBundle, bundle.StrictValidation)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	if !result.Valid {
		fmt.Printf("Bundle validation failed: %s\n", result.Message)
		for _, err := range result.Errors {
			fmt.Printf("- %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Bundle validation successful\n")
}
```

## Security Considerations

- **Key Management**: Securely store and manage signing keys
- **Air-Gapped Transfer**: Use secure methods for transferring bundles to air-gapped environments
- **Validation**: Always validate bundles at the strictest level appropriate for your security requirements
- **Audit Logging**: Enable audit logging for all bundle operations
- **Compliance**: Regularly update compliance mappings as standards evolve

## Best Practices

1. **Version Control**: Maintain clear versioning for all bundles
2. **Documentation**: Include comprehensive documentation with all bundles
3. **Compliance Mapping**: Map all templates to relevant compliance standards
4. **Incremental Updates**: Use incremental bundles for updates to minimize transfer size
5. **Validation**: Validate bundles before and after transfer to air-gapped environments
6. **Signing**: Always sign bundles with a trusted key
7. **Audit Trail**: Maintain an audit trail of all bundle operations
