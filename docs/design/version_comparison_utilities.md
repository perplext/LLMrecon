# Version Comparison Utilities

## Overview

This document defines the version comparison utilities for the LLMreconing Tool. It outlines the algorithms and functions needed to compare versions, check compatibility, and determine when updates are available.

## Version Parsing

### Semantic Version Parser

The system will implement a parser for semantic versions following the SemVer 2.0.0 specification:

```go
// Version represents a semantic version
type Version struct {
    Major      int
    Minor      int
    Patch      int
    PreRelease string
    BuildMeta  string
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(versionStr string) (Version, error) {
    // Implementation to parse "X.Y.Z-prerelease+buildmeta"
    // Returns parsed Version or error for invalid format
}

// String returns the string representation of a Version
func (v Version) String() string {
    // Implementation to convert Version back to string
}
```

### Version Constraint Parser

The system will implement a parser for version constraints:

```go
// Constraint represents a version constraint
type Constraint struct {
    Operator string // =, >, >=, <, <=, ~, ^
    Version  Version
}

// ConstraintSet represents a set of constraints (e.g., ">=1.0.0 <2.0.0")
type ConstraintSet []Constraint

// ParseConstraint parses a constraint string into a ConstraintSet
func ParseConstraint(constraintStr string) (ConstraintSet, error) {
    // Implementation to parse constraint expressions
    // Returns parsed ConstraintSet or error for invalid format
}
```

## Version Comparison

### Basic Comparison

The system will implement comparison functions following SemVer rules:

```go
// Compare compares two versions
// Returns:
//   -1 if v1 < v2
//    0 if v1 == v2
//   +1 if v1 > v2
func Compare(v1, v2 Version) int {
    // Compare major version
    if v1.Major != v2.Major {
        if v1.Major < v2.Major {
            return -1
        }
        return 1
    }
    
    // Compare minor version
    if v1.Minor != v2.Minor {
        if v1.Minor < v2.Minor {
            return -1
        }
        return 1
    }
    
    // Compare patch version
    if v1.Patch != v2.Patch {
        if v1.Patch < v2.Patch {
            return -1
        }
        return 1
    }
    
    // If we get here, the version components are equal
    // Compare pre-release identifiers (no pre-release > pre-release)
    if v1.PreRelease == "" && v2.PreRelease != "" {
        return 1
    }
    if v1.PreRelease != "" && v2.PreRelease == "" {
        return -1
    }
    
    // Compare pre-release identifiers if both have them
    if v1.PreRelease != "" && v2.PreRelease != "" {
        // Implementation of pre-release comparison
        // Following SemVer rules for pre-release precedence
    }
    
    // Build metadata does not affect precedence
    return 0
}

// Equal checks if two versions are equal
func Equal(v1, v2 Version) bool {
    return Compare(v1, v2) == 0
}

// LessThan checks if v1 is less than v2
func LessThan(v1, v2 Version) bool {
    return Compare(v1, v2) < 0
}

// GreaterThan checks if v1 is greater than v2
func GreaterThan(v1, v2 Version) bool {
    return Compare(v1, v2) > 0
}
```

### Constraint Satisfaction

The system will implement functions to check if a version satisfies constraints:

```go
// Satisfies checks if a version satisfies a constraint
func (c Constraint) Satisfies(v Version) bool {
    cmp := Compare(v, c.Version)
    
    switch c.Operator {
    case "=":
        return cmp == 0
    case ">":
        return cmp > 0
    case ">=":
        return cmp >= 0
    case "<":
        return cmp < 0
    case "<=":
        return cmp <= 0
    case "~":
        // Compatible with patch changes
        // e.g., ~1.2.3 matches 1.2.x
        return v.Major == c.Version.Major &&
               v.Minor == c.Version.Minor &&
               v.Patch >= c.Version.Patch
    case "^":
        // Compatible with minor and patch changes
        // e.g., ^1.2.3 matches 1.x.y where x >= 2, y >= 0
        return v.Major == c.Version.Major &&
               (v.Minor > c.Version.Minor ||
                (v.Minor == c.Version.Minor && v.Patch >= c.Version.Patch))
    }
    
    return false
}

// Satisfies checks if a version satisfies all constraints in a set
func (cs ConstraintSet) Satisfies(v Version) bool {
    for _, c := range cs {
        if !c.Satisfies(v) {
            return false
        }
    }
    return true
}
```

## Version Range Handling

### Version Range

The system will support defining and checking version ranges:

```go
// VersionRange represents a range of versions
type VersionRange struct {
    Min           Version
    Max           Version
    IncludeMin    bool
    IncludeMax    bool
}

// Contains checks if a version is within the range
func (vr VersionRange) Contains(v Version) bool {
    if vr.IncludeMin {
        if Compare(v, vr.Min) < 0 {
            return false
        }
    } else {
        if Compare(v, vr.Min) <= 0 {
            return false
        }
    }
    
    if vr.IncludeMax {
        if Compare(v, vr.Max) > 0 {
            return false
        }
    } else {
        if Compare(v, vr.Max) >= 0 {
            return false
        }
    }
    
    return true
}

// ParseRange parses a range expression like ">=1.0.0 <2.0.0"
func ParseRange(rangeStr string) (VersionRange, error) {
    // Implementation to parse range expressions
}
```

### Version Set

The system will support defining and checking sets of versions:

```go
// VersionSet represents a set of versions
type VersionSet []Version

// Contains checks if a version is in the set
func (vs VersionSet) Contains(v Version) bool {
    for _, setV := range vs {
        if Equal(v, setV) {
            return true
        }
    }
    return false
}

// AddVersion adds a version to the set if not already present
func (vs *VersionSet) AddVersion(v Version) {
    if !vs.Contains(v) {
        *vs = append(*vs, v)
    }
}
```

## Update Checking

### Version Difference

The system will implement functions to determine the type of version change:

```go
// VersionChangeType represents the type of version change
type VersionChangeType int

const (
    NoChange VersionChangeType = iota
    PatchChange
    MinorChange
    MajorChange
)

// DetermineChangeType determines the type of change between versions
func DetermineChangeType(oldV, newV Version) VersionChangeType {
    if oldV.Major != newV.Major {
        return MajorChange
    }
    if oldV.Minor != newV.Minor {
        return MinorChange
    }
    if oldV.Patch != newV.Patch {
        return PatchChange
    }
    return NoChange
}
```

### Update Availability

The system will implement functions to check for available updates:

```go
// UpdateInfo represents information about an available update
type UpdateInfo struct {
    Component       string
    CurrentVersion  Version
    LatestVersion   Version
    ChangeType      VersionChangeType
    ChangelogURL    string
}

// CheckForUpdates checks if updates are available for components
func CheckForUpdates(components map[string]Version) ([]UpdateInfo, error) {
    // Implementation to check remote repositories for newer versions
    // Returns list of available updates with change type and changelog URL
}
```

## Version Visualization

### Version Comparison Display

The system will implement functions to display version differences:

```go
// FormatVersionDiff formats the difference between versions
func FormatVersionDiff(oldV, newV Version) string {
    changeType := DetermineChangeType(oldV, newV)
    
    switch changeType {
    case MajorChange:
        return fmt.Sprintf("%s → %s (Major Update)", oldV, newV)
    case MinorChange:
        return fmt.Sprintf("%s → %s (Minor Update)", oldV, newV)
    case PatchChange:
        return fmt.Sprintf("%s → %s (Patch Update)", oldV, newV)
    default:
        return fmt.Sprintf("%s (No Change)", oldV)
    }
}
```

### Version History Visualization

The system will implement functions to visualize version history:

```go
// VersionHistoryEntry represents an entry in version history
type VersionHistoryEntry struct {
    Version      Version
    ReleaseDate  time.Time
    ChangeType   VersionChangeType
    Changes      map[string][]string // Category -> Changes
}

// FormatVersionHistory formats version history for display
func FormatVersionHistory(history []VersionHistoryEntry) string {
    // Implementation to format version history as text
    // with proper indentation and grouping
}
```

## Compatibility Checking

### Component Compatibility

The system will implement functions to check compatibility between components:

```go
// ComponentCompatibility represents compatibility information
type ComponentCompatibility struct {
    Component      string
    Version        Version
    CompatibleWith map[string]ConstraintSet // Component -> Constraints
}

// CheckCompatibility checks if components are compatible
func CheckCompatibility(components map[string]Version, 
                       compatibility []ComponentCompatibility) (bool, []string) {
    // Implementation to check if all components are compatible
    // Returns compatibility status and list of incompatibility messages
}
```

### Compatibility Matrix

The system will implement functions to generate a compatibility matrix:

```go
// CompatibilityMatrix represents a matrix of component compatibility
type CompatibilityMatrix map[string]map[string][]Version

// GenerateCompatibilityMatrix generates a compatibility matrix
func GenerateCompatibilityMatrix(compatibility []ComponentCompatibility) CompatibilityMatrix {
    // Implementation to generate a matrix showing which versions
    // of components are compatible with each other
}
```

## Performance Optimization

### Version Caching

To improve performance, the system will implement version caching:

```go
// VersionCache caches parsed versions
type VersionCache struct {
    versions map[string]Version
    mutex    sync.RWMutex
}

// GetVersion gets a version from cache or parses it
func (vc *VersionCache) GetVersion(versionStr string) (Version, error) {
    vc.mutex.RLock()
    v, found := vc.versions[versionStr]
    vc.mutex.RUnlock()
    
    if found {
        return v, nil
    }
    
    v, err := ParseVersion(versionStr)
    if err != nil {
        return Version{}, err
    }
    
    vc.mutex.Lock()
    vc.versions[versionStr] = v
    vc.mutex.Unlock()
    
    return v, nil
}
```

### Constraint Optimization

To improve constraint checking performance:

```go
// OptimizedConstraintSet represents an optimized set of constraints
type OptimizedConstraintSet struct {
    MinVersion Version
    MaxVersion Version
    ExactSet   VersionSet
    // Additional optimization fields
}

// Optimize converts a ConstraintSet to an OptimizedConstraintSet
func (cs ConstraintSet) Optimize() OptimizedConstraintSet {
    // Implementation to optimize constraint checking
    // by pre-computing min/max bounds and other optimizations
}

// Satisfies checks if a version satisfies the optimized constraints
func (ocs OptimizedConstraintSet) Satisfies(v Version) bool {
    // Implementation of optimized constraint checking
    // using pre-computed bounds for faster rejection
}
```

## Integration with Version Management System

### Version Storage Integration

The version comparison utilities will integrate with the version storage system:

```go
// VersionStorage represents the storage for version information
type VersionStorage interface {
    GetCurrentVersions() (map[string]Version, error)
    GetVersionHistory(component string) ([]VersionHistoryEntry, error)
    GetCompatibilityRules() ([]ComponentCompatibility, error)
    // Additional methods
}

// NewVersionComparator creates a version comparator with storage
func NewVersionComparator(storage VersionStorage) *VersionComparator {
    // Implementation to create a version comparator
    // that uses the provided storage
}
```

### Update System Integration

The version comparison utilities will integrate with the update system:

```go
// UpdateSystem represents the system for applying updates
type UpdateSystem interface {
    CheckForUpdates() ([]UpdateInfo, error)
    ApplyUpdate(component string, version Version) error
    // Additional methods
}

// NewUpdateChecker creates an update checker
func NewUpdateChecker(comparator *VersionComparator, 
                     updateSystem UpdateSystem) *UpdateChecker {
    // Implementation to create an update checker
    // that uses the comparator and update system
}
```

## Command-Line Interface

### CLI Commands

The version comparison utilities will be exposed through CLI commands:

```
# Check current versions
LLMrecon version

# Compare versions
LLMrecon version compare 1.2.3 1.3.0

# Check compatibility
LLMrecon version check-compatibility

# Check for updates
LLMrecon update check
```

### API Endpoints

The version comparison utilities will be exposed through API endpoints:

```
GET /api/v1/version
GET /api/v1/version/compare?v1=1.2.3&v2=1.3.0
GET /api/v1/version/compatibility
GET /api/v1/updates/check
```

## Implementation Considerations

### Language and Libraries

The version comparison utilities will be implemented in Go, using:

1. **Standard Library**: For basic functionality
2. **Semver Library**: Consider using an established semver library
3. **Custom Implementation**: For LLMreconing Tool specific needs

### Error Handling

The implementation will include robust error handling:

1. **Invalid Versions**: Clear error messages for invalid version strings
2. **Invalid Constraints**: Helpful error messages for invalid constraints
3. **Compatibility Issues**: Detailed information about incompatibilities

### Testing Strategy

The implementation will include comprehensive tests:

1. **Unit Tests**: For individual functions
2. **Property Tests**: For version comparison properties
3. **Edge Cases**: Test pre-release versions, build metadata, etc.
4. **Performance Tests**: Ensure efficient operation with large numbers of versions

## Conclusion

The version comparison utilities provide the core functionality needed to implement the version management system for the LLMreconing Tool. These utilities enable:

1. **Accurate Version Comparison**: Following SemVer rules
2. **Constraint Checking**: For compatibility verification
3. **Update Detection**: To identify available updates
4. **Visualization**: To help users understand version differences

By implementing these utilities, we ensure that the version management system can reliably track, compare, and manage versions across all components of the tool.
