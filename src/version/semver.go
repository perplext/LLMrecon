package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemVersion represents a semantic version
type SemVersion struct {
	// Major version number
	Major int
	
	// Minor version number
	Minor int
	
	// Patch version number
	Patch int
	
	// Prerelease identifiers (e.g., "alpha.1", "beta.2")
	Prerelease string
	
	// Build metadata (e.g., "build.123")
	Build string
}

// semverRegex is a regular expression for parsing semantic versions
var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-.]+))?(?:\+([0-9A-Za-z-.]+))?$`)

// Parse parses a version string into a SemVersion object
func Parse(version string) (*SemVersion, error) {
	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return nil, fmt.Errorf("invalid version: %s", version)
	}
	
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}
	
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches[2])
	}
	
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", matches[3])
	}
	
	prerelease := matches[4]
	build := matches[5]
	
	return &SemVersion{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Build:      build,
	}, nil
}

// MustParse parses a version string into a SemVersion object
// It panics if the version string is invalid
func MustParse(version string) *SemVersion {
	v, err := Parse(version)
	if err != nil {
		panic(err)
	}
	return v
}

// String returns the string representation of a version
func (v *SemVersion) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	
	if v.Prerelease != "" {
		result += "-" + v.Prerelease
	}
	
	if v.Build != "" {
		result += "+" + v.Build
	}
	
	return result
}

// Compare compares two versions
// Returns -1 if v < other, 0 if v == other, 1 if v > other
func (v *SemVersion) Compare(other *SemVersion) int {
	// Compare major version
	if v.Major < other.Major {
		return -1
	}
	if v.Major > other.Major {
		return 1
	}
	
	// Compare minor version
	if v.Minor < other.Minor {
		return -1
	}
	if v.Minor > other.Minor {
		return 1
	}
	
	// Compare patch version
	if v.Patch < other.Patch {
		return -1
	}
	if v.Patch > other.Patch {
		return 1
	}
	
	// Compare prerelease
	// No prerelease is greater than any prerelease
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}
	if v.Prerelease != "" && other.Prerelease != "" {
		return comparePrerelease(v.Prerelease, other.Prerelease)
	}
	
	// Versions are equal
	return 0
}

// comparePrerelease compares two prerelease strings
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func comparePrerelease(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	
	// Compare each part
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		// Check if parts are numeric
		aNum, aErr := strconv.Atoi(aParts[i])
		bNum, bErr := strconv.Atoi(bParts[i])
		
		// If both parts are numeric, compare them as numbers
		if aErr == nil && bErr == nil {
			if aNum < bNum {
				return -1
			}
			if aNum > bNum {
				return 1
			}
			continue
		}
		
		// If only one part is numeric, numeric is less
		if aErr == nil && bErr != nil {
			return -1
		}
		if aErr != nil && bErr == nil {
			return 1
		}
		
		// Both parts are non-numeric, compare them lexically
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	
	// If all parts are equal, the one with fewer parts is less
	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}
	
	// Prereleases are equal
	return 0
}

// LessThan returns true if v < other
func (v *SemVersion) LessThan(other *SemVersion) bool {
	return v.Compare(other) < 0
}

// GreaterThan returns true if v > other
func (v *SemVersion) GreaterThan(other *SemVersion) bool {
	return v.Compare(other) > 0
}

// Equal returns true if v == other
func (v *SemVersion) Equal(other *SemVersion) bool {
	return v.Compare(other) == 0
}

// IncrementMajor increments the major version and resets minor and patch to 0
func (v *SemVersion) IncrementMajor() *SemVersion {
	return &SemVersion{
		Major:      v.Major + 1,
		Minor:      0,
		Patch:      0,
		Prerelease: "",
		Build:      "",
	}
}

// IncrementMinor increments the minor version and resets patch to 0
func (v *SemVersion) IncrementMinor() *SemVersion {
	return &SemVersion{
		Major:      v.Major,
		Minor:      v.Minor + 1,
		Patch:      0,
		Prerelease: "",
		Build:      "",
	}
}

// IncrementPatch increments the patch version
func (v *SemVersion) IncrementPatch() *SemVersion {
	return &SemVersion{
		Major:      v.Major,
		Minor:      v.Minor,
		Patch:      v.Patch + 1,
		Prerelease: "",
		Build:      "",
	}
}

// WithPrerelease returns a new version with the given prerelease string
func (v *SemVersion) WithPrerelease(prerelease string) *SemVersion {
	return &SemVersion{
		Major:      v.Major,
		Minor:      v.Minor,
		Patch:      v.Patch,
		Prerelease: prerelease,
		Build:      v.Build,
	}
}

// WithBuild returns a new version with the given build string
func (v *SemVersion) WithBuild(build string) *SemVersion {
	return &SemVersion{
		Major:      v.Major,
		Minor:      v.Minor,
		Patch:      v.Patch,
		Prerelease: v.Prerelease,
		Build:      build,
	}
}

// IsCompatible checks if the version is compatible with the given version
// using semantic versioning rules (major version must match)
func (v *SemVersion) IsCompatible(other *SemVersion) bool {
	return v.Major == other.Major
}

// IsBackwardsCompatible checks if the version is backwards compatible
// with the given version using semantic versioning rules
// (major version must match and be >= the other version)
func (v *SemVersion) IsBackwardsCompatible(other *SemVersion) bool {
	// Major versions must match for backwards compatibility
	if v.Major != other.Major {
		return false
	}
	
	// If major versions match, this version must be >= the other version
	return v.Compare(other) >= 0
}
