// Package version provides utilities for semantic versioning
package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	BuildMeta  string
}

// Regular expression for parsing semantic versions
var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)

// ParseVersion parses a version string into a Version struct
func ParseVersion(versionStr string) (Version, error) {
	matches := semverRegex.FindStringSubmatch(versionStr)
	if matches == nil {
		return Version{}, fmt.Errorf("invalid semantic version: %s", versionStr)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	preRelease := ""
	if len(matches) > 4 && matches[4] != "" {
		preRelease = matches[4]
	}

	buildMeta := ""
	if len(matches) > 5 && matches[5] != "" {
		buildMeta = matches[5]
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
		BuildMeta:  buildMeta,
	}, nil

// String returns the string representation of a Version
func (v Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		result += "-" + v.PreRelease
	}
	if v.BuildMeta != "" {
		result += "+" + v.BuildMeta
	}
	return result

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
		return comparePreRelease(v1.PreRelease, v2.PreRelease)
	}

	// Build metadata does not affect precedence
	return 0

// comparePreRelease compares pre-release identifiers
// Returns:
//   -1 if pr1 < pr2
//    0 if pr1 == pr2
//   +1 if pr1 > pr2
func comparePreRelease(pr1, pr2 string) int {
	if pr1 == pr2 {
		return 0
	}

	// Split pre-release into identifiers
	ids1 := strings.Split(pr1, ".")
	ids2 := strings.Split(pr2, ".")

	// Compare each identifier
	minLen := len(ids1)
	if len(ids2) < minLen {
		minLen = len(ids2)
	}

	for i := 0; i < minLen; i++ {
		id1 := ids1[i]
		id2 := ids2[i]

		// Check if both are numeric
		isNum1 := true
		num1, err := strconv.Atoi(id1)
		if err != nil {
			isNum1 = false
		}

		isNum2 := true
		num2, err := strconv.Atoi(id2)
		if err != nil {
			isNum2 = false
		}

		// Numeric identifiers have lower precedence than non-numeric
		if isNum1 && !isNum2 {
			return -1
		}
		if !isNum1 && isNum2 {
			return 1
		}

		// If both are numeric, compare numerically
		if isNum1 && isNum2 {
			if num1 < num2 {
				return -1
			}
			if num1 > num2 {
				return 1
			}
			continue
		}

		// If both are non-numeric, compare lexically
		if id1 < id2 {
			return -1
		}
		if id1 > id2 {
			return 1
		}
	}

	// If we get here, one is a prefix of the other
	// The shorter one has lower precedence
	if len(ids1) < len(ids2) {
		return -1
	}
	if len(ids1) > len(ids2) {
		return 1
	}

	return 0

// Equal checks if two versions are equal
func Equal(v1, v2 Version) bool {
	return Compare(v1, v2) == 0

// LessThan checks if v1 is less than v2
func LessThan(v1, v2 Version) bool {
	return Compare(v1, v2) < 0

// GreaterThan checks if v1 is greater than v2
func GreaterThan(v1, v2 Version) bool {
	return Compare(v1, v2) > 0

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

// FormatVersionDiff formats the difference between versions
func FormatVersionDiff(oldV, newV Version) string {
	changeType := DetermineChangeType(oldV, newV)

	switch changeType {
	case MajorChange:
		return fmt.Sprintf("%s → %s (Major Update)", oldV.String(), newV.String())
	case MinorChange:
		return fmt.Sprintf("%s → %s (Minor Update)", oldV.String(), newV.String())
	case PatchChange:
		return fmt.Sprintf("%s → %s (Patch Update)", oldV.String(), newV.String())
	default:
		return fmt.Sprintf("%s (No Change)", oldV.String())
	}
}
}
}
}
}
}
}
}
