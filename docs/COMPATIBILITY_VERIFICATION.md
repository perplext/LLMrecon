# Bundle Compatibility Verification Specification

## Overview

This document specifies the logic and rules for verifying compatibility between bundle versions, components, and target environments. The system ensures that updates and installations maintain system integrity and prevent breaking changes.

## Compatibility Model

### 1. Version Compatibility

#### Semantic Versioning Rules
```
MAJOR.MINOR.PATCH-PRERELEASE+BUILD

Examples:
- 1.2.3 - stable release
- 2.0.0-beta.1 - beta release
- 1.3.0-rc.2+build.123 - release candidate with build metadata
```

#### Compatibility Matrix
| From Version | To Version | Compatibility | Update Path |
|--------------|------------|---------------|-------------|
| 1.0.0 | 1.0.1 | Direct | Patch update |
| 1.0.0 | 1.1.0 | Direct | Minor update |
| 1.0.0 | 2.0.0 | Migration | Major update with migration |
| 1.2.3 | 1.2.2 | Blocked | Downgrade protection |
| 1.0.0 | 1.0.0-beta | Blocked | Stable to prerelease blocked |

### 2. Component Compatibility

#### Component Dependencies
```json
{
  "components": {
    "templates": {
      "minVersion": "1.0.0",
      "maxVersion": "2.0.0",
      "required": true
    },
    "providers": {
      "minVersion": "1.1.0",
      "maxVersion": null,
      "required": true
    },
    "modules": {
      "minVersion": "1.0.0",
      "maxVersion": "1.x",
      "required": false
    }
  }
}
```

#### API Version Compatibility
```json
{
  "apiVersions": {
    "template": "v1",
    "provider": "v2",
    "module": "v1"
  },
  "supportedVersions": {
    "template": ["v1"],
    "provider": ["v1", "v2"],
    "module": ["v1"]
  }
}
```

### 3. Environment Compatibility

#### Platform Requirements
```json
{
  "platforms": {
    "linux": {
      "arch": ["amd64", "arm64"],
      "minKernel": "3.10",
      "glibc": "2.17"
    },
    "darwin": {
      "arch": ["amd64", "arm64"],
      "minVersion": "10.15"
    },
    "windows": {
      "arch": ["amd64"],
      "minVersion": "10",
      "build": 17763
    }
  }
}
```

#### Runtime Dependencies
```json
{
  "runtime": {
    "go": {
      "minVersion": "1.19",
      "required": false
    },
    "docker": {
      "minVersion": "20.10",
      "required": false,
      "features": ["buildkit"]
    }
  }
}
```

## Compatibility Verification Process

### 1. Pre-Installation Checks

```go
type CompatibilityChecker struct {
    currentVersion *Version
    targetVersion  *Version
    environment    *Environment
    config         *CompatibilityConfig
}

func (c *CompatibilityChecker) VerifyCompatibility() (*CompatibilityResult, error) {
    result := &CompatibilityResult{
        Compatible: true,
        Issues:     []CompatibilityIssue{},
        Warnings:   []string{},
    }
    
    // Version compatibility
    if err := c.checkVersionCompatibility(result); err != nil {
        return nil, err
    }
    
    // Component compatibility
    if err := c.checkComponentCompatibility(result); err != nil {
        return nil, err
    }
    
    // Environment compatibility
    if err := c.checkEnvironmentCompatibility(result); err != nil {
        return nil, err
    }
    
    // Feature compatibility
    if err := c.checkFeatureCompatibility(result); err != nil {
        return nil, err
    }
    
    return result, nil
}
```

### 2. Version Compatibility Rules

```go
func (c *CompatibilityChecker) checkVersionCompatibility(result *CompatibilityResult) error {
    current := c.currentVersion
    target := c.targetVersion
    
    // Rule 1: No downgrades in production
    if c.environment.IsProduction() && target.LessThan(current) {
        result.Compatible = false
        result.Issues = append(result.Issues, CompatibilityIssue{
            Type:     "version_downgrade",
            Severity: "error",
            Message:  fmt.Sprintf("Cannot downgrade from %s to %s in production", current, target),
        })
    }
    
    // Rule 2: Major version changes require migration
    if target.Major != current.Major {
        result.Issues = append(result.Issues, CompatibilityIssue{
            Type:     "major_version_change",
            Severity: "warning",
            Message:  "Major version change detected, migration may be required",
            Action:   "run migration tool",
        })
    }
    
    // Rule 3: Check supported upgrade paths
    if !c.isUpgradePathSupported(current, target) {
        result.Compatible = false
        result.Issues = append(result.Issues, CompatibilityIssue{
            Type:     "unsupported_upgrade_path",
            Severity: "error",
            Message:  fmt.Sprintf("Direct upgrade from %s to %s is not supported", current, target),
            Action:   "upgrade to intermediate version first",
        })
    }
    
    return nil
}
```

### 3. Component Compatibility Verification

```go
func (c *CompatibilityChecker) checkComponentCompatibility(result *CompatibilityResult) error {
    components := c.getInstalledComponents()
    requirements := c.targetVersion.ComponentRequirements
    
    for name, req := range requirements {
        component, exists := components[name]
        if !exists && req.Required {
            result.Compatible = false
            result.Issues = append(result.Issues, CompatibilityIssue{
                Type:     "missing_component",
                Severity: "error",
                Message:  fmt.Sprintf("Required component '%s' is missing", name),
            })
            continue
        }
        
        if exists {
            // Check version compatibility
            if !req.IsVersionCompatible(component.Version) {
                result.Compatible = false
                result.Issues = append(result.Issues, CompatibilityIssue{
                    Type:     "incompatible_component",
                    Severity: "error",
                    Message:  fmt.Sprintf("Component '%s' version %s is incompatible (requires %s)",
                        name, component.Version, req.VersionRange),
                })
            }
            
            // Check API compatibility
            if !c.checkAPICompatibility(component, req) {
                result.Issues = append(result.Issues, CompatibilityIssue{
                    Type:     "api_incompatible",
                    Severity: "warning",
                    Message:  fmt.Sprintf("Component '%s' uses deprecated API version", name),
                })
            }
        }
    }
    
    return nil
}
```

### 4. Environment Compatibility Checks

```go
func (c *CompatibilityChecker) checkEnvironmentCompatibility(result *CompatibilityResult) error {
    env := c.environment
    requirements := c.targetVersion.EnvironmentRequirements
    
    // Platform check
    if !requirements.IsPlatformSupported(env.OS, env.Arch) {
        result.Compatible = false
        result.Issues = append(result.Issues, CompatibilityIssue{
            Type:     "unsupported_platform",
            Severity: "error",
            Message:  fmt.Sprintf("Platform %s/%s is not supported", env.OS, env.Arch),
        })
    }
    
    // Disk space check
    if env.AvailableDiskSpace < requirements.MinDiskSpace {
        result.Compatible = false
        result.Issues = append(result.Issues, CompatibilityIssue{
            Type:     "insufficient_disk_space",
            Severity: "error",
            Message:  fmt.Sprintf("Insufficient disk space: %d MB available, %d MB required",
                env.AvailableDiskSpace/1024/1024, requirements.MinDiskSpace/1024/1024),
        })
    }
    
    // Memory check
    if env.AvailableMemory < requirements.MinMemory {
        result.Warnings = append(result.Warnings, 
            fmt.Sprintf("Low memory: %d MB available, %d MB recommended",
                env.AvailableMemory/1024/1024, requirements.MinMemory/1024/1024))
    }
    
    // Runtime dependencies
    for name, req := range requirements.RuntimeDependencies {
        if err := c.checkRuntimeDependency(name, req, result); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Migration Support

### 1. Migration Detection

```go
type MigrationDetector struct {
    fromVersion *Version
    toVersion   *Version
}

func (m *MigrationDetector) IsMigrationRequired() bool {
    // Major version change
    if m.toVersion.Major != m.fromVersion.Major {
        return true
    }
    
    // Check for specific migration points
    migrationPoints := []string{
        "1.5.0", // Template format change
        "2.0.0", // Major architecture change
        "2.3.0", // Database schema update
    }
    
    for _, point := range migrationPoints {
        v, _ := ParseVersion(point)
        if m.fromVersion.LessThan(v) && m.toVersion.GreaterThanOrEqual(v) {
            return true
        }
    }
    
    return false
}
```

### 2. Migration Planning

```go
type MigrationPlan struct {
    Steps           []MigrationStep
    EstimatedTime   time.Duration
    RequiresBackup  bool
    CanRollback     bool
}

func (m *MigrationDetector) CreateMigrationPlan() (*MigrationPlan, error) {
    plan := &MigrationPlan{
        Steps:          []MigrationStep{},
        RequiresBackup: true,
        CanRollback:    true,
    }
    
    // Add migration steps based on version changes
    if m.requiresTemplateFormatMigration() {
        plan.Steps = append(plan.Steps, MigrationStep{
            Name:        "template_format_migration",
            Description: "Migrate templates to new format",
            Handler:     migrateTemplateFormat,
            Rollback:    rollbackTemplateFormat,
        })
    }
    
    if m.requiresDatabaseMigration() {
        plan.Steps = append(plan.Steps, MigrationStep{
            Name:        "database_migration",
            Description: "Update database schema",
            Handler:     migrateDatabaseSchema,
            Rollback:    rollbackDatabaseSchema,
        })
    }
    
    return plan, nil
}
```

## Compatibility Configuration

### 1. Configuration Schema

```yaml
compatibility:
  version: "1.0"
  
  # Version compatibility rules
  versionRules:
    allowDowngrade: false
    allowPrerelease: false
    requireStableToStable: true
    
  # Supported upgrade paths
  upgradePaths:
    - from: "1.0.x"
      to: "1.1.x"
      direct: true
    - from: "1.x"
      to: "2.0.x"
      direct: false
      intermediate: "1.9.x"
      
  # Component requirements
  components:
    templates:
      minVersion: "1.0.0"
      required: true
    providers:
      minVersion: "1.1.0"
      required: true
      
  # Environment requirements
  environment:
    platforms:
      - os: linux
        arch: [amd64, arm64]
      - os: darwin
        arch: [amd64, arm64]
    diskSpace: 1073741824  # 1GB
    memory: 536870912      # 512MB
```

### 2. Override Mechanism

```go
type CompatibilityOverride struct {
    Force              bool     `json:"force"`
    SkipVersionCheck   bool     `json:"skipVersionCheck"`
    SkipComponentCheck bool     `json:"skipComponentCheck"`
    SkipEnvCheck       bool     `json:"skipEnvCheck"`
    AllowedRisks       []string `json:"allowedRisks"`
}

func (c *CompatibilityChecker) ApplyOverrides(override *CompatibilityOverride) {
    if override.Force {
        log.Warn("Forcing compatibility - all checks bypassed")
        c.config.EnforceStrict = false
    }
    
    if override.SkipVersionCheck {
        log.Warn("Skipping version compatibility checks")
        c.config.CheckVersion = false
    }
    
    // Log all overrides for audit trail
    c.logOverrides(override)
}
```

## Testing and Validation

### 1. Compatibility Test Suite

```go
func TestVersionCompatibility(t *testing.T) {
    tests := []struct {
        name     string
        from     string
        to       string
        expected bool
    }{
        {"patch update", "1.0.0", "1.0.1", true},
        {"minor update", "1.0.0", "1.1.0", true},
        {"major update", "1.0.0", "2.0.0", true},
        {"downgrade", "1.1.0", "1.0.0", false},
        {"prerelease", "1.0.0", "1.0.0-beta", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            checker := NewCompatibilityChecker()
            result := checker.CheckVersions(tt.from, tt.to)
            if result.Compatible != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result.Compatible)
            }
        })
    }
}
```

### 2. Integration Testing

```bash
# Test compatibility check
llm-redteam compatibility check --target-version 2.0.0

# Test with override
llm-redteam compatibility check --target-version 2.0.0 --force

# Dry run update with compatibility check
llm-redteam update --version 2.0.0 --dry-run
```

## Error Messages and User Guidance

### 1. Clear Error Messages

```go
var compatibilityErrors = map[string]string{
    "version_downgrade":        "Downgrading to an older version is not allowed in production environments",
    "unsupported_platform":     "Your platform (%s) is not supported by this version",
    "missing_component":        "Required component '%s' is not installed",
    "incompatible_api":         "API version mismatch: component uses %s, but %s is required",
    "insufficient_resources":   "Insufficient system resources: %s",
}
```

### 2. Actionable Recommendations

```go
func (c *CompatibilityChecker) GetRecommendations(result *CompatibilityResult) []string {
    recommendations := []string{}
    
    for _, issue := range result.Issues {
        switch issue.Type {
        case "version_downgrade":
            recommendations = append(recommendations, 
                "Use --force flag to override (not recommended in production)")
        case "unsupported_platform":
            recommendations = append(recommendations,
                "Check the compatibility matrix for supported platforms")
        case "missing_component":
            recommendations = append(recommendations,
                fmt.Sprintf("Install missing component: llm-redteam component install %s", issue.Component))
        }
    }
    
    return recommendations
}
```