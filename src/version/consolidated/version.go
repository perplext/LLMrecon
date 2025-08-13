package version

import (
)

// VersionInfo represents version information for a template or module
type VersionInfo struct {
	// ID is the unique identifier for the template or module
	ID string
	
	// Version is the semantic version
	Version *SemVersion
	
	// CreatedAt is the creation timestamp
	CreatedAt time.Time
	
	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time
	
	// Author is the author of the template or module
	Author string
	
	// Description is a description of the template or module
	Description string
	
	// Tags are tags associated with the template or module
	Tags []string
	
	// Metadata is additional metadata for the template or module
	Metadata map[string]interface{}
}

// NewVersionInfo creates a new version info object
func NewVersionInfo(id string, version string, author string, description string) (*VersionInfo, error) {
	semver, err := Parse(version)
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	
	return &VersionInfo{
		ID:          id,
		Version:     semver,
		CreatedAt:   now,
		UpdatedAt:   now,
		Author:      author,
		Description: description,
		Tags:        []string{},
		Metadata:    make(map[string]interface{}),
	}, nil
}

// WithTag adds a tag to the version info
func (v *VersionInfo) WithTag(tag string) *VersionInfo {
	v.Tags = append(v.Tags, tag)
	return v
}

// WithMetadata adds metadata to the version info
func (v *VersionInfo) WithMetadata(key string, value interface{}) *VersionInfo {
	v.Metadata[key] = value
	return v
}

// IsCompatible checks if the version info is compatible with another version info
func (v *VersionInfo) IsCompatible(other *VersionInfo) bool {
	return v.Version.IsCompatible(other.Version)
}

// IsBackwardsCompatible checks if the version info is backwards compatible with another version info
func (v *VersionInfo) IsBackwardsCompatible(other *VersionInfo) bool {
	return v.Version.IsBackwardsCompatible(other.Version)
}

// GetVersionString returns the string representation of the version
func (v *VersionInfo) GetVersionString() string {
	return v.Version.String()
}

// UpdateVersion updates the version and sets the updated timestamp
func (v *VersionInfo) UpdateVersion(version string) error {
	semver, err := Parse(version)
	if err != nil {
		return err
	}
	
	v.Version = semver
	v.UpdatedAt = time.Now()
	
	return nil
}

// IncrementMajor increments the major version
func (v *VersionInfo) IncrementMajor() {
	v.Version = v.Version.IncrementMajor()
	v.UpdatedAt = time.Now()
}

// IncrementMinor increments the minor version
func (v *VersionInfo) IncrementMinor() {
	v.Version = v.Version.IncrementMinor()
	v.UpdatedAt = time.Now()
}

// IncrementPatch increments the patch version
func (v *VersionInfo) IncrementPatch() {
	v.Version = v.Version.IncrementPatch()
	v.UpdatedAt = time.Now()
}
