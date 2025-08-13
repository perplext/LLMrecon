# Incremental Update Format Specification

## Overview

This document specifies the format and process for incremental updates to LLMrecon bundles. Incremental updates allow users to download and apply only the changes between versions, reducing bandwidth usage and update time.

## Design Principles

1. **Efficiency**: Minimize download size and processing time
2. **Reliability**: Ensure atomic updates with rollback capability
3. **Flexibility**: Support partial updates (templates only, binaries only, etc.)
4. **Compatibility**: Maintain backward compatibility with full bundles
5. **Security**: Maintain signature verification throughout the process

## Delta Bundle Format

### 1. Bundle Structure

```
delta-bundle/
├── delta-manifest.json
├── meta/
│   ├── from-version.json
│   ├── to-version.json
│   └── changelog.md
├── operations/
│   ├── add/
│   │   ├── templates/
│   │   └── modules/
│   ├── update/
│   │   ├── binary/
│   │   └── templates/
│   ├── delete/
│   │   └── manifest.json
│   └── patch/
│       └── patches/
├── validation/
│   ├── pre-conditions.json
│   └── post-conditions.json
└── signatures/
    ├── delta.sig
    └── manifest.json
```

### 2. Delta Manifest Schema

```json
{
  "version": "1.0",
  "deltaType": "incremental",
  "fromVersion": "1.2.3",
  "toVersion": "1.2.4",
  "timestamp": "2024-01-15T10:00:00Z",
  "size": {
    "compressed": 1048576,
    "uncompressed": 2097152
  },
  "operations": {
    "add": [
      {
        "path": "templates/owasp-llm/llm11-new-vulnerability.yaml",
        "type": "file",
        "size": 2048,
        "hash": "sha256:abc123..."
      }
    ],
    "update": [
      {
        "path": "binary/llm-redteam",
        "type": "binary",
        "oldHash": "sha256:old123...",
        "newHash": "sha256:new456...",
        "patchAvailable": true,
        "patchSize": 102400
      }
    ],
    "delete": [
      {
        "path": "templates/deprecated/old-test.yaml",
        "type": "file"
      }
    ],
    "patch": [
      {
        "path": "config/default.yaml",
        "type": "config",
        "patchFile": "patches/config-default.patch",
        "algorithm": "bsdiff"
      }
    ]
  },
  "dependencies": {
    "required": ["1.2.3"],
    "compatible": ["1.2.0", "1.2.1", "1.2.2", "1.2.3"]
  },
  "rollback": {
    "supported": true,
    "snapshotRequired": true
  }
}
```

### 3. Operation Types

#### 3.1 Add Operation
- Adds new files to the bundle
- Includes full file content
- Preserves file permissions and metadata

#### 3.2 Update Operation
- Replaces existing files
- Can use binary patches for large files
- Supports both full replacement and delta patches

#### 3.3 Delete Operation
- Removes files from the bundle
- Records deleted paths for rollback

#### 3.4 Patch Operation
- Applies text or binary patches
- Supports multiple patch formats:
  - Unified diff for text files
  - bsdiff for binary files
  - JSON patch (RFC 6902) for structured data

## Update Process

### 1. Pre-Update Phase

```go
type UpdateContext struct {
    CurrentVersion string
    TargetVersion  string
    BundlePath     string
    BackupPath     string
    DryRun         bool
}

func PrepareUpdate(ctx *UpdateContext) (*UpdatePlan, error) {
    // 1. Verify current version
    current, err := GetCurrentVersion(ctx.BundlePath)
    if err != nil {
        return nil, err
    }
    
    // 2. Check compatibility
    if !IsCompatible(current, ctx.TargetVersion) {
        return nil, ErrIncompatibleVersion
    }
    
    // 3. Create update plan
    plan := &UpdatePlan{
        Operations: []Operation{},
        SpaceRequired: 0,
        BackupSize: 0,
    }
    
    // 4. Validate pre-conditions
    if err := ValidatePreConditions(ctx, plan); err != nil {
        return nil, err
    }
    
    return plan, nil
}
```

### 2. Backup Phase

```go
func CreateBackup(ctx *UpdateContext) (*Backup, error) {
    backup := &Backup{
        Version:   ctx.CurrentVersion,
        Timestamp: time.Now(),
        Files:     make(map[string]FileBackup),
    }
    
    // Backup files that will be modified or deleted
    for _, op := range ctx.Plan.Operations {
        if op.Type == "update" || op.Type == "delete" {
            if err := backup.AddFile(op.Path); err != nil {
                return nil, err
            }
        }
    }
    
    // Save backup manifest
    if err := backup.Save(ctx.BackupPath); err != nil {
        return nil, err
    }
    
    return backup, nil
}
```

### 3. Application Phase

```go
func ApplyOperations(ctx *UpdateContext, ops []Operation) error {
    // Apply operations in order
    for _, op := range ops {
        switch op.Type {
        case "add":
            if err := applyAdd(ctx, op); err != nil {
                return err
            }
        case "update":
            if err := applyUpdate(ctx, op); err != nil {
                return err
            }
        case "delete":
            if err := applyDelete(ctx, op); err != nil {
                return err
            }
        case "patch":
            if err := applyPatch(ctx, op); err != nil {
                return err
            }
        }
        
        // Update progress
        ctx.Progress.Update(op)
    }
    
    return nil
}
```

### 4. Validation Phase

```go
func ValidateUpdate(ctx *UpdateContext) error {
    // 1. Verify all files
    for _, file := range ctx.UpdatedFiles {
        if err := verifyFileIntegrity(file); err != nil {
            return err
        }
    }
    
    // 2. Run post-conditions
    if err := ValidatePostConditions(ctx); err != nil {
        return err
    }
    
    // 3. Update version info
    if err := UpdateVersionInfo(ctx); err != nil {
        return err
    }
    
    return nil
}
```

### 5. Rollback Support

```go
func RollbackUpdate(ctx *UpdateContext, backup *Backup) error {
    // 1. Stop if update is in progress
    ctx.Cancel()
    
    // 2. Restore backed up files
    for path, backupFile := range backup.Files {
        if err := restoreFile(path, backupFile); err != nil {
            return err
        }
    }
    
    // 3. Remove added files
    for _, op := range ctx.AppliedOperations {
        if op.Type == "add" {
            os.Remove(op.Path)
        }
    }
    
    // 4. Restore version info
    if err := RestoreVersionInfo(backup.Version); err != nil {
        return err
    }
    
    return nil
}
```

## Binary Patching

### 1. Patch Generation

```go
func GenerateBinaryPatch(oldFile, newFile string) (*BinaryPatch, error) {
    // Use bsdiff algorithm for binary files
    patch := &BinaryPatch{
        Algorithm: "bsdiff",
        OldHash:   calculateHash(oldFile),
        NewHash:   calculateHash(newFile),
    }
    
    // Generate patch
    patchData, err := bsdiff.Generate(oldFile, newFile)
    if err != nil {
        return nil, err
    }
    
    // Compress patch
    patch.Data = compress(patchData)
    patch.CompressedSize = len(patch.Data)
    
    return patch, nil
}
```

### 2. Patch Application

```go
func ApplyBinaryPatch(targetFile string, patch *BinaryPatch) error {
    // Verify target file
    if calculateHash(targetFile) != patch.OldHash {
        return ErrInvalidTargetFile
    }
    
    // Decompress patch
    patchData := decompress(patch.Data)
    
    // Apply patch
    tempFile := targetFile + ".tmp"
    if err := bsdiff.Apply(targetFile, tempFile, patchData); err != nil {
        return err
    }
    
    // Verify result
    if calculateHash(tempFile) != patch.NewHash {
        os.Remove(tempFile)
        return ErrPatchFailed
    }
    
    // Replace original
    return os.Rename(tempFile, targetFile)
}
```

## Validation Rules

### 1. Pre-Conditions

```json
{
  "preConditions": [
    {
      "type": "version",
      "check": "equals",
      "value": "1.2.3"
    },
    {
      "type": "file_exists",
      "path": "binary/llm-redteam",
      "required": true
    },
    {
      "type": "disk_space",
      "required": 104857600,
      "unit": "bytes"
    },
    {
      "type": "permissions",
      "path": ".",
      "required": "write"
    }
  ]
}
```

### 2. Post-Conditions

```json
{
  "postConditions": [
    {
      "type": "version",
      "check": "equals",
      "value": "1.2.4"
    },
    {
      "type": "file_hash",
      "path": "binary/llm-redteam",
      "hash": "sha256:expected123..."
    },
    {
      "type": "command",
      "cmd": "llm-redteam version",
      "expectedOutput": "1.2.4"
    }
  ]
}
```

## Optimization Strategies

### 1. Delta Compression

- Use content-aware compression for different file types
- Template files: Use template-specific delta encoding
- Binary files: Use bsdiff or similar algorithms
- Configuration files: Use JSON/YAML-aware patching

### 2. Smart Chunking

- Break large files into chunks
- Only transfer changed chunks
- Use rolling hash for chunk boundaries

### 3. Parallel Downloads

- Download multiple delta files concurrently
- Verify each chunk independently
- Assemble in correct order

## Security Considerations

### 1. Signature Verification

- Each delta bundle is signed
- Verify signature before applying any operations
- Include operation hashes in signed manifest

### 2. Atomic Updates

- Use temporary files during update
- Only replace files after successful verification
- Maintain transaction log for recovery

### 3. Permission Preservation

- Record and restore file permissions
- Verify permission requirements before update
- Prevent privilege escalation

## CLI Integration

```bash
# Check for updates
llm-redteam update check

# Download and apply incremental update
llm-redteam update apply --incremental

# Apply specific delta bundle
llm-redteam update apply --delta bundle-1.2.3-to-1.2.4.delta

# Dry run to see what would change
llm-redteam update apply --dry-run

# Rollback last update
llm-redteam update rollback

# Force full update (skip incremental)
llm-redteam update apply --full
```

## Error Handling

### Common Errors

1. **VERSION_MISMATCH**: Current version doesn't match expected
2. **INSUFFICIENT_SPACE**: Not enough disk space for update
3. **PATCH_FAILED**: Binary patch application failed
4. **SIGNATURE_INVALID**: Delta bundle signature verification failed
5. **ROLLBACK_FAILED**: Unable to restore previous state

### Recovery Procedures

1. Automatic rollback on failure
2. Manual recovery using backup
3. Force full update as last resort
4. Detailed error logging for debugging