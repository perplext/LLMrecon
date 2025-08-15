package update

import (
	"github.com/perplext/LLMrecon/src/version"
)

// FormatVersionChangeType formats a version change type for display
func FormatVersionChangeType(changeType version.VersionChangeType) string {
	switch changeType {
	case version.MajorChange:
		return "Major"
	case version.MinorChange:
		return "Minor"
	case version.PatchChange:
		return "Patch"
	case version.PreReleaseChange:
		return "Pre-release"
	case version.BuildChange:
		return "Build"
	default:
		return "Unknown"
	}
