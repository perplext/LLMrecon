# Update Package Implementation Guide

## Overview

This document provides implementation guidelines for creating, verifying, and applying update packages for the LLMreconing Tool. The update package format is designed to ensure secure, reliable, and efficient updates while supporting differential updates to minimize download size.

## Creating Update Packages

### Prerequisites

To create an update package, you need:

1. The LLMreconing Tool with the `package` command
2. A valid manifest file in JSON format
3. The files to include in the package
4. A private key for signing the package (for official releases)

### Manifest File

The manifest file (`manifest.json`) contains metadata and integrity information for the update package. Here's an example:

```json
{
  "schema_version": "1.0",
  "package_id": "LLMrecon-update-20250518",
  "package_type": "full",
  "created_at": "2025-05-18T12:00:00Z",
  "expires_at": "2025-06-18T12:00:00Z",
  "publisher": {
    "name": "LLMrecon Project",
    "url": "https://github.com/perplext/LLMrecon",
    "public_key_id": "SHA256:abcdef1234567890"
  },
  "components": {
    "binary": {
      "version": "1.2.0",
      "platforms": ["linux", "darwin", "windows"],
      "min_version": "1.0.0",
      "required": true,
      "changelog_url": "https://github.com/perplext/LLMrecon/releases/tag/v1.2.0",
      "checksums": {
        "linux": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        "darwin": "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
        "windows": "sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
      }
    },
    "templates": {
      "version": "1.5.0",
      "min_version": "1.0.0",
      "required": false,
      "changelog_url": "https://github.com/perplext/LLMrecon/wiki/Template-Changes-1.5.0",
      "checksum": "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
      "categories": [
        "owasp-llm",
        "custom"
      ],
      "template_count": 150
    },
    "modules": [
      {
        "id": "openai-provider",
        "name": "OpenAI Provider Module",
        "version": "2.1.0",
        "min_version": "1.0.0",
        "required": false,
        "changelog_url": "https://github.com/perplext/LLMrecon/wiki/OpenAI-Provider-2.1.0",
        "checksum": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        "dependencies": [
          {
            "id": "core",
            "min_version": "1.1.0"
          }
        ]
      }
    ]
  },
  "compliance": {
    "owasp_llm_top10": {
      "version": "2025.1",
      "coverage": [
        "LLM01", "LLM02", "LLM03", "LLM04", "LLM05",
        "LLM06", "LLM07", "LLM08", "LLM09", "LLM10"
      ]
    },
    "iso_42001": {
      "version": "2023",
      "controls": [
        "5.2.1", "5.2.2", "5.3.1", "6.1.2", "7.1.1"
      ]
    }
  },
  "signature": "base64:ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz+/="
}
```

### Directory Structure

Before creating a package, organize your files in the following structure:

```
update-package/
├── manifest.json
├── binary/
│   ├── linux/
│   │   └── LLMrecon
│   ├── darwin/
│   │   └── LLMrecon
│   └── windows/
│       └── LLMrecon.exe
├── templates/
│   ├── owasp-llm/
│   │   ├── llm01-prompt-injection/
│   │   │   ├── basic.yaml
│   │   │   └── advanced.yaml
│   │   ├── llm02-insecure-output/
│   │   └── ...
│   └── custom/
├── modules/
│   ├── openai-provider/
│   │   ├── module.yaml
│   │   └── ...
├── signatures/
│   ├── binary.sig
│   ├── templates.sig
│   └── modules.sig
└── README.md
```

### Creating a Package

Use the `package create` command to create an update package:

```bash
LLMrecon package create manifest.json update-package.zip
```

This command will:
1. Read the manifest file
2. Calculate checksums for all components
3. Sign the manifest (if a private key is provided)
4. Create a ZIP archive with all files

### Signing Packages

For official releases, packages should be signed with the project's private key. The signing process is as follows:

1. Create a manifest without the signature field
2. Calculate the SHA-256 hash of the manifest JSON
3. Sign the hash with the private key using Ed25519
4. Add the signature to the manifest

```bash
# Generate a key pair (if you don't have one)
openssl genpkey -algorithm ED25519 -out private.pem
openssl pkey -in private.pem -pubout -out public.pem

# Sign the manifest
LLMrecon package sign manifest.json private.pem
```

## Creating Differential Updates

Differential updates contain only the changes between two versions, reducing download size. To create a differential update:

1. Prepare a manifest with `package_type` set to `"differential"`
2. Include the `patches` section in the `components` field
3. Generate binary patches using bsdiff
4. Generate YAML/JSON patches for templates and modules

Example patches section in the manifest:

```json
"patches": {
  "binary": [
    {
      "from_version": "1.1.0",
      "to_version": "1.2.0",
      "platforms": ["linux", "darwin", "windows"],
      "checksums": {
        "linux": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        "darwin": "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
        "windows": "sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
      }
    }
  ],
  "templates": [
    {
      "from_version": "1.4.0",
      "to_version": "1.5.0",
      "checksum": "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
    }
  ],
  "modules": [
    {
      "id": "openai-provider",
      "from_version": "2.0.0",
      "to_version": "2.1.0",
      "checksum": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    }
  ]
}
```

### Generating Binary Patches

Use bsdiff to generate binary patches:

```bash
bsdiff old-binary new-binary patch-file
```

### Generating YAML/JSON Patches

Use JSON Patch (RFC 6902) for templates and modules:

```json
[
  { "op": "replace", "path": "/version", "value": "1.5.0" },
  { "op": "add", "path": "/templates/llm01-prompt-injection/advanced.yaml", "value": "..." },
  { "op": "remove", "path": "/templates/deprecated-template.yaml" }
]
```

## Verifying Update Packages

Before applying an update package, verify its integrity and authenticity:

```bash
LLMrecon package verify update-package.zip --public-key public.pem
```

This command will:
1. Extract and parse the manifest
2. Verify the manifest signature using the public key
3. Verify checksums for all components
4. Check if the package has expired

## Applying Update Packages

To apply an update package:

```bash
LLMrecon package apply update-package.zip --public-key public.pem
```

This command will:
1. Verify the package (if not skipped)
2. Check compatibility with the current installation
3. Create backups of existing files
4. Apply the update
5. Update version information

### Options

- `--install-dir`: Installation directory (default: executable directory)
- `--temp-dir`: Temporary directory for update operations
- `--backup-dir`: Backup directory for update operations
- `--force`: Force update even if not compatible
- `--skip-verify`: Skip package verification (not recommended)

## Update Process

The update process follows these steps:

1. **Preparation**:
   - Create temporary directories for the update
   - Create backup directories for rollback

2. **Verification**:
   - Verify package signature
   - Verify component checksums
   - Check compatibility

3. **Backup**:
   - Create backups of all files to be updated

4. **Update**:
   - For full updates: Extract and replace files
   - For differential updates: Apply patches

5. **Finalization**:
   - Update version information
   - Clean up temporary files

6. **Rollback** (if needed):
   - Restore files from backups if any step fails

## Security Considerations

### Integrity Protection

All components include checksums to verify integrity:

```json
"checksums": {
  "linux": "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
  "darwin": "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
  "windows": "sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
}
```

### Authenticity Protection

Digital signatures ensure the package comes from a trusted source:

```json
"signature": "base64:ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz+/="
```

### Expiration

Packages include an expiration date to prevent replay attacks:

```json
"expires_at": "2025-06-18T12:00:00Z"
```

### Required Updates

Critical security updates can be marked as required:

```json
"required": true
```

## Best Practices

1. **Always Sign Packages**: Official packages should always be signed with the project's private key.

2. **Use Differential Updates**: For frequent updates, use differential updates to reduce download size.

3. **Include Changelogs**: Always include changelog URLs in the manifest for transparency.

4. **Set Reasonable Expiration**: Set an expiration date that gives users enough time to update.

5. **Backup Before Updating**: Always create backups before applying updates.

6. **Verify Before Applying**: Always verify packages before applying them.

7. **Test Updates**: Test updates on all supported platforms before releasing.

8. **Document Changes**: Document all changes in the update package.

## Troubleshooting

### Package Verification Fails

If package verification fails, check:
- Is the public key correct?
- Has the package been tampered with?
- Has the package expired?

### Update Application Fails

If update application fails, check:
- Is the package compatible with your current version?
- Do you have write permissions to the installation directory?
- Is the tool currently running?

### Rollback After Failed Update

If an update fails and automatic rollback doesn't work:
1. Find the backup directory (usually in `<install-dir>/backups/<package-id>`)
2. Manually restore files from the backup
3. Restart the tool

## Example: Creating a Full Update Package

Here's a complete example of creating a full update package:

1. Prepare the directory structure:
   ```bash
   mkdir -p update-package/binary/{linux,darwin,windows}
   mkdir -p update-package/templates/owasp-llm
   mkdir -p update-package/modules/openai-provider
   mkdir -p update-package/signatures
   ```

2. Copy files to the package directory:
   ```bash
   cp build/linux/LLMrecon update-package/binary/linux/
   cp build/darwin/LLMrecon update-package/binary/darwin/
   cp build/windows/LLMrecon.exe update-package/binary/windows/
   cp -r templates/* update-package/templates/
   cp -r modules/* update-package/modules/
   ```

3. Create the manifest file:
   ```bash
   cat > update-package/manifest.json << EOF
   {
     "schema_version": "1.0",
     "package_id": "LLMrecon-update-1.2.0",
     "package_type": "full",
     "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
     "expires_at": "$(date -u -v+30d +%Y-%m-%dT%H:%M:%SZ)",
     "publisher": {
       "name": "LLMrecon Project",
       "url": "https://github.com/perplext/LLMrecon",
       "public_key_id": "SHA256:abcdef1234567890"
     },
     "components": {
       "binary": {
         "version": "1.2.0",
         "platforms": ["linux", "darwin", "windows"],
         "min_version": "1.0.0",
         "required": true,
         "changelog_url": "https://github.com/perplext/LLMrecon/releases/tag/v1.2.0",
         "checksums": {}
       },
       "templates": {
         "version": "1.5.0",
         "min_version": "1.0.0",
         "required": false,
         "changelog_url": "https://github.com/perplext/LLMrecon/wiki/Template-Changes-1.5.0",
         "checksum": "",
         "categories": ["owasp-llm", "custom"],
         "template_count": 150
       },
       "modules": []
     },
     "compliance": {
       "owasp_llm_top10": {
         "version": "2025.1",
         "coverage": ["LLM01", "LLM02", "LLM03", "LLM04", "LLM05", "LLM06", "LLM07", "LLM08", "LLM09", "LLM10"]
       },
       "iso_42001": {
         "version": "2023",
         "controls": ["5.2.1", "5.2.2", "5.3.1", "6.1.2", "7.1.1"]
       }
     }
   }
   EOF
   ```

4. Calculate checksums and update the manifest:
   ```bash
   # This would be done by the package create command
   ```

5. Sign the manifest:
   ```bash
   # This would be done by the package create command
   ```

6. Create the package:
   ```bash
   LLMrecon package create update-package/manifest.json LLMrecon-update-1.2.0.zip
   ```

## Example: Creating a Differential Update Package

Here's a complete example of creating a differential update package:

1. Prepare the directory structure:
   ```bash
   mkdir -p update-package/patches/binary/{linux,darwin,windows}
   mkdir -p update-package/patches/templates
   mkdir -p update-package/patches/modules/openai-provider
   mkdir -p update-package/signatures
   ```

2. Generate binary patches:
   ```bash
   bsdiff old-linux/LLMrecon new-linux/LLMrecon update-package/patches/binary/linux/1.1.0-1.2.0.patch
   bsdiff old-darwin/LLMrecon new-darwin/LLMrecon update-package/patches/binary/darwin/1.1.0-1.2.0.patch
   bsdiff old-windows/LLMrecon.exe new-windows/LLMrecon.exe update-package/patches/binary/windows/1.1.0-1.2.0.patch
   ```

3. Generate template patches:
   ```bash
   # Generate JSON patch
   jq -n '[
     { "op": "replace", "path": "/version", "value": "1.5.0" },
     { "op": "add", "path": "/templates/llm01-prompt-injection/advanced.yaml", "value": "..." }
   ]' > update-package/patches/templates/1.4.0-1.5.0.patch
   ```

4. Create the manifest file with patches information:
   ```bash
   cat > update-package/manifest.json << EOF
   {
     "schema_version": "1.0",
     "package_id": "LLMrecon-update-1.1.0-1.2.0",
     "package_type": "differential",
     "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
     "expires_at": "$(date -u -v+30d +%Y-%m-%dT%H:%M:%SZ)",
     "publisher": {
       "name": "LLMrecon Project",
       "url": "https://github.com/perplext/LLMrecon",
       "public_key_id": "SHA256:abcdef1234567890"
     },
     "components": {
       "binary": {
         "version": "1.2.0",
         "platforms": ["linux", "darwin", "windows"],
         "min_version": "1.0.0",
         "required": true,
         "changelog_url": "https://github.com/perplext/LLMrecon/releases/tag/v1.2.0",
         "checksums": {}
       },
       "templates": {
         "version": "1.5.0",
         "min_version": "1.0.0",
         "required": false,
         "changelog_url": "https://github.com/perplext/LLMrecon/wiki/Template-Changes-1.5.0",
         "checksum": "",
         "categories": ["owasp-llm", "custom"],
         "template_count": 150
       },
       "modules": [],
       "patches": {
         "binary": [
           {
             "from_version": "1.1.0",
             "to_version": "1.2.0",
             "platforms": ["linux", "darwin", "windows"],
             "checksums": {}
           }
         ],
         "templates": [
           {
             "from_version": "1.4.0",
             "to_version": "1.5.0",
             "checksum": ""
           }
         ],
         "modules": []
       }
     },
     "compliance": {
       "owasp_llm_top10": {
         "version": "2025.1",
         "coverage": ["LLM01", "LLM02", "LLM03", "LLM04", "LLM05", "LLM06", "LLM07", "LLM08", "LLM09", "LLM10"]
       },
       "iso_42001": {
         "version": "2023",
         "controls": ["5.2.1", "5.2.2", "5.3.1", "6.1.2", "7.1.1"]
       }
     }
   }
   EOF
   ```

5. Calculate checksums and update the manifest:
   ```bash
   # This would be done by the package create command
   ```

6. Sign the manifest:
   ```bash
   # This would be done by the package create command
   ```

7. Create the package:
   ```bash
   LLMrecon package create update-package/manifest.json LLMrecon-update-1.1.0-1.2.0.zip
   ```

## Conclusion

The update package format provides a secure, reliable, and efficient way to distribute updates for the LLMreconing Tool. By following the guidelines in this document, you can create and apply update packages that ensure the integrity and authenticity of updates while minimizing download size.

For more information, see the [Update Package Format Specification](UPDATE_PACKAGE_FORMAT.md).
