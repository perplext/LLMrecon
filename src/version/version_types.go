package version

// Version is an alias for SemVersion for compatibility
type Version = SemVersion

// VersionChangeType represents the type of version change
type VersionChangeType string

const (
	MajorChange      VersionChangeType = "major"
	MinorChange      VersionChangeType = "minor"
	PatchChange      VersionChangeType = "patch"
	PreReleaseChange VersionChangeType = "prerelease"
	BuildChange      VersionChangeType = "build"
	NoChange         VersionChangeType = "none"
)

// GetChangeType determines the type of change between two versions
func (v *SemVersion) GetChangeType(other *SemVersion) VersionChangeType {
	if v.Major != other.Major {
		return MajorChange
	}
	if v.Minor != other.Minor {
		return MinorChange
	}
	if v.Patch != other.Patch {
		return PatchChange
	}
	if v.Prerelease != other.Prerelease {
		return PreReleaseChange
	}
	if v.Build != other.Build {
		return BuildChange
	}
	return NoChange

// ParseVersion parses a version string into a Version struct
func ParseVersion(versionStr string) (Version, error) {
	v, err := Parse(versionStr)
	if err != nil {
		return Version{}, err
	}
