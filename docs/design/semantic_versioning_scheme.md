# Semantic Versioning Scheme for LLMreconing Tool

## Overview

This document defines the semantic versioning strategy for the LLMreconing Tool. It outlines how version numbers will be assigned and incremented for the core binary, vulnerability templates, and provider modules.

## Core Binary Versioning

### Semantic Versioning Format

The core binary will follow the standard Semantic Versioning 2.0.0 specification (MAJOR.MINOR.PATCH):

```
X.Y.Z
```

Where:
- **X** (MAJOR): Incremented for incompatible API changes
- **Y** (MINOR): Incremented for backward-compatible functionality additions
- **Z** (PATCH): Incremented for backward-compatible bug fixes

### Version Increment Rules

1. **MAJOR Version (X)**:
   - Incremented when making incompatible API changes
   - Changes that break backward compatibility
   - Significant architectural changes
   - Changes to core command structure
   - Examples: Command syntax changes, removal of features, changes to output formats

2. **MINOR Version (Y)**:
   - Incremented when adding functionality in a backward-compatible manner
   - New commands or options
   - New features that don't break existing workflows
   - Deprecation notices (marking features for removal in future MAJOR versions)
   - Examples: Adding new scan types, new output formats, additional configuration options

3. **PATCH Version (Z)**:
   - Incremented for backward-compatible bug fixes
   - Performance improvements
   - Security fixes that don't change the API
   - Examples: Fixing incorrect scan results, resolving crashes, improving error messages

### Initial Development Phase

During initial development (before 1.0.0):
- Version 0.y.z will be used
- The public API should not be considered stable
- Breaking changes may occur in minor versions
- Development will progress through 0.1.0, 0.2.0, etc. for feature additions
- Patch versions (0.y.z) will be used for bug fixes

### Stable Release

- Version 1.0.0 defines the stable public API
- After 1.0.0, version increments will strictly follow semantic versioning rules
- Breaking changes will only occur in major version increments

### Pre-release Identifiers

For pre-release versions, the following format will be used:

```
X.Y.Z-identifier
```

Where identifier can be:
- `alpha`: Early testing, unstable
- `beta`: Feature complete for version, undergoing testing
- `rc.N`: Release candidates (e.g., `1.0.0-rc.1`, `1.0.0-rc.2`)

### Build Metadata

Build metadata may be appended using a plus sign:

```
X.Y.Z+metadata
```

Examples:
- `1.0.0+20250516`
- `1.2.3+git.commit.abc123`

Build metadata does not affect version precedence.

## Template Versioning

Templates will use a combination of versioning approaches:

### Individual Template Versioning

Each template will have:
- A unique identifier (e.g., `prompt-injection-basic`)
- A version number following semantic versioning (e.g., `1.2.0`)
- A last-updated timestamp

### Template Collection Versioning

The entire template collection will have:
- A version number (e.g., `1.1.0`) representing the overall state
- This version will increment when:
  - New templates are added (MINOR version increment)
  - Existing templates are updated (PATCH version increment)
  - Template format changes (MAJOR version increment)

### Template Source Tracking

Templates from different sources will be tracked separately:
- Official templates (from GitHub): `github_official` source
- Development templates (from GitLab): `gitlab_dev` source
- Each source will have its own version number

### Template Compatibility

Templates will include:
- Minimum core version required (`min_core_version`)
- Format version to ensure compatibility with the tool

## Module Versioning

Provider modules will follow a versioning scheme similar to the core binary:

### Module Version Format

```
X.Y.Z
```

Following semantic versioning principles:
- **MAJOR**: Incompatible changes to module API
- **MINOR**: New features, backward-compatible
- **PATCH**: Bug fixes, backward-compatible

### Module Compatibility

Each module will specify:
- Compatible core version range (e.g., `>=1.0.0 <2.0.0`)
- Supported provider API versions
- Supported models or features

### Module Update Independence

Modules can be updated independently of the core binary, allowing for:
- Quick updates to support new provider features
- Bug fixes specific to particular providers
- Optimizations for specific LLM platforms

## Version Display Format

When displaying version information to users:

### Core Binary Version

```
LLMreconing Tool v1.2.3
```

### Detailed Version Information

```
LLMreconing Tool v1.2.3
Build: 20250516-abc123
Templates: v1.1.0 (42 templates)
Modules: 2 modules installed
- OpenAI Provider v1.2.0
- Anthropic Provider v1.1.0
```

### Update Status

```
LLMreconing Tool v1.2.3
Updates available:
- Templates: v1.2.0 available (current: v1.1.0)
- OpenAI Provider: v1.3.0 available (current: v1.2.0)
```

## Version Comparison Logic

### Semantic Version Comparison

For comparing semantic versions:
1. Compare MAJOR version numbers
2. If equal, compare MINOR version numbers
3. If equal, compare PATCH version numbers
4. If equal, pre-release versions have lower precedence than normal versions
5. If both have pre-release identifiers, compare them according to SemVer rules

### Version Range Specification

For specifying version ranges:
- `=1.2.3`: Exact version match
- `>1.2.3`: Greater than specified version
- `>=1.2.3`: Greater than or equal to specified version
- `<1.2.3`: Less than specified version
- `<=1.2.3`: Less than or equal to specified version
- `~1.2.3`: Approximately equivalent to version (allows PATCH level changes)
- `^1.2.3`: Compatible with version (allows MINOR and PATCH level changes)
- `1.2.x`: Any PATCH level of specified MAJOR.MINOR version
- `1.x.x`: Any MINOR and PATCH level of specified MAJOR version

## Changelog Management

### Changelog Format

Changelogs will be maintained in a structured format:

```markdown
# Changelog

## [1.2.3] - 2025-05-16

### Added
- New feature A
- New feature B

### Changed
- Improved performance of X
- Updated dependency Y

### Fixed
- Bug in feature Z
- Issue with error handling

## [1.2.2] - 2025-05-01

...
```

### Changelog Categories

Changes will be categorized as:
- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Features that will be removed in upcoming releases
- **Removed**: Features removed in this release
- **Fixed**: Bug fixes
- **Security**: Security vulnerability fixes

## Implementation Guidelines

### Version Bumping

1. **Automated Tools**:
   - Use automated tools to bump versions based on commit messages
   - Consider conventional commits format for automatic version determination

2. **Manual Review**:
   - Major version bumps should always be manually reviewed
   - Ensure breaking changes are well-documented

3. **Release Process**:
   - Tag releases in version control
   - Generate changelogs automatically where possible
   - Include version information in build artifacts

### Version Storage

1. **Core Binary**:
   - Embed version in binary at build time
   - Make version accessible via CLI command (`version`)

2. **Templates and Modules**:
   - Store version information in manifest files
   - Include version in individual template/module files
   - Update manifests during update process

## Environment-Specific Considerations

### Development Environment
- Use pre-release identifiers (alpha, beta, rc)
- May use build metadata to indicate development builds
- Example: `1.2.3-alpha+dev.20250516`

### Test Environment
- Use release candidate identifiers
- Include test-specific metadata if needed
- Example: `1.2.3-rc.2+test.20250516`

### Production Environment
- Use stable release versions without pre-release identifiers
- May include build date or commit hash as metadata
- Example: `1.2.3+20250516`

## Version Management Responsibilities

### Development Team
- Determine appropriate version increments based on changes
- Maintain changelogs
- Ensure compatibility between components

### Release Management
- Verify version numbers follow the scheme
- Ensure changelogs are complete and accurate
- Communicate version changes to users

### Users
- Specify version constraints for dependencies
- Report version-specific issues
- Follow update recommendations

## Conclusion

This semantic versioning scheme provides a clear and consistent approach to versioning the LLMreconing Tool and its components. By following these guidelines, we ensure:

1. Users can understand the impact of updates
2. Dependencies can be managed effectively
3. Breaking changes are clearly communicated
4. Components can evolve independently while maintaining compatibility

The scheme aligns with industry best practices while addressing the specific needs of the LLMreconing Tool's architecture.
