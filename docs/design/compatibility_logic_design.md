# Compatibility Logic Design

## Overview

This document defines the compatibility logic for the LLMreconing Tool's version management system. It outlines how compatibility is determined between different versions of the core binary, templates, and provider modules, ensuring a consistent and reliable update process.

## Compatibility Relationships

The LLMreconing Tool has several components that need to maintain compatibility:

1. **Core Binary ↔ Templates**: The core binary must be compatible with the template format
2. **Core Binary ↔ Provider Modules**: The core binary must be compatible with provider modules
3. **Templates ↔ Provider Modules**: Some templates may require specific provider capabilities

## Core Binary Compatibility

### Core API Versioning

The core binary will define API versions for different interfaces:

1. **Template API**: Interface for parsing and executing templates
2. **Module API**: Interface for provider modules to integrate with the core
3. **CLI API**: Command-line interface structure and options
4. **REST API**: API endpoints and data formats

Each API will have its own version number, which may increment independently.

### Backward Compatibility Policy

1. **Major Version Changes (X.y.z → X+1.y.z)**:
   - Breaking changes are allowed
   - No backward compatibility guaranteed
   - Explicit migration steps required

2. **Minor Version Changes (x.Y.z → x.Y+1.z)**:
   - Must maintain backward compatibility
   - New features may be added
   - Existing functionality must work as before

3. **Patch Version Changes (x.y.Z → x.y.Z+1)**:
   - Only bug fixes and non-functional changes
   - Full backward compatibility required

### Forward Compatibility

The system will generally not guarantee forward compatibility (older core versions working with newer templates/modules). However, where possible:

1. **Template Forward Compatibility**: Newer template formats should include version information that allows older core versions to gracefully reject incompatible templates.

2. **Module Forward Compatibility**: Modules should detect core version and adjust behavior or provide clear error messages when used with incompatible core versions.

## Template Compatibility

### Template Format Versioning

Templates will include a format version that indicates their structure:

```yaml
id: prompt-injection-basic
name: Basic Prompt Injection Test
format_version: "1.0"
min_core_version: "1.0.0"
```

### Template Compatibility Rules

1. **Format Version Compatibility**:
   - Core binary defines which template format versions it supports
   - Templates specify their format version
   - Compatibility check compares these values

2. **Minimum Core Version**:
   - Templates specify the minimum core version they require
   - System verifies current core version meets this requirement
   - Example: Template requires core ≥ 1.2.0

3. **Category-Specific Compatibility**:
   - Some template categories may have specific compatibility requirements
   - Example: Data leakage templates may require core ≥ 1.1.0 for specific features

### Template Update Compatibility

When updating templates:

1. **Compatible Updates**: Templates with the same format version but newer content
   - Can be updated without core binary changes
   - Example: Adding new prompt injection tests with the same format

2. **Breaking Updates**: Templates with a new format version
   - May require core binary update
   - System should detect and notify user
   - Example: New template format with additional fields

## Provider Module Compatibility

### Module API Versioning

Provider modules will specify:

1. **Module API Version**: The version of the module API they implement
2. **Compatible Core Versions**: Range of core versions they work with

Example module metadata:

```json
{
  "id": "openai",
  "name": "OpenAI Provider",
  "version": "1.2.0",
  "module_api_version": "1.0",
  "compatible_core_versions": ">=1.0.0 <2.0.0"
}
```

### Module Compatibility Rules

1. **API Version Compatibility**:
   - Core binary defines which module API versions it supports
   - Modules specify which API version they implement
   - Compatibility check ensures these match

2. **Core Version Range**:
   - Modules specify a range of compatible core versions
   - System verifies current core version is within this range
   - Example: Module works with core ≥ 1.0.0 and < 2.0.0

3. **Feature-Specific Compatibility**:
   - Some module features may require specific core capabilities
   - Modules can check core version at runtime for feature enablement
   - Example: Advanced logging features require core ≥ 1.2.0

### Module Update Compatibility

When updating modules:

1. **Compatible Updates**: Module updates with the same API version
   - Can be updated without core binary changes
   - Example: Adding support for new LLM models

2. **Breaking Updates**: Module updates with a new API version
   - May require core binary update
   - System should detect and notify user
   - Example: Module using new core features

## Compatibility Verification Process

### Installation-Time Verification

When installing or updating components:

1. **Core Binary Update**:
   - Check compatibility with installed templates and modules
   - Warn if any components will become incompatible
   - Offer to update compatible components

2. **Template Update**:
   - Verify template format version is supported by core
   - Check minimum core version requirement
   - Skip incompatible templates with warning

3. **Module Update**:
   - Verify module API version is supported by core
   - Check core version is within module's compatible range
   - Skip incompatible modules with warning

### Runtime Verification

During operation:

1. **Template Loading**:
   - Verify template format version before execution
   - Check for required features in core
   - Skip incompatible templates with error message

2. **Module Loading**:
   - Verify module API version at initialization
   - Check for required features in core
   - Disable incompatible modules with error message

## Version Constraint Specification

### Constraint Syntax

The system will support the following version constraint syntax:

- `=1.2.3`: Exact version match
- `>1.2.3`: Greater than specified version
- `>=1.2.3`: Greater than or equal to specified version
- `<1.2.3`: Less than specified version
- `<=1.2.3`: Less than or equal to specified version
- `~1.2.3`: Approximately equivalent to version (allows patch level changes)
- `^1.2.3`: Compatible with version (allows minor and patch level changes)
- `1.2.x`: Any patch level of specified major.minor version
- `1.x.x`: Any minor and patch level of specified major version
- `>=1.0.0 <2.0.0`: Range specification (both conditions must be met)

### Constraint Evaluation

Version constraints will be evaluated using the following algorithm:

1. Parse the constraint into operator and version components
2. Parse the target version into major, minor, patch components
3. Apply the operator logic to determine if the constraint is satisfied
4. For range specifications, all conditions must be satisfied

## Compatibility Matrix

The system will maintain a compatibility matrix to document which versions work together:

| Core Version | Template Format Versions | Module API Versions |
|--------------|--------------------------|---------------------|
| 1.0.0        | 1.0                      | 1.0                 |
| 1.1.0        | 1.0, 1.1                 | 1.0                 |
| 1.2.0        | 1.0, 1.1, 1.2            | 1.0, 1.1            |
| 2.0.0        | 1.1, 1.2, 2.0            | 1.1, 2.0            |

This matrix will be used to:
- Document compatibility for users
- Guide development of new versions
- Inform automated compatibility checks

## Handling Incompatibilities

### User Notifications

When incompatibilities are detected:

1. **Clear Error Messages**:
   - Explain which components are incompatible
   - Specify version requirements
   - Suggest resolution steps

2. **Update Recommendations**:
   - Recommend updates to resolve incompatibilities
   - Provide command examples
   - Link to documentation

### Graceful Degradation

When possible, the system should gracefully degrade rather than fail completely:

1. **Skip Incompatible Templates**:
   - Continue with compatible templates
   - Report skipped templates in summary

2. **Disable Incompatible Modules**:
   - Continue with compatible modules
   - Report disabled modules in summary

3. **Limit Feature Availability**:
   - Disable features that require incompatible components
   - Clearly indicate limited functionality

## Implementation Approach

### Compatibility Check Implementation

The compatibility logic will be implemented as:

1. **Version Parser**: Parses version strings into comparable components
2. **Constraint Parser**: Parses constraint strings into operators and versions
3. **Compatibility Checker**: Applies rules to determine compatibility
4. **Notification System**: Generates appropriate messages for users

### Code Structure

```
version/
  ├── parser.go       # Version and constraint parsing
  ├── compatibility.go # Compatibility checking logic
  ├── constraints.go  # Version constraint definitions
  └── notification.go # User notification handling
```

### Testing Strategy

Comprehensive testing will include:

1. **Unit Tests**:
   - Test version parsing
   - Test constraint evaluation
   - Test compatibility rules

2. **Integration Tests**:
   - Test with various component combinations
   - Verify correct handling of incompatible components

3. **Edge Cases**:
   - Pre-release versions
   - Build metadata
   - Invalid version strings
   - Complex constraint expressions

## Future Considerations

### Evolving Compatibility Rules

As the tool evolves:

1. **New Component Types**:
   - Define compatibility rules for new component types
   - Integrate with existing compatibility system

2. **Enhanced Constraint Language**:
   - Support more complex constraint expressions
   - Add logical operators (AND, OR, NOT)

3. **Dynamic Compatibility**:
   - Runtime feature detection
   - Conditional feature enablement based on component versions

### Compatibility Database

For complex systems:

1. **Centralized Compatibility Database**:
   - Track tested component combinations
   - Record known issues
   - Provide compatibility API

2. **User Feedback Integration**:
   - Allow users to report compatibility issues
   - Aggregate compatibility data
   - Improve recommendations based on real-world usage

## Conclusion

This compatibility logic design provides a robust framework for ensuring that different versions of the LLMreconing Tool's components work together correctly. By implementing these compatibility checks and rules, we can provide users with a reliable update experience while allowing the tool to evolve over time.
